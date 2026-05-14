part of 'erxin_assessment_page.dart';

class _AgeMonthSection extends StatelessWidget {
  const _AgeMonthSection({
    required this.month,
    required this.isMainAge,
    required this.flashing,
    required this.items,
    required this.itemPasses,
    required this.selectedItemNo,
    required this.itemKeyFor,
    required this.onSelectItem,
    required this.onScore,
    super.key,
  });

  final int month;
  final bool isMainAge;
  final bool flashing;
  final List<ErxinItemSummary> items;
  final Map<int, bool> itemPasses;
  final int selectedItemNo;
  final GlobalKey Function(int itemNo) itemKeyFor;
  final ValueChanged<int> onSelectItem;
  final void Function(int itemNo, bool passed) onScore;

  @override
  Widget build(BuildContext context) {
    final List<ErxinItemSummary> displayItems = items;
    final int answered = items
        .where((ErxinItemSummary item) => itemPasses.containsKey(item.itemNo))
        .length;
    return Container(
      margin: const EdgeInsets.only(bottom: 6),
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: flashing ? const Color(0xFFFFF3BF) : Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      foregroundDecoration: BoxDecoration(
        border: Border.all(
          color: flashing ? _ErxinColors.orange : _ErxinColors.line,
          width: flashing ? 1.4 : 1,
        ),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: <Widget>[
          Container(
            height: 34,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            decoration: BoxDecoration(
              color:
                  isMainAge ? const Color(0xFFFFF1E8) : const Color(0xFFFFFAF5),
              borderRadius: const BorderRadius.vertical(
                top: Radius.circular(8),
              ),
              border: const Border(
                bottom: BorderSide(color: _ErxinColors.line),
              ),
            ),
            child: Row(
              children: <Widget>[
                Text(
                  '$month月龄',
                  style: TextStyle(
                    color: isMainAge ? _ErxinColors.blue : _ErxinColors.ink,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                if (isMainAge) ...<Widget>[
                  const SizedBox(width: 8),
                  Container(
                    height: 22,
                    padding: const EdgeInsets.symmetric(horizontal: 8),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(999),
                      border: Border.all(color: const Color(0xFFFFC8AD)),
                    ),
                    child: const Center(
                      child: Text(
                        '主测月龄',
                        style: TextStyle(
                          color: _ErxinColors.blue,
                          fontSize: 11,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                  ),
                ],
                const Spacer(),
                Text(
                  '已测 $answered/${items.length}',
                  style: const TextStyle(
                    color: _ErxinColors.body,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          for (final MapEntry<int, ErxinItemSummary> entry
              in displayItems.asMap().entries)
            _ItemScoreRow(
              key: itemKeyFor(entry.value.itemNo),
              item: entry.value,
              selected: entry.value.itemNo == selectedItemNo,
              passed: itemPasses[entry.value.itemNo],
              showBottomDivider: entry.key < displayItems.length - 1,
              onTap: () => onSelectItem(entry.value.itemNo),
              onScore: (bool passed) => onScore(entry.value.itemNo, passed),
            ),
        ],
      ),
    );
  }
}

class _WorkspaceMonthDivider extends StatelessWidget {
  const _WorkspaceMonthDivider({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(4, 8, 4, 10),
      child: Row(
        children: <Widget>[
          Expanded(
            child: Container(
              height: 1,
              color: const Color(0xFFE8D6C8),
            ),
          ),
          Container(
            margin: const EdgeInsets.symmetric(horizontal: 10),
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            decoration: BoxDecoration(
              color: const Color(0xFFFFFAF5),
              borderRadius: BorderRadius.circular(999),
              border: Border.all(color: _ErxinColors.line),
            ),
            child: Text(
              label,
              style: const TextStyle(
                color: _ErxinColors.body,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          Expanded(
            child: Container(
              height: 1,
              color: const Color(0xFFE8D6C8),
            ),
          ),
        ],
      ),
    );
  }
}

class _ItemScoreRow extends StatelessWidget {
  const _ItemScoreRow({
    required this.item,
    required this.selected,
    required this.passed,
    required this.showBottomDivider,
    required this.onTap,
    required this.onScore,
    super.key,
  });

  final ErxinItemSummary item;
  final bool selected;
  final bool? passed;
  final bool showBottomDivider;
  final VoidCallback onTap;
  final ValueChanged<bool> onScore;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Container(
        height: 48,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFFBEB) : Colors.white,
          border: showBottomDivider
              ? const Border(bottom: BorderSide(color: _ErxinColors.line))
              : null,
        ),
        child: Row(
          children: <Widget>[
            SizedBox(
              width: 42,
              child: Text(
                '${item.itemNo}',
                style: const TextStyle(
                  color: _ErxinColors.body,
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
            Expanded(
              child: Row(
                children: <Widget>[
                  Flexible(
                    child: Text(
                      item.itemTitle,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _ErxinColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ),
                  if (item.parentReportAllowed) ...<Widget>[
                    const SizedBox(width: 8),
                    const _MiniMarker(text: 'R'),
                  ],
                  if (item.attentionIfFailed) ...<Widget>[
                    const SizedBox(width: 6),
                    const _MiniMarker(text: '*', warning: true),
                  ],
                ],
              ),
            ),
            _ScoreButton(
              label: '通过',
              selected: passed == true,
              color: _ErxinColors.green,
              icon: Icons.check_circle_rounded,
              onTap: () => onScore(true),
            ),
            const SizedBox(width: 8),
            _ScoreButton(
              label: '不通过',
              selected: passed == false,
              color: _ErxinColors.red,
              icon: Icons.cancel_rounded,
              onTap: () => onScore(false),
            ),
          ],
        ),
      ),
    );
  }
}

class _ScoreButton extends StatelessWidget {
  const _ScoreButton({
    required this.label,
    required this.selected,
    required this.color,
    required this.icon,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final Color color;
  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color selectedFill = Color.alphaBlend(
      color.withOpacity(.12),
      Colors.white,
    );
    final Color selectedBorder = Color.alphaBlend(
      color.withOpacity(.48),
      Colors.white,
    );
    final Color contentColor = selected ? color : _ErxinColors.body;
    return SizedBox(
      width: label.length > 2 ? 104 : 88,
      height: 34,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(8),
          child: Ink(
            decoration: BoxDecoration(
              color: selected ? selectedFill : Colors.white,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(
                color: selected ? selectedBorder : _ErxinColors.line,
              ),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: <Widget>[
                Icon(icon, size: 17, color: contentColor),
                const SizedBox(width: 6),
                Text(
                  label,
                  maxLines: 1,
                  softWrap: false,
                  overflow: TextOverflow.visible,
                  style: TextStyle(
                    color: contentColor,
                    fontSize: 13,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _DetailTextBox extends StatelessWidget {
  const _DetailTextBox({required this.title, required this.text});

  final String title;
  final String text;

  @override
  Widget build(BuildContext context) {
    final String normalizedText = _inlineDetailText(text);
    return Container(
      padding: const EdgeInsets.fromLTRB(10, 8, 10, 8),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            title,
            style: const TextStyle(
              color: _ErxinColors.ink,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 4),
          Expanded(
            child: Text(
              normalizedText.isEmpty ? '暂无内容' : normalizedText,
              maxLines: 4,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _ErxinColors.body,
                fontSize: 13,
                height: 1.28,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

String _inlineDetailText(String value) {
  return value
      .trim()
      .replaceAll(RegExp(r'[\r\n]+\s*'), '')
      .replaceAll(RegExp(r'\s+'), ' ')
      .replaceAllMapped(
        RegExp(r'([，。；、：！？])\s+'),
        (Match match) => match.group(1) ?? '',
      )
      .trim();
}
