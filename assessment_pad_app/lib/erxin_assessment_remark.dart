part of 'erxin_assessment_page.dart';

class _RightRemarkSection extends StatelessWidget {
  const _RightRemarkSection({
    required this.height,
    required this.itemNo,
    required this.remark,
    required this.onChanged,
    required this.onEditingComplete,
  });

  final double height;
  final int itemNo;
  final String remark;
  final void Function(int itemNo, String remark) onChanged;
  final ValueChanged<int> onEditingComplete;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: height,
      child: Padding(
        padding: const EdgeInsets.only(top: _erxinDetailPanelTopPadding),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            const SizedBox(
              height: _erxinDetailHeaderHeight,
              child: Align(
                alignment: Alignment.centerLeft,
                child: Text(
                  '题目备注',
                  style: TextStyle(
                    color: _ErxinColors.ink,
                    fontSize: 15,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
            ),
            const SizedBox(height: _erxinDetailHeaderGap),
            SizedBox(
              height: _erxinDetailContentHeight,
              child: _RemarkBox(
                itemNo: itemNo,
                remark: remark,
                onChanged: onChanged,
                onEditingComplete: onEditingComplete,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _RemarkBox extends StatelessWidget {
  const _RemarkBox({
    required this.itemNo,
    required this.remark,
    required this.onChanged,
    required this.onEditingComplete,
  });

  final int itemNo;
  final String remark;
  final void Function(int itemNo, String remark) onChanged;
  final ValueChanged<int> onEditingComplete;

  @override
  Widget build(BuildContext context) {
    final bool enabled = itemNo > 0;
    final String preview = remark.trim();
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: enabled ? () => _openRemarkEditor(context) : null,
        borderRadius: BorderRadius.circular(8),
        child: Ink(
          padding: const EdgeInsets.all(9),
          decoration: BoxDecoration(
            color: const Color(0xFFFFFAF5),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: _ErxinColors.line),
          ),
          child: Align(
            alignment: Alignment.topLeft,
            child: Text(
              enabled
                  ? preview.isEmpty
                      ? '添加本题备注'
                      : preview
                  : '选择题目后添加备注',
              maxLines: 4,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                color: preview.isEmpty ? _ErxinColors.muted : _ErxinColors.body,
                fontSize: 13,
                height: 1.25,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _openRemarkEditor(BuildContext context) async {
    bool changed = false;
    await showDialog<void>(
      context: context,
      barrierColor: Colors.black.withOpacity(.18),
      builder: (BuildContext dialogContext) {
        return PadDialogViewport(
          alignment: Alignment.topCenter,
          child: _RemarkEditorDialog(
            initialValue: remark,
            onChanged: (String value) {
              changed = true;
              onChanged(itemNo, value);
            },
            onClear: () {
              changed = true;
              onChanged(itemNo, '');
            },
          ),
        );
      },
    );
    if (changed) {
      onEditingComplete(itemNo);
    }
  }
}

class _RemarkEditorDialog extends StatefulWidget {
  const _RemarkEditorDialog({
    required this.initialValue,
    required this.onChanged,
    required this.onClear,
  });

  final String initialValue;
  final ValueChanged<String> onChanged;
  final VoidCallback onClear;

  @override
  State<_RemarkEditorDialog> createState() => _RemarkEditorDialogState();
}

class _RemarkEditorDialogState extends State<_RemarkEditorDialog> {
  late final TextEditingController _controller;

  @override
  void initState() {
    super.initState();
    _controller = TextEditingController(text: widget.initialValue);
    _controller.selection = TextSelection.collapsed(
      offset: _controller.text.length,
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final double keyboardBottom = MediaQuery.viewInsetsOf(context).bottom;
    return AnimatedPadding(
      duration: const Duration(milliseconds: 180),
      curve: Curves.easeOutCubic,
      padding: EdgeInsets.fromLTRB(24, 108, 24, keyboardBottom + 24),
      child: Align(
        alignment: Alignment.topCenter,
        child: Material(
          color: Colors.white,
          borderRadius: BorderRadius.circular(10),
          clipBehavior: Clip.antiAlias,
          child: Container(
            width: 430,
            height: 258,
            padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
            decoration: BoxDecoration(
              border: Border.all(color: _ErxinColors.line),
              borderRadius: BorderRadius.circular(10),
              boxShadow: _erxinShadow(
                color: const Color(0x22000000),
                blur: 22,
                offset: const Offset(0, 10),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: <Widget>[
                Row(
                  children: <Widget>[
                    const Expanded(
                      child: Text(
                        '题目备注',
                        style: TextStyle(
                          color: _ErxinColors.ink,
                          fontSize: 16,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                    IconButton(
                      onPressed: () => Navigator.of(context).pop(),
                      icon: const Icon(Icons.close_rounded),
                      iconSize: 20,
                      visualDensity: VisualDensity.compact,
                      tooltip: '关闭',
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Expanded(
                  child: TextField(
                    controller: _controller,
                    autofocus: true,
                    maxLines: null,
                    expands: true,
                    textAlignVertical: TextAlignVertical.top,
                    onChanged: widget.onChanged,
                    onTapOutside: (_) =>
                        FocusManager.instance.primaryFocus?.unfocus(),
                    style: const TextStyle(fontSize: 14, height: 1.35),
                    decoration: InputDecoration(
                      hintText: '添加本题备注',
                      filled: true,
                      fillColor: const Color(0xFFFFFAF5),
                      contentPadding: const EdgeInsets.all(10),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: const BorderSide(color: _ErxinColors.line),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: const BorderSide(color: _ErxinColors.line),
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 10),
                Row(
                  children: <Widget>[
                    TextButton(
                      onPressed: () {
                        _controller.clear();
                        widget.onClear();
                      },
                      child: const Text('清空'),
                    ),
                    const Spacer(),
                    FilledButton(
                      onPressed: () => Navigator.of(context).pop(),
                      style: FilledButton.styleFrom(
                        backgroundColor: _ErxinColors.blue,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                      ),
                      child: const Text('完成'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
