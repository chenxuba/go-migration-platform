part of 'pep3_assessment_page.dart';

class _Pep3Footer extends StatelessWidget {
  const _Pep3Footer({
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
        border: Border.all(color: _Pep3Colors.line),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
        boxShadow: _pep3Shadow(color: const Color(0x14B05F32), blur: 16),
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
                    color: _Pep3Colors.ink,
                    fontSize: 27,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                TextSpan(
                  text: ' / $total',
                  style: const TextStyle(
                    color: _Pep3Colors.text,
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
              color: _Pep3Colors.text,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(width: 8),
          Switch(
            value: autoNext,
            activeColor: _Pep3Colors.orange,
            onChanged: onToggleAutoNext,
          ),
        ],
      ),
    );
  }
}
