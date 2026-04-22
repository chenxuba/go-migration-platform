<template>
	<view class="parent-page">
		<view class="parent-shell settings-shell">
			<view class="parent-header settings-header" :style="{ paddingTop: `${nav.top}px` }">
				<view class="parent-nav-row settings-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="settings-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="settings-back" @click="goBack">
							<view class="settings-back__icon"></view>
						</view>
					</view>
					<text class="settings-nav__title">设置</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>

				<view class="parent-card settings-summary-card">
					<text class="settings-summary-card__label">当前账号</text>
					<text class="settings-summary-card__title">{{ summaryTitle }}</text>
					<text class="settings-summary-card__desc">{{ summaryDesc }}</text>
				</view>
			</view>

			<view class="parent-card settings-list-card">
				<view
					class="settings-item"
					:class="{ 'settings-item--disabled': !isAuthenticated || canceling }"
					@click="openCancelConfirm"
				>
					<view class="settings-item__copy">
						<text class="settings-item__title">注销账号</text>
						<text class="settings-item__desc">清除当前微信绑定，后续可由新的微信重新绑定</text>
					</view>
					<text class="settings-item__arrow">›</text>
				</view>
			</view>

			<view class="settings-tip-card">
				<text class="settings-tip-card__text">注销后会退出当前账号，并移除当前微信对应的公众号/学员绑定关系，请谨慎操作。</text>
			</view>
		</view>

		<view v-if="showCancelConfirm" class="settings-modal-mask" @touchmove.stop.prevent>
			<view class="settings-modal" @click.stop>
				<text class="settings-modal__title">注销账号？</text>
				<text class="settings-modal__desc">注销后，将退出当前账号并清除当前微信绑定数据，请谨慎操作</text>
				<view class="settings-modal__actions">
					<view
						class="settings-modal__action settings-modal__action--danger"
						:class="{ 'settings-modal__action--disabled': canceling }"
						@click="confirmCancelAccount"
					>
						{{ canceling ? '注销中...' : '注销账号' }}
					</view>
					<view
						class="settings-modal__action"
						:class="{ 'settings-modal__action--disabled': canceling }"
						@click="closeCancelConfirm"
					>
						取消
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { cancelParentAccount as cancelParentAccountRequest } from '@/common/parent-api'
import { getNavLayout } from '@/common/nav-layout'
import { logoutParent, parentState } from '@/common/parent-state'

const nav = getNavLayout()
const showCancelConfirm = ref(false)
const canceling = ref(false)

const isAuthenticated = computed(() => !!parentState.isAuthenticated && !!`${parentState.authToken || ''}`.trim())
const summaryTitle = computed(() => {
	if (!isAuthenticated.value) {
		return '当前未登录'
	}
	return `${parentState.profile.nickname || '微信家长'}`
})
const summaryDesc = computed(() => {
	if (!isAuthenticated.value) {
		return '请先登录后再管理账号'
	}
	return `${parentState.profile.maskedPhone || '未绑定手机号'}`
})

function resolveErrorMessage(error) {
	const message = `${error?.message || error || ''}`.trim()
	if (!message) {
		return '注销失败，请稍后重试'
	}
	return message.length > 30 ? message.slice(0, 30) : message
}

function openCancelConfirm() {
	if (canceling.value) {
		return
	}
	if (!isAuthenticated.value) {
		uni.showToast({
			title: '请先登录账号',
			icon: 'none'
		})
		return
	}
	showCancelConfirm.value = true
}

function closeCancelConfirm() {
	if (canceling.value) {
		return
	}
	showCancelConfirm.value = false
}

async function confirmCancelAccount() {
	if (canceling.value) {
		return
	}

	const token = `${parentState.authToken || ''}`.trim()
	const miniOpenId = `${parentState.miniOpenId || ''}`.trim()
	const unionId = `${parentState.unionId || ''}`.trim()
	if (!token || (!miniOpenId && !unionId)) {
		uni.showToast({
			title: '当前登录态已失效，请重新登录',
			icon: 'none'
		})
		return
	}

	canceling.value = true
	try {
		await cancelParentAccountRequest(token, {
			miniOpenId,
			unionId
		})
		showCancelConfirm.value = false
		logoutParent()
		uni.showToast({
			title: '账号已注销',
			icon: 'success'
		})
		setTimeout(() => {
			uni.switchTab({
				url: '/pages/profile/index'
			})
		}, 500)
	} catch (error) {
		uni.showToast({
			title: resolveErrorMessage(error),
			icon: 'none'
		})
	} finally {
		canceling.value = false
	}
}

function goBack() {
	if (canceling.value) {
		return
	}
	uni.navigateBack({
		fail() {
			uni.switchTab({
				url: '/pages/profile/index'
			})
		}
	})
}
</script>

<style scoped>
.settings-shell {
	padding-bottom: 80rpx;
}

.settings-header {
	padding-bottom: 10rpx;
}

.settings-nav__side {
	display: flex;
	align-items: center;
}

.settings-nav__title {
	flex: 1;
	text-align: center;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.settings-back {
	width: 64rpx;
	height: 64rpx;
	border-radius: 22rpx;
	background: rgba(255, 255, 255, 0.76);
	border: 1rpx solid rgba(255, 255, 255, 0.92);
	display: flex;
	align-items: center;
	justify-content: center;
	box-shadow: 0 12rpx 28rpx rgba(162, 130, 71, 0.08);
}

.settings-back__icon {
	width: 16rpx;
	height: 16rpx;
	border-left: 4rpx solid #433e33;
	border-bottom: 4rpx solid #433e33;
	transform: translateX(4rpx) rotate(45deg);
}

.settings-summary-card {
	margin-top: 24rpx;
	padding: 26rpx 22rpx;
	background:
		radial-gradient(circle at top right, rgba(255, 226, 122, 0.14), transparent 28%),
		linear-gradient(180deg, rgba(255, 255, 255, 0.98) 0%, rgba(255, 252, 245, 0.98) 100%);
}

.settings-summary-card__label {
	display: block;
	font-size: 22rpx;
	font-weight: 700;
	letter-spacing: 2rpx;
	color: #8f8570;
}

.settings-summary-card__title {
	display: block;
	margin-top: 14rpx;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1.25;
	color: #1f1f1f;
}

.settings-summary-card__desc {
	display: block;
	margin-top: 10rpx;
	font-size: 24rpx;
	line-height: 1.6;
	color: var(--parent-subtext);
}

.settings-list-card {
	margin-top: 28rpx;
	padding: 0 18rpx;
	overflow: hidden;
}

.settings-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 18rpx;
	min-height: 108rpx;
}

.settings-item--disabled {
	opacity: 0.46;
}

.settings-item__copy {
	flex: 1;
	min-width: 0;
}

.settings-item__title {
	display: block;
	font-size: 28rpx;
	font-weight: 700;
	line-height: 1.35;
	color: #202020;
}

.settings-item__desc {
	display: block;
	margin-top: 10rpx;
	font-size: 22rpx;
	line-height: 1.6;
	color: var(--parent-subtext);
}

.settings-item__arrow {
	color: #bbb29f;
	font-size: 30rpx;
	flex-shrink: 0;
}

.settings-tip-card {
	margin-top: 20rpx;
	padding: 0 6rpx;
}

.settings-tip-card__text {
	display: block;
	font-size: 22rpx;
	line-height: 1.7;
	color: #8d8069;
}

.settings-modal-mask {
	position: fixed;
	inset: 0;
	background: rgba(23, 21, 16, 0.46);
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 0 52rpx;
	z-index: 90;
}

.settings-modal {
	width: 100%;
	border-radius: 34rpx;
	background: rgba(255, 255, 255, 0.98);
	overflow: hidden;
	box-shadow: 0 34rpx 80rpx rgba(41, 31, 10, 0.2);
}

.settings-modal__title {
	display: block;
	padding: 42rpx 36rpx 0;
	text-align: center;
	font-size: 40rpx;
	font-weight: 700;
	line-height: 1.3;
	color: #2b2b2b;
}

.settings-modal__desc {
	display: block;
	padding: 24rpx 46rpx 40rpx;
	text-align: center;
	font-size: 27rpx;
	line-height: 1.65;
	color: #8a8477;
}

.settings-modal__actions {
	display: flex;
	align-items: stretch;
	border-top: 1rpx solid rgba(232, 227, 218, 0.9);
}

.settings-modal__action {
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: center;
	min-height: 104rpx;
	font-size: 30rpx;
	font-weight: 700;
	color: #232323;
}

.settings-modal__action + .settings-modal__action {
	border-left: 1rpx solid rgba(232, 227, 218, 0.9);
}

.settings-modal__action--danger {
	color: #ff5f52;
}

.settings-modal__action--disabled {
	opacity: 0.5;
}
</style>
