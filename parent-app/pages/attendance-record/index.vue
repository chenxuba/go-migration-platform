<template>
	<view class="parent-page record-page">
		<view class="record-nav-fixed" :style="{ paddingTop: `${nav.top}px` }">
			<view class="record-nav-fixed__inner">
				<view class="parent-nav-row record-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="record-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="record-back" @click="goBack">
							<view class="record-back__icon"></view>
						</view>
					</view>
					<text class="record-nav__title">上课记录</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>
			</view>
		</view>

		<view class="parent-shell" :style="{ paddingTop: `${nav.top + nav.height + 18}px` }">
			<view class="parent-header record-header">
				<view v-if="isAuthenticated && displayStudents.length" class="parent-card record-student-card">
					<scroll-view class="record-student-scroll" :scroll-x="!shouldCenterStudents" show-scrollbar="false" enable-flex>
						<view
							class="record-student-list"
							:class="{ 'record-student-list--center': shouldCenterStudents }"
						>
							<view
								v-for="student in displayStudents"
								:key="student.id"
								class="record-student-item"
								:class="{
									'record-student-item--active': selectedStudentId === student.id,
									'record-student-item--center': shouldCenterStudents
								}"
								@click="handleStudentSelect(student.id)"
							>
								<text class="record-student-item__name">{{ student.name }}</text>
								<view class="record-student-item__line"></view>
							</view>
						</view>
					</scroll-view>
				</view>
			</view>

			<view class="parent-card record-list-card">
				<template v-if="!isAuthenticated">
					<view class="parent-empty-card record-empty-card">
						<view class="parent-empty-badge">记</view>
						<text class="parent-empty-title">登录后即可查看上课记录</text>
						<text class="parent-empty-desc">手机号授权后，系统会自动匹配已关注学员并同步上课记录。</text>
						<!-- #ifdef MP-WEIXIN -->
						<button class="parent-primary-button record-auth-button" :loading="authLoading" :disabled="authLoading" open-type="getPhoneNumber" @getphonenumber="handleWechatPhoneAuth">
							授权登录
						</button>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<view class="parent-primary-button record-auth-button" @click="handleMockPhoneAuth">授权登录</view>
						<!-- #endif -->
					</view>
				</template>

				<template v-else-if="pageLoading && !recordList.length">
					<view class="parent-empty-card record-empty-card">
						<view class="parent-empty-badge">载</view>
						<text class="parent-empty-title">正在加载上课记录</text>
						<text class="parent-empty-desc">稍等一下，系统正在同步最新记录。</text>
					</view>
				</template>

				<template v-else-if="recordList.length">
					<view
						v-for="item in recordList"
						:key="item.id"
						class="record-item"
					>
						<view class="record-item__header">
							<view class="record-item__title-wrap">
								<text class="record-item__title">{{ item.className }}</text>
							</view>
							<text class="record-item__status" :class="recordStatusClass(item.statusText)">{{ item.statusText }}</text>
						</view>

						<view class="record-item__meta">
							<view class="record-item__row">
								<text class="record-item__label">上课时间：</text>
								<text class="record-item__value">{{ formatTimeRange(item) }}</text>
							</view>
							<view v-if="item.teacherName !== '-'" class="record-item__row">
								<text class="record-item__label">授课老师：</text>
								<text class="record-item__value">{{ item.teacherName }}</text>
							</view>
							<view v-if="item.classroom !== '-'" class="record-item__row">
								<text class="record-item__label">上课教室：</text>
								<text class="record-item__value">{{ item.classroom }}</text>
							</view>
							<view v-if="item.showDeductQuantity" class="record-item__row">
								<text class="record-item__label">扣除课时：</text>
								<text class="record-item__value record-item__value--strong">{{ formatMetric(item.deductQuantity) }}</text>
							</view>
							<view v-if="item.showDeductDays" class="record-item__row">
								<text class="record-item__label">扣除天数：</text>
								<text class="record-item__value record-item__value--strong">{{ formatMetric(item.deductDays) }}</text>
							</view>
						</view>
					</view>

					<view v-if="moreLoading" class="record-loading-text">正在加载更多...</view>
					<view v-else-if="!hasMore && recordList.length >= pageSize" class="record-loading-text">没有更多记录了</view>
				</template>

				<template v-else>
					<view class="parent-empty-card record-empty-card">
						<view class="parent-empty-badge">空</view>
						<text class="parent-empty-title">{{ displayStudents.length ? '当前学员还没有上课记录' : '暂未绑定学员' }}</text>
						<text class="parent-empty-desc">{{ displayStudents.length ? '完成上课后，记录会自动同步到这里。' : '请先完成学员关注，再查看上课记录。' }}</text>
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
import { onReachBottom, onShow } from '@dcloudio/uni-app'
import BindSuccessDialog from '@/components/bind-success-dialog/bind-success-dialog.vue'
import { authorizeByWechatPhone } from '@/common/parent-auth'
import { listParentClassRecords } from '@/common/parent-api'
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
const recordItems = ref([])
const hasLoadedSummary = ref(false)
const selectedStudentId = ref('')
const pageLoading = ref(false)
const moreLoading = ref(false)
const hasMore = ref(false)
const authLoading = computed(() => parentState.authLoading)
const isAuthenticated = computed(() => parentState.isAuthenticated)
const bindSuccessVisible = computed(() => parentState.bindSuccessVisible)
const latestBindStudentName = computed(() => parentState.latestBindStudentName)
const pageIndex = ref(1)
const pageSize = 20

let classRecordRequestSerial = 0

const displayStudents = computed(() => {
	const backendStudents = normalizeStudents(summaryStudents.value || [])
	if (hasLoadedSummary.value) {
		return backendStudents
	}
	return normalizeStudents(parentState.students || [])
})
const shouldCenterStudents = computed(() => displayStudents.value.length > 0 && displayStudents.value.length <= 3)

const recordList = computed(() => normalizeClassRecordList(recordItems.value || []))

onShow(() => {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		hasLoadedSummary.value = false
		summaryStudents.value = []
		resetRecordPagination()
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
	refreshClassRecords({
		reset: true
	})
})

onReachBottom(() => {
	loadMoreClassRecords()
})

function normalizeStudents(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		id: `${item?.id || item?.studentId || item?.rawId || `student-${index + 1}`}`.trim(),
		name: `${item?.name || item?.studentName || '-'}`.trim() || '-'
	})).filter(item => item.id)
}

function normalizeClassRecordList(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		id: `${item?.id || item?.studentTeachingRecordId || `record-${index + 1}`}`.trim(),
		className: `${item?.className || item?.courseName || '课程记录'}`.trim() || '课程记录',
		courseName: `${item?.courseName || item?.className || ''}`.trim(),
		date: `${item?.date || ''}`.trim(),
		startTime: `${item?.startTime || ''}`.trim(),
		endTime: `${item?.endTime || ''}`.trim(),
		teacherName: `${item?.teacherName || '-'}`.trim() || '-',
		classroom: `${item?.classroom || '-'}`.trim() || '-',
		statusText: `${item?.statusText || '-'}`.trim() || '-',
		deductQuantity: Number(item?.deductQuantity || 0),
		deductDays: Number(item?.deductDays || 0),
		showDeductQuantity: !!item?.showDeductQuantity,
		showDeductDays: !!item?.showDeductDays
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

function resetRecordPagination() {
	recordItems.value = []
	pageIndex.value = 1
	hasMore.value = false
	pageLoading.value = false
	moreLoading.value = false
}

async function refreshClassRecords(options = {}) {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	const reset = options?.reset !== false

	ensureSelectedStudent()
	const token = `${parentState.authToken || ''}`.trim()
	const targetStudentId = `${selectedStudentId.value || ''}`.trim()
	if (!token || !targetStudentId) {
		summaryStudents.value = displayStudents.value
		resetRecordPagination()
		return
	}

	const requestPageIndex = reset ? 1 : pageIndex.value + 1
	if (reset) {
		pageLoading.value = true
	} else {
		moreLoading.value = true
	}
	const requestSerial = ++classRecordRequestSerial
	try {
		const result = await listParentClassRecords(token, {
			studentId: targetStudentId,
			pageIndex: requestPageIndex,
			pageSize
		})
		if (requestSerial !== classRecordRequestSerial || token !== `${parentState.authToken || ''}`.trim()) {
			return
		}
		hasLoadedSummary.value = true
		summaryStudents.value = result?.students || []
		const nextStudents = normalizeStudents(result?.students || [])
		if (nextStudents.length && !nextStudents.some(item => item.id === selectedStudentId.value)) {
			selectedStudentId.value = nextStudents[0].id
		}
		const nextItems = result?.items || []
		recordItems.value = reset ? nextItems : mergePagedRecordItems(recordItems.value, nextItems)
		pageIndex.value = Number(result?.pageIndex || requestPageIndex)
		hasMore.value = !!result?.hasMore
	} catch (error) {
		if (requestSerial !== classRecordRequestSerial) {
			return
		}
		console.warn('load parent class records failed', error)
		if (reset) {
			uni.showToast({
				title: `${error?.message || '加载失败'}`.slice(0, 24),
				icon: 'none'
			})
		}
	} finally {
		if (requestSerial === classRecordRequestSerial && reset) {
			pageLoading.value = false
		}
		if (requestSerial === classRecordRequestSerial && !reset) {
			moreLoading.value = false
		}
	}
}

function handleStudentSelect(studentId) {
	if (!studentId || studentId === selectedStudentId.value || pageLoading.value || moreLoading.value) {
		return
	}
	selectedStudentId.value = studentId
	resetRecordPagination()
	refreshClassRecords({
		reset: true
	})
	uni.pageScrollTo({
		scrollTop: 0,
		duration: 180
	})
}

function loadMoreClassRecords() {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	if (!hasMore.value || pageLoading.value || moreLoading.value) {
		return
	}
	refreshClassRecords({
		reset: false
	})
}

function handleMockPhoneAuth() {
	setPostAuthPage('/pages/attendance-record/index')
	authorizeByPhone(DEFAULT_PHONE)
	uni.navigateTo({
		url: '/pages/select-student/index'
	})
}

function handleWechatPhoneAuth(event) {
	authorizeByWechatPhone(event, {
		postAuthPage: '/pages/attendance-record/index'
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

function recordStatusClass(statusText = '') {
	const text = `${statusText || ''}`.trim()
	if (text === '到课') {
		return 'record-item__status--arrived'
	}
	if (text === '请假') {
		return 'record-item__status--leave'
	}
	if (text === '旷课') {
		return 'record-item__status--absent'
	}
	if (text === '未记录') {
		return 'record-item__status--pending'
	}
	return ''
}

function formatMetric(value) {
	const nextValue = Number(value || 0)
	if (!Number.isFinite(nextValue)) {
		return '-'
	}
	if (Math.abs(nextValue - Math.round(nextValue)) < 0.00001) {
		return `${Math.round(nextValue)}`
	}
	return `${nextValue.toFixed(2)}`.replace(/\.?0+$/u, '')
}

function formatTimeRange(item = {}) {
	const dateText = `${item?.date || ''}`.trim()
	const startText = `${item?.startTime || ''}`.trim()
	const endText = `${item?.endTime || ''}`.trim()
	if (!dateText) {
		return `${startText} ~ ${endText}`.trim()
	}
	if (!startText && !endText) {
		return dateText
	}
	return `${dateText} ${startText} ~ ${endText}`.trim()
}

function mergePagedRecordItems(currentItems = [], nextItems = []) {
	const list = Array.isArray(currentItems) ? [...currentItems] : []
	const seen = new Set(list.map(item => `${item?.id || item?.studentTeachingRecordId || ''}`.trim()).filter(Boolean))
	;(Array.isArray(nextItems) ? nextItems : []).forEach(item => {
		const key = `${item?.id || item?.studentTeachingRecordId || ''}`.trim()
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
.record-page {
	min-height: 100vh;
	background:
		radial-gradient(circle at 12% 6%, rgba(255, 221, 169, 0.2), transparent 22%),
		linear-gradient(180deg, #fff7eb 0%, #fff8ef 42%, #fff9f2 100%);
}

.record-header {
	padding-bottom: 18rpx;
}

.record-nav-fixed {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	z-index: 40;
	background:
		linear-gradient(180deg, rgba(255, 247, 235, 0.98) 0%, rgba(255, 248, 239, 0.96) 72%, rgba(255, 249, 242, 0.9) 100%);
	backdrop-filter: blur(12rpx);
}

.record-nav-fixed__inner {
	padding: 0 24rpx 10rpx;
}

.record-nav__side {
	display: flex;
	align-items: center;
}

.record-nav__title {
	flex: 1;
	text-align: center;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.record-back {
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

.record-back__icon {
	width: 16rpx;
	height: 16rpx;
	border-left: 4rpx solid #433e33;
	border-bottom: 4rpx solid #433e33;
	transform: translateX(4rpx) rotate(45deg);
}



.record-student-scroll {
	white-space: nowrap;
}

.record-student-list {
	display: inline-flex;
	align-items: stretch;
	min-width: 100%;
}

.record-student-list--center {
	justify-content: center;
	gap: 12rpx;
}

.record-student-item {
	position: relative;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	min-width: 156rpx;
	padding: 16rpx 14rpx 14rpx;
}

.record-student-item--center {
	flex: 0 1 180rpx;
	min-width: 0;
}

.record-student-item__name {
	font-size: 30rpx;
	font-weight: 600;
	line-height: 1.24;
	color: #8f8a80;
}

.record-student-item__line {
	width: 40rpx;
	height: 8rpx;
	margin-top: 14rpx;
	border-radius: 999rpx;
	background: transparent;
}

.record-student-item--active .record-student-item__name {
	color: #222222;
	font-weight: 700;
}

.record-student-item--active .record-student-item__line {
	background: linear-gradient(135deg, #ffe40f 0%, #ffcf0b 100%);
	box-shadow: 0 8rpx 18rpx rgba(255, 214, 10, 0.28);
}

.record-list-card {
	padding: 22rpx 20rpx 26rpx;
}

.record-list-card__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.record-list-card__title {
	font-size: 32rpx;
	font-weight: 700;
}

.record-list-card__count {
	font-size: 22rpx;
	color: var(--parent-subtext);
}

.record-empty-card {
	padding-top: 54rpx;
	padding-bottom: 54rpx;
}

.record-auth-button {
	width: 100%;
	margin-top: 20rpx;
}

.record-item {
	margin-top: 16rpx;
	padding: 22rpx 20rpx;
	border-radius: 24rpx;
	background: rgba(255, 255, 255, 0.92);
	border: 1rpx solid rgba(230, 221, 204, 0.78);
}

.record-item__header {
	display: flex;
	align-items: flex-start;
	justify-content: space-between;
	gap: 16rpx;
}

.record-item__title-wrap {
	flex: 1;
	min-width: 0;
	display: flex;
	flex-direction: column;
}

.record-item__title {
	font-size: 26rpx;
	font-weight: 700;
	line-height: 1.45;
	color: #5f5f5f;
}

.record-item__status {
	flex-shrink: 0;
	padding: 10rpx 18rpx;
	border-radius: 999rpx;
	font-size: 22rpx;
	font-weight: 700;
	line-height: 1;
	background: rgba(233, 237, 243, 0.96);
	color: #7c8796;
}

.record-item__status--arrived {
	background: rgba(228, 240, 255, 0.96);
	color: #2f86ff;
}

.record-item__status--leave {
	background: rgba(255, 241, 204, 0.98);
	color: #b27a00;
}

.record-item__status--absent {
	background: rgba(255, 230, 225, 0.96);
	color: #dd5b4f;
}

.record-item__status--pending {
	background: rgba(238, 239, 241, 0.96);
	color: #8f95a0;
}

.record-item__meta {
	margin-top: 18rpx;
	display: flex;
	flex-direction: column;
	gap: 12rpx;
}

.record-item__row {
	display: flex;
	align-items: flex-start;
	gap: 0;
}

.record-item__label {
	flex-shrink: 0;
	font-size: 24rpx;
	line-height: 1.5;
	color: #968f82;
}

.record-item__value {
	flex: 1;
	font-size: 24rpx;
	line-height: 1.5;
	color: #353535;
	text-align: left;
}

.record-item__value--strong {
	font-weight: 700;
	color: #1f1f1f;
}

.record-loading-text {
	margin-top: 18rpx;
	text-align: center;
	font-size: 22rpx;
	color: var(--parent-subtext);
}
</style>
