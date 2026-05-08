import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_scale_client.dart';
import 'erxin_assessment_client.dart';

class ErxinAssessmentPage extends StatefulWidget {
  const ErxinAssessmentPage({
    required this.onBack,
    this.args = const ErxinAssessmentLaunchArgs(),
    this.client = const ApiErxinAssessmentClient(),
    super.key,
  });

  final VoidCallback onBack;
  final ErxinAssessmentLaunchArgs args;
  final ErxinAssessmentClient client;

  @override
  State<ErxinAssessmentPage> createState() => _ErxinAssessmentPageState();
}

class _ErxinAssessmentPageState extends State<ErxinAssessmentPage> {
  static const String _authTokenStorageKey = 'auth_token';
  static const List<int> _standardAgeMonths = <int>[
    1,
    2,
    3,
    4,
    5,
    6,
    7,
    8,
    9,
    10,
    11,
    12,
    15,
    18,
    21,
    24,
    27,
    30,
    33,
    36,
    42,
    48,
    54,
    60,
    66,
    72,
    78,
    84,
  ];

  final Map<int, bool> _itemPasses = <int, bool>{};
  final Map<int, ErxinAssessmentItem> _itemDetailCache =
      <int, ErxinAssessmentItem>{};
  final Map<int, Future<ErxinAssessmentItem>> _itemDetailFetches =
      <int, Future<ErxinAssessmentItem>>{};

  ErxinTemplateSummary _template = ErxinTemplateSummary.empty;
  String _token = '';
  String _selectedDomainCode = '';
  int _selectedItemNo = 0;
  bool _loading = true;
  bool _showFutureMonths = false;
  String _errorMessage = '';
  String _autoSaveText = '等待作答';

  @override
  void initState() {
    super.initState();
    _initialize();
  }

  Future<void> _initialize() async {
    setState(() {
      _loading = true;
      _errorMessage = '';
    });
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_authTokenStorageKey) ?? '';
    if (token.trim().isEmpty) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = '请先登录后再进行测评';
      });
      return;
    }
    try {
      final ErxinTemplateSummary template =
          await widget.client.fetchTemplateSummary(token);
      if (!mounted) {
        return;
      }
      final String firstDomain = template.domains.isNotEmpty
          ? template.domains.first.domainCode
          : 'GM';
      setState(() {
        _token = token;
        _template = template;
        _selectedDomainCode = firstDomain;
        _selectedItemNo = _firstVisibleItemNo(firstDomain);
        _loading = false;
        _autoSaveText = '已准备';
      });
      _prefetchSelectedItem();
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = '儿心量表加载失败：$error';
      });
    }
  }

  int get _mainAgeMonth {
    final double months = _actualAgeMonths(
      widget.args.birthDate,
      widget.args.assessmentDate,
    );
    if (months <= 0) {
      return 0;
    }
    int selected = _standardAgeMonths.first;
    double bestDistance = (months - selected).abs();
    for (final int ageMonth in _standardAgeMonths.skip(1)) {
      final double distance = (months - ageMonth).abs();
      if (distance < bestDistance) {
        selected = ageMonth;
        bestDistance = distance;
      }
    }
    return selected;
  }

  List<int> get _initialVisibleMonths {
    final int mainAge = _mainAgeMonth;
    final int index = _standardAgeMonths.indexOf(mainAge);
    if (index < 0) {
      return <int>[];
    }
    final int start = math.max(0, index - 2);
    return _standardAgeMonths.sublist(start, index + 1);
  }

  List<int> get _futureMonths {
    final int mainAge = _mainAgeMonth;
    final int index = _standardAgeMonths.indexOf(mainAge);
    if (index < 0) {
      return <int>[];
    }
    final int end = math.min(_standardAgeMonths.length, index + 3);
    if (index + 1 >= end) {
      return <int>[];
    }
    return _standardAgeMonths.sublist(index + 1, end);
  }

  List<int> get _visibleMonths {
    return <int>[
      ..._initialVisibleMonths,
      if (_showFutureMonths) ..._futureMonths,
    ];
  }

  bool get _previousMonthsReady {
    final List<int> months = _initialVisibleMonths;
    if (months.length < 3) {
      return false;
    }
    final List<int> previous = months.take(months.length - 1).toList();
    return previous.every(
      (int month) =>
          _ageMonthComplete(_selectedDomainCode, month) &&
          _ageMonthAllPassed(_selectedDomainCode, month),
    );
  }

  bool get _mainMonthComplete {
    final int mainAge = _mainAgeMonth;
    return mainAge > 0 && _ageMonthComplete(_selectedDomainCode, mainAge);
  }

  bool get _canEnterFutureMonths {
    return !_showFutureMonths && _previousMonthsReady && _mainMonthComplete;
  }

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_errorMessage.trim().isNotEmpty) {
      return _ErrorView(message: _errorMessage, onBack: widget.onBack);
    }
    return Container(
      color: _ErxinColors.page,
      child: Column(
        children: <Widget>[
          _Header(
            args: widget.args,
            mainAgeMonth: _mainAgeMonth,
            autoSaveText: _autoSaveText,
            onBack: widget.onBack,
          ),
          Expanded(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: <Widget>[
                _DomainSidebar(
                  domains: _template.domains,
                  selectedCode: _selectedDomainCode,
                  progressForDomain: _domainProgress,
                  onSelect: _selectDomain,
                ),
                Expanded(
                  child: Column(
                    children: <Widget>[
                      Expanded(child: _buildWorkspace()),
                      _buildDetailPanel(),
                    ],
                  ),
                ),
                _buildRulePanel(),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildWorkspace() {
    return Container(
      padding: const EdgeInsets.fromLTRB(22, 18, 18, 12),
      color: Colors.white,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Text(
                '${_domainName(_selectedDomainCode)} · 首批测查',
                style: const TextStyle(
                  color: _ErxinColors.ink,
                  fontSize: 22,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(width: 14),
              _SmallBadge(text: '全部展开'),
            ],
          ),
          const SizedBox(height: 12),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
            decoration: BoxDecoration(
              color: const Color(0xFFEFF6FF),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: const Color(0xFFCFE2FF)),
            ),
            child: Text(
              '首批仅包含：${_initialVisibleMonths.join('月、')}月（主测）。'
              '前两个标准月龄均全通过后，系统才会显示往后测查。',
              style: const TextStyle(
                color: _ErxinColors.blue,
                fontSize: 14,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: <Widget>[
              for (final int month in _visibleMonths) _ageChip(month),
            ],
          ),
          const SizedBox(height: 14),
          Expanded(
            child: ListView(
              padding: EdgeInsets.zero,
              children: <Widget>[
                for (final int month in _visibleMonths)
                  _AgeMonthSection(
                    month: month,
                    titleSuffix: _ageMonthSuffix(month),
                    items: _itemsFor(_selectedDomainCode, month),
                    itemPasses: _itemPasses,
                    selectedItemNo: _selectedItemNo,
                    onSelectItem: _selectItem,
                    onScore: _scoreItem,
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _ageChip(int month) {
    final bool isMain = month == _mainAgeMonth;
    final bool complete = _ageMonthComplete(_selectedDomainCode, month);
    return Container(
      height: 36,
      padding: const EdgeInsets.symmetric(horizontal: 13),
      decoration: BoxDecoration(
        color: isMain
            ? const Color(0xFFEAF2FF)
            : complete
                ? const Color(0xFFEAF7EF)
                : const Color(0xFFF7F9FC),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(
          color: isMain
              ? _ErxinColors.blue
              : complete
                  ? _ErxinColors.green
                  : _ErxinColors.line,
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          if (complete && !isMain) ...const <Widget>[
            Icon(Icons.check_circle, size: 15, color: _ErxinColors.green),
            SizedBox(width: 5),
          ],
          Text(
            '$month月 ${_ageMonthSuffix(month)}',
            style: TextStyle(
              color: isMain ? _ErxinColors.blue : _ErxinColors.ink,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRulePanel() {
    final String nextText = _nextActionText();
    return Container(
      width: 316,
      padding: const EdgeInsets.fromLTRB(18, 18, 18, 14),
      decoration: const BoxDecoration(
        color: Color(0xFFFAFBFD),
        border: Border(left: BorderSide(color: _ErxinColors.line)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '规则判断',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 22,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 18),
          _RuleCard(
            title: '下一步',
            body: nextText,
            icon: Icons.arrow_forward_rounded,
            color: _ErxinColors.blue,
          ),
          const SizedBox(height: 18),
          _RuleChecklist(
            rows: <_RuleRow>[
              _RuleRow(
                label: '主测月龄$_mainAgeMonth月',
                value: _mainMonthComplete ? '已完成' : '未完成',
                done: _mainMonthComplete,
              ),
              for (final int month in _initialVisibleMonths
                  .where((int value) => value != _mainAgeMonth))
                _RuleRow(
                  label: '往前$month月',
                  value: _ageMonthAllPassed(_selectedDomainCode, month)
                      ? '全通过'
                      : _ageMonthComplete(_selectedDomainCode, month)
                          ? '未全通过'
                          : '未完成',
                  done: _ageMonthAllPassed(_selectedDomainCode, month),
                ),
            ],
          ),
          const SizedBox(height: 18),
          const Text(
            '是否进入往后测查',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 16,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 8),
          const Text(
            '前两个标准月龄连续全通过后，系统将显示下一批往后测查题目。',
            style: TextStyle(
              color: _ErxinColors.muted,
              fontSize: 13,
              height: 1.45,
            ),
          ),
          const SizedBox(height: 12),
          SizedBox(
            width: double.infinity,
            height: 44,
            child: FilledButton(
              onPressed: _canEnterFutureMonths
                  ? () => setState(() => _showFutureMonths = true)
                  : null,
              style: FilledButton.styleFrom(
                backgroundColor: _ErxinColors.blue,
                disabledBackgroundColor: const Color(0xFFE1E5EA),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
              ),
              child: const Text(
                '进入往后测查',
                style: TextStyle(fontWeight: FontWeight.w900),
              ),
            ),
          ),
          const Spacer(),
          _ProgressSummary(
            domainStatus:
                _domainProgress(_selectedDomainCode).answered >=
                        _domainProgress(_selectedDomainCode).total
                    ? '本能区：首批完成'
                    : '本能区：测查中',
            scaleStatus: '${_completedDomainCount()}/5 能区完成',
          ),
        ],
      ),
    );
  }

  Widget _buildDetailPanel() {
    final ErxinItemSummary? summary = _summaryByNo(_selectedItemNo);
    final String fallbackTitle = summary == null
        ? '当前题目说明'
        : '${summary.itemNo} ${summary.itemTitle}';
    return Container(
      height: 172,
      padding: const EdgeInsets.fromLTRB(22, 12, 18, 14),
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(top: BorderSide(color: _ErxinColors.line)),
      ),
      child: FutureBuilder<ErxinAssessmentItem>(
        future: _selectedItemNo > 0 ? _detailFuture(_selectedItemNo) : null,
        builder: (BuildContext context, AsyncSnapshot<ErxinAssessmentItem> snap) {
          final ErxinAssessmentItem? item = snap.data;
          final String title = item == null || item.itemNo <= 0
              ? fallbackTitle
              : '${item.itemNo} ${item.itemTitle}'
                  '${item.parentReportAllowed ? '（R）' : ''}';
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      '当前题目说明：$title',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _ErxinColors.ink,
                        fontSize: 16,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  OutlinedButton.icon(
                    onPressed: () => _showFullItemDetail(item, summary),
                    icon: const Icon(Icons.open_in_full_rounded, size: 16),
                    label: const Text('展开完整说明'),
                  ),
                ],
              ),
              const SizedBox(height: 10),
              Expanded(
                child: Row(
                  children: <Widget>[
                    Expanded(
                      child: _DetailTextBox(
                        title: '操作方法',
                        text: item?.method ?? '正在加载题目操作方法...',
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: _DetailTextBox(
                        title: '通过标准',
                        text: item?.passCriteria ?? '正在加载通过标准...',
                      ),
                    ),
                    const SizedBox(width: 12),
                    const SizedBox(
                      width: 220,
                      child: _RemarkBox(),
                    ),
                  ],
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Future<ErxinAssessmentItem> _detailFuture(int itemNo) {
    final ErxinAssessmentItem? cached = _itemDetailCache[itemNo];
    if (cached != null) {
      return Future<ErxinAssessmentItem>.value(cached);
    }
    return _itemDetailFetches.putIfAbsent(itemNo, () async {
      final ErxinAssessmentItem item = await widget.client.fetchTemplateItem(
        _token,
        itemNo: itemNo,
      );
      _itemDetailCache[itemNo] = item;
      return item;
    });
  }

  void _prefetchSelectedItem() {
    if (_selectedItemNo > 0 && _token.trim().isNotEmpty) {
      _detailFuture(_selectedItemNo);
    }
  }

  void _selectDomain(String domainCode) {
    setState(() {
      _selectedDomainCode = domainCode;
      _selectedItemNo = _firstVisibleItemNo(domainCode);
      _showFutureMonths = false;
    });
    _prefetchSelectedItem();
  }

  void _selectItem(int itemNo) {
    setState(() => _selectedItemNo = itemNo);
    _prefetchSelectedItem();
  }

  void _scoreItem(int itemNo, bool passed) {
    setState(() {
      _itemPasses[itemNo] = passed;
      _selectedItemNo = itemNo;
      _autoSaveText = '本地已记录，待自动保存';
    });
    _prefetchSelectedItem();
  }

  void _showFullItemDetail(
    ErxinAssessmentItem? item,
    ErxinItemSummary? summary,
  ) {
    final String title = item == null || item.itemNo <= 0
        ? summary == null
            ? '题目说明'
            : '${summary.itemNo} ${summary.itemTitle}'
        : '${item.itemNo} ${item.itemTitle}';
    showDialog<void>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(title),
          content: SizedBox(
            width: 640,
            child: SingleChildScrollView(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  _DialogTextBlock(title: '操作方法', text: item?.method ?? ''),
                  const SizedBox(height: 18),
                  _DialogTextBlock(
                    title: '通过标准',
                    text: item?.passCriteria ?? '',
                  ),
                ],
              ),
            ),
          ),
          actions: <Widget>[
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('关闭'),
            ),
          ],
        );
      },
    );
  }

  _DomainProgress _domainProgress(String domainCode) {
    final List<ErxinItemSummary> items = <ErxinItemSummary>[
      for (final int month in _visibleMonths) ..._itemsFor(domainCode, month),
    ];
    final int answered =
        items.where((ErxinItemSummary item) => _itemPasses.containsKey(item.itemNo)).length;
    return _DomainProgress(answered: answered, total: items.length);
  }

  int _completedDomainCount() {
    return _template.domains
        .where((ErxinDomain domain) {
          final _DomainProgress progress = _domainProgress(domain.domainCode);
          return progress.total > 0 && progress.answered >= progress.total;
        })
        .length
        .clamp(0, 5);
  }

  List<ErxinItemSummary> _itemsFor(String domainCode, int ageMonth) {
    for (final ErxinAgeGroup group in _template.ageGroups) {
      if (group.ageMonth == ageMonth) {
        return group.items
            .where((ErxinItemSummary item) => item.domainCode == domainCode)
            .toList();
      }
    }
    return <ErxinItemSummary>[];
  }

  bool _ageMonthComplete(String domainCode, int ageMonth) {
    final List<ErxinItemSummary> items = _itemsFor(domainCode, ageMonth);
    return items.isNotEmpty &&
        items.every((ErxinItemSummary item) => _itemPasses.containsKey(item.itemNo));
  }

  bool _ageMonthAllPassed(String domainCode, int ageMonth) {
    final List<ErxinItemSummary> items = _itemsFor(domainCode, ageMonth);
    return items.isNotEmpty &&
        items.every((ErxinItemSummary item) => _itemPasses[item.itemNo] == true);
  }

  int _firstVisibleItemNo(String domainCode) {
    for (final int month in _visibleMonths) {
      final List<ErxinItemSummary> items = _itemsFor(domainCode, month);
      if (items.isNotEmpty) {
        return items.first.itemNo;
      }
    }
    return 0;
  }

  ErxinItemSummary? _summaryByNo(int itemNo) {
    for (final ErxinAgeGroup group in _template.ageGroups) {
      for (final ErxinItemSummary item in group.items) {
        if (item.itemNo == itemNo) {
          return item;
        }
      }
    }
    return null;
  }

  String _domainName(String domainCode) {
    for (final ErxinDomain domain in _template.domains) {
      if (domain.domainCode == domainCode) {
        return domain.domainName.trim().isEmpty
            ? domain.domainCode
            : domain.domainName;
      }
    }
    return domainCode;
  }

  String _ageMonthSuffix(int month) {
    if (month == _mainAgeMonth) {
      return '主测';
    }
    final int mainIndex = _standardAgeMonths.indexOf(_mainAgeMonth);
    final int index = _standardAgeMonths.indexOf(month);
    if (mainIndex >= 0 && index >= 0 && index < mainIndex) {
      return '往前第${mainIndex - index}组';
    }
    if (mainIndex >= 0 && index >= 0 && index > mainIndex) {
      return '往后第${index - mainIndex}组';
    }
    return '';
  }

  String _nextActionText() {
    for (final int month in _visibleMonths) {
      for (final ErxinItemSummary item in _itemsFor(_selectedDomainCode, month)) {
        if (!_itemPasses.containsKey(item.itemNo)) {
          return '完成$month月龄第${item.itemNo}题';
        }
      }
    }
    if (_canEnterFutureMonths) {
      return '前测已达标，可以进入往后测查';
    }
    if (!_previousMonthsReady) {
      return '前两个标准月龄尚未连续全通过';
    }
    return '当前可见题目已完成';
  }
}

class _Header extends StatelessWidget {
  const _Header({
    required this.args,
    required this.mainAgeMonth,
    required this.autoSaveText,
    required this.onBack,
  });

  final ErxinAssessmentLaunchArgs args;
  final int mainAgeMonth;
  final String autoSaveText;
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 64,
      padding: const EdgeInsets.symmetric(horizontal: 20),
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(bottom: BorderSide(color: _ErxinColors.line)),
      ),
      child: Row(
        children: <Widget>[
          TextButton.icon(
            onPressed: onBack,
            icon: const Icon(Icons.arrow_back_rounded),
            label: const Text('返回'),
          ),
          const SizedBox(width: 8),
          const Text(
            '儿心量表-II 测评',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 22,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 22),
          Expanded(
            child: Text(
              '儿童：${args.studentName}｜出生：${args.birthDate}｜测查：${args.assessmentDate}｜实足年龄：${args.studentAge}',
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _ErxinColors.body,
                fontSize: 13,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
          _SmallBadge(text: '主测月龄 $mainAgeMonth月', strong: true),
          const SizedBox(width: 14),
          const Icon(Icons.check_circle, size: 18, color: _ErxinColors.green),
          const SizedBox(width: 5),
          Text(
            autoSaveText,
            style: const TextStyle(
              color: _ErxinColors.body,
              fontSize: 13,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(width: 14),
          OutlinedButton(onPressed: () {}, child: const Text('保存草稿')),
          const SizedBox(width: 8),
          FilledButton(
            onPressed: null,
            child: const Text('提交正式记录'),
          ),
        ],
      ),
    );
  }
}

class _DomainSidebar extends StatelessWidget {
  const _DomainSidebar({
    required this.domains,
    required this.selectedCode,
    required this.progressForDomain,
    required this.onSelect,
  });

  final List<ErxinDomain> domains;
  final String selectedCode;
  final _DomainProgress Function(String domainCode) progressForDomain;
  final ValueChanged<String> onSelect;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 230,
      padding: const EdgeInsets.fromLTRB(16, 18, 16, 14),
      decoration: const BoxDecoration(
        color: Color(0xFFFAFBFD),
        border: Border(right: BorderSide(color: _ErxinColors.line)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '能区进度',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 17,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 12),
          for (final ErxinDomain domain in domains)
            _DomainRow(
              domain: domain,
              selected: domain.domainCode == selectedCode,
              progress: progressForDomain(domain.domainCode),
              onTap: () => onSelect(domain.domainCode),
            ),
          const Spacer(),
          TextButton.icon(
            onPressed: () {},
            icon: const Icon(Icons.list_alt_rounded, size: 18),
            label: const Text('查看全部题目'),
          ),
        ],
      ),
    );
  }
}

class _DomainRow extends StatelessWidget {
  const _DomainRow({
    required this.domain,
    required this.selected,
    required this.progress,
    required this.onTap,
  });

  final ErxinDomain domain;
  final bool selected;
  final _DomainProgress progress;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final bool complete = progress.total > 0 && progress.answered >= progress.total;
    final double percent =
        progress.total <= 0 ? 0 : progress.answered / progress.total;
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: InkWell(
        borderRadius: BorderRadius.circular(8),
        onTap: onTap,
        child: Container(
          height: 56,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFEAF2FF) : Colors.white,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: _ErxinColors.line),
          ),
          child: Column(
            children: <Widget>[
              Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      domain.domainName,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        color: selected ? _ErxinColors.blue : _ErxinColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  Text(
                    complete ? '已完成' : progress.answered > 0 ? '测查中' : '待测',
                    style: TextStyle(
                      color: complete
                          ? _ErxinColors.green
                          : selected
                              ? _ErxinColors.blue
                              : _ErxinColors.muted,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(width: 6),
                  Text(
                    '${progress.answered}/${progress.total}',
                    style: const TextStyle(
                      color: _ErxinColors.body,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 7),
              ClipRRect(
                borderRadius: BorderRadius.circular(2),
                child: LinearProgressIndicator(
                  value: percent.clamp(0, 1),
                  minHeight: 4,
                  backgroundColor: const Color(0xFFE8EDF3),
                  valueColor: AlwaysStoppedAnimation<Color>(
                    complete ? _ErxinColors.green : _ErxinColors.blue,
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

class _AgeMonthSection extends StatelessWidget {
  const _AgeMonthSection({
    required this.month,
    required this.titleSuffix,
    required this.items,
    required this.itemPasses,
    required this.selectedItemNo,
    required this.onSelectItem,
    required this.onScore,
  });

  final int month;
  final String titleSuffix;
  final List<ErxinItemSummary> items;
  final Map<int, bool> itemPasses;
  final int selectedItemNo;
  final ValueChanged<int> onSelectItem;
  final void Function(int itemNo, bool passed) onScore;

  @override
  Widget build(BuildContext context) {
    final int answered =
        items.where((ErxinItemSummary item) => itemPasses.containsKey(item.itemNo)).length;
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: _ErxinColors.line),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: <Widget>[
          Container(
            height: 42,
            padding: const EdgeInsets.symmetric(horizontal: 14),
            decoration: const BoxDecoration(
              color: Color(0xFFF8FAFD),
              borderRadius: BorderRadius.vertical(top: Radius.circular(8)),
              border: Border(bottom: BorderSide(color: _ErxinColors.line)),
            ),
            child: Row(
              children: <Widget>[
                Text(
                  '$month月龄（$titleSuffix）',
                  style: const TextStyle(
                    color: _ErxinColors.ink,
                    fontSize: 16,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const Spacer(),
                Text(
                  '已测 $answered/${items.length}',
                  style: const TextStyle(
                    color: _ErxinColors.body,
                    fontSize: 13,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          for (final ErxinItemSummary item in items)
            _ItemScoreRow(
              item: item,
              selected: item.itemNo == selectedItemNo,
              passed: itemPasses[item.itemNo],
              onTap: () => onSelectItem(item.itemNo),
              onScore: (bool passed) => onScore(item.itemNo, passed),
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
    required this.onTap,
    required this.onScore,
  });

  final ErxinItemSummary item;
  final bool selected;
  final bool? passed;
  final VoidCallback onTap;
  final ValueChanged<bool> onScore;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Container(
        height: 58,
        padding: const EdgeInsets.symmetric(horizontal: 14),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFFBEB) : Colors.white,
          border: const Border(bottom: BorderSide(color: _ErxinColors.line)),
        ),
        child: Row(
          children: <Widget>[
            SizedBox(
              width: 48,
              child: Text(
                '${item.itemNo}',
                style: const TextStyle(
                  color: _ErxinColors.body,
                  fontSize: 15,
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
                        fontSize: 15,
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
            const SizedBox(width: 10),
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
    return SizedBox(
      width: 116,
      height: 40,
      child: OutlinedButton.icon(
        onPressed: onTap,
        icon: Icon(icon, size: 18),
        label: Text(label),
        style: OutlinedButton.styleFrom(
          foregroundColor: selected ? Colors.white : color,
          backgroundColor: selected ? color : Colors.white,
          side: BorderSide(color: selected ? color : _ErxinColors.line),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
          textStyle: const TextStyle(fontSize: 14, fontWeight: FontWeight.w900),
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
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: const Color(0xFFF8FAFD),
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
          const SizedBox(height: 5),
          Expanded(
            child: SingleChildScrollView(
              child: Text(
                text.trim().isEmpty ? '暂无内容' : text.trim(),
                style: const TextStyle(
                  color: _ErxinColors.body,
                  fontSize: 12,
                  height: 1.45,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _RemarkBox extends StatelessWidget {
  const _RemarkBox();

  @override
  Widget build(BuildContext context) {
    return TextField(
      maxLines: null,
      expands: true,
      decoration: InputDecoration(
        hintText: '添加本题备注',
        filled: true,
        fillColor: const Color(0xFFF8FAFD),
        contentPadding: const EdgeInsets.all(10),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: _ErxinColors.line),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: const BorderSide(color: _ErxinColors.line),
        ),
      ),
    );
  }
}

class _RuleCard extends StatelessWidget {
  const _RuleCard({
    required this.title,
    required this.body,
    required this.icon,
    required this.color,
  });

  final String title;
  final String body;
  final IconData icon;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: color, size: 24),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  title,
                  style: const TextStyle(
                    color: _ErxinColors.ink,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  body,
                  style: const TextStyle(
                    color: _ErxinColors.body,
                    fontSize: 13,
                    height: 1.35,
                    fontWeight: FontWeight.w700,
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

class _RuleChecklist extends StatelessWidget {
  const _RuleChecklist({required this.rows});

  final List<_RuleRow> rows;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Column(
        children: <Widget>[
          for (final _RuleRow row in rows)
            ListTile(
              dense: true,
              leading: Icon(
                row.done ? Icons.check_circle : Icons.radio_button_unchecked,
                color: row.done ? _ErxinColors.green : _ErxinColors.muted,
                size: 19,
              ),
              title: Text(
                row.label,
                style: const TextStyle(
                  color: _ErxinColors.ink,
                  fontWeight: FontWeight.w800,
                ),
              ),
              trailing: Text(
                row.value,
                style: TextStyle(
                  color: row.done ? _ErxinColors.green : _ErxinColors.body,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _ProgressSummary extends StatelessWidget {
  const _ProgressSummary({
    required this.domainStatus,
    required this.scaleStatus,
  });

  final String domainStatus;
  final String scaleStatus;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '完成情况',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 8),
          Text(domainStatus, style: _summaryStyle),
          const SizedBox(height: 5),
          Text('全量表：$scaleStatus', style: _summaryStyle),
        ],
      ),
    );
  }

  static const TextStyle _summaryStyle = TextStyle(
    color: _ErxinColors.body,
    fontSize: 13,
    fontWeight: FontWeight.w700,
  );
}

class _SmallBadge extends StatelessWidget {
  const _SmallBadge({required this.text, this.strong = false});

  final String text;
  final bool strong;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 11),
      decoration: BoxDecoration(
        color: strong ? const Color(0xFFEAF2FF) : const Color(0xFFF4F6FA),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: strong ? _ErxinColors.blue : _ErxinColors.line,
        ),
      ),
      alignment: Alignment.center,
      child: Text(
        text,
        style: TextStyle(
          color: strong ? _ErxinColors.blue : _ErxinColors.body,
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _MiniMarker extends StatelessWidget {
  const _MiniMarker({required this.text, this.warning = false});

  final String text;
  final bool warning;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 20,
      width: 20,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: warning ? const Color(0xFFFFF2E8) : const Color(0xFFEAF2FF),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: warning ? const Color(0xFFEA580C) : _ErxinColors.blue,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _DialogTextBlock extends StatelessWidget {
  const _DialogTextBlock({required this.title, required this.text});

  final String title;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          title,
          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w900),
        ),
        const SizedBox(height: 8),
        Text(
          text.trim().isEmpty ? '暂无内容' : text.trim(),
          style: const TextStyle(fontSize: 14, height: 1.55),
        ),
      ],
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.onBack});

  final String message;
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Text(message, style: const TextStyle(color: _ErxinColors.red)),
          const SizedBox(height: 16),
          OutlinedButton(onPressed: onBack, child: const Text('返回')),
        ],
      ),
    );
  }
}

class _RuleRow {
  const _RuleRow({
    required this.label,
    required this.value,
    required this.done,
  });

  final String label;
  final String value;
  final bool done;
}

class _DomainProgress {
  const _DomainProgress({required this.answered, required this.total});

  final int answered;
  final int total;
}

class _ErxinColors {
  static const Color page = Color(0xFFF5F7FA);
  static const Color ink = Color(0xFF172033);
  static const Color body = Color(0xFF566173);
  static const Color muted = Color(0xFF98A2B3);
  static const Color line = Color(0xFFE4E8EF);
  static const Color blue = Color(0xFF2563EB);
  static const Color green = Color(0xFF16A34A);
  static const Color red = Color(0xFFDC2626);
}

double _actualAgeMonths(String birthDate, String assessmentDate) {
  final DateTime? birth = DateTime.tryParse(birthDate);
  final DateTime? target = DateTime.tryParse(assessmentDate);
  if (birth == null || target == null || birth.isAfter(target)) {
    return 0;
  }
  final int days = target.difference(birth).inDays;
  return days / 30.0;
}
