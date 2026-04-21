<template>
	<view class="parent-page course-detail-page">
		<view class="course-detail-nav-fixed" :style="{ paddingTop: `${nav.top}px` }">
			<view class="course-detail-nav-fixed__inner">
				<view class="parent-nav-row course-detail-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="course-detail-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="course-detail-back" @click="goBack">
							<view class="course-detail-back__icon"></view>
						</view>
					</view>
					<text class="course-detail-nav__title">报读课程详情</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>
			</view>
		</view>

		<view class="parent-shell" :style="{ paddingTop: `${nav.top + nav.height + 18}px` }">
			<template v-if="!isAuthenticated">
				<view class="parent-card course-detail-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">报</view>
						<text class="parent-empty-title">登录后即可查看课程详情</text>
						<text class="parent-empty-desc">手机号授权后，系统会自动同步报读课程与变动明细。</text>
						<!-- #ifdef MP-WEIXIN -->
						<button class="parent-primary-button course-detail-auth-button" :loading="authLoading" :disabled="authLoading" open-type="getPhoneNumber" @getphonenumber="handleWechatPhoneAuth">
							授权登录
						</button>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<view class="parent-primary-button course-detail-auth-button" @click="handleMockPhoneAuth">授权登录</view>
						<!-- #endif -->
					</view>
				</view>
			</template>

			<template v-else-if="pageLoading && !courseDetail.id">
				<view class="parent-card course-detail-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">载</view>
						<text class="parent-empty-title">正在加载课程详情</text>
						<text class="parent-empty-desc">稍等一下，系统正在同步变动明细。</text>
					</view>
				</view>
			</template>

			<template v-else-if="courseDetail.id">
				<view class="parent-card course-detail-summary-card">
					<view class="course-detail-student">
						<view class="course-detail-student__avatar">
							<image
								v-if="studentAvatarUrl"
								class="course-detail-student__avatar-image"
								:src="studentAvatarUrl"
								mode="aspectFill"
							></image>
							<text v-else>{{ studentAvatarText }}</text>
						</view>
						<text class="course-detail-student__name">{{ studentName }}</text>
					</view>

					<text class="course-detail-summary-card__title">{{ courseDetail.lessonName }}</text>

					<view class="course-detail-summary-card__meta">
						<view class="course-detail-summary-card__row">
							<text class="course-detail-summary-card__label">剩余学费</text>
							<text class="course-detail-summary-card__value">{{ courseDetail.remainingTuitionText }}</text>
						</view>
						<view v-if="courseDetail.chargingMode !== 3" class="course-detail-summary-card__row">
							<text class="course-detail-summary-card__label">{{ courseDetail.remainingQuantityLabel }}</text>
							<text class="course-detail-summary-card__value">{{ courseDetail.remainingQuantityText }}</text>
						</view>
						<view v-if="courseDetail.lowBalanceText" class="course-detail-warning">
							<text class="course-detail-warning__icon">!</text>
							<text class="course-detail-warning__text">{{ courseDetail.lowBalanceText }}</text>
						</view>
						<view v-if="courseDetail.showValidRange" class="course-detail-summary-card__row course-detail-summary-card__row--range">
							<text class="course-detail-summary-card__label">有效时段</text>
							<view class="course-detail-summary-card__range">
								<text class="course-detail-summary-card__value">{{ courseDetail.validRangeText }}</text>
								<text class="course-detail-summary-card__status" :class="courseStatusClass(courseDetail.status)">{{ courseDetail.statusText }}</text>
							</view>
						</view>
					</view>
				</view>

				<view class="course-detail-section">
					<text class="course-detail-section__title">变动明细</text>
				</view>

				<view class="parent-card course-detail-list-card">
					<template v-if="flowList.length">
						<view
							v-for="(item, index) in flowList"
							:key="item.id"
							class="course-detail-flow"
							:class="{ 'course-detail-flow--last': index === flowList.length - 1 && !showFooterLine }"
						>
							<view class="course-detail-flow__main">
								<text class="course-detail-flow__title">{{ item.title }}</text>
								<text class="course-detail-flow__time">{{ formatFlowTime(item.createdAt) }}</text>
							</view>
							<text class="course-detail-flow__value" :class="{ 'course-detail-flow__value--positive': item.highlightPositive }">{{ item.quantityText }}</text>
						</view>
						<view v-if="moreLoading" class="course-detail-loading-text">正在加载更多...</view>
						<view v-else-if="!hasMore && flowList.length >= pageSize" class="course-detail-loading-text">没有更多记录了</view>
					</template>

					<template v-else>
						<view class="parent-empty-card course-detail-empty-card course-detail-empty-card--inner">
							<view class="parent-empty-badge">空</view>
							<text class="parent-empty-title">暂无变动明细</text>
							<text class="parent-empty-desc">当前课程还没有可展示的明细记录。</text>
						</view>
					</template>
				</view>
			</template>

			<template v-else>
				<view class="parent-card course-detail-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">空</view>
						<text class="parent-empty-title">未找到课程详情</text>
						<text class="parent-empty-desc">请返回报读课程列表后重新进入。</text>
					</view>
				</view>
			</template>
		</view>
	</view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onLoad, onReachBottom, onShow } from '@dcloudio/uni-app'
import { authorizeByWechatPhone } from '@/common/parent-auth'
import { getParentCourseEnrollmentDetail } from '@/common/parent-api'
import { getNavLayout } from '@/common/nav-layout'
import { DEFAULT_PHONE } from '@/common/mock-parent'
import {
	authorizeByPhone,
	parentState,
	setPostAuthPage
} from '@/common/parent-state'

const nav = getNavLayout()
const pageSize = 20
const routeStudentId = ref('')
const routeLessonId = ref('')
const routeChargingMode = ref(0)
const detailStudent = ref({})
const detailCourse = ref({})
const flowItems = ref([])
const pageIndex = ref(1)
const hasMore = ref(false)
const pageLoading = ref(false)
const moreLoading = ref(false)
const authLoading = computed(() => parentState.authLoading)
const isAuthenticated = computed(() => parentState.isAuthenticated)

let detailRequestSerial = 0

const courseDetail = computed(() => normalizeCourseDetail(detailCourse.value || {}))
const flowList = computed(() => normalizeFlowList(flowItems.value || []))
const studentName = computed(() => `${detailStudent.value?.name || courseDetail.value?.studentName || '学员'}`.trim() || '学员')
const studentAvatarUrl = computed(() => `${detailStudent.value?.avatarUrl || courseDetail.value?.studentAvatarUrl || ''}`.trim())
const studentAvatarText = computed(() => studentName.value.slice(0, 1))
const showFooterLine = computed(() => moreLoading.value || (!hasMore.value && flowList.value.length >= pageSize))

onLoad(query => {
	routeStudentId.value = `${query?.studentId || ''}`.trim()
	routeLessonId.value = `${query?.lessonId || ''}`.trim()
	routeChargingMode.value = Number(query?.chargingMode || 0)
})

onShow(() => {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		resetCourseDetail()
		return
	}
	if (!routeLessonId.value || !routeChargingMode.value) {
		return
	}
	refreshCourseDetail({
		reset: true
	})
})

onReachBottom(() => {
	loadMoreFlows()
})

function normalizeCourseDetail(item = {}) {
	return {
		id: `${item?.id || ''}`.trim(),
		studentId: `${item?.studentId || ''}`.trim(),
		studentName: `${item?.studentName || '-'}`.trim() || '-',
		studentAvatarUrl: `${item?.studentAvatarUrl || ''}`.trim(),
		lessonName: `${item?.lessonName || '报读课程'}`.trim() || '报读课程',
		chargingMode: Number(item?.chargingMode || 0),
		status: Number(item?.status || 0),
		statusText: `${item?.statusText || '-'}`.trim() || '-',
		remainingTuitionText: `${item?.remainingTuitionText || '-'}`.trim() || '-',
		remainingQuantityLabel: `${item?.remainingQuantityLabel || '剩余课时'}`.trim() || '剩余课时',
		remainingQuantityText: `${item?.remainingQuantityText || '-'}`.trim() || '-',
		showValidRange: !!item?.showValidRange,
		validRangeText: `${item?.validRangeText || ''}`.trim(),
		lowBalanceText: `${item?.lowBalanceText || ''}`.trim()
	}
}

function normalizeFlowList(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		id: `${item?.id || `flow-${index + 1}`}`.trim(),
		title: `${item?.title || '账户变动'}`.trim() || '账户变动',
		createdAt: `${item?.createdAt || ''}`.trim(),
		quantityText: `${item?.quantityText || '-'}`.trim() || '-',
		highlightPositive: !!item?.highlightPositive
	}))
}

function resetCourseDetail() {
	detailStudent.value = {}
	detailCourse.value = {}
	flowItems.value = []
	pageIndex.value = 1
	hasMore.value = false
	pageLoading.value = false
	moreLoading.value = false
}

async function refreshCourseDetail(options = {}) {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	if (!routeLessonId.value || !routeChargingMode.value) {
		return
	}

	const reset = options?.reset !== false
	const token = `${parentState.authToken || ''}`.trim()
	if (!token) {
		return
	}

	const requestPageIndex = reset ? 1 : pageIndex.value + 1
	if (reset) {
		pageLoading.value = true
	} else {
		moreLoading.value = true
	}
	const requestSerial = ++detailRequestSerial
	try {
		const result = await getParentCourseEnrollmentDetail(token, {
			studentId: routeStudentId.value,
			lessonId: routeLessonId.value,
			chargingMode: routeChargingMode.value,
			pageIndex: requestPageIndex,
			pageSize
		})
		if (requestSerial !== detailRequestSerial || token !== `${parentState.authToken || ''}`.trim()) {
			return
		}
		detailStudent.value = result?.student || {}
		detailCourse.value = result?.course || {}
		flowItems.value = reset ? (result?.items || []) : mergePagedFlowItems(flowItems.value, result?.items || [])
		pageIndex.value = Number(result?.pageIndex || requestPageIndex)
		hasMore.value = !!result?.hasMore
	} catch (error) {
		if (requestSerial !== detailRequestSerial) {
			return
		}
		console.warn('load parent course detail failed', error)
		if (reset) {
			uni.showToast({
				title: `${error?.message || '加载失败'}`.slice(0, 24),
				icon: 'none'
			})
		}
	} finally {
		if (requestSerial === detailRequestSerial && reset) {
			pageLoading.value = false
		}
		if (requestSerial === detailRequestSerial && !reset) {
			moreLoading.value = false
		}
	}
}

function loadMoreFlows() {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	if (!hasMore.value || pageLoading.value || moreLoading.value) {
		return
	}
	refreshCourseDetail({
		reset: false
	})
}

function mergePagedFlowItems(currentItems = [], nextItems = []) {
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

function formatFlowTime(createdAt = '') {
	const text = `${createdAt || ''}`.trim()
	if (!text) {
		return '-'
	}
	const normalized = text.replace('T', ' ').slice(0, 16)
	if (normalized.length < 16) {
		return normalized
	}
	const today = getTodayText()
	if (normalized.slice(0, 10) === today) {
		return normalized.slice(11, 16)
	}
	return normalized.slice(5, 16)
}

function getTodayText() {
	const today = new Date()
	const year = today.getFullYear()
	const month = `${today.getMonth() + 1}`.padStart(2, '0')
	const day = `${today.getDate()}`.padStart(2, '0')
	return `${year}-${month}-${day}`
}

function courseStatusClass(status) {
	if (Number(status) === 2) {
		return 'course-detail-summary-card__status--suspended'
	}
	if (Number(status) === 3) {
		return 'course-detail-summary-card__status--closed'
	}
	return 'course-detail-summary-card__status--active'
}

function handleMockPhoneAuth() {
	setPostAuthPage(`/pages/course-enrollment/detail?studentId=${encodeURIComponent(routeStudentId.value)}&lessonId=${encodeURIComponent(routeLessonId.value)}&chargingMode=${routeChargingMode.value}`)
	authorizeByPhone(DEFAULT_PHONE)
	uni.navigateTo({
		url: '/pages/select-student/index'
	})
}

function handleWechatPhoneAuth(event) {
	authorizeByWechatPhone(event, {
		postAuthPage: `/pages/course-enrollment/detail?studentId=${encodeURIComponent(routeStudentId.value)}&lessonId=${encodeURIComponent(routeLessonId.value)}&chargingMode=${routeChargingMode.value}`
	})
}

function goBack() {
	uni.navigateBack({
		fail() {
			uni.redirectTo({
				url: '/pages/course-enrollment/index'
			})
		}
	})
}
</script>

<style scoped>
.course-detail-page {
	min-height: 100vh;
	background:
		radial-gradient(circle at 12% 6%, rgba(255, 221, 169, 0.2), transparent 22%),
		linear-gradient(180deg, #fff7eb 0%, #fff8ef 42%, #fff9f2 100%);
}

.course-detail-nav-fixed {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	z-index: 40;
	background:
		linear-gradient(180deg, rgba(255, 247, 235, 0.98) 0%, rgba(255, 248, 239, 0.96) 72%, rgba(255, 249, 242, 0.9) 100%);
	backdrop-filter: blur(12rpx);
}

.course-detail-nav-fixed__inner {
	padding: 0 24rpx 10rpx;
}

.course-detail-nav__side {
	display: flex;
	align-items: center;
}

.course-detail-nav__title {
	flex: 1;
	text-align: center;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.course-detail-back {
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

.course-detail-back__icon {
	width: 16rpx;
	height: 16rpx;
	border-left: 4rpx solid #433e33;
	border-bottom: 4rpx solid #433e33;
	transform: translateX(4rpx) rotate(45deg);
}

.course-detail-empty-card {
	padding: 0;
}

.course-detail-empty-card--inner {
	padding-top: 54rpx;
	padding-bottom: 54rpx;
}

.course-detail-auth-button {
	width: 100%;
	margin-top: 20rpx;
}

.course-detail-summary-card {
	padding: 34rpx 26rpx 20rpx;
}

.course-detail-student {
	display: flex;
	flex-direction: column;
	align-items: center;
}

.course-detail-student__avatar {
	width: 112rpx;
	height: 112rpx;
	border-radius: 50%;
	background: #a9c8ff;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	font-size: 38rpx;
	font-weight: 700;
	overflow: hidden;
	box-shadow: 0 16rpx 34rpx rgba(126, 164, 235, 0.22);
}

.course-detail-student__avatar-image {
	width: 100%;
	height: 100%;
	border-radius: 50%;
}

.course-detail-student__name {
	margin-top: 18rpx;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1.3;
	color: #303030;
}

.course-detail-summary-card__title {
	display: block;
	margin-top: 34rpx;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1.4;
	color: #333333;
}

.course-detail-summary-card__meta {
	margin-top: 26rpx;
	display: flex;
	flex-direction: column;
	gap: 24rpx;
}

.course-detail-summary-card__row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 18rpx;
}

.course-detail-summary-card__row--range {
	align-items: flex-start;
}

.course-detail-summary-card__label {
	flex-shrink: 0;
	font-size: 26rpx;
	line-height: 1.45;
	color: #9b968a;
}

.course-detail-summary-card__value {
	flex: 1;
	text-align: right;
	font-size: 26rpx;
	font-weight: 700;
	line-height: 1.45;
	color: #2f2f2f;
}

.course-detail-summary-card__range {
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: flex-end;
	flex-wrap: wrap;
	gap: 12rpx;
}

.course-detail-summary-card__status {
	padding: 8rpx 16rpx;
	border-radius: 999rpx;
	font-size: 22rpx;
	font-weight: 700;
	line-height: 1;
}

.course-detail-summary-card__status--active {
	background: rgba(227, 239, 255, 0.96);
	color: #2f86ff;
}

.course-detail-summary-card__status--suspended {
	background: rgba(255, 241, 204, 0.98);
	color: #b27a00;
}

.course-detail-summary-card__status--closed {
	background: rgba(238, 239, 241, 0.96);
	color: #8f95a0;
}

.course-detail-warning {
	display: flex;
	align-items: center;
	justify-content: flex-end;
	gap: 10rpx;
}

.course-detail-warning__icon {
	width: 34rpx;
	height: 34rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, #ffba1b 0%, #ff9900 100%);
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	font-size: 22rpx;
	font-weight: 700;
	line-height: 1;
}

.course-detail-warning__text {
	font-size: 24rpx;
	font-weight: 600;
	line-height: 1.5;
	color: #d08a00;
}

.course-detail-section {
	margin: 22rpx -24rpx 0;
	padding: 22rpx 24rpx;
	background: rgba(245, 241, 233, 0.9);
}

.course-detail-section__title {
	font-size: 28rpx;
	font-weight: 600;
	color: #7f796d;
}

.course-detail-list-card {
	padding: 0 22rpx;
	border-top-left-radius: 0;
	border-top-right-radius: 0;
}

.course-detail-flow {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 20rpx;
	padding: 24rpx 0;
	border-bottom: 1rpx solid rgba(234, 228, 218, 0.9);
}

.course-detail-flow--last {
	border-bottom: none;
}

.course-detail-flow__main {
	flex: 1;
	min-width: 0;
	display: flex;
	flex-direction: column;
}

.course-detail-flow__title {
	font-size: 30rpx;
	font-weight: 700;
	line-height: 1.4;
	color: #3a3936;
}

.course-detail-flow__time {
	margin-top: 8rpx;
	font-size: 24rpx;
	line-height: 1.45;
	color: #9b968a;
}

.course-detail-flow__value {
	flex-shrink: 0;
	font-size: 26rpx;
	font-weight: 700;
	line-height: 1.3;
	color: #2f2f2f;
}

.course-detail-flow__value--positive {
	color: #f0c400;
}

.course-detail-loading-text {
	padding: 22rpx 0 26rpx;
	text-align: center;
	font-size: 22rpx;
	color: var(--parent-subtext);
}
</style>
