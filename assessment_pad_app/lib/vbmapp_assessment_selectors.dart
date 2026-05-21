part of 'vbmapp_assessment_page.dart';

extension _VbmappAssessmentSelectors on _VbmappAssessmentPageState {
  List<_VbmappItem> get _selectedItems {
    return _itemsForModule(_selectedModuleCode);
  }

  _VbmappItem get _selectedItem {
    final List<_VbmappItem> items = _selectedItems;
    return items[_selectedItemIndex.clamp(0, items.length - 1)];
  }

  int get _answeredCount {
    return _milestoneScores.length +
        _barrierScores.length +
        _transitionScores.length;
  }

  double get _progressPercent {
    return _answeredCount / _totalItemCount;
  }

  _VbmappScoreSnapshot get _scoreSnapshot {
    return _VbmappScoreSnapshot(
      milestoneTotal: _milestoneScores.values.fold<double>(
        0,
        (double total, double score) => total + score,
      ),
      milestoneMax: _milestoneItems.length,
      barrierTotal: _barrierScores.values.fold<int>(
        0,
        (int total, int score) => total + score,
      ),
      barrierMax: _barrierItems.length * 4,
      transitionTotal: _transitionScores.values.fold<int>(
        0,
        (int total, int score) => total + score,
      ),
      transitionMax: _transitionItems.length * 5,
      milestoneDomains: _milestoneDomainSummaries,
    );
  }

  List<_VbmappDomainScoreSummary> get _milestoneDomainSummaries {
    final Map<String, List<_VbmappItem>> groupedItems =
        <String, List<_VbmappItem>>{};
    for (final _VbmappItem item in _milestoneItems) {
      groupedItems
          .putIfAbsent(item.domainName, () => <_VbmappItem>[])
          .add(item);
    }
    return groupedItems.entries
        .map((MapEntry<String, List<_VbmappItem>> entry) {
      final List<_VbmappItem> items = entry.value;
      final int answered = items
          .where(
              (_VbmappItem item) => _milestoneScores.containsKey(item.itemCode))
          .length;
      final double score = items.fold<double>(
        0,
        (double total, _VbmappItem item) =>
            total + (_milestoneScores[item.itemCode] ?? 0),
      );
      return _VbmappDomainScoreSummary(
        name: entry.key,
        score: score,
        maxScore: items.length,
        answered: answered,
        total: items.length,
      );
    }).toList(growable: false);
  }

  String get _studentAgeText {
    final String fallback = _studentAge.trim().isEmpty ? '未知' : _studentAge;
    return formatAssessmentAgeText(
      birthDate: _birthDate,
      assessmentDate: _assessmentDate,
      fallback: fallback,
    );
  }
}
