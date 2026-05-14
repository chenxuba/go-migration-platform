part of 'erxin_assessment_page.dart';

class _CurrentItemsEmptyState extends StatelessWidget {
  const _CurrentItemsEmptyState({required this.text});

  final String text;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Container(
        width: 420,
        padding: const EdgeInsets.fromLTRB(22, 20, 22, 18),
        decoration: BoxDecoration(
          color: const Color(0xFFFFFAF5),
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: _ErxinColors.line),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: <Widget>[
            const Icon(
              Icons.fact_check_outlined,
              color: _ErxinColors.blue,
              size: 34,
            ),
            const SizedBox(height: 10),
            const Text(
              '当前题目已完成',
              style: TextStyle(
                color: _ErxinColors.ink,
                fontSize: 16,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              text,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.center,
              style: const TextStyle(
                color: _ErxinColors.body,
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

class _HistoryChangeConfirmDialog extends StatelessWidget {
  const _HistoryChangeConfirmDialog({
    required this.itemTitle,
    required this.nextStatus,
    required this.onCancel,
    required this.onConfirm,
  });

  final String itemTitle;
  final String nextStatus;
  final VoidCallback onCancel;
  final VoidCallback onConfirm;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: Container(
        width: 438,
        padding: const EdgeInsets.fromLTRB(22, 20, 22, 18),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: _ErxinColors.line),
          boxShadow: _erxinShadow(
            color: const Color(0x24172033),
            blur: 28,
            offset: const Offset(0, 14),
          ),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            const Text(
              '确认修改历史记录',
              style: TextStyle(
                color: _ErxinColors.ink,
                fontSize: 20,
                fontWeight: FontWeight.w900,
              ),
            ),
            const SizedBox(height: 10),
            Text(
              '$itemTitle 将改为“$nextStatus”。修改后会重新计算当前能区的前测基线、后测封顶和后续需测月龄。',
              style: const TextStyle(
                color: _ErxinColors.body,
                fontSize: 14,
                height: 1.45,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 18),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: <Widget>[
                SizedBox(
                  height: 36,
                  child: OutlinedButton(
                    onPressed: onCancel,
                    child: const Text('取消'),
                  ),
                ),
                const SizedBox(width: 10),
                SizedBox(
                  height: 36,
                  child: FilledButton(
                    onPressed: onConfirm,
                    style: FilledButton.styleFrom(
                      backgroundColor: _ErxinColors.orange,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    child: const Text(
                      '确认修改',
                      style: TextStyle(fontWeight: FontWeight.w900),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

class _ProgressSummary extends StatelessWidget {
  const _ProgressSummary({
    required this.domainStatus,
    required this.scaleStatus,
  });

  final String domainStatus;
  final String scaleStatus;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      height: _erxinProgressSummaryHeight,
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: _ErxinColors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '完成情况',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 15,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 6),
          Text(
            domainStatus,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: _summaryStyle,
          ),
          const SizedBox(height: 4),
          Text(
            '全量表：$scaleStatus',
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: _summaryStyle,
          ),
        ],
      ),
    );
  }

  static const TextStyle _summaryStyle = TextStyle(
    color: _ErxinColors.body,
    fontSize: 13,
    fontWeight: FontWeight.w700,
  );
}

class _SmallBadge extends StatelessWidget {
  const _SmallBadge({required this.text, this.strong = false});

  final String text;
  final bool strong;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 30,
      padding: const EdgeInsets.symmetric(horizontal: 11),
      decoration: BoxDecoration(
        color: strong ? const Color(0xFFFFF1E8) : const Color(0xFFFFFAF5),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: strong ? _ErxinColors.blue : _ErxinColors.line,
        ),
      ),
      alignment: Alignment.center,
      child: Text(
        text,
        style: TextStyle(
          color: strong ? _ErxinColors.blue : _ErxinColors.body,
          fontSize: 13,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _MiniMarker extends StatelessWidget {
  const _MiniMarker({required this.text, this.warning = false});

  final String text;
  final bool warning;

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 20,
      width: 20,
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: warning ? const Color(0xFFFFF2E8) : const Color(0xFFFFF1E8),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        text,
        style: TextStyle(
          color: warning ? const Color(0xFFEA580C) : _ErxinColors.blue,
          fontSize: 12,
          fontWeight: FontWeight.w900,
        ),
      ),
    );
  }
}

class _DialogTextBlock extends StatelessWidget {
  const _DialogTextBlock({required this.title, required this.text});

  final String title;
  final String text;

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: <Widget>[
        Text(
          title,
          style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w900),
        ),
        const SizedBox(height: 8),
        Text(
          text.trim().isEmpty ? '暂无内容' : text.trim(),
          style: const TextStyle(fontSize: 14, height: 1.55),
        ),
      ],
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.message, required this.onBack});

  final String message;
  final VoidCallback onBack;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: <Widget>[
          Text(message, style: const TextStyle(color: _ErxinColors.red)),
          const SizedBox(height: 16),
          OutlinedButton(onPressed: onBack, child: const Text('返回')),
        ],
      ),
    );
  }
}
