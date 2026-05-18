import 'dart:async';

import 'package:flutter/material.dart';

import 'pad_responsive.dart';

Future<DateTimeRange?> showPadDateRangePicker({
  required BuildContext context,
  required DateTimeRange initialRange,
  DateTime? today,
  DateTime? minDate,
  DateTime? maxDate,
}) {
  final DateTime currentDay = _dateOnly(today ?? DateTime.now());
  final Completer<DateTimeRange?> completer = Completer<DateTimeRange?>();
  final OverlayState overlay = Overlay.of(context, rootOverlay: true);
  late final OverlayEntry entry;

  void close([DateTimeRange? range]) {
    if (entry.mounted) {
      entry.remove();
    }
    if (!completer.isCompleted) {
      completer.complete(range);
    }
  }

  entry = OverlayEntry(
    builder: (BuildContext context) {
      return _PadDateRangeOverlay(
        initialRange: initialRange,
        today: currentDay,
        minDate: _dateOnly(minDate ?? DateTime(currentDay.year - 5)),
        maxDate: _dateOnly(maxDate ?? DateTime(currentDay.year + 1, 12, 31)),
        onCancel: close,
        onSubmit: close,
      );
    },
  );
  overlay.insert(entry);
  return completer.future;
}

Future<DateTime?> showPadDatePicker({
  required BuildContext context,
  required DateTime initialDate,
  String title = '选择评估日期',
  String helperText = '请选择实际评估日期',
  DateTime? today,
  DateTime? minDate,
  DateTime? maxDate,
  bool disableFutureDates = false,
  bool initiallySelectDate = true,
}) {
  final DateTime currentDay = _dateOnly(today ?? DateTime.now());
  final DateTime resolvedMinDate = _dateOnly(
    minDate ?? DateTime(currentDay.year - 5),
  );
  final DateTime requestedMaxDate = _dateOnly(
    maxDate ?? DateTime(currentDay.year + 1, 12, 31),
  );
  final DateTime resolvedMaxDate =
      disableFutureDates && requestedMaxDate.isAfter(currentDay)
          ? currentDay
          : requestedMaxDate;
  final DateTime resolvedInitialDate = _clampDayWithinBounds(
    _dateOnly(initialDate),
    resolvedMinDate,
    resolvedMaxDate,
  );
  final Completer<DateTime?> completer = Completer<DateTime?>();
  final OverlayState overlay = Overlay.of(context, rootOverlay: true);
  late final OverlayEntry entry;

  void close([DateTime? value]) {
    if (entry.mounted) {
      entry.remove();
    }
    if (!completer.isCompleted) {
      completer.complete(value);
    }
  }

  entry = OverlayEntry(
    builder: (BuildContext context) {
      return _PadSingleDateOverlay(
        initialDate: resolvedInitialDate,
        title: title,
        helperText: helperText,
        today: currentDay,
        minDate: resolvedMinDate,
        maxDate: resolvedMaxDate,
        disableFutureDates: disableFutureDates,
        initiallySelectDate: initiallySelectDate,
        onCancel: close,
        onSubmit: close,
      );
    },
  );
  overlay.insert(entry);
  return completer.future;
}

class _PickerColors {
  static const Color surface = Color(0xFFFFFDFA);
  static const Color ink = Color(0xFF3F2B22);
  static const Color text = Color(0xFF6F5B50);
  static const Color muted = Color(0xFFA7958B);
  static const Color line = Color(0xFFEAD7C9);
  static const Color lineSoft = Color(0xFFF4E8DF);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95D37);
}

class _PadSingleDateOverlay extends StatelessWidget {
  const _PadSingleDateOverlay({
    required this.initialDate,
    required this.title,
    required this.helperText,
    required this.today,
    required this.minDate,
    required this.maxDate,
    required this.disableFutureDates,
    required this.initiallySelectDate,
    required this.onCancel,
    required this.onSubmit,
  });

  final DateTime initialDate;
  final String title;
  final String helperText;
  final DateTime today;
  final DateTime minDate;
  final DateTime maxDate;
  final bool disableFutureDates;
  final bool initiallySelectDate;
  final VoidCallback onCancel;
  final ValueChanged<DateTime> onSubmit;

  @override
  Widget build(BuildContext context) {
    return Material(
      type: MaterialType.transparency,
      child: Stack(
        children: <Widget>[
          const Positioned.fill(
            child: ColoredBox(color: Color(0x33000000)),
          ),
          Positioned.fill(
            child: PadDialogViewport(
              child: _PadSingleDateDialog(
                initialDate: initialDate,
                title: title,
                helperText: helperText,
                today: today,
                minDate: minDate,
                maxDate: maxDate,
                disableFutureDates: disableFutureDates,
                initiallySelectDate: initiallySelectDate,
                onCancel: onCancel,
                onSubmit: onSubmit,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _PadSingleDateDialog extends StatefulWidget {
  const _PadSingleDateDialog({
    required this.initialDate,
    required this.title,
    required this.helperText,
    required this.today,
    required this.minDate,
    required this.maxDate,
    required this.disableFutureDates,
    required this.initiallySelectDate,
    required this.onCancel,
    required this.onSubmit,
  });

  final DateTime initialDate;
  final String title;
  final String helperText;
  final DateTime today;
  final DateTime minDate;
  final DateTime maxDate;
  final bool disableFutureDates;
  final bool initiallySelectDate;
  final VoidCallback onCancel;
  final ValueChanged<DateTime> onSubmit;

  @override
  State<_PadSingleDateDialog> createState() => _PadSingleDateDialogState();
}

class _PadSingleDateDialogState extends State<_PadSingleDateDialog> {
  late DateTime _visibleMonth;
  DateTime? _selectedDate;
  _SingleDatePanel _panel = _SingleDatePanel.day;

  @override
  void initState() {
    super.initState();
    _selectedDate =
        widget.initiallySelectDate ? _dateOnly(widget.initialDate) : null;
    _visibleMonth = _boundedVisibleMonth(_monthOnly(widget.initialDate));
  }

  DateTime _boundedVisibleMonth(DateTime month) {
    final DateTime minMonth = _monthOnly(widget.minDate);
    final DateTime maxMonth = _monthOnly(widget.maxDate);
    if (month.isBefore(minMonth)) {
      return minMonth;
    }
    if (month.isAfter(maxMonth)) {
      return maxMonth;
    }
    return month;
  }

  void _shiftMonth(int offset) {
    setState(() {
      _visibleMonth = _boundedVisibleMonth(_addMonths(_visibleMonth, offset));
      _panel = _SingleDatePanel.day;
    });
  }

  void _selectDay(DateTime day) {
    final DateTime value = _dateOnly(day);
    if (_isBeforeDay(value, widget.minDate) ||
        _isAfterDay(value, widget.maxDate) ||
        (widget.disableFutureDates && _isAfterDay(value, widget.today))) {
      return;
    }
    setState(() {
      _selectedDate = value;
      _panel = _SingleDatePanel.day;
    });
  }

  void _showYearPanel() {
    setState(() => _panel = _SingleDatePanel.year);
  }

  void _showMonthPanel() {
    setState(() => _panel = _SingleDatePanel.month);
  }

  void _selectYear(int year) {
    setState(() {
      _visibleMonth = _boundedVisibleMonth(
        DateTime(year, _visibleMonth.month),
      );
      _panel = _SingleDatePanel.month;
    });
  }

  void _selectMonth(int month) {
    setState(() {
      _visibleMonth = _boundedVisibleMonth(
        DateTime(_visibleMonth.year, month),
      );
      _panel = _SingleDatePanel.day;
    });
  }

  Widget _buildCalendarBody(DateTime? selectedDate) {
    switch (_panel) {
      case _SingleDatePanel.year:
        return _YearQuickGrid(
          minDate: widget.minDate,
          maxDate: widget.maxDate,
          today: widget.today,
          disableFutureDates: widget.disableFutureDates,
          visibleMonth: _visibleMonth,
          onSelect: _selectYear,
        );
      case _SingleDatePanel.month:
        return _MonthQuickGrid(
          minDate: widget.minDate,
          maxDate: widget.maxDate,
          today: widget.today,
          disableFutureDates: widget.disableFutureDates,
          visibleMonth: _visibleMonth,
          onSelect: _selectMonth,
        );
      case _SingleDatePanel.day:
        return _MonthCalendar(
          month: _visibleMonth,
          minDate: widget.minDate,
          maxDate: widget.maxDate,
          today: widget.today,
          disableFutureDates: widget.disableFutureDates,
          start: selectedDate,
          end: selectedDate,
          onSelect: _selectDay,
        );
    }
  }

  @override
  Widget build(BuildContext context) {
    final DateTime? selectedDate = _selectedDate;
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: 560,
        padding: const EdgeInsets.fromLTRB(24, 22, 24, 22),
        decoration: BoxDecoration(
          color: _PickerColors.surface,
          borderRadius: BorderRadius.circular(22),
          border: Border.all(color: _PickerColors.line),
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
                Text(
                  widget.title,
                  style: TextStyle(
                    color: _PickerColors.ink,
                    fontSize: 22,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const Spacer(),
                _RangePreview(
                  label: '日期',
                  value: selectedDate == null ? '未选择' : _dateText(selectedDate),
                ),
                const SizedBox(width: 10),
                _IconCircleButton(
                  icon: Icons.close_rounded,
                  onTap: widget.onCancel,
                ),
              ],
            ),
            const SizedBox(height: 22),
            Row(
              children: <Widget>[
                _IconCircleButton(
                  icon: Icons.chevron_left_rounded,
                  enabled: _panel == _SingleDatePanel.day && _canShiftPrevious,
                  onTap: () => _shiftMonth(-1),
                ),
                const Spacer(),
                _MonthJumpControl(
                  month: _visibleMonth,
                  panel: _panel,
                  onYearTap: _showYearPanel,
                  onMonthTap: _showMonthPanel,
                ),
                const Spacer(),
                _IconCircleButton(
                  icon: Icons.chevron_right_rounded,
                  enabled: _panel == _SingleDatePanel.day && _canShiftNext,
                  onTap: () => _shiftMonth(1),
                ),
              ],
            ),
            const SizedBox(height: 16),
            SizedBox(
              height: 306,
              child: _buildCalendarBody(selectedDate),
            ),
            const SizedBox(height: 18),
            Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: <Widget>[
                Text(
                  widget.helperText,
                  style: TextStyle(
                    color: _PickerColors.muted,
                    fontSize: 13,
                    fontWeight: FontWeight.w800,
                    height: 1.35,
                  ),
                ),
                const SizedBox(height: 14),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: <Widget>[
                    _DialogActionButton(
                      label: '取消',
                      onTap: widget.onCancel,
                    ),
                    const SizedBox(width: 10),
                    _DialogActionButton(
                      label: '确定',
                      filled: true,
                      enabled: selectedDate != null,
                      onTap: selectedDate == null
                          ? null
                          : () => widget.onSubmit(selectedDate),
                    ),
                  ],
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
    final DateTime nextMonth = _addMonths(_visibleMonth, 1);
    if (widget.disableFutureDates &&
        nextMonth.isAfter(_monthOnly(widget.today))) {
      return false;
    }
    return !nextMonth.isAfter(_monthOnly(widget.maxDate));
  }
}

enum _SingleDatePanel { day, year, month }

class _PadDateRangeOverlay extends StatelessWidget {
  const _PadDateRangeOverlay({
    required this.initialRange,
    required this.today,
    required this.minDate,
    required this.maxDate,
    required this.onCancel,
    required this.onSubmit,
  });

  final DateTimeRange initialRange;
  final DateTime today;
  final DateTime minDate;
  final DateTime maxDate;
  final VoidCallback onCancel;
  final ValueChanged<DateTimeRange> onSubmit;

  @override
  Widget build(BuildContext context) {
    return Material(
      type: MaterialType.transparency,
      child: Stack(
        children: <Widget>[
          Positioned.fill(
            child: GestureDetector(
              behavior: HitTestBehavior.opaque,
              onTap: onCancel,
              child: const ColoredBox(color: Color(0x33000000)),
            ),
          ),
          Positioned.fill(
            child: PadDialogViewport(
              child: _PadDateRangeDialog(
                initialRange: initialRange,
                today: today,
                minDate: minDate,
                maxDate: maxDate,
                onCancel: onCancel,
                onSubmit: onSubmit,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _PadDateRangeDialog extends StatefulWidget {
  const _PadDateRangeDialog({
    required this.initialRange,
    required this.today,
    required this.minDate,
    required this.maxDate,
    required this.onCancel,
    required this.onSubmit,
  });

  final DateTimeRange initialRange;
  final DateTime today;
  final DateTime minDate;
  final DateTime maxDate;
  final VoidCallback onCancel;
  final ValueChanged<DateTimeRange> onSubmit;

  @override
  State<_PadDateRangeDialog> createState() => _PadDateRangeDialogState();
}

class _PadDateRangeDialogState extends State<_PadDateRangeDialog> {
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
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: 920,
        padding: const EdgeInsets.fromLTRB(26, 24, 26, 24),
        decoration: BoxDecoration(
          color: _PickerColors.surface,
          borderRadius: BorderRadius.circular(22),
          border: Border.all(color: _PickerColors.line),
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
                    color: _PickerColors.ink,
                    fontSize: 22,
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
                  onTap: widget.onCancel,
                ),
              ],
            ),
            const SizedBox(height: 22),
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                SizedBox(
                  width: 136,
                  child: _RangePresetRail(
                    today: widget.today,
                    selectedLabel: _selectedPresetLabel,
                    onSelected: _applyPreset,
                  ),
                ),
                const SizedBox(width: 20),
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
                          const SizedBox(width: 132),
                          _MonthTitle(month: _addMonths(_visibleMonth, 1)),
                          const Spacer(),
                          _IconCircleButton(
                            icon: Icons.chevron_right_rounded,
                            enabled: _canShiftNext,
                            onTap: () => _shiftMonth(1),
                          ),
                        ],
                      ),
                      const SizedBox(height: 16),
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
                          const SizedBox(width: 22),
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
            const SizedBox(height: 22),
            Row(
              children: <Widget>[
                const Text(
                  '点击日期重新选择开始日期，再点击一次选择结束日期',
                  style: TextStyle(
                    color: _PickerColors.muted,
                    fontSize: 13,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const Spacer(),
                _DialogActionButton(
                  label: '取消',
                  onTap: widget.onCancel,
                ),
                const SizedBox(width: 10),
                _DialogActionButton(
                  label: '确定',
                  filled: true,
                  enabled: canSubmit,
                  onTap: canSubmit
                      ? () {
                          widget.onSubmit(
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
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 13),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF8F2),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _PickerColors.lineSoft),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Text(
            label,
            style: const TextStyle(
              color: _PickerColors.muted,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 7),
          Text(
            value,
            style: TextStyle(
              color: muted ? _PickerColors.muted : _PickerColors.ink,
              fontSize: 15,
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
            color: _PickerColors.muted,
            fontSize: 14,
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
      '上月',
      lastMonthStart,
      monthStart.subtract(const Duration(days: 1)),
    ),
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
          height: 42,
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFF0E7) : const Color(0xFFFFF8F2),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: selected ? _PickerColors.orange : _PickerColors.lineSoft,
            ),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              if (selected) ...<Widget>[
                const Icon(
                  Icons.check_rounded,
                  size: 17,
                  color: _PickerColors.orangeDeep,
                ),
                const SizedBox(width: 4),
              ],
              Flexible(
                child: Text(
                  label,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: selected
                        ? _PickerColors.orangeDeep
                        : _PickerColors.text,
                    fontSize: 15,
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

class _MonthJumpControl extends StatelessWidget {
  const _MonthJumpControl({
    required this.month,
    required this.panel,
    required this.onYearTap,
    required this.onMonthTap,
  });

  final DateTime month;
  final _SingleDatePanel panel;
  final VoidCallback onYearTap;
  final VoidCallback onMonthTap;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        _MonthJumpButton(
          label: '${month.year}年',
          active: panel == _SingleDatePanel.year,
          onTap: onYearTap,
        ),
        const SizedBox(width: 8),
        _MonthJumpButton(
          label: '${month.month}月',
          active: panel == _SingleDatePanel.month,
          onTap: onMonthTap,
        ),
      ],
    );
  }
}

class _MonthJumpButton extends StatelessWidget {
  const _MonthJumpButton({
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
        child: Container(
          height: 40,
          constraints: const BoxConstraints(minWidth: 84),
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: active ? const Color(0xFFFFF0E7) : const Color(0xFFFFF8F2),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: active ? _PickerColors.orange : _PickerColors.lineSoft,
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              Text(
                label,
                style: TextStyle(
                  color: active ? _PickerColors.orangeDeep : _PickerColors.ink,
                  fontSize: 17,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(width: 4),
              Icon(
                Icons.keyboard_arrow_down_rounded,
                size: 18,
                color: active ? _PickerColors.orange : _PickerColors.muted,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _YearQuickGrid extends StatelessWidget {
  const _YearQuickGrid({
    required this.minDate,
    required this.maxDate,
    required this.today,
    required this.disableFutureDates,
    required this.visibleMonth,
    required this.onSelect,
  });

  final DateTime minDate;
  final DateTime maxDate;
  final DateTime today;
  final bool disableFutureDates;
  final DateTime visibleMonth;
  final ValueChanged<int> onSelect;

  @override
  Widget build(BuildContext context) {
    final int minYear = minDate.year;
    final int maxYear = disableFutureDates && today.year < maxDate.year
        ? today.year
        : maxDate.year;
    final int count = (maxYear - minYear + 1).clamp(0, 200);
    return GridView.builder(
      padding: const EdgeInsets.fromLTRB(2, 2, 2, 4),
      physics: const BouncingScrollPhysics(),
      itemCount: count,
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 5,
        mainAxisSpacing: 10,
        crossAxisSpacing: 10,
        childAspectRatio: 1.78,
      ),
      itemBuilder: (BuildContext context, int index) {
        final int year = maxYear - index;
        return _QuickSelectCell(
          label: '$year',
          selected: year == visibleMonth.year,
          enabled: true,
          onTap: () => onSelect(year),
        );
      },
    );
  }
}

class _MonthQuickGrid extends StatelessWidget {
  const _MonthQuickGrid({
    required this.minDate,
    required this.maxDate,
    required this.today,
    required this.disableFutureDates,
    required this.visibleMonth,
    required this.onSelect,
  });

  final DateTime minDate;
  final DateTime maxDate;
  final DateTime today;
  final bool disableFutureDates;
  final DateTime visibleMonth;
  final ValueChanged<int> onSelect;

  @override
  Widget build(BuildContext context) {
    return GridView.builder(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
      physics: const NeverScrollableScrollPhysics(),
      itemCount: 12,
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 4,
        mainAxisSpacing: 12,
        crossAxisSpacing: 12,
        childAspectRatio: 2.1,
      ),
      itemBuilder: (BuildContext context, int index) {
        final int month = index + 1;
        final DateTime candidate = DateTime(visibleMonth.year, month);
        final bool enabled = !_monthBefore(candidate, minDate) &&
            !_monthAfter(candidate, maxDate) &&
            !(disableFutureDates && _monthAfter(candidate, today));
        return _QuickSelectCell(
          label: '$month月',
          selected: enabled && month == visibleMonth.month,
          enabled: enabled,
          onTap: () => onSelect(month),
        );
      },
    );
  }
}

class _QuickSelectCell extends StatelessWidget {
  const _QuickSelectCell({
    required this.label,
    required this.selected,
    required this.enabled,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final bool enabled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color foreground = selected
        ? _PickerColors.orangeDeep
        : enabled
            ? _PickerColors.text
            : _PickerColors.muted.withOpacity(.48);
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: enabled ? onTap : null,
        borderRadius: BorderRadius.circular(12),
        child: Ink(
          decoration: BoxDecoration(
            color: selected
                ? const Color(0xFFFFF0E7)
                : enabled
                    ? const Color(0xFFFFF8F2)
                    : const Color(0xFFF8EEE6).withOpacity(.56),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: selected ? _PickerColors.orange : _PickerColors.lineSoft,
            ),
          ),
          child: Center(
            child: Text(
              label,
              maxLines: 1,
              style: TextStyle(
                color: foreground,
                fontSize: 15,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
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
      width: 116,
      child: Text(
        '${month.year}年${month.month}月',
        textAlign: TextAlign.center,
        style: const TextStyle(
          color: _PickerColors.ink,
          fontSize: 18,
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
    this.disableFutureDates = false,
    required this.start,
    required this.end,
    required this.onSelect,
  });

  final DateTime month;
  final DateTime minDate;
  final DateTime maxDate;
  final DateTime today;
  final bool disableFutureDates;
  final DateTime? start;
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
        const SizedBox(height: 8),
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
                    disableFutureDates: disableFutureDates,
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
            color: _PickerColors.muted,
            fontSize: 13,
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
    required this.disableFutureDates,
    required this.start,
    required this.end,
    required this.onSelect,
  });

  final DateTime day;
  final DateTime visibleMonth;
  final DateTime minDate;
  final DateTime maxDate;
  final DateTime today;
  final bool disableFutureDates;
  final DateTime? start;
  final DateTime? end;
  final ValueChanged<DateTime> onSelect;

  @override
  Widget build(BuildContext context) {
    final DateTime current = _dateOnly(day);
    final bool inCurrentMonth = current.year == visibleMonth.year &&
        current.month == visibleMonth.month;
    final bool futureDisabled =
        disableFutureDates && _isAfterDay(current, today);
    final bool enabled = inCurrentMonth &&
        !_isBeforeDay(current, minDate) &&
        !_isAfterDay(current, maxDate) &&
        !futureDisabled;
    final bool isStart =
        inCurrentMonth && start != null && _sameDay(current, start!);
    final bool isEnd = inCurrentMonth && end != null && _sameDay(current, end!);
    final bool inRange = start != null &&
        end != null &&
        !current.isBefore(start!) &&
        !current.isAfter(end!) &&
        inCurrentMonth;
    final bool isToday = inCurrentMonth && _sameDay(current, today);
    final int lastDayOfMonth =
        DateTime(visibleMonth.year, visibleMonth.month + 1, 0).day;
    final bool startsRangeBand =
        isStart || current.weekday == DateTime.monday || current.day == 1;
    final bool endsRangeBand = isEnd ||
        current.weekday == DateTime.sunday ||
        current.day == lastDayOfMonth;
    final bool isSingleDay = isStart && isEnd;
    final bool shouldShowBand = inRange && !isSingleDay;
    final Color textColor = futureDisabled
        ? const Color(0xFFC9B7AC).withOpacity(.46)
        : isStart || isEnd
            ? Colors.white
            : enabled
                ? _PickerColors.text
                : _PickerColors.line;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: enabled ? () => onSelect(current) : null,
          borderRadius: BorderRadius.circular(10),
          child: SizedBox(
            height: 40,
            child: LayoutBuilder(
              builder: (BuildContext context, BoxConstraints constraints) {
                final double halfWidth = constraints.maxWidth / 2;
                return Stack(
                  alignment: Alignment.center,
                  children: <Widget>[
                    if (shouldShowBand)
                      Positioned(
                        top: 6,
                        bottom: 6,
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
                    if (futureDisabled && inCurrentMonth)
                      Container(
                        width: 34,
                        height: 30,
                        decoration: BoxDecoration(
                          color: const Color(0xFFF8EEE6).withOpacity(.64),
                          borderRadius: BorderRadius.circular(10),
                        ),
                      )
                    else if (isStart || isEnd)
                      Container(
                        width: 34,
                        height: 30,
                        decoration: BoxDecoration(
                          color: _PickerColors.orange,
                          borderRadius: BorderRadius.circular(10),
                        ),
                      )
                    else if (isToday)
                      Container(
                        width: 34,
                        height: 30,
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(10),
                          border: Border.all(
                            color: _PickerColors.orange.withOpacity(.42),
                          ),
                        ),
                      ),
                    Text(
                      '${current.day}',
                      style: TextStyle(
                        color: textColor,
                        fontSize: 14,
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
          width: 40,
          height: 40,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: enabled ? const Color(0xFFFFF8F2) : const Color(0xFFF8EEE6),
            shape: BoxShape.circle,
            border: Border.all(color: _PickerColors.lineSoft),
          ),
          child: Icon(
            icon,
            size: 23,
            color: enabled ? _PickerColors.text : _PickerColors.muted,
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
          height: 44,
          constraints: const BoxConstraints(minWidth: 88),
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(horizontal: 20),
          decoration: BoxDecoration(
            color: filled && enabled
                ? _PickerColors.orange
                : _PickerColors.surface,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color:
                  filled && enabled ? _PickerColors.orange : _PickerColors.line,
            ),
          ),
          child: Text(
            label,
            style: TextStyle(
              color: filled && enabled
                  ? Colors.white
                  : enabled
                      ? _PickerColors.text
                      : _PickerColors.muted,
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

DateTime _dateOnly(DateTime value) =>
    DateTime(value.year, value.month, value.day);

DateTime _monthOnly(DateTime value) => DateTime(value.year, value.month);

DateTime _addMonths(DateTime value, int months) {
  return DateTime(value.year, value.month + months);
}

DateTime _clampDayWithinBounds(DateTime value, DateTime min, DateTime max) {
  final DateTime day = _dateOnly(value);
  if (_isBeforeDay(day, min)) {
    return _dateOnly(min);
  }
  if (_isAfterDay(day, max)) {
    return _dateOnly(max);
  }
  return day;
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

bool _monthBefore(DateTime left, DateTime right) {
  return _monthOnly(left).isBefore(_monthOnly(right));
}

bool _monthAfter(DateTime left, DateTime right) {
  return _monthOnly(left).isAfter(_monthOnly(right));
}

String _dateText(DateTime value) {
  final String month = value.month.toString().padLeft(2, '0');
  final String day = value.day.toString().padLeft(2, '0');
  return '${value.year}-$month-$day';
}
