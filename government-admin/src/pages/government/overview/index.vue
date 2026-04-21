<script setup lang="ts">
import type { TableColumnsType } from 'ant-design-vue'
import { computed, onMounted, ref } from 'vue'
import { getGovernmentOverviewApi, type GovernmentOverviewEntry, type GovernmentOverviewPayload } from '@/api/government/overview'
import messageService from '@/utils/messageService'

const loading = ref(false)
const overview = ref<GovernmentOverviewPayload | null>(null)

const columns: TableColumnsType<GovernmentOverviewEntry> = [
  { title: '区域名称', dataIndex: 'regionName', key: 'regionName' },
  { title: '层级', dataIndex: 'levelLabel', key: 'levelLabel', width: 110, align: 'center' as const },
  { title: '机构数', dataIndex: 'institutionCount', key: 'institutionCount', width: 100, align: 'right' as const },
  { title: '在读学员', dataIndex: 'readingStudentCount', key: 'readingStudentCount', width: 110, align: 'right' as const },
  { title: '意向学员', dataIndex: 'intentStudentCount', key: 'intentStudentCount', width: 110, align: 'right' as const },
  { title: '订单总量', dataIndex: 'orderCount', key: 'orderCount', width: 110, align: 'right' as const },
]

const statCards = computed(() => {
  const payload = overview.value
  if (!payload) {
    return []
  }
  return [
    {
      key: 'institutions',
      label: '纳管机构',
      value: `${payload.institutionCount}`,
      unit: '家',
      tone: 'blue',
    },
    {
      key: 'regions',
      label: resolveRegionStatLabel(payload),
      value: `${resolveRegionStatValue(payload)}`,
      unit: '个',
      tone: 'cyan',
    },
    {
      key: 'students',
      label: '在读学员',
      value: `${payload.readingStudentCount}`,
      unit: '人',
      tone: 'green',
    },
    {
      key: 'orders',
      label: '订单总量',
      value: `${payload.orderCount}`,
      unit: '笔',
      tone: 'orange',
    },
  ]
})

const noteItems = computed(() => {
  const payload = overview.value
  if (!payload) {
    return [
      '当前页将按真实监管账号的辖区范围汇总机构台账、在读学员和订单数据。',
      '首页不再使用示意性的预警、整改、督导任务指标。',
    ]
  }

  return [
    '首页统计口径来自现有机构业务：机构台账、在读学员、意向学员和订单总量。',
    payload.scopeCount > 1
      ? `当前账号已配置 ${payload.scopeCount} 个监管范围，首页已按这些辖区合并汇总。`
      : `当前账号监管范围为 ${payload.scopeText}。`,
    `区域表按${payload.level === 'super' ? '省级' : payload.level === 'province' ? '市级' : '区县级'}维度聚合，方便先看辖区体量，再下钻机构明细。`,
  ]
})

function resolveRegionStatLabel(payload: GovernmentOverviewPayload) {
  switch (payload.level) {
    case 'super':
      return '覆盖省份'
    case 'province':
      return '下辖城市'
    case 'city':
      return '下辖区县'
    case 'district':
      return payload.scopeCount > 1 ? '监管区县' : '当前辖区'
    default:
      return '覆盖区域'
  }
}

function resolveRegionStatValue(payload: GovernmentOverviewPayload) {
  if (payload.level === 'district') {
    return payload.scopeCount || payload.regionalSummary.length || 0
  }
  return payload.subordinateRegionCount
}

function resolveRequestErrorMessage(error: any, fallback: string) {
  return String(error?.response?.data?.message || error?.message || fallback).trim() || fallback
}

function getLevelColor(levelLabel?: string) {
  const value = String(levelLabel || '').trim()
  if (value === '省级')
    return 'blue'
  if (value === '市级')
    return 'cyan'
  if (value === '区县级')
    return 'geekblue'
  return 'default'
}

async function fetchOverview() {
  loading.value = true
  try {
    const res = await getGovernmentOverviewApi()
    if (res.code !== 200 || !res.result) {
      messageService.error(res.message || '获取监管总览失败')
      return
    }
    overview.value = res.result
  }
  catch (error: any) {
    console.error('fetch government overview failed', error)
    messageService.error(resolveRequestErrorMessage(error, '获取监管总览失败'))
  }
  finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchOverview()
})
</script>

<template>
  <div class="gov-page">
    <section class="page-header">
      <div class="page-header__main">
        <div class="page-header__eyebrow">
          G 端 / 监管驾驶舱
        </div>
        <h1 class="page-header__title">
          康复机构监管平台
        </h1>
      </div>

      <div class="page-header__badge">
        {{ overview?.levelLabel || '监管视角' }}
      </div>
    </section>

    <section class="scope-strip">
      <div class="scope-strip__item">
        <span class="scope-strip__label">当前层级</span>
        <span class="scope-strip__value">{{ overview?.levelLabel || '--' }}</span>
      </div>
      <div class="scope-strip__item">
        <span class="scope-strip__label">监管区域</span>
        <span class="scope-strip__value">{{ overview?.scopeText || '--' }}</span>
      </div>
      <div class="scope-strip__item">
        <span class="scope-strip__label">行政区划码</span>
        <span class="scope-strip__value">{{ overview?.scopeCodeText || '--' }}</span>
      </div>
    </section>

    <section class="stat-grid">
      <article v-for="card in statCards" :key="card.key" class="stat-card" :class="`stat-card--${card.tone}`">
        <div class="stat-card__label">
          {{ card.label }}
        </div>
        <div class="stat-card__value">
          {{ card.value }}
          <span class="stat-card__unit">{{ card.unit }}</span>
        </div>
      </article>
    </section>

    <section class="content-grid">
      <a-card title="区域监管概览" :bordered="false">
        <a-table
          :columns="columns"
          :data-source="overview?.regionalSummary || []"
          :loading="loading"
          :pagination="false"
          row-key="regionCode"
          size="small"
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'levelLabel'">
              <a-tag :color="getLevelColor(record.levelLabel)">
                {{ record.levelLabel }}
              </a-tag>
            </template>
          </template>
        </a-table>
      </a-card>

      <a-card title="统计口径" :bordered="false">
        <div class="note-list">
          <div v-for="item in noteItems" :key="item" class="note-list__item">
            {{ item }}
          </div>
        </div>
      </a-card>
    </section>
  </div>
</template>

<style scoped lang="less">
.gov-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
  min-width: 0;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 22px;
  border-radius: 14px;
  background: linear-gradient(135deg, #eef4ff 0%, #ffffff 100%);
}

.page-header__main {
  min-width: 0;
}

.page-header__eyebrow {
  margin-bottom: 10px;
  color: #1d4ed8;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.page-header__title {
  margin: 0;
  color: #102a43;
  font-size: 20px;
  font-weight: 700;
}

.page-header__badge {
  flex-shrink: 0;
  padding: 8px 14px;
  border-radius: 999px;
  background: rgba(29, 78, 216, 0.08);
  color: #1d4ed8;
  font-size: 14px;
  font-weight: 700;
  line-height: 20px;
  white-space: nowrap;
}

.scope-strip {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.scope-strip__item {
  padding: 16px 18px;
  border-radius: 14px;
  background: #fff;
}

.scope-strip__label {
  display: block;
  margin-bottom: 6px;
  color: #8c8c8c;
  font-size: 13px;
}

.scope-strip__value {
  color: #1f2329;
  font-size: 18px;
  font-weight: 600;
}

.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.stat-card {
  padding: 20px 18px;
  border-radius: 16px;
  background: #fff;
}

.stat-card__label {
  margin-bottom: 14px;
  color: #6b7280;
  font-size: 13px;
}

.stat-card__value {
  color: #111827;
  font-size: 30px;
  font-weight: 700;
  line-height: 1;
}

.stat-card__unit {
  margin-left: 6px;
  color: #6b7280;
  font-size: 14px;
  font-weight: 500;
}

.stat-card--blue {
  box-shadow: inset 0 0 0 1px rgba(59, 130, 246, 0.12);
}

.stat-card--cyan {
  box-shadow: inset 0 0 0 1px rgba(6, 182, 212, 0.12);
}

.stat-card--green {
  box-shadow: inset 0 0 0 1px rgba(16, 185, 129, 0.12);
}

.stat-card--orange {
  box-shadow: inset 0 0 0 1px rgba(249, 115, 22, 0.12);
}

.content-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.7fr) minmax(300px, 1fr);
  gap: 16px;
}

.note-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  color: #4b5563;
  line-height: 24px;
}

.note-list__item {
  padding: 14px 16px;
  border-radius: 12px;
  background: #f8fafc;
}

@media (max-width: 1200px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .scope-strip,
  .stat-grid,
  .content-grid {
    grid-template-columns: 1fr;
  }
}
</style>
