import 'dart:convert';
import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'home_client.dart';
import 'pad_top_message.dart';
import 'pad_responsive.dart';
import 'pep3_assessment_client.dart';

class Pep3AssessmentPage extends StatefulWidget {
  const Pep3AssessmentPage({
    required this.onBack,
    this.args = const Pep3AssessmentLaunchArgs(),
    this.client = const ApiPep3AssessmentClient(),
    this.homeClient = const ApiHomeClient(),
    super.key,
  });

  final VoidCallback onBack;
  final Pep3AssessmentLaunchArgs args;
  final Pep3AssessmentClient client;
  final HomeClient homeClient;

  @override
  State<Pep3AssessmentPage> createState() => _Pep3AssessmentPageState();
}

class _Pep3AssessmentPageState extends State<Pep3AssessmentPage> {
  static const String _authTokenStorageKey = 'auth_token';

  Pep3TemplateSummary _template = Pep3TemplateSummary.empty;
  final Map<int, Pep3AssessmentItem> _itemCache = <int, Pep3AssessmentItem>{};
  final Map<int, Future<Pep3AssessmentItem>> _itemFetches =
      <int, Future<Pep3AssessmentItem>>{};
  final Map<int, int> _itemScores = <int, int>{};
  final Map<int, Map<String, dynamic>> _recordValues =
      <int, Map<String, dynamic>>{};
  final Set<int> _savingItems = <int>{};
  final Set<int> _savedItems = <int>{};
  final ScrollController _leftScrollController = ScrollController();
  final ScrollController _questionScrollController = ScrollController();
  final GlobalKey _activeNavItemKey = GlobalKey();
  final Map<String, GlobalKey> _pageGroupKeys = <String, GlobalKey>{};
  final PadMessageOverlayController _messageController =
      PadMessageOverlayController();

  Pep3DraftDetail? _draft;
  Pep3DraftSummary? _detectedDraft;
  Pep3CaregiverInvite? _caregiverInvite;
  int _caregiverInviteDraftId = 0;
  Future<void>? _caregiverInviteRequest;
  HomeSession _session = HomeSession.fallback;
  final Map<int, int> _previousItemScores = <int, int>{};
  int _currentItemNo = 0;
  String _expandedGroupKey = '';
  String _errorMessage = '';
  String _autoSaveText = '';
  String _studentName = '';
  String _studentAge = '';
  String _birthDate = '';
  String _assessmentDate = '';
  String _examinerName = '';
  String _previousAssessmentDate = '';
  bool _draftDialogShown = false;
  bool _loading = true;
  bool _itemLoading = false;
  bool _savingDraft = false;
  bool _submitting = false;
  bool _caregiverLoading = false;
  bool _autoNext = true;

  @override
  void initState() {
    super.initState();
    _studentName = widget.args.studentName;
    _studentAge = widget.args.studentAge;
    _birthDate = _normalizeDate(widget.args.birthDate) ?? '';
    _assessmentDate =
        _normalizeDate(widget.args.assessmentDate) ?? _todayIsoDate();
    _examinerName = widget.args.examinerName;
    _initialize();
  }

  @override
  void dispose() {
    _messageController.dispose();
    _leftScrollController.dispose();
    _questionScrollController.dispose();
    super.dispose();
  }

  Future<String> _readToken() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    return prefs.getString(_authTokenStorageKey) ?? '';
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
    final Pep3DraftDetail detail =
        await widget.client.fetchDraftDetail(token, draftId);
    if (_caregiverInviteDraftId != detail.id) {
      _caregiverInvite = null;
      _caregiverInviteDraftId = 0;
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

  void _showDetectedDraftDialogIfNeeded() {
    final Pep3DraftSummary? draft = _detectedDraft;
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
          return PadDialogViewport(
            child: _DraftResumeDialog(
              draft: draft,
              total: _totalCount,
              onRestart: _restartWithoutDetectedDraft,
              onContinue: () => _continueDetectedDraft(draft),
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
      await _loadDraftDetail(token, draft.id);
      _currentItemNo = _resolveInitialItemNo();
      _expandedGroupKey = _groupKeyForItem(_currentItemNo);
      await _loadCurrentItem(token);
      await _loadPreviousAssessment(token);
      if (!mounted) {
        return false;
      }
      setState(() {
        _detectedDraft = null;
        _autoSaveText = '已恢复最新草稿';
      });
      _keepActiveItemVisible();
      await _refreshCaregiverInvite(silent: true);
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
    final int studentId = _studentId;
    if (studentId <= 0) {
      _previousAssessmentDate = '';
      _previousItemScores.clear();
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
        _previousAssessmentDate = '';
        _previousItemScores.clear();
        return;
      }
      final Pep3RecordDetail detail =
          await widget.client.fetchRecordDetail(token, latest.id);
      _previousAssessmentDate = _normalizeDate(detail.assessmentDate) ??
          _normalizeDate(latest.assessmentDate) ??
          '';
      _previousItemScores
        ..clear()
        ..addAll(detail.input.itemScores);
    } on Object {
      _previousAssessmentDate = '';
      _previousItemScores.clear();
    }
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
    final List<int> missing = _draft?.progress.missingItemNos ?? <int>[];
    if (missing.isNotEmpty &&
        items.any((Pep3ItemSummary item) => item.itemNo == missing.first)) {
      return missing.first;
    }
    return items.first.itemNo;
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
      _autoSaveText = '自动保存中...';
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

  @override
  Widget build(BuildContext context) {
    if (_loading) {
      return _Pep3LoadingShell(
        title: _scaleTitle,
        studentName: _studentName,
        age: _studentAgeText,
        assessmentDate: _assessmentDate,
        examinerName: _examinerName,
        onBack: widget.onBack,
      );
    }
    if (_errorMessage.isNotEmpty) {
      return _Pep3ErrorShell(message: _errorMessage, onBack: widget.onBack);
    }
    return ColoredBox(
      color: _Pep3Colors.page,
      child: Column(
        children: <Widget>[
          _Pep3Header(
            title: _scaleTitle,
            studentName: _studentName.trim().isEmpty ? '-' : _studentName,
            age: _studentAgeText,
            assessmentDate: _assessmentDate,
            examinerName: _examinerName.trim().isEmpty ? '当前老师' : _examinerName,
            autoSaveText: _autoSaveText,
            saving: _savingDraft,
            submitting: _submitting,
            onBack: widget.onBack,
            onSave: () => _saveDraft(),
            onSubmit: _submitDraft,
          ),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(10, 10, 10, 10),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  SizedBox(
                    width: 226,
                    child: _Pep3PageSidebar(
                      groups: _template.itemGroups,
                      expandedGroupKey: _expandedGroupKey,
                      itemScores: _itemScores,
                      currentItemNo: _currentItemNo,
                      controller: _leftScrollController,
                      activeItemKey: _activeNavItemKey,
                      groupKeys: _pageGroupKeys,
                      onToggleGroup: (String key) {
                        final bool opening = _expandedGroupKey != key;
                        setState(() {
                          _expandedGroupKey = opening ? key : '';
                        });
                        if (opening) {
                          _keepPageGroupVisible(key);
                        }
                      },
                      onTapItem: _goToItem,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: _Pep3QuestionPanel(
                      controller: _questionScrollController,
                      loading: _itemLoading,
                      item: _currentItem,
                      summary: _currentSummary,
                      scoreOptions: _currentScoreOptions,
                      selectedScore: _itemScores[_currentItemNo],
                      previousScore: _previousItemScores[_currentItemNo],
                      previousAssessmentDate: _previousAssessmentDate,
                      saving: _savingItems.contains(_currentItemNo),
                      saved: _savedItems.contains(_currentItemNo),
                      recordValues:
                          _recordValues[_currentItemNo] ?? <String, dynamic>{},
                      onScore: _setScore,
                      onRecordValue: _setRecordValue,
                    ),
                  ),
                  const SizedBox(width: 10),
                  SizedBox(
                    width: 238,
                    child: _Pep3RightRail(
                      progressPercent: _progressPercent,
                      answered: _answeredCount,
                      total: _totalCount,
                      missing: _missingCount,
                      currentItemNo: _currentItemNo,
                      recordFields:
                          _currentItem?.recordFields ?? <Pep3RecordField>[],
                      recordValues:
                          _recordValues[_currentItemNo] ?? <String, dynamic>{},
                      caregiverInvite: _caregiverInvite,
                      caregiverLoading: _caregiverLoading,
                      onRecordValue: _setRecordValueForItem,
                      onSmsTap: () => _showMessage('短信发送功能暂未开放'),
                      onWechatTap: () => _showMessage('微信推送功能暂未开放'),
                    ),
                  ),
                ],
              ),
            ),
          ),
          _Pep3Footer(
            current: _currentIndex + 1,
            total: _totalCount,
            hasPrevious: _hasPreviousItem,
            hasNext: _hasNextItem,
            autoNext: _autoNext,
            onPrevious: _goPrevious,
            onNext: _goNext,
            onJumpMissing: _jumpToMissing,
            onToggleAutoNext: (bool value) => setState(() => _autoNext = value),
          ),
        ],
      ),
    );
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

class _DraftResumeDialog extends StatefulWidget {
  const _DraftResumeDialog({
    required this.draft,
    required this.total,
    required this.onRestart,
    required this.onContinue,
  });

  final Pep3DraftSummary draft;
  final int total;
  final VoidCallback onRestart;
  final Future<bool> Function() onContinue;

  @override
  State<_DraftResumeDialog> createState() => _DraftResumeDialogState();
}

class _DraftResumeDialogState extends State<_DraftResumeDialog> {
  static const Duration _closeDuration = Duration(milliseconds: 260);

  bool _continuing = false;
  bool _closing = false;

  Future<void> _closeAfterShrink({VoidCallback? afterClosed}) async {
    if (_closing) {
      return;
    }
    setState(() => _closing = true);
    await Future<void>.delayed(_closeDuration);
    if (!mounted) {
      return;
    }
    Navigator.of(context).pop();
    afterClosed?.call();
  }

  Future<void> _handleRestart() async {
    if (_continuing || _closing) {
      return;
    }
    await _closeAfterShrink(afterClosed: widget.onRestart);
  }

  Future<void> _handleContinue() async {
    if (_continuing || _closing) {
      return;
    }
    setState(() => _continuing = true);
    await WidgetsBinding.instance.endOfFrame;
    if (!mounted || !_continuing) {
      return;
    }
    final bool restored = await widget.onContinue();
    if (!mounted) {
      return;
    }
    if (restored) {
      await _closeAfterShrink();
      return;
    }
    setState(() => _continuing = false);
  }

  @override
  Widget build(BuildContext context) {
    final int answered = widget.draft.progress.answeredItemCount > 0
        ? widget.draft.progress.answeredItemCount
        : widget.draft.answeredItemCount;
    final int resolvedTotal = widget.draft.progress.itemCount > 0
        ? widget.draft.progress.itemCount
        : math.max(widget.total, answered);
    return AnimatedOpacity(
      opacity: _closing ? 0 : 1,
      duration: _closeDuration,
      curve: Curves.easeInCubic,
      child: AnimatedScale(
        scale: _closing ? .92 : 1,
        duration: _closeDuration,
        curve: Curves.easeInCubic,
        child: Dialog(
          insetPadding: const EdgeInsets.symmetric(horizontal: 32),
          backgroundColor: Colors.transparent,
          child: Container(
            width: 520,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(14),
              boxShadow: _pep3Shadow(
                color: const Color(0x33000000),
                blur: 30,
                offset: const Offset(0, 18),
              ),
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                const Padding(
                  padding: EdgeInsets.fromLTRB(20, 20, 20, 20),
                  child: Text(
                    '发现未完成草稿',
                    style: TextStyle(
                      color: _Pep3Colors.ink,
                      fontSize: 19,
                      height: 1,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                const Divider(height: 1, color: _Pep3Colors.lineSoft),
                Padding(
                  padding: const EdgeInsets.fromLTRB(30, 30, 30, 30),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      const Text(
                        '当前儿童存在一份未提交的 PEP-3 测评草稿。',
                        style: TextStyle(
                          color: _Pep3Colors.ink,
                          fontSize: 15,
                          height: 1.2,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      const SizedBox(height: 22),
                      Container(
                        width: double.infinity,
                        padding: const EdgeInsets.fromLTRB(18, 16, 18, 16),
                        decoration: BoxDecoration(
                          color: const Color(0xFFFFFBF7),
                          borderRadius: BorderRadius.circular(10),
                          border: Border.all(color: _Pep3Colors.line),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: <Widget>[
                            _DraftResumeMeta(
                              label: '已完成',
                              value: '$answered / $resolvedTotal 题',
                            ),
                            const SizedBox(height: 13),
                            _DraftResumeMeta(
                              label: '更新时间',
                              value: _formatDateTime(widget.draft.updatedTime),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                const Divider(height: 1, color: _Pep3Colors.lineSoft),
                Padding(
                  padding: const EdgeInsets.fromLTRB(30, 18, 30, 20),
                  child: _DraftResumeActionArea(
                    continuing: _continuing,
                    onRestart: _handleRestart,
                    onContinue: _handleContinue,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _DraftResumeActionArea extends StatelessWidget {
  const _DraftResumeActionArea({
    required this.continuing,
    required this.onRestart,
    required this.onContinue,
  });

  final bool continuing;
  final VoidCallback onRestart;
  final VoidCallback onContinue;

  @override
  Widget build(BuildContext context) {
    return Align(
      alignment: Alignment.centerRight,
      child: SizedBox(
        width: 236,
        height: 42,
        child: AnimatedSwitcher(
          duration: const Duration(milliseconds: 180),
          reverseDuration: const Duration(milliseconds: 120),
          switchInCurve: Curves.easeOutCubic,
          switchOutCurve: Curves.easeOutCubic,
          layoutBuilder: (
            Widget? currentChild,
            List<Widget> previousChildren,
          ) {
            return Stack(
              alignment: Alignment.centerRight,
              children: <Widget>[
                ...previousChildren,
                if (currentChild != null) currentChild,
              ],
            );
          },
          transitionBuilder: (Widget child, Animation<double> animation) {
            return FadeTransition(opacity: animation, child: child);
          },
          child: continuing
              ? const _DialogLoadingButton(
                  key: ValueKey<String>('draft-resume-loading'),
                )
              : Row(
                  key: const ValueKey<String>('draft-resume-actions'),
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: <Widget>[
                    _DialogActionButton(
                      label: '重新测评',
                      filled: false,
                      onTap: onRestart,
                    ),
                    const SizedBox(width: 12),
                    _DialogActionButton(
                      label: '继续测评',
                      filled: true,
                      onTap: onContinue,
                    ),
                  ],
                ),
        ),
      ),
    );
  }
}

class _DialogLoadingButton extends StatelessWidget {
  const _DialogLoadingButton({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 236,
      height: 42,
      decoration: BoxDecoration(
        color: _Pep3Colors.orange,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _Pep3Colors.orange),
      ),
      child: const Center(
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            SizedBox(
              width: 15,
              height: 15,
              child: CircularProgressIndicator(
                strokeWidth: 2.2,
                color: Colors.white,
              ),
            ),
            SizedBox(width: 8),
            Text(
              '题目填充中，请稍后...',
              style: TextStyle(
                color: Colors.white,
                fontSize: 14,
                fontWeight: FontWeight.w900,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _DraftResumeMeta extends StatelessWidget {
  const _DraftResumeMeta({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Text.rich(
      TextSpan(
        children: <InlineSpan>[
          TextSpan(text: '$label：'),
          TextSpan(
            text: value,
            style: const TextStyle(
              color: _Pep3Colors.ink,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
      style: const TextStyle(
        color: _Pep3Colors.text,
        fontSize: 14,
        height: 1.2,
        fontWeight: FontWeight.w800,
      ),
    );
  }
}

class _DialogActionButton extends StatelessWidget {
  const _DialogActionButton({
    required this.label,
    required this.filled,
    required this.onTap,
  });

  final String label;
  final bool filled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          width: 112,
          height: 42,
          decoration: BoxDecoration(
            color: filled ? _Pep3Colors.orange : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: filled ? _Pep3Colors.orange : _Pep3Colors.line,
            ),
          ),
          child: Center(
            child: Text(
              label,
              style: TextStyle(
                color: filled ? Colors.white : _Pep3Colors.ink,
                fontSize: 14,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _Pep3Header extends StatelessWidget {
  const _Pep3Header({
    required this.title,
    required this.studentName,
    required this.age,
    required this.assessmentDate,
    required this.examinerName,
    required this.autoSaveText,
    required this.saving,
    required this.submitting,
    required this.onBack,
    required this.onSave,
    required this.onSubmit,
  });

  final String title;
  final String studentName;
  final String age;
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
    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 18),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        border: Border.all(color: _Pep3Colors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _pep3Shadow(color: const Color(0x16B05F32), blur: 16),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1120;
          return Row(
            children: <Widget>[
              _HeaderIconButton(
                  icon: Icons.chevron_left_rounded, onTap: onBack),
              const SizedBox(width: 10),
              SizedBox(
                width: compact ? 206 : 250,
                child: Text(
                  '$title 测评工作台',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _Pep3Colors.ink,
                    fontSize: 23,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Row(
                  children: <Widget>[
                    Expanded(
                        child: _HeaderMeta(label: '儿童', value: studentName)),
                    Expanded(child: _HeaderMeta(label: '年龄', value: age)),
                    Expanded(
                      flex: 2,
                      child: _HeaderMeta(label: '测评日期', value: assessmentDate),
                    ),
                    Expanded(
                      child: _HeaderMeta(label: '施测者', value: examinerName),
                    ),
                  ],
                ),
              ),
              if (autoSaveText.trim().isNotEmpty)
                SizedBox(
                  width: compact ? 82 : 112,
                  child: Text(
                    autoSaveText,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    textAlign: TextAlign.right,
                    style: const TextStyle(
                      color: _Pep3Colors.muted,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ),
              const SizedBox(width: 10),
              _TopActionButton(
                label: '保存草稿',
                icon: Icons.save_outlined,
                loading: saving,
                filled: false,
                onTap: onSave,
              ),
              const SizedBox(width: 9),
              _TopActionButton(
                label: '提交记录',
                icon: Icons.fact_check_outlined,
                loading: submitting,
                filled: true,
                onTap: onSubmit,
              ),
            ],
          );
        },
      ),
    );
  }
}

class _HeaderMeta extends StatelessWidget {
  const _HeaderMeta({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(left: 10),
      padding: const EdgeInsets.only(left: 10),
      decoration: const BoxDecoration(
        border: Border(left: BorderSide(color: _Pep3Colors.line)),
      ),
      child: FittedBox(
        fit: BoxFit.scaleDown,
        alignment: Alignment.centerLeft,
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
          style: const TextStyle(
            color: _Pep3Colors.text,
            fontSize: 13,
            height: 1,
            fontWeight: FontWeight.w700,
          ),
        ),
      ),
    );
  }
}

class _Pep3PageSidebar extends StatelessWidget {
  const _Pep3PageSidebar({
    required this.groups,
    required this.expandedGroupKey,
    required this.itemScores,
    required this.currentItemNo,
    required this.controller,
    required this.activeItemKey,
    required this.groupKeys,
    required this.onToggleGroup,
    required this.onTapItem,
  });

  final List<Pep3ItemGroupSummary> groups;
  final String expandedGroupKey;
  final Map<int, int> itemScores;
  final int currentItemNo;
  final ScrollController controller;
  final GlobalKey activeItemKey;
  final Map<String, GlobalKey> groupKeys;
  final ValueChanged<String> onToggleGroup;
  final ValueChanged<int> onTapItem;

  @override
  Widget build(BuildContext context) {
    return _RailCard(
      child: Column(
        children: <Widget>[
          const _SidebarHeader(),
          Expanded(
            child: SingleChildScrollView(
              controller: controller,
              physics: const BouncingScrollPhysics(),
              child: Column(
                children: <Widget>[
                  for (final Pep3ItemGroupSummary group in groups)
                    _buildGroup(group),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildGroup(Pep3ItemGroupSummary group) {
    final bool expanded = group.key == expandedGroupKey;
    final int done = group.items
        .where((Pep3ItemSummary item) => itemScores.containsKey(item.itemNo))
        .length;
    final int total = group.items.length;
    final int percent = total == 0 ? 0 : ((done / total) * 100).round();
    return _PageGroup(
      key: groupKeys.putIfAbsent(
        group.key,
        () => GlobalKey(debugLabel: 'pep3-page-group-${group.key}'),
      ),
      group: group,
      expanded: expanded,
      done: done,
      total: total,
      percent: percent,
      itemScores: itemScores,
      currentItemNo: currentItemNo,
      activeItemKey: activeItemKey,
      onToggle: () => onToggleGroup(group.key),
      onTapItem: onTapItem,
    );
  }
}

class _SidebarHeader extends StatelessWidget {
  const _SidebarHeader();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 48,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: _Pep3Colors.lineSoft)),
      ),
      child: const Row(
        children: <Widget>[
          Text(
            '记录册页面',
            style: TextStyle(
              color: _Pep3Colors.ink,
              fontSize: 16,
              fontWeight: FontWeight.w900,
            ),
          ),
          Spacer(),
          Icon(Icons.tune_rounded, size: 18, color: _Pep3Colors.muted),
        ],
      ),
    );
  }
}

class _PageGroup extends StatelessWidget {
  const _PageGroup({
    super.key,
    required this.group,
    required this.expanded,
    required this.done,
    required this.total,
    required this.percent,
    required this.itemScores,
    required this.currentItemNo,
    required this.activeItemKey,
    required this.onToggle,
    required this.onTapItem,
  });

  final Pep3ItemGroupSummary group;
  final bool expanded;
  final int done;
  final int total;
  final int percent;
  final Map<int, int> itemScores;
  final int currentItemNo;
  final GlobalKey activeItemKey;
  final VoidCallback onToggle;
  final ValueChanged<int> onTapItem;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 10, 9),
      decoration: const BoxDecoration(
        border: Border(bottom: BorderSide(color: _Pep3Colors.lineSoft)),
      ),
      child: Column(
        children: <Widget>[
          InkWell(
            onTap: onToggle,
            borderRadius: BorderRadius.circular(8),
            child: SizedBox(
              height: 26,
              child: Row(
                children: <Widget>[
                  Icon(
                    expanded
                        ? Icons.keyboard_arrow_down_rounded
                        : Icons.chevron_right_rounded,
                    size: 20,
                    color: _Pep3Colors.ink,
                  ),
                  const SizedBox(width: 4),
                  Expanded(
                    child: Text(
                      group.displayTitle,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _Pep3Colors.ink,
                        fontSize: 13,
                        height: 1,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  Text(
                    '$done/$total 题',
                    style: const TextStyle(
                      color: _Pep3Colors.muted,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 7),
          Row(
            children: <Widget>[
              const SizedBox(width: 26),
              Expanded(
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(99),
                  child: LinearProgressIndicator(
                    value: percent / 100,
                    minHeight: 5,
                    backgroundColor: const Color(0xFFF2E6DC),
                    valueColor:
                        const AlwaysStoppedAnimation<Color>(_Pep3Colors.orange),
                  ),
                ),
              ),
              const SizedBox(width: 9),
              SizedBox(
                width: 34,
                child: Text(
                  '$percent%',
                  textAlign: TextAlign.right,
                  style: const TextStyle(
                    color: _Pep3Colors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
            ],
          ),
          if (expanded) ...<Widget>[
            const SizedBox(height: 7),
            for (final Pep3ItemSummary item in group.items)
              _QuestionNavItem(
                key: item.itemNo == currentItemNo ? activeItemKey : null,
                item: item,
                active: item.itemNo == currentItemNo,
                done: itemScores.containsKey(item.itemNo),
                onTap: () => onTapItem(item.itemNo),
              ),
          ],
        ],
      ),
    );
  }
}

class _QuestionNavItem extends StatelessWidget {
  const _QuestionNavItem({
    super.key,
    required this.item,
    required this.active,
    required this.done,
    required this.onTap,
  });

  final Pep3ItemSummary item;
  final bool active;
  final bool done;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 18, top: 3),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(10),
          child: Ink(
            height: 31,
            padding: const EdgeInsets.symmetric(horizontal: 9),
            decoration: BoxDecoration(
              color: active ? const Color(0xFFFFEEE5) : Colors.transparent,
              borderRadius: BorderRadius.circular(10),
            ),
            child: Row(
              children: <Widget>[
                SizedBox(
                  width: 48,
                  child: Text(
                    '第 ${item.itemNo} 题',
                    maxLines: 1,
                    style: TextStyle(
                      color: active ? _Pep3Colors.orangeDeep : _Pep3Colors.text,
                      fontSize: 12,
                      fontWeight: active ? FontWeight.w900 : FontWeight.w700,
                    ),
                  ),
                ),
                Expanded(
                  child: Text(
                    item.displayTitle,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: active ? _Pep3Colors.orangeDeep : _Pep3Colors.text,
                      fontSize: 12,
                      fontWeight: active ? FontWeight.w900 : FontWeight.w700,
                    ),
                  ),
                ),
                const SizedBox(width: 7),
                if (done)
                  const Icon(Icons.check_circle_rounded,
                      size: 17, color: _Pep3Colors.green)
                else
                  Container(
                    width: 15,
                    height: 15,
                    decoration: BoxDecoration(
                      color: active ? _Pep3Colors.orange : Colors.white,
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: active
                            ? _Pep3Colors.orange
                            : const Color(0xFFCAB8AA),
                      ),
                    ),
                    child: active
                        ? const Center(
                            child: DecoratedBox(
                              decoration: BoxDecoration(
                                color: Colors.white,
                                shape: BoxShape.circle,
                              ),
                              child: SizedBox(width: 5, height: 5),
                            ),
                          )
                        : null,
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _Pep3QuestionPanel extends StatelessWidget {
  const _Pep3QuestionPanel({
    required this.controller,
    required this.loading,
    required this.item,
    required this.summary,
    required this.scoreOptions,
    required this.selectedScore,
    required this.previousScore,
    required this.previousAssessmentDate,
    required this.saving,
    required this.saved,
    required this.recordValues,
    required this.onScore,
    required this.onRecordValue,
  });

  final ScrollController controller;
  final bool loading;
  final Pep3AssessmentItem? item;
  final Pep3ItemSummary? summary;
  final List<Pep3ScoreOption> scoreOptions;
  final int? selectedScore;
  final int? previousScore;
  final String previousAssessmentDate;
  final bool saving;
  final bool saved;
  final Map<String, dynamic> recordValues;
  final ValueChanged<int> onScore;
  final void Function(String key, dynamic value) onRecordValue;

  @override
  Widget build(BuildContext context) {
    final Pep3AssessmentItem resolved = item ?? Pep3AssessmentItem.empty;
    final Pep3ItemSummary? resolvedSummary = summary;
    return _RailCard(
      child: loading && item == null
          ? _QuestionLoadingView(
              summary: resolvedSummary,
              scoreOptions: scoreOptions,
            )
          : Padding(
              padding: const EdgeInsets.fromLTRB(16, 18, 16, 14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  Row(
                    children: <Widget>[
                      Expanded(
                        child: Text(
                          '第 ${resolvedSummary?.itemNo ?? resolved.itemNo} 题  ${resolvedSummary?.displayTitle ?? resolved.displayTitle}',
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            color: _Pep3Colors.ink,
                            fontSize: 23,
                            height: 1,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ),
                      const SizedBox(width: 10),
                      _DomainChip(
                        code: resolved.domainCode.isNotEmpty
                            ? resolved.domainCode
                            : resolvedSummary?.domainCode ?? '',
                        name: resolved.domainName.isNotEmpty
                            ? resolved.domainName
                            : resolvedSummary?.domainName ?? '',
                      ),
                      if (saving || saved) ...<Widget>[
                        const SizedBox(width: 8),
                        _SaveBadge(saving: saving),
                      ],
                    ],
                  ),
                  const SizedBox(height: 16),
                  Expanded(
                    child: ListView(
                      key: const ValueKey<String>(
                        'pep3-question-instruction-scroll',
                      ),
                      controller: controller,
                      padding: EdgeInsets.zero,
                      physics: const BouncingScrollPhysics(),
                      children: <Widget>[
                        _InstructionCard(
                          title: '材料',
                          icon: Icons.article_outlined,
                          body: _normalizeText(resolved.materials),
                        ),
                        _InstructionCard(
                          title: '操作标准',
                          icon: Icons.assignment_outlined,
                          body: _normalizeText(resolved.method),
                        ),
                        _InstructionCard(
                          title: '指导语',
                          icon: Icons.record_voice_over_outlined,
                          body: _normalizeText(resolved.guidance),
                        ),
                        _InstructionCard(
                          title: '评分标准',
                          icon: Icons.fact_check_outlined,
                          body: _scoreStandardText(resolved, scoreOptions),
                        ),
                        const SizedBox(height: 2),
                      ],
                    ),
                  ),
                  const SizedBox(height: 10),
                  _ScoreDock(
                    scoreOptions: scoreOptions,
                    selectedScore: selectedScore,
                    previousScore: previousScore,
                    previousAssessmentDate: previousAssessmentDate,
                    onScore: onScore,
                  ),
                ],
              ),
            ),
    );
  }
}

class _QuestionLoadingView extends StatelessWidget {
  const _QuestionLoadingView({
    required this.summary,
    required this.scoreOptions,
  });

  final Pep3ItemSummary? summary;
  final List<Pep3ScoreOption> scoreOptions;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 18, 16, 14),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Row(
            children: <Widget>[
              Expanded(
                child: Text(
                  '第 ${summary?.itemNo ?? 0} 题  ${summary?.displayTitle ?? ''}',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _Pep3Colors.ink,
                    fontSize: 23,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 10),
              _DomainChip(
                code: summary?.domainCode ?? '',
                name: summary?.domainName ?? '',
              ),
            ],
          ),
          const SizedBox(height: 16),
          Expanded(
            child: ListView(
              padding: EdgeInsets.zero,
              physics: const NeverScrollableScrollPhysics(),
              children: const <Widget>[
                _QuestionLoadingCard(title: '材料', height: 72),
                _QuestionLoadingCard(title: '操作标准', height: 86),
                _QuestionLoadingCard(title: '指导语', height: 86),
                _QuestionLoadingCard(title: '评分标准', height: 86),
                SizedBox(height: 2),
              ],
            ),
          ),
          const SizedBox(height: 10),
          _ScoreLoadingDock(scoreOptions: scoreOptions),
        ],
      ),
    );
  }
}

class _ScoreDock extends StatelessWidget {
  const _ScoreDock({
    required this.scoreOptions,
    required this.selectedScore,
    required this.previousScore,
    required this.previousAssessmentDate,
    required this.onScore,
  });

  final List<Pep3ScoreOption> scoreOptions;
  final int? selectedScore;
  final int? previousScore;
  final String previousAssessmentDate;
  final ValueChanged<int> onScore;

  @override
  Widget build(BuildContext context) {
    final bool hasPrevious =
        previousScore != null && previousAssessmentDate.trim().isNotEmpty;
    return Column(
      key: const ValueKey<String>('pep3-question-score-dock'),
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Row(
          children: <Widget>[
            const Text(
              '评分',
              style: TextStyle(
                color: _Pep3Colors.ink,
                fontSize: 15,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            const Spacer(),
            if (hasPrevious)
              _PreviousScoreSummary(
                score: previousScore!,
                date: previousAssessmentDate,
              ),
          ],
        ),
        const SizedBox(height: 10),
        Row(
          children: <Widget>[
            for (int i = 0; i < scoreOptions.length; i++) ...<Widget>[
              Expanded(
                child: _ScoreOptionCard(
                  option: scoreOptions[i],
                  selected: selectedScore == scoreOptions[i].value,
                  previous:
                      hasPrevious && previousScore == scoreOptions[i].value,
                  previousDate: previousAssessmentDate,
                  onTap: () => onScore(scoreOptions[i].value),
                ),
              ),
              if (i != scoreOptions.length - 1) const SizedBox(width: 14),
            ],
          ],
        ),
      ],
    );
  }
}

class _ScoreLoadingDock extends StatelessWidget {
  const _ScoreLoadingDock({required this.scoreOptions});

  final List<Pep3ScoreOption> scoreOptions;

  @override
  Widget build(BuildContext context) {
    final int optionCount = math.max(scoreOptions.length, 3);
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Row(
          children: <Widget>[
            const Text(
              '评分',
              style: TextStyle(
                color: _Pep3Colors.ink,
                fontSize: 15,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            const Spacer(),
            Container(
              height: 28,
              padding: const EdgeInsets.symmetric(horizontal: 12),
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: const Color(0xFFFFF2EA),
                borderRadius: BorderRadius.circular(9),
                border: Border.all(color: const Color(0xFFFFD6C3)),
              ),
              child: const Row(
                mainAxisSize: MainAxisSize.min,
                children: <Widget>[
                  SizedBox(
                    width: 13,
                    height: 13,
                    child: CircularProgressIndicator(
                      strokeWidth: 2,
                      color: _Pep3Colors.orange,
                    ),
                  ),
                  SizedBox(width: 8),
                  Text(
                    '题目加载中',
                    style: TextStyle(
                      color: _Pep3Colors.orangeDeep,
                      fontSize: 12,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        Row(
          children: <Widget>[
            for (int i = 0; i < optionCount; i++) ...<Widget>[
              const Expanded(child: _ScoreLoadingCard()),
              if (i != optionCount - 1) const SizedBox(width: 14),
            ],
          ],
        ),
      ],
    );
  }
}

class _QuestionLoadingCard extends StatelessWidget {
  const _QuestionLoadingCard({required this.title, required this.height});

  final String title;
  final double height;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      constraints: BoxConstraints(minHeight: height),
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.94),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _Pep3Colors.line),
        boxShadow: _pep3Shadow(
          color: const Color(0x0FB05F32),
          blur: 12,
          offset: const Offset(0, 6),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Row(
            children: <Widget>[
              Container(
                width: 17,
                height: 17,
                decoration: BoxDecoration(
                  color: const Color(0xFFFFEFE6),
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: const Color(0xFFFFCFB6)),
                ),
              ),
              const SizedBox(width: 7),
              Text(
                title,
                style: const TextStyle(
                  color: _Pep3Colors.ink,
                  fontSize: 16,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          const _LoadingLine(widthFactor: .72),
          const SizedBox(height: 8),
          const _LoadingLine(widthFactor: .92),
        ],
      ),
    );
  }
}

class _ScoreLoadingCard extends StatelessWidget {
  const _ScoreLoadingCard();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 94,
      padding: const EdgeInsets.fromLTRB(16, 15, 14, 14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _Pep3Colors.line),
      ),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: <Widget>[
          _LoadingLine(width: 54, height: 20),
          SizedBox(height: 12),
          _LoadingLine(width: 72, height: 12),
        ],
      ),
    );
  }
}

class _LoadingLine extends StatelessWidget {
  const _LoadingLine({
    this.width,
    this.widthFactor,
    this.height = 10,
  });

  final double? width;
  final double? widthFactor;
  final double height;

  @override
  Widget build(BuildContext context) {
    final Widget line = Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: const Color(0xFFF3E8DF),
        borderRadius: BorderRadius.circular(99),
      ),
    );
    if (widthFactor == null) {
      return line;
    }
    return FractionallySizedBox(
      alignment: Alignment.centerLeft,
      widthFactor: widthFactor,
      child: line,
    );
  }
}

class _InstructionCard extends StatelessWidget {
  const _InstructionCard({
    required this.title,
    required this.icon,
    required this.body,
  });

  final String title;
  final IconData icon;
  final String body;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.94),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _Pep3Colors.line),
        boxShadow: _pep3Shadow(
          color: const Color(0x0FB05F32),
          blur: 12,
          offset: const Offset(0, 6),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Container(
                width: 17,
                height: 17,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: const Color(0xFFFFEFE6),
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: const Color(0xFFFFCFB6)),
                ),
                child: Icon(icon, size: 12, color: _Pep3Colors.orange),
              ),
              const SizedBox(width: 7),
              Text(
                title,
                style: const TextStyle(
                  color: _Pep3Colors.ink,
                  fontSize: 16,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 11),
          Text(
            body,
            style: const TextStyle(
              color: _Pep3Colors.text,
              fontSize: 14,
              height: 1.55,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}

class _ScoreOptionCard extends StatelessWidget {
  const _ScoreOptionCard({
    required this.option,
    required this.selected,
    required this.previous,
    required this.previousDate,
    required this.onTap,
  });

  final Pep3ScoreOption option;
  final bool selected;
  final bool previous;
  final String previousDate;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color color = _scoreColor(option.value);
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 94,
          padding: const EdgeInsets.fromLTRB(16, 9, 14, 9),
          decoration: BoxDecoration(
            color: selected ? color.withOpacity(.08) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: selected ? color : _Pep3Colors.line,
              width: selected ? 1.5 : 1,
            ),
          ),
          child: Stack(
            fit: StackFit.expand,
            clipBehavior: Clip.none,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: <Widget>[
                        Text(
                          '${option.value} 分',
                          style: TextStyle(
                            color: color,
                            fontSize: 24,
                            height: 1,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          _shortScoreLabel(option.value, option.label),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            color: _Pep3Colors.text,
                            fontSize: 13,
                            height: 1,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 36),
                  Icon(
                    selected
                        ? Icons.check_circle_rounded
                        : Icons.radio_button_unchecked_rounded,
                    color: selected ? color : const Color(0xFFCAB8AA),
                    size: 23,
                  ),
                ],
              ),
              if (previous)
                Positioned(
                  top: -5,
                  right: -8,
                  child: _PreviousScoreBadge(date: previousDate, color: color),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _PreviousScoreSummary extends StatelessWidget {
  const _PreviousScoreSummary({required this.score, required this.date});

  final int score;
  final String date;

  @override
  Widget build(BuildContext context) {
    final Color color = _scoreColor(score);
    return Container(
      height: 32,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: color.withOpacity(.06),
        borderRadius: BorderRadius.circular(9),
        border: Border.all(color: color.withOpacity(.38)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Container(
            width: 7,
            height: 7,
            decoration: BoxDecoration(color: color, shape: BoxShape.circle),
          ),
          const SizedBox(width: 9),
          Text(
            '上次测评 ${_compactDateLabel(date)}',
            style: const TextStyle(
              color: _Pep3Colors.text,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(width: 10),
          Text(
            '$score 分 · ${_shortScoreLabel(score, '')}',
            style: TextStyle(
              color: color,
              fontSize: 14,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _PreviousScoreBadge extends StatelessWidget {
  const _PreviousScoreBadge({required this.date, required this.color});

  final String date;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(9),
        border: Border.all(color: color.withOpacity(.62), width: 1.1),
      ),
      child: Text(
        '上次 ${_shortDateLabel(date)}',
        maxLines: 1,
        style: TextStyle(
          color: color,
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _Pep3RightRail extends StatelessWidget {
  const _Pep3RightRail({
    required this.progressPercent,
    required this.answered,
    required this.total,
    required this.missing,
    required this.currentItemNo,
    required this.recordFields,
    required this.recordValues,
    required this.caregiverInvite,
    required this.caregiverLoading,
    required this.onRecordValue,
    required this.onSmsTap,
    required this.onWechatTap,
  });

  final int progressPercent;
  final int answered;
  final int total;
  final int missing;
  final int currentItemNo;
  final List<Pep3RecordField> recordFields;
  final Map<String, dynamic> recordValues;
  final Pep3CaregiverInvite? caregiverInvite;
  final bool caregiverLoading;
  final void Function(int itemNo, String key, dynamic value) onRecordValue;
  final VoidCallback onSmsTap;
  final VoidCallback onWechatTap;

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: EdgeInsets.zero,
      physics: const BouncingScrollPhysics(),
      children: <Widget>[
        _RailCard(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: _ProgressPanel(
              progressPercent: progressPercent,
              answered: answered,
              total: total,
              missing: missing,
            ),
          ),
        ),
        const SizedBox(height: 10),
        _RailCard(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: _TrainingRecordPanel(
              currentItemNo: currentItemNo,
              fields: recordFields,
              values: recordValues,
              onChanged: (String key, dynamic value) =>
                  onRecordValue(currentItemNo, key, value),
            ),
          ),
        ),
        const SizedBox(height: 10),
        _RailCard(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: _CaregiverPanel(
              invite: caregiverInvite,
              loading: caregiverLoading,
              onSmsTap: onSmsTap,
              onWechatTap: onWechatTap,
            ),
          ),
        ),
      ],
    );
  }
}

class _ProgressPanel extends StatelessWidget {
  const _ProgressPanel({
    required this.progressPercent,
    required this.answered,
    required this.total,
    required this.missing,
  });

  final int progressPercent;
  final int answered;
  final int total;
  final int missing;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        const Text(
          '当前进度',
          style: TextStyle(
            color: _Pep3Colors.ink,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 14),
        Row(
          children: <Widget>[
            SizedBox(
              width: 82,
              height: 82,
              child: CustomPaint(
                painter: _DonutPainter(percent: progressPercent),
                child: Center(
                  child: Text(
                    '$progressPercent%',
                    style: const TextStyle(
                      color: _Pep3Colors.ink,
                      fontSize: 21,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  _ProgressText(label: '已完成', value: '$answered / $total 题'),
                  const SizedBox(height: 10),
                  _ProgressText(
                    label: '缺题',
                    value: '$missing 题',
                    danger: true,
                  ),
                ],
              ),
            ),
          ],
        ),
      ],
    );
  }
}

class _TrainingRecordPanel extends StatelessWidget {
  const _TrainingRecordPanel({
    required this.currentItemNo,
    required this.fields,
    required this.values,
    required this.onChanged,
  });

  final int currentItemNo;
  final List<Pep3RecordField> fields;
  final Map<String, dynamic> values;
  final void Function(String key, dynamic value) onChanged;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        const Text(
          '儿童训练记录',
          style: TextStyle(
            color: _Pep3Colors.ink,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 14),
        if (fields.isEmpty)
          Container(
            height: 50,
            alignment: Alignment.centerLeft,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            decoration: BoxDecoration(
              color: const Color(0xFFFFFAF5),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: _Pep3Colors.lineSoft),
            ),
            child: const Text(
              '本题暂无训练记录项',
              style: TextStyle(
                color: _Pep3Colors.muted,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          )
        else
          for (final Pep3RecordField field in fields)
            _RecordFieldEditor(
              key: ValueKey<String>(
                'pep3-record-field-$currentItemNo-${field.key}',
              ),
              currentItemNo: currentItemNo,
              field: field,
              value: values[field.key],
              onChanged: (dynamic value) => onChanged(field.key, value),
            ),
      ],
    );
  }
}

class _CaregiverPanel extends StatelessWidget {
  const _CaregiverPanel({
    required this.invite,
    required this.loading,
    required this.onSmsTap,
    required this.onWechatTap,
  });

  final Pep3CaregiverInvite? invite;
  final bool loading;
  final VoidCallback onSmsTap;
  final VoidCallback onWechatTap;

  @override
  Widget build(BuildContext context) {
    final String qrValue = invite?.qrValue ?? '';
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        const Text(
          '照护者报告',
          style: TextStyle(
            color: _Pep3Colors.ink,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 14),
        Center(
          child: Container(
            width: 132,
            height: 132,
            alignment: Alignment.center,
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: _Pep3Colors.line),
            ),
            child: loading
                ? const CircularProgressIndicator(color: _Pep3Colors.orange)
                : _CaregiverQr(invite: invite, qrValue: qrValue),
          ),
        ),
        const SizedBox(height: 10),
        const Center(
          child: Text(
            '家长扫码填写照护者报告',
            style: TextStyle(
              color: _Pep3Colors.muted,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
        const SizedBox(height: 14),
        _CaregiverActionButton(
          label: '发送短信给家长',
          icon: Icons.sms_outlined,
          filled: false,
          loading: loading,
          onTap: onSmsTap,
        ),
        const SizedBox(height: 8),
        _CaregiverActionButton(
          label: '推送微信消息',
          icon: Icons.wechat,
          filled: true,
          loading: loading,
          onTap: onWechatTap,
        ),
      ],
    );
  }
}

class _CaregiverQr extends StatefulWidget {
  const _CaregiverQr({required this.invite, required this.qrValue});

  final Pep3CaregiverInvite? invite;
  final String qrValue;

  @override
  State<_CaregiverQr> createState() => _CaregiverQrState();
}

class _CaregiverQrState extends State<_CaregiverQr> {
  String _signature = '';
  Widget _cachedQr = const SizedBox.shrink();

  @override
  void initState() {
    super.initState();
    _cacheQr();
  }

  @override
  void didUpdateWidget(covariant _CaregiverQr oldWidget) {
    super.didUpdateWidget(oldWidget);
    final String nextSignature = _qrSignature(widget.invite, widget.qrValue);
    if (nextSignature != _signature) {
      _cacheQr();
    }
  }

  void _cacheQr() {
    final String signature = _qrSignature(widget.invite, widget.qrValue);
    _signature = signature;
    _cachedQr = _buildQr(widget.invite, widget.qrValue, signature);
  }

  @override
  Widget build(BuildContext context) {
    return RepaintBoundary(child: _cachedQr);
  }

  static String _qrSignature(Pep3CaregiverInvite? invite, String qrValue) {
    final String dataUrl = invite?.miniProgramCodeDataUrl.trim() ?? '';
    if (dataUrl.isNotEmpty) {
      return 'data:$dataUrl';
    }
    if (qrValue.trim().isNotEmpty) {
      return 'qr:${qrValue.trim()}';
    }
    return 'empty';
  }

  static Widget _buildQr(
    Pep3CaregiverInvite? invite,
    String qrValue,
    String signature,
  ) {
    final String dataUrl = invite?.miniProgramCodeDataUrl ?? '';
    final RegExpMatch? match =
        RegExp(r'^data:image/[^;]+;base64,(.+)$').firstMatch(dataUrl);
    if (match != null) {
      return Image.memory(
        base64Decode(match.group(1)!),
        key: ValueKey<String>('caregiver-image-$signature'),
        width: 116,
        height: 116,
        fit: BoxFit.contain,
        gaplessPlayback: true,
      );
    }
    if (qrValue.isNotEmpty) {
      return QrImageView(
        key: ValueKey<String>('caregiver-qr-$signature'),
        data: qrValue,
        version: QrVersions.auto,
        padding: EdgeInsets.zero,
        backgroundColor: Colors.white,
      );
    }
    return const Text(
      '暂无二维码',
      textAlign: TextAlign.center,
      style: TextStyle(
        color: _Pep3Colors.muted,
        fontSize: 13,
        fontWeight: FontWeight.w800,
      ),
    );
  }
}

class _Pep3Footer extends StatelessWidget {
  const _Pep3Footer({
    required this.current,
    required this.total,
    required this.hasPrevious,
    required this.hasNext,
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
        border: Border.all(color: _Pep3Colors.line),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
        boxShadow: _pep3Shadow(color: const Color(0x14B05F32), blur: 16),
      ),
      child: Row(
        children: <Widget>[
          _FooterButton(
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
                    color: _Pep3Colors.ink,
                    fontSize: 27,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                TextSpan(
                  text: ' / $total',
                  style: const TextStyle(
                    color: _Pep3Colors.text,
                    fontSize: 15,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          const Spacer(),
          _FooterButton(
            label: '下一题',
            icon: Icons.chevron_right_rounded,
            enabled: hasNext,
            filled: true,
            reverseIcon: true,
            onTap: onNext,
          ),
          const SizedBox(width: 14),
          _FooterButton(
            label: '跳到缺题',
            icon: Icons.swipe_right_alt_rounded,
            enabled: true,
            onTap: onJumpMissing,
          ),
          const SizedBox(width: 22),
          const Text(
            '自动下一题',
            style: TextStyle(
              color: _Pep3Colors.text,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(width: 8),
          Switch(
            value: autoNext,
            activeColor: _Pep3Colors.orange,
            onChanged: onToggleAutoNext,
          ),
        ],
      ),
    );
  }
}

class _RailCard extends StatelessWidget {
  const _RailCard({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.9),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _Pep3Colors.line),
        boxShadow: _pep3Shadow(
          color: const Color(0x12B05F32),
          blur: 15,
          offset: const Offset(0, 7),
        ),
      ),
      child: child,
    );
  }
}

class _RecordFieldEditor extends StatelessWidget {
  const _RecordFieldEditor({
    super.key,
    required this.currentItemNo,
    required this.field,
    required this.value,
    required this.onChanged,
  });

  final int currentItemNo;
  final Pep3RecordField field;
  final dynamic value;
  final ValueChanged<dynamic> onChanged;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _Pep3Colors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            field.label.trim().isEmpty ? field.key : field.label,
            style: const TextStyle(
              color: _Pep3Colors.text,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 8),
          if (field.fieldType == 'radio')
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: <Widget>[
                for (final Pep3RecordFieldOption option in field.options)
                  ChoiceChip(
                    label: Text(option.label),
                    selected: '$value' == option.value,
                    selectedColor: const Color(0xFFFFEEE5),
                    onSelected: (_) => onChanged(option.value),
                  ),
              ],
            )
          else if (field.fieldType == 'checkbox_group')
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: <Widget>[
                for (final Pep3RecordFieldOption option in field.options)
                  FilterChip(
                    label: Text(option.label),
                    selected:
                        value is List && (value as List).contains(option.value),
                    selectedColor: const Color(0xFFFFEEE5),
                    onSelected: (bool selected) {
                      final List<String> next = value is List
                          ? (value as List)
                              .map((dynamic item) => '$item')
                              .toList()
                          : <String>[];
                      if (selected) {
                        next.add(option.value);
                      } else {
                        next.remove(option.value);
                      }
                      onChanged(next);
                    },
                  ),
              ],
            )
          else
            SizedBox(
              height: field.fieldType == 'textarea' ? null : 52,
              child: TextFormField(
                key: ValueKey<String>(
                  'pep3-record-input-$currentItemNo-${field.key}',
                ),
                initialValue: value == null ? '' : '$value',
                minLines: field.fieldType == 'textarea' ? 2 : 1,
                maxLines: field.fieldType == 'textarea' ? 4 : 1,
                keyboardType:
                    field.fieldType == 'number' ? TextInputType.number : null,
                textAlignVertical: field.fieldType == 'textarea'
                    ? TextAlignVertical.top
                    : TextAlignVertical.center,
                style: const TextStyle(
                  color: _Pep3Colors.ink,
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                ),
                decoration: InputDecoration(
                  isDense: true,
                  contentPadding: EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: field.fieldType == 'textarea' ? 10 : 14,
                  ),
                  hintText: field.placeholder.trim().isEmpty
                      ? '请输入'
                      : field.placeholder,
                  hintStyle: const TextStyle(
                    color: _Pep3Colors.muted,
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                  ),
                  filled: true,
                  fillColor: Colors.white,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: const BorderSide(color: _Pep3Colors.line),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: const BorderSide(color: _Pep3Colors.line),
                  ),
                  focusedBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: const BorderSide(color: _Pep3Colors.orange),
                  ),
                ),
                onChanged: (String text) => onChanged(
                  field.fieldType == 'number'
                      ? num.tryParse(text.trim())
                      : text,
                ),
              ),
            ),
        ],
      ),
    );
  }
}

class _TopActionButton extends StatelessWidget {
  const _TopActionButton({
    required this.label,
    required this.icon,
    required this.loading,
    required this.filled,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final bool loading;
  final bool filled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: loading ? null : onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 36,
          padding: const EdgeInsets.symmetric(horizontal: 13),
          decoration: BoxDecoration(
            color: filled ? _Pep3Colors.orange : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _Pep3Colors.orange),
            boxShadow: filled
                ? _pep3Shadow(
                    color: const Color(0x28E96F43),
                    blur: 12,
                    offset: const Offset(0, 5),
                  )
                : null,
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              if (loading)
                SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    color: filled ? Colors.white : _Pep3Colors.orange,
                  ),
                )
              else
                Icon(
                  icon,
                  size: 17,
                  color: filled ? Colors.white : _Pep3Colors.orange,
                ),
              const SizedBox(width: 7),
              Text(
                label,
                style: TextStyle(
                  color: filled ? Colors.white : _Pep3Colors.orangeDeep,
                  fontSize: 14,
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

class _FooterButton extends StatelessWidget {
  const _FooterButton({
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
    final Color textColor = filled ? Colors.white : _Pep3Colors.orangeDeep;
    final List<Widget> children = <Widget>[
      Icon(icon, size: 22, color: textColor),
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
            color: filled
                ? _Pep3Colors.orange
                : enabled
                    ? Colors.white
                    : const Color(0xFFF7F1ED),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: enabled ? _Pep3Colors.orange : const Color(0xFFE2D6CE),
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

class _HeaderIconButton extends StatelessWidget {
  const _HeaderIconButton({required this.icon, required this.onTap});

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
            border: Border.all(color: _Pep3Colors.line),
          ),
          child: Icon(icon, color: _Pep3Colors.text, size: 34),
        ),
      ),
    );
  }
}

class _DomainChip extends StatelessWidget {
  const _DomainChip({required this.code, required this.name});

  final String code;
  final String name;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 11),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF1E8),
        borderRadius: BorderRadius.circular(9),
        border: Border.all(color: const Color(0xFFFFC8AD)),
      ),
      child: Text(
        '${code.trim()} ${name.trim()}'.trim(),
        style: const TextStyle(
          color: _Pep3Colors.orangeDeep,
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _SaveBadge extends StatelessWidget {
  const _SaveBadge({required this.saving});

  final bool saving;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 26,
      padding: const EdgeInsets.symmetric(horizontal: 9),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: saving ? const Color(0xFFFFF7EA) : const Color(0xFFEAF4E5),
        borderRadius: BorderRadius.circular(13),
      ),
      child: Text(
        saving ? '保存中' : '已保存',
        style: TextStyle(
          color: saving ? _Pep3Colors.orangeDeep : _Pep3Colors.green,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _CaregiverActionButton extends StatelessWidget {
  const _CaregiverActionButton({
    required this.label,
    required this.icon,
    required this.filled,
    required this.loading,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final bool filled;
  final bool loading;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: loading ? null : onTap,
        borderRadius: BorderRadius.circular(9),
        child: Ink(
          height: 34,
          decoration: BoxDecoration(
            color: filled ? _Pep3Colors.orange : Colors.white,
            borderRadius: BorderRadius.circular(9),
            border: Border.all(
                color: filled ? _Pep3Colors.orange : _Pep3Colors.line),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              Icon(icon,
                  size: 17, color: filled ? Colors.white : _Pep3Colors.ink),
              const SizedBox(width: 7),
              Text(
                label,
                style: TextStyle(
                  color: filled ? Colors.white : _Pep3Colors.ink,
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

class _ProgressText extends StatelessWidget {
  const _ProgressText({
    required this.label,
    required this.value,
    this.danger = false,
  });

  final String label;
  final String value;
  final bool danger;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          label,
          style: const TextStyle(
            color: _Pep3Colors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w800,
          ),
        ),
        const SizedBox(height: 5),
        Text(
          value,
          style: TextStyle(
            color: danger ? const Color(0xFFE04438) : _Pep3Colors.orangeDeep,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}

class _DonutPainter extends CustomPainter {
  const _DonutPainter({required this.percent});

  final int percent;

  @override
  void paint(Canvas canvas, Size size) {
    final Offset center = Offset(size.width / 2, size.height / 2);
    final double radius = math.min(size.width, size.height) / 2 - 6;
    final Paint track = Paint()
      ..color = const Color(0xFFE9DDD3)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 10
      ..strokeCap = StrokeCap.round;
    final Paint progress = Paint()
      ..color = _Pep3Colors.orange
      ..style = PaintingStyle.stroke
      ..strokeWidth = 10
      ..strokeCap = StrokeCap.round;
    canvas.drawCircle(center, radius, track);
    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      -math.pi / 2,
      math.pi * 2 * percent.clamp(0, 100) / 100,
      false,
      progress,
    );
  }

  @override
  bool shouldRepaint(covariant _DonutPainter oldDelegate) {
    return oldDelegate.percent != percent;
  }
}

class _Pep3LoadingShell extends StatelessWidget {
  const _Pep3LoadingShell({
    required this.title,
    required this.studentName,
    required this.age,
    required this.assessmentDate,
    required this.examinerName,
    required this.onBack,
  });

  final String title;
  final String studentName;
  final String age;
  final String assessmentDate;
  final String examinerName;
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: _Pep3Colors.page,
      child: Column(
        children: <Widget>[
          _Pep3LoadingHeader(
            title: title,
            studentName: studentName,
            age: age,
            assessmentDate: assessmentDate,
            examinerName: examinerName,
            onBack: onBack,
          ),
          const Expanded(
            child: Padding(
              padding: EdgeInsets.fromLTRB(10, 10, 10, 10),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  SizedBox(width: 226, child: _Pep3SidebarSkeleton()),
                  SizedBox(width: 10),
                  Expanded(child: _Pep3QuestionSkeleton()),
                  SizedBox(width: 10),
                  SizedBox(width: 238, child: _Pep3RightRailSkeleton()),
                ],
              ),
            ),
          ),
          const _Pep3FooterSkeleton(),
        ],
      ),
    );
  }
}

class _Pep3LoadingHeader extends StatelessWidget {
  const _Pep3LoadingHeader({
    required this.title,
    required this.studentName,
    required this.age,
    required this.assessmentDate,
    required this.examinerName,
    required this.onBack,
  });

  final String title;
  final String studentName;
  final String age;
  final String assessmentDate;
  final String examinerName;
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 18),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        border: Border.all(color: _Pep3Colors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _pep3Shadow(color: const Color(0x16B05F32), blur: 16),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1120;
          return Row(
            children: <Widget>[
              _HeaderIconButton(
                icon: Icons.chevron_left_rounded,
                onTap: onBack,
              ),
              const SizedBox(width: 10),
              SizedBox(
                width: compact ? 206 : 250,
                child: Text(
                  '$title 测评工作台',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _Pep3Colors.ink,
                    fontSize: 23,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Row(
                  children: <Widget>[
                    Expanded(
                      child: _HeaderLoadingMeta(
                        label: '儿童',
                        value: studentName,
                      ),
                    ),
                    Expanded(
                      child: _HeaderLoadingMeta(label: '年龄', value: age),
                    ),
                    Expanded(
                      flex: 2,
                      child: _HeaderLoadingMeta(
                        label: '测评日期',
                        value: assessmentDate,
                      ),
                    ),
                    Expanded(
                      child: _HeaderLoadingMeta(
                        label: '施测者',
                        value: examinerName,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              const _SkeletonPill(width: 78, height: 14),
              const SizedBox(width: 12),
              const _SkeletonButton(width: 112),
              const SizedBox(width: 9),
              const _SkeletonButton(width: 112, filled: true),
            ],
          );
        },
      ),
    );
  }
}

class _HeaderLoadingMeta extends StatelessWidget {
  const _HeaderLoadingMeta({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final String resolved = value.trim();
    return Container(
      margin: const EdgeInsets.only(left: 10),
      padding: const EdgeInsets.only(left: 10),
      decoration: const BoxDecoration(
        border: Border(left: BorderSide(color: _Pep3Colors.line)),
      ),
      child: FittedBox(
        fit: BoxFit.scaleDown,
        alignment: Alignment.centerLeft,
        child: resolved.isEmpty
            ? Row(
                mainAxisSize: MainAxisSize.min,
                children: <Widget>[
                  Text(
                    '$label：',
                    maxLines: 1,
                    style: const TextStyle(
                      color: _Pep3Colors.text,
                      fontSize: 13,
                      height: 1,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  const SizedBox(width: 5),
                  const _LoadingLine(width: 40, height: 12),
                ],
              )
            : Text.rich(
                TextSpan(
                  children: <InlineSpan>[
                    TextSpan(text: '$label：'),
                    TextSpan(
                      text: resolved,
                      style: const TextStyle(fontWeight: FontWeight.w900),
                    ),
                  ],
                ),
                maxLines: 1,
                style: const TextStyle(
                  color: _Pep3Colors.text,
                  fontSize: 13,
                  height: 1,
                  fontWeight: FontWeight.w700,
                ),
              ),
      ),
    );
  }
}

class _Pep3SidebarSkeleton extends StatelessWidget {
  const _Pep3SidebarSkeleton();

  @override
  Widget build(BuildContext context) {
    return _RailCard(
      child: Column(
        children: <Widget>[
          Container(
            height: 48,
            padding: const EdgeInsets.symmetric(horizontal: 14),
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: _Pep3Colors.lineSoft)),
            ),
            child: const Row(
              children: <Widget>[
                _LoadingLine(width: 88, height: 18),
                Spacer(),
                _SkeletonPill(width: 18, height: 18, radius: 6),
              ],
            ),
          ),
          Expanded(
            child: ListView.separated(
              padding: const EdgeInsets.fromLTRB(12, 13, 10, 13),
              physics: const NeverScrollableScrollPhysics(),
              itemBuilder: (BuildContext context, int index) {
                return _PageGroupSkeleton(expanded: index == 0);
              },
              separatorBuilder: (BuildContext context, int index) {
                return const Divider(height: 18, color: _Pep3Colors.lineSoft);
              },
              itemCount: 7,
            ),
          ),
        ],
      ),
    );
  }
}

class _PageGroupSkeleton extends StatelessWidget {
  const _PageGroupSkeleton({required this.expanded});

  final bool expanded;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        const Row(
          children: <Widget>[
            _SkeletonPill(width: 20, height: 20, radius: 7),
            SizedBox(width: 8),
            Expanded(child: _LoadingLine(widthFactor: .72, height: 14)),
            SizedBox(width: 10),
            _LoadingLine(width: 36, height: 12),
          ],
        ),
        const SizedBox(height: 9),
        const Row(
          children: <Widget>[
            SizedBox(width: 28),
            Expanded(child: _SkeletonPill(height: 5, radius: 99)),
            SizedBox(width: 9),
            _LoadingLine(width: 28, height: 12),
          ],
        ),
        if (expanded) ...<Widget>[
          const SizedBox(height: 10),
          for (int i = 0; i < 5; i++) ...<Widget>[
            const Padding(
              padding: EdgeInsets.only(left: 18, bottom: 8),
              child: Row(
                children: <Widget>[
                  _LoadingLine(width: 45, height: 12),
                  SizedBox(width: 8),
                  Expanded(child: _LoadingLine(widthFactor: .72, height: 12)),
                  SizedBox(width: 8),
                  _SkeletonPill(width: 15, height: 15, radius: 99),
                ],
              ),
            ),
          ],
        ],
      ],
    );
  }
}

class _Pep3QuestionSkeleton extends StatelessWidget {
  const _Pep3QuestionSkeleton();

  @override
  Widget build(BuildContext context) {
    return _RailCard(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 18, 16, 14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            const Row(
              children: <Widget>[
                Expanded(child: _LoadingLine(widthFactor: .72, height: 28)),
                SizedBox(width: 12),
                _SkeletonPill(width: 154, height: 34, radius: 10),
              ],
            ),
            const SizedBox(height: 16),
            Expanded(
              child: ListView(
                padding: EdgeInsets.zero,
                physics: const NeverScrollableScrollPhysics(),
                children: const <Widget>[
                  _QuestionLoadingCard(title: '材料', height: 72),
                  _QuestionLoadingCard(title: '操作标准', height: 86),
                  _QuestionLoadingCard(title: '指导语', height: 86),
                  _QuestionLoadingCard(title: '评分标准', height: 86),
                  SizedBox(height: 2),
                ],
              ),
            ),
            const SizedBox(height: 10),
            const Row(
              children: <Widget>[
                _LoadingLine(width: 40, height: 16),
                Spacer(),
                _SkeletonPill(width: 188, height: 30, radius: 9),
              ],
            ),
            const SizedBox(height: 10),
            const Row(
              children: <Widget>[
                Expanded(child: _ScoreLoadingCard()),
                SizedBox(width: 14),
                Expanded(child: _ScoreLoadingCard()),
                SizedBox(width: 14),
                Expanded(child: _ScoreLoadingCard()),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _Pep3RightRailSkeleton extends StatelessWidget {
  const _Pep3RightRailSkeleton();

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: EdgeInsets.zero,
      physics: const NeverScrollableScrollPhysics(),
      children: const <Widget>[
        _RailCard(
          child: Padding(
            padding: EdgeInsets.all(16),
            child: _ProgressSkeleton(),
          ),
        ),
        SizedBox(height: 10),
        _RailCard(
          child: Padding(
            padding: EdgeInsets.all(16),
            child: _TrainingRecordSkeleton(),
          ),
        ),
        SizedBox(height: 10),
        _RailCard(
          child: Padding(
            padding: EdgeInsets.all(16),
            child: _CaregiverSkeleton(),
          ),
        ),
      ],
    );
  }
}

class _ProgressSkeleton extends StatelessWidget {
  const _ProgressSkeleton();

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        const _SkeletonPill(width: 82, height: 82, radius: 99),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: const <Widget>[
              _LoadingLine(width: 72, height: 16),
              SizedBox(height: 14),
              _LoadingLine(widthFactor: .86, height: 13),
              SizedBox(height: 11),
              _LoadingLine(widthFactor: .68, height: 13),
            ],
          ),
        ),
      ],
    );
  }
}

class _TrainingRecordSkeleton extends StatelessWidget {
  const _TrainingRecordSkeleton();

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: const <Widget>[
        _LoadingLine(width: 96, height: 18),
        SizedBox(height: 14),
        _SkeletonInputBlock(),
        SizedBox(height: 10),
        _SkeletonInputBlock(),
      ],
    );
  }
}

class _CaregiverSkeleton extends StatelessWidget {
  const _CaregiverSkeleton();

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: const <Widget>[
        _LoadingLine(width: 82, height: 18),
        SizedBox(height: 16),
        Center(child: _SkeletonPill(width: 122, height: 122, radius: 10)),
        SizedBox(height: 14),
        Center(child: _LoadingLine(width: 112, height: 13)),
        SizedBox(height: 14),
        Row(
          children: <Widget>[
            Expanded(child: _SkeletonButton(width: double.infinity)),
            SizedBox(width: 8),
            Expanded(child: _SkeletonButton(width: double.infinity)),
          ],
        ),
      ],
    );
  }
}

class _Pep3FooterSkeleton extends StatelessWidget {
  const _Pep3FooterSkeleton();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 62,
      padding: const EdgeInsets.symmetric(horizontal: 20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.97),
        border: Border.all(color: _Pep3Colors.line),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
        boxShadow: _pep3Shadow(color: const Color(0x14B05F32), blur: 16),
      ),
      child: const Row(
        children: <Widget>[
          _SkeletonButton(width: 134),
          Spacer(),
          _LoadingLine(width: 80, height: 28),
          Spacer(),
          _SkeletonButton(width: 150, filled: true),
          SizedBox(width: 14),
          _SkeletonButton(width: 134),
          SizedBox(width: 22),
          _LoadingLine(width: 66, height: 13),
          SizedBox(width: 8),
          _SkeletonPill(width: 50, height: 30, radius: 99),
        ],
      ),
    );
  }
}

class _SkeletonInputBlock extends StatelessWidget {
  const _SkeletonInputBlock();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      height: 80,
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _Pep3Colors.lineSoft),
      ),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          _LoadingLine(width: 74, height: 13),
          SizedBox(height: 10),
          _SkeletonPill(height: 32, radius: 8),
        ],
      ),
    );
  }
}

class _SkeletonButton extends StatelessWidget {
  const _SkeletonButton({
    required this.width,
    this.filled = false,
  });

  final double width;
  final bool filled;

  @override
  Widget build(BuildContext context) {
    return _SkeletonPill(
      width: width,
      height: 38,
      radius: 10,
      color: filled ? const Color(0xFFF7C1A8) : null,
    );
  }
}

class _SkeletonPill extends StatelessWidget {
  const _SkeletonPill({
    this.width,
    required this.height,
    this.radius = 99,
    this.color,
  });

  final double? width;
  final double height;
  final double radius;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: color ?? const Color(0xFFF3E8DF),
        borderRadius: BorderRadius.circular(radius),
      ),
    );
  }
}

class _Pep3ErrorShell extends StatelessWidget {
  const _Pep3ErrorShell({required this.message, required this.onBack});

  final String message;
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: _Pep3Colors.page,
      child: Center(
        child: Container(
          width: 440,
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: _Pep3Colors.line),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(Icons.error_outline_rounded,
                  size: 34, color: _Pep3Colors.orange),
              const SizedBox(height: 12),
              Text(
                message,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: _Pep3Colors.text,
                  fontSize: 14,
                  height: 1.5,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 16),
              _FooterButton(
                label: '返回',
                icon: Icons.chevron_left_rounded,
                enabled: true,
                onTap: onBack,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _Pep3Colors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color ink = Color(0xFF432B22);
  static const Color text = Color(0xFF7F665A);
  static const Color muted = Color(0xFFBBA99C);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95735);
  static const Color green = Color(0xFF6F9F70);
  static const Color blue = Color(0xFF3F82D2);
  static const Color red = Color(0xFFD94A42);
  static const Color line = Color(0xFFF0DACB);
  static const Color lineSoft = Color(0xFFF6E7DC);
}

List<BoxShadow> _pep3Shadow({
  Color color = const Color(0x18000000),
  double blur = 18,
  Offset offset = const Offset(0, 9),
}) {
  return <BoxShadow>[
    BoxShadow(color: color, blurRadius: blur, offset: offset),
  ];
}

Color _scoreColor(int value) {
  if (value == 2) {
    return const Color(0xFF159947);
  }
  if (value == 0) {
    return _Pep3Colors.red;
  }
  return _Pep3Colors.blue;
}

String _shortScoreLabel(int value, String fallback) {
  if (value == 2) {
    return '通过';
  }
  if (value == 1) {
    return '部分通过';
  }
  if (value == 0) {
    return '未通过';
  }
  return fallback.trim();
}

String _scoreStandardText(
  Pep3AssessmentItem item,
  List<Pep3ScoreOption> options,
) {
  final String standard = _normalizeText(item.standard, fallback: '');
  if (standard.isNotEmpty) {
    return standard;
  }
  return options
      .map((Pep3ScoreOption option) =>
          '${option.value} 分（${_shortScoreLabel(option.value, option.label)}）：${option.description.trim().isEmpty ? option.label : option.description}')
      .join('\n');
}

String _normalizeText(String? value, {String fallback = '-'}) {
  final String text = '${value ?? ''}'.replaceAll(RegExp(r'\s+'), ' ').trim();
  return text.isEmpty ? fallback : text;
}

String _todayIsoDate() {
  final DateTime now = DateTime.now();
  return '${now.year.toString().padLeft(4, '0')}-${now.month.toString().padLeft(2, '0')}-${now.day.toString().padLeft(2, '0')}';
}

String? _normalizeDate(String? value) {
  final String text = '${value ?? ''}'.trim();
  if (text.isEmpty) {
    return null;
  }
  final DateTime? parsed = DateTime.tryParse(text);
  if (parsed == null) {
    return text.length >= 10 ? text.substring(0, 10) : text;
  }
  return '${parsed.year.toString().padLeft(4, '0')}-${parsed.month.toString().padLeft(2, '0')}-${parsed.day.toString().padLeft(2, '0')}';
}

String _formatDateTime(String value) {
  final String text = value.trim();
  if (text.isEmpty) {
    return '-';
  }
  final DateTime? parsed = DateTime.tryParse(text);
  if (parsed == null) {
    if (text.length >= 16) {
      return text.substring(0, 16).replaceFirst('T', ' ');
    }
    return text;
  }
  final DateTime local = parsed.toLocal();
  return '${local.year.toString().padLeft(4, '0')}-'
      '${local.month.toString().padLeft(2, '0')}-'
      '${local.day.toString().padLeft(2, '0')} '
      '${local.hour.toString().padLeft(2, '0')}:'
      '${local.minute.toString().padLeft(2, '0')}';
}

String _compactDateLabel(String value) {
  return _normalizeDate(value) ?? value.trim();
}

String _shortDateLabel(String value) {
  final String normalized = _normalizeDate(value) ?? value.trim();
  if (normalized.length >= 10) {
    return normalized.substring(5, 10);
  }
  return normalized;
}

String _assessmentAgeText(String birthDate, String assessmentDate) {
  final DateTime? birth = DateTime.tryParse(birthDate);
  final DateTime? target = DateTime.tryParse(assessmentDate);
  if (birth == null || target == null || birth.isAfter(target)) {
    return '';
  }
  int years = target.year - birth.year;
  DateTime anchor = DateTime(birth.year + years, birth.month, birth.day);
  if (anchor.isAfter(target)) {
    years -= 1;
    anchor = DateTime(birth.year + years, birth.month, birth.day);
  }
  int months = (target.year - anchor.year) * 12 + target.month - anchor.month;
  DateTime monthAnchor =
      DateTime(anchor.year, anchor.month + months, anchor.day);
  if (monthAnchor.isAfter(target)) {
    months -= 1;
    monthAnchor = DateTime(anchor.year, anchor.month + months, anchor.day);
  }
  final int days = target.difference(monthAnchor).inDays;
  final List<String> parts = <String>[];
  if (years > 0) {
    parts.add('$years岁');
  }
  if (months > 0) {
    parts.add('$months月');
  }
  if (days > 0 || parts.isEmpty) {
    parts.add('$days天');
  }
  return parts.join('');
}

bool _isEmptyRecordValue(dynamic value) {
  if (value == null) {
    return true;
  }
  if (value is String) {
    return value.trim().isEmpty;
  }
  if (value is Iterable) {
    return value.isEmpty;
  }
  return false;
}

extension _FirstOrNull<T> on Iterable<T> {
  T? get firstOrNull {
    final Iterator<T> iterator = this.iterator;
    if (iterator.moveNext()) {
      return iterator.current;
    }
    return null;
  }
}
