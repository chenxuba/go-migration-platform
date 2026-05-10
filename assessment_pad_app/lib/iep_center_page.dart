import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'iep_assessment_record_client.dart';
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
  bool _queueBootstrapLoading = true;
  bool _showConfirmIep = false;

  void _selectRecord(IepAssessmentRecordSummary record) {
    final IepAssessmentRecordSummary? current = _selectedRecord;
    if (current != null && _sameRecord(current, record)) {
      return;
    }
    setState(() {
      _selectedRecord = record;
      _showConfirmIep = false;
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
                ),
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
                  onInitialLoadSettled: _handleQueueInitialLoadSettled,
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
                  queueBootstrapLoading: _queueBootstrapLoading,
                  onConfirmAvailabilityChanged:
                      _handleConfirmAvailabilityChanged,
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
  const _IepTopBar({
    required this.onBack,
    required this.metrics,
    required this.showConfirmIep,
  });

  final VoidCallback onBack;
  final _IepMetrics metrics;
  final bool showConfirmIep;

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
          if (showConfirmIep) ...<Widget>[
            const SizedBox(width: 10),
            const _PrimaryActionButton(
                icon: Icons.check_circle_rounded, label: '确认IEP'),
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
