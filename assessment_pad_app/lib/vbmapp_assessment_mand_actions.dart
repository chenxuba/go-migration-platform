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
      _mandEventsByItem[_mandEventStorageKeyFor(item.itemCode)] = events;
      _milestoneScores[item.itemCode] = suggestedScore;
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
        _mandEventsByItem.remove(_mandEventStorageKeyFor(item.itemCode));
      } else {
        _mandEventsByItem[_mandEventStorageKeyFor(item.itemCode)] = events;
      }
      _milestoneScores[item.itemCode] = suggestedScore;
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
      _mandObservationByItem[_mandObservationStorageKeyFor(item.itemCode)] =
          observation;
      if (events.isNotEmpty || _milestoneScores.containsKey(item.itemCode)) {
        _milestoneScores[item.itemCode] = suggestedScore;
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
      responseSchema: responseSchema,
    );
    final _VbmappTimedMandStrategy? timedStrategy =
        _timedMandStrategyForItem(item, responseSchema: responseSchema);
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
        _mandEventsByItem[_mandEventStorageKeyFor(itemCode)] = events;
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
        _mandObservationByItem[_mandObservationStorageKeyFor(itemCode)] =
            observation;
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
          ? _fallbackSharedTimedMandGroupIdFor(item.itemCode)
          : rule.groupId.trim();
    }
    return _fallbackSharedTimedMandGroupIdFor(item.itemCode);
  }

  String _sharedTimedMandStorageKeyForGroup(String groupId) {
    final String normalized = groupId.trim();
    if (normalized.isEmpty || normalized == _vbmappMandLevel2TimedGroupId) {
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
    if (primaryCode.trim().isNotEmpty) {
      return primaryCode.trim();
    }
    if (groupId == _vbmappMandLevel3TimedGroupId) {
      return 'MAND_11M';
    }
    return 'MAND_04M';
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
    return _mandObservationByItem[_mandObservationStorageKeyFor(item.itemCode)];
  }

  bool _hasActiveMandObservation(_VbmappItem item) {
    final _VbmappObservationTimerState? observation = _mandObservationFor(item);
    return observation != null && observation.hasStarted && !observation.ended;
  }

  _VbmappItem? _activeMandObservationItem() {
    final List<_VbmappItem> activeItems = _activeMandObservationItems();
    return activeItems.isEmpty ? null : activeItems.first;
  }

  List<_VbmappItem> _activeMandObservationItems() {
    final List<_VbmappItem> activeItems = <_VbmappItem>[];
    final Set<String> visitedSharedGroups = <String>{};
    final Set<String> visitedStorageKeys = <String>{};
    for (final _VbmappItem item in _milestoneItems) {
      final String? groupId = _sharedTimedMandGroupIdFor(item);
      if (groupId == null || !visitedSharedGroups.add(groupId)) {
        continue;
      }
      final String storageKey = _sharedTimedMandStorageKeyForGroup(groupId);
      final _VbmappObservationTimerState? sharedObservation =
          _mandObservationByItem[storageKey];
      if (sharedObservation != null &&
          sharedObservation.hasStarted &&
          !sharedObservation.ended) {
        visitedStorageKeys.add(storageKey);
        final String primaryCode =
            _sharedTimedMandPrimaryItemCodeFor(item) ?? 'MAND_04M';
        activeItems.add(
          _milestoneItems.firstWhere(
            (_VbmappItem candidate) => candidate.itemCode == primaryCode,
            orElse: () => item,
          ),
        );
      }
    }
    for (final _VbmappItem item in _milestoneItems) {
      if (_sharedTimedMandGroupIdFor(item) != null) {
        continue;
      }
      final String storageKey = _mandObservationStorageKeyFor(item.itemCode);
      if (!visitedStorageKeys.add(storageKey)) {
        continue;
      }
      final _VbmappObservationTimerState? observation =
          _mandObservationByItem[storageKey];
      if (observation != null && observation.hasStarted && !observation.ended) {
        activeItems.add(item);
      }
    }
    return activeItems;
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
    return _mandEventsByItem[_mandEventStorageKeyFor(item.itemCode)] ??
        const <_VbmappMandEvent>[];
  }

  int _sharedTimedMandRecordCountForGroup(String groupId) {
    int total = 0;
    for (final _VbmappItem item in _sharedTimedMandItemsForGroup(groupId)) {
      total += _mandStoredEventsFor(item).length;
    }
    return total;
  }

  List<_VbmappActiveObservationSummary> _sharedTimedMandSummariesForGroup(
    String groupId,
  ) {
    return _sharedTimedMandItemsForGroup(groupId).map((_VbmappItem item) {
      final VbmappItemResponseSchema? schema = _schemaFor(item);
      final List<_VbmappMandEvent> events = _mandStoredEventsFor(item);
      final _VbmappObservationTimerState? observation =
          _mandObservationFor(item);
      final _VbmappTimedMandStrategy? strategy =
          _timedMandStrategyForItem(item, responseSchema: schema);
      final int qualified = _qualifiedMandCountForItem(
        item,
        events,
        observation: observation,
        responseSchema: schema,
      );
      final int target =
          strategy?.onePointCount(item) ?? _scoreCountThreshold(item, 1) ?? 1;
      final int multiWordTarget = strategy?.multiWordQualifiedMinCount ?? 0;
      final int phraseCount = multiWordTarget <= 0
          ? 0
          : _mandPhraseQualifiedCountForItem(
              item,
              events,
              observation: observation,
              responseSchema: schema,
            );
      final bool complete = qualified >= target &&
          (multiWordTarget <= 0 || phraseCount >= multiWordTarget);
      return _VbmappActiveObservationSummary(
        label: item.navCode,
        value: '$qualified/$target',
        plannedSeconds: _plannedMinutesForMandItem(
              item,
              responseSchema: schema,
            ) *
            Duration.secondsPerMinute,
        complete: complete,
      );
    }).toList(growable: false);
  }

  int _sharedTimedMandMaxPlannedMinutesForGroup(String groupId) {
    int maxMinutes = 0;
    for (final _VbmappItem item in _sharedTimedMandItemsForGroup(groupId)) {
      final int plannedMinutes = _plannedMinutesForMandItem(
        item,
        responseSchema: _schemaFor(item),
      );
      if (plannedMinutes > maxMinutes) {
        maxMinutes = plannedMinutes;
      }
    }
    return maxMinutes > 0 ? maxMinutes : 60;
  }

  String _sharedTimedMandStatusLabelForGroup(String groupId) {
    return '提要求观察中';
  }

  String _mandEventStorageKeyFor(String itemCode) {
    return itemCode;
  }

  String _mandObservationStorageKeyFor(String itemCode) {
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

  List<_VbmappItem> _sharedTimedMandItemsForGroup(String groupId) {
    final List<_VbmappItem> items = <_VbmappItem>[];
    for (final _VbmappItem item in _milestoneItems) {
      if (_sharedTimedMandGroupIdFor(item) == groupId) {
        items.add(item);
      }
    }
    if (items.isNotEmpty) {
      return items;
    }
    final Set<String> fallbackCodes = groupId == _vbmappMandLevel3TimedGroupId
        ? _vbmappMandLevel3TimedItemCodes
        : _vbmappMandLevel2TimedItemCodes;
    return _milestoneItems
        .where((_VbmappItem item) => fallbackCodes.contains(item.itemCode))
        .toList(growable: false);
  }

  String? _fallbackSharedTimedMandGroupIdFor(String itemCode) {
    if (_vbmappMandLevel2TimedItemCodes.contains(itemCode)) {
      return _vbmappMandLevel2TimedGroupId;
    }
    if (_vbmappMandLevel3TimedItemCodes.contains(itemCode)) {
      return _vbmappMandLevel3TimedGroupId;
    }
    return null;
  }

  Future<void> _openActiveMandQuickRecord([_VbmappItem? sourceItem]) async {
    final _VbmappItem? item = sourceItem ?? _activeMandObservationItem();
    if (item == null) {
      return;
    }
    final String? groupId = _sharedTimedMandGroupIdFor(item);
    final List<_VbmappItem> targetItems = groupId == null
        ? <_VbmappItem>[item]
        : _sharedTimedMandItemsForGroup(groupId);
    final List<_VbmappTimedMandQuickRecordTarget> targets =
        targetItems.map((_VbmappItem targetItem) {
      final VbmappItemResponseSchema? schema = _schemaFor(targetItem);
      final List<_VbmappMandEvent> events = _mandStoredEventsFor(targetItem);
      return _VbmappTimedMandQuickRecordTarget(
        item: targetItem,
        responseSchema: schema,
        materialProfile: _materialProfileFor(targetItem, schema),
        recordCount: events.length,
        qualifiedCount: _qualifiedMandCountForItem(
          targetItem,
          events,
          observation: _mandObservationFor(targetItem),
          responseSchema: schema,
        ),
      );
    }).toList(growable: false);
    final String initialItemCode = targets.any(
      (_VbmappTimedMandQuickRecordTarget target) =>
          target.item.itemCode == _selectedItem.itemCode,
    )
        ? _selectedItem.itemCode
        : item.itemCode;
    final _VbmappTimedMandQuickRecordResult? result =
        await showDialog<_VbmappTimedMandQuickRecordResult>(
      context: context,
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _VbmappMand4QuickRecordDialog(
            targets: targets,
            initialItemCode: initialItemCode,
          ),
        );
      },
    );
    if (result == null) {
      return;
    }
    await _addMandEvent(result.item, result.event);
  }

  Future<void> _confirmFinishActiveMandObservation([
    _VbmappItem? sourceItem,
  ]) async {
    final _VbmappItem? item = sourceItem ?? _activeMandObservationItem();
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
