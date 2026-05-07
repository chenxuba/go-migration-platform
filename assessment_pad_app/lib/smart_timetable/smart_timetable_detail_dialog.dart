part of '../smart_timetable_page.dart';

extension _SmartTimetableStateDetail on _SmartTimetablePageState {
  Future<void> _openLessonDetail(_LessonCell lesson) async {
    final String scheduleId = lesson.id.trim();
    if (!mounted || scheduleId.isEmpty) {
      return;
    }
    final String initialSubtitle = lesson.person.split('·').first.trim();
    _updateState(() {
      _teacherDropdownOpen = false;
      _periodGroupDropdownOpen = false;
      _schedulePanelOpen = false;
      _openFilterKind = null;
    });
    await showDialog<void>(
      context: context,
      barrierDismissible: true,
      barrierColor: Colors.black.withOpacity(.34),
      builder: (BuildContext dialogContext) {
        return PadDialogViewport(
          child: _ScheduleDetailDialog(
            scheduleId: scheduleId,
            initialTitle: lesson.title,
            initialSubtitle:
                initialSubtitle.isEmpty ? lesson.person : initialSubtitle,
            initialStatus: lesson.status,
            onClose: () => Navigator.of(dialogContext).maybePop(),
            loadDetail: () => _fetchScheduleDetail(scheduleId),
            onDelete: _deleteScheduleFromDetail,
          ),
        );
      },
    );
  }

  Future<TimetableScheduleDetail> _fetchScheduleDetail(
      String scheduleId) async {
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      throw const TimetableApiException('登录已失效，请重新登录');
    }
    return widget.timetableClient.fetchScheduleDetail(
      token,
      scheduleId: scheduleId,
    );
  }

  Future<void> _deleteScheduleFromDetail(
    TimetableScheduleDetail detail,
    ScheduleDeleteScope scope,
  ) async {
    final String token = await _readAuthToken();
    if (token.trim().isEmpty) {
      throw const TimetableApiException('登录已失效，请重新登录');
    }
    final int canceled = await widget.timetableClient.cancelScheduleScoped(
      token,
      scheduleId: detail.id,
      scope: scope,
    );
    if (canceled <= 0) {
      throw const TimetableApiException('删除失败，请刷新后重试');
    }
    await _loadTimetable(showSkeleton: false);
    _showScheduleMessage(
      scope == ScheduleDeleteScope.future ? '已删除后续全部日程' : '已删除当前日程',
      tone: PadMessageTone.success,
    );
  }
}

class _ScheduleDetailDialog extends StatefulWidget {
  const _ScheduleDetailDialog({
    required this.scheduleId,
    required this.initialTitle,
    required this.initialSubtitle,
    required this.initialStatus,
    required this.onClose,
    required this.loadDetail,
    required this.onDelete,
  });

  final String scheduleId;
  final String initialTitle;
  final String initialSubtitle;
  final _LessonStatus initialStatus;
  final VoidCallback onClose;
  final Future<TimetableScheduleDetail> Function() loadDetail;
  final Future<void> Function(
    TimetableScheduleDetail detail,
    ScheduleDeleteScope scope,
  ) onDelete;

  @override
  State<_ScheduleDetailDialog> createState() => _ScheduleDetailDialogState();
}

class _ScheduleDetailDialogState extends State<_ScheduleDetailDialog> {
  TimetableScheduleDetail? _detail;
  bool _loading = true;
  bool _deleting = false;
  String? _loadError;
  String? _actionError;

  @override
  void initState() {
    super.initState();
    unawaited(_loadDetail());
  }

  Future<void> _loadDetail() async {
    if (mounted) {
      setState(() {
        _loading = true;
        _loadError = null;
        _actionError = null;
      });
    }
    try {
      final TimetableScheduleDetail detail = await widget.loadDetail();
      if (!mounted) {
        return;
      }
      setState(() {
        _detail = detail;
        _loading = false;
      });
    } on TimetableApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _loadError = error.message;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _loadError = '加载日程详情失败：$error';
      });
    }
  }

  Future<void> _handleDelete(ScheduleDeleteScope scope) async {
    final TimetableScheduleDetail? detail = _detail;
    if (detail == null || _deleting) {
      return;
    }
    final bool confirmed = await _confirmDelete(scope);
    if (!confirmed || !mounted) {
      return;
    }
    setState(() {
      _deleting = true;
      _actionError = null;
    });
    try {
      await widget.onDelete(detail, scope);
      if (!mounted) {
        return;
      }
      Navigator.of(context).pop();
    } on TimetableApiException catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _actionError = error.message;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _actionError = '删除失败：$error';
      });
    } finally {
      if (!mounted) {
        return;
      }
      setState(() {
        _deleting = false;
      });
    }
  }

  Future<bool> _confirmDelete(ScheduleDeleteScope scope) async {
    final bool? confirmed = await showDialog<bool>(
      context: context,
      barrierDismissible: true,
      barrierColor: Colors.black.withOpacity(.34),
      builder: (BuildContext dialogContext) {
        return PadDialogViewport(
          child: _ScheduleDeleteConfirmDialog(
            scope: scope,
            onCancel: () => Navigator.of(dialogContext).pop(false),
            onConfirm: () => Navigator.of(dialogContext).pop(true),
          ),
        );
      },
    );
    return confirmed == true;
  }

  @override
  Widget build(BuildContext context) {
    final TimetableScheduleDetail? detail = _detail;
    final String title = detail == null
        ? widget.initialTitle
        : _scheduleDetailTitle(detail).trim().isEmpty
            ? widget.initialTitle
            : _scheduleDetailTitle(detail);
    final String subtitle = detail == null
        ? widget.initialSubtitle
        : _scheduleDetailSubtitle(detail);
    final _LessonStatus status = detail == null
        ? widget.initialStatus
        : _lessonStatusFromCallStatus(detail.callStatus);
    final bool hasBatchSchedule = detail != null && _hasBatchSchedule(detail);
    final bool deleteDisabled =
        detail == null || _deleteDisabledReason(detail).isNotEmpty || _deleting;
    final String timeLabel =
        detail == null ? '' : _scheduleDateTimeLabel(detail);
    final String headerMeta = <String>[
      if (subtitle.trim().isNotEmpty) subtitle.trim(),
      if (timeLabel.trim().isNotEmpty) timeLabel.trim(),
    ].join(' · ');

    return Material(
      type: MaterialType.transparency,
      child: Center(
        child: Container(
          key: const ValueKey<String>('schedule-detail-dialog'),
          width: 448,
          height: 368,
          decoration: BoxDecoration(
            color: _SmartColors.card,
            borderRadius: BorderRadius.circular(22),
            border: Border.all(color: const Color(0xFFF0E0D3)),
            boxShadow: const <BoxShadow>[
              BoxShadow(
                color: Color(0x24000000),
                blurRadius: 40,
                offset: Offset(0, 18),
              ),
            ],
          ),
          child: Column(
            children: <Widget>[
              Padding(
                padding: const EdgeInsets.fromLTRB(14, 12, 12, 11),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: <Widget>[
                    _ScheduleDetailTypeMark(
                      label: detail == null
                          ? _scheduleTypeLabelByClassType(null)
                          : _scheduleTypeLabelByClassType(detail.classType),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          Row(
                            children: <Widget>[
                              Flexible(
                                fit: FlexFit.loose,
                                child: Text(
                                  title,
                                  maxLines: 1,
                                  overflow: TextOverflow.ellipsis,
                                  style: const TextStyle(
                                    color: _SmartColors.ink,
                                    fontSize: 15.5,
                                    height: 1.15,
                                    fontWeight: FontWeight.w900,
                                  ),
                                ),
                              ),
                              const SizedBox(width: 8),
                              _ScheduleDetailBadge(
                                label: detail == null
                                    ? widget.initialStatus.label
                                    : _callStatusLabel(detail),
                                foreground: status.foreground,
                                background: status.background,
                              ),
                            ],
                          ),
                          if (headerMeta.isNotEmpty) ...<Widget>[
                            const SizedBox(height: 4),
                            Text(
                              headerMeta,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                color: _SmartColors.text,
                                fontSize: 11.5,
                                height: 1.15,
                                fontWeight: FontWeight.w700,
                              ),
                            ),
                          ],
                        ],
                      ),
                    ),
                    const SizedBox(width: 8),
                    _ScheduleDetailIconButton(
                      key: const ValueKey<String>('schedule-detail-close'),
                      icon: Icons.close_rounded,
                      onTap: widget.onClose,
                    ),
                  ],
                ),
              ),
              const Divider(height: 1, color: Color(0xFFF0E0D3)),
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(12, 10, 12, 0),
                  child: _buildBody(),
                ),
              ),
              const Divider(height: 1, color: Color(0xFFF0E0D3)),
              Padding(
                padding: const EdgeInsets.fromLTRB(12, 9, 12, 11),
                child: Row(
                  children: <Widget>[
                    Expanded(
                      child: SizedBox(
                        height: 18,
                        child: Align(
                          alignment: Alignment.centerLeft,
                          child: Text(
                            detail == null
                                ? (_loading ? '' : (_loadError ?? ''))
                                : (_actionError ??
                                    _deleteDisabledReason(detail)),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: const TextStyle(
                              color: _SmartColors.danger,
                              fontSize: 11,
                              fontWeight: FontWeight.w700,
                            ),
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    if (hasBatchSchedule) ...<Widget>[
                      _ScheduleDetailActionButton(
                        key: const ValueKey<String>(
                          'schedule-detail-delete-current',
                        ),
                        label: _deleting ? '删除中...' : '删除本节',
                        icon: Icons.delete_outline_rounded,
                        filled: true,
                        enabled: !deleteDisabled,
                        onTap: () => _handleDelete(ScheduleDeleteScope.current),
                      ),
                      const SizedBox(width: 8),
                      _ScheduleDetailActionButton(
                        key: const ValueKey<String>(
                          'schedule-detail-delete-future',
                        ),
                        label: _deleting ? '删除中...' : '删后续',
                        icon: Icons.delete_sweep_rounded,
                        filled: false,
                        enabled: !deleteDisabled,
                        onTap: () => _handleDelete(ScheduleDeleteScope.future),
                      ),
                    ] else
                      _ScheduleDetailActionButton(
                        key: const ValueKey<String>(
                          'schedule-detail-delete-current',
                        ),
                        label: _deleting ? '删除中...' : '删除本节',
                        icon: Icons.delete_outline_rounded,
                        filled: true,
                        enabled: !deleteDisabled,
                        onTap: () => _handleDelete(ScheduleDeleteScope.current),
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

  Widget _buildBody() {
    if (_loading) {
      return const _ScheduleDetailLoadingBody();
    }
    if (_loadError != null) {
      return _ScheduleDetailErrorBody(
        message: _loadError!,
        onRetry: _loadDetail,
      );
    }
    final TimetableScheduleDetail detail = _detail!;
    final List<_ScheduleDetailFieldItem> fields = <_ScheduleDetailFieldItem>[
      _ScheduleDetailFieldItem(
        label: '上课教师',
        value: _fallbackLabel(detail.teacherName, '-'),
      ),
      _ScheduleDetailFieldItem(
        label: '课程',
        value: _fallbackLabel(detail.lessonName, '-'),
      ),
      _ScheduleDetailFieldItem(
        label: '上课助教',
        value: _assistantNamesText(detail),
      ),
      _ScheduleDetailFieldItem(
        label: '上课学员',
        value: _studentSummaryText(_regularStudents(detail)),
      ),
      _ScheduleDetailFieldItem(
        label: '试听学员',
        value: _studentSummaryText(_trialStudents(detail)),
      ),
      _ScheduleDetailFieldItem(
        label: '请假学员',
        value: _studentSummaryText(detail.leaveStudents),
      ),
      _ScheduleDetailFieldItem(
        label: '对内备注',
        value: detail.remark.trim().isEmpty ? '-' : detail.remark.trim(),
        multiline: true,
      ),
    ];
    return SingleChildScrollView(
      physics: const BouncingScrollPhysics(),
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.fromLTRB(14, 10, 14, 10),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: const Color(0xFFF0E0D3)),
        ),
        child: Column(
          children: <Widget>[
            for (int index = 0; index < fields.length; index += 1) ...<Widget>[
              _ScheduleDetailFieldRow(item: fields[index]),
              if (index != fields.length - 1)
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 6),
                  child: Divider(height: 1, color: Color(0xFFF2E5DA)),
                ),
            ],
          ],
        ),
      ),
    );
  }
}

class _ScheduleDeleteConfirmDialog extends StatelessWidget {
  const _ScheduleDeleteConfirmDialog({
    required this.scope,
    required this.onCancel,
    required this.onConfirm,
  });

  final ScheduleDeleteScope scope;
  final VoidCallback onCancel;
  final VoidCallback onConfirm;

  @override
  Widget build(BuildContext context) {
    final bool deleteFuture = scope == ScheduleDeleteScope.future;
    return Material(
      type: MaterialType.transparency,
      child: Center(
        child: Container(
          width: 360,
          padding: const EdgeInsets.fromLTRB(18, 17, 18, 16),
          decoration: BoxDecoration(
            color: _SmartColors.card,
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: const Color(0xFFF0E0D3)),
            boxShadow: const <BoxShadow>[
              BoxShadow(
                color: Color(0x24000000),
                blurRadius: 32,
                offset: Offset(0, 16),
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Row(
                children: <Widget>[
                  Container(
                    width: 36,
                    height: 36,
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFEEE9),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Icon(
                      Icons.delete_outline_rounded,
                      color: _SmartColors.danger,
                      size: 21,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Text(
                      deleteFuture ? '删除后续全部日程' : '删除当前日程',
                      style: const TextStyle(
                        color: _SmartColors.ink,
                        fontSize: 17,
                        height: 1.1,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Text(
                deleteFuture
                    ? '当前日程及其后续关联日程会一起删除，删除后不可恢复。'
                    : '仅删除当前这节课，删除后不可恢复。',
                style: const TextStyle(
                  color: _SmartColors.text,
                  fontSize: 12.5,
                  height: 1.45,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 15),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: <Widget>[
                  _ScheduleDetailActionButton(
                    key: const ValueKey<String>(
                      'schedule-delete-confirm-cancel',
                    ),
                    label: '取消',
                    icon: Icons.close_rounded,
                    filled: false,
                    onTap: onCancel,
                  ),
                  const SizedBox(width: 10),
                  _ScheduleDetailActionButton(
                    key: const ValueKey<String>(
                      'schedule-delete-confirm-submit',
                    ),
                    label: '确认删除',
                    icon: Icons.delete_outline_rounded,
                    filled: true,
                    danger: true,
                    onTap: onConfirm,
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

class _ScheduleDetailBadge extends StatelessWidget {
  const _ScheduleDetailBadge({
    required this.label,
    required this.foreground,
    required this.background,
  });

  final String label;
  final Color foreground;
  final Color background;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 24,
      padding: const EdgeInsets.symmetric(horizontal: 9),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: foreground,
          fontSize: 10.5,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _ScheduleDetailTypeMark extends StatelessWidget {
  const _ScheduleDetailTypeMark({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 42,
      height: 30,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF1E8),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        label,
        textAlign: TextAlign.center,
        style: const TextStyle(
          color: _SmartColors.orangeDeep,
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _ScheduleDetailFieldItem {
  const _ScheduleDetailFieldItem({
    required this.label,
    required this.value,
    this.multiline = false,
  });

  final String label;
  final String value;
  final bool multiline;
}

class _ScheduleDetailFieldRow extends StatelessWidget {
  const _ScheduleDetailFieldRow({required this.item});

  final _ScheduleDetailFieldItem item;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment:
          item.multiline ? CrossAxisAlignment.start : CrossAxisAlignment.center,
      children: <Widget>[
        SizedBox(
          width: 78,
          child: Text(
            '${item.label}：',
            style: const TextStyle(
              color: _SmartColors.muted,
              fontSize: 12.5,
              height: 1.35,
              fontWeight: FontWeight.w700,
            ),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            item.value,
            maxLines: item.multiline ? 2 : 1,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: _SmartColors.ink,
              fontSize: 12.8,
              height: 1.35,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
      ],
    );
  }
}

class _ScheduleDetailIconButton extends StatelessWidget {
  const _ScheduleDetailIconButton({
    required this.icon,
    required this.onTap,
    super.key,
  });

  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: const Color(0xFFFFF6ED),
      borderRadius: BorderRadius.circular(10),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: SizedBox(
          width: 32,
          height: 32,
          child: Icon(icon, color: _SmartColors.text, size: 18),
        ),
      ),
    );
  }
}

class _ScheduleDetailActionButton extends StatelessWidget {
  const _ScheduleDetailActionButton({
    required this.label,
    required this.icon,
    required this.filled,
    required this.onTap,
    this.enabled = true,
    this.danger = false,
    super.key,
  });

  final String label;
  final IconData icon;
  final bool filled;
  final bool enabled;
  final bool danger;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color activeFill =
        danger ? _SmartColors.danger : _SmartColors.orangeDeep;
    final Color background = !enabled
        ? const Color(0xFFF5EFE9)
        : (filled ? activeFill : Colors.white);
    final Color borderColor = filled
        ? Colors.transparent
        : (!enabled ? const Color(0xFFE7D9CD) : const Color(0xFFE2D1C3));
    final Color textColor = !enabled
        ? _SmartColors.muted
        : (filled
            ? Colors.white
            : (danger ? _SmartColors.danger : _SmartColors.text));
    return Opacity(
      opacity: enabled ? 1 : .72,
      child: IgnorePointer(
        ignoring: !enabled,
        child: Material(
          color: Colors.transparent,
          borderRadius: BorderRadius.circular(10),
          child: InkWell(
            onTap: onTap,
            borderRadius: BorderRadius.circular(10),
            child: Container(
              height: 34,
              padding: const EdgeInsets.symmetric(horizontal: 12),
              decoration: BoxDecoration(
                color: background,
                borderRadius: BorderRadius.circular(10),
                border: Border.all(color: borderColor),
                boxShadow: filled && enabled
                    ? <BoxShadow>[
                        BoxShadow(
                          color: (danger
                                  ? _SmartColors.danger
                                  : _SmartColors.orangeDeep)
                              .withOpacity(.18),
                          blurRadius: 14,
                          offset: const Offset(0, 8),
                        ),
                      ]
                    : null,
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: <Widget>[
                  Icon(icon, color: textColor, size: 16),
                  const SizedBox(width: 5),
                  Text(
                    label,
                    style: TextStyle(
                      color: textColor,
                      fontSize: 12.5,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _ScheduleDetailLoadingBody extends StatelessWidget {
  const _ScheduleDetailLoadingBody();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(14, 10, 14, 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFF0E0D3)),
      ),
      child: Column(
        children: <Widget>[
          for (int index = 0; index < 7; index += 1) ...<Widget>[
            Row(
              children: const <Widget>[
                _ScheduleDetailSkeletonBox(width: 62, height: 13),
                SizedBox(width: 24),
                Expanded(
                  child: _ScheduleDetailSkeletonBox(height: 13),
                ),
              ],
            ),
            if (index != 6)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 7),
                child: Divider(height: 1, color: Color(0xFFF2E5DA)),
              ),
          ],
        ],
      ),
    );
  }
}

class _ScheduleDetailSkeletonBox extends StatelessWidget {
  const _ScheduleDetailSkeletonBox({
    this.width = double.infinity,
    required this.height,
  });

  final double width;
  final double height;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: const Color(0xFFF6EEE7),
        borderRadius: BorderRadius.circular(10),
      ),
    );
  }
}

class _ScheduleDetailErrorBody extends StatelessWidget {
  const _ScheduleDetailErrorBody({
    required this.message,
    required this.onRetry,
  });

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Container(
            width: 54,
            height: 54,
            decoration: BoxDecoration(
              color: const Color(0xFFFFF0E7),
              borderRadius: BorderRadius.circular(18),
            ),
            child: const Icon(
              Icons.error_outline_rounded,
              color: _SmartColors.orangeDeep,
              size: 30,
            ),
          ),
          const SizedBox(height: 14),
          Text(
            message,
            textAlign: TextAlign.center,
            style: const TextStyle(
              color: _SmartColors.text,
              fontSize: 13,
              height: 1.5,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 16),
          _ScheduleDetailActionButton(
            label: '重新加载',
            icon: Icons.refresh_rounded,
            filled: true,
            onTap: onRetry,
          ),
        ],
      ),
    );
  }
}

String _scheduleDetailTitle(TimetableScheduleDetail detail) {
  final String lessonName = detail.lessonName.trim();
  if (lessonName.isNotEmpty) {
    return lessonName;
  }
  final String className = detail.teachingClassName.trim();
  if (className.isNotEmpty) {
    return className;
  }
  return '日程详情';
}

String _scheduleDetailSubtitle(TimetableScheduleDetail detail) {
  final List<TimetableScheduleDetailStudent> students =
      _regularStudents(detail);
  if (students.isNotEmpty) {
    final String name = students.first.studentName.trim();
    if (name.isNotEmpty) {
      return name;
    }
  }
  final String className = detail.teachingClassName.trim();
  if (className.isNotEmpty) {
    return className;
  }
  final String fallbackStudent = detail.students
      .map((TimetableScheduleDetailStudent item) => item.studentName.trim())
      .firstWhere(
        (String item) => item.isNotEmpty,
        orElse: () => '',
      );
  return fallbackStudent;
}

String _scheduleDateTimeLabel(TimetableScheduleDetail detail) {
  final String dateLabel = _monthDayLabel(detail.lessonDate);
  final String weekLabel = _weekdayShortLabel(detail.lessonDate);
  final String startTime = _timeClockText(detail.startAt);
  final String endTime = _timeClockText(detail.endAt);
  final String timeText = startTime.isEmpty && endTime.isEmpty
      ? ''
      : (endTime.isEmpty ? startTime : '$startTime - $endTime');
  return '$dateLabel${weekLabel.isEmpty ? '' : ' $weekLabel'}${timeText.isEmpty ? '' : '  $timeText'}';
}

String _timeClockText(String raw) {
  final String trimmed = raw.trim();
  if (trimmed.isEmpty) {
    return '';
  }
  final DateTime? parsed = DateTime.tryParse(trimmed);
  if (parsed != null) {
    final DateTime local = parsed.isUtc ? parsed.toLocal() : parsed;
    return '${local.hour.toString().padLeft(2, '0')}:${local.minute.toString().padLeft(2, '0')}';
  }
  final RegExpMatch? match =
      RegExp(r'(\d{2}:\d{2})(?::\d{2})?$').firstMatch(trimmed);
  return match?.group(1) ?? trimmed;
}

String _assistantNamesText(TimetableScheduleDetail detail) {
  final List<String> names = detail.assistantNames
      .map((String item) => item.trim())
      .where((String item) => item.isNotEmpty)
      .toList();
  return names.isEmpty ? '-' : names.join('、');
}

String _fallbackLabel(String value, String fallback) {
  final String trimmed = value.trim();
  return trimmed.isEmpty ? fallback : trimmed;
}

List<TimetableScheduleDetailStudent> _regularStudents(
  TimetableScheduleDetail detail,
) {
  return detail.students.where((TimetableScheduleDetailStudent item) {
    return item.scheduleStudentType != 3;
  }).toList();
}

List<TimetableScheduleDetailStudent> _trialStudents(
  TimetableScheduleDetail detail,
) {
  return detail.students.where((TimetableScheduleDetailStudent item) {
    return item.scheduleStudentType == 3;
  }).toList();
}

String _studentSummaryText(List<TimetableScheduleDetailStudent> students) {
  if (students.isEmpty) {
    return '-';
  }
  final List<String> names = students
      .map((TimetableScheduleDetailStudent item) => item.studentName.trim())
      .where((String item) => item.isNotEmpty)
      .toList();
  if (names.isEmpty) {
    return '${students.length}人';
  }
  return '${students.length}人，${_studentNamesPreviewText(names)}';
}

String _studentNamesPreviewText(List<String> names) {
  if (names.length <= 2) {
    return names.join('、');
  }
  final List<String> preview = names.take(2).toList();
  return '${preview.join('、')} +${names.length - preview.length}';
}

bool _hasBatchSchedule(TimetableScheduleDetail detail) {
  if (detail.batchSize > 1 || detail.batchNo.trim().isNotEmpty) {
    return true;
  }
  final TimetableScheduleBatchMeta? meta = detail.batchMeta;
  if (meta == null) {
    return false;
  }
  return meta.plannedClassCount > 1 ||
      meta.freeSelectedDates.length > 1 ||
      meta.schedulingMode.trim() == 'free' ||
      (meta.schedulingMode.trim() == 'repeat' &&
          meta.repeatRule.trim().isNotEmpty &&
          meta.repeatRule.trim() != 'none');
}

String _deleteDisabledReason(TimetableScheduleDetail detail) {
  if (detail.callStatus == 2) {
    return '当前日程已点名，不可删除';
  }
  return '';
}

String _callStatusLabel(TimetableScheduleDetail detail) {
  final String text = detail.callStatusText.trim();
  return text.isNotEmpty ? text : _callStatusTextByCode(detail.callStatus);
}

String _callStatusTextByCode(int callStatus) {
  switch (callStatus) {
    case 2:
      return '已点名';
    case 3:
      return '部分点名';
    default:
      return '未点名';
  }
}

_LessonStatus _lessonStatusFromCallStatus(int callStatus) {
  switch (callStatus) {
    case 2:
      return _LessonStatus.signed;
    case 3:
      return _LessonStatus.partial;
    default:
      return _LessonStatus.unsigned;
  }
}

String _scheduleTypeLabelByClassType(int? classType) {
  return classType == 1 ? '班课' : '1v1';
}
