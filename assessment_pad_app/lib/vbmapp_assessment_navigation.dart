part of 'vbmapp_assessment_page.dart';

class _VbmappModuleRail extends StatefulWidget {
  const _VbmappModuleRail({
    required this.modules,
    required this.selectedCode,
    required this.selectedItemCodeListenable,
    required this.items,
    required this.answeredCount,
    required this.isAnswered,
    required this.hasActiveObservation,
    required this.onSelectModule,
    required this.onSelectItem,
  });

  final List<_VbmappModule> modules;
  final String selectedCode;
  final ValueNotifier<String> selectedItemCodeListenable;
  final List<_VbmappItem> items;
  final Map<String, int> answeredCount;
  final bool Function(_VbmappItem item) isAnswered;
  final bool Function(_VbmappItem item) hasActiveObservation;
  final ValueChanged<String> onSelectModule;
  final ValueChanged<_VbmappItem> onSelectItem;

  @override
  State<_VbmappModuleRail> createState() => _VbmappModuleRailState();
}

class _VbmappModuleRailState extends State<_VbmappModuleRail> {
  final Map<String, Set<String>> _expandedDomainsByModule =
      <String, Set<String>>{};
  final Map<String, List<_VbmappRailGroup>> _groupCache =
      <String, List<_VbmappRailGroup>>{};
  final Map<String, GlobalKey> _groupKeys = <String, GlobalKey>{};
  final Map<String, GlobalKey> _itemKeys = <String, GlobalKey>{};
  final ScrollController _scrollController = ScrollController();
  String _lastSelectedItemCode = '';
  String _lastSelectedDomainName = '';

  @override
  void initState() {
    super.initState();
    _lastSelectedItemCode = _selectedItemCode;
    _lastSelectedDomainName = _selectedDomainName;
    widget.selectedItemCodeListenable.addListener(_handleSelectedItemChanged);
    _ensureExpandedForSelection();
    _keepSelectedVisible();
  }

  @override
  void dispose() {
    widget.selectedItemCodeListenable.removeListener(
      _handleSelectedItemChanged,
    );
    _scrollController.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(covariant _VbmappModuleRail oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.selectedItemCodeListenable !=
        widget.selectedItemCodeListenable) {
      oldWidget.selectedItemCodeListenable.removeListener(
        _handleSelectedItemChanged,
      );
      widget.selectedItemCodeListenable.addListener(_handleSelectedItemChanged);
    }
    _pruneExpandedDomains();
    _ensureExpandedForSelection();
    if (oldWidget.selectedCode != widget.selectedCode ||
        oldWidget.items.length != widget.items.length) {
      _lastSelectedItemCode = _selectedItemCode;
      _lastSelectedDomainName = _selectedDomainName;
      _keepSelectedVisible(ensureGroupHeader: false, animated: false);
    }
  }

  String get _selectedItemCode => widget.selectedItemCodeListenable.value;

  Set<String> _expandedDomainsFor(String moduleCode) {
    return _expandedDomainsByModule.putIfAbsent(
      moduleCode,
      () => <String>{},
    );
  }

  void _pruneExpandedDomains() {
    final Set<String> validDomains = widget.items
        .map(((_VbmappItem item) => item.domainName.trim()))
        .where((String value) => value.isNotEmpty)
        .toSet();
    _expandedDomainsFor(widget.selectedCode)
        .removeWhere((String name) => !validDomains.contains(name));
  }

  bool _ensureExpandedForSelection() {
    final Set<String> expanded = _expandedDomainsFor(widget.selectedCode);
    _VbmappItem? selectedItem;
    for (final _VbmappItem item in widget.items) {
      if (item.itemCode == _selectedItemCode) {
        selectedItem = item;
        break;
      }
    }
    if (selectedItem != null && selectedItem.domainName.trim().isNotEmpty) {
      return expanded.add(selectedItem.domainName.trim());
    }
    if (expanded.isEmpty && widget.items.isNotEmpty) {
      return expanded.add(widget.items.first.domainName.trim());
    }
    return false;
  }

  void _handleSelectedItemChanged() {
    if (_lastSelectedItemCode == _selectedItemCode) {
      return;
    }
    _lastSelectedItemCode = _selectedItemCode;
    if (!_hasSelectedItemInCurrentModule) {
      _lastSelectedDomainName = '';
      return;
    }
    final String nextDomainName = _selectedDomainName;
    final bool domainChanged = nextDomainName != _lastSelectedDomainName;
    _lastSelectedDomainName = nextDomainName;
    final bool expandedChanged = _ensureExpandedForSelection();
    if (expandedChanged) {
      setState(() {});
    }
    _keepSelectedVisible(ensureGroupHeader: domainChanged || expandedChanged);
  }

  void _toggleDomain(String domainName) {
    final String normalized = domainName.trim();
    if (normalized.isEmpty) {
      return;
    }
    setState(() {
      final Set<String> expanded = _expandedDomainsFor(widget.selectedCode);
      if (!expanded.add(normalized)) {
        expanded.remove(normalized);
      }
    });
    if (normalized == _selectedDomainName) {
      _keepSelectedVisible();
    }
  }

  String get _selectedDomainName {
    for (final _VbmappItem item in widget.items) {
      if (item.itemCode == _selectedItemCode) {
        return item.domainName.trim();
      }
    }
    return '';
  }

  bool get _hasSelectedItemInCurrentModule {
    for (final _VbmappItem item in widget.items) {
      if (item.itemCode == _selectedItemCode) {
        return true;
      }
    }
    return false;
  }

  GlobalKey _groupKeyFor(String domainName) {
    return _groupKeys.putIfAbsent(domainName, () => GlobalKey());
  }

  GlobalKey _itemKeyFor(String itemCode) {
    return _itemKeys.putIfAbsent(itemCode, () => GlobalKey());
  }

  void _keepSelectedVisible({
    bool ensureGroupHeader = true,
    bool animated = true,
  }) {
    if (ensureGroupHeader) {
      _keepGroupHeaderVisible(_selectedDomainName, animated: animated);
    }
    _keepActiveItemVisible(animated: animated);
  }

  void _keepGroupHeaderVisible(String domainName, {required bool animated}) {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted ||
          !_scrollController.hasClients ||
          domainName.trim().isEmpty) {
        return;
      }
      final BuildContext? groupContext = _groupKeys[domainName]?.currentContext;
      if (groupContext == null) {
        return;
      }
      Scrollable.ensureVisible(
        groupContext,
        duration: animated ? const Duration(milliseconds: 90) : Duration.zero,
        curve: Curves.easeOut,
        alignment: .02,
        alignmentPolicy: ScrollPositionAlignmentPolicy.explicit,
      );
    });
  }

  void _keepActiveItemVisible({required bool animated}) {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted ||
          !_scrollController.hasClients ||
          _selectedItemCode.trim().isEmpty) {
        return;
      }
      final BuildContext? itemContext =
          _itemKeys[_selectedItemCode]?.currentContext;
      if (itemContext == null) {
        return;
      }
      Scrollable.ensureVisible(
        itemContext,
        duration: animated ? const Duration(milliseconds: 90) : Duration.zero,
        curve: Curves.easeOut,
        alignment: .34,
        alignmentPolicy: ScrollPositionAlignmentPolicy.explicit,
      );
    });
  }

  List<_VbmappRailGroup> get _groups {
    return _groupCache.putIfAbsent(
      widget.selectedCode,
      () => _buildRailGroups(widget.items),
    );
  }

  List<_VbmappRailGroup> _buildRailGroups(List<_VbmappItem> items) {
    final Map<String, List<_VbmappItem>> grouped =
        <String, List<_VbmappItem>>{};
    final List<String> domainOrder = <String>[];
    for (final _VbmappItem item in items) {
      if (!grouped.containsKey(item.domainName)) {
        domainOrder.add(item.domainName);
      }
      grouped.putIfAbsent(item.domainName, () => <_VbmappItem>[]).add(item);
    }
    return List<_VbmappRailGroup>.unmodifiable(
      domainOrder.map((String domainName) {
        final List<_VbmappItem> groupItems =
            grouped[domainName] ?? const <_VbmappItem>[];
        return _VbmappRailGroup(
          title: domainName,
          items: List<_VbmappItem>.unmodifiable(groupItems),
        );
      }),
    );
  }

  @override
  Widget build(BuildContext context) {
    final Set<String> expanded = _expandedDomainsFor(widget.selectedCode);
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: _vbmappCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          for (final _VbmappModule module in widget.modules) ...<Widget>[
            _VbmappModuleTile(
              module: module,
              selected: module.code == widget.selectedCode,
              answered: widget.answeredCount[module.code] ?? 0,
              onTap: () => widget.onSelectModule(module.code),
            ),
            const SizedBox(height: 6),
          ],
          const SizedBox(height: 2),
          const Divider(height: 1, color: _VbmappColors.lineSoft),
          const SizedBox(height: 8),
          Expanded(
            child: SingleChildScrollView(
              controller: _scrollController,
              child: Column(
                children: <Widget>[
                  for (int index = 0;
                      index < _groups.length;
                      index++) ...<Widget>[
                    _VbmappDomainGroupTile(
                      key: _groupKeyFor(_groups[index].title),
                      title: _groups[index].title,
                      subtitle: _vbmappRailGroupSubtitle(
                        widget.selectedCode,
                        _groups[index].items,
                      ),
                      answered: _groups[index]
                          .items
                          .where((_VbmappItem item) => widget.isAnswered(item))
                          .length,
                      total: _groups[index].items.length,
                      expanded: expanded.contains(_groups[index].title),
                      onTap: () => _toggleDomain(_groups[index].title),
                    ),
                    if (expanded.contains(_groups[index].title)) ...<Widget>[
                      const SizedBox(height: 6),
                      Padding(
                        padding: const EdgeInsets.only(left: 8),
                        child: Column(
                          children: <Widget>[
                            for (final _VbmappItem item
                                in _groups[index].items) ...<Widget>[
                              _VbmappItemNavTile(
                                key: _itemKeyFor(item.itemCode),
                                item: item,
                                selectedItemCodeListenable:
                                    widget.selectedItemCodeListenable,
                                answered: widget.isAnswered(item),
                                timing: widget.hasActiveObservation(item),
                                onTap: () => widget.onSelectItem(item),
                              ),
                              const SizedBox(height: 6),
                            ],
                          ],
                        ),
                      ),
                    ] else
                      const SizedBox(height: 6),
                  ],
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappRailGroup {
  const _VbmappRailGroup({required this.title, required this.items});

  final String title;
  final List<_VbmappItem> items;
}

class _VbmappModuleTile extends StatelessWidget {
  const _VbmappModuleTile({
    required this.module,
    required this.selected,
    required this.answered,
    required this.onTap,
  });

  final _VbmappModule module;
  final bool selected;
  final int answered;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color accent = selected ? module.color : _VbmappColors.body;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
          decoration: BoxDecoration(
            color: selected ? module.color.withOpacity(.12) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color:
                  selected ? module.color.withOpacity(.55) : _VbmappColors.line,
            ),
          ),
          child: Row(
            children: <Widget>[
              Icon(module.icon, size: 19, color: accent),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  module.title,
                  maxLines: 1,
                  softWrap: false,
                  style: TextStyle(
                    color: selected ? _VbmappColors.ink : _VbmappColors.body,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: selected
                      ? module.color.withOpacity(.14)
                      : const Color(0xFFFFF6EF),
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(
                    color: selected
                        ? module.color.withOpacity(.3)
                        : _VbmappColors.lineSoft,
                  ),
                ),
                child: Text(
                  '$answered/${module.itemCount}',
                  style: TextStyle(
                    color: selected ? module.color : _VbmappColors.body,
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

class _VbmappDomainGroupTile extends StatelessWidget {
  const _VbmappDomainGroupTile({
    super.key,
    required this.title,
    required this.subtitle,
    required this.answered,
    required this.total,
    required this.expanded,
    required this.onTap,
  });

  final String title;
  final String subtitle;
  final int answered;
  final int total;
  final bool expanded;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
          decoration: BoxDecoration(
            color: expanded ? const Color(0xFFFFF3E8) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _VbmappColors.lineSoft),
          ),
          child: Row(
            children: <Widget>[
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text(
                      title,
                      maxLines: 1,
                      softWrap: false,
                      style: const TextStyle(
                        color: _VbmappColors.ink,
                        fontSize: 13,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    if (subtitle.trim().isNotEmpty) ...<Widget>[
                      const SizedBox(height: 2),
                      Text(
                        subtitle,
                        maxLines: 1,
                        softWrap: false,
                        style: const TextStyle(
                          color: _VbmappColors.muted,
                          fontSize: 11,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFF6EF),
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(color: _VbmappColors.lineSoft),
                ),
                child: Text(
                  '$answered/$total',
                  style: const TextStyle(
                    color: _VbmappColors.body,
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 6),
              Icon(
                expanded
                    ? Icons.expand_less_rounded
                    : Icons.expand_more_rounded,
                size: 20,
                color: _VbmappColors.body,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappItemNavTile extends StatelessWidget {
  const _VbmappItemNavTile({
    super.key,
    required this.item,
    required this.selectedItemCodeListenable,
    required this.answered,
    required this.timing,
    required this.onTap,
  });

  final _VbmappItem item;
  final ValueNotifier<String> selectedItemCodeListenable;
  final bool answered;
  final bool timing;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return ValueListenableBuilder<String>(
      valueListenable: selectedItemCodeListenable,
      builder: (BuildContext context, String selectedItemCode, Widget? child) {
        final bool selected = item.itemCode == selectedItemCode;
        final Color accent = item.color;
        return Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: onTap,
            borderRadius: BorderRadius.circular(10),
            child: Ink(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
              decoration: BoxDecoration(
                color: selected ? accent.withOpacity(.12) : Colors.white,
                borderRadius: BorderRadius.circular(10),
                border: Border.all(
                  color: selected
                      ? accent.withOpacity(.55)
                      : _VbmappColors.lineSoft,
                ),
              ),
              child: Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      item.navCode,
                      maxLines: 1,
                      softWrap: false,
                      style: TextStyle(
                        color:
                            selected ? _VbmappColors.ink : _VbmappColors.body,
                        fontSize: 15,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  if (timing) ...<Widget>[
                    const SizedBox(width: 6),
                    Container(
                      width: 20,
                      height: 20,
                      alignment: Alignment.center,
                      decoration: BoxDecoration(
                        color: accent.withOpacity(.12),
                        shape: BoxShape.circle,
                        border: Border.all(color: accent.withOpacity(.28)),
                      ),
                      child: Icon(
                        Icons.schedule_rounded,
                        color: accent,
                        size: 13,
                      ),
                    ),
                  ],
                  const SizedBox(width: 6),
                  if (answered)
                    Container(
                      width: 20,
                      height: 20,
                      alignment: Alignment.center,
                      decoration: BoxDecoration(
                        color: accent,
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.check,
                        color: Colors.white,
                        size: 14,
                      ),
                    )
                  else
                    Container(
                      width: 20,
                      height: 20,
                      alignment: Alignment.center,
                      decoration: BoxDecoration(
                        color: const Color(0xFFFFF6EF),
                        shape: BoxShape.circle,
                        border: Border.all(color: _VbmappColors.line),
                      ),
                    ),
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}

String _vbmappRailGroupSubtitle(
  String moduleCode,
  List<_VbmappItem> items,
) {
  if (items.isEmpty) {
    return '';
  }
  if (moduleCode == 'milestones') {
    final String first = items.first.navCode;
    final String last = items.last.navCode;
    return first == last ? first : '$first - $last';
  }
  return '${items.length}项';
}

int _vbmappMilestoneNavOrder(_VbmappItem item) {
  final RegExpMatch? match = RegExp(r'^(\d+)M$').firstMatch(item.navCode);
  if (match == null) {
    return item.sequenceNo;
  }
  return int.tryParse(match.group(1) ?? '') ?? item.sequenceNo;
}
