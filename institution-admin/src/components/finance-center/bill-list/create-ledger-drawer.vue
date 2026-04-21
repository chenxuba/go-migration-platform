<script setup>
import * as qiniu from 'qiniu-js'
import {
  AuditOutlined,
  BankOutlined,
  CheckOutlined,
  CloseOutlined,
  CustomerServiceOutlined,
  ExperimentOutlined,
  HomeOutlined,
  NotificationOutlined,
  PaperClipOutlined,
  PlaySquareOutlined,
  PlusOutlined,
  ReadOutlined,
  SafetyCertificateOutlined,
  ShopOutlined,
  TagOutlined,
  ThunderboltOutlined,
  WalletOutlined,
} from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, nextTick, reactive, ref, watch } from 'vue'
import { createLedgerApi, updateLedgerApi } from '@/api/finance-center/ledger'
import { getQiniuToken } from '@/api/qiniu'
import { payMethodOptionsWithIcons } from '@/components/common/pay-method-options-data'
import StaffSelect from '@/components/common/staff-select.vue'
import { useDrawer } from '@/composables/useDrawer'
import { useUserStore } from '@/stores/user'
import messageService from '@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  record: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['update:open', 'saved'])

const { openDrawer } = useDrawer(props, emit)
const userStore = useUserStore()

const LEDGER_CATEGORY_OTHER_BUSINESS = 'manual-other-business'
const LEDGER_CATEGORY_MANAGEMENT = 'manual-management-expense'
const LEDGER_CATEGORY_SALES = 'manual-sales-expense'
const LEDGER_CATEGORY_FINANCE = 'manual-finance-expense'

function createDefaultFormState() {
  return {
    amount: undefined,
    ledgerType: 1,
    payMethod: undefined,
    payAccount: '1',
    billRemarks: '',
    payDate: dayjs().format('YYYY-MM-DD'),
    dealStaffId: userStore.instUserId || undefined,
  }
}

const formState = reactive(createDefaultFormState())
const hydrating = ref(false)
const submitting = ref(false)
const isEditMode = computed(() => Boolean(props.record?.id))

const activeExpenseCategory = ref(LEDGER_CATEGORY_MANAGEMENT)
const activeLedgerItemKey = ref('')
const checkOptions = [...payMethodOptionsWithIcons]
const fileList = ref([])
const accountList = ref([{ value: '1', label: '默认账户' }])
const previewVisible = ref(false)
const previewImage = ref('')
const previewTitle = ref('')

const incomeCategoryItems = [
  { key: 'manual-exam-fee', categoryId: LEDGER_CATEGORY_OTHER_BUSINESS, label: '考试费用', icon: ReadOutlined },
  { key: 'manual-performance-fee', categoryId: LEDGER_CATEGORY_OTHER_BUSINESS, label: '演出费用', icon: PlaySquareOutlined },
  { key: 'manual-instrument-fee', categoryId: LEDGER_CATEGORY_OTHER_BUSINESS, label: '乐器费用', icon: CustomerServiceOutlined },
  { key: 'manual-meal-fee', categoryId: LEDGER_CATEGORY_OTHER_BUSINESS, label: '餐饮费用', icon: ShopOutlined },
  { key: 'manual-other-fee', categoryId: LEDGER_CATEGORY_OTHER_BUSINESS, label: '其他', icon: TagOutlined },
]

const expenseCategoryGroups = [
  {
    key: LEDGER_CATEGORY_MANAGEMENT,
    label: '管理费用',
    items: [
      { key: 'manual-office-supplies', categoryId: LEDGER_CATEGORY_MANAGEMENT, label: '办公用品', icon: PaperClipOutlined },
      { key: 'manual-water-fee', categoryId: LEDGER_CATEGORY_MANAGEMENT, label: '水费', icon: ExperimentOutlined },
      { key: 'manual-electricity-fee', categoryId: LEDGER_CATEGORY_MANAGEMENT, label: '电费', icon: ThunderboltOutlined },
      { key: 'manual-rent-fee', categoryId: LEDGER_CATEGORY_MANAGEMENT, label: '房租', icon: BankOutlined },
      { key: 'manual-property-fee', categoryId: LEDGER_CATEGORY_MANAGEMENT, label: '物业', icon: HomeOutlined },
      { key: 'manual-salary-fee', categoryId: LEDGER_CATEGORY_MANAGEMENT, label: '工资', icon: WalletOutlined },
      { key: 'manual-housing-fund-fee', categoryId: LEDGER_CATEGORY_MANAGEMENT, label: '公积金', icon: SafetyCertificateOutlined },
      { key: 'manual-social-insurance-fee', categoryId: LEDGER_CATEGORY_MANAGEMENT, label: '社保', icon: AuditOutlined },
    ],
  },
  {
    key: LEDGER_CATEGORY_SALES,
    label: '销售费用',
    items: [
      { key: 'manual-marketing-fee', categoryId: LEDGER_CATEGORY_SALES, label: '营销费用', icon: NotificationOutlined },
    ],
  },
  {
    key: LEDGER_CATEGORY_FINANCE,
    label: '财务费用',
    items: [{ key: 'manual-tax-fee', categoryId: LEDGER_CATEGORY_FINANCE, label: '税', icon: BankOutlined }],
  },
]

const allLedgerItems = [
  ...incomeCategoryItems,
  ...expenseCategoryGroups.flatMap(group => group.items),
]

const ledgerItemMap = allLedgerItems.reduce((acc, item) => {
  acc[item.key] = item
  return acc
}, {})

const activeExpenseItems = computed(() => {
  return (
    expenseCategoryGroups.find(
      item => item.key === activeExpenseCategory.value,
    )?.items || []
  )
})

function disabledDate(current) {
  return current > dayjs().endOf('day')
}

function selectExpenseCategory(key) {
  activeExpenseCategory.value = key
  if (activeLedgerItemKey.value && ledgerItemMap[activeLedgerItemKey.value]?.categoryId !== key)
    activeLedgerItemKey.value = ''
}

function selectLedgerItem(key) {
  activeLedgerItemKey.value = key
}

function getBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result)
    reader.onerror = error => reject(error)
  })
}

async function handlePreview(file) {
  if (!file.url && !file.preview)
    file.preview = await getBase64(file.originFileObj)

  previewImage.value = file.url || file.preview
  previewVisible.value = true
  previewTitle.value = file.name || file.url?.substring(file.url.lastIndexOf('/') + 1) || ''
}

function handleCancelImg() {
  previewVisible.value = false
  previewTitle.value = ''
}

function beforeUpload(file) {
  const isImage = file.type === 'image/jpeg' || file.type === 'image/png' || file.type === 'image/bmp'
  if (!isImage) {
    messageService.error('只能上传 BMP / JPG / JPEG / PNG 格式的图片')
    return false
  }
  const isLt4M = file.size / 1024 / 1024 < 4
  if (!isLt4M) {
    messageService.error('图片大小不能超过 4 MB')
    return false
  }
  if (fileList.value.length >= 3) {
    messageService.warning('最多上传 3 张图片')
    return false
  }
  return true
}

function handleLedgerImageUpload(options) {
  const { file, onSuccess, onError, onProgress } = options
  const rawFile = file.originFileObj || file

  if (!beforeUpload(rawFile)) {
    onError?.(new Error('文件校验未通过'))
    return
  }

  ;(async () => {
    try {
      const tokenRes = await getQiniuToken()
      const { token, uuid, buckethostname } = tokenRes.result || {}
      if (!token || !uuid || !buckethostname)
        throw new Error('上传凭证无效')

      const ext = rawFile.name?.includes('.')
        ? rawFile.name.substring(rawFile.name.lastIndexOf('.'))
        : (rawFile.type === 'image/png' ? '.png' : '.jpg')
      const key = `finance/manual-ledger/${uuid}${ext}`

      const observable = qiniu.upload(rawFile, key, token, {
        fname: rawFile.name,
        mimeType: rawFile.type,
      }, {
        useCdnDomain: true,
        region: qiniu.region.z0,
      })

      observable.subscribe({
        next(res) {
          onProgress?.({ percent: Math.floor(res.total.percent) })
        },
        error(err) {
          console.error('上传记账图片失败:', err)
          messageService.error(err?.message || '图片上传失败')
          onError?.(err)
        },
        complete(res) {
          onSuccess?.({ url: `${buckethostname}${res.key}` }, file)
        },
      })
    }
    catch (error) {
      console.error('获取上传凭证失败:', error)
      messageService.error(error?.message || '获取上传凭证失败')
      onError?.(error)
    }
  })()
}

function normalizeFileListToUrls() {
  return fileList.value
    .map((file) => {
      const uploadedUrl = file.url || file.response?.url
      return typeof uploadedUrl === 'string' ? uploadedUrl : ''
    })
    .filter(Boolean)
}

function resetDrawerState() {
  Object.assign(formState, createDefaultFormState())
  activeExpenseCategory.value = LEDGER_CATEGORY_MANAGEMENT
  activeLedgerItemKey.value = ''
  fileList.value = []
  previewVisible.value = false
  previewImage.value = ''
  previewTitle.value = ''
}

function resolveRecordLedgerSelection(record) {
  const ledgerType = Number(record?.type || 1)
  const subCategoryId = String(record?.ledgerSubCategoryId || '').trim()
  const selectedItem = subCategoryId ? ledgerItemMap[subCategoryId] : undefined
  const fallbackCategoryId = ledgerType === 1
    ? LEDGER_CATEGORY_OTHER_BUSINESS
    : LEDGER_CATEGORY_MANAGEMENT

  return {
    categoryId: selectedItem?.categoryId || String(record?.ledgerCategoryId || '').trim() || fallbackCategoryId,
    subCategoryId: selectedItem?.key || subCategoryId,
  }
}

function fillDrawerByRecord(record) {
  hydrating.value = true
  const selection = resolveRecordLedgerSelection(record)
  Object.assign(formState, {
    amount: record?.amount ?? undefined,
    ledgerType: Number(record?.type || 1),
    payMethod: record?.payMethod ?? undefined,
    payAccount: String(record?.accountId || '1'),
    billRemarks: record?.paymentVoucher?.text || '',
    payDate: record?.payTime ? dayjs(record.payTime).format('YYYY-MM-DD') : dayjs().format('YYYY-MM-DD'),
    dealStaffId: record?.dealStaffId ? String(record.dealStaffId) : (userStore.instUserId || undefined),
  })
  activeExpenseCategory.value = selection.categoryId
  activeLedgerItemKey.value = selection.subCategoryId
  fileList.value = Array.isArray(record?.paymentVoucher?.images)
    ? record.paymentVoucher.images.map((url, index) => ({
        uid: `existing-${index}`,
        name: `image-${index + 1}.png`,
        status: 'done',
        url,
      }))
    : []
  nextTick(() => {
    hydrating.value = false
  })
}

function initDrawer() {
  if (isEditMode.value)
    fillDrawerByRecord(props.record)
  else
    resetDrawerState()
}

function validateBeforeSubmit() {
  if (!(Number(formState.amount) > 0)) {
    messageService.warning('请输入账单金额')
    return false
  }
  if (!formState.payDate) {
    messageService.warning('请选择支付日期')
    return false
  }
  if (!String(formState.dealStaffId || '').trim()) {
    messageService.warning('请选择经办人')
    return false
  }
  if (!activeLedgerItemKey.value || !ledgerItemMap[activeLedgerItemKey.value]) {
    messageService.warning('请选择账单分类')
    return false
  }
  if (!formState.payMethod) {
    messageService.warning('请选择收款方式')
    return false
  }
  if (!String(formState.payAccount || '').trim()) {
    messageService.warning('请选择支付账户')
    return false
  }
  if (fileList.value.some(file => file.status === 'uploading')) {
    messageService.warning('图片上传中，请稍候再提交')
    return false
  }
  return true
}

async function handleSubmit() {
  if (submitting.value || !validateBeforeSubmit())
    return

  const selectedLedgerItem = ledgerItemMap[activeLedgerItemKey.value]
  const payload = {
    id: isEditMode.value ? String(props.record?.id || '') : undefined,
    amount: Number(formState.amount),
    remark: formState.billRemarks || '',
    images: normalizeFileListToUrls(),
    payTime: formState.payDate,
    dealStaffId: String(formState.dealStaffId),
    payMethod: Number(formState.payMethod),
    type: Number(formState.ledgerType),
    ledgerCategoryId: selectedLedgerItem.categoryId,
    ledgerSubCategoryId: selectedLedgerItem.key,
    accountId: String(formState.payAccount),
  }

  try {
    submitting.value = true
    const res = isEditMode.value
      ? await updateLedgerApi(payload)
      : await createLedgerApi(payload)
    messageService.success(isEditMode.value ? '账单编辑成功' : '记账成功')
    emit('saved', {
      id: res?.result?.id || payload.id || '',
      isEdit: isEditMode.value,
    })
    openDrawer.value = false
  }
  catch (error) {
    console.error(isEditMode.value ? '编辑账单失败:' : '创建账单失败:', error)
    messageService.error(error?.message || (isEditMode.value ? '编辑账单失败' : '创建账单失败'))
  }
  finally {
    submitting.value = false
  }
}

watch(
  () => formState.ledgerType,
  () => {
    if (hydrating.value)
      return
    activeExpenseCategory.value = LEDGER_CATEGORY_MANAGEMENT
    activeLedgerItemKey.value = ''
  },
)

watch(
  () => props.open,
  (open) => {
    if (!open) {
      if (!isEditMode.value)
        resetDrawerState()
      return
    }
    initDrawer()
  },
  { immediate: true },
)

watch(
  () => props.record,
  () => {
    if (props.open)
      initDrawer()
  },
)
</script>

<template>
  <a-drawer
    v-model:open="openDrawer"
    :push="{ distance: 80 }"
    :body-style="{ padding: '0', background: '#f7f7fd' }"
    :closable="false"
    width="800px"
    placement="right"
  >
    <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            {{ isEditMode ? '编辑账单' : '记一笔' }}
          </div>
        <a-button type="text" class="close-btn" @click="openDrawer = false">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter">
      <div class="form-panel">
        <a-form layout="vertical" :model="formState" class="drawer-form">
          <div class="middleBox pb2 bg-white">
            <div>
              <div class="lebal text-#666">
                <span
                  class="text-#f03 mr1"
                  style="font-family: SimSun, sans-serif"
                >*</span>账单金额：
              </div>
              <div
                class="payPrice h-20 border border-b-#eee border-solid border-x-none border-t-none"
              >
                <a-input-number
                  v-model:value="formState.amount"
                  :bordered="false"
                  :controls="false"
                  class="h-100% w-100% text-12"
                  :min="0"
                  :max="100000"
                  :precision="2"
                  placeholder="输入金额"
                >
                  <template #addonBefore>
                    <span class="text-12">¥</span>
                  </template>
                </a-input-number>
              </div>
            </div>
            <div class="mt6">
              <div class="flex justify-between w-75%">
                <a-form-item class="drawer-form-item">
                  <template #label>
                    <span><span
                      class="text-#f03 mr1"
                      style="font-family: SimSun, sans-serif"
                    >*</span>支付日期：</span>
                  </template>
                  <a-date-picker
                    v-model:value="formState.payDate"
                    class="w-60"
                    :allow-clear="false"
                    :disabled-date="disabledDate"
                    format="YYYY-MM-DD"
                    value-format="YYYY-MM-DD"
                  />
                </a-form-item>
                <a-form-item class="drawer-form-item">
                  <template #label>
                    <span><span
                      class="text-#f03 mr1"
                      style="font-family: SimSun, sans-serif"
                    >*</span>经办人：</span>
                  </template>
                  <StaffSelect
                    v-model="formState.dealStaffId"
                    placeholder="请选择经办人"
                    width="240px"
                  />
                </a-form-item>
              </div>
            </div>
          </div>
        </a-form>
      </div>
    </div>
    <div class="p-24px">
      <!-- 收入、支出 -->
      <!-- a-radio-group -->
      <a-radio-group
        v-model:value="formState.ledgerType"
        button-style="solid"
        class="ledger-type-group"
      >
        <a-radio-button
          :value="1"
          class="ledger-type-option ledger-type-option--income w-175px text-center font-size-16px"
        >
          <span class="ledger-type-label">
            <span
              v-if="formState.ledgerType === 1"
              class="ledger-type-icon ledger-type-icon--income"
            >
              <CheckOutlined />
            </span>
            <span class="ledger-type-text">收入</span>
          </span>
        </a-radio-button>
        <a-radio-button
          :value="2"
          class="ledger-type-option ledger-type-option--expense w-175px text-center font-size-16px"
        >
          <span class="ledger-type-label">
            <span
              v-if="formState.ledgerType === 2"
              class="ledger-type-icon ledger-type-icon--expense"
            >
              <CheckOutlined />
            </span>
            <span class="ledger-type-text">支出</span>
          </span>
        </a-radio-button>
      </a-radio-group>

      <div class="ledger-category-entry">
        <div v-if="formState.ledgerType === 1" class="ledger-inline-panel">
          <div class="ledger-expense-tabs">
            <button
              type="button"
              class="ledger-expense-trigger ledger-expense-trigger--income-active"
            >
              其他业务
            </button>
          </div>
          <div class="ledger-category-surface">
            <div class="ledger-category-grid">
              <button
                v-for="item in incomeCategoryItems"
                :key="item.key"
                class="ledger-category-item ledger-category-item--income"
                :class="{
                  'ledger-category-item--active':
                    activeLedgerItemKey === item.key,
                }"
                type="button"
                @click="selectLedgerItem(item.key)"
              >
                <div class="ledger-category-item__icon">
                  <component :is="item.icon" />
                </div>
                <span class="ledger-category-item__label">{{
                  item.label
                }}</span>
              </button>
            </div>
          </div>
        </div>

        <div v-else class="ledger-inline-panel">
          <div class="ledger-expense-tabs">
            <button
              v-for="group in expenseCategoryGroups"
              :key="group.key"
              type="button"
              class="ledger-expense-trigger"
              :class="{
                'ledger-expense-trigger--active':
                  activeExpenseCategory === group.key,
              }"
              @click="selectExpenseCategory(group.key)"
            >
              {{ group.label }}
            </button>
          </div>
          <div class="ledger-category-surface">
            <div class="ledger-category-grid">
              <button
                v-for="item in activeExpenseItems"
                :key="item.key"
                class="ledger-category-item ledger-category-item--expense"
                :class="{
                  'ledger-category-item--active':
                    activeLedgerItemKey === item.key,
                }"
                type="button"
                @click="selectLedgerItem(item.key)"
              >
                <div class="ledger-category-item__icon">
                  <component :is="item.icon" />
                </div>
                <span class="ledger-category-item__label">{{
                  item.label
                }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
    <!-- 收款方式区域 -->
    <div class="bg-white px-24px pb-24px pt-24px">
      <div class="payList">
        <div class="payList-title">
          <span>*</span>收款方式
          <span class="payList-tip ml-2">请选择</span>
        </div>
        <div class="pay mt-3">
          <a-radio-group
            v-model:value="formState.payMethod"
            class="custom-radio w-full"
          >
            <a-space :size="16" class="flex-wrap">
              <label
                v-for="item in checkOptions"
                :key="item.id"
                class="pay-box"
                :class="{ active: formState.payMethod === item.id }"
              >
                <span>
                  <img :src="item.img" alt="">
                  {{ item.label }}
                </span>
                <a-radio :value="item.id" />
              </label>
            </a-space>
          </a-radio-group>
        </div>
      </div>
      <div class="bg-white pt4 pb0">
        <a-form layout="vertical" :model="formState" class="flex flex-col">
          <a-form-item class="w-60">
            <template #label>
              <span><span
                class="text-#f03 mr1"
                style="font-family: SimSun, sans-serif"
              >*</span>支付账户</span>
            </template>
            <a-select
              v-model:value="formState.payAccount"
              :allow-clear="false"
              placeholder="请选择支付账户"
              :options="accountList"
            />
          </a-form-item>
          <a-form-item label="账单备注（选填）">
            <a-textarea
              v-model:value="formState.billRemarks"
              placeholder="请输入内容，最多100字"
              :auto-size="{ minRows: 2, maxRows: 5 }"
            />
          </a-form-item>
        </a-form>
      </div>
      <div class="upload bg-white pt0 mt--4">
        <a-upload
          v-model:file-list="fileList"
          list-type="picture-card"
          :custom-request="handleLedgerImageUpload"
          :before-upload="beforeUpload"
          accept=".jpg,.jpeg,.png,.bmp"
          @preview="handlePreview"
        >
          <div v-if="fileList.length < 3">
            <PlusOutlined class="text-6" />
          </div>
        </a-upload>
        <span class="text-#888">最多上传 3 张图片，支持 BMP / JPG / JPEG / PNG，单张图片不超过 4 MB</span>
      </div>
    </div>
    <template #footer>
      <div class="flex justify-end">
        <a-button
          type="primary"
          class="w-140px h-48px font-size-18px font-weight-600"
          :loading="submitting"
          @click="handleSubmit"
        >
          {{ isEditMode ? '保存' : '完成' }}
        </a-button>
      </div>
    </template>
    <a-modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="handleCancelImg">
      <img alt="example" style="width: 100%" :src="previewImage">
    </a-modal>
  </a-drawer>
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

.contenter {
  background: #fff;
  padding: 24px 24px 0 24px;
  .form-panel {
    background: #fff;
  }
}

.middleBox {
  max-width: 760px;
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.ledger-type-group {
  :deep(.ant-radio-button-wrapper) {
    text-align: center;
  }

  :deep(
      .ant-radio-button-wrapper-checked:not(.ant-radio-button-wrapper-disabled)
    ) {
    color: #fff;
  }

  :deep(
      .ledger-type-option--income.ant-radio-button-wrapper-checked:not(
          .ant-radio-button-wrapper-disabled
        )
    ) {
    background: #1677ff;
    border-color: #1677ff;
  }

  :deep(
      .ledger-type-option--expense.ant-radio-button-wrapper-checked:not(
          .ant-radio-button-wrapper-disabled
        )
    ) {
    background: #f90;
    border-color: #f90;
  }
}

.ledger-category-entry {
  padding-top: 16px;
}

.ledger-inline-panel {
  width: 100%;
}

.ledger-expense-tabs {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
}

.ledger-expense-trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  position: relative;
  min-width: 88px;
  height: 28px;
  padding: 0 16px;
  border: none;
  border-radius: 999px;
  background: #fff;
  color: #222;
  cursor: pointer;
  line-height: 28px;
}

.ledger-expense-trigger--active {
  background: #fff4e6;
  color: #f90;
}

.ledger-expense-trigger--income-active {
  background: #edf5ff;
  color: #1677ff;
}

.ledger-expense-trigger--income-active::after,
.ledger-expense-trigger--active::after {
  content: "";
  width: 0;
  height: 0;
  position: absolute;
  top: 34px;
  left: 31px;
  border-right: 10px solid #0000;
  border-left: 10px solid #0000;
  border-bottom: 10px solid #fff;
}

.ledger-category-surface {
  min-height: 112px;
  margin: 0 8px 0 0;
  border-radius: 16px;
  background: #fff;
  padding: 16px;
}

.ledger-category-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.ledger-category-item {
  display: flex;
  width: 72px;
  padding: 0;
  border: none;
  background: transparent;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  color: #222;
  cursor: pointer;
}

.ledger-category-item__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #f6f7fb;
  color: #d0d4dc;
  font-size: 20px;
}

.ledger-category-item--income:hover .ledger-category-item__icon,
.ledger-category-item--income.ledger-category-item--active
  .ledger-category-item__icon {
  background: #edf5ff;
  color: #1677ff;
}

.ledger-category-item--expense:hover .ledger-category-item__icon,
.ledger-category-item--expense.ledger-category-item--active
  .ledger-category-item__icon {
  background: #fff4e6;
  color: #f90;
}

.ledger-category-item__label {
  font-size: 14px;
  line-height: 20px;
  text-align: center;
  white-space: nowrap;
}

.payList {
  span {
    color: red;
  }

  span.payList-tip {
    color: var(--pro-ant-color-primary);
  }
}

.pay {
  margin-top: 10px;
  margin-bottom: 20px;

  .pay-box {
    display: flex;
    align-items: center;
    padding: 12px 16px;
    border: 1px solid #eee;
    border-radius: 6px;
    cursor: pointer;
    user-select: none;

    &:hover {
      border-color: var(--pro-ant-color-primary);
    }

    span {
      display: flex;
      align-items: center;
      margin-right: 20px;
      color: #000;

      img {
        width: 20px;
        height: 20px;
        margin-right: 6px;
      }
    }
  }

  .active {
    border-color: var(--pro-ant-color-primary);
  }
}

.custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
.custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
}

.custom-radio ::v-deep(.ant-radio-inner) {
  border-color: #d9d9d9;
  background-color: transparent;
}

.custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary);
  background-color: transparent;
}

.custom-radio ::v-deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary);
  transform: scale(0.5);
}

:deep(.ant-upload) {
  width: 78px !important;
  height: 78px !important;
}

:deep(.ant-upload-list-item-container) {
  width: 78px !important;
  height: 78px !important;
}

.ledger-type-label {
  position: relative;
  display: inline-block;
}

.ledger-type-text {
  display: inline-block;
}

.ledger-type-icon {
  position: absolute;
  top: 50%;
  right: calc(100% + 6px);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 999px;
  background: #fff;
  font-size: 10px;
  line-height: 1;
  transform: translateY(-50%);
}

.ledger-type-icon--income {
  color: #1677ff;
}

.ledger-type-icon--expense {
  color: #f90;
}

.payPrice {
  :deep(.ant-input-number) {
    width: 100%;
  }

  :deep(.ant-input-number-input) {
    height: 80px;
    line-height: 80px;
    font-family: "DIN alternate", sans-serif;
  }

  :deep(.ant-input-number-group-addon) {
    background: transparent;
    border: none;
  }
}

@media (max-width: 960px) {
  .middleBox {
    max-width: 100%;
  }
}
</style>
