import 'dart:async';
import 'dart:math' as math;
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:printing/printing.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'iep_assessment_record_client.dart';
import 'downloaded_file_saver.dart';
import 'iep_plan_client.dart';
import 'pad_date_range_picker.dart';
import 'pad_responsive.dart';
import 'pad_top_message.dart';
import 'route_bootstrap.dart';

part 'iep_center_loading.dart';
part 'iep_center_plan_adapters.dart';
part 'iep_center_queue.dart';
part 'iep_center_workspace.dart';
part 'iep_center_dialogs.dart';
part 'iep_center_tables.dart';
part 'iep_center_print.dart';
part 'iep_lesson_session_page.dart';

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
  final GlobalKey<_IepWorkspaceState> _workspaceKey =
      GlobalKey<_IepWorkspaceState>();
  final GlobalKey<_StudentQueuePanelState> _queueKey =
      GlobalKey<_StudentQueuePanelState>();
  IepAssessmentRecordSummary? _selectedRecord;
  bool _queueBootstrapLoading = true;
  bool _showConfirmIep = false;
  bool _confirmingIep = false;
  bool _exportingWord = false;
  bool _printingPlan = false;
  late DateTimeRange _range;
  int _searchResetSeed = 0;
  final Map<String, String> _recordStatusOverrides = <String, String>{};
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();

  @override
  void initState() {
    super.initState();
    final DateTime today = _dateOnly(DateTime.now());
    _range = DateTimeRange(
      start: today.subtract(const Duration(days: 29)),
      end: today,
    );
  }

  void _selectRecord(IepAssessmentRecordSummary record) {
    final IepAssessmentRecordSummary? current = _selectedRecord;
    if (current != null && _sameRecord(current, record)) {
      return;
    }
    setState(() {
      _selectedRecord = record;
      _showConfirmIep = false;
      _confirmingIep = false;
    });
  }

  void _handleQueueInitialLoadSettled() {
    if (!_queueBootstrapLoading) {
      return;
    }
    setState(() {
      _queueBootstrapLoading = false;
    });
  }

  void _handleConfirmAvailabilityChanged(bool visible) {
    if (_showConfirmIep == visible) {
      return;
    }
    setState(() {
      _showConfirmIep = visible;
    });
  }

  void _handleRecordStatusChanged(
    IepAssessmentRecordSummary record,
    String status,
  ) {
    final String recordKey = _recordIdentityKey(record);
    final String normalizedStatus = status.trim();
    final String? current = _recordStatusOverrides[recordKey];
    if (normalizedStatus.isEmpty) {
      if (current == null) {
        return;
      }
      setState(() {
        _recordStatusOverrides.remove(recordKey);
      });
      return;
    }
    if (current == normalizedStatus) {
      return;
    }
    setState(() {
      _recordStatusOverrides[recordKey] = normalizedStatus;
    });
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
      topMargin: 12,
      key: 'iep-center-message',
    );
  }

  Future<void> _handleConfirmIepPressed() async {
    final _IepWorkspaceState? workspaceState = _workspaceKey.currentState;
    if (workspaceState == null || _confirmingIep) {
      return;
    }
    setState(() {
      _confirmingIep = true;
    });
    try {
      await workspaceState.requestConfirmIepPlan();
    } finally {
      if (!mounted) {
        return;
      }
      setState(() {
        _confirmingIep = false;
      });
    }
  }

  Future<void> _handleExportWordPressed() async {
    final _IepWorkspaceState? workspaceState = _workspaceKey.currentState;
    if (workspaceState == null || _exportingWord) {
      return;
    }
    setState(() {
      _exportingWord = true;
    });
    try {
      await workspaceState.exportCurrentPlanWord();
    } finally {
      if (!mounted) {
        return;
      }
      setState(() {
        _exportingWord = false;
      });
    }
  }

  Future<void> _handlePrintPressed() async {
    final _IepWorkspaceState? workspaceState = _workspaceKey.currentState;
    if (workspaceState == null || _printingPlan) {
      return;
    }
    setState(() {
      _printingPlan = true;
    });
    try {
      await workspaceState.printCurrentPlan();
    } finally {
      if (!mounted) {
        return;
      }
      setState(() {
        _printingPlan = false;
      });
    }
  }

  Future<void> _handleRangePressed() async {
    final DateTime today = _dateOnly(DateTime.now());
    final DateTimeRange? picked = await showPadDateRangePicker(
      context: context,
      initialRange: _range,
      today: today,
      minDate: DateTime(today.year - 5),
      maxDate: DateTime(today.year + 1, 12, 31),
    );
    if (picked == null || !mounted) {
      return;
    }
    final DateTimeRange nextRange = DateTimeRange(
      start: _dateOnly(picked.start),
      end: _dateOnly(picked.end),
    );
    setState(() {
      _range = nextRange;
    });
    _queueKey.currentState?.applyQueryFilters(
      assessmentDateBegin: _formatDateDash(nextRange.start),
      assessmentDateEnd: _formatDateDash(nextRange.end),
    );
  }

  void _handleSearchSubmitted(String value) {
    _queueKey.currentState?.applyQueryFilters(searchKey: value.trim());
  }

  void _handleResetFilters() {
    final DateTime today = _dateOnly(DateTime.now());
    final DateTimeRange nextRange = DateTimeRange(
      start: today.subtract(const Duration(days: 29)),
      end: today,
    );
    setState(() {
      _range = nextRange;
      _searchResetSeed += 1;
    });
    _queueKey.currentState?.resetQueryFilters(
      assessmentDateBegin: _formatDateDash(nextRange.start),
      assessmentDateEnd: _formatDateDash(nextRange.end),
    );
  }

  @override
  void dispose() {
    _messageController.dispose();
    super.dispose();
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
                child: _IepTopBar(
                  onBack: widget.onBack,
                  metrics: metrics,
                  showConfirmIep: _showConfirmIep,
                  confirmingIep: _confirmingIep,
                  exportingWord: _exportingWord,
                  printingPlan: _printingPlan,
                  range: _range,
                  searchResetSeed: _searchResetSeed,
                  onConfirmIep: _handleConfirmIepPressed,
                  onExportWord: _handleExportWordPressed,
                  onPrint: _handlePrintPressed,
                  onRangeTap: _handleRangePressed,
                  onSearchSubmitted: _handleSearchSubmitted,
                  onResetFilters: _handleResetFilters,
                ),
              ),
              Positioned(
                left: metrics.outer,
                top: 84,
                width: metrics.leftWidth,
                height: 660,
                child: _StudentQueuePanel(
                  key: _queueKey,
                  recordClient: widget.recordClient,
                  selectedRecord: _selectedRecord,
                  statusOverrides: _recordStatusOverrides,
                  onRecordSelected: _selectRecord,
                  onInitialLoadSettled: _handleQueueInitialLoadSettled,
                  initialAssessmentDateBegin: _formatDateDash(_range.start),
                  initialAssessmentDateEnd: _formatDateDash(_range.end),
                ),
              ),
              Positioned(
                left: metrics.contentLeft,
                top: 84,
                width: metrics.contentWidth,
                height: 660,
                child: _IepWorkspace(
                  key: _workspaceKey,
                  record: _selectedRecord,
                  planClient: widget.planClient,
                  queueBootstrapLoading: _queueBootstrapLoading,
                  onConfirmAvailabilityChanged:
                      _handleConfirmAvailabilityChanged,
                  onRecordStatusChanged: _handleRecordStatusChanged,
                  onMessage: _showMessage,
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

DateTime _addMonthsClamped(DateTime start, int months) {
  final DateTime targetMonth = DateTime(start.year, start.month + months, 1);
  final int lastDay = DateTime(targetMonth.year, targetMonth.month + 1, 0).day;
  final int targetDay = math.min(start.day, lastDay);
  return DateTime(targetMonth.year, targetMonth.month, targetDay);
}

DateTime _periodEndFor(DateTime start, int monthCount) {
  final DateTime periodStart = _dateOnly(start);
  return _addMonthsClamped(periodStart, monthCount);
}

DateTime _monthOnly(DateTime value) => DateTime(value.year, value.month);

String _recordIdentityKey(IepAssessmentRecordSummary record) {
  return '${record.source.trim().toUpperCase()}-${record.id}';
}

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
  return _periodMonthDates(start, _periodEndFor(start, monthCount))
      .map(_monthLabelFromDate)
      .toList(growable: false);
}

List<DateTime> _periodMonthDates(DateTime periodStart, DateTime periodEnd) {
  final List<DateTime> months = <DateTime>[];
  DateTime cursor = DateTime(periodStart.year, periodStart.month, 1);
  final DateTime normalizedEnd = DateTime(periodEnd.year, periodEnd.month, 1);
  while (!cursor.isAfter(normalizedEnd)) {
    months.add(cursor);
    cursor = DateTime(cursor.year, cursor.month + 1, 1);
  }
  return months;
}

DateTime _monthDateFromLabel(
  DateTime periodStart,
  int monthCount,
  String label,
) {
  for (final DateTime monthDate in _periodMonthDates(
      periodStart, _periodEndFor(periodStart, monthCount))) {
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

List<List<DateTime>> _workingWeekBucketsInMonthRange(
  DateTimeRange monthRange, {
  List<int> restWeekdays = const <int>[DateTime.sunday],
}) {
  DateTime cursor = _dateOnly(monthRange.start);
  final Set<int> blockedWeekdays = restWeekdays
      .where(
          (int value) => value >= DateTime.monday && value <= DateTime.sunday)
      .toSet();
  final List<List<DateTime>> buckets = <List<DateTime>>[];

  for (int currentWeek = 1; currentWeek <= 5; currentWeek += 1) {
    if (cursor.isAfter(monthRange.end)) {
      break;
    }
    final DateTime end = _earlierDate(_weekEndForStart(cursor), monthRange.end);
    final List<DateTime> dates = <DateTime>[];
    for (DateTime day = cursor;
        !day.isAfter(end);
        day = day.add(const Duration(days: 1))) {
      if (!blockedWeekdays.contains(day.weekday)) {
        dates.add(day);
      }
    }
    if (dates.isNotEmpty) {
      buckets.add(dates);
    }
    cursor = end.add(const Duration(days: 1));
  }
  return buckets;
}

List<DateTime> _weekDatesInMonthRange(
  DateTimeRange monthRange,
  int weekNumber, {
  List<int> restWeekdays = const <int>[DateTime.sunday],
}) {
  final List<List<DateTime>> buckets = _workingWeekBucketsInMonthRange(
    monthRange,
    restWeekdays: restWeekdays,
  );
  if (weekNumber < 1 || weekNumber > buckets.length) {
    return <DateTime>[];
  }
  return buckets[weekNumber - 1];
}

int _lastAvailableWeekInMonthRange(
  DateTimeRange monthRange, {
  List<int> restWeekdays = const <int>[DateTime.sunday],
}) {
  final int count = _workingWeekBucketsInMonthRange(
    monthRange,
    restWeekdays: restWeekdays,
  ).length;
  return count <= 0 ? 1 : count;
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
  const _IepTopBar({
    required this.onBack,
    required this.metrics,
    required this.showConfirmIep,
    required this.confirmingIep,
    required this.exportingWord,
    required this.printingPlan,
    required this.range,
    required this.searchResetSeed,
    required this.onConfirmIep,
    required this.onExportWord,
    required this.onPrint,
    required this.onRangeTap,
    required this.onSearchSubmitted,
    required this.onResetFilters,
  });

  final VoidCallback onBack;
  final _IepMetrics metrics;
  final bool showConfirmIep;
  final bool confirmingIep;
  final bool exportingWord;
  final bool printingPlan;
  final DateTimeRange range;
  final int searchResetSeed;
  final VoidCallback onConfirmIep;
  final VoidCallback onExportWord;
  final VoidCallback onPrint;
  final VoidCallback onRangeTap;
  final ValueChanged<String> onSearchSubmitted;
  final VoidCallback onResetFilters;

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
            _TopSelector(
              label:
                  '${_formatDateDash(range.start)} - ${_formatDateDash(range.end)}',
              width: 250,
              onTap: onRangeTap,
            ),
            const SizedBox(width: 12),
          ],
          _IepSearchBox(
            width: metrics.compact ? 204 : 248,
            resetSeed: searchResetSeed,
            onSubmitted: onSearchSubmitted,
          ),
          const SizedBox(width: 12),
          _SoftActionButton(
            icon: Icons.restart_alt_rounded,
            label: '重置',
            onTap: onResetFilters,
          ),
          const SizedBox(width: 10),
          _SoftActionButton(
            icon: Icons.file_download_outlined,
            label: exportingWord ? '导出中' : '导出Word',
            loading: exportingWord,
            onTap: exportingWord ? null : onExportWord,
          ),
          const SizedBox(width: 10),
          _SoftActionButton(
            icon: Icons.print_rounded,
            label: printingPlan ? '打印中' : '打印',
            loading: printingPlan,
            onTap: printingPlan ? null : onPrint,
          ),
          if (showConfirmIep) ...<Widget>[
            const SizedBox(width: 10),
            _PrimaryActionButton(
              icon: Icons.check_circle_rounded,
              label: '确认IEP',
              loading: confirmingIep,
              onTap: confirmingIep ? null : onConfirmIep,
            ),
          ],
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
  const _TopSelector({
    required this.label,
    required this.width,
    this.onTap,
  });

  final String label;
  final double width;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Container(
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
              const Icon(
                Icons.calendar_month_rounded,
                color: _IepColors.muted,
                size: 18,
              ),
              const SizedBox(width: 8),
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
        ),
      ),
    );
  }
}

class _IepSearchBox extends StatefulWidget {
  const _IepSearchBox({
    required this.width,
    required this.resetSeed,
    required this.onSubmitted,
  });

  final double width;
  final int resetSeed;
  final ValueChanged<String> onSubmitted;

  @override
  State<_IepSearchBox> createState() => _IepSearchBoxState();
}

class _IepSearchBoxState extends State<_IepSearchBox> {
  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController();
  }

  @override
  void didUpdateWidget(covariant _IepSearchBox oldWidget) {
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
      width: widget.width,
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.92),
        borderRadius: BorderRadius.circular(19),
        border: Border.all(color: _IepColors.line),
      ),
      child: TextField(
        controller: _controller,
        onSubmitted: widget.onSubmitted,
        textInputAction: TextInputAction.search,
        style: const TextStyle(
          color: _IepColors.ink,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
        decoration: const InputDecoration(
          isDense: true,
          border: InputBorder.none,
          prefixIcon: Icon(
            Icons.search_rounded,
            color: _IepColors.ink,
            size: 21,
          ),
          prefixIconConstraints: BoxConstraints(minWidth: 34),
          hintText: '搜索学员姓名',
          hintStyle: TextStyle(
            color: _IepColors.muted,
            fontSize: 13,
            fontWeight: FontWeight.w600,
          ),
          contentPadding: EdgeInsets.fromLTRB(0, 10, 0, 10),
        ),
      ),
    );
  }
}

class _SoftActionButton extends StatelessWidget {
  const _SoftActionButton({
    required this.icon,
    required this.label,
    this.loading = false,
    this.onTap,
  });

  final IconData icon;
  final String label;
  final bool loading;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(19),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(19),
        child: Container(
          height: 38,
          padding: const EdgeInsets.symmetric(horizontal: 13),
          decoration: BoxDecoration(
            color: onTap == null
                ? _IepColors.orangeSoft.withOpacity(.72)
                : _IepColors.orangeSoft,
            borderRadius: BorderRadius.circular(19),
            border: Border.all(color: const Color(0xFFFFCDB5)),
          ),
          child: Row(
            children: <Widget>[
              if (loading)
                const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    valueColor: AlwaysStoppedAnimation<Color>(
                      _IepColors.orangeDeep,
                    ),
                  ),
                )
              else
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
        ),
      ),
    );
  }
}

class _PrimaryActionButton extends StatelessWidget {
  const _PrimaryActionButton({
    required this.icon,
    required this.label,
    this.loading = false,
    this.onTap,
  });

  final IconData icon;
  final String label;
  final bool loading;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(19),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(19),
        child: Container(
          height: 38,
          padding: const EdgeInsets.symmetric(horizontal: 15),
          decoration: BoxDecoration(
            color: onTap == null
                ? _IepColors.orange.withOpacity(.72)
                : _IepColors.orange,
            borderRadius: BorderRadius.circular(19),
            boxShadow: _iepShadow(
              color: const Color(0x2CE96F43),
              blur: 14,
              offset: const Offset(0, 7),
            ),
          ),
          child: Row(
            children: <Widget>[
              if (loading)
                const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                  ),
                )
              else
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
        ),
      ),
    );
  }
}
