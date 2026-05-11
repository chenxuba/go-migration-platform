part of 'iep_center_page.dart';

enum _PlanTabTone { primary, week }

class _PlanTab extends StatelessWidget {
  const _PlanTab({
    required this.text,
    required this.width,
    this.active = false,
    this.generated = false,
    this.activeTone = _PlanTabTone.primary,
    this.rightGap = 6,
    this.onTap,
  });

  final String text;
  final double width;
  final bool active;
  final bool generated;
  final _PlanTabTone activeTone;
  final double rightGap;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final Color activeBg =
        activeTone == _PlanTabTone.week ? _IepColors.orange : _IepColors.ink;
    final Color inactiveBg = activeTone == _PlanTabTone.week
        ? const Color(0xFFFFF3EC)
        : const Color(0xFFFFFAF6);
    final Color inactiveText = activeTone == _PlanTabTone.week
        ? _IepColors.orangeDeep
        : _IepColors.text;
    final Color borderColor = activeTone == _PlanTabTone.week
        ? const Color(0xFFFFD8C3)
        : _IepColors.lightLine;
    final Color dotColor =
        generated ? const Color(0xFF6F9F70) : const Color(0xFFD0D6DE);

    return Padding(
      padding: EdgeInsets.only(right: rightGap),
      child: Material(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(15),
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(15),
          child: Container(
            width: width,
            height: 30,
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: active ? activeBg : inactiveBg,
              borderRadius: BorderRadius.circular(15),
              border: active ? null : Border.all(color: borderColor),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              mainAxisSize: MainAxisSize.min,
              children: <Widget>[
                Container(
                  width: 7,
                  height: 7,
                  decoration: BoxDecoration(
                    color: dotColor,
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(width: 6),
                Flexible(
                  child: Text(
                    text,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: active ? Colors.white : inactiveText,
                      fontSize: 11,
                      fontWeight: FontWeight.w900,
                    ),
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

class _PlanNavLabel extends StatelessWidget {
  const _PlanNavLabel({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 6, right: 6),
      child: Text(
        text,
        style: const TextStyle(
          color: _IepColors.muted,
          fontSize: 11,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _IepTablePreview extends StatelessWidget {
  const _IepTablePreview({
    required this.previewMode,
    required this.month,
    required this.weekNumber,
    required this.periodStart,
    required this.periodMonthCount,
    required this.record,
    required this.plan,
    required this.monthPlan,
    required this.weekPlan,
    required this.loading,
    required this.bootstrapLoading,
    required this.generatingPlan,
    required this.generationStatus,
    required this.generationText,
    required this.generationProgress,
    required this.generationCostAmountCny,
    required this.error,
    required this.onRetry,
    required this.onGeneratePlan,
    required this.totalPlanDomains,
    required this.selectedGoal,
    required this.onGoalTap,
    required this.onClearSelectedGoal,
  });

  final _IepPreviewMode previewMode;
  final String month;
  final int weekNumber;
  final DateTime periodStart;
  final int periodMonthCount;
  final IepAssessmentRecordSummary? record;
  final IepPlan? plan;
  final IepMonthlyPlan? monthPlan;
  final IepWeeklyPlan? weekPlan;
  final bool loading;
  final bool bootstrapLoading;
  final bool generatingPlan;
  final String generationStatus;
  final String generationText;
  final double generationProgress;
  final double generationCostAmountCny;
  final String error;
  final VoidCallback onRetry;
  final VoidCallback onGeneratePlan;
  final List<_DocDomainData> totalPlanDomains;
  final _GoalEditRequest? selectedGoal;
  final ValueChanged<_GoalEditRequest> onGoalTap;
  final VoidCallback onClearSelectedGoal;

  @override
  Widget build(BuildContext context) {
    final DateTime monthDate = _monthDateFromLabel(
      periodStart,
      periodMonthCount,
      month,
    );
    final DateTimeRange monthRange = _monthRangeInPeriod(
      periodStart: periodStart,
      monthCount: periodMonthCount,
      monthDate: monthDate,
    );
    final List<DateTime> weekDates = _weekDatesInMonthRange(
      monthRange,
      weekNumber,
    );
    final String contentSignature = Object.hashAll(<Object?>[
      record?.id,
      plan?.title,
      plan?.meta.startDate,
      plan?.meta.endDate,
      plan?.rows.length,
      monthPlan?.title,
      monthPlan?.rows.length,
      weekPlan?.title,
      weekPlan?.rows.length,
    ]).toString();

    Widget child;
    final String generationPlanLabel = switch (previewMode) {
      _IepPreviewMode.total => 'IEP计划',
      _IepPreviewMode.month => '$month计划',
      _IepPreviewMode.week => '$month第$weekNumber周计划',
    };
    if (generatingPlan) {
      final IepAssessmentRecordSummary currentRecord = record!;
      child = _IepGenerationStreamPanel(
        studentName: currentRecord.studentName.trim().isEmpty
            ? '当前学员'
            : currentRecord.studentName.trim(),
        planLabel: generationPlanLabel,
        status: generationStatus,
        streamText: generationText,
        progress: generationProgress,
        costAmountCny: generationCostAmountCny,
      );
    } else if (bootstrapLoading) {
      child = const _IepWordTableSkeleton();
    } else if (record == null) {
      child = const _PlanStateView(
        icon: Icons.touch_app_rounded,
        title: '请选择左侧评估记录',
        message: '选择学员后会读取对应IEP计划',
      );
    } else if (loading && !generatingPlan) {
      child = const _IepPlanLoadingState();
    } else if (error.trim().isNotEmpty) {
      child = _PlanStateView(
        icon: Icons.wifi_off_rounded,
        title: 'IEP计划加载失败',
        message: error,
        actionLabel: '重试',
        onAction: onRetry,
      );
    } else if (plan == null || totalPlanDomains.isEmpty) {
      final IepAssessmentRecordSummary currentRecord = record!;
      child = _IepEmptyGenerateState(
        studentName: currentRecord.studentName.trim().isEmpty
            ? '当前学员'
            : currentRecord.studentName.trim(),
        generating: generatingPlan,
        statusText: generationStatus,
        onGenerate: onGeneratePlan,
      );
    } else {
      child = switch (previewMode) {
        _IepPreviewMode.month => monthPlan == null
            ? _PlanStateView(
                icon: Icons.calendar_month_rounded,
                title: '$month计划未生成',
                message: '可基于当前IEP总计划生成$month的月计划模板',
                actionLabel: 'AI生成',
                onAction: onGeneratePlan,
              )
            : _WordTableFrame(
                child: _MonthPlanTable(
                  month: month,
                  monthRange: monthRange,
                  plan: monthPlan,
                ),
              ),
        _IepPreviewMode.week => weekPlan == null
            ? _PlanStateView(
                icon: Icons.view_week_rounded,
                title: '$month第$weekNumber周计划未生成',
                message: '可基于当前IEP总计划或月计划生成本周计划模板',
                actionLabel: 'AI生成',
                onAction: onGeneratePlan,
              )
            : _WordTableFrame(
                child: _WeekPlanTable(
                  month: month,
                  weekNumber: weekNumber,
                  weekDates: weekDates,
                  plan: weekPlan,
                ),
              ),
        _IepPreviewMode.total => _WordTableFrame(
            child: _WordTable(
              periodText: _formatZhRange(
                periodStart,
                _periodEndFor(periodStart, periodMonthCount),
              ),
              plan: plan,
              domains: totalPlanDomains,
              selectedGoal: selectedGoal,
              onGoalTap: onGoalTap,
              onClearSelectedGoal: onClearSelectedGoal,
            ),
          ),
      };
    }

    return AnimatedSwitcher(
      duration: const Duration(milliseconds: 220),
      switchInCurve: Curves.easeOutCubic,
      switchOutCurve: Curves.easeInCubic,
      child: KeyedSubtree(
        key: ValueKey<String>(
          '${previewMode.name}|$month|$weekNumber|${generatingPlan ? 'generating' : 'stable'}|${loading ? 'loading' : 'ready'}|${error.trim()}|$contentSignature',
        ),
        child: child,
      ),
    );
  }
}

class _IepGenerationStreamPanel extends StatefulWidget {
  const _IepGenerationStreamPanel({
    required this.studentName,
    required this.planLabel,
    required this.status,
    required this.streamText,
    required this.progress,
    required this.costAmountCny,
  });

  final String studentName;
  final String planLabel;
  final String status;
  final String streamText;
  final double progress;
  final double costAmountCny;

  @override
  State<_IepGenerationStreamPanel> createState() =>
      _IepGenerationStreamPanelState();
}

class _IepGenerationStreamPanelState extends State<_IepGenerationStreamPanel> {
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
  void didUpdateWidget(covariant _IepGenerationStreamPanel oldWidget) {
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
    final double progress = widget.progress.clamp(0, 1).toDouble();
    final String status =
        widget.status.trim().isEmpty ? 'AI正在生成IEP计划' : widget.status.trim();
    final String? costText = widget.costAmountCny > 0
        ? '¥${widget.costAmountCny.toStringAsFixed(widget.costAmountCny >= 1 ? 2 : 4)}'
        : null;
    final _IepReadableStream readable =
        _IepReadableStream.fromRaw(widget.streamText);
    final String progressText = '${(progress * 100).round()}%';

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: const Color(0xFFFFD8C3)),
        boxShadow: _iepShadow(
          color: const Color(0x12B05F32),
          blur: 14,
          offset: const Offset(0, 6),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Padding(
            padding: const EdgeInsets.fromLTRB(18, 16, 18, 12),
            child: Row(
              children: <Widget>[
                const _IepHourglassLoader(size: 24),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        '正在生成 ${widget.studentName} 的${widget.planLabel}',
                        style: const TextStyle(
                          color: _IepColors.ink,
                          fontSize: 16,
                          fontWeight: FontWeight.w900,
                          height: 1.15,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Row(
                        children: <Widget>[
                          Text(
                            status,
                            style: const TextStyle(
                              color: _IepColors.text,
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
                                value: progress <= 0 ? null : progress,
                                backgroundColor: const Color(0xFFFFEEE4),
                                color: _IepColors.orange,
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
                                color: _IepColors.orangeDeep,
                                fontSize: 11,
                                fontWeight: FontWeight.w900,
                                height: 1,
                              ),
                            ),
                          ),
                          if (costText != null) ...<Widget>[
                            const SizedBox(width: 8),
                            Text(
                              costText,
                              style: const TextStyle(
                                color: _IepColors.orangeDeep,
                                fontSize: 12,
                                fontWeight: FontWeight.w900,
                                height: 1,
                              ),
                            ),
                          ],
                          const SizedBox(width: 8),
                          ConstrainedBox(
                            constraints: const BoxConstraints(maxWidth: 118),
                            child: Text(
                              costText == null ? '正在估算金额' : '完成后切换真实消费',
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              textAlign: TextAlign.right,
                              style: const TextStyle(
                                color: _IepColors.muted,
                                fontSize: 10.5,
                                fontWeight: FontWeight.w700,
                                height: 1.1,
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
                border: Border.all(color: _IepColors.lightLine),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  Row(
                    children: const <Widget>[
                      Icon(
                        Icons.auto_awesome_rounded,
                        size: 16,
                        color: _IepColors.orangeDeep,
                      ),
                      SizedBox(width: 6),
                      Text(
                        'AI正在整理可预览内容',
                        style: TextStyle(
                          color: _IepColors.ink,
                          fontSize: 12.5,
                          fontWeight: FontWeight.w900,
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
                            color: _IepColors.text,
                            fontSize: 12.5,
                            fontWeight: FontWeight.w700,
                            height: 1.42,
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
                  color: _IepColors.orangeDeep,
                  size: 16,
                ),
                SizedBox(width: 6),
                Expanded(
                  child: Text(
                    '生成完成后将自动保存草稿，并切换为正式IEP表格预览',
                    style: TextStyle(
                      color: _IepColors.muted,
                      fontSize: 11.5,
                      fontWeight: FontWeight.w800,
                    ),
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

class _IepReadableStream {
  const _IepReadableStream({
    required this.content,
  });

  factory _IepReadableStream.fromRaw(String raw) {
    final String content = _formatIepStreamTextIncrementally(raw).trimRight();
    final String visible = content.isEmpty ? '正在连接AI生成服务，准备读取评估记录...' : content;
    return _IepReadableStream(
      content: visible,
    );
  }

  final String content;
}

String _formatIepStreamTextIncrementally(String raw) {
  final String text = raw.trim();
  if (text.isEmpty) {
    return '';
  }
  final StringBuffer output = StringBuffer();
  final StringBuffer token = StringBuffer();
  _IepReadableJsonMode mode = _IepReadableJsonMode.outside;
  String currentKey = '';
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

  void startVisibleField() {
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
      case '\\':
        writeText('\\');
      default:
        writeText(char);
    }
  }

  for (final int rune in text.runes) {
    final String char = String.fromCharCode(rune);
    if (escaping) {
      if (mode == _IepReadableJsonMode.visibleValue) {
        writeEscaped(char);
      } else if (mode == _IepReadableJsonMode.key) {
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
      case _IepReadableJsonMode.outside:
        if (char == '"') {
          token.clear();
          if (expectingValue) {
            if (_isIepReadableField(currentKey)) {
              startVisibleField();
              mode = _IepReadableJsonMode.visibleValue;
            } else {
              mode = _IepReadableJsonMode.hiddenValue;
            }
          } else {
            mode = _IepReadableJsonMode.key;
          }
        } else if (char == ':') {
          expectingValue = currentKey.isNotEmpty;
        } else if (char == ',' || char == '}' || char == ']') {
          if (expectingValue) {
            expectingValue = false;
            currentKey = '';
          }
        }
      case _IepReadableJsonMode.key:
        if (char == '"') {
          currentKey = token.toString();
          token.clear();
          mode = _IepReadableJsonMode.outside;
        } else {
          token.write(char);
        }
      case _IepReadableJsonMode.visibleValue:
        if (char == '"') {
          mode = _IepReadableJsonMode.outside;
          expectingValue = false;
          currentKey = '';
          if (output.isNotEmpty && !lastWasNewline) {
            writeText('\n');
          }
        } else {
          writeText(char);
        }
      case _IepReadableJsonMode.hiddenValue:
        if (char == '"') {
          mode = _IepReadableJsonMode.outside;
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

enum _IepReadableJsonMode { outside, key, visibleValue, hiddenValue }

bool _isIepReadableField(String key) {
  return const <String>{
    'title',
    'domain',
    'longGoal',
    'shortGoal',
    'courseForm',
    'startEndDate',
    'participant',
    'implementer',
    'planDate',
    'startDate',
    'endDate',
    'project',
    'content',
    'teacherName',
    'courseName',
    'trainingDate',
    'preparation',
    'monthLabel',
  }.contains(key);
}

int _gridFlex(List<int> columns, int start, int span) {
  return columns
      .skip(start)
      .take(span)
      .fold<int>(0, (int sum, int width) => sum + width);
}

class _FixedGridCell {
  const _FixedGridCell({
    required this.columns,
    required this.child,
  });

  final int columns;
  final Widget child;
}

class _FixedGridRow extends StatelessWidget {
  const _FixedGridRow({
    required this.columns,
    required this.cells,
  });

  final List<int> columns;
  final List<_FixedGridCell> cells;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        if (!constraints.maxWidth.isFinite || constraints.maxWidth <= 0) {
          return Row(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: _expandedCells(),
          );
        }
        final int totalFlex =
            columns.fold<int>(0, (int sum, int width) => sum + width);
        final List<Widget> children = <Widget>[];
        int columnIndex = 0;
        double usedWidth = 0;
        for (int index = 0; index < cells.length; index += 1) {
          final _FixedGridCell cell = cells[index];
          final double width = index == cells.length - 1
              ? constraints.maxWidth - usedWidth
              : constraints.maxWidth *
                  _gridFlex(columns, columnIndex, cell.columns) /
                  totalFlex;
          usedWidth += width;
          columnIndex += cell.columns;
          children.add(
            SizedBox(
              width: width < 0 ? 0 : width,
              child: cell.child,
            ),
          );
        }
        return Row(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: children,
        );
      },
    );
  }

  List<Widget> _expandedCells() {
    final List<Widget> children = <Widget>[];
    int columnIndex = 0;
    for (final _FixedGridCell cell in cells) {
      children.add(
        Expanded(
          flex: _gridFlex(columns, columnIndex, cell.columns),
          child: cell.child,
        ),
      );
      columnIndex += cell.columns;
    }
    return children;
  }
}

class _WordTable extends StatelessWidget {
  const _WordTable({
    required this.periodText,
    required this.plan,
    required this.domains,
    required this.selectedGoal,
    required this.onGoalTap,
    required this.onClearSelectedGoal,
  });

  final String periodText;
  final IepPlan? plan;
  final List<_DocDomainData> domains;
  final _GoalEditRequest? selectedGoal;
  final ValueChanged<_GoalEditRequest> onGoalTap;
  final VoidCallback onClearSelectedGoal;

  static const List<int> _columns = <int>[
    1038,
    1472,
    625,
    877,
    1260,
    1562,
    927,
    2319,
  ];
  static const double _minHeight = 820;
  static const double _headerHeight = 208;

  @override
  Widget build(BuildContext context) {
    final IepPlan? currentPlan = plan;
    return GestureDetector(
      behavior: HitTestBehavior.translucent,
      onTap: onClearSelectedGoal,
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final double tableWidth =
              constraints.maxWidth.isFinite && constraints.maxWidth > 0
                  ? constraints.maxWidth
                  : 1008;
          final double bodyHeight = _bodyHeightFor(domains, tableWidth);
          return Column(
            children: <Widget>[
              _WordTableTitle(title: currentPlan?.title ?? '康复教学季度计划'),
              _DocTableRow(
                height: 42,
                cells: <_DocCellData>[
                  _DocCellData(text: '姓名', columns: 1, bold: true),
                  _DocCellData(
                      text: currentPlan?.student.name ?? '-', columns: 1),
                  _DocCellData(text: '性别', columns: 1, bold: true),
                  _DocCellData(
                      text: currentPlan?.student.gender ?? '-', columns: 1),
                  _DocCellData(text: '出生年月', columns: 1, bold: true),
                  _DocCellData(
                    text: currentPlan?.student.birthDate ?? '-',
                    columns: 3,
                    last: true,
                  ),
                ],
              ),
              _DocTableRow(
                height: 42,
                cells: <_DocCellData>[
                  _DocCellData(text: '制定日期', columns: 1, bold: true),
                  _DocCellData(
                      text: currentPlan?.meta.planDate ?? '-', columns: 3),
                  _DocCellData(text: '计划参与者', columns: 1, bold: true),
                  _DocCellData(
                    text: currentPlan?.meta.participant ?? '-',
                    columns: 3,
                    last: true,
                  ),
                ],
              ),
              _DocTableRow(
                height: 42,
                cells: <_DocCellData>[
                  _DocCellData(text: '实施者', columns: 1, bold: true),
                  _DocCellData(
                      text: currentPlan?.meta.implementer ?? '-', columns: 3),
                  _DocCellData(text: '实施\n起止日期', columns: 1, bold: true),
                  _DocCellData(
                    text:
                        _metaRangeText(currentPlan?.meta, fallback: periodText),
                    columns: 3,
                    noWrap: true,
                    last: true,
                  ),
                ],
              ),
              _DocTableRow(
                height: 42,
                cells: <_DocCellData>[
                  _DocCellData(text: '康复\n领域', columns: 1, bold: true),
                  _DocCellData(text: '长期目标', columns: 3, bold: true),
                  _DocCellData(text: '短期目标', columns: 2, bold: true),
                  _DocCellData(text: '课程\n形式', columns: 1, bold: true),
                  _DocCellData(
                      text: '起止日期', columns: 1, bold: true, last: true),
                ],
              ),
              SizedBox(
                height: bodyHeight,
                child: _DocPlanRows(
                  domains: domains,
                  height: bodyHeight,
                  tableWidth: tableWidth,
                  selectedGoal: selectedGoal,
                  onGoalTap: onGoalTap,
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  static double _bodyHeightFor(List<_DocDomainData> domains, double width) {
    final double minBodyHeight = _minHeight - _headerHeight;
    final double contentHeight = _DocPlanRows.heightFor(domains, width);
    return contentHeight > minBodyHeight ? contentHeight : minBodyHeight;
  }
}

class _WordTableFrame extends StatefulWidget {
  const _WordTableFrame({
    required this.child,
  });

  final Widget child;

  @override
  State<_WordTableFrame> createState() => _WordTableFrameState();
}

class _WordTableFrameState extends State<_WordTableFrame> {
  late final ScrollController _scrollController;

  @override
  void initState() {
    super.initState();
    _scrollController = ScrollController();
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      foregroundDecoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: const Color(0xFFB98A71), width: 1.2),
      ),
      child: Padding(
        padding: const EdgeInsets.all(1.2),
        child: LayoutBuilder(
          builder: (BuildContext context, BoxConstraints constraints) {
            final double maxHeight =
                constraints.maxHeight.isFinite && constraints.maxHeight > 0
                    ? constraints.maxHeight
                    : 612;
            return ConstrainedBox(
              constraints: BoxConstraints(maxHeight: maxHeight),
              child: SingleChildScrollView(
                controller: _scrollController,
                physics: const ClampingScrollPhysics(),
                child: widget.child,
              ),
            );
          },
        ),
      ),
    );
  }
}

class _WeekPlanTable extends StatelessWidget {
  const _WeekPlanTable({
    required this.month,
    required this.weekNumber,
    required this.weekDates,
    required this.plan,
  });

  final String month;
  final int weekNumber;
  final List<DateTime> weekDates;
  final IepWeeklyPlan? plan;

  static const List<int> _columns = <int>[
    1300,
    1334,
    1333,
    1925,
    698,
    698,
    698,
    698,
    698,
    698,
  ];

  static const List<_WeekTrainingRow> _fallbackRows = <_WeekTrainingRow>[
    _WeekTrainingRow(
      project: '平衡木行走',
      content:
          '在感统室设置宽10cm、高20cm的平衡木，儿童独立行走3米，治疗师在旁保护但不接触，完成后给予口头表扬和代币强化，每日练习2次，记录掉落次数，目标连续3次不掉落。',
    ),
    _WeekTrainingRow(
      project: '直线剪纸',
      content:
          '使用彩色纸画有直线（线宽0.5cm），儿童独立剪10cm，治疗师用尺子测量偏差，偏差在0.5cm内给予代币，累计代币兑换偏好活动。',
    ),
    _WeekTrainingRow(
      project: '情绪指认',
      content:
          '在绘本阅读中，治疗师指向角色表情，问“他感觉怎么样？”，儿童从4张情绪卡片（高兴、生气、伤心、害怕）中选择对应卡片，正确率100%后，儿童尝试口头命名情绪。',
    ),
    _WeekTrainingRow(
      project: '动作序列模仿',
      content: '变换动作序列（如“拍肩-转圈-跳”），治疗师示范后儿童模仿，顺序正确率80%以上，逐渐撤除视觉提示，仅靠观察模仿。',
    ),
    _WeekTrainingRow(
      project: '主动邀请同伴',
      content:
          '使用社交故事《邀请朋友玩》，课前阅读，课上治疗师提示儿童使用邀请语“我们一起玩吧”，当儿童主动邀请时，同伴积极回应，形成自然强化，记录邀请次数，目标每节至少2次。',
    ),
    _WeekTrainingRow(
      project: '变换问答',
      content:
          '使用视觉提示卡“说不同的话”，当儿童在对话中重复同一回答时，治疗师出示卡片并等待3秒，儿童变换回答后立即表扬，逐渐减少提示，记录重复次数。',
    ),
    _WeekTrainingRow(
      project: '减少摇晃行为',
      content:
          '使用区别强化，当儿童保持安静坐好2分钟无摇晃，给予代币，累计代币兑换偏好活动；摇晃行为发生时，不给予关注，仅重新引导，目标每节课不超过2次。',
    ),
    _WeekTrainingRow(
      project: '多属性分类',
      content:
          '使用属性卡片（红色、圆形、大），儿童根据指令将物品放入对应盒子，如“把红色的放一起”，逐渐增加复杂度，同时按两个属性分类（如红色且圆形），正确后给予代币，正确率90%以上。',
    ),
    _WeekTrainingRow(
      project: '两步指令执行',
      content:
          '使用图片提示卡（拍手、摸头），治疗师发两步指令“先拍手再摸头”时同时出示图片，儿童执行，逐渐撤除图片，仅靠听觉理解，正确率90%以上。',
    ),
    _WeekTrainingRow(
      project: '完整句子描述',
      content:
          '提供缺少主语的图片，治疗师问“谁在做什么？”，儿童需补充完整句子，如“妈妈在做饭”，逐渐增加句子成分（加入地点“在厨房”），正确结构后给予代币。',
    ),
  ];

  @override
  Widget build(BuildContext context) {
    final IepWeeklyPlan? currentPlan = plan;
    final List<_WeekTrainingRow> rows = currentPlan?.rows
            .map(_weekTrainingRowFromPlanRow)
            .where((_WeekTrainingRow row) =>
                row.project.trim().isNotEmpty || row.content.trim().isNotEmpty)
            .toList() ??
        _fallbackRows;
    final List<DateTime> displayWeekDates =
        _dateListFromStrings(currentPlan?.weekDates) ?? weekDates;
    return Column(
      children: <Widget>[
        _WordTableTitle(
          title: currentPlan?.title.trim().isNotEmpty == true
              ? currentPlan!.title
              : '康复教学周计划日记录卡$month第$weekNumber周',
        ),
        _WeekInfoRows(
          plan: currentPlan,
          periodText: currentPlan?.trainingDate.trim().isNotEmpty == true
              ? currentPlan!.trainingDate
              : _weekRangeText(displayWeekDates),
        ),
        _WeekHeaderRows(dates: displayWeekDates),
        _WeekTrainingRows(rows: rows),
      ],
    );
  }
}

class _WeekInfoRows extends StatelessWidget {
  const _WeekInfoRows({required this.periodText, required this.plan});

  final String periodText;
  final IepWeeklyPlan? plan;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        _WeekDocTableRow(
          height: 42,
          cells: <_WeekDocCellData>[
            _WeekDocCellData(text: '姓名', columns: 1, bold: true),
            _WeekDocCellData(text: plan?.student.name ?? '-', columns: 1),
            _WeekDocCellData(text: '性别', columns: 1, bold: true),
            _WeekDocCellData(text: plan?.student.gender ?? '-', columns: 1),
            _WeekDocCellData(text: '出生年月', columns: 2, bold: true),
            _WeekDocCellData(
              text: plan?.student.birthDate ?? '-',
              columns: 4,
              last: true,
            ),
          ],
        ),
        _WeekDocTableRow(
          height: 42,
          cells: <_WeekDocCellData>[
            const _WeekDocCellData(text: '任教\n老师', columns: 1, bold: true),
            _WeekDocCellData(
              text: plan?.teacherName.trim().isNotEmpty == true
                  ? plan!.teacherName
                  : '-',
              columns: 1,
            ),
            const _WeekDocCellData(text: '课程\n名称', columns: 1, bold: true),
            _WeekDocCellData(
              text: plan?.courseName.trim().isNotEmpty == true
                  ? plan!.courseName
                  : '康复教学',
              columns: 1,
            ),
            const _WeekDocCellData(text: '训练日期', columns: 2, bold: true),
            _WeekDocCellData(
              text: periodText,
              columns: 4,
              last: true,
            ),
          ],
        ),
        _WeekDocAutoTableRow(
          cells: <_WeekDocCellData>[
            _WeekDocCellData(text: '训练前\n准备', columns: 1, bold: true),
            _WeekDocCellData(
              text: plan?.preparation.trim().isNotEmpty == true
                  ? plan!.preparation
                  : '训练材料、视觉提示卡、强化物、记录表',
              columns: 9,
              align: TextAlign.left,
              last: true,
            ),
          ],
        ),
      ],
    );
  }
}

class _WeekHeaderRows extends StatelessWidget {
  const _WeekHeaderRows({required this.dates});

  final List<DateTime> dates;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 70,
      child: _FixedGridRow(
        columns: _WeekPlanTable._columns,
        cells: <_FixedGridCell>[
          const _FixedGridCell(
            columns: 1,
            child: _WeekDocCellBox(
              data: _WeekDocCellData(text: '训练项目', columns: 1, bold: true),
            ),
          ),
          const _FixedGridCell(
            columns: 3,
            child: _WeekDocCellBox(
              data: _WeekDocCellData(text: '训练内容', columns: 3, bold: true),
            ),
          ),
          _FixedGridCell(
            columns: 6,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: <Widget>[
                const SizedBox(
                  height: 36,
                  child: _WeekDocCellBox(
                    data: _WeekDocCellData(
                      text: '完成情况',
                      columns: 6,
                      bold: true,
                      last: true,
                    ),
                  ),
                ),
                SizedBox(
                  height: 34,
                  child: _FixedGridRow(
                    columns: _WeekPlanTable._columns.sublist(4),
                    cells: List<_FixedGridCell>.generate(6, (int index) {
                      final String label = index < dates.length
                          ? _weekDateLabel(dates[index])
                          : '';
                      return _FixedGridCell(
                        columns: 1,
                        child: _WeekDocCellBox(
                          data: _WeekDocCellData(
                            text: label,
                            columns: 1,
                            bold: true,
                            last: index == 5,
                          ),
                        ),
                      );
                    }),
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

class _WeekDocCellData {
  const _WeekDocCellData({
    required this.text,
    required this.columns,
    this.bold = false,
    this.align = TextAlign.center,
    this.last = false,
  });

  final String text;
  final int columns;
  final bool bold;
  final TextAlign align;
  final bool last;
}

class _WeekDocTableRow extends StatelessWidget {
  const _WeekDocTableRow({required this.height, required this.cells});

  final double height;
  final List<_WeekDocCellData> cells;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: height,
      child: _FixedGridRow(
        columns: _WeekPlanTable._columns,
        cells: cells.map((_WeekDocCellData cell) {
          return _FixedGridCell(
            columns: cell.columns,
            child: _WeekDocCellBox(data: cell),
          );
        }).toList(),
      ),
    );
  }
}

class _WeekDocAutoTableRow extends StatelessWidget {
  const _WeekDocAutoTableRow({required this.cells});

  final List<_WeekDocCellData> cells;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width =
            constraints.maxWidth.isFinite ? constraints.maxWidth : 900;
        return SizedBox(
          height: _rowHeight(width),
          child: _FixedGridRow(
            columns: _WeekPlanTable._columns,
            cells: cells.map((_WeekDocCellData cell) {
              return _FixedGridCell(
                columns: cell.columns,
                child: _WeekDocCellBox(
                  data: cell,
                  verticalPadding: 4,
                ),
              );
            }).toList(),
          ),
        );
      },
    );
  }

  double _rowHeight(double tableWidth) {
    double maxHeight = 42;
    int columnIndex = 0;
    for (final _WeekDocCellData cell in cells) {
      final double cellWidth = _DocRowMetrics.scaledColumnWidth(
        tableWidth,
        _WeekPlanTable._columns,
        columnIndex,
        cell.columns,
      );
      maxHeight = math.max(
        maxHeight,
        _DocRowMetrics.textHeight(
          cell.text,
          width: cellWidth,
          fontSize: 10.8,
          lineHeight: 1.18,
          horizontalPadding: 14,
          verticalPadding: 4,
        ),
      );
      columnIndex += cell.columns;
    }
    return maxHeight;
  }
}

class _DocRowMetrics {
  const _DocRowMetrics._();

  static double scaledColumnWidth(
    double tableWidth,
    List<int> columns,
    int start,
    int span,
  ) {
    final int totalFlex =
        columns.fold<int>(0, (int sum, int width) => sum + width);
    final int spanFlex = columns.skip(start).take(span).fold<int>(
          0,
          (int sum, int width) => sum + width,
        );
    return tableWidth * spanFlex / totalFlex;
  }

  static double textHeight(
    String text, {
    required double width,
    required double fontSize,
    required double lineHeight,
    double horizontalPadding = 16,
    double verticalPadding = 4,
    double safety = 5,
  }) {
    final double maxWidth = math.max(24, width - horizontalPadding);
    final TextPainter painter = TextPainter(
      text: TextSpan(
        text: text.trim().isEmpty ? ' ' : text,
        style: TextStyle(
          fontSize: fontSize,
          fontWeight: FontWeight.w700,
          height: lineHeight,
        ),
      ),
      textAlign: TextAlign.left,
      textDirection: TextDirection.ltr,
    )..layout(maxWidth: maxWidth);
    return painter.height + verticalPadding * 2 + safety;
  }
}

class _WeekDocCellBox extends StatelessWidget {
  const _WeekDocCellBox({
    required this.data,
    this.rowLast = false,
    this.verticalPadding = 5,
  });

  final _WeekDocCellData data;
  final bool rowLast;
  final double verticalPadding;

  @override
  Widget build(BuildContext context) {
    return Container(
      alignment: Alignment.center,
      padding: EdgeInsets.symmetric(horizontal: 7, vertical: verticalPadding),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(
          right: data.last
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
          bottom: rowLast
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
        ),
      ),
      child: Text(
        data.text,
        overflow: TextOverflow.clip,
        textAlign: data.align,
        style: TextStyle(
          color: data.bold ? _IepColors.ink : _IepColors.text,
          fontSize: 10.8,
          fontWeight: data.bold ? FontWeight.w900 : FontWeight.w700,
          height: 1.18,
        ),
      ),
    );
  }
}

class _WeekTrainingRows extends StatelessWidget {
  const _WeekTrainingRows({required this.rows});

  final List<_WeekTrainingRow> rows;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: rows.asMap().entries.map((entry) {
        return _WeekTrainingTableRow(
          row: entry.value,
          last: entry.key == rows.length - 1,
        );
      }).toList(),
    );
  }
}

class _WeekTrainingRow {
  const _WeekTrainingRow({
    required this.project,
    required this.content,
    this.completion = const <String>[],
  });

  final String project;
  final String content;
  final List<String> completion;

  double get rowHeight {
    if (content.length >= 85) {
      return 62;
    }
    if (content.length >= 70) {
      return 54;
    }
    return 46;
  }
}

class _WeekTrainingTableRow extends StatelessWidget {
  const _WeekTrainingTableRow({
    required this.row,
    required this.last,
  });

  final _WeekTrainingRow row;
  final bool last;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: row.rowHeight,
      child: _FixedGridRow(
        columns: _WeekPlanTable._columns,
        cells: <_FixedGridCell>[
          _FixedGridCell(
            columns: 1,
            child: _WeekDocCellBox(
              data: _WeekDocCellData(text: row.project, columns: 1, bold: true),
              rowLast: last,
              verticalPadding: 3,
            ),
          ),
          _FixedGridCell(
            columns: 3,
            child: _WeekDocCellBox(
              data: _WeekDocCellData(
                text: row.content,
                columns: 3,
                align: TextAlign.left,
              ),
              rowLast: last,
              verticalPadding: 3,
            ),
          ),
          ...List<_FixedGridCell>.generate(6, (int index) {
            final String completionText = index < row.completion.length
                ? row.completion[index].trim()
                : '';
            return _FixedGridCell(
              columns: 1,
              child: _WeekDocCellBox(
                data: _WeekDocCellData(
                  text: completionText,
                  columns: 1,
                  bold: completionText.isNotEmpty,
                  last: index == 5,
                ),
                rowLast: last,
                verticalPadding: 3,
              ),
            );
          }),
        ],
      ),
    );
  }
}

class _MonthPlanTable extends StatelessWidget {
  const _MonthPlanTable({
    required this.month,
    required this.monthRange,
    required this.plan,
  });

  final String month;
  final DateTimeRange monthRange;
  final IepMonthlyPlan? plan;

  static const List<int> _columns = <int>[
    806,
    907,
    907,
    958,
    957,
    882,
    882,
    882,
    882,
    826,
    595,
    595,
  ];

  static const List<_MonthDomainData> _fallbackDomains = <_MonthDomainData>[
    _MonthDomainData(
      domain: '大肌肉',
      longGoal: '1. 提升动态平衡与协调能力，能独立完成单脚站立、走平衡木等动作\n2. 增强下肢力量与跳跃技能，能连续向前跳跃并保持稳定',
      shortGoal: '能在宽10cm、高20cm的平衡木上独立行走3米，不掉落',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 使用宽10cm、高20cm的平衡木，治疗师先示范双手侧平举保持平衡，儿童在平地上沿直线行走练习，逐步过渡到平衡木上，初始可单手扶墙辅助，逐渐撤除辅助，完成3米行走不掉落。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 在感统室设置平衡木，儿童独立行走3米，治疗师在旁保护但不接触，完成后给予口头表扬和代币强化，每日练习2次，记录掉落次数，目标连续3次不掉落。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至户外低矮花坛边缘（约10cm宽），儿童独立行走3米，治疗师在旁监护，鼓励儿童在不同材质上保持平衡，完成后奖励贴纸。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '小肌肉',
      longGoal: '1. 提高手眼协调与精细操作能力，能熟练使用剪刀、穿珠子等\n2. 增强手部小肌肉控制，能完成复杂拼图与书写前准备',
      shortGoal: '能沿直线剪纸，偏差不超过0.5cm，连续剪10cm',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 提供儿童安全剪刀和画有粗直线的纸条（宽2cm），治疗师手把手辅助儿童开合剪刀，沿直线剪，逐步减少辅助，要求偏差不超过0.5cm，剪完10cm。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用彩色纸画有直线（线宽0.5cm），儿童独立剪10cm，治疗师用尺子测量偏差，偏差在0.5cm内给予代币，累计代币兑换偏好活动。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至剪不同材质（如卡纸、杂志页），儿童沿直线剪10cm，偏差不超过0.5cm，完成后将剪下的纸条用于粘贴画，增加趣味性。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '情感表达',
      longGoal: '1. 能识别并命名基本情绪，理解他人情绪并做出恰当反应\n2. 在情境中表达自己的情绪，并学习简单的情绪调节策略',
      shortGoal: '能指认高兴、生气、伤心、害怕四种情绪图片，正确率100%',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 使用情绪卡片（高兴、生气、伤心、害怕各2张），治疗师呈现卡片并命名，儿童指认，每次4选1，正确后给予社会性强化，错误时示范正确卡片，连续2次正确率100%进入下一阶段。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 在绘本阅读中，治疗师指向角色表情，问“他感觉怎么样？”，儿童从4张情绪卡片中选择对应卡片，正确率100%后，儿童尝试口头命名情绪。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至真实情境，当同伴或家人表现出情绪时，治疗师提示儿童观察并指认情绪图片，正确后给予自然强化（如“你看到妹妹哭了，知道她伤心，真棒！”）。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '模仿 （视觉/动作）',
      longGoal: '1. 提升动作模仿的准确性和复杂性，能模仿多步骤动作序列\n2. 增强视觉注意与模仿的泛化能力，能在不同情境下模仿他人',
      shortGoal: '能模仿3个连续的动作序列（如拍手-摸头-跺脚），顺序正确',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 治疗师示范“拍手-摸头-跺脚”序列，边说边做，儿童模仿，初始可分解教学，先模仿单个动作，再串联，使用视觉提示卡辅助顺序，正确后给予代币。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 变换动作序列（如“拍肩-转圈-跳”），治疗师示范后儿童模仿，顺序正确率80%以上，逐渐撤除视觉提示，仅靠观察模仿。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 在集体课中泛化，治疗师带领小组做动作序列，儿童跟随模仿，顺序正确后担任小老师带领其他儿童，增强动机。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '社交互动',
      longGoal: '1. 提高与同伴的互动能力，能主动发起并维持简单的社交游戏\n2. 理解并遵守基本社交规则，如轮流、分享、等待',
      shortGoal: '在小组活动中，能主动邀请同伴一起玩，至少2次/节',
      lesson: '集体课',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 在集体课自由游戏时间，治疗师设置合作性玩具（如积木、拼图），示范邀请语言“我们一起玩吧”，儿童模仿邀请同伴，每成功邀请1次给予贴纸，目标2次/节。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用社交故事《邀请朋友玩》，课前阅读，课上治疗师提示儿童使用邀请语，当儿童主动邀请时，同伴积极回应，形成自然强化，记录邀请次数。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至户外活动，儿童在滑梯或沙池主动邀请同伴，治疗师在旁观察，必要时给予手势提示，每节至少2次主动邀请，完成后奖励额外自由时间。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '行为特征 -语言',
      longGoal: '1. 减少刻板语言，增加功能性语言的灵活运用\n2. 提高语言在社交情境中的恰当性，能根据对象调整语言',
      shortGoal: '在对话中，能根据对方的问题变换回答，减少重复同一句话',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 个训中，治疗师与儿童进行简单问答（如“你喜欢什么颜色？”），若儿童重复同一回答，治疗师示范不同回答并提示“换一种说法”，正确变换后给予强化。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用视觉提示卡“说不同的话”，当儿童重复时，治疗师出示卡片并等待3秒，儿童变换回答后立即表扬，逐渐减少提示，记录重复次数。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至与同伴对话，设置情境（如分享玩具），同伴问“我可以玩吗？”，儿童需根据情境回答（如“可以”或“等一下”），而非固定回答，治疗师在旁辅助。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '行为特征 -非语言',
      longGoal: '1. 减少自我刺激行为，增加功能性非语言沟通\n2. 提高对环境变化的适应能力，减少刻板行为',
      shortGoal: '在课堂上，无意义的摇晃身体行为减少至每节课不超过2次',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 个训中，治疗师观察摇晃行为，当行为出现时，立即提供替代感觉输入（如挤压球、坐垫），并口头提醒“坐好”，记录频率，目标每节课不超过2次。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用区别强化，当儿童保持安静坐好2分钟无摇晃，给予代币，累计代币兑换偏好活动；摇晃行为发生时，不给予关注，仅重新引导。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至集体课，治疗师与主教合作，在集体活动中监控摇晃行为，使用视觉提示卡“安静身体”提醒，每节课摇晃不超过2次，达成后给予集体奖励。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '认知 （语言/语前）',
      longGoal: '1. 提升分类与排序能力，能按多种属性进行归类\n2. 增强问题解决与逻辑思维能力，能完成简单的推理任务',
      shortGoal: '能按颜色、形状、大小三个属性对物品进行分类，正确率90%',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 提供不同颜色、形状、大小的积木，治疗师先示范按颜色分类，儿童模仿，然后按形状分类，最后按大小分类，每次分类后提问“为什么放在一起？”，正确率90%以上。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用属性卡片（红色、圆形、大），儿童根据指令将物品放入对应盒子，如“把红色的放一起”，逐渐增加复杂度，同时按两个属性分类（如红色且圆形），正确后给予代币。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至生活场景，整理玩具时，儿童按颜色或类型将玩具放回架子，治疗师在旁提示，正确率90%后自然强化（环境整洁）。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '语言理解',
      longGoal: '1. 提高对复杂指令的理解，能执行两步以上指令\n2. 增强对故事和对话的理解，能回答相关问题',
      shortGoal: '能执行包含两个步骤的指令（如“先拍手再摸头”），正确率90%',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 个训中，治疗师发出两步指令“先拍手再摸头”，初始可示范，儿童模仿，然后仅用语言指令，儿童执行，正确后给予强化，错误时退回一步分解。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 使用图片提示卡（拍手、摸头），治疗师发指令时同时出示图片，儿童执行，逐渐撤除图片，仅靠听觉理解，正确率90%以上。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至集体课，在音乐活动中，治疗师唱出两步指令“先跺脚再拍手”，儿童跟随，正确后担任小指挥，增加趣味性。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
    _MonthDomainData(
      domain: '语言表达',
      longGoal: '1. 扩展句子长度与复杂性，能使用完整句子描述事件\n2. 提高叙事能力，能连贯讲述个人经历或故事',
      shortGoal: '能使用“主语+谓语+宾语”结构描述图片，如“男孩吃苹果”',
      lesson: '个训',
      trainings: <_MonthTrainingData>[
        _MonthTrainingData(
            '1. 使用动作图片卡（如男孩吃苹果、女孩拍球），治疗师示范“男孩吃苹果”，儿童模仿，然后出示新图片，儿童独立描述，正确结构后给予代币。',
            '2026-05-01 - 2026-05-10'),
        _MonthTrainingData(
            '2. 提供缺少主语的图片，治疗师问“谁在做什么？”，儿童需补充完整句子，如“妈妈在做饭”，逐渐增加句子成分（加入地点“在厨房”）。',
            '2026-05-11 - 2026-05-20'),
        _MonthTrainingData(
            '3. 泛化至绘本阅读，儿童描述书中画面，治疗师提示使用完整句子，如“小狗在睡觉”，正确后自然强化（继续讲故事）。',
            '2026-05-21 - 2026-05-31'),
      ],
    ),
  ];

  @override
  Widget build(BuildContext context) {
    final IepMonthlyPlan? currentPlan = plan;
    final List<_MonthDomainData> domains = currentPlan?.rows
            .map(_monthDomainFromPlanRow)
            .where((_MonthDomainData item) =>
                item.domain.trim().isNotEmpty ||
                item.shortGoal.trim().isNotEmpty ||
                item.trainings.any((_MonthTrainingData training) =>
                    training.content.trim().isNotEmpty))
            .toList() ??
        _fallbackDomains;
    return Column(
      children: <Widget>[
        _WordTableTitle(
          title: currentPlan?.title.trim().isNotEmpty == true
              ? currentPlan!.title
              : '康复教学$month计划',
        ),
        _MonthInfoRows(
          plan: currentPlan,
          periodText: _metaRangeText(
            currentPlan?.meta,
            fallback: _formatZhRange(monthRange.start, monthRange.end),
          ),
        ),
        _MonthPlanRows(domains: domains, monthRange: monthRange),
      ],
    );
  }
}

class _MonthInfoRows extends StatelessWidget {
  const _MonthInfoRows({required this.periodText, required this.plan});

  final String periodText;
  final IepMonthlyPlan? plan;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        _MonthDocTableRow(
          height: 42,
          cells: <_MonthDocCellData>[
            _MonthDocCellData(text: '姓名', columns: 1, bold: true),
            _MonthDocCellData(text: plan?.student.name ?? '-', columns: 2),
            _MonthDocCellData(text: '性别', columns: 1, bold: true),
            _MonthDocCellData(text: plan?.student.gender ?? '-', columns: 1),
            _MonthDocCellData(text: '出生年月', columns: 2, bold: true),
            _MonthDocCellData(
              text: plan?.student.birthDate ?? '-',
              columns: 5,
              last: true,
            ),
          ],
        ),
        _MonthDocTableRow(
          height: 42,
          cells: <_MonthDocCellData>[
            const _MonthDocCellData(text: '制定\n日期', columns: 1, bold: true),
            _MonthDocCellData(text: plan?.meta.planDate ?? '-', columns: 2),
            const _MonthDocCellData(text: '计划参与者', columns: 4, bold: true),
            _MonthDocCellData(
              text: plan?.meta.participant ?? '-',
              columns: 5,
              last: true,
            ),
          ],
        ),
        _MonthDocTableRow(
          height: 42,
          cells: <_MonthDocCellData>[
            const _MonthDocCellData(text: '实施者', columns: 1, bold: true),
            _MonthDocCellData(text: plan?.meta.implementer ?? '-', columns: 2),
            const _MonthDocCellData(text: '实施起止日期', columns: 4, bold: true),
            _MonthDocCellData(
              text: periodText,
              columns: 5,
              noWrap: true,
              last: true,
            ),
          ],
        ),
        _MonthDocTableRow(
          height: 42,
          cells: const <_MonthDocCellData>[
            _MonthDocCellData(text: '康复\n领域', columns: 1, bold: true),
            _MonthDocCellData(text: '长期目标', columns: 2, bold: true),
            _MonthDocCellData(text: '短期目标', columns: 2, bold: true),
            _MonthDocCellData(text: '训练内容', columns: 4, bold: true),
            _MonthDocCellData(text: '课程\n形式', columns: 1, bold: true),
            _MonthDocCellData(text: '起止日期', columns: 2, bold: true, last: true),
          ],
        ),
      ],
    );
  }
}

class _MonthDocCellData {
  const _MonthDocCellData({
    required this.text,
    required this.columns,
    this.bold = false,
    this.align = TextAlign.center,
    this.last = false,
    this.noWrap = false,
  });

  final String text;
  final int columns;
  final bool bold;
  final TextAlign align;
  final bool last;
  final bool noWrap;
}

class _MonthDocTableRow extends StatelessWidget {
  const _MonthDocTableRow({required this.height, required this.cells});

  final double height;
  final List<_MonthDocCellData> cells;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: height,
      child: _FixedGridRow(
        columns: _MonthPlanTable._columns,
        cells: cells.map((_MonthDocCellData cell) {
          return _FixedGridCell(
            columns: cell.columns,
            child: _MonthDocCellBox(data: cell),
          );
        }).toList(),
      ),
    );
  }
}

class _MonthDocCellBox extends StatelessWidget {
  const _MonthDocCellBox({
    required this.data,
    this.rowLast = false,
    this.verticalPadding = 5,
  });

  final _MonthDocCellData data;
  final bool rowLast;
  final double verticalPadding;

  @override
  Widget build(BuildContext context) {
    return Container(
      alignment: Alignment.center,
      padding: EdgeInsets.symmetric(horizontal: 7, vertical: verticalPadding),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(
          right: data.last
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
          bottom: rowLast
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
        ),
      ),
      child: Text(
        data.text,
        maxLines: data.noWrap ? 1 : null,
        overflow: data.noWrap ? TextOverflow.ellipsis : TextOverflow.clip,
        textAlign: data.align,
        style: TextStyle(
          color: data.bold ? _IepColors.ink : _IepColors.text,
          fontSize: 10.6,
          fontWeight: data.bold ? FontWeight.w900 : FontWeight.w700,
          height: 1.18,
        ),
      ),
    );
  }
}

class _MonthPlanRows extends StatelessWidget {
  const _MonthPlanRows({
    required this.domains,
    required this.monthRange,
  });

  final List<_MonthDomainData> domains;
  final DateTimeRange monthRange;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: domains.asMap().entries.map((entry) {
        return _MonthDomainBlock(
          data: entry.value,
          last: entry.key == domains.length - 1,
          monthRange: monthRange,
        );
      }).toList(),
    );
  }
}

class _MonthDomainData {
  const _MonthDomainData({
    required this.domain,
    required this.longGoal,
    required this.shortGoal,
    required this.lesson,
    required this.trainings,
  });

  final String domain;
  final String longGoal;
  final String shortGoal;
  final String lesson;
  final List<_MonthTrainingData> trainings;
}

class _MonthTrainingData {
  const _MonthTrainingData(this.content, this.period);

  final String content;
  final String period;

  String get displayPeriod => period.replaceFirst(' - ', '\n至 ');

  double get rowHeight {
    if (content.length >= 82) {
      return 70;
    }
    if (content.length >= 68) {
      return 60;
    }
    return 50;
  }
}

class _MonthDomainBlock extends StatelessWidget {
  const _MonthDomainBlock({
    required this.data,
    required this.last,
    required this.monthRange,
  });

  final _MonthDomainData data;
  final bool last;
  final DateTimeRange monthRange;

  @override
  Widget build(BuildContext context) {
    final List<double> rowHeights =
        data.trainings.map((training) => training.rowHeight).toList();
    final double blockHeight =
        rowHeights.fold<double>(0, (double sum, double height) => sum + height);

    return SizedBox(
      height: blockHeight,
      child: _FixedGridRow(
        columns: _MonthPlanTable._columns,
        cells: <_FixedGridCell>[
          _FixedGridCell(
            columns: 1,
            child: _MonthDocCellBox(
              data:
                  _MonthDocCellData(text: data.domain, columns: 1, bold: true),
              rowLast: last,
            ),
          ),
          _FixedGridCell(
            columns: 2,
            child: _MonthDocCellBox(
              data: _MonthDocCellData(
                text: data.longGoal,
                columns: 2,
                align: TextAlign.left,
              ),
              rowLast: last,
            ),
          ),
          _FixedGridCell(
            columns: 2,
            child: _MonthDocCellBox(
              data: _MonthDocCellData(
                text: data.shortGoal,
                columns: 2,
                align: TextAlign.left,
              ),
              rowLast: last,
            ),
          ),
          _FixedGridCell(
            columns: 4,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: data.trainings.asMap().entries.map((entry) {
                return SizedBox(
                  height: rowHeights[entry.key],
                  child: _MonthDocCellBox(
                    data: _MonthDocCellData(
                      text: entry.value.content,
                      columns: 4,
                      align: TextAlign.left,
                    ),
                    rowLast: last && entry.key == data.trainings.length - 1,
                    verticalPadding: 3,
                  ),
                );
              }).toList(),
            ),
          ),
          _FixedGridCell(
            columns: 1,
            child: _MonthDocCellBox(
              data: _MonthDocCellData(
                text: data.lesson,
                columns: 1,
                noWrap: true,
              ),
              rowLast: last,
            ),
          ),
          _FixedGridCell(
            columns: 2,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: data.trainings.asMap().entries.map((entry) {
                final String periodText = _monthTrainingPeriodText(
                  monthRange,
                  entry.key,
                );
                return SizedBox(
                  height: rowHeights[entry.key],
                  child: _MonthDocCellBox(
                    data: _MonthDocCellData(
                      text: periodText,
                      columns: 2,
                      last: true,
                    ),
                    rowLast: last && entry.key == data.trainings.length - 1,
                    verticalPadding: 3,
                  ),
                );
              }).toList(),
            ),
          ),
        ],
      ),
    );
  }
}

class _WordTableTitle extends StatelessWidget {
  const _WordTableTitle({this.title = '康复教学季度计划'});

  final String title;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 40,
      alignment: Alignment.center,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(
          bottom: BorderSide(color: Color(0xFFB98A71), width: 1),
        ),
      ),
      child: Text(
        title,
        style: TextStyle(
          color: _IepColors.ink,
          fontSize: 19,
          fontWeight: FontWeight.w900,
          letterSpacing: 1,
        ),
      ),
    );
  }
}

class _DocCellData {
  const _DocCellData({
    required this.text,
    required this.columns,
    this.bold = false,
    this.align = TextAlign.center,
    this.last = false,
    this.noWrap = false,
    this.editable = false,
  });

  final String text;
  final int columns;
  final bool bold;
  final TextAlign align;
  final bool last;
  final bool noWrap;
  final bool editable;
}

class _DocTableRow extends StatelessWidget {
  const _DocTableRow({required this.height, required this.cells});

  final double height;
  final List<_DocCellData> cells;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: height,
      child: _FixedGridRow(
        columns: _WordTable._columns,
        cells: cells.map((_DocCellData cell) {
          return _FixedGridCell(
            columns: cell.columns,
            child: _DocCellBox(data: cell),
          );
        }).toList(),
      ),
    );
  }
}

class _DocCellBox extends StatelessWidget {
  const _DocCellBox({
    required this.data,
    this.rowLast = false,
    this.verticalPadding = 5,
    this.onTap,
  });

  final _DocCellData data;
  final bool rowLast;
  final double verticalPadding;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final Widget content = Container(
      alignment: Alignment.center,
      padding: EdgeInsets.fromLTRB(
        8,
        verticalPadding,
        data.editable ? 14 : 8,
        verticalPadding,
      ),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: data.editable
            ? const <BoxShadow>[
                BoxShadow(
                  color: Color(0x22E96F43),
                  blurRadius: 0,
                  spreadRadius: 1.4,
                ),
              ]
            : null,
        border: Border(
          right: data.last
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
          bottom: rowLast
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
        ),
      ),
      child: Stack(
        clipBehavior: Clip.none,
        children: <Widget>[
          Align(
            alignment: Alignment.center,
            child: Text(
              data.text,
              maxLines: data.noWrap ? 1 : null,
              overflow: data.noWrap ? TextOverflow.ellipsis : TextOverflow.clip,
              textAlign: data.align,
              style: TextStyle(
                color: data.bold ? _IepColors.ink : _IepColors.text,
                fontSize: 11.4,
                fontWeight: data.bold ? FontWeight.w900 : FontWeight.w700,
                height: 1.22,
              ),
            ),
          ),
          if (data.editable)
            const Positioned(
              right: -6,
              top: -2,
              child: Icon(
                Icons.edit_rounded,
                size: 10,
                color: _IepColors.orangeDeep,
              ),
            ),
        ],
      ),
    );
    if (onTap == null) {
      return content;
    }
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        child: content,
      ),
    );
  }
}

class _DocPlanRows extends StatelessWidget {
  const _DocPlanRows({
    required this.domains,
    required this.height,
    required this.tableWidth,
    required this.selectedGoal,
    required this.onGoalTap,
  });

  final List<_DocDomainData> domains;
  final double height;
  final double tableWidth;
  final _GoalEditRequest? selectedGoal;
  final ValueChanged<_GoalEditRequest> onGoalTap;

  static const double _minDomainHeight = 122.4;
  static const double _minShortGoalRowHeight = 32;
  static const double _textFontSize = 11.4;
  static const double _textLineHeight = 1.28;
  static const double _textHorizontalPadding = 16;
  static const double _textHeightSafety = 5;

  static double _scaledColumnWidth(
    double tableWidth,
    List<int> columns,
    int start,
    int span,
  ) {
    final int totalFlex =
        columns.fold<int>(0, (int sum, int width) => sum + width);
    final int spanFlex = columns.skip(start).take(span).fold<int>(
          0,
          (int sum, int width) => sum + width,
        );
    return tableWidth * spanFlex / totalFlex;
  }

  static double _textHeight(
    String text, {
    required double width,
    double verticalPadding = 4,
  }) {
    final double maxWidth = math.max(24, width - _textHorizontalPadding);
    final TextPainter painter = TextPainter(
      text: TextSpan(
        text: text.trim().isEmpty ? ' ' : text,
        style: const TextStyle(
          fontSize: _textFontSize,
          fontWeight: FontWeight.w700,
          height: _textLineHeight,
        ),
      ),
      textAlign: TextAlign.left,
      textDirection: TextDirection.ltr,
    )..layout(maxWidth: maxWidth);
    return painter.height + verticalPadding * 2 + _textHeightSafety;
  }

  static double _shortGoalRowHeightFor(
    _DocShortGoalData goal,
    double tableWidth,
  ) {
    final double shortGoalWidth =
        _scaledColumnWidth(tableWidth, _WordTable._columns, 4, 2);
    final double periodWidth =
        _scaledColumnWidth(tableWidth, _WordTable._columns, 7, 1);
    return math.max(
      _minShortGoalRowHeight,
      math.max(
        _textHeight(goal.goal, width: shortGoalWidth, verticalPadding: 4),
        _textHeight(goal.period, width: periodWidth, verticalPadding: 4),
      ),
    );
  }

  static List<double> rowHeightsFor(_DocDomainData domain, double tableWidth) {
    final List<_DocShortGoalData> shortGoals = domain.shortGoals.isEmpty
        ? <_DocShortGoalData>[const _DocShortGoalData('', '个训', '')]
        : domain.shortGoals;
    final List<double> rowHeights = shortGoals
        .map(
          (_DocShortGoalData goal) => _shortGoalRowHeightFor(goal, tableWidth),
        )
        .toList();
    final double rowsHeight = rowHeights.fold<double>(
      0,
      (double sum, double rowHeight) => sum + rowHeight,
    );
    final double longGoalHeight = _textHeight(
      domain.longGoals.join('\n'),
      width: _scaledColumnWidth(tableWidth, _WordTable._columns, 1, 3),
      verticalPadding: 6,
    );
    final double targetHeight =
        math.max(_minDomainHeight, math.max(rowsHeight, longGoalHeight));
    if (targetHeight > rowsHeight && rowHeights.isNotEmpty) {
      final double extraPerRow =
          (targetHeight - rowsHeight) / rowHeights.length;
      for (int index = 0; index < rowHeights.length; index += 1) {
        rowHeights[index] += extraPerRow;
      }
    }
    return rowHeights;
  }

  static double blockHeightFor(_DocDomainData domain, double tableWidth) {
    return rowHeightsFor(domain, tableWidth).fold<double>(
      0,
      (double height, double rowHeight) => height + rowHeight,
    );
  }

  static double heightFor(List<_DocDomainData> domains, double tableWidth) {
    return domains.fold<double>(
      0,
      (double height, _DocDomainData domain) =>
          height + blockHeightFor(domain, tableWidth),
    );
  }

  @override
  Widget build(BuildContext context) {
    final double contentHeight = heightFor(domains, tableWidth);
    final double fillerHeight =
        height > contentHeight ? height - contentHeight : 0;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        ...domains.asMap().entries.map((entry) {
          final bool hasFiller = fillerHeight > 0;
          final List<double> rowHeights = rowHeightsFor(
            entry.value,
            tableWidth,
          );
          return SizedBox(
            height: rowHeights.fold<double>(
              0,
              (double height, double rowHeight) => height + rowHeight,
            ),
            child: _DocDomainBlock(
              domainIndex: entry.key,
              data: entry.value,
              rowHeights: rowHeights,
              selected: entry.key == 0,
              last: !hasFiller && entry.key == domains.length - 1,
              selectedGoal: selectedGoal,
              onGoalTap: onGoalTap,
            ),
          );
        }),
        if (fillerHeight > 0)
          SizedBox(
            height: fillerHeight,
            child: const _DocPlanFillerRow(),
          ),
      ],
    );
  }
}

class _DocPlanFillerRow extends StatelessWidget {
  const _DocPlanFillerRow();

  @override
  Widget build(BuildContext context) {
    return _FixedGridRow(
      columns: _WordTable._columns,
      cells: const <_FixedGridCell>[
        _FixedGridCell(
          columns: 1,
          child: _DocCellBox(
            data: _DocCellData(text: '', columns: 1),
            rowLast: true,
          ),
        ),
        _FixedGridCell(
          columns: 3,
          child: _DocCellBox(
            data: _DocCellData(text: '', columns: 3),
            rowLast: true,
          ),
        ),
        _FixedGridCell(
          columns: 2,
          child: _DocCellBox(
            data: _DocCellData(text: '', columns: 2),
            rowLast: true,
          ),
        ),
        _FixedGridCell(
          columns: 1,
          child: _DocCellBox(
            data: _DocCellData(text: '', columns: 1),
            rowLast: true,
          ),
        ),
        _FixedGridCell(
          columns: 1,
          child: _DocCellBox(
            data: _DocCellData(text: '', columns: 1, last: true),
            rowLast: true,
          ),
        ),
      ],
    );
  }
}

class _DocDomainData {
  const _DocDomainData({
    required this.domain,
    required this.longGoals,
    required this.shortGoals,
  });

  final String domain;
  final List<String> longGoals;
  final List<_DocShortGoalData> shortGoals;

  _DocDomainData copyWith({
    List<String>? longGoals,
    List<_DocShortGoalData>? shortGoals,
  }) {
    return _DocDomainData(
      domain: domain,
      longGoals: longGoals ?? this.longGoals,
      shortGoals: shortGoals ?? this.shortGoals,
    );
  }
}

class _DocShortGoalData {
  const _DocShortGoalData(this.goal, this.lesson, this.period);

  final String goal;
  final String lesson;
  final String period;

  _DocShortGoalData copyWith({
    String? goal,
    String? lesson,
    String? period,
  }) {
    return _DocShortGoalData(
      goal ?? this.goal,
      lesson ?? this.lesson,
      period ?? this.period,
    );
  }
}

class _DocDomainBlock extends StatelessWidget {
  const _DocDomainBlock({
    required this.domainIndex,
    required this.data,
    required this.rowHeights,
    required this.selected,
    required this.last,
    required this.selectedGoal,
    required this.onGoalTap,
  });

  final int domainIndex;
  final _DocDomainData data;
  final List<double> rowHeights;
  final bool selected;
  final bool last;
  final _GoalEditRequest? selectedGoal;
  final ValueChanged<_GoalEditRequest> onGoalTap;

  @override
  Widget build(BuildContext context) {
    final _GoalEditRequest longGoalRequest =
        _GoalEditRequest.longGoal(domainIndex: domainIndex);
    return _FixedGridRow(
      columns: _WordTable._columns,
      cells: <_FixedGridCell>[
        _FixedGridCell(
          columns: 1,
          child: _DocMergedCell(text: data.domain, bold: true, rowLast: last),
        ),
        _FixedGridCell(
          columns: 3,
          child: _DocMergedCell(
            text: data.longGoals.join('\n'),
            align: TextAlign.left,
            rowLast: last,
            editable: selectedGoal == longGoalRequest,
            onTap: () => onGoalTap(longGoalRequest),
          ),
        ),
        _FixedGridCell(
          columns: 2,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: data.shortGoals.asMap().entries.map((entry) {
              final _GoalEditRequest request = _GoalEditRequest.shortGoal(
                domainIndex: domainIndex,
                shortGoalIndex: entry.key,
              );
              return SizedBox(
                height: rowHeights[entry.key],
                child: Stack(
                  children: <Widget>[
                    Positioned.fill(
                      child: _DocCellBox(
                        data: _DocCellData(
                          text: entry.value.goal,
                          columns: 2,
                          align: TextAlign.left,
                          editable: selectedGoal == request,
                        ),
                        rowLast:
                            last && entry.key == data.shortGoals.length - 1,
                        verticalPadding: 4,
                        onTap: () => onGoalTap(request),
                      ),
                    ),
                  ],
                ),
              );
            }).toList(),
          ),
        ),
        _FixedGridCell(
          columns: 1,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: data.shortGoals.asMap().entries.map((entry) {
              return SizedBox(
                height: rowHeights[entry.key],
                child: _DocCellBox(
                  data: _DocCellData(
                    text: entry.value.lesson,
                    columns: 1,
                    noWrap: true,
                  ),
                  rowLast: last && entry.key == data.shortGoals.length - 1,
                  verticalPadding: 4,
                ),
              );
            }).toList(),
          ),
        ),
        _FixedGridCell(
          columns: 1,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: data.shortGoals.asMap().entries.map((entry) {
              return SizedBox(
                height: rowHeights[entry.key],
                child: _DocCellBox(
                  data: _DocCellData(
                    text: entry.value.period,
                    columns: 1,
                    noWrap: true,
                    last: true,
                  ),
                  rowLast: last && entry.key == data.shortGoals.length - 1,
                  verticalPadding: 4,
                ),
              );
            }).toList(),
          ),
        ),
      ],
    );
  }
}

class _DocMergedCell extends StatelessWidget {
  const _DocMergedCell({
    required this.text,
    this.bold = false,
    this.align = TextAlign.center,
    this.rowLast = false,
    this.editable = false,
    this.onTap,
  });

  final String text;
  final bool bold;
  final TextAlign align;
  final bool rowLast;
  final bool editable;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return _DocCellBox(
      data: _DocCellData(
        text: text,
        columns: 1,
        bold: bold,
        align: align,
        editable: editable,
      ),
      rowLast: rowLast,
      verticalPadding: 6,
      onTap: onTap,
    );
  }
}
