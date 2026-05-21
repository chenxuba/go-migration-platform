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

class _VbmappTimedMandInlinePanel extends StatefulWidget {
  const _VbmappTimedMandInlinePanel({
    required this.item,
    required this.materialProfile,
    required this.events,
    required this.observation,
    required this.onSubmitEvent,
    required this.onDeleteEvent,
    required this.onChangeObservation,
  });

  final _VbmappItem item;
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
  String _presentation = '呈现物品';
  String _targetKind = '物品';
  String _promptMode = '自发地';
  int? _selectedRecordIndex;

  _VbmappObservationTimerState get _observation =>
      (widget.observation ?? const _VbmappObservationTimerState())
          .withPlannedMinutes(60);

  bool get _isMand8 => widget.item.itemCode == 'MAND_08M';

  bool get _isMand9 => widget.item.itemCode == 'MAND_09M';

  int get _onePointRequestCount => _scoreCountThreshold(widget.item, 1) ?? 5;

  int get _halfPointRequestCount => _scoreCountThreshold(widget.item, .5) ?? 2;

  int get _plannedMinutes => _plannedMinutesForMandItem(widget.item);

  int get _plannedSeconds => _plannedMinutes * Duration.secondsPerMinute;

  int get _elapsedSeconds => _observation.elapsedSecondsAt(DateTime.now());

  int get _remainingSeconds {
    final int remaining = _plannedSeconds - _elapsedSeconds;
    return remaining > 0 ? remaining : 0;
  }

  bool get _observationMet => _elapsedSeconds >= _plannedSeconds;

  int get _displayMinSlots => _isMand9 ? 6 : _onePointRequestCount;

  String get _recordTitle => '提要求${widget.item.navCode}观察记录';

  String get _scoreReference {
    if (_isMand8) {
      return '参考：60分钟观察窗内，$_halfPointRequestCount个不同要求计0.5分，'
          '$_onePointRequestCount个不同要求且系统判定至少2条为双词+计1分。';
    }
    if (_isMand9) {
      return '参考：30分钟观察窗内，$_halfPointRequestCount个自发不同要求计0.5分，'
          '$_onePointRequestCount个自发不同要求计1分。';
    }
    return '参考：60分钟观察窗内，$_halfPointRequestCount个计0.5分，'
        '$_onePointRequestCount个计1分；仅统计呈现物品条件下的自发要求。';
  }

  String get _observationHint {
    final int qualifiedCount = _qualifiedMandCountForItem(
      widget.item,
      widget.events,
      observation: widget.observation,
    );
    if (!_observation.hasStarted) {
      return '先开启${_plannedMinutes}分钟观察窗，再连续记录孩子的自然要求。';
    }
    if (!_observationMet && qualifiedCount >= _onePointRequestCount) {
      return _isMand8
          ? '已达到数量条件，但观察未满${_plannedMinutes}分钟；双词结构由系统根据原话自动判定。'
          : '已达到1分数量条件，但观察未满${_plannedMinutes}分钟，可继续观察或由老师确认结束。';
    }
    if (!_observationMet && qualifiedCount >= _halfPointRequestCount) {
      return '已达到0.5分数量条件，但观察未满${_plannedMinutes}分钟，建议继续观察。';
    }
    if (_observationMet) {
      return '${_plannedMinutes}分钟观察窗已完成，系统会继续保留新增记录供老师判断。';
    }
    if (_isMand8) {
      return '记录自然情境下的不同要求；双词结构由系统根据原话自动判定。';
    }
    if (_isMand9) {
      return '记录30分钟内的自发不同要求，系统自动去重统计。';
    }
    return '记录自然情境下的自发要求，系统自动统计不同要求数量。';
  }

  @override
  void initState() {
    super.initState();
    _syncClockTicker();
  }

  @override
  void didUpdateWidget(covariant _VbmappTimedMandInlinePanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    final int? selectedIndex = _selectedRecordIndex;
    if (selectedIndex != null && selectedIndex >= widget.events.length) {
      _selectedRecordIndex = null;
    }
    if (oldWidget.observation != widget.observation) {
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
    );
    final double suggestedScore = _suggestMandScore(
      widget.item,
      widget.events,
      observation: widget.observation,
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
                    label: '不同',
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
                label: _isMand8 ? '孩子要求原话' : '孩子要求内容',
                hintText: _isMand8 ? '如：跑快点、该我了、倒果汁' : '如：泡泡、出去、打开',
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
    final Widget presentationChoice = _VbmappMandInlineChoiceGroup(
      label: '目标呈现',
      value: _presentation,
      values: const <String>['呈现物品', '未呈现物品'],
      onChanged: (String value) => setState(() {
        _presentation = value;
      }),
    );
    final Widget targetChoice = _VbmappMandInlineChoiceGroup(
      label: '目标',
      value: _targetKind,
      values: const <String>['物品', '动作', '活动'],
      onChanged: (String value) => setState(() {
        _targetKind = value;
      }),
    );
    if (_isMand8) {
      return Row(
        children: <Widget>[
          Expanded(
            flex: 4,
            child: _VbmappMandInlineChoiceGroup(
              label: '诱发',
              value: _promptMode,
              values: const <String>['自发地', '提问下'],
              onChanged: (String value) => setState(() {
                _promptMode = value;
              }),
            ),
          ),
          const SizedBox(width: 10),
          Expanded(flex: 6, child: presentationChoice),
          const SizedBox(width: 10),
          Expanded(flex: 5, child: targetChoice),
        ],
      );
    }
    return Row(
      children: <Widget>[
        Expanded(flex: 5, child: presentationChoice),
        const SizedBox(width: 10),
        Expanded(flex: 5, child: targetChoice),
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
          ),
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

  void _syncClockTicker() {
    _clockTimer?.cancel();
    _clockTimer = null;
    if (!_observation.isRunning) {
      return;
    }
    _clockTimer = Timer.periodic(const Duration(seconds: 1), (Timer timer) {
      if (!mounted) {
        timer.cancel();
        return;
      }
      setState(() {});
    });
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
    if (!_observation.hasStarted) {
      widget.onChangeObservation(_observation.start(DateTime.now()));
    }
    widget.onSubmitEvent(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: '',
        environment: _presentation,
        targetKind: _targetKind,
        person: '',
        setting: '',
        example: '',
        responseMode:
            _isMand8 ? (_promptMode == '提问下' ? '提问下要求' : '自发要求') : '自发要求',
        promptLevel: _isMand8 ? _promptMode : '',
        phraseLevel: '',
        functional: true,
      ),
    );
    _requestController.clear();
    setState(() {
      _selectedRecordIndex = null;
      if (_isMand8) {
        _promptMode = '自发地';
      }
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

class _VbmappMandQuickPickLabel extends StatelessWidget {
  const _VbmappMandQuickPickLabel({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Text(
      label,
      style: const TextStyle(
        color: _VbmappColors.muted,
        fontSize: 11,
        fontWeight: FontWeight.w900,
      ),
    );
  }
}

class _VbmappMandCompactChip extends StatelessWidget {
  const _VbmappMandCompactChip({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(999),
        child: Ink(
          height: 28,
          padding: const EdgeInsets.symmetric(horizontal: 9),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFE6D9) : const Color(0xFFFFFCFA),
            borderRadius: BorderRadius.circular(999),
            border: Border.all(
              color: selected ? _VbmappColors.orange : _VbmappColors.lineSoft,
            ),
          ),
          child: Center(
            child: Text(
              label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: selected ? _VbmappColors.orangeDeep : _VbmappColors.body,
                fontSize: 11.5,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMand3CoverageColumn extends StatelessWidget {
  const _VbmappMand3CoverageColumn({
    required this.label,
    required this.values,
    required this.color,
  });

  final String label;
  final List<String> values;
  final Color color;

  @override
  Widget build(BuildContext context) {
    final int filledCount = values.length > 2 ? 2 : values.length;
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Text(
            '$label $filledCount/2',
            style: TextStyle(
              color: filledCount >= 2 ? color : _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 7),
          for (int index = 0; index < 2; index++) ...<Widget>[
            _VbmappMand3CoverageSlot(
              text: index < values.length ? values[index] : '待记录',
              filled: index < values.length,
              color: color,
            ),
            if (index == 0) const SizedBox(height: 6),
          ],
        ],
      ),
    );
  }
}

class _VbmappMand3CoverageSlot extends StatelessWidget {
  const _VbmappMand3CoverageSlot({
    required this.text,
    required this.filled,
    required this.color,
  });

  final String text;
  final bool filled;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      alignment: Alignment.center,
      padding: const EdgeInsets.symmetric(horizontal: 6),
      decoration: BoxDecoration(
        color: filled ? color.withOpacity(.09) : const Color(0xFFFFF6EF),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: filled ? color.withOpacity(.24) : _VbmappColors.lineSoft,
        ),
      ),
      child: Text(
        text,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: TextStyle(
          color: filled ? color : _VbmappColors.muted,
          fontSize: 11.5,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _VbmappMand3RecordList extends StatelessWidget {
  const _VbmappMand3RecordList({
    required this.events,
    required this.selectedIndex,
    required this.onSelectIndex,
    required this.onDeleteIndex,
  });

  final List<_VbmappMandEvent> events;
  final int? selectedIndex;
  final ValueChanged<int> onSelectIndex;
  final ValueChanged<int> onDeleteIndex;

  @override
  Widget build(BuildContext context) {
    final int itemCount = events.length < 6 ? 6 : events.length;
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final bool twoColumns = constraints.maxWidth >= 340;
        if (!twoColumns) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              for (int index = 0; index < itemCount; index++) ...<Widget>[
                _buildSlot(index),
                if (index < itemCount - 1) const SizedBox(height: 7),
              ],
            ],
          );
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            for (int index = 0; index < itemCount; index += 2) ...<Widget>[
              Row(
                children: <Widget>[
                  Expanded(child: _buildSlot(index)),
                  if (index + 1 < itemCount) ...<Widget>[
                    const SizedBox(width: 8),
                    Expanded(child: _buildSlot(index + 1)),
                  ],
                ],
              ),
              if (index + 2 < itemCount) const SizedBox(height: 7),
            ],
          ],
        );
      },
    );
  }

  Widget _buildSlot(int index) {
    return _VbmappMand3RecordSlot(
      index: index,
      event: index < events.length ? events[index] : null,
      selected: selectedIndex == index,
      onTap: () => onSelectIndex(index),
      onDelete: () => onDeleteIndex(index),
    );
  }
}

class _VbmappMand3RecordSlot extends StatelessWidget {
  const _VbmappMand3RecordSlot({
    required this.index,
    required this.event,
    required this.selected,
    required this.onTap,
    required this.onDelete,
  });

  final int index;
  final _VbmappMandEvent? event;
  final bool selected;
  final VoidCallback onTap;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    final _VbmappMandEvent? row = event;
    final bool filled = row != null;
    final bool qualified = row?.isQualified ?? false;
    final Color accent = qualified ? _VbmappColors.green : _VbmappColors.muted;
    final String dimensionText = row == null ? '' : _mand3DimensionText(row);
    return Material(
      key: ValueKey<String>('vbmapp-mand3-record-$index'),
      color: Colors.transparent,
      child: InkWell(
        onTap: filled ? onTap : null,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 42,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFFBF7) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: selected
                  ? _VbmappColors.orange
                  : filled
                      ? accent.withOpacity(.42)
                      : _VbmappColors.lineSoft,
            ),
          ),
          child: Row(
            children: <Widget>[
              SizedBox(
                width: 22,
                child: Text(
                  '${index + 1}',
                  style: TextStyle(
                    color: filled ? accent : _VbmappColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              Expanded(
                child: filled
                    ? Row(
                        children: <Widget>[
                          Flexible(
                            child: Text(
                              _mandRequestText(row),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                color: _VbmappColors.ink,
                                fontSize: 12.5,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                          ),
                          Container(
                            width: 1,
                            height: 14,
                            margin: const EdgeInsets.symmetric(horizontal: 7),
                            color: _VbmappColors.lineSoft,
                          ),
                          Flexible(
                            child: Text(
                              dimensionText,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                color: _VbmappColors.ink,
                                fontSize: 12.5,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                          ),
                        ],
                      )
                    : const Text(
                        '等待记录',
                        style: TextStyle(
                          color: _VbmappColors.muted,
                          fontSize: 12.5,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
              ),
              const SizedBox(width: 8),
              _VbmappMandRecordActions(
                filled: filled,
                qualified: qualified,
                selected: selected,
                accent: accent,
                onDelete: onDelete,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappMandInlineChoiceGroup extends StatelessWidget {
  const _VbmappMandInlineChoiceGroup({
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
    return Row(
      mainAxisSize: MainAxisSize.max,
      children: <Widget>[
        Text(
          label,
          maxLines: 1,
          softWrap: false,
          style: const TextStyle(
            color: _VbmappColors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(width: 7),
        Expanded(
          child: Row(
            children: <Widget>[
              for (int index = 0; index < values.length; index++) ...<Widget>[
                if (index > 0) const SizedBox(width: 5),
                Expanded(
                  child: _VbmappMandChoiceButton(
                    label: values[index],
                    selected: values[index] == value,
                    onTap: () => onChanged(values[index]),
                  ),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }
}

class _VbmappMandChoiceButton extends StatelessWidget {
  const _VbmappMandChoiceButton({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(9),
        child: Ink(
          height: 30,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            color: selected ? _VbmappColors.orange : Colors.white,
            borderRadius: BorderRadius.circular(9),
            border: Border.all(
              color: selected ? _VbmappColors.orange : _VbmappColors.line,
            ),
          ),
          child: Center(
            child: Text(
              label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              softWrap: false,
              textAlign: TextAlign.center,
              style: TextStyle(
                color: selected ? Colors.white : _VbmappColors.body,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMandInlineTextField extends StatelessWidget {
  const _VbmappMandInlineTextField({
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
      onTapOutside: (_) => FocusManager.instance.primaryFocus?.unfocus(),
      style: const TextStyle(
        color: _VbmappColors.ink,
        fontSize: 14,
        fontWeight: FontWeight.w800,
      ),
      decoration: InputDecoration(
        labelText: label,
        hintText: hintText,
        floatingLabelBehavior: FloatingLabelBehavior.auto,
        labelStyle: const TextStyle(
          color: Color(0xFFB8A79E),
          fontSize: 13,
          fontWeight: FontWeight.w600,
        ),
        floatingLabelStyle: const TextStyle(
          color: _VbmappColors.orangeDeep,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
        hintStyle: const TextStyle(
          color: Color(0xFFC7B9B1),
          fontSize: 13,
          fontWeight: FontWeight.w500,
        ),
        isDense: true,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 14, vertical: 13),
        filled: true,
        fillColor: Colors.white,
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: _VbmappColors.lineSoft),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: _VbmappColors.orange),
        ),
      ),
    );
  }
}

class _VbmappMandMaterialChip extends StatelessWidget {
  const _VbmappMandMaterialChip({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(999),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFE6D9) : Colors.white,
            borderRadius: BorderRadius.circular(999),
            border: Border.all(
              color: selected ? _VbmappColors.orange : _VbmappColors.lineSoft,
            ),
          ),
          child: Text(
            label,
            style: TextStyle(
              color: selected ? _VbmappColors.orangeDeep : _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMand1RecordGrid extends StatelessWidget {
  const _VbmappMand1RecordGrid({
    required this.events,
    required this.minSlots,
    required this.selectedIndex,
    required this.onSelectIndex,
    required this.onDeleteIndex,
    this.isQualified,
    this.metaTextBuilder,
  });

  final List<_VbmappMandEvent> events;
  final int minSlots;
  final int? selectedIndex;
  final ValueChanged<int> onSelectIndex;
  final ValueChanged<int> onDeleteIndex;
  final bool Function(_VbmappMandEvent event)? isQualified;
  final String Function(_VbmappMandEvent event)? metaTextBuilder;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        const double spacing = 8;
        final int itemCount =
            events.length < minSlots ? minSlots : events.length;
        final bool twoColumns = constraints.maxWidth >= 360;
        if (!twoColumns) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              for (int index = 0; index < itemCount; index++) ...<Widget>[
                _buildSlot(index),
                if (index < itemCount - 1) const SizedBox(height: spacing),
              ],
            ],
          );
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            for (int index = 0; index < itemCount; index += 2) ...<Widget>[
              if (index + 1 < itemCount)
                IntrinsicHeight(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: <Widget>[
                      Expanded(child: _buildSlot(index)),
                      const SizedBox(width: spacing),
                      Expanded(child: _buildSlot(index + 1)),
                    ],
                  ),
                )
              else
                _buildSlot(index),
              if (index + 2 < itemCount) const SizedBox(height: spacing),
            ],
          ],
        );
      },
    );
  }

  Widget _buildSlot(int index) {
    return _VbmappMand1RecordSlot(
      index: index,
      event: index < events.length ? events[index] : null,
      selected: selectedIndex == index,
      onTap: () => onSelectIndex(index),
      onDelete: () => onDeleteIndex(index),
      qualificationResolver: isQualified,
      metaTextBuilder: metaTextBuilder,
    );
  }
}

class _VbmappMand1RecordSlot extends StatelessWidget {
  const _VbmappMand1RecordSlot({
    required this.index,
    required this.event,
    required this.selected,
    required this.onTap,
    required this.onDelete,
    this.qualificationResolver,
    this.metaTextBuilder,
  });

  final int index;
  final _VbmappMandEvent? event;
  final bool selected;
  final VoidCallback onTap;
  final VoidCallback onDelete;
  final bool Function(_VbmappMandEvent event)? qualificationResolver;
  final String Function(_VbmappMandEvent event)? metaTextBuilder;

  @override
  Widget build(BuildContext context) {
    final _VbmappMandEvent? row = event;
    final bool filled = row != null;
    final bool qualified = row == null
        ? false
        : (qualificationResolver?.call(row) ?? row.isQualified);
    final Color accent = qualified ? _VbmappColors.green : _VbmappColors.muted;
    final String requestText = row == null ? '' : _mandRequestText(row);
    final BorderRadius radius = BorderRadius.circular(10);
    return Material(
      key: ValueKey<String>('vbmapp-mand-record-$index'),
      color: Colors.transparent,
      child: InkWell(
        onTap: filled ? onTap : null,
        borderRadius: radius,
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 9),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFFBF7) : Colors.white,
            borderRadius: radius,
            border: Border.all(
              color: selected
                  ? _VbmappColors.orange
                  : filled
                      ? accent.withOpacity(.42)
                      : _VbmappColors.lineSoft,
            ),
          ),
          child: ConstrainedBox(
            constraints: const BoxConstraints(minHeight: 34),
            child: Row(
              children: <Widget>[
                Container(
                  width: 28,
                  height: 28,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: filled
                        ? accent.withOpacity(.12)
                        : const Color(0xFFFFF6EF),
                    shape: BoxShape.circle,
                  ),
                  child: Text(
                    '${index + 1}',
                    style: TextStyle(
                      color: filled ? accent : _VbmappColors.muted,
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                const SizedBox(width: 9),
                Expanded(
                  child: filled
                      ? Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: <Widget>[
                            Text(
                              requestText,
                              maxLines: 2,
                              style: const TextStyle(
                                color: _VbmappColors.ink,
                                fontSize: 13,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                            const SizedBox(height: 3),
                            Text(
                              metaTextBuilder?.call(row) ??
                                  _mandRecordMetaText(row),
                              maxLines: 2,
                              style: const TextStyle(
                                color: _VbmappColors.body,
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                          ],
                        )
                      : const Text(
                          '等待记录一条有效要求',
                          style: TextStyle(
                            color: _VbmappColors.muted,
                            fontSize: 13,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                ),
                const SizedBox(width: 8),
                _VbmappMandRecordActions(
                  filled: filled,
                  qualified: qualified,
                  selected: selected,
                  accent: accent,
                  onDelete: onDelete,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMandRecordActions extends StatelessWidget {
  const _VbmappMandRecordActions({
    required this.filled,
    required this.qualified,
    required this.selected,
    required this.accent,
    required this.onDelete,
  });

  final bool filled;
  final bool qualified;
  final bool selected;
  final Color accent;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    if (!selected || !filled) {
      return Text(
        filled ? (qualified ? '计入' : '不计') : '-',
        style: TextStyle(
          color: accent,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      );
    }
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Text(
          qualified ? '计入' : '不计',
          style: TextStyle(
            color: accent,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(width: 6),
        Material(
          color: Colors.transparent,
          child: InkWell(
            key: const ValueKey<String>('vbmapp-mand-delete-record'),
            onTap: onDelete,
            borderRadius: BorderRadius.circular(999),
            child: Ink(
              height: 22,
              padding: const EdgeInsets.symmetric(horizontal: 7),
              decoration: BoxDecoration(
                color: const Color(0xFFFFE6D9),
                borderRadius: BorderRadius.circular(999),
                border: Border.all(color: _VbmappColors.orange),
              ),
              child: const Center(
                child: Text(
                  '删除',
                  style: TextStyle(
                    color: _VbmappColors.orangeDeep,
                    fontSize: 11,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _VbmappMandRecorderPanel extends StatelessWidget {
  const _VbmappMandRecorderPanel({
    required this.item,
    required this.materialProfile,
    required this.events,
    required this.onAddEvent,
  });

  final _VbmappItem item;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> events;
  final VoidCallback onAddEvent;

  @override
  Widget build(BuildContext context) {
    final int qualifiedCount = _qualifiedMandCount(events);
    final double suggestedScore = _suggestMandScore(item, events);
    final bool generalizationMode = item.itemCode == 'MAND_03M';
    final Map<String, int> generalizationCounts =
        _mandGeneralizationCounts(events);
    final List<_VbmappMandEvent> visibleEvents =
        events.length <= 4 ? events : events.sublist(events.length - 4);
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Row(
            children: <Widget>[
              Icon(Icons.record_voice_over_outlined,
                  color: item.color, size: 21),
              const SizedBox(width: 8),
              const Expanded(
                child: Text(
                  '提要求事件记录',
                  style: TextStyle(
                    color: _VbmappColors.ink,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              _VbmappEvidenceMetric(
                label: '有效',
                value: '$qualifiedCount',
                color: item.color,
              ),
              const SizedBox(width: 8),
              if (generalizationMode) ...<Widget>[
                _VbmappEvidenceMetric(
                  label: '人',
                  value: '${generalizationCounts['people'] ?? 0}/2',
                  color: item.color,
                ),
                const SizedBox(width: 8),
                _VbmappEvidenceMetric(
                  label: '环境',
                  value: '${generalizationCounts['settings'] ?? 0}/2',
                  color: item.color,
                ),
                const SizedBox(width: 8),
                _VbmappEvidenceMetric(
                  label: '例子',
                  value: '${generalizationCounts['examples'] ?? 0}/2',
                  color: item.color,
                ),
                const SizedBox(width: 8),
              ],
              _VbmappEvidenceMetric(
                label: '建议',
                value: _formatScore(suggestedScore),
                color: item.color,
              ),
            ],
          ),
          if (materialProfile != null) ...<Widget>[
            const SizedBox(height: 10),
            _VbmappInlineInfo(
              icon: Icons.inventory_2_outlined,
              text: materialProfile!.label,
            ),
          ],
          const SizedBox(height: 12),
          if (visibleEvents.isEmpty)
            const _VbmappEmptyEvidence(text: '还没有记录孩子实际发出的要求')
          else
            Column(
              children: <Widget>[
                for (final _VbmappMandEvent event in visibleEvents)
                  _VbmappMandEventRow(event: event),
              ],
            ),
          const SizedBox(height: 12),
          Align(
            alignment: Alignment.centerLeft,
            child: _VbmappSmallActionButton(
              icon: Icons.add_rounded,
              label: generalizationMode ? '记录一次泛化要求' : '记录一次要求',
              onTap: onAddEvent,
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappMandEventRow extends StatelessWidget {
  const _VbmappMandEventRow({required this.event});

  final _VbmappMandEvent event;

  @override
  Widget build(BuildContext context) {
    final Color color =
        event.isQualified ? _VbmappColors.green : _VbmappColors.muted;
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 9),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          Icon(
            event.isQualified
                ? Icons.check_circle_outline_rounded
                : Icons.radio_button_unchecked_rounded,
            color: color,
            size: 18,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              event.summary,
              style: const TextStyle(
                color: _VbmappColors.ink,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          const SizedBox(width: 8),
          Text(
            event.promptLevel,
            style: TextStyle(
              color: color,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappEvidenceMetric extends StatelessWidget {
  const _VbmappEvidenceMetric({
    required this.label,
    required this.value,
    required this.color,
  });

  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: color.withOpacity(.1),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        '$label $value',
        style: TextStyle(
          color: color,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _VbmappInlineInfo extends StatelessWidget {
  const _VbmappInlineInfo({
    required this.icon,
    required this.text,
  });

  final IconData icon;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Icon(icon, color: _VbmappColors.orange, size: 17),
        const SizedBox(width: 6),
        Expanded(
          child: Text(
            text,
            style: const TextStyle(
              color: _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
      ],
    );
  }
}

class _VbmappEmptyEvidence extends StatelessWidget {
  const _VbmappEmptyEvidence({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 13),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Text(
        text,
        style: const TextStyle(
          color: _VbmappColors.muted,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _VbmappSmallActionButton extends StatelessWidget {
  const _VbmappSmallActionButton({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 42,
          padding: const EdgeInsets.symmetric(horizontal: 13),
          decoration: BoxDecoration(
            color: _VbmappColors.orange,
            borderRadius: BorderRadius.circular(10),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(icon, size: 18, color: Colors.white),
              const SizedBox(width: 6),
              Text(
                label,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 13,
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

class _VbmappMand4TimerButton extends StatelessWidget {
  const _VbmappMand4TimerButton({
    required this.icon,
    required this.label,
    required this.onTap,
    this.filled = false,
    this.compact = false,
    super.key,
  });

  final IconData icon;
  final String label;
  final VoidCallback? onTap;
  final bool filled;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final bool enabled = onTap != null;
    final Color textColor = enabled
        ? filled
            ? Colors.white
            : _VbmappColors.orangeDeep
        : _VbmappColors.muted;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(9),
        child: Ink(
          height: compact ? 28 : 32,
          padding: EdgeInsets.symmetric(horizontal: compact ? 8 : 10),
          decoration: BoxDecoration(
            color: filled && enabled
                ? _VbmappColors.orange
                : enabled
                    ? const Color(0xFFFFFCFA)
                    : const Color(0xFFF7F1ED),
            borderRadius: BorderRadius.circular(9),
            border: Border.all(
              color: enabled ? _VbmappColors.orange : const Color(0xFFE2D6CE),
            ),
          ),
          child: Center(
            child: Row(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.center,
              children: <Widget>[
                Icon(icon, size: compact ? 14 : 16, color: textColor),
                SizedBox(width: compact ? 4 : 5),
                Text(
                  label,
                  style: TextStyle(
                    color: textColor,
                    fontSize: compact ? 11.5 : 12.5,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMand4TimerMetricChip extends StatelessWidget {
  const _VbmappMand4TimerMetricChip({
    required this.label,
    required this.value,
    required this.tone,
  });

  final String label;
  final String value;
  final Color tone;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
      decoration: BoxDecoration(
        color: tone.withOpacity(.1),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: tone.withOpacity(.18)),
      ),
      child: Text(
        '$label $value',
        style: TextStyle(
          color: tone,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _VbmappMand4QuickRecordDialog extends StatefulWidget {
  const _VbmappMand4QuickRecordDialog({required this.materialProfile});

  final VbmappMaterialProfile? materialProfile;

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

  @override
  void dispose() {
    _requestController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final List<String> materials = _smartMandQuickPicks(
      widget.materialProfile,
      targetKind: _targetKind,
      fallback: _fallbackMaterials,
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
              Row(
                children: <Widget>[
                  Expanded(
                    flex: 4,
                    child: _VbmappMandInlineChoiceGroup(
                      label: '诱发',
                      value: _promptMode,
                      values: const <String>['自发地', '提问下'],
                      onChanged: (String value) => setState(() {
                        _promptMode = value;
                      }),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    flex: 6,
                    child: _VbmappMandInlineChoiceGroup(
                      label: '目标呈现',
                      value: _presentation,
                      values: const <String>['呈现物品', '未呈现物品'],
                      onChanged: (String value) => setState(() {
                        _presentation = value;
                      }),
                    ),
                  ),
                  const SizedBox(width: 12),
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
                ],
              ),
              const SizedBox(height: 12),
              _VbmappMandInlineTextField(
                controller: _requestController,
                label: '孩子要求原话',
                hintText: '如：泡泡、出去、该我了',
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

  void _submit() {
    final String request = _requestController.text.trim();
    if (request.isEmpty) {
      return;
    }
    Navigator.of(context).pop(
      _VbmappMandEvent(
        utterance: request,
        target: request,
        motivationContext: '',
        environment: _presentation,
        targetKind: _targetKind,
        person: '',
        setting: '',
        example: '',
        responseMode: _promptMode == '提问下' ? '提问下要求' : '自发要求',
        promptLevel: _promptMode,
        phraseLevel: '',
        functional: true,
      ),
    );
  }
}

class _VbmappObservationFinishConfirmDialog extends StatelessWidget {
  const _VbmappObservationFinishConfirmDialog();

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
              const Text(
                '确认结束观察？',
                style: TextStyle(
                  color: _VbmappColors.ink,
                  fontSize: 18,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 10),
              const Text(
                '结束后会保留当前计时和记录。如只是暂时离开，建议点暂停。',
                style: TextStyle(
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
