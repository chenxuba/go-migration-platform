part of 'autismdev_assessment_page.dart';

class _AutismDevTopBar extends StatelessWidget {
  const _AutismDevTopBar({
    required this.title,
    required this.studentName,
    required this.studentAge,
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
  final String studentAge;
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
        border: Border.all(color: _AutismDevColors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _autismDevShadow(color: const Color(0x16B05F32), blur: 16),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1280;
          final List<Widget> headerChildren = <Widget>[
            Text(
              '$title 测评工作台',
              maxLines: 1,
              softWrap: false,
              style: const TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 23,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            _HeaderMeta(
              label: '儿童',
              value: studentName,
              compact: compact,
            ),
            _HeaderMeta(
              label: '年龄',
              value: studentAge,
              compact: compact,
            ),
            _HeaderMeta(
              label: compact ? '日期' : '测评日期',
              value: assessmentDate,
              compact: compact,
            ),
            _HeaderMeta(
              label: '施测者',
              value: examinerName,
              compact: compact,
            ),
          ];
          return Row(
            children: <Widget>[
              _HeaderIconButton(
                icon: Icons.chevron_left_rounded,
                onTap: onBack,
              ),
              const SizedBox(width: 10),
              Expanded(
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  physics: const ClampingScrollPhysics(),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: headerChildren,
                  ),
                ),
              ),
              if (autoSaveText.trim().isNotEmpty)
                _SaveStatusLabel(text: autoSaveText.trim(), saving: saving),
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

class _SaveStatusLabel extends StatelessWidget {
  const _SaveStatusLabel({required this.text, required this.saving});

  final String text;
  final bool saving;

  bool get _saving {
    return saving ||
        text.contains('保存中') ||
        text.contains('草稿保存中') ||
        text.contains('保存中');
  }

  bool get _failed => text.contains('失败');

  @override
  Widget build(BuildContext context) {
    final Color color = _failed
        ? _AutismDevColors.red
        : (_saving ? _AutismDevColors.orangeDeep : _AutismDevColors.body);
    return ConstrainedBox(
      constraints: const BoxConstraints(minWidth: 116),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.end,
        children: <Widget>[
          Icon(
            _failed
                ? Icons.error_outline_rounded
                : (_saving
                    ? Icons.sync_rounded
                    : Icons.check_circle_outline_rounded),
            color: _failed
                ? _AutismDevColors.red
                : (_saving
                    ? _AutismDevColors.orangeDeep
                    : _AutismDevColors.green),
            size: _saving ? 17 : 18,
          ),
          const SizedBox(width: 5),
          Text(
            text,
            maxLines: 1,
            softWrap: false,
            textAlign: TextAlign.right,
            style: TextStyle(
              color: color,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _HeaderMeta extends StatelessWidget {
  const _HeaderMeta({
    required this.label,
    required this.value,
    required this.compact,
  });

  final String label;
  final String value;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: EdgeInsets.only(left: compact ? 6 : 10),
      padding: EdgeInsets.only(left: compact ? 6 : 10),
      decoration: const BoxDecoration(
        border: Border(left: BorderSide(color: _AutismDevColors.line)),
      ),
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
        overflow: TextOverflow.ellipsis,
        style: const TextStyle(
          color: _AutismDevColors.body,
          fontSize: 13,
          height: 1,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

class _HeaderIconButton extends StatelessWidget {
  const _HeaderIconButton({required this.icon, required this.onTap});

  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Ink(
          width: 46,
          height: 46,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: _AutismDevColors.line),
          ),
          child: Icon(
            icon,
            color: _AutismDevColors.body,
            size: 34,
          ),
        ),
      ),
    );
  }
}

class _TopActionButton extends StatelessWidget {
  const _TopActionButton({
    required this.label,
    required this.icon,
    required this.loading,
    required this.filled,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final bool loading;
  final bool filled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: loading ? null : onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 36,
          padding: const EdgeInsets.symmetric(horizontal: 13),
          decoration: BoxDecoration(
            color: filled ? _AutismDevColors.orange : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _AutismDevColors.orange),
            boxShadow: filled
                ? _autismDevShadow(
                    color: const Color(0x28E96F43),
                    blur: 12,
                    offset: const Offset(0, 5),
                  )
                : null,
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              if (loading)
                SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    color: filled ? Colors.white : _AutismDevColors.orange,
                  ),
                )
              else
                Icon(
                  icon,
                  size: 17,
                  color: filled ? Colors.white : _AutismDevColors.orange,
                ),
              const SizedBox(width: 7),
              Text(
                label,
                style: TextStyle(
                  color: filled ? Colors.white : _AutismDevColors.orangeDeep,
                  fontSize: 14,
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
