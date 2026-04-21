<template>
	<view class="parent-page course-arrear-page">
		<view class="course-arrear-nav-fixed" :style="{ paddingTop: `${nav.top}px` }">
			<view class="course-arrear-nav-fixed__inner">
				<view class="parent-nav-row course-arrear-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="course-arrear-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="course-arrear-back" @click="goBack">
							<view class="course-arrear-back__icon"></view>
						</view>
					</view>
					<text class="course-arrear-nav__title">欠费记录</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>
			</view>
		</view>

		<view class="parent-shell course-arrear-shell" :style="{ paddingTop: `${nav.top + nav.height + 8}px` }">
			<template v-if="!isAuthenticated">
				<view class="parent-card course-arrear-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">欠</view>
						<text class="parent-empty-title">登录后即可查看欠费记录</text>
						<text class="parent-empty-desc">手机号授权后，系统会自动同步当前学员的课消欠费明细。</text>
						<!-- #ifdef MP-WEIXIN -->
						<button class="parent-primary-button course-arrear-auth-button" :loading="authLoading" :disabled="authLoading" open-type="getPhoneNumber" @getphonenumber="handleWechatPhoneAuth">
							授权登录
						</button>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<view class="parent-primary-button course-arrear-auth-button" @click="handleMockPhoneAuth">授权登录</view>
						<!-- #endif -->
					</view>
				</view>
			</template>

			<template v-else-if="pageLoading && !arrearCourses.length">
				<view class="parent-card course-arrear-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">载</view>
						<text class="parent-empty-title">正在加载欠费记录</text>
						<text class="parent-empty-desc">稍等一下，系统正在同步最新欠费数据。</text>
					</view>
				</view>
			</template>

			<template v-else-if="arrearCourses.length">
				<view class="course-arrear-overview">
					<text class="course-arrear-overview__text">当前共计 {{ courseCount }} 个课程</text>
				</view>

				<view
					v-for="course in arrearCourses"
					:key="course.id"
					class="course-arrear-section"
				>
					<view class="course-arrear-section__summary">
						<text class="course-arrear-card__title">{{ course.lessonName }}</text>
						<text class="course-arrear-card__total">{{ course.totalArrearLabel }}：{{ course.totalArrearText }}</text>
					</view>

					<view class="course-arrear-section__records">
						<view
							v-for="record in course.items"
							:key="record.id"
							class="course-arrear-record"
							@click="openArrearDetail(course, record)"
						>
							<view class="course-arrear-record__main">
								<text class="course-arrear-record__label">上课时间：</text>
								<text class="course-arrear-record__time">{{ record.lessonTime }}</text>
								<text class="course-arrear-record__value">{{ record.arrearText }}</text>
							</view>
							<view class="course-arrear-record__arrow"></view>
						</view>
					</view>
				</view>
			</template>

			<template v-else>
				<view class="parent-card course-arrear-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">空</view>
						<text class="parent-empty-title">暂无欠费记录</text>
						<text class="parent-empty-desc">当前学员暂时没有未补扣的课消欠费记录。</text>
					</view>
				</view>
			</template>
		</view>
	</view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { authorizeByWechatPhone } from '@/common/parent-auth'
import { getParentCourseArrears } from '@/common/parent-api'
import { getNavLayout } from '@/common/nav-layout'
import { DEFAULT_PHONE } from '@/common/mock-parent'
import {
	authorizeByPhone,
	parentState,
	setPostAuthPage
} from '@/common/parent-state'

const nav = getNavLayout()
const routeStudentId = ref('')
const routeLessonId = ref('')
const routeChargingMode = ref(0)
const arrearSummary = ref({})
const pageLoading = ref(false)
const authLoading = computed(() => parentState.authLoading)
const isAuthenticated = computed(() => parentState.isAuthenticated)

let arrearRequestSerial = 0

const arrearCourses = computed(() => normalizeArrearCourses(arrearSummary.value?.items || []))
const courseCount = computed(() => {
	const value = Number(arrearSummary.value?.courseCount || 0)
	if (value > 0) {
		return value
	}
	return arrearCourses.value.length
})

onLoad(query => {
	routeStudentId.value = `${query?.studentId || ''}`.trim()
	routeLessonId.value = `${query?.lessonId || ''}`.trim()
	routeChargingMode.value = Number(query?.chargingMode || 0)
})

onShow(() => {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		resetArrearSummary()
		return
	}
	if (!routeStudentId.value) {
		routeStudentId.value = resolveStudentId()
	}
	if (!routeStudentId.value) {
		return
	}
	refreshCourseArrears()
})

function normalizeArrearCourses(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		id: `${item?.id || `arrear-course-${index + 1}`}`.trim(),
		lessonId: `${item?.lessonId || ''}`.trim(),
		lessonName: `${item?.lessonName || '报读课程'}`.trim() || '报读课程',
		chargingMode: Number(item?.chargingMode || 0),
		chargingModeText: `${item?.chargingModeText || ''}`.trim(),
		totalArrearLabel: `${item?.totalArrearLabel || '总计拖欠课时'}`.trim() || '总计拖欠课时',
		totalArrearText: `${item?.totalArrearText || '-'}`.trim() || '-',
		recordCount: Number(item?.recordCount || 0),
		items: normalizeArrearRecords(item?.items || [])
	}))
}

function normalizeArrearRecords(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		id: `${item?.id || item?.studentTeachingRecordId || `arrear-record-${index + 1}`}`.trim(),
		studentTeachingRecordId: `${item?.studentTeachingRecordId || ''}`.trim(),
		lessonTime: `${item?.lessonTime || '-'}`.trim() || '-',
		arrearText: `${item?.arrearText || '-'}`.trim() || '-'
	}))
}

function resolveStudentId() {
	const preferred = `${parentState.currentStudentId || ''}`.trim()
	if (preferred) {
		return preferred
	}
	const firstStudent = Array.isArray(parentState.students) ? parentState.students[0] : null
	return `${firstStudent?.id || firstStudent?.studentId || ''}`.trim()
}

function resetArrearSummary() {
	arrearSummary.value = {}
	pageLoading.value = false
}

async function refreshCourseArrears() {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	const token = `${parentState.authToken || ''}`.trim()
	if (!token || !routeStudentId.value) {
		return
	}

	pageLoading.value = true
	const requestSerial = ++arrearRequestSerial
	try {
		const result = await getParentCourseArrears(token, {
			studentId: routeStudentId.value,
			lessonId: routeLessonId.value,
			chargingMode: routeChargingMode.value
		})
		if (requestSerial !== arrearRequestSerial || token !== `${parentState.authToken || ''}`.trim()) {
			return
		}
		arrearSummary.value = result || {}
	} catch (error) {
		if (requestSerial !== arrearRequestSerial) {
			return
		}
		console.warn('load parent course arrears failed', error)
		uni.showToast({
			title: `${error?.message || '加载失败'}`.slice(0, 24),
			icon: 'none'
		})
	} finally {
		if (requestSerial === arrearRequestSerial) {
			pageLoading.value = false
		}
	}
}

function buildCurrentPageURL() {
	const query = [`studentId=${encodeURIComponent(routeStudentId.value)}`]
	if (routeLessonId.value) {
		query.push(`lessonId=${encodeURIComponent(routeLessonId.value)}`)
	}
	if (routeChargingMode.value > 0) {
		query.push(`chargingMode=${routeChargingMode.value}`)
	}
	return `/pages/course-arrear/index?${query.join('&')}`
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
			uni.redirectTo({
				url: '/pages/course-enrollment/index'
			})
		}
	})
}

function openArrearDetail(course, record) {
	const lessonName = `${course?.lessonName || '欠费记录'}`.trim()
	const recordID = `${record?.studentTeachingRecordId || record?.id || ''}`.trim()
	console.log('open arrear detail placeholder', {
		lessonName,
		recordID
	})
	uni.showToast({
		title: '欠费详情暂未开放',
		icon: 'none'
	})
}
</script>

<style scoped>
.course-arrear-page {
	min-height: 100vh;
	background: #f5f6f8;
}

.course-arrear-nav-fixed {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	z-index: 40;
	background: rgba(255, 255, 255, 0.98);
	backdrop-filter: blur(12rpx);
	border-bottom: 1rpx solid #f0f0f0;
}

.course-arrear-nav-fixed__inner {
	padding: 0 24rpx 10rpx;
}

.course-arrear-nav__side {
	display: flex;
	align-items: center;
}

.course-arrear-nav__title {
	flex: 1;
	text-align: center;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.course-arrear-back {
	width: 56rpx;
	height: 56rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.course-arrear-back__icon {
	width: 18rpx;
	height: 18rpx;
	border-left: 4rpx solid #252525;
	border-bottom: 4rpx solid #252525;
	transform: translateX(6rpx) rotate(45deg);
}

.course-arrear-empty-card {
	margin: 0 24rpx;
	padding: 0;
}

.course-arrear-auth-button {
	width: 100%;
	margin-top: 20rpx;
}

.course-arrear-shell {
	padding-right: 0;
	padding-bottom: 56rpx;
	padding-left: 0;
}

.course-arrear-overview {
	margin: 0 0 16rpx;
	padding: 24rpx 32rpx;
	background: #f1f2f4;
}

.course-arrear-overview__text {
	font-size: 26rpx;
	font-weight: 600;
	line-height: 1.4;
	color: #5a5a5a;
}

.course-arrear-section {
	margin: 0 0 16rpx;
	background: #ffffff;
}

.course-arrear-section__summary {
	padding: 30rpx 32rpx 26rpx;
	display: flex;
	flex-direction: column;
	align-items: flex-start;
}

.course-arrear-card__title {
	font-size: 36rpx;
	font-weight: 700;
	line-height: 1.4;
	color: #232323;
}

.course-arrear-card__total {
	display: block;
	margin-top: 18rpx;
	font-size: 26rpx;
	font-weight: 700;
	line-height: 1.45;
	color: #ff5b43;
}

.course-arrear-section__records {
	background: #ffffff;
}

.course-arrear-record {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 20rpx;
	min-height: 94rpx;
	padding: 0 32rpx;
	border-top: 1rpx solid #f0f0f0;
	background: #ffffff;
}

.course-arrear-record__main {
	flex: 1;
	min-width: 0;
	display: flex;
	align-items: center;
	flex-wrap: nowrap;
	gap: 8rpx;
}

.course-arrear-record__label {
	flex-shrink: 0;
	font-size: 26rpx;
	line-height: 1.45;
	color: #6c6c6c;
}

.course-arrear-record__time {
	min-width: 0;
	font-size: 26rpx;
	line-height: 1.45;
	color: #6c6c6c;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.course-arrear-record__value {
	flex-shrink: 0;
	margin-left: 12rpx;
	font-size: 26rpx;
	line-height: 1.45;
	color: #6c6c6c;
	white-space: nowrap;
}

.course-arrear-record__arrow {
	width: 18rpx;
	height: 18rpx;
	flex-shrink: 0;
	border-top: 4rpx solid #a5a5a5;
	border-right: 4rpx solid #a5a5a5;
	transform: rotate(45deg);
}
</style>
