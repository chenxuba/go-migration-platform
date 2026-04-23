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

function drawStepTitle(ctx, step, x, y, beforeText, highlightText, afterText = '') {
  ctx.fillStyle = '#1677ff'
  ctx.beginPath()
  ctx.arc(x + 24, y + 24, 24, 0, Math.PI * 2)
  ctx.fill()

  ctx.fillStyle = '#ffffff'
  ctx.font = 'italic 700 26px "Times New Roman", serif'
  ctx.textBaseline = 'middle'
  ctx.textAlign = 'center'
  ctx.fillText(String(step), x + 24, y + 24)

  ctx.textAlign = 'left'
  ctx.textBaseline = 'alphabetic'
  let textX = x + 78
  ctx.fillStyle = '#1f1f1f'
  ctx.font = '700 28px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.fillText(beforeText, textX, y + 36)
  textX += ctx.measureText(beforeText).width
  ctx.fillStyle = '#1677ff'
  ctx.fillText(highlightText, textX, y + 36)
  textX += ctx.measureText(highlightText).width
  if (afterText) {
    ctx.fillStyle = '#1f1f1f'
    ctx.fillText(afterText, textX, y + 36)
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

  const canvas = document.createElement('canvas')
  canvas.width = 1125
  canvas.height = 2259
  const ctx = canvas.getContext('2d')
  if (!ctx)
    throw new Error('海报画布初始化失败')

  ctx.fillStyle = '#f5f7fb'
  ctx.fillRect(0, 0, canvas.width, canvas.height)

  drawRoundRect(ctx, 42, 70, 1041, 1830, 32, '#ffffff')

  ctx.strokeStyle = '#edf1f5'
  ctx.lineWidth = 2
  ctx.beginPath()
  ctx.moveTo(42, 360)
  ctx.lineTo(1083, 360)
  ctx.stroke()

  ctx.fillStyle = '#1677ff'
  ctx.beginPath()
  ctx.arc(145, 200, 62, 0, Math.PI * 2)
  ctx.fill()

  ctx.fillStyle = '#ffffff'
  ctx.font = '500 60px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(institutionInitial.value, 145, 200)

  ctx.textAlign = 'left'
  ctx.textBaseline = 'top'
  ctx.fillStyle = '#1f1f1f'
  ctx.font = '600 50px "PingFang SC", "Microsoft YaHei", sans-serif'
  drawTextLines(ctx, institutionName.value, 255, 130, 560, 68, 2)

  ctx.fillStyle = '#70757d'
  ctx.font = '500 38px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.fillText('邀请您关注', 820, 150)

  drawStepTitle(ctx, 1, 95, 420, '第一步： 扫描最下方二维码 ', '点击关注公众号')

  drawRoundRect(ctx, 96, 480, 830, 250, 22, '#f8fafc')
  ctx.fillStyle = '#ebedf0'
  ctx.beginPath()
  ctx.arc(175, 570, 40, 0, Math.PI * 2)
  ctx.fill()

  ctx.fillStyle = '#4b4b4b'
  ctx.font = '600 28px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.fillText(officialAccountDisplayName.value, 240, 542)

  ctx.fillStyle = '#99a1b0'
  ctx.font = '400 22px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.fillText('关注后可接收学校通知', 240, 590)

  drawRoundRect(ctx, 720, 520, 190, 92, 16, '#12c287')
  ctx.fillStyle = '#ffffff'
  ctx.font = '700 44px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText('关注', 815, 566)

  ctx.setLineDash([14, 12])
  ctx.strokeStyle = '#e2e7ef'
  ctx.lineWidth = 2
  ctx.beginPath()
  ctx.moveTo(95, 780)
  ctx.lineTo(925, 780)
  ctx.stroke()
  ctx.setLineDash([])

  drawStepTitle(ctx, 2, 95, 840, '第二步： ', '点击公众号推送消息', ' 立即关注学员')

  drawRoundRect(ctx, 96, 925, 830, 705, 22, '#f8fafc')
  drawRoundRect(ctx, 96, 925, 830, 96, 22, '#ffe9e9')
  ctx.fillStyle = '#ff4d4f'
  ctx.font = '600 24px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textAlign = 'left'
  ctx.textBaseline = 'middle'
  ctx.fillText('仅点击此推送消息才可成功关注学员，否则失效', 128, 973)

  ctx.drawImage(step2Image, 96, 1021, 830, 571)

  ctx.drawImage(qrImage, 96, 1668, 180, 180)
  ctx.fillStyle = '#1f1f1f'
  ctx.font = '600 34px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textBaseline = 'top'
  ctx.fillText(studentName.value, 330, 1704)

  ctx.fillStyle = '#5f6673'
  ctx.font = '400 30px "PingFang SC", "Microsoft YaHei", sans-serif'
  drawTextLines(ctx, '长按识别二维码，接收学员在校消息', 330, 1786, 480, 46, 2)

  ctx.fillStyle = '#d0d5df'
  ctx.font = '600 62px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textAlign = 'center'
  ctx.fillText('校宝 APP', canvas.width / 2, 2090)

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
    style="top: 12px"
    class="modal-content-box"
    wrap-class-name="qrcode-invitation-modal-wrap"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="860"
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
  width: 46px;
  height: 46px;
  border-radius: 999px;
  background: #1677ff;
  color: #fff;
  font-size: 24px;
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
  font-size: 15px;
  font-weight: 600;
  line-height: 1.35;
  word-break: break-all;
}

.header-highlight {
  margin-left: 10px;
  color: #70757d;
  font-size: 15px;
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
  width: 24px;
  height: 24px;
  margin-right: 8px;
  border-radius: 999px;
  background: #1677ff;
  color: #fff;
  font-size: 15px;
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
  padding: 16px;
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
  margin-top: 4px;
  color: #99a1b0;
  font-size: 11px;
}

.follow-card__action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 82px;
  height: 38px;
  margin-left: 10px;
  border-radius: 10px;
  background: #12c287;
  color: #fff;
  font-size: 15px;
  font-weight: 700;
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
  margin-top: 12px;
  border-radius: 18px;
  background: #f8fafc;
}

.message-card__banner {
  padding: 10px 14px;
  background: #ffe9e9;
  color: #ff4d4f;
  font-size: 13px;
  font-weight: 500;
}

.message-card__image {
  display: block;
  width: 100%;
}

.qrcode-section {
  display: flex;
  align-items: center;
  margin-top: 16px;
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
  min-height: 480px;
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
