part of '../smart_timetable_page.dart';

class _TeacherSelector extends StatelessWidget {
  const _TeacherSelector({
    required this.teacher,
    required this.isOpen,
    required this.onTap,
  });

  final _TeacherOption teacher;
  final bool isOpen;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: _ShellBox(
        height: 42,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        child: Row(
          children: <Widget>[
            const Icon(
              Icons.person_outline_rounded,
              color: _SmartColors.ink,
              size: 18,
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    teacher.label,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.muted,
                      fontSize: 10,
                      height: 1.1,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    teacher.name,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.ink,
                      fontSize: 14,
                      height: 1.05,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 6),
            AnimatedRotation(
              turns: isOpen ? .5 : 0,
              duration: const Duration(milliseconds: 160),
              child: const Icon(
                Icons.keyboard_arrow_down_rounded,
                color: _SmartColors.ink,
                size: 18,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _TeacherDropdownPanel extends StatelessWidget {
  const _TeacherDropdownPanel({
    required this.teachers,
    required this.selectedIndex,
    required this.onSelected,
  });

  final List<_TeacherOption> teachers;
  final int selectedIndex;
  final ValueChanged<int> onSelected;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: Container(
        padding: const EdgeInsets.fromLTRB(10, 10, 10, 8),
        decoration: BoxDecoration(
          color: _SmartColors.card,
          border: Border.all(color: _SmartColors.line),
          borderRadius: BorderRadius.circular(16),
          boxShadow: const <BoxShadow>[
            BoxShadow(
              color: Color(0x1AB05F32),
              blurRadius: 24,
              offset: Offset(0, 14),
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            const Padding(
              padding: EdgeInsets.fromLTRB(8, 2, 8, 8),
              child: Text(
                '切换老师课表',
                style: TextStyle(
                  color: _SmartColors.ink,
                  fontSize: 13,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            if (teachers.isEmpty)
              Container(
                height: 44,
                padding: const EdgeInsets.symmetric(horizontal: 9),
                alignment: Alignment.centerLeft,
                child: const Text(
                  '该时段组暂无老师',
                  style: TextStyle(
                    color: _SmartColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              )
            else
              for (int index = 0; index < teachers.length; index += 1)
                _TeacherDropdownItem(
                  teacher: teachers[index],
                  selected: index == selectedIndex,
                  onTap: () => onSelected(index),
                ),
          ],
        ),
      ),
    );
  }
}

class _TeacherDropdownItem extends StatelessWidget {
  const _TeacherDropdownItem({
    required this.teacher,
    required this.selected,
    required this.onTap,
  });

  final _TeacherOption teacher;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        height: 44,
        padding: const EdgeInsets.symmetric(horizontal: 9),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFF1E8) : Colors.transparent,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: <Widget>[
            Container(
              width: 26,
              height: 26,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: selected ? _SmartColors.orange : const Color(0xFFFFF7EE),
                borderRadius: BorderRadius.circular(9),
              ),
              child: Icon(
                selected ? Icons.check_rounded : Icons.person_outline_rounded,
                color: selected ? Colors.white : _SmartColors.text,
                size: 16,
              ),
            ),
            const SizedBox(width: 9),
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    teacher.name,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.ink,
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    teacher.label,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.muted,
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _DateSwitch extends StatelessWidget {
  const _DateSwitch({
    required this.width,
    required this.dateRange,
    required this.onPrev,
    required this.onNext,
  });

  final double width;
  final String dateRange;
  final VoidCallback onPrev;
  final VoidCallback onNext;

  @override
  Widget build(BuildContext context) {
    return _ShellBox(
      width: width,
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 6),
      child: Row(
        children: <Widget>[
          _MiniNavButton(
            label: '上一周',
            onTap: onPrev,
          ),
          Expanded(
            child: Text(
              dateRange,
              textAlign: TextAlign.center,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _SmartColors.ink,
                fontSize: 14,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          _MiniNavButton(
            label: '下一周',
            onTap: onNext,
          ),
        ],
      ),
    );
  }
}

class _MiniNavButton extends StatelessWidget {
  const _MiniNavButton({
    required this.label,
    required this.onTap,
  });

  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Container(
        width: 74,
        height: 32,
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: const Color(0xFFFFF0E5),
          border: Border.all(color: const Color(0xFFF3D5C4)),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Text(
          label,
          maxLines: 1,
          overflow: TextOverflow.clip,
          style: const TextStyle(
            color: _SmartColors.orangeDeep,
            fontSize: 12,
            height: 1,
            fontWeight: FontWeight.w900,
          ),
        ),
      ),
    );
  }
}

class _ToolbarButton extends StatelessWidget {
  const _ToolbarButton({
    required this.width,
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final double width;
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return _ShellBox(
      width: width,
      height: 42,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            Icon(icon, color: _SmartColors.text, size: 16),
            const SizedBox(width: 6),
            Text(
              label,
              style: const TextStyle(
                color: _SmartColors.text,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _PrimaryButton extends StatelessWidget {
  const _PrimaryButton({required this.width, required this.onTap});

  final double width;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: _SmartColors.orange,
      borderRadius: BorderRadius.circular(13),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: SizedBox(
          width: width,
          height: 42,
          child: const Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              Icon(Icons.add_rounded, color: Colors.white, size: 18),
              SizedBox(width: 5),
              Text(
                '新增排课',
                style: TextStyle(
                  color: Colors.white,
                  fontSize: 13,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ScheduleComposerBar extends StatelessWidget {
  const _ScheduleComposerBar({
    required this.compact,
    required this.mode,
    required this.selectedTarget,
    required this.panelOpen,
    required this.availabilityLoading,
    required this.availabilityMessage,
    required this.onPanelToggle,
    required this.onModeChanged,
    required this.onTargetCleared,
    required this.onAvailabilityRefresh,
  });

  final bool compact;
  final _ScheduleMode mode;
  final ScheduleTargetOption? selectedTarget;
  final bool panelOpen;
  final bool availabilityLoading;
  final String? availabilityMessage;
  final VoidCallback onPanelToggle;
  final ValueChanged<_ScheduleMode> onModeChanged;
  final VoidCallback onTargetCleared;
  final VoidCallback onAvailabilityRefresh;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        _ScheduleModeSwitch(
          width: compact ? 96 : 104,
          mode: mode,
          onChanged: onModeChanged,
        ),
        SizedBox(width: compact ? 7 : 8),
        SizedBox(
          width: compact ? 184 : 230,
          child: _ScheduleTargetSelector(
            mode: mode,
            target: selectedTarget,
            open: panelOpen,
            onTap: onPanelToggle,
            onClear: onTargetCleared,
          ),
        ),
        if (!compact) ...<Widget>[
          const SizedBox(width: 8),
          SizedBox(
            width: 168,
            child: selectedTarget == null &&
                    availabilityMessage == null &&
                    !availabilityLoading
                ? const SizedBox.shrink()
                : _AvailabilityStatusPill(
                    loading: availabilityLoading,
                    message: availabilityMessage ?? '检测中',
                    onTap: onAvailabilityRefresh,
                  ),
          ),
        ],
      ],
    );
  }
}

class _ScheduleModeSwitch extends StatelessWidget {
  const _ScheduleModeSwitch({
    required this.width,
    required this.mode,
    required this.onChanged,
  });

  final double width;
  final _ScheduleMode mode;
  final ValueChanged<_ScheduleMode> onChanged;

  @override
  Widget build(BuildContext context) {
    return _ShellBox(
      width: width,
      height: 34,
      padding: const EdgeInsets.all(3),
      borderRadius: 11,
      child: Row(
        children: <Widget>[
          Expanded(
            child: _ScheduleModeItem(
              label: '1v1',
              selected: mode == _ScheduleMode.oneToOne,
              onTap: () => onChanged(_ScheduleMode.oneToOne),
            ),
          ),
          Expanded(
            child: _ScheduleModeItem(
              label: '班课',
              selected: mode == _ScheduleMode.groupClass,
              onTap: () => onChanged(_ScheduleMode.groupClass),
            ),
          ),
        ],
      ),
    );
  }
}

class _ScheduleModeItem extends StatelessWidget {
  const _ScheduleModeItem({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 160),
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: selected ? _SmartColors.orange : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: selected ? Colors.white : _SmartColors.text,
            fontSize: 12,
            height: 1,
            fontWeight: FontWeight.w900,
          ),
        ),
      ),
    );
  }
}

class _ScheduleTargetSelector extends StatelessWidget {
  const _ScheduleTargetSelector({
    required this.mode,
    required this.target,
    required this.open,
    required this.onTap,
    required this.onClear,
  });

  final _ScheduleMode mode;
  final ScheduleTargetOption? target;
  final bool open;
  final VoidCallback onTap;
  final VoidCallback onClear;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      key: const ValueKey<String>('schedule-target-selector'),
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: _ShellBox(
        height: 34,
        padding: const EdgeInsets.symmetric(horizontal: 10),
        borderRadius: 11,
        child: Row(
          children: <Widget>[
            Icon(
              mode == _ScheduleMode.oneToOne
                  ? Icons.person_add_alt_1_rounded
                  : Icons.groups_2_outlined,
              color: _SmartColors.text,
              size: 16,
            ),
            const SizedBox(width: 7),
            Expanded(
              child: Text(
                target?.title ??
                    (mode == _ScheduleMode.oneToOne ? '选择1v1' : '选择班课'),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  color:
                      target == null ? _SmartColors.muted : _SmartColors.text,
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            if (target != null) ...<Widget>[
              const SizedBox(width: 4),
              InkWell(
                key: const ValueKey<String>('schedule-target-clear'),
                onTap: onClear,
                borderRadius: BorderRadius.circular(8),
                child: Container(
                  width: 22,
                  height: 22,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFF7EE),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Icon(
                    Icons.close_rounded,
                    color: _SmartColors.muted,
                    size: 15,
                  ),
                ),
              ),
              const SizedBox(width: 4),
            ],
            AnimatedRotation(
              turns: open ? .5 : 0,
              duration: const Duration(milliseconds: 160),
              child: const Icon(
                Icons.keyboard_arrow_down_rounded,
                color: _SmartColors.text,
                size: 18,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _AvailabilityStatusPill extends StatelessWidget {
  const _AvailabilityStatusPill({
    required this.loading,
    required this.message,
    required this.onTap,
  });

  final bool loading;
  final String message;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      child: InkWell(
        onTap: loading ? null : onTap,
        borderRadius: BorderRadius.circular(999),
        child: Container(
          height: 30,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color: loading ? const Color(0xFFFFF8EE) : const Color(0xFFEAF8E9),
            border: Border.all(
              color:
                  loading ? const Color(0xFFF0DDC9) : const Color(0xFFC9EACB),
            ),
            borderRadius: BorderRadius.circular(999),
          ),
          child: Row(
            children: <Widget>[
              if (loading)
                const SizedBox(
                  width: 13,
                  height: 13,
                  child: CircularProgressIndicator(
                    strokeWidth: 1.6,
                    color: _SmartColors.orangeDeep,
                  ),
                )
              else
                const Icon(
                  Icons.verified_outlined,
                  color: _SmartColors.green,
                  size: 15,
                ),
              const SizedBox(width: 6),
              Expanded(
                child: _AvailabilityStatusText(
                  loading: loading,
                  message: message,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _AvailabilityStatusText extends StatelessWidget {
  const _AvailabilityStatusText({
    required this.loading,
    required this.message,
  });

  final bool loading;
  final String message;

  @override
  Widget build(BuildContext context) {
    final TextStyle baseStyle = TextStyle(
      color: loading ? _SmartColors.orangeDeep : _SmartColors.green,
      fontSize: 11,
      fontWeight: FontWeight.w900,
    );
    if (loading) {
      return Text(
        message,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: baseStyle,
      );
    }

    final int conflictIndex = message.indexOf('，冲突');
    if (conflictIndex < 0) {
      return Text(
        message,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: baseStyle,
      );
    }

    return Text.rich(
      TextSpan(
        children: <InlineSpan>[
          TextSpan(text: message.substring(0, conflictIndex + 1)),
          TextSpan(
            text: message.substring(conflictIndex + 1),
            style: baseStyle.copyWith(color: _SmartColors.danger),
          ),
        ],
      ),
      maxLines: 1,
      overflow: TextOverflow.ellipsis,
      style: baseStyle,
    );
  }
}

class _SchedulePickerPanel extends StatelessWidget {
  const _SchedulePickerPanel({
    required this.mode,
    required this.oneToOneTargets,
    required this.groupClassTargets,
    required this.assistantOptions,
    required this.classroomOptions,
    required this.selectedTarget,
    required this.selectedAssistantIds,
    required this.selectedClassroom,
    required this.loading,
    required this.errorMessage,
    required this.onModeChanged,
    required this.onTargetSelected,
    required this.onAssistantToggled,
    required this.onClassroomSelected,
    required this.onRefresh,
    required this.onClose,
  });

  final _ScheduleMode mode;
  final List<ScheduleTargetOption> oneToOneTargets;
  final List<ScheduleTargetOption> groupClassTargets;
  final List<ScheduleStaffOption> assistantOptions;
  final List<ScheduleClassroomOption> classroomOptions;
  final ScheduleTargetOption? selectedTarget;
  final Set<String> selectedAssistantIds;
  final ScheduleClassroomOption? selectedClassroom;
  final bool loading;
  final String? errorMessage;
  final ValueChanged<_ScheduleMode> onModeChanged;
  final ValueChanged<ScheduleTargetOption> onTargetSelected;
  final ValueChanged<ScheduleStaffOption> onAssistantToggled;
  final ValueChanged<ScheduleClassroomOption?> onClassroomSelected;
  final VoidCallback onRefresh;
  final VoidCallback onClose;

  @override
  Widget build(BuildContext context) {
    final List<ScheduleTargetOption> targets =
        mode == _ScheduleMode.oneToOne ? oneToOneTargets : groupClassTargets;
    return Material(
      color: Colors.transparent,
      child: Container(
        height: 334,
        padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
        decoration: BoxDecoration(
          color: _SmartColors.card,
          border: Border.all(color: _SmartColors.line),
          borderRadius: BorderRadius.circular(16),
          boxShadow: const <BoxShadow>[
            BoxShadow(
              color: Color(0x1AB05F32),
              blurRadius: 24,
              offset: Offset(0, 14),
            ),
          ],
        ),
        child: Column(
          children: <Widget>[
            Row(
              children: <Widget>[
                _ScheduleModeSwitch(
                  width: 108,
                  mode: mode,
                  onChanged: onModeChanged,
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    mode == _ScheduleMode.oneToOne
                        ? '选择 1v1 后立即检测本周空闲点'
                        : '选择班课后立即检测本周空闲点',
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.ink,
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                if (loading)
                  const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(
                      strokeWidth: 1.8,
                      color: _SmartColors.orangeDeep,
                    ),
                  )
                else
                  _PanelIconButton(
                    icon: Icons.refresh_rounded,
                    onTap: onRefresh,
                  ),
                const SizedBox(width: 6),
                _PanelIconButton(
                  icon: Icons.close_rounded,
                  onTap: onClose,
                ),
              ],
            ),
            if (errorMessage != null) ...<Widget>[
              const SizedBox(height: 8),
              _PanelErrorBar(message: errorMessage!, onTap: onRefresh),
            ],
            const SizedBox(height: 10),
            Expanded(
              child: Row(
                children: <Widget>[
                  Expanded(
                    flex: 11,
                    child: _ScheduleTargetColumn(
                      title: mode == _ScheduleMode.oneToOne ? '排课对象' : '排课班级',
                      emptyText:
                          mode == _ScheduleMode.oneToOne ? '暂无可排1v1' : '暂无可排班课',
                      targets: targets,
                      selectedTarget: selectedTarget,
                      onSelected: onTargetSelected,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    flex: 8,
                    child: _ScheduleAssistantColumn(
                      assistants: assistantOptions,
                      selectedIds: selectedAssistantIds,
                      onToggled: onAssistantToggled,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    flex: 8,
                    child: _ScheduleClassroomColumn(
                      classrooms: classroomOptions,
                      selectedClassroom: selectedClassroom,
                      onSelected: onClassroomSelected,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _PanelIconButton extends StatelessWidget {
  const _PanelIconButton({required this.icon, required this.onTap});

  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(9),
      child: Container(
        width: 30,
        height: 30,
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: const Color(0xFFFFF7EE),
          border: Border.all(color: _SmartColors.lineSoft),
          borderRadius: BorderRadius.circular(9),
        ),
        child: Icon(icon, color: _SmartColors.text, size: 17),
      ),
    );
  }
}

class _PanelErrorBar extends StatelessWidget {
  const _PanelErrorBar({required this.message, required this.onTap});

  final String message;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Container(
        height: 30,
        padding: const EdgeInsets.symmetric(horizontal: 10),
        decoration: BoxDecoration(
          color: const Color(0xFFFFEFEA),
          border: Border.all(color: const Color(0xFFF4C8BB)),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Row(
          children: <Widget>[
            const Icon(
              Icons.info_outline_rounded,
              color: _SmartColors.orangeDeep,
              size: 15,
            ),
            const SizedBox(width: 6),
            Expanded(
              child: Text(
                message,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _SmartColors.orangeDeep,
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ScheduleTargetColumn extends StatelessWidget {
  const _ScheduleTargetColumn({
    required this.title,
    required this.emptyText,
    required this.targets,
    required this.selectedTarget,
    required this.onSelected,
  });

  final String title;
  final String emptyText;
  final List<ScheduleTargetOption> targets;
  final ScheduleTargetOption? selectedTarget;
  final ValueChanged<ScheduleTargetOption> onSelected;

  @override
  Widget build(BuildContext context) {
    return _PanelSection(
      title: title,
      child: targets.isEmpty
          ? _PanelEmptyText(emptyText)
          : ListView.separated(
              physics: const BouncingScrollPhysics(),
              itemBuilder: (BuildContext context, int index) {
                final ScheduleTargetOption target = targets[index];
                return _ScheduleTargetPanelItem(
                  target: target,
                  selected: selectedTarget?.id == target.id,
                  onTap: () => onSelected(target),
                );
              },
              separatorBuilder: (_, __) => const SizedBox(height: 7),
              itemCount: targets.length,
            ),
    );
  }
}

class _ScheduleTargetPanelItem extends StatelessWidget {
  const _ScheduleTargetPanelItem({
    required this.target,
    required this.selected,
    required this.onTap,
  });

  final ScheduleTargetOption target;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      key: ValueKey<String>('schedule-target-${target.id}'),
      onTap: onTap,
      borderRadius: BorderRadius.circular(11),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 160),
        constraints: const BoxConstraints(minHeight: 48),
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFF1E8) : Colors.white,
          border: Border.all(
            color: selected ? _SmartColors.orange : _SmartColors.lineSoft,
          ),
          borderRadius: BorderRadius.circular(11),
        ),
        child: Row(
          children: <Widget>[
            Container(
              width: 26,
              height: 26,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: selected ? _SmartColors.orange : const Color(0xFFFFF7EE),
                borderRadius: BorderRadius.circular(9),
              ),
              child: Icon(
                selected ? Icons.check_rounded : Icons.menu_book_outlined,
                color: selected ? Colors.white : _SmartColors.text,
                size: 15,
              ),
            ),
            const SizedBox(width: 9),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.center,
                children: <Widget>[
                  Text(
                    target.title,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: target.disabled
                          ? _SmartColors.muted
                          : _SmartColors.ink,
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  if (target.subtitle.trim().isNotEmpty) ...<Widget>[
                    const SizedBox(height: 3),
                    Text(
                      target.subtitle,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _SmartColors.muted,
                        fontSize: 10,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ScheduleAssistantColumn extends StatelessWidget {
  const _ScheduleAssistantColumn({
    required this.assistants,
    required this.selectedIds,
    required this.onToggled,
  });

  final List<ScheduleStaffOption> assistants;
  final Set<String> selectedIds;
  final ValueChanged<ScheduleStaffOption> onToggled;

  @override
  Widget build(BuildContext context) {
    return _PanelSection(
      title: '上课助教',
      child: assistants.isEmpty
          ? const _PanelEmptyText('当前组暂无其他老师')
          : ListView.separated(
              physics: const BouncingScrollPhysics(),
              itemBuilder: (BuildContext context, int index) {
                final ScheduleStaffOption assistant = assistants[index];
                return _ScheduleCheckItem(
                  title: assistant.name,
                  subtitle: assistant.subtitle,
                  selected: selectedIds.contains(assistant.id),
                  onTap: () => onToggled(assistant),
                );
              },
              separatorBuilder: (_, __) => const SizedBox(height: 7),
              itemCount: assistants.length,
            ),
    );
  }
}

class _ScheduleClassroomColumn extends StatelessWidget {
  const _ScheduleClassroomColumn({
    required this.classrooms,
    required this.selectedClassroom,
    required this.onSelected,
  });

  final List<ScheduleClassroomOption> classrooms;
  final ScheduleClassroomOption? selectedClassroom;
  final ValueChanged<ScheduleClassroomOption?> onSelected;

  @override
  Widget build(BuildContext context) {
    return _PanelSection(
      title: '上课教室',
      child: ListView.separated(
        physics: const BouncingScrollPhysics(),
        itemBuilder: (BuildContext context, int index) {
          if (index == 0) {
            return _ScheduleCheckItem(
              title: '不指定教室',
              subtitle: '不校验教室占用冲突',
              selected: selectedClassroom == null,
              onTap: () => onSelected(null),
            );
          }
          final ScheduleClassroomOption classroom = classrooms[index - 1];
          return _ScheduleCheckItem(
            title: classroom.name,
            subtitle: classroom.subtitle.trim().isEmpty
                ? '校验教室占用冲突'
                : '${classroom.subtitle} · 校验教室占用',
            selected: selectedClassroom?.id == classroom.id,
            onTap: () => onSelected(classroom),
          );
        },
        separatorBuilder: (_, __) => const SizedBox(height: 7),
        itemCount: classrooms.length + 1,
      ),
    );
  }
}

class _ScheduleCheckItem extends StatelessWidget {
  const _ScheduleCheckItem({
    required this.title,
    required this.subtitle,
    required this.selected,
    required this.onTap,
  });

  final String title;
  final String subtitle;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(11),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 160),
        constraints: const BoxConstraints(minHeight: 44),
        padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 7),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFF1E8) : Colors.white,
          border: Border.all(
            color: selected ? _SmartColors.orange : _SmartColors.lineSoft,
          ),
          borderRadius: BorderRadius.circular(11),
        ),
        child: Row(
          children: <Widget>[
            Icon(
              selected
                  ? Icons.check_circle_rounded
                  : Icons.radio_button_unchecked_rounded,
              color: selected ? _SmartColors.orange : _SmartColors.muted,
              size: 17,
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    title,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.ink,
                      fontSize: 12,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  if (subtitle.trim().isNotEmpty) ...<Widget>[
                    const SizedBox(height: 2),
                    Text(
                      subtitle,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _SmartColors.muted,
                        fontSize: 10,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _PanelSection extends StatelessWidget {
  const _PanelSection({required this.title, required this.child});

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(9, 9, 9, 8),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFBF7),
        border: Border.all(color: _SmartColors.lineSoft),
        borderRadius: BorderRadius.circular(13),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Padding(
            padding: const EdgeInsets.only(left: 2, bottom: 8),
            child: Text(
              title,
              style: const TextStyle(
                color: _SmartColors.ink,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          Expanded(child: child),
        ],
      ),
    );
  }
}

class _PanelEmptyText extends StatelessWidget {
  const _PanelEmptyText(this.text);

  final String text;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Text(
        text,
        style: const TextStyle(
          color: _SmartColors.muted,
          fontSize: 12,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}
