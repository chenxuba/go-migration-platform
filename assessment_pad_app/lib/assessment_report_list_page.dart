import 'dart:async';
import 'dart:collection';
import 'dart:convert';
import 'dart:math' as math;
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:printing/printing.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_scale_client.dart';
import 'autismdev_assessment_client.dart';
import 'erxin_assessment_client.dart';
import 'pad_date_range_picker.dart';
import 'pad_responsive.dart';
import 'pad_top_message.dart';
import 'pep3_assessment_client.dart';
import 'route_bootstrap.dart';
import 'timetable_client.dart';

part 'autismdev_report_preview.dart';

class AssessmentReportListScreen extends StatefulWidget {
  const AssessmentReportListScreen({
    required this.onBack,
    this.scaleClient = const ApiAssessmentScaleClient(),
    this.recordClient = const ApiPep3AssessmentClient(),
    this.erxinRecordClient = const ApiPep3AssessmentClient(
      recordsPagePath: defaultErxinRecordsPagePath,
      recordCategoryStatsPath: defaultErxinRecordCategoryStatsPath,
    ),
    this.autismDevRecordClient = const ApiPep3AssessmentClient(
      recordsPagePath: defaultAutismDevRecordsPagePath,
      recordCategoryStatsPath: defaultAutismDevRecordCategoryStatsPath,
      recordDetailPath: defaultAutismDevRecordDetailPath,
      recordConfigUpdatePath: defaultAutismDevRecordConfigUpdatePath,
    ),
    this.erxinClient = const ApiErxinAssessmentClient(),
    this.staffClient = const ApiTimetableClient(),
    super.key,
  });

  final VoidCallback onBack;
  final AssessmentScaleClient scaleClient;
  final Pep3AssessmentClient recordClient;
  final Pep3AssessmentClient erxinRecordClient;
  final Pep3AssessmentClient autismDevRecordClient;
  final ErxinAssessmentClient erxinClient;
  final TimetableClient staffClient;

  @override
  State<AssessmentReportListScreen> createState() =>
      _AssessmentReportListScreenState();
}

class _AssessmentReportListScreenState
    extends State<AssessmentReportListScreen> {
  static const String _authTokenStorageKey = 'auth_token';

  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();
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
  bool _bootstrapLoading = true;
  String _errorMessage = '';
  int _searchResetSeed = 0;

  @override
  void initState() {
    super.initState();
    final DateTime today = _dateOnly(DateTime.now());
    _range = DateTimeRange(
      start: today.subtract(const Duration(days: 29)),
      end: today,
    );
    runAfterRouteEntrance(context, () => _loadData(bootstrap: true));
  }

  Future<void> _loadData({
    String? selectedCategory,
    bool reloadCategories = false,
    bool bootstrap = false,
  }) async {
    final bool shouldLoadCategories = reloadCategories || _categories.isEmpty;
    final bool shouldShowCategorySkeleton =
        shouldLoadCategories && _categories.isEmpty;
    if (!bootstrap) {
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
    } else if (selectedCategory != null) {
      _selectedCategory = selectedCategory;
    }
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final String dateBegin = _dateText(_range.start);
      final String dateEnd = _dateText(_range.end);
      if (bootstrap) {
        final Future<List<String>> categoriesFuture = shouldLoadCategories
            ? widget.scaleClient.fetchCategories(token)
            : Future<List<String>>.value(_categories);
        final Future<Pep3RecordCategoryStats> statsFuture =
            _fetchRecordCategoryStats(
          token,
          searchKey: _searchKey,
          assessmentDateBegin: dateBegin,
          assessmentDateEnd: dateEnd,
        );
        final Pep3RecordPage page = await _fetchRecordsPage(
          token,
          pageIndex: 1,
          pageSize: 50,
          scaleCategory: _selectedCategory,
          searchKey: _searchKey,
          assessmentDateBegin: dateBegin,
          assessmentDateEnd: dateEnd,
        );
        if (!mounted) {
          return;
        }
        setState(() {
          _page = page;
          _bootstrapLoading = false;
          _listLoading = false;
        });
        unawaited(() async {
          try {
            final List<dynamic> results = await Future.wait<dynamic>(
              <Future<dynamic>>[categoriesFuture, statsFuture],
            );
            if (!mounted) {
              return;
            }
            final List<String> categories =
                List<String>.from(results[0] as List);
            final Pep3RecordCategoryStats stats =
                results[1] as Pep3RecordCategoryStats;
            final Map<String, int> counts = <String, int>{
              for (final String category in categories)
                category: stats.categoryCounts[category] ?? 0,
            };
            setState(() {
              if (shouldLoadCategories) {
                _categories = categories;
              }
              _categoryCounts = counts;
              _rangeTotal = stats.total;
              _categoryLoading = false;
            });
          } on Object catch (error) {
            if (!mounted) {
              return;
            }
            setState(() {
              if (shouldLoadCategories) {
                _categoryLoading = false;
              }
              if (_errorMessage.isEmpty) {
                _errorMessage = '$error';
              }
            });
          }
        }());
        return;
      }
      final Future<List<String>> categoriesFuture = shouldLoadCategories
          ? widget.scaleClient.fetchCategories(token)
          : Future<List<String>>.value(_categories);
      final List<dynamic> results = await Future.wait<dynamic>(
        <Future<dynamic>>[
          categoriesFuture,
          _fetchRecordsPage(
            token,
            pageIndex: 1,
            pageSize: 50,
            scaleCategory: _selectedCategory,
            searchKey: _searchKey,
            assessmentDateBegin: dateBegin,
            assessmentDateEnd: dateEnd,
          ),
          _fetchRecordCategoryStats(
            token,
            searchKey: _searchKey,
            assessmentDateBegin: dateBegin,
            assessmentDateEnd: dateEnd,
          ),
        ],
      );
      final List<String> categories = List<String>.from(results[0] as List);
      final Pep3RecordPage page = results[1] as Pep3RecordPage;
      final Pep3RecordCategoryStats stats =
          results[2] as Pep3RecordCategoryStats;
      final Map<String, int> counts = <String, int>{
        for (final String category in categories)
          category: stats.categoryCounts[category] ?? 0,
      };
      if (!mounted) {
        return;
      }
      setState(() {
        if (shouldLoadCategories) {
          _categories = categories;
        }
        _categoryCounts = counts;
        _rangeTotal = stats.total;
        _categoryLoading = false;
        _page = page;
        _bootstrapLoading = false;
        _listLoading = false;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _bootstrapLoading = false;
        _listLoading = false;
        if (shouldLoadCategories) {
          _categoryLoading = false;
        }
        _errorMessage = '$error';
      });
    }
  }

  Future<Pep3RecordPage> _fetchRecordsPage(
    String token, {
    required int pageIndex,
    required int pageSize,
    required String scaleCategory,
    required String searchKey,
    required String assessmentDateBegin,
    required String assessmentDateEnd,
  }) async {
    final List<Pep3RecordPage> pages = await Future.wait<Pep3RecordPage>(
      <Future<Pep3RecordPage>>[
        widget.recordClient.fetchRecordsPage(
          token,
          pageIndex: pageIndex,
          pageSize: pageSize,
          assessmentCode: '',
          scaleCategory: scaleCategory,
          searchKey: searchKey,
          assessmentDateBegin: assessmentDateBegin,
          assessmentDateEnd: assessmentDateEnd,
        ),
        widget.erxinRecordClient.fetchRecordsPage(
          token,
          pageIndex: pageIndex,
          pageSize: pageSize,
          assessmentCode: '',
          scaleCategory: scaleCategory,
          searchKey: searchKey,
          assessmentDateBegin: assessmentDateBegin,
          assessmentDateEnd: assessmentDateEnd,
        ),
        widget.autismDevRecordClient.fetchRecordsPage(
          token,
          pageIndex: pageIndex,
          pageSize: pageSize,
          assessmentCode: '',
          scaleCategory: scaleCategory,
          searchKey: searchKey,
          assessmentDateBegin: assessmentDateBegin,
          assessmentDateEnd: assessmentDateEnd,
        ),
      ],
    );
    final List<Pep3RecordSummary> items = <Pep3RecordSummary>[
      ...pages[0].items,
      ...pages[1].items,
      ...pages[2].items,
    ]..sort(_compareRecordSummaryDesc);
    final List<Pep3RecordSummary> visibleItems =
        items.take(pageSize).toList(growable: false);
    return Pep3RecordPage(
      items: visibleItems,
      total: pages.fold<int>(
          0, (int sum, Pep3RecordPage page) => sum + page.total),
      current: pageIndex,
      size: visibleItems.length,
    );
  }

  Future<Pep3RecordCategoryStats> _fetchRecordCategoryStats(
    String token, {
    required String searchKey,
    required String assessmentDateBegin,
    required String assessmentDateEnd,
  }) async {
    final List<Pep3RecordCategoryStats> statsList =
        await Future.wait<Pep3RecordCategoryStats>(
      <Future<Pep3RecordCategoryStats>>[
        widget.recordClient.fetchRecordCategoryStats(
          token,
          assessmentCode: '',
          searchKey: searchKey,
          assessmentDateBegin: assessmentDateBegin,
          assessmentDateEnd: assessmentDateEnd,
        ),
        widget.erxinRecordClient.fetchRecordCategoryStats(
          token,
          assessmentCode: '',
          searchKey: searchKey,
          assessmentDateBegin: assessmentDateBegin,
          assessmentDateEnd: assessmentDateEnd,
        ),
        widget.autismDevRecordClient.fetchRecordCategoryStats(
          token,
          assessmentCode: '',
          searchKey: searchKey,
          assessmentDateBegin: assessmentDateBegin,
          assessmentDateEnd: assessmentDateEnd,
        ),
      ],
    );
    final Map<String, int> counts = <String, int>{};
    for (final Pep3RecordCategoryStats stats in statsList) {
      stats.categoryCounts.forEach((String category, int count) {
        counts[category] = (counts[category] ?? 0) + count;
      });
    }
    return Pep3RecordCategoryStats(
      total: statsList.fold<int>(
        0,
        (int sum, Pep3RecordCategoryStats stats) => sum + stats.total,
      ),
      categoryCounts: counts,
    );
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

  void _resetFilters() {
    final DateTime today = _dateOnly(DateTime.now());
    setState(() {
      _range = DateTimeRange(
        start: today.subtract(const Duration(days: 29)),
        end: today,
      );
      _selectedCategory = '';
      _searchKey = '';
      _searchResetSeed += 1;
    });
    _loadData();
  }

  void _submitSearch(String value) {
    _searchKey = value.trim();
    _loadData();
  }

  @override
  void dispose() {
    _messageController.dispose();
    super.dispose();
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
        bootstrapLoading: _bootstrapLoading && _listLoading,
        categoryLoading: _categoryLoading,
        listLoading: _listLoading,
        errorMessage: _errorMessage,
        searchResetSeed: _searchResetSeed,
        onReset: _resetFilters,
        onRangeTap: _selectRange,
        onSearchSubmitted: _submitSearch,
        onViewReport: _openReportViewer,
        onConfigRecord: _openConfigDialog,
        onCategorySelected: (String category) => _loadData(
          selectedCategory: category,
        ),
      ),
    );
  }

  Future<void> _openReportViewer(Pep3RecordSummary record) async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_authTokenStorageKey) ?? '';
    if (!mounted) {
      return;
    }
    if (_isAutismDevRecord(record)) {
      showDialog<void>(
        context: context,
        barrierDismissible: false,
        barrierColor: Colors.black.withOpacity(.28),
        builder: (BuildContext dialogContext) {
          return PadDialogViewport(
            child: _AutismDevReportPreviewDialog(
              record: record,
              token: token,
              client: widget.autismDevRecordClient,
            ),
          );
        },
      );
      return;
    }
    showDialog<void>(
      context: context,
      barrierDismissible: false,
      barrierColor: Colors.black.withOpacity(.28),
      builder: (BuildContext dialogContext) {
        return PadDialogViewport(
          child: _isErxinRecord(record)
              ? _ErxinReportPreviewDialog(
                  record: record,
                  token: token,
                  client: widget.erxinClient,
                )
              : _ReportPreviewDialog(
                  record: record,
                  token: token,
                  client: widget.recordClient,
                ),
        );
      },
    );
  }

  Future<void> _openConfigDialog(Pep3RecordSummary record) async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_authTokenStorageKey) ?? '';
    if (!mounted) {
      return;
    }
    final Pep3RecordSummary? savedDetail = await showDialog<Pep3RecordSummary>(
      context: context,
      barrierDismissible: true,
      barrierColor: Colors.black.withOpacity(.28),
      builder: (BuildContext dialogContext) {
        return PadDialogViewport(
          child: _ReportConfigDialog(
            record: record,
            token: token,
            staffClient: widget.staffClient,
            loadConfigDetail: _loadRecordConfigDetail,
            saveConfig: _saveRecordConfig,
          ),
        );
      },
    );
    if (savedDetail == null || !mounted) {
      return;
    }
    setState(() {
      final List<Pep3RecordSummary> nextItems = _page.items
          .map(
            (Pep3RecordSummary item) =>
                item.id == savedDetail.id ? savedDetail : item,
          )
          .toList();
      _page = Pep3RecordPage(
        items: nextItems,
        total: _page.total,
        current: _page.current,
        size: _page.size,
      );
    });
    _showMessage('评估配置已保存', tone: PadMessageTone.success);
    unawaited(_loadData());
  }

  void _showMessage(
    String message, {
    PadMessageTone tone = PadMessageTone.info,
  }) {
    if (!mounted || message.trim().isEmpty) {
      return;
    }
    _messageController.show(
      context,
      message,
      tone: tone,
      topMargin: 12,
      key: 'report-list-top-message',
    );
  }

  bool _isErxinRecord(Pep3RecordSummary record) {
    return record.assessmentCode.trim().toUpperCase() == 'ERXIN2';
  }

  bool _isAutismDevRecord(Pep3RecordSummary record) {
    return record.assessmentCode.trim().toUpperCase() == 'AUTISMDEV';
  }

  Future<_RecordConfigDetail> _loadRecordConfigDetail(
    String token,
    Pep3RecordSummary record,
  ) async {
    if (_isErxinRecord(record)) {
      final ErxinRecordDetail detail =
          await widget.erxinClient.fetchRecordDetail(token, record.id);
      return _RecordConfigDetail(
        currentExaminerName: detail.examinerName,
        currentAssessmentDate: detail.assessmentDate,
        originalExaminerName: detail.input.examinerName,
        originalAssessmentDate:
            _originalAssessmentDateText(record) ?? record.assessmentDate,
      );
    }
    if (_isAutismDevRecord(record)) {
      final Pep3RecordDetail detail =
          await widget.autismDevRecordClient.fetchRecordDetail(
        token,
        record.id,
      );
      return _RecordConfigDetail(
        currentExaminerName: detail.examinerName,
        currentAssessmentDate: detail.assessmentDate,
        originalExaminerName: detail.input.examinerName,
        originalAssessmentDate:
            _originalAssessmentDateText(record) ?? record.assessmentDate,
      );
    }
    final Pep3RecordDetail detail =
        await widget.recordClient.fetchRecordDetail(token, record.id);
    return _RecordConfigDetail(
      currentExaminerName: detail.examinerName,
      currentAssessmentDate: detail.assessmentDate,
      originalExaminerName: detail.input.examinerName,
      originalAssessmentDate:
          _originalAssessmentDateText(record) ?? record.assessmentDate,
    );
  }

  Future<Pep3RecordSummary> _saveRecordConfig(
    String token,
    Pep3RecordSummary record, {
    required String examinerName,
    required String assessmentDate,
  }) async {
    if (_isErxinRecord(record)) {
      final ErxinRecordDetail detail =
          await widget.erxinClient.updateRecordConfig(
        token,
        record.id,
        examinerName: examinerName,
        assessmentDate: assessmentDate,
      );
      return Pep3RecordSummary(
        id: detail.id,
        studentId: detail.studentId,
        studentName: detail.studentName,
        studentGender: detail.studentGender,
        studentAvatar: detail.studentAvatar,
        studentPhone: detail.studentPhone,
        assessmentCode: detail.assessmentCode,
        assessmentName: detail.assessmentName,
        scaleCategory: detail.scaleCategory,
        scaleVersion: detail.scaleVersion,
        birthDate: detail.birthDate,
        assessmentDate: detail.assessmentDate,
        ageYears: detail.ageYears,
        ageMonths: detail.ageMonths,
        ageDays: detail.ageDays,
        normAgeMonths: detail.normAgeMonths,
        assessmentSequence: detail.assessmentSequence,
        examinerName: detail.examinerName,
        createdTime: detail.createdTime,
        updatedTime: detail.updatedTime,
      );
    }
    if (_isAutismDevRecord(record)) {
      return widget.autismDevRecordClient.updateRecordConfig(
        token,
        record.id,
        examinerName: examinerName,
        assessmentDate: assessmentDate,
      );
    }
    return widget.recordClient.updateRecordConfig(
      token,
      record.id,
      examinerName: examinerName,
      assessmentDate: assessmentDate,
    );
  }
}

class _RecordConfigDetail {
  const _RecordConfigDetail({
    required this.currentExaminerName,
    required this.currentAssessmentDate,
    required this.originalExaminerName,
    required this.originalAssessmentDate,
  });

  final String currentExaminerName;
  final String currentAssessmentDate;
  final String originalExaminerName;
  final String originalAssessmentDate;
}

int _compareRecordSummaryDesc(Pep3RecordSummary left, Pep3RecordSummary right) {
  final DateTime rightTime = _recordSortTime(right);
  final DateTime leftTime = _recordSortTime(left);
  final int timeCompare = rightTime.compareTo(leftTime);
  if (timeCompare != 0) {
    return timeCompare;
  }
  return right.id.compareTo(left.id);
}

DateTime _recordSortTime(Pep3RecordSummary record) {
  final DateTime? updatedTime = _tryParseRecordDateTime(record.updatedTime);
  if (updatedTime != null) {
    return updatedTime;
  }
  final DateTime? createdTime = _tryParseRecordDateTime(record.createdTime);
  if (createdTime != null) {
    return createdTime;
  }
  final DateTime? assessmentDate =
      _tryParseRecordDateTime(record.assessmentDate);
  if (assessmentDate != null) {
    return assessmentDate;
  }
  return DateTime.fromMillisecondsSinceEpoch(0);
}

DateTime? _tryParseRecordDateTime(String raw) {
  final String value = raw.trim();
  if (value.isEmpty) {
    return null;
  }
  return DateTime.tryParse(value) ??
      DateTime.tryParse(value.replaceFirst(' ', 'T'));
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
    required this.bootstrapLoading,
    required this.categoryLoading,
    required this.listLoading,
    required this.errorMessage,
    required this.searchResetSeed,
    required this.onReset,
    required this.onRangeTap,
    required this.onSearchSubmitted,
    required this.onViewReport,
    required this.onConfigRecord,
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
  final bool bootstrapLoading;
  final bool categoryLoading;
  final bool listLoading;
  final String errorMessage;
  final int searchResetSeed;
  final VoidCallback onReset;
  final VoidCallback onRangeTap;
  final ValueChanged<String> onSearchSubmitted;
  final ValueChanged<Pep3RecordSummary> onViewReport;
  final ValueChanged<Pep3RecordSummary> onConfigRecord;
  final ValueChanged<String> onCategorySelected;

  @override
  Widget build(BuildContext context) {
    return Builder(
      builder: (BuildContext context) {
        return Listener(
          behavior: HitTestBehavior.translucent,
          onPointerDown: (_) => FocusManager.instance.primaryFocus?.unfocus(),
          child: ColoredBox(
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
                        onReset: onReset,
                        searchResetSeed: searchResetSeed,
                        onSearchSubmitted: onSearchSubmitted,
                      ),
                      const SizedBox(height: 30),
                      Expanded(
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: <Widget>[
                            SizedBox(
                              width: sideWidth,
                              child: bootstrapLoading
                                  ? const _ReportBootstrapSidebar()
                                  : _DomainPanel(
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
                              child: bootstrapLoading
                                  ? const _ReportBootstrapPanel()
                                  : _ReportListPanel(
                                      records: records,
                                      total: total,
                                      loading: listLoading,
                                      errorMessage: errorMessage,
                                      onRetry: onReset,
                                      onViewReport: onViewReport,
                                      onConfigRecord: onConfigRecord,
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
    required this.onReset,
    required this.searchResetSeed,
    required this.onSearchSubmitted,
  });

  final VoidCallback onBack;
  final DateTimeRange range;
  final VoidCallback onRangeTap;
  final VoidCallback onReset;
  final int searchResetSeed;
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
          _SearchBox(
              resetSeed: searchResetSeed, onSubmitted: onSearchSubmitted),
          const SizedBox(width: 10),
          _ToolbarButton(
            label: '${_dateText(range.start)} - ${_dateText(range.end)}',
            icon: Icons.calendar_month_rounded,
            onTap: onRangeTap,
            triggerOnTapDown: true,
          ),
          const SizedBox(width: 10),
          _ToolbarButton(label: '重置', filled: true, onTap: onReset),
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

class _SearchBox extends StatefulWidget {
  const _SearchBox({
    required this.resetSeed,
    required this.onSubmitted,
  });

  final int resetSeed;
  final ValueChanged<String> onSubmitted;

  @override
  State<_SearchBox> createState() => _SearchBoxState();
}

class _SearchBoxState extends State<_SearchBox> {
  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController();
  }

  @override
  void didUpdateWidget(_SearchBox oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.resetSeed != widget.resetSeed) {
      _controller.clear();
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

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
        controller: _controller,
        onSubmitted: widget.onSubmitted,
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
          hintText: '搜索儿童姓名',
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
    this.triggerOnTapDown = false,
    super.key,
  });

  final String label;
  final bool filled;
  final IconData? icon;
  final VoidCallback? onTap;
  final bool triggerOnTapDown;

  @override
  Widget build(BuildContext context) {
    final Widget button = IntrinsicWidth(
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: triggerOnTapDown ? null : onTap,
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
      ),
    );
    if (!triggerOnTapDown || onTap == null) {
      return button;
    }
    return Listener(
      behavior: HitTestBehavior.opaque,
      onPointerDown: (_) => onTap!(),
      child: button,
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
    return ListView.separated(
      padding: EdgeInsets.zero,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: 6,
      separatorBuilder: (BuildContext context, int index) =>
          const SizedBox(height: 7),
      itemBuilder: (BuildContext context, int index) =>
          const _DomainSkeletonRow(),
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
    required this.onViewReport,
    required this.onConfigRecord,
  });

  final List<Pep3RecordSummary> records;
  final int total;
  final bool loading;
  final String errorMessage;
  final VoidCallback onRetry;
  final ValueChanged<Pep3RecordSummary> onViewReport;
  final ValueChanged<Pep3RecordSummary> onConfigRecord;

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
              const Expanded(
                child: _ReportSkeletonList(),
              )
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
              Expanded(
                child: ListView.builder(
                  padding: EdgeInsets.zero,
                  physics: const BouncingScrollPhysics(),
                  itemCount: records.length,
                  itemExtent: 73,
                  itemBuilder: (BuildContext context, int index) {
                    final Pep3RecordSummary record = records[index];
                    return _ReportRow(
                      record: record,
                      onViewReport: onViewReport,
                      onConfigRecord: onConfigRecord,
                    );
                  },
                ),
              ),
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
  const _ReportRow({
    required this.record,
    required this.onViewReport,
    required this.onConfigRecord,
  });

  final Pep3RecordSummary record;
  final ValueChanged<Pep3RecordSummary> onViewReport;
  final ValueChanged<Pep3RecordSummary> onConfigRecord;

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
            _ColumnCell(
              flex: 168,
              trailingGap: 0,
              child: Align(
                alignment: Alignment.centerRight,
                child: _RowActions(
                  record: record,
                  onViewReport: onViewReport,
                  onConfigRecord: onConfigRecord,
                ),
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

class _ReportSkeletonList extends StatelessWidget {
  const _ReportSkeletonList();

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      padding: EdgeInsets.zero,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: 4,
      itemExtent: 73,
      itemBuilder: (BuildContext context, int index) =>
          const _ReportSkeletonRow(),
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

class _ReportBootstrapSidebar extends StatelessWidget {
  const _ReportBootstrapSidebar();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 16, 12, 16),
      decoration: BoxDecoration(
        color: _ReportTheme.surface.withOpacity(.94),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _ReportTheme.line),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x12C26B3E),
            blurRadius: 20,
            offset: Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          const Text(
            '测评分类',
            style: TextStyle(
              color: _ReportTheme.ink,
              fontSize: 17,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 14),
          Expanded(
            child: Column(
              children: <Widget>[
                _DomainRow(
                  item: const _DomainItem('全部分类', 0, _ReportTheme.orange, ''),
                  selected: true,
                  onTap: _noopReportAction,
                ),
                const SizedBox(height: 7),
                const _DomainBootstrapPlaceholderRow(
                  color: _ReportTheme.blue,
                  width: 74,
                ),
                const SizedBox(height: 7),
                const _DomainBootstrapPlaceholderRow(
                  color: _ReportTheme.green,
                  width: 82,
                ),
                const SizedBox(height: 7),
                const _DomainBootstrapPlaceholderRow(
                  color: _ReportTheme.amber,
                  width: 68,
                ),
                const SizedBox(height: 7),
                const _DomainBootstrapPlaceholderRow(
                  color: _ReportTheme.rose,
                  width: 78,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ReportBootstrapPanel extends StatelessWidget {
  const _ReportBootstrapPanel();

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: _ReportTheme.surface.withOpacity(.96),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _ReportTheme.line),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x12C26B3E),
            blurRadius: 20,
            offset: Offset(0, 10),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
        child: Column(
          children: <Widget>[
            Container(
              height: 63,
              padding: const EdgeInsets.symmetric(horizontal: 16),
              decoration: const BoxDecoration(
                border: Border(bottom: BorderSide(color: _ReportTheme.line)),
              ),
              child: const Row(
                children: <Widget>[
                  Text(
                    '评估报告列表',
                    style: TextStyle(
                      color: _ReportTheme.ink,
                      fontSize: 19,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  SizedBox(width: 24),
                  _MetricChip(label: '近一月', value: '0'),
                ],
              ),
            ),
            const _ReportTableHeader(),
            const Expanded(child: _ReportSkeletonList()),
          ],
        ),
      ),
    );
  }
}

void _noopReportAction() {}

class _DomainBootstrapPlaceholderRow extends StatelessWidget {
  const _DomainBootstrapPlaceholderRow({
    required this.color,
    required this.width,
  });

  final Color color;
  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 47,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      child: Row(
        children: <Widget>[
          Container(
            width: 10,
            height: 10,
            decoration: BoxDecoration(color: color, shape: BoxShape.circle),
          ),
          const SizedBox(width: 10),
          _SkeletonBox(width: width, height: 14, radius: 7),
          const Spacer(),
          const Text(
            '0',
            style: TextStyle(
              color: _ReportTheme.muted,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
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
  const _RowActions({
    required this.record,
    required this.onViewReport,
    required this.onConfigRecord,
  });

  final Pep3RecordSummary record;
  final ValueChanged<Pep3RecordSummary> onViewReport;
  final ValueChanged<Pep3RecordSummary> onConfigRecord;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        _ActionButton(
          label: '查看',
          emphasized: true,
          onTap: () => onViewReport(record),
        ),
        const SizedBox(width: 8),
        _ActionButton(label: '配置', onTap: () => onConfigRecord(record)),
      ],
    );
  }
}

class _ActionButton extends StatelessWidget {
  const _ActionButton({
    required this.label,
    this.emphasized = false,
    this.onTap,
  });

  final String label;
  final bool emphasized;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(11),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(11),
        child: ConstrainedBox(
          constraints: const BoxConstraints(minWidth: 54),
          child: Ink(
            height: 32,
            padding: const EdgeInsets.symmetric(horizontal: 11),
            decoration: BoxDecoration(
              color:
                  emphasized ? const Color(0xFFFFF8F2) : _ReportTheme.surface,
              borderRadius: BorderRadius.circular(11),
              border: Border.all(
                color: emphasized ? const Color(0xFFF2CDBB) : _ReportTheme.line,
              ),
            ),
            child: Center(
              widthFactor: 1,
              child: Text(
                label,
                style: TextStyle(
                  color:
                      emphasized ? _ReportTheme.orangeDeep : _ReportTheme.text,
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _ReportConfigDialog extends StatefulWidget {
  const _ReportConfigDialog({
    required this.record,
    required this.token,
    required this.staffClient,
    required this.loadConfigDetail,
    required this.saveConfig,
  });

  final Pep3RecordSummary record;
  final String token;
  final TimetableClient staffClient;
  final Future<_RecordConfigDetail> Function(
    String token,
    Pep3RecordSummary record,
  ) loadConfigDetail;
  final Future<Pep3RecordSummary> Function(
    String token,
    Pep3RecordSummary record, {
    required String examinerName,
    required String assessmentDate,
  }) saveConfig;

  @override
  State<_ReportConfigDialog> createState() => _ReportConfigDialogState();
}

class _ReportConfigDialogState extends State<_ReportConfigDialog> {
  final LayerLink _teacherFieldLink = LayerLink();
  final GlobalKey _teacherFieldKey = GlobalKey();
  static const double _teacherFieldHeight = 56;
  late DateTime _assessmentDate;
  late List<String> _selectedExaminerNames;
  late String _originalExaminerName;
  DateTime? _originalAssessmentDate;
  List<ScheduleStaffOption> _teacherOptions = const <ScheduleStaffOption>[];
  bool _detailHydrating = true;
  bool _teacherLoading = false;
  bool _saving = false;
  String _errorMessage = '';
  String _teacherErrorMessage = '';
  bool _teacherSelectionTouched = false;
  bool _assessmentDateTouched = false;
  OverlayEntry? _teacherDropdownEntry;

  @override
  void initState() {
    super.initState();
    _assessmentDate = _configDateValue(widget.record.assessmentDate) ??
        _dateOnly(DateTime.now());
    _selectedExaminerNames = _uniqueExaminerNames(
      _splitExaminerNames(widget.record.examinerName),
    );
    _originalExaminerName = widget.record.examinerName.trim();
    _originalAssessmentDate = _originalAssessmentDateValue(widget.record) ??
        _configDateValue(widget.record.assessmentDate);
    _teacherOptions = _mergeTeacherOptions(const <ScheduleStaffOption>[]);
    unawaited(_loadTeacherOptions());
    unawaited(_hydrateOriginalDetail());
  }

  @override
  void dispose() {
    _removeTeacherDropdown();
    super.dispose();
  }

  Future<void> _hydrateOriginalDetail() async {
    final String token = widget.token.trim();
    if (token.isEmpty) {
      if (!mounted) {
        return;
      }
      setState(() {
        _detailHydrating = false;
      });
      return;
    }
    try {
      final _RecordConfigDetail detail = await widget.loadConfigDetail(
        token,
        widget.record,
      );
      if (!mounted) {
        return;
      }
      final String originalExaminerName =
          detail.originalExaminerName.trim().isNotEmpty
              ? detail.originalExaminerName.trim()
              : widget.record.examinerName.trim();
      final DateTime? originalAssessmentDate =
          _configDateValue(detail.originalAssessmentDate) ??
              _configDateValue(widget.record.assessmentDate);
      final List<String> currentExaminerNames = _uniqueExaminerNames(
        _splitExaminerNames(detail.currentExaminerName),
      );
      final DateTime? currentAssessmentDate =
          _configDateValue(detail.currentAssessmentDate);
      setState(() {
        if (!_teacherSelectionTouched && currentExaminerNames.isNotEmpty) {
          _selectedExaminerNames = currentExaminerNames;
        }
        if (!_assessmentDateTouched && currentAssessmentDate != null) {
          _assessmentDate = currentAssessmentDate;
        }
        _originalExaminerName = originalExaminerName;
        _originalAssessmentDate = originalAssessmentDate;
        _teacherOptions = _mergeTeacherOptions(_teacherOptions);
        _detailHydrating = false;
      });
    } on Object {
      if (!mounted) {
        return;
      }
      setState(() {
        if (!_teacherSelectionTouched) {
          _selectedExaminerNames = _uniqueExaminerNames(
            _splitExaminerNames(widget.record.examinerName),
          );
        }
        if (!_assessmentDateTouched) {
          _assessmentDate = _configDateValue(widget.record.assessmentDate) ??
              _dateOnly(DateTime.now());
        }
        _originalExaminerName = widget.record.examinerName.trim();
        _originalAssessmentDate = _originalAssessmentDateValue(widget.record) ??
            _configDateValue(widget.record.assessmentDate);
        _teacherOptions = _mergeTeacherOptions(_teacherOptions);
        _detailHydrating = false;
      });
    }
  }

  List<ScheduleStaffOption> _mergeTeacherOptions(
    List<ScheduleStaffOption> base,
  ) {
    final Map<String, ScheduleStaffOption> merged =
        <String, ScheduleStaffOption>{
      for (final ScheduleStaffOption option in base)
        if (option.name.trim().isNotEmpty) option.name.trim(): option,
    };
    for (final String name in <String>[
      ..._selectedExaminerNames,
      ..._splitExaminerNames(_originalExaminerName),
    ]) {
      final String trimmed = name.trim();
      if (trimmed.isEmpty || merged.containsKey(trimmed)) {
        continue;
      }
      merged[trimmed] = ScheduleStaffOption(id: 'name:$trimmed', name: trimmed);
    }
    return merged.values.toList();
  }

  Future<void> _loadTeacherOptions() async {
    final String token = widget.token.trim();
    if (token.isEmpty) {
      return;
    }
    setState(() {
      _teacherLoading = true;
      _teacherErrorMessage = '';
    });
    try {
      final List<ScheduleStaffOption> options =
          await widget.staffClient.fetchInstitutionStaffOptions(
        token,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _teacherOptions = _mergeTeacherOptions(options);
        _teacherLoading = false;
      });
      _markTeacherDropdownNeedsBuild();
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _teacherOptions = _mergeTeacherOptions(_teacherOptions);
        _teacherLoading = false;
        _teacherErrorMessage = '评估老师加载失败：$error';
      });
      _markTeacherDropdownNeedsBuild();
    }
  }

  void _toggleExaminerName(String name) {
    final String trimmed = name.trim();
    if (trimmed.isEmpty) {
      return;
    }
    setState(() {
      if (_selectedExaminerNames.contains(trimmed)) {
        _selectedExaminerNames = _selectedExaminerNames
            .where((String item) => item != trimmed)
            .toList();
      } else {
        _selectedExaminerNames = <String>[
          ..._selectedExaminerNames,
          trimmed,
        ];
      }
      _teacherSelectionTouched = true;
      _teacherOptions = _mergeTeacherOptions(_teacherOptions);
      _errorMessage = '';
    });
    _markTeacherDropdownNeedsBuild();
  }

  void _restoreOriginalExaminer() {
    final List<String> originalNames =
        _uniqueExaminerNames(_splitExaminerNames(_originalExaminerName));
    if (originalNames.isEmpty) {
      return;
    }
    setState(() {
      _selectedExaminerNames = originalNames;
      _teacherSelectionTouched = true;
      _teacherOptions = _mergeTeacherOptions(_teacherOptions);
      _errorMessage = '';
    });
    _markTeacherDropdownNeedsBuild();
  }

  void _restoreOriginalAssessmentDate() {
    final DateTime? originalDate = _originalAssessmentDate;
    if (originalDate == null) {
      return;
    }
    setState(() {
      _assessmentDate = originalDate;
      _assessmentDateTouched = true;
      _errorMessage = '';
    });
  }

  bool get _teacherDropdownOpen => _teacherDropdownEntry != null;

  void _toggleTeacherDropdown() {
    FocusManager.instance.primaryFocus?.unfocus();
    if (_teacherDropdownOpen) {
      _removeTeacherDropdown();
      return;
    }
    _showTeacherDropdown();
  }

  void _showTeacherDropdown() {
    final BuildContext? fieldContext = _teacherFieldKey.currentContext;
    if (fieldContext == null) {
      return;
    }
    final RenderBox box = fieldContext.findRenderObject()! as RenderBox;
    final Size fieldSize = box.size;
    _teacherDropdownEntry = OverlayEntry(
      builder: (BuildContext context) {
        return Stack(
          children: <Widget>[
            Positioned.fill(
              child: GestureDetector(
                behavior: HitTestBehavior.translucent,
                onTap: _removeTeacherDropdown,
                child: const SizedBox.expand(),
              ),
            ),
            CompositedTransformFollower(
              link: _teacherFieldLink,
              showWhenUnlinked: false,
              targetAnchor: Alignment.bottomLeft,
              followerAnchor: Alignment.topLeft,
              offset: const Offset(0, 8),
              child: Material(
                color: Colors.transparent,
                child: SizedBox(
                  width: fieldSize.width,
                  child: _ConfigTeacherDropdown(
                    loading: _teacherLoading,
                    errorMessage: _teacherErrorMessage,
                    options: _teacherOptions,
                    selectedNames: _selectedExaminerNames,
                    enabled: !_saving,
                    onToggleOption: (String name) => _toggleExaminerName(name),
                  ),
                ),
              ),
            ),
          ],
        );
      },
    );
    Overlay.of(context, rootOverlay: true).insert(_teacherDropdownEntry!);
    setState(() {});
    if (_teacherOptions.isEmpty && !_teacherLoading) {
      unawaited(_loadTeacherOptions());
    }
  }

  void _removeTeacherDropdown() {
    final OverlayEntry? entry = _teacherDropdownEntry;
    if (entry == null) {
      return;
    }
    _teacherDropdownEntry = null;
    entry.remove();
    if (mounted) {
      setState(() {});
    }
  }

  void _markTeacherDropdownNeedsBuild() {
    _teacherDropdownEntry?.markNeedsBuild();
  }

  Future<void> _pickAssessmentDate() async {
    FocusManager.instance.primaryFocus?.unfocus();
    _removeTeacherDropdown();
    final DateTime today = _dateOnly(DateTime.now());
    final DateTime? picked = await showPadDatePicker(
      context: context,
      initialDate:
          _dateOnly(_assessmentDate).isAfter(today) ? today : _assessmentDate,
      today: today,
      minDate: DateTime(today.year - 10, 1, 1),
      maxDate: today,
    );
    if (picked == null || !mounted) {
      return;
    }
    setState(() {
      _assessmentDate = _dateOnly(picked);
      _assessmentDateTouched = true;
      _errorMessage = '';
    });
  }

  Future<void> _saveConfig() async {
    FocusManager.instance.primaryFocus?.unfocus();
    _removeTeacherDropdown();
    final String token = widget.token.trim();
    final String examinerName = _joinExaminerNames(_selectedExaminerNames);
    if (examinerName.isEmpty) {
      setState(() => _errorMessage = '请选择评估老师');
      return;
    }
    if (token.isEmpty) {
      setState(() => _errorMessage = '请先登录后再保存');
      return;
    }
    setState(() {
      _saving = true;
      _errorMessage = '';
    });
    try {
      final Pep3RecordSummary detail = await widget.saveConfig(
        token,
        widget.record,
        examinerName: examinerName,
        assessmentDate: _dateText(_assessmentDate),
      );
      if (!mounted) {
        return;
      }
      Navigator.of(context).pop(detail);
    } on Pep3ApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _saving = false;
        _errorMessage = error.message;
      });
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _saving = false;
        _errorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _saving = false;
        _errorMessage = '保存评估配置失败：$error';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final String studentName = _studentName(widget.record);
    final String assessmentName = widget.record.assessmentName.trim().isEmpty
        ? 'PEP-3评估记录'
        : widget.record.assessmentName.trim();
    final bool configLocked = _saving || _detailHydrating;
    return PopScope(
      canPop: !_saving,
      child: Center(
        child: Material(
          color: Colors.transparent,
          child: Container(
            width: 560,
            constraints: const BoxConstraints(maxHeight: 640),
            padding: const EdgeInsets.fromLTRB(22, 22, 22, 20),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(color: _ReportTheme.line),
              boxShadow: const <BoxShadow>[
                BoxShadow(
                  color: Color(0x24000000),
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
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          const Text(
                            '配置评估记录',
                            style: TextStyle(
                              color: _ReportTheme.ink,
                              fontSize: 24,
                              fontWeight: FontWeight.w900,
                            ),
                          ),
                          const SizedBox(height: 7),
                          Text(
                            '$studentName / $assessmentName',
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              color: _ReportTheme.muted,
                              fontSize: 14,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 16),
                    Material(
                      color: Colors.transparent,
                      shape: const CircleBorder(),
                      child: InkWell(
                        onTap:
                            _saving ? null : () => Navigator.of(context).pop(),
                        customBorder: const CircleBorder(),
                        child: Container(
                          width: 38,
                          height: 38,
                          alignment: Alignment.center,
                          decoration: BoxDecoration(
                            color: const Color(0xFFFFF8F2),
                            shape: BoxShape.circle,
                            border: Border.all(color: _ReportTheme.lineSoft),
                          ),
                          child: const Icon(
                            Icons.close_rounded,
                            size: 22,
                            color: _ReportTheme.muted,
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 18),
                Container(
                  padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF7FBFF),
                    borderRadius: BorderRadius.circular(14),
                    border: Border.all(color: const Color(0xFFD9E9FF)),
                  ),
                  child: const Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Padding(
                        padding: EdgeInsets.only(top: 2),
                        child: Icon(
                          Icons.info_outline_rounded,
                          size: 18,
                          color: _ReportTheme.blue,
                        ),
                      ),
                      SizedBox(width: 10),
                      Expanded(
                        child: Text(
                          '仅修改评估老师和评估日期的展示信息，不重新计算测评结果，也不影响已生成IEP。',
                          style: TextStyle(
                            color: _ReportTheme.text,
                            fontSize: 13,
                            height: 1.5,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 16),
                Flexible(
                  child: SingleChildScrollView(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        const _RequiredFieldLabel(label: '评估老师'),
                        const SizedBox(height: 8),
                        _detailHydrating
                            ? const _ConfigLoadingField(
                                text: '正在加载评估老师',
                                height: _teacherFieldHeight,
                              )
                            : _ConfigTeacherSelectField(
                                fieldKey: _teacherFieldKey,
                                layerLink: _teacherFieldLink,
                                names: _selectedExaminerNames,
                                height: _teacherFieldHeight,
                                enabled: !_saving,
                                open: _teacherDropdownOpen,
                                onTap: _toggleTeacherDropdown,
                                onRemove: (String name) =>
                                    _toggleExaminerName(name),
                              ),
                        const SizedBox(height: 8),
                        Row(
                          children: <Widget>[
                            Expanded(
                              child: Text(
                                '原评估老师：${_detailHydrating ? '加载中...' : (_originalExaminerName.trim().isEmpty ? '-' : _originalExaminerName.trim())}',
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(
                                  color: _ReportTheme.muted,
                                  fontSize: 12,
                                  fontWeight: FontWeight.w800,
                                ),
                              ),
                            ),
                            const SizedBox(width: 12),
                            _InlineTextButton(
                              label: '恢复原评估老师',
                              enabled: !_detailHydrating &&
                                  _originalExaminerName.trim().isNotEmpty,
                              onTap: _restoreOriginalExaminer,
                            ),
                          ],
                        ),
                        const SizedBox(height: 16),
                        const _RequiredFieldLabel(label: '评估日期'),
                        const SizedBox(height: 8),
                        _detailHydrating
                            ? const _ConfigLoadingField(text: '正在加载评估日期')
                            : Material(
                                color: Colors.transparent,
                                borderRadius: BorderRadius.circular(14),
                                child: InkWell(
                                  onTap: _saving ? null : _pickAssessmentDate,
                                  borderRadius: BorderRadius.circular(14),
                                  child: Ink(
                                    height: 50,
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 14,
                                    ),
                                    decoration: BoxDecoration(
                                      color: const Color(0xFFFFFCF8),
                                      borderRadius: BorderRadius.circular(14),
                                      border:
                                          Border.all(color: _ReportTheme.line),
                                    ),
                                    child: Row(
                                      children: <Widget>[
                                        Expanded(
                                          child: Text(
                                            _dateText(_assessmentDate),
                                            style: const TextStyle(
                                              color: _ReportTheme.ink,
                                              fontSize: 15,
                                              fontWeight: FontWeight.w900,
                                            ),
                                          ),
                                        ),
                                        const Icon(
                                          Icons.calendar_month_rounded,
                                          size: 20,
                                          color: _ReportTheme.orangeDeep,
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              ),
                        const SizedBox(height: 8),
                        Row(
                          children: <Widget>[
                            Expanded(
                              child: Text(
                                '原评估日期：${_detailHydrating ? '加载中...' : (_originalAssessmentDate == null ? '-' : _dateText(_originalAssessmentDate!))}',
                                overflow: TextOverflow.ellipsis,
                                style: const TextStyle(
                                  color: _ReportTheme.muted,
                                  fontSize: 12,
                                  fontWeight: FontWeight.w800,
                                ),
                              ),
                            ),
                            const SizedBox(width: 12),
                            _InlineTextButton(
                              label: '恢复原评估日期',
                              enabled: !_detailHydrating &&
                                  _originalAssessmentDate != null,
                              onTap: _restoreOriginalAssessmentDate,
                            ),
                          ],
                        ),
                        if (_errorMessage.isNotEmpty) ...<Widget>[
                          const SizedBox(height: 16),
                          Container(
                            width: double.infinity,
                            padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
                            decoration: BoxDecoration(
                              color: const Color(0xFFFFF4F2),
                              borderRadius: BorderRadius.circular(14),
                              border: Border.all(
                                color: const Color(0xFFF4C7BE),
                              ),
                            ),
                            child: Text(
                              _errorMessage,
                              style: const TextStyle(
                                color: Color(0xFFB85A43),
                                fontSize: 13,
                                height: 1.45,
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 18),
                Row(
                  children: <Widget>[
                    Expanded(
                      child: _ConfigDialogButton(
                        label: '取消',
                        onTap:
                            _saving ? null : () => Navigator.of(context).pop(),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: _ConfigDialogButton(
                        label: _saving ? '保存中...' : '保存',
                        filled: true,
                        onTap: configLocked ? null : _saveConfig,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _ConfigTeacherSelectField extends StatelessWidget {
  const _ConfigTeacherSelectField({
    required this.fieldKey,
    required this.layerLink,
    required this.names,
    required this.height,
    required this.enabled,
    required this.open,
    required this.onTap,
    required this.onRemove,
  });

  final GlobalKey fieldKey;
  final LayerLink layerLink;
  final List<String> names;
  final double height;
  final bool enabled;
  final bool open;
  final VoidCallback onTap;
  final ValueChanged<String> onRemove;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final _ConfigTeacherFieldPreview preview = _buildTeacherFieldPreview(
          context,
          names,
          math.max(0, constraints.maxWidth - 12 - 12 - 10 - 22),
        );
        return CompositedTransformTarget(
          link: layerLink,
          child: Material(
            key: fieldKey,
            color: Colors.transparent,
            borderRadius: BorderRadius.circular(14),
            child: InkWell(
              onTap: enabled ? onTap : null,
              borderRadius: BorderRadius.circular(14),
              child: Ink(
                height: height,
                width: double.infinity,
                padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFFCF8),
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(
                    color: open ? _ReportTheme.orangeDeep : _ReportTheme.line,
                  ),
                ),
                child: Row(
                  children: <Widget>[
                    Expanded(
                      child: names.isEmpty
                          ? const Text(
                              '请选择评估老师',
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: TextStyle(
                                color: _ReportTheme.muted,
                                fontSize: 14,
                                fontWeight: FontWeight.w800,
                              ),
                            )
                          : SizedBox(
                              height: 32,
                              child: Row(
                                children: <Widget>[
                                  for (int index = 0;
                                      index < preview.visibleNames.length;
                                      index += 1) ...<Widget>[
                                    if (index > 0) const SizedBox(width: 8),
                                    ConstrainedBox(
                                      constraints: const BoxConstraints(
                                        maxWidth: 128,
                                      ),
                                      child: _ConfigTeacherChip(
                                        label: preview.visibleNames[index],
                                        enabled: enabled,
                                        onRemove: () => onRemove(
                                          preview.visibleNames[index],
                                        ),
                                      ),
                                    ),
                                  ],
                                  if (preview.hiddenCount > 0) ...<Widget>[
                                    if (preview.visibleNames.isNotEmpty)
                                      const SizedBox(width: 8),
                                    _ConfigTeacherOverflowChip(
                                      hiddenCount: preview.hiddenCount,
                                    ),
                                  ],
                                ],
                              ),
                            ),
                    ),
                    const SizedBox(width: 10),
                    Icon(
                      open
                          ? Icons.keyboard_arrow_up_rounded
                          : Icons.keyboard_arrow_down_rounded,
                      color: _ReportTheme.muted,
                      size: 22,
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }
}

class _ConfigTeacherFieldPreview {
  const _ConfigTeacherFieldPreview({
    required this.visibleNames,
    required this.hiddenCount,
  });

  final List<String> visibleNames;
  final int hiddenCount;
}

_ConfigTeacherFieldPreview _buildTeacherFieldPreview(
  BuildContext context,
  List<String> names,
  double maxWidth,
) {
  if (names.isEmpty || maxWidth <= 0) {
    return const _ConfigTeacherFieldPreview(
      visibleNames: <String>[],
      hiddenCount: 0,
    );
  }
  const double spacing = 8;
  double usedWidth = 0;
  final List<String> visibleNames = <String>[];
  for (int index = 0; index < names.length; index += 1) {
    final String name = names[index];
    final int remainingAfter = names.length - index - 1;
    final double chipWidth = _measureConfigTeacherChipWidth(context, name);
    final double nextWidth =
        usedWidth + (visibleNames.isEmpty ? 0 : spacing) + chipWidth;
    final double overflowWidth = remainingAfter > 0
        ? spacing +
            _measureConfigTeacherOverflowChipWidth(
              context,
              remainingAfter,
            )
        : 0;
    if (nextWidth + overflowWidth <= maxWidth || visibleNames.isEmpty) {
      visibleNames.add(name);
      usedWidth = nextWidth;
      continue;
    }
    return _ConfigTeacherFieldPreview(
      visibleNames: visibleNames,
      hiddenCount: names.length - visibleNames.length,
    );
  }
  return _ConfigTeacherFieldPreview(
    visibleNames: visibleNames,
    hiddenCount: 0,
  );
}

double _measureConfigTeacherChipWidth(BuildContext context, String name) {
  return math.min(
    128,
    _measureConfigTeacherTextWidth(
          context,
          name,
          const TextStyle(
            fontSize: 13,
            fontWeight: FontWeight.w900,
          ),
        ) +
        42,
  );
}

double _measureConfigTeacherOverflowChipWidth(
  BuildContext context,
  int hiddenCount,
) {
  return _measureConfigTeacherTextWidth(
        context,
        '+$hiddenCount',
        const TextStyle(
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ) +
      24;
}

double _measureConfigTeacherTextWidth(
  BuildContext context,
  String text,
  TextStyle style,
) {
  final TextPainter painter = TextPainter(
    text: TextSpan(text: text, style: style),
    textDirection: Directionality.of(context),
    maxLines: 1,
  )..layout();
  return painter.width;
}

class _ConfigLoadingField extends StatelessWidget {
  const _ConfigLoadingField({
    required this.text,
    this.height = 50,
  });

  final String text;
  final double height;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: height,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _ReportTheme.line),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: <Widget>[
          const SizedBox(
            width: 18,
            height: 18,
            child: CircularProgressIndicator(
              strokeWidth: 2.2,
              color: _ReportTheme.orangeDeep,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              text,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _ReportTheme.muted,
                fontSize: 14,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ConfigTeacherDropdown extends StatelessWidget {
  const _ConfigTeacherDropdown({
    required this.loading,
    required this.errorMessage,
    required this.options,
    required this.selectedNames,
    required this.enabled,
    required this.onToggleOption,
  });

  final bool loading;
  final String errorMessage;
  final List<ScheduleStaffOption> options;
  final List<String> selectedNames;
  final bool enabled;
  final ValueChanged<String> onToggleOption;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _ReportTheme.line),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x12000000),
            blurRadius: 16,
            offset: Offset(0, 8),
          ),
        ],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Container(
            height: 188,
            padding: const EdgeInsets.all(12),
            child: loading
                ? const Center(
                    child: SizedBox(
                      width: 22,
                      height: 22,
                      child: CircularProgressIndicator(
                        strokeWidth: 2.4,
                        color: _ReportTheme.orangeDeep,
                      ),
                    ),
                  )
                : options.isEmpty
                    ? Center(
                        child: Text(
                          errorMessage.isNotEmpty ? errorMessage : '暂无可选评估老师',
                          textAlign: TextAlign.center,
                          style: const TextStyle(
                            color: _ReportTheme.muted,
                            fontSize: 13,
                            height: 1.45,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      )
                    : ListView.separated(
                        itemCount: options.length,
                        separatorBuilder: (_, __) => const SizedBox(height: 8),
                        itemBuilder: (BuildContext context, int index) {
                          final ScheduleStaffOption option = options[index];
                          return _ConfigTeacherOptionTile(
                            title: option.name,
                            subtitle: option.subtitle,
                            selected: selectedNames.contains(option.name),
                            onTap: enabled
                                ? () => onToggleOption(option.name)
                                : null,
                          );
                        },
                      ),
          ),
        ],
      ),
    );
  }
}

class _RequiredFieldLabel extends StatelessWidget {
  const _RequiredFieldLabel({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return RichText(
      text: TextSpan(
        children: <InlineSpan>[
          const TextSpan(
            text: '* ',
            style: TextStyle(
              color: Color(0xFFFF5A4F),
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
          TextSpan(
            text: label,
            style: const TextStyle(
              color: _ReportTheme.ink,
              fontSize: 14,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _InlineTextButton extends StatelessWidget {
  const _InlineTextButton({
    required this.label,
    required this.enabled,
    required this.onTap,
  });

  final String label;
  final bool enabled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: enabled ? onTap : null,
        borderRadius: BorderRadius.circular(10),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 2, vertical: 2),
          child: Text(
            label,
            style: TextStyle(
              color: enabled ? _ReportTheme.blue : _ReportTheme.muted,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _ConfigTeacherChip extends StatelessWidget {
  const _ConfigTeacherChip({
    required this.label,
    required this.enabled,
    required this.onRemove,
  });

  final String label;
  final bool enabled;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 32,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF1E8),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: const Color(0xFFF1C7B0)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Flexible(
            child: Text(
              label,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _ReportTheme.orangeDeep,
                fontSize: 13,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          const SizedBox(width: 6),
          Material(
            color: Colors.transparent,
            shape: const CircleBorder(),
            child: InkWell(
              onTap: enabled ? onRemove : null,
              customBorder: const CircleBorder(),
              child: const Padding(
                padding: EdgeInsets.all(1),
                child: Icon(
                  Icons.close_rounded,
                  size: 16,
                  color: _ReportTheme.orangeDeep,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ConfigTeacherOverflowChip extends StatelessWidget {
  const _ConfigTeacherOverflowChip({required this.hiddenCount});

  final int hiddenCount;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 32,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFF7F1EB),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: const Color(0xFFE8D7C7)),
      ),
      alignment: Alignment.center,
      child: Text(
        '+$hiddenCount',
        style: const TextStyle(
          color: _ReportTheme.text,
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _ConfigTeacherOptionTile extends StatelessWidget {
  const _ConfigTeacherOptionTile({
    required this.title,
    required this.subtitle,
    required this.selected,
    required this.onTap,
  });

  final String title;
  final String subtitle;
  final bool selected;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(11),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(11),
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 160),
          constraints: const BoxConstraints(minHeight: 44),
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFF1E8) : Colors.white,
            borderRadius: BorderRadius.circular(11),
            border: Border.all(
              color: selected ? _ReportTheme.orange : _ReportTheme.lineSoft,
            ),
          ),
          child: Row(
            children: <Widget>[
              Icon(
                selected
                    ? Icons.check_circle_rounded
                    : Icons.radio_button_unchecked_rounded,
                size: 18,
                color: selected ? _ReportTheme.orangeDeep : _ReportTheme.muted,
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text(
                      title,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _ReportTheme.ink,
                        fontSize: 13,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    if (subtitle.trim().isNotEmpty) ...<Widget>[
                      const SizedBox(height: 2),
                      Text(
                        subtitle,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          color: _ReportTheme.muted,
                          fontSize: 10,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _ConfigDialogButton extends StatelessWidget {
  const _ConfigDialogButton({
    required this.label,
    this.filled = false,
    this.onTap,
  });

  final String label;
  final bool filled;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Ink(
          height: 48,
          decoration: BoxDecoration(
            color: filled ? _ReportTheme.orangeDeep : const Color(0xFFFFFCF8),
            borderRadius: BorderRadius.circular(14),
            border: Border.all(
              color: filled ? _ReportTheme.orangeDeep : _ReportTheme.line,
            ),
          ),
          child: Center(
            child: Text(
              label,
              style: TextStyle(
                color: filled ? Colors.white : _ReportTheme.text,
                fontSize: 15,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _ReportModuleOption {
  const _ReportModuleOption({
    required this.value,
    required this.label,
    required this.pages,
    required this.pageCount,
    required this.description,
    this.recommended = false,
  });

  final String value;
  final String label;
  final String pages;
  final int pageCount;
  final String description;
  final bool recommended;
}

const List<_ReportModuleOption> _reportModuleOptions = <_ReportModuleOption>[
  _ReportModuleOption(
    value: 'test_score',
    label: '测验分数',
    pages: '第 1 页',
    pageCount: 1,
    description: '导出首页测验分数汇总，适合快速归档总览。',
  ),
  _ReportModuleOption(
    value: 'development_profile',
    label: '发展表现图',
    pages: '第 19 页',
    pageCount: 1,
    description: '只导出发展表现图，用于查看各领域发展曲线。',
  ),
  _ReportModuleOption(
    value: 'score_and_profile',
    label: '分数+表现图',
    pages: '第 1、19 页',
    pageCount: 2,
    description: '包含测验分数汇总和发展表现图，适合简版报告。',
    recommended: true,
  ),
  _ReportModuleOption(
    value: 'scoring_tables',
    label: '评分表',
    pages: '第 2-18 页',
    pageCount: 17,
    description: '导出儿童表现记录、评分统计和照顾者评分表。',
  ),
];

const double _reportPreviewRasterDpi = 96;
const double _erxinReportPreviewRasterDpi = 216;

String _reportModuleInlineDescription(_ReportModuleOption option) {
  final String pages = option.pages.replaceAll(RegExp(r'\s+'), '');
  return '$pages：${option.description}';
}

class _ReportPreviewDialog extends StatefulWidget {
  const _ReportPreviewDialog({
    required this.record,
    required this.token,
    required this.client,
  });

  final Pep3RecordSummary record;
  final String token;
  final Pep3AssessmentClient client;

  @override
  State<_ReportPreviewDialog> createState() => _ReportPreviewDialogState();
}

class _ReportPreviewDialogState extends State<_ReportPreviewDialog> {
  late _ReportModuleOption _activeOption;
  late Pep3RecordSummary _displayRecord;
  Uint8List? _pdfBytes;
  ErxinReportInterpretation? _interpretation;
  Future<ErxinReportInterpretation>? _interpretationLoad;
  final Map<String, Uint8List> _modulePdfBytes = <String, Uint8List>{};
  final Map<String, Future<Uint8List>> _modulePdfLoads =
      <String, Future<Uint8List>>{};
  bool _loading = true;
  bool _interpretationLoading = false;
  bool _interpretationGenerating = false;
  bool _interpretationFetched = false;
  bool _showInterpretation = false;
  bool _printing = false;
  String _errorMessage = '';
  String _interpretationErrorMessage = '';
  String _interpretationProgressMessage = '准备生成报告解读...';
  String _interpretationStreamingText = '';
  int _interpretationGenerateSerial = 0;
  int _recordSyncSerial = 0;

  @override
  void initState() {
    super.initState();
    _displayRecord = widget.record;
    _activeOption = _reportModuleOptions.firstWhere(
      (_ReportModuleOption option) => option.recommended,
      orElse: () => _reportModuleOptions.first,
    );
    unawaited(_syncLatestRecord());
    unawaited(_bootstrapPreview());
  }

  Future<void> _bootstrapPreview() async {
    await _activateModule(_activeOption);
    await _prewarmOtherModules();
  }

  Future<void> _syncLatestRecord() async {
    final String token = widget.token.trim();
    if (token.isEmpty) {
      return;
    }
    final int serial = ++_recordSyncSerial;
    try {
      final Pep3RecordDetail detail = await widget.client.fetchRecordDetail(
        token,
        widget.record.id,
      );
      if (!mounted || serial != _recordSyncSerial) {
        return;
      }
      setState(() {
        _displayRecord = detail;
      });
    } on Object {
      if (!mounted || serial != _recordSyncSerial) {
        return;
      }
    }
  }

  void _retryPreview() {
    unawaited(_syncLatestRecord());
    unawaited(_refreshPreviewModules());
  }

  void _selectResultTab() {
    if (_showInterpretation) {
      setState(() {
        _showInterpretation = false;
      });
    }
  }

  void _selectInterpretationTab() {
    if (!_showInterpretation) {
      setState(() {
        _showInterpretation = true;
      });
    }
    if (!_interpretationFetched && !_interpretationLoading) {
      unawaited(_loadSavedInterpretation());
    }
  }

  Future<void> _loadSavedInterpretation() async {
    if (!mounted) {
      return;
    }
    final String token = widget.token.trim();
    if (token.isEmpty) {
      setState(() {
        _interpretationFetched = true;
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationErrorMessage = '请先登录后再查看报告解读';
      });
      return;
    }
    setState(() {
      _interpretationLoading = true;
      _interpretationGenerating = false;
      _interpretationErrorMessage = '';
      _interpretationStreamingText = '';
      _interpretationProgressMessage = '正在读取已保存的报告解读...';
    });
    final Future<ErxinReportInterpretation> future =
        widget.client.fetchRecordReportInterpretation(
      token,
      widget.record.id,
    );
    _interpretationLoad = future;
    try {
      final ErxinReportInterpretation interpretation = await future;
      if (!mounted || !identical(_interpretationLoad, future)) {
        return;
      }
      setState(() {
        _interpretation = interpretation;
        _interpretationFetched = true;
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationStreamingText = '';
        _interpretationProgressMessage =
            interpretation.isEmpty ? '报告解读尚未生成' : '已读取保存的报告解读';
      });
    } on Pep3ApiException catch (error) {
      if (!mounted || !identical(_interpretationLoad, future)) {
        return;
      }
      setState(() {
        _interpretationFetched = true;
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationErrorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted || !identical(_interpretationLoad, future)) {
        return;
      }
      setState(() {
        _interpretationFetched = true;
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationErrorMessage = '报告解读读取失败：$error';
      });
    }
  }

  Future<void> _generateInterpretation({bool regenerate = false}) async {
    if (!mounted) {
      return;
    }
    final String token = widget.token.trim();
    if (token.isEmpty) {
      setState(() {
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationErrorMessage = '请先登录后再生成报告解读';
      });
      return;
    }
    final int serial = ++_interpretationGenerateSerial;
    setState(() {
      _interpretationLoading = true;
      _interpretationGenerating = true;
      _interpretationFetched = true;
      _interpretationErrorMessage = '';
      _interpretationStreamingText = '';
      _interpretationProgressMessage =
          regenerate ? '正在重新生成报告解读...' : '正在生成报告解读...';
      if (regenerate) {
        _interpretation = null;
      }
    });
    try {
      bool completed = false;
      await for (final ErxinReportInterpretationStreamEvent event in widget
          .client
          .generateRecordReportInterpretationStream(token, widget.record.id)) {
        if (!mounted || serial != _interpretationGenerateSerial) {
          return;
        }
        if (event.type == 'status') {
          setState(() {
            _interpretationProgressMessage = event.message.trim().isEmpty
                ? 'AI 正在分析PEP-3评估结果...'
                : event.message.trim();
          });
          continue;
        }
        if (event.type == 'delta') {
          if (event.text.isEmpty) {
            continue;
          }
          setState(() {
            _interpretationStreamingText += event.text;
            _interpretationProgressMessage = 'AI 正在生成报告解读...';
          });
          continue;
        }
        if (event.type == 'error') {
          throw Pep3ApiException(
            event.message.trim().isEmpty ? '报告解读生成失败' : event.message.trim(),
          );
        }
        if (event.type == 'done') {
          final ErxinReportInterpretation interpretation =
              event.data ?? ErxinReportInterpretation.empty;
          setState(() {
            _interpretation = interpretation;
            _interpretationLoading = false;
            _interpretationGenerating = false;
            _interpretationStreamingText = '';
            _interpretationProgressMessage = '报告解读已生成';
          });
          completed = true;
          break;
        }
      }
      if (!mounted || serial != _interpretationGenerateSerial || completed) {
        return;
      }
      setState(() {
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationErrorMessage = '报告解读生成中断，请重新生成';
      });
    } on Pep3ApiException catch (error) {
      if (!mounted || serial != _interpretationGenerateSerial) {
        return;
      }
      setState(() {
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationStreamingText = '';
        _interpretationErrorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted || serial != _interpretationGenerateSerial) {
        return;
      }
      setState(() {
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationStreamingText = '';
        _interpretationErrorMessage = '报告解读生成失败：$error';
      });
    }
  }

  Future<void> _refreshPreviewModules() async {
    await _activateModule(_activeOption, refresh: true);
    await _prewarmOtherModules(refresh: true);
  }

  Future<void> _prewarmOtherModules({bool refresh = false}) async {
    for (final _ReportModuleOption option in _reportModuleOptions) {
      if (option.value == _activeOption.value) {
        continue;
      }
      try {
        await _ensureModulePdf(option, refresh: refresh);
      } on Object {
        // 预热失败不影响当前弹窗。
      }
    }
  }

  Future<Uint8List> _ensureModulePdf(
    _ReportModuleOption option, {
    bool refresh = false,
  }) {
    final String key = option.value;
    if (!refresh) {
      final Uint8List? cached = _modulePdfBytes[key];
      if (cached != null) {
        return Future<Uint8List>.value(cached);
      }
      final Future<Uint8List>? pending = _modulePdfLoads[key];
      if (pending != null) {
        return pending;
      }
    }
    final String token = widget.token.trim();
    if (token.isEmpty) {
      return Future<Uint8List>.error(
        const Pep3ApiException('请先登录后再查看评估报告'),
      );
    }
    late final Future<Uint8List> future;
    future = (() async {
      try {
        final Uint8List bytes = await widget.client.downloadRecordBookletPdf(
          token,
          widget.record.id,
          dimension: key,
        );
        _modulePdfBytes[key] = bytes;
        return bytes;
      } finally {
        if (identical(_modulePdfLoads[key], future)) {
          _modulePdfLoads.remove(key);
        }
      }
    })();
    _modulePdfLoads[key] = future;
    return future;
  }

  Future<void> _activateModule(
    _ReportModuleOption option, {
    bool refresh = false,
  }) async {
    final String key = option.value;
    if (!mounted) {
      return;
    }
    setState(() {
      _activeOption = option;
      _errorMessage = '';
    });

    final Uint8List? cached = !refresh ? _modulePdfBytes[key] : null;
    if (cached != null) {
      if (!mounted) {
        return;
      }
      setState(() {
        _pdfBytes = cached;
        _loading = false;
      });
      return;
    }

    if (!mounted) {
      return;
    }
    setState(() {
      _loading = true;
      _pdfBytes = null;
    });

    try {
      final Uint8List bytes = await _ensureModulePdf(option, refresh: refresh);
      if (!mounted || _activeOption.value != key) {
        return;
      }
      setState(() {
        _pdfBytes = bytes;
        _loading = false;
      });
    } on Pep3ApiException catch (error) {
      if (!mounted || _activeOption.value != key) {
        return;
      }
      setState(() {
        _pdfBytes = null;
        _loading = false;
        _errorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted || _activeOption.value != key) {
        return;
      }
      setState(() {
        _pdfBytes = null;
        _loading = false;
        _errorMessage = '评估报告加载失败：$error';
      });
    }
  }

  void _selectOption(_ReportModuleOption option) {
    if (_activeOption.value == option.value) {
      return;
    }
    unawaited(_activateModule(option));
  }

  Future<void> _printCurrentModule() async {
    if (_printing) {
      return;
    }
    final _ReportModuleOption option = _activeOption;
    setState(() {
      _printing = true;
      _errorMessage = '';
    });
    try {
      final Uint8List bytes =
          _pdfBytes ?? await _ensureModulePdf(option, refresh: false);
      if (!mounted || _activeOption.value != option.value) {
        return;
      }
      if (bytes.isEmpty) {
        setState(() {
          _errorMessage = '暂无可打印的评估报告';
        });
        return;
      }
      await Printing.layoutPdf(
        name: _pep3PrintFileName(_displayRecord, option.label),
        onLayout: (_) async => bytes,
      );
    } on Pep3ApiException catch (error) {
      if (!mounted || _activeOption.value != option.value) {
        return;
      }
      setState(() {
        _errorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted || _activeOption.value != option.value) {
        return;
      }
      setState(() {
        _errorMessage = '评估报告打印失败：$error';
      });
    } finally {
      if (mounted) {
        setState(() {
          _printing = false;
        });
      }
    }
  }

  Future<void> _printCurrentTab() async {
    if (_printing) {
      return;
    }
    if (_showInterpretation) {
      await _printInterpretation();
      return;
    }
    await _printCurrentModule();
  }

  Future<void> _printInterpretation() async {
    ErxinReportInterpretation? interpretation = _interpretation;
    if (interpretation == null || interpretation.isEmpty) {
      if (!_interpretationFetched && !_interpretationLoading) {
        await _loadSavedInterpretation();
        interpretation = _interpretation;
      }
    }
    if (!mounted) {
      return;
    }
    if (interpretation == null || interpretation.isEmpty) {
      setState(() {
        _interpretationErrorMessage = '请先生成报告解读后再打印';
      });
      return;
    }
    setState(() {
      _printing = true;
      _interpretationErrorMessage = '';
    });
    try {
      final Uint8List bytes =
          await widget.client.downloadRecordReportInterpretationPdf(
        widget.token.trim(),
        widget.record.id,
      );
      await Printing.layoutPdf(
        name: _pep3PrintFileName(_displayRecord, '报告解读'),
        onLayout: (_) async => bytes,
      );
    } on Pep3ApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _interpretationErrorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _interpretationErrorMessage = '报告解读打印失败：$error';
      });
    } finally {
      if (mounted) {
        setState(() {
          _printing = false;
        });
      }
    }
  }

  Future<void> _handleInterpretationGenerateTap() async {
    final bool regenerate =
        _interpretation != null && !_interpretation!.isEmpty;
    if (!regenerate) {
      await _generateInterpretation();
      return;
    }
    final bool confirmed = await _confirmRegenerateInterpretation();
    if (!mounted || !confirmed) {
      return;
    }
    await _generateInterpretation(regenerate: true);
  }

  Future<bool> _confirmRegenerateInterpretation() async {
    final bool? confirmed = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext dialogContext) {
        return PadDialogViewport(
          child: Center(
            child: _ErxinRegenerateInterpretationConfirmDialog(
              onCancel: () => Navigator.of(dialogContext).pop(false),
              onConfirm: () => Navigator.of(dialogContext).pop(true),
            ),
          ),
        );
      },
    );
    return confirmed == true;
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      child: Center(
        child: Material(
          color: Colors.transparent,
          child: Container(
            width: 980,
            height: 654,
            padding: const EdgeInsets.fromLTRB(22, 22, 22, 22),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(color: _ReportTheme.line),
              boxShadow: const <BoxShadow>[
                BoxShadow(
                  color: Color(0x24000000),
                  blurRadius: 34,
                  offset: Offset(0, 18),
                ),
              ],
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                _buildHeader(context),
                const SizedBox(height: 14),
                _buildTabBar(),
                const SizedBox(height: 12),
                Expanded(
                  child: _showInterpretation
                      ? _buildInterpretationContent()
                      : _buildContent(),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildHeader(BuildContext context) {
    final Pep3RecordSummary record = _displayRecord;
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              const Text(
                '评估报告',
                style: TextStyle(
                  color: _ReportTheme.ink,
                  fontSize: 24,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 7),
              Text.rich(
                TextSpan(
                  children: <InlineSpan>[
                    TextSpan(
                      text:
                          '${record.assessmentName.trim().isEmpty ? 'PEP-3测试员记录册' : record.assessmentName}   ${_studentName(record)} / ${_dateOnlyText(record.assessmentDate)}   ',
                    ),
                    TextSpan(
                      text: _showInterpretation
                          ? '报告解读：AI 会基于PEP-3评估结果生成并保存。'
                          : _reportModuleInlineDescription(_activeOption),
                      style: const TextStyle(color: _ReportTheme.blue),
                    ),
                  ],
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _ReportTheme.muted,
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(width: 16),
        Material(
          color: Colors.transparent,
          shape: const CircleBorder(),
          child: InkWell(
            onTap: () => Navigator.of(context).pop(),
            customBorder: const CircleBorder(),
            child: Container(
              width: 38,
              height: 38,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: const Color(0xFFFFF8F2),
                shape: BoxShape.circle,
                border: Border.all(color: _ReportTheme.lineSoft),
              ),
              child: const Icon(
                Icons.close_rounded,
                size: 22,
                color: _ReportTheme.muted,
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildTabBar() {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final bool stackMeta = constraints.maxWidth < 900;
        final Widget chips = Wrap(
          spacing: 8,
          runSpacing: 8,
          children: <Widget>[
            _ErxinReportTabChip(
              label: '测试员记录册',
              active: !_showInterpretation,
              onTap: _selectResultTab,
            ),
            _ErxinReportTabChip(
              label: '报告解读',
              active: _showInterpretation,
              onTap: _selectInterpretationTab,
            ),
            if (!_showInterpretation)
              for (final _ReportModuleOption option in _reportModuleOptions)
                _ReportModuleChip(
                  option: option,
                  active: option.value == _activeOption.value,
                  onTap: () => _selectOption(option),
                ),
          ],
        );
        final Widget generateButton = _ToolbarButton(
          label: _interpretationLoading
              ? (_interpretationGenerating ? '生成中' : '读取中')
              : (_interpretation == null || _interpretation!.isEmpty)
                  ? '生成解读'
                  : '重新生成解读',
          icon: Icons.auto_awesome_rounded,
          filled: true,
          onTap: _interpretationLoading
              ? null
              : () => unawaited(_handleInterpretationGenerateTap()),
        );
        final Widget printButton = _ToolbarButton(
          label: _printing ? '打印中' : '打印',
          icon: Icons.print_rounded,
          onTap: (_printing ||
                  (!_showInterpretation && _loading) ||
                  (_showInterpretation && _interpretationLoading))
              ? null
              : () => unawaited(_printCurrentTab()),
        );
        final Widget actions = Row(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            if (_showInterpretation) ...<Widget>[
              generateButton,
              const SizedBox(width: 8),
            ],
            printButton,
          ],
        );

        return Container(
          padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
          decoration: BoxDecoration(
            color: const Color(0xFFFFFCF8),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: _ReportTheme.lineSoft),
          ),
          child: stackMeta
              ? Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    chips,
                    const SizedBox(height: 8),
                    Align(
                      alignment: Alignment.centerRight,
                      child: actions,
                    ),
                  ],
                )
              : Row(
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: <Widget>[
                    Expanded(child: chips),
                    const SizedBox(width: 14),
                    actions,
                  ],
                ),
        );
      },
    );
  }

  Widget _buildContent() {
    if (_loading) {
      return const _ReportPreviewLoadingState(message: '评估报告加载中...');
    }
    if (_errorMessage.isNotEmpty) {
      return _ReportPreviewErrorState(
        message: _errorMessage,
        onRetry: _retryPreview,
      );
    }
    final Uint8List? bytes = _pdfBytes;
    if (bytes == null || bytes.isEmpty) {
      return const _ReportPreviewEmptyState(message: '暂无评估报告内容');
    }

    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFFDF8F3),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      clipBehavior: Clip.antiAlias,
      child: _LazyReportPdfPreview(
        key: ValueKey<String>(
          'report-pdf-${widget.record.id}-${widget.record.updatedTime}-${_activeOption.value}-${bytes.length}',
        ),
        bytes: bytes,
        pageCount: _activeOption.pageCount,
      ),
    );
  }

  Widget _buildInterpretationContent() {
    if (_interpretationLoading) {
      if (_interpretationGenerating) {
        return _ErxinInterpretationProgressState(
          message: _interpretationProgressMessage,
          streamingText: _interpretationStreamingText,
          domainSectionTitle: '领域表现',
        );
      }
      return _ErxinInterpretationReadLoadingState(
        message: _interpretationProgressMessage,
      );
    }
    if (_interpretationErrorMessage.isNotEmpty) {
      return _ReportPreviewErrorState(
        message: _interpretationErrorMessage,
        onRetry: () => unawaited(_generateInterpretation(regenerate: true)),
      );
    }
    final ErxinReportInterpretation? interpretation = _interpretation;
    if (interpretation == null || interpretation.isEmpty) {
      return _ErxinInterpretationEmptyState(
        message: '报告解读尚未生成',
        detail: '点击“生成解读”后，AI 会基于当前评估结果生成并保存。',
        actionLabel: '生成解读',
        onAction: () => unawaited(_generateInterpretation()),
      );
    }
    return _ErxinInterpretationView(
      interpretation: interpretation,
      domainSectionTitle: '领域表现',
    );
  }
}

class _ReportPdfPageSnapshot {
  const _ReportPdfPageSnapshot({
    required this.raster,
    required this.width,
    required this.height,
  });

  final PdfRaster raster;
  final int width;
  final int height;

  double get aspectRatio {
    if (width <= 0 || height <= 0) {
      return 0.76;
    }
    return width / height;
  }
}

class _ErxinReportPreviewDialog extends StatefulWidget {
  const _ErxinReportPreviewDialog({
    required this.record,
    required this.token,
    required this.client,
  });

  final Pep3RecordSummary record;
  final String token;
  final ErxinAssessmentClient client;

  @override
  State<_ErxinReportPreviewDialog> createState() =>
      _ErxinReportPreviewDialogState();
}

class _ErxinReportPreviewDialogState extends State<_ErxinReportPreviewDialog> {
  Pep3RecordSummary? _displayRecord;
  Uint8List? _pdfBytes;
  ErxinReportInterpretation? _interpretation;
  Future<ErxinReportInterpretation>? _interpretationLoad;
  bool _loading = true;
  bool _interpretationLoading = false;
  bool _interpretationGenerating = false;
  bool _interpretationFetched = false;
  bool _showInterpretation = false;
  bool _printing = false;
  String _errorMessage = '';
  String _interpretationErrorMessage = '';
  String _interpretationProgressMessage = '准备生成报告解读...';
  String _interpretationStreamingText = '';
  int _interpretationGenerateSerial = 0;
  int _recordSyncSerial = 0;
  Future<Uint8List>? _pdfLoad;

  @override
  void initState() {
    super.initState();
    _displayRecord = widget.record;
    unawaited(_syncLatestRecord());
    unawaited(_loadPdf());
  }

  Future<void> _syncLatestRecord() async {
    final String token = widget.token.trim();
    if (token.isEmpty) {
      return;
    }
    final int serial = ++_recordSyncSerial;
    try {
      final ErxinRecordDetail detail = await widget.client.fetchRecordDetail(
        token,
        widget.record.id,
      );
      if (!mounted || serial != _recordSyncSerial) {
        return;
      }
      setState(() {
        _displayRecord = Pep3RecordSummary(
          id: detail.id,
          studentId: detail.studentId,
          studentName: detail.studentName,
          studentGender: detail.studentGender,
          studentAvatar: detail.studentAvatar,
          studentPhone: detail.studentPhone,
          assessmentCode: detail.assessmentCode,
          assessmentName: detail.assessmentName,
          scaleCategory: detail.scaleCategory,
          scaleVersion: detail.scaleVersion,
          birthDate: detail.birthDate,
          assessmentDate: detail.assessmentDate,
          ageYears: detail.ageYears,
          ageMonths: detail.ageMonths,
          ageDays: detail.ageDays,
          normAgeMonths: detail.normAgeMonths,
          assessmentSequence: detail.assessmentSequence,
          examinerName: detail.examinerName,
          createdTime: detail.createdTime,
          updatedTime: detail.updatedTime,
        );
      });
    } on Object {
      if (!mounted || serial != _recordSyncSerial) {
        return;
      }
    }
  }

  Future<void> _loadPdf({bool refresh = false}) async {
    if (!mounted) {
      return;
    }
    if (!refresh && _pdfBytes != null && _pdfBytes!.isNotEmpty) {
      setState(() {
        _loading = false;
        _errorMessage = '';
      });
      return;
    }
    setState(() {
      _loading = true;
      _errorMessage = '';
      if (refresh) {
        _pdfBytes = null;
      }
    });
    final String token = widget.token.trim();
    if (token.isEmpty) {
      setState(() {
        _loading = false;
        _errorMessage = '请先登录后再查看评估报告';
      });
      return;
    }
    final Future<Uint8List> future =
        widget.client.downloadRecordReportPdf(token, widget.record.id);
    _pdfLoad = future;
    try {
      final Uint8List bytes = await future;
      if (!mounted || !identical(_pdfLoad, future)) {
        return;
      }
      setState(() {
        _pdfBytes = bytes;
        _loading = false;
      });
    } on AssessmentScaleApiException catch (error) {
      if (!mounted || !identical(_pdfLoad, future)) {
        return;
      }
      setState(() {
        _pdfBytes = null;
        _loading = false;
        _errorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted || !identical(_pdfLoad, future)) {
        return;
      }
      setState(() {
        _pdfBytes = null;
        _loading = false;
        _errorMessage = '评估报告加载失败：$error';
      });
    }
  }

  void _retryPreview() {
    unawaited(_syncLatestRecord());
    unawaited(_loadPdf(refresh: true));
  }

  void _selectResultTab() {
    if (_showInterpretation) {
      setState(() {
        _showInterpretation = false;
      });
    }
  }

  void _selectInterpretationTab() {
    if (!_showInterpretation) {
      setState(() {
        _showInterpretation = true;
      });
    }
    if (!_interpretationFetched && !_interpretationLoading) {
      unawaited(_loadSavedInterpretation());
    }
  }

  Future<void> _loadSavedInterpretation() async {
    if (!mounted) {
      return;
    }
    final String token = widget.token.trim();
    if (token.isEmpty) {
      setState(() {
        _interpretationFetched = true;
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationErrorMessage = '请先登录后再查看报告解读';
      });
      return;
    }
    setState(() {
      _interpretationLoading = true;
      _interpretationGenerating = false;
      _interpretationErrorMessage = '';
      _interpretationStreamingText = '';
      _interpretationProgressMessage = '正在读取已保存的报告解读...';
    });
    final Future<ErxinReportInterpretation> future =
        widget.client.fetchRecordReportInterpretation(
      token,
      widget.record.id,
    );
    _interpretationLoad = future;
    try {
      final ErxinReportInterpretation interpretation = await future;
      if (!mounted || !identical(_interpretationLoad, future)) {
        return;
      }
      setState(() {
        _interpretation = interpretation;
        _interpretationFetched = true;
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationStreamingText = '';
        _interpretationProgressMessage =
            interpretation.isEmpty ? '报告解读尚未生成' : '已读取保存的报告解读';
      });
    } on AssessmentScaleApiException catch (error) {
      if (!mounted || !identical(_interpretationLoad, future)) {
        return;
      }
      setState(() {
        _interpretationFetched = true;
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationErrorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted || !identical(_interpretationLoad, future)) {
        return;
      }
      setState(() {
        _interpretationFetched = true;
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationErrorMessage = '报告解读读取失败：$error';
      });
    }
  }

  Future<void> _generateInterpretation({bool regenerate = false}) async {
    if (!mounted) {
      return;
    }
    final String token = widget.token.trim();
    if (token.isEmpty) {
      setState(() {
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationErrorMessage = '请先登录后再生成报告解读';
      });
      return;
    }
    final int serial = ++_interpretationGenerateSerial;
    setState(() {
      _interpretationLoading = true;
      _interpretationGenerating = true;
      _interpretationFetched = true;
      _interpretationErrorMessage = '';
      _interpretationStreamingText = '';
      _interpretationProgressMessage =
          regenerate ? '正在重新生成报告解读...' : '正在生成报告解读...';
      if (regenerate) {
        _interpretation = null;
      }
    });
    try {
      bool completed = false;
      await for (final ErxinReportInterpretationStreamEvent event in widget
          .client
          .generateRecordReportInterpretationStream(token, widget.record.id)) {
        if (!mounted || serial != _interpretationGenerateSerial) {
          return;
        }
        if (event.type == 'status') {
          setState(() {
            _interpretationProgressMessage = event.message.trim().isEmpty
                ? 'AI 正在分析全量表与五大能区结果...'
                : event.message.trim();
          });
          continue;
        }
        if (event.type == 'delta') {
          if (event.text.isEmpty) {
            continue;
          }
          setState(() {
            _interpretationStreamingText += event.text;
            _interpretationProgressMessage = 'AI 正在生成报告解读...';
          });
          continue;
        }
        if (event.type == 'error') {
          throw AssessmentScaleApiException(
            event.message.trim().isEmpty ? '报告解读生成失败' : event.message.trim(),
          );
        }
        if (event.type == 'done') {
          final ErxinReportInterpretation interpretation =
              event.data ?? ErxinReportInterpretation.empty;
          setState(() {
            _interpretation = interpretation;
            _interpretationLoading = false;
            _interpretationGenerating = false;
            _interpretationStreamingText = '';
            _interpretationProgressMessage = '报告解读已生成';
          });
          completed = true;
          break;
        }
      }
      if (!mounted || serial != _interpretationGenerateSerial || completed) {
        return;
      }
      setState(() {
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationErrorMessage = '报告解读生成中断，请重新生成';
      });
    } on AssessmentScaleApiException catch (error) {
      if (!mounted || serial != _interpretationGenerateSerial) {
        return;
      }
      setState(() {
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationStreamingText = '';
        _interpretationErrorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted || serial != _interpretationGenerateSerial) {
        return;
      }
      setState(() {
        _interpretationLoading = false;
        _interpretationGenerating = false;
        _interpretationStreamingText = '';
        _interpretationErrorMessage = '报告解读生成失败：$error';
      });
    }
  }

  Future<void> _printCurrentTab() async {
    if (_printing) {
      return;
    }
    if (_showInterpretation) {
      await _printInterpretation();
      return;
    }
    await _printResultRecord();
  }

  Future<void> _printResultRecord() async {
    Uint8List? bytes = _pdfBytes;
    if (bytes == null || bytes.isEmpty) {
      await _loadPdf();
      bytes = _pdfBytes;
    }
    if (!mounted) {
      return;
    }
    if (bytes == null || bytes.isEmpty) {
      setState(() {
        _errorMessage = '暂无可打印的评估结果记录';
      });
      return;
    }
    setState(() {
      _printing = true;
      _errorMessage = '';
    });
    try {
      final Pep3RecordSummary record = _displayRecord ?? widget.record;
      await Printing.layoutPdf(
        name: _erxinPrintFileName(record, '评估结果记录'),
        onLayout: (_) async => bytes!,
      );
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _errorMessage = '评估结果记录打印失败：$error';
      });
    } finally {
      if (mounted) {
        setState(() {
          _printing = false;
        });
      }
    }
  }

  Future<void> _printInterpretation() async {
    ErxinReportInterpretation? interpretation = _interpretation;
    if (interpretation == null || interpretation.isEmpty) {
      if (!_interpretationFetched && !_interpretationLoading) {
        await _loadSavedInterpretation();
        interpretation = _interpretation;
      }
    }
    if (!mounted) {
      return;
    }
    if (interpretation == null || interpretation.isEmpty) {
      setState(() {
        _interpretationErrorMessage = '请先生成报告解读后再打印';
      });
      return;
    }
    setState(() {
      _printing = true;
      _interpretationErrorMessage = '';
    });
    try {
      final Uint8List bytes =
          await widget.client.downloadRecordReportInterpretationPdf(
        widget.token.trim(),
        widget.record.id,
      );
      final Pep3RecordSummary record = _displayRecord ?? widget.record;
      await Printing.layoutPdf(
        name: _erxinPrintFileName(record, '报告解读'),
        onLayout: (_) async => bytes,
      );
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _interpretationErrorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _interpretationErrorMessage = '报告解读打印失败：$error';
      });
    } finally {
      if (mounted) {
        setState(() {
          _printing = false;
        });
      }
    }
  }

  Future<void> _handleInterpretationGenerateTap() async {
    final bool regenerate =
        _interpretation != null && !_interpretation!.isEmpty;
    if (!regenerate) {
      await _generateInterpretation();
      return;
    }
    final bool confirmed = await _confirmRegenerateInterpretation();
    if (!mounted || !confirmed) {
      return;
    }
    await _generateInterpretation(regenerate: true);
  }

  Future<bool> _confirmRegenerateInterpretation() async {
    final bool? confirmed = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext dialogContext) {
        return PadDialogViewport(
          child: Center(
            child: _ErxinRegenerateInterpretationConfirmDialog(
              onCancel: () => Navigator.of(dialogContext).pop(false),
              onConfirm: () => Navigator.of(dialogContext).pop(true),
            ),
          ),
        );
      },
    );
    return confirmed == true;
  }

  @override
  Widget build(BuildContext context) {
    final Pep3RecordSummary record = _displayRecord ?? widget.record;
    return PopScope(
      canPop: false,
      child: Center(
        child: Material(
          color: Colors.transparent,
          child: Container(
            width: 980,
            height: 654,
            padding: const EdgeInsets.fromLTRB(22, 22, 22, 22),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(color: _ReportTheme.line),
              boxShadow: const <BoxShadow>[
                BoxShadow(
                  color: Color(0x24000000),
                  blurRadius: 34,
                  offset: Offset(0, 18),
                ),
              ],
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                _buildHeader(context, record),
                const SizedBox(height: 14),
                _buildTabBar(),
                const SizedBox(height: 12),
                Expanded(
                  child: _showInterpretation
                      ? _buildInterpretationContent()
                      : _buildContent(record),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildHeader(BuildContext context, Pep3RecordSummary record) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              const Text(
                '评估报告',
                style: TextStyle(
                  color: _ReportTheme.ink,
                  fontSize: 24,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 7),
              Text(
                '${record.assessmentName.trim().isEmpty ? '儿心量表-II发育行为评估报告' : record.assessmentName}   ${_studentName(record)} / ${_dateOnlyText(record.assessmentDate)}',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _ReportTheme.muted,
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(width: 16),
        Material(
          color: Colors.transparent,
          shape: const CircleBorder(),
          child: InkWell(
            onTap: () => Navigator.of(context).pop(),
            customBorder: const CircleBorder(),
            child: Container(
              width: 38,
              height: 38,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: const Color(0xFFFFF8F2),
                shape: BoxShape.circle,
                border: Border.all(color: _ReportTheme.lineSoft),
              ),
              child: const Icon(
                Icons.close_rounded,
                size: 22,
                color: _ReportTheme.muted,
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildTabBar() {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          _ErxinReportTabChip(
            label: '评估结果记录',
            active: !_showInterpretation,
            onTap: _selectResultTab,
          ),
          const SizedBox(width: 8),
          _ErxinReportTabChip(
            label: '报告解读',
            active: _showInterpretation,
            onTap: _selectInterpretationTab,
          ),
          const Spacer(),
          const SizedBox(width: 12),
          if (_showInterpretation)
            _ToolbarButton(
              label: _interpretationLoading
                  ? (_interpretationGenerating ? '生成中' : '读取中')
                  : (_interpretation == null || _interpretation!.isEmpty)
                      ? '生成解读'
                      : '重新生成解读',
              icon: Icons.auto_awesome_rounded,
              filled: true,
              onTap: _interpretationLoading
                  ? null
                  : () => unawaited(_handleInterpretationGenerateTap()),
            ),
          if (_showInterpretation) const SizedBox(width: 8),
          _ToolbarButton(
            label: _printing ? '打印中' : '打印',
            icon: Icons.print_rounded,
            onTap: (_printing ||
                    (!_showInterpretation && _loading) ||
                    (_showInterpretation && _interpretationLoading))
                ? null
                : () => unawaited(_printCurrentTab()),
          ),
        ],
      ),
    );
  }

  Widget _buildContent(Pep3RecordSummary record) {
    if (_loading) {
      return const _ReportPreviewLoadingState(message: '评估报告加载中...');
    }
    if (_errorMessage.isNotEmpty) {
      return _ReportPreviewErrorState(
        message: _errorMessage,
        onRetry: _retryPreview,
      );
    }
    final Uint8List? bytes = _pdfBytes;
    if (bytes == null || bytes.isEmpty) {
      return const _ReportPreviewEmptyState(message: '暂无评估报告内容');
    }
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFFDF8F3),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      clipBehavior: Clip.antiAlias,
      child: _LazyReportPdfPreview(
        key: ValueKey<String>(
          'erxin-report-pdf-${record.id}-${record.updatedTime}-${bytes.length}',
        ),
        bytes: bytes,
        pageCount: 1,
        dpi: _erxinReportPreviewRasterDpi,
        maxPageWidth: 920,
      ),
    );
  }

  Widget _buildInterpretationContent() {
    if (_interpretationLoading) {
      if (_interpretationGenerating) {
        return _ErxinInterpretationProgressState(
          message: _interpretationProgressMessage,
          streamingText: _interpretationStreamingText,
        );
      }
      return _ErxinInterpretationReadLoadingState(
        message: _interpretationProgressMessage,
      );
    }
    if (_interpretationErrorMessage.isNotEmpty) {
      return _ReportPreviewErrorState(
        message: _interpretationErrorMessage,
        onRetry: () => unawaited(_generateInterpretation(regenerate: true)),
      );
    }
    final ErxinReportInterpretation? interpretation = _interpretation;
    if (interpretation == null || interpretation.isEmpty) {
      return _ErxinInterpretationEmptyState(
        message: '报告解读尚未生成',
        detail: '点击“生成解读”后，AI 会基于当前评估结果生成并保存。',
        actionLabel: '生成解读',
        onAction: () => unawaited(_generateInterpretation()),
      );
    }
    return _ErxinInterpretationView(interpretation: interpretation);
  }
}

class _ErxinReportTabChip extends StatelessWidget {
  const _ErxinReportTabChip({
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
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Ink(
          height: 40,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: active ? const Color(0xFFF2F7FF) : Colors.white,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: active ? _ReportTheme.blue : _ReportTheme.lineSoft,
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(
                active
                    ? Icons.radio_button_checked_rounded
                    : Icons.radio_button_unchecked_rounded,
                size: 16,
                color: active ? _ReportTheme.blue : _ReportTheme.muted,
              ),
              const SizedBox(width: 7),
              Text(
                label,
                style: TextStyle(
                  color: active ? _ReportTheme.blue : _ReportTheme.text,
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

class _ErxinRegenerateInterpretationConfirmDialog extends StatelessWidget {
  const _ErxinRegenerateInterpretationConfirmDialog({
    required this.onCancel,
    required this.onConfirm,
  });

  final VoidCallback onCancel;
  final VoidCallback onConfirm;

  @override
  Widget build(BuildContext context) {
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: 430,
        padding: const EdgeInsets.fromLTRB(22, 20, 22, 18),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: _ReportTheme.line),
          boxShadow: const <BoxShadow>[
            BoxShadow(
              color: Color(0x24000000),
              blurRadius: 30,
              offset: Offset(0, 14),
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            const Text(
              '确认重新生成解读',
              style: TextStyle(
                color: _ReportTheme.ink,
                fontSize: 20,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 10),
            const Text(
              '重新生成会覆盖当前已保存的报告解读，确认继续吗？',
              style: TextStyle(
                color: _ReportTheme.text,
                fontSize: 14,
                height: 1.45,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 18),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: <Widget>[
                SizedBox(
                  width: 92,
                  child: _ToolbarButton(
                    label: '取消',
                    onTap: onCancel,
                  ),
                ),
                const SizedBox(width: 10),
                SizedBox(
                  width: 122,
                  child: _ToolbarButton(
                    label: '确认重新生成',
                    filled: true,
                    onTap: onConfirm,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _ErxinInterpretationReadLoadingState extends StatelessWidget {
  const _ErxinInterpretationReadLoadingState({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Center(
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            const SizedBox(
              width: 18,
              height: 18,
              child: CircularProgressIndicator(
                strokeWidth: 2.4,
                color: _ReportTheme.orange,
              ),
            ),
            const SizedBox(width: 10),
            Text(
              message.trim().isEmpty ? '正在读取已保存的报告解读...' : message,
              style: const TextStyle(
                color: _ReportTheme.muted,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ErxinInterpretationProgressState extends StatefulWidget {
  const _ErxinInterpretationProgressState({
    required this.message,
    required this.streamingText,
    this.domainSectionTitle = '能区表现',
  });

  final String message;
  final String streamingText;
  final String domainSectionTitle;

  @override
  State<_ErxinInterpretationProgressState> createState() =>
      _ErxinInterpretationProgressStateState();
}

class _ErxinInterpretationProgressStateState
    extends State<_ErxinInterpretationProgressState> {
  final ScrollController _scrollController = ScrollController();

  @override
  void didUpdateWidget(_ErxinInterpretationProgressState oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.streamingText != widget.streamingText) {
      _scrollToBottom();
    }
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted || !_scrollController.hasClients) {
        return;
      }
      _scrollController.animateTo(
        _scrollController.position.maxScrollExtent,
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOutCubic,
      );
    });
  }

  @override
  Widget build(BuildContext context) {
    final _ErxinInterpretationStreamingPreview preview =
        _erxinInterpretationStreamingPreview(widget.streamingText);
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(18, 16, 18, 18),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: <Widget>[
              Container(
                width: 32,
                height: 32,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: const Color(0xFFFFF1E8),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: const Color(0xFFFFD8BD)),
                ),
                child: const Icon(
                  Icons.auto_awesome_rounded,
                  color: _ReportTheme.orangeDeep,
                  size: 18,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    const Text(
                      'AI 正在生成报告解读',
                      style: TextStyle(
                        color: _ReportTheme.text,
                        fontSize: 15,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    const SizedBox(height: 3),
                    Text(
                      widget.message,
                      style: const TextStyle(
                        color: _ReportTheme.muted,
                        fontSize: 12,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ],
                ),
              ),
              const _ErxinStreamingIndicator(),
            ],
          ),
          const SizedBox(height: 12),
          Expanded(
            child: preview.isEmpty
                ? const Center(
                    child: Text(
                      '正在建立报告结构，稍后开始输出解读内容...',
                      style: TextStyle(
                        color: _ReportTheme.muted,
                        fontSize: 13,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  )
                : SingleChildScrollView(
                    controller: _scrollController,
                    padding: const EdgeInsets.only(bottom: 4),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        _ErxinInterpretationStreamingSection(
                          title: '综合解读',
                          paragraphs: <String>[preview.summary],
                        ),
                        _ErxinInterpretationStreamingSection(
                          title: widget.domainSectionTitle,
                          paragraphs: preview.domainAnalysis,
                        ),
                        _ErxinInterpretationStreamingSection(
                          title: '发展建议',
                          paragraphs: preview.suggestions,
                        ),
                        _ErxinInterpretationStreamingSection(
                          title: '注意事项',
                          paragraphs: preview.notes,
                        ),
                      ],
                    ),
                  ),
          ),
        ],
      ),
    );
  }
}

class _ErxinStreamingIndicator extends StatefulWidget {
  const _ErxinStreamingIndicator();

  @override
  State<_ErxinStreamingIndicator> createState() =>
      _ErxinStreamingIndicatorState();
}

class _ErxinStreamingIndicatorState extends State<_ErxinStreamingIndicator>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (BuildContext context, Widget? child) {
        return Opacity(
          opacity: .45 + .45 * math.sin(_controller.value * math.pi),
          child: child,
        );
      },
      child: Container(
        height: 24,
        padding: const EdgeInsets.symmetric(horizontal: 9),
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: const Color(0xFFFFF1E8),
          borderRadius: BorderRadius.circular(999),
          border: Border.all(color: const Color(0xFFFFD8BD)),
        ),
        child: const Text(
          '流式生成',
          style: TextStyle(
            color: _ReportTheme.orangeDeep,
            fontSize: 11,
            fontWeight: FontWeight.w900,
          ),
        ),
      ),
    );
  }
}

class _ErxinInterpretationStreamingSection extends StatelessWidget {
  const _ErxinInterpretationStreamingSection({
    required this.title,
    required this.paragraphs,
  });

  final String title;
  final List<String> paragraphs;

  @override
  Widget build(BuildContext context) {
    final List<String> items = paragraphs
        .map((String item) => item.trim())
        .where((String item) => item.isNotEmpty)
        .toList();
    if (items.isEmpty) {
      return const SizedBox.shrink();
    }
    return _ErxinInterpretationSection(
      title: title,
      paragraphs: items,
    );
  }
}

class _ErxinInterpretationStreamingPreview {
  const _ErxinInterpretationStreamingPreview({
    required this.summary,
    required this.domainAnalysis,
    required this.suggestions,
    required this.notes,
  });

  final String summary;
  final List<String> domainAnalysis;
  final List<String> suggestions;
  final List<String> notes;

  bool get isEmpty =>
      summary.trim().isEmpty &&
      domainAnalysis.isEmpty &&
      suggestions.isEmpty &&
      notes.isEmpty;
}

_ErxinInterpretationStreamingPreview _erxinInterpretationStreamingPreview(
  String raw,
) {
  final String text = raw.trim();
  if (text.isEmpty) {
    return const _ErxinInterpretationStreamingPreview(
      summary: '',
      domainAnalysis: <String>[],
      suggestions: <String>[],
      notes: <String>[],
    );
  }
  try {
    final Object? decoded = jsonDecode(text);
    if (decoded is Map) {
      return _erxinInterpretationPreviewFromMap(
        Map<String, dynamic>.from(decoded),
      );
    }
  } on Object {
    // Partial JSON while streaming. Fall through to lightweight extraction.
  }
  return _ErxinInterpretationStreamingPreview(
    summary: _extractPartialJsonStringValue(text, 'summary') ?? '',
    domainAnalysis: _extractPartialJsonArrayValues(text, 'domainAnalysis'),
    suggestions: _extractPartialJsonArrayValues(text, 'suggestions'),
    notes: _extractPartialJsonArrayValues(text, 'notes'),
  );
}

_ErxinInterpretationStreamingPreview _erxinInterpretationPreviewFromMap(
  Map<String, dynamic> json,
) {
  return _ErxinInterpretationStreamingPreview(
    summary: '${json['summary'] ?? ''}'.trim(),
    domainAnalysis: _previewStringListFrom(json['domainAnalysis']),
    suggestions: _previewStringListFrom(json['suggestions']),
    notes: _previewStringListFrom(json['notes']),
  );
}

List<String> _previewStringListFrom(Object? value) {
  if (value is List) {
    return value
        .map((Object? item) => '${item ?? ''}'.trim())
        .where((String item) => item.isNotEmpty)
        .toList();
  }
  return <String>[];
}

String? _extractPartialJsonStringValue(String text, String key) {
  final RegExpMatch? match =
      RegExp('"$key"\\s*:\\s*"((?:\\\\.|[^"\\\\])*)').firstMatch(text);
  if (match == null) {
    return null;
  }
  return _decodePartialJsonString(match.group(1) ?? '');
}

List<String> _extractPartialJsonArrayValues(String text, String key) {
  final int keyIndex = text.indexOf('"$key"');
  if (keyIndex < 0) {
    return const <String>[];
  }
  final int arrayStart = text.indexOf('[', keyIndex);
  if (arrayStart < 0) {
    return const <String>[];
  }
  final int arrayEnd = text.indexOf(']', arrayStart);
  final String body = text.substring(
    arrayStart + 1,
    arrayEnd >= 0 ? arrayEnd : text.length,
  );
  final List<String> values = RegExp('"((?:\\\\.|[^"\\\\])*)"')
      .allMatches(body)
      .map(
          (RegExpMatch match) => _decodePartialJsonString(match.group(1) ?? ''))
      .where((String item) => item.trim().isNotEmpty)
      .toList();
  final String? trailing = _extractTrailingPartialJsonArrayString(body);
  if (trailing != null && trailing.trim().isNotEmpty) {
    values.add(trailing.trim());
  }
  return values;
}

String? _extractTrailingPartialJsonArrayString(String body) {
  int quoteIndex = -1;
  bool escaped = false;
  bool inString = false;
  for (int index = 0; index < body.length; index++) {
    final String char = body[index];
    if (escaped) {
      escaped = false;
      continue;
    }
    if (char == r'\') {
      escaped = true;
      continue;
    }
    if (char == '"') {
      inString = !inString;
      quoteIndex = index;
    }
  }
  if (!inString || quoteIndex < 0 || quoteIndex >= body.length - 1) {
    return null;
  }
  return _decodePartialJsonString(body.substring(quoteIndex + 1));
}

String _decodePartialJsonString(String value) {
  try {
    return jsonDecode('"$value"') as String;
  } on Object {
    return value
        .replaceAll(r'\"', '"')
        .replaceAll(r'\n', '\n')
        .replaceAll(r'\\', r'\');
  }
}

class _ErxinInterpretationEmptyState extends StatelessWidget {
  const _ErxinInterpretationEmptyState({
    required this.message,
    required this.detail,
    required this.actionLabel,
    required this.onAction,
  });

  final String message;
  final String detail;
  final String actionLabel;
  final VoidCallback onAction;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            const Icon(
              Icons.auto_awesome_outlined,
              size: 32,
              color: _ReportTheme.orangeDeep,
            ),
            const SizedBox(height: 14),
            Text(
              message,
              style: const TextStyle(
                color: _ReportTheme.text,
                fontSize: 16,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              detail,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _ReportTheme.muted,
                fontSize: 12,
                fontWeight: FontWeight.w800,
              ),
            ),
            const SizedBox(height: 16),
            SizedBox(
              width: 236,
              child: _ToolbarButton(
                key: const ValueKey<String>(
                  'erxin-interpretation-empty-action',
                ),
                label: actionLabel,
                filled: true,
                icon: Icons.auto_awesome_rounded,
                onTap: onAction,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ErxinInterpretationView extends StatelessWidget {
  const _ErxinInterpretationView({
    required this.interpretation,
    this.domainSectionTitle = '能区表现',
  });

  final ErxinReportInterpretation interpretation;
  final String domainSectionTitle;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      clipBehavior: Clip.antiAlias,
      child: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(18, 16, 18, 18),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                const Icon(
                  Icons.auto_awesome_rounded,
                  size: 20,
                  color: _ReportTheme.orangeDeep,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    interpretation.title.trim().isEmpty
                        ? '报告解读'
                        : interpretation.title.trim(),
                    style: const TextStyle(
                      color: _ReportTheme.ink,
                      fontSize: 20,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                _ErxinInterpretationSourceBadge(
                  generatedBy: interpretation.generatedBy,
                ),
              ],
            ),
            const SizedBox(height: 14),
            _ErxinInterpretationSection(
              title: '综合解读',
              paragraphs: <String>[interpretation.summary],
            ),
            _ErxinInterpretationSection(
              title: domainSectionTitle,
              paragraphs: interpretation.domainAnalysis,
              numbered: true,
            ),
            _ErxinInterpretationSection(
              title: '发展建议',
              paragraphs: interpretation.suggestions,
              numbered: true,
            ),
            if (interpretation.notes.isNotEmpty)
              _ErxinInterpretationSection(
                title: '注意事项',
                paragraphs: interpretation.notes,
              ),
          ],
        ),
      ),
    );
  }
}

class _ErxinInterpretationSourceBadge extends StatelessWidget {
  const _ErxinInterpretationSourceBadge({required this.generatedBy});

  final String generatedBy;

  @override
  Widget build(BuildContext context) {
    final bool ai = generatedBy.trim().toLowerCase() == 'ai';
    return Container(
      height: 28,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: ai ? const Color(0xFFF2F7FF) : const Color(0xFFFFF1E8),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(
          color: ai ? const Color(0xFFD8E6FF) : const Color(0xFFFFD8BD),
        ),
      ),
      child: Text(
        ai ? 'AI生成' : '规则解读',
        style: TextStyle(
          color: ai ? _ReportTheme.blue : _ReportTheme.orangeDeep,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _ErxinInterpretationSection extends StatelessWidget {
  const _ErxinInterpretationSection({
    required this.title,
    required this.paragraphs,
    this.numbered = false,
  });

  final String title;
  final List<String> paragraphs;
  final bool numbered;

  @override
  Widget build(BuildContext context) {
    final List<String> items = paragraphs
        .map((String item) => item.trim())
        .where((String item) => item.isNotEmpty)
        .toList();
    if (items.isEmpty) {
      return const SizedBox.shrink();
    }
    return Padding(
      padding: const EdgeInsets.only(bottom: 14),
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: _ReportTheme.lineSoft),
        ),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
          child: SizedBox(
            width: double.infinity,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  title,
                  style: const TextStyle(
                    color: _ReportTheme.orangeDeep,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 8),
                for (int index = 0; index < items.length; index++) ...<Widget>[
                  Text(
                    numbered ? '${index + 1}. ${items[index]}' : items[index],
                    style: const TextStyle(
                      color: _ReportTheme.text,
                      fontSize: 14,
                      height: 1.55,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  if (index != items.length - 1) const SizedBox(height: 6),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}

// ignore: unused_element
Future<Uint8List> _buildErxinInterpretationPrintPdf(
  Pep3RecordSummary record,
  ErxinReportInterpretation interpretation,
  PdfPageFormat format, {
  String title = '0岁～6岁儿童发育行为评估量表（儿心量表-II）报告解读',
  String domainSectionTitle = '能区表现',
}) async {
  final pw.Font baseFont =
      await fontFromAssetBundle('assets/fonts/NotoSansSC-Regular.ttf');
  final PdfPageFormat pageFormat = format.copyWith(
    marginLeft: 28,
    marginTop: 28,
    marginRight: 28,
    marginBottom: 28,
  );
  final pw.Document document = pw.Document(
    theme: pw.ThemeData.withFont(
      base: baseFont,
      bold: baseFont,
      fontFallback: <pw.Font>[baseFont],
    ),
  );
  const PdfColor ink = PdfColor.fromInt(0xff172033);
  const PdfColor orange = PdfColor.fromInt(0xfff57c00);
  const PdfColor orangeSoft = PdfColor.fromInt(0xfffffcf8);
  const PdfColor border = PdfColor.fromInt(0xffead8c8);

  document.addPage(
    pw.MultiPage(
      pageFormat: pageFormat,
      build: (pw.Context context) => <pw.Widget>[
        pw.Container(
          width: double.infinity,
          alignment: pw.Alignment.center,
          child: pw.Text(
            title,
            textAlign: pw.TextAlign.center,
            style: pw.TextStyle(
              color: ink,
              fontSize: 16,
              fontWeight: pw.FontWeight.bold,
            ),
          ),
        ),
        pw.SizedBox(height: 20),
        _erxinInterpretationPrintSection(
          title: '综合解读',
          paragraphs: <String>[interpretation.summary],
          orange: orange,
          orangeSoft: orangeSoft,
          border: border,
          ink: ink,
        ),
        _erxinInterpretationPrintSection(
          title: domainSectionTitle,
          paragraphs: interpretation.domainAnalysis,
          numbered: true,
          orange: orange,
          orangeSoft: orangeSoft,
          border: border,
          ink: ink,
        ),
        _erxinInterpretationPrintSection(
          title: '发展建议',
          paragraphs: interpretation.suggestions,
          numbered: true,
          orange: orange,
          orangeSoft: orangeSoft,
          border: border,
          ink: ink,
        ),
        _erxinInterpretationPrintSection(
          title: '注意事项',
          paragraphs: interpretation.notes,
          orange: orange,
          orangeSoft: orangeSoft,
          border: border,
          ink: ink,
        ),
      ],
    ),
  );
  return document.save();
}

pw.Widget _erxinInterpretationPrintSection({
  required String title,
  required List<String> paragraphs,
  required PdfColor orange,
  required PdfColor orangeSoft,
  required PdfColor border,
  required PdfColor ink,
  bool numbered = false,
}) {
  final List<String> items = paragraphs
      .map((String item) => item.trim())
      .where((String item) => item.isNotEmpty)
      .toList();
  if (items.isEmpty) {
    return pw.SizedBox();
  }
  return pw.Container(
    width: double.infinity,
    margin: const pw.EdgeInsets.only(bottom: 14),
    padding: const pw.EdgeInsets.fromLTRB(10, 8, 10, 8),
    decoration: pw.BoxDecoration(
      color: orangeSoft,
      borderRadius: pw.BorderRadius.circular(12),
      border: pw.Border.all(color: border),
    ),
    child: pw.Column(
      crossAxisAlignment: pw.CrossAxisAlignment.start,
      children: <pw.Widget>[
        pw.Text(
          title,
          style: pw.TextStyle(
            color: orange,
            fontSize: 11.5,
            fontWeight: pw.FontWeight.bold,
          ),
        ),
        pw.SizedBox(height: 4),
        for (int index = 0; index < items.length; index++) ...<pw.Widget>[
          if (numbered)
            pw.Row(
              crossAxisAlignment: pw.CrossAxisAlignment.start,
              children: <pw.Widget>[
                pw.SizedBox(
                  width: 15,
                  child: pw.Text(
                    '${index + 1}.',
                    style: pw.TextStyle(
                      color: ink,
                      fontSize: 9.8,
                      fontWeight: pw.FontWeight.bold,
                    ),
                  ),
                ),
                pw.Expanded(
                  child: pw.Text(
                    _erxinStripLeadingNumber(items[index]),
                    style: pw.TextStyle(
                      color: ink,
                      fontSize: 9.8,
                      lineSpacing: 2,
                    ),
                  ),
                ),
              ],
            )
          else
            pw.Text(
              items[index],
              style: pw.TextStyle(
                color: ink,
                fontSize: 9.8,
                lineSpacing: 2,
              ),
            ),
          if (index != items.length - 1) pw.SizedBox(height: 2),
        ],
      ],
    ),
  );
}

String _erxinStripLeadingNumber(String value) {
  return value
      .trim()
      .replaceFirst(RegExp(r'^(?:\d+|[一二三四五六七八九十]+)[\.．、:：]?\s*'), '')
      .trim();
}

String _erxinPrintFileName(Pep3RecordSummary record, String suffix) {
  final String name = _studentName(record).trim().isEmpty
      ? '未命名儿童'
      : _studentName(record).trim();
  final String date = _dateOnlyText(record.assessmentDate).replaceAll('-', '');
  return '$name-儿心量表-$suffix${date.isEmpty ? '' : '-$date'}.pdf';
}

String _pep3PrintFileName(Pep3RecordSummary record, String suffix) {
  final String name = _studentName(record).trim().isEmpty
      ? '未命名儿童'
      : _studentName(record).trim();
  final String date = _dateOnlyText(record.assessmentDate).replaceAll('-', '');
  return '$name-PEP3-$suffix${date.isEmpty ? '' : '-$date'}.pdf';
}

class _LazyReportPdfPreview extends StatefulWidget {
  const _LazyReportPdfPreview({
    required this.bytes,
    required this.pageCount,
    this.dpi = _reportPreviewRasterDpi,
    this.maxPageWidth = 860,
    super.key,
  });

  final Uint8List bytes;
  final int pageCount;
  final double dpi;
  final double maxPageWidth;

  @override
  State<_LazyReportPdfPreview> createState() => _LazyReportPdfPreviewState();
}

class _LazyReportPdfPreviewState extends State<_LazyReportPdfPreview> {
  static const int _pageCacheLimit = 6;

  final Map<int, Future<_ReportPdfPageSnapshot>> _pendingPageFutures =
      <int, Future<_ReportPdfPageSnapshot>>{};
  final LinkedHashMap<int, _ReportPdfPageSnapshot> _pageCache =
      LinkedHashMap<int, _ReportPdfPageSnapshot>();

  @override
  void initState() {
    super.initState();
    _warmAround(0);
  }

  Future<_ReportPdfPageSnapshot> _loadPage(int pageIndex) {
    final _ReportPdfPageSnapshot? cached = _pageCache.remove(pageIndex);
    if (cached != null) {
      _pageCache[pageIndex] = cached;
      return Future<_ReportPdfPageSnapshot>.value(cached);
    }
    final Future<_ReportPdfPageSnapshot>? pending =
        _pendingPageFutures[pageIndex];
    if (pending != null) {
      return pending;
    }
    final Future<_ReportPdfPageSnapshot> future = (() async {
      final PdfRaster raster = await Printing.raster(
        widget.bytes,
        pages: <int>[pageIndex],
        dpi: widget.dpi,
      ).first;
      final _ReportPdfPageSnapshot snapshot = _ReportPdfPageSnapshot(
        raster: raster,
        width: raster.width,
        height: raster.height,
      );
      _pageCache[pageIndex] = snapshot;
      while (_pageCache.length > _pageCacheLimit) {
        _pageCache.remove(_pageCache.keys.first);
      }
      return snapshot;
    })();
    _pendingPageFutures[pageIndex] = future;
    future.whenComplete(() {
      if (identical(_pendingPageFutures[pageIndex], future)) {
        _pendingPageFutures.remove(pageIndex);
      }
    });
    return future;
  }

  void _warmAround(int pageIndex) {
    if (pageIndex < 0 || pageIndex >= widget.pageCount) {
      return;
    }
    _loadPage(pageIndex);
    if (pageIndex + 1 < widget.pageCount) {
      _loadPage(pageIndex + 1);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (widget.pageCount == 1) {
      _warmAround(0);
      return SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(12, 12, 12, 12),
        child: _LazyReportPdfPageCard(
          pageIndex: 0,
          pageCount: 1,
          pageFuture: _loadPage(0),
          maxPageWidth: widget.maxPageWidth,
          badgeText: '第 1 / 1 页',
          compact: true,
        ),
      );
    }
    return ListView.builder(
      padding: const EdgeInsets.fromLTRB(16, 14, 16, 16),
      cacheExtent: 720,
      itemCount: widget.pageCount,
      itemBuilder: (BuildContext context, int index) {
        _warmAround(index);
        return _LazyReportPdfPageCard(
          pageIndex: index,
          pageCount: widget.pageCount,
          pageFuture: _loadPage(index),
          maxPageWidth: widget.maxPageWidth,
          badgeText: '第 ${index + 1} / ${widget.pageCount} 页',
        );
      },
    );
  }
}

class _LazyReportPdfPageCard extends StatelessWidget {
  const _LazyReportPdfPageCard({
    required this.pageIndex,
    required this.pageCount,
    required this.pageFuture,
    required this.maxPageWidth,
    required this.badgeText,
    this.compact = false,
  });

  final int pageIndex;
  final int pageCount;
  final Future<_ReportPdfPageSnapshot> pageFuture;
  final double maxPageWidth;
  final String badgeText;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(
          bottom: compact || pageIndex == pageCount - 1 ? 0 : 16),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final double cardWidth = math.min(constraints.maxWidth, maxPageWidth);
          return Center(
            child: SizedBox(
              width: cardWidth,
              child: DecoratedBox(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(compact ? 10 : 12),
                  boxShadow: compact
                      ? const <BoxShadow>[]
                      : const <BoxShadow>[
                          BoxShadow(
                            color: Color(0x11000000),
                            blurRadius: 12,
                            offset: Offset(0, 6),
                          ),
                        ],
                  border:
                      compact ? Border.all(color: _ReportTheme.lineSoft) : null,
                ),
                child: Padding(
                  padding: EdgeInsets.all(compact ? 8 : 12),
                  child: Stack(
                    children: <Widget>[
                      FutureBuilder<_ReportPdfPageSnapshot>(
                        future: pageFuture,
                        builder: (
                          BuildContext context,
                          AsyncSnapshot<_ReportPdfPageSnapshot> snapshot,
                        ) {
                          if (snapshot.hasError) {
                            return _ReportPdfPagePlaceholder(
                              child: Text(
                                '第 ${pageIndex + 1} 页渲染失败：${snapshot.error}',
                                textAlign: TextAlign.center,
                                style: const TextStyle(
                                  color: _ReportTheme.text,
                                  fontSize: 13,
                                  height: 1.4,
                                  fontWeight: FontWeight.w800,
                                ),
                              ),
                            );
                          }
                          if (!snapshot.hasData) {
                            return const _ReportPdfPagePlaceholder(
                              child: _ReportPreviewLoadingState(
                                message: '评估报告渲染中...',
                              ),
                            );
                          }
                          final _ReportPdfPageSnapshot data = snapshot.data!;
                          return AspectRatio(
                            aspectRatio: data.aspectRatio,
                            child: Image(
                              image: PdfRasterImage(data.raster),
                              fit: BoxFit.contain,
                              filterQuality: FilterQuality.medium,
                              gaplessPlayback: true,
                            ),
                          );
                        },
                      ),
                      Positioned(
                        top: 10,
                        right: 10,
                        child: IgnorePointer(
                          child: _ReportPdfPageBadge(
                            text: badgeText,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}

class _ReportPdfPagePlaceholder extends StatelessWidget {
  const _ReportPdfPagePlaceholder({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return AspectRatio(
      aspectRatio: 0.76,
      child: DecoratedBox(
        decoration: BoxDecoration(
          color: const Color(0xFFFFFBF7),
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: _ReportTheme.lineSoft),
        ),
        child: Center(child: child),
      ),
    );
  }
}

class _ReportPdfPageBadge extends StatelessWidget {
  const _ReportPdfPageBadge({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        color: const Color(0xEFFFFFFB),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: const Color(0xFFE9D8CA)),
        boxShadow: const <BoxShadow>[
          BoxShadow(
            color: Color(0x12000000),
            blurRadius: 8,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
        child: Text(
          text,
          style: const TextStyle(
            color: _ReportTheme.text,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
      ),
    );
  }
}

class _ReportModuleChip extends StatelessWidget {
  const _ReportModuleChip({
    required this.option,
    required this.active,
    required this.onTap,
  });

  final _ReportModuleOption option;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Ink(
          height: 40,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: active ? const Color(0xFFF2F7FF) : Colors.white,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: active ? _ReportTheme.blue : _ReportTheme.lineSoft,
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Container(
                width: 8,
                height: 8,
                decoration: BoxDecoration(
                  color: active ? _ReportTheme.blue : const Color(0xFFD5DDE6),
                  shape: BoxShape.circle,
                ),
              ),
              const SizedBox(width: 8),
              Text(
                option.label,
                style: TextStyle(
                  color: active ? _ReportTheme.blue : _ReportTheme.text,
                  fontSize: 13,
                  fontWeight: FontWeight.w900,
                ),
              ),
              if (option.recommended) ...<Widget>[
                const SizedBox(width: 8),
                Text(
                  '推荐',
                  style: TextStyle(
                    color: active ? _ReportTheme.blue : _ReportTheme.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _ReportPreviewLoadingState extends StatelessWidget {
  const _ReportPreviewLoadingState({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          const SizedBox(
            width: 26,
            height: 26,
            child: CircularProgressIndicator(
              strokeWidth: 3,
              color: _ReportTheme.orange,
            ),
          ),
          const SizedBox(height: 14),
          Text(
            message,
            style: const TextStyle(
              color: _ReportTheme.text,
              fontSize: 14,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _ReportPreviewErrorState extends StatelessWidget {
  const _ReportPreviewErrorState({
    required this.message,
    required this.onRetry,
  });

  final String message;
  final VoidCallback onRetry;

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
              color: _ReportTheme.text,
              fontSize: 14,
              height: 1.4,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 14),
          _ToolbarButton(
            label: '重新加载',
            filled: true,
            onTap: onRetry,
          ),
        ],
      ),
    );
  }
}

class _ReportPreviewEmptyState extends StatelessWidget {
  const _ReportPreviewEmptyState({required this.message});

  final String message;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Text(
        message,
        style: const TextStyle(
          color: _ReportTheme.muted,
          fontSize: 14,
          fontWeight: FontWeight.w800,
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

DateTime? _configDateValue(String raw) {
  final DateTime? parsed = _parseDateTime(raw);
  return parsed == null ? null : _dateOnly(parsed.toLocal());
}

DateTime? _originalAssessmentDateValue(Pep3RecordSummary record) {
  final DateTime? created = _configDateValue(record.createdTime);
  if (created != null) {
    return created;
  }
  return _configDateValue(record.assessmentDate);
}

String? _originalAssessmentDateText(Pep3RecordSummary record) {
  final DateTime? value = _originalAssessmentDateValue(record);
  if (value == null) {
    return null;
  }
  return _dateText(value);
}

List<String> _splitExaminerNames(String raw) {
  return raw
      .split(RegExp(r'[、,，]'))
      .map((String item) => item.trim())
      .where((String item) => item.isNotEmpty)
      .toList();
}

List<String> _uniqueExaminerNames(Iterable<String> names) {
  final Set<String> seen = <String>{};
  final List<String> result = <String>[];
  for (final String item in names) {
    final String trimmed = item.trim();
    if (trimmed.isEmpty || seen.contains(trimmed)) {
      continue;
    }
    seen.add(trimmed);
    result.add(trimmed);
  }
  return result;
}

String _joinExaminerNames(Iterable<String> names) {
  return _uniqueExaminerNames(names).join('、');
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
