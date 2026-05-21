part of 'vbmapp_assessment_page.dart';

Map<String, VbmappMaterialProfile> _itemMaterialProfileMapFromCatalog(
  VbmappMaterialCatalog? catalog,
) {
  if (catalog == null || catalog.items.isEmpty) {
    return const <String, VbmappMaterialProfile>{};
  }
  final Map<String, VbmappMaterialProfile> out =
      <String, VbmappMaterialProfile>{};
  for (final VbmappMaterialCatalogItem item in catalog.items) {
    if (item.itemCode.isEmpty) {
      continue;
    }
    out[item.itemCode] = item.toMaterialProfile();
  }
  return out;
}

class _VbmappScoreSnapshot {
  const _VbmappScoreSnapshot({
    required this.milestoneTotal,
    required this.milestoneMax,
    required this.barrierTotal,
    required this.barrierMax,
    required this.transitionTotal,
    required this.transitionMax,
    required this.milestoneDomains,
  });

  final double milestoneTotal;
  final int milestoneMax;
  final int barrierTotal;
  final int barrierMax;
  final int transitionTotal;
  final int transitionMax;
  final List<_VbmappDomainScoreSummary> milestoneDomains;

  String get milestoneScoreText {
    return milestoneTotal.toStringAsFixed(1);
  }
}

class _VbmappDomainScoreSummary {
  const _VbmappDomainScoreSummary({
    required this.name,
    required this.score,
    required this.maxScore,
    required this.answered,
    required this.total,
  });

  final String name;
  final double score;
  final int maxScore;
  final int answered;
  final int total;

  double get percent {
    if (maxScore <= 0) {
      return 0;
    }
    return (score / maxScore).clamp(0, 1).toDouble();
  }

  String get scoreText {
    return score.toStringAsFixed(1);
  }
}

class _VbmappModule {
  const _VbmappModule({
    required this.code,
    required this.title,
    required this.subtitle,
    required this.itemCount,
    required this.icon,
    required this.color,
  });

  final String code;
  final String title;
  final String subtitle;
  final int itemCount;
  final IconData icon;
  final Color color;
}

class _VbmappItem {
  const _VbmappItem({
    required this.sequenceNo,
    required this.moduleCode,
    required this.itemCode,
    required this.label,
    required this.domainName,
    required this.ageBand,
    required this.assessmentMode,
    required this.title,
    required this.scoreTitle,
    required this.scoreOptions,
    required this.materialHint,
    required this.color,
  });

  final int sequenceNo;
  final String moduleCode;
  final String itemCode;
  final String label;
  final String domainName;
  final String ageBand;
  final String assessmentMode;
  final String title;
  final String scoreTitle;
  final List<_VbmappScoreOption> scoreOptions;
  final String materialHint;
  final Color color;

  int get localNo {
    switch (moduleCode) {
      case 'barriers':
        return sequenceNo - 170;
      case 'transition':
        return sequenceNo - 194;
      case 'milestones':
      default:
        return sequenceNo;
    }
  }

  String get navCode {
    if (moduleCode == 'milestones') {
      final RegExpMatch? labelMatch =
          RegExp(r'(\d+)\s*-\s*M').firstMatch(label);
      if (labelMatch != null) {
        return '${labelMatch.group(1)}M';
      }
      final RegExpMatch? codeMatch = RegExp(r'_(\d+)M$').firstMatch(itemCode);
      if (codeMatch != null) {
        return '${int.parse(codeMatch.group(1)!)}M';
      }
      return '${localNo}M';
    }
    return itemCode;
  }
}

class _VbmappMandEvent {
  const _VbmappMandEvent({
    required this.utterance,
    required this.target,
    required this.motivationContext,
    this.environment = '',
    this.targetKind = '',
    required this.person,
    required this.setting,
    required this.example,
    required this.responseMode,
    required this.promptLevel,
    this.phraseLevel = '',
    this.recordedAtIso = '',
    this.sourceItemCode = '',
    required this.functional,
  });

  factory _VbmappMandEvent.fromJson(Map<String, dynamic> json) {
    return _VbmappMandEvent(
      utterance: _safeText(json['utterance']),
      target: _safeText(json['target']),
      motivationContext: _safeText(json['motivationContext']),
      environment: _safeText(json['environment']),
      targetKind: _safeText(json['targetKind']),
      person: _safeText(json['person']),
      setting: _safeText(json['setting']),
      example: _safeText(json['example']),
      responseMode: _safeText(json['responseMode']),
      promptLevel: _safeText(json['promptLevel']),
      phraseLevel: _safeText(
        json['phraseLevel'] ?? json['languageLevel'] ?? json['phrase_level'],
      ),
      recordedAtIso: _safeText(
        json['recordedAtIso'] ?? json['recorded_at'] ?? '',
      ),
      sourceItemCode: _safeText(
        json['sourceItemCode'] ?? json['source_item_code'] ?? '',
      ),
      functional: json['functional'] != false,
    );
  }

  final String utterance;
  final String target;
  final String motivationContext;
  final String environment;
  final String targetKind;
  final String person;
  final String setting;
  final String example;
  final String responseMode;
  final String promptLevel;
  final String phraseLevel;
  final String recordedAtIso;
  final String sourceItemCode;
  final bool functional;

  bool get isNotEmpty => utterance.isNotEmpty || target.isNotEmpty;

  DateTime? get recordedAt => _safeDateTimeParse(recordedAtIso);

  bool get hasPhysicalPrompt => promptLevel == '肢体辅助';

  bool get hasDisallowedPrompt =>
      hasPhysicalPrompt ||
      promptLevel == '口头辅助' ||
      promptLevel == '有口头辅助' ||
      promptLevel == '有额外辅助' ||
      promptLevel == '额外辅助' ||
      promptLevel == '其他辅助';

  bool get isQualified => functional && isNotEmpty && !hasDisallowedPrompt;

  String get uniqueKey {
    final String text = target.trim().isNotEmpty ? target : utterance;
    return text.trim().toLowerCase();
  }

  String get summary {
    final String spoken = utterance.trim().isEmpty ? '未记录表达' : utterance;
    final String targetText = target.trim().isEmpty ? '未记录目标' : target;
    final List<String> dimensions = <String>[
      if (person.trim().isNotEmpty) person.trim(),
      if (setting.trim().isNotEmpty) setting.trim(),
      if (example.trim().isNotEmpty) example.trim(),
      if (environment.trim().isNotEmpty) environment.trim(),
      if (targetKind.trim().isNotEmpty) targetKind.trim(),
      if (phraseLevel.trim().isNotEmpty) phraseLevel.trim(),
    ];
    final String dimensionText =
        dimensions.isEmpty ? '' : ' · ${dimensions.join('/')}';
    return '$spoken -> $targetText · $responseMode$dimensionText';
  }

  _VbmappMandEvent copyWith({
    String? utterance,
    String? target,
    String? motivationContext,
    String? environment,
    String? targetKind,
    String? person,
    String? setting,
    String? example,
    String? responseMode,
    String? promptLevel,
    String? phraseLevel,
    String? recordedAtIso,
    String? sourceItemCode,
    bool? functional,
  }) {
    return _VbmappMandEvent(
      utterance: utterance ?? this.utterance,
      target: target ?? this.target,
      motivationContext: motivationContext ?? this.motivationContext,
      environment: environment ?? this.environment,
      targetKind: targetKind ?? this.targetKind,
      person: person ?? this.person,
      setting: setting ?? this.setting,
      example: example ?? this.example,
      responseMode: responseMode ?? this.responseMode,
      promptLevel: promptLevel ?? this.promptLevel,
      phraseLevel: phraseLevel ?? this.phraseLevel,
      recordedAtIso: recordedAtIso ?? this.recordedAtIso,
      sourceItemCode: sourceItemCode ?? this.sourceItemCode,
      functional: functional ?? this.functional,
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'utterance': utterance,
      'target': target,
      'motivationContext': motivationContext,
      'environment': environment,
      'targetKind': targetKind,
      'person': person,
      'setting': setting,
      'example': example,
      'responseMode': responseMode,
      'promptLevel': promptLevel,
      'phraseLevel': phraseLevel,
      'recordedAtIso': recordedAtIso,
      'sourceItemCode': sourceItemCode,
      'functional': functional,
    };
  }
}

class _VbmappObservationTimerState {
  const _VbmappObservationTimerState({
    this.plannedMinutes = 60,
    this.accumulatedSeconds = 0,
    this.runningSinceIso = '',
    this.startedAtIso = '',
    this.ended = false,
  });

  factory _VbmappObservationTimerState.fromJson(Map<String, dynamic> json) {
    int intFrom(Object? value) {
      if (value is num) {
        return value.toInt();
      }
      return int.tryParse(_safeText(value)) ?? 0;
    }

    final int plannedMinutes =
        intFrom(json['plannedMinutes'] ?? json['planned_minutes']);
    final int accumulatedSeconds = intFrom(
      json['accumulatedSeconds'] ??
          json['accumulated_seconds'] ??
          json['actualObservationSeconds'],
    );
    return _VbmappObservationTimerState(
      plannedMinutes: plannedMinutes <= 0 ? 60 : plannedMinutes,
      accumulatedSeconds: accumulatedSeconds < 0 ? 0 : accumulatedSeconds,
      runningSinceIso: _safeText(
        json['runningSinceIso'] ?? json['running_since'] ?? '',
      ),
      startedAtIso: _safeText(
        json['startedAtIso'] ?? json['startTime'] ?? json['start_time'] ?? '',
      ),
      ended: json['ended'] == true || _safeText(json['status']) == 'ended',
    );
  }

  final int plannedMinutes;
  final int accumulatedSeconds;
  final String runningSinceIso;
  final String startedAtIso;
  final bool ended;

  bool get isRunning => runningSinceIso.trim().isNotEmpty;

  bool get hasStarted =>
      startedAtIso.trim().isNotEmpty || accumulatedSeconds > 0 || isRunning;

  bool get isEmpty => !hasStarted && !ended;

  int get plannedSeconds => plannedMinutes * Duration.secondsPerMinute;

  DateTime? get runningSince => _safeDateTimeParse(runningSinceIso);

  DateTime? get startedAt => _safeDateTimeParse(startedAtIso);

  int elapsedSecondsAt(DateTime now) {
    if (!isRunning) {
      return accumulatedSeconds;
    }
    final DateTime? since = runningSince;
    if (since == null) {
      return accumulatedSeconds;
    }
    final int delta = now.difference(since).inSeconds;
    return accumulatedSeconds + (delta > 0 ? delta : 0);
  }

  _VbmappObservationTimerState start(DateTime now) {
    if (isRunning) {
      return this;
    }
    final String iso = now.toIso8601String();
    return _VbmappObservationTimerState(
      plannedMinutes: plannedMinutes,
      accumulatedSeconds: accumulatedSeconds,
      runningSinceIso: iso,
      startedAtIso: startedAtIso.trim().isEmpty ? iso : startedAtIso,
      ended: false,
    );
  }

  _VbmappObservationTimerState resume(DateTime now) {
    return start(now);
  }

  _VbmappObservationTimerState pause(DateTime now) {
    return _VbmappObservationTimerState(
      plannedMinutes: plannedMinutes,
      accumulatedSeconds: elapsedSecondsAt(now),
      runningSinceIso: '',
      startedAtIso: startedAtIso,
      ended: false,
    );
  }

  _VbmappObservationTimerState finish(DateTime now) {
    return _VbmappObservationTimerState(
      plannedMinutes: plannedMinutes,
      accumulatedSeconds: elapsedSecondsAt(now),
      runningSinceIso: '',
      startedAtIso: startedAtIso,
      ended: true,
    );
  }

  _VbmappObservationTimerState restart(DateTime now) {
    final String iso = now.toIso8601String();
    return _VbmappObservationTimerState(
      plannedMinutes: plannedMinutes,
      accumulatedSeconds: 0,
      runningSinceIso: iso,
      startedAtIso: iso,
      ended: false,
    );
  }

  _VbmappObservationTimerState withPlannedMinutes(int value) {
    if (value <= 0 || value == plannedMinutes) {
      return this;
    }
    return _VbmappObservationTimerState(
      plannedMinutes: value,
      accumulatedSeconds: accumulatedSeconds,
      runningSinceIso: runningSinceIso,
      startedAtIso: startedAtIso,
      ended: ended,
    );
  }

  Map<String, dynamic> toJson() {
    return <String, dynamic>{
      'plannedMinutes': plannedMinutes,
      'planned_seconds': plannedSeconds,
      'accumulatedSeconds': accumulatedSeconds,
      'runningSinceIso': runningSinceIso,
      'startTime': startedAtIso,
      'status': ended
          ? 'ended'
          : isRunning
              ? 'running'
              : hasStarted
                  ? 'paused'
                  : 'idle',
      'ended': ended,
    };
  }
}

class _VbmappScoreOption {
  const _VbmappScoreOption({
    required this.score,
    required this.label,
  });

  final num score;
  final String label;

  String get displayScore {
    if (score is int || score == score.roundToDouble()) {
      return score.toInt().toString();
    }
    return score.toString();
  }
}

String _sessionExaminerName(HomeSession session) {
  if (session.nickName.trim().isNotEmpty) {
    return session.nickName.trim();
  }
  if (session.username.trim().isNotEmpty) {
    return session.username.trim();
  }
  return '';
}

String _todayIsoDate() {
  return _dateOnlyText(DateTime.now());
}

String _formatClock(DateTime value) {
  return '${value.hour.toString().padLeft(2, '0')}:'
      '${value.minute.toString().padLeft(2, '0')}';
}

String _vbmappDurationText(int totalSeconds) {
  final int seconds = totalSeconds < 0 ? 0 : totalSeconds;
  final int minutesPart = seconds ~/ Duration.secondsPerMinute;
  final int secondsPart = seconds % Duration.secondsPerMinute;
  return '${minutesPart.toString().padLeft(2, '0')}:'
      '${secondsPart.toString().padLeft(2, '0')}';
}

bool _isSimpleMandRecorder(
  _VbmappItem item,
  VbmappItemResponseSchema? schema,
) {
  if (schema?.uiPattern != 'mand_event_recorder') {
    return false;
  }
  return item.itemCode == 'MAND_01M' ||
      item.itemCode == 'MAND_02M' ||
      item.itemCode == 'MAND_03M' ||
      item.itemCode == 'MAND_04M';
}

bool _isTimedMandItemCode(String itemCode) {
  return itemCode == 'MAND_04M' ||
      itemCode == 'MAND_08M' ||
      itemCode == 'MAND_09M';
}

int _plannedMinutesForMandItem(_VbmappItem item) {
  switch (item.itemCode) {
    case 'MAND_09M':
      return 30;
    case 'MAND_04M':
    case 'MAND_08M':
    default:
      return 60;
  }
}

String _mandInitiationText(_VbmappMandEvent event) {
  final String prompt = event.promptLevel.trim();
  if (prompt == '提问下' || prompt == '自发地') {
    return prompt;
  }
  if (event.responseMode.contains('自发')) {
    return '自发地';
  }
  return '';
}

bool _mandEventWithinWindow(
  _VbmappMandEvent event,
  _VbmappObservationTimerState? observation, {
  required int plannedMinutes,
}) {
  final DateTime? startedAt = observation?.startedAt;
  final DateTime? recordedAt = event.recordedAt;
  if (startedAt == null || recordedAt == null) {
    return true;
  }
  if (recordedAt.isBefore(startedAt)) {
    return false;
  }
  return !recordedAt.isAfter(
    startedAt.add(Duration(minutes: plannedMinutes)),
  );
}

bool _mandEventCountsForItem(
  _VbmappItem item,
  _VbmappMandEvent event, {
  _VbmappObservationTimerState? observation,
}) {
  if (!event.isQualified) {
    return false;
  }
  if (_isTimedMandItemCode(item.itemCode) &&
      !_mandEventWithinWindow(
        event,
        observation,
        plannedMinutes: _plannedMinutesForMandItem(item),
      )) {
    return false;
  }
  switch (item.itemCode) {
    case 'MAND_05M':
      return event.environment.trim() == '呈现物品' &&
          _mandInitiationText(event) != '提问下';
    case 'MAND_04M':
      return event.environment.trim() == '呈现物品' &&
          _mandInitiationText(event) != '提问下';
    case 'MAND_08M':
      return true;
    case 'MAND_09M':
      return _mandInitiationText(event) != '提问下';
    default:
      return event.isQualified;
  }
}

int _qualifiedMandCount(List<_VbmappMandEvent> events) {
  final Set<String> uniqueTargets = <String>{};
  for (final _VbmappMandEvent event in events) {
    if (event.isQualified && event.uniqueKey.isNotEmpty) {
      uniqueTargets.add(event.uniqueKey);
    }
  }
  return uniqueTargets.length;
}

int _qualifiedMandCountForItem(
  _VbmappItem item,
  List<_VbmappMandEvent> events, {
  _VbmappObservationTimerState? observation,
}) {
  final Set<String> uniqueTargets = <String>{};
  for (final _VbmappMandEvent event in events) {
    if (_mandEventCountsForItem(item, event, observation: observation) &&
        event.uniqueKey.isNotEmpty) {
      uniqueTargets.add(event.uniqueKey);
    }
  }
  return uniqueTargets.length;
}

int _mandPhraseQualifiedCountForItem(
  _VbmappItem item,
  List<_VbmappMandEvent> events, {
  _VbmappObservationTimerState? observation,
}) {
  final Set<String> uniqueTargets = <String>{};
  for (final _VbmappMandEvent event in events) {
    if (!_mandEventCountsForItem(item, event, observation: observation)) {
      continue;
    }
    if (!_isLikelyMultiWordMand(event)) {
      continue;
    }
    if (event.uniqueKey.isNotEmpty) {
      uniqueTargets.add(event.uniqueKey);
    }
  }
  return uniqueTargets.length;
}

bool _isLikelyMultiWordMand(_VbmappMandEvent event) {
  final String explicitLevel = event.phraseLevel.trim();
  if (explicitLevel == '双词+') {
    return true;
  }
  if (explicitLevel == '单词') {
    return false;
  }

  String text = _mandRequestText(event)
      .replaceAll(RegExp(r'[，。！？、,.!?;；:/\\]+'), ' ')
      .replaceAll(RegExp(r'\s+'), ' ')
      .trim();
  if (text.isEmpty) {
    return false;
  }

  final List<String> spacedTokens = text
      .split(' ')
      .map((String part) => part.trim())
      .where((String part) => part.isNotEmpty)
      .toList(growable: false);
  if (spacedTokens.length >= 2) {
    return true;
  }

  text = text.replaceFirst(RegExp(r'^我想要'), '').trim();
  if (text.isEmpty) {
    return false;
  }

  if (RegExp(r'(我|你|他|她|它|我们|我要|给我|帮我)').hasMatch(text) &&
      text.runes.length >= 3) {
    return true;
  }
  if (RegExp(r'(快点|一下|一会|给我|帮我|让我|一起|该我了|不要|好了|开门)$').hasMatch(text)) {
    return true;
  }
  if (text.runes.length >= 3 &&
      RegExp(r'^(打开|帮我|给我|让我|带我|一起|还要|再来|我要|我们|该我|跑|倒|推|拿|去|来|开)')
          .hasMatch(text)) {
    return true;
  }

  return false;
}

String _mandRequestText(_VbmappMandEvent event) {
  final String utterance = event.utterance.trim();
  if (utterance.isNotEmpty) {
    return utterance;
  }
  final String target = event.target.trim();
  if (target.isNotEmpty) {
    return target;
  }
  return '未记录要求';
}

String _mandRecordMetaText(
  _VbmappMandEvent event, {
  _VbmappItem? item,
}) {
  final List<String> values = <String>[];

  void addMeta(String raw) {
    final String value = raw.trim();
    if (value.isEmpty || values.contains(value)) {
      return;
    }
    values.add(value);
  }

  addMeta(_mandInitiationText(event));
  addMeta(event.environment);
  addMeta(event.targetKind);
  addMeta(event.phraseLevel);
  addMeta(event.promptLevel);
  if (item != null && _isTimedMandItemCode(item.itemCode)) {
    final DateTime? recordedAt = event.recordedAt;
    if (recordedAt != null) {
      values.add(_formatClock(recordedAt));
    }
  }
  return values.isEmpty ? '未记录条件' : values.join(' · ');
}

int _effectiveObservationSecondsForItem(
  _VbmappItem item,
  _VbmappObservationTimerState? observation,
) {
  if (observation == null) {
    return 0;
  }
  final int elapsed = observation.elapsedSecondsAt(DateTime.now());
  final int maxSeconds =
      _plannedMinutesForMandItem(item) * Duration.secondsPerMinute;
  return elapsed > maxSeconds ? maxSeconds : elapsed;
}

String _mandTimedScoreBasisText(
  _VbmappItem item,
  double suggestedScore,
  int qualifiedCount,
  int actualObservationSeconds,
  int multiWordCount,
) {
  final String baseDuration = '${_plannedMinutesForMandItem(item)}分钟观察窗';
  switch (item.itemCode) {
    case 'MAND_08M':
      return '系统按$baseDuration内的不同要求数量建议${_formatScore(suggestedScore)}分，'
          '当前计入$qualifiedCount条，其中双词+$multiWordCount条，'
          '已观察${_vbmappDurationText(actualObservationSeconds)}，老师可在下方评分区覆盖。';
    case 'MAND_09M':
      return '系统按$baseDuration内的自发不同要求数量建议${_formatScore(suggestedScore)}分，'
          '当前计入$qualifiedCount条，已观察${_vbmappDurationText(actualObservationSeconds)}，'
          '老师可在下方评分区覆盖。';
    case 'MAND_04M':
    default:
      return '系统按$baseDuration内的有效自发要求数量建议${_formatScore(suggestedScore)}分，'
          '当前计入$qualifiedCount条，已观察${_vbmappDurationText(actualObservationSeconds)}，'
          '老师可在下方评分区覆盖。';
  }
}

double _suggestMandScore(
  _VbmappItem item,
  List<_VbmappMandEvent> events, {
  _VbmappObservationTimerState? observation,
}) {
  if (item.itemCode == 'MAND_03M') {
    final Map<String, int> counts = _mandGeneralizationCounts(events);
    final bool onePoint = (counts['people'] ?? 0) >= 2 &&
        (counts['settings'] ?? 0) >= 2 &&
        (counts['examples'] ?? 0) >= 2;
    if (onePoint) {
      return 1;
    }
    final bool halfPoint = (counts['people'] ?? 0) >= 1 &&
        (counts['settings'] ?? 0) >= 1 &&
        (counts['examples'] ?? 0) >= 1;
    return halfPoint ? .5 : 0;
  }
  final int count = _qualifiedMandCountForItem(
    item,
    events,
    observation: observation,
  );
  if (item.itemCode == 'MAND_08M') {
    final int multiWordCount = _mandPhraseQualifiedCountForItem(
      item,
      events,
      observation: observation,
    );
    if (count >= 5 && multiWordCount >= 2) {
      return 1;
    }
    if (count >= 2) {
      return .5;
    }
    return 0;
  }
  final int onePointCount = _scoreCountThreshold(item, 1) ?? 1;
  final int halfPointCount = _scoreCountThreshold(item, .5) ?? onePointCount;
  if (count >= onePointCount) {
    return 1;
  }
  if (count >= halfPointCount) {
    return .5;
  }
  return 0;
}

Map<String, int> _mandGeneralizationCounts(List<_VbmappMandEvent> events) {
  final Map<String, List<String>> values = _mandGeneralizationValues(events);
  return <String, int>{
    'people': values['people']?.length ?? 0,
    'settings': values['settings']?.length ?? 0,
    'examples': values['examples']?.length ?? 0,
  };
}

Map<String, List<String>> _mandGeneralizationValues(
  List<_VbmappMandEvent> events,
) {
  final Set<String> people = <String>{};
  final Set<String> settings = <String>{};
  final Set<String> examples = <String>{};
  final List<String> peopleValues = <String>[];
  final List<String> settingValues = <String>[];
  final List<String> exampleValues = <String>[];
  for (final _VbmappMandEvent event in events) {
    if (!event.isQualified) {
      continue;
    }
    if (event.person.trim().isNotEmpty) {
      final String value = event.person.trim();
      final String normalized = value.toLowerCase();
      if (people.add(normalized)) {
        peopleValues.add(value);
      }
    }
    if (event.setting.trim().isNotEmpty) {
      final String value = event.setting.trim();
      final String normalized = value.toLowerCase();
      if (settings.add(normalized)) {
        settingValues.add(value);
      }
    }
    if (event.example.trim().isNotEmpty) {
      final String value = event.example.trim();
      final String normalized = value.toLowerCase();
      if (examples.add(normalized)) {
        exampleValues.add(value);
      }
    }
  }
  return <String, List<String>>{
    'people': peopleValues,
    'settings': settingValues,
    'examples': exampleValues,
  };
}

String _mand3DimensionText(_VbmappMandEvent event) {
  if (event.person.trim().isNotEmpty) {
    return '人物：${event.person.trim()}';
  }
  if (event.setting.trim().isNotEmpty) {
    return '环境：${event.setting.trim()}';
  }
  if (event.example.trim().isNotEmpty) {
    return '例子：${event.example.trim()}';
  }
  return '未记录';
}

int? _scoreCountThreshold(_VbmappItem item, num score) {
  for (final _VbmappScoreOption option in item.scoreOptions) {
    if (option.score == score) {
      final RegExpMatch? match =
          RegExp(r'[：:]\s*(\d+)').firstMatch(option.label) ??
              RegExp(r'(\d+)').firstMatch(option.label);
      if (match != null) {
        return int.tryParse(match.group(1)!);
      }
    }
  }
  return null;
}

String _schemaKey(String moduleCode, String itemCode) {
  return '${_safeText(moduleCode).toLowerCase()}::${_safeText(itemCode).toUpperCase()}';
}

String _formatScore(num score) {
  if (score == score.roundToDouble()) {
    return score.toInt().toString();
  }
  return score.toString();
}

Map<String, List<String>> _normalizedMaterialQuickPicks(
  Map<String, Object?> raw,
) {
  if (raw.isEmpty) {
    return const <String, List<String>>{};
  }
  final Map<String, List<String>> out = <String, List<String>>{};
  raw.forEach((String key, Object? value) {
    final List<String> values = _materialStringList(value);
    if (key.trim().isNotEmpty && values.isNotEmpty) {
      out[key.trim()] = values;
    }
  });
  return out;
}

List<String> _materialStringList(Object? raw) {
  if (raw is! List) {
    return const <String>[];
  }
  return raw
      .map((Object? value) => _safeText(value))
      .where((String value) => value.isNotEmpty)
      .toList(growable: false);
}

String _materialFieldLabel(String key) {
  if (key.contains('people')) {
    return '人物';
  }
  if (key.contains('settings')) {
    return '环境';
  }
  if (key.contains('examples')) {
    return '例子';
  }
  return '词库';
}

List<String> _deduplicatedTexts(List<String> values) {
  final Set<String> seen = <String>{};
  final List<String> out = <String>[];
  for (final String value in values) {
    final String normalized = value.trim();
    if (normalized.isEmpty || seen.contains(normalized)) {
      continue;
    }
    seen.add(normalized);
    out.add(normalized);
  }
  return out;
}

List<String> _smartMandQuickPicks(
  VbmappMaterialProfile? profile, {
  required String targetKind,
  required List<String> fallback,
  int limit = 8,
}) {
  final List<String> typed = <String>[
    for (final VbmappMaterialSuggestion material
        in profile?.recommendedMaterials ?? const <VbmappMaterialSuggestion>[])
      if (_materialMatchesMandTarget(
        name: material.name,
        type: material.type,
        targetKind: targetKind,
      ))
        material.name,
  ];
  final List<String> untyped = profile?.quickPicks ?? const <String>[];
  final List<String> targetFallback = _mandFallbackQuickPicksForTarget(
    targetKind,
  );
  final List<String> filteredFallback = fallback
      .where((String value) => _materialMatchesMandTarget(
            name: value,
            type: '',
            targetKind: targetKind,
          ))
      .toList(growable: false);
  final List<String> source = typed.isEmpty
      ? <String>[
          ...untyped,
          ...targetFallback,
          ...filteredFallback,
        ]
      : <String>[...typed, ...targetFallback, ...untyped, ...filteredFallback];
  return _deduplicatedTexts(source).take(limit).toList(growable: false);
}

List<String> _mandFallbackQuickPicksForTarget(String targetKind) {
  switch (targetKind.trim()) {
    case '动作':
      return const <String>['打开', '出去', '帮我', '给我', '推', '倒果汁'];
    case '活动':
      return const <String>['音乐', '秋千', '泡泡', '转圈', '一起玩', '出去'];
    case '物品':
    default:
      return const <String>['饼干', '书', '球', '泡泡', '车', '积木', '彩虹弹簧'];
  }
}

bool _materialMatchesMandTarget({
  required String name,
  required String type,
  required String targetKind,
}) {
  final String normalizedName = name.trim();
  final String normalizedType = type.trim();
  if (normalizedName.isEmpty) {
    return false;
  }
  switch (targetKind.trim()) {
    case '动作':
      return normalizedType.contains('动作') ||
          normalizedType.contains('帮助') ||
          _looksLikeActionMand(normalizedName);
    case '活动':
      return normalizedType.contains('活动') ||
          normalizedType.contains('社交游戏') ||
          _looksLikeActivityMand(normalizedName);
    case '物品':
    default:
      return !_looksLikeActionMand(normalizedName) &&
          !_looksLikeActivityMand(normalizedName);
  }
}

bool _looksLikeActionMand(String value) {
  return RegExp(r'(打开|出去|帮我|给我|推|倒|拿|开门|关门|再来|快点|该我了)').hasMatch(value.trim());
}

bool _looksLikeActivityMand(String value) {
  return RegExp(r'(音乐|秋千|泡泡|转圈|一起玩|游戏|出去)').hasMatch(value.trim());
}

Map<String, dynamic> _dynamicMap(Object? raw) {
  if (raw is Map) {
    final Map<String, dynamic> out = <String, dynamic>{};
    raw.forEach((Object? key, Object? value) {
      final String normalizedKey = _safeText(key);
      if (normalizedKey.isNotEmpty) {
        out[normalizedKey] = value;
      }
    });
    return out;
  }
  return <String, dynamic>{};
}

DateTime? _safeDateTimeParse(Object? value) {
  final String text = _safeText(value);
  if (text.isEmpty) {
    return null;
  }
  return DateTime.tryParse(text);
}

String _dateOnlyText(Object? value) {
  if (value is DateTime) {
    return '${value.year.toString().padLeft(4, '0')}-'
        '${value.month.toString().padLeft(2, '0')}-'
        '${value.day.toString().padLeft(2, '0')}';
  }
  final String text = _safeText(value);
  if (text.length >= 10) {
    return text.substring(0, 10);
  }
  return text;
}

String _safeText(Object? value) {
  if (value == null) {
    return '';
  }
  if (value is String) {
    return value.trim();
  }
  return '$value'.trim();
}
