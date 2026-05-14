part of 'pep3_assessment_page.dart';

class _Pep3Header extends StatelessWidget {
  const _Pep3Header({
    required this.title,
    required this.studentName,
    required this.age,
    required this.assessmentDate,
    required this.examinerName,
    required this.autoSaveText,
    required this.saving,
    required this.submitting,
    required this.onBack,
    required this.onSave,
    required this.onSubmit,
  });

  final String title;
  final String studentName;
  final String age;
  final String assessmentDate;
  final String examinerName;
  final String autoSaveText;
  final bool saving;
  final bool submitting;
  final VoidCallback onBack;
  final VoidCallback onSave;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 18),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        border: Border.all(color: _Pep3Colors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _pep3Shadow(color: const Color(0x16B05F32), blur: 16),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1120;
          return Row(
            children: <Widget>[
              _HeaderIconButton(
                  icon: Icons.chevron_left_rounded, onTap: onBack),
              const SizedBox(width: 10),
              SizedBox(
                width: compact ? 206 : 250,
                child: Text(
                  '$title 测评工作台',
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
              Expanded(
                child: Row(
                  children: <Widget>[
                    Expanded(
                        child: _HeaderMeta(label: '儿童', value: studentName)),
                    Expanded(child: _HeaderMeta(label: '年龄', value: age)),
                    Expanded(
                      flex: 2,
                      child: _HeaderMeta(label: '测评日期', value: assessmentDate),
                    ),
                    Expanded(
                      child: _HeaderMeta(label: '施测者', value: examinerName),
                    ),
                  ],
                ),
              ),
              if (autoSaveText.trim().isNotEmpty)
                SizedBox(
                  width: compact ? 82 : 112,
                  child: Text(
                    autoSaveText,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    textAlign: TextAlign.right,
                    style: const TextStyle(
                      color: _Pep3Colors.muted,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ),
              const SizedBox(width: 10),
              _TopActionButton(
                label: '保存草稿',
                icon: Icons.save_outlined,
                loading: saving,
                filled: false,
                onTap: onSave,
              ),
              const SizedBox(width: 9),
              _TopActionButton(
                label: '提交记录',
                icon: Icons.fact_check_outlined,
                loading: submitting,
                filled: true,
                onTap: onSubmit,
              ),
            ],
          );
        },
      ),
    );
  }
}

class _HeaderMeta extends StatelessWidget {
  const _HeaderMeta({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(left: 10),
      padding: const EdgeInsets.only(left: 10),
      decoration: const BoxDecoration(
        border: Border(left: BorderSide(color: _Pep3Colors.line)),
      ),
      child: FittedBox(
        fit: BoxFit.scaleDown,
        alignment: Alignment.centerLeft,
        child: Text.rich(
          TextSpan(
            children: <InlineSpan>[
              TextSpan(text: '$label：'),
              TextSpan(
                text: value,
                style: const TextStyle(fontWeight: FontWeight.w900),
              ),
            ],
          ),
          maxLines: 1,
          style: const TextStyle(
            color: _Pep3Colors.text,
            fontSize: 13,
            height: 1,
            fontWeight: FontWeight.w700,
          ),
        ),
      ),
    );
  }
}
