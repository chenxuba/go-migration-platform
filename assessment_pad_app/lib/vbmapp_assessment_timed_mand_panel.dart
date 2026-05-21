part of 'vbmapp_assessment_page.dart';

class _VbmappTimedMandInlinePanel extends StatefulWidget {
  const _VbmappTimedMandInlinePanel({
    required this.item,
    required this.responseSchema,
    required this.materialProfile,
    required this.events,
    required this.observation,
    required this.onSubmitEvent,
    required this.onDeleteEvent,
    required this.onChangeObservation,
  });

  final _VbmappItem item;
  final VbmappItemResponseSchema? responseSchema;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> events;
  final _VbmappObservationTimerState? observation;
  final ValueChanged<_VbmappMandEvent> onSubmitEvent;
  final ValueChanged<int> onDeleteEvent;
  final ValueChanged<_VbmappObservationTimerState> onChangeObservation;

  @override
  State<_VbmappTimedMandInlinePanel> createState() =>
      _VbmappTimedMandInlinePanelState();
}

class _VbmappTimedMandInlinePanelState
    extends State<_VbmappTimedMandInlinePanel> {
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

  Timer? _clockTimer;
  bool _autoFinishRequested = false;
  String _presentation = '呈现物品';
  String _targetKind = '物品';
  String _promptMode = '自发地';
  String _ability = '';
  int? _selectedRecordIndex;

  _VbmappTimedMandStrategy get _strategy =>
      _timedMandStrategyForItem(
        widget.item,
        responseSchema: widget.responseSchema,
      ) ??
      _vbmappTimedMandStrategies['MAND_04M']!;

  _VbmappObservationTimerState get _observation =>
      (widget.observation ?? const _VbmappObservationTimerState())
          .withPlannedMinutes(_strategy.plannedMinutes);

  int get _onePointRequestCount => _strategy.onePointCount(widget.item);

  int get _halfPointRequestCount => _strategy.halfPointCount(widget.item);

  int get _plannedMinutes => _strategy.plannedMinutes;

  int get _plannedSeconds => _plannedMinutes * Duration.secondsPerMinute;

  int get _elapsedSeconds => _observation.elapsedSecondsAt(DateTime.now());

  int get _remainingSeconds {
    final int remaining = _plannedSeconds - _elapsedSeconds;
    return remaining > 0 ? remaining : 0;
  }

  bool get _observationMet => _elapsedSeconds >= _plannedSeconds;

  int get _displayMinSlots => _strategy.resolvedDisplayMinSlots(widget.item);

  String get _currentPromptMode {
    if (_strategy.promptOptions.contains(_promptMode)) {
      return _promptMode;
    }
    return _strategy.defaultPromptMode;
  }

  String get _recordTitle => '提要求${widget.item.navCode}观察记录';

  String get _scoreReference => _strategy.scoreReference(widget.item);

  String get _observationHint => _strategy.observationHint(
        item: widget.item,
        events: widget.events,
        observation: widget.observation,
        responseSchema: widget.responseSchema,
      );

  @override
  void initState() {
    super.initState();
    _presentation = _strategy.defaultPresentation;
    _targetKind = _strategy.defaultTargetKind;
    _promptMode = _strategy.defaultPromptMode;
    _ability = _strategy.defaultAbility;
    _syncClockTicker();
  }

  @override
  void didUpdateWidget(covariant _VbmappTimedMandInlinePanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    final int? selectedIndex = _selectedRecordIndex;
    if (selectedIndex != null && selectedIndex >= widget.events.length) {
      _selectedRecordIndex = null;
    }
    if (oldWidget.item.itemCode != widget.item.itemCode) {
      _presentation = _strategy.defaultPresentation;
      _targetKind = _strategy.defaultTargetKind;
      _promptMode = _strategy.defaultPromptMode;
      _ability = _strategy.defaultAbility;
    }
    if (oldWidget.observation != widget.observation) {
      _autoFinishRequested = false;
      _syncClockTicker();
    }
  }

  @override
  void dispose() {
    _clockTimer?.cancel();
    _requestController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final int qualifiedCount = _qualifiedMandCountForItem(
      widget.item,
      widget.events,
      observation: widget.observation,
      responseSchema: widget.responseSchema,
    );
    final double suggestedScore = _suggestMandScore(
      widget.item,
      widget.events,
      observation: widget.observation,
      responseSchema: widget.responseSchema,
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
          final bool narrow = constraints.maxWidth < 860;
          final Widget form = _buildRecordForm();
          final Widget summary = _buildRecordSummary();
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Icon(Icons.timer_outlined,
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
                    label: '观察',
                    value: _vbmappDurationText(_elapsedSeconds),
                    color: widget.item.color,
                  ),
                  const SizedBox(width: 8),
                  _VbmappEvidenceMetric(
                    label: _strategy.countMetricLabel,
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
              const SizedBox(height: 10),
              _buildObservationBar(),
              const SizedBox(height: 10),
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

  Widget _buildObservationBar() {
    final bool running = _observation.isRunning;
    final bool ended = _observation.ended;
    final String primaryActionLabel = !_observation.hasStarted
        ? '开始观察'
        : ended
            ? '重新观察'
            : running
                ? '暂停'
                : '继续观察';
    final IconData primaryActionIcon = !_observation.hasStarted
        ? Icons.play_arrow_rounded
        : ended
            ? Icons.replay_rounded
            : running
                ? Icons.pause_rounded
                : Icons.play_arrow_rounded;
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(11),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Row(
            children: <Widget>[
              Expanded(
                child: Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: <Widget>[
                    _VbmappMand4TimerMetricChip(
                      label: '观察窗',
                      value: '${_plannedMinutes}分钟',
                      tone: widget.item.color,
                    ),
                    _VbmappMand4TimerMetricChip(
                      label: '已观察',
                      value: _vbmappDurationText(_elapsedSeconds),
                      tone: widget.item.color,
                    ),
                    _VbmappMand4TimerMetricChip(
                      label: '剩余',
                      value: _vbmappDurationText(_remainingSeconds),
                      tone: _remainingSeconds == 0
                          ? _VbmappColors.green
                          : _VbmappColors.orangeDeep,
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 10),
              Row(
                mainAxisSize: MainAxisSize.min,
                children: <Widget>[
                  _VbmappMand4TimerButton(
                    key: const ValueKey<String>('vbmapp-mand4-primary-timer'),
                    icon: primaryActionIcon,
                    label: primaryActionLabel,
                    filled: true,
                    onTap: _handlePrimaryTimerAction,
                  ),
                  const SizedBox(width: 8),
                  _VbmappMand4TimerButton(
                    key: const ValueKey<String>('vbmapp-mand4-stop-timer'),
                    icon: Icons.stop_rounded,
                    label: '结束',
                    onTap: _observation.hasStarted ? _finishObservation : null,
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            _observationHint,
            style: const TextStyle(
              color: _VbmappColors.body,
              fontSize: 12,
              height: 1.35,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRecordForm() {
    final List<String> materials =
        _mandMaterialQuickPicks(widget.materialProfile);
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
                label: _strategy.inputLabel,
                hintText: _strategy.inputHint,
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
    final List<_VbmappTimedMandChoiceSpec> choices =
        <_VbmappTimedMandChoiceSpec>[
      if (_strategy.showPromptSelector)
        _VbmappTimedMandChoiceSpec(
          flex: 4,
          label: _strategy.promptSelectorLabel,
          value: _currentPromptMode,
          values: _strategy.promptOptions,
          onChanged: (String value) => setState(() {
            _promptMode = value;
          }),
        ),
      _VbmappTimedMandChoiceSpec(
        flex: 5,
        label: _strategy.presentationSelectorLabel,
        value: _presentation,
        values: const <String>['呈现物品', '未呈现物品'],
        onChanged: (String value) => setState(() {
          _presentation = value;
        }),
      ),
      if (_strategy.targetOptions.isNotEmpty)
        _VbmappTimedMandChoiceSpec(
          flex: 5,
          label: '目标',
          value: _targetKind,
          values: _strategy.targetOptions,
          onChanged: (String value) => setState(() {
            _targetKind = value;
          }),
        ),
      if (_strategy.abilityOptions.isNotEmpty)
        _VbmappTimedMandChoiceSpec(
          flex: 5,
          label: _strategy.abilitySelectorLabel,
          value: _currentAbility,
          values: _strategy.abilityOptions,
          onChanged: (String value) => setState(() {
            _ability = value;
          }),
        ),
    ];
    if (choices.length == 1) {
      final _VbmappTimedMandChoiceSpec choice = choices.first;
      return Align(
        alignment: Alignment.centerLeft,
        child: SizedBox(width: 300, child: _buildChoice(choice)),
      );
    }
    return Row(
      children: <Widget>[
        for (int index = 0; index < choices.length; index++) ...<Widget>[
          if (index > 0) const SizedBox(width: 10),
          Expanded(
            flex: choices[index].flex,
            child: _buildChoice(choices[index]),
          ),
        ],
      ],
    );
  }

  String get _currentAbility {
    if (_strategy.abilityOptions.contains(_ability)) {
      return _ability;
    }
    return _strategy.defaultAbility;
  }

  Widget _buildChoice(_VbmappTimedMandChoiceSpec choice) {
    return _VbmappMandInlineChoiceGroup(
      label: choice.label,
      value: choice.value,
      values: choice.values,
      onChanged: choice.onChanged,
    );
  }

  List<String> _mandMaterialQuickPicks(VbmappMaterialProfile? profile) {
    return _smartMandQuickPicks(
      profile,
      targetKind: _targetKind,
      fallback: _strategy.quickPickFallback.isEmpty
          ? _fallbackMaterials
          : _strategy.quickPickFallback,
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
          '观察记录',
          style: TextStyle(
            color: _VbmappColors.ink,
            fontSize: 14,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 8),
        _VbmappMand1RecordGrid(
          events: widget.events,
          minSlots: _displayMinSlots,
          selectedIndex: _selectedRecordIndex,
          onSelectIndex: _selectRecord,
          onDeleteIndex: _deleteRecord,
          isQualified: (_VbmappMandEvent event) => _mandEventCountsForItem(
            widget.item,
            event,
            observation: widget.observation,
            responseSchema: widget.responseSchema,
          ),
          metaTextBuilder: (_VbmappMandEvent event) => _mandRecordMetaText(
            event,
            item: widget.item,
            responseSchema: widget.responseSchema,
          ),
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

  void _syncClockTicker() {
    _clockTimer?.cancel();
    _clockTimer = null;
    if (!_observation.isRunning) {
      return;
    }
    if (_finishIfWindowElapsed()) {
      return;
    }
    _clockTimer = Timer.periodic(const Duration(seconds: 1), (Timer timer) {
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
    if (_autoFinishRequested || !_observation.isRunning) {
      return false;
    }
    if (_elapsedSeconds < _plannedSeconds) {
      return false;
    }
    _autoFinishRequested = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        widget.onChangeObservation(
          _observation.finishAtPlannedEnd(DateTime.now()),
        );
      }
    });
    return true;
  }

  void _handlePrimaryTimerAction() {
    final DateTime now = DateTime.now();
    if (!_observation.hasStarted || _observation.ended) {
      widget.onChangeObservation(_observation.restart(now));
      return;
    }
    if (_observation.isRunning) {
      widget.onChangeObservation(_observation.pause(now));
      return;
    }
    widget.onChangeObservation(_observation.resume(now));
  }

  void _finishObservation() {
    _confirmFinishObservation();
  }

  Future<void> _confirmFinishObservation() async {
    final bool? confirmed = await showDialog<bool>(
      context: context,
      builder: (BuildContext context) {
        return const PadDialogViewport(
          child: _VbmappObservationFinishConfirmDialog(),
        );
      },
    );
    if (confirmed != true) {
      return;
    }
    widget.onChangeObservation(_observation.finish(DateTime.now()));
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
    final _VbmappMandPhraseAssessment phraseAssessment = _assessMandPhrase(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: '',
        person: '',
        setting: '',
        example: '',
        responseMode: _strategy.responseModeForPrompt(_currentPromptMode),
        promptLevel: _strategy.promptLevelForPrompt(_currentPromptMode),
        functional: true,
      ),
    );
    if (!_observation.hasStarted) {
      widget.onChangeObservation(_observation.start(DateTime.now()));
    }
    widget.onSubmitEvent(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: '',
        environment: _presentation,
        targetKind: _strategy.targetOptions.isEmpty
            ? _strategy.defaultTargetKind
            : _targetKind,
        person: '',
        setting: '',
        example: '',
        responseMode: _strategy.responseModeForPrompt(_currentPromptMode),
        promptLevel: _strategy.promptLevelForPrompt(_currentPromptMode),
        phraseLevel: _strategy.abilityOptions.isNotEmpty
            ? _currentAbility
            : _strategy.multiWordQualifiedMinCount > 0
                ? phraseAssessment.label
                : '',
        functional: true,
      ),
    );
    _requestController.clear();
    setState(() {
      _selectedRecordIndex = null;
      _promptMode = _strategy.defaultPromptMode;
      _ability = _strategy.defaultAbility;
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

class _VbmappTimedMandChoiceSpec {
  const _VbmappTimedMandChoiceSpec({
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
