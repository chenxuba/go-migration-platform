import 'dart:math' as math;

import 'package:flutter/material.dart';

class TrainingCenterPage extends StatelessWidget {
  const TrainingCenterPage({required this.onBack, super.key});

  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width =
            constraints.maxWidth.isFinite ? constraints.maxWidth : 1366;
        final bool compact = width < 1180;
        final _TrainingMetrics metrics = _TrainingMetrics.forWidth(width);

        return ColoredBox(
          color: _TrainingColors.page,
          child: Stack(
            children: <Widget>[
              Positioned.fill(child: CustomPaint(painter: _PageGlowPainter())),
              Positioned(
                left: 0,
                top: 0,
                right: 0,
                child: _TrainingTopBar(
                  onBack: onBack,
                  compact: compact,
                  horizontalPadding: metrics.outer,
                ),
              ),
              Positioned(
                left: metrics.outer,
                top: 86,
                width: metrics.leftWidth,
                height: 664,
                child: _LeftRail(compact: compact),
              ),
              Positioned(
                left: metrics.centerLeft,
                top: 86,
                width: metrics.centerWidth,
                height: 664,
                child: const _CenterContent(),
              ),
            ],
          ),
        );
      },
    );
  }
}

class _TrainingMetrics {
  const _TrainingMetrics({
    required this.outer,
    required this.gap,
    required this.leftWidth,
    required this.centerLeft,
    required this.centerWidth,
  });

  factory _TrainingMetrics.forWidth(double width) {
    final bool compact = width < 1180;
    final double outer = compact ? 12 : 18;
    final double gap = compact ? 8 : 12;
    final double leftWidth = compact ? 208 : 236;
    final double centerLeft = outer + leftWidth + gap;
    final double centerWidth = width - outer * 2 - leftWidth - gap;

    return _TrainingMetrics(
      outer: outer,
      gap: gap,
      leftWidth: leftWidth,
      centerLeft: centerLeft,
      centerWidth: centerWidth,
    );
  }

  final double outer;
  final double gap;
  final double leftWidth;
  final double centerLeft;
  final double centerWidth;
}

class _TrainingColors {
  static const Color page = Color(0xFFFAFAF8);
  static const Color surface = Color(0xFFFFFFFF);
  static const Color ink = Color(0xFF181B20);
  static const Color text = Color(0xFF3F454E);
  static const Color muted = Color(0xFF858B96);
  static const Color line = Color(0xFFE6E8EC);
  static const Color orange = Color(0xFFFF6B12);
  static const Color orangeSoft = Color(0xFFFFF2E7);
  static const Color blue = Color(0xFF2E79F6);
  static const Color green = Color(0xFF20A856);
  static const Color teal = Color(0xFF129A91);
  static const Color purple = Color(0xFF7B55E6);
  static const Color red = Color(0xFFEF4D4D);
  static const Color yellow = Color(0xFFFFB000);
}

List<BoxShadow> _panelShadow() {
  return <BoxShadow>[
    BoxShadow(
      color: const Color(0xFF1D2433).withOpacity(.06),
      blurRadius: 18,
      offset: const Offset(0, 8),
    ),
  ];
}

class _TrainingTopBar extends StatelessWidget {
  const _TrainingTopBar({
    required this.onBack,
    required this.compact,
    required this.horizontalPadding,
  });

  final VoidCallback onBack;
  final bool compact;
  final double horizontalPadding;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 72,
      padding: EdgeInsets.symmetric(horizontal: horizontalPadding),
      decoration: BoxDecoration(
        color: _TrainingColors.surface,
        border: Border(
          bottom: BorderSide(color: _TrainingColors.line.withOpacity(.78)),
        ),
      ),
      child: Row(
        children: <Widget>[
          _TrainingBackButton(onTap: onBack),
          const SizedBox(width: 16),
          const Text(
            '训练中心',
            style: TextStyle(
              color: _TrainingColors.ink,
              fontSize: 25,
              fontWeight: FontWeight.w900,
              height: 1,
            ),
          ),
          const Spacer(),
          if (!compact) ...<Widget>[
            const _StudentSelector(),
            const SizedBox(width: 14),
            const _StageBadge(),
            const SizedBox(width: 110),
          ],
          _SearchField(width: compact ? 210 : 224),
          const SizedBox(width: 14),
          const _FilterButton(),
          const SizedBox(width: 16),
          const _NotificationButton(),
        ],
      ),
    );
  }
}

class _StudentSelector extends StatelessWidget {
  const _StudentSelector();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 184,
      height: 38,
      padding: const EdgeInsets.fromLTRB(10, 5, 10, 5),
      decoration: BoxDecoration(
        color: _TrainingColors.surface,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _TrainingColors.line),
      ),
      child: Row(
        children: <Widget>[
          const _StudentAvatar(size: 28),
          const SizedBox(width: 12),
          const Expanded(
            child: Text(
              '陈小宇 · 6岁',
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: _TrainingColors.ink,
                fontSize: 14,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          Icon(
            Icons.keyboard_arrow_down_rounded,
            color: _TrainingColors.ink.withOpacity(.9),
            size: 22,
          ),
        ],
      ),
    );
  }
}

class _StudentAvatar extends StatelessWidget {
  const _StudentAvatar({required this.size});

  final double size;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: <Color>[Color(0xFFFFD5AA), Color(0xFFFFA15F)],
        ),
        border: Border.all(color: Colors.white, width: 2),
      ),
      child: Icon(
        Icons.face_rounded,
        size: size * .72,
        color: const Color(0xFF5B392A),
      ),
    );
  }
}

class _StageBadge extends StatelessWidget {
  const _StageBadge();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 32,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF7EF),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: const Color(0xFFFFBC82)),
      ),
      child: const Text(
        '训练阶段：基础巩固期',
        style: TextStyle(
          color: _TrainingColors.orange,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _SearchField extends StatelessWidget {
  const _SearchField({required this.width});

  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: BoxDecoration(
        color: _TrainingColors.surface,
        borderRadius: BorderRadius.circular(19),
        border: Border.all(color: _TrainingColors.line),
      ),
      child: Row(
        children: const <Widget>[
          Icon(Icons.search_rounded, color: Color(0xFF20242A), size: 22),
          SizedBox(width: 9),
          Expanded(
            child: Text(
              '搜索训练游戏',
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: _TrainingColors.muted,
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

class _FilterButton extends StatelessWidget {
  const _FilterButton();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 78,
      height: 38,
      decoration: BoxDecoration(
        color: _TrainingColors.surface,
        borderRadius: BorderRadius.circular(19),
        border: Border.all(color: _TrainingColors.line),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: const <Widget>[
          Icon(Icons.filter_alt_outlined, color: _TrainingColors.ink, size: 22),
          SizedBox(width: 6),
          Text(
            '筛选',
            style: TextStyle(
              color: _TrainingColors.ink,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _NotificationButton extends StatelessWidget {
  const _NotificationButton();

  @override
  Widget build(BuildContext context) {
    return Stack(
      clipBehavior: Clip.none,
      children: <Widget>[
        Container(
          width: 38,
          height: 38,
          alignment: Alignment.center,
          child: const Icon(
            Icons.notifications_none_rounded,
            color: _TrainingColors.ink,
            size: 27,
          ),
        ),
        Positioned(
          right: -2,
          top: -5,
          child: Container(
            width: 19,
            height: 19,
            alignment: Alignment.center,
            decoration: const BoxDecoration(
              color: Color(0xFFFF3D30),
              shape: BoxShape.circle,
            ),
            child: const Text(
              '3',
              style: TextStyle(
                color: Colors.white,
                fontSize: 11,
                fontWeight: FontWeight.w900,
                height: 1,
              ),
            ),
          ),
        ),
      ],
    );
  }
}

class _TrainingBackButton extends StatelessWidget {
  const _TrainingBackButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: _TrainingColors.surface.withOpacity(.94),
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Container(
          width: 42,
          height: 42,
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: _TrainingColors.line),
          ),
          child: const Icon(
            Icons.chevron_left_rounded,
            color: _TrainingColors.text,
            size: 28,
          ),
        ),
      ),
    );
  }
}

class _LeftRail extends StatelessWidget {
  const _LeftRail({required this.compact});

  final bool compact;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        _CategoryPanel(compact: compact),
        const SizedBox(height: 12),
        const Expanded(child: _ProgressPanel()),
      ],
    );
  }
}

class _CategoryPanel extends StatelessWidget {
  const _CategoryPanel({required this.compact});

  final bool compact;

  @override
  Widget build(BuildContext context) {
    return _Panel(
      height: compact ? 372 : 372,
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 12),
      child: Column(
        children: List<Widget>.generate(_categoryItems.length, (int index) {
          final _CategoryItem item = _categoryItems[index];
          return Padding(
            padding: EdgeInsets.only(
                bottom: index == _categoryItems.length - 1 ? 0 : 5),
            child: _CategoryTile(item: item, selected: index == 0),
          );
        }),
      ),
    );
  }
}

class _CategoryTile extends StatelessWidget {
  const _CategoryTile({required this.item, required this.selected});

  final _CategoryItem item;
  final bool selected;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 43,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: selected ? _TrainingColors.orangeSoft : Colors.transparent,
        borderRadius: BorderRadius.circular(14),
      ),
      child: Row(
        children: <Widget>[
          Container(
            width: 10,
            height: 10,
            decoration:
                BoxDecoration(color: item.color, shape: BoxShape.circle),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              item.label,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: selected ? _TrainingColors.orange : _TrainingColors.ink,
                fontSize: 15,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          const SizedBox(width: 10),
          Text(
            '${item.count}',
            style: TextStyle(
              color: selected ? _TrainingColors.orange : _TrainingColors.muted,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _CategoryItem {
  const _CategoryItem({
    required this.label,
    required this.color,
    required this.count,
  });

  final String label;
  final Color color;
  final int count;
}

const List<_CategoryItem> _categoryItems = <_CategoryItem>[
  _CategoryItem(
    label: '全部游戏',
    color: _TrainingColors.orange,
    count: 24,
  ),
  _CategoryItem(
    label: '认知理解',
    color: _TrainingColors.blue,
    count: 6,
  ),
  _CategoryItem(
    label: '语言表达',
    color: _TrainingColors.green,
    count: 5,
  ),
  _CategoryItem(
    label: '精细动作',
    color: _TrainingColors.orange,
    count: 4,
  ),
  _CategoryItem(
    label: '感统协调',
    color: _TrainingColors.teal,
    count: 3,
  ),
  _CategoryItem(
    label: '社交互动',
    color: _TrainingColors.purple,
    count: 4,
  ),
  _CategoryItem(
    label: '情绪管理',
    color: _TrainingColors.red,
    count: 2,
  ),
];

class _ProgressPanel extends StatelessWidget {
  const _ProgressPanel();

  @override
  Widget build(BuildContext context) {
    return _Panel(
      padding: const EdgeInsets.fromLTRB(18, 18, 18, 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '今日训练进度',
            style: TextStyle(
              color: _TrainingColors.ink,
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
          const Spacer(),
          Center(
            child: SizedBox(
              width: 142,
              height: 142,
              child: CustomPaint(
                painter: _DonutProgressPainter(progress: .6),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: const <Widget>[
                    Text.rich(
                      TextSpan(
                        children: <InlineSpan>[
                          TextSpan(
                            text: '6',
                            style: TextStyle(
                              color: _TrainingColors.orange,
                              fontSize: 36,
                              fontWeight: FontWeight.w900,
                              height: 1,
                            ),
                          ),
                          TextSpan(
                            text: '/10',
                            style: TextStyle(
                              color: _TrainingColors.text,
                              fontSize: 24,
                              fontWeight: FontWeight.w700,
                              height: 1,
                            ),
                          ),
                        ],
                      ),
                    ),
                    SizedBox(height: 8),
                    Text(
                      '已完成项目',
                      style: TextStyle(
                        color: _TrainingColors.muted,
                        fontSize: 13,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    SizedBox(height: 7),
                    Icon(Icons.star_rounded,
                        color: _TrainingColors.yellow, size: 19),
                  ],
                ),
              ),
            ),
          ),
          const Spacer(),
          Center(
            child: Column(
              children: const <Widget>[
                Text(
                  '持续努力，明天会更棒！',
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    color: _TrainingColors.muted,
                    fontSize: 12,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                SizedBox(height: 14),
                _DetailLink(),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _DetailLink extends StatelessWidget {
  const _DetailLink();

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: const <Widget>[
        Text(
          '查看详情',
          style: TextStyle(
            color: _TrainingColors.blue,
            fontSize: 12,
            fontWeight: FontWeight.w800,
          ),
        ),
        SizedBox(width: 3),
        Icon(
          Icons.keyboard_arrow_right_rounded,
          color: _TrainingColors.blue,
          size: 18,
        ),
      ],
    );
  }
}

class _CenterContent extends StatelessWidget {
  const _CenterContent();

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final bool wide = constraints.maxWidth >= 980;
        if (!wide) {
          return const Column(
            children: <Widget>[
              _RecommendedPanel(height: 256),
              SizedBox(height: 10),
              Expanded(child: _GameCollectionPanel()),
            ],
          );
        }

        return const Column(
          children: <Widget>[
            _RecommendedPanel(height: 256),
            SizedBox(height: 10),
            Expanded(
              child: Row(
                children: <Widget>[
                  Expanded(child: _GameCollectionPanel()),
                  SizedBox(width: 12),
                  SizedBox(width: 340, child: _TrainingFocusColumn()),
                ],
              ),
            ),
          ],
        );
      },
    );
  }
}

class _RecommendedPanel extends StatelessWidget {
  const _RecommendedPanel({required this.height});

  final double height;

  @override
  Widget build(BuildContext context) {
    return _Panel(
      height: height,
      padding: const EdgeInsets.fromLTRB(14, 13, 14, 12),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final double gap = constraints.maxWidth >= 900 ? 16 : 12;
          const int columns = 4;
          final double cardWidth =
              (constraints.maxWidth - gap * (columns - 1)) / columns;

          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              const _PanelTitleRow(
                title: '推荐训练游戏',
                leadingColor: _TrainingColors.orange,
                action: _RefreshAction(),
              ),
              const SizedBox(height: 12),
              Row(
                children: <Widget>[
                  for (int index = 0; index < _recommendedGames.length; index++)
                    Padding(
                      padding: EdgeInsets.only(
                        right: index == _recommendedGames.length - 1 ? 0 : gap,
                      ),
                      child: _RecommendedGameCard(
                        width: cardWidth,
                        game: _recommendedGames[index],
                      ),
                    ),
                ],
              ),
            ],
          );
        },
      ),
    );
  }
}

class _RefreshAction extends StatelessWidget {
  const _RefreshAction();

  @override
  Widget build(BuildContext context) {
    return Row(
      children: const <Widget>[
        Text(
          '换一换',
          style: TextStyle(
            color: _TrainingColors.text,
            fontSize: 13,
            fontWeight: FontWeight.w700,
          ),
        ),
        SizedBox(width: 4),
        Icon(Icons.refresh_rounded, size: 18, color: _TrainingColors.text),
      ],
    );
  }
}

class _RecommendedGameCard extends StatelessWidget {
  const _RecommendedGameCard({required this.width, required this.game});

  final double width;
  final _RecommendedGame game;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 190,
      decoration: BoxDecoration(
        color: _TrainingColors.surface,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _TrainingColors.line),
      ),
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          SizedBox(
            height: 92,
            width: double.infinity,
            child: CustomPaint(
              painter: _HeroIllustrationPainter(game.illustration),
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(14, 8, 14, 0),
            child: Text(
              game.title,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _TrainingColors.ink,
                fontSize: 15,
                fontWeight: FontWeight.w900,
                height: 1.05,
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.fromLTRB(14, 5, 14, 0),
            child: SizedBox(
              height: 21,
              child: FittedBox(
                fit: BoxFit.scaleDown,
                alignment: Alignment.centerLeft,
                child: Row(
                  children: <Widget>[
                    for (int index = 0; index < game.tags.length; index++)
                      Padding(
                        padding: EdgeInsets.only(
                          right: index == game.tags.length - 1 ? 0 : 6,
                        ),
                        child: _TagChip(
                          label: game.tags[index].label,
                          color: game.tags[index].color,
                        ),
                      ),
                  ],
                ),
              ),
            ),
          ),
          const Spacer(),
          Padding(
            padding: const EdgeInsets.fromLTRB(14, 0, 14, 10),
            child: Row(
              children: <Widget>[
                _RatingStars(score: game.rating, size: width < 210 ? 14 : 17),
                const Spacer(),
                _StartButton(compact: width < 210),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _StartButton extends StatelessWidget {
  const _StartButton({this.compact = false});

  final bool compact;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: compact ? 72 : 80,
      height: 28,
      decoration: BoxDecoration(
        color: _TrainingColors.orange,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: const <Widget>[
          Icon(Icons.play_arrow_rounded, color: Colors.white, size: 18),
          SizedBox(width: 2),
          Text(
            '开始游戏',
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

class _RatingStars extends StatelessWidget {
  const _RatingStars({required this.score, this.size = 17});

  final int score;
  final double size;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: List<Widget>.generate(5, (int index) {
        return Icon(
          index < score ? Icons.star_rounded : Icons.star_border_rounded,
          size: size,
          color: index < score
              ? _TrainingColors.yellow
              : _TrainingColors.muted.withOpacity(.65),
        );
      }),
    );
  }
}

class _GameCollectionPanel extends StatelessWidget {
  const _GameCollectionPanel();

  @override
  Widget build(BuildContext context) {
    return _Panel(
      padding: const EdgeInsets.fromLTRB(14, 12, 14, 10),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          const double gap = 12;
          const int columns = 3;
          const int rows = 2;
          final List<_CollectionGame> visibleGames =
              _collectionGames.take(columns * rows).toList();
          final double cardWidth =
              (constraints.maxWidth - gap * (columns - 1)) / columns;
          final double cardHeight = (constraints.maxHeight - 40 - gap - 31) / 2;

          return Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              const _PanelTitleRow(title: '游戏合集'),
              const SizedBox(height: 10),
              for (int row = 0; row < rows; row++)
                Padding(
                  padding: EdgeInsets.only(bottom: row == rows - 1 ? 0 : gap),
                  child: Row(
                    children: <Widget>[
                      for (int col = 0; col < columns; col++)
                        Padding(
                          padding: EdgeInsets.only(
                              right: col == columns - 1 ? 0 : gap),
                          child: _CollectionGameCard(
                            width: cardWidth,
                            height: cardHeight,
                            game: visibleGames[row * columns + col],
                          ),
                        ),
                    ],
                  ),
                ),
              const Spacer(),
              const Center(child: _MoreGamesLink()),
            ],
          );
        },
      ),
    );
  }
}

class _CollectionGameCard extends StatelessWidget {
  const _CollectionGameCard({
    required this.width,
    required this.height,
    required this.game,
  });

  final double width;
  final double height;
  final _CollectionGame game;

  @override
  Widget build(BuildContext context) {
    final bool compact = width < 174;
    final double padding = compact ? 8 : 10;
    final double imageSize = compact ? 46 : 58;
    final double horizontalGap = compact ? 8 : 10;
    final double titleFontSize = compact ? 13 : 14;
    final double contentGap = compact ? 5 : 8;
    final Color progressColor = game.status == _GameStatus.practice
        ? _TrainingColors.orange
        : game.color;

    return Container(
      width: width,
      height: height,
      padding: EdgeInsets.all(padding),
      decoration: BoxDecoration(
        color: _TrainingColors.surface,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _TrainingColors.line),
      ),
      child: Column(
        children: <Widget>[
          Expanded(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                _MiniGameImage(
                  color: game.color,
                  illustration: game.illustration,
                  size: imageSize,
                ),
                SizedBox(width: horizontalGap),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        game.title,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(
                          color: _TrainingColors.ink,
                          fontSize: titleFontSize,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                      const SizedBox(height: 4),
                      _TagChip(
                        label: game.tag.label,
                        color: game.tag.color,
                      ),
                      const SizedBox(height: 4),
                      Text.rich(
                        TextSpan(
                          children: <InlineSpan>[
                            const TextSpan(text: '最近表现：'),
                            TextSpan(
                              text: game.performance,
                              style: TextStyle(
                                color: game.performanceColor,
                                fontWeight: FontWeight.w900,
                              ),
                            ),
                          ],
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: const TextStyle(
                          color: _TrainingColors.text,
                          fontSize: 10,
                          fontWeight: FontWeight.w700,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          SizedBox(height: contentGap),
          Row(
            children: <Widget>[
              Expanded(
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(8),
                  child: LinearProgressIndicator(
                    minHeight: 6,
                    value: game.progress,
                    backgroundColor: _TrainingColors.line,
                    valueColor: AlwaysStoppedAnimation<Color>(progressColor),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Text(
                '${(game.progress * 100).round()}%',
                style: const TextStyle(
                  color: _TrainingColors.text,
                  fontSize: 11,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ],
          ),
          SizedBox(height: contentGap),
          _GameStateLabel(status: game.status, color: progressColor),
        ],
      ),
    );
  }
}

class _MiniGameImage extends StatelessWidget {
  const _MiniGameImage({
    required this.color,
    required this.illustration,
    required this.size,
  });

  final Color color;
  final _MiniIllustration illustration;
  final double size;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: color.withOpacity(.16),
        borderRadius: BorderRadius.circular(8),
      ),
      child:
          CustomPaint(painter: _MiniIllustrationPainter(illustration, color)),
    );
  }
}

class _GameStateLabel extends StatelessWidget {
  const _GameStateLabel({required this.status, required this.color});

  final _GameStatus status;
  final Color color;

  @override
  Widget build(BuildContext context) {
    final IconData icon;
    final String label;
    switch (status) {
      case _GameStatus.done:
        icon = Icons.check_circle_outline_rounded;
        label = '已完成';
      case _GameStatus.playing:
        icon = Icons.timelapse_rounded;
        label = '进行中';
      case _GameStatus.practice:
        icon = Icons.schedule_rounded;
        label = '待练习';
    }

    return Row(
      children: <Widget>[
        Icon(icon, size: 15, color: color),
        const SizedBox(width: 4),
        Text(
          label,
          style: TextStyle(
            color: color,
            fontSize: 11,
            fontWeight: FontWeight.w900,
          ),
        ),
      ],
    );
  }
}

class _MoreGamesLink extends StatelessWidget {
  const _MoreGamesLink();

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: const <Widget>[
        Text(
          '查看更多游戏',
          style: TextStyle(
            color: _TrainingColors.blue,
            fontSize: 13,
            fontWeight: FontWeight.w800,
          ),
        ),
        SizedBox(width: 5),
        Icon(Icons.expand_more_rounded, color: _TrainingColors.blue, size: 20),
      ],
    );
  }
}

class _TrainingFocusColumn extends StatelessWidget {
  const _TrainingFocusColumn();

  @override
  Widget build(BuildContext context) {
    return const Column(
      children: <Widget>[
        _TodayTaskPanel(),
        SizedBox(height: 12),
        Expanded(child: _RecentRecordsPanel()),
      ],
    );
  }
}

class _TodayTaskPanel extends StatelessWidget {
  const _TodayTaskPanel();

  @override
  Widget build(BuildContext context) {
    return _Panel(
      height: 200,
      padding: const EdgeInsets.fromLTRB(16, 13, 16, 12),
      child: Column(
        children: <Widget>[
          const _PanelTitleRow(title: '今日任务', actionText: '3/4'),
          const SizedBox(height: 8),
          for (int index = 0; index < _todayTasks.length; index++)
            _TaskRow(task: _todayTasks[index]),
        ],
      ),
    );
  }
}

class _TaskRow extends StatelessWidget {
  const _TaskRow({required this.task});

  final _TrainingTask task;

  @override
  Widget build(BuildContext context) {
    final bool done = task.state == _TaskState.done;
    final Color actionColor = done
        ? _TrainingColors.green
        : task.state == _TaskState.continueTask
            ? _TrainingColors.orange
            : _TrainingColors.orange;

    return Expanded(
      child: Row(
        children: <Widget>[
          Icon(
            done
                ? Icons.check_circle_rounded
                : Icons.radio_button_unchecked_rounded,
            color: done ? _TrainingColors.green : _TrainingColors.text,
            size: 22,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              task.title,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: done ? _TrainingColors.muted : _TrainingColors.text,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          const SizedBox(width: 8),
          SizedBox(
            width: 68,
            child: Text(
              task.minutes,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _TrainingColors.text,
                fontSize: 12,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
          const SizedBox(width: 10),
          Container(
            width: 66,
            height: 28,
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(6),
              border: Border.all(color: actionColor.withOpacity(.45)),
            ),
            child: Text(
              task.button,
              style: TextStyle(
                color: actionColor,
                fontSize: 13,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _RecentRecordsPanel extends StatelessWidget {
  const _RecentRecordsPanel();

  @override
  Widget build(BuildContext context) {
    return _Panel(
      padding: const EdgeInsets.fromLTRB(16, 13, 16, 12),
      child: Column(
        children: <Widget>[
          const _PanelTitleRow(title: '最近记录', action: _MoreAction()),
          const SizedBox(height: 10),
          for (int index = 0; index < _recentRecords.length; index++)
            _RecordRow(
              record: _recentRecords[index],
              showDivider: index < _recentRecords.length - 1,
            ),
        ],
      ),
    );
  }
}

class _RecordRow extends StatelessWidget {
  const _RecordRow({required this.record, required this.showDivider});

  final _RecentRecord record;
  final bool showDivider;

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: DecoratedBox(
        decoration: BoxDecoration(
          border: Border(
            bottom: showDivider
                ? BorderSide(color: _TrainingColors.line.withOpacity(.8))
                : BorderSide.none,
          ),
        ),
        child: Row(
          children: <Widget>[
            Container(
              width: 32,
              height: 32,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: record.color.withOpacity(.14),
              ),
              child: Icon(record.icon, size: 19, color: record.color),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                record.title,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _TrainingColors.text,
                  fontSize: 13,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
            const SizedBox(width: 8),
            SizedBox(
              width: 86,
              child: Text(
                record.time,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _TrainingColors.muted,
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
            SizedBox(
              width: 74,
              child: Text(
                '正确率  ${record.accuracy}%',
                textAlign: TextAlign.right,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: _TrainingColors.text,
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _PanelTitleRow extends StatelessWidget {
  const _PanelTitleRow({
    required this.title,
    this.leadingColor,
    this.action,
    this.actionText,
  });

  final String title;
  final Color? leadingColor;
  final Widget? action;
  final String? actionText;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 24,
      child: Row(
        children: <Widget>[
          if (leadingColor != null) ...<Widget>[
            Container(
              width: 5,
              height: 22,
              decoration: BoxDecoration(
                color: leadingColor,
                borderRadius: BorderRadius.circular(5),
              ),
            ),
            const SizedBox(width: 9),
          ],
          Text(
            title,
            style: const TextStyle(
              color: _TrainingColors.ink,
              fontSize: 17,
              fontWeight: FontWeight.w900,
              height: 1,
            ),
          ),
          const Spacer(),
          if (action != null)
            action!
          else if (actionText != null)
            Text(
              actionText!,
              style: const TextStyle(
                color: _TrainingColors.text,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
        ],
      ),
    );
  }
}

class _MoreAction extends StatelessWidget {
  const _MoreAction();

  @override
  Widget build(BuildContext context) {
    return Row(
      children: const <Widget>[
        Text(
          '更多',
          style: TextStyle(
            color: _TrainingColors.text,
            fontSize: 12,
            fontWeight: FontWeight.w700,
          ),
        ),
        SizedBox(width: 3),
        Icon(
          Icons.keyboard_arrow_right_rounded,
          color: _TrainingColors.text,
          size: 17,
        ),
      ],
    );
  }
}

class _TagChip extends StatelessWidget {
  const _TagChip({
    required this.label,
    required this.color,
  });

  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) {
    return IntrinsicWidth(
      child: Container(
        height: 19,
        padding: const EdgeInsets.symmetric(horizontal: 7),
        alignment: Alignment.center,
        decoration: BoxDecoration(
          color: color.withOpacity(.04),
          borderRadius: BorderRadius.circular(6),
          border: Border.all(color: color.withOpacity(.10)),
        ),
        child: Transform.translate(
          offset: const Offset(0, .5),
          child: Text(
            label,
            textAlign: TextAlign.center,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            strutStyle: const StrutStyle(
              fontSize: 10,
              height: 1,
              forceStrutHeight: true,
            ),
            style: TextStyle(
              color: color.withOpacity(.84),
              fontSize: 10,
              fontWeight: FontWeight.w700,
              height: 1,
            ),
          ),
        ),
      ),
    );
  }
}

class _Panel extends StatelessWidget {
  const _Panel({
    required this.child,
    this.height,
    this.padding = EdgeInsets.zero,
  });

  final Widget child;
  final double? height;
  final EdgeInsetsGeometry padding;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: height,
      width: double.infinity,
      padding: padding,
      decoration: BoxDecoration(
        color: _TrainingColors.surface,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _TrainingColors.line),
        boxShadow: _panelShadow(),
      ),
      child: child,
    );
  }
}

class _RecommendedGame {
  const _RecommendedGame({
    required this.title,
    required this.rating,
    required this.tags,
    required this.illustration,
  });

  final String title;
  final int rating;
  final List<_Tag> tags;
  final _HeroIllustration illustration;
}

class _CollectionGame {
  const _CollectionGame({
    required this.title,
    required this.tag,
    required this.performance,
    required this.performanceColor,
    required this.progress,
    required this.status,
    required this.color,
    required this.illustration,
  });

  final String title;
  final _Tag tag;
  final String performance;
  final Color performanceColor;
  final double progress;
  final _GameStatus status;
  final Color color;
  final _MiniIllustration illustration;
}

class _Tag {
  const _Tag({required this.label, required this.color});

  final String label;
  final Color color;
}

enum _HeroIllustration { rainbow, train, maze, rhythm }

enum _MiniIllustration {
  animals,
  ear,
  shapes,
  emotion,
  rhythm,
  story,
  numbers,
  cards,
}

enum _GameStatus { done, playing, practice }

const _Tag _cognitionTag = _Tag(label: '认知理解', color: _TrainingColors.blue);
const _Tag _languageTag = _Tag(label: '语言表达', color: _TrainingColors.green);
const _Tag _fineMotorTag = _Tag(label: '精细动作', color: _TrainingColors.orange);
const _Tag _sensoryTag = _Tag(label: '感统协调', color: _TrainingColors.teal);
const _Tag _emotionTag = _Tag(label: '情绪管理', color: _TrainingColors.red);

const List<_RecommendedGame> _recommendedGames = <_RecommendedGame>[
  _RecommendedGame(
    title: '颜色配对乐园',
    rating: 4,
    tags: <_Tag>[
      _cognitionTag,
      _Tag(label: '视觉分辨', color: _TrainingColors.blue)
    ],
    illustration: _HeroIllustration.rainbow,
  ),
  _RecommendedGame(
    title: '词语小火车',
    rating: 4,
    tags: <_Tag>[
      _languageTag,
      _Tag(label: '词汇理解', color: _TrainingColors.green)
    ],
    illustration: _HeroIllustration.train,
  ),
  _RecommendedGame(
    title: '手指迷宫',
    rating: 4,
    tags: <_Tag>[
      _fineMotorTag,
      _Tag(label: '手眼协调', color: _TrainingColors.orange)
    ],
    illustration: _HeroIllustration.maze,
  ),
  _RecommendedGame(
    title: '节奏敲击',
    rating: 4,
    tags: <_Tag>[_sensoryTag, _Tag(label: '听觉节奏', color: _TrainingColors.teal)],
    illustration: _HeroIllustration.rhythm,
  ),
];

const List<_CollectionGame> _collectionGames = <_CollectionGame>[
  _CollectionGame(
    title: '动物找朋友',
    tag: _cognitionTag,
    performance: '优秀',
    performanceColor: _TrainingColors.green,
    progress: .8,
    status: _GameStatus.done,
    color: _TrainingColors.green,
    illustration: _MiniIllustration.animals,
  ),
  _CollectionGame(
    title: '声音辨认',
    tag: _languageTag,
    performance: '良好',
    performanceColor: _TrainingColors.orange,
    progress: .6,
    status: _GameStatus.playing,
    color: _TrainingColors.green,
    illustration: _MiniIllustration.ear,
  ),
  _CollectionGame(
    title: '图形拼拼看',
    tag: _cognitionTag,
    performance: '良好',
    performanceColor: _TrainingColors.orange,
    progress: .7,
    status: _GameStatus.playing,
    color: _TrainingColors.blue,
    illustration: _MiniIllustration.shapes,
  ),
  _CollectionGame(
    title: '表情猜猜',
    tag: _emotionTag,
    performance: '一般',
    performanceColor: _TrainingColors.orange,
    progress: .4,
    status: _GameStatus.practice,
    color: _TrainingColors.red,
    illustration: _MiniIllustration.emotion,
  ),
  _CollectionGame(
    title: '节奏敲击',
    tag: _sensoryTag,
    performance: '优秀',
    performanceColor: _TrainingColors.green,
    progress: .9,
    status: _GameStatus.done,
    color: _TrainingColors.teal,
    illustration: _MiniIllustration.rhythm,
  ),
  _CollectionGame(
    title: '故事排序',
    tag: _languageTag,
    performance: '良好',
    performanceColor: _TrainingColors.orange,
    progress: .65,
    status: _GameStatus.playing,
    color: _TrainingColors.green,
    illustration: _MiniIllustration.story,
  ),
  _CollectionGame(
    title: '数字跳格',
    tag: _cognitionTag,
    performance: '良好',
    performanceColor: _TrainingColors.orange,
    progress: .55,
    status: _GameStatus.practice,
    color: _TrainingColors.blue,
    illustration: _MiniIllustration.numbers,
  ),
  _CollectionGame(
    title: '记忆翻牌',
    tag: _cognitionTag,
    performance: '一般',
    performanceColor: _TrainingColors.orange,
    progress: .3,
    status: _GameStatus.practice,
    color: _TrainingColors.blue,
    illustration: _MiniIllustration.cards,
  ),
];

enum _TaskState { done, continueTask, start }

class _TrainingTask {
  const _TrainingTask({
    required this.title,
    required this.minutes,
    required this.button,
    required this.state,
  });

  final String title;
  final String minutes;
  final String button;
  final _TaskState state;
}

const List<_TrainingTask> _todayTasks = <_TrainingTask>[
  _TrainingTask(
    title: '颜色配对乐园',
    minutes: '5-8分钟',
    button: '复习',
    state: _TaskState.done,
  ),
  _TrainingTask(
    title: '词语小火车',
    minutes: '8-10分钟',
    button: '复习',
    state: _TaskState.done,
  ),
  _TrainingTask(
    title: '手指迷宫',
    minutes: '5-7分钟',
    button: '继续',
    state: _TaskState.continueTask,
  ),
  _TrainingTask(
    title: '节奏敲击',
    minutes: '6-8分钟',
    button: '开始',
    state: _TaskState.start,
  ),
];

class _RecentRecord {
  const _RecentRecord({
    required this.title,
    required this.time,
    required this.accuracy,
    required this.icon,
    required this.color,
  });

  final String title;
  final String time;
  final int accuracy;
  final IconData icon;
  final Color color;
}

const List<_RecentRecord> _recentRecords = <_RecentRecord>[
  _RecentRecord(
    title: '手指迷宫',
    time: '今天  10:30',
    accuracy: 85,
    icon: Icons.accessibility_new_rounded,
    color: _TrainingColors.green,
  ),
  _RecentRecord(
    title: '动物找朋友',
    time: '今天  09:15',
    accuracy: 90,
    icon: Icons.pan_tool_rounded,
    color: _TrainingColors.orange,
  ),
  _RecentRecord(
    title: '词语小火车',
    time: '昨天  16:40',
    accuracy: 78,
    icon: Icons.directions_train_rounded,
    color: _TrainingColors.blue,
  ),
];

class _PageGlowPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint top = Paint()
      ..shader = LinearGradient(
        begin: Alignment.topCenter,
        end: Alignment.bottomCenter,
        colors: <Color>[
          const Color(0xFFFFF5EA).withOpacity(.32),
          Colors.transparent,
        ],
      ).createShader(Rect.fromLTWH(0, 0, size.width, 210));
    canvas.drawRect(Rect.fromLTWH(0, 72, size.width, 210), top);

    final Paint bottom = Paint()
      ..shader = RadialGradient(
        colors: <Color>[
          const Color(0xFFFFEDDA).withOpacity(.24),
          Colors.transparent,
        ],
      ).createShader(
        Rect.fromCircle(
          center: Offset(size.width * .5, size.height + 40),
          radius: size.width * .6,
        ),
      );
    canvas.drawCircle(
        Offset(size.width * .5, size.height + 40), size.width * .6, bottom);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _DonutProgressPainter extends CustomPainter {
  const _DonutProgressPainter({required this.progress});

  final double progress;

  @override
  void paint(Canvas canvas, Size size) {
    final Offset center = Offset(size.width / 2, size.height / 2);
    final double radius = math.min(size.width, size.height) / 2 - 9;
    final Rect rect = Rect.fromCircle(center: center, radius: radius);
    final Paint track = Paint()
      ..color = const Color(0xFFE8E9EC)
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round
      ..strokeWidth = 12;
    final Paint active = Paint()
      ..shader = const SweepGradient(
        startAngle: -math.pi / 2,
        endAngle: math.pi * 1.5,
        colors: <Color>[Color(0xFFFFA800), _TrainingColors.orange],
      ).createShader(rect)
      ..style = PaintingStyle.stroke
      ..strokeCap = StrokeCap.round
      ..strokeWidth = 12;

    canvas.drawArc(rect, 0, math.pi * 2, false, track);
    canvas.drawArc(rect, -math.pi / 2, math.pi * 2 * progress, false, active);
  }

  @override
  bool shouldRepaint(covariant _DonutProgressPainter oldDelegate) {
    return oldDelegate.progress != progress;
  }
}

class _HeroIllustrationPainter extends CustomPainter {
  const _HeroIllustrationPainter(this.kind);

  final _HeroIllustration kind;

  @override
  void paint(Canvas canvas, Size size) {
    final Rect rect = Offset.zero & size;
    final Paint bg = Paint()
      ..shader = LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: _backgroundColors(),
      ).createShader(rect);
    canvas.drawRect(rect, bg);

    switch (kind) {
      case _HeroIllustration.rainbow:
        _drawRainbow(canvas, size);
      case _HeroIllustration.train:
        _drawTrain(canvas, size);
      case _HeroIllustration.maze:
        _drawMaze(canvas, size);
      case _HeroIllustration.rhythm:
        _drawRhythm(canvas, size);
    }
  }

  List<Color> _backgroundColors() {
    switch (kind) {
      case _HeroIllustration.rainbow:
        return const <Color>[Color(0xFFC6E9FF), Color(0xFFEDF7C7)];
      case _HeroIllustration.train:
        return const <Color>[Color(0xFFE9D9FF), Color(0xFFFFF4DF)];
      case _HeroIllustration.maze:
        return const <Color>[Color(0xFFFFEEBD), Color(0xFFFFF8DD)];
      case _HeroIllustration.rhythm:
        return const <Color>[Color(0xFFE7F8F3), Color(0xFFFFF2DF)];
    }
  }

  void _drawRainbow(Canvas canvas, Size size) {
    final Paint cloud = Paint()..color = Colors.white.withOpacity(.88);
    canvas.drawCircle(Offset(size.width * .12, size.height * .18), 9, cloud);
    canvas.drawCircle(Offset(size.width * .18, size.height * .15), 12, cloud);
    canvas.drawCircle(Offset(size.width * .25, size.height * .2), 8, cloud);
    canvas.drawCircle(Offset(size.width * .78, size.height * .15), 5, cloud);
    canvas.drawCircle(Offset(size.width * .84, size.height * .18), 4, cloud);

    final Offset center = Offset(size.width * .24, size.height * .87);
    const List<Color> colors = <Color>[
      Color(0xFFFF5B5B),
      Color(0xFFFFAF31),
      Color(0xFFFFEB5D),
      Color(0xFF44C565),
      Color(0xFF39A9F7),
      Color(0xFF8069E8),
    ];
    for (int i = 0; i < colors.length; i++) {
      final Paint paint = Paint()
        ..style = PaintingStyle.stroke
        ..strokeWidth = 8
        ..strokeCap = StrokeCap.round
        ..color = colors[i];
      canvas.drawArc(
        Rect.fromCircle(center: center, radius: 76 - i * 10),
        math.pi,
        math.pi,
        false,
        paint,
      );
    }

    final Paint hill = Paint()..color = const Color(0xFFA6D66B);
    final Path path = Path()
      ..moveTo(0, size.height)
      ..quadraticBezierTo(
          size.width * .35, size.height * .63, size.width, size.height * .82)
      ..lineTo(size.width, size.height)
      ..close();
    canvas.drawPath(path, hill);

    _drawShape(canvas, Offset(size.width * .65, size.height * .55),
        _TrainingColors.red, 0);
    _drawShape(canvas, Offset(size.width * .82, size.height * .33),
        _TrainingColors.green, 1);
    _drawShape(canvas, Offset(size.width * .54, size.height * .33),
        _TrainingColors.teal, 2);
    _drawStar(canvas, Offset(size.width * .92, size.height * .22), 12,
        _TrainingColors.yellow);
    canvas.drawCircle(
      Offset(size.width * .83, size.height * .72),
      19,
      Paint()..color = const Color(0xFFFF922D),
    );
    canvas.drawCircle(
      Offset(size.width * .83, size.height * .72),
      9,
      Paint()..color = const Color(0xFFFFC468),
    );
  }

  void _drawTrain(Canvas canvas, Size size) {
    final Paint cloud = Paint()..color = Colors.white.withOpacity(.86);
    canvas.drawCircle(Offset(size.width * .12, size.height * .25), 9, cloud);
    canvas.drawCircle(Offset(size.width * .18, size.height * .22), 13, cloud);
    canvas.drawCircle(Offset(size.width * .24, size.height * .26), 8, cloud);
    canvas.drawCircle(Offset(size.width * .75, size.height * .18), 11, cloud);
    canvas.drawCircle(Offset(size.width * .82, size.height * .16), 13, cloud);
    canvas.drawCircle(Offset(size.width * .9, size.height * .2), 9, cloud);

    final Paint ground = Paint()..color = const Color(0xFF9DCA83);
    final Path groundPath = Path()
      ..moveTo(0, size.height)
      ..lineTo(0, size.height * .86)
      ..quadraticBezierTo(
          size.width * .42, size.height * .62, size.width, size.height * .8)
      ..lineTo(size.width, size.height)
      ..close();
    canvas.drawPath(groundPath, ground);

    final double y = size.height * .56;
    final Paint red = Paint()..color = const Color(0xFFE55835);
    final Paint orange = Paint()..color = const Color(0xFFFF7A24);
    final Paint blue = Paint()..color = const Color(0xFF3385CC);
    final RRect engine = RRect.fromRectAndRadius(
      Rect.fromLTWH(size.width * .2, y - 18, 92, 42),
      const Radius.circular(7),
    );
    canvas.drawRRect(engine, red);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(size.width * .2 + 54, y - 8, 26, 20),
        const Radius.circular(3),
      ),
      Paint()..color = const Color(0xFF9EE2FF),
    );
    canvas.drawRect(Rect.fromLTWH(size.width * .2 + 18, y - 43, 15, 26), red);
    canvas.drawCircle(Offset(size.width * .2 + 25, y - 44), 9, red);
    canvas.drawRect(Rect.fromLTWH(size.width * .2 + 92, y - 4, 24, 6), red);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(size.width * .2 + 116, y - 10, 72, 34),
        const Radius.circular(7),
      ),
      blue,
    );
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(size.width * .2 + 132, y - 1, 28, 14),
        const Radius.circular(3),
      ),
      Paint()..color = const Color(0xFFC9E9FF),
    );
    for (final double dx in <double>[
      size.width * .2 + 22,
      size.width * .2 + 68,
      size.width * .2 + 136,
      size.width * .2 + 174
    ]) {
      canvas.drawCircle(
          Offset(dx, y + 26), 11, Paint()..color = const Color(0xFF4C4543));
      canvas.drawCircle(Offset(dx, y + 26), 5, orange);
    }
    _drawShape(canvas, Offset(size.width * .78, size.height * .42),
        _TrainingColors.green, 0);
    _drawShape(canvas, Offset(size.width * .88, size.height * .55),
        _TrainingColors.orange, 2);
  }

  void _drawMaze(Canvas canvas, Size size) {
    final Paint track = Paint()
      ..color = const Color(0xFFD9AF61)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 4
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;
    final Rect maze = Rect.fromLTWH(size.width * .1, 16, size.width * .76, 64);
    canvas.drawRRect(
      RRect.fromRectAndRadius(maze, const Radius.circular(12)),
      Paint()
        ..color = const Color(0xFFFFF4D3)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 5,
    );
    final Path path = Path()
      ..moveTo(maze.left + 16, maze.top + 14)
      ..lineTo(maze.left + 90, maze.top + 14)
      ..lineTo(maze.left + 90, maze.top + 46)
      ..lineTo(maze.left + 42, maze.top + 46)
      ..lineTo(maze.left + 42, maze.top + 30)
      ..lineTo(maze.left + 142, maze.top + 30)
      ..lineTo(maze.left + 142, maze.top + 12)
      ..lineTo(maze.left + 202, maze.top + 12)
      ..lineTo(maze.left + 202, maze.top + 50)
      ..lineTo(maze.right - 20, maze.top + 50);
    canvas.drawPath(path, track);
    final Paint green = Paint()
      ..color = const Color(0xFF8BC66D)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 5
      ..strokeCap = StrokeCap.round;
    final Path route = Path()
      ..moveTo(maze.left + 104, maze.top + 4)
      ..cubicTo(maze.left + 102, maze.top + 28, maze.left + 128, maze.top + 18,
          maze.left + 130, maze.top + 38)
      ..cubicTo(maze.left + 136, maze.top + 62, maze.left + 170, maze.top + 50,
          maze.left + 164, maze.top + 72);
    canvas.drawPath(route, green);
    canvas.drawCircle(Offset(maze.left + 104, maze.top + 4), 5,
        Paint()..color = const Color(0xFF8BC66D));

    final Paint skin = Paint()..color = const Color(0xFFFFB184);
    final Path hand = Path()
      ..moveTo(size.width * .48, size.height)
      ..lineTo(size.width * .48, size.height * .78)
      ..quadraticBezierTo(size.width * .49, size.height * .64, size.width * .55,
          size.height * .67)
      ..lineTo(size.width * .66, size.height * .98)
      ..close();
    canvas.drawPath(hand, skin);
    canvas.drawCircle(Offset(size.width * .51, size.height * .65), 9, skin);
    canvas.drawCircle(Offset(size.width * .51, size.height * .65), 5,
        Paint()..color = const Color(0xFFFFD3B9));
  }

  void _drawRhythm(Canvas canvas, Size size) {
    final Paint ground = Paint()..color = const Color(0xFFB7DA91);
    final Path groundPath = Path()
      ..moveTo(0, size.height)
      ..lineTo(0, size.height * .84)
      ..quadraticBezierTo(
          size.width * .42, size.height * .64, size.width, size.height * .82)
      ..lineTo(size.width, size.height)
      ..close();
    canvas.drawPath(groundPath, ground);

    final List<Color> colors = <Color>[
      const Color(0xFFFF5B5B),
      const Color(0xFFFFB000),
      const Color(0xFF2E79F6),
      const Color(0xFF20A856),
      const Color(0xFF7B55E6),
    ];
    for (int i = 0; i < colors.length; i++) {
      final double x = size.width * .16 + i * size.width * .1;
      final double h = 28 + (i % 3) * 10;
      canvas.drawRRect(
        RRect.fromRectAndRadius(
          Rect.fromLTWH(x, size.height * .6 - h / 2, 17, h),
          const Radius.circular(9),
        ),
        Paint()..color = colors[i],
      );
    }

    final Paint stick = Paint()
      ..color = const Color(0xFFE07C38)
      ..strokeWidth = 5
      ..strokeCap = StrokeCap.round;
    canvas.drawLine(Offset(size.width * .62, size.height * .62),
        Offset(size.width * .76, size.height * .38), stick);
    canvas.drawLine(Offset(size.width * .70, size.height * .70),
        Offset(size.width * .87, size.height * .48), stick);
    canvas.drawCircle(Offset(size.width * .76, size.height * .38), 9,
        Paint()..color = const Color(0xFFFFC15C));
    canvas.drawCircle(Offset(size.width * .87, size.height * .48), 9,
        Paint()..color = const Color(0xFFFFC15C));

    final TextPainter notePainter = TextPainter(
      text: const TextSpan(
        text: '♪',
        style: TextStyle(
          color: _TrainingColors.teal,
          fontSize: 30,
          fontWeight: FontWeight.w900,
        ),
      ),
      textDirection: TextDirection.ltr,
    )..layout();
    notePainter.paint(canvas, Offset(size.width * .78, size.height * .12));
  }

  void _drawShape(Canvas canvas, Offset center, Color color, int kind) {
    final Paint paint = Paint()..color = color.withOpacity(.88);
    if (kind == 0) {
      final Path p = Path()
        ..moveTo(center.dx, center.dy - 15)
        ..lineTo(center.dx + 18, center.dy + 14)
        ..lineTo(center.dx - 18, center.dy + 14)
        ..close();
      canvas.drawPath(p, paint);
    } else if (kind == 1) {
      final Path p = Path()
        ..moveTo(center.dx, center.dy - 18)
        ..lineTo(center.dx + 24, center.dy + 16)
        ..lineTo(center.dx - 24, center.dy + 16)
        ..close();
      canvas.drawPath(p, paint);
    } else {
      canvas.drawPath(
        Path()
          ..moveTo(center.dx, center.dy - 17)
          ..lineTo(center.dx + 14, center.dy + 9)
          ..lineTo(center.dx - 14, center.dy + 9)
          ..close(),
        Paint()
          ..color = Colors.transparent
          ..style = PaintingStyle.fill,
      );
      canvas.drawPath(
        Path()
          ..moveTo(center.dx, center.dy - 17)
          ..lineTo(center.dx + 14, center.dy + 9)
          ..lineTo(center.dx - 14, center.dy + 9)
          ..close(),
        Paint()
          ..color = color.withOpacity(.22)
          ..style = PaintingStyle.fill,
      );
      canvas.drawPath(
        Path()
          ..moveTo(center.dx, center.dy - 17)
          ..lineTo(center.dx + 14, center.dy + 9)
          ..lineTo(center.dx - 14, center.dy + 9)
          ..close(),
        Paint()
          ..color = color
          ..style = PaintingStyle.stroke
          ..strokeWidth = 4
          ..strokeJoin = StrokeJoin.round,
      );
    }
  }

  void _drawStar(Canvas canvas, Offset center, double radius, Color color) {
    final Path path = Path();
    for (int i = 0; i < 10; i++) {
      final double angle = -math.pi / 2 + i * math.pi / 5;
      final double r = i.isEven ? radius : radius * .48;
      final Offset point = Offset(
          center.dx + math.cos(angle) * r, center.dy + math.sin(angle) * r);
      if (i == 0) {
        path.moveTo(point.dx, point.dy);
      } else {
        path.lineTo(point.dx, point.dy);
      }
    }
    path.close();
    canvas.drawPath(path, Paint()..color = color);
  }

  @override
  bool shouldRepaint(covariant _HeroIllustrationPainter oldDelegate) {
    return oldDelegate.kind != kind;
  }
}

class _MiniIllustrationPainter extends CustomPainter {
  const _MiniIllustrationPainter(this.kind, this.color);

  final _MiniIllustration kind;
  final Color color;

  @override
  void paint(Canvas canvas, Size size) {
    const double baseSide = 58;
    const Size drawSize = Size(baseSide, baseSide);
    final double side = math.min(size.width, size.height);
    final double scale = side / baseSide;

    canvas.save();
    canvas.translate((size.width - side) / 2, (size.height - side) / 2);
    canvas.scale(scale);

    final Paint paint = Paint()..color = color;
    switch (kind) {
      case _MiniIllustration.animals:
        _drawFace(canvas, Offset(drawSize.width * .48, drawSize.height * .41),
            18, const Color(0xFFE98D28), true);
        _drawFace(canvas, Offset(drawSize.width * .32, drawSize.height * .73),
            12, Colors.white, false);
        _drawFace(canvas, Offset(drawSize.width * .7, drawSize.height * .73),
            12, Colors.white, false);
      case _MiniIllustration.ear:
        final Paint stroke = Paint()
          ..color = color
          ..style = PaintingStyle.stroke
          ..strokeWidth = 4
          ..strokeCap = StrokeCap.round;
        canvas.drawArc(Rect.fromLTWH(26, 12, 22, 32), -math.pi * .55,
            math.pi * 1.25, false, stroke);
        canvas.drawArc(Rect.fromLTWH(32, 20, 12, 18), -math.pi * .55, math.pi,
            false, stroke);
        canvas.drawArc(Rect.fromLTWH(16, 17, 10, 20), -math.pi / 2, math.pi,
            false, stroke..strokeWidth = 3);
        canvas.drawArc(Rect.fromLTWH(10, 12, 14, 30), -math.pi / 2, math.pi,
            false, stroke);
      case _MiniIllustration.shapes:
        canvas.drawCircle(
            Offset(drawSize.width * .35, drawSize.height * .34), 12, paint);
        final Path triangle = Path()
          ..moveTo(drawSize.width * .68, drawSize.height * .15)
          ..lineTo(drawSize.width * .84, drawSize.height * .5)
          ..lineTo(drawSize.width * .52, drawSize.height * .5)
          ..close();
        canvas.drawPath(triangle, Paint()..color = const Color(0xFF65BE62));
        canvas.drawRRect(
          RRect.fromRectAndRadius(
              Rect.fromLTWH(11, 34, 20, 18), const Radius.circular(3)),
          Paint()..color = const Color(0xFF4FA4E7),
        );
        canvas.drawRRect(
          RRect.fromRectAndRadius(
              Rect.fromLTWH(38, 34, 18, 18), const Radius.circular(3)),
          Paint()
            ..color = Colors.white.withOpacity(.45)
            ..style = PaintingStyle.stroke
            ..strokeWidth = 2,
        );
      case _MiniIllustration.emotion:
        _drawPerson(canvas, Offset(drawSize.width * .35, drawSize.height * .46),
            const Color(0xFFFFB795));
        _drawPerson(canvas, Offset(drawSize.width * .65, drawSize.height * .46),
            const Color(0xFFFFC6A1));
      case _MiniIllustration.rhythm:
        for (int i = 0; i < 4; i++) {
          final double x = 9 + i * 11;
          canvas.drawRRect(
            RRect.fromRectAndRadius(Rect.fromLTWH(x, 14 + i % 2 * 5, 8, 30),
                const Radius.circular(4)),
            Paint()
              ..color = <Color>[
                const Color(0xFFEF4D4D),
                const Color(0xFFFFB000),
                const Color(0xFF2E79F6),
                const Color(0xFF20A856),
              ][i],
          );
        }
        canvas.drawCircle(
            Offset(43, 42), 5, Paint()..color = const Color(0xFFFFB000));
        canvas.drawLine(
            Offset(45, 38),
            Offset(52, 24),
            Paint()
              ..color = color
              ..strokeWidth = 3
              ..strokeCap = StrokeCap.round);
      case _MiniIllustration.story:
        canvas.drawRRect(
          RRect.fromRectAndRadius(
              Rect.fromLTWH(9, 12, 40, 34), const Radius.circular(5)),
          Paint()..color = Colors.white.withOpacity(.72),
        );
        canvas.drawLine(
            Offset(29, 14),
            Offset(29, 46),
            Paint()
              ..color = color.withOpacity(.35)
              ..strokeWidth = 2);
        canvas.drawCircle(
            Offset(21, 26), 6, Paint()..color = const Color(0xFFFFB000));
        canvas.drawRRect(
            RRect.fromRectAndRadius(
                Rect.fromLTWH(34, 24, 8, 12), const Radius.circular(3)),
            Paint()..color = const Color(0xFFFF7A24));
        canvas.drawPath(
          Path()
            ..moveTo(14, 38)
            ..quadraticBezierTo(21, 32, 28, 38),
          Paint()
            ..color = color
            ..style = PaintingStyle.stroke
            ..strokeWidth = 2
            ..strokeCap = StrokeCap.round,
        );
      case _MiniIllustration.numbers:
        _numberTile(canvas, const Rect.fromLTWH(8, 10, 21, 31), '1', 0);
        _numberTile(canvas, const Rect.fromLTWH(30, 9, 22, 32), '2', -.08);
        _numberTile(canvas, const Rect.fromLTWH(23, 35, 22, 18), '3', .12);
      case _MiniIllustration.cards:
        _card(canvas, const Rect.fromLTWH(12, 12, 22, 34), -.14,
            const Color(0xFF2E79F6), Icons.star_rounded);
        _card(canvas, const Rect.fromLTWH(29, 15, 22, 34), .18,
            const Color(0xFF8FA6C8), Icons.pets_rounded);
    }
    canvas.restore();
  }

  void _drawFace(
      Canvas canvas, Offset center, double radius, Color fill, bool mane) {
    if (mane) {
      canvas.drawCircle(
          center, radius + 5, Paint()..color = const Color(0xFFC86C18));
    }
    canvas.drawCircle(center, radius, Paint()..color = fill);
    canvas.drawCircle(center + Offset(-radius * .35, -radius * .12),
        radius * .11, Paint()..color = Colors.black87);
    canvas.drawCircle(center + Offset(radius * .35, -radius * .12),
        radius * .11, Paint()..color = Colors.black87);
    canvas.drawCircle(center + Offset(0, radius * .22), radius * .14,
        Paint()..color = Colors.black87);
  }

  void _drawPerson(Canvas canvas, Offset center, Color skin) {
    canvas.drawCircle(center, 12, Paint()..color = skin);
    canvas.drawArc(
        Rect.fromCircle(center: center + const Offset(0, -4), radius: 12),
        math.pi,
        math.pi,
        false,
        Paint()
          ..color = const Color(0xFF5B392A)
          ..strokeWidth = 5
          ..style = PaintingStyle.stroke);
    canvas.drawCircle(
        center + const Offset(-4, -1), 1.5, Paint()..color = Colors.black87);
    canvas.drawCircle(
        center + const Offset(4, -1), 1.5, Paint()..color = Colors.black87);
    canvas.drawArc(
        Rect.fromCenter(
            center: center + const Offset(0, 4), width: 8, height: 5),
        0,
        math.pi,
        false,
        Paint()
          ..color = Colors.black87
          ..style = PaintingStyle.stroke
          ..strokeWidth = 1.4);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
          Rect.fromCenter(
              center: center + const Offset(0, 23), width: 25, height: 18),
          const Radius.circular(8)),
      Paint()..color = color.withOpacity(.75),
    );
  }

  void _numberTile(Canvas canvas, Rect rect, String text, double rotate) {
    canvas.save();
    canvas.translate(rect.center.dx, rect.center.dy);
    canvas.rotate(rotate);
    final Rect local = Rect.fromCenter(
        center: Offset.zero, width: rect.width, height: rect.height);
    canvas.drawRRect(RRect.fromRectAndRadius(local, const Radius.circular(4)),
        Paint()..color = Colors.white.withOpacity(.78));
    canvas.drawRRect(
        RRect.fromRectAndRadius(local, const Radius.circular(4)),
        Paint()
          ..color = color
          ..style = PaintingStyle.stroke
          ..strokeWidth = 2);
    final TextPainter painter = TextPainter(
      text: TextSpan(
        text: text,
        style: const TextStyle(
            color: _TrainingColors.ink,
            fontSize: 20,
            fontWeight: FontWeight.w900),
      ),
      textDirection: TextDirection.ltr,
    )..layout();
    painter.paint(canvas, Offset(-painter.width / 2, -painter.height / 2));
    canvas.restore();
  }

  void _card(
      Canvas canvas, Rect rect, double rotate, Color fill, IconData icon) {
    canvas.save();
    canvas.translate(rect.center.dx, rect.center.dy);
    canvas.rotate(rotate);
    final Rect local = Rect.fromCenter(
        center: Offset.zero, width: rect.width, height: rect.height);
    canvas.drawRRect(RRect.fromRectAndRadius(local, const Radius.circular(4)),
        Paint()..color = fill);
    canvas.drawRRect(
        RRect.fromRectAndRadius(local.deflate(3), const Radius.circular(3)),
        Paint()
          ..color = Colors.white.withOpacity(.18)
          ..style = PaintingStyle.stroke
          ..strokeWidth = 1.6);
    final TextPainter painter = TextPainter(
      text: TextSpan(
        text: icon == Icons.star_rounded ? '★' : '♪',
        style: const TextStyle(
            color: Color(0xFFFFD35C),
            fontSize: 17,
            fontWeight: FontWeight.w900),
      ),
      textDirection: TextDirection.ltr,
    )..layout();
    painter.paint(canvas, Offset(-painter.width / 2, -painter.height / 2));
    canvas.restore();
  }

  @override
  bool shouldRepaint(covariant _MiniIllustrationPainter oldDelegate) {
    return oldDelegate.kind != kind || oldDelegate.color != color;
  }
}
