part of 'pep3_assessment_page.dart';

class _Pep3RightRail extends StatelessWidget {
  const _Pep3RightRail({
    required this.progressPercent,
    required this.answered,
    required this.total,
    required this.missing,
    required this.currentItemNo,
    required this.recordFields,
    required this.recordValues,
    required this.caregiverInvite,
    required this.caregiverLoading,
    required this.onRecordValue,
    required this.onSmsTap,
    required this.onWechatTap,
  });

  final int progressPercent;
  final int answered;
  final int total;
  final int missing;
  final int currentItemNo;
  final List<Pep3RecordField> recordFields;
  final Map<String, dynamic> recordValues;
  final Pep3CaregiverInvite? caregiverInvite;
  final bool caregiverLoading;
  final void Function(int itemNo, String key, dynamic value) onRecordValue;
  final VoidCallback onSmsTap;
  final VoidCallback onWechatTap;

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: EdgeInsets.zero,
      physics: const BouncingScrollPhysics(),
      children: <Widget>[
        _RailCard(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: _ProgressPanel(
              progressPercent: progressPercent,
              answered: answered,
              total: total,
              missing: missing,
            ),
          ),
        ),
        const SizedBox(height: 10),
        _RailCard(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: _TrainingRecordPanel(
              currentItemNo: currentItemNo,
              fields: recordFields,
              values: recordValues,
              onChanged: (String key, dynamic value) =>
                  onRecordValue(currentItemNo, key, value),
            ),
          ),
        ),
        const SizedBox(height: 10),
        _RailCard(
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: _CaregiverPanel(
              invite: caregiverInvite,
              loading: caregiverLoading,
              onSmsTap: onSmsTap,
              onWechatTap: onWechatTap,
            ),
          ),
        ),
      ],
    );
  }
}

class _ProgressPanel extends StatelessWidget {
  const _ProgressPanel({
    required this.progressPercent,
    required this.answered,
    required this.total,
    required this.missing,
  });

  final int progressPercent;
  final int answered;
  final int total;
  final int missing;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        const Text(
          '当前进度',
          style: TextStyle(
            color: _Pep3Colors.ink,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 14),
        Row(
          children: <Widget>[
            SizedBox(
              width: 82,
              height: 82,
              child: CustomPaint(
                painter: _DonutPainter(percent: progressPercent),
                child: Center(
                  child: Text(
                    '$progressPercent%',
                    style: const TextStyle(
                      color: _Pep3Colors.ink,
                      fontSize: 21,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  _ProgressText(label: '已完成', value: '$answered / $total 题'),
                  const SizedBox(height: 10),
                  _ProgressText(
                    label: '缺题',
                    value: '$missing 题',
                    danger: true,
                  ),
                ],
              ),
            ),
          ],
        ),
      ],
    );
  }
}

class _TrainingRecordPanel extends StatelessWidget {
  const _TrainingRecordPanel({
    required this.currentItemNo,
    required this.fields,
    required this.values,
    required this.onChanged,
  });

  final int currentItemNo;
  final List<Pep3RecordField> fields;
  final Map<String, dynamic> values;
  final void Function(String key, dynamic value) onChanged;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        const Text(
          '儿童训练记录',
          style: TextStyle(
            color: _Pep3Colors.ink,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 14),
        if (fields.isEmpty)
          Container(
            height: 50,
            alignment: Alignment.centerLeft,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            decoration: BoxDecoration(
              color: const Color(0xFFFFFAF5),
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: _Pep3Colors.lineSoft),
            ),
            child: const Text(
              '本题暂无训练记录项',
              style: TextStyle(
                color: _Pep3Colors.muted,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          )
        else
          for (final Pep3RecordField field in fields)
            _RecordFieldEditor(
              key: ValueKey<String>(
                'pep3-record-field-$currentItemNo-${field.key}',
              ),
              currentItemNo: currentItemNo,
              field: field,
              value: values[field.key],
              onChanged: (dynamic value) => onChanged(field.key, value),
            ),
      ],
    );
  }
}

class _CaregiverPanel extends StatelessWidget {
  const _CaregiverPanel({
    required this.invite,
    required this.loading,
    required this.onSmsTap,
    required this.onWechatTap,
  });

  final Pep3CaregiverInvite? invite;
  final bool loading;
  final VoidCallback onSmsTap;
  final VoidCallback onWechatTap;

  @override
  Widget build(BuildContext context) {
    final String qrValue = invite?.qrValue ?? '';
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        const Text(
          '照护者报告',
          style: TextStyle(
            color: _Pep3Colors.ink,
            fontSize: 16,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(height: 14),
        Center(
          child: Container(
            width: 132,
            height: 132,
            alignment: Alignment.center,
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: _Pep3Colors.line),
            ),
            child: loading
                ? const CircularProgressIndicator(color: _Pep3Colors.orange)
                : _CaregiverQr(invite: invite, qrValue: qrValue),
          ),
        ),
        const SizedBox(height: 10),
        const Center(
          child: Text(
            '家长扫码填写照护者报告',
            style: TextStyle(
              color: _Pep3Colors.muted,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
        const SizedBox(height: 14),
        _CaregiverActionButton(
          label: '发送短信给家长',
          icon: Icons.sms_outlined,
          filled: false,
          loading: loading,
          onTap: onSmsTap,
        ),
        const SizedBox(height: 8),
        _CaregiverActionButton(
          label: '推送微信消息',
          icon: Icons.wechat,
          filled: true,
          loading: loading,
          onTap: onWechatTap,
        ),
      ],
    );
  }
}

class _CaregiverQr extends StatefulWidget {
  const _CaregiverQr({required this.invite, required this.qrValue});

  final Pep3CaregiverInvite? invite;
  final String qrValue;

  @override
  State<_CaregiverQr> createState() => _CaregiverQrState();
}

class _CaregiverQrState extends State<_CaregiverQr> {
  String _signature = '';
  Widget _cachedQr = const SizedBox.shrink();

  @override
  void initState() {
    super.initState();
    _cacheQr();
  }

  @override
  void didUpdateWidget(covariant _CaregiverQr oldWidget) {
    super.didUpdateWidget(oldWidget);
    final String nextSignature = _qrSignature(widget.invite, widget.qrValue);
    if (nextSignature != _signature) {
      _cacheQr();
    }
  }

  void _cacheQr() {
    final String signature = _qrSignature(widget.invite, widget.qrValue);
    _signature = signature;
    _cachedQr = _buildQr(widget.invite, widget.qrValue, signature);
  }

  @override
  Widget build(BuildContext context) {
    return RepaintBoundary(child: _cachedQr);
  }

  static String _qrSignature(Pep3CaregiverInvite? invite, String qrValue) {
    final String dataUrl = invite?.miniProgramCodeDataUrl.trim() ?? '';
    if (dataUrl.isNotEmpty) {
      return 'data:$dataUrl';
    }
    if (qrValue.trim().isNotEmpty) {
      return 'qr:${qrValue.trim()}';
    }
    return 'empty';
  }

  static Widget _buildQr(
    Pep3CaregiverInvite? invite,
    String qrValue,
    String signature,
  ) {
    final String dataUrl = invite?.miniProgramCodeDataUrl ?? '';
    final RegExpMatch? match =
        RegExp(r'^data:image/[^;]+;base64,(.+)$').firstMatch(dataUrl);
    if (match != null) {
      return Image.memory(
        base64Decode(match.group(1)!),
        key: ValueKey<String>('caregiver-image-$signature'),
        width: 116,
        height: 116,
        fit: BoxFit.contain,
        gaplessPlayback: true,
      );
    }
    if (qrValue.isNotEmpty) {
      return QrImageView(
        key: ValueKey<String>('caregiver-qr-$signature'),
        data: qrValue,
        version: QrVersions.auto,
        padding: EdgeInsets.zero,
        backgroundColor: Colors.white,
      );
    }
    return const Text(
      '暂无二维码',
      textAlign: TextAlign.center,
      style: TextStyle(
        color: _Pep3Colors.muted,
        fontSize: 13,
        fontWeight: FontWeight.w800,
      ),
    );
  }
}
