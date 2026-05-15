part of 'assessment_report_list_page.dart';

enum _AutismDevReportTab {
  assessmentInfo,
  resultAnalysis,
  strengthWeakness,
  training,
  iepPlan,
  developmentProfile,
  behaviorProfile,
}

const List<_AutismDevReportTabSpec> _autismDevReportTabs =
    <_AutismDevReportTabSpec>[
  _AutismDevReportTabSpec('评估情况', _AutismDevReportTab.assessmentInfo),
  _AutismDevReportTabSpec('评估结果分析', _AutismDevReportTab.resultAnalysis),
  _AutismDevReportTabSpec('优劣势分析', _AutismDevReportTab.strengthWeakness),
  _AutismDevReportTabSpec('训练情况', _AutismDevReportTab.training),
  _AutismDevReportTabSpec('IEP训练计划', _AutismDevReportTab.iepPlan),
  _AutismDevReportTabSpec('发展情况剖面图', _AutismDevReportTab.developmentProfile),
  _AutismDevReportTabSpec('情绪行为表现图', _AutismDevReportTab.behaviorProfile),
];

const String _autismDevDevelopmentProfileAsset =
    'assets/reports/autismdev_development_profile.png';
const String _autismDevBehaviorProfileAsset =
    'assets/reports/autismdev_behavior_profile.png';

class _AutismDevReportPreviewDialog extends StatefulWidget {
  const _AutismDevReportPreviewDialog({required this.record});

  final Pep3RecordSummary record;

  @override
  State<_AutismDevReportPreviewDialog> createState() =>
      _AutismDevReportPreviewDialogState();
}

class _AutismDevReportPreviewDialogState
    extends State<_AutismDevReportPreviewDialog> {
  _AutismDevReportTab _activeTab = _AutismDevReportTab.assessmentInfo;

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      child: Center(
        child: Material(
          color: Colors.transparent,
          child: Container(
            width: 980,
            height: 654,
            padding: const EdgeInsets.fromLTRB(22, 22, 22, 22),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(color: _ReportTheme.line),
              boxShadow: const <BoxShadow>[
                BoxShadow(
                  color: Color(0x24000000),
                  blurRadius: 34,
                  offset: Offset(0, 18),
                ),
              ],
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                _buildHeader(context),
                const SizedBox(height: 14),
                _buildTabBar(),
                const SizedBox(height: 12),
                Expanded(child: _buildContent()),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildHeader(BuildContext context) {
    final Pep3RecordSummary record = widget.record;
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              const Text(
                '孤独症儿童发展评估报告',
                style: TextStyle(
                  color: _ReportTheme.ink,
                  fontSize: 24,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 7),
              Text(
                '${record.assessmentName.trim().isEmpty ? '孤独症儿童发展评估表' : record.assessmentName}   ${_studentName(record)} / ${_dateOnlyText(record.assessmentDate)}',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _ReportTheme.muted,
                  fontSize: 14,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(width: 16),
        Material(
          color: Colors.transparent,
          shape: const CircleBorder(),
          child: InkWell(
            onTap: () => Navigator.of(context).pop(),
            customBorder: const CircleBorder(),
            child: Container(
              width: 38,
              height: 38,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: const Color(0xFFFFF8F2),
                shape: BoxShape.circle,
                border: Border.all(color: _ReportTheme.lineSoft),
              ),
              child: const Icon(
                Icons.close_rounded,
                size: 22,
                color: _ReportTheme.muted,
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildTabBar() {
    return _AutismDevReportTabBar(
      activeTab: _activeTab,
      onSelected: _selectTab,
    );
  }

  void _selectTab(_AutismDevReportTab tab) {
    if (_activeTab == tab) {
      return;
    }
    setState(() {
      _activeTab = tab;
    });
  }

  Widget _buildContent() {
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFFDF8F3),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      clipBehavior: Clip.antiAlias,
      child: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(18, 16, 18, 18),
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 880),
            child: DecoratedBox(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: _ReportTheme.lineSoft),
                boxShadow: const <BoxShadow>[
                  BoxShadow(
                    color: Color(0x0F000000),
                    blurRadius: 16,
                    offset: Offset(0, 8),
                  ),
                ],
              ),
              child: Padding(
                padding: const EdgeInsets.fromLTRB(24, 22, 24, 24),
                child: _buildReportPage(),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildReportPage() {
    final bool showHeader =
        _activeTab != _AutismDevReportTab.developmentProfile &&
            _activeTab != _AutismDevReportTab.behaviorProfile;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        if (showHeader) ...<Widget>[
          _AutismDevReportPageHeader(record: widget.record),
          const SizedBox(height: 18),
        ],
        _buildActiveSection(),
      ],
    );
  }

  Widget _buildActiveSection() {
    switch (_activeTab) {
      case _AutismDevReportTab.assessmentInfo:
        return const _AutismDevOverviewSection();
      case _AutismDevReportTab.resultAnalysis:
        return const _AutismDevAnalysisSection();
      case _AutismDevReportTab.strengthWeakness:
        return const _AutismDevStrengthWeaknessSection();
      case _AutismDevReportTab.training:
        return const _AutismDevTrainingSection();
      case _AutismDevReportTab.iepPlan:
        return const _AutismDevIepPlanSection();
      case _AutismDevReportTab.developmentProfile:
        return const _AutismDevDevelopmentProfileSection();
      case _AutismDevReportTab.behaviorProfile:
        return const _AutismDevBehaviorProfileSection();
    }
  }
}

class _AutismDevReportTabBar extends StatelessWidget {
  const _AutismDevReportTabBar({
    required this.activeTab,
    required this.onSelected,
  });

  final _AutismDevReportTab activeTab;
  final ValueChanged<_AutismDevReportTab> onSelected;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: <Widget>[
            for (int index = 0; index < _autismDevReportTabs.length; index++)
              Padding(
                padding: EdgeInsets.only(
                  right: index == _autismDevReportTabs.length - 1 ? 0 : 8,
                ),
                child: _ErxinReportTabChip(
                  label: _autismDevReportTabs[index].label,
                  active: activeTab == _autismDevReportTabs[index].tab,
                  onTap: () => onSelected(_autismDevReportTabs[index].tab),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _AutismDevReportPageHeader extends StatelessWidget {
  const _AutismDevReportPageHeader({required this.record});

  final Pep3RecordSummary record;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: <Widget>[
        const Center(
          child: Text(
            '孤独症儿童发展评估报告',
            style: TextStyle(
              color: _ReportTheme.ink,
              fontSize: 24,
              height: 1.2,
              fontWeight: FontWeight.w900,
            ),
          ),
        ),
        const SizedBox(height: 16),
        _AutismDevInfoGrid(
          items: <_AutismDevInfoItem>[
            _AutismDevInfoItem('儿童姓名', _studentName(record)),
            _AutismDevInfoItem('测评年龄', _ageText(record)),
            _AutismDevInfoItem('测评日期', _dateOnlyText(record.assessmentDate)),
            _AutismDevInfoItem(
              '评估者',
              record.examinerName.trim().isEmpty
                  ? '-'
                  : record.examinerName.trim(),
            ),
            _AutismDevInfoItem(
                '量表版本',
                record.scaleVersion.trim().isEmpty
                    ? '2010修订训练师版'
                    : record.scaleVersion.trim()),
            _AutismDevInfoItem(
                '测评次数', _sequenceText(record.assessmentSequence)),
          ],
        ),
      ],
    );
  }
}

class _AutismDevOverviewSection extends StatelessWidget {
  const _AutismDevOverviewSection();

  @override
  Widget build(BuildContext context) {
    final int developmentP = _autismDevDevelopmentScores.fold<int>(
      0,
      (int total, _AutismDevDevelopmentScore item) => total + item.p,
    );
    final int developmentTotal = _autismDevDevelopmentScores.fold<int>(
      0,
      (int total, _AutismDevDevelopmentScore item) => total + item.total,
    );
    final int trainingTargets = _autismDevDevelopmentScores.fold<int>(
      0,
      (int total, _AutismDevDevelopmentScore item) => total + item.e,
    );
    final int behaviorAttention = _autismDevBehaviorScores.fold<int>(
      0,
      (int total, _AutismDevBehaviorScore item) => total + item.m + item.s,
    );

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Wrap(
          spacing: 12,
          runSpacing: 12,
          children: <Widget>[
            _AutismDevOverviewMetric(
              label: '发展领域通过项',
              value: '$developmentP',
              suffix: '/ $developmentTotal',
              color: _ReportTheme.blue,
            ),
            _AutismDevOverviewMetric(
              label: '训练目标候选项',
              value: '$trainingTargets',
              suffix: '个E项',
              color: _ReportTheme.orangeDeep,
            ),
            _AutismDevOverviewMetric(
              label: '情绪行为需关注',
              value: '$behaviorAttention',
              suffix: '项',
              color: _ReportTheme.rose,
            ),
          ],
        ),
        const SizedBox(height: 18),
        const _AutismDevSectionTitle(
          title: '发展能力计分汇总表',
          subtitle: '七个发展领域按 P / E+F(X) 汇总；情绪与行为按 A / M / S 汇总。',
        ),
        const SizedBox(height: 10),
        const _AutismDevDevelopmentScoreTable(),
        const SizedBox(height: 18),
        const _AutismDevSectionTitle(
          title: '结果摘要',
          subtitle: '根据各领域通过项、中间反应项和情绪行为分布整理。',
        ),
        const SizedBox(height: 10),
        const _AutismDevReportSummaryGrid(),
      ],
    );
  }
}

class _AutismDevDevelopmentSection extends StatelessWidget {
  const _AutismDevDevelopmentSection();

  @override
  Widget build(BuildContext context) {
    return const Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _AutismDevSectionTitle(
          title: '发展能力柱状图',
          subtitle: 'P 项表示当前能力现状；E 项优先转化为个别化训练目标。',
        ),
        SizedBox(height: 12),
        _AutismDevDevelopmentBarChart(),
        SizedBox(height: 18),
        _AutismDevSectionTitle(
          title: '发展领域明细',
          subtitle: '汇总各领域 P、E、F、X 与通过率。',
        ),
        SizedBox(height: 10),
        _AutismDevDevelopmentDetailTable(),
      ],
    );
  }
}

class _AutismDevBehaviorSection extends StatelessWidget {
  const _AutismDevBehaviorSection();

  @override
  Widget build(BuildContext context) {
    return const Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _AutismDevSectionTitle(
          title: '情绪与行为发展能力柱状图',
          subtitle: 'A 为无异常，M 为轻度异常，S 为重度异常。',
        ),
        SizedBox(height: 12),
        _AutismDevBehaviorBarChart(),
        SizedBox(height: 18),
        _AutismDevSectionTitle(
          title: '情绪行为汇总',
          subtitle: '按手册七个范围展示 A / M / S 的数量分布。',
        ),
        SizedBox(height: 10),
        _AutismDevBehaviorScoreTable(),
      ],
    );
  }
}

class _AutismDevStrengthWeaknessSection extends StatelessWidget {
  const _AutismDevStrengthWeaknessSection();

  @override
  Widget build(BuildContext context) {
    return const Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _AutismDevSectionTitle(
          title: '优劣势分析',
          subtitle: '按发展领域整理优势能力和当前支持重点。',
        ),
        SizedBox(height: 10),
        _AutismDevStrengthWeaknessTable(),
      ],
    );
  }
}

class _AutismDevTrainingSection extends StatelessWidget {
  const _AutismDevTrainingSection();

  @override
  Widget build(BuildContext context) {
    return const Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _AutismDevSectionTitle(
          title: '训练情况',
          subtitle: '根据评估结果中的中间反应项整理训练目标候选。',
        ),
        SizedBox(height: 10),
        _AutismDevTrainingTargetTable(),
      ],
    );
  }
}

class _AutismDevIepPlanSection extends StatelessWidget {
  const _AutismDevIepPlanSection();

  @override
  Widget build(BuildContext context) {
    return const Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _AutismDevSectionTitle(
          title: 'IEP训练计划',
          subtitle: '围绕优先领域拆分阶段目标和训练内容。',
        ),
        SizedBox(height: 10),
        _AutismDevIepPlanTable(),
      ],
    );
  }
}

class _AutismDevDevelopmentProfileSection extends StatelessWidget {
  const _AutismDevDevelopmentProfileSection();

  @override
  Widget build(BuildContext context) {
    return const _AutismDevReportFigure(
      assetPath: _autismDevDevelopmentProfileAsset,
      overlayPainter: _AutismDevDevelopmentProfilePainter(),
    );
  }
}

class _AutismDevBehaviorProfileSection extends StatelessWidget {
  const _AutismDevBehaviorProfileSection();

  @override
  Widget build(BuildContext context) {
    return const _AutismDevReportFigure(
      assetPath: _autismDevBehaviorProfileAsset,
      overlayPainter: _AutismDevBehaviorProfilePainter(),
    );
  }
}

class _AutismDevAnalysisSection extends StatelessWidget {
  const _AutismDevAnalysisSection();

  @override
  Widget build(BuildContext context) {
    return const Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _AutismDevSectionTitle(
          title: '孤独症儿童评估结果分析表',
          subtitle: '按领域归纳能力现状、优劣势与训练目标。',
        ),
        SizedBox(height: 10),
        _AutismDevAnalysisTable(),
      ],
    );
  }
}

class _AutismDevSectionTitle extends StatelessWidget {
  const _AutismDevSectionTitle({
    required this.title,
    required this.subtitle,
  });

  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Container(
          width: 4,
          height: 32,
          margin: const EdgeInsets.only(top: 2),
          decoration: BoxDecoration(
            color: _ReportTheme.orange,
            borderRadius: BorderRadius.circular(999),
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
                  color: _ReportTheme.ink,
                  fontSize: 18,
                  height: 1.2,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                subtitle,
                style: const TextStyle(
                  color: _ReportTheme.muted,
                  fontSize: 12,
                  height: 1.35,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}

class _AutismDevInfoGrid extends StatelessWidget {
  const _AutismDevInfoGrid({required this.items});

  final List<_AutismDevInfoItem> items;

  @override
  Widget build(BuildContext context) {
    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: <Widget>[
        for (final _AutismDevInfoItem item in items)
          SizedBox(
            width: 264,
            child: _AutismDevInfoCell(item: item),
          ),
      ],
    );
  }
}

class _AutismDevInfoCell extends StatelessWidget {
  const _AutismDevInfoCell({required this.item});

  final _AutismDevInfoItem item;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 46,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          Text(
            item.label,
            style: const TextStyle(
              color: _ReportTheme.muted,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              item.value.trim().isEmpty ? '-' : item.value.trim(),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.right,
              style: const TextStyle(
                color: _ReportTheme.ink,
                fontSize: 14,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevOverviewMetric extends StatelessWidget {
  const _AutismDevOverviewMetric({
    required this.label,
    required this.value,
    required this.suffix,
    required this.color,
  });

  final String label;
  final String value;
  final String suffix;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 264,
      height: 82,
      padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            label,
            style: const TextStyle(
              color: _ReportTheme.muted,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
          const Spacer(),
          Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: <Widget>[
              Text(
                value,
                style: TextStyle(
                  color: color,
                  fontSize: 30,
                  height: .9,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(width: 7),
              Padding(
                padding: const EdgeInsets.only(bottom: 2),
                child: Text(
                  suffix,
                  style: const TextStyle(
                    color: _ReportTheme.text,
                    fontSize: 13,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _AutismDevReportSummaryGrid extends StatelessWidget {
  const _AutismDevReportSummaryGrid();

  @override
  Widget build(BuildContext context) {
    const List<_AutismDevSummaryItem> items = <_AutismDevSummaryItem>[
      _AutismDevSummaryItem(
        '能力现状',
        '认知、生活自理和感知觉通过项占比较高。',
        Icons.trending_up_rounded,
      ),
      _AutismDevSummaryItem(
        '目标候选',
        '语言表达、精细动作和社会交往E项较集中。',
        Icons.flag_rounded,
      ),
      _AutismDevSummaryItem(
        '行为关注',
        '感觉偏好、情绪调节和特殊行为需要持续观察。',
        Icons.visibility_rounded,
      ),
      _AutismDevSummaryItem(
        '训练建议',
        '优先从E项转化短期目标，再补充F项前置能力。',
        Icons.assignment_turned_in_rounded,
      ),
    ];
    return Wrap(
      spacing: 10,
      runSpacing: 10,
      children: <Widget>[
        for (final _AutismDevSummaryItem item in items)
          SizedBox(
            width: 198,
            child: _AutismDevSummaryTile(item: item),
          ),
      ],
    );
  }
}

class _AutismDevSummaryTile extends StatelessWidget {
  const _AutismDevSummaryTile({required this.item});

  final _AutismDevSummaryItem item;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 84,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          Container(
            width: 34,
            height: 34,
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: const Color(0xFFFFF1E8),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(
              item.icon,
              color: _ReportTheme.orangeDeep,
              size: 18,
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  item.title,
                  style: const TextStyle(
                    color: _ReportTheme.ink,
                    fontSize: 14,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 5),
                Text(
                  item.description,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _ReportTheme.muted,
                    fontSize: 11,
                    height: 1.25,
                    fontWeight: FontWeight.w800,
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

class _AutismDevDevelopmentScoreTable extends StatelessWidget {
  const _AutismDevDevelopmentScoreTable();

  @override
  Widget build(BuildContext context) {
    return _AutismDevTableFrame(
      child: Table(
        columnWidths: const <int, TableColumnWidth>{
          0: FlexColumnWidth(1.7),
          1: FlexColumnWidth(1),
          2: FlexColumnWidth(1),
          3: FlexColumnWidth(1),
        },
        border: TableBorder.all(color: _ReportTheme.lineSoft),
        children: <TableRow>[
          _autismDevTableHeaderRow(<String>['领域', 'P', 'E+F(X)', '总分']),
          for (final _AutismDevDevelopmentScore item
              in _autismDevDevelopmentScores)
            _autismDevTableRow(
              <String>[
                item.label,
                '${item.p}',
                '${item.supportCount}',
                '${item.p}',
              ],
              emphFirst: true,
            ),
        ],
      ),
    );
  }
}

class _AutismDevDevelopmentDetailTable extends StatelessWidget {
  const _AutismDevDevelopmentDetailTable();

  @override
  Widget build(BuildContext context) {
    return _AutismDevTableFrame(
      child: Table(
        columnWidths: const <int, TableColumnWidth>{
          0: FlexColumnWidth(1.55),
          1: FlexColumnWidth(.7),
          2: FlexColumnWidth(.7),
          3: FlexColumnWidth(.7),
          4: FlexColumnWidth(.7),
          5: FlexColumnWidth(.8),
        },
        border: TableBorder.all(color: _ReportTheme.lineSoft),
        children: <TableRow>[
          _autismDevTableHeaderRow(<String>['领域', 'P', 'E', 'F', 'X', '通过率']),
          for (final _AutismDevDevelopmentScore item
              in _autismDevDevelopmentScores)
            _autismDevTableRow(
              <String>[
                item.label,
                '${item.p}',
                '${item.e}',
                '${item.f}',
                '${item.x}',
                '${item.passRate.toStringAsFixed(0)}%',
              ],
              emphFirst: true,
            ),
        ],
      ),
    );
  }
}

class _AutismDevBehaviorScoreTable extends StatelessWidget {
  const _AutismDevBehaviorScoreTable();

  @override
  Widget build(BuildContext context) {
    return _AutismDevTableFrame(
      child: Table(
        columnWidths: const <int, TableColumnWidth>{
          0: FlexColumnWidth(1.8),
          1: FlexColumnWidth(.8),
          2: FlexColumnWidth(.8),
          3: FlexColumnWidth(.8),
          4: FlexColumnWidth(.9),
        },
        border: TableBorder.all(color: _ReportTheme.lineSoft),
        children: <TableRow>[
          _autismDevTableHeaderRow(<String>['范围', 'A', 'M', 'S', '关注项']),
          for (final _AutismDevBehaviorScore item in _autismDevBehaviorScores)
            _autismDevTableRow(
              <String>[
                item.label,
                '${item.a}',
                '${item.m}',
                '${item.s}',
                '${item.m + item.s}',
              ],
              emphFirst: true,
            ),
        ],
      ),
    );
  }
}

class _AutismDevStrengthWeaknessTable extends StatelessWidget {
  const _AutismDevStrengthWeaknessTable();

  @override
  Widget build(BuildContext context) {
    return _AutismDevTableFrame(
      child: Table(
        columnWidths: const <int, TableColumnWidth>{
          0: FlexColumnWidth(1.05),
          1: FlexColumnWidth(3.6),
        },
        border: TableBorder.all(color: _ReportTheme.lineSoft),
        children: <TableRow>[
          _autismDevTableHeaderRow(<String>['领域', '优劣势分析'], tall: true),
          for (final _AutismDevAnalysisItem item in _autismDevAnalysisItems)
            TableRow(
              children: <Widget>[
                _autismDevTableCell(item.domain, emph: true, minHeight: 66),
                _autismDevTableCell(item.strength, minHeight: 66),
              ],
            ),
        ],
      ),
    );
  }
}

class _AutismDevTrainingTargetTable extends StatelessWidget {
  const _AutismDevTrainingTargetTable();

  @override
  Widget build(BuildContext context) {
    return _AutismDevTableFrame(
      child: Table(
        columnWidths: const <int, TableColumnWidth>{
          0: FlexColumnWidth(1.05),
          1: FlexColumnWidth(3.6),
        },
        border: TableBorder.all(color: _ReportTheme.lineSoft),
        children: <TableRow>[
          _autismDevTableHeaderRow(<String>['领域', '训练目标'], tall: true),
          for (final _AutismDevAnalysisItem item in _autismDevAnalysisItems)
            TableRow(
              children: <Widget>[
                _autismDevTableCell(item.domain, emph: true, minHeight: 66),
                _autismDevTableCell(item.target, minHeight: 66),
              ],
            ),
        ],
      ),
    );
  }
}

class _AutismDevIepPlanTable extends StatelessWidget {
  const _AutismDevIepPlanTable();

  @override
  Widget build(BuildContext context) {
    return _AutismDevTableFrame(
      child: Table(
        columnWidths: const <int, TableColumnWidth>{
          0: FlexColumnWidth(1.05),
          1: FlexColumnWidth(2),
          2: FlexColumnWidth(2),
        },
        border: TableBorder.all(color: _ReportTheme.lineSoft),
        children: <TableRow>[
          _autismDevTableHeaderRow(<String>['领域', '阶段目标', '训练内容'], tall: true),
          for (final _AutismDevAnalysisItem item in _autismDevAnalysisItems)
            TableRow(
              children: <Widget>[
                _autismDevTableCell(item.domain, emph: true, minHeight: 74),
                _autismDevTableCell(item.target, minHeight: 74),
                _autismDevTableCell(item.status, minHeight: 74),
              ],
            ),
        ],
      ),
    );
  }
}

class _AutismDevAnalysisTable extends StatelessWidget {
  const _AutismDevAnalysisTable();

  @override
  Widget build(BuildContext context) {
    return _AutismDevTableFrame(
      child: Table(
        columnWidths: const <int, TableColumnWidth>{
          0: FlexColumnWidth(1.05),
          1: FlexColumnWidth(2.15),
          2: FlexColumnWidth(2.15),
          3: FlexColumnWidth(2.1),
        },
        border: TableBorder.all(color: _ReportTheme.lineSoft),
        children: <TableRow>[
          _autismDevTableHeaderRow(
            <String>['领域', '能力现状描述', '优劣分析', '训练目标'],
            tall: true,
          ),
          for (final _AutismDevAnalysisItem item in _autismDevAnalysisItems)
            TableRow(
              children: <Widget>[
                _autismDevTableCell(item.domain, emph: true, minHeight: 74),
                _autismDevTableCell(item.status, minHeight: 74),
                _autismDevTableCell(item.strength, minHeight: 74),
                _autismDevTableCell(item.target, minHeight: 74),
              ],
            ),
        ],
      ),
    );
  }
}

class _AutismDevReportFigure extends StatelessWidget {
  const _AutismDevReportFigure({
    required this.assetPath,
    required this.overlayPainter,
  });

  final String assetPath;
  final CustomPainter overlayPainter;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 820),
        child: AspectRatio(
          aspectRatio: 1488 / 2103,
          child: Stack(
            fit: StackFit.expand,
            children: <Widget>[
              Image.asset(assetPath, fit: BoxFit.contain),
              CustomPaint(painter: overlayPainter),
            ],
          ),
        ),
      ),
    );
  }
}

class _AutismDevDevelopmentProfilePainter extends CustomPainter {
  const _AutismDevDevelopmentProfilePainter();

  static const double _figureWidth = 1488;
  static const double _figureHeight = 2103;
  static const List<String> _profileDomains = <String>[
    '感知觉',
    '粗大动作',
    '精细动作',
    '语言与沟通',
    '认知',
    '社会交往',
    '生活自理',
  ];
  static const List<double> _profileColumnXs = <double>[
    380,
    478.5,
    581.5,
    681,
    780.5,
    883.5,
    990.5,
  ];
  static const List<double> _profilePointXs = <double>[
    376.5,
    478.5,
    578,
    677.5,
    777,
    883.5,
    987,
  ];
  static const double _developmentTotalX = 1094;
  static const double _developmentTotalPointX = 1090.5;
  static const double _pScoreY = 1650;
  static const double _eScoreY = 1739;

  @override
  void paint(Canvas canvas, Size size) {
    final List<Offset> abilityPoints = <Offset>[
      for (int index = 0; index < _profileDomains.length; index++)
        _point(
          size,
          _profilePointXs[index],
          _scoreYFor(
            _profileDomains[index],
            _developmentScore(_profileDomains[index]).p,
          ),
        ),
      _point(
        size,
        _developmentTotalPointX,
        _scoreYFor('发展分数', _developmentTotalP),
      ),
    ];
    final List<Offset> targetPoints = <Offset>[
      for (int index = 0; index < _profileDomains.length; index++)
        _point(
          size,
          _profilePointXs[index],
          _scoreYFor(
            _profileDomains[index],
            _developmentScore(_profileDomains[index]).p +
                _developmentScore(_profileDomains[index]).e,
          ),
        ),
    ];
    if (abilityPoints.isEmpty || targetPoints.isEmpty) {
      return;
    }
    final double strokeWidth = size.width / 360;
    final Paint abilityLinePaint = Paint()
      ..color = _ReportTheme.blue.withOpacity(.86)
      ..strokeWidth = strokeWidth
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;
    final Paint targetLinePaint = Paint()
      ..color = _ReportTheme.orangeDeep.withOpacity(.84)
      ..strokeWidth = strokeWidth
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;
    final Paint haloPaint = Paint()
      ..color = Colors.white.withOpacity(.92)
      ..style = PaintingStyle.fill;
    final Paint abilityDotPaint = Paint()
      ..color = _ReportTheme.blue
      ..style = PaintingStyle.fill;
    final Paint targetDotPaint = Paint()
      ..color = _ReportTheme.orangeDeep
      ..style = PaintingStyle.fill;
    _drawDashedPath(canvas, _pathFor(targetPoints), targetLinePaint);
    canvas.drawPath(_pathFor(abilityPoints), abilityLinePaint);
    for (final Offset point in targetPoints) {
      canvas.drawCircle(point, strokeWidth * 2.6, haloPaint);
      canvas.drawCircle(point, strokeWidth * 1.55, targetDotPaint);
    }
    for (final Offset point in abilityPoints) {
      canvas.drawCircle(point, strokeWidth * 2.6, haloPaint);
      canvas.drawCircle(point, strokeWidth * 1.65, abilityDotPaint);
    }
    _drawScoreBoxes(canvas, size);
  }

  static _AutismDevDevelopmentScore _developmentScore(String label) {
    return _autismDevDevelopmentScores.firstWhere(
      (_AutismDevDevelopmentScore item) => item.label == label,
    );
  }

  static void _drawScoreBoxes(Canvas canvas, Size size) {
    for (int index = 0; index < _profileDomains.length; index++) {
      final _AutismDevDevelopmentScore score =
          _developmentScore(_profileDomains[index]);
      _drawScoreText(
        canvas,
        size,
        _profileColumnXs[index],
        _pScoreY,
        '${score.p}',
      );
      _drawScoreText(
        canvas,
        size,
        _profileColumnXs[index],
        _eScoreY,
        '${score.e}',
      );
    }
    _drawScoreText(
      canvas,
      size,
      _developmentTotalX,
      _pScoreY,
      '$_developmentTotalP',
    );
  }

  static int get _developmentTotalP => _autismDevDevelopmentScores.fold<int>(
        0,
        (int total, _AutismDevDevelopmentScore item) => total + item.p,
      );

  static void _drawScoreText(
    Canvas canvas,
    Size size,
    double sourceX,
    double sourceY,
    String value,
  ) {
    final double scale = size.width / _figureWidth;
    final TextPainter painter = TextPainter(
      text: TextSpan(
        text: value,
        style: TextStyle(
          color: _ReportTheme.ink,
          fontSize: 28 * scale,
          height: 1,
          fontWeight: FontWeight.w900,
        ),
      ),
      textAlign: TextAlign.center,
      textDirection: TextDirection.ltr,
    )..layout();
    final Offset center = _point(size, sourceX, sourceY);
    painter.paint(
      canvas,
      Offset(center.dx - painter.width / 2, center.dy - painter.height / 2),
    );
  }

  static double _scoreYFor(String domain, int score) {
    final List<_AutismDevProfileScalePoint> points =
        _profileScalePoints[domain] ?? const <_AutismDevProfileScalePoint>[];
    if (points.isEmpty) {
      return 1589;
    }
    final double value = score.toDouble();
    if (value >= points.first.score) {
      return points.first.y;
    }
    for (int index = 0; index < points.length - 1; index++) {
      final _AutismDevProfileScalePoint upper = points[index];
      final _AutismDevProfileScalePoint lower = points[index + 1];
      if (value <= upper.score && value >= lower.score) {
        final double span = upper.score - lower.score;
        if (span <= 0) {
          return upper.y;
        }
        final double progress = (upper.score - value) / span;
        return upper.y + (lower.y - upper.y) * progress;
      }
    }
    return points.last.y;
  }

  static Offset _point(Size size, double x, double y) {
    return Offset(
      x / _figureWidth * size.width,
      y / _figureHeight * size.height,
    );
  }

  static Path _pathFor(List<Offset> points) {
    final Path path = Path()..moveTo(points.first.dx, points.first.dy);
    for (final Offset point in points.skip(1)) {
      path.lineTo(point.dx, point.dy);
    }
    return path;
  }

  static void _drawDashedPath(Canvas canvas, Path path, Paint paint) {
    for (final metric in path.computeMetrics()) {
      double distance = 0;
      while (distance < metric.length) {
        final double next =
            math.min(distance + paint.strokeWidth * 4, metric.length);
        canvas.drawPath(metric.extractPath(distance, next), paint);
        distance += paint.strokeWidth * 6.4;
      }
    }
  }

  @override
  bool shouldRepaint(
      covariant _AutismDevDevelopmentProfilePainter oldDelegate) {
    return false;
  }
}

class _AutismDevProfileScalePoint {
  const _AutismDevProfileScalePoint(this.score, this.y);

  final double score;
  final double y;
}

const Map<String, List<_AutismDevProfileScalePoint>> _profileScalePoints =
    <String, List<_AutismDevProfileScalePoint>>{
  '感知觉': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(55, 586),
    _AutismDevProfileScalePoint(52, 760),
    _AutismDevProfileScalePoint(47, 862),
    _AutismDevProfileScalePoint(44, 940),
    _AutismDevProfileScalePoint(40, 1026),
    _AutismDevProfileScalePoint(37, 1093),
    _AutismDevProfileScalePoint(36, 1120),
    _AutismDevProfileScalePoint(29, 1210),
    _AutismDevProfileScalePoint(27, 1262),
    _AutismDevProfileScalePoint(21, 1344),
    _AutismDevProfileScalePoint(19, 1404),
    _AutismDevProfileScalePoint(16, 1430),
    _AutismDevProfileScalePoint(10, 1511),
    _AutismDevProfileScalePoint(5, 1524),
    _AutismDevProfileScalePoint(2, 1539),
    _AutismDevProfileScalePoint(1, 1554),
    _AutismDevProfileScalePoint(0, 1568),
  ],
  '粗大动作': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(72, 529),
    _AutismDevProfileScalePoint(65, 727),
    _AutismDevProfileScalePoint(64, 760),
    _AutismDevProfileScalePoint(47, 862),
    _AutismDevProfileScalePoint(46, 938),
    _AutismDevProfileScalePoint(35, 1004),
    _AutismDevProfileScalePoint(34, 1117),
    _AutismDevProfileScalePoint(24, 1164),
    _AutismDevProfileScalePoint(22, 1238),
    _AutismDevProfileScalePoint(21, 1270),
    _AutismDevProfileScalePoint(19, 1313),
    _AutismDevProfileScalePoint(7, 1376),
    _AutismDevProfileScalePoint(6, 1418),
    _AutismDevProfileScalePoint(5, 1451),
    _AutismDevProfileScalePoint(1, 1482),
    _AutismDevProfileScalePoint(0, 1496),
  ],
  '精细动作': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(66, 529),
    _AutismDevProfileScalePoint(63, 740),
    _AutismDevProfileScalePoint(62, 755),
    _AutismDevProfileScalePoint(51, 787),
    _AutismDevProfileScalePoint(50, 817),
    _AutismDevProfileScalePoint(49, 865),
    _AutismDevProfileScalePoint(48, 910),
    _AutismDevProfileScalePoint(47, 938),
    _AutismDevProfileScalePoint(39, 965),
    _AutismDevProfileScalePoint(35, 1027),
    _AutismDevProfileScalePoint(34, 1055),
    _AutismDevProfileScalePoint(33, 1130),
    _AutismDevProfileScalePoint(24, 1178),
    _AutismDevProfileScalePoint(23, 1211),
    _AutismDevProfileScalePoint(22, 1240),
    _AutismDevProfileScalePoint(21, 1262),
    _AutismDevProfileScalePoint(20, 1280),
    _AutismDevProfileScalePoint(11, 1375),
    _AutismDevProfileScalePoint(9, 1419),
    _AutismDevProfileScalePoint(4, 1435),
    _AutismDevProfileScalePoint(3, 1470),
    _AutismDevProfileScalePoint(2, 1500),
    _AutismDevProfileScalePoint(1, 1530),
    _AutismDevProfileScalePoint(0, 1542),
  ],
  '语言与沟通': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(79, 562),
    _AutismDevProfileScalePoint(76, 760),
    _AutismDevProfileScalePoint(67, 955),
    _AutismDevProfileScalePoint(53, 1090),
    _AutismDevProfileScalePoint(52, 1135),
    _AutismDevProfileScalePoint(36, 1268),
    _AutismDevProfileScalePoint(27, 1360),
    _AutismDevProfileScalePoint(21, 1415),
    _AutismDevProfileScalePoint(18, 1480),
    _AutismDevProfileScalePoint(8, 1492),
    _AutismDevProfileScalePoint(2, 1512),
    _AutismDevProfileScalePoint(1, 1530),
    _AutismDevProfileScalePoint(0, 1544),
  ],
  '认知': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(55, 530),
    _AutismDevProfileScalePoint(50, 586),
    _AutismDevProfileScalePoint(42, 758),
    _AutismDevProfileScalePoint(30, 940),
    _AutismDevProfileScalePoint(20, 1118),
    _AutismDevProfileScalePoint(10, 1270),
    _AutismDevProfileScalePoint(9, 1315),
    _AutismDevProfileScalePoint(5, 1362),
    _AutismDevProfileScalePoint(4, 1388),
    _AutismDevProfileScalePoint(2, 1426),
    _AutismDevProfileScalePoint(1, 1455),
    _AutismDevProfileScalePoint(0, 1498),
  ],
  '社会交往': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(47, 558),
    _AutismDevProfileScalePoint(45, 758),
    _AutismDevProfileScalePoint(40, 910),
    _AutismDevProfileScalePoint(30, 1090),
    _AutismDevProfileScalePoint(24, 1225),
    _AutismDevProfileScalePoint(19, 1265),
    _AutismDevProfileScalePoint(15, 1315),
    _AutismDevProfileScalePoint(14, 1355),
    _AutismDevProfileScalePoint(11, 1450),
    _AutismDevProfileScalePoint(4, 1498),
    _AutismDevProfileScalePoint(1, 1530),
    _AutismDevProfileScalePoint(0, 1546),
  ],
  '生活自理': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(67, 530),
    _AutismDevProfileScalePoint(62, 758),
    _AutismDevProfileScalePoint(46, 940),
    _AutismDevProfileScalePoint(34, 1090),
    _AutismDevProfileScalePoint(33, 1120),
    _AutismDevProfileScalePoint(18, 1178),
    _AutismDevProfileScalePoint(15, 1265),
    _AutismDevProfileScalePoint(12, 1315),
    _AutismDevProfileScalePoint(8, 1350),
    _AutismDevProfileScalePoint(6, 1380),
    _AutismDevProfileScalePoint(5, 1408),
    _AutismDevProfileScalePoint(3, 1450),
    _AutismDevProfileScalePoint(2, 1495),
    _AutismDevProfileScalePoint(1, 1512),
    _AutismDevProfileScalePoint(0, 1545),
  ],
  '发展分数': <_AutismDevProfileScalePoint>[
    _AutismDevProfileScalePoint(441, 532),
    _AutismDevProfileScalePoint(421, 562),
    _AutismDevProfileScalePoint(416, 592),
    _AutismDevProfileScalePoint(405, 760),
    _AutismDevProfileScalePoint(330, 790),
    _AutismDevProfileScalePoint(329, 820),
    _AutismDevProfileScalePoint(328, 862),
    _AutismDevProfileScalePoint(323, 910),
    _AutismDevProfileScalePoint(312, 940),
    _AutismDevProfileScalePoint(267, 956),
    _AutismDevProfileScalePoint(253, 970),
    _AutismDevProfileScalePoint(249, 1000),
    _AutismDevProfileScalePoint(248, 1030),
    _AutismDevProfileScalePoint(244, 1062),
    _AutismDevProfileScalePoint(243, 1092),
    _AutismDevProfileScalePoint(234, 1125),
    _AutismDevProfileScalePoint(192, 1138),
    _AutismDevProfileScalePoint(167, 1168),
    _AutismDevProfileScalePoint(163, 1182),
    _AutismDevProfileScalePoint(160, 1210),
    _AutismDevProfileScalePoint(157, 1226),
    _AutismDevProfileScalePoint(152, 1240),
    _AutismDevProfileScalePoint(149, 1270),
    _AutismDevProfileScalePoint(123, 1318),
    _AutismDevProfileScalePoint(93, 1355),
    _AutismDevProfileScalePoint(89, 1372),
    _AutismDevProfileScalePoint(79, 1388),
    _AutismDevProfileScalePoint(75, 1398),
    _AutismDevProfileScalePoint(73, 1418),
    _AutismDevProfileScalePoint(68, 1432),
    _AutismDevProfileScalePoint(58, 1446),
    _AutismDevProfileScalePoint(51, 1462),
    _AutismDevProfileScalePoint(28, 1480),
    _AutismDevProfileScalePoint(27, 1492),
    _AutismDevProfileScalePoint(26, 1500),
    _AutismDevProfileScalePoint(16, 1520),
    _AutismDevProfileScalePoint(9, 1534),
    _AutismDevProfileScalePoint(2, 1548),
    _AutismDevProfileScalePoint(1, 1560),
    _AutismDevProfileScalePoint(0, 1572),
  ],
};

class _AutismDevBehaviorProfilePainter extends CustomPainter {
  const _AutismDevBehaviorProfilePainter();

  static const double _figureWidth = 1488;
  static const double _figureHeight = 2103;
  static const Offset _sourceCenter = Offset(737, 939);
  static const double _innerRadius = 240;
  static const double _outerRadius = 456;

  @override
  void paint(Canvas canvas, Size size) {
    final Offset center = _point(size, _sourceCenter.dx, _sourceCenter.dy);
    final double radiusScale = size.width / _figureWidth;
    final List<Offset> points = <Offset>[];
    for (int index = 0; index < _autismDevBehaviorScores.length; index++) {
      final _AutismDevBehaviorScore score = _autismDevBehaviorScores[index];
      final double severity =
          score.total <= 0 ? 0 : (score.m + score.s * 2) / (score.total * 2);
      final double radius =
          (_innerRadius + (_outerRadius - _innerRadius) * severity) *
              radiusScale;
      final double angle =
          -math.pi / 2 + math.pi * 2 * index / _autismDevBehaviorScores.length;
      points.add(
        Offset(
          center.dx + math.cos(angle) * radius,
          center.dy + math.sin(angle) * radius,
        ),
      );
    }
    if (points.isEmpty) {
      return;
    }
    final Path polygon = Path()..moveTo(points.first.dx, points.first.dy);
    for (final Offset point in points.skip(1)) {
      polygon.lineTo(point.dx, point.dy);
    }
    polygon.close();

    final Paint fillPaint = Paint()
      ..color = _ReportTheme.orange.withOpacity(.18)
      ..style = PaintingStyle.fill;
    final Paint strokePaint = Paint()
      ..color = _ReportTheme.orangeDeep.withOpacity(.88)
      ..strokeWidth = size.width / 380
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;
    final Paint spokePaint = Paint()
      ..color = _ReportTheme.orange.withOpacity(.32)
      ..strokeWidth = size.width / 620
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round;
    final Paint dotPaint = Paint()
      ..color = _ReportTheme.orangeDeep
      ..style = PaintingStyle.fill;

    canvas.drawPath(polygon, fillPaint);
    for (final Offset point in points) {
      canvas.drawLine(center, point, spokePaint);
    }
    canvas.drawPath(polygon, strokePaint);
    for (final Offset point in points) {
      canvas.drawCircle(point, size.width / 180, dotPaint);
    }
    canvas.drawCircle(center, size.width / 220, dotPaint);
  }

  static Offset _point(Size size, double x, double y) {
    return Offset(
      x / _figureWidth * size.width,
      y / _figureHeight * size.height,
    );
  }

  @override
  bool shouldRepaint(covariant _AutismDevBehaviorProfilePainter oldDelegate) {
    return false;
  }
}

class _AutismDevDevelopmentBarChart extends StatelessWidget {
  const _AutismDevDevelopmentBarChart();

  @override
  Widget build(BuildContext context) {
    return _AutismDevChartFrame(
      legend: const <_AutismDevLegendItem>[
        _AutismDevLegendItem('P 通过', _ReportTheme.blue),
        _AutismDevLegendItem('E 中间反应', _ReportTheme.orange),
        _AutismDevLegendItem('F/X 未通过或不适用', _ReportTheme.line),
      ],
      child: Column(
        children: <Widget>[
          for (final _AutismDevDevelopmentScore item
              in _autismDevDevelopmentScores)
            _AutismDevDevelopmentBarRow(item: item),
        ],
      ),
    );
  }
}

class _AutismDevBehaviorBarChart extends StatelessWidget {
  const _AutismDevBehaviorBarChart();

  @override
  Widget build(BuildContext context) {
    return _AutismDevChartFrame(
      legend: const <_AutismDevLegendItem>[
        _AutismDevLegendItem('A 无异常', _ReportTheme.green),
        _AutismDevLegendItem('M 轻度异常', _ReportTheme.amber),
        _AutismDevLegendItem('S 重度异常', _ReportTheme.rose),
      ],
      child: Column(
        children: <Widget>[
          for (final _AutismDevBehaviorScore item in _autismDevBehaviorScores)
            _AutismDevBehaviorBarRow(item: item),
        ],
      ),
    );
  }
}

class _AutismDevChartFrame extends StatelessWidget {
  const _AutismDevChartFrame({
    required this.legend,
    required this.child,
  });

  final List<_AutismDevLegendItem> legend;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(14, 14, 14, 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Wrap(
            spacing: 12,
            runSpacing: 6,
            children: <Widget>[
              for (final _AutismDevLegendItem item in legend)
                _AutismDevLegend(item: item),
            ],
          ),
          const SizedBox(height: 14),
          child,
        ],
      ),
    );
  }
}

class _AutismDevDevelopmentBarRow extends StatelessWidget {
  const _AutismDevDevelopmentBarRow({required this.item});

  final _AutismDevDevelopmentScore item;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: <Widget>[
          SizedBox(
            width: 94,
            child: Text(
              item.label,
              style: const TextStyle(
                color: _ReportTheme.text,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          Expanded(
            child: ClipRRect(
              borderRadius: BorderRadius.circular(999),
              child: SizedBox(
                height: 18,
                child: Row(
                  children: <Widget>[
                    _AutismDevBarSegment(
                      value: item.p,
                      total: item.total,
                      color: _ReportTheme.blue,
                    ),
                    _AutismDevBarSegment(
                      value: item.e,
                      total: item.total,
                      color: _ReportTheme.orange,
                    ),
                    _AutismDevBarSegment(
                      value: item.f + item.x,
                      total: item.total,
                      color: _ReportTheme.line,
                    ),
                  ],
                ),
              ),
            ),
          ),
          SizedBox(
            width: 58,
            child: Text(
              '${item.passRate.toStringAsFixed(0)}%',
              textAlign: TextAlign.right,
              style: const TextStyle(
                color: _ReportTheme.ink,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevBehaviorBarRow extends StatelessWidget {
  const _AutismDevBehaviorBarRow({required this.item});

  final _AutismDevBehaviorScore item;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: <Widget>[
          SizedBox(
            width: 116,
            child: Text(
              item.label,
              style: const TextStyle(
                color: _ReportTheme.text,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          Expanded(
            child: ClipRRect(
              borderRadius: BorderRadius.circular(999),
              child: SizedBox(
                height: 18,
                child: Row(
                  children: <Widget>[
                    _AutismDevBarSegment(
                      value: item.a,
                      total: item.total,
                      color: _ReportTheme.green,
                    ),
                    _AutismDevBarSegment(
                      value: item.m,
                      total: item.total,
                      color: _ReportTheme.amber,
                    ),
                    _AutismDevBarSegment(
                      value: item.s,
                      total: item.total,
                      color: _ReportTheme.rose,
                    ),
                  ],
                ),
              ),
            ),
          ),
          SizedBox(
            width: 64,
            child: Text(
              '${item.m + item.s}项',
              textAlign: TextAlign.right,
              style: const TextStyle(
                color: _ReportTheme.ink,
                fontSize: 12,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevBarSegment extends StatelessWidget {
  const _AutismDevBarSegment({
    required this.value,
    required this.total,
    required this.color,
  });

  final int value;
  final int total;
  final Color color;

  @override
  Widget build(BuildContext context) {
    if (value <= 0 || total <= 0) {
      return const SizedBox.shrink();
    }
    return Expanded(
      flex: value,
      child: ColoredBox(color: color),
    );
  }
}

class _AutismDevLegend extends StatelessWidget {
  const _AutismDevLegend({required this.item});

  final _AutismDevLegendItem item;

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        Container(
          width: 9,
          height: 9,
          decoration: BoxDecoration(
            color: item.color,
            shape: BoxShape.circle,
          ),
        ),
        const SizedBox(width: 6),
        Text(
          item.label,
          style: const TextStyle(
            color: _ReportTheme.text,
            fontSize: 12,
            fontWeight: FontWeight.w800,
          ),
        ),
      ],
    );
  }
}

class _AutismDevTableFrame extends StatelessWidget {
  const _AutismDevTableFrame({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(10),
      child: child,
    );
  }
}

TableRow _autismDevTableHeaderRow(List<String> values, {bool tall = false}) {
  return TableRow(
    decoration: const BoxDecoration(color: Color(0xFFFFF8F2)),
    children: <Widget>[
      for (final String value in values)
        _autismDevTableCell(
          value,
          emph: true,
          header: true,
          minHeight: tall ? 44 : 38,
        ),
    ],
  );
}

TableRow _autismDevTableRow(List<String> values, {bool emphFirst = false}) {
  return TableRow(
    children: <Widget>[
      for (int index = 0; index < values.length; index++)
        _autismDevTableCell(
          values[index],
          emph: emphFirst && index == 0,
        ),
    ],
  );
}

Widget _autismDevTableCell(
  String value, {
  bool emph = false,
  bool header = false,
  double minHeight = 36,
}) {
  return Container(
    constraints: BoxConstraints(minHeight: minHeight),
    alignment: Alignment.center,
    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
    color: header ? const Color(0xFFFFF8F2) : Colors.white,
    child: Text(
      value,
      textAlign: TextAlign.center,
      style: TextStyle(
        color: emph ? _ReportTheme.ink : _ReportTheme.text,
        fontSize: header ? 13 : 12,
        height: 1.35,
        fontWeight: emph ? FontWeight.w900 : FontWeight.w800,
      ),
    ),
  );
}

class _AutismDevInfoItem {
  const _AutismDevInfoItem(this.label, this.value);

  final String label;
  final String value;
}

class _AutismDevReportTabSpec {
  const _AutismDevReportTabSpec(this.label, this.tab);

  final String label;
  final _AutismDevReportTab tab;
}

class _AutismDevSummaryItem {
  const _AutismDevSummaryItem(this.title, this.description, this.icon);

  final String title;
  final String description;
  final IconData icon;
}

class _AutismDevLegendItem {
  const _AutismDevLegendItem(this.label, this.color);

  final String label;
  final Color color;
}

class _AutismDevDevelopmentScore {
  const _AutismDevDevelopmentScore({
    required this.label,
    required this.p,
    required this.e,
    required this.f,
    required this.x,
    required this.total,
  });

  final String label;
  final int p;
  final int e;
  final int f;
  final int x;
  final int total;

  int get supportCount => e + f + x;
  double get passRate => total <= 0 ? 0 : p * 100 / total;
}

class _AutismDevBehaviorScore {
  const _AutismDevBehaviorScore({
    required this.label,
    required this.a,
    required this.m,
    required this.s,
  });

  final String label;
  final int a;
  final int m;
  final int s;

  int get total => a + m + s;
}

class _AutismDevAnalysisItem {
  const _AutismDevAnalysisItem({
    required this.domain,
    required this.status,
    required this.strength,
    required this.target,
  });

  final String domain;
  final String status;
  final String strength;
  final String target;
}

const List<_AutismDevDevelopmentScore> _autismDevDevelopmentScores =
    <_AutismDevDevelopmentScore>[
  _AutismDevDevelopmentScore(
    label: '语言与沟通',
    p: 29,
    e: 18,
    f: 25,
    x: 7,
    total: 79,
  ),
  _AutismDevDevelopmentScore(
    label: '认知',
    p: 34,
    e: 10,
    f: 9,
    x: 2,
    total: 55,
  ),
  _AutismDevDevelopmentScore(
    label: '生活自理',
    p: 41,
    e: 12,
    f: 10,
    x: 4,
    total: 67,
  ),
  _AutismDevDevelopmentScore(
    label: '感知觉',
    p: 38,
    e: 8,
    f: 7,
    x: 2,
    total: 55,
  ),
  _AutismDevDevelopmentScore(
    label: '粗大动作',
    p: 44,
    e: 13,
    f: 11,
    x: 4,
    total: 72,
  ),
  _AutismDevDevelopmentScore(
    label: '精细动作',
    p: 35,
    e: 14,
    f: 13,
    x: 4,
    total: 66,
  ),
  _AutismDevDevelopmentScore(
    label: '社会交往',
    p: 25,
    e: 9,
    f: 11,
    x: 2,
    total: 47,
  ),
];

const List<_AutismDevBehaviorScore> _autismDevBehaviorScores =
    <_AutismDevBehaviorScore>[
  _AutismDevBehaviorScore(label: '依附情绪行为', a: 4, m: 2, s: 1),
  _AutismDevBehaviorScore(label: '情绪理解', a: 3, m: 2, s: 0),
  _AutismDevBehaviorScore(label: '情绪表达与调节', a: 3, m: 3, s: 1),
  _AutismDevBehaviorScore(label: '关系与情感', a: 4, m: 2, s: 1),
  _AutismDevBehaviorScore(label: '对物品的兴趣', a: 5, m: 3, s: 1),
  _AutismDevBehaviorScore(label: '感觉偏好', a: 5, m: 4, s: 1),
  _AutismDevBehaviorScore(label: '特殊行为', a: 4, m: 3, s: 0),
];

const List<_AutismDevAnalysisItem> _autismDevAnalysisItems =
    <_AutismDevAnalysisItem>[
  _AutismDevAnalysisItem(
    domain: '感知觉',
    status: '视觉、听觉反应较稳定，触觉与味觉辨别任务仍需持续观察。',
    strength: '优势：熟悉刺激反应明确。劣势：复杂辨别与记忆任务稳定性不足。',
    target: '优先安排触觉辨别、味觉辨别与多感官配对训练。',
  ),
  _AutismDevAnalysisItem(
    domain: '粗大动作',
    status: '基本姿势与移动能力较好，球类操作与平衡动作存在波动。',
    strength: '优势：移动类项目完成度较高。劣势：抛接、踢、拍等协调动作较弱。',
    target: '设置抛接球、平衡木行走与双手协调游戏目标。',
  ),
  _AutismDevAnalysisItem(
    domain: '精细动作',
    status: '基础抓握和摆弄物品较稳定，握笔写画及工具使用需加强。',
    strength: '优势：基本操作可配合。劣势：双手配合和精细控制持续时间短。',
    target: '围绕穿珠、剪纸、仿画和工具使用建立分步训练。',
  ),
  _AutismDevAnalysisItem(
    domain: '语言与沟通',
    status: '理解简单名称和动作指令较好，主动表达和复述能力不足。',
    strength: '优势：名称指认有基础。劣势：短语、句子和主动提问较少。',
    target: '以表达要求、回答问题、短语扩展和主动提问作为核心目标。',
  ),
  _AutismDevAnalysisItem(
    domain: '认知',
    status: '配对、分类和部分颜色概念可完成，数量及关系概念仍需支持。',
    strength: '优势：具体物品配对较好。劣势：抽象概念和数概念掌握不稳定。',
    target: '训练大小、多少、长短、颜色分类和1-5数量操作。',
  ),
  _AutismDevAnalysisItem(
    domain: '社会交往',
    status: '熟悉情境中可回应互动，陌生情境与社交礼仪需提示。',
    strength: '优势：近距离互动可建立。劣势：主动打招呼、告别与感谢不足。',
    target: '设计打招呼、告别、请求帮助和表示感谢的情境练习。',
  ),
  _AutismDevAnalysisItem(
    domain: '生活自理',
    status: '进食和部分家居自理较稳定，穿衣梳洗流程仍需辅助。',
    strength: '优势：日常熟悉流程接受度较好。劣势：多步骤任务独立性不足。',
    target: '拆分穿衣、梳洗、物品归位和收拾餐具等连续任务。',
  ),
  _AutismDevAnalysisItem(
    domain: '情绪与行为',
    status: '轻度异常项目集中在情绪调节、感觉偏好和特殊行为。',
    strength: '优势：多数项目无重度异常。劣势：转变适应与感官偏好需关注。',
    target: '建立等待、转换活动、情绪表达和感官调节策略。',
  ),
];
