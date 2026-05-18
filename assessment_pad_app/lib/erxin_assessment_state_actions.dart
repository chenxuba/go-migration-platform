part of 'erxin_assessment_page.dart';

extension _ErxinAssessmentActions on _ErxinAssessmentPageState {
  ErxinAssessmentLaunchArgs _headerArgs() {
    return ErxinAssessmentLaunchArgs(
      studentId: _studentId,
      studentName: _studentName,
      studentAge: _resolvedStudentAgeText(),
      birthDate: _dateOnlyText(_birthDate),
      assessmentDate: _dateOnlyText(_assessmentDate),
      examinerName: _examinerName,
      scaleName: widget.args.scaleName,
    );
  }

  Widget _buildLoadingShell() {
    return ColoredBox(
      key: const ValueKey<String>('erxin-loading-shell'),
      color: _ErxinColors.page,
      child: Column(
        children: <Widget>[
          _Header(
            args: _headerArgs(),
            autoSaveText: '加载中...',
            saving: false,
            submitting: false,
            actionsEnabled: false,
            onBack: widget.onBack,
            onSave: () {},
            onSubmit: () {},
          ),
          Expanded(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(10, 10, 10, 10),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  _buildLoadingSidebar(),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Container(
                      clipBehavior: Clip.antiAlias,
                      decoration: _erxinPanelDecoration(),
                      child: Column(
                        children: <Widget>[
                          Expanded(child: _buildLoadingWorkspace()),
                          const _ErxinLoadingDetailPanel(),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(width: 10),
                  _buildLoadingRulePanel(),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLoadingSidebar() {
    return Container(
      key: const ValueKey<String>('erxin-loading-sidebar'),
      width: 214,
      padding: const EdgeInsets.fromLTRB(
        14,
        16,
        12,
        _erxinSidebarBottomPadding,
      ),
      decoration: _erxinPanelDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: const <Widget>[
          Text(
            '能区进度',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 17,
              fontWeight: FontWeight.w900,
            ),
          ),
          SizedBox(height: 12),
          _ErxinLoadingDomainRow(selected: true),
          _ErxinLoadingDomainRow(selected: false),
          _ErxinLoadingDomainRow(selected: false),
          _ErxinLoadingDomainRow(selected: false),
          _ErxinLoadingDomainRow(selected: false),
          SizedBox(height: 2),
          _AllItemsButton(),
          Spacer(),
          _ErxinLoadingProgressSummary(),
          SizedBox(height: _erxinProgressSummaryBottomGap),
        ],
      ),
    );
  }

  Widget _buildLoadingWorkspace() {
    final int mainAge = _mainAgeMonth;
    return Container(
      key: const ValueKey<String>('erxin-loading-workspace'),
      padding: const EdgeInsets.fromLTRB(16, 10, 12, 6),
      color: Colors.white,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              const _ErxinSkeletonBlock(width: 70, height: 20),
              const Text(
                ' · 当前测查',
                style: TextStyle(
                  color: _ErxinColors.ink,
                  fontSize: 20,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const Spacer(),
              if (mainAge > 0)
                _SmallBadge(text: '主测月龄 $mainAge月龄', strong: true)
              else
                const _ErxinSkeletonBlock(width: 118, height: 24, radius: 999),
            ],
          ),
          const SizedBox(height: 8),
          Expanded(
            child: ListView(
              padding: EdgeInsets.zero,
              physics: const NeverScrollableScrollPhysics(),
              children: const <Widget>[
                _ErxinLoadingMonthSection(rowCount: 3),
                _ErxinLoadingMonthSection(rowCount: 2),
                _ErxinLoadingMonthSection(rowCount: 2),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLoadingRulePanel() {
    return Container(
      key: const ValueKey<String>('erxin-loading-rule'),
      width: 296,
      padding: const EdgeInsets.fromLTRB(
        14,
        14,
        14,
        _erxinRulePanelBottomPadding,
      ),
      clipBehavior: Clip.antiAlias,
      decoration: _erxinPanelDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '规则判断',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 20,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: const <Widget>[
                _ErxinLoadingNextCard(),
                SizedBox(height: 10),
                Text(
                  '测评记录',
                  style: TextStyle(
                    color: _ErxinColors.ink,
                    fontSize: 16,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                SizedBox(height: 6),
                Expanded(child: _ErxinLoadingRecordList()),
                SizedBox(height: 12),
                Text(
                  '测查推进',
                  style: TextStyle(
                    color: _ErxinColors.ink,
                    fontSize: 16,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                SizedBox(height: 7),
                _ErxinSkeletonBlock(height: 12),
                SizedBox(height: 7),
                _ErxinSkeletonBlock(widthFactor: .72, height: 12),
                SizedBox(height: 8),
                _ErxinSkeletonBlock(height: 38, radius: 8),
              ],
            ),
          ),
          const SizedBox(height: 10),
          _RightRemarkSection(
            height: _erxinRightRemarkSectionHeight,
            itemNo: 0,
            remark: '',
            onChanged: _noopRemarkChange,
            onEditingComplete: _noopRemarkComplete,
          ),
        ],
      ),
    );
  }

  Widget _buildWorkspace() {
    final List<int> months = _centerMonths;
    final bool reviewing = _isReviewingRecord;
    final int? reviewMonth = _reviewMonth;
    final List<int> previousMonths =
        _previousMonthsForDomain(_selectedDomainCode);
    final List<int> futureMonths = _futureMonthsForDomain(_selectedDomainCode);
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 10, 12, 6),
      color: Colors.white,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Text(
                reviewing && reviewMonth != null
                    ? '${_domainName(_selectedDomainCode)} · ${reviewMonth}月龄记录'
                    : '${_domainName(_selectedDomainCode)} · 当前测查',
                style: const TextStyle(
                  color: _ErxinColors.ink,
                  fontSize: 20,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const Spacer(),
              if (reviewing) ...<Widget>[
                SizedBox(
                  height: 30,
                  child: OutlinedButton.icon(
                    onPressed: _returnToCurrentAssessment,
                    icon: const Icon(Icons.arrow_back_rounded, size: 14),
                    label: const Text('返回当前测查'),
                    style: OutlinedButton.styleFrom(
                      visualDensity: VisualDensity.compact,
                      padding: const EdgeInsets.symmetric(horizontal: 10),
                      textStyle: const TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 8),
              ],
              _SmallBadge(text: '主测月龄 $_mainAgeMonth月龄', strong: true),
            ],
          ),
          const SizedBox(height: 8),
          Expanded(
            child: months.isEmpty
                ? const _CurrentItemsEmptyState(text: '请按右侧规则继续推进测查')
                : SingleChildScrollView(
                    controller: _workspaceScrollController,
                    padding: EdgeInsets.zero,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: <Widget>[
                        for (final MapEntry<int, int> entry
                            in months.asMap().entries) ...<Widget>[
                          if (!reviewing &&
                              entry.key == 1 &&
                              previousMonths.isNotEmpty)
                            const _WorkspaceMonthDivider(label: '往前测查'),
                          if (!reviewing &&
                              futureMonths.isNotEmpty &&
                              entry.value == futureMonths.first)
                            const _WorkspaceMonthDivider(label: '往后测查'),
                          _AgeMonthSection(
                            key: _monthSectionKeyFor(
                              _selectedDomainCode,
                              entry.value,
                            ),
                            month: entry.value,
                            isMainAge: entry.value == _mainAgeMonth,
                            flashing:
                                _workspaceFlashMonths.contains(entry.value),
                            items: _itemsFor(_selectedDomainCode, entry.value),
                            itemPasses: _itemPasses,
                            selectedItemNo: _selectedItemNo,
                            itemKeyFor: _itemRowKeyFor,
                            onSelectItem: _selectItem,
                            onScore: _scoreItem,
                          ),
                        ],
                      ],
                    ),
                  ),
          ),
        ],
      ),
    );
  }

  List<_RuleRow> _recordRowsForDomain(String domainCode) {
    final int mainAge = _mainAgeMonth;
    final int? reviewMonth = _reviewMonthByDomain[domainCode];
    final List<int> baselineMonths = _previousBaselineMonthsForDomain(
      domainCode,
    );
    final List<int> ceilingMonths = _futureCeilingMonthsForDomain(domainCode);
    final bool previousBoundaryStop = _previousBoundaryStopForDomain(
      domainCode,
    );
    final bool previousResolved =
        _hasPreviousBaselineForDomain(domainCode) || previousBoundaryStop;
    final bool futureBoundaryStop = _futureBoundaryStopForDomain(domainCode);
    final bool futureResolved =
        _hasFutureCeilingForDomain(domainCode) || futureBoundaryStop;
    final List<_RuleRow> rows = <_RuleRow>[
      for (final int month in _recordMonthsForDomain(domainCode))
        _RuleRow(
          label: month == mainAge
              ? '主测月龄$month月龄'
              : month < mainAge
                  ? '往前$month月龄'
                  : '往后$month月龄',
          value: _recordStatusText(domainCode, month),
          done: _recordDone(domainCode, month),
          month: month,
          selected: reviewMonth == month,
        ),
      _RuleRow(
        label: '前测基线',
        value: _hasPreviousBaselineForDomain(domainCode)
            ? '已建立'
            : previousBoundaryStop
                ? '已到最低月龄'
                : '未形成',
        done: previousResolved,
        targetMonths: baselineMonths,
      ),
    ];
    if (_futureVisibleDomains.contains(domainCode) ||
        _hasAnsweredFutureMonth(domainCode) ||
        futureBoundaryStop) {
      rows.add(
        _RuleRow(
          label: '后测封顶',
          value: _hasFutureCeilingForDomain(domainCode)
              ? '已建立'
              : futureBoundaryStop
                  ? '已到最高月龄'
                  : '未形成',
          done: futureResolved,
          targetMonths: ceilingMonths,
        ),
      );
    }
    return rows;
  }

  Widget _buildRulePanel() {
    final String nextText = _nextActionText();
    final bool futureShown =
        _futureVisibleDomains.contains(_selectedDomainCode);
    final bool canContinuePrevious = _canContinuePreviousMonths;
    final bool canEnterFuture = _canEnterFutureMonths;
    final bool canContinueFuture = _canContinueFutureMonths;
    final bool domainComplete = _domainStopRuleComplete(_selectedDomainCode);
    final bool canLocateCurrentItem = !canContinuePrevious &&
        !canEnterFuture &&
        !canContinueFuture &&
        !domainComplete &&
        _firstPendingAssessmentItemNoForDomain(_selectedDomainCode) > 0;
    final String actionLabel = canContinuePrevious
        ? '继续往前测查'
        : canContinueFuture
            ? '继续往后测查'
            : canLocateCurrentItem
                ? futureShown
                    ? '往后测查已展开'
                    : '往前测查已展开'
                : domainComplete
                    ? '本能区测查完成'
                    : futureShown
                        ? _hasFutureCeiling
                            ? '本能区测查完成'
                            : '往后测查已展开'
                        : '进入往后测查';
    final VoidCallback? actionHandler = canContinuePrevious
        ? _continuePreviousMonths
        : canEnterFuture
            ? _enterFutureMonths
            : canContinueFuture
                ? _continueFutureMonths
                : canLocateCurrentItem
                    ? _locateCurrentAssessmentItem
                    : null;
    return Container(
      width: 296,
      padding: const EdgeInsets.fromLTRB(
        14,
        14,
        14,
        _erxinRulePanelBottomPadding,
      ),
      clipBehavior: Clip.antiAlias,
      decoration: _erxinPanelDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '规则判断',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 20,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                _RuleCard(
                  title: '下一步',
                  body: nextText,
                  icon: canContinuePrevious
                      ? Icons.keyboard_double_arrow_left_rounded
                      : canContinueFuture
                          ? Icons.keyboard_double_arrow_right_rounded
                          : canLocateCurrentItem
                              ? Icons.center_focus_strong_rounded
                              : domainComplete
                                  ? Icons.check_circle_rounded
                                  : Icons.arrow_forward_rounded,
                  color: _ErxinColors.blue,
                ),
                const SizedBox(height: 10),
                Row(
                  children: const <Widget>[
                    Text(
                      '测评记录',
                      style: TextStyle(
                        color: _ErxinColors.ink,
                        fontSize: 16,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                Expanded(
                  child: _RuleChecklist(
                    rows: _recordRowsForDomain(_selectedDomainCode),
                    revealMonths: _recordRevealMonths,
                    revealSerial: _recordRevealSerial,
                    onTapMonth: _openAssessmentRecord,
                    onRevealTargets: _revealWorkspaceMonths,
                  ),
                ),
                const SizedBox(height: 12),
                const Text(
                  '测查推进',
                  style: TextStyle(
                    color: _ErxinColors.ink,
                    fontSize: 16,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 5),
                Text(
                  _hasFutureCeiling
                      ? '往后测查已形成连续两个标准月龄全不通过，本能区达到停止规则。'
                      : _futureBoundaryStopForDomain(_selectedDomainCode)
                          ? '已测至最高标准月龄，仍未形成后测封顶，按边界规则强行停止。'
                          : _hasPreviousBaseline
                              ? '前测已形成连续两个标准月龄全通过，继续往后寻找连续两个标准月龄全不通过。'
                              : _previousBoundaryStopForDomain(
                                      _selectedDomainCode)
                                  ? '已测至最低标准月龄，仍未形成前测基线，按边界规则进入往后测查。'
                                  : '前测尚未形成连续两个标准月龄全通过，需继续向更低月龄追测。',
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _ErxinColors.muted,
                    fontSize: 12,
                    height: 1.28,
                  ),
                ),
                const SizedBox(height: 6),
                SizedBox(
                  width: double.infinity,
                  height: 38,
                  child: FilledButton(
                    onPressed: actionHandler,
                    style: FilledButton.styleFrom(
                      backgroundColor: _ErxinColors.blue,
                      disabledBackgroundColor: const Color(0xFFE2D6CE),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    child: Text(
                      actionLabel,
                      style: const TextStyle(fontWeight: FontWeight.w900),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 10),
          _RightRemarkSection(
            height: _erxinRightRemarkSectionHeight,
            itemNo: _selectedItemNo,
            remark: _itemRemarks[_selectedItemNo] ?? '',
            onChanged: _updateItemRemark,
            onEditingComplete: _finishItemRemarkEdit,
          ),
        ],
      ),
    );
  }

  Widget _buildDetailPanel() {
    final ErxinItemSummary? summary = _summaryByNo(_selectedItemNo);
    final String fallbackTitle =
        summary == null ? '当前题目说明' : '${summary.itemNo} ${summary.itemTitle}';
    return Container(
      height: _erxinDetailPanelHeight,
      padding: const EdgeInsets.fromLTRB(
        16,
        _erxinDetailPanelTopPadding,
        12,
        _erxinDetailPanelBottomPadding,
      ),
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(top: BorderSide(color: _ErxinColors.line)),
      ),
      child: FutureBuilder<ErxinAssessmentItem>(
        future: _selectedItemNo > 0 ? _detailFuture(_selectedItemNo) : null,
        builder:
            (BuildContext context, AsyncSnapshot<ErxinAssessmentItem> snap) {
          final ErxinAssessmentItem? item = snap.data;
          final String title = item == null || item.itemNo <= 0
              ? fallbackTitle
              : '${item.itemNo} ${item.itemTitle}'
                  '${item.parentReportAllowed ? '（R）' : ''}';
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Expanded(
                    child: Text(
                      '当前题目说明：$title',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _ErxinColors.ink,
                        fontSize: 16,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  SizedBox(
                    height: 28,
                    child: OutlinedButton.icon(
                      onPressed: () => _showFullItemDetail(item, summary),
                      icon: const Icon(Icons.open_in_full_rounded, size: 14),
                      label: const Text('完整说明'),
                      style: OutlinedButton.styleFrom(
                        visualDensity: VisualDensity.compact,
                        padding: const EdgeInsets.symmetric(horizontal: 9),
                        textStyle: const TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: _erxinDetailHeaderGap),
              Expanded(
                child: Row(
                  children: <Widget>[
                    Expanded(
                      child: _DetailTextBox(
                        title: '操作方法',
                        text: item?.method ?? '正在加载题目操作方法...',
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: _DetailTextBox(
                        title: '通过标准',
                        text: item?.passCriteria ?? '正在加载通过标准...',
                      ),
                    ),
                  ],
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Future<ErxinAssessmentItem> _detailFuture(int itemNo) {
    final ErxinAssessmentItem? cached = _itemDetailCache[itemNo];
    if (cached != null) {
      return Future<ErxinAssessmentItem>.value(cached);
    }
    return _itemDetailFetches.putIfAbsent(itemNo, () async {
      final ErxinAssessmentItem item = await widget.client.fetchTemplateItem(
        _token,
        itemNo: itemNo,
      );
      _itemDetailCache[itemNo] = item;
      return item;
    });
  }

  void _prefetchSelectedItem() {
    if (_selectedItemNo > 0 && _token.trim().isNotEmpty) {
      _detailFuture(_selectedItemNo);
    }
  }

  GlobalKey _itemRowKeyFor(int itemNo) {
    return _itemRowKeys.putIfAbsent(
      itemNo,
      () => GlobalKey(debugLabel: 'erxin-item-row-$itemNo'),
    );
  }

  GlobalKey _monthSectionKeyFor(String domainCode, int month) {
    final String key = '$domainCode-$month';
    return _monthSectionKeys.putIfAbsent(
      key,
      () => GlobalKey(debugLabel: 'erxin-month-section-$key'),
    );
  }

  void _revealWorkspaceMonths(List<int> months) {
    final Set<int> rawTargets = months.toSet();
    if (rawTargets.isEmpty) {
      return;
    }
    final bool reviewing =
        _reviewMonthByDomain.containsKey(_selectedDomainCode);
    if (reviewing) {
      setState(() {
        _reviewMonthByDomain.remove(_selectedDomainCode);
      });
    }
    final List<int> targets = _centerMonths
        .where((int month) => rawTargets.contains(month))
        .toList(growable: false);
    if (targets.isEmpty) {
      return;
    }
    setState(() {
      _workspaceFlashMonths = targets;
      _workspaceFlashSerial++;
    });
    _scheduleWorkspaceFlashClear(_workspaceFlashSerial);
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) {
        return;
      }
      _scrollWorkspaceMonthsIntoView(targets);
    });
  }

  void _scheduleWorkspaceFlashClear(int serial) {
    _workspaceFlashTimer?.cancel();
    final List<int> targetMonths = _workspaceFlashMonths;
    int ticks = 0;
    _workspaceFlashTimer = Timer.periodic(const Duration(milliseconds: 220), (
      Timer timer,
    ) {
      if (!mounted || serial != _workspaceFlashSerial) {
        timer.cancel();
        return;
      }
      ticks++;
      setState(() {
        _workspaceFlashMonths = ticks.isOdd ? const <int>[] : targetMonths;
      });
      if (ticks >= 6) {
        timer.cancel();
        if (mounted && serial == _workspaceFlashSerial) {
          setState(() => _workspaceFlashMonths = const <int>[]);
        }
      }
    });
  }

  void _scrollWorkspaceMonthsIntoView(List<int> months) {
    if (!_workspaceScrollController.hasClients || months.isEmpty) {
      return;
    }
    final int firstMonth = months.first;
    final BuildContext? monthContext =
        _monthSectionKeys['$_selectedDomainCode-$firstMonth']?.currentContext;
    if (monthContext == null) {
      return;
    }
    unawaited(
      Scrollable.ensureVisible(
        monthContext,
        alignment: 0,
        duration: const Duration(milliseconds: 240),
        curve: Curves.easeOutCubic,
      ),
    );
  }

  void _revealSelectedItem() {
    final int itemNo = _selectedItemNo;
    if (itemNo <= 0) {
      return;
    }
    void reveal() {
      if (!mounted || _selectedItemNo != itemNo) {
        return;
      }
      final BuildContext? rowContext = _itemRowKeys[itemNo]?.currentContext;
      if (rowContext == null) {
        return;
      }
      unawaited(
        Scrollable.ensureVisible(
          rowContext,
          alignment: 0.1,
          duration: const Duration(milliseconds: 260),
          curve: Curves.easeOutCubic,
        ),
      );
    }

    WidgetsBinding.instance.addPostFrameCallback((_) {
      reveal();
      unawaited(
        Future<void>.delayed(
          const Duration(milliseconds: 120),
          reveal,
        ),
      );
    });
  }

  void _selectDomain(String domainCode) {
    setState(() {
      _selectedDomainCode = domainCode;
      _reviewMonthByDomain.remove(domainCode);
      _selectedItemNo = _firstCurrentItemNo(domainCode);
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstVisibleItemNo(domainCode);
      }
    });
    _prefetchSelectedItem();
  }

  void _selectItem(int itemNo) {
    setState(() => _selectedItemNo = itemNo);
    _prefetchSelectedItem();
  }

  void _showAllItemsOverview() {
    showGeneralDialog<void>(
      context: context,
      barrierDismissible: true,
      barrierLabel: '关闭全部题目总览',
      barrierColor: Colors.black.withOpacity(.26),
      transitionDuration: Duration.zero,
      pageBuilder: (
        BuildContext dialogContext,
        Animation<double> _,
        Animation<double> __,
      ) {
        return PadDialogViewport(
          child: _ErxinAllItemsOverviewDialog(
            domains: _template.domains,
            ageGroups: _template.ageGroups,
            itemPasses: _itemPasses,
            selectedItemNo: _selectedItemNo,
            mainAgeMonth: _mainAgeMonth,
            onClose: () => Navigator.of(dialogContext).pop(),
          ),
        );
      },
    );
  }

  void _updateItemRemark(int itemNo, String remark) {
    if (itemNo <= 0) {
      return;
    }
    final String normalized = remark.trim();
    setState(() {
      if (normalized.isEmpty) {
        _itemRemarks.remove(itemNo);
      } else {
        _itemRemarks[itemNo] = normalized;
      }
    });
  }

  void _finishItemRemarkEdit(int itemNo) {
    if (itemNo <= 0 ||
        !_itemPasses.containsKey(itemNo) ||
        _token.trim().isEmpty) {
      return;
    }
    unawaited(_saveItem(itemNo));
  }

  void _noopRemarkChange(int itemNo, String remark) {}

  void _noopRemarkComplete(int itemNo) {}

  void _scoreItem(int itemNo, bool passed) {
    unawaited(_scoreItemInternal(itemNo, passed));
  }

  Future<void> _scoreItemInternal(int itemNo, bool passed) async {
    final ErxinItemSummary? summary = _summaryByNo(itemNo);
    final String domainCode = summary?.domainCode ?? _selectedDomainCode;
    final bool historyReview = _reviewMonthByDomain[domainCode] != null &&
        summary?.ageMonth == _reviewMonthByDomain[domainCode];
    final bool changedExisting =
        _itemPasses.containsKey(itemNo) && _itemPasses[itemNo] != passed;
    if (historyReview && changedExisting) {
      final bool confirmed = await _confirmHistoryChange(
        itemNo: itemNo,
        passed: passed,
      );
      if (!confirmed || !mounted) {
        return;
      }
    }
    if (!mounted) {
      return;
    }
    int revealedItemNo = itemNo;
    setState(() {
      _itemPasses[itemNo] = passed;
      _reconcileAfterScore(domainCode);
      _selectedDomainCode = domainCode;
      final int nextSelected = _nextSelectedItemNoForDomain(domainCode, itemNo);
      _selectedItemNo = nextSelected > 0 ? nextSelected : itemNo;
      revealedItemNo = _selectedItemNo;
      _autoSaveText = _token.trim().isEmpty ? '本地已记录' : '保存中...';
    });
    _prefetchSelectedItem();
    if (revealedItemNo != itemNo) {
      _revealSelectedItem();
    }
    if (_token.trim().isNotEmpty) {
      unawaited(_saveItem(itemNo));
    }
  }

  void _openAssessmentRecord(int month) {
    setState(() {
      _reviewMonthByDomain[_selectedDomainCode] = month;
      _selectedItemNo = _firstItemNoForMonth(_selectedDomainCode, month);
      _autoSaveText = '正在查看历史记录';
    });
    _prefetchSelectedItem();
  }

  void _returnToCurrentAssessment() {
    setState(() {
      _reviewMonthByDomain.remove(_selectedDomainCode);
      _selectedItemNo = _firstCurrentItemNo(_selectedDomainCode);
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstVisibleItemNo(_selectedDomainCode);
      }
      _autoSaveText = '返回当前测查';
    });
    _prefetchSelectedItem();
  }

  void _locateCurrentAssessmentItem() {
    setState(() {
      _reviewMonthByDomain.remove(_selectedDomainCode);
      final int pendingItemNo =
          _firstPendingAssessmentItemNoForDomain(_selectedDomainCode);
      _selectedItemNo = pendingItemNo > 0
          ? pendingItemNo
          : _firstCurrentItemNo(_selectedDomainCode);
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstVisibleItemNo(_selectedDomainCode);
      }
      _autoSaveText = '已定位当前题';
    });
    _prefetchSelectedItem();
    _revealSelectedItem();
  }

  Future<bool> _confirmHistoryChange({
    required int itemNo,
    required bool passed,
  }) async {
    final ErxinItemSummary? summary = _summaryByNo(itemNo);
    final bool? confirmed = await showDialog<bool>(
      context: context,
      barrierDismissible: false,
      barrierColor: Colors.black.withOpacity(.24),
      builder: (BuildContext dialogContext) {
        return PadDialogViewport(
          child: Center(
            child: _HistoryChangeConfirmDialog(
              itemTitle: summary == null
                  ? '第$itemNo题'
                  : '${summary.itemNo} ${summary.itemTitle}',
              nextStatus: passed ? '通过' : '不通过',
              onCancel: () => Navigator.of(dialogContext).pop(false),
              onConfirm: () => Navigator.of(dialogContext).pop(true),
            ),
          ),
        );
      },
    );
    return confirmed == true;
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
      key: 'erxin-top-message',
    );
  }

  Future<ErxinDraftDetail?> _saveDraft({bool silent = false}) {
    final Future<ErxinDraftDetail?>? inFlight = _saveDraftFuture;
    if (inFlight != null) {
      if (!silent && _saveDraftFutureSilent && !_saveDraftJoinedByManual) {
        _saveDraftJoinedByManual = true;
        return _joinSilentDraftSave(inFlight);
      }
      return inFlight;
    }
    late final Future<ErxinDraftDetail?> trackedFuture;
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

  Future<ErxinDraftDetail?> _joinSilentDraftSave(
    Future<ErxinDraftDetail?> inFlight,
  ) async {
    if (mounted) {
      setState(() => _autoSaveText = '草稿保存中...');
    }
    final ErxinDraftDetail? detail = await inFlight;
    if (!mounted) {
      return detail;
    }
    if (detail == null) {
      _showMessage('保存草稿失败，请稍后重试');
    } else {
      _showMessage('草稿已保存', tone: PadMessageTone.success);
    }
    return detail;
  }

  Future<ErxinDraftDetail?> _performSaveDraft({required bool silent}) async {
    if (_token.trim().isEmpty) {
      _showMessage('请先登录后再保存草稿');
      return null;
    }
    if (_studentId <= 0 || _studentName.trim().isEmpty) {
      _showMessage('缺少儿童信息，无法保存草稿');
      return null;
    }
    setState(() => _savingDraft = true);
    try {
      final ErxinDraftDetail detail = await _sendSaveDraftWithRetry();
      if (!mounted) {
        return detail;
      }
      setState(() {
        _draftId = detail.id;
        _draftProgress = detail.progress;
        _autoSaveText = '已保存 ${_formatClock(DateTime.now())}';
      });
      if (!silent) {
        _showMessage('草稿已保存', tone: PadMessageTone.success);
      }
      return detail;
    } on Object catch (error) {
      if (!silent) {
        _showMessage('保存草稿失败：$error');
      } else if (mounted) {
        setState(() => _autoSaveText = '保存失败');
      }
      return null;
    } finally {
      if (mounted) {
        setState(() => _savingDraft = false);
      }
    }
  }

  Future<ErxinDraftDetail> _sendSaveDraftWithRetry() async {
    final int attemptedDraftId = _draftId;
    try {
      return await widget.client.saveDraft(
        _token,
        _buildDraftPayload(),
      );
    } on Object catch (error) {
      if (attemptedDraftId <= 0 || !_isDraftNotFoundError(error)) {
        rethrow;
      }
      if (mounted) {
        setState(() {
          _draftId = 0;
          _autoSaveText = '正在创建新草稿...';
        });
      } else {
        _draftId = 0;
      }
      return widget.client.saveDraft(
        _token,
        _buildDraftPayload(),
      );
    }
  }

  Future<void> _saveItem(int itemNo) async {
    if (itemNo <= 0 || !_itemPasses.containsKey(itemNo)) {
      return;
    }
    int draftId = _draftId;
    if (draftId <= 0) {
      final ErxinDraftDetail? created = await _saveDraft(silent: true);
      draftId = created?.id ?? _draftId;
    }
    if (draftId <= 0 || !mounted) {
      return;
    }
    setState(() {
      _autoSaveText = '保存中...';
    });
    try {
      final ErxinDraftDetail detail = await _sendSaveDraftItem(
        itemNo: itemNo,
        draftId: draftId,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _draftId = detail.id;
        _draftProgress = detail.progress;
        _autoSaveText = '已保存 ${_formatClock(DateTime.now())}';
      });
    } on Object catch (error) {
      if (_isDraftNotFoundError(error)) {
        final bool retried = await _retrySaveItemAfterMissingDraft(itemNo);
        if (retried) {
          return;
        }
      }
      if (mounted) {
        setState(() {
          _autoSaveText = '保存失败';
        });
      }
      _showMessage('第$itemNo题自动保存失败：$error');
    }
  }

  Future<ErxinDraftDetail> _sendSaveDraftItem({
    required int itemNo,
    required int draftId,
  }) {
    return widget.client.saveDraftItem(
      _token,
      <String, dynamic>{
        'draftId': draftId,
        'itemNo': itemNo,
        'passed': _itemPasses[itemNo],
        'remark': _itemRemarks[itemNo]?.trim() ?? '',
      },
    );
  }

  Future<bool> _retrySaveItemAfterMissingDraft(int itemNo) async {
    if (!mounted || itemNo <= 0 || !_itemPasses.containsKey(itemNo)) {
      return false;
    }
    setState(() {
      _draftId = 0;
      _autoSaveText = '正在重建草稿...';
    });
    try {
      final ErxinDraftDetail? created = await _saveDraft(silent: true);
      final int retryDraftId = created?.id ?? _draftId;
      if (retryDraftId <= 0 || !mounted) {
        return false;
      }
      final ErxinDraftDetail detail = await _sendSaveDraftItem(
        itemNo: itemNo,
        draftId: retryDraftId,
      );
      if (!mounted) {
        return true;
      }
      setState(() {
        _draftId = detail.id;
        _draftProgress = detail.progress;
        _autoSaveText = '已保存 ${_formatClock(DateTime.now())}';
      });
      return true;
    } on Object {
      return false;
    }
  }

  bool _isDraftNotFoundError(Object error) {
    return error
        .toString()
        .toLowerCase()
        .contains('assessment draft not found');
  }

  Future<void> _submitDraft() async {
    if (_submitting) {
      return;
    }
    final String? blocker = _localSubmitBlocker();
    if (blocker != null) {
      _showMessage(blocker);
      return;
    }
    setState(() => _submitting = true);
    try {
      final ErxinDraftDetail? detail = await _saveDraft(silent: true);
      final int draftId = detail?.id ?? _draftId;
      if (draftId <= 0) {
        _showMessage('请先保存草稿，再提交正式记录');
        return;
      }
      await widget.client.submitDraft(_token, draftId);
      if (!mounted) {
        return;
      }
      _showMessage('已提交正式测评记录', tone: PadMessageTone.success);
      await Future<void>.delayed(const Duration(milliseconds: 650));
      if (mounted) {
        widget.onBack();
      }
    } on Object catch (error) {
      if (mounted) {
        _showMessage('提交记录失败：$error');
      }
    } finally {
      if (mounted) {
        setState(() => _submitting = false);
      }
    }
  }

  Map<String, dynamic> _buildDraftPayload() {
    return <String, dynamic>{
      if (_draftId > 0) 'id': _draftId,
      'studentId': _studentId,
      'studentName': _studentName.trim(),
      'examinerName': _examinerName.trim(),
      'birthDate': _dateOnlyText(_birthDate),
      'assessmentDate': _dateOnlyText(_assessmentDate),
      'itemPassList': _itemPassList(),
      if (_itemRemarks.isNotEmpty) 'itemRemarkList': _itemRemarkList(),
    };
  }

  List<Map<String, dynamic>> _itemPassList() {
    final List<int> itemNos = _itemPasses.keys.toList()..sort();
    return <Map<String, dynamic>>[
      for (final int itemNo in itemNos)
        <String, dynamic>{
          'itemNo': itemNo,
          'passed': _itemPasses[itemNo],
          'remark': _itemRemarks[itemNo]?.trim() ?? '',
        },
    ];
  }

  List<Map<String, dynamic>> _itemRemarkList() {
    final List<int> itemNos = _itemRemarks.keys.toList()..sort();
    return <Map<String, dynamic>>[
      for (final int itemNo in itemNos)
        if ((_itemRemarks[itemNo] ?? '').trim().isNotEmpty)
          <String, dynamic>{
            'itemNo': itemNo,
            'remark': (_itemRemarks[itemNo] ?? '').trim(),
          },
    ];
  }

  void _applyDraftDetail(ErxinDraftDetail detail) {
    _draftId = detail.id;
    _draftProgress = detail.progress;
    if (detail.studentId > 0) {
      _studentId = detail.studentId;
    }
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
    _itemPasses
      ..clear()
      ..addAll(detail.input.itemPasses);
    _itemRemarks
      ..clear()
      ..addAll(detail.input.itemRemarks);
    _restoreAssessmentWindowsFromAnswers();
  }

  void _restoreAssessmentWindowsFromAnswers() {
    _previousStartIndexByDomain.clear();
    _futureEndIndexByDomain.clear();
    _futureVisibleDomains.clear();
    _reviewMonthByDomain.clear();
    final int mainIndex = _mainAgeIndex;
    if (mainIndex < 0) {
      return;
    }
    for (final ErxinDomain domain in _template.domains) {
      final String domainCode = domain.domainCode;
      int previousStart = _defaultPreviousStartIndex;
      bool hasPreviousAnswer = false;
      int futureEnd = math.min(
          _ErxinAssessmentPageState._standardAgeMonths.length - 1,
          mainIndex + 2);
      bool hasFutureAnswer = false;
      for (final ErxinAgeGroup group in _template.ageGroups) {
        final int ageIndex = _ErxinAssessmentPageState._standardAgeMonths
            .indexOf(group.ageMonth);
        if (ageIndex < 0) {
          continue;
        }
        final bool hasAnsweredItem = group.items.any(
          (ErxinItemSummary item) =>
              item.domainCode == domainCode &&
              _itemPasses.containsKey(item.itemNo),
        );
        if (!hasAnsweredItem) {
          continue;
        }
        if (ageIndex < mainIndex) {
          hasPreviousAnswer = true;
          previousStart = math.min(previousStart, ageIndex);
        } else if (ageIndex > mainIndex) {
          hasFutureAnswer = true;
          futureEnd = math.max(futureEnd, ageIndex);
        }
      }
      if (hasPreviousAnswer) {
        _previousStartIndexByDomain[domainCode] = previousStart;
      }
      if (hasFutureAnswer) {
        _futureVisibleDomains.add(domainCode);
        _futureEndIndexByDomain[domainCode] = math.min(
            futureEnd, _ErxinAssessmentPageState._standardAgeMonths.length - 1);
      }
      _trimPreviousWindowToActiveBaseline(domainCode);
      if (_futureVisibleDomains.contains(domainCode)) {
        _futureEndIndexByDomain[domainCode] = _trimFutureEndToActiveCeiling(
          domainCode,
          _futureEndIndexByDomain[domainCode] ?? futureEnd,
        );
      }
    }
  }

  String? _localSubmitBlocker() {
    if (_token.trim().isEmpty) {
      return '请先登录后再提交正式记录';
    }
    if (_birthDate.trim().isEmpty || _assessmentDate.trim().isEmpty) {
      return '缺少出生日期或测查日期，不能提交正式记录';
    }
    for (final ErxinDomain domain in _template.domains) {
      final String code = domain.domainCode;
      for (final int month in _visibleMonthsForDomain(code)) {
        for (final ErxinItemSummary item in _itemsFor(code, month)) {
          if (!_itemPasses.containsKey(item.itemNo)) {
            _selectDomain(code);
            return '${_domainName(code)}还有当前可见题目未记录，请补全后再提交';
          }
        }
      }
      if (_canContinuePreviousMonthsForDomain(code)) {
        _selectDomain(code);
        return '${_domainName(code)}尚未形成连续两个往前月龄全通过，请继续往前测查';
      }
      if (_canEnterFutureMonthsForDomain(code)) {
        _selectDomain(code);
        return _hasPreviousBaselineForDomain(code)
            ? '${_domainName(code)}已建立前测基线，请先进入往后测查'
            : '${_domainName(code)}已测至最低月龄，请先进入往后测查';
      }
      if (_canContinueFutureMonthsForDomain(code)) {
        _selectDomain(code);
        return '${_domainName(code)}尚未形成连续两个往后月龄全不通过，请继续往后测查';
      }
      if (!_domainStopRuleComplete(code)) {
        _selectDomain(code);
        return '${_domainName(code)}尚未满足儿心量表停止规则，请完成规则提示的测查';
      }
    }
    return null;
  }

  void _continuePreviousMonths() {
    final String domainCode = _selectedDomainCode;
    final int currentStart = _previousStartIndexForDomain(domainCode);
    if (currentStart <= 0) {
      return;
    }
    setState(() {
      _reviewMonthByDomain.remove(domainCode);
      int start = currentStart;
      while (start > 0) {
        final int lowestVisibleMonth =
            _ErxinAssessmentPageState._standardAgeMonths[start];
        final int step =
            _ageMonthAllPassed(domainCode, lowestVisibleMonth) ? 1 : 2;
        final int nextStart = math.max(0, start - step);
        if (nextStart == start) {
          break;
        }
        _previousStartIndexByDomain[domainCode] = nextStart;
        start = nextStart;
        if (_firstPendingCurrentItemNo(domainCode) > 0 ||
            _hasPreviousBaselineForDomain(domainCode) ||
            !_previousMonthsCompleteForDomain(domainCode)) {
          break;
        }
      }
      _selectedItemNo = _firstPendingCurrentItemNo(domainCode);
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstCurrentItemNo(domainCode);
      }
      if (_selectedItemNo <= 0) {
        _selectedItemNo = _firstVisibleItemNo(domainCode);
      }
      final List<int> addedMonths = _ErxinAssessmentPageState._standardAgeMonths
          .sublist(start, currentStart)
          .reversed
          .toList(growable: false);
      if (addedMonths.isNotEmpty) {
        _recordRevealMonths = addedMonths;
        _recordRevealSerial++;
      }
      _autoSaveText = '已追加往前测查';
    });
    _prefetchSelectedItem();
    _revealSelectedItem();
  }

  void _enterFutureMonths() {
    final int index = _mainAgeIndex;
    final String domainCode = _selectedDomainCode;
    final int endIndex = math.min(
        _ErxinAssessmentPageState._standardAgeMonths.length - 1, index + 2);
    setState(() {
      _reviewMonthByDomain.remove(domainCode);
      _futureVisibleDomains.add(domainCode);
      _futureEndIndexByDomain[domainCode] = endIndex;
      _selectedItemNo = _firstItemNoForMonth(
        domainCode,
        _ErxinAssessmentPageState._standardAgeMonths[index + 1],
      );
      _recordRevealMonths = _ErxinAssessmentPageState._standardAgeMonths
          .sublist(index + 1, endIndex + 1)
          .toList(growable: false);
      if (_recordRevealMonths.isNotEmpty) {
        _recordRevealSerial++;
      }
      _autoSaveText = '已进入往后测查';
    });
    _prefetchSelectedItem();
    _revealSelectedItem();
  }

  void _continueFutureMonths() {
    final int index = _mainAgeIndex;
    if (index < 0) {
      return;
    }
    final String domainCode = _selectedDomainCode;
    final int currentEnd = _futureEndIndexByDomain[domainCode] ??
        math.min(
            _ErxinAssessmentPageState._standardAgeMonths.length - 1, index + 2);
    if (currentEnd >= _ErxinAssessmentPageState._standardAgeMonths.length - 1) {
      return;
    }
    final int highestVisibleMonth =
        _ErxinAssessmentPageState._standardAgeMonths[currentEnd];
    final int step = _ageMonthAllFailed(
      domainCode,
      highestVisibleMonth,
    )
        ? 1
        : 2;
    final int nextEnd = math.min(
        _ErxinAssessmentPageState._standardAgeMonths.length - 1,
        currentEnd + step);
    setState(() {
      _reviewMonthByDomain.remove(domainCode);
      _futureEndIndexByDomain[domainCode] = nextEnd;
      _selectedItemNo = _firstItemNoForMonth(
        domainCode,
        _ErxinAssessmentPageState._standardAgeMonths[currentEnd + 1],
      );
      _recordRevealMonths = _ErxinAssessmentPageState._standardAgeMonths
          .sublist(currentEnd + 1, nextEnd + 1)
          .toList(growable: false);
      if (_recordRevealMonths.isNotEmpty) {
        _recordRevealSerial++;
      }
      _autoSaveText = '已追加往后测查';
    });
    _prefetchSelectedItem();
    _revealSelectedItem();
  }

  void _showFullItemDetail(
    ErxinAssessmentItem? item,
    ErxinItemSummary? summary,
  ) {
    final String title = item == null || item.itemNo <= 0
        ? summary == null
            ? '题目说明'
            : '${summary.itemNo} ${summary.itemTitle}'
        : '${item.itemNo} ${item.itemTitle}';
    showDialog<void>(
      context: context,
      barrierColor: Colors.black.withOpacity(.24),
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: Center(
            child: AlertDialog(
              title: Text(title),
              content: SizedBox(
                width: 640,
                child: SingleChildScrollView(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      _DialogTextBlock(title: '操作方法', text: item?.method ?? ''),
                      const SizedBox(height: 18),
                      _DialogTextBlock(
                        title: '通过标准',
                        text: item?.passCriteria ?? '',
                      ),
                    ],
                  ),
                ),
              ),
              actions: <Widget>[
                TextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: const Text('关闭'),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  _DomainProgress _domainProgress(String domainCode) {
    final List<ErxinItemSummary> items = <ErxinItemSummary>[
      for (final int month in _visibleMonthsForDomain(domainCode))
        ..._itemsFor(domainCode, month),
    ];
    final int answered = items
        .where((ErxinItemSummary item) => _itemPasses.containsKey(item.itemNo))
        .length;
    return _DomainProgress(answered: answered, total: items.length);
  }

  int _completedDomainCount() {
    return _template.domains
        .where(
            (ErxinDomain domain) => _domainStopRuleComplete(domain.domainCode))
        .length
        .clamp(0, 5);
  }

  bool _domainStopRuleComplete(String domainCode) {
    if (!_mainMonthCompleteForDomain(domainCode)) {
      return false;
    }
    if (!_previousSearchResolvedForDomain(domainCode)) {
      return false;
    }
    return _futureSearchResolvedForDomain(domainCode);
  }

  String _recordStatusText(String domainCode, int month) {
    if (month == _mainAgeMonth) {
      return _ageMonthComplete(domainCode, month) ? '已完成' : '未完成';
    }
    if (month < _mainAgeMonth) {
      return _ageMonthAllPassed(domainCode, month)
          ? '全通过'
          : _ageMonthComplete(domainCode, month)
              ? '未全通过'
              : '未完成';
    }
    return _ageMonthAllFailed(domainCode, month)
        ? '全不通过'
        : _ageMonthComplete(domainCode, month)
            ? '未全不通过'
            : '未完成';
  }

  bool _recordDone(String domainCode, int month) {
    if (month == _mainAgeMonth) {
      return _ageMonthComplete(domainCode, month);
    }
    if (month < _mainAgeMonth) {
      return _ageMonthAllPassed(domainCode, month);
    }
    return _ageMonthAllFailed(domainCode, month);
  }

  bool _hasAnsweredFutureMonth(String domainCode) {
    return _highestAnsweredFutureIndex(domainCode) >= 0;
  }

  int _highestAnsweredFutureIndex(String domainCode) {
    int highest = -1;
    for (final ErxinAgeGroup group in _template.ageGroups) {
      final int index =
          _ErxinAssessmentPageState._standardAgeMonths.indexOf(group.ageMonth);
      if (index <= _mainAgeIndex) {
        continue;
      }
      final bool hasAnswered = group.items.any(
        (ErxinItemSummary item) =>
            item.domainCode == domainCode &&
            _itemPasses.containsKey(item.itemNo),
      );
      if (hasAnswered) {
        highest = math.max(highest, index);
      }
    }
    return highest;
  }

  void _reconcileAfterScore(String domainCode) {
    _trimPreviousWindowToActiveBaseline(domainCode);
    final bool previousReady = _mainMonthCompleteForDomain(domainCode) &&
        _previousSearchResolvedForDomain(domainCode);
    if (!previousReady) {
      _futureVisibleDomains.remove(domainCode);
      _futureEndIndexByDomain.remove(domainCode);
      return;
    }

    final int highestAnsweredFuture = _highestAnsweredFutureIndex(domainCode);
    if (!_futureVisibleDomains.contains(domainCode) &&
        highestAnsweredFuture < 0) {
      _futureEndIndexByDomain.remove(domainCode);
      return;
    }

    _futureVisibleDomains.add(domainCode);
    final int defaultEnd = math.min(
        _ErxinAssessmentPageState._standardAgeMonths.length - 1,
        _mainAgeIndex + 2);
    final int existingEnd = _futureEndIndexByDomain[domainCode] ?? defaultEnd;
    final int end = math.max(existingEnd, highestAnsweredFuture);
    final int normalizedEnd = math.max(defaultEnd,
        math.min(end, _ErxinAssessmentPageState._standardAgeMonths.length - 1));
    _futureEndIndexByDomain[domainCode] =
        _trimFutureEndToActiveCeiling(domainCode, normalizedEnd);
  }

  void _trimPreviousWindowToActiveBaseline(String domainCode) {
    final int mainIndex = _mainAgeIndex;
    if (mainIndex <= 0) {
      return;
    }
    final int currentStart = _previousStartIndexForDomain(domainCode);
    int? bestStart;
    for (int index = mainIndex - 2; index >= currentStart; index--) {
      final int current = _ErxinAssessmentPageState._standardAgeMonths[index];
      final int next = _ErxinAssessmentPageState._standardAgeMonths[index + 1];
      if (_ageMonthAllPassed(domainCode, current) &&
          _ageMonthAllPassed(domainCode, next)) {
        bestStart = index;
        break;
      }
    }
    if (bestStart != null && bestStart != currentStart) {
      _previousStartIndexByDomain[domainCode] = bestStart;
    }
  }

  int _trimFutureEndToActiveCeiling(String domainCode, int endIndex) {
    final int mainIndex = _mainAgeIndex;
    if (mainIndex < 0 || endIndex <= mainIndex + 1) {
      return endIndex;
    }
    for (int index = mainIndex + 1; index < endIndex; index++) {
      final int current = _ErxinAssessmentPageState._standardAgeMonths[index];
      final int next = _ErxinAssessmentPageState._standardAgeMonths[index + 1];
      if (_ageMonthAllFailed(domainCode, current) &&
          _ageMonthAllFailed(domainCode, next)) {
        return index + 1;
      }
    }
    return endIndex;
  }

  List<ErxinItemSummary> _itemsFor(String domainCode, int ageMonth) {
    for (final ErxinAgeGroup group in _template.ageGroups) {
      if (group.ageMonth == ageMonth) {
        return group.items
            .where((ErxinItemSummary item) => item.domainCode == domainCode)
            .toList();
      }
    }
    return <ErxinItemSummary>[];
  }

  bool _ageMonthComplete(String domainCode, int ageMonth) {
    final List<ErxinItemSummary> items = _itemsFor(domainCode, ageMonth);
    return items.isNotEmpty &&
        items.every(
            (ErxinItemSummary item) => _itemPasses.containsKey(item.itemNo));
  }

  bool _ageMonthAllPassed(String domainCode, int ageMonth) {
    final List<ErxinItemSummary> items = _itemsFor(domainCode, ageMonth);
    return items.isNotEmpty &&
        items
            .every((ErxinItemSummary item) => _itemPasses[item.itemNo] == true);
  }

  bool _ageMonthAllFailed(String domainCode, int ageMonth) {
    final List<ErxinItemSummary> items = _itemsFor(domainCode, ageMonth);
    return items.isNotEmpty &&
        items.every(
            (ErxinItemSummary item) => _itemPasses[item.itemNo] == false);
  }

  int _firstVisibleItemNo(String domainCode) {
    for (final int month in _visibleMonthsForDomain(domainCode)) {
      final int itemNo = _firstItemNoForMonth(domainCode, month);
      if (itemNo > 0) {
        return itemNo;
      }
    }
    return 0;
  }

  int _firstCurrentItemNo(String domainCode) {
    final int pendingItemNo = _firstPendingCurrentItemNo(domainCode);
    if (pendingItemNo > 0) {
      return pendingItemNo;
    }
    for (final int month in _centerMonthsForDomain(domainCode)) {
      final int itemNo = _firstItemNoForMonth(domainCode, month);
      if (itemNo > 0) {
        return itemNo;
      }
    }
    return 0;
  }

  int _firstPendingCurrentItemNo(String domainCode) {
    for (final int month in _centerMonthsForDomain(domainCode)) {
      for (final ErxinItemSummary item in _itemsFor(domainCode, month)) {
        if (!_itemPasses.containsKey(item.itemNo)) {
          return item.itemNo;
        }
      }
    }
    return 0;
  }

  int _firstPendingAssessmentItemNoForDomain(String domainCode) {
    for (final int month in _visibleMonthsForDomain(domainCode)) {
      for (final ErxinItemSummary item in _itemsFor(domainCode, month)) {
        if (!_itemPasses.containsKey(item.itemNo)) {
          return item.itemNo;
        }
      }
    }
    return 0;
  }

  int _nextSelectedItemNoForDomain(String domainCode, int fallbackItemNo) {
    if (_reviewMonthByDomain.containsKey(domainCode)) {
      return fallbackItemNo;
    }
    final int pendingItemNo = _firstPendingCurrentItemNo(domainCode);
    return pendingItemNo > 0 ? pendingItemNo : fallbackItemNo;
  }

  int _firstItemNoForMonth(String domainCode, int ageMonth) {
    for (final ErxinItemSummary item in _itemsFor(domainCode, ageMonth)) {
      return item.itemNo;
    }
    return 0;
  }

  ErxinItemSummary? _summaryByNo(int itemNo) {
    for (final ErxinAgeGroup group in _template.ageGroups) {
      for (final ErxinItemSummary item in group.items) {
        if (item.itemNo == itemNo) {
          return item;
        }
      }
    }
    return null;
  }

  String _domainName(String domainCode) {
    for (final ErxinDomain domain in _template.domains) {
      if (domain.domainCode == domainCode) {
        return domain.domainName.trim().isEmpty
            ? domain.domainCode
            : domain.domainName;
      }
    }
    return domainCode;
  }

  String _resolvedStudentAgeText() {
    final String calculated = formatAssessmentAgeText(
      birthDate: _birthDate,
      assessmentDate: _assessmentDate,
    );
    if (calculated.isNotEmpty) {
      return calculated;
    }
    final String explicit = _studentAge.trim();
    if (explicit.isNotEmpty && explicit != '未知') {
      return explicit;
    }
    final double months = _actualAgeMonths(_birthDate, _assessmentDate);
    if (months <= 0) {
      return '未知';
    }
    final int years = (months ~/ 12).clamp(0, 6);
    final int remainingMonths = (months - years * 12).round().clamp(0, 11);
    if (years > 0) {
      return remainingMonths > 0 ? '$years岁$remainingMonths个月' : '$years岁';
    }
    return '$remainingMonths个月';
  }

  String _nextActionText() {
    for (final int month in _visibleMonths) {
      for (final ErxinItemSummary item
          in _itemsFor(_selectedDomainCode, month)) {
        if (!_itemPasses.containsKey(item.itemNo)) {
          return '完成$month月龄第${item.itemNo}题';
        }
      }
    }
    if (_canEnterFutureMonths) {
      return _hasPreviousBaseline ? '前测已达标，可以进入往后测查' : '已到最低月龄，前测强行停止，可进入往后测查';
    }
    if (_canContinuePreviousMonths) {
      final int currentStart =
          _previousStartIndexForDomain(_selectedDomainCode);
      final int lowestVisibleMonth =
          _ErxinAssessmentPageState._standardAgeMonths[currentStart];
      final int step =
          _ageMonthAllPassed(_selectedDomainCode, lowestVisibleMonth) ? 1 : 2;
      final int nextStart = math.max(0, currentStart - step);
      final List<int> nextMonths = _ErxinAssessmentPageState._standardAgeMonths
          .sublist(nextStart, currentStart);
      return '前测未形成连续全通过，继续追加${nextMonths.join('月、')}月';
    }
    if (_canContinueFutureMonths) {
      final int index = _mainAgeIndex;
      final int currentEnd = _futureEndIndexByDomain[_selectedDomainCode] ??
          math.min(_ErxinAssessmentPageState._standardAgeMonths.length - 1,
              index + 2);
      final int highestVisibleMonth =
          _ErxinAssessmentPageState._standardAgeMonths[currentEnd];
      final int step = _ageMonthAllFailed(
        _selectedDomainCode,
        highestVisibleMonth,
      )
          ? 1
          : 2;
      final int nextEnd = math.min(
          _ErxinAssessmentPageState._standardAgeMonths.length - 1,
          currentEnd + step);
      final List<int> nextMonths = _ErxinAssessmentPageState._standardAgeMonths
          .sublist(currentEnd + 1, nextEnd + 1);
      return '后测未形成连续全不通过，继续追加${nextMonths.join('月、')}月';
    }
    if (_hasFutureCeiling) {
      return '当前能区停止规则已满足，可切换下一个能区或提交';
    }
    if (_domainStopRuleComplete(_selectedDomainCode)) {
      return '当前能区边界停止，可切换下一个能区或提交';
    }
    if (_futureVisibleDomains.contains(_selectedDomainCode) &&
        !_hasFutureCeiling) {
      return _futureBoundaryStopForDomain(_selectedDomainCode)
          ? '已到最高月龄，后测强行停止'
          : _futureMonthsComplete
              ? '已到最高可追测月龄，仍未形成连续全不通过'
              : '先完成当前可见的往后测查题目';
    }
    if (!_hasPreviousBaseline) {
      return _previousBoundaryStopForDomain(_selectedDomainCode)
          ? '已到最低可追测月龄，仍未形成连续全通过'
          : _previousMonthsComplete
              ? '已到最低可追测月龄，仍未形成连续全通过'
              : '先完成当前可见的往前测查题目';
    }
    return '当前可见题目已完成';
  }
}
