part of 'vbmapp_assessment_page.dart';

class _VbmappMand6InlinePanel extends StatefulWidget {
  const _VbmappMand6InlinePanel({
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
  State<_VbmappMand6InlinePanel> createState() =>
      _VbmappMand6InlinePanelState();
}

class _VbmappMand6InlinePanelState extends State<_VbmappMand6InlinePanel> {
  static const List<String> _fallbackMissingItems = <String>[
    '纸张',
    '蜡笔',
    '勺子',
    '碗',
    '吸管',
    '拼图块',
    '车轮',
    '钥匙',
    '盖子',
    '泡泡棒',
    '积木',
    '胶水',
  ];

  final TextEditingController _requestController = TextEditingController();

  String _promptChoice = '提问下';
  int? _selectedRecordIndex;

  int get _onePointRequestCount => _scoreCountThreshold(widget.item, 1) ?? 20;

  int get _halfPointRequestCount => _scoreCountThreshold(widget.item, .5) ?? 10;

  String get _recordTitle =>
      '${_vbmappQuestionDomainLabel(widget.item)}${widget.item.navCode}缺失物品记录';

  String get _scoreReference {
    return '参考：0个计0分，$_halfPointRequestCount个计0.5分，'
        '$_onePointRequestCount个计1分；仅提供“你想要什么？”提问或等待自发要求，'
        '系统按不同缺失物品自动去重。';
  }

  @override
  void didUpdateWidget(covariant _VbmappMand6InlinePanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    final int? selectedIndex = _selectedRecordIndex;
    if (selectedIndex != null && selectedIndex >= widget.events.length) {
      _selectedRecordIndex = null;
    }
  }

  @override
  void dispose() {
    _requestController.dispose();
    super.dispose();
  }

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
                    Icons.extension_outlined,
                    color: widget.item.color,
                    size: 19,
                  ),
                  const SizedBox(width: 7),
                  Expanded(
                    child: Text(
                      _recordTitle,
                      style: const TextStyle(
                        color: _VbmappColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  _VbmappEvidenceMetric(
                    label: '有效',
                    value: '$qualifiedCount/$_onePointRequestCount',
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
    final List<String> missingItems = _missingItemQuickPicks();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        Align(
          alignment: Alignment.centerLeft,
          child: SizedBox(
            width: 280,
            child: _VbmappMandInlineChoiceGroup(
              label: '辅助',
              value: _promptChoice,
              values: const <String>['提问下', '自发地'],
              onChanged: (String value) => setState(() {
                _promptChoice = value;
              }),
            ),
          ),
        ),
        const SizedBox(height: 12),
        Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            Expanded(
              child: _VbmappMandInlineTextField(
                controller: _requestController,
                label: '孩子要求的缺失物',
                hintText: '如：纸张、勺子、拼图块、车轮',
              ),
            ),
            const SizedBox(width: 10),
            _VbmappSmallActionButton(
              icon: Icons.add_rounded,
              label: '记录缺失物要求',
              onTap: _submit,
            ),
          ],
        ),
        const SizedBox(height: 8),
        Wrap(
          spacing: 8,
          runSpacing: 6,
          children: <Widget>[
            for (final String item in missingItems)
              _VbmappMandMaterialChip(
                label: item,
                selected: _requestController.text.trim() == item,
                onTap: () => _selectMissingItem(item),
              ),
          ],
        ),
      ],
    );
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
        const Text(
          '不同缺失物品',
          style: TextStyle(
            color: _VbmappColors.ink,
            fontSize: 14,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 8),
        _VbmappMand1RecordGrid(
          events: widget.events,
          minSlots: _onePointRequestCount,
          selectedIndex: _selectedRecordIndex,
          onSelectIndex: _selectRecord,
          onDeleteIndex: _deleteRecord,
          isQualified: (_VbmappMandEvent event) =>
              _mandEventCountsForItem(widget.item, event),
          metaTextBuilder: (_VbmappMandEvent event) =>
              _mandRecordMetaText(event, item: widget.item),
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

  List<String> _missingItemQuickPicks() {
    final List<String> profiled = <String>[
      for (final VbmappMaterialSuggestion material
          in widget.materialProfile?.recommendedMaterials ??
              const <VbmappMaterialSuggestion>[])
        if (material.type.contains('缺失') || material.name.trim().isNotEmpty)
          material.name,
      ...?widget.materialProfile?.quickPicks,
    ];
    return _deduplicatedTexts(<String>[
      ...profiled,
      ..._fallbackMissingItems,
    ]).take(12).toList(growable: false);
  }

  void _selectMissingItem(String item) {
    setState(() {
      _requestController.text = item;
    });
  }

  void _submit() {
    final String request = _requestController.text.trim();
    if (request.isEmpty) {
      return;
    }
    widget.onSubmitEvent(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: '缺失物品',
        environment: '缺失物品',
        targetKind: '缺失物',
        person: '',
        setting: '',
        example: '',
        responseMode: _promptChoice == '提问下' ? '提问下要求' : '自发要求',
        promptLevel: _promptChoice,
        functional: true,
      ),
    );
    _requestController.clear();
    setState(() {
      _selectedRecordIndex = null;
      _promptChoice = '提问下';
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
