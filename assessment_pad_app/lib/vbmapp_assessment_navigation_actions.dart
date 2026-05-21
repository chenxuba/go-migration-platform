part of 'vbmapp_assessment_page.dart';

extension _VbmappAssessmentNavigationActions on _VbmappAssessmentPageState {
  void _applySelection({
    required String moduleCode,
    required int itemIndex,
  }) {
    final List<_VbmappItem> items = _itemsForModule(moduleCode);
    if (items.isEmpty) {
      return;
    }
    final int normalizedIndex = itemIndex.clamp(0, items.length - 1);
    final bool changed = _selectedModuleCode != moduleCode ||
        _selectedItemIndex != normalizedIndex;
    _selectedModuleCode = moduleCode;
    _selectedItemIndex = normalizedIndex;
    if (changed) {
      _selectionRevision.value++;
    }
  }

  void _selectModule(String code) {
    if (_selectedModuleCode == code) {
      return;
    }
    _applySelection(moduleCode: code, itemIndex: 0);
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
      _applySelection(
        moduleCode: _selectedModuleCode,
        itemIndex: _selectedItemIndex - 1,
      );
      return;
    }
    final int moduleIndex = _vbmappModules.indexWhere(
      (_VbmappModule module) => module.code == _selectedModuleCode,
    );
    if (moduleIndex <= 0) {
      return;
    }
    final String previousCode = _vbmappModules[moduleIndex - 1].code;
    _applySelection(
      moduleCode: previousCode,
      itemIndex: _itemsForModule(previousCode).length - 1,
    );
  }

  void _goNext() {
    final List<_VbmappItem> items = _selectedItems;
    if (_selectedItemIndex < items.length - 1) {
      _applySelection(
        moduleCode: _selectedModuleCode,
        itemIndex: _selectedItemIndex + 1,
      );
      return;
    }
    final int moduleIndex = _vbmappModules.indexWhere(
      (_VbmappModule module) => module.code == _selectedModuleCode,
    );
    if (moduleIndex < 0 || moduleIndex >= _vbmappModules.length - 1) {
      return;
    }
    _applySelection(
      moduleCode: _vbmappModules[moduleIndex + 1].code,
      itemIndex: 0,
    );
  }

  void _jumpFirstMissing() {
    final _VbmappItem? missing = _firstMissingItem();
    if (missing == null) {
      return;
    }
    _applySelection(
      moduleCode: missing.moduleCode,
      itemIndex: _itemsForModule(missing.moduleCode)
          .indexWhere((_VbmappItem item) => item.itemCode == missing.itemCode),
    );
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
    _applySelection(moduleCode: item.moduleCode, itemIndex: index);
  }
}
