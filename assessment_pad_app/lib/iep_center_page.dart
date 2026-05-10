import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'iep_assessment_record_client.dart';
import 'iep_plan_client.dart';
import 'pad_date_range_picker.dart';
import 'pad_responsive.dart';
import 'pad_top_message.dart';

class IepCenterPage extends StatefulWidget {
  const IepCenterPage({
    required this.onBack,
    this.recordClient = const ApiIepAssessmentRecordClient(),
    this.planClient = const ApiIepPlanClient(),
    super.key,
  });

  final VoidCallback onBack;
  final IepAssessmentRecordClient recordClient;
  final IepPlanClient planClient;

  @override
  State<IepCenterPage> createState() => _IepCenterPageState();
}

class _IepCenterPageState extends State<IepCenterPage> {
  IepAssessmentRecordSummary? _selectedRecord;

  void _selectRecord(IepAssessmentRecordSummary record) {
    final IepAssessmentRecordSummary? current = _selectedRecord;
    if (current != null && _sameRecord(current, record)) {
      return;
    }
    setState(() {
      _selectedRecord = record;
    });
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width =
            constraints.maxWidth.isFinite ? constraints.maxWidth : 1366;
        final _IepMetrics metrics = _IepMetrics.forWidth(width);

        return ColoredBox(
          color: _IepColors.page,
          child: Stack(
            children: <Widget>[
              Positioned.fill(child: CustomPaint(painter: _IepPagePainter())),
              Positioned(
                left: 0,
                top: 0,
                right: 0,
                child: _IepTopBar(onBack: widget.onBack, metrics: metrics),
              ),
              Positioned(
                left: metrics.outer,
                top: 84,
                width: metrics.leftWidth,
                height: 660,
                child: _StudentQueuePanel(
                  recordClient: widget.recordClient,
                  selectedRecord: _selectedRecord,
                  onRecordSelected: _selectRecord,
                ),
              ),
              Positioned(
                left: metrics.contentLeft,
                top: 84,
                width: metrics.contentWidth,
                height: 660,
                child: _IepWorkspace(
                  record: _selectedRecord,
                  planClient: widget.planClient,
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}

class _IepMetrics {
  const _IepMetrics({
    required this.outer,
    required this.gap,
    required this.leftWidth,
    required this.contentLeft,
    required this.contentWidth,
    required this.compact,
  });

  factory _IepMetrics.forWidth(double width) {
    final bool compact = width < 1180;
    final double outer = compact ? 14 : 24;
    final double gap = compact ? 10 : 14;
    final double leftWidth = compact ? 246 : 284;
    final double contentLeft = outer + leftWidth + gap;
    final double contentWidth = width - outer * 2 - leftWidth - gap;
    return _IepMetrics(
      outer: outer,
      gap: gap,
      leftWidth: leftWidth,
      contentLeft: contentLeft,
      contentWidth: contentWidth,
      compact: compact,
    );
  }

  final double outer;
  final double gap;
  final double leftWidth;
  final double contentLeft;
  final double contentWidth;
  final bool compact;
}

class _IepColors {
  static const Color page = Color(0xFFFFF6EC);
  static const Color surface = Color(0xFFFFFEFB);
  static const Color ink = Color(0xFF3E2A22);
  static const Color text = Color(0xFF72594D);
  static const Color muted = Color(0xFFB39B8C);
  static const Color line = Color(0xFFF0D9C8);
  static const Color lightLine = Color(0xFFF6E8DD);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95735);
  static const Color orangeSoft = Color(0xFFFFEEE4);
  static const Color green = Color(0xFF76A971);
  static const Color greenSoft = Color(0xFFEAF4E5);
  static const Color yellow = Color(0xFFE6A93A);
  static const Color yellowSoft = Color(0xFFFFF3D8);
}

List<BoxShadow> _iepShadow({
  Color color = const Color(0x16B05F32),
  double blur = 18,
  Offset offset = const Offset(0, 9),
}) {
  return <BoxShadow>[
    BoxShadow(color: color, blurRadius: blur, offset: offset),
  ];
}

DateTime _dateOnly(DateTime value) =>
    DateTime(value.year, value.month, value.day);

DateTime _periodEndFor(DateTime start, int monthCount) {
  final DateTime periodStart = _dateOnly(start);
  return DateTime(periodStart.year, periodStart.month + monthCount, 0);
}

DateTime _monthOnly(DateTime value) => DateTime(value.year, value.month);

DateTime _monthEnd(DateTime value) => DateTime(value.year, value.month + 1, 0);

DateTime _laterDate(DateTime left, DateTime right) =>
    left.isAfter(right) ? left : right;

DateTime _earlierDate(DateTime left, DateTime right) =>
    left.isBefore(right) ? left : right;

String _twoDigits(int value) => value.toString().padLeft(2, '0');

String _formatDateDash(DateTime value) {
  return '${value.year}-${_twoDigits(value.month)}-${_twoDigits(value.day)}';
}

String _formatDateDot(DateTime value) {
  return '${value.year}.${_twoDigits(value.month)}.${_twoDigits(value.day)}';
}

String _formatDotRange(DateTime start, DateTime end) {
  return '${_formatDateDot(start)}-${_formatDateDot(end)}';
}

String _formatZhRange(DateTime start, DateTime end) {
  return '${_formatDateDash(start)} 至 ${_formatDateDash(end)}';
}

String _monthLabelFromDate(DateTime value) => '${value.month}月';

List<String> _periodMonthLabels(DateTime start, int monthCount) {
  return List<String>.generate(monthCount, (int index) {
    return _monthLabelFromDate(DateTime(start.year, start.month + index));
  });
}

DateTime _monthDateFromLabel(
  DateTime periodStart,
  int monthCount,
  String label,
) {
  for (int index = 0; index < monthCount; index += 1) {
    final DateTime monthDate =
        DateTime(periodStart.year, periodStart.month + index);
    if (_monthLabelFromDate(monthDate) == label) {
      return monthDate;
    }
  }
  return _monthOnly(periodStart);
}

DateTimeRange _monthRangeInPeriod({
  required DateTime periodStart,
  required int monthCount,
  required DateTime monthDate,
}) {
  final DateTime start =
      _laterDate(_monthOnly(monthDate), _dateOnly(periodStart));
  final DateTime end = _earlierDate(
    _monthEnd(monthDate),
    _periodEndFor(periodStart, monthCount),
  );
  return DateTimeRange(start: start, end: end);
}

String _monthTrainingPeriodText(DateTimeRange monthRange, int index) {
  final int endDay = monthRange.end.day;
  final DateTime segmentStart = switch (index) {
    0 => DateTime(monthRange.start.year, monthRange.start.month, 1),
    1 => DateTime(monthRange.start.year, monthRange.start.month, 11),
    _ => DateTime(monthRange.start.year, monthRange.start.month, 21),
  };
  final DateTime segmentEnd = switch (index) {
    0 => DateTime(monthRange.start.year, monthRange.start.month, 10),
    1 => DateTime(monthRange.start.year, monthRange.start.month, 20),
    _ => DateTime(monthRange.start.year, monthRange.start.month, endDay),
  };
  final DateTime start = _laterDate(segmentStart, monthRange.start);
  final DateTime end = _earlierDate(segmentEnd, monthRange.end);
  if (start.isAfter(end)) {
    return '';
  }
  return '${_formatDateDash(start)}\n至 ${_formatDateDash(end)}';
}

List<DateTime> _weekDatesInMonthRange(
    DateTimeRange monthRange, int weekNumber) {
  DateTime cursor = _dateOnly(monthRange.start);
  while (cursor.weekday == DateTime.sunday && !cursor.isAfter(monthRange.end)) {
    cursor = cursor.add(const Duration(days: 1));
  }

  for (int currentWeek = 1; currentWeek <= 5; currentWeek += 1) {
    if (cursor.isAfter(monthRange.end)) {
      return <DateTime>[];
    }
    final DateTime end = _earlierDate(_weekEndForStart(cursor), monthRange.end);
    final List<DateTime> dates = <DateTime>[];
    for (DateTime day = cursor;
        !day.isAfter(end);
        day = day.add(const Duration(days: 1))) {
      if (day.weekday != DateTime.sunday) {
        dates.add(day);
      }
    }
    if (currentWeek == weekNumber) {
      return dates;
    }
    cursor = end.add(const Duration(days: 1));
    while (
        cursor.weekday == DateTime.sunday && !cursor.isAfter(monthRange.end)) {
      cursor = cursor.add(const Duration(days: 1));
    }
  }
  return <DateTime>[];
}

int _lastAvailableWeekInMonthRange(DateTimeRange monthRange) {
  int lastWeek = 1;
  for (int weekNumber = 1; weekNumber <= 5; weekNumber += 1) {
    if (_weekDatesInMonthRange(monthRange, weekNumber).isNotEmpty) {
      lastWeek = weekNumber;
    }
  }
  return lastWeek;
}

DateTime _weekEndForStart(DateTime value) {
  final int daysUntilSunday =
      (DateTime.sunday - value.weekday) % DateTime.daysPerWeek;
  return value.add(Duration(days: daysUntilSunday));
}

String _weekRangeText(List<DateTime> dates) {
  if (dates.isEmpty) {
    return '暂无日期';
  }
  return _formatZhRange(dates.first, dates.last);
}

String _weekDateLabel(DateTime value) {
  return '${_twoDigits(value.month)}.${_twoDigits(value.day)}';
}

class _IepPagePainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint cream = Paint()..color = const Color(0xFFFFE7C6);
    final Paint pale = Paint()..color = const Color(0xFFFFFBF4);
    canvas.drawCircle(Offset(size.width * .02, -100), 275, cream);
    canvas.drawOval(
      Rect.fromLTWH(size.width * .66, size.height - 76, 430, 210),
      cream..color = const Color(0xFFFFEED1),
    );
    canvas.drawOval(
      Rect.fromLTWH(-90, size.height - 38, 520, 150),
      pale,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _IepTopBar extends StatelessWidget {
  const _IepTopBar({required this.onBack, required this.metrics});

  final VoidCallback onBack;
  final _IepMetrics metrics;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 72,
      padding: EdgeInsets.symmetric(horizontal: metrics.outer),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.86),
        border: Border(
          bottom: BorderSide(color: _IepColors.line.withOpacity(.74)),
        ),
      ),
      child: Row(
        children: <Widget>[
          _IepBackButton(onTap: onBack),
          const SizedBox(width: 15),
          const Text(
            'IEP中心',
            style: TextStyle(
              color: _IepColors.ink,
              fontSize: 25,
              fontWeight: FontWeight.w900,
              height: 1,
            ),
          ),
          const Spacer(),
          if (!metrics.compact) ...<Widget>[
            const _TopSelector(label: '近30天', width: 108),
            const SizedBox(width: 12),
          ],
          _SearchBox(width: metrics.compact ? 194 : 224),
          const SizedBox(width: 12),
          const _SoftActionButton(
              icon: Icons.file_download_outlined, label: '导出Word'),
          const SizedBox(width: 10),
          const _SoftActionButton(icon: Icons.print_rounded, label: '打印'),
          const SizedBox(width: 10),
          const _PrimaryActionButton(
              icon: Icons.check_circle_rounded, label: '确认IEP'),
        ],
      ),
    );
  }
}

class _IepBackButton extends StatelessWidget {
  const _IepBackButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: _IepColors.surface.withOpacity(.94),
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          width: 42,
          height: 42,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: _IepColors.line),
          ),
          child: const Icon(
            Icons.chevron_left_rounded,
            color: _IepColors.text,
            size: 28,
          ),
        ),
      ),
    );
  }
}

class _TopSelector extends StatelessWidget {
  const _TopSelector({required this.label, required this.width});

  final String label;
  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.9),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _IepColors.line),
      ),
      child: Row(
        children: <Widget>[
          Expanded(
            child: Text(
              label,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _IepColors.text,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          const Icon(
            Icons.keyboard_arrow_down_rounded,
            color: _IepColors.muted,
            size: 20,
          ),
        ],
      ),
    );
  }
}

class _SearchBox extends StatelessWidget {
  const _SearchBox({required this.width});

  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.92),
        borderRadius: BorderRadius.circular(19),
        border: Border.all(color: _IepColors.line),
      ),
      child: Row(
        children: const <Widget>[
          Icon(Icons.search_rounded, color: _IepColors.ink, size: 21),
          SizedBox(width: 8),
          Expanded(
            child: Text(
              '搜索学员/评估老师',
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: _IepColors.muted,
                fontSize: 13,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _SoftActionButton extends StatelessWidget {
  const _SoftActionButton({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 13),
      decoration: BoxDecoration(
        color: _IepColors.orangeSoft,
        borderRadius: BorderRadius.circular(19),
        border: Border.all(color: const Color(0xFFFFCDB5)),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: _IepColors.orangeDeep, size: 19),
          const SizedBox(width: 5),
          Text(
            label,
            style: const TextStyle(
              color: _IepColors.orangeDeep,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _PrimaryActionButton extends StatelessWidget {
  const _PrimaryActionButton({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 15),
      decoration: BoxDecoration(
        color: _IepColors.orange,
        borderRadius: BorderRadius.circular(19),
        boxShadow: _iepShadow(
          color: const Color(0x2CE96F43),
          blur: 14,
          offset: const Offset(0, 7),
        ),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: Colors.white, size: 19),
          const SizedBox(width: 6),
          Text(
            label,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _StudentQueuePanel extends StatefulWidget {
  const _StudentQueuePanel({
    required this.recordClient,
    required this.selectedRecord,
    required this.onRecordSelected,
  });

  final IepAssessmentRecordClient recordClient;
  final IepAssessmentRecordSummary? selectedRecord;
  final ValueChanged<IepAssessmentRecordSummary> onRecordSelected;

  @override
  State<_StudentQueuePanel> createState() => _StudentQueuePanelState();
}

class _StudentQueuePanelState extends State<_StudentQueuePanel> {
  static const String _authTokenStorageKey = 'auth_token';

  List<IepAssessmentRecordSummary> _records = <IepAssessmentRecordSummary>[];
  bool _loading = true;
  String _error = '';
  int _totalCount = 0;
  _QueueFilter _filter = _QueueFilter.all;

  @override
  void initState() {
    super.initState();
    _loadRecords();
  }

  Future<void> _loadRecords() async {
    setState(() {
      _loading = true;
      _error = '';
    });
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepAssessmentRecordPage page =
          await widget.recordClient.fetchRecordsPage(
        token,
        pageIndex: 1,
        pageSize: 30,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _records = page.items;
        _totalCount = page.total;
        _loading = false;
      });
      final IepAssessmentRecordSummary? selectedRecord =
          _selectedRecordFrom(_visibleRecordsFor(_filter, page.items));
      if (selectedRecord != null) {
        widget.onRecordSelected(selectedRecord);
      }
    } on IepAssessmentRecordApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _error = error.message;
        _loading = false;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _error = '评估记录加载失败：$error';
        _loading = false;
      });
    }
  }

  IepAssessmentRecordSummary? _selectedRecordFrom(
    List<IepAssessmentRecordSummary> records,
  ) {
    if (records.isEmpty) {
      return null;
    }
    final IepAssessmentRecordSummary? selectedRecord = widget.selectedRecord;
    if (selectedRecord != null) {
      for (final IepAssessmentRecordSummary record in records) {
        if (_sameRecord(record, selectedRecord)) {
          return null;
        }
      }
    }
    return records.first;
  }

  void _changeFilter(_QueueFilter filter) {
    if (_filter == filter) {
      return;
    }
    setState(() {
      _filter = filter;
    });
    final IepAssessmentRecordSummary? selectedRecord =
        _selectedRecordFrom(_visibleRecordsFor(filter, _records));
    if (selectedRecord != null) {
      widget.onRecordSelected(selectedRecord);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.9),
        borderRadius: BorderRadius.circular(18),
        boxShadow: _iepShadow(),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: const <Widget>[
              Expanded(
                child: Text(
                  '学员IEP队列',
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: _IepColors.ink,
                    fontSize: 17,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              _QueueFilterButton(),
            ],
          ),
          const SizedBox(height: 12),
          _QueueTabs(selected: _filter, onChanged: _changeFilter),
          const SizedBox(height: 12),
          _CompactStatsStrip(records: _records, totalCount: _totalCount),
          const SizedBox(height: 12),
          Expanded(
            child: _QueueList(
              records: _visibleRecords,
              selectedRecord: widget.selectedRecord,
              loading: _loading,
              error: _error,
              onRetry: _loadRecords,
              onRecordSelected: widget.onRecordSelected,
            ),
          ),
        ],
      ),
    );
  }

  List<IepAssessmentRecordSummary> get _visibleRecords {
    return _visibleRecordsFor(_filter, _records);
  }
}

class _QueueFilterButton extends StatelessWidget {
  const _QueueFilterButton();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: _IepColors.orangeSoft,
        borderRadius: BorderRadius.circular(15),
      ),
      child: Row(
        children: const <Widget>[
          Icon(Icons.filter_alt_outlined, size: 16, color: _IepColors.orange),
          SizedBox(width: 4),
          Text(
            '筛选',
            style: TextStyle(
              color: _IepColors.orange,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

enum _QueueFilter { all, pending, draft }

class _QueueTabs extends StatelessWidget {
  const _QueueTabs({
    required this.selected,
    required this.onChanged,
  });

  final _QueueFilter selected;
  final ValueChanged<_QueueFilter> onChanged;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 34,
      padding: const EdgeInsets.all(3),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF4EC),
        borderRadius: BorderRadius.circular(17),
      ),
      child: Row(
        children: <Widget>[
          Expanded(
            child: _QueueTab(
              label: '全部',
              active: selected == _QueueFilter.all,
              onTap: () => onChanged(_QueueFilter.all),
            ),
          ),
          Expanded(
            child: _QueueTab(
              label: '待生成',
              active: selected == _QueueFilter.pending,
              onTap: () => onChanged(_QueueFilter.pending),
            ),
          ),
          Expanded(
            child: _QueueTab(
              label: '草稿',
              active: selected == _QueueFilter.draft,
              onTap: () => onChanged(_QueueFilter.draft),
            ),
          ),
        ],
      ),
    );
  }
}

class _QueueTab extends StatelessWidget {
  const _QueueTab({
    required this.label,
    required this.onTap,
    this.active = false,
  });

  final String label;
  final VoidCallback onTap;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: active ? Colors.white : Colors.transparent,
            borderRadius: BorderRadius.circular(14),
            boxShadow: active
                ? _iepShadow(
                    color: const Color(0x0FB05F32),
                    blur: 8,
                    offset: const Offset(0, 3),
                  )
                : const <BoxShadow>[],
          ),
          child: Text(
            label,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              color: active ? _IepColors.orangeDeep : _IepColors.text,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _CompactStatsStrip extends StatelessWidget {
  const _CompactStatsStrip({
    required this.records,
    required this.totalCount,
  });

  final List<IepAssessmentRecordSummary> records;
  final int totalCount;

  @override
  Widget build(BuildContext context) {
    final int pendingCount = records
        .where((IepAssessmentRecordSummary record) =>
            _QueueStatusStyle.fromPlanStatus(record.iepPlanStatus).label ==
            '待生成')
        .length;
    final int draftCount = records
        .where((IepAssessmentRecordSummary record) =>
            _QueueStatusStyle.fromPlanStatus(record.iepPlanStatus).label ==
            '草稿')
        .length;
    final int confirmedCount = records
        .where((IepAssessmentRecordSummary record) =>
            _QueueStatusStyle.fromPlanStatus(record.iepPlanStatus).label ==
            '已确认')
        .length;
    final String totalText =
        totalCount > 0 ? totalCount.toString() : records.length.toString();
    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        children: <Widget>[
          Expanded(child: _SmallStat(number: totalText, label: '评估记录')),
          const _StatDivider(),
          Expanded(child: _SmallStat(number: '$pendingCount', label: '待生成')),
          const _StatDivider(),
          Expanded(child: _SmallStat(number: '$draftCount', label: '草稿')),
          const _StatDivider(),
          Expanded(child: _SmallStat(number: '$confirmedCount', label: '已确认')),
        ],
      ),
    );
  }
}

class _SmallStat extends StatelessWidget {
  const _SmallStat({required this.number, required this.label});

  final String number;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: <Widget>[
        Text(
          number,
          style: const TextStyle(
            color: _IepColors.ink,
            fontSize: 18,
            fontWeight: FontWeight.w900,
            height: 1,
          ),
        ),
        const SizedBox(height: 6),
        Text(
          label,
          overflow: TextOverflow.ellipsis,
          style: const TextStyle(
            color: _IepColors.muted,
            fontSize: 10,
            fontWeight: FontWeight.w700,
          ),
        ),
      ],
    );
  }
}

class _StatDivider extends StatelessWidget {
  const _StatDivider();

  @override
  Widget build(BuildContext context) {
    return Container(width: 1, height: 28, color: _IepColors.lightLine);
  }
}

class _QueueList extends StatelessWidget {
  const _QueueList({
    required this.records,
    required this.selectedRecord,
    required this.loading,
    required this.error,
    required this.onRetry,
    required this.onRecordSelected,
  });

  final List<IepAssessmentRecordSummary> records;
  final IepAssessmentRecordSummary? selectedRecord;
  final bool loading;
  final String error;
  final VoidCallback onRetry;
  final ValueChanged<IepAssessmentRecordSummary> onRecordSelected;

  @override
  Widget build(BuildContext context) {
    if (loading) {
      return const _QueueStateView(
        icon: Icons.hourglass_top_rounded,
        title: '正在加载评估记录',
      );
    }
    if (error.trim().isNotEmpty) {
      return _QueueStateView(
        icon: Icons.wifi_off_rounded,
        title: '评估记录加载失败',
        message: error,
        actionLabel: '重试',
        onAction: onRetry,
      );
    }
    if (records.isEmpty) {
      return const _QueueStateView(
        icon: Icons.assignment_outlined,
        title: '暂无评估记录',
        message: '完成评估后会出现在这里',
      );
    }
    return ListView.separated(
      physics: const ClampingScrollPhysics(),
      padding: EdgeInsets.zero,
      itemCount: records.length,
      separatorBuilder: (_, __) => const SizedBox(height: 10),
      itemBuilder: (BuildContext context, int index) {
        final IepAssessmentRecordSummary record = records[index];
        return _QueueStudentCard(
          student: _QueueStudent.fromRecord(
            record,
            active: selectedRecord == null
                ? index == 0
                : _sameRecord(record, selectedRecord!),
          ),
          onTap: () => onRecordSelected(record),
        );
      },
    );
  }
}

class _QueueStateView extends StatelessWidget {
  const _QueueStateView({
    required this.icon,
    required this.title,
    this.message = '',
    this.actionLabel = '',
    this.onAction,
  });

  final IconData icon;
  final String title;
  final String message;
  final String actionLabel;
  final VoidCallback? onAction;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 16),
        decoration: BoxDecoration(
          color: const Color(0xFFFFFAF5),
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: _IepColors.lightLine),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Icon(icon, size: 28, color: _IepColors.orangeDeep),
            const SizedBox(height: 8),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _IepColors.ink,
                fontSize: 13,
                fontWeight: FontWeight.w900,
              ),
            ),
            if (message.trim().isNotEmpty) ...<Widget>[
              const SizedBox(height: 6),
              Text(
                message,
                textAlign: TextAlign.center,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _IepColors.text,
                  fontSize: 11,
                  fontWeight: FontWeight.w700,
                  height: 1.25,
                ),
              ),
            ],
            if (actionLabel.trim().isNotEmpty && onAction != null) ...<Widget>[
              const SizedBox(height: 10),
              _MiniQueueAction(label: actionLabel, onTap: onAction!),
            ],
          ],
        ),
      ),
    );
  }
}

class _MiniQueueAction extends StatelessWidget {
  const _MiniQueueAction({required this.label, required this.onTap});

  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: _IepColors.orangeSoft,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          height: 28,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          alignment: Alignment.center,
          child: Text(
            label,
            style: const TextStyle(
              color: _IepColors.orangeDeep,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _QueueStudent {
  const _QueueStudent({
    required this.name,
    required this.age,
    required this.status,
    required this.statusColor,
    required this.statusBg,
    required this.assessment,
    required this.period,
    required this.avatarAsset,
    this.active = false,
  });

  factory _QueueStudent.fromRecord(
    IepAssessmentRecordSummary record, {
    required bool active,
  }) {
    final _QueueStatusStyle status = _QueueStatusStyle.fromPlanStatus(
      record.iepPlanStatus,
    );
    return _QueueStudent(
      name: record.studentName.trim().isEmpty ? '未命名学员' : record.studentName,
      age: _recordAgeText(record),
      status: status.label,
      statusColor: status.color,
      statusBg: status.background,
      assessment:
          '${_recordAssessmentName(record)} · ${_recordDateText(record.assessmentDate)}',
      period: _recordPeriodText(record),
      avatarAsset: _avatarAssetForRecord(record),
      active: active,
    );
  }

  final String name;
  final String age;
  final String status;
  final Color statusColor;
  final Color statusBg;
  final String assessment;
  final String period;
  final String avatarAsset;
  final bool active;
}

class _QueueStatusStyle {
  const _QueueStatusStyle({
    required this.label,
    required this.color,
    required this.background,
  });

  factory _QueueStatusStyle.fromPlanStatus(String status) {
    return switch (status.trim()) {
      'confirmed' => const _QueueStatusStyle(
          label: '已确认',
          color: _IepColors.green,
          background: _IepColors.greenSoft,
        ),
      'draft' => const _QueueStatusStyle(
          label: '草稿',
          color: _IepColors.yellow,
          background: _IepColors.yellowSoft,
        ),
      _ => const _QueueStatusStyle(
          label: '待生成',
          color: _IepColors.orange,
          background: _IepColors.orangeSoft,
        ),
    };
  }

  final String label;
  final Color color;
  final Color background;
}

const List<String> _queueAvatarAssets = <String>[
  'assets/avatars/student_chenxu.png',
  'assets/avatars/student_chenxiaoyu.png',
  'assets/avatars/student_linyinuo.png',
  'assets/avatars/student_zhoushuyan.png',
  'assets/avatars/student_tangmuchen.png',
];

String _avatarAssetForRecord(IepAssessmentRecordSummary record) {
  final String seed = '${record.studentId}:${record.id}:${record.studentName}';
  int hash = 0;
  for (int index = 0; index < seed.length; index += 1) {
    hash = (hash * 31 + seed.codeUnitAt(index)) & 0x7fffffff;
  }
  return _queueAvatarAssets[hash % _queueAvatarAssets.length];
}

String _recordAgeText(IepAssessmentRecordSummary record) {
  if (record.ageYears > 0 || record.ageMonths > 0) {
    return '${record.ageYears}岁${record.ageMonths}月';
  }
  final DateTime? birthDate = DateTime.tryParse(record.birthDate);
  final DateTime? assessmentDate = DateTime.tryParse(record.assessmentDate);
  if (birthDate == null || assessmentDate == null) {
    return '年龄未知';
  }
  int totalMonths = (assessmentDate.year - birthDate.year) * 12 +
      assessmentDate.month -
      birthDate.month;
  if (assessmentDate.day < birthDate.day) {
    totalMonths -= 1;
  }
  if (totalMonths < 0) {
    return '年龄未知';
  }
  return '${totalMonths ~/ 12}岁${totalMonths % 12}月';
}

String _recordAssessmentName(IepAssessmentRecordSummary record) {
  if (record.source == 'ERXIN') {
    return '儿心量表';
  }
  if (record.assessmentName.trim().isNotEmpty) {
    return record.assessmentName.trim();
  }
  return record.assessmentCode == 'PEP3' ? 'PEP-3' : '评估记录';
}

String _recordDateText(String value) {
  final DateTime? date = DateTime.tryParse(value.trim());
  if (date == null) {
    return value.trim().isEmpty ? '-' : value.trim();
  }
  return _formatDateDash(date);
}

String _recordPeriodText(IepAssessmentRecordSummary record) {
  final DateTime? start = DateTime.tryParse(record.assessmentDate.trim());
  if (start == null) {
    return record.iepPlanStatus.trim().isEmpty ? '待确认周期' : '周期待同步';
  }
  final DateTime end = _periodEndFor(start, 3);
  return _formatDotRange(start, end);
}

bool _sameRecord(
  IepAssessmentRecordSummary left,
  IepAssessmentRecordSummary right,
) {
  return left.id == right.id &&
      left.source.trim().toUpperCase() == right.source.trim().toUpperCase();
}

List<IepAssessmentRecordSummary> _visibleRecordsFor(
  _QueueFilter filter,
  List<IepAssessmentRecordSummary> records,
) {
  return records.where((IepAssessmentRecordSummary record) {
    final String status =
        _QueueStatusStyle.fromPlanStatus(record.iepPlanStatus).label;
    return switch (filter) {
      _QueueFilter.all => true,
      _QueueFilter.pending => status == '待生成',
      _QueueFilter.draft => status == '草稿',
    };
  }).toList(growable: false);
}

String _workspaceTitle(IepAssessmentRecordSummary? record, IepPlan? plan) {
  final String studentName = plan?.student.name.trim().isNotEmpty == true
      ? plan!.student.name.trim()
      : (record?.studentName.trim().isNotEmpty == true
          ? record!.studentName.trim()
          : '未选择学员');
  final String planTitle =
      plan?.title.trim().isNotEmpty == true ? plan!.title.trim() : '康复教学计划';
  return '$studentName · $planTitle';
}

String _planStatusText(String? status) {
  return status?.trim() == 'confirmed' ? '已确认' : '草稿';
}

List<_DocDomainData> _docDomainsFromPlan(IepPlan plan) {
  final Map<String, List<IepPlanRow>> grouped = <String, List<IepPlanRow>>{};
  for (final IepPlanRow row in plan.rows) {
    final String domain = row.domain.trim().isEmpty ? '未分领域' : row.domain;
    grouped.putIfAbsent(domain, () => <IepPlanRow>[]).add(row);
  }
  return grouped.entries.map((MapEntry<String, List<IepPlanRow>> entry) {
    final List<String> longGoals = entry.value
        .map((IepPlanRow row) => row.longGoal.trim())
        .where((String value) => value.isNotEmpty)
        .toSet()
        .toList();
    final List<_DocShortGoalData> shortGoals =
        entry.value.map((IepPlanRow row) {
      return _DocShortGoalData(
        row.shortGoal,
        row.courseForm.trim().isEmpty ? '个训' : row.courseForm,
        row.startEndDate,
      );
    }).toList();
    return _DocDomainData(
      domain: entry.key,
      longGoals: longGoals.isEmpty ? <String>[''] : longGoals,
      shortGoals: shortGoals.isEmpty
          ? <_DocShortGoalData>[const _DocShortGoalData('', '个训', '')]
          : shortGoals,
    );
  }).toList();
}

_MonthDomainData _monthDomainFromPlanRow(IepMonthlyPlanRow row) {
  return _MonthDomainData(
    domain: row.domain,
    longGoal: row.longGoal,
    shortGoal: row.shortGoal,
    lesson: row.courseForm.trim().isEmpty ? '个训' : row.courseForm,
    trainings: row.trainingItems.isEmpty
        ? <_MonthTrainingData>[const _MonthTrainingData('', '')]
        : row.trainingItems.map((IepMonthlyTrainingItem item) {
            return _MonthTrainingData(item.content, item.startEndDate);
          }).toList(),
  );
}

_WeekTrainingRow _weekTrainingRowFromPlanRow(IepWeeklyPlanRow row) {
  return _WeekTrainingRow(project: row.project, content: row.content);
}

String _metaRangeText(IepPlanMeta? meta, {required String fallback}) {
  if (meta == null) {
    return fallback;
  }
  if (meta.startDate.isEmpty || meta.endDate.isEmpty) {
    return fallback;
  }
  return '${meta.startDate} 至 ${meta.endDate}';
}

List<DateTime>? _dateListFromStrings(List<String>? values) {
  if (values == null || values.isEmpty) {
    return null;
  }
  final List<DateTime> dates = values
      .map((String value) => DateTime.tryParse(value.trim()))
      .whereType<DateTime>()
      .map(_dateOnly)
      .toList();
  return dates.isEmpty ? null : dates;
}

class _QueueStudentCard extends StatelessWidget {
  const _QueueStudentCard({required this.student, required this.onTap});

  final _QueueStudent student;
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
          height: 80,
          padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 8),
          decoration: BoxDecoration(
            color: student.active ? const Color(0xFFFFF3EB) : Colors.white,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(
              color: student.active
                  ? const Color(0xFFFFB792)
                  : _IepColors.lightLine,
              width: student.active ? 1.4 : 1,
            ),
          ),
          child: Row(
            children: <Widget>[
              _QueueAvatar(asset: student.avatarAsset, active: student.active),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Row(
                      children: <Widget>[
                        Expanded(
                          child: Text(
                            '${student.name} · ${student.age}',
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              color: _IepColors.ink,
                              fontSize: 13,
                              fontWeight: FontWeight.w900,
                            ),
                          ),
                        ),
                        _StatusPill(
                          text: student.status,
                          color: student.statusColor,
                          bg: student.statusBg,
                        ),
                      ],
                    ),
                    const SizedBox(height: 6),
                    Text(
                      student.assessment,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _IepColors.text,
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        height: 1,
                      ),
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: <Widget>[
                        const Icon(
                          Icons.date_range_rounded,
                          size: 13,
                          color: _IepColors.muted,
                        ),
                        const SizedBox(width: 4),
                        Expanded(
                          child: Text(
                            student.period,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              color: _IepColors.muted,
                              fontSize: 11,
                              fontWeight: FontWeight.w700,
                              height: 1,
                            ),
                          ),
                        ),
                      ],
                    ),
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

class _QueueAvatar extends StatelessWidget {
  const _QueueAvatar({required this.asset, required this.active});

  final String asset;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 44,
      height: 44,
      padding: const EdgeInsets.all(2),
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: Colors.white,
        border: Border.all(
          color: active ? const Color(0xFFFFA878) : const Color(0xFFFFDFC8),
          width: active ? 2 : 1.4,
        ),
        boxShadow: active
            ? _iepShadow(
                color: const Color(0x22E96F43),
                blur: 10,
                offset: const Offset(0, 4),
              )
            : const <BoxShadow>[],
      ),
      child: ClipOval(
        child: Image.asset(
          asset,
          width: 40,
          height: 40,
          fit: BoxFit.cover,
          filterQuality: FilterQuality.medium,
        ),
      ),
    );
  }
}

class _StatusPill extends StatelessWidget {
  const _StatusPill({
    required this.text,
    required this.color,
    required this.bg,
  });

  final String text;
  final Color color;
  final Color bg;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 23,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: color,
          fontSize: 10,
          fontWeight: FontWeight.w900,
          height: 1,
        ),
      ),
    );
  }
}

enum _IepPreviewMode { total, month, week }

class _IepWorkspace extends StatefulWidget {
  const _IepWorkspace({
    required this.record,
    required this.planClient,
  });

  final IepAssessmentRecordSummary? record;
  final IepPlanClient planClient;

  @override
  State<_IepWorkspace> createState() => _IepWorkspaceState();
}

class _IepWorkspaceState extends State<_IepWorkspace> {
  static const String _authTokenStorageKey = 'auth_token';

  _IepPreviewMode _previewMode = _IepPreviewMode.total;
  String _previewMonth = '5月';
  int _previewWeek = 2;
  DateTime _periodStart = DateTime(2026, 5);
  DateTime? _periodEndOverride;
  int _periodMonthCount = 3;
  _GoalEditRequest? _selectedGoal;
  List<_DocDomainData> _totalPlanDomains = <_DocDomainData>[];
  IepPlanSaved? _savedPlan;
  IepExecutionPlansSaved? _executionPlans;
  bool _loadingPlan = false;
  bool _syncingPeriod = false;
  String _planError = '';
  int _loadTicket = 0;
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();

  DateTime get _periodEnd =>
      _periodEndOverride ?? _periodEndFor(_periodStart, _periodMonthCount);

  List<String> get _periodMonths =>
      _periodMonthLabels(_periodStart, _periodMonthCount);

  @override
  void initState() {
    super.initState();
    _syncPreviewMonthToPeriod();
    _loadPlanBundle();
  }

  @override
  void didUpdateWidget(covariant _IepWorkspace oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.record?.id != widget.record?.id ||
        oldWidget.record?.source != widget.record?.source) {
      _previewMode = _IepPreviewMode.total;
      _selectedGoal = null;
      _totalPlanDomains = <_DocDomainData>[];
      _savedPlan = null;
      _executionPlans = null;
      _planError = '';
      _initPeriodFromRecord(widget.record);
      _syncPreviewMonthToPeriod();
      _loadPlanBundle();
    }
  }

  @override
  void dispose() {
    _messageController.dispose();
    super.dispose();
  }

  Future<void> _loadPlanBundle() async {
    final IepAssessmentRecordSummary? record = widget.record;
    if (record == null) {
      setState(() {
        _loadingPlan = false;
        _planError = '';
        _savedPlan = null;
        _executionPlans = null;
        _totalPlanDomains = <_DocDomainData>[];
      });
      return;
    }
    final int ticket = ++_loadTicket;
    setState(() {
      _loadingPlan = true;
      _planError = '';
    });
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepPlanSaved savedPlan = await widget.planClient.fetchIepPlan(
        token,
        record: record,
        durationMonths: _periodMonthCount,
      );
      IepExecutionPlansSaved executionPlans =
          IepExecutionPlansSaved.empty(_periodMonthCount);
      if (savedPlan.hasContent) {
        executionPlans = await widget.planClient.fetchExecutionPlans(
          token,
          record: record,
          durationMonths: savedPlan.durationMonths,
        );
      }
      if (!mounted || ticket != _loadTicket) {
        return;
      }
      setState(() {
        _loadingPlan = false;
        _savedPlan = savedPlan;
        _executionPlans = executionPlans;
        _periodMonthCount = savedPlan.durationMonths == 6 ? 6 : 3;
        _applyPeriodFromPlan(savedPlan.plan, record);
        _totalPlanDomains = savedPlan.plan == null
            ? <_DocDomainData>[]
            : _docDomainsFromPlan(savedPlan.plan!);
        _syncPreviewMonthToPeriod();
      });
    } on IepPlanApiException catch (error) {
      if (!mounted || ticket != _loadTicket) {
        return;
      }
      setState(() {
        _loadingPlan = false;
        _planError = error.message;
      });
    } on Object catch (error) {
      if (!mounted || ticket != _loadTicket) {
        return;
      }
      setState(() {
        _loadingPlan = false;
        _planError = 'IEP计划加载失败：$error';
      });
    }
  }

  void _initPeriodFromRecord(IepAssessmentRecordSummary? record) {
    final DateTime? assessmentDate =
        DateTime.tryParse(record?.assessmentDate.trim() ?? '');
    if (assessmentDate == null) {
      return;
    }
    _periodStart = DateTime(assessmentDate.year, assessmentDate.month);
    _periodEndOverride = null;
  }

  void _applyPeriodFromPlan(IepPlan? plan, IepAssessmentRecordSummary record) {
    final DateTime? planStart = DateTime.tryParse(plan?.meta.startDate ?? '');
    final DateTime? planEnd = DateTime.tryParse(plan?.meta.endDate ?? '');
    if (planStart != null) {
      _periodStart = _dateOnly(planStart);
    } else {
      _initPeriodFromRecord(record);
    }
    _periodEndOverride = planEnd == null ? null : _dateOnly(planEnd);
  }

  void _syncPreviewMonthToPeriod() {
    final List<String> months = _periodMonths;
    if (months.isEmpty) {
      return;
    }
    if (!months.contains(_previewMonth)) {
      _previewMonth = months.first;
      _previewWeek = 1;
    }
  }

  int _previewMonthIndex() {
    final int index = _periodMonths.indexOf(_previewMonth);
    return index < 0 ? 1 : index + 1;
  }

  Future<void> _showEditPeriodDialog() async {
    if (_syncingPeriod) {
      return;
    }
    final _IepPeriodDraft? draft = await showDialog<_IepPeriodDraft>(
      context: context,
      barrierColor: const Color(0x33000000),
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _IepPeriodEditDialog(
            initialStart: _periodStart,
            monthCount: _periodMonthCount,
          ),
        );
      },
    );
    if (draft == null || !mounted) {
      return;
    }
    await _syncPeriodStart(draft.start);
  }

  Future<void> _syncPeriodStart(DateTime start) async {
    final IepAssessmentRecordSummary? record = widget.record;
    if (record == null) {
      _showMessage('请先选择左侧评估记录');
      return;
    }
    if (_savedPlan?.hasContent != true) {
      _showMessage('请先生成IEP计划后再调整周期');
      return;
    }
    final DateTime nextStart = _dateOnly(start);
    if (_dateOnly(_periodStart) == nextStart) {
      return;
    }
    ++_loadTicket;
    final int sourceDurationMonths = _savedPlan?.durationMonths == 6 ? 6 : 3;
    setState(() {
      _syncingPeriod = true;
      _planError = '';
    });
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepPlanPeriodSyncResult result =
          await widget.planClient.syncIepPlanPeriod(
        token,
        record: record,
        durationMonths: _periodMonthCount,
        sourceDurationMonths: sourceDurationMonths,
        startDate: nextStart,
      );
      if (!mounted) {
        return;
      }
      _applySyncedPlanBundle(result.iepPlan, result.executionPlans, record);
      _showMessage('计划周期和关联月/周计划日期已同步保存', tone: PadMessageTone.success);
    } on IepPlanApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _syncingPeriod = false;
        _planError = error.message;
      });
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      final String message = '同步计划周期失败：$error';
      setState(() {
        _syncingPeriod = false;
        _planError = message;
      });
      _showMessage(message);
    }
  }

  void _applySyncedPlanBundle(
    IepPlanSaved savedPlan,
    IepExecutionPlansSaved executionPlans,
    IepAssessmentRecordSummary record,
  ) {
    setState(() {
      _syncingPeriod = false;
      _loadingPlan = false;
      _savedPlan = savedPlan;
      _executionPlans = executionPlans;
      _periodMonthCount = savedPlan.durationMonths == 6 ? 6 : 3;
      _applyPeriodFromPlan(savedPlan.plan, record);
      _totalPlanDomains = savedPlan.plan == null
          ? <_DocDomainData>[]
          : _docDomainsFromPlan(savedPlan.plan!);
      _syncPreviewMonthToPeriod();
      _ensurePreviewWeekInRange();
    });
  }

  void _ensurePreviewWeekInRange() {
    if (_previewMode != _IepPreviewMode.week) {
      return;
    }
    final DateTime monthDate = _monthDateFromLabel(
      _periodStart,
      _periodMonthCount,
      _previewMonth,
    );
    final DateTimeRange monthRange = _monthRangeInPeriod(
      periodStart: _periodStart,
      monthCount: _periodMonthCount,
      monthDate: monthDate,
    );
    if (_weekDatesInMonthRange(monthRange, _previewWeek).isEmpty) {
      _previewWeek = 1;
    }
  }

  void _showMessage(String message,
      {PadMessageTone tone = PadMessageTone.info}) {
    if (!mounted || message.trim().isEmpty) {
      return;
    }
    _messageController.show(
      context,
      message,
      tone: tone,
      topMargin: 84,
      key: 'iep-center-message',
    );
  }

  void _showTotalPlan() {
    setState(() {
      _previewMode = _IepPreviewMode.total;
      _selectedGoal = null;
    });
  }

  void _showMonthPlan(String month) {
    setState(() {
      _previewMode = _IepPreviewMode.month;
      _previewMonth = month;
      _selectedGoal = null;
    });
  }

  void _showWeekPlan(String month, int weekNumber) {
    final DateTime monthDate = _monthDateFromLabel(
      _periodStart,
      _periodMonthCount,
      month,
    );
    final DateTimeRange monthRange = _monthRangeInPeriod(
      periodStart: _periodStart,
      monthCount: _periodMonthCount,
      monthDate: monthDate,
    );
    final int week = _weekDatesInMonthRange(monthRange, weekNumber).isEmpty
        ? _lastAvailableWeekInMonthRange(monthRange)
        : weekNumber;
    setState(() {
      _previewMode = _IepPreviewMode.week;
      _previewMonth = month;
      _previewWeek = week;
      _selectedGoal = null;
    });
  }

  void _changePeriodMonthCount(int monthCount) {
    if (_periodMonthCount == monthCount) {
      return;
    }
    setState(() {
      _periodMonthCount = monthCount;
      _periodEndOverride = null;
      final List<String> months = _periodMonths;
      if (!months.contains(_previewMonth)) {
        _previewMonth = months.first;
        _previewWeek = 1;
      }
      _ensurePreviewWeekInRange();
    });
    _loadPlanBundle();
  }

  Future<void> _showGoalEditDialog(_GoalEditRequest request) async {
    setState(() {
      _selectedGoal = request;
    });
    final _DocDomainData domain = _totalPlanDomains[request.domainIndex];
    final _GoalEditResult? result = await showDialog<_GoalEditResult>(
      context: context,
      barrierColor: const Color(0x33000000),
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _IepGoalEditDialog(
            domain: domain,
            request: request,
          ),
        );
      },
    );
    if (result == null || !mounted) {
      return;
    }
    setState(() {
      final List<_DocDomainData> nextDomains =
          List<_DocDomainData>.from(_totalPlanDomains);
      if (result.longGoals != null) {
        nextDomains[request.domainIndex] =
            domain.copyWith(longGoals: result.longGoals);
      }
      if (result.shortGoals != null) {
        nextDomains[request.domainIndex] =
            domain.copyWith(shortGoals: result.shortGoals);
      }
      _totalPlanDomains = nextDomains;
    });
  }

  void _clearSelectedGoal() {
    if (_selectedGoal == null) {
      return;
    }
    setState(() {
      _selectedGoal = null;
    });
  }

  void _handleGoalTap(_GoalEditRequest request) {
    if (_selectedGoal == request) {
      _showGoalEditDialog(request);
      return;
    }
    setState(() {
      _selectedGoal = request;
    });
  }

  @override
  Widget build(BuildContext context) {
    final IepAssessmentRecordSummary? record = widget.record;
    final IepPlan? plan = _savedPlan?.plan;
    final IepMonthlyPlan? monthPlan =
        _executionPlans?.monthPlan(_previewMonthIndex());
    final IepWeeklyPlan? weekPlan =
        _executionPlans?.weekPlan(_previewMonthIndex(), _previewWeek);
    final String title = _workspaceTitle(record, plan);
    final String statusText = _planStatusText(_savedPlan?.status);
    return Container(
      padding: const EdgeInsets.fromLTRB(18, 16, 18, 16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.92),
        borderRadius: BorderRadius.circular(18),
        boxShadow: _iepShadow(),
      ),
      child: Column(
        children: <Widget>[
          _WorkspaceHeader(
            title: title,
            statusText: statusText,
            periodText: _formatDotRange(_periodStart, _periodEnd),
          ),
          const SizedBox(height: 10),
          _PlanToolbar(
            onShowTotalPlan: _showTotalPlan,
            onShowMonthPlan: _showMonthPlan,
            onShowWeekPlan: _showWeekPlan,
            onEditPeriod: _showEditPeriodDialog,
            monthLabels: _periodMonths,
            periodMonthCount: _periodMonthCount,
            periodStart: _periodStart,
            onPeriodMonthCountChanged: _changePeriodMonthCount,
            syncingPeriod: _syncingPeriod,
          ),
          const SizedBox(height: 10),
          Expanded(
            child: _IepTablePreview(
              previewMode: _previewMode,
              month: _previewMonth,
              weekNumber: _previewWeek,
              periodStart: _periodStart,
              periodMonthCount: _periodMonthCount,
              record: record,
              plan: plan,
              monthPlan: monthPlan,
              weekPlan: weekPlan,
              loading: _loadingPlan,
              error: _planError,
              onRetry: _loadPlanBundle,
              totalPlanDomains: _totalPlanDomains,
              selectedGoal: _selectedGoal,
              onGoalTap: _handleGoalTap,
              onClearSelectedGoal: _clearSelectedGoal,
            ),
          ),
        ],
      ),
    );
  }
}

class _WorkspaceHeader extends StatelessWidget {
  const _WorkspaceHeader({
    required this.title,
    required this.statusText,
    required this.periodText,
  });

  final String title;
  final String statusText;
  final String periodText;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 42,
      child: Row(
        children: <Widget>[
          Expanded(
            child: Text(
              title,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _IepColors.ink,
                fontSize: 19,
                fontWeight: FontWeight.w900,
                height: 1,
              ),
            ),
          ),
          _HeaderMetaPill(
            icon: statusText == '已确认'
                ? Icons.verified_rounded
                : Icons.pending_actions_rounded,
            text: statusText,
            iconColor:
                statusText == '已确认' ? _IepColors.green : _IepColors.yellow,
          ),
          const SizedBox(width: 10),
          _HeaderMetaPill(
            icon: Icons.date_range_rounded,
            text: periodText,
          ),
          const SizedBox(width: 10),
          const _ClassContextPill(),
          const SizedBox(width: 10),
          const _StartClassButton(),
        ],
      ),
    );
  }
}

class _ClassContextPill extends StatelessWidget {
  const _ClassContextPill();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF6EE),
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: const Color(0xFFFFD3BA)),
      ),
      child: const Text(
        '第2月 · 第1周',
        style: TextStyle(
          color: _IepColors.orangeDeep,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _HeaderMetaPill extends StatelessWidget {
  const _HeaderMetaPill({
    required this.icon,
    required this.text,
    this.iconColor = _IepColors.muted,
  });

  final IconData icon;
  final String text;
  final Color iconColor;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF6),
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, size: 15, color: iconColor),
          const SizedBox(width: 5),
          Text(
            text,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: _IepColors.text,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _StartClassButton extends StatelessWidget {
  const _StartClassButton();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 36,
      padding: const EdgeInsets.symmetric(horizontal: 15),
      decoration: BoxDecoration(
        color: _IepColors.orange,
        borderRadius: BorderRadius.circular(18),
        boxShadow: _iepShadow(
          color: const Color(0x32E96F43),
          blur: 12,
          offset: const Offset(0, 5),
        ),
      ),
      child: Row(
        children: const <Widget>[
          Icon(Icons.play_circle_fill_rounded, color: Colors.white, size: 19),
          SizedBox(width: 6),
          Text(
            '开始上课',
            style: TextStyle(
              color: Colors.white,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _PlanToolbar extends StatelessWidget {
  const _PlanToolbar({
    required this.onShowTotalPlan,
    required this.onShowMonthPlan,
    required this.onShowWeekPlan,
    required this.onEditPeriod,
    required this.monthLabels,
    required this.periodMonthCount,
    required this.periodStart,
    required this.onPeriodMonthCountChanged,
    required this.syncingPeriod,
  });

  final VoidCallback onShowTotalPlan;
  final ValueChanged<String> onShowMonthPlan;
  final void Function(String month, int weekNumber) onShowWeekPlan;
  final VoidCallback onEditPeriod;
  final List<String> monthLabels;
  final int periodMonthCount;
  final DateTime periodStart;
  final ValueChanged<int> onPeriodMonthCountChanged;
  final bool syncingPeriod;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 46,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: <Widget>[
          _PeriodSwitch(
            selectedMonthCount: periodMonthCount,
            onChanged: onPeriodMonthCountChanged,
          ),
          const _ToolbarDivider(),
          Expanded(
            child: _ScrollablePlanNav(
              onShowTotalPlan: onShowTotalPlan,
              onShowMonthPlan: onShowMonthPlan,
              onShowWeekPlan: onShowWeekPlan,
              monthLabels: monthLabels,
              periodStart: periodStart,
              periodMonthCount: periodMonthCount,
            ),
          ),
          const _ToolbarDivider(),
          _TableTinyAction(
            icon: syncingPeriod
                ? Icons.hourglass_top_rounded
                : Icons.edit_calendar_rounded,
            label: syncingPeriod ? '同步中' : '编辑周期',
            onTap: syncingPeriod ? null : onEditPeriod,
          ),
          const SizedBox(width: 8),
          const _TableTinyAction(icon: Icons.refresh_rounded, label: '重新生成'),
        ],
      ),
    );
  }
}

class _ScrollablePlanNav extends StatefulWidget {
  const _ScrollablePlanNav({
    required this.onShowTotalPlan,
    required this.onShowMonthPlan,
    required this.onShowWeekPlan,
    required this.monthLabels,
    required this.periodStart,
    required this.periodMonthCount,
  });

  final VoidCallback onShowTotalPlan;
  final ValueChanged<String> onShowMonthPlan;
  final void Function(String month, int weekNumber) onShowWeekPlan;
  final List<String> monthLabels;
  final DateTime periodStart;
  final int periodMonthCount;

  @override
  State<_ScrollablePlanNav> createState() => _ScrollablePlanNavState();
}

class _ScrollablePlanNavState extends State<_ScrollablePlanNav>
    with SingleTickerProviderStateMixin {
  late final ScrollController _scrollController;
  late final AnimationController _hintController;
  late final Animation<double> _hintOffset;
  bool _showLeftHint = false;
  bool _showRightHint = false;
  String _selectedSection = 'iep';
  String _selectedMonth = '5月';
  int? _selectedWeek;

  @override
  void didUpdateWidget(covariant _ScrollablePlanNav oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!widget.monthLabels.contains(_selectedMonth) &&
        widget.monthLabels.isNotEmpty) {
      _selectedMonth = widget.monthLabels.first;
      _selectedWeek = null;
    }
    final int maxWeek = _weekCountForSelectedMonth();
    if (_selectedWeek != null && _selectedWeek! > maxWeek) {
      _selectedWeek = maxWeek;
    }
    WidgetsBinding.instance.addPostFrameCallback((_) => _syncHints());
  }

  @override
  void initState() {
    super.initState();
    _scrollController = ScrollController();
    _scrollController.addListener(_syncHints);
    _hintController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 760),
    )..repeat(reverse: true);
    _hintOffset = CurvedAnimation(
      parent: _hintController,
      curve: Curves.easeInOutCubic,
    );
    WidgetsBinding.instance.addPostFrameCallback((_) => _syncHints());
  }

  void _syncHints() {
    if (!_scrollController.hasClients) {
      return;
    }
    final ScrollPosition position = _scrollController.position;
    final bool canScroll = position.maxScrollExtent > 1;
    final bool nextLeft = canScroll && position.pixels > 1;
    final bool nextRight =
        canScroll && position.pixels < position.maxScrollExtent - 1;
    if (_showLeftHint == nextLeft && _showRightHint == nextRight) {
      return;
    }
    setState(() {
      _showLeftHint = nextLeft;
      _showRightHint = nextRight;
    });
  }

  void _selectPlan(String plan) {
    setState(() {
      _selectedSection = plan;
      _selectedWeek = null;
    });
    if (plan == 'iep') {
      widget.onShowTotalPlan();
    }
    WidgetsBinding.instance.addPostFrameCallback((_) => _syncHints());
  }

  void _selectMonth(String month) {
    setState(() {
      _selectedSection = 'month';
      _selectedMonth = month;
      _selectedWeek = null;
    });
    widget.onShowMonthPlan(month);
    WidgetsBinding.instance.addPostFrameCallback((_) => _syncHints());
  }

  void _selectWeek(int weekNumber) {
    setState(() {
      _selectedSection = 'week';
      _selectedWeek = weekNumber;
    });
    widget.onShowWeekPlan(_selectedMonth, weekNumber);
    WidgetsBinding.instance.addPostFrameCallback((_) => _syncHints());
  }

  @override
  void dispose() {
    _scrollController.removeListener(_syncHints);
    _scrollController.dispose();
    _hintController.dispose();
    super.dispose();
  }

  int _weekCountForSelectedMonth() {
    final DateTime selectedMonthDate = _monthDateFromLabel(
      widget.periodStart,
      widget.periodMonthCount,
      _selectedMonth,
    );
    final DateTimeRange selectedMonthRange = _monthRangeInPeriod(
      periodStart: widget.periodStart,
      monthCount: widget.periodMonthCount,
      monthDate: selectedMonthDate,
    );
    return _lastAvailableWeekInMonthRange(selectedMonthRange);
  }

  @override
  Widget build(BuildContext context) {
    final int weekCount = _weekCountForSelectedMonth();
    return Center(
      child: SizedBox(
        height: 34,
        child: Stack(
          alignment: Alignment.center,
          children: <Widget>[
            Center(
              child: SingleChildScrollView(
                controller: _scrollController,
                scrollDirection: Axis.horizontal,
                physics: const BouncingScrollPhysics(
                  parent: ClampingScrollPhysics(),
                ),
                padding: const EdgeInsets.only(right: 2),
                child: SizedBox(
                  height: 34,
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.center,
                    children: <Widget>[
                      _PlanTab(
                        text: 'IEP总计划',
                        active: _selectedSection == 'iep',
                        width: 92,
                        onTap: () => _selectPlan('iep'),
                      ),
                      const _PlanNavLabel(text: '月计划'),
                      ...widget.monthLabels.map((String month) {
                        return _PlanTab(
                          text: month,
                          active: _selectedSection == 'month' &&
                              _selectedMonth == month,
                          width: 54,
                          onTap: () => _selectMonth(month),
                        );
                      }),
                      const _PlanNavLabel(text: '周计划'),
                      ...List<Widget>.generate(weekCount, (int index) {
                        final int weekNumber = index + 1;
                        return _PlanTab(
                          text: '${_selectedMonth} W$weekNumber',
                          width: 68,
                          active: _selectedSection == 'week' &&
                              _selectedWeek == weekNumber,
                          activeTone: _PlanTabTone.week,
                          rightGap: weekNumber == weekCount ? 2 : 6,
                          onTap: () => _selectWeek(weekNumber),
                        );
                      }),
                    ],
                  ),
                ),
              ),
            ),
            Align(
              alignment: Alignment.centerLeft,
              child: _PlanScrollHint(
                visible: _showLeftHint,
                alignment: Alignment.centerLeft,
                direction: AxisDirection.left,
                animation: _hintOffset,
              ),
            ),
            Align(
              alignment: Alignment.centerRight,
              child: _PlanScrollHint(
                visible: _showRightHint,
                alignment: Alignment.centerRight,
                direction: AxisDirection.right,
                animation: _hintOffset,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _PlanScrollHint extends StatelessWidget {
  const _PlanScrollHint({
    required this.visible,
    required this.alignment,
    required this.direction,
    required this.animation,
  });

  final bool visible;
  final Alignment alignment;
  final AxisDirection direction;
  final Animation<double> animation;

  @override
  Widget build(BuildContext context) {
    final bool right = direction == AxisDirection.right;
    return IgnorePointer(
      child: AnimatedOpacity(
        opacity: visible ? 1 : 0,
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOut,
        child: Container(
          width: 62,
          height: 34,
          alignment: alignment,
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: right ? Alignment.centerLeft : Alignment.centerRight,
              end: right ? Alignment.centerRight : Alignment.centerLeft,
              colors: const <Color>[
                Color(0x00FFFAF5),
                Color(0xEFFFFAF5),
                Color(0xFFFFFAF5),
              ],
            ),
          ),
          child: AnimatedBuilder(
            animation: animation,
            builder: (BuildContext context, Widget? child) {
              final double dx = (animation.value * 5 + 1) * (right ? 1 : -1);
              return Transform.translate(
                offset: Offset(dx, 0),
                child: child,
              );
            },
            child: Padding(
              padding: EdgeInsets.only(
                left: right ? 0 : 4,
                right: right ? 4 : 0,
              ),
              child: Icon(
                right
                    ? Icons.chevron_right_rounded
                    : Icons.chevron_left_rounded,
                size: 27,
                color: _IepColors.orangeDeep.withOpacity(.86),
                shadows: const <Shadow>[
                  Shadow(
                    color: Color(0x22E96F43),
                    blurRadius: 5,
                    offset: Offset(0, 2),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _PeriodSwitch extends StatelessWidget {
  const _PeriodSwitch({
    required this.selectedMonthCount,
    required this.onChanged,
  });

  final int selectedMonthCount;
  final ValueChanged<int> onChanged;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.all(2),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: _IepColors.line),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: <Widget>[
          _PeriodOption(
            text: '3个月',
            active: selectedMonthCount == 3,
            onTap: () => onChanged(3),
          ),
          _PeriodOption(
            text: '6个月',
            active: selectedMonthCount == 6,
            onTap: () => onChanged(6),
          ),
        ],
      ),
    );
  }
}

class _PeriodOption extends StatelessWidget {
  const _PeriodOption({
    required this.text,
    required this.onTap,
    this.active = false,
  });

  final String text;
  final VoidCallback onTap;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(13),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Container(
          height: 26,
          width: 50,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: active ? _IepColors.orange : Colors.transparent,
            borderRadius: BorderRadius.circular(13),
          ),
          child: Text(
            text,
            style: TextStyle(
              color: active ? Colors.white : _IepColors.text,
              fontSize: 11,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _ToolbarDivider extends StatelessWidget {
  const _ToolbarDivider();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 1,
      height: 24,
      margin: const EdgeInsets.symmetric(horizontal: 8),
      color: _IepColors.lightLine,
    );
  }
}

class _TableTinyAction extends StatelessWidget {
  const _TableTinyAction({
    required this.icon,
    required this.label,
    this.onTap,
  });

  final IconData icon;
  final String label;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(15),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(15),
        child: Container(
          height: 30,
          padding: const EdgeInsets.symmetric(horizontal: 9),
          decoration: BoxDecoration(
            color: onTap == null ? const Color(0xFFF8EEE6) : Colors.white,
            borderRadius: BorderRadius.circular(15),
            border: Border.all(color: _IepColors.line),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.center,
            children: <Widget>[
              Icon(
                icon,
                color: onTap == null ? _IepColors.muted : _IepColors.text,
                size: 16,
              ),
              const SizedBox(width: 5),
              Text(
                label,
                style: TextStyle(
                  color: onTap == null ? _IepColors.muted : _IepColors.text,
                  fontSize: 11,
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

class _IepPeriodDraft {
  const _IepPeriodDraft({required this.start});

  final DateTime start;
}

class _IepPeriodEditDialog extends StatefulWidget {
  const _IepPeriodEditDialog({
    required this.initialStart,
    required this.monthCount,
  });

  final DateTime initialStart;
  final int monthCount;

  @override
  State<_IepPeriodEditDialog> createState() => _IepPeriodEditDialogState();
}

class _IepPeriodEditDialogState extends State<_IepPeriodEditDialog> {
  late DateTime _start;

  DateTime get _end => _periodEndFor(_start, widget.monthCount);

  @override
  void initState() {
    super.initState();
    _start = _dateOnly(widget.initialStart);
  }

  Future<void> _pickStartDate() async {
    final DateTime? picked = await showPadDatePicker(
      context: context,
      title: '选择周期开始日期',
      helperText: '请选择周期开始日期，结束日期将按自然月自动计算',
      initialDate: _start,
      today: _start,
      minDate: DateTime(_start.year - 1, 1),
      maxDate: DateTime(_start.year + 1, 12, 31),
    );
    if (picked == null || !mounted) {
      return;
    }
    setState(() {
      _start = _dateOnly(picked);
    });
  }

  void _submit() {
    Navigator.of(context).pop(_IepPeriodDraft(start: _start));
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: 540,
        padding: const EdgeInsets.fromLTRB(24, 22, 24, 22),
        decoration: BoxDecoration(
          color: _IepColors.surface,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: _IepColors.line),
          boxShadow: _iepShadow(
            color: const Color(0x20B05F32),
            blur: 32,
            offset: const Offset(0, 16),
          ),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                const Text(
                  '编辑周期',
                  style: TextStyle(
                    color: _IepColors.ink,
                    fontSize: 22,
                    fontWeight: FontWeight.w900,
                    height: 1,
                  ),
                ),
                const Spacer(),
                _IepPeriodTypePill(text: '${widget.monthCount}个月周期'),
                const SizedBox(width: 10),
                _IepDialogIconButton(
                  icon: Icons.close_rounded,
                  onTap: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 22),
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: const Color(0xFFFFFAF6),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: _IepColors.lightLine),
              ),
              child: Column(
                children: <Widget>[
                  _IepPeriodDateTile(
                    label: '周期开始',
                    value: _formatDateDash(_start),
                    icon: Icons.event_available_rounded,
                    onTap: _pickStartDate,
                  ),
                  const SizedBox(height: 10),
                  _IepPeriodDateTile(
                    label: '周期结束',
                    value: _formatDateDash(_end),
                    icon: Icons.event_repeat_rounded,
                    trailingText: '自动计算',
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
            Row(
              children: <Widget>[
                const Icon(
                  Icons.info_outline_rounded,
                  size: 16,
                  color: _IepColors.muted,
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    '同步后会按${widget.monthCount}个月周期更新表格中的实施日期，并重算对应月计划、周计划起止日期。',
                    style: const TextStyle(
                      color: _IepColors.text,
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      height: 1.35,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 22),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: <Widget>[
                _IepDialogAction(
                  label: '取消',
                  onTap: () => Navigator.of(context).pop(),
                ),
                const SizedBox(width: 10),
                _IepDialogAction(
                  label: '确认同步',
                  filled: true,
                  onTap: _submit,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _IepPeriodTypePill extends StatelessWidget {
  const _IepPeriodTypePill({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: _IepColors.orangeSoft,
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: const Color(0xFFFFD8C6)),
      ),
      child: Text(
        text,
        style: const TextStyle(
          color: _IepColors.orangeDeep,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _IepPeriodDateTile extends StatelessWidget {
  const _IepPeriodDateTile({
    required this.label,
    required this.value,
    required this.icon,
    this.trailingText,
    this.onTap,
  });

  final String label;
  final String value;
  final IconData icon;
  final String? trailingText;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final bool clickable = onTap != null;
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(13),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Container(
          height: 58,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(13),
            border: Border.all(
              color: clickable ? const Color(0xFFFFCDB4) : _IepColors.line,
            ),
          ),
          child: Row(
            children: <Widget>[
              Container(
                width: 34,
                height: 34,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: clickable
                      ? _IepColors.orangeSoft
                      : const Color(0xFFF8EEE6),
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  icon,
                  size: 19,
                  color: clickable ? _IepColors.orangeDeep : _IepColors.muted,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text(
                      label,
                      style: const TextStyle(
                        color: _IepColors.muted,
                        fontSize: 11,
                        fontWeight: FontWeight.w900,
                        height: 1,
                      ),
                    ),
                    const SizedBox(height: 7),
                    Text(
                      value,
                      style: const TextStyle(
                        color: _IepColors.ink,
                        fontSize: 17,
                        fontWeight: FontWeight.w900,
                        height: 1,
                      ),
                    ),
                  ],
                ),
              ),
              if (trailingText != null)
                Text(
                  trailingText!,
                  style: const TextStyle(
                    color: _IepColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                )
              else
                const Icon(
                  Icons.chevron_right_rounded,
                  color: _IepColors.orangeDeep,
                  size: 24,
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _IepDialogIconButton extends StatelessWidget {
  const _IepDialogIconButton({required this.icon, required this.onTap});

  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      shape: const CircleBorder(),
      child: InkWell(
        onTap: onTap,
        customBorder: const CircleBorder(),
        child: Container(
          width: 36,
          height: 36,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: const Color(0xFFFFFAF6),
            shape: BoxShape.circle,
            border: Border.all(color: _IepColors.lightLine),
          ),
          child: Icon(icon, size: 21, color: _IepColors.text),
        ),
      ),
    );
  }
}

class _IepDialogAction extends StatelessWidget {
  const _IepDialogAction({
    required this.label,
    required this.onTap,
    this.filled = false,
    this.icon,
  });

  final String label;
  final VoidCallback onTap;
  final bool filled;
  final IconData? icon;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          height: 42,
          constraints: const BoxConstraints(minWidth: 92),
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(horizontal: 20),
          decoration: BoxDecoration(
            color: filled ? _IepColors.orange : Colors.white,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(
              color: filled ? _IepColors.orange : _IepColors.line,
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              if (icon != null) ...<Widget>[
                Icon(
                  icon,
                  size: 16,
                  color: filled ? Colors.white : _IepColors.text,
                ),
                const SizedBox(width: 5),
              ],
              Text(
                label,
                style: TextStyle(
                  color: filled ? Colors.white : _IepColors.text,
                  fontSize: 14,
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

enum _GoalEditType { longGoal, shortGoal }

class _GoalEditRequest {
  const _GoalEditRequest._({
    required this.domainIndex,
    required this.type,
    this.shortGoalIndex,
  });

  factory _GoalEditRequest.longGoal({required int domainIndex}) {
    return _GoalEditRequest._(
      domainIndex: domainIndex,
      type: _GoalEditType.longGoal,
    );
  }

  factory _GoalEditRequest.shortGoal({
    required int domainIndex,
    required int shortGoalIndex,
  }) {
    return _GoalEditRequest._(
      domainIndex: domainIndex,
      type: _GoalEditType.shortGoal,
      shortGoalIndex: shortGoalIndex,
    );
  }

  final int domainIndex;
  final _GoalEditType type;
  final int? shortGoalIndex;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        other is _GoalEditRequest &&
            other.domainIndex == domainIndex &&
            other.type == type &&
            other.shortGoalIndex == shortGoalIndex;
  }

  @override
  int get hashCode => Object.hash(domainIndex, type, shortGoalIndex);
}

class _GoalEditResult {
  const _GoalEditResult({
    this.longGoals,
    this.shortGoals,
    this.syncRelatedPlans = false,
  });

  final List<String>? longGoals;
  final List<_DocShortGoalData>? shortGoals;
  final bool syncRelatedPlans;
}

class _IepGoalEditDialog extends StatefulWidget {
  const _IepGoalEditDialog({
    required this.domain,
    required this.request,
  });

  final _DocDomainData domain;
  final _GoalEditRequest request;

  @override
  State<_IepGoalEditDialog> createState() => _IepGoalEditDialogState();
}

class _IepGoalEditDialogState extends State<_IepGoalEditDialog> {
  final List<TextEditingController> _longGoalControllers =
      <TextEditingController>[];
  final List<_ShortGoalDraft> _shortGoalDrafts = <_ShortGoalDraft>[];

  bool get _editingLongGoal => widget.request.type == _GoalEditType.longGoal;

  bool get _canRemoveCurrentShortGoal {
    return !_editingLongGoal &&
        (widget.domain.shortGoals.length > 1 || _shortGoalDrafts.length > 1);
  }

  int get _shortGoalIndex {
    final int index = widget.request.shortGoalIndex ?? 0;
    if (widget.domain.shortGoals.isEmpty || index <= 0) {
      return 0;
    }
    final int lastIndex = widget.domain.shortGoals.length - 1;
    return index > lastIndex ? lastIndex : index;
  }

  _DocShortGoalData get _shortGoal {
    if (widget.domain.shortGoals.isEmpty) {
      return const _DocShortGoalData('', '个训', '');
    }
    return widget.domain.shortGoals[_shortGoalIndex];
  }

  @override
  void initState() {
    super.initState();
    _longGoalControllers.addAll(
      widget.domain.longGoals
          .map((String goal) => TextEditingController(text: goal)),
    );
    if (!_editingLongGoal) {
      _shortGoalDrafts.add(
        _ShortGoalDraft(
          goal: _shortGoal.goal,
          lesson: _normalizeLesson(_shortGoal.lesson),
          period: _shortGoal.period,
        ),
      );
    }
  }

  @override
  void dispose() {
    for (final TextEditingController controller in _longGoalControllers) {
      controller.dispose();
    }
    for (final _ShortGoalDraft draft in _shortGoalDrafts) {
      draft.dispose();
    }
    super.dispose();
  }

  String _normalizeLesson(String lesson) {
    return lesson == '集体课' ? '集体课' : '个训';
  }

  void _addLongGoal() {
    setState(() {
      _longGoalControllers.add(TextEditingController());
    });
  }

  void _removeLongGoal(int index) {
    if (_longGoalControllers.length <= 1) {
      return;
    }
    setState(() {
      _longGoalControllers.removeAt(index).dispose();
    });
  }

  void _addShortGoal() {
    setState(() {
      _shortGoalDrafts.add(
        _ShortGoalDraft(
          goal: '',
          lesson: '个训',
          period: _shortGoal.period,
        ),
      );
    });
  }

  void _removeShortGoal(int index) {
    if (index < 0 || index >= _shortGoalDrafts.length) {
      return;
    }
    if (index == 0 && !_canRemoveCurrentShortGoal) {
      return;
    }
    setState(() {
      _shortGoalDrafts.removeAt(index).dispose();
    });
  }

  void _changeShortGoalLesson(int index, String lesson) {
    setState(() {
      _shortGoalDrafts[index].lesson = lesson;
    });
  }

  void _submit({required bool syncRelatedPlans}) {
    if (_editingLongGoal) {
      final List<String> goals = _longGoalControllers
          .map((TextEditingController controller) => controller.text.trim())
          .where((String value) => value.isNotEmpty)
          .toList();
      Navigator.of(context).pop(
        _GoalEditResult(
          longGoals: goals.isEmpty ? <String>[''] : goals,
          syncRelatedPlans: syncRelatedPlans,
        ),
      );
      return;
    }
    final List<_DocShortGoalData> editedShortGoals = _shortGoalDrafts
        .map((_ShortGoalDraft draft) => draft.toData())
        .where((_DocShortGoalData data) => data.goal.trim().isNotEmpty)
        .toList();
    final List<_DocShortGoalData> nextShortGoals =
        List<_DocShortGoalData>.from(widget.domain.shortGoals);
    if (nextShortGoals.isEmpty) {
      nextShortGoals.addAll(
        editedShortGoals.isEmpty
            ? <_DocShortGoalData>[_shortGoal.copyWith(goal: '')]
            : editedShortGoals,
      );
    } else if (editedShortGoals.isEmpty) {
      if (nextShortGoals.length > 1) {
        nextShortGoals.removeAt(_shortGoalIndex);
      } else {
        nextShortGoals[_shortGoalIndex] = _shortGoal.copyWith(goal: '');
      }
    } else {
      nextShortGoals
        ..removeAt(_shortGoalIndex)
        ..insertAll(_shortGoalIndex, editedShortGoals);
    }
    Navigator.of(context).pop(
      _GoalEditResult(
        shortGoals: nextShortGoals,
        syncRelatedPlans: syncRelatedPlans,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final int? shortGoalIndex = widget.request.shortGoalIndex;
    final String title = _editingLongGoal ? '编辑长期目标' : '编辑短期目标';
    final String location = _editingLongGoal
        ? '${widget.domain.domain} · 长期目标'
        : '${widget.domain.domain} · 短期目标${(shortGoalIndex ?? 0) + 1}';
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: _editingLongGoal ? 660 : 620,
        padding: const EdgeInsets.fromLTRB(24, 22, 24, 22),
        decoration: BoxDecoration(
          color: _IepColors.surface,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: _IepColors.line),
          boxShadow: _iepShadow(
            color: const Color(0x20B05F32),
            blur: 32,
            offset: const Offset(0, 16),
          ),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                Text(
                  title,
                  style: const TextStyle(
                    color: _IepColors.ink,
                    fontSize: 22,
                    fontWeight: FontWeight.w900,
                    height: 1,
                  ),
                ),
                const SizedBox(width: 12),
                _IepPeriodTypePill(text: location),
                const Spacer(),
                _IepDialogIconButton(
                  icon: Icons.close_rounded,
                  onTap: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 20),
            if (_editingLongGoal)
              _LongGoalEditor(
                controllers: _longGoalControllers,
                onAdd: _addLongGoal,
                onRemove: _removeLongGoal,
              )
            else
              _ShortGoalEditor(
                drafts: _shortGoalDrafts,
                firstShortGoalNumber: _shortGoalIndex + 1,
                canRemoveCurrent: _canRemoveCurrentShortGoal,
                onLessonChanged: _changeShortGoalLesson,
                onAdd: _addShortGoal,
                onRemove: _removeShortGoal,
              ),
            const SizedBox(height: 12),
            const _GoalSyncHint(),
            const SizedBox(height: 20),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: <Widget>[
                _IepDialogAction(
                  label: '取消',
                  onTap: () => Navigator.of(context).pop(),
                ),
                const SizedBox(width: 10),
                _IepDialogAction(
                  label: '仅保存当前表格',
                  icon: Icons.save_outlined,
                  onTap: () => _submit(syncRelatedPlans: false),
                ),
                const SizedBox(width: 10),
                _IepDialogAction(
                  label: '保存并同步',
                  filled: true,
                  icon: Icons.sync_rounded,
                  onTap: () => _submit(syncRelatedPlans: true),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _ShortGoalDraft {
  _ShortGoalDraft({
    required String goal,
    required this.lesson,
    required String period,
  })  : goalController = TextEditingController(text: goal),
        periodController = TextEditingController(text: period);

  final TextEditingController goalController;
  final TextEditingController periodController;
  String lesson;

  _DocShortGoalData toData() {
    return _DocShortGoalData(
      goalController.text.trim(),
      lesson,
      periodController.text.trim(),
    );
  }

  void dispose() {
    goalController.dispose();
    periodController.dispose();
  }
}

class _LongGoalEditor extends StatelessWidget {
  const _LongGoalEditor({
    required this.controllers,
    required this.onAdd,
    required this.onRemove,
  });

  final List<TextEditingController> controllers;
  final VoidCallback onAdd;
  final ValueChanged<int> onRemove;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(maxHeight: 330),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF6),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: SingleChildScrollView(
        child: Column(
          children: <Widget>[
            for (int index = 0;
                index < controllers.length;
                index += 1) ...<Widget>[
              _GoalTextField(
                label: '长期目标 ${index + 1}',
                controller: controllers[index],
                minLines: 2,
                maxLines: 3,
                trailing: _SmallIconAction(
                  icon: Icons.delete_outline_rounded,
                  enabled: controllers.length > 1,
                  onTap: () => onRemove(index),
                ),
              ),
              const SizedBox(height: 10),
            ],
            Align(
              alignment: Alignment.centerLeft,
              child: _AddGoalButton(onTap: onAdd),
            ),
          ],
        ),
      ),
    );
  }
}

class _ShortGoalEditor extends StatelessWidget {
  const _ShortGoalEditor({
    required this.drafts,
    required this.firstShortGoalNumber,
    required this.canRemoveCurrent,
    required this.onLessonChanged,
    required this.onAdd,
    required this.onRemove,
  });

  final List<_ShortGoalDraft> drafts;
  final int firstShortGoalNumber;
  final bool canRemoveCurrent;
  final void Function(int index, String lesson) onLessonChanged;
  final VoidCallback onAdd;
  final ValueChanged<int> onRemove;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(maxHeight: 372),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF6),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: <Widget>[
              const Text(
                '当前短期目标',
                style: TextStyle(
                  color: _IepColors.text,
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const Spacer(),
              _AddGoalButton(
                label: '新增一条短期目标',
                onTap: onAdd,
              ),
            ],
          ),
          const SizedBox(height: 10),
          Expanded(
            child: SingleChildScrollView(
              child: Column(
                children: <Widget>[
                  for (int index = 0; index < drafts.length; index += 1)
                    Padding(
                      padding: EdgeInsets.only(
                        bottom: index == drafts.length - 1 ? 0 : 12,
                      ),
                      child: _ShortGoalDraftCard(
                        index: index,
                        number: firstShortGoalNumber + index,
                        draft: drafts[index],
                        canRemove: index > 0 || canRemoveCurrent,
                        onLessonChanged: (String value) =>
                            onLessonChanged(index, value),
                        onRemove: () => onRemove(index),
                      ),
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

class _ShortGoalDraftCard extends StatelessWidget {
  const _ShortGoalDraftCard({
    required this.index,
    required this.number,
    required this.draft,
    required this.canRemove,
    required this.onLessonChanged,
    required this.onRemove,
  });

  final int index;
  final int number;
  final _ShortGoalDraft draft;
  final bool canRemove;
  final ValueChanged<String> onLessonChanged;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _IepColors.line),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: <Widget>[
              Text(
                '短期目标 $number',
                style: const TextStyle(
                  color: _IepColors.ink,
                  fontSize: 13,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const Spacer(),
              _SmallIconAction(
                icon: Icons.delete_outline_rounded,
                enabled: canRemove,
                onTap: onRemove,
              ),
            ],
          ),
          const SizedBox(height: 8),
          _GoalTextField(
            fieldKey: ValueKey<String>('short-goal-$index-goal'),
            label: '目标内容',
            controller: draft.goalController,
            minLines: 2,
            maxLines: 4,
          ),
          const SizedBox(height: 12),
          Row(
            children: <Widget>[
              Expanded(
                flex: 8,
                child: _LessonSegmentedPicker(
                  label: '课程形式',
                  value: draft.lesson,
                  optionKeyPrefix: 'short-goal-$index-lesson',
                  onChanged: onLessonChanged,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                flex: 13,
                child: _GoalTextField(
                  label: '起止日期',
                  controller: draft.periodController,
                  minLines: 1,
                  maxLines: 1,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _LessonSegmentedPicker extends StatelessWidget {
  const _LessonSegmentedPicker({
    required this.label,
    required this.value,
    required this.onChanged,
    this.optionKeyPrefix,
  });

  final String label;
  final String value;
  final ValueChanged<String> onChanged;
  final String? optionKeyPrefix;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          label,
          style: const TextStyle(
            color: _IepColors.text,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 7),
        Container(
          height: 42,
          padding: const EdgeInsets.all(3),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: _IepColors.line),
          ),
          child: Row(
            children: <Widget>[
              _LessonOption(
                label: '个训',
                active: value == '个训',
                optionKey: optionKeyPrefix == null
                    ? null
                    : ValueKey<String>('$optionKeyPrefix-个训'),
                onTap: () => onChanged('个训'),
              ),
              _LessonOption(
                label: '集体课',
                active: value == '集体课',
                optionKey: optionKeyPrefix == null
                    ? null
                    : ValueKey<String>('$optionKeyPrefix-集体课'),
                onTap: () => onChanged('集体课'),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _LessonOption extends StatelessWidget {
  const _LessonOption({
    required this.label,
    required this.active,
    required this.onTap,
    this.optionKey,
  });

  final String label;
  final bool active;
  final VoidCallback onTap;
  final Key? optionKey;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Material(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(9),
        child: InkWell(
          key: optionKey,
          onTap: onTap,
          borderRadius: BorderRadius.circular(9),
          child: Container(
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: active ? _IepColors.orange : Colors.transparent,
              borderRadius: BorderRadius.circular(9),
            ),
            child: Text(
              label,
              style: TextStyle(
                color: active ? Colors.white : _IepColors.text,
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

class _GoalTextField extends StatelessWidget {
  const _GoalTextField({
    required this.label,
    required this.controller,
    required this.minLines,
    required this.maxLines,
    this.fieldKey,
    this.trailing,
  });

  final String label;
  final TextEditingController controller;
  final int minLines;
  final int maxLines;
  final Key? fieldKey;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Row(
          children: <Widget>[
            Text(
              label,
              style: const TextStyle(
                color: _IepColors.text,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
            const Spacer(),
            if (trailing != null) trailing!,
          ],
        ),
        const SizedBox(height: 7),
        TextField(
          key: fieldKey,
          controller: controller,
          minLines: minLines,
          maxLines: maxLines,
          style: const TextStyle(
            color: _IepColors.ink,
            fontSize: 14,
            fontWeight: FontWeight.w700,
            height: 1.35,
          ),
          decoration: InputDecoration(
            isDense: true,
            filled: true,
            fillColor: Colors.white,
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 12, vertical: 11),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: _IepColors.line),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide:
                  const BorderSide(color: _IepColors.orange, width: 1.2),
            ),
          ),
        ),
      ],
    );
  }
}

class _AddGoalButton extends StatelessWidget {
  const _AddGoalButton({
    required this.onTap,
    this.label = '新增一条',
  });

  final VoidCallback onTap;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(13),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Container(
          height: 32,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(13),
            border: Border.all(color: const Color(0xFFFFD8C6)),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(Icons.add_rounded,
                  size: 17, color: _IepColors.orangeDeep),
              const SizedBox(width: 4),
              Text(
                label,
                style: const TextStyle(
                  color: _IepColors.orangeDeep,
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

class _SmallIconAction extends StatelessWidget {
  const _SmallIconAction({
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
          width: 26,
          height: 26,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: enabled ? Colors.white : const Color(0xFFF8EEE6),
            shape: BoxShape.circle,
            border: Border.all(color: _IepColors.lightLine),
          ),
          child: Icon(
            icon,
            size: 16,
            color: enabled ? _IepColors.text : _IepColors.muted,
          ),
        ),
      ),
    );
  }
}

class _GoalSyncHint extends StatelessWidget {
  const _GoalSyncHint();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: _IepColors.orangeSoft,
        borderRadius: BorderRadius.circular(13),
        border: Border.all(color: const Color(0xFFFFD8C6)),
      ),
      child: const Row(
        children: <Widget>[
          Icon(Icons.info_outline_rounded,
              size: 16, color: _IepColors.orangeDeep),
          SizedBox(width: 7),
          Expanded(
            child: Text(
              '修改目标后，可选择仅更新当前IEP总表，也可以同步更新关联月计划、周计划。',
              style: TextStyle(
                color: _IepColors.text,
                fontSize: 12,
                fontWeight: FontWeight.w800,
                height: 1.35,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

enum _PlanTabTone { primary, week }

class _PlanTab extends StatelessWidget {
  const _PlanTab({
    required this.text,
    required this.width,
    this.active = false,
    this.activeTone = _PlanTabTone.primary,
    this.rightGap = 6,
    this.onTap,
  });

  final String text;
  final double width;
  final bool active;
  final _PlanTabTone activeTone;
  final double rightGap;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final Color activeBg =
        activeTone == _PlanTabTone.week ? _IepColors.orange : _IepColors.ink;
    final Color inactiveBg = activeTone == _PlanTabTone.week
        ? const Color(0xFFFFF3EC)
        : const Color(0xFFFFFAF6);
    final Color inactiveText = activeTone == _PlanTabTone.week
        ? _IepColors.orangeDeep
        : _IepColors.text;
    final Color borderColor = activeTone == _PlanTabTone.week
        ? const Color(0xFFFFD8C3)
        : _IepColors.lightLine;

    return Padding(
      padding: EdgeInsets.only(right: rightGap),
      child: Material(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(15),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(15),
          child: Container(
            width: width,
            height: 30,
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: active ? activeBg : inactiveBg,
              borderRadius: BorderRadius.circular(15),
              border: active ? null : Border.all(color: borderColor),
            ),
            child: Text(
              text,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: active ? Colors.white : inactiveText,
                fontSize: 11,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _PlanNavLabel extends StatelessWidget {
  const _PlanNavLabel({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 6, right: 6),
      child: Text(
        text,
        style: const TextStyle(
          color: _IepColors.muted,
          fontSize: 11,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _IepTablePreview extends StatelessWidget {
  const _IepTablePreview({
    required this.previewMode,
    required this.month,
    required this.weekNumber,
    required this.periodStart,
    required this.periodMonthCount,
    required this.record,
    required this.plan,
    required this.monthPlan,
    required this.weekPlan,
    required this.loading,
    required this.error,
    required this.onRetry,
    required this.totalPlanDomains,
    required this.selectedGoal,
    required this.onGoalTap,
    required this.onClearSelectedGoal,
  });

  final _IepPreviewMode previewMode;
  final String month;
  final int weekNumber;
  final DateTime periodStart;
  final int periodMonthCount;
  final IepAssessmentRecordSummary? record;
  final IepPlan? plan;
  final IepMonthlyPlan? monthPlan;
  final IepWeeklyPlan? weekPlan;
  final bool loading;
  final String error;
  final VoidCallback onRetry;
  final List<_DocDomainData> totalPlanDomains;
  final _GoalEditRequest? selectedGoal;
  final ValueChanged<_GoalEditRequest> onGoalTap;
  final VoidCallback onClearSelectedGoal;

  @override
  Widget build(BuildContext context) {
    final DateTime monthDate = _monthDateFromLabel(
      periodStart,
      periodMonthCount,
      month,
    );
    final DateTimeRange monthRange = _monthRangeInPeriod(
      periodStart: periodStart,
      monthCount: periodMonthCount,
      monthDate: monthDate,
    );
    final List<DateTime> weekDates = _weekDatesInMonthRange(
      monthRange,
      weekNumber,
    );

    Widget child;
    if (record == null) {
      child = const _PlanStateView(
        icon: Icons.touch_app_rounded,
        title: '请选择左侧评估记录',
        message: '选择学员后会读取对应IEP计划',
      );
    } else if (loading) {
      child = const _PlanStateView(
        icon: Icons.hourglass_top_rounded,
        title: '正在加载IEP计划',
      );
    } else if (error.trim().isNotEmpty) {
      child = _PlanStateView(
        icon: Icons.wifi_off_rounded,
        title: 'IEP计划加载失败',
        message: error,
        actionLabel: '重试',
        onAction: onRetry,
      );
    } else if (plan == null || totalPlanDomains.isEmpty) {
      child = const _PlanStateView(
        icon: Icons.assignment_outlined,
        title: '暂无已生成IEP',
        message: '确认IEP后会在这里展示总计划、月计划和周计划',
      );
    } else {
      child = switch (previewMode) {
        _IepPreviewMode.month => monthPlan == null
            ? _PlanStateView(
                icon: Icons.calendar_month_rounded,
                title: '$month计划未生成',
                message: plan?.title ?? 'IEP总计划',
              )
            : _WordTableFrame(
                child: _MonthPlanTable(
                  month: month,
                  monthRange: monthRange,
                  plan: monthPlan,
                ),
              ),
        _IepPreviewMode.week => weekPlan == null
            ? _PlanStateView(
                icon: Icons.view_week_rounded,
                title: '$month W$weekNumber 周计划未生成',
                message: plan?.title ?? 'IEP总计划',
              )
            : _WordTableFrame(
                child: _WeekPlanTable(
                  month: month,
                  weekNumber: weekNumber,
                  weekDates: weekDates,
                  plan: weekPlan,
                ),
              ),
        _IepPreviewMode.total => _WordTableFrame(
            child: _WordTable(
              periodText: _formatZhRange(
                periodStart,
                _periodEndFor(periodStart, periodMonthCount),
              ),
              plan: plan,
              domains: totalPlanDomains,
              selectedGoal: selectedGoal,
              onGoalTap: onGoalTap,
              onClearSelectedGoal: onClearSelectedGoal,
            ),
            height: _WordTable.heightFor(totalPlanDomains),
          ),
      };
    }

    return Container(
      padding: const EdgeInsets.fromLTRB(14, 14, 14, 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _IepColors.line),
      ),
      child: child,
    );
  }
}

class _PlanStateView extends StatelessWidget {
  const _PlanStateView({
    required this.icon,
    required this.title,
    this.message = '',
    this.actionLabel = '',
    this.onAction,
  });

  final IconData icon;
  final String title;
  final String message;
  final String actionLabel;
  final VoidCallback? onAction;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        width: 340,
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 22),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: _IepColors.lightLine),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Icon(icon, size: 34, color: _IepColors.orangeDeep),
            const SizedBox(height: 10),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _IepColors.ink,
                fontSize: 15,
                fontWeight: FontWeight.w900,
              ),
            ),
            if (message.trim().isNotEmpty) ...<Widget>[
              const SizedBox(height: 7),
              Text(
                message,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: _IepColors.text,
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                  height: 1.35,
                ),
              ),
            ],
            if (actionLabel.trim().isNotEmpty && onAction != null) ...<Widget>[
              const SizedBox(height: 12),
              _MiniQueueAction(label: actionLabel, onTap: onAction!),
            ],
          ],
        ),
      ),
    );
  }
}

class _WordTable extends StatelessWidget {
  const _WordTable({
    required this.periodText,
    required this.plan,
    required this.domains,
    required this.selectedGoal,
    required this.onGoalTap,
    required this.onClearSelectedGoal,
  });

  final String periodText;
  final IepPlan? plan;
  final List<_DocDomainData> domains;
  final _GoalEditRequest? selectedGoal;
  final ValueChanged<_GoalEditRequest> onGoalTap;
  final VoidCallback onClearSelectedGoal;

  static const List<int> _columns = <int>[
    1038,
    1472,
    625,
    877,
    1260,
    1562,
    927,
    2319,
  ];
  static const double _minHeight = 820;
  static const double _headerHeight = 208;

  static double heightFor(List<_DocDomainData> domains) {
    final double contentHeight =
        _headerHeight + _DocPlanRows.heightFor(domains);
    return contentHeight > _minHeight ? contentHeight : _minHeight;
  }

  @override
  Widget build(BuildContext context) {
    final IepPlan? currentPlan = plan;
    return GestureDetector(
      behavior: HitTestBehavior.translucent,
      onTap: onClearSelectedGoal,
      child: Column(
        children: <Widget>[
          _WordTableTitle(title: currentPlan?.title ?? '康复教学季度计划'),
          _DocTableRow(
            height: 42,
            cells: <_DocCellData>[
              _DocCellData(text: '姓名', columns: 1, bold: true),
              _DocCellData(text: currentPlan?.student.name ?? '-', columns: 1),
              _DocCellData(text: '性别', columns: 1, bold: true),
              _DocCellData(
                  text: currentPlan?.student.gender ?? '-', columns: 1),
              _DocCellData(text: '出生年月', columns: 1, bold: true),
              _DocCellData(
                text: currentPlan?.student.birthDate ?? '-',
                columns: 3,
                last: true,
              ),
            ],
          ),
          _DocTableRow(
            height: 42,
            cells: <_DocCellData>[
              _DocCellData(text: '制定日期', columns: 1, bold: true),
              _DocCellData(text: currentPlan?.meta.planDate ?? '-', columns: 3),
              _DocCellData(text: '计划参与者', columns: 1, bold: true),
              _DocCellData(
                text: currentPlan?.meta.participant ?? '-',
                columns: 3,
                last: true,
              ),
            ],
          ),
          _DocTableRow(
            height: 42,
            cells: <_DocCellData>[
              _DocCellData(text: '实施者', columns: 1, bold: true),
              _DocCellData(
                  text: currentPlan?.meta.implementer ?? '-', columns: 3),
              _DocCellData(text: '实施\n起止日期', columns: 1, bold: true),
              _DocCellData(
                text: _metaRangeText(currentPlan?.meta, fallback: periodText),
                columns: 3,
                noWrap: true,
                last: true,
              ),
            ],
          ),
          _DocTableRow(
            height: 42,
            cells: <_DocCellData>[
              _DocCellData(text: '康复\n领域', columns: 1, bold: true),
              _DocCellData(text: '长期目标', columns: 3, bold: true),
              _DocCellData(text: '短期目标', columns: 2, bold: true),
              _DocCellData(text: '课程\n形式', columns: 1, bold: true),
              _DocCellData(text: '起止日期', columns: 1, bold: true, last: true),
            ],
          ),
          SizedBox(
            height: _DocPlanRows.heightFor(domains),
            child: _DocPlanRows(
              domains: domains,
              selectedGoal: selectedGoal,
              onGoalTap: onGoalTap,
            ),
          ),
        ],
      ),
    );
  }
}

class _WordTableFrame extends StatelessWidget {
  const _WordTableFrame({
    required this.child,
    this.height,
  });

  final Widget child;
  final double? height;

  @override
  Widget build(BuildContext context) {
    return Container(
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      foregroundDecoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: const Color(0xFFB98A71), width: 1.2),
      ),
      child: Padding(
        padding: const EdgeInsets.all(1.2),
        child: SingleChildScrollView(
          physics: const ClampingScrollPhysics(),
          child:
              height == null ? child : SizedBox(height: height, child: child),
        ),
      ),
    );
  }
}

class _WeekPlanTable extends StatelessWidget {
  const _WeekPlanTable({
    required this.month,
    required this.weekNumber,
    required this.weekDates,
    required this.plan,
  });

  final String month;
  final int weekNumber;
  final List<DateTime> weekDates;
  final IepWeeklyPlan? plan;

  static const List<int> _columns = <int>[
    1300,
    1334,
    1333,
    1925,
    698,
    698,
    698,
    698,
    698,
    698,
  ];

  static const List<_WeekTrainingRow> _fallbackRows = <_WeekTrainingRow>[
    _WeekTrainingRow(
      project: '平衡木行走',
      content:
          '在感统室设置宽10cm、高20cm的平衡木，儿童独立行走3米，治疗师在旁保护但不接触，完成后给予口头表扬和代币强化，每日练习2次，记录掉落次数，目标连续3次不掉落。',
    ),
    _WeekTrainingRow(
      project: '直线剪纸',
      content:
          '使用彩色纸画有直线（线宽0.5cm），儿童独立剪10cm，治疗师用尺子测量偏差，偏差在0.5cm内给予代币，累计代币兑换偏好活动。',
    ),
    _WeekTrainingRow(
      project: '情绪指认',
      content:
          '在绘本阅读中，治疗师指向角色表情，问“他感觉怎么样？”，儿童从4张情绪卡片（高兴、生气、伤心、害怕）中选择对应卡片，正确率100%后，儿童尝试口头命名情绪。',
    ),
    _WeekTrainingRow(
      project: '动作序列模仿',
      content: '变换动作序列（如“拍肩-转圈-跳”），治疗师示范后儿童模仿，顺序正确率80%以上，逐渐撤除视觉提示，仅靠观察模仿。',
    ),
    _WeekTrainingRow(
      project: '主动邀请同伴',
      content:
          '使用社交故事《邀请朋友玩》，课前阅读，课上治疗师提示儿童使用邀请语“我们一起玩吧”，当儿童主动邀请时，同伴积极回应，形成自然强化，记录邀请次数，目标每节至少2次。',
    ),
    _WeekTrainingRow(
      project: '变换问答',
      content:
          '使用视觉提示卡“说不同的话”，当儿童在对话中重复同一回答时，治疗师出示卡片并等待3秒，儿童变换回答后立即表扬，逐渐减少提示，记录重复次数。',
    ),
    _WeekTrainingRow(
      project: '减少摇晃行为',
      content:
          '使用区别强化，当儿童保持安静坐好2分钟无摇晃，给予代币，累计代币兑换偏好活动；摇晃行为发生时，不给予关注，仅重新引导，目标每节课不超过2次。',
    ),
    _WeekTrainingRow(
      project: '多属性分类',
      content:
          '使用属性卡片（红色、圆形、大），儿童根据指令将物品放入对应盒子，如“把红色的放一起”，逐渐增加复杂度，同时按两个属性分类（如红色且圆形），正确后给予代币，正确率90%以上。',
    ),
    _WeekTrainingRow(
      project: '两步指令执行',
      content:
          '使用图片提示卡（拍手、摸头），治疗师发两步指令“先拍手再摸头”时同时出示图片，儿童执行，逐渐撤除图片，仅靠听觉理解，正确率90%以上。',
    ),
    _WeekTrainingRow(
      project: '完整句子描述',
      content:
          '提供缺少主语的图片，治疗师问“谁在做什么？”，儿童需补充完整句子，如“妈妈在做饭”，逐渐增加句子成分（加入地点“在厨房”），正确结构后给予代币。',
    ),
  ];

  @override
  Widget build(BuildContext context) {
    final IepWeeklyPlan? currentPlan = plan;
    final List<_WeekTrainingRow> rows = currentPlan?.rows
            .map(_weekTrainingRowFromPlanRow)
            .where((_WeekTrainingRow row) =>
                row.project.trim().isNotEmpty || row.content.trim().isNotEmpty)
            .toList() ??
        _fallbackRows;
    final List<DateTime> displayWeekDates =
        _dateListFromStrings(currentPlan?.weekDates) ?? weekDates;
    return Column(
      children: <Widget>[
        _WordTableTitle(
          title: currentPlan?.title.trim().isNotEmpty == true
              ? currentPlan!.title
              : '康复教学周计划日记录卡$month第$weekNumber周',
        ),
        _WeekInfoRows(
          plan: currentPlan,
          periodText: currentPlan?.trainingDate.trim().isNotEmpty == true
              ? currentPlan!.trainingDate
              : _weekRangeText(displayWeekDates),
        ),
        _WeekHeaderRows(dates: displayWeekDates),
        _WeekTrainingRows(rows: rows),
      ],
    );
  }
}

class _WeekInfoRows extends StatelessWidget {
  const _WeekInfoRows({required this.periodText, required this.plan});

  final String periodText;
  final IepWeeklyPlan? plan;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        _WeekDocTableRow(
          height: 42,
          cells: <_WeekDocCellData>[
            _WeekDocCellData(text: '姓名', columns: 1, bold: true),
            _WeekDocCellData(text: plan?.student.name ?? '-', columns: 1),
            _WeekDocCellData(text: '性别', columns: 1, bold: true),
            _WeekDocCellData(text: plan?.student.gender ?? '-', columns: 1),
            _WeekDocCellData(text: '出生年月', columns: 2, bold: true),
            _WeekDocCellData(
              text: plan?.student.birthDate ?? '-',
              columns: 4,
              last: true,
            ),
          ],
        ),
        _WeekDocTableRow(
          height: 42,
          cells: <_WeekDocCellData>[
            const _WeekDocCellData(text: '任教\n老师', columns: 1, bold: true),
            _WeekDocCellData(
              text: plan?.teacherName.trim().isNotEmpty == true
                  ? plan!.teacherName
                  : '-',
              columns: 1,
            ),
            const _WeekDocCellData(text: '课程\n名称', columns: 1, bold: true),
            _WeekDocCellData(
              text: plan?.courseName.trim().isNotEmpty == true
                  ? plan!.courseName
                  : '康复教学',
              columns: 1,
            ),
            const _WeekDocCellData(text: '训练日期', columns: 2, bold: true),
            _WeekDocCellData(
              text: periodText,
              columns: 4,
              last: true,
            ),
          ],
        ),
        _WeekDocTableRow(
          height: 54,
          cells: <_WeekDocCellData>[
            _WeekDocCellData(text: '训练前\n准备', columns: 1, bold: true),
            _WeekDocCellData(
              text: plan?.preparation.trim().isNotEmpty == true
                  ? plan!.preparation
                  : '训练材料、视觉提示卡、强化物、记录表',
              columns: 9,
              align: TextAlign.left,
              last: true,
            ),
          ],
        ),
      ],
    );
  }
}

class _WeekHeaderRows extends StatelessWidget {
  const _WeekHeaderRows({required this.dates});

  final List<DateTime> dates;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 70,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          const Expanded(
            flex: 1300,
            child: _WeekDocCellBox(
              data: _WeekDocCellData(text: '训练项目', columns: 1, bold: true),
            ),
          ),
          const Expanded(
            flex: 4583,
            child: _WeekDocCellBox(
              data: _WeekDocCellData(text: '训练内容', columns: 3, bold: true),
            ),
          ),
          Expanded(
            flex: 4188,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: <Widget>[
                SizedBox(
                  height: 36,
                  child: const _WeekDocCellBox(
                    data: _WeekDocCellData(
                      text: '完成情况',
                      columns: 6,
                      bold: true,
                      last: true,
                    ),
                  ),
                ),
                SizedBox(
                  height: 34,
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: List<Widget>.generate(6, (int index) {
                      final String label = index < dates.length
                          ? _weekDateLabel(dates[index])
                          : '';
                      return Expanded(
                        child: _WeekDocCellBox(
                          data: _WeekDocCellData(
                            text: label,
                            columns: 1,
                            bold: true,
                            last: index == 5,
                          ),
                        ),
                      );
                    }),
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

class _WeekDocCellData {
  const _WeekDocCellData({
    required this.text,
    required this.columns,
    this.bold = false,
    this.align = TextAlign.center,
    this.last = false,
  });

  final String text;
  final int columns;
  final bool bold;
  final TextAlign align;
  final bool last;
}

class _WeekDocTableRow extends StatelessWidget {
  const _WeekDocTableRow({required this.height, required this.cells});

  final double height;
  final List<_WeekDocCellData> cells;

  @override
  Widget build(BuildContext context) {
    final List<Widget> children = <Widget>[];
    int columnIndex = 0;
    for (final _WeekDocCellData cell in cells) {
      final int flex = _WeekPlanTable._columns
          .skip(columnIndex)
          .take(cell.columns)
          .fold<int>(0, (int sum, int width) => sum + width);
      columnIndex += cell.columns;
      children.add(
        Expanded(
          flex: flex,
          child: _WeekDocCellBox(data: cell),
        ),
      );
    }

    return SizedBox(
      height: height,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: children,
      ),
    );
  }
}

class _WeekDocCellBox extends StatelessWidget {
  const _WeekDocCellBox({
    required this.data,
    this.rowLast = false,
    this.verticalPadding = 5,
  });

  final _WeekDocCellData data;
  final bool rowLast;
  final double verticalPadding;

  @override
  Widget build(BuildContext context) {
    return Container(
      alignment: Alignment.center,
      padding: EdgeInsets.symmetric(horizontal: 7, vertical: verticalPadding),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(
          right: data.last
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
          bottom: rowLast
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
        ),
      ),
      child: Text(
        data.text,
        overflow: TextOverflow.clip,
        textAlign: data.align,
        style: TextStyle(
          color: data.bold ? _IepColors.ink : _IepColors.text,
          fontSize: 10.8,
          fontWeight: data.bold ? FontWeight.w900 : FontWeight.w700,
          height: 1.18,
        ),
      ),
    );
  }
}

class _WeekTrainingRows extends StatelessWidget {
  const _WeekTrainingRows({required this.rows});

  final List<_WeekTrainingRow> rows;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: rows.asMap().entries.map((entry) {
        return _WeekTrainingTableRow(
          row: entry.value,
          last: entry.key == rows.length - 1,
        );
      }).toList(),
    );
  }
}

class _WeekTrainingRow {
  const _WeekTrainingRow({
    required this.project,
    required this.content,
  });

  final String project;
  final String content;

  double get rowHeight {
    if (content.length >= 85) {
      return 62;
    }
    if (content.length >= 70) {
      return 54;
    }
    return 46;
  }
}

class _WeekTrainingTableRow extends StatelessWidget {
  const _WeekTrainingTableRow({
    required this.row,
    required this.last,
  });

  final _WeekTrainingRow row;
  final bool last;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: row.rowHeight,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Expanded(
            flex: _WeekPlanTable._columns[0],
            child: _WeekDocCellBox(
              data: _WeekDocCellData(text: row.project, columns: 1, bold: true),
              rowLast: last,
              verticalPadding: 3,
            ),
          ),
          Expanded(
            flex: _WeekPlanTable._columns[1] +
                _WeekPlanTable._columns[2] +
                _WeekPlanTable._columns[3],
            child: _WeekDocCellBox(
              data: _WeekDocCellData(
                text: row.content,
                columns: 3,
                align: TextAlign.left,
              ),
              rowLast: last,
              verticalPadding: 3,
            ),
          ),
          ...List<Widget>.generate(6, (int index) {
            return Expanded(
              flex: _WeekPlanTable._columns[index + 4],
              child: _WeekDocCellBox(
                data: _WeekDocCellData(
                  text: '',
                  columns: 1,
                  last: index == 5,
                ),
                rowLast: last,
                verticalPadding: 3,
              ),
            );
          }),
        ],
      ),
    );
  }
}

class _MonthPlanTable extends StatelessWidget {
  const _MonthPlanTable({
    required this.month,
    required this.monthRange,
    required this.plan,
  });

  final String month;
  final DateTimeRange monthRange;
  final IepMonthlyPlan? plan;

  static const List<int> _columns = <int>[
    806,
    907,
    907,
    958,
    957,
    882,
    882,
    882,
    882,
    826,
    595,
    595,
  ];

  static const List<_MonthDomainData> _fallbackDomains = <_MonthDomainData>[
    _MonthDomainData(
      domain: '大肌肉',
      longGoal: '1. 提升动态平衡与协调能力，能独立完成单脚站立、走平衡木等动作\n2. 增强下肢力量与跳跃技能，能连续向前跳跃并保持稳定',
      shortGoal: '能在宽10cm、高20cm的平衡木上独立行走3米，不掉落',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 使用宽10cm、高20cm的平衡木，治疗师先示范双手侧平举保持平衡，儿童在平地上沿直线行走练习，逐步过渡到平衡木上，初始可单手扶墙辅助，逐渐撤除辅助，完成3米行走不掉落。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 在感统室设置平衡木，儿童独立行走3米，治疗师在旁保护但不接触，完成后给予口头表扬和代币强化，每日练习2次，记录掉落次数，目标连续3次不掉落。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至户外低矮花坛边缘（约10cm宽），儿童独立行走3米，治疗师在旁监护，鼓励儿童在不同材质上保持平衡，完成后奖励贴纸。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '小肌肉',
      longGoal: '1. 提高手眼协调与精细操作能力，能熟练使用剪刀、穿珠子等\n2. 增强手部小肌肉控制，能完成复杂拼图与书写前准备',
      shortGoal: '能沿直线剪纸，偏差不超过0.5cm，连续剪10cm',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 提供儿童安全剪刀和画有粗直线的纸条（宽2cm），治疗师手把手辅助儿童开合剪刀，沿直线剪，逐步减少辅助，要求偏差不超过0.5cm，剪完10cm。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用彩色纸画有直线（线宽0.5cm），儿童独立剪10cm，治疗师用尺子测量偏差，偏差在0.5cm内给予代币，累计代币兑换偏好活动。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至剪不同材质（如卡纸、杂志页），儿童沿直线剪10cm，偏差不超过0.5cm，完成后将剪下的纸条用于粘贴画，增加趣味性。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '情感表达',
      longGoal: '1. 能识别并命名基本情绪，理解他人情绪并做出恰当反应\n2. 在情境中表达自己的情绪，并学习简单的情绪调节策略',
      shortGoal: '能指认高兴、生气、伤心、害怕四种情绪图片，正确率100%',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 使用情绪卡片（高兴、生气、伤心、害怕各2张），治疗师呈现卡片并命名，儿童指认，每次4选1，正确后给予社会性强化，错误时示范正确卡片，连续2次正确率100%进入下一阶段。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 在绘本阅读中，治疗师指向角色表情，问“他感觉怎么样？”，儿童从4张情绪卡片中选择对应卡片，正确率100%后，儿童尝试口头命名情绪。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至真实情境，当同伴或家人表现出情绪时，治疗师提示儿童观察并指认情绪图片，正确后给予自然强化（如“你看到妹妹哭了，知道她伤心，真棒！”）。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '模仿 （视觉/动作）',
      longGoal: '1. 提升动作模仿的准确性和复杂性，能模仿多步骤动作序列\n2. 增强视觉注意与模仿的泛化能力，能在不同情境下模仿他人',
      shortGoal: '能模仿3个连续的动作序列（如拍手-摸头-跺脚），顺序正确',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 治疗师示范“拍手-摸头-跺脚”序列，边说边做，儿童模仿，初始可分解教学，先模仿单个动作，再串联，使用视觉提示卡辅助顺序，正确后给予代币。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 变换动作序列（如“拍肩-转圈-跳”），治疗师示范后儿童模仿，顺序正确率80%以上，逐渐撤除视觉提示，仅靠观察模仿。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 在集体课中泛化，治疗师带领小组做动作序列，儿童跟随模仿，顺序正确后担任小老师带领其他儿童，增强动机。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '社交互动',
      longGoal: '1. 提高与同伴的互动能力，能主动发起并维持简单的社交游戏\n2. 理解并遵守基本社交规则，如轮流、分享、等待',
      shortGoal: '在小组活动中，能主动邀请同伴一起玩，至少2次/节',
      lesson: '集体课',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 在集体课自由游戏时间，治疗师设置合作性玩具（如积木、拼图），示范邀请语言“我们一起玩吧”，儿童模仿邀请同伴，每成功邀请1次给予贴纸，目标2次/节。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用社交故事《邀请朋友玩》，课前阅读，课上治疗师提示儿童使用邀请语，当儿童主动邀请时，同伴积极回应，形成自然强化，记录邀请次数。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至户外活动，儿童在滑梯或沙池主动邀请同伴，治疗师在旁观察，必要时给予手势提示，每节至少2次主动邀请，完成后奖励额外自由时间。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '行为特征 -语言',
      longGoal: '1. 减少刻板语言，增加功能性语言的灵活运用\n2. 提高语言在社交情境中的恰当性，能根据对象调整语言',
      shortGoal: '在对话中，能根据对方的问题变换回答，减少重复同一句话',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 个训中，治疗师与儿童进行简单问答（如“你喜欢什么颜色？”），若儿童重复同一回答，治疗师示范不同回答并提示“换一种说法”，正确变换后给予强化。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用视觉提示卡“说不同的话”，当儿童重复时，治疗师出示卡片并等待3秒，儿童变换回答后立即表扬，逐渐减少提示，记录重复次数。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至与同伴对话，设置情境（如分享玩具），同伴问“我可以玩吗？”，儿童需根据情境回答（如“可以”或“等一下”），而非固定回答，治疗师在旁辅助。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '行为特征 -非语言',
      longGoal: '1. 减少自我刺激行为，增加功能性非语言沟通\n2. 提高对环境变化的适应能力，减少刻板行为',
      shortGoal: '在课堂上，无意义的摇晃身体行为减少至每节课不超过2次',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 个训中，治疗师观察摇晃行为，当行为出现时，立即提供替代感觉输入（如挤压球、坐垫），并口头提醒“坐好”，记录频率，目标每节课不超过2次。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用区别强化，当儿童保持安静坐好2分钟无摇晃，给予代币，累计代币兑换偏好活动；摇晃行为发生时，不给予关注，仅重新引导。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至集体课，治疗师与主教合作，在集体活动中监控摇晃行为，使用视觉提示卡“安静身体”提醒，每节课摇晃不超过2次，达成后给予集体奖励。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '认知 （语言/语前）',
      longGoal: '1. 提升分类与排序能力，能按多种属性进行归类\n2. 增强问题解决与逻辑思维能力，能完成简单的推理任务',
      shortGoal: '能按颜色、形状、大小三个属性对物品进行分类，正确率90%',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 提供不同颜色、形状、大小的积木，治疗师先示范按颜色分类，儿童模仿，然后按形状分类，最后按大小分类，每次分类后提问“为什么放在一起？”，正确率90%以上。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用属性卡片（红色、圆形、大），儿童根据指令将物品放入对应盒子，如“把红色的放一起”，逐渐增加复杂度，同时按两个属性分类（如红色且圆形），正确后给予代币。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至生活场景，整理玩具时，儿童按颜色或类型将玩具放回架子，治疗师在旁提示，正确率90%后自然强化（环境整洁）。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '语言理解',
      longGoal: '1. 提高对复杂指令的理解，能执行两步以上指令\n2. 增强对故事和对话的理解，能回答相关问题',
      shortGoal: '能执行包含两个步骤的指令（如“先拍手再摸头”），正确率90%',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 个训中，治疗师发出两步指令“先拍手再摸头”，初始可示范，儿童模仿，然后仅用语言指令，儿童执行，正确后给予强化，错误时退回一步分解。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用图片提示卡（拍手、摸头），治疗师发指令时同时出示图片，儿童执行，逐渐撤除图片，仅靠听觉理解，正确率90%以上。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至集体课，在音乐活动中，治疗师唱出两步指令“先跺脚再拍手”，儿童跟随，正确后担任小指挥，增加趣味性。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '语言表达',
      longGoal: '1. 扩展句子长度与复杂性，能使用完整句子描述事件\n2. 提高叙事能力，能连贯讲述个人经历或故事',
      shortGoal: '能使用“主语+谓语+宾语”结构描述图片，如“男孩吃苹果”',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 使用动作图片卡（如男孩吃苹果、女孩拍球），治疗师示范“男孩吃苹果”，儿童模仿，然后出示新图片，儿童独立描述，正确结构后给予代币。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 提供缺少主语的图片，治疗师问“谁在做什么？”，儿童需补充完整句子，如“妈妈在做饭”，逐渐增加句子成分（加入地点“在厨房”）。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至绘本阅读，儿童描述书中画面，治疗师提示使用完整句子，如“小狗在睡觉”，正确后自然强化（继续讲故事）。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
  ];

  @override
  Widget build(BuildContext context) {
    final IepMonthlyPlan? currentPlan = plan;
    final List<_MonthDomainData> domains = currentPlan?.rows
            .map(_monthDomainFromPlanRow)
            .where((_MonthDomainData item) =>
                item.domain.trim().isNotEmpty ||
                item.shortGoal.trim().isNotEmpty ||
                item.trainings.any((_MonthTrainingData training) =>
                    training.content.trim().isNotEmpty))
            .toList() ??
        _fallbackDomains;
    return Column(
      children: <Widget>[
        _WordTableTitle(
          title: currentPlan?.title.trim().isNotEmpty == true
              ? currentPlan!.title
              : '康复教学$month计划',
        ),
        _MonthInfoRows(
          plan: currentPlan,
          periodText: _metaRangeText(
            currentPlan?.meta,
            fallback: _formatZhRange(monthRange.start, monthRange.end),
          ),
        ),
        _MonthPlanRows(domains: domains, monthRange: monthRange),
      ],
    );
  }
}

class _MonthInfoRows extends StatelessWidget {
  const _MonthInfoRows({required this.periodText, required this.plan});

  final String periodText;
  final IepMonthlyPlan? plan;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        _MonthDocTableRow(
          height: 42,
          cells: <_MonthDocCellData>[
            _MonthDocCellData(text: '姓名', columns: 1, bold: true),
            _MonthDocCellData(text: plan?.student.name ?? '-', columns: 2),
            _MonthDocCellData(text: '性别', columns: 1, bold: true),
            _MonthDocCellData(text: plan?.student.gender ?? '-', columns: 1),
            _MonthDocCellData(text: '出生年月', columns: 2, bold: true),
            _MonthDocCellData(
              text: plan?.student.birthDate ?? '-',
              columns: 5,
              last: true,
            ),
          ],
        ),
        _MonthDocTableRow(
          height: 42,
          cells: <_MonthDocCellData>[
            const _MonthDocCellData(text: '制定\n日期', columns: 1, bold: true),
            _MonthDocCellData(text: plan?.meta.planDate ?? '-', columns: 2),
            const _MonthDocCellData(text: '计划参与者', columns: 4, bold: true),
            _MonthDocCellData(
              text: plan?.meta.participant ?? '-',
              columns: 5,
              last: true,
            ),
          ],
        ),
        _MonthDocTableRow(
          height: 42,
          cells: <_MonthDocCellData>[
            const _MonthDocCellData(text: '实施者', columns: 1, bold: true),
            _MonthDocCellData(text: plan?.meta.implementer ?? '-', columns: 2),
            const _MonthDocCellData(text: '实施起止日期', columns: 4, bold: true),
            _MonthDocCellData(
              text: periodText,
              columns: 5,
              noWrap: true,
              last: true,
            ),
          ],
        ),
        _MonthDocTableRow(
          height: 42,
          cells: const <_MonthDocCellData>[
            _MonthDocCellData(text: '康复\n领域', columns: 1, bold: true),
            _MonthDocCellData(text: '长期目标', columns: 2, bold: true),
            _MonthDocCellData(text: '短期目标', columns: 2, bold: true),
            _MonthDocCellData(text: '训练内容', columns: 4, bold: true),
            _MonthDocCellData(text: '课程\n形式', columns: 1, bold: true),
            _MonthDocCellData(text: '起止日期', columns: 2, bold: true, last: true),
          ],
        ),
      ],
    );
  }
}

class _MonthDocCellData {
  const _MonthDocCellData({
    required this.text,
    required this.columns,
    this.bold = false,
    this.align = TextAlign.center,
    this.last = false,
    this.noWrap = false,
  });

  final String text;
  final int columns;
  final bool bold;
  final TextAlign align;
  final bool last;
  final bool noWrap;
}

class _MonthDocTableRow extends StatelessWidget {
  const _MonthDocTableRow({required this.height, required this.cells});

  final double height;
  final List<_MonthDocCellData> cells;

  @override
  Widget build(BuildContext context) {
    final List<Widget> children = <Widget>[];
    int columnIndex = 0;
    for (final _MonthDocCellData cell in cells) {
      final int flex = _MonthPlanTable._columns
          .skip(columnIndex)
          .take(cell.columns)
          .fold<int>(0, (int sum, int width) => sum + width);
      columnIndex += cell.columns;
      children.add(
        Expanded(
          flex: flex,
          child: _MonthDocCellBox(data: cell),
        ),
      );
    }

    return SizedBox(
      height: height,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: children,
      ),
    );
  }
}

class _MonthDocCellBox extends StatelessWidget {
  const _MonthDocCellBox({
    required this.data,
    this.rowLast = false,
    this.verticalPadding = 5,
  });

  final _MonthDocCellData data;
  final bool rowLast;
  final double verticalPadding;

  @override
  Widget build(BuildContext context) {
    return Container(
      alignment: Alignment.center,
      padding: EdgeInsets.symmetric(horizontal: 7, vertical: verticalPadding),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(
          right: data.last
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
          bottom: rowLast
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
        ),
      ),
      child: Text(
        data.text,
        maxLines: data.noWrap ? 1 : null,
        overflow: data.noWrap ? TextOverflow.ellipsis : TextOverflow.clip,
        textAlign: data.align,
        style: TextStyle(
          color: data.bold ? _IepColors.ink : _IepColors.text,
          fontSize: 10.6,
          fontWeight: data.bold ? FontWeight.w900 : FontWeight.w700,
          height: 1.18,
        ),
      ),
    );
  }
}

class _MonthPlanRows extends StatelessWidget {
  const _MonthPlanRows({
    required this.domains,
    required this.monthRange,
  });

  final List<_MonthDomainData> domains;
  final DateTimeRange monthRange;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: domains.asMap().entries.map((entry) {
        return _MonthDomainBlock(
          data: entry.value,
          last: entry.key == domains.length - 1,
          monthRange: monthRange,
        );
      }).toList(),
    );
  }
}

class _MonthDomainData {
  const _MonthDomainData({
    required this.domain,
    required this.longGoal,
    required this.shortGoal,
    required this.lesson,
    required this.trainings,
  });

  final String domain;
  final String longGoal;
  final String shortGoal;
  final String lesson;
  final List<_MonthTrainingData> trainings;
}

class _MonthTrainingData {
  const _MonthTrainingData(this.content, this.period);

  final String content;
  final String period;

  String get displayPeriod => period.replaceFirst(' - ', '\n至 ');

  double get rowHeight {
    if (content.length >= 82) {
      return 70;
    }
    if (content.length >= 68) {
      return 60;
    }
    return 50;
  }
}

class _MonthDomainBlock extends StatelessWidget {
  const _MonthDomainBlock({
    required this.data,
    required this.last,
    required this.monthRange,
  });

  final _MonthDomainData data;
  final bool last;
  final DateTimeRange monthRange;

  @override
  Widget build(BuildContext context) {
    final List<double> rowHeights =
        data.trainings.map((training) => training.rowHeight).toList();
    final double blockHeight =
        rowHeights.fold<double>(0, (double sum, double height) => sum + height);

    return SizedBox(
      height: blockHeight,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Expanded(
            flex: _MonthPlanTable._columns[0],
            child: _MonthDocCellBox(
              data:
                  _MonthDocCellData(text: data.domain, columns: 1, bold: true),
              rowLast: last,
            ),
          ),
          Expanded(
            flex: _MonthPlanTable._columns[1] + _MonthPlanTable._columns[2],
            child: _MonthDocCellBox(
              data: _MonthDocCellData(
                text: data.longGoal,
                columns: 2,
                align: TextAlign.left,
              ),
              rowLast: last,
            ),
          ),
          Expanded(
            flex: _MonthPlanTable._columns[3] + _MonthPlanTable._columns[4],
            child: _MonthDocCellBox(
              data: _MonthDocCellData(
                text: data.shortGoal,
                columns: 2,
                align: TextAlign.left,
              ),
              rowLast: last,
            ),
          ),
          Expanded(
            flex: _MonthPlanTable._columns[5] +
                _MonthPlanTable._columns[6] +
                _MonthPlanTable._columns[7] +
                _MonthPlanTable._columns[8],
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: data.trainings.asMap().entries.map((entry) {
                return SizedBox(
                  height: rowHeights[entry.key],
                  child: _MonthDocCellBox(
                    data: _MonthDocCellData(
                      text: entry.value.content,
                      columns: 4,
                      align: TextAlign.left,
                    ),
                    rowLast: last && entry.key == data.trainings.length - 1,
                    verticalPadding: 3,
                  ),
                );
              }).toList(),
            ),
          ),
          Expanded(
            flex: _MonthPlanTable._columns[9],
            child: _MonthDocCellBox(
              data: _MonthDocCellData(
                text: data.lesson,
                columns: 1,
                noWrap: true,
              ),
              rowLast: last,
            ),
          ),
          Expanded(
            flex: _MonthPlanTable._columns[10] + _MonthPlanTable._columns[11],
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: data.trainings.asMap().entries.map((entry) {
                final String periodText = _monthTrainingPeriodText(
                  monthRange,
                  entry.key,
                );
                return SizedBox(
                  height: rowHeights[entry.key],
                  child: _MonthDocCellBox(
                    data: _MonthDocCellData(
                      text: periodText,
                      columns: 2,
                      last: true,
                    ),
                    rowLast: last && entry.key == data.trainings.length - 1,
                    verticalPadding: 3,
                  ),
                );
              }).toList(),
            ),
          ),
        ],
      ),
    );
  }
}

class _WordTableTitle extends StatelessWidget {
  const _WordTableTitle({this.title = '康复教学季度计划'});

  final String title;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 40,
      alignment: Alignment.center,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(
          bottom: BorderSide(color: Color(0xFFB98A71), width: 1),
        ),
      ),
      child: Text(
        title,
        style: TextStyle(
          color: _IepColors.ink,
          fontSize: 19,
          fontWeight: FontWeight.w900,
          letterSpacing: 1,
        ),
      ),
    );
  }
}

class _DocCellData {
  const _DocCellData({
    required this.text,
    required this.columns,
    this.bold = false,
    this.align = TextAlign.center,
    this.last = false,
    this.noWrap = false,
    this.editable = false,
  });

  final String text;
  final int columns;
  final bool bold;
  final TextAlign align;
  final bool last;
  final bool noWrap;
  final bool editable;
}

class _DocTableRow extends StatelessWidget {
  const _DocTableRow({required this.height, required this.cells});

  final double height;
  final List<_DocCellData> cells;

  @override
  Widget build(BuildContext context) {
    final List<Widget> children = <Widget>[];
    int columnIndex = 0;
    for (final _DocCellData cell in cells) {
      final int flex = _WordTable._columns
          .skip(columnIndex)
          .take(cell.columns)
          .fold<int>(0, (int sum, int width) => sum + width);
      columnIndex += cell.columns;
      children.add(
        Expanded(
          flex: flex,
          child: _DocCellBox(data: cell),
        ),
      );
    }

    return SizedBox(
      height: height,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: children,
      ),
    );
  }
}

class _DocCellBox extends StatelessWidget {
  const _DocCellBox({
    required this.data,
    this.rowLast = false,
    this.verticalPadding = 5,
    this.onTap,
  });

  final _DocCellData data;
  final bool rowLast;
  final double verticalPadding;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final Widget content = Container(
      alignment: Alignment.center,
      padding: EdgeInsets.fromLTRB(
        8,
        verticalPadding,
        data.editable ? 14 : 8,
        verticalPadding,
      ),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: data.editable
            ? const <BoxShadow>[
                BoxShadow(
                  color: Color(0x22E96F43),
                  blurRadius: 0,
                  spreadRadius: 1.4,
                ),
              ]
            : null,
        border: Border(
          right: data.last
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
          bottom: rowLast
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
        ),
      ),
      child: Stack(
        clipBehavior: Clip.none,
        children: <Widget>[
          Align(
            alignment: Alignment.center,
            child: Text(
              data.text,
              maxLines: data.noWrap ? 1 : 4,
              overflow: TextOverflow.ellipsis,
              textAlign: data.align,
              style: TextStyle(
                color: data.bold ? _IepColors.ink : _IepColors.text,
                fontSize: 11.4,
                fontWeight: data.bold ? FontWeight.w900 : FontWeight.w700,
                height: 1.22,
              ),
            ),
          ),
          if (data.editable)
            const Positioned(
              right: -6,
              top: -2,
              child: Icon(
                Icons.edit_rounded,
                size: 10,
                color: _IepColors.orangeDeep,
              ),
            ),
        ],
      ),
    );
    if (onTap == null) {
      return content;
    }
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        child: content,
      ),
    );
  }
}

class _DocPlanRows extends StatelessWidget {
  const _DocPlanRows({
    required this.domains,
    required this.selectedGoal,
    required this.onGoalTap,
  });

  final List<_DocDomainData> domains;
  final _GoalEditRequest? selectedGoal;
  final ValueChanged<_GoalEditRequest> onGoalTap;

  static const double _minDomainHeight = 122.4;
  static const double _shortGoalRowHeight = 39;

  static double blockHeightFor(_DocDomainData domain) {
    final int shortGoalCount =
        domain.shortGoals.isEmpty ? 1 : domain.shortGoals.length;
    final double contentHeight = shortGoalCount * _shortGoalRowHeight;
    return contentHeight > _minDomainHeight ? contentHeight : _minDomainHeight;
  }

  static double heightFor(List<_DocDomainData> domains) {
    return domains.fold<double>(
      0,
      (double height, _DocDomainData domain) => height + blockHeightFor(domain),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: domains.asMap().entries.map((entry) {
        return SizedBox(
          height: blockHeightFor(entry.value),
          child: _DocDomainBlock(
            domainIndex: entry.key,
            data: entry.value,
            selected: entry.key == 0,
            last: entry.key == domains.length - 1,
            selectedGoal: selectedGoal,
            onGoalTap: onGoalTap,
          ),
        );
      }).toList(),
    );
  }
}

class _DocDomainData {
  const _DocDomainData({
    required this.domain,
    required this.longGoals,
    required this.shortGoals,
  });

  final String domain;
  final List<String> longGoals;
  final List<_DocShortGoalData> shortGoals;

  _DocDomainData copyWith({
    List<String>? longGoals,
    List<_DocShortGoalData>? shortGoals,
  }) {
    return _DocDomainData(
      domain: domain,
      longGoals: longGoals ?? this.longGoals,
      shortGoals: shortGoals ?? this.shortGoals,
    );
  }
}

class _DocShortGoalData {
  const _DocShortGoalData(this.goal, this.lesson, this.period);

  final String goal;
  final String lesson;
  final String period;

  _DocShortGoalData copyWith({
    String? goal,
    String? lesson,
    String? period,
  }) {
    return _DocShortGoalData(
      goal ?? this.goal,
      lesson ?? this.lesson,
      period ?? this.period,
    );
  }
}

class _DocDomainBlock extends StatelessWidget {
  const _DocDomainBlock({
    required this.domainIndex,
    required this.data,
    required this.selected,
    required this.last,
    required this.selectedGoal,
    required this.onGoalTap,
  });

  final int domainIndex;
  final _DocDomainData data;
  final bool selected;
  final bool last;
  final _GoalEditRequest? selectedGoal;
  final ValueChanged<_GoalEditRequest> onGoalTap;

  @override
  Widget build(BuildContext context) {
    final _GoalEditRequest longGoalRequest =
        _GoalEditRequest.longGoal(domainIndex: domainIndex);
    return Row(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        Expanded(
          flex: _WordTable._columns[0],
          child: _DocMergedCell(text: data.domain, bold: true, rowLast: last),
        ),
        Expanded(
          flex: _WordTable._columns[1] +
              _WordTable._columns[2] +
              _WordTable._columns[3],
          child: _DocMergedCell(
            text: data.longGoals.join('\n'),
            align: TextAlign.left,
            rowLast: last,
            editable: selectedGoal == longGoalRequest,
            onTap: () => onGoalTap(longGoalRequest),
          ),
        ),
        Expanded(
          flex: _WordTable._columns[4] + _WordTable._columns[5],
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: data.shortGoals.asMap().entries.map((entry) {
              final _GoalEditRequest request = _GoalEditRequest.shortGoal(
                domainIndex: domainIndex,
                shortGoalIndex: entry.key,
              );
              return Expanded(
                child: _DocCellBox(
                  data: _DocCellData(
                    text: entry.value.goal,
                    columns: 2,
                    align: TextAlign.left,
                    editable: selectedGoal == request,
                  ),
                  rowLast: last && entry.key == data.shortGoals.length - 1,
                  verticalPadding: 4,
                  onTap: () => onGoalTap(request),
                ),
              );
            }).toList(),
          ),
        ),
        Expanded(
          flex: _WordTable._columns[6],
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: data.shortGoals.asMap().entries.map((entry) {
              return Expanded(
                child: _DocCellBox(
                  data: _DocCellData(
                    text: entry.value.lesson,
                    columns: 1,
                    noWrap: true,
                  ),
                  rowLast: last && entry.key == data.shortGoals.length - 1,
                  verticalPadding: 4,
                ),
              );
            }).toList(),
          ),
        ),
        Expanded(
          flex: _WordTable._columns[7],
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: data.shortGoals.asMap().entries.map((entry) {
              return Expanded(
                child: _DocCellBox(
                  data: _DocCellData(
                    text: entry.value.period,
                    columns: 1,
                    noWrap: true,
                    last: true,
                  ),
                  rowLast: last && entry.key == data.shortGoals.length - 1,
                  verticalPadding: 4,
                ),
              );
            }).toList(),
          ),
        ),
      ],
    );
  }
}

class _DocMergedCell extends StatelessWidget {
  const _DocMergedCell({
    required this.text,
    this.bold = false,
    this.align = TextAlign.center,
    this.rowLast = false,
    this.editable = false,
    this.onTap,
  });

  final String text;
  final bool bold;
  final TextAlign align;
  final bool rowLast;
  final bool editable;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return _DocCellBox(
      data: _DocCellData(
        text: text,
        columns: 1,
        bold: bold,
        align: align,
        editable: editable,
      ),
      rowLast: rowLast,
      verticalPadding: 6,
      onTap: onTap,
    );
  }
}
