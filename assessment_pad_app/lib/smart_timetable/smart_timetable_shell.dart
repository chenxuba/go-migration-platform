part of '../smart_timetable_page.dart';

const SystemUiOverlayStyle _smartImmersiveOverlayStyle = SystemUiOverlayStyle(
  statusBarColor: Colors.transparent,
  statusBarIconBrightness: Brightness.dark,
  statusBarBrightness: Brightness.light,
  systemStatusBarContrastEnforced: false,
  systemNavigationBarColor: Colors.transparent,
  systemNavigationBarDividerColor: Colors.transparent,
  systemNavigationBarIconBrightness: Brightness.dark,
  systemNavigationBarContrastEnforced: false,
);

class _SmartTimetableViewport extends StatelessWidget {
  const _SmartTimetableViewport({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return AnnotatedRegion<SystemUiOverlayStyle>(
      value: _smartImmersiveOverlayStyle,
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final Size screenSize = MediaQuery.sizeOf(context);
          final double viewportWidth = constraints.maxWidth.isFinite
              ? constraints.maxWidth
              : screenSize.width;
          final double viewportHeight = constraints.maxHeight.isFinite
              ? constraints.maxHeight
              : screenSize.height;
          final double aspect = viewportWidth / viewportHeight;
          final double designWidth =
              math.max(_smartMinDesignWidth, _smartDesignHeight * aspect);

          return ColoredBox(
            color: _SmartColors.page,
            child: Center(
              child: FittedBox(
                fit: BoxFit.contain,
                child: SizedBox(
                  width: designWidth,
                  height: _smartDesignHeight,
                  child: child,
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}

class _SmartTimetableScreen extends StatelessWidget {
  const _SmartTimetableScreen({
    required this.teacher,
    required this.periodGroups,
    required this.periodGroupIndex,
    required this.teachers,
    required this.teacherIndex,
    required this.teacherDropdownOpen,
    required this.schedulePanelOpen,
    required this.scheduleMode,
    required this.selectedScheduleTarget,
    required this.oneToOneTargets,
    required this.groupClassTargets,
    required this.assistantOptions,
    required this.classroomOptions,
    required this.selectedAssistantIds,
    required this.selectedClassroom,
    required this.scheduleOptionsLoading,
    required this.scheduleOptionsError,
    required this.availabilityLoading,
    required this.availabilityMessage,
    required this.slotAvailability,
    required this.dragSlotAvailability,
    required this.dragCheckingSlotKeys,
    required this.dragValidationActive,
    required this.creatingSlotKey,
    required this.scheduleRows,
    required this.weekDays,
    required this.timeSlots,
    required this.summary,
    required this.errorMessage,
    required this.dateRange,
    required this.onBack,
    required this.onPrevWeek,
    required this.onNextWeek,
    required this.onToday,
    required this.onPeriodGroupSelected,
    required this.onTeacherToggle,
    required this.onTeacherSelected,
    required this.onTeacherDropdownClose,
    required this.onSchedulePanelToggle,
    required this.onSchedulePanelClose,
    required this.onScheduleModeChanged,
    required this.onScheduleTargetSelected,
    required this.onScheduleTargetCleared,
    required this.onAssistantToggled,
    required this.onClassroomSelected,
    required this.onScheduleOptionsRefresh,
    required this.onAvailabilityRefresh,
    required this.onRefresh,
    required this.onPrimaryScheduleTap,
    required this.onLessonMove,
    required this.onLessonDragStarted,
    required this.onLessonDragEnded,
    required this.onEmptySlotTap,
  });

  final _TeacherOption teacher;
  final List<_PeriodGroupOption> periodGroups;
  final int periodGroupIndex;
  final List<_TeacherOption> teachers;
  final int teacherIndex;
  final bool teacherDropdownOpen;
  final bool schedulePanelOpen;
  final _ScheduleMode scheduleMode;
  final ScheduleTargetOption? selectedScheduleTarget;
  final List<ScheduleTargetOption> oneToOneTargets;
  final List<ScheduleTargetOption> groupClassTargets;
  final List<ScheduleStaffOption> assistantOptions;
  final List<ScheduleClassroomOption> classroomOptions;
  final Set<String> selectedAssistantIds;
  final ScheduleClassroomOption? selectedClassroom;
  final bool scheduleOptionsLoading;
  final String? scheduleOptionsError;
  final bool availabilityLoading;
  final String? availabilityMessage;
  final Map<String, _SlotAvailability> slotAvailability;
  final Map<String, _SlotAvailability> dragSlotAvailability;
  final Set<String> dragCheckingSlotKeys;
  final bool dragValidationActive;
  final String? creatingSlotKey;
  final List<List<_LessonCell?>> scheduleRows;
  final List<_WeekDay> weekDays;
  final List<_TimeSlot> timeSlots;
  final TimetableSummary summary;
  final String? errorMessage;
  final String dateRange;
  final VoidCallback onBack;
  final VoidCallback onPrevWeek;
  final VoidCallback onNextWeek;
  final VoidCallback onToday;
  final ValueChanged<int> onPeriodGroupSelected;
  final VoidCallback onTeacherToggle;
  final ValueChanged<int> onTeacherSelected;
  final VoidCallback onTeacherDropdownClose;
  final VoidCallback onSchedulePanelToggle;
  final VoidCallback onSchedulePanelClose;
  final ValueChanged<_ScheduleMode> onScheduleModeChanged;
  final ValueChanged<ScheduleTargetOption> onScheduleTargetSelected;
  final VoidCallback onScheduleTargetCleared;
  final ValueChanged<ScheduleStaffOption> onAssistantToggled;
  final ValueChanged<ScheduleClassroomOption?> onClassroomSelected;
  final VoidCallback onScheduleOptionsRefresh;
  final VoidCallback onAvailabilityRefresh;
  final VoidCallback onRefresh;
  final VoidCallback onPrimaryScheduleTap;
  final void Function(_LessonDragData source, int targetRow, int targetColumn)
      onLessonMove;
  final ValueChanged<_LessonDragData> onLessonDragStarted;
  final VoidCallback onLessonDragEnded;
  final void Function(int row, int column) onEmptySlotTap;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width = constraints.maxWidth.isFinite
            ? constraints.maxWidth
            : _smartWideDesignWidth;
        final bool compact = width < 1180;
        final double pagePadding = compact ? 24 : 32;

        final double teacherWidth = compact ? 190 : 224;
        final double primaryWidth = compact ? 104 : 110;

        return ColoredBox(
          color: _SmartColors.page,
          child: Stack(
            children: <Widget>[
              Padding(
                padding: EdgeInsets.fromLTRB(pagePadding, 24, pagePadding, 24),
                child: Column(
                  children: <Widget>[
                    _TimetableTopBar(
                      compact: compact,
                      teacher: teacher,
                      teacherDropdownOpen: teacherDropdownOpen,
                      teacherWidth: teacherWidth,
                      primaryWidth: primaryWidth,
                      dateRange: dateRange,
                      onBack: onBack,
                      onPrevWeek: onPrevWeek,
                      onNextWeek: onNextWeek,
                      onToday: onToday,
                      onTeacherToggle: onTeacherToggle,
                      onPrimaryScheduleTap: onPrimaryScheduleTap,
                    ),
                    const SizedBox(height: 10),
                    _TimetableSubBar(
                      compact: compact,
                      scheduleMode: scheduleMode,
                      selectedScheduleTarget: selectedScheduleTarget,
                      schedulePanelOpen: schedulePanelOpen,
                      availabilityLoading: availabilityLoading,
                      availabilityMessage: availabilityMessage,
                      periodGroups: periodGroups,
                      periodGroupIndex: periodGroupIndex,
                      errorMessage: errorMessage,
                      onSchedulePanelToggle: onSchedulePanelToggle,
                      onScheduleModeChanged: onScheduleModeChanged,
                      onScheduleTargetCleared: onScheduleTargetCleared,
                      onAvailabilityRefresh: onAvailabilityRefresh,
                      onPeriodGroupSelected: onPeriodGroupSelected,
                      onRefresh: onRefresh,
                    ),
                    const SizedBox(height: 4),
                    _TimetableSummary(compact: compact, summary: summary),
                    const SizedBox(height: 4),
                    Expanded(
                      child: _TimetableBoard(
                        compact: compact,
                        rows: scheduleRows,
                        weekDays: weekDays,
                        timeSlots: timeSlots,
                        selectedTeacherId: teacher.id,
                        scheduleTargetSelected: selectedScheduleTarget != null,
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
              if (teacherDropdownOpen) ...<Widget>[
                Positioned.fill(
                  child: GestureDetector(
                    behavior: HitTestBehavior.translucent,
                    onTap: onTeacherDropdownClose,
                  ),
                ),
                Positioned(
                  top: 82,
                  right: pagePadding + primaryWidth + 10,
                  width: teacherWidth,
                  child: _TeacherDropdownPanel(
                    teachers: teachers,
                    selectedIndex: teacherIndex,
                    onSelected: onTeacherSelected,
                  ),
                ),
              ],
              if (schedulePanelOpen) ...<Widget>[
                Positioned.fill(
                  child: GestureDetector(
                    behavior: HitTestBehavior.translucent,
                    onTap: onSchedulePanelClose,
                  ),
                ),
                Positioned(
                  top: 134,
                  left: pagePadding,
                  width: compact ? 690 : 780,
                  child: _SchedulePickerPanel(
                    mode: scheduleMode,
                    oneToOneTargets: oneToOneTargets,
                    groupClassTargets: groupClassTargets,
                    assistantOptions: assistantOptions,
                    classroomOptions: classroomOptions,
                    selectedTarget: selectedScheduleTarget,
                    selectedAssistantIds: selectedAssistantIds,
                    selectedClassroom: selectedClassroom,
                    loading: scheduleOptionsLoading,
                    errorMessage: scheduleOptionsError,
                    onModeChanged: onScheduleModeChanged,
                    onTargetSelected: onScheduleTargetSelected,
                    onAssistantToggled: onAssistantToggled,
                    onClassroomSelected: onClassroomSelected,
                    onRefresh: onScheduleOptionsRefresh,
                    onClose: onSchedulePanelClose,
                  ),
                ),
              ],
            ],
          ),
        );
      },
    );
  }
}
