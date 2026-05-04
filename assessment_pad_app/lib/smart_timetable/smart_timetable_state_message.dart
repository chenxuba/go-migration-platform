part of '../smart_timetable_page.dart';

extension _SmartTimetableStateMessage on _SmartTimetablePageState {
  void _showScheduleMessage(
    String message, {
    _ScheduleMessageTone tone = _ScheduleMessageTone.info,
  }) {
    if (!mounted || message.trim().isEmpty) {
      return;
    }
    _scheduleMessageTimer?.cancel();
    _scheduleMessageText = message.trim();
    _scheduleMessageTone = tone;
    final OverlayState overlay = Overlay.of(context, rootOverlay: true);
    if (_scheduleMessageEntry == null) {
      _scheduleMessageVisible = false;
      _scheduleMessageEntry = OverlayEntry(
        builder: (BuildContext context) {
          return _ScheduleTopMessage(
            visible: _scheduleMessageVisible,
            message: _scheduleMessageText,
            tone: _scheduleMessageTone,
          );
        },
      );
      overlay.insert(_scheduleMessageEntry!);
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted || _scheduleMessageEntry == null) {
          return;
        }
        _scheduleMessageVisible = true;
        _scheduleMessageEntry?.markNeedsBuild();
      });
    } else {
      _scheduleMessageVisible = true;
      _scheduleMessageEntry!.markNeedsBuild();
    }
    _scheduleMessageTimer = Timer(const Duration(milliseconds: 1900), () {
      _scheduleMessageVisible = false;
      _scheduleMessageEntry?.markNeedsBuild();
      _scheduleMessageTimer = Timer(
        const Duration(milliseconds: 180),
        _removeScheduleMessage,
      );
    });
  }

  void _removeScheduleMessage() {
    _scheduleMessageEntry?.remove();
    _scheduleMessageEntry = null;
  }
}
