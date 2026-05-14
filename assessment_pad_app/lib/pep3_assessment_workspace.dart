part of 'pep3_assessment_page.dart';

class _Pep3QuestionPanel extends StatelessWidget {
  const _Pep3QuestionPanel({
    required this.controller,
    required this.loading,
    required this.item,
    required this.summary,
    required this.scoreOptions,
    required this.selectedScore,
    required this.previousScore,
    required this.previousAssessmentDate,
    required this.saving,
    required this.saved,
    required this.recordValues,
    required this.onScore,
    required this.onRecordValue,
  });

  final ScrollController controller;
  final bool loading;
  final Pep3AssessmentItem? item;
  final Pep3ItemSummary? summary;
  final List<Pep3ScoreOption> scoreOptions;
  final int? selectedScore;
  final int? previousScore;
  final String previousAssessmentDate;
  final bool saving;
  final bool saved;
  final Map<String, dynamic> recordValues;
  final ValueChanged<int> onScore;
  final void Function(String key, dynamic value) onRecordValue;

  @override
  Widget build(BuildContext context) {
    final Pep3AssessmentItem resolved = item ?? Pep3AssessmentItem.empty;
    final Pep3ItemSummary? resolvedSummary = summary;
    return _RailCard(
      child: loading && item == null
          ? _QuestionLoadingView(
              summary: resolvedSummary,
              scoreOptions: scoreOptions,
            )
          : Padding(
              padding: const EdgeInsets.fromLTRB(16, 18, 16, 14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  Row(
                    children: <Widget>[
                      Expanded(
                        child: Text(
                          '第 ${resolvedSummary?.itemNo ?? resolved.itemNo} 题  ${resolvedSummary?.displayTitle ?? resolved.displayTitle}',
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            color: _Pep3Colors.ink,
                            fontSize: 23,
                            height: 1,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),
                      _DomainChip(
                        code: resolved.domainCode.isNotEmpty
                            ? resolved.domainCode
                            : resolvedSummary?.domainCode ?? '',
                        name: resolved.domainName.isNotEmpty
                            ? resolved.domainName
                            : resolvedSummary?.domainName ?? '',
                      ),
                      if (saving || saved) ...<Widget>[
                        const SizedBox(width: 8),
                        _SaveBadge(saving: saving),
                      ],
                    ],
                  ),
                  const SizedBox(height: 16),
                  Expanded(
                    child: ListView(
                      key: const ValueKey<String>(
                        'pep3-question-instruction-scroll',
                      ),
                      controller: controller,
                      padding: EdgeInsets.zero,
                      physics: const BouncingScrollPhysics(),
                      children: <Widget>[
                        _InstructionCard(
                          title: '材料',
                          icon: Icons.article_outlined,
                          body: _normalizeText(resolved.materials),
                        ),
                        _InstructionCard(
                          title: '操作标准',
                          icon: Icons.assignment_outlined,
                          body: _normalizeText(resolved.method),
                        ),
                        _InstructionCard(
                          title: '指导语',
                          icon: Icons.record_voice_over_outlined,
                          body: _normalizeText(resolved.guidance),
                        ),
                        _InstructionCard(
                          title: '评分标准',
                          icon: Icons.fact_check_outlined,
                          body: _scoreStandardText(resolved, scoreOptions),
                        ),
                        const SizedBox(height: 2),
                      ],
                    ),
                  ),
                  const SizedBox(height: 10),
                  _ScoreDock(
                    scoreOptions: scoreOptions,
                    selectedScore: selectedScore,
                    previousScore: previousScore,
                    previousAssessmentDate: previousAssessmentDate,
                    onScore: onScore,
                  ),
                ],
              ),
            ),
    );
  }
}

class _QuestionLoadingView extends StatelessWidget {
  const _QuestionLoadingView({
    required this.summary,
    required this.scoreOptions,
  });

  final Pep3ItemSummary? summary;
  final List<Pep3ScoreOption> scoreOptions;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 18, 16, 14),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Row(
            children: <Widget>[
              Expanded(
                child: Text(
                  '第 ${summary?.itemNo ?? 0} 题  ${summary?.displayTitle ?? ''}',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _Pep3Colors.ink,
                    fontSize: 23,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 10),
              _DomainChip(
                code: summary?.domainCode ?? '',
                name: summary?.domainName ?? '',
              ),
            ],
          ),
          const SizedBox(height: 16),
          Expanded(
            child: ListView(
              padding: EdgeInsets.zero,
              physics: const NeverScrollableScrollPhysics(),
              children: const <Widget>[
                _QuestionLoadingCard(title: '材料', height: 72),
                _QuestionLoadingCard(title: '操作标准', height: 86),
                _QuestionLoadingCard(title: '指导语', height: 86),
                _QuestionLoadingCard(title: '评分标准', height: 86),
                SizedBox(height: 2),
              ],
            ),
          ),
          const SizedBox(height: 10),
          _ScoreLoadingDock(scoreOptions: scoreOptions),
        ],
      ),
    );
  }
}

class _ScoreDock extends StatelessWidget {
  const _ScoreDock({
    required this.scoreOptions,
    required this.selectedScore,
    required this.previousScore,
    required this.previousAssessmentDate,
    required this.onScore,
  });

  final List<Pep3ScoreOption> scoreOptions;
  final int? selectedScore;
  final int? previousScore;
  final String previousAssessmentDate;
  final ValueChanged<int> onScore;

  @override
  Widget build(BuildContext context) {
    final bool hasPrevious =
        previousScore != null && previousAssessmentDate.trim().isNotEmpty;
    return Column(
      key: const ValueKey<String>('pep3-question-score-dock'),
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Row(
          children: <Widget>[
            const Text(
              '评分',
              style: TextStyle(
                color: _Pep3Colors.ink,
                fontSize: 15,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            const Spacer(),
            if (hasPrevious)
              _PreviousScoreSummary(
                score: previousScore!,
                date: previousAssessmentDate,
              ),
          ],
        ),
        const SizedBox(height: 10),
        Row(
          children: <Widget>[
            for (int i = 0; i < scoreOptions.length; i++) ...<Widget>[
              Expanded(
                child: _ScoreOptionCard(
                  option: scoreOptions[i],
                  selected: selectedScore == scoreOptions[i].value,
                  previous:
                      hasPrevious && previousScore == scoreOptions[i].value,
                  previousDate: previousAssessmentDate,
                  onTap: () => onScore(scoreOptions[i].value),
                ),
              ),
              if (i != scoreOptions.length - 1) const SizedBox(width: 14),
            ],
          ],
        ),
      ],
    );
  }
}

class _ScoreLoadingDock extends StatelessWidget {
  const _ScoreLoadingDock({required this.scoreOptions});

  final List<Pep3ScoreOption> scoreOptions;

  @override
  Widget build(BuildContext context) {
    final int optionCount = math.max(scoreOptions.length, 3);
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Row(
          children: <Widget>[
            const Text(
              '评分',
              style: TextStyle(
                color: _Pep3Colors.ink,
                fontSize: 15,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            const Spacer(),
            Container(
              height: 28,
              padding: const EdgeInsets.symmetric(horizontal: 12),
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: const Color(0xFFFFF2EA),
                borderRadius: BorderRadius.circular(9),
                border: Border.all(color: const Color(0xFFFFD6C3)),
              ),
              child: const Row(
                mainAxisSize: MainAxisSize.min,
                children: <Widget>[
                  SizedBox(
                    width: 13,
                    height: 13,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: _Pep3Colors.orange,
                    ),
                  ),
                  SizedBox(width: 8),
                  Text(
                    '题目加载中',
                    style: TextStyle(
                      color: _Pep3Colors.orangeDeep,
                      fontSize: 12,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        Row(
          children: <Widget>[
            for (int i = 0; i < optionCount; i++) ...<Widget>[
              const Expanded(child: _ScoreLoadingCard()),
              if (i != optionCount - 1) const SizedBox(width: 14),
            ],
          ],
        ),
      ],
    );
  }
}

class _QuestionLoadingCard extends StatelessWidget {
  const _QuestionLoadingCard({required this.title, required this.height});

  final String title;
  final double height;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      constraints: BoxConstraints(minHeight: height),
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.94),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _Pep3Colors.line),
        boxShadow: _pep3Shadow(
          color: const Color(0x0FB05F32),
          blur: 12,
          offset: const Offset(0, 6),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Row(
            children: <Widget>[
              Container(
                width: 17,
                height: 17,
                decoration: BoxDecoration(
                  color: const Color(0xFFFFEFE6),
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: const Color(0xFFFFCFB6)),
                ),
              ),
              const SizedBox(width: 7),
              Text(
                title,
                style: const TextStyle(
                  color: _Pep3Colors.ink,
                  fontSize: 16,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          const _LoadingLine(widthFactor: .72),
          const SizedBox(height: 8),
          const _LoadingLine(widthFactor: .92),
        ],
      ),
    );
  }
}

class _ScoreLoadingCard extends StatelessWidget {
  const _ScoreLoadingCard();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 94,
      padding: const EdgeInsets.fromLTRB(16, 15, 14, 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _Pep3Colors.line),
      ),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: <Widget>[
          _LoadingLine(width: 54, height: 20),
          SizedBox(height: 12),
          _LoadingLine(width: 72, height: 12),
        ],
      ),
    );
  }
}

class _LoadingLine extends StatelessWidget {
  const _LoadingLine({
    this.width,
    this.widthFactor,
    this.height = 10,
  });

  final double? width;
  final double? widthFactor;
  final double height;

  @override
  Widget build(BuildContext context) {
    final Widget line = Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: const Color(0xFFF3E8DF),
        borderRadius: BorderRadius.circular(99),
      ),
    );
    if (widthFactor == null) {
      return line;
    }
    return FractionallySizedBox(
      alignment: Alignment.centerLeft,
      widthFactor: widthFactor,
      child: line,
    );
  }
}

class _InstructionCard extends StatelessWidget {
  const _InstructionCard({
    required this.title,
    required this.icon,
    required this.body,
  });

  final String title;
  final IconData icon;
  final String body;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.94),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _Pep3Colors.line),
        boxShadow: _pep3Shadow(
          color: const Color(0x0FB05F32),
          blur: 12,
          offset: const Offset(0, 6),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Container(
                width: 17,
                height: 17,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: const Color(0xFFFFEFE6),
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: const Color(0xFFFFCFB6)),
                ),
                child: Icon(icon, size: 12, color: _Pep3Colors.orange),
              ),
              const SizedBox(width: 7),
              Text(
                title,
                style: const TextStyle(
                  color: _Pep3Colors.ink,
                  fontSize: 16,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 11),
          Text(
            body,
            style: const TextStyle(
              color: _Pep3Colors.text,
              fontSize: 14,
              height: 1.55,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}

class _ScoreOptionCard extends StatelessWidget {
  const _ScoreOptionCard({
    required this.option,
    required this.selected,
    required this.previous,
    required this.previousDate,
    required this.onTap,
  });

  final Pep3ScoreOption option;
  final bool selected;
  final bool previous;
  final String previousDate;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color color = _scoreColor(option.value);
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 94,
          padding: const EdgeInsets.fromLTRB(16, 9, 14, 9),
          decoration: BoxDecoration(
            color: selected ? color.withOpacity(.08) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: selected ? color : _Pep3Colors.line,
              width: selected ? 1.5 : 1,
            ),
          ),
          child: Stack(
            fit: StackFit.expand,
            clipBehavior: Clip.none,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: <Widget>[
                        Text(
                          '${option.value} 分',
                          style: TextStyle(
                            color: color,
                            fontSize: 24,
                            height: 1,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          _shortScoreLabel(option.value, option.label),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            color: _Pep3Colors.text,
                            fontSize: 13,
                            height: 1,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 36),
                  Icon(
                    selected
                        ? Icons.check_circle_rounded
                        : Icons.radio_button_unchecked_rounded,
                    color: selected ? color : const Color(0xFFCAB8AA),
                    size: 23,
                  ),
                ],
              ),
              if (previous)
                Positioned(
                  top: -5,
                  right: -8,
                  child: _PreviousScoreBadge(date: previousDate, color: color),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _PreviousScoreSummary extends StatelessWidget {
  const _PreviousScoreSummary({required this.score, required this.date});

  final int score;
  final String date;

  @override
  Widget build(BuildContext context) {
    final Color color = _scoreColor(score);
    return Container(
      height: 32,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: color.withOpacity(.06),
        borderRadius: BorderRadius.circular(9),
        border: Border.all(color: color.withOpacity(.38)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Container(
            width: 7,
            height: 7,
            decoration: BoxDecoration(color: color, shape: BoxShape.circle),
          ),
          const SizedBox(width: 9),
          Text(
            '上次测评 ${_compactDateLabel(date)}',
            style: const TextStyle(
              color: _Pep3Colors.text,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(width: 10),
          Text(
            '$score 分 · ${_shortScoreLabel(score, '')}',
            style: TextStyle(
              color: color,
              fontSize: 14,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _PreviousScoreBadge extends StatelessWidget {
  const _PreviousScoreBadge({required this.date, required this.color});

  final String date;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(9),
        border: Border.all(color: color.withOpacity(.62), width: 1.1),
      ),
      child: Text(
        '上次 ${_shortDateLabel(date)}',
        maxLines: 1,
        style: TextStyle(
          color: color,
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}
