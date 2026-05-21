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

class _VbmappMandPhraseAssessment {
  const _VbmappMandPhraseAssessment({
    required this.label,
    required this.isMultiWord,
    required this.normalizedText,
    required this.reason,
  });

  final String label;
  final bool isMultiWord;
  final String normalizedText;
  final String reason;
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

class _VbmappTimedMandStrategy {
  const _VbmappTimedMandStrategy({
    required this.itemCode,
    required this.plannedMinutes,
    required this.countMetricLabel,
    required this.inputLabel,
    required this.inputHint,
    required this.targetOptions,
    required this.defaultObservationHint,
    this.defaultPromptMode = '自发地',
    this.defaultPresentation = '呈现物品',
    this.defaultTargetKind = '物品',
    this.showPromptSelector = false,
    this.promptSelectorLabel = '诱发',
    this.promptOptions = const <String>['自发地', '提问下'],
    this.requirePresentedEnvironment = false,
    this.excludePromptedEvents = false,
    this.multiWordQualifiedMinCount = 0,
    this.displayMinSlots,
  });

  final String itemCode;
  final int plannedMinutes;
  final String countMetricLabel;
  final String inputLabel;
  final String inputHint;
  final List<String> targetOptions;
  final String defaultObservationHint;
  final String defaultPromptMode;
  final String defaultPresentation;
  final String defaultTargetKind;
  final bool showPromptSelector;
  final String promptSelectorLabel;
  final List<String> promptOptions;
  final bool requirePresentedEnvironment;
  final bool excludePromptedEvents;
  final int multiWordQualifiedMinCount;
  final int? displayMinSlots;

  int onePointCount(_VbmappItem item) => _scoreCountThreshold(item, 1) ?? 5;

  int halfPointCount(_VbmappItem item) => _scoreCountThreshold(item, .5) ?? 2;

  int resolvedDisplayMinSlots(_VbmappItem item) {
    return displayMinSlots ?? onePointCount(item);
  }

  String responseModeForPrompt(String promptMode) {
    if (showPromptSelector && promptMode == '提问下') {
      return '提问下要求';
    }
    return '自发要求';
  }

  String promptLevelForPrompt(String promptMode) {
    return showPromptSelector ? promptMode : '';
  }

  String uniqueKeyForEvent(_VbmappMandEvent event) {
    if (itemCode == 'MAND_09M') {
      return _normalizeMandDistinctKey(event);
    }
    return _defaultMandUniqueKey(event);
  }

  String scoreReference(_VbmappItem item) {
    final int halfPoint = halfPointCount(item);
    final int onePoint = onePointCount(item);
    if (multiWordQualifiedMinCount > 0) {
      return '参考：${plannedMinutes}分钟观察窗内，$halfPoint个不同要求计0.5分，'
          '$onePoint个不同要求且系统判定至少$multiWordQualifiedMinCount条为双词+计1分。';
    }
    if (excludePromptedEvents) {
      return '参考：${plannedMinutes}分钟观察窗内，$halfPoint个自发不同要求计0.5分，'
          '$onePoint个自发不同要求计1分。';
    }
    if (requirePresentedEnvironment) {
      return '参考：${plannedMinutes}分钟观察窗内，$halfPoint个计0.5分，'
          '$onePoint个计1分；仅统计呈现物品条件下的自发要求。';
    }
    return '参考：${plannedMinutes}分钟观察窗内，$halfPoint个计0.5分，'
        '$onePoint个计1分。';
  }

  String observationHint({
    required _VbmappItem item,
    required List<_VbmappMandEvent> events,
    required _VbmappObservationTimerState? observation,
  }) {
    final _VbmappObservationTimerState timerState =
        (observation ?? const _VbmappObservationTimerState())
            .withPlannedMinutes(plannedMinutes);
    final int qualifiedCount = _qualifiedMandCountForItem(
      item,
      events,
      observation: observation,
    );
    final bool observationMet = timerState.elapsedSecondsAt(DateTime.now()) >=
        plannedMinutes * Duration.secondsPerMinute;
    if (!timerState.hasStarted) {
      return '先开启${plannedMinutes}分钟观察窗，再连续记录孩子的自然要求。';
    }
    if (!observationMet && qualifiedCount >= onePointCount(item)) {
      if (multiWordQualifiedMinCount > 0) {
        return '已达到数量条件，但观察未满${plannedMinutes}分钟；双词结构由系统根据原话自动判定。';
      }
      return '已达到1分数量条件，但观察未满${plannedMinutes}分钟，可继续观察或由老师确认结束。';
    }
    if (!observationMet && qualifiedCount >= halfPointCount(item)) {
      return '已达到0.5分数量条件，但观察未满${plannedMinutes}分钟，建议继续观察。';
    }
    if (observationMet) {
      return '${plannedMinutes}分钟观察窗已完成，系统会继续保留新增记录供老师判断。';
    }
    return defaultObservationHint;
  }

  int qualifiedCount(
    _VbmappItem item,
    List<_VbmappMandEvent> events, {
    _VbmappObservationTimerState? observation,
  }) {
    final Set<String> uniqueTargets = <String>{};
    for (final _VbmappMandEvent event in events) {
      if (!countsEvent(item, event, observation: observation)) {
        continue;
      }
      final String uniqueKey = uniqueKeyForEvent(event);
      if (uniqueKey.isNotEmpty) {
        uniqueTargets.add(uniqueKey);
      }
    }
    return uniqueTargets.length;
  }

  int qualifiedMultiWordCount(
    _VbmappItem item,
    List<_VbmappMandEvent> events, {
    _VbmappObservationTimerState? observation,
  }) {
    final Set<String> uniqueTargets = <String>{};
    for (final _VbmappMandEvent event in events) {
      if (!countsEvent(item, event, observation: observation) ||
          !_isLikelyMultiWordMand(event)) {
        continue;
      }
      final String uniqueKey = uniqueKeyForEvent(event);
      if (uniqueKey.isNotEmpty) {
        uniqueTargets.add(uniqueKey);
      }
    }
    return uniqueTargets.length;
  }

  bool countsEvent(
    _VbmappItem item,
    _VbmappMandEvent event, {
    _VbmappObservationTimerState? observation,
  }) {
    if (!event.isQualified) {
      return false;
    }
    if (!_mandEventWithinWindow(
      event,
      observation,
      plannedMinutes: plannedMinutes,
    )) {
      return false;
    }
    if (requirePresentedEnvironment && event.environment.trim() != '呈现物品') {
      return false;
    }
    if (excludePromptedEvents && _mandInitiationText(event) == '提问下') {
      return false;
    }
    return true;
  }

  String recordMetaText(
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
    if (multiWordQualifiedMinCount > 0) {
      addMeta(_assessMandPhrase(event).label);
    } else {
      addMeta(event.phraseLevel);
    }
    if (item?.itemCode == 'MAND_09M') {
      final String uniqueKey = uniqueKeyForEvent(event);
      if (uniqueKey.isNotEmpty && uniqueKey != _mandRequestText(event).trim()) {
        addMeta('归一:$uniqueKey');
      }
    }
    addMeta(event.promptLevel);
    final DateTime? recordedAt = event.recordedAt;
    if (recordedAt != null) {
      values.add(_formatClock(recordedAt));
    }
    return values.isEmpty ? '未记录条件' : values.join(' · ');
  }

  double suggestScore(
    _VbmappItem item,
    List<_VbmappMandEvent> events, {
    _VbmappObservationTimerState? observation,
  }) {
    final int qualifiedCountValue =
        qualifiedCount(item, events, observation: observation);
    if (multiWordQualifiedMinCount > 0) {
      final int multiWordCount =
          qualifiedMultiWordCount(item, events, observation: observation);
      if (qualifiedCountValue >= onePointCount(item) &&
          multiWordCount >= multiWordQualifiedMinCount) {
        return 1;
      }
      if (qualifiedCountValue >= halfPointCount(item)) {
        return .5;
      }
      return 0;
    }
    if (qualifiedCountValue >= onePointCount(item)) {
      return 1;
    }
    if (qualifiedCountValue >= halfPointCount(item)) {
      return .5;
    }
    return 0;
  }

  String scoreBasisText(
    _VbmappItem item,
    double suggestedScore,
    int qualifiedCount,
    int actualObservationSeconds,
    int multiWordCount,
  ) {
    final String baseDuration = '${plannedMinutes}分钟观察窗';
    if (multiWordQualifiedMinCount > 0) {
      return '系统按$baseDuration内的不同要求数量建议${_formatScore(suggestedScore)}分，'
          '当前计入$qualifiedCount条，其中双词+$multiWordCount条，'
          '已观察${_vbmappDurationText(actualObservationSeconds)}，老师可在下方评分区覆盖。';
    }
    if (excludePromptedEvents) {
      return '系统按$baseDuration内的自发不同要求数量建议${_formatScore(suggestedScore)}分，'
          '当前计入$qualifiedCount条，已观察${_vbmappDurationText(actualObservationSeconds)}，'
          '老师可在下方评分区覆盖。';
    }
    return '系统按$baseDuration内的有效自发要求数量建议${_formatScore(suggestedScore)}分，'
        '当前计入$qualifiedCount条，已观察${_vbmappDurationText(actualObservationSeconds)}，'
        '老师可在下方评分区覆盖。';
  }

  Map<String, dynamic> buildEvidenceFields(
    _VbmappItem item,
    List<_VbmappMandEvent> events,
    double suggestedScore, {
    _VbmappObservationTimerState? observation,
  }) {
    final _VbmappObservationTimerState? timerState = observation;
    final int qualifiedCountValue =
        qualifiedCount(item, events, observation: observation);
    final int actualObservationSeconds =
        timerState?.elapsedSecondsAt(DateTime.now()) ?? 0;
    final int effectiveObservationSeconds =
        _effectiveObservationSecondsForItem(item, timerState);
    final int multiWordCount = multiWordQualifiedMinCount > 0
        ? qualifiedMultiWordCount(item, events, observation: observation)
        : 0;
    final List<Map<String, dynamic>> phraseEvaluations =
        multiWordQualifiedMinCount > 0
            ? events.map(((_VbmappMandEvent event) {
                final _VbmappMandPhraseAssessment assessment =
                    _assessMandPhrase(event);
                return <String, dynamic>{
                  'utterance': _mandRequestText(event),
                  'label': assessment.label,
                  'isMultiWord': assessment.isMultiWord,
                  'normalizedText': assessment.normalizedText,
                  'reason': assessment.reason,
                };
              })).toList(growable: false)
            : const <Map<String, dynamic>>[];
    final List<Map<String, dynamic>> eventUniqueKeys = events
        .map(((_VbmappMandEvent event) => <String, dynamic>{
              'utterance': _mandRequestText(event),
              'uniqueKey': uniqueKeyForEvent(event),
              'counts': countsEvent(item, event, observation: observation),
            }))
        .toList(growable: false);
    final List<String> uniqueTargetKeys = <String>{
      for (final _VbmappMandEvent event in events)
        if (countsEvent(item, event, observation: observation) &&
            uniqueKeyForEvent(event).isNotEmpty)
          uniqueKeyForEvent(event),
    }.toList(growable: false);
    return <String, dynamic>{
      'qualifiedCount': qualifiedCountValue,
      'uniqueTargetCount': qualifiedCountValue,
      'uniqueTargetKeys': uniqueTargetKeys,
      'eventUniqueKeys': eventUniqueKeys,
      if (timerState != null) 'timer': timerState.toJson(),
      if (timerState != null)
        'actualObservationMinutes':
            actualObservationSeconds / Duration.secondsPerMinute,
      if (timerState != null)
        'actualObservationSeconds': actualObservationSeconds,
      if (timerState != null)
        'effectiveObservationSeconds': effectiveObservationSeconds,
      if (timerState != null)
        'effectiveObservationMinutes':
            effectiveObservationSeconds / Duration.secondsPerMinute,
      if (multiWordQualifiedMinCount > 0)
        'multiWordQualifiedCount': multiWordCount,
      if (multiWordQualifiedMinCount > 0)
        'phraseAssessments': phraseEvaluations,
      'scoreBasis': scoreBasisText(
        item,
        suggestedScore,
        qualifiedCountValue,
        actualObservationSeconds,
        multiWordCount,
      ),
    };
  }
}

const Map<String, _VbmappTimedMandStrategy> _vbmappTimedMandStrategies =
    <String, _VbmappTimedMandStrategy>{
  'MAND_04M': _VbmappTimedMandStrategy(
    itemCode: 'MAND_04M',
    plannedMinutes: 60,
    countMetricLabel: '有效',
    inputLabel: '孩子要求内容',
    inputHint: '如：泡泡、出去、打开',
    targetOptions: <String>['物品', '动作', '活动'],
    requirePresentedEnvironment: true,
    excludePromptedEvents: true,
    defaultObservationHint: '记录自然情境下的自发要求，系统自动统计不同要求数量。',
  ),
  'MAND_08M': _VbmappTimedMandStrategy(
    itemCode: 'MAND_08M',
    plannedMinutes: 60,
    countMetricLabel: '不同',
    inputLabel: '孩子要求原话',
    inputHint: '如：跑快点、该我了、倒果汁',
    targetOptions: <String>['物品', '动作', '活动'],
    showPromptSelector: true,
    multiWordQualifiedMinCount: 2,
    defaultObservationHint: '记录自然情境下的不同要求；双词结构由系统根据原话自动判定。',
  ),
  'MAND_09M': _VbmappTimedMandStrategy(
    itemCode: 'MAND_09M',
    plannedMinutes: 30,
    countMetricLabel: '自发',
    inputLabel: '孩子要求内容',
    inputHint: '如：泡泡、出去、打开',
    targetOptions: <String>['物品', '动作', '活动'],
    excludePromptedEvents: true,
    displayMinSlots: 6,
    defaultObservationHint: '记录30分钟内的自发不同要求，系统自动去重统计。',
  ),
};

_VbmappTimedMandStrategy? _timedMandStrategyForItem(_VbmappItem item) {
  return _vbmappTimedMandStrategies[item.itemCode];
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
  return _vbmappTimedMandStrategies.containsKey(itemCode);
}

int _plannedMinutesForMandItem(_VbmappItem item) {
  return _timedMandStrategyForItem(item)?.plannedMinutes ?? 60;
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
  final _VbmappTimedMandStrategy? timedStrategy =
      _timedMandStrategyForItem(item);
  if (timedStrategy != null) {
    return timedStrategy.countsEvent(
      item,
      event,
      observation: observation,
    );
  }
  if (!event.isQualified) {
    return false;
  }
  switch (item.itemCode) {
    case 'MAND_05M':
      return event.environment.trim() == '呈现物品' &&
          _mandInitiationText(event) != '提问下';
    default:
      return event.isQualified;
  }
}

String _defaultMandUniqueKey(_VbmappMandEvent event) {
  final String text = event.target.trim().isNotEmpty
      ? event.target.trim()
      : event.utterance.trim();
  return text.toLowerCase();
}

String _normalizeMandDistinctKey(_VbmappMandEvent event) {
  String text = _normalizeMandPhraseText(_mandRequestText(event)).toLowerCase();
  if (text.isEmpty) {
    return '';
  }

  text = text
      .replaceFirst(RegExp(r'^(我想要|我要|我还要|我再要|给我|帮我|让我|替我)\s*'), '')
      .replaceFirst(RegExp(r'^(我们一起|一起)\s*'), '')
      .replaceFirst(RegExp(r'^(请|麻烦|想要)\s*'), '')
      .replaceFirst(RegExp(r'^(我要个|我要吃|我要喝)\s*'), '')
      .replaceFirst(RegExp(r'^(我想吃|我想喝)\s*'), '')
      .trim();

  text = text.replaceAll(RegExp(r'(吧|呀|啊|呢|啦|嘛|哦|喔)$'), '').trim();

  if (text.isEmpty) {
    text = _normalizeMandPhraseText(_mandRequestText(event)).toLowerCase();
  }

  return text;
}

int _qualifiedMandCount(List<_VbmappMandEvent> events) {
  final Set<String> uniqueTargets = <String>{};
  for (final _VbmappMandEvent event in events) {
    final String uniqueKey = _defaultMandUniqueKey(event);
    if (event.isQualified && uniqueKey.isNotEmpty) {
      uniqueTargets.add(uniqueKey);
    }
  }
  return uniqueTargets.length;
}

int _qualifiedMandCountForItem(
  _VbmappItem item,
  List<_VbmappMandEvent> events, {
  _VbmappObservationTimerState? observation,
}) {
  final _VbmappTimedMandStrategy? timedStrategy =
      _timedMandStrategyForItem(item);
  if (timedStrategy != null) {
    return timedStrategy.qualifiedCount(
      item,
      events,
      observation: observation,
    );
  }
  final Set<String> uniqueTargets = <String>{};
  for (final _VbmappMandEvent event in events) {
    final String uniqueKey = _defaultMandUniqueKey(event);
    if (_mandEventCountsForItem(item, event, observation: observation) &&
        uniqueKey.isNotEmpty) {
      uniqueTargets.add(uniqueKey);
    }
  }
  return uniqueTargets.length;
}

int _mandPhraseQualifiedCountForItem(
  _VbmappItem item,
  List<_VbmappMandEvent> events, {
  _VbmappObservationTimerState? observation,
}) {
  final _VbmappTimedMandStrategy? timedStrategy =
      _timedMandStrategyForItem(item);
  if (timedStrategy != null) {
    return timedStrategy.qualifiedMultiWordCount(
      item,
      events,
      observation: observation,
    );
  }
  final Set<String> uniqueTargets = <String>{};
  for (final _VbmappMandEvent event in events) {
    if (!_mandEventCountsForItem(item, event, observation: observation)) {
      continue;
    }
    if (!_assessMandPhrase(event).isMultiWord) {
      continue;
    }
    final String uniqueKey = _defaultMandUniqueKey(event);
    if (uniqueKey.isNotEmpty) {
      uniqueTargets.add(uniqueKey);
    }
  }
  return uniqueTargets.length;
}

bool _isLikelyMultiWordMand(_VbmappMandEvent event) {
  return _assessMandPhrase(event).isMultiWord;
}

_VbmappMandPhraseAssessment _assessMandPhrase(_VbmappMandEvent event) {
  final String explicitLevel = event.phraseLevel.trim();
  if (explicitLevel == '双词+') {
    return const _VbmappMandPhraseAssessment(
      label: '双词+',
      isMultiWord: true,
      normalizedText: '',
      reason: 'explicit_multi_word',
    );
  }
  if (explicitLevel == '单词') {
    return const _VbmappMandPhraseAssessment(
      label: '单词',
      isMultiWord: false,
      normalizedText: '',
      reason: 'explicit_single_word',
    );
  }

  final String rawText = _mandRequestText(event);
  String text = _normalizeMandPhraseText(rawText);
  if (text.isEmpty) {
    return const _VbmappMandPhraseAssessment(
      label: '未判定',
      isMultiWord: false,
      normalizedText: '',
      reason: 'empty_text',
    );
  }

  final List<String> spacedTokens = text
      .split(' ')
      .map((String part) => part.trim())
      .where((String part) => part.isNotEmpty)
      .toList(growable: false);
  if (spacedTokens.length >= 2) {
    return _VbmappMandPhraseAssessment(
      label: '双词+',
      isMultiWord: true,
      normalizedText: text,
      reason: 'space_token_count',
    );
  }

  final String withoutPrefix = text.replaceFirst(RegExp(r'^我想要'), '').trim();
  if (withoutPrefix.isEmpty) {
    return _VbmappMandPhraseAssessment(
      label: '单词',
      isMultiWord: false,
      normalizedText: text,
      reason: 'prefix_only',
    );
  }

  if (RegExp(r'(我|你|他|她|它|我们|我要|给我|帮我)').hasMatch(withoutPrefix) &&
      withoutPrefix.runes.length >= 3) {
    return _VbmappMandPhraseAssessment(
      label: '双词+',
      isMultiWord: true,
      normalizedText: withoutPrefix,
      reason: 'pronoun_plus_content',
    );
  }
  if (RegExp(
    r'(快点|一下|一会|给我|帮我|让我|一起|该我了|不要|好了|开门)$',
  ).hasMatch(withoutPrefix)) {
    return _VbmappMandPhraseAssessment(
      label: '双词+',
      isMultiWord: true,
      normalizedText: withoutPrefix,
      reason: 'suffix_phrase_pattern',
    );
  }
  if (withoutPrefix.runes.length >= 3 &&
      RegExp(r'^(打开|帮我|给我|让我|带我|一起|还要|再来|我要|我们|该我|跑|倒|推|拿|去|来|开)')
          .hasMatch(withoutPrefix)) {
    return _VbmappMandPhraseAssessment(
      label: '双词+',
      isMultiWord: true,
      normalizedText: withoutPrefix,
      reason: 'verb_phrase_pattern',
    );
  }

  return _VbmappMandPhraseAssessment(
    label: '单词',
    isMultiWord: false,
    normalizedText: withoutPrefix,
    reason: 'default_single_word',
  );
}

String _normalizeMandPhraseText(String rawText) {
  return rawText
      .replaceAll(RegExp(r'[，。！？、,.!?;；:/\\]+'), ' ')
      .replaceAll(RegExp(r'\s+'), ' ')
      .trim();
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
  if (item != null) {
    final _VbmappTimedMandStrategy? timedStrategy =
        _timedMandStrategyForItem(item);
    if (timedStrategy != null) {
      return timedStrategy.recordMetaText(event, item: item);
    }
  }
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
  final _VbmappTimedMandStrategy? strategy = _timedMandStrategyForItem(item);
  if (strategy == null) {
    return '系统按有效要求数量建议${_formatScore(suggestedScore)}分，老师可在下方评分区覆盖。';
  }
  return strategy.scoreBasisText(
    item,
    suggestedScore,
    qualifiedCount,
    actualObservationSeconds,
    multiWordCount,
  );
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
  final _VbmappTimedMandStrategy? timedStrategy =
      _timedMandStrategyForItem(item);
  if (timedStrategy != null) {
    return timedStrategy.suggestScore(
      item,
      events,
      observation: observation,
    );
  }
  final int count = _qualifiedMandCountForItem(
    item,
    events,
    observation: observation,
  );
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
