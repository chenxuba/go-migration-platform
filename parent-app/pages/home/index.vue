<template>
	<view class="parent-page">
		<view class="parent-shell">
			<view class="parent-header" :style="{ paddingTop: `${statusBarHeight + 20}px` }">
				<view class="home-top-row">
					<view class="home-campus-trigger" @click="goCampusPage">
						<text class="home-campus-trigger__icon">切</text>
						<text class="home-campus-trigger__text">切换校区</text>
					</view>
				</view>
				<view class="parent-card home-campus-card">
					<view class="home-campus-brand">
						<view class="home-campus-logo">校</view>
						<view class="home-campus-brand__content">
							<text class="home-campus-brand__name">{{ currentCampus.brandName }}</text>
							<text class="home-campus-brand__campus">{{ currentCampus.name }}</text>
						</view>
						<view class="home-campus-brand__mute">静</view>
					</view>
					<view v-if="currentStudent" class="home-student-bar">
						<text class="home-student-bar__label">当前关注学员</text>
						<text class="home-student-bar__value">{{ currentStudent.name }}</text>
					</view>
				</view>
			</view>

			<view class="parent-card home-grid-card">
				<view class="home-grid">
					<view
						v-for="item in featureList"
						:key="item.key"
						class="home-grid-item"
						@click="handleFeature(item)"
					>
						<view class="home-grid-item__badge" :style="{ background: item.accent }">
							{{ item.shortLabel }}
						</view>
						<text class="home-grid-item__text">{{ item.title }}</text>
					</view>
				</view>
			</view>

			<view class="parent-card home-notice-card">
				<view class="home-notice-card__header">
					<text class="home-notice-card__title">系统通知</text>
					<text class="home-notice-card__more">全部</text>
				</view>
				<view
					v-for="item in noticeList"
					:key="item.id"
					class="home-notice-item"
					@click="handleNotice(item)"
				>
					<view class="home-notice-item__dot"></view>
					<view class="home-notice-item__content">
						<text class="home-notice-item__title">{{ item.title }}</text>
						<text class="home-notice-item__desc">{{ item.summary }}</text>
					</view>
				</view>
			</view>
		</view>

	</view>
</template>

<script setup>
import { computed } from 'vue'
import { parentState, getCurrentCampus } from '@/common/parent-state'

const statusBarHeight = uni.getSystemInfoSync().statusBarHeight || 0

const currentCampus = computed(() => getCurrentCampus())
const currentStudent = computed(() => parentState.students.find(item => item.id === parentState.currentStudentId) || null)
const featureList = computed(() => parentState.featureList)
const noticeList = computed(() => parentState.noticeList)

function goCampusPage() {
	uni.navigateTo({
		url: '/pages/select-campus/index'
	})
}

function handleFeature(item) {
	uni.showToast({
		title: `${item.title}建设中`,
		icon: 'none'
	})
}

function handleNotice(item) {
	uni.showToast({
		title: item.title,
		icon: 'none'
	})
}
</script>

<style scoped>
.home-top-row {
	display: flex;
	justify-content: flex-start;
}

.home-campus-trigger {
	display: inline-flex;
	align-items: center;
	gap: 12rpx;
	padding: 12rpx 18rpx;
	border-radius: 999rpx;
	background: rgba(255, 255, 255, 0.62);
	border: 1rpx solid rgba(255, 255, 255, 0.8);
}

.home-campus-trigger__icon {
	width: 32rpx;
	height: 32rpx;
	border-radius: 12rpx;
	background: rgba(255, 214, 10, 0.18);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 18rpx;
	font-weight: 700;
	color: #7a5e00;
}

.home-campus-trigger__text {
	font-size: 26rpx;
	font-weight: 600;
}

.home-campus-card {
	margin-top: 24rpx;
	padding: 24rpx;
}

.home-campus-brand {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.home-campus-logo {
	width: 88rpx;
	height: 88rpx;
	border-radius: 28rpx;
	background: linear-gradient(135deg, #1dd277 0%, #24c05f 100%);
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	font-size: 36rpx;
	font-weight: 700;
	box-shadow: 0 16rpx 34rpx rgba(29, 210, 119, 0.25);
}

.home-campus-brand__content {
	flex: 1;
	display: flex;
	flex-direction: column;
}

.home-campus-brand__name {
	font-size: 36rpx;
	font-weight: 700;
	line-height: 1.3;
}

.home-campus-brand__campus {
	margin-top: 8rpx;
	font-size: 22rpx;
	line-height: 1.5;
	color: var(--parent-subtext);
}

.home-campus-brand__mute {
	width: 58rpx;
	height: 58rpx;
	border-radius: 50%;
	background: rgba(159, 159, 159, 0.2);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 22rpx;
	color: #8c8c8c;
}

.home-student-bar {
	margin-top: 26rpx;
	padding: 14rpx 18rpx;
	border-radius: 24rpx;
	background: rgba(255, 248, 226, 0.9);
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.home-student-bar__label {
	font-size: 22rpx;
	color: var(--parent-subtext);
}

.home-student-bar__value {
	font-size: 26rpx;
	font-weight: 700;
}

.home-grid-card {
	padding: 24rpx 16rpx 4rpx;
}

.home-grid {
	display: grid;
	grid-template-columns: repeat(4, 1fr);
	gap: 22rpx 8rpx;
}

.home-grid-item {
	display: flex;
	flex-direction: column;
	align-items: center;
}

.home-grid-item__badge {
	width: 78rpx;
	height: 78rpx;
	border-radius: 28rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	font-size: 28rpx;
	font-weight: 700;
	box-shadow: 0 14rpx 28rpx rgba(65, 65, 65, 0.08);
}

.home-grid-item__text {
	margin-top: 12rpx;
	font-size: 22rpx;
	text-align: center;
	line-height: 1.4;
}

.home-notice-card {
	margin-top: 26rpx;
	padding: 24rpx;
}

.home-notice-card__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-bottom: 10rpx;
}

.home-notice-card__title {
	font-size: 28rpx;
	font-weight: 700;
}

.home-notice-card__more {
	font-size: 22rpx;
	color: var(--parent-subtext);
}

.home-notice-item {
	display: flex;
	gap: 18rpx;
	padding: 24rpx 0;
	border-top: 1rpx solid rgba(234, 225, 207, 0.76);
}

.home-notice-item:first-of-type {
	margin-top: 16rpx;
}

.home-notice-item__dot {
	width: 16rpx;
	height: 16rpx;
	border-radius: 50%;
	background: var(--parent-brand-deep);
	margin-top: 14rpx;
	flex-shrink: 0;
}

.home-notice-item__content {
	display: flex;
	flex-direction: column;
}

.home-notice-item__title {
	font-size: 26rpx;
	font-weight: 600;
}

.home-notice-item__desc {
	margin-top: 8rpx;
	font-size: 22rpx;
	line-height: 1.5;
	color: var(--parent-subtext);
}
</style>
