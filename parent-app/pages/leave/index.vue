<template>
	<view class="parent-page leave-page">
		<view class="leave-nav-fixed" :style="{ paddingTop: `${nav.top}px` }">
			<view class="leave-nav-fixed__inner">
				<view class="parent-nav-row leave-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="leave-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="leave-back" @click="goBack">
							<view class="leave-back__icon"></view>
						</view>
					</view>
					<text class="leave-nav__title">请假</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>
			</view>
		</view>

		<view class="parent-shell" :style="{ paddingTop: `${nav.top + nav.height + 18}px` }">
			<view class="parent-header leave-header">
				<view v-if="isAuthenticated && displayStudents.length" class="parent-card leave-student-card">
					<scroll-view class="leave-student-scroll" :scroll-x="!shouldCenterStudents" show-scrollbar="false" enable-flex>
						<view
							class="leave-student-list"
							:class="{ 'leave-student-list--center': shouldCenterStudents }"
						>
							<view
								v-for="student in displayStudents"
								:key="student.id"
								class="leave-student-item"
								:class="{
									'leave-student-item--active': selectedStudentId === student.id,
									'leave-student-item--center': shouldCenterStudents
								}"
								@click="handleStudentSelect(student.id)"
							>
								<text class="leave-student-item__name">{{ student.name }}</text>
								<view class="leave-student-item__line"></view>
							</view>
						</view>
					</scroll-view>
				</view>
			</view>

			<view class="parent-card leave-list-card">
				<template v-if="!isAuthenticated">
					<view class="parent-empty-card leave-empty-card">
						<view class="parent-empty-badge">假</view>
						<text class="parent-empty-title">登录后即可查看请假记录</text>
						<text class="parent-empty-desc">手机号授权后，系统会自动同步当前学员的请假记录。</text>
						<!-- #ifdef MP-WEIXIN -->
						<button class="parent-primary-button leave-auth-button" :loading="authLoading" :disabled="authLoading" open-type="getPhoneNumber" @getphonenumber="handleWechatPhoneAuth">
							授权登录
						</button>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<view class="parent-primary-button leave-auth-button" @click="handleMockPhoneAuth">授权登录</view>
						<!-- #endif -->
					</view>
				</template>

				<template v-else-if="pageLoading && !leaveList.length">
					<view class="parent-empty-card leave-empty-card">
						<view class="parent-empty-badge">载</view>
						<text class="parent-empty-title">正在加载请假记录</text>
						<text class="parent-empty-desc">稍等一下，系统正在同步最新请假数据。</text>
					</view>
				</template>

				<template v-else-if="leaveList.length">
					<view
						v-for="item in leaveList"
						:key="item.id"
						class="leave-item"
					>
						<view class="leave-item__avatar">
							<image
								v-if="item.studentAvatarUrl"
								class="leave-item__avatar-image"
								:src="item.studentAvatarUrl"
								mode="aspectFill"
							></image>
							<text v-else>{{ item.avatarText }}</text>
						</view>

						<view class="leave-item__content">
							<view class="leave-item__header">
								<text class="leave-item__title">{{ item.title }}</text>
								<text class="leave-item__status" :class="leaveStatusClass(item.statusText)">{{ item.statusText }}</text>
							</view>

							<view class="leave-item__meta">
								<view class="leave-item__row">
									<text class="leave-item__label">请假类型：</text>
									<text class="leave-item__value">{{ item.leaveTypeText }}</text>
								</view>
								<view class="leave-item__row">
									<text class="leave-item__label">开始时间：</text>
									<text class="leave-item__value">{{ item.startTime }}</text>
								</view>
								<view class="leave-item__row">
									<text class="leave-item__label">结束时间：</text>
									<text class="leave-item__value">{{ item.endTime }}</text>
								</view>
							</view>
						</view>
					</view>

					<view v-if="moreLoading" class="leave-loading-text">正在加载更多...</view>
					<view v-else-if="!hasMore && leaveList.length >= pageSize" class="leave-loading-text">没有更多记录了</view>
				</template>

				<template v-else>
					<view class="parent-empty-card leave-empty-card">
						<view class="parent-empty-badge">空</view>
						<text class="parent-empty-title">{{ displayStudents.length ? '当前学员还没有请假记录' : '暂未绑定学员' }}</text>
						<text class="parent-empty-desc">{{ displayStudents.length ? '发起请假后，记录会自动同步到这里。' : '请先完成学员关注，再查看请假记录。' }}</text>
					</view>
				</template>
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
import { computed, ref } from 'vue'
import { onLoad, onReachBottom, onShow } from '@dcloudio/uni-app'
import BindSuccessDialog from '@/components/bind-success-dialog/bind-success-dialog.vue'
import { authorizeByWechatPhone } from '@/common/parent-auth'
import { listParentLeaves } from '@/common/parent-api'
import { getNavLayout } from '@/common/nav-layout'
import { DEFAULT_PHONE } from '@/common/mock-parent'
import {
	authorizeByPhone,
	dismissBindSuccess,
	parentState,
	setPostAuthPage
} from '@/common/parent-state'

const nav = getNavLayout()
const summaryStudents = ref([])
const leaveItems = ref([])
const hasLoadedSummary = ref(false)
const selectedStudentId = ref('')
const pageLoading = ref(false)
const moreLoading = ref(false)
const hasMore = ref(false)
const routeStudentId = ref('')
const authLoading = computed(() => parentState.authLoading)
const isAuthenticated = computed(() => parentState.isAuthenticated)
const bindSuccessVisible = computed(() => parentState.bindSuccessVisible)
const latestBindStudentName = computed(() => parentState.latestBindStudentName)
const pageIndex = ref(1)
const pageSize = 20

let leaveRequestSerial = 0

const displayStudents = computed(() => {
	const backendStudents = normalizeStudents(summaryStudents.value || [])
	if (hasLoadedSummary.value) {
		return backendStudents
	}
	return normalizeStudents(parentState.students || [])
})
const shouldCenterStudents = computed(() => displayStudents.value.length > 0 && displayStudents.value.length <= 3)
const leaveList = computed(() => normalizeLeaveList(leaveItems.value || []))

onLoad(query => {
	routeStudentId.value = `${query?.studentId || ''}`.trim()
})

onShow(() => {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		hasLoadedSummary.value = false
		summaryStudents.value = []
		resetLeavePagination()
		return
	}
	const cachedStudents = normalizeStudents(summaryStudents.value || [])
	const sessionStudents = normalizeStudents(parentState.students || [])
	const needSyncStudents = sessionStudents.length !== cachedStudents.length
		|| sessionStudents.some((item, index) => item.id !== cachedStudents[index]?.id)
	if (needSyncStudents || parentState.bindSuccessVisible) {
		hasLoadedSummary.value = false
	}
	ensureSelectedStudent()
	refreshLeaves({
		reset: true
	})
})

onReachBottom(() => {
	loadMoreLeaves()
})

function normalizeStudents(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		id: `${item?.id || item?.studentId || item?.rawId || `student-${index + 1}`}`.trim(),
		name: `${item?.name || item?.studentName || '-'}`.trim() || '-'
	})).filter(item => item.id)
}

function normalizeLeaveList(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => {
		const studentName = `${item?.studentName || '-'}`.trim() || '-'
		return {
			id: `${item?.id || `leave-${index + 1}`}`.trim(),
			title: studentName === '-' ? '请假记录' : studentName,
			studentName,
			studentAvatarUrl: `${item?.studentAvatarUrl || ''}`.trim(),
			avatarText: studentName.slice(0, 1) || '学',
			leaveTypeText: `${item?.leaveTypeText || '请假'}`.trim() || '请假',
			statusText: `${item?.statusText || '-'}`.trim() || '-',
			startTime: `${item?.startTime || '-'}`.trim() || '-',
			endTime: `${item?.endTime || '-'}`.trim() || '-'
		}
	})
}

function ensureSelectedStudent() {
	const students = displayStudents.value
	if (!students.length) {
		selectedStudentId.value = ''
		return
	}
	if (routeStudentId.value) {
		const matchedByRoute = students.find(item => item.id === routeStudentId.value)
		if (matchedByRoute) {
			selectedStudentId.value = matchedByRoute.id
			routeStudentId.value = ''
			return
		}
	}
	const hasCurrent = students.some(item => item.id === selectedStudentId.value)
	if (hasCurrent) {
		return
	}
	const preferred = `${parentState.currentStudentId || ''}`.trim()
	const matched = students.find(item => item.id === preferred)
	selectedStudentId.value = matched?.id || students[0].id
}

function resetLeavePagination() {
	leaveItems.value = []
	pageIndex.value = 1
	hasMore.value = false
	pageLoading.value = false
	moreLoading.value = false
}

async function refreshLeaves(options = {}) {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	const reset = options?.reset !== false

	ensureSelectedStudent()
	const token = `${parentState.authToken || ''}`.trim()
	const targetStudentId = `${selectedStudentId.value || ''}`.trim()
	if (!token || !targetStudentId) {
		summaryStudents.value = displayStudents.value
		resetLeavePagination()
		return
	}

	const requestPageIndex = reset ? 1 : pageIndex.value + 1
	if (reset) {
		pageLoading.value = true
	} else {
		moreLoading.value = true
	}
	const requestSerial = ++leaveRequestSerial
	try {
		const result = await listParentLeaves(token, {
			studentId: targetStudentId,
			pageIndex: requestPageIndex,
			pageSize
		})
		if (requestSerial !== leaveRequestSerial || token !== `${parentState.authToken || ''}`.trim()) {
			return
		}
		hasLoadedSummary.value = true
		summaryStudents.value = result?.students || []
		const nextStudents = normalizeStudents(result?.students || [])
		if (nextStudents.length && !nextStudents.some(item => item.id === selectedStudentId.value)) {
			selectedStudentId.value = nextStudents[0].id
		}
		const nextItems = result?.items || []
		leaveItems.value = reset ? nextItems : mergePagedLeaveItems(leaveItems.value, nextItems)
		pageIndex.value = Number(result?.pageIndex || requestPageIndex)
		hasMore.value = !!result?.hasMore
	} catch (error) {
		if (requestSerial !== leaveRequestSerial) {
			return
		}
		console.warn('load parent leaves failed', error)
		if (reset) {
			uni.showToast({
				title: `${error?.message || '加载失败'}`.slice(0, 24),
				icon: 'none'
			})
		}
	} finally {
		if (requestSerial === leaveRequestSerial && reset) {
			pageLoading.value = false
		}
		if (requestSerial === leaveRequestSerial && !reset) {
			moreLoading.value = false
		}
	}
}

function handleStudentSelect(studentId) {
	if (!studentId || studentId === selectedStudentId.value || pageLoading.value || moreLoading.value) {
		return
	}
	selectedStudentId.value = studentId
	resetLeavePagination()
	refreshLeaves({
		reset: true
	})
	uni.pageScrollTo({
		scrollTop: 0,
		duration: 180
	})
}

function loadMoreLeaves() {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	if (!hasMore.value || pageLoading.value || moreLoading.value) {
		return
	}
	refreshLeaves({
		reset: false
	})
}

function buildCurrentPageURL() {
	const studentID = `${selectedStudentId.value || routeStudentId.value || ''}`.trim()
	if (!studentID) {
		return '/pages/leave/index'
	}
	return `/pages/leave/index?studentId=${encodeURIComponent(studentID)}`
}

function handleMockPhoneAuth() {
	setPostAuthPage(buildCurrentPageURL())
	authorizeByPhone(DEFAULT_PHONE)
	uni.navigateTo({
		url: '/pages/select-student/index'
	})
}

function handleWechatPhoneAuth(event) {
	authorizeByWechatPhone(event, {
		postAuthPage: buildCurrentPageURL()
	})
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

function leaveStatusClass(statusText = '') {
	const text = `${statusText || ''}`.trim()
	if (text === '待处理') {
		return 'leave-item__status--pending'
	}
	if (text === '已通过') {
		return 'leave-item__status--approved'
	}
	if (text === '已拒绝') {
		return 'leave-item__status--rejected'
	}
	if (text === '已撤销') {
		return 'leave-item__status--revoked'
	}
	return ''
}

function mergePagedLeaveItems(currentItems = [], nextItems = []) {
	const list = Array.isArray(currentItems) ? [...currentItems] : []
	const seen = new Set(list.map(item => `${item?.id || ''}`.trim()).filter(Boolean))
	;(Array.isArray(nextItems) ? nextItems : []).forEach(item => {
		const key = `${item?.id || ''}`.trim()
		if (key && seen.has(key)) {
			return
		}
		if (key) {
			seen.add(key)
		}
		list.push(item)
	})
	return list
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
.leave-page {
	min-height: 100vh;
	background:
		radial-gradient(circle at 12% 6%, rgba(255, 221, 169, 0.2), transparent 22%),
		linear-gradient(180deg, #fff7eb 0%, #fff8ef 42%, #fff9f2 100%);
}

.leave-header {
	padding-bottom: 18rpx;
}

.leave-nav-fixed {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	z-index: 40;
	background:
		linear-gradient(180deg, rgba(255, 247, 235, 0.98) 0%, rgba(255, 248, 239, 0.96) 72%, rgba(255, 249, 242, 0.9) 100%);
	backdrop-filter: blur(12rpx);
}

.leave-nav-fixed__inner {
	padding: 0 24rpx 10rpx;
}

.leave-nav__side {
	display: flex;
	align-items: center;
}

.leave-nav__title {
	flex: 1;
	text-align: center;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.leave-back {
	width: 64rpx;
	height: 64rpx;
	border-radius: 22rpx;
	background: rgba(255, 255, 255, 0.78);
	border: 1rpx solid rgba(255, 255, 255, 0.92);
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 12rpx 28rpx rgba(162, 130, 71, 0.08);
}

.leave-back__icon {
	width: 16rpx;
	height: 16rpx;
	border-left: 4rpx solid #433e33;
	border-bottom: 4rpx solid #433e33;
	transform: translateX(4rpx) rotate(45deg);
}

.leave-student-card {
	padding: 8rpx 14rpx 4rpx;
}

.leave-student-scroll {
	white-space: nowrap;
}

.leave-student-list {
	display: inline-flex;
	align-items: stretch;
	min-width: 100%;
}

.leave-student-list--center {
	justify-content: center;
	gap: 12rpx;
}

.leave-student-item {
	position: relative;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	min-width: 156rpx;
	padding: 16rpx 14rpx 14rpx;
}

.leave-student-item--center {
	flex: 0 1 180rpx;
	min-width: 0;
}

.leave-student-item__name {
	font-size: 30rpx;
	font-weight: 600;
	line-height: 1.24;
	color: #8f8a80;
}

.leave-student-item__line {
	width: 40rpx;
	height: 8rpx;
	margin-top: 14rpx;
	border-radius: 999rpx;
	background: transparent;
}

.leave-student-item--active .leave-student-item__name {
	color: #222222;
	font-weight: 700;
}

.leave-student-item--active .leave-student-item__line {
	background: linear-gradient(135deg, #ffe40f 0%, #ffcf0b 100%);
	box-shadow: 0 8rpx 18rpx rgba(255, 214, 10, 0.28);
}

.leave-list-card {
	padding: 22rpx 20rpx 26rpx;
}

.leave-empty-card {
	padding-top: 54rpx;
	padding-bottom: 54rpx;
}

.leave-auth-button {
	width: 100%;
	margin-top: 20rpx;
}

.leave-item {
	margin-top: 16rpx;
	padding: 24rpx 22rpx;
	border-radius: 24rpx;
	background: rgba(255, 255, 255, 0.92);
	border: 1rpx solid var(--parent-card-border);
	display: flex;
	align-items: flex-start;
	gap: 18rpx;
}

.leave-item__avatar {
	width: 76rpx;
	height: 76rpx;
	border-radius: 50%;
	background: linear-gradient(180deg, #e8f1ff 0%, #d7e7ff 100%);
	display: flex;
	align-items: center;
	justify-content: center;
	overflow: hidden;
	flex-shrink: 0;
	color: #7e9fda;
	font-size: 28rpx;
	font-weight: 700;
}

.leave-item__avatar-image {
	width: 100%;
	height: 100%;
	display: block;
}

.leave-item__content {
	flex: 1;
	min-width: 0;
}

.leave-item__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 16rpx;
}

.leave-item__title {
	flex: 1;
	min-width: 0;
	font-size: 30rpx;
	font-weight: 700;
	line-height: 1.4;
	color: #2f2c26;
}

.leave-item__status {
	flex-shrink: 0;
	padding: 10rpx 18rpx;
	border-radius: 999rpx;
	font-size: 22rpx;
	font-weight: 700;
	line-height: 1;
	background: rgba(236, 239, 242, 0.98);
	color: #7f8790;
}

.leave-item__status--pending {
	background: rgba(255, 245, 212, 0.98);
	color: #b27a00;
}

.leave-item__status--approved {
	background: rgba(227, 245, 232, 0.96);
	color: #26985b;
}

.leave-item__status--rejected {
	background: rgba(255, 231, 225, 0.96);
	color: #dd5b4f;
}

.leave-item__status--revoked {
	background: rgba(255, 247, 223, 0.96);
	color: #c79a1c;
}

.leave-item__meta {
	margin-top: 14rpx;
	display: flex;
	flex-direction: column;
	gap: 10rpx;
}

.leave-item__row {
	display: flex;
	align-items: flex-start;
	gap: 0;
}

.leave-item__label {
	flex-shrink: 0;
	font-size: 24rpx;
	line-height: 1.6;
	color: #989183;
}

.leave-item__value {
	flex: 1;
	font-size: 24rpx;
	line-height: 1.6;
	color: #45403a;
}

.leave-loading-text {
	margin-top: 18rpx;
	text-align: center;
	font-size: 22rpx;
	color: var(--parent-subtext);
}
</style>
