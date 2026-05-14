part of 'autismdev_assessment_page.dart';

class _AutismDevItemDetailCard extends StatelessWidget {
  const _AutismDevItemDetailCard({
    required this.item,
    required this.detail,
  });

  final AutismDevItemSummary item;
  final AutismDevAssessmentItem? detail;

  @override
  Widget build(BuildContext context) {
    final AutismDevAssessmentItem? currentDetail = detail;
    final String rangeText = _assessmentRangeText(item, currentDetail);
    final String ageText =
        _detailText(currentDetail?.ageSegment, item.ageSegment);
    final String methodText =
        _nonEmptyDetailText(currentDetail?.method, item.method);
    final String criteriaText =
        _nonEmptyDetailText(currentDetail?.passCriteria, item.passCriteria);
    final List<_AutismDevDetailField> cards = <_AutismDevDetailField>[
      _AutismDevDetailField(
        icon: Icons.inventory_2_outlined,
        label: '评估材料',
        value: _detailText(currentDetail?.materials, item.materials),
      ),
      _AutismDevDetailField(
        icon: Icons.assignment_outlined,
        label: '评估方法',
        value: methodText.isEmpty ? ' ' : methodText,
      ),
      _AutismDevDetailField(
        icon: Icons.article_outlined,
        label: '评分标准',
        value: criteriaText.isEmpty ? ' ' : _criteriaDisplayText(criteriaText),
      ),
    ];
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        LayoutBuilder(
          builder: (BuildContext context, BoxConstraints constraints) {
            final bool compact = constraints.maxWidth < 540;
            if (compact) {
              return Column(
                children: <Widget>[
                  _AutismDevMetaCard(
                    icon: Icons.account_tree_outlined,
                    label: '评估范围',
                    value: rangeText,
                  ),
                  const SizedBox(height: 10),
                  _AutismDevMetaCard(
                    icon: Icons.calendar_month_outlined,
                    label: '参考年龄',
                    value: ageText,
                  ),
                ],
              );
            }
            return Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Expanded(
                  child: _AutismDevMetaCard(
                    icon: Icons.account_tree_outlined,
                    label: '评估范围',
                    value: rangeText,
                  ),
                ),
                const SizedBox(width: 10),
                SizedBox(
                  width: 158,
                  child: _AutismDevMetaCard(
                    icon: Icons.calendar_month_outlined,
                    label: '参考年龄',
                    value: ageText,
                  ),
                ),
              ],
            );
          },
        ),
        const SizedBox(height: 10),
        for (final _AutismDevDetailField card in cards)
          _AutismDevDetailInfoCard(field: card),
      ],
    );
  }
}

class _AutismDevDetailField {
  const _AutismDevDetailField({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final String value;
}

class _AutismDevMetaCard extends StatelessWidget {
  const _AutismDevMetaCard({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(minHeight: 70),
      padding: const EdgeInsets.fromLTRB(13, 12, 13, 12),
      decoration: _autismDevDetailCardDecoration(),
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
                child: Icon(
                  icon,
                  size: 12,
                  color: _AutismDevColors.orange,
                ),
              ),
              const SizedBox(width: 7),
              Text(
                label,
                style: const TextStyle(
                  color: _AutismDevColors.ink,
                  fontSize: 13,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            value,
            style: const TextStyle(
              color: _AutismDevColors.body,
              fontSize: 14,
              height: 1.28,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevDetailInfoCard extends StatelessWidget {
  const _AutismDevDetailInfoCard({required this.field});

  final _AutismDevDetailField field;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 14),
      decoration: _autismDevDetailCardDecoration(),
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
                child: Icon(
                  field.icon,
                  size: 12,
                  color: _AutismDevColors.orange,
                ),
              ),
              const SizedBox(width: 7),
              Text(
                field.label,
                style: const TextStyle(
                  color: _AutismDevColors.ink,
                  fontSize: 16,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 11),
          Text(
            field.value,
            style: const TextStyle(
              color: _AutismDevColors.body,
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

BoxDecoration _autismDevDetailCardDecoration() {
  return BoxDecoration(
    color: Colors.white.withOpacity(.94),
    borderRadius: BorderRadius.circular(10),
    border: Border.all(color: _AutismDevColors.line),
    boxShadow: _autismDevShadow(
      color: const Color(0x0FB05F32),
      blur: 12,
      offset: const Offset(0, 6),
    ),
  );
}
