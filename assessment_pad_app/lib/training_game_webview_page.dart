import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:webview_flutter/webview_flutter.dart';

class TrainingGameWebViewPage extends StatefulWidget {
  const TrainingGameWebViewPage({
    required this.title,
    required this.url,
    super.key,
  });

  final String title;
  final Uri url;

  @override
  State<TrainingGameWebViewPage> createState() =>
      _TrainingGameWebViewPageState();
}

class _TrainingGameWebViewPageState extends State<TrainingGameWebViewPage> {
  late final WebViewController _controller;
  int _progress = 0;
  bool _hasError = false;
  bool _gameReady = false;
  bool _openingOverlayMounted = true;
  bool _openingOverlayVisible = true;
  bool _closing = false;
  Timer? _openingOverlayRemovalTimer;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _controller = WebViewController()
      ..setJavaScriptMode(JavaScriptMode.unrestricted)
      ..setBackgroundColor(const Color(0xFF7BDFF2))
      ..addJavaScriptChannel(
        'FlutterTrainingGame',
        onMessageReceived: _handleGameMessage,
      )
      ..setNavigationDelegate(
        NavigationDelegate(
          onProgress: (int progress) {
            if (mounted) {
              setState(() => _progress = progress);
            }
          },
          onPageStarted: (_) {
            if (mounted) {
              _openingOverlayRemovalTimer?.cancel();
              setState(() {
                _hasError = false;
                _gameReady = false;
                _openingOverlayMounted = true;
                _openingOverlayVisible = true;
                _errorMessage = null;
                _progress = 0;
              });
            }
          },
          onPageFinished: (_) {
            if (mounted) {
              setState(() => _progress = 100);
            }
            _checkGameReady();
          },
          onWebResourceError: (WebResourceError error) {
            if (mounted) {
              setState(() {
                _hasError = true;
                _errorMessage = error.description;
              });
            }
          },
        ),
      )
      ..loadRequest(widget.url);
  }

  @override
  void dispose() {
    _openingOverlayRemovalTimer?.cancel();
    _stopGameAudio();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final EdgeInsets viewPadding = MediaQuery.viewPaddingOf(context);

    return Scaffold(
      backgroundColor: const Color(0xFF7BDFF2),
      body: PopScope(
        canPop: true,
        onPopInvoked: (bool didPop) {
          if (didPop) {
            unawaited(_stopGameAudio());
          }
        },
        child: Stack(
          children: <Widget>[
            Positioned.fill(child: WebViewWidget(controller: _controller)),
            if (_openingOverlayMounted && !_hasError)
              Positioned.fill(
                child: AnimatedOpacity(
                  opacity: _openingOverlayVisible ? 1 : 0,
                  duration: const Duration(milliseconds: 220),
                  curve: Curves.easeOutCubic,
                  child: _GameOpeningOverlay(progress: _progress),
                ),
              ),
            Positioned(
              left: 18 + viewPadding.left,
              top: 124 + viewPadding.top,
              child: _GameBackButton(onTap: _closePage),
            ),
            if (_progress < 100)
              Positioned(
                left: 0,
                top: viewPadding.top,
                right: 0,
                child: LinearProgressIndicator(
                  value: _progress / 100,
                  minHeight: 3,
                  backgroundColor: Colors.white.withOpacity(.18),
                  color: const Color(0xFFFF6B12),
                ),
              ),
            if (_hasError)
              Positioned.fill(
                child: _GameLoadError(
                  message: _errorMessage,
                  url: widget.url.toString(),
                  onRetry: () {
                    _stopGameAudio();
                    setState(() {
                      _hasError = false;
                      _errorMessage = null;
                    });
                    _controller.loadRequest(widget.url);
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }

  void _handleGameMessage(JavaScriptMessage message) {
    try {
      final Object? decoded = jsonDecode(message.message);
      if (decoded is! Map<String, dynamic>) {
        return;
      }
      final Object? type = decoded['type'];
      if (type == 'training-game-close' && mounted) {
        _closePage();
      } else if (type == 'training-game-ready' && mounted) {
        _markGameReady();
      } else if (type == 'training-game-error' && mounted) {
        final Object? payload = decoded['payload'];
        setState(() {
          _hasError = true;
          _errorMessage = payload is Map<String, dynamic>
              ? '${payload['message'] ?? '游戏脚本运行失败'}'
              : '游戏脚本运行失败';
        });
      }
    } catch (_) {
      return;
    }
  }

  void _closePage() {
    if (_closing) {
      return;
    }
    _closing = true;
    unawaited(_stopGameAudio());
    if (mounted) {
      Navigator.of(context).pop();
    }
  }

  Future<void> _stopGameAudio() async {
    try {
      await _controller.runJavaScript(
        'window.__COLOR_MATCH_STOP_AUDIO__ && window.__COLOR_MATCH_STOP_AUDIO__();',
      );
    } catch (_) {
      return;
    }
  }

  void _markGameReady() {
    if (_gameReady) {
      return;
    }

    _openingOverlayRemovalTimer?.cancel();
    setState(() {
      _gameReady = true;
      _progress = 100;
    });

    _openingOverlayRemovalTimer = Timer(const Duration(milliseconds: 80), () {
      if (!mounted) {
        return;
      }
      setState(() => _openingOverlayVisible = false);
      _openingOverlayRemovalTimer =
          Timer(const Duration(milliseconds: 240), () {
        if (mounted) {
          setState(() => _openingOverlayMounted = false);
        }
      });
    });
  }

  Future<void> _checkGameReady() async {
    for (int attempt = 0; attempt < 15; attempt += 1) {
      await Future<void>.delayed(const Duration(seconds: 1));
      if (!mounted || _hasError || _gameReady) {
        return;
      }

      try {
        final Object ready = await _controller.runJavaScriptReturningResult(
          'Boolean(window.__COLOR_MATCH_READY__)',
        );
        final bool isReady = ready == true || ready.toString() == 'true';
        if (isReady && mounted) {
          _markGameReady();
          return;
        }
      } catch (error) {
        if (mounted && attempt == 14) {
          setState(() {
            _hasError = true;
            _errorMessage = '无法检测游戏启动状态：$error';
          });
        }
      }
    }

    if (!mounted || _hasError || _gameReady) {
      return;
    }
    setState(() {
      _hasError = true;
      _errorMessage =
          '页面已加载，但游戏场景没有启动。请检查 H5 控制台或确认当前 WebView 支持 Canvas/JavaScript。';
    });
  }
}

class _GameOpeningOverlay extends StatelessWidget {
  const _GameOpeningOverlay({required this.progress});

  final int progress;

  @override
  Widget build(BuildContext context) {
    final double value = (progress.clamp(8, 100)) / 100;

    return DecoratedBox(
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: <Color>[
            Color(0xFF8BE7FF),
            Color(0xFFDFFAFF),
            Color(0xFF75D18C),
          ],
        ),
      ),
      child: Center(
        child: Container(
          width: 420,
          padding: const EdgeInsets.fromLTRB(34, 32, 34, 30),
          decoration: BoxDecoration(
            color: Colors.white.withOpacity(.93),
            borderRadius: BorderRadius.circular(30),
            border: Border.all(color: const Color(0xFFFFD447), width: 4),
            boxShadow: <BoxShadow>[
              BoxShadow(
                color: const Color(0xFF24546B).withOpacity(.18),
                blurRadius: 30,
                offset: const Offset(0, 14),
              ),
            ],
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const _LoadingBubbles(),
              const SizedBox(height: 18),
              const Text(
                '颜色配对乐园',
                style: TextStyle(
                  color: Color(0xFF24546B),
                  fontSize: 32,
                  fontWeight: FontWeight.w900,
                  height: 1.1,
                ),
              ),
              const SizedBox(height: 24),
              ClipRRect(
                borderRadius: BorderRadius.circular(999),
                child: LinearProgressIndicator(
                  value: value,
                  minHeight: 18,
                  backgroundColor: const Color(0xFFD8F3FB),
                  color: const Color(0xFFFF6B8A),
                ),
              ),
              const SizedBox(height: 14),
              Text(
                '游戏加载中 ${progress.clamp(0, 100)}%',
                style: const TextStyle(
                  color: Color(0xFF24546B),
                  fontSize: 18,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _LoadingBubbles extends StatelessWidget {
  const _LoadingBubbles();

  @override
  Widget build(BuildContext context) {
    const List<Color> colors = <Color>[
      Color(0xFFFF5A65),
      Color(0xFFFFD447),
      Color(0xFF4E9CFF),
      Color(0xFF54CE73),
    ];

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: colors
          .map(
            (Color color) => Container(
              width: 38,
              height: 38,
              margin: const EdgeInsets.symmetric(horizontal: 7),
              decoration: BoxDecoration(
                color: color,
                shape: BoxShape.circle,
                border:
                    Border.all(color: Colors.white.withOpacity(.72), width: 4),
                boxShadow: <BoxShadow>[
                  BoxShadow(
                    color: const Color(0xFF24546B).withOpacity(.12),
                    offset: const Offset(4, 6),
                  ),
                ],
              ),
            ),
          )
          .toList(),
    );
  }
}

class _GameBackButton extends StatelessWidget {
  const _GameBackButton({required this.onTap});

  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return DecoratedBox(
      decoration: BoxDecoration(
        boxShadow: <BoxShadow>[
          BoxShadow(
            color: const Color(0xFF24546B).withOpacity(.16),
            blurRadius: 14,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: Material(
        color: Colors.white.withOpacity(.9),
        shape: const CircleBorder(),
        clipBehavior: Clip.antiAlias,
        child: InkWell(
          customBorder: const CircleBorder(),
          onTap: onTap,
          child: const SizedBox(
            width: 48,
            height: 48,
            child: Icon(
              Icons.arrow_back_rounded,
              color: Color(0xFF24546B),
              size: 28,
            ),
          ),
        ),
      ),
    );
  }
}

class _GameLoadError extends StatelessWidget {
  const _GameLoadError({
    required this.onRetry,
    required this.url,
    this.message,
  });

  final VoidCallback onRetry;
  final String url;
  final String? message;

  @override
  Widget build(BuildContext context) {
    return ColoredBox(
      color: const Color(0xFF7BDFF2),
      child: Center(
        child: Container(
          width: 420,
          padding: const EdgeInsets.fromLTRB(26, 24, 26, 22),
          decoration: BoxDecoration(
            color: Colors.white.withOpacity(.96),
            borderRadius: BorderRadius.circular(18),
            border: Border.all(color: const Color(0xFFEAD7C9)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: <Widget>[
              const Icon(
                Icons.wifi_off_rounded,
                size: 42,
                color: Color(0xFFFF6B12),
              ),
              const SizedBox(height: 12),
              const Text(
                '游戏加载失败',
                style: TextStyle(
                  color: Color(0xFF24546B),
                  fontSize: 22,
                  fontWeight: FontWeight.w900,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                message ?? '请确认游戏服务已启动，或检查 Pad 能否访问游戏地址。',
                textAlign: TextAlign.center,
                style: TextStyle(
                  color: Color(0xFF607487),
                  fontSize: 14,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 10),
              Text(
                url,
                textAlign: TextAlign.center,
                maxLines: 3,
                overflow: TextOverflow.ellipsis,
                style: const TextStyle(
                  color: Color(0xFF7C8B96),
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 18),
              ElevatedButton(
                onPressed: onRetry,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFFFF6B12),
                  foregroundColor: Colors.white,
                  fixedSize: const Size(132, 40),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
                child: const Text(
                  '重新加载',
                  style: TextStyle(fontWeight: FontWeight.w900),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
