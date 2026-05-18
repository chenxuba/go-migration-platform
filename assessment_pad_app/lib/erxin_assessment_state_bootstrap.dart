part of 'erxin_assessment_page.dart';

extension _ErxinAssessmentBootstrap on _ErxinAssessmentPageState {
  Future<void> _initialize() async {
    setState(() {
      _loading = true;
      _errorMessage = '';
      _autoSaveText = '等待作答';
    });
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token =
        prefs.getString(_ErxinAssessmentPageState._authTokenStorageKey) ?? '';
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
      final Future<HomeSession> sessionRequest = widget.homeClient == null
          ? Future<HomeSession>.value(HomeSession.fallback)
          : widget.homeClient!.fetchCurrentSession(token).then<HomeSession>(
                (HomeSession session) => session,
                onError: (Object _, StackTrace __) => HomeSession.fallback,
              );
      final List<Object> bootstrapResult =
          await Future.wait<Object>(<Future<Object>>[
        widget.client.fetchTemplateSummary(token),
        sessionRequest,
      ]);
      if (!mounted) {
        return;
      }
      final ErxinTemplateSummary template =
          bootstrapResult[0] as ErxinTemplateSummary;
      final HomeSession session = bootstrapResult[1] as HomeSession;
      final String firstDomain = template.domains.isNotEmpty
          ? template.domains.first.domainCode
          : 'GM';
      _token = token;
      _template = template;
      _selectedDomainCode = firstDomain;
      if (widget.args.draftId > 0) {
        final ErxinDraftDetail detail = await widget.client.fetchDraftDetail(
          token,
          widget.args.draftId,
        );
        _applyDraftDetail(detail);
      } else {
        _detectedDraft = await _findLatestDraft(token);
        _prefetchDetectedDraftDetail(token, _detectedDraft);
      }
      if (_examinerName.trim().isEmpty) {
        _examinerName = _sessionExaminerName(session);
      }
      _selectedItemNo = _firstCurrentItemNo(_selectedDomainCode);
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstVisibleItemNo(_selectedDomainCode);
      }
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
        _errorMessage = '儿心量表加载失败：$error';
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
    final Future<ErxinDraftDetail> request =
        widget.client.fetchDraftDetail(token, draft.id);
    _detectedDraftDetailDraftId = draft.id;
    _detectedDraftDetailRequest = request;
    unawaited(
      request.then<void>(
        (ErxinDraftDetail _) {},
        onError: (Object _, StackTrace __) {},
      ),
    );
  }

  Future<ErxinDraftDetail> _resolveDetectedDraftDetail(
    AssessmentDraftSummary draft,
  ) async {
    final Future<ErxinDraftDetail>? prefetched = _detectedDraftDetailRequest;
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
                message: '当前儿童存在一份未提交的儿心量表测评草稿。',
                metaRows: <AssessmentDraftResumeMetaRow>[
                  AssessmentDraftResumeMetaRow(
                    label: '已记录',
                    value: '${draft.answeredItemCount} 题',
                  ),
                  AssessmentDraftResumeMetaRow(
                    label: '更新时间',
                    value: _formatErxinDateTime(draft.updatedTime),
                  ),
                ],
                accentColor: _ErxinColors.orange,
                inkColor: _ErxinColors.ink,
                bodyColor: _ErxinColors.body,
                lineColor: _ErxinColors.line,
                lineSoftColor: _ErxinColors.line,
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
      _draftProgress = ErxinDraftProgress.empty;
      _itemPasses.clear();
      _itemRemarks.clear();
      _previousStartIndexByDomain.clear();
      _futureEndIndexByDomain.clear();
      _futureVisibleDomains.clear();
      _reviewMonthByDomain.clear();
      _selectedItemNo = _firstVisibleItemNo(_selectedDomainCode);
      _autoSaveText = '已开始新的测评';
    });
    _prefetchSelectedItem();
    final ErxinDraftDetail? detail = await _saveDraft(silent: true);
    if (!mounted) {
      return;
    }
    if (detail == null) {
      _showMessage('新测评草稿创建失败，请手动保存草稿');
      return;
    }
    setState(() {
      _autoSaveText = '新草稿已创建';
    });
  }

  Future<bool> _continueDetectedDraft(AssessmentDraftSummary draft) async {
    if (draft.id <= 0 || _token.trim().isEmpty) {
      return false;
    }
    try {
      final ErxinDraftDetail detail = await _resolveDetectedDraftDetail(draft);
      if (!mounted) {
        return false;
      }
      setState(() {
        _applyDraftDetail(detail);
        _detectedDraft = null;
        _detectedDraftDetailDraftId = 0;
        _detectedDraftDetailRequest = null;
        _selectedItemNo = _firstCurrentItemNo(_selectedDomainCode);
        if (_selectedItemNo <= 0) {
          _selectedItemNo = _firstVisibleItemNo(_selectedDomainCode);
        }
        _autoSaveText = '已恢复最新草稿';
      });
      _prefetchSelectedItem();
      _revealSelectedItem();
      return true;
    } on Object catch (error) {
      if (mounted) {
        _showMessage('恢复草稿失败：$error');
      }
      return false;
    }
  }

  int get _mainAgeMonth {
    final double months = _actualAgeMonths(
      _birthDate,
      _assessmentDate,
    );
    if (months <= 0) {
      return 0;
    }
    int selected = _ErxinAssessmentPageState._standardAgeMonths.first;
    double bestDistance = (months - selected).abs();
    for (final int ageMonth
        in _ErxinAssessmentPageState._standardAgeMonths.skip(1)) {
      final double distance = (months - ageMonth).abs();
      if (distance < bestDistance) {
        selected = ageMonth;
        bestDistance = distance;
      }
    }
    return selected;
  }

  int get _mainAgeIndex {
    final int mainAge = _mainAgeMonth;
    return _ErxinAssessmentPageState._standardAgeMonths.indexOf(mainAge);
  }

  int get _defaultPreviousStartIndex {
    final int index = _mainAgeIndex;
    return index < 0 ? 0 : math.max(0, index - 2);
  }

  int _previousStartIndexForDomain(String domainCode) {
    final int index = _mainAgeIndex;
    if (index < 0) {
      return 0;
    }
    return (_previousStartIndexByDomain[domainCode] ??
            _defaultPreviousStartIndex)
        .clamp(0, index);
  }

  List<int> _previousMonthsForDomain(String domainCode) {
    final int index = _mainAgeIndex;
    if (index < 0) {
      return <int>[];
    }
    final int start = _previousStartIndexForDomain(domainCode);
    return _ErxinAssessmentPageState._standardAgeMonths.sublist(start, index);
  }

  List<int> _visibleMonthsBeforeFutureForDomain(String domainCode) {
    final int mainAge = _mainAgeMonth;
    final List<int> previous = _previousMonthsForDomain(domainCode);
    return <int>[
      mainAge,
      ...previous.reversed,
    ];
  }

  List<int> _futureMonthsForDomain(String domainCode) {
    if (!_futureVisibleDomains.contains(domainCode)) {
      return <int>[];
    }
    final int index = _mainAgeIndex;
    if (index < 0) {
      return <int>[];
    }
    final int end = (_futureEndIndexByDomain[domainCode] ??
            math.min(_ErxinAssessmentPageState._standardAgeMonths.length - 1,
                index + 2))
        .clamp(index, _ErxinAssessmentPageState._standardAgeMonths.length - 1);
    if (index + 1 > end) {
      return <int>[];
    }
    return _ErxinAssessmentPageState._standardAgeMonths
        .sublist(index + 1, end + 1);
  }

  List<int> _visibleMonthsForDomain(String domainCode) {
    return <int>[
      ..._visibleMonthsBeforeFutureForDomain(domainCode),
      ..._futureMonthsForDomain(domainCode),
    ];
  }

  List<int> get _visibleMonths => _visibleMonthsForDomain(_selectedDomainCode);

  bool get _isReviewingRecord {
    return _reviewMonthByDomain.containsKey(_selectedDomainCode);
  }

  int? get _reviewMonth {
    return _reviewMonthByDomain[_selectedDomainCode];
  }

  List<int> _centerMonthsForDomain(String domainCode) {
    final int? reviewMonth = _reviewMonthByDomain[domainCode];
    if (reviewMonth != null) {
      return <int>[reviewMonth];
    }
    return <int>[
      for (final int month in _visibleMonthsForDomain(domainCode))
        if (_itemsFor(domainCode, month).isNotEmpty) month,
    ];
  }

  List<int> get _centerMonths => _centerMonthsForDomain(_selectedDomainCode);

  List<int> _recordMonthsForDomain(String domainCode) {
    final Set<int> months = <int>{..._visibleMonthsForDomain(domainCode)};
    for (final ErxinAgeGroup group in _template.ageGroups) {
      final bool hasAnsweredItem = group.items.any(
        (ErxinItemSummary item) =>
            item.domainCode == domainCode &&
            _itemPasses.containsKey(item.itemNo),
      );
      if (hasAnsweredItem) {
        months.add(group.ageMonth);
      }
    }

    final int mainAge = _mainAgeMonth;
    final List<int> previous = months
        .where((int month) => month > 0 && month < mainAge)
        .toList()
      ..sort((int left, int right) => right.compareTo(left));
    final List<int> future =
        months.where((int month) => month > mainAge).toList()..sort();
    return <int>[
      if (months.contains(mainAge)) mainAge,
      ...previous,
      ...future,
    ];
  }

  bool get _previousMonthsComplete {
    return _previousMonthsCompleteForDomain(_selectedDomainCode);
  }

  bool _previousMonthsCompleteForDomain(String domainCode) {
    final List<int> previous = _previousMonthsForDomain(domainCode);
    return previous.isNotEmpty &&
        previous.every(
          (int month) => _ageMonthComplete(domainCode, month),
        );
  }

  bool get _hasPreviousBaseline {
    return _hasPreviousBaselineForDomain(_selectedDomainCode);
  }

  bool _hasPreviousBaselineForDomain(String domainCode) {
    return _previousBaselineMonthsForDomain(domainCode).isNotEmpty;
  }

  List<int> _previousBaselineMonthsForDomain(String domainCode) {
    final int mainIndex = _mainAgeIndex;
    if (mainIndex <= 1) {
      return const <int>[];
    }
    final int currentStart = _previousStartIndexForDomain(domainCode);
    for (int index = mainIndex - 2; index >= currentStart; index--) {
      final int lowerMonth =
          _ErxinAssessmentPageState._standardAgeMonths[index];
      final int upperMonth =
          _ErxinAssessmentPageState._standardAgeMonths[index + 1];
      if (_ageMonthAllPassed(domainCode, lowerMonth) &&
          _ageMonthAllPassed(domainCode, upperMonth)) {
        return <int>[upperMonth, lowerMonth];
      }
    }
    return const <int>[];
  }

  bool _mainMonthCompleteForDomain(String domainCode) {
    final int mainAge = _mainAgeMonth;
    return mainAge > 0 && _ageMonthComplete(domainCode, mainAge);
  }

  bool get _canContinuePreviousMonths {
    return _canContinuePreviousMonthsForDomain(_selectedDomainCode);
  }

  bool _canContinuePreviousMonthsForDomain(String domainCode) {
    if (_hasPreviousBaselineForDomain(domainCode) ||
        !_previousMonthsCompleteForDomain(domainCode) ||
        !_mainMonthCompleteForDomain(domainCode) ||
        _previousStartIndexForDomain(domainCode) <= 0) {
      return false;
    }
    return true;
  }

  bool get _canEnterFutureMonths {
    return _canEnterFutureMonthsForDomain(_selectedDomainCode);
  }

  bool _canEnterFutureMonthsForDomain(String domainCode) {
    return !_futureVisibleDomains.contains(domainCode) &&
        _mainAgeIndex <
            _ErxinAssessmentPageState._standardAgeMonths.length - 1 &&
        _mainMonthCompleteForDomain(domainCode) &&
        _previousSearchResolvedForDomain(domainCode);
  }

  bool get _futureMonthsComplete {
    return _futureMonthsCompleteForDomain(_selectedDomainCode);
  }

  bool _futureMonthsCompleteForDomain(String domainCode) {
    final List<int> future = _futureMonthsForDomain(domainCode);
    return future.isNotEmpty &&
        future.every(
          (int month) => _ageMonthComplete(domainCode, month),
        );
  }

  bool get _hasFutureCeiling {
    return _hasFutureCeilingForDomain(_selectedDomainCode);
  }

  bool _hasFutureCeilingForDomain(String domainCode) {
    return _futureCeilingMonthsForDomain(domainCode).isNotEmpty;
  }

  List<int> _futureCeilingMonthsForDomain(String domainCode) {
    final List<int> future = _futureMonthsForDomain(domainCode);
    for (int index = 0; index < future.length - 1; index++) {
      final int current = future[index];
      final int next = future[index + 1];
      final int currentIndex =
          _ErxinAssessmentPageState._standardAgeMonths.indexOf(current);
      final int nextIndex =
          _ErxinAssessmentPageState._standardAgeMonths.indexOf(next);
      if (nextIndex - currentIndex != 1) {
        continue;
      }
      if (_ageMonthAllFailed(domainCode, current) &&
          _ageMonthAllFailed(domainCode, next)) {
        return <int>[current, next];
      }
    }
    return const <int>[];
  }

  bool get _canContinueFutureMonths {
    return _canContinueFutureMonthsForDomain(_selectedDomainCode);
  }

  bool _canContinueFutureMonthsForDomain(String domainCode) {
    if (!_futureVisibleDomains.contains(domainCode) ||
        _hasFutureCeilingForDomain(domainCode) ||
        !_futureMonthsCompleteForDomain(domainCode)) {
      return false;
    }
    final int index = _mainAgeIndex;
    final int end = _futureEndIndexByDomain[domainCode] ??
        math.min(
            _ErxinAssessmentPageState._standardAgeMonths.length - 1, index + 2);
    return end < _ErxinAssessmentPageState._standardAgeMonths.length - 1;
  }

  bool _previousBoundaryStopForDomain(String domainCode) {
    if (_hasPreviousBaselineForDomain(domainCode) ||
        _previousStartIndexForDomain(domainCode) > 0) {
      return false;
    }
    final List<int> previous = _previousMonthsForDomain(domainCode);
    return previous.isEmpty || _previousMonthsCompleteForDomain(domainCode);
  }

  bool _previousSearchResolvedForDomain(String domainCode) {
    return _hasPreviousBaselineForDomain(domainCode) ||
        _previousBoundaryStopForDomain(domainCode);
  }

  bool _futureBoundaryStopForDomain(String domainCode) {
    final int mainIndex = _mainAgeIndex;
    if (mainIndex < 0) {
      return false;
    }
    if (!_futureVisibleDomains.contains(domainCode)) {
      return mainIndex >=
          _ErxinAssessmentPageState._standardAgeMonths.length - 1;
    }
    if (!_futureMonthsCompleteForDomain(domainCode)) {
      return false;
    }
    final int end = _futureEndIndexByDomain[domainCode] ??
        math.min(_ErxinAssessmentPageState._standardAgeMonths.length - 1,
            mainIndex + 2);
    return end >= _ErxinAssessmentPageState._standardAgeMonths.length - 1 &&
        !_hasFutureCeilingForDomain(domainCode);
  }

  bool _futureSearchResolvedForDomain(String domainCode) {
    return _hasFutureCeilingForDomain(domainCode) ||
        _futureBoundaryStopForDomain(domainCode);
  }
}
