<script setup>
import { CloseOutlined, DeleteOutlined, FileWordOutlined, PictureOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { Upload } from 'ant-design-vue'
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { debounce } from 'lodash-es'
import * as qiniu from 'qiniu-js'
import ActiveCourseModal from '@/components/edu-center/registr-renewal/step01/active-course-modal.vue'
import MicroSchoolSettingsFields from './micro-school-settings-fields.vue'
import PackageItemCard from './package-item-card.vue'
import CustomTitle from '@/components/common/custom-title.vue'
import { getCoursePropertyOptionsApi } from '~@/api/edu-center/course-list'
import { getProcessContentPageApi } from '~@/api/edu-center/registr-renewal'
import { createProductPackageApi } from '@/api/edu-center/product-package'
import { getQiniuToken } from '@/api/qiniu'
import { useCourseAttribute } from '@/composables/useCourseAttribute'
import messageService from '@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:open', 'created'])

const isMobile = ref(false)
const isTablet = ref(false)
const saving = ref(false)
const productLoading = ref(false)
const propertyLoading = ref(false)
const microSchoolSettingModalOpen = ref(false)
const openSelectProducts = ref(false)
const previewVisible = ref(false)
const previewImage = ref('')
const previewTitle = ref('')
const showSettingCourseError = ref(false)
const activeCourseModalRef = ref(null)
const itemCardRefs = ref([])
const formRef = ref()
const propertyFormRef = ref()
const settingFormRef = ref()
const propertyMap = ref({})
const courseListOptions = ref([])
const courseListLoading = ref(false)
const courseListPagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  hasMore: true,
})
const courseSearchKey = ref('')

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

const drawerWidth = computed(() => {
  if (isMobile.value)
    return '100%'
  if (isTablet.value)
    return '90%'
  return '1165px'
})

const responsiveLabelCol = computed(() => {
  if (isMobile.value)
    return { span: 24 }
  return { span: 4 }
})

const responsiveWrapperCol = computed(() => {
  if (isMobile.value)
    return { span: 24 }
  return { span: 20 }
})

const settingResponsiveLabelCol = computed(() => {
  if (isMobile.value)
    return { span: 24 }
  return { span: 3 }
})

const settingResponsiveWrapperCol = computed(() => {
  if (isMobile.value)
    return { span: 24 }
  return { span: 21 }
})

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const { enabledCourseProperties, getEnabledCourseProperties } = useCourseAttribute()

const formState = reactive({
  name: '',
  onlineSale: true,
  isAllowEditWhenEnroll: false,
  productPackageProperties: {},
  items: [],
})

const settingFormState = reactive({
  title: '',
  images: [],
  description: [],
  isShow: false,
  isOnlineSale: false,
  buyLimit: false,
  oldBuy: true,
  newBuy: true,
  buyOne: false,
  allowType: 1,
  courseListIds: [],
  studentStatuses: [1, 2],
})

const selectedProductCount = computed(() => formState.items.length)

watch(() => formState.name, (value) => {
  if (!settingFormState.title) {
    settingFormState.title = value || ''
  }
})

watch(microSchoolSettingModalOpen, (value) => {
  if (value && !settingFormState.title) {
    settingFormState.title = formState.name || ''
  }
})

function resetForm() {
  formState.name = ''
  formState.onlineSale = true
  formState.isAllowEditWhenEnroll = false
  formState.productPackageProperties = {}
  formState.items = []

  settingFormState.title = ''
  settingFormState.images = []
  settingFormState.description = []
  settingFormState.isShow = false
  settingFormState.isOnlineSale = false
  settingFormState.buyLimit = false
  settingFormState.oldBuy = true
  settingFormState.newBuy = true
  settingFormState.buyOne = false
  settingFormState.allowType = 1
  settingFormState.courseListIds = []
  settingFormState.studentStatuses = [1, 2]

  showSettingCourseError.value = false
  previewVisible.value = false
  previewImage.value = ''
  previewTitle.value = ''
  courseListOptions.value = []
  courseSearchKey.value = ''
  courseListPagination.value = {
    current: 1,
    pageSize: 20,
    total: 0,
    hasMore: true,
  }
  formRef.value?.clearValidate?.()
  propertyFormRef.value?.clearValidate?.()
  settingFormRef.value?.clearValidate?.()
}

async function loadPropertyOptions() {
  propertyLoading.value = true
  try {
    await getEnabledCourseProperties()
    await Promise.all(enabledCourseProperties.value.map(async (item) => {
      const res = await getCoursePropertyOptionsApi({ propertyId: item.id })
      if (res.code === 200) {
        propertyMap.value[item.id] = res.result || []
      }
    }))
  }
  catch (error) {
    console.error('加载套餐属性选项失败:', error)
  }
  finally {
    propertyLoading.value = false
  }
}

async function loadCourseOptions(searchKey = '', isLoadMore = false) {
  if (courseListLoading.value) return

  courseListLoading.value = true
  try {
    const pageIndex = isLoadMore ? courseListPagination.value.current + 1 : 1
    const res = await getProcessContentPageApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: courseListPagination.value.pageSize,
        pageIndex,
        skipCount: (pageIndex - 1) * courseListPagination.value.pageSize,
      },
      queryModel: {
        delFlag: false,
        productType: 1,
        saleStatus: true,
        ...(searchKey ? { searchKey } : {}),
      },
      sortModel: {},
    })
    if (res.code === 200) {
      const newData = Array.isArray(res.result) ? res.result : []
      if (isLoadMore) {
        courseListOptions.value = [...courseListOptions.value, ...newData]
        courseListPagination.value.current = pageIndex
      } else {
        courseListOptions.value = newData
        courseListPagination.value.current = 1
      }
      courseListPagination.value.total = res.total || 0
      courseListPagination.value.hasMore = courseListOptions.value.length < courseListPagination.value.total
    }
  }
  catch (error) {
    console.error('加载课程选项失败:', error)
  }
  finally {
    courseListLoading.value = false
  }
}

const searchCourse = debounce((value) => {
  courseSearchKey.value = value
  loadCourseOptions(value, false)
}, 400)

function loadMoreCourses() {
  if (!courseListLoading.value && courseListPagination.value.hasMore) {
    loadCourseOptions(courseSearchKey.value, true)
  }
}

function handleCourseSelectPopupScroll(e) {
  const { target } = e
  if (target.scrollTop + target.offsetHeight === target.scrollHeight) {
    loadMoreCourses()
  }
}

function handleCourseSelectionChange(value) {
  const filteredValue = value.filter(id => !['load-more', 'loading-more', 'no-more'].includes(id))
  settingFormState.courseListIds = filteredValue
  showSettingCourseError.value = false
}

function generateSkuLabel(sku) {
  const { unit, quantity, price } = sku
  switch (unit) {
    case 1:
      return `${quantity}课时｜${price}元`
    case 2:
      return `${quantity}天｜${price}元`
    case 3:
      return `${quantity}月｜${price}元`
    case 4:
      return `${quantity}年｜${price}元`
    case 5:
      return `${price}元`
    default:
      return `${price}元`
  }
}

function buildPriceList(product) {
  const priceList = []
  ;(product.productSku || []).forEach((sku) => {
    priceList.push({
      label: sku.name,
      options: [
        {
          value: sku.id,
          label: generateSkuLabel(sku),
          lessonAudition: sku.lessonAudition,
          lessonModel: sku.lessonModel,
          unit: sku.unit,
          quantity: sku.quantity,
          price: sku.price,
        },
      ],
    })
  })
  priceList.sort((a, b) => {
    const unitPriority = { 1: 1, 2: 2, 3: 3, 4: 4, 5: 5 }
    const aOption = a.options[0]
    const bOption = b.options[0]
    const aPriority = unitPriority[aOption.unit] || 6
    const bPriority = unitPriority[bOption.unit] || 6
    if (aPriority !== bPriority) {
      return aPriority - bPriority
    }
    if ((aOption.quantity || 0) !== (bOption.quantity || 0)) {
      return (aOption.quantity || 0) - (bOption.quantity || 0)
    }
    return (bOption.price || 0) - (aOption.price || 0)
  })
  return priceList
}

function getItemKey(item) {
  return `${item.productType || 1}_${item.id}`
}

function normalizeSelectedProduct(item) {
  return {
    ...item,
    productSkuId: undefined,
    skuCount: 0,
    freeQuantity: 0,
    discountType: '0',
    discountNumber: undefined,
    discountRate: undefined,
    priceList: buildPriceList(item),
  }
}

function handleSelectProducts() {
  openSelectProducts.value = true
}

function handleProductModalConfirm(selectedProducts) {
  const selectedKeys = new Set(selectedProducts.map(getItemKey))
  const keepItems = formState.items.filter(item => selectedKeys.has(getItemKey(item)))
  const existingKeys = new Set(formState.items.map(getItemKey))
  const newItems = selectedProducts
    .filter(item => !existingKeys.has(getItemKey(item)))
    .map(normalizeSelectedProduct)

  formState.items.splice(0, formState.items.length, ...keepItems, ...newItems)
}

function cancelSelectProduct(index) {
  const target = formState.items[index]
  formState.items.splice(index, 1)
  activeCourseModalRef.value?.cancelCourseSelection?.(target)
}

function calculateItemTotal(item) {
  const sku = (item.productSku || []).find(sku => String(sku.id) === String(item.productSkuId))
  if (!sku || !item.skuCount) {
    return 0
  }
  let subtotal = Number(sku.price || 0) * Number(item.skuCount || 0)
  if (item.discountType === '1' && item.discountNumber) {
    subtotal -= Number(item.discountNumber || 0)
  }
  else if (item.discountType === '2' && item.discountRate) {
    subtotal = subtotal * (Number(item.discountRate || 0) / 10)
  }
  return Math.max(0, subtotal)
}

function getBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result)
    reader.onerror = error => reject(error)
  })
}

async function normalizeUploadFile(file) {
  const url = await getBase64(file)
  return {
    uid: file.uid || `${Date.now()}`,
    name: file.name,
    status: 'done',
    url,
    thumbUrl: url,
    originFileObj: file,
  }
}

function beforePackageImageUpload(file) {
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

function handlePackageImageUpload(options) {
  const { file, onSuccess, onError, onProgress } = options
  const rawFile = file.originFileObj || file

  if (!beforePackageImageUpload(rawFile)) {
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
      const key = `product-package/${uuid}${ext}`

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
          console.error('套餐主图上传失败:', err)
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
  previewImage.value = fileUrl || file.preview || ''
  previewVisible.value = true
  previewTitle.value = file.name || '图片预览'
}

function handlePackageImageChange(info) {
  settingFormState.images = (info.fileList || []).filter(file => file.status !== 'error')
}

function handlePackageImageRemove(file) {
  settingFormState.images = (settingFormState.images || []).filter(item => item.uid !== file.uid)
}

function buildImagesPayload() {
  return JSON.stringify(
    (settingFormState.images || [])
      .map(item => item.url || item.thumbUrl || item.response?.url)
      .filter(Boolean),
  )
}

function buildDescriptionPayload() {
  return JSON.stringify(settingFormState.description || [])
}

function buildPropertyPayload() {
  const propertyRows = []
  const subjectIds = []
  enabledCourseProperties.value.forEach((property) => {
    const selectedValue = formState.productPackageProperties[property.id]
    if (selectedValue === undefined || selectedValue === null || selectedValue === '') {
      return
    }
    const values = Array.isArray(selectedValue) ? selectedValue : [selectedValue]
    values.forEach((value) => {
      propertyRows.push({
        productPackagePropertyId: String(property.id),
        productPackagePropertyValue: String(value),
      })
      if (property.name === '科目') {
        subjectIds.push(Number(value))
      }
    })
  })
  return { propertyRows, subjectIds }
}

function validateMicroSchoolSettings(showMessage = true) {
  if (!settingFormState.buyLimit || !settingFormState.oldBuy || settingFormState.allowType !== 2) {
    showSettingCourseError.value = false
    return true
  }
  const isValid = Array.isArray(settingFormState.courseListIds) && settingFormState.courseListIds.length > 0
  showSettingCourseError.value = !isValid
  if (!isValid && showMessage) {
    messageService.error('请选择允许购买的课程')
  }
  return isValid
}

async function submitMicroSchoolSettingModal() {
  try {
    await settingFormRef.value?.validate?.()
  }
  catch {
    return
  }
  if (!validateMicroSchoolSettings(true)) {
    return
  }
  microSchoolSettingModalOpen.value = false
}

function cancelMicroSchoolSettingModal() {
  showSettingCourseError.value = false
  microSchoolSettingModalOpen.value = false
}

async function handleSubmit() {
  try {
    await formRef.value?.validate?.()
    await propertyFormRef.value?.validate?.()
  }
  catch {
    return
  }

  if (!formState.items.length) {
    messageService.error('请添加套餐商品')
    return
  }

  try {
    await Promise.all((itemCardRefs.value || []).filter(Boolean).map(ref => ref.validate?.()))
  }
  catch {
    return
  }

  if (!validateMicroSchoolSettings(true)) {
    microSchoolSettingModalOpen.value = true
    return
  }

  const { propertyRows, subjectIds } = buildPropertyPayload()
  saving.value = true
  try {
    const res = await createProductPackageApi({
      name: formState.name.trim(),
      onlineSale: formState.onlineSale,
      isAllowEditWhenEnroll: formState.isAllowEditWhenEnroll,
      title: settingFormState.title.trim() || formState.name.trim(),
      images: buildImagesPayload(),
      description: buildDescriptionPayload(),
      isShowMicoSchool: settingFormState.isShow,
      isOnlineSaleMicoSchool: settingFormState.isOnlineSale,
      buyRule: {
        enableBuyLimit: settingFormState.buyLimit,
        isAllowReturningStudent: settingFormState.oldBuy,
        isAllowFreshmanStudent: settingFormState.newBuy,
        limitOnePer: settingFormState.buyOne,
        allowType: settingFormState.allowType,
        relateProductIds: settingFormState.allowType === 2 ? settingFormState.courseListIds : [],
        studentStatuses: settingFormState.studentStatuses,
      },
      items: formState.items.map(item => ({
        productId: String(item.id),
        skuId: String(item.productSkuId),
        skuCount: Number(item.skuCount || 0),
        freeQuantity: Number(item.freeQuantity || 0),
        discountType: item.discountType === '0' ? undefined : Number(item.discountType),
        discountNumber: item.discountType === '2'
          ? Number(item.discountRate || 0)
          : Number(item.discountNumber || 0),
      })),
      subjectIds,
      productPackageProperties: propertyRows,
    })
    if (res.code === 200) {
      const createdId = res.result || res.data
      const productIds = [...new Set(formState.items.map(item => String(item.id)).filter(Boolean))]
      messageService.success('创建套餐成功')
      openDrawer.value = false
      emit('created', {
        id: String(createdId),
        productIds,
      })
      resetForm()
    }
  }
  catch (error) {
    console.error('创建套餐失败:', error)
    messageService.error('创建套餐失败')
  }
  finally {
    saving.value = false
  }
}

watch(openDrawer, async (value) => {
  if (!value)
    return
  resetForm()
  await Promise.all([
    loadPropertyOptions(),
    loadCourseOptions(),
  ])
})
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer"
      :width="drawerWidth"
      placement="right"
      :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false"
      @close="openDrawer = false"
    >
      <template #title>
        <div class="custom-header">
          <div class="drawer-title">
            创建套餐
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <a-spin :spinning="propertyLoading || productLoading || saving">
        <div class="package-drawer-body">
          <a-alert
            class="package-alert"
            message="微校和后台报名套餐时，可正常使用整单优惠和优惠券，请注意套餐优惠的力度，避免重复优惠"
            type="warning"
            show-icon
          />

          <div class="section-card">
            <CustomTitle title="基础设置" font-size="20px" font-weight="500" class="mb-24px" />

            <a-form
              ref="formRef"
              :model="formState"
              :label-col="responsiveLabelCol"
              :wrapper-col="responsiveWrapperCol"
            >
              <a-form-item
                label="套餐名称:"
                name="name"
                :rules="[
                  { required: true, message: '请输入套餐名称' },
                  { max: 20, message: '套餐名称需控制在 20 字以内' },
                ]"
              >
                <a-input
                  v-model:value="formState.name"
                  :maxlength="20"
                  placeholder="请输入（20字以内）"
                  class="field-width"
                />
              </a-form-item>

              <a-form-item
                label="售卖状态:"
                name="onlineSale"
                :rules="[{ required: true, message: '请选择售卖状态' }]"
              >
                <a-radio-group v-model:value="formState.onlineSale" class="custom-radio custom-radio2">
                  <a-radio :value="true">
                    在售
                  </a-radio>
                  <a-radio :value="false">
                    停售
                  </a-radio>
                </a-radio-group>
              </a-form-item>

              <a-form-item label="报名时修改办理内容:" name="isAllowEditWhenEnroll">
                <a-radio-group v-model:value="formState.isAllowEditWhenEnroll" class="custom-radio custom-radio2">
                  <a-radio :value="false">
                    禁止修改
                  </a-radio>
                  <a-radio :value="true">
                    允许修改
                  </a-radio>
                </a-radio-group>
              </a-form-item>

              <a-form-item :colon="false" name="items" :rules="[{ validator: async () => {
                if (!formState.items.length)
                  throw new Error('请添加套餐商品')
              } }]">
                <template #label>
                  <span class="required-star">*</span> 套餐商品:
                </template>
                <div class="product-entry-box">
                  <a-button type="primary" ghost @click="handleSelectProducts">
                    添加课程/学杂费/教材用品
                  </a-button>
                  <span class="entry-tip">（上限 20 个，当前选择 {{ selectedProductCount }} 个）</span>
                </div>
              </a-form-item>
            </a-form>

            <div v-if="formState.items.length" class="conductContent mt-5">
              <package-item-card
                v-for="(item, index) in formState.items"
                :ref="el => itemCardRefs[index] = el"
                :key="getItemKey(item)"
                :item="item"
                :total="calculateItemTotal(item)"
                @cancel="cancelSelectProduct(index)"
              />
            </div>
          </div>

          <div class="section-card">
            <CustomTitle title="套餐属性" font-size="20px" font-weight="500" class="mb-24px" />

            <a-form
              ref="propertyFormRef"
              :model="formState"
              :label-col="responsiveLabelCol"
              :wrapper-col="responsiveWrapperCol"
            >
              <a-form-item v-for="property in enabledCourseProperties" :key="property.id" :label="`${property.name}:`">
                <a-select
                  v-model:value="formState.productPackageProperties[property.id]"
                  :mode="property.name === '科目' ? 'multiple' : undefined"
                  allow-clear
                  show-search
                  :placeholder="`搜索${property.name}`"
                  class="field-width"
                >
                  <a-select-option
                    v-for="option in propertyMap[property.id] || []"
                    :key="option.id"
                    :value="option.id"
                  >
                    {{ option.name }}
                  </a-select-option>
                </a-select>
              </a-form-item>
            </a-form>
          </div>

          <div class="section-card">
            <CustomTitle title="微校设置" font-size="20px" font-weight="500" class="mb-24px" />

            <div class="micro-setting-row">
              <div class="micro-setting-label">
                套餐详情:
              </div>
              <a-button type="primary" ghost @click="microSchoolSettingModalOpen = true">
                编辑微校套餐详情
              </a-button>
            </div>
          </div>
        </div>
      </a-spin>

      <template #footer>
        <div class="drawer-footer">
          <a-button ghost type="primary" class="footer-btn" @click="openDrawer = false">
            取消
          </a-button>
          <a-button type="primary" class="footer-btn" :loading="saving" @click="handleSubmit">
            确认
          </a-button>
        </div>
      </template>
    </a-drawer>

    <ActiveCourseModal
      ref="activeCourseModalRef"
      v-model:open="openSelectProducts"
      :selected-courses="formState.items"
      @confirm="handleProductModalConfirm"
    />

    <a-modal
      v-model:open="microSchoolSettingModalOpen"
      centered
      wrap-class-name="microSchoolSettingModal"
      :keyboard="false"
      :closable="false"
      :mask-closable="false"
      :width="800"
      :body-style="{ padding: 0 }"
      @ok="submitMicroSchoolSettingModal"
      @cancel="cancelMicroSchoolSettingModal"
    >
      <template #title>
        <div class="modal-header">
          <span>编辑微校套餐详情</span>
          <a-button type="text" class="close-btn" @click="cancelMicroSchoolSettingModal">
            <template #icon>
              <CloseOutlined class="close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <div>
        <a-alert
          class="micro-alert"
          message="编辑微校套餐详情，主要用于此套餐在微校内的展示和售卖"
          type="info"
          show-icon
        />

        <a-form
          ref="settingFormRef"
          :model="settingFormState"
          class="setting-form"
          :label-col="settingResponsiveLabelCol"
          :wrapper-col="settingResponsiveWrapperCol"
        >
          <div class="px-24px py-16px">
            <CustomTitle title="套餐基本信息" font-size="16px" font-weight="500" class="mb-24px" />

            <a-form-item label="商品名称：" name="title" :rules="[{ required: true, message: '请输入商品名称' }]">
              <a-input v-model:value="settingFormState.title" class="w-300px" placeholder="请输入" />
            </a-form-item>

            <a-form-item label="商品主图：" name="images">
              <a-upload v-model:file-list="settingFormState.images" class="upload-list-inline"
                list-type="picture-card" :custom-request="handlePackageImageUpload" :before-upload="beforePackageImageUpload"
                @preview="handlePreview" @change="handlePackageImageChange" @remove="handlePackageImageRemove">
                <div v-if="settingFormState.images.length < 2">
                  <PlusOutlined class="text-16px" />
                </div>
              </a-upload>
              <div class="text-12px text-#888">
                建议比例 4:3
              </div>
            </a-form-item>

            <a-form-item label="详情介绍：" name="description">
              <a-space :size="16">
                <a-button type="primary" ghost>
                  <template #icon>
                    <PictureOutlined />
                  </template>
                  添加图片
                </a-button>
                <a-button type="primary" ghost>
                  <template #icon>
                    <FileWordOutlined />
                  </template>
                  添加文字
                </a-button>
              </a-space>
            </a-form-item>

            <micro-school-settings-fields
              :setting-form-state="settingFormState"
              settings-title="微校套餐设置"
              :show-micro-sale="true"
              micro-show-popover="关闭后，学员将无法在微校看到此套餐，但仍可通过分享购买"
              micro-sale-popover="开启后，学员能够在微校中查看并购买此套餐，或通过分享链接购买此套餐"
              purchase-target-label="套餐"
              :show-error="showSettingCourseError"
              :course-list-options="courseListOptions"
              :course-list-loading="courseListLoading"
              :course-list-pagination="courseListPagination"
              @search-course="searchCourse"
              @course-dropdown-visible-change="(open) => open && loadCourseOptions()"
              @course-popup-scroll="handleCourseSelectPopupScroll"
              @course-selection-change="handleCourseSelectionChange"
              @load-more-courses="loadMoreCourses"
            />
          </div>
        </a-form>
      </div>
    </a-modal>

    <a-modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="previewVisible = false">
      <img alt="preview" style="width: 100%" :src="previewImage">
    </a-modal>
  </div>
</template>

<style scoped lang="less">
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.custom-header,
.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.drawer-title {
  font-size: 20px;
  font-weight: 600;
  color: #1f2329;
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.close-icon {
  font-size: 20px;
  color: #1f2329;
}

.package-drawer-body {
  padding: 16px;
}

.package-alert,
.micro-alert {
  margin-bottom: 12px;
  border: 0;
  border-radius: 0;
}

.section-card {
  margin-top: 12px;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
}

.field-width {
  width: 60%;
}

.micro-field-width {
  width: 300px;
}

.required-star {
  color: #ff4d4f;
  font-family: SimSun, sans-serif;
  margin-right: 1px;
}

.action-outline-btn {
  height: 40px;
  padding-inline: 18px;
  border-radius: 10px;
  font-weight: 500;
}

.product-entry-box {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.entry-tip {
  color: #666;
}

.container-box {
  background: #fafafa;
  border-radius: 16px;
}

.container-box-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  height: 44px;
  padding: 10px 24px;
  background: #f0f5fe;
  border-top-left-radius: 16px;
  border-top-right-radius: 16px;
}

.right {
  display: flex;
  align-items: center;
  white-space: nowrap;
}

.right .price {
  font-weight: 600;
  color: #222;
}

.right .line {
  width: 1px;
  height: 12px;
  background: #ccc;
  margin: 0 18px;
}

.right .cancel {
  color: var(--pro-ant-color-primary);
  font-weight: bold;
  cursor: pointer;
}

.container-box-bottom {
  padding: 10px 8px 16px 24px;
  overflow-x: auto;
}

.container-box-bottom :deep(.ant-form-item-label label) {
  white-space: nowrap;
}

.container-box-bottom :deep(.ant-space) {
  align-items: flex-start;
  flex-wrap: nowrap;
}

.container-box-bottom :deep(.ant-space-item) {
  flex: 0 0 auto;
}

.tagCustom {
  display: inline-block;
  padding: 2px 5px;
  background: linear-gradient(212deg, #fad961, #f76b1c);
  border-radius: 8px;
  font-size: 10px;
  font-weight: 500;
  color: #fff;
  line-height: 12px;
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 16px;
  padding: 16px 24px;
}

.footer-btn {
  width: 140px;
  height: 50px;
  font-size: 18px;
}

.setting-body {
  padding: 16px 24px;
}

.micro-image-layout {
  display: flex;
  align-items: flex-start;
  gap: 20px;
  flex-wrap: wrap;
}

.cover-card {
  width: 160px;
  height: 160px;
  border: 1px dashed #b7cdfb;
  border-radius: 12px;
  overflow: hidden;
  background: #fafcff;
}

.cover-empty,
.cover-preview {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  cursor: pointer;
}

.cover-empty {
  font-size: 30px;
  color: #94a3b8;
}

.cover-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-actions {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
}

.remove-cover-btn {
  padding: 0;
}

.cover-hint {
  color: #86909c;
  font-size: 12px;
}

.description-editor {
  width: 100%;
}

.description-toolbar {
  margin-top: 16px;
}

.detail-image-list {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 16px;
}

.detail-image-card {
  position: relative;
  width: 104px;
  height: 104px;
  border-radius: 10px;
  overflow: hidden;
  background: #f5f7fa;
  border: 1px solid #edf0f5;
}

.detail-image-card img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  cursor: pointer;
}

.detail-image-remove {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 28px;
  height: 28px;
  background: rgba(255, 255, 255, 0.92);
  border-radius: 50%;
}

.switch-line {
  display: flex;
  align-items: center;
}

.switch-tip-icon {
  margin-left: 6px;
}

.allow-range-wrap {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 4px;
}

.allow-range-line {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: nowrap;
}

.allow-range-tag {
  display: inline-flex;
  align-items: center;
  height: 28px;
  padding-inline: 16px;
  color: #86909c;
  background: #f7f8fa;
  border-radius: 6px;
}

.allow-course-select {
  flex: 1;
  min-width: 0;
}

.allow-status-select {
  width: 380px;
  min-width: 380px;
}

.allow-range-text {
  color: #4e5969;
}

.micro-setting-row {
  display: flex;
  align-items: center;
  gap: 16px;
  min-height: 32px;
}

.micro-setting-label {
  width: 16.666667%;
  color: rgba(0, 0, 0, 0.88);
  text-align: right;
}

.setting-form {
  :deep(.ant-form-item) {
    margin-bottom: 10px;
  }
}

.custom-radio2 {
  :deep(.ant-radio-wrapper) {
    min-width: 145px;
  }
}

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

.mgnone {
  margin: 0 !important;
  flex: 0 0 auto;
}

.mgnone :deep(.ant-form-item-explain-error) {
  position: absolute;
}

:deep(.ant-form-item-explain-error) {
  font-size: 12px !important;
}

@media (max-width: 1024px) {
  .field-width {
    width: 100%;
  }

  .selected-card__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .package-drawer-body,
  .setting-body {
    padding: 12px;
  }

  .section-card,
  .container-box-bottom {
    padding: 14px;
  }

  .container-box-top,
  .right {
    align-items: flex-start;
    flex-direction: column;
    height: auto;
  }

  .drawer-footer {
    flex-direction: column;
    gap: 12px;
    padding: 12px;
  }

  .footer-btn,
  .micro-field-width {
    width: 100%;
  }

  .allow-course-select,
  .allow-status-select {
    width: 100%;
    min-width: 100%;
  }

  .allow-range-line {
    flex-wrap: wrap;
  }

  .micro-setting-row {
    flex-direction: column;
    align-items: flex-start;
  }

  .micro-setting-label {
    width: 100%;
    text-align: left;
  }
}
</style>
