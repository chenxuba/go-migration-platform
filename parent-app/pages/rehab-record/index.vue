<template>
	<view class="parent-page rehab-page">
		<view class="rehab-nav-fixed" :style="{ paddingTop: `${nav.top}px` }">
			<view class="rehab-nav-fixed__inner">
				<view class="parent-nav-row rehab-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="rehab-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="rehab-back" @click="goBack">
							<view class="rehab-back__icon"></view>
						</view>
					</view>
					<text class="rehab-nav__title">康复记录</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>
			</view>
		</view>

		<view class="parent-shell" :style="{ paddingTop: `${nav.top + nav.height + 18}px` }">
			<view class="parent-header rehab-header">
				<view v-if="isAuthenticated && displayStudents.length" class="parent-card rehab-student-card">
					<scroll-view class="rehab-student-scroll" :scroll-x="!shouldCenterStudents" show-scrollbar="false" enable-flex>
						<view class="rehab-student-list" :class="{ 'rehab-student-list--center': shouldCenterStudents }">
							<view
								v-for="student in displayStudents"
								:key="student.id"
								class="rehab-student-item"
								:class="{
									'rehab-student-item--active': selectedStudentId === student.id,
									'rehab-student-item--center': shouldCenterStudents
								}"
								@click="handleStudentSelect(student.id)"
							>
								<text class="rehab-student-item__name">{{ student.name }}</text>
								<view class="rehab-student-item__line"></view>
							</view>
						</view>
					</scroll-view>
				</view>
			</view>

			<view class="parent-card rehab-list-card">
				<template v-if="!isAuthenticated">
					<view class="parent-card rehab-empty-panel">
						<view class="parent-empty-card rehab-empty-card">
							<view class="parent-empty-badge">康</view>
							<text class="parent-empty-title">登录后即可查看康复记录</text>
							<text class="parent-empty-desc">手机号授权后，系统会自动同步当前学员的已记录康复内容。</text>
							<!-- #ifdef MP-WEIXIN -->
							<button class="parent-primary-button rehab-auth-button" :loading="authLoading" :disabled="authLoading" open-type="getPhoneNumber" @getphonenumber="handleWechatPhoneAuth">
								授权登录
							</button>
							<!-- #endif -->
							<!-- #ifndef MP-WEIXIN -->
							<view class="parent-primary-button rehab-auth-button" @click="handleMockPhoneAuth">授权登录</view>
							<!-- #endif -->
						</view>
					</view>
				</template>

				<template v-else-if="pageLoading && !rehabList.length">
					<view class="parent-card rehab-empty-panel">
						<view class="parent-empty-card rehab-empty-card">
							<view class="parent-empty-badge">载</view>
							<text class="parent-empty-title">正在加载康复记录</text>
							<text class="parent-empty-desc">稍等一下，系统正在同步最新记录。</text>
						</view>
					</view>
				</template>

				<template v-else-if="rehabList.length">
					<view
						v-for="item in rehabList"
						:key="item.id"
						class="rehab-item"
						@click="openRehabDetail(item)"
					>
						<view class="rehab-item__header">
							<view class="rehab-item__header-main">
								<text class="rehab-item__title">{{ item.className }}</text>
								<text v-if="item.courseName !== '-'" class="rehab-item__subtitle">{{ item.courseName }}</text>
							</view>
							<text class="rehab-item__status">{{ item.statusText }}</text>
						</view>

						<view class="rehab-item__meta">
							<view class="rehab-item__meta-fact">
								<text class="rehab-item__meta-label">上课时间</text>
								<text class="rehab-item__meta-value">{{ item.lessonTime }}</text>
							</view>
							<view v-if="item.teacherName !== '-'" class="rehab-item__meta-fact">
								<text class="rehab-item__meta-label">授课老师</text>
								<text class="rehab-item__meta-value">{{ item.teacherName }}</text>
							</view>
						</view>

						<view class="rehab-item__content">
							<view v-if="item.trainingTarget !== '-'" class="rehab-item__row">
								<text class="rehab-item__label">训练目标：</text>
								<text class="rehab-item__value">{{ item.trainingTarget }}</text>
							</view>
							<view v-if="item.performance !== '-'" class="rehab-item__row">
								<text class="rehab-item__label">课堂表现：</text>
								<text class="rehab-item__value">{{ item.performance }}</text>
							</view>
							<view v-if="item.suggestion !== '-'" class="rehab-item__row">
								<text class="rehab-item__label">康复建议：</text>
								<text class="rehab-item__value">{{ item.suggestion }}</text>
							</view>
							<view v-if="item.showSummaryRow" class="rehab-item__row">
								<text class="rehab-item__label">记录摘要：</text>
								<text class="rehab-item__value">{{ item.summaryText }}</text>
							</view>
						</view>

						<view class="rehab-item__footer">
							<view class="rehab-item__footer-main">
								<text class="rehab-item__footer-text">记录时间：{{ item.updatedTime }}</text>
								<text v-if="item.updatedStaffName !== '-'" class="rehab-item__footer-text">记录老师：{{ item.updatedStaffName }}</text>
							</view>
							<view class="rehab-item__footer-action">
								<text class="rehab-item__footer-action-text">详情</text>
								<view class="rehab-item__arrow"></view>
							</view>
						</view>
					</view>

					<view v-if="moreLoading" class="rehab-loading-text">正在加载更多...</view>
					<view v-else-if="!hasMore && rehabList.length >= pageSize" class="rehab-loading-text">没有更多记录了</view>
				</template>

				<template v-else>
					<view class="parent-card rehab-empty-panel">
						<view class="parent-empty-card rehab-empty-card">
							<view class="parent-empty-badge">空</view>
							<text class="parent-empty-title">{{ displayStudents.length ? '当前学员还没有康复记录' : '暂未绑定学员' }}</text>
							<text class="parent-empty-desc">{{ displayStudents.length ? '老师完成记录后，康复内容会自动同步到这里。' : '请先完成学员关注，再查看康复记录。' }}</text>
						</view>
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
import { listParentRehabRecords } from '@/common/parent-api'
import { getNavLayout } from '@/common/nav-layout'
import { DEFAULT_PHONE } from '@/common/mock-parent'
import {
	authorizeByPhone,
	dismissBindSuccess,
	parentState,
	setPostAuthPage
} from '@/common/parent-state'

const nav = getNavLayout()
const routeStudentId = ref('')
const summaryStudents = ref([])
const rehabItems = ref([])
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

let rehabRequestSerial = 0

const displayStudents = computed(() => {
	const backendStudents = normalizeStudents(summaryStudents.value || [])
	if (hasLoadedSummary.value) {
		return backendStudents
	}
	return normalizeStudents(parentState.students || [])
})
const shouldCenterStudents = computed(() => displayStudents.value.length > 0 && displayStudents.value.length <= 3)
const rehabList = computed(() => normalizeRehabList(rehabItems.value || []))

onLoad(query => {
	routeStudentId.value = `${query?.studentId || ''}`.trim()
})

onShow(() => {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		hasLoadedSummary.value = false
		summaryStudents.value = []
		resetRehabPagination()
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
	refreshRehabRecords({
		reset: true
	})
})

onReachBottom(() => {
	loadMoreRehabRecords()
})

function normalizeStudents(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		id: `${item?.id || item?.studentId || item?.rawId || `student-${index + 1}`}`.trim(),
		name: `${item?.name || item?.studentName || '-'}`.trim() || '-'
	})).filter(item => item.id)
}

function normalizeRehabList(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => {
		const trainingTarget = `${item?.trainingTarget || ''}`.trim()
		const performance = `${item?.performance || ''}`.trim()
		const suggestion = `${item?.suggestion || ''}`.trim()
		const summaryText = `${item?.summaryText || '-'}`.trim() || '-'
		const className = `${item?.className || item?.courseName || '康复记录'}`.trim() || '康复记录'
		const courseName = `${item?.courseName || item?.className || '-'}`.trim() || '-'
		return {
			id: `${item?.id || item?.studentTeachingRecordId || `rehab-${index + 1}`}`.trim(),
			studentTeachingRecordId: `${item?.studentTeachingRecordId || item?.id || ''}`.trim(),
			className,
			courseName: courseName === className ? '-' : courseName,
			lessonTime: `${item?.lessonTime || ''}`.trim() || buildLessonTimeText(item),
			teacherName: `${item?.teacherName || '-'}`.trim() || '-',
			summaryText,
			trainingTarget: trainingTarget || '-',
			performance: performance || '-',
			suggestion: suggestion || '-',
			updatedTime: `${item?.updatedTime || '-'}`.trim() || '-',
			updatedStaffName: `${item?.updatedStaffName || '-'}`.trim() || '-',
			statusText: `${item?.statusText || '已记录'}`.trim() || '已记录',
			showSummaryRow: !trainingTarget && !performance && !suggestion && summaryText !== '-'
		}
	})
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

function ensureSelectedStudent() {
	const students = displayStudents.value
	if (!students.length) {
		selectedStudentId.value = ''
		return
	}
	const candidateId = `${selectedStudentId.value || routeStudentId.value || parentState.currentStudentId || ''}`.trim()
	const matched = students.find(item => item.id === candidateId)
	selectedStudentId.value = matched?.id || students[0].id
}

function resetRehabPagination() {
	rehabItems.value = []
	pageIndex.value = 1
	hasMore.value = false
	pageLoading.value = false
	moreLoading.value = false
}

async function refreshRehabRecords(options = {}) {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	const reset = options?.reset !== false

	ensureSelectedStudent()
	const token = `${parentState.authToken || ''}`.trim()
	const targetStudentId = `${selectedStudentId.value || ''}`.trim()
	if (!token || !targetStudentId) {
		summaryStudents.value = displayStudents.value
		resetRehabPagination()
		return
	}

	const requestPageIndex = reset ? 1 : pageIndex.value + 1
	if (reset) {
		pageLoading.value = true
	} else {
		moreLoading.value = true
	}
	const requestSerial = ++rehabRequestSerial
	try {
		const result = await listParentRehabRecords(token, {
			studentId: targetStudentId,
			pageIndex: requestPageIndex,
			pageSize
		})
		if (requestSerial !== rehabRequestSerial || token !== `${parentState.authToken || ''}`.trim()) {
			return
		}
		hasLoadedSummary.value = true
		summaryStudents.value = result?.students || []
		const nextStudents = normalizeStudents(result?.students || [])
		if (nextStudents.length && !nextStudents.some(item => item.id === selectedStudentId.value)) {
			selectedStudentId.value = nextStudents[0].id
		}
		const nextItems = result?.items || []
		rehabItems.value = reset ? nextItems : mergePagedRehabItems(rehabItems.value, nextItems)
		pageIndex.value = Number(result?.pageIndex || requestPageIndex)
		hasMore.value = !!result?.hasMore
	} catch (error) {
		if (requestSerial !== rehabRequestSerial) {
			return
		}
		console.warn('load parent rehab records failed', error)
		if (reset) {
			uni.showToast({
				title: `${error?.message || '加载失败'}`.slice(0, 24),
				icon: 'none'
			})
		}
	} finally {
		if (requestSerial === rehabRequestSerial && reset) {
			pageLoading.value = false
		}
		if (requestSerial === rehabRequestSerial && !reset) {
			moreLoading.value = false
		}
	}
}

function handleStudentSelect(studentId) {
	if (!studentId || studentId === selectedStudentId.value || pageLoading.value || moreLoading.value) {
		return
	}
	selectedStudentId.value = studentId
	resetRehabPagination()
	refreshRehabRecords({
		reset: true
	})
	uni.pageScrollTo({
		scrollTop: 0,
		duration: 180
	})
}

function loadMoreRehabRecords() {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	if (!hasMore.value || pageLoading.value || moreLoading.value) {
		return
	}
	refreshRehabRecords({
		reset: false
	})
}

function buildCurrentPageURL() {
	const studentID = `${selectedStudentId.value || routeStudentId.value || ''}`.trim()
	if (!studentID) {
		return '/pages/rehab-record/index'
	}
	return `/pages/rehab-record/index?studentId=${encodeURIComponent(studentID)}`
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

function openRehabDetail(item) {
	const recordID = `${item?.studentTeachingRecordId || item?.id || ''}`.trim()
	const studentID = `${selectedStudentId.value || ''}`.trim()
	if (!recordID) {
		return
	}
	const query = [`studentTeachingRecordId=${encodeURIComponent(recordID)}`]
	if (studentID) {
		query.push(`studentId=${encodeURIComponent(studentID)}`)
	}
	uni.navigateTo({
		url: `/pages/rehab-record/detail?${query.join('&')}`
	})
}

function mergePagedRehabItems(currentItems = [], nextItems = []) {
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
.rehab-page {
	min-height: 100vh;
	background:
		radial-gradient(circle at 10% 4%, rgba(187, 226, 181, 0.18), transparent 20%),
		radial-gradient(circle at 100% 0%, rgba(255, 228, 174, 0.26), transparent 26%),
		linear-gradient(180deg, #fff7ec 0%, #fff9f2 42%, #fffdf8 100%);
}

.rehab-header {
	padding-bottom: 18rpx;
}

.rehab-nav-fixed {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	z-index: 40;
	background:
		linear-gradient(180deg, rgba(255, 247, 236, 0.98) 0%, rgba(255, 249, 242, 0.96) 72%, rgba(255, 253, 248, 0.9) 100%);
	backdrop-filter: blur(12rpx);
}

.rehab-nav-fixed__inner {
	padding: 0 24rpx 10rpx;
}

.rehab-nav__side {
	display: flex;
	align-items: center;
}

.rehab-nav__title {
	flex: 1;
	text-align: center;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.rehab-back {
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

.rehab-back__icon {
	width: 16rpx;
	height: 16rpx;
	border-left: 4rpx solid #433e33;
	border-bottom: 4rpx solid #433e33;
	transform: translateX(4rpx) rotate(45deg);
}

.rehab-student-card {
	padding: 8rpx 14rpx 4rpx;
}

.rehab-student-scroll {
	white-space: nowrap;
}

.rehab-student-list {
	display: inline-flex;
	align-items: stretch;
	min-width: 100%;
}

.rehab-student-list--center {
	justify-content: center;
	gap: 12rpx;
}

.rehab-student-item {
	position: relative;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
	min-width: 156rpx;
	padding: 16rpx 14rpx 14rpx;
}

.rehab-student-item--center {
	flex: 0 1 180rpx;
	min-width: 0;
}

.rehab-student-item__name {
	font-size: 30rpx;
	font-weight: 600;
	line-height: 1.24;
	color: #918978;
}

.rehab-student-item__line {
	margin-top: 12rpx;
	width: 44rpx;
	height: 6rpx;
	border-radius: 999rpx;
	background: transparent;
}

.rehab-student-item--active .rehab-student-item__name {
	color: #1f1f1f;
}

.rehab-student-item--active .rehab-student-item__line {
	background: linear-gradient(90deg, #ffd40b 0%, #f1b800 100%);
}

.rehab-list-card {
	padding: 6rpx 0 10rpx;
	background: transparent;
	border: none;
	box-shadow: none;
	backdrop-filter: none;
}

.rehab-empty-panel {
	padding: 0;
}

.rehab-empty-card {
	padding-top: 56rpx;
	padding-bottom: 62rpx;
}

.rehab-auth-button {
	width: 100%;
	margin-top: 34rpx;
}

.rehab-item {
	margin-top: 18rpx;
	padding: 28rpx 24rpx 26rpx;
	border-radius: 30rpx;
	background: rgba(255, 255, 255, 0.96);
	border: 1rpx solid rgba(245, 237, 225, 0.96);
	box-shadow: 0 18rpx 44rpx rgba(154, 127, 77, 0.08);
}

.rehab-item:first-child {
	margin-top: 0;
}

.rehab-item__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 20rpx;
}

.rehab-item__header-main {
	flex: 1;
	min-width: 0;
}

.rehab-item__title {
	display: block;
	font-size: 32rpx;
	font-weight: 700;
	line-height: 1.42;
	color: #2a2721;
}

.rehab-item__subtitle {
	display: block;
	margin-top: 8rpx;
	font-size: 24rpx;
	line-height: 1.45;
	color: #9f9789;
}

.rehab-item__status {
	flex-shrink: 0;
	padding: 10rpx 16rpx;
	border-radius: 999rpx;
	background: rgba(233, 246, 237, 0.82);
	border: 1rpx solid rgba(175, 219, 188, 0.58);
	color: #2c9a67;
	font-size: 22rpx;
	font-weight: 700;
	line-height: 1;
}

.rehab-item__meta {
	display: flex;
	flex-wrap: nowrap;
	gap: 12rpx;
	margin-top: 20rpx;
}

.rehab-item__meta-fact {
	flex: 1 1 0;
	min-width: 0;
	padding: 16rpx 18rpx;
	border-radius: 22rpx;
	background: linear-gradient(180deg, rgba(252, 248, 241, 0.98) 0%, rgba(255, 253, 248, 0.96) 100%);
	border: 1rpx solid rgba(244, 237, 224, 0.92);
}

.rehab-item__meta-label {
	display: block;
	font-size: 22rpx;
	line-height: 1.2;
	color: #afa18d;
}

.rehab-item__meta-value {
	display: block;
	margin-top: 10rpx;
	font-size: 23rpx;
	font-weight: 600;
	line-height: 1.3;
	color: #4a463d;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.rehab-item__content {
	margin-top: 20rpx;
	padding-top: 18rpx;
	border-top: 1rpx solid var(--parent-divider);
}

.rehab-item__row {
	display: flex;
	align-items: flex-start;
	padding: 10rpx 0;
	gap: 10rpx;
}

.rehab-item__label {
	flex-shrink: 0;
	width: 124rpx;
	font-size: 24rpx;
	font-weight: 600;
	line-height: 1.6;
	color: #8e8575;
}

.rehab-item__value {
	flex: 1;
	min-width: 0;
	font-size: 25rpx;
	line-height: 1.68;
	color: #37332d;
	display: -webkit-box;
	-webkit-box-orient: vertical;
	-webkit-line-clamp: 2;
	overflow: hidden;
}

.rehab-item__footer {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 16rpx;
	margin-top: 20rpx;
	padding-top: 18rpx;
	border-top: 1rpx solid var(--parent-divider);
}

.rehab-item__footer-main {
	flex: 1;
	min-width: 0;
	display: flex;
	align-items: center;
	gap: 18rpx;
	overflow: hidden;
}

.rehab-item__footer-text {
	flex-shrink: 0;
	font-size: 22rpx;
	line-height: 1.45;
	color: #aba391;
	white-space: nowrap;
}

.rehab-item__footer-action {
	display: inline-flex;
	align-items: center;
	gap: 10rpx;
	flex-shrink: 0;
	margin-left: auto;
}

.rehab-item__footer-action-text {
	font-size: 22rpx;
	font-weight: 600;
	line-height: 1;
	color: #b5ac9a;
}

.rehab-item__arrow {
	width: 16rpx;
	height: 16rpx;
	border-top: 3rpx solid #c7beae;
	border-right: 3rpx solid #c7beae;
	transform: rotate(45deg);
	flex-shrink: 0;
}

.rehab-loading-text {
	padding: 30rpx 0 18rpx;
	text-align: center;
	font-size: 22rpx;
	color: var(--parent-subtext);
}
</style>
