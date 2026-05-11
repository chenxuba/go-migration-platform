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
  return <pw.Widget>[
    _IepSpanningPrintTable(
      columns: columns,
      rowHeights: rowHeights,
      mergedCells: <_IepSpanningMergedCell>[
        _IepSpanningMergedCell(
          startColumn: 0,
          columnSpan: 1,
          text: domain.domain,
          bold: true,
        ),
        _IepSpanningMergedCell(
          startColumn: 1,
          columnSpan: 3,
          text: domain.longGoals.join('\n'),
          align: pw.TextAlign.left,
        ),
      ],
      rowCells: shortGoals.map((goal) {
        return <_IepSpanningCell>[
          _IepSpanningCell(
            startColumn: 4,
            columnSpan: 2,
            text: goal.goal,
            align: pw.TextAlign.left,
          ),
          _IepSpanningCell(startColumn: 6, columnSpan: 1, text: goal.lesson),
          _IepSpanningCell(startColumn: 7, columnSpan: 1, text: goal.period),
        ];
      }).toList(),
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
  return <pw.Widget>[
    _IepSpanningPrintTable(
      columns: columns,
      rowHeights: rowHeights,
      mergedCells: <_IepSpanningMergedCell>[
        _IepSpanningMergedCell(
          startColumn: 0,
          columnSpan: 1,
          text: domain.domain,
          bold: true,
        ),
        _IepSpanningMergedCell(
          startColumn: 1,
          columnSpan: 2,
          text: domain.longGoal,
          align: pw.TextAlign.left,
        ),
        _IepSpanningMergedCell(
          startColumn: 3,
          columnSpan: 2,
          text: domain.shortGoal,
          align: pw.TextAlign.left,
        ),
        _IepSpanningMergedCell(
          startColumn: 9,
          columnSpan: 1,
          text: domain.lesson,
        ),
      ],
      rowCells: trainings.asMap().entries.map((entry) {
        return <_IepSpanningCell>[
          _IepSpanningCell(
            startColumn: 5,
            columnSpan: 4,
            text: entry.value.content,
            align: pw.TextAlign.left,
          ),
          _IepSpanningCell(
            startColumn: 10,
            columnSpan: 2,
            text: _monthTrainingPeriodText(monthRange, entry.key),
          ),
        ];
      }).toList(),
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

class _IepSpanningCell {
  const _IepSpanningCell({
    required this.startColumn,
    required this.columnSpan,
    required this.text,
    this.bold = false,
    this.align = pw.TextAlign.center,
  });

  final int startColumn;
  final int columnSpan;
  final String text;
  final bool bold;
  final pw.TextAlign align;
}

class _IepSpanningMergedCell extends _IepSpanningCell {
  const _IepSpanningMergedCell({
    required super.startColumn,
    required super.columnSpan,
    required super.text,
    super.bold = false,
    super.align = pw.TextAlign.center,
  });
}

class _IepSpanningPrintTableContext extends pw.WidgetContext {
  int firstRow = 0;
  int lastRow = 0;

  @override
  void apply(covariant _IepSpanningPrintTableContext other) {
    firstRow = other.firstRow;
    lastRow = other.lastRow;
  }

  @override
  _IepSpanningPrintTableContext clone() {
    return _IepSpanningPrintTableContext()..apply(this);
  }
}

class _IepSpanningPrintTable extends pw.Widget with pw.SpanningWidget {
  _IepSpanningPrintTable({
    required this.columns,
    required this.rowHeights,
    required this.mergedCells,
    required this.rowCells,
  });

  final List<int> columns;
  final List<double> rowHeights;
  final List<_IepSpanningMergedCell> mergedCells;
  final List<List<_IepSpanningCell>> rowCells;
  final _IepSpanningPrintTableContext _context =
      _IepSpanningPrintTableContext();

  @override
  bool get canSpan => true;

  @override
  bool get hasMoreWidgets => _context.lastRow < rowHeights.length;

  @override
  void restoreContext(covariant _IepSpanningPrintTableContext context) {
    _context.firstRow = context.lastRow;
    _context.lastRow = context.lastRow;
  }

  @override
  pw.WidgetContext saveContext() {
    return _context;
  }

  @override
  void layout(
    pw.Context context,
    pw.BoxConstraints constraints, {
    bool parentUsesSize = false,
  }) {
    final double width = constraints.hasBoundedWidth
        ? constraints.maxWidth
        : columns.fold<int>(0, (int sum, int item) => sum + item).toDouble();
    final int start = _context.firstRow.clamp(0, rowHeights.length);
    int end = start;
    double height = 0;
    final double maxHeight =
        constraints.hasBoundedHeight ? constraints.maxHeight : double.infinity;

    while (end < rowHeights.length) {
      final double rowHeight = rowHeights[end];
      if (height > 0 && height + rowHeight > maxHeight) {
        break;
      }
      height += rowHeight;
      end += 1;
      if (height >= maxHeight) {
        break;
      }
    }

    if (end == start && start < rowHeights.length) {
      height = math.min(rowHeights[start], maxHeight);
      end = start + 1;
    }

    _context.firstRow = start;
    _context.lastRow = end;
    box = PdfRect(0, 0, width, height);
  }

  @override
  void paint(pw.Context context) {
    super.paint(context);
    if (box == null || _context.lastRow <= _context.firstRow) {
      return;
    }
    final Matrix4 transform = Matrix4.identity()..translate(box!.x, box!.y);
    context.canvas
      ..saveContext()
      ..setTransform(transform);

    final List<double> xPositions = _columnPositions(box!.width);
    final int start = _context.firstRow;
    final int end = _context.lastRow;
    final double segmentHeight = _segmentHeight(start, end);

    for (final _IepSpanningMergedCell cell in mergedCells) {
      final PdfRect rect = _cellRect(
        xPositions,
        cell.startColumn,
        cell.columnSpan,
        0,
        segmentHeight,
      );
      _drawBorder(context, rect);
      _paintTextInRect(
        context,
        rect,
        start == 0 ? cell.text : '',
        bold: cell.bold,
        align: cell.align,
      );
    }

    double topCursor = segmentHeight;
    for (int row = start; row < end; row += 1) {
      final double rowHeight = rowHeights[row];
      final double y = topCursor - rowHeight;
      for (final _IepSpanningCell cell in rowCells[row]) {
        final PdfRect rect = _cellRect(
          xPositions,
          cell.startColumn,
          cell.columnSpan,
          y,
          rowHeight,
        );
        _drawBorder(context, rect);
        _paintTextInRect(
          context,
          rect,
          cell.text,
          bold: cell.bold,
          align: cell.align,
        );
      }
      topCursor = y;
    }

    context.canvas.restoreContext();
  }

  List<double> _columnPositions(double width) {
    final int total = columns.fold<int>(0, (int sum, int item) => sum + item);
    double cursor = 0;
    final List<double> positions = <double>[0];
    for (final int column in columns) {
      cursor += width * column / total;
      positions.add(cursor);
    }
    return positions;
  }

  PdfRect _cellRect(
    List<double> xPositions,
    int startColumn,
    int columnSpan,
    double y,
    double height,
  ) {
    final double x = xPositions[startColumn];
    final double right = xPositions[startColumn + columnSpan];
    return PdfRect(x, y, right - x, height);
  }

  double _segmentHeight(int start, int end) {
    return rowHeights
        .sublist(start, end)
        .fold<double>(0, (double sum, double item) => sum + item);
  }

  void _drawBorder(pw.Context context, PdfRect rect) {
    context.canvas
      ..setStrokeColor(_iepPrintBorder)
      ..setLineWidth(.55)
      ..drawBox(rect)
      ..strokePath();
  }

  void _paintTextInRect(
    pw.Context context,
    PdfRect rect,
    String text, {
    required bool bold,
    required pw.TextAlign align,
  }) {
    final pw.Widget child = pw.Container(
      alignment: align == pw.TextAlign.left
          ? pw.Alignment.centerLeft
          : pw.Alignment.center,
      padding: const pw.EdgeInsets.symmetric(horizontal: 5, vertical: 4),
      child: pw.Text(
        text.trim().isEmpty ? ' ' : text,
        textAlign: align,
        style: pw.TextStyle(
          color: bold ? _iepPrintInk : _iepPrintText,
          fontSize: 7.8,
          fontWeight: bold ? pw.FontWeight.bold : pw.FontWeight.normal,
          lineSpacing: 1.1,
          height: 1,
        ),
      ),
    );
    child.layout(
      context,
      pw.BoxConstraints.tightFor(width: rect.width, height: rect.height),
    );
    child.box = rect;
    child.paint(context);
  }
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
