<template>
	<view class="parent-page record-detail-page">
		<view class="record-detail-nav-fixed" :style="{ paddingTop: `${nav.top}px` }">
			<view class="record-detail-nav-fixed__inner">
				<view class="parent-nav-row record-detail-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="record-detail-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="record-detail-back" @click="goBack">
							<view class="record-detail-back__icon"></view>
						</view>
					</view>
					<text class="record-detail-nav__title">上课记录详情</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>
			</view>
		</view>

		<view class="parent-shell record-detail-shell" :style="{ paddingTop: `${nav.top + nav.height + 18}px` }">
			<template v-if="!isAuthenticated">
				<view class="parent-card record-detail-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">记</view>
						<text class="parent-empty-title">登录后即可查看上课记录详情</text>
						<text class="parent-empty-desc">手机号授权后，系统会自动同步当前学员的上课记录明细。</text>
						<!-- #ifdef MP-WEIXIN -->
						<button class="parent-primary-button record-detail-auth-button" :loading="authLoading" :disabled="authLoading" open-type="getPhoneNumber" @getphonenumber="handleWechatPhoneAuth">
							授权登录
						</button>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<view class="parent-primary-button record-detail-auth-button" @click="handleMockPhoneAuth">授权登录</view>
						<!-- #endif -->
					</view>
				</view>
			</template>

			<template v-else-if="pageLoading && !recordDetail.id">
				<view class="parent-card record-detail-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">载</view>
						<text class="parent-empty-title">正在加载上课记录详情</text>
						<text class="parent-empty-desc">稍等一下，系统正在同步最新记录。</text>
					</view>
				</view>
			</template>

			<template v-else-if="recordDetail.id">
				<view class="record-detail-hero">
					<text class="record-detail-hero__value">{{ recordDetail.summaryText }}</text>
					<text class="record-detail-hero__label">{{ recordDetail.summaryLabel }}</text>
				</view>

				<view class="parent-card record-detail-card">
					<view class="record-detail-card__header">
						<text class="record-detail-card__title">{{ detailTitle }}</text>
						<text class="record-detail-card__status" :class="recordStatusClass(recordDetail.statusText)">{{ recordDetail.statusText }}</text>
					</view>

					<view class="record-detail-card__list">
						<view
							v-for="row in detailRows"
							:key="row.label"
							class="record-detail-row"
						>
							<text class="record-detail-row__label">{{ row.label }}</text>
							<text class="record-detail-row__value" :class="{ 'record-detail-row__value--danger': row.danger }">{{ row.value }}</text>
						</view>
					</view>
				</view>
			</template>

			<template v-else>
				<view class="parent-card record-detail-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">空</view>
						<text class="parent-empty-title">未找到上课记录详情</text>
						<text class="parent-empty-desc">请返回上课记录列表后重新进入。</text>
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
import { getParentClassRecordDetail } from '@/common/parent-api'
import { getNavLayout } from '@/common/nav-layout'
import { DEFAULT_PHONE } from '@/common/mock-parent'
import {
	authorizeByPhone,
	parentState,
	setPostAuthPage
} from '@/common/parent-state'

const nav = getNavLayout()
const routeStudentId = ref('')
const routeRecordId = ref('')
const detailData = ref({})
const pageLoading = ref(false)
const authLoading = computed(() => parentState.authLoading)
const isAuthenticated = computed(() => parentState.isAuthenticated)

let detailRequestSerial = 0

const recordDetail = computed(() => normalizeRecordDetail(detailData.value || {}))
const detailTitle = computed(() => {
	const studentName = recordDetail.value.studentName === '-' ? '' : recordDetail.value.studentName
	const courseName = recordDetail.value.courseName === '-' ? recordDetail.value.className : recordDetail.value.courseName
	const parts = [studentName, courseName].filter(Boolean)
	return parts.length ? parts.join('-') : '上课记录'
})
const detailRows = computed(() => {
	const rows = [
		{
			label: '学员',
			value: recordDetail.value.studentName
		},
		{
			label: '课程',
			value: recordDetail.value.courseName === '-' ? recordDetail.value.className : recordDetail.value.courseName
		},
		{
			label: '上课时间',
			value: recordDetail.value.lessonTime
		}
	]
	if (recordDetail.value.arrearText) {
		rows.push({
			label: recordDetail.value.arrearLabel || '拖欠数量',
			value: recordDetail.value.arrearText,
			danger: true
		})
	}
	if (recordDetail.value.deductText) {
		rows.push({
			label: recordDetail.value.deductLabel || '扣除课时',
			value: recordDetail.value.deductText
		})
	}
	if (recordDetail.value.teacherName !== '-') {
		rows.push({
			label: '上课教师',
			value: recordDetail.value.teacherName
		})
	}
	if (recordDetail.value.classroom !== '-') {
		rows.push({
			label: '上课教室',
			value: recordDetail.value.classroom
		})
	}
	rows.push({
		label: '学校',
		value: recordDetail.value.campusName
	})
	return rows
})

onLoad(query => {
	routeStudentId.value = `${query?.studentId || ''}`.trim()
	routeRecordId.value = `${query?.studentTeachingRecordId || ''}`.trim()
})

onShow(() => {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		resetRecordDetail()
		return
	}
	if (!routeStudentId.value) {
		routeStudentId.value = `${parentState.currentStudentId || ''}`.trim()
	}
	if (!routeRecordId.value) {
		return
	}
	refreshRecordDetail()
})

function normalizeRecordDetail(item = {}) {
	return {
		id: `${item?.id || item?.studentTeachingRecordId || ''}`.trim(),
		studentTeachingRecordId: `${item?.studentTeachingRecordId || item?.id || ''}`.trim(),
		studentId: `${item?.studentId || ''}`.trim(),
		studentName: `${item?.studentName || '-'}`.trim() || '-',
		campusName: `${item?.campusName || '-'}`.trim() || '-',
		className: `${item?.className || item?.courseName || '课程记录'}`.trim() || '课程记录',
		courseName: `${item?.courseName || item?.className || '-'}`.trim() || '-',
		teacherName: `${item?.teacherName || '-'}`.trim() || '-',
		classroom: `${item?.classroom || '-'}`.trim() || '-',
		statusText: `${item?.statusText || '-'}`.trim() || '-',
		summaryText: `${item?.summaryText || '-'}`.trim() || '-',
		summaryLabel: `${item?.summaryLabel || '上课记录'}`.trim() || '上课记录',
		deductLabel: `${item?.deductLabel || ''}`.trim(),
		deductText: `${item?.deductText || ''}`.trim(),
		arrearLabel: `${item?.arrearLabel || ''}`.trim(),
		arrearText: `${item?.arrearText || ''}`.trim(),
		lessonTime: `${item?.lessonTime || ''}`.trim() || buildLessonTimeText(item),
		date: `${item?.date || ''}`.trim(),
		startTime: `${item?.startTime || ''}`.trim(),
		endTime: `${item?.endTime || ''}`.trim()
	}
}

function buildLessonTimeText(item = {}) {
	const dateText = `${item?.date || ''}`.trim()
	const startText = `${item?.startTime || ''}`.trim()
	const endText = `${item?.endTime || ''}`.trim()
	if (dateText && startText && endText) {
		return `${dateText} ${startText}~${endText}`
	}
	if (dateText && startText) {
		return `${dateText} ${startText}`
	}
	return dateText || startText || '-'
}

function resetRecordDetail() {
	detailData.value = {}
	pageLoading.value = false
}

async function refreshRecordDetail() {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	if (!routeRecordId.value) {
		return
	}

	const token = `${parentState.authToken || ''}`.trim()
	if (!token) {
		return
	}

	pageLoading.value = true
	const requestSerial = ++detailRequestSerial
	try {
		const result = await getParentClassRecordDetail(token, {
			studentId: routeStudentId.value,
			studentTeachingRecordId: routeRecordId.value
		})
		if (requestSerial !== detailRequestSerial || token !== `${parentState.authToken || ''}`.trim()) {
			return
		}
		detailData.value = result || {}
	} catch (error) {
		if (requestSerial !== detailRequestSerial) {
			return
		}
		console.warn('load parent class record detail failed', error)
		uni.showToast({
			title: `${error?.message || '加载失败'}`.slice(0, 24),
			icon: 'none'
		})
	} finally {
		if (requestSerial === detailRequestSerial) {
			pageLoading.value = false
		}
	}
}

function buildCurrentPageURL() {
	const query = [`studentTeachingRecordId=${encodeURIComponent(routeRecordId.value)}`]
	if (routeStudentId.value) {
		query.push(`studentId=${encodeURIComponent(routeStudentId.value)}`)
	}
	return `/pages/attendance-record/detail?${query.join('&')}`
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
				url: '/pages/attendance-record/index'
			})
		}
	})
}

function recordStatusClass(statusText = '') {
	const text = `${statusText || ''}`.trim()
	if (text === '到课') {
		return 'record-detail-card__status--arrived'
	}
	if (text === '请假') {
		return 'record-detail-card__status--leave'
	}
	if (text === '旷课') {
		return 'record-detail-card__status--absent'
	}
	if (text === '未记录') {
		return 'record-detail-card__status--pending'
	}
	return ''
}
</script>

<style scoped>
.record-detail-page {
	min-height: 100vh;
	background:
		radial-gradient(circle at 12% 6%, rgba(255, 221, 169, 0.2), transparent 22%),
		linear-gradient(180deg, #fff7eb 0%, #fff8ef 42%, #fff9f2 100%);
}

.record-detail-nav-fixed {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	z-index: 40;
	background:
		linear-gradient(180deg, rgba(255, 247, 235, 0.98) 0%, rgba(255, 248, 239, 0.96) 72%, rgba(255, 249, 242, 0.9) 100%);
	backdrop-filter: blur(12rpx);
}

.record-detail-nav-fixed__inner {
	padding: 0 24rpx 10rpx;
}

.record-detail-nav__side {
	display: flex;
	align-items: center;
}

.record-detail-nav__title {
	flex: 1;
	text-align: center;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.record-detail-back {
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

.record-detail-back__icon {
	width: 16rpx;
	height: 16rpx;
	border-left: 4rpx solid #433e33;
	border-bottom: 4rpx solid #433e33;
	transform: translateX(4rpx) rotate(45deg);
}

.record-detail-shell {
	padding-bottom: 56rpx;
}

.record-detail-empty-card {
	padding: 0;
}

.record-detail-auth-button {
	width: 100%;
	margin-top: 20rpx;
}

.record-detail-hero {
	padding: 56rpx 0 40rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	text-align: center;
}

.record-detail-hero__value {
	font-size: 78rpx;
	font-weight: 700;
	line-height: 1;
	color: #1f1f1f;
}

.record-detail-hero__label {
	margin-top: 18rpx;
	font-size: 28rpx;
	line-height: 1.4;
	color: #9a9387;
}

.record-detail-card {
	padding: 30rpx 28rpx 12rpx;
}

.record-detail-card__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 18rpx;
}

.record-detail-card__title {
	flex: 1;
	min-width: 0;
	font-size: 32rpx;
	font-weight: 700;
	line-height: 1.45;
	color: #2a2721;
}

.record-detail-card__status {
	flex-shrink: 0;
	padding: 10rpx 18rpx;
	border-radius: 999rpx;
	font-size: 22rpx;
	font-weight: 700;
	line-height: 1;
	background: rgba(233, 237, 243, 0.96);
	color: #7c8796;
}

.record-detail-card__status--arrived {
	background: rgba(228, 240, 255, 0.96);
	color: #2f86ff;
}

.record-detail-card__status--leave {
	background: rgba(255, 241, 204, 0.98);
	color: #b27a00;
}

.record-detail-card__status--absent {
	background: rgba(255, 230, 225, 0.96);
	color: #dd5b4f;
}

.record-detail-card__status--pending {
	background: rgba(238, 239, 241, 0.96);
	color: #8f95a0;
}

.record-detail-card__list {
	margin-top: 16rpx;
}

.record-detail-row {
	display: flex;
	align-items: flex-start;
	padding: 24rpx 0;
	border-top: 1rpx solid var(--parent-divider);
	gap: 18rpx;
}

.record-detail-row__label {
	width: 132rpx;
	flex-shrink: 0;
	font-size: 26rpx;
	line-height: 1.65;
	color: #938c80;
}

.record-detail-row__value {
	flex: 1;
	font-size: 26rpx;
	line-height: 1.65;
	color: #2d2a24;
	text-align: right;
	word-break: break-all;
}

.record-detail-row__value--danger {
	color: #e46349;
	font-weight: 700;
}
</style>
