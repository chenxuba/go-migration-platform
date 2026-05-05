import 'dart:async';
import 'dart:math' as math;

import 'package:assessment_pad_app/assessment_scale_client.dart';
import 'package:assessment_pad_app/chinese_ime_engine.dart';
import 'package:assessment_pad_app/pep3_assessment_client.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:shared_preferences/shared_preferences.dart';

class AssessmentScaleCategoryScreen extends StatefulWidget {
  const AssessmentScaleCategoryScreen({
    required this.onBack,
    this.scaleClient = const ApiAssessmentScaleClient(),
    super.key,
  });

  final VoidCallback onBack;
  final AssessmentScaleClient scaleClient;

  @override
  State<AssessmentScaleCategoryScreen> createState() =>
      _AssessmentScaleCategoryScreenState();
}

class _AssessmentScaleCategoryScreenState
    extends State<AssessmentScaleCategoryScreen> {
  static const String _authTokenStorageKey = 'auth_token';
  static const ChineseImeEngine _searchImeEngine =
      ChineseImeEngine(dictionary: assessmentScaleImeDictionary);

  ChineseImeEditingValue _searchImeValue = const ChineseImeEditingValue();
  bool _searchKeyboardVisible = false;
  bool _searchKeyboardShifted = false;
  Timer? _searchDebounceTimer;
  List<String> _categories = <String>[];
  Map<String, int> _categoryCounts = <String, int>{};
  List<AssessmentScaleItem> _scales = <AssessmentScaleItem>[];
  List<AssessmentDraftSummary> _drafts = <AssessmentDraftSummary>[];
  List<AssessmentStudentCandidate> _studentCandidates =
      <AssessmentStudentCandidate>[];
  AssessmentStudentCandidate? _selectedStudent;
  AssessmentScaleLibrarySummary _summary =
      const AssessmentScaleLibrarySummary();
  String _selectedCategory = '';
  bool _categoryLoading = true;
  bool _scalesLoading = true;
  bool _scalesInitialized = false;
  bool _draftsLoading = true;
  bool _studentsLoading = false;
  int _draftCount = 0;
  int _scaleLoadSerial = 0;
  String? _categoryErrorMessage;
  String? _scaleErrorMessage;
  String? _draftErrorMessage;
  String? _studentErrorMessage;

  String get _searchQuery => _searchImeValue.text;

  String get _activeCategoryTitle {
    if (_selectedCategory.trim().isNotEmpty) {
      return _selectedCategory.trim();
    }
    if (_searchQuery.trim().isNotEmpty) {
      return '搜索结果';
    }
    return '全部量表';
  }

  @override
  void initState() {
    super.initState();
    _loadInitialData();
  }

  @override
  void dispose() {
    _searchDebounceTimer?.cancel();
    super.dispose();
  }

  Future<String> _readToken() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    return prefs.getString(_authTokenStorageKey) ?? '';
  }

  Future<void> _loadInitialData() async {
    if (mounted) {
      setState(() {
        _categoryLoading = true;
        _scalesLoading = true;
        _scalesInitialized = false;
        _draftsLoading = true;
        _categoryErrorMessage = null;
        _scaleErrorMessage = null;
        _draftErrorMessage = null;
      });
    }
    final String token = await _readToken();
    if (token.trim().isEmpty) {
      if (!mounted) {
        return;
      }
      setState(() {
        _categoryLoading = false;
        _scalesLoading = false;
        _scalesInitialized = true;
        _draftsLoading = false;
        _categoryErrorMessage = '请先登录后再查看量表分类';
        _scaleErrorMessage = '请先登录后再查看量表';
        _draftErrorMessage = '请先登录后再查看草稿';
      });
      return;
    }

    unawaited(_refreshDrafts());

    try {
      final List<String> categories =
          await widget.scaleClient.fetchCategories(token);
      if (!mounted) {
        return;
      }
      final String currentSelected = _selectedCategory.trim();
      final String nextSelected =
          currentSelected.isEmpty || categories.contains(currentSelected)
              ? currentSelected
              : '';
      setState(() {
        _categories = categories;
        _selectedCategory = nextSelected;
        _categoryLoading = false;
        _categoryErrorMessage = null;
      });
      unawaited(_loadScales());
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _categoryLoading = false;
        _scalesLoading = false;
        _scalesInitialized = true;
        _categoryErrorMessage = error.message;
        _scaleErrorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _categoryLoading = false;
        _scalesLoading = false;
        _scalesInitialized = true;
        _categoryErrorMessage = '分类加载失败：$error';
        _scaleErrorMessage = '量表加载失败：$error';
      });
    }
  }

  Future<void> _loadScales() async {
    final int serial = ++_scaleLoadSerial;
    if (mounted) {
      setState(() {
        _scalesLoading = true;
        _scaleErrorMessage = null;
      });
    }
    final String token = await _readToken();
    if (token.trim().isEmpty) {
      if (!mounted || serial != _scaleLoadSerial) {
        return;
      }
      setState(() {
        _scalesLoading = false;
        _scalesInitialized = true;
        _scaleErrorMessage = '请先登录后再查看量表';
      });
      return;
    }
    try {
      final AssessmentScaleLibrary result =
          await widget.scaleClient.fetchScaleLibrary(
        token,
        keyword: _searchQuery,
        category: _selectedCategory,
      );
      if (!mounted || serial != _scaleLoadSerial) {
        return;
      }
      final List<String> mergedCategories =
          _mergeCategories(_categories, result.filterOptions.categories);
      setState(() {
        _scales = result.items;
        _summary = result.summary;
        _categoryCounts = result.filterOptions.categoryCounts;
        if (_categories.length != mergedCategories.length ||
            !_sameStringList(_categories, mergedCategories)) {
          _categories = mergedCategories;
        }
        _scalesLoading = false;
        _scalesInitialized = true;
        _scaleErrorMessage = null;
      });
    } on AssessmentScaleApiException catch (error) {
      if (!mounted || serial != _scaleLoadSerial) {
        return;
      }
      setState(() {
        _scalesLoading = false;
        _scalesInitialized = true;
        _scaleErrorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted || serial != _scaleLoadSerial) {
        return;
      }
      setState(() {
        _scalesLoading = false;
        _scalesInitialized = true;
        _scaleErrorMessage = '量表加载失败：$error';
      });
    }
  }

  Future<void> _refreshDrafts({bool openAfterLoad = false}) async {
    if (mounted) {
      setState(() {
        _draftsLoading = true;
        _draftErrorMessage = null;
      });
    }
    final String token = await _readToken();
    if (token.trim().isEmpty) {
      if (!mounted) {
        return;
      }
      setState(() {
        _draftsLoading = false;
        _draftErrorMessage = '请先登录后再查看草稿';
      });
      if (openAfterLoad) {
        _showDraftsSheet();
      }
      return;
    }
    try {
      final AssessmentDraftPage drafts = await widget.scaleClient
          .fetchDraftsPage(token, pageSize: 100, latestOnly: true);
      if (!mounted) {
        return;
      }
      setState(() {
        _drafts = drafts.items;
        _draftCount = drafts.total;
        _draftsLoading = false;
        _draftErrorMessage = null;
      });
      if (openAfterLoad) {
        _showDraftsSheet();
      }
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _draftsLoading = false;
        _draftErrorMessage = error.message;
      });
      if (openAfterLoad) {
        _showDraftsSheet();
      }
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _draftsLoading = false;
        _draftErrorMessage = '草稿加载失败：$error';
      });
      if (openAfterLoad) {
        _showDraftsSheet();
      }
    }
  }

  void _selectCategory(String category) {
    if (_selectedCategory == category) {
      return;
    }
    _searchDebounceTimer?.cancel();
    setState(() => _selectedCategory = category);
    _loadScales();
  }

  void _scheduleSearchReload() {
    _searchDebounceTimer?.cancel();
    _searchDebounceTimer = Timer(
      const Duration(milliseconds: 320),
      _loadScales,
    );
  }

  void _updateSearchValue(ChineseImeEditingValue value) {
    final String previousQuery = _searchQuery;
    setState(() => _searchImeValue = value);
    if (_normalizeSearchText(previousQuery) !=
        _normalizeSearchText(_searchQuery)) {
      _scheduleSearchReload();
    }
  }

  void _showDraftsSheet() {
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      barrierColor: Colors.black.withOpacity(.16),
      elevation: 0,
      clipBehavior: Clip.none,
      builder: (BuildContext context) {
        return _DraftSheet(
          drafts: _drafts,
          total: _draftCount,
          loading: _draftsLoading,
          errorMessage: _draftErrorMessage,
          onRetry: () {
            Navigator.of(context).pop();
            _refreshDrafts(openAfterLoad: true);
          },
          onOpenDraft: (AssessmentDraftSummary draft) {
            Navigator.of(context).pop();
            _openDraft(draft);
          },
        );
      },
    );
  }

  Future<void> _openStudentSheet([
    AssessmentScaleItem? scaleToOpenAfterConfirm,
  ]) async {
    if (_studentCandidates.isEmpty && !_studentsLoading) {
      await _loadStudentCandidates();
    }
    if (!mounted) {
      return;
    }
    _showStudentSheet(scaleToOpenAfterConfirm);
  }

  Future<void> _loadStudentCandidates() async {
    if (mounted) {
      setState(() {
        _studentsLoading = true;
        _studentErrorMessage = null;
      });
    }
    final String token = await _readToken();
    if (token.trim().isEmpty) {
      if (!mounted) {
        return;
      }
      setState(() {
        _studentsLoading = false;
        _studentErrorMessage = '请先登录后再选择学员';
      });
      return;
    }
    try {
      final AssessmentStudentCandidatePage page =
          await widget.scaleClient.fetchStudentCandidates(token);
      if (!mounted) {
        return;
      }
      setState(() {
        _studentCandidates = page.items;
        _studentsLoading = false;
        _studentErrorMessage = null;
      });
    } on AssessmentScaleApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _studentsLoading = false;
        _studentErrorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _studentsLoading = false;
        _studentErrorMessage = '学员加载失败：$error';
      });
    }
  }

  void _showStudentSheet([
    AssessmentScaleItem? scaleToOpenAfterConfirm,
  ]) {
    showDialog<void>(
      context: context,
      barrierColor: Colors.black.withOpacity(.18),
      builder: (BuildContext dialogContext) {
        return _StudentDialog(
          students: _studentCandidates,
          selectedStudent: _selectedStudent,
          loading: _studentsLoading,
          errorMessage: _studentErrorMessage,
          confirmLabel: scaleToOpenAfterConfirm == null ? '确认选择' : '确认选择并进入测评',
          onRetry: () {
            Navigator.of(dialogContext).pop();
            _openStudentSheet(scaleToOpenAfterConfirm);
          },
          onConfirm: (AssessmentStudentCandidate student) {
            setState(() => _selectedStudent = student);
            Navigator.of(dialogContext).pop();
            final AssessmentScaleItem? nextScale = scaleToOpenAfterConfirm;
            if (nextScale != null) {
              WidgetsBinding.instance.addPostFrameCallback((_) {
                if (mounted) {
                  _chooseScale(nextScale);
                }
              });
            }
          },
        );
      },
    );
  }

  void _chooseScale(AssessmentScaleItem scale) {
    final AssessmentStudentCandidate? student = _selectedStudent;
    if (student == null || !scale.available) {
      return;
    }
    if (_isPep3Scale(scale)) {
      _openPep3Assessment(
        Pep3AssessmentLaunchArgs(
          studentId: student.id,
          studentName: student.displayName,
          studentAge: student.age.trim().isEmpty ? '未知' : student.age.trim(),
          birthDate: student.birthDate,
          assessmentDate: _todayIsoDate(),
          scaleName: scale.name,
        ),
      );
      return;
    }
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('${scale.name} 的作答页待接入'),
      ),
    );
  }

  void _openDraft(AssessmentDraftSummary draft) {
    if (_isPep3Draft(draft)) {
      _openPep3Assessment(
        Pep3AssessmentLaunchArgs(
          draftId: draft.id,
          studentName: draft.studentName,
          assessmentDate: _todayIsoDate(),
          examinerName: draft.examinerName,
          scaleName: draft.assessmentName.trim().isEmpty
              ? 'PEP-3'
              : draft.assessmentName.trim(),
        ),
      );
      return;
    }
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('${draft.assessmentName} 的作答页待接入')),
    );
  }

  void _openPep3Assessment(Pep3AssessmentLaunchArgs args) {
    Navigator.of(context).pushNamed('/pep3-assessment', arguments: args);
  }

  void _openSearchKeyboard() {
    setState(() => _searchKeyboardVisible = true);
  }

  void _finishSearchKeyboard() {
    final String previousQuery = _searchQuery;
    final ChineseImeEditingValue committed =
        _searchImeEngine.commit(_searchImeValue);
    setState(() {
      _searchImeValue = committed;
      _searchKeyboardVisible = false;
      _searchKeyboardShifted = false;
    });
    if (_normalizeSearchText(previousQuery) !=
        _normalizeSearchText(committed.text)) {
      _scheduleSearchReload();
    }
  }

  void _insertSearchText(String value) {
    _updateSearchValue(_searchImeEngine.handleKey(_searchImeValue, value));
  }

  void _replaceSearchText(String value) {
    _updateSearchValue(_searchImeEngine.replace(value));
  }

  void _commitPinyinCandidate(String value) {
    _updateSearchValue(
      _searchImeEngine.commitCandidate(
        _searchImeValue,
        value,
      ),
    );
  }

  void _deleteSearchText() {
    _updateSearchValue(_searchImeEngine.backspace(_searchImeValue));
  }

  void _clearSearchText() {
    _updateSearchValue(_searchImeEngine.clear());
  }

  void _toggleSearchShift() {
    setState(() => _searchKeyboardShifted = !_searchKeyboardShifted);
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width =
            constraints.maxWidth.isFinite ? constraints.maxWidth : 1366;
        final bool compact = width < 1180;
        final double margin = compact ? 24 : 32;
        final double leftWidth = compact ? 214 : 232;
        final double contentGap = compact ? 12 : 22;
        const double searchKeyboardTop = 88;

        return ColoredBox(
          color: _ScaleColors.page,
          child: Stack(
            children: <Widget>[
              const Positioned.fill(child: _ScalePageBackground()),
              Padding(
                padding: EdgeInsets.fromLTRB(margin, 26, margin, 22),
                child: Column(
                  children: <Widget>[
                    _ScaleTopBar(
                      onBack: widget.onBack,
                      compact: compact,
                      searchQuery: _searchQuery,
                      searchActive: _searchKeyboardVisible,
                      selectedStudent: _selectedStudent,
                      onSearchTap: _openSearchKeyboard,
                      onSearchClear: _clearSearchText,
                      onStudentTap: _openStudentSheet,
                    ),
                    const SizedBox(height: 22),
                    Expanded(
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: <Widget>[
                          SizedBox(
                            width: leftWidth,
                            child: _ScaleCategorySidebar(
                              categories: _categories,
                              categoryCounts: _categoryCounts,
                              selectedCategory: _selectedCategory,
                              totalCount: _summary.total,
                              loading: _categoryLoading,
                              errorMessage: _categoryErrorMessage,
                              draftCount: _draftCount,
                              draftsLoading: _draftsLoading,
                              draftErrorMessage: _draftErrorMessage,
                              onCategoryTap: _selectCategory,
                              onDraftTap: () =>
                                  _refreshDrafts(openAfterLoad: true),
                              onRetry: _loadInitialData,
                            ),
                          ),
                          SizedBox(width: contentGap),
                          Expanded(
                            child: _ScaleMainContent(
                              searchQuery: _searchQuery,
                              categoryTitle: _activeCategoryTitle,
                              scales: _scales,
                              summary: _summary,
                              loading: _scalesLoading,
                              initialLoading:
                                  _scalesLoading && !_scalesInitialized,
                              errorMessage: _scaleErrorMessage,
                              hasSelectedStudent: _selectedStudent != null,
                              onChooseScale: _chooseScale,
                              onRequireStudent: _openStudentSheet,
                              onRetry: _loadScales,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              if (_searchKeyboardVisible)
                Positioned.fill(
                  top: searchKeyboardTop,
                  child: GestureDetector(
                    behavior: HitTestBehavior.translucent,
                    onTap: _finishSearchKeyboard,
                  ),
                ),
              if (_searchKeyboardVisible)
                Positioned(
                  right: margin,
                  top: searchKeyboardTop,
                  child: GestureDetector(
                    behavior: HitTestBehavior.opaque,
                    onTap: () {},
                    child: ChineseImeKeyboard(
                      composing: _searchImeValue.composing,
                      candidates: _searchImeValue.candidates,
                      shifted: _searchKeyboardShifted,
                      compact: compact,
                      keyPrefix: 'scale-search',
                      onKey: _insertSearchText,
                      onReplace: _replaceSearchText,
                      onCandidate: _commitPinyinCandidate,
                      onBackspace: _deleteSearchText,
                      onClear: _clearSearchText,
                      onShift: _toggleSearchShift,
                      onClose: _finishSearchKeyboard,
                    ),
                  ),
                ),
            ],
          ),
        );
      },
    );
  }
}

class _ScaleColors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color card = Color(0xFFFFFEFB);
  static const Color ink = Color(0xFF3F2B22);
  static const Color text = Color(0xFF6F5B50);
  static const Color muted = Color(0xFFA7958B);
  static const Color line = Color(0xFFEAD7C9);
  static const Color lineSoft = Color(0xFFF4E8DF);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95D37);
}

List<BoxShadow> _scaleShadow({
  Color color = const Color(0x12B05F32),
  double blur = 24,
  Offset offset = const Offset(0, 12),
}) {
  return <BoxShadow>[BoxShadow(color: color, blurRadius: blur, offset: offset)];
}

class _ScalePageBackground extends StatelessWidget {
  const _ScalePageBackground();

  @override
  Widget build(BuildContext context) {
    return CustomPaint(painter: _ScalePageBackgroundPainter());
  }
}

class _ScalePageBackgroundPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint top = Paint()..color = const Color(0xFFFFF1E3);
    canvas.drawPath(
      Path()
        ..moveTo(0, 0)
        ..lineTo(size.width, 0)
        ..lineTo(size.width, 126)
        ..quadraticBezierTo(size.width * .55, 82, 0, 132)
        ..close(),
      top,
    );

    canvas.drawOval(
      Rect.fromLTWH(size.width - 240, 72, 210, 90),
      Paint()..color = const Color(0x34FFE0C2),
    );
    canvas.drawOval(
      Rect.fromLTWH(16, size.height - 118, 220, 92),
      Paint()..color = const Color(0x22F4C492),
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _ScaleTopBar extends StatelessWidget {
  const _ScaleTopBar({
    required this.onBack,
    required this.compact,
    required this.searchQuery,
    required this.searchActive,
    required this.selectedStudent,
    required this.onSearchTap,
    required this.onSearchClear,
    required this.onStudentTap,
  });

  final VoidCallback onBack;
  final bool compact;
  final String searchQuery;
  final bool searchActive;
  final AssessmentStudentCandidate? selectedStudent;
  final VoidCallback onSearchTap;
  final VoidCallback onSearchClear;
  final VoidCallback onStudentTap;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 62,
      child: Row(
        children: <Widget>[
          _IconShell(
            icon: Icons.chevron_left_rounded,
            onTap: onBack,
            size: 46,
            iconSize: 34,
          ),
          const SizedBox(width: 16),
          const Text(
            '开始测评',
            style: TextStyle(
              color: _ScaleColors.ink,
              fontSize: 30,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          const Spacer(),
          _SearchBox(
            width: compact ? 280 : 328,
            value: searchQuery,
            active: searchActive,
            onTap: onSearchTap,
            onClear: onSearchClear,
          ),
          const SizedBox(width: 14),
          _StudentChip(
            student: selectedStudent,
            onTap: onStudentTap,
          ),
          const SizedBox(width: 14),
          const _AvailableFilterChip(),
        ],
      ),
    );
  }
}

class _IconShell extends StatelessWidget {
  const _IconShell({
    required this.icon,
    required this.onTap,
    this.size = 44,
    this.iconSize = 25,
  });

  final IconData icon;
  final VoidCallback onTap;
  final double size;
  final double iconSize;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Ink(
          width: size,
          height: size,
          decoration: BoxDecoration(
            color: _ScaleColors.card,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: _ScaleColors.line),
          ),
          child: Icon(icon, size: iconSize, color: _ScaleColors.text),
        ),
      ),
    );
  }
}

class _SearchBox extends StatelessWidget {
  const _SearchBox({
    required this.width,
    required this.value,
    required this.active,
    required this.onTap,
    required this.onClear,
  });

  final double width;
  final String value;
  final bool active;
  final VoidCallback onTap;
  final VoidCallback onClear;

  @override
  Widget build(BuildContext context) {
    final bool hasValue = value.trim().isNotEmpty;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(15),
        child: Ink(
          width: width,
          height: 46,
          padding: const EdgeInsets.symmetric(horizontal: 17),
          decoration: BoxDecoration(
            color: Colors.white.withOpacity(.88),
            borderRadius: BorderRadius.circular(15),
            border: Border.all(
              color: active ? _ScaleColors.orange : _ScaleColors.line,
              width: active ? 1.4 : 1,
            ),
            boxShadow: active
                ? _scaleShadow(
                    color: const Color(0x18E96F43),
                    blur: 16,
                    offset: const Offset(0, 7),
                  )
                : null,
          ),
          child: Row(
            children: <Widget>[
              Icon(
                Icons.search_rounded,
                size: 22,
                color: active ? _ScaleColors.orange : _ScaleColors.text,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  key: const ValueKey<String>('scale-search-display-text'),
                  hasValue ? value : '搜索量表名称 / 编码',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: hasValue ? _ScaleColors.ink : _ScaleColors.muted,
                    fontSize: 15,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
              if (hasValue) ...<Widget>[
                const SizedBox(width: 8),
                GestureDetector(
                  key: const ValueKey<String>('scale-search-clear'),
                  behavior: HitTestBehavior.opaque,
                  onTap: onClear,
                  child: const Icon(
                    Icons.cancel_rounded,
                    size: 19,
                    color: _ScaleColors.muted,
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class ChineseImeKeyboard extends StatelessWidget {
  const ChineseImeKeyboard({
    required this.composing,
    required this.candidates,
    required this.shifted,
    required this.compact,
    required this.onKey,
    required this.onReplace,
    required this.onCandidate,
    required this.onBackspace,
    required this.onClear,
    required this.onShift,
    required this.onClose,
    this.keyPrefix = 'chinese-ime',
    this.quickWords = _defaultQuickWords,
  });

  final String composing;
  final List<String> candidates;
  final bool shifted;
  final bool compact;
  final String keyPrefix;
  final List<String> quickWords;
  final ValueChanged<String> onKey;
  final ValueChanged<String> onReplace;
  final ValueChanged<String> onCandidate;
  final VoidCallback onBackspace;
  final VoidCallback onClear;
  final VoidCallback onShift;
  final VoidCallback onClose;

  @override
  Widget build(BuildContext context) {
    return Container(
      key: ValueKey<String>('$keyPrefix-keyboard'),
      width: compact ? 600 : 660,
      padding: const EdgeInsets.fromLTRB(16, 14, 16, 16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.97),
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: _ScaleColors.line),
        boxShadow: _scaleShadow(
          color: const Color(0x28B05F32),
          blur: 28,
          offset: const Offset(0, 15),
        ),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Row(
            children: <Widget>[
              const Icon(
                Icons.keyboard_alt_outlined,
                color: _ScaleColors.orange,
                size: 21,
              ),
              const SizedBox(width: 8),
              const Text(
                '搜索量表',
                maxLines: 1,
                style: TextStyle(
                  color: _ScaleColors.orangeDeep,
                  fontSize: 16,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const Spacer(),
              _SearchKeyboardAction(
                label: '关闭',
                icon: Icons.close_rounded,
                onTap: onClose,
              ),
            ],
          ),
          const SizedBox(height: 12),
          _PinyinCandidateBar(
            keyPrefix: keyPrefix,
            pinyin: composing,
            candidates: candidates,
            onCandidate: onCandidate,
          ),
          const SizedBox(height: 12),
          Align(
            alignment: Alignment.centerLeft,
            child: Wrap(
              alignment: WrapAlignment.start,
              spacing: 9,
              runSpacing: 9,
              children: <Widget>[
                for (final String word in quickWords)
                  _SearchKeyboardQuickKey(
                    keyPrefix: keyPrefix,
                    label: word,
                    onTap: () => onReplace(word),
                  ),
              ],
            ),
          ),
          const SizedBox(height: 12),
          _keyRow(_digitKeys),
          const SizedBox(height: 9),
          _keyRow(_letterKeys('qwertyuiop')),
          const SizedBox(height: 9),
          _keyRow(_letterKeys('asdfghjkl')),
          const SizedBox(height: 9),
          _keyRow(<Widget>[
            _SearchKeyboardKey(
              keyPrefix: keyPrefix,
              label: '大写',
              active: shifted,
              flex: 2,
              onTap: onShift,
            ),
            ..._letterKeys('zxcvbnm'),
            _SearchKeyboardKey(
              keyPrefix: keyPrefix,
              label: '删除',
              flex: 2,
              onTap: onBackspace,
            ),
          ]),
          const SizedBox(height: 9),
          _keyRow(<Widget>[
            _SearchKeyboardKey(
              keyPrefix: keyPrefix,
              label: '-',
              onTap: () => onKey('-'),
            ),
            _SearchKeyboardKey(
              keyPrefix: keyPrefix,
              label: '/',
              onTap: () => onKey('/'),
            ),
            _SearchKeyboardKey(
              keyPrefix: keyPrefix,
              label: '空格',
              flex: 2,
              onTap: () => onKey(' '),
            ),
            _SearchKeyboardKey(
              keyPrefix: keyPrefix,
              label: '清空',
              flex: 2,
              muted: true,
              onTap: onClear,
            ),
            _SearchKeyboardKey(
              keyPrefix: keyPrefix,
              label: '完成',
              flex: 2,
              primary: true,
              onTap: onClose,
            ),
          ]),
        ],
      ),
    );
  }

  List<Widget> get _digitKeys {
    return '1234567890'
        .split('')
        .map((String value) => _SearchKeyboardKey(
              keyPrefix: keyPrefix,
              label: value,
              onTap: () => onKey(value),
            ))
        .toList();
  }

  List<Widget> _letterKeys(String values) {
    return values.split('').map((String value) {
      final String label = shifted ? value.toUpperCase() : value;
      return _SearchKeyboardKey(
        keyPrefix: keyPrefix,
        label: label,
        onTap: () => onKey(label),
      );
    }).toList();
  }

  Widget _keyRow(List<Widget> children) {
    return Row(
      children: <Widget>[
        for (int index = 0; index < children.length; index++) ...<Widget>[
          if (index > 0) const SizedBox(width: 8),
          children[index],
        ],
      ],
    );
  }

  static const List<String> _defaultQuickWords = <String>[
    'PEP-3',
    '语言',
    '沟通',
    '筛查',
    '口语',
    '表达',
    '社交',
    '综合',
  ];
}

class _PinyinCandidateBar extends StatelessWidget {
  const _PinyinCandidateBar({
    required this.keyPrefix,
    required this.pinyin,
    required this.candidates,
    required this.onCandidate,
  });

  final String keyPrefix;
  final String pinyin;
  final List<String> candidates;
  final ValueChanged<String> onCandidate;

  static const double _height = 52;
  static const double _contentHeight = 32;

  @override
  Widget build(BuildContext context) {
    final bool composing = pinyin.trim().isNotEmpty;
    return SizedBox(
      height: _height,
      width: double.infinity,
      child: Container(
        padding: const EdgeInsets.fromLTRB(10, 9, 10, 9),
        decoration: BoxDecoration(
          color: const Color(0xFFFFFAF5),
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: _ScaleColors.lineSoft),
        ),
        child: Row(
          children: <Widget>[
            Container(
              height: _contentHeight,
              constraints: const BoxConstraints(minWidth: 58),
              padding: const EdgeInsets.symmetric(horizontal: 10),
              alignment: Alignment.centerLeft,
              decoration: BoxDecoration(
                color: const Color(0xFFFFF1E8),
                borderRadius: BorderRadius.circular(10),
              ),
              child: Text(
                composing ? pinyin : '拼音',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _ScaleColors.orangeDeep,
                  fontSize: 14,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: SizedBox(
                height: _contentHeight,
                child: _buildCandidateContent(composing),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCandidateContent(bool composing) {
    if (!composing) {
      return const Align(
        alignment: Alignment.centerLeft,
        child: Text(
          '输入拼音后在这里选择汉字候选',
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: TextStyle(
            color: _ScaleColors.muted,
            fontSize: 13,
            height: 1,
            fontWeight: FontWeight.w700,
          ),
        ),
      );
    }
    if (candidates.isEmpty) {
      return const Align(
        alignment: Alignment.centerLeft,
        child: Text(
          '暂无候选，继续输入或直接搜索编码',
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: TextStyle(
            color: _ScaleColors.muted,
            fontSize: 13,
            height: 1,
            fontWeight: FontWeight.w700,
          ),
        ),
      );
    }
    return ListView.separated(
      scrollDirection: Axis.horizontal,
      padding: EdgeInsets.zero,
      physics: const BouncingScrollPhysics(),
      itemCount: candidates.length,
      separatorBuilder: (_, __) => const SizedBox(width: 8),
      itemBuilder: (BuildContext context, int index) {
        return _SearchKeyboardQuickKey(
          keyPrefix: keyPrefix,
          label: '${index + 1}.${candidates[index]}',
          onTap: () => onCandidate(candidates[index]),
        );
      },
    );
  }
}

class _SearchKeyboardQuickKey extends StatefulWidget {
  const _SearchKeyboardQuickKey({
    required this.keyPrefix,
    required this.label,
    required this.onTap,
  });

  final String keyPrefix;
  final String label;
  final VoidCallback onTap;

  @override
  State<_SearchKeyboardQuickKey> createState() =>
      _SearchKeyboardQuickKeyState();
}

class _SearchKeyboardQuickKeyState extends State<_SearchKeyboardQuickKey> {
  bool _pressed = false;
  bool _showBubble = false;
  int _feedbackToken = 0;

  void _handleTapDown(TapDownDetails _) {
    HapticFeedback.selectionClick();
    _feedbackToken++;
    setState(() {
      _pressed = true;
      _showBubble = true;
    });
  }

  void _hidePressBubbleSoon() {
    final int token = _feedbackToken;
    Future<void>.delayed(const Duration(milliseconds: 90), () {
      if (mounted && token == _feedbackToken) {
        setState(() => _pressed = false);
      }
    });
    Future<void>.delayed(const Duration(milliseconds: 210), () {
      if (mounted && token == _feedbackToken) {
        setState(() => _showBubble = false);
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      key: ValueKey<String>('${widget.keyPrefix}-key-${widget.label}'),
      onTapDown: _handleTapDown,
      onTapUp: (_) => _hidePressBubbleSoon(),
      onTapCancel: () {
        if (mounted) {
          setState(() {
            _pressed = false;
            _showBubble = false;
          });
        }
      },
      onTap: widget.onTap,
      child: Stack(
        clipBehavior: Clip.none,
        alignment: Alignment.center,
        children: <Widget>[
          AnimatedScale(
            scale: _pressed ? .96 : 1,
            duration: const Duration(milliseconds: 70),
            curve: Curves.easeOut,
            child: AnimatedContainer(
              duration: const Duration(milliseconds: 70),
              curve: Curves.easeOut,
              height: 32,
              padding: const EdgeInsets.symmetric(horizontal: 14),
              decoration: BoxDecoration(
                color: _pressed
                    ? const Color(0xFFFFE8DA)
                    : const Color(0xFFFFF8F1),
                borderRadius: BorderRadius.circular(999),
                border: Border.all(
                  color: _pressed ? _ScaleColors.orange : _ScaleColors.lineSoft,
                  width: _pressed ? 1.3 : 1,
                ),
              ),
              child: Center(
                widthFactor: 1,
                child: Text(
                  widget.label,
                  style: const TextStyle(
                    color: _ScaleColors.text,
                    fontSize: 13,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
          ),
          if (_showBubble)
            Positioned(
              bottom: 40,
              child: _SearchKeyboardBubble(label: widget.label),
            ),
        ],
      ),
    );
  }
}

class _SearchKeyboardBubble extends StatelessWidget {
  const _SearchKeyboardBubble({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return IgnorePointer(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Container(
            height: 46,
            constraints: const BoxConstraints(minWidth: 46),
            padding: const EdgeInsets.symmetric(horizontal: 15),
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: _ScaleColors.orange,
              borderRadius: BorderRadius.circular(14),
              boxShadow: _scaleShadow(
                color: const Color(0x24D15E36),
                blur: 14,
                offset: const Offset(0, 8),
              ),
            ),
            child: Text(
              label,
              maxLines: 1,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 20,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          ClipPath(
            clipper: _BubbleTriangleClipper(),
            child: Container(
              width: 12,
              height: 7,
              color: _ScaleColors.orange,
            ),
          ),
        ],
      ),
    );
  }
}

class _BubbleTriangleClipper extends CustomClipper<Path> {
  @override
  Path getClip(Size size) {
    return Path()
      ..moveTo(0, 0)
      ..lineTo(size.width, 0)
      ..lineTo(size.width / 2, size.height)
      ..close();
  }

  @override
  bool shouldReclip(covariant CustomClipper<Path> oldClipper) => false;
}

class _SearchKeyboardKey extends StatefulWidget {
  const _SearchKeyboardKey({
    required this.keyPrefix,
    required this.label,
    required this.onTap,
    this.flex = 1,
    this.primary = false,
    this.active = false,
    this.muted = false,
  });

  final String keyPrefix;
  final String label;
  final VoidCallback onTap;
  final int flex;
  final bool primary;
  final bool active;
  final bool muted;

  @override
  State<_SearchKeyboardKey> createState() => _SearchKeyboardKeyState();
}

class _SearchKeyboardKeyState extends State<_SearchKeyboardKey> {
  bool _pressed = false;
  bool _showBubble = false;
  int _feedbackToken = 0;

  void _handleTapDown(TapDownDetails _) {
    HapticFeedback.selectionClick();
    _feedbackToken++;
    setState(() {
      _pressed = true;
      _showBubble = true;
    });
  }

  void _hidePressBubbleSoon() {
    final int token = _feedbackToken;
    Future<void>.delayed(const Duration(milliseconds: 90), () {
      if (mounted && token == _feedbackToken) {
        setState(() => _pressed = false);
      }
    });
    Future<void>.delayed(const Duration(milliseconds: 210), () {
      if (mounted && token == _feedbackToken) {
        setState(() => _showBubble = false);
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final Color background = widget.primary
        ? _ScaleColors.orange
        : widget.active
            ? const Color(0xFFFFE7D9)
            : widget.muted
                ? const Color(0xFFF7EFE9)
                : const Color(0xFFFFFAF5);
    final Color foreground = widget.primary
        ? Colors.white
        : widget.active
            ? _ScaleColors.orangeDeep
            : _ScaleColors.ink;
    final Color pressedBackground = widget.primary
        ? _ScaleColors.orangeDeep
        : widget.active
            ? const Color(0xFFFFD6C2)
            : const Color(0xFFFFE8DA);
    final Color keyBackground = _pressed ? pressedBackground : background;
    final Color keyBorder =
        _pressed ? _ScaleColors.orange : _ScaleColors.lineSoft;

    return Expanded(
      flex: widget.flex,
      child: GestureDetector(
        key: ValueKey<String>('${widget.keyPrefix}-key-${widget.label}'),
        onTapDown: _handleTapDown,
        onTapUp: (_) => _hidePressBubbleSoon(),
        onTapCancel: () {
          if (mounted) {
            setState(() {
              _pressed = false;
              _showBubble = false;
            });
          }
        },
        onTap: widget.onTap,
        child: Stack(
          clipBehavior: Clip.none,
          alignment: Alignment.center,
          children: <Widget>[
            AnimatedScale(
              scale: _pressed ? .96 : 1,
              duration: const Duration(milliseconds: 70),
              curve: Curves.easeOut,
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 70),
                curve: Curves.easeOut,
                height: 48,
                decoration: BoxDecoration(
                  color: keyBackground,
                  borderRadius: BorderRadius.circular(12),
                  border: widget.primary
                      ? null
                      : Border.all(color: keyBorder, width: _pressed ? 1.4 : 1),
                  boxShadow: widget.primary || _pressed
                      ? _scaleShadow(
                          color: _pressed
                              ? const Color(0x2FD15E36)
                              : const Color(0x20D15E36),
                          blur: _pressed ? 16 : 12,
                          offset: Offset(0, _pressed ? 5 : 7),
                        )
                      : null,
                ),
                child: Center(
                  child: Text(
                    widget.label,
                    maxLines: 1,
                    overflow: TextOverflow.clip,
                    strutStyle: const StrutStyle(
                      fontSize: 15,
                      height: 1,
                      forceStrutHeight: true,
                    ),
                    style: TextStyle(
                      color: foreground,
                      fontSize: widget.label.length > 2 ? 13 : 16,
                      height: 1,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ),
            ),
            if (_showBubble)
              Positioned(
                bottom: 56,
                child: _SearchKeyboardBubble(label: widget.label),
              ),
          ],
        ),
      ),
    );
  }
}

class _SearchKeyboardAction extends StatelessWidget {
  const _SearchKeyboardAction({
    required this.label,
    required this.icon,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
          child: Row(
            children: <Widget>[
              Icon(icon, color: _ScaleColors.text, size: 18),
              const SizedBox(width: 3),
              Text(
                label,
                style: const TextStyle(
                  color: _ScaleColors.text,
                  fontSize: 12,
                  height: 1,
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

class _StudentChip extends StatelessWidget {
  const _StudentChip({required this.student, required this.onTap});

  final AssessmentStudentCandidate? student;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final AssessmentStudentCandidate? selected = student;
    final bool hasStudent = selected != null;
    final String label =
        selected == null ? '未选择学员' : _studentEchoLabel(selected);
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(15),
        child: Ink(
          width: 202,
          height: 46,
          padding: const EdgeInsets.symmetric(horizontal: 18),
          decoration: BoxDecoration(
            color: Colors.white.withOpacity(.86),
            borderRadius: BorderRadius.circular(15),
            border: Border.all(
              color: hasStudent ? _ScaleColors.orange : _ScaleColors.line,
              width: hasStudent ? 1.3 : 1,
            ),
          ),
          child: Row(
            children: <Widget>[
              Icon(
                Icons.person_outline_rounded,
                size: 23,
                color: hasStudent ? _ScaleColors.orange : _ScaleColors.text,
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  label,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color:
                        hasStudent ? _ScaleColors.orangeDeep : _ScaleColors.ink,
                    fontSize: 15,
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

String _studentEchoLabel(AssessmentStudentCandidate student) {
  final String age = student.age.trim().isNotEmpty ? student.age.trim() : '未知';
  return '${student.displayName} * $age';
}

bool _isPep3Scale(AssessmentScaleItem scale) {
  return _isPep3Text(
    <String>[
      scale.executionEntry,
      scale.apiPackage,
      scale.code,
      scale.name,
    ].join(' '),
  );
}

bool _isPep3Draft(AssessmentDraftSummary draft) {
  return _isPep3Text(
    <String>[
      draft.assessmentCode,
      draft.assessmentName,
    ].join(' '),
  );
}

bool _isPep3Text(String value) {
  final String normalized =
      value.toLowerCase().replaceAll(RegExp(r'[\s_\-]'), '');
  return normalized.contains('pep3');
}

String _todayIsoDate() {
  final DateTime now = DateTime.now();
  return '${now.year.toString().padLeft(4, '0')}-'
      '${now.month.toString().padLeft(2, '0')}-'
      '${now.day.toString().padLeft(2, '0')}';
}

class _StudentDialog extends StatefulWidget {
  const _StudentDialog({
    required this.students,
    required this.selectedStudent,
    required this.loading,
    required this.errorMessage,
    required this.confirmLabel,
    required this.onRetry,
    required this.onConfirm,
  });

  final List<AssessmentStudentCandidate> students;
  final AssessmentStudentCandidate? selectedStudent;
  final bool loading;
  final String? errorMessage;
  final String confirmLabel;
  final VoidCallback onRetry;
  final ValueChanged<AssessmentStudentCandidate> onConfirm;

  @override
  State<_StudentDialog> createState() => _StudentDialogState();
}

class _StudentDialogState extends State<_StudentDialog> {
  AssessmentStudentCandidate? _pendingStudent;

  @override
  void initState() {
    super.initState();
    _pendingStudent = widget.selectedStudent;
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Material(
        color: Colors.transparent,
        child: Container(
          width: 820,
          constraints: const BoxConstraints(maxHeight: 560),
          padding: const EdgeInsets.fromLTRB(24, 22, 24, 22),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(24),
            border: Border.all(color: _ScaleColors.line),
            boxShadow: _scaleShadow(
              color: const Color(0x30B05F32),
              blur: 36,
              offset: const Offset(0, 18),
            ),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Container(
                    width: 43,
                    height: 43,
                    decoration: const BoxDecoration(
                      color: Color(0xFFFFE8DA),
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.person_search_rounded,
                        size: 24, color: _ScaleColors.orange),
                  ),
                  const SizedBox(width: 14),
                  const Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Text(
                          '选择学员',
                          style: TextStyle(
                            color: _ScaleColors.ink,
                            fontSize: 16,
                            height: 1.1,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                        SizedBox(height: 9),
                        Text(
                          '开始测评前，请先选择本次测评对象。',
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            color: _ScaleColors.muted,
                            fontSize: 12,
                            height: 1,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ],
                    ),
                  ),
                  _SearchKeyboardAction(
                    label: '关闭',
                    icon: Icons.close_rounded,
                    onTap: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
              const SizedBox(height: 18),
              Flexible(child: _buildContent()),
              const SizedBox(height: 18),
              _buildFooter(context),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildContent() {
    if (widget.loading) {
      return const SizedBox(
        height: 260,
        child: Center(
          child: CircularProgressIndicator(
            strokeWidth: 3,
            color: _ScaleColors.orange,
          ),
        ),
      );
    }
    if (widget.errorMessage != null) {
      return SizedBox(
        height: 260,
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Text(
                widget.errorMessage!,
                maxLines: 2,
                textAlign: TextAlign.center,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _ScaleColors.text,
                  fontSize: 14,
                  height: 1.4,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 12),
              _MiniActionButton(label: '重试', onTap: widget.onRetry),
            ],
          ),
        ),
      );
    }
    if (widget.students.isEmpty) {
      return const SizedBox(
        height: 260,
        child: Center(
          child: Text(
            '暂无可选择学员',
            style: TextStyle(
              color: _ScaleColors.muted,
              fontSize: 15,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
      );
    }
    return Container(
      height: 300,
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _ScaleColors.lineSoft),
      ),
      child: Column(
        children: <Widget>[
          const _StudentDialogHeaderRow(),
          Expanded(
            child: ListView.separated(
              padding: const EdgeInsets.fromLTRB(12, 0, 12, 12),
              physics: const BouncingScrollPhysics(),
              itemCount: widget.students.length,
              separatorBuilder: (_, __) => const SizedBox(height: 8),
              itemBuilder: (BuildContext context, int index) {
                final AssessmentStudentCandidate student =
                    widget.students[index];
                return _StudentDialogItem(
                  student: student,
                  selected: _pendingStudent?.id == student.id,
                  onTap: () => setState(() => _pendingStudent = student),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFooter(BuildContext context) {
    final AssessmentStudentCandidate? pending = _pendingStudent;
    return Row(
      children: <Widget>[
        Expanded(
          child: Text(
            pending == null ? '已选择：未选择' : '已选择：${_studentEchoLabel(pending)}',
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: _ScaleColors.text,
              fontSize: 14,
              height: 1,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
        _DialogActionButton(
          label: '取消',
          primary: false,
          onTap: () => Navigator.of(context).pop(),
        ),
        const SizedBox(width: 12),
        _DialogActionButton(
          label: widget.confirmLabel,
          primary: true,
          enabled: pending != null,
          onTap: pending == null ? null : () => widget.onConfirm(pending),
        ),
      ],
    );
  }
}

class _StudentDialogHeaderRow extends StatelessWidget {
  const _StudentDialogHeaderRow();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: const Row(
        children: <Widget>[
          Expanded(flex: 3, child: _StudentDialogHeaderText('儿童姓名')),
          Expanded(child: _StudentDialogHeaderText('性别')),
          Expanded(child: _StudentDialogHeaderText('年龄')),
          Expanded(flex: 2, child: _StudentDialogHeaderText('联系方式')),
          Expanded(flex: 2, child: _StudentDialogHeaderText('最近测评')),
          SizedBox(width: 34),
        ],
      ),
    );
  }
}

class _StudentDialogHeaderText extends StatelessWidget {
  const _StudentDialogHeaderText(this.label);

  final String label;

  @override
  Widget build(BuildContext context) {
    return Text(
      label,
      maxLines: 1,
      overflow: TextOverflow.ellipsis,
      style: const TextStyle(
        color: _ScaleColors.muted,
        fontSize: 12,
        height: 1,
        fontWeight: FontWeight.w900,
      ),
    );
  }
}

class _StudentDialogItem extends StatelessWidget {
  const _StudentDialogItem({
    required this.student,
    required this.selected,
    required this.onTap,
  });

  final AssessmentStudentCandidate student;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Ink(
          height: 56,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFF0E7) : Colors.white,
            borderRadius: BorderRadius.circular(13),
            border: Border.all(
              color: selected ? _ScaleColors.orange : _ScaleColors.lineSoft,
              width: selected ? 1.3 : 1,
            ),
          ),
          child: Row(
            children: <Widget>[
              Expanded(
                flex: 3,
                child: Row(
                  children: <Widget>[
                    Container(
                      width: 34,
                      height: 34,
                      alignment: Alignment.center,
                      decoration: BoxDecoration(
                        color: selected
                            ? const Color(0xFFFFDCCB)
                            : const Color(0xFFFFE8DA),
                        shape: BoxShape.circle,
                      ),
                      child: Text(
                        student.displayShortName,
                        style: const TextStyle(
                          color: _ScaleColors.orangeDeep,
                          fontSize: 14,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: _StudentDialogCell(
                        student.displayName,
                        strong: true,
                      ),
                    ),
                  ],
                ),
              ),
              Expanded(child: _StudentDialogCell(student.gender)),
              Expanded(child: _StudentDialogCell(student.age)),
              Expanded(
                  flex: 2, child: _StudentDialogCell(student.contactPhone)),
              Expanded(
                flex: 2,
                child: _StudentDialogCell(student.latestAssessment),
              ),
              const SizedBox(width: 12),
              Icon(
                selected
                    ? Icons.check_circle_rounded
                    : Icons.radio_button_unchecked_rounded,
                size: 22,
                color: selected ? _ScaleColors.orange : _ScaleColors.muted,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StudentDialogCell extends StatelessWidget {
  const _StudentDialogCell(this.value, {this.strong = false});

  final String value;
  final bool strong;

  @override
  Widget build(BuildContext context) {
    return Text(
      value.trim().isEmpty ? '-' : value.trim(),
      maxLines: 1,
      overflow: TextOverflow.ellipsis,
      style: TextStyle(
        color: strong ? _ScaleColors.ink : _ScaleColors.text,
        fontSize: strong ? 14 : 13,
        height: 1,
        fontWeight: strong ? FontWeight.w900 : FontWeight.w800,
      ),
    );
  }
}

class _DialogActionButton extends StatelessWidget {
  const _DialogActionButton({
    required this.label,
    required this.primary,
    required this.onTap,
    this.enabled = true,
  });

  final String label;
  final bool primary;
  final VoidCallback? onTap;
  final bool enabled;

  @override
  Widget build(BuildContext context) {
    final bool active = enabled && onTap != null;
    final Color background = primary
        ? active
            ? _ScaleColors.orange
            : const Color(0xFFD8CAC2)
        : Colors.white;
    final Color foreground = primary
        ? Colors.white
        : active
            ? _ScaleColors.text
            : _ScaleColors.muted;
    final double width = label.length > 5 ? 168 : 116;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: active ? onTap : null,
        borderRadius: BorderRadius.circular(12),
        child: Ink(
          width: width,
          height: 42,
          decoration: BoxDecoration(
            color: background,
            borderRadius: BorderRadius.circular(12),
            border: primary ? null : Border.all(color: _ScaleColors.line),
          ),
          child: Center(
            child: Text(
              label,
              maxLines: 1,
              style: TextStyle(
                color: foreground,
                fontSize: 14,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _AvailableFilterChip extends StatelessWidget {
  const _AvailableFilterChip();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 46,
      padding: const EdgeInsets.symmetric(horizontal: 17),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.86),
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: _ScaleColors.line),
      ),
      child: const Row(
        children: <Widget>[
          Icon(Icons.filter_alt_outlined, size: 22, color: _ScaleColors.text),
          SizedBox(width: 9),
          Text(
            '停用量表',
            style: TextStyle(
              color: _ScaleColors.ink,
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
          SizedBox(width: 9),
          _FilterDot(),
        ],
      ),
    );
  }
}

class _FilterDot extends StatelessWidget {
  const _FilterDot();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 8,
      height: 8,
      decoration: const BoxDecoration(
        color: _ScaleColors.orange,
        shape: BoxShape.circle,
      ),
    );
  }
}

class _ScaleCategorySidebar extends StatelessWidget {
  const _ScaleCategorySidebar({
    required this.categories,
    required this.categoryCounts,
    required this.selectedCategory,
    required this.totalCount,
    required this.loading,
    required this.errorMessage,
    required this.draftCount,
    required this.draftsLoading,
    required this.draftErrorMessage,
    required this.onCategoryTap,
    required this.onDraftTap,
    required this.onRetry,
  });

  final List<String> categories;
  final Map<String, int> categoryCounts;
  final String selectedCategory;
  final int totalCount;
  final bool loading;
  final String? errorMessage;
  final int draftCount;
  final bool draftsLoading;
  final String? draftErrorMessage;
  final ValueChanged<String> onCategoryTap;
  final VoidCallback onDraftTap;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    final List<String> visibleCategories = categories
        .map((String item) => item.trim())
        .where((String item) => item.isNotEmpty)
        .toList();
    return Column(
      children: <Widget>[
        Expanded(
          child: Container(
            padding: const EdgeInsets.fromLTRB(14, 17, 14, 16),
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(.88),
              borderRadius: BorderRadius.circular(20),
              border: Border.all(color: _ScaleColors.line),
              boxShadow: _scaleShadow(),
            ),
            child: Column(
              children: <Widget>[
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 2),
                  child: Row(
                    children: <Widget>[
                      const Text(
                        '分类',
                        style: TextStyle(
                          color: _ScaleColors.ink,
                          fontSize: 19,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                      const Spacer(),
                      if (loading)
                        const _ScaleSkeletonBlock(width: 24, height: 12)
                      else
                        Text(
                          '$totalCount',
                          style: const TextStyle(
                            color: _ScaleColors.muted,
                            fontSize: 12,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                    ],
                  ),
                ),
                const SizedBox(height: 14),
                Expanded(
                  child: _buildCategoryContent(visibleCategories),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(height: 14),
        _DraftCard(
          count: draftCount,
          loading: draftsLoading,
          errorMessage: draftErrorMessage,
          onTap: onDraftTap,
        ),
      ],
    );
  }

  Widget _buildCategoryContent(List<String> visibleCategories) {
    final int allCategoryCount = _allCategoryCount;
    if (loading && visibleCategories.isEmpty) {
      return const _CategorySkeletonList();
    }
    if (errorMessage != null && visibleCategories.isEmpty) {
      return _ScaleSidebarMessage(
        text: errorMessage!,
        actionLabel: '重试',
        onAction: onRetry,
      );
    }
    return _CategoryScrollViewport(
      items: <_CategoryItemData>[
        for (int index = 0; index < visibleCategories.length + 1; index++)
          _categoryItemDataFor(index, visibleCategories, allCategoryCount),
      ],
      onCategoryTap: (String name) => onCategoryTap(name == '全部' ? '' : name),
    );
  }

  _CategoryItemData _categoryItemDataFor(
    int index,
    List<String> visibleCategories,
    int allCategoryCount,
  ) {
    final bool allCategory = index == 0;
    final String name = allCategory ? '全部' : visibleCategories[index - 1];
    return _CategoryItemData(
      name,
      allCategory ? allCategoryCount : categoryCounts[name] ?? 0,
      _categoryAccentColor(index),
      active: allCategory
          ? selectedCategory.trim().isEmpty
          : name == selectedCategory,
    );
  }

  int get _allCategoryCount {
    final int countedTotal = categoryCounts.values.fold<int>(
      0,
      (int total, int count) => total + count,
    );
    if (countedTotal > 0) {
      return countedTotal;
    }
    return totalCount;
  }
}

class _CategoryItemData {
  const _CategoryItemData(
    this.name,
    this.count,
    this.color, {
    this.active = false,
  });

  final String name;
  final int count;
  final Color color;
  final bool active;
}

class _CategoryScrollViewport extends StatefulWidget {
  const _CategoryScrollViewport({
    required this.items,
    required this.onCategoryTap,
  });

  final List<_CategoryItemData> items;
  final ValueChanged<String> onCategoryTap;

  @override
  State<_CategoryScrollViewport> createState() =>
      _CategoryScrollViewportState();
}

class _CategoryScrollViewportState extends State<_CategoryScrollViewport> {
  final ScrollController _controller = ScrollController();
  bool _canScrollUp = false;
  bool _canScrollDown = false;

  @override
  void initState() {
    super.initState();
    _controller.addListener(_updateScrollHints);
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateScrollHints());
  }

  @override
  void didUpdateWidget(covariant _CategoryScrollViewport oldWidget) {
    super.didUpdateWidget(oldWidget);
    WidgetsBinding.instance.addPostFrameCallback((_) => _updateScrollHints());
  }

  @override
  void dispose() {
    _controller
      ..removeListener(_updateScrollHints)
      ..dispose();
    super.dispose();
  }

  void _updateScrollHints() {
    if (!mounted || !_controller.hasClients) {
      return;
    }
    final ScrollPosition position = _controller.position;
    final bool canScroll = position.maxScrollExtent > 1;
    final bool nextCanScrollUp = canScroll && position.pixels > 6;
    final bool nextCanScrollDown =
        canScroll && position.pixels < position.maxScrollExtent - 6;
    if (_canScrollUp == nextCanScrollUp &&
        _canScrollDown == nextCanScrollDown) {
      return;
    }
    setState(() {
      _canScrollUp = nextCanScrollUp;
      _canScrollDown = nextCanScrollDown;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: <Widget>[
        ListView.separated(
          controller: _controller,
          padding: EdgeInsets.zero,
          physics: const BouncingScrollPhysics(),
          itemCount: widget.items.length,
          separatorBuilder: (_, __) => const SizedBox(height: 5),
          itemBuilder: (BuildContext context, int index) {
            final _CategoryItemData data = widget.items[index];
            return _CategoryItem(
              data: data,
              onTap: () => widget.onCategoryTap(data.name),
            );
          },
        ),
        _CategoryScrollEdgeHint(
          alignment: Alignment.topCenter,
          visible: _canScrollUp,
          icon: Icons.keyboard_arrow_up_rounded,
          top: true,
        ),
        _CategoryScrollEdgeHint(
          alignment: Alignment.bottomCenter,
          visible: _canScrollDown,
          icon: Icons.keyboard_arrow_down_rounded,
          label: '继续下滑',
        ),
      ],
    );
  }
}

class _CategoryScrollEdgeHint extends StatelessWidget {
  const _CategoryScrollEdgeHint({
    required this.alignment,
    required this.visible,
    required this.icon,
    this.label = '',
    this.top = false,
  });

  final Alignment alignment;
  final bool visible;
  final IconData icon;
  final String label;
  final bool top;

  @override
  Widget build(BuildContext context) {
    return Positioned(
      left: 0,
      right: 0,
      top: top ? 0 : null,
      bottom: top ? null : 0,
      height: label.isEmpty ? 28 : 48,
      child: IgnorePointer(
        child: AnimatedOpacity(
          opacity: visible ? 1 : 0,
          duration: const Duration(milliseconds: 180),
          child: DecoratedBox(
            decoration: BoxDecoration(
              gradient: LinearGradient(
                begin: top ? Alignment.topCenter : Alignment.bottomCenter,
                end: top ? Alignment.bottomCenter : Alignment.topCenter,
                colors: const <Color>[
                  Color(0xFFFFFFFF),
                  Color(0x00FFFFFF),
                ],
              ),
            ),
            child: Align(
              alignment: alignment,
              child: label.isEmpty
                  ? Icon(icon, size: 20, color: _ScaleColors.orange)
                  : Container(
                      height: 28,
                      margin: const EdgeInsets.only(bottom: 2),
                      padding: const EdgeInsets.symmetric(horizontal: 10),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(.94),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: _ScaleColors.lineSoft),
                        boxShadow: _scaleShadow(
                          color: const Color(0x14B05F32),
                          blur: 8,
                          offset: const Offset(0, 3),
                        ),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: <Widget>[
                          Text(
                            label,
                            style: const TextStyle(
                              color: _ScaleColors.orangeDeep,
                              fontSize: 11,
                              height: 1,
                              fontWeight: FontWeight.w900,
                            ),
                          ),
                          const SizedBox(width: 2),
                          Icon(
                            icon,
                            size: 18,
                            color: _ScaleColors.orange,
                          ),
                        ],
                      ),
                    ),
            ),
          ),
        ),
      ),
    );
  }
}

class _CategoryItem extends StatelessWidget {
  const _CategoryItem({required this.data, required this.onTap});

  final _CategoryItemData data;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Ink(
          height: 46,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color: data.active ? const Color(0xFFFFF0E7) : Colors.transparent,
            borderRadius: BorderRadius.circular(14),
          ),
          child: Row(
            children: <Widget>[
              Container(
                width: 10,
                height: 10,
                decoration:
                    BoxDecoration(color: data.color, shape: BoxShape.circle),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  data.name,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: data.active
                        ? _ScaleColors.orangeDeep
                        : _ScaleColors.text,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Text(
                '${data.count}',
                style: TextStyle(
                  color: data.active
                      ? _ScaleColors.orangeDeep
                      : _ScaleColors.muted,
                  fontSize: 12,
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

class _DraftCard extends StatelessWidget {
  const _DraftCard({
    required this.count,
    required this.loading,
    required this.errorMessage,
    required this.onTap,
  });

  final int count;
  final bool loading;
  final String? errorMessage;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final String countText = loading
        ? '正在统计草稿...'
        : errorMessage != null
            ? '草稿统计失败，点击重试'
            : '当前共$count条草稿';
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(18),
        child: Ink(
          height: 118,
          padding: const EdgeInsets.fromLTRB(15, 14, 15, 13),
          decoration: BoxDecoration(
            color: Colors.white.withOpacity(.74),
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: _ScaleColors.line),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              const Row(
                children: <Widget>[
                  Icon(Icons.assignment_outlined,
                      size: 18, color: _ScaleColors.ink),
                  SizedBox(width: 8),
                  Text(
                    '继续草稿',
                    style: TextStyle(
                      color: _ScaleColors.ink,
                      fontSize: 16,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  Spacer(),
                  Icon(Icons.chevron_right_rounded,
                      size: 21, color: _ScaleColors.muted),
                ],
              ),
              const SizedBox(height: 9),
              const Text(
                '未完成的测评可直接恢复，支持断点续测！',
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  color: _ScaleColors.text,
                  fontSize: 12,
                  height: 1.45,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const Spacer(),
              Text(
                countText,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: TextStyle(
                  color: errorMessage == null
                      ? _ScaleColors.orange
                      : _ScaleColors.muted,
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

class _CategorySkeletonList extends StatelessWidget {
  const _CategorySkeletonList();

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      padding: EdgeInsets.zero,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: 8,
      separatorBuilder: (BuildContext context, int index) {
        return const SizedBox(height: 5);
      },
      itemBuilder: (BuildContext context, int index) {
        return _CategorySkeleton(
          key: ValueKey<String>('category-skeleton-$index'),
          index: index,
        );
      },
    );
  }
}

class _CategorySkeleton extends StatelessWidget {
  const _CategorySkeleton({
    super.key,
    required this.index,
  });

  final int index;

  @override
  Widget build(BuildContext context) {
    final bool active = index == 0;
    return Container(
      height: 46,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: active ? const Color(0xFFFFF0E7) : Colors.transparent,
        borderRadius: BorderRadius.circular(14),
      ),
      child: Row(
        children: <Widget>[
          Container(
            width: 10,
            height: 10,
            decoration: BoxDecoration(
              color: _categoryAccentColor(index).withOpacity(.72),
              shape: BoxShape.circle,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: _ScaleSkeletonBlock(
              widthFactor: active ? .52 : .7,
              height: 14,
            ),
          ),
          const SizedBox(width: 12),
          const _ScaleSkeletonBlock(width: 18, height: 12),
        ],
      ),
    );
  }
}

class _ScaleSidebarMessage extends StatelessWidget {
  const _ScaleSidebarMessage({
    required this.text,
    this.actionLabel = '',
    this.onAction,
  });

  final String text;
  final String actionLabel;
  final VoidCallback? onAction;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Text(
            text,
            maxLines: 3,
            textAlign: TextAlign.center,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: _ScaleColors.muted,
              fontSize: 13,
              height: 1.4,
              fontWeight: FontWeight.w800,
            ),
          ),
          if (onAction != null && actionLabel.trim().isNotEmpty) ...<Widget>[
            const SizedBox(height: 10),
            _MiniActionButton(label: actionLabel, onTap: onAction!),
          ],
        ],
      ),
    );
  }
}

class _MiniActionButton extends StatelessWidget {
  const _MiniActionButton({required this.label, required this.onTap});

  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 32,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: const Color(0xFFFFF1E8),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _ScaleColors.lineSoft),
          ),
          child: Center(
            child: Text(
              label,
              style: const TextStyle(
                color: _ScaleColors.orangeDeep,
                fontSize: 13,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _ScaleGridSkeleton extends StatelessWidget {
  const _ScaleGridSkeleton({
    required this.columns,
    required this.cardHeight,
  });

  final int columns;
  final double cardHeight;

  @override
  Widget build(BuildContext context) {
    return GridView.builder(
      padding: EdgeInsets.zero,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: columns * 2,
      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: columns,
        crossAxisSpacing: 14,
        mainAxisSpacing: 14,
        mainAxisExtent: cardHeight,
      ),
      itemBuilder: (BuildContext context, int index) {
        return _ScaleCardSkeleton(
          key: ValueKey<String>('scale-card-skeleton-$index'),
        );
      },
    );
  }
}

class _ScaleCardSkeleton extends StatelessWidget {
  const _ScaleCardSkeleton({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: _ScaleColors.card.withOpacity(.9),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _ScaleColors.line, width: 1.1),
        boxShadow: _scaleShadow(
          color: const Color(0x0DB05F32),
          blur: 18,
          offset: const Offset(0, 8),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: const <Widget>[
          Expanded(child: _ScaleCoverSkeleton()),
          SizedBox(height: 11),
          _ScaleSkeletonBlock(widthFactor: .78, height: 20),
          SizedBox(height: 10),
          Row(
            children: <Widget>[
              _ScaleSkeletonBlock(width: 54, height: 24, radius: 10),
              SizedBox(width: 8),
              _ScaleSkeletonBlock(width: 64, height: 24, radius: 10),
              SizedBox(width: 8),
              _ScaleSkeletonBlock(width: 58, height: 24, radius: 10),
            ],
          ),
          SizedBox(height: 12),
          _ScaleSkeletonBlock(height: 38, radius: 12),
        ],
      ),
    );
  }
}

class _ScaleCoverSkeleton extends StatelessWidget {
  const _ScaleCoverSkeleton();

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(16),
      child: ColoredBox(
        color: const Color(0xFFFFF1E8),
        child: Stack(
          children: const <Widget>[
            Positioned(
              left: 22,
              top: 22,
              child: _ScaleSkeletonBlock(width: 74, height: 74, radius: 99),
            ),
            Positioned(
              left: 114,
              top: 32,
              right: 24,
              child: _ScaleSkeletonBlock(height: 16, radius: 99),
            ),
            Positioned(
              left: 114,
              top: 60,
              right: 48,
              child: _ScaleSkeletonBlock(height: 14, radius: 99),
            ),
            Positioned(
              left: 114,
              top: 86,
              right: 72,
              child: _ScaleSkeletonBlock(height: 14, radius: 99),
            ),
          ],
        ),
      ),
    );
  }
}

class _ScaleSkeletonBlock extends StatelessWidget {
  const _ScaleSkeletonBlock({
    this.width,
    this.widthFactor,
    required this.height,
    this.radius = 99,
  });

  final double? width;
  final double? widthFactor;
  final double height;
  final double radius;

  @override
  Widget build(BuildContext context) {
    final Widget block = Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: const Color(0xFFF3E6DD),
        borderRadius: BorderRadius.circular(radius),
      ),
    );
    if (widthFactor == null) {
      return block;
    }
    return FractionallySizedBox(
      alignment: Alignment.centerLeft,
      widthFactor: widthFactor,
      child: block,
    );
  }
}

class _ScaleInlineLoading extends StatelessWidget {
  const _ScaleInlineLoading();

  @override
  Widget build(BuildContext context) {
    return const Center(child: _ScaleLoadingPill());
  }
}

class _ScaleLoadingOverlay extends StatelessWidget {
  const _ScaleLoadingOverlay();

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: Color(0x33FFF7EE),
      child: Align(
        alignment: Alignment.topRight,
        child: Padding(
          padding: const EdgeInsets.only(top: 8, right: 8),
          child: _ScaleLoadingPill(),
        ),
      ),
    );
  }
}

class _ScaleLoadingPill extends StatelessWidget {
  const _ScaleLoadingPill();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 36,
      padding: const EdgeInsets.symmetric(horizontal: 13),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.92),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _ScaleColors.lineSoft),
        boxShadow: _scaleShadow(
          color: const Color(0x10B05F32),
          blur: 10,
          offset: const Offset(0, 4),
        ),
      ),
      child: const Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          SizedBox(
            width: 14,
            height: 14,
            child: CircularProgressIndicator(
              strokeWidth: 2,
              color: _ScaleColors.orange,
            ),
          ),
          SizedBox(width: 8),
          Text(
            '正在加载量表',
            style: TextStyle(
              color: _ScaleColors.text,
              fontSize: 12,
              height: 1,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _ScaleErrorState extends StatelessWidget {
  const _ScaleErrorState({required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        width: 360,
        padding: const EdgeInsets.symmetric(horizontal: 22, vertical: 24),
        decoration: BoxDecoration(
          color: Colors.white.withOpacity(.76),
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: _ScaleColors.lineSoft),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            const Icon(Icons.wifi_off_rounded,
                color: _ScaleColors.muted, size: 34),
            const SizedBox(height: 10),
            Text(
              message,
              maxLines: 3,
              textAlign: TextAlign.center,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _ScaleColors.text,
                fontSize: 14,
                height: 1.4,
                fontWeight: FontWeight.w800,
              ),
            ),
            const SizedBox(height: 14),
            _MiniActionButton(label: '重新加载', onTap: onRetry),
          ],
        ),
      ),
    );
  }
}

class _DraftSheet extends StatelessWidget {
  const _DraftSheet({
    required this.drafts,
    required this.total,
    required this.loading,
    required this.errorMessage,
    required this.onRetry,
    required this.onOpenDraft,
  });

  final List<AssessmentDraftSummary> drafts;
  final int total;
  final bool loading;
  final String? errorMessage;
  final VoidCallback onRetry;
  final ValueChanged<AssessmentDraftSummary> onOpenDraft;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      top: false,
      child: Align(
        alignment: Alignment.bottomCenter,
        child: Container(
          width: 760,
          constraints: const BoxConstraints(maxHeight: 430),
          margin: const EdgeInsets.only(bottom: 18),
          padding: const EdgeInsets.fromLTRB(20, 18, 20, 20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(22),
            border: Border.all(color: _ScaleColors.line),
            boxShadow: _scaleShadow(
              color: const Color(0x14000000),
              blur: 18,
              offset: const Offset(0, 8),
            ),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Row(
                children: <Widget>[
                  const Icon(Icons.assignment_outlined,
                      size: 22, color: _ScaleColors.orange),
                  const SizedBox(width: 9),
                  const Text(
                    '继续草稿',
                    style: TextStyle(
                      color: _ScaleColors.ink,
                      fontSize: 18,
                      height: 1,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Text(
                    loading ? '统计中' : '共$total条',
                    style: const TextStyle(
                      color: _ScaleColors.muted,
                      fontSize: 13,
                      height: 1,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const Spacer(),
                  _SearchKeyboardAction(
                    label: '关闭',
                    icon: Icons.close_rounded,
                    onTap: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
              const SizedBox(height: 14),
              Flexible(child: _buildContent()),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildContent() {
    if (loading) {
      return const SizedBox(
        height: 148,
        child: Center(
          child: CircularProgressIndicator(
            strokeWidth: 3,
            color: _ScaleColors.orange,
          ),
        ),
      );
    }
    if (errorMessage != null) {
      return SizedBox(
        height: 148,
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Text(
                errorMessage!,
                maxLines: 2,
                textAlign: TextAlign.center,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _ScaleColors.text,
                  fontSize: 14,
                  height: 1.4,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 12),
              _MiniActionButton(label: '重试', onTap: onRetry),
            ],
          ),
        ),
      );
    }
    if (drafts.isEmpty) {
      return const SizedBox(
        height: 148,
        child: Center(
          child: Text(
            '当前没有未完成草稿',
            style: TextStyle(
              color: _ScaleColors.muted,
              fontSize: 15,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
      );
    }
    return ListView.separated(
      shrinkWrap: true,
      padding: EdgeInsets.zero,
      physics: const BouncingScrollPhysics(),
      itemCount: drafts.length,
      separatorBuilder: (_, __) => const SizedBox(height: 10),
      itemBuilder: (BuildContext context, int index) {
        final AssessmentDraftSummary draft = drafts[index];
        return _DraftSheetItem(
          draft: draft,
          onTap: () => onOpenDraft(draft),
        );
      },
    );
  }
}

class _DraftSheetItem extends StatelessWidget {
  const _DraftSheetItem({required this.draft, required this.onTap});

  final AssessmentDraftSummary draft;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final String student = draft.studentName.trim().isNotEmpty
        ? draft.studentName.trim()
        : '未选择学员';
    final String title = draft.assessmentName.trim().isNotEmpty
        ? draft.assessmentName.trim()
        : draft.assessmentCode.trim();
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Ink(
          height: 72,
          padding: const EdgeInsets.symmetric(horizontal: 15),
          decoration: BoxDecoration(
            color: const Color(0xFFFFFAF5),
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: _ScaleColors.lineSoft),
          ),
          child: Row(
            children: <Widget>[
              Container(
                width: 42,
                height: 42,
                alignment: Alignment.center,
                decoration: const BoxDecoration(
                  color: Color(0xFFFFE8DA),
                  shape: BoxShape.circle,
                ),
                child: Text(
                  '${draft.completionPercentInt}%',
                  style: const TextStyle(
                    color: _ScaleColors.orangeDeep,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const SizedBox(width: 13),
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text(
                      '$student · $title',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _ScaleColors.ink,
                        fontSize: 15,
                        height: 1,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      '已答${draft.answeredItemCount}题 · 更新${draft.displayUpdatedDate}',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _ScaleColors.muted,
                        fontSize: 12,
                        height: 1,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              const Icon(Icons.chevron_right_rounded,
                  size: 24, color: _ScaleColors.muted),
            ],
          ),
        ),
      ),
    );
  }
}

class _ScaleMainContent extends StatelessWidget {
  const _ScaleMainContent({
    required this.searchQuery,
    required this.categoryTitle,
    required this.scales,
    required this.summary,
    required this.loading,
    required this.initialLoading,
    required this.errorMessage,
    required this.hasSelectedStudent,
    required this.onChooseScale,
    required this.onRequireStudent,
    required this.onRetry,
  });

  final String searchQuery;
  final String categoryTitle;
  final List<AssessmentScaleItem> scales;
  final AssessmentScaleLibrarySummary summary;
  final bool loading;
  final bool initialLoading;
  final String? errorMessage;
  final bool hasSelectedStudent;
  final ValueChanged<AssessmentScaleItem> onChooseScale;
  final ValueChanged<AssessmentScaleItem> onRequireStudent;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        const double gap = 14;
        const double toolbarHeight = 44;
        const double toolbarGap = 10;
        final int columns = constraints.maxWidth < 760 ? 2 : 3;
        final double cardHeight =
            ((constraints.maxHeight - toolbarHeight - gap - toolbarGap) / 2)
                .clamp(252.0, 292.0)
                .toDouble();

        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            _ScaleToolbar(
              title: categoryTitle,
              searchQuery: searchQuery,
              visibleCount: scales.length,
              summary: summary,
              loading: loading,
            ),
            const SizedBox(height: toolbarGap),
            if (errorMessage != null)
              Expanded(
                child: _ScaleErrorState(
                  message: errorMessage!,
                  onRetry: onRetry,
                ),
              )
            else if (initialLoading)
              Expanded(
                child: _ScaleGridSkeleton(
                  columns: columns,
                  cardHeight: cardHeight,
                ),
              )
            else if (scales.isEmpty)
              Expanded(
                child: loading
                    ? const _ScaleInlineLoading()
                    : _ScaleSearchEmpty(searchQuery: searchQuery),
              )
            else
              Expanded(
                child: Stack(
                  children: <Widget>[
                    GridView.builder(
                      padding: EdgeInsets.zero,
                      physics: const BouncingScrollPhysics(),
                      itemCount: scales.length,
                      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                        crossAxisCount: columns,
                        crossAxisSpacing: gap,
                        mainAxisSpacing: gap,
                        mainAxisExtent: cardHeight,
                      ),
                      itemBuilder: (BuildContext context, int index) {
                        return _ScaleCard(
                          data: scales[index],
                          enabled:
                              hasSelectedStudent && scales[index].available,
                          onChoose: () {
                            if (hasSelectedStudent) {
                              onChooseScale(scales[index]);
                              return;
                            }
                            onRequireStudent(scales[index]);
                          },
                          canRequestStudent:
                              !hasSelectedStudent && scales[index].available,
                        );
                      },
                    ),
                    if (loading)
                      const Positioned.fill(
                        child: AbsorbPointer(child: _ScaleLoadingOverlay()),
                      ),
                  ],
                ),
              ),
          ],
        );
      },
    );
  }
}

class _ScaleToolbar extends StatelessWidget {
  const _ScaleToolbar({
    required this.title,
    required this.searchQuery,
    required this.visibleCount,
    required this.summary,
    required this.loading,
  });

  final String title;
  final String searchQuery;
  final int visibleCount;
  final AssessmentScaleLibrarySummary summary;
  final bool loading;

  @override
  Widget build(BuildContext context) {
    final String statsText = loading
        ? '正在加载量表'
        : searchQuery.trim().isNotEmpty
            ? '$visibleCount 个匹配结果'
            : '${summary.available} 个可用，${summary.unavailable} 个暂不可用';
    return SizedBox(
      height: 44,
      child: Transform.translate(
        offset: const Offset(0, 0),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            Expanded(
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: <Widget>[
                  Flexible(
                    child: Text(
                      title,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _ScaleColors.ink,
                        fontSize: 27,
                        height: 1,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  const SizedBox(width: 17),
                  Flexible(
                    child: Text(
                      statsText,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _ScaleColors.muted,
                        fontSize: 14,
                        height: 1,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            Container(
              height: 42,
              padding: const EdgeInsets.all(4),
              decoration: BoxDecoration(
                color: const Color(0xFFFFF1E8).withOpacity(.78),
                borderRadius: BorderRadius.circular(15),
              ),
              child: const Row(
                children: <Widget>[
                  _Segment(label: '可用', active: true),
                  _Segment(label: '全部'),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _Segment extends StatelessWidget {
  const _Segment({required this.label, this.active = false});

  final String label;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 90,
      height: 34,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: active ? Colors.white : Colors.transparent,
        borderRadius: BorderRadius.circular(12),
        boxShadow: active
            ? _scaleShadow(
                color: const Color(0x12B05F32),
                blur: 12,
                offset: const Offset(0, 5),
              )
            : null,
      ),
      child: Text(
        label,
        strutStyle: const StrutStyle(
          fontSize: 14,
          height: 1,
          forceStrutHeight: true,
        ),
        style: TextStyle(
          color: active ? _ScaleColors.orangeDeep : _ScaleColors.text,
          fontSize: 14,
          height: 1,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

String _normalizeSearchText(String value) {
  return value.trim().toLowerCase().replaceAll(RegExp(r'[\s\-/_.]'), '');
}

List<String> _mergeCategories(List<String> primary, List<String> secondary) {
  final List<String> out = <String>[];
  final Set<String> seen = <String>{};
  for (final String item in <String>[...primary, ...secondary]) {
    final String normalized = item.trim();
    if (normalized.isEmpty || seen.contains(normalized)) {
      continue;
    }
    seen.add(normalized);
    out.add(normalized);
  }
  return out;
}

bool _sameStringList(List<String> a, List<String> b) {
  if (a.length != b.length) {
    return false;
  }
  for (int index = 0; index < a.length; index++) {
    if (a[index] != b[index]) {
      return false;
    }
  }
  return true;
}

Color _categoryAccentColor(int index) {
  const List<Color> colors = <Color>[
    Color(0xFFE96F43),
    Color(0xFF3F82D2),
    Color(0xFF6F9F70),
    Color(0xFFD99427),
    Color(0xFF63A999),
    Color(0xFFD96A7F),
    Color(0xFF7F77C8),
  ];
  return colors[index % colors.length];
}

_CoverType _coverTypeForScale(AssessmentScaleItem item) {
  final String target = _normalizeSearchText(
    <String>[item.name, item.code, item.category, item.scenario].join(' '),
  );
  if (target.contains('pep') || target.contains('book')) {
    return _CoverType.book;
  }
  if (target.contains('社交') ||
      target.contains('social') ||
      target.contains('情绪')) {
    return _CoverType.social;
  }
  if (target.contains('筛查') || target.contains('screen')) {
    return _CoverType.screen;
  }
  if (target.contains('表达') || target.contains('沟通')) {
    return _CoverType.express;
  }
  if (target.contains('口语') || target.contains('语言')) {
    return _CoverType.talk;
  }
  return _CoverType.review;
}

class _ScaleSearchEmpty extends StatelessWidget {
  const _ScaleSearchEmpty({required this.searchQuery});

  final String searchQuery;

  @override
  Widget build(BuildContext context) {
    final bool searching = searchQuery.trim().isNotEmpty;
    return Center(
      child: Container(
        width: 320,
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 22),
        decoration: BoxDecoration(
          color: Colors.white.withOpacity(.72),
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: _ScaleColors.lineSoft),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Icon(
              searching ? Icons.search_off_rounded : Icons.inventory_2_outlined,
              color: _ScaleColors.muted,
              size: 34,
            ),
            const SizedBox(height: 10),
            Text(
              searching ? '没有匹配的量表' : '当前分类暂无量表',
              style: const TextStyle(
                color: _ScaleColors.ink,
                fontSize: 17,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              searching ? '可尝试更换关键词或切换分类' : '可切换左侧分类查看其它量表',
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _ScaleColors.muted,
                fontSize: 13,
                height: 1.35,
                fontWeight: FontWeight.w700,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ScaleCard extends StatelessWidget {
  const _ScaleCard({
    required this.data,
    required this.enabled,
    required this.onChoose,
    required this.canRequestStudent,
  });

  final AssessmentScaleItem data;
  final bool enabled;
  final VoidCallback onChoose;
  final bool canRequestStudent;

  @override
  Widget build(BuildContext context) {
    final List<String> tags = data.tags;
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: _ScaleColors.card.withOpacity(.96),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _ScaleColors.line, width: 1.1),
        boxShadow: _scaleShadow(),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Expanded(
            child: ClipRRect(
              borderRadius: BorderRadius.circular(16),
              child: CustomPaint(
                painter: _ScaleCoverPainter(type: _coverTypeForScale(data)),
                child: const SizedBox.expand(),
              ),
            ),
          ),
          const SizedBox(height: 11),
          Row(
            children: <Widget>[
              Expanded(
                child: Text(
                  data.name,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _ScaleColors.ink,
                    fontSize: 20,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: <Widget>[
              for (int i = 0; i < tags.length; i++) ...<Widget>[
                _InfoTag(label: tags[i]),
                if (i != tags.length - 1) const SizedBox(width: 6),
              ],
            ],
          ),
          const SizedBox(height: 10),
          _ChooseButton(
            enabled: enabled,
            feedbackEnabled: canRequestStudent,
            onTap: onChoose,
          ),
        ],
      ),
    );
  }
}

class _InfoTag extends StatelessWidget {
  const _InfoTag({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF8F1),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _ScaleColors.lineSoft),
      ),
      child: Text(
        label,
        maxLines: 1,
        style: const TextStyle(
          color: _ScaleColors.text,
          fontSize: 12,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _ChooseButton extends StatelessWidget {
  const _ChooseButton({
    required this.enabled,
    required this.feedbackEnabled,
    required this.onTap,
  });

  final bool enabled;
  final bool feedbackEnabled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final bool tappable = enabled || feedbackEnabled;
    final Color borderColor =
        enabled ? _ScaleColors.orange : const Color(0xFFE2D6CE);
    final Color textColor =
        enabled ? _ScaleColors.orangeDeep : _ScaleColors.muted;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: tappable ? onTap : null,
        borderRadius: BorderRadius.circular(12),
        child: Ink(
          height: 38,
          width: double.infinity,
          decoration: BoxDecoration(
            color: enabled ? Colors.white : const Color(0xFFF7F1ED),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: borderColor, width: 1.2),
          ),
          child: Center(
            child: Text(
              '开始测评',
              maxLines: 1,
              textAlign: TextAlign.center,
              strutStyle: const StrutStyle(
                fontSize: 15,
                height: 1,
                forceStrutHeight: true,
              ),
              style: TextStyle(
                color: textColor,
                fontSize: 15,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

enum _CoverType { book, talk, screen, express, social, review }

class _ScaleCoverPainter extends CustomPainter {
  const _ScaleCoverPainter({required this.type});

  final _CoverType type;

  @override
  void paint(Canvas canvas, Size size) {
    final Rect rect = Offset.zero & size;
    final Paint bg = Paint()
      ..shader = LinearGradient(
        colors: _backgroundColors(type),
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
      ).createShader(rect);
    canvas.drawRect(rect, bg);

    final Paint soft = Paint()..color = Colors.white.withOpacity(.34);
    canvas.drawCircle(Offset(size.width * .9, size.height * .05), 70, soft);
    canvas.drawCircle(Offset(size.width * .08, size.height * .95), 58, soft);

    switch (type) {
      case _CoverType.book:
        _drawBook(canvas, size);
      case _CoverType.talk:
        _drawTalk(canvas, size);
      case _CoverType.screen:
        _drawScreen(canvas, size);
      case _CoverType.express:
        _drawExpress(canvas, size);
      case _CoverType.social:
        _drawSocial(canvas, size);
      case _CoverType.review:
        _drawReview(canvas, size);
    }
  }

  List<Color> _backgroundColors(_CoverType type) {
    switch (type) {
      case _CoverType.book:
        return const <Color>[Color(0xFFFFF3E4), Color(0xFFFFD8BC)];
      case _CoverType.talk:
        return const <Color>[Color(0xFFF5F7EA), Color(0xFFDDEBD2)];
      case _CoverType.screen:
        return const <Color>[Color(0xFFF1F7FF), Color(0xFFD5E8F9)];
      case _CoverType.express:
        return const <Color>[Color(0xFFFFF6E1), Color(0xFFFFDFA7)];
      case _CoverType.social:
        return const <Color>[Color(0xFFF6F3EA), Color(0xFFDDECCF)];
      case _CoverType.review:
        return const <Color>[Color(0xFFFFF2EA), Color(0xFFF8D5C9)];
    }
  }

  void _drawBook(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    _drawChild(canvas, Offset(w * .28, h * .54),
        shirt: const Color(0xFF7FA1B5));
    final Paint page = Paint()..color = Colors.white.withOpacity(.86);
    final RRect left = RRect.fromRectAndRadius(
      Rect.fromLTWH(w * .34, h * .42, w * .2, h * .26),
      const Radius.circular(10),
    );
    final RRect right = RRect.fromRectAndRadius(
      Rect.fromLTWH(w * .53, h * .42, w * .2, h * .26),
      const Radius.circular(10),
    );
    canvas.drawRRect(left, page);
    canvas.drawRRect(right, page);
    final Paint line = Paint()
      ..color = const Color(0xFFE6A16B)
      ..strokeWidth = 4
      ..strokeCap = StrokeCap.round;
    canvas.drawLine(Offset(w * .535, h * .45), Offset(w * .535, h * .66), line);
    canvas.drawLine(Offset(w * .39, h * .5), Offset(w * .49, h * .5), line);
    canvas.drawLine(Offset(w * .58, h * .5), Offset(w * .68, h * .5), line);
    _drawSpeechBubble(canvas, Offset(w * .72, h * .26), const Color(0xFFFFFFFF),
        const Color(0xFFE96F43));
  }

  void _drawTalk(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    final Paint bridge = Paint()
      ..color = Colors.white.withOpacity(.42)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 5
      ..strokeCap = StrokeCap.round;
    canvas.drawPath(
      Path()
        ..moveTo(w * .35, h * .78)
        ..cubicTo(w * .43, h * .7, w * .57, h * .7, w * .65, h * .78),
      bridge,
    );
    _drawChild(canvas, Offset(w * .34, h * .6), shirt: const Color(0xFF8FB279));
    _drawChild(canvas, Offset(w * .66, h * .6), shirt: const Color(0xFFE5A552));
    _drawCompactSpeechBubble(
      canvas,
      Offset(w * .27, h * .18),
      const Color(0xFFFFFFFF),
      const Color(0xFF6F9F70),
    );
    _drawCompactSpeechBubble(
      canvas,
      Offset(w * .73, h * .18),
      const Color(0xFFFFFFFF),
      const Color(0xFF6F9F70),
    );
  }

  void _drawScreen(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    final Paint board = Paint()..color = Colors.white.withOpacity(.9);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(w * .27, h * .16, w * .34, h * .64),
        const Radius.circular(16),
      ),
      board,
    );
    final Paint clip = Paint()..color = const Color(0xFFD9B27B);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(w * .36, h * .1, w * .16, 22),
        const Radius.circular(9),
      ),
      clip,
    );
    final Paint check = Paint()
      ..color = const Color(0xFF6F9F70)
      ..strokeWidth = 5
      ..strokeCap = StrokeCap.round
      ..style = PaintingStyle.stroke;
    for (final double y in <double>[.32, .48, .64]) {
      canvas.drawPath(
        Path()
          ..moveTo(w * .34, h * y)
          ..lineTo(w * .38, h * (y + .04))
          ..lineTo(w * .47, h * (y - .06)),
        check,
      );
      canvas.drawLine(Offset(w * .5, h * y), Offset(w * .57, h * y), check);
    }
    final Paint lens = Paint()
      ..color = const Color(0x803F82D2)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 10;
    canvas.drawCircle(Offset(w * .69, h * .48), 36, lens);
    canvas.drawLine(
      Offset(w * .72, h * .58),
      Offset(w * .82, h * .72),
      Paint()
        ..color = const Color(0xFF3F82D2)
        ..strokeWidth = 10
        ..strokeCap = StrokeCap.round,
    );
  }

  void _drawExpress(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    final Paint paper = Paint()..color = Colors.white.withOpacity(.86);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(w * .35, h * .16, w * .32, h * .66),
        const Radius.circular(15),
      ),
      paper,
    );
    final Paint pencil = Paint()..color = const Color(0xFFE6A13D);
    canvas.save();
    canvas.translate(w * .58, h * .51);
    canvas.rotate(-.72);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(-12, -54, 24, 108),
        const Radius.circular(9),
      ),
      pencil,
    );
    canvas.drawPath(
      Path()
        ..moveTo(-12, 54)
        ..lineTo(12, 54)
        ..lineTo(0, 73)
        ..close(),
      Paint()..color = const Color(0xFF8D5B36),
    );
    canvas.restore();
    _drawSpeechBubble(canvas, Offset(w * .24, h * .55), Colors.white,
        const Color(0xFFE96F43));
  }

  void _drawSocial(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    _drawChild(canvas, Offset(w * .36, h * .46),
        shirt: const Color(0xFF7FA1B5));
    _drawChild(canvas, Offset(w * .62, h * .47),
        shirt: const Color(0xFFE5A17A));
    final List<Color> colors = <Color>[
      const Color(0xFFE96F43),
      const Color(0xFFF6C45F),
      const Color(0xFF6F9F70),
      const Color(0xFF3F82D2),
    ];
    for (int i = 0; i < 6; i++) {
      final double x = w * (.32 + (i % 3) * .13);
      final double y = h * (.72 - (i ~/ 3) * .11);
      canvas.drawRRect(
        RRect.fromRectAndRadius(
          Rect.fromLTWH(x, y, 35, 30),
          const Radius.circular(7),
        ),
        Paint()..color = colors[i % colors.length].withOpacity(.88),
      );
    }
  }

  void _drawReview(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    final Paint sheet = Paint()..color = Colors.white.withOpacity(.88);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(w * .25, h * .18, w * .44, h * .58),
        const Radius.circular(16),
      ),
      sheet,
    );
    final Paint line = Paint()
      ..color = const Color(0xFF7FA1B5)
      ..strokeWidth = 4
      ..strokeCap = StrokeCap.round
      ..style = PaintingStyle.stroke;
    canvas.drawPath(
      Path()
        ..moveTo(w * .32, h * .58)
        ..lineTo(w * .42, h * .48)
        ..lineTo(w * .5, h * .54)
        ..lineTo(w * .62, h * .35),
      line,
    );
    for (final double x in <double>[.32, .42, .5, .62]) {
      canvas.drawCircle(Offset(w * x, x == .62 ? h * .35 : h * .52), 5,
          Paint()..color = const Color(0xFF7FA1B5));
    }
    final Offset medal = Offset(w * .72, h * .62);
    canvas.drawCircle(medal, 31, Paint()..color = const Color(0xFFE6A13D));
    canvas.drawCircle(medal, 21, Paint()..color = const Color(0xFFFFDFA7));
    canvas.drawPath(
      Path()
        ..moveTo(medal.dx, medal.dy - 13)
        ..lineTo(medal.dx + 5, medal.dy - 2)
        ..lineTo(medal.dx + 17, medal.dy - 1)
        ..lineTo(medal.dx + 8, medal.dy + 6)
        ..lineTo(medal.dx + 11, medal.dy + 18)
        ..lineTo(medal.dx, medal.dy + 11)
        ..lineTo(medal.dx - 11, medal.dy + 18)
        ..lineTo(medal.dx - 8, medal.dy + 6)
        ..lineTo(medal.dx - 17, medal.dy - 1)
        ..lineTo(medal.dx - 5, medal.dy - 2)
        ..close(),
      Paint()..color = const Color(0xFFE6A13D),
    );
  }

  void _drawChild(Canvas canvas, Offset center, {required Color shirt}) {
    final Paint skin = Paint()..color = const Color(0xFFFFC79A);
    final Paint hair = Paint()..color = const Color(0xFF5E3C2A);
    canvas.drawCircle(center.translate(0, -28), 26, skin);
    canvas.drawArc(
      Rect.fromCircle(center: center.translate(0, -34), radius: 27),
      math.pi,
      math.pi,
      true,
      hair,
    );
    canvas.drawCircle(
        center.translate(-8, -31), 3, Paint()..color = Colors.black);
    canvas.drawCircle(
        center.translate(9, -31), 3, Paint()..color = Colors.black);
    canvas.drawArc(
      Rect.fromCenter(center: center.translate(0, -21), width: 15, height: 10),
      0,
      math.pi,
      false,
      Paint()
        ..color = const Color(0xFFB24A37)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 2.4,
    );
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromCenter(center: center.translate(0, 18), width: 62, height: 64),
        const Radius.circular(24),
      ),
      Paint()..color = shirt,
    );
  }

  void _drawSpeechBubble(
    Canvas canvas,
    Offset center,
    Color fill,
    Color dotColor,
  ) {
    final RRect bubble = RRect.fromRectAndRadius(
      Rect.fromCenter(center: center, width: 84, height: 54),
      const Radius.circular(24),
    );
    canvas.drawRRect(bubble, Paint()..color = fill.withOpacity(.9));
    canvas.drawPath(
      Path()
        ..moveTo(center.dx - 20, center.dy + 20)
        ..lineTo(center.dx - 34, center.dy + 38)
        ..lineTo(center.dx - 10, center.dy + 24)
        ..close(),
      Paint()..color = fill.withOpacity(.9),
    );
    for (final double dx in <double>[-17, 0, 17]) {
      canvas.drawCircle(
        Offset(center.dx + dx, center.dy),
        5,
        Paint()..color = dotColor.withOpacity(.72),
      );
    }
  }

  void _drawCompactSpeechBubble(
    Canvas canvas,
    Offset center,
    Color fill,
    Color dotColor,
  ) {
    final Paint bubblePaint = Paint()..color = fill.withOpacity(.9);
    final RRect bubble = RRect.fromRectAndRadius(
      Rect.fromCenter(center: center, width: 72, height: 40),
      const Radius.circular(18),
    );
    canvas.drawRRect(bubble, bubblePaint);
    for (final double dx in <double>[-13, 0, 13]) {
      canvas.drawCircle(
        Offset(center.dx + dx, center.dy),
        4.2,
        Paint()..color = dotColor.withOpacity(.72),
      );
    }
  }

  @override
  bool shouldRepaint(covariant _ScaleCoverPainter oldDelegate) {
    return oldDelegate.type != type;
  }
}
