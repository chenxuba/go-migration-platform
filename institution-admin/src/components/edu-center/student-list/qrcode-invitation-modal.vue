<script setup>
import { computed, ref } from 'vue'
import { CloseOutlined } from '@ant-design/icons-vue'
import { useClipboard } from '@v-c/utils'
import invitationStep2Image from '@/assets/images/qrcode-invitation-step-2.png'
import messageService from '~@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  invitationInfo: {
    type: Object,
    default: () => ({}),
  },
})

const emit = defineEmits(['update:open'])
const { copy } = useClipboard()
const downloading = ref(false)

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const institutionName = computed(() => String(props.invitationInfo?.institutionName || '').trim() || '机构服务中心')
const officialAccountName = computed(() => String(props.invitationInfo?.officialAccountName || '').trim() || 'irts家校云')
const officialAccountDisplayName = computed(() => {
  return officialAccountName.value.includes('公众号')
    ? officialAccountName.value
    : `${officialAccountName.value}公众号`
})
const studentName = computed(() => String(props.invitationInfo?.studentName || '').trim() || '学员')
const qrCodeSrc = computed(() => String(props.invitationInfo?.qrCodeDataUrl || props.invitationInfo?.qrCodeUrl || '').trim())
const institutionInitial = computed(() => Array.from(institutionName.value.replace(/\s+/g, ''))[0] || '校')
const hasInvitationInfo = computed(() => Boolean(qrCodeSrc.value))

const invitationText = computed(() => {
  const displayName = institutionName.value === '机构服务中心' ? '我校' : institutionName.value
  return [
    '亲爱的家长：',
    `为了进一步提升${displayName}的家校服务质量，便于您及时接收学校通知，随时查看和反馈孩子的学习情况，我校将采用“irts家校云”作为家校服务工具，请各位家长扫描二维码关注${officialAccountDisplayName.value}，和我们一起记录孩子成长的每一个精彩瞬间。`,
  ].join('\n')
})

function closeFun() {
  openModal.value = false
}

function copyInvitationText() {
  copy(invitationText.value)
  messageService.success('邀请文案已复制')
}

function drawRoundRect(ctx, x, y, width, height, radius, fillStyle, strokeStyle = '') {
  const safeRadius = Math.min(radius, width / 2, height / 2)
  ctx.beginPath()
  ctx.moveTo(x + safeRadius, y)
  ctx.arcTo(x + width, y, x + width, y + height, safeRadius)
  ctx.arcTo(x + width, y + height, x, y + height, safeRadius)
  ctx.arcTo(x, y + height, x, y, safeRadius)
  ctx.arcTo(x, y, x + width, y, safeRadius)
  ctx.closePath()
  if (fillStyle) {
    ctx.fillStyle = fillStyle
    ctx.fill()
  }
  if (strokeStyle) {
    ctx.strokeStyle = strokeStyle
    ctx.stroke()
  }
}

function clipRoundRectPath(ctx, x, y, width, height, radius, corners = ['tl', 'tr', 'br', 'bl']) {
  const safeRadius = Math.min(radius, width / 2, height / 2)
  const topLeft = corners.includes('tl') ? safeRadius : 0
  const topRight = corners.includes('tr') ? safeRadius : 0
  const bottomRight = corners.includes('br') ? safeRadius : 0
  const bottomLeft = corners.includes('bl') ? safeRadius : 0

  ctx.beginPath()
  ctx.moveTo(x + topLeft, y)
  ctx.lineTo(x + width - topRight, y)
  if (topRight)
    ctx.arcTo(x + width, y, x + width, y + topRight, topRight)
  else
    ctx.lineTo(x + width, y)
  ctx.lineTo(x + width, y + height - bottomRight)
  if (bottomRight)
    ctx.arcTo(x + width, y + height, x + width - bottomRight, y + height, bottomRight)
  else
    ctx.lineTo(x + width, y + height)
  ctx.lineTo(x + bottomLeft, y + height)
  if (bottomLeft)
    ctx.arcTo(x, y + height, x, y + height - bottomLeft, bottomLeft)
  else
    ctx.lineTo(x, y + height)
  ctx.lineTo(x, y + topLeft)
  if (topLeft)
    ctx.arcTo(x, y, x + topLeft, y, topLeft)
  else
    ctx.lineTo(x, y)
  ctx.closePath()
}

function drawRoundedImage(ctx, image, x, y, width, height, radius, corners) {
  ctx.save()
  clipRoundRectPath(ctx, x, y, width, height, radius, corners)
  ctx.clip()
  ctx.drawImage(image, x, y, width, height)
  ctx.restore()
}

function loadImage(src) {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.crossOrigin = 'anonymous'
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error('图片加载失败'))
    image.src = src
  })
}

function getWrappedLines(ctx, text, maxWidth, maxLines = Infinity) {
  const lines = []
  const paragraphs = String(text || '').split('\n')

  paragraphs.forEach((paragraph, paragraphIndex) => {
    let current = ''
    for (const char of Array.from(paragraph)) {
      const next = current + char
      if (ctx.measureText(next).width <= maxWidth || !current) {
        current = next
        continue
      }
      lines.push(current)
      current = char
      if (lines.length >= maxLines)
        return
    }
    if (current && lines.length < maxLines)
      lines.push(current)
    if (paragraphIndex < paragraphs.length - 1 && lines.length < maxLines)
      lines.push('')
  })

  return lines.slice(0, maxLines)
}

function drawTextLines(ctx, text, x, y, maxWidth, lineHeight, maxLines = Infinity) {
  const lines = getWrappedLines(ctx, text, maxWidth, maxLines)
  lines.forEach((line, index) => {
    ctx.fillText(line, x, y + index * lineHeight)
  })
  return lines.length
}

function drawEllipsisText(ctx, text, x, y, maxWidth) {
  const raw = String(text || '')
  if (ctx.measureText(raw).width <= maxWidth) {
    ctx.fillText(raw, x, y)
    return
  }

  let result = raw
  while (result.length > 0 && ctx.measureText(`${result}...`).width > maxWidth) {
    result = result.slice(0, -1)
  }
  ctx.fillText(`${result}...`, x, y)
}

function drawStepTitle(ctx, options) {
  const {
    scale,
    x,
    y,
    step,
    beforeText,
    highlightText,
    afterText = '',
  } = options

  const circleSize = 22 * scale
  const circleRadius = circleSize / 2
  const titleY = y + 14 * scale

  ctx.fillStyle = '#1677ff'
  ctx.beginPath()
  ctx.arc(x + circleRadius, y + circleRadius, circleRadius, 0, Math.PI * 2)
  ctx.fill()

  ctx.fillStyle = '#ffffff'
  ctx.font = `italic 700 ${14 * scale}px "Times New Roman", serif`
  ctx.textBaseline = 'middle'
  ctx.textAlign = 'center'
  ctx.fillText(String(step), x + circleRadius, y + circleRadius)

  ctx.textAlign = 'left'
  ctx.textBaseline = 'alphabetic'
  let textX = x + circleSize + 8 * scale
  ctx.fillStyle = '#1f1f1f'
  ctx.font = `700 ${14 * scale}px "PingFang SC", "Microsoft YaHei", sans-serif`
  ctx.fillText(beforeText, textX, titleY)
  textX += ctx.measureText(beforeText).width
  ctx.fillStyle = '#1677ff'
  ctx.fillText(highlightText, textX, titleY)
  textX += ctx.measureText(highlightText).width
  if (afterText) {
    ctx.fillStyle = '#1f1f1f'
    ctx.fillText(afterText, textX, titleY)
  }
}

function canvasToBlob(canvas, type = 'image/jpeg', quality = 0.92) {
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (blob) {
        resolve(blob)
        return
      }
      reject(new Error('海报生成失败'))
    }, type, quality)
  })
}

async function buildInvitationPosterBlob() {
  const [step2Image, qrImage] = await Promise.all([
    loadImage(invitationStep2Image),
    loadImage(qrCodeSrc.value),
  ])

  const scale = 3
  const panelWidth = 343
  const cardPadding = 16
  const cardBorderWidth = 1
  const cardRadius = 18
  const contentWidth = panelWidth - (cardPadding * 2) - (cardBorderWidth * 2)
  const titleBlockHeight = 0

  const headerHeight = 36
  const headerPaddingBottom = 14
  const headerBorder = 1

  const stepTitleHeight = Math.ceil(14 * 1.5)
  const followCardMarginTop = 12
  const followCardHeight = 64
  const dividerBlockHeight = 18 + 1 + 2

  const messageBannerHeight = Math.ceil((12 * 1.45) + 16)
  const messageImageWidth = contentWidth * 0.82
  const messageImageHeight = messageImageWidth * (step2Image.height / step2Image.width)
  const messageImageTop = 8
  const messageCardHeight = messageBannerHeight + messageImageTop + messageImageHeight + 10

  const qrSectionHeight = 88
  const cardHeight = Math.ceil(
    cardPadding
    + headerHeight + headerPaddingBottom + headerBorder
    + 16 + stepTitleHeight + followCardMarginTop + followCardHeight
    + dividerBlockHeight
    + 16 + stepTitleHeight + 10 + messageCardHeight
    + 12 + qrSectionHeight
    + cardPadding,
  )

  const canvas = document.createElement('canvas')
  canvas.width = panelWidth * scale
  canvas.height = (titleBlockHeight + cardHeight) * scale
  const ctx = canvas.getContext('2d')
  if (!ctx)
    throw new Error('海报画布初始化失败')

  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, canvas.width, canvas.height)

  const cardTop = titleBlockHeight * scale
  const outerBorderColor = '#eef2f6'
  const cardStrokeWidth = cardBorderWidth * scale
  const cardStrokeInset = cardStrokeWidth / 2
  ctx.lineWidth = cardStrokeWidth
  drawRoundRect(
    ctx,
    cardStrokeInset,
    cardTop + cardStrokeInset,
    canvas.width - cardStrokeWidth,
    cardHeight * scale - cardStrokeWidth,
    cardRadius * scale,
    '#ffffff',
    outerBorderColor,
  )

  const cardLeft = (cardPadding + cardBorderWidth) * scale
  const cardRight = canvas.width - cardLeft
  let currentY = cardTop + (cardPadding + cardBorderWidth) * scale

  const headerAvatarSize = 36 * scale
  const headerInfoGap = 10 * scale
  const headerHighlightGap = 10 * scale
  const headerHighlightWidth = 56 * scale
  const institutionTextX = cardLeft + headerAvatarSize + headerInfoGap
  const institutionTextMaxWidth = cardRight - institutionTextX - headerHighlightGap - headerHighlightWidth

  ctx.fillStyle = '#1677ff'
  ctx.beginPath()
  ctx.arc(cardLeft + headerAvatarSize / 2, currentY + headerAvatarSize / 2, headerAvatarSize / 2, 0, Math.PI * 2)
  ctx.fill()

  ctx.fillStyle = '#ffffff'
  ctx.font = `600 ${16 * scale}px "PingFang SC", "Microsoft YaHei", sans-serif`
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(institutionInitial.value, cardLeft + headerAvatarSize / 2, currentY + headerAvatarSize / 2)

  ctx.textAlign = 'left'
  ctx.textBaseline = 'top'
  ctx.fillStyle = '#1f1f1f'
  ctx.font = `600 ${14 * scale}px "PingFang SC", "Microsoft YaHei", sans-serif`
  const institutionLineHeight = Math.ceil(14 * 1.35 * scale)
  const institutionLines = getWrappedLines(ctx, institutionName.value, institutionTextMaxWidth, 2)
  const institutionTextTop = currentY + Math.max((headerAvatarSize - (institutionLines.length * institutionLineHeight)) / 2, 0) + 3 * scale
  institutionLines.forEach((line, index) => {
    ctx.fillText(line, institutionTextX, institutionTextTop + index * institutionLineHeight)
  })

  ctx.fillStyle = '#70757d'
  ctx.font = `500 ${14 * scale}px "PingFang SC", "Microsoft YaHei", sans-serif`
  ctx.textBaseline = 'middle'
  ctx.textAlign = 'right'
  ctx.fillText('邀您关注', cardRight, currentY + (headerAvatarSize / 2))

  currentY += headerHeight * scale
  currentY += headerPaddingBottom * scale

  ctx.strokeStyle = outerBorderColor
  ctx.lineWidth = cardBorderWidth * scale
  ctx.beginPath()
  ctx.moveTo(cardLeft, currentY)
  ctx.lineTo(cardRight, currentY)
  ctx.stroke()

  currentY += headerBorder * scale
  currentY += 16 * scale

  drawStepTitle(ctx, {
    scale,
    x: cardLeft,
    y: currentY,
    step: 1,
    beforeText: '第一步： 扫描最下方二维码 ',
    highlightText: '点击关注公众号',
  })

  currentY += stepTitleHeight * scale
  currentY += followCardMarginTop * scale

  const followCardX = cardLeft
  const followCardY = currentY
  drawRoundRect(ctx, followCardX, followCardY, contentWidth * scale, followCardHeight * scale, 18 * scale, '#f8fafc')

  const followAvatarSize = 40 * scale
  const followButtonWidth = 60 * scale
  const followButtonHeight = 26 * scale
  const followButtonRadius = 6 * scale
  const followInnerPaddingX = 16 * scale
  const followButtonX = followCardX + (contentWidth * scale) - followInnerPaddingX - followButtonWidth
  const followButtonY = followCardY + ((followCardHeight * scale - followButtonHeight) / 2)
  const followAvatarX = followCardX + followInnerPaddingX
  const followAvatarY = followCardY + ((followCardHeight * scale - followAvatarSize) / 2)
  const followTextX = followAvatarX + followAvatarSize + (10 * scale)
  const followTextMaxWidth = followButtonX - (10 * scale) - followTextX

  ctx.fillStyle = '#ebedf0'
  ctx.beginPath()
  ctx.arc(followAvatarX + followAvatarSize / 2, followAvatarY + followAvatarSize / 2, followAvatarSize / 2, 0, Math.PI * 2)
  ctx.fill()

  ctx.fillStyle = '#4b4b4b'
  ctx.font = `600 ${14 * scale}px "PingFang SC", "Microsoft YaHei", sans-serif`
  ctx.textBaseline = 'top'
  drawEllipsisText(ctx, officialAccountDisplayName.value, followTextX, followCardY + 14 * scale, followTextMaxWidth)

  ctx.fillStyle = '#99a1b0'
  ctx.font = `400 ${11 * scale}px "PingFang SC", "Microsoft YaHei", sans-serif`
  drawTextLines(ctx, '关注后可接收学校通知', followTextX, followCardY + 33 * scale, followTextMaxWidth, Math.ceil(11 * 1.3 * scale), 2)

  drawRoundRect(ctx, followButtonX, followButtonY, followButtonWidth, followButtonHeight, followButtonRadius, '#12c287')
  ctx.fillStyle = '#ffffff'
  ctx.font = `500 ${13 * scale}px "PingFang SC", "Microsoft YaHei", sans-serif`
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText('关注', followButtonX + followButtonWidth / 2, followButtonY + followButtonHeight / 2)

  currentY += followCardHeight * scale
  currentY += 18 * scale

  ctx.setLineDash([16 * scale, 12 * scale])
  ctx.strokeStyle = '#dfe5ec'
  ctx.lineWidth = 1 * scale
  ctx.beginPath()
  ctx.moveTo(cardLeft, currentY)
  ctx.lineTo(cardRight, currentY)
  ctx.stroke()
  ctx.setLineDash([])

  currentY += (1 + 2) * scale
  currentY += 16 * scale

  drawStepTitle(ctx, {
    scale,
    x: cardLeft,
    y: currentY,
    step: 2,
    beforeText: '第二步： ',
    highlightText: '点击公众号推送消息',
    afterText: ' 立即关注学员',
  })

  currentY += stepTitleHeight * scale
  currentY += 10 * scale

  const messageCardX = cardLeft
  const messageCardY = currentY
  drawRoundRect(ctx, messageCardX, messageCardY, contentWidth * scale, messageCardHeight * scale, 16 * scale, '#f8fafc')
  ctx.save()
  clipRoundRectPath(ctx, messageCardX, messageCardY, contentWidth * scale, messageBannerHeight * scale, 16 * scale, ['tl', 'tr'])
  ctx.fillStyle = '#ffe9e9'
  ctx.fill()
  ctx.restore()

  ctx.fillStyle = '#ff4d4f'
  ctx.font = `500 ${12 * scale}px "PingFang SC", "Microsoft YaHei", sans-serif`
  ctx.textAlign = 'left'
  ctx.textBaseline = 'middle'
  ctx.fillText('仅点击此推送消息才可成功关注学员，否则失效', messageCardX + 12 * scale, messageCardY + (messageBannerHeight * scale / 2))

  const messageImageDrawWidth = messageImageWidth * scale
  const messageImageDrawHeight = messageImageHeight * scale
  const messageImageX = messageCardX + ((contentWidth * scale) - messageImageDrawWidth) / 2
  const messageImageY = messageCardY + messageBannerHeight * scale + messageImageTop * scale
  drawRoundedImage(ctx, step2Image, messageImageX, messageImageY, messageImageDrawWidth, messageImageDrawHeight, 12 * scale)

  currentY += messageCardHeight * scale
  currentY += 12 * scale

  const qrImageSize = 88 * scale
  const qrImageY = currentY
  drawRoundedImage(ctx, qrImage, cardLeft, qrImageY, qrImageSize, qrImageSize, 12 * scale)

  const qrTextX = cardLeft + qrImageSize + 12 * scale
  const qrDescMaxWidth = contentWidth * scale - qrImageSize - 12 * scale
  const qrDescLineHeight = Math.ceil(13 * 1.6 * scale)
  const studentLineHeight = Math.ceil(16 * 1.3 * scale)
  const qrDescText = '长按识别二维码，接收学员在校消息'
  const qrDescLines = getWrappedLines(ctx, qrDescText, qrDescMaxWidth, 2)
  const qrTextBlockHeight = studentLineHeight + (6 * scale) + (qrDescLines.length * qrDescLineHeight)
  const qrTextBlockTop = qrImageY + Math.max((qrImageSize - qrTextBlockHeight) / 2, 0) + 4 * scale
  const studentTextTop = qrTextBlockTop
  ctx.fillStyle = '#1f1f1f'
  ctx.font = `600 ${16 * scale}px "PingFang SC", "Microsoft YaHei", sans-serif`
  ctx.textAlign = 'left'
  ctx.textBaseline = 'top'
  drawEllipsisText(ctx, studentName.value, qrTextX, studentTextTop, qrDescMaxWidth)

  ctx.fillStyle = '#5f6673'
  ctx.font = `400 ${13 * scale}px "PingFang SC", "Microsoft YaHei", sans-serif`
  const qrDescY = studentTextTop + studentLineHeight + (6 * scale)
  qrDescLines.forEach((line, index) => {
    ctx.fillText(line, qrTextX, qrDescY + index * qrDescLineHeight)
  })

  return canvasToBlob(canvas)
}

async function handleSubmit() {
  if (props.loading) {
    messageService.warning('二维码生成中，请稍候')
    return
  }
  if (!hasInvitationInfo.value) {
    messageService.warning('暂无可下载的邀请图片')
    return
  }

  downloading.value = true
  try {
    const blob = await buildInvitationPosterBlob()
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${studentName.value}-二维码邀请.jpg`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    messageService.success('邀请图片下载成功')
  }
  catch (error) {
    console.error('download invitation poster failed', error)
    messageService.error(error?.message || '邀请图片下载失败')
  }
  finally {
    downloading.value = false
  }
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    class="modal-content-box"
    wrap-class-name="qrcode-invitation-modal-wrap"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="800"
    centered
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>二维码邀请</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="contenter">
      <a-spin :spinning="loading">
        <div class="invitation-layout">
          <div class="invitation-panel">
            <div class="panel-title">
              带二维码的邀请图片
            </div>

            <div class="poster-preview">
              <div class="poster-header">
                <div class="institution-badge">
                  {{ institutionInitial }}
                </div>
                <div class="institution-info">
                  <div class="institution-name">
                    {{ institutionName }}
                  </div>
                </div>
                <div class="header-highlight">
                  邀您关注
                </div>
              </div>

              <div class="step-section">
                <div class="step-title">
                  <span class="step-index">1</span>
                  <span>第一步： 扫描最下方二维码 </span>
                  <span class="step-highlight">点击关注公众号</span>
                </div>

                <div class="follow-card">
                  <div class="follow-card__avatar" />
                  <div class="follow-card__meta">
                    <div class="follow-card__name">
                      {{ officialAccountDisplayName }}
                    </div>
                    <div class="follow-card__desc">
                      关注后可接收学校通知
                    </div>
                  </div>
                  <div class="follow-card__action">
                    关注
                  </div>
                </div>
              </div>

              <div class="step-divider" />

              <div class="step-section">
                <div class="step-title">
                  <span class="step-index">2</span>
                  <span>第二步： </span>
                  <span class="step-highlight">点击公众号推送消息</span>
                  <span> 立即关注学员</span>
                </div>

                <div class="message-card">
                  <div class="message-card__banner">
                    仅点击此推送消息才可成功关注学员，否则失效
                  </div>
                  <img class="message-card__image" :src="invitationStep2Image" alt="点击公众号推送消息" />
                </div>
              </div>

              <div class="qrcode-section">
                <img v-if="hasInvitationInfo" class="qrcode-image" :src="qrCodeSrc" alt="邀请二维码">
                <div v-else class="qrcode-placeholder">
                  二维码生成中
                </div>
                <div class="qrcode-content">
                  <div class="qrcode-student">
                    {{ studentName }}
                  </div>
                  <div class="qrcode-desc">
                    长按识别二维码，接收学员在校消息
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="copy-panel" @click="copyInvitationText">
            <div class="panel-title">
              邀请文案
            </div>

            <div class="copy-card">
              <div class="copy-text">
                {{ invitationText }}
              </div>
              <div class="copy-action">
                <a-button @click.stop="copyInvitationText">
                  复制文案
                </a-button>
              </div>
            </div>
          </div>
        </div>
      </a-spin>
    </div>

    <template #footer>
      <a-button @click="closeFun">
        取消
      </a-button>
      <a-button type="primary" :loading="downloading" :disabled="loading || !hasInvitationInfo" @click="handleSubmit">
        下载邀请图片
      </a-button>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.contenter {
  padding: 20px 24px 18px;
  background: #fff;
}

.invitation-layout {
  display: grid;
  grid-template-columns: 343px minmax(0, 1fr);
  gap: 18px;
  align-items: start;
}

.invitation-panel {
  width: 343px;
}

.panel-title {
  margin-bottom: 14px;
  color: #1f1f1f;
  font-size: 18px;
  font-weight: 600;
}

.poster-preview {
  padding: 16px;
  border: 1px solid #eef2f6;
  border-radius: 18px;
  background: #fff;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.poster-header {
  display: flex;
  align-items: center;
  padding-bottom: 14px;
  border-bottom: 1px solid #eef2f6;
}

.institution-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 999px;
  background: #1677ff;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  flex-shrink: 0;
}

.institution-info {
  min-width: 0;
  margin-left: 10px;
  flex: 1;
}

.institution-name {
  color: #1f1f1f;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.35;
  word-break: break-all;
}

.header-highlight {
  margin-left: 10px;
  color: #70757d;
  font-size: 14px;
  font-weight: 500;
  flex-shrink: 0;
}

.step-section {
  margin-top: 16px;
}

.step-title {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  color: #1f1f1f;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
}

.step-index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  margin-right: 8px;
  border-radius: 999px;
  background: #1677ff;
  color: #fff;
  font-size: 14px;
  font-style: italic;
  font-weight: 700;
}

.step-highlight {
  color: #1677ff;
}

.follow-card {
  display: flex;
  align-items: center;
  margin-top: 12px;
  padding: 12px 16px;
  border-radius: 18px;
  background: #f8fafc;
}

.follow-card__avatar {
  width: 40px;
  height: 40px;
  border-radius: 999px;
  background: #ebedf0;
  flex-shrink: 0;
}

.follow-card__meta {
  min-width: 0;
  margin-left: 10px;
  flex: 1;
}

.follow-card__name {
  overflow: hidden;
  color: #4b4b4b;
  font-size: 14px;
  font-weight: 600;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.follow-card__desc {
  color: #99a1b0;
  font-size: 11px;
}

.follow-card__action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 60px;
  height: 26px;
  margin-left: 10px;
  border-radius: 6px;
  background: #12c287;
  color: #fff;
  font-size: 13px;
  flex-shrink: 0;
}

.step-divider {
  height: 1px;
  margin: 18px 0 2px;
  background-image: linear-gradient(to right, #dfe5ec 50%, transparent 0%);
  background-position: top;
  background-size: 16px 1px;
  background-repeat: repeat-x;
}

.message-card {
  overflow: hidden;
  margin-top: 10px;
  padding-bottom: 10px;
  border-radius: 16px;
  background: #f8fafc;
}

.message-card__banner {
  padding: 8px 12px;
  background: #ffe9e9;
  color: #ff4d4f;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.45;
}

.message-card__image {
  display: block;
  width: 82%;
  margin: 8px auto 0;
  border-radius: 12px;
}

.qrcode-section {
  display: flex;
  align-items: center;
  margin-top: 12px;
}

.qrcode-image,
.qrcode-placeholder {
  width: 88px;
  height: 88px;
  border-radius: 12px;
  flex-shrink: 0;
}

.qrcode-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed #d9dfe8;
  background: #f8fafc;
  color: #99a1b0;
  font-size: 12px;
}

.qrcode-content {
  margin-left: 12px;
}

.qrcode-student {
  color: #1f1f1f;
  font-size: 16px;
  font-weight: 600;
}

.qrcode-desc {
  margin-top: 6px;
  color: #5f6673;
  font-size: 13px;
  line-height: 1.6;
}

.copy-panel {
  min-width: 0;
}

.copy-card {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 20px;
  border-radius: 18px;
  background: #f6f8fb;
  cursor: pointer;
  transition: box-shadow 0.2s ease, transform 0.2s ease;

  &:hover {
    box-shadow: 0 10px 24px rgba(31, 35, 41, 0.08);
    transform: translateY(-1px);
  }
}

.copy-text {
  color: #4f5560;
  font-size: 15px;
  line-height: 1.9;
  white-space: pre-line;
}

.copy-action {
  display: flex;
  justify-content: flex-end;
  margin-top: 18px;
}

@media (max-width: 960px) {
  .invitation-layout {
    grid-template-columns: 1fr;
  }

  .invitation-panel {
    width: 100%;
  }

  .copy-card {
    min-height: 260px;
  }
}
</style>

<style>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}

.modal-content-box .ant-modal-footer {
  padding: 12px 24px 20px !important;
  border-top: 1px solid #eef2f6;
}

.qrcode-invitation-modal-wrap .ant-modal-content {
  overflow: hidden;
  border-radius: 20px;
}
</style>
