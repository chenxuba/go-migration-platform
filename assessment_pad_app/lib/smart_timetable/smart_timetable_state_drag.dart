part of '../smart_timetable_page.dart';

extension _SmartTimetableStateDrag on _SmartTimetablePageState {
  void _moveLesson(_LessonDragData source, int targetRow, int targetColumn) {
    unawaited(_moveLessonToSlot(source, targetRow, targetColumn));
  }

  Future<void> _moveLessonToSlot(
    _LessonDragData source,
    int targetRow,
    int targetColumn,
  ) async {
    if (_movingScheduleId != null) {
      return;
    }
    if (source.row == targetRow && source.column == targetColumn) {
      return;
    }
    if (source.row < 0 ||
        source.column < 0 ||
        source.row >= _scheduleRows.length ||
        source.column >= _scheduleRows[source.row].length ||
        targetRow < 0 ||
        targetColumn < 0 ||
        targetRow >= _scheduleRows.length ||
        targetColumn >= _scheduleRows[targetRow].length) {
      return;
    }

    final _LessonCell? sourceLesson = _scheduleRows[source.row][source.column];
    if (sourceLesson == null) {
      return;
    }
    if (sourceLesson.id.trim().isEmpty) {
      _showScheduleMessage('缺少日程ID，无法调课');
      return;
    }
    if (_scheduleRows[targetRow][targetColumn] != null) {
      _showScheduleMessage('目标时段已有课程，不能调课');
      return;
    }
    final _ScheduleCellSlot? targetSlot =
        _scheduleCellSlotAt(targetRow, targetColumn);
    if (targetSlot == null) {
      _showScheduleMessage('目标时段无效，不能调课');
      return;
    }
    final _SlotAvailability? localAvailability =
        _localDragValidation(sourceLesson, targetSlot);
    final _SlotAvailability? cachedAvailability =
        _dragSlotAvailability[targetSlot.key];
    final _SlotAvailability? availability =
        localAvailability ?? cachedAvailability;
    if (availability == null) {
      _showScheduleMessage(
        _dragCheckingSlotKeys.contains(targetSlot.key)
            ? '正在检测调课空点，请稍后'
            : '请先等待调课空点检测完成',
      );
      return;
    }
    if (!availability.valid) {
      _showScheduleMessage(availability.message);
      return;
    }

    final List<List<_LessonCell?>> previousRows =
        _cloneScheduleRows(_scheduleRows);
    _updateState(() {
      _movingScheduleId = sourceLesson.id;
      final List<List<_LessonCell?>> nextRows =
          _cloneScheduleRows(_scheduleRows);
      nextRows[source.row][source.column] = null;
      nextRows[targetRow][targetColumn] = sourceLesson;
      _scheduleRows = nextRows;
      _dragValidationActive = false;
      _dragCheckingSlotKeys = const <String>{};
    });
    try {
      final String token = await _readAuthToken();
      if (token.trim().isEmpty) {
        _showScheduleMessage('登录已失效，请重新登录');
        _rollbackOptimisticLessonMove(previousRows);
        return;
      }
      await widget.timetableClient.updateScheduleSlot(
        token,
        scheduleId: sourceLesson.id,
        teacherId: _selectedTeacherId,
        assistantIds: sourceLesson.assistantIds,
        classroomId: sourceLesson.classroomId,
        slot: ScheduleSlotRequest(
          teacherId: _selectedTeacherId,
          lessonDate: targetSlot.date,
          startTime: targetSlot.startTime,
          endTime: targetSlot.endTime,
          assistantIds: sourceLesson.assistantIds ?? const <String>[],
          classroomId: sourceLesson.classroomId ?? '',
        ),
      );
      _showScheduleMessage(
        '调课成功，已刷新课表',
        tone: PadMessageTone.success,
      );
      await _loadTimetable();
    } on TimetableApiException catch (error) {
      _showScheduleMessage(error.message);
      _rollbackOptimisticLessonMove(previousRows);
    } on Object catch (error) {
      _showScheduleMessage('调课失败：$error');
      _rollbackOptimisticLessonMove(previousRows);
    } finally {
      if (mounted) {
        _updateState(() {
          _movingScheduleId = null;
        });
      }
    }
  }

  void _rollbackOptimisticLessonMove(List<List<_LessonCell?>> previousRows) {
    if (!mounted) {
      return;
    }
    _updateState(() {
      _scheduleRows = previousRows;
    });
  }

  void _startLessonDrag(_LessonDragData source) {
    _updateState(() {
      _dragValidationActive = true;
      _dragSlotAvailability = const <String, _SlotAvailability>{};
      _dragCheckingSlotKeys = const <String>{};
    });
    unawaited(_primeDragSlotAvailability(source));
  }

  void _endLessonDrag() {
    _updateState(() {
      _dragValidationActive = false;
      _dragCheckingSlotKeys = const <String>{};
    });
  }

  Future<void> _primeDragSlotAvailability(_LessonDragData source) async {
    final int sequence = ++_dragValidationSequence;
    final _LessonCell? sourceLesson = _lessonAt(source.row, source.column);
    if (sourceLesson == null) {
      return;
    }
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      return;
    }
    final List<_ScheduleCellSlot> cells = _emptyScheduleSlotsForCurrentWeek();
    final Map<String, _SlotAvailability> localResults =
        <String, _SlotAvailability>{};
    final List<_ScheduleCellSlot> remoteCells = <_ScheduleCellSlot>[];
    for (final _ScheduleCellSlot cell in cells) {
      final _SlotAvailability? local = _localDragValidation(sourceLesson, cell);
      if (local != null) {
        localResults[cell.key] = local;
      } else {
        remoteCells.add(cell);
      }
    }
    if (!mounted || sequence != _dragValidationSequence) {
      return;
    }
    _updateState(() {
      _dragValidationActive = true;
      _dragSlotAvailability = localResults;
      _dragCheckingSlotKeys =
          remoteCells.map((_ScheduleCellSlot cell) => cell.key).toSet();
    });
    if (remoteCells.isEmpty) {
      return;
    }
    try {
      final Map<String, _SlotAvailability> remoteResults =
          await _fetchDragTargetsValidation(sourceLesson, remoteCells, token);
      if (!mounted || sequence != _dragValidationSequence) {
        return;
      }
      _updateState(() {
        _dragSlotAvailability = <String, _SlotAvailability>{
          ...localResults,
          ...remoteResults,
        };
        _dragCheckingSlotKeys = const <String>{};
      });
    } on Object {
      if (!mounted || sequence != _dragValidationSequence) {
        return;
      }
      _updateState(() {
        _dragCheckingSlotKeys = const <String>{};
      });
    }
  }

  Future<Map<String, _SlotAvailability>> _fetchDragTargetsValidation(
    _LessonCell sourceLesson,
    List<_ScheduleCellSlot> targetSlots,
    String token,
  ) async {
    if (targetSlots.isEmpty) {
      return const <String, _SlotAvailability>{};
    }
    try {
      final ScheduleValidationResult result =
          await widget.timetableClient.validateScheduleSlots(
        token,
        type: sourceLesson.scheduleTargetType!,
        targetId: sourceLesson.teachingClassId,
        teacherId: _selectedTeacherId,
        assistantIds: sourceLesson.assistantIds ?? const <String>[],
        classroomId: sourceLesson.classroomId ?? '',
        slots: targetSlots.map((_ScheduleCellSlot targetSlot) {
          return ScheduleSlotRequest(
            teacherId: _selectedTeacherId,
            lessonDate: targetSlot.date,
            startTime: targetSlot.startTime,
            endTime: targetSlot.endTime,
            assistantIds: sourceLesson.assistantIds ?? const <String>[],
            classroomId: sourceLesson.classroomId ?? '',
          );
        }).toList(),
        excludeIds: <String>[sourceLesson.id],
      );
      final Map<String, ScheduleValidationItem> itemByKey =
          <String, ScheduleValidationItem>{};
      for (final ScheduleValidationItem item in result.items) {
        itemByKey[_availabilityKey(
          item.teacherId.trim().isEmpty ? _selectedTeacherId : item.teacherId,
          item.lessonDate,
          item.startTime,
          item.endTime,
        )] = item;
      }
      final bool hasItemResult = result.items.isNotEmpty;
      final Map<String, _SlotAvailability> next = <String, _SlotAvailability>{};
      for (final _ScheduleCellSlot cell in targetSlots) {
        final ScheduleValidationItem? item = itemByKey[cell.key];
        final bool valid = hasItemResult ? (item?.valid ?? true) : result.valid;
        final String message = item?.message.trim().isNotEmpty == true
            ? item!.message
            : valid
                ? '可调课'
                : (result.message.trim().isEmpty ? '当前空点不可调课' : result.message);
        next[cell.key] = _SlotAvailability(
          valid: valid,
          message: message,
          conflictTypes: item?.conflictTypes ?? const <String>[],
        );
      }
      return next;
    } on TimetableApiException catch (error) {
      return <String, _SlotAvailability>{
        for (final _ScheduleCellSlot cell in targetSlots)
          cell.key: _SlotAvailability(valid: false, message: error.message),
      };
    } on Object catch (error) {
      return <String, _SlotAvailability>{
        for (final _ScheduleCellSlot cell in targetSlots)
          cell.key: _SlotAvailability(
            valid: false,
            message: '检测调课空点失败：$error',
          ),
      };
    }
  }

  _SlotAvailability? _localDragValidation(
    _LessonCell sourceLesson,
    _ScheduleCellSlot targetSlot,
  ) {
    if (sourceLesson.id.trim().isEmpty) {
      return const _SlotAvailability(valid: false, message: '缺少日程ID，无法调课');
    }
    if (sourceLesson.teachingClassId.trim().isEmpty ||
        sourceLesson.scheduleTargetType == null) {
      return const _SlotAvailability(valid: false, message: '当前课程信息不完整，无法检测');
    }
    if (_isPastIsoDate(targetSlot.date)) {
      return const _SlotAvailability(valid: false, message: '不允许调课到过去日期');
    }
    final _LessonCell? targetLesson =
        _lessonAt(targetSlot.row, targetSlot.column);
    if (targetLesson != null) {
      return const _SlotAvailability(valid: false, message: '目标时段已有课程，不能调课');
    }
    return null;
  }

  _LessonCell? _lessonAt(int row, int column) {
    if (row < 0 ||
        column < 0 ||
        row >= _scheduleRows.length ||
        column >= _scheduleRows[row].length) {
      return null;
    }
    return _scheduleRows[row][column];
  }

  List<List<_LessonCell?>> _cloneScheduleRows(List<List<_LessonCell?>> rows) {
    return rows
        .map((List<_LessonCell?> row) => List<_LessonCell?>.of(row))
        .toList();
  }

  bool _isPastIsoDate(String isoDate) {
    final DateTime? parsed = DateTime.tryParse(isoDate);
    if (parsed == null) {
      return false;
    }
    final DateTime target = DateTime(parsed.year, parsed.month, parsed.day);
    final DateTime now = DateTime.now();
    final DateTime today = DateTime(now.year, now.month, now.day);
    return target.isBefore(today);
  }
}
