<script setup>
import { computed, ref, watch } from 'vue'
import { CloseOutlined, EditOutlined, InfoCircleOutlined } from '@ant-design/icons-vue'
import { getRechargeAccountDetailPageApi, getStudentDetailApi, updateRechargeAccountApi } from '@/api/finance-center/recharge-account'
import { useStudentStore } from '@/stores/student'
import messageService from '@/utils/messageService'
import AccountRefundDrawer from './account-refund-drawer.vue'

const DEFAULT_AVATAR_URL = 'https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png'

const FLOW_TYPE_OPTIONS = [
  { id: 1, value: '储值账户充值' },
  { id: 2, value: '储值账户退费' },
  { id: 3, value: '报名订单支出' },
  { id: 4, value: '退费订单退回' },
  { id: 5, value: '作废储值充值' },
  { id: 6, value: '转课少补支出' },
  { id: 7, value: '转课退回' },
  { id: 8, value: '作废转课少补支出' },
  { id: 9, value: '作废转课退回' },
  { id: 10, value: '场地预约支出' },
  { id: 11, value: '场地预约退回' },
  { id: 12, value: '作废储值退费' },
]

const FLOW_TYPE_LABEL_MAP = FLOW_TYPE_OPTIONS.reduce((acc, item) => {
  acc[item.id] = item.value
  return acc
}, {})

const SUMMARY_FIELD_TIP_MAP = {
  rechargeIncome: '累计充值包括充值金额+退回金额',
  givingIncome: '累计赠送包括充值赠送金额+退回赠送金额',
  residualIncome: '累计残联包括充值残联金额+退回残联金额',
}

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  account: {
    type: Object,
    default: () => ({}),
  },
})

const emit = defineEmits(['update:open', 'updated', 'open-linked-order-detail'])

const drawerOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const accountSnapshot = ref({
  rechargeAccountId: '',
  rechargeAccountName: '',
  phone: '',
  balanceTotal: 0,
  rechargeBalance: 0,
  residualBalance: 0,
  givingBalance: 0,
  rechargeAccountStudents: [],
})

const loading = ref(false)
const summaryLoading = ref(false)
const dataSource = ref([])
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showQuickJumper: true,
  showTotal: total => `共 ${total} 条`,
})

const accountFlowSummary = ref({
  rechargeIncome: 0,
  rechargeExpend: 0,
  givingIncome: 0,
  givingExpend: 0,
  residualIncome: 0,
  residualExpend: 0,
})
const studentProfileMap = ref({})

const editModalOpen = ref(false)
const editAccountName = ref('')
const saveNameLoading = ref(false)

const openStudentDrawer = ref(false)
const openAccountRechargeDrawer = ref(false)
const openAccountRefundDrawer = ref(false)
const rechargeStudentId = ref(undefined)
const refundAccountId = ref(undefined)
const studentStore = useStudentStore()

const accountId = computed(() => String(accountSnapshot.value.rechargeAccountId || ''))
const accountName = computed(() => accountSnapshot.value.rechargeAccountName || '默认账户')

const accountColumns = [
  { title: '操作时间', dataIndex: 'createTime', key: 'createTime', width: 176, fixed: 'left' },
  { title: '明细类型', dataIndex: 'rechargeAccountFlowSourceType', key: 'rechargeAccountFlowSourceType', width: 160 },
  { title: '金额', dataIndex: 'amount', key: 'amount', width: 140 },
  { title: '赠送金额', dataIndex: 'givingAmount', key: 'givingAmount', width: 140 },
  { title: '残联金额', dataIndex: 'residualAmount', key: 'residualAmount', width: 140 },
  { title: '订单编号', dataIndex: 'sourceOrderNumber', key: 'sourceOrderNumber', width: 260 },
  { title: '明细关联学员', dataIndex: 'studentName', key: 'studentName', width: 260 },
  { title: '账单备注', dataIndex: 'remark', key: 'remark', width: 180 },
  { title: '总计', dataIndex: 'totalAmount', key: 'totalAmount', width: 160, fixed: 'right', align: 'right' },
]

const summaryItems = computed(() => [
  { key: 'availableTotal', label: '可用总余额（元）', value: formatSummaryMoney(accountSnapshot.value.balanceTotal), highlight: true },
  { key: 'rechargeBalance', label: '充值余额（元）', value: formatSummaryMoney(accountSnapshot.value.rechargeBalance) },
  { key: 'rechargeIncome', label: '累计充值（元）', value: formatSummaryMoney(accountFlowSummary.value.rechargeIncome) },
  { key: 'rechargeExpend', label: '累计充值消费（元）', value: formatSummaryMoney(accountFlowSummary.value.rechargeExpend) },
  { key: 'givingBalance', label: '赠送余额（元）', value: formatSummaryMoney(accountSnapshot.value.givingBalance) },
  { key: 'givingIncome', label: '累计赠送（元）', value: formatSummaryMoney(accountFlowSummary.value.givingIncome) },
  { key: 'givingExpend', label: '累计赠送消费（元）', value: formatSummaryMoney(accountFlowSummary.value.givingExpend) },
  { key: 'residualBalance', label: '残联余额（元）', value: formatSummaryMoney(accountSnapshot.value.residualBalance) },
  { key: 'residualIncome', label: '累计残联（元）', value: formatSummaryMoney(accountFlowSummary.value.residualIncome) },
  { key: 'residualExpend', label: '累计残联消费（元）', value: formatSummaryMoney(accountFlowSummary.value.residualExpend) },
])

const mainStudentId = computed(() => {
  const students = accountSnapshot.value.rechargeAccountStudents || []
  const mainStudent = students.find(item => item.isMainStudent) || students[0]
  return mainStudent?.studentId
})

watch(
  () => [props.open, props.account],
  ([open]) => {
    if (!open)
      return
    syncAccountSnapshot()
    hydrateStudentProfiles()
    pagination.value.current = 1
    fetchDetailData()
    fetchAccountFlowSummary()
  },
  { deep: true },
)

function syncAccountSnapshot() {
  const source = props.account || {}
  accountSnapshot.value = {
    rechargeAccountId: source.rechargeAccountId || '',
    rechargeAccountName: source.rechargeAccountName || '',
    phone: source.phone || '',
    balanceTotal: Number(source.balanceTotal || 0),
    rechargeBalance: Number(source.rechargeBalance || 0),
    residualBalance: Number(source.residualBalance || 0),
    givingBalance: Number(source.givingBalance || 0),
    rechargeAccountStudents: Array.isArray(source.rechargeAccountStudents) ? source.rechargeAccountStudents : [],
  }
}

async function hydrateStudentProfiles() {
  const students = Array.isArray(accountSnapshot.value.rechargeAccountStudents) ? accountSnapshot.value.rechargeAccountStudents : []
  const ids = students.map(item => String(item.studentId || '')).filter(Boolean)
  if (!ids.length) {
    studentProfileMap.value = {}
    return
  }
  const entries = await Promise.all(ids.map(async (studentId) => {
    try {
      const { result } = await getStudentDetailApi({ studentId })
      return [
        studentId,
        {
          avatar: result?.avatar || '',
          phone: result?.phone || '',
        },
      ]
    }
    catch (error) {
      return [studentId, { avatar: '', phone: '' }]
    }
  }))
  studentProfileMap.value = Object.fromEntries(entries)
}

function formatDate(dateStr) {
  if (!dateStr)
    return '-'
  return String(dateStr).replace('T', ' ').slice(0, 16)
}

function formatSummaryMoney(value) {
  return Number(value || 0).toFixed(2)
}

function formatTableMoney(value) {
  const num = Number(value || 0)
  if (!num)
    return '¥0.00'
  if (num < 0)
    return `-¥${Math.abs(num).toFixed(2)}`
  return `¥${num.toFixed(2)}`
}

function accumulateSummaryFromRows(list = []) {
  const summary = {
    rechargeIncome: 0,
    rechargeExpend: 0,
    givingIncome: 0,
    givingExpend: 0,
    residualIncome: 0,
    residualExpend: 0,
  }

  list.forEach((item) => {
    const amount = Number(item.amount || 0)
    const givingAmount = Number(item.givingAmount || 0)
    const residualAmount = Number(item.residualAmount || 0)

    if (amount >= 0)
      summary.rechargeIncome += amount
    else
      summary.rechargeExpend += Math.abs(amount)

    if (givingAmount >= 0)
      summary.givingIncome += givingAmount
    else
      summary.givingExpend += Math.abs(givingAmount)

    if (residualAmount >= 0)
      summary.residualIncome += residualAmount
    else
      summary.residualExpend += Math.abs(residualAmount)
  })

  accountFlowSummary.value = summary
}

async function fetchDetailData() {
  if (!accountId.value)
    return
  try {
    loading.value = true
    const { result } = await getRechargeAccountDetailPageApi({
      queryModel: {
        rechargeAccountId: accountId.value,
      },
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: (pagination.value.current - 1) * pagination.value.pageSize,
      },
      sortModel: {
        orderByCreatedTime: 0,
      },
    })

    const list = Array.isArray(result?.list) ? result.list : []
    dataSource.value = list
    pagination.value.total = Number(result?.total || 0)
  }
  catch (error) {
    console.error('获取储值账户详情失败:', error)
    messageService.error('获取储值账户详情失败')
  }
  finally {
    loading.value = false
  }
}

async function fetchAccountFlowSummary() {
  if (!accountId.value)
    return
  summaryLoading.value = true
  try {
    const pageSize = 200
    const maxRound = 30
    let pageIndex = 1
    let total = 0
    let loaded = 0
    const rows = []

    while (pageIndex <= maxRound) {
      const res = await getRechargeAccountDetailPageApi({
        queryModel: {
          rechargeAccountId: accountId.value,
        },
        pageRequestModel: {
          needTotal: true,
          pageSize,
          pageIndex,
          skipCount: (pageIndex - 1) * pageSize,
        },
        sortModel: {
          orderByCreatedTime: 0,
        },
      })

      const list = Array.isArray(res.result?.list) ? res.result.list : []
      total = Number(res.result?.total || 0)
      rows.push(...list)
      loaded += list.length

      if (!list.length || loaded >= total)
        break
      pageIndex += 1
    }

    accumulateSummaryFromRows(rows)
  }
  catch (error) {
    console.error('获取储值账户汇总失败:', error)
    accountFlowSummary.value = {
      rechargeIncome: 0,
      rechargeExpend: 0,
      givingIncome: 0,
      givingExpend: 0,
      residualIncome: 0,
      residualExpend: 0,
    }
  }
  finally {
    summaryLoading.value = false
  }
}

function handleTableChange(pag) {
  pagination.value.current = pag.current
  pagination.value.pageSize = pag.pageSize
  fetchDetailData()
}

function handleOpenEditName() {
  editAccountName.value = accountName.value
  editModalOpen.value = true
}

async function handleSaveAccountName() {
  const name = String(editAccountName.value || '').trim()
  if (!name) {
    messageService.error('请输入账户名称')
    return
  }
  if (!accountId.value) {
    messageService.error('缺少储值账户ID')
    return
  }

  try {
    saveNameLoading.value = true
    const res = await updateRechargeAccountApi({
      rechargeAccountId: accountId.value,
      rechargeAccountName: name,
    })

    if (res.code === 200) {
      accountSnapshot.value.rechargeAccountName = name
      messageService.success('账户名称修改成功')
      editModalOpen.value = false
      emit('updated')
      return
    }

    messageService.error(res.message || '账户名称修改失败')
  }
  catch (error) {
    console.error('账户名称修改失败:', error)
    const status = error?.response?.status
    const backendMessage = error?.response?.data?.message || error?.message
    if (status === 404) {
      messageService.error('当前环境未实现“修改储值账户”接口（404）')
      return
    }
    messageService.error(backendMessage || '账户名称修改失败')
  }
  finally {
    saveNameLoading.value = false
  }
}

function handleExport() {
  messageService.info('导出功能开发中')
}

function handleRefund() {
  if (!accountId.value) {
    messageService.error('缺少储值账户ID')
    return
  }
  refundAccountId.value = accountId.value
  openAccountRefundDrawer.value = true
}

function handleRecharge() {
  if (!mainStudentId.value) {
    messageService.error('该储值账户暂无关联学员，无法发起充值')
    return
  }
  rechargeStudentId.value = mainStudentId.value
  openAccountRechargeDrawer.value = true
}

function handleSeeStuData(studentId = '') {
  if (!studentId) {
    messageService.error('invalid studentId')
    return
  }
  studentStore.setStudentId(String(studentId))
  openStudentDrawer.value = true
}

function handleOrderDetail(orderId = '') {
  if (!orderId) {
    messageService.error('invalid orderId')
    return
  }
  emit('open-linked-order-detail', String(orderId))
}

function handleSubmitted() {
  emit('updated')
  fetchDetailData()
  fetchAccountFlowSummary()
}

function formatMobile(mobile = '') {
  const text = String(mobile || '')
  if (!text)
    return '-'
  if (text.length === 11)
    return text.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')
  return text
}

function getStudentAvatar(student = {}) {
  const studentId = String(student.studentId || '')
  return student.avatarUrl
    || student.avatar
    || student.studentAvatar
    || studentProfileMap.value?.[studentId]?.avatar
    || DEFAULT_AVATAR_URL
}

function getStudentPhone(student = {}) {
  const studentId = String(student.studentId || '')
  return formatMobile(
    student.studentPhone
    || student.phone
    || studentProfileMap.value?.[studentId]?.phone
    || accountSnapshot.value.phone,
  )
}

function handleEditRelation() {
  messageService.info('修改关联功能开发中')
}

function closeDrawer() {
  drawerOpen.value = false
}
</script>

<template>
  <a-drawer
    v-model:open="drawerOpen"
    :body-style="{ padding: '0', background: '#f5f7fb' }"
    :push="false"
    :closable="false"
    width="1160"
    placement="right"
    destroy-on-close
  >
    <template #title>
      <div class="custom-header flex justify-between h-4 flex-items-center">
        <div class="text-5">
          储值账户详情
        </div>
        <a-button type="text" class="close-btn" @click="closeDrawer">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="drawer-inner">
      <div class="head-card">
        <div class="account-title-wrap">
          <span class="account-label">储值账户：</span>
          <span class="account-name">{{ accountName }}</span>
          <EditOutlined class="edit-icon" @click="handleOpenEditName" />
        </div>
        <a-space :size="12">
          <a-button danger @click="handleRefund">
            退款
          </a-button>
          <a-button type="primary" @click="handleRecharge">
            充值
          </a-button>
        </a-space>
      </div>

      <div class="student-card">
        <div class="section-title">
          已关联学员：{{ (accountSnapshot.rechargeAccountStudents || []).length }}位
        </div>
        <div v-if="(accountSnapshot.rechargeAccountStudents || []).length" class="student-list">
          <div
            v-for="student in (accountSnapshot.rechargeAccountStudents || [])"
            :key="student.studentId"
            class="linked-student-card"
          >
            <span v-if="student.isMainStudent" class="main-student-badge">主学员</span>
            <a-tooltip placement="top">
              <template #title>
                查看学员档案
              </template>
              <div class="linked-student-main" @click="handleSeeStuData(student.studentId)">
                <div class="student-avatar-wrap">
                  <img
                    class="student-avatar-image"
                    :src="getStudentAvatar(student)"
                    :alt="student.studentName || '学员头像'"
                  >
                </div>
                <div class="student-name">
                  {{ student.studentName || '-' }}
                </div>
                <div class="student-phone">
                  {{ getStudentPhone(student) }}
                </div>
              </div>
            </a-tooltip>
            <a class="edit-relation-link" @click="handleEditRelation">修改关联</a>
          </div>
        </div>
        <div v-else class="text-#999">
          -
        </div>
      </div>

      <a-spin :spinning="summaryLoading">
        <div class="summary-scroll">
          <div class="summary-grid">
            <div
              v-for="item in summaryItems"
              :key="item.key"
              class="summary-item"
              :class="{
                'summary-item-highlight': item.highlight,
                'summary-item-fixed-left': item.key === 'availableTotal',
              }"
            >
              <div class="summary-label">
                <span>{{ item.label }}</span>
                <a-popover v-if="SUMMARY_FIELD_TIP_MAP[item.key]" placement="top" trigger="hover">
                  <template #title>
                    <div style="font-weight:600;color:#222;">
                      字段说明
                    </div>
                  </template>
                  <template #content>
                    <div style="color:#666;line-height:1.6;max-width:360px;">
                      {{ SUMMARY_FIELD_TIP_MAP[item.key] }}
                    </div>
                  </template>
                  <InfoCircleOutlined class="summary-tip-icon" />
                </a-popover>
              </div>
              <div class="summary-value" :class="{ 'summary-value-emphasis': item.key === 'availableTotal' }">
                {{ item.value }}
              </div>
            </div>
          </div>
        </div>
      </a-spin>

      <div class="detail-card">
        <div class="table-title">
          <div class="total">
            共 {{ pagination.total }} 个
          </div>
          <a-button @click="handleExport">
            导出数据
          </a-button>
        </div>
        <a-table
          :data-source="dataSource"
          :pagination="pagination"
          :columns="accountColumns"
          :scroll="{ x: 1640 }"
          :loading="loading"
          row-key="rechargeAccountFlowId"
          size="small"
          @change="handleTableChange"
        >
          <template #headerCell="{ column }">
            <template v-if="column.key === 'studentName'">
              <span class="mr-1">明细关联学员</span>
              <a-popover placement="top" trigger="hover">
                <template #content>
                  <div class="help-popover">
                    <div class="help-title">
                      明细关联学员
                    </div>
                    <div>创建储值账户明细流水时的关联学员。</div>
                    <div>如：储值充值/退费时，默认显示主要关联学员；发起报课消费、退课退回时的订单关联学员。</div>
                  </div>
                </template>
                <InfoCircleOutlined class="cursor-pointer text-#999 hover:text-#1677ff" />
              </a-popover>
            </template>
            <template v-if="column.key === 'rechargeAccountFlowSourceType'">
              <span class="mr-1">明细类型</span>
              <a-popover placement="top" trigger="hover">
                <template #content>
                    <div class="help-popover type-popover">
                      <div class="help-title">
                        明细类型
                      </div>
                    <div>储值账户充值：完成储值账户充值订单时</div>
                    <div>储值账户退费：完成储值账户退费订单时</div>
                    <div>报名订单支出：报名缴费订单使用储值支付时</div>
                    <div>退费订单退回：退课、退教材费、退学杂费的金额退回储值账户时</div>
                    <div>作废储值充值：作废储值账户充值订单时</div>
                    <div>转课少补支出：转课存在差价时，使用储值支付转课订单时</div>
                    <div>转课退回：转课存在差价时，多出金额退回储值账户时</div>
                    <div>作废转课少补支出：作废使用储值支付的转课订单时</div>
                    <div>作废转课退回：作废转课退回订单时</div>
                    <div>场地预约支出：场地预约订单使用储值账户时</div>
                    <div>场地预约退回：场地预约订单作废/订单关闭时</div>
                    <div>作废储值退费：作废储值退费订单时</div>
                  </div>
                </template>
                <InfoCircleOutlined class="cursor-pointer text-#999 hover:text-#1677ff" />
              </a-popover>
            </template>
            <template v-if="column.key === 'totalAmount'">
              <div class="flex justify-end">
                总计
              </div>
            </template>
          </template>

          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'createTime'">
              {{ formatDate(record.createTime) }}
            </template>
            <template v-if="column.key === 'rechargeAccountFlowSourceType'">
              <span
                class="flow-tag"
                :class="{ 'flow-tag--refund': Number(record.rechargeAccountFlowSourceType) === 2 }"
              >{{ FLOW_TYPE_LABEL_MAP[record.rechargeAccountFlowSourceType] || '-' }}</span>
            </template>
            <template v-if="column.key === 'amount'">
              {{ formatTableMoney(record.amount) }}
            </template>
            <template v-if="column.key === 'givingAmount'">
              {{ formatTableMoney(record.givingAmount) }}
            </template>
            <template v-if="column.key === 'residualAmount'">
              {{ formatTableMoney(record.residualAmount) }}
            </template>
            <template v-if="column.key === 'sourceOrderNumber'">
              <a
                v-if="record.sourceOrderNumber && Number(record.sourceId || 0) > 0"
                class="text-#06f cursor-pointer"
                @click="handleOrderDetail(record.sourceId)"
              >
                {{ record.sourceOrderNumber }}
              </a>
              <span v-else-if="record.sourceOrderNumber">{{ record.sourceOrderNumber }}</span>
              <span v-else>-</span>
            </template>
            <template v-if="column.key === 'studentName'">
              <student-avatar
                :id="record.studentId"
                :name="record.studentName || '-'"
                :phone="record.studentPhone || ''"
                :avatar-url="record.studentAvatar"
                :show-gender="false"
                :show-age="false"
                default-active-key="0"
              />
            </template>
            <template v-if="column.key === 'remark'">
              {{ record.remark || '-' }}
            </template>
            <template v-if="column.key === 'totalAmount'">
              <div class="flex justify-end font-800">
                {{ formatTableMoney(record.totalAmount) }}
              </div>
            </template>
          </template>
        </a-table>
      </div>
    </div>

    <a-modal
      v-model:open="editModalOpen"
      title="修改储值账户"
      width="500px"
      ok-text="确定"
      cancel-text="取消"
      :confirm-loading="saveNameLoading"
      @ok="handleSaveAccountName"
    >
      <a-input
        v-model:value="editAccountName"
        maxlength="30"
        show-count
        size="large"
        placeholder="请输入储值账户名称"
      />
    </a-modal>

    <student-info-drawer v-model:open="openStudentDrawer" />
    <account-recharge-drawer
      v-model:open="openAccountRechargeDrawer"
      :stu-id="rechargeStudentId"
      @submitted="handleSubmitted"
    />
    <account-refund-drawer
      v-model:open="openAccountRefundDrawer"
      :recharge-account-id="refundAccountId"
      @submitted="handleSubmitted"
    />
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

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.drawer-inner {
  min-height: 100%;
  padding: 0 20px 20px;
}

.head-card {
  align-items: center;
  background: #dce8fb;
  border-radius: 0;
  display: flex;
  justify-content: space-between;
  margin: 0 -20px;
  padding: 12px 20px;
}

.account-title-wrap {
  align-items: center;
  display: flex;
}

.account-label {
  color: #1f1f1f;
  font-size: 16px;
  font-weight: 700;
  line-height: 24px;
}

.account-name {
  color: #1f1f1f;
  font-size: 16px;
  font-weight: 700;
  line-height: 24px;
  margin-left: 8px;
}

.edit-icon {
  color: #1677ff;
  cursor: pointer;
  font-size: 18px;
  margin-left: 10px;
}

.student-card {
  background: #fff;
  border-radius: 16px;
  margin-top: 16px;
  padding: 16px 20px;
}

.section-title {
  color: #222;
  font-size: 16px;
  font-weight: 600;
  line-height: 24px;
  margin-bottom: 14px;
}

.student-list {
  align-items: flex-start;
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding-top: 8px;
}

.linked-student-card {
  align-items: center;
  background: #f5f5f7;
  border-radius: 20px;
  display: flex;
  height: 40px;
  max-width: 100%;
  padding: 4px 14px 4px 8px;
  position: relative;
  width: fit-content;
}

.main-student-badge {
  background: #1677ff;
  border: 2px solid #fff;
  border-radius: 9px;
  box-shadow: 0 2px 6px rgba(22, 119, 255, 0.25);
  color: #fff;
  font-size: 10px;
  left: 3px;
  line-height: 16px;
  padding: 0 8px;
  position: absolute;
  top: -14px;
  z-index: 2;
}

.linked-student-main {
  align-items: center;
  cursor: pointer;
  display: flex;
  flex: 0 1 auto;
}

.student-avatar-wrap {
  border-radius: 999px;
  height: 40px;
  margin-right: 10px;
  overflow: hidden;
  width: 40px;
}

.student-avatar-image {
  display: block;
  height: 100%;
  object-fit: cover;
  width: 100%;
}

.student-name {
  color: #222;
  font-size: 18px;
  font-weight: 600;
  line-height: 22px;
  margin-right: 12px;
  white-space: nowrap;
}

.student-phone {
  color: #333;
  font-size: 14px;
  line-height: 20px;
  margin-right: 12px;
  white-space: nowrap;
}

.edit-relation-link {
  color: #1677ff;
  font-size: 14px;
  font-weight: 500;
  line-height: 22px;
  margin-left: 10px;
  white-space: nowrap;
}

.summary-scroll {
  margin-top: 12px;
  overflow-x: auto;
  overflow-y: hidden;
  padding-bottom: 8px;

  &::-webkit-scrollbar {
    height: 6px;
  }

  &::-webkit-scrollbar-track {
    background: transparent;
  }

  &::-webkit-scrollbar-thumb {
    background: #d9e6ff;
    border-radius: 999px;
  }

  &::-webkit-scrollbar-thumb:hover {
    background: #b7d0ff;
  }

  scrollbar-width: thin;
  scrollbar-color: #d9e6ff transparent;
}

.summary-grid {
  --summary-side-padding: 12px;
  background: #fff;
  border-radius: 16px;
  display: grid;
  gap: 0;
  grid-auto-columns: minmax(145px, 1fr);
  grid-auto-flow: column;
  min-width: max-content;
  padding: 4px var(--summary-side-padding);
}

.summary-item {
  border-right: 1px solid #f0f0f0;
  min-height: 82px;
  padding: 12px 8px;
}

.summary-item:last-child {
  border-right: none;
}

.summary-item-highlight .summary-value {
  color: #1677ff;
}

.summary-item-fixed-left {
  background: #fff;
  left: var(--summary-side-padding);
  position: sticky;
  z-index: 3;
}

.summary-item-fixed-left::before {
  background: #fff;
  bottom: 0;
  content: "";
  left: calc(-1 * var(--summary-side-padding));
  pointer-events: none;
  position: absolute;
  top: 0;
  width: var(--summary-side-padding);
}

.summary-label {
  align-items: center;
  color: #888;
  display: flex;
  font-size: 13px;
  gap: 4px;
  line-height: 20px;
}

.summary-tip-icon {
  color: #999;
  cursor: pointer;
  font-size: 14px;
}

.summary-tip-icon:hover {
  color: #1677ff;
}

.summary-value {
  color: #222;
  font-family: DINAlternate-Bold, DINAlternate, sans-serif;
  font-size: 20px;
  font-weight: 700;
  line-height: 26px;
  margin-top: 6px;
}

.summary-value-emphasis {
  font-size: 24px;
  line-height: 30px;
}

.detail-card {
  background: #fff;
  border-radius: 16px;
  margin-top: 12px;
  padding: 16px;
}

.table-title {
  align-items: center;
  display: flex;
  justify-content: space-between;
  margin-bottom: 12px;
}

.help-popover {
  color: #666;
  line-height: 1.6;
  max-width: 480px;
}

.type-popover {
  max-width: 680px;
}

.help-title {
  border-bottom: 1px solid #f0f0f0;
  color: #222;
  font-weight: 500;
  margin-bottom: 8px;
  padding-bottom: 8px;
}

.flow-tag {
  background: #e6f0ff;
  border-radius: 999px;
  color: #06f;
  display: inline-block;
  font-size: 12px;
  line-height: 18px;
  padding: 2px 10px;
}

.flow-tag--refund {
  background: #fff1f0;
  color: #cf1322;
}

.total {
  align-items: center;
  color: #222;
  display: flex;
  padding-left: 10px;
  position: relative;

  &::before {
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    display: inline-block;
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

:deep(.ant-drawer-title) {
  color: #222;
  font-size: 20px;
  font-weight: 700;
}

:deep(.ant-table-cell) {
  white-space: nowrap;
}

:deep(.ant-tag) {
  margin-right: 0;
}
</style>
