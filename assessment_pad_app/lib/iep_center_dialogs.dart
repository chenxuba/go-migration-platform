part of 'iep_center_page.dart';

class _IepPeriodDraft {
  const _IepPeriodDraft({required this.start});

  final DateTime start;
}

class _IepPeriodEditDialog extends StatefulWidget {
  const _IepPeriodEditDialog({
    required this.initialStart,
    required this.monthCount,
  });

  final DateTime initialStart;
  final int monthCount;

  @override
  State<_IepPeriodEditDialog> createState() => _IepPeriodEditDialogState();
}

class _IepPeriodEditDialogState extends State<_IepPeriodEditDialog> {
  late DateTime _start;

  DateTime get _end => _periodEndFor(_start, widget.monthCount);

  @override
  void initState() {
    super.initState();
    _start = _dateOnly(widget.initialStart);
  }

  Future<void> _pickStartDate() async {
    final DateTime? picked = await showPadDatePicker(
      context: context,
      title: '选择周期开始日期',
      helperText: '请选择周期开始日期，结束日期将按自然月自动计算',
      initialDate: _start,
      today: _start,
      minDate: DateTime(_start.year - 1, 1),
      maxDate: DateTime(_start.year + 1, 12, 31),
    );
    if (picked == null || !mounted) {
      return;
    }
    setState(() {
      _start = _dateOnly(picked);
    });
  }

  void _submit() {
    Navigator.of(context).pop(_IepPeriodDraft(start: _start));
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: 540,
        padding: const EdgeInsets.fromLTRB(24, 22, 24, 22),
        decoration: BoxDecoration(
          color: _IepColors.surface,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: _IepColors.line),
          boxShadow: _iepShadow(
            color: const Color(0x20B05F32),
            blur: 32,
            offset: const Offset(0, 16),
          ),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                const Text(
                  '编辑周期',
                  style: TextStyle(
                    color: _IepColors.ink,
                    fontSize: 22,
                    fontWeight: FontWeight.w900,
                    height: 1,
                  ),
                ),
                const Spacer(),
                _IepPeriodTypePill(text: '${widget.monthCount}个月周期'),
                const SizedBox(width: 10),
                _IepDialogIconButton(
                  icon: Icons.close_rounded,
                  onTap: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 22),
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: const Color(0xFFFFFAF6),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: _IepColors.lightLine),
              ),
              child: Column(
                children: <Widget>[
                  _IepPeriodDateTile(
                    label: '周期开始',
                    value: _formatDateDash(_start),
                    icon: Icons.event_available_rounded,
                    onTap: _pickStartDate,
                  ),
                  const SizedBox(height: 10),
                  _IepPeriodDateTile(
                    label: '周期结束',
                    value: _formatDateDash(_end),
                    icon: Icons.event_repeat_rounded,
                    trailingText: '自动计算',
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
            Row(
              children: <Widget>[
                const Icon(
                  Icons.info_outline_rounded,
                  size: 16,
                  color: _IepColors.muted,
                ),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    '同步后会按${widget.monthCount}个月周期更新表格中的实施日期，并重算对应月计划、周计划起止日期。',
                    style: const TextStyle(
                      color: _IepColors.text,
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      height: 1.35,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 22),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: <Widget>[
                _IepDialogAction(
                  label: '取消',
                  onTap: () => Navigator.of(context).pop(),
                ),
                const SizedBox(width: 10),
                _IepDialogAction(
                  label: '确认同步',
                  filled: true,
                  onTap: _submit,
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _IepPeriodTypePill extends StatelessWidget {
  const _IepPeriodTypePill({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: _IepColors.orangeSoft,
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: const Color(0xFFFFD8C6)),
      ),
      child: Text(
        text,
        style: const TextStyle(
          color: _IepColors.orangeDeep,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _IepPeriodDateTile extends StatelessWidget {
  const _IepPeriodDateTile({
    required this.label,
    required this.value,
    required this.icon,
    this.trailingText,
    this.onTap,
  });

  final String label;
  final String value;
  final IconData icon;
  final String? trailingText;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final bool clickable = onTap != null;
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(13),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Container(
          height: 58,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(13),
            border: Border.all(
              color: clickable ? const Color(0xFFFFCDB4) : _IepColors.line,
            ),
          ),
          child: Row(
            children: <Widget>[
              Container(
                width: 34,
                height: 34,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: clickable
                      ? _IepColors.orangeSoft
                      : const Color(0xFFF8EEE6),
                  shape: BoxShape.circle,
                ),
                child: Icon(
                  icon,
                  size: 19,
                  color: clickable ? _IepColors.orangeDeep : _IepColors.muted,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Text(
                      label,
                      style: const TextStyle(
                        color: _IepColors.muted,
                        fontSize: 11,
                        fontWeight: FontWeight.w900,
                        height: 1,
                      ),
                    ),
                    const SizedBox(height: 7),
                    Text(
                      value,
                      style: const TextStyle(
                        color: _IepColors.ink,
                        fontSize: 17,
                        fontWeight: FontWeight.w900,
                        height: 1,
                      ),
                    ),
                  ],
                ),
              ),
              if (trailingText != null)
                Text(
                  trailingText!,
                  style: const TextStyle(
                    color: _IepColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                  ),
                )
              else
                const Icon(
                  Icons.chevron_right_rounded,
                  color: _IepColors.orangeDeep,
                  size: 24,
                ),
            ],
          ),
        ),
      ),
    );
  }
}

class _IepDialogIconButton extends StatelessWidget {
  const _IepDialogIconButton({required this.icon, required this.onTap});

  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      shape: const CircleBorder(),
      child: InkWell(
        onTap: onTap,
        customBorder: const CircleBorder(),
        child: Container(
          width: 36,
          height: 36,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: const Color(0xFFFFFAF6),
            shape: BoxShape.circle,
            border: Border.all(color: _IepColors.lightLine),
          ),
          child: Icon(icon, size: 21, color: _IepColors.text),
        ),
      ),
    );
  }
}

class _IepDialogAction extends StatelessWidget {
  const _IepDialogAction({
    required this.label,
    required this.onTap,
    this.filled = false,
    this.icon,
  });

  final String label;
  final VoidCallback onTap;
  final bool filled;
  final IconData? icon;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          height: 42,
          constraints: const BoxConstraints(minWidth: 92),
          alignment: Alignment.center,
          padding: const EdgeInsets.symmetric(horizontal: 20),
          decoration: BoxDecoration(
            color: filled ? _IepColors.orange : Colors.white,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(
              color: filled ? _IepColors.orange : _IepColors.line,
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              if (icon != null) ...<Widget>[
                Icon(
                  icon,
                  size: 16,
                  color: filled ? Colors.white : _IepColors.text,
                ),
                const SizedBox(width: 5),
              ],
              Text(
                label,
                style: TextStyle(
                  color: filled ? Colors.white : _IepColors.text,
                  fontSize: 14,
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

class _IepRegenerateConfirmDialog extends StatelessWidget {
  const _IepRegenerateConfirmDialog();

  @override
  Widget build(BuildContext context) {
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: 500,
        padding: const EdgeInsets.fromLTRB(24, 22, 24, 22),
        decoration: BoxDecoration(
          color: _IepColors.surface,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: _IepColors.line),
          boxShadow: _iepShadow(
            color: const Color(0x20B05F32),
            blur: 32,
            offset: const Offset(0, 16),
          ),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                Container(
                  width: 42,
                  height: 42,
                  alignment: Alignment.center,
                  decoration: const BoxDecoration(
                    color: _IepColors.orangeSoft,
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(
                    Icons.refresh_rounded,
                    color: _IepColors.orangeDeep,
                    size: 23,
                  ),
                ),
                const SizedBox(width: 12),
                const Expanded(
                  child: Text(
                    '确认重新生成IEP？',
                    style: TextStyle(
                      color: _IepColors.ink,
                      fontSize: 22,
                      fontWeight: FontWeight.w900,
                      height: 1,
                    ),
                  ),
                ),
                _IepDialogIconButton(
                  icon: Icons.close_rounded,
                  onTap: () => Navigator.of(context).pop(false),
                ),
              ],
            ),
            const SizedBox(height: 18),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
              decoration: BoxDecoration(
                color: const Color(0xFFFFFAF6),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: _IepColors.lightLine),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: const <Widget>[
                  Icon(
                    Icons.info_outline_rounded,
                    size: 18,
                    color: _IepColors.orangeDeep,
                  ),
                  SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '重新生成会用AI生成一份新的IEP计划，并覆盖当前页面已有的IEP内容。已确认后请谨慎操作。',
                      style: TextStyle(
                        color: _IepColors.text,
                        fontSize: 13,
                        fontWeight: FontWeight.w700,
                        height: 1.45,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 22),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: <Widget>[
                _IepDialogAction(
                  label: '取消',
                  onTap: () => Navigator.of(context).pop(false),
                ),
                const SizedBox(width: 10),
                _IepDialogAction(
                  label: '确认重新生成',
                  filled: true,
                  icon: Icons.refresh_rounded,
                  onTap: () => Navigator.of(context).pop(true),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _IepWeeklyPlanMissingMonthConfirmDialog extends StatelessWidget {
  const _IepWeeklyPlanMissingMonthConfirmDialog({
    required this.monthLabel,
    required this.weekNumber,
    required this.planTitle,
  });

  final String monthLabel;
  final int weekNumber;
  final String planTitle;

  @override
  Widget build(BuildContext context) {
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: 468,
        padding: const EdgeInsets.fromLTRB(22, 22, 22, 20),
        decoration: BoxDecoration(
          color: _IepColors.surface,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: _IepColors.line),
          boxShadow: _iepShadow(
            color: const Color(0x20B05F32),
            blur: 32,
            offset: const Offset(0, 16),
          ),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: <Widget>[
                Container(
                  width: 42,
                  height: 42,
                  alignment: Alignment.center,
                  decoration: const BoxDecoration(
                    color: _IepColors.orangeSoft,
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(
                    Icons.info_outline_rounded,
                    color: _IepColors.orangeDeep,
                    size: 23,
                  ),
                ),
                const SizedBox(width: 12),
                const Expanded(
                  child: Text(
                    '当前还没有月计划',
                    style: TextStyle(
                      color: _IepColors.ink,
                      fontSize: 22,
                      fontWeight: FontWeight.w900,
                      height: 1,
                    ),
                  ),
                ),
                _IepDialogIconButton(
                  icon: Icons.close_rounded,
                  onTap: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 18),
            Container(
              width: double.infinity,
              padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
              decoration: BoxDecoration(
                color: const Color(0xFFFFFAF6),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: _IepColors.lightLine),
              ),
              child: Text(
                '是否直接基于$planTitle生成$monthLabel第$weekNumber周周计划？',
                style: const TextStyle(
                  color: _IepColors.text,
                  fontSize: 13,
                  fontWeight: FontWeight.w700,
                  height: 1.45,
                ),
              ),
            ),
            const SizedBox(height: 22),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: <Widget>[
                _IepDialogAction(
                  label: '先生成月计划',
                  onTap: () => Navigator.of(context).pop(false),
                ),
                const SizedBox(width: 10),
                _IepDialogAction(
                  label: '直接生成周计划',
                  filled: true,
                  icon: Icons.auto_awesome_rounded,
                  onTap: () => Navigator.of(context).pop(true),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

enum _GoalEditType { longGoal, shortGoal }

class _GoalEditRequest {
  const _GoalEditRequest._({
    required this.domainIndex,
    required this.type,
    this.shortGoalIndex,
  });

  factory _GoalEditRequest.longGoal({required int domainIndex}) {
    return _GoalEditRequest._(
      domainIndex: domainIndex,
      type: _GoalEditType.longGoal,
    );
  }

  factory _GoalEditRequest.shortGoal({
    required int domainIndex,
    required int shortGoalIndex,
  }) {
    return _GoalEditRequest._(
      domainIndex: domainIndex,
      type: _GoalEditType.shortGoal,
      shortGoalIndex: shortGoalIndex,
    );
  }

  final int domainIndex;
  final _GoalEditType type;
  final int? shortGoalIndex;

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        other is _GoalEditRequest &&
            other.domainIndex == domainIndex &&
            other.type == type &&
            other.shortGoalIndex == shortGoalIndex;
  }

  @override
  int get hashCode => Object.hash(domainIndex, type, shortGoalIndex);
}

class _GoalEditResult {
  const _GoalEditResult({
    this.longGoals,
    this.shortGoals,
    this.syncRelatedPlans = false,
  });

  final List<String>? longGoals;
  final List<_DocShortGoalData>? shortGoals;
  final bool syncRelatedPlans;
}

class _IepGoalEditDialog extends StatefulWidget {
  const _IepGoalEditDialog({
    required this.domain,
    required this.request,
  });

  final _DocDomainData domain;
  final _GoalEditRequest request;

  @override
  State<_IepGoalEditDialog> createState() => _IepGoalEditDialogState();
}

class _IepGoalEditDialogState extends State<_IepGoalEditDialog> {
  final List<TextEditingController> _longGoalControllers =
      <TextEditingController>[];
  final List<_ShortGoalDraft> _shortGoalDrafts = <_ShortGoalDraft>[];

  bool get _editingLongGoal => widget.request.type == _GoalEditType.longGoal;

  bool get _canRemoveCurrentShortGoal {
    return !_editingLongGoal &&
        (widget.domain.shortGoals.length > 1 || _shortGoalDrafts.length > 1);
  }

  int get _shortGoalIndex {
    final int index = widget.request.shortGoalIndex ?? 0;
    if (widget.domain.shortGoals.isEmpty || index <= 0) {
      return 0;
    }
    final int lastIndex = widget.domain.shortGoals.length - 1;
    return index > lastIndex ? lastIndex : index;
  }

  _DocShortGoalData get _shortGoal {
    if (widget.domain.shortGoals.isEmpty) {
      return const _DocShortGoalData('', '个训', '');
    }
    return widget.domain.shortGoals[_shortGoalIndex];
  }

  @override
  void initState() {
    super.initState();
    _longGoalControllers.addAll(
      widget.domain.longGoals
          .map((String goal) => TextEditingController(text: goal)),
    );
    if (!_editingLongGoal) {
      _shortGoalDrafts.add(
        _ShortGoalDraft(
          goal: _shortGoal.goal,
          lesson: _normalizeLesson(_shortGoal.lesson),
          period: _shortGoal.period,
        ),
      );
    }
  }

  @override
  void dispose() {
    for (final TextEditingController controller in _longGoalControllers) {
      controller.dispose();
    }
    for (final _ShortGoalDraft draft in _shortGoalDrafts) {
      draft.dispose();
    }
    super.dispose();
  }

  String _normalizeLesson(String lesson) {
    return lesson == '集体课' ? '集体课' : '个训';
  }

  void _addLongGoal() {
    setState(() {
      _longGoalControllers.add(TextEditingController());
    });
  }

  void _removeLongGoal(int index) {
    if (_longGoalControllers.length <= 1) {
      return;
    }
    setState(() {
      _longGoalControllers.removeAt(index).dispose();
    });
  }

  void _addShortGoal() {
    setState(() {
      _shortGoalDrafts.add(
        _ShortGoalDraft(
          goal: '',
          lesson: '个训',
          period: _shortGoal.period,
        ),
      );
    });
  }

  void _removeShortGoal(int index) {
    if (index < 0 || index >= _shortGoalDrafts.length) {
      return;
    }
    if (index == 0 && !_canRemoveCurrentShortGoal) {
      return;
    }
    setState(() {
      _shortGoalDrafts.removeAt(index).dispose();
    });
  }

  void _changeShortGoalLesson(int index, String lesson) {
    setState(() {
      _shortGoalDrafts[index].lesson = lesson;
    });
  }

  void _submit({required bool syncRelatedPlans}) {
    if (_editingLongGoal) {
      final List<String> goals = _longGoalControllers
          .map((TextEditingController controller) => controller.text.trim())
          .where((String value) => value.isNotEmpty)
          .toList();
      Navigator.of(context).pop(
        _GoalEditResult(
          longGoals: goals.isEmpty ? <String>[''] : goals,
          syncRelatedPlans: syncRelatedPlans,
        ),
      );
      return;
    }
    final List<_DocShortGoalData> editedShortGoals = _shortGoalDrafts
        .map((_ShortGoalDraft draft) => draft.toData())
        .where((_DocShortGoalData data) => data.goal.trim().isNotEmpty)
        .toList();
    final List<_DocShortGoalData> nextShortGoals =
        List<_DocShortGoalData>.from(widget.domain.shortGoals);
    if (nextShortGoals.isEmpty) {
      nextShortGoals.addAll(
        editedShortGoals.isEmpty
            ? <_DocShortGoalData>[_shortGoal.copyWith(goal: '')]
            : editedShortGoals,
      );
    } else if (editedShortGoals.isEmpty) {
      if (nextShortGoals.length > 1) {
        nextShortGoals.removeAt(_shortGoalIndex);
      } else {
        nextShortGoals[_shortGoalIndex] = _shortGoal.copyWith(goal: '');
      }
    } else {
      nextShortGoals
        ..removeAt(_shortGoalIndex)
        ..insertAll(_shortGoalIndex, editedShortGoals);
    }
    Navigator.of(context).pop(
      _GoalEditResult(
        shortGoals: nextShortGoals,
        syncRelatedPlans: syncRelatedPlans,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final int? shortGoalIndex = widget.request.shortGoalIndex;
    final String title = _editingLongGoal ? '编辑长期目标' : '编辑短期目标';
    final String location = _editingLongGoal
        ? '${widget.domain.domain} · 长期目标'
        : '${widget.domain.domain} · 短期目标${(shortGoalIndex ?? 0) + 1}';
    return Dialog(
      elevation: 0,
      backgroundColor: Colors.transparent,
      insetPadding: const EdgeInsets.symmetric(horizontal: 40, vertical: 28),
      child: Container(
        width: _editingLongGoal ? 660 : 620,
        padding: const EdgeInsets.fromLTRB(24, 22, 24, 22),
        decoration: BoxDecoration(
          color: _IepColors.surface,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: _IepColors.line),
          boxShadow: _iepShadow(
            color: const Color(0x20B05F32),
            blur: 32,
            offset: const Offset(0, 16),
          ),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                Text(
                  title,
                  style: const TextStyle(
                    color: _IepColors.ink,
                    fontSize: 22,
                    fontWeight: FontWeight.w900,
                    height: 1,
                  ),
                ),
                const SizedBox(width: 12),
                _IepPeriodTypePill(text: location),
                const Spacer(),
                _IepDialogIconButton(
                  icon: Icons.close_rounded,
                  onTap: () => Navigator.of(context).pop(),
                ),
              ],
            ),
            const SizedBox(height: 20),
            if (_editingLongGoal)
              _LongGoalEditor(
                controllers: _longGoalControllers,
                onAdd: _addLongGoal,
                onRemove: _removeLongGoal,
              )
            else
              _ShortGoalEditor(
                drafts: _shortGoalDrafts,
                firstShortGoalNumber: _shortGoalIndex + 1,
                canRemoveCurrent: _canRemoveCurrentShortGoal,
                onLessonChanged: _changeShortGoalLesson,
                onAdd: _addShortGoal,
                onRemove: _removeShortGoal,
              ),
            const SizedBox(height: 12),
            const _GoalSyncHint(),
            const SizedBox(height: 20),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: <Widget>[
                _IepDialogAction(
                  label: '取消',
                  onTap: () => Navigator.of(context).pop(),
                ),
                const SizedBox(width: 10),
                _IepDialogAction(
                  label: '仅保存当前表格',
                  icon: Icons.save_outlined,
                  onTap: () => _submit(syncRelatedPlans: false),
                ),
                const SizedBox(width: 10),
                _IepDialogAction(
                  label: '保存并同步',
                  filled: true,
                  icon: Icons.sync_rounded,
                  onTap: () => _submit(syncRelatedPlans: true),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _ShortGoalDraft {
  _ShortGoalDraft({
    required String goal,
    required this.lesson,
    required String period,
  })  : goalController = TextEditingController(text: goal),
        periodController = TextEditingController(text: period);

  final TextEditingController goalController;
  final TextEditingController periodController;
  String lesson;

  _DocShortGoalData toData() {
    return _DocShortGoalData(
      goalController.text.trim(),
      lesson,
      periodController.text.trim(),
    );
  }

  void dispose() {
    goalController.dispose();
    periodController.dispose();
  }
}

class _LongGoalEditor extends StatelessWidget {
  const _LongGoalEditor({
    required this.controllers,
    required this.onAdd,
    required this.onRemove,
  });

  final List<TextEditingController> controllers;
  final VoidCallback onAdd;
  final ValueChanged<int> onRemove;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(maxHeight: 330),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF6),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: SingleChildScrollView(
        child: Column(
          children: <Widget>[
            for (int index = 0;
                index < controllers.length;
                index += 1) ...<Widget>[
              _GoalTextField(
                label: '长期目标 ${index + 1}',
                controller: controllers[index],
                minLines: 2,
                maxLines: 3,
                trailing: _SmallIconAction(
                  icon: Icons.delete_outline_rounded,
                  enabled: controllers.length > 1,
                  onTap: () => onRemove(index),
                ),
              ),
              const SizedBox(height: 10),
            ],
            Align(
              alignment: Alignment.centerLeft,
              child: _AddGoalButton(onTap: onAdd),
            ),
          ],
        ),
      ),
    );
  }
}

class _ShortGoalEditor extends StatelessWidget {
  const _ShortGoalEditor({
    required this.drafts,
    required this.firstShortGoalNumber,
    required this.canRemoveCurrent,
    required this.onLessonChanged,
    required this.onAdd,
    required this.onRemove,
  });

  final List<_ShortGoalDraft> drafts;
  final int firstShortGoalNumber;
  final bool canRemoveCurrent;
  final void Function(int index, String lesson) onLessonChanged;
  final VoidCallback onAdd;
  final ValueChanged<int> onRemove;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(maxHeight: 372),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF6),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: <Widget>[
              const Text(
                '当前短期目标',
                style: TextStyle(
                  color: _IepColors.text,
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const Spacer(),
              _AddGoalButton(
                label: '新增一条短期目标',
                onTap: onAdd,
              ),
            ],
          ),
          const SizedBox(height: 10),
          Expanded(
            child: SingleChildScrollView(
              child: Column(
                children: <Widget>[
                  for (int index = 0; index < drafts.length; index += 1)
                    Padding(
                      padding: EdgeInsets.only(
                        bottom: index == drafts.length - 1 ? 0 : 12,
                      ),
                      child: _ShortGoalDraftCard(
                        index: index,
                        number: firstShortGoalNumber + index,
                        draft: drafts[index],
                        canRemove: index > 0 || canRemoveCurrent,
                        onLessonChanged: (String value) =>
                            onLessonChanged(index, value),
                        onRemove: () => onRemove(index),
                      ),
                    ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ShortGoalDraftCard extends StatelessWidget {
  const _ShortGoalDraftCard({
    required this.index,
    required this.number,
    required this.draft,
    required this.canRemove,
    required this.onLessonChanged,
    required this.onRemove,
  });

  final int index;
  final int number;
  final _ShortGoalDraft draft;
  final bool canRemove;
  final ValueChanged<String> onLessonChanged;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _IepColors.line),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: <Widget>[
              Text(
                '短期目标 $number',
                style: const TextStyle(
                  color: _IepColors.ink,
                  fontSize: 13,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const Spacer(),
              _SmallIconAction(
                icon: Icons.delete_outline_rounded,
                enabled: canRemove,
                onTap: onRemove,
              ),
            ],
          ),
          const SizedBox(height: 8),
          _GoalTextField(
            fieldKey: ValueKey<String>('short-goal-$index-goal'),
            label: '目标内容',
            controller: draft.goalController,
            minLines: 2,
            maxLines: 4,
          ),
          const SizedBox(height: 12),
          Row(
            children: <Widget>[
              Expanded(
                flex: 8,
                child: _LessonSegmentedPicker(
                  label: '课程形式',
                  value: draft.lesson,
                  optionKeyPrefix: 'short-goal-$index-lesson',
                  onChanged: onLessonChanged,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                flex: 13,
                child: _GoalTextField(
                  label: '起止日期',
                  controller: draft.periodController,
                  minLines: 1,
                  maxLines: 1,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _LessonSegmentedPicker extends StatelessWidget {
  const _LessonSegmentedPicker({
    required this.label,
    required this.value,
    required this.onChanged,
    this.optionKeyPrefix,
  });

  final String label;
  final String value;
  final ValueChanged<String> onChanged;
  final String? optionKeyPrefix;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          label,
          style: const TextStyle(
            color: _IepColors.text,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 7),
        Container(
          height: 42,
          padding: const EdgeInsets.all(3),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: _IepColors.line),
          ),
          child: Row(
            children: <Widget>[
              _LessonOption(
                label: '个训',
                active: value == '个训',
                optionKey: optionKeyPrefix == null
                    ? null
                    : ValueKey<String>('$optionKeyPrefix-个训'),
                onTap: () => onChanged('个训'),
              ),
              _LessonOption(
                label: '集体课',
                active: value == '集体课',
                optionKey: optionKeyPrefix == null
                    ? null
                    : ValueKey<String>('$optionKeyPrefix-集体课'),
                onTap: () => onChanged('集体课'),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _LessonOption extends StatelessWidget {
  const _LessonOption({
    required this.label,
    required this.active,
    required this.onTap,
    this.optionKey,
  });

  final String label;
  final bool active;
  final VoidCallback onTap;
  final Key? optionKey;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Material(
        color: Colors.transparent,
        borderRadius: BorderRadius.circular(9),
        child: InkWell(
          key: optionKey,
          onTap: onTap,
          borderRadius: BorderRadius.circular(9),
          child: Container(
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: active ? _IepColors.orange : Colors.transparent,
              borderRadius: BorderRadius.circular(9),
            ),
            child: Text(
              label,
              style: TextStyle(
                color: active ? Colors.white : _IepColors.text,
                fontSize: 13,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _GoalTextField extends StatelessWidget {
  const _GoalTextField({
    required this.label,
    required this.controller,
    required this.minLines,
    required this.maxLines,
    this.fieldKey,
    this.trailing,
  });

  final String label;
  final TextEditingController controller;
  final int minLines;
  final int maxLines;
  final Key? fieldKey;
  final Widget? trailing;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Row(
          children: <Widget>[
            Text(
              label,
              style: const TextStyle(
                color: _IepColors.text,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
            const Spacer(),
            if (trailing != null) trailing!,
          ],
        ),
        const SizedBox(height: 7),
        TextField(
          key: fieldKey,
          controller: controller,
          minLines: minLines,
          maxLines: maxLines,
          style: const TextStyle(
            color: _IepColors.ink,
            fontSize: 14,
            fontWeight: FontWeight.w700,
            height: 1.35,
          ),
          decoration: InputDecoration(
            isDense: true,
            filled: true,
            fillColor: Colors.white,
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 12, vertical: 11),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: _IepColors.line),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide:
                  const BorderSide(color: _IepColors.orange, width: 1.2),
            ),
          ),
        ),
      ],
    );
  }
}

class _AddGoalButton extends StatelessWidget {
  const _AddGoalButton({
    required this.onTap,
    this.label = '新增一条',
  });

  final VoidCallback onTap;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(13),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Container(
          height: 32,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(13),
            border: Border.all(color: const Color(0xFFFFD8C6)),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(Icons.add_rounded,
                  size: 17, color: _IepColors.orangeDeep),
              const SizedBox(width: 4),
              Text(
                label,
                style: const TextStyle(
                  color: _IepColors.orangeDeep,
                  fontSize: 12,
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

class _SmallIconAction extends StatelessWidget {
  const _SmallIconAction({
    required this.icon,
    required this.onTap,
    this.enabled = true,
  });

  final IconData icon;
  final VoidCallback onTap;
  final bool enabled;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      shape: const CircleBorder(),
      child: InkWell(
        onTap: enabled ? onTap : null,
        customBorder: const CircleBorder(),
        child: Container(
          width: 26,
          height: 26,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: enabled ? Colors.white : const Color(0xFFF8EEE6),
            shape: BoxShape.circle,
            border: Border.all(color: _IepColors.lightLine),
          ),
          child: Icon(
            icon,
            size: 16,
            color: enabled ? _IepColors.text : _IepColors.muted,
          ),
        ),
      ),
    );
  }
}

class _GoalSyncHint extends StatelessWidget {
  const _GoalSyncHint();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: _IepColors.orangeSoft,
        borderRadius: BorderRadius.circular(13),
        border: Border.all(color: const Color(0xFFFFD8C6)),
      ),
      child: const Row(
        children: <Widget>[
          Icon(Icons.info_outline_rounded,
              size: 16, color: _IepColors.orangeDeep),
          SizedBox(width: 7),
          Expanded(
            child: Text(
              '修改目标后，可选择仅更新当前IEP总表，也可以同步更新关联月计划、周计划。',
              style: TextStyle(
                color: _IepColors.text,
                fontSize: 12,
                fontWeight: FontWeight.w800,
                height: 1.35,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
