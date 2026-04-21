<script>
export default {
	onLaunch() {
		console.log('Parent App Launch')
	},
	onShow() {
		console.log('Parent App Show')
	},
	onHide() {
		console.log('Parent App Hide')
	}
}
</script>

<style lang="scss">
page {
	--parent-bg: #fffaf0;
	--parent-bg-soft: #fff4e1;
	--parent-card: rgba(255, 255, 255, 0.9);
	--parent-card-solid: #ffffff;
	--parent-line: rgba(225, 208, 176, 0.75);
	--parent-text: #1f1f1f;
	--parent-subtext: #7a7466;
	--parent-muted: #b8b2a4;
	--parent-brand: #ffd60a;
	--parent-brand-deep: #efb900;
	--parent-blue: #2f86ff;
	--parent-green: #22bf73;
	--parent-danger: #ff7a66;
	--parent-shadow: 0 24rpx 60rpx rgba(159, 124, 67, 0.12);
	background:
		radial-gradient(circle at 8% 8%, rgba(255, 216, 162, 0.35), transparent 24%),
		radial-gradient(circle at 100% 0%, rgba(255, 232, 183, 0.45), transparent 26%),
		linear-gradient(180deg, #fff6e9 0%, #fffdf8 40%, #ffffff 100%);
	color: var(--parent-text);
	font-size: 24rpx;
	min-height: 100%;
}

view,
text,
button {
	box-sizing: border-box;
}

button {
	margin: 0;
	padding: 0;
	line-height: 1;
	background: transparent;
}

button::after {
	border: none;
}

.parent-page {
	min-height: 100vh;
}

.parent-shell {
	padding: 0 24rpx 56rpx;
}

.parent-header {
	padding-bottom: 24rpx;
}

.parent-nav-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.parent-nav-spacer {
	flex-shrink: 0;
}

.parent-card {
	background: var(--parent-card);
	border: 1rpx solid rgba(255, 255, 255, 0.88);
	border-radius: 28rpx;
	box-shadow: var(--parent-shadow);
	backdrop-filter: blur(14rpx);
}

.parent-title {
	font-size: 52rpx;
	font-weight: 700;
	line-height: 1.12;
	color: var(--parent-text);
}

.parent-subtitle {
	margin-top: 10rpx;
	font-size: 24rpx;
	line-height: 1.6;
	color: var(--parent-subtext);
}

.parent-primary-button {
	display: flex;
	align-items: center;
	justify-content: center;
	height: 84rpx;
	border-radius: 999rpx;
	background: linear-gradient(135deg, #ffe60d 0%, #ffd10b 100%);
	box-shadow: 0 20rpx 42rpx rgba(255, 214, 10, 0.28);
	color: #2b250c;
	font-size: 28rpx;
	font-weight: 700;
}

.parent-secondary-button {
	display: flex;
	align-items: center;
	justify-content: center;
	height: 78rpx;
	border-radius: 26rpx;
	border: 1rpx solid var(--parent-line);
	background: rgba(255, 255, 255, 0.82);
	color: var(--parent-green);
	font-size: 26rpx;
	font-weight: 600;
}

.parent-empty-card {
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 60rpx 32rpx;
	text-align: center;
}

.parent-empty-badge {
	width: 148rpx;
	height: 148rpx;
	border-radius: 50%;
	background: linear-gradient(180deg, rgba(255, 255, 255, 0.95), rgba(244, 239, 228, 0.92));
	box-shadow: inset 0 0 0 1rpx rgba(255, 255, 255, 0.96);
	display: flex;
	align-items: center;
	justify-content: center;
	color: #c3bcaf;
	font-size: 62rpx;
	font-weight: 700;
}

.parent-empty-title {
	margin-top: 22rpx;
	font-size: 30rpx;
	font-weight: 700;
}

.parent-empty-desc {
	margin-top: 12rpx;
	font-size: 24rpx;
	line-height: 1.6;
	color: var(--parent-subtext);
}
</style>
