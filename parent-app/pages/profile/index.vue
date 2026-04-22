<template>
	<view class="parent-page">
		<view class="parent-shell parent-shell--profile">
			<view class="parent-header profile-header" :style="{ paddingTop: `${nav.top}px` }">
				<view class="parent-nav-row profile-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
					<text class="profile-nav__title">我的</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>

				<view
					class="parent-card profile-hero-card"
					:class="{ 'profile-hero-card--link': isAuthenticated }"
					@click="openSettings"
				>
					<view class="profile-hero-card__identity">
						<view class="profile-hero-card__avatar">{{ avatarText }}</view>
						<view class="profile-hero-card__copy">
							<text class="profile-hero-card__title">{{ heroTitle }}</text>
							<text v-if="heroSubtitle" class="profile-hero-card__subtitle">{{ heroSubtitle }}</text>
						</view>
					</view>
					<text v-if="isAuthenticated" class="profile-hero-card__arrow">›</text>
				</view>
			</view>

			<view v-if="isAuthenticated" class="profile-block">
				<text class="profile-block__title">{{ studentBlockTitle }}</text>

				<template v-if="students.length">
					<view
						v-for="student in students"
						:key="student.id"
						class="parent-card profile-student-card"
					>
						<view class="profile-student-card__main">
							<view class="profile-student-card__avatar" :style="{ background: student.avatarColor }">
								<image v-if="student.avatarUrl" class="profile-student-card__avatar-image" :src="student.avatarUrl" mode="aspectFill"></image>
								<text v-else>{{ student.name.slice(0, 1) }}</text>
							</view>
							<view class="profile-student-card__copy">
								<view class="profile-student-card__headline">
									<text class="profile-student-card__name">{{ student.name }}</text>
								</view>
								<text class="profile-student-card__campus">{{ simplifyCampusName(student.campusName) }}</text>
								<view class="profile-student-card__meta">
									<text class="profile-student-card__tag profile-student-card__tag--warm">{{ student.relation }}</text>
									<text class="profile-student-card__tag profile-student-card__tag--blue">{{ student.classLabel }}</text>
								</view>
							</view>
						</view>
					</view>
				</template>

				<template v-else-if="pendingCount">
					<view class="parent-card profile-state-card">
						<text class="profile-state-card__title">发现 {{ pendingCount }} 位待绑定学员</text>
						<view class="profile-state-card__button" @click="openPendingStudents">去选择学员</view>
					</view>
				</template>

				<template v-else>
					<view class="parent-card profile-state-card profile-state-card--empty">
						<text class="profile-state-card__title">暂无已绑定学员</text>
					</view>
				</template>
			</view>

			<view class="profile-block profile-block--menu">
				<text class="profile-block__title">常用功能</text>

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
							<text class="profile-menu-item__title">{{ item.title }}</text>
						</view>
						<view class="profile-menu-item__right">
							<view v-if="item.showDot" class="profile-menu-item__dot"></view>
							<text class="profile-menu-item__arrow">›</text>
						</view>
					</view>
				</view>
			</view>

			<view v-if="!isAuthenticated" class="profile-auth-block">
				<!-- #ifdef MP-WEIXIN -->
				<button class="parent-primary-button profile-auth-block__button" :loading="authLoading" :disabled="authLoading" open-type="getPhoneNumber" @getphonenumber="handleWechatPhoneAuth">
					授权登录
				</button>
				<!-- #endif -->
				<!-- #ifndef MP-WEIXIN -->
				<view class="parent-primary-button profile-auth-block__button" @click="handleMockPhoneAuth">授权登录</view>
				<!-- #endif -->
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
import { onShow } from '@dcloudio/uni-app'
import BindSuccessDialog from '@/components/bind-success-dialog/bind-success-dialog.vue'
import { authorizeByWechatPhone } from '@/common/parent-auth'
import { listParentBoundStudents, listParentPendingStudents } from '@/common/parent-api'
import { getNavLayout } from '@/common/nav-layout'
import { DEFAULT_PHONE } from '@/common/mock-parent'
import {
	applyParentBoundStudentSummary,
	applyParentPendingStudentSummary,
	authorizeByPhone,
	dismissBindSuccess,
	parentState
} from '@/common/parent-state'

const nav = getNavLayout()

const authLoading = computed(() => parentState.authLoading)
const isAuthenticated = computed(() => parentState.isAuthenticated)
const students = computed(() => parentState.students)
const pendingCount = computed(() => parentState.pendingCandidates.length)
const bindSuccessVisible = computed(() => parentState.bindSuccessVisible)
const latestBindStudentName = computed(() => parentState.latestBindStudentName)
const maskedPhone = computed(() => parentState.profile.maskedPhone)
const avatarText = computed(() => (isAuthenticated.value ? parentState.profile.avatarText : '家'))
const heroTitle = computed(() => (isAuthenticated.value ? parentState.profile.nickname : '未登录'))
const heroSubtitle = computed(() => (isAuthenticated.value ? maskedPhone.value : ''))
const studentBlockTitle = computed(() => {
	if (students.value.length) {
		return '我的学员'
	}
	return '待绑定学员'
})
const displayMenus = computed(() => {
	const sourceList = isAuthenticated.value
		? parentState.profileMenus
		: parentState.profileMenus.filter(item => item.key === 'notice')

	return sourceList.map(item => ({
		...item,
		showDot: item.key === 'pending' && pendingCount.value > 0
	}))
})

let profileRefreshSerial = 0

onShow(() => {
	refreshProfileStudents()
})

function simplifyCampusName(name = '') {
	const text = `${name || ''}`.trim()
	if (!text) {
		return '-'
	}
	return text.replace(/\s*(总校区|控江校区)\s*$/u, '').trim() || text
}

function completeMockAuth() {
	authorizeByPhone(DEFAULT_PHONE)
	uni.navigateTo({
		url: '/pages/select-student/index'
	})
}

function handleMockPhoneAuth() {
	completeMockAuth()
}

function handleWechatPhoneAuth(event) {
	authorizeByWechatPhone(event, {
		postAuthPage: '/pages/profile/index'
	})
}

function openPendingStudents() {
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
			title: '暂无待绑定学员',
			icon: 'none'
		})
		return
	}

	uni.navigateTo({
		url: '/pages/select-student/index'
	})
}

function openSettings() {
	if (!isAuthenticated.value) {
		return
	}
	uni.navigateTo({
		url: '/pages/settings/index'
	})
}

function handleMenuClick(item) {
	if (item.key === 'pending') {
		openPendingStudents()
		return
	}

	uni.showToast({
		title: `${item.title}建设中`,
		icon: 'none'
	})
}

async function refreshProfileStudents() {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}

	const token = `${parentState.authToken || ''}`.trim()
	if (!token) {
		return
	}

	const requestSerial = ++profileRefreshSerial
	try {
		const [boundSummary, pendingSummary] = await Promise.all([
			listParentBoundStudents(token),
			listParentPendingStudents(token)
		])
		if (requestSerial !== profileRefreshSerial || token !== `${parentState.authToken || ''}`.trim()) {
			return
		}
		applyParentBoundStudentSummary(boundSummary)
		applyParentPendingStudentSummary(pendingSummary)
	} catch (error) {
		if (requestSerial !== profileRefreshSerial) {
			return
		}
		console.warn('refresh profile students failed', error)
	}
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
.parent-shell--profile {
	padding-bottom: 80rpx;
}

.profile-header {
	padding-bottom: 8rpx;
}

.profile-nav__title {
	flex: 1;
	text-align: center;
	font-size: 36rpx;
	font-weight: 700;
	line-height: 1;
}

.profile-hero-card {
	margin-top: 20rpx;
	padding: 26rpx 22rpx 22rpx;
	border-radius: 32rpx;
	background:
		radial-gradient(circle at top right, rgba(255, 226, 122, 0.14), transparent 26%),
		linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(255, 252, 245, 0.98) 100%);
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 18rpx;
}

.profile-hero-card--link {
	cursor: pointer;
}

.profile-hero-card__identity {
	display: flex;
	align-items: center;
	gap: 16rpx;
	flex: 1;
	min-width: 0;
}

.profile-hero-card__avatar {
	width: 88rpx;
	height: 88rpx;
	border-radius: 28rpx;
	background: linear-gradient(180deg, #fffef8 0%, #f3edde 100%);
	box-shadow:
		inset 0 0 0 1rpx rgba(255, 255, 255, 0.96),
		0 10rpx 24rpx rgba(157, 126, 70, 0.08);
	display: flex;
	align-items: center;
	justify-content: center;
	color: #b8af9d;
	font-size: 32rpx;
	font-weight: 700;
	flex-shrink: 0;
}

.profile-hero-card__copy {
	flex: 1;
	min-width: 0;
}

.profile-hero-card__title {
	display: block;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1.25;
	color: #1f1f1f;
}

.profile-hero-card__subtitle {
	display: block;
	margin-top: 10rpx;
	font-size: 23rpx;
	line-height: 1.5;
	color: #726958;
}

.profile-hero-card__arrow {
	color: #bbb29f;
	font-size: 30rpx;
	flex-shrink: 0;
}

.profile-block {
	margin-top: 30rpx;
}

.profile-auth-block {
	margin-top: 100rpx;
}

.profile-auth-block__button {
	width: 100%;
}

.profile-block--menu {
	margin-top: 34rpx;
}

.profile-block__title {
	display: block;
	padding: 0 4rpx;
	font-size: 30rpx;
	font-weight: 700;
	line-height: 1.3;
	color: #1f1f1f;
}

.profile-student-card {
	margin-top: 16rpx;
	padding: 20rpx 18rpx;
}

.profile-student-card__main {
	display: flex;
	align-items: flex-start;
	gap: 16rpx;
}

.profile-student-card__avatar {
	width: 72rpx;
	height: 72rpx;
	border-radius: 22rpx;
	overflow: hidden;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	font-size: 26rpx;
	font-weight: 700;
	box-shadow: 0 10rpx 22rpx rgba(71, 109, 178, 0.14);
	flex-shrink: 0;
	margin-top: 4rpx;
}

.profile-student-card__avatar-image {
	width: 100%;
	height: 100%;
	display: block;
}

.profile-student-card__copy {
	flex: 1;
	min-width: 0;
}

.profile-student-card__headline {
	display: flex;
	align-items: center;
	gap: 10rpx;
}

.profile-student-card__name {
	font-size: 29rpx;
	font-weight: 700;
	line-height: 1.35;
	color: #242424;
}

.profile-student-card__campus {
	display: block;
	margin-top: 10rpx;
	font-size: 22rpx;
	line-height: 1.5;
	color: var(--parent-subtext);
}

.profile-student-card__meta {
	display: flex;
	align-items: center;
	gap: 10rpx;
	margin-top: 14rpx;
	flex-wrap: wrap;
}

.profile-student-card__tag {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	height: 38rpx;
	padding: 0 14rpx;
	border-radius: 999rpx;
	font-size: 18rpx;
	font-weight: 600;
}

.profile-student-card__tag--warm {
	background: rgba(255, 222, 75, 0.16);
	color: #7d6200;
}

.profile-student-card__tag--blue {
	background: rgba(47, 134, 255, 0.12);
	color: var(--parent-blue);
}

.profile-state-card {
	margin-top: 16rpx;
	padding: 24rpx 20rpx;
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 18rpx;
}

.profile-state-card--empty {
	justify-content: center;
}

.profile-state-card__title {
	font-size: 27rpx;
	font-weight: 700;
	line-height: 1.4;
	color: #242424;
}

.profile-state-card__button {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	height: 68rpx;
	padding: 0 22rpx;
	border-radius: 999rpx;
	background: linear-gradient(135deg, #ffe60d 0%, #ffd10b 100%);
	color: #2b250c;
	font-size: 23rpx;
	font-weight: 700;
	flex-shrink: 0;
}

.profile-menu-card {
	margin-top: 16rpx;
	padding: 0 18rpx;
	overflow: hidden;
}

.profile-menu-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	min-height: 104rpx;
	padding: 0;
	border-top: 1rpx solid var(--parent-divider);
}

.profile-menu-item:first-child {
	border-top: none;
}

.profile-menu-item__left {
	display: flex;
	align-items: center;
	gap: 14rpx;
	min-width: 0;
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
	flex-shrink: 0;
}

.profile-menu-item__title {
	font-size: 26rpx;
	font-weight: 700;
	color: #232323;
}

.profile-menu-item__right {
	display: flex;
	align-items: center;
	gap: 12rpx;
	flex-shrink: 0;
}

.profile-menu-item__dot {
	width: 16rpx;
	height: 16rpx;
	border-radius: 50%;
	background: #ff5b43;
	box-shadow: 0 6rpx 14rpx rgba(255, 91, 67, 0.22);
}

.profile-menu-item__arrow {
	color: #bbb29f;
	font-size: 28rpx;
}
</style>
