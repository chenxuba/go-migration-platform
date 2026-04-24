<script setup lang="ts">
import { DeleteOutlined, FileImageOutlined, PlayCircleOutlined, VideoCameraOutlined } from '@ant-design/icons-vue'
import * as qiniu from 'qiniu-js'
import { Modal } from 'ant-design-vue'
import { computed, h, ref, watch } from 'vue'
import type { RehabRecordMediaItem } from '@/api/edu-center/class-record'
import { getQiniuToken, getVideoUploadToken } from '@/api/qiniu'
import messageService from '@/utils/messageService'
import { validateUploadFileByToken } from '@/utils/upload-limit'

const props = withDefaults(defineProps<{
  modelValue?: RehabRecordMediaItem[]
  disabled?: boolean
  hintText?: string
  imageLimit?: number
  videoLimit?: number
}>(), {
  modelValue: () => [],
  disabled: false,
  hintText: '导出不包含图片/视频',
  imageLimit: 6,
  videoLimit: 2,
})

const emit = defineEmits<{
  (e: 'update:modelValue', value: RehabRecordMediaItem[]): void
}>()

const mediaItems = ref<RehabRecordMediaItem[]>([])
const imageUploading = ref(false)
const videoUploading = ref(false)
const previewImageOpen = ref(false)
const previewImageSrc = ref('')
const previewImageTitle = ref('')
const previewVideoOpen = ref(false)
const previewVideoItem = ref<RehabRecordMediaItem | null>(null)
const previewVideoRef = ref<HTMLVideoElement | null>(null)
const imageInputRef = ref<HTMLInputElement | null>(null)
const videoInputRef = ref<HTMLInputElement | null>(null)

const IMAGE_MAX_SIZE_MB = 10
const VIDEO_MAX_SIZE_MB = 100

interface InvalidSelectionItem {
  name: string
  reason: string
}

function normalizeTextValue(value?: string | number | null) {
  return String(value ?? '').trim()
}

function inferMediaType(url: string) {
  if (/\.(mp4|mov|webm|ogg|m4v)(\?.*)?$/i.test(url))
    return 'video'
  return 'image'
}

function normalizeMediaItem(item?: Partial<RehabRecordMediaItem> | string | null) {
  if (!item)
    return null

  if (typeof item === 'string') {
    const url = normalizeTextValue(item)
    if (!url)
      return null
    return {
      mediaType: inferMediaType(url),
      url,
      fileName: '',
    } satisfies RehabRecordMediaItem
  }

  const url = normalizeTextValue(item.url)
  if (!url)
    return null

  const rawMediaType = normalizeTextValue(item.mediaType).toLowerCase()
  const mediaType = rawMediaType === 'image' || rawMediaType === 'video'
    ? rawMediaType
    : inferMediaType(url)
  const size = Number(item.size || 0)

  return {
    mediaType,
    url,
    fileName: normalizeTextValue(item.fileName),
    size: Number.isFinite(size) && size > 0 ? size : undefined,
  } satisfies RehabRecordMediaItem
}

function normalizeMediaList(items?: RehabRecordMediaItem[] | null) {
  return Array.isArray(items)
    ? items.map(item => normalizeMediaItem(item)).filter(Boolean) as RehabRecordMediaItem[]
    : []
}

function buildMediaSignature(items?: RehabRecordMediaItem[] | null) {
  return JSON.stringify((items || []).map(item => ({
    mediaType: normalizeTextValue(item.mediaType).toLowerCase(),
    url: normalizeTextValue(item.url),
    fileName: normalizeTextValue(item.fileName),
    size: Number(item.size || 0) || 0,
  })))
}

watch(() => props.modelValue, (value) => {
  const normalized = normalizeMediaList(value)
  if (buildMediaSignature(normalized) === buildMediaSignature(mediaItems.value))
    return
  mediaItems.value = normalized
}, { deep: true, immediate: true })

function emitValue() {
  emit('update:modelValue', mediaItems.value.map(item => ({ ...item })))
}

const imageCount = computed(() => mediaItems.value.filter(item => item.mediaType === 'image').length)
const videoCount = computed(() => mediaItems.value.filter(item => item.mediaType === 'video').length)

function openImagePicker() {
  if (props.disabled || imageUploading.value || imageCount.value >= props.imageLimit)
    return
  imageInputRef.value?.click()
}

function openVideoPicker() {
  if (props.disabled || videoUploading.value || videoCount.value >= props.videoLimit)
    return
  videoInputRef.value?.click()
}

function buildSelectionWarningContent(invalidItems: InvalidSelectionItem[], prefixText: string) {
  return h('div', { class: 'media-selection-dialog' }, [
    h('div', { class: 'media-selection-dialog__intro' }, prefixText),
    ...invalidItems.map(item => h('div', { class: 'media-selection-dialog__item' }, `${item.name}：${item.reason}`)),
  ])
}

function validateImageFiles(files: File[]) {
  const invalidItems: InvalidSelectionItem[] = []
  const acceptedFiles: File[] = []

  files.forEach((file) => {
    const isImage = file.type.startsWith('image/')
    if (!isImage) {
      invalidItems.push({ name: file.name, reason: '文件格式不支持' })
      return
    }


    acceptedFiles.push(file)
  })

  const remainingCount = Math.max(props.imageLimit - imageCount.value, 0)
  const validFiles = acceptedFiles.slice(0, remainingCount)
  if (acceptedFiles.length > remainingCount) {
    acceptedFiles.slice(remainingCount).forEach((file) => {
      invalidItems.push({ name: file.name, reason: `最多还能上传 ${remainingCount} 张图片` })
    })
  }
  return { validFiles, invalidItems }
}

function validateVideoFiles(files: File[]) {
  const invalidItems: InvalidSelectionItem[] = []
  const acceptedFiles: File[] = []

  files.forEach((file) => {
    const isVideo = file.type.startsWith('video/')
    if (!isVideo) {
      invalidItems.push({ name: file.name, reason: '文件格式不支持' })
      return
    }


    acceptedFiles.push(file)
  })

  const remainingCount = Math.max(props.videoLimit - videoCount.value, 0)
  const validFiles = acceptedFiles.slice(0, remainingCount)
  if (acceptedFiles.length > remainingCount) {
    acceptedFiles.slice(remainingCount).forEach((file) => {
      invalidItems.push({ name: file.name, reason: `最多还能上传 ${remainingCount} 个视频` })
    })
  }
  return { validFiles, invalidItems }
}

function resetFileInput(input?: HTMLInputElement | null) {
  if (input)
    input.value = ''
}

function uploadSingleImage(file: File) {
  return new Promise<RehabRecordMediaItem>(async (resolve, reject) => {
    try {
      const tokenRes: any = await getQiniuToken()
      const { token, uuid, buckethostname } = tokenRes.result || {}
      if (!token || !uuid || !buckethostname)
        throw new Error('上传凭证缺失')

      const ext = file.name?.includes('.')
        ? file.name.substring(file.name.lastIndexOf('.'))
        : (file.type === 'image/png' ? '.png' : '.jpg')
      const key = `rehab-record/${uuid}${ext}`

      const observable = qiniu.upload(file, key, token, {
        fname: file.name,
        mimeType: file.type,
      }, {
        useCdnDomain: true,
        region: qiniu.region.z0,
      })

      observable.subscribe({
        error(err) {
          reject(err)
        },
        complete(res) {
          resolve({
            mediaType: 'image',
            url: `${buckethostname}${res.key}`,
            fileName: normalizeTextValue(file.name),
            size: Number(file.size || 0) || undefined,
          })
        },
      })
    }
    catch (error) {
      reject(error)
    }
  })
}

function uploadSingleVideo(file: File) {
  return new Promise<RehabRecordMediaItem>(async (resolve, reject) => {
    try {
      const tokenRes: any = await getVideoUploadToken()
      const { token, uuid, buckethostname } = tokenRes.result || {}
      const key = normalizeTextValue(uuid)
      if (!token || !key || !buckethostname)
        throw new Error('视频上传凭证缺失')
      validateUploadFileByToken(file, tokenRes.result, '视频')

      const observable = qiniu.upload(file, key, token, {
        fname: file.name,
        mimeType: file.type,
      }, {
        useCdnDomain: true,
        region: qiniu.region.z0,
      })

      observable.subscribe({
        error(err) {
          reject(err)
        },
        complete(res) {
          resolve({
            mediaType: 'video',
            url: `${buckethostname}${res.key}`,
            fileName: normalizeTextValue(file.name) || '视频附件',
            size: Number(file.size || 0) || undefined,
          })
        },
      })
    }
    catch (error) {
      reject(error)
    }
  })
}

async function uploadImageFiles(files: File[]) {
  if (files.length === 0)
    return

  imageUploading.value = true
  try {
    const results = await Promise.allSettled(files.map(file => uploadSingleImage(file)))
    const successItems = results
      .filter(result => result.status === 'fulfilled')
      .map(result => result.value)
    const failedCount = results.length - successItems.length

    if (successItems.length > 0) {
      mediaItems.value = [...mediaItems.value, ...successItems]
      emitValue()
      messageService.success(`已上传 ${successItems.length} 张图片`)
    }

    if (failedCount > 0) {
      messageService.error(`${failedCount} 张图片上传失败，请稍后重试`)
    }
  }
  finally {
    imageUploading.value = false
  }
}

async function uploadVideoFiles(files: File[]) {
  if (files.length === 0)
    return

  videoUploading.value = true
  try {
    const results = await Promise.allSettled(files.map(file => uploadSingleVideo(file)))
    const successItems = results
      .filter(result => result.status === 'fulfilled')
      .map(result => result.value)
    const failedCount = results.length - successItems.length

    if (successItems.length > 0) {
      mediaItems.value = [...mediaItems.value, ...successItems]
      emitValue()
      messageService.success(`已上传 ${successItems.length} 个视频`)
    }

    if (failedCount > 0) {
      messageService.error(`${failedCount} 个视频上传失败，请稍后重试`)
    }
  }
  finally {
    videoUploading.value = false
  }
}

function handleImageFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  resetFileInput(input)
  if (files.length === 0)
    return

  const { validFiles, invalidItems } = validateImageFiles(files)
  if (invalidItems.length === 0) {
    uploadImageFiles(validFiles)
    return
  }

  if (validFiles.length === 0) {
    Modal.warning({
      title: '图片无法上传',
      content: buildSelectionWarningContent(invalidItems, '以下图片不支持上传：'),
      okText: '知道了',
    })
    return
  }

  Modal.confirm({
    title: '部分图片无法上传',
    content: buildSelectionWarningContent(invalidItems, '以下图片不支持上传：'),
    okText: `继续上传剩余 ${validFiles.length} 张`,
    cancelText: '返回修改',
    onOk: async () => {
      await uploadImageFiles(validFiles)
    },
  })
}

function handleVideoFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  const files = Array.from(input.files || [])
  resetFileInput(input)
  if (files.length === 0)
    return

  const { validFiles, invalidItems } = validateVideoFiles(files)
  if (invalidItems.length === 0) {
    uploadVideoFiles(validFiles)
    return
  }

  if (validFiles.length === 0) {
    Modal.warning({
      title: '视频无法上传',
      content: buildSelectionWarningContent(invalidItems, '以下视频不支持上传：'),
      okText: '知道了',
    })
    return
  }

  Modal.confirm({
    title: '部分视频无法上传',
    content: buildSelectionWarningContent(invalidItems, '以下视频不支持上传：'),
    okText: `继续上传剩余 ${validFiles.length} 个`,
    cancelText: '返回修改',
    onOk: async () => {
      await uploadVideoFiles(validFiles)
    },
  })
}

function resolveMediaLabel(item: RehabRecordMediaItem, index: number) {
  const fileName = normalizeTextValue(item.fileName)
  if (fileName)
    return fileName
  return item.mediaType === 'video' ? `视频${index + 1}` : `图片${index + 1}`
}

function handleMediaPreview(item: RehabRecordMediaItem, index = 0) {
  if (item.mediaType === 'video') {
    previewVideoItem.value = item
    previewVideoOpen.value = true
    return
  }
  previewImageSrc.value = item.url || ''
  previewImageTitle.value = resolveMediaLabel(item, index)
  previewImageOpen.value = true
}

function handleRemove(item: RehabRecordMediaItem) {
  mediaItems.value = mediaItems.value.filter(target => target.url !== item.url)
  emitValue()
}

function closeVideoPreview() {
  const video = previewVideoRef.value
  if (video) {
    video.pause()
    video.currentTime = 0
    video.removeAttribute('src')
    video.load()
  }
  previewVideoOpen.value = false
  previewVideoItem.value = null
}
</script>

<template>
  <div class="rehab-media-upload">
    <div v-if="!disabled" class="media-action-row">
      <input
        ref="imageInputRef"
        class="media-hidden-input"
        type="file"
        accept=".jpg,.jpeg,.png,.bmp,.webp,.gif"
        multiple
        @change="handleImageFileChange"
      >
      <input
        ref="videoInputRef"
        class="media-hidden-input"
        type="file"
        accept=".mp4,.mov,.webm,.ogg,.m4v"
        multiple
        @change="handleVideoFileChange"
      >

      <a-button :loading="imageUploading" :disabled="imageCount >= imageLimit" @click="openImagePicker">
        <template #icon>
          <FileImageOutlined />
        </template>
        上传图片
      </a-button>

      <a-button :loading="videoUploading" :disabled="videoCount >= videoLimit" @click="openVideoPicker">
        <template #icon>
          <VideoCameraOutlined />
        </template>
        上传视频
      </a-button>

      <span v-if="hintText" class="media-hint">{{ hintText }}</span>
    </div>

    <div v-if="mediaItems.length" class="media-thumb-list">
      <div
        v-for="(item, index) in mediaItems"
        :key="`${item.mediaType}-${item.url}`"
        class="media-thumb-card"
        @click="handleMediaPreview(item, index)"
      >
        <img
          v-if="item.mediaType === 'image'"
          class="media-thumb-cover"
          :src="item.url"
          :alt="resolveMediaLabel(item, index)"
        >
        <div v-else class="media-thumb-video">
          <video
            class="media-thumb-cover"
            :src="item.url"
            muted
            playsinline
            preload="metadata"
          />
          <div class="media-thumb-play">
            <PlayCircleOutlined />
          </div>
        </div>

        <div class="media-thumb-label">
          {{ resolveMediaLabel(item, index) }}
        </div>

        <a-button
          v-if="!disabled"
          type="text"
          size="small"
          class="media-thumb-remove"
          @click.stop="handleRemove(item)"
        >
          <template #icon>
            <DeleteOutlined />
          </template>
        </a-button>
      </div>
    </div>

    <a-modal
      :open="previewImageOpen"
      :title="previewImageTitle"
      :footer="null"
      @cancel="previewImageOpen = false"
    >
      <img
        alt="图片预览"
        style="width: 100%;"
        :src="previewImageSrc"
      >
    </a-modal>

    <a-modal
      :open="previewVideoOpen"
      :title="previewVideoItem?.fileName || '视频预览'"
      :footer="null"
      width="720px"
      :destroy-on-close="true"
      @cancel="closeVideoPreview"
      @after-close="closeVideoPreview"
    >
      <video
        v-if="previewVideoItem?.url"
        ref="previewVideoRef"
        class="rehab-preview-video"
        :src="previewVideoItem.url"
        controls
        playsinline
      />
    </a-modal>
  </div>
</template>

<style lang="less" scoped>
.rehab-media-upload {
  width: 100%;
}

.media-action-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
}

.media-hidden-input {
  display: none;
}

.media-hint {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 20px;
}

.media-thumb-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 12px;
}

.media-thumb-card {
  position: relative;
  width: 124px;
  height: 72px;
  flex: 0 0 auto;
  overflow: hidden;
  border-radius: 10px;
  border: 1px solid #edf0f5;
  background: #f5f7fb;
  cursor: pointer;
}

.media-thumb-cover {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.media-thumb-video {
  position: relative;
  width: 100%;
  height: 100%;
  background: #0f172a;
}

.media-thumb-play {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 22px;
  background: rgba(15, 23, 42, 0.2);
}

.media-thumb-label {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  padding: 0 8px;
  height: 24px;
  color: #fff;
  font-size: 12px;
  line-height: 24px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  background: linear-gradient(180deg, rgba(15, 23, 42, 0) 0%, rgba(15, 23, 42, 0.72) 100%);
}

.media-thumb-remove {
  position: absolute;
  top: 6px;
  right: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  padding: 0;
  color: #fff;
  border-radius: 999px;
  background: rgba(15, 23, 42, 0.52);
}

.media-thumb-remove:hover {
  color: #fff;
  background: rgba(15, 23, 42, 0.72);
}

.rehab-preview-video {
  width: 100%;
  max-height: 70vh;
  border-radius: 12px;
  background: #000;
}

:deep(.media-selection-dialog__intro) {
  margin-bottom: 8px;
  color: #595959;
  line-height: 22px;
}

:deep(.media-selection-dialog__item) {
  color: #262626;
  line-height: 22px;
  word-break: break-all;
}
</style>
