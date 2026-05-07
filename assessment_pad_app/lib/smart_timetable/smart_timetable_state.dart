part of '../smart_timetable_page.dart';

class _SmartTimetablePageState extends State<SmartTimetablePage> {
  int _periodGroupIndex = 0;
  int _teacherIndex = 0;
  int _weekOffset = 0;
  int _loadSequence = 0;
  int _scheduleOptionsSequence = 0;
  int _filterOptionsSequence = 0;
  int _availabilitySequence = 0;
  final LayerLink _periodGroupDropdownLink = LayerLink();
  bool _teacherDropdownOpen = false;
  bool _periodGroupDropdownOpen = false;
  bool _schedulePanelOpen = false;
  _TimetableFilterKind? _openFilterKind;
  bool _scheduleOptionsLoading = false;
  bool _scheduleOptionsLoaded = false;
  bool _filterOptionsLoading = false;
  bool _filterOptionsLoaded = false;
  bool _availabilityLoading = false;
  bool _loading = true;
  bool _bootstrapLoading = true;
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
  List<_TimetableFilterOption> _studentCatalogOptions =
      const <_TimetableFilterOption>[];
  List<_TimetableFilterOption> _courseCatalogOptions =
      const <_TimetableFilterOption>[];
  List<ScheduleStaffOption> _assistantOptions = const <ScheduleStaffOption>[];
  List<ScheduleClassroomOption> _classroomOptions =
      const <ScheduleClassroomOption>[];
  Set<String> _selectedAssistantIds = <String>{};
  Set<String> _selectedStudentFilters = <String>{};
  Set<String> _selectedCourseFilters = <String>{};
  Set<String> _selectedCallStatusFilters = <String>{};
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
    runAfterRouteEntrance(context, () => this._loadTimetable(bootstrap: true));
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
    final bool availabilityBusy =
        _availabilityLoading || _dragCheckingSlotKeys.isNotEmpty;
    final String? availabilityStatusMessage = _availabilityLoading
        ? _availabilityMessage
        : (_dragCheckingSlotKeys.isNotEmpty ? '检测中' : _availabilityMessage);
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
          periodGroupDropdownOpen: _periodGroupDropdownOpen,
          periodGroupDropdownLink: _periodGroupDropdownLink,
          schedulePanelOpen: _schedulePanelOpen,
          bootstrapLoading: _bootstrapLoading && _loading,
          loading: _loading,
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
          openFilterKind: _openFilterKind,
          studentFilterLabel: _filterButtonLabel(_TimetableFilterKind.student),
          courseFilterLabel: _filterButtonLabel(_TimetableFilterKind.course),
          callStatusFilterLabel:
              _filterButtonLabel(_TimetableFilterKind.callStatus),
          studentFilterOptions: _studentFilterOptions,
          courseFilterOptions: _courseFilterOptions,
          callStatusFilterOptions: _callStatusFilterOptions,
          selectedStudentFilters: _selectedStudentFilters,
          selectedCourseFilters: _selectedCourseFilters,
          selectedCallStatusFilters: _selectedCallStatusFilters,
          availabilityLoading: availabilityBusy,
          availabilityMessage: availabilityStatusMessage,
          slotAvailability: _slotAvailability,
          dragSlotAvailability: _dragSlotAvailability,
          dragCheckingSlotKeys: _dragCheckingSlotKeys,
          dragValidationActive: _dragValidationActive,
          creatingSlotKey: _creatingSlotKey,
          scheduleRows: _visibleScheduleRows,
          weekDays: _weekDays,
          timeSlots: _timeSlots,
          summary: _visibleTimetableSummary,
          errorMessage: _errorMessage,
          dateRange: _dateRangeText(_weekOffset),
          isCurrentWeek: _weekOffset == 0,
          onBack: () => Navigator.of(context).maybePop(),
          onPrevWeek: () => _changeWeek(-1),
          onNextWeek: () => _changeWeek(1),
          onToday: this._backToCurrentWeek,
          onPeriodGroupToggle: this._togglePeriodGroupDropdown,
          onPeriodGroupSelected: this._selectPeriodGroup,
          onPeriodGroupDropdownClose: this._closePeriodGroupDropdown,
          onTeacherToggle: () => setState(() {
            _teacherDropdownOpen = !_teacherDropdownOpen;
            _periodGroupDropdownOpen = false;
            _schedulePanelOpen = false;
            _openFilterKind = null;
          }),
          onTeacherSelected: this._selectTeacher,
          onTeacherDropdownClose: () =>
              setState(() => _teacherDropdownOpen = false),
          onSchedulePanelToggle: this._toggleSchedulePanel,
          onSchedulePanelClose: this._closeSchedulePanel,
          onFilterToggle: this._toggleFilterPanel,
          onFilterClose: this._closeFilterPanel,
          onStudentFilterToggled: this._toggleStudentFilter,
          onCourseFilterToggled: this._toggleCourseFilter,
          onCallStatusFilterToggled: this._toggleCallStatusFilter,
          onFilterCleared: this._clearFilter,
          onScheduleModeChanged: this._setScheduleMode,
          onScheduleTargetSelected: this._selectScheduleTarget,
          onScheduleTargetCleared: this._clearScheduleTarget,
          onAssistantToggled: this._toggleAssistant,
          onClassroomSelected: this._selectClassroom,
          onScheduleOptionsRefresh: _loadScheduleOptions,
          onAvailabilityRefresh: this._detectScheduleAvailability,
          onRefresh: _loadTimetable,
          onPrimaryScheduleTap: this._handlePrimaryScheduleTap,
          onLessonTap: this._openLessonDetail,
          onLessonMove: this._moveLesson,
          onLessonDragStarted: this._startLessonDrag,
          onLessonDragEnded: this._endLessonDrag,
          onEmptySlotTap: this._handleEmptySlotTap,
        ),
      ),
    );
  }
}
