part of 'vbmapp_assessment_page.dart';

class _VbmappModuleRail extends StatefulWidget {
  const _VbmappModuleRail({
    required this.modules,
    required this.selectedCode,
    required this.selectedItemCode,
    required this.items,
    required this.answeredCount,
    required this.isAnswered,
    required this.onSelectModule,
    required this.onSelectItem,
  });

  final List<_VbmappModule> modules;
  final String selectedCode;
  final String selectedItemCode;
  final List<_VbmappItem> items;
  final Map<String, int> answeredCount;
  final bool Function(_VbmappItem item) isAnswered;
  final ValueChanged<String> onSelectModule;
  final ValueChanged<_VbmappItem> onSelectItem;

  @override
  State<_VbmappModuleRail> createState() => _VbmappModuleRailState();
}

class _VbmappModuleRailState extends State<_VbmappModuleRail> {
  final Map<String, Set<String>> _expandedDomainsByModule =
      <String, Set<String>>{};
  final GlobalKey _activeItemKey = GlobalKey();
  final Map<String, GlobalKey> _groupKeys = <String, GlobalKey>{};
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _ensureExpandedForSelection();
    _keepSelectedVisible();
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(covariant _VbmappModuleRail oldWidget) {
    super.didUpdateWidget(oldWidget);
    _pruneExpandedDomains();
    _ensureExpandedForSelection();
    if (oldWidget.selectedCode != widget.selectedCode ||
        oldWidget.selectedItemCode != widget.selectedItemCode ||
        oldWidget.items.length != widget.items.length) {
      _keepSelectedVisible();
    }
  }

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

  void _ensureExpandedForSelection() {
    final Set<String> expanded = _expandedDomainsFor(widget.selectedCode);
    _VbmappItem? selectedItem;
    for (final _VbmappItem item in widget.items) {
      if (item.itemCode == widget.selectedItemCode) {
        selectedItem = item;
        break;
      }
    }
    if (selectedItem != null && selectedItem.domainName.trim().isNotEmpty) {
      expanded.add(selectedItem.domainName.trim());
      return;
    }
    if (expanded.isEmpty && widget.items.isNotEmpty) {
      expanded.add(widget.items.first.domainName.trim());
    }
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
      if (item.itemCode == widget.selectedItemCode) {
        return item.domainName.trim();
      }
    }
    return '';
  }

  GlobalKey _groupKeyFor(String domainName) {
    return _groupKeys.putIfAbsent(domainName, () => GlobalKey());
  }

  void _keepSelectedVisible() {
    _keepGroupHeaderVisible(_selectedDomainName);
    _keepActiveItemVisible();
  }

  void _keepGroupHeaderVisible(String domainName) {
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
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOut,
        alignment: .02,
        alignmentPolicy: ScrollPositionAlignmentPolicy.explicit,
      );
    });
  }

  void _keepActiveItemVisible() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted ||
          !_scrollController.hasClients ||
          widget.selectedItemCode.trim().isEmpty) {
        return;
      }
      final BuildContext? itemContext = _activeItemKey.currentContext;
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

  List<_VbmappRailGroup> get _groups {
    final Map<String, List<_VbmappItem>> grouped =
        <String, List<_VbmappItem>>{};
    for (final _VbmappItem item in widget.items) {
      grouped.putIfAbsent(item.domainName, () => <_VbmappItem>[]).add(item);
    }
    return grouped.entries.map((MapEntry<String, List<_VbmappItem>> entry) {
      final List<_VbmappItem> groupItems = List<_VbmappItem>.from(entry.value);
      if (widget.selectedCode == 'milestones') {
        groupItems.sort((_VbmappItem a, _VbmappItem b) {
          final int stageCompare = _vbmappMilestoneNavOrder(a)
              .compareTo(_vbmappMilestoneNavOrder(b));
          if (stageCompare != 0) {
            return stageCompare;
          }
          return a.sequenceNo.compareTo(b.sequenceNo);
        });
      }
      return _VbmappRailGroup(title: entry.key, items: groupItems);
    }).toList(growable: false);
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
                                key: item.itemCode == widget.selectedItemCode
                                    ? _activeItemKey
                                    : ValueKey<String>(
                                        'vbmapp-nav-${item.itemCode}',
                                      ),
                                item: item,
                                selected:
                                    item.itemCode == widget.selectedItemCode,
                                answered: widget.isAnswered(item),
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
    required this.selected,
    required this.answered,
    required this.onTap,
  });

  final _VbmappItem item;
  final bool selected;
  final bool answered;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
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
              color:
                  selected ? accent.withOpacity(.55) : _VbmappColors.lineSoft,
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
                    color: selected ? _VbmappColors.ink : _VbmappColors.body,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              if (answered)
                Container(
                  width: 20,
                  height: 20,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: accent,
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(Icons.check, color: Colors.white, size: 14),
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
