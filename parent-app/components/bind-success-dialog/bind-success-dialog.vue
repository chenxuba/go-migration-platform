<template>
	<view v-if="show" class="success-mask" @touchmove.stop.prevent>
		<view class="success-panel">
			<view class="success-cloud">
				<view class="success-bear">家</view>
			</view>
			<text class="success-title">恭喜你</text>
			<text class="success-desc">成功绑定“{{ studentName }}”的家庭</text>
			<text class="success-desc success-desc--sub">你可邀请家人进入此家庭，共同关注学员成长</text>
			<view class="success-action" @click="$emit('invite')">邀请家人</view>
		</view>
		<view class="success-close" @click="$emit('close')">×</view>
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
	background: rgba(26, 21, 12, 0.5);
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	padding: 40rpx;
	z-index: 95;
}

.success-panel {
	width: 100%;
	padding: 44rpx 40rpx 38rpx;
	border-radius: 38rpx;
	background: linear-gradient(180deg, #ffffff 0%, #fffdf8 100%);
	box-shadow: 0 30rpx 86rpx rgba(38, 30, 15, 0.2);
	text-align: center;
}

.success-cloud {
	position: relative;
	width: 100%;
	height: 220rpx;
	margin-bottom: 10rpx;
	border-radius: 30rpx 30rpx 110rpx 110rpx;
	background: linear-gradient(180deg, #3480ff 0%, #4a93ff 100%);
	overflow: hidden;
}

.success-cloud::before,
.success-cloud::after {
	content: '';
	position: absolute;
	bottom: -36rpx;
	width: 220rpx;
	height: 140rpx;
	border-radius: 120rpx;
	background: rgba(255, 255, 255, 0.96);
}

.success-cloud::before {
	left: -10rpx;
}

.success-cloud::after {
	right: -10rpx;
}

.success-bear {
	position: absolute;
	left: 50%;
	top: 18rpx;
	transform: translateX(-50%);
	width: 176rpx;
	height: 176rpx;
	border-radius: 88rpx;
	background: linear-gradient(180deg, #ffffff 0%, #fbfaf8 100%);
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ff5d7e;
	font-size: 72rpx;
	font-weight: 700;
	box-shadow: 0 16rpx 40rpx rgba(52, 128, 255, 0.28);
}

.success-title {
	display: block;
	font-size: 74rpx;
	font-weight: 700;
	color: #171717;
}

.success-desc {
	display: block;
	margin-top: 28rpx;
	font-size: 34rpx;
	line-height: 1.6;
	color: #2f2f2f;
}

.success-desc--sub {
	margin-top: 10rpx;
	color: #454545;
}

.success-action {
	display: flex;
	align-items: center;
	justify-content: center;
	height: 96rpx;
	margin-top: 34rpx;
	border-radius: 999rpx;
	background: linear-gradient(135deg, #ffe70b 0%, #ffd30c 100%);
	color: #2b250f;
	font-size: 34rpx;
	font-weight: 700;
}

.success-close {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 96rpx;
	height: 96rpx;
	margin-top: 34rpx;
	border-radius: 50%;
	border: 4rpx solid rgba(255, 255, 255, 0.9);
	color: #ffffff;
	font-size: 58rpx;
	line-height: 1;
}
</style>
