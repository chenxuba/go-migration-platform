import 'package:assessment_pad_app/auth_client.dart';
import 'package:assessment_pad_app/main.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  testWidgets('login page opens the home dashboard after real login callback',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{});
    await tester.pumpWidget(AssessmentPadApp(authClient: _FakeAuthClient()));

    expect(find.text('测评云端'), findsOneWidget);
    expect(find.text('机构账号登录'), findsOneWidget);
    expect(find.text('验证码登录'), findsNothing);

    await tester.enterText(find.byType(TextField).at(0), 'chenrui');
    await tester.enterText(find.byType(TextField).at(1), '123456');
    await tester.tap(find.text('登 录'));
    await tester.pumpAndSettle();

    expect(find.text('上午好，启明成长中心'), findsOneWidget);
    expect(find.text('开始测评'), findsOneWidget);
    expect(find.byIcon(Icons.search_rounded), findsOneWidget);
  });

  testWidgets('login page switches to qr login and back',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{});
    await tester.pumpWidget(AssessmentPadApp(authClient: _FakeAuthClient()));

    await tester.tap(find.text('二维码登录'));
    await tester.pumpAndSettle();

    expect(find.text('二维码登录'), findsOneWidget);
    expect(find.text('账号登录'), findsOneWidget);
    expect(find.text('刷新二维码'), findsOneWidget);

    await tester.tap(find.text('账号登录'));
    await tester.pumpAndSettle();

    expect(find.text('机构账号登录'), findsOneWidget);
    expect(find.text('忘记密码'), findsOneWidget);
  });

  testWidgets('login page shows styled institution picker for multiple options',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{});
    await tester.pumpWidget(
      AssessmentPadApp(authClient: _MultiInstitutionAuthClient()),
    );

    await tester.enterText(find.byType(TextField).at(0), 'chenrui');
    await tester.enterText(find.byType(TextField).at(1), '123456');
    await tester.tap(find.text('登 录'));
    await tester.pumpAndSettle();

    expect(find.text('选择登录机构'), findsOneWidget);
    expect(find.text('当前账号关联多个机构，请选择本次进入的后台'), findsOneWidget);
    expect(find.text('启明成长中心'), findsOneWidget);
    expect(find.text('南山训练中心'), findsOneWidget);
    expect(find.text('超管'), findsOneWidget);
    expect(find.text('正常'), findsOneWidget);

    await tester.tap(find.text('南山训练中心'));
    await tester.pumpAndSettle();

    expect(find.text('上午好，启明成长中心'), findsOneWidget);
  });
}

class _FakeAuthClient implements AuthClient {
  @override
  Uri buildQrLoginUri(String nonce) {
    return Uri.parse('https://example.com/qr?nonce=$nonce');
  }

  @override
  Future<List<InstitutionLoginOption>> listInstitutionOptions(
    String identifier,
  ) async {
    return <InstitutionLoginOption>[];
  }

  @override
  Future<LoginResult> login({
    required String username,
    required String password,
    InstitutionLoginOption? institution,
  }) async {
    return <LoginResult>[
      LoginResult(
        token: 'fake-token',
        loginType: 'org',
        tenantId: 'tenant-a',
        orgId: 1,
        raw: <String, dynamic>{
          'token': 'fake-token',
          'loginType': 'org',
          'tenantId': 'tenant-a',
          'orgId': 1,
        },
      ),
    ].first;
  }
}

class _MultiInstitutionAuthClient extends _FakeAuthClient {
  @override
  Future<List<InstitutionLoginOption>> listInstitutionOptions(
    String identifier,
  ) async {
    return const <InstitutionLoginOption>[
      InstitutionLoginOption(
        userId: 1,
        instId: 11,
        orgName: '启明成长中心',
        loginName: 'chenrui',
        nickName: '陈老师',
        mobile: '19900000001',
        admin: true,
        institutionReadonly: false,
        institutionStatus: 'normal',
      ),
      InstitutionLoginOption(
        userId: 2,
        instId: 12,
        orgName: '南山训练中心',
        loginName: 'chenrui',
        nickName: '陈老师',
        mobile: '19900000002',
        admin: false,
        institutionReadonly: false,
        institutionStatus: 'warning',
      ),
    ];
  }
}
