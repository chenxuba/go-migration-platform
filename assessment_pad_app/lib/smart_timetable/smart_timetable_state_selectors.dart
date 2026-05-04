part of '../smart_timetable_page.dart';

extension _SmartTimetableStateSelectors on _SmartTimetablePageState {
  ScheduleTargetOption? get _selectedScheduleTarget {
    return _scheduleMode == _ScheduleMode.oneToOne
        ? _selectedOneToOneTarget
        : _selectedGroupClassTarget;
  }

  ScheduleTargetType get _scheduleTargetType {
    return _scheduleMode == _ScheduleMode.oneToOne
        ? ScheduleTargetType.oneToOne
        : ScheduleTargetType.groupClass;
  }

  List<ScheduleStaffOption> get _currentGroupAssistantOptions {
    final String teacherId = _selectedTeacherId.trim();
    final Map<String, ScheduleStaffOption> staffById =
        <String, ScheduleStaffOption>{
      for (final ScheduleStaffOption staff in _assistantOptions)
        if (staff.id.trim().isNotEmpty) staff.id.trim(): staff,
    };
    final List<ScheduleStaffOption> assistants = <ScheduleStaffOption>[];
    final Set<String> seenIds = <String>{};
    for (final _TeacherOption teacher in _teachers) {
      final String id = teacher.id.trim();
      if (id.isEmpty || id == teacherId || seenIds.contains(id)) {
        continue;
      }
      seenIds.add(id);
      final ScheduleStaffOption? staff = staffById[id];
      assistants.add(
        ScheduleStaffOption(
          id: id,
          name: (staff?.name.trim().isNotEmpty == true)
              ? staff!.name.trim()
              : teacher.name,
          subtitle: (staff?.subtitle.trim().isNotEmpty == true)
              ? staff!.subtitle.trim()
              : '当前时段组老师',
        ),
      );
    }
    return assistants;
  }

  List<String> get _normalizedAssistantIds {
    final String teacherId = _selectedTeacherId.trim();
    final Set<String> validAssistantIds = _currentGroupAssistantOptions
        .map((ScheduleStaffOption item) => item.id)
        .toSet();
    return _selectedAssistantIds
        .map((String id) => id.trim())
        .where(
          (String id) =>
              id.isNotEmpty &&
              id != teacherId &&
              validAssistantIds.contains(id),
        )
        .toList();
  }

  String get _selectedClassroomId => _selectedClassroom?.id.trim() ?? '';

  void _pruneSelectedAssistantsToCurrentGroup() {
    final Set<String> validAssistantIds = _currentGroupAssistantOptions
        .map((ScheduleStaffOption item) => item.id)
        .toSet();
    _selectedAssistantIds = _selectedAssistantIds
        .where((String id) => validAssistantIds.contains(id.trim()))
        .toSet();
  }

  void _resetAvailabilityFields({bool cancelPending = true}) {
    if (cancelPending) {
      _availabilitySequence += 1;
    }
    _availabilityLoading = false;
    _availabilityMessage = null;
    _slotAvailability = const <String, _SlotAvailability>{};
    _creatingSlotKey = null;
  }
}
