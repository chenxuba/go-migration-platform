<script setup>
import { CloseOutlined, DeleteOutlined, PlusOutlined } from '@ant-design/icons-vue'
import { computed, reactive, ref, watch } from 'vue'
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

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const { enabledCourseProperties, getEnabledCourseProperties } = useCourseAttribute()
const saving = ref(false)
const courseLoading = ref(false)
const courseOptions = ref([])
const propertyLoading = ref(false)
const propertyMap = ref({})
const formRef = ref()

const packageItemDiscountTypeOptions = [
  { label: '无优惠', value: 0 },
  { label: '减免', value: 1 },
  { label: '折扣', value: 2 },
]

function createItemRow() {
  return {
    productId: undefined,
    skuId: undefined,
    skuCount: 1,
    freeQuantity: 0,
    discountType: 0,
    discountNumber: 0,
  }
}

const formState = reactive({
  name: '',
  title: '',
  imageUrl: '',
  descriptionText: '',
  onlineSale: true,
  isAllowEditWhenEnroll: false,
  isShowMicoSchool: true,
  isOnlineSaleMicoSchool: true,
  buyRule: {
    enableBuyLimit: true,
    isAllowReturningStudent: true,
    isAllowFreshmanStudent: true,
    limitOnePer: true,
    studentStatuses: [1, 2],
    relateType: 2,
    relateClassIds: [],
  },
  productPackageProperties: {},
  items: [createItemRow()],
})

function resetForm() {
  formState.name = ''
  formState.title = ''
  formState.imageUrl = ''
  formState.descriptionText = ''
  formState.onlineSale = true
  formState.isAllowEditWhenEnroll = false
  formState.isShowMicoSchool = true
  formState.isOnlineSaleMicoSchool = true
  formState.buyRule = {
    enableBuyLimit: true,
    isAllowReturningStudent: true,
    isAllowFreshmanStudent: true,
    limitOnePer: true,
    studentStatuses: [1, 2],
    relateType: 2,
    relateClassIds: [],
  }
  formState.productPackageProperties = {}
  formState.items = [createItemRow()]
  formRef.value?.clearValidate?.()
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
  courseLoading.value = true
  try {
    const res = await getProcessContentPageApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 50,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        delFlag: false,
        productType: 1,
        name: searchKey,
        saleStatus: true,
      },
      sortModel: {},
    })
    if (res.code === 200) {
      courseOptions.value = Array.isArray(res.result) ? res.result : []
    }
  }
  catch (error) {
    console.error('加载套餐商品失败:', error)
  }
  finally {
    courseLoading.value = false
  }
}

function getSkuOptionsByCourseId(courseId) {
  const course = courseOptions.value.find(item => String(item.id) === String(courseId))
  return Array.isArray(course?.productSku) ? course.productSku : []
}

function handleCourseChange(row) {
  row.skuId = undefined
}

function addItemRow() {
  formState.items.push(createItemRow())
}

function deleteItemRow(index) {
  if (formState.items.length === 1) {
    messageService.warning('至少保留一项套餐商品')
    return
  }
  formState.items.splice(index, 1)
}

function buildDescriptionPayload() {
  const text = `${formState.descriptionText || ''}`.trim()
  if (!text) {
    return '[]'
  }
  return JSON.stringify([{ type: 'text', text }])
}

function buildImagesPayload() {
  const url = `${formState.imageUrl || ''}`.trim()
  if (!url) {
    return '[]'
  }
  return JSON.stringify([url])
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

async function handleSubmit() {
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  const invalidItem = formState.items.find(item => !item.productId || !item.skuId || Number(item.skuCount || 0) <= 0)
  if (invalidItem) {
    messageService.error('请完善套餐内商品信息')
    return
  }

  const { propertyRows, subjectIds } = buildPropertyPayload()
  saving.value = true
  try {
    const res = await createProductPackageApi({
      name: formState.name.trim(),
      onlineSale: formState.onlineSale,
      isAllowEditWhenEnroll: formState.isAllowEditWhenEnroll,
      title: formState.title.trim() || formState.name.trim(),
      images: buildImagesPayload(),
      description: buildDescriptionPayload(),
      isShowMicoSchool: formState.isShowMicoSchool,
      isOnlineSaleMicoSchool: formState.isOnlineSaleMicoSchool,
      buyRule: formState.buyRule,
      items: formState.items.map(item => ({
        productId: String(item.productId),
        skuId: String(item.skuId),
        skuCount: Number(item.skuCount || 0),
        freeQuantity: Number(item.freeQuantity || 0),
        discountType: Number(item.discountType || 0) || undefined,
        discountNumber: Number(item.discountNumber || 0),
      })),
      subjectIds,
      productPackageProperties: propertyRows,
    })
    if (res.code === 200) {
      const createdId = res.result || res.data
      const productIds = [...new Set(formState.items.map(item => String(item.productId)).filter(Boolean))]
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
  <a-drawer
    v-model:open="openDrawer"
    :width="960"
    placement="right"
    :body-style="{ padding: '20px', background: '#f7f7fd' }"
    :closable="false"
  >
    <template #title>
      <div class="drawer-header">
        <span class="drawer-title">创建套餐</span>
        <a-button type="text" class="close-btn" @click="openDrawer = false">
          <CloseOutlined />
        </a-button>
      </div>
    </template>

    <a-spin :spinning="propertyLoading || courseLoading || saving">
      <a-form ref="formRef" :model="formState" layout="vertical" class="package-form">
        <div class="section-card">
          <CustomTitle title="基础设置" font-size="18px" font-weight="500" class="mb-16px" />
          <div class="grid-2">
            <a-form-item label="套餐名称" name="name" :rules="[{ required: true, message: '请输入套餐名称' }]">
              <a-input v-model:value="formState.name" placeholder="请输入套餐名称" />
            </a-form-item>
            <a-form-item label="微校标题" name="title">
              <a-input v-model:value="formState.title" placeholder="默认与套餐名称一致" />
            </a-form-item>
            <a-form-item label="售卖状态">
              <a-radio-group v-model:value="formState.onlineSale">
                <a-radio :value="true">
                  在售
                </a-radio>
                <a-radio :value="false">
                  停售
                </a-radio>
              </a-radio-group>
            </a-form-item>
            <a-form-item label="是否允许报读后编辑">
              <a-radio-group v-model:value="formState.isAllowEditWhenEnroll">
                <a-radio :value="false">
                  不允许
                </a-radio>
                <a-radio :value="true">
                  允许
                </a-radio>
              </a-radio-group>
            </a-form-item>
            <a-form-item label="开启微校售卖">
              <a-switch v-model:checked="formState.isOnlineSaleMicoSchool" />
            </a-form-item>
            <a-form-item label="开启微校展示">
              <a-switch v-model:checked="formState.isShowMicoSchool" />
            </a-form-item>
            <a-form-item label="封面图片 URL">
              <a-input v-model:value="formState.imageUrl" placeholder="可选，单图地址" />
            </a-form-item>
            <a-form-item label="详情描述">
              <a-textarea v-model:value="formState.descriptionText" :rows="4" placeholder="可选，先按纯文本录入" />
            </a-form-item>
          </div>
        </div>

        <div class="section-card">
          <CustomTitle title="套餐属性" font-size="18px" font-weight="500" class="mb-16px" />
          <div class="grid-3">
            <a-form-item v-for="property in enabledCourseProperties" :key="property.id" :label="property.name">
              <a-select
                v-model:value="formState.productPackageProperties[property.id]"
                :mode="property.name === '科目' ? 'multiple' : undefined"
                allow-clear
                show-search
                :placeholder="`请选择${property.name}`"
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
          </div>
        </div>

        <div class="section-card">
          <CustomTitle title="购买规则" font-size="18px" font-weight="500" class="mb-16px" />
          <div class="rule-grid">
            <a-checkbox v-model:checked="formState.buyRule.enableBuyLimit">
              启用购买限制
            </a-checkbox>
            <a-checkbox v-model:checked="formState.buyRule.isAllowReturningStudent">
              允许老学员购买
            </a-checkbox>
            <a-checkbox v-model:checked="formState.buyRule.isAllowFreshmanStudent">
              允许新学员购买
            </a-checkbox>
            <a-checkbox v-model:checked="formState.buyRule.limitOnePer">
              每人仅限购一份
            </a-checkbox>
          </div>
        </div>

        <div class="section-card">
          <div class="item-header">
            <CustomTitle title="套餐内商品" font-size="18px" font-weight="500" />
            <a-button type="primary" ghost @click="addItemRow">
              <PlusOutlined />
              添加商品
            </a-button>
          </div>
          <div v-for="(item, index) in formState.items" :key="index" class="package-item-row">
            <div class="grid-5">
              <a-form-item label="课程商品">
                <a-select
                  v-model:value="item.productId"
                  allow-clear
                  show-search
                  placeholder="请选择课程"
                  @change="handleCourseChange(item)"
                >
                  <a-select-option v-for="course in courseOptions" :key="course.id" :value="course.id">
                    {{ course.name }}
                  </a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item label="报价单">
                <a-select v-model:value="item.skuId" allow-clear show-search placeholder="请选择报价单">
                  <a-select-option
                    v-for="sku in getSkuOptionsByCourseId(item.productId)"
                    :key="sku.id"
                    :value="sku.id"
                  >
                    {{ sku.name }}
                  </a-select-option>
                </a-select>
              </a-form-item>
              <a-form-item label="数量">
                <a-input-number v-model:value="item.skuCount" :min="1" :precision="2" style="width: 100%" />
              </a-form-item>
              <a-form-item label="赠送数量">
                <a-input-number v-model:value="item.freeQuantity" :min="0" :precision="2" style="width: 100%" />
              </a-form-item>
              <a-form-item label="优惠类型">
                <a-select v-model:value="item.discountType" :options="packageItemDiscountTypeOptions" />
              </a-form-item>
            </div>
            <div class="grid-2">
              <a-form-item label="优惠值">
                <a-input-number
                  v-model:value="item.discountNumber"
                  :precision="item.discountType === 2 ? 1 : 2"
                  :min="0"
                  style="width: 100%"
                  :placeholder="item.discountType === 2 ? '折扣，例 9.5' : '减免金额'"
                />
              </a-form-item>
              <div class="delete-cell">
                <a-button danger type="link" @click="deleteItemRow(index)">
                  <DeleteOutlined />
                  删除
                </a-button>
              </div>
            </div>
          </div>
        </div>

        <div class="drawer-footer">
          <a-button @click="openDrawer = false">
            取消
          </a-button>
          <a-button type="primary" :loading="saving" @click="handleSubmit">
            创建套餐
          </a-button>
        </div>
      </a-form>
    </a-spin>
  </a-drawer>
</template>

<style scoped lang="less">
.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.drawer-title {
  font-size: 20px;
  font-weight: 600;
  color: #222;
}

.package-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-card {
  background: #fff;
  border-radius: 14px;
  padding: 18px 20px 20px;
}

.grid-2 {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.grid-3 {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0 16px;
}

.grid-5 {
  display: grid;
  grid-template-columns: 2fr 2fr 1fr 1fr 1fr;
  gap: 0 16px;
}

.rule-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px 24px;
}

.item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.package-item-row {
  border: 1px solid #edf0f5;
  border-radius: 12px;
  padding: 12px 14px 4px;
  margin-bottom: 12px;
}

.delete-cell {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding-top: 30px;
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
