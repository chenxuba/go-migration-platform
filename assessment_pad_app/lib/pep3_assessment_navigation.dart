part of 'pep3_assessment_page.dart';

class _Pep3PageSidebar extends StatelessWidget {
  const _Pep3PageSidebar({
    required this.groups,
    required this.expandedGroupKey,
    required this.itemScores,
    required this.currentItemNo,
    required this.controller,
    required this.activeItemKey,
    required this.groupKeys,
    required this.onCollapseAll,
    required this.onToggleGroup,
    required this.onTapItem,
  });

  final List<Pep3ItemGroupSummary> groups;
  final String expandedGroupKey;
  final Map<int, int> itemScores;
  final int currentItemNo;
  final ScrollController controller;
  final GlobalKey activeItemKey;
  final Map<String, GlobalKey> groupKeys;
  final VoidCallback onCollapseAll;
  final ValueChanged<String> onToggleGroup;
  final ValueChanged<int> onTapItem;

  @override
  Widget build(BuildContext context) {
    return _RailCard(
      child: Column(
        children: <Widget>[
          _SidebarHeader(
            canCollapse: expandedGroupKey.isNotEmpty,
            onCollapseAll: onCollapseAll,
          ),
          Expanded(
            child: SingleChildScrollView(
              controller: controller,
              physics: const BouncingScrollPhysics(),
              child: Column(
                children: <Widget>[
                  for (final Pep3ItemGroupSummary group in groups)
                    _buildGroup(group),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildGroup(Pep3ItemGroupSummary group) {
    final bool expanded = group.key == expandedGroupKey;
    final int done = group.items
        .where((Pep3ItemSummary item) => itemScores.containsKey(item.itemNo))
        .length;
    final int total = group.items.length;
    final int percent = total == 0 ? 0 : ((done / total) * 100).round();
    return _PageGroup(
      key: groupKeys.putIfAbsent(
        group.key,
        () => GlobalKey(debugLabel: 'pep3-page-group-${group.key}'),
      ),
      group: group,
      expanded: expanded,
      done: done,
      total: total,
      percent: percent,
      itemScores: itemScores,
      currentItemNo: currentItemNo,
      activeItemKey: activeItemKey,
      onToggle: () => onToggleGroup(group.key),
      onTapItem: onTapItem,
    );
  }
}

class _SidebarHeader extends StatelessWidget {
  const _SidebarHeader({
    required this.canCollapse,
    required this.onCollapseAll,
  });

  final bool canCollapse;
  final VoidCallback onCollapseAll;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 48,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: _Pep3Colors.lineSoft)),
      ),
      child: Row(
        children: <Widget>[
          const Text(
            '记录册页面',
            style: TextStyle(
              color: _Pep3Colors.ink,
              fontSize: 16,
              fontWeight: FontWeight.w900,
            ),
          ),
          const Spacer(),
          Tooltip(
            message: '全部收起',
            child: Material(
              color: Colors.transparent,
              child: InkWell(
                onTap: canCollapse ? onCollapseAll : null,
                borderRadius: BorderRadius.circular(8),
                child: Ink(
                  width: 30,
                  height: 30,
                  decoration: BoxDecoration(
                    color: canCollapse
                        ? const Color(0xFFFFF7F2)
                        : Colors.transparent,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(
                      color: canCollapse
                          ? const Color(0xFFFFD7C4)
                          : Colors.transparent,
                    ),
                  ),
                  child: Icon(
                    Icons.unfold_less_rounded,
                    size: 19,
                    color: canCollapse
                        ? _Pep3Colors.orangeDeep
                        : _Pep3Colors.muted.withOpacity(.38),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _PageGroup extends StatelessWidget {
  const _PageGroup({
    super.key,
    required this.group,
    required this.expanded,
    required this.done,
    required this.total,
    required this.percent,
    required this.itemScores,
    required this.currentItemNo,
    required this.activeItemKey,
    required this.onToggle,
    required this.onTapItem,
  });

  final Pep3ItemGroupSummary group;
  final bool expanded;
  final int done;
  final int total;
  final int percent;
  final Map<int, int> itemScores;
  final int currentItemNo;
  final GlobalKey activeItemKey;
  final VoidCallback onToggle;
  final ValueChanged<int> onTapItem;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 10, 9),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: _Pep3Colors.lineSoft)),
      ),
      child: Column(
        children: <Widget>[
          Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: onToggle,
              borderRadius: BorderRadius.circular(8),
              child: Column(
                children: <Widget>[
                  SizedBox(
                    height: 26,
                    child: Row(
                      children: <Widget>[
                        Icon(
                          expanded
                              ? Icons.keyboard_arrow_down_rounded
                              : Icons.chevron_right_rounded,
                          size: 20,
                          color: _Pep3Colors.ink,
                        ),
                        const SizedBox(width: 4),
                        Expanded(
                          child: Text(
                            group.displayTitle,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              color: _Pep3Colors.ink,
                              fontSize: 13,
                              height: 1,
                              fontWeight: FontWeight.w900,
                            ),
                          ),
                        ),
                        Text(
                          '$done/$total 题',
                          style: const TextStyle(
                            color: _Pep3Colors.muted,
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 7),
                  Row(
                    children: <Widget>[
                      const SizedBox(width: 26),
                      Expanded(
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(99),
                          child: LinearProgressIndicator(
                            value: percent / 100,
                            minHeight: 4,
                            color: _Pep3Colors.orange.withOpacity(.46),
                            backgroundColor: const Color(0xFFF6EEE8),
                          ),
                        ),
                      ),
                      const SizedBox(width: 9),
                      SizedBox(
                        width: 34,
                        child: Text(
                          '$percent%',
                          textAlign: TextAlign.right,
                          style: const TextStyle(
                            color: _Pep3Colors.muted,
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          if (expanded) ...<Widget>[
            const SizedBox(height: 7),
            for (final Pep3ItemSummary item in group.items)
              _QuestionNavItem(
                key: item.itemNo == currentItemNo ? activeItemKey : null,
                item: item,
                active: item.itemNo == currentItemNo,
                done: itemScores.containsKey(item.itemNo),
                onTap: () => onTapItem(item.itemNo),
              ),
          ],
        ],
      ),
    );
  }
}

class _QuestionNavItem extends StatelessWidget {
  const _QuestionNavItem({
    super.key,
    required this.item,
    required this.active,
    required this.done,
    required this.onTap,
  });

  final Pep3ItemSummary item;
  final bool active;
  final bool done;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 18, top: 3),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(10),
          child: Ink(
            height: 31,
            padding: const EdgeInsets.symmetric(horizontal: 9),
            decoration: BoxDecoration(
              color: active ? const Color(0xFFFFEEE5) : Colors.transparent,
              borderRadius: BorderRadius.circular(10),
            ),
            child: Row(
              children: <Widget>[
                SizedBox(
                  width: 48,
                  child: Text(
                    '第 ${item.itemNo} 题',
                    maxLines: 1,
                    style: TextStyle(
                      color: active ? _Pep3Colors.orangeDeep : _Pep3Colors.text,
                      fontSize: 12,
                      fontWeight: active ? FontWeight.w900 : FontWeight.w700,
                    ),
                  ),
                ),
                Expanded(
                  child: Text(
                    item.displayTitle,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: active ? _Pep3Colors.orangeDeep : _Pep3Colors.text,
                      fontSize: 12,
                      fontWeight: active ? FontWeight.w900 : FontWeight.w700,
                    ),
                  ),
                ),
                const SizedBox(width: 7),
                if (done)
                  const Icon(Icons.check_circle_rounded,
                      size: 17, color: _Pep3Colors.green)
                else
                  Container(
                    width: 15,
                    height: 15,
                    decoration: BoxDecoration(
                      color: active ? _Pep3Colors.orange : Colors.white,
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: active
                            ? _Pep3Colors.orange
                            : const Color(0xFFCAB8AA),
                      ),
                    ),
                    child: active
                        ? const Center(
                            child: DecoratedBox(
                              decoration: BoxDecoration(
                                color: Colors.white,
                                shape: BoxShape.circle,
                              ),
                              child: SizedBox(width: 5, height: 5),
                            ),
                          )
                        : null,
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
