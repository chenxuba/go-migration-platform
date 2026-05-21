part of 'vbmapp_assessment_page.dart';

class _VbmappMandQuickPickLabel extends StatelessWidget {
  const _VbmappMandQuickPickLabel({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Text(
      label,
      style: const TextStyle(
        color: _VbmappColors.muted,
        fontSize: 11,
        fontWeight: FontWeight.w900,
      ),
    );
  }
}

class _VbmappMandCompactChip extends StatelessWidget {
  const _VbmappMandCompactChip({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(999),
        child: Ink(
          height: 28,
          padding: const EdgeInsets.symmetric(horizontal: 9),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFE6D9) : const Color(0xFFFFFCFA),
            borderRadius: BorderRadius.circular(999),
            border: Border.all(
              color: selected ? _VbmappColors.orange : _VbmappColors.lineSoft,
            ),
          ),
          child: Center(
            child: Text(
              label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: selected ? _VbmappColors.orangeDeep : _VbmappColors.body,
                fontSize: 11.5,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMand3CoverageColumn extends StatelessWidget {
  const _VbmappMand3CoverageColumn({
    required this.label,
    required this.values,
    required this.color,
  });

  final String label;
  final List<String> values;
  final Color color;

  @override
  Widget build(BuildContext context) {
    final int filledCount = values.length > 2 ? 2 : values.length;
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Text(
            '$label $filledCount/2',
            style: TextStyle(
              color: filledCount >= 2 ? color : _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 7),
          for (int index = 0; index < 2; index++) ...<Widget>[
            _VbmappMand3CoverageSlot(
              text: index < values.length ? values[index] : '待记录',
              filled: index < values.length,
              color: color,
            ),
            if (index == 0) const SizedBox(height: 6),
          ],
        ],
      ),
    );
  }
}

class _VbmappMand3CoverageSlot extends StatelessWidget {
  const _VbmappMand3CoverageSlot({
    required this.text,
    required this.filled,
    required this.color,
  });

  final String text;
  final bool filled;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      alignment: Alignment.center,
      padding: const EdgeInsets.symmetric(horizontal: 6),
      decoration: BoxDecoration(
        color: filled ? color.withOpacity(.09) : const Color(0xFFFFF6EF),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: filled ? color.withOpacity(.24) : _VbmappColors.lineSoft,
        ),
      ),
      child: Text(
        text,
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: TextStyle(
          color: filled ? color : _VbmappColors.muted,
          fontSize: 11.5,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _VbmappMand3RecordList extends StatelessWidget {
  const _VbmappMand3RecordList({
    required this.events,
    required this.selectedIndex,
    required this.onSelectIndex,
    required this.onDeleteIndex,
  });

  final List<_VbmappMandEvent> events;
  final int? selectedIndex;
  final ValueChanged<int> onSelectIndex;
  final ValueChanged<int> onDeleteIndex;

  @override
  Widget build(BuildContext context) {
    final int itemCount = events.length < 6 ? 6 : events.length;
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final bool twoColumns = constraints.maxWidth >= 340;
        if (!twoColumns) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              for (int index = 0; index < itemCount; index++) ...<Widget>[
                _buildSlot(index),
                if (index < itemCount - 1) const SizedBox(height: 7),
              ],
            ],
          );
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            for (int index = 0; index < itemCount; index += 2) ...<Widget>[
              Row(
                children: <Widget>[
                  Expanded(child: _buildSlot(index)),
                  if (index + 1 < itemCount) ...<Widget>[
                    const SizedBox(width: 8),
                    Expanded(child: _buildSlot(index + 1)),
                  ],
                ],
              ),
              if (index + 2 < itemCount) const SizedBox(height: 7),
            ],
          ],
        );
      },
    );
  }

  Widget _buildSlot(int index) {
    return _VbmappMand3RecordSlot(
      index: index,
      event: index < events.length ? events[index] : null,
      selected: selectedIndex == index,
      onTap: () => onSelectIndex(index),
      onDelete: () => onDeleteIndex(index),
    );
  }
}

class _VbmappMand3RecordSlot extends StatelessWidget {
  const _VbmappMand3RecordSlot({
    required this.index,
    required this.event,
    required this.selected,
    required this.onTap,
    required this.onDelete,
  });

  final int index;
  final _VbmappMandEvent? event;
  final bool selected;
  final VoidCallback onTap;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    final _VbmappMandEvent? row = event;
    final bool filled = row != null;
    final bool qualified = row?.isQualified ?? false;
    final Color accent = qualified ? _VbmappColors.green : _VbmappColors.muted;
    final String dimensionText = row == null ? '' : _mand3DimensionText(row);
    return Material(
      key: ValueKey<String>('vbmapp-mand3-record-$index'),
      color: Colors.transparent,
      child: InkWell(
        onTap: filled ? onTap : null,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 42,
          padding: const EdgeInsets.symmetric(horizontal: 10),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFFBF7) : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(
              color: selected
                  ? _VbmappColors.orange
                  : filled
                      ? accent.withOpacity(.42)
                      : _VbmappColors.lineSoft,
            ),
          ),
          child: Row(
            children: <Widget>[
              SizedBox(
                width: 22,
                child: Text(
                  '${index + 1}',
                  style: TextStyle(
                    color: filled ? accent : _VbmappColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              Expanded(
                child: filled
                    ? Row(
                        children: <Widget>[
                          Flexible(
                            child: Text(
                              _mandRequestText(row),
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                color: _VbmappColors.ink,
                                fontSize: 12.5,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                          ),
                          Container(
                            width: 1,
                            height: 14,
                            margin: const EdgeInsets.symmetric(horizontal: 7),
                            color: _VbmappColors.lineSoft,
                          ),
                          Flexible(
                            child: Text(
                              dimensionText,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                color: _VbmappColors.ink,
                                fontSize: 12.5,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                          ),
                        ],
                      )
                    : const Text(
                        '等待记录',
                        style: TextStyle(
                          color: _VbmappColors.muted,
                          fontSize: 12.5,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
              ),
              const SizedBox(width: 8),
              _VbmappMandRecordActions(
                filled: filled,
                qualified: qualified,
                selected: selected,
                accent: accent,
                onDelete: onDelete,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappMandInlineChoiceGroup extends StatelessWidget {
  const _VbmappMandInlineChoiceGroup({
    required this.label,
    required this.value,
    required this.values,
    required this.onChanged,
  });

  final String label;
  final String value;
  final List<String> values;
  final ValueChanged<String> onChanged;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.max,
      children: <Widget>[
        Text(
          label,
          maxLines: 1,
          softWrap: false,
          style: const TextStyle(
            color: _VbmappColors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(width: 7),
        Expanded(
          child: Row(
            children: <Widget>[
              for (int index = 0; index < values.length; index++) ...<Widget>[
                if (index > 0) const SizedBox(width: 5),
                Expanded(
                  child: _VbmappMandChoiceButton(
                    label: values[index],
                    selected: values[index] == value,
                    onTap: () => onChanged(values[index]),
                  ),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }
}

class _VbmappMandChoiceButton extends StatelessWidget {
  const _VbmappMandChoiceButton({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(9),
        child: Ink(
          height: 30,
          padding: const EdgeInsets.symmetric(horizontal: 8),
          decoration: BoxDecoration(
            color: selected ? _VbmappColors.orange : Colors.white,
            borderRadius: BorderRadius.circular(9),
            border: Border.all(
              color: selected ? _VbmappColors.orange : _VbmappColors.line,
            ),
          ),
          child: Center(
            child: Text(
              label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              softWrap: false,
              textAlign: TextAlign.center,
              style: TextStyle(
                color: selected ? Colors.white : _VbmappColors.body,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMandInlineTextField extends StatelessWidget {
  const _VbmappMandInlineTextField({
    required this.controller,
    required this.label,
    required this.hintText,
  });

  final TextEditingController controller;
  final String label;
  final String hintText;

  @override
  Widget build(BuildContext context) {
    return TextField(
      controller: controller,
      onTapOutside: (_) => FocusManager.instance.primaryFocus?.unfocus(),
      style: const TextStyle(
        color: _VbmappColors.ink,
        fontSize: 14,
        fontWeight: FontWeight.w800,
      ),
      decoration: InputDecoration(
        labelText: label,
        hintText: hintText,
        floatingLabelBehavior: FloatingLabelBehavior.auto,
        labelStyle: const TextStyle(
          color: Color(0xFFB8A79E),
          fontSize: 13,
          fontWeight: FontWeight.w600,
        ),
        floatingLabelStyle: const TextStyle(
          color: _VbmappColors.orangeDeep,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
        hintStyle: const TextStyle(
          color: Color(0xFFC7B9B1),
          fontSize: 13,
          fontWeight: FontWeight.w500,
        ),
        isDense: true,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 14, vertical: 13),
        filled: true,
        fillColor: Colors.white,
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: _VbmappColors.lineSoft),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: const BorderSide(color: _VbmappColors.orange),
        ),
      ),
    );
  }
}

class _VbmappMandMaterialChip extends StatelessWidget {
  const _VbmappMandMaterialChip({
    required this.label,
    required this.selected,
    required this.onTap,
  });

  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(999),
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFE6D9) : Colors.white,
            borderRadius: BorderRadius.circular(999),
            border: Border.all(
              color: selected ? _VbmappColors.orange : _VbmappColors.lineSoft,
            ),
          ),
          child: Text(
            label,
            style: TextStyle(
              color: selected ? _VbmappColors.orangeDeep : _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMand1RecordGrid extends StatelessWidget {
  const _VbmappMand1RecordGrid({
    required this.events,
    required this.minSlots,
    required this.selectedIndex,
    required this.onSelectIndex,
    required this.onDeleteIndex,
    this.isQualified,
    this.metaTextBuilder,
  });

  final List<_VbmappMandEvent> events;
  final int minSlots;
  final int? selectedIndex;
  final ValueChanged<int> onSelectIndex;
  final ValueChanged<int> onDeleteIndex;
  final bool Function(_VbmappMandEvent event)? isQualified;
  final String Function(_VbmappMandEvent event)? metaTextBuilder;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        const double spacing = 8;
        final int itemCount =
            events.length < minSlots ? minSlots : events.length;
        final bool twoColumns = constraints.maxWidth >= 360;
        if (!twoColumns) {
          return Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: <Widget>[
              for (int index = 0; index < itemCount; index++) ...<Widget>[
                _buildSlot(index),
                if (index < itemCount - 1) const SizedBox(height: spacing),
              ],
            ],
          );
        }
        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            for (int index = 0; index < itemCount; index += 2) ...<Widget>[
              if (index + 1 < itemCount)
                IntrinsicHeight(
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: <Widget>[
                      Expanded(child: _buildSlot(index)),
                      const SizedBox(width: spacing),
                      Expanded(child: _buildSlot(index + 1)),
                    ],
                  ),
                )
              else
                _buildSlot(index),
              if (index + 2 < itemCount) const SizedBox(height: spacing),
            ],
          ],
        );
      },
    );
  }

  Widget _buildSlot(int index) {
    return _VbmappMand1RecordSlot(
      index: index,
      event: index < events.length ? events[index] : null,
      selected: selectedIndex == index,
      onTap: () => onSelectIndex(index),
      onDelete: () => onDeleteIndex(index),
      qualificationResolver: isQualified,
      metaTextBuilder: metaTextBuilder,
    );
  }
}

class _VbmappMand1RecordSlot extends StatelessWidget {
  const _VbmappMand1RecordSlot({
    required this.index,
    required this.event,
    required this.selected,
    required this.onTap,
    required this.onDelete,
    this.qualificationResolver,
    this.metaTextBuilder,
  });

  final int index;
  final _VbmappMandEvent? event;
  final bool selected;
  final VoidCallback onTap;
  final VoidCallback onDelete;
  final bool Function(_VbmappMandEvent event)? qualificationResolver;
  final String Function(_VbmappMandEvent event)? metaTextBuilder;

  @override
  Widget build(BuildContext context) {
    final _VbmappMandEvent? row = event;
    final bool filled = row != null;
    final bool qualified = row == null
        ? false
        : (qualificationResolver?.call(row) ?? row.isQualified);
    final Color accent = qualified ? _VbmappColors.green : _VbmappColors.muted;
    final String requestText = row == null ? '' : _mandRequestText(row);
    final BorderRadius radius = BorderRadius.circular(10);
    return Material(
      key: ValueKey<String>('vbmapp-mand-record-$index'),
      color: Colors.transparent,
      child: InkWell(
        onTap: filled ? onTap : null,
        borderRadius: radius,
        child: Ink(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 9),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFFBF7) : Colors.white,
            borderRadius: radius,
            border: Border.all(
              color: selected
                  ? _VbmappColors.orange
                  : filled
                      ? accent.withOpacity(.42)
                      : _VbmappColors.lineSoft,
            ),
          ),
          child: ConstrainedBox(
            constraints: const BoxConstraints(minHeight: 34),
            child: Row(
              children: <Widget>[
                Container(
                  width: 28,
                  height: 28,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: filled
                        ? accent.withOpacity(.12)
                        : const Color(0xFFFFF6EF),
                    shape: BoxShape.circle,
                  ),
                  child: Text(
                    '${index + 1}',
                    style: TextStyle(
                      color: filled ? accent : _VbmappColors.muted,
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                const SizedBox(width: 9),
                Expanded(
                  child: filled
                      ? Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: <Widget>[
                            Text(
                              requestText,
                              maxLines: 2,
                              style: const TextStyle(
                                color: _VbmappColors.ink,
                                fontSize: 13,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                            const SizedBox(height: 3),
                            Text(
                              metaTextBuilder?.call(row) ??
                                  _mandRecordMetaText(row),
                              maxLines: 2,
                              style: const TextStyle(
                                color: _VbmappColors.body,
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                          ],
                        )
                      : const Text(
                          '等待记录一条有效要求',
                          style: TextStyle(
                            color: _VbmappColors.muted,
                            fontSize: 13,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                ),
                const SizedBox(width: 8),
                _VbmappMandRecordActions(
                  filled: filled,
                  qualified: qualified,
                  selected: selected,
                  accent: accent,
                  onDelete: onDelete,
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMandRecordActions extends StatelessWidget {
  const _VbmappMandRecordActions({
    required this.filled,
    required this.qualified,
    required this.selected,
    required this.accent,
    required this.onDelete,
  });

  final bool filled;
  final bool qualified;
  final bool selected;
  final Color accent;
  final VoidCallback onDelete;

  @override
  Widget build(BuildContext context) {
    if (!selected || !filled) {
      return Text(
        filled ? (qualified ? '计入' : '不计') : '-',
        style: TextStyle(
          color: accent,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      );
    }
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Text(
          qualified ? '计入' : '不计',
          style: TextStyle(
            color: accent,
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
        const SizedBox(width: 6),
        Material(
          color: Colors.transparent,
          child: InkWell(
            key: const ValueKey<String>('vbmapp-mand-delete-record'),
            onTap: onDelete,
            borderRadius: BorderRadius.circular(999),
            child: Ink(
              height: 22,
              padding: const EdgeInsets.symmetric(horizontal: 7),
              decoration: BoxDecoration(
                color: const Color(0xFFFFE6D9),
                borderRadius: BorderRadius.circular(999),
                border: Border.all(color: _VbmappColors.orange),
              ),
              child: const Center(
                child: Text(
                  '删除',
                  style: TextStyle(
                    color: _VbmappColors.orangeDeep,
                    fontSize: 11,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _VbmappMandRecorderPanel extends StatelessWidget {
  const _VbmappMandRecorderPanel({
    required this.item,
    required this.materialProfile,
    required this.events,
    required this.onAddEvent,
  });

  final _VbmappItem item;
  final VbmappMaterialProfile? materialProfile;
  final List<_VbmappMandEvent> events;
  final VoidCallback onAddEvent;

  @override
  Widget build(BuildContext context) {
    final int qualifiedCount = _qualifiedMandCount(events);
    final double suggestedScore = _suggestMandScore(item, events);
    final bool generalizationMode = item.itemCode == 'MAND_03M';
    final Map<String, int> generalizationCounts =
        _mandGeneralizationCounts(events);
    final List<_VbmappMandEvent> visibleEvents =
        events.length <= 4 ? events : events.sublist(events.length - 4);
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Row(
            children: <Widget>[
              Icon(Icons.record_voice_over_outlined,
                  color: item.color, size: 21),
              const SizedBox(width: 8),
              const Expanded(
                child: Text(
                  '提要求事件记录',
                  style: TextStyle(
                    color: _VbmappColors.ink,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              _VbmappEvidenceMetric(
                label: '有效',
                value: '$qualifiedCount',
                color: item.color,
              ),
              const SizedBox(width: 8),
              if (generalizationMode) ...<Widget>[
                _VbmappEvidenceMetric(
                  label: '人',
                  value: '${generalizationCounts['people'] ?? 0}/2',
                  color: item.color,
                ),
                const SizedBox(width: 8),
                _VbmappEvidenceMetric(
                  label: '环境',
                  value: '${generalizationCounts['settings'] ?? 0}/2',
                  color: item.color,
                ),
                const SizedBox(width: 8),
                _VbmappEvidenceMetric(
                  label: '例子',
                  value: '${generalizationCounts['examples'] ?? 0}/2',
                  color: item.color,
                ),
                const SizedBox(width: 8),
              ],
              _VbmappEvidenceMetric(
                label: '建议',
                value: _formatScore(suggestedScore),
                color: item.color,
              ),
            ],
          ),
          if (materialProfile != null) ...<Widget>[
            const SizedBox(height: 10),
            _VbmappInlineInfo(
              icon: Icons.inventory_2_outlined,
              text: materialProfile!.label,
            ),
          ],
          const SizedBox(height: 12),
          if (visibleEvents.isEmpty)
            const _VbmappEmptyEvidence(text: '还没有记录孩子实际发出的要求')
          else
            Column(
              children: <Widget>[
                for (final _VbmappMandEvent event in visibleEvents)
                  _VbmappMandEventRow(event: event),
              ],
            ),
          const SizedBox(height: 12),
          Align(
            alignment: Alignment.centerLeft,
            child: _VbmappSmallActionButton(
              icon: Icons.add_rounded,
              label: generalizationMode ? '记录一次泛化要求' : '记录一次要求',
              onTap: onAddEvent,
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappMandEventRow extends StatelessWidget {
  const _VbmappMandEventRow({required this.event});

  final _VbmappMandEvent event;

  @override
  Widget build(BuildContext context) {
    final Color color =
        event.isQualified ? _VbmappColors.green : _VbmappColors.muted;
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 9),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          Icon(
            event.isQualified
                ? Icons.check_circle_outline_rounded
                : Icons.radio_button_unchecked_rounded,
            color: color,
            size: 18,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              event.summary,
              style: const TextStyle(
                color: _VbmappColors.ink,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          const SizedBox(width: 8),
          Text(
            event.promptLevel,
            style: TextStyle(
              color: color,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _VbmappEvidenceMetric extends StatelessWidget {
  const _VbmappEvidenceMetric({
    required this.label,
    required this.value,
    required this.color,
  });

  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: color.withOpacity(.1),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        '$label $value',
        style: TextStyle(
          color: color,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _VbmappInlineInfo extends StatelessWidget {
  const _VbmappInlineInfo({
    required this.icon,
    required this.text,
  });

  final IconData icon;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Icon(icon, color: _VbmappColors.orange, size: 17),
        const SizedBox(width: 6),
        Expanded(
          child: Text(
            text,
            style: const TextStyle(
              color: _VbmappColors.body,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
      ],
    );
  }
}

class _VbmappEmptyEvidence extends StatelessWidget {
  const _VbmappEmptyEvidence({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 13),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _VbmappColors.lineSoft),
      ),
      child: Text(
        text,
        style: const TextStyle(
          color: _VbmappColors.muted,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _VbmappSmallActionButton extends StatelessWidget {
  const _VbmappSmallActionButton({
    required this.icon,
    required this.label,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          height: 42,
          padding: const EdgeInsets.symmetric(horizontal: 13),
          decoration: BoxDecoration(
            color: _VbmappColors.orange,
            borderRadius: BorderRadius.circular(10),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(icon, size: 18, color: Colors.white),
              const SizedBox(width: 6),
              Text(
                label,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 13,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _VbmappMand4TimerButton extends StatelessWidget {
  const _VbmappMand4TimerButton({
    required this.icon,
    required this.label,
    required this.onTap,
    this.filled = false,
    this.compact = false,
    super.key,
  });

  final IconData icon;
  final String label;
  final VoidCallback? onTap;
  final bool filled;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final bool enabled = onTap != null;
    final Color textColor = enabled
        ? filled
            ? Colors.white
            : _VbmappColors.orangeDeep
        : _VbmappColors.muted;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(9),
        child: Ink(
          height: compact ? 28 : 32,
          padding: EdgeInsets.symmetric(horizontal: compact ? 8 : 10),
          decoration: BoxDecoration(
            color: filled && enabled
                ? _VbmappColors.orange
                : enabled
                    ? const Color(0xFFFFFCFA)
                    : const Color(0xFFF7F1ED),
            borderRadius: BorderRadius.circular(9),
            border: Border.all(
              color: enabled ? _VbmappColors.orange : const Color(0xFFE2D6CE),
            ),
          ),
          child: Center(
            child: Row(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.center,
              children: <Widget>[
                Icon(icon, size: compact ? 14 : 16, color: textColor),
                SizedBox(width: compact ? 4 : 5),
                Text(
                  label,
                  style: TextStyle(
                    color: textColor,
                    fontSize: compact ? 11.5 : 12.5,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _VbmappMand4TimerMetricChip extends StatelessWidget {
  const _VbmappMand4TimerMetricChip({
    required this.label,
    required this.value,
    required this.tone,
  });

  final String label;
  final String value;
  final Color tone;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 7),
      decoration: BoxDecoration(
        color: tone.withOpacity(.1),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: tone.withOpacity(.18)),
      ),
      child: Text(
        '$label $value',
        style: TextStyle(
          color: tone,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}
