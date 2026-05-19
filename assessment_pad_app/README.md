# 测评云端 Pad App

Flutter 横屏 Pad 静态页项目，当前包含：

- 登录页
- 首页/机构工作台
- Android、iOS、Web、macOS、Windows 工程模板

## 运行

```bash
flutter pub get
flutter run -d macos
```

登录页点击「登录」会进入首页。

## 打包

macOS：

```bash
flutter build macos
```

Windows `.exe` 需要在 Windows 环境中执行：

```bash
flutter build windows
```

Windows 安装包需要在 Windows 环境安装 Flutter、Visual Studio C++ 桌面开发组件和 Inno Setup 6，然后执行：

```powershell
.\scripts\build_windows_installer.ps1 -Target prod -BuildMode release
```

当前页面代码集中在 `lib/main.dart`，后续接接口时可以再拆分为 `screens/`、`widgets/` 和 `models/`。
