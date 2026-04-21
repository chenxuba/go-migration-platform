export function getNavLayout() {
	const systemInfo = uni.getSystemInfoSync()
	const statusBarHeight = systemInfo.statusBarHeight || 20
	const windowWidth = systemInfo.windowWidth || 375
	let menuButtonRect = null

	// #ifdef MP-WEIXIN
	if (typeof uni.getMenuButtonBoundingClientRect === 'function') {
		menuButtonRect = uni.getMenuButtonBoundingClientRect()
	}
	// #endif

	if (menuButtonRect && menuButtonRect.width) {
		return {
			statusBarHeight,
			top: menuButtonRect.top,
			height: menuButtonRect.height,
			width: menuButtonRect.width
		}
	}

	return {
		statusBarHeight,
		top: statusBarHeight + 8,
		height: 32,
		width: Math.max(96, Math.round(windowWidth * 0.24))
	}
}
