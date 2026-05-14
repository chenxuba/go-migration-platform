part of 'pep3_assessment_page.dart';

class _RailCard extends StatelessWidget {
  const _RailCard({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.9),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _Pep3Colors.line),
        boxShadow: _pep3Shadow(
          color: const Color(0x12B05F32),
          blur: 15,
          offset: const Offset(0, 7),
        ),
      ),
      child: child,
    );
  }
}

class _RecordFieldEditor extends StatelessWidget {
  const _RecordFieldEditor({
    super.key,
    required this.currentItemNo,
    required this.field,
    required this.value,
    required this.onChanged,
  });

  final int currentItemNo;
  final Pep3RecordField field;
  final dynamic value;
  final ValueChanged<dynamic> onChanged;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _Pep3Colors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            field.label.trim().isEmpty ? field.key : field.label,
            style: const TextStyle(
              color: _Pep3Colors.text,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 8),
          if (field.fieldType == 'radio')
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: <Widget>[
                for (final Pep3RecordFieldOption option in field.options)
                  ChoiceChip(
                    label: Text(option.label),
                    selected: '$value' == option.value,
                    selectedColor: const Color(0xFFFFEEE5),
                    onSelected: (_) => onChanged(option.value),
                  ),
              ],
            )
          else if (field.fieldType == 'checkbox_group')
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: <Widget>[
                for (final Pep3RecordFieldOption option in field.options)
                  FilterChip(
                    label: Text(option.label),
                    selected:
                        value is List && (value as List).contains(option.value),
                    selectedColor: const Color(0xFFFFEEE5),
                    onSelected: (bool selected) {
                      final List<String> next = value is List
                          ? (value as List)
                              .map((dynamic item) => '$item')
                              .toList()
                          : <String>[];
                      if (selected) {
                        next.add(option.value);
                      } else {
                        next.remove(option.value);
                      }
                      onChanged(next);
                    },
                  ),
              ],
            )
          else
            SizedBox(
              height: field.fieldType == 'textarea' ? null : 52,
              child: TextFormField(
                key: ValueKey<String>(
                  'pep3-record-input-$currentItemNo-${field.key}',
                ),
                initialValue: value == null ? '' : '$value',
                minLines: field.fieldType == 'textarea' ? 2 : 1,
                maxLines: field.fieldType == 'textarea' ? 4 : 1,
                keyboardType:
                    field.fieldType == 'number' ? TextInputType.number : null,
                textAlignVertical: field.fieldType == 'textarea'
                    ? TextAlignVertical.top
                    : TextAlignVertical.center,
                style: const TextStyle(
                  color: _Pep3Colors.ink,
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                ),
                decoration: InputDecoration(
                  isDense: true,
                  contentPadding: EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: field.fieldType == 'textarea' ? 10 : 14,
                  ),
                  hintText: field.placeholder.trim().isEmpty
                      ? '请输入'
                      : field.placeholder,
                  hintStyle: const TextStyle(
                    color: _Pep3Colors.muted,
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                  ),
                  filled: true,
                  fillColor: Colors.white,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: const BorderSide(color: _Pep3Colors.line),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: const BorderSide(color: _Pep3Colors.line),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: const BorderSide(color: _Pep3Colors.orange),
                  ),
                ),
                onChanged: (String text) => onChanged(
                  field.fieldType == 'number'
                      ? num.tryParse(text.trim())
                      : text,
                ),
              ),
            ),
        ],
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
            color: filled ? _Pep3Colors.orange : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _Pep3Colors.orange),
            boxShadow: filled
                ? _pep3Shadow(
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
                    color: filled ? Colors.white : _Pep3Colors.orange,
                  ),
                )
              else
                Icon(
                  icon,
                  size: 17,
                  color: filled ? Colors.white : _Pep3Colors.orange,
                ),
              const SizedBox(width: 7),
              Text(
                label,
                style: TextStyle(
                  color: filled ? Colors.white : _Pep3Colors.orangeDeep,
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

class _FooterButton extends StatelessWidget {
  const _FooterButton({
    required this.label,
    required this.icon,
    required this.enabled,
    required this.onTap,
    this.filled = false,
    this.reverseIcon = false,
  });

  final String label;
  final IconData icon;
  final bool enabled;
  final VoidCallback onTap;
  final bool filled;
  final bool reverseIcon;

  @override
  Widget build(BuildContext context) {
    final Color textColor = filled ? Colors.white : _Pep3Colors.orangeDeep;
    final List<Widget> children = <Widget>[
      Icon(icon, size: 22, color: textColor),
      const SizedBox(width: 8),
      Text(
        label,
        style: TextStyle(
          color: textColor,
          fontSize: 15,
          fontWeight: FontWeight.w900,
        ),
      ),
    ];
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: enabled ? onTap : null,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          width: filled ? 140 : 128,
          height: 38,
          decoration: BoxDecoration(
            color: filled
                ? _Pep3Colors.orange
                : enabled
                    ? Colors.white
                    : const Color(0xFFF7F1ED),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: enabled ? _Pep3Colors.orange : const Color(0xFFE2D6CE),
            ),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: reverseIcon ? children.reversed.toList() : children,
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
            border: Border.all(color: _Pep3Colors.line),
          ),
          child: Icon(icon, color: _Pep3Colors.text, size: 34),
        ),
      ),
    );
  }
}

class _DomainChip extends StatelessWidget {
  const _DomainChip({required this.code, required this.name});

  final String code;
  final String name;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 11),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF1E8),
        borderRadius: BorderRadius.circular(9),
        border: Border.all(color: const Color(0xFFFFC8AD)),
      ),
      child: Text(
        '${code.trim()} ${name.trim()}'.trim(),
        style: const TextStyle(
          color: _Pep3Colors.orangeDeep,
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _SaveBadge extends StatelessWidget {
  const _SaveBadge({required this.saving});

  final bool saving;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 26,
      padding: const EdgeInsets.symmetric(horizontal: 9),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: saving ? const Color(0xFFFFF7EA) : const Color(0xFFEAF4E5),
        borderRadius: BorderRadius.circular(13),
      ),
      child: Text(
        saving ? '保存中' : '已保存',
        style: TextStyle(
          color: saving ? _Pep3Colors.orangeDeep : _Pep3Colors.green,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _CaregiverActionButton extends StatelessWidget {
  const _CaregiverActionButton({
    required this.label,
    required this.icon,
    required this.filled,
    required this.loading,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final bool filled;
  final bool loading;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: loading ? null : onTap,
        borderRadius: BorderRadius.circular(9),
        child: Ink(
          height: 34,
          decoration: BoxDecoration(
            color: filled ? _Pep3Colors.orange : Colors.white,
            borderRadius: BorderRadius.circular(9),
            border: Border.all(
                color: filled ? _Pep3Colors.orange : _Pep3Colors.line),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              Icon(icon,
                  size: 17, color: filled ? Colors.white : _Pep3Colors.ink),
              const SizedBox(width: 7),
              Text(
                label,
                style: TextStyle(
                  color: filled ? Colors.white : _Pep3Colors.ink,
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
            color: _Pep3Colors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w800,
          ),
        ),
        const SizedBox(height: 5),
        Text(
          value,
          style: TextStyle(
            color: danger ? const Color(0xFFE04438) : _Pep3Colors.orangeDeep,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}

class _DonutPainter extends CustomPainter {
  const _DonutPainter({required this.percent});

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
      ..color = _Pep3Colors.orange
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
  bool shouldRepaint(covariant _DonutPainter oldDelegate) {
    return oldDelegate.percent != percent;
  }
}
