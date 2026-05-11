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
  final String normalized = status?.trim() ?? '';
  if (normalized == 'confirmed') {
    return '已确认';
  }
  if (normalized == 'draft') {
    return '待确认';
  }
  return '待生成';
}

List<_DocDomainData> _docDomainsFromPlan(
  IepPlan plan, {
  bool defaultMissingCourseForm = true,
}) {
  final Map<String, List<IepPlanRow>> grouped = <String, List<IepPlanRow>>{};
  for (final IepPlanRow row in plan.rows) {
    final String domain = row.domain.trim().isEmpty ? '未分领域' : row.domain;
    grouped.putIfAbsent(domain, () => <IepPlanRow>[]).add(row);
  }
  return grouped.entries.map((MapEntry<String, List<IepPlanRow>> entry) {
    final List<String> longGoals = _normalizedNumberedTextLines(
      entry.value.map((IepPlanRow row) => row.longGoal),
    );
    final List<_DocShortGoalData> shortGoals =
        entry.value.map((IepPlanRow row) {
      final String courseForm = row.courseForm.trim();
      return _DocShortGoalData(
        _normalizeNumberedText(row.shortGoal),
        courseForm.isEmpty && defaultMissingCourseForm ? '个训' : courseForm,
        row.startEndDate,
      );
    }).toList();
    return _DocDomainData(
      domain: entry.key,
      longGoals: longGoals.isEmpty ? <String>[''] : longGoals,
      shortGoals: shortGoals.isEmpty
          ? <_DocShortGoalData>[
              _DocShortGoalData(
                '',
                defaultMissingCourseForm ? '个训' : '',
                '',
              ),
            ]
          : shortGoals,
    );
  }).toList();
}

_MonthDomainData _monthDomainFromPlanRow(IepMonthlyPlanRow row) {
  return _MonthDomainData(
    domain: row.domain,
    longGoal: _normalizeNumberedText(row.longGoal),
    shortGoal: _normalizeNumberedText(row.shortGoal),
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

List<String> _normalizedNumberedTextLines(Iterable<String> values) {
  final List<String> lines = <String>[];
  final String normalized = _normalizeNumberedText(values.join('\n'));
  final List<String> segments = normalized
      .split('\n')
      .map((String item) => item.trim())
      .where((String item) => item.isNotEmpty)
      .toList();
  for (final String segment in segments) {
    if (!lines.contains(segment)) {
      lines.add(segment);
    }
  }
  return lines;
}

String _normalizeNumberedText(String value) {
  final List<String> segments = value
      .replaceAll('\r\n', '\n')
      .replaceAll('\r', '\n')
      .split('\n')
      .map((String item) => item.trim())
      .where((String item) => item.isNotEmpty)
      .toList();
  final List<String> lines = <String>[];
  String? pendingMarker;
  for (final String segment in segments) {
    if (_isStandaloneNumberMarker(segment)) {
      if (pendingMarker != null) {
        lines.add(pendingMarker);
      }
      pendingMarker = segment;
      continue;
    }
    if (pendingMarker != null) {
      lines.add('$pendingMarker $segment');
      pendingMarker = null;
    } else {
      lines.add(segment);
    }
  }
  if (pendingMarker != null) {
    lines.add(pendingMarker);
  }
  return lines.join('\n');
}

bool _isStandaloneNumberMarker(String value) {
  final String text = value.trim();
  return RegExp(r'^(?:\d{1,2}|[（(]?\d{1,2}[）)]|\d{1,2}[\.．、])$')
          .hasMatch(text) ||
      RegExp(
        r'^(?:[一二三四五六七八九十]{1,3}|[（(]?[一二三四五六七八九十]{1,3}[）)]|[一二三四五六七八九十]{1,3}[\.．、])$',
      ).hasMatch(text);
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
