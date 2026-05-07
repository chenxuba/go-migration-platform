part of '../smart_timetable_page.dart';

class _FilterButton extends StatelessWidget {
  const _FilterButton({
    required this.icon,
    required this.label,
    required this.onTap,
    this.selected = false,
    this.width = 126,
    super.key,
  });

  final IconData icon;
  final String label;
  final VoidCallback onTap;
  final bool selected;
  final double width;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: selected ? const Color(0xFFFFF0E5) : Colors.transparent,
      borderRadius: BorderRadius.circular(11),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(11),
        child: _ShellBox(
          width: width,
          height: 34,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          borderRadius: 11,
          child: Row(
            children: <Widget>[
              Icon(
                icon,
                color: selected ? _SmartColors.orangeDeep : _SmartColors.text,
                size: 16,
              ),
              const SizedBox(width: 7),
              Expanded(
                child: Text(
                  label,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color:
                        selected ? _SmartColors.orangeDeep : _SmartColors.text,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
              const SizedBox(width: 4),
              AnimatedRotation(
                turns: selected ? .5 : 0,
                duration: const Duration(milliseconds: 160),
                child: Icon(
                  Icons.keyboard_arrow_down_rounded,
                  color: selected ? _SmartColors.orangeDeep : _SmartColors.text,
                  size: 16,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _PeriodGroupDropdownButton extends StatelessWidget {
  const _PeriodGroupDropdownButton({
    required this.group,
    required this.open,
    required this.layerLink,
    required this.onTap,
  });

  final _PeriodGroupOption group;
  final bool open;
  final LayerLink layerLink;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return CompositedTransformTarget(
      link: layerLink,
      child: Material(
        color: open ? const Color(0xFFFFF0E5) : Colors.transparent,
        borderRadius: BorderRadius.circular(11),
        child: InkWell(
          key: const ValueKey<String>('period-group-dropdown'),
          onTap: onTap,
          borderRadius: BorderRadius.circular(11),
          child: _ShellBox(
            height: 34,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            borderRadius: 11,
            child: Row(
              children: <Widget>[
                Icon(
                  Icons.view_week_outlined,
                  color: open ? _SmartColors.orangeDeep : _SmartColors.text,
                  size: 16,
                ),
                const SizedBox(width: 7),
                Expanded(
                  child: Text(
                    group.name,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: open ? _SmartColors.orangeDeep : _SmartColors.text,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ),
                const SizedBox(width: 4),
                AnimatedRotation(
                  turns: open ? .5 : 0,
                  duration: const Duration(milliseconds: 160),
                  child: Icon(
                    Icons.keyboard_arrow_down_rounded,
                    color: open ? _SmartColors.orangeDeep : _SmartColors.text,
                    size: 16,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _PeriodGroupDropdownPanel extends StatelessWidget {
  const _PeriodGroupDropdownPanel({
    required this.groups,
    required this.selectedIndex,
    required this.onSelected,
  });

  final List<_PeriodGroupOption> groups;
  final int selectedIndex;
  final ValueChanged<int> onSelected;

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
    return Material(
      color: Colors.transparent,
      child: Container(
        constraints: const BoxConstraints(maxHeight: 320),
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
              padding: EdgeInsets.fromLTRB(4, 0, 4, 8),
              child: Text(
                '分类组',
                style: TextStyle(
                  color: _SmartColors.ink,
                  fontSize: 13,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            Flexible(
              child: ListView.separated(
                shrinkWrap: true,
                itemCount: displayGroups.length,
                separatorBuilder: (_, __) => const SizedBox(height: 6),
                itemBuilder: (BuildContext context, int index) {
                  final _PeriodGroupOption group = displayGroups[index];
                  return _PeriodGroupDropdownItem(
                    key: ValueKey<String>('period-group-option-${group.id}'),
                    group: group,
                    selected: index == safeIndex,
                    onTap: () => onSelected(index),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _PeriodGroupDropdownItem extends StatelessWidget {
  const _PeriodGroupDropdownItem({
    required this.group,
    required this.selected,
    required this.onTap,
    super.key,
  });

  final _PeriodGroupOption group;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        height: 44,
        padding: const EdgeInsets.symmetric(horizontal: 10),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFF1E8) : Colors.transparent,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: <Widget>[
            Container(
              width: 22,
              height: 22,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: selected ? _SmartColors.orange : const Color(0xFFFFF7EE),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(
                selected
                    ? Icons.check_rounded
                    : Icons.radio_button_unchecked_rounded,
                color: selected ? Colors.white : _SmartColors.text,
                size: 14,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    group.name,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: _SmartColors.ink,
                      fontSize: 12,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    group.meta,
                    maxLines: 1,
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

class _TimetableFilterPanel extends StatelessWidget {
  const _TimetableFilterPanel({
    required this.kind,
    required this.options,
    required this.selectedIds,
    required this.onOptionToggled,
    required this.onClear,
  });

  final _TimetableFilterKind kind;
  final List<_TimetableFilterOption> options;
  final Set<String> selectedIds;
  final ValueChanged<String> onOptionToggled;
  final VoidCallback onClear;

  String get _title {
    return switch (kind) {
      _TimetableFilterKind.student => '上课学员',
      _TimetableFilterKind.course => '全部课程',
      _TimetableFilterKind.callStatus => '点名状态',
    };
  }

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: Container(
        constraints: const BoxConstraints(maxHeight: 300),
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
            Row(
              children: <Widget>[
                Text(
                  _title,
                  style: const TextStyle(
                    color: _SmartColors.ink,
                    fontSize: 13,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const Spacer(),
                TextButton(
                  onPressed: onClear,
                  style: TextButton.styleFrom(
                    visualDensity: VisualDensity.compact,
                    padding: const EdgeInsets.symmetric(horizontal: 8),
                  ),
                  child: const Text('清空'),
                ),
              ],
            ),
            const SizedBox(height: 4),
            if (options.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(horizontal: 8, vertical: 12),
                child: Text(
                  '暂无可筛选项',
                  style: TextStyle(
                    color: _SmartColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              )
            else
              Flexible(
                child: ListView.separated(
                  shrinkWrap: true,
                  itemCount: options.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 6),
                  itemBuilder: (BuildContext context, int index) {
                    final _TimetableFilterOption option = options[index];
                    final bool selected = selectedIds.contains(option.id);
                    return _TimetableFilterItem(
                      key: ValueKey<String>(
                        'smart-filter-option-${kind.name}-${option.id}',
                      ),
                      label: option.label,
                      selected: selected,
                      onTap: () => onOptionToggled(option.id),
                    );
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _TimetableFilterItem extends StatelessWidget {
  const _TimetableFilterItem({
    required this.label,
    required this.selected,
    required this.onTap,
    super.key,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(12),
      child: Container(
        height: 40,
        padding: const EdgeInsets.symmetric(horizontal: 10),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFF1E8) : Colors.transparent,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: <Widget>[
            Container(
              width: 22,
              height: 22,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: selected ? _SmartColors.orange : const Color(0xFFFFF7EE),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Icon(
                selected ? Icons.check_rounded : Icons.check_box_outline_blank,
                color: selected ? Colors.white : _SmartColors.text,
                size: 14,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                label,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _SmartColors.ink,
                  fontSize: 12,
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

class _TimetableLoadStatus extends StatelessWidget {
  const _TimetableLoadStatus({
    required this.message,
    required this.onRefresh,
  });

  final String message;
  final VoidCallback onRefresh;

  @override
  Widget build(BuildContext context) {
    return ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 230),
      child: InkWell(
        onTap: onRefresh,
        borderRadius: BorderRadius.circular(999),
        child: Container(
          height: 30,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color: const Color(0xFFFFEFEA),
            border: Border.all(color: const Color(0xFFF4C8BB)),
            borderRadius: BorderRadius.circular(999),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(
                Icons.refresh_rounded,
                color: _SmartColors.orangeDeep,
                size: 15,
              ),
              const SizedBox(width: 5),
              Flexible(
                child: Text(
                  message,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _SmartColors.orangeDeep,
                    fontSize: 11,
                    height: 1,
                    fontWeight: FontWeight.w800,
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

class _SummaryAccent extends StatelessWidget {
  const _SummaryAccent();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 4,
      height: 16,
      decoration: BoxDecoration(
        color: _SmartColors.orange,
        borderRadius: BorderRadius.circular(999),
      ),
    );
  }
}

class _LegendItem extends StatelessWidget {
  const _LegendItem({required this.color, required this.label});

  final Color color;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Container(
          width: 16,
          height: 4,
          decoration: BoxDecoration(
            color: color,
            borderRadius: BorderRadius.circular(999),
          ),
        ),
        const SizedBox(width: 6),
        Text(
          label,
          style: const TextStyle(
            color: _SmartColors.text,
            fontSize: 11,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    );
  }
}

class _IconShell extends StatelessWidget {
  const _IconShell({required this.size, required this.icon, this.onTap});

  final double size;
  final IconData icon;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return _ShellBox(
      width: size,
      height: size,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Icon(icon, color: _SmartColors.ink, size: 24),
      ),
    );
  }
}

class _ShellBox extends StatelessWidget {
  const _ShellBox({
    required this.child,
    this.width,
    this.height,
    this.padding,
    this.borderRadius = 13,
  });

  final Widget child;
  final double? width;
  final double? height;
  final EdgeInsetsGeometry? padding;
  final double borderRadius;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: _SmartColors.card.withOpacity(.92),
      borderRadius: BorderRadius.circular(borderRadius),
      child: Container(
        width: width,
        height: height,
        padding: padding,
        decoration: BoxDecoration(
          border: Border.all(color: _SmartColors.line),
          borderRadius: BorderRadius.circular(borderRadius),
        ),
        child: child,
      ),
    );
  }
}

class _TimetableSkeletonBox extends StatelessWidget {
  const _TimetableSkeletonBox({
    this.width,
    this.height = 14,
    this.radius = 13,
  });

  final double? width;
  final double height;
  final double radius;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: const Color(0xFFF3E6DA),
        borderRadius: BorderRadius.circular(radius),
        border: Border.all(
          color: const Color(0xFFF0DFD1),
        ),
      ),
    );
  }
}

class _DiagonalPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint line = Paint()
      ..color = _SmartColors.line
      ..strokeWidth = 1;
    canvas.drawLine(Offset.zero, Offset(size.width, size.height), line);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
