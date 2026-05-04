import 'package:assessment_pad_app/auth_client.dart';
import 'package:assessment_pad_app/main.dart';
import 'package:flutter/foundation.dart';
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

    await _enterWithCustomKeyboard(tester, 0, 'chenrui');
    await _enterWithCustomKeyboard(tester, 1, '123456');
    await tester.tap(find.text('登 录'));
    await tester.pumpAndSettle();

    expect(find.text('上午好，启明成长中心'), findsOneWidget);
    expect(find.text('开始测评'), findsOneWidget);
    expect(find.byIcon(Icons.search_rounded), findsOneWidget);

    await tester.binding.handlePopRoute();
    await tester.pumpAndSettle();

    expect(find.text('上午好，启明成长中心'), findsOneWidget);
    expect(find.text('机构账号登录'), findsNothing);
  });

  testWidgets('login page redirects to home when token already exists',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{
      'auth_token': 'existing-token',
    });
    await tester.pumpWidget(AssessmentPadApp(authClient: _FakeAuthClient()));
    await tester.pumpAndSettle();

    expect(find.text('上午好，启明成长中心'), findsOneWidget);
    expect(find.text('机构账号登录'), findsNothing);
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

    await _enterWithCustomKeyboard(tester, 0, 'chenrui');
    await _enterWithCustomKeyboard(tester, 1, '123456');
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

  testWidgets('wrong password does not open institution picker',
      (WidgetTester tester) async {
    SharedPreferences.setMockInitialValues(<String, Object>{});
    await tester.pumpWidget(
      AssessmentPadApp(authClient: _PasswordCheckingAuthClient()),
    );

    await _enterWithCustomKeyboard(tester, 0, 'chenrui');
    await _enterWithCustomKeyboard(tester, 1, 'wrong123');
    await tester.tap(find.text('登 录'));
    await tester.pumpAndSettle();

    expect(find.text('登录失败,用户名或密码错误'), findsOneWidget);
    expect(find.text('选择登录机构'), findsNothing);
  });

  testWidgets('desktop login fields use native input without custom keyboard',
      (WidgetTester tester) async {
    debugDefaultTargetPlatformOverride = TargetPlatform.macOS;
    try {
      SharedPreferences.setMockInitialValues(<String, Object>{});
      await tester.pumpWidget(AssessmentPadApp(authClient: _FakeAuthClient()));

      await tester.tap(find.byType(TextField).first);
      await tester.pumpAndSettle();

      expect(find.byKey(const ValueKey<String>('login-key-1')), findsNothing);

      await tester.enterText(find.byType(TextField).first, 'chenrui');
      await tester.pumpAndSettle();

      expect(find.text('chenrui'), findsOneWidget);
    } finally {
      debugDefaultTargetPlatformOverride = null;
    }
  });
}

Future<void> _enterWithCustomKeyboard(
  WidgetTester tester,
  int fieldIndex,
  String value,
) async {
  await tester.tap(find.byType(TextField).at(fieldIndex));
  await tester.pumpAndSettle();

  for (final String character in value.split('')) {
    await tester.tap(find.byKey(ValueKey<String>('login-key-$character')));
    await tester.pump(const Duration(milliseconds: 20));
  }
}

class _FakeAuthClient implements AuthClient {
  @override
  Uri buildQrLoginUri(String nonce) {
    return Uri.parse('https://example.com/qr?nonce=$nonce');
  }

  @override
  Future<List<InstitutionLoginOption>> listInstitutionOptions(
    String identifier, {
    String password = '',
  }) async {
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
    String identifier, {
    String password = '',
  }) async {
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

class _PasswordCheckingAuthClient extends _MultiInstitutionAuthClient {
  @override
  Future<List<InstitutionLoginOption>> listInstitutionOptions(
    String identifier, {
    String password = '',
  }) async {
    if (password != '123456') {
      throw const AuthException('登录失败,用户名或密码错误');
    }
    return super.listInstitutionOptions(identifier, password: password);
  }
}
