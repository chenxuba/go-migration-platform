import 'dart:math' as math;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'assessment_scale_client.dart';
import 'assessment_scale_category_page.dart';
import 'assessment_report_list_page.dart';
import 'auth_client.dart';
import 'erxin_assessment_client.dart';
import 'erxin_assessment_page.dart';
import 'home_client.dart';
import 'pep3_assessment_client.dart';
import 'pep3_assessment_page.dart';
import 'pad_responsive.dart';
import 'route_bootstrap.dart';
import 'smart_timetable_page.dart';
import 'timetable_client.dart';
import 'training_center_page.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await SystemChrome.setPreferredOrientations(
    <DeviceOrientation>[
      DeviceOrientation.landscapeLeft,
      DeviceOrientation.landscapeRight,
    ],
  );
  await SystemChrome.setEnabledSystemUIMode(SystemUiMode.immersiveSticky);
  SystemChrome.setSystemUIOverlayStyle(_immersiveOverlayStyle);
  runApp(const AssessmentPadApp());
  WidgetsBinding.instance.addPostFrameCallback((_) {
    SystemChrome.setEnabledSystemUIMode(SystemUiMode.immersiveSticky);
    SystemChrome.setSystemUIOverlayStyle(_immersiveOverlayStyle);
  });
}

const SystemUiOverlayStyle _immersiveOverlayStyle = SystemUiOverlayStyle(
  statusBarColor: Colors.transparent,
  statusBarIconBrightness: Brightness.dark,
  statusBarBrightness: Brightness.light,
  systemStatusBarContrastEnforced: false,
  systemNavigationBarColor: Colors.transparent,
  systemNavigationBarDividerColor: Colors.transparent,
  systemNavigationBarIconBrightness: Brightness.dark,
  systemNavigationBarContrastEnforced: false,
);

class AssessmentPadApp extends StatelessWidget {
  const AssessmentPadApp({
    this.authClient = const IamAuthClient(),
    this.homeClient = const ApiHomeClient(),
    this.scaleClient = const ApiAssessmentScaleClient(),
    this.pep3Client = const ApiPep3AssessmentClient(),
    this.erxinClient = const ApiErxinAssessmentClient(),
    this.timetableClient = const ApiTimetableClient(),
    super.key,
  });

  final AuthClient authClient;
  final HomeClient homeClient;
  final AssessmentScaleClient scaleClient;
  final Pep3AssessmentClient pep3Client;
  final ErxinAssessmentClient erxinClient;
  final TimetableClient timetableClient;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: '评估助手',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        useMaterial3: true,
        scaffoldBackgroundColor: AppColors.page,
        fontFamily: 'PingFang SC',
        fontFamilyFallback: const <String>[
          'Microsoft YaHei',
          'Heiti SC',
          'Arial',
        ],
        colorScheme: ColorScheme.fromSeed(seedColor: AppColors.orange),
      ),
      onGenerateRoute: _buildRoute,
    );
  }

  Route<dynamic> _buildRoute(RouteSettings settings) {
    switch (settings.name) {
      case '/':
        return MaterialPageRoute<void>(
          settings: settings,
          builder: (_) => LoginPage(authClient: authClient),
        );
      case '/home':
        return MaterialPageRoute<void>(
          settings: settings,
          builder: (_) => HomePage(homeClient: homeClient),
        );
      case '/smart-timetable':
        return MaterialPageRoute<void>(
          settings: settings,
          builder: (_) => SmartTimetablePage(timetableClient: timetableClient),
        );
      case '/training-center':
        return MaterialPageRoute<void>(
          settings: settings,
          builder: (BuildContext context) => Scaffold(
            body: PadViewport(
              child: TrainingCenterPage(
                onBack: () => Navigator.of(context).maybePop(),
              ),
            ),
          ),
        );
      case '/assessment-reports':
        return MaterialPageRoute<void>(
          settings: settings,
          builder: (BuildContext context) => Scaffold(
            body: PadViewport(
              child: AssessmentReportListScreen(
                scaleClient: scaleClient,
                recordClient: pep3Client,
                onBack: () => Navigator.of(context).maybePop(),
              ),
            ),
          ),
        );
      case '/assessment-scales':
        return MaterialPageRoute<void>(
          settings: settings,
          builder: (BuildContext context) => Scaffold(
            body: PadViewport(
              child: AssessmentScaleCategoryScreen(
                scaleClient: scaleClient,
                onBack: () => Navigator.of(context).maybePop(),
              ),
            ),
          ),
        );
      case '/pep3-assessment':
        final Object? rawArgs = settings.arguments;
        final Pep3AssessmentLaunchArgs args =
            rawArgs is Pep3AssessmentLaunchArgs
                ? rawArgs
                : const Pep3AssessmentLaunchArgs();
        return MaterialPageRoute<void>(
          settings: settings,
          builder: (BuildContext context) => Scaffold(
            body: PadViewport(
              child: Pep3AssessmentPage(
                args: args,
                client: pep3Client,
                homeClient: homeClient,
                onBack: () => Navigator.of(context).maybePop(),
              ),
            ),
          ),
        );
      case '/erxin-assessment':
        final Object? rawArgs = settings.arguments;
        final ErxinAssessmentLaunchArgs args =
            rawArgs is ErxinAssessmentLaunchArgs
                ? rawArgs
                : const ErxinAssessmentLaunchArgs();
        return MaterialPageRoute<void>(
          settings: settings,
          builder: (BuildContext context) => Scaffold(
            resizeToAvoidBottomInset: false,
            body: PadViewport(
              child: ErxinAssessmentPage(
                args: args,
                client: erxinClient,
                onBack: () => Navigator.of(context).maybePop(),
              ),
            ),
          ),
        );
      default:
        return MaterialPageRoute<void>(
          settings: settings,
          builder: (_) => LoginPage(authClient: authClient),
        );
    }
  }
}

class AppColors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color pageLight = Color(0xFFFFFBF4);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95735);
  static const Color orangeSoft = Color(0xFFFFE1CF);
  static const Color yellow = Color(0xFFF6C45F);
  static const Color green = Color(0xFF8DB886);
  static const Color greenSoft = Color(0xFFEAF4E5);
  static const Color blueGray = Color(0xFFAAB7C3);
  static const Color ink = Color(0xFF432B22);
  static const Color body = Color(0xFF7F665A);
  static const Color muted = Color(0xFFBBA99C);
  static const Color line = Color(0xFFF0DACB);
  static const Color card = Color(0xFFFFFEFB);
}

const double _designHeight = 768;
const double _minDesignWidth = 1024;
const double _wideDesignWidth = 1366;
const double _loginLeftShiftCompact = 24;
const double _loginLeftShiftWide = 64;
const double _loginCardWidth = 408;
const double _loginCardHeight = 548;
const double _loginCardRight = 64;
const double _loginCardTop = 162;
const double _loginKeyboardWidth = 540;
const double _loginKeyboardHeight = 408;
const double _loginKeyboardKeyHeight = 54;
const double _homeMainTop = 124;
const double _homeStartHeight = 334;
const double _homeStatsHeight = 102;
const double _homeStatsScheduleGap = 16;
const double _homeMainShortcutGap = 16;
const double _homeScheduleTop =
    _homeMainTop + _homeStatsHeight + _homeStatsScheduleGap;
const double _homeScheduleHeight =
    _homeMainTop + _homeStartHeight - _homeScheduleTop;
const double _homeShortcutTop =
    _homeMainTop + _homeStartHeight + _homeMainShortcutGap;
const String _authTokenStorageKey = 'auth_token';
const String _authLoginTypeStorageKey = 'auth_login_type';
const String _authTenantIdStorageKey = 'auth_tenant_id';
const String _authOrgIdStorageKey = 'auth_org_id';
const String _defaultLoginUsername = String.fromEnvironment(
  'DEFAULT_LOGIN_USERNAME',
  defaultValue: '',
);
const String _defaultLoginPassword = String.fromEnvironment(
  'DEFAULT_LOGIN_PASSWORD',
  defaultValue: '',
);

List<BoxShadow> _softShadow({
  Color color = const Color(0x22000000),
  double blur = 28,
  Offset offset = const Offset(0, 14),
}) {
  return <BoxShadow>[
    BoxShadow(color: color, blurRadius: blur, offset: offset),
  ];
}

bool get _usesCustomLoginKeyboard {
  switch (defaultTargetPlatform) {
    case TargetPlatform.android:
    case TargetPlatform.iOS:
      return true;
    case TargetPlatform.fuchsia:
    case TargetPlatform.linux:
    case TargetPlatform.macOS:
    case TargetPlatform.windows:
      return false;
  }
}

class PadViewport extends StatelessWidget {
  const PadViewport({required this.child, super.key});

  final Widget child;

  @override
  Widget build(BuildContext context) {
    return AnnotatedRegion<SystemUiOverlayStyle>(
      value: _immersiveOverlayStyle,
      child: LayoutBuilder(
        builder: (BuildContext context, BoxConstraints constraints) {
          final Size screenSize = MediaQuery.sizeOf(context);
          final double viewportWidth = constraints.maxWidth.isFinite
              ? constraints.maxWidth
              : screenSize.width;
          final double viewportHeight = constraints.maxHeight.isFinite
              ? constraints.maxHeight
              : screenSize.height;
          final double aspect = viewportWidth / viewportHeight;
          final double designWidth =
              math.max(_minDesignWidth, _designHeight * aspect);

          return ColoredBox(
            color: AppColors.page,
            child: Center(
              child: FittedBox(
                fit: BoxFit.contain,
                child: SizedBox(
                  width: designWidth,
                  height: _designHeight,
                  child: child,
                ),
              ),
            ),
          );
        },
      ),
    );
  }
}

class LoginPage extends StatefulWidget {
  const LoginPage({required this.authClient, super.key});

  final AuthClient authClient;

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  @override
  void initState() {
    super.initState();
    _redirectIfLoggedIn();
  }

  Future<void> _redirectIfLoggedIn() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_authTokenStorageKey) ?? '';
    if (!mounted || token.trim().isEmpty) {
      return;
    }
    Navigator.of(context).pushNamedAndRemoveUntil(
      '/home',
      (Route<dynamic> route) => false,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: PadViewport(child: LoginScreen(authClient: widget.authClient)),
    );
  }
}

class HomePage extends StatelessWidget {
  const HomePage({required this.homeClient, super.key});

  final HomeClient homeClient;

  @override
  Widget build(BuildContext context) {
    return PopScope(
      canPop: false,
      child: Scaffold(
          body: PadViewport(child: HomeScreen(homeClient: homeClient))),
    );
  }
}

class LoginScreen extends StatelessWidget {
  const LoginScreen({required this.authClient, super.key});

  final AuthClient authClient;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width = constraints.maxWidth.isFinite
            ? constraints.maxWidth
            : _wideDesignWidth;
        final double leftShift =
            width < 1200 ? _loginLeftShiftCompact : _loginLeftShiftWide;
        return Stack(
          children: <Widget>[
            CustomPaint(
                size: Size(width, _designHeight),
                painter: LoginBackgroundPainter()),
            const Positioned(top: 70, right: 70, child: PadExperienceChip()),
            Positioned(
              left: 92 + leftShift,
              top: 170,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Row(
                    children: <Widget>[
                      CustomPaint(
                        size: const Size(72, 56),
                        painter: CloudCheckPainter(color: AppColors.orange),
                      ),
                      const SizedBox(width: 16),
                      const Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: <Widget>[
                          Text(
                            '评估助手',
                            style: TextStyle(
                              color: AppColors.ink,
                              fontSize: 40,
                              fontWeight: FontWeight.w800,
                              height: 1,
                            ),
                          ),
                          SizedBox(height: 12),
                          Text(
                            '机构测评与服务工作台',
                            style: TextStyle(
                              color: AppColors.body,
                              fontSize: 18,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                  const SizedBox(height: 42),
                  const Text(
                    '专业测评，科学赋能成长',
                    style: TextStyle(
                      color: AppColors.orange,
                      fontSize: 17,
                      fontWeight: FontWeight.w500,
                      letterSpacing: 1.2,
                    ),
                  ),
                ],
              ),
            ),
            Positioned(
              left: 85 + leftShift,
              top: 320,
              child: const LoginIllustration(),
            ),
            Positioned.fill(
              child: LoginCard(
                authClient: authClient,
                onLoginSuccess: () => Navigator.of(context)
                    .pushNamedAndRemoveUntil('/home', (_) => false),
              ),
            ),
          ],
        );
      },
    );
  }
}

class LoginBackgroundPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint base = Paint()..color = AppColors.pageLight;
    canvas.drawRect(Offset.zero & size, base);

    final Path topPanel = Path()
      ..moveTo(0, 0)
      ..lineTo(size.width * .44, 0)
      ..quadraticBezierTo(size.width * .36, 118, size.width * .16, 170)
      ..quadraticBezierTo(size.width * .07, 198, 0, 210)
      ..close();
    canvas.drawPath(topPanel, Paint()..color = const Color(0xFFFFEBCB));

    final Path leftPanel = Path()
      ..moveTo(0, 0)
      ..lineTo(size.width * .18, 0)
      ..quadraticBezierTo(size.width * .04, 150, 0, 320)
      ..close();
    canvas.drawPath(leftPanel, Paint()..color = const Color(0xFFFFF1D7));

    final Path footer = Path()
      ..moveTo(0, size.height - 72)
      ..quadraticBezierTo(150, size.height - 116, 278, size.height - 58)
      ..quadraticBezierTo(390, size.height - 6, 542, size.height - 40)
      ..lineTo(size.width, size.height - 92)
      ..lineTo(size.width, size.height)
      ..lineTo(0, size.height)
      ..close();
    canvas.drawPath(footer, Paint()..color = const Color(0xFFFFF0D9));

    canvas.drawCircle(
      Offset(size.width - 78, size.height - 72),
      130,
      Paint()..color = const Color(0xFFFFE4AE).withOpacity(.46),
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class PadExperienceChip extends StatelessWidget {
  const PadExperienceChip({super.key});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.72),
        borderRadius: BorderRadius.circular(26),
        boxShadow: _softShadow(color: const Color(0x18C96D34), blur: 18),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Container(
            width: 25,
            height: 25,
            decoration: const BoxDecoration(
              color: Color(0xFFFFF1DE),
              shape: BoxShape.circle,
            ),
            child: const Icon(
              Icons.tablet_mac_rounded,
              color: AppColors.orange,
              size: 15,
            ),
          ),
          const SizedBox(width: 8),
          const Text(
            'Pad 端专属体验',
            style: TextStyle(
              color: AppColors.orange,
              fontSize: 15,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }
}

enum _LoginInputTarget { username, password }

class LoginCard extends StatefulWidget {
  const LoginCard({
    required this.authClient,
    required this.onLoginSuccess,
    super.key,
  });

  final AuthClient authClient;
  final VoidCallback onLoginSuccess;

  @override
  State<LoginCard> createState() => _LoginCardState();
}

class _LoginCardState extends State<LoginCard> {
  static const String _rememberPasswordKey = 'login_remember_password';
  static const String _rememberedUsernameKey = 'login_remember_username';
  static const String _rememberedPasswordKey = 'login_remember_password_value';
  static const String _sessionTokenKey = _authTokenStorageKey;
  static const String _sessionLoginTypeKey = _authLoginTypeStorageKey;
  static const String _sessionTenantIdKey = _authTenantIdStorageKey;
  static const String _sessionOrgIdKey = _authOrgIdStorageKey;

  final TextEditingController _usernameController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  final FocusNode _usernameFocusNode = FocusNode();
  final FocusNode _passwordFocusNode = FocusNode();

  bool _rememberPassword = false;
  bool _passwordVisible = false;
  bool _loading = false;
  bool _qrMode = false;
  bool _customKeyboardVisible = false;
  bool _keyboardShift = false;
  bool _keyboardWasMoved = false;
  _LoginInputTarget _activeInput = _LoginInputTarget.username;
  Offset _keyboardOffset = const Offset(360, 198);
  String _qrNonce = DateTime.now().millisecondsSinceEpoch.toString();
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _restoreRememberedCredentials();
  }

  @override
  void dispose() {
    _usernameController.dispose();
    _passwordController.dispose();
    _usernameFocusNode.dispose();
    _passwordFocusNode.dispose();
    super.dispose();
  }

  Future<void> _restoreRememberedCredentials() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final bool remember = prefs.getBool(_rememberPasswordKey) ?? false;
    if (!mounted) {
      return;
    }
    setState(() {
      _rememberPassword = remember;
      if (remember) {
        _usernameController.text =
            prefs.getString(_rememberedUsernameKey) ?? '';
        _passwordController.text =
            prefs.getString(_rememberedPasswordKey) ?? '';
      } else {
        _usernameController.text = _defaultLoginUsername;
        _passwordController.text = _defaultLoginPassword;
      }
    });
  }

  Future<void> _persistRememberedCredentials() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    await prefs.setBool(_rememberPasswordKey, _rememberPassword);
    if (_rememberPassword) {
      await prefs.setString(
        _rememberedUsernameKey,
        _usernameController.text.trim(),
      );
      await prefs.setString(_rememberedPasswordKey, _passwordController.text);
      return;
    }
    await prefs.remove(_rememberedUsernameKey);
    await prefs.remove(_rememberedPasswordKey);
  }

  Future<void> _persistSession(LoginResult result) async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    await prefs.setString(_sessionTokenKey, result.token);
    await prefs.setString(_sessionLoginTypeKey, result.loginType);
    await prefs.setString(_sessionTenantIdKey, result.tenantId);
    if (result.orgId != null) {
      await prefs.setInt(_sessionOrgIdKey, result.orgId!);
    } else {
      await prefs.remove(_sessionOrgIdKey);
    }
  }

  void _clearError() {
    if (_errorMessage == null) {
      return;
    }
    setState(() => _errorMessage = null);
  }

  void _dismissKeyboard({bool hideCustomKeyboard = true}) {
    FocusManager.instance.primaryFocus?.unfocus();
    SystemChannels.textInput.invokeMethod<void>('TextInput.hide');
    if (hideCustomKeyboard && _customKeyboardVisible && mounted) {
      setState(() => _customKeyboardVisible = false);
    }
  }

  void _openCustomKeyboard(_LoginInputTarget target) {
    if (!_usesCustomLoginKeyboard) {
      setState(() {
        _activeInput = target;
        _customKeyboardVisible = false;
        _errorMessage = null;
      });
      if (target == _LoginInputTarget.username) {
        _usernameFocusNode.requestFocus();
      } else {
        _passwordFocusNode.requestFocus();
      }
      return;
    }

    final Size stageSize = _keyboardStageSize();
    setState(() {
      _activeInput = target;
      _customKeyboardVisible = true;
      if (!_keyboardWasMoved) {
        _keyboardOffset = _defaultKeyboardOffset(stageSize);
      } else {
        _keyboardOffset = _clampKeyboardOffset(_keyboardOffset, stageSize);
      }
      _errorMessage = null;
    });
    if (target == _LoginInputTarget.username) {
      _usernameFocusNode.requestFocus();
    } else {
      _passwordFocusNode.requestFocus();
    }
    SystemChannels.textInput.invokeMethod<void>('TextInput.hide');
  }

  void _moveKeyboard(DragUpdateDetails details) {
    final Offset nextOffset = _clampKeyboardOffset(
      _keyboardOffset + details.delta,
      _keyboardStageSize(),
    );
    setState(() {
      _keyboardWasMoved = true;
      _keyboardOffset = nextOffset;
    });
  }

  Size _keyboardStageSize() {
    return context.size ?? const Size(_wideDesignWidth, _designHeight);
  }

  Offset _defaultKeyboardOffset(Size stageSize) {
    final double cardLeft = stageSize.width - _loginCardRight - _loginCardWidth;
    return _clampKeyboardOffset(
      Offset(cardLeft - _loginKeyboardWidth - 36, _loginCardTop + 34),
      stageSize,
    );
  }

  Offset _clampKeyboardOffset(Offset offset, Size stageSize) {
    final double maxDx = math.max(0, stageSize.width - _loginKeyboardWidth);
    final double maxDy = math.max(0, stageSize.height - _loginKeyboardHeight);
    return Offset(
      offset.dx.clamp(0.0, maxDx),
      offset.dy.clamp(0.0, maxDy),
    );
  }

  TextEditingController get _activeController {
    return _activeInput == _LoginInputTarget.username
        ? _usernameController
        : _passwordController;
  }

  void _insertKeyboardText(String value) {
    final TextEditingController controller = _activeController;
    final TextSelection selection = controller.selection;
    final String text = controller.text;
    final int start = selection.isValid ? selection.start : text.length;
    final int end = selection.isValid ? selection.end : text.length;
    final String nextText = text.replaceRange(start, end, value);
    controller.value = TextEditingValue(
      text: nextText,
      selection: TextSelection.collapsed(offset: start + value.length),
    );
    _clearError();
  }

  void _deleteKeyboardText() {
    final TextEditingController controller = _activeController;
    final TextSelection selection = controller.selection;
    final String text = controller.text;
    if (text.isEmpty) {
      return;
    }

    if (selection.isValid && !selection.isCollapsed) {
      controller.value = TextEditingValue(
        text: text.replaceRange(selection.start, selection.end, ''),
        selection: TextSelection.collapsed(offset: selection.start),
      );
      return;
    }

    final int cursor = selection.isValid ? selection.start : text.length;
    if (cursor <= 0) {
      return;
    }
    controller.value = TextEditingValue(
      text: text.replaceRange(cursor - 1, cursor, ''),
      selection: TextSelection.collapsed(offset: cursor - 1),
    );
    _clearError();
  }

  void _clearActiveInput() {
    _activeController.clear();
    _clearError();
  }

  void _focusNextInput() {
    if (_activeInput == _LoginInputTarget.username) {
      _openCustomKeyboard(_LoginInputTarget.password);
      return;
    }
    _submit();
  }

  Future<void> _submit() async {
    if (_loading) {
      return;
    }
    _dismissKeyboard();
    final String username = _usernameController.text.trim();
    final String password = _passwordController.text;
    if (username.isEmpty || password.trim().isEmpty) {
      setState(() => _errorMessage = '请输入账号和密码');
      return;
    }

    setState(() {
      _loading = true;
      _errorMessage = null;
    });

    try {
      final List<InstitutionLoginOption> options =
          await widget.authClient.listInstitutionOptions(
        username,
        password: password,
      );
      InstitutionLoginOption? selectedInstitution;
      if (options.length > 1) {
        if (mounted) {
          setState(() => _loading = false);
        }
        selectedInstitution = await _chooseInstitution(options);
        if (selectedInstitution == null) {
          return;
        }
        if (mounted) {
          setState(() => _loading = true);
        }
      } else if (options.length == 1) {
        selectedInstitution = options.first;
      }

      final LoginResult result = await widget.authClient.login(
        username: username,
        password: password,
        institution: selectedInstitution,
      );
      await _persistSession(result);
      await _persistRememberedCredentials();
      if (!mounted) {
        return;
      }
      _dismissKeyboard();
      widget.onLoginSuccess();
    } on AuthException catch (error) {
      if (mounted) {
        setState(() => _errorMessage = error.message);
      }
    } on Object catch (error) {
      if (mounted) {
        setState(() => _errorMessage = '登录失败：$error');
      }
    } finally {
      if (mounted) {
        setState(() => _loading = false);
      }
    }
  }

  Future<InstitutionLoginOption?> _chooseInstitution(
    List<InstitutionLoginOption> options,
  ) {
    _dismissKeyboard();
    return showDialog<InstitutionLoginOption>(
      context: context,
      builder: (BuildContext context) {
        return PadDialogViewport(
          child: InstitutionPickerDialog(options: options),
        );
      },
    );
  }

  void _showForgotPasswordHint() {
    _dismissKeyboard();
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('请联系机构管理员重置密码')),
    );
  }

  void _refreshQrCode() {
    setState(() {
      _qrNonce = DateTime.now().microsecondsSinceEpoch.toString();
    });
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        return Stack(
          clipBehavior: Clip.none,
          children: <Widget>[
            Positioned(
              right: _loginCardRight,
              top: _loginCardTop,
              child: _buildCard(),
            ),
            if (_customKeyboardVisible && !_qrMode)
              Positioned(
                left: _keyboardOffset.dx,
                top: _keyboardOffset.dy,
                child: FloatingLoginKeyboard(
                  targetLabel:
                      _activeInput == _LoginInputTarget.username ? '账号' : '密码',
                  shifted: _keyboardShift,
                  onDragUpdate: _moveKeyboard,
                  onKey: _insertKeyboardText,
                  onBackspace: _deleteKeyboardText,
                  onClear: _clearActiveInput,
                  onShift: () {
                    setState(() => _keyboardShift = !_keyboardShift);
                  },
                  onNext: _focusNextInput,
                  onClose: () => _dismissKeyboard(),
                ),
              ),
          ],
        );
      },
    );
  }

  Widget _buildCard() {
    return Container(
      width: _loginCardWidth,
      height: _loginCardHeight,
      padding: const EdgeInsets.fromLTRB(40, 62, 40, 38),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.91),
        borderRadius: BorderRadius.circular(28),
        border: Border.all(color: Colors.white, width: 1.4),
        boxShadow: _softShadow(
          color: const Color(0x1FCB8C65),
          blur: 34,
          offset: const Offset(0, 18),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: _qrMode ? _buildQrLogin() : _buildAccountLogin(),
      ),
    );
  }

  List<Widget> _buildAccountLogin() {
    final bool useCustomKeyboard = _usesCustomLoginKeyboard;
    return <Widget>[
      Row(
        children: <Widget>[
          const Expanded(
            child: Text(
              '机构账号登录',
              style: TextStyle(
                color: AppColors.ink,
                fontSize: 27,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          LoginModeButton(
            icon: Icons.qr_code_rounded,
            label: '二维码登录',
            onTap: () {
              _dismissKeyboard();
              setState(() {
                _qrMode = true;
                _errorMessage = null;
              });
            },
          ),
        ],
      ),
      const SizedBox(height: 34),
      LoginTextField(
        controller: _usernameController,
        focusNode: _usernameFocusNode,
        icon: Icons.person_outline_rounded,
        hint: '手机号 / 账号',
        useCustomKeyboard: useCustomKeyboard,
        active: useCustomKeyboard &&
            _activeInput == _LoginInputTarget.username &&
            _customKeyboardVisible,
        onTap: useCustomKeyboard
            ? () => _openCustomKeyboard(_LoginInputTarget.username)
            : null,
        textInputAction: TextInputAction.next,
        onChanged: (_) => _clearError(),
      ),
      const SizedBox(height: 18),
      LoginTextField(
        controller: _passwordController,
        focusNode: _passwordFocusNode,
        icon: Icons.lock_outline_rounded,
        hint: '密码',
        useCustomKeyboard: useCustomKeyboard,
        active: useCustomKeyboard &&
            _activeInput == _LoginInputTarget.password &&
            _customKeyboardVisible,
        onTap: useCustomKeyboard
            ? () => _openCustomKeyboard(_LoginInputTarget.password)
            : null,
        obscureText: !_passwordVisible,
        textInputAction: TextInputAction.done,
        onSubmitted: (_) => _submit(),
        onChanged: (_) => _clearError(),
        suffix: IconButton(
          splashRadius: 20,
          onPressed: () {
            setState(() => _passwordVisible = !_passwordVisible);
          },
          icon: Icon(
            _passwordVisible
                ? Icons.visibility_outlined
                : Icons.visibility_off_outlined,
            size: 20,
            color: AppColors.muted,
          ),
        ),
      ),
      const SizedBox(height: 20),
      RememberPasswordRow(
        value: _rememberPassword,
        onChanged: (bool value) {
          setState(() => _rememberPassword = value);
        },
        onForgotPassword: _showForgotPasswordHint,
      ),
      const SizedBox(height: 18),
      AnimatedSwitcher(
        duration: const Duration(milliseconds: 160),
        child: _errorMessage == null
            ? const SizedBox(key: ValueKey<String>('empty-error'), height: 24)
            : SizedBox(
                key: const ValueKey<String>('login-error'),
                height: 24,
                child: Text(
                  _errorMessage!,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: AppColors.orangeDeep,
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
      ),
      const SizedBox(height: 16),
      LoginPrimaryButton(
        loading: _loading,
        label: _loading ? '登录中...' : '登 录',
        onTap: _submit,
      ),
    ];
  }

  List<Widget> _buildQrLogin() {
    final String qrData =
        widget.authClient.buildQrLoginUri(_qrNonce).toString();
    return <Widget>[
      Row(
        children: <Widget>[
          const Expanded(
            child: Text(
              '二维码登录',
              style: TextStyle(
                color: AppColors.ink,
                fontSize: 27,
                fontWeight: FontWeight.w800,
              ),
            ),
          ),
          LoginModeButton(
            icon: Icons.person_outline_rounded,
            label: '账号登录',
            onTap: () {
              _dismissKeyboard();
              setState(() {
                _qrMode = false;
                _errorMessage = null;
              });
            },
          ),
        ],
      ),
      const SizedBox(height: 28),
      Center(
        child: Container(
          width: 216,
          height: 216,
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: AppColors.line, width: 1.2),
            boxShadow: _softShadow(
              color: const Color(0x12D46B48),
              blur: 18,
              offset: const Offset(0, 10),
            ),
          ),
          child: QrImageView(
            data: qrData,
            version: QrVersions.auto,
            backgroundColor: Colors.white,
            eyeStyle: const QrEyeStyle(
              eyeShape: QrEyeShape.square,
              color: AppColors.ink,
            ),
            dataModuleStyle: const QrDataModuleStyle(
              dataModuleShape: QrDataModuleShape.square,
              color: AppColors.ink,
            ),
          ),
        ),
      ),
      const SizedBox(height: 22),
      const Center(
        child: Text(
          '请使用机构端 App 扫码确认',
          style: TextStyle(
            color: AppColors.body,
            fontSize: 15,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      const Spacer(),
      GestureDetector(
        onTap: _refreshQrCode,
        child: Container(
          height: 46,
          decoration: BoxDecoration(
            color: const Color(0xFFFFF1E8),
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: const Color(0xFFFFD3BD), width: 1.2),
          ),
          alignment: Alignment.center,
          child: const Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: <Widget>[
              Icon(Icons.refresh_rounded, color: AppColors.orange, size: 20),
              SizedBox(width: 8),
              Text(
                '刷新二维码',
                style: TextStyle(
                  color: AppColors.orange,
                  fontSize: 15,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ],
          ),
        ),
      ),
    ];
  }
}

class InstitutionPickerDialog extends StatelessWidget {
  const InstitutionPickerDialog({required this.options, super.key});

  final List<InstitutionLoginOption> options;

  @override
  Widget build(BuildContext context) {
    return Dialog(
      elevation: 0,
      insetPadding: const EdgeInsets.symmetric(horizontal: 24, vertical: 18),
      backgroundColor: Colors.transparent,
      child: Container(
        width: 700,
        constraints: const BoxConstraints(maxHeight: 620),
        padding: const EdgeInsets.fromLTRB(30, 18, 30, 28),
        decoration: BoxDecoration(
          color: Colors.white.withOpacity(.96),
          borderRadius: BorderRadius.circular(24),
          border: Border.all(color: Colors.white, width: 1.4),
          boxShadow: _softShadow(
            color: const Color(0x24CB8C65),
            blur: 32,
            offset: const Offset(0, 18),
          ),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Row(
              children: <Widget>[
                Container(
                  width: 50,
                  height: 50,
                  decoration: const BoxDecoration(
                    color: Color(0xFFFFF1E8),
                    shape: BoxShape.circle,
                  ),
                  child: const Icon(
                    Icons.business_rounded,
                    color: AppColors.orange,
                    size: 26,
                  ),
                ),
                const SizedBox(width: 16),
                const Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        '选择登录机构',
                        style: TextStyle(
                          color: AppColors.ink,
                          fontSize: 20,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      SizedBox(height: 3),
                      Text(
                        '当前账号关联多个机构，请选择本次进入的后台',
                        style: TextStyle(
                          color: AppColors.body,
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ],
                  ),
                ),
                IconButton(
                  splashRadius: 20,
                  onPressed: () => Navigator.of(context).pop(),
                  icon: const Icon(
                    Icons.close_rounded,
                    color: AppColors.muted,
                    size: 24,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 22),
            Flexible(
              child: ListView.separated(
                shrinkWrap: true,
                itemCount: options.length,
                separatorBuilder: (_, __) => const SizedBox(height: 16),
                itemBuilder: (BuildContext context, int index) {
                  return InstitutionOptionTile(option: options[index]);
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class InstitutionOptionTile extends StatelessWidget {
  const InstitutionOptionTile({required this.option, super.key});

  final InstitutionLoginOption option;

  @override
  Widget build(BuildContext context) {
    final String subtitle = option.nickName.trim().isNotEmpty
        ? '${option.nickName} · ${option.mobile}'
        : option.mobile;

    return GestureDetector(
      onTap: () => Navigator.of(context).pop(option),
      child: Container(
        constraints: const BoxConstraints(minHeight: 136),
        padding: const EdgeInsets.fromLTRB(20, 18, 18, 18),
        decoration: BoxDecoration(
          color: const Color(0xFFFFFBF6),
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: const Color(0xFFF3DACB), width: 1.2),
        ),
        child: Row(
          children: <Widget>[
            _InstitutionAvatar(option: option),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    option.orgName,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: AppColors.ink,
                      fontSize: 18,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    subtitle.trim().isNotEmpty ? subtitle : option.loginName,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: const TextStyle(
                      color: AppColors.body,
                      fontSize: 15,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 10),
                  Row(
                    children: <Widget>[
                      _InstitutionTag(
                        icon: Icons.verified_user_outlined,
                        label: option.admin ? '超管' : '员工',
                        color: AppColors.orangeDeep,
                        background: const Color(0xFFFFF5EF),
                        border: const Color(0xFFFFDDCC),
                      ),
                      const SizedBox(width: 8),
                      _InstitutionTag(
                        icon: _statusIcon(option.institutionStatus),
                        label: _statusLabel(option.institutionStatus),
                        color: _statusColor(option.institutionStatus),
                        background: _statusBackground(option.institutionStatus),
                        border: _statusBorder(option.institutionStatus),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            Container(
              width: 42,
              height: 42,
              decoration: const BoxDecoration(
                color: Color(0xFFFFEFE6),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.arrow_forward_ios_rounded,
                color: AppColors.orange,
                size: 18,
              ),
            ),
          ],
        ),
      ),
    );
  }

  static String _statusLabel(String status) {
    switch (status) {
      case 'warning':
        return '即将到期';
      case 'disabled':
        return '已停用';
      case 'trial_expired':
        return '试用到期';
      case 'expired_readonly':
        return '只读模式';
      default:
        return '正常';
    }
  }

  static IconData _statusIcon(String status) {
    switch (status) {
      case 'warning':
        return Icons.schedule_rounded;
      case 'disabled':
      case 'trial_expired':
      case 'expired_readonly':
        return Icons.lock_outline_rounded;
      default:
        return Icons.check_circle_outline_rounded;
    }
  }

  static Color _statusColor(String status) {
    switch (status) {
      case 'warning':
        return const Color(0xFFC7821E);
      case 'disabled':
      case 'trial_expired':
      case 'expired_readonly':
        return const Color(0xFF9E6D5D);
      default:
        return const Color(0xFF7BA36F);
    }
  }

  static Color _statusBackground(String status) {
    switch (status) {
      case 'warning':
        return const Color(0xFFFFFAEA);
      case 'disabled':
      case 'trial_expired':
      case 'expired_readonly':
        return const Color(0xFFFAF2EE);
      default:
        return const Color(0xFFF2F8EE);
    }
  }

  static Color _statusBorder(String status) {
    switch (status) {
      case 'warning':
        return const Color(0xFFFFE4A8);
      case 'disabled':
      case 'trial_expired':
      case 'expired_readonly':
        return const Color(0xFFEBD8CF);
      default:
        return const Color(0xFFDDEED6);
    }
  }
}

class _InstitutionAvatar extends StatelessWidget {
  const _InstitutionAvatar({required this.option});

  final InstitutionLoginOption option;

  @override
  Widget build(BuildContext context) {
    final String logo = option.logo.trim();
    return Container(
      width: 64,
      height: 64,
      decoration: BoxDecoration(
        color: const Color(0xFFFFE7D8),
        borderRadius: BorderRadius.circular(20),
      ),
      clipBehavior: Clip.antiAlias,
      child: logo.isEmpty
          ? const Icon(
              Icons.apartment_rounded,
              color: AppColors.orange,
              size: 32,
            )
          : Image.network(
              logo,
              fit: BoxFit.cover,
              errorBuilder: (_, __, ___) => const Icon(
                Icons.apartment_rounded,
                color: AppColors.orange,
                size: 32,
              ),
            ),
    );
  }
}

class _InstitutionTag extends StatelessWidget {
  const _InstitutionTag({
    required this.icon,
    required this.label,
    required this.color,
    required this.background,
    required this.border,
  });

  final IconData icon;
  final String label;
  final Color color;
  final Color background;
  final Color border;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 26,
      padding: const EdgeInsets.symmetric(horizontal: 9),
      decoration: BoxDecoration(
        color: background,
        borderRadius: BorderRadius.circular(999),
        border: Border.all(color: border, width: .8),
      ),
      alignment: Alignment.center,
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Icon(icon, size: 12, color: color),
          const SizedBox(width: 5),
          Text(
            label,
            style: TextStyle(
              color: color,
              fontSize: 12,
              fontWeight: FontWeight.w800,
              height: 1,
            ),
          ),
        ],
      ),
    );
  }
}

class FloatingLoginKeyboard extends StatelessWidget {
  const FloatingLoginKeyboard({
    required this.targetLabel,
    required this.shifted,
    required this.onDragUpdate,
    required this.onKey,
    required this.onBackspace,
    required this.onClear,
    required this.onShift,
    required this.onNext,
    required this.onClose,
    super.key,
  });

  final String targetLabel;
  final bool shifted;
  final GestureDragUpdateCallback onDragUpdate;
  final ValueChanged<String> onKey;
  final VoidCallback onBackspace;
  final VoidCallback onClear;
  final VoidCallback onShift;
  final VoidCallback onNext;
  final VoidCallback onClose;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: _loginKeyboardWidth,
      height: _loginKeyboardHeight,
      padding: const EdgeInsets.fromLTRB(14, 12, 14, 16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.96),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.white, width: 1.2),
        boxShadow: _softShadow(
          color: const Color(0x24CB8C65),
          blur: 26,
          offset: const Offset(0, 14),
        ),
      ),
      child: Column(
        children: <Widget>[
          GestureDetector(
            behavior: HitTestBehavior.opaque,
            onPanUpdate: onDragUpdate,
            child: SizedBox(
              height: 38,
              child: Row(
                children: <Widget>[
                  const Icon(
                    Icons.drag_indicator_rounded,
                    color: AppColors.muted,
                    size: 21,
                  ),
                  const SizedBox(width: 6),
                  Container(
                    height: 24,
                    padding: const EdgeInsets.symmetric(horizontal: 10),
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFF1E8),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    alignment: Alignment.center,
                    child: Text(
                      targetLabel,
                      style: const TextStyle(
                        color: AppColors.orange,
                        fontSize: 13,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ),
                  const Spacer(),
                  IconButton(
                    splashRadius: 18,
                    padding: EdgeInsets.zero,
                    onPressed: onClose,
                    icon: const Icon(
                      Icons.close_rounded,
                      color: AppColors.muted,
                      size: 21,
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 10),
          _keyRow(_digitKeys),
          const SizedBox(height: 9),
          _keyRow(_letterKeys('qwertyuiop')),
          const SizedBox(height: 9),
          _keyRow(_letterKeys('asdfghjkl')),
          const SizedBox(height: 9),
          _keyRow(<Widget>[
            _KeyboardButton(
              id: 'shift',
              flex: 2,
              active: shifted,
              icon: Icons.keyboard_capslock_rounded,
              onTap: onShift,
            ),
            ..._letterKeys('zxcvbnm'),
            _KeyboardButton(
              id: 'backspace',
              flex: 2,
              icon: Icons.backspace_outlined,
              onTap: onBackspace,
            ),
          ]),
          const SizedBox(height: 9),
          _keyRow(<Widget>[
            _textKey('@'),
            _textKey('.'),
            _textKey('_'),
            _textKey('-'),
            _KeyboardButton(
              id: 'space',
              flex: 3,
              label: '空格',
              onTap: () => onKey(' '),
            ),
            _KeyboardButton(
              id: 'clear',
              flex: 2,
              label: '清空',
              muted: true,
              onTap: onClear,
            ),
            _KeyboardButton(
              id: 'next',
              flex: 3,
              label: targetLabel == '账号' ? '下一项' : '登录',
              primary: true,
              onTap: onNext,
            ),
          ]),
        ],
      ),
    );
  }

  List<Widget> get _digitKeys {
    return '1234567890'.split('').map(_textKey).toList();
  }

  List<Widget> _letterKeys(String values) {
    return values.split('').map((String value) {
      final String label = shifted ? value.toUpperCase() : value;
      return _KeyboardButton(
        id: value,
        label: label,
        onTap: () => onKey(label),
      );
    }).toList();
  }

  Widget _textKey(String value) {
    return _KeyboardButton(
      id: value,
      label: value,
      onTap: () => onKey(value),
    );
  }

  Widget _keyRow(List<Widget> children) {
    return Row(
      children: <Widget>[
        for (int index = 0; index < children.length; index++) ...<Widget>[
          if (index > 0) const SizedBox(width: 8),
          children[index],
        ],
      ],
    );
  }
}

class _KeyboardButton extends StatefulWidget {
  const _KeyboardButton({
    required this.id,
    required this.onTap,
    this.label,
    this.icon,
    this.flex = 1,
    this.primary = false,
    this.active = false,
    this.muted = false,
  });

  final String id;
  final String? label;
  final IconData? icon;
  final int flex;
  final bool primary;
  final bool active;
  final bool muted;
  final VoidCallback onTap;

  @override
  State<_KeyboardButton> createState() => _KeyboardButtonState();
}

class _KeyboardButtonState extends State<_KeyboardButton> {
  bool _pressed = false;
  bool _showBubble = false;

  void _handleTapDown(TapDownDetails _) {
    HapticFeedback.selectionClick();
    setState(() {
      _pressed = true;
      _showBubble = true;
    });
  }

  void _hidePressBubbleSoon() {
    setState(() => _pressed = false);
    Future<void>.delayed(const Duration(milliseconds: 130), () {
      if (mounted) {
        setState(() => _showBubble = false);
      }
    });
  }

  String get _bubbleLabel {
    if (widget.label != null && widget.label!.trim().isNotEmpty) {
      return widget.label!;
    }
    switch (widget.id) {
      case 'shift':
        return '大写';
      case 'backspace':
        return '删除';
      default:
        return widget.id;
    }
  }

  @override
  Widget build(BuildContext context) {
    final Color background = widget.primary
        ? AppColors.orange
        : widget.active
            ? const Color(0xFFFFE5D6)
            : widget.muted
                ? const Color(0xFFF6EFEA)
                : const Color(0xFFFFFAF5);
    final Color foreground = widget.primary
        ? Colors.white
        : widget.active
            ? AppColors.orangeDeep
            : widget.muted
                ? AppColors.body
                : AppColors.ink;
    final Color pressedBackground = widget.primary
        ? AppColors.orangeDeep
        : widget.active
            ? const Color(0xFFFFD6C2)
            : const Color(0xFFFFE8DA);
    final Color keyBackground = _pressed ? pressedBackground : background;
    final Color keyBorder =
        _pressed ? AppColors.orange : const Color(0xFFF0DACB);

    return Expanded(
      flex: widget.flex,
      child: GestureDetector(
        key: ValueKey<String>('login-key-${widget.id}'),
        onTapDown: _handleTapDown,
        onTapUp: (_) => _hidePressBubbleSoon(),
        onTapCancel: () {
          if (mounted) {
            setState(() {
              _pressed = false;
              _showBubble = false;
            });
          }
        },
        onTap: widget.onTap,
        child: Stack(
          clipBehavior: Clip.none,
          alignment: Alignment.center,
          children: <Widget>[
            AnimatedScale(
              scale: _pressed ? .96 : 1,
              duration: const Duration(milliseconds: 70),
              curve: Curves.easeOut,
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 70),
                curve: Curves.easeOut,
                height: _loginKeyboardKeyHeight,
                decoration: BoxDecoration(
                  color: keyBackground,
                  borderRadius: BorderRadius.circular(13),
                  border: widget.primary
                      ? null
                      : Border.all(color: keyBorder, width: _pressed ? 1.4 : 1),
                  boxShadow: widget.primary || _pressed
                      ? _softShadow(
                          color: _pressed
                              ? const Color(0x2FD15E36)
                              : const Color(0x20D15E36),
                          blur: _pressed ? 16 : 12,
                          offset: Offset(0, _pressed ? 5 : 7),
                        )
                      : null,
                ),
                alignment: Alignment.center,
                child: widget.icon == null
                    ? Text(
                        widget.label ?? '',
                        maxLines: 1,
                        overflow: TextOverflow.clip,
                        style: TextStyle(
                          color: foreground,
                          fontSize:
                              widget.label != null && widget.label!.length > 2
                                  ? 15
                                  : 19,
                          fontWeight: FontWeight.w800,
                          height: 1,
                        ),
                      )
                    : Icon(widget.icon, color: foreground, size: 22),
              ),
            ),
            if (_showBubble)
              Positioned(
                bottom: _loginKeyboardKeyHeight + 8,
                child: IgnorePointer(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: <Widget>[
                      Container(
                        height: 46,
                        constraints: const BoxConstraints(minWidth: 46),
                        padding: const EdgeInsets.symmetric(horizontal: 15),
                        decoration: BoxDecoration(
                          color: AppColors.orange,
                          borderRadius: BorderRadius.circular(14),
                          boxShadow: _softShadow(
                            color: const Color(0x24D15E36),
                            blur: 14,
                            offset: const Offset(0, 8),
                          ),
                        ),
                        alignment: Alignment.center,
                        child: Text(
                          _bubbleLabel,
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 20,
                            fontWeight: FontWeight.w900,
                            height: 1,
                          ),
                        ),
                      ),
                      ClipPath(
                        clipper: TriangleClipper(),
                        child: Container(
                          width: 16,
                          height: 9,
                          color: AppColors.orange,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class LoginTextField extends StatelessWidget {
  const LoginTextField({
    required this.controller,
    required this.focusNode,
    required this.icon,
    required this.hint,
    required this.active,
    required this.useCustomKeyboard,
    this.suffix,
    this.obscureText = false,
    this.textInputAction,
    this.onTap,
    this.onChanged,
    this.onSubmitted,
    super.key,
  });

  final TextEditingController controller;
  final FocusNode focusNode;
  final IconData icon;
  final Widget? suffix;
  final String hint;
  final bool active;
  final bool useCustomKeyboard;
  final bool obscureText;
  final TextInputAction? textInputAction;
  final VoidCallback? onTap;
  final ValueChanged<String>? onChanged;
  final ValueChanged<String>? onSubmitted;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 58,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(
          color: active ? AppColors.orange : AppColors.line,
          width: active ? 1.8 : 1.4,
        ),
      ),
      child: Row(
        children: <Widget>[
          const SizedBox(width: 18),
          Icon(icon, size: 21, color: AppColors.muted),
          const SizedBox(width: 14),
          Expanded(
            child: TextField(
              controller: controller,
              focusNode: focusNode,
              obscureText: obscureText,
              readOnly: useCustomKeyboard,
              showCursor: useCustomKeyboard ? active : null,
              keyboardType:
                  useCustomKeyboard ? TextInputType.none : TextInputType.text,
              enableInteractiveSelection: !useCustomKeyboard,
              textInputAction: textInputAction,
              onTap: onTap,
              onChanged: onChanged,
              onSubmitted: onSubmitted,
              cursorColor: AppColors.orange,
              style: const TextStyle(
                color: AppColors.ink,
                fontSize: 16,
                fontWeight: FontWeight.w600,
              ),
              decoration: InputDecoration(
                hintText: hint,
                border: InputBorder.none,
                isCollapsed: true,
                hintStyle: const TextStyle(
                  color: AppColors.muted,
                  fontSize: 16,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
          ),
          if (suffix != null)
            Padding(
              padding: const EdgeInsets.only(right: 8),
              child: suffix,
            ),
        ],
      ),
    );
  }
}

class RememberPasswordRow extends StatelessWidget {
  const RememberPasswordRow({
    required this.value,
    required this.onChanged,
    required this.onForgotPassword,
    super.key,
  });

  final bool value;
  final ValueChanged<bool> onChanged;
  final VoidCallback onForgotPassword;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        GestureDetector(
          onTap: () => onChanged(!value),
          child: Container(
            width: 18,
            height: 18,
            decoration: BoxDecoration(
              color: value ? AppColors.orange : Colors.white,
              borderRadius: BorderRadius.circular(5),
              border: Border.all(
                color: value ? AppColors.orange : const Color(0xFFE5CBB8),
                width: 1.3,
              ),
            ),
            child: value
                ? const Icon(Icons.check_rounded, color: Colors.white, size: 14)
                : null,
          ),
        ),
        const SizedBox(width: 9),
        GestureDetector(
          onTap: () => onChanged(!value),
          child: const Text(
            '记住密码',
            style: TextStyle(
              color: AppColors.body,
              fontSize: 14,
              fontWeight: FontWeight.w500,
            ),
          ),
        ),
        const Spacer(),
        GestureDetector(
          onTap: onForgotPassword,
          child: const Text(
            '忘记密码',
            style: TextStyle(
              color: AppColors.orange,
              fontSize: 14,
              fontWeight: FontWeight.w700,
            ),
          ),
        ),
      ],
    );
  }
}

class LoginModeButton extends StatelessWidget {
  const LoginModeButton({
    required this.icon,
    required this.label,
    required this.onTap,
    super.key,
  });

  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        height: 34,
        padding: const EdgeInsets.symmetric(horizontal: 12),
        decoration: BoxDecoration(
          color: const Color(0xFFFFF1E8),
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: const Color(0xFFFFD5C0), width: 1),
        ),
        child: Row(
          children: <Widget>[
            Icon(icon, size: 16, color: AppColors.orange),
            const SizedBox(width: 6),
            Text(
              label,
              style: const TextStyle(
                color: AppColors.orange,
                fontSize: 13,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class LoginPrimaryButton extends StatelessWidget {
  const LoginPrimaryButton({
    required this.loading,
    required this.label,
    required this.onTap,
    super.key,
  });

  final bool loading;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: loading ? null : onTap,
      child: Opacity(
        opacity: loading ? .78 : 1,
        child: Container(
          height: 60,
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              colors: <Color>[Color(0xFFE86E43), Color(0xFFD95B35)],
            ),
            borderRadius: BorderRadius.circular(10),
            boxShadow: _softShadow(
              color: const Color(0x28D15E36),
              blur: 18,
              offset: const Offset(0, 10),
            ),
          ),
          alignment: Alignment.center,
          child: loading
              ? const SizedBox(
                  width: 22,
                  height: 22,
                  child: CircularProgressIndicator(
                    strokeWidth: 2.4,
                    valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                  ),
                )
              : Text(
                  label,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 18,
                    fontWeight: FontWeight.w700,
                  ),
                ),
        ),
      ),
    );
  }
}

class LoginIllustration extends StatelessWidget {
  const LoginIllustration({super.key});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 435,
      height: 360,
      child: Stack(
        clipBehavior: Clip.none,
        children: <Widget>[
          Positioned(
            left: -64,
            bottom: 76,
            child: CustomPaint(
              size: const Size(92, 160),
              painter: LeafClusterPainter(),
            ),
          ),
          Positioned(
            left: 34,
            bottom: 28,
            child: Container(
              width: 326,
              height: 58,
              decoration: BoxDecoration(
                color: const Color(0xFFF7CDA9).withOpacity(.34),
                borderRadius: BorderRadius.circular(40),
              ),
            ),
          ),
          Positioned(
            left: 28,
            bottom: 72,
            child: Container(
              width: 128,
              height: 128,
              decoration: BoxDecoration(
                color: const Color(0xFFFFE5C9),
                borderRadius: BorderRadius.circular(70),
                boxShadow: _softShadow(
                  color: const Color(0x22D7743B),
                  blur: 18,
                  offset: const Offset(0, 12),
                ),
              ),
            ),
          ),
          Positioned(
            left: 52,
            bottom: 111,
            child: CustomPaint(
              size: const Size(88, 82),
              painter: PieMarkPainter(),
            ),
          ),
          Positioned(
            left: 58,
            bottom: 21,
            child: Container(
              width: 244,
              height: 76,
              decoration: BoxDecoration(
                color: const Color(0xFFF8D8BC),
                borderRadius: BorderRadius.circular(48),
              ),
            ),
          ),
          Positioned(
            left: 118,
            top: 65,
            child: Transform.rotate(
              angle: -.01,
              child: Container(
                width: 238,
                height: 184,
                padding: const EdgeInsets.all(21),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: <Color>[Color(0xFFFFF0E0), Color(0xFFFFDABB)],
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.white, width: 4),
                  boxShadow: _softShadow(
                    color: const Color(0x2DD9794A),
                    blur: 24,
                    offset: const Offset(0, 18),
                  ),
                ),
                child: const EvaluationCardGraphic(),
              ),
            ),
          ),
          Positioned(
            left: 292,
            top: 44,
            child: Container(
              width: 102,
              height: 76,
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  colors: <Color>[Color(0xFFFFCBB7), Color(0xFFFFE2D2)],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(38),
                boxShadow: _softShadow(
                  color: const Color(0x1BD46B48),
                  blur: 20,
                  offset: const Offset(0, 12),
                ),
              ),
              child: const Icon(
                Icons.favorite_rounded,
                color: Colors.white,
                size: 34,
              ),
            ),
          ),
          Positioned(
            left: 320,
            bottom: 84,
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: <Widget>[
                _bar(22, 88, const Color(0xFFFFB35E)),
                const SizedBox(width: 10),
                _bar(22, 118, const Color(0xFFE97845)),
              ],
            ),
          ),
          Positioned(
            right: 8,
            bottom: 7,
            child: Container(
              width: 104,
              height: 84,
              decoration: BoxDecoration(
                color: const Color(0xFFFFE5C7),
                borderRadius: BorderRadius.circular(54),
              ),
            ),
          ),
          Positioned(
            right: -4,
            bottom: 48,
            child: Container(
              width: 82,
              height: 82,
              decoration: BoxDecoration(
                color: const Color(0xFFFBD7B9),
                borderRadius: BorderRadius.circular(50),
              ),
            ),
          ),
        ],
      ),
    );
  }

  static Widget _bar(double width, double height, Color color) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: color,
        borderRadius: BorderRadius.circular(7),
        boxShadow: _softShadow(
          color: color.withOpacity(.23),
          blur: 16,
          offset: const Offset(0, 8),
        ),
      ),
    );
  }
}

class EvaluationCardGraphic extends StatelessWidget {
  const EvaluationCardGraphic({super.key});

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Container(
              width: 40,
              height: 40,
              decoration: const BoxDecoration(
                color: Color(0xFFFFE7CB),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.person_rounded,
                color: Color(0xFFFFA647),
                size: 25,
              ),
            ),
            const SizedBox(height: 16),
            _line(46),
            const SizedBox(height: 13),
            _line(76),
            const SizedBox(height: 13),
            _line(58),
          ],
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              const SizedBox(height: 7),
              _line(72),
              const SizedBox(height: 16),
              SizedBox(
                width: 104,
                height: 96,
                child: CustomPaint(painter: HexRadarPainter()),
              ),
            ],
          ),
        ),
      ],
    );
  }

  static Widget _line(double width) {
    return Container(
      width: width,
      height: 7,
      decoration: BoxDecoration(
        color: const Color(0xFFE9A17A).withOpacity(.36),
        borderRadius: BorderRadius.circular(8),
      ),
    );
  }
}

class HexRadarPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Offset center = Offset(size.width / 2, size.height / 2 + 4);
    final Paint axisPaint = Paint()
      ..color = const Color(0xFFF0B27D).withOpacity(.42)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1;
    for (int ring = 1; ring <= 4; ring++) {
      final double radius = 14.0 * ring;
      final Path path = _polygon(center, radius);
      canvas.drawPath(path, axisPaint);
    }
    for (int i = 0; i < 6; i++) {
      final double angle = -math.pi / 2 + math.pi / 3 * i;
      canvas.drawLine(
        center,
        Offset(
            center.dx + math.cos(angle) * 58, center.dy + math.sin(angle) * 58),
        axisPaint,
      );
    }
    final Path value = Path();
    final List<double> radii = <double>[42, 34, 51, 39, 45, 30];
    for (int i = 0; i < radii.length; i++) {
      final double angle = -math.pi / 2 + math.pi / 3 * i;
      final Offset point = Offset(
        center.dx + math.cos(angle) * radii[i],
        center.dy + math.sin(angle) * radii[i],
      );
      if (i == 0) {
        value.moveTo(point.dx, point.dy);
      } else {
        value.lineTo(point.dx, point.dy);
      }
    }
    value.close();
    canvas.drawPath(
        value, Paint()..color = const Color(0xFFFFBA67).withOpacity(.47));
    canvas.drawPath(
      value,
      Paint()
        ..color = const Color(0xFFF39C42)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 1.4,
    );
  }

  Path _polygon(Offset center, double radius) {
    final Path path = Path();
    for (int i = 0; i < 6; i++) {
      final double angle = -math.pi / 2 + math.pi / 3 * i;
      final Offset point = Offset(
        center.dx + math.cos(angle) * radius,
        center.dy + math.sin(angle) * radius,
      );
      if (i == 0) {
        path.moveTo(point.dx, point.dy);
      } else {
        path.lineTo(point.dx, point.dy);
      }
    }
    path.close();
    return path;
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class CloudCheckPainter extends CustomPainter {
  const CloudCheckPainter({required this.color});

  final Color color;

  @override
  void paint(Canvas canvas, Size size) {
    final Paint paint = Paint()
      ..color = color
      ..style = PaintingStyle.stroke
      ..strokeWidth = 4.4
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round;
    final Path cloud = Path()
      ..moveTo(size.width * .18, size.height * .68)
      ..cubicTo(size.width * .06, size.height * .66, size.width * .04,
          size.height * .38, size.width * .24, size.height * .38)
      ..cubicTo(size.width * .30, size.height * .10, size.width * .70,
          size.height * .12, size.width * .74, size.height * .42)
      ..cubicTo(size.width * .92, size.height * .42, size.width * .95,
          size.height * .66, size.width * .80, size.height * .72);
    canvas.drawPath(cloud, paint);
    final Path check = Path()
      ..moveTo(size.width * .48, size.height * .58)
      ..lineTo(size.width * .58, size.height * .70)
      ..lineTo(size.width * .76, size.height * .47);
    canvas.drawPath(check, paint);
  }

  @override
  bool shouldRepaint(covariant CloudCheckPainter oldDelegate) =>
      oldDelegate.color != color;
}

class LeafClusterPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint stem = Paint()
      ..color = const Color(0xFF91AE75)
      ..strokeWidth = 5
      ..strokeCap = StrokeCap.round;
    canvas.drawLine(Offset(size.width * .45, size.height),
        Offset(size.width * .58, 26), stem);
    _leaf(canvas, Offset(48, 82), 58, -1.15, const Color(0xFF89A86E));
    _leaf(canvas, Offset(37, 124), 68, -.84, const Color(0xFFA0BE84));
    _leaf(canvas, Offset(66, 56), 46, -.18, const Color(0xFF7E9E68));
  }

  void _leaf(
      Canvas canvas, Offset center, double length, double angle, Color color) {
    final Path path = Path();
    final Offset tip = Offset(center.dx + math.cos(angle) * length,
        center.dy + math.sin(angle) * length);
    final Offset c1 = Offset(center.dx - math.sin(angle) * length * .38,
        center.dy + math.cos(angle) * length * .38);
    final Offset c2 = Offset(tip.dx - math.sin(angle) * length * .34,
        tip.dy + math.cos(angle) * length * .34);
    final Offset c3 = Offset(center.dx + math.sin(angle) * length * .34,
        center.dy - math.cos(angle) * length * .34);
    final Offset c4 = Offset(tip.dx + math.sin(angle) * length * .28,
        tip.dy - math.cos(angle) * length * .28);
    path
      ..moveTo(center.dx, center.dy)
      ..cubicTo(c1.dx, c1.dy, c2.dx, c2.dy, tip.dx, tip.dy)
      ..cubicTo(c4.dx, c4.dy, c3.dx, c3.dy, center.dx, center.dy)
      ..close();
    canvas.drawPath(path, Paint()..color = color.withOpacity(.9));
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class PieMarkPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Rect rect = Offset.zero & size;
    canvas.drawArc(
      rect.deflate(2),
      -math.pi / 2,
      math.pi * 1.55,
      true,
      Paint()..color = const Color(0xFFFFEBD4),
    );
    canvas.drawArc(
      rect.deflate(2),
      -math.pi / 2,
      math.pi * .68,
      true,
      Paint()..color = AppColors.orange,
    );
    canvas.drawArc(
      rect.deflate(2),
      math.pi * .18,
      math.pi * .62,
      true,
      Paint()..color = const Color(0xFFF7B15F),
    );
    canvas.drawCircle(
      Offset(size.width / 2, size.height / 2),
      26,
      Paint()..color = const Color(0xFFFFF0DF),
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class HomeScreen extends StatefulWidget {
  const HomeScreen({required this.homeClient, super.key});

  final HomeClient homeClient;

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  HomeSummary _summary = HomeSummary.fallback();
  HomeSession _session = HomeSession.fallback;
  bool _loading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    runAfterRouteEntrance(context, () => _loadHomeData(bootstrap: true));
  }

  Future<void> _loadHomeData({bool bootstrap = false}) async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    final String token = prefs.getString(_authTokenStorageKey) ?? '';
    if (token.trim().isEmpty) {
      await _logout();
      return;
    }
    if (mounted && !bootstrap) {
      setState(() {
        _loading = true;
        _errorMessage = null;
      });
    }
    try {
      final List<Object> result = await Future.wait<Object>(<Future<Object>>[
        widget.homeClient.fetchCurrentSession(token),
        widget.homeClient.fetchSummary(token),
      ]);
      if (!mounted) {
        return;
      }
      setState(() {
        _session = result[0] as HomeSession;
        _summary = result[1] as HomeSummary;
        _loading = false;
      });
    } on HomeApiException catch (error) {
      if (!mounted) {
        return;
      }
      if (error.unauthorized) {
        await _logout();
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = error.message;
      });
    } on Object catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _loading = false;
        _errorMessage = '首页数据加载失败：$error';
      });
    }
  }

  Future<void> _logout() async {
    final SharedPreferences prefs = await SharedPreferences.getInstance();
    await prefs.remove(_authTokenStorageKey);
    await prefs.remove(_authLoginTypeStorageKey);
    await prefs.remove(_authTenantIdStorageKey);
    await prefs.remove(_authOrgIdStorageKey);
    if (!mounted) {
      return;
    }
    Navigator.of(context).pushNamedAndRemoveUntil(
      '/',
      (Route<dynamic> route) => false,
    );
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width = constraints.maxWidth.isFinite
            ? constraints.maxWidth
            : _wideDesignWidth;
        final bool compact = width < 1200;
        final _HomeMetrics metrics =
            compact ? _HomeMetrics.compact(width) : _HomeMetrics.wide(width);

        return Stack(
          children: <Widget>[
            CustomPaint(
              size: Size(width, _designHeight),
              painter: HomeBackgroundPainter(),
            ),
            Positioned(
              left: metrics.margin,
              top: 38,
              right: metrics.margin,
              child: HomeHeader(
                session: _session,
                weather: _summary.weather,
                date: _summary.date,
                weekday: _summary.weekday,
                loading: _loading,
                errorMessage: _errorMessage,
                onRefresh: _loadHomeData,
                onLogout: _logout,
              ),
            ),
            Positioned(
              left: metrics.margin,
              top: _homeMainTop,
              child: StartAssessmentCard(width: metrics.startWidth),
            ),
            Positioned(
              left: metrics.rightColumnLeft,
              top: _homeMainTop,
              child: HomeStatsRow(
                width: metrics.scheduleWidth,
                cardWidth: metrics.statWidth,
                spacing: metrics.statSpacing,
                stats: _summary.assessmentStats,
              ),
            ),
            Positioned(
              left: metrics.rightColumnLeft,
              top: _homeScheduleTop,
              child: ScheduleCard(
                width: metrics.scheduleWidth,
                height: _homeScheduleHeight,
                items: _summary.schedule,
                loading: _loading,
              ),
            ),
            Positioned(
              left: metrics.margin,
              top: _homeShortcutTop,
              child: FeatureShortcutRow(
                cardWidth: metrics.shortcutWidth,
                spacing: metrics.shortcutSpacing,
                onTimetableTap: () =>
                    Navigator.of(context).pushNamed('/smart-timetable'),
                onReportTap: () =>
                    Navigator.of(context).pushNamed('/assessment-reports'),
                onTrainingTap: () =>
                    Navigator.of(context).pushNamed('/training-center'),
              ),
            ),
            Positioned(
              left: metrics.margin,
              bottom: 24,
              child: ProgressOverviewCard(
                width: metrics.progressWidth,
                stats: _summary.assessmentStats,
              ),
            ),
            Positioned(
              right: metrics.margin,
              bottom: 24,
              child: MorePlansCard(width: metrics.moreWidth),
            ),
          ],
        );
      },
    );
  }
}

class _HomeMetrics {
  const _HomeMetrics({
    required this.margin,
    required this.startWidth,
    required this.rightColumnLeft,
    required this.statWidth,
    required this.statSpacing,
    required this.scheduleWidth,
    required this.shortcutWidth,
    required this.shortcutSpacing,
    required this.progressWidth,
    required this.moreWidth,
  });

  factory _HomeMetrics.compact(double width) {
    const double margin = 54;
    const double startWidth = 336;
    const double rightLeft = 420;
    final double available = width - rightLeft - margin;
    return _HomeMetrics(
      margin: margin,
      startWidth: startWidth,
      rightColumnLeft: rightLeft,
      statWidth: (available - 36) / 3,
      statSpacing: 18,
      scheduleWidth: available,
      shortcutWidth: (width - margin * 2 - 70) / 6,
      shortcutSpacing: 14,
      progressWidth: 612,
      moreWidth: width - margin * 2 - 612 - 18,
    );
  }

  factory _HomeMetrics.wide(double width) {
    const double margin = 86;
    const double startWidth = 430;
    const double rightLeft = 552;
    final double available = width - rightLeft - margin;
    return _HomeMetrics(
      margin: margin,
      startWidth: startWidth,
      rightColumnLeft: rightLeft,
      statWidth: (available - 36) / 3,
      statSpacing: 18,
      scheduleWidth: available,
      shortcutWidth: (width - margin * 2 - 70) / 6,
      shortcutSpacing: 14,
      progressWidth: (width - margin * 2 - 24) * .68,
      moreWidth: (width - margin * 2 - 24) * .32,
    );
  }

  final double margin;
  final double startWidth;
  final double rightColumnLeft;
  final double statWidth;
  final double statSpacing;
  final double scheduleWidth;
  final double shortcutWidth;
  final double shortcutSpacing;
  final double progressWidth;
  final double moreWidth;
}

class HomeBackgroundPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    canvas.drawRect(Offset.zero & size, Paint()..color = AppColors.page);

    final Path top = Path()
      ..moveTo(0, 0)
      ..lineTo(size.width, 0)
      ..lineTo(size.width, 182)
      ..quadraticBezierTo(size.width * .72, 126, size.width * .48, 148)
      ..quadraticBezierTo(size.width * .18, 180, 0, 122)
      ..close();
    canvas.drawPath(top, Paint()..color = const Color(0xFFFFF4E6));

    final Path lower = Path()
      ..moveTo(0, size.height - 118)
      ..quadraticBezierTo(size.width * .23, size.height - 174, size.width * .48,
          size.height - 118)
      ..quadraticBezierTo(
          size.width * .72, size.height - 68, size.width, size.height - 128)
      ..lineTo(size.width, size.height)
      ..lineTo(0, size.height)
      ..close();
    canvas.drawPath(lower, Paint()..color = const Color(0xFFFFF9F0));
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class HomeHeader extends StatelessWidget {
  const HomeHeader({
    required this.session,
    required this.weather,
    required this.date,
    required this.weekday,
    required this.loading,
    required this.errorMessage,
    required this.onRefresh,
    required this.onLogout,
    super.key,
  });

  final HomeSession session;
  final HomeWeather weather;
  final String date;
  final String weekday;
  final bool loading;
  final String? errorMessage;
  final VoidCallback onRefresh;
  final VoidCallback onLogout;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        WeatherIcon(weather: weather, size: 58),
        const SizedBox(width: 16),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(
              _homeHeaderTitle(session),
              style: const TextStyle(
                color: AppColors.ink,
                fontSize: 24,
                fontWeight: FontWeight.w800,
                height: 1,
              ),
            ),
            const SizedBox(height: 9),
            Text(
              _homeHeaderSubtitle(date, weekday, weather),
              style: const TextStyle(
                color: AppColors.body,
                fontSize: 13,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
        const Spacer(),
        if (errorMessage != null)
          HomeLoadStatus(message: errorMessage!, onRefresh: onRefresh)
        else if (loading)
          const SizedBox(
            width: 18,
            height: 18,
            child: CircularProgressIndicator(
              strokeWidth: 2.2,
              color: AppColors.orange,
            ),
          ),
        if (loading || errorMessage != null) const SizedBox(width: 24),
        SizedBox(
          width: 31,
          height: 31,
          child: IconButton(
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints.tightFor(width: 31, height: 31),
            splashRadius: 18,
            tooltip: '刷新首页',
            onPressed: loading ? null : onRefresh,
            icon: const Icon(
              Icons.refresh_rounded,
              size: 31,
              color: AppColors.ink,
            ),
          ),
        ),
        const SizedBox(width: 26),
        Stack(
          clipBehavior: Clip.none,
          children: <Widget>[
            const Icon(
              Icons.notifications_none_rounded,
              size: 31,
              color: AppColors.ink,
            ),
            Positioned(
              right: -2,
              top: -6,
              child: Container(
                width: 20,
                height: 20,
                decoration: const BoxDecoration(
                  color: Color(0xFFE94C3F),
                  shape: BoxShape.circle,
                ),
                alignment: Alignment.center,
                child: const Text(
                  '3',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ),
          ],
        ),
        const SizedBox(width: 26),
        AvatarMenuButton(
          session: session,
          onLogout: onLogout,
        ),
      ],
    );
  }
}

class HomeLoadStatus extends StatelessWidget {
  const HomeLoadStatus({
    required this.message,
    required this.onRefresh,
    super.key,
  });

  final String message;
  final VoidCallback onRefresh;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onRefresh,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        constraints: const BoxConstraints(maxWidth: 260),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: Colors.white.withOpacity(.78),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: AppColors.line),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            const Icon(
              Icons.refresh_rounded,
              size: 16,
              color: AppColors.orange,
            ),
            const SizedBox(width: 6),
            Flexible(
              child: Text(
                message,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: AppColors.body,
                  fontSize: 11,
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

class WeatherIcon extends StatelessWidget {
  const WeatherIcon({
    required this.weather,
    this.size = 58,
    super.key,
  });

  final HomeWeather weather;
  final double size;

  @override
  Widget build(BuildContext context) {
    final String condition = weather.condition.trim().toLowerCase();
    final CustomPainter painter = switch (condition) {
      'rain' => RainWeatherPainter(),
      'overcast' || 'cloudy' => CloudWeatherPainter(),
      'partly_cloudy' => PartlyCloudyWeatherPainter(),
      _ => SunPainter(),
    };
    return CustomPaint(size: Size(size, size), painter: painter);
  }
}

String _homeOrganizationName(HomeSession session) {
  final String orgName = session.orgName.trim();
  if (orgName.isNotEmpty) {
    return orgName;
  }
  final String nickName = session.nickName.trim();
  if (nickName.isNotEmpty) {
    return nickName;
  }
  return '';
}

String _homeHeaderTitle(HomeSession session) {
  final String greeting = _homeGreeting();
  final String organization = _homeOrganizationName(session);
  if (organization.isEmpty) {
    return greeting;
  }
  return '$greeting，$organization';
}

String _homeGreeting() {
  final int hour = DateTime.now().hour;
  if (hour < 12) {
    return '上午好';
  }
  if (hour < 18) {
    return '下午好';
  }
  return '晚上好';
}

String _homeHeaderSubtitle(
  String date,
  String weekday,
  HomeWeather weather,
) {
  final String dateText = _formatChineseDate(date);
  final String weekdayText = weekday.trim().isEmpty ? _weekdayLabel() : weekday;
  final String weatherText = _weatherBrief(weather);
  if (weatherText.isEmpty) {
    return '今天是 $dateText    $weekdayText';
  }
  return '今天是 $dateText    $weekdayText    $weatherText';
}

String _formatChineseDate(String rawDate) {
  final DateTime? parsed = DateTime.tryParse(rawDate.trim());
  final DateTime date = parsed ?? DateTime.now();
  return '${date.year}年${date.month}月${date.day}日';
}

String _weekdayLabel() {
  switch (DateTime.now().weekday) {
    case DateTime.monday:
      return '星期一';
    case DateTime.tuesday:
      return '星期二';
    case DateTime.wednesday:
      return '星期三';
    case DateTime.thursday:
      return '星期四';
    case DateTime.friday:
      return '星期五';
    case DateTime.saturday:
      return '星期六';
    default:
      return '星期日';
  }
}

String _weatherBrief(HomeWeather weather) {
  final String display = weather.displayName.trim();
  if (display.isEmpty) {
    return '';
  }
  if (weather.temperature == 0 && weather.source.trim().isEmpty) {
    return display;
  }
  final int temperature = weather.temperature.round();
  return '$display $temperature°C';
}

class SunPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Offset center = Offset(size.width / 2, size.height / 2);
    final Paint ray = Paint()
      ..color = const Color(0xFFF6B64E)
      ..strokeWidth = 3.2
      ..strokeCap = StrokeCap.round;
    for (int i = 0; i < 10; i++) {
      final double angle = math.pi * 2 / 10 * i;
      final Offset start =
          center + Offset(math.cos(angle), math.sin(angle)) * 22;
      final Offset end = center + Offset(math.cos(angle), math.sin(angle)) * 28;
      canvas.drawLine(start, end, ray);
    }
    canvas.drawCircle(center, 14, Paint()..color = const Color(0xFFFFC75B));
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class CloudWeatherPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint cloud = Paint()..color = const Color(0xFFAAB7C3);
    final Paint shadow = Paint()..color = const Color(0xFFDCE5EB);
    final Rect base = Rect.fromLTWH(
      size.width * .18,
      size.height * .46,
      size.width * .64,
      size.height * .26,
    );
    canvas.drawRRect(
      RRect.fromRectAndRadius(base.translate(1, 3), const Radius.circular(20)),
      shadow,
    );
    canvas.drawCircle(
        Offset(size.width * .35, size.height * .47), size.width * .18, cloud);
    canvas.drawCircle(
        Offset(size.width * .53, size.height * .40), size.width * .22, cloud);
    canvas.drawCircle(
        Offset(size.width * .67, size.height * .50), size.width * .15, cloud);
    canvas.drawRRect(
      RRect.fromRectAndRadius(base, const Radius.circular(20)),
      cloud,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class PartlyCloudyWeatherPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Offset sunCenter = Offset(size.width * .38, size.height * .36);
    final Paint ray = Paint()
      ..color = const Color(0xFFF6B64E)
      ..strokeWidth = 2.6
      ..strokeCap = StrokeCap.round;
    for (int i = 0; i < 8; i++) {
      final double angle = math.pi * 2 / 8 * i;
      final Offset start =
          sunCenter + Offset(math.cos(angle), math.sin(angle)) * 14;
      final Offset end =
          sunCenter + Offset(math.cos(angle), math.sin(angle)) * 20;
      canvas.drawLine(start, end, ray);
    }
    canvas.drawCircle(
      sunCenter,
      size.width * .18,
      Paint()..color = const Color(0xFFFFC75B),
    );
    CloudWeatherPainter().paint(canvas, size);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class RainWeatherPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    CloudWeatherPainter().paint(canvas, size);
    final Paint drop = Paint()
      ..color = const Color(0xFF6FA9D6)
      ..strokeWidth = 3
      ..strokeCap = StrokeCap.round;
    for (final Offset start in <Offset>[
      Offset(size.width * .33, size.height * .73),
      Offset(size.width * .50, size.height * .77),
      Offset(size.width * .66, size.height * .73),
    ]) {
      canvas.drawLine(
        start,
        start.translate(-size.width * .04, size.height * .11),
        drop,
      );
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

enum _AvatarAction { logout }

class AvatarMenuButton extends StatelessWidget {
  const AvatarMenuButton({
    required this.session,
    required this.onLogout,
    super.key,
  });

  final HomeSession session;
  final VoidCallback onLogout;

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<_AvatarAction>(
      offset: const Offset(0, 58),
      elevation: 10,
      color: Colors.white,
      surfaceTintColor: Colors.transparent,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      onSelected: (_AvatarAction action) {
        switch (action) {
          case _AvatarAction.logout:
            onLogout();
        }
      },
      itemBuilder: (BuildContext context) {
        return <PopupMenuEntry<_AvatarAction>>[
          PopupMenuItem<_AvatarAction>(
            enabled: false,
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 10),
            child: SizedBox(
              width: 190,
              child: Row(
                children: <Widget>[
                  PersonAvatar(size: 34, imageUrl: session.avatar),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: <Widget>[
                        Text(
                          session.nickName.trim().isEmpty
                              ? '当前账号'
                              : session.nickName.trim(),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            color: AppColors.ink,
                            fontSize: 13,
                            fontWeight: FontWeight.w800,
                          ),
                        ),
                        const SizedBox(height: 3),
                        Text(
                          _homeOrganizationName(session),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            color: AppColors.body,
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
          const PopupMenuDivider(height: 1),
          const PopupMenuItem<_AvatarAction>(
            value: _AvatarAction.logout,
            padding: EdgeInsets.symmetric(horizontal: 16, vertical: 4),
            child: Row(
              children: <Widget>[
                Icon(
                  Icons.logout_rounded,
                  size: 18,
                  color: AppColors.orange,
                ),
                SizedBox(width: 9),
                Text(
                  '退出登录',
                  style: TextStyle(
                    color: AppColors.ink,
                    fontSize: 13,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
          ),
        ];
      },
      child: PersonAvatar(size: 52, imageUrl: session.avatar),
    );
  }
}

class PersonAvatar extends StatelessWidget {
  const PersonAvatar({required this.size, this.imageUrl = '', super.key});

  final double size;
  final String imageUrl;

  @override
  Widget build(BuildContext context) {
    final String avatarUrl = imageUrl.trim();
    return Container(
      width: size,
      height: size,
      decoration: BoxDecoration(
        color: const Color(0xFFFFE1C9),
        shape: BoxShape.circle,
        boxShadow: _softShadow(
          color: const Color(0x18B65C3A),
          blur: 14,
          offset: const Offset(0, 7),
        ),
      ),
      clipBehavior: Clip.antiAlias,
      child: avatarUrl.isNotEmpty
          ? Image.network(
              avatarUrl,
              fit: BoxFit.cover,
              errorBuilder:
                  (BuildContext context, Object error, StackTrace? stackTrace) {
                return _FallbackAvatarGraphic(size: size);
              },
            )
          : _FallbackAvatarGraphic(size: size),
    );
  }
}

class _FallbackAvatarGraphic extends StatelessWidget {
  const _FallbackAvatarGraphic({required this.size});

  final double size;

  @override
  Widget build(BuildContext context) {
    return Stack(
      alignment: Alignment.center,
      children: <Widget>[
        Positioned(
          top: size * .18,
          child: Container(
            width: size * .55,
            height: size * .46,
            decoration: const BoxDecoration(
              color: Color(0xFF74402E),
              borderRadius: BorderRadius.vertical(top: Radius.circular(22)),
            ),
          ),
        ),
        Positioned(
          top: size * .27,
          child: Container(
            width: size * .45,
            height: size * .43,
            decoration: const BoxDecoration(
              color: Color(0xFFFFC9A8),
              shape: BoxShape.circle,
            ),
          ),
        ),
        Positioned(
          bottom: size * .05,
          child: Container(
            width: size * .58,
            height: size * .22,
            decoration: const BoxDecoration(
              color: Color(0xFFE98A66),
              borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
            ),
          ),
        ),
        Positioned(
          top: size * .47,
          left: size * .35,
          child: Container(
              width: 3,
              height: 3,
              decoration: const BoxDecoration(
                  color: AppColors.ink, shape: BoxShape.circle)),
        ),
        Positioned(
          top: size * .47,
          right: size * .35,
          child: Container(
              width: 3,
              height: 3,
              decoration: const BoxDecoration(
                  color: AppColors.ink, shape: BoxShape.circle)),
        ),
      ],
    );
  }
}

class StartAssessmentCard extends StatelessWidget {
  const StartAssessmentCard({this.width = 430, super.key});

  final double width;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      behavior: HitTestBehavior.opaque,
      onTap: () => Navigator.of(context).pushNamed('/assessment-scales'),
      child: Container(
        width: width,
        height: 334,
        padding: const EdgeInsets.fromLTRB(30, 34, 30, 28),
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            colors: <Color>[Color(0xFFF28A58), Color(0xFFD85C36)],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: BorderRadius.circular(18),
          boxShadow: _softShadow(
            color: const Color(0x30C85C38),
            blur: 22,
            offset: const Offset(0, 14),
          ),
        ),
        child: Stack(
          children: <Widget>[
            Positioned(
              right: -10,
              bottom: 26,
              child: ClipOval(
                child: Container(
                  width: 140,
                  height: 58,
                  color: const Color(0x22FFFFFF),
                ),
              ),
            ),
            const Positioned(
              right: 6,
              bottom: 46,
              child: ClipboardPencilIllustration(),
            ),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                const Text(
                  '开始测评',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 33,
                    fontWeight: FontWeight.w800,
                    height: 1,
                  ),
                ),
                const SizedBox(height: 17),
                const Text(
                  '科学评估 · 全面了解 · 助力成长',
                  style: TextStyle(
                    color: Color(0xFFFFE9DA),
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const Spacer(),
                _StartButton(
                  icon: Icons.add_rounded,
                  label: '新建测评',
                  filled: true,
                  onTap: () =>
                      Navigator.of(context).pushNamed('/assessment-scales'),
                ),
                const SizedBox(height: 16),
                _StartButton(
                  icon: Icons.play_arrow_rounded,
                  label: '继续测评',
                  filled: false,
                  onTap: () =>
                      Navigator.of(context).pushNamed('/assessment-scales'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _StartButton extends StatelessWidget {
  const _StartButton({
    required this.icon,
    required this.label,
    required this.filled,
    required this.onTap,
  });

  final IconData icon;
  final String label;
  final bool filled;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: 174,
        height: 50,
        decoration: BoxDecoration(
          color: filled ? Colors.white : Colors.transparent,
          borderRadius: BorderRadius.circular(12),
          border: filled ? null : Border.all(color: Colors.white, width: 1.6),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            Icon(
              icon,
              size: 23,
              color: filled ? AppColors.orangeDeep : Colors.white,
            ),
            const SizedBox(width: 9),
            Text(
              label,
              style: TextStyle(
                color: filled ? AppColors.orangeDeep : Colors.white,
                fontSize: 16,
                fontWeight: FontWeight.w800,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class ClipboardPencilIllustration extends StatelessWidget {
  const ClipboardPencilIllustration({super.key});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 158,
      height: 168,
      child: Stack(
        clipBehavior: Clip.none,
        children: <Widget>[
          Positioned(
            left: 26,
            top: 23,
            child: Transform.rotate(
              angle: .11,
              child: Container(
                width: 94,
                height: 128,
                padding: const EdgeInsets.fromLTRB(16, 24, 10, 12),
                decoration: BoxDecoration(
                  color: const Color(0xFFFFE0C0),
                  borderRadius: BorderRadius.circular(18),
                  boxShadow: _softShadow(
                    color: const Color(0x25A64B25),
                    blur: 18,
                    offset: const Offset(0, 12),
                  ),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    _clipLine(45),
                    const SizedBox(height: 13),
                    Row(
                      children: <Widget>[
                        const Icon(Icons.check_circle_rounded,
                            size: 18, color: Color(0xFFFFB769)),
                        const SizedBox(width: 9),
                        _clipLine(37),
                      ],
                    ),
                    const SizedBox(height: 13),
                    _clipLine(58),
                    const SizedBox(height: 11),
                    _clipLine(42),
                  ],
                ),
              ),
            ),
          ),
          Positioned(
            left: 57,
            top: 9,
            child: Transform.rotate(
              angle: .11,
              child: Container(
                width: 58,
                height: 22,
                decoration: BoxDecoration(
                  color: const Color(0xFFF6B26B),
                  borderRadius: BorderRadius.circular(7),
                  border: Border.all(color: const Color(0xFFFFD6AA), width: 2),
                ),
              ),
            ),
          ),
          Positioned(
            right: 14,
            bottom: 30,
            child: Transform.rotate(
              angle: .64,
              child: Column(
                children: <Widget>[
                  Container(
                    width: 23,
                    height: 84,
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFBE66),
                      borderRadius: BorderRadius.circular(9),
                    ),
                  ),
                  ClipPath(
                    clipper: TriangleClipper(),
                    child: Container(
                      width: 23,
                      height: 21,
                      color: const Color(0xFF6B3A2A),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  static Widget _clipLine(double width) {
    return Container(
      width: width,
      height: 7,
      decoration: BoxDecoration(
        color: const Color(0xFFE59E70).withOpacity(.38),
        borderRadius: BorderRadius.circular(6),
      ),
    );
  }
}

class TriangleClipper extends CustomClipper<Path> {
  @override
  Path getClip(Size size) {
    return Path()
      ..moveTo(0, 0)
      ..lineTo(size.width, 0)
      ..lineTo(size.width / 2, size.height)
      ..close();
  }

  @override
  bool shouldReclip(covariant CustomClipper<Path> oldClipper) => false;
}

class HomeStatsRow extends StatelessWidget {
  const HomeStatsRow({
    required this.width,
    required this.stats,
    this.cardWidth = 226,
    this.spacing = 18,
    super.key,
  });

  final double width;
  final HomeAssessmentStats stats;
  final double cardWidth;
  final double spacing;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: width,
      height: _homeStatsHeight,
      child: Row(
        children: <Widget>[
          Expanded(
            child: StatusSummaryCard(
              label: '在读学员',
              number: '${stats.enrolledStudents}',
              icon: Icons.group_rounded,
              tint: const Color(0xFFFFE4D3),
              iconColor: const Color(0xFFE9905E),
              width: cardWidth,
            ),
          ),
          SizedBox(width: spacing),
          Expanded(
            child: StatusSummaryCard(
              label: '评估进行中',
              number: '${stats.inProgressDrafts}',
              icon: Icons.edit_note_rounded,
              tint: const Color(0xFFFFF0C7),
              iconColor: const Color(0xFFE4AD42),
              width: cardWidth,
            ),
          ),
          SizedBox(width: spacing),
          Expanded(
            child: StatusSummaryCard(
              label: '待生成IEP',
              number: '${stats.pendingIep}',
              icon: Icons.edit_document,
              tint: const Color(0xFFEAF2E1),
              iconColor: const Color(0xFF9AB986),
              width: cardWidth,
            ),
          ),
        ],
      ),
    );
  }
}

class StatusSummaryCard extends StatelessWidget {
  const StatusSummaryCard({
    required this.label,
    required this.number,
    required this.icon,
    required this.tint,
    required this.iconColor,
    this.width = 226,
    super.key,
  });

  final String label;
  final String number;
  final IconData icon;
  final Color tint;
  final Color iconColor;
  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 102,
      padding: const EdgeInsets.fromLTRB(18, 18, 16, 14),
      decoration: BoxDecoration(
        color: tint.withOpacity(.82),
        borderRadius: BorderRadius.circular(16),
        boxShadow: _softShadow(
          color: const Color(0x13B05F32),
          blur: 18,
          offset: const Offset(0, 9),
        ),
      ),
      child: Row(
        children: <Widget>[
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Text(
                  label,
                  style: TextStyle(
                    color: iconColor.withOpacity(.88),
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const Spacer(),
                Text(
                  number,
                  style: const TextStyle(
                    color: AppColors.ink,
                    fontSize: 32,
                    fontWeight: FontWeight.w800,
                    height: 1,
                  ),
                ),
              ],
            ),
          ),
          Container(
            width: 43,
            height: 43,
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(.52),
              shape: BoxShape.circle,
            ),
            child: Icon(icon, color: iconColor, size: 26),
          ),
        ],
      ),
    );
  }
}

class ScheduleCard extends StatelessWidget {
  const ScheduleCard({
    required this.items,
    this.loading = false,
    this.width = 726,
    this.height = 224,
    super.key,
  });

  final List<HomeScheduleItem> items;
  final bool loading;
  final double width;
  final double height;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      padding: const EdgeInsets.fromLTRB(24, 22, 24, 18),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.9),
        borderRadius: BorderRadius.circular(18),
        boxShadow: _softShadow(
          color: const Color(0x13B05F32),
          blur: 18,
          offset: const Offset(0, 9),
        ),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: const <Widget>[
              Text(
                '今日排课',
                style: TextStyle(
                  color: AppColors.ink,
                  fontSize: 17,
                  fontWeight: FontWeight.w800,
                ),
              ),
              Spacer(),
              Text(
                '查看全部 〉',
                style: TextStyle(
                  color: Color(0xFFD79B78),
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          if (loading && items.isEmpty)
            const ScheduleSkeletonList()
          else if (items.isEmpty)
            const Expanded(child: ScheduleEmptyState())
          else
            ...items.asMap().entries.map(
                  (MapEntry<int, HomeScheduleItem> entry) => ScheduleRow(
                    color: _scheduleDotColor(entry.key),
                    time: entry.value.time,
                    title: entry.value.title,
                    place: entry.value.place,
                    state: entry.value.state,
                    stateColor: _scheduleStateColor(entry.value.state),
                    stateBg: _scheduleStateBg(entry.value.state),
                    isLast: entry.key == items.length - 1,
                  ),
                ),
        ],
      ),
    );
  }
}

class ScheduleSkeletonList extends StatelessWidget {
  const ScheduleSkeletonList({super.key});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: List<Widget>.generate(
        4,
        (int index) => ScheduleSkeletonRow(
          key: ValueKey<String>('schedule-skeleton-$index'),
          isLast: index == 3,
        ),
      ),
    );
  }
}

class ScheduleSkeletonRow extends StatelessWidget {
  const ScheduleSkeletonRow({
    this.isLast = false,
    super.key,
  });

  final bool isLast;

  static const double _rowHeight = 34;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: _rowHeight,
      child: Row(
        children: <Widget>[
          SizedBox(
            width: 20,
            height: _rowHeight,
            child: Stack(
              clipBehavior: Clip.none,
              alignment: Alignment.center,
              children: <Widget>[
                if (!isLast)
                  Positioned(
                    top: _rowHeight / 2,
                    bottom: -_rowHeight / 2,
                    child: Container(
                      width: 2,
                      decoration: BoxDecoration(
                        color: const Color(0xFFEFE9E1),
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ),
                  ),
                const _SkeletonBlock(width: 11, height: 11, radius: 99),
              ],
            ),
          ),
          const SizedBox(width: 10),
          const _SkeletonBlock(width: 48, height: 13, radius: 5),
          const SizedBox(width: 10),
          const Expanded(
            child: Align(
              alignment: Alignment.centerLeft,
              child: FractionallySizedBox(
                widthFactor: .78,
                child: _SkeletonBlock(height: 14, radius: 6),
              ),
            ),
          ),
          const SizedBox(width: 14),
          const _SkeletonBlock(width: 74, height: 13, radius: 5),
          const SizedBox(width: 14),
          const _SkeletonBlock(width: 63, height: 25, radius: 8),
        ],
      ),
    );
  }
}

class _SkeletonBlock extends StatelessWidget {
  const _SkeletonBlock({
    this.width,
    required this.height,
    required this.radius,
  });

  final double? width;
  final double height;
  final double radius;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        color: const Color(0xFFF0E9E2),
        borderRadius: BorderRadius.circular(radius),
      ),
    );
  }
}

class ScheduleEmptyState extends StatelessWidget {
  const ScheduleEmptyState({super.key});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: const <Widget>[
          Icon(
            Icons.event_available_rounded,
            size: 23,
            color: AppColors.green,
          ),
          SizedBox(width: 8),
          Text(
            '今日暂无排课',
            style: TextStyle(
              color: AppColors.body,
              fontSize: 13,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}

Color _scheduleDotColor(int index) {
  const List<Color> colors = <Color>[
    AppColors.orange,
    AppColors.yellow,
    AppColors.green,
    AppColors.blueGray,
  ];
  return colors[index % colors.length];
}

Color _scheduleStateColor(String state) {
  switch (state) {
    case '进行中':
    case '即将开始':
      return const Color(0xFFE87042);
    case '已点名':
      return const Color(0xFF6FA477);
    default:
      return const Color(0xFF9E9A96);
  }
}

Color _scheduleStateBg(String state) {
  switch (state) {
    case '进行中':
    case '即将开始':
      return const Color(0xFFFFEEE5);
    case '已点名':
      return const Color(0xFFEAF3E7);
    default:
      return const Color(0xFFF3F0EC);
  }
}

class ScheduleRow extends StatelessWidget {
  const ScheduleRow({
    required this.color,
    required this.time,
    required this.title,
    required this.place,
    required this.state,
    required this.stateColor,
    required this.stateBg,
    this.isLast = false,
    super.key,
  });

  final Color color;
  final String time;
  final String title;
  final String place;
  final String state;
  final Color stateColor;
  final Color stateBg;
  final bool isLast;

  static const double _rowHeight = 34;
  static const double _dotSize = 11;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: _rowHeight,
      child: Row(
        children: <Widget>[
          SizedBox(
            width: 20,
            height: _rowHeight,
            child: Stack(
              clipBehavior: Clip.none,
              alignment: Alignment.center,
              children: <Widget>[
                if (!isLast)
                  Positioned(
                    top: _rowHeight / 2,
                    bottom: -_rowHeight / 2,
                    child: Container(
                      width: 2,
                      decoration: BoxDecoration(
                        color: color.withOpacity(.34),
                        borderRadius: BorderRadius.circular(2),
                      ),
                    ),
                  ),
                Container(
                  width: _dotSize,
                  height: _dotSize,
                  decoration:
                      BoxDecoration(color: color, shape: BoxShape.circle),
                ),
              ],
            ),
          ),
          const SizedBox(width: 10),
          SizedBox(
            width: 54,
            child: Text(
              time,
              style: const TextStyle(
                color: AppColors.ink,
                fontSize: 13,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          Expanded(
            child: Text(
              title,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: AppColors.ink,
                fontSize: 13,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
          SizedBox(
            width: 88,
            child: Text(
              place,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: Color(0xFF95867E),
                fontSize: 12,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          Container(
            width: 63,
            height: 25,
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: stateBg,
              borderRadius: BorderRadius.circular(8),
            ),
            child: Text(
              state,
              style: TextStyle(
                color: stateColor,
                fontSize: 11,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class FeatureShortcutRow extends StatelessWidget {
  const FeatureShortcutRow({
    this.cardWidth = 187,
    this.spacing = 14,
    this.onTimetableTap,
    this.onReportTap,
    this.onTrainingTap,
    super.key,
  });

  final double cardWidth;
  final double spacing;
  final VoidCallback? onTimetableTap;
  final VoidCallback? onReportTap;
  final VoidCallback? onTrainingTap;

  @override
  Widget build(BuildContext context) {
    return Row(
      children: <Widget>[
        ShortcutCard(
          title: '学员档案',
          desc1: '成长记录',
          desc2: '全面管理',
          icon: Icons.badge_rounded,
          iconColor: const Color(0xFF74AA79),
          bg: const Color(0xFFEFF7EC),
          width: cardWidth,
        ),
        SizedBox(width: spacing),
        ShortcutCard(
          title: '评估报告',
          desc1: '测评结果',
          desc2: '快速查看',
          icon: Icons.article_outlined,
          iconColor: const Color(0xFF3F82D2),
          bg: const Color(0xFFEDF5FF),
          width: cardWidth,
          onTap: onReportTap,
        ),
        SizedBox(width: spacing),
        ShortcutCard(
          title: 'IEP中心',
          desc1: '个训计划',
          desc2: '智能生成',
          icon: Icons.assignment_rounded,
          iconColor: const Color(0xFFE87952),
          bg: const Color(0xFFFFF0E6),
          width: cardWidth,
        ),
        SizedBox(width: spacing),
        ShortcutCard(
          title: '督导管理',
          desc1: '专业督导',
          desc2: '闭环跟进',
          icon: Icons.supervisor_account_rounded,
          iconColor: const Color(0xFF8C6DD8),
          bg: const Color(0xFFF1EDFF),
          width: cardWidth,
        ),
        SizedBox(width: spacing),
        ShortcutCard(
          title: '排课日程',
          desc1: '课程安排',
          desc2: '一目了然',
          icon: Icons.calendar_month_rounded,
          iconColor: const Color(0xFFE87B52),
          bg: const Color(0xFFFFF5EA),
          width: cardWidth,
          onTap: onTimetableTap,
        ),
        SizedBox(width: spacing),
        ShortcutCard(
          title: '训练中心',
          desc1: '训练动态',
          desc2: '即时掌握',
          icon: Icons.sports_esports_rounded,
          iconColor: const Color(0xFF4D9C8E),
          bg: const Color(0xFFE8F7F3),
          width: cardWidth,
          onTap: onTrainingTap,
        ),
      ],
    );
  }
}

class ShortcutCard extends StatelessWidget {
  const ShortcutCard({
    required this.title,
    required this.desc1,
    required this.desc2,
    required this.icon,
    required this.iconColor,
    required this.bg,
    this.badge,
    this.width = 187,
    this.onTap,
    super.key,
  });

  final String title;
  final String desc1;
  final String desc2;
  final IconData icon;
  final Color iconColor;
  final Color bg;
  final String? badge;
  final double width;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final BorderRadius borderRadius = BorderRadius.circular(16);
    return Material(
      color: Colors.white.withOpacity(.88),
      borderRadius: borderRadius,
      elevation: 0,
      child: InkWell(
        onTap: onTap,
        borderRadius: borderRadius,
        child: Container(
          width: width,
          height: 118,
          padding: const EdgeInsets.fromLTRB(17, 16, 14, 12),
          decoration: BoxDecoration(
            borderRadius: borderRadius,
            boxShadow: _softShadow(
              color: const Color(0x12B05F32),
              blur: 16,
              offset: const Offset(0, 8),
            ),
          ),
          child: Stack(
            children: <Widget>[
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  Text(
                    title,
                    style: const TextStyle(
                      color: AppColors.ink,
                      fontSize: 16,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    desc1,
                    style: const TextStyle(
                      color: AppColors.body,
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 3),
                  Text(
                    desc2,
                    style: const TextStyle(
                      color: AppColors.body,
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
              Positioned(
                right: 0,
                bottom: 0,
                child: Stack(
                  clipBehavior: Clip.none,
                  children: <Widget>[
                    Container(
                      width: 54,
                      height: 54,
                      decoration: BoxDecoration(
                        color: bg,
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: Icon(icon, color: iconColor, size: 32),
                    ),
                    if (badge != null)
                      Positioned(
                        right: -5,
                        top: -7,
                        child: Container(
                          width: 23,
                          height: 23,
                          decoration: const BoxDecoration(
                            color: Color(0xFFE8463A),
                            shape: BoxShape.circle,
                          ),
                          alignment: Alignment.center,
                          child: Text(
                            badge!,
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 12,
                              fontWeight: FontWeight.w800,
                            ),
                          ),
                        ),
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

class ProgressOverviewCard extends StatelessWidget {
  const ProgressOverviewCard({
    required this.stats,
    this.width = 820,
    super.key,
  });

  final HomeAssessmentStats stats;
  final double width;

  @override
  Widget build(BuildContext context) {
    final int total = stats.total <= 0 ? 1 : stats.total;
    final int recordTotal =
        stats.completedRecords <= 0 ? 1 : stats.completedRecords;
    final double progress = stats.coverageRate.clamp(0.0, 1.0).toDouble();
    final int progressText = (progress * 100).round();
    return Container(
      width: width,
      height: 132,
      padding: const EdgeInsets.fromLTRB(22, 17, 18, 16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.9),
        borderRadius: BorderRadius.circular(18),
        boxShadow: _softShadow(
          color: const Color(0x12B05F32),
          blur: 16,
          offset: const Offset(0, 8),
        ),
      ),
      child: Column(
        children: <Widget>[
          Row(
            children: const <Widget>[
              Text(
                '在读学员测评覆盖率',
                style: TextStyle(
                  color: AppColors.ink,
                  fontSize: 16,
                  fontWeight: FontWeight.w800,
                ),
              ),
              Spacer(),
              Text(
                '学员测评 〉',
                style: TextStyle(
                  color: Color(0xFFD79B78),
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: <Widget>[
              SizedBox(
                width: 58,
                height: 58,
                child: Stack(
                  alignment: Alignment.center,
                  children: <Widget>[
                    CustomPaint(
                      size: const Size(58, 58),
                      painter: DonutPainter(progress: progress),
                    ),
                    Text(
                      '$progressText%',
                      style: const TextStyle(
                        color: AppColors.ink,
                        fontSize: 15,
                        fontWeight: FontWeight.w800,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 16),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: <Widget>[
                  const Text(
                    '测评覆盖率',
                    style: TextStyle(
                      color: AppColors.ink,
                      fontSize: 13,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(height: 7),
                  Text(
                    '已测评 ${stats.assessedStudents}/${stats.enrolledStudents}',
                    style: const TextStyle(
                      color: AppColors.orange,
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ],
              ),
              const SizedBox(width: 28),
              Expanded(
                child: Row(
                  children: <Widget>[
                    Expanded(
                      child: ProgressStat(
                          color: AppColors.blueGray,
                          label: '未测评',
                          number: '${stats.unassessedStudents}',
                          desc:
                              '占比 ${_percent(stats.unassessedStudents, total)}%'),
                    ),
                    Expanded(
                      child: ProgressStat(
                          color: AppColors.orange,
                          label: '已测评',
                          number: '${stats.assessedStudents}',
                          desc:
                              '占比 ${_percent(stats.assessedStudents, total)}%'),
                    ),
                    Expanded(
                      child: ProgressStat(
                          color: AppColors.green,
                          label: '待生成IEP',
                          number: '${stats.pendingIep}',
                          desc:
                              '占记录 ${_percent(stats.pendingIep, recordTotal)}%'),
                    ),
                    Expanded(
                      child: ProgressStat(
                          color: AppColors.yellow,
                          label: '已出IEP',
                          number: '${stats.generatedIep}',
                          desc:
                              '占记录 ${_percent(stats.generatedIep, recordTotal)}%'),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

int _percent(int value, int total) {
  if (total <= 0) {
    return 0;
  }
  return (value / total * 100).round();
}

class DonutPainter extends CustomPainter {
  const DonutPainter({required this.progress});

  final double progress;

  @override
  void paint(Canvas canvas, Size size) {
    final Rect rect = Offset.zero & size;
    final Paint bg = Paint()
      ..color = const Color(0xFFEFE9E1)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 9
      ..strokeCap = StrokeCap.round;
    final Paint orange = Paint()
      ..color = AppColors.orange
      ..style = PaintingStyle.stroke
      ..strokeWidth = 9
      ..strokeCap = StrokeCap.round;
    canvas.drawArc(rect.deflate(7), 0, math.pi * 2, false, bg);
    canvas.drawArc(
        rect.deflate(7), -math.pi / 2, math.pi * 2 * progress, false, orange);
  }

  @override
  bool shouldRepaint(covariant DonutPainter oldDelegate) =>
      oldDelegate.progress != progress;
}

class ProgressStat extends StatelessWidget {
  const ProgressStat({
    required this.color,
    required this.label,
    required this.number,
    required this.desc,
    super.key,
  });

  final Color color;
  final String label;
  final String number;
  final String desc;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Container(
                  width: 6,
                  height: 6,
                  decoration:
                      BoxDecoration(color: color, shape: BoxShape.circle)),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  label,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: AppColors.body,
                    fontSize: 10,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            number,
            style: const TextStyle(
              color: AppColors.ink,
              fontSize: 18,
              fontWeight: FontWeight.w800,
              height: 1,
            ),
          ),
          const SizedBox(height: 5),
          Text(
            desc,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(
              color: AppColors.body,
              fontSize: 10,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }
}

class MorePlansCard extends StatelessWidget {
  const MorePlansCard({this.width = 360, super.key});

  final double width;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: 132,
      padding: const EdgeInsets.fromLTRB(20, 17, 18, 14),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.9),
        borderRadius: BorderRadius.circular(18),
        boxShadow: _softShadow(
          color: const Color(0x12B05F32),
          blur: 16,
          offset: const Offset(0, 8),
        ),
      ),
      child: Stack(
        children: <Widget>[
          const Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: <Widget>[
              Text(
                '了解更多测评方案',
                style: TextStyle(
                  color: AppColors.ink,
                  fontSize: 16,
                  fontWeight: FontWeight.w800,
                ),
              ),
              SizedBox(height: 8),
              Text(
                '为不同需求提供专业支持',
                style: TextStyle(
                  color: AppColors.body,
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
          Positioned(
            left: 0,
            bottom: 0,
            child: Container(
              width: 91,
              height: 34,
              decoration: BoxDecoration(
                color: AppColors.orangeDeep,
                borderRadius: BorderRadius.circular(18),
              ),
              alignment: Alignment.center,
              child: const Text(
                '去探索  〉',
                style: TextStyle(
                  color: Colors.white,
                  fontSize: 13,
                  fontWeight: FontWeight.w800,
                ),
              ),
            ),
          ),
          const Positioned(right: 0, bottom: -4, child: AdvisorGraphic()),
        ],
      ),
    );
  }
}

class AdvisorGraphic extends StatelessWidget {
  const AdvisorGraphic({super.key});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: 126,
      height: 98,
      child: Stack(
        children: <Widget>[
          Positioned(
            right: 0,
            bottom: 0,
            child: Container(
              width: 92,
              height: 22,
              decoration: BoxDecoration(
                color: const Color(0xFFE8B18A).withOpacity(.35),
                borderRadius: BorderRadius.circular(18),
              ),
            ),
          ),
          Positioned(
            right: 20,
            top: 4,
            child: Container(
              width: 43,
              height: 48,
              decoration: const BoxDecoration(
                color: Color(0xFF75402E),
                borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
              ),
            ),
          ),
          Positioned(
            right: 25,
            top: 15,
            child: Container(
              width: 34,
              height: 35,
              decoration: const BoxDecoration(
                color: Color(0xFFFFC6A5),
                shape: BoxShape.circle,
              ),
            ),
          ),
          Positioned(
            right: 11,
            bottom: 7,
            child: Container(
              width: 68,
              height: 43,
              decoration: const BoxDecoration(
                color: Color(0xFFFFF0E2),
                borderRadius: BorderRadius.vertical(top: Radius.circular(22)),
              ),
            ),
          ),
          Positioned(
            right: 43,
            bottom: 8,
            child: Transform.rotate(
              angle: -.24,
              child: Container(
                width: 54,
                height: 36,
                decoration: BoxDecoration(
                  color: const Color(0xFFCFE7BF),
                  borderRadius: BorderRadius.circular(8),
                  boxShadow: _softShadow(
                    color: const Color(0x1683A46F),
                    blur: 10,
                    offset: const Offset(0, 5),
                  ),
                ),
              ),
            ),
          ),
          Positioned(
            left: 12,
            bottom: 4,
            child: CustomPaint(
                size: const Size(42, 48), painter: SmallLeavesPainter()),
          ),
        ],
      ),
    );
  }
}

class SmallLeavesPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint stem = Paint()
      ..color = const Color(0xFF87A969)
      ..strokeWidth = 3
      ..strokeCap = StrokeCap.round;
    canvas.drawLine(
        Offset(size.width / 2, size.height), Offset(size.width / 2, 10), stem);
    _leaf(canvas, Offset(20, 29), const Size(22, 11), -.6);
    _leaf(canvas, Offset(22, 21), const Size(24, 12), .25);
    _leaf(canvas, Offset(19, 13), const Size(18, 9), -.45);
  }

  void _leaf(Canvas canvas, Offset center, Size size, double angle) {
    canvas.save();
    canvas.translate(center.dx, center.dy);
    canvas.rotate(angle);
    canvas.drawOval(
      Rect.fromCenter(
          center: Offset.zero, width: size.width, height: size.height),
      Paint()..color = const Color(0xFF9CBD78),
    );
    canvas.restore();
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
