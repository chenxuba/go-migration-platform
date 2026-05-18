part of 'pep3_assessment_page.dart';

class _Pep3LoadingShell extends StatelessWidget {
  const _Pep3LoadingShell({
    required this.title,
    required this.studentName,
    required this.age,
    required this.assessmentDate,
    required this.examinerName,
    required this.autoSaveText,
    required this.onBack,
  });

  final String title;
  final String studentName;
  final String age;
  final String assessmentDate;
  final String examinerName;
  final String autoSaveText;
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: _Pep3Colors.page,
      child: Column(
        children: <Widget>[
          _Pep3LoadingHeader(
            title: title,
            studentName: studentName,
            age: age,
            assessmentDate: assessmentDate,
            examinerName: examinerName,
            autoSaveText: autoSaveText,
            onBack: onBack,
          ),
          const Expanded(
            child: Padding(
              padding: EdgeInsets.fromLTRB(10, 10, 10, 10),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  SizedBox(width: 226, child: _Pep3SidebarSkeleton()),
                  SizedBox(width: 10),
                  Expanded(child: _Pep3QuestionSkeleton()),
                  SizedBox(width: 10),
                  SizedBox(width: 238, child: _Pep3RightRailSkeleton()),
                ],
              ),
            ),
          ),
          const _Pep3FooterSkeleton(),
        ],
      ),
    );
  }
}

class _Pep3LoadingHeader extends StatelessWidget {
  const _Pep3LoadingHeader({
    required this.title,
    required this.studentName,
    required this.age,
    required this.assessmentDate,
    required this.examinerName,
    required this.autoSaveText,
    required this.onBack,
  });

  final String title;
  final String studentName;
  final String age;
  final String assessmentDate;
  final String examinerName;
  final String autoSaveText;
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 58,
      padding: const EdgeInsets.symmetric(horizontal: 18),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        border: Border.all(color: _Pep3Colors.line),
        borderRadius: const BorderRadius.vertical(bottom: Radius.circular(12)),
        boxShadow: _pep3Shadow(color: const Color(0x16B05F32), blur: 16),
      ),
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final bool compact = constraints.maxWidth < 1280;
          final List<Widget> headerChildren = <Widget>[
            Text(
              '$title 测评工作台',
              maxLines: 1,
              softWrap: false,
              style: const TextStyle(
                color: _Pep3Colors.ink,
                fontSize: 23,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            _HeaderLoadingMeta(
              label: '儿童',
              value: studentName,
              compact: compact,
            ),
            _HeaderLoadingMeta(label: '年龄', value: age, compact: compact),
            _HeaderLoadingMeta(
              label: compact ? '日期' : '测评日期',
              value: assessmentDate,
              compact: compact,
            ),
            _HeaderLoadingMeta(
              label: '施测者',
              value: examinerName,
              compact: compact,
            ),
          ];
          return Row(
            children: <Widget>[
              _HeaderIconButton(
                icon: Icons.chevron_left_rounded,
                onTap: onBack,
              ),
              const SizedBox(width: 10),
              Expanded(
                child: SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  physics: const ClampingScrollPhysics(),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: headerChildren,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              _SaveStatusLabel(
                text:
                    autoSaveText.trim().isEmpty ? '等待作答' : autoSaveText.trim(),
                saving: false,
              ),
              const SizedBox(width: 12),
              const _SkeletonButton(width: 112),
              const SizedBox(width: 9),
              const _SkeletonButton(width: 112, filled: true),
            ],
          );
        },
      ),
    );
  }
}

class _HeaderLoadingMeta extends StatelessWidget {
  const _HeaderLoadingMeta({
    required this.label,
    required this.value,
    required this.compact,
  });

  final String label;
  final String value;
  final bool compact;

  @override
  Widget build(BuildContext context) {
    final String resolved = value.trim();
    return Container(
      margin: EdgeInsets.only(left: compact ? 6 : 10),
      padding: EdgeInsets.only(left: compact ? 6 : 10),
      decoration: const BoxDecoration(
        border: Border(left: BorderSide(color: _Pep3Colors.line)),
      ),
      child: resolved.isEmpty
          ? Row(
              mainAxisSize: MainAxisSize.min,
              children: <Widget>[
                Text(
                  '$label：',
                  maxLines: 1,
                  style: const TextStyle(
                    color: _Pep3Colors.text,
                    fontSize: 13,
                    height: 1,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(width: 5),
                const _LoadingLine(width: 40, height: 12),
              ],
            )
          : Text.rich(
              TextSpan(
                children: <InlineSpan>[
                  TextSpan(text: '$label：'),
                  TextSpan(
                    text: resolved,
                    style: const TextStyle(fontWeight: FontWeight.w900),
                  ),
                ],
              ),
              maxLines: 1,
              softWrap: false,
              style: const TextStyle(
                color: _Pep3Colors.text,
                fontSize: 13,
                height: 1,
                fontWeight: FontWeight.w700,
              ),
            ),
    );
  }
}

class _Pep3SidebarSkeleton extends StatelessWidget {
  const _Pep3SidebarSkeleton();

  @override
  Widget build(BuildContext context) {
    return _RailCard(
      child: Column(
        children: <Widget>[
          Container(
            height: 48,
            padding: const EdgeInsets.symmetric(horizontal: 14),
            decoration: const BoxDecoration(
              border: Border(bottom: BorderSide(color: _Pep3Colors.lineSoft)),
            ),
            child: const Row(
              children: <Widget>[
                _LoadingLine(width: 88, height: 18),
                Spacer(),
                _SkeletonPill(width: 18, height: 18, radius: 6),
              ],
            ),
          ),
          Expanded(
            child: ListView.separated(
              padding: const EdgeInsets.fromLTRB(12, 13, 10, 13),
              physics: const NeverScrollableScrollPhysics(),
              itemBuilder: (BuildContext context, int index) {
                return _PageGroupSkeleton(expanded: index == 0);
              },
              separatorBuilder: (BuildContext context, int index) {
                return const Divider(height: 18, color: _Pep3Colors.lineSoft);
              },
              itemCount: 7,
            ),
          ),
        ],
      ),
    );
  }
}

class _PageGroupSkeleton extends StatelessWidget {
  const _PageGroupSkeleton({required this.expanded});

  final bool expanded;

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        const Row(
          children: <Widget>[
            _SkeletonPill(width: 20, height: 20, radius: 7),
            SizedBox(width: 8),
            Expanded(child: _LoadingLine(widthFactor: .72, height: 14)),
            SizedBox(width: 10),
            _LoadingLine(width: 36, height: 12),
          ],
        ),
        const SizedBox(height: 9),
        const Row(
          children: <Widget>[
            SizedBox(width: 28),
            Expanded(child: _SkeletonPill(height: 5, radius: 99)),
            SizedBox(width: 9),
            _LoadingLine(width: 28, height: 12),
          ],
        ),
        if (expanded) ...<Widget>[
          const SizedBox(height: 10),
          for (int i = 0; i < 5; i++) ...<Widget>[
            const Padding(
              padding: EdgeInsets.only(left: 18, bottom: 8),
              child: Row(
                children: <Widget>[
                  _LoadingLine(width: 45, height: 12),
                  SizedBox(width: 8),
                  Expanded(child: _LoadingLine(widthFactor: .72, height: 12)),
                  SizedBox(width: 8),
                  _SkeletonPill(width: 15, height: 15, radius: 99),
                ],
              ),
            ),
          ],
        ],
      ],
    );
  }
}

class _Pep3QuestionSkeleton extends StatelessWidget {
  const _Pep3QuestionSkeleton();

  @override
  Widget build(BuildContext context) {
    return _RailCard(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 18, 16, 14),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            const Row(
              children: <Widget>[
                Expanded(child: _LoadingLine(widthFactor: .72, height: 28)),
                SizedBox(width: 12),
                _SkeletonPill(width: 154, height: 34, radius: 10),
              ],
            ),
            const SizedBox(height: 16),
            Expanded(
              child: ListView(
                padding: EdgeInsets.zero,
                physics: const NeverScrollableScrollPhysics(),
                children: const <Widget>[
                  _QuestionLoadingCard(title: '材料', height: 72),
                  _QuestionLoadingCard(title: '操作标准', height: 86),
                  _QuestionLoadingCard(title: '指导语', height: 86),
                  _QuestionLoadingCard(title: '评分标准', height: 86),
                  SizedBox(height: 2),
                ],
              ),
            ),
            const SizedBox(height: 10),
            const Row(
              children: <Widget>[
                _LoadingLine(width: 40, height: 16),
                Spacer(),
                _SkeletonPill(width: 188, height: 30, radius: 9),
              ],
            ),
            const SizedBox(height: 10),
            const Row(
              children: <Widget>[
                Expanded(child: _ScoreLoadingCard()),
                SizedBox(width: 14),
                Expanded(child: _ScoreLoadingCard()),
                SizedBox(width: 14),
                Expanded(child: _ScoreLoadingCard()),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _Pep3RightRailSkeleton extends StatelessWidget {
  const _Pep3RightRailSkeleton();

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: EdgeInsets.zero,
      physics: const NeverScrollableScrollPhysics(),
      children: const <Widget>[
        _RailCard(
          child: Padding(
            padding: EdgeInsets.all(16),
            child: _ProgressSkeleton(),
          ),
        ),
        SizedBox(height: 10),
        _RailCard(
          child: Padding(
            padding: EdgeInsets.all(16),
            child: _TrainingRecordSkeleton(),
          ),
        ),
        SizedBox(height: 10),
        _RailCard(
          child: Padding(
            padding: EdgeInsets.all(16),
            child: _CaregiverSkeleton(),
          ),
        ),
      ],
    );
  }
}

class _ProgressSkeleton extends StatelessWidget {
  const _ProgressSkeleton();

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        const _SkeletonPill(width: 82, height: 82, radius: 99),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: const <Widget>[
              _LoadingLine(width: 72, height: 16),
              SizedBox(height: 14),
              _LoadingLine(widthFactor: .86, height: 13),
              SizedBox(height: 11),
              _LoadingLine(widthFactor: .68, height: 13),
            ],
          ),
        ),
      ],
    );
  }
}

class _TrainingRecordSkeleton extends StatelessWidget {
  const _TrainingRecordSkeleton();

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: const <Widget>[
        _LoadingLine(width: 96, height: 18),
        SizedBox(height: 14),
        _SkeletonInputBlock(),
        SizedBox(height: 10),
        _SkeletonInputBlock(),
      ],
    );
  }
}

class _CaregiverSkeleton extends StatelessWidget {
  const _CaregiverSkeleton();

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: const <Widget>[
        _LoadingLine(width: 82, height: 18),
        SizedBox(height: 16),
        Center(child: _SkeletonPill(width: 122, height: 122, radius: 10)),
        SizedBox(height: 14),
        Center(child: _LoadingLine(width: 112, height: 13)),
        SizedBox(height: 14),
        Row(
          children: <Widget>[
            Expanded(child: _SkeletonButton(width: double.infinity)),
            SizedBox(width: 8),
            Expanded(child: _SkeletonButton(width: double.infinity)),
          ],
        ),
      ],
    );
  }
}

class _Pep3FooterSkeleton extends StatelessWidget {
  const _Pep3FooterSkeleton();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 62,
      padding: const EdgeInsets.symmetric(horizontal: 20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.97),
        border: Border.all(color: _Pep3Colors.line),
        borderRadius: const BorderRadius.vertical(top: Radius.circular(12)),
        boxShadow: _pep3Shadow(color: const Color(0x14B05F32), blur: 16),
      ),
      child: const Row(
        children: <Widget>[
          _SkeletonButton(width: 134),
          Spacer(),
          _LoadingLine(width: 80, height: 28),
          Spacer(),
          _SkeletonButton(width: 150, filled: true),
          SizedBox(width: 14),
          _SkeletonButton(width: 134),
          SizedBox(width: 22),
          _LoadingLine(width: 66, height: 13),
          SizedBox(width: 8),
          _SkeletonPill(width: 50, height: 30, radius: 99),
        ],
      ),
    );
  }
}

class _SkeletonInputBlock extends StatelessWidget {
  const _SkeletonInputBlock();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      height: 80,
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _Pep3Colors.lineSoft),
      ),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          _LoadingLine(width: 74, height: 13),
          SizedBox(height: 10),
          _SkeletonPill(height: 32, radius: 8),
        ],
      ),
    );
  }
}

class _SkeletonButton extends StatelessWidget {
  const _SkeletonButton({
    required this.width,
    this.filled = false,
  });

  final double width;
  final bool filled;

  @override
  Widget build(BuildContext context) {
    return _SkeletonPill(
      width: width,
      height: 38,
      radius: 10,
      color: filled ? const Color(0xFFF7C1A8) : null,
    );
  }
}

class _SkeletonPill extends StatelessWidget {
  const _SkeletonPill({
    this.width,
    required this.height,
    this.radius = 99,
    this.color,
  });

  final double? width;
  final double height;
  final double radius;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: color ?? const Color(0xFFF3E8DF),
        borderRadius: BorderRadius.circular(radius),
      ),
    );
  }
}
