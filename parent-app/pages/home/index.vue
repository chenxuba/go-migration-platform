<template>
	<view class="parent-page">
		<view class="parent-shell" :class="{ 'parent-shell--home-guide': showOfficialGuide }">
			<view class="parent-header" :style="{ paddingTop: `${nav.top}px` }">
				<view class="home-top-row" :style="{ minHeight: `${nav.height}px` }">
					<view v-if="canSwitchCampus" class="home-campus-trigger" @click="goCampusPage">
						<text class="home-campus-trigger__icon">切</text>
						<text class="home-campus-trigger__text">切换校区</text>
					</view>
				</view>
				<view class="parent-card home-campus-card">
					<view class="home-campus-brand">
						<view class="home-campus-logo">
							<image v-if="currentCampus.logoUrl" class="home-campus-logo__image" :src="currentCampus.logoUrl" mode="aspectFill"></image>
							<text v-else class="home-campus-logo__text">{{ campusLogoText }}</text>
						</view>
						<view class="home-campus-brand__content">
							<text class="home-campus-brand__name">{{ campusTitle }}</text>
							<text v-if="campusSubtitle" class="home-campus-brand__campus">{{ campusSubtitle }}</text>
						</view>
					</view>
				</view>
			</view>

			<view class="parent-card home-grid-card">
				<view class="home-grid">
					<view
						v-for="item in featureList"
						:key="item.key"
						class="home-grid-item"
						@click="handleFeature(item)"
					>
						<view class="home-grid-item__badge" :style="{ background: item.accent }">
							{{ item.shortLabel }}
						</view>
						<text class="home-grid-item__text">{{ item.title }}</text>
					</view>
				</view>
			</view>

			<view class="parent-card home-notice-card">
				<view class="home-notice-card__header">
					<text class="home-notice-card__title">评估报告</text>
					<text class="home-notice-card__more">全部</text>
				</view>
				<view
					v-for="item in reportList"
					:key="item.id"
					class="home-notice-item"
					@click="handleReport(item)"
				>
					<view class="home-notice-item__dot"></view>
					<view class="home-notice-item__content">
						<text class="home-notice-item__title">{{ item.title }}</text>
						<text class="home-notice-item__desc">{{ item.summary }}</text>
					</view>
				</view>
			</view>
		</view>

		<view v-if="showOfficialGuide" class="home-follow-mask" @touchmove.stop.prevent>
			<view class="home-follow-dialog" @click.stop>
				<view class="home-follow-dialog__close" @click="closeOfficialGuide">×</view>
				<text class="home-follow-dialog__title">{{ officialGuideTitle }}</text>
				<text class="home-follow-dialog__desc">{{ officialGuideDescPrimary }}</text>
				<text class="home-follow-dialog__desc">{{ officialGuideDescSecondary }}</text>
				<text v-if="officialGuideDescTertiary" class="home-follow-dialog__desc">{{ officialGuideDescTertiary }}</text>

				<view v-if="showOfficialBindAction" class="home-follow-dialog__actions">
					<!-- #ifdef MP-WEIXIN -->
					<button
						v-if="!isAuthenticated"
						class="parent-primary-button home-follow-dialog__action"
						:loading="authLoading"
						:disabled="authLoading"
						open-type="getPhoneNumber"
						@getphonenumber="handleWechatOfficialAuth"
					>
						授权并绑定
					</button>
					<!-- #endif -->
					<!-- #ifndef MP-WEIXIN -->
					<view v-if="!isAuthenticated" class="parent-primary-button home-follow-dialog__action" @click="handleMockOfficialAuth">授权并绑定</view>
					<!-- #endif -->
					<view
						v-else
						class="parent-primary-button home-follow-dialog__action"
						@click="goOfficialBindingSelection"
					>
						去选择学员
					</view>
				</view>
			</view>
		</view>

		<view v-if="showOfficialGuide" class="home-follow-bar">
			<view class="home-follow-bar__panel">
				<view class="home-follow-bar__icon">{{ officialBarIconText }}</view>
				<view class="home-follow-bar__copy">
					<text class="home-follow-bar__title">{{ officialBarTitle }}</text>
					<text class="home-follow-bar__desc">{{ officialBarDesc }}</text>
				</view>
				<view v-if="showOfficialBindAction" class="home-follow-bar__action">
					<!-- #ifdef MP-WEIXIN -->
					<button
						v-if="!isAuthenticated"
						class="parent-primary-button home-follow-bar__button"
						:loading="authLoading"
						:disabled="authLoading"
						open-type="getPhoneNumber"
						@getphonenumber="handleWechatOfficialAuth"
					>
						绑定
					</button>
					<!-- #endif -->
					<!-- #ifndef MP-WEIXIN -->
					<view v-if="!isAuthenticated" class="parent-primary-button home-follow-bar__button" @click="handleMockOfficialAuth">绑定</view>
					<!-- #endif -->
					<view
						v-else
						class="parent-primary-button home-follow-bar__button"
						@click="goOfficialBindingSelection"
					>
						继续
					</view>
				</view>
				<!-- #ifdef MP-WEIXIN -->
				<view v-else class="home-follow-bar__official">
					<official-account></official-account>
				</view>
				<!-- #endif -->
				<view class="home-follow-bar__close" @click="closeOfficialGuide">×</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { computed } from 'vue'
import { onLoad, onShow } from '@dcloudio/uni-app'
import { authorizeByWechatPhone } from '@/common/parent-auth'
import {
	getParentWeChatOfficialStatus,
	getWeChatOfficialBindTicketPreview,
	listParentBoundStudents,
	listParentCampuses,
	listParentPendingStudents
} from '@/common/parent-api'
import { getNavLayout } from '@/common/nav-layout'
import { DEFAULT_PHONE } from '@/common/mock-parent'
import {
	applyParentBoundStudentSummary,
	applyParentCampusSummary,
	applyParentPendingStudentSummary,
	applyWeChatOfficialBindPreview,
	applyWeChatOfficialStatus,
	authorizeByPhone,
	dismissWeChatOfficialGuide,
	getCurrentCampus,
	hasActiveWeChatOfficialBinding,
	parentState
} from '@/common/parent-state'

const nav = getNavLayout()
const authLoading = computed(() => parentState.authLoading)
const isAuthenticated = computed(() => parentState.isAuthenticated)
const currentCampus = computed(() => getCurrentCampus())
const featureList = computed(() => parentState.featureList)
const reportList = computed(() => parentState.noticeList)
const pendingCount = computed(() => parentState.pendingCandidates.length)
const officialStatus = computed(() => parentState.officialStatus || {})
const canSwitchCampus = computed(() => parentState.isAuthenticated && parentState.campusList.length > 1)
const officialBindTicket = computed(() => `${parentState.officialBindTicket || ''}`.trim())
const officialBindPreview = computed(() => parentState.officialBindPreview || {})
const hasActiveOfficialBinding = computed(() => hasActiveWeChatOfficialBinding())
const showOfficialBindAction = computed(() => hasActiveOfficialBinding.value)
const officialAccountName = computed(() => officialStatus.value?.officialAccountName || 'irts家校云')
const institutionName = computed(() => officialBindPreview.value?.institutionName || campusTitle.value || '当前机构')
const showOfficialGuide = computed(() => {
	if (parentState.officialGuideDismissed) {
		return false
	}
	if (showOfficialBindAction.value) {
		return !officialBindPreview.value?.hasBoundStudent
	}
	return isAuthenticated.value && !!officialStatus.value?.needFollowGuide
})
const officialGuideTitle = computed(() => {
	if (showOfficialBindAction.value) {
		return '请完成公众号绑定'
	}
	return '请关注公众号！'
})
const officialGuideDescPrimary = computed(() => {
	if (showOfficialBindAction.value) {
		return `已进入${institutionName.value}公众号绑定流程。`
	}
	return '否则无法接收学员在校消息。'
})
const officialGuideDescSecondary = computed(() => {
	if (showOfficialBindAction.value) {
		return isAuthenticated.value ? '请点击下方按钮选择学员，完成本次绑定。' : '请先点击下方按钮授权手机号，系统会自动匹配可绑定学员。'
	}
	return '请点击右下方按钮，立即关注。'
})
const officialGuideDescTertiary = computed(() => {
	if (showOfficialBindAction.value) {
		return officialBindPreview.value?.sceneStudentId ? '若二维码已指定学员，系统会优先定位该学员。' : '完成绑定后，课表、通知等内容会自动同步到当前家长端。'
	}
	return '若当前已关注，请通过公众号消息卡片重新进入小程序完成绑定。'
})
const officialBarIconText = computed(() => (showOfficialBindAction.value ? '绑' : '公'))
const officialBarTitle = computed(() => {
	if (showOfficialBindAction.value) {
		return isAuthenticated.value ? `继续绑定“${institutionName.value}”学员` : `进入“${institutionName.value}”绑定流程`
	}
	return `未关注“${officialAccountName.value}”公众号`
})
const officialBarDesc = computed(() => {
	if (showOfficialBindAction.value) {
		return isAuthenticated.value ? '选择学员后即可完成本次公众号绑定' : '授权手机号后即可匹配待绑定学员'
	}
	return '关注公众号后才能稳定接收在校通知'
})
const campusTitle = computed(() => {
	if (!isAuthenticated.value && !showOfficialBindAction.value) {
		return '家长端服务'
	}
	return buildCampusTitle(currentCampus.value)
})
const campusSubtitle = computed(() => {
	if (!isAuthenticated.value && !showOfficialBindAction.value) {
		return '授权手机号后查看学员课表、记录和通知'
	}
	return buildCampusSubtitle(currentCampus.value)
})
const campusLogoText = computed(() => {
	if (!isAuthenticated.value && !showOfficialBindAction.value) {
		return '家'
	}
	const title = campusTitle.value || currentCampus.value?.shortName || '校'
	return `${title}`.slice(0, 1)
})

onLoad(query => {
	redirectPEP3CaregiverReport(query)
})

onShow(() => {
	refreshHomeSummary()
	refreshWeChatOfficialPreview()
})

function redirectPEP3CaregiverReport(query = {}) {
	const target = `${query?.target || ''}`.trim()
	const token = `${query?.pep3Token || query?.token || ''}`.trim()
	const ticket = resolvePEP3CaregiverTicket(query)
	if (!ticket && (target !== 'pep3_caregiver_report' || !token)) {
		return
	}
	const params = ticket ? `ticket=${encodeURIComponent(ticket)}` : `token=${encodeURIComponent(token)}`
	uni.redirectTo({
		url: `/pages/pep3-caregiver-report/index?${params}`
	})
}

function resolvePEP3CaregiverTicket(query = {}) {
	const direct = `${query?.ticket || ''}`.trim()
	if (direct) {
		return direct
	}
	const scene = decodeURIComponent(`${query?.scene || ''}`.trim())
	if (!scene) {
		return ''
	}
	if (scene.startsWith('pc')) {
		return scene
	}
	const params = scene.split('&').reduce((out, part) => {
		const [key, value = ''] = part.split('=')
		out[decodeURIComponent(key || '')] = decodeURIComponent(value || '')
		return out
	}, {})
	return `${params.ticket || ''}`.trim()
}

function goCampusPage() {
	uni.navigateTo({
		url: '/pages/select-campus/index'
	})
}

function handleFeature(item) {
	if (item?.key === 'attendance') {
		uni.navigateTo({
			url: '/pages/attendance-record/index'
		})
		return
	}
	if (item?.key === 'course') {
		uni.navigateTo({
			url: '/pages/course-enrollment/index'
		})
		return
	}
	if (item?.key === 'leave') {
		uni.navigateTo({
			url: '/pages/leave/index'
		})
		return
	}
	if (item?.key === 'comment') {
		uni.navigateTo({
			url: '/pages/rehab-record/index'
		})
		return
	}
	uni.showToast({
		title: `${item.title}建设中`,
		icon: 'none'
	})
}

function handleReport(item) {
	uni.showToast({
		title: item.title,
		icon: 'none'
	})
}

function closeOfficialGuide() {
	dismissWeChatOfficialGuide()
}

function handleWechatOfficialAuth(event) {
	authorizeByWechatPhone(event, {
		postAuthPage: '/pages/select-student/index'
	})
}

function handleMockOfficialAuth() {
	authorizeByPhone(DEFAULT_PHONE)
	uni.navigateTo({
		url: '/pages/select-student/index'
	})
}

function goOfficialBindingSelection() {
	if (!isAuthenticated.value) {
		uni.showToast({
			title: '请先完成手机号授权',
			icon: 'none'
		})
		return
	}
	if (!pendingCount.value) {
		uni.showToast({
			title: '当前手机号暂无可绑定学员',
			icon: 'none'
		})
		return
	}
	uni.navigateTo({
		url: '/pages/select-student/index'
	})
}

async function refreshHomeSummary() {
	if (!parentState.isAuthenticated || !parentState.authToken) {
		return
	}
	try {
		const [campusSummary, boundSummary, pendingSummary] = await Promise.all([
			listParentCampuses(parentState.authToken),
			listParentBoundStudents(parentState.authToken),
			listParentPendingStudents(parentState.authToken)
		])
		applyParentCampusSummary(campusSummary)
		applyParentBoundStudentSummary(boundSummary)
		applyParentPendingStudentSummary(pendingSummary)
	} catch (error) {
		console.warn('refresh parent home summary failed', error)
	}

	try {
		const status = await getParentWeChatOfficialStatus(parentState.authToken)
		applyWeChatOfficialStatus(status)
	} catch (error) {
		console.warn('refresh parent wechat official status failed', error)
	}
}

async function refreshWeChatOfficialPreview() {
	if (!officialBindTicket.value) {
		return
	}
	try {
		const preview = await getWeChatOfficialBindTicketPreview({
			bindTicket: officialBindTicket.value
		})
		applyWeChatOfficialBindPreview(preview)
	} catch (error) {
		console.warn('refresh wechat official bind preview failed', error)
	}
}

function buildCampusTitle(campus = {}) {
	const brandName = `${campus?.brandName || ''}`.trim()
	const simpleName = simplifyCampusName(campus?.name)
	return brandName || simpleName || '机构'
}

function buildCampusSubtitle(campus = {}) {
	const title = buildCampusTitle(campus)
	const fullName = `${campus?.name || ''}`.trim()
	const simpleName = simplifyCampusName(fullName)
	if (!fullName || fullName === title || simpleName === title) {
		return ''
	}
	return fullName
}

function simplifyCampusName(name = '') {
	return `${name || ''}`
		.replace(/\s*(总校区|控江校区|校区|分校|院区)\s*$/u, '')
		.trim()
}
</script>

<style scoped>
.home-top-row {
	display: flex;
	justify-content: flex-start;
	align-items: center;
}

.parent-shell--home-guide {
	padding-bottom: 236rpx;
}

.home-campus-trigger {
	display: inline-flex;
	align-items: center;
	gap: 12rpx;
	padding: 12rpx 18rpx;
	border-radius: 999rpx;
	background: rgba(255, 255, 255, 0.62);
	border: 1rpx solid rgba(255, 255, 255, 0.8);
}

.home-campus-trigger__icon {
	width: 32rpx;
	height: 32rpx;
	border-radius: 12rpx;
	background: rgba(255, 214, 10, 0.18);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 18rpx;
	font-weight: 700;
	color: #7a5e00;
}

.home-campus-trigger__text {
	font-size: 26rpx;
	font-weight: 600;
}

.home-campus-card {
	margin-top: 16rpx;
	padding: 24rpx;
}

.home-campus-brand {
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.home-campus-logo {
	width: 88rpx;
	height: 88rpx;
	border-radius: 28rpx;
	background: linear-gradient(135deg, #1dd277 0%, #24c05f 100%);
	display: flex;
	align-items: center;
	justify-content: center;
	overflow: hidden;
	box-shadow: 0 16rpx 34rpx rgba(29, 210, 119, 0.25);
	flex-shrink: 0;
}

.home-campus-logo__image {
	width: 100%;
	height: 100%;
	display: block;
}

.home-campus-logo__text {
	color: #ffffff;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.home-campus-brand__content {
	flex: 1;
	display: flex;
	flex-direction: column;
	justify-content: center;
	min-height: 88rpx;
}

.home-campus-brand__name {
	font-size: 36rpx;
	font-weight: 700;
	line-height: 1.3;
}

.home-campus-brand__campus {
	margin-top: 8rpx;
	font-size: 22rpx;
	line-height: 1.5;
	color: var(--parent-subtext);
}

.home-grid-card {
	padding: 32rpx 16rpx 32rpx;
}

.home-grid {
	display: grid;
	grid-template-columns: repeat(4, 1fr);
	gap: 22rpx 8rpx;
}

.home-grid-item {
	display: flex;
	flex-direction: column;
	align-items: center;
}

.home-grid-item__badge {
	width: 78rpx;
	height: 78rpx;
	border-radius: 28rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	font-size: 28rpx;
	font-weight: 700;
	box-shadow: 0 14rpx 28rpx rgba(65, 65, 65, 0.08);
}

.home-grid-item__text {
	margin-top: 12rpx;
	font-size: 22rpx;
	text-align: center;
	line-height: 1.4;
}

.home-notice-card {
	margin-top: 26rpx;
	padding: 24rpx;
}

.home-notice-card__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	padding-bottom: 10rpx;
}

.home-notice-card__title {
	font-size: 28rpx;
	font-weight: 700;
}

.home-notice-card__more {
	font-size: 22rpx;
	color: var(--parent-subtext);
}

.home-notice-item {
	display: flex;
	gap: 18rpx;
	padding: 24rpx 0;
	border-top: 1rpx solid var(--parent-divider);
}

.home-notice-item:first-of-type {
	margin-top: 16rpx;
}

.home-notice-item__dot {
	width: 16rpx;
	height: 16rpx;
	border-radius: 50%;
	background: var(--parent-brand-deep);
	margin-top: 14rpx;
	flex-shrink: 0;
}

.home-notice-item__content {
	display: flex;
	flex-direction: column;
}

.home-notice-item__title {
	font-size: 26rpx;
	font-weight: 600;
}

.home-notice-item__desc {
	margin-top: 8rpx;
	font-size: 22rpx;
	line-height: 1.5;
	color: var(--parent-subtext);
}

.home-follow-mask {
	position: fixed;
	inset: 0;
	z-index: 88;
	pointer-events: none;
}

.home-follow-dialog {
	position: absolute;
	left: 50%;
	bottom: calc(env(safe-area-inset-bottom) + 120rpx);
	transform: translateX(-50%);
	width: calc(100vw - 96rpx);
	padding: 28rpx 32rpx 30rpx;
	border-radius: 24rpx;
	background: rgba(39, 39, 39, 0.96);
	box-shadow: 0 32rpx 72rpx rgba(0, 0, 0, 0.2);
	color: #ffffff;
	pointer-events: auto;
}

.home-follow-dialog__close {
	position: absolute;
	top: 18rpx;
	right: 22rpx;
	font-size: 44rpx;
	line-height: 1;
	color: rgba(255, 255, 255, 0.5);
}

.home-follow-dialog__title {
	display: block;
	padding-right: 48rpx;
	font-size: 46rpx;
	font-weight: 700;
	line-height: 1.3;
}

.home-follow-dialog__desc {
	display: block;
	margin-top: 18rpx;
	font-size: 28rpx;
	line-height: 1.7;
	color: rgba(255, 255, 255, 0.92);
}

.home-follow-dialog__actions {
	margin-top: 28rpx;
}

.home-follow-dialog__action {
	width: 100%;
	height: 84rpx;
}

.home-follow-bar {
	position: fixed;
	left: 24rpx;
	right: 24rpx;
	bottom: calc(env(safe-area-inset-bottom));
	z-index: 89;
}

.home-follow-bar__panel {
	position: relative;
	display: flex;
	align-items: center;
	gap: 20rpx;
	padding: 18rpx 26rpx;
	border-radius: 999rpx;
	background: rgba(36, 36, 36, 0.98);
	box-shadow: 0 28rpx 48rpx rgba(0, 0, 0, 0.16);
}

.home-follow-bar__icon {
	width: 72rpx;
	height: 72rpx;
	border-radius: 50%;
	background: linear-gradient(135deg, #fff26a 0%, #ffe20a 100%);
	display: flex;
	align-items: center;
	justify-content: center;
	color: #2d2708;
	font-size: 28rpx;
	font-weight: 700;
	flex-shrink: 0;
}

.home-follow-bar__copy {
	flex: 1;
	min-width: 0;
	display: flex;
	flex-direction: column;
}

.home-follow-bar__title {
	font-size: 30rpx;
	font-weight: 700;
	line-height: 1.3;
	color: #ffffff;
}

.home-follow-bar__desc {
	margin-top: 6rpx;
	font-size: 22rpx;
	line-height: 1.5;
	color: rgba(255, 255, 255, 0.72);
}

.home-follow-bar__action,
.home-follow-bar__official {
	flex-shrink: 0;
}

.home-follow-bar__button {
	min-width: 164rpx;
	height: 78rpx;
	padding: 0 34rpx;
}

.home-follow-bar__close {
	position: absolute;
	top: -10rpx;
	right: -10rpx;
	width: 54rpx;
	height: 54rpx;
	border-radius: 50%;
	background: #d9d9d9;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #484848;
	font-size: 34rpx;
	line-height: 1;
}
</style>
