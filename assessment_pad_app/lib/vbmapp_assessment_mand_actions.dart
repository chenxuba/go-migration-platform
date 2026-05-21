part of 'vbmapp_assessment_page.dart';

extension _VbmappAssessmentMandActions on _VbmappAssessmentPageState {
  Future<void> _openMandEventDialog(_VbmappItem item) async {
    final _VbmappMandEvent? event = await showDialog<_VbmappMandEvent>(
      context: context,
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _VbmappMandEventDialog(
            generalizationMode: item.itemCode == 'MAND_03M',
          ),
        );
      },
    );
    if (event == null) {
      return;
    }
    await _addMandEvent(item, event);
  }

  Future<void> _addMandEvent(_VbmappItem item, _VbmappMandEvent event) async {
    final DateTime now = DateTime.now();
    final VbmappItemResponseSchema? schema = _schemaFor(item);
    final List<_VbmappMandEvent> events =
        List<_VbmappMandEvent>.from(_mandStoredEventsFor(item))
          ..add(
            event.recordedAtIso.trim().isEmpty
                ? event.copyWith(
                    recordedAtIso: now.toIso8601String(),
                    sourceItemCode: item.itemCode,
                  )
                : event,
          );
    final _VbmappObservationTimerState? observation = _mandObservationFor(item);
    final double suggestedScore = _suggestMandScore(
      item,
      events,
      observation: observation,
      responseSchema: schema,
    );
    setState(() {
      _mandEventsByItem[_mandStorageKeyFor(item.itemCode)] = events;
      _milestoneScores[item.itemCode] = suggestedScore;
      if (_usesSharedTimedMandObservation(item, schema: schema)) {
        _syncSharedTimedMandScores();
      }
      _rebuildScoreDerivedState();
      _autoSaveText = '保存中...';
    });
    await _saveMandEvidence(
      item,
      events,
      suggestedScore,
      observation: observation,
      responseSchema: schema,
    );
  }

  Future<void> _deleteMandEvent(_VbmappItem item, int index) async {
    final List<_VbmappMandEvent> events =
        List<_VbmappMandEvent>.from(_mandStoredEventsFor(item));
    if (index < 0 || index >= events.length) {
      return;
    }
    final VbmappItemResponseSchema? schema = _schemaFor(item);
    events.removeAt(index);
    final _VbmappObservationTimerState? observation = _mandObservationFor(item);
    final double suggestedScore = _suggestMandScore(
      item,
      events,
      observation: observation,
      responseSchema: schema,
    );
    setState(() {
      if (events.isEmpty) {
        _mandEventsByItem.remove(_mandStorageKeyFor(item.itemCode));
      } else {
        _mandEventsByItem[_mandStorageKeyFor(item.itemCode)] = events;
      }
      _milestoneScores[item.itemCode] = suggestedScore;
      if (_usesSharedTimedMandObservation(item, schema: schema)) {
        _syncSharedTimedMandScores();
      }
      _rebuildScoreDerivedState();
      _autoSaveText = '保存中...';
    });
    await _saveMandEvidence(
      item,
      events,
      suggestedScore,
      observation: observation,
      responseSchema: schema,
    );
  }

  Future<void> _updateMandObservation(
    _VbmappItem item,
    _VbmappObservationTimerState observation,
  ) async {
    final List<_VbmappMandEvent> events = _mandStoredEventsFor(item);
    final VbmappItemResponseSchema? schema = _schemaFor(item);
    final double suggestedScore = _suggestMandScore(
      item,
      events,
      observation: observation,
      responseSchema: schema,
    );
    setState(() {
      _mandObservationByItem[_mandStorageKeyFor(item.itemCode)] = observation;
      _milestoneScores[item.itemCode] = suggestedScore;
      if (_usesSharedTimedMandObservation(item, schema: schema)) {
        _syncSharedTimedMandScores();
      }
      _rebuildScoreDerivedState();
      _autoSaveText = '保存中...';
    });
    await _saveMandEvidence(
      item,
      events,
      suggestedScore,
      observation: observation,
      responseSchema: schema,
    );
  }

  Future<void> _saveMandEvidence(
    _VbmappItem item,
    List<_VbmappMandEvent> events,
    double suggestedScore, {
    _VbmappObservationTimerState? observation,
    VbmappItemResponseSchema? responseSchema,
  }) async {
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存证据', tone: PadMessageTone.error);
      return;
    }
    final int draftId = await _saveDraft(silent: true);
    if (draftId <= 0) {
      return;
    }
    final int qualifiedCount = _qualifiedMandCountForItem(
      item,
      events,
      observation: observation,
      responseSchema: responseSchema,
    );
    final _VbmappObservationTimerState? timerState = observation;
    final int actualObservationSeconds =
        timerState?.elapsedSecondsAt(DateTime.now()) ?? 0;
    final int effectiveObservationSeconds = _effectiveObservationSecondsForItem(
      item,
      timerState,
    );
    final _VbmappTimedMandStrategy? timedStrategy =
        _timedMandStrategyForItem(item);
    final int multiWordCount =
        (timedStrategy?.multiWordQualifiedMinCount ?? 0) > 0
            ? _mandPhraseQualifiedCountForItem(
                item,
                events,
                observation: observation,
                responseSchema: responseSchema,
              )
            : 0;
    try {
      final Map<String, dynamic> evidence = <String, dynamic>{
        'mandEvents': events
            .map((_VbmappMandEvent event) => event.toJson())
            .toList(growable: false),
      };
      if (timedStrategy != null) {
        evidence.addAll(
          timedStrategy.buildEvidenceFields(
            item,
            events,
            suggestedScore,
            observation: observation,
            responseSchema: responseSchema,
          ),
        );
      } else {
        evidence.addAll(<String, dynamic>{
          'qualifiedCount': qualifiedCount,
          'uniqueTargetCount': qualifiedCount,
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
          if ((timedStrategy?.multiWordQualifiedMinCount ?? 0) > 0)
            'multiWordQualifiedCount': multiWordCount,
          'scoreBasis':
              '系统按有效要求数量建议${_formatScore(suggestedScore)}分，老师可在下方评分区覆盖。',
        });
      }
      if (item.itemCode == 'MAND_03M') {
        evidence['generalizationCounts'] = _mandGeneralizationCounts(events);
        evidence['scoreBasis'] =
            '系统按互动对象、环境、不同例子的泛化记录建议${_formatScore(suggestedScore)}分，老师可在下方评分区覆盖。';
      }
      final VbmappDraftDetail detail = await widget.client.saveDraftItem(
        _token,
        <String, dynamic>{
          'draftId': draftId,
          'moduleCode': item.moduleCode,
          'itemCode': item.itemCode,
          'score': suggestedScore,
          'suggestedScore': suggestedScore,
          'teacherConfirmed': false,
          'recordStatus': 'auto_suggested',
          'evidence': evidence,
        },
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _draftId = detail.id > 0 ? detail.id : _draftId;
        _autoSaveText = '已保存 ${_formatClock(DateTime.now())}';
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() => _autoSaveText = '保存失败');
      _showMessage('VB-MAPP单题证据保存失败：$error', tone: PadMessageTone.error);
    }
  }

  void _restoreMandEvents(
    Map<String, Map<String, Map<String, dynamic>>> itemResponses,
  ) {
    _mandEventsByItem.clear();
    final Map<String, Map<String, dynamic>> milestoneResponses =
        itemResponses['milestones'] ?? const <String, Map<String, dynamic>>{};
    milestoneResponses.forEach((String itemCode, Map<String, dynamic> value) {
      final Object? evidenceRaw = value['evidence'];
      if (evidenceRaw is! Map) {
        return;
      }
      final Object? eventsRaw = evidenceRaw['mandEvents'];
      if (eventsRaw is! List) {
        return;
      }
      final List<_VbmappMandEvent> events = eventsRaw
          .map((Object? raw) => _VbmappMandEvent.fromJson(_dynamicMap(raw)))
          .where((_VbmappMandEvent event) => event.isNotEmpty)
          .toList(growable: false);
      if (events.isNotEmpty) {
        _mandEventsByItem[_mandStorageKeyFor(itemCode)] = events;
      }
    });
  }

  void _restoreMandObservations(
    Map<String, Map<String, Map<String, dynamic>>> itemResponses,
  ) {
    _mandObservationByItem.clear();
    final Map<String, Map<String, dynamic>> milestoneResponses =
        itemResponses['milestones'] ?? const <String, Map<String, dynamic>>{};
    milestoneResponses.forEach((String itemCode, Map<String, dynamic> value) {
      final Object? evidenceRaw = value['evidence'];
      if (evidenceRaw is! Map) {
        return;
      }
      final Object? timerRaw = evidenceRaw['timer'];
      if (timerRaw is! Map) {
        return;
      }
      final _VbmappObservationTimerState observation =
          _VbmappObservationTimerState.fromJson(_dynamicMap(timerRaw));
      if (!observation.isEmpty) {
        _mandObservationByItem[_mandStorageKeyFor(itemCode)] = observation;
      }
    });
  }

  num? _scoreFor(_VbmappItem item) {
    switch (item.moduleCode) {
      case 'milestones':
        return _milestoneScores[item.itemCode];
      case 'barriers':
        return _barrierScores[item.itemCode];
      case 'transition':
        return _transitionScores[item.itemCode];
    }
    return null;
  }

  VbmappItemResponseSchema? _schemaFor(_VbmappItem item) {
    return _itemSchemas[_schemaKey(item.moduleCode, item.itemCode)];
  }

  VbmappItemResponseSchema? _schemaForMilestoneCode(String itemCode) {
    return _itemSchemas[_schemaKey('milestones', itemCode)];
  }

  String? _sharedTimedMandGroupIdFor(
    _VbmappItem item, {
    VbmappItemResponseSchema? schema,
  }) {
    final VbmappItemResponseSchema? resolvedSchema = schema ?? _schemaFor(item);
    final VbmappSharedObservationRule? rule =
        resolvedSchema?.smartRules.sharedObservation;
    if (rule != null && rule.enabled) {
      return rule.groupId.trim().isEmpty
          ? 'mand_timed_shared_v1'
          : rule.groupId.trim();
    }
    if (_vbmappSharedTimedMandItemCodes.contains(item.itemCode)) {
      return 'mand_timed_shared_v1';
    }
    return null;
  }

  String _sharedTimedMandStorageKeyForGroup(String groupId) {
    final String normalized = groupId.trim();
    if (normalized.isEmpty || normalized == 'mand_timed_shared_v1') {
      return _vbmappSharedTimedMandStorageKey;
    }
    return '$_vbmappSharedTimedMandStorageKey::$normalized';
  }

  String? _sharedTimedMandPrimaryItemCodeFor(
    _VbmappItem item, {
    VbmappItemResponseSchema? schema,
  }) {
    final String? groupId = _sharedTimedMandGroupIdFor(item, schema: schema);
    if (groupId == null) {
      return null;
    }
    final VbmappItemResponseSchema? resolvedSchema = schema ?? _schemaFor(item);
    final String primaryCode =
        resolvedSchema?.smartRules.sharedObservation?.primaryMilestoneId ?? '';
    return primaryCode.trim().isEmpty ? 'MAND_04M' : primaryCode.trim();
  }

  bool _usesSharedTimedMandObservation(
    _VbmappItem item, {
    VbmappItemResponseSchema? schema,
  }) {
    return _sharedTimedMandGroupIdFor(item, schema: schema) != null;
  }

  VbmappMaterialProfile? _materialProfileFor(
    _VbmappItem item,
    VbmappItemResponseSchema? schema,
  ) {
    final VbmappMaterialProfile? itemProfile =
        _itemMaterialProfiles[item.itemCode];
    if (itemProfile != null) {
      return itemProfile;
    }
    if (schema == null || schema.materialProfileId.isEmpty) {
      return null;
    }
    return _materialProfiles[schema.materialProfileId];
  }

  List<_VbmappMandEvent> _mandEventsFor(_VbmappItem item) {
    return _mandStoredEventsFor(item);
  }

  _VbmappObservationTimerState? _mandObservationFor(_VbmappItem item) {
    return _mandObservationByItem[_mandStorageKeyFor(item.itemCode)];
  }

  _VbmappItem? _activeMandObservationItem() {
    final Set<String> visitedSharedGroups = <String>{};
    for (final _VbmappItem item in _milestoneItems) {
      final String? groupId = _sharedTimedMandGroupIdFor(item);
      if (groupId == null || !visitedSharedGroups.add(groupId)) {
        continue;
      }
      final _VbmappObservationTimerState? sharedObservation =
          _mandObservationByItem[_sharedTimedMandStorageKeyForGroup(groupId)];
      if (sharedObservation != null &&
          sharedObservation.hasStarted &&
          !sharedObservation.ended) {
        final String primaryCode =
            _sharedTimedMandPrimaryItemCodeFor(item) ?? 'MAND_04M';
        return _milestoneItems.firstWhere(
          (_VbmappItem candidate) => candidate.itemCode == primaryCode,
          orElse: () => item,
        );
      }
    }
    for (final _VbmappItem item in _milestoneItems) {
      final _VbmappObservationTimerState? observation =
          _mandObservationByItem[item.itemCode];
      if (observation != null && observation.hasStarted && !observation.ended) {
        return item;
      }
    }
    return null;
  }

  int _activeMandObservationQualifiedCount(_VbmappItem item) {
    return _qualifiedMandCountForItem(
      item,
      _mandStoredEventsFor(item),
      observation: _mandObservationFor(item),
      responseSchema: _schemaFor(item),
    );
  }

  List<_VbmappMandEvent> _mandStoredEventsFor(_VbmappItem item) {
    return _mandEventsByItem[_mandStorageKeyFor(item.itemCode)] ??
        const <_VbmappMandEvent>[];
  }

  bool _hasSharedTimedMandEvidenceForGroup(String groupId) {
    final String storageKey = _sharedTimedMandStorageKeyForGroup(groupId);
    final List<_VbmappMandEvent> sharedEvents =
        _mandEventsByItem[storageKey] ?? const <_VbmappMandEvent>[];
    if (sharedEvents.isNotEmpty) {
      return true;
    }
    final _VbmappObservationTimerState? sharedObservation =
        _mandObservationByItem[storageKey];
    if (sharedObservation == null) {
      return false;
    }
    return sharedObservation.hasStarted ||
        sharedObservation.accumulatedSeconds > 0 ||
        sharedObservation.ended;
  }

  bool _hasSharedTimedMandEvidence() {
    final Set<String> groupIds = <String>{};
    for (final _VbmappItem item in _milestoneItems) {
      final String? groupId = _sharedTimedMandGroupIdFor(item);
      if (groupId != null) {
        groupIds.add(groupId);
      }
    }
    for (final String groupId in groupIds) {
      if (_hasSharedTimedMandEvidenceForGroup(groupId)) {
        return true;
      }
    }
    return false;
  }

  void _clearBuggedSharedTimedMandScores() {
    final Map<String, List<String>> groupCodes = <String, List<String>>{};
    for (final _VbmappItem item in _milestoneItems) {
      final String? groupId = _sharedTimedMandGroupIdFor(item);
      if (groupId == null) {
        continue;
      }
      groupCodes.putIfAbsent(groupId, () => <String>[]).add(item.itemCode);
    }
    if (groupCodes.isEmpty) {
      return;
    }
    for (final List<String> codes in groupCodes.values) {
      final bool allZero = codes.every(
        (String code) => (_milestoneScores[code] ?? -1) == 0,
      );
      if (!allZero) {
        continue;
      }
      for (final String code in codes) {
        _milestoneScores.remove(code);
      }
    }
  }

  String _mandStorageKeyFor(String itemCode) {
    final VbmappItemResponseSchema? schema = _schemaForMilestoneCode(itemCode);
    _VbmappItem? item;
    for (final _VbmappItem candidate in _milestoneItems) {
      if (candidate.itemCode == itemCode) {
        item = candidate;
        break;
      }
    }
    if (item != null) {
      final String? groupId = _sharedTimedMandGroupIdFor(item, schema: schema);
      if (groupId != null) {
        return _sharedTimedMandStorageKeyForGroup(groupId);
      }
    }
    return itemCode;
  }

  void _syncSharedTimedMandScores() {
    final Map<String, List<_VbmappItem>> groups = <String, List<_VbmappItem>>{};
    for (final _VbmappItem item in _milestoneItems) {
      final String? groupId = _sharedTimedMandGroupIdFor(item);
      if (groupId == null) {
        continue;
      }
      groups.putIfAbsent(groupId, () => <_VbmappItem>[]).add(item);
    }
    for (final MapEntry<String, List<_VbmappItem>> entry in groups.entries) {
      if (!_hasSharedTimedMandEvidenceForGroup(entry.key)) {
        continue;
      }
      for (final _VbmappItem item in entry.value) {
        final List<_VbmappMandEvent> events = _mandStoredEventsFor(item);
        final _VbmappObservationTimerState? observation =
            _mandObservationFor(item);
        _milestoneScores[item.itemCode] = _suggestMandScore(
          item,
          events,
          observation: observation,
          responseSchema: _schemaFor(item),
        );
      }
    }
  }

  Future<void> _openActiveMandQuickRecord() async {
    final _VbmappItem? item = _activeMandObservationItem();
    if (item == null) {
      return;
    }
    final VbmappMaterialProfile? profile =
        _materialProfileFor(item, _schemaFor(item));
    final _VbmappMandEvent? event = await showDialog<_VbmappMandEvent>(
      context: context,
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _VbmappMand4QuickRecordDialog(materialProfile: profile),
        );
      },
    );
    if (event == null) {
      return;
    }
    await _addMandEvent(item, event);
  }

  Future<void> _confirmFinishActiveMandObservation() async {
    final _VbmappItem? item = _activeMandObservationItem();
    if (item == null) {
      return;
    }
    final bool? confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return const PadDialogViewport(
          child: _VbmappObservationFinishConfirmDialog(),
        );
      },
    );
    if (confirmed != true) {
      return;
    }
    final _VbmappObservationTimerState? observation = _mandObservationFor(item);
    if (observation == null) {
      return;
    }
    await _updateMandObservation(item, observation.finish(DateTime.now()));
  }
}
