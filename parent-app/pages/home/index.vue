<template>
	<view class="parent-page">
		<view class="parent-shell">
			<view class="parent-header" :style="{ paddingTop: `${nav.top}px` }">
				<view class="home-top-row" :style="{ minHeight: `${nav.height}px` }">
					<view v-if="canSwitchCampus" class="home-campus-trigger" @click="goCampusPage">
						<text class="home-campus-trigger__icon">切</text>
						<text class="home-campus-trigger__text">切换校区</text>
					</view>
				</view>
				<view class="parent-card home-campus-card">
					<view class="home-campus-brand">
						<view class="home-campus-logo">
							<image v-if="currentCampus.logoUrl" class="home-campus-logo__image" :src="currentCampus.logoUrl" mode="aspectFill"></image>
							<text v-else class="home-campus-logo__text">{{ campusLogoText }}</text>
						</view>
						<view class="home-campus-brand__content">
							<text class="home-campus-brand__name">{{ campusTitle }}</text>
							<text v-if="campusSubtitle" class="home-campus-brand__campus">{{ campusSubtitle }}</text>
						</view>
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
					<text class="home-notice-card__title">评估报告</text>
					<text class="home-notice-card__more">全部</text>
				</view>
				<view
					v-for="item in reportList"
					:key="item.id"
					class="home-notice-item"
					@click="handleReport(item)"
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
import { onShow } from '@dcloudio/uni-app'
import { listParentCampuses } from '@/common/parent-api'
import { getNavLayout } from '@/common/nav-layout'
import { applyParentCampusSummary, parentState, getCurrentCampus } from '@/common/parent-state'

const nav = getNavLayout()
const currentCampus = computed(() => getCurrentCampus())
const featureList = computed(() => parentState.featureList)
const reportList = computed(() => parentState.noticeList)
const canSwitchCampus = computed(() => parentState.isAuthenticated && parentState.campusList.length > 1)
const campusTitle = computed(() => {
	return buildCampusTitle(currentCampus.value)
})
const campusSubtitle = computed(() => {
	return buildCampusSubtitle(currentCampus.value)
})
const campusLogoText = computed(() => {
	const title = campusTitle.value || currentCampus.value?.shortName || '校'
	return `${title}`.slice(0, 1)
})

onShow(() => {
	refreshCampusSummary()
})

function goCampusPage() {
	uni.navigateTo({
		url: '/pages/select-campus/index'
	})
}

function handleFeature(item) {
	if (item?.key === 'attendance') {
		uni.navigateTo({
			url: '/pages/attendance-record/index'
		})
		return
	}
	if (item?.key === 'course') {
		uni.navigateTo({
			url: '/pages/course-enrollment/index'
		})
		return
	}
	uni.showToast({
		title: `${item.title}建设中`,
		icon: 'none'
	})
}

function handleReport(item) {
	uni.showToast({
		title: item.title,
		icon: 'none'
	})
}

async function refreshCampusSummary() {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	try {
		const summary = await listParentCampuses(parentState.authToken)
		applyParentCampusSummary(summary)
	} catch (error) {
		console.warn('refresh parent campuses failed', error)
	}
}

function buildCampusTitle(campus = {}) {
	const brandName = `${campus?.brandName || ''}`.trim()
	const simpleName = simplifyCampusName(campus?.name)
	return brandName || simpleName || '机构'
}

function buildCampusSubtitle(campus = {}) {
	const title = buildCampusTitle(campus)
	const fullName = `${campus?.name || ''}`.trim()
	const simpleName = simplifyCampusName(fullName)
	if (!fullName || fullName === title || simpleName === title) {
		return ''
	}
	return fullName
}

function simplifyCampusName(name = '') {
	return `${name || ''}`
		.replace(/\s*(总校区|控江校区|校区|分校|院区)\s*$/u, '')
		.trim()
}
</script>

<style scoped>
.home-top-row {
	display: flex;
	justify-content: flex-start;
	align-items: center;
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
	margin-top: 16rpx;
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
	overflow: hidden;
	box-shadow: 0 16rpx 34rpx rgba(29, 210, 119, 0.25);
	flex-shrink: 0;
}

.home-campus-logo__image {
	width: 100%;
	height: 100%;
	display: block;
}

.home-campus-logo__text {
	color: #ffffff;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.home-campus-brand__content {
	flex: 1;
	display: flex;
	flex-direction: column;
	justify-content: center;
	min-height: 88rpx;
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

.home-grid-card {
	padding: 32rpx 16rpx 32rpx;
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
	border-top: 1rpx solid var(--parent-divider);
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
