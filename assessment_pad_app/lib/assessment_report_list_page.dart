import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_scale_client.dart';
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
    final DateTimeRange? picked = await showDialog<DateTimeRange>(
      context: context,
      barrierColor: const Color(0x33000000),
      builder: (BuildContext context) {
        return _ReportDateRangeDialog(
          initialRange: _range,
          today: today,
          minDate: DateTime(today.year - 5),
          maxDate: DateTime(today.year + 1, 12, 31),
        );
      },
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

class _ReportDateRangeDialog extends StatefulWidget {
  const _ReportDateRangeDialog({
    required this.initialRange,
    required this.today,
    required this.minDate,
    required this.maxDate,
  });

  final DateTimeRange initialRange;
  final DateTime today;
  final DateTime minDate;
  final DateTime maxDate;

  @override
  State<_ReportDateRangeDialog> createState() => _ReportDateRangeDialogState();
}

class _ReportDateRangeDialogState extends State<_ReportDateRangeDialog> {
  late DateTime _visibleMonth;
  late DateTime _start;
  DateTime? _end;
  String _selectedPresetLabel = '';

  @override
  void initState() {
    super.initState();
    _start = _dateOnly(widget.initialRange.start);
    _end = _dateOnly(widget.initialRange.end);
    _visibleMonth = _boundedVisibleMonth(_monthOnly(_start));
    _selectedPresetLabel = _presetLabelForRange(_start, _end);
  }

  DateTime _boundedVisibleMonth(DateTime month) {
    final DateTime minMonth = _monthOnly(widget.minDate);
    final DateTime maxFirstMonth = _addMonths(_monthOnly(widget.maxDate), -1);
    if (month.isBefore(minMonth)) {
      return minMonth;
    }
    if (month.isAfter(maxFirstMonth)) {
      return maxFirstMonth;
    }
    return month;
  }

  void _shiftMonth(int offset) {
    setState(() {
      _visibleMonth = _boundedVisibleMonth(_addMonths(_visibleMonth, offset));
    });
  }

  void _selectDay(DateTime day) {
    final DateTime value = _dateOnly(day);
    if (_isBeforeDay(value, widget.minDate) ||
        _isAfterDay(value, widget.maxDate)) {
      return;
    }
    setState(() {
      if (_end != null) {
        _start = value;
        _end = null;
        _selectedPresetLabel = '';
        return;
      }
      if (value.isBefore(_start)) {
        _end = _start;
        _start = value;
      } else {
        _end = value;
      }
      _selectedPresetLabel = _presetLabelForRange(_start, _end);
    });
  }

  void _applyPreset(_RangePreset preset) {
    final DateTime clippedStart = _clampDay(preset.start);
    final DateTime clippedEnd = _clampDay(preset.end);
    setState(() {
      if (clippedStart.isAfter(clippedEnd)) {
        _start = clippedEnd;
        _end = clippedStart;
      } else {
        _start = clippedStart;
        _end = clippedEnd;
      }
      _selectedPresetLabel = preset.label;
      _visibleMonth = _boundedVisibleMonth(_monthOnly(_start));
    });
  }

  String _presetLabelForRange(DateTime start, DateTime? end) {
    if (end == null) {
      return '';
    }
    for (final _RangePreset preset in _rangePresets(widget.today)) {
      if (_sameDay(start, preset.start) && _sameDay(end, preset.end)) {
        return preset.label;
      }
    }
    return '';
  }

  DateTime _clampDay(DateTime value) {
    final DateTime day = _dateOnly(value);
    if (_isBeforeDay(day, widget.minDate)) {
      return _dateOnly(widget.minDate);
    }
    if (_isAfterDay(day, widget.maxDate)) {
      return _dateOnly(widget.maxDate);
    }
    return day;
  }

  @override
  Widget build(BuildContext context) {
    final DateTime? end = _end;
    final bool canSubmit = end != null;
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 56, vertical: 42),
      child: Container(
        width: 760,
        padding: const EdgeInsets.fromLTRB(20, 18, 20, 18),
        decoration: BoxDecoration(
          color: _ReportTheme.surface,
          borderRadius: BorderRadius.circular(22),
          border: Border.all(color: _ReportTheme.line),
          boxShadow: const <BoxShadow>[
            BoxShadow(
              color: Color(0x1FC26B3E),
              blurRadius: 34,
              offset: Offset(0, 18),
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                const Text(
                  '选择日期范围',
                  style: TextStyle(
                    color: _ReportTheme.ink,
                    fontSize: 20,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const Spacer(),
                _RangePreview(label: '开始', value: _dateText(_start)),
                const SizedBox(width: 8),
                _RangePreview(
                  label: '结束',
                  value: end == null ? '请选择' : _dateText(end),
                  muted: end == null,
                ),
                const SizedBox(width: 10),
                _IconCircleButton(
                  icon: Icons.close_rounded,
                  onTap: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 18),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                SizedBox(
                  width: 118,
                  child: _RangePresetRail(
                    today: widget.today,
                    selectedLabel: _selectedPresetLabel,
                    onSelected: _applyPreset,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    children: <Widget>[
                      Row(
                        children: <Widget>[
                          _IconCircleButton(
                            icon: Icons.chevron_left_rounded,
                            enabled: _canShiftPrevious,
                            onTap: () => _shiftMonth(-1),
                          ),
                          const Spacer(),
                          _MonthTitle(month: _visibleMonth),
                          const SizedBox(width: 108),
                          _MonthTitle(month: _addMonths(_visibleMonth, 1)),
                          const Spacer(),
                          _IconCircleButton(
                            icon: Icons.chevron_right_rounded,
                            enabled: _canShiftNext,
                            onTap: () => _shiftMonth(1),
                          ),
                        ],
                      ),
                      const SizedBox(height: 12),
                      Row(
                        children: <Widget>[
                          Expanded(
                            child: _MonthCalendar(
                              month: _visibleMonth,
                              minDate: widget.minDate,
                              maxDate: widget.maxDate,
                              today: widget.today,
                              start: _start,
                              end: end,
                              onSelect: _selectDay,
                            ),
                          ),
                          const SizedBox(width: 18),
                          Expanded(
                            child: _MonthCalendar(
                              month: _addMonths(_visibleMonth, 1),
                              minDate: widget.minDate,
                              maxDate: widget.maxDate,
                              today: widget.today,
                              start: _start,
                              end: end,
                              onSelect: _selectDay,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 18),
            Row(
              children: <Widget>[
                const Text(
                  '点击日期重新选择开始日期，再点击一次选择结束日期',
                  style: TextStyle(
                    color: _ReportTheme.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const Spacer(),
                _DialogActionButton(
                  label: '取消',
                  onTap: () => Navigator.of(context).pop(),
                ),
                const SizedBox(width: 10),
                _DialogActionButton(
                  label: '确定',
                  filled: true,
                  enabled: canSubmit,
                  onTap: canSubmit
                      ? () {
                          Navigator.of(context).pop(
                            DateTimeRange(start: _start, end: end),
                          );
                        }
                      : null,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  bool get _canShiftPrevious {
    return !_addMonths(_visibleMonth, -1).isBefore(_monthOnly(widget.minDate));
  }

  bool get _canShiftNext {
    final DateTime nextSecondMonth = _addMonths(_visibleMonth, 2);
    return !nextSecondMonth.isAfter(_monthOnly(widget.maxDate));
  }
}

class _RangePreview extends StatelessWidget {
  const _RangePreview({
    required this.label,
    required this.value,
    this.muted = false,
  });

  final String label;
  final String value;
  final bool muted;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 36,
      padding: const EdgeInsets.symmetric(horizontal: 11),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF8F2),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Text(
            label,
            style: const TextStyle(
              color: _ReportTheme.muted,
              fontSize: 11,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 7),
          Text(
            value,
            style: TextStyle(
              color: muted ? _ReportTheme.muted : _ReportTheme.ink,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _RangePresetRail extends StatelessWidget {
  const _RangePresetRail({
    required this.today,
    required this.selectedLabel,
    required this.onSelected,
  });

  final DateTime today;
  final String selectedLabel;
  final ValueChanged<_RangePreset> onSelected;

  @override
  Widget build(BuildContext context) {
    final List<_RangePreset> presets = _rangePresets(today);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        const Text(
          '快捷选择',
          style: TextStyle(
            color: _ReportTheme.muted,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 10),
        for (final _RangePreset preset in presets) ...<Widget>[
          _PresetButton(
            label: preset.label,
            selected: preset.label == selectedLabel,
            onTap: () => onSelected(preset),
          ),
          const SizedBox(height: 8),
        ],
      ],
    );
  }
}

class _RangePreset {
  const _RangePreset(this.label, this.start, this.end);

  final String label;
  final DateTime start;
  final DateTime end;
}

List<_RangePreset> _rangePresets(DateTime today) {
  final DateTime day = _dateOnly(today);
  final DateTime monthStart = DateTime(day.year, day.month);
  final DateTime lastMonthStart = DateTime(day.year, day.month - 1);
  return <_RangePreset>[
    _RangePreset('近7天', day.subtract(const Duration(days: 6)), day),
    _RangePreset('近30天', day.subtract(const Duration(days: 29)), day),
    _RangePreset('本月', monthStart, day),
    _RangePreset(
        '上月', lastMonthStart, monthStart.subtract(const Duration(days: 1))),
  ];
}

class _PresetButton extends StatelessWidget {
  const _PresetButton({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          height: 36,
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFF0E7) : const Color(0xFFFFF8F2),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: selected ? _ReportTheme.orange : _ReportTheme.lineSoft,
            ),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              if (selected) ...<Widget>[
                const Icon(
                  Icons.check_rounded,
                  size: 16,
                  color: _ReportTheme.orangeDeep,
                ),
                const SizedBox(width: 4),
              ],
              Flexible(
                child: Text(
                  label,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color:
                        selected ? _ReportTheme.orangeDeep : _ReportTheme.text,
                    fontSize: 13,
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

class _MonthTitle extends StatelessWidget {
  const _MonthTitle({required this.month});

  final DateTime month;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 96,
      child: Text(
        '${month.year}年${month.month}月',
        textAlign: TextAlign.center,
        style: const TextStyle(
          color: _ReportTheme.ink,
          fontSize: 15,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _MonthCalendar extends StatelessWidget {
  const _MonthCalendar({
    required this.month,
    required this.minDate,
    required this.maxDate,
    required this.today,
    required this.start,
    required this.end,
    required this.onSelect,
  });

  final DateTime month;
  final DateTime minDate;
  final DateTime maxDate;
  final DateTime today;
  final DateTime start;
  final DateTime? end;
  final ValueChanged<DateTime> onSelect;

  @override
  Widget build(BuildContext context) {
    final DateTime firstDay = DateTime(month.year, month.month);
    final int leadingDays = firstDay.weekday - DateTime.monday;
    final DateTime gridStart = firstDay.subtract(Duration(days: leadingDays));
    return Column(
      children: <Widget>[
        const Row(
          children: <Widget>[
            _WeekdayLabel('一'),
            _WeekdayLabel('二'),
            _WeekdayLabel('三'),
            _WeekdayLabel('四'),
            _WeekdayLabel('五'),
            _WeekdayLabel('六'),
            _WeekdayLabel('日'),
          ],
        ),
        const SizedBox(height: 6),
        for (int row = 0; row < 6; row++)
          Row(
            children: <Widget>[
              for (int column = 0; column < 7; column++)
                Expanded(
                  child: _DateCell(
                    day: gridStart.add(Duration(days: row * 7 + column)),
                    visibleMonth: month,
                    minDate: minDate,
                    maxDate: maxDate,
                    today: today,
                    start: start,
                    end: end,
                    onSelect: onSelect,
                  ),
                ),
            ],
          ),
      ],
    );
  }
}

class _WeekdayLabel extends StatelessWidget {
  const _WeekdayLabel(this.label);

  final String label;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Center(
        child: Text(
          label,
          style: const TextStyle(
            color: _ReportTheme.muted,
            fontSize: 11,
            fontWeight: FontWeight.w900,
          ),
        ),
      ),
    );
  }
}

class _DateCell extends StatelessWidget {
  const _DateCell({
    required this.day,
    required this.visibleMonth,
    required this.minDate,
    required this.maxDate,
    required this.today,
    required this.start,
    required this.end,
    required this.onSelect,
  });

  final DateTime day;
  final DateTime visibleMonth;
  final DateTime minDate;
  final DateTime maxDate;
  final DateTime today;
  final DateTime start;
  final DateTime? end;
  final ValueChanged<DateTime> onSelect;

  @override
  Widget build(BuildContext context) {
    final DateTime current = _dateOnly(day);
    final bool inCurrentMonth = current.year == visibleMonth.year &&
        current.month == visibleMonth.month;
    final bool enabled = inCurrentMonth &&
        !_isBeforeDay(current, minDate) &&
        !_isAfterDay(current, maxDate);
    final bool isStart = _sameDay(current, start);
    final bool isEnd = end != null && _sameDay(current, end!);
    final bool inRange = end != null &&
        !current.isBefore(start) &&
        !current.isAfter(end!) &&
        inCurrentMonth;
    final bool isToday = _sameDay(current, today);
    final int lastDayOfMonth =
        DateTime(visibleMonth.year, visibleMonth.month + 1, 0).day;
    final bool startsRangeBand =
        isStart || current.weekday == DateTime.monday || current.day == 1;
    final bool endsRangeBand = isEnd ||
        current.weekday == DateTime.sunday ||
        current.day == lastDayOfMonth;
    final bool isSingleDay = isStart && isEnd;
    final bool shouldShowBand = inRange && !isSingleDay;
    final Color textColor = isStart || isEnd
        ? Colors.white
        : enabled
            ? _ReportTheme.text
            : _ReportTheme.line;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: enabled ? () => onSelect(current) : null,
          borderRadius: BorderRadius.circular(10),
          child: SizedBox(
            height: 34,
            child: LayoutBuilder(
              builder: (BuildContext context, BoxConstraints constraints) {
                final double halfWidth = constraints.maxWidth / 2;
                return Stack(
                  alignment: Alignment.center,
                  children: <Widget>[
                    if (shouldShowBand)
                      Positioned(
                        top: 5,
                        bottom: 5,
                        left: isStart ? halfWidth : 0,
                        right: isEnd ? halfWidth : 0,
                        child: DecoratedBox(
                          decoration: BoxDecoration(
                            color: const Color(0xFFFFF0E7),
                            borderRadius: BorderRadius.horizontal(
                              left: Radius.circular(
                                startsRangeBand && !isStart ? 10 : 0,
                              ),
                              right: Radius.circular(
                                endsRangeBand && !isEnd ? 10 : 0,
                              ),
                            ),
                          ),
                        ),
                      ),
                    if (isStart || isEnd)
                      Container(
                        width: 32,
                        height: 28,
                        decoration: BoxDecoration(
                          color: _ReportTheme.orange,
                          borderRadius: BorderRadius.circular(10),
                        ),
                      )
                    else if (isToday)
                      Container(
                        width: 32,
                        height: 28,
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(10),
                          border: Border.all(
                            color: _ReportTheme.orange.withOpacity(.42),
                          ),
                        ),
                      ),
                    Text(
                      '${current.day}',
                      style: TextStyle(
                        color: textColor,
                        fontSize: 12,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ],
                );
              },
            ),
          ),
        ),
      ),
    );
  }
}

class _IconCircleButton extends StatelessWidget {
  const _IconCircleButton({
    required this.icon,
    required this.onTap,
    this.enabled = true,
  });

  final IconData icon;
  final VoidCallback onTap;
  final bool enabled;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      shape: const CircleBorder(),
      child: InkWell(
        onTap: enabled ? onTap : null,
        customBorder: const CircleBorder(),
        child: Container(
          width: 34,
          height: 34,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: enabled ? const Color(0xFFFFF8F2) : const Color(0xFFF8EEE6),
            shape: BoxShape.circle,
            border: Border.all(color: _ReportTheme.lineSoft),
          ),
          child: Icon(
            icon,
            size: 20,
            color: enabled ? _ReportTheme.text : _ReportTheme.muted,
          ),
        ),
      ),
    );
  }
}

class _DialogActionButton extends StatelessWidget {
  const _DialogActionButton({
    required this.label,
    required this.onTap,
    this.filled = false,
    this.enabled = true,
  });

  final String label;
  final VoidCallback? onTap;
  final bool filled;
  final bool enabled;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: enabled ? onTap : null,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          height: 38,
          constraints: const BoxConstraints(minWidth: 76),
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(horizontal: 16),
          decoration: BoxDecoration(
            color:
                filled && enabled ? _ReportTheme.orange : _ReportTheme.surface,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color:
                  filled && enabled ? _ReportTheme.orange : _ReportTheme.line,
            ),
          ),
          child: Text(
            label,
            style: TextStyle(
              color: filled && enabled
                  ? Colors.white
                  : enabled
                      ? _ReportTheme.text
                      : _ReportTheme.muted,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
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

DateTime _monthOnly(DateTime value) => DateTime(value.year, value.month);

DateTime _addMonths(DateTime value, int months) {
  return DateTime(value.year, value.month + months);
}

bool _sameDay(DateTime left, DateTime right) {
  return left.year == right.year &&
      left.month == right.month &&
      left.day == right.day;
}

bool _isBeforeDay(DateTime left, DateTime right) {
  return _dateOnly(left).isBefore(_dateOnly(right));
}

bool _isAfterDay(DateTime left, DateTime right) {
  return _dateOnly(left).isAfter(_dateOnly(right));
}

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
