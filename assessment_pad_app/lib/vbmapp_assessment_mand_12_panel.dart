part of 'vbmapp_assessment_page.dart';

class _VbmappMand1InlinePanel extends StatefulWidget {
  const _VbmappMand1InlinePanel({
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
  State<_VbmappMand1InlinePanel> createState() =>
      _VbmappMand1InlinePanelState();
}

class _VbmappMand1InlinePanelState extends State<_VbmappMand1InlinePanel> {
  static const List<String> _fallbackMaterials = <String>[
    '饼干',
    '书',
    '球',
    '泡泡',
    '音乐',
    '车',
    '秋千',
    '积木',
  ];

  final TextEditingController _requestController = TextEditingController();

  String _environment = '呈现物品';
  String _targetKind = '物品';
  String _promptChoice = '否';
  int? _selectedRecordIndex;

  bool get _usesExtraPromptRule => widget.item.itemCode == 'MAND_02M';

  int get _onePointRequestCount =>
      _scoreCountThreshold(widget.item, 1) ?? (_usesExtraPromptRule ? 4 : 2);

  int get _halfPointRequestCount =>
      _scoreCountThreshold(widget.item, .5) ?? (_usesExtraPromptRule ? 3 : 1);

  String get _recordTitle =>
      '${_vbmappQuestionDomainLabel(widget.item)}${widget.item.navCode}现场记录';

  String get _promptLabel => _usesExtraPromptRule ? '辅助' : '肢体辅助';

  List<String> get _promptValues => _usesExtraPromptRule
      ? const <String>['提问下', '自发地']
      : const <String>['否', '是'];

  String get _currentPromptChoice {
    final List<String> values = _promptValues;
    if (values.contains(_promptChoice)) {
      return _promptChoice;
    }
    return values.first;
  }

  String get _promptLevel {
    if (_usesExtraPromptRule) {
      return _currentPromptChoice;
    }
    return _currentPromptChoice == '是' ? '肢体辅助' : '无肢体辅助';
  }

  String get _requestHint => _usesExtraPromptRule ? '如：音乐、彩虹弹簧、球' : '如：饼干、书、打开';

  String get _scoreReference {
    final String promptRule =
        _usesExtraPromptRule ? '提问下仅限“你想要什么？”，自发地也计入有效要求。' : '肢体辅助不计入有效要求。';
    return '参考：0个计0分，$_halfPointRequestCount个计0.5分，'
        '$_onePointRequestCount个计1分；$promptRule';
  }

  @override
  void didUpdateWidget(covariant _VbmappMand1InlinePanel oldWidget) {
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
    final int qualifiedCount = _qualifiedMandCount(widget.events);
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
    final List<String> materials = _mandMaterialQuickPicks(
      widget.materialProfile,
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
                hintText: _requestHint,
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
    final Widget environmentChoice = Expanded(
      flex: 5,
      child: _VbmappMandInlineChoiceGroup(
        label: '环境',
        value: _environment,
        values: const <String>['呈现物品', '未呈现物品'],
        onChanged: (String value) => setState(() {
          _environment = value;
        }),
      ),
    );
    final Widget targetChoice = Expanded(
      flex: 4,
      child: _VbmappMandInlineChoiceGroup(
        label: '对象',
        value: _targetKind,
        values: const <String>['物品', '动作'],
        onChanged: (String value) => setState(() {
          _targetKind = value;
        }),
      ),
    );
    final Widget promptChoice = Expanded(
      flex: _usesExtraPromptRule ? 5 : 4,
      child: _VbmappMandInlineChoiceGroup(
        label: _promptLabel,
        value: _currentPromptChoice,
        values: _promptValues,
        onChanged: (String value) => setState(() {
          _promptChoice = value;
        }),
      ),
    );
    final List<Widget> choices = _usesExtraPromptRule
        ? <Widget>[promptChoice, environmentChoice, targetChoice]
        : <Widget>[environmentChoice, targetChoice, promptChoice];
    return Row(
      children: <Widget>[
        for (int index = 0; index < choices.length; index++) ...<Widget>[
          if (index > 0) const SizedBox(width: 10),
          choices[index],
        ],
      ],
    );
  }

  List<String> _mandMaterialQuickPicks(VbmappMaterialProfile? profile) {
    return _smartMandQuickPicks(
      profile,
      targetKind: _targetKind,
      fallback: _fallbackMaterials,
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
        responseMode: '要求',
        promptLevel: _promptLevel,
        functional: true,
      ),
    );
    _requestController.clear();
    setState(() {
      _promptChoice = _promptValues.first;
      _selectedRecordIndex = null;
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
