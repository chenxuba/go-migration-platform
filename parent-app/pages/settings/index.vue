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

				<view class="parent-card settings-account-card">
					<text class="settings-account-card__label">当前账号</text>
					<view class="settings-account-card__main">
						<view class="settings-account-card__avatar">
							<text class="settings-account-card__avatar-text">{{ accountInitial }}</text>
						</view>
						<view class="settings-account-card__copy">
							<text class="settings-account-card__title">{{ summaryTitle }}</text>
							<text class="settings-account-card__desc">{{ summaryDesc }}</text>
						</view>
						<view
							class="settings-account-card__status"
							:class="isAuthenticated ? 'settings-account-card__status--online' : 'settings-account-card__status--offline'"
						>
							<text class="settings-account-card__status-dot"></text>
							<text>{{ isAuthenticated ? '已登录' : '未登录' }}</text>
						</view>
					</view>
				</view>
			</view>

			<view class="parent-card settings-action-card">
				<view
					class="settings-item"
					:class="{ 'settings-item--disabled': !isAuthenticated || canceling }"
					@click="openCancelConfirm"
				>
					<view class="settings-item__icon-wrap">
						<view class="settings-item__icon">
							<text class="settings-item__icon-mark">!</text>
						</view>
					</view>
					<view class="settings-item__copy">
						<text class="settings-item__title">注销账号</text>
						<text class="settings-item__desc">停止当前微信接收推送，后续可由新微信重新绑定</text>
					</view>
					<text class="settings-item__arrow">›</text>
				</view>
			</view>

			<view class="parent-card settings-note-card">
				<text class="settings-note-card__title">注销后将会</text>
				<view class="settings-note-card__item">
					<text class="settings-note-card__dot">1</text>
					<text class="settings-note-card__text">当前微信不再接收学员通知</text>
				</view>
				<view class="settings-note-card__item">
					<text class="settings-note-card__dot">2</text>
					<text class="settings-note-card__text">新微信登录后可重新绑定</text>
				</view>
				<text class="settings-note-card__tip">如果只是暂时不用当前手机，不需要注销账号。</text>
			</view>
		</view>

		<view v-if="showCancelConfirm" class="settings-modal-mask" @touchmove.stop.prevent>
			<view class="settings-modal" @click.stop>
				<text class="settings-modal__title">确认注销当前账号？</text>
				<text class="settings-modal__desc">注销后，当前微信将不再接收通知；如需更换微信，可在新微信登录后重新绑定。</text>
				<view class="settings-modal__actions">
					<view
						class="settings-modal__action settings-modal__action--secondary"
						:class="{ 'settings-modal__action--disabled': canceling }"
						@click="closeCancelConfirm"
					>
						取消
					</view>
					<view
						class="settings-modal__action settings-modal__action--danger"
						:class="{ 'settings-modal__action--disabled': canceling }"
						@click="confirmCancelAccount"
					>
						{{ canceling ? '注销中...' : '确认注销' }}
					</view>
				</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { cancelParentAccount as cancelParentAccountRequest } from '@/common/parent-api'
import { repairCurrentParentWeChatIdentity } from '@/common/parent-auth'
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
const accountInitial = computed(() => {
	const nickname = `${parentState.profile.nickname || ''}`.trim()
	if (nickname) {
		return nickname.slice(0, 1)
	}
	return isAuthenticated.value ? '家' : '未'
})

function resolveErrorMessage(error) {
	const message = `${error?.message || error || ''}`.trim()
	if (!message) {
		return '注销失败，请稍后重试'
	}
	return message.length > 30 ? message.slice(0, 30) : message
}

function hasCurrentWeChatIdentity() {
	return !!`${parentState.miniOpenId || ''}`.trim() || !!`${parentState.unionId || ''}`.trim()
}

async function ensureCurrentWeChatIdentity() {
	if (hasCurrentWeChatIdentity()) {
		return true
	}

	const token = `${parentState.authToken || ''}`.trim()
	if (!token) {
		return false
	}

	await repairCurrentParentWeChatIdentity(token)
	return hasCurrentWeChatIdentity()
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

	canceling.value = true
	try {
		const token = `${parentState.authToken || ''}`.trim()
		if (!token) {
			throw new Error('当前登录态已失效，请重新登录')
		}

		const prepared = await ensureCurrentWeChatIdentity()
		if (!prepared) {
			throw new Error('当前登录态已失效，请重新登录')
		}

		const miniOpenId = `${parentState.miniOpenId || ''}`.trim()
		const unionId = `${parentState.unionId || ''}`.trim()
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
	padding-bottom: 14rpx;
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

.settings-account-card {
	margin-top: 24rpx;
	padding: 26rpx 24rpx;
	background:
		radial-gradient(circle at 100% 0%, rgba(255, 225, 127, 0.18), transparent 34%),
		linear-gradient(145deg, rgba(255, 255, 255, 0.99) 0%, rgba(255, 251, 244, 0.98) 100%);
}

.settings-account-card__label {
	display: block;
	font-size: 22rpx;
	font-weight: 700;
	letter-spacing: 2rpx;
	color: #958569;
}

.settings-account-card__main {
	display: flex;
	align-items: center;
	gap: 22rpx;
	margin-top: 14rpx;
}

.settings-account-card__avatar {
	width: 92rpx;
	height: 92rpx;
	border-radius: 28rpx;
	background: linear-gradient(160deg, rgba(255, 227, 133, 0.95) 0%, rgba(255, 205, 87, 0.92) 100%);
	box-shadow: 0 14rpx 28rpx rgba(236, 184, 61, 0.18);
	display: flex;
	align-items: center;
	justify-content: center;
}

.settings-account-card__avatar-text {
	font-size: 38rpx;
	font-weight: 800;
	color: #5f4300;
}

.settings-account-card__copy {
	flex: 1;
	min-width: 0;
}

.settings-account-card__status {
	flex-shrink: 0;
	display: inline-flex;
	align-items: center;
	gap: 8rpx;
	padding: 10rpx 14rpx;
	border-radius: 999rpx;
	font-size: 20rpx;
	font-weight: 700;
}

.settings-account-card__status--online {
	background: rgba(34, 191, 115, 0.12);
	color: #169456;
}

.settings-account-card__status--offline {
	background: rgba(143, 133, 112, 0.12);
	color: #8d8069;
}

.settings-account-card__status-dot {
	width: 10rpx;
	height: 10rpx;
	border-radius: 50%;
	background: currentColor;
}

.settings-account-card__title {
	display: block;
	font-size: 36rpx;
	font-weight: 700;
	line-height: 1.25;
	color: #1f1f1f;
}

.settings-account-card__desc {
	display: block;
	margin-top: 8rpx;
	font-size: 24rpx;
	line-height: 1.6;
	color: var(--parent-subtext);
}

.settings-action-card {
	margin-top: 22rpx;
	padding: 0 20rpx;
}

.settings-item {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 18rpx;
	min-height: 124rpx;
}

.settings-item--disabled {
	opacity: 0.46;
}

.settings-item__icon-wrap {
	flex-shrink: 0;
}

.settings-item__icon {
	width: 72rpx;
	height: 72rpx;
	border-radius: 22rpx;
	background: linear-gradient(160deg, #ffb7ab 0%, #ff7b66 100%);
	box-shadow: 0 14rpx 26rpx rgba(255, 122, 102, 0.16);
	display: flex;
	align-items: center;
	justify-content: center;
}

.settings-item__icon-mark {
	font-size: 34rpx;
	font-weight: 800;
	color: #ffffff;
	line-height: 1;
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
	margin-top: 8rpx;
	font-size: 23rpx;
	line-height: 1.6;
	color: var(--parent-subtext);
}

.settings-item__arrow {
	color: #c5af9f;
	font-size: 38rpx;
	flex-shrink: 0;
}

.settings-note-card {
	margin-top: 22rpx;
	padding: 24rpx;
}

.settings-note-card__title {
	display: block;
	font-size: 30rpx;
	font-weight: 700;
	line-height: 1.35;
	color: #22211f;
}

.settings-note-card__item {
	display: flex;
	align-items: center;
	gap: 16rpx;
	margin-top: 18rpx;
}

.settings-note-card__dot {
	width: 34rpx;
	height: 34rpx;
	border-radius: 50%;
	background: rgba(255, 214, 10, 0.16);
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 20rpx;
	font-weight: 800;
	color: #9b7600;
	flex-shrink: 0;
}

.settings-note-card__text {
	flex: 1;
	font-size: 23rpx;
	line-height: 1.6;
	color: #786e60;
}

.settings-note-card__tip {
	display: block;
	margin-top: 18rpx;
	font-size: 22rpx;
	line-height: 1.6;
	color: #938774;
}

.settings-modal-mask {
	position: fixed;
	inset: 0;
	background: rgba(23, 21, 16, 0.46);
	display: flex;
	align-items: center;
	justify-content: center;
	padding: 0 44rpx;
	z-index: 90;
}

.settings-modal {
	width: 100%;
	border-radius: 32rpx;
	background: rgba(255, 255, 255, 0.98);
	padding: 40rpx 30rpx 30rpx;
	box-shadow: 0 34rpx 80rpx rgba(41, 31, 10, 0.2);
}

.settings-modal__title {
	display: block;
	text-align: center;
	font-size: 38rpx;
	font-weight: 700;
	line-height: 1.3;
	color: #2b2b2b;
}

.settings-modal__desc {
	display: block;
	margin-top: 18rpx;
	text-align: center;
	font-size: 25rpx;
	line-height: 1.65;
	color: #8a8477;
}

.settings-modal__actions {
	display: flex;
	gap: 18rpx;
	margin-top: 26rpx;
}

.settings-modal__action {
	flex: 1;
	display: flex;
	align-items: center;
	justify-content: center;
	min-height: 88rpx;
	border-radius: 26rpx;
	font-size: 29rpx;
	font-weight: 700;
	color: #232323;
}

.settings-modal__action--danger {
	background: linear-gradient(160deg, #ff8f79 0%, #ff6d57 100%);
	box-shadow: 0 18rpx 34rpx rgba(255, 122, 102, 0.2);
	color: #ffffff;
}

.settings-modal__action--secondary {
	border: 1rpx solid rgba(232, 227, 218, 0.95);
	background: rgba(255, 255, 255, 0.98);
	color: #6d6559;
}

.settings-modal__action--disabled {
	opacity: 0.5;
}
</style>
