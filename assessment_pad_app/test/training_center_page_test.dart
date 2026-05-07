import 'package:assessment_pad_app/main.dart' show PadViewport;
import 'package:assessment_pad_app/training_center_page.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets(
      'training center page renders compact pad layout without overflow',
      (WidgetTester tester) async {
    tester.view.physicalSize = const Size(1024, 768);
    tester.view.devicePixelRatio = 1;
    addTearDown(() {
      tester.view.resetPhysicalSize();
      tester.view.resetDevicePixelRatio();
    });

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: PadViewport(
            child: TrainingCenterPage(onBack: () {}),
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('训练中心'), findsOneWidget);
    expect(find.text('推荐训练游戏'), findsOneWidget);
    expect(find.text('游戏合集'), findsOneWidget);
    expect(find.text('能力雷达'), findsNothing);
  });
}
