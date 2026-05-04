part of '../smart_timetable_page.dart';

class _TimetableTopBar extends StatelessWidget {
  const _TimetableTopBar({
    required this.compact,
    required this.teacher,
    required this.teacherDropdownOpen,
    required this.teacherWidth,
    required this.primaryWidth,
    required this.dateRange,
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
    required this.onSelected,
    required this.compact,
  });

  final List<_PeriodGroupOption> groups;
  final int selectedIndex;
  final ValueChanged<int> onSelected;
  final bool compact;

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
    return SizedBox(
      height: 34,
      child: Row(
        children: <Widget>[
          Expanded(
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              physics: const BouncingScrollPhysics(),
              itemBuilder: (BuildContext context, int index) {
                final _PeriodGroupOption group = displayGroups[index];
                return _PeriodGroupTab(
                  key: ValueKey<String>('period-group-tab-${group.id}'),
                  group: group,
                  selected: index == selectedIndex,
                  compact: compact,
                  onTap: () => onSelected(index),
                );
              },
              separatorBuilder: (_, __) => SizedBox(width: compact ? 6 : 8),
              itemCount: displayGroups.length,
            ),
          ),
        ],
      ),
    );
  }
}

class _PeriodGroupTab extends StatelessWidget {
  const _PeriodGroupTab({
    required this.group,
    required this.selected,
    required this.compact,
    required this.onTap,
    super.key,
  });

  final _PeriodGroupOption group;
  final bool selected;
  final bool compact;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(11),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(11),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 160),
          width: compact ? 74 : 82,
          height: 34,
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFF0E5) : _SmartColors.card,
            border: Border.all(
              color: selected ? _SmartColors.orange : _SmartColors.line,
            ),
            borderRadius: BorderRadius.circular(11),
            boxShadow: selected
                ? const <BoxShadow>[
                    BoxShadow(
                      color: Color(0x24E96F43),
                      blurRadius: 14,
                      offset: Offset(0, 6),
                    ),
                  ]
                : null,
          ),
          child: Text(
            group.name,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              color: _SmartColors.text,
              fontSize: 12,
              height: 1,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
      ),
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
    required this.errorMessage,
    required this.onSchedulePanelToggle,
    required this.onScheduleModeChanged,
    required this.onScheduleTargetCleared,
    required this.onAvailabilityRefresh,
    required this.onPeriodGroupSelected,
    required this.onRefresh,
  });

  final bool compact;
  final _ScheduleMode scheduleMode;
  final ScheduleTargetOption? selectedScheduleTarget;
  final bool schedulePanelOpen;
  final bool availabilityLoading;
  final String? availabilityMessage;
  final List<_PeriodGroupOption> periodGroups;
  final int periodGroupIndex;
  final String? errorMessage;
  final VoidCallback onSchedulePanelToggle;
  final ValueChanged<_ScheduleMode> onScheduleModeChanged;
  final VoidCallback onScheduleTargetCleared;
  final VoidCallback onAvailabilityRefresh;
  final ValueChanged<int> onPeriodGroupSelected;
  final VoidCallback onRefresh;

  @override
  Widget build(BuildContext context) {
    final double periodGroupWidth = compact ? 250 : 284;
    return SizedBox(
      height: 44,
      child: Row(
        children: <Widget>[
          Expanded(
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
              compact: compact,
              groups: periodGroups,
              selectedIndex: periodGroupIndex,
              onSelected: onPeriodGroupSelected,
            ),
          ),
          SizedBox(width: compact ? 8 : 10),
          if (errorMessage != null)
            _TimetableLoadStatus(message: errorMessage!, onRefresh: onRefresh),
          if (errorMessage != null) const SizedBox(width: 10),
          const _FilterButton(icon: Icons.filter_list_rounded, label: '全部课程'),
          const SizedBox(width: 8),
          const _FilterButton(
            icon: Icons.library_books_outlined,
            label: '全部状态',
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
