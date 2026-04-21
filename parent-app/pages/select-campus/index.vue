<template>
	<view class="parent-page">
		<view class="parent-shell">
			<view class="parent-header campus-header" :style="{ paddingTop: `${statusBarHeight + 18}px` }">
				<view class="campus-header__back" @click="goBack">‹</view>
				<text class="campus-header__title">选择学校</text>
				<view class="campus-header__placeholder"></view>
			</view>

			<view class="campus-label">当前校区</view>
			<view
				v-for="item in campusList"
				:key="item.id"
				class="parent-card campus-item"
				@click="selectCampus(item.id)"
			>
				<view class="campus-item__left">
					<view class="campus-item__avatar">{{ item.shortName }}</view>
					<text class="campus-item__name">{{ item.name }}</text>
				</view>
				<text v-if="currentCampusId === item.id" class="campus-item__check">√</text>
			</view>
		</view>
	</view>
</template>

<script setup>
import { computed } from 'vue'
import { parentState, switchCurrentCampus } from '@/common/parent-state'

const statusBarHeight = uni.getSystemInfoSync().statusBarHeight || 0
const campusList = computed(() => parentState.campusList)
const currentCampusId = computed(() => parentState.currentCampusId)

function selectCampus(campusId) {
	switchCurrentCampus(campusId)
	goBack()
}

function goBack() {
	uni.navigateBack({
		fail() {
			uni.switchTab({
				url: '/pages/home/index'
			})
		}
	})
}
</script>

<style scoped>
.campus-header {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.campus-header__back,
.campus-header__placeholder {
	width: 74rpx;
	height: 74rpx;
}

.campus-header__back {
	border-radius: 50%;
	background: rgba(255, 255, 255, 0.72);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 60rpx;
	color: #1f1f1f;
}

.campus-header__title {
	font-size: 40rpx;
	font-weight: 700;
}

.campus-label {
	margin-top: 14rpx;
	padding: 22rpx 0;
	font-size: 24rpx;
	color: var(--parent-subtext);
}

.campus-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 22rpx 20rpx;
	margin-bottom: 16rpx;
}

.campus-item__left {
	display: flex;
	align-items: center;
	gap: 18rpx;
}

.campus-item__avatar {
	width: 68rpx;
	height: 68rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, #ffe213 0%, #ffc907 100%);
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	font-size: 28rpx;
	font-weight: 700;
}

.campus-item__name {
	font-size: 28rpx;
	line-height: 1.4;
}

.campus-item__check {
	font-size: 34rpx;
	font-weight: 700;
	color: var(--parent-brand-deep);
}
</style>
