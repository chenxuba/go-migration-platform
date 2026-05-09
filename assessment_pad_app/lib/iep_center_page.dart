import 'package:flutter/material.dart';

class IepCenterPage extends StatelessWidget {
  const IepCenterPage({required this.onBack, super.key});

  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width =
            constraints.maxWidth.isFinite ? constraints.maxWidth : 1366;
        final _IepMetrics metrics = _IepMetrics.forWidth(width);

        return ColoredBox(
          color: _IepColors.page,
          child: Stack(
            children: <Widget>[
              Positioned.fill(child: CustomPaint(painter: _IepPagePainter())),
              Positioned(
                left: 0,
                top: 0,
                right: 0,
                child: _IepTopBar(onBack: onBack, metrics: metrics),
              ),
              Positioned(
                left: metrics.outer,
                top: 84,
                width: metrics.leftWidth,
                height: 660,
                child: const _StudentQueuePanel(),
              ),
              Positioned(
                left: metrics.contentLeft,
                top: 84,
                width: metrics.contentWidth,
                height: 660,
                child: const _IepWorkspace(),
              ),
            ],
          ),
        );
      },
    );
  }
}

class _IepMetrics {
  const _IepMetrics({
    required this.outer,
    required this.gap,
    required this.leftWidth,
    required this.contentLeft,
    required this.contentWidth,
    required this.compact,
  });

  factory _IepMetrics.forWidth(double width) {
    final bool compact = width < 1180;
    final double outer = compact ? 14 : 24;
    final double gap = compact ? 10 : 14;
    final double leftWidth = compact ? 246 : 284;
    final double contentLeft = outer + leftWidth + gap;
    final double contentWidth = width - outer * 2 - leftWidth - gap;
    return _IepMetrics(
      outer: outer,
      gap: gap,
      leftWidth: leftWidth,
      contentLeft: contentLeft,
      contentWidth: contentWidth,
      compact: compact,
    );
  }

  final double outer;
  final double gap;
  final double leftWidth;
  final double contentLeft;
  final double contentWidth;
  final bool compact;
}

class _IepColors {
  static const Color page = Color(0xFFFFF6EC);
  static const Color surface = Color(0xFFFFFEFB);
  static const Color ink = Color(0xFF3E2A22);
  static const Color text = Color(0xFF72594D);
  static const Color muted = Color(0xFFB39B8C);
  static const Color line = Color(0xFFF0D9C8);
  static const Color lightLine = Color(0xFFF6E8DD);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95735);
  static const Color orangeSoft = Color(0xFFFFEEE4);
  static const Color green = Color(0xFF76A971);
  static const Color greenSoft = Color(0xFFEAF4E5);
  static const Color yellow = Color(0xFFE6A93A);
  static const Color yellowSoft = Color(0xFFFFF3D8);
}

List<BoxShadow> _iepShadow({
  Color color = const Color(0x16B05F32),
  double blur = 18,
  Offset offset = const Offset(0, 9),
}) {
  return <BoxShadow>[
    BoxShadow(color: color, blurRadius: blur, offset: offset),
  ];
}

class _IepPagePainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint cream = Paint()..color = const Color(0xFFFFE7C6);
    final Paint pale = Paint()..color = const Color(0xFFFFFBF4);
    canvas.drawCircle(Offset(size.width * .02, -100), 275, cream);
    canvas.drawOval(
      Rect.fromLTWH(size.width * .66, size.height - 76, 430, 210),
      cream..color = const Color(0xFFFFEED1),
    );
    canvas.drawOval(
      Rect.fromLTWH(-90, size.height - 38, 520, 150),
      pale,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _IepTopBar extends StatelessWidget {
  const _IepTopBar({required this.onBack, required this.metrics});

  final VoidCallback onBack;
  final _IepMetrics metrics;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 72,
      padding: EdgeInsets.symmetric(horizontal: metrics.outer),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.86),
        border: Border(
          bottom: BorderSide(color: _IepColors.line.withOpacity(.74)),
        ),
      ),
      child: Row(
        children: <Widget>[
          _IepBackButton(onTap: onBack),
          const SizedBox(width: 15),
          const Text(
            'IEP中心',
            style: TextStyle(
              color: _IepColors.ink,
              fontSize: 25,
              fontWeight: FontWeight.w900,
              height: 1,
            ),
          ),
          const Spacer(),
          if (!metrics.compact) ...<Widget>[
            const _TopSelector(label: '近30天', width: 108),
            const SizedBox(width: 12),
          ],
          _SearchBox(width: metrics.compact ? 194 : 224),
          const SizedBox(width: 12),
          const _GhostActionButton(icon: Icons.tune_rounded, label: '筛选'),
          const SizedBox(width: 12),
          const _SoftActionButton(
              icon: Icons.file_download_outlined, label: '导出Word'),
          const SizedBox(width: 10),
          const _PrimaryActionButton(
              icon: Icons.auto_fix_high_rounded, label: '生成IEP'),
        ],
      ),
    );
  }
}

class _IepBackButton extends StatelessWidget {
  const _IepBackButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: _IepColors.surface.withOpacity(.94),
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          width: 42,
          height: 42,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: _IepColors.line),
          ),
          child: const Icon(
            Icons.chevron_left_rounded,
            color: _IepColors.text,
            size: 28,
          ),
        ),
      ),
    );
  }
}

class _TopSelector extends StatelessWidget {
  const _TopSelector({required this.label, required this.width});

  final String label;
  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.9),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _IepColors.line),
      ),
      child: Row(
        children: <Widget>[
          Expanded(
            child: Text(
              label,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _IepColors.text,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          const Icon(
            Icons.keyboard_arrow_down_rounded,
            color: _IepColors.muted,
            size: 20,
          ),
        ],
      ),
    );
  }
}

class _SearchBox extends StatelessWidget {
  const _SearchBox({required this.width});

  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.92),
        borderRadius: BorderRadius.circular(19),
        border: Border.all(color: _IepColors.line),
      ),
      child: Row(
        children: const <Widget>[
          Icon(Icons.search_rounded, color: _IepColors.ink, size: 21),
          SizedBox(width: 8),
          Expanded(
            child: Text(
              '搜索学员/评估老师',
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: _IepColors.muted,
                fontSize: 13,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _GhostActionButton extends StatelessWidget {
  const _GhostActionButton({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.92),
        borderRadius: BorderRadius.circular(19),
        border: Border.all(color: _IepColors.line),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: _IepColors.ink, size: 19),
          const SizedBox(width: 5),
          Text(
            label,
            style: const TextStyle(
              color: _IepColors.ink,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _SoftActionButton extends StatelessWidget {
  const _SoftActionButton({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 13),
      decoration: BoxDecoration(
        color: _IepColors.orangeSoft,
        borderRadius: BorderRadius.circular(19),
        border: Border.all(color: const Color(0xFFFFCDB5)),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: _IepColors.orangeDeep, size: 19),
          const SizedBox(width: 5),
          Text(
            label,
            style: const TextStyle(
              color: _IepColors.orangeDeep,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _PrimaryActionButton extends StatelessWidget {
  const _PrimaryActionButton({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 15),
      decoration: BoxDecoration(
        color: _IepColors.orange,
        borderRadius: BorderRadius.circular(19),
        boxShadow: _iepShadow(
          color: const Color(0x2CE96F43),
          blur: 14,
          offset: const Offset(0, 7),
        ),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: Colors.white, size: 19),
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
    );
  }
}

class _StudentQueuePanel extends StatelessWidget {
  const _StudentQueuePanel();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.9),
        borderRadius: BorderRadius.circular(18),
        boxShadow: _iepShadow(),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: const <Widget>[
              Expanded(
                child: Text(
                  '学员IEP队列',
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: _IepColors.ink,
                    fontSize: 17,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              _QueueFilterButton(),
            ],
          ),
          const SizedBox(height: 12),
          const _QueueTabs(),
          const SizedBox(height: 12),
          const _CompactStatsStrip(),
          const SizedBox(height: 12),
          const Expanded(child: _QueueList()),
        ],
      ),
    );
  }
}

class _QueueFilterButton extends StatelessWidget {
  const _QueueFilterButton();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: _IepColors.orangeSoft,
        borderRadius: BorderRadius.circular(15),
      ),
      child: Row(
        children: const <Widget>[
          Icon(Icons.filter_alt_outlined, size: 16, color: _IepColors.orange),
          SizedBox(width: 4),
          Text(
            '筛选',
            style: TextStyle(
              color: _IepColors.orange,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _QueueTabs extends StatelessWidget {
  const _QueueTabs();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 34,
      padding: const EdgeInsets.all(3),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF4EC),
        borderRadius: BorderRadius.circular(17),
      ),
      child: Row(
        children: const <Widget>[
          Expanded(child: _QueueTab(label: '全部', active: true)),
          Expanded(child: _QueueTab(label: '待生成')),
          Expanded(child: _QueueTab(label: '草稿')),
        ],
      ),
    );
  }
}

class _QueueTab extends StatelessWidget {
  const _QueueTab({required this.label, this.active = false});

  final String label;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: active ? Colors.white : Colors.transparent,
        borderRadius: BorderRadius.circular(14),
        boxShadow: active
            ? _iepShadow(
                color: const Color(0x0FB05F32),
                blur: 8,
                offset: const Offset(0, 3),
              )
            : const <BoxShadow>[],
      ),
      child: Text(
        label,
        overflow: TextOverflow.ellipsis,
        style: TextStyle(
          color: active ? _IepColors.orangeDeep : _IepColors.text,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _CompactStatsStrip extends StatelessWidget {
  const _CompactStatsStrip();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        children: const <Widget>[
          Expanded(child: _SmallStat(number: '18', label: '待生成')),
          _StatDivider(),
          Expanded(child: _SmallStat(number: '7', label: '草稿')),
          _StatDivider(),
          Expanded(child: _SmallStat(number: '26', label: '已确认')),
          _StatDivider(),
          Expanded(child: _SmallStat(number: '34', label: '本周计划')),
        ],
      ),
    );
  }
}

class _SmallStat extends StatelessWidget {
  const _SmallStat({required this.number, required this.label});

  final String number;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: <Widget>[
        Text(
          number,
          style: const TextStyle(
            color: _IepColors.ink,
            fontSize: 18,
            fontWeight: FontWeight.w900,
            height: 1,
          ),
        ),
        const SizedBox(height: 6),
        Text(
          label,
          overflow: TextOverflow.ellipsis,
          style: const TextStyle(
            color: _IepColors.muted,
            fontSize: 10,
            fontWeight: FontWeight.w700,
          ),
        ),
      ],
    );
  }
}

class _StatDivider extends StatelessWidget {
  const _StatDivider();

  @override
  Widget build(BuildContext context) {
    return Container(width: 1, height: 28, color: _IepColors.lightLine);
  }
}

class _QueueList extends StatelessWidget {
  const _QueueList();

  static const List<_QueueStudent> _students = <_QueueStudent>[
    _QueueStudent(
      name: '陈旭',
      age: '4岁0月',
      status: '已确认',
      statusColor: _IepColors.green,
      statusBg: _IepColors.greenSoft,
      assessment: '儿心量表 · 2026-05-07',
      period: '05.01-07.31',
      avatarAsset: 'assets/avatars/student_chenxu.png',
      active: true,
    ),
    _QueueStudent(
      name: '陈小宇',
      age: '6岁1月',
      status: '草稿',
      statusColor: _IepColors.yellow,
      statusBg: _IepColors.yellowSoft,
      assessment: '儿心量表 · 2026-05-02',
      period: '05.02-08.01',
      avatarAsset: 'assets/avatars/student_chenxiaoyu.png',
    ),
    _QueueStudent(
      name: '林一诺',
      age: '4岁9月',
      status: '待生成',
      statusColor: _IepColors.orange,
      statusBg: _IepColors.orangeSoft,
      assessment: 'PEP-3 · 2026-04-29',
      period: '待确认周期',
      avatarAsset: 'assets/avatars/student_linyinuo.png',
    ),
    _QueueStudent(
      name: '周书言',
      age: '7岁2月',
      status: '已确认',
      statusColor: _IepColors.green,
      statusBg: _IepColors.greenSoft,
      assessment: 'PEP-3 · 2026-04-26',
      period: '04.26-07.25',
      avatarAsset: 'assets/avatars/student_zhoushuyan.png',
    ),
    _QueueStudent(
      name: '唐沐辰',
      age: '5岁8月',
      status: '草稿',
      statusColor: _IepColors.yellow,
      statusBg: _IepColors.yellowSoft,
      assessment: '儿心量表 · 2026-04-21',
      period: '04.21-07.20',
      avatarAsset: 'assets/avatars/student_tangmuchen.png',
    ),
  ];

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      physics: const ClampingScrollPhysics(),
      padding: EdgeInsets.zero,
      itemCount: _students.length,
      separatorBuilder: (_, __) => const SizedBox(height: 10),
      itemBuilder: (BuildContext context, int index) {
        return _QueueStudentCard(student: _students[index]);
      },
    );
  }
}

class _QueueStudent {
  const _QueueStudent({
    required this.name,
    required this.age,
    required this.status,
    required this.statusColor,
    required this.statusBg,
    required this.assessment,
    required this.period,
    required this.avatarAsset,
    this.active = false,
  });

  final String name;
  final String age;
  final String status;
  final Color statusColor;
  final Color statusBg;
  final String assessment;
  final String period;
  final String avatarAsset;
  final bool active;
}

class _QueueStudentCard extends StatelessWidget {
  const _QueueStudentCard({required this.student});

  final _QueueStudent student;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 80,
      padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 8),
      decoration: BoxDecoration(
        color: student.active ? const Color(0xFFFFF3EB) : Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color:
              student.active ? const Color(0xFFFFB792) : _IepColors.lightLine,
          width: student.active ? 1.4 : 1,
        ),
      ),
      child: Row(
        children: <Widget>[
          _QueueAvatar(asset: student.avatarAsset, active: student.active),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Row(
                  children: <Widget>[
                    Expanded(
                      child: Text(
                        '${student.name} · ${student.age}',
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          color: _IepColors.ink,
                          fontSize: 13,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                    _StatusPill(
                      text: student.status,
                      color: student.statusColor,
                      bg: student.statusBg,
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                Text(
                  student.assessment,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _IepColors.text,
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                    height: 1,
                  ),
                ),
                const SizedBox(height: 6),
                Row(
                  children: <Widget>[
                    const Icon(
                      Icons.date_range_rounded,
                      size: 13,
                      color: _IepColors.muted,
                    ),
                    const SizedBox(width: 4),
                    Expanded(
                      child: Text(
                        student.period,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          color: _IepColors.muted,
                          fontSize: 11,
                          fontWeight: FontWeight.w700,
                          height: 1,
                        ),
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

class _QueueAvatar extends StatelessWidget {
  const _QueueAvatar({required this.asset, required this.active});

  final String asset;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 44,
      height: 44,
      padding: const EdgeInsets.all(2),
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: Colors.white,
        border: Border.all(
          color: active ? const Color(0xFFFFA878) : const Color(0xFFFFDFC8),
          width: active ? 2 : 1.4,
        ),
        boxShadow: active
            ? _iepShadow(
                color: const Color(0x22E96F43),
                blur: 10,
                offset: const Offset(0, 4),
              )
            : const <BoxShadow>[],
      ),
      child: ClipOval(
        child: Image.asset(
          asset,
          width: 40,
          height: 40,
          fit: BoxFit.cover,
          filterQuality: FilterQuality.medium,
        ),
      ),
    );
  }
}

class _StatusPill extends StatelessWidget {
  const _StatusPill({
    required this.text,
    required this.color,
    required this.bg,
  });

  final String text;
  final Color color;
  final Color bg;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 23,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: color,
          fontSize: 10,
          fontWeight: FontWeight.w900,
          height: 1,
        ),
      ),
    );
  }
}

class _IepWorkspace extends StatelessWidget {
  const _IepWorkspace();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(18, 16, 18, 16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.92),
        borderRadius: BorderRadius.circular(18),
        boxShadow: _iepShadow(),
      ),
      child: Column(
        children: const <Widget>[
          _WorkspaceHeader(),
          SizedBox(height: 12),
          _PlanToolbar(),
          SizedBox(height: 10),
          _PlanTabs(),
          SizedBox(height: 12),
          Expanded(child: _IepTablePreview()),
        ],
      ),
    );
  }
}

class _WorkspaceHeader extends StatelessWidget {
  const _WorkspaceHeader();

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 42,
      child: Row(
        children: <Widget>[
          const Expanded(
            child: Text(
              '陈旭 · 康复教学季度计划',
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: _IepColors.ink,
                fontSize: 19,
                fontWeight: FontWeight.w900,
                height: 1,
              ),
            ),
          ),
          const _PeriodPill(text: '3个月', active: true),
          const SizedBox(width: 8),
          const _PeriodPill(text: '6个月'),
          const SizedBox(width: 10),
          const _HeaderMetaPill(icon: Icons.verified_rounded, text: '已确认'),
          const SizedBox(width: 10),
          const _HeaderMetaPill(
            icon: Icons.date_range_rounded,
            text: '2026.05.01-2026.07.31',
          ),
          const SizedBox(width: 12),
          Container(
            height: 36,
            padding: const EdgeInsets.fromLTRB(12, 0, 6, 0),
            decoration: BoxDecoration(
              color: const Color(0xFFFFF6EE),
              borderRadius: BorderRadius.circular(18),
              border: Border.all(color: const Color(0xFFFFD3BA)),
            ),
            child: Row(
              children: const <Widget>[
                Text(
                  '第2月 · 第1周',
                  style: TextStyle(
                    color: _IepColors.orangeDeep,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                SizedBox(width: 9),
                _StartClassButton(),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _PeriodPill extends StatelessWidget {
  const _PeriodPill({required this.text, this.active = false});

  final String text;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: active ? _IepColors.orange : const Color(0xFFFFF8F2),
        borderRadius: BorderRadius.circular(15),
        border: active ? null : Border.all(color: _IepColors.lightLine),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: active ? Colors.white : _IepColors.text,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _HeaderMetaPill extends StatelessWidget {
  const _HeaderMetaPill({required this.icon, required this.text});

  final IconData icon;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF6),
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, size: 15, color: _IepColors.muted),
          const SizedBox(width: 5),
          Text(
            text,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: _IepColors.text,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _StartClassButton extends StatelessWidget {
  const _StartClassButton();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 28,
      padding: const EdgeInsets.symmetric(horizontal: 11),
      decoration: BoxDecoration(
        color: _IepColors.orange,
        borderRadius: BorderRadius.circular(14),
      ),
      child: Row(
        children: const <Widget>[
          Icon(Icons.play_circle_fill_rounded, color: Colors.white, size: 17),
          SizedBox(width: 5),
          Text(
            '开始上课',
            style: TextStyle(
              color: Colors.white,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _PlanToolbar extends StatelessWidget {
  const _PlanToolbar();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 48,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        children: const <Widget>[
          _ToolbarInfo(
              icon: Icons.event_available_rounded,
              label: '制定日期',
              value: '2026.05.07',
              minWidth: 118,
              maxWidth: 118),
          _ToolbarDivider(),
          _ToolbarInfo(
              icon: Icons.group_rounded,
              label: '计划参与者',
              value: '陈瑞',
              minWidth: 132,
              maxWidth: 220),
          _ToolbarDivider(),
          _ToolbarInfo(
              icon: Icons.person_pin_circle_rounded,
              label: '实施者',
              value: '陈瑞',
              minWidth: 94,
              maxWidth: 140),
          Spacer(),
          _TableTinyAction(icon: Icons.edit_calendar_rounded, label: '调整周期'),
          SizedBox(width: 8),
          _TableTinyAction(icon: Icons.save_alt_rounded, label: '保存草稿'),
        ],
      ),
    );
  }
}

class _ToolbarInfo extends StatelessWidget {
  const _ToolbarInfo({
    required this.icon,
    required this.label,
    required this.value,
    required this.minWidth,
    required this.maxWidth,
  });

  final IconData icon;
  final String label;
  final String value;
  final double minWidth;
  final double maxWidth;

  @override
  Widget build(BuildContext context) {
    return ConstrainedBox(
      constraints: BoxConstraints(minWidth: minWidth, maxWidth: maxWidth),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Icon(icon, color: _IepColors.orange, size: 18),
          const SizedBox(width: 7),
          Flexible(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  label,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _IepColors.muted,
                    fontSize: 10,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 1),
                Text(
                  value,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _IepColors.ink,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
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

class _ToolbarDivider extends StatelessWidget {
  const _ToolbarDivider();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 1,
      height: 26,
      margin: const EdgeInsets.only(right: 14),
      color: _IepColors.lightLine,
    );
  }
}

class _TableTinyAction extends StatelessWidget {
  const _TableTinyAction({required this.icon, required this.label});

  final IconData icon;
  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: _IepColors.line),
      ),
      child: Row(
        children: <Widget>[
          Icon(icon, color: _IepColors.text, size: 16),
          const SizedBox(width: 5),
          Text(
            label,
            style: const TextStyle(
              color: _IepColors.text,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _PlanTabs extends StatelessWidget {
  const _PlanTabs();

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 36,
      child: Row(
        children: const <Widget>[
          _PlanTab(text: 'IEP总计划', active: true, width: 96),
          SizedBox(width: 8),
          _PlanTab(text: '第1个月', width: 78),
          SizedBox(width: 8),
          _PlanTab(text: '第2个月', width: 78),
          SizedBox(width: 8),
          _PlanTab(text: '第3个月', width: 78),
          SizedBox(width: 8),
          _PlanTab(text: '周计划', width: 74),
          Spacer(),
          _CollapseNavButton(),
        ],
      ),
    );
  }
}

class _PlanTab extends StatelessWidget {
  const _PlanTab({
    required this.text,
    required this.width,
    this.active = false,
  });

  final String text;
  final double width;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 34,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: active ? _IepColors.ink : const Color(0xFFFFFAF6),
        borderRadius: BorderRadius.circular(17),
        border: active ? null : Border.all(color: _IepColors.lightLine),
      ),
      child: Text(
        text,
        overflow: TextOverflow.ellipsis,
        style: TextStyle(
          color: active ? Colors.white : _IepColors.text,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _CollapseNavButton extends StatelessWidget {
  const _CollapseNavButton();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 34,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF6),
        borderRadius: BorderRadius.circular(17),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        children: const <Widget>[
          Icon(Icons.keyboard_double_arrow_left_rounded,
              size: 17, color: _IepColors.muted),
          SizedBox(width: 4),
          Text(
            '折叠导航',
            style: TextStyle(
              color: _IepColors.muted,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _IepTablePreview extends StatelessWidget {
  const _IepTablePreview();

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(14, 14, 14, 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _IepColors.line),
      ),
      child: Column(
        children: const <Widget>[
          Expanded(
            child: _WordTableFrame(),
          ),
          SizedBox(height: 10),
          _TableFooterBar(),
        ],
      ),
    );
  }
}

class _WordTable extends StatelessWidget {
  const _WordTable();

  static const List<int> _columns = <int>[
    1038,
    1472,
    625,
    877,
    1260,
    1562,
    927,
    2319,
  ];

  @override
  Widget build(BuildContext context) {
    return Column(
      children: const <Widget>[
        _WordTableTitle(),
        _DocTableRow(
          height: 42,
          cells: <_DocCellData>[
            _DocCellData(text: '姓名', columns: 1, bold: true),
            _DocCellData(text: '陈旭', columns: 1),
            _DocCellData(text: '性别', columns: 1, bold: true),
            _DocCellData(text: '-', columns: 1),
            _DocCellData(text: '出生年月', columns: 1, bold: true),
            _DocCellData(text: '2022-05-11', columns: 3, last: true),
          ],
        ),
        _DocTableRow(
          height: 42,
          cells: <_DocCellData>[
            _DocCellData(text: '制定日期', columns: 1, bold: true),
            _DocCellData(text: '2026-05-07', columns: 3),
            _DocCellData(text: '计划参与者', columns: 1, bold: true),
            _DocCellData(text: '陈瑞', columns: 3, last: true),
          ],
        ),
        _DocTableRow(
          height: 42,
          cells: <_DocCellData>[
            _DocCellData(text: '实施者', columns: 1, bold: true),
            _DocCellData(text: '陈瑞', columns: 3),
            _DocCellData(text: '实施\n起止日期', columns: 1, bold: true),
            _DocCellData(
                text: '2026-05-01 至 2026-07-31',
                columns: 3,
                noWrap: true,
                last: true),
          ],
        ),
        _DocTableRow(
          height: 42,
          cells: <_DocCellData>[
            _DocCellData(text: '康复\n领域', columns: 1, bold: true),
            _DocCellData(text: '长期目标', columns: 3, bold: true),
            _DocCellData(text: '短期目标', columns: 2, bold: true),
            _DocCellData(text: '课程\n形式', columns: 1, bold: true),
            _DocCellData(text: '起止日期', columns: 1, bold: true, last: true),
          ],
        ),
        Expanded(
          child: _DocPlanRows(),
        ),
      ],
    );
  }
}

class _WordTableFrame extends StatelessWidget {
  const _WordTableFrame();

  @override
  Widget build(BuildContext context) {
    return Container(
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      foregroundDecoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: const Color(0xFFB98A71), width: 1.2),
      ),
      child: const Padding(
        padding: EdgeInsets.all(1.2),
        child: SingleChildScrollView(
          physics: ClampingScrollPhysics(),
          child: SizedBox(height: 820, child: _WordTable()),
        ),
      ),
    );
  }
}

class _WordTableTitle extends StatelessWidget {
  const _WordTableTitle();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 40,
      alignment: Alignment.center,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(
          bottom: BorderSide(color: Color(0xFFB98A71), width: 1),
        ),
      ),
      child: const Text(
        '康复教学季度计划',
        style: TextStyle(
          color: _IepColors.ink,
          fontSize: 19,
          fontWeight: FontWeight.w900,
          letterSpacing: 1,
        ),
      ),
    );
  }
}

class _DocCellData {
  const _DocCellData({
    required this.text,
    required this.columns,
    this.bold = false,
    this.align = TextAlign.center,
    this.last = false,
    this.noWrap = false,
  });

  final String text;
  final int columns;
  final bool bold;
  final TextAlign align;
  final bool last;
  final bool noWrap;
}

class _DocTableRow extends StatelessWidget {
  const _DocTableRow({required this.height, required this.cells});

  final double height;
  final List<_DocCellData> cells;

  @override
  Widget build(BuildContext context) {
    final List<Widget> children = <Widget>[];
    int columnIndex = 0;
    for (final _DocCellData cell in cells) {
      final int flex = _WordTable._columns
          .skip(columnIndex)
          .take(cell.columns)
          .fold<int>(0, (int sum, int width) => sum + width);
      columnIndex += cell.columns;
      children.add(
        Expanded(
          flex: flex,
          child: _DocCellBox(data: cell),
        ),
      );
    }

    return SizedBox(
      height: height,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: children,
      ),
    );
  }
}

class _DocCellBox extends StatelessWidget {
  const _DocCellBox({
    required this.data,
    this.rowLast = false,
    this.verticalPadding = 5,
  });

  final _DocCellData data;
  final bool rowLast;
  final double verticalPadding;

  @override
  Widget build(BuildContext context) {
    return Container(
      alignment: Alignment.center,
      padding: EdgeInsets.symmetric(horizontal: 8, vertical: verticalPadding),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border(
          right: data.last
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
          bottom: rowLast
              ? BorderSide.none
              : const BorderSide(color: Color(0xFFB98A71), width: .8),
        ),
      ),
      child: Text(
        data.text,
        maxLines: data.noWrap ? 1 : 4,
        overflow: TextOverflow.ellipsis,
        textAlign: data.align,
        style: TextStyle(
          color: data.bold ? _IepColors.ink : _IepColors.text,
          fontSize: 11.4,
          fontWeight: data.bold ? FontWeight.w900 : FontWeight.w700,
          height: 1.22,
        ),
      ),
    );
  }
}

class _DocPlanRows extends StatelessWidget {
  const _DocPlanRows();

  static const List<_DocDomainData> _domains = <_DocDomainData>[
    _DocDomainData(
      domain: '大肌肉',
      longGoals: <String>[
        '1. 提升动态平衡与协调能力，能在移动中稳定控制身体',
        '2. 增强下肢力量与跳跃技能，完成连续跳跃动作',
      ],
      shortGoals: <_DocShortGoalData>[
        _DocShortGoalData('能单脚站立保持平衡5秒以上', '个训', '2026-05-01 - 2026-05-31'),
        _DocShortGoalData('能双脚连续向前跳5步以上', '集体课', '2026-05-01 - 2026-05-31'),
        _DocShortGoalData('能在平衡木上独立行走2米', '个训', '2026-05-01 - 2026-05-31'),
      ],
    ),
    _DocDomainData(
      domain: '小肌肉',
      longGoals: <String>[
        '1. 提高手眼协调与精细操作能力，完成复杂拼插任务',
        '2. 增强手部力量与控制，熟练使用剪刀沿直线裁剪',
      ],
      shortGoals: <_DocShortGoalData>[
        _DocShortGoalData('能独立完成12块以上的拼图', '个训', '2026-05-01 - 2026-05-31'),
        _DocShortGoalData(
            '能用剪刀沿直线剪开10厘米长的纸条', '集体课', '2026-05-01 - 2026-05-31'),
        _DocShortGoalData('能串起直径1厘米的珠子10颗', '个训', '2026-05-01 - 2026-05-31'),
      ],
    ),
    _DocDomainData(
      domain: '情感表达',
      longGoals: <String>[
        '1. 能识别并命名基本情绪，理解情绪产生的原因',
        '2. 在情境中恰当地表达自己的情绪，并用语言描述感受',
      ],
      shortGoals: <_DocShortGoalData>[
        _DocShortGoalData(
            '能指认高兴、生气、伤心、害怕四种基本情绪图片', '集体课', '2026-05-01 - 2026-05-31'),
        _DocShortGoalData(
            '在角色扮演游戏中，能说出角色可能的情绪', '集体课', '2026-05-01 - 2026-05-31'),
        _DocShortGoalData(
            '当自己情绪波动时，能用简单语言表达感受', '个训', '2026-05-01 - 2026-05-31'),
      ],
    ),
    _DocDomainData(
      domain: '模仿（视觉/动作）',
      longGoals: <String>[
        '1. 提高动作模仿的准确性和复杂性，能模仿多步骤动作序列',
        '2. 增强视觉记忆与动作再现能力，完成延迟模仿任务',
      ],
      shortGoals: <_DocShortGoalData>[
        _DocShortGoalData('能模仿3个步骤的粗大动作序列', '个训', '2026-05-01 - 2026-05-31'),
        _DocShortGoalData('能模仿搭建6块积木的造型', '个训', '2026-06-01 - 2026-06-30'),
        _DocShortGoalData(
            '观察教师动作10秒后，能延迟模仿该动作', '集体课', '2026-06-01 - 2026-06-30'),
      ],
    ),
    _DocDomainData(
      domain: '社交互动',
      longGoals: <String>[
        '1. 提升与同伴的合作与轮流意识，能在游戏中遵守规则',
        '2. 增强社交发起与回应能力，主动参与小组活动',
      ],
      shortGoals: <_DocShortGoalData>[
        _DocShortGoalData(
            '在集体游戏中，能等待轮流并遵守简单规则', '集体课', '2026-06-01 - 2026-06-30'),
        _DocShortGoalData('能主动邀请同伴一起玩', '集体课', '2026-06-01 - 2026-06-30'),
        _DocShortGoalData(
            '在角色扮演中，能与同伴合作完成一个场景', '集体课', '2026-06-01 - 2026-06-30'),
      ],
    ),
  ];

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: _domains.asMap().entries.map((entry) {
        return Expanded(
          child: _DocDomainBlock(
            data: entry.value,
            selected: entry.key == 0,
            last: entry.key == _domains.length - 1,
          ),
        );
      }).toList(),
    );
  }
}

class _DocDomainData {
  const _DocDomainData({
    required this.domain,
    required this.longGoals,
    required this.shortGoals,
  });

  final String domain;
  final List<String> longGoals;
  final List<_DocShortGoalData> shortGoals;
}

class _DocShortGoalData {
  const _DocShortGoalData(this.goal, this.lesson, this.period);

  final String goal;
  final String lesson;
  final String period;
}

class _DocDomainBlock extends StatelessWidget {
  const _DocDomainBlock({
    required this.data,
    required this.selected,
    required this.last,
  });

  final _DocDomainData data;
  final bool selected;
  final bool last;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        Expanded(
          flex: _WordTable._columns[0],
          child: _DocMergedCell(text: data.domain, bold: true, rowLast: last),
        ),
        Expanded(
          flex: _WordTable._columns[1] +
              _WordTable._columns[2] +
              _WordTable._columns[3],
          child: _DocMergedCell(
            text: data.longGoals.join('\n'),
            align: TextAlign.left,
            rowLast: last,
          ),
        ),
        Expanded(
          flex: _WordTable._columns[4] + _WordTable._columns[5],
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: data.shortGoals.asMap().entries.map((entry) {
              return Expanded(
                child: _DocCellBox(
                  data: _DocCellData(
                    text: entry.value.goal,
                    columns: 2,
                    align: TextAlign.left,
                  ),
                  rowLast: last && entry.key == data.shortGoals.length - 1,
                  verticalPadding: 4,
                ),
              );
            }).toList(),
          ),
        ),
        Expanded(
          flex: _WordTable._columns[6],
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: data.shortGoals.asMap().entries.map((entry) {
              return Expanded(
                child: _DocCellBox(
                  data: _DocCellData(
                    text: entry.value.lesson,
                    columns: 1,
                    noWrap: true,
                  ),
                  rowLast: last && entry.key == data.shortGoals.length - 1,
                  verticalPadding: 4,
                ),
              );
            }).toList(),
          ),
        ),
        Expanded(
          flex: _WordTable._columns[7],
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: data.shortGoals.asMap().entries.map((entry) {
              return Expanded(
                child: _DocCellBox(
                  data: _DocCellData(
                    text: entry.value.period,
                    columns: 1,
                    noWrap: true,
                    last: true,
                  ),
                  rowLast: last && entry.key == data.shortGoals.length - 1,
                  verticalPadding: 4,
                ),
              );
            }).toList(),
          ),
        ),
      ],
    );
  }
}

class _DocMergedCell extends StatelessWidget {
  const _DocMergedCell({
    required this.text,
    this.bold = false,
    this.align = TextAlign.center,
    this.rowLast = false,
  });

  final String text;
  final bool bold;
  final TextAlign align;
  final bool rowLast;

  @override
  Widget build(BuildContext context) {
    return _DocCellBox(
      data: _DocCellData(
        text: text,
        columns: 1,
        bold: bold,
        align: align,
      ),
      rowLast: rowLast,
      verticalPadding: 6,
    );
  }
}

class _TableFooterBar extends StatelessWidget {
  const _TableFooterBar();

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 30,
      child: Row(
        children: const <Widget>[
          Icon(Icons.edit_note_rounded, color: _IepColors.orange, size: 17),
          SizedBox(width: 6),
          Text(
            '当前选中：大肌肉 / 短期目标 1',
            style: TextStyle(
              color: _IepColors.text,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
          Spacer(),
          Text(
            'Word表格预览',
            style: TextStyle(
              color: _IepColors.muted,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}
