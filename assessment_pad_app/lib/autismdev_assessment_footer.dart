part of 'autismdev_assessment_page.dart';

class _AutismDevFooter extends StatelessWidget {
  const _AutismDevFooter({
    required this.current,
    required this.total,
    required this.hasPrevious,
    required this.hasNext,
    required this.autoNext,
    required this.onPrevious,
    required this.onNext,
    required this.onJumpMissing,
    required this.onToggleAutoNext,
  });

  final int current;
  final int total;
  final bool hasPrevious;
  final bool hasNext;
  final bool autoNext;
  final VoidCallback onPrevious;
  final VoidCallback onNext;
  final VoidCallback onJumpMissing;
  final ValueChanged<bool> onToggleAutoNext;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 62,
      padding: const EdgeInsets.symmetric(horizontal: 20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.97),
        border: Border.all(color: _AutismDevColors.line),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
        boxShadow: _autismDevShadow(color: const Color(0x14B05F32), blur: 16),
      ),
      child: Row(
        children: <Widget>[
          _FooterButton(
            label: '上一题',
            icon: Icons.chevron_left_rounded,
            enabled: hasPrevious,
            onTap: onPrevious,
          ),
          const Spacer(),
          Text.rich(
            TextSpan(
              children: <InlineSpan>[
                TextSpan(
                  text: '$current',
                  style: const TextStyle(
                    color: _AutismDevColors.ink,
                    fontSize: 27,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                TextSpan(
                  text: ' / $total',
                  style: const TextStyle(
                    color: _AutismDevColors.body,
                    fontSize: 15,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          const Spacer(),
          _FooterButton(
            label: '下一题',
            icon: Icons.chevron_right_rounded,
            enabled: hasNext,
            filled: true,
            reverseIcon: true,
            onTap: onNext,
          ),
          const SizedBox(width: 14),
          _FooterButton(
            label: '跳到缺题',
            icon: Icons.swipe_right_alt_rounded,
            enabled: true,
            onTap: onJumpMissing,
          ),
          const SizedBox(width: 22),
          const Text(
            '自动下一题',
            style: TextStyle(
              color: _AutismDevColors.body,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(width: 8),
          Switch(
            value: autoNext,
            activeColor: _AutismDevColors.orange,
            onChanged: onToggleAutoNext,
          ),
        ],
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
    final Color textColor = enabled
        ? filled
            ? Colors.white
            : _AutismDevColors.orangeDeep
        : _AutismDevColors.muted;
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
            color: filled && enabled
                ? _AutismDevColors.orange
                : enabled
                    ? Colors.white
                    : const Color(0xFFF7F1ED),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color:
                  enabled ? _AutismDevColors.orange : const Color(0xFFE2D6CE),
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
