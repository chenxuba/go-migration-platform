<script setup>
import { CloseOutlined, ExclamationCircleOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { Upload } from 'ant-design-vue'
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { debounce } from 'lodash-es'
import * as qiniu from 'qiniu-js'
import DescriptionBlockEditor from './description-block-editor.vue'
import SelectCourseRangeModal from './selectCourseRangeModal.vue'
import MicroSchoolSettingsFields from './micro-school-settings-fields.vue'
import CustomTitle from '@/components/common/custom-title.vue'
import { getCourseCategoryPageApi, getCourseDetailApi, getCoursePropertyOptionsApi, getCoursePageApi } from '~@/api/edu-center/course-list'
import { getQiniuToken } from '@/api/qiniu'
import { useCourseAttribute } from '@/composables/useCourseAttribute'
import messageService from '~@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  modalType: {
    type: String,
    default: 'create',
  },
  editCourseId: {
    type: [String, Number],
    default: null,
  },
})
const emit = defineEmits(['update:open', 'handleSubmit'])

// 响应式断点检测
const isMobile = ref(false)
const isTablet = ref(false)
const btnLoading = ref(false)
const microSchoolSettingModalOpen = ref(false)
function checkScreenSize() {
  isMobile.value = window.innerWidth < 768
  isTablet.value = window.innerWidth >= 768 && window.innerWidth < 1024
}

onMounted(() => {
  checkScreenSize()
  window.addEventListener('resize', checkScreenSize)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkScreenSize)
})

// 响应式drawer宽度
const drawerWidth = computed(() => {
  if (isMobile.value)
    return '100%'
  if (isTablet.value)
    return '90%'
  return '1165px'
})

// 响应式表单布局
const responsiveLabelCol = computed(() => {
  if (isMobile.value)
    return { span: 24 }
  return { span: 3 }
})

const responsiveWrapperCol = computed(() => {
  if (isMobile.value)
    return { span: 24 }
  return { span: 21 }
})

// 处理双向绑定
const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
const { enabledCourseProperties, getEnabledCourseProperties } = useCourseAttribute()
const editLoading = ref(false)
watch(enabledCourseProperties, async (newVal) => {
  // console.log(newVal);
  editLoading.value = true
  if (newVal.length > 0) {
    // 使用Promise.all等待所有异步操作完成
    try {
      await Promise.all(newVal.map(async (item) => {
        const res = await getCoursePropertyOptionsApi({
          'propertyId': item.id,
        })
        if (res.code === 200) {
          item.options = res.result
          if (props.modalType === 'create') {
            editLoading.value = false
          }
        }
      }))

    }
    catch (error) {
      console.error('获取课程属性选项失败:', error)
    }
  } else {
    editLoading.value = false
  }
  // 所有属性选项加载完成后，如果是编辑模式，获取课程详情
  if ((props.modalType === 'edit' || props.modalType === 'copy') && props.editCourseId) {
    await getCourseDetail(props.editCourseId)
  }
})
// 获取课程类别
const courseCategoryOptions = ref([])
// 获取课程类别
async function getCourseCategory(value) {
  // 获取课程类别
  try {
    const res = await getCourseCategoryPageApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': 1000, // 获取所有数据
        'pageIndex': 1,
        'skipCount': 1,
      },
      'queryModel': {
        'searchKey': value, // 使用单独的搜索关键字
      },
      'sortModel': {
        'orderBySortquantity': 2,
      },
    })
    if (res.code === 200) {
      courseCategoryOptions.value = res.result
    }
  }
  catch (error) {
    console.log(error)
  }
}

// 课程列表相关
const courseListOptions = ref([])
const courseListLoading = ref(false)
const courseListPagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  hasMore: true
})
const courseSearchKey = ref('')

// 获取课程列表
async function getCourseList(searchKey = '', isLoadMore = false) {
  if (courseListLoading.value) return

  courseListLoading.value = true

  try {
    const pageIndex = isLoadMore ? courseListPagination.value.current + 1 : 1

    const res = await getCoursePageApi({
      'pageRequestModel': {
        'needTotal': true,
        'pageSize': courseListPagination.value.pageSize,
        'pageIndex': pageIndex,
        'skipCount': (pageIndex - 1) * courseListPagination.value.pageSize,
      },
      'queryModel': {
        'searchKey': searchKey,
        'delFlag': false, // 只获取在售课程
      },
      'sortModel': {
        'byTotalSales': 0,
        'byUpdateTime': 0,
      },
    })

    if (res.code === 200) {
      const newData = res.result || []

      if (isLoadMore) {
        // 加载更多时追加数据
        courseListOptions.value = [...courseListOptions.value, ...newData]
        courseListPagination.value.current = pageIndex
      } else {
        // 首次加载或搜索时替换数据
        courseListOptions.value = newData
        courseListPagination.value.current = 1
      }

      courseListPagination.value.total = res.total || 0
      courseListPagination.value.hasMore = courseListOptions.value.length < courseListPagination.value.total
    }
  } catch (error) {
    console.error('获取课程列表失败:', error)
  } finally {
    courseListLoading.value = false
  }
}

// 搜索课程（带防抖）
const searchCourse = debounce((searchKey) => {
  courseSearchKey.value = searchKey
  getCourseList(searchKey, false)
}, 500)

// 加载更多课程
function loadMoreCourses() {
  if (!courseListLoading.value && courseListPagination.value.hasMore) {
    getCourseList(courseSearchKey.value, true)
  }
}

// 处理下拉框滚动到底部
function handleCourseSelectPopupScroll(e) {
  const { target } = e
  if (target.scrollTop + target.offsetHeight === target.scrollHeight) {
    loadMoreCourses()
  }
}

// 处理课程选择变化，过滤掉特殊值
function handleCourseSelectionChange(value) {
  // 过滤掉特殊的加载提示值
  const filteredValue = value.filter(id => !['load-more', 'loading-more', 'no-more'].includes(id))
  settingFormState.courseListIds = filteredValue
}

// 监听弹窗打开
watch(openDrawer, (newVal) => {
  if (newVal) {
    btnLoading.value = false
    // 获取课程类别
    getCourseCategory()
    getEnabledCourseProperties()
    // 初始化课程列表
    getCourseList()
    resetForm()
  }
})

const formRef = ref(null)
const settingFormRef = ref(null)
const settingFormState = reactive({
  title: undefined,
  images: [],
  descriptionBlocks: [],
  isShow: true,
  buyLimit: false,
  oldBuy: true,
  newBuy: true,
  buyOne: false,
  allowType: 1,
  courseListIds: [],
  studentStatuses: [1, 2],
})
const previewImage = ref('')
const previewVisible = ref(false)
const previewTitle = ref('')
const showError = ref(false)
// 表单数据
const formState = ref({
  name: undefined, // 课程名称
  title: undefined, // 微校的课程名称
  images: undefined, // 微校的商品主图
  description: undefined, // 微校的详情介绍
  isShowMicoSchool: false, // 是否开启微校展示
  courseCategory: undefined, // 课程类别
  courseProductProperties: [], // 课程属性  学季/学年/年级/班型/课程属性
  subjectIds: [], // 科目
  saleStatus: true, // 售卖状态
  courseType: 1, // 通用课程
  courseScope: [], // 课程范围
  teachMethod: 1, // 授课方式 1 班级授课 2 一对一授课
  type: 1, // 课程类型
  productSku: [
    // {
    //   id:undefined,//新增的时候无id
    //   lessonModel:undefined,//报价单模式
    //   name:undefined,//报价单名称
    //   unit:undefined,//计价方式
    //   quantity:undefined,//数量
    //   price:undefined,//总价金额
    //   lessonAudition:false,//体验价
    //   onlineSale:false,//开启微校售卖
    //   remark:undefined,//备注
    // }
  ], // 报价单
  buyRule: {
    enableBuyLimit: false,
    isAllowFreshmanStudent: true,
    relateProductIds: [],
    studentStatuses: [1, 2],
    isAllowReturningStudent: true,
    limitOnePer: false,
  },
  // 按课时报价单
  regularPrice: [],
  // 按时间段报价单
  timeBasedPrice: [],
  // 按金额报价单
  amountPrice: [],
})

// 监听microSchoolSettingModalOpen
watch(microSchoolSettingModalOpen, (newVal) => {
  if (newVal) {
    // 如果title为空 就 取formState的name
    if (!settingFormState.title) {
      settingFormState.title = formState.value.name
    }
  }
})

function getBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result)
    reader.onerror = error => reject(error)
  })
}

function beforeCourseImageUpload(file) {
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

function handleCourseImageUpload(options) {
  const { file, onSuccess, onError, onProgress } = options
  const rawFile = file.originFileObj || file

  if (!beforeCourseImageUpload(rawFile)) {
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
      const key = `course-micro/${uuid}${ext}`

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
          console.error('课程主图上传失败:', err)
          messageService.error(`上传失败: ${err?.message || '未知错误'}`)
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
      messageService.error('获取上传凭证失败')
      onError?.(error)
    }
  })()
}

async function handlePreview(file) {
  const fileUrl = file.url || file.response?.url || file.thumbUrl
  if (!fileUrl && !file.preview) {
    file.preview = await getBase64(file.originFileObj)
  }
  previewImage.value = fileUrl || file.preview
  previewVisible.value = true
  const urlStr = typeof fileUrl === 'string' ? fileUrl : ''
  previewTitle.value = file.name || (urlStr ? urlStr.substring(urlStr.lastIndexOf('/') + 1) : 'preview')
}
function handleCancelImg() {
  previewVisible.value = false
  previewTitle.value = ''
}

function handleCourseImageChange(info) {
  settingFormState.images = normalizeUploadFileList(info.fileList || [])
}

function handleCourseImageRemove(file) {
  settingFormState.images = (settingFormState.images || []).filter(item => item.uid !== file.uid)
}

function getUploadedImageUrl(item) {
  if (!item)
    return ''

  if (typeof item === 'string') {
    const text = item.trim()
    if (!text || text.startsWith('data:'))
      return ''
    return text
  }

  const candidates = [
    item.url,
    item.response?.url,
  ]
  const match = candidates.find(url => typeof url === 'string' && url.trim() && !url.startsWith('data:'))
  return match ? match.trim() : ''
}

function normalizeUploadImageList(imageList) {
  if (!Array.isArray(imageList))
    return []

  return imageList
    .map((item, index) => {
      if (item && typeof item === 'object' && item.uid) {
        return item
      }
      const url = getUploadedImageUrl(item)
      if (!url)
        return null
      return {
        uid: `course-image-${index}`,
        name: url.split('/').pop() || `image-${index + 1}`,
        status: 'done',
        url,
      }
    })
    .filter(Boolean)
}

function normalizeUploadFileList(fileList) {
  return (fileList || [])
    .filter(file => file?.status !== 'error')
    .map((item, index) => {
      const url = getUploadedImageUrl(item)
      if (!url)
        return item

      return {
        ...item,
        uid: item.uid || `course-image-${index}`,
        name: item.name || url.split('/').pop() || `image-${index + 1}`,
        status: 'done',
        url,
        thumbUrl: url,
        response: {
          ...(item.response || {}),
          url,
        },
      }
    })
}

function hasPendingCourseImages() {
  return (settingFormState.images || []).some(item => item?.status && item.status !== 'done')
}

function buildCourseImagePayload() {
  return JSON.stringify(
    (settingFormState.images || [])
      .map(item => getUploadedImageUrl(item))
      .filter(Boolean),
  )
}

// 课程属性选项
const courseLevelOptions = ref([
  // 可以从后端获取这些选项
])

// 添加一个强制更新的 ref
const forceUpdateKey = ref(0)

function resetForm() {
  // 先清除表单验证状态
  formRef.value?.clearValidate()
  // 重置表单字段
  formRef.value?.resetFields()
  microSchoolSettingFlag.value = false
  // 重置表单基础数据
  formState.value.regularPrice = []
  formState.value.timeBasedPrice = []
  formState.value.amountPrice = []

  // 重置课程属性字段 - 清空当前已知的属性
  enabledCourseProperties.value.forEach((property) => {
    formState.value[property.name] = undefined
  })

  // 额外清空可能存在的历史属性数据
  // 遍历 formState 中所有的动态属性字段并清空
  Object.keys(formState.value).forEach((key) => {
    // 如果是对象且包含 coursePropertyId 字段，说明是课程属性数据
    if (formState.value[key] && typeof formState.value[key] === 'object' && formState.value[key].coursePropertyId) {
      formState.value[key] = undefined
    }
  })

  // 重置基础字段到初始状态
  formState.value.name = undefined
  formState.value.id = undefined
  formState.value.uuid = undefined
  formState.value.version = undefined
  formState.value.title = undefined
  formState.value.images = undefined
  formState.value.description = undefined
  formState.value.isShowMicoSchool = true
  formState.value.courseCategory = undefined
  formState.value.courseProductProperties = []
  formState.value.subjectIds = []
  formState.value.saleStatus = true
  formState.value.courseType = 1
  formState.value.courseScope = []
  formState.value.teachMethod = 1
  allCommonTypeSelection.value = 'all-class'
  formState.value.type = 1
  formState.value.productSku = []

  // 重置购买规则到默认值
  formState.value.buyRule = {
    enableBuyLimit: false,
    isAllowFreshmanStudent: true,
    relateProductIds: [],
    studentStatuses: [1, 2],
    isAllowReturningStudent: true,
    limitOnePer: false,
    allowType: 1, // 默认任意课程
  }
  // 重置微校设置表单 settingFormState
  settingFormState.title = undefined
  settingFormState.images = []
  settingFormState.descriptionBlocks = []
  settingFormState.isShow = true
  settingFormState.buyLimit = false
  settingFormState.oldBuy = true
  settingFormState.newBuy = true
  settingFormState.buyOne = false
  settingFormState.allowType = 1
  settingFormState.courseListIds = []
  settingFormState.studentStatuses = [1, 2]
  settingFormState.relateProductIds = []

  // 清除微校设置表单验证状态
  settingFormRef.value?.clearValidate()
  settingFormRef.value?.resetFields()

  // 重置课程列表相关状态
  courseListOptions.value = []
  courseSearchKey.value = ''
  courseListPagination.value = {
    current: 1,
    pageSize: 20,
    total: 0,
    hasMore: true
  }

  // 重置选中的课程
  selectedCourses.value = []

  // 强制更新组件以确保视图刷新
  forceUpdateKey.value++

  // 使用 nextTick 确保 DOM 更新后再关闭弹窗
  // nextTick(() => {
  //   openDrawer.value = false;
  // });
}
// 查询课程详情
async function getCourseDetail(id) {
  console.log('查询课程详情')
  try {
    const res = await getCourseDetailApi(id)
    if (res.code === 200 && res.result) {
      const courseData = res.result
      console.log('课程详情数据:', courseData)

      // 辅助函数：根据属性ID和选项值获取显示名称
      const getPropertyValueName = (propertyId, propertyValue) => {
        const property = enabledCourseProperties.value.find(prop => prop.id === propertyId)
        if (!property || !property.options)
          return ''

        if (Array.isArray(propertyValue)) {
          // 多选情况（如科目）
          const names = propertyValue.map((value) => {
            const option = property.options.find(opt => opt.id === value)
            return option ? option.name : ''
          }).filter(name => name)
          return names.join(', ')
        }
        else {
          // 单选情况
          const option = property.options.find(opt => opt.id === propertyValue)
          return option ? option.name : ''
        }
      }

      // 回显基础信息
      formState.value.name = courseData.name
      if (props.modalType === 'copy') {
        formState.value.name = `${formState.value.name}（复制）`
      }

      // 非copy模式时设置课程的id、uuid、version字段
      if (props.modalType !== 'copy') {
        formState.value.id = courseData.id
        formState.value.uuid = courseData.uuid
        formState.value.version = courseData.version
      }

      formState.value.courseCategory = courseData.courseCategory
      formState.value.saleStatus = courseData.saleStatus
      formState.value.courseType = courseData.courseType || 1
      formState.value.teachMethod = courseData.teachMethod || 1
      allCommonTypeSelection.value = courseData.courseType === 4
        ? 'all-course'
        : ((courseData.teachMethod || 1) === 2 ? 'all-1v1' : 'all-class')
      formState.value.type = courseData.type || 1

      // 回显课程范围选择（使用 courseScopeInfo 字段进行回显）
      if (courseData.courseScopeInfo && courseData.courseScopeInfo.length > 0) {
        // 使用 courseScopeInfo 字段，直接包含了课程的 id 和 name 信息
        selectedCourses.value = courseData.courseScopeInfo.map(course => ({
          id: course.id,
          key: course.id.toString(),
          title: course.name,
          name: course.name,
        }))
        // 同时设置 courseScope 字段为课程ID数组，用于编辑已选
        formState.value.courseScope = courseData.courseScopeInfo.map(course => course.id)
      } else {
        selectedCourses.value = []
        formState.value.courseScope = []
      }

      // 回显课程属性
      if (courseData.courseProductProperties && Array.isArray(courseData.courseProductProperties)) {
        console.log('课程属性数据:', courseData.courseProductProperties)
        console.log('已启用属性:', enabledCourseProperties.value)

        courseData.courseProductProperties.forEach((property) => {
          // 根据coursePropertyId查找对应的enabled属性
          const enabledProperty = enabledCourseProperties.value.find(prop => prop.id === property.coursePropertyId)

          if (enabledProperty) {
            // 使用enabled属性中的name作为key，确保映射正确
            const mappedPropertyName = enabledProperty.name

            console.log(`映射属性: ${property.propertyName || property.propertyIdName} -> ${mappedPropertyName}`, {
              originalProperty: property,
              enabledProperty,
            })

            formState.value[mappedPropertyName] = {
              coursePropertyId: property.coursePropertyId,
              propertyName: mappedPropertyName,
              coursePropertyValue: property.coursePropertyValue,
              propertyValueName: property.propertyValueName || getPropertyValueName(property.coursePropertyId, property.coursePropertyValue),
            }
          }
          else {
            console.warn(`未找到匹配的属性:`, {
              coursePropertyId: property.coursePropertyId,
              propertyName: property.propertyName || property.propertyIdName,
              enabledProperties: enabledCourseProperties.value,
            })
          }
        })
      }

      // 回显科目信息
      if (courseData.subjectIds && Array.isArray(courseData.subjectIds)) {
        // 查找科目属性
        const subjectProperty = enabledCourseProperties.value.find(prop => prop.name === '科目')
        if (subjectProperty) {
          // 检查是否已经在courseProductProperties中处理过科目
          const alreadyProcessed = courseData.courseProductProperties?.some(prop => prop.coursePropertyId === subjectProperty.id)

          if (!alreadyProcessed) {
            const subjectValueName = getPropertyValueName(subjectProperty.id, courseData.subjectIds)
            formState.value['科目'] = {
              coursePropertyId: subjectProperty.id,
              propertyName: '科目',
              coursePropertyValue: courseData.subjectIds,
              propertyValueName: subjectValueName,
            }
            console.log('单独处理科目数据:', formState.value['科目'])
          }
          else {
            console.log('科目数据已在courseProductProperties中处理，跳过重复处理')
          }
        }
        formState.value.subjectIds = courseData.subjectIds
      }

      // 回显报价单信息
      if (courseData.productSku && Array.isArray(courseData.productSku)) {
        // 清空现有报价单
        formState.value.regularPrice = []
        formState.value.timeBasedPrice = []
        formState.value.amountPrice = []

        courseData.productSku.forEach((sku) => {
          const skuData = {
            name: sku.name,
            quantity: sku.quantity,
            price: sku.price,
            lessonAudition: sku.lessonAudition || false,
            onlineSale: sku.onlineSale || false,
            remark: sku.remark,
            unit: sku.unit,
            lessonModel: sku.lessonModel,
          }

          // 非copy模式时保留id、uuid、version字段
          if (props.modalType !== 'copy') {
            skuData.id = sku.id
            skuData.uuid = sku.uuid
            skuData.version = sku.version
          }

          // 根据lessonModel分类推入对应数组
          if (sku.lessonModel === 1) {
            // 按课时报价单
            formState.value.regularPrice.push(skuData)
          }
          else if (sku.lessonModel === 2) {
            // 按时间段报价单
            formState.value.timeBasedPrice.push(skuData)
          }
          else if (sku.lessonModel === 3) {
            // 按套餐报价单
            formState.value.amountPrice.push(skuData)
          }
        })
      }

      // 回显购买规则
      if (courseData.buyRule) {
        formState.value.buyRule = {
          enableBuyLimit: courseData.buyRule.enableBuyLimit || false,
          isAllowFreshmanStudent: courseData.buyRule.isAllowFreshmanStudent !== false, // 默认true
          relateProductIds: courseData.buyRule.relateProductIds || [],
          studentStatuses: courseData.buyRule.studentStatuses || [1, 2],
          isAllowReturningStudent: courseData.buyRule.isAllowReturningStudent !== false, // 默认true
          limitOnePer: courseData.buyRule.limitOnePer || false,
          allowType: courseData.buyRule.allowType || 1, // 默认任意课程
        }
      }

      // 回显微校设置
      formState.value.title = courseData.title || courseData.name
      formState.value.isShowMicoSchool = courseData.isShowMicoSchool !== false // 默认true
      formState.value.images = courseData.images
      formState.value.description = courseData.description

      // 设置微校设置表单数据
      settingFormState.title = courseData.title || courseData.name
      settingFormState.isShow = courseData.isShowMicoSchool !== false

      // 处理图片数据
      if (courseData.images) {
        try {
          const imageList = typeof courseData.images === 'string'
            ? JSON.parse(courseData.images)
            : courseData.images
          settingFormState.images = normalizeUploadImageList(imageList)
        }
        catch (e) {
          settingFormState.images = []
        }
      }

      // 处理详情描述
      if (courseData.description) {
        try {
          const descList = typeof courseData.description === 'string'
            ? JSON.parse(courseData.description)
            : courseData.description
          settingFormState.descriptionBlocks = Array.isArray(descList) ? descList : []
        }
        catch (e) {
          settingFormState.descriptionBlocks = []
        }
      }

      // 设置购买限制相关
      if (courseData.buyRule) {
        settingFormState.buyLimit = courseData.buyRule.enableBuyLimit || false
        settingFormState.oldBuy = courseData.buyRule.isAllowReturningStudent !== false
        settingFormState.newBuy = courseData.buyRule.isAllowFreshmanStudent !== false
        settingFormState.buyOne = courseData.buyRule.limitOnePer || false
        settingFormState.courseListIds = courseData.buyRule.relateProductIds || []
        settingFormState.studentStatuses = courseData.buyRule.studentStatuses || [1, 2]

        // 处理关联课程信息回显 - 如果有 relateProductInfos，将其添加到课程列表中
        if (courseData.buyRule.relateProductInfos && Array.isArray(courseData.buyRule.relateProductInfos)) {
          courseData.buyRule.relateProductInfos.forEach(productInfo => {
            // 检查课程是否已存在于课程列表中
            const existingCourse = courseListOptions.value.find(course => course.id === productInfo.id)
            if (!existingCourse) {
              // 如果不存在，则添加到课程列表中
              courseListOptions.value.push({
                id: productInfo.id,
                name: productInfo.name,
                // 可能需要的其他字段
                delFlag: false
              })
            }
          })
        }

        // 设置购买范围类型 - 优先使用后端返回的 allowType，如果没有则根据 relateProductIds 判断
        if (courseData.buyRule.allowType !== undefined && courseData.buyRule.allowType !== null) {
          settingFormState.allowType = courseData.buyRule.allowType
        } else {
          // 兼容旧数据：根据 relateProductIds 的长度判断
          if (courseData.buyRule.relateProductIds && courseData.buyRule.relateProductIds.length > 0) {
            settingFormState.allowType = 2 // 部分课程
          } else {
            settingFormState.allowType = 1 // 任意课程
          }
        }
      }

      // 强制更新组件
      forceUpdateKey.value++

      // 标记微校设置已配置
      if (courseData.title || courseData.images || courseData.description) {
        microSchoolSettingFlag.value = true
      }

      console.log('数据回显完成', {
        formState: formState.value,
        settingFormState,
        courseProperties: Object.keys(formState.value).filter(key =>
          formState.value[key]
          && typeof formState.value[key] === 'object'
          && formState.value[key].coursePropertyId,
        ).map(key => ({
          propertyName: key,
          data: formState.value[key],
        })),
      })
    }
    editLoading.value = false
  }
  catch (err) {
    console.error('获取课程详情失败:', err)
    editLoading.value = false
  }
}

// 添加按课时报价单
function addRegularPrice() {
  // 实现添加按课时报价单的逻辑
  console.log('添加按课时报价单')
  // 最多只能添加10条报价单
  if (formState.value.regularPrice.length >= 10) {
    messageService.error('最多只能添加10条报价单')
    return
  }
  formState.value.regularPrice.push({
    lessonModel: 1,
    name: undefined,
    unit: 1,
    quantity: undefined,
    price: undefined,
    lessonAudition: false,
    onlineSale: false,
    remark: undefined,
  })
  // 触发报价单校验
  formRef.value?.validateFields(['productSku'])
}

// 添加按时间段报价单
function addTimeBasedPrice() {
  // 实现添加按时间段报价单的逻辑
  console.log('添加按时间段报价单')
  // 最多只能添加10条报价单
  if (formState.value.timeBasedPrice.length >= 10) {
    messageService.error('最多只能添加10条报价单')
    return
  }
  formState.value.timeBasedPrice.push({
    lessonModel: 2,
    name: '',
    // 默认按天
    unit: 2,
    quantity: null,
    price: null,
    lessonAudition: false,
    onlineSale: false,
  })
  // 触发报价单校验
  formRef.value?.validateFields(['productSku'])
}

// 添加按套餐报价单
function addPackagePrice() {
  // 实现添加按套餐报价单的逻辑
  console.log('添加按套餐报价单')
  // 最多只能添加10条报价单
  if (formState.value.amountPrice.length >= 10) {
    messageService.error('最多只能添加10条报价单')
    return
  }
  formState.value.amountPrice.push({
    lessonModel: 3,
    name: '',
    lessonAudition: false,
    unit: 5,
    price: null,
    onlineSale: false,
  })
  // 触发报价单校验
  formRef.value?.validateFields(['productSku'])
}

function deletePriceItem(type, record) {
  const index = formState.value[type].findIndex(item => item === record)
  if (index > -1) {
    formState.value[type].splice(index, 1)
    messageService.success('删除成功')
  }
}

const microSchoolSettingFlag = ref(false)
// 提交微校modal
function submitMicroSchoolSettingModal() {
  if (settingFormState.courseListIds.length === 0 && settingFormState.allowType == 2) {
    showError.value = true
    messageService.error('请选择课程')
    return
  }
  settingFormRef.value.validate().then(async () => {
    microSchoolSettingFlag.value = true
    // 关闭modal
    microSchoolSettingModalOpen.value = false
  }).catch(() => {})
}

function cancelMicroSchoolSettingModal() {
  if (!microSchoolSettingFlag.value) {
    settingFormRef.value?.resetFields(['title'])
  }
  if (settingFormState.courseListIds.length === 0 && settingFormState.allowType == 2) {
    showError.value = false
    settingFormState.allowType = 1
  }
  microSchoolSettingModalOpen.value = false
}

// 原始提交函数
function _handleSubmit() {
  formRef.value.validate().then(async () => {
    if (hasPendingCourseImages()) {
      messageService.warning('请等待商品主图上传完成后再保存')
      return
    }
    const payload = buildCourseSubmitPayload()
    btnLoading.value = true
    emit('handleSubmit', payload)
  }).catch(() => {
    btnLoading.value = false
  })
}

// 防抖保存函数 (500ms延迟)
const handleSubmit = debounce(_handleSubmit, 500, {
  leading: true,
  trailing: false,
})

const openModal = ref(false)
// 选择课程弹出modal
const selectCourseRangeModalOpen = ref(false)
// 已选择的课程
const selectedCourses = ref([])
const allCommonTypeSelection = ref('all-class')

// 计算课程名称字符串
const courseNames = computed(() => {
  if (selectedCourses.value.length === 0) return ''

  // 获取课程名称，优先使用title，其次name
  const names = selectedCourses.value.map(course => course.title || course.name || '未知课程')

  if (names.length <= 3) {
    return names.join('、')
  } else {
    const firstThree = names.slice(0, 3).join('、')
    return `${firstThree}等${names.length}门课程`
  }
})

const isCommonCourse = computed({
  get: () => formState.value.courseType !== 1,
  set(value) {
    if (!value) {
      formState.value.courseType = 1
      return
    }
    if (formState.value.courseType === 1) {
      formState.value.courseType = allCommonTypeSelection.value === 'all-course' ? 4 : 2
    }
  },
})

const commonCourseMode = computed({
  get: () => formState.value.courseType === 3 ? 'partial' : 'all',
  set(value) {
    if (value === 'partial') {
      formState.value.courseType = 3
    } else {
      formState.value.courseType = allCommonTypeSelection.value === 'all-course' ? 4 : 2
    }
  },
})

// 仅改「全部课程 / 全部班课 / 全部1对1」时，v-model 只动 allCommonTypeSelection，不会改 courseType，导致保存仍为 2、重开被回显成全部1对1
watch(allCommonTypeSelection, (val) => {
  if (formState.value.courseType !== 2 && formState.value.courseType !== 4)
    return
  const next = val === 'all-course' ? 4 : 2
  if (formState.value.courseType !== next)
    formState.value.courseType = next
})

function selectCourseRange(index) {
  selectCourseRangeModalOpen.value = true
}

// 处理课程确认选择
function handleCourseConfirm(courses) {
  selectedCourses.value = [...courses]
  // 将选中的课程ID存储到表单数据中
  formState.value.courseScope = courses.map(course => course.id)
  console.log('父组件接收到选中的课程:', courses)
}

function handlePropertyChange(propertyId, value, propertyName) {
  if (value) {
    // 找到对应的属性对象
    const property = enabledCourseProperties.value.find(prop => prop.id === propertyId)
    // 从属性的选项中找到对应的选项名称
    const selectedOption = property?.options?.find(option => option.id === value)
    const propertyValueName = selectedOption?.name || ''

    // 设置为包含 propertyId 和选项 id 的对象格式
    formState.value[propertyName] = {
      coursePropertyId: propertyId,
      propertyName,
      coursePropertyValue: value,
      propertyValueName,
    }
  }
  else {
    // 如果没有选择值，则清空
    formState.value[propertyName] = undefined
  }
  console.log(`Property ${propertyName} changed:`, formState.value[propertyName])
}

function normalizeSubmitLessonMode(value) {
  return value === 3 ? 4 : value
}

function buildCourseSubmitPayload() {
  const lessonProductProperties = []
  formState.value.subjectIds = []

  enabledCourseProperties.value.forEach((property) => {
    const propertyData = formState.value[property.name]
    if (propertyData && propertyData.coursePropertyValue) {
      if (property.name === '科目') {
        if (Array.isArray(propertyData.coursePropertyValue)) {
          formState.value.subjectIds = [...propertyData.coursePropertyValue]
        } else {
          formState.value.subjectIds = [propertyData.coursePropertyValue]
        }
      } else {
        lessonProductProperties.push({
          lessonPropertyId: propertyData.coursePropertyId,
          lessonPropertyName: propertyData.propertyName,
          lessonPropertyValue: propertyData.coursePropertyValue,
          propertyValueName: propertyData.propertyValueName,
        })
      }
    }
  })

  return {
    id: formState.value.id,
    uuid: formState.value.uuid,
    version: formState.value.version,
    name: formState.value.name,
    courseCategory: formState.value.courseCategory,
    type: formState.value.type,
    title: formState.value.title || formState.value.name,
    images: buildCourseImagePayload(),
    description: JSON.stringify(settingFormState.descriptionBlocks || []),
    isShowMicoSchool: settingFormState.isShow,
    buyRule: {
      enableBuyLimit: settingFormState.buyLimit,
      allowType: settingFormState.allowType,
      isAllowFreshmanStudent: settingFormState.newBuy,
      isAllowReturningStudent: settingFormState.oldBuy,
      limitOnePer: settingFormState.buyOne,
      relateProductIds: settingFormState.courseListIds,
      studentStatuses: settingFormState.studentStatuses,
    },
    status: formState.value.saleStatus,
    productSku: [...formState.value.amountPrice, ...formState.value.regularPrice, ...formState.value.timeBasedPrice].map(item => ({
      id: item.id,
      uuid: item.uuid,
      version: item.version,
      name: item.name,
      quantity: item.quantity,
      unit: item.unit,
      price: item.price,
      onlineSale: item.onlineSale,
      lessonAudition: item.lessonAudition,
      lessonMode: normalizeSubmitLessonMode(item.lessonModel),
      remark: item.remark,
    })),
    lessonType: formState.value.teachMethod,
    lessonScope: formState.value.courseType,
    allowedLessonIds: formState.value.courseScope,
    subjectIds: formState.value.subjectIds,
    lessonProductProperties,
  }
}
</script>

<template>
  <div>
    <a-drawer v-model:open="openDrawer" :body-style="{ padding: '0', background: '#f7f7fd' }" :closable="false"
      :width="drawerWidth" placement="right" @close="openDrawer = false">
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            {{ props.modalType === 'edit' ? '编辑课程' : '创建课程' }}
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <a-spin :spinning="editLoading">
        <a-form ref="formRef" :model="formState" :label-col="responsiveLabelCol" :wrapper-col="responsiveWrapperCol">
          <div class="contenter flex flex-col p-2 md:p-4">
            <div class="content bg-white w-100% p-2 md:p-4 rounded-8px">
              <CustomTitle title="基础设置" font-size="20px" font-weight="500" class="mb-24px" />

              <a-form-item label="课程名称:" name="name" :rules="[{ required: true, message: '请输入课程名称' }]">
                <a-input v-model:value="formState.name" placeholder="请输入" class="w-full md:w-60%" />
              </a-form-item>

              <a-form-item label="课程类别:" name="courseCategory">
                <a-select v-model:value="formState.courseCategory" placeholder="搜索课程类别"
                  :style="{ width: isMobile ? '100%' : '60%' }" allow-clear show-search option-filter-prop="label">
                  <a-select-option v-for="item in courseCategoryOptions" :key="item.id" :label="item.name"
                    :value="item.id">
                    {{ item.name }}
                  </a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item v-for="item in enabledCourseProperties"
                :key="`${item.id}-${formState[item.name]?.coursePropertyValue || 'empty'}-${forceUpdateKey}`"
                :label="item.name" :name="item.name">
                <a-select :mode="item.name == '科目' ? 'multiple' : 'default'"
                  :value="formState[item.name]?.coursePropertyValue" :placeholder="`搜索${item.name}`"
                  :style="{ width: isMobile ? '100%' : '60%' }" allow-clear show-search option-filter-prop="label"
                  @change="(value) => handlePropertyChange(item.id, value, item.name)">
                  <a-select-option v-for="option in item.options" :key="option.id" :value="option.id"
                    :label="option.name">
                    {{ option.name }}
                  </a-select-option>
                </a-select>
              </a-form-item>

              <a-form-item label="售卖状态:" name="saleStatus" :rules="[{ required: true, message: '请选择售卖状态' }]">
                <a-radio-group v-model:value="formState.saleStatus" class="custom-radio custom-radio2">
                  <a-radio :value="true">
                    在售
                  </a-radio>
                  <a-radio :value="false">
                    停售
                  </a-radio>
                </a-radio-group>
              </a-form-item>

                <a-form-item label="授课方式:" name="teachMethod" :rules="[{ required: true, message: '请选择授课方式' }]">
                  <a-radio-group v-model:value="formState.teachMethod" class="custom-radio custom-radio2">
                    <a-radio :value="1">
                      班级授课
                    </a-radio>
                    <a-radio :value="2">
                      1v1授课
                    </a-radio>
                  </a-radio-group>
                </a-form-item>

                <a-form-item label="是否通用课程:" name="isCommonCourse">
                  <a-radio-group v-model:value="isCommonCourse" class="custom-radio custom-radio2">
                    <a-radio :value="false">
                      否
                    </a-radio>
                    <a-radio :value="true">
                      是
                    </a-radio>
                  </a-radio-group>
                </a-form-item>

                <a-form-item v-if="isCommonCourse" label="通用课程类型:" name="commonCourseMode">
                  <a-radio-group v-model:value="commonCourseMode" class="custom-radio custom-radio2">
                    <a-radio value="all">
                      全部通用
                    </a-radio>
                    <a-radio value="partial">
                      部分通用
                    </a-radio>
                  </a-radio-group>
                </a-form-item>

                <a-form-item v-if="isCommonCourse && commonCourseMode === 'all'" label="全部通用类型:" name="allCommonType">
                  <a-radio-group v-model:value="allCommonTypeSelection" class="custom-radio common-type-radio-group">
                    <a-radio value="all-course">
                      全部课程
                    </a-radio>
                    <a-radio value="all-class">
                      全部班课
                    </a-radio>
                    <a-radio value="all-1v1">
                      全部1对1
                    </a-radio>
                  </a-radio-group>
                </a-form-item>

                <a-form-item v-if="commonCourseMode === 'partial'" label="课程范围:"
                  name="courseScope" :rules="[{ required: true, message: '请选择课程范围' }]">
                  <div class="course-range-container">
                    <a-button type="primary" ghost class="w-full sm:w-auto"
                      @click="selectCourseRange(formState.teachMethod)">
                      {{ selectedCourses.length === 0 ? '选择课程' : '编辑已选' }}
                    </a-button>

                    <div v-if="selectedCourses.length > 0" class="selected-courses-display">
                      已选择{{ selectedCourses.length }}门：{{ courseNames }}
                    </div>
                  </div>
                </a-form-item>

              <a-form-item name="productSku" :rules="[
                {
                  validator: async (rule, value) => {
                    const isValid
                      = formState.regularPrice.length > 0
                      || formState.timeBasedPrice.length > 0
                      || formState.amountPrice.length > 0;

                    if (!isValid) {
                      throw new Error('最少添加一条报价单');
                    }
                  },
                  trigger: ['change', 'blur'],
                },
              ]">
                <template #label>
                  <span style="color: #ff4d4f;font-family: SimSun, sans-serif;" class="mr-px">*</span> 添加报价单
                </template>
                <div class="flex  sm:flex-row gap-3">
                  <a-button type="primary" ghost class="w-full sm:w-auto text-sm" @click="addRegularPrice">
                    添加按课时报价单
                  </a-button>
                  <a-button type="primary" ghost class="w-full sm:w-auto text-sm" @click="addTimeBasedPrice">
                    添加按时间段报价单
                  </a-button>
                  <a-button type="primary" ghost class="w-full sm:w-auto text-sm" @click="addPackagePrice">
                    添加按套餐报价单
                  </a-button>
                </div>
              </a-form-item>
              <!-- 按课时报价单 -->
              <div v-if="formState.regularPrice.length > 0" class="mx-0 md:mx-2 md:mr-10">
                <div
                  class="topBar bg-#f0f5fe rounded-8px rounded-lb-none rounded-rb-none flex justify-between items-center h-38px px-4 ">
                  <span class="font500 text-3.5">按课时</span>
                  <a-button type="primary" ghost size="small" @click="addRegularPrice">
                    <span class="text-3">添加报价单</span>
                  </a-button>
                </div>
                <div class="tables">
                  <div class="overflow-x-auto">
                    <a-table :pagination="false" :data-source="formState.regularPrice" :scroll="{ x: 800 }">
                      <a-table-column key="name" title="报价单名称" data-index="name" :width="isMobile ? 200 : 400" />
                      <a-table-column key="quantity" data-index="quantity" :width="isMobile ? 120 : 150">
                        <!-- 自定义title -->
                        <template #title>
                          <div><span class="text-red mr-1px">*</span>数量(课时)</div>
                        </template>
                      </a-table-column>
                      <a-table-column key="price" data-index="price" :width="isMobile ? 120 : 150">
                        <template #title>
                          <div><span class="text-red mr-1px">*</span>总价金额</div>
                        </template>
                      </a-table-column>
                      <!-- 体验价  -->
                      <a-table-column key="lessonAudition" title="体验价" data-index="lessonAudition" width="100" />
                      <!-- 开启微校售卖 自定义title -->
                      <a-table-column key="onlineSale" data-index="onlineSale" :width="isMobile ? 120 : 150">
                        <template #title>
                          <div class="whitespace-nowrap">
                            开启微校售卖
                            <a-popover color="#fff" title="开启微校售卖">
                              <template #content>
                                <div class="w-300px">
                                  开启后，学员能够在微校中查看并购买此课程，或通过分享链接购买此课程
                                </div>
                              </template>
                              <ExclamationCircleOutlined />
                            </a-popover>
                          </div>
                        </template>
                      </a-table-column>
                      <!-- 操作 -->
                      <a-table-column key="action" title="操作" data-index="action" width="80" />
                      <template #bodyCell="{ column, record, index }">
                        <!-- 双向绑定 报价单名称 -->
                        <template v-if="column.key === 'name'">
                          <a-form-item :name="['regularPrice', index, 'name']"
                            :rules="[{ required: true, message: '请输入' }]" :validate-trigger="['change', 'blur']">
                            <a-input v-model:value="record.name" placeholder="请输入"
                              :class="isMobile ? 'w-180px' : 'w-360px'" />
                          </a-form-item>
                        </template>
                        <!-- 双向绑定 数量 数字输入框 只能输入数字 限制两位小数 用a-input-number     -->
                        <template v-if="column.key === 'quantity'">
                          <a-form-item :name="['regularPrice', index, 'quantity']"
                            :rules="[{ required: true, message: '请输入' }]" :validate-trigger="['change', 'blur']">
                            <a-input-number v-model:value="record.quantity" placeholder="请输入"
                              :class="isMobile ? 'w-80px' : 'w-100px'" :min="1" :max="9999" :controls="false"
                              :precision="2" />
                          </a-form-item>
                        </template>
                        <!-- 双向绑定 总价金额 -->
                        <template v-if="column.key === 'price'">
                          <a-form-item :name="['regularPrice', index, 'price']"
                            :rules="[{ required: true, message: '请输入' }]" :validate-trigger="['change', 'blur']">
                            <a-input-number v-model:value="record.price" placeholder="请输入"
                              :class="isMobile ? 'w-80px' : 'w-100px'" :min="1" :max="99999" :controls="false"
                              :precision="2" />
                          </a-form-item>
                        </template>
                        <!-- 体验价 switch -->
                        <template v-if="column.key === 'lessonAudition'">
                          <a-form-item class="w-50px">
                            <a-switch v-model:checked="record.lessonAudition" />
                          </a-form-item>
                        </template>
                        <!-- 开启微校售卖 switch -->
                        <template v-if="column.key === 'onlineSale'">
                          <a-form-item class="w-100px">
                            <a-switch v-model:checked="record.onlineSale" />
                          </a-form-item>
                        </template>
                        <!-- 操作 -->
                        <template v-if="column.key === 'action'">
                          <a-form-item class="w-60px">
                            <a class="text-sm" @click="deletePriceItem('regularPrice', record)">删除</a>
                          </a-form-item>
                        </template>
                      </template>
                    </a-table>
                  </div>
                </div>
              </div>
              <!-- 按时间段报价单 -->
              <div v-if="formState.timeBasedPrice.length > 0" class="mx-0 md:mx-2 md:mr-10 mt-4px">
                <div
                  class="topBar bg-#f0f5fe rounded-8px rounded-lb-none rounded-rb-none flex justify-between items-center h-38px px-4 ">
                  <span class="font500 text-3.5">按时段</span>
                  <a-button type="primary" ghost size="small" @click="addTimeBasedPrice">
                    <span class="text-3">添加报价单</span>
                  </a-button>
                </div>
                <div class="tables">
                  <div class="overflow-x-auto">
                    <a-table :pagination="false" :data-source="formState.timeBasedPrice" :scroll="{ x: 900 }">
                      <a-table-column key="name" title="报价单名称" data-index="name" :width="isMobile ? 200 : 276" />
                      <!-- 计价方式 按天 按月 按年 -->
                      <a-table-column key="unit" data-index="unit" :width="isMobile ? 100 : 120">
                        <template #title>
                          <div><span class="text-red mr-1px">*</span>计价方式</div>
                        </template>
                      </a-table-column>
                      <a-table-column key="quantity" data-index="quantity" :width="isMobile ? 120 : 150">
                        <template #title>
                          <div><span class="text-red mr-1px">*</span>数量</div>
                        </template>
                      </a-table-column>
                      <a-table-column key="price" data-index="price" :width="isMobile ? 120 : 150">
                        <template #title>
                          <div><span class="text-red mr-1px">*</span>总价金额</div>
                        </template>
                      </a-table-column>
                      <!-- 体验价 switch -->
                      <a-table-column key="lessonAudition" title="体验价" data-index="lessonAudition" width="100" />
                      <!-- 开启微校售卖 switch -->
                      <a-table-column key="onlineSale" data-index="onlineSale" :width="isMobile ? 120 : 150">
                        <template #title>
                          <div class="whitespace-nowrap">
                            开启微校售卖
                            <a-popover color="#fff" title="开启微校售卖">
                              <template #content>
                                <div class="w-300px">
                                  开启后，学员能够在微校中查看并购买此课程，或通过分享链接购买此课程
                                </div>
                              </template>
                              <ExclamationCircleOutlined />
                            </a-popover>
                          </div>
                        </template>
                      </a-table-column>
                      <!-- 操作 -->
                      <a-table-column key="action" title="操作" data-index="action" width="80" />
                      <template #bodyCell="{ column, record, index }">
                        <template v-if="column.key === 'name'">
                          <a-form-item :name="['timeBasedPrice', index, 'name']"
                            :rules="[{ required: true, message: '请输入' }]" :validate-trigger="['change', 'blur']">
                            <a-input v-model:value="record.name" placeholder="请输入"
                              :class="isMobile ? 'w-180px' : 'w-250px'" />
                          </a-form-item>
                        </template>
                        <template v-if="column.key === 'unit'">
                          <a-form-item :name="['timeBasedPrice', index, 'unit']"
                            :rules="[{ required: true, message: '请选择' }]" :validate-trigger="['change', 'blur']">
                            <a-select v-model:value="record.unit" placeholder="请选择"
                              :style="{ width: isMobile ? '80px' : '80px' }">
                              <a-select-option :value="2">
                                按天
                              </a-select-option>
                              <a-select-option :value="3">
                                按月
                              </a-select-option>
                              <a-select-option :value="4">
                                按年
                              </a-select-option>
                            </a-select>
                          </a-form-item>
                        </template>
                        <template v-if="column.key === 'quantity'">
                          <a-form-item :name="['timeBasedPrice', index, 'quantity']"
                            :rules="[{ required: true, message: '请输入' }]" :validate-trigger="['change', 'blur']">
                            <a-input-number v-model:value="record.quantity" placeholder="请输入"
                              :class="isMobile ? 'w-80px' : 'w-100px'" :min="1" :max="9999" :controls="false"
                              :precision="0" />
                          </a-form-item>
                        </template>
                        <template v-if="column.key === 'price'">
                          <a-form-item :name="['timeBasedPrice', index, 'price']"
                            :rules="[{ required: true, message: '请输入' }]" :validate-trigger="['change', 'blur']">
                            <a-input-number v-model:value="record.price" placeholder="请输入"
                              :class="isMobile ? 'w-80px' : 'w-100px'" :min="1" :max="99999" :controls="false"
                              :precision="2" />
                          </a-form-item>
                        </template>
                        <template v-if="column.key === 'lessonAudition'">
                          <a-form-item class="w-50px">
                            <a-switch v-model:checked="record.lessonAudition" />
                          </a-form-item>
                        </template>
                        <template v-if="column.key === 'onlineSale'">
                          <a-form-item class="w-100px">
                            <a-switch v-model:checked="record.onlineSale" />
                          </a-form-item>
                        </template>
                        <template v-if="column.key === 'action'">
                          <a-form-item class="w-60px">
                            <a class="text-sm" @click="deletePriceItem('timeBasedPrice', record)">删除</a>
                          </a-form-item>
                        </template>
                      </template>
                    </a-table>
                  </div>
                </div>
              </div>
              <!-- 按金额报价单 -->
              <div v-if="formState.amountPrice.length > 0" class="mx-0 md:mx-2 md:mr-10 mt-4px">
                <div
                  class="topBar bg-#f0f5fe rounded-8px rounded-lb-none rounded-rb-none flex justify-between items-center h-38px px-5 ">
                  <span class="font500 text-3.5">按金额</span>
                  <a-button type="primary" ghost size="small" @click="addPackagePrice">
                    <span class="text-3">添加报价单</span>
                  </a-button>
                </div>
                <div class="tables">
                  <div class="overflow-x-auto">
                    <a-table :pagination="false" :data-source="formState.amountPrice" :scroll="{ x: 600 }">
                      <a-table-column key="name" title="报价单名称" data-index="name" :width="isMobile ? 200 : 400" />
                      <a-table-column key="price" data-index="price" :width="isMobile ? 200 : 400">
                        <template #title>
                          <div><span class="text-red mr-1px">*</span>总价金额</div>
                        </template>
                      </a-table-column>
                      <!-- 开启微校售卖 switch -->
                      <a-table-column key="onlineSale" data-index="onlineSale" :width="isMobile ? 120 : 150">
                        <template #title>
                          <div class="whitespace-nowrap">
                            开启微校售卖
                            <a-popover color="#fff" title="开启微校售卖">
                              <template #content>
                                <div class="w-300px">
                                  开启后，学员能够在微校中查看并购买此课程，或通过分享链接购买此课程
                                </div>
                              </template>
                              <ExclamationCircleOutlined />
                            </a-popover>
                          </div>
                        </template>
                      </a-table-column>
                      <!-- 操作 -->
                      <a-table-column key="action" title="操作" data-index="action" width="80" />
                      <template #bodyCell="{ column, record, index }">
                        <template v-if="column.key === 'name'">
                          <a-form-item :name="['amountPrice', index, 'name']"
                            :rules="[{ required: true, message: '请输入' }]" :validate-trigger="['change', 'blur']">
                            <a-input v-model:value="record.name" placeholder="请输入"
                              :class="isMobile ? 'w-180px' : 'w-360px'" />
                          </a-form-item>
                        </template>
                        <template v-if="column.key === 'price'">
                          <a-form-item :name="['amountPrice', index, 'price']"
                            :rules="[{ required: true, message: '请输入' }]" :validate-trigger="['change', 'blur']">
                            <a-input-number v-model:value="record.price" placeholder="请输入"
                              :class="isMobile ? 'w-180px' : 'w-320px'" :min="1" :max="99999" :controls="false"
                              :precision="2" />
                          </a-form-item>
                        </template>
                        <template v-if="column.key === 'onlineSale'">
                          <a-form-item class="w-100px">
                            <a-switch v-model:checked="record.onlineSale" />
                          </a-form-item>
                        </template>
                        <template v-if="column.key === 'action'">
                          <a-form-item class="w-60px">
                            <a class="text-sm" @click="deletePriceItem('amountPrice', record)">删除</a>
                          </a-form-item>
                        </template>
                      </template>
                    </a-table>
                  </div>
                </div>
              </div>
            </div>
            <div class="content bg-white w-100% p-2 md:p-4 rounded-8px mt-12px">
              <CustomTitle title="微校设置" font-size="20px" font-weight="500" class="mb-24px" />
              <a-form-item label="课程详情">
                <a-button type="primary" ghost class="w-full sm:w-auto text-sm"
                  @click="microSchoolSettingModalOpen = true">
                  编辑微校课程详情
                </a-button>
              </a-form-item>
            </div>
          </div>
        </a-form>
      </a-spin>
      <template #footer>
        <div class="flex  sm:flex-row justify-end gap-4 p-4">
          <a-button ghost type="primary" :class="isMobile ? 'w-full h-50px' : 'w-140px h-50px'" class="text-5"
            @click="openDrawer = false">
            取消
          </a-button>
          <a-button type="primary" :class="isMobile ? 'w-full h-50px' : 'w-140px h-50px'" class="text-5"
            :loading="btnLoading" @click="handleSubmit">
            保存
          </a-button>
        </div>
      </template>
    </a-drawer>
    <SelectCourseRangeModal v-model:open="selectCourseRangeModalOpen" :selected-courses="selectedCourses"
      @confirm="handleCourseConfirm" />
    <!-- 微校设置modal -->
    <a-modal v-model:open="microSchoolSettingModalOpen" centered wrap-class-name="microSchoolSettingModal"
      :keyboard="false" :closable="false" :mask-closable="false" :width="800" :body-style="{ padding: 0 }"
      @ok="submitMicroSchoolSettingModal" @cancel="cancelMicroSchoolSettingModal">
      <template #title>
        <div class="text-5 flex justify-between flex-center mb-0">
          <span>编辑微校课程详情</span>
          <a-button type="text" class="close-btn" @click="cancelMicroSchoolSettingModal">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div>
        <a-alert class="mb2 mt--2 text-#06f border-0 rounded-0" message="编辑微校课程详情，主要用于此课程在微校内的展示和售卖" type="info"
          show-icon />
        <a-form ref="settingFormRef" :model="settingFormState" class="setting-form" :label-col="responsiveLabelCol"
          :wrapper-col="responsiveWrapperCol">
          <div class="px-24px py-16px">
            <CustomTitle title="课程基本信息" font-size="16px" font-weight="500" class="mb-24px" />
            <a-form-item label="课程名称：" name="title" :rules="[{ required: true, message: '请输入课程名称' }]">
              <a-input v-model:value="settingFormState.title" class="w-300px" placeholder="请输入" />
            </a-form-item>
            <!-- 商品主图 -->
            <a-form-item label="商品主图：" name="images">
              <a-upload v-model:file-list="settingFormState.images" class="upload-list-inline"
                list-type="picture-card" :custom-request="handleCourseImageUpload" :before-upload="beforeCourseImageUpload"
                @preview="handlePreview" @change="handleCourseImageChange" @remove="handleCourseImageRemove">
                <div v-if="settingFormState.images.length < 2">
                  <PlusOutlined class="text-16px" />
                </div>
              </a-upload>
              <div class="text-12px text-#888">
                建议比例 4:3
              </div>
            </a-form-item>
            <!-- 详情介绍 -->
            <a-form-item label="详情介绍：" name="description">
              <description-block-editor
                v-model="settingFormState.descriptionBlocks"
                upload-folder="course/detail"
              />
            </a-form-item>
            <micro-school-settings-fields
              :setting-form-state="settingFormState"
              settings-title="微校课程设置"
              :show-micro-sale="false"
              micro-show-popover="关闭后，学员将无法在微校看到此课程，但仍可通过分享购买"
              purchase-target-label="课程"
              :show-error="showError"
              :course-list-options="courseListOptions"
              :course-list-loading="courseListLoading"
              :course-list-pagination="courseListPagination"
              @search-course="searchCourse"
              @course-dropdown-visible-change="(open) => open && getCourseList()"
              @course-popup-scroll="handleCourseSelectPopupScroll"
              @course-selection-change="handleCourseSelectionChange"
              @load-more-courses="loadMoreCourses"
            />
          </div>
        </a-form>
        <a-modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="handleCancelImg">
          <img alt="example" style="width: 100%" :src="previewImage">
        </a-modal>
      </div>
    </a-modal>
  </div>
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

// 水平对齐radio选项
.custom-radio2 {
  :deep(.ant-radio-wrapper) {
    min-width: 145px;

    // // 移动端缩短间距
    // @media (max-width: 768px) {
    //   min-width: 145px;
    //   margin-right: 8px;
    // }
  }
}

.tables {
  :deep(.ant-form-item) {
    margin-bottom: 0px;

    .ant-form-item-explain-error {
      position: absolute;
      font-size: 12px;
    }
  }
}

.setting-form {
  :deep(.ant-form-item) {
    margin-bottom: 10px;
  }
}



.common-type-radio-group {
  display: flex;
  gap: 8px 24px;
  flex-wrap: wrap;

  :deep(.ant-radio-wrapper) {
    min-width: 120px;
  }
}

// 课程范围选择样式
.course-range-container {
  .selected-courses-display {
    padding: 8px 12px;
    background: #e6f7ff;
    border: 1px solid #91d5ff;
    border-radius: 4px;
    color: #002766;
    font-size: 14px;
    line-height: 1.5;
    margin-top: 8px;
  }
}

// 响应式表单样式
@media (max-width: 768px) {
  .custom-header {
    padding: 0 16px;

    .text-5 {
      font-size: 16px;
    }
  }

  .contenter {
    padding: 8px;
  }

  .content {
    padding: 16px 12px;
  }

  // 移动端表单标签全宽显示
  :deep(.ant-form-item-label) {
    text-align: left;
    padding-bottom: 4px;
  }

  // 移动端按钮组垂直排列
  .flex-col {
    flex-direction: column;

    .a-button {
      margin-bottom: 8px;
    }
  }

  // 移动端表格横向滚动优化
  .overflow-x-auto {
    -webkit-overflow-scrolling: touch;
  }

  // 移动端footer按钮样式
  .template\:footer {
    .flex-col {
      gap: 12px;
    }
  }
}

// 平板样式
@media (min-width: 769px) and (max-width: 1024px) {
  .custom-header {
    .text-5 {
      font-size: 18px;
    }
  }

  .contenter {
    padding: 12px;
  }

  .content {
    padding: 20px 16px;
  }
}

.upload-list-inline {
  :deep(.ant-upload.ant-upload-select) {
    width: 160px !important;
    height: 107px !important;
  }

  :deep(.ant-upload-list-item-container) {
    width: 160px !important;
    height: 107px !important;
  }
}
</style>
