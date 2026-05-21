part of 'vbmapp_assessment_page.dart';

class _VbmappMand3InlinePanel extends StatefulWidget {
  const _VbmappMand3InlinePanel({
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
  State<_VbmappMand3InlinePanel> createState() =>
      _VbmappMand3InlinePanelState();
}

class _VbmappMand3InlinePanelState extends State<_VbmappMand3InlinePanel> {
  static const List<String> _fallbackMaterials = <String>[
    '泡泡',
    '音乐',
    '球',
    '彩虹弹簧',
    '秋千',
    '车',
    '饼干',
    '书',
  ];
  static const List<String> _fallbackPeople = <String>[
    '爸爸',
    '妈妈',
    '老师',
    '治疗师',
    '爷爷',
    '奶奶',
  ];
  static const List<String> _fallbackSettings = <String>[
    '屋里',
    '屋外',
    '教室',
    '游戏区',
    '桌面',
    '走廊',
  ];
  static const List<String> _fallbackExamples = <String>[
    '红瓶泡泡',
    '蓝瓶泡泡',
    '大球',
    '小球',
    '红车',
    '蓝车',
  ];

  final TextEditingController _requestController = TextEditingController();
  final TextEditingController _dimensionController = TextEditingController();

  String _dimension = '人物';
  String _promptChoice = '提问下';
  int? _selectedRecordIndex;

  String get _dimensionHint {
    switch (_dimension) {
      case '环境':
        return '如：屋里、屋外';
      case '例子':
        return '如：红瓶泡泡、蓝瓶泡泡';
      case '人物':
      default:
        return '如：爸爸、妈妈';
    }
  }

  @override
  void initState() {
    super.initState();
    _requestController.addListener(_handleRequestTextChanged);
  }

  @override
  void didUpdateWidget(covariant _VbmappMand3InlinePanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    final int? selectedIndex = _selectedRecordIndex;
    if (selectedIndex != null && selectedIndex >= widget.events.length) {
      _selectedRecordIndex = null;
    }
  }

  @override
  void dispose() {
    _requestController.removeListener(_handleRequestTextChanged);
    _requestController.dispose();
    _dimensionController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final Map<String, int> counts = _mandGeneralizationCounts(widget.events);
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
          final Widget summary = _buildGeneralizationSummary(counts);
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Icon(Icons.hub_outlined, color: widget.item.color, size: 19),
                  const SizedBox(width: 7),
                  const Expanded(
                    child: Text(
                      '提要求3M泛化记录',
                      style: TextStyle(
                        color: _VbmappColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  _VbmappEvidenceMetric(
                    label: '人物',
                    value: '${counts['people'] ?? 0}/2',
                    color: widget.item.color,
                  ),
                  const SizedBox(width: 8),
                  _VbmappEvidenceMetric(
                    label: '环境',
                    value: '${counts['settings'] ?? 0}/2',
                    color: widget.item.color,
                  ),
                  const SizedBox(width: 8),
                  _VbmappEvidenceMetric(
                    label: '例子',
                    value: '${counts['examples'] ?? 0}/2',
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
                    summary,
                  ],
                )
              else
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Expanded(flex: 5, child: form),
                    const SizedBox(width: 12),
                    Expanded(flex: 4, child: summary),
                  ],
                ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildRecordForm() {
    final List<String> materials = _mand3MaterialQuickPicks(
      widget.materialProfile,
    );
    final List<String> dimensionQuickPicks = _dimensionQuickPicks();
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        Row(
          children: <Widget>[
            Expanded(
              flex: 6,
              child: _VbmappMandInlineChoiceGroup(
                label: '本次记录',
                value: _dimension,
                values: const <String>['人物', '环境', '例子'],
                onChanged: (String value) => setState(() {
                  _dimension = value;
                  _dimensionController.clear();
                }),
              ),
            ),
            const SizedBox(width: 10),
            Expanded(
              flex: 4,
              child: _VbmappMandInlineChoiceGroup(
                label: '辅助',
                value: _promptChoice,
                values: const <String>['提问下', '自发地'],
                onChanged: (String value) => setState(() {
                  _promptChoice = value;
                }),
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            Expanded(
              child: _VbmappMandInlineTextField(
                controller: _dimensionController,
                label: _dimension,
                hintText: _dimensionHint,
              ),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: _VbmappMandInlineTextField(
                controller: _requestController,
                label: '要求内容',
                hintText: '如：泡泡、音乐、球',
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
        _buildQuickPickBar(
          dimensionValues: dimensionQuickPicks,
          materialValues: materials,
        ),
      ],
    );
  }

  Widget _buildQuickPickBar({
    required List<String> dimensionValues,
    required List<String> materialValues,
  }) {
    if (dimensionValues.isEmpty && materialValues.isEmpty) {
      return const SizedBox.shrink();
    }
    return SizedBox(
      height: 30,
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: <Widget>[
            if (dimensionValues.isNotEmpty) ...<Widget>[
              _VbmappMandQuickPickLabel(label: _dimension),
              const SizedBox(width: 5),
              for (final String value in dimensionValues) ...<Widget>[
                _VbmappMandCompactChip(
                  label: value,
                  selected: _dimensionController.text.trim() == value,
                  onTap: () => _selectDimensionValue(value),
                ),
                const SizedBox(width: 5),
              ],
            ],
            if (dimensionValues.isNotEmpty &&
                materialValues.isNotEmpty) ...<Widget>[
              Container(
                width: 1,
                height: 18,
                margin: const EdgeInsets.symmetric(horizontal: 4),
                color: _VbmappColors.lineSoft,
              ),
              const SizedBox(width: 2),
            ],
            if (materialValues.isNotEmpty) ...<Widget>[
              const _VbmappMandQuickPickLabel(label: '要求'),
              const SizedBox(width: 5),
              for (final String material in materialValues) ...<Widget>[
                _VbmappMandCompactChip(
                  label: material,
                  selected: _requestController.text.trim() == material,
                  onTap: () => _selectMaterial(material),
                ),
                const SizedBox(width: 5),
              ],
            ],
          ],
        ),
      ),
    );
  }

  List<String> _mand3MaterialQuickPicks(VbmappMaterialProfile? profile) {
    final List<String> configured =
        profile?.quickPickLabels ?? const <String>[];
    if (configured.isEmpty) {
      return _fallbackMaterials;
    }
    final List<String> preferred = <String>[
      '泡泡',
      '音乐',
      '球',
      '彩虹弹簧',
      '秋千',
      ...configured,
    ];
    final Set<String> seen = <String>{};
    return preferred
        .where((String value) {
          final String normalized = value.trim();
          if (normalized.isEmpty || seen.contains(normalized)) {
            return false;
          }
          seen.add(normalized);
          return true;
        })
        .take(8)
        .toList(growable: false);
  }

  List<String> _dimensionQuickPicks() {
    switch (_dimension) {
      case '环境':
        return _configuredDimensionQuickPicks(
          'mand3_settings',
          _fallbackSettings,
        );
      case '例子':
        final String key = _exampleQuickPickFieldKey();
        final List<String> configured = _configuredDimensionQuickPicks(
          key,
          const <String>[],
        );
        if (configured.isNotEmpty) {
          return configured;
        }
        return _configuredDimensionQuickPicks(
          'mand3_examples_default',
          _fallbackExamples,
        );
      case '人物':
      default:
        return _configuredDimensionQuickPicks('mand3_people', _fallbackPeople);
    }
  }

  List<String> _configuredDimensionQuickPicks(
    String fieldKey,
    List<String> fallback,
  ) {
    final List<String> configured =
        widget.materialProfile?.quickPicksFor(fieldKey) ?? const <String>[];
    final List<String> source = configured.isEmpty ? fallback : configured;
    return source.take(6).toList(growable: false);
  }

  String _exampleQuickPickFieldKey() {
    final String request = _requestController.text.trim();
    if (request.contains('泡泡')) {
      return 'mand3_examples_bubbles';
    }
    if (request.contains('球')) {
      return 'mand3_examples_ball';
    }
    if (request.contains('车')) {
      return 'mand3_examples_car';
    }
    if (request.contains('音乐') || request.contains('歌')) {
      return 'mand3_examples_music';
    }
    if (request.contains('彩虹') || request.contains('弹簧')) {
      return 'mand3_examples_slinky';
    }
    if (request.contains('秋千')) {
      return 'mand3_examples_swing';
    }
    return 'mand3_examples_default';
  }

  Widget _buildGeneralizationSummary(Map<String, int> counts) {
    final Map<String, List<String>> values =
        _mandGeneralizationValues(widget.events);
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
          '达标进度',
          style: TextStyle(
            color: _VbmappColors.ink,
            fontSize: 14,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 8),
        Row(
          children: <Widget>[
            Expanded(
              child: _VbmappMand3CoverageColumn(
                label: '人物',
                values: values['people'] ?? const <String>[],
                color: widget.item.color,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: _VbmappMand3CoverageColumn(
                label: '环境',
                values: values['settings'] ?? const <String>[],
                color: widget.item.color,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: _VbmappMand3CoverageColumn(
                label: '例子',
                values: values['examples'] ?? const <String>[],
                color: widget.item.color,
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        _VbmappMand3RecordList(
          events: widget.events,
          selectedIndex: _selectedRecordIndex,
          onSelectIndex: _selectRecord,
          onDeleteIndex: _deleteRecord,
        ),
        const SizedBox(height: 10),
        Text(
          '参考：资料要求同一强化物覆盖2个人、2个环境、2个不同例子；各1个计0.5分，各2个计1分。',
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

  void _selectDimensionValue(String value) {
    setState(() {
      _dimensionController.text = value;
    });
  }

  void _handleRequestTextChanged() {
    if (mounted) {
      setState(() {});
    }
  }

  void _submit() {
    final String request = _requestController.text.trim();
    final String dimensionValue = _dimensionController.text.trim();
    if (request.isEmpty || dimensionValue.isEmpty) {
      return;
    }
    widget.onSubmitEvent(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: '',
        person: _dimension == '人物' ? dimensionValue : '',
        setting: _dimension == '环境' ? dimensionValue : '',
        example: _dimension == '例子' ? dimensionValue : '',
        responseMode: '要求',
        promptLevel: _promptChoice,
        functional: true,
      ),
    );
    _requestController.clear();
    _dimensionController.clear();
    setState(() {
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
