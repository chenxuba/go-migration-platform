part of 'autismdev_assessment_page.dart';

class _AutismDevWorkspacePanel extends StatelessWidget {
  const _AutismDevWorkspacePanel({
    required this.item,
    required this.detail,
    required this.selectedScore,
    required this.scoreOptions,
    required this.onOpenQuestionPreference,
    required this.onScore,
  });

  final AutismDevItemSummary item;
  final AutismDevAssessmentItem? detail;
  final String? selectedScore;
  final List<AutismDevScoreOption> scoreOptions;
  final VoidCallback onOpenQuestionPreference;
  final ValueChanged<String> onScore;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: _panelDecoration(),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 18, 16, 14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            SizedBox(
              height: 38,
              child: Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      '第 ${item.itemNo} 项  ${_displayItemTitle(item)}',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _AutismDevColors.ink,
                        fontSize: 23,
                        height: 1,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  _QuestionPreferenceChip(
                    onTap: onOpenQuestionPreference,
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            Expanded(
              child: ListView(
                padding: EdgeInsets.zero,
                physics: const BouncingScrollPhysics(),
                children: <Widget>[
                  _AutismDevItemDetailCard(item: item, detail: detail),
                ],
              ),
            ),
            const SizedBox(height: 10),
            _AutismDevScoreDock(
              scoreOptions: scoreOptions,
              selectedScore: selectedScore,
              onScore: onScore,
            ),
          ],
        ),
      ),
    );
  }
}
