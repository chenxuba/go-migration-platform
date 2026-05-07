import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'pad_responsive.dart';
import 'pad_top_message.dart';
import 'timetable_client.dart';

part 'smart_timetable/smart_timetable_state.dart';
part 'smart_timetable/smart_timetable_state_selectors.dart';
part 'smart_timetable/smart_timetable_state_loading.dart';
part 'smart_timetable/smart_timetable_state_drag.dart';
part 'smart_timetable/smart_timetable_state_schedule.dart';
part 'smart_timetable/smart_timetable_state_message.dart';
part 'smart_timetable/smart_timetable_helpers.dart';
part 'smart_timetable/smart_timetable_shell.dart';
part 'smart_timetable/smart_timetable_toolbar.dart';
part 'smart_timetable/smart_timetable_board.dart';
part 'smart_timetable/smart_timetable_controls.dart';
part 'smart_timetable/smart_timetable_common.dart';
part 'smart_timetable/smart_timetable_models.dart';

class SmartTimetablePage extends StatefulWidget {
  const SmartTimetablePage({
    this.timetableClient = const ApiTimetableClient(),
    super.key,
  });

  final TimetableClient timetableClient;

  @override
  State<SmartTimetablePage> createState() => _SmartTimetablePageState();
}
