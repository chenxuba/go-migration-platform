<script setup>
import { LeftOutlined } from '@ant-design/icons-vue'
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getOrderImportTaskDetailApi,
  getOrderImportTaskRecordListApi,
} from '~@/api/finance-center/order-import'
import {
  getRechargeAccountImportTaskDetailApi,
  getRechargeAccountImportTaskRecordListApi,
} from '@/api/finance-center/recharge-account'
import messageService from '~@/utils/messageService'

const router = useRouter()
const route = useRoute()
const taskId = computed(() => String(route.params.id || ''))
const isRechargeImport = computed(() => route.path.includes('/import-recharge-account'))
const pageTitle = computed(() => isRechargeImport.value ? '储值账户导入明细' : '导入明细')
const detail = reactive({
  fileName: '',
  uploadStaffName: '',
  createdTime: '',
  totalRows: 0,
  executedRows: 0,
  errorRows: 0,
  status: 0,
})
const columns = ref([])
const rows = ref([])
let pollingTimer = null

function getColumnWidth(title) {
  switch (`${title || ''}`.trim()) {
    case '学员姓名':
      return 140
    case '手机号':
      return 150
    case '手机号归属人':
      return 170
    case '渠道':
      return 170
    case '性别':
      return 110
    case '生日':
    case '经办日期':
    case '有效开始日期':
    case '有效结束日期':
    case '有效结束日期(含赠送天数)':
    case '有效期至':
      return 180
    case '微信号':
    case '年级':
    case '会员号':
      return 140
    case '报读课程':
    case '报读班级':
      return 180
    case '购买课时数':
    case '赠送课时数':
    case '已上课时数':
    case '赠送天数':
    case '已上天数':
    case '购买金额':
    case '残联金额':
    case '赠送金额':
    case '已上金额':
    case '实收金额':
    case '欠费金额':
      return 150
    case '收款方式':
    case '收款账户':
    case '支付单号':
    case '对方账户':
    case '订单标签':
    case '销售':
    case '订单销售员':
      return 160
    case '订单备注':
      return 180
    default:
      return 150
  }
}

const tableMinWidth = computed(() => {
  const dynamicWidth = columns.value.reduce((total, column) => total + getColumnWidth(column.title), 0)
  return dynamicWidth + 70 + 150
})

function goBack() {
  router.replace(isRechargeImport.value ? '/import-center/import-recharge-account/record' : '/import-center/import-order/record')
}

function resultText(row) {
  return row.status === 1 ? '导入成功' : (row.result || '导入失败')
}

function getCellDisplayText(cell) {
  const text = `${cell?.value || ''}`.trim()
  return text || '-'
}

async function loadDetail() {
  try {
    const [detailRes, abnormalRes, normalRes] = await Promise.all([
      (isRechargeImport.value ? getRechargeAccountImportTaskDetailApi : getOrderImportTaskDetailApi)({ taskId: taskId.value }),
      (isRechargeImport.value ? getRechargeAccountImportTaskRecordListApi : getOrderImportTaskRecordListApi)({
        queryModel: { taskId: taskId.value, type: 0 },
        sortModel: '',
        pageRequestModel: { needTotal: true, pageSize: 1000, pageIndex: 1, skipCount: 0 },
      }),
      (isRechargeImport.value ? getRechargeAccountImportTaskRecordListApi : getOrderImportTaskRecordListApi)({
        queryModel: { taskId: taskId.value, type: 1 },
        sortModel: '',
        pageRequestModel: { needTotal: true, pageSize: 1000, pageIndex: 1, skipCount: 0 },
      }),
    ])

    const detailData = detailRes.result || detailRes.data || {}
    Object.assign(detail, detailData)

    const abnormal = abnormalRes.result || abnormalRes.data || {}
    const normal = normalRes.result || normalRes.data || {}
    columns.value = abnormal.columns?.length ? abnormal.columns : (normal.columns || [])
    rows.value = [...(abnormal.list || []), ...(normal.list || [])]
  }
  catch (error) {
    console.error(error)
    messageService.error('加载导入详情失败')
  }
}

onMounted(() => {
  loadDetail()
  pollingTimer = window.setInterval(() => {
    if (detail.status === 4)
      loadDetail()
  }, 2000)
})

onUnmounted(() => {
  if (pollingTimer) {
    clearInterval(pollingTimer)
    pollingTimer = null
  }
})
</script>

<template>
  <div class="import-record-layout">
    <div class="work-top flex justify-between items-center h-56px bg-#fff">
      <div class="work-top-left flex items-center">
        <div class="import-header-logo" title="导入中心" aria-hidden="true" />
        <span class="text-20px text-#06f font500 ml-16px flex items-center cursor-pointer" @click="goBack">
          <LeftOutlined class="mt--1px" /> 返回
        </span>
      </div>
    </div>

    <div class="work-main">
      <div class="work-main-card">
        <div class="record-title-row">
          <div class="record-title">
            {{ detail.fileName }}
          </div>
        </div>

        <div class="summary-strip">
          <div class="summary-title">
            {{ pageTitle }}
          </div>
          <div class="summary-items">
            <span class="summary-item">
              <span class="summary-label">导入时间</span>
              <span class="summary-value">{{ detail.createdTime ? detail.createdTime.replace('T', ' ').slice(0, 16) : '-' }}</span>
            </span>
            <span class="summary-item">
              <span class="summary-label">导入人</span>
              <span class="summary-value">{{ detail.uploadStaffName || '-' }}</span>
            </span>
            <span class="summary-item">
              <span class="summary-label">导入结果</span>
              <span class="summary-value">
                共计 {{ detail.totalRows || 0 }}（<span :class="detail.executedRows > 0 ? 'success-count' : 'neutral-count'">成功{{ detail.executedRows || 0 }}</span>/<span :class="detail.errorRows > 0 ? 'fail-count' : 'neutral-count'">失败{{ detail.errorRows || 0 }}</span>）
              </span>
            </span>
          </div>
        </div>

        <div class="table-wrap">
          <table class="detail-table" :style="{ minWidth: `${tableMinWidth}px` }">
            <colgroup>
              <col style="width: 70px">
              <col v-for="column in columns" :key="column.key" :style="{ width: `${getColumnWidth(column.title)}px` }">
              <col style="width: 150px">
            </colgroup>
            <thead>
              <tr>
                <th>序号</th>
                <th v-for="column in columns" :key="column.key">
                  <span v-if="column.required" class="required">*</span>{{ column.title }}
                </th>
                <th class="result-column">导入结果</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in rows" :key="row.id">
                <td>{{ row.rowNo }}</td>
                <td v-for="cell in row.cells" :key="cell.key">
                  {{ getCellDisplayText(cell) }}
                </td>
                <td class="result-column" :class="row.status === 1 ? 'success' : 'fail'">
                  {{ resultText(row) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.import-record-layout {
  min-height: 100vh;
  background: #f7f7fd;
}

.import-header-logo {
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.22) 0%, transparent 42%),
    linear-gradient(145deg, #2b8cff 0%, #0066ff 45%, #0050d8 100%);
  position: relative;
  overflow: hidden;
}

.import-header-logo::before {
  content: '';
  position: absolute;
  left: 12px;
  top: 15px;
  width: 32px;
  height: 26px;
  background-color: rgba(255, 255, 255, 0.94);
  background-image:
    linear-gradient(rgba(0, 102, 255, 0.22), rgba(0, 102, 255, 0.22)),
    linear-gradient(rgba(0, 102, 255, 0.18), rgba(0, 102, 255, 0.18)),
    linear-gradient(rgba(0, 102, 255, 0.14), rgba(0, 102, 255, 0.14));
  background-size: 24px 2px, 18px 2px, 22px 2px;
  background-position: 4px 8px, 4px 14px, 4px 20px;
  background-repeat: no-repeat;
}

.work-main {
  display: flex;
  justify-content: center;
  padding: 40px 0;
}

.work-main-card {
  width: 1300px;
  min-height: 720px;
  padding: 40px 48px 48px;
  border-radius: 24px;
  background: #fff;
  box-shadow: 0 0 32px rgba(0, 0, 0, 0.08);
}

.record-title {
  font-size: 24px;
  font-weight: 600;
  color: #222;
}

.summary-strip {
  margin-top: 24px;
  padding: 14px 18px;
  border-radius: 12px;
  background: #f5f7ff;
  display: flex;
  align-items: center;
  gap: 24px;
}

.summary-title {
  flex-shrink: 0;
  color: #222;
  font-size: 18px;
  font-weight: 600;
}

.summary-items {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 20px;
}

.summary-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #666;
  font-size: 14px;
}

.summary-label {
  color: #8a8f99;
}

.summary-value {
  color: #222;
}

.success-count {
  color: #16a34a;
  font-weight: 600;
}

.fail-count {
  color: #ff4d4f;
  font-weight: 600;
}

.neutral-count {
  color: #999;
  font-weight: 500;
}

.table-wrap {
  margin-top: 24px;
  overflow-x: auto;
  overflow-y: hidden;
}

.detail-table {
  width: max-content;
  min-width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.detail-table th,
.detail-table td {
  padding: 14px 12px;
  border-bottom: 1px solid #f0f0f0;
  text-align: left;
  white-space: nowrap;
  background: #fff;
}

.detail-table th {
  background: #fafafa;
  color: #222;
  font-weight: 600;
}

.detail-table .result-column {
  position: sticky;
  right: 0;
  z-index: 2;
  box-shadow: -8px 0 12px rgba(255, 255, 255, 0.96);
}

.detail-table th.result-column {
  z-index: 3;
  background: #fafafa;
}

.required {
  color: #ff4d4f;
}

.success {
  color: #16a34a;
}

.fail {
  color: #ff4d4f;
}
</style>
