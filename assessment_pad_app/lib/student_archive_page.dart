import 'package:flutter/material.dart';

class StudentArchivePage extends StatelessWidget {
  const StudentArchivePage({required this.onBack, super.key});

  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width =
            constraints.maxWidth.isFinite ? constraints.maxWidth : 1366;
        final bool compact = width < 1180;
        final double outer = compact ? 16 : 24;
        final double gap = compact ? 12 : 16;
        final double leftWidth = compact ? 260 : 286;
        final double rightWidth = compact ? 244 : 260;
        final double centerWidth =
            width - outer * 2 - leftWidth - rightWidth - gap * 2;

        return ColoredBox(
          color: const Color(0xFFFFF6EE),
          child: Padding(
            padding: EdgeInsets.fromLTRB(outer, 18, outer, 18),
            child: Column(
              children: <Widget>[
                _ArchiveTopBar(onBack: onBack),
                const SizedBox(height: 14),
                Expanded(
                  child: Row(
                    children: <Widget>[
                      SizedBox(width: leftWidth, child: const _StudentIndex()),
                      SizedBox(width: gap),
                      SizedBox(
                        width: centerWidth > 0 ? centerWidth : 0,
                        child: const _ArchiveCenter(),
                      ),
                      SizedBox(width: gap),
                      SizedBox(
                          width: rightWidth, child: const _ArchiveRightRail()),
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

class _ArchiveTopBar extends StatelessWidget {
  const _ArchiveTopBar({required this.onBack});

  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        _BackButton(onTap: onBack),
        const SizedBox(width: 14),
        const Text(
          '学员档案',
          style: TextStyle(
            color: Color(0xFF3F2B22),
            fontSize: 24,
            fontWeight: FontWeight.w900,
            height: 1,
          ),
        ),
        const Spacer(),
        const _SearchBox(),
        const SizedBox(width: 10),
        const _TopChip(label: '在训学员', icon: Icons.groups_rounded),
        const SizedBox(width: 10),
        const _TopChip(label: '新建档案', icon: Icons.add_rounded, accent: true),
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
      color: Colors.white.withOpacity(.95),
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: onTap,
        child: Container(
          width: 42,
          height: 42,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: const Color(0xFFE6D4C4)),
          ),
          child: const Icon(
            Icons.chevron_left_rounded,
            size: 30,
            color: Color(0xFF6F5B50),
          ),
        ),
      ),
    );
  }
}

class _SearchBox extends StatelessWidget {
  const _SearchBox();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 304,
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.94),
        borderRadius: BorderRadius.circular(13),
        border: Border.all(color: const Color(0xFFE6D4C4)),
      ),
      child: const Row(
        children: <Widget>[
          Icon(Icons.search_rounded, color: Color(0xFFA7968B), size: 20),
          SizedBox(width: 8),
          Expanded(
            child: Text(
              '搜索学员 / 家长电话',
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: Color(0xFFA7968B),
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _TopChip extends StatelessWidget {
  const _TopChip({
    required this.label,
    required this.icon,
    this.accent = false,
  });

  final String label;
  final IconData icon;
  final bool accent;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: BoxDecoration(
        color: accent ? const Color(0xFFE96F43) : Colors.white.withOpacity(.94),
        borderRadius: BorderRadius.circular(13),
        border: accent ? null : Border.all(color: const Color(0xFFE6D4C4)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Icon(
            icon,
            size: 18,
            color: accent ? Colors.white : const Color(0xFF6F5B50),
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
    );
  }
}

class _StudentIndex extends StatelessWidget {
  const _StudentIndex();

  @override
  Widget build(BuildContext context) {
    return _Panel(
      child: Column(
        children: const <Widget>[
          _PanelHeader(title: '学员索引', subtitle: '共 128 人 · 本月更新 12'),
          Divider(height: 1, color: Color(0xFFF1E3D7)),
          _StatusStrip(),
          Divider(height: 1, color: Color(0xFFF1E3D7)),
          Expanded(
            child: SingleChildScrollView(
              physics: ClampingScrollPhysics(),
              child: Column(
                children: <Widget>[
                  _StudentRow(
                    active: true,
                    name: '陈小宇',
                    tag: '在训',
                    note: 'IEP执行中 · 张老师',
                    color: Color(0xFFE96F43),
                  ),
                  _StudentRow(
                    name: '刘一诺',
                    tag: '待复核',
                    note: 'DTT记录待复核 · 李老师',
                    color: Color(0xFF4B82D1),
                  ),
                  _StudentRow(
                    name: '周子航',
                    tag: '待调整',
                    note: '目标需微调 · 王老师',
                    color: Color(0xFF7D73C9),
                  ),
                  _StudentRow(
                    name: '孙乐乐',
                    tag: '高关注',
                    note: '家校反馈波动 · 陈老师',
                    color: Color(0xFFD76C7D),
                  ),
                  _StudentRow(
                    name: '王晨',
                    tag: '评估',
                    note: '近期评估完成 · 赵老师',
                    color: Color(0xFF63A999),
                  ),
                  _StudentRow(
                    name: '朱思涵',
                    tag: '完成',
                    note: '阶段结案 · 徐老师',
                    color: Color(0xFF9DA4B1),
                  ),
                ],
              ),
            ),
          ),
          Padding(
            padding: EdgeInsets.fromLTRB(14, 0, 14, 14),
            child: _ReminderCard(),
          ),
        ],
      ),
    );
  }
}

class _StatusStrip extends StatelessWidget {
  const _StatusStrip();

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 48,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 14),
        child: Row(
          children: const <Widget>[
            _MiniStatusChip(label: '全部', count: '128', active: true),
            SizedBox(width: 8),
            _MiniStatusChip(label: '在训', count: '92'),
            SizedBox(width: 8),
            _MiniStatusChip(label: '高关注', count: '6'),
          ],
        ),
      ),
    );
  }
}

class _MiniStatusChip extends StatelessWidget {
  const _MiniStatusChip({
    required this.label,
    required this.count,
    this.active = false,
  });

  final String label;
  final String count;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: active ? const Color(0xFFFFF1E8) : Colors.white,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(
          color: active ? const Color(0xFFFFD8BD) : const Color(0xFFF1E3D7),
        ),
      ),
      child: Row(
        children: <Widget>[
          Text(
            label,
            style: TextStyle(
              color: active ? const Color(0xFFC95D37) : const Color(0xFF6F5B50),
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(width: 6),
          Text(
            count,
            style: TextStyle(
              color: active ? const Color(0xFFC95D37) : const Color(0xFF8F7D70),
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _StudentRow extends StatelessWidget {
  const _StudentRow({
    required this.name,
    required this.tag,
    required this.note,
    required this.color,
    this.active = false,
  });

  final String name;
  final String tag;
  final String note;
  final Color color;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 74,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      color: active ? const Color(0xFFFFF7F0) : Colors.transparent,
      child: Row(
        children: <Widget>[
          Container(
            width: 38,
            height: 38,
            alignment: Alignment.center,
            decoration: BoxDecoration(color: color, shape: BoxShape.circle),
            child: Text(
              name.substring(0, 1),
              style: const TextStyle(
                color: Colors.white,
                fontSize: 16,
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
                    Expanded(
                      child: Text(
                        name,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          color: Color(0xFF3F2B22),
                          fontSize: 14,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                    _TagChip(text: tag, color: color),
                  ],
                ),
                const SizedBox(height: 4),
                Text(
                  note,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: Color(0xFF8D7B6E),
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

class _TagChip extends StatelessWidget {
  const _TagChip({required this.text, required this.color});

  final String text;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 22,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: color.withOpacity(.12),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: color,
          fontSize: 10,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _ReminderCard extends StatelessWidget {
  const _ReminderCard();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFBF7),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFF1E3D7)),
      ),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            '本周提醒',
            style: TextStyle(
              color: Color(0xFF3F2B22),
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
          SizedBox(height: 8),
          Text(
            '优先查看高关注学员的 IEP 调整，及时补齐最近两次 DTT 复核。',
            style: TextStyle(
              color: Color(0xFF7A6659),
              fontSize: 12,
              height: 1.5,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}

class _ArchiveCenter extends StatelessWidget {
  const _ArchiveCenter();

  @override
  Widget build(BuildContext context) {
    return _Panel(
      child: Column(
        children: const <Widget>[
          _ArchiveHeader(),
          Divider(height: 1, color: Color(0xFFF1E3D7)),
          _ArchiveTabs(),
          Divider(height: 1, color: Color(0xFFF1E3D7)),
          Expanded(
            child: SingleChildScrollView(
              physics: ClampingScrollPhysics(),
              padding: EdgeInsets.all(14),
              child: Column(
                children: <Widget>[
                  _ProfileBlock(),
                  SizedBox(height: 10),
                  _GridSummary(),
                  SizedBox(height: 10),
                  _TimelineBlock(),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ArchiveHeader extends StatelessWidget {
  const _ArchiveHeader();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: <Widget>[
          Container(
            width: 50,
            height: 50,
            alignment: Alignment.center,
            decoration: const BoxDecoration(
              color: Color(0xFFE96F43),
              shape: BoxShape.circle,
            ),
            child: const Text(
              '陈',
              style: TextStyle(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: const <Widget>[
                Text(
                  '陈小宇 · 6岁',
                  style: TextStyle(
                    color: Color(0xFF3F2B22),
                    fontSize: 20,
                    fontWeight: FontWeight.w900,
                    height: 1,
                  ),
                ),
                SizedBox(height: 5),
                Row(
                  children: <Widget>[
                    Flexible(
                      child: Text(
                        '入训 2026-03-12 · 主责张老师',
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(
                          color: Color(0xFF7A6659),
                          fontSize: 11,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ),
                    SizedBox(width: 10),
                    SingleChildScrollView(
                      scrollDirection: Axis.horizontal,
                      physics: ClampingScrollPhysics(),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: <Widget>[
                          _StateBadge(label: '在训', color: Color(0xFF6F9F70)),
                          SizedBox(width: 6),
                          _StateBadge(
                            label: 'IEP执行中',
                            color: Color(0xFFE96F43),
                          ),
                          SizedBox(width: 6),
                          _StateBadge(
                            label: '记录待复核',
                            color: Color(0xFF4B82D1),
                          ),
                        ],
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

class _ArchiveTabs extends StatelessWidget {
  const _ArchiveTabs();

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 48,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 14),
        child: Row(
          children: const <Widget>[
            _MiniTab(label: '概览', active: true),
            SizedBox(width: 8),
            _MiniTab(label: 'IEP'),
            SizedBox(width: 8),
            _MiniTab(label: '评估'),
            SizedBox(width: 8),
            _MiniTab(label: 'DTT'),
            SizedBox(width: 8),
            _MiniTab(label: '报告'),
          ],
        ),
      ),
    );
  }
}

class _MiniTab extends StatelessWidget {
  const _MiniTab({required this.label, this.active = false});

  final String label;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: active ? const Color(0xFFFFF1E8) : Colors.white,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(
          color: active ? const Color(0xFFFFD8BD) : const Color(0xFFF1E3D7),
        ),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: active ? const Color(0xFFC95D37) : const Color(0xFF6F5B50),
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _ProfileBlock extends StatelessWidget {
  const _ProfileBlock();

  @override
  Widget build(BuildContext context) {
    return _SectionCard(
      title: '当前档案摘要',
      child: Column(
        children: const <Widget>[
          _SummaryRow(
            label: '当前重点',
            value: '主动表达需求，降低完整口头提示依赖',
          ),
          SizedBox(height: 10),
          _SummaryRow(
            label: '责任教师',
            value: '张老师 · 语言表达',
          ),
          SizedBox(height: 10),
          _SummaryRow(
            label: '最近动作',
            value: '完成一次 DTT 复核，等待下次 IEP 调整',
          ),
        ],
      ),
    );
  }
}

class _GridSummary extends StatelessWidget {
  const _GridSummary();

  @override
  Widget build(BuildContext context) {
    return _SectionCard(
      title: '关键进度',
      child: Row(
        children: const <Widget>[
          Expanded(
            child: _SmallMetricCard(title: 'IEP周期', value: '62%'),
          ),
          SizedBox(width: 10),
          Expanded(
            child: _SmallMetricCard(title: '记录完成', value: '78%'),
          ),
          SizedBox(width: 10),
          Expanded(
            child: _SmallMetricCard(title: '目标达成', value: '58%'),
          ),
        ],
      ),
    );
  }
}

class _SmallMetricCard extends StatelessWidget {
  const _SmallMetricCard({required this.title, required this.value});

  final String title;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 88,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFF1E3D7)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: <Widget>[
          Text(
            title,
            style: const TextStyle(
              color: Color(0xFF7A6659),
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
          Text(
            value,
            style: const TextStyle(
              color: Color(0xFF3F2B22),
              fontSize: 26,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _TimelineBlock extends StatelessWidget {
  const _TimelineBlock();

  @override
  Widget build(BuildContext context) {
    return _SectionCard(
      title: '最近记录',
      child: Column(
        children: const <Widget>[
          _TimelineItem(
            time: '今日',
            title: 'DTT 记录复核',
            desc: '完成率 66% · 提示层级仍偏高',
          ),
          _TimelineItem(
            time: '05-12',
            title: 'PEP-3 评估',
            desc: '评估已完成，结果已入档',
          ),
          _TimelineItem(
            time: '05-08',
            title: 'IEP 目标调整',
            desc: '新增主动表达与等待时长两个观察项',
          ),
        ],
      ),
    );
  }
}

class _TimelineItem extends StatelessWidget {
  const _TimelineItem({
    required this.time,
    required this.title,
    required this.desc,
  });

  final String time;
  final String title;
  final String desc;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          SizedBox(
            width: 52,
            child: Text(
              time,
              style: const TextStyle(
                color: Color(0xFFAA978B),
                fontSize: 11,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          Container(
            width: 8,
            height: 8,
            margin: const EdgeInsets.only(top: 5),
            decoration: const BoxDecoration(
              color: Color(0xFFE96F43),
              shape: BoxShape.circle,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  title,
                  style: const TextStyle(
                    color: Color(0xFF3F2B22),
                    fontSize: 13,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 3),
                Text(
                  desc,
                  style: const TextStyle(
                    color: Color(0xFF7A6659),
                    fontSize: 11,
                    height: 1.4,
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

class _ArchiveRightRail extends StatelessWidget {
  const _ArchiveRightRail();

  @override
  Widget build(BuildContext context) {
    return _Panel(
      child: Column(
        children: const <Widget>[
          _PanelHeader(title: '档案动作', subtitle: '当前：IEP执行中'),
          Divider(height: 1, color: Color(0xFFF1E3D7)),
          Expanded(
            child: SingleChildScrollView(
              physics: ClampingScrollPhysics(),
              padding: EdgeInsets.all(14),
              child: Column(
                children: <Widget>[
                  _ActionSummary(),
                  SizedBox(height: 10),
                  _ActionList(),
                  SizedBox(height: 10),
                  _QuickActionGrid(),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ActionSummary extends StatelessWidget {
  const _ActionSummary();

  @override
  Widget build(BuildContext context) {
    return _SectionCard(
      title: '处理进度',
      child: const Column(
        children: <Widget>[
          _ProgressLine(label: 'IEP周期', value: 0.62, text: '62%'),
          SizedBox(height: 10),
          _ProgressLine(label: '记录完成', value: 0.78, text: '78%'),
          SizedBox(height: 10),
          _ProgressLine(label: '目标达成', value: 0.58, text: '58%'),
        ],
      ),
    );
  }
}

class _ProgressLine extends StatelessWidget {
  const _ProgressLine({
    required this.label,
    required this.value,
    required this.text,
  });

  final String label;
  final double value;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        Row(
          children: <Widget>[
            Expanded(
              child: Text(
                label,
                style: const TextStyle(
                  color: Color(0xFF6F5B50),
                  fontSize: 12,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            Text(
              text,
              style: const TextStyle(
                color: Color(0xFF6F5B50),
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        ClipRRect(
          borderRadius: BorderRadius.circular(999),
          child: LinearProgressIndicator(
            minHeight: 8,
            value: value,
            backgroundColor: const Color(0xFFEAF0F3),
            valueColor: const AlwaysStoppedAnimation<Color>(Color(0xFFE96F43)),
          ),
        ),
      ],
    );
  }
}

class _ActionList extends StatelessWidget {
  const _ActionList();

  @override
  Widget build(BuildContext context) {
    return _SectionCard(
      title: '最近动作',
      child: const Column(
        children: <Widget>[
          _ActionRow(title: '查看 IEP', desc: '当前执行版本', color: Color(0xFFE96F43)),
          _ActionRow(title: '填写 DTT', desc: '补录课堂记录', color: Color(0xFF4B82D1)),
          _ActionRow(title: '生成报告', desc: '导出阶段性摘要', color: Color(0xFF63A999)),
        ],
      ),
    );
  }
}

class _ActionRow extends StatelessWidget {
  const _ActionRow({
    required this.title,
    required this.desc,
    required this.color,
  });

  final String title;
  final String desc;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Container(
        width: double.infinity,
        padding: const EdgeInsets.all(10),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border.all(color: const Color(0xFFF1E3D7)),
        ),
        child: Row(
          children: <Widget>[
            Container(
              width: 28,
              height: 28,
              decoration: BoxDecoration(
                  color: color.withOpacity(.14), shape: BoxShape.circle),
              child: Icon(Icons.bolt_rounded, size: 16, color: color),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    title,
                    style: const TextStyle(
                      color: Color(0xFF3F2B22),
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                  const SizedBox(height: 3),
                  Text(
                    desc,
                    style: const TextStyle(
                      color: Color(0xFF7A6659),
                      fontSize: 11,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _QuickActionGrid extends StatelessWidget {
  const _QuickActionGrid();

  @override
  Widget build(BuildContext context) {
    return _SectionCard(
      title: '快捷入口',
      child: Row(
        children: const <Widget>[
          Expanded(child: _QuickButton(label: '生成档案单', accent: false)),
          SizedBox(width: 10),
          Expanded(child: _QuickButton(label: '填写记录', accent: true)),
        ],
      ),
    );
  }
}

class _QuickButton extends StatelessWidget {
  const _QuickButton({required this.label, required this.accent});

  final String label;
  final bool accent;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 44,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: accent ? const Color(0xFFE96F43) : Colors.white,
        borderRadius: BorderRadius.circular(13),
        border: accent ? null : Border.all(color: const Color(0xFFE6D4C4)),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: accent ? Colors.white : const Color(0xFF6F5B50),
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _StateBadge extends StatelessWidget {
  const _StateBadge({required this.label, required this.color});

  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 22,
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
          fontSize: 10,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _SectionCard extends StatelessWidget {
  const _SectionCard({required this.title, required this.child});

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
        border: Border.all(color: const Color(0xFFF1E3D7)),
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

class _SummaryRow extends StatelessWidget {
  const _SummaryRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        SizedBox(
          width: 76,
          child: Text(
            label,
            style: const TextStyle(
              color: Color(0xFFA7968B),
              fontSize: 11,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
        Expanded(
          child: Text(
            value,
            style: const TextStyle(
              color: Color(0xFF3F2B22),
              fontSize: 12,
              height: 1.45,
              fontWeight: FontWeight.w700,
            ),
          ),
        ),
      ],
    );
  }
}

class _Panel extends StatelessWidget {
  const _Panel({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFE6D4C4)),
      ),
      child: child,
    );
  }
}

class _PanelHeader extends StatelessWidget {
  const _PanelHeader({required this.title, required this.subtitle});

  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      height: 60,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      alignment: Alignment.centerLeft,
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
              color: Color(0xFFA7968B),
              fontSize: 11,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}
