part of 'autismdev_assessment_page.dart';

extension _AutismDevAssessmentStateActions on _AutismDevAssessmentPageState {
  Future<void> _initialize() async {
    setState(() {
      _loading = true;
      _errorMessage = '';
      _detectedDraft = null;
      _detectedDraftDetailDraftId = 0;
      _detectedDraftDetailRequest = null;
      _draftDialogShown = false;
    });
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token =
        prefs.getString(_AutismDevAssessmentPageState._authTokenStorageKey) ??
            '';
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
      final AutismDevTemplateSummary template =
          await widget.client.fetchTemplateSummary(token);
      if (!mounted) {
        return;
      }
      _token = token;
      _template = template;
      final List<AutismDevDomainGroup> displayGroups = _displayDomainGroups;
      _selectedDomainCode =
          displayGroups.isNotEmpty ? displayGroups.first.domainCode : '';
      _selectedItemNo = _firstItemNoInDomain(_selectedDomainCode);
      if (_draftId > 0) {
        final AutismDevDraftDetail detail =
            await widget.client.fetchDraftDetail(token, _draftId);
        _applyDraftDetail(detail);
      } else {
        _detectedDraft = await _findLatestDraft(token);
        _prefetchDetectedDraftDetail(token, _detectedDraft);
      }
      _selectedRangeFilter = '';
      _ensureSelectedDisplayItem();
      _syncRemarkController();
      setState(() {
        _loading = false;
        _autoSaveText = _draftId > 0 ? '已载入草稿' : '已准备';
      });
      _prefetchSelectedItem();
      _showDetectedDraftDialogIfNeeded();
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = '孤独症发展评估表加载失败：$error';
      });
    }
  }

  Future<AssessmentDraftSummary?> _findLatestDraft(String token) async {
    if (_studentId <= 0) {
      return null;
    }
    final AssessmentDraftPage page = await widget.client.fetchDraftsPage(
      token,
      studentId: _studentId,
      pageSize: 1,
      latestOnly: true,
    );
    if (page.items.isEmpty || page.items.first.id <= 0) {
      return null;
    }
    return page.items.first;
  }

  void _prefetchDetectedDraftDetail(
    String token,
    AssessmentDraftSummary? draft,
  ) {
    if (draft == null || draft.id <= 0) {
      _detectedDraftDetailDraftId = 0;
      _detectedDraftDetailRequest = null;
      return;
    }
    if (_detectedDraftDetailDraftId == draft.id &&
        _detectedDraftDetailRequest != null) {
      return;
    }
    final Future<AutismDevDraftDetail> request =
        widget.client.fetchDraftDetail(token, draft.id);
    _detectedDraftDetailDraftId = draft.id;
    _detectedDraftDetailRequest = request;
    unawaited(
      request.then<void>(
        (AutismDevDraftDetail _) {},
        onError: (Object _, StackTrace __) {},
      ),
    );
  }

  Future<AutismDevDraftDetail> _resolveDetectedDraftDetail(
    AssessmentDraftSummary draft,
  ) async {
    final Future<AutismDevDraftDetail>? prefetched =
        _detectedDraftDetailRequest;
    if (_detectedDraftDetailDraftId == draft.id && prefetched != null) {
      try {
        return await prefetched;
      } on Object {
        // Retry below if the prefetch failed.
      }
    }
    return widget.client.fetchDraftDetail(_token, draft.id);
  }

  void _showDetectedDraftDialogIfNeeded() {
    final AssessmentDraftSummary? draft = _detectedDraft;
    if (!mounted || _draftDialogShown || draft == null || draft.id <= 0) {
      return;
    }
    final int answered = draft.answeredItemCount;
    final int total = math.max(_fullTotalCount, answered);
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
                message: '当前儿童存在一份未提交的孤独症儿童发展评估表草稿。',
                metaRows: <AssessmentDraftResumeMetaRow>[
                  AssessmentDraftResumeMetaRow(
                    label: '已完成',
                    value: '$answered / $total 题',
                  ),
                  AssessmentDraftResumeMetaRow(
                    label: '更新时间',
                    value: _formatAutismDevDateTime(draft.updatedTime),
                  ),
                ],
                accentColor: _AutismDevColors.orange,
                inkColor: _AutismDevColors.ink,
                bodyColor: _AutismDevColors.body,
                lineColor: _AutismDevColors.line,
                lineSoftColor: _AutismDevColors.lineSoft,
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
      _detectedDraftDetailDraftId = 0;
      _detectedDraftDetailRequest = null;
      _draftId = 0;
      _itemScores.clear();
      _itemRemarks.clear();
      _selectedRangeFilter = '';
      _assessmentDate = _dateOnlyText(widget.args.assessmentDate).isNotEmpty
          ? _dateOnlyText(widget.args.assessmentDate)
          : _todayIsoDate();
      _examinerName = widget.args.examinerName.trim().isNotEmpty
          ? widget.args.examinerName
          : _examinerName;
      _selectedDomainCode = _displayDomainGroups.isNotEmpty
          ? _displayDomainGroups.first.domainCode
          : '';
      _selectedItemNo = _firstItemNoInDomain(_selectedDomainCode);
      _autoSaveText = '已开始新的测评';
    });
    _syncRemarkController();
    _prefetchSelectedItem();
    final AutismDevDraftDetail? detail = await _saveDraft(silent: true);
    if (!mounted) {
      return;
    }
    if (detail == null) {
      _showMessage('新测评草稿创建失败，请手动保存草稿');
      return;
    }
    setState(() => _autoSaveText = '新草稿已创建');
  }

  Future<bool> _continueDetectedDraft(AssessmentDraftSummary draft) async {
    if (draft.id <= 0 || _token.trim().isEmpty) {
      return false;
    }
    setState(() => _autoSaveText = '正在恢复草稿...');
    try {
      final AutismDevDraftDetail detail =
          await _resolveDetectedDraftDetail(draft);
      if (!mounted) {
        return false;
      }
      setState(() {
        _applyDraftDetail(detail);
        _detectedDraft = null;
        _detectedDraftDetailDraftId = 0;
        _detectedDraftDetailRequest = null;
        _selectedRangeFilter = '';
        _ensureSelectedDisplayItem();
        _autoSaveText = '已恢复最新草稿';
      });
      _syncRemarkController();
      _prefetchSelectedItem();
      return true;
    } on Object catch (error) {
      if (mounted) {
        _showMessage('恢复草稿失败：$error');
      }
      return false;
    }
  }

  void _applyDraftDetail(AutismDevDraftDetail detail) {
    _draftId = detail.id;
    _studentId = detail.studentId > 0 ? detail.studentId : _studentId;
    _studentName = detail.studentName.trim().isNotEmpty
        ? detail.studentName.trim()
        : _studentName;
    _birthDate = detail.birthDate.trim().isNotEmpty
        ? detail.birthDate.trim()
        : _birthDate;
    _assessmentDate = detail.assessmentDate.trim().isNotEmpty
        ? detail.assessmentDate.trim()
        : _assessmentDate;
    _examinerName = detail.examinerName.trim().isNotEmpty
        ? detail.examinerName.trim()
        : _examinerName;
    _itemScores
      ..clear()
      ..addAll(detail.input.itemScores);
    _itemRemarks
      ..clear()
      ..addAll(detail.input.itemRemarks);
    final int firstUnanswered = _firstUnansweredItemNo();
    if (firstUnanswered > 0) {
      _selectedItemNo = firstUnanswered;
      _selectedDomainCode =
          _summaryByNo(firstUnanswered)?.domainCode ?? _selectedDomainCode;
    }
  }

  AutismDevDomainGroup? get _selectedGroup {
    for (final AutismDevDomainGroup group in _displayDomainGroups) {
      if (group.domainCode == _selectedDomainCode) {
        return group;
      }
    }
    return _displayDomainGroups.isNotEmpty ? _displayDomainGroups.first : null;
  }

  AutismDevItemSummary? get _selectedSummary => _summaryByNo(_selectedItemNo);

  AutismDevAssessmentItem? get _selectedDetail =>
      _itemDetailCache[_selectedItemNo];

  int get _fullAnsweredCount => _allItems
      .where(
          (AutismDevItemSummary item) => _itemScores.containsKey(item.itemNo))
      .length;

  int get _visibleAnsweredCount => _displayItems
      .where(
          (AutismDevItemSummary item) => _itemScores.containsKey(item.itemNo))
      .length;

  int get _fullTotalCount {
    if (_template.itemCount > 0) {
      return _template.itemCount;
    }
    return _template.domainGroups.fold<int>(
      0,
      (int total, AutismDevDomainGroup group) => total + group.items.length,
    );
  }

  int get _visibleTotalCount => _displayItems.length;

  int get _visibleMissingCount =>
      math.max(_visibleTotalCount - _visibleAnsweredCount, 0);

  int? get _studentAgeMonths => _resolvedStudentAgeMonths(
        birthDate: _birthDate,
        assessmentDate: _assessmentDate,
        ageText: _studentAge,
      );

  int get _currentIndex {
    final int index = _displayItems.indexWhere(
      (AutismDevItemSummary item) => item.itemNo == _selectedItemNo,
    );
    return index < 0 ? 0 : index;
  }

  bool get _hasPreviousItem => _currentIndex > 0;

  bool get _hasNextItem => _currentIndex < _displayItems.length - 1;

  List<AutismDevScoreOption> get _currentScoreOptions {
    final AutismDevAssessmentItem? detail = _selectedDetail;
    if (detail != null && detail.scoreOptions.isNotEmpty) {
      return detail.scoreOptions;
    }
    final String scoreType =
        detail?.scoreType ?? _selectedSummary?.scoreType ?? '';
    return _template.scoreOptions
        .where((AutismDevScoreOption option) =>
            option.scoreType.toUpperCase() == scoreType.toUpperCase())
        .toList();
  }

  void _selectItem(AutismDevItemSummary item) {
    setState(() {
      if (_selectedRangeFilter.isNotEmpty &&
          _assessmentRangeBucket(item) != _selectedRangeFilter) {
        _selectedRangeFilter = '';
      }
      _selectedItemNo = item.itemNo;
      _selectedDomainCode = item.domainCode;
    });
    _syncRemarkController();
    _prefetchSelectedItem();
  }

  void _selectRangeFilter(String rangeFilter) {
    final AutismDevDomainGroup? group = _selectedGroup;
    if (group == null) {
      return;
    }
    final String nextFilter = rangeFilter.trim();
    if (nextFilter.isEmpty) {
      setState(() => _selectedRangeFilter = '');
      return;
    }
    final List<AutismDevItemSummary> filteredItems = group.items
        .where((AutismDevItemSummary item) =>
            _assessmentRangeBucket(item) == nextFilter)
        .toList(growable: false);
    if (filteredItems.isEmpty) {
      setState(() => _selectedRangeFilter = nextFilter);
      return;
    }
    final AutismDevItemSummary target = filteredItems.firstWhere(
      (AutismDevItemSummary item) => !_itemScores.containsKey(item.itemNo),
      orElse: () => filteredItems.first,
    );
    setState(() {
      _selectedRangeFilter = nextFilter;
      _selectedItemNo = target.itemNo;
      _selectedDomainCode = target.domainCode;
    });
    _syncRemarkController();
    _prefetchSelectedItem();
  }

  Future<void> _openQuestionPreferenceDialog() async {
    final _AutismDevQuestionDisplayPreference? selected =
        await showDialog<_AutismDevQuestionDisplayPreference>(
      context: context,
      barrierColor: Colors.black.withOpacity(.28),
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _QuestionPreferenceDialog(
            selected: _questionDisplayPreference,
            studentAgeMonths: _studentAgeMonths,
          ),
        );
      },
    );
    if (selected == null || selected == _questionDisplayPreference) {
      return;
    }
    setState(() {
      _questionDisplayPreference = selected;
      _selectedRangeFilter = '';
      _ensureSelectedDisplayItem();
    });
    _syncRemarkController();
    _prefetchSelectedItem();
  }

  void _syncRemarkController() {
    final String remark = _itemRemarks[_selectedItemNo] ?? '';
    if (_remarkController.text != remark) {
      _remarkController.text = remark;
    }
  }

  void _syncCurrentRemarkToState() {
    final int itemNo = _selectedItemNo;
    if (itemNo <= 0) {
      return;
    }
    final String remark = _remarkController.text.trim();
    if (remark.isNotEmpty || _itemScores.containsKey(itemNo)) {
      _itemRemarks[itemNo] = remark;
    }
  }

  void _prefetchSelectedItem() {
    final int itemNo = _selectedItemNo;
    _prefetchItemDetail(itemNo, updateStateOnComplete: true);
    _prefetchNextItem();
  }

  void _prefetchNextItem() {
    final List<AutismDevItemSummary> displayItems = _displayItems;
    final int nextIndex = _currentIndex + 1;
    if (nextIndex < 0 || nextIndex >= displayItems.length) {
      return;
    }
    _prefetchItemDetail(
      displayItems[nextIndex].itemNo,
      updateStateOnComplete: false,
    );
  }

  void _prefetchItemDetail(
    int itemNo, {
    required bool updateStateOnComplete,
  }) {
    if (itemNo <= 0 || _itemDetailCache.containsKey(itemNo)) {
      return;
    }
    final Future<AutismDevAssessmentItem> future = _itemDetailFetches[itemNo] ??
        widget.client
            .fetchTemplateItem(_token, itemNo: itemNo)
            .whenComplete(() => _itemDetailFetches.remove(itemNo));
    _itemDetailFetches[itemNo] = future;
    unawaited(
      future.then<void>((AutismDevAssessmentItem item) {
        if (!mounted) {
          _itemDetailCache[itemNo] = item;
          return;
        }
        if (updateStateOnComplete || _selectedItemNo == itemNo) {
          setState(() => _itemDetailCache[itemNo] = item);
        } else {
          _itemDetailCache[itemNo] = item;
        }
      }, onError: (Object _) {}),
    );
  }

  Future<void> _selectScore(String score, {bool moveNext = false}) async {
    if (_selectedItemNo <= 0) {
      return;
    }
    setState(() {
      _itemScores[_selectedItemNo] = score;
      _itemRemarks[_selectedItemNo] = _remarkController.text.trim();
    });
    await _saveCurrentItem(moveNext: moveNext, silent: true);
  }

  Future<void> _saveCurrentItem({
    bool moveNext = false,
    bool silent = false,
  }) async {
    final int itemNo = _selectedItemNo;
    final String? score = _itemScores[itemNo];
    if (itemNo <= 0 || score == null || score.trim().isEmpty) {
      _showMessage('请先选择本题评分');
      return;
    }
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存');
      return;
    }
    setState(() {
      _saving = true;
      _autoSaveText = '保存中...';
      _itemRemarks[itemNo] = _remarkController.text.trim();
    });
    try {
      final AutismDevDraftDetail detail;
      if (_draftId <= 0) {
        detail = await widget.client.saveDraft(_token, _draftPayload());
      } else {
        detail = await widget.client.saveDraftItem(
          _token,
          <String, dynamic>{
            'draftId': _draftId,
            'itemNo': itemNo,
            'score': score,
            'remark': _remarkController.text.trim(),
          },
        );
      }
      if (!mounted) {
        return;
      }
      setState(() {
        _draftId = detail.id;
        _itemScores
          ..clear()
          ..addAll(detail.input.itemScores);
        _itemRemarks
          ..clear()
          ..addAll(detail.input.itemRemarks);
        _saving = false;
        _autoSaveText = '已保存 ${_formatClock(DateTime.now())}';
      });
      _syncRemarkController();
      if (moveNext) {
        _goNextItem();
      } else if (!silent) {
        _showMessage('已保存本题', tone: PadMessageTone.success);
      }
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _saving = false;
        _autoSaveText = '保存失败';
      });
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _saving = false;
        _autoSaveText = '保存失败';
      });
      _showMessage('保存失败：$error');
    }
  }

  Future<AutismDevDraftDetail?> _saveDraft({bool silent = false}) async {
    final Future<AutismDevDraftDetail?>? inFlight = _saveDraftFuture;
    if (inFlight != null) {
      if (!silent && _saveDraftFutureSilent && !_saveDraftJoinedByManual) {
        _saveDraftJoinedByManual = true;
        return _joinSilentDraftSave(inFlight);
      }
      return inFlight;
    }
    late final Future<AutismDevDraftDetail?> trackedFuture;
    _saveDraftFutureSilent = silent;
    _saveDraftJoinedByManual = false;
    trackedFuture = _performSaveDraft(silent: silent).whenComplete(() {
      if (identical(_saveDraftFuture, trackedFuture)) {
        _saveDraftFuture = null;
        _saveDraftFutureSilent = false;
        _saveDraftJoinedByManual = false;
      }
    });
    _saveDraftFuture = trackedFuture;
    return trackedFuture;
  }

  Future<AutismDevDraftDetail?> _joinSilentDraftSave(
    Future<AutismDevDraftDetail?> inFlight,
  ) async {
    if (mounted) {
      setState(() => _autoSaveText = '草稿保存中...');
    }
    final AutismDevDraftDetail? detail = await inFlight;
    if (!mounted) {
      return detail;
    }
    if (detail == null) {
      _showMessage('保存草稿失败，请稍后重试');
      return null;
    }
    _showMessage('草稿已保存', tone: PadMessageTone.success);
    return detail;
  }

  Future<AutismDevDraftDetail?> _performSaveDraft({
    required bool silent,
  }) async {
    if (_saving) {
      return null;
    }
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存');
      return null;
    }
    _syncCurrentRemarkToState();
    setState(() {
      _saving = true;
      _autoSaveText = '保存中...';
    });
    try {
      final AutismDevDraftDetail detail =
          await widget.client.saveDraft(_token, _draftPayload());
      if (!mounted) {
        return detail;
      }
      setState(() {
        _draftId = detail.id;
        _itemScores
          ..clear()
          ..addAll(detail.input.itemScores);
        _itemRemarks
          ..clear()
          ..addAll(detail.input.itemRemarks);
        _saving = false;
        _autoSaveText = '已保存 ${_formatClock(DateTime.now())}';
      });
      _syncRemarkController();
      if (!silent) {
        _showMessage('草稿已保存', tone: PadMessageTone.success);
      }
      return detail;
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return null;
      }
      setState(() {
        _saving = false;
        _autoSaveText = '保存失败';
      });
      _showMessage(error.message);
      return null;
    } on Object catch (error) {
      if (!mounted) {
        return null;
      }
      setState(() {
        _saving = false;
        _autoSaveText = '保存失败';
      });
      _showMessage('保存失败：$error');
      return null;
    }
  }

  Future<void> _submitDraft() async {
    if (_submitting) {
      return;
    }
    _syncCurrentRemarkToState();
    if (_fullTotalCount > 0 && _fullAnsweredCount < _fullTotalCount) {
      _showMessage(
        '还有 ${_fullTotalCount - _fullAnsweredCount} 道题未评分，完成后再提交',
      );
      return;
    }
    setState(() => _submitting = true);
    try {
      if (_draftId <= 0) {
        final AutismDevDraftDetail detail =
            await widget.client.saveDraft(_token, _draftPayload());
        _draftId = detail.id;
      }
      await widget.client.submitDraft(_token, _draftId);
      if (!mounted) {
        return;
      }
      setState(() => _submitting = false);
      _showMessage('正式记录已提交', tone: PadMessageTone.success);
      await Future<void>.delayed(const Duration(milliseconds: 500));
      if (mounted) {
        widget.onBack();
      }
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() => _submitting = false);
      _showMessage(error.message);
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() => _submitting = false);
      _showMessage('提交失败：$error');
    }
  }

  Map<String, dynamic> _draftPayload() {
    return <String, dynamic>{
      if (_draftId > 0) 'id': _draftId,
      if (_studentId > 0) 'studentId': _studentId,
      'studentName': _studentName.trim(),
      'examinerName': _examinerName.trim(),
      'birthDate': _birthDate.trim(),
      'assessmentDate': _assessmentDate.trim(),
      'itemScoreList': _itemScoreList(),
    };
  }

  List<Map<String, dynamic>> _itemScoreList() {
    final List<int> itemNos = _itemScores.keys.toList()..sort();
    return itemNos.map((int itemNo) {
      return <String, dynamic>{
        'itemNo': itemNo,
        'score': _itemScores[itemNo],
        if ((_itemRemarks[itemNo] ?? '').trim().isNotEmpty)
          'remark': _itemRemarks[itemNo]!.trim(),
      };
    }).toList();
  }

  void _goNextItem() {
    final List<AutismDevItemSummary> displayItems = _displayItems;
    final int currentIndex = displayItems.indexWhere(
        (AutismDevItemSummary item) => item.itemNo == _selectedItemNo);
    if (currentIndex >= 0 && currentIndex < displayItems.length - 1) {
      final AutismDevItemSummary next = displayItems[currentIndex + 1];
      _selectItem(next);
    }
  }

  void _goPreviousItem() {
    final List<AutismDevItemSummary> displayItems = _displayItems;
    final int currentIndex = displayItems.indexWhere(
        (AutismDevItemSummary item) => item.itemNo == _selectedItemNo);
    if (currentIndex > 0) {
      _selectItem(displayItems[currentIndex - 1]);
    }
  }

  void _jumpToMissing() {
    final int itemNo = _firstUnansweredItemNo();
    if (itemNo <= 0) {
      _showMessage('当前没有缺题', tone: PadMessageTone.success);
      return;
    }
    final AutismDevItemSummary? item = _summaryByNo(itemNo);
    if (item != null) {
      _selectItem(item);
    }
  }

  List<AutismDevItemSummary> get _allItems {
    return _template.domainGroups
        .expand((AutismDevDomainGroup group) => group.items)
        .toList();
  }

  List<AutismDevDomainGroup> get _displayDomainGroups {
    return _template.domainGroups
        .map(_displayGroupForPreference)
        .where((AutismDevDomainGroup group) => group.items.isNotEmpty)
        .toList(growable: false);
  }

  List<AutismDevItemSummary> get _displayItems {
    return _displayDomainGroups
        .expand((AutismDevDomainGroup group) => group.items)
        .toList(growable: false);
  }

  AutismDevDomainGroup _displayGroupForPreference(
    AutismDevDomainGroup group,
  ) {
    final List<AutismDevItemSummary> items = group.items
        .where(_shouldDisplayItemForPreference)
        .toList(growable: false);
    return AutismDevDomainGroup(
      groupCode: group.groupCode,
      title: group.title,
      domainCode: group.domainCode,
      domainName: group.domainName,
      scoreType: group.scoreType,
      itemCount: items.length,
      items: items,
    );
  }

  bool _shouldDisplayItemForPreference(AutismDevItemSummary item) {
    if (_isEmotionBehaviorItem(item) ||
        _questionDisplayPreference == _AutismDevQuestionDisplayPreference.all) {
      return true;
    }
    final int? ageMonths = _studentAgeMonths;
    if (ageMonths == null) {
      return true;
    }
    final int minMonth = item.ageMinMonth;
    final int maxMonth = item.ageMaxMonth;
    if (minMonth <= 0 && maxMonth <= 0) {
      return true;
    }
    return switch (_questionDisplayPreference) {
      _AutismDevQuestionDisplayPreference.all => true,
      _AutismDevQuestionDisplayPreference.matchingAge =>
        minMonth <= ageMonths && (maxMonth <= 0 || ageMonths <= maxMonth),
      _AutismDevQuestionDisplayPreference.ageAndBelow => minMonth <= ageMonths,
    };
  }

  int _firstItemNoInDomain(String domainCode) {
    for (final AutismDevDomainGroup group in _displayDomainGroups) {
      if (group.domainCode == domainCode && group.items.isNotEmpty) {
        return group.items.first.itemNo;
      }
    }
    return _displayItems.isNotEmpty ? _displayItems.first.itemNo : 0;
  }

  int _firstUnansweredItemNo() {
    for (final AutismDevItemSummary item in _displayItems) {
      if (!_itemScores.containsKey(item.itemNo)) {
        return item.itemNo;
      }
    }
    return 0;
  }

  void _ensureSelectedDisplayItem() {
    final List<AutismDevItemSummary> displayItems = _displayItems;
    if (displayItems.isEmpty) {
      _selectedItemNo = 0;
      _selectedDomainCode = '';
      return;
    }
    for (final AutismDevItemSummary item in displayItems) {
      if (item.itemNo == _selectedItemNo) {
        _selectedDomainCode = item.domainCode;
        return;
      }
    }
    final AutismDevItemSummary target = displayItems.firstWhere(
      (AutismDevItemSummary item) => !_itemScores.containsKey(item.itemNo),
      orElse: () => displayItems.first,
    );
    _selectedItemNo = target.itemNo;
    _selectedDomainCode = target.domainCode;
  }

  AutismDevItemSummary? _summaryByNo(int itemNo) {
    for (final AutismDevItemSummary item in _allItems) {
      if (item.itemNo == itemNo) {
        return item;
      }
    }
    return null;
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
      key: 'autismdev-top-message',
    );
  }

  Widget _buildBody() {
    if (_loading) {
      return const _AutismDevLoadingBody();
    }
    if (_errorMessage.isNotEmpty) {
      return _AutismDevStateBody(
        title: '加载失败',
        message: _errorMessage,
        actionText: '重新加载',
        onAction: _initialize,
      );
    }
    final AutismDevDomainGroup? group = _selectedGroup;
    final AutismDevItemSummary? summary = _selectedSummary;
    if (group == null || summary == null) {
      return _AutismDevStateBody(
        title: '暂无题目',
        message: '当前量表模板没有可用题目',
        actionText: '返回',
        onAction: widget.onBack,
      );
    }
    return Column(
      children: <Widget>[
        Expanded(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(10, 10, 10, 10),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: <Widget>[
                SizedBox(
                  width: 226,
                  child: _AutismDevDomainPanel(
                    groups: _displayDomainGroups,
                    selectedDomainCode: _selectedDomainCode,
                    selectedRangeFilter: _selectedRangeFilter,
                    selectedItemNo: _selectedItemNo,
                    itemScores: _itemScores,
                    onSelectItem: _selectItem,
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: _AutismDevWorkspacePanel(
                    item: summary,
                    detail: _selectedDetail,
                    selectedScore: _itemScores[_selectedItemNo],
                    scoreOptions: _currentScoreOptions,
                    onOpenQuestionPreference: _openQuestionPreferenceDialog,
                    onScore: (String score) =>
                        _selectScore(score, moveNext: _autoNext),
                  ),
                ),
                const SizedBox(width: 10),
                SizedBox(
                  width: 238,
                  child: _AutismDevRightRail(
                    group: group,
                    item: summary,
                    remarkController: _remarkController,
                    selectedRangeFilter: _selectedRangeFilter,
                    itemScores: _itemScores,
                    answeredCount: _visibleAnsweredCount,
                    totalCount: _visibleTotalCount,
                    missingCount: _visibleMissingCount,
                    onSelectRangeFilter: _selectRangeFilter,
                  ),
                ),
              ],
            ),
          ),
        ),
        _AutismDevFooter(
          current: _currentIndex + 1,
          total: _visibleTotalCount,
          hasPrevious: _hasPreviousItem,
          hasNext: _hasNextItem,
          autoNext: _autoNext,
          onPrevious: _goPreviousItem,
          onNext: _goNextItem,
          onJumpMissing: _jumpToMissing,
          onToggleAutoNext: (bool value) => setState(() => _autoNext = value),
        ),
      ],
    );
  }
}
