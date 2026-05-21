part of 'vbmapp_assessment_page.dart';

class _VbmappMand5InlinePanel extends StatefulWidget {
  const _VbmappMand5InlinePanel({
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
  State<_VbmappMand5InlinePanel> createState() =>
      _VbmappMand5InlinePanelState();
}

class _VbmappMand5InlinePanelState extends State<_VbmappMand5InlinePanel> {
  static const List<String> _fallbackMaterials = <String>[
    '泡泡',
    '球',
    '音乐',
    '彩虹弹簧',
    '秋千',
    '车',
    '饼干',
    '打开',
  ];

  final TextEditingController _requestController = TextEditingController();

  String _environment = '呈现物品';
  String _targetKind = '物品';
  String _promptChoice = '自发地';
  int? _selectedRecordIndex;

  String get _currentPromptChoice {
    if (_promptChoice == '提问下') {
      return '提问下';
    }
    return '自发地';
  }

  int get _onePointRequestCount => _scoreCountThreshold(widget.item, 1) ?? 10;

  int get _halfPointRequestCount => _scoreCountThreshold(widget.item, .5) ?? 8;

  String get _recordTitle => '提要求${widget.item.navCode}现场记录';

  String get _scoreReference {
    return '参考：0个计0分，$_halfPointRequestCount个计0.5分，'
        '$_onePointRequestCount个计1分；仅统计呈现物品条件下的自发不同要求，提问下记录保留但不计入有效要求。';
  }

  @override
  void didUpdateWidget(covariant _VbmappMand5InlinePanel oldWidget) {
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
    final double suggestedScore = _suggestMandScore(widget.item, widget.events);
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
                  Icon(Icons.record_voice_over_outlined,
                      color: widget.item.color, size: 19),
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
    final List<String> materials = _smartMandQuickPicks(
      widget.materialProfile,
      targetKind: _targetKind,
      fallback: _fallbackMaterials,
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        _buildChoiceRow(),
        const SizedBox(height: 12),
        Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            Expanded(
              child: _VbmappMandInlineTextField(
                controller: _requestController,
                label: '孩子要求内容',
                hintText: '如：泡泡、秋千、打开、出去',
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
            for (final String material in materials)
              _VbmappMandMaterialChip(
                label: material,
                selected: _requestController.text.trim() == material,
                onTap: () => _selectMaterial(material),
              ),
          ],
        ),
      ],
    );
  }

  Widget _buildChoiceRow() {
    return Row(
      children: <Widget>[
        Expanded(
          flex: 5,
          child: _VbmappMandInlineChoiceGroup(
            label: '环境',
            value: _environment,
            values: const <String>['呈现物品', '未呈现物品'],
            onChanged: (String value) => setState(() {
              _environment = value;
            }),
          ),
        ),
        const SizedBox(width: 10),
        Expanded(
          flex: 5,
          child: _VbmappMandInlineChoiceGroup(
            label: '目标',
            value: _targetKind,
            values: const <String>['物品', '动作', '活动'],
            onChanged: (String value) => setState(() {
              _targetKind = value;
            }),
          ),
        ),
        const SizedBox(width: 10),
        Expanded(
          flex: 4,
          child: _VbmappMandInlineChoiceGroup(
            label: '辅助',
            value: _currentPromptChoice,
            values: const <String>['提问下', '自发地'],
            onChanged: (String value) => setState(() {
              _promptChoice = value;
            }),
          ),
        ),
      ],
    );
  }

  Widget _buildRecordSummary() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        if (widget.materialProfile != null) ...<Widget>[
          _VbmappInlineInfo(
            icon: Icons.inventory_2_outlined,
            text: widget.materialProfile!.label,
          ),
          const SizedBox(height: 10),
        ],
        const Text(
          '有效要求记录',
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

  void _selectMaterial(String material) {
    setState(() {
      _requestController.text = material;
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
        motivationContext: '',
        environment: _environment,
        targetKind: _targetKind,
        person: '',
        setting: '',
        example: '',
        responseMode: _currentPromptChoice == '提问下' ? '提问下要求' : '自发要求',
        promptLevel: _currentPromptChoice,
        functional: true,
      ),
    );
    _requestController.clear();
    setState(() {
      _selectedRecordIndex = null;
      _promptChoice = '自发地';
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
