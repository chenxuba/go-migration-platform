part of '../smart_timetable_page.dart';

extension _SmartTimetableStateSchedule on _SmartTimetablePageState {
  void _toggleSchedulePanel() {
    _updateState(() {
      _schedulePanelOpen = !_schedulePanelOpen;
      _teacherDropdownOpen = false;
    });
    if (_schedulePanelOpen &&
        !_scheduleOptionsLoading &&
        !_scheduleOptionsLoaded) {
      unawaited(_loadScheduleOptions());
    }
  }

  void _handlePrimaryScheduleTap() {
    if (!_schedulePanelOpen) {
      return;
    }
    _updateState(() {
      _schedulePanelOpen = false;
    });
  }

  void _closeSchedulePanel() {
    if (!_schedulePanelOpen) {
      return;
    }
    _updateState(() {
      _schedulePanelOpen = false;
    });
  }

  void _setScheduleMode(_ScheduleMode mode) {
    if (mode == _scheduleMode) {
      return;
    }
    _updateState(() {
      _scheduleMode = mode;
      _resetAvailabilityFields();
    });
    if (_selectedScheduleTarget != null) {
      unawaited(_detectScheduleAvailability());
    }
  }

  void _selectScheduleTarget(ScheduleTargetOption option) {
    if (option.disabled) {
      _showScheduleMessage('该排课对象当前状态不可排');
      return;
    }
    _updateState(() {
      if (_scheduleMode == _ScheduleMode.oneToOne) {
        _selectedOneToOneTarget = option;
      } else {
        _selectedGroupClassTarget = option;
      }
      _schedulePanelOpen = false;
      _resetAvailabilityFields();
    });
    unawaited(_detectScheduleAvailability());
  }

  void _clearScheduleTarget() {
    _updateState(() {
      if (_scheduleMode == _ScheduleMode.oneToOne) {
        _selectedOneToOneTarget = null;
      } else {
        _selectedGroupClassTarget = null;
      }
      _schedulePanelOpen = false;
      _resetAvailabilityFields();
    });
  }

  void _toggleAssistant(ScheduleStaffOption assistant) {
    final String id = assistant.id.trim();
    if (id.isEmpty) {
      return;
    }
    if (id == _selectedTeacherId.trim()) {
      _showScheduleMessage('主教与助教不能为同一人，已自动忽略');
      _updateState(() {
        _selectedAssistantIds.remove(id);
        _resetAvailabilityFields();
      });
      unawaited(_detectScheduleAvailability());
      return;
    }
    final bool inCurrentGroup = _currentGroupAssistantOptions.any(
      (ScheduleStaffOption item) => item.id == id,
    );
    if (!inCurrentGroup) {
      _showScheduleMessage('助教只能选择当前时段组的其他老师');
      return;
    }
    _updateState(() {
      final Set<String> next = Set<String>.from(_selectedAssistantIds);
      if (next.contains(id)) {
        next.remove(id);
      } else {
        next.add(id);
      }
      _selectedAssistantIds = next;
      _resetAvailabilityFields();
    });
    unawaited(_detectScheduleAvailability());
  }

  void _selectClassroom(ScheduleClassroomOption? classroom) {
    _updateState(() {
      _selectedClassroom = classroom;
      _resetAvailabilityFields();
    });
    unawaited(_detectScheduleAvailability());
  }

  Future<void> _detectScheduleAvailability() async {
    final int sequence = ++_availabilitySequence;
    final ScheduleTargetOption? target = _selectedScheduleTarget;
    final String teacherId = _selectedTeacherId.trim();
    if (target == null) {
      if (!mounted) {
        return;
      }
      _updateState(() {
        _resetAvailabilityFields(cancelPending: false);
      });
      return;
    }
    if (teacherId.isEmpty) {
      if (!mounted) {
        return;
      }
      _updateState(() {
        _availabilityLoading = false;
        _availabilityMessage = '当前课表没有可用老师';
        _slotAvailability = const <String, _SlotAvailability>{};
      });
      return;
    }
    final List<_ScheduleCellSlot> cells = _emptyScheduleSlotsForCurrentWeek();
    if (cells.isEmpty) {
      if (!mounted) {
        return;
      }
      _updateState(() {
        _availabilityLoading = false;
        _availabilityMessage = '本周暂无空闲时段';
        _slotAvailability = const <String, _SlotAvailability>{};
      });
      return;
    }
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      if (!mounted || sequence != _availabilitySequence) {
        return;
      }
      _updateState(() {
        _availabilityLoading = false;
        _availabilityMessage = '登录已失效，请重新登录';
        _slotAvailability = const <String, _SlotAvailability>{};
      });
      return;
    }
    final List<String> assistantIds = _normalizedAssistantIds;
    final String classroomId = _selectedClassroomId;
    if (mounted) {
      _updateState(() {
        _availabilityLoading = true;
        _availabilityMessage = '检测中';
        _slotAvailability = const <String, _SlotAvailability>{};
      });
    }
    try {
      final ScheduleValidationResult result =
          await widget.timetableClient.validateScheduleSlots(
        token,
        type: _scheduleTargetType,
        targetId: target.id,
        teacherId: teacherId,
        assistantIds: assistantIds,
        classroomId: classroomId,
        slots: cells.map((_ScheduleCellSlot cell) {
          return ScheduleSlotRequest(
            teacherId: teacherId,
            lessonDate: cell.date,
            startTime: cell.startTime,
            endTime: cell.endTime,
            assistantIds: assistantIds,
            classroomId: classroomId,
          );
        }).toList(),
      );
      if (!mounted || sequence != _availabilitySequence) {
        return;
      }
      final Map<String, ScheduleValidationItem> itemByKey =
          <String, ScheduleValidationItem>{};
      for (final ScheduleValidationItem item in result.items) {
        itemByKey[_availabilityKey(
          item.teacherId.trim().isEmpty ? teacherId : item.teacherId,
          item.lessonDate,
          item.startTime,
          item.endTime,
        )] = item;
      }
      final bool hasItemResult = result.items.isNotEmpty;
      final Map<String, _SlotAvailability> next = <String, _SlotAvailability>{};
      for (final _ScheduleCellSlot cell in cells) {
        final ScheduleValidationItem? item = itemByKey[cell.key];
        final bool valid = hasItemResult ? (item?.valid ?? true) : result.valid;
        final List<String> conflictTypes =
            item?.conflictTypes ?? const <String>[];
        final String message = item?.message.trim().isNotEmpty == true
            ? item!.message
            : valid
                ? '空闲时段可排'
                : (result.message.trim().isEmpty ? '该时间段不可排课' : result.message);
        next[cell.key] = _SlotAvailability(
          valid: valid,
          message: message,
          conflictTypes: conflictTypes,
        );
      }
      final int validCount =
          next.values.where((_SlotAvailability item) => item.valid).length;
      final int invalidCount = next.length - validCount;
      _updateState(() {
        _availabilityLoading = false;
        _slotAvailability = next;
        _availabilityMessage = invalidCount == 0
            ? '可排 $validCount'
            : '可排 $validCount，冲突 $invalidCount';
      });
    } on TimetableApiException catch (error) {
      if (!mounted || sequence != _availabilitySequence) {
        return;
      }
      _updateState(() {
        _availabilityLoading = false;
        _availabilityMessage = error.message;
        _slotAvailability = const <String, _SlotAvailability>{};
      });
    } on Object catch (error) {
      if (!mounted || sequence != _availabilitySequence) {
        return;
      }
      _updateState(() {
        _availabilityLoading = false;
        _availabilityMessage = '空闲点检测失败：$error';
        _slotAvailability = const <String, _SlotAvailability>{};
      });
    }
  }

  Future<void> _handleEmptySlotTap(int row, int column) async {
    final _ScheduleCellSlot? cell = _scheduleCellSlotAt(row, column);
    if (cell == null) {
      return;
    }
    final ScheduleTargetOption? target = _selectedScheduleTarget;
    if (target == null) {
      _updateState(() {
        _schedulePanelOpen = true;
        _teacherDropdownOpen = false;
      });
      _showScheduleMessage('请先选择要排课的1v1或班课');
      return;
    }
    if (_availabilityLoading) {
      _showScheduleMessage('正在检测本周空闲点，请稍后');
      return;
    }
    final _SlotAvailability? availability = _slotAvailability[cell.key];
    if (availability == null) {
      _showScheduleMessage('请先等待空闲点检测完成');
      unawaited(_detectScheduleAvailability());
      return;
    }
    if (!availability.valid) {
      _showScheduleMessage(availability.message);
      return;
    }
    if (_creatingSlotKey != null) {
      return;
    }
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      _showScheduleMessage('登录已失效，请重新登录');
      return;
    }
    final List<String> assistantIds = _normalizedAssistantIds;
    final String classroomId = _selectedClassroomId;
    _updateState(() {
      _creatingSlotKey = cell.key;
    });
    try {
      await widget.timetableClient.createSchedule(
        token,
        type: _scheduleTargetType,
        targetId: target.id,
        teacherId: _selectedTeacherId,
        assistantIds: assistantIds,
        classroomId: classroomId,
        slot: ScheduleSlotRequest(
          teacherId: _selectedTeacherId,
          lessonDate: cell.date,
          startTime: cell.startTime,
          endTime: cell.endTime,
          assistantIds: assistantIds,
          classroomId: classroomId,
        ),
      );
      _showScheduleMessage(
        '排课成功，已刷新课表',
        tone: PadMessageTone.success,
      );
      await _loadTimetable();
    } on TimetableApiException catch (error) {
      _showScheduleMessage(error.message);
      unawaited(_detectScheduleAvailability());
    } on Object catch (error) {
      _showScheduleMessage('创建排课失败：$error');
      unawaited(_detectScheduleAvailability());
    } finally {
      if (mounted) {
        _updateState(() {
          _creatingSlotKey = null;
        });
      }
    }
  }

  List<_ScheduleCellSlot> _emptyScheduleSlotsForCurrentWeek() {
    final List<_ScheduleCellSlot> cells = <_ScheduleCellSlot>[];
    for (int row = 0; row < _timeSlots.length; row += 1) {
      for (int column = 0; column < _weekDays.length; column += 1) {
        if (row < _scheduleRows.length &&
            column < _scheduleRows[row].length &&
            _scheduleRows[row][column] != null) {
          continue;
        }
        final _ScheduleCellSlot? cell = _scheduleCellSlotAt(row, column);
        if (cell != null) {
          cells.add(cell);
        }
      }
    }
    return cells;
  }

  _ScheduleCellSlot? _scheduleCellSlotAt(int row, int column) {
    if (row < 0 ||
        column < 0 ||
        row >= _timeSlots.length ||
        column >= _weekDays.length) {
      return null;
    }
    final _TimeSlot slot = _timeSlots[row];
    final _WeekDay day = _weekDays[column];
    final String teacherId = _selectedTeacherId.trim();
    if (teacherId.isEmpty ||
        day.isoDate.trim().isEmpty ||
        slot.startTime.trim().isEmpty ||
        slot.endTime.trim().isEmpty) {
      return null;
    }
    return _ScheduleCellSlot(
      row: row,
      column: column,
      key: _availabilityKey(
        teacherId,
        day.isoDate,
        slot.startTime,
        slot.endTime,
      ),
      date: day.isoDate,
      startTime: slot.startTime,
      endTime: slot.endTime,
    );
  }
}
