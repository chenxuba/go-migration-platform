part of 'vbmapp_assessment_page.dart';

class _VbmappTimedMandQuickRecordTarget {
  const _VbmappTimedMandQuickRecordTarget({
    required this.item,
    required this.responseSchema,
    required this.materialProfile,
    required this.recordCount,
    required this.qualifiedCount,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? responseSchema;
  final VbmappMaterialProfile? materialProfile;
  final int recordCount;
  final int qualifiedCount;
}

class _VbmappTimedMandQuickRecordResult {
  const _VbmappTimedMandQuickRecordResult({
    required this.item,
    required this.event,
  });

  final _VbmappItem item;
  final _VbmappMandEvent event;
}

class _VbmappMand4QuickRecordDialog extends StatefulWidget {
  const _VbmappMand4QuickRecordDialog({
    required this.targets,
    required this.initialItemCode,
  });

  final List<_VbmappTimedMandQuickRecordTarget> targets;
  final String initialItemCode;

  @override
  State<_VbmappMand4QuickRecordDialog> createState() =>
      _VbmappMand4QuickRecordDialogState();
}

class _VbmappMand4QuickRecordDialogState
    extends State<_VbmappMand4QuickRecordDialog> {
  static const List<String> _fallbackMaterials = <String>[
    '泡泡',
    '球',
    '音乐',
    '彩虹弹簧',
    '出去',
    '打开',
    '秋千',
    '车',
  ];

  final TextEditingController _requestController = TextEditingController();
  String _promptMode = '自发地';
  String _presentation = '呈现物品';
  String _targetKind = '物品';
  String _ability = '';
  String _selectedItemCode = '';

  @override
  void initState() {
    super.initState();
    _selectedItemCode = widget.targets.any(
      (_VbmappTimedMandQuickRecordTarget target) =>
          target.item.itemCode == widget.initialItemCode,
    )
        ? widget.initialItemCode
        : (widget.targets.isEmpty ? '' : widget.targets.first.item.itemCode);
    if (widget.targets.isNotEmpty) {
      _resetChoicesForTarget(_selectedTarget);
    }
  }

  _VbmappTimedMandQuickRecordTarget get _selectedTarget {
    for (final _VbmappTimedMandQuickRecordTarget target in widget.targets) {
      if (target.item.itemCode == _selectedItemCode) {
        return target;
      }
    }
    return widget.targets.first;
  }

  _VbmappTimedMandStrategy _strategyFor(
    _VbmappTimedMandQuickRecordTarget target,
  ) {
    return _timedMandStrategyForItem(
          target.item,
          responseSchema: target.responseSchema,
        ) ??
        _vbmappTimedMandStrategies['MAND_04M']!;
  }

  void _resetChoicesForTarget(_VbmappTimedMandQuickRecordTarget target) {
    final _VbmappTimedMandStrategy strategy = _strategyFor(target);
    _promptMode = strategy.defaultPromptMode;
    _presentation = strategy.defaultPresentation;
    _targetKind = strategy.defaultTargetKind;
    _ability = strategy.defaultAbility;
  }

  @override
  void dispose() {
    _requestController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (widget.targets.isEmpty) {
      return const SizedBox.shrink();
    }
    final _VbmappTimedMandQuickRecordTarget selectedTarget = _selectedTarget;
    final _VbmappTimedMandStrategy strategy = _strategyFor(selectedTarget);
    final List<String> materials = _smartMandQuickPicks(
      selectedTarget.materialProfile,
      targetKind: _targetKind,
      fallback: strategy.quickPickFallback.isEmpty
          ? _fallbackMaterials
          : strategy.quickPickFallback,
    );
    return Center(
      child: Material(
        color: Colors.transparent,
        child: Container(
          width: 760,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: _VbmappColors.line),
            boxShadow: _vbmappShadow(blur: 24),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              const Text(
                '补记一条提要求观察记录',
                style: TextStyle(
                  color: _VbmappColors.ink,
                  fontSize: 18,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 14),
              _buildTargetPicker(),
              const SizedBox(height: 12),
              _buildChoiceRow(strategy),
              const SizedBox(height: 12),
              _VbmappMandInlineTextField(
                controller: _requestController,
                label: strategy.inputLabel,
                hintText: strategy.inputHint,
              ),
              const SizedBox(height: 10),
              Wrap(
                spacing: 8,
                runSpacing: 6,
                children: <Widget>[
                  for (final String material in materials)
                    _VbmappMandMaterialChip(
                      label: material,
                      selected: _requestController.text.trim() == material,
                      onTap: () => setState(() {
                        _requestController.text = material;
                      }),
                    ),
                ],
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: <Widget>[
                  _VbmappMand4TimerButton(
                    icon: Icons.close_rounded,
                    label: '取消',
                    onTap: () => Navigator.of(context).pop(),
                  ),
                  const SizedBox(width: 8),
                  _VbmappMand4TimerButton(
                    key: const ValueKey<String>('vbmapp-global-mand4-submit'),
                    icon: Icons.check_rounded,
                    label: '保存记录',
                    filled: true,
                    onTap: _submit,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildTargetPicker() {
    return Row(
      children: <Widget>[
        const Text(
          '记录到',
          style: TextStyle(
            color: _VbmappColors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(width: 10),
        for (int index = 0; index < widget.targets.length; index++) ...<Widget>[
          if (index > 0) const SizedBox(width: 8),
          Expanded(
            child: _VbmappTimedMandTargetButton(
              target: widget.targets[index],
              selected:
                  widget.targets[index].item.itemCode == _selectedItemCode,
              onTap: () => _selectTarget(widget.targets[index]),
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildChoiceRow(_VbmappTimedMandStrategy strategy) {
    final List<_VbmappQuickMandChoiceSpec> choices =
        <_VbmappQuickMandChoiceSpec>[
      if (strategy.showPromptSelector)
        _VbmappQuickMandChoiceSpec(
          flex: 4,
          label: strategy.promptSelectorLabel,
          value: _currentPromptMode(strategy),
          values: strategy.promptOptions,
          onChanged: (String value) => setState(() {
            _promptMode = value;
          }),
        ),
      _VbmappQuickMandChoiceSpec(
        flex: 5,
        label: strategy.presentationSelectorLabel,
        value: _presentation,
        values: const <String>['呈现物品', '未呈现物品'],
        onChanged: (String value) => setState(() {
          _presentation = value;
        }),
      ),
      if (strategy.targetOptions.isNotEmpty)
        _VbmappQuickMandChoiceSpec(
          flex: 5,
          label: '目标',
          value: _targetKind,
          values: strategy.targetOptions,
          onChanged: (String value) => setState(() {
            _targetKind = value;
          }),
        ),
      if (strategy.abilityOptions.isNotEmpty)
        _VbmappQuickMandChoiceSpec(
          flex: 5,
          label: strategy.abilitySelectorLabel,
          value: _currentAbility(strategy),
          values: strategy.abilityOptions,
          onChanged: (String value) => setState(() {
            _ability = value;
          }),
        ),
    ];
    if (choices.length == 1) {
      final _VbmappQuickMandChoiceSpec choice = choices.first;
      return Align(
        alignment: Alignment.centerLeft,
        child: SizedBox(width: 300, child: _buildChoice(choice)),
      );
    }
    return Row(
      children: <Widget>[
        for (int index = 0; index < choices.length; index++) ...<Widget>[
          if (index > 0) const SizedBox(width: 12),
          Expanded(
              flex: choices[index].flex, child: _buildChoice(choices[index])),
        ],
      ],
    );
  }

  String _currentPromptMode(_VbmappTimedMandStrategy strategy) {
    if (strategy.promptOptions.contains(_promptMode)) {
      return _promptMode;
    }
    return strategy.defaultPromptMode;
  }

  String _currentAbility(_VbmappTimedMandStrategy strategy) {
    if (strategy.abilityOptions.contains(_ability)) {
      return _ability;
    }
    return strategy.defaultAbility;
  }

  Widget _buildChoice(_VbmappQuickMandChoiceSpec choice) {
    return _VbmappMandInlineChoiceGroup(
      label: choice.label,
      value: choice.value,
      values: choice.values,
      onChanged: choice.onChanged,
    );
  }

  void _selectTarget(_VbmappTimedMandQuickRecordTarget target) {
    setState(() {
      _selectedItemCode = target.item.itemCode;
      _resetChoicesForTarget(target);
    });
  }

  void _submit() {
    final String request = _requestController.text.trim();
    if (request.isEmpty) {
      return;
    }
    final _VbmappTimedMandQuickRecordTarget target = _selectedTarget;
    final _VbmappTimedMandStrategy strategy = _strategyFor(target);
    final _VbmappMandPhraseAssessment phraseAssessment = _assessMandPhrase(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: '',
        person: '',
        setting: '',
        example: '',
        responseMode: strategy.responseModeForPrompt(
          _currentPromptMode(strategy),
        ),
        promptLevel: strategy.promptLevelForPrompt(
          _currentPromptMode(strategy),
        ),
        functional: true,
      ),
    );
    Navigator.of(context).pop(
      _VbmappTimedMandQuickRecordResult(
        item: target.item,
        event: _VbmappMandEvent(
          utterance: request,
          target: request,
          motivationContext: '',
          environment: _presentation,
          targetKind: strategy.targetOptions.isEmpty ? '' : _targetKind,
          person: '',
          setting: '',
          example: '',
          responseMode: strategy.responseModeForPrompt(
            _currentPromptMode(strategy),
          ),
          promptLevel: strategy.promptLevelForPrompt(
            _currentPromptMode(strategy),
          ),
          phraseLevel: strategy.abilityOptions.isNotEmpty
              ? _currentAbility(strategy)
              : strategy.multiWordQualifiedMinCount > 0
                  ? phraseAssessment.label
                  : '',
          functional: true,
        ),
      ),
    );
  }
}

class _VbmappQuickMandChoiceSpec {
  const _VbmappQuickMandChoiceSpec({
    required this.flex,
    required this.label,
    required this.value,
    required this.values,
    required this.onChanged,
  });

  final int flex;
  final String label;
  final String value;
  final List<String> values;
  final ValueChanged<String> onChanged;
}

class _VbmappTimedMandTargetButton extends StatelessWidget {
  const _VbmappTimedMandTargetButton({
    required this.target,
    required this.selected,
    required this.onTap,
  });

  final _VbmappTimedMandQuickRecordTarget target;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color color = target.item.color;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 48,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color: selected ? color.withOpacity(.12) : const Color(0xFFFFFCFA),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: selected ? color.withOpacity(.55) : _VbmappColors.lineSoft,
            ),
          ),
          child: Row(
            children: <Widget>[
              Text(
                target.item.navCode,
                style: TextStyle(
                  color: selected ? _VbmappColors.ink : _VbmappColors.body,
                  fontSize: 15,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  _quickTargetLabel(target.item.itemCode),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _VbmappColors.body,
                    fontSize: 11.5,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
              const SizedBox(width: 6),
              Text(
                '${target.qualifiedCount}/${target.recordCount}',
                style: TextStyle(
                  color: selected ? color : _VbmappColors.muted,
                  fontSize: 11.5,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  static String _quickTargetLabel(String itemCode) {
    switch (itemCode) {
      case 'MAND_08M':
        return '双词不同要求';
      case 'MAND_09M':
        return '30分钟自发不同';
      case 'MAND_11M':
        return '信息要求';
      case 'MAND_13M':
        return '形容介副词要求';
      case 'MAND_04M':
      default:
        return '60分钟自发要求';
    }
  }
}

class _VbmappObservationFinishConfirmDialog extends StatelessWidget {
  const _VbmappObservationFinishConfirmDialog({
    this.title = '确认结束观察？',
    this.message = '结束后会保留当前计时和记录。如只是暂时离开，建议点暂停。',
  });

  final String title;
  final String message;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Material(
        color: Colors.transparent,
        child: Container(
          width: 480,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: _VbmappColors.line),
            boxShadow: _vbmappShadow(blur: 24),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Text(
                title,
                style: const TextStyle(
                  color: _VbmappColors.ink,
                  fontSize: 18,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 10),
              Text(
                message,
                style: const TextStyle(
                  color: _VbmappColors.body,
                  fontSize: 13,
                  height: 1.45,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: <Widget>[
                  _VbmappMand4TimerButton(
                    icon: Icons.close_rounded,
                    label: '取消',
                    onTap: () => Navigator.of(context).pop(false),
                  ),
                  const SizedBox(width: 8),
                  _VbmappMand4TimerButton(
                    key: const ValueKey<String>(
                        'vbmapp-confirm-finish-observation'),
                    icon: Icons.stop_rounded,
                    label: '确认结束',
                    filled: true,
                    onTap: () => Navigator.of(context).pop(true),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappMandEventDialog extends StatefulWidget {
  const _VbmappMandEventDialog({required this.generalizationMode});

  final bool generalizationMode;

  @override
  State<_VbmappMandEventDialog> createState() => _VbmappMandEventDialogState();
}

class _VbmappMandEventDialogState extends State<_VbmappMandEventDialog> {
  final TextEditingController _utteranceController = TextEditingController();
  final TextEditingController _targetController = TextEditingController();
  final TextEditingController _contextController = TextEditingController();
  final TextEditingController _personController = TextEditingController();
  final TextEditingController _settingController = TextEditingController();
  final TextEditingController _exampleController = TextEditingController();

  String _responseMode = '口语';
  String _promptLevel = '无辅助';
  bool _functional = true;

  @override
  void dispose() {
    _utteranceController.dispose();
    _targetController.dispose();
    _contextController.dispose();
    _personController.dispose();
    _settingController.dispose();
    _exampleController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Material(
        color: Colors.transparent,
        child: Container(
          width: widget.generalizationMode ? 720 : 560,
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: _VbmappColors.line),
            boxShadow: _vbmappShadow(blur: 24),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Text(
                widget.generalizationMode ? '记录一次泛化要求' : '记录一次提要求',
                style: TextStyle(
                  color: _VbmappColors.ink,
                  fontSize: 22,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 14),
              _VbmappDialogTextField(
                controller: _utteranceController,
                label: '孩子发出的词语 / 手语 / 图片交换',
                hintText: '例如：饼干、球、打开',
              ),
              const SizedBox(height: 10),
              _VbmappDialogTextField(
                controller: _targetController,
                label: '要求的物品或活动',
                hintText: '例如：饼干、泡泡、秋千',
              ),
              const SizedBox(height: 10),
              _VbmappDialogTextField(
                controller: _contextController,
                label: '动机情境',
                hintText: '例如：看到饼干但拿不到',
              ),
              if (widget.generalizationMode) ...<Widget>[
                const SizedBox(height: 10),
                Row(
                  children: <Widget>[
                    Expanded(
                      child: _VbmappDialogTextField(
                        controller: _personController,
                        label: '互动对象',
                        hintText: '例如：妈妈、老师',
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: _VbmappDialogTextField(
                        controller: _settingController,
                        label: '环境',
                        hintText: '例如：教室、户外',
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: _VbmappDialogTextField(
                        controller: _exampleController,
                        label: '不同例子',
                        hintText: '例如：红泡泡、蓝泡泡',
                      ),
                    ),
                  ],
                ),
              ],
              const SizedBox(height: 12),
              Row(
                children: <Widget>[
                  Expanded(
                    child: _VbmappDialogDropdown(
                      label: '沟通形式',
                      value: _responseMode,
                      values: const <String>['口语', '手语', '图片交换', '手势', '其他'],
                      onChanged: (String value) {
                        setState(() => _responseMode = value);
                      },
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: _VbmappDialogDropdown(
                      label: '辅助水平',
                      value: _promptLevel,
                      values: const <String>[
                        '无辅助',
                        '口头提示',
                        '仿说/模仿',
                        '其他辅助',
                        '肢体辅助',
                      ],
                      onChanged: (String value) {
                        setState(() => _promptLevel = value);
                      },
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              CheckboxListTile(
                contentPadding: EdgeInsets.zero,
                value: _functional,
                activeColor: _VbmappColors.orange,
                onChanged: (bool? value) {
                  setState(() => _functional = value ?? true);
                },
                title: const Text(
                  '本次反应是功能性要求，并获得或指向目标物',
                  style: TextStyle(
                    color: _VbmappColors.body,
                    fontSize: 14,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
              const SizedBox(height: 14),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: <Widget>[
                  TextButton(
                    onPressed: () => Navigator.of(context).pop(),
                    child: const Text('取消'),
                  ),
                  const SizedBox(width: 10),
                  ElevatedButton(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: _VbmappColors.orange,
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 18,
                        vertical: 12,
                      ),
                    ),
                    onPressed: _submit,
                    child: const Text('保存事件'),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _submit() {
    final String utterance = _utteranceController.text.trim();
    final String target = _targetController.text.trim();
    if (utterance.isEmpty && target.isEmpty) {
      return;
    }
    Navigator.of(context).pop(
      _VbmappMandEvent(
        utterance: utterance,
        target: target,
        motivationContext: _contextController.text.trim(),
        person: _personController.text.trim(),
        setting: _settingController.text.trim(),
        example: _exampleController.text.trim(),
        responseMode: _responseMode,
        promptLevel: _promptLevel,
        functional: _functional,
      ),
    );
  }
}

class _VbmappDialogTextField extends StatelessWidget {
  const _VbmappDialogTextField({
    required this.controller,
    required this.label,
    required this.hintText,
  });

  final TextEditingController controller;
  final String label;
  final String hintText;

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      decoration: InputDecoration(
        labelText: label,
        hintText: hintText,
        filled: true,
        fillColor: const Color(0xFFFFFAF5),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: _VbmappColors.lineSoft),
        ),
      ),
    );
  }
}

class _VbmappDialogDropdown extends StatelessWidget {
  const _VbmappDialogDropdown({
    required this.label,
    required this.value,
    required this.values,
    required this.onChanged,
  });

  final String label;
  final String value;
  final List<String> values;
  final ValueChanged<String> onChanged;

  @override
  Widget build(BuildContext context) {
    return DropdownButtonFormField<String>(
      value: value,
      decoration: InputDecoration(
        labelText: label,
        filled: true,
        fillColor: const Color(0xFFFFFAF5),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: _VbmappColors.lineSoft),
        ),
      ),
      items: <DropdownMenuItem<String>>[
        for (final String option in values)
          DropdownMenuItem<String>(
            value: option,
            child: Text(option),
          ),
      ],
      onChanged: (String? value) {
        if (value != null) {
          onChanged(value);
        }
      },
    );
  }
}
