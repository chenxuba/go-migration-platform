part of 'erxin_assessment_page.dart';

class _Header extends StatelessWidget {
  const _Header({
    required this.args,
    required this.autoSaveText,
    required this.saving,
    required this.submitting,
    this.actionsEnabled = true,
    required this.onBack,
    required this.onSave,
    required this.onSubmit,
  });

  final ErxinAssessmentLaunchArgs args;
  final String autoSaveText;
  final bool saving;
  final bool submitting;
  final bool actionsEnabled;
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
        border: Border.all(color: _ErxinColors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _erxinShadow(color: const Color(0x12172033), blur: 14),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1260;
          return Row(
            children: <Widget>[
              _HeaderIconButton(
                icon: Icons.chevron_left_rounded,
                onTap: onBack,
              ),
              const SizedBox(width: 10),
              SizedBox(
                width: compact ? 282 : 306,
                child: FittedBox(
                  fit: BoxFit.scaleDown,
                  alignment: Alignment.centerLeft,
                  child: const Text(
                    '儿心量表-II 测评工作台',
                    maxLines: 1,
                    softWrap: false,
                    style: TextStyle(
                      color: _ErxinColors.ink,
                      fontSize: 22,
                      height: 1,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Row(
                  children: <Widget>[
                    SizedBox(
                      width: compact ? 104 : 118,
                      child: _HeaderMeta(
                        label: '儿童',
                        value: args.studentName.trim().isEmpty
                            ? '-'
                            : args.studentName.trim(),
                      ),
                    ),
                    Expanded(
                      flex: compact ? 4 : 5,
                      child: _HeaderMeta(
                        label: '出生日期',
                        value: args.birthDate.trim().isEmpty
                            ? '-'
                            : args.birthDate.trim(),
                      ),
                    ),
                    Expanded(
                      flex: compact ? 4 : 5,
                      child: _HeaderMeta(
                        label: '测查日期',
                        value: args.assessmentDate.trim().isEmpty
                            ? '-'
                            : args.assessmentDate.trim(),
                      ),
                    ),
                    Expanded(
                      flex: compact ? 4 : 5,
                      child: _HeaderMeta(
                        label: '实足年龄',
                        value: args.studentAge.trim().isEmpty
                            ? '-'
                            : args.studentAge.trim(),
                      ),
                    ),
                  ],
                ),
              ),
              if (autoSaveText.trim().isNotEmpty)
                SizedBox(
                  width: compact ? 86 : 106,
                  child: Text(
                    autoSaveText,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    textAlign: TextAlign.right,
                    style: const TextStyle(
                      color: _ErxinColors.muted,
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
                enabled: actionsEnabled,
                onTap: onSave,
              ),
              const SizedBox(width: 9),
              _TopActionButton(
                label: '提交记录',
                icon: Icons.fact_check_outlined,
                loading: submitting,
                filled: true,
                enabled: actionsEnabled,
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
      margin: const EdgeInsets.only(left: 8),
      padding: const EdgeInsets.only(left: 8),
      decoration: const BoxDecoration(
        border: Border(left: BorderSide(color: _ErxinColors.line)),
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
          softWrap: false,
          style: const TextStyle(
            color: _ErxinColors.body,
            fontSize: 13,
            height: 1,
            fontWeight: FontWeight.w700,
          ),
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
            border: Border.all(color: _ErxinColors.line),
          ),
          child: Icon(
            icon,
            color: _ErxinColors.body,
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
    this.enabled = true,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final bool loading;
  final bool filled;
  final bool enabled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: loading || !enabled ? null : onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 36,
          padding: const EdgeInsets.symmetric(horizontal: 13),
          decoration: BoxDecoration(
            color: filled
                ? (enabled
                    ? _ErxinColors.orange
                    : _ErxinColors.orange.withOpacity(.45))
                : (enabled ? Colors.white : Colors.white.withOpacity(.72)),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: enabled
                  ? _ErxinColors.orange
                  : _ErxinColors.orange.withOpacity(.45),
            ),
            boxShadow: filled
                ? enabled
                    ? _erxinShadow(
                        color: const Color(0x28E96F43),
                        blur: 12,
                        offset: const Offset(0, 5),
                      )
                    : null
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
                    color: filled
                        ? Colors.white
                        : (enabled
                            ? _ErxinColors.orange
                            : _ErxinColors.orange.withOpacity(.45)),
                  ),
                )
              else
                Icon(
                  icon,
                  size: 17,
                  color: filled
                      ? Colors.white
                      : (enabled
                          ? _ErxinColors.orange
                          : _ErxinColors.orange.withOpacity(.45)),
                ),
              const SizedBox(width: 7),
              Text(
                label,
                maxLines: 1,
                softWrap: false,
                style: TextStyle(
                  color: filled
                      ? Colors.white
                      : (enabled
                          ? _ErxinColors.orangeDeep
                          : _ErxinColors.orangeDeep.withOpacity(.45)),
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
