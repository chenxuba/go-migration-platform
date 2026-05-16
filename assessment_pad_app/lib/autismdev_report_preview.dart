part of 'assessment_report_list_page.dart';

enum _AutismDevReportTab {
  assessmentInfo,
  resultAnalysis,
  strengthWeakness,
  training,
  developmentProfile,
  behaviorProfile,
}

const List<_AutismDevReportTabSpec> _autismDevReportTabs =
    <_AutismDevReportTabSpec>[
  _AutismDevReportTabSpec('评估情况', _AutismDevReportTab.assessmentInfo),
  _AutismDevReportTabSpec('评估结果分析', _AutismDevReportTab.resultAnalysis),
  _AutismDevReportTabSpec('优劣势分析', _AutismDevReportTab.strengthWeakness),
  _AutismDevReportTabSpec('训练效果', _AutismDevReportTab.training),
  _AutismDevReportTabSpec('发展情况剖面图', _AutismDevReportTab.developmentProfile),
  _AutismDevReportTabSpec('情绪行为表现图', _AutismDevReportTab.behaviorProfile),
];

const String _autismDevDevelopmentProfileAsset =
    'assets/reports/autismdev_development_profile.png';
const String _autismDevBehaviorProfileAsset =
    'assets/reports/autismdev_behavior_profile.png';

class _AutismDevReportPreviewDialog extends StatefulWidget {
  const _AutismDevReportPreviewDialog({
    required this.record,
    required this.token,
    required this.client,
  });

  final Pep3RecordSummary record;
  final String token;
  final Pep3AssessmentClient client;

  @override
  State<_AutismDevReportPreviewDialog> createState() =>
      _AutismDevReportPreviewDialogState();
}

class _AutismDevReportPreviewDialogState
    extends State<_AutismDevReportPreviewDialog> {
  _AutismDevReportTab _activeTab = _AutismDevReportTab.assessmentInfo;
  late Pep3RecordSummary _displayRecord;
  final GlobalKey _printContentKey = GlobalKey();
  final Map<_AutismDevReportTab, Future<Uint8List>> _profilePdfFutures =
      <_AutismDevReportTab, Future<Uint8List>>{};
  final Map<_AutismDevReportTab, Uint8List> _profilePdfBytes =
      <_AutismDevReportTab, Uint8List>{};
  bool _printing = false;
  AutismDevResultAnalysis _resultAnalysis = _emptyAutismDevResultAnalysis();
  bool _resultAnalysisGenerating = false;
  String _resultAnalysisGenerationStatus = '';
  String _resultAnalysisGenerationError = '';
  String _resultAnalysisStreamText = '';
  int _resultAnalysisGenerateSerial = 0;
  int _resultAnalysisLoadSerial = 0;
  _AutismDevAnalysisEditRequest? _selectedAnalysisCell;

  @override
  void initState() {
    super.initState();
    _displayRecord = widget.record;
    unawaited(_loadRecordDetail());
    unawaited(_loadSavedResultAnalysis());
  }

  Future<void> _loadRecordDetail() async {
    final String token = widget.token.trim();
    if (token.isEmpty || widget.record.id <= 0) {
      return;
    }
    try {
      final Pep3RecordDetail detail = await widget.client.fetchRecordDetail(
        token,
        widget.record.id,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _displayRecord = detail;
      });
    } catch (_) {
      // The preview still has static report sections; keep them visible if
      // the detail endpoint is temporarily unavailable.
    }
  }

  Future<void> _loadSavedResultAnalysis() async {
    final String token = widget.token.trim();
    if (token.isEmpty || widget.record.id <= 0) {
      return;
    }
    final int serial = ++_resultAnalysisLoadSerial;
    try {
      final AutismDevResultAnalysis saved =
          await widget.client.fetchAutismDevResultAnalysis(
        token,
        widget.record.id,
      );
      if (!mounted || serial != _resultAnalysisLoadSerial) {
        return;
      }
      if (saved.isEmpty ||
          _resultAnalysisGenerating ||
          !_resultAnalysis.isEmpty) {
        return;
      }
      setState(() {
        _resultAnalysis = _mergeAutismDevResultAnalysis(saved);
        _resultAnalysisGenerationError = '';
      });
    } catch (_) {
      // Keep the page usable even if the cache lookup fails.
    }
  }

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      child: Center(
        child: Material(
          color: Colors.transparent,
          child: Container(
            width: 980,
            height: 654,
            padding: const EdgeInsets.fromLTRB(22, 22, 22, 22),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(color: _ReportTheme.line),
              boxShadow: const <BoxShadow>[
                BoxShadow(
                  color: Color(0x24000000),
                  blurRadius: 34,
                  offset: Offset(0, 18),
                ),
              ],
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                _buildHeader(context),
                const SizedBox(height: 14),
                _buildTabBar(),
                const SizedBox(height: 12),
                Expanded(child: _buildContent()),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildHeader(BuildContext context) {
    final Pep3RecordSummary record = _displayRecord;
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              const Text(
                '孤独症儿童发展评估报告',
                style: TextStyle(
                  color: _ReportTheme.ink,
                  fontSize: 24,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 7),
              Text(
                '${record.assessmentName.trim().isEmpty ? '孤独症儿童发展评估表' : record.assessmentName}   ${_studentName(record)} / ${_dateOnlyText(record.assessmentDate)}',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _ReportTheme.muted,
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(width: 16),
        Material(
          color: Colors.transparent,
          shape: const CircleBorder(),
          child: InkWell(
            onTap: () => Navigator.of(context).pop(),
            customBorder: const CircleBorder(),
            child: Container(
              width: 38,
              height: 38,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: const Color(0xFFFFF8F2),
                shape: BoxShape.circle,
                border: Border.all(color: _ReportTheme.lineSoft),
              ),
              child: const Icon(
                Icons.close_rounded,
                size: 22,
                color: _ReportTheme.muted,
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildTabBar() {
    return _AutismDevReportTabBar(
      activeTab: _activeTab,
      onSelected: _selectTab,
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          if (_activeTab == _AutismDevReportTab.resultAnalysis) ...<Widget>[
            _ToolbarButton(
              label: _resultAnalysisGenerating ? '生成中' : 'AI生成',
              icon: Icons.auto_awesome_rounded,
              filled: true,
              onTap: _resultAnalysisGenerating
                  ? null
                  : () => unawaited(_generateResultAnalysis()),
            ),
            const SizedBox(width: 8),
          ],
          _ToolbarButton(
            label: _printing ? '打印中' : '打印',
            icon: Icons.print_rounded,
            onTap: _printing ? null : () => unawaited(_printCurrentTab()),
          ),
        ],
      ),
    );
  }

  void _selectTab(_AutismDevReportTab tab) {
    if (_activeTab == tab) {
      return;
    }
    setState(() {
      _activeTab = tab;
      if (tab != _AutismDevReportTab.resultAnalysis) {
        _selectedAnalysisCell = null;
      }
    });
  }

  Future<void> _generateResultAnalysis() async {
    if (_resultAnalysisGenerating) {
      return;
    }
    final String token = widget.token.trim();
    if (token.isEmpty || widget.record.id <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('缺少AI生成参数')),
      );
      return;
    }
    final int serial = ++_resultAnalysisGenerateSerial;
    setState(() {
      _resultAnalysisGenerating = true;
      _resultAnalysisGenerationStatus = '正在生成评估结果分析';
      _resultAnalysisGenerationError = '';
      _resultAnalysisStreamText = '';
      _selectedAnalysisCell = null;
      _resultAnalysis = _emptyAutismDevResultAnalysis();
    });
    try {
      await for (final AutismDevResultAnalysisStreamEvent event in widget.client
          .generateAutismDevResultAnalysisStream(token, widget.record.id)) {
        if (!mounted || serial != _resultAnalysisGenerateSerial) {
          return;
        }
        switch (event.type) {
          case 'status':
            setState(() {
              _resultAnalysisGenerationStatus =
                  event.message.trim().isEmpty ? '正在生成评估结果分析' : event.message;
            });
          case 'delta':
            await _appendResultAnalysisDeltaWithTypewriter(event.text, serial);
            if (!mounted || serial != _resultAnalysisGenerateSerial) {
              return;
            }
            setState(() {
              _resultAnalysisGenerationStatus = 'AI正在生成评估结果分析';
            });
          case 'done':
            setState(() {
              _resultAnalysis = _mergeAutismDevResultAnalysis(event.data);
              _resultAnalysisGenerating = false;
              _resultAnalysisGenerationStatus = '';
              _resultAnalysisGenerationError = '';
            });
          case 'error':
            throw Pep3ApiException(
              event.message.trim().isEmpty ? '评估结果分析生成失败' : event.message,
            );
          default:
            break;
        }
      }
      if (mounted && serial == _resultAnalysisGenerateSerial) {
        setState(() {
          _resultAnalysisGenerating = false;
          _resultAnalysisGenerationStatus = '';
        });
      }
    } catch (error) {
      if (!mounted || serial != _resultAnalysisGenerateSerial) {
        return;
      }
      setState(() {
        _resultAnalysisGenerating = false;
        _resultAnalysisGenerationStatus = '';
        _resultAnalysisGenerationError = '$error';
      });
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('AI生成失败：$error')),
      );
    }
  }

  Future<void> _appendResultAnalysisDeltaWithTypewriter(
    String delta,
    int serial,
  ) async {
    if (delta.isEmpty) {
      return;
    }
    for (final int codePoint in delta.runes) {
      if (!mounted || serial != _resultAnalysisGenerateSerial) {
        return;
      }
      setState(() {
        _resultAnalysisStreamText += String.fromCharCode(codePoint);
      });
      await Future<void>.delayed(const Duration(milliseconds: 4));
    }
  }

  Widget _buildContent() {
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFFDF8F3),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      clipBehavior: Clip.antiAlias,
      child: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(18, 16, 18, 18),
        child: Center(
          child: RepaintBoundary(
            key: _printContentKey,
            child: ConstrainedBox(
              constraints: const BoxConstraints(maxWidth: 880),
              child: DecoratedBox(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: _ReportTheme.lineSoft),
                  boxShadow: const <BoxShadow>[
                    BoxShadow(
                      color: Color(0x0F000000),
                      blurRadius: 16,
                      offset: Offset(0, 8),
                    ),
                  ],
                ),
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(24, 22, 24, 24),
                  child: _buildReportPage(),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _printCurrentTab() async {
    if (_printing) {
      return;
    }
    setState(() {
      _printing = true;
    });
    try {
      if (_isBackendProfileTab(_activeTab)) {
        final Uint8List bytes = await _loadProfilePdf(_activeTab);
        if (bytes.isEmpty) {
          throw StateError('暂无可打印内容');
        }
        await Printing.layoutPdf(
          name:
              '孤独症儿童发展评估报告-${_autismDevReportTabs.firstWhere((e) => e.tab == _activeTab).label}.pdf',
          onLayout: (_) async => bytes,
        );
        return;
      }
      final Uint8List imageBytes = await _capturePrintContent();
      final pw.Document document = pw.Document();
      final pw.MemoryImage image = pw.MemoryImage(imageBytes);
      document.addPage(
        pw.Page(
          pageFormat: PdfPageFormat.a4,
          margin: const pw.EdgeInsets.all(24),
          build: (pw.Context context) {
            return pw.Center(
              child: pw.Image(image, fit: pw.BoxFit.contain),
            );
          },
        ),
      );
      await Printing.layoutPdf(
        name:
            '孤独症儿童发展评估报告-${_autismDevReportTabs.firstWhere((e) => e.tab == _activeTab).label}.pdf',
        onLayout: (_) async => document.save(),
      );
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('打印失败：$error')),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _printing = false;
        });
      }
    }
  }

  Future<Uint8List> _capturePrintContent() async {
    await WidgetsBinding.instance.endOfFrame;
    final BuildContext? boundaryContext = _printContentKey.currentContext;
    final RenderRepaintBoundary? boundary =
        boundaryContext?.findRenderObject() as RenderRepaintBoundary?;
    if (boundary == null) {
      throw StateError('暂无可打印内容');
    }
    final ui.Image image = await boundary.toImage(pixelRatio: 2.5);
    final ByteData? byteData =
        await image.toByteData(format: ui.ImageByteFormat.png);
    image.dispose();
    if (byteData == null) {
      throw StateError('打印内容生成失败');
    }
    return byteData.buffer.asUint8List();
  }

  Widget _buildReportPage() {
    final bool showHeader =
        !_AutismDevReportPreviewDialogState._isBackendProfileTab(_activeTab) &&
            _activeTab != _AutismDevReportTab.resultAnalysis;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        if (showHeader) ...<Widget>[
          _AutismDevReportPageHeader(record: _displayRecord),
          const SizedBox(height: 18),
        ],
        _buildActiveSection(),
      ],
    );
  }

  Widget _buildActiveSection() {
    switch (_activeTab) {
      case _AutismDevReportTab.assessmentInfo:
        return const _AutismDevOverviewSection();
      case _AutismDevReportTab.resultAnalysis:
        return _AutismDevAnalysisSection(
          record: _displayRecord,
          analysis: _resultAnalysis,
          generating: _resultAnalysisGenerating,
          generationStatus: _resultAnalysisGenerationStatus,
          generationError: _resultAnalysisGenerationError,
          streamText: _resultAnalysisStreamText,
          selectedCell: _selectedAnalysisCell,
          onCellTap: _handleResultAnalysisCellTap,
        );
      case _AutismDevReportTab.strengthWeakness:
        return const _AutismDevStrengthWeaknessSection();
      case _AutismDevReportTab.training:
        return _AutismDevProfilePdfSection(
          record: _displayRecord,
          tab: _AutismDevReportTab.training,
          future: _profilePdfFuture(_AutismDevReportTab.training),
          onRetry: () => _retryProfilePdf(_AutismDevReportTab.training),
        );
      case _AutismDevReportTab.developmentProfile:
        return _AutismDevProfilePdfSection(
          record: _displayRecord,
          tab: _AutismDevReportTab.developmentProfile,
          future: _profilePdfFuture(_AutismDevReportTab.developmentProfile),
          onRetry: () =>
              _retryProfilePdf(_AutismDevReportTab.developmentProfile),
        );
      case _AutismDevReportTab.behaviorProfile:
        return _AutismDevProfilePdfSection(
          record: _displayRecord,
          tab: _AutismDevReportTab.behaviorProfile,
          future: _profilePdfFuture(_AutismDevReportTab.behaviorProfile),
          onRetry: () => _retryProfilePdf(_AutismDevReportTab.behaviorProfile),
        );
    }
  }

  Future<Uint8List> _profilePdfFuture(_AutismDevReportTab tab) {
    if (!_isBackendProfileTab(tab)) {
      return Future<Uint8List>.value(Uint8List(0));
    }
    return _profilePdfFutures.putIfAbsent(tab, () => _loadProfilePdf(tab));
  }

  Future<Uint8List> _loadProfilePdf(_AutismDevReportTab tab) async {
    final Uint8List? cached = _profilePdfBytes[tab];
    if (cached != null) {
      return cached;
    }
    final String token = widget.token.trim();
    if (token.isEmpty || widget.record.id <= 0) {
      throw StateError('缺少报告加载参数');
    }
    final Uint8List bytes =
        await widget.client.downloadAutismDevRecordProfilePdf(
      token,
      widget.record.id,
      profile: _profilePdfCode(tab),
    );
    if (mounted) {
      _profilePdfBytes[tab] = bytes;
    }
    return bytes;
  }

  void _retryProfilePdf(_AutismDevReportTab tab) {
    setState(() {
      _profilePdfBytes.remove(tab);
      _profilePdfFutures.remove(tab);
    });
  }

  void _handleResultAnalysisCellTap(_AutismDevAnalysisEditRequest request) {
    if (_selectedAnalysisCell == request) {
      unawaited(_showResultAnalysisEditDialog(request));
      return;
    }
    setState(() {
      _selectedAnalysisCell = request;
    });
  }

  Future<void> _showResultAnalysisEditDialog(
    _AutismDevAnalysisEditRequest request,
  ) async {
    final List<AutismDevResultAnalysisRow> rows = _resultAnalysis.rows;
    if (request.rowIndex < 0 || request.rowIndex >= rows.length) {
      return;
    }
    final AutismDevResultAnalysisRow row = rows[request.rowIndex];
    final _AutismDevAnalysisEditResult? result =
        await showDialog<_AutismDevAnalysisEditResult>(
      context: context,
      barrierColor: const Color(0x33000000),
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _AutismDevAnalysisEditDialog(
            domain: row.domain,
            request: request,
            row: row,
          ),
        );
      },
    );
    if (result == null || !mounted) {
      return;
    }
    final List<AutismDevResultAnalysisRow> nextRows =
        List<AutismDevResultAnalysisRow>.from(_resultAnalysis.rows);
    nextRows[request.rowIndex] = row.copyWith(
      status: result.status,
      strengths: result.strengths,
      weaknesses: result.weaknesses,
      targets: result.targets,
    );
    final AutismDevResultAnalysis nextAnalysis = AutismDevResultAnalysis(
      title: _resultAnalysis.title,
      model: _resultAnalysis.model,
      generatedBy: _resultAnalysis.generatedBy,
      generatedAt: _resultAnalysis.generatedAt,
      rows: nextRows,
    );
    setState(() {
      _resultAnalysis = nextAnalysis;
      _selectedAnalysisCell = null;
    });
    try {
      final AutismDevResultAnalysis saved =
          await widget.client.saveAutismDevResultAnalysis(
        widget.token.trim(),
        widget.record.id,
        nextAnalysis,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _resultAnalysis = _mergeAutismDevResultAnalysis(saved);
      });
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('保存修改失败：$error')),
        );
      }
    }
  }

  static bool _isBackendProfileTab(_AutismDevReportTab tab) {
    return tab == _AutismDevReportTab.training ||
        tab == _AutismDevReportTab.developmentProfile ||
        tab == _AutismDevReportTab.behaviorProfile;
  }

  static String _profilePdfCode(_AutismDevReportTab tab) {
    return switch (tab) {
      _AutismDevReportTab.behaviorProfile => 'behavior',
      _AutismDevReportTab.training => 'training',
      _ => 'development',
    };
  }
}

class _AutismDevProfilePdfSection extends StatelessWidget {
  const _AutismDevProfilePdfSection({
    required this.record,
    required this.tab,
    required this.future,
    required this.onRetry,
  });

  final Pep3RecordSummary record;
  final _AutismDevReportTab tab;
  final Future<Uint8List> future;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 512,
      child: FutureBuilder<Uint8List>(
        future: future,
        builder: (BuildContext context, AsyncSnapshot<Uint8List> snapshot) {
          if (snapshot.connectionState != ConnectionState.done) {
            return const _ReportPreviewLoadingState(message: 'PDF加载中...');
          }
          if (snapshot.hasError) {
            return _ReportPreviewErrorState(
              message: 'PDF加载失败：${snapshot.error}',
              onRetry: onRetry,
            );
          }
          final Uint8List? bytes = snapshot.data;
          if (bytes == null || bytes.isEmpty) {
            return const _ReportPreviewEmptyState(message: '暂无PDF内容');
          }
          return _LazyReportPdfPreview(
            key: ValueKey<String>(
              'autismdev-profile-${record.id}-${record.updatedTime}-${tab.name}-${bytes.length}',
            ),
            bytes: bytes,
            pageCount: _autismDevPdfPageCount(bytes),
            maxPageWidth: 820,
          );
        },
      ),
    );
  }
}

int _autismDevPdfPageCount(Uint8List bytes) {
  if (bytes.isEmpty) {
    return 1;
  }
  final String source = latin1.decode(bytes, allowInvalid: true);
  final RegExp pageObjectPattern = RegExp(r'/Type\s*/Page\b');
  final int count = pageObjectPattern.allMatches(source).length;
  return math.max(1, count);
}

class _AutismDevReportTabBar extends StatelessWidget {
  const _AutismDevReportTabBar({
    required this.activeTab,
    required this.onSelected,
    required this.trailing,
  });

  final _AutismDevReportTab activeTab;
  final ValueChanged<_AutismDevReportTab> onSelected;
  final Widget trailing;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: <Widget>[
          Expanded(
            child: SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: Row(
                children: <Widget>[
                  for (int index = 0;
                      index < _autismDevReportTabs.length;
                      index++)
                    Padding(
                      padding: EdgeInsets.only(
                        right: index == _autismDevReportTabs.length - 1 ? 0 : 8,
                      ),
                      child: _ErxinReportTabChip(
                        label: _autismDevReportTabs[index].label,
                        active: activeTab == _autismDevReportTabs[index].tab,
                        onTap: () =>
                            onSelected(_autismDevReportTabs[index].tab),
                      ),
                    ),
                ],
              ),
            ),
          ),
          const SizedBox(width: 12),
          Container(
            width: 1,
            height: 28,
            color: _ReportTheme.lineSoft,
          ),
          const SizedBox(width: 12),
          trailing,
        ],
      ),
    );
  }
}

class _AutismDevReportPageHeader extends StatelessWidget {
  const _AutismDevReportPageHeader({required this.record});

  final Pep3RecordSummary record;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        const Center(
          child: Text(
            '孤独症儿童发展评估报告',
            style: TextStyle(
              color: _ReportTheme.ink,
              fontSize: 24,
              height: 1.2,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
        const SizedBox(height: 16),
        _AutismDevInfoGrid(
          items: <_AutismDevInfoItem>[
            _AutismDevInfoItem('儿童姓名', _studentName(record)),
            _AutismDevInfoItem('测评年龄', _ageText(record)),
            _AutismDevInfoItem('测评日期', _dateOnlyText(record.assessmentDate)),
            _AutismDevInfoItem(
              '评估者',
              record.examinerName.trim().isEmpty
                  ? '-'
                  : record.examinerName.trim(),
            ),
            _AutismDevInfoItem(
                '量表版本',
                record.scaleVersion.trim().isEmpty
                    ? '2010修订训练师版'
                    : record.scaleVersion.trim()),
            _AutismDevInfoItem(
                '测评次数', _sequenceText(record.assessmentSequence)),
          ],
        ),
      ],
    );
  }
}

class _AutismDevOverviewSection extends StatelessWidget {
  const _AutismDevOverviewSection();

  @override
  Widget build(BuildContext context) {
    final int developmentP = _autismDevDevelopmentScores.fold<int>(
      0,
      (int total, _AutismDevDevelopmentScore item) => total + item.p,
    );
    final int developmentTotal = _autismDevDevelopmentScores.fold<int>(
      0,
      (int total, _AutismDevDevelopmentScore item) => total + item.total,
    );
    final int trainingTargets = _autismDevDevelopmentScores.fold<int>(
      0,
      (int total, _AutismDevDevelopmentScore item) => total + item.e,
    );
    final int behaviorAttention = _autismDevBehaviorScores.fold<int>(
      0,
      (int total, _AutismDevBehaviorScore item) => total + item.m + item.s,
    );

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Wrap(
          spacing: 12,
          runSpacing: 12,
          children: <Widget>[
            _AutismDevOverviewMetric(
              label: '发展领域通过项',
              value: '$developmentP',
              suffix: '/ $developmentTotal',
              color: _ReportTheme.blue,
            ),
            _AutismDevOverviewMetric(
              label: '训练目标候选项',
              value: '$trainingTargets',
              suffix: '个E项',
              color: _ReportTheme.orangeDeep,
            ),
            _AutismDevOverviewMetric(
              label: '情绪行为需关注',
              value: '$behaviorAttention',
              suffix: '项',
              color: _ReportTheme.rose,
            ),
          ],
        ),
        const SizedBox(height: 18),
        const _AutismDevSectionTitle(
          title: '发展能力计分汇总表',
          subtitle: '七个发展领域按 P / E+F(X) 汇总；情绪与行为按 A / M / S 汇总。',
        ),
        const SizedBox(height: 10),
        const _AutismDevDevelopmentScoreTable(),
        const SizedBox(height: 18),
        const _AutismDevSectionTitle(
          title: '结果摘要',
          subtitle: '根据各领域通过项、中间反应项和情绪行为分布整理。',
        ),
        const SizedBox(height: 10),
        const _AutismDevReportSummaryGrid(),
      ],
    );
  }
}

class _AutismDevStrengthWeaknessSection extends StatelessWidget {
  const _AutismDevStrengthWeaknessSection();

  @override
  Widget build(BuildContext context) {
    return const Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _AutismDevSectionTitle(
          title: '优劣势分析',
          subtitle: '按发展领域整理优势能力和当前支持重点。',
        ),
        SizedBox(height: 10),
        _AutismDevStrengthWeaknessTable(),
      ],
    );
  }
}

// ignore: unused_element
class _AutismDevDevelopmentProfileSection extends StatelessWidget {
  const _AutismDevDevelopmentProfileSection();

  @override
  Widget build(BuildContext context) {
    return const _AutismDevReportFigure(
      assetPath: _autismDevDevelopmentProfileAsset,
      overlayPainter: _AutismDevDevelopmentProfilePainter(),
    );
  }
}

// ignore: unused_element
class _AutismDevBehaviorProfileSection extends StatelessWidget {
  const _AutismDevBehaviorProfileSection({required this.itemScoreLabels});

  final Map<int, String> itemScoreLabels;

  @override
  Widget build(BuildContext context) {
    return _AutismDevReportFigure(
      assetPath: _autismDevBehaviorProfileAsset,
      overlayPainter: _AutismDevBehaviorProfilePainter(
        itemLevels: _autismDevBehaviorItemLevelsFromScores(itemScoreLabels),
      ),
    );
  }
}

class _AutismDevAnalysisSection extends StatelessWidget {
  const _AutismDevAnalysisSection({
    required this.record,
    required this.analysis,
    required this.generating,
    required this.generationStatus,
    required this.generationError,
    required this.streamText,
    required this.selectedCell,
    required this.onCellTap,
  });

  final Pep3RecordSummary record;
  final AutismDevResultAnalysis analysis;
  final bool generating;
  final String generationStatus;
  final String generationError;
  final String streamText;
  final _AutismDevAnalysisEditRequest? selectedCell;
  final ValueChanged<_AutismDevAnalysisEditRequest> onCellTap;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 790),
        child: _AutismDevResultAnalysisSheet(
          record: record,
          analysis: analysis,
          generating: generating,
          generationStatus: generationStatus,
          generationError: generationError,
          streamText: streamText,
          selectedCell: selectedCell,
          onCellTap: onCellTap,
        ),
      ),
    );
  }
}

class _AutismDevResultAnalysisSheet extends StatelessWidget {
  const _AutismDevResultAnalysisSheet({
    required this.record,
    required this.analysis,
    required this.generating,
    required this.generationStatus,
    required this.generationError,
    required this.streamText,
    required this.selectedCell,
    required this.onCellTap,
  });

  final Pep3RecordSummary record;
  final AutismDevResultAnalysis analysis;
  final bool generating;
  final String generationStatus;
  final String generationError;
  final String streamText;
  final _AutismDevAnalysisEditRequest? selectedCell;
  final ValueChanged<_AutismDevAnalysisEditRequest> onCellTap;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        const SizedBox(height: 4),
        const Text(
          '孤独症儿童评估结果分析表',
          textAlign: TextAlign.center,
          style: TextStyle(
            color: Colors.black,
            fontSize: 22,
            height: 1.2,
          ),
        ),
        const SizedBox(height: 16),
        _AutismDevResultAnalysisMeta(record: record),
        const SizedBox(height: 2),
        if (generating) ...<Widget>[
          const SizedBox(height: 8),
          _AutismDevAnalysisStreamPanel(
            studentName: _studentName(record),
            status: generationStatus,
            streamText: streamText,
          ),
        ] else ...<Widget>[
          if (generationError.trim().isNotEmpty) ...<Widget>[
            _AutismDevAnalysisGenerationBanner(
              message: generationError.trim(),
              isError: true,
            ),
            const SizedBox(height: 8),
          ],
          _AutismDevResultAnalysisTable(
            rows: analysis.rows,
            selectedCell: selectedCell,
            onCellTap: onCellTap,
          ),
        ],
      ],
    );
  }
}

class _AutismDevAnalysisStreamPanel extends StatefulWidget {
  const _AutismDevAnalysisStreamPanel({
    required this.studentName,
    required this.status,
    required this.streamText,
  });

  final String studentName;
  final String status;
  final String streamText;

  @override
  State<_AutismDevAnalysisStreamPanel> createState() =>
      _AutismDevAnalysisStreamPanelState();
}

class _AutismDevAnalysisStreamPanelState
    extends State<_AutismDevAnalysisStreamPanel> {
  late final ScrollController _scrollController;
  bool _stickToBottom = true;
  bool _scrollSyncScheduled = false;

  @override
  void initState() {
    super.initState();
    _scrollController = ScrollController();
    _scrollController.addListener(_handleScroll);
  }

  @override
  void didUpdateWidget(covariant _AutismDevAnalysisStreamPanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.streamText != oldWidget.streamText) {
      _scheduleStickToBottom();
    }
  }

  @override
  void dispose() {
    _scrollController.removeListener(_handleScroll);
    _scrollController.dispose();
    super.dispose();
  }

  void _handleScroll() {
    if (!_scrollController.hasClients) {
      return;
    }
    final ScrollPosition position = _scrollController.position;
    _stickToBottom = position.maxScrollExtent - position.pixels <= 48;
  }

  void _scheduleStickToBottom() {
    if (_scrollSyncScheduled) {
      return;
    }
    _scrollSyncScheduled = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _scrollSyncScheduled = false;
      if (!mounted || !_scrollController.hasClients || !_stickToBottom) {
        return;
      }
      _jumpToBottom();
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted || !_scrollController.hasClients || !_stickToBottom) {
          return;
        }
        _jumpToBottom();
      });
    });
  }

  void _jumpToBottom() {
    final ScrollPosition position = _scrollController.position;
    final double target = position.maxScrollExtent
        .clamp(position.minScrollExtent, double.infinity);
    if ((position.pixels - target).abs() <= .5) {
      return;
    }
    _scrollController.jumpTo(target);
  }

  @override
  Widget build(BuildContext context) {
    final String status =
        widget.status.trim().isEmpty ? 'AI正在生成评估结果分析' : widget.status.trim();
    final double progress =
        _autismDevResultAnalysisStreamProgress(widget.streamText);
    final String progressText = '${(progress * 100).round()}%';
    final _AutismDevAnalysisReadableStream readable =
        _AutismDevAnalysisReadableStream.fromRaw(widget.streamText);

    return SizedBox(
      height: 560,
      child: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: const Color(0xFFFFD8C3)),
          boxShadow: const <BoxShadow>[
            BoxShadow(
              color: Color(0x12B05F32),
              blurRadius: 14,
              offset: Offset(0, 6),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            Padding(
              padding: const EdgeInsets.fromLTRB(18, 16, 18, 12),
              child: Row(
                children: <Widget>[
                  SizedBox(
                    width: 24,
                    height: 24,
                    child: CircularProgressIndicator(
                      strokeWidth: 2.4,
                      color: _ReportTheme.orangeDeep,
                      backgroundColor: const Color(0xFFFFEEE4),
                      value: progress >= .98 ? progress : null,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Text(
                          '正在生成 ${widget.studentName} 的评估结果分析',
                          style: const TextStyle(
                            color: _ReportTheme.ink,
                            fontSize: 16,
                            fontWeight: FontWeight.w800,
                            height: 1.15,
                          ),
                        ),
                        const SizedBox(height: 6),
                        Row(
                          children: <Widget>[
                            Text(
                              status,
                              style: const TextStyle(
                                color: _ReportTheme.text,
                                fontSize: 12,
                                fontWeight: FontWeight.w700,
                                height: 1.1,
                              ),
                            ),
                            const SizedBox(width: 10),
                            Expanded(
                              child: ClipRRect(
                                borderRadius: BorderRadius.circular(99),
                                child: LinearProgressIndicator(
                                  minHeight: 8,
                                  value: progress,
                                  backgroundColor: const Color(0xFFFFEEE4),
                                  color: _ReportTheme.orange,
                                ),
                              ),
                            ),
                            const SizedBox(width: 8),
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: const Color(0xFFFFF3EA),
                                borderRadius: BorderRadius.circular(999),
                              ),
                              child: Text(
                                progressText,
                                style: const TextStyle(
                                  color: _ReportTheme.orangeDeep,
                                  fontSize: 11,
                                  fontWeight: FontWeight.w800,
                                  height: 1,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: Container(
                margin: const EdgeInsets.fromLTRB(18, 0, 18, 14),
                padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFFCF8),
                  borderRadius: BorderRadius.circular(9),
                  border: Border.all(color: _ReportTheme.lineSoft),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: <Widget>[
                    Row(
                      children: const <Widget>[
                        Icon(
                          Icons.auto_awesome_rounded,
                          size: 16,
                          color: _ReportTheme.orangeDeep,
                        ),
                        SizedBox(width: 6),
                        Text(
                          'AI正在整理可预览内容',
                          style: TextStyle(
                            color: _ReportTheme.ink,
                            fontSize: 12.5,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Expanded(
                      child: SingleChildScrollView(
                        controller: _scrollController,
                        physics: const ClampingScrollPhysics(),
                        child: Padding(
                          padding: const EdgeInsets.only(bottom: 26),
                          child: Text(
                            readable.content,
                            style: const TextStyle(
                              color: Colors.black,
                              fontSize: 12.8,
                              height: 1.48,
                            ),
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(18, 0, 18, 16),
              child: Row(
                children: const <Widget>[
                  Icon(
                    Icons.auto_awesome_rounded,
                    color: _ReportTheme.orangeDeep,
                    size: 16,
                  ),
                  SizedBox(width: 6),
                  Expanded(
                    child: Text(
                      '生成完成后将自动切换为正式分析表，可继续手动修改',
                      style: TextStyle(
                        color: _ReportTheme.muted,
                        fontSize: 11.5,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _AutismDevResultAnalysisMeta extends StatelessWidget {
  const _AutismDevResultAnalysisMeta({required this.record});

  final Pep3RecordSummary record;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Expanded(
          flex: 31,
          child: _AutismDevAnalysisMetaLine(
            label: '儿童姓名：',
            value: _studentName(record),
          ),
        ),
        const SizedBox(width: 32),
        Expanded(
          flex: 25,
          child: _AutismDevAnalysisMetaLine(
            label: '评估者：',
            value: record.examinerName.trim(),
          ),
        ),
        const SizedBox(width: 32),
        Expanded(
          flex: 32,
          child: _AutismDevAnalysisMetaLine(
            label: '评估时间：',
            value: _dateOnlyText(record.assessmentDate),
          ),
        ),
      ],
    );
  }
}

class _AutismDevAnalysisMetaLine extends StatelessWidget {
  const _AutismDevAnalysisMetaLine({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Text(
          label,
          style: const TextStyle(
            color: Colors.black,
            fontSize: 17,
            height: 1.2,
          ),
        ),
        Expanded(
          child: Container(
            height: 24,
            alignment: Alignment.center,
            decoration: const BoxDecoration(
              border: Border(
                bottom: BorderSide(color: Colors.black, width: .8),
              ),
            ),
            child: Text(
              value.trim(),
              maxLines: 1,
              style: const TextStyle(
                color: Colors.black,
                fontSize: 16,
                height: 1.15,
              ),
            ),
          ),
        ),
      ],
    );
  }
}

enum _AutismDevAnalysisEditField {
  status,
  strengthWeakness,
  targets,
}

class _AutismDevAnalysisEditRequest {
  const _AutismDevAnalysisEditRequest({
    required this.rowIndex,
    required this.field,
  });

  final int rowIndex;
  final _AutismDevAnalysisEditField field;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        other is _AutismDevAnalysisEditRequest &&
            other.rowIndex == rowIndex &&
            other.field == field;
  }

  @override
  int get hashCode => Object.hash(rowIndex, field);
}

class _AutismDevAnalysisEditResult {
  const _AutismDevAnalysisEditResult({
    this.status = '',
    this.strengths = '',
    this.weaknesses = '',
    this.targets = '',
  });

  final String status;
  final String strengths;
  final String weaknesses;
  final String targets;
}

class _AutismDevAnalysisGenerationBanner extends StatelessWidget {
  const _AutismDevAnalysisGenerationBanner({
    required this.message,
    required this.isError,
  });

  final String message;
  final bool isError;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: isError ? const Color(0xFFFFF0EE) : const Color(0xFFFFF7EC),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isError ? const Color(0xFFE05D4F) : _ReportTheme.lineSoft,
        ),
      ),
      child: Row(
        children: <Widget>[
          Icon(
            isError ? Icons.error_outline_rounded : Icons.auto_awesome_rounded,
            size: 16,
            color: isError ? const Color(0xFFE05D4F) : _ReportTheme.orangeDeep,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              message,
              style: TextStyle(
                color: isError ? const Color(0xFFB64437) : Colors.black,
                fontSize: 12.5,
                height: 1.35,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevAnalysisEditDialog extends StatefulWidget {
  const _AutismDevAnalysisEditDialog({
    required this.domain,
    required this.request,
    required this.row,
  });

  final String domain;
  final _AutismDevAnalysisEditRequest request;
  final AutismDevResultAnalysisRow row;

  @override
  State<_AutismDevAnalysisEditDialog> createState() =>
      _AutismDevAnalysisEditDialogState();
}

class _AutismDevAnalysisEditDialogState
    extends State<_AutismDevAnalysisEditDialog> {
  late final TextEditingController _statusController;
  late final TextEditingController _strengthsController;
  late final TextEditingController _weaknessesController;
  late final TextEditingController _targetsController;

  bool get _editingStrengthWeakness =>
      widget.request.field == _AutismDevAnalysisEditField.strengthWeakness;

  @override
  void initState() {
    super.initState();
    _statusController = TextEditingController(text: widget.row.status);
    _strengthsController = TextEditingController(text: widget.row.strengths);
    _weaknessesController = TextEditingController(text: widget.row.weaknesses);
    _targetsController = TextEditingController(text: widget.row.targets);
  }

  @override
  void dispose() {
    _statusController.dispose();
    _strengthsController.dispose();
    _weaknessesController.dispose();
    _targetsController.dispose();
    super.dispose();
  }

  void _submit() {
    Navigator.of(context).pop(
      _AutismDevAnalysisEditResult(
        status: _statusController.text.trim(),
        strengths: _strengthsController.text.trim(),
        weaknesses: _weaknessesController.text.trim(),
        targets: _targetsController.text.trim(),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final String title = switch (widget.request.field) {
      _AutismDevAnalysisEditField.status => '编辑能力现状描述',
      _AutismDevAnalysisEditField.strengthWeakness => '编辑优劣分析',
      _AutismDevAnalysisEditField.targets => '编辑训练目标',
    };
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: _editingStrengthWeakness ? 720 : 620,
        padding: const EdgeInsets.fromLTRB(24, 22, 24, 22),
        decoration: BoxDecoration(
          color: const Color(0xFFFFFCF8),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: _ReportTheme.lineSoft),
          boxShadow: const <BoxShadow>[
            BoxShadow(
              color: Color(0x20000000),
              blurRadius: 30,
              offset: Offset(0, 16),
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                Text(
                  title,
                  style: const TextStyle(
                    color: Colors.black,
                    fontSize: 22,
                    fontWeight: FontWeight.w900,
                    height: 1,
                  ),
                ),
                const SizedBox(width: 12),
                _AutismDevEditPill(text: widget.domain),
                const Spacer(),
                _AutismDevDialogIconButton(
                  icon: Icons.close_rounded,
                  onTap: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 18),
            if (widget.request.field == _AutismDevAnalysisEditField.status)
              _AutismDevAnalysisDialogField(
                label: '能力现状描述',
                controller: _statusController,
                minLines: 4,
                maxLines: 8,
              )
            else if (widget.request.field ==
                _AutismDevAnalysisEditField.strengthWeakness) ...<Widget>[
              _AutismDevAnalysisDialogField(
                label: '优势',
                controller: _strengthsController,
                minLines: 3,
                maxLines: 6,
              ),
              const SizedBox(height: 12),
              _AutismDevAnalysisDialogField(
                label: '劣势',
                controller: _weaknessesController,
                minLines: 3,
                maxLines: 6,
              ),
            ] else
              _AutismDevAnalysisDialogField(
                label: '训练目标',
                controller: _targetsController,
                minLines: 4,
                maxLines: 8,
              ),
            const SizedBox(height: 14),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: <Widget>[
                _AutismDevDialogAction(
                  label: '取消',
                  onTap: () => Navigator.of(context).pop(),
                ),
                const SizedBox(width: 10),
                _AutismDevDialogAction(
                  label: '保存',
                  filled: true,
                  icon: Icons.save_outlined,
                  onTap: _submit,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _AutismDevAnalysisDialogField extends StatelessWidget {
  const _AutismDevAnalysisDialogField({
    required this.label,
    required this.controller,
    required this.minLines,
    required this.maxLines,
  });

  final String label;
  final TextEditingController controller;
  final int minLines;
  final int maxLines;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          label,
          style: const TextStyle(
            color: Colors.black,
            fontSize: 13,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 7),
        TextField(
          controller: controller,
          minLines: minLines,
          maxLines: maxLines,
          style: const TextStyle(
            color: Colors.black,
            fontSize: 14.5,
            height: 1.35,
          ),
          decoration: InputDecoration(
            isDense: true,
            filled: true,
            fillColor: Colors.white,
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 12, vertical: 11),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: _ReportTheme.lineSoft),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide:
                  const BorderSide(color: _ReportTheme.orange, width: 1.2),
            ),
          ),
        ),
      ],
    );
  }
}

class _AutismDevEditPill extends StatelessWidget {
  const _AutismDevEditPill({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 26,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF3EA),
        borderRadius: BorderRadius.circular(99),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Text(
        text,
        style: const TextStyle(
          color: Colors.black,
          fontSize: 12,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _AutismDevDialogIconButton extends StatelessWidget {
  const _AutismDevDialogIconButton({
    required this.icon,
    required this.onTap,
  });

  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      shape: const CircleBorder(),
      child: InkWell(
        onTap: onTap,
        customBorder: const CircleBorder(),
        child: Container(
          width: 34,
          height: 34,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: const Color(0xFFFFF6F0),
            shape: BoxShape.circle,
            border: Border.all(color: _ReportTheme.lineSoft),
          ),
          child: Icon(
            icon,
            size: 20,
            color: _ReportTheme.muted,
          ),
        ),
      ),
    );
  }
}

class _AutismDevDialogAction extends StatelessWidget {
  const _AutismDevDialogAction({
    required this.label,
    required this.onTap,
    this.filled = false,
    this.icon,
  });

  final String label;
  final VoidCallback onTap;
  final bool filled;
  final IconData? icon;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(13),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Container(
          height: 38,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: filled ? _ReportTheme.orange : Colors.white,
            borderRadius: BorderRadius.circular(13),
            border: filled ? null : Border.all(color: _ReportTheme.lineSoft),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              if (icon != null) ...<Widget>[
                Icon(
                  icon,
                  size: 17,
                  color: filled ? Colors.white : Colors.black,
                ),
                const SizedBox(width: 6),
              ],
              Text(
                label,
                style: TextStyle(
                  color: filled ? Colors.white : Colors.black,
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

class _AutismDevSectionTitle extends StatelessWidget {
  const _AutismDevSectionTitle({
    required this.title,
    required this.subtitle,
  });

  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Container(
          width: 4,
          height: 32,
          margin: const EdgeInsets.only(top: 2),
          decoration: BoxDecoration(
            color: _ReportTheme.orange,
            borderRadius: BorderRadius.circular(999),
          ),
        ),
        const SizedBox(width: 10),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Text(
                title,
                style: const TextStyle(
                  color: _ReportTheme.ink,
                  fontSize: 18,
                  height: 1.2,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                subtitle,
                style: const TextStyle(
                  color: _ReportTheme.muted,
                  fontSize: 12,
                  height: 1.35,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _AutismDevInfoGrid extends StatelessWidget {
  const _AutismDevInfoGrid({required this.items});

  final List<_AutismDevInfoItem> items;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: <Widget>[
        for (final _AutismDevInfoItem item in items)
          SizedBox(
            width: 264,
            child: _AutismDevInfoCell(item: item),
          ),
      ],
    );
  }
}

class _AutismDevInfoCell extends StatelessWidget {
  const _AutismDevInfoCell({required this.item});

  final _AutismDevInfoItem item;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 46,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          Text(
            item.label,
            style: const TextStyle(
              color: _ReportTheme.muted,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              item.value.trim().isEmpty ? '-' : item.value.trim(),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.right,
              style: const TextStyle(
                color: _ReportTheme.ink,
                fontSize: 14,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevOverviewMetric extends StatelessWidget {
  const _AutismDevOverviewMetric({
    required this.label,
    required this.value,
    required this.suffix,
    required this.color,
  });

  final String label;
  final String value;
  final String suffix;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 264,
      height: 82,
      padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            label,
            style: const TextStyle(
              color: _ReportTheme.muted,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
          const Spacer(),
          Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: <Widget>[
              Text(
                value,
                style: TextStyle(
                  color: color,
                  fontSize: 30,
                  height: .9,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(width: 7),
              Padding(
                padding: const EdgeInsets.only(bottom: 2),
                child: Text(
                  suffix,
                  style: const TextStyle(
                    color: _ReportTheme.text,
                    fontSize: 13,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _AutismDevReportSummaryGrid extends StatelessWidget {
  const _AutismDevReportSummaryGrid();

  @override
  Widget build(BuildContext context) {
    const List<_AutismDevSummaryItem> items = <_AutismDevSummaryItem>[
      _AutismDevSummaryItem(
        '能力现状',
        '认知、生活自理和感知觉通过项占比较高。',
        Icons.trending_up_rounded,
      ),
      _AutismDevSummaryItem(
        '目标候选',
        '语言表达、精细动作和社会交往E项较集中。',
        Icons.flag_rounded,
      ),
      _AutismDevSummaryItem(
        '行为关注',
        '感觉偏好、情绪调节和特殊行为需要持续观察。',
        Icons.visibility_rounded,
      ),
      _AutismDevSummaryItem(
        '训练建议',
        '优先从E项转化短期目标，再补充F项前置能力。',
        Icons.assignment_turned_in_rounded,
      ),
    ];
    return Wrap(
      spacing: 10,
      runSpacing: 10,
      children: <Widget>[
        for (final _AutismDevSummaryItem item in items)
          SizedBox(
            width: 198,
            child: _AutismDevSummaryTile(item: item),
          ),
      ],
    );
  }
}

class _AutismDevSummaryTile extends StatelessWidget {
  const _AutismDevSummaryTile({required this.item});

  final _AutismDevSummaryItem item;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 84,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          Container(
            width: 34,
            height: 34,
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: const Color(0xFFFFF1E8),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(
              item.icon,
              color: _ReportTheme.orangeDeep,
              size: 18,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  item.title,
                  style: const TextStyle(
                    color: _ReportTheme.ink,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 5),
                Text(
                  item.description,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _ReportTheme.muted,
                    fontSize: 11,
                    height: 1.25,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevDevelopmentScoreTable extends StatelessWidget {
  const _AutismDevDevelopmentScoreTable();

  @override
  Widget build(BuildContext context) {
    return _AutismDevTableFrame(
      child: Table(
        columnWidths: const <int, TableColumnWidth>{
          0: FlexColumnWidth(1.7),
          1: FlexColumnWidth(1),
          2: FlexColumnWidth(1),
          3: FlexColumnWidth(1),
        },
        border: TableBorder.all(color: _ReportTheme.lineSoft),
        children: <TableRow>[
          _autismDevTableHeaderRow(<String>['领域', 'P', 'E+F(X)', '总分']),
          for (final _AutismDevDevelopmentScore item
              in _autismDevDevelopmentScores)
            _autismDevTableRow(
              <String>[
                item.label,
                '${item.p}',
                '${item.supportCount}',
                '${item.p}',
              ],
              emphFirst: true,
            ),
        ],
      ),
    );
  }
}

class _AutismDevStrengthWeaknessTable extends StatelessWidget {
  const _AutismDevStrengthWeaknessTable();

  @override
  Widget build(BuildContext context) {
    return _AutismDevTableFrame(
      child: Table(
        columnWidths: const <int, TableColumnWidth>{
          0: FlexColumnWidth(1.05),
          1: FlexColumnWidth(3.6),
        },
        border: TableBorder.all(color: _ReportTheme.lineSoft),
        children: <TableRow>[
          _autismDevTableHeaderRow(<String>['领域', '优劣势分析'], tall: true),
          for (final _AutismDevAnalysisItem item in _autismDevAnalysisItems)
            TableRow(
              children: <Widget>[
                _autismDevTableCell(item.domain, emph: true, minHeight: 66),
                _autismDevTableCell(item.strength, minHeight: 66),
              ],
            ),
        ],
      ),
    );
  }
}

class _AutismDevResultAnalysisTable extends StatelessWidget {
  const _AutismDevResultAnalysisTable({
    required this.rows,
    required this.selectedCell,
    required this.onCellTap,
  });

  final List<AutismDevResultAnalysisRow> rows;
  final _AutismDevAnalysisEditRequest? selectedCell;
  final ValueChanged<_AutismDevAnalysisEditRequest> onCellTap;

  @override
  Widget build(BuildContext context) {
    final List<AutismDevResultAnalysisRow> displayRows =
        _normalizeAutismDevResultAnalysisRows(rows);
    return ClipRect(
      child: Table(
        columnWidths: const <int, TableColumnWidth>{
          0: FlexColumnWidth(1.05),
          1: FlexColumnWidth(1.45),
          2: FlexColumnWidth(2.08),
          3: FlexColumnWidth(1.95),
        },
        border: TableBorder.all(color: Colors.black, width: .8),
        children: <TableRow>[
          TableRow(
            children: <Widget>[
              _autismDevAnalysisHeaderCell('领   域'),
              _autismDevAnalysisHeaderCell('能力现状描述'),
              _autismDevAnalysisHeaderCell('优劣分析'),
              _autismDevAnalysisHeaderCell('训练目标'),
            ],
          ),
          for (int index = 0; index < displayRows.length; index += 1)
            TableRow(
              children: <Widget>[
                _autismDevAnalysisDomainCell(displayRows[index].domain),
                _autismDevAnalysisTextCell(
                  displayRows[index].status,
                  editable: selectedCell ==
                      _AutismDevAnalysisEditRequest(
                        rowIndex: index,
                        field: _AutismDevAnalysisEditField.status,
                      ),
                  onTap: () => onCellTap(
                    _AutismDevAnalysisEditRequest(
                      rowIndex: index,
                      field: _AutismDevAnalysisEditField.status,
                    ),
                  ),
                ),
                _autismDevAnalysisStrengthCell(
                  displayRows[index],
                  editable: selectedCell ==
                      _AutismDevAnalysisEditRequest(
                        rowIndex: index,
                        field: _AutismDevAnalysisEditField.strengthWeakness,
                      ),
                  onTap: () => onCellTap(
                    _AutismDevAnalysisEditRequest(
                      rowIndex: index,
                      field: _AutismDevAnalysisEditField.strengthWeakness,
                    ),
                  ),
                ),
                _autismDevAnalysisTextCell(
                  displayRows[index].targets,
                  editable: selectedCell ==
                      _AutismDevAnalysisEditRequest(
                        rowIndex: index,
                        field: _AutismDevAnalysisEditField.targets,
                      ),
                  onTap: () => onCellTap(
                    _AutismDevAnalysisEditRequest(
                      rowIndex: index,
                      field: _AutismDevAnalysisEditField.targets,
                    ),
                  ),
                ),
              ],
            ),
        ],
      ),
    );
  }
}

Widget _autismDevAnalysisHeaderCell(String text) {
  return Container(
    height: 38,
    alignment: Alignment.center,
    padding: const EdgeInsets.symmetric(horizontal: 8),
    color: Colors.white,
    child: Text(
      text,
      textAlign: TextAlign.center,
      style: const TextStyle(
        color: Colors.black,
        fontSize: 18,
        height: 1.15,
      ),
    ),
  );
}

Widget _autismDevAnalysisDomainCell(String domain) {
  return Container(
    constraints: const BoxConstraints(minHeight: 246),
    alignment: Alignment.center,
    padding: const EdgeInsets.symmetric(horizontal: 8),
    color: Colors.white,
    child: Text(
      domain,
      textAlign: TextAlign.center,
      style: const TextStyle(
        color: Colors.black,
        fontSize: 19,
        height: 1.2,
      ),
    ),
  );
}

Widget _autismDevAnalysisTextCell(
  String text, {
  required bool editable,
  required VoidCallback onTap,
}) {
  return _autismDevAnalysisEditableFrame(
    editable: editable,
    onTap: onTap,
    child: Text(
      text.trim(),
      textAlign: TextAlign.left,
      style: const TextStyle(
        color: Colors.black,
        fontSize: 15.5,
        height: 1.36,
      ),
    ),
  );
}

Widget _autismDevAnalysisStrengthCell(
  AutismDevResultAnalysisRow row, {
  required bool editable,
  required VoidCallback onTap,
}) {
  return _autismDevAnalysisEditableFrame(
    editable: editable,
    onTap: onTap,
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _autismDevAnalysisLabeledText(
          label: '优势：',
          text: row.strengths,
        ),
        const SizedBox(height: 24),
        _autismDevAnalysisLabeledText(
          label: '劣势：',
          text: row.weaknesses,
        ),
      ],
    ),
  );
}

Widget _autismDevAnalysisLabeledText({
  required String label,
  required String text,
}) {
  const TextStyle style = TextStyle(
    color: Colors.black,
    fontSize: 15.5,
    height: 1.36,
  );
  return Text.rich(
    TextSpan(
      style: style,
      children: <InlineSpan>[
        TextSpan(text: label),
        TextSpan(text: text.trim()),
      ],
    ),
    textAlign: TextAlign.left,
  );
}

Widget _autismDevAnalysisEditableFrame({
  required Widget child,
  required bool editable,
  required VoidCallback onTap,
}) {
  final Widget content = Container(
    constraints: const BoxConstraints(minHeight: 246),
    alignment: Alignment.topLeft,
    padding: EdgeInsets.fromLTRB(10, 9, editable ? 20 : 10, 9),
    decoration: BoxDecoration(
      color: Colors.white,
      boxShadow: editable
          ? const <BoxShadow>[
              BoxShadow(
                color: Color(0x22E96F43),
                blurRadius: 0,
                spreadRadius: 1.4,
              ),
            ]
          : null,
    ),
    child: Stack(
      clipBehavior: Clip.none,
      children: <Widget>[
        SizedBox(width: double.infinity, child: child),
        if (editable)
          const Positioned(
            right: -10,
            top: -2,
            child: Icon(
              Icons.edit_rounded,
              size: 12,
              color: _ReportTheme.orangeDeep,
            ),
          ),
      ],
    ),
  );
  return Material(
    color: Colors.transparent,
    child: InkWell(
      onTap: onTap,
      child: content,
    ),
  );
}

class _AutismDevReportFigure extends StatelessWidget {
  const _AutismDevReportFigure({
    required this.assetPath,
    required this.overlayPainter,
  });

  final String assetPath;
  final CustomPainter overlayPainter;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 820),
        child: AspectRatio(
          aspectRatio: 1488 / 2103,
          child: Stack(
            fit: StackFit.expand,
            children: <Widget>[
              Image.asset(assetPath, fit: BoxFit.contain),
              CustomPaint(painter: overlayPainter),
            ],
          ),
        ),
      ),
    );
  }
}

class _AutismDevDevelopmentProfilePainter extends CustomPainter {
  const _AutismDevDevelopmentProfilePainter();

  static const double _figureWidth = 1488;
  static const double _figureHeight = 2103;
  static const List<String> _profileDomains = <String>[
    '感知觉',
    '粗大动作',
    '精细动作',
    '语言与沟通',
    '认知',
    '社会交往',
    '生活自理',
  ];
  static const List<double> _profileColumnXs = <double>[
    380,
    478.5,
    581.5,
    681,
    780.5,
    883.5,
    990.5,
  ];
  static const List<double> _profilePointXs = <double>[
    376.5,
    478.5,
    578,
    677.5,
    777,
    883.5,
    987,
  ];
  static const double _developmentTotalX = 1094;
  static const double _developmentTotalPointX = 1090.5;
  static const double _pScoreY = 1650;
  static const double _eScoreY = 1739;

  @override
  void paint(Canvas canvas, Size size) {
    final List<Offset> abilityPoints = <Offset>[
      for (int index = 0; index < _profileDomains.length; index++)
        _point(
          size,
          _profilePointXs[index],
          _scoreYFor(
            _profileDomains[index],
            _developmentScore(_profileDomains[index]).p,
          ),
        ),
      _point(
        size,
        _developmentTotalPointX,
        _scoreYFor('发展分数', _developmentTotalP),
      ),
    ];
    final List<Offset> targetPoints = <Offset>[
      for (int index = 0; index < _profileDomains.length; index++)
        _point(
          size,
          _profilePointXs[index],
          _scoreYFor(
            _profileDomains[index],
            _developmentScore(_profileDomains[index]).p +
                _developmentScore(_profileDomains[index]).e,
          ),
        ),
    ];
    if (abilityPoints.isEmpty || targetPoints.isEmpty) {
      return;
    }
    final double strokeWidth = size.width / 360;
    final Paint abilityLinePaint = Paint()
      ..color = _ReportTheme.blue.withOpacity(.86)
      ..strokeWidth = strokeWidth
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;
    final Paint targetLinePaint = Paint()
      ..color = _ReportTheme.orangeDeep.withOpacity(.84)
      ..strokeWidth = strokeWidth
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;
    final Paint haloPaint = Paint()
      ..color = Colors.white.withOpacity(.92)
      ..style = PaintingStyle.fill;
    final Paint abilityDotPaint = Paint()
      ..color = _ReportTheme.blue
      ..style = PaintingStyle.fill;
    final Paint targetDotPaint = Paint()
      ..color = _ReportTheme.orangeDeep
      ..style = PaintingStyle.fill;
    _drawDashedPath(canvas, _pathFor(targetPoints), targetLinePaint);
    canvas.drawPath(_pathFor(abilityPoints), abilityLinePaint);
    for (final Offset point in targetPoints) {
      canvas.drawCircle(point, strokeWidth * 2.6, haloPaint);
      canvas.drawCircle(point, strokeWidth * 1.55, targetDotPaint);
    }
    for (final Offset point in abilityPoints) {
      canvas.drawCircle(point, strokeWidth * 2.6, haloPaint);
      canvas.drawCircle(point, strokeWidth * 1.65, abilityDotPaint);
    }
    _drawScoreBoxes(canvas, size);
  }

  static _AutismDevDevelopmentScore _developmentScore(String label) {
    return _autismDevDevelopmentScores.firstWhere(
      (_AutismDevDevelopmentScore item) => item.label == label,
    );
  }

  static void _drawScoreBoxes(Canvas canvas, Size size) {
    for (int index = 0; index < _profileDomains.length; index++) {
      final _AutismDevDevelopmentScore score =
          _developmentScore(_profileDomains[index]);
      _drawScoreText(
        canvas,
        size,
        _profileColumnXs[index],
        _pScoreY,
        '${score.p}',
      );
      _drawScoreText(
        canvas,
        size,
        _profileColumnXs[index],
        _eScoreY,
        '${score.e}',
      );
    }
    _drawScoreText(
      canvas,
      size,
      _developmentTotalX,
      _pScoreY,
      '$_developmentTotalP',
    );
  }

  static int get _developmentTotalP => _autismDevDevelopmentScores.fold<int>(
        0,
        (int total, _AutismDevDevelopmentScore item) => total + item.p,
      );

  static void _drawScoreText(
    Canvas canvas,
    Size size,
    double sourceX,
    double sourceY,
    String value,
  ) {
    final double scale = size.width / _figureWidth;
    final TextPainter painter = TextPainter(
      text: TextSpan(
        text: value,
        style: TextStyle(
          color: _ReportTheme.ink,
          fontSize: 28 * scale,
          height: 1,
          fontWeight: FontWeight.w900,
        ),
      ),
      textAlign: TextAlign.center,
      textDirection: TextDirection.ltr,
    )..layout();
    final Offset center = _point(size, sourceX, sourceY);
    painter.paint(
      canvas,
      Offset(center.dx - painter.width / 2, center.dy - painter.height / 2),
    );
  }

  static double _scoreYFor(String domain, int score) {
    final List<_AutismDevProfileScalePoint> points =
        _profileScalePoints[domain] ?? const <_AutismDevProfileScalePoint>[];
    if (points.isEmpty) {
      return 1589;
    }
    final double value = score.toDouble();
    if (value >= points.first.score) {
      return points.first.y;
    }
    for (int index = 0; index < points.length - 1; index++) {
      final _AutismDevProfileScalePoint upper = points[index];
      final _AutismDevProfileScalePoint lower = points[index + 1];
      if (value <= upper.score && value >= lower.score) {
        final double span = upper.score - lower.score;
        if (span <= 0) {
          return upper.y;
        }
        final double progress = (upper.score - value) / span;
        return upper.y + (lower.y - upper.y) * progress;
      }
    }
    return points.last.y;
  }

  static Offset _point(Size size, double x, double y) {
    return Offset(
      x / _figureWidth * size.width,
      y / _figureHeight * size.height,
    );
  }

  static Path _pathFor(List<Offset> points) {
    final Path path = Path()..moveTo(points.first.dx, points.first.dy);
    for (final Offset point in points.skip(1)) {
      path.lineTo(point.dx, point.dy);
    }
    return path;
  }

  static void _drawDashedPath(Canvas canvas, Path path, Paint paint) {
    for (final metric in path.computeMetrics()) {
      double distance = 0;
      while (distance < metric.length) {
        final double next =
            math.min(distance + paint.strokeWidth * 4, metric.length);
        canvas.drawPath(metric.extractPath(distance, next), paint);
        distance += paint.strokeWidth * 6.4;
      }
    }
  }

  @override
  bool shouldRepaint(
      covariant _AutismDevDevelopmentProfilePainter oldDelegate) {
    return false;
  }
}

class _AutismDevProfileScalePoint {
  const _AutismDevProfileScalePoint(this.score, this.y);

  final double score;
  final double y;
}

const Map<String, List<_AutismDevProfileScalePoint>> _profileScalePoints =
    <String, List<_AutismDevProfileScalePoint>>{
  '感知觉': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(55, 586),
    _AutismDevProfileScalePoint(52, 760),
    _AutismDevProfileScalePoint(47, 862),
    _AutismDevProfileScalePoint(44, 940),
    _AutismDevProfileScalePoint(40, 1026),
    _AutismDevProfileScalePoint(37, 1093),
    _AutismDevProfileScalePoint(36, 1120),
    _AutismDevProfileScalePoint(29, 1210),
    _AutismDevProfileScalePoint(27, 1262),
    _AutismDevProfileScalePoint(21, 1344),
    _AutismDevProfileScalePoint(19, 1404),
    _AutismDevProfileScalePoint(16, 1430),
    _AutismDevProfileScalePoint(10, 1511),
    _AutismDevProfileScalePoint(5, 1524),
    _AutismDevProfileScalePoint(2, 1539),
    _AutismDevProfileScalePoint(1, 1554),
    _AutismDevProfileScalePoint(0, 1568),
  ],
  '粗大动作': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(72, 529),
    _AutismDevProfileScalePoint(65, 727),
    _AutismDevProfileScalePoint(64, 760),
    _AutismDevProfileScalePoint(47, 862),
    _AutismDevProfileScalePoint(46, 938),
    _AutismDevProfileScalePoint(35, 1004),
    _AutismDevProfileScalePoint(34, 1117),
    _AutismDevProfileScalePoint(24, 1164),
    _AutismDevProfileScalePoint(22, 1238),
    _AutismDevProfileScalePoint(21, 1270),
    _AutismDevProfileScalePoint(19, 1313),
    _AutismDevProfileScalePoint(7, 1376),
    _AutismDevProfileScalePoint(6, 1418),
    _AutismDevProfileScalePoint(5, 1451),
    _AutismDevProfileScalePoint(1, 1482),
    _AutismDevProfileScalePoint(0, 1496),
  ],
  '精细动作': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(66, 529),
    _AutismDevProfileScalePoint(63, 740),
    _AutismDevProfileScalePoint(62, 755),
    _AutismDevProfileScalePoint(51, 787),
    _AutismDevProfileScalePoint(50, 817),
    _AutismDevProfileScalePoint(49, 865),
    _AutismDevProfileScalePoint(48, 910),
    _AutismDevProfileScalePoint(47, 938),
    _AutismDevProfileScalePoint(39, 965),
    _AutismDevProfileScalePoint(35, 1027),
    _AutismDevProfileScalePoint(34, 1055),
    _AutismDevProfileScalePoint(33, 1130),
    _AutismDevProfileScalePoint(24, 1178),
    _AutismDevProfileScalePoint(23, 1211),
    _AutismDevProfileScalePoint(22, 1240),
    _AutismDevProfileScalePoint(21, 1262),
    _AutismDevProfileScalePoint(20, 1280),
    _AutismDevProfileScalePoint(11, 1375),
    _AutismDevProfileScalePoint(9, 1419),
    _AutismDevProfileScalePoint(4, 1435),
    _AutismDevProfileScalePoint(3, 1470),
    _AutismDevProfileScalePoint(2, 1500),
    _AutismDevProfileScalePoint(1, 1530),
    _AutismDevProfileScalePoint(0, 1542),
  ],
  '语言与沟通': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(79, 562),
    _AutismDevProfileScalePoint(76, 760),
    _AutismDevProfileScalePoint(67, 955),
    _AutismDevProfileScalePoint(53, 1090),
    _AutismDevProfileScalePoint(52, 1135),
    _AutismDevProfileScalePoint(36, 1268),
    _AutismDevProfileScalePoint(27, 1360),
    _AutismDevProfileScalePoint(21, 1415),
    _AutismDevProfileScalePoint(18, 1480),
    _AutismDevProfileScalePoint(8, 1492),
    _AutismDevProfileScalePoint(2, 1512),
    _AutismDevProfileScalePoint(1, 1530),
    _AutismDevProfileScalePoint(0, 1544),
  ],
  '认知': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(55, 530),
    _AutismDevProfileScalePoint(50, 586),
    _AutismDevProfileScalePoint(42, 758),
    _AutismDevProfileScalePoint(30, 940),
    _AutismDevProfileScalePoint(20, 1118),
    _AutismDevProfileScalePoint(10, 1270),
    _AutismDevProfileScalePoint(9, 1315),
    _AutismDevProfileScalePoint(5, 1362),
    _AutismDevProfileScalePoint(4, 1388),
    _AutismDevProfileScalePoint(2, 1426),
    _AutismDevProfileScalePoint(1, 1455),
    _AutismDevProfileScalePoint(0, 1498),
  ],
  '社会交往': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(47, 558),
    _AutismDevProfileScalePoint(45, 758),
    _AutismDevProfileScalePoint(40, 910),
    _AutismDevProfileScalePoint(30, 1090),
    _AutismDevProfileScalePoint(24, 1225),
    _AutismDevProfileScalePoint(19, 1265),
    _AutismDevProfileScalePoint(15, 1315),
    _AutismDevProfileScalePoint(14, 1355),
    _AutismDevProfileScalePoint(11, 1450),
    _AutismDevProfileScalePoint(4, 1498),
    _AutismDevProfileScalePoint(1, 1530),
    _AutismDevProfileScalePoint(0, 1546),
  ],
  '生活自理': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(67, 530),
    _AutismDevProfileScalePoint(62, 758),
    _AutismDevProfileScalePoint(46, 940),
    _AutismDevProfileScalePoint(34, 1090),
    _AutismDevProfileScalePoint(33, 1120),
    _AutismDevProfileScalePoint(18, 1178),
    _AutismDevProfileScalePoint(15, 1265),
    _AutismDevProfileScalePoint(12, 1315),
    _AutismDevProfileScalePoint(8, 1350),
    _AutismDevProfileScalePoint(6, 1380),
    _AutismDevProfileScalePoint(5, 1408),
    _AutismDevProfileScalePoint(3, 1450),
    _AutismDevProfileScalePoint(2, 1495),
    _AutismDevProfileScalePoint(1, 1512),
    _AutismDevProfileScalePoint(0, 1545),
  ],
  '发展分数': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(441, 532),
    _AutismDevProfileScalePoint(421, 562),
    _AutismDevProfileScalePoint(416, 592),
    _AutismDevProfileScalePoint(405, 760),
    _AutismDevProfileScalePoint(330, 790),
    _AutismDevProfileScalePoint(329, 820),
    _AutismDevProfileScalePoint(328, 862),
    _AutismDevProfileScalePoint(323, 910),
    _AutismDevProfileScalePoint(312, 940),
    _AutismDevProfileScalePoint(267, 956),
    _AutismDevProfileScalePoint(253, 970),
    _AutismDevProfileScalePoint(249, 1000),
    _AutismDevProfileScalePoint(248, 1030),
    _AutismDevProfileScalePoint(244, 1062),
    _AutismDevProfileScalePoint(243, 1092),
    _AutismDevProfileScalePoint(234, 1125),
    _AutismDevProfileScalePoint(192, 1138),
    _AutismDevProfileScalePoint(167, 1168),
    _AutismDevProfileScalePoint(163, 1182),
    _AutismDevProfileScalePoint(160, 1210),
    _AutismDevProfileScalePoint(157, 1226),
    _AutismDevProfileScalePoint(152, 1240),
    _AutismDevProfileScalePoint(149, 1270),
    _AutismDevProfileScalePoint(123, 1318),
    _AutismDevProfileScalePoint(93, 1355),
    _AutismDevProfileScalePoint(89, 1372),
    _AutismDevProfileScalePoint(79, 1388),
    _AutismDevProfileScalePoint(75, 1398),
    _AutismDevProfileScalePoint(73, 1418),
    _AutismDevProfileScalePoint(68, 1432),
    _AutismDevProfileScalePoint(58, 1446),
    _AutismDevProfileScalePoint(51, 1462),
    _AutismDevProfileScalePoint(28, 1480),
    _AutismDevProfileScalePoint(27, 1492),
    _AutismDevProfileScalePoint(26, 1500),
    _AutismDevProfileScalePoint(16, 1520),
    _AutismDevProfileScalePoint(9, 1534),
    _AutismDevProfileScalePoint(2, 1548),
    _AutismDevProfileScalePoint(1, 1560),
    _AutismDevProfileScalePoint(0, 1572),
  ],
};

class _AutismDevBehaviorProfilePainter extends CustomPainter {
  const _AutismDevBehaviorProfilePainter({required this.itemLevels});

  final List<String> itemLevels;

  static const double _figureWidth = 1488;
  static const double _figureHeight = 2103;
  static const List<Offset> _behaviorSPoints = <Offset>[
    Offset(745.0, 735.0),
    Offset(775.0, 737.0),
    Offset(796.0, 741.0),
    Offset(821.0, 748.0),
    Offset(849.0, 759.0),
    Offset(869.0, 770.0),
    Offset(889.0, 786.0),
    Offset(906.0, 801.0),
    Offset(925.0, 822.0),
    Offset(938.0, 839.0),
    Offset(947.0, 861.0),
    Offset(956.0, 884.0),
    Offset(965.0, 909.0),
    Offset(965.0, 930.0),
    Offset(963.0, 954.0),
    Offset(960.0, 977.0),
    Offset(953.0, 1002.0),
    Offset(944.0, 1025.0),
    Offset(930.0, 1046.0),
    Offset(913.0, 1069.0),
    Offset(895.0, 1086.0),
    Offset(876.0, 1101.0),
    Offset(851.0, 1116.0),
    Offset(827.0, 1127.0),
    Offset(801.0, 1135.0),
    Offset(777.0, 1140.0),
    Offset(748.0, 1142.0),
    Offset(723.0, 1141.0),
    Offset(693.0, 1137.0),
    Offset(670.0, 1130.0),
    Offset(645.0, 1120.0),
    Offset(618.0, 1105.0),
    Offset(600.0, 1091.0),
    Offset(580.0, 1073.0),
    Offset(563.0, 1053.0),
    Offset(549.0, 1032.0),
    Offset(537.0, 1006.0),
    Offset(530.0, 984.0),
    Offset(526.0, 957.0),
    Offset(524.0, 936.0),
    Offset(522.0, 910.0),
    Offset(533.0, 884.0),
    Offset(542.0, 860.0),
    Offset(550.0, 837.0),
    Offset(566.0, 820.0),
    Offset(585.0, 798.0),
    Offset(601.0, 784.0),
    Offset(626.0, 767.0),
    Offset(647.0, 754.0),
    Offset(669.0, 747.0),
    Offset(697.0, 739.0),
    Offset(722.0, 735.0),
  ];

  static const List<Offset> _behaviorMPoints = <Offset>[
    Offset(746.2, 628.9),
    Offset(789.2, 631.1),
    Offset(824.4, 636.9),
    Offset(863.6, 647.7),
    Offset(902.8, 664.1),
    Offset(934.8, 681.8),
    Offset(967.5, 704.9),
    Offset(993.5, 729.5),
    Offset(1020.1, 760.9),
    Offset(1039.5, 789.1),
    Offset(1057.5, 823.0),
    Offset(1070.0, 858.4),
    Offset(1077.2, 896.0),
    Offset(1080.8, 928.0),
    Offset(1079.4, 964.7),
    Offset(1073.2, 1001.7),
    Offset(1062.1, 1038.4),
    Offset(1046.4, 1073.0),
    Offset(1026.1, 1105.5),
    Offset(1000.9, 1136.8),
    Offset(974.0, 1163.7),
    Offset(943.2, 1186.8),
    Offset(907.3, 1208.1),
    Offset(869.3, 1224.2),
    Offset(831.5, 1236.3),
    Offset(792.4, 1242.7),
    Offset(749.9, 1245.0),
    Offset(710.7, 1243.3),
    Offset(669.5, 1237.2),
    Offset(631.5, 1226.5),
    Offset(595.4, 1211.5),
    Offset(556.6, 1190.0),
    Offset(527.8, 1168.6),
    Offset(498.6, 1141.2),
    Offset(473.3, 1111.6),
    Offset(451.9, 1079.6),
    Offset(434.4, 1042.8),
    Offset(423.3, 1007.4),
    Offset(416.6, 970.0),
    Offset(414.2, 934.6),
    Offset(417.6, 897.7),
    Offset(425.2, 858.1),
    Offset(437.8, 822.4),
    Offset(455.2, 789.4),
    Offset(474.6, 759.1),
    Offset(501.2, 728.3),
    Offset(528.8, 705.1),
    Offset(563.3, 680.5),
    Offset(596.0, 663.1),
    Offset(631.2, 648.1),
    Offset(670.3, 637.3),
    Offset(709.6, 630.8),
  ];

  static const List<Offset> _behaviorAPoints = <Offset>[
    Offset(747.5, 516.9),
    Offset(804.0, 521.1),
    Offset(853.9, 528.8),
    Offset(907.4, 544.6),
    Offset(958.3, 566.2),
    Offset(1002.7, 590.8),
    Offset(1046.4, 623.3),
    Offset(1081.8, 657.4),
    Offset(1118.1, 697.9),
    Offset(1144.5, 737.5),
    Offset(1167.2, 785.3),
    Offset(1183.2, 833.0),
    Offset(1192.9, 882.6),
    Offset(1196.3, 926.0),
    Offset(1193.9, 975.2),
    Offset(1185.1, 1026.1),
    Offset(1169.8, 1074.3),
    Offset(1147.8, 1120.5),
    Offset(1120.5, 1163.9),
    Offset(1085.6, 1202.2),
    Offset(1051.7, 1240.1),
    Offset(1008.9, 1270.6),
    Offset(961.5, 1296.8),
    Offset(911.0, 1320.0),
    Offset(861.2, 1334.9),
    Offset(807.6, 1344.1),
    Offset(751.8, 1348.0),
    Offset(698.3, 1346.1),
    Offset(646.0, 1337.5),
    Offset(593.0, 1323.1),
    Offset(545.4, 1303.8),
    Offset(494.5, 1275.9),
    Offset(455.6, 1246.2),
    Offset(415.8, 1210.6),
    Offset(380.4, 1172.3),
    Offset(351.8, 1128.7),
    Offset(328.0, 1081.0),
    Offset(311.5, 1031.9),
    Offset(301.9, 983.6),
    Offset(298.7, 933.1),
    Offset(301.9, 884.1),
    Offset(312.9, 831.1),
    Offset(329.2, 783.2),
    Offset(351.5, 737.3),
    Offset(379.3, 695.6),
    Offset(413.9, 655.7),
    Offset(451.8, 621.0),
    Offset(497.0, 589.0),
    Offset(540.7, 564.6),
    Offset(591.2, 543.5),
    Offset(641.9, 529.0),
    Offset(695.5, 520.2),
  ];

  @override
  void paint(Canvas canvas, Size size) {
    final int itemCount = itemLevels.length < _autismDevBehaviorItemCount
        ? itemLevels.length
        : _autismDevBehaviorItemCount;
    final List<Offset> points = <Offset>[
      for (int index = 0; index < itemCount; index++)
        _behaviorPoint(
          size,
          itemNo: index + 1,
          level: itemLevels[index],
        ),
    ];
    if (points.isEmpty) {
      return;
    }
    final Path outerPath = Path()..moveTo(points.first.dx, points.first.dy);
    for (final Offset point in points.skip(1)) {
      outerPath.lineTo(point.dx, point.dy);
    }
    outerPath.close();
    final Path centerCutoutPath = _pathFromSourcePoints(size, _behaviorSPoints);
    final Path sunPath = Path()
      ..fillType = PathFillType.evenOdd
      ..addPath(outerPath, Offset.zero)
      ..addPath(centerCutoutPath, Offset.zero);

    final double strokeWidth = size.width / 420;
    final double dotRadius = size.width / 300;
    final Color profileColor = _ReportTheme.rose;
    final Paint fillPaint = Paint()
      ..color = profileColor.withOpacity(.22)
      ..style = PaintingStyle.fill;
    final Paint strokePaint = Paint()
      ..color = profileColor.withOpacity(.9)
      ..strokeWidth = strokeWidth
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;
    final Paint innerStrokePaint = Paint()
      ..color = profileColor.withOpacity(.34)
      ..strokeWidth = strokeWidth * .6
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;
    final Paint dotHaloPaint = Paint()
      ..color = Colors.white.withOpacity(.94)
      ..style = PaintingStyle.fill;
    final Paint dotPaint = Paint()
      ..color = profileColor
      ..style = PaintingStyle.fill;
    final Paint dotRingPaint = Paint()
      ..color = profileColor.withOpacity(.82)
      ..strokeWidth = strokeWidth * .72
      ..style = PaintingStyle.stroke;

    canvas.drawPath(sunPath, fillPaint);
    canvas.drawPath(centerCutoutPath, innerStrokePaint);
    canvas.drawPath(outerPath, strokePaint);
    for (final Offset point in points) {
      canvas.drawCircle(point, dotRadius * 1.65, dotHaloPaint);
      canvas.drawCircle(point, dotRadius, dotPaint);
      canvas.drawCircle(point, dotRadius * 1.65, dotRingPaint);
    }
  }

  static Offset _point(Size size, double x, double y) {
    return Offset(
      x / _figureWidth * size.width,
      y / _figureHeight * size.height,
    );
  }

  static Path _pathFromSourcePoints(Size size, List<Offset> sourcePoints) {
    final Path path = Path();
    if (sourcePoints.isEmpty) {
      return path;
    }
    final Offset first =
        _point(size, sourcePoints.first.dx, sourcePoints.first.dy);
    path.moveTo(first.dx, first.dy);
    for (final Offset sourcePoint in sourcePoints.skip(1)) {
      final Offset point = _point(size, sourcePoint.dx, sourcePoint.dy);
      path.lineTo(point.dx, point.dy);
    }
    return path..close();
  }

  static Offset _behaviorPoint(
    Size size, {
    required int itemNo,
    required String level,
  }) {
    final int index = itemNo - 1;
    final List<Offset> sourcePoints = _pointsForLevel(level);
    return _point(size, sourcePoints[index].dx, sourcePoints[index].dy);
  }

  static List<Offset> _pointsForLevel(String level) {
    switch (level) {
      case 'S':
        return _behaviorSPoints;
      case 'M':
        return _behaviorMPoints;
      case 'A':
      default:
        return _behaviorAPoints;
    }
  }

  @override
  bool shouldRepaint(covariant _AutismDevBehaviorProfilePainter oldDelegate) {
    if (oldDelegate.itemLevels.length != itemLevels.length) {
      return true;
    }
    for (int index = 0; index < itemLevels.length; index++) {
      if (oldDelegate.itemLevels[index] != itemLevels[index]) {
        return true;
      }
    }
    return false;
  }
}

class _AutismDevTableFrame extends StatelessWidget {
  const _AutismDevTableFrame({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(10),
      child: child,
    );
  }
}

TableRow _autismDevTableHeaderRow(List<String> values, {bool tall = false}) {
  return TableRow(
    decoration: const BoxDecoration(color: Color(0xFFFFF8F2)),
    children: <Widget>[
      for (final String value in values)
        _autismDevTableCell(
          value,
          emph: true,
          header: true,
          minHeight: tall ? 44 : 38,
        ),
    ],
  );
}

TableRow _autismDevTableRow(List<String> values, {bool emphFirst = false}) {
  return TableRow(
    children: <Widget>[
      for (int index = 0; index < values.length; index++)
        _autismDevTableCell(
          values[index],
          emph: emphFirst && index == 0,
        ),
    ],
  );
}

Widget _autismDevTableCell(
  String value, {
  bool emph = false,
  bool header = false,
  double minHeight = 36,
}) {
  return Container(
    constraints: BoxConstraints(minHeight: minHeight),
    alignment: Alignment.center,
    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
    color: header ? const Color(0xFFFFF8F2) : Colors.white,
    child: Text(
      value,
      textAlign: TextAlign.center,
      style: TextStyle(
        color: emph ? _ReportTheme.ink : _ReportTheme.text,
        fontSize: header ? 13 : 12,
        height: 1.35,
        fontWeight: emph ? FontWeight.w900 : FontWeight.w800,
      ),
    ),
  );
}

class _AutismDevInfoItem {
  const _AutismDevInfoItem(this.label, this.value);

  final String label;
  final String value;
}

class _AutismDevReportTabSpec {
  const _AutismDevReportTabSpec(this.label, this.tab);

  final String label;
  final _AutismDevReportTab tab;
}

class _AutismDevSummaryItem {
  const _AutismDevSummaryItem(this.title, this.description, this.icon);

  final String title;
  final String description;
  final IconData icon;
}

class _AutismDevDevelopmentScore {
  const _AutismDevDevelopmentScore({
    required this.label,
    required this.p,
    required this.e,
    required this.f,
    required this.x,
    required this.total,
  });

  final String label;
  final int p;
  final int e;
  final int f;
  final int x;
  final int total;

  int get supportCount => e + f + x;
}

class _AutismDevBehaviorScore {
  const _AutismDevBehaviorScore({
    required this.label,
    required this.a,
    required this.m,
    required this.s,
  });

  final String label;
  final int a;
  final int m;
  final int s;
}

class _AutismDevAnalysisItem {
  const _AutismDevAnalysisItem({
    required this.domain,
    required this.status,
    required this.strength,
    required this.target,
  });

  final String domain;
  final String status;
  final String strength;
  final String target;
}

const List<_AutismDevDevelopmentScore> _autismDevDevelopmentScores =
    <_AutismDevDevelopmentScore>[
  _AutismDevDevelopmentScore(
    label: '语言与沟通',
    p: 29,
    e: 18,
    f: 25,
    x: 7,
    total: 79,
  ),
  _AutismDevDevelopmentScore(
    label: '认知',
    p: 34,
    e: 10,
    f: 9,
    x: 2,
    total: 55,
  ),
  _AutismDevDevelopmentScore(
    label: '生活自理',
    p: 41,
    e: 12,
    f: 10,
    x: 4,
    total: 67,
  ),
  _AutismDevDevelopmentScore(
    label: '感知觉',
    p: 38,
    e: 8,
    f: 7,
    x: 2,
    total: 55,
  ),
  _AutismDevDevelopmentScore(
    label: '粗大动作',
    p: 44,
    e: 13,
    f: 11,
    x: 4,
    total: 72,
  ),
  _AutismDevDevelopmentScore(
    label: '精细动作',
    p: 35,
    e: 14,
    f: 13,
    x: 4,
    total: 66,
  ),
  _AutismDevDevelopmentScore(
    label: '社会交往',
    p: 25,
    e: 9,
    f: 11,
    x: 2,
    total: 47,
  ),
];

const int _autismDevBehaviorFirstItemNo = 442;
const int _autismDevBehaviorItemCount = 52;

const List<String> _autismDevBehaviorFallbackItemLevels = <String>[
  'A',
  'M',
  'A',
  'M',
  'A',
  'A',
  'M',
  'A',
  'S',
  'M',
  'A',
  'M',
  'A',
  'A',
  'A',
  'M',
  'A',
  'M',
  'S',
  'A',
  'M',
  'A',
  'A',
  'M',
  'A',
  'M',
  'A',
  'A',
  'M',
  'A',
  'S',
  'M',
  'A',
  'A',
  'A',
  'M',
  'A',
  'M',
  'A',
  'M',
  'S',
  'A',
  'M',
  'A',
  'A',
  'M',
  'A',
  'M',
  'S',
  'A',
  'M',
  'A',
];

List<String> _autismDevBehaviorItemLevelsFromScores(
  Map<int, String> itemScoreLabels,
) {
  if (itemScoreLabels.isEmpty) {
    return _autismDevBehaviorFallbackItemLevels;
  }
  return <String>[
    for (int index = 0; index < _autismDevBehaviorItemCount; index++)
      _autismDevBehaviorLevel(
            itemScoreLabels[_autismDevBehaviorFirstItemNo + index] ??
                itemScoreLabels[index + 1],
          ) ??
          'A',
  ];
}

String? _autismDevBehaviorLevel(String? raw) {
  final String score = (raw ?? '').trim().toUpperCase();
  if (score == 'A' || score.contains('无') || score.contains('没有')) {
    return 'A';
  }
  if (score == 'M' || score.contains('轻度')) {
    return 'M';
  }
  if (score == 'S' || score.contains('重度')) {
    return 'S';
  }
  return null;
}

const List<_AutismDevBehaviorScore> _autismDevBehaviorScores =
    <_AutismDevBehaviorScore>[
  _AutismDevBehaviorScore(label: '依附情绪行为', a: 4, m: 2, s: 1),
  _AutismDevBehaviorScore(label: '情绪理解', a: 3, m: 2, s: 0),
  _AutismDevBehaviorScore(label: '情绪表达与调节', a: 3, m: 3, s: 1),
  _AutismDevBehaviorScore(label: '关系与情感', a: 4, m: 2, s: 1),
  _AutismDevBehaviorScore(label: '对物品的兴趣', a: 5, m: 3, s: 1),
  _AutismDevBehaviorScore(label: '感觉偏好', a: 5, m: 4, s: 1),
  _AutismDevBehaviorScore(label: '特殊行为', a: 4, m: 3, s: 0),
];

AutismDevResultAnalysis _emptyAutismDevResultAnalysis() {
  return AutismDevResultAnalysis(
    title: '孤独症儿童评估结果分析表',
    rows: _autismDevResultAnalysisDomains
        .map(
          (String domain) => AutismDevResultAnalysisRow(domain: domain),
        )
        .toList(),
  );
}

AutismDevResultAnalysis _mergeAutismDevResultAnalysis(
  AutismDevResultAnalysis? source,
) {
  final AutismDevResultAnalysis fallback = _emptyAutismDevResultAnalysis();
  if (source == null || source.rows.isEmpty) {
    return fallback;
  }
  final List<AutismDevResultAnalysisRow> nextRows =
      <AutismDevResultAnalysisRow>[];
  final Set<String> seenDomains = <String>{};
  for (final AutismDevResultAnalysisRow row in source.rows) {
    final String canonical =
        _canonicalAutismDevResultAnalysisDomain(row.domain);
    if (canonical.isEmpty) {
      continue;
    }
    if (seenDomains.add(canonical)) {
      nextRows.add(row.copyWith(domain: canonical));
    }
  }
  if (nextRows.isEmpty) {
    return fallback;
  }
  return AutismDevResultAnalysis(
    title: source.title.trim().isEmpty ? fallback.title : source.title.trim(),
    model: source.model,
    generatedBy: source.generatedBy,
    generatedAt: source.generatedAt,
    rows: nextRows,
  );
}

List<AutismDevResultAnalysisRow> _normalizeAutismDevResultAnalysisRows(
  List<AutismDevResultAnalysisRow> rows,
) {
  return _mergeAutismDevResultAnalysis(
    AutismDevResultAnalysis(title: '', rows: rows),
  ).rows;
}

String _canonicalAutismDevResultAnalysisDomain(String value) {
  final String normalized =
      value.replaceAll(RegExp(r'\s+'), '').replaceAll('和', '与').trim();
  for (final String domain in _autismDevResultAnalysisDomains) {
    final String candidate =
        domain.replaceAll(RegExp(r'\s+'), '').replaceAll('和', '与').trim();
    if (normalized == candidate) {
      return domain;
    }
  }
  return '';
}

class _AutismDevAnalysisReadableStream {
  const _AutismDevAnalysisReadableStream({required this.content});

  factory _AutismDevAnalysisReadableStream.fromRaw(String raw) {
    final String content =
        _formatAutismDevAnalysisStreamTextIncrementally(raw).trimRight();
    return _AutismDevAnalysisReadableStream(
      content: content.isEmpty ? '正在连接AI生成服务，准备读取评估记录...' : content,
    );
  }

  final String content;
}

String _formatAutismDevAnalysisStreamTextIncrementally(String raw) {
  final String text = raw.trim();
  if (text.isEmpty) {
    return '';
  }
  final StringBuffer output = StringBuffer();
  final StringBuffer token = StringBuffer();
  _AutismDevReadableJsonMode mode = _AutismDevReadableJsonMode.outside;
  String currentKey = '';
  String visibleKey = '';
  bool expectingValue = false;
  bool escaping = false;
  bool lastWasNewline = true;

  void writeText(String value) {
    if (value.isEmpty) {
      return;
    }
    output.write(value);
    lastWasNewline = value.endsWith('\n');
  }

  void startVisibleField(String key) {
    if (output.isNotEmpty && !lastWasNewline) {
      writeText('\n');
    }
    if (key == 'domain') {
      if (output.isNotEmpty) {
        writeText('\n');
      }
      writeText('【');
      return;
    }
    writeText('${_autismDevAnalysisStreamLabel(key)}：');
  }

  void endVisibleField(String key) {
    if (key == 'domain') {
      writeText('】\n');
      return;
    }
    if (output.isNotEmpty && !lastWasNewline) {
      writeText('\n');
    }
  }

  void writeEscaped(String char) {
    switch (char) {
      case 'n':
      case 'r':
      case 't':
        writeText(' ');
      case '"':
        writeText('"');
      case '/':
        writeText('/');
      case r'\':
        writeText(r'\');
      default:
        writeText(char);
    }
  }

  for (final int rune in text.runes) {
    final String char = String.fromCharCode(rune);
    if (escaping) {
      if (mode == _AutismDevReadableJsonMode.visibleValue) {
        writeEscaped(char);
      } else if (mode == _AutismDevReadableJsonMode.key) {
        token.write(char);
      }
      escaping = false;
      continue;
    }
    if (char == r'\') {
      escaping = true;
      continue;
    }
    switch (mode) {
      case _AutismDevReadableJsonMode.outside:
        if (char == '"') {
          token.clear();
          if (expectingValue) {
            if (_isAutismDevAnalysisReadableField(currentKey)) {
              visibleKey = currentKey;
              startVisibleField(visibleKey);
              mode = _AutismDevReadableJsonMode.visibleValue;
            } else {
              mode = _AutismDevReadableJsonMode.hiddenValue;
            }
          } else {
            mode = _AutismDevReadableJsonMode.key;
          }
        } else if (char == ':') {
          expectingValue = currentKey.isNotEmpty;
        } else if (char == ',' || char == '}' || char == ']') {
          if (expectingValue) {
            expectingValue = false;
            currentKey = '';
          }
        }
      case _AutismDevReadableJsonMode.key:
        if (char == '"') {
          currentKey = token.toString();
          token.clear();
          mode = _AutismDevReadableJsonMode.outside;
        } else {
          token.write(char);
        }
      case _AutismDevReadableJsonMode.visibleValue:
        if (char == '"') {
          endVisibleField(visibleKey);
          mode = _AutismDevReadableJsonMode.outside;
          expectingValue = false;
          currentKey = '';
          visibleKey = '';
        } else {
          writeText(char);
        }
      case _AutismDevReadableJsonMode.hiddenValue:
        if (char == '"') {
          mode = _AutismDevReadableJsonMode.outside;
          expectingValue = false;
          currentKey = '';
        }
    }
  }

  final String normalized = output.toString().trimRight();
  if (normalized.isNotEmpty) {
    return normalized;
  }
  return text
      .replaceAll('{', '')
      .replaceAll('}', '')
      .replaceAll('[', '')
      .replaceAll(']', '')
      .replaceAll('"', '')
      .replaceAll(',', '')
      .trim();
}

enum _AutismDevReadableJsonMode { outside, key, visibleValue, hiddenValue }

bool _isAutismDevAnalysisReadableField(String key) {
  return const <String>{
    'domain',
    'status',
    'strengths',
    'weaknesses',
    'targets',
  }.contains(key);
}

String _autismDevAnalysisStreamLabel(String key) {
  return switch (key) {
    'status' => '能力现状描述',
    'strengths' => '优势',
    'weaknesses' => '劣势',
    'targets' => '训练目标',
    _ => key,
  };
}

double _autismDevResultAnalysisStreamProgress(String text) {
  final String trimmed = text.trim();
  if (trimmed.isEmpty) {
    return .12;
  }
  final double length = trimmed.runes.length.toDouble();
  const double floor = .12;
  const double ceiling = .975;
  if (length <= 360) {
    return _autismDevLerpProgress(floor, .32, length / 360);
  }
  if (length <= 1600) {
    return _autismDevLerpProgress(.32, .88, (length - 360) / 1240);
  }
  final double tail = 1 - math.exp(-(length - 1600) / 720);
  return _autismDevLerpProgress(.88, ceiling, tail.clamp(0, 1));
}

double _autismDevLerpProgress(double start, double end, double t) {
  final double normalized = t.clamp(0, 1).toDouble();
  return start + (end - start) * normalized;
}

const List<String> _autismDevResultAnalysisDomains = <String>[
  '感知觉',
  '粗大动作',
  '精细动作',
  '语言与沟通',
  '认知',
  '社会交往',
  '生活自理',
];

const List<_AutismDevAnalysisItem> _autismDevAnalysisItems =
    <_AutismDevAnalysisItem>[
  _AutismDevAnalysisItem(
    domain: '感知觉',
    status: '视觉、听觉反应较稳定，触觉与味觉辨别任务仍需持续观察。',
    strength: '优势：熟悉刺激反应明确。劣势：复杂辨别与记忆任务稳定性不足。',
    target: '优先安排触觉辨别、味觉辨别与多感官配对训练。',
  ),
  _AutismDevAnalysisItem(
    domain: '粗大动作',
    status: '基本姿势与移动能力较好，球类操作与平衡动作存在波动。',
    strength: '优势：移动类项目完成度较高。劣势：抛接、踢、拍等协调动作较弱。',
    target: '设置抛接球、平衡木行走与双手协调游戏目标。',
  ),
  _AutismDevAnalysisItem(
    domain: '精细动作',
    status: '基础抓握和摆弄物品较稳定，握笔写画及工具使用需加强。',
    strength: '优势：基本操作可配合。劣势：双手配合和精细控制持续时间短。',
    target: '围绕穿珠、剪纸、仿画和工具使用建立分步训练。',
  ),
  _AutismDevAnalysisItem(
    domain: '语言与沟通',
    status: '理解简单名称和动作指令较好，主动表达和复述能力不足。',
    strength: '优势：名称指认有基础。劣势：短语、句子和主动提问较少。',
    target: '以表达要求、回答问题、短语扩展和主动提问作为核心目标。',
  ),
  _AutismDevAnalysisItem(
    domain: '认知',
    status: '配对、分类和部分颜色概念可完成，数量及关系概念仍需支持。',
    strength: '优势：具体物品配对较好。劣势：抽象概念和数概念掌握不稳定。',
    target: '训练大小、多少、长短、颜色分类和1-5数量操作。',
  ),
  _AutismDevAnalysisItem(
    domain: '社会交往',
    status: '熟悉情境中可回应互动，陌生情境与社交礼仪需提示。',
    strength: '优势：近距离互动可建立。劣势：主动打招呼、告别与感谢不足。',
    target: '设计打招呼、告别、请求帮助和表示感谢的情境练习。',
  ),
  _AutismDevAnalysisItem(
    domain: '生活自理',
    status: '进食和部分家居自理较稳定，穿衣梳洗流程仍需辅助。',
    strength: '优势：日常熟悉流程接受度较好。劣势：多步骤任务独立性不足。',
    target: '拆分穿衣、梳洗、物品归位和收拾餐具等连续任务。',
  ),
  _AutismDevAnalysisItem(
    domain: '情绪与行为',
    status: '轻度异常项目集中在情绪调节、感觉偏好和特殊行为。',
    strength: '优势：多数项目无重度异常。劣势：转变适应与感官偏好需关注。',
    target: '建立等待、转换活动、情绪表达和感官调节策略。',
  ),
];
