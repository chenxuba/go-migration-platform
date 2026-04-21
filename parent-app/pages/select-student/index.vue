<template>
	<view class="parent-page">
		<view class="parent-shell">
			<view class="parent-header select-header" :style="{ paddingTop: `${statusBarHeight + 18}px` }">
				<view class="select-header__back" @click="handleBack">‹</view>
				<view>
					<text class="select-title">Hi，欢迎来到家校助手</text>
					<text class="select-subtitle">{{ maskedPhone }} 下待关注学员</text>
				</view>
			</view>

			<view
				v-for="item in pendingCandidates"
				:key="item.id"
				class="parent-card select-student-card"
				:class="{ 'select-student-card--active': selectedIds.includes(item.id) }"
				@click="toggleSelect(item.id)"
			>
				<view class="select-student-card__body">
					<view>
						<text class="select-student-card__name">{{ item.name }}</text>
						<text class="select-student-card__campus">{{ item.campusName }}</text>
					</view>
					<view class="select-student-card__check">
						{{ selectedIds.includes(item.id) ? '√' : '' }}
					</view>
				</view>
			</view>

			<view v-if="!pendingCandidates.length" class="parent-card parent-empty-card">
				<view class="parent-empty-badge">空</view>
				<text class="parent-empty-title">当前号码暂无待关注学员</text>
				<text class="parent-empty-desc">你可以返回上一页或换个手机号继续尝试。</text>
			</view>
		</view>

		<view class="select-footer">
			<view class="parent-primary-button" @click="confirmBinding">确认关注</view>
			<view class="select-footer__switch" @click="changePhone">换个手机号试试</view>
		</view>
	</view>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { confirmStudentBinding, logoutParent, parentState, setPendingSelection } from '@/common/parent-state'

const statusBarHeight = uni.getSystemInfoSync().statusBarHeight || 0
const selectedIds = ref([])

const pendingCandidates = computed(() => parentState.pendingCandidates)
const maskedPhone = computed(() => parentState.profile.maskedPhone)

watch(
	() => parentState.selectedCandidateIds,
	value => {
		selectedIds.value = [...value]
	},
	{ immediate: true }
)

function toggleSelect(studentId) {
	if (selectedIds.value.includes(studentId)) {
		selectedIds.value = selectedIds.value.filter(item => item !== studentId)
	} else {
		selectedIds.value = [...selectedIds.value, studentId]
	}
}

function confirmBinding() {
	if (!selectedIds.value.length) {
		uni.showToast({
			title: '请先选择学员',
			icon: 'none'
		})
		return
	}
	setPendingSelection(selectedIds.value)
	const nextPage = confirmStudentBinding()
	if (!nextPage) {
		return
	}
	uni.switchTab({
		url: nextPage
	})
}

function handleBack() {
	uni.navigateBack({
		fail() {
			uni.switchTab({
				url: '/pages/profile/index'
			})
		}
	})
}

function changePhone() {
	logoutParent()
	uni.switchTab({
		url: '/pages/profile/index'
	})
}
</script>

<style scoped>
.select-header {
	display: flex;
	gap: 20rpx;
}

.select-header__back {
	width: 74rpx;
	height: 74rpx;
	border-radius: 50%;
	background: rgba(255, 255, 255, 0.72);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 44rpx;
	color: #1f1f1f;
	flex-shrink: 0;
}

.select-title {
	display: block;
	font-size: 60rpx;
	font-weight: 700;
	line-height: 1.2;
}

.select-subtitle {
	display: block;
	margin-top: 26rpx;
	font-size: 28rpx;
	font-weight: 600;
	color: #6e6c66;
}

.select-student-card {
	margin-top: 18rpx;
	padding: 22rpx;
	border: 2rpx solid transparent;
}

.select-student-card--active {
	border-color: rgba(255, 214, 10, 0.95);
}

.select-student-card__body {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 20rpx;
}

.select-student-card__name {
	display: block;
	font-size: 34rpx;
	font-weight: 700;
}

.select-student-card__campus {
	display: block;
	margin-top: 12rpx;
	font-size: 24rpx;
	line-height: 1.45;
	color: var(--parent-subtext);
}

.select-student-card__check {
	width: 48rpx;
	height: 48rpx;
	border-radius: 50%;
	background: rgba(34, 191, 115, 0.14);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 28rpx;
	font-weight: 700;
	color: var(--parent-green);
	flex-shrink: 0;
}

.select-footer {
	position: fixed;
	left: 0;
	right: 0;
	bottom: 0;
	padding: 18rpx 36rpx calc(env(safe-area-inset-bottom) + 30rpx);
	background: linear-gradient(180deg, rgba(255, 255, 255, 0) 0%, #ffffff 28%);
}

.select-footer__switch {
	margin-top: 18rpx;
	text-align: center;
	font-size: 24rpx;
	color: var(--parent-brand-deep);
	font-weight: 600;
}
</style>
