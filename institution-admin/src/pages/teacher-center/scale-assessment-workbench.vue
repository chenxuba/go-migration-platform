<script setup lang="ts">
import {
  ArrowLeftOutlined,
  CheckCircleFilled,
  FileDoneOutlined,
  FileTextOutlined,
  InfoCircleOutlined,
  LeftOutlined,
  RightOutlined,
  SaveOutlined,
  SlidersOutlined,
  SwapOutlined,
} from '@ant-design/icons-vue'
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const selectedScore = ref(1)
const autoNext = ref(true)

const studentName = computed(() => String(route.query.childName || '浩浩'))
const studentAge = computed(() => String(route.query.childAge || '3;2'))
const scaleTitle = computed(() => {
  const code = String(route.query.scaleCode || 'PEP3').trim()
  if (code.toUpperCase() === 'PEP3')
    return 'PEP-3'
  return code || 'PEP-3'
})

const pageGroups = [
  { title: '第 1 页  准备页', count: '0/10 题', percent: 0, expanded: false, items: [] },
  {
    title: '第 2 页  1-14题',
    count: '10/14 题',
    percent: 71,
    expanded: true,
    items: [
      { no: 8, name: '模仿动作', status: 'done' },
      { no: 9, name: '身体部位指认', status: 'done' },
      { no: 10, name: '图片命名', status: 'done' },
      { no: 11, name: '动作指令', status: 'done' },
      { no: 12, name: '物件配对', status: 'active' },
      { no: 13, name: '图片分类', status: 'todo' },
      { no: 14, name: '物品功能', status: 'todo' },
    ],
  },
  { title: '第 3 页  15-28题', count: '3/14 题', percent: 21, expanded: false, items: [] },
  { title: '第 4 页  29-42题', count: '0/14 题', percent: 0, expanded: false, items: [] },
  { title: '第 5 页  43-56题', count: '0/14 题', percent: 0, expanded: false, items: [] },
]

const rawScores = [
  ['CVP 认知（语言/语前）', 28],
  ['EXP 表达性语言', 24],
  ['RCP 接受性语言', 26],
  ['VSI 视觉感知', 21],
  ['Fine 精细动作', 18],
  ['GMI 粗大动作', 17],
  ['SOC 社会交往', 22],
  ['SED 社会情绪发展', 16],
  ['Self 自理技能', 14],
]

const scoreOptions = [
  { value: 2, title: '2 分', desc: '通过', tone: 'green', checkColor: '#0d9749' },
  { value: 1, title: '1 分', desc: '部分通过', tone: 'blue', checkColor: '#0757e6' },
  { value: 0, title: '0 分', desc: '未通过', tone: 'red', checkColor: '#d41f1f' },
]

function goBack() {
  void router.push('/teacherCenter/scale-library')
}
</script>

<template>
  <div class="pep3-workbench-page">
    <header class="workbench-header">
      <button type="button" class="back-button" aria-label="返回量表库" @click="goBack">
        <ArrowLeftOutlined />
      </button>
      <strong class="workbench-title">{{ scaleTitle }} PC测评工作台</strong>
      <span class="header-divider"></span>
      <span class="header-meta">儿童：<b>{{ studentName }}</b></span>
      <span class="header-divider"></span>
      <span class="header-meta">年龄：<b>{{ studentAge }}</b></span>
      <span class="header-divider"></span>
      <span class="header-meta">测评日期：<b>2025-05-20</b></span>
      <span class="header-divider"></span>
      <span class="header-meta">施测者：<b>张老师</b></span>
      <div class="header-actions">
        <a-button size="large" class="outline-action">
          <template #icon>
            <SaveOutlined />
          </template>
          保存草稿
        </a-button>
        <a-button size="large" type="primary" class="primary-action">
          <template #icon>
            <FileDoneOutlined />
          </template>
          提交记录
        </a-button>
      </div>
    </header>

    <main class="workbench-main">
      <aside class="page-sidebar">
        <div class="sidebar-title">
          <span>记录册页面</span>
          <SlidersOutlined />
        </div>
        <div v-for="group in pageGroups" :key="group.title" class="page-group">
          <div class="page-group__head">
            <span class="page-group__title">
              <RightOutlined v-if="!group.expanded" />
              <span v-else class="chevron-down">⌄</span>
              {{ group.title }}
            </span>
            <span>{{ group.count }}</span>
          </div>
          <div class="page-group__progress">
            <div class="progress-line">
              <i :style="{ width: `${group.percent}%` }"></i>
            </div>
            <span class="page-group__percent">{{ group.percent }}%</span>
          </div>

          <div v-if="group.expanded" class="question-list">
            <button
              v-for="item in group.items"
              :key="item.no"
              type="button"
              class="question-item"
              :class="`is-${item.status}`"
            >
              <span>第 {{ item.no }} 题</span>
              <strong>{{ item.name }}</strong>
              <CheckCircleFilled v-if="item.status === 'done'" />
              <i v-else-if="item.status === 'active'"></i>
              <b v-else></b>
            </button>
          </div>
        </div>
      </aside>

      <section class="question-panel">
        <div class="question-title-row">
          <h1>第 1 题&nbsp;&nbsp;拧开瓶盖</h1>
          <a-tag color="blue">CVP 认知（语言/语前）</a-tag>
        </div>

        <article class="instruction-card">
          <h2><FileTextOutlined />材料</h2>
          <p>泡泡瓶</p>
        </article>

        <article class="instruction-card">
          <h2><FileTextOutlined />操作标准</h2>
          <p>测试员将泡泡瓶放在桌子上，并示意儿童打开瓶盖。若儿童无法完成，测试员可进行示范后再让儿童尝试。</p>
        </article>

        <article class="instruction-card">
          <h2><FileTextOutlined />指导语</h2>
          <p>把泡泡瓶盖打开，我们来吹泡泡。</p>
        </article>

        <article class="instruction-card">
          <h2><FileTextOutlined />评分标准</h2>
          <p><b>2 分（通过）：</b> 能自行拧开瓶盖。</p>
          <p><b>1 分（部分通过）：</b> 未能拧开瓶盖，但做出所需动作，即把手放在瓶盖上并做出拧动动作。</p>
          <p><b>0 分（未通过）：</b> 未能拧开瓶盖或做出所需动作。</p>
        </article>

        <div class="score-section">
          <h2>评分</h2>
          <div class="score-options">
            <button
              v-for="item in scoreOptions"
              :key="item.value"
              type="button"
              class="score-option"
              :class="[`score-${item.tone}`, { 'is-selected': selectedScore === item.value }]"
              @click="selectedScore = item.value"
            >
              <strong>{{ item.title }}</strong>
              <span>{{ item.desc }}</span>
              <CheckCircleFilled
                v-if="selectedScore === item.value"
                class="score-option__check"
                :style="{ color: item.checkColor }"
              />
            </button>
          </div>
        </div>
      </section>

      <aside class="score-sidebar">
        <section class="right-card progress-card">
          <h2>当前进度</h2>
          <div class="progress-card__body">
            <div class="donut">68%</div>
            <div class="progress-stats">
              <span>已完成</span>
              <strong>117 <i>/ 172 题</i></strong>
              <span>缺题</span>
              <strong class="danger">9 <i>题</i></strong>
            </div>
          </div>
        </section>

        <section class="right-card caregiver-card">
          <h2>照护者报告原始分 <InfoCircleOutlined /></h2>
          <label><span>PB（行为问题）</span><em>-</em></label>
          <label><span>PSC（问题严重性）</span><em>-</em></label>
          <label><span>AB（适应行为）</span><em>-</em></label>
        </section>

        <section class="right-card raw-score-card">
          <h2>自动汇总原始分 <span>（已完成部分）</span></h2>
          <div v-for="item in rawScores" :key="item[0]" class="raw-score-row">
            <span>{{ item[0] }}</span>
            <strong>{{ item[1] }}</strong>
          </div>
        </section>
      </aside>
    </main>

    <footer class="workbench-footer">
      <a-button size="large" class="nav-button">
        <template #icon>
          <LeftOutlined />
        </template>
        上一题
      </a-button>
      <div class="question-counter">
        <strong>1</strong>
        <span>/ 172</span>
      </div>
      <a-button size="large" type="primary" class="next-button">
        下一题
        <RightOutlined />
      </a-button>
      <a-button size="large" class="nav-button">
        <template #icon>
          <SwapOutlined />
        </template>
        跳到缺题
      </a-button>
      <div class="auto-next">
        <span>自动下一题</span>
        <a-switch v-model:checked="autoNext" />
      </div>
    </footer>
  </div>
</template>

<style scoped lang="less">
.pep3-workbench-page {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  margin: 0;
  color: #1f2937;
  background: #f3f5f9;
}

.workbench-header {
  position: sticky;
  top: 0;
  z-index: 30;
  display: flex;
  align-items: center;
  min-height: 52px;
  padding: 0 14px;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid #d8dfe8;
  border-radius: 0 0 10px 10px;
  box-shadow: 0 8px 22px rgba(15, 23, 42, 0.08);
  backdrop-filter: blur(8px);
}

.back-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  margin-right: 12px;
  color: #0f2a5f;
  background: transparent;
  border: 0;
  border-radius: 6px;
  cursor: pointer;
  font-size: 18px;

  &:hover {
    background: #eef4ff;
    color: #155bdc;
  }
}

.workbench-title {
  color: #0f2a5f;
  font-size: 18px;
  font-weight: 800;
  white-space: nowrap;
}

.header-divider {
  width: 1px;
  height: 18px;
  margin: 0 12px;
  background: #cbd5e1;
}

.header-meta {
  color: #111827;
  font-size: 13px;
  white-space: nowrap;

  b {
    font-weight: 700;
  }
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
}

.outline-action,
.primary-action {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 7px;
  min-width: 104px;
  height: 32px;
  padding: 0 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 700;
  line-height: 1;

  :deep(.ant-btn-icon) {
    display: inline-flex;
    align-items: center;
    margin-inline-end: 0;
    line-height: 1;
  }

  :deep(.anticon) {
    display: inline-flex;
    align-items: center;
    line-height: 1;
  }

  :deep(.anticon + span) {
    margin-inline-start: 0;
  }
}

.outline-action {
  color: #155bdc;
  border-color: #2f6bff;
}

.primary-action {
  background: #0757e6;
  box-shadow: 0 10px 20px rgba(7, 87, 230, 0.24);
}

.workbench-main {
  display: grid;
  grid-template-columns: 240px minmax(420px, 1fr) 256px;
  gap: 10px;
  flex: 1;
  min-height: 0;
  padding: 10px 10px 0;
}

.page-sidebar,
.question-panel,
.right-card {
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid #e1e7f0;
  border-radius: 8px;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.06);
}

.page-sidebar {
  overflow: hidden;
}

.sidebar-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 34px;
  padding: 0 16px;
  border-bottom: 1px solid #e5eaf1;
  font-size: 13px;
  font-weight: 700;
}

.page-group {
  padding: 10px 12px 8px;
  border-bottom: 1px solid #e5eaf1;
}

.page-group__head {
  display: flex;
  justify-content: space-between;
  color: #667085;
  font-size: 12px;
}

.page-group__title {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  color: #111827;
  font-weight: 700;
}

.chevron-down {
  margin-top: -6px;
  font-size: 18px;
  line-height: 12px;
}

.page-group__progress {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 38px;
  align-items: center;
  gap: 10px;
  margin: 8px 0 2px 18px;
}

.progress-line {
  height: 5px;
  overflow: hidden;
  background: #e5e7eb;
  border-radius: 999px;

  i {
    display: block;
    height: 100%;
    background: #18a957;
    border-radius: inherit;
  }
}

.page-group__percent {
  color: #667085;
  text-align: right;
  font-size: 12px;
  white-space: nowrap;
}

.question-list {
  margin-top: 6px;
}

.question-item {
  display: grid;
  grid-template-columns: 50px minmax(0, 1fr) 16px;
  align-items: center;
  width: 100%;
  min-height: 30px;
  padding: 0 8px 0 26px;
  color: #4b5563;
  background: transparent;
  border: 0;
  border-radius: 0;
  cursor: pointer;
  font-size: 12px;
  text-align: left;

  strong {
    color: inherit;
  }

  .anticon {
    color: #18a957;
    font-size: 14px;
  }

  i,
  b {
    display: inline-block;
    width: 14px;
    height: 14px;
    border-radius: 50%;
  }

  i {
    background: radial-gradient(circle at center, #fff 19%, #1769e8 22% 100%);
  }

  b {
    border: 1px solid #b8c1cf;
  }

  &.is-active {
    position: relative;
    color: #0757e6;
    background: #eaf3ff;
    font-weight: 800;

    &::before {
      position: absolute;
      left: 0;
      width: 4px;
      height: 100%;
      background: #0757e6;
      border-radius: 0 4px 4px 0;
      content: "";
    }
  }
}

.question-panel {
  padding: 14px 16px 12px;
}

.question-title-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;

  h1 {
    margin: 0;
    color: #111827;
    font-size: 20px;
    font-weight: 900;
  }

  :deep(.ant-tag) {
    padding: 2px 8px;
    border-radius: 7px;
    font-size: 12px;
  }
}

.instruction-card {
  padding: 11px 13px;
  margin-bottom: 8px;
  background: #fff;
  border: 1px solid #d8e0eb;
  border-radius: 8px;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.06);

  h2 {
    display: flex;
    align-items: center;
    gap: 6px;
    margin: 0 0 7px;
    color: #263247;
    font-size: 14px;
    font-weight: 700;

    .anticon {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 14px;
      height: 14px;
      color: #155bdc;
      background: #edf4ff;
      border: 1px solid #c8dcff;
      border-radius: 4px;
      font-size: 9px;
    }
  }

  ul,
  ol {
    padding-left: 20px;
    margin: 0;
  }

  li,
  p {
    margin: 4px 0;
    color: #3f4856;
    font-size: 13px;
    line-height: 1.45;
  }
}

.score-section {
  h2 {
    margin: 6px 0;
    font-size: 13px;
  }
}

.score-options {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

.score-option {
  --score-color: #0757e6;
  --score-hover-bg: #f8fbff;
  --score-selected-bg: #f4f8ff;
  --score-check-bg: #0757e6;

  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: center;
  min-height: 64px;
  padding: 10px 42px 10px 16px;
  color: #111827;
  text-align: left;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
  transition:
    background 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;

  &::after {
    position: absolute;
    top: 13px;
    right: 14px;
    width: 15px;
    height: 15px;
    background: #fff;
    border: 1px solid #cbd5e1;
    border-radius: 50%;
    content: "";
  }

  &:hover {
    background: var(--score-hover-bg);
    border-color: var(--score-color);
  }

  strong,
  span {
    display: block;
  }

  strong {
    color: var(--score-color);
    font-size: 19px;
    font-weight: 700;
    line-height: 1.2;
  }

  span {
    margin-top: 4px;
    color: #334155;
    font-size: 13px;
    font-weight: 500;
  }

  &.score-green {
    --score-color: #0d9749;
    --score-hover-bg: #f7fff9;
    --score-selected-bg: #f6fff9;
    --score-check-bg: #0d9749;
  }

  &.score-blue {
    --score-color: #0757e6;
    --score-hover-bg: #f7faff;
    --score-selected-bg: #f7faff;
    --score-check-bg: #0757e6;
  }

  &.score-red {
    --score-color: #d41f1f;
    --score-hover-bg: #fff8f7;
    --score-selected-bg: #fff8f7;
    --score-check-bg: #d41f1f;
  }

  &.is-selected {
    background: var(--score-selected-bg);
    border-color: var(--score-color);
    box-shadow: none;

    &::after {
      opacity: 0;
    }
  }
}

.score-option__check {
  position: absolute;
  top: 11px;
  right: 13px;
  color: var(--score-check-bg);
  font-size: 17px;

  :deep(svg) {
    color: inherit;
    fill: currentcolor;
  }
}

.score-sidebar {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.right-card {
  padding: 11px 12px 10px;

  h2 {
    margin: 0 0 8px;
    font-size: 13px;
    font-weight: 800;
  }
}

.progress-card__body {
  display: grid;
  grid-template-columns: 88px 1fr;
  align-items: center;
  gap: 10px;
}

.donut {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 82px;
  height: 82px;
  color: #1f2a44;
  background:
    radial-gradient(circle at center, #fff 54%, transparent 55%),
    conic-gradient(#2563eb 0 68%, #e5e7eb 68% 100%);
  border-radius: 50%;
  font-size: 20px;
  font-weight: 700;
}

.progress-stats {
  display: grid;
  gap: 4px;
  color: #667085;
  font-size: 12px;

  strong {
    color: #0757e6;
    font-size: 15px;

    i {
      color: #4b5563;
      font-size: 12px;
      font-style: normal;
      font-weight: 500;
    }
  }

  .danger {
    color: #f04438;
  }
}

.caregiver-card {
  label {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 78px;
    align-items: center;
    gap: 8px;
    margin-top: 6px;
    color: #374151;
    font-size: 12px;

    em {
      height: 26px;
      color: #374151;
      background: #fbfdff;
      border: 1px solid #dbe3ed;
      border-radius: 6px;
      font-style: normal;
      line-height: 24px;
      text-align: center;
    }
  }
}

.raw-score-card {
  padding: 14px 16px;

  h2 {
    margin-bottom: 10px;
  }

  h2 span {
    color: #6b7280;
    font-size: 12px;
    font-weight: 500;
  }
}

.raw-score-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 34px;
  align-items: center;
  column-gap: 12px;
  min-height: 28px;
  color: #374151;
  border-bottom: 1px solid #edf1f6;
  font-size: 13px;

  span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  strong {
    color: #0b8f43;
    font-size: 14px;
    text-align: right;
  }
}

.workbench-footer {
  position: sticky;
  bottom: 0;
  z-index: 11;
  display: grid;
  grid-template-columns: 136px 1fr 146px 158px 142px;
  align-items: center;
  gap: 16px;
  min-height: 58px;
  padding: 0 20px;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid #d8dfe8;
  border-radius: 10px 10px 0 0;
  box-shadow: 0 -8px 24px rgba(15, 23, 42, 0.08);
}

.nav-button,
.next-button {
  height: 34px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 800;
}

.nav-button {
  color: #0757e6;
  border-color: #9bbcff;
}

.next-button {
  background: #0757e6;
  box-shadow: 0 10px 20px rgba(7, 87, 230, 0.22);
}

.question-counter {
  text-align: center;

  strong {
    color: #1f2937;
    font-size: 24px;
    letter-spacing: 0;
  }

  span {
    margin-left: 6px;
    color: #1f2937;
    font-size: 14px;
  }
}

.auto-next {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  color: #374151;
  font-size: 12px;
}

@media (max-width: 1400px) {
  .workbench-main {
    grid-template-columns: 220px minmax(400px, 1fr) 240px;
  }

  .header-divider {
    margin: 0 16px;
  }

  .header-meta {
    font-size: 14px;
  }

  .workbench-footer {
    gap: 24px;
  }
}
</style>
