<template>
	<view class="parent-page">
		<view class="parent-shell">
			<view class="parent-header" :style="{ paddingTop: `${nav.top}px` }">
				<view class="parent-nav-row schedule-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px` }"></view>
					<text class="schedule-nav__title">课表</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>

				<view class="parent-card schedule-calendar-card">
					<view
						id="schedule-week-viewport"
						class="schedule-week-viewport"
						@touchstart="handleWeekTouchStart"
						@touchmove.stop="handleWeekTouchMove"
						@touchend="handleWeekTouchEnd"
						@touchcancel="handleWeekTouchCancel"
					>
						<view class="schedule-week-track" :style="weekTrackStyle">
							<view
								v-for="slide in weekSlides"
								:key="slide.key"
								class="schedule-week-panel"
								:style="weekPanelStyle"
							>
								<view class="schedule-week-row">
									<view
										v-for="day in slide.days"
										:key="day.date"
										class="schedule-week-item"
										:class="{ 'schedule-week-item--active': day.isActive }"
										@tap="handleDaySelect(day.date)"
									>
										<text class="schedule-week-item__label">{{ day.weekLabel }}</text>
										<text class="schedule-week-item__day">{{ day.day }}</text>
										<view class="schedule-week-item__indicator">
											<view
												v-if="day.hasCourse"
												class="schedule-week-item__dot"
												:class="{ 'schedule-week-item__dot--active': day.isActive }"
											></view>
										</view>
									</view>
								</view>
							</view>
						</view>
					</view>
					<view class="schedule-date-text">{{ selectedDateText }}</view>
				</view>
			</view>

			<view class="parent-card schedule-list-card">
				<text class="schedule-list-card__title">课程安排</text>

				<template v-if="!isAuthenticated">
					<view class="parent-empty-card schedule-empty-card">
						<view class="parent-empty-badge">课</view>
						<text class="parent-empty-title">登录后即可查看课程安排</text>
						<text class="parent-empty-desc">手机号授权后，系统会自动帮你匹配孩子并同步课表。</text>
						<!-- #ifdef MP-WEIXIN -->
						<button class="parent-primary-button schedule-auth-button" open-type="getPhoneNumber" @getphonenumber="handleWechatPhoneAuth">
							授权登录
						</button>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<view class="parent-primary-button schedule-auth-button" @click="handleMockPhoneAuth">授权登录</view>
						<!-- #endif -->
					</view>
				</template>

				<template v-else-if="displaySchedules.length">
					<view
						v-for="item in displaySchedules"
						:key="item.id"
						class="schedule-item"
					>
						<view class="schedule-item__top">
							<text class="schedule-item__time">{{ item.startTime }} ~ {{ item.endTime }}</text>
							<text class="schedule-item__status">{{ item.statusText }}</text>
						</view>
						<view class="schedule-item__body">
							<view class="schedule-item__avatar">{{ item.studentName.slice(0, 1) }}</view>
							<view class="schedule-item__content">
								<text class="schedule-item__student">{{ item.studentName }}【{{ item.className }}】</text>
								<text class="schedule-item__detail">课程：{{ item.courseName }}</text>
								<text class="schedule-item__detail">校区：{{ campusName }}</text>
								<text class="schedule-item__detail">教室：{{ item.classroom }}</text>
								<text class="schedule-item__detail">教师：{{ item.teacherName }}</text>
								<text class="schedule-item__detail">备注：{{ item.note }}</text>
							</view>
						</view>
					</view>
				</template>

				<template v-else>
					<view class="parent-empty-card schedule-empty-card">
						<view class="parent-empty-badge">空</view>
						<text class="parent-empty-title">当前日期暂无课程安排</text>
						<text class="parent-empty-desc">可以切换日期或校区查看其它课程。</text>
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
import { computed, nextTick, ref } from 'vue'
import { onReady, onUnload } from '@dcloudio/uni-app'
import BindSuccessDialog from '@/components/bind-success-dialog/bind-success-dialog.vue'
import { getNavLayout } from '@/common/nav-layout'
import { BASE_DATE, DEFAULT_PHONE } from '@/common/mock-parent'
import {
	authorizeByPhone,
	dismissBindSuccess,
	getCurrentCampus,
	parentState,
	setPostAuthPage
} from '@/common/parent-state'

const nav = getNavLayout()
const systemInfo = uni.getSystemInfoSync()
const selectedDate = ref(BASE_DATE)
const currentWeekStart = ref(getWeekStartText(BASE_DATE))
const weekViewportWidth = ref(Math.max((systemInfo.windowWidth || 375) - 32, 280))
const dragOffsetX = ref(0)
const isDraggingWeek = ref(false)
const isWeekAnimating = ref(false)
const isAuthenticated = computed(() => parentState.isAuthenticated)
const campusName = computed(() => getCurrentCampus().name)
const bindSuccessVisible = computed(() => parentState.bindSuccessVisible)
const latestBindStudentName = computed(() => parentState.latestBindStudentName)
const weekNames = ['一', '二', '三', '四', '五', '六', '日']
const WEEK_ANIMATION_DURATION = 240
const WEEK_TAP_SLOP = 8
let weekTouchStartX = 0
let weekAnimationTimer = null
let weekMeasureTimer = null

function parseDate(dateText) {
	const [year, month, day] = dateText.split('-').map(Number)
	return new Date(year, month - 1, day)
}

function formatDate(date) {
	const year = date.getFullYear()
	const month = `${date.getMonth() + 1}`.padStart(2, '0')
	const day = `${date.getDate()}`.padStart(2, '0')
	return `${year}-${month}-${day}`
}

function addDays(date, offset) {
	const nextDate = new Date(date)
	nextDate.setDate(nextDate.getDate() + offset)
	return nextDate
}

function getWeekMonday(dateText) {
	const currentDate = parseDate(dateText)
	const dayIndex = currentDate.getDay() || 7
	return addDays(currentDate, 1 - dayIndex)
}

function getWeekStartText(dateText) {
	return formatDate(getWeekMonday(dateText))
}

const campusScheduleDateSet = computed(() => {
	if (!parentState.isAuthenticated) {
		return new Set()
	}
	return new Set(
		parentState.scheduleEntries
			.filter(item => item.campusId === parentState.currentCampusId)
			.map(item => item.date)
	)
})

function createWeekDays(weekMondayDate) {
	return Array.from({ length: 7 }).map((_, index) => {
		const date = addDays(weekMondayDate, index)
		const dateText = formatDate(date)
		return {
			date: dateText,
			weekLabel: weekNames[index],
			day: `${date.getDate()}`.padStart(2, '0'),
			isActive: dateText === selectedDate.value,
			hasCourse: campusScheduleDateSet.value.has(dateText)
		}
	})
}

const weekSlides = computed(() => {
	return [-1, 0, 1].map(offset => {
		const weekMonday = addDays(parseDate(currentWeekStart.value), offset * 7)
		return {
			key: formatDate(weekMonday),
			days: createWeekDays(weekMonday)
		}
	})
})

const weekPanelStyle = computed(() => ({
	width: `${weekViewportWidth.value}px`
}))

const weekTrackStyle = computed(() => ({
	width: `${weekViewportWidth.value * weekSlides.value.length}px`,
	transform: `translate3d(${dragOffsetX.value - weekViewportWidth.value}px, 0, 0)`,
	transition: isDraggingWeek.value || !isWeekAnimating.value
		? 'none'
		: `transform ${WEEK_ANIMATION_DURATION}ms cubic-bezier(0.22, 1, 0.36, 1)`
}))

const selectedDateText = computed(() => {
	const currentDate = parseDate(selectedDate.value)
	const weekIndex = currentDate.getDay() === 0 ? 6 : currentDate.getDay() - 1
	return `${selectedDate.value.slice(5)}（周${weekNames[weekIndex]}）`
})

const displaySchedules = computed(() => {
	return parentState.scheduleEntries
		.filter(item => item.date === selectedDate.value && item.campusId === parentState.currentCampusId)
		.sort((prev, next) => prev.startTime.localeCompare(next.startTime))
})

function completeMockAuth() {
	setPostAuthPage('/pages/schedule/index')
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

function measureWeekViewport() {
	nextTick(() => {
		uni.createSelectorQuery()
			.select('#schedule-week-viewport')
			.boundingClientRect(rect => {
				if (rect?.width) {
					weekViewportWidth.value = rect.width
				}
			})
			.exec()
	})
}

function clearWeekAnimationTimer() {
	if (weekAnimationTimer) {
		clearTimeout(weekAnimationTimer)
		weekAnimationTimer = null
	}
}

function runWeekAnimation(targetOffsetX, weekOffset = 0, nextSelectedDate = '') {
	clearWeekAnimationTimer()
	isWeekAnimating.value = true
	if (nextSelectedDate) {
		selectedDate.value = nextSelectedDate
	}
	dragOffsetX.value = targetOffsetX
	weekAnimationTimer = setTimeout(() => {
		if (weekOffset) {
			const nextWeekStartDate = addDays(parseDate(currentWeekStart.value), weekOffset * 7)
			const nextWeekStart = formatDate(nextWeekStartDate)
			currentWeekStart.value = nextWeekStart
			selectedDate.value = nextSelectedDate || nextWeekStart
		}
		isWeekAnimating.value = false
		dragOffsetX.value = 0
		weekAnimationTimer = null
	}, WEEK_ANIMATION_DURATION)
}

function handleDaySelect(dateText) {
	if (isWeekAnimating.value || Math.abs(dragOffsetX.value) > WEEK_TAP_SLOP) {
		return
	}
	selectedDate.value = dateText
	const nextWeekStart = getWeekStartText(dateText)
	if (nextWeekStart !== currentWeekStart.value) {
		currentWeekStart.value = nextWeekStart
		dragOffsetX.value = 0
	}
}

function handleWeekTouchStart(event) {
	if (isWeekAnimating.value) {
		return
	}
	const touch = event.touches?.[0]
	if (!touch) {
		return
	}
	clearWeekAnimationTimer()
	isDraggingWeek.value = true
	dragOffsetX.value = 0
	weekTouchStartX = touch.pageX
}

function handleWeekTouchMove(event) {
	if (!isDraggingWeek.value || isWeekAnimating.value) {
		return
	}
	const touch = event.touches?.[0]
	if (!touch) {
		return
	}
	const nextOffset = touch.pageX - weekTouchStartX
	dragOffsetX.value = Math.max(-weekViewportWidth.value, Math.min(weekViewportWidth.value, nextOffset))
}

function handleWeekTouchEnd() {
	if (!isDraggingWeek.value) {
		return
	}
	isDraggingWeek.value = false
	const absoluteOffset = Math.abs(dragOffsetX.value)
	if (absoluteOffset <= WEEK_TAP_SLOP) {
		dragOffsetX.value = 0
		return
	}
	const threshold = Math.min(Math.max(weekViewportWidth.value * 0.18, 44), 96)
	if (dragOffsetX.value <= -threshold) {
		const nextSelectedDate = formatDate(addDays(parseDate(currentWeekStart.value), 7))
		runWeekAnimation(-weekViewportWidth.value, 1, nextSelectedDate)
		return
	}
	if (dragOffsetX.value >= threshold) {
		const nextSelectedDate = formatDate(addDays(parseDate(currentWeekStart.value), -7))
		runWeekAnimation(weekViewportWidth.value, -1, nextSelectedDate)
		return
	}
	runWeekAnimation(0)
}

function handleWeekTouchCancel() {
	if (!isDraggingWeek.value) {
		return
	}
	isDraggingWeek.value = false
	runWeekAnimation(0)
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

onReady(() => {
	measureWeekViewport()
	weekMeasureTimer = setTimeout(measureWeekViewport, 80)
})

onUnload(() => {
	clearWeekAnimationTimer()
	if (weekMeasureTimer) {
		clearTimeout(weekMeasureTimer)
		weekMeasureTimer = null
	}
})
</script>

<style scoped>
.schedule-nav__title {
	flex: 1;
	text-align: center;
	font-size: 38rpx;
	font-weight: 700;
	line-height: 1;
}

.schedule-calendar-card {
	margin-top: 18rpx;
	padding: 18rpx 14rpx 16rpx;
}

.schedule-week-viewport {
	height: 132rpx;
	overflow: hidden;
}

.schedule-week-track {
	display: flex;
	height: 100%;
	will-change: transform;
}

.schedule-week-panel {
	flex-shrink: 0;
	padding: 0 2rpx;
}

.schedule-week-row {
	display: grid;
	grid-template-columns: repeat(7, 1fr);
	gap: 8rpx;
}

.schedule-week-item {
	padding: 8rpx 0 10rpx;
	border-radius: 24rpx;
	display: flex;
	flex-direction: column;
	align-items: center;
	justify-content: center;
}

.schedule-week-item--active {
	background: linear-gradient(135deg, #ffe40f 0%, #ffcf0b 100%);
	box-shadow: 0 12rpx 24rpx rgba(255, 214, 10, 0.24);
}

.schedule-week-item__label {
	font-size: 18rpx;
	color: #615b4f;
}

.schedule-week-item__day {
	margin-top: 8rpx;
	font-size: 32rpx;
	font-weight: 700;
	color: #272727;
}

.schedule-week-item__indicator {
	height: 16rpx;
	margin-top: 8rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.schedule-week-item__dot {
	width: 8rpx;
	height: 8rpx;
	border-radius: 50%;
	background: #ff6a5b;
	box-shadow: 0 4rpx 10rpx rgba(255, 106, 91, 0.32);
}

.schedule-week-item__dot--active {
	box-shadow:
		0 0 0 4rpx rgba(255, 255, 255, 0.72),
		0 4rpx 10rpx rgba(255, 106, 91, 0.28);
}

.schedule-date-text {
	margin-top: -6px;
	text-align: center;
	font-size: 22rpx;
	font-weight: 600;
	color: #4f4a40;
}

.schedule-list-card {
	padding: 20rpx;
}

.schedule-list-card__title {
	font-size: 36rpx;
	font-weight: 700;
}

.schedule-empty-card {
	padding-top: 50rpx;
	padding-bottom: 54rpx;
}

.schedule-auth-button {
	width: 100%;
	margin-top: 20rpx;
}

.schedule-item {
	margin-top: 16rpx;
	padding: 20rpx;
	border-radius: 24rpx;
	background: rgba(255, 255, 255, 0.88);
	border: 1rpx solid rgba(230, 221, 204, 0.78);
}

.schedule-item__top {
	display: flex;
	align-items: center;
	justify-content: space-between;
}

.schedule-item__time {
	font-size: 32rpx;
	font-weight: 700;
	color: #808080;
}

.schedule-item__status {
	padding: 8rpx 14rpx;
	border-radius: 999rpx;
	background: rgba(243, 243, 243, 0.95);
	font-size: 20rpx;
	color: #999999;
}

.schedule-item__body {
	display: flex;
	gap: 16rpx;
	margin-top: 18rpx;
}

.schedule-item__avatar {
	width: 72rpx;
	height: 72rpx;
	border-radius: 50%;
	background: #a9c8ff;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	font-size: 28rpx;
	font-weight: 700;
	flex-shrink: 0;
}

.schedule-item__content {
	display: flex;
	flex-direction: column;
}

.schedule-item__student {
	font-size: 26rpx;
	font-weight: 700;
	line-height: 1.45;
	color: #5f5f5f;
}

.schedule-item__detail {
	margin-top: 6rpx;
	font-size: 22rpx;
	line-height: 1.45;
	color: #868686;
}
</style>
