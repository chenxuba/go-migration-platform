<script setup lang="ts">
import { computed, ref } from 'vue'
import { institutionList } from '../shared/mock-data'

const regionKeyword = ref('')
const riskLevel = ref<string | undefined>(undefined)
const status = ref<string | undefined>(undefined)

const riskOptions = [
  { label: '全部风险', value: undefined },
  { label: '低风险', value: '低风险' },
  { label: '中风险', value: '中风险' },
  { label: '高风险', value: '高风险' },
]

const statusOptions = [
  { label: '全部状态', value: undefined },
  { label: '正常营业', value: '正常营业' },
  { label: '限期整改', value: '限期整改' },
  { label: '暂停接诊', value: '暂停接诊' },
]

const columns = [
  { title: '机构名称', dataIndex: 'institutionName', key: 'institutionName', width: 220 },
  { title: '所属区域', dataIndex: 'regionName', key: 'regionName', width: 180 },
  { title: '监管层级', dataIndex: 'level', key: 'level', width: 100 },
  { title: '机构类型', dataIndex: 'institutionType', key: 'institutionType', width: 160 },
  { title: '风险等级', dataIndex: 'riskLevel', key: 'riskLevel', width: 110 },
  { title: '营业状态', dataIndex: 'status', key: 'status', width: 110 },
  { title: '最近巡查', dataIndex: 'latestInspectionAt', key: 'latestInspectionAt', width: 120 },
  { title: '负责人', dataIndex: 'principal', key: 'principal', width: 90 },
  { title: '联系电话', dataIndex: 'phone', key: 'phone', width: 130 },
]

const filteredData = computed(() =>
  institutionList.filter((item) => {
    const matchesKeyword = !regionKeyword.value
      || item.institutionName.includes(regionKeyword.value)
      || item.regionName.includes(regionKeyword.value)
    const matchesRisk = !riskLevel.value || item.riskLevel === riskLevel.value
    const matchesStatus = !status.value || item.status === status.value
    return matchesKeyword && matchesRisk && matchesStatus
  }),
)

function resetFilters() {
  regionKeyword.value = ''
  riskLevel.value = undefined
  status.value = undefined
}
</script>

<template>
  <div class="gov-page">
    <a-card :bordered="false">
      <div class="page-head">
        <div>
          <div class="page-head__title">
            机构监管
          </div>
          <div class="page-head__desc">
            这里先放了一版 G 端机构监管页骨架，后续可以继续接机构备案、风险评级、处罚记录和跨区域监管台账。
          </div>
        </div>
      </div>

      <div class="filter-row">
        <a-input
          v-model:value="regionKeyword"
          allow-clear
          placeholder="搜索机构名称 / 所属区域"
          class="filter-item filter-item--lg"
        />
        <a-select v-model:value="riskLevel" :options="riskOptions" allow-clear placeholder="风险等级" class="filter-item" />
        <a-select v-model:value="status" :options="statusOptions" allow-clear placeholder="营业状态" class="filter-item" />
        <a-button @click="resetFilters">
          清空筛选
        </a-button>
      </div>
    </a-card>

    <a-card :bordered="false">
      <a-table :columns="columns" :data-source="filteredData" :pagination="false" :scroll="{ x: 1280 }" size="small">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'riskLevel'">
            <span class="pill" :class="{
              'pill--green': record.riskLevel === '低风险',
              'pill--orange': record.riskLevel === '中风险',
              'pill--red': record.riskLevel === '高风险',
            }">
              {{ record.riskLevel }}
            </span>
          </template>
          <template v-else-if="column.key === 'status'">
            <span class="pill" :class="{
              'pill--blue': record.status === '正常营业',
              'pill--orange': record.status === '限期整改',
              'pill--red': record.status === '暂停接诊',
            }">
              {{ record.status }}
            </span>
          </template>
        </template>
      </a-table>
    </a-card>
  </div>
</template>

<style scoped lang="less">
.gov-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
}

.page-head__title {
  margin-bottom: 8px;
  color: #1f2329;
  font-size: 24px;
  font-weight: 700;
}

.page-head__desc {
  color: #6b7280;
  line-height: 24px;
}

.filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 18px;
}

.filter-item {
  width: 180px;
}

.filter-item--lg {
  width: 280px;
}

.pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 72px;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
}

.pill--green {
  color: #047857;
  background: #ecfdf5;
}

.pill--orange {
  color: #c2410c;
  background: #fff7ed;
}

.pill--red {
  color: #b91c1c;
  background: #fef2f2;
}

.pill--blue {
  color: #1d4ed8;
  background: #eff6ff;
}
</style>
