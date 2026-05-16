import 'package:flutter/material.dart';

class SupervisionWorkbenchPage extends StatefulWidget {
  const SupervisionWorkbenchPage({required this.onBack, super.key});

  final VoidCallback onBack;

  @override
  State<SupervisionWorkbenchPage> createState() =>
      _SupervisionWorkbenchPageState();
}

class _SupervisionWorkbenchPageState extends State<SupervisionWorkbenchPage> {
  static const List<_SupervisionItem> _items = <_SupervisionItem>[
    _SupervisionItem(
      id: 'item-1',
      kind: _SupervisionKind.iep,
      childName: '陈小宇',
      age: 6,
      className: '语言表达班 A101',
      teacherName: '张老师',
      dueText: '今日 16:30',
      status: '待审核',
      target: '能在提示下主动表达需求',
      baseline: '目前需完整口头提示',
      criteria: '连续 3 次课，80% 机会轻提示完成',
      recentRecord: '最近 2 次 DTT 记录完整，独立完成率 58%',
      note: '先看目标能不能直接落到课堂记录，再决定是否拆分。',
      accent: Color(0xFFE96F43),
    ),
    _SupervisionItem(
      id: 'item-2',
      kind: _SupervisionKind.record,
      childName: '刘一诺',
      age: 7,
      className: '社交沟通班 B205',
      teacherName: '李老师',
      dueText: '明日 10:00',
      status: '待复核',
      target: '能在同伴互动中完成轮流回应',
      baseline: '需要成人完整提示',
      criteria: '5 次机会中，至少 4 次轻提示完成',
      recentRecord: '课堂记录显示正确率 3/5，提示等待不足',
      note: '优先核对记录和目标是否一致。',
      accent: Color(0xFF3F82D2),
    ),
    _SupervisionItem(
      id: 'item-3',
      kind: _SupervisionKind.adjust,
      childName: '周子航',
      age: 5,
      className: '精细动作班 C301',
      teacherName: '王老师',
      dueText: '今天 18:00',
      status: '待调整',
      target: '能完成双手协作的穿珠任务',
      baseline: '完成需要分步身体辅助',
      criteria: '连续 2 周，独立完成率达到 70%',
      recentRecord: '记录显示材料难度偏高，完成率波动大',
      note: '记录结果要回写到下周期目标和材料难度。',
      accent: Color(0xFF7F77C8),
    ),
    _SupervisionItem(
      id: 'item-4',
      kind: _SupervisionKind.iep,
      childName: '孙乐乐',
      age: 6,
      className: '情绪管理班 A107',
      teacherName: '陈老师',
      dueText: '明日 14:00',
      status: '待审核',
      target: '在情绪升高时能接受成人提示',
      baseline: '情绪失控时需要成人介入',
      criteria: '3 次情境中，2 次能在轻提示下恢复',
      recentRecord: '上周期记录完整，但目标描述略泛',
      note: '目标要更可记录，避免只写意图不写标准。',
      accent: Color(0xFFD96A7F),
    ),
  ];

  static const List<_StageTab> _tabs = <_StageTab>[
    _StageTab(label: '全部', kind: null),
    _StageTab(label: '待审核 IEP', kind: _SupervisionKind.iep),
    _StageTab(label: '待复核记录', kind: _SupervisionKind.record),
    _StageTab(label: '待调整计划', kind: _SupervisionKind.adjust),
  ];

  final TextEditingController _searchController = TextEditingController();
  int _selectedTabIndex = 0;
  String _query = '';
  String _selectedItemId = 'item-1';
  String _lastAction = '待处理';

  @override
  void initState() {
    super.initState();
    _searchController.addListener(() {
      final String next = _searchController.text.trim().toLowerCase();
      if (next == _query) {
        return;
      }
      setState(() {
        _query = next;
      });
      _normalizeSelection();
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  List<_SupervisionItem> get _visibleItems {
    final _StageTab tab = _tabs[_selectedTabIndex];
    return _items.where((_SupervisionItem item) {
      final bool tabMatch = tab.kind == null || item.kind == tab.kind;
      if (!tabMatch) {
        return false;
      }
      if (_query.isEmpty) {
        return true;
      }
      final String haystack =
          '${item.childName} ${item.className} ${item.teacherName} ${item.target} ${item.recentRecord} ${item.note}'
              .toLowerCase();
      return haystack.contains(_query);
    }).toList(growable: false);
  }

  _SupervisionItem get _selectedItem {
    final List<_SupervisionItem> visible = _visibleItems;
    if (visible.isEmpty) {
      return _items.first;
    }
    return visible.firstWhere(
      (_SupervisionItem item) => item.id == _selectedItemId,
      orElse: () => visible.first,
    );
  }

  int get _visibleIepCount => _items
      .where((_SupervisionItem item) => item.kind == _SupervisionKind.iep)
      .length;

  int get _visibleRecordCount => _items
      .where((_SupervisionItem item) => item.kind == _SupervisionKind.record)
      .length;

  int get _visibleAdjustCount => _items
      .where((_SupervisionItem item) => item.kind == _SupervisionKind.adjust)
      .length;

  void _selectTab(int index) {
    setState(() {
      _selectedTabIndex = index;
    });
    _normalizeSelection();
  }

  void _selectItem(String itemId) {
    setState(() {
      _selectedItemId = itemId;
    });
  }

  void _setAction(String action) {
    setState(() {
      _lastAction = action;
    });
  }

  void _normalizeSelection() {
    final List<_SupervisionItem> visible = _visibleItems;
    if (visible.isEmpty) {
      return;
    }
    final bool stillVisible =
        visible.any((_SupervisionItem item) => item.id == _selectedItemId);
    if (stillVisible) {
      return;
    }
    setState(() {
      _selectedItemId = visible.first.id;
    });
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width =
            constraints.maxWidth.isFinite ? constraints.maxWidth : 1366;
        final bool compact = width < 1180;
        final double outer = compact ? 16 : 24;
        final double gap = compact ? 12 : 16;
        final double leftWidth = compact ? 280 : 300;
        final double rightWidth = compact ? 292 : 316;
        final double centerWidth =
            width - outer * 2 - leftWidth - rightWidth - gap * 2;
        final List<_SupervisionItem> visibleItems = _visibleItems;
        final _SupervisionItem selectedItem = _selectedItem;

        return ColoredBox(
          color: const Color(0xFFFFF7EE),
          child: Padding(
            padding: EdgeInsets.fromLTRB(outer, 20, outer, 18),
            child: Column(
              children: <Widget>[
                _TopBar(
                  onBack: widget.onBack,
                  searchController: _searchController,
                  totalCount: _items.length,
                ),
                const SizedBox(height: 14),
                _TabStrip(
                  selectedIndex: _selectedTabIndex,
                  tabs: _tabs,
                  visibleIepCount: _visibleIepCount,
                  visibleRecordCount: _visibleRecordCount,
                  visibleAdjustCount: _visibleAdjustCount,
                  onChanged: _selectTab,
                ),
                const SizedBox(height: 14),
                Expanded(
                  child: Row(
                    children: <Widget>[
                      SizedBox(
                        width: leftWidth,
                        child: _ItemListPanel(
                          items: visibleItems,
                          selectedItemId: selectedItem.id,
                          onItemSelected: _selectItem,
                        ),
                      ),
                      SizedBox(width: gap),
                      SizedBox(
                        width: centerWidth > 0 ? centerWidth : 0,
                        child: _DetailWorkspace(item: selectedItem),
                      ),
                      SizedBox(width: gap),
                      SizedBox(
                        width: rightWidth,
                        child: _ReviewActions(
                          item: selectedItem,
                          lastAction: _lastAction,
                          onAction: _setAction,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

class _TopBar extends StatelessWidget {
  const _TopBar({
    required this.onBack,
    required this.searchController,
    required this.totalCount,
  });

  final VoidCallback onBack;
  final TextEditingController searchController;
  final int totalCount;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        _BackButton(onTap: onBack),
        const SizedBox(width: 14),
        const Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(
              '督导工作台',
              style: TextStyle(
                color: Color(0xFF3F2B22),
                fontSize: 25,
                fontWeight: FontWeight.w900,
                height: 1,
              ),
            ),
            SizedBox(height: 6),
            Text(
              'IEP 制定、审核、记录复核、计划调整',
              style: TextStyle(
                color: Color(0xFF6F5B50),
                fontSize: 12,
                fontWeight: FontWeight.w700,
                height: 1,
              ),
            ),
          ],
        ),
        const Spacer(),
        SizedBox(
          width: 320,
          height: 42,
          child: TextField(
            controller: searchController,
            decoration: InputDecoration(
              hintText: '搜索学员 / 老师 / 目标',
              hintStyle: const TextStyle(
                color: Color(0xFFA7958B),
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
              prefixIcon: const Icon(
                Icons.search_rounded,
                color: Color(0xFFA7958B),
                size: 20,
              ),
              filled: true,
              fillColor: Colors.white.withOpacity(.94),
              contentPadding: const EdgeInsets.symmetric(horizontal: 14),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(13),
                borderSide: const BorderSide(color: Color(0xFFEAD7C9)),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(13),
                borderSide: const BorderSide(color: Color(0xFFEAD7C9)),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(13),
                borderSide: const BorderSide(color: Color(0xFFE96F43)),
              ),
            ),
            style: const TextStyle(
              color: Color(0xFF3F2B22),
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
        ),
        const SizedBox(width: 10),
        _TopAction(
            label: '共 $totalCount 项', icon: Icons.assignment_turned_in_rounded),
        const SizedBox(width: 10),
        _TopAction(
          label: '新建 IEP',
          icon: Icons.add_rounded,
          accent: true,
          onTap: () {},
        ),
      ],
    );
  }
}

class _BackButton extends StatelessWidget {
  const _BackButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white.withOpacity(.92),
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: onTap,
        child: Container(
          width: 42,
          height: 42,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: const Color(0xFFEAD7C9)),
          ),
          child: const Icon(
            Icons.chevron_left_rounded,
            color: Color(0xFF6F5B50),
            size: 30,
          ),
        ),
      ),
    );
  }
}

class _TopAction extends StatelessWidget {
  const _TopAction({
    required this.label,
    required this.icon,
    this.accent = false,
    this.onTap,
  });

  final String label;
  final IconData icon;
  final bool accent;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Container(
          height: 42,
          padding: const EdgeInsets.symmetric(horizontal: 14),
          decoration: BoxDecoration(
            color: accent
                ? const Color(0xFFE96F43)
                : Colors.white.withOpacity(.94),
            borderRadius: BorderRadius.circular(13),
            border: accent ? null : Border.all(color: const Color(0xFFEAD7C9)),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(
                icon,
                color: accent ? Colors.white : const Color(0xFF6F5B50),
                size: 18,
              ),
              const SizedBox(width: 6),
              Text(
                label,
                style: TextStyle(
                  color: accent ? Colors.white : const Color(0xFF6F5B50),
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

class _TabStrip extends StatelessWidget {
  const _TabStrip({
    required this.selectedIndex,
    required this.tabs,
    required this.visibleIepCount,
    required this.visibleRecordCount,
    required this.visibleAdjustCount,
    required this.onChanged,
  });

  final int selectedIndex;
  final List<_StageTab> tabs;
  final int visibleIepCount;
  final int visibleRecordCount;
  final int visibleAdjustCount;
  final ValueChanged<int> onChanged;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        Container(
          height: 42,
          padding: const EdgeInsets.all(4),
          decoration: BoxDecoration(
            color: const Color(0xFFFFF1E8).withOpacity(.82),
            borderRadius: BorderRadius.circular(13),
            border: Border.all(color: const Color(0xFFEAD7C9)),
          ),
          child: Row(
            children: <Widget>[
              for (int index = 0; index < tabs.length; index++)
                _TabChip(
                  label: tabs[index].label,
                  active: index == selectedIndex,
                  onTap: () => onChanged(index),
                ),
            ],
          ),
        ),
        const Spacer(),
        _MiniStat(label: 'IEP $visibleIepCount'),
        const SizedBox(width: 12),
        _MiniStat(label: '记录 $visibleRecordCount'),
        const SizedBox(width: 12),
        _MiniStat(label: '调整 $visibleAdjustCount'),
      ],
    );
  }
}

class _TabChip extends StatelessWidget {
  const _TabChip({
    required this.label,
    required this.active,
    required this.onTap,
  });

  final String label;
  final bool active;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(10),
      child: Container(
        height: 34,
        padding: const EdgeInsets.symmetric(horizontal: 14),
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: active ? Colors.white : Colors.transparent,
          borderRadius: BorderRadius.circular(10),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: active ? const Color(0xFFC95D37) : const Color(0xFF6F5B50),
            fontSize: 12,
            fontWeight: FontWeight.w900,
          ),
        ),
      ),
    );
  }
}

class _MiniStat extends StatelessWidget {
  const _MiniStat({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Text(
      label,
      style: const TextStyle(
        color: Color(0xFF6F5B50),
        fontSize: 12,
        fontWeight: FontWeight.w800,
      ),
    );
  }
}

class _ItemListPanel extends StatelessWidget {
  const _ItemListPanel({
    required this.items,
    required this.selectedItemId,
    required this.onItemSelected,
  });

  final List<_SupervisionItem> items;
  final String selectedItemId;
  final ValueChanged<String> onItemSelected;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFEAD7C9)),
      ),
      child: Column(
        children: <Widget>[
          const _PanelHeader(),
          const Divider(height: 1, color: Color(0xFFF4E8DF)),
          Expanded(
            child: items.isEmpty
                ? const _EmptyState()
                : ListView.separated(
                    padding: EdgeInsets.zero,
                    itemCount: items.length,
                    separatorBuilder: (_, __) =>
                        const Divider(height: 1, color: Color(0xFFF4E8DF)),
                    itemBuilder: (BuildContext context, int index) {
                      final _SupervisionItem item = items[index];
                      return _ItemRow(
                        item: item,
                        selected: item.id == selectedItemId,
                        onTap: () => onItemSelected(item.id),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}

class _PanelHeader extends StatelessWidget {
  const _PanelHeader();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 60,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Row(
        children: const <Widget>[
          Expanded(
            child: Text(
              '闭环队列',
              style: TextStyle(
                color: Color(0xFF3F2B22),
                fontSize: 17,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          _MiniBadge(label: '今日处理'),
        ],
      ),
    );
  }
}

class _MiniBadge extends StatelessWidget {
  const _MiniBadge({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF8F2),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: const Color(0xFFF4E8DF)),
      ),
      child: Text(
        label,
        style: const TextStyle(
          color: Color(0xFF6F5B50),
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  const _SectionHeader({
    required this.title,
    required this.subtitle,
  });

  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 60,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Row(
        children: <Widget>[
          Expanded(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  title,
                  style: const TextStyle(
                    color: Color(0xFF3F2B22),
                    fontSize: 17,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  subtitle,
                  style: const TextStyle(
                    color: Color(0xFFA7958B),
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ItemRow extends StatelessWidget {
  const _ItemRow({
    required this.item,
    required this.selected,
    required this.onTap,
  });

  final _SupervisionItem item;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: selected ? const Color(0xFFFFF5EC) : Colors.transparent,
      child: InkWell(
        onTap: onTap,
        child: SizedBox(
          height: 76,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 14),
            child: Row(
              children: <Widget>[
                Container(
                  width: 36,
                  height: 36,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: item.accent,
                    shape: BoxShape.circle,
                  ),
                  child: Text(
                    item.childName.substring(0, 1),
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 15,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Row(
                        children: <Widget>[
                          _KindBadge(
                              label: item.kind.label, color: item.accent),
                          const SizedBox(width: 6),
                          Expanded(
                            child: Text(
                              '${item.childName} · ${item.age}岁',
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                color: Color(0xFF3F2B22),
                                fontSize: 13,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 3),
                      Text(
                        item.className,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          color: Color(0xFF9B897D),
                          fontSize: 11,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 8),
                Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: <Widget>[
                    Text(
                      item.dueText,
                      style: const TextStyle(
                        color: Color(0xFF6F5B50),
                        fontSize: 11,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                    const SizedBox(height: 6),
                    _StatusBadge(label: item.status, color: item.accent),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _KindBadge extends StatelessWidget {
  const _KindBadge({required this.label, required this.color});

  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 18,
      padding: const EdgeInsets.symmetric(horizontal: 6),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: color.withOpacity(.13),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: color,
          fontSize: 10,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _StatusBadge extends StatelessWidget {
  const _StatusBadge({required this.label, required this.color});

  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 24,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: color.withOpacity(.13),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: color,
          fontSize: 11,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _EmptyState extends StatelessWidget {
  const _EmptyState();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Text(
        '没有符合条件的待办',
        style: TextStyle(
          color: Color(0xFFA7958B),
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _DetailWorkspace extends StatelessWidget {
  const _DetailWorkspace({required this.item});

  final _SupervisionItem item;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFEAD7C9)),
      ),
      child: Column(
        children: <Widget>[
          _SectionHeader(
            title: 'IEP 复核区',
            subtitle: '${item.childName} · ${item.className}',
          ),
          const Divider(height: 1, color: Color(0xFFF4E8DF)),
          Expanded(
            child: SingleChildScrollView(
              physics: const ClampingScrollPhysics(),
              padding: const EdgeInsets.all(14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  _WorkspaceCard(
                    title: 'IEP 核心',
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        _KeyLine(label: '目标', value: item.target),
                        _KeyLine(label: '基线', value: item.baseline),
                        _KeyLine(label: '达成标准', value: item.criteria),
                        _KeyLine(label: '责任教师', value: item.teacherName),
                      ],
                    ),
                  ),
                  const SizedBox(height: 10),
                  _WorkspaceCard(
                    title: 'DTT / 课堂记录',
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        _KeyLine(label: '最新记录', value: item.recentRecord),
                        const SizedBox(height: 8),
                        const _InfoPillRow(),
                      ],
                    ),
                  ),
                  const SizedBox(height: 10),
                  _WorkspaceCard(
                    title: '督导判断',
                    child: Text(
                      item.note,
                      style: const TextStyle(
                        color: Color(0xFF6F5B50),
                        fontSize: 12,
                        height: 1.45,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _KeyLine extends StatelessWidget {
  const _KeyLine({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            label,
            style: const TextStyle(
              color: Color(0xFFA7958B),
              fontSize: 10,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 3),
          Text(
            value,
            style: const TextStyle(
              color: Color(0xFF3F2B22),
              fontSize: 13,
              height: 1.35,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _InfoPillRow extends StatelessWidget {
  const _InfoPillRow();

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 6,
      runSpacing: 6,
      children: const <Widget>[
        _InfoPill(label: '目标可记录'),
        _InfoPill(label: '提示层级'),
        _InfoPill(label: '闭环复核'),
      ],
    );
  }
}

class _InfoPill extends StatelessWidget {
  const _InfoPill({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 26,
      padding: const EdgeInsets.symmetric(horizontal: 9),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF6EA),
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: const Color(0xFFF1DFCD)),
      ),
      child: Text(
        label,
        style: const TextStyle(
          color: Color(0xFF6F5B50),
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _WorkspaceCard extends StatelessWidget {
  const _WorkspaceCard({required this.title, required this.child});

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFF4E8DF)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            title,
            style: const TextStyle(
              color: Color(0xFF3F2B22),
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 10),
          child,
        ],
      ),
    );
  }
}

class _ReviewActions extends StatelessWidget {
  const _ReviewActions({
    required this.item,
    required this.lastAction,
    required this.onAction,
  });

  final _SupervisionItem item;
  final String lastAction;
  final ValueChanged<String> onAction;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFEAD7C9)),
      ),
      child: Column(
        children: <Widget>[
          _SectionHeader(
            title: '处理面板',
            subtitle: '当前状态：${item.status}',
          ),
          const Divider(height: 1, color: Color(0xFFF4E8DF)),
          Expanded(
            child: SingleChildScrollView(
              physics: const ClampingScrollPhysics(),
              padding: const EdgeInsets.all(14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  _WorkspaceCard(
                    title: '流程步骤',
                    child: Column(
                      children: <Widget>[
                        _StepLine(
                          index: 1,
                          title: '审核 IEP',
                          desc: '确认目标可执行、可记录',
                          active: item.kind == _SupervisionKind.iep,
                        ),
                        _StepLine(
                          index: 2,
                          title: '复核记录',
                          desc: '看课堂数据是否和目标一致',
                          active: item.kind == _SupervisionKind.record,
                        ),
                        _StepLine(
                          index: 3,
                          title: '回写调整',
                          desc: '把记录结果写回下一周期计划',
                          active: item.kind == _SupervisionKind.adjust,
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 10),
                  _WorkspaceCard(
                    title: '最近动作',
                    child: Text(
                      lastAction,
                      style: const TextStyle(
                        color: Color(0xFF6F5B50),
                        fontSize: 12,
                        height: 1.4,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(14),
            child: Column(
              children: <Widget>[
                Row(
                  children: <Widget>[
                    Expanded(
                      child: _ActionButton(
                        label: '通过审核',
                        filled: true,
                        onTap: () => onAction('已通过'),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: _ActionButton(
                        label: '退回修改',
                        onTap: () => onAction('需修改'),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Row(
                  children: <Widget>[
                    Expanded(
                      child: _ActionButton(
                        label: '标记复核',
                        onTap: () => onAction('已复核'),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: _ActionButton(
                        label: '发起调整',
                        filled: true,
                        onTap: () => onAction('待调整'),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _StepLine extends StatelessWidget {
  const _StepLine({
    required this.index,
    required this.title,
    required this.desc,
    required this.active,
  });

  final int index;
  final String title;
  final String desc;
  final bool active;

  @override
  Widget build(BuildContext context) {
    final Color color =
        active ? const Color(0xFFE96F43) : const Color(0xFFB7A89C);
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Container(
            width: 22,
            height: 22,
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: color.withOpacity(.13),
              shape: BoxShape.circle,
            ),
            child: Text(
              '$index',
              style: TextStyle(
                color: color,
                fontSize: 11,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  title,
                  style: TextStyle(
                    color: active
                        ? const Color(0xFF3F2B22)
                        : const Color(0xFF6F5B50),
                    fontSize: 13,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  desc,
                  style: const TextStyle(
                    color: Color(0xFF9B897D),
                    fontSize: 11,
                    height: 1.35,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ActionButton extends StatelessWidget {
  const _ActionButton({
    required this.label,
    required this.onTap,
    this.filled = false,
  });

  final String label;
  final VoidCallback onTap;
  final bool filled;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(13),
        child: Container(
          height: 40,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: filled ? const Color(0xFFE96F43) : Colors.white,
            borderRadius: BorderRadius.circular(13),
            border: filled ? null : Border.all(color: const Color(0xFFEAD7C9)),
          ),
          child: Text(
            label,
            style: TextStyle(
              color: filled ? Colors.white : const Color(0xFF6F5B50),
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
      ),
    );
  }
}

class _StageTab {
  const _StageTab({
    required this.label,
    required this.kind,
  });

  final String label;
  final _SupervisionKind? kind;
}

enum _SupervisionKind { iep, record, adjust }

extension on _SupervisionKind {
  String get label {
    switch (this) {
      case _SupervisionKind.iep:
        return 'IEP';
      case _SupervisionKind.record:
        return '记录';
      case _SupervisionKind.adjust:
        return '调整';
    }
  }
}

class _SupervisionItem {
  const _SupervisionItem({
    required this.id,
    required this.kind,
    required this.childName,
    required this.age,
    required this.className,
    required this.teacherName,
    required this.dueText,
    required this.status,
    required this.target,
    required this.baseline,
    required this.criteria,
    required this.recentRecord,
    required this.note,
    required this.accent,
  });

  final String id;
  final _SupervisionKind kind;
  final String childName;
  final int age;
  final String className;
  final String teacherName;
  final String dueText;
  final String status;
  final String target;
  final String baseline;
  final String criteria;
  final String recentRecord;
  final String note;
  final Color accent;
}
