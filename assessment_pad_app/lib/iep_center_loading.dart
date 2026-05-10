part of 'iep_center_page.dart';

class _QueueListSkeleton extends StatelessWidget {
  const _QueueListSkeleton();

  @override
  Widget build(BuildContext context) {
    return ListView.separated(
      padding: EdgeInsets.zero,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: 6,
      separatorBuilder: (BuildContext context, int index) {
        return const SizedBox(height: 10);
      },
      itemBuilder: (BuildContext context, int index) {
        return _QueueStudentSkeletonCard(
          key: ValueKey<String>('iep-queue-skeleton-$index'),
          active: index == 0,
        );
      },
    );
  }
}

class _QueueStudentSkeletonCard extends StatelessWidget {
  const _QueueStudentSkeletonCard({
    required this.active,
    super.key,
  });

  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 80,
      padding: const EdgeInsets.symmetric(horizontal: 11, vertical: 8),
      decoration: BoxDecoration(
        color: active ? const Color(0xFFFFF3EB) : Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: active ? const Color(0xFFFFD0B7) : _IepColors.lightLine,
          width: active ? 1.2 : 1,
        ),
      ),
      child: Row(
        children: <Widget>[
          const _IepSkeletonBlock(width: 44, height: 44, radius: 99),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: const <Widget>[
                Row(
                  children: <Widget>[
                    Expanded(
                      child: _IepSkeletonBlock(
                        widthFactor: .62,
                        height: 13,
                      ),
                    ),
                    SizedBox(width: 8),
                    _IepSkeletonBlock(width: 44, height: 20, radius: 10),
                  ],
                ),
                SizedBox(height: 9),
                _IepSkeletonBlock(widthFactor: .72, height: 11),
                SizedBox(height: 9),
                _IepSkeletonBlock(widthFactor: .58, height: 11),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _IepWordTableSkeleton extends StatelessWidget {
  const _IepWordTableSkeleton();

  @override
  Widget build(BuildContext context) {
    return _WordTableFrame(
      height: 820,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: const <Widget>[
          _WordSkeletonTitle(),
          _WordSkeletonRow(
              height: 42, widths: <double>[.12, .2, .1, .16, .12, .24]),
          _WordSkeletonRow(height: 42, widths: <double>[.12, .22, .12, .28]),
          _WordSkeletonRow(height: 42, widths: <double>[.12, .18, .14, .3]),
          _WordSkeletonHeaderRow(),
          _WordSkeletonDomainBlock(active: true),
          _WordSkeletonDomainBlock(),
          _WordSkeletonDomainBlock(),
          _WordSkeletonDomainBlock(shortRows: 2),
        ],
      ),
    );
  }
}

class _WordSkeletonTitle extends StatelessWidget {
  const _WordSkeletonTitle();

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
      child: const _IepSkeletonBlock(width: 154, height: 16, radius: 6),
    );
  }
}

class _WordSkeletonHeaderRow extends StatelessWidget {
  const _WordSkeletonHeaderRow();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 42,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(
          bottom: BorderSide(color: Color(0xFFB98A71), width: .8),
        ),
      ),
      child: Row(
        children: const <Widget>[
          Expanded(
              flex: 12,
              child: Center(child: _IepSkeletonBlock(width: 38, height: 12))),
          _WordSkeletonDivider(),
          Expanded(
              flex: 30,
              child: Center(child: _IepSkeletonBlock(width: 88, height: 12))),
          _WordSkeletonDivider(),
          Expanded(
              flex: 22,
              child: Center(child: _IepSkeletonBlock(width: 70, height: 12))),
          _WordSkeletonDivider(),
          Expanded(
              flex: 13,
              child: Center(child: _IepSkeletonBlock(width: 42, height: 12))),
          _WordSkeletonDivider(),
          Expanded(
              flex: 23,
              child: Center(child: _IepSkeletonBlock(width: 58, height: 12))),
        ],
      ),
    );
  }
}

class _WordSkeletonRow extends StatelessWidget {
  const _WordSkeletonRow({
    required this.height,
    required this.widths,
  });

  final double height;
  final List<double> widths;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: height,
      decoration: const BoxDecoration(
        color: Colors.white,
        border: Border(
          bottom: BorderSide(color: Color(0xFFB98A71), width: .8),
        ),
      ),
      child: Row(
        children: <Widget>[
          for (int index = 0; index < widths.length; index += 1) ...<Widget>[
            Expanded(
              flex: (widths[index] * 100).round(),
              child: Center(
                child: _IepSkeletonBlock(
                  widthFactor: index.isEven ? .46 : .68,
                  height: 12,
                  radius: 5,
                ),
              ),
            ),
            if (index != widths.length - 1) const _WordSkeletonDivider(),
          ],
        ],
      ),
    );
  }
}

class _WordSkeletonDomainBlock extends StatelessWidget {
  const _WordSkeletonDomainBlock({
    this.active = false,
    this.shortRows = 3,
  });

  final bool active;
  final int shortRows;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: shortRows * 39.0 > 122.4 ? shortRows * 39.0 : 122.4,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          const Expanded(
            flex: 12,
            child: Center(child: _IepSkeletonBlock(width: 46, height: 14)),
          ),
          const _WordSkeletonDivider(),
          Expanded(
            flex: 30,
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  _IepSkeletonBlock(
                      widthFactor: active ? .92 : .72, height: 12),
                  const SizedBox(height: 9),
                  const _IepSkeletonBlock(widthFactor: .82, height: 12),
                ],
              ),
            ),
          ),
          const _WordSkeletonDivider(),
          Expanded(
            flex: 22,
            child: Column(
              children: List<Widget>.generate(shortRows, (int index) {
                return Expanded(
                  child: Container(
                    alignment: Alignment.centerLeft,
                    padding: const EdgeInsets.symmetric(horizontal: 10),
                    decoration: BoxDecoration(
                      border: Border(
                        bottom: index == shortRows - 1
                            ? BorderSide.none
                            : const BorderSide(
                                color: Color(0xFFB98A71),
                                width: .8,
                              ),
                      ),
                    ),
                    child: _IepSkeletonBlock(
                      widthFactor: index == 0 ? .82 : .64,
                      height: 11,
                    ),
                  ),
                );
              }),
            ),
          ),
          const _WordSkeletonDivider(),
          const Expanded(
            flex: 13,
            child: Center(child: _IepSkeletonBlock(width: 36, height: 12)),
          ),
          const _WordSkeletonDivider(),
          const Expanded(
            flex: 23,
            child: Center(child: _IepSkeletonBlock(width: 96, height: 12)),
          ),
        ],
      ),
    );
  }
}

class _WordSkeletonDivider extends StatelessWidget {
  const _WordSkeletonDivider();

  @override
  Widget build(BuildContext context) {
    return Container(width: .8, color: const Color(0xFFB98A71));
  }
}

class _IepPlanLoadingState extends StatelessWidget {
  const _IepPlanLoadingState();

  @override
  Widget build(BuildContext context) {
    return const Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          _IepHourglassLoader(),
          SizedBox(height: 14),
          Text(
            '正在读取IEP计划',
            style: TextStyle(
              color: _IepColors.text,
              fontSize: 13,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _IepHourglassLoader extends StatefulWidget {
  const _IepHourglassLoader({this.size = 34});

  final double size;

  @override
  State<_IepHourglassLoader> createState() => _IepHourglassLoaderState();
}

class _IepHourglassLoaderState extends State<_IepHourglassLoader>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 980),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (BuildContext context, Widget? child) {
        return Transform.rotate(
          angle: _controller.value * 3.1415926,
          child: child,
        );
      },
      child: SizedBox(
        width: widget.size,
        height: widget.size,
        child: const CustomPaint(painter: _IepHourglassPainter()),
      ),
    );
  }
}

class _IepHourglassPainter extends CustomPainter {
  const _IepHourglassPainter();

  @override
  void paint(Canvas canvas, Size size) {
    final double left = size.width * .25;
    final double right = size.width * .75;
    final double top = size.height * .16;
    final double middle = size.height * .5;
    final double bottom = size.height * .84;
    final double centerX = size.width * .5;

    final Paint framePaint = Paint()
      ..color = _IepColors.orangeDeep
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2.2
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;
    final Paint sandPaint = Paint()
      ..color = _IepColors.orangeDeep.withOpacity(.72)
      ..style = PaintingStyle.fill;

    final Path frame = Path()
      ..moveTo(left, top)
      ..lineTo(right, top)
      ..moveTo(left, bottom)
      ..lineTo(right, bottom)
      ..moveTo(left + 1, top + 1)
      ..quadraticBezierTo(centerX - 5, middle - 2, centerX, middle)
      ..quadraticBezierTo(centerX - 5, middle + 2, left + 1, bottom - 1)
      ..moveTo(right - 1, top + 1)
      ..quadraticBezierTo(centerX + 5, middle - 2, centerX, middle)
      ..quadraticBezierTo(centerX + 5, middle + 2, right - 1, bottom - 1);

    final Path bottomSand = Path()
      ..moveTo(centerX, middle + 1)
      ..lineTo(right - 3, bottom - 2)
      ..lineTo(left + 3, bottom - 2)
      ..close();
    final Path topSand = Path()
      ..moveTo(left + 5, top + 3)
      ..lineTo(right - 5, top + 3)
      ..lineTo(centerX, middle - 2)
      ..close();

    canvas.drawPath(
        topSand, sandPaint..color = sandPaint.color.withOpacity(.2));
    canvas.drawPath(bottomSand, sandPaint..color = _IepColors.orangeDeep);
    canvas.drawCircle(Offset(centerX, middle), 1.4, sandPaint);
    canvas.drawPath(frame, framePaint);
  }

  @override
  bool shouldRepaint(covariant _IepHourglassPainter oldDelegate) => false;
}

class _PlanStateView extends StatelessWidget {
  const _PlanStateView({
    required this.icon,
    required this.title,
    this.message = '',
    this.actionLabel = '',
    this.onAction,
  });

  final IconData icon;
  final String title;
  final String message;
  final String actionLabel;
  final VoidCallback? onAction;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: SizedBox(
        width: 340,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Icon(icon, size: 34, color: _IepColors.orangeDeep),
            const SizedBox(height: 10),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _IepColors.ink,
                fontSize: 15,
                fontWeight: FontWeight.w900,
              ),
            ),
            if (message.trim().isNotEmpty) ...<Widget>[
              const SizedBox(height: 7),
              Text(
                message,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: _IepColors.text,
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                  height: 1.35,
                ),
              ),
            ],
            if (actionLabel.trim().isNotEmpty && onAction != null) ...<Widget>[
              const SizedBox(height: 12),
              _MiniQueueAction(label: actionLabel, onTap: onAction!),
            ],
          ],
        ),
      ),
    );
  }
}

class _IepEmptyGenerateState extends StatelessWidget {
  const _IepEmptyGenerateState({
    required this.studentName,
    required this.generating,
    required this.statusText,
    required this.onGenerate,
  });

  final String studentName;
  final bool generating;
  final String statusText;
  final VoidCallback onGenerate;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: SizedBox(
        width: 460,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            generating
                ? const _IepHourglassLoader()
                : const _IepEmptyIllustration(),
            const SizedBox(height: 18),
            Text(
              generating ? '正在生成 $studentName 的IEP计划' : '$studentName 暂无IEP计划',
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _IepColors.ink,
                fontSize: 18,
                fontWeight: FontWeight.w900,
                height: 1.15,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              generating
                  ? (statusText.trim().isEmpty ? 'AI正在准备表格内容' : statusText)
                  : '可基于当前评估记录生成IEP总计划，生成后会继续展示月计划和周计划入口。',
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _IepColors.text,
                fontSize: 13,
                fontWeight: FontWeight.w700,
                height: 1.45,
              ),
            ),
            const SizedBox(height: 20),
            _AiGenerateButton(
              generating: generating,
              onTap: generating ? null : onGenerate,
            ),
          ],
        ),
      ),
    );
  }
}

class _IepEmptyIllustration extends StatelessWidget {
  const _IepEmptyIllustration();

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 128,
      height: 92,
      child: Stack(
        alignment: Alignment.center,
        children: <Widget>[
          Positioned(
            bottom: 4,
            child: Container(
              width: 106,
              height: 58,
              decoration: BoxDecoration(
                color: const Color(0xFFFFF3EC).withOpacity(.62),
                borderRadius: BorderRadius.circular(29),
              ),
            ),
          ),
          Positioned(
            left: 21,
            top: 18,
            child: Transform.rotate(
              angle: -0.08,
              child: Container(
                width: 70,
                height: 58,
                padding: const EdgeInsets.fromLTRB(10, 10, 10, 8),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(10),
                  border: Border.all(color: const Color(0xFFEFCBB7)),
                  boxShadow: _iepShadow(
                    color: const Color(0x0FB05F32),
                    blur: 10,
                    offset: const Offset(0, 4),
                  ),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: const <Widget>[
                    _EmptyDocLine(width: 42, strong: true),
                    SizedBox(height: 7),
                    _EmptyDocLine(width: 50),
                    SizedBox(height: 6),
                    _EmptyDocLine(width: 35),
                  ],
                ),
              ),
            ),
          ),
          Positioned(
            right: 20,
            top: 10,
            child: Container(
              width: 42,
              height: 42,
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: _IepColors.orange,
                shape: BoxShape.circle,
                boxShadow: _iepShadow(
                  color: const Color(0x30E96F43),
                  blur: 14,
                  offset: const Offset(0, 6),
                ),
              ),
              child: const Icon(
                Icons.auto_awesome_rounded,
                size: 23,
                color: Colors.white,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _EmptyDocLine extends StatelessWidget {
  const _EmptyDocLine({required this.width, this.strong = false});

  final double width;
  final bool strong;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: strong ? 5 : 4,
      decoration: BoxDecoration(
        color: strong ? _IepColors.orangeSoft : const Color(0xFFF3DED1),
        borderRadius: BorderRadius.circular(3),
      ),
    );
  }
}

class _AiGenerateButton extends StatelessWidget {
  const _AiGenerateButton({
    required this.generating,
    required this.onTap,
  });

  final bool generating;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: generating ? const Color(0xFFEFC1A8) : _IepColors.orange,
      borderRadius: BorderRadius.circular(18),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(18),
        child: Container(
          height: 42,
          padding: const EdgeInsets.symmetric(horizontal: 26),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(18),
            boxShadow: _iepShadow(
              color: const Color(0x2FE96F43),
              blur: 16,
              offset: const Offset(0, 7),
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              Icon(
                generating
                    ? Icons.hourglass_top_rounded
                    : Icons.auto_awesome_rounded,
                size: 18,
                color: Colors.white,
              ),
              const SizedBox(width: 7),
              Text(
                generating ? '生成中' : 'AI生成',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 14,
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

class _IepGenerationStatusStrip extends StatelessWidget {
  const _IepGenerationStatusStrip({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return IgnorePointer(
      child: Center(
        child: Container(
          height: 34,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: Colors.white.withOpacity(.92),
            borderRadius: BorderRadius.circular(17),
            border: Border.all(color: const Color(0xFFFFD8C3)),
            boxShadow: _iepShadow(
              color: const Color(0x18B05F32),
              blur: 12,
              offset: const Offset(0, 5),
            ),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const _IepHourglassLoader(size: 18),
              const SizedBox(width: 8),
              Text(
                text,
                style: const TextStyle(
                  color: _IepColors.orangeDeep,
                  fontSize: 12,
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

class _IepSkeletonBlock extends StatelessWidget {
  const _IepSkeletonBlock({
    this.width,
    this.widthFactor,
    required this.height,
    this.radius = 99,
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
        color: const Color(0xFFF1E1D6),
        borderRadius: BorderRadius.circular(radius),
      ),
    );
    if (widthFactor == null) {
      return block;
    }
    return FractionallySizedBox(
      widthFactor: widthFactor,
      alignment: Alignment.centerLeft,
      child: block,
    );
  }
}
