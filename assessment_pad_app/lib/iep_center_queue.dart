part of 'iep_center_page.dart';

class _StudentQueuePanel extends StatefulWidget {
  const _StudentQueuePanel({
    required this.recordClient,
    required this.selectedRecord,
    required this.statusOverrides,
    required this.onRecordSelected,
    required this.onInitialLoadSettled,
  });

  final IepAssessmentRecordClient recordClient;
  final IepAssessmentRecordSummary? selectedRecord;
  final Map<String, String> statusOverrides;
  final ValueChanged<IepAssessmentRecordSummary> onRecordSelected;
  final VoidCallback onInitialLoadSettled;

  @override
  State<_StudentQueuePanel> createState() => _StudentQueuePanelState();
}

class _StudentQueuePanelState extends State<_StudentQueuePanel> {
  static const String _authTokenStorageKey = 'auth_token';

  List<IepAssessmentRecordSummary> _records = <IepAssessmentRecordSummary>[];
  bool _loading = true;
  String _error = '';
  int _totalCount = 0;
  _QueueFilter _filter = _QueueFilter.all;

  @override
  void initState() {
    super.initState();
    runAfterRouteEntrance(context, _loadRecords);
  }

  Future<void> _loadRecords() async {
    if (mounted && !_loading) {
      setState(() {
        _loading = true;
        _error = '';
      });
    }
    try {
      final SharedPreferences prefs = await SharedPreferences.getInstance();
      final String token = prefs.getString(_authTokenStorageKey) ?? '';
      final IepAssessmentRecordPage page =
          await widget.recordClient.fetchRecordsPage(
        token,
        pageIndex: 1,
        pageSize: 30,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _records = page.items;
        _totalCount = page.total;
        _loading = false;
      });
      final IepAssessmentRecordSummary? selectedRecord =
          _selectedRecordFrom(
        _visibleRecordsForWithOverrides(
          _filter,
          page.items,
          widget.statusOverrides,
        ),
      );
      if (selectedRecord != null) {
        widget.onRecordSelected(selectedRecord);
      }
      widget.onInitialLoadSettled();
    } on IepAssessmentRecordApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _error = error.message;
        _loading = false;
      });
      widget.onInitialLoadSettled();
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _error = '评估记录加载失败：$error';
        _loading = false;
      });
      widget.onInitialLoadSettled();
    }
  }

  IepAssessmentRecordSummary? _selectedRecordFrom(
    List<IepAssessmentRecordSummary> records,
  ) {
    if (records.isEmpty) {
      return null;
    }
    final IepAssessmentRecordSummary? selectedRecord = widget.selectedRecord;
    if (selectedRecord != null) {
      for (final IepAssessmentRecordSummary record in records) {
        if (_sameRecord(record, selectedRecord)) {
          return null;
        }
      }
    }
    return records.first;
  }

  void _changeFilter(_QueueFilter filter) {
    if (_filter == filter) {
      return;
    }
    setState(() {
      _filter = filter;
    });
    final IepAssessmentRecordSummary? selectedRecord =
        _selectedRecordFrom(
      _visibleRecordsForWithOverrides(
        filter,
        _records,
        widget.statusOverrides,
      ),
    );
    if (selectedRecord != null) {
      widget.onRecordSelected(selectedRecord);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.9),
        borderRadius: BorderRadius.circular(18),
        boxShadow: _iepShadow(),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: const <Widget>[
              Expanded(
                child: Text(
                  '学员IEP队列',
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: _IepColors.ink,
                    fontSize: 17,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              _QueueFilterButton(),
            ],
          ),
          const SizedBox(height: 12),
          _QueueTabs(selected: _filter, onChanged: _changeFilter),
          const SizedBox(height: 12),
          _CompactStatsStrip(
            records: _records,
            totalCount: _totalCount,
            statusOverrides: widget.statusOverrides,
          ),
          const SizedBox(height: 12),
          Expanded(
            child: _QueueList(
              records: _visibleRecords,
              statusOverrides: widget.statusOverrides,
              selectedRecord: widget.selectedRecord,
              loading: _loading,
              error: _error,
              onRetry: _loadRecords,
              onRecordSelected: widget.onRecordSelected,
            ),
          ),
        ],
      ),
    );
  }

  List<IepAssessmentRecordSummary> get _visibleRecords {
    return _visibleRecordsForWithOverrides(
      _filter,
      _records,
      widget.statusOverrides,
    );
  }
}

class _QueueFilterButton extends StatelessWidget {
  const _QueueFilterButton();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: _IepColors.orangeSoft,
        borderRadius: BorderRadius.circular(15),
      ),
      child: Row(
        children: const <Widget>[
          Icon(Icons.filter_alt_outlined, size: 16, color: _IepColors.orange),
          SizedBox(width: 4),
          Text(
            '筛选',
            style: TextStyle(
              color: _IepColors.orange,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

enum _QueueFilter { all, pending, awaitingConfirm, confirmed }

class _QueueTabs extends StatelessWidget {
  const _QueueTabs({
    required this.selected,
    required this.onChanged,
  });

  final _QueueFilter selected;
  final ValueChanged<_QueueFilter> onChanged;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 34,
      padding: const EdgeInsets.all(3),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF4EC),
        borderRadius: BorderRadius.circular(17),
      ),
      child: Row(
        children: <Widget>[
          Expanded(
            child: _QueueTab(
              label: '全部',
              active: selected == _QueueFilter.all,
              onTap: () => onChanged(_QueueFilter.all),
            ),
          ),
          Expanded(
            child: _QueueTab(
              label: '待生成',
              active: selected == _QueueFilter.pending,
              onTap: () => onChanged(_QueueFilter.pending),
            ),
          ),
          Expanded(
            child: _QueueTab(
              label: '待确认',
              active: selected == _QueueFilter.awaitingConfirm,
              onTap: () => onChanged(_QueueFilter.awaitingConfirm),
            ),
          ),
          Expanded(
            child: _QueueTab(
              label: '已确认',
              active: selected == _QueueFilter.confirmed,
              onTap: () => onChanged(_QueueFilter.confirmed),
            ),
          ),
        ],
      ),
    );
  }
}

class _QueueTab extends StatelessWidget {
  const _QueueTab({
    required this.label,
    required this.onTap,
    this.active = false,
  });

  final String label;
  final VoidCallback onTap;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: active ? Colors.white : Colors.transparent,
            borderRadius: BorderRadius.circular(14),
            boxShadow: active
                ? _iepShadow(
                    color: const Color(0x0FB05F32),
                    blur: 8,
                    offset: const Offset(0, 3),
                  )
                : const <BoxShadow>[],
          ),
          child: Text(
            label,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              color: active ? _IepColors.orangeDeep : _IepColors.text,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _CompactStatsStrip extends StatelessWidget {
  const _CompactStatsStrip({
    required this.records,
    required this.totalCount,
    this.statusOverrides = const <String, String>{},
  });

  final List<IepAssessmentRecordSummary> records;
  final int totalCount;
  final Map<String, String> statusOverrides;

  @override
  Widget build(BuildContext context) {
    final int pendingCount = records
        .where((IepAssessmentRecordSummary record) {
          final String label = _QueueStatusStyle.fromPlanStatus(
            _effectiveRecordStatus(record, statusOverrides),
          ).label;
          return label == '待生成' || label == '生成中';
        })
        .length;
    final int awaitingConfirmCount = records
        .where((IepAssessmentRecordSummary record) =>
            _QueueStatusStyle.fromPlanStatus(
                  _effectiveRecordStatus(record, statusOverrides),
                ).label ==
            '待确认')
        .length;
    final int confirmedCount = records
        .where((IepAssessmentRecordSummary record) =>
            _QueueStatusStyle.fromPlanStatus(
                  _effectiveRecordStatus(record, statusOverrides),
                ).label ==
            '已确认')
        .length;
    final String totalText =
        totalCount > 0 ? totalCount.toString() : records.length.toString();
    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        children: <Widget>[
          Expanded(child: _SmallStat(number: totalText, label: '评估记录')),
          const _StatDivider(),
          Expanded(child: _SmallStat(number: '$pendingCount', label: '待生成')),
          const _StatDivider(),
          Expanded(
              child: _SmallStat(number: '$awaitingConfirmCount', label: '待确认')),
          const _StatDivider(),
          Expanded(child: _SmallStat(number: '$confirmedCount', label: '已确认')),
        ],
      ),
    );
  }
}

class _SmallStat extends StatelessWidget {
  const _SmallStat({required this.number, required this.label});

  final String number;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: <Widget>[
        Text(
          number,
          style: const TextStyle(
            color: _IepColors.ink,
            fontSize: 18,
            fontWeight: FontWeight.w900,
            height: 1,
          ),
        ),
        const SizedBox(height: 6),
        Text(
          label,
          overflow: TextOverflow.ellipsis,
          style: const TextStyle(
            color: _IepColors.muted,
            fontSize: 10,
            fontWeight: FontWeight.w700,
          ),
        ),
      ],
    );
  }
}

class _StatDivider extends StatelessWidget {
  const _StatDivider();

  @override
  Widget build(BuildContext context) {
    return Container(width: 1, height: 28, color: _IepColors.lightLine);
  }
}

class _QueueList extends StatelessWidget {
  const _QueueList({
    required this.records,
    required this.statusOverrides,
    required this.selectedRecord,
    required this.loading,
    required this.error,
    required this.onRetry,
    required this.onRecordSelected,
  });

  final List<IepAssessmentRecordSummary> records;
  final Map<String, String> statusOverrides;
  final IepAssessmentRecordSummary? selectedRecord;
  final bool loading;
  final String error;
  final VoidCallback onRetry;
  final ValueChanged<IepAssessmentRecordSummary> onRecordSelected;

  @override
  Widget build(BuildContext context) {
    if (loading) {
      return const _QueueListSkeleton();
    }
    if (error.trim().isNotEmpty) {
      return _QueueStateView(
        icon: Icons.wifi_off_rounded,
        title: '评估记录加载失败',
        message: error,
        actionLabel: '重试',
        onAction: onRetry,
      );
    }
    if (records.isEmpty) {
      return const _QueueStateView(
        icon: Icons.assignment_outlined,
        title: '暂无评估记录',
        message: '完成评估后会出现在这里',
      );
    }
    return ListView.separated(
      physics: const ClampingScrollPhysics(),
      padding: EdgeInsets.zero,
      itemCount: records.length,
      separatorBuilder: (_, __) => const SizedBox(height: 10),
      itemBuilder: (BuildContext context, int index) {
        final IepAssessmentRecordSummary record = records[index];
        return _QueueStudentCard(
          student: _QueueStudent.fromRecord(
            record,
            statusOverride: _effectiveRecordStatus(record, statusOverrides),
            active: selectedRecord == null
                ? index == 0
                : _sameRecord(record, selectedRecord!),
          ),
          onTap: () => onRecordSelected(record),
        );
      },
    );
  }
}

class _QueueStateView extends StatelessWidget {
  const _QueueStateView({
    required this.icon,
    required this.title,
    this.message = '',
    this.actionLabel = '',
    this.onAction,
  });

  final IconData icon;
  final String title;
  final String message;
  final String actionLabel;
  final VoidCallback? onAction;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 16),
        decoration: BoxDecoration(
          color: const Color(0xFFFFFAF5),
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: _IepColors.lightLine),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Icon(icon, size: 28, color: _IepColors.orangeDeep),
            const SizedBox(height: 8),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _IepColors.ink,
                fontSize: 13,
                fontWeight: FontWeight.w900,
              ),
            ),
            if (message.trim().isNotEmpty) ...<Widget>[
              const SizedBox(height: 6),
              Text(
                message,
                textAlign: TextAlign.center,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _IepColors.text,
                  fontSize: 11,
                  fontWeight: FontWeight.w700,
                  height: 1.25,
                ),
              ),
            ],
            if (actionLabel.trim().isNotEmpty && onAction != null) ...<Widget>[
              const SizedBox(height: 10),
              _MiniQueueAction(label: actionLabel, onTap: onAction!),
            ],
          ],
        ),
      ),
    );
  }
}

class _MiniQueueAction extends StatelessWidget {
  const _MiniQueueAction({required this.label, required this.onTap});

  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: _IepColors.orangeSoft,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          height: 28,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          alignment: Alignment.center,
          child: Text(
            label,
            style: const TextStyle(
              color: _IepColors.orangeDeep,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _QueueStudent {
  const _QueueStudent({
    required this.name,
    required this.age,
    required this.status,
    required this.statusColor,
    required this.statusBg,
    required this.assessment,
    required this.period,
    required this.avatarAsset,
    this.active = false,
  });

  factory _QueueStudent.fromRecord(
    IepAssessmentRecordSummary record, {
    required bool active,
    String? statusOverride,
  }) {
    final _QueueStatusStyle status = _QueueStatusStyle.fromPlanStatus(
      statusOverride ?? record.iepPlanStatus,
    );
    return _QueueStudent(
      name: record.studentName.trim().isEmpty ? '未命名学员' : record.studentName,
      age: _recordAgeText(record),
      status: status.label,
      statusColor: status.color,
      statusBg: status.background,
      assessment:
          '${_recordAssessmentName(record)} · ${_recordDateText(record.assessmentDate)}',
      period: _recordPeriodText(record),
      avatarAsset: _avatarAssetForRecord(record),
      active: active,
    );
  }

  final String name;
  final String age;
  final String status;
  final Color statusColor;
  final Color statusBg;
  final String assessment;
  final String period;
  final String avatarAsset;
  final bool active;
}

class _QueueStatusStyle {
  const _QueueStatusStyle({
    required this.label,
    required this.color,
    required this.background,
  });

  factory _QueueStatusStyle.fromPlanStatus(String status) {
    return switch (status.trim()) {
      'generating' => const _QueueStatusStyle(
          label: '生成中',
          color: _IepColors.orangeDeep,
          background: Color(0xFFFFE7D8),
        ),
      'confirmed' => const _QueueStatusStyle(
          label: '已确认',
          color: _IepColors.green,
          background: _IepColors.greenSoft,
        ),
      'draft' => const _QueueStatusStyle(
          label: '待确认',
          color: _IepColors.yellow,
          background: _IepColors.yellowSoft,
        ),
      _ => const _QueueStatusStyle(
          label: '待生成',
          color: _IepColors.orange,
          background: _IepColors.orangeSoft,
        ),
    };
  }

  final String label;
  final Color color;
  final Color background;
}

const List<String> _queueAvatarAssets = <String>[
  'assets/avatars/student_chenxu.png',
  'assets/avatars/student_chenxiaoyu.png',
  'assets/avatars/student_linyinuo.png',
  'assets/avatars/student_zhoushuyan.png',
  'assets/avatars/student_tangmuchen.png',
];

String _avatarAssetForRecord(IepAssessmentRecordSummary record) {
  final String seed = '${record.studentId}:${record.id}:${record.studentName}';
  int hash = 0;
  for (int index = 0; index < seed.length; index += 1) {
    hash = (hash * 31 + seed.codeUnitAt(index)) & 0x7fffffff;
  }
  return _queueAvatarAssets[hash % _queueAvatarAssets.length];
}

String _recordAgeText(IepAssessmentRecordSummary record) {
  if (record.ageYears > 0 || record.ageMonths > 0) {
    return '${record.ageYears}岁${record.ageMonths}月';
  }
  final DateTime? birthDate = DateTime.tryParse(record.birthDate);
  final DateTime? assessmentDate = DateTime.tryParse(record.assessmentDate);
  if (birthDate == null || assessmentDate == null) {
    return '年龄未知';
  }
  int totalMonths = (assessmentDate.year - birthDate.year) * 12 +
      assessmentDate.month -
      birthDate.month;
  if (assessmentDate.day < birthDate.day) {
    totalMonths -= 1;
  }
  if (totalMonths < 0) {
    return '年龄未知';
  }
  return '${totalMonths ~/ 12}岁${totalMonths % 12}月';
}

String _recordAssessmentName(IepAssessmentRecordSummary record) {
  if (record.source == 'ERXIN') {
    return '儿心量表';
  }
  if (record.assessmentName.trim().isNotEmpty) {
    return record.assessmentName.trim();
  }
  return record.assessmentCode == 'PEP3' ? 'PEP-3' : '评估记录';
}

String _recordDateText(String value) {
  final DateTime? date = DateTime.tryParse(value.trim());
  if (date == null) {
    return value.trim().isEmpty ? '-' : value.trim();
  }
  return _formatDateDash(date);
}

String _recordPeriodText(IepAssessmentRecordSummary record) {
  final DateTime? start = DateTime.tryParse(record.assessmentDate.trim());
  if (start == null) {
    return record.iepPlanStatus.trim().isEmpty ? '待确认周期' : '周期待同步';
  }
  final DateTime end = _periodEndFor(start, 3);
  return _formatDotRange(start, end);
}

bool _sameRecord(
  IepAssessmentRecordSummary left,
  IepAssessmentRecordSummary right,
) {
  return left.id == right.id &&
      left.source.trim().toUpperCase() == right.source.trim().toUpperCase();
}

List<IepAssessmentRecordSummary> _visibleRecordsForWithOverrides(
  _QueueFilter filter,
  List<IepAssessmentRecordSummary> records,
  Map<String, String> statusOverrides,
) {
  return records.where((IepAssessmentRecordSummary record) {
    final String status =
        _QueueStatusStyle.fromPlanStatus(
          _effectiveRecordStatus(record, statusOverrides),
        ).label;
    return switch (filter) {
      _QueueFilter.all => true,
      _QueueFilter.pending => status == '待生成' || status == '生成中',
      _QueueFilter.awaitingConfirm => status == '待确认',
      _QueueFilter.confirmed => status == '已确认',
    };
  }).toList(growable: false);
}

String _effectiveRecordStatus(
  IepAssessmentRecordSummary record,
  Map<String, String> statusOverrides,
) {
  return statusOverrides[_recordIdentityKey(record)] ?? record.iepPlanStatus;
}

class _QueueStudentCard extends StatelessWidget {
  const _QueueStudentCard({required this.student, required this.onTap});

  final _QueueStudent student;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      borderRadius: BorderRadius.circular(14),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          height: 80,
          padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 8),
          decoration: BoxDecoration(
            color: student.active ? const Color(0xFFFFF3EB) : Colors.white,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(
              color: student.active
                  ? const Color(0xFFFFB792)
                  : _IepColors.lightLine,
              width: student.active ? 1.4 : 1,
            ),
          ),
          child: Row(
            children: <Widget>[
              _QueueAvatar(asset: student.avatarAsset, active: student.active),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Row(
                      children: <Widget>[
                        Expanded(
                          child: Text(
                            '${student.name} · ${student.age}',
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              color: _IepColors.ink,
                              fontSize: 13,
                              fontWeight: FontWeight.w900,
                            ),
                          ),
                        ),
                        _StatusPill(
                          text: student.status,
                          color: student.statusColor,
                          bg: student.statusBg,
                        ),
                      ],
                    ),
                    const SizedBox(height: 6),
                    Text(
                      student.assessment,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _IepColors.text,
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        height: 1,
                      ),
                    ),
                    const SizedBox(height: 6),
                    Row(
                      children: <Widget>[
                        const Icon(
                          Icons.date_range_rounded,
                          size: 13,
                          color: _IepColors.muted,
                        ),
                        const SizedBox(width: 4),
                        Expanded(
                          child: Text(
                            student.period,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              color: _IepColors.muted,
                              fontSize: 11,
                              fontWeight: FontWeight.w700,
                              height: 1,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _QueueAvatar extends StatelessWidget {
  const _QueueAvatar({required this.asset, required this.active});

  final String asset;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 44,
      height: 44,
      padding: const EdgeInsets.all(2),
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: Colors.white,
        border: Border.all(
          color: active ? const Color(0xFFFFA878) : const Color(0xFFFFDFC8),
          width: active ? 2 : 1.4,
        ),
        boxShadow: active
            ? _iepShadow(
                color: const Color(0x22E96F43),
                blur: 10,
                offset: const Offset(0, 4),
              )
            : const <BoxShadow>[],
      ),
      child: ClipOval(
        child: Image.asset(
          asset,
          width: 40,
          height: 40,
          fit: BoxFit.cover,
          filterQuality: FilterQuality.medium,
        ),
      ),
    );
  }
}

class _StatusPill extends StatelessWidget {
  const _StatusPill({
    required this.text,
    required this.color,
    required this.bg,
  });

  final String text;
  final Color color;
  final Color bg;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 23,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: color,
          fontSize: 10,
          fontWeight: FontWeight.w900,
          height: 1,
        ),
      ),
    );
  }
}
