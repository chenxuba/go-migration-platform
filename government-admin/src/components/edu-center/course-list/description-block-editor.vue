<script setup>
import { DeleteOutlined, FileWordOutlined, HolderOutlined, PictureOutlined } from '@ant-design/icons-vue'
import { Upload } from 'ant-design-vue'
import { nextTick, onBeforeUnmount, ref, watch } from 'vue'
import Sortable from 'sortablejs'
import * as qiniu from 'qiniu-js'
import { getQiniuToken } from '@/api/qiniu'
import messageService from '@/utils/messageService'

const props = defineProps({
  modelValue: {
    type: Array,
    default: () => [],
  },
  uploadFolder: {
    type: String,
    default: 'detail',
  },
})

const emit = defineEmits(['update:modelValue'])

const descriptionBlockListRef = ref(null)
let descriptionSortable = null

function syncBlocks(nextBlocks) {
  emit('update:modelValue', nextBlocks)
}

function createDescriptionBlock(type, payload = {}) {
  return {
    id: `${type}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    type,
    ...payload,
  }
}

function initDescriptionSortable() {
  if (!descriptionBlockListRef.value || props.modelValue.length <= 1) return
  destroyDescriptionSortable()
  descriptionSortable = Sortable.create(descriptionBlockListRef.value, {
    animation: 180,
    handle: '.description-block-drag',
    ghostClass: 'description-block-ghost',
    onEnd(evt) {
      const { oldIndex, newIndex } = evt
      if (oldIndex == null || newIndex == null || oldIndex === newIndex) return
      const next = [...props.modelValue]
      const moved = next.splice(oldIndex, 1)[0]
      next.splice(newIndex, 0, moved)
      syncBlocks(next)
    },
  })
}

function destroyDescriptionSortable() {
  if (descriptionSortable) {
    descriptionSortable.destroy()
    descriptionSortable = null
  }
}

watch(() => props.modelValue.length, () => {
  nextTick(() => {
    if (props.modelValue.length <= 1) {
      destroyDescriptionSortable()
    } else {
      initDescriptionSortable()
    }
  })
}, { immediate: true })

onBeforeUnmount(() => {
  destroyDescriptionSortable()
})

function addTextBlock() {
  syncBlocks([
    ...props.modelValue,
    createDescriptionBlock('text', { text: '' }),
  ])
}

function updateTextBlock(index, text) {
  const next = [...props.modelValue]
  next[index] = { ...next[index], text }
  syncBlocks(next)
}

function removeDescriptionBlock(index) {
  const next = [...props.modelValue]
  next.splice(index, 1)
  syncBlocks(next)
}

function getDescriptionImageName(block) {
  if (block.name) return block.name
  const url = `${block.url || ''}`
  return url.split('/').pop() || '图片'
}

function beforeDescriptionImageUpload(file) {
  const isImage = ['image/jpeg', 'image/png', 'image/bmp', 'image/webp'].includes(file.type)
  if (!isImage) {
    messageService.error('只能上传 BMP、JPG、JPEG、PNG、WEBP 格式的图片')
    return Upload.LIST_IGNORE
  }
  const isLt4M = file.size / 1024 / 1024 < 4
  if (!isLt4M) {
    messageService.error('图片大小不能超过 4MB')
    return Upload.LIST_IGNORE
  }
  return true
}

function handleDescriptionImageUpload(options) {
  const { file, onSuccess, onError, onProgress } = options
  const rawFile = file.originFileObj || file

  if (beforeDescriptionImageUpload(rawFile) !== true) {
    onError?.(new Error('文件校验未通过'))
    return
  }

  ;(async () => {
    try {
      const tokenRes = await getQiniuToken()
      const { token, uuid, buckethostname } = tokenRes.result

      const ext = rawFile.name?.includes('.')
        ? rawFile.name.substring(rawFile.name.lastIndexOf('.'))
        : (rawFile.type === 'image/png' ? '.png' : '.jpg')
      const key = `${props.uploadFolder}/${uuid}${ext}`

      const config = {
        useCdnDomain: true,
        region: qiniu.region.z0,
      }
      const putExtra = {
        fname: rawFile.name,
        mimeType: rawFile.type,
      }

      const observable = qiniu.upload(rawFile, key, token, putExtra, config)

      observable.subscribe({
        next(res) {
          onProgress?.({ percent: Math.floor(res.total.percent) })
        },
        error(err) {
          console.error('详情图片上传失败:', err)
          messageService.error(`上传失败: ${err?.message || '未知错误'}`)
          onError?.(err)
        },
        complete(res) {
          const fileUrl = buckethostname + res.key
          syncBlocks([
            ...props.modelValue,
            createDescriptionBlock('image', {
              url: fileUrl,
              name: rawFile.name,
            }),
          ])
          onSuccess?.({ url: fileUrl }, file)
        },
      })
    }
    catch (error) {
      console.error('获取七牛 token 失败:', error)
      messageService.error('获取上传凭证失败')
      onError?.(error)
    }
  })()
}
</script>

<template>
  <div class="description-block-editor" :class="{ 'description-block-editor--empty': !modelValue.length }">
    <div v-if="modelValue.length" ref="descriptionBlockListRef" class="description-block-list">
      <div
        v-for="(block, index) in modelValue"
        :key="block.id"
        class="description-block"
      >
        <div
          class="description-block-card"
          :class="block.type === 'image' ? 'description-block-card--image' : 'description-block-card--text'"
        >
          <template v-if="block.type === 'text'">
            <a-textarea
              :value="block.text"
              :maxlength="500"
              :auto-size="{ minRows: 4, maxRows: 6 }"
              placeholder="请输入，最多 500 字"
              @update:value="updateTextBlock(index, $event)"
            />
          </template>
          <template v-else>
            <div class="description-image-row">
              <img :src="block.url" alt="详情图片">
              <div class="description-image-name">
                {{ getDescriptionImageName(block) }}
              </div>
            </div>
          </template>
        </div>
        <button type="button" class="description-block-drag" title="拖拽排序">
          <HolderOutlined />
        </button>
        <a-button type="text" class="description-block-delete" @click="removeDescriptionBlock(index)">
          <DeleteOutlined />
        </a-button>
      </div>
    </div>

    <div v-if="modelValue.length" class="description-sort-tip">
      鼠标拖拽可以变动以上文字和图片的排列顺序
    </div>

    <div class="description-toolbar">
      <a-upload
        :show-upload-list="false"
        :custom-request="handleDescriptionImageUpload"
        :before-upload="beforeDescriptionImageUpload"
      >
        <a-button type="primary" ghost>
          <template #icon>
            <PictureOutlined />
          </template>
          添加图片
        </a-button>
      </a-upload>
      <a-button type="primary" ghost @click="addTextBlock">
        <template #icon>
          <FileWordOutlined />
        </template>
        添加文字
      </a-button>
    </div>
  </div>
</template>

<style scoped lang="less">
.description-block-editor {
  width: 100%;
}

.description-block-editor--empty {
  padding-top: 6px;
}

.description-block-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
  margin-bottom: 16px;
}

.description-block {
  position: relative;
  padding-right: 56px;
}

.description-block-ghost .description-block-card {
  border-color: #9ec5ff;
  background: #f7fbff;
}

.description-block-card {
  border: 1px solid #e7ebf0;
  border-radius: 14px;
  background: #fff;
}

.description-block-card--text {
  padding-top: 0;
  padding-bottom: 0;
  padding-left: 0;
}

.description-block-card--text :deep(.ant-input-textarea) {
  display: block;
}

.description-block-card--image {
  display: flex;
  align-items: center;
  padding: 14px;
  min-height: 108px;
}

.description-image-row {
  display: flex;
  align-items: center;
  gap: 14px;
  min-width: 0;
}

.description-image-row img {
  width: 72px;
  height: 72px;
  border-radius: 10px;
  object-fit: cover;
  flex: 0 0 auto;
}

.description-image-name {
  min-width: 0;
  color: #2b2f36;
  font-size: 15px;
  line-height: 1.5;
  word-break: break-all;
}

.description-block-delete {
  position: absolute;
  right:6px;
  top: 26px;
  color: #8b95a7;
}

.description-block-drag {
  position: absolute;
  right: 16px;
  bottom: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: 0;
  background: transparent;
  color: #a0a9b8;
  cursor: grab;
}

.description-block-drag:active {
  cursor: grabbing;
}

.description-sort-tip {
  margin-bottom: 14px;
  color: #8b95a7;
  font-size: 13px;
}

.description-toolbar {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
}
</style>
