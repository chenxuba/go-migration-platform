<script setup lang="ts">
import { computed, ref } from 'vue'
import { supervisionTasks } from '../shared/mock-data'

const status = ref<string | undefined>(undefined)

const statusOptions = [
  { label: '全部状态', value: undefined },
  { label: '待下发', value: '待下发' },
  { label: '进行中', value: '进行中' },
  { label: '待复核', value: '待复核' },
]

const columns = [
  { title: '任务名称', dataIndex: 'taskName', key: 'taskName', width: 260 },
  { title: '任务类型', dataIndex: 'taskType', key: 'taskType', width: 120 },
  { title: '所属区域', dataIndex: 'regionName', key: 'regionName', width: 120 },
  { title: '状态', dataIndex: 'status', key: 'status', width: 100 },
  { title: '责任单位', dataIndex: 'owner', key: 'owner', width: 150 },
  { title: '截止时间', dataIndex: 'dueDate', key: 'dueDate', width: 120 },
  { title: '完成度', dataIndex: 'completionRate', key: 'completionRate', width: 100 },
]

const filteredData = computed(() =>
  supervisionTasks.filter(item => !status.value || item.status === status.value),
)
</script>

<template>
  <div class="gov-page">
    <a-card :bordered="false">
      <div class="page-head">
        <div>
          <div class="page-head__title">
            督导任务
          </div>
          <div class="page-head__desc">
            这里作为 G 端后续承接巡查下发、整改复核、风险预警闭环的入口。当前先放示意数据，方便把政府端菜单和页面骨架单独拆出来。
          </div>
        </div>
        <div class="page-head__actions">
          <a-select v-model:value="status" :options="statusOptions" allow-clear placeholder="筛选状态" style="width: 180px" />
          <a-button type="primary">
            新建督导任务
          </a-button>
        </div>
      </div>
    </a-card>

    <a-card :bordered="false">
      <a-table :columns="columns" :data-source="filteredData" :pagination="false" :scroll="{ x: 1100 }" size="small">
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'status'">
            <span class="pill" :class="{
              'pill--blue': record.status === '待下发',
              'pill--green': record.status === '进行中',
              'pill--orange': record.status === '待复核',
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
  max-width: 780px;
  color: #6b7280;
  line-height: 24px;
}

.page-head__actions {
  display: flex;
  gap: 12px;
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

.pill--blue {
  color: #1d4ed8;
  background: #eff6ff;
}

@media (max-width: 1200px) {
  .page-head {
    flex-direction: column;
  }
}
</style>
