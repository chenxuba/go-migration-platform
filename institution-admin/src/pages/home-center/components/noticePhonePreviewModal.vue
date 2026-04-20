<script setup lang="ts">
import { CloseOutlined, LeftOutlined } from '@ant-design/icons-vue'
import noticePreviewPhoneBg from '@/assets/images/home-center-notice-preview-bg.png'

const props = withDefaults(defineProps<{
  loading?: boolean
  title?: string
  contentHtml?: string
  coverUrl?: string
  primaryText?: string
  publishText?: string
  description?: string
  navTitle?: string
  emptyText?: string
}>(), {
  loading: false,
  title: '',
  contentHtml: '',
  coverUrl: '',
  primaryText: '通知预览',
  publishText: '',
  description: '家长端接收效果预览，请以实际发布页面为准',
  navTitle: '预览通知公告',
  emptyText: '暂无通知内容',
})

const emit = defineEmits<{
  (e: 'after-close'): void
}>()

const open = defineModel<boolean>({
  default: false,
})

const defaultTemplateCover = 'https://prod-cdn.schoolpal.cn/training/next-erp/h5/static/images/notice/cjbk2025.png'

const resolvedTitle = computed(() => String(props.title || '').trim() || '未填写通知标题')
const resolvedCoverUrl = computed(() => String(props.coverUrl || '').trim() || defaultTemplateCover)
const resolvedContentHtml = computed(() => String(props.contentHtml || '').trim())
const showCover = computed(() => !!resolvedCoverUrl.value && !/<img[\s>]/i.test(resolvedContentHtml.value))
const hasMeta = computed(() => !!String(props.primaryText || '').trim() || !!String(props.publishText || '').trim())

function closePreview() {
  open.value = false
}

function handleAfterClose() {
  emit('after-close')
}
</script>

<template>
  <a-modal
    v-model:open="open"
    centered
    wrap-class-name="notice-phone-preview-modal"
    :footer="false"
    :closable="false"
    :mask-closable="true"
    :keyboard="true"
    :width="430"
    :body-style="{ padding: 0, background: 'transparent' }"
    destroy-on-close
    @cancel="closePreview"
    @after-close="handleAfterClose"
  >
    <div class="noticeTemplatePreview">
      <div class="noticeTemplatePreview__phone" :style="{ backgroundImage: `url(${noticePreviewPhoneBg})` }">
        <div class="noticeTemplatePreview__nav">
          <button type="button" class="noticeTemplatePreview__back" @click="closePreview">
            <LeftOutlined />
          </button>
          <div class="noticeTemplatePreview__navTitle">
            {{ navTitle }}
          </div>
        </div>

        <div class="noticeTemplatePreview__viewport">
          <a-spin :spinning="loading">
            <div class="noticeTemplatePreview__tip">
              <span class="noticeTemplatePreview__tipIcon">i</span>
              <span>此页面为预览页面，不支持操作</span>
            </div>

            <div class="noticeTemplatePreview__content">
              <div class="noticeTemplatePreview__title">
                {{ resolvedTitle }}
              </div>

              <div v-if="hasMeta" class="noticeTemplatePreview__meta">
                <span v-if="primaryText" class="noticeTemplatePreview__metaPrimary">{{ primaryText }}</span>
                <span v-if="publishText" class="noticeTemplatePreview__metaSecondary">发布于 {{ publishText }}</span>
              </div>

              <div class="noticeTemplatePreview__metaSub">
                {{ description }}
              </div>

              <img
                v-if="showCover"
                class="noticeTemplatePreview__cover"
                :src="resolvedCoverUrl"
                alt=""
              >

              <div
                class="noticeTemplatePreview__richtext"
                v-html="resolvedContentHtml || `<p>${emptyText}</p>`"
              />
            </div>
          </a-spin>
        </div>
      </div>

      <button type="button" class="noticeTemplatePreview__close" @click="closePreview">
        <CloseOutlined />
      </button>
    </div>
  </a-modal>
</template>

<style lang="less">
.notice-phone-preview-modal {
  .ant-modal {
    max-width: calc(100vw - 24px);
    margin: 0 auto;
  }

  .ant-modal-content {
    background: transparent;
    box-shadow: none;
    padding: 0;
  }

  .ant-modal-body {
    padding: 0;
  }
}
</style>

<style scoped lang="less">
.noticeTemplatePreview {
  --phone-base-width: 378;
  --phone-base-height: 802;
  --phone-max-height: calc(100vh - 168px);
  --phone-width: min(338px, calc(var(--phone-max-height) * var(--phone-base-width) / var(--phone-base-height)), calc(100vw - 56px));
  --phone-height: calc(var(--phone-width) * var(--phone-base-height) / var(--phone-base-width));
  --phone-gap: max(12px, calc(var(--phone-width) * 18 / var(--phone-base-width)));
  --close-size: max(42px, calc(var(--phone-width) * 50 / var(--phone-base-width)));
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--phone-gap);
  padding: 8px 0;
}

.noticeTemplatePreview__phone {
  position: relative;
  width: var(--phone-width);
  height: var(--phone-height);
  background-repeat: no-repeat;
  background-position: center;
  background-size: 100% 100%;
}

.noticeTemplatePreview__nav {
  position: absolute;
  top: calc(var(--phone-height) * 52 / var(--phone-base-height));
  left: calc(var(--phone-width) * 18 / var(--phone-base-width));
  right: calc(var(--phone-width) * 18 / var(--phone-base-width));
  display: flex;
  align-items: center;
  justify-content: center;
  height: calc(var(--phone-height) * 54 / var(--phone-base-height));
  background: #fff;
}

.noticeTemplatePreview__back {
  position: absolute;
  left: calc(var(--phone-width) * 8 / var(--phone-base-width));
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: calc(var(--phone-width) * 32 / var(--phone-base-width));
  height: calc(var(--phone-width) * 32 / var(--phone-base-width));
  padding: 0;
  border: none;
  background: transparent;
  color: #242b38;
  font-size: clamp(15px, calc(var(--phone-width) * 18 / var(--phone-base-width)), 18px);
  cursor: pointer;
}

.noticeTemplatePreview__navTitle {
  font-size: clamp(14px, calc(var(--phone-width) * 16 / var(--phone-base-width)), 16px);
  font-weight: 600;
  line-height: 1.5;
  color: #1f2329;
}

.noticeTemplatePreview__viewport {
  position: absolute;
  top: calc(var(--phone-height) * 106 / var(--phone-base-height));
  left: calc(var(--phone-width) * 18 / var(--phone-base-width));
  right: calc(var(--phone-width) * 18 / var(--phone-base-width));
  bottom: calc(var(--phone-height) * 78 / var(--phone-base-height));
  overflow-y: auto;
  overflow-x: hidden;
  scrollbar-width: thin;
  scrollbar-color: rgba(31, 35, 41, 0.18) transparent;
}

.noticeTemplatePreview__viewport::-webkit-scrollbar {
  width: 4px;
}

.noticeTemplatePreview__viewport::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgba(31, 35, 41, 0.18);
}

.noticeTemplatePreview__viewport::-webkit-scrollbar-track {
  background: transparent;
}

.noticeTemplatePreview__tip {
  display: flex;
  align-items: center;
  gap: calc(var(--phone-width) * 6 / var(--phone-base-width));
  min-height: calc(var(--phone-height) * 36 / var(--phone-base-height));
  padding: calc(var(--phone-height) * 8 / var(--phone-base-height)) calc(var(--phone-width) * 12 / var(--phone-base-width));
  background: #dbe9ff;
  color: #2267df;
  font-size: clamp(11px, calc(var(--phone-width) * 13 / var(--phone-base-width)), 13px);
  line-height: 1.5;
}

.noticeTemplatePreview__tipIcon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: calc(var(--phone-width) * 16 / var(--phone-base-width));
  height: calc(var(--phone-width) * 16 / var(--phone-base-width));
  border-radius: 50%;
  background: #2267df;
  color: #fff;
  font-size: clamp(10px, calc(var(--phone-width) * 12 / var(--phone-base-width)), 12px);
  font-weight: 700;
  line-height: 1;
}

.noticeTemplatePreview__content {
  padding: calc(var(--phone-height) * 16 / var(--phone-base-height)) calc(var(--phone-width) * 12 / var(--phone-base-width)) calc(var(--phone-height) * 24 / var(--phone-base-height));
  color: #444;
}

.noticeTemplatePreview__title {
  font-size: clamp(16px, calc(var(--phone-width) * 19 / var(--phone-base-width)), 19px);
  font-weight: 700;
  line-height: 1.55;
  color: #222;
}

.noticeTemplatePreview__meta {
  display: flex;
  align-items: center;
  gap: calc(var(--phone-width) * 4 / var(--phone-base-width));
  margin-top: calc(var(--phone-height) * 2 / var(--phone-base-height));
  font-size: clamp(11px, calc(var(--phone-width) * 13 / var(--phone-base-width)), 13px);
  line-height: 1.5;
}

.noticeTemplatePreview__metaPrimary {
  color: #2468f2;
}

.noticeTemplatePreview__metaSecondary {
  color: #7c8593;
}

.noticeTemplatePreview__metaSub {
  margin-top: calc(var(--phone-height) * 2 / var(--phone-base-height));
  font-size: clamp(11px, calc(var(--phone-width) * 13 / var(--phone-base-width)), 13px);
  line-height: 1.5;
  color: #4f5662;
}

.noticeTemplatePreview__cover {
  display: block;
  width: 100%;
  margin-top: calc(var(--phone-height) * 12 / var(--phone-base-height));
}

.noticeTemplatePreview__richtext {
  margin-top: calc(var(--phone-height) * 12 / var(--phone-base-height));
  font-size: clamp(13px, calc(var(--phone-width) * 15 / var(--phone-base-width)), 15px);
  line-height: 1.8;
  color: #444;
  word-break: break-word;
}

.noticeTemplatePreview__richtext:deep(p) {
  margin: 0 0 calc(var(--phone-height) * 8 / var(--phone-base-height));
}

.noticeTemplatePreview__richtext:deep(img) {
  display: block;
  max-width: 100%;
  width: 100%;
  height: auto;
  margin: calc(var(--phone-height) * 12 / var(--phone-base-height)) 0;
}

.noticeTemplatePreview__richtext:deep(h1),
.noticeTemplatePreview__richtext:deep(h2),
.noticeTemplatePreview__richtext:deep(h3),
.noticeTemplatePreview__richtext:deep(h4) {
  margin: calc(var(--phone-height) * 14 / var(--phone-base-height)) 0 calc(var(--phone-height) * 8 / var(--phone-base-height));
  color: #222;
  line-height: 1.5;
}

.noticeTemplatePreview__richtext:deep(ul),
.noticeTemplatePreview__richtext:deep(ol) {
  padding-left: calc(var(--phone-width) * 20 / var(--phone-base-width));
  margin: calc(var(--phone-height) * 8 / var(--phone-base-height)) 0;
}

.noticeTemplatePreview__richtext:deep(li) {
  margin-bottom: calc(var(--phone-height) * 6 / var(--phone-base-height));
}

.noticeTemplatePreview__richtext:deep(a) {
  color: #2468f2;
}

.noticeTemplatePreview__close {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: var(--close-size);
  height: var(--close-size);
  padding: 0;
  border: none;
  border-radius: 50%;
  background: #2468f2;
  color: #fff;
  font-size: clamp(18px, calc(var(--phone-width) * 20 / var(--phone-base-width)), 20px);
  box-shadow: 0 10px 20px rgba(36, 104, 242, 0.28);
  cursor: pointer;
}

@media (max-height: 820px) {
  .noticeTemplatePreview {
    --phone-max-height: calc(100vh - 140px);
  }
}
</style>
