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

IepPlan? _streamingIepPlanFromText({
  required String text,
  required IepAssessmentRecordSummary record,
  required DateTime periodStart,
  required int durationMonths,
  IepPlan? fallbackPlan,
}) {
  final String content = text.trim();
  if (content.isEmpty) {
    return null;
  }
  try {
    final Object? parsed = jsonDecode(_extractCompleteJsonContent(content));
    if (parsed is Map) {
      final IepPlan plan = IepPlan.fromJson(Map<String, dynamic>.from(parsed));
      if (plan.rows.isNotEmpty) {
        return plan;
      }
    }
  } on Object {
    // The stream often contains an incomplete JSON object. Rows below are
    // parsed as soon as each object is complete.
  }

  final List<IepPlanRow> rows = _collectCompleteJsonObjects(
    _extractRowsArrayText(content),
  )
      .map((String raw) {
        try {
          final Object? decoded = jsonDecode(raw);
          return decoded is Map
              ? IepPlanRow.fromJson(Map<String, dynamic>.from(decoded))
              : null;
        } on Object {
          return null;
        }
      })
      .whereType<IepPlanRow>()
      .where((IepPlanRow row) {
        return row.domain.trim().isNotEmpty ||
            row.longGoal.trim().isNotEmpty ||
            row.shortGoal.trim().isNotEmpty;
      })
      .toList(growable: false);

  if (rows.isEmpty) {
    return null;
  }
  return IepPlan(
    title: _extractJsonStringField(content, 'title').trim().isNotEmpty
        ? _extractJsonStringField(content, 'title')
        : (fallbackPlan?.title.trim().isNotEmpty == true
            ? fallbackPlan!.title
            : (durationMonths == 6 ? '康复教学半年计划' : '康复教学季度计划')),
    student: IepPlanStudent(
      name: record.studentName,
      gender: record.studentGender,
      birthDate: record.birthDate,
    ),
    meta: IepPlanMeta(
      planDate: record.assessmentDate.trim().isNotEmpty
          ? record.assessmentDate
          : _formatDateDash(DateTime.now()),
      participant: record.examinerName,
      implementer: record.examinerName,
      startDate: _formatDateDash(_dateOnly(periodStart)),
      endDate: _formatDateDash(_periodEndFor(periodStart, durationMonths)),
    ),
    rows: rows,
  );
}

String _extractCompleteJsonContent(String text) {
  final int start = text.indexOf('{');
  final int end = text.lastIndexOf('}');
  if (start >= 0 && end > start) {
    return text.substring(start, end + 1);
  }
  return text;
}

String _extractRowsArrayText(String text) {
  final int keyIndex = text.indexOf('"rows"');
  if (keyIndex < 0) {
    return '';
  }
  final int arrayStart = text.indexOf('[', keyIndex);
  if (arrayStart < 0) {
    return '';
  }
  return text.substring(arrayStart + 1);
}

List<String> _collectCompleteJsonObjects(String text) {
  final List<String> objects = <String>[];
  int start = -1;
  int depth = 0;
  bool inString = false;
  bool escaped = false;
  for (int index = 0; index < text.length; index += 1) {
    final String char = text[index];
    if (inString) {
      if (escaped) {
        escaped = false;
      } else if (char == '\\') {
        escaped = true;
      } else if (char == '"') {
        inString = false;
      }
      continue;
    }
    if (char == '"') {
      inString = true;
      continue;
    }
    if (char == '{') {
      if (depth == 0) {
        start = index;
      }
      depth += 1;
      continue;
    }
    if (char == '}' && depth > 0) {
      depth -= 1;
      if (depth == 0 && start >= 0) {
        objects.add(text.substring(start, index + 1));
        start = -1;
      }
    }
  }
  return objects;
}

String _extractJsonStringField(String text, String key) {
  final RegExp pattern = RegExp(
    '"${RegExp.escape(key)}"\\s*:\\s*"((?:\\\\.|[^"\\\\])*)"',
  );
  final RegExpMatch? match = pattern.firstMatch(text);
  if (match == null) {
    return '';
  }
  final String raw = match.group(1) ?? '';
  try {
    return jsonDecode('"$raw"') as String;
  } on Object {
    return raw.replaceAll(r'\"', '"').replaceAll(r'\\', '\\');
  }
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
