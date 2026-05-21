part of 'vbmapp_assessment_page.dart';

extension _VbmappAssessmentNavigationActions on _VbmappAssessmentPageState {
  void _selectModule(String code) {
    if (_selectedModuleCode == code) {
      return;
    }
    setState(() {
      _selectedModuleCode = code;
      _selectedItemIndex = 0;
    });
  }

  void _selectScore(num score) {
    final _VbmappItem item = _selectedItem;
    setState(() {
      switch (item.moduleCode) {
        case 'milestones':
          _milestoneScores[item.itemCode] = score.toDouble();
          break;
        case 'barriers':
          _barrierScores[item.itemCode] = score.toInt();
          break;
        case 'transition':
          _transitionScores[item.itemCode] = score.toInt();
          break;
      }
      _rebuildScoreDerivedState();
      _autoSaveText = '已记录';
    });
    if (_autoNext) {
      Future<void>.delayed(const Duration(milliseconds: 220), () {
        if (mounted) {
          _goNext();
        }
      });
    }
  }

  void _goPrevious() {
    if (_selectedItemIndex > 0) {
      setState(() => _selectedItemIndex -= 1);
      return;
    }
    final int moduleIndex = _vbmappModules.indexWhere(
      (_VbmappModule module) => module.code == _selectedModuleCode,
    );
    if (moduleIndex <= 0) {
      return;
    }
    final String previousCode = _vbmappModules[moduleIndex - 1].code;
    setState(() {
      _selectedModuleCode = previousCode;
      _selectedItemIndex = _itemsForModule(previousCode).length - 1;
    });
  }

  void _goNext() {
    final List<_VbmappItem> items = _selectedItems;
    if (_selectedItemIndex < items.length - 1) {
      setState(() => _selectedItemIndex += 1);
      return;
    }
    final int moduleIndex = _vbmappModules.indexWhere(
      (_VbmappModule module) => module.code == _selectedModuleCode,
    );
    if (moduleIndex < 0 || moduleIndex >= _vbmappModules.length - 1) {
      return;
    }
    setState(() {
      _selectedModuleCode = _vbmappModules[moduleIndex + 1].code;
      _selectedItemIndex = 0;
    });
  }

  void _jumpFirstMissing() {
    final _VbmappItem? missing = _firstMissingItem();
    if (missing == null) {
      return;
    }
    setState(() {
      _selectedModuleCode = missing.moduleCode;
      _selectedItemIndex = _itemsForModule(missing.moduleCode)
          .indexWhere((_VbmappItem item) => item.itemCode == missing.itemCode);
    });
  }

  _VbmappItem? _firstMissingItem() {
    for (final _VbmappModule module in _vbmappModules) {
      final List<_VbmappItem> items = _itemsForModule(module.code);
      final int index =
          items.indexWhere((_VbmappItem item) => _scoreFor(item) == null);
      if (index >= 0) {
        return items[index];
      }
    }
    return null;
  }

  void _selectItem(_VbmappItem item) {
    final List<_VbmappItem> items = _itemsForModule(item.moduleCode);
    final int index = items.indexWhere(
      (_VbmappItem candidate) => candidate.itemCode == item.itemCode,
    );
    if (index < 0) {
      return;
    }
    setState(() {
      _selectedModuleCode = item.moduleCode;
      _selectedItemIndex = index;
    });
  }
}
