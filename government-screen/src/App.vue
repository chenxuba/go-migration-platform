<template>
  <main class="screen-shell">
    <div class="screen-frame" :style="screenStyle">
      <header class="screen-header">
        <div class="supervision-unit">
          <span class="unit-icon">盾</span>
          <span>监管单位：XXX市教育局</span>
        </div>
        <div class="title-block">
          <h1>康复教育机构监管驾驶舱</h1>
          <p>辖区机构运营 · 教学服务 · 资金安全 · 风险预警</p>
        </div>
        <div class="header-tools">
          <div class="clock">2026-04-25&nbsp;&nbsp;09:30:18</div>
          <div class="filters">
            <span>全市</span>
            <span>区县</span>
            <span>机构</span>
          </div>
          <div class="updated">数据更新时间：2026-04-25 09:25:12 <b>↻</b></div>
        </div>
      </header>

      <section class="kpi-grid">
        <article v-for="item in kpis" :key="item.label" class="kpi-card" :class="item.level">
          <div class="kpi-icon">{{ item.icon }}</div>
          <div>
            <p>{{ item.label }}</p>
            <strong>{{ item.value }}</strong>
            <span>{{ item.unit }}</span>
          </div>
        </article>
      </section>

      <section class="dashboard-grid">
        <aside class="left-column">
          <Panel title="机构运营监管" class="operation-panel">
            <div class="sub-title">机构活跃度排行 TOP5 <span>近30天</span></div>
            <div class="ranking-list">
              <div v-for="org in orgRanks" :key="org.name" class="rank-row">
                <i>{{ org.rank }}</i>
                <span>{{ org.name }}</span>
                <div class="bar"><b :style="{ width: `${org.score}%` }" /></div>
                <em>{{ org.score }}</em>
              </div>
            </div>
          </Panel>

          <Panel title="学员结构" class="student-panel">
            <div class="donut-grid">
              <div class="donut-card">
                <div class="donut age-donut"><span>6,842<br>总人数</span></div>
                <ul>
                  <li><i class="green" />0-3岁 811</li>
                  <li><i class="cyan" />4-6岁 2,456</li>
                  <li><i class="blue" />7-12岁 2,637</li>
                  <li><i class="purple" />13-18岁 938</li>
                </ul>
              </div>
              <div class="donut-card">
                <div class="donut type-donut" />
                <ul>
                  <li><i class="cyan" />语言 38.6%</li>
                  <li><i class="yellow" />智力 22.1%</li>
                  <li><i class="blue" />感统 12.8%</li>
                  <li><i class="green" />其他 26.5%</li>
                </ul>
              </div>
            </div>
            <div class="status-stack">
              <span class="active" /><span class="paused" /><span class="done" />
            </div>
            <div class="legend-row"><span>在训 5,215</span><span>停课 862</span><span>结课 765</span></div>
          </Panel>

          <Panel title="家长绑定与通知触达" class="touch-panel">
            <div class="metric-pair">
              <MiniMetric label="家长绑定率" value="88.7%" color="cyan" />
              <MiniMetric label="通知阅读率" value="76.4%" color="green" />
            </div>
            <svg class="sparkline" viewBox="0 0 310 70" aria-label="触达趋势">
              <polyline points="4,46 26,38 48,42 70,35 92,48 114,30 136,36 158,27 180,34 202,19 224,30 246,16 268,23 306,10" />
              <polyline class="green-line" points="4,54 26,50 48,44 70,49 92,38 114,42 136,31 158,36 180,25 202,29 224,21 246,26 268,17 306,14" />
            </svg>
          </Panel>
        </aside>

        <section class="center-column">
          <Panel title="辖区机构分布地图" class="map-panel">
            <template #extra>
              <select><option>全部机构</option></select>
            </template>
            <div class="map-legend">
              <span><i class="normal" />正常</span><span><i class="focus" />关注</span><span><i class="warn" />预警</span><span><i class="danger" />高风险</span>
            </div>
            <div class="map-canvas">
              <Map3D />
              <div class="map-hud-lines">
                <span class="scan-line" />
                <span class="scan-line reverse" />
              </div>
              <div class="tooltip-card">
                <strong>星启康复中心 <em>关注</em></strong>
                <p>今日课程 <b>46</b> 节</p>
                <p>到课率 <b>89%</b></p>
                <p class="danger-text">风险：课消异常</p>
                <button>查看详情</button>
              </div>
              <div class="risk-stat">
                <p>风险机构统计</p>
                <span><i class="danger-bg" />高风险 8家</span>
                <span><i class="warn-bg" />预警 16家</span>
                <span><i class="focus-bg" />关注 32家</span>
                <span><i class="normal-bg" />正常 72家</span>
                <b>机构总数 128家</b>
              </div>
            </div>
          </Panel>

          <Panel title="实时动态" class="feed-panel">
            <div class="feed-tabs"><span>排课</span><span>点名</span><span>收费</span><span>审批</span></div>
            <div class="feed-list">
              <div v-for="event in events" :key="event.time" class="feed-item" :class="event.type">
                <i>{{ event.icon }}</i>
                <strong>{{ event.org }}</strong>
                <span>{{ event.text }}</span>
                <em>{{ event.time }}</em>
              </div>
            </div>
          </Panel>
        </section>

        <aside class="right-column">
          <Panel title="教学服务监管" class="teaching-panel">
            <div class="panel-note">今日排课与到课趋势 <b>今日到课率 92.6%</b></div>
            <svg class="combo-chart" viewBox="0 0 450 170">
              <g class="bars">
                <rect v-for="(bar, index) in attendanceBars" :key="index" :x="18 + index * 25" :y="150 - bar" width="10" :height="bar" rx="3" />
              </g>
              <polyline points="18,72 43,88 68,60 93,50 118,74 143,78 168,64 193,82 218,96 243,112 268,56 293,86 318,74 343,92 368,78 393,96 418,88" />
            </svg>
          </Panel>

          <div class="right-two">
            <Panel title="老师授课负载 TOP5">
              <div class="teacher-list">
                <div v-for="teacher in teachers" :key="teacher.name">
                  <span>{{ teacher.name }}</span><b>{{ teacher.lessons }}</b><i :style="{ width: teacher.load }" /><em>{{ teacher.rate }}</em>
                </div>
              </div>
            </Panel>
            <Panel title="点名异常">
              <div class="abnormal-donut"><span>86<br>异常总数</span></div>
              <ul class="abnormal-list">
                <li><i class="blue" />未点名 28</li>
                <li><i class="orange" />迟到 22</li>
                <li><i class="green" />缺勤 18</li>
                <li><i class="yellow" />请假 18</li>
              </ul>
            </Panel>
          </div>

          <div class="right-two">
            <Panel title="康复记录完成率">
              <RingMetric value="86.3%" text1="已完成 8,965" text2="未完成 1,424" />
            </Panel>
            <Panel title="家长反馈率">
              <RingMetric value="74.1%" text1="已反馈 4,152" text2="待反馈 1,448" />
            </Panel>
          </div>
        </aside>
      </section>

      <section class="bottom-grid">
        <Panel title="资金与课消安全监管" class="money-panel">
          <div class="panel-note">资金趋势分析 <span>近30天 · 单位：万元</span></div>
          <svg class="money-chart" viewBox="0 0 620 190">
            <g class="grid-lines"><line v-for="line in 5" :key="line" x1="0" :y1="line * 34" x2="620" :y2="line * 34" /></g>
            <polyline class="income" points="8,116 40,88 72,108 104,72 136,68 168,98 200,76 232,84 264,58 296,66 328,92 360,54 392,70 424,48 456,62 488,38 520,58 552,36 612,18" />
            <polyline class="consume" points="8,136 40,120 72,128 104,104 136,112 168,92 200,100 232,78 264,84 296,72 328,82 360,64 392,74 424,52 456,60 488,42 520,54 552,34 612,28" />
            <polyline class="recharge" points="8,160 40,148 72,154 104,132 136,140 168,126 200,136 232,112 264,124 296,116 328,132 360,108 392,118 424,96 456,106 488,88 520,98 552,74 612,56" />
          </svg>
        </Panel>

        <Panel title="欠费预警 TOP5" class="arrears-panel">
          <table>
            <thead><tr><th>机构名称</th><th>学员数</th><th>欠费金额</th><th>风险</th></tr></thead>
            <tbody>
              <tr v-for="row in arrears" :key="row.name"><td>{{ row.name }}</td><td>{{ row.students }}</td><td>{{ row.amount }}</td><td><span :class="['risk-tag', row.level]">{{ row.text }}</span></td></tr>
            </tbody>
          </table>
        </Panel>

        <Panel title="待续费提醒" class="renew-panel">
          <div class="renew-cards">
            <div><b>7天内</b><strong>156人</strong><span>¥128,560</span></div>
            <div><b>15天内</b><strong>342人</strong><span>¥285,420</span></div>
            <div><b>30天内</b><strong>689人</strong><span>¥568,730</span></div>
          </div>
          <p>提醒规则：以合同到期日为准</p>
        </Panel>

        <Panel title="审批流监控" class="approval-panel">
          <div class="approval-tabs"><span>请假</span><span>停复课</span><span>退费</span><span>关账</span></div>
          <div class="flow">
            <div><i>人</i><span>发起申请</span><b>38</b></div>
            <em />
            <div><i>审</i><span>机构审批</span><b>27</b><small>超时3</small></div>
            <em />
            <div><i>监</i><span>监管审批</span><b>19</b><small>超时2</small></div>
            <em />
            <div><i>✓</i><span>完成</span><b>16</b></div>
          </div>
          <p class="overdue">超时审批 5 项</p>
        </Panel>
      </section>
    </div>
  </main>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import Panel from './components/Panel.vue'
import Map3D from './components/Map3D.vue'
import MiniMetric from './components/MiniMetric.vue'
import RingMetric from './components/RingMetric.vue'

const viewport = ref({ width: window.innerWidth, height: window.innerHeight })
const updateViewport = () => {
  viewport.value = { width: window.innerWidth, height: window.innerHeight }
}
onMounted(() => window.addEventListener('resize', updateViewport))
onBeforeUnmount(() => window.removeEventListener('resize', updateViewport))
const screenStyle = computed(() => ({
  transform: `scale(${Math.min(viewport.value.width / 1920, viewport.value.height / 1080)})`,
}))

const kpis = [
  { label: '监管机构数', value: '128', unit: '家', icon: '楼', level: 'blue' },
  { label: '在册学员', value: '6,842', unit: '人', icon: '员', level: 'cyan' },
  { label: '今日排课', value: '1,236', unit: '节', icon: '课', level: 'blue' },
  { label: '今日到课率', value: '92.6', unit: '%', icon: '趋', level: 'cyan' },
  { label: '待处理审批', value: '37', unit: '项', icon: '审', level: 'orange' },
  { label: '资金监管余额', value: '¥18,560,000', unit: '', icon: '¥', level: 'blue' },
]

const orgRanks = [
  { rank: 1, name: '星启康复中心', score: 96.5 },
  { rank: 2, name: '启智康复中心', score: 92.1 },
  { rank: 3, name: '阳光康复中心', score: 88.7 },
  { rank: 4, name: '未来星康复中心', score: 86.3 },
  { rank: 5, name: '希望之家', score: 82.4 },
]


const events = [
  { org: '星启康复中心', text: '新增排课 12节', time: '09:29:45', icon: '课', type: 'course' },
  { org: '启智康复中心', text: '完成点名 28人', time: '09:29:12', icon: '点', type: 'check' },
  { org: '阳光康复中心', text: '收费成功 ¥3,560', time: '09:28:31', icon: '收', type: 'money' },
  { org: '未来星康复中心', text: '退费申请待审批', time: '09:27:58', icon: '审', type: 'approval' },
  { org: '希望之家', text: '停复课申请待审批', time: '09:27:11', icon: '审', type: 'approval' },
]

const attendanceBars = [18, 22, 28, 34, 40, 46, 58, 72, 96, 134, 152, 110, 92, 76, 88, 94, 82]
const teachers = [
  { name: '张老师', lessons: 18, load: '90%', rate: '90%' },
  { name: '李老师', lessons: 16, load: '80%', rate: '80%' },
  { name: '王老师', lessons: 15, load: '75%', rate: '75%' },
  { name: '陈老师', lessons: 14, load: '72%', rate: '72%' },
  { name: '刘老师', lessons: 12, load: '60%', rate: '60%' },
]
const arrears = [
  { name: '阳光康复中心', students: 68, amount: '¥128,600', level: 'high', text: '高风险' },
  { name: '未来星康复中心', students: 55, amount: '¥96,320', level: 'high', text: '高风险' },
  { name: '启智康复中心', students: 42, amount: '¥72,300', level: 'middle', text: '中风险' },
  { name: '星启康复中心', students: 38, amount: '¥58,400', level: 'middle', text: '中风险' },
  { name: '希望之家', students: 31, amount: '¥46,200', level: 'low', text: '关注' },
]
</script>
