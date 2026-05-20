import 'dart:async';

import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_age_formatter.dart';
import 'assessment_scale_client.dart';
import 'home_client.dart';
import 'pad_top_message.dart';
import 'vbmapp_assessment_client.dart';

part 'vbmapp_assessment_data.dart';

class VbmappAssessmentLaunchArgs {
  const VbmappAssessmentLaunchArgs({
    this.draftId = 0,
    this.studentId = 0,
    this.studentName = '',
    this.studentAge = '',
    this.birthDate = '',
    this.assessmentDate = '',
    this.examinerName = '',
    this.scaleName = 'VB-MAPP语言行为里程碑评估及安置计划',
  });

  final int draftId;
  final int studentId;
  final String studentName;
  final String studentAge;
  final String birthDate;
  final String assessmentDate;
  final String examinerName;
  final String scaleName;
}

class VbmappAssessmentPage extends StatefulWidget {
  const VbmappAssessmentPage({
    required this.onBack,
    this.args = const VbmappAssessmentLaunchArgs(),
    this.client = const ApiVbmappAssessmentClient(),
    this.homeClient = const ApiHomeClient(),
    super.key,
  });

  final VoidCallback onBack;
  final VbmappAssessmentLaunchArgs args;
  final VbmappAssessmentClient client;
  final HomeClient homeClient;

  @override
  State<VbmappAssessmentPage> createState() => _VbmappAssessmentPageState();
}

class _VbmappAssessmentPageState extends State<VbmappAssessmentPage> {
  static const String _authTokenStorageKey = 'auth_token';
  static const String _scaleVersion = 'VBMAPP_CN_2ND_DRAFT_2026_05';
  static const int _totalItemCount = 212;

  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();
  final Map<String, double> _milestoneScores = <String, double>{};
  final Map<String, int> _barrierScores = <String, int>{};
  final Map<String, int> _transitionScores = <String, int>{};

  String _token = '';
  String _studentName = '';
  String _studentAge = '';
  String _birthDate = '';
  String _examinerName = '';
  String _assessmentDate = '';
  String _selectedModuleCode = _vbmappModules.first.code;
  String _autoSaveText = '等待作答';
  int _draftId = 0;
  int _studentId = 0;
  int _selectedItemIndex = 0;
  bool _loading = true;
  bool _saving = false;
  bool _submitting = false;
  bool _autoNext = true;

  @override
  void initState() {
    super.initState();
    _draftId = widget.args.draftId;
    _studentId = widget.args.studentId;
    _studentName = widget.args.studentName.trim();
    _studentAge = widget.args.studentAge.trim();
    _birthDate = _dateOnlyText(widget.args.birthDate);
    _assessmentDate = _dateOnlyText(widget.args.assessmentDate).isEmpty
        ? _todayIsoDate()
        : _dateOnlyText(widget.args.assessmentDate);
    _examinerName = widget.args.examinerName.trim();
    unawaited(_initialize());
  }

  @override
  void dispose() {
    _messageController.dispose();
    super.dispose();
  }

  Future<void> _initialize() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_authTokenStorageKey) ?? '';
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
      if (launchDraft != null) {
        _applyDraftDetail(launchDraft);
      }
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
    final _VbmappItem? firstMissing = _firstMissingItem();
    if (firstMissing != null) {
      _selectedModuleCode = firstMissing.moduleCode;
      final int missingIndex = _itemsForModule(firstMissing.moduleCode)
          .indexWhere(
              (_VbmappItem item) => item.itemCode == firstMissing.itemCode);
      _selectedItemIndex = missingIndex < 0 ? 0 : missingIndex;
    }
  }

  List<_VbmappItem> get _selectedItems {
    return _itemsForModule(_selectedModuleCode);
  }

  _VbmappItem get _selectedItem {
    final List<_VbmappItem> items = _selectedItems;
    return items[_selectedItemIndex.clamp(0, items.length - 1)];
  }

  int get _answeredCount {
    return _milestoneScores.length +
        _barrierScores.length +
        _transitionScores.length;
  }

  double get _progressPercent {
    return _answeredCount / _totalItemCount;
  }

  _VbmappScoreSnapshot get _scoreSnapshot {
    return _VbmappScoreSnapshot(
      milestoneTotal: _milestoneScores.values.fold<double>(
        0,
        (double total, double score) => total + score,
      ),
      milestoneMax: _milestoneItems.length,
      barrierTotal: _barrierScores.values.fold<int>(
        0,
        (int total, int score) => total + score,
      ),
      barrierMax: _barrierItems.length * 4,
      transitionTotal: _transitionScores.values.fold<int>(
        0,
        (int total, int score) => total + score,
      ),
      transitionMax: _transitionItems.length * 5,
      milestoneDomains: _milestoneDomainSummaries,
    );
  }

  List<_VbmappDomainScoreSummary> get _milestoneDomainSummaries {
    final Map<String, List<_VbmappItem>> groupedItems =
        <String, List<_VbmappItem>>{};
    for (final _VbmappItem item in _milestoneItems) {
      groupedItems
          .putIfAbsent(item.domainName, () => <_VbmappItem>[])
          .add(item);
    }
    return groupedItems.entries
        .map((MapEntry<String, List<_VbmappItem>> entry) {
      final List<_VbmappItem> items = entry.value;
      final int answered = items
          .where(
              (_VbmappItem item) => _milestoneScores.containsKey(item.itemCode))
          .length;
      final double score = items.fold<double>(
        0,
        (double total, _VbmappItem item) =>
            total + (_milestoneScores[item.itemCode] ?? 0),
      );
      return _VbmappDomainScoreSummary(
        name: entry.key,
        score: score,
        maxScore: items.length,
        answered: answered,
        total: items.length,
      );
    }).toList(growable: false);
  }

  String get _studentAgeText {
    final String fallback = _studentAge.trim().isEmpty ? '未知' : _studentAge;
    return formatAssessmentAgeText(
      birthDate: _birthDate,
      assessmentDate: _assessmentDate,
      fallback: fallback,
    );
  }

  void _selectModule(String code) {
    if (_selectedModuleCode == code) {
      return;
    }
    setState(() {
      _selectedModuleCode = code;
      _selectedItemIndex = 0;
    });
  }

  void _selectScore(num score) {
    final _VbmappItem item = _selectedItem;
    setState(() {
      switch (item.moduleCode) {
        case 'milestones':
          _milestoneScores[item.itemCode] = score.toDouble();
          break;
        case 'barriers':
          _barrierScores[item.itemCode] = score.toInt();
          break;
        case 'transition':
          _transitionScores[item.itemCode] = score.toInt();
          break;
      }
      _autoSaveText = '已记录';
    });
    if (_autoNext) {
      Future<void>.delayed(const Duration(milliseconds: 220), () {
        if (mounted) {
          _goNext();
        }
      });
    }
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

  void _goPrevious() {
    if (_selectedItemIndex > 0) {
      setState(() => _selectedItemIndex -= 1);
      return;
    }
    final int moduleIndex = _vbmappModules.indexWhere(
      (_VbmappModule module) => module.code == _selectedModuleCode,
    );
    if (moduleIndex <= 0) {
      return;
    }
    final String previousCode = _vbmappModules[moduleIndex - 1].code;
    setState(() {
      _selectedModuleCode = previousCode;
      _selectedItemIndex = _itemsForModule(previousCode).length - 1;
    });
  }

  void _goNext() {
    final List<_VbmappItem> items = _selectedItems;
    if (_selectedItemIndex < items.length - 1) {
      setState(() => _selectedItemIndex += 1);
      return;
    }
    final int moduleIndex = _vbmappModules.indexWhere(
      (_VbmappModule module) => module.code == _selectedModuleCode,
    );
    if (moduleIndex < 0 || moduleIndex >= _vbmappModules.length - 1) {
      return;
    }
    setState(() {
      _selectedModuleCode = _vbmappModules[moduleIndex + 1].code;
      _selectedItemIndex = 0;
    });
  }

  void _jumpFirstMissing() {
    final _VbmappItem? missing = _firstMissingItem();
    if (missing == null) {
      return;
    }
    setState(() {
      _selectedModuleCode = missing.moduleCode;
      _selectedItemIndex = _itemsForModule(missing.moduleCode)
          .indexWhere((_VbmappItem item) => item.itemCode == missing.itemCode);
    });
  }

  _VbmappItem? _firstMissingItem() {
    for (final _VbmappModule module in _vbmappModules) {
      final List<_VbmappItem> items = _itemsForModule(module.code);
      final int index =
          items.indexWhere((_VbmappItem item) => _scoreFor(item) == null);
      if (index >= 0) {
        return items[index];
      }
    }
    return null;
  }

  void _selectItem(_VbmappItem item) {
    final List<_VbmappItem> items = _itemsForModule(item.moduleCode);
    final int index = items.indexWhere(
      (_VbmappItem candidate) => candidate.itemCode == item.itemCode,
    );
    if (index < 0) {
      return;
    }
    setState(() {
      _selectedModuleCode = item.moduleCode;
      _selectedItemIndex = index;
    });
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
      _autoSaveText = '保存中';
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
          'scaleVersion': _scaleVersion,
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
        _autoSaveText = '草稿已保存';
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
    final int missingCount = _totalItemCount - _answeredCount;
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
      _autoSaveText = '提交中';
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

  @override
  Widget build(BuildContext context) {
    final _VbmappItem item = _selectedItem;
    final _VbmappScoreSnapshot scoreSnapshot = _scoreSnapshot;
    return ColoredBox(
      color: _VbmappColors.page,
      child: Column(
        children: <Widget>[
          _VbmappTopBar(
            scaleName: widget.args.scaleName,
            studentName: _studentName,
            studentAge: _studentAgeText,
            assessmentDate: _assessmentDate,
            examinerName: _examinerName,
            autoSaveText: _autoSaveText,
            saving: _saving,
            submitting: _submitting,
            onBack: widget.onBack,
            onSave: () => unawaited(_saveDraft()),
            onSubmit: () => unawaited(_submitDraft()),
          ),
          if (_loading)
            const Expanded(child: _VbmappLoadingState())
          else ...<Widget>[
            Expanded(
              child: Padding(
                padding: const EdgeInsets.fromLTRB(12, 10, 12, 8),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: <Widget>[
                    SizedBox(
                      width: 250,
                      child: _VbmappModuleRail(
                        modules: _vbmappModules,
                        selectedCode: _selectedModuleCode,
                        selectedItemCode: item.itemCode,
                        items: _selectedItems,
                        answeredCount: _answeredCountByModule,
                        isAnswered: (_VbmappItem item) =>
                            _scoreFor(item) != null,
                        onSelectModule: _selectModule,
                        onSelectItem: _selectItem,
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: _VbmappWorkspace(
                        item: item,
                        score: _scoreFor(item),
                        onSelectScore: _selectScore,
                      ),
                    ),
                    const SizedBox(width: 10),
                    SizedBox(
                      width: 278,
                      child: _VbmappRightRail(
                        progressPercent: _progressPercent,
                        answered: _answeredCount,
                        total: _totalItemCount,
                        selectedModule: _moduleByCode(_selectedModuleCode),
                        scoreSnapshot: scoreSnapshot,
                      ),
                    ),
                  ],
                ),
              ),
            ),
            _VbmappFooterDock(
              current: item.sequenceNo,
              total: _totalItemCount,
              hasPrevious: item.sequenceNo > 1,
              hasNext: item.sequenceNo < _totalItemCount,
              hasMissing: _answeredCount < _totalItemCount,
              autoNext: _autoNext,
              onPrevious: _goPrevious,
              onNext: _goNext,
              onJumpMissing: _jumpFirstMissing,
              onToggleAutoNext: (bool value) =>
                  setState(() => _autoNext = value),
            ),
          ],
        ],
      ),
    );
  }

  Map<String, int> get _answeredCountByModule {
    return <String, int>{
      'milestones': _milestoneScores.length,
      'barriers': _barrierScores.length,
      'transition': _transitionScores.length,
    };
  }
}

class _VbmappColors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color ink = Color(0xFF3F2B22);
  static const Color body = Color(0xFF705B50);
  static const Color muted = Color(0xFFA7958B);
  static const Color line = Color(0xFFEAD7C9);
  static const Color lineSoft = Color(0xFFF4E8DF);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95D37);
  static const Color green = Color(0xFF7FA874);
  static const Color blue = Color(0xFF5D7F9F);
}

List<BoxShadow> _vbmappShadow({
  Color color = const Color(0x12B05F32),
  double blur = 18,
  Offset offset = const Offset(0, 10),
}) {
  return <BoxShadow>[BoxShadow(color: color, blurRadius: blur, offset: offset)];
}

class _VbmappTopBar extends StatelessWidget {
  const _VbmappTopBar({
    required this.scaleName,
    required this.studentName,
    required this.studentAge,
    required this.assessmentDate,
    required this.examinerName,
    required this.autoSaveText,
    required this.saving,
    required this.submitting,
    required this.onBack,
    required this.onSave,
    required this.onSubmit,
  });

  final String scaleName;
  final String studentName;
  final String studentAge;
  final String assessmentDate;
  final String examinerName;
  final String autoSaveText;
  final bool saving;
  final bool submitting;
  final VoidCallback onBack;
  final VoidCallback onSave;
  final VoidCallback onSubmit;

  @override
  Widget build(BuildContext context) {
    final String title =
        scaleName.trim().isEmpty || scaleName.contains('VB-MAPP')
            ? 'VB-MAPP语言行为评估'
            : scaleName.trim();
    final String student =
        studentName.trim().isEmpty ? '-' : studentName.trim();
    final String age = studentAge.trim().isEmpty ? '未知' : studentAge.trim();
    final String date =
        assessmentDate.trim().isEmpty ? _todayIsoDate() : assessmentDate;
    final String examiner =
        examinerName.trim().isEmpty ? '-' : examinerName.trim();
    final String status =
        autoSaveText.trim().isEmpty ? '等待作答' : autoSaveText.trim();

    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.98),
        border: Border.all(color: _VbmappColors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _vbmappShadow(color: const Color(0x14B05F32), blur: 14),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1500;
          final List<Widget> headerChildren = <Widget>[
            Text(
              '$title 测评工作台',
              maxLines: 1,
              softWrap: false,
              style: const TextStyle(
                color: _VbmappColors.ink,
                fontSize: 23,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            _VbmappHeaderMeta(label: '儿童', value: student, compact: compact),
            _VbmappHeaderMeta(label: '年龄', value: age, compact: compact),
            _VbmappHeaderMeta(
              label: compact ? '日期' : '测评日期',
              value: date,
              compact: compact,
            ),
            _VbmappHeaderMeta(
              label: '施测者',
              value: examiner,
              compact: compact,
            ),
          ];
          return Row(
            children: <Widget>[
              _VbmappIconButtonBox(
                icon: Icons.chevron_left_rounded,
                onTap: onBack,
              ),
              const SizedBox(width: 8),
              Expanded(
                child: compact
                    ? SingleChildScrollView(
                        scrollDirection: Axis.horizontal,
                        physics: const ClampingScrollPhysics(),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: headerChildren,
                        ),
                      )
                    : Row(children: headerChildren),
              ),
              _VbmappSaveStatusLabel(text: status, saving: saving),
              const SizedBox(width: 8),
              _VbmappTopActionButton(
                label: saving ? '保存中' : '保存草稿',
                icon: Icons.save_outlined,
                filled: false,
                onTap: saving ? null : onSave,
              ),
              const SizedBox(width: 6),
              _VbmappTopActionButton(
                label: submitting ? '提交中' : '提交记录',
                icon: Icons.fact_check_outlined,
                filled: true,
                onTap: submitting ? null : onSubmit,
              ),
            ],
          );
        },
      ),
    );
  }
}

class _VbmappHeaderMeta extends StatelessWidget {
  const _VbmappHeaderMeta({
    required this.label,
    required this.value,
    required this.compact,
  });

  final String label;
  final String value;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: EdgeInsets.only(left: compact ? 6 : 10),
      padding: EdgeInsets.only(left: compact ? 6 : 10),
      decoration: const BoxDecoration(
        border: Border(left: BorderSide(color: _VbmappColors.line)),
      ),
      child: Text.rich(
        TextSpan(
          children: <InlineSpan>[
            TextSpan(text: '$label：'),
            TextSpan(
              text: value,
              style: const TextStyle(fontWeight: FontWeight.w900),
            ),
          ],
        ),
        maxLines: 1,
        softWrap: false,
        style: const TextStyle(
          color: _VbmappColors.body,
          fontSize: 13,
          height: 1,
          fontWeight: FontWeight.w700,
        ),
      ),
    );
  }
}

class _VbmappSaveStatusLabel extends StatelessWidget {
  const _VbmappSaveStatusLabel({required this.text, required this.saving});

  final String text;
  final bool saving;

  bool get _activeSaving {
    return saving || text.contains('保存中');
  }

  @override
  Widget build(BuildContext context) {
    return ConstrainedBox(
      constraints: const BoxConstraints(minWidth: 112),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.end,
        children: <Widget>[
          Icon(
            _activeSaving
                ? Icons.sync_rounded
                : Icons.check_circle_outline_rounded,
            color:
                _activeSaving ? _VbmappColors.orangeDeep : _VbmappColors.green,
            size: _activeSaving ? 17 : 18,
          ),
          const SizedBox(width: 5),
          Text(
            text,
            maxLines: 1,
            softWrap: false,
            textAlign: TextAlign.right,
            style: TextStyle(
              color:
                  _activeSaving ? _VbmappColors.orangeDeep : _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappIconButtonBox extends StatelessWidget {
  const _VbmappIconButtonBox({required this.icon, required this.onTap});

  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Ink(
          width: 46,
          height: 46,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: _VbmappColors.line),
          ),
          child: Icon(icon, color: _VbmappColors.body, size: 34),
        ),
      ),
    );
  }
}

class _VbmappTopActionButton extends StatelessWidget {
  const _VbmappTopActionButton({
    required this.label,
    required this.icon,
    required this.filled,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final bool filled;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final bool enabled = onTap != null;
    final Color foreground = filled
        ? Colors.white
        : enabled
            ? _VbmappColors.orangeDeep
            : _VbmappColors.muted;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 34,
          padding: const EdgeInsets.symmetric(horizontal: 11),
          decoration: BoxDecoration(
            color: filled
                ? enabled
                    ? _VbmappColors.orange
                    : const Color(0xFFE7DDD6)
                : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: enabled ? _VbmappColors.orange : const Color(0xFFE2D6CE),
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(icon, size: 18, color: foreground),
              const SizedBox(width: 6),
              Text(
                label,
                style: TextStyle(
                  color: foreground,
                  fontSize: 13,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappModuleRail extends StatelessWidget {
  const _VbmappModuleRail({
    required this.modules,
    required this.selectedCode,
    required this.selectedItemCode,
    required this.items,
    required this.answeredCount,
    required this.isAnswered,
    required this.onSelectModule,
    required this.onSelectItem,
  });

  final List<_VbmappModule> modules;
  final String selectedCode;
  final String selectedItemCode;
  final List<_VbmappItem> items;
  final Map<String, int> answeredCount;
  final bool Function(_VbmappItem item) isAnswered;
  final ValueChanged<String> onSelectModule;
  final ValueChanged<_VbmappItem> onSelectItem;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: _vbmappCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          for (final _VbmappModule module in modules) ...<Widget>[
            _VbmappModuleTile(
              module: module,
              selected: module.code == selectedCode,
              answered: answeredCount[module.code] ?? 0,
              onTap: () => onSelectModule(module.code),
            ),
            const SizedBox(height: 6),
          ],
          const SizedBox(height: 2),
          const Divider(height: 1, color: _VbmappColors.lineSoft),
          const SizedBox(height: 8),
          Expanded(
            child: ListView.builder(
              padding: EdgeInsets.zero,
              itemCount: items.length,
              itemBuilder: (BuildContext context, int index) {
                final _VbmappItem item = items[index];
                final _VbmappItem? previous =
                    index > 0 ? items[index - 1] : null;
                final bool showHeader = previous == null ||
                    previous.domainName != item.domainName ||
                    previous.ageBand != item.ageBand;
                return Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: <Widget>[
                    if (showHeader)
                      Padding(
                        padding: EdgeInsets.only(
                          top: index == 0 ? 0 : 10,
                          bottom: 6,
                        ),
                        child: Text(
                          '${item.domainName} · ${item.ageBand}',
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            color: _VbmappColors.muted,
                            fontSize: 12,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ),
                    _VbmappItemNavTile(
                      item: item,
                      selected: item.itemCode == selectedItemCode,
                      answered: isAnswered(item),
                      onTap: () => onSelectItem(item),
                    ),
                    const SizedBox(height: 6),
                  ],
                );
              },
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappModuleTile extends StatelessWidget {
  const _VbmappModuleTile({
    required this.module,
    required this.selected,
    required this.answered,
    required this.onTap,
  });

  final _VbmappModule module;
  final bool selected;
  final int answered;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color accent = selected ? module.color : _VbmappColors.body;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
          decoration: BoxDecoration(
            color: selected ? module.color.withOpacity(.12) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color:
                  selected ? module.color.withOpacity(.55) : _VbmappColors.line,
            ),
          ),
          child: Row(
            children: <Widget>[
              Icon(module.icon, size: 19, color: accent),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  module.title,
                  maxLines: 1,
                  softWrap: false,
                  style: TextStyle(
                    color: selected ? _VbmappColors.ink : _VbmappColors.body,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: selected
                      ? module.color.withOpacity(.14)
                      : const Color(0xFFFFF6EF),
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(
                    color: selected
                        ? module.color.withOpacity(.3)
                        : _VbmappColors.lineSoft,
                  ),
                ),
                child: Text(
                  '$answered/${module.itemCount}',
                  style: TextStyle(
                    color: selected ? module.color : _VbmappColors.body,
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappItemNavTile extends StatelessWidget {
  const _VbmappItemNavTile({
    required this.item,
    required this.selected,
    required this.answered,
    required this.onTap,
  });

  final _VbmappItem item;
  final bool selected;
  final bool answered;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color accent = item.color;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
          decoration: BoxDecoration(
            color: selected ? accent.withOpacity(.12) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color:
                  selected ? accent.withOpacity(.55) : _VbmappColors.lineSoft,
            ),
          ),
          child: Row(
            children: <Widget>[
              Expanded(
                child: Text(
                  item.navCode,
                  maxLines: 1,
                  softWrap: false,
                  style: TextStyle(
                    color: selected ? _VbmappColors.ink : _VbmappColors.body,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              if (answered)
                Container(
                  width: 20,
                  height: 20,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: accent,
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(Icons.check, color: Colors.white, size: 14),
                )
              else
                Container(
                  width: 20,
                  height: 20,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFF6EF),
                    shape: BoxShape.circle,
                    border: Border.all(color: _VbmappColors.line),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappWorkspace extends StatelessWidget {
  const _VbmappWorkspace({
    required this.item,
    required this.score,
    required this.onSelectScore,
  });

  final _VbmappItem item;
  final num? score;
  final ValueChanged<num> onSelectScore;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(22, 20, 22, 18),
      decoration: _vbmappCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Row(
            children: <Widget>[
              _VbmappPill(text: item.domainName),
              const SizedBox(width: 8),
              _VbmappPill(text: item.ageBand),
              const SizedBox(width: 8),
              _VbmappPill(text: item.assessmentMode),
              const Spacer(),
              Text(
                '${item.sequenceNo} / 212',
                style: const TextStyle(
                  color: _VbmappColors.body,
                  fontSize: 14,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 18),
          Expanded(
            child: SingleChildScrollView(
              padding: EdgeInsets.zero,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  Text(
                    item.label,
                    style: const TextStyle(
                      color: _VbmappColors.ink,
                      fontSize: 18,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const SizedBox(height: 10),
                  Text(
                    item.title,
                    style: const TextStyle(
                      color: _VbmappColors.ink,
                      fontSize: 24,
                      height: 1.38,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),
          const Divider(height: 1, color: _VbmappColors.lineSoft),
          const SizedBox(height: 14),
          _VbmappScoreDock(
            item: item,
            score: score,
            onSelectScore: onSelectScore,
          ),
        ],
      ),
    );
  }
}

class _VbmappScoreDock extends StatelessWidget {
  const _VbmappScoreDock({
    required this.item,
    required this.score,
    required this.onSelectScore,
  });

  final _VbmappItem item;
  final num? score;
  final ValueChanged<num> onSelectScore;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Text(
          item.scoreTitle,
          style: const TextStyle(
            color: _VbmappColors.body,
            fontSize: 15,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 12),
        _VbmappScoreOptionGrid(
          options: item.scoreOptions,
          selectedScore: score,
          accent: item.color,
          onSelectScore: onSelectScore,
        ),
        const SizedBox(height: 12),
        _VbmappMaterialHint(item: item),
      ],
    );
  }
}

class _VbmappScoreOptionGrid extends StatelessWidget {
  const _VbmappScoreOptionGrid({
    required this.options,
    required this.selectedScore,
    required this.accent,
    required this.onSelectScore,
  });

  final List<_VbmappScoreOption> options;
  final num? selectedScore;
  final Color accent;
  final ValueChanged<num> onSelectScore;

  int _columnCount(double width) {
    if (options.length <= 3) {
      return options.length.clamp(1, 3);
    }
    return width < 560 ? 2 : 3;
  }

  List<List<_VbmappScoreOption>> _optionRows(int columns) {
    final List<List<_VbmappScoreOption>> rows = <List<_VbmappScoreOption>>[];
    for (int start = 0; start < options.length; start += columns) {
      final int end = (start + columns).clamp(0, options.length);
      rows.add(options.sublist(start, end));
    }
    return rows;
  }

  @override
  Widget build(BuildContext context) {
    const double spacing = 10;
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final int columns = _columnCount(constraints.maxWidth);
        final List<List<_VbmappScoreOption>> rows = _optionRows(columns);
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            for (int rowIndex = 0;
                rowIndex < rows.length;
                rowIndex++) ...<Widget>[
              if (rowIndex > 0) const SizedBox(height: spacing),
              Row(
                children: <Widget>[
                  for (int index = 0;
                      index < rows[rowIndex].length;
                      index++) ...<Widget>[
                    if (index > 0) const SizedBox(width: spacing),
                    Expanded(
                      child: _VbmappScoreOptionButton(
                        option: rows[rowIndex][index],
                        selected: selectedScore == rows[rowIndex][index].score,
                        accent: accent,
                        onTap: () => onSelectScore(rows[rowIndex][index].score),
                      ),
                    ),
                  ],
                ],
              ),
            ],
          ],
        );
      },
    );
  }
}

class _VbmappScoreOptionButton extends StatelessWidget {
  const _VbmappScoreOptionButton({
    required this.option,
    required this.selected,
    required this.accent,
    required this.onTap,
  });

  final _VbmappScoreOption option;
  final bool selected;
  final Color accent;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Ink(
          width: double.infinity,
          height: 90,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            color: selected ? accent : Colors.white,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: selected ? accent : _VbmappColors.line),
            boxShadow: selected
                ? _vbmappShadow(color: accent.withOpacity(.16), blur: 14)
                : null,
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Text(
                option.displayScore,
                style: TextStyle(
                  color: selected ? Colors.white : _VbmappColors.ink,
                  fontSize: 22,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 5),
              Text(
                option.label,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  color: selected ? Colors.white : _VbmappColors.body,
                  fontSize: 12,
                  height: 1.18,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappMaterialHint extends StatelessWidget {
  const _VbmappMaterialHint({required this.item});

  final _VbmappItem item;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(maxHeight: 90),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Icon(Icons.inventory_2_outlined, color: item.color, size: 21),
          const SizedBox(width: 10),
          Expanded(
            child: SingleChildScrollView(
              padding: EdgeInsets.zero,
              physics: const ClampingScrollPhysics(),
              child: Text(
                item.materialHint,
                style: const TextStyle(
                  color: _VbmappColors.body,
                  fontSize: 13,
                  height: 1.38,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappRightRail extends StatelessWidget {
  const _VbmappRightRail({
    required this.progressPercent,
    required this.answered,
    required this.total,
    required this.selectedModule,
    required this.scoreSnapshot,
  });

  final double progressPercent;
  final int answered;
  final int total;
  final _VbmappModule selectedModule;
  final _VbmappScoreSnapshot scoreSnapshot;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(15),
      decoration: _vbmappCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          const Text(
            '测评进度',
            style: TextStyle(
              color: _VbmappColors.ink,
              fontSize: 18,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 14),
          ClipRRect(
            borderRadius: BorderRadius.circular(999),
            child: LinearProgressIndicator(
              value: progressPercent.clamp(0, 1).toDouble(),
              minHeight: 10,
              color: selectedModule.color,
              backgroundColor: _VbmappColors.lineSoft,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            '$answered / $total 项',
            style: const TextStyle(
              color: _VbmappColors.body,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 18),
          _VbmappSummaryStrip(
            label: selectedModule.title,
            value: '${selectedModule.itemCount}',
            subValue: selectedModule.subtitle,
            icon: selectedModule.icon,
            color: selectedModule.color,
          ),
          const SizedBox(height: 12),
          Expanded(
            child: SingleChildScrollView(
              padding: EdgeInsets.zero,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  _VbmappCurrentScoreCard(snapshot: scoreSnapshot),
                  const SizedBox(height: 12),
                  _VbmappMilestoneDomainScoreCard(
                    domains: scoreSnapshot.milestoneDomains,
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 12),
          const _VbmappLegend(),
        ],
      ),
    );
  }
}

class _VbmappSummaryStrip extends StatelessWidget {
  const _VbmappSummaryStrip({
    required this.label,
    required this.value,
    required this.subValue,
    required this.icon,
    this.color = _VbmappColors.orange,
  });

  final String label;
  final String value;
  final String subValue;
  final IconData icon;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: color.withOpacity(.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(.24)),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: color, size: 24),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  label,
                  maxLines: 1,
                  softWrap: false,
                  style: const TextStyle(
                    color: _VbmappColors.body,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  subValue,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _VbmappColors.ink,
                    fontSize: 13,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ),
          ),
          Text(
            value,
            style: const TextStyle(
              color: _VbmappColors.ink,
              fontSize: 20,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappCurrentScoreCard extends StatelessWidget {
  const _VbmappCurrentScoreCard({required this.snapshot});

  final _VbmappScoreSnapshot snapshot;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Row(
            children: <Widget>[
              Icon(
                Icons.insights_rounded,
                color: _VbmappColors.orange,
                size: 20,
              ),
              SizedBox(width: 8),
              Text(
                '当前得分',
                style: TextStyle(
                  color: _VbmappColors.ink,
                  fontSize: 15,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          _VbmappTinyMetric(
            label: '里程碑',
            value: '${snapshot.milestoneScoreText} / ${snapshot.milestoneMax}',
            color: _VbmappColors.orange,
          ),
          const SizedBox(height: 8),
          _VbmappTinyMetric(
            label: '障碍',
            value: '${snapshot.barrierTotal} / ${snapshot.barrierMax}',
            color: _VbmappColors.blue,
          ),
          const SizedBox(height: 8),
          _VbmappTinyMetric(
            label: '转衔',
            value: '${snapshot.transitionTotal} / ${snapshot.transitionMax}',
            color: _VbmappColors.green,
          ),
        ],
      ),
    );
  }
}

class _VbmappMilestoneDomainScoreCard extends StatelessWidget {
  const _VbmappMilestoneDomainScoreCard({required this.domains});

  final List<_VbmappDomainScoreSummary> domains;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          const Text(
            '里程碑领域',
            style: TextStyle(
              color: _VbmappColors.ink,
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 10),
          for (final _VbmappDomainScoreSummary domain in domains) ...<Widget>[
            _VbmappDomainScoreRow(domain: domain),
            if (domain != domains.last) const SizedBox(height: 10),
          ],
        ],
      ),
    );
  }
}

class _VbmappDomainScoreRow extends StatelessWidget {
  const _VbmappDomainScoreRow({required this.domain});

  final _VbmappDomainScoreSummary domain;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        Row(
          children: <Widget>[
            Expanded(
              child: Text(
                domain.name,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _VbmappColors.body,
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            Text(
              '${domain.scoreText}/${domain.maxScore}',
              style: const TextStyle(
                color: _VbmappColors.ink,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(width: 6),
            Text(
              '${domain.answered}/${domain.total}项',
              style: const TextStyle(
                color: _VbmappColors.muted,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        ClipRRect(
          borderRadius: BorderRadius.circular(999),
          child: LinearProgressIndicator(
            value: domain.percent,
            minHeight: 7,
            color: _VbmappColors.orange,
            backgroundColor: _VbmappColors.lineSoft,
          ),
        ),
      ],
    );
  }
}

class _VbmappTinyMetric extends StatelessWidget {
  const _VbmappTinyMetric({
    required this.label,
    required this.value,
    required this.color,
  });

  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Container(
          width: 8,
          height: 8,
          decoration: BoxDecoration(color: color, shape: BoxShape.circle),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            label,
            style: const TextStyle(
              color: _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
        Text(
          value,
          style: const TextStyle(
            color: _VbmappColors.ink,
            fontSize: 14,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}

class _VbmappLegend extends StatelessWidget {
  const _VbmappLegend();

  @override
  Widget build(BuildContext context) {
    return const Row(
      children: <Widget>[
        _VbmappLegendItem(color: _VbmappColors.orange, text: '0 / .5 / 1'),
        SizedBox(width: 8),
        _VbmappLegendItem(color: _VbmappColors.blue, text: '0-4'),
        SizedBox(width: 8),
        _VbmappLegendItem(color: _VbmappColors.green, text: '1-5'),
      ],
    );
  }
}

class _VbmappLegendItem extends StatelessWidget {
  const _VbmappLegendItem({required this.color, required this.text});

  final Color color;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Row(
        children: <Widget>[
          Container(
            width: 8,
            height: 8,
            decoration: BoxDecoration(color: color, shape: BoxShape.circle),
          ),
          const SizedBox(width: 4),
          Expanded(
            child: Text(
              text,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _VbmappColors.body,
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappFooterDock extends StatelessWidget {
  const _VbmappFooterDock({
    required this.current,
    required this.total,
    required this.hasPrevious,
    required this.hasNext,
    required this.hasMissing,
    required this.autoNext,
    required this.onPrevious,
    required this.onNext,
    required this.onJumpMissing,
    required this.onToggleAutoNext,
  });

  final int current;
  final int total;
  final bool hasPrevious;
  final bool hasNext;
  final bool hasMissing;
  final bool autoNext;
  final VoidCallback onPrevious;
  final VoidCallback onNext;
  final VoidCallback onJumpMissing;
  final ValueChanged<bool> onToggleAutoNext;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 62,
      padding: const EdgeInsets.symmetric(horizontal: 20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.97),
        border: Border.all(color: _VbmappColors.line),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(8)),
        boxShadow: _vbmappShadow(color: const Color(0x14B05F32), blur: 16),
      ),
      child: Row(
        children: <Widget>[
          _VbmappFooterButton(
            label: '上一题',
            icon: Icons.chevron_left_rounded,
            enabled: hasPrevious,
            onTap: onPrevious,
          ),
          const Spacer(),
          Text.rich(
            TextSpan(
              children: <InlineSpan>[
                TextSpan(
                  text: '$current',
                  style: const TextStyle(
                    color: _VbmappColors.ink,
                    fontSize: 26,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                TextSpan(
                  text: ' / $total',
                  style: const TextStyle(
                    color: _VbmappColors.body,
                    fontSize: 16,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          const Spacer(),
          _VbmappFooterButton(
            label: '下一题',
            icon: Icons.arrow_forward_rounded,
            enabled: hasNext,
            filled: true,
            reverseIcon: true,
            onTap: onNext,
          ),
          const SizedBox(width: 14),
          _VbmappFooterButton(
            label: '跳到缺题',
            icon: Icons.format_list_bulleted_rounded,
            enabled: hasMissing,
            onTap: onJumpMissing,
          ),
          const SizedBox(width: 22),
          const Text(
            '自动下一题',
            style: TextStyle(
              color: _VbmappColors.body,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 8),
          Switch(
            value: autoNext,
            activeColor: _VbmappColors.orange,
            onChanged: onToggleAutoNext,
          ),
        ],
      ),
    );
  }
}

class _VbmappFooterButton extends StatelessWidget {
  const _VbmappFooterButton({
    required this.label,
    required this.icon,
    required this.enabled,
    required this.onTap,
    this.filled = false,
    this.reverseIcon = false,
  });

  final String label;
  final IconData icon;
  final bool enabled;
  final VoidCallback onTap;
  final bool filled;
  final bool reverseIcon;

  @override
  Widget build(BuildContext context) {
    final Color textColor = enabled
        ? filled
            ? Colors.white
            : _VbmappColors.orangeDeep
        : _VbmappColors.muted;
    final List<Widget> children = <Widget>[
      Icon(icon, size: 21, color: textColor),
      const SizedBox(width: 8),
      Text(
        label,
        style: TextStyle(
          color: textColor,
          fontSize: 15,
          fontWeight: FontWeight.w900,
        ),
      ),
    ];
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: enabled ? onTap : null,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          width: filled ? 140 : 128,
          height: 38,
          decoration: BoxDecoration(
            color: filled && enabled
                ? _VbmappColors.orange
                : enabled
                    ? Colors.white
                    : const Color(0xFFF7F1ED),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: enabled ? _VbmappColors.orange : const Color(0xFFE2D6CE),
            ),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: reverseIcon ? children.reversed.toList() : children,
          ),
        ),
      ),
    );
  }
}

class _VbmappPill extends StatelessWidget {
  const _VbmappPill({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF1E6),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Text(
        text,
        style: const TextStyle(
          color: _VbmappColors.body,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _VbmappLoadingState extends StatelessWidget {
  const _VbmappLoadingState();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Text(
        '正在载入VB-MAPP测评',
        style: TextStyle(
          color: _VbmappColors.body,
          fontSize: 18,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

BoxDecoration _vbmappCardDecoration() {
  return BoxDecoration(
    color: Colors.white.withOpacity(.94),
    borderRadius: BorderRadius.circular(16),
    border: Border.all(color: _VbmappColors.line),
    boxShadow: _vbmappShadow(),
  );
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

String _dateOnlyText(Object? value) {
  if (value is DateTime) {
    return '${value.year.toString().padLeft(4, '0')}-'
        '${value.month.toString().padLeft(2, '0')}-'
        '${value.day.toString().padLeft(2, '0')}';
  }
  final String text = '${value ?? ''}'.trim();
  if (text.length >= 10) {
    return text.substring(0, 10);
  }
  return text;
}
