part of '../smart_timetable_page.dart';

class _TimetableTopBar extends StatelessWidget {
  const _TimetableTopBar({
    required this.compact,
    required this.teacher,
    required this.teacherDropdownOpen,
    required this.teacherWidth,
    required this.primaryWidth,
    required this.dateRange,
    required this.isCurrentWeek,
    required this.onBack,
    required this.onPrevWeek,
    required this.onNextWeek,
    required this.onToday,
    required this.onTeacherToggle,
    required this.onPrimaryScheduleTap,
  });

  final bool compact;
  final _TeacherOption teacher;
  final bool teacherDropdownOpen;
  final double teacherWidth;
  final double primaryWidth;
  final String dateRange;
  final bool isCurrentWeek;
  final VoidCallback onBack;
  final VoidCallback onPrevWeek;
  final VoidCallback onNextWeek;
  final VoidCallback onToday;
  final VoidCallback onTeacherToggle;
  final VoidCallback onPrimaryScheduleTap;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 56,
      child: Row(
        children: <Widget>[
          SizedBox(
            width: compact ? 188 : 210,
            child: Row(
              children: <Widget>[
                _IconShell(
                  size: 42,
                  icon: Icons.chevron_left_rounded,
                  onTap: onBack,
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
                  dateRange: dateRange,
                  isCurrentWeek: isCurrentWeek,
                  onPrev: onPrevWeek,
                  onNext: onNextWeek,
                ),
                const SizedBox(width: 10),
                _ToolbarButton(
                  width: compact ? 74 : 82,
                  icon: Icons.format_list_bulleted_rounded,
                  label: '今天',
                  onTap: onToday,
                ),
                const SizedBox(width: 10),
                SizedBox(
                  width: teacherWidth,
                  child: _TeacherSelector(
                    teacher: teacher,
                    isOpen: teacherDropdownOpen,
                    onTap: onTeacherToggle,
                  ),
                ),
                const SizedBox(width: 10),
                _PrimaryButton(
                  width: primaryWidth,
                  onTap: onPrimaryScheduleTap,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _PeriodGroupTabs extends StatelessWidget {
  const _PeriodGroupTabs({
    required this.groups,
    required this.selectedIndex,
    required this.isOpen,
    required this.layerLink,
    required this.onToggle,
  });

  final List<_PeriodGroupOption> groups;
  final int selectedIndex;
  final bool isOpen;
  final LayerLink layerLink;
  final VoidCallback onToggle;

  @override
  Widget build(BuildContext context) {
    final List<_PeriodGroupOption> displayGroups = groups.isEmpty
        ? const <_PeriodGroupOption>[
            _PeriodGroupOption(
              id: 'default',
              name: '默认时段',
              meta: '08:00 - 18:20 · 11节',
            ),
          ]
        : groups;
    final int safeIndex = selectedIndex.clamp(0, displayGroups.length - 1);
    final _PeriodGroupOption current = displayGroups[safeIndex];
    return _PeriodGroupDropdownButton(
      group: current,
      open: isOpen,
      layerLink: layerLink,
      onTap: onToggle,
    );
  }
}

class _TimetableSubBar extends StatelessWidget {
  const _TimetableSubBar({
    required this.compact,
    required this.scheduleMode,
    required this.selectedScheduleTarget,
    required this.schedulePanelOpen,
    required this.availabilityLoading,
    required this.availabilityMessage,
    required this.periodGroups,
    required this.periodGroupIndex,
    required this.periodGroupDropdownOpen,
    required this.periodGroupDropdownLink,
    required this.errorMessage,
    required this.openFilterKind,
    required this.studentFilterLabel,
    required this.courseFilterLabel,
    required this.callStatusFilterLabel,
    required this.onPeriodGroupToggle,
    required this.onSchedulePanelToggle,
    required this.onScheduleModeChanged,
    required this.onScheduleTargetCleared,
    required this.onAvailabilityRefresh,
    required this.onRefresh,
    required this.onFilterToggle,
  });

  final bool compact;
  final _ScheduleMode scheduleMode;
  final ScheduleTargetOption? selectedScheduleTarget;
  final bool schedulePanelOpen;
  final bool availabilityLoading;
  final String? availabilityMessage;
  final List<_PeriodGroupOption> periodGroups;
  final int periodGroupIndex;
  final bool periodGroupDropdownOpen;
  final LayerLink periodGroupDropdownLink;
  final String? errorMessage;
  final _TimetableFilterKind? openFilterKind;
  final String studentFilterLabel;
  final String courseFilterLabel;
  final String callStatusFilterLabel;
  final VoidCallback onPeriodGroupToggle;
  final VoidCallback onSchedulePanelToggle;
  final ValueChanged<_ScheduleMode> onScheduleModeChanged;
  final VoidCallback onScheduleTargetCleared;
  final VoidCallback onAvailabilityRefresh;
  final VoidCallback onRefresh;
  final ValueChanged<_TimetableFilterKind> onFilterToggle;

  @override
  Widget build(BuildContext context) {
    final double periodGroupWidth = compact ? 126 : 136;
    final double statusWidth = compact ? 96 : 128;
    return SizedBox(
      height: 44,
      child: Row(
        children: <Widget>[
          Flexible(
            fit: FlexFit.loose,
            child: _ScheduleComposerBar(
              compact: compact,
              mode: scheduleMode,
              selectedTarget: selectedScheduleTarget,
              panelOpen: schedulePanelOpen,
              availabilityLoading: availabilityLoading,
              availabilityMessage: availabilityMessage,
              onPanelToggle: onSchedulePanelToggle,
              onModeChanged: onScheduleModeChanged,
              onTargetCleared: onScheduleTargetCleared,
              onAvailabilityRefresh: onAvailabilityRefresh,
            ),
          ),
          SizedBox(width: compact ? 8 : 10),
          SizedBox(
            width: periodGroupWidth,
            child: _PeriodGroupTabs(
              groups: periodGroups,
              selectedIndex: periodGroupIndex,
              isOpen: periodGroupDropdownOpen,
              layerLink: periodGroupDropdownLink,
              onToggle: onPeriodGroupToggle,
            ),
          ),
          SizedBox(width: compact ? 8 : 10),
          if (errorMessage != null)
            SizedBox(
              width: statusWidth,
              child: _TimetableLoadStatus(
                message: errorMessage!,
                onRefresh: onRefresh,
              ),
            ),
          if (errorMessage != null) const SizedBox(width: 10),
          _FilterButton(
            key: const ValueKey<String>('smart-filter-student'),
            icon: Icons.person_outline_rounded,
            label: studentFilterLabel,
            width: compact ? 118 : 130,
            selected: openFilterKind == _TimetableFilterKind.student,
            onTap: () => onFilterToggle(_TimetableFilterKind.student),
          ),
          const SizedBox(width: 8),
          _FilterButton(
            key: const ValueKey<String>('smart-filter-course'),
            icon: Icons.menu_book_outlined,
            label: courseFilterLabel,
            width: compact ? 118 : 130,
            selected: openFilterKind == _TimetableFilterKind.course,
            onTap: () => onFilterToggle(_TimetableFilterKind.course),
          ),
          const SizedBox(width: 8),
          _FilterButton(
            key: const ValueKey<String>('smart-filter-call-status'),
            icon: Icons.fact_check_outlined,
            label: callStatusFilterLabel,
            width: compact ? 110 : 122,
            selected: openFilterKind == _TimetableFilterKind.callStatus,
            onTap: () => onFilterToggle(_TimetableFilterKind.callStatus),
          ),
        ],
      ),
    );
  }
}

class _TimetableSummary extends StatelessWidget {
  const _TimetableSummary({required this.compact, required this.summary});

  final bool compact;
  final TimetableSummary summary;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 38,
      child: Row(
        children: <Widget>[
          SizedBox(width: compact ? 6 : 12),
          const _SummaryAccent(),
          const SizedBox(width: 11),
          Text.rich(
            TextSpan(
              children: <InlineSpan>[
                const TextSpan(text: '共 '),
                TextSpan(
                  text: '${summary.total}',
                  style: const TextStyle(color: _SmartColors.orangeDeep),
                ),
                const TextSpan(text: ' 个日程，未点名 '),
                TextSpan(
                  text: '${summary.unsigned}',
                  style: const TextStyle(color: _SmartColors.orangeDeep),
                ),
                const TextSpan(text: ' 个，冲突 '),
                TextSpan(
                  text: '${summary.conflict}',
                  style: const TextStyle(color: _SmartColors.orangeDeep),
                ),
                const TextSpan(text: ' 个'),
              ],
            ),
            style: const TextStyle(
              color: _SmartColors.ink,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const Spacer(),
          const _LegendItem(color: _SmartColors.blue, label: '未点名'),
          const SizedBox(width: 14),
          const _LegendItem(color: _SmartColors.gray, label: '已点名'),
          const SizedBox(width: 14),
          const _LegendItem(color: _SmartColors.amber, label: '部分点名'),
          const SizedBox(width: 14),
          const _LegendItem(color: _SmartColors.green, label: '试听'),
          const SizedBox(width: 14),
          const _LegendItem(color: _SmartColors.danger, label: '冲突'),
          const SizedBox(width: 2),
        ],
      ),
    );
  }
}
