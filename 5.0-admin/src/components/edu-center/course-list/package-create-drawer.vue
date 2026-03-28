<script setup>
import { CloseOutlined, DeleteOutlined, ExclamationCircleOutlined, FileWordOutlined, PictureOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import { debounce } from 'lodash-es'
import ActiveCourseModal from '@/components/edu-center/registr-renewal/step01/active-course-modal.vue'
import PackageItemCard from './package-item-card.vue'
import CustomTitle from '@/components/common/custom-title.vue'
import { getCoursePropertyOptionsApi } from '~@/api/edu-center/course-list'
import { getProcessContentPageApi } from '~@/api/edu-center/registr-renewal'
import { createProductPackageApi } from '@/api/edu-center/product-package'
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
const descriptionTextareaRef = ref()
const propertyMap = ref({})
const courseOptions = ref([])

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
  descriptionText: '',
  detailImages: [],
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
  settingFormState.descriptionText = ''
  settingFormState.detailImages = []
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

async function loadCourseOptions(searchKey = '') {
  productLoading.value = true
  try {
    const res = await getProcessContentPageApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 200,
        pageIndex: 1,
        skipCount: 0,
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
      courseOptions.value = Array.isArray(res.result) ? res.result : []
    }
  }
  catch (error) {
    console.error('加载课程选项失败:', error)
  }
  finally {
    productLoading.value = false
  }
}

const searchCourse = debounce((value) => {
  loadCourseOptions(value)
}, 400)

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

function focusDescription() {
  nextTick(() => {
    descriptionTextareaRef.value?.focus?.()
  })
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

async function handleMainImageBeforeUpload(file) {
  settingFormState.images = [await normalizeUploadFile(file)]
  return false
}

async function handleDetailImageBeforeUpload(file) {
  if (settingFormState.detailImages.length >= 9) {
    messageService.warning('详情图片最多上传 9 张')
    return false
  }
  settingFormState.detailImages.push(await normalizeUploadFile(file))
  return false
}

function removeMainImage() {
  settingFormState.images = []
}

function removeDetailImage(index) {
  settingFormState.detailImages.splice(index, 1)
}

async function handlePreview(file) {
  if (!file.url && !file.preview && file.originFileObj) {
    file.preview = await getBase64(file.originFileObj)
  }
  previewImage.value = file.url || file.preview || ''
  previewVisible.value = true
  previewTitle.value = file.name || '图片预览'
}

function buildImagesPayload() {
  return JSON.stringify(
    (settingFormState.images || [])
      .map(item => item.url || item.thumbUrl)
      .filter(Boolean),
  )
}

function buildDescriptionPayload() {
  const blocks = []
  const text = `${settingFormState.descriptionText || ''}`.trim()
  if (text) {
    blocks.push({ type: 'text', text })
  }
  ;(settingFormState.detailImages || []).forEach((item) => {
    const url = item.url || item.thumbUrl
    if (url) {
      blocks.push({ type: 'image', url })
    }
  })
  return JSON.stringify(blocks)
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
          :label-col="responsiveLabelCol"
          :wrapper-col="responsiveWrapperCol"
        >
          <div class="setting-body">
            <CustomTitle title="套餐基本信息" font-size="16px" font-weight="500" class="mb-24px" />

            <a-form-item label="商品名称：" name="title" :rules="[{ required: true, message: '请输入商品名称' }]">
              <a-input v-model:value="settingFormState.title" class="micro-field-width" placeholder="请输入" />
            </a-form-item>

            <a-form-item label="商品主图：" name="images">
              <div class="micro-image-layout">
                <div class="cover-card">
                  <div
                    v-if="settingFormState.images.length"
                    class="cover-preview"
                    @click="handlePreview(settingFormState.images[0])"
                  >
                    <img :src="settingFormState.images[0].url || settingFormState.images[0].thumbUrl" alt="商品主图">
                  </div>
                  <div v-else class="cover-empty">
                    <PlusOutlined />
                  </div>
                </div>

                <div class="cover-actions">
                  <a-upload
                    :show-upload-list="false"
                    accept=".jpg,.jpeg,.png,.webp"
                    :before-upload="handleMainImageBeforeUpload"
                  >
                    <a-button type="primary" ghost class="action-outline-btn">
                      <template #icon>
                        <PictureOutlined />
                      </template>
                      {{ settingFormState.images.length ? '重新上传主图' : '上传主图' }}
                    </a-button>
                  </a-upload>
                  <a-button
                    v-if="settingFormState.images.length"
                    type="link"
                    danger
                    class="remove-cover-btn"
                    @click="removeMainImage"
                  >
                    删除图片
                  </a-button>
                  <div class="cover-hint">
                    建议比例 4:3
                  </div>
                </div>
              </div>
            </a-form-item>

            <a-form-item label="详情介绍：" name="descriptionText">
              <div class="description-editor">
                <a-textarea
                  ref="descriptionTextareaRef"
                  v-model:value="settingFormState.descriptionText"
                  :rows="5"
                  :maxlength="500"
                  show-count
                  placeholder="请输入，最多 500 字"
                />
                <a-space :size="16" class="description-toolbar">
                  <a-upload
                    :show-upload-list="false"
                    accept=".jpg,.jpeg,.png,.webp"
                    :before-upload="handleDetailImageBeforeUpload"
                  >
                    <a-button type="primary" ghost class="action-outline-btn">
                      <template #icon>
                        <PictureOutlined />
                      </template>
                      添加图片
                    </a-button>
                  </a-upload>
                  <a-button type="primary" ghost class="action-outline-btn" @click="focusDescription">
                    <template #icon>
                      <FileWordOutlined />
                    </template>
                    添加文字
                  </a-button>
                </a-space>

                <div v-if="settingFormState.detailImages.length" class="detail-image-list">
                  <div
                    v-for="(image, index) in settingFormState.detailImages"
                    :key="image.uid || index"
                    class="detail-image-card"
                  >
                    <img
                      :src="image.url || image.thumbUrl"
                      alt="详情图片"
                      @click="handlePreview(image)"
                    >
                    <a-button type="text" danger class="detail-image-remove" @click="removeDetailImage(index)">
                      <DeleteOutlined />
                    </a-button>
                  </div>
                </div>
              </div>
            </a-form-item>

            <CustomTitle title="微校套餐设置" font-size="16px" font-weight="500" class="mb-24px mt-24px" />

            <a-form-item label="微校展示：" name="isShow">
              <div class="switch-line">
                <a-switch v-model:checked="settingFormState.isShow" />
                <a-popover content="关闭后，学员将无法在微校看到此套餐，但仍可通过分享购买" title="微校展示">
                  <ExclamationCircleOutlined class="switch-tip-icon" />
                </a-popover>
              </div>
            </a-form-item>

            <a-form-item label="微校售卖：" name="isOnlineSale">
              <div class="switch-line">
                <a-switch v-model:checked="settingFormState.isOnlineSale" />
                <a-popover title="微校售卖">
                  <template #content>
                    开启后，学员能够在微校中查看并购买此套餐，或通过分享链接购买此套餐
                  </template>
                  <ExclamationCircleOutlined class="switch-tip-icon" />
                </a-popover>
              </div>
            </a-form-item>

            <a-form-item label="购买限制：" name="buyLimit">
              <a-switch v-model:checked="settingFormState.buyLimit" />
            </a-form-item>

            <div v-if="settingFormState.buyLimit">
              <a-form-item label="允许老生购买：" name="oldBuy">
                <div class="switch-line">
                  <a-switch v-model:checked="settingFormState.oldBuy" />
                  <a-popover title="允许老生购买">
                    <template #content>
                      <div class="w-440px">
                        <div>老生：在读、历史学员。</div>
                        <div>如果选择任意课程的在读学员，则机构所有在读学员均可购买此套餐。</div>
                        <div>如果选择部分课程，则只有购买过这些课程的学员可购买此套餐。</div>
                      </div>
                    </template>
                    <ExclamationCircleOutlined class="switch-tip-icon" />
                  </a-popover>
                </div>
              </a-form-item>

              <a-form-item v-if="settingFormState.oldBuy" label="允许：" name="allowType">
                <div class="allow-range-wrap">
                  <a-form-item-rest>
                    <a-radio-group v-model:value="settingFormState.allowType" class="custom-radio whitespace-nowrap">
                      <a-radio :value="1">
                        任意课程
                      </a-radio>
                      <a-radio :value="2">
                        部分课程
                      </a-radio>
                    </a-radio-group>
                  </a-form-item-rest>

                  <div class="allow-range-line">
                    <a-tag v-if="settingFormState.allowType === 1" class="allow-range-tag">
                      任意课程
                    </a-tag>
                    <template v-else>
                      <a-select
                        v-model:value="settingFormState.courseListIds"
                        mode="multiple"
                        :status="showSettingCourseError ? 'error' : ''"
                        class="allow-course-select"
                        placeholder="请选择课程（可多选）"
                        :max-tag-count="2"
                        :filter-option="false"
                        show-search
                        @search="searchCourse"
                        @change="showSettingCourseError = false"
                      >
                        <a-select-option v-for="course in courseOptions" :key="course.id" :value="course.id">
                          {{ course.name }}
                        </a-select-option>
                      </a-select>
                      <span class="allow-range-text">的</span>
                      <a-select
                        v-model:value="settingFormState.studentStatuses"
                        mode="multiple"
                        class="allow-status-select"
                        placeholder="请选择学员状态"
                      >
                        <a-select-option :value="1">
                          在读学员
                        </a-select-option>
                        <a-select-option :value="2">
                          历史学员
                        </a-select-option>
                      </a-select>
                      <span class="allow-range-text">购买</span>
                    </template>
                  </div>
                </div>
              </a-form-item>

              <a-form-item label="允许新生购买：" name="newBuy">
                <div class="switch-line">
                  <a-switch v-model:checked="settingFormState.newBuy" />
                  <a-popover title="允许新生购买">
                    <template #content>
                      新生：意向学员（未录入的学员）
                    </template>
                    <ExclamationCircleOutlined class="switch-tip-icon" />
                  </a-popover>
                </div>
              </a-form-item>

              <a-form-item label="限购 1 单：" name="buyOne">
                <div class="switch-line">
                  <a-switch v-model:checked="settingFormState.buyOne" />
                  <a-popover title="限购 1 单">
                    <template #content>
                      <div class="w-300px">
                        开启后，每个学员仅允许在微校购买此套餐一次，购买完成后不能重复购买。
                      </div>
                    </template>
                    <ExclamationCircleOutlined class="switch-tip-icon" />
                  </a-popover>
                </div>
              </a-form-item>
            </div>
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
