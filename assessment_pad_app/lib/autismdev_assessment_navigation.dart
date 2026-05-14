part of 'autismdev_assessment_page.dart';

class _AutismDevDomainPanel extends StatefulWidget {
  const _AutismDevDomainPanel({
    required this.groups,
    required this.allGroups,
    required this.selectedDomainCode,
    required this.selectedRangeFilter,
    required this.selectedItemNo,
    required this.itemScores,
    required this.editingScope,
    required this.draftScopeMode,
    required this.draftScopeDomainCodes,
    required this.scopeSummaryText,
    required this.onSelectItem,
    required this.onBeginScopeEdit,
    required this.onSelectDraftScopeMode,
    required this.onToggleDraftScopeDomain,
    required this.onCancelScopeEdit,
    required this.onApplyScopeEdit,
  });

  final List<AutismDevDomainGroup> groups;
  final List<AutismDevDomainGroup> allGroups;
  final String selectedDomainCode;
  final String selectedRangeFilter;
  final int selectedItemNo;
  final Map<int, String> itemScores;
  final bool editingScope;
  final _AutismDevAssessmentScopeMode draftScopeMode;
  final Set<String> draftScopeDomainCodes;
  final String scopeSummaryText;
  final ValueChanged<AutismDevItemSummary> onSelectItem;
  final VoidCallback onBeginScopeEdit;
  final ValueChanged<_AutismDevAssessmentScopeMode> onSelectDraftScopeMode;
  final ValueChanged<String> onToggleDraftScopeDomain;
  final VoidCallback onCancelScopeEdit;
  final VoidCallback onApplyScopeEdit;

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
    if (widget.editingScope) {
      return _AutismDevScopeEditorPanel(
        groups: widget.allGroups,
        itemScores: widget.itemScores,
        scopeMode: widget.draftScopeMode,
        selectedDomainCodes: widget.draftScopeDomainCodes,
        onSelectMode: widget.onSelectDraftScopeMode,
        onToggleDomain: widget.onToggleDraftScopeDomain,
        onCancel: widget.onCancelScopeEdit,
        onApply: widget.onApplyScopeEdit,
      );
    }
    return Container(
      decoration: _panelDecoration(),
      child: Column(
        children: <Widget>[
          _AutismDevSidebarHeader(
            canCollapse: _expandedDomainCode.isNotEmpty,
            scopeSummaryText: widget.scopeSummaryText,
            onCollapseAll: _collapseAllDomains,
            onEditScope: widget.onBeginScopeEdit,
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

class _AutismDevScopeEditorPanel extends StatelessWidget {
  const _AutismDevScopeEditorPanel({
    required this.groups,
    required this.itemScores,
    required this.scopeMode,
    required this.selectedDomainCodes,
    required this.onSelectMode,
    required this.onToggleDomain,
    required this.onCancel,
    required this.onApply,
  });

  final List<AutismDevDomainGroup> groups;
  final Map<int, String> itemScores;
  final _AutismDevAssessmentScopeMode scopeMode;
  final Set<String> selectedDomainCodes;
  final ValueChanged<_AutismDevAssessmentScopeMode> onSelectMode;
  final ValueChanged<String> onToggleDomain;
  final VoidCallback onCancel;
  final VoidCallback onApply;

  int get _selectedDomainCount =>
      scopeMode == _AutismDevAssessmentScopeMode.full
          ? groups.length
          : selectedDomainCodes.length;

  int get _selectedItemCount {
    if (scopeMode == _AutismDevAssessmentScopeMode.full) {
      return groups.fold<int>(
        0,
        (int total, AutismDevDomainGroup group) => total + group.itemCount,
      );
    }
    return groups
        .where((AutismDevDomainGroup group) =>
            selectedDomainCodes.contains(group.domainCode.trim()))
        .fold<int>(
          0,
          (int total, AutismDevDomainGroup group) => total + group.itemCount,
        );
  }

  @override
  Widget build(BuildContext context) {
    final bool custom = scopeMode == _AutismDevAssessmentScopeMode.custom;
    final bool canApply = !custom || selectedDomainCodes.isNotEmpty;
    return Container(
      decoration: _panelDecoration(),
      child: Column(
        children: <Widget>[
          Container(
            height: 48,
            padding: const EdgeInsets.symmetric(horizontal: 14),
            decoration: const BoxDecoration(
              border: Border(
                bottom: BorderSide(color: _AutismDevColors.lineSoft),
              ),
            ),
            child: Row(
              children: <Widget>[
                const Text(
                  '测评范围',
                  style: TextStyle(
                    color: _AutismDevColors.ink,
                    fontSize: 16,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const Spacer(),
                Text(
                  '$_selectedDomainCount领域 · $_selectedItemCount题',
                  style: const TextStyle(
                    color: _AutismDevColors.orangeDeep,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(12, 10, 12, 0),
            child: _ScopeModeSegmented(
              mode: scopeMode,
              onSelect: onSelectMode,
            ),
          ),
          const Padding(
            padding: EdgeInsets.fromLTRB(12, 10, 12, 0),
            child: _ScopeEditNotice(),
          ),
          Expanded(
            child: ListView.separated(
              padding: const EdgeInsets.fromLTRB(10, 10, 10, 8),
              physics: const BouncingScrollPhysics(),
              itemBuilder: (BuildContext context, int index) {
                final AutismDevDomainGroup group = groups[index];
                final int done = group.items
                    .where((AutismDevItemSummary item) =>
                        itemScores.containsKey(item.itemNo))
                    .length;
                final bool selected = !custom ||
                    selectedDomainCodes.contains(group.domainCode.trim());
                return _ScopeDomainRow(
                  group: group,
                  done: done,
                  selected: selected,
                  disabled: !custom,
                  onTap: () => onToggleDomain(group.domainCode),
                );
              },
              separatorBuilder: (BuildContext context, int index) =>
                  const SizedBox(height: 7),
              itemCount: groups.length,
            ),
          ),
          Container(
            height: 58,
            padding: const EdgeInsets.fromLTRB(10, 9, 10, 9),
            decoration: const BoxDecoration(
              border: Border(top: BorderSide(color: _AutismDevColors.lineSoft)),
            ),
            child: Row(
              children: <Widget>[
                Expanded(
                  child: _ScopeActionButton(
                    label: '取消',
                    filled: false,
                    enabled: true,
                    onTap: onCancel,
                  ),
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: _ScopeActionButton(
                    label: '应用范围',
                    filled: true,
                    enabled: canApply,
                    onTap: onApply,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ScopeModeSegmented extends StatelessWidget {
  const _ScopeModeSegmented({required this.mode, required this.onSelect});

  final _AutismDevAssessmentScopeMode mode;
  final ValueChanged<_AutismDevAssessmentScopeMode> onSelect;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 40,
      padding: const EdgeInsets.all(4),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF8F1),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _AutismDevColors.line),
      ),
      child: Row(
        children: <Widget>[
          Expanded(
            child: _ScopeModeButton(
              label: '全量',
              active: mode == _AutismDevAssessmentScopeMode.full,
              onTap: () => onSelect(_AutismDevAssessmentScopeMode.full),
            ),
          ),
          const SizedBox(width: 4),
          Expanded(
            child: _ScopeModeButton(
              label: '自定义',
              active: mode == _AutismDevAssessmentScopeMode.custom,
              onTap: () => onSelect(_AutismDevAssessmentScopeMode.custom),
            ),
          ),
        ],
      ),
    );
  }
}

class _ScopeModeButton extends StatelessWidget {
  const _ScopeModeButton({
    required this.label,
    required this.active,
    required this.onTap,
  });

  final String label;
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
          decoration: BoxDecoration(
            color: active ? _AutismDevColors.orange : Colors.transparent,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Center(
            child: Text(
              label,
              style: TextStyle(
                color: active ? Colors.white : _AutismDevColors.body,
                fontSize: 13,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _ScopeEditNotice extends StatelessWidget {
  const _ScopeEditNotice();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(10, 9, 10, 9),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF8E9),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: const Color(0xFFF0D6B0)),
      ),
      child: const Text(
        '自定义可只选 1 个领域。应用后工作台、缺题和进度只按所选领域计算。',
        style: TextStyle(
          color: Color(0xFF8D642B),
          fontSize: 12,
          height: 1.35,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _ScopeDomainRow extends StatelessWidget {
  const _ScopeDomainRow({
    required this.group,
    required this.done,
    required this.selected,
    required this.disabled,
    required this.onTap,
  });

  final AutismDevDomainGroup group;
  final int done;
  final bool selected;
  final bool disabled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Opacity(
      opacity: disabled ? .72 : 1,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: disabled ? null : onTap,
          borderRadius: BorderRadius.circular(8),
          child: Ink(
            padding: const EdgeInsets.fromLTRB(8, 9, 8, 9),
            decoration: BoxDecoration(
              color: selected ? const Color(0xFFFFF6EF) : Colors.white,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(
                color: selected
                    ? const Color(0xFFFFC8AD)
                    : _AutismDevColors.lineSoft,
              ),
            ),
            child: Row(
              children: <Widget>[
                Container(
                  width: 22,
                  height: 22,
                  decoration: BoxDecoration(
                    color: selected ? _AutismDevColors.orange : Colors.white,
                    borderRadius: BorderRadius.circular(7),
                    border: Border.all(
                      color: selected
                          ? _AutismDevColors.orange
                          : const Color(0xFFD9C7BB),
                      width: 1.5,
                    ),
                  ),
                  child: selected
                      ? const Icon(
                          Icons.check_rounded,
                          size: 16,
                          color: Colors.white,
                        )
                      : null,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        group.domainName,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          color: _AutismDevColors.ink,
                          fontSize: 13,
                          height: 1.1,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                      const SizedBox(height: 5),
                      Text(
                        '草稿进度 $done/${group.itemCount}',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          color: _AutismDevColors.muted,
                          fontSize: 10,
                          height: 1,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  '${group.itemCount}题',
                  style: const TextStyle(
                    color: _AutismDevColors.orangeDeep,
                    fontSize: 12,
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

class _ScopeActionButton extends StatelessWidget {
  const _ScopeActionButton({
    required this.label,
    required this.filled,
    required this.enabled,
    required this.onTap,
  });

  final String label;
  final bool filled;
  final bool enabled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color textColor = enabled
        ? filled
            ? Colors.white
            : _AutismDevColors.ink
        : _AutismDevColors.muted;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: enabled ? onTap : null,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          decoration: BoxDecoration(
            color: filled
                ? enabled
                    ? _AutismDevColors.orange
                    : _AutismDevColors.lineSoft
                : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: filled && enabled
                  ? _AutismDevColors.orange
                  : _AutismDevColors.line,
            ),
          ),
          child: Center(
            child: Text(
              label,
              style: TextStyle(
                color: textColor,
                fontSize: 14,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
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
    required this.scopeSummaryText,
    required this.onCollapseAll,
    required this.onEditScope,
  });

  final bool canCollapse;
  final String scopeSummaryText;
  final VoidCallback onCollapseAll;
  final VoidCallback onEditScope;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 48,
      padding: const EdgeInsets.symmetric(horizontal: 10),
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
          Text(
            scopeSummaryText,
            maxLines: 1,
            style: const TextStyle(
              color: _AutismDevColors.muted,
              fontSize: 11,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 6),
          Tooltip(
            message: '测评范围',
            child: Material(
              color: Colors.transparent,
              child: InkWell(
                onTap: onEditScope,
                borderRadius: BorderRadius.circular(8),
                child: Ink(
                  height: 28,
                  padding: const EdgeInsets.symmetric(horizontal: 10),
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFF2EA),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: _AutismDevColors.orange),
                  ),
                  child: const Center(
                    child: Text(
                      '范围',
                      maxLines: 1,
                      style: TextStyle(
                        color: _AutismDevColors.orangeDeep,
                        fontSize: 12,
                        height: 1,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(width: 6),
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
