part of '../smart_timetable_page.dart';

class _SmartTimetablePageState extends State<SmartTimetablePage> {
  int _periodGroupIndex = 0;
  int _teacherIndex = 0;
  int _weekOffset = 0;
  int _loadSequence = 0;
  int _scheduleOptionsSequence = 0;
  int _availabilitySequence = 0;
  bool _teacherDropdownOpen = false;
  bool _schedulePanelOpen = false;
  bool _scheduleOptionsLoading = false;
  bool _scheduleOptionsLoaded = false;
  bool _availabilityLoading = false;
  bool _loading = true;
  String? _errorMessage;
  String? _scheduleOptionsError;
  String? _availabilityMessage;
  String _selectedPeriodGroupId = '';
  String _selectedTeacherId = '';
  _ScheduleMode _scheduleMode = _ScheduleMode.oneToOne;
  ScheduleTargetOption? _selectedOneToOneTarget;
  ScheduleTargetOption? _selectedGroupClassTarget;
  ScheduleClassroomOption? _selectedClassroom;
  List<ScheduleTargetOption> _oneToOneTargets = const <ScheduleTargetOption>[];
  List<ScheduleTargetOption> _groupClassTargets =
      const <ScheduleTargetOption>[];
  List<ScheduleStaffOption> _assistantOptions = const <ScheduleStaffOption>[];
  List<ScheduleClassroomOption> _classroomOptions =
      const <ScheduleClassroomOption>[];
  Set<String> _selectedAssistantIds = <String>{};
  Map<String, _SlotAvailability> _slotAvailability =
      const <String, _SlotAvailability>{};
  Map<String, _SlotAvailability> _dragSlotAvailability =
      const <String, _SlotAvailability>{};
  Set<String> _dragCheckingSlotKeys = const <String>{};
  int _dragValidationSequence = 0;
  bool _dragValidationActive = false;
  String? _creatingSlotKey;
  String? _movingScheduleId;
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();
  TimetableData _data = TimetableData.fallback();
  List<_PeriodGroupOption> _periodGroups = const <_PeriodGroupOption>[];
  List<_TeacherOption> _teachers = const <_TeacherOption>[];
  List<_WeekDay> _weekDays = const <_WeekDay>[];
  List<_TimeSlot> _timeSlots = const <_TimeSlot>[];
  List<List<_LessonCell?>> _scheduleRows = const <List<_LessonCell?>>[];

  @override
  void initState() {
    super.initState();
    this._applyTimetableData(_data, preserveTeacherSelection: false);
    this._loadTimetable();
  }

  @override
  void dispose() {
    _messageController.dispose();
    super.dispose();
  }

  void _updateState(VoidCallback callback) {
    setState(callback);
  }

  @override
  Widget build(BuildContext context) {
    final _TeacherOption teacher = _teachers.isEmpty
        ? const _TeacherOption(id: '', name: '当前老师', label: '当前老师')
        : _teachers[_teacherIndex.clamp(0, _teachers.length - 1)];
    final List<ScheduleStaffOption> currentGroupAssistantOptions =
        this._currentGroupAssistantOptions;
    final Set<String> currentGroupSelectedAssistantIds = _selectedAssistantIds
        .where(
          (String id) => currentGroupAssistantOptions.any(
            (ScheduleStaffOption item) => item.id == id,
          ),
        )
        .toSet();
    return Scaffold(
      body: _SmartTimetableViewport(
        child: _SmartTimetableScreen(
          teacher: teacher,
          periodGroups: _periodGroups,
          periodGroupIndex: _periodGroupIndex,
          teachers: _teachers,
          teacherIndex: _teacherIndex,
          teacherDropdownOpen: _teacherDropdownOpen,
          schedulePanelOpen: _schedulePanelOpen,
          scheduleMode: _scheduleMode,
          selectedScheduleTarget: this._selectedScheduleTarget,
          oneToOneTargets: _oneToOneTargets,
          groupClassTargets: _groupClassTargets,
          assistantOptions: currentGroupAssistantOptions,
          classroomOptions: _classroomOptions,
          selectedAssistantIds: currentGroupSelectedAssistantIds,
          selectedClassroom: _selectedClassroom,
          scheduleOptionsLoading: _scheduleOptionsLoading,
          scheduleOptionsError: _scheduleOptionsError,
          availabilityLoading: _availabilityLoading,
          availabilityMessage: _availabilityMessage,
          slotAvailability: _slotAvailability,
          dragSlotAvailability: _dragSlotAvailability,
          dragCheckingSlotKeys: _dragCheckingSlotKeys,
          dragValidationActive: _dragValidationActive,
          creatingSlotKey: _creatingSlotKey,
          scheduleRows: _scheduleRows,
          weekDays: _weekDays,
          timeSlots: _timeSlots,
          summary: _data.summary,
          errorMessage: _errorMessage,
          dateRange: _dateRangeText(_weekOffset),
          onBack: () => Navigator.of(context).maybePop(),
          onPrevWeek: () => _changeWeek(-1),
          onNextWeek: () => _changeWeek(1),
          onToday: this._backToCurrentWeek,
          onPeriodGroupSelected: this._selectPeriodGroup,
          onTeacherToggle: () => setState(() {
            _teacherDropdownOpen = !_teacherDropdownOpen;
            _schedulePanelOpen = false;
          }),
          onTeacherSelected: this._selectTeacher,
          onTeacherDropdownClose: () =>
              setState(() => _teacherDropdownOpen = false),
          onSchedulePanelToggle: this._toggleSchedulePanel,
          onSchedulePanelClose: this._closeSchedulePanel,
          onScheduleModeChanged: this._setScheduleMode,
          onScheduleTargetSelected: this._selectScheduleTarget,
          onScheduleTargetCleared: this._clearScheduleTarget,
          onAssistantToggled: this._toggleAssistant,
          onClassroomSelected: this._selectClassroom,
          onScheduleOptionsRefresh: _loadScheduleOptions,
          onAvailabilityRefresh: this._detectScheduleAvailability,
          onRefresh: _loadTimetable,
          onPrimaryScheduleTap: this._handlePrimaryScheduleTap,
          onLessonMove: this._moveLesson,
          onLessonDragStarted: this._startLessonDrag,
          onLessonDragEnded: this._endLessonDrag,
          onEmptySlotTap: this._handleEmptySlotTap,
        ),
      ),
    );
  }
}
