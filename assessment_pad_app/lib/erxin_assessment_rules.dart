part of 'erxin_assessment_page.dart';

class _RuleCard extends StatelessWidget {
  const _RuleCard({
    required this.title,
    required this.body,
    required this.icon,
    required this.color,
  });

  final String title;
  final String body;
  final IconData icon;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: color, size: 22),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  title,
                  style: const TextStyle(
                    color: _ErxinColors.ink,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 3),
                Text(
                  body,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _ErxinColors.body,
                    fontSize: 12,
                    height: 1.3,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _RuleChecklist extends StatefulWidget {
  const _RuleChecklist({
    required this.rows,
    required this.revealMonths,
    required this.revealSerial,
    required this.onTapMonth,
    required this.onRevealTargets,
  });

  final List<_RuleRow> rows;
  final List<int> revealMonths;
  final int revealSerial;
  final ValueChanged<int> onTapMonth;
  final ValueChanged<List<int>> onRevealTargets;

  @override
  State<_RuleChecklist> createState() => _RuleChecklistState();
}

class _RuleChecklistState extends State<_RuleChecklist> {
  static const double _rowExtent = 41;

  final ScrollController _scrollController = ScrollController();
  final Set<int> _flashingMonths = <int>{};
  Timer? _flashTimer;

  @override
  void dispose() {
    _flashTimer?.cancel();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(covariant _RuleChecklist oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.revealSerial == widget.revealSerial ||
        widget.revealMonths.isEmpty) {
      return;
    }
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) {
        return;
      }
      _revealMonths(widget.revealMonths);
    });
  }

  void _revealTargetMonths(_RuleRow row) {
    _revealMonths(row.targetMonths);
    widget.onRevealTargets(row.targetMonths);
  }

  void _revealMonths(List<int> months) {
    if (months.isEmpty) {
      return;
    }
    final Set<int> targetMonths = months.toSet();
    final List<_RuleRow> monthRows = widget.rows
        .where((_RuleRow candidate) => candidate.month != null)
        .toList(growable: false);
    final int targetIndex = monthRows.indexWhere(
      (_RuleRow candidate) => targetMonths.contains(candidate.month),
    );
    if (targetIndex < 0) {
      return;
    }

    _flashTimer?.cancel();
    setState(() {
      _flashingMonths
        ..clear()
        ..addAll(targetMonths);
    });
    _flashTimer = Timer(const Duration(milliseconds: 900), () {
      if (!mounted) {
        return;
      }
      setState(_flashingMonths.clear);
    });

    if (!_scrollController.hasClients) {
      return;
    }
    final double maxOffset = _scrollController.position.maxScrollExtent;
    final double targetOffset = math.min(
      maxOffset,
      math.max(0, 6 + targetIndex * _rowExtent),
    );
    unawaited(
      _scrollController.animateTo(
        targetOffset,
        duration: const Duration(milliseconds: 260),
        curve: Curves.easeOutCubic,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final List<_RuleRow> monthRows = widget.rows
        .where((_RuleRow row) => row.month != null)
        .toList(growable: false);
    final List<_RuleRow> pinnedRows = widget.rows
        .where((_RuleRow row) => row.month == null)
        .toList(growable: false);
    return Container(
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Column(
        children: <Widget>[
          Expanded(
            child: ListView.separated(
              controller: _scrollController,
              padding: const EdgeInsets.symmetric(vertical: 6),
              physics: const BouncingScrollPhysics(),
              itemCount: monthRows.length,
              separatorBuilder: (_, __) => const Divider(
                height: 1,
                thickness: 1,
                color: _ErxinColors.line,
              ),
              itemBuilder: (BuildContext context, int index) {
                final _RuleRow row = monthRows[index];
                return _RuleChecklistRow(
                  row: row,
                  highlighted:
                      row.month != null && _flashingMonths.contains(row.month),
                  onTapMonth: widget.onTapMonth,
                );
              },
            ),
          ),
          if (pinnedRows.isNotEmpty) ...<Widget>[
            const Divider(height: 1, thickness: 1, color: _ErxinColors.line),
            for (final MapEntry<int, _RuleRow> entry
                in pinnedRows.asMap().entries) ...<Widget>[
              _RuleChecklistRow(
                row: entry.value,
                onTapMonth: widget.onTapMonth,
                onTapTargets: () => _revealTargetMonths(entry.value),
              ),
              if (entry.key < pinnedRows.length - 1)
                const Divider(
                  height: 1,
                  thickness: 1,
                  color: _ErxinColors.line,
                ),
            ],
          ],
        ],
      ),
    );
  }
}

class _RuleChecklistRow extends StatelessWidget {
  const _RuleChecklistRow({
    required this.row,
    required this.onTapMonth,
    this.highlighted = false,
    this.onTapTargets,
  });

  final _RuleRow row;
  final ValueChanged<int> onTapMonth;
  final bool highlighted;
  final VoidCallback? onTapTargets;

  @override
  Widget build(BuildContext context) {
    final bool targetClickable =
        row.month == null && row.done && row.targetMonths.isNotEmpty;
    final bool clickable = row.month != null || targetClickable;
    final bool unmetResult = _ruleRowHasUnmetResult(row);
    return Material(
      color: highlighted
          ? const Color(0xFFFFF3BF)
          : row.selected
              ? const Color(0xFFFFF1E8)
              : Colors.white,
      child: InkWell(
        onTap: row.month != null
            ? () => onTapMonth(row.month!)
            : targetClickable
                ? onTapTargets
                : null,
        child: SizedBox(
          height: 40,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12),
            child: Row(
              children: <Widget>[
                Icon(
                  row.done
                      ? Icons.check_circle
                      : unmetResult
                          ? Icons.cancel_rounded
                          : Icons.radio_button_unchecked,
                  color: row.done
                      ? _ErxinColors.green
                      : unmetResult
                          ? _ErxinColors.red
                          : _ErxinColors.muted,
                  size: 18,
                ),
                const SizedBox(width: 9),
                Expanded(
                  child: Text(
                    row.label,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color:
                          row.selected ? _ErxinColors.blue : _ErxinColors.ink,
                      fontSize: 13,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  row.value,
                  maxLines: 1,
                  softWrap: false,
                  style: TextStyle(
                    color: row.done
                        ? _ErxinColors.green
                        : unmetResult
                            ? _ErxinColors.red
                            : _ErxinColors.body,
                    fontSize: 13,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                if (clickable) ...<Widget>[
                  const SizedBox(width: 5),
                  Icon(
                    targetClickable
                        ? Icons.center_focus_strong_rounded
                        : row.selected
                            ? Icons.edit_note_rounded
                            : Icons.history_rounded,
                    color: targetClickable || row.selected
                        ? _ErxinColors.blue
                        : _ErxinColors.muted,
                    size: 16,
                  ),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}

bool _ruleRowHasUnmetResult(_RuleRow row) {
  if (row.done || row.month == null) {
    return false;
  }
  return row.value.contains('未全') || row.value.contains('未通过');
}
