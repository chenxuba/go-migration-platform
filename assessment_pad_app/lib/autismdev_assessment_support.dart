part of 'autismdev_assessment_page.dart';

class _AutismDevStateBody extends StatelessWidget {
  const _AutismDevStateBody({
    required this.title,
    required this.message,
    required this.actionText,
    required this.onAction,
  });

  final String title;
  final String message;
  final String actionText;
  final VoidCallback onAction;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        width: 420,
        padding: const EdgeInsets.all(28),
        decoration: _panelDecoration(),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Text(
              title,
              style: const TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 24,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 12),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _AutismDevColors.body,
                fontSize: 16,
                height: 1.45,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 22),
            FilledButton(
              onPressed: onAction,
              child: Text(actionText),
            ),
          ],
        ),
      ),
    );
  }
}

class _AutismDevColors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color ink = Color(0xFF432B22);
  static const Color body = Color(0xFF7F665A);
  static const Color muted = Color(0xFFBBA99C);
  static const Color line = Color(0xFFF0DACB);
  static const Color lineSoft = Color(0xFFF6E7DC);
  static const Color softPanel = Color(0xFFFFFBF4);
  static const Color blue = Color(0xFF3F82D2);
  static const Color green = Color(0xFF6F9F70);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95735);
  static const Color red = Color(0xFFD94A42);
  static const Color teal = Color(0xFF00A7A7);
  static const Color purple = Color(0xFF7C5CFC);
  static const Color pink = Color(0xFFD0578B);
}

List<BoxShadow> _autismDevShadow({
  Color color = const Color(0x18000000),
  double blur = 18,
  Offset offset = const Offset(0, 9),
}) {
  return <BoxShadow>[
    BoxShadow(color: color, blurRadius: blur, offset: offset),
  ];
}

BoxDecoration _panelDecoration() {
  return BoxDecoration(
    color: Colors.white.withOpacity(.92),
    border: Border.all(color: _AutismDevColors.line),
    borderRadius: BorderRadius.circular(8),
    boxShadow: _autismDevShadow(color: const Color(0x12B05F32), blur: 15),
  );
}

Color _domainColor(String code) {
  switch (code.toUpperCase()) {
    case 'GM':
      return _AutismDevColors.green;
    case 'FM':
      return _AutismDevColors.orange;
    case 'LC':
      return _AutismDevColors.purple;
    case 'COG':
      return _AutismDevColors.teal;
    case 'SOC':
      return _AutismDevColors.pink;
    case 'ADL':
      return const Color(0xFF6A7A20);
    case 'EB':
      return _AutismDevColors.red;
    case 'SP':
    default:
      return _AutismDevColors.blue;
  }
}

Color _scoreColor(String score) {
  switch (score.toUpperCase()) {
    case 'P':
    case 'A':
      return _AutismDevColors.green;
    case 'E':
    case 'M':
      return _AutismDevColors.orange;
    case 'F':
    case 'S':
      return _AutismDevColors.red;
    case 'X':
    default:
      return _AutismDevColors.muted;
  }
}

String _optionTitle(AutismDevScoreOption option) {
  final String label = option.label.trim();
  if (label.isEmpty) {
    return option.value;
  }
  final String prefix = option.value.trim();
  if (prefix.isNotEmpty && label.startsWith(prefix)) {
    return label.substring(prefix.length).trim();
  }
  return label;
}

String _displayItemTitle(AutismDevItemSummary item) {
  return item.itemTitle.trim().isNotEmpty
      ? item.itemTitle.trim()
      : item.testItem.trim();
}

String _autismDevScaleTitle(String raw) {
  final String title = raw.trim().isEmpty ? '孤独症儿童发展评估' : raw.trim();
  final String cleaned =
      title.replaceFirst(RegExp(r'\s*[（(]?\s*试行\s*[）)]?\s*$'), '').trim();
  return cleaned.isEmpty ? title : cleaned;
}

String _assessmentRangeText(
  AutismDevItemSummary item,
  AutismDevAssessmentItem? detail,
) {
  final String range = (detail?.assessmentRange ?? item.assessmentRange).trim();
  return range.isEmpty ? '-' : range;
}

String _assessmentRangeBucket(AutismDevItemSummary item) {
  final List<String> parts = item.assessmentRange
      .split('/')
      .map((String part) => part.trim())
      .where((String part) => part.isNotEmpty)
      .toList(growable: false);
  if (parts.length >= 2) {
    return parts[1];
  }
  if (parts.isNotEmpty) {
    return parts.first;
  }
  return '未分类';
}

List<_AutismDevRangeOption> _rangeOptionsForGroup(
  AutismDevDomainGroup? group,
  Map<int, String> itemScores,
) {
  if (group == null) {
    return const <_AutismDevRangeOption>[];
  }
  final Map<String, _AutismDevRangeOption> optionByLabel =
      <String, _AutismDevRangeOption>{};
  for (final AutismDevItemSummary item in group.items) {
    final String label = _assessmentRangeBucket(item);
    final _AutismDevRangeOption option =
        optionByLabel.putIfAbsent(label, () => _AutismDevRangeOption(label));
    option.total += 1;
    if (itemScores.containsKey(item.itemNo)) {
      option.done += 1;
    }
  }
  return optionByLabel.values.toList(growable: false);
}

String _detailText(String? preferred, [String fallback = '']) {
  final String value = (preferred ?? '').trim();
  if (value.isNotEmpty) {
    return value;
  }
  final String fallbackValue = fallback.trim();
  return fallbackValue.isNotEmpty ? fallbackValue : '-';
}

String _nonEmptyDetailText(String? preferred, [String fallback = '']) {
  final String value = (preferred ?? '').trim();
  if (value.isNotEmpty) {
    return value;
  }
  return fallback.trim();
}

List<String> _criteriaDisplayLines(String? value) {
  final String text = _detailText(value);
  if (text == '-') {
    return const <String>[];
  }
  return text
      .split(RegExp(r'[\r\n]+'))
      .map((String line) => line.trim())
      .where((String line) => line.isNotEmpty)
      .toList();
}

String _criteriaDisplayText(String? value) {
  final List<String> lines = _criteriaDisplayLines(value);
  return lines.isEmpty ? '-' : lines.join('；');
}

bool _isEmotionBehaviorItem(AutismDevItemSummary item) {
  return item.domainCode.trim().toUpperCase() == 'EB' ||
      item.scoreType.trim().toUpperCase() == 'AMS';
}

int? _resolvedStudentAgeMonths({
  required String birthDate,
  required String assessmentDate,
  required String ageText,
}) {
  final int? dateAge = _ageMonthsFromDates(birthDate, assessmentDate);
  if (dateAge != null) {
    return dateAge;
  }
  return _ageMonthsFromText(ageText);
}

int? _ageMonthsFromDates(String birthDate, String assessmentDate) {
  final DateTime? birth = DateTime.tryParse(_dateOnlyText(birthDate));
  final DateTime? target = DateTime.tryParse(_dateOnlyText(assessmentDate));
  if (birth == null || target == null || birth.isAfter(target)) {
    return null;
  }
  int months = (target.year - birth.year) * 12 + target.month - birth.month;
  if (target.day < birth.day) {
    months -= 1;
  }
  return math.max(months, 0);
}

int? _ageMonthsFromText(String ageText) {
  final String text = ageText.trim();
  if (text.isEmpty || text == '未知') {
    return null;
  }
  final RegExpMatch? monthMatch = RegExp(r'(\d+)\s*(?:个)?月').firstMatch(text);
  if (monthMatch != null) {
    return int.tryParse(monthMatch.group(1)!);
  }
  final RegExpMatch? yearMonthMatch =
      RegExp(r'(\d+)\s*岁\s*(\d+)?').firstMatch(text);
  if (yearMonthMatch != null) {
    final int years = int.tryParse(yearMonthMatch.group(1)!) ?? 0;
    final int months = int.tryParse(yearMonthMatch.group(2) ?? '') ?? 0;
    return years * 12 + months;
  }
  return int.tryParse(text);
}

String _dateOnlyText(String value) {
  final String text = value.trim();
  if (text.length >= 10) {
    return text.substring(0, 10);
  }
  return text;
}

String _todayIsoDate() {
  final DateTime now = DateTime.now();
  return '${now.year.toString().padLeft(4, '0')}-'
      '${now.month.toString().padLeft(2, '0')}-'
      '${now.day.toString().padLeft(2, '0')}';
}

String _formatClock(DateTime value) {
  return '${value.hour.toString().padLeft(2, '0')}:'
      '${value.minute.toString().padLeft(2, '0')}';
}

String _formatAutismDevDateTime(String value) {
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
