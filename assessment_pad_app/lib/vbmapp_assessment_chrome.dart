part of 'vbmapp_assessment_page.dart';

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
      padding: const EdgeInsets.symmetric(horizontal: 18),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.98),
        border: Border.all(color: _VbmappColors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _vbmappShadow(color: const Color(0x14B05F32), blur: 14),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1280;
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
              const SizedBox(width: 10),
              Expanded(
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  physics: const ClampingScrollPhysics(),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: headerChildren,
                  ),
                ),
              ),
              _VbmappSaveStatusLabel(text: status, saving: saving),
              const SizedBox(width: 10),
              _VbmappTopActionButton(
                label: '保存草稿',
                icon: Icons.save_outlined,
                loading: saving,
                filled: false,
                onTap: onSave,
              ),
              const SizedBox(width: 9),
              _VbmappTopActionButton(
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
    return saving ||
        text.contains('保存中') ||
        text.contains('提交中') ||
        text.contains('草稿保存中') ||
        text.startsWith('正在');
  }

  bool get _failed {
    return text.contains('失败');
  }

  @override
  Widget build(BuildContext context) {
    return ConstrainedBox(
      constraints: const BoxConstraints(minWidth: 116),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        mainAxisAlignment: MainAxisAlignment.end,
        children: <Widget>[
          Icon(
            _failed
                ? Icons.error_outline_rounded
                : (_activeSaving
                    ? Icons.sync_rounded
                    : Icons.check_circle_outline_rounded),
            color: _failed
                ? _VbmappColors.red
                : (_activeSaving
                    ? _VbmappColors.orangeDeep
                    : _VbmappColors.green),
            size: _activeSaving ? 17 : 18,
          ),
          const SizedBox(width: 5),
          Text(
            text,
            maxLines: 1,
            softWrap: false,
            textAlign: TextAlign.right,
            style: TextStyle(
              color: _failed
                  ? _VbmappColors.red
                  : (_activeSaving
                      ? _VbmappColors.orangeDeep
                      : _VbmappColors.body),
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
    required this.loading,
    required this.filled,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final bool loading;
  final bool filled;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final bool enabled = onTap != null;
    final bool interactive = enabled && !loading;
    final Color foreground = filled
        ? Colors.white
        : enabled
            ? _VbmappColors.orangeDeep
            : _VbmappColors.muted;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: interactive ? onTap : null,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 36,
          padding: const EdgeInsets.symmetric(horizontal: 13),
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
            boxShadow: filled && enabled
                ? _vbmappShadow(
                    color: const Color(0x2AE96F43),
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
                    color: filled ? Colors.white : _VbmappColors.orange,
                  ),
                )
              else
                Icon(icon, size: 17, color: foreground),
              const SizedBox(width: 7),
              Text(
                label,
                style: TextStyle(
                  color: foreground,
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

class _VbmappRightRail extends StatelessWidget {
  const _VbmappRightRail({
    required this.progressPercent,
    required this.answered,
    required this.total,
    required this.selectedModule,
    required this.scoreDetails,
  });

  final double progressPercent;
  final int answered;
  final int total;
  final _VbmappModule selectedModule;
  final Widget scoreDetails;

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
              child: scoreDetails,
            ),
          ),
          const SizedBox(height: 12),
          const _VbmappLegend(),
        ],
      ),
    );
  }
}

class _VbmappRightRailScoreDetails extends StatelessWidget {
  const _VbmappRightRailScoreDetails({required this.snapshot});

  final _VbmappScoreSnapshot snapshot;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        _VbmappCurrentScoreCard(snapshot: snapshot),
        const SizedBox(height: 12),
        _VbmappMilestoneDomainScoreCard(
          domains: snapshot.milestoneDomains,
        ),
      ],
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

class _VbmappActiveObservationBar extends StatefulWidget {
  const _VbmappActiveObservationBar({
    required this.tone,
    required this.observation,
    required this.statusLabel,
    required this.sharedSummaryMode,
    required this.summaries,
    required this.recordCount,
    required this.qualifiedCount,
    required this.onePointTarget,
    required this.onJump,
    required this.onQuickRecord,
    required this.onPrimaryAction,
    required this.onFinish,
    required this.onAutoFinish,
  });

  final Color tone;
  final _VbmappObservationTimerState observation;
  final String statusLabel;
  final bool sharedSummaryMode;
  final List<_VbmappActiveObservationSummary> summaries;
  final int recordCount;
  final int qualifiedCount;
  final int onePointTarget;
  final VoidCallback onJump;
  final VoidCallback onQuickRecord;
  final VoidCallback onPrimaryAction;
  final VoidCallback onFinish;
  final VoidCallback onAutoFinish;

  @override
  State<_VbmappActiveObservationBar> createState() =>
      _VbmappActiveObservationBarState();
}

class _VbmappActiveObservationBarState
    extends State<_VbmappActiveObservationBar> {
  Timer? _ticker;
  bool _autoFinishRequested = false;

  @override
  void initState() {
    super.initState();
    _syncTicker();
  }

  @override
  void didUpdateWidget(covariant _VbmappActiveObservationBar oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.observation != widget.observation) {
      _autoFinishRequested = false;
      _syncTicker();
    }
  }

  @override
  void dispose() {
    _ticker?.cancel();
    super.dispose();
  }

  void _syncTicker() {
    _ticker?.cancel();
    _ticker = null;
    if (!widget.observation.isRunning) {
      return;
    }
    if (_finishIfWindowElapsed()) {
      return;
    }
    _ticker = Timer.periodic(const Duration(seconds: 1), (Timer timer) {
      if (!mounted) {
        timer.cancel();
        return;
      }
      if (_finishIfWindowElapsed()) {
        timer.cancel();
        return;
      }
      setState(() {});
    });
  }

  bool _finishIfWindowElapsed() {
    if (_autoFinishRequested || !widget.observation.isRunning) {
      return false;
    }
    if (widget.observation.elapsedSecondsAt(DateTime.now()) <
        widget.observation.plannedSeconds) {
      return false;
    }
    _autoFinishRequested = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        widget.onAutoFinish();
      }
    });
    return true;
  }

  @override
  Widget build(BuildContext context) {
    final bool running = widget.observation.isRunning;
    final int elapsedSeconds =
        widget.observation.elapsedSecondsAt(DateTime.now());
    final String elapsedText = _vbmappDurationText(elapsedSeconds);
    final String statusLabel = running
        ? widget.statusLabel
        : widget.statusLabel.replaceFirst('观察中', '观察暂停');
    final String summaryText = widget.sharedSummaryMode
        ? '$elapsedText · 已记录 ${widget.recordCount} 条'
        : '$elapsedText · ${widget.qualifiedCount}/${widget.onePointTarget}';
    final List<_VbmappActiveObservationSummary> summaries = widget.summaries;
    return Container(
      key: const ValueKey<String>('vbmapp-active-observation-bar'),
      height: 40,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.97),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
        boxShadow: _vbmappShadow(color: const Color(0x14B05F32), blur: 12),
      ),
      child: Row(
        children: <Widget>[
          InkWell(
            onTap: widget.onJump,
            borderRadius: BorderRadius.circular(999),
            child: Ink(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
              decoration: BoxDecoration(
                color: widget.tone.withOpacity(.1),
                borderRadius: BorderRadius.circular(999),
              ),
              child: Text(
                statusLabel,
                style: TextStyle(
                  color: widget.tone,
                  fontSize: 11,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
          ),
          const SizedBox(width: 7),
          if (widget.sharedSummaryMode && summaries.isNotEmpty)
            Row(
              mainAxisSize: MainAxisSize.min,
              children: <Widget>[
                Text(
                  elapsedText,
                  style: const TextStyle(
                    color: _VbmappColors.ink,
                    fontSize: 11.5,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(width: 8),
                for (int index = 0;
                    index < summaries.length;
                    index++) ...<Widget>[
                  if (index > 0) const SizedBox(width: 5),
                  _VbmappActiveObservationSummaryChip(
                    summary: summaries[index],
                    elapsedSeconds: elapsedSeconds,
                    tone: widget.tone,
                  ),
                ],
              ],
            )
          else
            Text(
              summaryText,
              style: const TextStyle(
                color: _VbmappColors.ink,
                fontSize: 11.5,
                fontWeight: FontWeight.w900,
              ),
            ),
          const Spacer(),
          _VbmappMand4TimerButton(
            key: const ValueKey<String>(
                'vbmapp-active-observation-quick-record'),
            icon: Icons.add_rounded,
            label: '记一条',
            filled: true,
            compact: true,
            onTap: widget.onQuickRecord,
          ),
          const SizedBox(width: 5),
          _VbmappObservationMiniIconButton(
            key: const ValueKey<String>('vbmapp-active-observation-primary'),
            icon: running ? Icons.pause_rounded : Icons.play_arrow_rounded,
            tooltip: running ? '暂停观察' : '继续观察',
            onTap: widget.onPrimaryAction,
          ),
          const SizedBox(width: 5),
          _VbmappObservationMiniIconButton(
            key: const ValueKey<String>('vbmapp-active-observation-finish'),
            icon: Icons.stop_rounded,
            tooltip: '结束观察',
            onTap: widget.onFinish,
          ),
        ],
      ),
    );
  }
}

class _VbmappActiveObservationSummary {
  const _VbmappActiveObservationSummary({
    required this.label,
    required this.value,
    required this.plannedSeconds,
    this.complete = false,
  });

  final String label;
  final String value;
  final int plannedSeconds;
  final bool complete;
}

class _VbmappActiveObservationSummaryChip extends StatelessWidget {
  const _VbmappActiveObservationSummaryChip({
    required this.summary,
    required this.elapsedSeconds,
    required this.tone,
  });

  final _VbmappActiveObservationSummary summary;
  final int elapsedSeconds;
  final Color tone;

  @override
  Widget build(BuildContext context) {
    final Color color = summary.complete ? _VbmappColors.green : tone;
    final int remainingSeconds = summary.plannedSeconds - elapsedSeconds;
    final String remainingText = _vbmappDurationText(
      remainingSeconds > 0 ? remainingSeconds : 0,
    );
    return Container(
      height: 24,
      padding: const EdgeInsets.symmetric(horizontal: 7),
      decoration: BoxDecoration(
        color: color.withOpacity(summary.complete ? .13 : .1),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: color.withOpacity(.24)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Text(
            '${summary.label} ${summary.value}',
            style: TextStyle(
              color: summary.complete ? _VbmappColors.green : _VbmappColors.ink,
              fontSize: 11,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 5),
          Text(
            '剩$remainingText',
            style: TextStyle(
              color:
                  summary.complete ? _VbmappColors.green : _VbmappColors.body,
              fontSize: 10.5,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappObservationMiniIconButton extends StatelessWidget {
  const _VbmappObservationMiniIconButton({
    required this.icon,
    required this.tooltip,
    required this.onTap,
    super.key,
  });

  final IconData icon;
  final String tooltip;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message: tooltip,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: onTap,
          borderRadius: BorderRadius.circular(8),
          child: Ink(
            width: 28,
            height: 28,
            decoration: BoxDecoration(
              color: const Color(0xFFFFFCFA),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: _VbmappColors.lineSoft),
            ),
            child: Icon(icon, size: 16, color: _VbmappColors.orangeDeep),
          ),
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
