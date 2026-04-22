<template>
	<view class="rehab-signature-page" :style="{ paddingTop: `${pageTopInset}px`, paddingBottom: `${pageBottomInset}px` }">
		<view class="rehab-signature-page__board">
			<view class="rehab-signature-page__tip-rail">
				<view class="rehab-signature-page__tip-rotated">
					<text class="rehab-signature-page__tip">请在签名区域内手写签名，完成后会自动返回康复记录详情。</text>
				</view>
			</view>

			<view class="rehab-signature-page__canvas-panel">
				<canvas
					:id="signatureCanvasId"
					:canvas-id="signatureCanvasId"
					class="rehab-signature-page__canvas"
					:style="{ width: `${signatureCanvasWidth}px`, height: `${signatureCanvasHeight}px` }"
					:width="signatureCanvasWidth"
					:height="signatureCanvasHeight"
					disable-scroll="true"
					@touchstart.stop.prevent="handleSignatureTouchStart"
					@touchmove.stop.prevent="handleSignatureTouchMove"
					@touchend.stop.prevent="handleSignatureTouchEnd"
					@touchcancel.stop.prevent="handleSignatureTouchEnd"
				></canvas>
			</view>

			<view class="rehab-signature-page__side-rail">
				<view class="rehab-signature-page__side-rotated">
					<text class="rehab-signature-page__title">在线签名</text>
					<view class="rehab-signature-page__actions">
						<view class="rehab-signature-page__action" @click="resetSignatureCanvas">
							<text>重写</text>
						</view>
						<view class="rehab-signature-page__action" @click="goBack">
							<text>取消</text>
						</view>
						<view class="rehab-signature-page__action rehab-signature-page__action--primary" @click="confirmSignature">
							<text>完成</text>
						</view>
					</view>
				</view>
			</view>

			<canvas
				:id="signatureExportCanvasId"
				:canvas-id="signatureExportCanvasId"
				class="rehab-signature-page__export-canvas"
				:style="{ width: `${signatureExportCanvasWidth}px`, height: `${signatureExportCanvasHeight}px` }"
				:width="signatureExportCanvasWidth"
				:height="signatureExportCanvasHeight"
			></canvas>
		</view>
	</view>
</template>

<script setup>
import { getCurrentInstance, nextTick, ref } from 'vue'
import { onReady } from '@dcloudio/uni-app'
import { getNavLayout } from '@/common/nav-layout'

const instance = getCurrentInstance()
const nav = getNavLayout()
const pageTopInset = ref(nav.top + nav.height + 8)
const pageBottomInset = ref(26)

const signatureCanvasId = 'parentRehabSignatureCanvasPage'
const signatureExportCanvasId = 'parentRehabSignatureExportCanvasPage'
const signatureCanvasWidth = ref(0)
const signatureCanvasHeight = ref(0)
const signatureCanvasReady = ref(false)
const signatureHasStroke = ref(false)
const signatureDraftStorageKey = 'parent_rehab_signature_draft'
const signatureExportCanvasWidth = ref(1)
const signatureExportCanvasHeight = ref(1)

let signatureContext = null
let signatureExportContext = null
let signatureCanvasRect = null
let signatureLastPoint = null
let signatureDrawQueue = Promise.resolve()
let openerEventChannel = null

onReady(() => {
	nextTick(() => {
		initSignatureCanvas()
	})
})

function initSignatureCanvas() {
	const systemInfo = uni.getSystemInfoSync()
	const windowWidth = Number(systemInfo?.windowWidth || 667)
	const windowHeight = Number(systemInfo?.windowHeight || 812)
	const safeBottom = Math.max(0, Number(systemInfo?.safeAreaInsets?.bottom || 0))
	pageBottomInset.value = safeBottom + 26
	signatureCanvasWidth.value = Math.max(240, Math.floor(windowWidth - 132))
	signatureCanvasHeight.value = Math.max(420, Math.floor(windowHeight - pageTopInset.value - pageBottomInset.value - 52))
	signatureContext = uni.createCanvasContext(signatureCanvasId, instance?.proxy)
	signatureExportContext = uni.createCanvasContext(signatureExportCanvasId, instance?.proxy)
	signatureCanvasReady.value = true
	resetSignatureCanvas()
	queryCanvasRect()
}

function queryCanvasRect() {
	const query = uni.createSelectorQuery()
	if (instance?.proxy) {
		query.in(instance.proxy)
	}
	query.select(`#${signatureCanvasId}`).boundingClientRect(rect => {
		signatureCanvasRect = rect || null
	}).exec()
}

function resetSignatureCanvas() {
	if (!signatureContext || !signatureCanvasWidth.value || !signatureCanvasHeight.value) {
		return
	}
	signatureHasStroke.value = false
	signatureLastPoint = null
	signatureDrawQueue = Promise.resolve()
	queueSignatureCanvasDraw(false, () => {
		signatureContext.setFillStyle('#ffffff')
		signatureContext.fillRect(0, 0, signatureCanvasWidth.value, signatureCanvasHeight.value)
	})
}

function resolveOpenerEventChannel() {
	if (openerEventChannel) {
		return openerEventChannel
	}
	let channel = instance?.proxy?.getOpenerEventChannel?.()
	if (!channel || typeof channel.emit !== 'function') {
		const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []
		const currentPage = Array.isArray(pages) ? pages[pages.length - 1] : null
		channel = currentPage?.getOpenerEventChannel?.()
	}
	if (channel && typeof channel.emit === 'function') {
		openerEventChannel = channel
	}
	return openerEventChannel
}

function queueSignatureCanvasDraw(reserve, painter) {
	if (!signatureContext) {
		return Promise.resolve()
	}
	signatureDrawQueue = signatureDrawQueue
		.catch(() => undefined)
		.then(() => new Promise((resolve, reject) => {
			try {
				let settled = false
				const finish = () => {
					if (settled) {
						return
					}
					settled = true
					resolve()
				}
				const fallbackTimer = setTimeout(finish, 36)
				painter()
				signatureContext.draw(reserve, () => {
					clearTimeout(fallbackTimer)
					finish()
				})
			} catch (error) {
				reject(error)
			}
		}))
	return signatureDrawQueue
}

function resolveTouchPoint(event) {
	const detailX = Number(event?.detail?.x ?? NaN)
	const detailY = Number(event?.detail?.y ?? NaN)
	if (Number.isFinite(detailX) && Number.isFinite(detailY)) {
		return clampSignaturePoint(detailX, detailY)
	}

	const touch = event?.touches?.[0] || event?.changedTouches?.[0]
	if (!touch) {
		return null
	}

	const localX = Number(touch.x ?? NaN)
	const localY = Number(touch.y ?? NaN)
	if (Number.isFinite(localX) && Number.isFinite(localY)) {
		return clampSignaturePoint(localX, localY)
	}

	const pageX = Number(touch.pageX ?? touch.clientX ?? NaN)
	const pageY = Number(touch.pageY ?? touch.clientY ?? NaN)
	if (signatureCanvasRect && Number.isFinite(pageX) && Number.isFinite(pageY)) {
		return clampSignaturePoint(
			pageX - Number(signatureCanvasRect.left || 0),
			pageY - Number(signatureCanvasRect.top || 0)
		)
	}
	return null
}

function clampSignaturePoint(x, y) {
	return {
		x: Math.max(0, Math.min(signatureCanvasWidth.value, Number(x || 0))),
		y: Math.max(0, Math.min(signatureCanvasHeight.value, Number(y || 0)))
	}
}

function drawSignatureLine(from, to) {
	if (!signatureContext || !from || !to) {
		return
	}
	queueSignatureCanvasDraw(true, () => {
		signatureContext.setStrokeStyle('#2f2b25')
		signatureContext.setLineWidth(4)
		signatureContext.setLineCap('round')
		signatureContext.setLineJoin('round')
		signatureContext.beginPath()
		signatureContext.moveTo(from.x, from.y)
		signatureContext.lineTo(to.x, to.y)
		signatureContext.stroke()
	})
}

function handleSignatureTouchStart(event) {
	if (!signatureCanvasReady.value) {
		return
	}
	if (!signatureCanvasRect) {
		queryCanvasRect()
	}
	const point = resolveTouchPoint(event)
	if (!point) {
		return
	}
	signatureHasStroke.value = true
	signatureLastPoint = point
	drawSignatureLine(point, {
		x: point.x + 0.1,
		y: point.y + 0.1
	})
}

function handleSignatureTouchMove(event) {
	if (!signatureLastPoint) {
		return
	}
	const point = resolveTouchPoint(event)
	if (!point) {
		return
	}
	drawSignatureLine(signatureLastPoint, point)
	signatureLastPoint = point
}

function handleSignatureTouchEnd() {
	signatureLastPoint = null
}

async function confirmSignature() {
	if (!signatureHasStroke.value) {
		uni.showToast({
			title: '请先完成签名',
			icon: 'none'
		})
		return
	}
	try {
		await signatureDrawQueue.catch(() => undefined)
		await waitCanvasFlush()
		const rawTempFilePath = await exportSignatureTempFilePath()
		const tempFilePath = await rotateSignatureTempFilePath(rawTempFilePath)
		uni.setStorageSync(signatureDraftStorageKey, tempFilePath)
		const eventChannel = resolveOpenerEventChannel()
		if (eventChannel && typeof eventChannel.emit === 'function') {
			eventChannel.emit('parentRehabSignatureComplete', {
				tempFilePath
			})
		}
		setTimeout(() => {
			uni.navigateBack()
		}, 16)
	} catch (error) {
		console.warn('export signature temp file failed', error)
		uni.showToast({
			title: `${error?.message || '签名生成失败'}`.slice(0, 24),
			icon: 'none'
		})
	}
}

function exportSignatureTempFilePath() {
	return new Promise((resolve, reject) => {
		uni.canvasToTempFilePath({
			x: 0,
			y: 0,
			width: signatureCanvasWidth.value,
			height: signatureCanvasHeight.value,
			destWidth: signatureCanvasWidth.value * 2,
			destHeight: signatureCanvasHeight.value * 2,
			canvasId: signatureCanvasId,
			fileType: 'png',
			quality: 1,
			success: result => {
				const tempFilePath = `${result?.tempFilePath || ''}`.trim()
				if (!tempFilePath) {
					reject(new Error('签名内容为空'))
					return
				}
				resolve(tempFilePath)
			},
			fail: error => {
				reject(new Error(error?.errMsg || '签名导出失败'))
			}
		}, instance?.proxy)
	})
}

function getImageInfo(src) {
	return new Promise((resolve, reject) => {
		uni.getImageInfo({
			src,
			success: result => resolve(result || {}),
			fail: error => reject(new Error(error?.errMsg || '签名读取失败'))
		})
	})
}

function drawSignatureExportCanvas(painter) {
	if (!signatureExportContext) {
		return Promise.reject(new Error('签名导出画布未准备好'))
	}
	return new Promise((resolve, reject) => {
		try {
			painter()
			signatureExportContext.draw(false, () => {
				resolve()
			})
		} catch (error) {
			reject(error)
		}
	})
}

function exportRotatedSignatureTempFilePath() {
	return new Promise((resolve, reject) => {
		uni.canvasToTempFilePath({
			x: 0,
			y: 0,
			width: signatureExportCanvasWidth.value,
			height: signatureExportCanvasHeight.value,
			destWidth: signatureExportCanvasWidth.value * 2,
			destHeight: signatureExportCanvasHeight.value * 2,
			canvasId: signatureExportCanvasId,
			fileType: 'png',
			quality: 1,
			success: result => {
				const tempFilePath = `${result?.tempFilePath || ''}`.trim()
				if (!tempFilePath) {
					reject(new Error('签名内容为空'))
					return
				}
				resolve(tempFilePath)
			},
			fail: error => {
				reject(new Error(error?.errMsg || '签名导出失败'))
			}
		}, instance?.proxy)
	})
}

async function rotateSignatureTempFilePath(sourcePath) {
	const imageInfo = await getImageInfo(sourcePath)
	const sourceWidth = Math.max(1, Number(imageInfo?.width || signatureCanvasWidth.value))
	const sourceHeight = Math.max(1, Number(imageInfo?.height || signatureCanvasHeight.value))
	if (sourceWidth >= sourceHeight) {
		return sourcePath
	}
	signatureExportCanvasWidth.value = sourceHeight
	signatureExportCanvasHeight.value = sourceWidth
	await nextTick()
	signatureExportContext = uni.createCanvasContext(signatureExportCanvasId, instance?.proxy)
	await drawSignatureExportCanvas(() => {
		signatureExportContext.setFillStyle('#ffffff')
		signatureExportContext.fillRect(0, 0, signatureExportCanvasWidth.value, signatureExportCanvasHeight.value)
		signatureExportContext.translate(0, signatureExportCanvasHeight.value)
		signatureExportContext.rotate(-Math.PI / 2)
		signatureExportContext.drawImage(sourcePath, 0, 0, sourceWidth, sourceHeight)
	})
	await waitCanvasFlush()
	return exportRotatedSignatureTempFilePath()
}

function waitCanvasFlush() {
	return new Promise(resolve => {
		setTimeout(resolve, 80)
	})
}

function goBack() {
	uni.navigateBack()
}
</script>

<style scoped>
.rehab-signature-page {
	min-height: 100vh;
	padding-left: 18rpx;
	padding-right: 18rpx;
	box-sizing: border-box;
	background:
		radial-gradient(circle at 14% 18%, rgba(187, 226, 181, 0.18), transparent 22%),
		linear-gradient(180deg, #fff8ef 0%, #fffaf4 46%, #fffdf9 100%);
	display: flex;
	align-items: center;
	justify-content: center;
}

.rehab-signature-page__board {
	width: 100%;
	height: 100%;
	max-height: calc(100vh - 24rpx);
	padding: 18rpx 16rpx;
	border-radius: 36rpx;
	background: rgba(255, 253, 249, 0.98);
	border: 1rpx solid rgba(245, 235, 220, 0.92);
	box-shadow: 0 18rpx 48rpx rgba(63, 45, 12, 0.12);
	box-sizing: border-box;
	display: flex;
	align-items: stretch;
	justify-content: space-between;
	gap: 12rpx;
}

.rehab-signature-page__tip-rail {
	position: relative;
	width: 68rpx;
	display: flex;
	align-items: center;
	justify-content: center;
}

.rehab-signature-page__tip-rotated {
	position: absolute;
	left: 50%;
	top: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	transform: translate(-50%, -50%) rotate(90deg);
	transform-origin: center;
}

.rehab-signature-page__canvas-panel {
	flex: 1;
	display: flex;
	align-items: center;
	padding: 10rpx;
	border-radius: 32rpx;
	background: rgba(255, 252, 246, 0.96);
	border: 1rpx solid rgba(242, 233, 219, 0.96);
	box-shadow: inset 0 0 0 1rpx rgba(255, 255, 255, 0.6);
	display: flex;
	align-items: center;
	justify-content: center;
}

.rehab-signature-page__canvas {
	display: block;
	background: #ffffff;
	border-radius: 24rpx;
}

.rehab-signature-page__side-rail {
	position: relative;
	width: 92rpx;
	display: flex;
	align-items: center;
	justify-content: center;
	overflow: visible;
}

.rehab-signature-page__side-rotated {
	position: absolute;
	left: 50%;
	top: 50%;
	display: flex;
	align-items: center;
	justify-content: center;
	gap: 28rpx;
	transform: translate(-50%, -50%) rotate(90deg);
	transform-origin: center;
}

.rehab-signature-page__title {
	font-size: 32rpx;
	font-weight: 700;
	line-height: 1.1;
	color: #1f1f1f;
	flex-shrink: 0;
	letter-spacing: 1rpx;
}

.rehab-signature-page__actions {
	display: flex;
	align-items: center;
	gap: 18rpx;
}

.rehab-signature-page__action {
	min-width: 104rpx;
	height: 56rpx;
	padding: 0 20rpx;
	border-radius: 999rpx;
	background: rgba(255, 255, 255, 0.94);
	border: 1rpx solid rgba(237, 226, 211, 0.98);
	box-sizing: border-box;
	display: flex;
	align-items: center;
	justify-content: center;
	font-size: 24rpx;
	line-height: 1.3;
	color: #8f8576;
	box-shadow: 0 10rpx 20rpx rgba(168, 134, 77, 0.08);
}

.rehab-signature-page__action--primary {
	color: #b98905;
	font-weight: 700;
	background: linear-gradient(180deg, #fff4cc 0%, #ffe8a4 100%);
	border-color: rgba(228, 194, 101, 0.78);
}

.rehab-signature-page__tip {
	font-size: 20rpx;
	line-height: 1.4;
	color: #9a9081;
	white-space: nowrap;
}

.rehab-signature-page__export-canvas {
	position: fixed;
	left: -9999px;
	top: -9999px;
	opacity: 0;
	pointer-events: none;
}
</style>
