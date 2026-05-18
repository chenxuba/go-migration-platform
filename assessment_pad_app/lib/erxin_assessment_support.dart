part of 'erxin_assessment_page.dart';

class _RuleRow {
  const _RuleRow({
    required this.label,
    required this.value,
    required this.done,
    this.month,
    this.targetMonths = const <int>[],
    this.selected = false,
  });

  final String label;
  final String value;
  final bool done;
  final int? month;
  final List<int> targetMonths;
  final bool selected;
}

class _DomainProgress {
  const _DomainProgress({required this.answered, required this.total});

  final int answered;
  final int total;
}

class _ErxinColors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color ink = Color(0xFF432B22);
  static const Color body = Color(0xFF7F665A);
  static const Color muted = Color(0xFFBBA99C);
  static const Color line = Color(0xFFF0DACB);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95D37);
  static const Color blue = Color(0xFFE96F43);
  static const Color green = Color(0xFF6F9F70);
  static const Color red = Color(0xFFD94A42);
}

BoxDecoration _erxinPanelDecoration() {
  return BoxDecoration(
    color: Colors.white.withOpacity(.9),
    borderRadius: BorderRadius.circular(10),
    border: Border.all(color: _ErxinColors.line),
    boxShadow: _erxinShadow(
      color: const Color(0x12B05F32),
      blur: 15,
      offset: const Offset(0, 7),
    ),
  );
}

List<BoxShadow> _erxinShadow({
  Color color = const Color(0x16000000),
  double blur = 16,
  Offset offset = const Offset(0, 8),
}) {
  return <BoxShadow>[BoxShadow(color: color, blurRadius: blur, offset: offset)];
}

double _actualAgeMonths(String birthDate, String assessmentDate) {
  final DateTime? birth = DateTime.tryParse(birthDate);
  final DateTime? target = DateTime.tryParse(assessmentDate);
  if (birth == null || target == null || birth.isAfter(target)) {
    return 0;
  }
  final int days = target.difference(birth).inDays;
  return days / 30.0;
}

String _dateOnlyText(String value) {
  final String text = value.trim();
  if (text.isEmpty) {
    return '';
  }
  final RegExpMatch? match =
      RegExp(r'^(\d{4})[-/](\d{1,2})[-/](\d{1,2})').firstMatch(text);
  if (match != null) {
    final String year = match.group(1)!;
    final String month = match.group(2)!.padLeft(2, '0');
    final String day = match.group(3)!.padLeft(2, '0');
    return '$year-$month-$day';
  }
  final DateTime? parsed = DateTime.tryParse(text);
  if (parsed == null) {
    return text;
  }
  return '${parsed.year.toString().padLeft(4, '0')}-'
      '${parsed.month.toString().padLeft(2, '0')}-'
      '${parsed.day.toString().padLeft(2, '0')}';
}

String _formatErxinDateTime(String value) {
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

String _formatClock(DateTime value) {
  return '${value.hour.toString().padLeft(2, '0')}:'
      '${value.minute.toString().padLeft(2, '0')}';
}

String _sessionExaminerName(HomeSession session) {
  final String nickName = session.nickName.trim();
  if (nickName.isNotEmpty) {
    return nickName;
  }
  final String username = session.username.trim();
  if (username.isNotEmpty) {
    return username;
  }
  return session.mobile.trim();
}
