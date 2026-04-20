<script setup lang="ts">
import { computed, ref } from 'vue'
import { regionalSummary, scopeMap } from '../shared/mock-data'

const levelOptions = [
  { label: '省级视角', value: '省级' },
  { label: '市级视角', value: '市级' },
  { label: '区级视角', value: '区级' },
]

const currentLevel = ref<'省级' | '市级' | '区级'>('省级')

const currentScope = computed(() => scopeMap[currentLevel.value])

const statCards = computed(() => [
  {
    key: 'institutions',
    label: '纳管机构',
    value: `${currentScope.value.institutionCount}`,
    unit: '家',
    tone: 'blue',
  },
  {
    key: 'regions',
    label: '下辖区域',
    value: `${currentScope.value.subordinateRegions}`,
    unit: '个',
    tone: 'cyan',
  },
  {
    key: 'alerts',
    label: '待处理预警',
    value: `${currentScope.value.pendingAlerts}`,
    unit: '条',
    tone: 'orange',
  },
  {
    key: 'tasks',
    label: '待办任务',
    value: `${currentScope.value.pendingTasks}`,
    unit: '项',
    tone: 'green',
  },
])

const columns = [
  { title: '区域名称', dataIndex: 'regionName', key: 'regionName' },
  { title: '层级', dataIndex: 'level', key: 'level', width: 100 },
  { title: '机构数', dataIndex: 'institutionCount', key: 'institutionCount', width: 100 },
  { title: '在办任务', dataIndex: 'ongoingTasks', key: 'ongoingTasks', width: 100 },
  { title: '风险预警', dataIndex: 'alerts', key: 'alerts', width: 100 },
  { title: '整改完成率', dataIndex: 'completionRate', key: 'completionRate', width: 120 },
]
</script>

<template>
  <div class="gov-page">
    <section class="hero-card">
      <div>
        <div class="hero-card__eyebrow">
          G 端 / 监管驾驶舱
        </div>
        <h1 class="hero-card__title">
          康复机构监管平台
        </h1>
        <p class="hero-card__desc">
          省、市、区三级共用一套代码，后端通过行政区划编码和数据权限控制监管范围。当前这套前端已经独立拆出，可继续承接统计总览、机构监管、督导整改和账号管理。
        </p>
      </div>

      <a-radio-group v-model:value="currentLevel" :options="levelOptions" option-type="button" button-style="solid" />
    </section>

    <section class="scope-strip">
      <div class="scope-strip__item">
        <span class="scope-strip__label">当前层级</span>
        <span class="scope-strip__value">{{ currentScope.level }}</span>
      </div>
      <div class="scope-strip__item">
        <span class="scope-strip__label">监管区域</span>
        <span class="scope-strip__value">{{ currentScope.regionName }}</span>
      </div>
      <div class="scope-strip__item">
        <span class="scope-strip__label">行政区划码</span>
        <span class="scope-strip__value">{{ currentScope.regionCode }}</span>
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
        <a-table :columns="columns" :data-source="regionalSummary" :pagination="false" size="small" />
      </a-card>

      <a-card title="拆分说明" :bordered="false">
        <div class="note-list">
          <div class="note-list__item">
            现在已经从现有前端底座拆出一套独立的 `government-admin`，后续可以单独部署。
          </div>
          <div class="note-list__item">
            省、市、区不需要拆三套代码，继续共用这一个监管端即可。
          </div>
          <div class="note-list__item">
            后续只要补齐监管菜单、接口鉴权和区域数据权限，就能逐步落地真实业务。
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
}

.hero-card {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  padding: 24px;
  border-radius: 16px;
  background: linear-gradient(135deg, #e9f2ff 0%, #f7fbff 54%, #ffffff 100%);
}

.hero-card__eyebrow {
  margin-bottom: 10px;
  color: #1d4ed8;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.hero-card__title {
  margin: 0 0 10px;
  color: #102a43;
  font-size: 28px;
  font-weight: 700;
}

.hero-card__desc {
  margin: 0;
  max-width: 760px;
  color: #52606d;
  line-height: 24px;
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

.stat-card--orange {
  box-shadow: inset 0 0 0 1px rgba(249, 115, 22, 0.12);
}

.stat-card--green {
  box-shadow: inset 0 0 0 1px rgba(16, 185, 129, 0.12);
}

.content-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.6fr) minmax(300px, 1fr);
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
  .scope-strip,
  .stat-grid,
  .content-grid {
    grid-template-columns: 1fr;
  }

  .hero-card {
    flex-direction: column;
  }
}
</style>
