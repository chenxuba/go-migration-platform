part of 'vbmapp_assessment_page.dart';

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
  static const Color red = Color(0xFFD85F4A);
  static const Color blue = Color(0xFF5D7F9F);
}

List<BoxShadow> _vbmappShadow({
  Color color = const Color(0x12B05F32),
  double blur = 18,
  Offset offset = const Offset(0, 10),
}) {
  return <BoxShadow>[BoxShadow(color: color, blurRadius: blur, offset: offset)];
}

class _VbmappWorkspace extends StatelessWidget {
  const _VbmappWorkspace({
    required this.item,
    required this.score,
    required this.responseSchema,
    required this.materialProfile,
    required this.mandEvents,
    required this.mandObservation,
    required this.onAddMandEvent,
    required this.onSubmitMandEvent,
    required this.onDeleteMandEvent,
    required this.onChangeMandObservation,
    required this.onSelectScore,
  });

  final _VbmappItem item;
  final num? score;
  final VbmappItemResponseSchema? responseSchema;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> mandEvents;
  final _VbmappObservationTimerState? mandObservation;
  final VoidCallback onAddMandEvent;
  final ValueChanged<_VbmappMandEvent> onSubmitMandEvent;
  final ValueChanged<int> onDeleteMandEvent;
  final ValueChanged<_VbmappObservationTimerState> onChangeMandObservation;
  final ValueChanged<num> onSelectScore;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
      decoration: _vbmappCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Expanded(
            child: SingleChildScrollView(
              padding: EdgeInsets.zero,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  _VbmappQuestionHeader(
                    item: item,
                    schema: responseSchema,
                    materialProfile: materialProfile,
                  ),
                  const SizedBox(height: 12),
                  _VbmappSmartEvidencePanel(
                    item: item,
                    schema: responseSchema,
                    materialProfile: materialProfile,
                    mandEvents: mandEvents,
                    mandObservation: mandObservation,
                    onAddMandEvent: onAddMandEvent,
                    onSubmitMandEvent: onSubmitMandEvent,
                    onDeleteMandEvent: onDeleteMandEvent,
                    onChangeMandObservation: onChangeMandObservation,
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

bool _hasPreparationData(
  VbmappItemResponseSchema? schema,
  VbmappMaterialProfile? materialProfile,
) {
  return (materialProfile != null &&
          (materialProfile.quickPickLabels.isNotEmpty ||
              materialProfile.preparationChecks.isNotEmpty ||
              materialProfile.quickPicksByField.isNotEmpty)) ||
      (schema != null && schema.qualityChecks.isNotEmpty);
}

bool _shouldShowPreparationEntry(
  VbmappItemResponseSchema? schema,
  VbmappMaterialProfile? materialProfile,
) {
  return schema?.showPreparationEntry == true &&
      _hasPreparationData(schema, materialProfile);
}

class _VbmappPreparationDialog extends StatelessWidget {
  const _VbmappPreparationDialog({
    required this.item,
    required this.schema,
    required this.materialProfile,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? schema;
  final VbmappMaterialProfile? materialProfile;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Material(
        color: Colors.transparent,
        child: Container(
          width: 860,
          padding: const EdgeInsets.fromLTRB(20, 18, 20, 18),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: _VbmappColors.line),
            boxShadow: _vbmappShadow(blur: 24),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Icon(
                    Icons.inventory_2_outlined,
                    color: item.color,
                    size: 20,
                  ),
                  const SizedBox(width: 8),
                  const Expanded(
                    child: Text(
                      '题前准备',
                      style: TextStyle(
                        color: _VbmappColors.ink,
                        fontSize: 18,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  _VbmappIconButtonBox(
                    icon: Icons.close_rounded,
                    onTap: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Flexible(
                fit: FlexFit.loose,
                child: ConstrainedBox(
                  constraints: const BoxConstraints(maxHeight: 560),
                  child: SingleChildScrollView(
                    padding: EdgeInsets.zero,
                    child: _VbmappPreparationContent(
                      item: item,
                      schema: schema,
                      materialProfile: materialProfile,
                    ),
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

class _VbmappPreparationContent extends StatelessWidget {
  const _VbmappPreparationContent({
    required this.item,
    required this.schema,
    required this.materialProfile,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? schema;
  final VbmappMaterialProfile? materialProfile;

  @override
  Widget build(BuildContext context) {
    final VbmappMaterialProfile? profile = materialProfile;
    final List<String> quickPicks =
        profile?.quickPickLabels ?? const <String>[];
    final List<String> checks = _deduplicatedTexts(<String>[
      ...?profile?.preparationChecks,
      ...?schema?.qualityChecks,
    ]);
    final Map<String, List<String>> fieldQuickPicks =
        _normalizedMaterialQuickPicks(
      profile?.quickPicksByField ?? const <String, Object?>{},
    );
    final List<MapEntry<String, List<String>>> visibleFieldQuickPicks =
        fieldQuickPicks.entries.take(6).toList(growable: false);
    return Container(
      padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          if (profile != null && profile.label.trim().isNotEmpty) ...<Widget>[
            Text(
              profile.label.trim(),
              style: const TextStyle(
                color: _VbmappColors.ink,
                fontSize: 13,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 10),
          ],
          if (quickPicks.isNotEmpty) ...<Widget>[
            _VbmappPreparationSection(
              title: '推荐素材',
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                children: <Widget>[
                  for (final String label in quickPicks.take(16))
                    _VbmappPreparationChip(label: label),
                ],
              ),
            ),
            const SizedBox(height: 12),
          ],
          if (fieldQuickPicks.isNotEmpty) ...<Widget>[
            _VbmappPreparationSection(
              title: '快捷词',
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  for (int index = 0;
                      index < visibleFieldQuickPicks.length;
                      index++) ...<Widget>[
                    if (index > 0) const SizedBox(height: 6),
                    Text(
                      '${_materialFieldLabel(visibleFieldQuickPicks[index].key)}：${visibleFieldQuickPicks[index].value.join('、')}',
                      style: const TextStyle(
                        color: _VbmappColors.body,
                        fontSize: 12.5,
                        height: 1.4,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const SizedBox(height: 12),
          ],
          if (checks.isNotEmpty)
            _VbmappPreparationSection(
              title: '准备检查',
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  for (final String check in checks)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 6),
                      child: Text(
                        '- $check',
                        style: const TextStyle(
                          color: _VbmappColors.body,
                          fontSize: 12.5,
                          height: 1.4,
                          fontWeight: FontWeight.w700,
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

class _VbmappPreparationSection extends StatelessWidget {
  const _VbmappPreparationSection({
    required this.title,
    required this.child,
  });

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        Text(
          title,
          style: const TextStyle(
            color: _VbmappColors.ink,
            fontSize: 12.5,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 6),
        child,
      ],
    );
  }
}

class _VbmappPreparationChip extends StatelessWidget {
  const _VbmappPreparationChip({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _VbmappColors.line),
      ),
      child: Text(
        label,
        style: const TextStyle(
          color: _VbmappColors.body,
          fontSize: 12,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _VbmappQuestionHeader extends StatelessWidget {
  const _VbmappQuestionHeader({
    required this.item,
    required this.schema,
    required this.materialProfile,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? schema;
  final VbmappMaterialProfile? materialProfile;

  @override
  Widget build(BuildContext context) {
    final Color accent = item.color;
    final _VbmappQuestionHeaderCopy headerCopy =
        _vbmappQuestionHeaderCopy(item);
    final Widget badge = Container(
      padding: const EdgeInsets.fromLTRB(9, 6, 10, 6),
      decoration: BoxDecoration(
        color: accent.withOpacity(.08),
        borderRadius: BorderRadius.circular(9),
        border: Border.all(color: accent.withOpacity(.22)),
      ),
      child: Wrap(
        spacing: 7,
        runSpacing: 5,
        crossAxisAlignment: WrapCrossAlignment.center,
        children: <Widget>[
          Container(
            width: 3,
            height: 18,
            decoration: BoxDecoration(
              color: accent,
              borderRadius: BorderRadius.circular(999),
            ),
          ),
          Text(
            _vbmappQuestionDomainLabel(item),
            style: TextStyle(
              color: accent,
              fontSize: 14,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          Text(
            _vbmappQuestionStepLabel(item),
            style: TextStyle(
              color: accent,
              fontSize: 16,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          Container(
            width: 1,
            height: 14,
            color: accent.withOpacity(.28),
          ),
          Text(
            '${_vbmappQuestionStageLabel(item)} ${item.ageBand}',
            style: TextStyle(
              color: _VbmappColors.body,
              fontSize: 12,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          Container(
            width: 1,
            height: 14,
            color: accent.withOpacity(.28),
          ),
          Text(
            _vbmappQuestionModuleLabel(item.moduleCode),
            style: const TextStyle(
              color: _VbmappColors.body,
              fontSize: 12,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 7, vertical: 3),
            decoration: BoxDecoration(
              color: accent.withOpacity(.11),
              borderRadius: BorderRadius.circular(999),
              border: Border.all(color: accent.withOpacity(.18)),
            ),
            child: Text(
              item.assessmentMode,
              style: TextStyle(
                color: accent,
                fontSize: 11.5,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ],
      ),
    );

    final Widget title = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        for (int index = 0;
            index < headerCopy.lines.length;
            index++) ...<Widget>[
          if (index > 0) const SizedBox(height: 3),
          Text(
            headerCopy.lines[index],
            style: const TextStyle(
              color: _VbmappColors.ink,
              fontSize: 20,
              height: 1.2,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ],
    );

    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final bool hasPreparationContent =
            _shouldShowPreparationEntry(schema, materialProfile);
        final Widget? preparationAction = hasPreparationContent
            ? _VbmappPreparationEntryButton(
                item: item,
                schema: schema,
                materialProfile: materialProfile,
              )
            : null;
        final Widget badgeSlot = Expanded(
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            physics: const ClampingScrollPhysics(),
            child: Align(
              alignment: Alignment.centerLeft,
              child: badge,
            ),
          ),
        );
        if (constraints.maxWidth < 760) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: <Widget>[
                  badgeSlot,
                  if (preparationAction != null) ...<Widget>[
                    const SizedBox(width: 8),
                    preparationAction,
                  ],
                ],
              ),
              const SizedBox(height: 7),
              title,
            ],
          );
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: <Widget>[
                badgeSlot,
                if (preparationAction != null) ...<Widget>[
                  const SizedBox(width: 10),
                  preparationAction,
                ],
              ],
            ),
            const SizedBox(height: 8),
            title,
          ],
        );
      },
    );
  }
}

class _VbmappPreparationEntryButton extends StatelessWidget {
  const _VbmappPreparationEntryButton({
    required this.item,
    required this.schema,
    required this.materialProfile,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? schema;
  final VbmappMaterialProfile? materialProfile;

  @override
  Widget build(BuildContext context) {
    final VbmappMaterialProfile? profile = materialProfile;
    final List<String> quickPicks =
        profile?.quickPickLabels ?? const <String>[];
    final List<String> checks = _deduplicatedTexts(<String>[
      ...?profile?.preparationChecks,
      ...?schema?.qualityChecks,
    ]);
    final Color accent = item.color;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        key: const ValueKey<String>('vbmapp-preparation-entry'),
        onTap: () => _openPreparationDialog(context),
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: _VbmappColors.lineSoft),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(Icons.inventory_2_outlined, color: accent, size: 16),
              const SizedBox(width: 6),
              const Text(
                '题前准备',
                style: TextStyle(
                  color: _VbmappColors.ink,
                  fontSize: 12.5,
                  height: 1,
                  fontWeight: FontWeight.w900,
                ),
              ),
              if (quickPicks.isNotEmpty || checks.isNotEmpty) ...<Widget>[
                const SizedBox(width: 7),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
                  decoration: BoxDecoration(
                    color: accent.withOpacity(.08),
                    borderRadius: BorderRadius.circular(999),
                  ),
                  child: Text(
                    '${quickPicks.length}/${checks.length}',
                    style: TextStyle(
                      color: accent,
                      fontSize: 10.5,
                      height: 1,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _openPreparationDialog(BuildContext context) async {
    await showDialog<void>(
      context: context,
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: _VbmappPreparationDialog(
            item: item,
            schema: schema,
            materialProfile: materialProfile,
          ),
        );
      },
    );
  }
}

class _VbmappQuestionHeaderCopy {
  const _VbmappQuestionHeaderCopy({
    required this.lines,
  });

  final List<String> lines;
}

String _vbmappQuestionDomainLabel(_VbmappItem item) {
  final String domain = item.domainName.trim();
  if (domain.isNotEmpty) {
    return domain;
  }
  return _vbmappQuestionModuleLabel(item.moduleCode);
}

String _vbmappQuestionStepLabel(_VbmappItem item) {
  final String label = item.label.trim();
  final String domain = item.domainName.trim();
  if (domain.isNotEmpty && label.startsWith(domain)) {
    final String step = label.substring(domain.length).trim();
    if (step.isNotEmpty) {
      return step;
    }
  }
  if (item.moduleCode == 'milestones') {
    final RegExpMatch? match = RegExp(r'^(\d+)M$').firstMatch(item.navCode);
    if (match != null) {
      return '${match.group(1)}-M';
    }
    return item.navCode;
  }
  return item.itemCode;
}

_VbmappQuestionHeaderCopy _vbmappQuestionHeaderCopy(_VbmappItem item) {
  switch (item.itemCode) {
    case 'MAND_08M':
      return const _VbmappQuestionHeaderCopy(
        lines: <String>[
          '能提出5个不同的要求，其中至少要包含2个或2个以上的单词（不包括“我想要”）（如：跑快点、该我了、倒果汁）（TO：60分钟）',
        ],
      );
    case 'MAND_04M':
      return const _VbmappQuestionHeaderCopy(
        lines: <String>[
          '自发性地提出（没有口头辅助）5项要求，所要的物件可在眼前（TO：60分钟）',
        ],
      );
    case 'MAND_09M':
      return const _VbmappQuestionHeaderCopy(
        lines: <String>[
          '自发性地提出15个不同的要求，（如我们一起玩、打开、我想要书）（TO：30分钟）',
        ],
      );
  }
  return _VbmappQuestionHeaderCopy(
    lines: <String>[_vbmappQuestionTitleWithoutMode(item)],
  );
}

String _vbmappQuestionTitleWithoutMode(_VbmappItem item) {
  final String title = item.title.trim();
  if (title.isEmpty) {
    return '';
  }
  final String assessmentMode = item.assessmentMode.trim();
  if (assessmentMode.isEmpty) {
    return title;
  }
  return title
      .replaceFirst('（$assessmentMode）', '')
      .replaceFirst('($assessmentMode)', '')
      .trim();
}

String _vbmappQuestionStageLabel(_VbmappItem item) {
  switch (item.ageBand.trim()) {
    case '0-18个月':
      return '第一阶段';
    case '18-30个月':
      return '第二阶段';
    case '30-48个月':
      return '第三阶段';
    default:
      return '阶段';
  }
}

String _vbmappQuestionModuleLabel(String moduleCode) {
  switch (moduleCode) {
    case 'barriers':
      return '障碍评估';
    case 'transition':
      return '转衔评估';
    case 'milestones':
    default:
      return '里程碑评估';
  }
}

class _VbmappSmartEvidencePanel extends StatelessWidget {
  const _VbmappSmartEvidencePanel({
    required this.item,
    required this.schema,
    required this.materialProfile,
    required this.mandEvents,
    required this.mandObservation,
    required this.onAddMandEvent,
    required this.onSubmitMandEvent,
    required this.onDeleteMandEvent,
    required this.onChangeMandObservation,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? schema;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> mandEvents;
  final _VbmappObservationTimerState? mandObservation;
  final VoidCallback onAddMandEvent;
  final ValueChanged<_VbmappMandEvent> onSubmitMandEvent;
  final ValueChanged<int> onDeleteMandEvent;
  final ValueChanged<_VbmappObservationTimerState> onChangeMandObservation;

  @override
  Widget build(BuildContext context) {
    if (item.itemCode == 'MAND_01M' || item.itemCode == 'MAND_02M') {
      return _VbmappMand1InlinePanel(
        item: item,
        materialProfile: materialProfile,
        events: mandEvents,
        onSubmitEvent: onSubmitMandEvent,
        onDeleteEvent: onDeleteMandEvent,
      );
    }
    if (item.itemCode == 'MAND_03M') {
      return _VbmappMand3InlinePanel(
        item: item,
        materialProfile: materialProfile,
        events: mandEvents,
        onSubmitEvent: onSubmitMandEvent,
        onDeleteEvent: onDeleteMandEvent,
      );
    }
    if (item.itemCode == 'MAND_05M') {
      return _VbmappMand5InlinePanel(
        item: item,
        materialProfile: materialProfile,
        events: mandEvents,
        onSubmitEvent: onSubmitMandEvent,
        onDeleteEvent: onDeleteMandEvent,
      );
    }
    if (item.itemCode == 'MAND_06M') {
      return _VbmappMand6InlinePanel(
        item: item,
        responseSchema: schema,
        materialProfile: materialProfile,
        events: mandEvents,
        onSubmitEvent: onSubmitMandEvent,
        onDeleteEvent: onDeleteMandEvent,
      );
    }
    if (_isTimedMandItem(item, schema)) {
      return _VbmappTimedMandInlinePanel(
        item: item,
        responseSchema: schema,
        materialProfile: materialProfile,
        events: mandEvents,
        observation: mandObservation,
        onSubmitEvent: onSubmitMandEvent,
        onDeleteEvent: onDeleteMandEvent,
        onChangeObservation: onChangeMandObservation,
      );
    }
    if (_isLateMandRecorderItem(item, schema)) {
      return _VbmappLateMandInlinePanel(
        item: item,
        responseSchema: schema,
        materialProfile: materialProfile,
        events: mandEvents,
        onSubmitEvent: onSubmitMandEvent,
        onDeleteEvent: onDeleteMandEvent,
      );
    }
    if (_isSimpleMandRecorder(item, schema)) {
      return _VbmappMandRecorderPanel(
        item: item,
        materialProfile: materialProfile,
        events: mandEvents,
        onAddEvent: onAddMandEvent,
      );
    }
    return const SizedBox.shrink();
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
        _VbmappScoreOptionGrid(
          options: item.scoreOptions,
          selectedScore: score,
          accent: item.color,
          onSelectScore: onSelectScore,
        ),
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
          height: 58,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          decoration: BoxDecoration(
            color: selected ? accent : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: selected ? accent : _VbmappColors.line),
            boxShadow: selected
                ? _vbmappShadow(color: accent.withOpacity(.16), blur: 14)
                : null,
          ),
          child: Row(
            children: <Widget>[
              SizedBox(
                width: 42,
                child: Text(
                  option.displayScore,
                  style: TextStyle(
                    color: selected ? Colors.white : _VbmappColors.ink,
                    fontSize: 20,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              Container(
                width: 1,
                height: 28,
                color: selected
                    ? Colors.white.withOpacity(.28)
                    : _VbmappColors.lineSoft,
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  option.label,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: selected ? Colors.white : _VbmappColors.body,
                    fontSize: 11.5,
                    height: 1.18,
                    fontWeight: FontWeight.w800,
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

BoxDecoration _vbmappCardDecoration() {
  return BoxDecoration(
    color: Colors.white.withOpacity(.94),
    borderRadius: BorderRadius.circular(16),
    border: Border.all(color: _VbmappColors.line),
    boxShadow: _vbmappShadow(),
  );
}
