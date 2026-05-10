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
    final List<String> longGoals = entry.value
        .map((IepPlanRow row) => row.longGoal.trim())
        .where((String value) => value.isNotEmpty)
        .toSet()
        .toList();
    final List<_DocShortGoalData> shortGoals =
        entry.value.map((IepPlanRow row) {
      final String courseForm = row.courseForm.trim();
      return _DocShortGoalData(
        row.shortGoal,
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

  final List<_StreamingPlanRow> rows = _stableStreamingPlanRows(
    _collectStreamingJsonObjects(_extractRowsArrayText(content))
        .map((_StreamingJsonObject object) {
          final IepPlanRow? row = _streamingIepPlanRowFromObject(object);
          return row == null
              ? null
              : _StreamingPlanRow(row: row, complete: object.complete);
        })
        .whereType<_StreamingPlanRow>()
        .where((_StreamingPlanRow item) {
          final IepPlanRow row = item.row;
          return row.domain.trim().isNotEmpty ||
              row.longGoal.trim().isNotEmpty ||
              row.shortGoal.trim().isNotEmpty;
        })
        .toList(growable: false),
  );

  final List<IepPlanRow> planRows =
      rows.map((_StreamingPlanRow item) => item.row).where((IepPlanRow row) {
    return row.domain.trim().isNotEmpty ||
        row.longGoal.trim().isNotEmpty ||
        row.shortGoal.trim().isNotEmpty;
  }).toList(growable: false);

  if (planRows.isEmpty) {
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
    rows: planRows,
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

class _StreamingJsonObject {
  const _StreamingJsonObject({required this.raw, required this.complete});

  final String raw;
  final bool complete;
}

class _StreamingPlanRow {
  const _StreamingPlanRow({required this.row, required this.complete});

  final IepPlanRow row;
  final bool complete;
}

List<_StreamingPlanRow> _stableStreamingPlanRows(
  List<_StreamingPlanRow> rows,
) {
  final Map<String, List<String>> seenLongGoalsByDomain =
      <String, List<String>>{};
  final Set<String> seenDomains = <String>{};
  final List<_StreamingPlanRow> result = <_StreamingPlanRow>[];

  for (final _StreamingPlanRow item in rows) {
    IepPlanRow row = item.row;
    final String rawDomain = row.domain.trim();
    final String? repeatedDomain = !item.complete && rawDomain.isNotEmpty
        ? _matchingStreamingDomain(seenDomains, rawDomain)
        : null;
    if (repeatedDomain != null &&
        row.longGoal.trim().isEmpty &&
        row.shortGoal.trim().isEmpty &&
        row.courseForm.trim().isEmpty &&
        row.startEndDate.trim().isEmpty) {
      continue;
    }
    if (repeatedDomain != null && rawDomain != repeatedDomain) {
      row = IepPlanRow(
        domain: repeatedDomain,
        longGoal: row.longGoal,
        shortGoal: row.shortGoal,
        courseForm: row.courseForm,
        startEndDate: row.startEndDate,
      );
    }
    final String domainKey =
        row.domain.trim().isEmpty ? '未分领域' : row.domain.trim();
    final String longGoal = row.longGoal.trim();
    final List<String> seenLongGoals =
        seenLongGoalsByDomain.putIfAbsent(domainKey, () => <String>[]);

    if (!item.complete &&
        longGoal.isNotEmpty &&
        seenLongGoals
            .any((String seen) => _sameStreamingLongGoal(seen, longGoal))) {
      row = IepPlanRow(
        domain: row.domain,
        longGoal: '',
        shortGoal: row.shortGoal,
        courseForm: row.courseForm,
        startEndDate: row.startEndDate,
      );
    }

    if (seenDomains.contains(domainKey) &&
        row.longGoal.trim().isEmpty &&
        row.shortGoal.trim().isEmpty &&
        row.courseForm.trim().isEmpty &&
        row.startEndDate.trim().isEmpty) {
      continue;
    }

    final String stableLongGoal = row.longGoal.trim();
    if (stableLongGoal.isNotEmpty &&
        !seenLongGoals.any(
          (String seen) => _sameStreamingLongGoal(seen, stableLongGoal),
        )) {
      seenLongGoals.add(stableLongGoal);
    }
    seenDomains.add(domainKey);
    result.add(_StreamingPlanRow(row: row, complete: item.complete));
  }

  return result;
}

String? _matchingStreamingDomain(Set<String> seenDomains, String incoming) {
  for (final String seen in seenDomains) {
    if (_sameStreamingTextPrefix(seen, incoming)) {
      return seen;
    }
  }
  return null;
}

bool _sameStreamingLongGoal(String existing, String incoming) {
  return _sameStreamingTextPrefix(existing, incoming);
}

bool _sameStreamingTextPrefix(String existing, String incoming) {
  final String left = _compactGoalText(existing);
  final String right = _compactGoalText(incoming);
  if (left.isEmpty || right.isEmpty) {
    return false;
  }
  return left == right || left.startsWith(right) || right.startsWith(left);
}

String _compactGoalText(String text) {
  return text.replaceAll(RegExp(r'\s+'), '');
}

List<_StreamingJsonObject> _collectStreamingJsonObjects(String text) {
  final List<_StreamingJsonObject> objects = <_StreamingJsonObject>[];
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
        objects.add(
          _StreamingJsonObject(
            raw: text.substring(start, index + 1),
            complete: true,
          ),
        );
        start = -1;
      }
    }
  }
  if (start >= 0 && depth > 0) {
    objects.add(
      _StreamingJsonObject(raw: text.substring(start), complete: false),
    );
  }
  return objects;
}

IepPlanRow? _streamingIepPlanRowFromObject(_StreamingJsonObject object) {
  if (object.complete) {
    try {
      final Object? decoded = jsonDecode(object.raw);
      if (decoded is Map) {
        return IepPlanRow.fromJson(Map<String, dynamic>.from(decoded));
      }
    } on Object {
      // Fall through to tolerant parsing below.
    }
  }
  final String domain = _extractJsonStringField(
    object.raw,
    'domain',
    allowPartial: !object.complete,
  );
  final String longGoal = _extractJsonStringField(
    object.raw,
    'longGoal',
    allowPartial: !object.complete,
  );
  final String shortGoal = _extractJsonStringField(
    object.raw,
    'shortGoal',
    allowPartial: !object.complete,
  );
  final String courseForm = _extractJsonStringField(
    object.raw,
    'courseForm',
    allowPartial: !object.complete,
  );
  final String startEndDate = _extractJsonStringField(
    object.raw,
    'startEndDate',
    allowPartial: !object.complete,
  );
  if (domain.isEmpty &&
      longGoal.isEmpty &&
      shortGoal.isEmpty &&
      courseForm.isEmpty &&
      startEndDate.isEmpty) {
    return null;
  }
  return IepPlanRow(
    domain: domain,
    longGoal: longGoal,
    shortGoal: shortGoal,
    courseForm: courseForm,
    startEndDate: startEndDate,
  );
}

String _extractJsonStringField(
  String text,
  String key, {
  bool allowPartial = false,
}) {
  final RegExp pattern = RegExp(
    '"${RegExp.escape(key)}"\\s*:\\s*"',
  );
  final RegExpMatch? match = pattern.firstMatch(text);
  if (match == null) {
    return '';
  }
  final StringBuffer raw = StringBuffer();
  bool escaped = false;
  for (int index = match.end; index < text.length; index += 1) {
    final String char = text[index];
    if (escaped) {
      raw
        ..write('\\')
        ..write(char);
      escaped = false;
      continue;
    }
    if (char == '\\') {
      escaped = true;
      continue;
    }
    if (char == '"') {
      return _decodeJsonString(raw.toString());
    }
    raw.write(char);
  }
  if (!allowPartial) {
    return '';
  }
  if (escaped) {
    raw.write('\\');
  }
  return _decodeJsonString(raw.toString());
}

String _decodeJsonString(String raw) {
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
