import 'dart:math' as math;

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class AssessmentScaleCategoryScreen extends StatefulWidget {
  const AssessmentScaleCategoryScreen({required this.onBack, super.key});

  final VoidCallback onBack;

  @override
  State<AssessmentScaleCategoryScreen> createState() =>
      _AssessmentScaleCategoryScreenState();
}

class _AssessmentScaleCategoryScreenState
    extends State<AssessmentScaleCategoryScreen> {
  String _searchQuery = '';
  String _pinyinBuffer = '';
  bool _searchKeyboardVisible = false;
  bool _searchKeyboardShifted = false;

  void _openSearchKeyboard() {
    setState(() => _searchKeyboardVisible = true);
  }

  void _finishSearchKeyboard() {
    final List<String> candidates = _pinyinCandidates(_pinyinBuffer);
    setState(() {
      if (_pinyinBuffer.isNotEmpty && candidates.isNotEmpty) {
        final int keepLength =
            math.max(0, _searchQuery.length - _pinyinBuffer.length);
        _searchQuery = _searchQuery.substring(0, keepLength) + candidates.first;
      }
      _searchKeyboardVisible = false;
      _searchKeyboardShifted = false;
      _pinyinBuffer = '';
    });
  }

  void _insertSearchText(String value) {
    if (_isAsciiLetter(value)) {
      setState(() {
        _searchQuery += value;
        _pinyinBuffer += value.toLowerCase();
      });
      return;
    }

    final List<String> candidates = _pinyinCandidates(_pinyinBuffer);
    if (_pinyinBuffer.isNotEmpty && value.length == 1) {
      final int? digit = int.tryParse(value);
      if (digit != null && digit >= 1 && digit <= candidates.length) {
        _commitPinyinCandidate(candidates[digit - 1]);
        return;
      }
      if (digit != null) {
        setState(() {
          _searchQuery += value;
          _pinyinBuffer += value;
        });
        return;
      }
    }

    if (value == ' ' && candidates.isNotEmpty) {
      _commitPinyinCandidate(candidates.first);
      return;
    }

    setState(() {
      _searchQuery += value;
      _pinyinBuffer = '';
    });
  }

  void _replaceSearchText(String value) {
    setState(() {
      _searchQuery = value;
      _pinyinBuffer = '';
    });
  }

  void _commitPinyinCandidate(String value) {
    if (_pinyinBuffer.isEmpty) {
      setState(() => _searchQuery += value);
      return;
    }
    final int keepLength =
        math.max(0, _searchQuery.length - _pinyinBuffer.length);
    setState(() {
      _searchQuery = _searchQuery.substring(0, keepLength) + value;
      _pinyinBuffer = '';
    });
  }

  void _deleteSearchText() {
    if (_searchQuery.isEmpty) {
      return;
    }
    final List<int> runes = _searchQuery.runes.toList();
    setState(() {
      _searchQuery = String.fromCharCodes(runes.take(runes.length - 1));
      if (_pinyinBuffer.isNotEmpty) {
        final List<int> pinyinRunes = _pinyinBuffer.runes.toList();
        _pinyinBuffer =
            String.fromCharCodes(pinyinRunes.take(pinyinRunes.length - 1));
      }
    });
  }

  void _clearSearchText() {
    setState(() {
      _searchQuery = '';
      _pinyinBuffer = '';
    });
  }

  void _toggleSearchShift() {
    setState(() => _searchKeyboardShifted = !_searchKeyboardShifted);
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        final double width =
            constraints.maxWidth.isFinite ? constraints.maxWidth : 1366;
        final bool compact = width < 1180;
        final double margin = compact ? 24 : 32;
        final double leftWidth = compact ? 214 : 232;
        final double contentGap = compact ? 12 : 22;

        return ColoredBox(
          color: _ScaleColors.page,
          child: Stack(
            children: <Widget>[
              const Positioned.fill(child: _ScalePageBackground()),
              Padding(
                padding: EdgeInsets.fromLTRB(margin, 26, margin, 22),
                child: Column(
                  children: <Widget>[
                    _ScaleTopBar(
                      onBack: widget.onBack,
                      compact: compact,
                      searchQuery: _searchQuery,
                      searchActive: _searchKeyboardVisible,
                      onSearchTap: _openSearchKeyboard,
                      onSearchClear: _clearSearchText,
                    ),
                    const SizedBox(height: 22),
                    Expanded(
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: <Widget>[
                          SizedBox(
                            width: leftWidth,
                            child: const _ScaleCategorySidebar(),
                          ),
                          SizedBox(width: contentGap),
                          Expanded(
                            child: _ScaleMainContent(searchQuery: _searchQuery),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              if (_searchKeyboardVisible)
                Positioned.fill(
                  child: GestureDetector(
                    behavior: HitTestBehavior.translucent,
                    onTap: _finishSearchKeyboard,
                  ),
                ),
              if (_searchKeyboardVisible)
                Positioned(
                  right: margin,
                  top: 88,
                  child: GestureDetector(
                    behavior: HitTestBehavior.opaque,
                    onTap: () {},
                    child: _ScaleSearchKeyboard(
                      value: _searchQuery,
                      pinyinBuffer: _pinyinBuffer,
                      shifted: _searchKeyboardShifted,
                      compact: compact,
                      onKey: _insertSearchText,
                      onReplace: _replaceSearchText,
                      onCandidate: _commitPinyinCandidate,
                      onBackspace: _deleteSearchText,
                      onClear: _clearSearchText,
                      onShift: _toggleSearchShift,
                      onClose: _finishSearchKeyboard,
                    ),
                  ),
                ),
            ],
          ),
        );
      },
    );
  }
}

class _ScaleColors {
  static const Color page = Color(0xFFFFF7EE);
  static const Color card = Color(0xFFFFFEFB);
  static const Color ink = Color(0xFF3F2B22);
  static const Color text = Color(0xFF6F5B50);
  static const Color muted = Color(0xFFA7958B);
  static const Color line = Color(0xFFEAD7C9);
  static const Color lineSoft = Color(0xFFF4E8DF);
  static const Color orange = Color(0xFFE96F43);
  static const Color orangeDeep = Color(0xFFC95D37);
}

List<BoxShadow> _scaleShadow({
  Color color = const Color(0x12B05F32),
  double blur = 24,
  Offset offset = const Offset(0, 12),
}) {
  return <BoxShadow>[BoxShadow(color: color, blurRadius: blur, offset: offset)];
}

class _ScalePageBackground extends StatelessWidget {
  const _ScalePageBackground();

  @override
  Widget build(BuildContext context) {
    return CustomPaint(painter: _ScalePageBackgroundPainter());
  }
}

class _ScalePageBackgroundPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final Paint top = Paint()..color = const Color(0xFFFFF1E3);
    canvas.drawPath(
      Path()
        ..moveTo(0, 0)
        ..lineTo(size.width, 0)
        ..lineTo(size.width, 126)
        ..quadraticBezierTo(size.width * .55, 82, 0, 132)
        ..close(),
      top,
    );

    canvas.drawOval(
      Rect.fromLTWH(size.width - 240, 72, 210, 90),
      Paint()..color = const Color(0x34FFE0C2),
    );
    canvas.drawOval(
      Rect.fromLTWH(16, size.height - 118, 220, 92),
      Paint()..color = const Color(0x22F4C492),
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _ScaleTopBar extends StatelessWidget {
  const _ScaleTopBar({
    required this.onBack,
    required this.compact,
    required this.searchQuery,
    required this.searchActive,
    required this.onSearchTap,
    required this.onSearchClear,
  });

  final VoidCallback onBack;
  final bool compact;
  final String searchQuery;
  final bool searchActive;
  final VoidCallback onSearchTap;
  final VoidCallback onSearchClear;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 62,
      child: Row(
        children: <Widget>[
          _IconShell(
            icon: Icons.chevron_left_rounded,
            onTap: onBack,
            size: 46,
            iconSize: 34,
          ),
          const SizedBox(width: 16),
          const Text(
            '开始测评',
            style: TextStyle(
              color: _ScaleColors.ink,
              fontSize: 30,
              height: 1,
              fontWeight: FontWeight.w900,
            ),
          ),
          const Spacer(),
          _SearchBox(
            width: compact ? 280 : 328,
            value: searchQuery,
            active: searchActive,
            onTap: onSearchTap,
            onClear: onSearchClear,
          ),
          const SizedBox(width: 14),
          const _StudentChip(),
          const SizedBox(width: 14),
          const _AvailableFilterChip(),
        ],
      ),
    );
  }
}

class _IconShell extends StatelessWidget {
  const _IconShell({
    required this.icon,
    required this.onTap,
    this.size = 44,
    this.iconSize = 25,
  });

  final IconData icon;
  final VoidCallback onTap;
  final double size;
  final double iconSize;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(14),
        child: Ink(
          width: size,
          height: size,
          decoration: BoxDecoration(
            color: _ScaleColors.card,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: _ScaleColors.line),
          ),
          child: Icon(icon, size: iconSize, color: _ScaleColors.text),
        ),
      ),
    );
  }
}

class _SearchBox extends StatelessWidget {
  const _SearchBox({
    required this.width,
    required this.value,
    required this.active,
    required this.onTap,
    required this.onClear,
  });

  final double width;
  final String value;
  final bool active;
  final VoidCallback onTap;
  final VoidCallback onClear;

  @override
  Widget build(BuildContext context) {
    final bool hasValue = value.trim().isNotEmpty;
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(15),
        child: Ink(
          width: width,
          height: 46,
          padding: const EdgeInsets.symmetric(horizontal: 17),
          decoration: BoxDecoration(
            color: Colors.white.withOpacity(.88),
            borderRadius: BorderRadius.circular(15),
            border: Border.all(
              color: active ? _ScaleColors.orange : _ScaleColors.line,
              width: active ? 1.4 : 1,
            ),
            boxShadow: active
                ? _scaleShadow(
                    color: const Color(0x18E96F43),
                    blur: 16,
                    offset: const Offset(0, 7),
                  )
                : null,
          ),
          child: Row(
            children: <Widget>[
              Icon(
                Icons.search_rounded,
                size: 22,
                color: active ? _ScaleColors.orange : _ScaleColors.text,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  hasValue ? value : '搜索量表名称 / 编码',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: hasValue ? _ScaleColors.ink : _ScaleColors.muted,
                    fontSize: 15,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ),
              if (hasValue) ...<Widget>[
                const SizedBox(width: 8),
                GestureDetector(
                  behavior: HitTestBehavior.opaque,
                  onTap: onClear,
                  child: const Icon(
                    Icons.cancel_rounded,
                    size: 19,
                    color: _ScaleColors.muted,
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _ScaleSearchKeyboard extends StatelessWidget {
  const _ScaleSearchKeyboard({
    required this.value,
    required this.pinyinBuffer,
    required this.shifted,
    required this.compact,
    required this.onKey,
    required this.onReplace,
    required this.onCandidate,
    required this.onBackspace,
    required this.onClear,
    required this.onShift,
    required this.onClose,
  });

  final String value;
  final String pinyinBuffer;
  final bool shifted;
  final bool compact;
  final ValueChanged<String> onKey;
  final ValueChanged<String> onReplace;
  final ValueChanged<String> onCandidate;
  final VoidCallback onBackspace;
  final VoidCallback onClear;
  final VoidCallback onShift;
  final VoidCallback onClose;

  @override
  Widget build(BuildContext context) {
    final List<String> candidates = _pinyinCandidates(pinyinBuffer);
    return Container(
      width: compact ? 600 : 660,
      padding: const EdgeInsets.fromLTRB(16, 14, 16, 16),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.97),
        borderRadius: BorderRadius.circular(22),
        border: Border.all(color: _ScaleColors.line),
        boxShadow: _scaleShadow(
          color: const Color(0x28B05F32),
          blur: 28,
          offset: const Offset(0, 15),
        ),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Row(
            children: <Widget>[
              const Icon(
                Icons.keyboard_alt_outlined,
                color: _ScaleColors.orange,
                size: 21,
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Align(
                  alignment: Alignment.centerLeft,
                  child: Container(
                    height: 25,
                    constraints: const BoxConstraints(maxWidth: 430),
                    padding: const EdgeInsets.symmetric(horizontal: 10),
                    alignment: Alignment.centerLeft,
                    decoration: BoxDecoration(
                      color: const Color(0xFFFFF1E8),
                      borderRadius: BorderRadius.circular(999),
                    ),
                    child: Text(
                      value.isEmpty ? '搜索量表' : value,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      textAlign: TextAlign.left,
                      style: const TextStyle(
                        color: _ScaleColors.orangeDeep,
                        fontSize: 13,
                        height: 1,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                ),
              ),
              _SearchKeyboardAction(
                label: '关闭',
                icon: Icons.close_rounded,
                onTap: onClose,
              ),
            ],
          ),
          const SizedBox(height: 12),
          _PinyinCandidateBar(
            pinyin: pinyinBuffer,
            candidates: candidates,
            onCandidate: onCandidate,
          ),
          const SizedBox(height: 12),
          Wrap(
            spacing: 9,
            runSpacing: 9,
            children: <Widget>[
              for (final String word in _quickWords)
                _SearchKeyboardQuickKey(
                  label: word,
                  onTap: () => onReplace(word),
                ),
            ],
          ),
          const SizedBox(height: 12),
          _keyRow(_digitKeys),
          const SizedBox(height: 9),
          _keyRow(_letterKeys('qwertyuiop')),
          const SizedBox(height: 9),
          _keyRow(_letterKeys('asdfghjkl')),
          const SizedBox(height: 9),
          _keyRow(<Widget>[
            _SearchKeyboardKey(
              label: '大写',
              active: shifted,
              flex: 2,
              onTap: onShift,
            ),
            ..._letterKeys('zxcvbnm'),
            _SearchKeyboardKey(
              label: '删除',
              flex: 2,
              onTap: onBackspace,
            ),
          ]),
          const SizedBox(height: 9),
          _keyRow(<Widget>[
            _SearchKeyboardKey(label: '-', onTap: () => onKey('-')),
            _SearchKeyboardKey(label: '/', onTap: () => onKey('/')),
            _SearchKeyboardKey(label: '空格', flex: 2, onTap: () => onKey(' ')),
            _SearchKeyboardKey(
                label: '清空', flex: 2, muted: true, onTap: onClear),
            _SearchKeyboardKey(
              label: '完成',
              flex: 2,
              primary: true,
              onTap: onClose,
            ),
          ]),
        ],
      ),
    );
  }

  List<Widget> get _digitKeys {
    return '1234567890'
        .split('')
        .map((String value) => _SearchKeyboardKey(
              label: value,
              onTap: () => onKey(value),
            ))
        .toList();
  }

  List<Widget> _letterKeys(String values) {
    return values.split('').map((String value) {
      final String label = shifted ? value.toUpperCase() : value;
      return _SearchKeyboardKey(label: label, onTap: () => onKey(label));
    }).toList();
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

  static const List<String> _quickWords = <String>[
    'PEP-3',
    '语言',
    '沟通',
    '筛查',
    '口语',
    '表达',
    '社交',
    '综合',
  ];
}

class _PinyinCandidateBar extends StatelessWidget {
  const _PinyinCandidateBar({
    required this.pinyin,
    required this.candidates,
    required this.onCandidate,
  });

  final String pinyin;
  final List<String> candidates;
  final ValueChanged<String> onCandidate;

  @override
  Widget build(BuildContext context) {
    final bool composing = pinyin.trim().isNotEmpty;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(10, 9, 10, 9),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: _ScaleColors.lineSoft),
      ),
      child: Row(
        children: <Widget>[
          Container(
            height: 30,
            constraints: const BoxConstraints(minWidth: 58),
            padding: const EdgeInsets.symmetric(horizontal: 10),
            alignment: Alignment.centerLeft,
            decoration: BoxDecoration(
              color: const Color(0xFFFFF1E8),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Text(
              composing ? pinyin : '拼音',
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: _ScaleColors.orangeDeep,
                fontSize: 14,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: !composing
                ? const Text(
                    '输入拼音后在这里选择汉字候选',
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                    style: TextStyle(
                      color: _ScaleColors.muted,
                      fontSize: 13,
                      height: 1,
                      fontWeight: FontWeight.w700,
                    ),
                  )
                : candidates.isEmpty
                    ? const Text(
                        '暂无候选，继续输入或直接搜索编码',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: TextStyle(
                          color: _ScaleColors.muted,
                          fontSize: 13,
                          height: 1,
                          fontWeight: FontWeight.w700,
                        ),
                      )
                    : Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: <Widget>[
                          for (int index = 0;
                              index < candidates.length;
                              index++)
                            _SearchKeyboardQuickKey(
                              label: '${index + 1}.${candidates[index]}',
                              onTap: () => onCandidate(candidates[index]),
                            ),
                        ],
                      ),
          ),
        ],
      ),
    );
  }
}

class _SearchKeyboardQuickKey extends StatefulWidget {
  const _SearchKeyboardQuickKey({required this.label, required this.onTap});

  final String label;
  final VoidCallback onTap;

  @override
  State<_SearchKeyboardQuickKey> createState() =>
      _SearchKeyboardQuickKeyState();
}

class _SearchKeyboardQuickKeyState extends State<_SearchKeyboardQuickKey> {
  bool _pressed = false;
  bool _showBubble = false;
  int _feedbackToken = 0;

  void _handleTapDown(TapDownDetails _) {
    HapticFeedback.selectionClick();
    _feedbackToken++;
    setState(() {
      _pressed = true;
      _showBubble = true;
    });
  }

  void _hidePressBubbleSoon() {
    final int token = _feedbackToken;
    Future<void>.delayed(const Duration(milliseconds: 90), () {
      if (mounted && token == _feedbackToken) {
        setState(() => _pressed = false);
      }
    });
    Future<void>.delayed(const Duration(milliseconds: 210), () {
      if (mounted && token == _feedbackToken) {
        setState(() => _showBubble = false);
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      key: ValueKey<String>('scale-search-key-${widget.label}'),
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
              height: 32,
              padding: const EdgeInsets.symmetric(horizontal: 14),
              decoration: BoxDecoration(
                color: _pressed
                    ? const Color(0xFFFFE8DA)
                    : const Color(0xFFFFF8F1),
                borderRadius: BorderRadius.circular(999),
                border: Border.all(
                  color: _pressed ? _ScaleColors.orange : _ScaleColors.lineSoft,
                  width: _pressed ? 1.3 : 1,
                ),
              ),
              child: Center(
                widthFactor: 1,
                child: Text(
                  widget.label,
                  style: const TextStyle(
                    color: _ScaleColors.text,
                    fontSize: 13,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
          ),
          if (_showBubble)
            Positioned(
              bottom: 40,
              child: _SearchKeyboardBubble(label: widget.label),
            ),
        ],
      ),
    );
  }
}

class _SearchKeyboardBubble extends StatelessWidget {
  const _SearchKeyboardBubble({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return IgnorePointer(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Container(
            height: 46,
            constraints: const BoxConstraints(minWidth: 46),
            padding: const EdgeInsets.symmetric(horizontal: 15),
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: _ScaleColors.orange,
              borderRadius: BorderRadius.circular(14),
              boxShadow: _scaleShadow(
                color: const Color(0x24D15E36),
                blur: 14,
                offset: const Offset(0, 8),
              ),
            ),
            child: Text(
              label,
              maxLines: 1,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 20,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          ClipPath(
            clipper: _BubbleTriangleClipper(),
            child: Container(
              width: 12,
              height: 7,
              color: _ScaleColors.orange,
            ),
          ),
        ],
      ),
    );
  }
}

class _BubbleTriangleClipper extends CustomClipper<Path> {
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

class _SearchKeyboardKey extends StatefulWidget {
  const _SearchKeyboardKey({
    required this.label,
    required this.onTap,
    this.flex = 1,
    this.primary = false,
    this.active = false,
    this.muted = false,
  });

  final String label;
  final VoidCallback onTap;
  final int flex;
  final bool primary;
  final bool active;
  final bool muted;

  @override
  State<_SearchKeyboardKey> createState() => _SearchKeyboardKeyState();
}

class _SearchKeyboardKeyState extends State<_SearchKeyboardKey> {
  bool _pressed = false;
  bool _showBubble = false;
  int _feedbackToken = 0;

  void _handleTapDown(TapDownDetails _) {
    HapticFeedback.selectionClick();
    _feedbackToken++;
    setState(() {
      _pressed = true;
      _showBubble = true;
    });
  }

  void _hidePressBubbleSoon() {
    final int token = _feedbackToken;
    Future<void>.delayed(const Duration(milliseconds: 90), () {
      if (mounted && token == _feedbackToken) {
        setState(() => _pressed = false);
      }
    });
    Future<void>.delayed(const Duration(milliseconds: 210), () {
      if (mounted && token == _feedbackToken) {
        setState(() => _showBubble = false);
      }
    });
  }

  @override
  Widget build(BuildContext context) {
    final Color background = widget.primary
        ? _ScaleColors.orange
        : widget.active
            ? const Color(0xFFFFE7D9)
            : widget.muted
                ? const Color(0xFFF7EFE9)
                : const Color(0xFFFFFAF5);
    final Color foreground = widget.primary
        ? Colors.white
        : widget.active
            ? _ScaleColors.orangeDeep
            : _ScaleColors.ink;
    final Color pressedBackground = widget.primary
        ? _ScaleColors.orangeDeep
        : widget.active
            ? const Color(0xFFFFD6C2)
            : const Color(0xFFFFE8DA);
    final Color keyBackground = _pressed ? pressedBackground : background;
    final Color keyBorder =
        _pressed ? _ScaleColors.orange : _ScaleColors.lineSoft;

    return Expanded(
      flex: widget.flex,
      child: GestureDetector(
        key: ValueKey<String>('scale-search-key-${widget.label}'),
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
                height: 48,
                decoration: BoxDecoration(
                  color: keyBackground,
                  borderRadius: BorderRadius.circular(12),
                  border: widget.primary
                      ? null
                      : Border.all(color: keyBorder, width: _pressed ? 1.4 : 1),
                  boxShadow: widget.primary || _pressed
                      ? _scaleShadow(
                          color: _pressed
                              ? const Color(0x2FD15E36)
                              : const Color(0x20D15E36),
                          blur: _pressed ? 16 : 12,
                          offset: Offset(0, _pressed ? 5 : 7),
                        )
                      : null,
                ),
                child: Center(
                  child: Text(
                    widget.label,
                    maxLines: 1,
                    overflow: TextOverflow.clip,
                    strutStyle: const StrutStyle(
                      fontSize: 15,
                      height: 1,
                      forceStrutHeight: true,
                    ),
                    style: TextStyle(
                      color: foreground,
                      fontSize: widget.label.length > 2 ? 13 : 16,
                      height: 1,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
              ),
            ),
            if (_showBubble)
              Positioned(
                bottom: 56,
                child: _SearchKeyboardBubble(label: widget.label),
              ),
          ],
        ),
      ),
    );
  }
}

class _SearchKeyboardAction extends StatelessWidget {
  const _SearchKeyboardAction({
    required this.label,
    required this.icon,
    required this.onTap,
  });

  final String label;
  final IconData icon;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 5),
          child: Row(
            children: <Widget>[
              Icon(icon, color: _ScaleColors.text, size: 18),
              const SizedBox(width: 3),
              Text(
                label,
                style: const TextStyle(
                  color: _ScaleColors.text,
                  fontSize: 12,
                  height: 1,
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

class _StudentChip extends StatelessWidget {
  const _StudentChip();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 46,
      padding: const EdgeInsets.symmetric(horizontal: 18),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.86),
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: _ScaleColors.line),
      ),
      child: const Row(
        children: <Widget>[
          Icon(Icons.person_outline_rounded,
              size: 23, color: _ScaleColors.text),
          SizedBox(width: 10),
          Text(
            '未选择学员',
            style: TextStyle(
              color: _ScaleColors.ink,
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _AvailableFilterChip extends StatelessWidget {
  const _AvailableFilterChip();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 46,
      padding: const EdgeInsets.symmetric(horizontal: 17),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.86),
        borderRadius: BorderRadius.circular(15),
        border: Border.all(color: _ScaleColors.line),
      ),
      child: const Row(
        children: <Widget>[
          Icon(Icons.filter_alt_outlined, size: 22, color: _ScaleColors.text),
          SizedBox(width: 9),
          Text(
            '停用量表',
            style: TextStyle(
              color: _ScaleColors.ink,
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
          SizedBox(width: 9),
          _FilterDot(),
        ],
      ),
    );
  }
}

class _FilterDot extends StatelessWidget {
  const _FilterDot();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 8,
      height: 8,
      decoration: const BoxDecoration(
        color: _ScaleColors.orange,
        shape: BoxShape.circle,
      ),
    );
  }
}

class _ScaleCategorySidebar extends StatelessWidget {
  const _ScaleCategorySidebar();

  static const List<_CategoryItemData> _items = <_CategoryItemData>[
    _CategoryItemData('发展筛查', 12, Color(0xFFE96F43)),
    _CategoryItemData('语言与沟通能力', 16, Color(0xFF3F82D2), active: true),
    _CategoryItemData('社交情绪评估', 14, Color(0xFF6F9F70)),
    _CategoryItemData('感觉统合', 11, Color(0xFFD99427)),
    _CategoryItemData('动作与精细运动', 10, Color(0xFF63A999)),
    _CategoryItemData('生活自理能力', 9, Color(0xFFD96A7F)),
    _CategoryItemData('适应行为', 8, Color(0xFF6F9F70)),
    _CategoryItemData('IEP 目标库', 6, Color(0xFF7F77C8)),
  ];

  @override
  Widget build(BuildContext context) {
    return Column(
      children: <Widget>[
        Expanded(
          child: Container(
            padding: const EdgeInsets.fromLTRB(14, 17, 14, 16),
            decoration: BoxDecoration(
              color: Colors.white.withOpacity(.88),
              borderRadius: BorderRadius.circular(20),
              border: Border.all(color: _ScaleColors.line),
              boxShadow: _scaleShadow(),
            ),
            child: Column(
              children: <Widget>[
                const Padding(
                  padding: EdgeInsets.symmetric(horizontal: 2),
                  child: Row(
                    children: <Widget>[
                      Text(
                        '分类',
                        style: TextStyle(
                          color: _ScaleColors.ink,
                          fontSize: 19,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                      Spacer(),
                      Text(
                        '128',
                        style: TextStyle(
                          color: _ScaleColors.muted,
                          fontSize: 12,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 14),
                for (int index = 0; index < _items.length; index++) ...<Widget>[
                  _CategoryItem(data: _items[index]),
                  if (index != _items.length - 1) const SizedBox(height: 5),
                ],
              ],
            ),
          ),
        ),
        const SizedBox(height: 14),
        const _DraftCard(),
      ],
    );
  }
}

class _CategoryItemData {
  const _CategoryItemData(
    this.name,
    this.count,
    this.color, {
    this.active = false,
  });

  final String name;
  final int count;
  final Color color;
  final bool active;
}

class _CategoryItem extends StatelessWidget {
  const _CategoryItem({required this.data});

  final _CategoryItemData data;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 46,
      padding: const EdgeInsets.symmetric(horizontal: 10),
      decoration: BoxDecoration(
        color: data.active ? const Color(0xFFFFF0E7) : Colors.transparent,
        borderRadius: BorderRadius.circular(14),
      ),
      child: Row(
        children: <Widget>[
          Container(
            width: 10,
            height: 10,
            decoration:
                BoxDecoration(color: data.color, shape: BoxShape.circle),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              data.name,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color:
                    data.active ? _ScaleColors.orangeDeep : _ScaleColors.text,
                fontSize: 15,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          const SizedBox(width: 12),
          Text(
            '${data.count}',
            style: TextStyle(
              color: data.active ? _ScaleColors.orangeDeep : _ScaleColors.muted,
              fontSize: 12,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _DraftCard extends StatelessWidget {
  const _DraftCard();

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 118,
      padding: const EdgeInsets.fromLTRB(15, 14, 15, 13),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(.74),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: _ScaleColors.line),
      ),
      child: const Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          Row(
            children: <Widget>[
              Icon(Icons.assignment_outlined,
                  size: 18, color: _ScaleColors.ink),
              SizedBox(width: 8),
              Text(
                '继续草稿',
                style: TextStyle(
                  color: _ScaleColors.ink,
                  fontSize: 16,
                  fontWeight: FontWeight.w900,
                ),
              ),
            ],
          ),
          SizedBox(height: 9),
          Text(
            '未完成的测评可直接恢复，支持断点续测！',
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              color: _ScaleColors.text,
              fontSize: 12,
              height: 1.45,
              fontWeight: FontWeight.w700,
            ),
          ),
          Spacer(),
          Text(
            '当前共3条草稿',
            style: TextStyle(
              color: _ScaleColors.orange,
              fontSize: 14,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
    );
  }
}

class _ScaleMainContent extends StatelessWidget {
  const _ScaleMainContent({required this.searchQuery});

  final String searchQuery;

  static const List<_ScaleCardData> _scales = <_ScaleCardData>[
    _ScaleCardData(
      title: 'PEP-3语言理解评核量表',
      code: 'PEP3-CVP',
      tags: <String>['56题', '25分钟', '2-7岁'],
      aliases: <String>['pep3', 'yuyan', 'lijie', 'pinghe', 'liangbiao'],
      type: _CoverType.book,
    ),
    _ScaleCardData(
      title: '口语发起与互动',
      code: 'ORAL-INT',
      tags: <String>['42题', '18分钟', '3-8岁'],
      aliases: <String>['kouyu', 'kou', 'hudong', 'faqi', 'oral'],
      type: _CoverType.talk,
    ),
    _ScaleCardData(
      title: '语言发展筛查表',
      code: 'LANG-SCREEN',
      tags: <String>['32题', '15分钟', '12-48月'],
      aliases: <String>['yuyan', 'fazhan', 'shaicha', 'liangbiao', 'lang'],
      type: _CoverType.screen,
    ),
    _ScaleCardData(
      title: '表达沟通量表',
      code: 'EXP-COMM',
      tags: <String>['38题', '22分钟', '学龄前'],
      aliases: <String>['biaoda', 'goutong', 'liangbiao', 'exp', 'comm'],
      type: _CoverType.express,
    ),
    _ScaleCardData(
      title: '社交沟通观察表',
      code: 'SOC-COMM',
      tags: <String>['44题', '20分钟', '3-10岁'],
      aliases: <String>['shejiao', 'goutong', 'guancha', 'liangbiao', 'soc'],
      type: _CoverType.social,
    ),
    _ScaleCardData(
      title: '综合语言复评卡',
      code: 'LANG-REVIEW',
      tags: <String>['48题', '30分钟', '4-8岁'],
      aliases: <String>['zonghe', 'yuyan', 'fuping', 'review', 'lang'],
      type: _CoverType.review,
    ),
  ];

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (BuildContext context, BoxConstraints constraints) {
        const double gap = 14;
        const double toolbarHeight = 44;
        const double toolbarGap = 10;
        final List<_ScaleCardData> visibleScales = _filterScales(searchQuery);
        final double cardWidth = (constraints.maxWidth - gap * 2) / 3;
        final double cardHeight =
            (constraints.maxHeight - toolbarHeight - gap - toolbarGap) / 2;

        return Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            const _ScaleToolbar(),
            const SizedBox(height: toolbarGap),
            if (visibleScales.isEmpty)
              const Expanded(child: _ScaleSearchEmpty())
            else
              Wrap(
                spacing: gap,
                runSpacing: gap,
                children: <Widget>[
                  for (final _ScaleCardData data in visibleScales)
                    SizedBox(
                      width: cardWidth,
                      height: cardHeight.clamp(252, 292),
                      child: _ScaleCard(data: data),
                    ),
                ],
              ),
          ],
        );
      },
    );
  }

  static List<_ScaleCardData> _filterScales(String query) {
    final String normalizedQuery = _normalizeSearchText(query);
    if (normalizedQuery.isEmpty) {
      return _scales;
    }
    return _scales.where((_ScaleCardData item) {
      final String target = _normalizeSearchText(
        <String>[item.title, item.code, ...item.tags, ...item.aliases]
            .join(' '),
      );
      return target.contains(normalizedQuery);
    }).toList();
  }
}

class _ScaleToolbar extends StatelessWidget {
  const _ScaleToolbar();

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 44,
      child: Transform.translate(
        offset: const Offset(0, 0),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: <Widget>[
            const Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: <Widget>[
                Text(
                  '语言沟通量表',
                  style: TextStyle(
                    color: _ScaleColors.ink,
                    fontSize: 27,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
                SizedBox(width: 17),
                Text(
                  '16 个可用，5 个常用',
                  style: TextStyle(
                    color: _ScaleColors.muted,
                    fontSize: 14,
                    height: 1,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
            const Spacer(),
            Container(
              height: 42,
              padding: const EdgeInsets.all(4),
              decoration: BoxDecoration(
                color: const Color(0xFFFFF1E8).withOpacity(.78),
                borderRadius: BorderRadius.circular(15),
              ),
              child: const Row(
                children: <Widget>[
                  _Segment(label: '常用', active: true),
                  _Segment(label: '全部'),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _Segment extends StatelessWidget {
  const _Segment({required this.label, this.active = false});

  final String label;
  final bool active;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 90,
      height: 34,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: active ? Colors.white : Colors.transparent,
        borderRadius: BorderRadius.circular(12),
        boxShadow: active
            ? _scaleShadow(
                color: const Color(0x12B05F32),
                blur: 12,
                offset: const Offset(0, 5),
              )
            : null,
      ),
      child: Text(
        label,
        strutStyle: const StrutStyle(
          fontSize: 14,
          height: 1,
          forceStrutHeight: true,
        ),
        style: TextStyle(
          color: active ? _ScaleColors.orangeDeep : _ScaleColors.text,
          fontSize: 14,
          height: 1,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _ScaleCardData {
  const _ScaleCardData({
    required this.title,
    required this.code,
    required this.tags,
    required this.aliases,
    required this.type,
  });

  final String title;
  final String code;
  final List<String> tags;
  final List<String> aliases;
  final _CoverType type;
}

String _normalizeSearchText(String value) {
  return value.trim().toLowerCase().replaceAll(RegExp(r'[\s\-/_.]'), '');
}

bool _isAsciiLetter(String value) {
  return value.length == 1 && RegExp(r'^[A-Za-z]$').hasMatch(value);
}

List<String> _pinyinCandidates(String value) {
  final String pinyin =
      value.toLowerCase().replaceAll(RegExp(r'[^a-z0-9]'), '');
  if (pinyin.isEmpty) {
    return const <String>[];
  }
  final List<String> exact = _pinyinCandidateMap[pinyin] ?? const <String>[];
  final List<String> prefix = <String>[];
  for (final MapEntry<String, List<String>> entry
      in _pinyinCandidateMap.entries) {
    if (entry.key.startsWith(pinyin)) {
      prefix.addAll(entry.value);
    }
  }
  return <String>[...exact, ...prefix]
      .where((String item) => item.trim().isNotEmpty)
      .toSet()
      .take(8)
      .toList();
}

const Map<String, List<String>> _pinyinCandidateMap = <String, List<String>>{
  'p': <String>['评', 'PEP-3'],
  'pe': <String>['PEP-3'],
  'pep': <String>['PEP-3'],
  'pep3': <String>['PEP-3'],
  'y': <String>['语', '言'],
  'yu': <String>['语', '语言'],
  'yan': <String>['言'],
  'yuyan': <String>['语言'],
  'g': <String>['沟'],
  'go': <String>['沟'],
  'gou': <String>['沟'],
  'tong': <String>['通'],
  'gt': <String>['沟通'],
  'goutong': <String>['沟通'],
  's': <String>['筛', '社'],
  'shai': <String>['筛'],
  'cha': <String>['查'],
  'sc': <String>['筛查'],
  'shaicha': <String>['筛查'],
  'kou': <String>['口'],
  'ky': <String>['口语'],
  'kouyu': <String>['口语'],
  'biao': <String>['表'],
  'da': <String>['达'],
  'bd': <String>['表达'],
  'biaoda': <String>['表达'],
  'she': <String>['社'],
  'jiao': <String>['交'],
  'sj': <String>['社交'],
  'shejiao': <String>['社交'],
  'zong': <String>['综'],
  'he': <String>['合'],
  'zh': <String>['综合'],
  'zonghe': <String>['综合'],
  'fa': <String>['发'],
  'zhan': <String>['展'],
  'fz': <String>['发展'],
  'fazhan': <String>['发展'],
  'ping': <String>['评'],
  'heping': <String>['评核'],
  'ph': <String>['评核'],
  'pinghe': <String>['评核'],
  'liang': <String>['量'],
  'lb': <String>['量表'],
  'liangbiao': <String>['量表'],
  'guan': <String>['观'],
  'gc': <String>['观察'],
  'guancha': <String>['观察'],
  'fu': <String>['复'],
  'fp': <String>['复评'],
  'fuping': <String>['复评'],
  'er': <String>['儿'],
  'et': <String>['儿童'],
  'ertong': <String>['儿童'],
  'li': <String>['理'],
  'jie': <String>['解'],
  'lj': <String>['理解'],
  'lijie': <String>['理解'],
};

class _ScaleSearchEmpty extends StatelessWidget {
  const _ScaleSearchEmpty();

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        width: 320,
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 22),
        decoration: BoxDecoration(
          color: Colors.white.withOpacity(.72),
          borderRadius: BorderRadius.circular(18),
          border: Border.all(color: _ScaleColors.lineSoft),
        ),
        child: const Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            Icon(Icons.search_off_rounded, color: _ScaleColors.muted, size: 34),
            SizedBox(height: 10),
            Text(
              '没有匹配的量表',
              style: TextStyle(
                color: _ScaleColors.ink,
                fontSize: 17,
                height: 1,
                fontWeight: FontWeight.w900,
              ),
            ),
            SizedBox(height: 8),
            Text(
              '可尝试输入 PEP3、语言、筛查等关键词',
              textAlign: TextAlign.center,
              style: TextStyle(
                color: _ScaleColors.muted,
                fontSize: 13,
                height: 1.35,
                fontWeight: FontWeight.w700,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _ScaleCard extends StatelessWidget {
  const _ScaleCard({required this.data});

  final _ScaleCardData data;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: _ScaleColors.card.withOpacity(.96),
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: _ScaleColors.line, width: 1.1),
        boxShadow: _scaleShadow(),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: <Widget>[
          Expanded(
            child: ClipRRect(
              borderRadius: BorderRadius.circular(16),
              child: CustomPaint(
                painter: _ScaleCoverPainter(type: data.type),
                child: const SizedBox.expand(),
              ),
            ),
          ),
          const SizedBox(height: 11),
          Row(
            children: <Widget>[
              Expanded(
                child: Text(
                  data.title,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: const TextStyle(
                    color: _ScaleColors.ink,
                    fontSize: 20,
                    height: 1,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: <Widget>[
              for (int i = 0; i < data.tags.length; i++) ...<Widget>[
                _InfoTag(label: data.tags[i]),
                if (i != data.tags.length - 1) const SizedBox(width: 6),
              ],
              const Spacer(),
              const _ChooseButton(),
            ],
          ),
        ],
      ),
    );
  }
}

class _InfoTag extends StatelessWidget {
  const _InfoTag({required this.label});

  final String label;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 8),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: const Color(0xFFFFF8F1),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _ScaleColors.lineSoft),
      ),
      child: Text(
        label,
        maxLines: 1,
        style: const TextStyle(
          color: _ScaleColors.text,
          fontSize: 13,
          fontWeight: FontWeight.w800,
        ),
      ),
    );
  }
}

class _ChooseButton extends StatelessWidget {
  const _ChooseButton();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 64,
      height: 38,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _ScaleColors.orange, width: 1.2),
      ),
      child: const Text(
        '选择',
        maxLines: 1,
        textAlign: TextAlign.center,
        strutStyle: StrutStyle(
          fontSize: 15,
          height: 1,
          forceStrutHeight: true,
        ),
        style: TextStyle(
          color: _ScaleColors.orangeDeep,
          fontSize: 15,
          height: 1,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

enum _CoverType { book, talk, screen, express, social, review }

class _ScaleCoverPainter extends CustomPainter {
  const _ScaleCoverPainter({required this.type});

  final _CoverType type;

  @override
  void paint(Canvas canvas, Size size) {
    final Rect rect = Offset.zero & size;
    final Paint bg = Paint()
      ..shader = LinearGradient(
        colors: _backgroundColors(type),
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
      ).createShader(rect);
    canvas.drawRect(rect, bg);

    final Paint soft = Paint()..color = Colors.white.withOpacity(.34);
    canvas.drawCircle(Offset(size.width * .9, size.height * .05), 70, soft);
    canvas.drawCircle(Offset(size.width * .08, size.height * .95), 58, soft);

    switch (type) {
      case _CoverType.book:
        _drawBook(canvas, size);
      case _CoverType.talk:
        _drawTalk(canvas, size);
      case _CoverType.screen:
        _drawScreen(canvas, size);
      case _CoverType.express:
        _drawExpress(canvas, size);
      case _CoverType.social:
        _drawSocial(canvas, size);
      case _CoverType.review:
        _drawReview(canvas, size);
    }
  }

  List<Color> _backgroundColors(_CoverType type) {
    switch (type) {
      case _CoverType.book:
        return const <Color>[Color(0xFFFFF3E4), Color(0xFFFFD8BC)];
      case _CoverType.talk:
        return const <Color>[Color(0xFFF5F7EA), Color(0xFFDDEBD2)];
      case _CoverType.screen:
        return const <Color>[Color(0xFFF1F7FF), Color(0xFFD5E8F9)];
      case _CoverType.express:
        return const <Color>[Color(0xFFFFF6E1), Color(0xFFFFDFA7)];
      case _CoverType.social:
        return const <Color>[Color(0xFFF6F3EA), Color(0xFFDDECCF)];
      case _CoverType.review:
        return const <Color>[Color(0xFFFFF2EA), Color(0xFFF8D5C9)];
    }
  }

  void _drawBook(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    _drawChild(canvas, Offset(w * .28, h * .42),
        shirt: const Color(0xFF7FA1B5));
    final Paint page = Paint()..color = Colors.white.withOpacity(.86);
    final RRect left = RRect.fromRectAndRadius(
      Rect.fromLTWH(w * .34, h * .42, w * .2, h * .26),
      const Radius.circular(10),
    );
    final RRect right = RRect.fromRectAndRadius(
      Rect.fromLTWH(w * .53, h * .42, w * .2, h * .26),
      const Radius.circular(10),
    );
    canvas.drawRRect(left, page);
    canvas.drawRRect(right, page);
    final Paint line = Paint()
      ..color = const Color(0xFFE6A16B)
      ..strokeWidth = 4
      ..strokeCap = StrokeCap.round;
    canvas.drawLine(Offset(w * .535, h * .45), Offset(w * .535, h * .66), line);
    canvas.drawLine(Offset(w * .39, h * .5), Offset(w * .49, h * .5), line);
    canvas.drawLine(Offset(w * .58, h * .5), Offset(w * .68, h * .5), line);
    _drawSpeechBubble(canvas, Offset(w * .72, h * .26), const Color(0xFFFFFFFF),
        const Color(0xFFE96F43));
  }

  void _drawTalk(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    _drawChild(canvas, Offset(w * .34, h * .49),
        shirt: const Color(0xFF8FB279));
    _drawChild(canvas, Offset(w * .65, h * .5), shirt: const Color(0xFFE5A552));
    _drawSpeechBubble(
        canvas, Offset(w * .34, h * .2), Colors.white, const Color(0xFF6F9F70));
    _drawSpeechBubble(canvas, Offset(w * .68, h * .26), Colors.white,
        const Color(0xFF6F9F70));
  }

  void _drawScreen(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    final Paint board = Paint()..color = Colors.white.withOpacity(.9);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(w * .27, h * .16, w * .34, h * .64),
        const Radius.circular(16),
      ),
      board,
    );
    final Paint clip = Paint()..color = const Color(0xFFD9B27B);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(w * .36, h * .1, w * .16, 22),
        const Radius.circular(9),
      ),
      clip,
    );
    final Paint check = Paint()
      ..color = const Color(0xFF6F9F70)
      ..strokeWidth = 5
      ..strokeCap = StrokeCap.round
      ..style = PaintingStyle.stroke;
    for (final double y in <double>[.32, .48, .64]) {
      canvas.drawPath(
        Path()
          ..moveTo(w * .34, h * y)
          ..lineTo(w * .38, h * (y + .04))
          ..lineTo(w * .47, h * (y - .06)),
        check,
      );
      canvas.drawLine(Offset(w * .5, h * y), Offset(w * .57, h * y), check);
    }
    final Paint lens = Paint()
      ..color = const Color(0x803F82D2)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 10;
    canvas.drawCircle(Offset(w * .69, h * .48), 36, lens);
    canvas.drawLine(
      Offset(w * .72, h * .58),
      Offset(w * .82, h * .72),
      Paint()
        ..color = const Color(0xFF3F82D2)
        ..strokeWidth = 10
        ..strokeCap = StrokeCap.round,
    );
  }

  void _drawExpress(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    final Paint paper = Paint()..color = Colors.white.withOpacity(.86);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(w * .35, h * .16, w * .32, h * .66),
        const Radius.circular(15),
      ),
      paper,
    );
    final Paint pencil = Paint()..color = const Color(0xFFE6A13D);
    canvas.save();
    canvas.translate(w * .58, h * .51);
    canvas.rotate(-.72);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(-12, -54, 24, 108),
        const Radius.circular(9),
      ),
      pencil,
    );
    canvas.drawPath(
      Path()
        ..moveTo(-12, 54)
        ..lineTo(12, 54)
        ..lineTo(0, 73)
        ..close(),
      Paint()..color = const Color(0xFF8D5B36),
    );
    canvas.restore();
    _drawSpeechBubble(canvas, Offset(w * .24, h * .55), Colors.white,
        const Color(0xFFE96F43));
  }

  void _drawSocial(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    _drawChild(canvas, Offset(w * .36, h * .46),
        shirt: const Color(0xFF7FA1B5));
    _drawChild(canvas, Offset(w * .62, h * .47),
        shirt: const Color(0xFFE5A17A));
    final List<Color> colors = <Color>[
      const Color(0xFFE96F43),
      const Color(0xFFF6C45F),
      const Color(0xFF6F9F70),
      const Color(0xFF3F82D2),
    ];
    for (int i = 0; i < 6; i++) {
      final double x = w * (.32 + (i % 3) * .13);
      final double y = h * (.72 - (i ~/ 3) * .11);
      canvas.drawRRect(
        RRect.fromRectAndRadius(
          Rect.fromLTWH(x, y, 35, 30),
          const Radius.circular(7),
        ),
        Paint()..color = colors[i % colors.length].withOpacity(.88),
      );
    }
  }

  void _drawReview(Canvas canvas, Size size) {
    final double w = size.width;
    final double h = size.height;
    final Paint sheet = Paint()..color = Colors.white.withOpacity(.88);
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromLTWH(w * .25, h * .18, w * .44, h * .58),
        const Radius.circular(16),
      ),
      sheet,
    );
    final Paint line = Paint()
      ..color = const Color(0xFF7FA1B5)
      ..strokeWidth = 4
      ..strokeCap = StrokeCap.round
      ..style = PaintingStyle.stroke;
    canvas.drawPath(
      Path()
        ..moveTo(w * .32, h * .58)
        ..lineTo(w * .42, h * .48)
        ..lineTo(w * .5, h * .54)
        ..lineTo(w * .62, h * .35),
      line,
    );
    for (final double x in <double>[.32, .42, .5, .62]) {
      canvas.drawCircle(Offset(w * x, x == .62 ? h * .35 : h * .52), 5,
          Paint()..color = const Color(0xFF7FA1B5));
    }
    final Offset medal = Offset(w * .72, h * .62);
    canvas.drawCircle(medal, 31, Paint()..color = const Color(0xFFE6A13D));
    canvas.drawCircle(medal, 21, Paint()..color = const Color(0xFFFFDFA7));
    canvas.drawPath(
      Path()
        ..moveTo(medal.dx, medal.dy - 13)
        ..lineTo(medal.dx + 5, medal.dy - 2)
        ..lineTo(medal.dx + 17, medal.dy - 1)
        ..lineTo(medal.dx + 8, medal.dy + 6)
        ..lineTo(medal.dx + 11, medal.dy + 18)
        ..lineTo(medal.dx, medal.dy + 11)
        ..lineTo(medal.dx - 11, medal.dy + 18)
        ..lineTo(medal.dx - 8, medal.dy + 6)
        ..lineTo(medal.dx - 17, medal.dy - 1)
        ..lineTo(medal.dx - 5, medal.dy - 2)
        ..close(),
      Paint()..color = const Color(0xFFE6A13D),
    );
  }

  void _drawChild(Canvas canvas, Offset center, {required Color shirt}) {
    final Paint skin = Paint()..color = const Color(0xFFFFC79A);
    final Paint hair = Paint()..color = const Color(0xFF5E3C2A);
    canvas.drawCircle(center.translate(0, -28), 26, skin);
    canvas.drawArc(
      Rect.fromCircle(center: center.translate(0, -34), radius: 27),
      math.pi,
      math.pi,
      true,
      hair,
    );
    canvas.drawCircle(
        center.translate(-8, -31), 3, Paint()..color = Colors.black);
    canvas.drawCircle(
        center.translate(9, -31), 3, Paint()..color = Colors.black);
    canvas.drawArc(
      Rect.fromCenter(center: center.translate(0, -21), width: 15, height: 10),
      0,
      math.pi,
      false,
      Paint()
        ..color = const Color(0xFFB24A37)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 2.4,
    );
    canvas.drawRRect(
      RRect.fromRectAndRadius(
        Rect.fromCenter(center: center.translate(0, 18), width: 62, height: 64),
        const Radius.circular(24),
      ),
      Paint()..color = shirt,
    );
  }

  void _drawSpeechBubble(
    Canvas canvas,
    Offset center,
    Color fill,
    Color dotColor,
  ) {
    final RRect bubble = RRect.fromRectAndRadius(
      Rect.fromCenter(center: center, width: 84, height: 54),
      const Radius.circular(24),
    );
    canvas.drawRRect(bubble, Paint()..color = fill.withOpacity(.9));
    canvas.drawPath(
      Path()
        ..moveTo(center.dx - 20, center.dy + 20)
        ..lineTo(center.dx - 34, center.dy + 38)
        ..lineTo(center.dx - 10, center.dy + 24)
        ..close(),
      Paint()..color = fill.withOpacity(.9),
    );
    for (final double dx in <double>[-17, 0, 17]) {
      canvas.drawCircle(
        Offset(center.dx + dx, center.dy),
        5,
        Paint()..color = dotColor.withOpacity(.72),
      );
    }
  }

  @override
  bool shouldRepaint(covariant _ScaleCoverPainter oldDelegate) {
    return oldDelegate.type != type;
  }
}
