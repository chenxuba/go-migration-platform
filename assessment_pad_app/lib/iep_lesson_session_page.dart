part of 'iep_center_page.dart';

class _IepLessonSessionDraft {
  const _IepLessonSessionDraft({
    required this.studentName,
    required this.gender,
    required this.ageLabel,
    required this.teacherName,
    required this.courseName,
    required this.planTitle,
    required this.stageLabel,
    required this.periodLabel,
    required this.weekLabel,
    required this.initialSelectedDateIndex,
    String? trainingDateLabel,
    List<String>? weekDateOptions,
    List<String>? completionColumnLabels,
    required this.preparation,
    required this.goals,
    required this.tasks,
  })  : _trainingDateLabel = trainingDateLabel,
        _weekDateOptions = weekDateOptions,
        _completionColumnLabels = completionColumnLabels;

  final String studentName;
  final String gender;
  final String ageLabel;
  final String teacherName;
  final String courseName;
  final String planTitle;
  final String stageLabel;
  final String periodLabel;
  final String weekLabel;
  final int initialSelectedDateIndex;
  final String? _trainingDateLabel;
  final List<String>? _weekDateOptions;
  final List<String>? _completionColumnLabels;
  final String preparation;
  final List<String> goals;
  final List<_IepLessonTaskDraft> tasks;

  String get trainingDateLabel => _trainingDateLabel ?? weekLabel;

  List<String> get weekDateOptions => _weekDateOptions ?? const <String>[];

  List<String> get completionColumnLabels =>
      _completionColumnLabels ?? const <String>[];
}

class _IepLessonFullscreenViewport extends StatelessWidget {
  const _IepLessonFullscreenViewport({required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final Size screenSize = MediaQuery.sizeOf(context);
        final double viewportWidth = constraints.maxWidth.isFinite
            ? constraints.maxWidth
            : screenSize.width;
        final double viewportHeight = constraints.maxHeight.isFinite
            ? constraints.maxHeight
            : screenSize.height;
        final double designWidth = padDesignWidthForViewport(
          viewportWidth,
          viewportHeight,
        );
        return ColoredBox(
          color: _IepColors.page,
          child: Center(
            child: FittedBox(
              fit: BoxFit.contain,
              child: SizedBox(
                width: designWidth,
                height: padDesignHeight,
                child: child,
              ),
            ),
          ),
        );
      },
    );
  }
}

class _IepLessonTaskDraft {
  const _IepLessonTaskDraft({
    required this.title,
    required this.subtitle,
    required this.domain,
    required this.goal,
    required this.materials,
    required this.steps,
    required this.tips,
    List<String>? completionCodes,
  }) : _completionCodes = completionCodes;

  final String title;
  final String subtitle;
  final String domain;
  final String goal;
  final String materials;
  final List<String> steps;
  final List<String> tips;
  final List<String>? _completionCodes;

  List<String> get completionCodes => _completionCodes ?? const <String>[];
}

class _IepLessonSessionPage extends StatefulWidget {
  const _IepLessonSessionPage({
    required this.onBack,
    required this.draft,
  });

  final VoidCallback onBack;
  final _IepLessonSessionDraft draft;

  @override
  State<_IepLessonSessionPage> createState() => _IepLessonSessionPageState();
}

class _IepLessonSessionPageState extends State<_IepLessonSessionPage> {
  int _selectedTaskIndex = 0;
  int _selectedDateIndex = 0;
  List<List<String>> _taskCompletionCodes = <List<String>>[];

  @override
  void initState() {
    super.initState();
    _taskCompletionCodes = _normalizedCompletionCodes(widget.draft);
    _selectedDateIndex = _initialSelectedDateIndexFor(widget.draft);
  }

  @override
  void didUpdateWidget(covariant _IepLessonSessionPage oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (!identical(oldWidget.draft, widget.draft)) {
      _taskCompletionCodes = _normalizedCompletionCodes(widget.draft);
      _selectedTaskIndex = 0;
      _selectedDateIndex = _initialSelectedDateIndexFor(widget.draft);
    }
  }

  int _initialSelectedDateIndexFor(_IepLessonSessionDraft draft) {
    if (draft.weekDateOptions.isEmpty) {
      return 0;
    }
    return draft.initialSelectedDateIndex.clamp(
      0,
      math.max(0, draft.weekDateOptions.length - 1),
    );
  }

  List<List<String>> _normalizedCompletionCodes(_IepLessonSessionDraft draft) {
    final int columnCount = math.max(
      1,
      draft.completionColumnLabels.isNotEmpty
          ? draft.completionColumnLabels.length
          : draft.weekDateOptions.length,
    );
    return draft.tasks.map((_IepLessonTaskDraft task) {
      return List<String>.generate(columnCount, (int index) {
        if (index < task.completionCodes.length) {
          return task.completionCodes[index].trim();
        }
        return '';
      });
    }).toList(growable: false);
  }

  void _updateCompletionCode(String code) {
    if (_taskCompletionCodes.isEmpty) {
      return;
    }
    final int taskIndex = _selectedTaskIndex.clamp(
      0,
      math.max(0, _taskCompletionCodes.length - 1),
    );
    final List<String> taskCodes = _taskCompletionCodes[taskIndex];
    if (taskCodes.isEmpty) {
      return;
    }
    final int dateIndex = _selectedDateIndex.clamp(
      0,
      math.max(0, taskCodes.length - 1),
    );
    final List<List<String>> next = _taskCompletionCodes
        .map((List<String> item) => List<String>.from(item))
        .toList(growable: false);
    next[taskIndex][dateIndex] = next[taskIndex][dateIndex] == code ? '' : code;
    setState(() {
      _taskCompletionCodes = next;
    });
  }

  @override
  Widget build(BuildContext context) {
    final _IepLessonSessionDraft draft = widget.draft;
    final List<_IepLessonTaskDraft> tasks = draft.tasks;
    final List<String> weekDateOptions = draft.weekDateOptions;
    if (_taskCompletionCodes.length != tasks.length) {
      _taskCompletionCodes = _normalizedCompletionCodes(draft);
    }
    final int selectedIndex = tasks.isEmpty
        ? 0
        : _selectedTaskIndex.clamp(0, math.max(0, tasks.length - 1));
    final int selectedDateIndex = weekDateOptions.isEmpty
        ? 0
        : _selectedDateIndex.clamp(0, math.max(0, weekDateOptions.length - 1));
    final _IepLessonTaskDraft? selectedTask =
        tasks.isEmpty ? null : tasks[selectedIndex];
    final String selectedDateLabel = weekDateOptions.isEmpty
        ? draft.trainingDateLabel
        : weekDateOptions[selectedDateIndex];
    final int recordedCount = _taskCompletionCodes.where((List<String> item) {
      return selectedDateIndex < item.length &&
          item[selectedDateIndex].isNotEmpty;
    }).length;

    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width =
            constraints.maxWidth.isFinite ? constraints.maxWidth : 1366;
        final bool compact = width < 1180;
        final double outer = compact ? 14 : 20;
        final double gap = compact ? 10 : 14;
        final double leftWidth = compact ? 248 : 272;
        final double rightWidth = compact ? 256 : 292;
        final double centerWidth =
            width - outer * 2 - leftWidth - rightWidth - gap * 2;

        return ColoredBox(
          color: const Color(0xFFFFF7EE),
          child: Stack(
            children: <Widget>[
              Positioned.fill(
                child: CustomPaint(painter: _IepLessonBackgroundPainter()),
              ),
              Positioned(
                left: 0,
                right: 0,
                top: 0,
                child: _IepLessonTopBar(
                  onBack: widget.onBack,
                  draft: draft,
                  selectedDateLabel: selectedDateLabel,
                ),
              ),
              Positioned(
                left: outer,
                top: 84,
                width: leftWidth,
                height: 660,
                child: _IepLessonTaskRail(
                  draft: draft,
                  selectedIndex: selectedIndex,
                  selectedDateIndex: selectedDateIndex,
                  recordedCount: recordedCount,
                  completionCodes: _taskCompletionCodes,
                  onTaskSelected: (int index) {
                    setState(() {
                      _selectedTaskIndex = index;
                    });
                  },
                ),
              ),
              Positioned(
                left: outer + leftWidth + gap,
                top: 84,
                width: centerWidth,
                height: 660,
                child: _IepLessonMainPanel(
                  draft: draft,
                  task: selectedTask,
                  taskIndex: selectedIndex,
                  selectedDateLabel: selectedDateLabel,
                ),
              ),
              Positioned(
                right: outer,
                top: 84,
                width: rightWidth,
                height: 660,
                child: _IepLessonRecordPanel(
                  draft: draft,
                  task: selectedTask,
                  currentCodes: selectedIndex < _taskCompletionCodes.length
                      ? _taskCompletionCodes[selectedIndex]
                      : const <String>[],
                  weekDateOptions: weekDateOptions,
                  selectedDateIndex: selectedDateIndex,
                  onCodeSelected: _updateCompletionCode,
                  selectedDateLabel: selectedDateLabel,
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}

class _IepLessonBackgroundPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint wash = Paint()
      ..shader = const LinearGradient(
        begin: Alignment.topCenter,
        end: Alignment.bottomCenter,
        colors: <Color>[Color(0xFFFFFBF7), Color(0xFFFFF3E7)],
      ).createShader(Offset.zero & size);
    canvas.drawRect(Offset.zero & size, wash);

    final Paint circleA = Paint()..color = const Color(0x22F3C39D);
    final Paint circleB = Paint()..color = const Color(0x14E9854E);
    canvas.drawCircle(
        Offset(size.width * .14, size.height * .08), 120, circleA);
    canvas.drawCircle(
        Offset(size.width * .84, size.height * .18), 140, circleB);
    canvas.drawCircle(
        Offset(size.width * .74, size.height * .82), 180, circleA);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _IepLessonTopBar extends StatelessWidget {
  const _IepLessonTopBar({
    required this.onBack,
    required this.draft,
    required this.selectedDateLabel,
  });

  final VoidCallback onBack;
  final _IepLessonSessionDraft draft;
  final String selectedDateLabel;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 72,
      padding: const EdgeInsets.symmetric(horizontal: 20),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        border: Border(
          bottom: BorderSide(color: _IepColors.line.withOpacity(.78)),
        ),
      ),
      child: Row(
        children: <Widget>[
          _IepLessonBackButton(onTap: onBack),
          const SizedBox(width: 16),
          _IepLessonStudentCard(draft: draft),
          const SizedBox(width: 12),
          _IepLessonMetaBadge(
            icon: Icons.assignment_rounded,
            text: draft.weekLabel,
          ),
          const SizedBox(width: 8),
          _IepLessonMetaBadge(
            icon: Icons.today_rounded,
            text: selectedDateLabel,
          ),
          const SizedBox(width: 8),
          _IepLessonMetaBadge(
            icon: Icons.schedule_rounded,
            text: '进行中 32:18',
            tone: _IepLessonBadgeTone.orange,
          ),
          const Spacer(),
          _IepLessonActionButton(
            label: '暂停',
            icon: Icons.pause_circle_outline_rounded,
            filled: false,
          ),
          const SizedBox(width: 10),
          _IepLessonActionButton(
            label: '结束上课',
            icon: Icons.stop_circle_rounded,
            filled: true,
          ),
        ],
      ),
    );
  }
}

class _IepLessonBackButton extends StatelessWidget {
  const _IepLessonBackButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.white,
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

class _IepLessonStudentCard extends StatelessWidget {
  const _IepLessonStudentCard({required this.draft});

  final _IepLessonSessionDraft draft;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 44,
      padding: const EdgeInsets.fromLTRB(6, 6, 14, 6),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: _IepColors.line),
      ),
      child: Row(
        children: <Widget>[
          _IepLessonAvatar(name: draft.studentName, size: 32),
          const SizedBox(width: 10),
          Text(
            '${draft.studentName} · ${draft.ageLabel}',
            style: const TextStyle(
              color: _IepColors.ink,
              fontSize: 14,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

enum _IepLessonBadgeTone { neutral, orange }

class _IepLessonMetaBadge extends StatelessWidget {
  const _IepLessonMetaBadge({
    required this.icon,
    required this.text,
    this.tone = _IepLessonBadgeTone.neutral,
  });

  final IconData icon;
  final String text;
  final _IepLessonBadgeTone tone;

  @override
  Widget build(BuildContext context) {
    final bool orange = tone == _IepLessonBadgeTone.orange;
    return Container(
      height: 34,
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: orange ? const Color(0xFFFFEFE4) : const Color(0xFFFFFBF7),
        borderRadius: BorderRadius.circular(17),
        border: Border.all(
          color: orange ? const Color(0xFFF4D0B6) : _IepColors.lightLine,
        ),
      ),
      child: Row(
        children: <Widget>[
          Icon(
            icon,
            size: 16,
            color: orange ? _IepColors.orangeDeep : _IepColors.muted,
          ),
          const SizedBox(width: 6),
          Text(
            text,
            style: TextStyle(
              color: orange ? _IepColors.orangeDeep : _IepColors.text,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonActionButton extends StatelessWidget {
  const _IepLessonActionButton({
    required this.label,
    required this.icon,
    required this.filled,
  });

  final String label;
  final IconData icon;
  final bool filled;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 38,
      padding: const EdgeInsets.symmetric(horizontal: 14),
      decoration: BoxDecoration(
        color: filled ? _IepColors.orange : Colors.white,
        borderRadius: BorderRadius.circular(19),
        border: Border.all(
          color: filled ? _IepColors.orange : _IepColors.line,
        ),
        boxShadow: filled
            ? _iepShadow(
                color: const Color(0x26E96F43),
                blur: 14,
                offset: const Offset(0, 5),
              )
            : const <BoxShadow>[],
      ),
      child: Row(
        children: <Widget>[
          Icon(
            icon,
            size: 18,
            color: filled ? Colors.white : _IepColors.text,
          ),
          const SizedBox(width: 6),
          Text(
            label,
            style: TextStyle(
              color: filled ? Colors.white : _IepColors.text,
              fontSize: 13,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonTaskRail extends StatelessWidget {
  const _IepLessonTaskRail({
    required this.draft,
    required this.selectedIndex,
    required this.selectedDateIndex,
    required this.recordedCount,
    required this.completionCodes,
    required this.onTaskSelected,
  });

  final _IepLessonSessionDraft draft;
  final int selectedIndex;
  final int selectedDateIndex;
  final int recordedCount;
  final List<List<String>> completionCodes;
  final ValueChanged<int> onTaskSelected;

  @override
  Widget build(BuildContext context) {
    final List<_IepLessonTaskDraft> tasks = draft.tasks;
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.95),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _IepColors.line),
        boxShadow: _iepShadow(),
      ),
      padding: const EdgeInsets.fromLTRB(14, 14, 14, 14),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '训练项目',
            style: TextStyle(
              color: _IepColors.ink,
              fontSize: 16,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 6),
          Text(
            '${recordedCount.clamp(0, tasks.length)}/${tasks.length} 已记录',
            style: const TextStyle(
              color: _IepColors.muted,
              fontSize: 12,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            draft.weekDateOptions.isEmpty
                ? draft.trainingDateLabel
                : draft.weekDateOptions[selectedDateIndex],
            style: const TextStyle(
              color: _IepColors.orangeDeep,
              fontSize: 12,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 14),
          Container(
            height: 8,
            clipBehavior: Clip.antiAlias,
            decoration: BoxDecoration(
              color: const Color(0xFFFFF0E5),
              borderRadius: BorderRadius.circular(999),
            ),
            child: FractionallySizedBox(
              widthFactor: tasks.isEmpty ? 0 : recordedCount / tasks.length,
              alignment: Alignment.centerLeft,
              child: Container(color: _IepColors.orange),
            ),
          ),
          const SizedBox(height: 14),
          Expanded(
            child: SingleChildScrollView(
              physics: const BouncingScrollPhysics(),
              child: Column(
                children: List<Widget>.generate(tasks.length, (int index) {
                  final _IepLessonTaskDraft task = tasks[index];
                  final bool selected = index == selectedIndex;
                  final List<String> taskCodes = index < completionCodes.length
                      ? completionCodes[index]
                      : const <String>[];
                  final String currentCode =
                      selectedDateIndex < taskCodes.length
                          ? taskCodes[selectedDateIndex].trim()
                          : '';
                  return Padding(
                    padding: EdgeInsets.only(
                      bottom: index == tasks.length - 1 ? 0 : 10,
                    ),
                    child: _IepLessonTaskCard(
                      index: index,
                      task: task,
                      selected: selected,
                      currentCode: currentCode,
                      onTap: () => onTaskSelected(index),
                    ),
                  );
                }),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonTaskCard extends StatelessWidget {
  const _IepLessonTaskCard({
    required this.index,
    required this.task,
    required this.selected,
    required this.currentCode,
    required this.onTap,
  });

  final int index;
  final _IepLessonTaskDraft task;
  final bool selected;
  final String currentCode;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final Color borderColor =
        selected ? _IepColors.orange.withOpacity(.55) : _IepColors.lightLine;
    final Color fillColor = selected ? const Color(0xFFFFF6EE) : Colors.white;
    final bool recorded = currentCode.isNotEmpty;
    final String stateLabel = recorded ? _lessonCodeLabel(currentCode) : '待记录';
    final Color stateColor =
        recorded ? _lessonCodeColor(currentCode) : _IepColors.muted;

    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(16),
        child: Container(
          padding: const EdgeInsets.fromLTRB(12, 12, 12, 12),
          decoration: BoxDecoration(
            color: fillColor,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: borderColor),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Container(
                width: 26,
                height: 26,
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: recorded
                      ? _lessonCodeColor(currentCode).withOpacity(.12)
                      : selected
                          ? _IepColors.orangeSoft
                          : const Color(0xFFFFFAF6),
                  shape: BoxShape.circle,
                ),
                child: recorded
                    ? Text(
                        currentCode,
                        style: TextStyle(
                          color: _lessonCodeColor(currentCode),
                          fontSize: 12,
                          fontWeight: FontWeight.w900,
                        ),
                      )
                    : Text(
                        '${index + 1}',
                        style: const TextStyle(
                          color: _IepColors.orangeDeep,
                          fontSize: 12,
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
                      task.title,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: _IepColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                        height: 1.35,
                      ),
                    ),
                    const SizedBox(height: 10),
                    Row(
                      children: <Widget>[
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: selected
                                ? const Color(0xFFFFEEDF)
                                : const Color(0xFFFFFBF7),
                            borderRadius: BorderRadius.circular(999),
                          ),
                          child: Text(
                            '第${index + 1}项',
                            style: const TextStyle(
                              color: _IepColors.muted,
                              fontSize: 11,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                        ),
                        const Spacer(),
                        Text(
                          stateLabel,
                          style: TextStyle(
                            color: stateColor,
                            fontSize: 11,
                            fontWeight: FontWeight.w900,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _IepLessonMainPanel extends StatelessWidget {
  const _IepLessonMainPanel({
    required this.draft,
    required this.task,
    required this.taskIndex,
    required this.selectedDateLabel,
  });

  final _IepLessonSessionDraft draft;
  final _IepLessonTaskDraft? task;
  final int taskIndex;
  final String selectedDateLabel;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _IepColors.line),
        boxShadow: _iepShadow(),
      ),
      padding: const EdgeInsets.fromLTRB(20, 18, 20, 18),
      child: task == null
          ? const Center(
              child: Text(
                '暂无训练任务',
                style: TextStyle(
                  color: _IepColors.muted,
                  fontSize: 14,
                  fontWeight: FontWeight.w700,
                ),
              ),
            )
          : Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                _IepLessonHeroCard(
                  eyebrow: '当前训练',
                  title: task!.title,
                  content: task!.subtitle,
                  indexLabel: '第${taskIndex + 1}项',
                ),
                const SizedBox(height: 10),
                _IepLessonMetaStrip(
                  teacherName: draft.teacherName,
                  courseName: draft.courseName,
                  selectedDateLabel: selectedDateLabel,
                ),
                const SizedBox(height: 12),
                Expanded(
                  child: SingleChildScrollView(
                    physics: const BouncingScrollPhysics(),
                    child: _IepLessonPreparationCard(
                      preparation: draft.preparation,
                    ),
                  ),
                ),
              ],
            ),
    );
  }
}

class _IepLessonMetaStrip extends StatelessWidget {
  const _IepLessonMetaStrip({
    required this.teacherName,
    required this.courseName,
    required this.selectedDateLabel,
  });

  final String teacherName;
  final String courseName;
  final String selectedDateLabel;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 4),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFBF7),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _IepColors.lightLine),
      ),
      child: Row(
        children: <Widget>[
          Expanded(
            child: _IepLessonMetaItem(
              label: '任教老师',
              value: teacherName,
            ),
          ),
          const _IepLessonMetaDivider(),
          Expanded(
            child: _IepLessonMetaItem(
              label: '课程名称',
              value: courseName,
            ),
          ),
          const _IepLessonMetaDivider(),
          Expanded(
            child: _IepLessonMetaItem(
              label: '训练日期',
              value: selectedDateLabel,
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonMetaItem extends StatelessWidget {
  const _IepLessonMetaItem({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(12, 8, 12, 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Text(
            label,
            style: const TextStyle(
              color: _IepColors.muted,
              fontSize: 11,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 5),
          Text(
            value.trim().isEmpty ? '-' : value.trim(),
            style: const TextStyle(
              color: _IepColors.ink,
              fontSize: 13,
              fontWeight: FontWeight.w900,
              height: 1.35,
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonMetaDivider extends StatelessWidget {
  const _IepLessonMetaDivider();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 1,
      height: 34,
      color: _IepColors.lightLine,
    );
  }
}

class _IepLessonHeroCard extends StatelessWidget {
  const _IepLessonHeroCard({
    required this.eyebrow,
    required this.title,
    required this.content,
    required this.indexLabel,
  });

  final String eyebrow;
  final String title;
  final String content;
  final String indexLabel;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(18, 18, 18, 18),
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(20),
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: <Color>[Color(0xFFFFFCF8), Color(0xFFFFF4EA)],
        ),
        border: Border.all(color: const Color(0xFFF2D9C6)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(.72),
                  borderRadius: BorderRadius.circular(999),
                  border: Border.all(color: const Color(0xFFF0D7C5)),
                ),
                child: Text(
                  eyebrow,
                  style: const TextStyle(
                    color: _IepColors.orangeDeep,
                    fontSize: 11,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              const Spacer(),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFF8F1),
                  borderRadius: BorderRadius.circular(14),
                  border: Border.all(color: const Color(0xFFF0D7C5)),
                ),
                child: Text(
                  indexLabel,
                  style: const TextStyle(
                    color: _IepColors.orangeDeep,
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          Text(
            title.trim().isEmpty ? '-' : title.trim(),
            style: const TextStyle(
              color: _IepColors.ink,
              fontSize: 26,
              fontWeight: FontWeight.w900,
              height: 1.22,
            ),
          ),
          const SizedBox(height: 12),
          Text(
            content.trim().isEmpty ? '-' : content.trim(),
            style: const TextStyle(
              color: _IepColors.text,
              fontSize: 16,
              fontWeight: FontWeight.w800,
              height: 1.65,
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonPreparationCard extends StatelessWidget {
  const _IepLessonPreparationCard({
    required this.preparation,
  });

  final String preparation;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(18, 16, 18, 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _IepColors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '训练前准备',
            style: TextStyle(
              color: _IepColors.orangeDeep,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 12),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
            decoration: BoxDecoration(
              color: const Color(0xFFFFFAF6),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: _IepColors.lightLine),
            ),
            child: Text(
              preparation.trim().isEmpty ? '-' : preparation.trim(),
              style: const TextStyle(
                color: _IepColors.text,
                fontSize: 14,
                fontWeight: FontWeight.w800,
                height: 1.72,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _IepLessonRecordPanel extends StatelessWidget {
  const _IepLessonRecordPanel({
    required this.draft,
    required this.task,
    required this.currentCodes,
    required this.weekDateOptions,
    required this.selectedDateIndex,
    required this.onCodeSelected,
    required this.selectedDateLabel,
  });

  final _IepLessonSessionDraft draft;
  final _IepLessonTaskDraft? task;
  final List<String> currentCodes;
  final List<String> weekDateOptions;
  final int selectedDateIndex;
  final ValueChanged<String> onCodeSelected;
  final String selectedDateLabel;

  @override
  Widget build(BuildContext context) {
    final List<String> dateOptions = weekDateOptions.isEmpty
        ? draft.completionColumnLabels
        : weekDateOptions;
    final int safeDateIndex = dateOptions.isEmpty
        ? 0
        : selectedDateIndex.clamp(0, math.max(0, dateOptions.length - 1));
    final String currentCode = safeDateIndex < currentCodes.length
        ? currentCodes[safeDateIndex].trim()
        : '';
    const List<String> codeOptions = <String>[
      '√',
      '✗',
      'S',
      'G',
      'M',
      'V',
      'P'
    ];
    return Container(
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.95),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _IepColors.line),
        boxShadow: _iepShadow(),
      ),
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '记录区域',
            style: TextStyle(
              color: _IepColors.ink,
              fontSize: 16,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 6),
          Text(
            task?.title ?? '请选择训练项目',
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: _IepColors.muted,
              fontSize: 12,
              fontWeight: FontWeight.w700,
              height: 1.35,
            ),
          ),
          const SizedBox(height: 14),
          Expanded(
            child: SingleChildScrollView(
              physics: const BouncingScrollPhysics(),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  _IepLessonRecordSection(
                    title: '训练日期',
                    child: _IepLessonDateChip(
                      label: selectedDateLabel,
                      selected: true,
                    ),
                  ),
                  const SizedBox(height: 12),
                  _IepLessonRecordSection(
                    title: '当前记录',
                    child: Container(
                      padding: const EdgeInsets.fromLTRB(12, 10, 12, 10),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: _IepColors.lightLine),
                      ),
                      child: Row(
                        children: <Widget>[
                          Container(
                            width: 42,
                            height: 42,
                            alignment: Alignment.center,
                            decoration: BoxDecoration(
                              color: _lessonCodeColor(currentCode)
                                  .withOpacity(.12),
                              shape: BoxShape.circle,
                            ),
                            child: Text(
                              currentCode.isEmpty ? '·' : currentCode,
                              style: TextStyle(
                                color: _lessonCodeColor(currentCode),
                                fontSize: currentCode.isEmpty ? 22 : 20,
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
                                  currentCode.isEmpty
                                      ? '待记录'
                                      : _lessonCodeLabel(currentCode),
                                  style: TextStyle(
                                    color: _lessonCodeColor(currentCode),
                                    fontSize: 16,
                                    fontWeight: FontWeight.w900,
                                  ),
                                ),
                                const SizedBox(height: 2),
                                Text(
                                  currentCode.isEmpty
                                      ? '请选择一项数据记录'
                                      : '当前记录为 $currentCode',
                                  style: const TextStyle(
                                    color: _IepColors.text,
                                    fontSize: 11,
                                    fontWeight: FontWeight.w700,
                                    height: 1.35,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 12),
                  _IepLessonRecordSection(
                    title: '数据记录',
                    child: _IepLessonCodeGrid(
                      codes: codeOptions,
                      currentCode: currentCode,
                      enabled: task != null,
                      onCodeSelected: onCodeSelected,
                    ),
                  ),
                ],
              ),
            ),
          ),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.fromLTRB(14, 12, 14, 12),
            decoration: BoxDecoration(
              color: const Color(0xFFFFFBF7),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(color: _IepColors.lightLine),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                const Icon(
                  Icons.lightbulb_outline_rounded,
                  size: 18,
                  color: _IepColors.orangeDeep,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    '结束上课后，可将本次记录回写到 ${draft.weekLabel} 中 $selectedDateLabel 的完成情况。',
                    style: const TextStyle(
                      color: _IepColors.text,
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      height: 1.45,
                    ),
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

class _IepLessonRecordSection extends StatelessWidget {
  const _IepLessonRecordSection({
    required this.title,
    required this.child,
  });

  final String title;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          title,
          style: const TextStyle(
            color: _IepColors.muted,
            fontSize: 12,
            fontWeight: FontWeight.w800,
          ),
        ),
        const SizedBox(height: 8),
        child,
      ],
    );
  }
}

class _IepLessonDateChip extends StatelessWidget {
  const _IepLessonDateChip({
    required this.label,
    required this.selected,
  });

  final String label;
  final bool selected;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      decoration: BoxDecoration(
        color: selected ? const Color(0xFFFFF1E6) : Colors.white,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(
          color: selected ? _IepColors.orange : _IepColors.line,
        ),
      ),
      child: Text(
        label,
        textAlign: TextAlign.center,
        style: TextStyle(
          color: selected ? _IepColors.orangeDeep : _IepColors.text,
          fontSize: 12,
          fontWeight: FontWeight.w800,
          height: 1.35,
        ),
      ),
    );
  }
}

class _IepLessonCodeGrid extends StatelessWidget {
  const _IepLessonCodeGrid({
    required this.codes,
    required this.currentCode,
    required this.enabled,
    required this.onCodeSelected,
  });

  final List<String> codes;
  final String currentCode;
  final bool enabled;
  final ValueChanged<String> onCodeSelected;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double itemWidth = (constraints.maxWidth - 10) / 2;
        return Wrap(
          spacing: 10,
          runSpacing: 10,
          children: codes.map((String code) {
            return _IepLessonCodeCard(
              code: code,
              label: _lessonCodeLabel(code),
              width: itemWidth,
              selected: currentCode == code,
              enabled: enabled,
              onTap: () => onCodeSelected(code),
            );
          }).toList(),
        );
      },
    );
  }
}

class _IepLessonCodeCard extends StatelessWidget {
  const _IepLessonCodeCard({
    required this.code,
    required this.label,
    required this.width,
    required this.selected,
    required this.enabled,
    required this.onTap,
  });

  final String code;
  final String label;
  final double width;
  final bool selected;
  final bool enabled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Opacity(
      opacity: enabled ? 1 : .45,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: enabled ? onTap : null,
          borderRadius: BorderRadius.circular(16),
          child: Container(
            width: width,
            padding: const EdgeInsets.fromLTRB(12, 12, 12, 12),
            decoration: BoxDecoration(
              color: selected
                  ? _lessonCodeColor(code).withOpacity(.12)
                  : Colors.white,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: selected ? _lessonCodeColor(code) : _IepColors.line,
                width: selected ? 1.4 : 1,
              ),
            ),
            child: Row(
              children: <Widget>[
                Container(
                  width: 36,
                  height: 36,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: _lessonCodeColor(code).withOpacity(.12),
                    shape: BoxShape.circle,
                  ),
                  child: Text(
                    code,
                    style: TextStyle(
                      color: _lessonCodeColor(code),
                      fontSize: 18,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    label,
                    style: TextStyle(
                      color:
                          selected ? _lessonCodeColor(code) : _IepColors.text,
                      fontSize: 13,
                      fontWeight: FontWeight.w800,
                      height: 1.25,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

String _lessonCodeLabel(String code) {
  switch (code) {
    case '√':
      return '独立完成';
    case '✗':
      return '未完成';
    case 'S':
      return '语言提示';
    case 'G':
      return '手势提示';
    case 'M':
      return '示范辅助';
    case 'V':
      return '视觉提示';
    case 'P':
      return '肢体辅助';
    default:
      return '待记录';
  }
}

Color _lessonCodeColor(String code) {
  switch (code) {
    case '√':
      return _IepColors.green;
    case '✗':
      return const Color(0xFFD2573F);
    case 'S':
      return const Color(0xFFE0A339);
    case 'G':
      return const Color(0xFFCE7F3B);
    case 'M':
      return const Color(0xFFB77BCE);
    case 'V':
      return const Color(0xFF5E98C9);
    case 'P':
      return const Color(0xFF8D6E63);
    default:
      return _IepColors.muted;
  }
}

class _IepLessonAvatar extends StatelessWidget {
  const _IepLessonAvatar({
    required this.name,
    required this.size,
  });

  final String name;
  final double size;

  @override
  Widget build(BuildContext context) {
    final int seed = name.runes.fold<int>(0, (int sum, int item) => sum + item);
    final List<List<Color>> palettes = <List<Color>>[
      const <Color>[Color(0xFFFFD8C2), Color(0xFFFFA36F)],
      const <Color>[Color(0xFFFFE0B7), Color(0xFFFFB067)],
      const <Color>[Color(0xFFFFD2C8), Color(0xFFFF8E75)],
      const <Color>[Color(0xFFFFE5BF), Color(0xFFFFB74E)],
    ];
    final List<Color> colors = palettes[seed % palettes.length];
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: colors,
        ),
      ),
      child: Icon(
        Icons.face_rounded,
        size: size * .72,
        color: const Color(0xFF6B4336),
      ),
    );
  }
}
