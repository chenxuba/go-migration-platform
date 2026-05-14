part of 'pep3_assessment_page.dart';

class _Pep3ErrorShell extends StatelessWidget {
  const _Pep3ErrorShell({required this.message, required this.onBack});

  final String message;
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: _Pep3Colors.page,
      child: Center(
        child: Container(
          width: 440,
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: _Pep3Colors.line),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(Icons.error_outline_rounded,
                  size: 34, color: _Pep3Colors.orange),
              const SizedBox(height: 12),
              Text(
                message,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: _Pep3Colors.text,
                  fontSize: 14,
                  height: 1.5,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 16),
              _FooterButton(
                label: '返回',
                icon: Icons.chevron_left_rounded,
                enabled: true,
                onTap: onBack,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _Pep3Colors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color ink = Color(0xFF432B22);
  static const Color text = Color(0xFF7F665A);
  static const Color muted = Color(0xFFBBA99C);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95735);
  static const Color green = Color(0xFF6F9F70);
  static const Color blue = Color(0xFF3F82D2);
  static const Color red = Color(0xFFD94A42);
  static const Color line = Color(0xFFF0DACB);
  static const Color lineSoft = Color(0xFFF6E7DC);
}

List<BoxShadow> _pep3Shadow({
  Color color = const Color(0x18000000),
  double blur = 18,
  Offset offset = const Offset(0, 9),
}) {
  return <BoxShadow>[
    BoxShadow(color: color, blurRadius: blur, offset: offset),
  ];
}

Color _scoreColor(int value) {
  if (value == 2) {
    return const Color(0xFF159947);
  }
  if (value == 0) {
    return _Pep3Colors.red;
  }
  return _Pep3Colors.blue;
}

String _shortScoreLabel(int value, String fallback) {
  if (value == 2) {
    return '通过';
  }
  if (value == 1) {
    return '部分通过';
  }
  if (value == 0) {
    return '未通过';
  }
  return fallback.trim();
}

String _scoreStandardText(
  Pep3AssessmentItem item,
  List<Pep3ScoreOption> options,
) {
  final String standard = _normalizeText(item.standard, fallback: '');
  if (standard.isNotEmpty) {
    return standard;
  }
  return options
      .map((Pep3ScoreOption option) =>
          '${option.value} 分（${_shortScoreLabel(option.value, option.label)}）：${option.description.trim().isEmpty ? option.label : option.description}')
      .join('\n');
}

String _normalizeText(String? value, {String fallback = '-'}) {
  final String text = '${value ?? ''}'.replaceAll(RegExp(r'\s+'), ' ').trim();
  return text.isEmpty ? fallback : text;
}

String _todayIsoDate() {
  final DateTime now = DateTime.now();
  return '${now.year.toString().padLeft(4, '0')}-${now.month.toString().padLeft(2, '0')}-${now.day.toString().padLeft(2, '0')}';
}

String? _normalizeDate(String? value) {
  final String text = '${value ?? ''}'.trim();
  if (text.isEmpty) {
    return null;
  }
  final DateTime? parsed = DateTime.tryParse(text);
  if (parsed == null) {
    return text.length >= 10 ? text.substring(0, 10) : text;
  }
  return '${parsed.year.toString().padLeft(4, '0')}-${parsed.month.toString().padLeft(2, '0')}-${parsed.day.toString().padLeft(2, '0')}';
}

String _formatDateTime(String value) {
  final String text = value.trim();
  if (text.isEmpty) {
    return '-';
  }
  final DateTime? parsed = DateTime.tryParse(text);
  if (parsed == null) {
    if (text.length >= 16) {
      return text.substring(0, 16).replaceFirst('T', ' ');
    }
    return text;
  }
  final DateTime local = parsed.toLocal();
  return '${local.year.toString().padLeft(4, '0')}-'
      '${local.month.toString().padLeft(2, '0')}-'
      '${local.day.toString().padLeft(2, '0')} '
      '${local.hour.toString().padLeft(2, '0')}:'
      '${local.minute.toString().padLeft(2, '0')}';
}

String _compactDateLabel(String value) {
  return _normalizeDate(value) ?? value.trim();
}

String _shortDateLabel(String value) {
  final String normalized = _normalizeDate(value) ?? value.trim();
  if (normalized.length >= 10) {
    return normalized.substring(5, 10);
  }
  return normalized;
}

String _assessmentAgeText(String birthDate, String assessmentDate) {
  return formatAssessmentAgeText(
    birthDate: birthDate,
    assessmentDate: assessmentDate,
  );
}

bool _isEmptyRecordValue(dynamic value) {
  if (value == null) {
    return true;
  }
  if (value is String) {
    return value.trim().isEmpty;
  }
  if (value is Iterable) {
    return value.isEmpty;
  }
  return false;
}

extension _FirstOrNull<T> on Iterable<T> {
  T? get firstOrNull {
    final Iterator<T> iterator = this.iterator;
    if (iterator.moveNext()) {
      return iterator.current;
    }
    return null;
  }
}
