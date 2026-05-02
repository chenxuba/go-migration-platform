<script setup>
import {
  CloseOutlined,
  CopyOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  PlusOutlined,
  ReloadOutlined,
} from '@ant-design/icons-vue'
import { computed, ref } from 'vue'
import { downloadPEP3IEPPlanWordApi } from '@/api/edu-center/pep3-assessment'
import messageService from '@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  record: {
    type: Object,
    default: () => ({}),
  },
})

const emit = defineEmits(['update:open'])

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const planDuration = ref('3')
const activeDomainKey = ref('language')
const previewOpen = ref(false)
const previewGoal = ref(null)
const previewStage = ref(null)
const exportingWord = ref(false)

const domains = [
  { key: 'language', name: '语言沟通', icon: '语', longCount: 3, shortCount: 6 },
  { key: 'social', name: '社交互动', icon: '社', longCount: 2, shortCount: 5 },
  { key: 'cognition', name: '认知理解', icon: '认', longCount: 2, shortCount: 4 },
  { key: 'fineMotor', name: '精细动作', icon: '精', longCount: 2, shortCount: 4 },
  { key: 'sensory', name: '感统运动', icon: '感', longCount: 2, shortCount: 5 },
  { key: 'selfCare', name: '生活自理', icon: '自', longCount: 1, shortCount: 4 },
]

const iepData = {
  language: {
    name: '语言沟通',
    longGoals: [
      '在自然情境中主动表达需求，能使用2-3词短句完成请求、拒绝和简单回应。',
      '提升语言模仿与功能性表达能力，减少仅用动作或哭闹表达需求的情况。',
      '将课堂训练中的表达方式泛化到家庭、集体活动和日常生活场景。',
    ],
    homePlan: [
      '家庭训练围绕日常高频场景开展，包括进餐、玩具选择、穿衣、外出准备和睡前互动。家长在孩子出现需求前先等待3-5秒，鼓励其使用单词或2词短句表达，再给予物品或活动强化。',
      '每天安排2-3次短时训练，每次10-15分钟，重点练习请求、拒绝、选择和简单问答。训练时减少直接替孩子表达，优先使用实物提示、图片提示和语言示范，并逐步降低提示强度。',
      '家庭成员保持一致回应方式，当孩子用哭闹、拉人或指物替代表达时，先引导其完成可接受表达，再满足需求；如出现明显抗拒，可降低难度并回到单词表达。',
    ],
  },
  social: {
    name: '社交互动',
    longGoals: [
      '提升共同注意和互动回应能力，能在课堂活动中完成简单轮替互动。',
      '在成人提示下参与同伴活动，逐步增加主动发起和等待轮到自己的行为。',
    ],
    homePlan: [
      '家庭互动以轮流游戏、共同阅读和生活等待场景为主，家长每次只设置一个清晰互动目标。',
      '家长先示范等待、回应和轮替，再邀请孩子模仿，出现回应后立即给予具体表扬和活动强化。',
      '若孩子回避互动，可缩短互动时间并降低要求，先稳定完成目光回应或动作轮替。',
    ],
  },
  cognition: {
    name: '认知理解',
    longGoals: [
      '提升物品功能、类别配对和两步指令理解能力。',
      '能在课堂和家庭情境中根据口头指令完成简单任务。',
    ],
    homePlan: [
      '家庭训练结合收纳、取物和整理玩具进行，重点练习分类、配对和简单指令。',
      '家长每次给出简短明确指令，必要时配合手势提示，完成后逐步减少辅助。',
      '训练中保持材料少量、目标明确，避免一次呈现过多干扰物。',
    ],
  },
  fineMotor: {
    name: '精细动作',
    longGoals: [
      '提升手眼协调、抓握控制和双手协作能力。',
      '能完成穿珠、描线、夹取等精细动作任务并迁移到生活操作。',
    ],
    homePlan: [
      '家庭训练可使用夹子、积木、穿珠和粗蜡笔，每次选择一种材料进行短时练习。',
      '家长先示范动作，再提供手势或轻微身体辅助，重点观察抓握稳定性和持续时间。',
      '训练结束后安排简单生活应用，如整理餐具、贴贴纸或打开小盒子。',
    ],
  },
  sensory: {
    name: '感统运动',
    longGoals: [
      '提升前庭、本体和动作计划能力，增强课堂坐姿稳定和活动参与。',
      '在规则明确的运动活动中完成等待、启动和停止。',
    ],
    homePlan: [
      '家庭干预可安排跳跃、推拉、爬行和障碍跨越等活动，控制强度并观察情绪状态。',
      '每次运动前说明规则，活动中使用开始、停止、再来一次等简单口令。',
      '如出现兴奋过高或抗拒，应降低强度，加入深压、抱枕推压或安静整理活动。',
    ],
  },
  selfCare: {
    name: '生活自理',
    longGoals: [
      '提升穿脱、进餐和如厕相关生活自理能力。',
    ],
    homePlan: [
      '家庭训练优先选择固定生活流程，如洗手、穿鞋、收拾餐具，保持步骤和口令一致。',
      '家长将任务拆成小步骤，先完成最后一步再逐步前移，降低挫败感。',
      '每次完成后给予具体反馈，避免长期替代完成。',
    ],
  },
}

const stagePlans = {
  3: [
    {
      month: '第1个月',
      title: '建立表达基础',
      tag: '3项目标',
      note: '从提示表达过渡到单词表达',
      tone: 'blue',
      goals: [
        {
          title: '用单词表达需求',
          content: [
            '使用偏好物进行请求表达',
            '通过图片卡选择想要的物品',
            '在等待情境中引导说出目标词',
          ],
          method: {
            method: 'DTT + 自然情境教学',
            desc: '先给予视觉提示，再逐步减少辅助',
            steps: '呈现实物 -> 等待反应 -> 语言提示 -> 即时强化',
            standard: '连续3天在80%机会中完成表达',
          },
        },
        {
          title: '模仿功能词：要、不要',
          content: [
            '进行口型和发音模仿',
            '练习“要”“不要”等功能词跟读',
            '在请求和拒绝场景中进行功能词使用',
          ],
          method: {
            method: '模仿训练',
            desc: '从单音节模仿过渡到功能词表达',
            steps: '示范发音 -> 跟读 -> 场景使用 -> 强化',
            standard: '10次机会中完成7次',
          },
        },
        {
          title: '二选一情境选择表达',
          collapsed: true,
        },
      ],
    },
    {
      month: '第2个月',
      title: '提升主动沟通',
      tag: '3项目标',
      note: '增加主动发起与简单问答',
      tone: 'green',
      goals: [
        {
          title: '主动使用2词短句提出请求',
          content: [
            '进行目标物命名',
            '练习“我要+物品”短句表达',
            '在课堂、游戏和点心场景中发起表达',
          ],
          method: {
            method: '功能性沟通训练',
            desc: '围绕真实需求设置表达机会',
            steps: '创设需求 -> 等待主动表达 -> 扩展短句 -> 强化',
            standard: '5次机会中完成4次',
          },
        },
        {
          title: '回应简单问句',
          content: [
            '识别“要什么”“在哪里”等问句类型',
            '使用实物或图片辅助完成回应',
            '逐步扩展回答内容并减少提示',
          ],
          method: {
            method: '问答轮替训练',
            desc: '从封闭式问题过渡到简单开放问题',
            steps: '提问 -> 等待 -> 提示 -> 回答 -> 反馈强化',
            standard: '简单问句正确回应率达到80%',
          },
        },
        {
          title: '小组活动中发起沟通',
          collapsed: true,
        },
      ],
    },
    {
      month: '第3个月',
      title: '泛化与稳定',
      tag: '3项目标',
      note: '迁移到家庭与集体场景',
      tone: 'orange',
      goals: [
        {
          title: '表达拒绝或帮助需求',
          content: [
            '练习“不要”“帮帮我”等功能表达',
            '在任务受阻时替代哭闹或拉人行为',
            '在课堂、家庭和游戏中进行场景泛化',
          ],
          method: {
            method: '情境化训练',
            desc: '在真实问题中练习功能性表达',
            steps: '设置困难 -> 等待表达 -> 给予选择 -> 反馈调整',
            standard: '跨2个场景稳定出现',
          },
        },
        {
          title: '家庭场景简单问答',
          content: [
            '完成吃饭、玩具、外出等主题问答',
            '使用短句回应家长提问',
            '通过家庭短视频进行老师复盘',
          ],
          method: {
            method: '家庭泛化训练',
            desc: '把课堂目标迁移到真实生活场景',
            steps: '家长提问 -> 等待回应 -> 必要提示 -> 记录表现',
            standard: '家庭连续1周有稳定记录',
          },
        },
        {
          title: '受挫时使用语言表达情绪',
          collapsed: true,
        },
      ],
    },
  ],
  6: [
    {
      month: '第1个月',
      title: '表达启动',
      tag: '3项目标',
      note: '建立请求表达和模仿基础',
      tone: 'blue',
    },
    {
      month: '第2个月',
      title: '词汇积累',
      tag: '4项目标',
      note: '增加功能词和常见名词',
      tone: 'cyan',
    },
    {
      month: '第3个月',
      title: '短句表达',
      tag: '4项目标',
      note: '形成2词短句请求',
      tone: 'green',
    },
    {
      month: '第4个月',
      title: '问答轮替',
      tag: '4项目标',
      note: '回应简单问句和轮替沟通',
      tone: 'purple',
    },
    {
      month: '第5个月',
      title: '场景泛化',
      tag: '4项目标',
      note: '迁移到家庭与集体活动',
      tone: 'orange',
    },
    {
      month: '第6个月',
      title: '稳定维持',
      tag: '3项目标',
      note: '稳定表达并减少替代行为',
      tone: 'red',
    },
  ],
}

const sixMonthCollapsedGoalTitles = [
  '等待情境中主动请求帮助',
  '功能词在生活场景中泛化',
  '受挫时使用语言表达情绪',
  '同伴轮替中发起简单沟通',
  '家庭外出场景主动表达',
  '稳定使用短句完成需求表达',
]

const currentDomain = computed(() => {
  return iepData[activeDomainKey.value] || iepData.language
})

const stages = computed(() => {
  const plan = stagePlans[planDuration.value] || stagePlans[3]
  if (planDuration.value === '3')
    return plan

  return plan.map((stage, index) => {
    const visibleGoals = stagePlans[3][index % 3].goals.filter(goal => !goal.collapsed)
    const collapsedGoal = {
      title: sixMonthCollapsedGoalTitles[index] || '更多短期目标',
      collapsed: true,
    }

    return {
      ...stage,
      tag: `${visibleGoals.length + 1}项目标`,
      goals: [...visibleGoals, collapsedGoal],
    }
  })
})

const totalShortGoalCount = computed(() => {
  return planDuration.value === '3' ? 25 : 48
})

const longGoalText = computed(() => {
  return currentDomain.value.longGoals.map((item, index) => `${index + 1}. ${item}`).join('\n')
})

const homePlanText = computed(() => {
  return currentDomain.value.homePlan.map((item, index) => `${index + 1}. ${item}`).join('\n')
})

const studentMeta = computed(() => {
  const name = props.record?.studentName || '张一鸣'
  const gender = props.record?.studentGender || '男'
  const age = formatAge(props.record) || '3岁1月'
  const date = formatDate(props.record?.assessmentDate) || '2026-05-02'
  return `${name} · ${gender} · ${age} · 依据 ${date} 评估记录生成`
})

function formatDate(value) {
  if (!value)
    return ''
  return String(value).slice(0, 10)
}

function formatAge(row = {}) {
  const parts = []
  if (row.ageYears)
    parts.push(`${row.ageYears}岁`)
  if (row.ageMonths)
    parts.push(`${row.ageMonths}月`)
  if (row.ageDays)
    parts.push(`${row.ageDays}天`)
  return parts.join('')
}

function closeModal() {
  openModal.value = false
}

function parseAttachmentFilename(headerValue) {
  const text = `${headerValue || ''}`
  const encodedMatch = text.match(/filename\*=UTF-8''([^;]+)/i)
  if (encodedMatch?.[1])
    return decodeURIComponent(encodedMatch[1])
  const match = text.match(/filename="?([^";]+)"?/i)
  return match?.[1] || ''
}

function triggerDownload(response, fallbackName) {
  const contentType = String(response.headers?.['content-type'] || '')
  const blob = new Blob([response.data], {
    type: contentType || 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = parseAttachmentFilename(response.headers?.['content-disposition']) || fallbackName
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.setTimeout(() => URL.revokeObjectURL(url), 1000)
}

async function exportIepWord() {
  if (exportingWord.value)
    return
  const studentName = props.record?.studentName || '张一鸣'
  exportingWord.value = true
  try {
    const response = await downloadPEP3IEPPlanWordApi({
      id: props.record?.id,
      duration: planDuration.value,
    })
    triggerDownload(response, `${studentName}-康复个别化教育计划-${planDuration.value}个月.docx`)
    messageService.success('导出成功')
  }
  catch (error) {
    console.error('export iep plan failed', error)
    const text = await error?.response?.data?.text?.()
    if (text) {
      try {
        const payload = JSON.parse(text)
        messageService.error(payload?.message || '导出失败')
        return
      }
      catch {
      }
    }
    messageService.error('导出失败')
  }
  finally {
    exportingWord.value = false
  }
}

function openGoalPreview(goal, stage) {
  if (!goal || goal.collapsed)
    return
  previewGoal.value = goal
  previewStage.value = stage
  previewOpen.value = true
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    :centered="true"
    :closable="false"
    :footer="null"
    :keyboard="false"
    :mask-closable="false"
    :width="1360"
    :body-style="{ padding: 0 }"
    wrap-class-name="generate-iep-modal-wrap"
  >
    <section class="iep-modal">
      <header class="iep-modal__header">
        <div>
          <h2>生成IEP训练计划</h2>
          <p>{{ studentMeta }}</p>
        </div>
        <a-button type="text" class="iep-close-btn" @click="closeModal">
          <template #icon>
            <CloseOutlined />
          </template>
        </a-button>
      </header>

      <div class="iep-modal__summary">
        <div class="summary-tags">
          <a-tag color="blue">
            IEP草案
          </a-tag>
          <a-tag color="green">
            可编辑
          </a-tag>
          <a-tag color="orange">
            按领域生成
          </a-tag>
        </div>
        <div class="period-switch">
          <span>计划周期</span>
          <a-segmented
            v-model:value="planDuration"
            :options="[
              { label: '3个月', value: '3' },
              { label: '6个月', value: '6' },
            ]"
          />
          <em>切换后将重新划分阶段与目标数量</em>
        </div>
        <div class="summary-count">
          <strong>6</strong> 个领域 · <strong>{{ totalShortGoalCount }}</strong> 个短期目标 · 家庭干预同步生成
        </div>
      </div>

      <nav class="domain-tabs" aria-label="训练领域">
        <span class="domain-tabs__label">训练领域</span>
        <div class="domain-tab-list">
          <button
            v-for="domain in domains"
            :key="domain.key"
            type="button"
            class="domain-tab"
            :class="{ 'domain-tab--active': activeDomainKey === domain.key }"
            @click="activeDomainKey = domain.key"
          >
            <strong>{{ domain.name }}</strong>
            <small>长期{{ domain.longCount }} · 短期{{ domain.shortCount }}</small>
          </button>
        </div>
      </nav>

      <main class="iep-modal__body">
        <div class="domain-heading">
          <div>
            <h3>{{ currentDomain.name }}</h3>
            <a-tag color="blue">
              {{ planDuration }}个月计划
            </a-tag>
          </div>
          <a-space :size="8">
            <a-button size="small">
              <template #icon>
                <ReloadOutlined />
              </template>
              重新生成
            </a-button>
            <a-button size="small">
              <template #icon>
                <PlusOutlined />
              </template>
              新增短期目标
            </a-button>
          </a-space>
        </div>

        <section class="iep-section">
          <div class="section-title">
            <span>长期目标（1-3条）</span>
            <a-button type="link" size="small">
              <template #icon>
                <EditOutlined />
              </template>
              编辑
            </a-button>
          </div>
          <div class="readonly-field readonly-field--goal">
            {{ longGoalText }}
          </div>
        </section>

        <section class="iep-section short-goals">
          <div class="section-title">
            <span>短期目标（按阶段横向排列）</span>
            <div class="section-actions">
              <a-button type="link" size="small">
                展开全部
              </a-button>
              <a-button type="link" size="small">
                批量编辑
              </a-button>
            </div>
          </div>

          <div class="stage-board">
            <article
              v-for="(stage, stageIndex) in stages"
              :key="`${planDuration}-${stage.month}`"
              class="stage-column"
              :class="`stage-column--${stage.tone}`"
            >
              <header class="stage-column__header">
                <div class="stage-badge">
                  {{ String(stageIndex + 1).padStart(2, '0') }}
                </div>
                <div>
                  <div class="stage-title">
                    <strong>{{ stage.month }}</strong>
                    <span>{{ stage.title }}</span>
                  </div>
                  <p>{{ stage.note }}</p>
                </div>
                <a-tag>{{ stage.tag }}</a-tag>
              </header>

              <div class="stage-column__goals">
                <article
                  v-for="goal in stage.goals"
                  :key="goal.title"
                  class="goal-card"
                  :class="{ 'goal-card--collapsed': goal.collapsed }"
                >
                  <template v-if="goal.collapsed">
                    <div class="goal-card__compact">
                      <div>
                        <strong>{{ goal.title }}</strong>
                        <span>训练内容 3项</span>
                      </div>
                      <a-button type="link" size="small">
                        展开
                      </a-button>
                    </div>
                  </template>
                  <template v-else>
                    <div class="goal-card__head">
                      <strong>{{ goal.title }}</strong>
                      <a-space :size="4" class="goal-card__actions">
                        <a-button type="text" size="small" @click="openGoalPreview(goal, stage)">
                          <template #icon>
                            <EyeOutlined />
                          </template>
                        </a-button>
                        <a-button type="text" size="small">
                          <template #icon>
                            <EditOutlined />
                          </template>
                        </a-button>
                        <a-button type="text" size="small">
                          <template #icon>
                            <CopyOutlined />
                          </template>
                        </a-button>
                        <a-button type="text" size="small">
                          <template #icon>
                            <DeleteOutlined />
                          </template>
                        </a-button>
                      </a-space>
                    </div>
                    <div class="goal-card__content">
                      <div class="goal-panel goal-panel--content">
                        <label>训练内容</label>
                        <div class="training-content-list">
                          <div
                            v-for="(item, itemIndex) in goal.content"
                            :key="item"
                            class="training-content-item"
                          >
                            <span class="training-content-item__index">{{ itemIndex + 1 }}</span>
                            <span class="training-content-item__text">{{ item }}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  </template>
                </article>
              </div>
            </article>
          </div>
        </section>

        <section class="iep-section home-plan">
          <div class="section-title">
            <span>家庭干预计划</span>
            <a-button type="link" size="small">
              <template #icon>
                <EditOutlined />
              </template>
              编辑
            </a-button>
          </div>
          <div class="readonly-field readonly-field--home">
            {{ homePlanText }}
          </div>
        </section>
      </main>

      <footer class="iep-modal__footer">
        <div class="footer-hint">
          当前为{{ planDuration }}个月IEP草案；{{ planDuration === '3' ? '切换为6个月后横向扩展为6个阶段，可左右滚动查看。' : '当前横向展示6个阶段，可左右滚动查看。' }}
        </div>
        <div class="footer-actions">
          <a-button @click="closeModal">
            取消
          </a-button>
          <a-button>
            保存草稿
          </a-button>
          <a-button :loading="exportingWord" @click="exportIepWord">
            导出
          </a-button>
          <a-button type="primary" @click="closeModal">
            确认生成IEP
          </a-button>
        </div>
      </footer>
    </section>

    <a-modal
      v-model:open="previewOpen"
      :footer="null"
      width="680px"
      centered
      title="短期目标预览"
      wrap-class-name="iep-goal-preview-modal"
    >
      <div v-if="previewGoal" class="goal-preview">
        <div class="goal-preview__head">
          <a-tag color="blue">
            {{ previewStage?.month }}
          </a-tag>
          <strong>{{ previewGoal.title }}</strong>
        </div>
        <div class="goal-preview__section">
          <h4>训练内容</h4>
          <ol>
            <li v-for="item in previewGoal.content" :key="item">
              {{ item }}
            </li>
          </ol>
        </div>
      </div>
    </a-modal>
  </a-modal>
</template>

<style lang="less" scoped>
.iep-modal {
  overflow: hidden;
  color: #1f2937;
  background: #fff;
  border-radius: 8px;
}

.iep-modal__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 18px 22px 14px;
  border-bottom: 1px solid #edf0f5;

  h2 {
    margin: 0;
    color: #111827;
    font-size: 22px;
    font-weight: 650;
    line-height: 30px;
  }

  p {
    margin: 4px 0 0;
    color: #5f6b7a;
    font-size: 13px;
    line-height: 20px;
  }
}

.iep-close-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  color: #1f2937;
  font-size: 18px;
}

.iep-modal__summary {
  display: grid;
  grid-template-columns: 280px 1fr auto;
  gap: 16px;
  align-items: center;
  padding: 12px 22px;
  background: #fbfcfe;
  border-bottom: 1px solid #edf0f5;
}

.summary-tags {
  display: flex;
  gap: 8px;
  align-items: center;
}

.period-switch {
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: center;
  min-width: 0;

  span {
    color: #374151;
    font-size: 13px;
    font-weight: 500;
  }

  em {
    overflow: hidden;
    color: #6b7280;
    font-size: 12px;
    font-style: normal;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.summary-count {
  color: #4b5563;
  font-size: 12px;
  white-space: nowrap;

  strong {
    color: #1677ff;
    font-weight: 650;
  }
}

.domain-tabs {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 8px 22px;
  border-bottom: 1px solid #edf0f5;
}

.domain-tabs__label {
  flex: 0 0 auto;
  color: #374151;
  font-size: 13px;
  font-weight: 650;
  line-height: 20px;
}

.domain-tab-list {
  display: flex;
  flex: 1;
  gap: 8px;
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;
}

.domain-tab-list::-webkit-scrollbar {
  display: none;
}

.domain-tab {
  display: flex;
  flex: 0 0 auto;
  gap: 6px;
  align-items: center;
  height: 32px;
  padding: 0 12px;
  text-align: left;
  cursor: pointer;
  background: #fff;
  border: 1px solid #e5eaf3;
  border-radius: 6px;
  transition: all 0.18s ease;

  &:hover {
    border-color: #9ec5ff;
  }
}

.domain-tab--active {
  background: #eef6ff;
  border-color: #1677ff;
  box-shadow: 0 2px 8px rgba(22, 119, 255, 0.12);
}

.domain-tab strong {
  color: #1f2937;
  font-size: 13px;
  line-height: 18px;
  white-space: nowrap;
}

.domain-tab small {
  color: #64748b;
  font-size: 12px;
  line-height: 18px;
  white-space: nowrap;
}

.domain-tab--active strong,
.domain-tab--active small {
  color: #1677ff;
}

.iep-modal__body {
  max-height: calc(100vh - 284px);
  padding: 14px 22px 16px;
  overflow: auto;
  background: #fff;
  scrollbar-color: rgba(148, 163, 184, 0.7) transparent;
  scrollbar-width: thin;
}

.iep-modal__body::-webkit-scrollbar,
.stage-board::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.iep-modal__body::-webkit-scrollbar-thumb,
.stage-board::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.55);
  border-radius: 999px;
}

.iep-modal__body::-webkit-scrollbar-track,
.stage-board::-webkit-scrollbar-track {
  background: transparent;
}

.domain-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;

  > div:first-child {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  h3 {
    margin: 0;
    color: #111827;
    font-size: 20px;
    font-weight: 650;
    line-height: 28px;
  }
}

.iep-section + .iep-section {
  margin-top: 14px;
}

.section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;

  span {
    color: #111827;
    font-size: 15px;
    font-weight: 650;
    line-height: 22px;
  }
}

.section-actions {
  display: flex;
  gap: 2px;
  align-items: center;
}

.readonly-field {
  white-space: pre-wrap;
  background: #fff;
  border: 1px solid #dfe5ee;
  border-radius: 8px;
}

.readonly-field--goal {
  min-height: 82px;
  padding: 12px 14px;
  color: #374151;
  font-size: 13px;
  line-height: 24px;
}

.stage-board {
  display: grid;
  grid-auto-columns: minmax(382px, 1fr);
  grid-auto-flow: column;
  gap: 12px;
  padding-bottom: 6px;
  overflow-x: auto;
  overflow-y: visible;
}

.stage-column {
  display: flex;
  flex-direction: column;
  min-width: 382px;
  background: #fbfcfe;
  border: 1px solid #e3e9f2;
  border-radius: 10px;
}

.stage-column__header {
  display: grid;
  grid-template-columns: 38px 1fr auto;
  gap: 10px;
  align-items: center;
  padding: 12px;
  background: #fff;
  border-bottom: 1px solid #e8eef6;
}

.stage-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  color: #1677ff;
  font-weight: 700;
  background: #eaf3ff;
  border-radius: 10px;
}

.stage-title {
  display: flex;
  gap: 8px;
  align-items: center;

  strong,
  span {
    color: #111827;
    font-size: 14px;
    line-height: 20px;
  }

  strong {
    font-weight: 650;
  }
}

.stage-column__header p {
  margin: 2px 0 0;
  overflow: hidden;
  color: #667085;
  font-size: 12px;
  line-height: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.stage-column--green .stage-badge {
  color: #0f9f6e;
  background: #e9fbf4;
}

.stage-column--orange .stage-badge {
  color: #f97316;
  background: #fff4e8;
}

.stage-column--cyan .stage-badge {
  color: #0891b2;
  background: #ecfeff;
}

.stage-column--purple .stage-badge {
  color: #7c3aed;
  background: #f3efff;
}

.stage-column--red .stage-badge {
  color: #dc2626;
  background: #fff1f2;
}

.stage-column__goals {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 10px;
}

.goal-card {
  flex: 0 0 auto;
  padding: 10px;
  background: #fff;
  border: 1px solid #e4e9f1;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.03);
}

.goal-card__head {
  display: flex;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;

  strong {
    overflow: hidden;
    color: #111827;
    font-size: 13px;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.goal-card__actions {
  flex: 0 0 auto;
  opacity: 0.74;
}

.goal-card__content {
  display: block;
}

.goal-panel {
  min-height: 112px;
  padding: 8px;
  background: #f8fafc;
  border: 1px solid #edf1f7;
  border-radius: 6px;

  label {
    display: block;
    margin-bottom: 6px;
    color: #475569;
    font-size: 12px;
    font-weight: 650;
    line-height: 18px;
  }
}

.goal-panel--content {
  text-align: left;
}

.training-content-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.training-content-item {
  display: grid;
  grid-template-columns: 18px minmax(0, 1fr);
  gap: 6px;
  align-items: flex-start;
  color: #344054;
  font-size: 12px;
  line-height: 20px;
  text-align: left;
}

.training-content-item__index {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  margin-top: 2px;
  color: #1677ff;
  font-size: 11px;
  font-weight: 650;
  line-height: 16px;
  background: #eaf3ff;
  border-radius: 50%;
}

.training-content-item__text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.goal-card--collapsed {
  padding: 0;
}

.goal-card__compact {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 54px;
  padding: 10px 12px;

  > div {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }

  strong {
    overflow: hidden;
    color: #111827;
    font-size: 13px;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  span {
    color: #667085;
    font-size: 12px;
    line-height: 18px;
  }
}

.readonly-field--home {
  min-height: 112px;
  padding: 12px 14px;
  color: #374151;
  font-size: 13px;
  line-height: 24px;
  background: #fbfdff;
}

.iep-modal__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 22px;
  background: #fff;
  border-top: 1px solid #edf0f5;
}

.footer-hint {
  position: relative;
  padding-left: 18px;
  color: #5f6b7a;
  font-size: 12px;
  line-height: 20px;

  &::before {
    position: absolute;
    top: 5px;
    left: 0;
    width: 10px;
    height: 10px;
    background: #1677ff;
    border-radius: 50%;
    content: "";
  }
}

.footer-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.goal-preview {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.goal-preview__head {
  display: flex;
  gap: 8px;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 1px solid #edf0f5;

  strong {
    color: #111827;
    font-size: 16px;
    line-height: 24px;
  }
}

.goal-preview__section {
  padding: 12px;
  background: #f8fafc;
  border: 1px solid #edf1f7;
  border-radius: 8px;

  h4 {
    margin: 0 0 8px;
    color: #111827;
    font-size: 14px;
    font-weight: 650;
    line-height: 22px;
  }

  ol {
    padding-left: 22px;
    margin: 0;
    color: #374151;
    font-size: 13px;
    line-height: 24px;
  }

  dl {
    display: flex;
    flex-direction: column;
    gap: 8px;
    margin: 0;
  }

  dl > div {
    display: grid;
    grid-template-columns: 42px minmax(0, 1fr);
    gap: 10px;
  }

  dt {
    color: #64748b;
    font-size: 13px;
    line-height: 22px;
  }

  dd {
    margin: 0;
    color: #374151;
    font-size: 13px;
    line-height: 22px;
  }
}
</style>

<style lang="less">
.generate-iep-modal-wrap {
  .ant-modal-content {
    overflow: hidden;
    padding: 0;
    border-radius: 10px;
  }

  .ant-modal {
    max-width: calc(100vw - 48px);
  }

  .ant-segmented {
    padding: 2px;
  }

  .ant-segmented-item {
    min-width: 72px;
  }
}
</style>
