import 'dart:async';

import 'package:flutter/material.dart';

class AssessmentDraftResumeMetaRow {
  const AssessmentDraftResumeMetaRow({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;
}

class AssessmentDraftResumeDialog extends StatefulWidget {
  const AssessmentDraftResumeDialog({
    required this.message,
    required this.metaRows,
    required this.accentColor,
    required this.inkColor,
    required this.bodyColor,
    required this.lineColor,
    required this.lineSoftColor,
    required this.onRestart,
    required this.onContinue,
    this.closeDuration = const Duration(milliseconds: 140),
    this.metaBackgroundColor = const Color(0xFFFFFBF7),
    super.key,
  });

  final String message;
  final List<AssessmentDraftResumeMetaRow> metaRows;
  final Color accentColor;
  final Color inkColor;
  final Color bodyColor;
  final Color lineColor;
  final Color lineSoftColor;
  final Color metaBackgroundColor;
  final Future<void> Function() onRestart;
  final Future<bool> Function() onContinue;
  final Duration closeDuration;

  @override
  State<AssessmentDraftResumeDialog> createState() =>
      _AssessmentDraftResumeDialogState();
}

class _AssessmentDraftResumeDialogState
    extends State<AssessmentDraftResumeDialog> {
  bool _continuing = false;
  bool _closing = false;

  Future<void> _closeAfterShrink() async {
    if (_closing) {
      return;
    }
    setState(() => _closing = true);
    if (widget.closeDuration > Duration.zero) {
      await Future<void>.delayed(widget.closeDuration);
    }
    if (!mounted) {
      return;
    }
    Navigator.of(context).pop();
  }

  Future<void> _handleRestart() async {
    if (_continuing || _closing) {
      return;
    }
    await _closeAfterShrink();
    unawaited(widget.onRestart());
  }

  Future<void> _handleContinue() async {
    if (_continuing || _closing) {
      return;
    }
    setState(() => _continuing = true);
    await WidgetsBinding.instance.endOfFrame;
    if (!mounted || !_continuing) {
      return;
    }
    final bool restored = await widget.onContinue();
    if (!mounted) {
      return;
    }
    if (restored) {
      await _closeAfterShrink();
      return;
    }
    setState(() => _continuing = false);
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedOpacity(
      opacity: _closing ? 0 : 1,
      duration: widget.closeDuration,
      curve: Curves.easeInCubic,
      child: AnimatedScale(
        scale: _closing ? .92 : 1,
        duration: widget.closeDuration,
        curve: Curves.easeInCubic,
        child: Dialog(
          insetPadding: const EdgeInsets.symmetric(horizontal: 32),
          backgroundColor: Colors.transparent,
          child: Container(
            width: 520,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(14),
              boxShadow: const <BoxShadow>[
                BoxShadow(
                  color: Color(0x33000000),
                  blurRadius: 30,
                  offset: Offset(0, 18),
                ),
              ],
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Padding(
                  padding: const EdgeInsets.fromLTRB(20, 20, 20, 20),
                  child: Text(
                    '发现未完成草稿',
                    style: TextStyle(
                      color: widget.inkColor,
                      fontSize: 19,
                      height: 1,
                      fontWeight: FontWeight.w900,
                    ),
                  ),
                ),
                Divider(height: 1, color: widget.lineSoftColor),
                Padding(
                  padding: const EdgeInsets.fromLTRB(30, 30, 30, 30),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: <Widget>[
                      Text(
                        widget.message,
                        style: TextStyle(
                          color: widget.inkColor,
                          fontSize: 15,
                          height: 1.2,
                          fontWeight: FontWeight.w800,
                        ),
                      ),
                      const SizedBox(height: 22),
                      Container(
                        width: double.infinity,
                        padding: const EdgeInsets.fromLTRB(18, 16, 18, 16),
                        decoration: BoxDecoration(
                          color: widget.metaBackgroundColor,
                          borderRadius: BorderRadius.circular(10),
                          border: Border.all(color: widget.lineColor),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: <Widget>[
                            for (int index = 0;
                                index < widget.metaRows.length;
                                index += 1) ...<Widget>[
                              if (index > 0) const SizedBox(height: 13),
                              _AssessmentDraftResumeMeta(
                                row: widget.metaRows[index],
                                inkColor: widget.inkColor,
                                bodyColor: widget.bodyColor,
                              ),
                            ],
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                Divider(height: 1, color: widget.lineSoftColor),
                Padding(
                  padding: const EdgeInsets.fromLTRB(30, 18, 30, 20),
                  child: _AssessmentDraftResumeActionArea(
                    continuing: _continuing,
                    accentColor: widget.accentColor,
                    inkColor: widget.inkColor,
                    lineColor: widget.lineColor,
                    onRestart: _handleRestart,
                    onContinue: _handleContinue,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _AssessmentDraftResumeActionArea extends StatelessWidget {
  const _AssessmentDraftResumeActionArea({
    required this.continuing,
    required this.accentColor,
    required this.inkColor,
    required this.lineColor,
    required this.onRestart,
    required this.onContinue,
  });

  final bool continuing;
  final Color accentColor;
  final Color inkColor;
  final Color lineColor;
  final VoidCallback onRestart;
  final VoidCallback onContinue;

  @override
  Widget build(BuildContext context) {
    return Align(
      alignment: Alignment.centerRight,
      child: SizedBox(
        width: 236,
        height: 42,
        child: AnimatedSwitcher(
          duration:
              continuing ? Duration.zero : const Duration(milliseconds: 120),
          reverseDuration: Duration.zero,
          switchInCurve: Curves.easeOutCubic,
          switchOutCurve: Curves.easeOutCubic,
          layoutBuilder: (
            Widget? currentChild,
            List<Widget> previousChildren,
          ) {
            return Stack(
              alignment: Alignment.centerRight,
              children: <Widget>[
                ...previousChildren,
                if (currentChild != null) currentChild,
              ],
            );
          },
          transitionBuilder: (Widget child, Animation<double> animation) {
            if (continuing) {
              return child;
            }
            return FadeTransition(opacity: animation, child: child);
          },
          child: continuing
              ? _AssessmentDraftResumeLoadingButton(
                  key: const ValueKey<String>('draft-resume-loading'),
                  accentColor: accentColor,
                )
              : Row(
                  key: const ValueKey<String>('draft-resume-actions'),
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: <Widget>[
                    _AssessmentDraftResumeActionButton(
                      label: '重新测评',
                      filled: false,
                      accentColor: accentColor,
                      inkColor: inkColor,
                      lineColor: lineColor,
                      onTap: onRestart,
                    ),
                    const SizedBox(width: 12),
                    _AssessmentDraftResumeActionButton(
                      label: '继续测评',
                      filled: true,
                      accentColor: accentColor,
                      inkColor: inkColor,
                      lineColor: lineColor,
                      onTap: onContinue,
                    ),
                  ],
                ),
        ),
      ),
    );
  }
}

class _AssessmentDraftResumeLoadingButton extends StatelessWidget {
  const _AssessmentDraftResumeLoadingButton({
    required this.accentColor,
    super.key,
  });

  final Color accentColor;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 236,
      height: 42,
      decoration: BoxDecoration(
        color: accentColor,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: accentColor),
      ),
      child: const Center(
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            SizedBox(
              width: 15,
              height: 15,
              child: CircularProgressIndicator(
                strokeWidth: 2.2,
                color: Colors.white,
              ),
            ),
            SizedBox(width: 8),
            Text(
              '题目填充中，请稍后...',
              style: TextStyle(
                color: Colors.white,
                fontSize: 14,
                fontWeight: FontWeight.w900,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _AssessmentDraftResumeActionButton extends StatelessWidget {
  const _AssessmentDraftResumeActionButton({
    required this.label,
    required this.filled,
    required this.accentColor,
    required this.inkColor,
    required this.lineColor,
    required this.onTap,
  });

  final String label;
  final bool filled;
  final Color accentColor;
  final Color inkColor;
  final Color lineColor;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(10),
        child: Ink(
          width: 112,
          height: 42,
          decoration: BoxDecoration(
            color: filled ? accentColor : Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: filled ? accentColor : lineColor),
          ),
          child: Center(
            child: Text(
              label,
              style: TextStyle(
                color: filled ? Colors.white : inkColor,
                fontSize: 14,
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _AssessmentDraftResumeMeta extends StatelessWidget {
  const _AssessmentDraftResumeMeta({
    required this.row,
    required this.inkColor,
    required this.bodyColor,
  });

  final AssessmentDraftResumeMetaRow row;
  final Color inkColor;
  final Color bodyColor;

  @override
  Widget build(BuildContext context) {
    return Text.rich(
      TextSpan(
        children: <InlineSpan>[
          TextSpan(text: '${row.label}：'),
          TextSpan(
            text: row.value,
            style: TextStyle(
              color: inkColor,
              fontWeight: FontWeight.w900,
            ),
          ),
        ],
      ),
      style: TextStyle(
        color: bodyColor,
        fontSize: 14,
        height: 1.2,
        fontWeight: FontWeight.w800,
      ),
    );
  }
}
