part of 'iep_center_page.dart';

const PdfColor _iepPrintInk = PdfColor.fromInt(0xff111111);
const PdfColor _iepPrintText = PdfColor.fromInt(0xff111111);
const PdfColor _iepPrintBorder = PdfColor.fromInt(0xff000000);
const PdfColor _iepPrintWhite = PdfColor.fromInt(0xffffffff);
const List<int> _iepTotalPrintColumns = <int>[
  659130,
  934720,
  396875,
  556895,
  800100,
  991870,
  588645,
  1472565,
];
const List<int> _iepMonthPrintColumns = <int>[
  511810,
  575945,
  575945,
  608330,
  607695,
  560070,
  560070,
  560070,
  560070,
  524510,
  377825,
  377825,
];
const List<int> _iepWeekPrintColumns = <int>[
  825500,
  847090,
  846455,
  1222375,
  443230,
  443230,
  443230,
  443230,
  443230,
  443230,
];

Future<Uint8List> _buildIepPlanPrintPdf({
  required PdfPageFormat format,
  required _IepPreviewMode mode,
  required IepPlan? totalPlan,
  required List<_DocDomainData> totalDomains,
  required IepMonthlyPlan? monthPlan,
  required IepWeeklyPlan? weekPlan,
  required String periodText,
  required String monthLabel,
  required int weekNumber,
  required DateTimeRange monthRange,
  required List<DateTime> weekDates,
}) async {
  final pw.Font baseFont =
      await fontFromAssetBundle('assets/fonts/NotoSansSC-Regular.ttf');
  final pw.Document document = pw.Document(
    theme: pw.ThemeData.withFont(
      base: baseFont,
      bold: baseFont,
      fontFallback: <pw.Font>[baseFont],
    ),
  );
  final PdfPageFormat baseFormat =
      format.width.isFinite && format.height.isFinite
          ? format.portrait
          : PdfPageFormat.a4;
  final PdfPageFormat pageFormat = baseFormat.copyWith(
    marginLeft: 28,
    marginTop: 32,
    marginRight: 28,
    marginBottom: 32,
  );

  document.addPage(
    pw.MultiPage(
      pageFormat: pageFormat,
      orientation: pw.PageOrientation.portrait,
      build: (pw.Context context) {
        return switch (mode) {
          _IepPreviewMode.total => _iepTotalPrintWidgets(
              plan: totalPlan!,
              domains: totalDomains,
              periodText: periodText,
            ),
          _IepPreviewMode.month => _iepMonthPrintWidgets(
              plan: monthPlan!,
              monthLabel: monthLabel,
              monthRange: monthRange,
            ),
          _IepPreviewMode.week => _iepWeekPrintWidgets(
              plan: weekPlan!,
              monthLabel: monthLabel,
              weekNumber: weekNumber,
              weekDates: weekDates,
            ),
        };
      },
    ),
  );
  return document.save();
}

List<pw.Widget> _iepTotalPrintWidgets({
  required IepPlan plan,
  required List<_DocDomainData> domains,
  required String periodText,
}) {
  const List<int> columns = _iepTotalPrintColumns;
  return <pw.Widget>[
    _iepPrintTitle(
      plan.title.trim().isEmpty ? '康复教学季度计划' : plan.title,
    ),
    _iepPrintRow(columns, <_IepPrintCell>[
      _IepPrintCell('姓名', 1, bold: true),
      _IepPrintCell(plan.student.name, 1),
      _IepPrintCell('性别', 1, bold: true),
      _IepPrintCell(plan.student.gender, 1),
      _IepPrintCell('出生年月', 1, bold: true),
      _IepPrintCell(plan.student.birthDate, 3),
    ]),
    _iepPrintRow(columns, <_IepPrintCell>[
      _IepPrintCell('制定日期', 1, bold: true),
      _IepPrintCell(plan.meta.planDate, 3),
      _IepPrintCell('计划参与者', 1, bold: true),
      _IepPrintCell(plan.meta.participant, 3),
    ]),
    _iepPrintRow(columns, <_IepPrintCell>[
      _IepPrintCell('实施者', 1, bold: true),
      _IepPrintCell(plan.meta.implementer, 3),
      _IepPrintCell('实施\n起止日期', 1, bold: true),
      _IepPrintCell(_metaRangeText(plan.meta, fallback: periodText), 3),
    ]),
    _iepPrintRow(columns, <_IepPrintCell>[
      _IepPrintCell('康复\n领域', 1, bold: true),
      _IepPrintCell('长期目标', 3, bold: true),
      _IepPrintCell('短期目标', 2, bold: true),
      _IepPrintCell('课程\n形式', 1, bold: true),
      _IepPrintCell('起止日期', 1, bold: true),
    ]),
    ...domains.expand((domain) => _iepTotalDomainWidgets(columns, domain)),
  ];
}

List<pw.Widget> _iepTotalDomainWidgets(
  List<int> columns,
  _DocDomainData domain,
) {
  final List<_DocShortGoalData> shortGoals = domain.shortGoals.isEmpty
      ? <_DocShortGoalData>[const _DocShortGoalData('', '', '')]
      : domain.shortGoals;
  final List<double> rowHeights = shortGoals.map((goal) {
    return math.max(
      _iepTextHeightEstimate(goal.goal, charsPerLine: 22, minHeight: 32),
      _iepTextHeightEstimate(goal.period, charsPerLine: 12, minHeight: 32),
    );
  }).toList();
  final double longGoalHeight = _iepTextHeightEstimate(
    domain.longGoals.join('\n'),
    charsPerLine: 35,
    minHeight: 72,
  );
  final double rowsHeight =
      rowHeights.fold<double>(0, (double sum, double item) => sum + item);
  if (longGoalHeight > rowsHeight && rowHeights.isNotEmpty) {
    final double extra = (longGoalHeight - rowsHeight) / rowHeights.length;
    for (int index = 0; index < rowHeights.length; index += 1) {
      rowHeights[index] += extra;
    }
  }
  final double blockHeight =
      rowHeights.fold<double>(0, (double sum, double item) => sum + item);
  return <pw.Widget>[
    pw.NewPage(freeSpace: blockHeight),
    pw.SizedBox(
      height: blockHeight,
      child: pw.Row(
        crossAxisAlignment: pw.CrossAxisAlignment.stretch,
        children: <pw.Widget>[
          _iepFlexCell(
            columns,
            _IepPrintCell(domain.domain, 1, bold: true),
            columnOffset: 0,
          ),
          _iepFlexCell(
            columns,
            _IepPrintCell(
              domain.longGoals.join('\n'),
              3,
              align: pw.TextAlign.left,
            ),
            columnOffset: 1,
          ),
          pw.Expanded(
            flex: _iepColumnFlex(columns, 4, 4),
            child: pw.Column(
              crossAxisAlignment: pw.CrossAxisAlignment.stretch,
              children: shortGoals.asMap().entries.map((entry) {
                final _DocShortGoalData goal = entry.value;
                return _iepPrintRow(
                  columns.sublist(4),
                  <_IepPrintCell>[
                    _IepPrintCell(goal.goal, 2, align: pw.TextAlign.left),
                    _IepPrintCell(goal.lesson, 1),
                    _IepPrintCell(goal.period, 1),
                  ],
                  height: rowHeights[entry.key],
                );
              }).toList(),
            ),
          ),
        ],
      ),
    ),
  ];
}

List<pw.Widget> _iepMonthPrintWidgets({
  required IepMonthlyPlan plan,
  required String monthLabel,
  required DateTimeRange monthRange,
}) {
  const List<int> columns = _iepMonthPrintColumns;
  final List<_MonthDomainData> domains = plan.rows
      .map(_monthDomainFromPlanRow)
      .where((_MonthDomainData item) =>
          item.domain.trim().isNotEmpty ||
          item.shortGoal.trim().isNotEmpty ||
          item.trainings.any(
            (_MonthTrainingData training) => training.content.trim().isNotEmpty,
          ))
      .toList();
  return <pw.Widget>[
    _iepPrintTitle(
      plan.title.trim().isEmpty ? '康复教学$monthLabel计划' : plan.title,
    ),
    _iepPrintRow(columns, <_IepPrintCell>[
      _IepPrintCell('姓名', 1, bold: true),
      _IepPrintCell(plan.student.name, 2),
      _IepPrintCell('性别', 1, bold: true),
      _IepPrintCell(plan.student.gender, 1),
      _IepPrintCell('出生年月', 2, bold: true),
      _IepPrintCell(plan.student.birthDate, 5),
    ]),
    _iepPrintRow(columns, <_IepPrintCell>[
      _IepPrintCell('制定\n日期', 1, bold: true),
      _IepPrintCell(plan.meta.planDate, 2),
      _IepPrintCell('计划参与者', 4, bold: true),
      _IepPrintCell(plan.meta.participant, 5),
    ]),
    _iepPrintRow(columns, <_IepPrintCell>[
      _IepPrintCell('实施者', 1, bold: true),
      _IepPrintCell(plan.meta.implementer, 2),
      _IepPrintCell('实施起止日期', 4, bold: true),
      _IepPrintCell(
        _metaRangeText(
          plan.meta,
          fallback: _formatZhRange(monthRange.start, monthRange.end),
        ),
        5,
      ),
    ]),
    _iepPrintRow(columns, <_IepPrintCell>[
      _IepPrintCell('康复\n领域', 1, bold: true),
      _IepPrintCell('长期目标', 2, bold: true),
      _IepPrintCell('短期目标', 2, bold: true),
      _IepPrintCell('训练内容', 4, bold: true),
      _IepPrintCell('课程\n形式', 1, bold: true),
      _IepPrintCell('起止日期', 2, bold: true),
    ]),
    ...domains.expand(
        (domain) => _iepMonthDomainWidgets(columns, domain, monthRange)),
  ];
}

List<pw.Widget> _iepMonthDomainWidgets(
  List<int> columns,
  _MonthDomainData domain,
  DateTimeRange monthRange,
) {
  final List<_MonthTrainingData> trainings = domain.trainings.isEmpty
      ? <_MonthTrainingData>[const _MonthTrainingData('', '')]
      : domain.trainings;
  final List<double> rowHeights = trainings.map((training) {
    return _iepTextHeightEstimate(
      training.content,
      charsPerLine: 42,
      minHeight: 42,
    );
  }).toList();
  final double mergedHeight = math.max(
    _iepTextHeightEstimate(domain.longGoal, charsPerLine: 20, minHeight: 80),
    _iepTextHeightEstimate(domain.shortGoal, charsPerLine: 16, minHeight: 80),
  );
  final double rowsHeight =
      rowHeights.fold<double>(0, (double sum, double item) => sum + item);
  if (mergedHeight > rowsHeight && rowHeights.isNotEmpty) {
    final double extra = (mergedHeight - rowsHeight) / rowHeights.length;
    for (int index = 0; index < rowHeights.length; index += 1) {
      rowHeights[index] += extra;
    }
  }
  final double blockHeight =
      rowHeights.fold<double>(0, (double sum, double item) => sum + item);
  return <pw.Widget>[
    pw.NewPage(freeSpace: blockHeight),
    pw.SizedBox(
      height: blockHeight,
      child: pw.Row(
        crossAxisAlignment: pw.CrossAxisAlignment.stretch,
        children: <pw.Widget>[
          _iepFlexCell(
            columns,
            _IepPrintCell(domain.domain, 1, bold: true),
            columnOffset: 0,
          ),
          _iepFlexCell(
            columns,
            _IepPrintCell(domain.longGoal, 2, align: pw.TextAlign.left),
            columnOffset: 1,
          ),
          _iepFlexCell(
            columns,
            _IepPrintCell(domain.shortGoal, 2, align: pw.TextAlign.left),
            columnOffset: 3,
          ),
          pw.Expanded(
            flex: _iepColumnFlex(columns, 5, 4),
            child: pw.Column(
              crossAxisAlignment: pw.CrossAxisAlignment.stretch,
              children: trainings.asMap().entries.map((entry) {
                return _iepPrintRow(
                  columns.sublist(5, 9),
                  <_IepPrintCell>[
                    _IepPrintCell(
                      entry.value.content,
                      4,
                      align: pw.TextAlign.left,
                    ),
                  ],
                  height: rowHeights[entry.key],
                );
              }).toList(),
            ),
          ),
          _iepFlexCell(
            columns,
            _IepPrintCell(domain.lesson, 1),
            columnOffset: 9,
          ),
          pw.Expanded(
            flex: _iepColumnFlex(columns, 10, 2),
            child: pw.Column(
              crossAxisAlignment: pw.CrossAxisAlignment.stretch,
              children: trainings.asMap().entries.map((entry) {
                return _iepPrintRow(
                  columns.sublist(10),
                  <_IepPrintCell>[
                    _IepPrintCell(
                        _monthTrainingPeriodText(monthRange, entry.key), 2),
                  ],
                  height: rowHeights[entry.key],
                );
              }).toList(),
            ),
          ),
        ],
      ),
    ),
  ];
}

List<pw.Widget> _iepWeekPrintWidgets({
  required IepWeeklyPlan plan,
  required String monthLabel,
  required int weekNumber,
  required List<DateTime> weekDates,
}) {
  const List<int> columns = _iepWeekPrintColumns;
  final List<DateTime> displayWeekDates =
      _dateListFromStrings(plan.weekDates) ?? weekDates;
  final List<_WeekTrainingRow> trainingRows = plan.rows
      .map(_weekTrainingRowFromPlanRow)
      .where((_WeekTrainingRow row) =>
          row.project.trim().isNotEmpty || row.content.trim().isNotEmpty)
      .toList();
  return <pw.Widget>[
    _iepPrintTitle(
      plan.title.trim().isEmpty
          ? '康复教学周计划日记录卡$monthLabel第$weekNumber周'
          : plan.title,
    ),
    _iepPrintRow(columns, <_IepPrintCell>[
      _IepPrintCell('姓名', 1, bold: true),
      _IepPrintCell(plan.student.name, 1),
      _IepPrintCell('性别', 1, bold: true),
      _IepPrintCell(plan.student.gender, 1),
      _IepPrintCell('出生年月', 2, bold: true),
      _IepPrintCell(plan.student.birthDate, 4),
    ]),
    _iepPrintRow(columns, <_IepPrintCell>[
      _IepPrintCell('任教\n老师', 1, bold: true),
      _IepPrintCell(
          plan.teacherName.trim().isEmpty ? '-' : plan.teacherName, 1),
      _IepPrintCell('课程\n名称', 1, bold: true),
      _IepPrintCell(
          plan.courseName.trim().isEmpty ? '康复教学' : plan.courseName, 1),
      _IepPrintCell('训练日期', 2, bold: true),
      _IepPrintCell(
        plan.trainingDate.trim().isEmpty
            ? _weekRangeText(displayWeekDates)
            : plan.trainingDate,
        4,
      ),
    ]),
    _iepPrintRow(
        columns,
        <_IepPrintCell>[
          _IepPrintCell('训练前\n准备', 1, bold: true),
          _IepPrintCell(
            plan.preparation.trim().isEmpty
                ? '训练材料、视觉提示卡、强化物、记录表'
                : plan.preparation,
            9,
            align: pw.TextAlign.left,
          ),
        ],
        minHeight: _iepTextHeightEstimate(
          plan.preparation.trim().isEmpty
              ? '训练材料、视觉提示卡、强化物、记录表'
              : plan.preparation,
          charsPerLine: 84,
          minHeight: 38,
        )),
    _iepWeekHeader(columns, displayWeekDates),
    ...trainingRows.map((row) {
      return _iepPrintRow(
          columns,
          <_IepPrintCell>[
            _IepPrintCell(row.project, 1, bold: true),
            _IepPrintCell(row.content, 3, align: pw.TextAlign.left),
            ...List<_IepPrintCell>.generate(6, (_) => _IepPrintCell('', 1)),
          ],
          minHeight: math.max(
            _iepTextHeightEstimate(row.project,
                charsPerLine: 10, minHeight: 38),
            _iepTextHeightEstimate(row.content,
                charsPerLine: 36, minHeight: 38),
          ));
    }),
  ];
}

pw.Widget _iepWeekHeader(List<int> columns, List<DateTime> dates) {
  return pw.SizedBox(
    height: 48,
    child: pw.Row(
      crossAxisAlignment: pw.CrossAxisAlignment.stretch,
      children: <pw.Widget>[
        _iepFlexCell(
          columns,
          _IepPrintCell('训练项目', 1, bold: true),
          columnOffset: 0,
        ),
        _iepFlexCell(
          columns,
          _IepPrintCell('训练内容', 3, bold: true),
          columnOffset: 1,
        ),
        pw.Expanded(
          flex: _iepColumnFlex(columns, 4, 6),
          child: pw.Column(
            crossAxisAlignment: pw.CrossAxisAlignment.stretch,
            children: <pw.Widget>[
              _iepPrintRow(
                  columns.sublist(4),
                  <_IepPrintCell>[
                    _IepPrintCell('完成情况', 6, bold: true),
                  ],
                  height: 24),
              _iepPrintRow(
                columns.sublist(4),
                List<_IepPrintCell>.generate(6, (int index) {
                  final String label =
                      index < dates.length ? _weekDateLabel(dates[index]) : '';
                  return _IepPrintCell(label, 1, bold: true);
                }),
                height: 24,
              ),
            ],
          ),
        ),
      ],
    ),
  );
}

class _IepPrintCell {
  const _IepPrintCell(
    this.text,
    this.columns, {
    this.bold = false,
    this.align = pw.TextAlign.center,
  });

  final String text;
  final int columns;
  final bool bold;
  final pw.TextAlign align;
}

pw.Widget _iepPrintTitle(String title) {
  return pw.Container(
    width: double.infinity,
    alignment: pw.Alignment.center,
    margin: const pw.EdgeInsets.only(bottom: 8),
    child: pw.Text(
      title,
      textAlign: pw.TextAlign.center,
      style: pw.TextStyle(
        color: _iepPrintInk,
        fontSize: 14,
        fontWeight: pw.FontWeight.bold,
        lineSpacing: 1.3,
      ),
    ),
  );
}

pw.Widget _iepPrintRow(
  List<int> columns,
  List<_IepPrintCell> cells, {
  double? height,
  double minHeight = 28,
}) {
  int columnOffset = 0;
  final List<pw.Widget> children = <pw.Widget>[];
  for (final _IepPrintCell cell in cells) {
    children.add(_iepFlexCell(columns, cell, columnOffset: columnOffset));
    columnOffset += cell.columns;
  }
  final pw.Widget row = pw.Row(
    crossAxisAlignment: pw.CrossAxisAlignment.stretch,
    children: children,
  );
  return pw.SizedBox(height: height ?? minHeight, child: row);
}

pw.Widget _iepFlexCell(
  List<int> columns,
  _IepPrintCell cell, {
  required int columnOffset,
}) {
  return pw.Expanded(
    flex: _iepColumnFlex(columns, columnOffset, cell.columns),
    child: _iepPrintCell(cell),
  );
}

int _iepColumnFlex(List<int> columns, int start, int span) {
  return columns.skip(start).take(span).fold<int>(0, (int sum, int item) {
    return sum + item;
  });
}

double _iepTextHeightEstimate(
  String text, {
  required double charsPerLine,
  double minHeight = 28,
}) {
  final List<String> lines =
      text.replaceAll('\r\n', '\n').replaceAll('\r', '\n').split('\n');
  int visualLines = 0;
  for (final String line in lines) {
    final int length = line.trim().isEmpty ? 1 : line.trim().length;
    visualLines += math.max(1, (length / charsPerLine).ceil());
  }
  return math.max(minHeight, visualLines * 10.2 + 10);
}

pw.Widget _iepPrintCell(_IepPrintCell cell) {
  final bool compactHeader = _isCompactPrintHeader(cell);
  return pw.Container(
    alignment: pw.Alignment.center,
    padding: compactHeader
        ? const pw.EdgeInsets.fromLTRB(5, 2, 5, 6)
        : const pw.EdgeInsets.symmetric(horizontal: 5, vertical: 4),
    decoration: pw.BoxDecoration(
      color: _iepPrintWhite,
      border: pw.Border.all(color: _iepPrintBorder, width: .55),
    ),
    child: pw.Text(
      cell.text.trim().isEmpty ? ' ' : cell.text,
      textAlign: cell.align,
      style: pw.TextStyle(
        color: cell.bold ? _iepPrintInk : _iepPrintText,
        fontSize: 7.8,
        fontWeight: cell.bold ? pw.FontWeight.bold : pw.FontWeight.normal,
        lineSpacing: compactHeader ? 0 : 1.1,
        height: compactHeader ? 0.94 : 1,
      ),
    ),
  );
}

bool _isCompactPrintHeader(_IepPrintCell cell) {
  if (!cell.bold) {
    return false;
  }
  final String text = cell.text.replaceAll(RegExp(r'\s+'), '');
  return text == '实施起止日期' || text == '课程形式' || text == '康复领域';
}
