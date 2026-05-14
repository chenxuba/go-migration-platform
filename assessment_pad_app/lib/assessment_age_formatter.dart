import 'package:flutter/foundation.dart';

String formatAssessmentAgeText({
  required String birthDate,
  required String assessmentDate,
  String fallback = '',
}) {
  final DateTime? birth = parseAssessmentDateOnly(birthDate);
  final DateTime? target = parseAssessmentDateOnly(assessmentDate);
  if (birth == null || target == null || birth.isAfter(target)) {
    return fallback;
  }

  int years = target.year - birth.year;
  DateTime yearAnchor = DateTime(birth.year + years, birth.month, birth.day);
  if (yearAnchor.isAfter(target)) {
    years -= 1;
    yearAnchor = DateTime(birth.year + years, birth.month, birth.day);
  }

  int months =
      (target.year - yearAnchor.year) * 12 + target.month - yearAnchor.month;
  DateTime monthAnchor =
      DateTime(yearAnchor.year, yearAnchor.month + months, yearAnchor.day);
  if (monthAnchor.isAfter(target)) {
    months -= 1;
    monthAnchor =
        DateTime(yearAnchor.year, yearAnchor.month + months, yearAnchor.day);
  }

  final int days = target.difference(monthAnchor).inDays;
  return formatAssessmentAgeParts(years: years, months: months, days: days);
}

@visibleForTesting
String formatAssessmentAgeParts({
  required int years,
  required int months,
  required int days,
}) {
  if (years > 0) {
    if (months > 0) {
      return '$years岁$months个月';
    }
    if (days > 0) {
      return '$years岁$days天';
    }
    return '$years岁';
  }
  if (months > 0) {
    return '$months个月';
  }
  return '${days.clamp(0, 999)}天';
}

DateTime? parseAssessmentDateOnly(String value) {
  final String trimmed = value.trim();
  if (trimmed.isEmpty || trimmed == '未知') {
    return null;
  }
  final RegExpMatch? match =
      RegExp(r'^(\d{4})-(\d{1,2})-(\d{1,2})').firstMatch(trimmed);
  if (match != null) {
    final int? year = int.tryParse(match.group(1)!);
    final int? month = int.tryParse(match.group(2)!);
    final int? day = int.tryParse(match.group(3)!);
    if (year != null && month != null && day != null) {
      return DateTime(year, month, day);
    }
  }
  final DateTime? parsed = DateTime.tryParse(trimmed);
  if (parsed == null) {
    return null;
  }
  return DateTime(parsed.year, parsed.month, parsed.day);
}
