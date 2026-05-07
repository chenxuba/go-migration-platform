part of '../smart_timetable_page.dart';

class _TimetableBoard extends StatelessWidget {
  const _TimetableBoard({
    required this.compact,
    required this.rows,
    required this.weekDays,
    required this.timeSlots,
    required this.selectedTeacherId,
    required this.scheduleTargetSelected,
    required this.availabilityLoading,
    required this.slotAvailability,
    required this.dragSlotAvailability,
    required this.dragCheckingSlotKeys,
    required this.dragValidationActive,
    required this.creatingSlotKey,
    required this.onLessonMove,
    required this.onLessonDragStarted,
    required this.onLessonDragEnded,
    required this.onEmptySlotTap,
  });

  final bool compact;
  final List<List<_LessonCell?>> rows;
  final List<_WeekDay> weekDays;
  final List<_TimeSlot> timeSlots;
  final String selectedTeacherId;
  final bool scheduleTargetSelected;
  final bool availabilityLoading;
  final Map<String, _SlotAvailability> slotAvailability;
  final Map<String, _SlotAvailability> dragSlotAvailability;
  final Set<String> dragCheckingSlotKeys;
  final bool dragValidationActive;
  final String? creatingSlotKey;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;
  final ValueChanged<_LessonDragData> onLessonDragStarted;
  final VoidCallback onLessonDragEnded;
  final void Function(int row, int column) onEmptySlotTap;

  @override
  Widget build(BuildContext context) {
    final double leftWidth = compact ? 112 : 118;
    final int rowCount = math.max(timeSlots.length, rows.length);
    final List<List<_LessonCell?>> displayRows =
        List<List<_LessonCell?>>.generate(rowCount, (int rowIndex) {
      final List<_LessonCell?> source =
          rowIndex < rows.length ? rows[rowIndex] : const <_LessonCell?>[];
      return List<_LessonCell?>.generate(
        weekDays.length,
        (int column) => column < source.length ? source[column] : null,
      );
    });

    return Align(
      alignment: Alignment.topCenter,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(14),
        child: Container(
          key: const ValueKey<String>('smart-timetable-board'),
          color: Colors.white,
          foregroundDecoration: BoxDecoration(
            border: Border.all(color: _SmartColors.line),
            borderRadius: BorderRadius.circular(14),
          ),
          child: Column(
            children: <Widget>[
              SizedBox(
                key: const ValueKey<String>('smart-timetable-header'),
                height: _headerHeight,
                child: Row(
                  children: <Widget>[
                    SizedBox(
                      width: leftWidth,
                      child: const _DiagonalHeaderCell(),
                    ),
                    Expanded(
                      child: _WeekHeaderRow(
                        key: const ValueKey<String>('smart-week-header'),
                        weekDays: weekDays,
                      ),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: SingleChildScrollView(
                  key: const ValueKey<String>('smart-timetable-scroll-body'),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      SizedBox(
                        width: leftWidth,
                        child: _TimeRail(
                          rowCount: rowCount,
                          timeSlots: timeSlots,
                        ),
                      ),
                      Expanded(
                        child: _ScheduleGrid(
                          rows: displayRows,
                          weekDays: weekDays,
                          timeSlots: timeSlots,
                          selectedTeacherId: selectedTeacherId,
                          scheduleTargetSelected: scheduleTargetSelected,
                          availabilityLoading: availabilityLoading,
                          slotAvailability: slotAvailability,
                          dragSlotAvailability: dragSlotAvailability,
                          dragCheckingSlotKeys: dragCheckingSlotKeys,
                          dragValidationActive: dragValidationActive,
                          creatingSlotKey: creatingSlotKey,
                          onLessonMove: onLessonMove,
                          onLessonDragStarted: onLessonDragStarted,
                          onLessonDragEnded: onLessonDragEnded,
                          onEmptySlotTap: onEmptySlotTap,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _TimetableBoardSkeleton extends StatelessWidget {
  const _TimetableBoardSkeleton();

  @override
  Widget build(BuildContext context) {
    const int columnCount = 7;
    const int rowCount = 8;
    const double leftWidth = 118;

    return Align(
      alignment: Alignment.topCenter,
      child: ClipRRect(
        borderRadius: BorderRadius.circular(14),
        child: Container(
          key: const ValueKey<String>('smart-timetable-skeleton-board'),
          color: Colors.white,
          foregroundDecoration: BoxDecoration(
            border: Border.all(color: _SmartColors.line),
            borderRadius: BorderRadius.circular(14),
          ),
          child: Column(
            children: <Widget>[
              SizedBox(
                height: _headerHeight,
                child: Row(
                  children: <Widget>[
                    Container(
                      width: leftWidth,
                      height: _headerHeight,
                      decoration: const BoxDecoration(
                        color: _SmartColors.surface,
                        border: Border(
                          right: BorderSide(color: _SmartColors.line),
                          bottom: BorderSide(color: _SmartColors.line),
                        ),
                      ),
                      child: const Padding(
                        padding: EdgeInsets.fromLTRB(10, 10, 10, 8),
                        child: Row(
                          children: <Widget>[
                            Expanded(
                              child: Align(
                                alignment: Alignment.topRight,
                                child: _TimetableSkeletonBox(
                                  width: 24,
                                  height: 10,
                                  radius: 6,
                                ),
                              ),
                            ),
                            SizedBox(width: 12),
                            Align(
                              alignment: Alignment.bottomLeft,
                              child: _TimetableSkeletonBox(
                                width: 24,
                                height: 10,
                                radius: 6,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    Expanded(
                      child: Row(
                        children:
                            List<Widget>.generate(columnCount, (int index) {
                          return Expanded(
                            child: Container(
                              decoration: BoxDecoration(
                                color: _SmartColors.surface,
                                border: Border(
                                  right: BorderSide(
                                    color: index == columnCount - 1
                                        ? Colors.transparent
                                        : _SmartColors.line,
                                  ),
                                  bottom: const BorderSide(
                                      color: _SmartColors.line),
                                ),
                              ),
                              child: const Center(
                                child: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  children: <Widget>[
                                    _TimetableSkeletonBox(
                                      width: 32,
                                      height: 12,
                                      radius: 6,
                                    ),
                                    SizedBox(height: 5),
                                    _TimetableSkeletonBox(
                                      width: 52,
                                      height: 10,
                                      radius: 6,
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          );
                        }),
                      ),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: SingleChildScrollView(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      SizedBox(
                        width: leftWidth,
                        child: ColoredBox(
                          color: _SmartColors.slot,
                          child: Column(
                            children:
                                List<Widget>.generate(rowCount, (int index) {
                              return Container(
                                key: ValueKey<String>(
                                  'smart-timetable-skeleton-row-$index',
                                ),
                                height: _rowHeight,
                                width: double.infinity,
                                padding:
                                    const EdgeInsets.fromLTRB(10, 8, 10, 8),
                                decoration: BoxDecoration(
                                  color: _SmartColors.slot,
                                  border: Border(
                                    right: const BorderSide(
                                        color: _SmartColors.line),
                                    bottom: BorderSide(
                                      color: index == rowCount - 1
                                          ? Colors.transparent
                                          : _SmartColors.lineSoft,
                                    ),
                                  ),
                                ),
                                child: const Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: <Widget>[
                                    _TimetableSkeletonBox(
                                      width: 30,
                                      height: 12,
                                      radius: 6,
                                    ),
                                    SizedBox(height: 8),
                                    _TimetableSkeletonBox(
                                      width: 66,
                                      height: 10,
                                      radius: 6,
                                    ),
                                  ],
                                ),
                              );
                            }),
                          ),
                        ),
                      ),
                      Expanded(
                        child: Column(
                          children: List<Widget>.generate(rowCount, (int row) {
                            return SizedBox(
                              height: _rowHeight,
                              child: Row(
                                children: List<Widget>.generate(columnCount,
                                    (int column) {
                                  final bool showCard = row == 0 ||
                                      row == 2 &&
                                          (column == 1 || column == 4) ||
                                      row == 4 &&
                                          (column == 0 || column == 5) ||
                                      row == 6 && column == 3;
                                  return Expanded(
                                    child: Container(
                                      padding: const EdgeInsets.all(6),
                                      decoration: BoxDecoration(
                                        color: Colors.white.withOpacity(.72),
                                        border: Border(
                                          right: BorderSide(
                                            color: column == columnCount - 1
                                                ? Colors.transparent
                                                : _SmartColors.lineSoft,
                                          ),
                                          bottom: BorderSide(
                                            color: row == rowCount - 1
                                                ? Colors.transparent
                                                : _SmartColors.lineSoft,
                                          ),
                                        ),
                                      ),
                                      child: showCard
                                          ? const _TimetableSkeletonLessonCard()
                                          : const SizedBox.expand(),
                                    ),
                                  );
                                }),
                              ),
                            );
                          }),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _TimetableSkeletonLessonCard extends StatelessWidget {
  const _TimetableSkeletonLessonCard();

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFFFF7F0),
        border: Border.all(color: const Color(0xFFF2E3D6)),
        borderRadius: BorderRadius.circular(11),
      ),
      padding: const EdgeInsets.fromLTRB(9, 7, 9, 6),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: <Widget>[
          _TimetableSkeletonBox(height: 10, radius: 6),
          Row(
            children: <Widget>[
              Expanded(
                child: _TimetableSkeletonBox(height: 8, radius: 6),
              ),
              SizedBox(width: 8),
              _TimetableSkeletonBox(width: 28, height: 12, radius: 6),
            ],
          ),
        ],
      ),
    );
  }
}

class _TimeRail extends StatelessWidget {
  const _TimeRail({required this.rowCount, required this.timeSlots});

  final int rowCount;
  final List<_TimeSlot> timeSlots;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      key: const ValueKey<String>('smart-time-rail'),
      color: _SmartColors.slot,
      child: Column(
        children: <Widget>[
          for (int index = 0; index < rowCount; index += 1)
            _TimeSlotCell(
              slot: _timeSlotForIndex(index, timeSlots),
              isLast: index == rowCount - 1,
            ),
        ],
      ),
    );
  }
}

_TimeSlot _timeSlotForIndex(int index, List<_TimeSlot> timeSlots) {
  if (index < timeSlots.length) {
    return timeSlots[index];
  }
  return _TimeSlot(
    title: '第${index + 1}节',
    time: '',
    startTime: '',
    endTime: '',
  );
}

class _DiagonalHeaderCell extends StatelessWidget {
  const _DiagonalHeaderCell();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: _headerHeight,
      decoration: BoxDecoration(
        color: _SmartColors.surface,
        border: const Border(
          right: BorderSide(color: _SmartColors.line),
          bottom: BorderSide(color: _SmartColors.line),
        ),
      ),
      child: Stack(
        children: <Widget>[
          Positioned.fill(child: CustomPaint(painter: _DiagonalPainter())),
          const Positioned(
            left: 10,
            bottom: 8,
            child: Text(
              '时段',
              style: TextStyle(
                color: _SmartColors.muted,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          const Positioned(
            right: 10,
            top: 8,
            child: Text(
              '日期',
              style: TextStyle(
                color: _SmartColors.muted,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _TimeSlotCell extends StatelessWidget {
  const _TimeSlotCell({required this.slot, required this.isLast});

  final _TimeSlot slot;
  final bool isLast;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: _rowHeight,
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(10, 8, 4, 0),
      decoration: BoxDecoration(
        color: _SmartColors.slot,
        border: Border(
          right: const BorderSide(color: _SmartColors.line),
          bottom: BorderSide(
            color: isLast ? Colors.transparent : _SmartColors.lineSoft,
          ),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            slot.title,
            style: const TextStyle(
              color: _SmartColors.ink,
              fontSize: 13,
              height: 1.1,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 7),
          Text(
            slot.time,
            style: const TextStyle(
              color: _SmartColors.muted,
              fontSize: 11,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _ScheduleGrid extends StatelessWidget {
  const _ScheduleGrid({
    required this.rows,
    required this.weekDays,
    required this.timeSlots,
    required this.selectedTeacherId,
    required this.scheduleTargetSelected,
    required this.availabilityLoading,
    required this.slotAvailability,
    required this.dragSlotAvailability,
    required this.dragCheckingSlotKeys,
    required this.dragValidationActive,
    required this.creatingSlotKey,
    required this.onLessonMove,
    required this.onLessonDragStarted,
    required this.onLessonDragEnded,
    required this.onEmptySlotTap,
  });

  final List<List<_LessonCell?>> rows;
  final List<_WeekDay> weekDays;
  final List<_TimeSlot> timeSlots;
  final String selectedTeacherId;
  final bool scheduleTargetSelected;
  final bool availabilityLoading;
  final Map<String, _SlotAvailability> slotAvailability;
  final Map<String, _SlotAvailability> dragSlotAvailability;
  final Set<String> dragCheckingSlotKeys;
  final bool dragValidationActive;
  final String? creatingSlotKey;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;
  final ValueChanged<_LessonDragData> onLessonDragStarted;
  final VoidCallback onLessonDragEnded;
  final void Function(int row, int column) onEmptySlotTap;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      key: const ValueKey<String>('smart-schedule-grid'),
      color: Colors.white,
      child: Column(
        children: <Widget>[
          for (int row = 0; row < rows.length; row += 1)
            _ScheduleGridRow(
              rowIndex: row,
              row: rows[row],
              weekDays: weekDays,
              timeSlot: row < timeSlots.length ? timeSlots[row] : null,
              selectedTeacherId: selectedTeacherId,
              scheduleTargetSelected: scheduleTargetSelected,
              availabilityLoading: availabilityLoading,
              slotAvailability: slotAvailability,
              dragSlotAvailability: dragSlotAvailability,
              dragCheckingSlotKeys: dragCheckingSlotKeys,
              dragValidationActive: dragValidationActive,
              creatingSlotKey: creatingSlotKey,
              isLastRow: row == rows.length - 1,
              onLessonMove: onLessonMove,
              onLessonDragStarted: onLessonDragStarted,
              onLessonDragEnded: onLessonDragEnded,
              onEmptySlotTap: onEmptySlotTap,
            ),
        ],
      ),
    );
  }
}

class _WeekHeaderRow extends StatelessWidget {
  const _WeekHeaderRow({required this.weekDays, super.key});

  final List<_WeekDay> weekDays;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: _headerHeight,
      child: Row(
        children: <Widget>[
          for (int index = 0; index < weekDays.length; index += 1)
            Expanded(
              child: _WeekHeaderCell(
                day: weekDays[index],
                isLast: index == weekDays.length - 1,
              ),
            ),
        ],
      ),
    );
  }
}

class _WeekHeaderCell extends StatelessWidget {
  const _WeekHeaderCell({
    required this.day,
    required this.isLast,
  });

  final _WeekDay day;
  final bool isLast;

  @override
  Widget build(BuildContext context) {
    final bool isToday = day.isToday;
    return Container(
      height: _headerHeight,
      decoration: BoxDecoration(
        color: isToday ? _SmartColors.todayHeader : _SmartColors.surface,
        border: Border(
          right: BorderSide(
            color: isLast ? Colors.transparent : _SmartColors.lineSoft,
          ),
          bottom: const BorderSide(color: _SmartColors.line),
        ),
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: <Widget>[
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              Text(
                day.label,
                style: TextStyle(
                  color: isToday ? _SmartColors.orangeDeep : _SmartColors.ink,
                  fontSize: 13,
                  fontWeight: FontWeight.w900,
                ),
              ),
              if (isToday) ...<Widget>[
                const SizedBox(width: 5),
                Container(
                  height: 16,
                  padding: const EdgeInsets.symmetric(horizontal: 6),
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFE1D1),
                    borderRadius: BorderRadius.circular(999),
                  ),
                  child: const Text(
                    '今天',
                    style: TextStyle(
                      color: _SmartColors.orangeDeep,
                      fontSize: 10,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ],
            ],
          ),
          const SizedBox(height: 2),
          Text(
            day.date,
            style: TextStyle(
              color: isToday ? _SmartColors.orangeDeep : _SmartColors.muted,
              fontSize: 11,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _ScheduleGridRow extends StatelessWidget {
  const _ScheduleGridRow({
    required this.rowIndex,
    required this.row,
    required this.weekDays,
    required this.timeSlot,
    required this.selectedTeacherId,
    required this.scheduleTargetSelected,
    required this.availabilityLoading,
    required this.slotAvailability,
    required this.dragSlotAvailability,
    required this.dragCheckingSlotKeys,
    required this.dragValidationActive,
    required this.creatingSlotKey,
    required this.isLastRow,
    required this.onLessonMove,
    required this.onLessonDragStarted,
    required this.onLessonDragEnded,
    required this.onEmptySlotTap,
  });

  final int rowIndex;
  final List<_LessonCell?> row;
  final List<_WeekDay> weekDays;
  final _TimeSlot? timeSlot;
  final String selectedTeacherId;
  final bool scheduleTargetSelected;
  final bool availabilityLoading;
  final Map<String, _SlotAvailability> slotAvailability;
  final Map<String, _SlotAvailability> dragSlotAvailability;
  final Set<String> dragCheckingSlotKeys;
  final bool dragValidationActive;
  final String? creatingSlotKey;
  final bool isLastRow;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;
  final ValueChanged<_LessonDragData> onLessonDragStarted;
  final VoidCallback onLessonDragEnded;
  final void Function(int row, int column) onEmptySlotTap;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: _rowHeight,
      child: Row(
        children: <Widget>[
          for (int column = 0; column < row.length; column += 1)
            Expanded(
              child: Builder(
                builder: (BuildContext context) {
                  final String availabilityKey =
                      _availabilityKeyForGridCell(column);
                  return _ScheduleGridCell(
                    lesson: row[column],
                    rowIndex: rowIndex,
                    columnIndex: column,
                    availability: slotAvailability[availabilityKey],
                    dragAvailability: dragSlotAvailability[availabilityKey],
                    dragChecking:
                        dragCheckingSlotKeys.contains(availabilityKey),
                    dragValidationActive: dragValidationActive,
                    scheduleTargetSelected: scheduleTargetSelected,
                    availabilityLoading: availabilityLoading,
                    creating: creatingSlotKey == availabilityKey,
                    isToday: column < weekDays.length
                        ? weekDays[column].isToday
                        : false,
                    isLastColumn: column == row.length - 1,
                    isLastRow: isLastRow,
                    onLessonMove: onLessonMove,
                    onLessonDragStarted: onLessonDragStarted,
                    onLessonDragEnded: onLessonDragEnded,
                    onEmptySlotTap: onEmptySlotTap,
                  );
                },
              ),
            ),
        ],
      ),
    );
  }

  String _availabilityKeyForGridCell(int column) {
    if (timeSlot == null ||
        column >= weekDays.length ||
        selectedTeacherId.trim().isEmpty) {
      return '';
    }
    return _availabilityKey(
      selectedTeacherId,
      weekDays[column].isoDate,
      timeSlot!.startTime,
      timeSlot!.endTime,
    );
  }
}

class _ScheduleGridCell extends StatelessWidget {
  const _ScheduleGridCell({
    required this.lesson,
    required this.rowIndex,
    required this.columnIndex,
    required this.availability,
    required this.dragAvailability,
    required this.dragChecking,
    required this.dragValidationActive,
    required this.scheduleTargetSelected,
    required this.availabilityLoading,
    required this.creating,
    required this.isToday,
    required this.isLastColumn,
    required this.isLastRow,
    required this.onLessonMove,
    required this.onLessonDragStarted,
    required this.onLessonDragEnded,
    required this.onEmptySlotTap,
  });

  final _LessonCell? lesson;
  final int rowIndex;
  final int columnIndex;
  final _SlotAvailability? availability;
  final _SlotAvailability? dragAvailability;
  final bool dragChecking;
  final bool dragValidationActive;
  final bool scheduleTargetSelected;
  final bool availabilityLoading;
  final bool creating;
  final bool isToday;
  final bool isLastColumn;
  final bool isLastRow;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;
  final ValueChanged<_LessonDragData> onLessonDragStarted;
  final VoidCallback onLessonDragEnded;
  final void Function(int row, int column) onEmptySlotTap;

  @override
  Widget build(BuildContext context) {
    return DragTarget<_LessonDragData>(
      onWillAcceptWithDetails: (DragTargetDetails<_LessonDragData> details) =>
          details.data.row != rowIndex || details.data.column != columnIndex,
      onAcceptWithDetails: (DragTargetDetails<_LessonDragData> details) =>
          onLessonMove(details.data, rowIndex, columnIndex),
      builder: (
        BuildContext context,
        List<_LessonDragData?> candidateData,
        List<dynamic> rejectedData,
      ) {
        final bool highlighted = candidateData.isNotEmpty;
        return AnimatedContainer(
          key: ValueKey<String>('schedule-cell-$rowIndex-$columnIndex'),
          duration: const Duration(milliseconds: 120),
          height: _rowHeight,
          padding: const EdgeInsets.all(6),
          decoration: BoxDecoration(
            color: highlighted
                ? const Color(0xFFFFF0E6)
                : isToday
                    ? _SmartColors.todayCell
                    : Colors.white.withOpacity(.72),
            border: Border(
              right: BorderSide(
                color:
                    isLastColumn ? Colors.transparent : _SmartColors.lineSoft,
              ),
              bottom: BorderSide(
                color: isLastRow ? Colors.transparent : _SmartColors.lineSoft,
              ),
            ),
          ),
          child: Stack(
            children: <Widget>[
              if (highlighted)
                Positioned.fill(
                  child: DecoratedBox(
                    decoration: BoxDecoration(
                      border: Border.all(
                        color: _SmartColors.orange.withOpacity(.42),
                        width: 1.2,
                      ),
                      borderRadius: BorderRadius.circular(10),
                    ),
                  ),
                ),
              if (lesson == null)
                _EmptyScheduleSlot(
                  rowIndex: rowIndex,
                  columnIndex: columnIndex,
                  availability: availability,
                  dragAvailability: dragAvailability,
                  dragChecking: dragChecking,
                  dragValidationActive: dragValidationActive,
                  scheduleTargetSelected: scheduleTargetSelected,
                  checking: availabilityLoading && availability == null,
                  creating: creating,
                  onTap: () => onEmptySlotTap(rowIndex, columnIndex),
                )
              else
                _DraggableLessonBlock(
                  lesson: lesson!,
                  source: _LessonDragData(row: rowIndex, column: columnIndex),
                  onDragStarted: onLessonDragStarted,
                  onDragEnded: onLessonDragEnded,
                ),
            ],
          ),
        );
      },
    );
  }
}

class _EmptyScheduleSlot extends StatelessWidget {
  const _EmptyScheduleSlot({
    required this.rowIndex,
    required this.columnIndex,
    required this.availability,
    required this.dragAvailability,
    required this.dragChecking,
    required this.dragValidationActive,
    required this.scheduleTargetSelected,
    required this.checking,
    required this.creating,
    required this.onTap,
  });

  final int rowIndex;
  final int columnIndex;
  final _SlotAvailability? availability;
  final _SlotAvailability? dragAvailability;
  final bool dragChecking;
  final bool dragValidationActive;
  final bool scheduleTargetSelected;
  final bool checking;
  final bool creating;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final bool showingDragState = dragValidationActive;
    if (!scheduleTargetSelected && !showingDragState) {
      return const SizedBox.expand();
    }
    final _SlotAvailability? state =
        showingDragState ? dragAvailability : availability;
    final bool checkingNow =
        showingDragState ? dragChecking && state == null : checking;
    final bool valid = state?.valid == true;
    final bool invalid = state?.valid == false;
    final Color background = creating
        ? const Color(0xFFFFF1E8)
        : valid
            ? const Color(0xFFF4FAF4)
            : invalid
                ? const Color(0xFFFFECEC)
                : const Color(0xFFFFF8EE);
    final Color border = valid
        ? const Color(0xFFD9E7DB)
        : invalid
            ? const Color(0xFFF3B7B7)
            : const Color(0xFFF0DDC9);
    final Color foreground = valid
        ? const Color(0xFF5E846C)
        : invalid
            ? _SmartColors.danger
            : _SmartColors.orangeDeep;
    final String label = creating
        ? '排课中'
        : valid
            ? (showingDragState ? '可调课' : '空闲时段(可排)')
            : invalid
                ? _invalidLabel(state!, drag: showingDragState)
                : checkingNow
                    ? '检测中'
                    : '待检测';

    return Material(
      key: ValueKey<String>('empty-slot-$rowIndex-$columnIndex'),
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(9),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 160),
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: background,
            border: Border.all(color: border),
            borderRadius: BorderRadius.circular(9),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              if (creating || checkingNow) ...<Widget>[
                SizedBox(
                  width: 11,
                  height: 11,
                  child: CircularProgressIndicator(
                    strokeWidth: 1.5,
                    color: foreground,
                  ),
                ),
                const SizedBox(width: 6),
              ],
              Flexible(
                child: Text(
                  label,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: foreground,
                    fontSize: valid ? 11.5 : 11,
                    height: 1,
                    fontWeight: valid ? FontWeight.w700 : FontWeight.w700,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _invalidLabel(_SlotAvailability state, {required bool drag}) {
    if (drag) {
      if (state.conflictTypes.isEmpty) {
        return '不可调';
      }
      return '${state.conflictTypes.join('/')}冲突';
    }
    if (state.conflictTypes.isEmpty) {
      return '冲突(不可排)';
    }
    return '${state.conflictTypes.join('/')}冲突';
  }
}

class _DraggableLessonBlock extends StatefulWidget {
  const _DraggableLessonBlock({
    required this.lesson,
    required this.source,
    required this.onDragStarted,
    required this.onDragEnded,
  });

  final _LessonCell lesson;
  final _LessonDragData source;
  final ValueChanged<_LessonDragData> onDragStarted;
  final VoidCallback onDragEnded;

  @override
  State<_DraggableLessonBlock> createState() => _DraggableLessonBlockState();
}

class _DraggableLessonBlockState extends State<_DraggableLessonBlock> {
  Timer? _armingTimer;
  bool _arming = false;
  bool _dragging = false;
  Offset? _downPosition;

  @override
  void dispose() {
    _armingTimer?.cancel();
    super.dispose();
  }

  void _handlePointerDown(PointerDownEvent event) {
    _armingTimer?.cancel();
    _downPosition = event.position;
    _armingTimer = Timer(const Duration(milliseconds: 150), () {
      if (!mounted || _dragging) {
        return;
      }
      setState(() {
        _arming = true;
      });
      unawaited(HapticFeedback.selectionClick());
    });
  }

  void _handlePointerMove(PointerMoveEvent event) {
    final Offset? down = _downPosition;
    if (down == null || _dragging) {
      return;
    }
    if ((event.position - down).distance > 12) {
      _clearArming();
    }
  }

  void _clearArming() {
    _armingTimer?.cancel();
    _armingTimer = null;
    _downPosition = null;
    if (_arming && mounted) {
      setState(() {
        _arming = false;
      });
    }
  }

  void _handleDragStarted() {
    _armingTimer?.cancel();
    _downPosition = null;
    if (mounted) {
      setState(() {
        _arming = false;
        _dragging = true;
      });
    }
    widget.onDragStarted(widget.source);
    unawaited(HapticFeedback.mediumImpact());
  }

  void _handleDragEnded() {
    if (mounted) {
      setState(() {
        _arming = false;
        _dragging = false;
      });
    }
    widget.onDragEnded();
  }

  @override
  Widget build(BuildContext context) {
    final Widget child = _LessonBlock(
      widget.lesson,
      elevated: _arming || _dragging,
      armed: _arming,
    );
    return Listener(
      onPointerDown: _handlePointerDown,
      onPointerMove: _handlePointerMove,
      onPointerUp: (_) => _clearArming(),
      onPointerCancel: (_) => _clearArming(),
      child: LongPressDraggable<_LessonDragData>(
        key: ValueKey<String>(
          'lesson-${widget.source.row}-${widget.source.column}',
        ),
        data: widget.source,
        delay: const Duration(milliseconds: 300),
        hapticFeedbackOnStart: true,
        maxSimultaneousDrags: 1,
        dragAnchorStrategy: pointerDragAnchorStrategy,
        feedback: Material(
          color: Colors.transparent,
          child: Transform.translate(
            offset: const Offset(-78, -25),
            child: SizedBox(
              width: 156,
              height: 50,
              child:
                  _LessonBlock(widget.lesson, elevated: true, dragging: true),
            ),
          ),
        ),
        childWhenDragging: Opacity(
          opacity: .28,
          child: _LessonBlock(widget.lesson),
        ),
        onDragStarted: _handleDragStarted,
        onDragCompleted: _handleDragEnded,
        onDraggableCanceled: (_, __) => _handleDragEnded(),
        onDragEnd: (_) => _handleDragEnded(),
        child: child,
      ),
    );
  }
}

class _LessonBlock extends StatelessWidget {
  const _LessonBlock(
    this.lesson, {
    this.elevated = false,
    this.armed = false,
    this.dragging = false,
  });

  final _LessonCell lesson;
  final bool elevated;
  final bool armed;
  final bool dragging;

  @override
  Widget build(BuildContext context) {
    return AnimatedContainer(
      duration: const Duration(milliseconds: 140),
      height: 50,
      padding: const EdgeInsets.fromLTRB(9, 7, 9, 6),
      decoration: BoxDecoration(
        color: lesson.status.background,
        border: Border.all(
          color:
              armed || dragging ? _SmartColors.orangeDeep : Colors.transparent,
          width: armed || dragging ? 1.4 : 1,
        ),
        borderRadius: BorderRadius.circular(9),
        boxShadow: elevated
            ? const <BoxShadow>[
                BoxShadow(
                  color: Color(0x33B05F32),
                  blurRadius: 16,
                  offset: Offset(0, 8),
                ),
              ]
            : null,
      ),
      child: Stack(
        children: <Widget>[
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: <Widget>[
              Text(
                lesson.title,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _SmartColors.ink,
                  fontSize: 12,
                  height: 1.1,
                  fontWeight: FontWeight.w900,
                ),
              ),
              Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      lesson.person,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _SmartColors.text,
                        fontSize: 10,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    lesson.status.label,
                    style: TextStyle(
                      color: lesson.status.foreground,
                      fontSize: 10,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ],
              ),
            ],
          ),
          if (lesson.conflict)
            const Positioned(
              right: 0,
              top: 0,
              child: DecoratedBox(
                decoration: BoxDecoration(
                  color: _SmartColors.danger,
                  shape: BoxShape.circle,
                ),
                child: SizedBox(width: 7, height: 7),
              ),
            ),
        ],
      ),
    );
  }
}
