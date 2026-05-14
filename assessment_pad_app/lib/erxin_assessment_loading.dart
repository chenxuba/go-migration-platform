part of 'erxin_assessment_page.dart';

class _ErxinLoadingDomainRow extends StatelessWidget {
  const _ErxinLoadingDomainRow({required this.selected});

  final bool selected;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 9),
      child: Container(
        height: 66,
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        decoration: BoxDecoration(
          color: selected ? const Color(0xFFFFEEE5) : Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: _ErxinColors.line),
        ),
        child: Column(
          children: <Widget>[
            Row(
              children: <Widget>[
                _ErxinSkeletonBlock(
                  width: 24,
                  height: 24,
                  radius: 7,
                  highlight: selected,
                ),
                const SizedBox(width: 8),
                const Expanded(
                  child: _ErxinSkeletonBlock(height: 14),
                ),
                const SizedBox(width: 10),
                const _ErxinSkeletonBlock(width: 34, height: 12, radius: 6),
              ],
            ),
            const Spacer(),
            Row(
              children: const <Widget>[
                Expanded(
                  child: _ErxinSkeletonBlock(height: 4, radius: 2),
                ),
                SizedBox(width: 9),
                _ErxinSkeletonBlock(width: 28, height: 12, radius: 6),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _ErxinLoadingProgressSummary extends StatelessWidget {
  const _ErxinLoadingProgressSummary();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: _erxinProgressSummaryHeight,
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: const <Widget>[
          _ErxinSkeletonBlock(width: 96, height: 12, radius: 6),
          SizedBox(height: 10),
          _ErxinSkeletonBlock(widthFactor: .82, height: 12),
          SizedBox(height: 8),
          _ErxinSkeletonBlock(widthFactor: .6, height: 12),
        ],
      ),
    );
  }
}

class _ErxinLoadingMonthSection extends StatelessWidget {
  const _ErxinLoadingMonthSection({required this.rowCount});

  final int rowCount;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 6),
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      foregroundDecoration: BoxDecoration(
        border: Border.all(color: _ErxinColors.line),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: <Widget>[
          Container(
            height: 34,
            padding: const EdgeInsets.symmetric(horizontal: 12),
            decoration: const BoxDecoration(
              color: Color(0xFFFFFAF5),
              borderRadius: BorderRadius.vertical(top: Radius.circular(8)),
              border: Border(bottom: BorderSide(color: _ErxinColors.line)),
            ),
            child: const Row(
              children: <Widget>[
                _ErxinSkeletonBlock(width: 74, height: 14),
                Spacer(),
                _ErxinSkeletonBlock(width: 54, height: 12, radius: 6),
              ],
            ),
          ),
          for (int index = 0; index < rowCount; index++)
            const _ErxinLoadingItemRow(),
        ],
      ),
    );
  }
}

class _ErxinLoadingItemRow extends StatelessWidget {
  const _ErxinLoadingItemRow();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 48,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(bottom: BorderSide(color: _ErxinColors.line)),
      ),
      child: Row(
        children: const <Widget>[
          SizedBox(
            width: 42,
            child: _ErxinSkeletonBlock(height: 14),
          ),
          Expanded(
            child: _ErxinSkeletonBlock(height: 14),
          ),
          SizedBox(width: 8),
          _ErxinSkeletonBlock(width: 64, height: 28, radius: 10),
          SizedBox(width: 8),
          _ErxinSkeletonBlock(width: 74, height: 28, radius: 10),
        ],
      ),
    );
  }
}

class _ErxinLoadingNextCard extends StatelessWidget {
  const _ErxinLoadingNextCard();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 108,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: const <Widget>[
          _ErxinSkeletonBlock(width: 48, height: 12, radius: 6),
          SizedBox(height: 10),
          _ErxinSkeletonBlock(height: 14),
          SizedBox(height: 8),
          _ErxinSkeletonBlock(widthFactor: .76, height: 12),
          SizedBox(height: 8),
          _ErxinSkeletonBlock(widthFactor: .56, height: 12),
        ],
      ),
    );
  }
}

class _ErxinLoadingRecordList extends StatelessWidget {
  const _ErxinLoadingRecordList();

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      padding: EdgeInsets.zero,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: 5,
      separatorBuilder: (_, __) => const SizedBox(height: 8),
      itemBuilder: (BuildContext context, int index) {
        return _ErxinLoadingRecordRow(selected: index == 1);
      },
    );
  }
}

class _ErxinLoadingRecordRow extends StatelessWidget {
  const _ErxinLoadingRecordRow({required this.selected});

  final bool selected;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 42,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: selected ? const Color(0xFFFFF3E8) : const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Row(
        children: const <Widget>[
          _ErxinSkeletonBlock(width: 16, height: 16, radius: 8),
          SizedBox(width: 8),
          Expanded(child: _ErxinSkeletonBlock(height: 12)),
          SizedBox(width: 8),
          _ErxinSkeletonBlock(width: 36, height: 12, radius: 6),
        ],
      ),
    );
  }
}

class _ErxinLoadingDetailPanel extends StatelessWidget {
  const _ErxinLoadingDetailPanel();

  @override
  Widget build(BuildContext context) {
    return Container(
      key: const ValueKey<String>('erxin-loading-detail'),
      height: _erxinDetailPanelHeight,
      padding: const EdgeInsets.fromLTRB(
        16,
        _erxinDetailPanelTopPadding,
        12,
        _erxinDetailPanelBottomPadding,
      ),
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(top: BorderSide(color: _ErxinColors.line)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              const Expanded(
                child: Text(
                  '当前题目说明：',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: _ErxinColors.ink,
                    fontSize: 16,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              SizedBox(
                height: 28,
                child: OutlinedButton(
                  onPressed: null,
                  child: const Text('完整说明'),
                ),
              ),
            ],
          ),
          const SizedBox(height: _erxinDetailHeaderGap),
          const Expanded(
            child: Row(
              children: <Widget>[
                Expanded(
                  child: _ErxinLoadingTextPanel(title: '操作方法'),
                ),
                SizedBox(width: 10),
                Expanded(
                  child: _ErxinLoadingTextPanel(title: '通过标准'),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ErxinLoadingTextPanel extends StatelessWidget {
  const _ErxinLoadingTextPanel({required this.title});

  final String title;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            title,
            style: const TextStyle(
              color: _ErxinColors.ink,
              fontSize: 14,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 8),
          const _ErxinSkeletonBlock(height: 12),
          const SizedBox(height: 7),
          const _ErxinSkeletonBlock(widthFactor: .92, height: 12),
          const SizedBox(height: 7),
          const _ErxinSkeletonBlock(widthFactor: .7, height: 12),
        ],
      ),
    );
  }
}

class _ErxinSkeletonBlock extends StatelessWidget {
  const _ErxinSkeletonBlock({
    this.width,
    this.widthFactor = 1,
    required this.height,
    this.radius = 6,
    this.highlight = false,
  });

  final double? width;
  final double widthFactor;
  final double height;
  final double radius;
  final bool highlight;

  @override
  Widget build(BuildContext context) {
    final Color fill =
        highlight ? const Color(0xFFFFE5D3) : const Color(0xFFF3E3D8);
    final Widget block = Container(
      height: height,
      decoration: BoxDecoration(
        color: fill,
        borderRadius: BorderRadius.circular(radius),
      ),
    );
    if (width != null) {
      return SizedBox(width: width, child: block);
    }
    return FractionallySizedBox(
      widthFactor: widthFactor,
      alignment: Alignment.centerLeft,
      child: block,
    );
  }
}
