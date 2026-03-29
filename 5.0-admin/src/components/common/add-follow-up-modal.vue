<script setup>
import { CloseOutlined, PlusOutlined, QuestionCircleOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import dayjs from 'dayjs'
import * as qiniu from 'qiniu-js'
import { FollowMethod, FollowMethodLabel, FollowUpStatusLabel, IntentionLevelLabel } from '@/enums'
import emitter, { EVENTS } from '@/utils/eventBus'
import { getProcessContentPageApi } from '@/api/edu-center/registr-renewal'
import { getQiniuToken } from '@/api/qiniu'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  record: {
    type: Object,
    default: () => ({}),
  },
  editRecord: {
    type: Object,
    default: () => ({}),
  },
  isEdit: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open', 'handleFollowUpSubmit'])
const formRef = ref(null)
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
// 在组件中补充
const MAX_COUNT = 3

// 注：已移除硬编码的 typeMap，改用动态标签生成

// 计算属性
const maxCountPlaceholder = computed(() => {
  if (courseLoading.value) {
    return '正在加载课程列表...'
  }
  if (searchLoading.value) {
    return '搜索中...'
  }
  return formState.intentCourseIds.length >= MAX_COUNT
    ? '已达最大选择数量'
    : '请搜索或选择课程（最多3个）'
})

// 选择事件处理
function handleCourseChange(selected) {
  if (selected.length > MAX_COUNT) {
    message.warn(`最多选择 ${MAX_COUNT} 个课程`)
    formState.intentCourseIds = selected.slice(0, MAX_COUNT)
  }

  // 选择后重新显示全部课程
  nextTick(() => {
    searchKey.value = ''
    filterCourseList('')
  })
}
const courseList = ref([]) // 显示的课程列表
const allCourseList = ref([]) // 完整的课程数据缓存
const courseLoading = ref(false)
// 搜索相关状态
const searchLoading = ref(false)
const searchKey = ref('')

// 搜索防抖定时器
let searchTimer = null

function getBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result)
    reader.onerror = error => reject(error)
  })
}

/** 学员详情里 intendedCourse 可能是 "1,2" 字符串或 [1,2] 数组 */
function intendedCourseToSelectValues(raw) {
  if (raw == null || raw === '')
    return []
  if (Array.isArray(raw))
    return raw.map(v => String(v)).filter(Boolean)
  if (typeof raw === 'string')
    return raw.split(',').map(s => s.trim()).filter(Boolean)
  return [String(raw)]
}

/** Upload 的 picture-card 内部会对 url 做 extname(url).split，必须是 string */
function ensureUploadUrl(raw) {
  if (raw == null || raw === '')
    return ''
  if (typeof raw === 'string')
    return raw
  if (typeof raw === 'object' && raw.url != null)
    return ensureUploadUrl(raw.url)
  return String(raw)
}

function normalizeUploadFileItem(file) {
  const url = ensureUploadUrl(
    file.url ?? file.response?.url ?? file.thumbUrl ?? file.preview,
  )
  const next = { ...file, url }
  if (file.thumbUrl != null) {
    const t = ensureUploadUrl(file.thumbUrl)
    if (t)
      next.thumbUrl = t
    else
      delete next.thumbUrl
  }
  return next
}
const previewVisible = ref(false)
const previewImage = ref('')
const previewTitle = ref('')
const fileList = ref([])
function handleCancel() {
  previewVisible.value = false
  previewTitle.value = ''
}
async function handlePreview(file) {
  if (!file.url && !file.preview) {
    file.preview = await getBase64(file.originFileObj)
  }
  previewImage.value = file.url || file.preview
  previewVisible.value = true
  const urlStr = typeof file.url === 'string' ? file.url : ''
  previewTitle.value = file.name || (urlStr ? urlStr.substring(urlStr.lastIndexOf('/') + 1) : 'preview')
}
const formState = reactive({
  followMethod: Number(FollowMethod.None),
  intentLevel: props.record.intentLevel,
  nextFollowUpTime: '',
  followUpStatus: props.record.followUpStatus, 
  content: '',
  followImages: '',
  intentCourseIds: [],
})

// 监听editRecord和isEdit的变化，用于数据回显
watch(() => [props.editRecord, props.isEdit], ([newEditRecord, newIsEdit]) => {
  if (newIsEdit && newEditRecord) {
    // 编辑模式，回显数据
    formState.followMethod = newEditRecord.followMethod || Number(FollowMethod.None)
    formState.intentLevel = newEditRecord.intentionLevel || props.record.intentLevel
    formState.nextFollowUpTime = newEditRecord.nextFollowUpTime || ''
    formState.followUpStatus = newEditRecord.followUpStatus || props.record.followUpStatus
    formState.content = newEditRecord.content || ''
    formState.followImages = newEditRecord.followImages || ''
    
    // 优先编辑记录上的意向课程（intentCourseIds 或后端返回的 intendedCourse），再回退学员详情 lessons
    if (newEditRecord.intentCourseIds?.length) {
      formState.intentCourseIds = newEditRecord.intentCourseIds.map(v => String(v))
    }
    else {
      const fromIntended = intendedCourseToSelectValues(newEditRecord.intendedCourse)
      if (fromIntended.length) {
        formState.intentCourseIds = fromIntended
      }
      else if (props.record.lessons && Array.isArray(props.record.lessons)) {
        formState.intentCourseIds = props.record.lessons.map(course => course.id.toString())
      }
      else {
        formState.intentCourseIds = []
      }
    }

    formState.studentId = newEditRecord.studentId || props.record.id

    // 处理编辑模式下的图片显示
    if (newEditRecord.followImages) {
      try {
        const images = typeof newEditRecord.followImages === 'string'
          ? JSON.parse(newEditRecord.followImages)
          : newEditRecord.followImages
        if (Array.isArray(images)) {
          fileList.value = images.map((img, index) => ({
            uid: `-${index + 1}`,
            name: `image-${index + 1}.png`,
            status: 'done',
            url: ensureUploadUrl(typeof img === 'string' ? img : img?.url),
          }))
        }
      }
      catch (error) {
        console.warn('解析编辑记录中的图片失败:', error)
        fileList.value = []
      }
    }
    else {
      fileList.value = []
    }

    // 同步更新 followImages 字段
    nextTick(() => {
      updateFollowImages()
    })
  }
  else {
    // 新增模式，重置为默认值
    formState.followMethod = Number(FollowMethod.None)
    formState.intentLevel = props.record.intentLevel
    formState.nextFollowUpTime = ''
    formState.followUpStatus = props.record.followUpStatus
    formState.content = ''
    formState.followImages = ''
    formState.intentCourseIds = intendedCourseToSelectValues(props.record.intendedCourse)

    formState.studentId = props.record.id

    // 新增模式下的默认图片
    fileList.value = []

    // 同步更新 followImages 字段
    nextTick(() => {
      updateFollowImages()
    })
  }
}, { immediate: true, deep: true })

watch(() => props.record, (newVal) => {
  if (!props.isEdit) {
    formState.studentId = newVal.id
    formState.intentLevel = newVal.intentLevel
    formState.followUpStatus = newVal.followUpStatus
    
    // 回显意向课程：优先 lessons，否则兼容 intendedCourse 为数组或逗号字符串
    if (newVal.lessons && Array.isArray(newVal.lessons)) {
      formState.intentCourseIds = newVal.lessons.map(course => course.id.toString())
    }
    else {
      formState.intentCourseIds = intendedCourseToSelectValues(newVal.intendedCourse)
    }
  }
}, { immediate: true })

const btnLoading = ref(false)

// 课程标签生成函数（参考外部组件）
function getCourseTagList(item) {
  const tags = []

  if (item.teachMethod === 1) {
    tags.push({
      text: '班级授课',
      color: '#e6f0ff',
      textColor: '#0066ff',
      type: 'normal',
    })
  }

  if (item.teachMethod === 2) {
    tags.push({
      text: '1v1授课',
      color: '#e6f0ff',
      textColor: '#0066ff',
      type: 'normal',
    })
  }

  // 收费方式标签
  if (item.chargeMethods) {
    tags.push({
      text: item.chargeMethods,
      color: '#e6f0ff',
      textColor: '#0066ff',
      type: 'normal',
    })
  }

  // 是否有体验价
  if (item.hasExperiencePrice) {
    tags.push({
      text: '体验价',
      color: '#fff5e6',
      textColor: '#ff9900',
      type: 'normal',
    })
  }

  return tags
}

// 通用标签样式
function getTagStyle(type = 'normal') {
  const baseStyle = {
    borderRadius: '20px',
    marginRight: '0',
    height: '20px',
  }

  if (type === 'primary') {
    return {
      ...baseStyle,
      color: '#fff',
    }
  }

  return {
    ...baseStyle,
    color: '#0066ff',
  }
}

// 获取课程列表（只在初始化时调用接口）
async function getCourseList() {
  if (courseLoading.value) return

  try {
    courseLoading.value = true

    const params = {
      pageRequestModel: {
        needTotal: true,
        pageSize: 1000, // 获取所有课程数据
        pageIndex: 1,
        skipCount: 0
      },
      queryModel: {
        delFlag: false,
        productType: 1, // 1-课程商品
        saleStatus: true, // 只获取可销售的商品
      },
      sortModel: {}
    }

    const res = await getProcessContentPageApi(params)

    if (res.code === 200) {
      const { result = [], total = 0 } = res || {}

      // 处理数据格式，保留完整的原始数据
      const courses = result.map(course => ({
        label: course.name || '未命名课程',
        value: course.id.toString(),
        ...course // 保留所有原始数据，用于标签渲染
      }))

      // 保存完整数据到缓存
      allCourseList.value = courses
      courseList.value = courses

      console.log('获取课程列表成功:', courses.length, '条数据')
    }
  } catch (error) {
    console.error('获取课程列表失败:', error)
    message.error('获取课程列表失败')
  } finally {
    courseLoading.value = false
  }
}

// 本地过滤课程数据
function filterCourseList(keyword = '') {
  if (!keyword) {
    // 无搜索关键词，显示所有课程
    courseList.value = allCourseList.value
  } else {
    // 根据关键词过滤课程（不区分大小写）
    courseList.value = allCourseList.value.filter(course =>
      course.label.toLowerCase().includes(keyword.toLowerCase())
    )
  }
}

// 加载更多数据（本地搜索模式下不需要）
async function loadMoreCourses() {
  // 本地搜索模式下不需要加载更多
  return
}

// 本地搜索课程
function searchCourses(keyword) {
  // 清除之前的定时器
  if (searchTimer) {
    clearTimeout(searchTimer)
  }

  // 设置搜索loading（为了更好的用户体验）
  searchLoading.value = true
  // 防抖处理
  searchTimer = setTimeout(() => {
    searchKey.value = keyword || ''
    filterCourseList(keyword)

    // 短暂延迟后关闭loading，模拟搜索过程
    setTimeout(() => {
      searchLoading.value = false
    }, 100)
  }, 300)
}

// 处理下拉框滚动事件
function handleSelectScroll(e) {
  const { target } = e
  const { scrollTop, scrollHeight, clientHeight } = target

  // 触底判断：距离底部20px时触发加载
  if (scrollHeight - scrollTop - clientHeight <= 20) {
    loadMoreCourses()
  }
}

// 监听modal打开状态，打开时获取课程列表
watch(openModal, (newValue, oldValue) => {
  if (newValue && !oldValue) {
    // modal从关闭变为打开时，重置搜索状态
    searchKey.value = ''
    filterCourseList('')

    // 如果还没有获取过课程数据，则获取
    if (allCourseList.value.length === 0) {
      getCourseList()
    }
  }
})

// 监听关闭loading事件
onMounted(() => {
  emitter.on(EVENTS.CLOSE_LOADING_EVENT, () => {
    btnLoading.value = false
    formRef.value?.resetFields()
  })
})

// 组件卸载时移除事件监听
onUnmounted(() => {
  emitter.off(EVENTS.CLOSE_LOADING_EVENT)
  // 清理搜索定时器
  if (searchTimer) {
    clearTimeout(searchTimer)
    searchTimer = null
  }
})

// 处理图片删除
function handleRemove(file) {
  // 从 fileList 中移除图片后，同步更新 followImages
  updateFollowImages()
  return true
}

// 处理图片上传变化
function handleUploadChange(info) {
  fileList.value = info.fileList.map(normalizeUploadFileItem)
  updateFollowImages()
}

// 同步更新 followImages 字段
function updateFollowImages() {
  const images = fileList.value.map((file) => ({
    type: 1,
    url: ensureUploadUrl(file.url || file.response?.url || file.thumbUrl || ''),
  }))
  formState.followImages = images
}

// 手动触发验证
async function handleSubmit() {
  btnLoading.value = true
  try {
    await formRef.value.validate() // 关键3：通过引用调用验证方法

    // 准备提交数据，将 followImages 转为字符串
    const submitData = {
      ...formState,
      followImages: JSON.stringify(formState.followImages || []),
    }

    console.log('验证通过，提交数据:', submitData)
    emit('handleFollowUpSubmit', submitData)
  }
  catch (error) {
    console.log('验证失败:', error)
    btnLoading.value = false
  }
}
function closeFun() {
  formRef.value?.resetFields()
  openModal.value = false
  btnLoading.value = false

  // 清理搜索定时器
  if (searchTimer) {
    clearTimeout(searchTimer)
    searchTimer = null
  }

  // 重置图片列表为默认状态
  fileList.value = []

  // 重置搜索状态
  searchKey.value = ''
  searchLoading.value = false

  // 重置表单状态
  formState.followMethod = Number(FollowMethod.None)
  formState.intentLevel = props.record.intentLevel
  formState.nextFollowUpTime = ''
  formState.followUpStatus = props.record.followUpStatus
  formState.content = ''
  formState.followImages = ''
  
  // 重置意向课程时保留原始数据
  if (props.record.lessons && Array.isArray(props.record.lessons)) {
    formState.intentCourseIds = props.record.lessons.map(course => course.id.toString())
  } else {
    formState.intentCourseIds = []
  }
  
  formState.studentId = props.record.id

  // 同步更新 followImages 字段
  nextTick(() => {
    updateFollowImages()
  })
}
// 禁止当前时间之前的时间，不包含当前时间
function disabledDate(current) {
  return current && current < dayjs().startOf('day')
}
function disabledTime(current) {
  // 禁止当前时间之前的时间
  return {
    disabledHours: () => range(0, dayjs().hour()),
    disabledMinutes: (selectedHour) => {
      if (selectedHour === dayjs().hour()) {
        return range(0, dayjs().minute())
      }
      return []
    },
  }
}
function range(start, end) {
  const result = []
  for (let i = start; i < end; i++) {
    result.push(i)
  }
  return result
}
function beforeUpload(file) {
  const isImage = file.type === 'image/jpeg' || file.type === 'image/png' || file.type === 'image/bmp'
  if (!isImage) {
    message.error('只能上传BMP、JPG、JPEG、PNG格式的图片!')
    return false
  }
  const isLt4M = file.size / 1024 / 1024 < 4
  if (!isLt4M) {
    message.error('图片大小不能超过 4MB!')
    return false
  }
  return true
}

/**
 * 跟进记录附图：与学员头像一致，走七牛 customRequest（原 mocky action 仅为占位，无真实存储）
 */
function handleFollowImageUpload(options) {
  const { file, onSuccess, onError, onProgress } = options
  const rawFile = file.originFileObj || file

  if (!beforeUpload(rawFile)) {
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
      const key = `follow-up/${uuid}${ext}`

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
          console.error('跟进图片上传失败:', err)
          message.error(`上传失败: ${err?.message || '未知错误'}`)
          onError?.(err)
        },
        complete(res) {
          const fileUrl = buckethostname + res.key
          onSuccess?.({ url: fileUrl }, file)
        },
      })
    }
    catch (error) {
      console.error('获取七牛 token 失败:', error)
      message.error('获取上传凭证失败')
      onError?.(error)
    }
  })()
}
</script>

<template>
  <a-modal v-model:open="openModal" class="modal-content-box" :keyboard="false" :closable="false" :mask-closable="false"
    :width="800">
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ isEdit ? '编辑' : '添加' }}跟进记录{{ record.id }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <a-form ref="formRef" layout="vertical" :model="formState">
        <a-row :gutter="50">
          <a-col :span="12">
            <a-form-item label="跟进方式：" name="followMethod">
              <a-select v-model:value="formState.followMethod">
                <a-select-option v-for="(label, value) in FollowMethodLabel" :key="value" :value="Number(value)">
                  {{ label }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
          <!-- 意向度 单选 -->
          <a-col :span="12">
            <a-form-item name="intentLevel">
              <!-- 自定义label -->
              <template #label>
                <span>意向度：</span>
                <a-popover title="意向度">
                  <template #content>
                    意向度填写后将会同步更新学员对应信息
                  </template>
                  <!-- 问号icon -->
                  <QuestionCircleOutlined />
                </a-popover>
              </template>
              <a-radio-group v-model:value="formState.intentLevel" class="custom-radio">
                <a-radio v-for="(label, value) in IntentionLevelLabel" :key="value" :value="Number(value)">
                  {{ label }}
                </a-radio>
              </a-radio-group>
            </a-form-item>
          </a-col>
        </a-row>
        <!-- row 下次跟进时间 -->
        <a-row :gutter="50">
          <a-col :span="12">
            <a-form-item name="nextFollowUpTime">
              <template #label>
                <span>下次跟进时间：</span>
                <a-popover title="下次跟进时间">
                  <template #content>
                    1. 设置后，系统会提前一天以及一个小时前，发送通知提醒 <br>
                    2. 将有相关回访提醒，回访后可手动标记为已回访
                  </template>
                  <!-- 问号icon -->
                  <QuestionCircleOutlined />
                </a-popover>
              </template>
              <!-- 日期时间选择器，禁止当前时间之前的时间 -->
              <a-date-picker v-model:value="formState.nextFollowUpTime" :show-time="{ format: 'HH:mm' }"
                format="YYYY-MM-DD HH:mm" value-format="YYYY-MM-DD HH:mm" class="w-100%" :disabled-date="disabledDate"
                :disabled-time="disabledTime" />
            </a-form-item>
          </a-col>
          <!-- 跟进状态 -->
          <a-col :span="12">
            <a-form-item name="followUpStatus" :rules="[{ required: true, message: '请选择跟进状态' }]">
              <template #label>
                <span>跟进状态：</span>
                <a-popover title="跟进状态">
                  <template #content>
                    1. 跟进状态为手动标记，仅为区分当前的跟进状态，与学员实际系统状态无关 <br>
                    2. 跟进状态填写后将会同步更新学员对应信息
                  </template>
                  <!-- 问号icon -->
                  <QuestionCircleOutlined />
                </a-popover>
              </template>
              <a-select v-model:value="formState.followUpStatus">
                <a-select-option v-for="(label, value) in FollowUpStatusLabel" :key="value" :value="Number(value)">
                  {{ label }}
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
        <!-- row沟通内容 -->
        <a-row>
          <a-col :span="24">
            <a-form-item label="沟通内容：" name="content" :rules="[{ required: true, message: '请输入沟通内容' }]">
              <a-textarea v-model:value="formState.content" class="scrollbar" placeholder="请输入沟通内容"
                :auto-size="{ minRows: 5, maxRows: 5 }" />
            </a-form-item>
            <a-form-item-rest>
              <div class="mt--10px">
                <a-upload v-model:file-list="fileList" list-type="picture-card"
                  :custom-request="handleFollowImageUpload" :before-upload="beforeUpload" accept=".jpg,.jpeg,.png,.bmp"
                  @preview="handlePreview" @remove="handleRemove" @change="handleUploadChange">
                  <div v-if="fileList.length < 6">
                    <PlusOutlined class="text-20px" />
                  </div>
                </a-upload>
                <span class="text-#888 text-12px">最多上传6张，支持BMP、JPG、JPEG、PNG，单张图片不超过 4 MB</span>
                <a-modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="handleCancel">
                  <img alt="example" style="width: 100%" :src="previewImage">
                </a-modal>
              </div>
            </a-form-item-rest>
          </a-col>
        </a-row>
        <!-- 意向课程 row 24 -->
        <a-row class="mt-12px">
          <a-col :span="24">
            <a-form-item label="意向课程：" name="intentCourseIds">
              <!-- 多选 回显 label 最多选择3个课程，超出提示 最多选择3个课程 -->
              <a-select v-model:value="formState.intentCourseIds" mode="multiple" :max-tag-count="3"
                option-label-prop="label" :placeholder="maxCountPlaceholder" :loading="courseLoading || searchLoading"
                show-search :filter-option="false" :not-found-content="searchLoading ? undefined : '暂无数据'"
                @change="handleCourseChange" @search="searchCourses">
                <a-select-option v-for="item in courseList" :key="item.value" :value="item.value" :label="item.label">
                  <div class="flex flex-col py-2">
                    <div class="text-sm font-medium text-#222 mb-2">{{ item.label }}</div>
                    <a-space :size="5" class="w-100% flex flex-wrap">
                      <template v-for="tag in getCourseTagList(item)" :key="tag.key || tag.text">
                        <a-tag v-if="tag.type === 'tooltip'" :style="getTagStyle(tag.type)" :color="tag.color">
                          {{ tag.text }}
                          <a-tooltip>
                            <template #title>
                              {{ tag.tooltipTitle }}
                            </template>
                            <InfoCircleOutlined class="ml-1" />
                          </a-tooltip>
                        </a-tag>
                        <a-tag v-else :style="getTagStyle(tag.type)" :color="tag.color">
                          <span class="text-#ff9900" v-if="tag.text === '体验价'">{{ tag.text }}</span>
                          <span v-else>{{ tag.text }}</span>
                        </a-tag>
                      </template>
                    </a-space>
                  </div>
                </a-select-option>
                <!-- 空状态提示 -->
                <a-select-option v-if="!courseLoading && !searchLoading && courseList.length === 0" disabled
                  value="empty">
                  <div class="text-center py-3 text-#999">
                    {{ searchKey ? '未找到相关课程' : '暂无课程数据' }}
                  </div>
                </a-select-option>
              </a-select>
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </div>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        取消
      </a-button>
      <a-button type="primary" ghost :loading="btnLoading" @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
/* 添加旋转动画 */
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
  padding: 24px;
}

/* 自定义镂空样式 */
.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}

/* 课程选择器样式优化 */
.ant-select-dropdown .ant-select-item-option-content {
  padding: 4px 0;
}

.ant-select-dropdown .ant-tag {
  margin-right: 4px;
  margin-bottom: 4px;
}

/* 加载状态样式 */
.ant-select-dropdown .ant-select-item-option-disabled {
  cursor: default;
}

.ant-select-dropdown .ant-select-item-option-disabled:hover {
  background-color: transparent;
}

/* 搜索框样式优化 */
.ant-select-search-input {
  border: none !important;
  box-shadow: none !important;
}

.ant-select-search-input:focus {
  border: none !important;
  box-shadow: none !important;
}

/* Spin加载样式优化 */
.ant-spin-container {
  min-height: 40px;
}

.ant-spin-spinning .ant-select {
  pointer-events: auto;
}

.ant-spin-spinning .ant-select-disabled {
  pointer-events: none;
}

/* 下拉框loading状态样式 */
.ant-select-dropdown .ant-select-item-option[value="search-loading"],
.ant-select-dropdown .ant-select-item-option[value="empty"] {
  background-color: #fafafa !important;
}

.ant-select-dropdown .ant-select-item-option[value="search-loading"]:hover,
.ant-select-dropdown .ant-select-item-option[value="empty"]:hover {
  background-color: #fafafa !important;
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
</style>
