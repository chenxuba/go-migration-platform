part of 'vbmapp_assessment_page.dart';

class _VbmappLateMandInlinePanel extends StatefulWidget {
  const _VbmappLateMandInlinePanel({
    required this.item,
    required this.materialProfile,
    required this.events,
    required this.onSubmitEvent,
    required this.onDeleteEvent,
  });

  final _VbmappItem item;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> events;
  final ValueChanged<_VbmappMandEvent> onSubmitEvent;
  final ValueChanged<int> onDeleteEvent;

  @override
  State<_VbmappLateMandInlinePanel> createState() =>
      _VbmappLateMandInlinePanelState();
}

class _VbmappLateMandInlinePanelState
    extends State<_VbmappLateMandInlinePanel> {
  final TextEditingController _requestController = TextEditingController();
  final TextEditingController _environmentController = TextEditingController();

  late _VbmappLateMandConfig _config;
  String _promptChoice = '';
  String _environment = '';
  String _ability = '';
  String _targetKind = '';
  int? _selectedRecordIndex;

  @override
  void initState() {
    super.initState();
    _syncConfig();
  }

  @override
  void didUpdateWidget(covariant _VbmappLateMandInlinePanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.item.itemCode != widget.item.itemCode) {
      _syncConfig();
    }
    final int? selectedIndex = _selectedRecordIndex;
    if (selectedIndex != null && selectedIndex >= widget.events.length) {
      _selectedRecordIndex = null;
    }
  }

  @override
  void dispose() {
    _requestController.dispose();
    _environmentController.dispose();
    super.dispose();
  }

  void _syncConfig() {
    _config = _lateMandConfigFor(widget.item);
    _promptChoice = _firstLateMandOption(_config.promptOptions);
    _environment = _firstLateMandOption(_config.environmentOptions);
    _ability = _firstLateMandOption(_config.abilityOptions);
    _targetKind = _firstLateMandOption(
      _config.targetOptions,
      fallback: _config.defaultTargetKind,
    );
    _environmentController.text = '';
    _selectedRecordIndex = null;
  }

  int get _onePointCount => _scoreCountThreshold(widget.item, 1) ?? 5;

  int get _halfPointCount => _scoreCountThreshold(widget.item, .5) ?? 2;

  @override
  Widget build(BuildContext context) {
    final int qualifiedCount = _qualifiedMandCountForItem(
      widget.item,
      widget.events,
    );
    final double suggestedScore = _suggestMandScore(
      widget.item,
      widget.events,
    );
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool narrow = constraints.maxWidth < 780;
          final Widget form = _buildRecordForm();
          final Widget records = _buildRecordSummary();
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Icon(
                    Icons.record_voice_over_outlined,
                    color: widget.item.color,
                    size: 19,
                  ),
                  const SizedBox(width: 7),
                  Expanded(
                    child: Text(
                      '${_vbmappQuestionDomainLabel(widget.item)}'
                      '${widget.item.navCode}${_config.recordTitleSuffix}',
                      style: const TextStyle(
                        color: _VbmappColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  _VbmappEvidenceMetric(
                    label: _config.metricLabel,
                    value: '$qualifiedCount/$_onePointCount',
                    color: widget.item.color,
                  ),
                  const SizedBox(width: 8),
                  _VbmappEvidenceMetric(
                    label: '建议',
                    value: '${_formatScore(suggestedScore)}分',
                    color: widget.item.color,
                  ),
                ],
              ),
              const SizedBox(height: 9),
              if (narrow)
                Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: <Widget>[
                    form,
                    const SizedBox(height: 10),
                    records,
                  ],
                )
              else
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Expanded(flex: 5, child: form),
                    const SizedBox(width: 12),
                    Expanded(flex: 4, child: records),
                  ],
                ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildRecordForm() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        if (_choiceRows.isNotEmpty) ...<Widget>[
          for (final List<Widget> row in _choiceRows) ...<Widget>[
            if (row.length == 1)
              Align(alignment: Alignment.centerLeft, child: row.first)
            else
              Row(children: row),
            const SizedBox(height: 10),
          ],
          const SizedBox(height: 2),
        ],
        if (_config.freeTextEnvironment) ...<Widget>[
          _VbmappMandInlineTextField(
            controller: _environmentController,
            label: '环境',
            hintText: '输入本次发生的具体环境',
          ),
          const SizedBox(height: 10),
        ],
        Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            Expanded(
              child: _VbmappMandInlineTextField(
                controller: _requestController,
                label: _config.inputLabel,
                hintText: _config.inputHint,
              ),
            ),
            const SizedBox(width: 10),
            _VbmappSmallActionButton(
              icon: Icons.add_rounded,
              label: '记录本次要求',
              onTap: _submit,
            ),
          ],
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 6,
          children: <Widget>[
            for (final String value in _quickPicks)
              _VbmappMandMaterialChip(
                label: value,
                selected: _requestController.text.trim() == value,
                onTap: () => _selectQuickPick(value),
              ),
          ],
        ),
      ],
    );
  }

  List<List<Widget>> get _choiceRows {
    final List<Widget> firstRow = <Widget>[
      if (_config.promptOptions.isNotEmpty)
        _choiceCell(
          flex: 4,
          label: '辅助',
          value: _promptChoice,
          values: _config.promptOptions,
          onChanged: (String value) => setState(() => _promptChoice = value),
        ),
      if (_config.environmentOptions.isNotEmpty)
        _choiceCell(
          flex: 5,
          label: '环境',
          value: _environment,
          values: _config.environmentOptions,
          onChanged: (String value) => setState(() => _environment = value),
        ),
    ];
    final List<Widget> secondRow = <Widget>[
      if (_config.abilityOptions.isNotEmpty)
        _choiceCell(
          flex: 5,
          label: '能力',
          value: _ability,
          values: _config.abilityOptions,
          onChanged: (String value) => setState(() => _ability = value),
        ),
      if (_config.targetOptions.isNotEmpty)
        _choiceCell(
          flex: _config.targetOptions.length > 2 ? 7 : 5,
          label: '对象',
          value: _targetKind,
          values: _config.targetOptions,
          onChanged: (String value) => setState(() => _targetKind = value),
        ),
    ];
    return <List<Widget>>[
      if (firstRow.isNotEmpty) _withGaps(firstRow),
      if (secondRow.isNotEmpty) _withGaps(secondRow),
    ];
  }

  Widget _choiceCell({
    required int flex,
    required String label,
    required String value,
    required List<String> values,
    required ValueChanged<String> onChanged,
  }) {
    final Widget choice = _VbmappMandInlineChoiceGroup(
      label: label,
      value: value,
      values: values,
      onChanged: onChanged,
    );
    if (values.length <= 2) {
      return SizedBox(width: 280, child: choice);
    }
    return Expanded(
      flex: flex,
      child: choice,
    );
  }

  List<Widget> _withGaps(List<Widget> children) {
    final List<Widget> out = <Widget>[];
    for (int index = 0; index < children.length; index++) {
      if (index > 0) {
        out.add(const SizedBox(width: 10));
      }
      out.add(children[index]);
    }
    return out;
  }

  Widget _buildRecordSummary() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        _VbmappInlineInfo(
          icon: Icons.inventory_2_outlined,
          text: widget.materialProfile?.label ?? '潜在强化物/活动',
        ),
        const SizedBox(height: 10),
        Text(
          _config.recordListTitle,
          style: const TextStyle(
            color: _VbmappColors.ink,
            fontSize: 14,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 8),
        _VbmappMand1RecordGrid(
          events: widget.events,
          minSlots: _onePointCount,
          selectedIndex: _selectedRecordIndex,
          onSelectIndex: _selectRecord,
          onDeleteIndex: _deleteRecord,
          isQualified: (_VbmappMandEvent event) =>
              _mandEventCountsForItem(widget.item, event),
          metaTextBuilder: _recordMetaText,
        ),
        const SizedBox(height: 10),
        Text(
          _scoreReference,
          style: const TextStyle(
            color: _VbmappColors.body,
            fontSize: 12,
            height: 1.35,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    );
  }

  String get _scoreReference {
    return '参考：0${_config.unit}计0分，$_halfPointCount${_config.unit}计0.5分，'
        '$_onePointCount${_config.unit}计1分；${_config.countReferenceText}'
        '老师可在下方评分区覆盖。';
  }

  List<String> get _quickPicks {
    final List<String> profilePicks = <String>[
      ...?widget.materialProfile?.quickPicks,
      for (final VbmappMaterialSuggestion material
          in widget.materialProfile?.recommendedMaterials ??
              const <VbmappMaterialSuggestion>[])
        material.name,
    ];
    return _deduplicatedTexts(<String>[
      ..._config.quickPicks,
      ...profilePicks,
    ]).take(12).toList(growable: false);
  }

  String _recordMetaText(_VbmappMandEvent event) {
    final List<String> values = <String>[];

    void addMeta(String value) {
      final String normalized = value.trim();
      if (normalized.isEmpty || values.contains(normalized)) {
        return;
      }
      values.add(normalized);
    }

    addMeta(_mandInitiationText(event));
    addMeta(event.environment);
    addMeta(event.phraseLevel);
    if (!_shouldHideTargetKindInMandMeta(widget.item)) {
      addMeta(event.targetKind);
    }
    return values.isEmpty ? '未记录条件' : values.join(' · ');
  }

  void _selectQuickPick(String value) {
    setState(() {
      _requestController.text = value;
    });
  }

  void _submit() {
    final String request = _requestController.text.trim();
    if (request.isEmpty) {
      return;
    }
    final String environment = _config.freeTextEnvironment
        ? _environmentController.text.trim()
        : _environment;
    widget.onSubmitEvent(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: _config.motivationContext,
        environment: environment,
        targetKind: _config.targetOptions.isEmpty
            ? ''
            : _targetKind.trim().isEmpty
                ? _config.defaultTargetKind
                : _targetKind.trim(),
        person: '',
        setting: '',
        example: '',
        responseMode:
            _promptChoice == '提问下' ? '提问下要求' : _config.defaultResponseMode,
        promptLevel: _promptChoice,
        phraseLevel: _ability,
        functional: true,
      ),
    );
    _requestController.clear();
    setState(() {
      _selectedRecordIndex = null;
      _promptChoice = _firstLateMandOption(_config.promptOptions);
    });
  }

  void _selectRecord(int index) {
    setState(() {
      _selectedRecordIndex = _selectedRecordIndex == index ? null : index;
    });
  }

  void _deleteRecord(int index) {
    setState(() {
      _selectedRecordIndex = null;
    });
    widget.onDeleteEvent(index);
  }
}

class _VbmappLateMandConfig {
  const _VbmappLateMandConfig({
    required this.itemCode,
    required this.recordTitleSuffix,
    required this.recordListTitle,
    required this.metricLabel,
    required this.inputLabel,
    required this.inputHint,
    required this.quickPicks,
    required this.unit,
    required this.motivationContext,
    this.countReferenceText = '系统按有效要求记录自动去重统计，',
    this.promptOptions = const <String>[],
    this.environmentOptions = const <String>[],
    this.freeTextEnvironment = false,
    this.abilityOptions = const <String>[],
    this.targetOptions = const <String>[],
    this.defaultTargetKind = '',
    this.defaultResponseMode = '自发要求',
  });

  final String itemCode;
  final String recordTitleSuffix;
  final String recordListTitle;
  final String metricLabel;
  final String inputLabel;
  final String inputHint;
  final List<String> quickPicks;
  final String unit;
  final String motivationContext;
  final String countReferenceText;
  final List<String> promptOptions;
  final List<String> environmentOptions;
  final bool freeTextEnvironment;
  final List<String> abilityOptions;
  final List<String> targetOptions;
  final String defaultTargetKind;
  final String defaultResponseMode;
}

bool _isLateMandRecorderItem(_VbmappItem item) {
  return _lateMandConfigs.containsKey(item.itemCode);
}

String _firstLateMandOption(List<String> options, {String fallback = ''}) {
  return options.isEmpty ? fallback : options.first;
}

_VbmappLateMandConfig _lateMandConfigFor(_VbmappItem item) {
  return _lateMandConfigs[item.itemCode] ??
      _VbmappLateMandConfig(
        itemCode: item.itemCode,
        recordTitleSuffix: '现场记录',
        recordListTitle: '有效要求记录',
        metricLabel: '有效',
        inputLabel: '孩子要求内容',
        inputHint: '输入孩子实际发出的要求',
        quickPicks: const <String>[],
        unit: '个',
        motivationContext: '提要求',
      );
}

const Map<String, _VbmappLateMandConfig> _lateMandConfigs =
    <String, _VbmappLateMandConfig>{
  'MAND_07M': _VbmappLateMandConfig(
    itemCode: 'MAND_07M',
    recordTitleSuffix: '行动要求记录',
    recordListTitle: '不同行动要求',
    metricLabel: '有效',
    inputLabel: '孩子要求的行动',
    inputHint: '如：开门、推',
    quickPicks: <String>['开门', '推'],
    unit: '个',
    motivationContext: '要求他人行动',
    promptOptions: <String>['提问下', '自发地'],
    environmentOptions: <String>['呈现物品', '未呈现物品'],
    defaultTargetKind: '动作',
  ),
  'MAND_10M': _VbmappLateMandConfig(
    itemCode: 'MAND_10M',
    recordTitleSuffix: '新要求记录',
    recordListTitle: '新的要求',
    metricLabel: '有效',
    inputLabel: '孩子提出的新要求',
    inputHint: '如：小猫在哪里？',
    quickPicks: <String>['小猫在哪里？'],
    unit: '个',
    motivationContext: '新的要求',
    promptOptions: <String>['提问下', '自发地'],
    environmentOptions: <String>['呈现物品', '未呈现物品'],
    abilityOptions: <String>['新形式', '新内容'],
  ),
  'MAND_12M': _VbmappLateMandConfig(
    itemCode: 'MAND_12M',
    recordTitleSuffix: '礼貌拒绝记录',
    recordListTitle: '礼貌停止或移除要求',
    metricLabel: '有效',
    inputLabel: '孩子的礼貌要求',
    inputHint: '如：请别再推我、不了，谢谢你',
    quickPicks: <String>['请别再推我', '不了，谢谢你', '对不起', '你能让一下吗？'],
    unit: '个',
    motivationContext: '停止不喜欢活动或移除反感条件',
    countReferenceText: '系统按不同环境中的有效记录统计，',
    freeTextEnvironment: true,
    defaultTargetKind: '动作（终止或移除）',
  ),
  'MAND_14M': _VbmappLateMandConfig(
    itemCode: 'MAND_14M',
    recordTitleSuffix: '说明要求记录',
    recordListTitle: '指出/说明/解释记录',
    metricLabel: '有效',
    inputLabel: '孩子指出或说明的内容',
    inputHint: '如：你先涂胶水，再把它贴好',
    quickPicks: <String>['你先涂胶水，再把它贴好', '你坐在这里，我去拿一本书'],
    unit: '次',
    motivationContext: '说明如何做事或参与活动',
    countReferenceText: '系统按有效发生次数统计，',
    promptOptions: <String>['提问下', '自发地'],
    environmentOptions: <String>['呈现物品', '未呈现物品'],
    defaultTargetKind: '活动',
  ),
  'MAND_15M': _VbmappLateMandConfig(
    itemCode: 'MAND_15M',
    recordTitleSuffix: '对话注意记录',
    recordListTitle: '要求他人注意对话',
    metricLabel: '有效',
    inputLabel: '孩子发起注意的对话',
    inputHint: '如：听我说、让我来告诉你',
    quickPicks: <String>['听我说', '让我来告诉你', '当时发生的是', '我来说个故事'],
    unit: '次',
    motivationContext: '要求他人注意自己的对话行为',
    countReferenceText: '系统按有效发生次数统计，',
    environmentOptions: <String>['呈现物品', '未呈现物品'],
    defaultTargetKind: '对话行为',
  ),
};
