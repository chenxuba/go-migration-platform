<template>
	<view class="parent-page">
		<view class="parent-shell parent-shell--with-footer">
			<view class="parent-header select-header" :style="{ paddingTop: `${nav.top}px` }">
				<view class="parent-nav-row select-nav" :style="{ minHeight: `${nav.height}px` }">
					<view class="select-nav__side" :style="{ width: `${nav.width}px`, minHeight: `${nav.height}px` }">
						<view class="select-back" @click="handleBack">
							<view class="select-back__icon"></view>
						</view>
					</view>
					<text class="select-nav__title">关注学员</text>
					<view class="parent-nav-spacer" :style="{ width: `${nav.width}px`, height: `${nav.height}px` }"></view>
				</view>

				<view class="select-hero">
					<text class="select-hero__title">选择关联学员</text>
					<text class="select-hero__subtitle">{{ maskedPhone }} 已匹配到以下孩子</text>
					<view class="select-hero__tag">选中后即可同步课表、通知和订单信息</view>
				</view>
			</view>

			<view
				v-for="item in pendingCandidates"
				:key="item.id"
				class="parent-card select-student-card"
				:class="{ 'select-student-card--active': selectedIds.includes(item.id) }"
				@click="toggleSelect(item.id)"
			>
				<view class="select-student-card__main">
					<view class="select-student-card__avatar" :style="{ background: item.avatarColor }">
						{{ item.name.slice(0, 1) }}
					</view>
					<view class="select-student-card__content">
						<view class="select-student-card__title-row">
							<text class="select-student-card__name">{{ item.name }}</text>
							<view
								class="select-student-card__status"
								:class="{ 'select-student-card__status--active': selectedIds.includes(item.id) }"
							>
								<view
									class="select-student-card__status-dot"
									:class="{ 'select-student-card__status-dot--active': selectedIds.includes(item.id) }"
								></view>
								<text>{{ selectedIds.includes(item.id) ? '已选中' : '点击选择' }}</text>
							</view>
						</view>
						<text class="select-student-card__campus">{{ item.campusName }}</text>
						<view class="select-student-card__tags">
							<text class="select-student-card__tag select-student-card__tag--warm">{{ item.relation }}</text>
							<text class="select-student-card__tag select-student-card__tag--blue">{{ item.classLabel }}</text>
						</view>
					</view>
				</view>
			</view>

			<view v-if="!pendingCandidates.length" class="parent-card parent-empty-card select-empty-card">
				<view class="parent-empty-badge">空</view>
				<text class="parent-empty-title">当前号码暂无待关注学员</text>
				<text class="parent-empty-desc">可以返回上一页，或换个手机号继续尝试。</text>
			</view>
		</view>

		<view class="select-footer">
			<view class="select-footer__panel">
				<view class="parent-primary-button" @click="confirmBinding">{{ confirmButtonText }}</view>
				<view class="select-footer__switch" @click="changePhone">换个手机号试试</view>
			</view>
		</view>
	</view>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { getNavLayout } from '@/common/nav-layout'
import { confirmStudentBinding, logoutParent, parentState, setPendingSelection } from '@/common/parent-state'

const nav = getNavLayout()
const selectedIds = ref([])

const pendingCandidates = computed(() => parentState.pendingCandidates)
const maskedPhone = computed(() => parentState.profile.maskedPhone)
const confirmButtonText = computed(() => {
	return selectedIds.value.length ? `确认关注（${selectedIds.value.length}）` : '确认关注'
})

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
.parent-shell--with-footer {
	padding-bottom: 218rpx;
}

.select-nav__side {
	display: flex;
	align-items: center;
}

.select-nav__title {
	flex: 1;
	text-align: center;
	font-size: 34rpx;
	font-weight: 700;
	line-height: 1;
}

.select-back {
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

.select-back__icon {
	width: 16rpx;
	height: 16rpx;
	border-left: 4rpx solid #433e33;
	border-bottom: 4rpx solid #433e33;
	transform: translateX(4rpx) rotate(45deg);
}

.select-hero {
	margin-top: 24rpx;
	padding: 0 4rpx;
}

.select-hero__title {
	display: block;
	font-size: 36rpx;
	font-weight: 700;
	line-height: 1.24;
	color: #262626;
}

.select-hero__subtitle {
	display: block;
	margin-top: 12rpx;
	font-size: 24rpx;
	line-height: 1.6;
	color: var(--parent-subtext);
}

.select-hero__tag {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	margin-top: 16rpx;
	height: 44rpx;
	padding: 0 18rpx;
	border-radius: 999rpx;
	background: rgba(255, 244, 209, 0.86);
	color: #7d6200;
	font-size: 20rpx;
	font-weight: 600;
}

.select-student-card {
	margin-top: 16rpx;
	padding: 22rpx 20rpx;
	border: 2rpx solid transparent;
	transition: all 0.2s ease;
}

.select-student-card--active {
	background: linear-gradient(135deg, rgba(255, 247, 213, 0.9) 0%, rgba(255, 255, 255, 0.96) 100%);
	border-color: rgba(255, 214, 10, 0.94);
	box-shadow: 0 18rpx 36rpx rgba(255, 214, 10, 0.12);
}

.select-student-card__main {
	display: flex;
	align-items: flex-start;
	gap: 18rpx;
}

.select-student-card__avatar {
	width: 76rpx;
	height: 76rpx;
	border-radius: 24rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	color: #ffffff;
	font-size: 28rpx;
	font-weight: 700;
	box-shadow: 0 12rpx 24rpx rgba(71, 109, 178, 0.14);
	flex-shrink: 0;
}

.select-student-card__content {
	flex: 1;
	min-width: 0;
}

.select-student-card__title-row {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 16rpx;
}

.select-student-card__name {
	font-size: 30rpx;
	font-weight: 700;
	line-height: 1.35;
	color: #242424;
}

.select-student-card__status {
	display: inline-flex;
	align-items: center;
	gap: 8rpx;
	height: 42rpx;
	padding: 0 14rpx;
	border-radius: 999rpx;
	background: rgba(245, 241, 231, 0.96);
	color: #8c856f;
	font-size: 20rpx;
	font-weight: 600;
	flex-shrink: 0;
}

.select-student-card__status--active {
	background: rgba(255, 222, 75, 0.22);
	color: #7d6200;
}

.select-student-card__status-dot {
	width: 12rpx;
	height: 12rpx;
	border-radius: 50%;
	background: rgba(184, 178, 164, 0.88);
}

.select-student-card__status-dot--active {
	background: var(--parent-brand-deep);
	box-shadow: 0 0 0 6rpx rgba(255, 214, 10, 0.16);
}

.select-student-card__campus {
	display: block;
	margin-top: 10rpx;
	font-size: 22rpx;
	line-height: 1.55;
	color: var(--parent-subtext);
}

.select-student-card__tags {
	display: flex;
	align-items: center;
	gap: 10rpx;
	margin-top: 14rpx;
}

.select-student-card__tag {
	display: inline-flex;
	align-items: center;
	justify-content: center;
	height: 40rpx;
	padding: 0 14rpx;
	border-radius: 999rpx;
	font-size: 20rpx;
	font-weight: 600;
}

.select-student-card__tag--warm {
	background: rgba(255, 222, 75, 0.18);
	color: #7d6200;
}

.select-student-card__tag--blue {
	background: rgba(47, 134, 255, 0.12);
	color: var(--parent-blue);
}

.select-empty-card {
	margin-top: 16rpx;
}

.select-footer {
	position: fixed;
	left: 0;
	right: 0;
	bottom: 0;
	padding: 18rpx 24rpx calc(env(safe-area-inset-bottom) + 28rpx);
	background: linear-gradient(180deg, rgba(255, 255, 255, 0) 0%, rgba(255, 253, 248, 0.92) 30%, #ffffff 100%);
}

.select-footer__panel {
	padding: 16rpx;
	border-radius: 30rpx 30rpx 0 0;
	background: rgba(255, 255, 255, 0.88);
	box-shadow: 0 -8rpx 24rpx rgba(159, 124, 67, 0.08);
	backdrop-filter: blur(14rpx);
}

.select-footer__switch {
	margin-top: 16rpx;
	text-align: center;
	font-size: 22rpx;
	color: #8c856f;
	font-weight: 600;
}
</style>
