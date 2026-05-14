import 'dart:async';

import 'package:flutter/material.dart';

enum PadMessageTone {
  info,
  success,
  error,
}

extension PadMessageToneView on PadMessageTone {
  IconData get icon {
    switch (this) {
      case PadMessageTone.info:
        return Icons.info_outline_rounded;
      case PadMessageTone.success:
        return Icons.check_circle_outline_rounded;
      case PadMessageTone.error:
        return Icons.error_outline_rounded;
    }
  }

  Color get foreground {
    switch (this) {
      case PadMessageTone.info:
        return const Color(0xFFC95735);
      case PadMessageTone.success:
        return const Color(0xFF6F9F70);
      case PadMessageTone.error:
        return const Color(0xFFD92D20);
    }
  }

  Color get textColor {
    switch (this) {
      case PadMessageTone.info:
        return const Color(0xFF432B22);
      case PadMessageTone.success:
        return const Color(0xFF426D44);
      case PadMessageTone.error:
        return const Color(0xFF7A271A);
    }
  }

  Color get background {
    switch (this) {
      case PadMessageTone.info:
        return const Color(0xFFFFF8EE);
      case PadMessageTone.success:
        return const Color(0xFFF0FAEF);
      case PadMessageTone.error:
        return const Color(0xFFFFF1F0);
    }
  }

  Color get border {
    switch (this) {
      case PadMessageTone.info:
        return const Color(0xFFF0DDC9);
      case PadMessageTone.success:
        return const Color(0xFFCBEACB);
      case PadMessageTone.error:
        return const Color(0xFFFFB4AB);
    }
  }
}

class PadMessageOverlayController {
  OverlayEntry? _entry;
  Timer? _timer;
  bool _visible = false;
  String _message = '';
  PadMessageTone _tone = PadMessageTone.info;
  double _topMargin = 12;
  String _key = 'pad-top-message';

  void show(
    BuildContext context,
    String message, {
    PadMessageTone tone = PadMessageTone.info,
    Duration duration = const Duration(milliseconds: 1900),
    double topMargin = 12,
    String key = 'pad-top-message',
    bool rootOverlay = true,
  }) {
    if (message.trim().isEmpty) {
      return;
    }
    _timer?.cancel();
    _message = message.trim();
    _tone = tone;
    _topMargin = topMargin;
    _key = key;
    final OverlayState? overlay =
        Overlay.maybeOf(context, rootOverlay: rootOverlay);
    if (overlay == null) {
      return;
    }
    if (_entry == null) {
      _visible = false;
      _entry = OverlayEntry(builder: _build);
      overlay.insert(_entry!);
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (_entry == null) {
          return;
        }
        _visible = true;
        _entry?.markNeedsBuild();
      });
    } else {
      _visible = true;
      _entry!.markNeedsBuild();
    }
    _timer = Timer(duration, () {
      _visible = false;
      _entry?.markNeedsBuild();
      _timer = Timer(const Duration(milliseconds: 180), dispose);
    });
  }

  Widget _build(BuildContext context) {
    return PadTopMessage(
      visible: _visible,
      message: _message,
      tone: _tone,
      topMargin: _topMargin,
      messageKey: ValueKey<String>(_key),
    );
  }

  void dispose() {
    _timer?.cancel();
    _timer = null;
    _entry?.remove();
    _entry = null;
  }
}

class PadTopMessage extends StatelessWidget {
  const PadTopMessage({
    required this.visible,
    required this.message,
    this.tone = PadMessageTone.info,
    this.topMargin = 12,
    this.messageKey,
    super.key,
  });

  final bool visible;
  final String message;
  final PadMessageTone tone;
  final double topMargin;
  final Key? messageKey;

  @override
  Widget build(BuildContext context) {
    return IgnorePointer(
      child: SafeArea(
        bottom: false,
        child: Align(
          alignment: Alignment.topCenter,
          child: AnimatedSlide(
            offset: visible ? Offset.zero : const Offset(0, -.75),
            duration: const Duration(milliseconds: 180),
            curve: Curves.easeOutCubic,
            child: AnimatedOpacity(
              opacity: visible ? 1 : 0,
              duration: const Duration(milliseconds: 160),
              curve: Curves.easeOutCubic,
              child: Container(
                key: messageKey,
                constraints: const BoxConstraints(maxWidth: 680),
                margin: EdgeInsets.only(top: topMargin),
                padding: const EdgeInsets.fromLTRB(14, 11, 16, 11),
                decoration: BoxDecoration(
                  color: tone.background,
                  border: Border.all(color: tone.border),
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: const <BoxShadow>[
                    BoxShadow(
                      color: Color(0x1A4A2F22),
                      blurRadius: 18,
                      offset: Offset(0, 10),
                    ),
                  ],
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: <Widget>[
                    Padding(
                      padding: const EdgeInsets.only(top: 1),
                      child: Icon(tone.icon, color: tone.foreground, size: 18),
                    ),
                    const SizedBox(width: 8),
                    Flexible(
                      child: Text(
                        message,
                        style: TextStyle(
                          color: tone.textColor,
                          fontSize: 13,
                          height: 1.35,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
