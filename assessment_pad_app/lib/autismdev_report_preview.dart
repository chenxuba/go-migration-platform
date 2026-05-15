part of 'assessment_report_list_page.dart';

enum _AutismDevReportTab { overview, development, behavior, analysis }

class _AutismDevReportPreviewDialog extends StatefulWidget {
  const _AutismDevReportPreviewDialog({required this.record});

  final Pep3RecordSummary record;

  @override
  State<_AutismDevReportPreviewDialog> createState() =>
      _AutismDevReportPreviewDialogState();
}

class _AutismDevReportPreviewDialogState
    extends State<_AutismDevReportPreviewDialog> {
  _AutismDevReportTab _activeTab = _AutismDevReportTab.overview;

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
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFCF8),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: _ReportTheme.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          _AutismDevReportTabChip(
            label: '报告总览',
            tab: _AutismDevReportTab.overview,
            activeTab: _activeTab,
            onSelected: _selectTab,
          ),
          const SizedBox(width: 8),
          _AutismDevReportTabChip(
            label: '发展能力',
            tab: _AutismDevReportTab.development,
            activeTab: _activeTab,
            onSelected: _selectTab,
          ),
          const SizedBox(width: 8),
          _AutismDevReportTabChip(
            label: '情绪行为',
            tab: _AutismDevReportTab.behavior,
            activeTab: _activeTab,
            onSelected: _selectTab,
          ),
          const SizedBox(width: 8),
          _AutismDevReportTabChip(
            label: '结果分析',
            tab: _AutismDevReportTab.analysis,
            activeTab: _activeTab,
            onSelected: _selectTab,
          ),
          const Spacer(),
        ],
      ),
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
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        _AutismDevReportPageHeader(record: widget.record),
        const SizedBox(height: 18),
        if (_activeTab == _AutismDevReportTab.overview)
          const _AutismDevOverviewSection()
        else if (_activeTab == _AutismDevReportTab.development)
          const _AutismDevDevelopmentSection()
        else if (_activeTab == _AutismDevReportTab.behavior)
          const _AutismDevBehaviorSection()
        else
          const _AutismDevAnalysisSection(),
      ],
    );
  }
}

class _AutismDevReportTabChip extends StatelessWidget {
  const _AutismDevReportTabChip({
    required this.label,
    required this.tab,
    required this.activeTab,
    required this.onSelected,
  });

  final String label;
  final _AutismDevReportTab tab;
  final _AutismDevReportTab activeTab;
  final ValueChanged<_AutismDevReportTab> onSelected;

  @override
  Widget build(BuildContext context) {
    return _ErxinReportTabChip(
      label: label,
      active: activeTab == tab,
      onTap: () => onSelected(tab),
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
