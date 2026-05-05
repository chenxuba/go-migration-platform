import 'dart:collection';
import 'dart:math' as math;

class ChineseImeEditingValue {
  const ChineseImeEditingValue({
    this.text = '',
    this.composing = '',
    this.preview = '',
    this.candidates = const <String>[],
  });

  final String text;
  final String composing;
  final String preview;
  final List<String> candidates;
}

class ChineseImeDictionary {
  const ChineseImeDictionary(this.entries);

  final Map<String, List<String>> entries;

  List<String> candidates(String value, {int limit = 8}) {
    final String pinyin =
        value.toLowerCase().replaceAll(RegExp(r'[^a-z0-9]'), '');
    if (pinyin.isEmpty) {
      return const <String>[];
    }

    final LinkedHashSet<String> result = LinkedHashSet<String>();
    for (final String item in entries[pinyin] ?? const <String>[]) {
      if (item.trim().isNotEmpty) {
        result.add(item);
      }
    }
    for (final MapEntry<String, List<String>> entry in entries.entries) {
      if (entry.key.startsWith(pinyin)) {
        for (final String item in entry.value) {
          if (item.trim().isNotEmpty) {
            result.add(item);
          }
        }
      }
    }

    return result.take(limit).toList();
  }
}

class ChineseImeEngine {
  const ChineseImeEngine({required this.dictionary});

  final ChineseImeDictionary dictionary;

  ChineseImeEditingValue replace(String text) {
    return ChineseImeEditingValue(text: text);
  }

  ChineseImeEditingValue handleKey(
    ChineseImeEditingValue value,
    String key,
  ) {
    if (_isAsciiLetter(key)) {
      return _withComposing(value, value.composing + key.toLowerCase());
    }

    final int? digit = int.tryParse(key);
    if (value.composing.isNotEmpty && key.length == 1 && digit != null) {
      if (digit >= 1 && digit <= value.candidates.length) {
        return commitCandidate(value, value.candidates[digit - 1]);
      }
      return _withComposing(value, value.composing + key);
    }

    if (key == ' ' && value.composing.isNotEmpty) {
      return commit(value);
    }

    final ChineseImeEditingValue committed = commit(value);
    return ChineseImeEditingValue(text: committed.text + key);
  }

  ChineseImeEditingValue commit(ChineseImeEditingValue value) {
    if (value.composing.isEmpty) {
      return value;
    }
    return ChineseImeEditingValue(text: _prefix(value) + value.preview);
  }

  ChineseImeEditingValue commitCandidate(
    ChineseImeEditingValue value,
    String candidate,
  ) {
    if (value.composing.isEmpty) {
      return ChineseImeEditingValue(text: value.text + candidate);
    }
    return ChineseImeEditingValue(text: _prefix(value) + candidate);
  }

  ChineseImeEditingValue backspace(ChineseImeEditingValue value) {
    if (value.composing.isNotEmpty) {
      final List<int> composingRunes = value.composing.runes.toList();
      return _withComposing(
        value,
        String.fromCharCodes(composingRunes.take(composingRunes.length - 1)),
      );
    }

    if (value.text.isEmpty) {
      return value;
    }
    final List<int> textRunes = value.text.runes.toList();
    return ChineseImeEditingValue(
      text: String.fromCharCodes(textRunes.take(textRunes.length - 1)),
    );
  }

  ChineseImeEditingValue clear() {
    return const ChineseImeEditingValue();
  }

  ChineseImeEditingValue _withComposing(
    ChineseImeEditingValue value,
    String composing,
  ) {
    final String normalized =
        composing.toLowerCase().replaceAll(RegExp(r'[^a-z0-9]'), '');
    final List<String> candidates = dictionary.candidates(normalized);
    final String preview = normalized.isEmpty
        ? ''
        : candidates.isNotEmpty
            ? candidates.first
            : normalized;

    return ChineseImeEditingValue(
      text: _prefix(value) + preview,
      composing: normalized,
      preview: preview,
      candidates: candidates,
    );
  }

  String _prefix(ChineseImeEditingValue value) {
    final int prefixLength =
        math.max(0, value.text.length - value.preview.length);
    return value.text.substring(0, prefixLength);
  }
}

bool _isAsciiLetter(String value) {
  return value.length == 1 && RegExp(r'^[A-Za-z]$').hasMatch(value);
}

const ChineseImeDictionary assessmentScaleImeDictionary =
    ChineseImeDictionary(_assessmentScaleCandidateMap);

const Map<String, List<String>> _assessmentScaleCandidateMap =
    <String, List<String>>{
  'p': <String>['评', 'PEP-3'],
  'pe': <String>['PEP-3'],
  'pep': <String>['PEP-3'],
  'pep3': <String>['PEP-3'],
  'y': <String>['语', '言'],
  'yu': <String>['语', '语言'],
  'yan': <String>['言'],
  'yuyan': <String>['语言'],
  'g': <String>['沟'],
  'go': <String>['沟'],
  'gou': <String>['沟'],
  'tong': <String>['通'],
  'gt': <String>['沟通'],
  'goutong': <String>['沟通'],
  's': <String>['筛', '社'],
  'shai': <String>['筛'],
  'cha': <String>['查'],
  'sc': <String>['筛查'],
  'shaicha': <String>['筛查'],
  'kou': <String>['口'],
  'ky': <String>['口语'],
  'kouyu': <String>['口语'],
  'biao': <String>['表'],
  'da': <String>['达'],
  'bd': <String>['表达'],
  'biaoda': <String>['表达'],
  'she': <String>['社'],
  'jiao': <String>['交'],
  'sj': <String>['社交'],
  'shejiao': <String>['社交'],
  'zong': <String>['综'],
  'he': <String>['合'],
  'zh': <String>['综合'],
  'zonghe': <String>['综合'],
  'fa': <String>['发'],
  'zhan': <String>['展'],
  'fz': <String>['发展'],
  'fazhan': <String>['发展'],
  'ping': <String>['评'],
  'heping': <String>['评核'],
  'ph': <String>['评核'],
  'pinghe': <String>['评核'],
  'liang': <String>['量'],
  'lb': <String>['量表'],
  'liangbiao': <String>['量表'],
  'guan': <String>['观'],
  'gc': <String>['观察'],
  'guancha': <String>['观察'],
  'fu': <String>['复'],
  'fp': <String>['复评'],
  'fuping': <String>['复评'],
  'er': <String>['儿'],
  'et': <String>['儿童'],
  'ertong': <String>['儿童'],
  'li': <String>['理'],
  'jie': <String>['解'],
  'lj': <String>['理解'],
  'lijie': <String>['理解'],
};
