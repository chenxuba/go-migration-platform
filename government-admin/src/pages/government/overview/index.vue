<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { getGovernmentOverviewApi, type GovernmentOverviewEntry, type GovernmentOverviewPayload } from '@/api/government/overview'
import messageService from '@/utils/messageService'

interface RegionalCard extends GovernmentOverviewEntry {
  rank: number
  institutionPercent: number
  studentPercent: number
  orderPercent: number
}

const loading = ref(false)
const overview = ref<GovernmentOverviewPayload | null>(null)

const intentStudentCount = computed(() =>
  (overview.value?.regionalSummary || []).reduce((sum, item) => sum + Number(item.intentStudentCount || 0), 0),
)

const totalStudentCount = computed(() =>
  Number(overview.value?.readingStudentCount || 0) + intentStudentCount.value,
)

const regionStatValue = computed(() => {
  const payload = overview.value
  if (!payload)
    return 0

  if (payload.level === 'district')
    return payload.scopeCount || payload.regionalSummary.length || 0

  return payload.subordinateRegionCount
})

const actualMetricCards = computed(() => {
  const payload = overview.value
  if (!payload) {
    return []
  }

  return [
    {
      key: 'institutions',
      label: '纳管机构',
      value: payload.institutionCount,
      unit: '家',
      foot: '机构台账总量',
      tone: 'blue',
    },
    {
      key: 'regions',
      label: resolveRegionStatLabel(payload),
      value: regionStatValue.value,
      unit: '个',
      foot: resolveRegionDimensionLabel(payload),
      tone: 'cyan',
    },
    {
      key: 'reading',
      label: '在读学员',
      value: payload.readingStudentCount,
      unit: '人',
      foot: '现有报读学员',
      tone: 'green',
    },
    {
      key: 'intent',
      label: '意向学员',
      value: intentStudentCount.value,
      unit: '人',
      foot: '招生跟进池',
      tone: 'amber',
    },
    {
      key: 'orders',
      label: '订单总量',
      value: payload.orderCount,
      unit: '笔',
      foot: '财务业务累计',
      tone: 'violet',
    },
  ]
})

const boundaryItems = computed(() => {
  const payload = overview.value
  return [
    {
      key: 'code',
      label: '行政区划码',
      value: payload?.scopeCodeText || '--',
      accent: 'cyan',
    },
    {
      key: 'dimension',
      label: '汇总口径',
      value: payload ? resolveRegionDimensionLabel(payload) : '--',
      accent: 'amber',
    },
    {
      key: 'regions',
      label: payload ? resolveRegionStatLabel(payload) : '覆盖单元',
      value: payload ? `${formatNumber(regionStatValue.value)} 个` : '--',
      accent: 'green',
    },
    {
      key: 'students',
      label: '学员池规模',
      value: `${formatNumber(totalStudentCount.value)} 人`,
      accent: 'blue',
    },
  ]
})

const coreQuickMetrics = computed(() => {
  const payload = overview.value
  if (!payload) {
    return []
  }

  return [
    {
      key: 'regions',
      label: resolveRegionStatLabel(payload),
      value: regionStatValue.value,
      unit: '个',
    },
    {
      key: 'reading',
      label: '在读学员',
      value: payload.readingStudentCount,
      unit: '人',
    },
    {
      key: 'orders',
      label: '订单总量',
      value: payload.orderCount,
      unit: '笔',
    },
  ]
})

const studentMix = computed(() => {
  const reading = Number(overview.value?.readingStudentCount || 0)
  const intent = intentStudentCount.value
  const total = totalStudentCount.value

  return {
    reading,
    intent,
    total,
    readingPercent: calcPercent(reading, total),
    intentPercent: calcPercent(intent, total),
  }
})

const regionCards = computed<RegionalCard[]>(() => {
  const source = overview.value?.regionalSummary || []
  if (!source.length) {
    return []
  }

  const sorted = [...source].sort((left, right) => {
    if (right.institutionCount !== left.institutionCount)
      return right.institutionCount - left.institutionCount
    if (right.readingStudentCount !== left.readingStudentCount)
      return right.readingStudentCount - left.readingStudentCount
    return right.orderCount - left.orderCount
  })

  const maxInstitution = Math.max(...sorted.map(item => Number(item.institutionCount || 0)), 1)
  const maxStudent = Math.max(...sorted.map(item => Number(item.readingStudentCount || 0)), 1)
  const maxOrder = Math.max(...sorted.map(item => Number(item.orderCount || 0)), 1)

  return sorted.map((item, index) => ({
    ...item,
    rank: index + 1,
    institutionPercent: calcPercent(item.institutionCount, maxInstitution),
    studentPercent: calcPercent(item.readingStudentCount, maxStudent),
    orderPercent: calcPercent(item.orderCount, maxOrder),
  }))
})

const focusRegion = computed(() => regionCards.value[0] || null)

const focusRegionShare = computed(() => {
  const payload = overview.value
  const focus = focusRegion.value
  if (!payload || !focus || !payload.institutionCount) {
    return 0
  }
  return calcPercent(focus.institutionCount, payload.institutionCount)
})

const derivedInsightCards = computed(() => {
  const payload = overview.value
  const institutionCount = Number(payload?.institutionCount || 0)
  const regionCount = Math.max(regionStatValue.value, 1)
  const reading = Number(payload?.readingStudentCount || 0)
  const orders = Number(payload?.orderCount || 0)
  const totalStudents = totalStudentCount.value

  return [
    {
      key: 'readingRatio',
      label: '在读占比',
      value: `${studentMix.value.readingPercent}%`,
      desc: `在读 ${formatNumber(reading)} / 学员池 ${formatNumber(totalStudents)}`,
      tone: 'cyan',
    },
    {
      key: 'intentRatio',
      label: '意向占比',
      value: `${studentMix.value.intentPercent}%`,
      desc: `意向 ${formatNumber(intentStudentCount.value)} / 学员池 ${formatNumber(totalStudents)}`,
      tone: 'amber',
    },
    {
      key: 'ordersPerInstitution',
      label: '平均每机构订单',
      value: formatAverage(orders, institutionCount),
      desc: `订单 ${formatNumber(orders)} / 机构 ${formatNumber(institutionCount)}`,
      tone: 'violet',
    },
    {
      key: 'institutionPerRegion',
      label: '平均每区域机构',
      value: formatAverage(institutionCount, regionCount),
      desc: `机构 ${formatNumber(institutionCount)} / 区域 ${formatNumber(regionCount)}`,
      tone: 'green',
    },
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

function resolveRegionDimensionLabel(payload: GovernmentOverviewPayload) {
  switch (payload.level) {
    case 'super':
      return '按省级维度聚合'
    case 'province':
      return '按市级维度聚合'
    case 'city':
      return '按区县维度聚合'
    case 'district':
      return '按当前辖区聚合'
    default:
      return '按区域维度聚合'
  }
}

function resolveRequestErrorMessage(error: any, fallback: string) {
  return String(error?.response?.data?.message || error?.message || fallback).trim() || fallback
}

function calcPercent(value?: number, total?: number) {
  const currentValue = Number(value || 0)
  const currentTotal = Number(total || 0)
  if (!currentTotal) {
    return 0
  }
  return Math.min(100, Math.round(currentValue / currentTotal * 100))
}

function formatNumber(value?: number | string) {
  return new Intl.NumberFormat('zh-CN').format(Number(value || 0))
}

function formatAverage(numerator?: number, denominator?: number) {
  const currentNumerator = Number(numerator || 0)
  const currentDenominator = Number(denominator || 0)
  if (!currentDenominator) {
    return '0.0'
  }

  const average = currentNumerator / currentDenominator
  return average >= 100 ? average.toFixed(0) : average.toFixed(1)
}

function padRank(rank: number) {
  return String(rank || 0).padStart(2, '0')
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
  <a-spin :spinning="loading" tip="加载监管数据中...">
    <div class="screen-page">
      <section class="hero-banner">
        <div class="hero-banner__main">
          <div class="hero-banner__eyebrow">
            G 端 / 数据大屏
          </div>
          <h1 class="hero-banner__title">
            康复机构监管数据大屏
          </h1>
          <div class="hero-banner__signals">
            <span>机构台账</span>
            <span>区域聚合</span>
            <span>学员规模</span>
            <span>订单总量</span>
          </div>
        </div>

        <div class="hero-banner__meta">
          <div class="hero-banner__badge">
            {{ overview?.levelLabel || '--' }}
          </div>
          <div class="hero-banner__chip">
            <span>监管区域</span>
            <strong>{{ overview?.scopeText || '--' }}</strong>
          </div>
          <div class="hero-banner__chip">
            <span>{{ overview ? resolveRegionStatLabel(overview) : '覆盖单元' }}</span>
            <strong>{{ formatNumber(regionStatValue) }} 个</strong>
          </div>
        </div>
      </section>

      <section class="metric-ribbon">
        <article v-for="card in actualMetricCards" :key="card.key" class="metric-card" :class="`metric-card--${card.tone}`">
          <div class="metric-card__label">
            {{ card.label }}
          </div>
          <div class="metric-card__value">
            {{ formatNumber(card.value) }}
            <span class="metric-card__unit">{{ card.unit }}</span>
          </div>
          <div class="metric-card__foot">
            {{ card.foot }}
          </div>
        </article>
      </section>

      <section class="dashboard-grid">
        <article class="panel panel--identity">
          <div class="panel__title">
            监管边界
          </div>
          <div class="identity-list">
            <div v-for="item in boundaryItems" :key="item.key" class="identity-item" :class="`identity-item--${item.accent}`">
              <div class="identity-item__label">
                {{ item.label }}
              </div>
              <div class="identity-item__value">
                {{ item.value }}
              </div>
            </div>
          </div>
        </article>

        <article class="panel panel--core">
          <div class="panel__title">
            监管核心
          </div>
          <div class="core-stage">
            <div class="core-orb-wrap">
              <div class="core-orb">
                <div class="core-orb__ring core-orb__ring--outer" />
                <div class="core-orb__ring core-orb__ring--inner" />
                <div class="core-orb__content">
                  <div class="core-orb__value">
                    {{ formatNumber(overview?.institutionCount) }}
                  </div>
                  <div class="core-orb__label">
                    纳管机构
                  </div>
                  <div class="core-orb__sub">
                    {{ overview?.levelLabel || '--' }}视角
                  </div>
                </div>
              </div>
            </div>

            <div class="core-side">
              <div class="core-quick-metrics">
                <div v-for="item in coreQuickMetrics" :key="item.key" class="core-quick-metric">
                  <div class="core-quick-metric__label">
                    {{ item.label }}
                  </div>
                  <div class="core-quick-metric__value">
                    {{ formatNumber(item.value) }}
                    <span>{{ item.unit }}</span>
                  </div>
                </div>
              </div>

              <div class="mix-board">
                <div class="mix-board__head">
                  学员结构
                </div>
                <div class="mix-board__track">
                  <span
                    class="mix-board__segment mix-board__segment--reading"
                    :style="{ width: `${studentMix.readingPercent}%` }"
                  />
                  <span
                    class="mix-board__segment mix-board__segment--intent"
                    :style="{ width: `${studentMix.intentPercent}%` }"
                  />
                </div>
                <div class="mix-board__legend">
                  <div class="mix-board__legend-item">
                    <span class="mix-board__dot mix-board__dot--reading" />
                    在读 {{ formatNumber(studentMix.reading) }} 人
                  </div>
                  <div class="mix-board__legend-item">
                    <span class="mix-board__dot mix-board__dot--intent" />
                    意向 {{ formatNumber(studentMix.intent) }} 人
                  </div>
                </div>
              </div>
            </div>
          </div>
        </article>

        <article class="panel panel--insight">
          <div class="panel__title">
            结构解析
          </div>

          <div v-if="focusRegion" class="focus-region">
            <div class="focus-region__label">
              最高体量辖区
            </div>
            <div class="focus-region__name">
              {{ focusRegion.regionName }}
            </div>
            <div class="focus-region__meta">
              <span>{{ focusRegion.levelLabel }}</span>
              <span>{{ formatNumber(focusRegion.institutionCount) }} 家机构</span>
              <span>占全部机构 {{ focusRegionShare }}%</span>
            </div>
          </div>

          <div class="insight-grid">
            <article v-for="item in derivedInsightCards" :key="item.key" class="insight-card" :class="`insight-card--${item.tone}`">
              <div class="insight-card__label">
                {{ item.label }}
              </div>
              <div class="insight-card__value">
                {{ item.value }}
              </div>
              <div class="insight-card__desc">
                {{ item.desc }}
              </div>
            </article>
          </div>

          <div class="panel__note">
            比率与均值均由真实总量自动测算。
          </div>
        </article>
      </section>

      <section class="panel panel--regions">
        <div class="panel__title">
          区域监管矩阵
        </div>

        <div v-if="regionCards.length" class="region-grid">
          <article v-for="item in regionCards" :key="item.regionCode" class="region-card">
            <div class="region-card__header">
              <div>
                <div class="region-card__rank">
                  NO.{{ padRank(item.rank) }}
                </div>
                <div class="region-card__name">
                  {{ item.regionName }}
                </div>
              </div>
              <div class="region-card__tag">
                {{ item.levelLabel }}
              </div>
            </div>

            <div class="region-card__stats">
              <div class="region-stat">
                <span>机构</span>
                <strong>{{ formatNumber(item.institutionCount) }}</strong>
              </div>
              <div class="region-stat">
                <span>在读</span>
                <strong>{{ formatNumber(item.readingStudentCount) }}</strong>
              </div>
              <div class="region-stat">
                <span>意向</span>
                <strong>{{ formatNumber(item.intentStudentCount) }}</strong>
              </div>
              <div class="region-stat">
                <span>订单</span>
                <strong>{{ formatNumber(item.orderCount) }}</strong>
              </div>
            </div>

            <div class="region-bars">
              <div class="region-bar">
                <span>机构体量</span>
                <div class="region-bar__track">
                  <div class="region-bar__fill region-bar__fill--blue" :style="{ width: `${item.institutionPercent}%` }" />
                </div>
                <em>{{ item.institutionPercent }}%</em>
              </div>
              <div class="region-bar">
                <span>学员体量</span>
                <div class="region-bar__track">
                  <div class="region-bar__fill region-bar__fill--cyan" :style="{ width: `${item.studentPercent}%` }" />
                </div>
                <em>{{ item.studentPercent }}%</em>
              </div>
              <div class="region-bar">
                <span>订单体量</span>
                <div class="region-bar__track">
                  <div class="region-bar__fill region-bar__fill--amber" :style="{ width: `${item.orderPercent}%` }" />
                </div>
                <em>{{ item.orderPercent }}%</em>
              </div>
            </div>
          </article>
        </div>

        <div v-else class="empty-panel">
          当前监管范围内暂无可展示的区域聚合数据。
        </div>
      </section>
    </div>
  </a-spin>
</template>

<style scoped lang="less">
.screen-page {
  --screen-bg: #061428;
  --screen-bg-soft: rgba(7, 27, 52, 0.88);
  --screen-border: rgba(124, 196, 255, 0.16);
  --screen-text: #eef6ff;
  --screen-muted: rgba(214, 233, 255, 0.72);
  --screen-number-font: 'DIN Alternate', 'Bahnschrift', 'PingFang SC', 'Microsoft YaHei', sans-serif;
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 18px;
  width: 100%;
  min-width: 0;
  padding: 22px;
  overflow: hidden;
  border-radius: 28px;
  background:
    radial-gradient(circle at 12% 18%, rgba(0, 214, 255, 0.18), transparent 26%),
    radial-gradient(circle at 88% 4%, rgba(70, 109, 255, 0.18), transparent 24%),
    linear-gradient(180deg, #071427 0%, #091d35 52%, #06111f 100%);
  color: var(--screen-text);
  box-shadow:
    inset 0 0 0 1px rgba(124, 196, 255, 0.1),
    0 22px 60px rgba(1, 9, 22, 0.28);
}

.screen-page::before {
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(124, 196, 255, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(124, 196, 255, 0.05) 1px, transparent 1px);
  background-size: 24px 24px;
  opacity: 0.65;
  content: '';
  pointer-events: none;
}

.screen-page::after {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(120deg, transparent 18%, rgba(255, 255, 255, 0.06) 32%, transparent 46%) no-repeat;
  background-size: 220% 100%;
  animation: screen-scan 10s linear infinite;
  content: '';
  pointer-events: none;
}

.hero-banner,
.panel,
.metric-card {
  position: relative;
  overflow: hidden;
  border: 1px solid var(--screen-border);
  background: linear-gradient(180deg, rgba(7, 27, 52, 0.9) 0%, rgba(6, 17, 33, 0.84) 100%);
  box-shadow:
    inset 0 0 0 1px rgba(160, 215, 255, 0.03),
    0 14px 32px rgba(2, 11, 26, 0.22);
  backdrop-filter: blur(16px);
}

.hero-banner {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 24px;
  align-items: center;
  padding: 26px 28px;
  border-radius: 24px;
}

.hero-banner__eyebrow {
  margin-bottom: 10px;
  color: #84d8ff;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.hero-banner__title {
  margin: 0;
  color: #f4f9ff;
  font-size: 34px;
  font-weight: 700;
  letter-spacing: 0.02em;
}

.hero-banner__signals {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 14px;
}

.hero-banner__signals span {
  padding: 6px 12px;
  border: 1px solid rgba(124, 196, 255, 0.14);
  border-radius: 999px;
  background: rgba(10, 39, 71, 0.46);
  color: var(--screen-muted);
  font-size: 12px;
  letter-spacing: 0.06em;
}

.hero-banner__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
}

.hero-banner__badge {
  padding: 10px 18px;
  border-radius: 999px;
  background: linear-gradient(135deg, rgba(26, 93, 255, 0.18) 0%, rgba(0, 214, 255, 0.12) 100%);
  box-shadow: inset 0 0 0 1px rgba(132, 216, 255, 0.18);
  color: #8de7ff;
  font-size: 16px;
  font-weight: 700;
}

.hero-banner__chip {
  display: flex;
  flex-direction: column;
  min-width: 140px;
  padding: 10px 14px;
  border-radius: 16px;
  background: rgba(9, 35, 64, 0.7);
}

.hero-banner__chip span {
  margin-bottom: 6px;
  color: rgba(214, 233, 255, 0.6);
  font-size: 12px;
}

.hero-banner__chip strong {
  color: #f4f9ff;
  font-size: 16px;
  font-weight: 600;
}

.metric-ribbon {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 14px;
}

.metric-card {
  padding: 20px 20px 18px;
  border-radius: 20px;
}

.metric-card::before {
  position: absolute;
  left: 18px;
  right: 18px;
  top: 0;
  height: 2px;
  border-radius: 999px;
  content: '';
}

.metric-card__label {
  color: rgba(214, 233, 255, 0.72);
  font-size: 13px;
  letter-spacing: 0.04em;
}

.metric-card__value {
  margin-top: 12px;
  font-family: var(--screen-number-font);
  font-size: 38px;
  font-weight: 700;
  line-height: 1;
}

.metric-card__unit {
  margin-left: 8px;
  color: rgba(214, 233, 255, 0.7);
  font-size: 14px;
  font-weight: 500;
}

.metric-card__foot {
  margin-top: 14px;
  color: rgba(214, 233, 255, 0.56);
  font-size: 12px;
}

.metric-card--blue::before {
  background: linear-gradient(90deg, #64b5ff, transparent);
}

.metric-card--cyan::before {
  background: linear-gradient(90deg, #74f0ff, transparent);
}

.metric-card--green::before {
  background: linear-gradient(90deg, #54e5a6, transparent);
}

.metric-card--amber::before {
  background: linear-gradient(90deg, #ffc86a, transparent);
}

.metric-card--violet::before {
  background: linear-gradient(90deg, #7e93ff, transparent);
}

.dashboard-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.38fr) 332px;
  grid-template-areas:
    'core identity'
    'insight insight';
  gap: 18px;
  align-items: start;
}

.panel {
  border-radius: 24px;
  padding: 22px;
}

.panel--core {
  grid-area: core;
}

.panel--identity {
  grid-area: identity;
}

.panel--insight {
  grid-area: insight;
}

.panel__title {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 18px;
  color: #f4f9ff;
  font-size: 18px;
  font-weight: 700;
}

.panel__title::before {
  width: 10px;
  height: 10px;
  border-radius: 999px;
  background: linear-gradient(135deg, #74f0ff 0%, #5f79ff 100%);
  box-shadow: 0 0 14px rgba(116, 240, 255, 0.45);
  content: '';
}

.identity-list {
  display: grid;
  gap: 10px;
}

.identity-item {
  padding: 12px 14px;
  border-radius: 16px;
  background: rgba(9, 35, 64, 0.72);
  box-shadow: inset 0 0 0 1px rgba(124, 196, 255, 0.08);
}

.identity-item__label {
  margin-bottom: 8px;
  color: rgba(214, 233, 255, 0.56);
  font-size: 12px;
}

.identity-item__value {
  color: #f4f9ff;
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;
  word-break: break-all;
}

.identity-item--blue {
  box-shadow: inset 0 0 0 1px rgba(100, 181, 255, 0.14);
}

.identity-item--cyan {
  box-shadow: inset 0 0 0 1px rgba(116, 240, 255, 0.14);
}

.identity-item--amber {
  box-shadow: inset 0 0 0 1px rgba(255, 200, 106, 0.14);
}

.identity-item--green {
  box-shadow: inset 0 0 0 1px rgba(84, 229, 166, 0.14);
}

.core-stage {
  display: grid;
  grid-template-columns: minmax(260px, 0.95fr) minmax(0, 1fr);
  gap: 24px;
  align-items: center;
}

.core-orb-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 320px;
}

.core-orb {
  position: relative;
  width: 230px;
  height: 230px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background:
    radial-gradient(circle at 50% 45%, rgba(132, 216, 255, 0.28) 0%, rgba(38, 90, 156, 0.22) 34%, rgba(7, 27, 52, 0.2) 68%, transparent 100%),
    radial-gradient(circle, rgba(7, 27, 52, 0.85) 0%, rgba(5, 16, 31, 0.98) 100%);
  box-shadow:
    0 0 32px rgba(76, 162, 255, 0.18),
    inset 0 0 0 1px rgba(132, 216, 255, 0.16);
}

.core-orb__ring {
  position: absolute;
  inset: 0;
  border-radius: 50%;
}

.core-orb__ring--outer {
  inset: -18px;
  border: 1px dashed rgba(132, 216, 255, 0.24);
  animation: rotate-screen-ring 18s linear infinite;
}

.core-orb__ring--inner {
  inset: 14px;
  border: 1px solid rgba(132, 216, 255, 0.12);
  box-shadow: inset 0 0 18px rgba(116, 240, 255, 0.08);
  animation: pulse-screen-glow 3.4s ease-in-out infinite;
}

.core-orb__content {
  position: relative;
  z-index: 1;
  text-align: center;
}

.core-orb__value {
  font-family: var(--screen-number-font);
  color: #ffffff;
  font-size: 58px;
  font-weight: 700;
  line-height: 1;
}

.core-orb__label {
  margin-top: 12px;
  color: rgba(214, 233, 255, 0.82);
  font-size: 14px;
  letter-spacing: 0.12em;
}

.core-orb__sub {
  margin-top: 10px;
  color: rgba(132, 216, 255, 0.84);
  font-size: 13px;
}

.core-side {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.core-quick-metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.core-quick-metric {
  padding: 14px;
  border-radius: 18px;
  background: rgba(9, 35, 64, 0.72);
  box-shadow: inset 0 0 0 1px rgba(124, 196, 255, 0.08);
}

.core-quick-metric__label {
  color: rgba(214, 233, 255, 0.58);
  font-size: 12px;
}

.core-quick-metric__value {
  margin-top: 10px;
  font-family: var(--screen-number-font);
  color: #ffffff;
  font-size: 28px;
  font-weight: 700;
}

.core-quick-metric__value span {
  margin-left: 6px;
  color: rgba(214, 233, 255, 0.66);
  font-size: 13px;
  font-weight: 500;
}

.mix-board {
  padding: 16px 18px;
  border-radius: 18px;
  background: rgba(8, 31, 58, 0.78);
  box-shadow: inset 0 0 0 1px rgba(124, 196, 255, 0.08);
}

.mix-board__head {
  color: #f4f9ff;
  font-size: 14px;
  font-weight: 600;
}

.mix-board__track {
  display: flex;
  height: 14px;
  margin-top: 14px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
}

.mix-board__segment {
  height: 100%;
}

.mix-board__segment--reading {
  background: linear-gradient(90deg, #65dcff 0%, #4ca1ff 100%);
}

.mix-board__segment--intent {
  background: linear-gradient(90deg, #ffc86a 0%, #ff9f5a 100%);
}

.mix-board__legend {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
  margin-top: 14px;
}

.mix-board__legend-item {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: rgba(214, 233, 255, 0.76);
  font-size: 12px;
}

.mix-board__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.mix-board__dot--reading {
  background: #65dcff;
}

.mix-board__dot--intent {
  background: #ffc86a;
}

.focus-region {
  padding: 16px 18px;
  margin-bottom: 16px;
  border-radius: 18px;
  background:
    linear-gradient(135deg, rgba(20, 78, 149, 0.24) 0%, rgba(8, 31, 58, 0.22) 100%);
  box-shadow: inset 0 0 0 1px rgba(124, 196, 255, 0.1);
}

.focus-region__label {
  color: rgba(214, 233, 255, 0.56);
  font-size: 12px;
}

.focus-region__name {
  margin-top: 8px;
  color: #ffffff;
  font-size: 24px;
  font-weight: 700;
  line-height: 1.25;
}

.focus-region__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 10px;
  margin-top: 12px;
  color: rgba(214, 233, 255, 0.78);
  font-size: 12px;
}

.focus-region__meta span {
  padding: 4px 10px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.05);
}

.insight-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
}

.insight-card {
  padding: 16px;
  border-radius: 18px;
  background: rgba(9, 35, 64, 0.72);
}

.insight-card__label {
  color: rgba(214, 233, 255, 0.58);
  font-size: 12px;
}

.insight-card__value {
  margin-top: 10px;
  color: #ffffff;
  font-family: var(--screen-number-font);
  font-size: 28px;
  font-weight: 700;
  line-height: 1;
}

.insight-card__desc {
  margin-top: 12px;
  color: rgba(214, 233, 255, 0.62);
  font-size: 12px;
  line-height: 18px;
}

.insight-card--cyan {
  box-shadow: inset 0 0 0 1px rgba(101, 220, 255, 0.12);
}

.insight-card--amber {
  box-shadow: inset 0 0 0 1px rgba(255, 200, 106, 0.12);
}

.insight-card--violet {
  box-shadow: inset 0 0 0 1px rgba(126, 147, 255, 0.12);
}

.insight-card--green {
  box-shadow: inset 0 0 0 1px rgba(84, 229, 166, 0.12);
}

.panel__note {
  margin-top: 16px;
  color: rgba(214, 233, 255, 0.5);
  font-size: 12px;
  line-height: 18px;
}

.panel--regions {
  min-height: 320px;
}

.region-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.region-card {
  padding: 18px;
  border-radius: 20px;
  background:
    linear-gradient(180deg, rgba(10, 38, 68, 0.8) 0%, rgba(7, 24, 44, 0.92) 100%);
  box-shadow: inset 0 0 0 1px rgba(124, 196, 255, 0.08);
}

.region-card__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.region-card__rank {
  color: rgba(132, 216, 255, 0.74);
  font-size: 12px;
  letter-spacing: 0.12em;
}

.region-card__name {
  margin-top: 8px;
  color: #ffffff;
  font-size: 24px;
  font-weight: 700;
  line-height: 1.2;
}

.region-card__tag {
  padding: 6px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.06);
  color: #9ee9ff;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.region-card__stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
  margin-top: 18px;
}

.region-stat {
  padding: 12px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.04);
  text-align: center;
}

.region-stat span {
  display: block;
  color: rgba(214, 233, 255, 0.58);
  font-size: 12px;
}

.region-stat strong {
  display: block;
  margin-top: 8px;
  color: #ffffff;
  font-family: var(--screen-number-font);
  font-size: 24px;
  font-weight: 700;
}

.region-bars {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-top: 18px;
}

.region-bar {
  display: grid;
  grid-template-columns: 64px minmax(0, 1fr) 42px;
  align-items: center;
  gap: 10px;
}

.region-bar span,
.region-bar em {
  color: rgba(214, 233, 255, 0.62);
  font-size: 12px;
  font-style: normal;
}

.region-bar__track {
  height: 8px;
  overflow: hidden;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.06);
}

.region-bar__fill {
  height: 100%;
  border-radius: 999px;
}

.region-bar__fill--blue {
  background: linear-gradient(90deg, #64b5ff 0%, #6a8cff 100%);
}

.region-bar__fill--cyan {
  background: linear-gradient(90deg, #74f0ff 0%, #4ca1ff 100%);
}

.region-bar__fill--amber {
  background: linear-gradient(90deg, #ffc86a 0%, #ff9f5a 100%);
}

.empty-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 220px;
  border-radius: 20px;
  background: rgba(9, 35, 64, 0.42);
  color: rgba(214, 233, 255, 0.62);
  font-size: 14px;
}

@keyframes rotate-screen-ring {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

@keyframes pulse-screen-glow {
  0%,
  100% {
    box-shadow: inset 0 0 18px rgba(116, 240, 255, 0.08);
    opacity: 0.8;
  }
  50% {
    box-shadow: inset 0 0 26px rgba(116, 240, 255, 0.18);
    opacity: 1;
  }
}

@keyframes screen-scan {
  from {
    background-position: 140% 0;
  }
  to {
    background-position: -120% 0;
  }
}

@media (max-width: 1440px) {
  .metric-ribbon {
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  }

  .dashboard-grid {
    grid-template-columns: minmax(0, 1fr) 300px;
  }
}

@media (max-width: 1200px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
    grid-template-areas:
      'core'
      'identity'
      'insight';
  }

  .core-stage {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 960px) {
  .screen-page {
    padding: 16px;
    border-radius: 20px;
  }

  .hero-banner {
    grid-template-columns: 1fr;
    padding: 20px;
  }

  .hero-banner__title {
    font-size: 26px;
  }

  .hero-banner__meta {
    justify-content: flex-start;
  }

  .core-quick-metrics,
  .insight-grid,
  .region-card__stats {
    grid-template-columns: 1fr;
  }

  .identity-item__value {
    font-size: 16px;
  }

  .region-bar {
    grid-template-columns: 56px minmax(0, 1fr) 40px;
  }
}
</style>
