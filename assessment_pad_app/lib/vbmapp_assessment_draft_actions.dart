part of 'vbmapp_assessment_page.dart';

extension _VbmappAssessmentDraftActions on _VbmappAssessmentPageState {
  Future<void> _initialize() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_vbmappAuthTokenStorageKey) ?? '';
    if (token.trim().isEmpty) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _autoSaveText = '请先登录';
      });
      _showMessage('请先登录后再进行VB-MAPP测评', tone: PadMessageTone.error);
      return;
    }
    HomeSession session = HomeSession.fallback;
    try {
      session = await widget.homeClient.fetchCurrentSession(token);
    } on Object {
      session = HomeSession.fallback;
    }
    VbmappAssessmentSchema? smartSchema;
    VbmappMaterialCatalog? materialCatalog;
    try {
      smartSchema = await widget.client.fetchAssessmentSchema(token);
    } on Object catch (error) {
      if (mounted) {
        _showMessage('VB-MAPP智能题库载入失败，先使用基础题库：$error');
      }
    }
    try {
      materialCatalog = await widget.client.fetchMaterialCatalog(
        token,
        moduleCode: 'milestones',
      );
    } on Object catch (error) {
      if (mounted) {
        _showMessage('VB-MAPP素材目录载入失败，先使用基础素材：$error');
      }
    }
    VbmappDraftDetail? launchDraft;
    if (_draftId > 0) {
      try {
        launchDraft = await widget.client.fetchDraftDetail(token, _draftId);
      } on Object catch (error) {
        if (mounted) {
          _showMessage('VB-MAPP草稿载入失败：$error', tone: PadMessageTone.error);
        }
      }
    }
    if (!mounted) {
      return;
    }
    setState(() {
      _token = token;
      if (smartSchema != null) {
        _itemSchemas
          ..clear()
          ..addAll(smartSchema.itemSchemas);
        _materialProfiles
          ..clear()
          ..addAll(smartSchema.materialProfiles);
      }
      _itemMaterialProfiles
        ..clear()
        ..addAll(_itemMaterialProfileMapFromCatalog(materialCatalog));
      if (launchDraft != null) {
        _applyDraftDetail(launchDraft);
      }
      _rebuildScoreDerivedState();
      if (_examinerName.isEmpty) {
        _examinerName = _sessionExaminerName(session);
      }
      _autoSaveText = _draftId > 0 ? '草稿已载入' : '等待作答';
      _loading = false;
    });
  }

  void _applyDraftDetail(VbmappDraftDetail detail) {
    _draftId = detail.id > 0 ? detail.id : _draftId;
    _studentId = detail.studentId > 0 ? detail.studentId : _studentId;
    if (detail.studentName.trim().isNotEmpty) {
      _studentName = detail.studentName.trim();
    }
    if (detail.birthDate.trim().isNotEmpty) {
      _birthDate = _dateOnlyText(detail.birthDate);
    }
    if (detail.assessmentDate.trim().isNotEmpty) {
      _assessmentDate = _dateOnlyText(detail.assessmentDate);
    }
    if (detail.examinerName.trim().isNotEmpty) {
      _examinerName = detail.examinerName.trim();
    }
    _milestoneScores
      ..clear()
      ..addAll(detail.milestoneScores);
    _barrierScores
      ..clear()
      ..addAll(detail.barrierScores);
    _transitionScores
      ..clear()
      ..addAll(detail.transitionScores);
    _restoreMandEvents(detail.itemResponses);
    _restoreMandObservations(detail.itemResponses);
    _rebuildScoreDerivedState();
    final _VbmappItem? firstMissing = _firstMissingItem();
    if (firstMissing != null) {
      _selectedModuleCode = firstMissing.moduleCode;
      final int missingIndex = _itemsForModule(firstMissing.moduleCode)
          .indexWhere(
              (_VbmappItem item) => item.itemCode == firstMissing.itemCode);
      _selectedItemIndex = missingIndex < 0 ? 0 : missingIndex;
      _selectedItemCode.value = _selectedItem.itemCode;
    }
  }

  Future<int> _saveDraft({bool silent = false}) async {
    if (_saving) {
      return _draftId;
    }
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存草稿', tone: PadMessageTone.error);
      return 0;
    }
    if (_studentId <= 0) {
      _showMessage('请先从开始测评页选择学员', tone: PadMessageTone.error);
      return 0;
    }
    setState(() {
      _saving = true;
      _autoSaveText = '保存中...';
    });
    try {
      final VbmappDraftSaveResult result = await widget.client.saveDraft(
        _token,
        <String, dynamic>{
          if (_draftId > 0) 'id': _draftId,
          'studentId': _studentId,
          'studentName': _studentName,
          'examinerName': _examinerName,
          'birthDate': _birthDate,
          'assessmentDate': _assessmentDate,
          'scaleVersion': _vbmappScaleVersion,
          'milestoneScores': _milestoneScores,
          'barrierScores': _barrierScores,
          'transitionScores': _transitionScores,
        },
      );
      if (!mounted) {
        return 0;
      }
      setState(() {
        _draftId = result.id > 0 ? result.id : _draftId;
        _autoSaveText = '已保存 ${_formatClock(DateTime.now())}';
        _saving = false;
      });
      if (!silent) {
        _showMessage('VB-MAPP草稿已保存', tone: PadMessageTone.success);
      }
      return _draftId;
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return 0;
      }
      setState(() {
        _autoSaveText = '保存失败';
        _saving = false;
      });
      _showMessage(error.message, tone: PadMessageTone.error);
      return 0;
    } on Object catch (error) {
      if (!mounted) {
        return 0;
      }
      setState(() {
        _autoSaveText = '保存失败';
        _saving = false;
      });
      _showMessage('VB-MAPP草稿保存失败：$error', tone: PadMessageTone.error);
      return 0;
    }
  }

  Future<void> _submitDraft() async {
    if (_submitting) {
      return;
    }
    final int missingCount = _vbmappTotalItemCount - _answeredCount;
    if (missingCount > 0) {
      _showMessage(
        'VB-MAPP还有 $missingCount 个项目未评分，完成后才能提交正式记录',
        tone: PadMessageTone.error,
      );
      _jumpFirstMissing();
      return;
    }
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再提交记录', tone: PadMessageTone.error);
      return;
    }
    setState(() {
      _submitting = true;
      _autoSaveText = '提交中...';
    });
    try {
      final int draftId = await _saveDraft(silent: true);
      if (draftId <= 0) {
        if (mounted) {
          setState(() => _submitting = false);
        }
        return;
      }
      await widget.client.submitDraft(_token, draftId);
      if (!mounted) {
        return;
      }
      setState(() {
        _autoSaveText = '已提交';
        _submitting = false;
      });
      _showMessage('VB-MAPP正式记录已提交', tone: PadMessageTone.success);
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _autoSaveText = '提交失败';
        _submitting = false;
      });
      _showMessage(error.message, tone: PadMessageTone.error);
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _autoSaveText = '提交失败';
        _submitting = false;
      });
      _showMessage('VB-MAPP记录提交失败：$error', tone: PadMessageTone.error);
    }
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
      key: 'vbmapp-assessment-top-message',
    );
  }

  void _dismissEditingFocus() {
    final FocusManager manager = FocusManager.instance;
    if (manager.primaryFocus != null) {
      manager.primaryFocus?.unfocus();
    }
  }

  Map<String, int> get _answeredCountByModule {
    return _cachedAnsweredCountByModule;
  }
}
