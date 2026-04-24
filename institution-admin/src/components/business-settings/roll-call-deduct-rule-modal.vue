<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'
import { computed, reactive, ref, watch } from 'vue'
import { type InstConfig, setInstConfigModuleApi } from '~@/api/common/config'
import {
  getCourseRollCallDeductRulePageApi,
  updateCourseRollCallDeductPriceApi,
} from '~@/api/edu-center/course-list'
import messageService from '~@/utils/messageService'

interface CourseDeductRuleRow {
  id: number
  name: string
  rollCallDeductPrice: number | null
}

const props = defineProps<{
  open: boolean
  instConfig?: Partial<InstConfig>
}>()

const emit = defineEmits<{
  'update:open': [value: boolean]
  saved: []
}>()

const loading = ref(false)
const submitting = ref(false)
const rowSavingId = ref<number | null>(null)
const rows = ref<CourseDeductRuleRow[]>([])
const defaultPrice = ref<number>(100)

const editingState = reactive({
  courseId: 0,
  price: null as number | null,
})

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const singleCourseCount = computed(() => rows.value.filter(item => item.rollCallDeductPrice != null).length)

const tableColumns = [
  {
    title: '课程名称',
    dataIndex: 'name',
    key: 'name',
    width: 220,
  },
  {
    title: '扣费方式',
    key: 'mode',
    width: 150,
  },
  {
    title: '扣费金额',
    key: 'price',
    width: 180,
  },
  {
    title: '操作',
    key: 'action',
    width: 160,
  },
]

function extractPagedItems(res: any): CourseDeductRuleRow[] {
  const list = res?.result?.items ?? res?.result ?? res?.data?.items ?? res?.data ?? []
  return Array.isArray(list) ? list : []
}

function normalizePrice(value: unknown, fallback: number) {
  if (value == null)
    return fallback
  if (typeof value === 'string' && value.trim() === '')
    return fallback
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed < 0)
    return fallback
  return Number(parsed.toFixed(2))
}

function normalizeRow(record: any): CourseDeductRuleRow {
  return {
    id: Number(record?.id || 0),
    name: String(record?.name || ''),
    rollCallDeductPrice: record?.rollCallDeductPrice == null || record?.rollCallDeductPrice === ''
      ? null
      : Number(record.rollCallDeductPrice),
  }
}

function getDisplayPrice(record: any) {
  const row = normalizeRow(record)
  return row.rollCallDeductPrice != null ? Number(row.rollCallDeductPrice) : Number(defaultPrice.value || 0)
}

function closeFun() {
  if (submitting.value || rowSavingId.value)
    return
  openModal.value = false
}

function startEdit(record: any) {
  const row = normalizeRow(record)
  editingState.courseId = row.id
  editingState.price = row.rollCallDeductPrice != null ? Number(row.rollCallDeductPrice) : null
}

function cancelEdit() {
  editingState.courseId = 0
  editingState.price = null
}

async function loadCourseRules() {
  loading.value = true
  try {
    const res = await getCourseRollCallDeductRulePageApi({
      pageRequestModel: {
        needTotal: true,
        pageIndex: 1,
        pageSize: 200,
      },
      sortModel: {
        byUpdateTime: -1,
        byTotalSales: 0,
      },
      queryModel: {
        chargeTypes: [3],
        delFlag: false,
      },
    })
    rows.value = extractPagedItems(res).map((item: any) => ({
      id: Number(item?.id || 0),
      name: String(item?.name || ''),
      rollCallDeductPrice: item?.rollCallDeductPrice == null || item?.rollCallDeductPrice === ''
        ? null
        : Number(item.rollCallDeductPrice),
    })).filter(item => item.id > 0)
  }
  catch (error) {
    console.error('load course deduct rules failed', error)
    messageService.error('获取课程扣费规则失败')
  }
  finally {
    loading.value = false
  }
}

async function handleSaveDefaultPrice() {
  if (!Number.isFinite(Number(defaultPrice.value)) || Number(defaultPrice.value) < 0) {
    messageService.error('默认扣费金额不能小于 0')
    return false
  }

  await setInstConfigModuleApi('course', {
    chargeByPriceDefaultPrice: String(normalizePrice(defaultPrice.value, 100)),
  })
  return true
}

async function handleSubmit() {
  if (submitting.value)
    return
  submitting.value = true
  try {
    await handleSaveDefaultPrice()
    messageService.success('扣费规则已更新')
    emit('saved')
    openModal.value = false
  }
  catch (error) {
    console.error('save deduct rule modal failed', error)
    messageService.error('保存失败，请稍后重试')
  }
  finally {
    submitting.value = false
  }
}

async function saveRow(record: any) {
  const row = normalizeRow(record)
  if (rowSavingId.value)
    return
  if (editingState.courseId !== row.id)
    return
  if (editingState.price != null && Number(editingState.price) < 0) {
    messageService.error('扣费金额不能小于 0')
    return
  }

  rowSavingId.value = row.id
  try {
    await updateCourseRollCallDeductPriceApi({
      courseId: row.id,
      rollCallDeductPrice: editingState.price == null ? null : normalizePrice(editingState.price, 0),
    })
    await loadCourseRules()
    cancelEdit()
    emit('saved')
    messageService.success('课程扣费金额已更新')
  }
  catch (error) {
    console.error('save course deduct rule failed', error)
    messageService.error('保存失败，请稍后重试')
  }
  finally {
    rowSavingId.value = null
  }
}

watch(() => props.open, async (value) => {
  if (!value)
    return
  defaultPrice.value = normalizePrice(props.instConfig?.chargeByPriceDefaultPrice, 100)
  cancelEdit()
  await loadCourseRules()
}, { immediate: true })
</script>

<template>
  <a-modal
    v-model:open="openModal"
    class="modal-content-box"
    wrap-class-name="roll-call-deduct-rule-modal-wrap"
    centered
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="780"
    destroy-on-close
    :confirm-loading="submitting"
    :body-style="{ maxHeight: '68vh', overflowY: 'auto', padding: '0' }"
  >
    <template #title>
      <div class="modal-title-bar">
        <span>编辑扣费规则</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="deduct-rule-modal__body">
      <div class="deduct-rule-modal__tip">
        <div class="deduct-rule-modal__tip-line">
          <span class="deduct-rule-modal__tip-q">Q：</span>
          <span>课程扣费规则的作用是什么？</span>
        </div>
        <div class="deduct-rule-modal__tip-line">
          <span class="deduct-rule-modal__tip-q">A：</span>
          <span>作为学员报名按金额收费课程后的课消扣费依据，未设置单课扣费时将按默认扣费金额执行。</span>
        </div>
      </div>

      <div class="deduct-rule-modal__default-row">
        <span class="deduct-rule-modal__default-label">默认扣费：</span>
        <a-input-number
          v-model:value="defaultPrice"
          :min="0"
          :precision="2"
          :controls="false"
          style="width: 160px"
        />
        <span class="deduct-rule-modal__default-unit">元</span>
        <span class="deduct-rule-modal__default-tip">（仅对未设置单课扣费的课程有效）</span>
      </div>

      <div class="deduct-rule-modal__desc">
        如需修改单个课程的扣费金额，可直接在下方列表中逐条编辑。当前已设置单课扣费 <span class="text-primary">{{ singleCourseCount }}</span> 门课程。
      </div>

      <a-table
        :columns="tableColumns"
        :data-source="rows"
        :loading="loading"
        :pagination="false"
        :scroll="{ y: 420 }"
        row-key="id"
        class="deduct-rule-modal__table"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'mode'">
            {{ record.rollCallDeductPrice != null ? '单课扣费' : '默认扣费' }}
          </template>

          <template v-else-if="column.key === 'price'">
            <template v-if="editingState.courseId === record.id">
              <a-input-number
                v-model:value="editingState.price"
                :min="0"
                :precision="2"
                :controls="false"
                style="width: 140px"
                placeholder="留空则恢复默认"
              />
            </template>
            <template v-else>
              {{ getDisplayPrice(record) }}
            </template>
          </template>

          <template v-else-if="column.key === 'action'">
            <template v-if="editingState.courseId === record.id">
              <a-button type="link" class="settings-link" :loading="rowSavingId === record.id" @click="saveRow(record)">
                保存
              </a-button>
              <a-button type="link" class="settings-link settings-link--muted" @click="cancelEdit">
                取消
              </a-button>
            </template>
            <template v-else>
              <a-button type="link" class="settings-link" @click="startEdit(record)">
                编辑
              </a-button>
            </template>
          </template>
        </template>
      </a-table>
    </div>

    <template #footer>
      <a-button danger ghost @click="closeFun">
        关闭
      </a-button>
      <a-button type="primary" ghost :loading="submitting" @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
</template>

<style lang="less" scoped>
.modal-title-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 20px;
  font-weight: 600;
  color: #1f2329;
}

.close-btn {
  &:hover {
    background: transparent;
  }
}

.close-icon {
  font-size: 18px;
  color: #909399;
}

.deduct-rule-modal__body {
  padding: 20px 24px 8px;
  background: #fff;
}

.deduct-rule-modal__tip {
  padding: 14px 16px;
  margin-bottom: 18px;
  background: #fff7e8;
  border: 1px solid #f6c88f;
  border-radius: 12px;
}

.deduct-rule-modal__tip-line {
  display: flex;
  gap: 6px;
  line-height: 24px;
  color: #5c3b16;
}

.deduct-rule-modal__tip-q {
  font-weight: 600;
}

.deduct-rule-modal__default-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 14px;
}

.deduct-rule-modal__default-label {
  font-weight: 600;
  color: #1f2329;
}

.deduct-rule-modal__default-unit {
  color: #1f2329;
}

.deduct-rule-modal__default-tip {
  color: #86909c;
}

.deduct-rule-modal__desc {
  margin-bottom: 16px;
  color: #4e5969;
}

.deduct-rule-modal__table {
  :deep(.ant-table-thead > tr > th) {
    color: #1f2329;
    font-weight: 600;
    background: #fafafa;
  }

  :deep(.ant-table-tbody > tr > td) {
    vertical-align: middle;
  }
}

.settings-link {
  padding: 0;
}

.settings-link--muted {
  color: #86909c;
}

.text-primary {
  color: #1677ff;
}
</style>

<style>
.roll-call-deduct-rule-modal-wrap .ant-modal-content {
  overflow: hidden;
  border-radius: 16px;
}

.roll-call-deduct-rule-modal-wrap .ant-modal-header {
  padding: 14px 20px !important;
  margin-bottom: 0;
  border-bottom: 1px solid #f0f0f0;
}

.roll-call-deduct-rule-modal-wrap .ant-modal-body {
  padding: 0 !important;
}

.roll-call-deduct-rule-modal-wrap .ant-modal-footer {
  padding: 12px 20px 16px;
  border-top: 1px solid #f0f0f0;
}
</style>
