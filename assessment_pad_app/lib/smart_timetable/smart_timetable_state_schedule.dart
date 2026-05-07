part of '../smart_timetable_page.dart';

const String _scheduleCreateConfirmSkipDateKey =
    'smart_timetable_schedule_create_confirm_skip_date';

extension _SmartTimetableStateSchedule on _SmartTimetablePageState {
  void _toggleSchedulePanel() {
    _updateState(() {
      _schedulePanelOpen = !_schedulePanelOpen;
      _teacherDropdownOpen = false;
      _periodGroupDropdownOpen = false;
      _openFilterKind = null;
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
      _periodGroupDropdownOpen = false;
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
      _periodGroupDropdownOpen = false;
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
    final bool skipConfirm = await _shouldSkipScheduleCreateConfirm();
    if (!skipConfirm) {
      final _ScheduleCreatePreviewData preview =
          _buildScheduleCreatePreviewData(cell);
      final _ScheduleCreateConfirmResult? result =
          await _confirmScheduleCreation(preview);
      if (!mounted || result == null || !result.confirmed) {
        return;
      }
      if (result.dontAskAgainToday) {
        await _rememberScheduleCreateConfirmSkipToday();
      }
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
      await _loadTimetable(showSkeleton: false);
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

  Future<bool> _shouldSkipScheduleCreateConfirm() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    return prefs.getString(_scheduleCreateConfirmSkipDateKey) ==
        _formatApiDate(DateTime.now());
  }

  Future<void> _rememberScheduleCreateConfirmSkipToday() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    await prefs.setString(
      _scheduleCreateConfirmSkipDateKey,
      _formatApiDate(DateTime.now()),
    );
  }

  Future<_ScheduleCreateConfirmResult?> _confirmScheduleCreation(
    _ScheduleCreatePreviewData preview,
  ) async {
    if (!mounted) {
      return null;
    }
    return showDialog<_ScheduleCreateConfirmResult>(
      context: context,
      barrierDismissible: true,
      barrierColor: Colors.black.withOpacity(.34),
      builder: (BuildContext dialogContext) {
        return PadDialogViewport(
          child: _ScheduleCreateConfirmDialog(
            data: preview,
            onCancel: () => Navigator.of(dialogContext).pop(),
            onConfirm: (bool dontAskAgainToday) {
              Navigator.of(dialogContext).pop(
                _ScheduleCreateConfirmResult(
                  confirmed: true,
                  dontAskAgainToday: dontAskAgainToday,
                ),
              );
            },
          ),
        );
      },
    );
  }

  _ScheduleCreatePreviewData _buildScheduleCreatePreviewData(
    _ScheduleCellSlot cell,
  ) {
    final ScheduleTargetOption? target = _selectedScheduleTarget;
    final String targetTitle = target?.title.trim().isNotEmpty == true
        ? target!.title.trim()
        : '未命名对象';
    final String targetSubtitle = target?.subtitle.trim().isNotEmpty == true
        ? target!.subtitle.trim()
        : (target?.lessonName.trim().isNotEmpty == true
            ? target!.lessonName.trim()
            : '');
    final String titleLine =
        targetSubtitle.isEmpty ? targetTitle : '$targetTitle-$targetSubtitle';
    final String teacherName = _teachers.isEmpty
        ? '当前老师'
        : _teachers[_teacherIndex.clamp(0, _teachers.length - 1)].name.trim();
    final Map<String, String> assistantNameById = <String, String>{
      for (final ScheduleStaffOption staff in _currentGroupAssistantOptions)
        if (staff.id.trim().isNotEmpty && staff.name.trim().isNotEmpty)
          staff.id.trim(): staff.name.trim(),
      for (final _TeacherOption teacher in _teachers)
        if (teacher.id.trim().isNotEmpty && teacher.name.trim().isNotEmpty)
          teacher.id.trim(): teacher.name.trim(),
    };
    final List<String> assistantNames = <String>[];
    for (final String id in _normalizedAssistantIds) {
      final String name = assistantNameById[id.trim()] ?? '';
      if (name.isNotEmpty && !assistantNames.contains(name)) {
        assistantNames.add(name);
      }
    }
    final String assistantLabel =
        assistantNames.isEmpty ? '未安排' : assistantNames.join('、');
    final String classroomLabel =
        _selectedClassroom?.name.trim().isNotEmpty == true
            ? _selectedClassroom!.name.trim()
            : '未设置教室';
    final String slotLabel = cell.row >= 0 && cell.row < _timeSlots.length
        ? _timeSlots[cell.row].title.trim().isNotEmpty
            ? _timeSlots[cell.row].title.trim()
            : '第${cell.row + 1}节'
        : '第${cell.row + 1}节';
    final String dateLabel = _monthDayLabel(cell.date);
    final String weekdayLabel = _weekdayShortLabel(cell.date);
    final String timeLabel =
        '$dateLabel $weekdayLabel $slotLabel · ${cell.startTime}-${cell.endTime}';
    return _ScheduleCreatePreviewData(
      typeLabel:
          _scheduleTargetType == ScheduleTargetType.oneToOne ? '1v1' : '班课',
      titleLine: titleLine,
      targetLabel: targetTitle,
      timeLabel: timeLabel,
      teacherLabel: teacherName,
      assistantLabel: assistantLabel,
      classroomLabel: classroomLabel,
      groupLabel: _selectedPeriodGroupName,
    );
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
