<template>
	<view class="parent-page caregiver-report-page">
		<view class="caregiver-nav-fixed" :style="{ paddingTop: `${nav.top}px` }">
			<view class="caregiver-nav-fixed__inner">
				<view class="parent-nav-row caregiver-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="caregiver-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="caregiver-back" @click="goBack">
							<view class="caregiver-back__icon"></view>
						</view>
					</view>
					<text class="caregiver-nav__title">照顾者报告</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>
			</view>
		</view>

		<view class="parent-shell caregiver-shell" :style="{ paddingTop: `${nav.top + nav.height + 18}px` }">
			<template v-if="pageLoading">
				<view class="parent-card caregiver-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">载</view>
						<text class="parent-empty-title">正在加载照顾者报告</text>
						<text class="parent-empty-desc">请稍候。</text>
					</view>
				</view>
			</template>

			<template v-else-if="loadError">
				<view class="parent-card caregiver-empty-card">
					<view class="parent-empty-card">
						<view class="parent-empty-badge">失</view>
						<text class="parent-empty-title">无法打开照顾者报告</text>
						<text class="parent-empty-desc">{{ loadError }}</text>
					</view>
				</view>
			</template>

			<template v-else>
				<view class="caregiver-hero">
					<view>
						<text class="caregiver-hero__title">{{ studentNameText }}</text>
						<text class="caregiver-hero__subtitle">请根据儿童近期日常表现填写。</text>
					</view>
					<view class="caregiver-progress">
						<text>{{ answeredScoredCount }}</text>
						<text>/{{ totalScoredCount }}</text>
					</view>
				</view>

				<view v-if="submitted" class="parent-card caregiver-success-card">
					<text class="caregiver-success-card__title">已提交</text>
					<text class="caregiver-success-card__desc">照顾者报告已同步到老师端PEP-3草稿。</text>
				</view>

				<view class="parent-card caregiver-basic-card">
					<text class="caregiver-section-title">填写人信息</text>
					<view class="caregiver-field">
						<text class="caregiver-field__label">填写人姓名</text>
						<input
							class="caregiver-input"
							:value="respondentName"
							placeholder="可选"
							@input="event => setRespondentName(event.detail.value)"
						/>
					</view>
					<view class="caregiver-field">
						<text class="caregiver-field__label">与儿童关系</text>
						<input
							class="caregiver-input"
							:value="relationship"
							placeholder="例如：父亲、母亲、主要照顾者"
							@input="event => setRelationship(event.detail.value)"
						/>
					</view>
				</view>

				<view v-for="section in sections" :key="section.sectionCode" class="parent-card caregiver-section-card">
					<view class="caregiver-section-card__header">
						<text class="caregiver-section-title">{{ section.title }}</text>
						<text v-if="section.scored" class="caregiver-section-count">{{ sectionAnsweredCount(section) }}/{{ sectionScoredCount(section) }}</text>
					</view>
					<text v-if="section.description" class="caregiver-section-desc">{{ section.description }}</text>

					<view v-if="section.inputType === 'diagnosis_matrix'" class="diagnosis-list">
						<view v-for="category in section.diagnosisCategories || []" :key="category.key" class="diagnosis-item">
							<text class="diagnosis-item__title">{{ category.label }}</text>
							<view class="diagnosis-item__group">
								<text class="diagnosis-item__label">诊断</text>
								<radio-group class="caregiver-radio-grid caregiver-radio-grid--compact" @change="event => setDiagnosisAnswer(category.key, 'status', event.detail.value)">
									<label v-for="option in diagnosisStatusOptions" :key="option.value" class="caregiver-radio-option">
										<radio :value="option.value" :checked="getDiagnosisAnswer(category.key, 'status') === option.value" color="#f1b800" />
										<text>{{ option.label }}</text>
									</label>
								</radio-group>
							</view>
							<view class="diagnosis-item__group">
								<text class="diagnosis-item__label">程度</text>
								<radio-group class="caregiver-radio-grid caregiver-radio-grid--compact" @change="event => setDiagnosisAnswer(category.key, 'severity', event.detail.value)">
									<label v-for="option in diagnosisSeverityOptions" :key="option.value" class="caregiver-radio-option">
										<radio :value="option.value" :checked="getDiagnosisAnswer(category.key, 'severity') === option.value" color="#f1b800" />
										<text>{{ option.label }}</text>
									</label>
								</radio-group>
							</view>
						</view>
					</view>

					<view v-else class="caregiver-item-list">
						<view v-for="item in section.items || []" :key="item.key" class="caregiver-item">
							<view class="caregiver-item__prompt-row">
								<text class="caregiver-item__no">{{ item.itemNo }}</text>
								<text class="caregiver-item__prompt">{{ item.prompt }}</text>
							</view>

							<view v-if="item.fieldType === 'number'" class="caregiver-number-wrap">
								<input
									class="caregiver-input caregiver-input--number"
									type="number"
									:value="getAnswer(section.sectionCode, item.key)"
									placeholder="请输入"
									@input="event => setAnswer(section.sectionCode, item.key, event.detail.value)"
								/>
								<text v-if="item.unit" class="caregiver-input-unit">{{ item.unit }}</text>
							</view>

							<radio-group
								v-else-if="item.fieldType === 'radio'"
								class="caregiver-radio-grid"
								@change="event => setAnswer(section.sectionCode, item.key, event.detail.value)"
							>
								<label v-for="option in item.options || []" :key="option.value" class="caregiver-radio-option">
									<radio :value="option.value" :checked="getAnswer(section.sectionCode, item.key) === option.value" color="#f1b800" />
									<text>{{ option.label }}</text>
								</label>
							</radio-group>

							<textarea
								v-else
								class="caregiver-textarea"
								:value="getAnswer(section.sectionCode, item.key)"
								placeholder="请输入"
								maxlength="500"
								@input="event => setAnswer(section.sectionCode, item.key, event.detail.value)"
							/>
						</view>
					</view>
				</view>

				<view class="caregiver-submit-bar">
					<button class="parent-primary-button caregiver-submit-button" :loading="submitting" :disabled="submitting || submitted" @click="submitReport">
						{{ submitted ? '已提交' : '提交照顾者报告' }}
					</button>
					<text v-if="missingScoredCount" class="caregiver-submit-hint">还有 {{ missingScoredCount }} 道计分题未完成</text>
				</view>
			</template>
		</view>
	</view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { onLoad } from '@dcloudio/uni-app'
import { getNavLayout } from '@/common/nav-layout'
import { getPEP3CaregiverReportTemplate, submitPEP3CaregiverReport } from '@/common/parent-api'

const nav = getNavLayout()
const routeToken = ref('')
const pageLoading = ref(false)
const submitting = ref(false)
const loadError = ref('')
const submitted = ref(false)
const pageData = ref({})
const answers = ref({})
const respondentName = ref('')
const relationship = ref('')

const diagnosisStatusOptions = [
	{ value: 'yes', label: '是' },
	{ value: 'no', label: '不是' },
	{ value: 'unknown', label: '不知道' }
]

const diagnosisSeverityOptions = [
	{ value: 'none', label: '没有' },
	{ value: 'mild', label: '轻微' },
	{ value: 'moderate', label: '中度' },
	{ value: 'severe', label: '严重' }
]

const template = computed(() => pageData.value?.template || {})
const sections = computed(() => template.value?.sections || [])
const studentNameText = computed(() => `${pageData.value?.studentName || '儿童'}`.trim() || '儿童')
const scoredItems = computed(() => {
	return sections.value.flatMap(section => {
		if (!section.scored) {
			return []
		}
		return (section.items || [])
			.filter(item => item.scored)
			.map(item => ({ sectionCode: section.sectionCode, item }))
	})
})
const totalScoredCount = computed(() => scoredItems.value.length)
const answeredScoredCount = computed(() => scoredItems.value.filter(({ sectionCode, item }) => !!getAnswer(sectionCode, item.key)).length)
const missingScoredCount = computed(() => Math.max(totalScoredCount.value - answeredScoredCount.value, 0))

onLoad(query => {
	routeToken.value = `${query?.token || ''}`.trim()
	refreshTemplate()
})

async function refreshTemplate() {
	if (!routeToken.value) {
		loadError.value = '缺少照顾者报告入口参数'
		return
	}
	pageLoading.value = true
	loadError.value = ''
	try {
		const result = await getPEP3CaregiverReportTemplate(routeToken.value)
		pageData.value = result || {}
		applySubmission(result?.submission || {})
	}
	catch (error) {
		console.warn('load pep3 caregiver report failed', error)
		loadError.value = `${error?.message || '入口无效或已过期'}`.trim()
	}
	finally {
		pageLoading.value = false
	}
}

function applySubmission(submission = {}) {
	answers.value = clonePlainObject(submission?.answers || {})
	respondentName.value = `${submission?.respondentName || ''}`.trim()
	relationship.value = `${submission?.relationship || ''}`.trim()
	submitted.value = !!submission?.submittedAt
}

function clonePlainObject(value = {}) {
	try {
		return JSON.parse(JSON.stringify(value || {}))
	}
	catch (error) {
		return {}
	}
}

function getAnswer(sectionCode, key) {
	const value = answers.value?.[sectionCode]?.[key]
	if (value === undefined || value === null) {
		return ''
	}
	return `${value}`
}

function setAnswer(sectionCode, key, value) {
	const next = clonePlainObject(answers.value)
	if (!next[sectionCode]) {
		next[sectionCode] = {}
	}
	const normalized = `${value ?? ''}`.trim()
	if (!normalized) {
		delete next[sectionCode][key]
	}
	else {
		next[sectionCode][key] = normalized
	}
	answers.value = next
	submitted.value = false
}

function setRespondentName(value) {
	respondentName.value = `${value || ''}`.trim()
	submitted.value = false
}

function setRelationship(value) {
	relationship.value = `${value || ''}`.trim()
	submitted.value = false
}

function getDiagnosisAnswer(categoryKey, fieldKey) {
	const value = answers.value?.diagnosis?.[categoryKey]
	if (!value || typeof value !== 'object') {
		return ''
	}
	return `${value[fieldKey] || ''}`.trim()
}

function setDiagnosisAnswer(categoryKey, fieldKey, value) {
	const next = clonePlainObject(answers.value)
	if (!next.diagnosis) {
		next.diagnosis = {}
	}
	if (!next.diagnosis[categoryKey] || typeof next.diagnosis[categoryKey] !== 'object') {
		next.diagnosis[categoryKey] = {}
	}
	next.diagnosis[categoryKey][fieldKey] = `${value || ''}`.trim()
	answers.value = next
	submitted.value = false
}

function sectionScoredCount(section) {
	if (!section?.scored) {
		return 0
	}
	return (section.items || []).filter(item => item.scored).length
}

function sectionAnsweredCount(section) {
	if (!section?.scored) {
		return 0
	}
	return (section.items || []).filter(item => item.scored && !!getAnswer(section.sectionCode, item.key)).length
}

function firstMissingScoredItem() {
	return scoredItems.value.find(({ sectionCode, item }) => !getAnswer(sectionCode, item.key))
}

async function submitReport() {
	if (submitting.value || submitted.value) {
		return
	}
	const missing = firstMissingScoredItem()
	if (missing) {
		uni.showToast({
			title: `请先完成第${missing.item.itemNo}题`,
			icon: 'none'
		})
		return
	}

	submitting.value = true
	try {
		await submitPEP3CaregiverReport({
			token: routeToken.value,
			respondentName: respondentName.value.trim(),
			relationship: relationship.value.trim(),
			answers: answers.value
		})
		submitted.value = true
		uni.showToast({
			title: '已提交',
			icon: 'success'
		})
	}
	catch (error) {
		console.warn('submit pep3 caregiver report failed', error)
		uni.showToast({
			title: `${error?.message || '提交失败'}`.slice(0, 24),
			icon: 'none'
		})
	}
	finally {
		submitting.value = false
	}
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
</script>

<style lang="scss" scoped>
.caregiver-report-page {
	padding-bottom: 40rpx;
}

.caregiver-nav-fixed {
	position: fixed;
	z-index: 20;
	top: 0;
	right: 0;
	left: 0;
	background: rgba(255, 250, 240, 0.94);
	backdrop-filter: blur(20rpx);
}

.caregiver-nav-fixed__inner {
	padding: 0 24rpx;
}

.caregiver-nav__side {
	display: flex;
	align-items: center;
}

.caregiver-back {
	display: flex;
	align-items: center;
	justify-content: center;
	width: 64rpx;
	height: 64rpx;
	border-radius: 50%;
	background: rgba(255, 255, 255, 0.82);
}

.caregiver-back__icon {
	width: 18rpx;
	height: 18rpx;
	border-bottom: 4rpx solid #242424;
	border-left: 4rpx solid #242424;
	transform: rotate(45deg);
}

.caregiver-nav__title {
	flex: 1;
	text-align: center;
	color: var(--parent-text);
	font-size: 30rpx;
	font-weight: 700;
}

.caregiver-shell {
	display: flex;
	flex-direction: column;
	gap: 22rpx;
}

.caregiver-empty-card {
	margin-top: 80rpx;
}

.caregiver-hero {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 24rpx;
	padding: 12rpx 6rpx 8rpx;
}

.caregiver-hero__title {
	display: block;
	color: var(--parent-text);
	font-size: 46rpx;
	font-weight: 800;
	line-height: 1.18;
}

.caregiver-hero__subtitle {
	display: block;
	margin-top: 10rpx;
	color: var(--parent-subtext);
	font-size: 24rpx;
	line-height: 1.5;
}

.caregiver-progress {
	display: flex;
	align-items: baseline;
	justify-content: center;
	min-width: 128rpx;
	height: 92rpx;
	padding: 18rpx 20rpx;
	border-radius: 28rpx;
	background: rgba(255, 255, 255, 0.9);
	box-shadow: var(--parent-shadow);
	color: #7b6210;
	font-weight: 800;

	text:first-child {
		font-size: 36rpx;
	}

	text:last-child {
		font-size: 22rpx;
	}
}

.caregiver-basic-card,
.caregiver-section-card,
.caregiver-success-card {
	padding: 26rpx;
}

.caregiver-success-card {
	background: rgba(234, 251, 242, 0.92);
	border-color: rgba(185, 237, 207, 0.82);
}

.caregiver-success-card__title {
	display: block;
	color: #0f8f52;
	font-size: 30rpx;
	font-weight: 800;
}

.caregiver-success-card__desc {
	display: block;
	margin-top: 8rpx;
	color: #397c5a;
	font-size: 24rpx;
	line-height: 1.5;
}

.caregiver-section-card__header {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 16rpx;
}

.caregiver-section-title {
	color: var(--parent-text);
	font-size: 30rpx;
	font-weight: 800;
	line-height: 1.35;
}

.caregiver-section-count {
	flex-shrink: 0;
	padding: 8rpx 16rpx;
	border-radius: 999rpx;
	background: rgba(255, 214, 10, 0.18);
	color: #806500;
	font-size: 22rpx;
	font-weight: 700;
}

.caregiver-section-desc {
	display: block;
	margin-top: 12rpx;
	color: var(--parent-subtext);
	font-size: 24rpx;
	line-height: 1.6;
}

.caregiver-field {
	margin-top: 22rpx;
}

.caregiver-field__label {
	display: block;
	margin-bottom: 10rpx;
	color: var(--parent-subtext);
	font-size: 24rpx;
	font-weight: 600;
}

.caregiver-input {
	width: 100%;
	height: 78rpx;
	padding: 0 22rpx;
	border: 1rpx solid var(--parent-card-border);
	border-radius: 20rpx;
	background: rgba(255, 255, 255, 0.9);
	color: var(--parent-text);
	font-size: 26rpx;
}

.caregiver-item-list,
.diagnosis-list {
	display: flex;
	flex-direction: column;
	gap: 18rpx;
	margin-top: 20rpx;
}

.caregiver-item,
.diagnosis-item {
	padding: 22rpx;
	border: 1rpx solid rgba(240, 235, 226, 0.95);
	border-radius: 22rpx;
	background: rgba(255, 255, 255, 0.82);
}

.caregiver-item__prompt-row {
	display: flex;
	gap: 14rpx;
}

.caregiver-item__no {
	display: flex;
	align-items: center;
	justify-content: center;
	flex-shrink: 0;
	width: 42rpx;
	height: 42rpx;
	border-radius: 50%;
	background: rgba(255, 214, 10, 0.22);
	color: #7d6400;
	font-size: 22rpx;
	font-weight: 800;
}

.caregiver-item__prompt {
	flex: 1;
	color: var(--parent-text);
	font-size: 26rpx;
	font-weight: 650;
	line-height: 1.55;
}

.caregiver-number-wrap {
	display: flex;
	align-items: center;
	gap: 12rpx;
	margin-top: 18rpx;
}

.caregiver-input--number {
	flex: 1;
}

.caregiver-input-unit {
	color: var(--parent-subtext);
	font-size: 24rpx;
}

.caregiver-radio-grid {
	display: flex;
	flex-direction: column;
	gap: 14rpx;
	margin-top: 18rpx;
}

.caregiver-radio-grid--compact {
	flex-direction: row;
	flex-wrap: wrap;
	gap: 12rpx 18rpx;
	margin-top: 8rpx;
}

.caregiver-radio-option {
	display: flex;
	align-items: flex-start;
	gap: 10rpx;
	color: #363226;
	font-size: 24rpx;
	line-height: 1.45;
}

.caregiver-radio-option radio {
	margin-top: 2rpx;
	transform: scale(0.82);
}

.caregiver-textarea {
	width: 100%;
	min-height: 150rpx;
	margin-top: 18rpx;
	padding: 18rpx 20rpx;
	border: 1rpx solid var(--parent-card-border);
	border-radius: 20rpx;
	background: rgba(255, 255, 255, 0.9);
	color: var(--parent-text);
	font-size: 26rpx;
	line-height: 1.5;
}

.diagnosis-item__title {
	display: block;
	color: var(--parent-text);
	font-size: 26rpx;
	font-weight: 750;
}

.diagnosis-item__group {
	margin-top: 16rpx;
}

.diagnosis-item__label {
	display: block;
	color: var(--parent-subtext);
	font-size: 22rpx;
	font-weight: 650;
}

.caregiver-submit-bar {
	position: sticky;
	bottom: 0;
	z-index: 10;
	margin: 8rpx -24rpx -56rpx;
	padding: 18rpx 24rpx 30rpx;
	background: linear-gradient(180deg, rgba(255, 255, 255, 0), rgba(255, 250, 240, 0.98) 22%);
}

.caregiver-submit-button {
	width: 100%;
}

.caregiver-submit-hint {
	display: block;
	margin-top: 12rpx;
	text-align: center;
	color: var(--parent-subtext);
	font-size: 22rpx;
}
</style>
