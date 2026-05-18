part of 'pep3_assessment_page.dart';

extension _Pep3AssessmentStateActions on _Pep3AssessmentPageState {
  Future<String> _readToken() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    return prefs.getString(_Pep3AssessmentPageState._authTokenStorageKey) ?? '';
  }

  Future<void> _initialize() async {
    setState(() {
      _loading = true;
      _errorMessage = '';
    });
    final String token = await _readToken();
    if (token.trim().isEmpty) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = '请先登录后再进行测评';
      });
      return;
    }
    try {
      final List<Object> result = await Future.wait<Object>(<Future<Object>>[
        widget.client.fetchTemplateSummary(token),
        widget.homeClient.fetchCurrentSession(token),
      ]);
      if (!mounted) {
        return;
      }
      _template = result[0] as Pep3TemplateSummary;
      _session = result[1] as HomeSession;
      if (_examinerName.trim().isEmpty) {
        _examinerName = _session.nickName.trim().isNotEmpty
            ? _session.nickName.trim()
            : _session.username.trim();
      }
      if (widget.args.draftId > 0) {
        await _loadDraftDetail(token, widget.args.draftId);
      } else {
        _detectedDraft = await _findLatestDraft(token);
        _prefetchDetectedDraftDetail(token, _detectedDraft);
      }
      _currentItemNo = _resolveInitialItemNo();
      _expandedGroupKey = _groupKeyForItem(_currentItemNo);
      await _loadCurrentItem(token);
      await _loadPreviousAssessment(token);
      if (_draft?.id != null && _draft!.id > 0) {
        _refreshCaregiverInvite(silent: true);
      }
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
      });
      _keepActiveItemVisible();
      _showDetectedDraftDialogIfNeeded();
      if (_detectedDraft == null && (_draft?.id ?? 0) <= 0) {
        await _ensureDraftAndInvite();
      }
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = 'PEP-3测评页面加载失败：$error';
      });
    }
  }

  Future<Pep3DraftSummary?> _findLatestDraft(String token) async {
    final int studentId = widget.args.studentId;
    if (studentId <= 0) {
      return null;
    }
    final Pep3DraftPage page = await widget.client.fetchDraftsPage(
      token,
      studentId: studentId,
      pageSize: 1,
      latestOnly: true,
    );
    if (page.items.isEmpty || page.items.first.id <= 0) {
      return null;
    }
    return page.items.first;
  }

  Future<void> _loadDraftDetail(String token, int draftId) async {
    final Pep3DraftDetail detail = await _fetchDraftDetail(token, draftId);
    _applyDraftDetail(detail);
  }

  Future<Pep3DraftDetail> _fetchDraftDetail(String token, int draftId) {
    return widget.client.fetchDraftDetail(token, draftId);
  }

  void _applyDraftDetail(Pep3DraftDetail detail) {
    if (_caregiverInviteDraftId != detail.id) {
      _caregiverInvite = null;
      _caregiverInviteDraftId = 0;
      _caregiverLoading = detail.id > 0;
    }
    _draft = detail;
    _studentName = detail.studentName.trim().isNotEmpty
        ? detail.studentName.trim()
        : _studentName;
    _birthDate = _normalizeDate(detail.birthDate) ?? _birthDate;
    _assessmentDate = _normalizeDate(detail.assessmentDate) ?? _assessmentDate;
    _examinerName = detail.examinerName.trim().isNotEmpty
        ? detail.examinerName.trim()
        : _examinerName;
    _applyDraftInput(detail.input);
  }

  void _prefetchDetectedDraftDetail(String token, Pep3DraftSummary? draft) {
    if (draft == null || draft.id <= 0) {
      _detectedDraftDetailDraftId = 0;
      _detectedDraftDetailRequest = null;
      return;
    }
    if (_detectedDraftDetailDraftId == draft.id &&
        _detectedDraftDetailRequest != null) {
      return;
    }
    final Future<Pep3DraftDetail> request = _fetchDraftDetail(token, draft.id);
    _detectedDraftDetailDraftId = draft.id;
    _detectedDraftDetailRequest = request;
    unawaited(
      request.then<void>(
        (Pep3DraftDetail _) {},
        onError: (Object _, StackTrace __) {},
      ),
    );
  }

  Future<Pep3DraftDetail> _resolveDetectedDraftDetail(
    String token,
    Pep3DraftSummary draft,
  ) async {
    final Future<Pep3DraftDetail>? prefetched = _detectedDraftDetailRequest;
    if (_detectedDraftDetailDraftId == draft.id && prefetched != null) {
      try {
        return await prefetched;
      } on Object {
        // Retry below if the prefetch failed.
      }
    }
    return _fetchDraftDetail(token, draft.id);
  }

  void _showDetectedDraftDialogIfNeeded() {
    final Pep3DraftSummary? draft = _detectedDraft;
    if (!mounted || _draftDialogShown || draft == null || draft.id <= 0) {
      return;
    }
    final int answered = draft.progress.answeredItemCount > 0
        ? draft.progress.answeredItemCount
        : draft.answeredItemCount;
    final int total = draft.progress.itemCount > 0
        ? draft.progress.itemCount
        : math.max(_totalCount, answered);
    _draftDialogShown = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) {
        return;
      }
      showDialog<void>(
        context: context,
        barrierDismissible: false,
        barrierColor: Colors.black.withOpacity(.32),
        builder: (BuildContext dialogContext) {
          return PopScope(
            canPop: false,
            child: PadDialogViewport(
              child: AssessmentDraftResumeDialog(
                message: '当前儿童存在一份未提交的 PEP-3 测评草稿。',
                metaRows: <AssessmentDraftResumeMetaRow>[
                  AssessmentDraftResumeMetaRow(
                    label: '已完成',
                    value: '$answered / $total 题',
                  ),
                  AssessmentDraftResumeMetaRow(
                    label: '更新时间',
                    value: _formatDateTime(draft.updatedTime),
                  ),
                ],
                accentColor: _Pep3Colors.orange,
                inkColor: _Pep3Colors.ink,
                bodyColor: _Pep3Colors.text,
                lineColor: _Pep3Colors.line,
                lineSoftColor: _Pep3Colors.lineSoft,
                closeDuration: const Duration(milliseconds: 260),
                onRestart: _restartWithoutDetectedDraft,
                onContinue: () => _continueDetectedDraft(draft),
              ),
            ),
          );
        },
      );
    });
  }

  Future<void> _restartWithoutDetectedDraft() async {
    if (!mounted) {
      return;
    }
    setState(() {
      _detectedDraft = null;
      _draft = null;
      _itemScores.clear();
      _recordValues.clear();
      _savedItems.clear();
      _caregiverInvite = null;
      _caregiverInviteDraftId = 0;
      _detectedDraftDetailDraftId = 0;
      _detectedDraftDetailRequest = null;
      _autoSaveText = '已开始新的测评';
      _assessmentDate =
          _normalizeDate(widget.args.assessmentDate) ?? _todayIsoDate();
      _examinerName = widget.args.examinerName.trim().isNotEmpty
          ? widget.args.examinerName.trim()
          : _examinerName;
      _currentItemNo =
          _template.allItems.isEmpty ? 0 : _template.allItems.first.itemNo;
      _expandedGroupKey = _groupKeyForItem(_currentItemNo);
    });
    await _loadCurrentItem();
    await _ensureDraftAndInvite();
  }

  Future<void> _ensureDraftAndInvite() async {
    if ((_draft?.id ?? 0) > 0) {
      await _refreshCaregiverInvite(silent: true);
      return;
    }
    final Pep3DraftDetail? detail = await _saveDraft(silent: true);
    if ((detail?.id ?? 0) > 0) {
      await _refreshCaregiverInvite(silent: true);
    }
  }

  Future<bool> _continueDetectedDraft(Pep3DraftSummary draft) async {
    if (draft.id <= 0) {
      return false;
    }
    setState(() {
      _autoSaveText = '';
    });
    try {
      final String token = await _readToken();
      if (token.trim().isEmpty) {
        _showMessage('请先登录后再恢复草稿');
        return false;
      }
      final Pep3DraftDetail detail =
          await _resolveDetectedDraftDetail(token, draft);
      _applyDraftDetail(detail);
      if (!mounted) {
        return false;
      }
      final int restoredItemNo = _resolveInitialItemNo();
      final bool shouldLoadItem = !_itemCache.containsKey(restoredItemNo);
      setState(() {
        _currentItemNo = restoredItemNo;
        _expandedGroupKey = _groupKeyForItem(restoredItemNo);
        _itemLoading = shouldLoadItem;
        _detectedDraft = null;
        _autoSaveText = '已恢复最新草稿';
      });
      _detectedDraftDetailDraftId = 0;
      _detectedDraftDetailRequest = null;
      if (_questionScrollController.hasClients) {
        _questionScrollController.jumpTo(0);
      }
      _keepActiveItemVisible();
      unawaited(_loadCurrentItem(token));
      if (_currentPreviousAssessmentLookupKey != _previousAssessmentLookupKey) {
        unawaited(_loadPreviousAssessment(token));
      }
      unawaited(_refreshCaregiverInvite(silent: true));
      return true;
    } on Object catch (error) {
      if (!mounted) {
        return false;
      }
      _showMessage('恢复草稿失败：$error');
      return false;
    }
  }

  Future<void> _loadPreviousAssessment(String token) async {
    final String lookupKey = _currentPreviousAssessmentLookupKey;
    final int studentId = _studentId;
    String nextPreviousAssessmentDate = '';
    final Map<int, int> nextPreviousItemScores = <int, int>{};
    if (studentId <= 0) {
      _applyPreviousAssessmentState(
        lookupKey: '',
        previousAssessmentDate: '',
        previousItemScores: nextPreviousItemScores,
      );
      return;
    }
    try {
      final Pep3RecordPage page = await widget.client.fetchRecordsPage(
        token,
        studentId: studentId,
        assessmentDateEnd: _assessmentDate,
        pageSize: 5,
      );
      final Pep3RecordSummary? latest = page.items
          .where((Pep3RecordSummary item) => item.id > 0)
          .cast<Pep3RecordSummary?>()
          .firstOrNull;
      if (latest == null) {
        _applyPreviousAssessmentState(
          lookupKey: lookupKey,
          previousAssessmentDate: '',
          previousItemScores: nextPreviousItemScores,
        );
        return;
      }
      final Pep3RecordDetail detail =
          await widget.client.fetchRecordDetail(token, latest.id);
      nextPreviousAssessmentDate = _normalizeDate(detail.assessmentDate) ??
          _normalizeDate(latest.assessmentDate) ??
          '';
      nextPreviousItemScores.addAll(detail.input.itemScores);
    } on Object {
      nextPreviousAssessmentDate = '';
      nextPreviousItemScores.clear();
    }
    _applyPreviousAssessmentState(
      lookupKey: lookupKey,
      previousAssessmentDate: nextPreviousAssessmentDate,
      previousItemScores: nextPreviousItemScores,
    );
  }

  void _applyDraftInput(Pep3DraftInput input) {
    if (input.studentName.trim().isNotEmpty) {
      _studentName = input.studentName.trim();
    }
    if (input.birthDate.trim().isNotEmpty) {
      _birthDate = _normalizeDate(input.birthDate) ?? _birthDate;
    }
    if (input.assessmentDate.trim().isNotEmpty) {
      _assessmentDate = _normalizeDate(input.assessmentDate) ?? _assessmentDate;
    }
    if (input.examinerName.trim().isNotEmpty) {
      _examinerName = input.examinerName.trim();
    }
    _itemScores
      ..clear()
      ..addAll(input.itemScores);
    _recordValues
      ..clear()
      ..addAll(input.itemRecordValues);
  }

  int _resolveInitialItemNo() {
    final List<Pep3ItemSummary> items = _template.allItems;
    if (items.isEmpty) {
      return 0;
    }
    final List<int> missing = _draft?.progress.missingItemNos ??
        _detectedDraft?.progress.missingItemNos ??
        <int>[];
    if (missing.isNotEmpty &&
        items.any((Pep3ItemSummary item) => item.itemNo == missing.first)) {
      return missing.first;
    }
    return items.first.itemNo;
  }

  String get _currentPreviousAssessmentLookupKey {
    final int studentId = _studentId;
    final String assessmentDate = _assessmentDate.trim();
    if (studentId <= 0 || assessmentDate.isEmpty) {
      return '';
    }
    return '$studentId|$assessmentDate';
  }

  void _applyPreviousAssessmentState({
    required String lookupKey,
    required String previousAssessmentDate,
    required Map<int, int> previousItemScores,
  }) {
    if (!mounted) {
      _previousAssessmentLookupKey = lookupKey;
      _previousAssessmentDate = previousAssessmentDate;
      _previousItemScores
        ..clear()
        ..addAll(previousItemScores);
      return;
    }
    setState(() {
      _previousAssessmentLookupKey = lookupKey;
      _previousAssessmentDate = previousAssessmentDate;
      _previousItemScores
        ..clear()
        ..addAll(previousItemScores);
    });
  }

  Future<void> _loadCurrentItem([String? token]) async {
    final int itemNo = _currentItemNo;
    if (itemNo <= 0) {
      return;
    }
    if (_itemCache.containsKey(itemNo)) {
      if (mounted && _itemLoading) {
        setState(() => _itemLoading = false);
      }
      _prefetchNextItem(token);
      return;
    }
    if (mounted && !_itemLoading) {
      setState(() => _itemLoading = true);
    }
    try {
      await _fetchAndCacheItem(itemNo, token);
      _prefetchNextItem(token);
    } on Object catch (error) {
      _showMessage('第$itemNo题加载失败：$error');
    } finally {
      if (mounted && _currentItemNo == itemNo) {
        setState(() => _itemLoading = false);
      }
    }
  }

  Future<Pep3AssessmentItem> _fetchAndCacheItem(
    int itemNo, [
    String? token,
  ]) {
    final Pep3AssessmentItem? cached = _itemCache[itemNo];
    if (cached != null) {
      return Future<Pep3AssessmentItem>.value(cached);
    }
    final Future<Pep3AssessmentItem>? existing = _itemFetches[itemNo];
    if (existing != null) {
      return existing;
    }
    final Future<Pep3AssessmentItem> future = (() async {
      try {
        final String resolvedToken = token ?? await _readToken();
        final Pep3AssessmentItem item =
            await widget.client.fetchTemplateItem(resolvedToken, itemNo);
        if (mounted) {
          setState(() => _itemCache[itemNo] = item);
        } else {
          _itemCache[itemNo] = item;
        }
        return item;
      } finally {
        _itemFetches.remove(itemNo);
      }
    })();
    _itemFetches[itemNo] = future;
    return future;
  }

  void _prefetchNextItem([String? token]) {
    final int nextIndex = _currentIndex + 1;
    if (nextIndex < 0 || nextIndex >= _template.allItems.length) {
      return;
    }
    final int nextItemNo = _template.allItems[nextIndex].itemNo;
    if (nextItemNo <= 0 || _itemCache.containsKey(nextItemNo)) {
      return;
    }
    _fetchAndCacheItem(nextItemNo, token).catchError((Object _) {
      return Pep3AssessmentItem.empty;
    });
  }

  Future<Pep3DraftDetail?> _saveDraft({bool silent = false}) async {
    final String token = await _readToken();
    if (token.trim().isEmpty) {
      _showMessage('请先登录后再保存草稿');
      return null;
    }
    if (_studentId <= 0 || _studentName.trim().isEmpty) {
      _showMessage('缺少学员信息，无法保存草稿');
      return null;
    }
    setState(() => _savingDraft = true);
    try {
      final Pep3DraftDetail detail =
          await widget.client.saveDraft(token, _buildDraftPayload());
      if (!mounted) {
        return detail;
      }
      setState(() {
        _draft = detail;
        _autoSaveText = '草稿已保存';
      });
      if (!silent) {
        _showMessage('草稿已保存', tone: PadMessageTone.success);
      }
      return detail;
    } on Object catch (error) {
      _showMessage('保存草稿失败：$error');
      return null;
    } finally {
      if (mounted) {
        setState(() => _savingDraft = false);
      }
    }
  }

  Future<void> _saveCurrentItem() => _saveItem(_currentItemNo);

  Future<void> _saveItem(int itemNo) async {
    if (itemNo <= 0) {
      return;
    }
    final String token = await _readToken();
    if (token.trim().isEmpty) {
      _showMessage('请先登录后再保存题目');
      return;
    }
    int draftId = _draft?.id ?? 0;
    if (draftId <= 0) {
      final Pep3DraftDetail? created = await _saveDraft(silent: true);
      draftId = created?.id ?? 0;
    }
    if (draftId <= 0) {
      return;
    }
    setState(() {
      _savingItems.add(itemNo);
      _savedItems.remove(itemNo);
      _autoSaveText = '保存中...';
    });
    try {
      final Pep3DraftDetail detail = await widget.client.saveDraftItem(
        token,
        <String, dynamic>{
          'draftId': draftId,
          'itemNo': itemNo,
          if (_itemScores.containsKey(itemNo)) 'score': _itemScores[itemNo],
          'recordValues': _cleanRecordValues(itemNo),
        },
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _draft = detail;
        _savingItems.remove(itemNo);
        _savedItems.add(itemNo);
        _autoSaveText = '已自动保存';
      });
    } on Object catch (error) {
      if (mounted) {
        setState(() => _savingItems.remove(itemNo));
      }
      _showMessage('第$itemNo题自动保存失败：$error');
    }
  }

  Future<void> _submitDraft() async {
    if (_submitting) {
      return;
    }
    if (_missingCount > 0) {
      _showMessage('还有 $_missingCount 道题未评分，请补全后再提交');
      _jumpToMissing();
      return;
    }
    setState(() => _submitting = true);
    try {
      final Pep3DraftDetail? detail = await _saveDraft(silent: true);
      final int draftId = detail?.id ?? _draft?.id ?? 0;
      if (draftId <= 0) {
        return;
      }
      final String token = await _readToken();
      await widget.client.submitDraft(token, draftId);
      _showMessage('已提交正式测评记录', tone: PadMessageTone.success);
      await Future<void>.delayed(const Duration(milliseconds: 650));
      if (mounted) {
        widget.onBack();
      }
    } on Object catch (error) {
      _showMessage('提交记录失败：$error');
    } finally {
      if (mounted) {
        setState(() => _submitting = false);
      }
    }
  }

  Future<void> _refreshCaregiverInvite({bool silent = false}) async {
    final int draftId = _draft?.id ?? 0;
    if (draftId <= 0) {
      if (!silent) {
        _showMessage('请先保存草稿，再生成照护者报告入口');
      }
      return;
    }
    if (_caregiverInviteDraftId == draftId && _caregiverInvite != null) {
      return;
    }
    final Future<void>? currentRequest = _caregiverInviteRequest;
    if (currentRequest != null) {
      await currentRequest;
      return;
    }
    final Future<void> request = _loadCaregiverInvite(draftId, silent: silent);
    _caregiverInviteRequest = request;
    try {
      await request;
    } finally {
      if (identical(_caregiverInviteRequest, request)) {
        _caregiverInviteRequest = null;
      }
    }
  }

  Future<void> _loadCaregiverInvite(
    int draftId, {
    required bool silent,
  }) async {
    setState(() => _caregiverLoading = true);
    try {
      final String token = await _readToken();
      final Pep3CaregiverInvite invite =
          await widget.client.inviteCaregiverReport(token, draftId);
      if (!mounted) {
        return;
      }
      if ((_draft?.id ?? 0) != draftId) {
        return;
      }
      setState(() {
        _caregiverInvite = invite;
        _caregiverInviteDraftId = draftId;
      });
      if (!silent) {
        _showMessage('照护者报告入口已生成', tone: PadMessageTone.success);
      }
    } on Object catch (error) {
      if (!silent) {
        _showMessage('生成照护者报告入口失败：$error');
      }
    } finally {
      if (mounted) {
        setState(() => _caregiverLoading = false);
      }
    }
  }

  Map<String, dynamic> _buildDraftPayload() {
    final List<Map<String, int>> itemScoreList = _itemScores.entries
        .map((MapEntry<int, int> entry) => <String, int>{
              'itemNo': entry.key,
              'score': entry.value,
            })
        .toList()
      ..sort((Map<String, int> a, Map<String, int> b) =>
          a['itemNo']!.compareTo(b['itemNo']!));
    final List<Map<String, dynamic>> recordValueList = <Map<String, dynamic>>[];
    for (final MapEntry<int, Map<String, dynamic>> item
        in _recordValues.entries) {
      for (final MapEntry<String, dynamic> field in item.value.entries) {
        if (!_isEmptyRecordValue(field.value)) {
          recordValueList.add(<String, dynamic>{
            'itemNo': item.key,
            'fieldKey': field.key,
            'value': field.value,
          });
        }
      }
    }
    recordValueList.sort((Map<String, dynamic> a, Map<String, dynamic> b) {
      final int itemCompare =
          (a['itemNo'] as int).compareTo(b['itemNo'] as int);
      if (itemCompare != 0) {
        return itemCompare;
      }
      return '${a['fieldKey']}'.compareTo('${b['fieldKey']}');
    });
    return <String, dynamic>{
      if ((_draft?.id ?? widget.args.draftId) > 0)
        'id': _draft?.id ?? widget.args.draftId,
      'studentId': _studentId,
      'studentName': _studentName,
      'examinerName': _examinerName,
      if (_birthDate.trim().isNotEmpty) 'birthDate': _birthDate,
      'assessmentDate': _assessmentDate,
      'allowMissingItems': true,
      'itemScoreList': itemScoreList,
      'itemRecordValueList': recordValueList,
    };
  }

  Map<String, dynamic> _cleanRecordValues(int itemNo) {
    final Map<String, dynamic> values =
        _recordValues[itemNo] ?? <String, dynamic>{};
    return Map<String, dynamic>.fromEntries(
      values.entries.where((MapEntry<String, dynamic> entry) =>
          entry.key.trim().isNotEmpty && !_isEmptyRecordValue(entry.value)),
    );
  }

  void _setScore(int score) {
    if (_currentItemNo <= 0) {
      return;
    }
    setState(() => _itemScores[_currentItemNo] = score);
    _saveCurrentItem();
    if (_autoNext && _hasNextItem) {
      Future<void>.delayed(const Duration(milliseconds: 180), () {
        if (mounted) {
          _goNext();
        }
      });
    }
  }

  void _setRecordValue(String key, dynamic value) {
    _setRecordValueForItem(_currentItemNo, key, value);
  }

  void _setRecordValueForItem(int itemNo, String key, dynamic value) {
    if (itemNo <= 0) {
      return;
    }
    setState(() {
      if (_isEmptyRecordValue(value)) {
        _recordValues[itemNo]?.remove(key);
      } else {
        _recordValues.putIfAbsent(itemNo, () => <String, dynamic>{});
        _recordValues[itemNo]![key] = value;
      }
    });
    _saveItem(itemNo);
  }

  Future<void> _goToItem(int itemNo) async {
    if (!_template.allItems
        .any((Pep3ItemSummary item) => item.itemNo == itemNo)) {
      return;
    }
    if (itemNo == _currentItemNo && !_itemLoading) {
      _keepActiveItemVisible();
      return;
    }
    FocusManager.instance.primaryFocus?.unfocus();
    final bool shouldLoad = !_itemCache.containsKey(itemNo);
    setState(() {
      _currentItemNo = itemNo;
      _expandedGroupKey = _groupKeyForItem(itemNo);
      _itemLoading = shouldLoad;
    });
    if (_questionScrollController.hasClients) {
      _questionScrollController.jumpTo(0);
    }
    await _loadCurrentItem();
    _keepActiveItemVisible();
  }

  void _goPrevious() {
    if (!_hasPreviousItem) {
      return;
    }
    _goToItem(_template.allItems[_currentIndex - 1].itemNo);
  }

  void _goNext() {
    if (!_hasNextItem) {
      return;
    }
    _goToItem(_template.allItems[_currentIndex + 1].itemNo);
  }

  void _jumpToMissing() {
    final Pep3ItemSummary? missing = _template.allItems
        .where((Pep3ItemSummary item) => !_itemScores.containsKey(item.itemNo))
        .cast<Pep3ItemSummary?>()
        .firstOrNull;
    if (missing == null) {
      _showMessage('当前没有缺题');
      return;
    }
    _goToItem(missing.itemNo);
  }

  void _keepActiveItemVisible() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted ||
          !_leftScrollController.hasClients ||
          _currentItemNo <= 0) {
        return;
      }
      final BuildContext? itemContext = _activeNavItemKey.currentContext;
      if (itemContext == null) {
        return;
      }
      Scrollable.ensureVisible(
        itemContext,
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOut,
        alignment: .34,
        alignmentPolicy: ScrollPositionAlignmentPolicy.explicit,
      );
    });
  }

  void _keepPageGroupVisible(String groupKey) {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted || !_leftScrollController.hasClients) {
        return;
      }
      final BuildContext? groupContext =
          _pageGroupKeys[groupKey]?.currentContext;
      if (groupContext == null) {
        return;
      }
      Scrollable.ensureVisible(
        groupContext,
        duration: const Duration(milliseconds: 180),
        curve: Curves.easeOut,
        alignment: .02,
        alignmentPolicy: ScrollPositionAlignmentPolicy.explicit,
      );
    });
  }

  void _showMessage(
    String message, {
    PadMessageTone tone = PadMessageTone.info,
  }) {
    if (!mounted || message.trim().isEmpty) {
      return;
    }
    _messageController.show(
      context,
      message,
      tone: tone,
      topMargin: 12,
      key: 'pep3-top-message',
    );
  }

  int get _studentId => _draft?.studentId != null && _draft!.studentId > 0
      ? _draft!.studentId
      : widget.args.studentId;

  int get _totalCount =>
      _template.itemCount > 0 ? _template.itemCount : _template.allItems.length;

  int get _answeredCount => _itemScores.length;

  int get _missingCount => math.max(_totalCount - _answeredCount, 0);

  int get _progressPercent =>
      _totalCount == 0 ? 0 : ((_answeredCount / _totalCount) * 100).round();

  int get _currentIndex {
    final int index = _template.allItems
        .indexWhere((Pep3ItemSummary item) => item.itemNo == _currentItemNo);
    return index < 0 ? 0 : index;
  }

  bool get _hasPreviousItem => _currentIndex > 0;

  bool get _hasNextItem => _currentIndex < _template.allItems.length - 1;

  Pep3ItemSummary? get _currentSummary {
    if (_template.allItems.isEmpty) {
      return null;
    }
    return _template.allItems[_currentIndex];
  }

  Pep3AssessmentItem? get _currentItem => _itemCache[_currentItemNo];

  List<Pep3ScoreOption> get _currentScoreOptions {
    final List<Pep3ScoreOption> options =
        _currentItem?.scoreOptions ?? <Pep3ScoreOption>[];
    final List<Pep3ScoreOption> resolved =
        options.isNotEmpty ? options : _template.scoreOptions;
    return resolved.toList()
      ..sort(
          (Pep3ScoreOption a, Pep3ScoreOption b) => b.value.compareTo(a.value));
  }

  String _groupKeyForItem(int itemNo) {
    for (final Pep3ItemGroupSummary group in _template.itemGroups) {
      if (group.items.any((Pep3ItemSummary item) => item.itemNo == itemNo)) {
        return group.key;
      }
    }
    return '';
  }

  String get _scaleTitle {
    final String raw = widget.args.scaleName.trim();
    if (raw.isEmpty ||
        RegExp(r'pep[-\s]?3', caseSensitive: false).hasMatch(raw)) {
      return 'PEP-3';
    }
    return raw;
  }

  String get _studentAgeText {
    final String calculated = _assessmentAgeText(_birthDate, _assessmentDate);
    if (calculated.isNotEmpty) {
      return calculated;
    }
    return _studentAge.trim().isEmpty ? '未知' : _studentAge.trim();
  }
}
