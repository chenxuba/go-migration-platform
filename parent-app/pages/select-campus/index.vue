<template>
	<view class="parent-page">
		<view class="parent-shell">
			<view class="parent-header campus-header" :style="{ paddingTop: `${nav.top}px` }">
				<view class="parent-nav-row campus-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="campus-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="campus-back" @click="goBack">
							<view class="campus-back__icon"></view>
						</view>
					</view>
					<text class="campus-nav__title">选择学校</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>

				<view class="campus-hero">
					<text class="campus-hero__label">当前校区</text>
					<text class="campus-hero__desc">切换后，首页和课表会同步展示对应校区内容。</text>
				</view>
			</view>

			<view
				v-for="item in campusList"
				:key="item.id"
				class="parent-card campus-item"
				:class="{ 'campus-item--active': currentCampusId === item.id }"
				@click="selectCampus(item.id)"
			>
				<view class="campus-item__main">
					<view class="campus-item__avatar" :class="{ 'campus-item__avatar--active': currentCampusId === item.id }">
						{{ item.shortName }}
					</view>
					<view class="campus-item__content">
						<view class="campus-item__title-row">
							<text class="campus-item__name">{{ item.name }}</text>
						</view>
						<text class="campus-item__brand">{{ item.brandName }}</text>
						<view class="campus-item__meta">
							<text
								class="campus-item__meta-chip"
								:class="{ 'campus-item__meta-chip--active': currentCampusId === item.id }"
							>
								{{ currentCampusId === item.id ? '当前校区' : '点击切换' }}
							</text>
						</view>
					</view>
				</view>

				<view
					class="campus-item__status"
					:class="{ 'campus-item__status--active': currentCampusId === item.id }"
				>
					<view
						class="campus-item__status-dot"
						:class="{ 'campus-item__status-dot--active': currentCampusId === item.id }"
					></view>
					<text class="campus-item__status-text">{{ currentCampusId === item.id ? '已选中' : '未选中' }}</text>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { computed } from 'vue'
import { getNavLayout } from '@/common/nav-layout'
import { parentState, switchCurrentCampus } from '@/common/parent-state'

const nav = getNavLayout()
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
.campus-nav__side {
	display: flex;
	align-items: center;
}

.campus-nav__title {
	flex: 1;
	text-align: center;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.campus-back {
	width: 64rpx;
	height: 64rpx;
	border-radius: 22rpx;
	background: rgba(255, 255, 255, 0.76);
	border: 1rpx solid rgba(255, 255, 255, 0.92);
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 12rpx 28rpx rgba(162, 130, 71, 0.08);
}

.campus-back__icon {
	width: 16rpx;
	height: 16rpx;
	border-left: 4rpx solid #433e33;
	border-bottom: 4rpx solid #433e33;
	transform: translateX(4rpx) rotate(45deg);
}

.campus-hero {
	margin-top: 22rpx;
	padding: 0 4rpx;
}

.campus-hero__label {
	display: block;
	font-size: 22rpx;
	font-weight: 700;
	letter-spacing: 2rpx;
	color: #8f8570;
}

.campus-hero__desc {
	display: block;
	margin-top: 12rpx;
	font-size: 24rpx;
	line-height: 1.6;
	color: var(--parent-subtext);
}

.campus-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 18rpx;
	padding: 22rpx 20rpx;
	margin-bottom: 16rpx;
	border: 2rpx solid transparent;
	transition: all 0.2s ease;
}

.campus-item--active {
	background: linear-gradient(135deg, rgba(255, 246, 206, 0.92) 0%, rgba(255, 255, 255, 0.96) 100%);
	border-color: rgba(255, 214, 10, 0.92);
	box-shadow: 0 18rpx 36rpx rgba(255, 214, 10, 0.12);
}

.campus-item__main {
	flex: 1;
	display: flex;
	align-items: center;
	gap: 18rpx;
	min-width: 0;
}

.campus-item__avatar {
	width: 76rpx;
	height: 76rpx;
	border-radius: 24rpx;
	background: linear-gradient(135deg, rgba(255, 226, 19, 0.4) 0%, rgba(255, 201, 7, 0.82) 100%);
	display: flex;
	align-items: center;
	justify-content: center;
	color: #8c6500;
	font-size: 28rpx;
	font-weight: 700;
	flex-shrink: 0;
}

.campus-item__avatar--active {
	background: linear-gradient(135deg, #ffe213 0%, #ffc907 100%);
	color: #ffffff;
	box-shadow: 0 12rpx 24rpx rgba(255, 214, 10, 0.26);
}

.campus-item__content {
	flex: 1;
	min-width: 0;
}

.campus-item__title-row {
	display: flex;
	align-items: center;
}

.campus-item__name {
	font-size: 28rpx;
	font-weight: 700;
	line-height: 1.5;
	color: #242424;
}

.campus-item__brand {
	display: block;
	margin-top: 8rpx;
	font-size: 22rpx;
	line-height: 1.5;
	color: var(--parent-subtext);
}

.campus-item__meta {
	margin-top: 14rpx;
}

.campus-item__meta-chip {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	height: 42rpx;
	padding: 0 16rpx;
	border-radius: 999rpx;
	background: rgba(245, 241, 231, 0.96);
	color: #8c856f;
	font-size: 20rpx;
	font-weight: 600;
}

.campus-item__meta-chip--active {
	background: rgba(255, 222, 75, 0.22);
	color: #7d6200;
}

.campus-item__status {
	display: flex;
	align-items: center;
	gap: 10rpx;
	padding: 10rpx 14rpx;
	border-radius: 999rpx;
	background: rgba(247, 244, 237, 0.96);
	flex-shrink: 0;
}

.campus-item__status--active {
	background: rgba(255, 240, 178, 0.78);
}

.campus-item__status-dot {
	width: 14rpx;
	height: 14rpx;
	border-radius: 50%;
	background: rgba(184, 178, 164, 0.88);
}

.campus-item__status-dot--active {
	background: var(--parent-brand-deep);
	box-shadow: 0 0 0 6rpx rgba(255, 214, 10, 0.18);
}

.campus-item__status-text {
	font-size: 20rpx;
	font-weight: 600;
	color: #766d5d;
}
</style>
