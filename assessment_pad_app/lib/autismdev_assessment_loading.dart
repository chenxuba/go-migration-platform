part of 'autismdev_assessment_page.dart';

class _AutismDevLoadingBody extends StatelessWidget {
  const _AutismDevLoadingBody();

  @override
  Widget build(BuildContext context) {
    return const Column(
      children: <Widget>[
        Expanded(
          child: Padding(
            padding: EdgeInsets.fromLTRB(10, 10, 10, 10),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: <Widget>[
                SizedBox(width: 226, child: _AutismDevSidebarSkeleton()),
                SizedBox(width: 10),
                Expanded(child: _AutismDevQuestionSkeleton()),
                SizedBox(width: 10),
                SizedBox(width: 238, child: _AutismDevRightRailSkeleton()),
              ],
            ),
          ),
        ),
        _AutismDevLoadingFooter(),
      ],
    );
  }
}

class _AutismDevSidebarSkeleton extends StatelessWidget {
  const _AutismDevSidebarSkeleton();

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: _panelDecoration(),
      child: Column(
        children: <Widget>[
          Container(
            height: 48,
            padding: const EdgeInsets.symmetric(horizontal: 14),
            decoration: const BoxDecoration(
              border: Border(
                bottom: BorderSide(color: _AutismDevColors.lineSoft),
              ),
            ),
            child: Row(
              children: <Widget>[
                const Text(
                  '领域任务',
                  style: TextStyle(
                    color: _AutismDevColors.ink,
                    fontSize: 16,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const Spacer(),
                _AutismDevSkeletonBlock(width: 30, height: 30, radius: 8),
              ],
            ),
          ),
          Expanded(
            child: SingleChildScrollView(
              physics: const NeverScrollableScrollPhysics(),
              child: Column(
                children: const <Widget>[
                  _AutismDevLoadingDomainRow(selected: true, expanded: true),
                  _AutismDevLoadingDomainRow(selected: false),
                  _AutismDevLoadingDomainRow(selected: false),
                  _AutismDevLoadingDomainRow(selected: false),
                  _AutismDevLoadingDomainRow(selected: false),
                  _AutismDevLoadingDomainRow(selected: false),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _AutismDevLoadingDomainRow extends StatelessWidget {
  const _AutismDevLoadingDomainRow({
    required this.selected,
    this.expanded = false,
  });

  final bool selected;
  final bool expanded;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 10, 9),
      decoration: BoxDecoration(
        color: selected ? const Color(0xFFFFFBF8) : Colors.transparent,
        border: const Border(
          bottom: BorderSide(color: _AutismDevColors.lineSoft),
        ),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: <Widget>[
              _AutismDevSkeletonBlock(width: 20, height: 20, radius: 7),
              const SizedBox(width: 4),
              _AutismDevSkeletonBlock(width: 9, height: 9, radius: 99),
              const SizedBox(width: 8),
              const Expanded(
                child: _AutismDevSkeletonBlock(widthFactor: .72, height: 13),
              ),
              const SizedBox(width: 9),
              _AutismDevSkeletonBlock(width: 34, height: 12, radius: 6),
            ],
          ),
          const SizedBox(height: 9),
          Row(
            children: <Widget>[
              const SizedBox(width: 28),
              const Expanded(
                child: _AutismDevSkeletonBlock(height: 4, radius: 99),
              ),
              const SizedBox(width: 9),
              _AutismDevSkeletonBlock(width: 28, height: 12, radius: 6),
            ],
          ),
          if (expanded) ...const <Widget>[
            SizedBox(height: 9),
            _AutismDevLoadingNavRow(),
            _AutismDevLoadingNavRow(),
            _AutismDevLoadingNavRow(),
          ],
        ],
      ),
    );
  }
}

class _AutismDevLoadingNavRow extends StatelessWidget {
  const _AutismDevLoadingNavRow();

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 18, top: 3),
      child: SizedBox(
        height: 31,
        child: Row(
          children: <Widget>[
            _AutismDevSkeletonBlock(width: 42, height: 12),
            const SizedBox(width: 8),
            const Expanded(
              child: _AutismDevSkeletonBlock(widthFactor: .72, height: 12),
            ),
            const SizedBox(width: 7),
            _AutismDevSkeletonBlock(width: 15, height: 15, radius: 99),
          ],
        ),
      ),
    );
  }
}

class _AutismDevQuestionSkeleton extends StatelessWidget {
  const _AutismDevQuestionSkeleton();

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: _panelDecoration(),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 18, 16, 14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            SizedBox(
              height: 38,
              child: Row(
                children: <Widget>[
                  const Expanded(
                    child: _AutismDevSkeletonBlock(
                      widthFactor: .68,
                      height: 26,
                    ),
                  ),
                  const SizedBox(width: 8),
                  Container(
                    height: 30,
                    padding: const EdgeInsets.symmetric(horizontal: 11),
                    alignment: Alignment.center,
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFF1E8),
                      borderRadius: BorderRadius.circular(9),
                      border: Border.all(color: const Color(0xFFFFC8AD)),
                    ),
                    child: const Text(
                      '题目偏好配置',
                      style: TextStyle(
                        color: _AutismDevColors.orangeDeep,
                        fontSize: 13,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
            const Expanded(
              child: SingleChildScrollView(
                physics: NeverScrollableScrollPhysics(),
                child: Column(
                  children: <Widget>[
                    _AutismDevLoadingMetaRow(),
                    SizedBox(height: 10),
                    _AutismDevLoadingDetailCard(title: '评估材料', height: 68),
                    _AutismDevLoadingDetailCard(title: '评估方法', height: 78),
                    _AutismDevLoadingDetailCard(title: '评分标准', height: 78),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 10),
            const _AutismDevScoreSkeleton(),
          ],
        ),
      ),
    );
  }
}

class _AutismDevLoadingMetaRow extends StatelessWidget {
  const _AutismDevLoadingMetaRow();

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: const <Widget>[
        Expanded(child: _AutismDevLoadingMetaCard(label: '评估范围')),
        SizedBox(width: 10),
        SizedBox(
          width: 158,
          child: _AutismDevLoadingMetaCard(label: '参考年龄'),
        ),
      ],
    );
  }
}

class _AutismDevLoadingMetaCard extends StatelessWidget {
  const _AutismDevLoadingMetaCard({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: const BoxConstraints(minHeight: 70),
      padding: const EdgeInsets.fromLTRB(13, 12, 13, 12),
      decoration: _autismDevDetailCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            label,
            style: const TextStyle(
              color: _AutismDevColors.ink,
              fontSize: 13,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 10),
          const _AutismDevSkeletonBlock(widthFactor: .8, height: 13),
          const SizedBox(height: 7),
          const _AutismDevSkeletonBlock(widthFactor: .52, height: 13),
        ],
      ),
    );
  }
}

class _AutismDevLoadingDetailCard extends StatelessWidget {
  const _AutismDevLoadingDetailCard({
    required this.title,
    required this.height,
  });

  final String title;
  final double height;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      constraints: BoxConstraints(minHeight: height),
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 14),
      decoration: _autismDevDetailCardDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            title,
            style: const TextStyle(
              color: _AutismDevColors.ink,
              fontSize: 16,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 13),
          const _AutismDevSkeletonBlock(widthFactor: .92, height: 12),
          const SizedBox(height: 8),
          const _AutismDevSkeletonBlock(widthFactor: .68, height: 12),
        ],
      ),
    );
  }
}

class _AutismDevScoreSkeleton extends StatelessWidget {
  const _AutismDevScoreSkeleton();

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: <Widget>[
        const Row(
          children: <Widget>[
            Text(
              '评分',
              style: TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 15,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            Spacer(),
          ],
        ),
        const SizedBox(height: 8),
        LayoutBuilder(
          builder: (BuildContext context, BoxConstraints constraints) {
            const double spacing = 12;
            final bool singleColumn = constraints.maxWidth < 430;
            final double cardWidth = singleColumn
                ? constraints.maxWidth
                : (constraints.maxWidth - spacing) / 2;
            return Wrap(
              spacing: spacing,
              runSpacing: 8,
              children: <Widget>[
                SizedBox(
                  width: cardWidth,
                  child: const _AutismDevScoreCardSkeleton(),
                ),
                SizedBox(
                  width: cardWidth,
                  child: const _AutismDevScoreCardSkeleton(),
                ),
              ],
            );
          },
        ),
      ],
    );
  }
}

class _AutismDevScoreCardSkeleton extends StatelessWidget {
  const _AutismDevScoreCardSkeleton();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 78,
      padding: const EdgeInsets.fromLTRB(13, 9, 11, 9),
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: _AutismDevColors.line),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Row(
        children: const <Widget>[
          _AutismDevSkeletonBlock(width: 36, height: 36, radius: 18),
          SizedBox(width: 10),
          Expanded(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                _AutismDevSkeletonBlock(widthFactor: .72, height: 14),
                SizedBox(height: 8),
                _AutismDevSkeletonBlock(widthFactor: .9, height: 11),
              ],
            ),
          ),
          SizedBox(width: 8),
          _AutismDevSkeletonBlock(width: 20, height: 20, radius: 99),
        ],
      ),
    );
  }
}

class _AutismDevRightRailSkeleton extends StatelessWidget {
  const _AutismDevRightRailSkeleton();

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: _panelDecoration(),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 6),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            const Text(
              '当前进度',
              style: TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 16,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 14),
            Row(
              children: const <Widget>[
                _AutismDevSkeletonBlock(width: 82, height: 82, radius: 99),
                SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      _AutismDevSkeletonBlock(width: 62, height: 12),
                      SizedBox(height: 8),
                      _AutismDevSkeletonBlock(widthFactor: .82, height: 16),
                      SizedBox(height: 12),
                      _AutismDevSkeletonBlock(width: 40, height: 12),
                      SizedBox(height: 8),
                      _AutismDevSkeletonBlock(widthFactor: .58, height: 16),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            const Expanded(child: _AutismDevRangeSkeleton()),
            const SizedBox(height: 16),
            const Text(
              '备注',
              style: TextStyle(
                color: _AutismDevColors.ink,
                fontSize: 16,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 10),
            const _AutismDevSkeletonBlock(height: 66, radius: 8),
          ],
        ),
      ),
    );
  }
}

class _AutismDevRangeSkeleton extends StatelessWidget {
  const _AutismDevRangeSkeleton();

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: const <Widget>[
        Row(
          children: <Widget>[
            Expanded(
              child: Text(
                '分类',
                style: TextStyle(
                  color: _AutismDevColors.ink,
                  fontSize: 16,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ),
            _AutismDevSkeletonBlock(width: 28, height: 12),
          ],
        ),
        SizedBox(height: 10),
        _AutismDevRangeListSkeleton(active: true),
        SizedBox(height: 7),
        _AutismDevRangeListSkeleton(active: false),
        SizedBox(height: 7),
        _AutismDevRangeListSkeleton(active: false),
        SizedBox(height: 7),
        _AutismDevRangeListSkeleton(active: false),
      ],
    );
  }
}

class _AutismDevRangeListSkeleton extends StatelessWidget {
  const _AutismDevRangeListSkeleton({required this.active});

  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 39,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: active ? const Color(0xFFFFF1E8) : _AutismDevColors.softPanel,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: active ? const Color(0xFFFFC8AD) : _AutismDevColors.line,
        ),
      ),
      child: Row(
        children: const <Widget>[
          _AutismDevSkeletonBlock(width: 7, height: 7, radius: 99),
          SizedBox(width: 8),
          Expanded(
              child: _AutismDevSkeletonBlock(widthFactor: .72, height: 12)),
          SizedBox(width: 8),
          _AutismDevSkeletonBlock(width: 38, height: 22, radius: 11),
        ],
      ),
    );
  }
}

class _AutismDevLoadingFooter extends StatelessWidget {
  const _AutismDevLoadingFooter();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 62,
      padding: const EdgeInsets.symmetric(horizontal: 20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.97),
        border: Border.all(color: _AutismDevColors.line),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
        boxShadow: _autismDevShadow(
          color: const Color(0x14B05F32),
          blur: 16,
        ),
      ),
      child: Row(
        children: <Widget>[
          _FooterButton(
            label: '上一题',
            icon: Icons.chevron_left_rounded,
            enabled: false,
            onTap: () {},
          ),
          const Spacer(),
          Text.rich(
            const TextSpan(
              children: <InlineSpan>[
                TextSpan(
                  text: '0',
                  style: TextStyle(
                    color: _AutismDevColors.ink,
                    fontSize: 27,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                TextSpan(
                  text: ' / 0',
                  style: TextStyle(
                    color: _AutismDevColors.body,
                    fontSize: 15,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
          const Spacer(),
          _FooterButton(
            label: '下一题',
            icon: Icons.chevron_right_rounded,
            enabled: false,
            filled: true,
            reverseIcon: true,
            onTap: () {},
          ),
          const SizedBox(width: 14),
          _FooterButton(
            label: '跳到缺题',
            icon: Icons.swipe_right_alt_rounded,
            enabled: false,
            onTap: () {},
          ),
          const SizedBox(width: 22),
          const Text(
            '自动下一题',
            style: TextStyle(
              color: _AutismDevColors.body,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(width: 8),
          const Switch(
            value: true,
            activeColor: _AutismDevColors.orange,
            onChanged: null,
          ),
        ],
      ),
    );
  }
}

class _AutismDevSkeletonBlock extends StatelessWidget {
  const _AutismDevSkeletonBlock({
    this.width,
    this.widthFactor,
    required this.height,
    this.radius = 6,
  });

  final double? width;
  final double? widthFactor;
  final double height;
  final double radius;

  @override
  Widget build(BuildContext context) {
    final Widget block = Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: const Color(0xFFF3E6DB),
        borderRadius: BorderRadius.circular(radius),
      ),
    );
    final double? factor = widthFactor;
    if (factor != null) {
      return FractionallySizedBox(
        widthFactor: factor,
        alignment: Alignment.centerLeft,
        child: block,
      );
    }
    return block;
  }
}
