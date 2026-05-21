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
    return _cachedAnsweredCount;
  }

  double get _progressPercent {
    return _cachedProgressPercent;
  }

  _VbmappScoreSnapshot get _scoreSnapshot {
    return _cachedScoreSnapshot;
  }

  List<_VbmappDomainScoreSummary> get _milestoneDomainSummaries {
    return _cachedScoreSnapshot.milestoneDomains;
  }

  void _rebuildScoreDerivedState() {
    final int milestoneAnswered = _milestoneScores.length;
    final int barrierAnswered = _barrierScores.length;
    final int transitionAnswered = _transitionScores.length;
    final int answeredCount =
        milestoneAnswered + barrierAnswered + transitionAnswered;
    final double milestoneTotal = _milestoneScores.values.fold<double>(
      0,
      (double total, double score) => total + score,
    );
    final int barrierTotal = _barrierScores.values.fold<int>(
      0,
      (int total, int score) => total + score,
    );
    final int transitionTotal = _transitionScores.values.fold<int>(
      0,
      (int total, int score) => total + score,
    );
    final Map<String, List<_VbmappItem>> groupedItems =
        <String, List<_VbmappItem>>{};
    for (final _VbmappItem item in _milestoneItems) {
      groupedItems
          .putIfAbsent(item.domainName, () => <_VbmappItem>[])
          .add(item);
    }
    final List<_VbmappDomainScoreSummary> milestoneDomains =
        groupedItems.entries.map((MapEntry<String, List<_VbmappItem>> entry) {
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
    _cachedAnsweredCount = answeredCount;
    _cachedAnsweredCountByModule = <String, int>{
      'milestones': milestoneAnswered,
      'barriers': barrierAnswered,
      'transition': transitionAnswered,
    };
    _cachedProgressPercent = answeredCount / _vbmappTotalItemCount;
    _cachedScoreSnapshot = _VbmappScoreSnapshot(
      milestoneTotal: milestoneTotal,
      milestoneMax: _milestoneItems.length,
      barrierTotal: barrierTotal,
      barrierMax: _barrierItems.length * 4,
      transitionTotal: transitionTotal,
      transitionMax: _transitionItems.length * 5,
      milestoneDomains: milestoneDomains,
    );
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
