part of 'erxin_assessment_page.dart';

class _ErxinAllItemsOverviewDialog extends StatefulWidget {
  const _ErxinAllItemsOverviewDialog({
    required this.domains,
    required this.ageGroups,
    required this.itemPasses,
    required this.selectedItemNo,
    required this.mainAgeMonth,
    required this.onClose,
  });

  final List<ErxinDomain> domains;
  final List<ErxinAgeGroup> ageGroups;
  final Map<int, bool> itemPasses;
  final int selectedItemNo;
  final int mainAgeMonth;
  final VoidCallback onClose;

  @override
  State<_ErxinAllItemsOverviewDialog> createState() =>
      _ErxinAllItemsOverviewDialogState();
}

class _ErxinAllItemsOverviewDialogState
    extends State<_ErxinAllItemsOverviewDialog> {
  static const double _estimatedSectionExtent = 332;

  final Map<int, GlobalKey> _sectionKeys = <int, GlobalKey>{};
  List<ErxinAgeGroup> _groups = const <ErxinAgeGroup>[];
  List<List<ErxinAgeGroup>> _sections = const <List<ErxinAgeGroup>>[];
  List<ErxinDomain> _sortedDomains = const <ErxinDomain>[];
  int _answeredCount = 0;
  int _totalCount = 0;
  int _targetSectionIndex = 0;
  ScrollController? _scrollController;
  double _initialTargetOffset = 0;
  bool _contentReady = false;
  bool _dataReady = false;
  int _visibleSectionCount = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) {
        return;
      }
      _prepareOverviewData();
      setState(() {
        _dataReady = true;
        _contentReady = true;
        _visibleSectionCount = _initialVisibleSectionCount;
      });
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _expandVisibleSections();
      });
    });
  }

  @override
  void dispose() {
    _scrollController?.dispose();
    super.dispose();
  }

  void _prepareOverviewData() {
    _groups = widget.ageGroups
        .where((ErxinAgeGroup group) => group.items.isNotEmpty)
        .toList()
      ..sort((ErxinAgeGroup left, ErxinAgeGroup right) =>
          left.ageMonth.compareTo(right.ageMonth));
    _sections = <List<ErxinAgeGroup>>[];
    for (int index = 0; index < _groups.length; index += 5) {
      _sections.add(
        _groups.sublist(index, math.min(index + 5, _groups.length)),
      );
    }
    for (int index = 0; index < _sections.length; index++) {
      _sectionKeys[index] = GlobalKey(debugLabel: 'erxin-overview-$index');
    }
    _sortedDomains = widget.domains.toList()
      ..sort((ErxinDomain left, ErxinDomain right) =>
          left.sortNo.compareTo(right.sortNo));
    _answeredCount = widget.itemPasses.length;
    _totalCount = _groups.fold<int>(
      0,
      (int total, ErxinAgeGroup group) => total + group.items.length,
    );
    _targetSectionIndex = _sectionIndexForMonth(_targetAgeMonth);
    _initialTargetOffset = _estimatedInitialOffset;
    _scrollController?.dispose();
    _scrollController = ScrollController(
      initialScrollOffset: _initialTargetOffset,
    );
  }

  int get _targetAgeMonth {
    for (final ErxinAgeGroup group in _groups) {
      final bool containsSelected = group.items.any(
        (ErxinItemSummary item) => item.itemNo == widget.selectedItemNo,
      );
      if (containsSelected) {
        return group.ageMonth;
      }
    }
    return widget.mainAgeMonth;
  }

  int _sectionIndexForMonth(int month) {
    if (month <= 0) {
      return 0;
    }
    final int index = _sections.indexWhere(
      (List<ErxinAgeGroup> section) =>
          section.any((ErxinAgeGroup group) => group.ageMonth == month),
    );
    return index < 0 ? 0 : index;
  }

  double get _estimatedInitialOffset {
    if (_targetSectionIndex <= 0) {
      return 0;
    }
    return math.max(0, _targetSectionIndex * _estimatedSectionExtent - 16);
  }

  int get _initialVisibleSectionCount {
    if (_sections.isEmpty) {
      return 0;
    }
    return math.min(_sections.length, math.max(2, _targetSectionIndex + 2));
  }

  void _expandVisibleSections() {
    if (!mounted ||
        !_contentReady ||
        _visibleSectionCount >= _sections.length) {
      return;
    }
    Future<void>.delayed(const Duration(milliseconds: 16), () {
      if (!mounted || _visibleSectionCount >= _sections.length) {
        return;
      }
      setState(() {
        _visibleSectionCount = math.min(
          _sections.length,
          _visibleSectionCount + 2,
        );
      });
      _expandVisibleSections();
    });
  }

  @override
  Widget build(BuildContext context) {
    final ScrollController? scrollController = _scrollController;
    return Dialog(
      insetPadding: const EdgeInsets.symmetric(horizontal: 24, vertical: 20),
      backgroundColor: Colors.transparent,
      child: Container(
        width: 1330,
        height: 704,
        decoration: BoxDecoration(
          color: const Color(0xFFFFFEFC),
          borderRadius: BorderRadius.circular(12),
          boxShadow: _erxinShadow(
            color: const Color(0x33000000),
            blur: 30,
            offset: const Offset(0, 18),
          ),
        ),
        child: Column(
          children: <Widget>[
            Container(
              height: 58,
              padding: const EdgeInsets.fromLTRB(18, 0, 12, 0),
              decoration: const BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.vertical(top: Radius.circular(12)),
                border: Border(bottom: BorderSide(color: _ErxinColors.line)),
              ),
              child: Row(
                children: <Widget>[
                  const Text(
                    '全部题目总览',
                    style: TextStyle(
                      color: _ErxinColors.ink,
                      fontSize: 20,
                      height: 1,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const SizedBox(width: 14),
                  _OverviewChip(
                    text:
                        _dataReady ? '已测 $_answeredCount/$_totalCount' : '加载中',
                  ),
                  if (widget.mainAgeMonth > 0) ...<Widget>[
                    const SizedBox(width: 8),
                    _OverviewChip(text: '主测月龄 ${widget.mainAgeMonth}月龄'),
                  ],
                  const Spacer(),
                  IconButton(
                    tooltip: '关闭',
                    onPressed: widget.onClose,
                    icon: const Icon(Icons.close_rounded),
                    color: _ErxinColors.body,
                    iconSize: 24,
                  ),
                ],
              ),
            ),
            Expanded(
              child: _sections.isEmpty
                  ? _dataReady
                      ? const Center(
                          child: Text(
                            '暂无题目',
                            style: TextStyle(
                              color: _ErxinColors.body,
                              fontSize: 15,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                        )
                      : const _OverviewDeferredBody()
                  : !_contentReady
                      ? const _OverviewDeferredBody()
                      : scrollController == null
                          ? const _OverviewDeferredBody()
                          : Scrollbar(
                              thumbVisibility: true,
                              controller: scrollController,
                              child: ListView.builder(
                                controller: scrollController,
                                padding:
                                    const EdgeInsets.fromLTRB(18, 16, 18, 18),
                                cacheExtent: 720,
                                itemCount: _visibleSectionCount,
                                itemBuilder: (BuildContext context, int index) {
                                  return Padding(
                                    key: _sectionKeys[index],
                                    padding: EdgeInsets.only(
                                      bottom: index == _sections.length - 1
                                          ? 0
                                          : 14,
                                    ),
                                    child: _PaperOverviewSection(
                                      ageGroups: _sections[index],
                                      domains: _sortedDomains,
                                      itemPasses: widget.itemPasses,
                                      selectedItemNo: widget.selectedItemNo,
                                      mainAgeMonth: widget.mainAgeMonth,
                                      highlighted: index == _targetSectionIndex,
                                    ),
                                  );
                                },
                              ),
                            ),
            ),
          ],
        ),
      ),
    );
  }
}

class _OverviewDeferredBody extends StatelessWidget {
  const _OverviewDeferredBody();

  @override
  Widget build(BuildContext context) {
    return const SingleChildScrollView(
      padding: EdgeInsets.fromLTRB(18, 16, 18, 18),
      physics: NeverScrollableScrollPhysics(),
      child: Column(
        children: <Widget>[
          _PaperOverviewSkeletonSection(
            months: <int>[1, 2, 3, 4, 5],
            highlighted: false,
          ),
          SizedBox(height: 14),
          _PaperOverviewSkeletonSection(
            months: <int>[6, 7, 8, 9, 10],
            highlighted: false,
          ),
        ],
      ),
    );
  }
}

class _PaperOverviewSkeletonSection extends StatelessWidget {
  const _PaperOverviewSkeletonSection({
    required this.months,
    required this.highlighted,
  });

  final List<int> months;
  final bool highlighted;

  @override
  Widget build(BuildContext context) {
    return Container(
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: highlighted ? const Color(0xFFFFFCF8) : Colors.white,
        borderRadius: BorderRadius.circular(4),
        border: Border.all(
          color: highlighted ? _ErxinColors.orange : _ErxinColors.ink,
          width: highlighted ? 1.4 : 1.1,
        ),
      ),
      child: Table(
        columnWidths: <int, TableColumnWidth>{
          0: const FixedColumnWidth(82),
          for (int index = 0; index < months.length; index++)
            index + 1: const FlexColumnWidth(),
        },
        border: TableBorder(
          horizontalInside:
              const BorderSide(color: Color(0xFF2F241E), width: .6),
          verticalInside: const BorderSide(color: Color(0xFF2F241E), width: .6),
        ),
        defaultVerticalAlignment: TableCellVerticalAlignment.middle,
        children: <TableRow>[
          TableRow(
            decoration: const BoxDecoration(color: Color(0xFFFFFBF7)),
            children: <Widget>[
              const _PaperHeaderCell(text: '项目'),
              for (final int month in months)
                _PaperHeaderCell(text: '$month 月龄'),
            ],
          ),
          const TableRow(
            children: <Widget>[
              _PaperSkeletonDomainCell(text: '大 运 动'),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
            ],
          ),
          const TableRow(
            children: <Widget>[
              _PaperSkeletonDomainCell(text: '精细动作'),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 1),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 1),
            ],
          ),
          const TableRow(
            children: <Widget>[
              _PaperSkeletonDomainCell(text: '适应能力'),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
            ],
          ),
          const TableRow(
            children: <Widget>[
              _PaperSkeletonDomainCell(text: '语　　言'),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 1),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 1),
            ],
          ),
          const TableRow(
            children: <Widget>[
              _PaperSkeletonDomainCell(text: '社会行为'),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 2),
              _PaperSkeletonItemsCell(rowCount: 1),
            ],
          ),
        ],
      ),
    );
  }
}

class _PaperSkeletonDomainCell extends StatelessWidget {
  const _PaperSkeletonDomainCell({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(minHeight: 48),
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 8),
      alignment: Alignment.center,
      child: Text(
        text,
        textAlign: TextAlign.center,
        style: const TextStyle(
          color: _ErxinColors.ink,
          fontSize: 16,
          height: 1.15,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _PaperSkeletonItemsCell extends StatelessWidget {
  const _PaperSkeletonItemsCell({required this.rowCount});

  final int rowCount;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(minHeight: 48),
      padding: const EdgeInsets.fromLTRB(11, 8, 11, 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          for (int index = 0; index < rowCount; index++) ...<Widget>[
            Row(
              children: <Widget>[
                const _ErxinSkeletonBlock(width: 15, height: 15, radius: 3),
                const SizedBox(width: 5),
                Expanded(
                  child: _ErxinSkeletonBlock(
                    widthFactor: index.isEven ? .86 : .68,
                    height: 12,
                    radius: 5,
                  ),
                ),
              ],
            ),
            if (index < rowCount - 1) const SizedBox(height: 8),
          ],
        ],
      ),
    );
  }
}

class _OverviewChip extends StatelessWidget {
  const _OverviewChip({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF6EF),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: _ErxinColors.line),
      ),
      alignment: Alignment.center,
      child: Text(
        text,
        style: const TextStyle(
          color: _ErxinColors.body,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _PaperOverviewSection extends StatelessWidget {
  const _PaperOverviewSection({
    required this.ageGroups,
    required this.domains,
    required this.itemPasses,
    required this.selectedItemNo,
    required this.mainAgeMonth,
    required this.highlighted,
  });

  final List<ErxinAgeGroup> ageGroups;
  final List<ErxinDomain> domains;
  final Map<int, bool> itemPasses;
  final int selectedItemNo;
  final int mainAgeMonth;
  final bool highlighted;

  @override
  Widget build(BuildContext context) {
    return Container(
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: highlighted ? const Color(0xFFFFFCF8) : Colors.white,
        borderRadius: BorderRadius.circular(4),
        border: Border.all(
          color: highlighted ? _ErxinColors.orange : _ErxinColors.ink,
          width: highlighted ? 1.4 : 1.1,
        ),
      ),
      child: Table(
        columnWidths: <int, TableColumnWidth>{
          0: const FixedColumnWidth(82),
          for (int index = 0; index < ageGroups.length; index++)
            index + 1: const FlexColumnWidth(),
        },
        border: TableBorder(
          horizontalInside:
              const BorderSide(color: Color(0xFF2F241E), width: .6),
          verticalInside: const BorderSide(color: Color(0xFF2F241E), width: .6),
        ),
        defaultVerticalAlignment: TableCellVerticalAlignment.middle,
        children: <TableRow>[
          TableRow(
            decoration: const BoxDecoration(color: Color(0xFFFFFBF7)),
            children: <Widget>[
              const _PaperHeaderCell(text: '项目'),
              for (final ErxinAgeGroup group in ageGroups)
                _PaperHeaderCell(
                  text: '${group.ageMonth} 月龄',
                  highlighted: group.ageMonth == mainAgeMonth,
                ),
            ],
          ),
          for (final ErxinDomain domain in domains)
            TableRow(
              children: <Widget>[
                _PaperDomainCell(domain: domain),
                for (final ErxinAgeGroup group in ageGroups)
                  _PaperMonthItemsCell(
                    items: _itemsForDomain(group.items, domain.domainCode),
                    itemPasses: itemPasses,
                    selectedItemNo: selectedItemNo,
                  ),
              ],
            ),
        ],
      ),
    );
  }

  static List<ErxinItemSummary> _itemsForDomain(
    List<ErxinItemSummary> items,
    String domainCode,
  ) {
    return items
        .where((ErxinItemSummary item) => item.domainCode == domainCode)
        .toList()
      ..sort((ErxinItemSummary left, ErxinItemSummary right) =>
          left.itemNo.compareTo(right.itemNo));
  }
}

class _PaperHeaderCell extends StatelessWidget {
  const _PaperHeaderCell({required this.text, this.highlighted = false});

  final String text;
  final bool highlighted;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 34,
      color: highlighted ? const Color(0xFFFFF1E8) : null,
      alignment: Alignment.center,
      child: Text(
        text,
        style: TextStyle(
          color: highlighted ? _ErxinColors.orangeDeep : _ErxinColors.ink,
          fontSize: 15,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _PaperDomainCell extends StatelessWidget {
  const _PaperDomainCell({required this.domain});

  final ErxinDomain domain;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(minHeight: 48),
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 8),
      alignment: Alignment.center,
      child: Text(
        _paperDomainName(domain.domainName),
        textAlign: TextAlign.center,
        style: const TextStyle(
          color: _ErxinColors.ink,
          fontSize: 16,
          height: 1.15,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _PaperMonthItemsCell extends StatelessWidget {
  const _PaperMonthItemsCell({
    required this.items,
    required this.itemPasses,
    required this.selectedItemNo,
  });

  final List<ErxinItemSummary> items;
  final Map<int, bool> itemPasses;
  final int selectedItemNo;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(minHeight: 48),
      padding: const EdgeInsets.fromLTRB(8, 5, 8, 5),
      child: items.isEmpty
          ? const SizedBox.shrink()
          : Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: <Widget>[
                for (final ErxinItemSummary item in items)
                  _PaperOverviewItem(
                    item: item,
                    passed: itemPasses[item.itemNo],
                    selected: item.itemNo == selectedItemNo,
                  ),
              ],
            ),
    );
  }
}

class _PaperOverviewItem extends StatelessWidget {
  const _PaperOverviewItem({
    required this.item,
    required this.passed,
    required this.selected,
  });

  final ErxinItemSummary item;
  final bool? passed;
  final bool selected;

  @override
  Widget build(BuildContext context) {
    final Color statusColor = passed == null
        ? _ErxinColors.muted
        : passed!
            ? _ErxinColors.green
            : _ErxinColors.red;
    return ColoredBox(
      color: selected ? const Color(0xFFFFF4E8) : Colors.transparent,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 3, vertical: 3),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Padding(
              padding: const EdgeInsets.only(top: 1),
              child: Icon(
                passed == null
                    ? Icons.check_box_outline_blank_rounded
                    : passed!
                        ? Icons.check_box_rounded
                        : Icons.indeterminate_check_box_rounded,
                size: 15,
                color: statusColor,
              ),
            ),
            const SizedBox(width: 3),
            Expanded(
              child: Text.rich(
                TextSpan(
                  children: <InlineSpan>[
                    TextSpan(
                      text: '${item.itemNo} ',
                      style: const TextStyle(fontWeight: FontWeight.w900),
                    ),
                    TextSpan(text: item.itemTitle),
                    if (item.parentReportAllowed)
                      const TextSpan(
                        text: ' ᴿ',
                        style: TextStyle(
                          color: _ErxinColors.blue,
                          fontSize: 11,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                  ],
                ),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _ErxinColors.ink,
                  fontSize: 13,
                  height: 1.15,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

String _paperDomainName(String value) {
  final String text = value.trim();
  if (text.length <= 3) {
    return text;
  }
  if (text.contains('大运动')) {
    return '大 运 动';
  }
  if (text.contains('精细')) {
    return '精细动作';
  }
  if (text.contains('适应')) {
    return '适应能力';
  }
  if (text.contains('语言')) {
    return '语　　言';
  }
  if (text.contains('社会')) {
    return '社会行为';
  }
  return text;
}
