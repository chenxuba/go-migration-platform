<template>
	<view class="parent-page">
		<view class="parent-shell">
			<view class="parent-header" :style="{ paddingTop: `${nav.top}px` }">
				<view class="profile-top-row" :style="{ minHeight: `${nav.height}px` }">
					<view class="profile-hero">
						<view class="profile-hero__avatar">{{ avatarText }}</view>
						<view class="profile-hero__content">
							<text class="profile-hero__title">{{ profileName }}</text>
							<text v-if="isAuthenticated" class="profile-hero__phone">{{ maskedPhone }}</text>
						</view>
					</view>
					<view class="profile-hero__spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>
			</view>

			<view class="profile-section">
				<text class="profile-section__title">我的学员</text>
				<template v-if="students.length">
					<view
						v-for="student in students"
						:key="student.id"
						class="parent-card profile-student-card"
						:class="{ 'profile-student-card--active': currentStudentId === student.id }"
						@click="handleStudentSwitch(student.id)"
					>
						<view class="profile-student-card__avatar" :style="{ background: student.avatarColor }">
							{{ student.name.slice(0, 1) }}
						</view>
						<view class="profile-student-card__content">
							<view class="profile-student-card__header">
								<text class="profile-student-card__name">{{ student.name }}</text>
								<text class="profile-student-card__balance">余额：￥{{ student.balance }}</text>
							</view>
							<text class="profile-student-card__campus">{{ student.campusName }}</text>
							<view class="profile-student-card__relation">
								<text class="profile-student-card__relation-badge">{{ student.relation }}</text>
								<text class="profile-student-card__tag">{{ student.classLabel }}</text>
							</view>
						</view>
					</view>
				</template>
				<template v-else>
					<view class="parent-card parent-empty-card profile-empty-card">
						<view class="parent-empty-badge">家</view>
						<text class="parent-empty-title">暂无相关家庭，快去添加学员吧</text>
						<text class="parent-empty-desc">手机号授权后会自动反查孩子并进入关注流程。</text>
						<!-- #ifdef MP-WEIXIN -->
						<button class="parent-primary-button profile-auth-button" open-type="getPhoneNumber" @getphonenumber="handleWechatPhoneAuth">
							授权登录
						</button>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<view class="parent-primary-button profile-auth-button" @click="handleMockPhoneAuth">授权登录</view>
						<!-- #endif -->
					</view>
				</template>
			</view>

			<view class="parent-card profile-menu-card">
				<view
					v-for="item in displayMenus"
					:key="item.key"
					class="profile-menu-item"
					@click="handleMenuClick(item)"
				>
					<view class="profile-menu-item__left">
						<view class="profile-menu-item__badge" :style="{ background: item.accent }">
							{{ item.shortLabel }}
						</view>
						<text class="profile-menu-item__title">
							{{ item.title }}
							<text v-if="item.key === 'pending' && pendingCount" class="profile-menu-item__count">（{{ pendingCount }}）</text>
						</text>
					</view>
					<text class="profile-menu-item__arrow">›</text>
				</view>
			</view>
		</view>

		<bind-success-dialog
			:show="bindSuccessVisible"
			:student-name="latestBindStudentName"
			@close="closeBindSuccess"
			@invite="inviteFamily"
		/>
	</view>
</template>

<script setup>
import { computed } from 'vue'
import BindSuccessDialog from '@/components/bind-success-dialog/bind-success-dialog.vue'
import { getNavLayout } from '@/common/nav-layout'
import { DEFAULT_PHONE } from '@/common/mock-parent'
import {
	authorizeByPhone,
	dismissBindSuccess,
	parentState,
	setPostAuthPage,
	switchCurrentStudent
} from '@/common/parent-state'

const nav = getNavLayout()
const isAuthenticated = computed(() => parentState.isAuthenticated)
const students = computed(() => parentState.students)
const currentStudentId = computed(() => parentState.currentStudentId)
const pendingCount = computed(() => parentState.pendingCandidates.length)
const bindSuccessVisible = computed(() => parentState.bindSuccessVisible)
const latestBindStudentName = computed(() => parentState.latestBindStudentName)
const displayMenus = computed(() => {
	if (isAuthenticated.value) {
		return parentState.profileMenus
	}
	return parentState.profileMenus.filter(item => item.key === 'notice')
})

const profileName = computed(() => (isAuthenticated.value ? `Hi，${parentState.profile.nickname}` : '请登录'))
const maskedPhone = computed(() => parentState.profile.maskedPhone)
const avatarText = computed(() => (isAuthenticated.value ? parentState.profile.avatarText : '家'))

function completeMockAuth() {
	setPostAuthPage('/pages/profile/index')
	authorizeByPhone(DEFAULT_PHONE)
	uni.navigateTo({
		url: '/pages/select-student/index'
	})
}

function handleMockPhoneAuth() {
	completeMockAuth()
}

function handleWechatPhoneAuth(event) {
	const detail = event?.detail || {}
	if (detail.errMsg && !detail.errMsg.includes('ok')) {
		uni.showToast({
			title: '你已取消手机号授权',
			icon: 'none'
		})
		return
	}
	// 真实接入时，这里应把 detail.code 传给后端换取手机号并建立登录态。
	completeMockAuth()
}

function handleStudentSwitch(studentId) {
	switchCurrentStudent(studentId)
}

function handleMenuClick(item) {
	if (item.key === 'pending') {
		if (!isAuthenticated.value) {
			// #ifdef MP-WEIXIN
			uni.showToast({
				title: '请先点击授权登录',
				icon: 'none'
			})
			// #endif
			// #ifndef MP-WEIXIN
			handleMockPhoneAuth()
			// #endif
			return
		}
		if (!pendingCount.value) {
			uni.showToast({
				title: '暂无待关注学员',
				icon: 'none'
			})
			return
		}
		uni.navigateTo({
			url: '/pages/select-student/index'
		})
		return
	}
	uni.showToast({
		title: `${item.title}建设中`,
		icon: 'none'
	})
}

function closeBindSuccess() {
	dismissBindSuccess()
}

function inviteFamily() {
	dismissBindSuccess()
	uni.showToast({
		title: '邀请链接稍后接入',
		icon: 'none'
	})
}
</script>

<style scoped>
.profile-top-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.profile-hero {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.profile-hero__avatar {
	width: 84rpx;
	height: 84rpx;
	border-radius: 50%;
	background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(241, 237, 228, 0.94));
	border: 2rpx solid rgba(255, 255, 255, 0.88);
	display: flex;
	align-items: center;
	justify-content: center;
	color: #c0b9ad;
	font-size: 30rpx;
	font-weight: 700;
}

.profile-hero__content {
	display: flex;
	flex-direction: column;
}

.profile-hero__title {
	font-size: 44rpx;
	font-weight: 700;
	line-height: 1.15;
}

.profile-hero__phone {
	margin-top: 10rpx;
	font-size: 24rpx;
	color: #555555;
}

.profile-section {
	margin-top: 18rpx;
}

.profile-section__title {
	display: block;
	font-size: 38rpx;
	font-weight: 700;
}

.profile-student-card {
	display: flex;
	gap: 16rpx;
	margin-top: 16rpx;
	padding: 18rpx;
	border: 2rpx solid transparent;
}

.profile-student-card--active {
	border-color: rgba(255, 214, 10, 0.74);
}

.profile-student-card__avatar {
	width: 86rpx;
	height: 86rpx;
	border-radius: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	font-size: 30rpx;
	font-weight: 700;
	flex-shrink: 0;
}

.profile-student-card__content {
	flex: 1;
}

.profile-student-card__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 12rpx;
}

.profile-student-card__name {
	font-size: 30rpx;
	font-weight: 700;
}

.profile-student-card__balance {
	font-size: 22rpx;
	font-weight: 600;
	color: #565656;
}

.profile-student-card__campus {
	display: block;
	margin-top: 8rpx;
	font-size: 22rpx;
	line-height: 1.45;
	color: var(--parent-subtext);
}

.profile-student-card__relation {
	display: flex;
	align-items: center;
	gap: 10rpx;
	margin-top: 12rpx;
}

.profile-student-card__relation-badge,
.profile-student-card__tag {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	padding: 6rpx 12rpx;
	border-radius: 999rpx;
	font-size: 18rpx;
}

.profile-student-card__relation-badge {
	background: rgba(255, 214, 10, 0.22);
	color: #7d6200;
}

.profile-student-card__tag {
	background: rgba(47, 134, 255, 0.12);
	color: #2f86ff;
}

.profile-empty-card {
	margin-top: 16rpx;
}

.profile-auth-button {
	width: 100%;
	margin-top: 20rpx;
}

.profile-menu-card {
	margin-top: 22rpx;
	padding: 0 20rpx;
}

.profile-menu-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding: 20rpx 0;
	border-top: 1rpx solid rgba(234, 225, 207, 0.76);
}

.profile-menu-item:first-child {
	border-top: none;
}

.profile-menu-item__left {
	display: flex;
	align-items: center;
	gap: 14rpx;
}

.profile-menu-item__badge {
	width: 48rpx;
	height: 48rpx;
	border-radius: 16rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	font-size: 20rpx;
	font-weight: 700;
}

.profile-menu-item__title {
	font-size: 26rpx;
	font-weight: 600;
}

.profile-menu-item__count {
	color: var(--parent-blue);
}

.profile-menu-item__arrow {
	font-size: 32rpx;
	color: #c7c1b5;
}
</style>
