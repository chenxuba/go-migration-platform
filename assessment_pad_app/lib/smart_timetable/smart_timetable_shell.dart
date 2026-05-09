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
    required this.periodGroupDropdownOpen,
    required this.periodGroupDropdownLink,
    required this.schedulePanelOpen,
    required this.bootstrapLoading,
    required this.loading,
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
    required this.openFilterKind,
    required this.studentFilterLabel,
    required this.courseFilterLabel,
    required this.callStatusFilterLabel,
    required this.studentFilterOptions,
    required this.courseFilterOptions,
    required this.callStatusFilterOptions,
    required this.selectedStudentFilters,
    required this.selectedCourseFilters,
    required this.selectedCallStatusFilters,
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
    required this.isCurrentWeek,
    required this.onBack,
    required this.onPrevWeek,
    required this.onNextWeek,
    required this.onToday,
    required this.onPeriodGroupToggle,
    required this.onPeriodGroupSelected,
    required this.onPeriodGroupDropdownClose,
    required this.onTeacherToggle,
    required this.onTeacherSelected,
    required this.onTeacherDropdownClose,
    required this.onSchedulePanelToggle,
    required this.onSchedulePanelClose,
    required this.onFilterToggle,
    required this.onFilterClose,
    required this.onStudentFilterToggled,
    required this.onCourseFilterToggled,
    required this.onCallStatusFilterToggled,
    required this.onFilterCleared,
    required this.onScheduleModeChanged,
    required this.onScheduleTargetSelected,
    required this.onScheduleTargetCleared,
    required this.onAssistantToggled,
    required this.onClassroomSelected,
    required this.onScheduleOptionsRefresh,
    required this.onAvailabilityRefresh,
    required this.onRefresh,
    required this.onPrimaryScheduleTap,
    required this.onLessonTap,
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
  final bool periodGroupDropdownOpen;
  final LayerLink periodGroupDropdownLink;
  final bool schedulePanelOpen;
  final bool bootstrapLoading;
  final bool loading;
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
  final _TimetableFilterKind? openFilterKind;
  final String studentFilterLabel;
  final String courseFilterLabel;
  final String callStatusFilterLabel;
  final List<_TimetableFilterOption> studentFilterOptions;
  final List<_TimetableFilterOption> courseFilterOptions;
  final List<_TimetableFilterOption> callStatusFilterOptions;
  final Set<String> selectedStudentFilters;
  final Set<String> selectedCourseFilters;
  final Set<String> selectedCallStatusFilters;
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
  final bool isCurrentWeek;
  final VoidCallback onBack;
  final VoidCallback onPrevWeek;
  final VoidCallback onNextWeek;
  final VoidCallback onToday;
  final VoidCallback onPeriodGroupToggle;
  final ValueChanged<int> onPeriodGroupSelected;
  final VoidCallback onPeriodGroupDropdownClose;
  final VoidCallback onTeacherToggle;
  final ValueChanged<int> onTeacherSelected;
  final VoidCallback onTeacherDropdownClose;
  final VoidCallback onSchedulePanelToggle;
  final VoidCallback onSchedulePanelClose;
  final ValueChanged<_TimetableFilterKind> onFilterToggle;
  final VoidCallback onFilterClose;
  final ValueChanged<String> onStudentFilterToggled;
  final ValueChanged<String> onCourseFilterToggled;
  final ValueChanged<String> onCallStatusFilterToggled;
  final ValueChanged<_TimetableFilterKind> onFilterCleared;
  final ValueChanged<_ScheduleMode> onScheduleModeChanged;
  final ValueChanged<ScheduleTargetOption> onScheduleTargetSelected;
  final VoidCallback onScheduleTargetCleared;
  final ValueChanged<ScheduleStaffOption> onAssistantToggled;
  final ValueChanged<ScheduleClassroomOption?> onClassroomSelected;
  final VoidCallback onScheduleOptionsRefresh;
  final VoidCallback onAvailabilityRefresh;
  final VoidCallback onRefresh;
  final VoidCallback onPrimaryScheduleTap;
  final ValueChanged<_LessonCell> onLessonTap;
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
        final Widget pageBody = bootstrapLoading
            ? _TimetableBootstrapScaffold(
                compact: compact,
                primaryWidth: primaryWidth,
                teacherWidth: teacherWidth,
              )
            : loading
                ? _TimetableLoadingScaffold(
                    compact: compact,
                    primaryWidth: primaryWidth,
                    teacherWidth: teacherWidth,
                  )
                : Column(
                    children: <Widget>[
                      _TimetableTopBar(
                        compact: compact,
                        teacher: teacher,
                        teacherDropdownOpen: teacherDropdownOpen,
                        teacherWidth: teacherWidth,
                        primaryWidth: primaryWidth,
                        dateRange: dateRange,
                        isCurrentWeek: isCurrentWeek,
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
                        periodGroupDropdownOpen: periodGroupDropdownOpen,
                        periodGroupDropdownLink: periodGroupDropdownLink,
                        errorMessage: errorMessage,
                        openFilterKind: openFilterKind,
                        studentFilterLabel: studentFilterLabel,
                        courseFilterLabel: courseFilterLabel,
                        callStatusFilterLabel: callStatusFilterLabel,
                        onPeriodGroupToggle: onPeriodGroupToggle,
                        onSchedulePanelToggle: onSchedulePanelToggle,
                        onScheduleModeChanged: onScheduleModeChanged,
                        onScheduleTargetCleared: onScheduleTargetCleared,
                        onAvailabilityRefresh: onAvailabilityRefresh,
                        onRefresh: onRefresh,
                        onFilterToggle: onFilterToggle,
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
                          scheduleTargetSelected:
                              selectedScheduleTarget != null,
                          availabilityLoading: availabilityLoading,
                          slotAvailability: slotAvailability,
                          dragSlotAvailability: dragSlotAvailability,
                          dragCheckingSlotKeys: dragCheckingSlotKeys,
                          dragValidationActive: dragValidationActive,
                          creatingSlotKey: creatingSlotKey,
                          onLessonTap: onLessonTap,
                          onLessonMove: onLessonMove,
                          onLessonDragStarted: onLessonDragStarted,
                          onLessonDragEnded: onLessonDragEnded,
                          onEmptySlotTap: onEmptySlotTap,
                        ),
                      ),
                    ],
                  );

        return ColoredBox(
          color: _SmartColors.page,
          child: Stack(
            children: <Widget>[
              Padding(
                padding: EdgeInsets.fromLTRB(pagePadding, 24, pagePadding, 24),
                child: pageBody,
              ),
              if (!loading && teacherDropdownOpen) ...<Widget>[
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
              if (!loading && periodGroupDropdownOpen) ...<Widget>[
                Positioned.fill(
                  child: GestureDetector(
                    behavior: HitTestBehavior.translucent,
                    onTap: onPeriodGroupDropdownClose,
                  ),
                ),
                CompositedTransformFollower(
                  link: periodGroupDropdownLink,
                  showWhenUnlinked: false,
                  targetAnchor: Alignment.bottomLeft,
                  followerAnchor: Alignment.topLeft,
                  offset: const Offset(0, 8),
                  child: Material(
                    color: Colors.transparent,
                    child: SizedBox(
                      width: compact ? 188 : 200,
                      child: _PeriodGroupDropdownPanel(
                        groups: periodGroups,
                        selectedIndex: periodGroupIndex,
                        onSelected: onPeriodGroupSelected,
                      ),
                    ),
                  ),
                ),
              ],
              if (!loading && schedulePanelOpen) ...<Widget>[
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
              if (!loading && openFilterKind != null) ...<Widget>[
                Positioned.fill(
                  child: GestureDetector(
                    behavior: HitTestBehavior.translucent,
                    onTap: onFilterClose,
                  ),
                ),
                Positioned(
                  top: 134,
                  right: _filterPanelRightOffset(
                    kind: openFilterKind!,
                    compact: compact,
                    pagePadding: pagePadding,
                  ),
                  width: compact ? 220 : 236,
                  child: _TimetableFilterPanel(
                    kind: openFilterKind!,
                    options: switch (openFilterKind!) {
                      _TimetableFilterKind.student => studentFilterOptions,
                      _TimetableFilterKind.course => courseFilterOptions,
                      _TimetableFilterKind.callStatus =>
                        callStatusFilterOptions,
                    },
                    selectedIds: switch (openFilterKind!) {
                      _TimetableFilterKind.student => selectedStudentFilters,
                      _TimetableFilterKind.course => selectedCourseFilters,
                      _TimetableFilterKind.callStatus =>
                        selectedCallStatusFilters,
                    },
                    onOptionToggled: switch (openFilterKind!) {
                      _TimetableFilterKind.student => onStudentFilterToggled,
                      _TimetableFilterKind.course => onCourseFilterToggled,
                      _TimetableFilterKind.callStatus =>
                        onCallStatusFilterToggled,
                    },
                    onClear: () => onFilterCleared(openFilterKind!),
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

class _TimetableLoadingScaffold extends StatelessWidget {
  const _TimetableLoadingScaffold({
    required this.compact,
    required this.primaryWidth,
    required this.teacherWidth,
  });

  final bool compact;
  final double primaryWidth;
  final double teacherWidth;

  @override
  Widget build(BuildContext context) {
    final double periodGroupWidth = compact ? 126 : 136;
    return Column(
      key: const ValueKey<String>('smart-timetable-skeleton'),
      children: <Widget>[
        SizedBox(
          height: 56,
          child: Row(
            children: <Widget>[
              SizedBox(
                width: compact ? 188 : 210,
                child: Row(
                  children: <Widget>[
                    const _TimetableSkeletonBox(
                      width: 42,
                      height: 42,
                      radius: 12,
                    ),
                    const SizedBox(width: 14),
                    const Expanded(
                      child: _TimetableSkeletonBox(height: 30, radius: 11),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: <Widget>[
                    _TimetableSkeletonBox(
                      width: compact ? 302 : 352,
                      height: 42,
                    ),
                    const SizedBox(width: 10),
                    _TimetableSkeletonBox(
                      width: compact ? 74 : 82,
                      height: 42,
                    ),
                    const SizedBox(width: 10),
                    _TimetableSkeletonBox(
                      width: teacherWidth,
                      height: 42,
                    ),
                    const SizedBox(width: 10),
                    _TimetableSkeletonBox(
                      width: primaryWidth,
                      height: 42,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        SizedBox(
          height: 44,
          child: Row(
            children: <Widget>[
              Expanded(
                child: Row(
                  children: <Widget>[
                    Expanded(
                      child: _TimetableSkeletonBox(
                        height: 44,
                        radius: 14,
                      ),
                    ),
                    SizedBox(width: compact ? 8 : 10),
                    const _TimetableSkeletonBox(
                      width: 92,
                      height: 34,
                      radius: 11,
                    ),
                    const SizedBox(width: 8),
                    const _TimetableSkeletonBox(
                      width: 92,
                      height: 34,
                      radius: 11,
                    ),
                  ],
                ),
              ),
              SizedBox(width: compact ? 8 : 10),
              _TimetableSkeletonBox(
                width: periodGroupWidth,
                height: 34,
                radius: 11,
              ),
              SizedBox(width: compact ? 8 : 10),
              const _TimetableSkeletonBox(width: 88, height: 34, radius: 11),
              const SizedBox(width: 8),
              const _TimetableSkeletonBox(width: 88, height: 34, radius: 11),
            ],
          ),
        ),
        const SizedBox(height: 4),
        const _TimetableSummarySkeleton(),
        const SizedBox(height: 4),
        const Expanded(child: _TimetableBoardSkeleton()),
      ],
    );
  }
}

class _TimetableBootstrapScaffold extends StatelessWidget {
  const _TimetableBootstrapScaffold({
    required this.compact,
    required this.primaryWidth,
    required this.teacherWidth,
  });

  final bool compact;
  final double primaryWidth;
  final double teacherWidth;

  @override
  Widget build(BuildContext context) {
    final double panelRadius = compact ? 18 : 20;
    final double periodGroupWidth = compact ? 126 : 136;
    return Column(
      key: const ValueKey<String>('smart-timetable-skeleton'),
      children: <Widget>[
        SizedBox(
          height: 56,
          child: Row(
            children: <Widget>[
              SizedBox(
                width: compact ? 188 : 210,
                child: Row(
                  children: <Widget>[
                    const _IconShell(
                      size: 42,
                      icon: Icons.chevron_left_rounded,
                    ),
                    const SizedBox(width: 14),
                    const Expanded(
                      child: Align(
                        alignment: Alignment.centerLeft,
                        child: Text(
                          '排课日程',
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            color: _SmartColors.ink,
                            fontSize: 25,
                            height: 1.05,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: <Widget>[
                    _DateSwitch(
                      width: compact ? 302 : 352,
                      dateRange: _dateRangeText(0),
                      isCurrentWeek: true,
                      onPrev: _noop,
                      onNext: _noop,
                    ),
                    const SizedBox(width: 10),
                    _ToolbarButton(
                      width: compact ? 74 : 82,
                      icon: Icons.format_list_bulleted_rounded,
                      label: '今天',
                      onTap: _noop,
                    ),
                    const SizedBox(width: 10),
                    _TimetableSkeletonBox(width: teacherWidth, height: 42),
                    const SizedBox(width: 10),
                    _PrimaryButton(
                      width: primaryWidth,
                      onTap: _noop,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 10),
        SizedBox(
          height: 44,
          child: Row(
            children: <Widget>[
              Expanded(
                child: _ScheduleComposerBar(
                  compact: compact,
                  mode: _ScheduleMode.oneToOne,
                  selectedTarget: null,
                  panelOpen: false,
                  availabilityLoading: false,
                  availabilityMessage: null,
                  onPanelToggle: _noop,
                  onModeChanged: _noopScheduleModeChanged,
                  onTargetCleared: _noop,
                  onAvailabilityRefresh: _noop,
                ),
              ),
              SizedBox(width: compact ? 8 : 10),
              SizedBox(
                width: periodGroupWidth,
                child: const _TimetableFieldSkeleton(
                  icon: Icons.view_week_outlined,
                  lineWidth: 52,
                ),
              ),
              const SizedBox(width: 8),
              _FilterButton(
                icon: Icons.person_outline_rounded,
                label: '上课学员',
                width: compact ? 118 : 130,
                onTap: _noop,
              ),
              const SizedBox(width: 8),
              _FilterButton(
                icon: Icons.menu_book_outlined,
                label: '全部课程',
                width: compact ? 118 : 130,
                onTap: _noop,
              ),
              const SizedBox(width: 8),
              _FilterButton(
                icon: Icons.fact_check_outlined,
                label: '点名状态',
                width: compact ? 110 : 122,
                onTap: _noop,
              ),
            ],
          ),
        ),
        const SizedBox(height: 4),
        _TimetableSummary(
          compact: compact,
          summary: const TimetableSummary(),
        ),
        const SizedBox(height: 4),
        Expanded(
          child: DecoratedBox(
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(.90),
              borderRadius: BorderRadius.circular(panelRadius),
              border: Border.all(color: _SmartColors.line),
              boxShadow: const <BoxShadow>[
                BoxShadow(
                  color: Color(0x12C07A4A),
                  blurRadius: 22,
                  offset: Offset(0, 10),
                ),
              ],
            ),
            child: const Padding(
              padding: EdgeInsets.fromLTRB(16, 14, 16, 16),
              child: _TimetableBoardSkeleton(),
            ),
          ),
        ),
      ],
    );
  }
}

void _noop() {}

void _noopScheduleModeChanged(_ScheduleMode _) {}

class _TimetableFieldSkeleton extends StatelessWidget {
  const _TimetableFieldSkeleton({
    required this.icon,
    required this.lineWidth,
  });

  final IconData icon;
  final double lineWidth;

  @override
  Widget build(BuildContext context) {
    return _ShellBox(
      height: 34,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      borderRadius: 11,
      child: Row(
        children: <Widget>[
          Icon(icon, color: _SmartColors.text, size: 16),
          const SizedBox(width: 7),
          Expanded(
            child: Align(
              alignment: Alignment.centerLeft,
              child: _TimetableSkeletonBox(
                width: lineWidth,
                height: 10,
                radius: 6,
              ),
            ),
          ),
          const SizedBox(width: 4),
          const Icon(
            Icons.keyboard_arrow_down_rounded,
            color: _SmartColors.text,
            size: 18,
          ),
        ],
      ),
    );
  }
}

class _TimetableSummarySkeleton extends StatelessWidget {
  const _TimetableSummarySkeleton();

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 38,
      child: Row(
        children: const <Widget>[
          SizedBox(width: 12),
          _SummaryAccent(),
          SizedBox(width: 11),
          _TimetableSkeletonBox(width: 218, height: 16, radius: 7),
          Spacer(),
          _TimetableSkeletonBox(width: 64, height: 12, radius: 6),
          SizedBox(width: 14),
          _TimetableSkeletonBox(width: 64, height: 12, radius: 6),
          SizedBox(width: 14),
          _TimetableSkeletonBox(width: 64, height: 12, radius: 6),
          SizedBox(width: 14),
          _TimetableSkeletonBox(width: 64, height: 12, radius: 6),
          SizedBox(width: 14),
          _TimetableSkeletonBox(width: 64, height: 12, radius: 6),
          SizedBox(width: 2),
        ],
      ),
    );
  }
}

double _filterPanelRightOffset({
  required _TimetableFilterKind kind,
  required bool compact,
  required double pagePadding,
}) {
  final double courseWidth = compact ? 118 : 130;
  final double statusWidth = compact ? 110 : 122;
  const double gap = 8;
  return switch (kind) {
    _TimetableFilterKind.student =>
      pagePadding + courseWidth + gap + statusWidth + gap,
    _TimetableFilterKind.course => pagePadding + statusWidth + gap,
    _TimetableFilterKind.callStatus => pagePadding,
  };
}
