import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_scale_client.dart';
import 'pad_date_range_picker.dart';
import 'pep3_assessment_client.dart';

class AssessmentReportListScreen extends StatefulWidget {
  const AssessmentReportListScreen({
    required this.onBack,
    this.scaleClient = const ApiAssessmentScaleClient(),
    this.recordClient = const ApiPep3AssessmentClient(),
    super.key,
  });

  final VoidCallback onBack;
  final AssessmentScaleClient scaleClient;
  final Pep3AssessmentClient recordClient;

  @override
  State<AssessmentReportListScreen> createState() =>
      _AssessmentReportListScreenState();
}

class _AssessmentReportListScreenState
    extends State<AssessmentReportListScreen> {
  static const String _authTokenStorageKey = 'auth_token';

  late DateTimeRange _range;
  List<String> _categories = const <String>[];
  Map<String, int> _categoryCounts = const <String, int>{};
  String _selectedCategory = '';
  int _rangeTotal = 0;
  Pep3RecordPage _page = const Pep3RecordPage(
    items: <Pep3RecordSummary>[],
    total: 0,
    current: 1,
    size: 0,
  );
  String _searchKey = '';
  bool _listLoading = true;
  bool _categoryLoading = true;
  String _errorMessage = '';

  @override
  void initState() {
    super.initState();
    final DateTime today = _dateOnly(DateTime.now());
    _range = DateTimeRange(
      start: today.subtract(const Duration(days: 29)),
      end: today,
    );
    _loadData();
  }

  Future<void> _loadData({
    String? selectedCategory,
    bool reloadCategories = true,
  }) async {
    final bool shouldReloadCategories = reloadCategories || _categories.isEmpty;
    final bool shouldShowCategorySkeleton =
        shouldReloadCategories && _categories.isEmpty;
    setState(() {
      _listLoading = true;
      if (shouldShowCategorySkeleton) {
        _categoryLoading = true;
      }
      _errorMessage = '';
      if (selectedCategory != null) {
        _selectedCategory = selectedCategory;
      }
    });
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final List<String> categories = shouldReloadCategories
          ? await widget.scaleClient.fetchCategories(token)
          : _categories;
      final Pep3RecordPage page = await widget.recordClient.fetchRecordsPage(
        token,
        pageIndex: 1,
        pageSize: 50,
        assessmentCode: '',
        scaleCategory: _selectedCategory,
        searchKey: _searchKey,
        assessmentDateBegin: _dateText(_range.start),
        assessmentDateEnd: _dateText(_range.end),
      );
      Pep3RecordPage? allCountPage;
      Map<String, int>? counts;
      if (shouldReloadCategories) {
        allCountPage = _selectedCategory.isEmpty
            ? page
            : await widget.recordClient.fetchRecordsPage(
                token,
                pageIndex: 1,
                pageSize: 1,
                assessmentCode: '',
                searchKey: _searchKey,
                assessmentDateBegin: _dateText(_range.start),
                assessmentDateEnd: _dateText(_range.end),
              );
        counts = <String, int>{};
        for (final String category in categories) {
          final Pep3RecordPage countPage =
              await widget.recordClient.fetchRecordsPage(
            token,
            pageIndex: 1,
            pageSize: 1,
            assessmentCode: '',
            scaleCategory: category,
            searchKey: _searchKey,
            assessmentDateBegin: _dateText(_range.start),
            assessmentDateEnd: _dateText(_range.end),
          );
          counts[category] = countPage.total;
        }
      }
      if (!mounted) {
        return;
      }
      setState(() {
        if (shouldReloadCategories) {
          _categories = categories;
          _categoryCounts = counts ?? const <String, int>{};
          _rangeTotal = allCountPage?.total ?? page.total;
          _categoryLoading = false;
        }
        _page = page;
        _listLoading = false;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _listLoading = false;
        if (shouldReloadCategories) {
          _categoryLoading = false;
        }
        _errorMessage = '$error';
      });
    }
  }

  Future<void> _selectRange() async {
    final DateTime today = _dateOnly(DateTime.now());
    final DateTimeRange? picked = await showPadDateRangePicker(
      context: context,
      initialRange: _range,
      today: today,
      minDate: DateTime(today.year - 5),
      maxDate: DateTime(today.year + 1, 12, 31),
    );
    if (picked == null) {
      return;
    }
    setState(() {
      _range = DateTimeRange(
        start: _dateOnly(picked.start),
        end: _dateOnly(picked.end),
      );
    });
    await _loadData();
  }

  void _submitSearch(String value) {
    _searchKey = value.trim();
    _loadData();
  }

  @override
  Widget build(BuildContext context) {
    return _ReportTheme(
      child: _AssessmentReportListBody(
        onBack: widget.onBack,
        categories: _categories,
        categoryCounts: _categoryCounts,
        selectedCategory: _selectedCategory,
        records: _page.items,
        rangeTotal: _rangeTotal,
        total: _page.total,
        range: _range,
        categoryLoading: _categoryLoading,
        listLoading: _listLoading,
        errorMessage: _errorMessage,
        onRefresh: _loadData,
        onRangeTap: _selectRange,
        onSearchSubmitted: _submitSearch,
        onCategorySelected: (String category) => _loadData(
          selectedCategory: category,
          reloadCategories: false,
        ),
      ),
    );
  }
}

class _ReportTheme extends InheritedWidget {
  const _ReportTheme({required super.child});

  static const Color page = Color(0xFFFFF7EE);
  static const Color surface = Color(0xFFFFFDFA);
  static const Color ink = Color(0xFF3F2B22);
  static const Color text = Color(0xFF6F5B50);
  static const Color muted = Color(0xFFA7958B);
  static const Color line = Color(0xFFEAD7C9);
  static const Color lineSoft = Color(0xFFF4E8DF);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95D37);
  static const Color blue = Color(0xFF3F82D2);
  static const Color green = Color(0xFF6F9F70);
  static const Color amber = Color(0xFFD99427);
  static const Color rose = Color(0xFFD96A7F);
  static const Color violet = Color(0xFF7F77C8);

  @override
  bool updateShouldNotify(_ReportTheme oldWidget) => false;
}

class _AssessmentReportListBody extends StatelessWidget {
  const _AssessmentReportListBody({
    required this.onBack,
    required this.categories,
    required this.categoryCounts,
    required this.selectedCategory,
    required this.records,
    required this.rangeTotal,
    required this.total,
    required this.range,
    required this.categoryLoading,
    required this.listLoading,
    required this.errorMessage,
    required this.onRefresh,
    required this.onRangeTap,
    required this.onSearchSubmitted,
    required this.onCategorySelected,
  });

  final VoidCallback onBack;
  final List<String> categories;
  final Map<String, int> categoryCounts;
  final String selectedCategory;
  final List<Pep3RecordSummary> records;
  final int rangeTotal;
  final int total;
  final DateTimeRange range;
  final bool categoryLoading;
  final bool listLoading;
  final String errorMessage;
  final VoidCallback onRefresh;
  final VoidCallback onRangeTap;
  final ValueChanged<String> onSearchSubmitted;
  final ValueChanged<String> onCategorySelected;

  @override
  Widget build(BuildContext context) {
    return Builder(
      builder: (BuildContext context) {
        return ColoredBox(
          color: _ReportTheme.page,
          child: LayoutBuilder(
            builder: (BuildContext context, BoxConstraints constraints) {
              final double width = constraints.maxWidth;
              final double horizontalPadding = width >= 1200 ? 32 : 24;
              final double contentWidth = width - horizontalPadding * 2;
              final double sideWidth = width >= 1200 ? 214 : 198;
              final double gap = width >= 1200 ? 18 : 14;
              final double listWidth = contentWidth - sideWidth - gap;

              return Padding(
                padding: EdgeInsets.fromLTRB(
                  horizontalPadding,
                  31,
                  horizontalPadding,
                  42,
                ),
                child: Column(
                  children: <Widget>[
                    _TopBar(
                      onBack: onBack,
                      range: range,
                      onRangeTap: onRangeTap,
                      onRefresh: onRefresh,
                      onSearchSubmitted: onSearchSubmitted,
                    ),
                    const SizedBox(height: 30),
                    Expanded(
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: <Widget>[
                          SizedBox(
                            width: sideWidth,
                            child: _DomainPanel(
                              categories: categories,
                              counts: categoryCounts,
                              selectedCategory: selectedCategory,
                              total: rangeTotal,
                              loading: categoryLoading,
                              onSelected: onCategorySelected,
                            ),
                          ),
                          SizedBox(width: gap),
                          SizedBox(
                            width: listWidth,
                            child: _ReportListPanel(
                              records: records,
                              total: total,
                              loading: listLoading,
                              errorMessage: errorMessage,
                              onRetry: onRefresh,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              );
            },
          ),
        );
      },
    );
  }
}

class _TopBar extends StatelessWidget {
  const _TopBar({
    required this.onBack,
    required this.range,
    required this.onRangeTap,
    required this.onRefresh,
    required this.onSearchSubmitted,
  });

  final VoidCallback onBack;
  final DateTimeRange range;
  final VoidCallback onRangeTap;
  final VoidCallback onRefresh;
  final ValueChanged<String> onSearchSubmitted;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 42,
      child: Row(
        children: <Widget>[
          _BackButton(onTap: onBack),
          const SizedBox(width: 16),
          const Text(
            '评估报告',
            style: TextStyle(
              color: _ReportTheme.ink,
              fontSize: 25,
              height: 1.2,
              fontWeight: FontWeight.w900,
            ),
          ),
          const Spacer(),
          _SearchBox(onSubmitted: onSearchSubmitted),
          const SizedBox(width: 10),
          _ToolbarButton(
            label: '${_dateText(range.start)} - ${_dateText(range.end)}',
            icon: Icons.calendar_month_rounded,
            onTap: onRangeTap,
          ),
          const SizedBox(width: 10),
          _ToolbarButton(label: '刷新列表', filled: true, onTap: onRefresh),
        ],
      ),
    );
  }
}

class _BackButton extends StatelessWidget {
  const _BackButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: _ReportTheme.surface.withOpacity(.94),
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          width: 42,
          height: 42,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: _ReportTheme.line),
          ),
          child: const Icon(
            Icons.chevron_left_rounded,
            color: _ReportTheme.text,
            size: 28,
          ),
        ),
      ),
    );
  }
}

class _SearchBox extends StatelessWidget {
  const _SearchBox({required this.onSubmitted});

  final ValueChanged<String> onSubmitted;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 324,
      height: 42,
      decoration: BoxDecoration(
        color: _ReportTheme.surface.withOpacity(.94),
        borderRadius: BorderRadius.circular(13),
        border: Border.all(color: _ReportTheme.line),
      ),
      child: TextField(
        onSubmitted: onSubmitted,
        textInputAction: TextInputAction.search,
        style: const TextStyle(
          color: _ReportTheme.ink,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
        decoration: const InputDecoration(
          isDense: true,
          border: InputBorder.none,
          prefixIcon: Icon(
            Icons.search_rounded,
            size: 22,
            color: _ReportTheme.muted,
          ),
          prefixIconConstraints: BoxConstraints(minWidth: 42),
          hintText: '搜索儿童姓名 / 报告编号',
          hintStyle: TextStyle(
            color: _ReportTheme.muted,
            fontSize: 13,
            fontWeight: FontWeight.w800,
          ),
          contentPadding: EdgeInsets.fromLTRB(0, 13, 14, 12),
        ),
      ),
    );
  }
}

class _ToolbarButton extends StatelessWidget {
  const _ToolbarButton({
    required this.label,
    this.filled = false,
    this.icon,
    this.onTap,
  });

  final String label;
  final bool filled;
  final IconData? icon;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Container(
          height: 42,
          constraints: BoxConstraints(minWidth: filled ? 118 : 72),
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: filled ? _ReportTheme.orange : _ReportTheme.surface,
            borderRadius: BorderRadius.circular(13),
            border: filled ? null : Border.all(color: _ReportTheme.line),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              if (icon != null) ...<Widget>[
                Icon(
                  icon,
                  color: filled ? Colors.white : _ReportTheme.text,
                  size: 18,
                ),
                const SizedBox(width: 7),
              ],
              Text(
                label,
                style: TextStyle(
                  color: filled ? Colors.white : _ReportTheme.text,
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

class _DomainPanel extends StatelessWidget {
  const _DomainPanel({
    required this.categories,
    required this.counts,
    required this.selectedCategory,
    required this.total,
    required this.loading,
    required this.onSelected,
  });

  final List<String> categories;
  final Map<String, int> counts;
  final String selectedCategory;
  final int total;
  final bool loading;
  final ValueChanged<String> onSelected;

  @override
  Widget build(BuildContext context) {
    final List<_DomainItem> domains = <_DomainItem>[
      _DomainItem('全部分类', total, _ReportTheme.orange, ''),
      for (int index = 0; index < categories.length; index++)
        _DomainItem(
          categories[index],
          counts[categories[index]] ?? 0,
          _domainColor(index),
          categories[index],
        ),
    ];
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 16, 12, 16),
      decoration: BoxDecoration(
        color: _ReportTheme.surface.withOpacity(.94),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _ReportTheme.line),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x14C26B3E),
            blurRadius: 24,
            offset: Offset(0, 12),
          ),
        ],
      ),
      child: Column(
        children: <Widget>[
          const Padding(
            padding: EdgeInsets.symmetric(horizontal: 4),
            child: Row(
              children: <Widget>[
                Text(
                  '测评分类',
                  style: TextStyle(
                    color: _ReportTheme.ink,
                    fontSize: 17,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 14),
          Expanded(
            child: loading
                ? const _DomainSkeletonList()
                : ListView.separated(
                    padding: EdgeInsets.zero,
                    itemCount: domains.length,
                    separatorBuilder: (BuildContext context, int index) =>
                        const SizedBox(height: 7),
                    itemBuilder: (BuildContext context, int index) {
                      final _DomainItem item = domains[index];
                      return _DomainRow(
                        item: item,
                        selected: item.value == selectedCategory,
                        onTap: () => onSelected(item.value),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}

class _DomainSkeletonList extends StatelessWidget {
  const _DomainSkeletonList();

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final int rowCount = math.max(
          8,
          ((constraints.maxHeight + 7) / 54).ceil(),
        );
        return ListView.separated(
          padding: EdgeInsets.zero,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: rowCount,
          separatorBuilder: (BuildContext context, int index) =>
              const SizedBox(height: 7),
          itemBuilder: (BuildContext context, int index) =>
              const _DomainSkeletonRow(),
        );
      },
    );
  }
}

class _DomainSkeletonRow extends StatelessWidget {
  const _DomainSkeletonRow();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 47,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      child: const Row(
        children: <Widget>[
          _SkeletonBox(width: 10, height: 10, radius: 999),
          SizedBox(width: 10),
          _SkeletonBox(width: 86, height: 14, radius: 7),
          Spacer(),
          _SkeletonBox(width: 18, height: 12, radius: 6),
        ],
      ),
    );
  }
}

class _DomainRow extends StatelessWidget {
  const _DomainRow({
    required this.item,
    required this.selected,
    required this.onTap,
  });

  final _DomainItem item;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          height: 47,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFF0E7) : Colors.transparent,
            borderRadius: BorderRadius.circular(14),
          ),
          child: Row(
            children: <Widget>[
              Container(
                width: 10,
                height: 10,
                decoration:
                    BoxDecoration(color: item.color, shape: BoxShape.circle),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  item.label,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color:
                        selected ? _ReportTheme.orangeDeep : _ReportTheme.text,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              Text(
                '${item.count}',
                style: TextStyle(
                  color:
                      selected ? _ReportTheme.orangeDeep : _ReportTheme.muted,
                  fontSize: 12,
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

class _ReportListPanel extends StatelessWidget {
  const _ReportListPanel({
    required this.records,
    required this.total,
    required this.loading,
    required this.errorMessage,
    required this.onRetry,
  });

  final List<Pep3RecordSummary> records;
  final int total;
  final bool loading;
  final String errorMessage;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: _ReportTheme.surface.withOpacity(.96),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _ReportTheme.line),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x14C26B3E),
            blurRadius: 24,
            offset: Offset(0, 12),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
        child: Column(
          children: <Widget>[
            _ReportPanelHeader(total: total, loading: loading),
            const _ReportTableHeader(),
            if (loading)
              for (int index = 0; index < 6; index++) const _ReportSkeletonRow()
            else if (errorMessage.isNotEmpty)
              Expanded(
                child: _ReportState(
                  message: errorMessage,
                  actionLabel: '重试',
                  onAction: onRetry,
                ),
              )
            else if (records.isEmpty)
              const Expanded(child: _ReportState(message: '暂无评估报告'))
            else
              for (final Pep3RecordSummary record in records)
                _ReportRow(record: record),
          ],
        ),
      ),
    );
  }
}

class _ReportPanelHeader extends StatelessWidget {
  const _ReportPanelHeader({required this.total, required this.loading});

  final int total;
  final bool loading;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 63,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: _ReportTheme.line)),
      ),
      child: Row(
        children: <Widget>[
          const Text(
            '评估报告列表',
            style: TextStyle(
              color: _ReportTheme.ink,
              fontSize: 19,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 26),
          if (loading)
            const _MetricSkeletonChip()
          else
            _MetricChip(label: '近一月', value: '$total'),
        ],
      ),
    );
  }
}

class _MetricSkeletonChip extends StatelessWidget {
  const _MetricSkeletonChip();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      width: 86,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF8F2),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: const Row(
        children: <Widget>[
          _SkeletonBox(width: 38, height: 11, radius: 6),
          SizedBox(width: 8),
          _SkeletonBox(width: 18, height: 15, radius: 7),
        ],
      ),
    );
  }
}

class _MetricChip extends StatelessWidget {
  const _MetricChip({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF8F2),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          Text(
            label,
            style: const TextStyle(
              color: _ReportTheme.text,
              fontSize: 11,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 6),
          Text(
            value,
            style: const TextStyle(
              color: _ReportTheme.ink,
              fontSize: 16,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _ReportTableHeader extends StatelessWidget {
  const _ReportTableHeader();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      color: const Color(0xC7FFF8F2),
      child: const _ReportColumns(
        child: DefaultTextStyle(
          style: TextStyle(
            color: _ReportTheme.muted,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
          child: Row(
            children: <Widget>[
              _ColumnCell(flex: 250, child: Text('儿童信息')),
              _ColumnCell(flex: 220, trailingGap: 24, child: Text('测评量表')),
              _ColumnCell(flex: 130, child: Text('测评年龄')),
              _ColumnCell(flex: 145, child: Text('测评日期')),
              _ColumnCell(flex: 145, child: Text('报告时间')),
              _ColumnCell(
                flex: 168,
                trailingGap: 0,
                child:
                    Align(alignment: Alignment.centerRight, child: Text('操作')),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ReportRow extends StatelessWidget {
  const _ReportRow({required this.record});

  final Pep3RecordSummary record;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 73,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: _ReportTheme.lineSoft)),
      ),
      child: _ReportColumns(
        child: Row(
          children: <Widget>[
            _ColumnCell(flex: 250, child: _ChildInfo(record: record)),
            _ColumnCell(
              flex: 220,
              trailingGap: 24,
              child: _ScaleInfo(record: record),
            ),
            _ColumnCell(flex: 130, child: _PlainCell(_ageText(record))),
            _ColumnCell(
              flex: 145,
              child: _PlainCell(_dateOnlyText(record.assessmentDate)),
            ),
            _ColumnCell(
              flex: 145,
              child: _ReportTimeCell(_reportTimeRaw(record)),
            ),
            const _ColumnCell(
              flex: 168,
              trailingGap: 0,
              child: Align(
                alignment: Alignment.centerRight,
                child: _RowActions(),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ReportSkeletonRow extends StatelessWidget {
  const _ReportSkeletonRow();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 73,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: _ReportTheme.lineSoft)),
      ),
      child: const _ReportColumns(
        child: Row(
          children: <Widget>[
            _ColumnCell(flex: 250, child: _ChildInfoSkeleton()),
            _ColumnCell(
              flex: 220,
              trailingGap: 24,
              child: _ScaleInfoSkeleton(),
            ),
            _ColumnCell(flex: 130, child: _SkeletonTextCell(width: 58)),
            _ColumnCell(flex: 145, child: _SkeletonTextCell(width: 76)),
            _ColumnCell(flex: 145, child: _SkeletonTextCell(width: 86)),
            _ColumnCell(
              flex: 168,
              trailingGap: 0,
              child: Align(
                alignment: Alignment.centerRight,
                child: _ActionSkeleton(),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ChildInfoSkeleton extends StatelessWidget {
  const _ChildInfoSkeleton();

  @override
  Widget build(BuildContext context) {
    return const Row(
      children: <Widget>[
        _SkeletonBox(width: 38, height: 38, radius: 999),
        SizedBox(width: 11),
        Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                _SkeletonBox(width: 52, height: 15),
                SizedBox(width: 8),
                _SkeletonBox(width: 48, height: 22, radius: 999),
              ],
            ),
            SizedBox(height: 7),
            _SkeletonBox(width: 132, height: 11),
          ],
        ),
      ],
    );
  }
}

class _ScaleInfoSkeleton extends StatelessWidget {
  const _ScaleInfoSkeleton();

  @override
  Widget build(BuildContext context) {
    return const Column(
      mainAxisAlignment: MainAxisAlignment.center,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _SkeletonBox(width: 104, height: 14),
        SizedBox(height: 7),
        _SkeletonBox(width: 48, height: 22, radius: 999),
      ],
    );
  }
}

class _ActionSkeleton extends StatelessWidget {
  const _ActionSkeleton();

  @override
  Widget build(BuildContext context) {
    return const Row(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        _SkeletonBox(width: 54, height: 32, radius: 11),
        SizedBox(width: 8),
        _SkeletonBox(width: 54, height: 32, radius: 11),
      ],
    );
  }
}

class _ReportColumns extends StatelessWidget {
  const _ReportColumns({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return child;
  }
}

class _ColumnCell extends StatelessWidget {
  const _ColumnCell({
    required this.flex,
    required this.child,
    this.trailingGap = 12,
  });

  final int flex;
  final Widget child;
  final double trailingGap;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      flex: flex,
      child: Padding(
        padding: EdgeInsets.only(right: trailingGap),
        child: child,
      ),
    );
  }
}

class _ChildInfo extends StatelessWidget {
  const _ChildInfo({required this.record});

  final Pep3RecordSummary record;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Container(
          width: 38,
          height: 38,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: _avatarColor(record),
            shape: BoxShape.circle,
          ),
          child: Text(
            _studentInitial(record),
            style: const TextStyle(
              color: Colors.white,
              fontSize: 16,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
        const SizedBox(width: 11),
        Expanded(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Flexible(
                    child: Text(
                      _studentName(record),
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _ReportTheme.ink,
                        fontSize: 15,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  _Tag(
                    label: _assessmentCodeText(record.assessmentCode),
                    textColor: _codeColor(record),
                    bgColor: _codeColor(record).withOpacity(.12),
                  ),
                ],
              ),
              const SizedBox(height: 5),
              Text(
                _studentMeta(record),
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _ReportTheme.muted,
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _ScaleInfo extends StatelessWidget {
  const _ScaleInfo({required this.record});

  final Pep3RecordSummary record;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          record.assessmentName.trim().isEmpty ? '-' : record.assessmentName,
          overflow: TextOverflow.ellipsis,
          style: const TextStyle(
            color: _ReportTheme.ink,
            fontSize: 13,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 5),
        _Tag(
          label: _sequenceText(record.assessmentSequence),
          textColor: _attemptColor(record),
          bgColor: _attemptColor(record).withOpacity(.12),
        ),
      ],
    );
  }
}

class _PlainCell extends StatelessWidget {
  const _PlainCell(this.text);

  final String text;

  @override
  Widget build(BuildContext context) {
    return Text(
      text,
      overflow: TextOverflow.ellipsis,
      style: const TextStyle(
        color: _ReportTheme.text,
        fontSize: 13,
        fontWeight: FontWeight.w800,
      ),
    );
  }
}

class _ReportTimeCell extends StatelessWidget {
  const _ReportTimeCell(this.raw);

  final String raw;

  @override
  Widget build(BuildContext context) {
    final DateTime? parsed = _parseDateTime(raw);
    if (parsed == null) {
      return Text(
        raw.trim().isEmpty ? '-' : raw.trim(),
        maxLines: 2,
        overflow: TextOverflow.ellipsis,
        style: const TextStyle(
          color: _ReportTheme.text,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
      );
    }

    final DateTime local = parsed.toLocal();
    final String hour = local.hour.toString().padLeft(2, '0');
    final String minute = local.minute.toString().padLeft(2, '0');
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          _dateText(local),
          overflow: TextOverflow.ellipsis,
          style: const TextStyle(
            color: _ReportTheme.text,
            fontSize: 13,
            fontWeight: FontWeight.w800,
          ),
        ),
        const SizedBox(height: 3),
        Text(
          '$hour:$minute',
          overflow: TextOverflow.ellipsis,
          style: const TextStyle(
            color: _ReportTheme.muted,
            fontSize: 12,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    );
  }
}

class _Tag extends StatelessWidget {
  const _Tag({
    required this.label,
    required this.textColor,
    required this.bgColor,
  });

  final String label;
  final Color textColor;
  final Color bgColor;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 24,
      padding: const EdgeInsets.symmetric(horizontal: 9),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Center(
        widthFactor: 1,
        child: Text(
          label,
          style: TextStyle(
            color: textColor,
            fontSize: 12,
            height: 1,
            fontWeight: FontWeight.w900,
          ),
        ),
      ),
    );
  }
}

class _RowActions extends StatelessWidget {
  const _RowActions();

  @override
  Widget build(BuildContext context) {
    return const Row(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        _ActionButton(label: '查看', emphasized: true),
        SizedBox(width: 8),
        _ActionButton(label: '配置'),
      ],
    );
  }
}

class _ActionButton extends StatelessWidget {
  const _ActionButton({required this.label, this.emphasized = false});

  final String label;
  final bool emphasized;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 32,
      constraints: const BoxConstraints(minWidth: 54),
      alignment: Alignment.center,
      padding: const EdgeInsets.symmetric(horizontal: 11),
      decoration: BoxDecoration(
        color: emphasized ? const Color(0xFFFFF8F2) : _ReportTheme.surface,
        borderRadius: BorderRadius.circular(11),
        border: Border.all(
          color: emphasized ? const Color(0xFFF2CDBB) : _ReportTheme.line,
        ),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: emphasized ? _ReportTheme.orangeDeep : _ReportTheme.text,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _DomainItem {
  const _DomainItem(this.label, this.count, this.color, this.value);

  final String label;
  final int count;
  final Color color;
  final String value;
}

class _ReportState extends StatelessWidget {
  const _ReportState({
    required this.message,
    this.actionLabel = '',
    this.onAction,
  });

  final String message;
  final String actionLabel;
  final VoidCallback? onAction;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Text(
            message,
            textAlign: TextAlign.center,
            style: const TextStyle(
              color: _ReportTheme.muted,
              fontSize: 14,
              fontWeight: FontWeight.w800,
            ),
          ),
          if (actionLabel.isNotEmpty && onAction != null) ...<Widget>[
            const SizedBox(height: 12),
            GestureDetector(
              onTap: onAction,
              child: Container(
                height: 32,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: _ReportTheme.orange,
                  borderRadius: BorderRadius.circular(11),
                ),
                child: Text(
                  actionLabel,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class _SkeletonBox extends StatelessWidget {
  const _SkeletonBox({
    required this.width,
    required this.height,
    this.radius = 8,
  });

  final double width;
  final double height;
  final double radius;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: const Color(0xFFF3E5DA),
        borderRadius: BorderRadius.circular(radius),
      ),
    );
  }
}

class _SkeletonTextCell extends StatelessWidget {
  const _SkeletonTextCell({required this.width});

  final double width;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double available =
            constraints.maxWidth.isFinite ? constraints.maxWidth : width;
        double actualWidth = width;
        if (available > 24 && actualWidth > available - 10) {
          actualWidth = available - 10;
        } else if (available <= 24) {
          actualWidth = available;
        }
        return Align(
          alignment: Alignment.centerLeft,
          child: _SkeletonBox(width: actualWidth, height: 14),
        );
      },
    );
  }
}

DateTime _dateOnly(DateTime value) =>
    DateTime(value.year, value.month, value.day);

String _dateText(DateTime value) {
  final String month = value.month.toString().padLeft(2, '0');
  final String day = value.day.toString().padLeft(2, '0');
  return '${value.year}-$month-$day';
}

DateTime? _parseDateTime(String raw) {
  final String text = raw.trim();
  if (text.isEmpty) {
    return null;
  }
  return DateTime.tryParse(text) ??
      DateTime.tryParse(text.replaceFirst(' ', 'T'));
}

String _dateOnlyText(String raw) {
  final DateTime? parsed = _parseDateTime(raw);
  if (parsed == null) {
    return raw.trim().isEmpty ? '-' : raw.trim();
  }
  return _dateText(parsed.toLocal());
}

String _ageText(Pep3RecordSummary record) {
  if (record.ageYears <= 0 && record.ageMonths <= 0 && record.ageDays <= 0) {
    return '-';
  }
  return _formatAgeParts(record.ageYears, record.ageMonths, record.ageDays);
}

String _formatAgeParts(
  int years,
  int months,
  int days, {
  bool showZeroDayWhenEmpty = false,
}) {
  final List<String> parts = <String>[];
  if (years > 0) {
    parts.add('$years岁');
  }
  if (months > 0) {
    parts.add('$months月');
  }
  if (days > 0) {
    parts.add('$days天');
  }
  if (parts.isEmpty && showZeroDayWhenEmpty) {
    return '0天';
  }
  return parts.join();
}

String _realAgeText(Pep3RecordSummary record) {
  final DateTime? birth = DateTime.tryParse(record.birthDate);
  if (birth == null) {
    return '-';
  }
  final DateTime start = _dateOnly(birth.toLocal());
  final DateTime end = _dateOnly(DateTime.now());
  if (start.isAfter(end)) {
    return '-';
  }

  int years = end.year - start.year;
  DateTime yearAnchor =
      _clampedDate(start.year + years, start.month, start.day);
  if (yearAnchor.isAfter(end)) {
    years -= 1;
    yearAnchor = _clampedDate(start.year + years, start.month, start.day);
  }

  int months = (end.year - yearAnchor.year) * 12 + end.month - yearAnchor.month;
  DateTime monthAnchor = _addMonthsClamped(yearAnchor, months);
  if (monthAnchor.isAfter(end)) {
    months -= 1;
    monthAnchor = _addMonthsClamped(yearAnchor, months);
  }

  final int days = end.difference(monthAnchor).inDays;
  return _formatAgeParts(years, months, days, showZeroDayWhenEmpty: true);
}

DateTime _addMonthsClamped(DateTime value, int months) {
  final int totalMonths = value.year * 12 + value.month - 1 + months;
  final int year = totalMonths ~/ 12;
  final int month = totalMonths % 12 + 1;
  return _clampedDate(year, month, value.day);
}

DateTime _clampedDate(int year, int month, int day) {
  final int lastDay = DateTime(year, month + 1, 0).day;
  return DateTime(year, month, math.min(day, lastDay));
}

String _studentName(Pep3RecordSummary record) {
  final String name = record.studentName.trim();
  return name.isEmpty ? '未命名儿童' : name;
}

String _studentInitial(Pep3RecordSummary record) {
  final String name = _studentName(record);
  return name.characters.first;
}

String _studentMeta(Pep3RecordSummary record) {
  final List<String> parts = <String>[];
  final String gender = record.studentGender.trim();
  if (gender.isNotEmpty) {
    parts.add(gender);
  }
  final String age = _realAgeText(record);
  if (age != '-') {
    parts.add(age);
  }
  final String phone = record.studentPhone.trim();
  if (phone.isNotEmpty) {
    parts.add(phone);
  }
  return parts.isEmpty ? '-' : parts.join(' · ');
}

String _assessmentCodeText(String raw) {
  final String code = raw.trim();
  if (code.toUpperCase() == 'PEP3') {
    return 'PEP-3';
  }
  return code.isEmpty ? '-' : code;
}

String _sequenceText(int value) => value <= 0 ? '-' : '第$value次';

String _reportTimeRaw(Pep3RecordSummary record) {
  final String createdTime = record.createdTime.trim();
  return createdTime.isNotEmpty ? createdTime : record.updatedTime;
}

Color _domainColor(int index) {
  const List<Color> colors = <Color>[
    _ReportTheme.blue,
    _ReportTheme.rose,
    Color(0xFF63A999),
    _ReportTheme.violet,
    _ReportTheme.amber,
    _ReportTheme.green,
  ];
  return colors[index % colors.length];
}

Color _avatarColor(Pep3RecordSummary record) {
  const List<Color> colors = <Color>[
    _ReportTheme.blue,
    _ReportTheme.orange,
    _ReportTheme.green,
    _ReportTheme.violet,
    _ReportTheme.rose,
  ];
  return colors[record.id.abs() % colors.length];
}

Color _codeColor(Pep3RecordSummary record) {
  switch (record.assessmentCode.trim().toUpperCase()) {
    case 'PEP3':
      return _ReportTheme.blue;
    default:
      return _domainColor(record.assessmentCode.hashCode.abs());
  }
}

Color _attemptColor(Pep3RecordSummary record) {
  const List<Color> colors = <Color>[
    _ReportTheme.orangeDeep,
    _ReportTheme.green,
    _ReportTheme.amber,
    _ReportTheme.rose,
    _ReportTheme.blue,
  ];
  final int sequence =
      record.assessmentSequence <= 0 ? 1 : record.assessmentSequence;
  return colors[(sequence - 1) % colors.length];
}
