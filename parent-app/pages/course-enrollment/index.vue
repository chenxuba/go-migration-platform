<template>
	<view class="parent-page course-page">
		<view class="course-nav-fixed" :style="{ paddingTop: `${nav.top}px` }">
			<view class="course-nav-fixed__inner">
				<view class="parent-nav-row course-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="course-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="course-back" @click="goBack">
							<view class="course-back__icon"></view>
						</view>
					</view>
					<text class="course-nav__title">报读课程</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>
			</view>
		</view>

		<view class="parent-shell" :style="{ paddingTop: `${nav.top + nav.height + 18}px` }">
			<view class="parent-header course-header">
				<view v-if="isAuthenticated && displayStudents.length" class="parent-card course-student-card">
					<scroll-view class="course-student-scroll" :scroll-x="!shouldCenterStudents" show-scrollbar="false" enable-flex>
						<view
							class="course-student-list"
							:class="{ 'course-student-list--center': shouldCenterStudents }"
						>
							<view
								v-for="student in displayStudents"
								:key="student.id"
								class="course-student-item"
								:class="{
									'course-student-item--active': selectedStudentId === student.id,
									'course-student-item--center': shouldCenterStudents
								}"
								@click="handleStudentSelect(student.id)"
							>
								<text class="course-student-item__name">{{ student.name }}</text>
								<view class="course-student-item__line"></view>
							</view>
						</view>
					</scroll-view>
				</view>
			</view>

			<view class="parent-card course-list-card">
				<template v-if="!isAuthenticated">
					<view class="parent-empty-card course-empty-card">
						<view class="parent-empty-badge">报</view>
						<text class="parent-empty-title">登录后即可查看报读课程</text>
						<text class="parent-empty-desc">手机号授权后，系统会自动同步当前已关注学员的报读课程。</text>
						<!-- #ifdef MP-WEIXIN -->
						<button class="parent-primary-button course-auth-button" :loading="authLoading" :disabled="authLoading" open-type="getPhoneNumber" @getphonenumber="handleWechatPhoneAuth">
							授权登录
						</button>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<view class="parent-primary-button course-auth-button" @click="handleMockPhoneAuth">授权登录</view>
						<!-- #endif -->
					</view>
				</template>

				<template v-else-if="pageLoading && !courseList.length">
					<view class="parent-empty-card course-empty-card">
						<view class="parent-empty-badge">载</view>
						<text class="parent-empty-title">正在加载报读课程</text>
						<text class="parent-empty-desc">稍等一下，系统正在同步最新报读信息。</text>
					</view>
				</template>

				<template v-else-if="courseList.length">
					<view
						v-for="item in courseList"
						:key="item.id"
						class="course-item"
						@click="openCourseDetail(item)"
					>
						<view class="course-item__header">
							<text class="course-item__title">{{ item.lessonName }}</text>
							<view class="course-item__arrow"></view>
						</view>
						<view class="course-item__divider"></view>
						<view class="course-item__meta">
							<view class="course-item__row">
								<text class="course-item__label">学员姓名：</text>
								<text class="course-item__value">{{ item.studentName }}</text>
							</view>
							<view class="course-item__row">
								<text class="course-item__label">{{ item.remainingQuantityLabel }}：</text>
								<text class="course-item__value course-item__value--strong">{{ item.remainingQuantityText }}</text>
							</view>
							<view v-if="item.showValidRange" class="course-item__row">
								<text class="course-item__label">有效时段：</text>
								<text class="course-item__value">{{ item.validRangeText }}</text>
							</view>
						</view>
					</view>
				</template>

				<template v-else>
					<view class="parent-empty-card course-empty-card">
						<view class="parent-empty-badge">空</view>
						<text class="parent-empty-title">{{ displayStudents.length ? '当前学员暂无报读课程' : '暂未绑定学员' }}</text>
						<text class="parent-empty-desc">{{ displayStudents.length ? '完成报名或续费后，报读课程会自动同步到这里。' : '请先完成学员关注，再查看报读课程。' }}</text>
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
import { onShow } from '@dcloudio/uni-app'
import BindSuccessDialog from '@/components/bind-success-dialog/bind-success-dialog.vue'
import { authorizeByWechatPhone } from '@/common/parent-auth'
import { listParentCourseEnrollments } from '@/common/parent-api'
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
const courseItems = ref([])
const hasLoadedSummary = ref(false)
const selectedStudentId = ref('')
const pageLoading = ref(false)
const authLoading = computed(() => parentState.authLoading)
const isAuthenticated = computed(() => parentState.isAuthenticated)
const bindSuccessVisible = computed(() => parentState.bindSuccessVisible)
const latestBindStudentName = computed(() => parentState.latestBindStudentName)

let courseRequestSerial = 0

const displayStudents = computed(() => {
	const backendStudents = normalizeStudents(summaryStudents.value || [])
	if (hasLoadedSummary.value) {
		return backendStudents
	}
	return normalizeStudents(parentState.students || [])
})
const shouldCenterStudents = computed(() => displayStudents.value.length > 0 && displayStudents.value.length <= 3)
const courseList = computed(() => normalizeCourseList(courseItems.value || []))

onShow(() => {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		hasLoadedSummary.value = false
		summaryStudents.value = []
		courseItems.value = []
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
	refreshCourseEnrollments()
})

function normalizeStudents(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		id: `${item?.id || item?.studentId || item?.rawId || `student-${index + 1}`}`.trim(),
		name: `${item?.name || item?.studentName || '-'}`.trim() || '-'
	})).filter(item => item.id)
}

function normalizeCourseList(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		id: `${item?.id || `${item?.lessonId || 'lesson'}-${item?.chargingMode || index + 1}`}`.trim(),
		lessonId: `${item?.lessonId || ''}`.trim(),
		lessonName: `${item?.lessonName || '报读课程'}`.trim() || '报读课程',
		studentId: `${item?.studentId || ''}`.trim(),
		studentName: `${item?.studentName || '-'}`.trim() || '-',
		chargingMode: Number(item?.chargingMode || 0),
		remainingQuantityLabel: `${item?.remainingQuantityLabel || '剩余课时'}`.trim() || '剩余课时',
		remainingQuantityText: `${item?.remainingQuantityText || '-'}`.trim() || '-',
		showValidRange: !!item?.showValidRange,
		validRangeText: `${item?.validRangeText || ''}`.trim(),
		status: Number(item?.status || 0),
		statusText: `${item?.statusText || ''}`.trim(),
		lowBalance: !!item?.lowBalance,
		lowBalanceText: `${item?.lowBalanceText || ''}`.trim()
	}))
}

function ensureSelectedStudent() {
	const students = displayStudents.value
	if (!students.length) {
		selectedStudentId.value = ''
		return
	}
	const hasCurrent = students.some(item => item.id === selectedStudentId.value)
	if (hasCurrent) {
		return
	}
	const preferred = `${parentState.currentStudentId || ''}`.trim()
	const matched = students.find(item => item.id === preferred)
	selectedStudentId.value = matched?.id || students[0].id
}

async function refreshCourseEnrollments() {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}

	ensureSelectedStudent()
	const token = `${parentState.authToken || ''}`.trim()
	const targetStudentId = `${selectedStudentId.value || ''}`.trim()
	if (!token) {
		return
	}

	pageLoading.value = true
	const requestSerial = ++courseRequestSerial
	try {
		const result = await listParentCourseEnrollments(token, {
			studentId: targetStudentId
		})
		if (requestSerial !== courseRequestSerial || token !== `${parentState.authToken || ''}`.trim()) {
			return
		}
		hasLoadedSummary.value = true
		summaryStudents.value = result?.students || []
		const nextStudents = normalizeStudents(result?.students || [])
		if (nextStudents.length && !nextStudents.some(item => item.id === selectedStudentId.value)) {
			selectedStudentId.value = nextStudents[0].id
		}
		courseItems.value = result?.items || []
	} catch (error) {
		if (requestSerial !== courseRequestSerial) {
			return
		}
		console.warn('load parent course enrollments failed', error)
		uni.showToast({
			title: `${error?.message || '加载失败'}`.slice(0, 24),
			icon: 'none'
		})
	} finally {
		if (requestSerial === courseRequestSerial) {
			pageLoading.value = false
		}
	}
}

function handleStudentSelect(studentId) {
	if (!studentId || studentId === selectedStudentId.value || pageLoading.value) {
		return
	}
	selectedStudentId.value = studentId
	refreshCourseEnrollments()
	uni.pageScrollTo({
		scrollTop: 0,
		duration: 180
	})
}

function openCourseDetail(item) {
	if (!item?.lessonId || !item?.chargingMode) {
		return
	}
	uni.navigateTo({
		url: `/pages/course-enrollment/detail?studentId=${encodeURIComponent(item.studentId)}&lessonId=${encodeURIComponent(item.lessonId)}&chargingMode=${item.chargingMode}`
	})
}

function handleMockPhoneAuth() {
	setPostAuthPage('/pages/course-enrollment/index')
	authorizeByPhone(DEFAULT_PHONE)
	uni.navigateTo({
		url: '/pages/select-student/index'
	})
}

function handleWechatPhoneAuth(event) {
	authorizeByWechatPhone(event, {
		postAuthPage: '/pages/course-enrollment/index'
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
.course-page {
	min-height: 100vh;
	background:
		radial-gradient(circle at 12% 6%, rgba(255, 221, 169, 0.2), transparent 22%),
		linear-gradient(180deg, #fff7eb 0%, #fff8ef 42%, #fff9f2 100%);
}

.course-header {
	padding-bottom: 18rpx;
}

.course-nav-fixed {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	z-index: 40;
	background:
		linear-gradient(180deg, rgba(255, 247, 235, 0.98) 0%, rgba(255, 248, 239, 0.96) 72%, rgba(255, 249, 242, 0.9) 100%);
	backdrop-filter: blur(12rpx);
}

.course-nav-fixed__inner {
	padding: 0 24rpx 10rpx;
}

.course-nav__side {
	display: flex;
	align-items: center;
}

.course-nav__title {
	flex: 1;
	text-align: center;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.course-back {
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

.course-back__icon {
	width: 16rpx;
	height: 16rpx;
	border-left: 4rpx solid #433e33;
	border-bottom: 4rpx solid #433e33;
	transform: translateX(4rpx) rotate(45deg);
}

.course-student-card {
	margin-top: 18rpx;
	padding: 8rpx 14rpx 4rpx;
}

.course-student-scroll {
	white-space: nowrap;
}

.course-student-list {
	display: inline-flex;
	align-items: stretch;
	min-width: 100%;
}

.course-student-list--center {
	justify-content: center;
	gap: 12rpx;
}

.course-student-item {
	position: relative;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	min-width: 156rpx;
	padding: 16rpx 14rpx 14rpx;
}

.course-student-item--center {
	flex: 0 1 180rpx;
	min-width: 0;
}

.course-student-item__name {
	font-size: 30rpx;
	font-weight: 600;
	line-height: 1.24;
	color: #8f8a80;
}

.course-student-item__line {
	width: 40rpx;
	height: 8rpx;
	margin-top: 14rpx;
	border-radius: 999rpx;
	background: transparent;
}

.course-student-item--active .course-student-item__name {
	color: #222222;
	font-weight: 700;
}

.course-student-item--active .course-student-item__line {
	background: linear-gradient(135deg, #ffe40f 0%, #ffcf0b 100%);
	box-shadow: 0 8rpx 18rpx rgba(255, 214, 10, 0.28);
}

.course-list-card {
	padding: 22rpx 20rpx 26rpx;
}

.course-empty-card {
	padding-top: 54rpx;
	padding-bottom: 54rpx;
}

.course-auth-button {
	width: 100%;
	margin-top: 20rpx;
}

.course-item {
	margin-top: 16rpx;
	padding: 24rpx 22rpx;
	border-radius: 24rpx;
	background: rgba(255, 255, 255, 0.92);
	border: 1rpx solid rgba(230, 221, 204, 0.78);
	box-shadow: var(--parent-shadow);
}

.course-item__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 16rpx;
}

.course-item__title {
	flex: 1;
	min-width: 0;
	font-size: 30rpx;
	font-weight: 700;
	line-height: 1.4;
	color: #333333;
}

.course-item__arrow {
	width: 20rpx;
	height: 20rpx;
	border-top: 4rpx solid #9d9a92;
	border-right: 4rpx solid #9d9a92;
	transform: rotate(45deg);
	flex-shrink: 0;
}

.course-item__divider {
	height: 1rpx;
	margin: 20rpx 0 18rpx;
	background: rgba(234, 228, 218, 0.9);
}

.course-item__meta {
	display: flex;
	flex-direction: column;
	gap: 12rpx;
}

.course-item__row {
	display: flex;
	align-items: flex-start;
}

.course-item__label {
	flex-shrink: 0;
	font-size: 24rpx;
	line-height: 1.5;
	color: #968f82;
}

.course-item__value {
	flex: 1;
	font-size: 24rpx;
	line-height: 1.5;
	color: #6f6b63;
	text-align: left;
}

.course-item__value--strong {
	font-weight: 700;
	color: #353535;
}
</style>
