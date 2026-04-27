<script setup lang="ts">
import { CloseOutlined, QuestionCircleFilled } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, createVNode, reactive, ref, watch } from 'vue'
import {
  getGroupClassFinishCoursePreviewApi,
  type GroupClassStudentPagedItem,
} from '@/api/edu-center/group-class'
import { addCloseTuitionAccountOrderApi } from '@/api/edu-center/tuition-account'
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

const emit = defineEmits(['update:open', 'success'])

const formRef = ref()
const loading = ref(false)
const submitLoading = ref(false)
const rows = ref<GroupClassStudentPagedItem[]>([])
let requestSeq = 0

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const formState = reactive({
  remark: '',
})

const classId = computed(() => String(props.record?.id || '').trim())
const className = computed(() => String(props.record?.name || '').trim() || '班级')
const closeDateText = computed(() => `${dayjs().format('YYYY-MM-DD')}（今天）`)
const closableRows = computed(() => rows.value.filter(canCloseCourse))

watch(
  () => props.open,
  (value) => {
    if (!value) {
      rows.value = []
      formState.remark = ''
      formRef.value?.resetFields?.()
      return
    }
    formState.remark = ''
    loadPreviewData()
  },
)

function closeFun() {
  formRef.value?.resetFields?.()
  formState.remark = ''
  openModal.value = false
}

function formatGender(value?: number) {
  if (value === 1)
    return '男'
  if (value === 0)
    return '女'
  return '未知'
}

function getStatusInfo(status?: number) {
  if (status === 2)
    return { text: '停课', className: 'text-#f90 bg-#fff5e6' }
  return { text: '在读', className: 'text-#06f bg-#e6f0ff' }
}

function formatNumber(value?: number) {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return '0'
  if (Number.isInteger(num))
    return String(num)
  return num.toFixed(2).replace(/\.?0+$/, '')
}

function formatMoney(value?: number) {
  return `¥ ${formatNumber(value)}`
}

function formatExpireDate(record: GroupClassStudentPagedItem) {
  if (!record.enableExpireTime)
    return '不限制'
  const value = record.expireTime || record.classStudentTuitionAccountInfo?.expireTime
  if (!value || `${value}`.startsWith('0001-01-01'))
    return '-'
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD') : '-'
}

function getQuantityUnit(mode?: number) {
  if (mode === 2)
    return '天'
  if (mode === 3 || mode === 4)
    return '元'
  return '课时'
}

function getRemainQuantity(record: GroupClassStudentPagedItem) {
  const remainQuantity = Number(record.quantity || 0)
  const remainFreeQuantity = Number(record.freeQuantity || 0)
  return remainQuantity + remainFreeQuantity
}

function getRemainQuantityText(record: GroupClassStudentPagedItem) {
  const unit = getQuantityUnit(record.classStudentTuitionAccountInfo?.lessonChargingMode)
  return `${formatNumber(getRemainQuantity(record))} ${unit}`.trim()
}

function resolveTuitionAccountId(record: GroupClassStudentPagedItem) {
  return String(record.classStudentTuitionAccountInfo?.tuitionAccountId || record.tuitionAccountId || '').trim()
}

function canCloseCourse(record: GroupClassStudentPagedItem) {
  const tuitionAccountId = resolveTuitionAccountId(record)
  if (!tuitionAccountId)
    return false
  return true
}

async function loadPreviewData() {
  if (!classId.value) {
    rows.value = []
    return
  }
  const seq = ++requestSeq
  loading.value = true
  try {
    const res = await getGroupClassFinishCoursePreviewApi({
      queryModel: {
        id: classId.value,
        classStudentStatus: [1],
      },
      sortModel: {
        orderByJoinTime: 0,
        totalTuition: 0,
        tuition: 0,
        confirmedTuition: 0,
        expireTime: 0,
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: 1000,
        pageIndex: 1,
        skipCount: 0,
      },
    })
    if (seq !== requestSeq)
      return
    if (res.code !== 200)
      throw new Error(res.message || '加载结课数据失败')
    const list = Array.isArray(res.result?.list) ? res.result.list : []
    rows.value = list.filter(canCloseCourse)
  }
  catch (error: any) {
    if (seq !== requestSeq)
      return
    rows.value = []
    messageService.error(error?.message || '加载结课数据失败')
  }
  finally {
    if (seq === requestSeq)
      loading.value = false
  }
}

async function handleSubmit() {
  if (!closableRows.value.length) {
    messageService.warning('当前班级无可结课的课程账户')
    return
  }
  Modal.confirm({
    title: '确认结课',
    centered: true,
    closable: false,
    maskClosable: false,
    keyboard: false,
    icon: createVNode(QuestionCircleFilled, { style: { color: '#ff4d4f', fontSize: '20px' } }),
    content: '结课后，学员班内默认课程账户的剩余天数将全部扣除，机构获得相应课消收入，请谨慎操作！',
    okText: '确定结课',
    cancelText: '再想想',
    okButtonProps: { danger: true, ghost: true },
    wrapClassName: 'group-finish-course-confirm-modal',
    async onOk() {
      submitLoading.value = true
      try {
        for (const row of closableRows.value) {
          const res = await addCloseTuitionAccountOrderApi({
            tuitionAccountId: resolveTuitionAccountId(row),
            quantity: Number(row.quantity || 0),
            freeQuantity: Number(row.freeQuantity || 0),
            tuition: Number(row.tuition || 0),
            remark: formState.remark.trim(),
          })
          if (res.code !== 200)
            throw new Error(`${row.name || '学员'}结课失败：${res.message || '结课失败'}`)
        }
        messageService.success('结课成功')
        emit('success')
        closeFun()
      }
      catch (error: any) {
        messageService.error(error?.message || '结课失败')
        return Promise.reject(error)
      }
      finally {
        submitLoading.value = false
      }
    },
  })
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    style="top:12px"
    class="modal-content-box"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="1000"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>结课</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
      <div class="finish-course-banner">
        结课后，学员报读课程的剩余课时将全部扣除，机构获得相应课消收入，请谨慎操作！
      </div>

      <div class="finish-course-title">
        {{ className }}
      </div>
      <div class="finish-course-count">
        共 {{ closableRows.length }} 个学员
      </div>

      <a-table
        :data-source="closableRows"
        :loading="loading"
        :pagination="false"
        :scroll="{ y: 280 }"
        row-key="id"
        size="small"
        class="finish-course-table"
      >
        <a-table-column title="学员/性别" key="name" data-index="name" width="180">
          <template #default="{ record }">
            <div class="flex items-center gap-12px">
              <a-avatar :src="record.avatar || undefined" :size="36">
                {{ record.name?.slice(0, 1) || '学' }}
              </a-avatar>
              <div>
                <div class="text-#222 leading-20px">
                  {{ record.name || '-' }}
                </div>
                <div class="text-#888 text-12px leading-18px">
                  {{ formatGender(record.sex) }}
                </div>
              </div>
            </div>
          </template>
        </a-table-column>
        <a-table-column title="在班状态" key="status" data-index="status" width="120">
          <template #default="{ record }">
            <span class="rounded-10 px-8px py-2px text-12px" :class="getStatusInfo(record.status).className">
              {{ getStatusInfo(record.status).text }}
            </span>
          </template>
        </a-table-column>
        <a-table-column title="班内默认课账户" key="account" width="180">
          <template #default="{ record }">
            {{ record.classStudentTuitionAccountInfo?.productName || '-' }}
          </template>
        </a-table-column>
        <a-table-column title="剩余数量" key="quantity" width="120">
          <template #default="{ record }">
            {{ getRemainQuantityText(record) }}
          </template>
        </a-table-column>
        <a-table-column title="剩余学费" key="tuition" width="120">
          <template #default="{ record }">
            {{ formatMoney(record.tuition) }}
          </template>
        </a-table-column>
        <a-table-column title="学费欠费金额" key="arrearTuition" width="140">
          <template #default="{ record }">
            {{ formatMoney(record.classStudentTuitionAccountInfo?.arrearTuition) }}
          </template>
        </a-table-column>
        <a-table-column title="有效日期" key="expireTime" width="140">
          <template #default="{ record }">
            {{ formatExpireDate(record) }}
          </template>
        </a-table-column>
      </a-table>

      <div class="finish-course-form-box">
        <a-form ref="formRef" layout="vertical" :model="formState">
          <div class="finish-course-form-row">
            <span class="finish-course-form-row__label">结课日期:</span>
            <span class="finish-course-form-row__value finish-course-form-row__value--danger">{{ closeDateText }}</span>
          </div>
          <div class="finish-course-form-row finish-course-form-row--remark">
            <span class="finish-course-form-row__label">结课备注:</span>
            <a-input v-model:value="formState.remark" placeholder="选填" />
          </div>
        </a-form>
      </div>
    </div>
    <template #footer>
      <a-button @click="closeFun">
        取消
      </a-button>
      <a-button type="primary" :loading="submitLoading" :disabled="loading || !closableRows.length" @click="handleSubmit">
        确认
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
  padding-bottom: 16px;
  max-height: calc(100vh - 140px);
  overflow-y: auto;
}

.finish-course-banner {
  padding: 12px 16px;
  color: #fa8c16;
  background: #fff5e6;
}

.finish-course-title {
  padding: 18px 24px 8px;
  color: #222;
  font-size: 18px;
  font-weight: 600;
  line-height: 28px;
}

.finish-course-count {
  position: relative;
  padding-left: 12px;
  margin: 0 24px 14px;
  color: #222;
  font-size: 16px;
  line-height: 24px;

  &::before {
    position: absolute;
    top: 4px;
    left: 0;
    width: 4px;
    height: 16px;
    border-radius: 2px;
    background: var(--pro-ant-color-primary);
    content: '';
  }
}

.finish-course-table {
  margin: 0 24px;

  :deep(.ant-table-thead > tr > th) {
    color: #222;
    font-weight: 500;
    background: #fafafa;
  }
}

.finish-course-form-box {
  padding: 16px 24px 0;
}

.finish-course-form-box :deep(.ant-form) {
  padding: 20px 24px;
  border-radius: 16px;
  background: #f7f8ff;
}

.finish-course-form-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.finish-course-form-row--remark {
  margin-top: 18px;
  align-items: flex-start;
}

.finish-course-form-row__label {
  width: 72px;
  color: #222;
  line-height: 32px;
  flex: none;
}

.finish-course-form-row__value {
  line-height: 32px;
}

.finish-course-form-row__value--danger {
  color: #ff4d4f;
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

.group-finish-course-confirm-modal .ant-modal-content {
  border-radius: 20px;
  overflow: hidden;
}

.group-finish-course-confirm-modal .ant-modal-confirm-body-wrapper {
  padding: 6px 8px 8px;
}

.group-finish-course-confirm-modal .ant-modal-confirm-title {
  font-size: 30px;
  font-weight: 600;
  line-height: 42px;
}

.group-finish-course-confirm-modal .ant-modal-confirm-content {
  margin-top: 14px !important;
  margin-left: 34px !important;
  color: #666;
  font-size: 18px;
  line-height: 34px;
}

.group-finish-course-confirm-modal .ant-modal-confirm-btns {
  margin-top: 28px !important;
}
</style>
