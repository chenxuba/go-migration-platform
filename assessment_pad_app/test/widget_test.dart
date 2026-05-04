import 'package:assessment_pad_app/main.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('login page opens the home dashboard',
      (WidgetTester tester) async {
    await tester.pumpWidget(const AssessmentPadApp());

    expect(find.text('测评云端'), findsOneWidget);
    expect(find.text('机构账号登录'), findsOneWidget);

    await tester.tap(find.text('登 录'));
    await tester.pumpAndSettle();

    expect(find.text('上午好，启明成长中心'), findsOneWidget);
    expect(find.text('开始测评'), findsOneWidget);
    expect(find.byIcon(Icons.search_rounded), findsOneWidget);
  });
}
