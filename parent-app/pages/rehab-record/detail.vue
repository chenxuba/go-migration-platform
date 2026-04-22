<template>
	<view class="parent-page rehab-detail-page">
		<view class="rehab-detail-nav-fixed" :style="{ paddingTop: `${nav.top}px` }">
			<view class="rehab-detail-nav-fixed__inner">
				<view class="parent-nav-row rehab-detail-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="rehab-detail-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="rehab-detail-back" @click="goBack">
							<view class="rehab-detail-back__icon"></view>
						</view>
					</view>
					<text class="rehab-detail-nav__title">康复记录详情</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>
			</view>
		</view>

		<view class="parent-shell rehab-detail-shell" :style="{ paddingTop: `${nav.top + nav.height + 18}px` }">
			<template v-if="!isAuthenticated">
				<view class="parent-card rehab-detail-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">康</view>
						<text class="parent-empty-title">登录后即可查看康复记录详情</text>
						<text class="parent-empty-desc">手机号授权后，系统会自动同步当前学员的康复记录详情。</text>
						<!-- #ifdef MP-WEIXIN -->
						<button class="parent-primary-button rehab-detail-auth-button" :loading="authLoading" :disabled="authLoading" open-type="getPhoneNumber" @getphonenumber="handleWechatPhoneAuth">
							授权登录
						</button>
						<!-- #endif -->
						<!-- #ifndef MP-WEIXIN -->
						<view class="parent-primary-button rehab-detail-auth-button" @click="handleMockPhoneAuth">授权登录</view>
						<!-- #endif -->
					</view>
				</view>
			</template>

			<template v-else-if="pageLoading && !record.id">
				<view class="parent-card rehab-detail-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">载</view>
						<text class="parent-empty-title">正在加载康复记录详情</text>
						<text class="parent-empty-desc">稍等一下，系统正在同步最新记录。</text>
					</view>
				</view>
			</template>

			<template v-else-if="record.id">
				<view class="rehab-detail-hero">
					<view class="rehab-detail-hero__title-row">
						<text class="rehab-detail-hero__title">{{ detailTitle }}</text>
						<text class="rehab-detail-hero__badge">已记录</text>
					</view>
					<text class="rehab-detail-hero__meta">{{ record.lessonTime }}</text>
				</view>

				<view class="parent-card rehab-detail-card rehab-detail-card--basic">
					<view class="rehab-detail-card__header rehab-detail-card__header--simple">
						<text class="rehab-detail-card__title rehab-detail-card__title--subtle">基本信息</text>
					</view>

					<view class="rehab-detail-card__list">
						<view v-for="row in detailRows" :key="row.label" class="rehab-detail-row">
							<text class="rehab-detail-row__label">{{ row.label }}</text>
							<text class="rehab-detail-row__value">{{ row.value }}</text>
						</view>
					</view>
				</view>

				<view v-if="contentBlocks.length" class="rehab-detail-section-list">
					<view
						v-for="block in contentBlocks"
						:key="block.key"
						class="parent-card rehab-detail-section-card"
					>
						<text class="rehab-detail-section-card__title">{{ block.title }}</text>
						<text class="rehab-detail-section-card__content">{{ block.content }}</text>
					</view>
				</view>

				<view v-if="trainingItems.length" class="parent-card rehab-detail-card rehab-detail-training-card">
					<view class="rehab-detail-card__header">
						<text class="rehab-detail-card__title">训练内容</text>
					</view>
					<view class="rehab-detail-training-list">
						<view v-for="(item, index) in trainingItems" :key="`training-${index}`" class="rehab-detail-training-item">
							<text v-if="item.title" class="rehab-detail-training-item__title">{{ item.title }}</text>
							<text class="rehab-detail-training-item__content">{{ item.content }}</text>
						</view>
					</view>
				</view>

				<view v-if="previousSummary" class="parent-card rehab-detail-note-card">
					<text class="rehab-detail-note-card__title">上一条记录</text>
					<text class="rehab-detail-note-card__content">{{ previousSummary }}</text>
				</view>

				<view class="parent-card rehab-detail-feedback-card">
					<view class="rehab-detail-card__header">
						<text class="rehab-detail-card__title">家长反馈</text>
					</view>

					<textarea
						v-model="feedbackInput"
						class="rehab-detail-feedback__textarea"
						placeholder="请输入家长反馈内容"
						maxlength="300"
						:disabled="feedbackSaving"
					/>

					<view class="rehab-detail-signature">
						<view class="rehab-detail-signature__header">
							<text class="rehab-detail-signature__title">家长签名</text>
							<view class="rehab-detail-signature__actions">
								<text v-if="signatureValue" class="rehab-detail-signature__action" @click="clearSignature">清空</text>
								<text class="rehab-detail-signature__action rehab-detail-signature__action--primary" @click="openSignatureDialog">{{ signatureActionText }}</text>
							</view>
						</view>

						<view class="rehab-detail-signature__panel" @click="openSignatureDialog">
							<image v-if="signatureValue" class="rehab-detail-signature__image" :src="signatureValue" mode="aspectFit"></image>
							<view v-else class="rehab-detail-signature__placeholder">
								<text class="rehab-detail-signature__placeholder-title">点击完成在线电子签名</text>
								<text class="rehab-detail-signature__placeholder-desc">提交反馈前请先签名确认</text>
							</view>
						</view>

						<text class="rehab-detail-signature__hint">{{ signatureHint }}</text>
					</view>

					<view class="rehab-detail-feedback__footer">
						<text class="rehab-detail-feedback__hint">{{ feedbackHint }}</text>
						<button class="parent-primary-button rehab-detail-feedback__submit" :loading="feedbackSaving" :disabled="feedbackSaving || !feedbackDirty" @click="submitParentFeedback">
							提交反馈
						</button>
					</view>
				</view>
			</template>

			<template v-else>
				<view class="parent-card rehab-detail-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">空</view>
						<text class="parent-empty-title">未找到康复记录详情</text>
						<text class="parent-empty-desc">请返回康复记录列表后重新进入。</text>
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
import { getParentRehabRecordDetail, saveParentRehabFeedback, uploadParentRehabSignature } from '@/common/parent-api'
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
const feedbackInput = ref('')
const signatureValue = ref('')
const signatureDraftFilePath = ref('')
const feedbackSaving = ref(false)
const authLoading = computed(() => parentState.authLoading)
const isAuthenticated = computed(() => parentState.isAuthenticated)
const signatureDraftStorageKey = 'parent_rehab_signature_draft'

let detailRequestSerial = 0
let skipNextOnShowRefresh = false

const detail = computed(() => normalizeDetail(detailData.value || {}))
const record = computed(() => detail.value.record)
const published = computed(() => detail.value.published)
const trainingItems = computed(() => published.value.trainingItems || [])
const signatureActionText = computed(() => signatureValue.value ? '重签' : '在线签名')
const signatureHint = computed(() => {
	if (signatureValue.value) {
		return published.value.feedbackDate ? `已签名，反馈日期 ${published.value.feedbackDate}` : '已完成电子签名，可重新签名覆盖当前内容'
	}
	return '请先完成电子签名，再提交家长反馈'
})
const feedbackHint = computed(() => {
	if (published.value.feedbackDate) {
		return `最近反馈日期：${published.value.feedbackDate}`
	}
	return '提交后会同步到当前康复记录'
})
const feedbackDirty = computed(() => {
	return feedbackInput.value.trim() !== (published.value.parentFeedback || '')
		|| signatureValue.value.trim() !== (published.value.parentSignature || '')
})
const detailTitle = computed(() => {
	const studentName = record.value.studentName === '-' ? '' : record.value.studentName
	const courseName = record.value.courseName === '-' ? record.value.className : record.value.courseName
	const parts = [studentName, courseName].filter(Boolean)
	return parts.length ? parts.join(' · ') : '康复记录'
})
const detailRows = computed(() => {
	const rows = []
	if (record.value.teacherName !== '-') {
		rows.push({
			label: '授课老师',
			value: record.value.teacherName
		})
	}
	if (record.value.classroom !== '-') {
		rows.push({
			label: '上课教室',
			value: record.value.classroom
		})
	}
	rows.push({
		label: '训练机构',
		value: record.value.campusName
	})
	rows.push({
		label: '记录时间',
		value: record.value.updatedTime
	})
	if (record.value.updatedStaffName !== '-' && record.value.updatedStaffName !== record.value.teacherName) {
		rows.push({
			label: '记录老师',
			value: record.value.updatedStaffName
		})
	}
	return rows
})
const contentBlocks = computed(() => {
	const blocks = []
	if (published.value.trainingTarget) {
		blocks.push({
			key: 'target',
			title: '训练目标',
			content: published.value.trainingTarget
		})
	}
	if (published.value.performance) {
		blocks.push({
			key: 'performance',
			title: '课堂表现',
			content: published.value.performance
		})
	}
	if (published.value.suggestion) {
		blocks.push({
			key: 'suggestion',
			title: '康复建议',
			content: published.value.suggestion
		})
	}
	if (!blocks.length && record.value.summaryText !== '-') {
		blocks.push({
			key: 'summary',
			title: '记录摘要',
			content: record.value.summaryText
		})
	}
	return blocks
})
const previousSummary = computed(() => {
	const source = detail.value.previousPublished
	const candidates = [
		source.trainingTarget,
		source.performance,
		source.suggestion
	].filter(Boolean)
	if (candidates.length) {
		return candidates[0]
	}
	return ''
})

onLoad(query => {
	routeStudentId.value = `${query?.studentId || ''}`.trim()
	routeRecordId.value = `${query?.studentTeachingRecordId || ''}`.trim()
})

onShow(() => {
	consumeSignatureDraft()
	if (skipNextOnShowRefresh) {
		skipNextOnShowRefresh = false
		return
	}
	if (!parentState.isAuthenticated || !parentState.authToken) {
		resetDetail()
		return
	}
	if (!routeStudentId.value) {
		routeStudentId.value = `${parentState.currentStudentId || ''}`.trim()
	}
	if (!routeRecordId.value) {
		return
	}
	refreshDetail()
})

function normalizeDetail(item = {}) {
	const recordItem = item?.record || {}
	const publishedItem = item?.published?.content || {}
	const previousItem = item?.previousPublished?.content || {}
	return {
		record: {
			id: `${recordItem?.id || recordItem?.studentTeachingRecordId || ''}`.trim(),
			studentTeachingRecordId: `${recordItem?.studentTeachingRecordId || ''}`.trim(),
			studentName: `${recordItem?.studentName || item?.student?.name || '-'}`.trim() || '-',
			campusName: `${recordItem?.campusName || item?.student?.campusName || '-'}`.trim() || '-',
			className: `${recordItem?.className || recordItem?.courseName || '康复记录'}`.trim() || '康复记录',
			courseName: `${recordItem?.courseName || recordItem?.className || '-'}`.trim() || '-',
			lessonTime: `${recordItem?.lessonTime || ''}`.trim() || buildLessonTimeText(recordItem),
			teacherName: `${recordItem?.teacherName || '-'}`.trim() || '-',
			classroom: `${recordItem?.classroom || '-'}`.trim() || '-',
			summaryText: `${recordItem?.summaryText || '-'}`.trim() || '-',
			updatedTime: `${recordItem?.updatedTime || '-'}`.trim() || '-',
			updatedStaffName: `${recordItem?.updatedStaffName || '-'}`.trim() || '-'
		},
		published: {
			trainingTarget: `${publishedItem?.trainingTarget || ''}`.trim(),
			performance: `${publishedItem?.performance || ''}`.trim(),
			suggestion: `${publishedItem?.suggestion || ''}`.trim(),
			parentFeedback: `${publishedItem?.parentFeedback || ''}`.trim(),
			parentSignature: `${publishedItem?.parentSignature || ''}`.trim(),
			feedbackDate: `${publishedItem?.feedbackDate || ''}`.trim(),
			trainingItems: normalizeTrainingItems(publishedItem?.trainingItems || [])
		},
		previousPublished: {
			trainingTarget: `${previousItem?.trainingTarget || ''}`.trim(),
			performance: `${previousItem?.performance || ''}`.trim(),
			suggestion: `${previousItem?.suggestion || ''}`.trim()
		}
	}
}

function normalizeTrainingItems(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		key: `${item?.title || 'item'}-${index + 1}`,
		title: `${item?.title || ''}`.trim(),
		content: `${item?.content || ''}`.trim() || '未填写'
	})).filter(item => item.title || item.content)
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

function resetDetail() {
	detailData.value = {}
	feedbackInput.value = ''
	signatureValue.value = ''
	signatureDraftFilePath.value = ''
	pageLoading.value = false
}

async function refreshDetail() {
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
		const result = await getParentRehabRecordDetail(token, {
			studentId: routeStudentId.value,
			studentTeachingRecordId: routeRecordId.value
		})
		if (requestSerial !== detailRequestSerial || token !== `${parentState.authToken || ''}`.trim()) {
			return
		}
		detailData.value = result || {}
		applyFeedbackFormState(result || {})
	} catch (error) {
		if (requestSerial !== detailRequestSerial) {
			return
		}
		console.warn('load parent rehab record detail failed', error)
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
	return `/pages/rehab-record/detail?${query.join('&')}`
}

function applyFeedbackFormState(item = {}) {
	feedbackInput.value = `${item?.published?.content?.parentFeedback || ''}`.trim()
	signatureValue.value = `${item?.published?.content?.parentSignature || ''}`.trim()
	signatureDraftFilePath.value = ''
}

function applySignatureDraft(tempFilePath = '') {
	const nextFilePath = `${tempFilePath || ''}`.trim()
	if (!nextFilePath) {
		return
	}
	signatureValue.value = nextFilePath
	signatureDraftFilePath.value = nextFilePath
}

function consumeSignatureDraft() {
	try {
		const tempFilePath = `${uni.getStorageSync(signatureDraftStorageKey) || ''}`.trim()
		if (!tempFilePath) {
			return
		}
		uni.removeStorageSync(signatureDraftStorageKey)
		applySignatureDraft(tempFilePath)
	} catch (error) {
		console.warn('consume parent rehab signature draft failed', error)
	}
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
			uni.navigateTo({
				url: '/pages/rehab-record/index'
			})
		}
	})
}

function openSignatureDialog() {
	if (feedbackSaving.value) {
		return
	}
	skipNextOnShowRefresh = true
	uni.navigateTo({
		url: '/pages/rehab-record/signature',
		events: {
			parentRehabSignatureComplete(payload) {
				applySignatureDraft(payload?.tempFilePath)
			}
		}
	})
}

function clearSignature() {
	signatureValue.value = ''
	signatureDraftFilePath.value = ''
	uni.removeStorageSync(signatureDraftStorageKey)
}

function isLocalSignatureFilePath(value = '') {
	const text = `${value || ''}`.trim()
	if (!text) {
		return false
	}
	return text.startsWith('wxfile://')
		|| text.startsWith('file://')
		|| text.startsWith('/private/')
		|| text.startsWith('/var/mobile/')
}

async function submitParentFeedback() {
	if (feedbackSaving.value) {
		return
	}
	if (!feedbackDirty.value) {
		uni.showToast({
			title: '暂无变更',
			icon: 'none'
		})
		return
	}
	if (!signatureValue.value.trim()) {
		uni.showToast({
			title: '请先完成电子签名',
			icon: 'none'
		})
		return
	}

	const token = `${parentState.authToken || ''}`.trim()
	if (!token || !routeRecordId.value) {
		return
	}

	feedbackSaving.value = true
	try {
		let signatureSubmitValue = signatureValue.value.trim()
		const draftFilePath = signatureDraftFilePath.value.trim()
		if (draftFilePath) {
			const uploadResult = await uploadParentRehabSignature(token, draftFilePath)
			signatureSubmitValue = `${uploadResult?.url || ''}`.trim()
		} else if (isLocalSignatureFilePath(signatureSubmitValue)) {
			const uploadResult = await uploadParentRehabSignature(token, signatureSubmitValue)
			signatureSubmitValue = `${uploadResult?.url || ''}`.trim()
		}
		if (!signatureSubmitValue) {
			throw new Error('签名上传失败')
		}
		signatureValue.value = signatureSubmitValue
		signatureDraftFilePath.value = ''
		await saveParentRehabFeedback(token, {
			studentId: routeStudentId.value,
			studentTeachingRecordId: routeRecordId.value,
			parentFeedback: feedbackInput.value,
			parentSignature: signatureSubmitValue
		})
		await refreshDetail()
		uni.showToast({
			title: '反馈已提交',
			icon: 'success'
		})
	} catch (error) {
		console.warn('save parent rehab feedback failed', error)
		uni.showToast({
			title: `${error?.message || '提交失败'}`.slice(0, 24),
			icon: 'none'
		})
	} finally {
		feedbackSaving.value = false
	}
}
</script>

<style scoped>
.rehab-detail-page {
	min-height: 100vh;
	background:
		radial-gradient(circle at 10% 4%, rgba(187, 226, 181, 0.16), transparent 22%),
		linear-gradient(180deg, #fff8ef 0%, #fffaf4 42%, #fffdf9 100%);
}

.rehab-detail-shell {
	padding-bottom: calc(60rpx + env(safe-area-inset-bottom));
}

.rehab-detail-nav-fixed {
	position: fixed;
	top: 0;
	left: 0;
	right: 0;
	z-index: 40;
	background:
		linear-gradient(180deg, rgba(255, 248, 239, 0.98) 0%, rgba(255, 250, 244, 0.96) 72%, rgba(255, 253, 249, 0.9) 100%);
	backdrop-filter: blur(12rpx);
}

.rehab-detail-nav-fixed__inner {
	padding: 0 24rpx 10rpx;
}

.rehab-detail-nav__side {
	display: flex;
	align-items: center;
}

.rehab-detail-nav__title {
	flex: 1;
	text-align: center;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.rehab-detail-back {
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

.rehab-detail-back__icon {
	width: 16rpx;
	height: 16rpx;
	border-left: 4rpx solid #433e33;
	border-bottom: 4rpx solid #433e33;
	transform: translateX(4rpx) rotate(45deg);
}

.rehab-detail-empty-card {
	margin-top: 10rpx;
}

.rehab-detail-auth-button {
	width: 100%;
	margin-top: 34rpx;
}

.rehab-detail-hero {
	padding: 18rpx 6rpx 24rpx;
	display: flex;
	flex-direction: column;
	align-items: flex-start;
}

.rehab-detail-hero__title-row {
	width: 100%;
	display: flex;
	align-items: center;
	gap: 16rpx;
}

.rehab-detail-hero__title {
	flex: 1;
	min-width: 0;
	font-size: 38rpx;
	font-weight: 700;
	line-height: 1.34;
	color: #1f1f1f;
}

.rehab-detail-hero__badge {
	flex-shrink: 0;
	padding: 10rpx 16rpx;
	border-radius: 999rpx;
	background: rgba(231, 244, 234, 0.96);
	color: #23915f;
	font-size: 22rpx;
	font-weight: 700;
	line-height: 1;
}

.rehab-detail-hero__meta {
	margin-top: 10rpx;
	font-size: 24rpx;
	line-height: 1.5;
	color: #978f80;
}

.rehab-detail-card {
	padding: 30rpx 28rpx 12rpx;
}

.rehab-detail-card--basic {
	padding-top: 22rpx;
}

.rehab-detail-card__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 18rpx;
}

.rehab-detail-card__header--simple {
	justify-content: flex-start;
}

.rehab-detail-card__title {
	flex: 1;
	min-width: 0;
	font-size: 32rpx;
	font-weight: 700;
	line-height: 1.45;
	color: #2a2721;
}

.rehab-detail-card__title--subtle {
	font-size: 26rpx;
	font-weight: 600;
	color: #8f8778;
}

.rehab-detail-card__status {
	flex-shrink: 0;
	padding: 10rpx 18rpx;
	border-radius: 999rpx;
	background: rgba(255, 249, 236, 0.96);
	color: #b78400;
	font-size: 22rpx;
	font-weight: 700;
	line-height: 1;
}

.rehab-detail-card__list {
	margin-top: 10rpx;
}

.rehab-detail-row {
	display: flex;
	align-items: flex-start;
	padding: 22rpx 0;
	border-top: 1rpx solid var(--parent-divider);
	gap: 18rpx;
}

.rehab-detail-row__label {
	width: 140rpx;
	flex-shrink: 0;
	font-size: 26rpx;
	line-height: 1.5;
	color: #928b7d;
}

.rehab-detail-row__value {
	flex: 1;
	min-width: 0;
	font-size: 27rpx;
	line-height: 1.62;
	color: #2f2b25;
}

.rehab-detail-section-list {
	margin-top: 24rpx;
	display: flex;
	flex-direction: column;
	gap: 20rpx;
}

.rehab-detail-section-card {
	padding: 28rpx 28rpx 30rpx;
}

.rehab-detail-section-card__title {
	display: block;
	position: relative;
	padding-left: 18rpx;
	font-size: 30rpx;
	font-weight: 700;
	line-height: 1.3;
	color: #1f1f1f;
}

.rehab-detail-section-card__title::before {
	content: '';
	position: absolute;
	left: 0;
	top: 50%;
	transform: translateY(-50%);
	width: 6rpx;
	height: 28rpx;
	border-radius: 999rpx;
	background: linear-gradient(180deg, #9bcf6b 0%, #6fb84b 100%);
}

.rehab-detail-section-card__content {
	display: block;
	margin-top: 18rpx;
	font-size: 28rpx;
	line-height: 1.75;
	color: #35312b;
}

.rehab-detail-training-card {
	margin-top: 24rpx;
	padding-bottom: 22rpx;
}

.rehab-detail-training-list {
	margin-top: 16rpx;
}

.rehab-detail-training-item {
	padding: 22rpx 0;
	border-top: 1rpx solid var(--parent-divider);
}

.rehab-detail-training-item__title {
	display: block;
	font-size: 28rpx;
	font-weight: 700;
	line-height: 1.42;
	color: #2f2b25;
}

.rehab-detail-training-item__content {
	display: block;
	margin-top: 10rpx;
	font-size: 26rpx;
	line-height: 1.7;
	color: #645d50;
}

.rehab-detail-note-card {
	margin-top: 24rpx;
	padding: 28rpx;
	background: rgba(255, 251, 242, 0.94);
}

.rehab-detail-note-card__title {
	display: block;
	font-size: 28rpx;
	font-weight: 700;
	line-height: 1.35;
	color: #9a7f2c;
}

.rehab-detail-note-card__content {
	display: block;
	margin-top: 12rpx;
	font-size: 25rpx;
	line-height: 1.65;
	color: #6b6048;
}

.rehab-detail-feedback-card {
	margin-top: 24rpx;
	padding: 28rpx;
}

.rehab-detail-feedback__textarea {
	width: 100%;
	min-height: 180rpx;
	margin-top: 18rpx;
	padding: 22rpx 20rpx;
	border-radius: 24rpx;
	background: rgba(255, 252, 246, 0.96);
	border: 1rpx solid rgba(245, 239, 229, 0.92);
	font-size: 28rpx;
	line-height: 1.7;
	color: #2f2b25;
}

.rehab-detail-signature {
	margin-top: 24rpx;
}

.rehab-detail-signature__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 16rpx;
}

.rehab-detail-signature__title {
	font-size: 28rpx;
	font-weight: 700;
	line-height: 1.35;
	color: #2a2721;
}

.rehab-detail-signature__actions {
	display: inline-flex;
	align-items: center;
	gap: 18rpx;
	flex-shrink: 0;
}

.rehab-detail-signature__action {
	font-size: 24rpx;
	line-height: 1.3;
	color: #a59a87;
}

.rehab-detail-signature__action--primary {
	color: #b98905;
	font-weight: 600;
}

.rehab-detail-signature__panel {
	margin-top: 16rpx;
	padding: 18rpx;
	border-radius: 24rpx;
	background: rgba(255, 252, 246, 0.96);
	border: 1rpx solid rgba(245, 239, 229, 0.92);
	min-height: 200rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	overflow: hidden;
}

.rehab-detail-signature__image {
	width: 100%;
	height: 164rpx;
	display: block;
}

.rehab-detail-signature__placeholder {
	display: flex;
	flex-direction: column;
	align-items: center;
	text-align: center;
}

.rehab-detail-signature__placeholder-title {
	font-size: 28rpx;
	font-weight: 600;
	line-height: 1.4;
	color: #3b362f;
}

.rehab-detail-signature__placeholder-desc {
	margin-top: 10rpx;
	font-size: 24rpx;
	line-height: 1.5;
	color: #a09684;
}

.rehab-detail-signature__hint {
	display: block;
	margin-top: 12rpx;
	font-size: 22rpx;
	line-height: 1.5;
	color: #a79d8c;
}

.rehab-detail-feedback__footer {
	margin-top: 26rpx;
	display: flex;
	flex-direction: column;
	gap: 18rpx;
}

.rehab-detail-feedback__hint {
	font-size: 22rpx;
	line-height: 1.5;
	color: #a79d8c;
}

.rehab-detail-feedback__submit {
	width: 100%;
}
</style>
