part of 'iep_center_page.dart';

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
