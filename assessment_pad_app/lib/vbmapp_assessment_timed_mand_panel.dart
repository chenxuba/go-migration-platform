part of 'vbmapp_assessment_page.dart';

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
