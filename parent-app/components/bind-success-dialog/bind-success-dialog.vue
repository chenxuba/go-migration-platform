<template>
	<view v-if="show" class="success-mask" @click.self="$emit('close')" @touchmove.stop.prevent>
		<view class="success-panel" @click.stop>
			<view class="success-close" @click="$emit('close')">
				<view class="success-close__line success-close__line--first"></view>
				<view class="success-close__line success-close__line--second"></view>
			</view>

			<view class="success-hero">
				<view class="success-orbit success-orbit--left"></view>
				<view class="success-orbit success-orbit--right"></view>
				<view class="success-badge">
					<view class="success-badge__ring"></view>
					<view class="success-badge__core">
						<view class="success-badge__check"></view>
					</view>
				</view>
				<view class="success-chip">绑定完成</view>
			</view>

			<view class="success-copy">
				<text class="success-title">已成功加入家庭</text>
				<view class="success-student-pill">
					<text class="success-student-pill__label">绑定学员</text>
					<text class="success-student-pill__name">{{ studentName || '新学员' }}</text>
				</view>
				<text class="success-desc">课表、系统通知和订单等内容会自动同步到当前家长端。</text>
			</view>

			<view class="success-summary">
				<view class="success-summary__item">
					<text class="success-summary__value">课表</text>
					<text class="success-summary__label">已同步</text>
				</view>
				<view class="success-summary__divider"></view>
				<view class="success-summary__item">
					<text class="success-summary__value">通知</text>
					<text class="success-summary__label">可查看</text>
				</view>
				<view class="success-summary__divider"></view>
				<view class="success-summary__item">
					<text class="success-summary__value">家庭</text>
					<text class="success-summary__label">可邀请</text>
				</view>
			</view>

			<view class="success-actions">
				<view class="parent-primary-button success-actions__primary" @click="$emit('invite')">邀请家人一起查看</view>
				<view class="success-actions__secondary" @click="$emit('close')">先看看</view>
			</view>
		</view>
	</view>
</template>

<script setup>
defineProps({
	show: {
		type: Boolean,
		default: false
	},
	studentName: {
		type: String,
		default: ''
	}
})

defineEmits(['close', 'invite'])
</script>

<style scoped>
.success-mask {
	position: fixed;
	inset: 0;
	background:
		radial-gradient(circle at 50% 18%, rgba(255, 231, 121, 0.16), transparent 24%),
		rgba(25, 20, 13, 0.42);
	backdrop-filter: blur(10rpx);
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 36rpx;
	z-index: 95;
}

.success-panel {
	position: relative;
	width: 100%;
	padding: 34rpx 30rpx 28rpx;
	border-radius: 40rpx;
	background:
		radial-gradient(circle at top, rgba(255, 243, 194, 0.86), transparent 34%),
		linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(255, 252, 245, 0.98) 100%);
	box-shadow: 0 32rpx 88rpx rgba(40, 31, 17, 0.22);
	overflow: hidden;
}

.success-panel::after {
	content: '';
	position: absolute;
	left: 24rpx;
	right: 24rpx;
	top: 24rpx;
	height: 1rpx;
	background: linear-gradient(90deg, transparent 0%, rgba(255, 213, 80, 0.5) 50%, transparent 100%);
}

.success-close {
	position: absolute;
	top: 20rpx;
	right: 20rpx;
	width: 56rpx;
	height: 56rpx;
	border-radius: 18rpx;
	background: rgba(248, 244, 236, 0.96);
	display: flex;
	align-items: center;
	justify-content: center;
}

.success-close__line {
	position: absolute;
	width: 22rpx;
	height: 3rpx;
	border-radius: 999rpx;
	background: #857b69;
}

.success-close__line--first {
	transform: rotate(45deg);
}

.success-close__line--second {
	transform: rotate(-45deg);
}

.success-hero {
	position: relative;
	padding-top: 18rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
}

.success-orbit {
	position: absolute;
	top: 48rpx;
	width: 116rpx;
	height: 116rpx;
	border-radius: 50%;
	border: 2rpx solid rgba(255, 214, 10, 0.18);
}

.success-orbit--left {
	left: 78rpx;
}

.success-orbit--right {
	right: 78rpx;
}

.success-badge {
	position: relative;
	width: 164rpx;
	height: 164rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.success-badge__ring {
	position: absolute;
	inset: 0;
	border-radius: 50%;
	background: radial-gradient(circle, rgba(255, 214, 10, 0.28) 0%, rgba(255, 214, 10, 0.08) 54%, transparent 72%);
}

.success-badge__core {
	position: relative;
	width: 108rpx;
	height: 108rpx;
	border-radius: 34rpx;
	background: linear-gradient(180deg, #ffe670 0%, #ffd20f 100%);
	box-shadow: 0 18rpx 36rpx rgba(255, 214, 10, 0.26);
	display: flex;
	align-items: center;
	justify-content: center;
}

.success-badge__check {
	width: 34rpx;
	height: 18rpx;
	border-left: 6rpx solid #ffffff;
	border-bottom: 6rpx solid #ffffff;
	transform: rotate(-45deg) translateY(-6rpx);
}

.success-chip {
	margin-top: -4rpx;
	height: 44rpx;
	padding: 0 20rpx;
	border-radius: 999rpx;
	background: rgba(255, 244, 209, 0.92);
	color: #866200;
	font-size: 20rpx;
	font-weight: 700;
	display: inline-flex;
	align-items: center;
	justify-content: center;
	letter-spacing: 1rpx;
}

.success-copy {
	margin-top: 18rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	text-align: center;
}

.success-title {
	display: block;
	font-size: 40rpx;
	font-weight: 700;
	line-height: 1.24;
	color: #1f1f1f;
}

.success-student-pill {
	display: inline-flex;
	align-items: center;
	gap: 12rpx;
	margin-top: 18rpx;
	padding: 14rpx 18rpx;
	border-radius: 22rpx;
	background: rgba(249, 246, 239, 0.98);
	border: 1rpx solid rgba(238, 226, 194, 0.76);
}

.success-student-pill__label {
	font-size: 20rpx;
	color: #8d846f;
	font-weight: 600;
}

.success-student-pill__name {
	font-size: 26rpx;
	color: #262626;
	font-weight: 700;
}

.success-desc {
	display: block;
	margin-top: 18rpx;
	font-size: 24rpx;
	line-height: 1.7;
	color: #6e6758;
}

.success-summary {
	display: flex;
	align-items: center;
	justify-content: space-between;
	margin-top: 26rpx;
	padding: 24rpx 18rpx;
	border-radius: 28rpx;
	background: linear-gradient(180deg, rgba(255, 249, 233, 0.9) 0%, rgba(255, 255, 255, 0.92) 100%);
	border: 1rpx solid rgba(245, 226, 176, 0.62);
}

.success-summary__item {
	flex: 1;
	display: flex;
	flex-direction: column;
	align-items: center;
	text-align: center;
}

.success-summary__value {
	font-size: 26rpx;
	font-weight: 700;
	color: #2b2924;
}

.success-summary__label {
	margin-top: 8rpx;
	font-size: 20rpx;
	color: #8a806f;
}

.success-summary__divider {
	width: 1rpx;
	height: 42rpx;
	background: rgba(231, 216, 181, 0.86);
}

.success-actions {
	margin-top: 24rpx;
}

.success-actions__primary {
	height: 84rpx;
	font-size: 28rpx;
}

.success-actions__secondary {
	margin-top: 16rpx;
	height: 80rpx;
	border-radius: 999rpx;
	background: rgba(247, 243, 236, 0.96);
	color: #6f6758;
	font-size: 24rpx;
	font-weight: 600;
	display: flex;
	align-items: center;
	justify-content: center;
}
</style>
