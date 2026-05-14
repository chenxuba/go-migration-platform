part of 'autismdev_assessment_page.dart';

class _AutismDevDomainPanel extends StatefulWidget {
  const _AutismDevDomainPanel({
    required this.groups,
    required this.selectedDomainCode,
    required this.selectedRangeFilter,
    required this.selectedItemNo,
    required this.itemScores,
    required this.onSelectItem,
  });

  final List<AutismDevDomainGroup> groups;
  final String selectedDomainCode;
  final String selectedRangeFilter;
  final int selectedItemNo;
  final Map<int, String> itemScores;
  final ValueChanged<AutismDevItemSummary> onSelectItem;

  @override
  State<_AutismDevDomainPanel> createState() => _AutismDevDomainPanelState();
}

class _AutismDevDomainPanelState extends State<_AutismDevDomainPanel> {
  final ScrollController _scrollController = ScrollController();
  final GlobalKey _activeNavItemKey = GlobalKey();
  final Map<String, GlobalKey> _domainHeaderKeys = <String, GlobalKey>{};
  String _expandedDomainCode = '';

  @override
  void initState() {
    super.initState();
    _syncExpandedDomain();
    _keepActiveItemVisible();
  }

  @override
  void didUpdateWidget(covariant _AutismDevDomainPanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    _domainHeaderKeys
        .removeWhere((String code, GlobalKey key) => !_hasDomainCode(code));
    final bool selectedDomainChanged =
        oldWidget.selectedDomainCode != widget.selectedDomainCode;
    final bool selectedItemChanged =
        oldWidget.selectedItemNo != widget.selectedItemNo;
    if (selectedDomainChanged || !_hasDomainCode(_expandedDomainCode)) {
      _syncExpandedDomain();
    }
    if (selectedDomainChanged || selectedItemChanged) {
      _keepActiveItemVisible();
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  bool _hasDomainCode(String domainCode) {
    return widget.groups.any(
      (AutismDevDomainGroup group) => group.domainCode == domainCode,
    );
  }

  void _syncExpandedDomain() {
    _expandedDomainCode = widget.selectedDomainCode.trim();
  }

  void _toggleDomain(AutismDevDomainGroup group) {
    final String domainCode = group.domainCode;
    final bool opening = _expandedDomainCode != domainCode;
    setState(() {
      _expandedDomainCode = opening ? domainCode : '';
    });
    if (opening) {
      _keepDomainHeaderVisible(domainCode);
    }
  }

  void _collapseAllDomains() {
    if (_expandedDomainCode.isEmpty) {
      return;
    }
    setState(() => _expandedDomainCode = '');
  }

  void _keepDomainHeaderVisible(String domainCode) {
    WidgetsBinding.instance.addPostFrameCallback((Duration _) {
      if (!mounted ||
          !_scrollController.hasClients ||
          domainCode.trim().isEmpty) {
        return;
      }
      final BuildContext? headerContext =
          _domainHeaderKeys[domainCode]?.currentContext;
      if (headerContext == null) {
        return;
      }
      Scrollable.ensureVisible(
        headerContext,
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOut,
        alignment: .02,
        alignmentPolicy: ScrollPositionAlignmentPolicy.explicit,
      );
    });
  }

  void _keepActiveItemVisible() {
    WidgetsBinding.instance.addPostFrameCallback((Duration _) {
      if (!mounted ||
          !_scrollController.hasClients ||
          widget.selectedItemNo <= 0) {
        return;
      }
      final BuildContext? itemContext = _activeNavItemKey.currentContext;
      if (itemContext == null) {
        return;
      }
      Scrollable.ensureVisible(
        itemContext,
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOut,
        alignment: .34,
        alignmentPolicy: ScrollPositionAlignmentPolicy.explicit,
      );
    });
  }

  List<AutismDevItemSummary> _visibleItemsForGroup(
    AutismDevDomainGroup group,
  ) {
    final String rangeFilter = _rangeFilterForGroup(group);
    if (rangeFilter.isEmpty) {
      return group.items;
    }
    return group.items
        .where((AutismDevItemSummary item) =>
            _assessmentRangeBucket(item) == rangeFilter)
        .toList(growable: false);
  }

  String _rangeFilterForGroup(AutismDevDomainGroup group) {
    if (group.domainCode == widget.selectedDomainCode &&
        widget.selectedRangeFilter.trim().isNotEmpty) {
      return widget.selectedRangeFilter.trim();
    }
    return '';
  }

  Widget _buildDomainTile(AutismDevDomainGroup group) {
    final int done = group.items
        .where((AutismDevItemSummary item) =>
            widget.itemScores.containsKey(item.itemNo))
        .length;
    final List<AutismDevItemSummary> visibleItems =
        _visibleItemsForGroup(group);
    return _AutismDevDomainTile(
      group: group,
      headerKey: _domainHeaderKeys.putIfAbsent(
        group.domainCode,
        () => GlobalKey(
          debugLabel: 'autismdev-domain-header-${group.domainCode}',
        ),
      ),
      visibleItems: visibleItems,
      done: done,
      expanded: _expandedDomainCode == group.domainCode,
      selected: group.domainCode == widget.selectedDomainCode,
      itemScores: widget.itemScores,
      selectedItemNo: widget.selectedItemNo,
      activeItemKey: _activeNavItemKey,
      onTap: () => _toggleDomain(group),
      onTapItem: widget.onSelectItem,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: _panelDecoration(),
      child: Column(
        children: <Widget>[
          _AutismDevSidebarHeader(
            canCollapse: _expandedDomainCode.isNotEmpty,
            onCollapseAll: _collapseAllDomains,
          ),
          Expanded(
            child: SingleChildScrollView(
              controller: _scrollController,
              physics: const BouncingScrollPhysics(),
              child: Column(
                children: <Widget>[
                  for (final AutismDevDomainGroup group in widget.groups)
                    _buildDomainTile(group),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevDomainTile extends StatelessWidget {
  const _AutismDevDomainTile({
    required this.group,
    required this.headerKey,
    required this.visibleItems,
    required this.done,
    required this.expanded,
    required this.selected,
    required this.itemScores,
    required this.selectedItemNo,
    required this.activeItemKey,
    required this.onTap,
    required this.onTapItem,
  });

  final AutismDevDomainGroup group;
  final Key headerKey;
  final List<AutismDevItemSummary> visibleItems;
  final int done;
  final bool expanded;
  final bool selected;
  final Map<int, String> itemScores;
  final int selectedItemNo;
  final GlobalKey activeItemKey;
  final VoidCallback onTap;
  final ValueChanged<AutismDevItemSummary> onTapItem;

  @override
  Widget build(BuildContext context) {
    final Color color = _domainColor(group.domainCode);
    final double progress = group.itemCount <= 0
        ? 0
        : (done / group.itemCount).clamp(0, 1).toDouble();
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 10, 9),
      decoration: BoxDecoration(
        color: selected ? const Color(0xFFFFFBF8) : Colors.transparent,
        border: const Border(
          bottom: BorderSide(color: _AutismDevColors.lineSoft),
        ),
      ),
      child: Column(
        children: <Widget>[
          Material(
            key: headerKey,
            color: Colors.transparent,
            child: InkWell(
              onTap: onTap,
              borderRadius: BorderRadius.circular(8),
              child: Column(
                children: <Widget>[
                  SizedBox(
                    height: 28,
                    child: Row(
                      children: <Widget>[
                        Icon(
                          expanded
                              ? Icons.keyboard_arrow_down_rounded
                              : Icons.chevron_right_rounded,
                          size: 20,
                          color: _AutismDevColors.ink,
                        ),
                        const SizedBox(width: 4),
                        Container(
                          width: 9,
                          height: 9,
                          decoration: BoxDecoration(
                              color: color, shape: BoxShape.circle),
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            group.domainName,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: TextStyle(
                              color: selected
                                  ? _AutismDevColors.orangeDeep
                                  : _AutismDevColors.ink,
                              fontSize: 13,
                              height: 1,
                              fontWeight:
                                  selected ? FontWeight.w900 : FontWeight.w800,
                            ),
                          ),
                        ),
                        Text(
                          '$done/${group.itemCount}',
                          style: const TextStyle(
                            color: _AutismDevColors.muted,
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
                      const SizedBox(width: 28),
                      Expanded(
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(99),
                          child: LinearProgressIndicator(
                            value: progress,
                            minHeight: 4,
                            color: _AutismDevColors.orange.withOpacity(.46),
                            backgroundColor: const Color(0xFFF6EEE8),
                          ),
                        ),
                      ),
                      const SizedBox(width: 9),
                      SizedBox(
                        width: 34,
                        child: Text(
                          '${(progress * 100).round()}%',
                          textAlign: TextAlign.right,
                          style: const TextStyle(
                            color: _AutismDevColors.muted,
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
            for (final AutismDevItemSummary item in visibleItems)
              _AutismDevQuestionNavItem(
                key: item.itemNo == selectedItemNo
                    ? activeItemKey
                    : ValueKey<int>(item.itemNo),
                item: item,
                active: item.itemNo == selectedItemNo,
                done: itemScores.containsKey(item.itemNo),
                onTap: () => onTapItem(item),
              ),
          ],
        ],
      ),
    );
  }
}

class _AutismDevRangeQuickFilter extends StatelessWidget {
  const _AutismDevRangeQuickFilter({
    required this.group,
    required this.selectedRangeFilter,
    required this.itemScores,
    required this.onSelectRangeFilter,
  });

  final AutismDevDomainGroup? group;
  final String selectedRangeFilter;
  final Map<int, String> itemScores;
  final ValueChanged<String> onSelectRangeFilter;

  @override
  Widget build(BuildContext context) {
    final List<_AutismDevRangeOption> options =
        _rangeOptionsForGroup(group, itemScores);
    final int total = group?.itemCount ?? 0;
    final int done = group?.items
            .where((AutismDevItemSummary item) =>
                itemScores.containsKey(item.itemNo))
            .length ??
        0;
    final String title = (group?.domainName ?? '').trim();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        Row(
          children: <Widget>[
            Expanded(
              child: Text(
                title.isEmpty ? '分类' : title,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _AutismDevColors.ink,
                  fontSize: 16,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            const SizedBox(width: 8),
            Text(
              '${options.length}类',
              style: const TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 12,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        Expanded(
          child: options.isEmpty
              ? const Center(
                  child: Text(
                    '暂无分类',
                    style: TextStyle(
                      color: _AutismDevColors.muted,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                )
              : ClipRect(
                  child: Stack(
                    children: <Widget>[
                      Scrollbar(
                        child: ListView.separated(
                          padding: const EdgeInsets.only(bottom: 26),
                          physics: const BouncingScrollPhysics(),
                          itemBuilder: (BuildContext context, int index) {
                            if (index == 0) {
                              return _AutismDevRangeListItem(
                                label: '全部',
                                total: total,
                                done: done,
                                active: selectedRangeFilter.trim().isEmpty,
                                onTap: () => onSelectRangeFilter(''),
                              );
                            }
                            final _AutismDevRangeOption option =
                                options[index - 1];
                            return _AutismDevRangeListItem(
                              label: option.label,
                              total: option.total,
                              done: option.done,
                              active:
                                  selectedRangeFilter.trim() == option.label,
                              onTap: () => onSelectRangeFilter(option.label),
                            );
                          },
                          separatorBuilder: (BuildContext context, int index) =>
                              const SizedBox(height: 7),
                          itemCount: options.length + 1,
                        ),
                      ),
                      const Positioned(
                        left: 0,
                        right: 0,
                        bottom: 0,
                        child: IgnorePointer(
                          child: DecoratedBox(
                            decoration: BoxDecoration(
                              gradient: LinearGradient(
                                begin: Alignment.topCenter,
                                end: Alignment.bottomCenter,
                                colors: <Color>[
                                  Color(0x00FFFFFF),
                                  Color(0xF6FFFFFF),
                                ],
                              ),
                            ),
                            child: SizedBox(height: 30),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
        ),
      ],
    );
  }
}

class _AutismDevRangeListItem extends StatelessWidget {
  const _AutismDevRangeListItem({
    required this.label,
    required this.total,
    required this.done,
    required this.active,
    required this.onTap,
  });

  final String label;
  final int total;
  final int done;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Ink(
          height: 39,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color:
                active ? _AutismDevColors.orange : _AutismDevColors.softPanel,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(
              color: active ? _AutismDevColors.orange : _AutismDevColors.line,
            ),
          ),
          child: Row(
            children: <Widget>[
              Container(
                width: 7,
                height: 7,
                decoration: BoxDecoration(
                  color: active ? Colors.white : _AutismDevColors.orange,
                  shape: BoxShape.circle,
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  label,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: active ? Colors.white : _AutismDevColors.ink,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Container(
                height: 22,
                padding: const EdgeInsets.symmetric(horizontal: 7),
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: active ? Colors.white.withOpacity(.16) : Colors.white,
                  borderRadius: BorderRadius.circular(11),
                  border: Border.all(
                    color: active
                        ? Colors.white.withOpacity(.26)
                        : _AutismDevColors.lineSoft,
                  ),
                ),
                child: Text(
                  '$done/$total',
                  style: TextStyle(
                    color: active ? Colors.white : _AutismDevColors.body,
                    fontSize: 11,
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

class _AutismDevRangeOption {
  _AutismDevRangeOption(this.label);

  final String label;
  int total = 0;
  int done = 0;
}

class _AutismDevSidebarHeader extends StatelessWidget {
  const _AutismDevSidebarHeader({
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
        border: Border(bottom: BorderSide(color: _AutismDevColors.lineSoft)),
      ),
      child: Row(
        children: <Widget>[
          const Text(
            '领域任务',
            style: TextStyle(
              color: _AutismDevColors.ink,
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
                        ? _AutismDevColors.orangeDeep
                        : _AutismDevColors.muted.withOpacity(.38),
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

class _AutismDevQuestionNavItem extends StatelessWidget {
  const _AutismDevQuestionNavItem({
    required this.item,
    required this.active,
    required this.done,
    required this.onTap,
    super.key,
  });

  final AutismDevItemSummary item;
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
                  width: 42,
                  child: Text(
                    '第 ${item.itemNo}',
                    maxLines: 1,
                    style: TextStyle(
                      color: active
                          ? _AutismDevColors.orangeDeep
                          : _AutismDevColors.body,
                      fontSize: 12,
                      fontWeight: active ? FontWeight.w900 : FontWeight.w700,
                    ),
                  ),
                ),
                Expanded(
                  child: Text(
                    _displayItemTitle(item),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: active
                          ? _AutismDevColors.orangeDeep
                          : _AutismDevColors.body,
                      fontSize: 12,
                      fontWeight: active ? FontWeight.w900 : FontWeight.w700,
                    ),
                  ),
                ),
                const SizedBox(width: 7),
                if (done)
                  const Icon(
                    Icons.check_circle_rounded,
                    size: 17,
                    color: _AutismDevColors.green,
                  )
                else
                  Container(
                    width: 15,
                    height: 15,
                    decoration: BoxDecoration(
                      color: active ? _AutismDevColors.orange : Colors.white,
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: active
                            ? _AutismDevColors.orange
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
