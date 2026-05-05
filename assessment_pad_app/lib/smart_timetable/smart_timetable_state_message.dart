part of '../smart_timetable_page.dart';

extension _SmartTimetableStateMessage on _SmartTimetablePageState {
  void _showScheduleMessage(
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
      key: 'schedule-top-message',
    );
  }
}
