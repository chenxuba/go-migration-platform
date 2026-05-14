part of 'autismdev_assessment_page.dart';

class _QuestionPreferenceChip extends StatelessWidget {
  const _QuestionPreferenceChip({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(9),
        child: Container(
          height: 30,
          padding: const EdgeInsets.symmetric(horizontal: 11),
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: const Color(0xFFFFF1E8),
            borderRadius: BorderRadius.circular(9),
            border: Border.all(color: const Color(0xFFFFC8AD)),
          ),
          child: const Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(
                Icons.tune_rounded,
                size: 15,
                color: _AutismDevColors.orangeDeep,
              ),
              SizedBox(width: 5),
              Text(
                '题目偏好配置',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  color: _AutismDevColors.orangeDeep,
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

class _QuestionPreferenceDialog extends StatefulWidget {
  const _QuestionPreferenceDialog({
    required this.selected,
    required this.studentAgeMonths,
  });

  final _AutismDevQuestionDisplayPreference selected;
  final int? studentAgeMonths;

  @override
  State<_QuestionPreferenceDialog> createState() =>
      _QuestionPreferenceDialogState();
}

class _QuestionPreferenceDialogState extends State<_QuestionPreferenceDialog> {
  late _AutismDevQuestionDisplayPreference _selected = widget.selected;

  void _confirm() {
    Navigator.of(context).pop(_selected);
  }

  @override
  Widget build(BuildContext context) {
    final String ageLabel = widget.studentAgeMonths == null
        ? '当前月龄未知'
        : '当前月龄：${widget.studentAgeMonths}月';
    return Center(
      child: SizedBox(
        width: 560,
        child: Material(
          color: Colors.white,
          borderRadius: BorderRadius.circular(14),
          child: Padding(
            padding: const EdgeInsets.fromLTRB(18, 18, 18, 14),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: <Widget>[
                const Text(
                  '题目偏好配置',
                  style: TextStyle(
                    color: _AutismDevColors.ink,
                    fontSize: 24,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 10),
                Text(
                  '$ageLabel，情绪与行为始终展示',
                  style: const TextStyle(
                    color: _AutismDevColors.body,
                    fontSize: 14,
                    height: 1.2,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 18),
                _PreferenceOptionTile(
                  title: '展示所有题目',
                  selected:
                      _selected == _AutismDevQuestionDisplayPreference.all,
                  onTap: () => setState(
                    () => _selected = _AutismDevQuestionDisplayPreference.all,
                  ),
                ),
                const SizedBox(height: 10),
                _PreferenceOptionTile(
                  title: '展示匹配月龄的题目',
                  selected: _selected ==
                      _AutismDevQuestionDisplayPreference.matchingAge,
                  onTap: () => setState(
                    () => _selected =
                        _AutismDevQuestionDisplayPreference.matchingAge,
                  ),
                ),
                const SizedBox(height: 10),
                _PreferenceOptionTile(
                  title: '展示匹配月龄及以下的题目',
                  selected: _selected ==
                      _AutismDevQuestionDisplayPreference.ageAndBelow,
                  onTap: () => setState(
                    () => _selected =
                        _AutismDevQuestionDisplayPreference.ageAndBelow,
                  ),
                ),
                const SizedBox(height: 18),
                Row(
                  children: <Widget>[
                    const Spacer(),
                    TextButton(
                      onPressed: () => Navigator.of(context).pop(),
                      child: const Text('取消'),
                    ),
                    const SizedBox(width: 10),
                    FilledButton(
                      onPressed: _confirm,
                      child: const Text('确认'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _PreferenceOptionTile extends StatelessWidget {
  const _PreferenceOptionTile({
    required this.title,
    required this.selected,
    required this.onTap,
  });

  final String title;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color accent =
        selected ? _AutismDevColors.orange : _AutismDevColors.line;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 52,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFF2EA) : const Color(0xFFFDF8F4),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: accent),
          ),
          child: Row(
            children: <Widget>[
              Container(
                width: 18,
                height: 18,
                decoration: BoxDecoration(
                  color: selected ? _AutismDevColors.orange : Colors.white,
                  shape: BoxShape.circle,
                  border: Border.all(color: accent, width: 1.5),
                ),
                child: selected
                    ? const Icon(
                        Icons.check_rounded,
                        size: 13,
                        color: Colors.white,
                      )
                    : null,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  title,
                  style: TextStyle(
                    color: selected
                        ? _AutismDevColors.orangeDeep
                        : _AutismDevColors.ink,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
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
