part of 'erxin_assessment_page.dart';

class _DomainSidebar extends StatelessWidget {
  const _DomainSidebar({
    required this.domains,
    required this.selectedCode,
    required this.progressForDomain,
    required this.domainCompleteForDomain,
    required this.completedDomainCount,
    required this.savedItemCount,
    required this.onSelect,
    required this.onShowAllItems,
  });

  final List<ErxinDomain> domains;
  final String selectedCode;
  final _DomainProgress Function(String domainCode) progressForDomain;
  final bool Function(String domainCode) domainCompleteForDomain;
  final int completedDomainCount;
  final int savedItemCount;
  final ValueChanged<String> onSelect;
  final VoidCallback onShowAllItems;

  @override
  Widget build(BuildContext context) {
    final _DomainProgress selectedProgress = progressForDomain(selectedCode);
    final bool selectedComplete = domainCompleteForDomain(selectedCode);
    final bool selectedVisibleComplete = selectedProgress.total > 0 &&
        selectedProgress.answered >= selectedProgress.total;
    return Container(
      width: 214,
      padding: const EdgeInsets.fromLTRB(
        14,
        16,
        12,
        _erxinSidebarBottomPadding,
      ),
      decoration: _erxinPanelDecoration(),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: <Widget>[
          const Text(
            '能区进度',
            style: TextStyle(
              color: _ErxinColors.ink,
              fontSize: 17,
              fontWeight: FontWeight.w900,
            ),
          ),
          const SizedBox(height: 12),
          for (final ErxinDomain domain in domains)
            _DomainRow(
              domain: domain,
              selected: domain.domainCode == selectedCode,
              progress: progressForDomain(domain.domainCode),
              complete: domainCompleteForDomain(domain.domainCode),
              onTap: () => onSelect(domain.domainCode),
            ),
          const SizedBox(height: 2),
          _AllItemsButton(onTap: onShowAllItems),
          const Spacer(),
          _ProgressSummary(
            domainStatus: selectedComplete
                ? '本能区：已完成'
                : selectedVisibleComplete
                    ? '本能区：当前可见完成（待推进）'
                    : selectedProgress.answered > 0
                        ? '本能区：测查中'
                        : '本能区：待测',
            scaleStatus: savedItemCount > 0
                ? '$completedDomainCount/5 能区完成\n已保存$savedItemCount题'
                : '$completedDomainCount/5 能区完成',
          ),
          const SizedBox(height: _erxinProgressSummaryBottomGap),
        ],
      ),
    );
  }
}

class _DomainRow extends StatelessWidget {
  const _DomainRow({
    required this.domain,
    required this.selected,
    required this.progress,
    required this.complete,
    required this.onTap,
  });

  final ErxinDomain domain;
  final bool selected;
  final _DomainProgress progress;
  final bool complete;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final bool visibleComplete =
        progress.total > 0 && progress.answered >= progress.total;
    final double percent =
        progress.total <= 0 ? 0 : progress.answered / progress.total;
    final String status = complete
        ? '已完成'
        : visibleComplete
            ? '待推进'
            : progress.answered > 0
                ? '测查中'
                : '待测';
    final Color statusColor = complete
        ? _ErxinColors.green
        : (selected || visibleComplete)
            ? _ErxinColors.blue
            : _ErxinColors.muted;
    return Padding(
      padding: const EdgeInsets.only(bottom: 9),
      child: InkWell(
        borderRadius: BorderRadius.circular(8),
        onTap: onTap,
        child: Container(
          height: 66,
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            color: selected ? const Color(0xFFFFEEE5) : Colors.white,
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: _ErxinColors.line),
          ),
          child: Column(
            children: <Widget>[
              Row(
                children: <Widget>[
                  _DomainIcon(
                    icon: _domainIconFor(domain),
                    selected: selected,
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      domain.domainName,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        color: selected ? _ErxinColors.blue : _ErxinColors.ink,
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                      ),
                    ),
                  ),
                  Text(
                    '${progress.answered}/${progress.total}',
                    style: const TextStyle(
                      color: _ErxinColors.body,
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ],
              ),
              const Spacer(),
              Row(
                children: <Widget>[
                  Expanded(
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(2),
                      child: LinearProgressIndicator(
                        value: percent.clamp(0, 1),
                        minHeight: 4,
                        backgroundColor: const Color(0xFFF2E6DC),
                        valueColor: AlwaysStoppedAnimation<Color>(
                          complete ? _ErxinColors.green : _ErxinColors.blue,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 9),
                  Text(
                    status,
                    style: TextStyle(
                      color: statusColor,
                      fontSize: 12,
                      height: 1,
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _DomainIcon extends StatelessWidget {
  const _DomainIcon({required this.icon, required this.selected});

  final IconData icon;
  final bool selected;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 24,
      height: 24,
      decoration: BoxDecoration(
        color: selected ? _ErxinColors.blue : const Color(0xFFFFF2EA),
        borderRadius: BorderRadius.circular(7),
      ),
      alignment: Alignment.center,
      child: Icon(
        icon,
        size: 15,
        color: selected ? Colors.white : _ErxinColors.body,
      ),
    );
  }
}

IconData _domainIconFor(ErxinDomain domain) {
  final String code = domain.domainCode.toUpperCase();
  final String name = domain.domainName;
  if (code == 'GM' || name.contains('大运动')) {
    return Icons.directions_run_rounded;
  }
  if (code == 'FM' || name.contains('精细')) {
    return Icons.gesture_rounded;
  }
  if (code == 'AD' || name.contains('适应')) {
    return Icons.psychology_alt_rounded;
  }
  if (code == 'LANG' || name.contains('语言')) {
    return Icons.record_voice_over_rounded;
  }
  if (code == 'SOC' || name.contains('社会') || name.contains('社交')) {
    return Icons.groups_2_rounded;
  }
  return Icons.extension_rounded;
}

class _AllItemsButton extends StatelessWidget {
  const _AllItemsButton({this.onTap});

  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Ink(
          height: 38,
          padding: const EdgeInsets.symmetric(horizontal: 12),
          decoration: BoxDecoration(
            color: const Color(0xFFFFFAF5),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: _ErxinColors.line),
          ),
          child: Row(
            children: const <Widget>[
              Icon(
                Icons.list_alt_rounded,
                size: 17,
                color: _ErxinColors.blue,
              ),
              SizedBox(width: 8),
              Expanded(
                child: Text(
                  '查看全部题目',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    color: _ErxinColors.blue,
                    fontSize: 13,
                    fontWeight: FontWeight.w900,
                  ),
                ),
              ),
              Icon(
                Icons.chevron_right_rounded,
                size: 18,
                color: _ErxinColors.blue,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
