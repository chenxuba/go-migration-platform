part of 'autismdev_assessment_page.dart';

class _AutismDevRightRail extends StatelessWidget {
  const _AutismDevRightRail({
    required this.group,
    required this.item,
    required this.remarkController,
    required this.selectedRangeFilter,
    required this.itemScores,
    required this.answeredCount,
    required this.totalCount,
    required this.missingCount,
    required this.onSelectRangeFilter,
  });

  final AutismDevDomainGroup group;
  final AutismDevItemSummary item;
  final TextEditingController remarkController;
  final String selectedRangeFilter;
  final Map<int, String> itemScores;
  final int answeredCount;
  final int totalCount;
  final int missingCount;
  final ValueChanged<String> onSelectRangeFilter;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: _panelDecoration(),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 6),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            _ProgressSummary(
              answeredCount: answeredCount,
              totalCount: totalCount,
              missing: missingCount,
            ),
            const SizedBox(height: 16),
            Expanded(
              child: _AutismDevRangeQuickFilter(
                group: group,
                selectedRangeFilter: selectedRangeFilter,
                itemScores: itemScores,
                onSelectRangeFilter: onSelectRangeFilter,
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              '备注',
              style: TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 16,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 10),
            SizedBox(
              height: 66,
              child: TextField(
                controller: remarkController,
                minLines: 2,
                maxLines: 2,
                onTapOutside: (_) {
                  FocusManager.instance.primaryFocus?.unfocus();
                },
                onEditingComplete: () {
                  FocusManager.instance.primaryFocus?.unfocus();
                },
                textAlignVertical: TextAlignVertical.top,
                style: const TextStyle(
                  color: _AutismDevColors.body,
                  fontSize: 15,
                  height: 1.35,
                  fontWeight: FontWeight.w600,
                ),
                decoration: InputDecoration(
                  hintText: '请输入备注',
                  hintStyle: const TextStyle(
                    color: _AutismDevColors.muted,
                    fontSize: 12,
                    height: 1.25,
                    fontWeight: FontWeight.w600,
                  ),
                  filled: true,
                  fillColor: _AutismDevColors.softPanel,
                  contentPadding: const EdgeInsets.fromLTRB(12, 9, 12, 9),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: const BorderSide(color: _AutismDevColors.line),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide:
                        const BorderSide(color: _AutismDevColors.orange),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _AutismDevScoreDock extends StatelessWidget {
  const _AutismDevScoreDock({
    required this.scoreOptions,
    required this.selectedScore,
    required this.onScore,
  });

  final List<AutismDevScoreOption> scoreOptions;
  final String? selectedScore;
  final ValueChanged<String> onScore;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        const Row(
          children: <Widget>[
            Text(
              '评分',
              style: TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 15,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            Spacer(),
          ],
        ),
        const SizedBox(height: 8),
        LayoutBuilder(
          builder: (BuildContext context, BoxConstraints constraints) {
            const double spacing = 12;
            final bool singleColumn = constraints.maxWidth < 460;
            final double cardWidth = scoreOptions.length == 1 || singleColumn
                ? constraints.maxWidth
                : (constraints.maxWidth - spacing) / 2;
            return Wrap(
              spacing: spacing,
              runSpacing: 8,
              children: <Widget>[
                for (final AutismDevScoreOption option in scoreOptions)
                  SizedBox(
                    width: cardWidth,
                    child: _ScoreOptionButton(
                      option: option,
                      selected: selectedScore == option.value,
                      onTap: () => onScore(option.value),
                    ),
                  ),
              ],
            );
          },
        ),
      ],
    );
  }
}

class _ScoreOptionButton extends StatelessWidget {
  const _ScoreOptionButton({
    required this.option,
    required this.selected,
    required this.onTap,
  });

  final AutismDevScoreOption option;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color color = _scoreColor(option.value);
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(10),
      child: InkWell(
        borderRadius: BorderRadius.circular(10),
        onTap: onTap,
        child: Ink(
          height: 78,
          padding: const EdgeInsets.fromLTRB(13, 9, 11, 9),
          decoration: BoxDecoration(
            color: selected ? color.withOpacity(.08) : Colors.white,
            border: Border.all(
              color: selected ? color : _AutismDevColors.line,
              width: selected ? 1.5 : 1,
            ),
            borderRadius: BorderRadius.circular(10),
          ),
          child: Row(
            children: <Widget>[
              Container(
                width: 36,
                height: 36,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: color.withOpacity(selected ? .14 : .08),
                  borderRadius: BorderRadius.circular(18),
                  border: Border.all(color: color.withOpacity(.38)),
                ),
                child: Text(
                  option.value,
                  style: TextStyle(
                    color: color,
                    fontSize: 21,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text(
                      _optionTitle(option),
                      maxLines: 2,
                      style: TextStyle(
                        color: _AutismDevColors.ink,
                        fontSize: 14,
                        height: 1.12,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      option.description,
                      maxLines: 2,
                      style: TextStyle(
                        color: _AutismDevColors.muted,
                        fontSize: 10.5,
                        height: 1.12,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              Icon(
                selected
                    ? Icons.check_circle_rounded
                    : Icons.radio_button_unchecked_rounded,
                color: selected ? color : const Color(0xFFCAB8AA),
                size: 20,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ProgressSummary extends StatelessWidget {
  const _ProgressSummary({
    required this.answeredCount,
    required this.totalCount,
    required this.missing,
  });

  final int answeredCount;
  final int totalCount;
  final int missing;

  @override
  Widget build(BuildContext context) {
    final double percent = totalCount <= 0 ? 0 : answeredCount / totalCount;
    final double normalizedPercent = percent.clamp(0, 1).toDouble();
    final int percentText = (normalizedPercent * 100).round();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        const Text(
          '当前进度',
          style: TextStyle(
            color: _AutismDevColors.ink,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 14),
        Row(
          children: <Widget>[
            SizedBox(
              width: 82,
              height: 82,
              child: CustomPaint(
                painter: _AutismDevDonutPainter(percent: percentText),
                child: Center(
                  child: Text(
                    '$percentText%',
                    style: const TextStyle(
                      color: _AutismDevColors.ink,
                      fontSize: 21,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  _ProgressText(
                    label: '已完成',
                    value: '$answeredCount / $totalCount 项',
                  ),
                  const SizedBox(height: 10),
                  _ProgressText(
                    label: '缺题',
                    value: '$missing 项',
                    danger: true,
                  ),
                ],
              ),
            ),
          ],
        ),
      ],
    );
  }
}

class _ProgressText extends StatelessWidget {
  const _ProgressText({
    required this.label,
    required this.value,
    this.danger = false,
  });

  final String label;
  final String value;
  final bool danger;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          label,
          style: const TextStyle(
            color: _AutismDevColors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w800,
          ),
        ),
        const SizedBox(height: 5),
        Text(
          value,
          style: TextStyle(
            color:
                danger ? const Color(0xFFE04438) : _AutismDevColors.orangeDeep,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}

class _AutismDevDonutPainter extends CustomPainter {
  const _AutismDevDonutPainter({required this.percent});

  final int percent;

  @override
  void paint(Canvas canvas, Size size) {
    final Offset center = Offset(size.width / 2, size.height / 2);
    final double radius = math.min(size.width, size.height) / 2 - 6;
    final Paint track = Paint()
      ..color = const Color(0xFFE9DDD3)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 10
      ..strokeCap = StrokeCap.round;
    final Paint progress = Paint()
      ..color = _AutismDevColors.orange
      ..style = PaintingStyle.stroke
      ..strokeWidth = 10
      ..strokeCap = StrokeCap.round;
    canvas.drawCircle(center, radius, track);
    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      -math.pi / 2,
      math.pi * 2 * percent.clamp(0, 100) / 100,
      false,
      progress,
    );
  }

  @override
  bool shouldRepaint(covariant _AutismDevDonutPainter oldDelegate) {
    return oldDelegate.percent != percent;
  }
}
