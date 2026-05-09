<script setup>
import {
  CloseOutlined,
  DeleteOutlined,
  PlusOutlined,
} from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { resolveIEPPlanAssessmentAdapter } from './iep-plan-adapters'
import { useUserStore } from '@/stores/user'
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

const emit = defineEmits(['update:open', 'saved', 'confirmed'])
const userStore = useUserStore()

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const assessmentAdapter = computed(() => resolveIEPPlanAssessmentAdapter(props.record))
const planDuration = ref('3')
const activeDomainKey = ref('language')
const exportingWord = ref(false)
const aiGenerating = ref(false)
const generatingExecutionPlan = ref(false)
const executionPlanGeneratingType = ref('')
const loadingSavedPlan = ref(false)
const savingDraft = ref(false)
const confirmingPlan = ref(false)
const savingExecutionPlan = ref(false)
const showPlanLoadingOverlay = ref(false)
const aiStreamStatus = ref('')
const aiStreamText = ref('')
const streamingPlan = ref(null)
const generatedPlan = ref(null)
const editablePlan = ref(null)
const savedPlanStatus = ref('')
const executionPlanView = ref('iep')
const selectedExecutionMonth = ref(1)
const selectedExecutionWeek = ref(1)
const monthlyPlans = ref({})
const monthlyPlan = computed({
  get: () => monthlyPlans.value[String(clampExecutionMonth(selectedExecutionMonth.value))] || null,
  set(value) {
    const key = String(clampExecutionMonth(selectedExecutionMonth.value))
    const nextPlans = { ...monthlyPlans.value }
    if (value)
      nextPlans[key] = value
    else
      delete nextPlans[key]
    monthlyPlans.value = nextPlans
  },
})
const weeklyPlans = ref({})
const weeklyPlan = computed({
  get: () => weeklyPlans.value[executionWeekStorageKey.value] || null,
  set(value) {
    const key = executionWeekStorageKey.value
    const nextPlans = { ...weeklyPlans.value }
    if (value)
      nextPlans[key] = value
    else
      delete nextPlans[key]
    weeklyPlans.value = nextPlans
  },
})
const editingPlanView = ref('')
const selectedPlanRowIndex = ref(0)
const selectedMonthlyRowIndex = ref(0)
const selectedMonthlyContentIndex = ref(0)
const selectedWeeklyRowIndex = ref(0)
const iepModalBodyRef = ref(null)
const aiStreamAbortController = ref(null)
const executionPlanStreamAbortController = ref(null)
let aiGenerationRequestKey = 0
let executionPlanRequestKey = 0
let loadPlanRequestKey = 0
let ignoreNextPlanDurationWatch = false
let planLoadingOverlayTimer = 0

const domains = [
  { key: 'language', name: '语言沟通', icon: '语', longCount: 3, shortCount: 6 },
  { key: 'social', name: '社交互动', icon: '社', longCount: 2, shortCount: 5 },
  { key: 'cognition', name: '认知理解', icon: '认', longCount: 2, shortCount: 4 },
  { key: 'fineMotor', name: '精细动作', icon: '精', longCount: 2, shortCount: 4 },
  { key: 'sensory', name: '感统运动', icon: '感', longCount: 2, shortCount: 5 },
  { key: 'selfCare', name: '生活自理', icon: '自', longCount: 1, shortCount: 4 },
]

const courseFormOptions = [
  { label: '个训', value: '个训' },
  { label: '集体课', value: '集体课' },
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

const currentTeacherName = computed(() => {
  return String(userStore.nickname || userStore.userInfo?.nickName || props.record?.examinerName || '当前老师').trim()
})

const assessmentPlanDate = computed(() => {
  return formatDate(props.record?.assessmentDate) || formatDate(new Date())
})

const defaultPlanDateRange = computed(() => {
  const baseDate = assessmentPlanDate.value || '2026-05-01'
  const start = firstDayOfMonth(baseDate)
  const end = lastDayAfterMonths(start, Number(planDuration.value || 6))
  return { start, end }
})

const planTitle = computed(() => {
  return planDuration.value === '3' ? '康复教学季度计划' : '康复教学半年计划'
})

const defaultPlanRows = computed(() => {
  const { start, end } = defaultPlanDateRange.value
  return [
    {
      domain: '语言沟通',
      longGoal: iepData.language.longGoals.join('\n'),
      shortGoal: [
        '用单词或2词短句表达需求、拒绝和帮助。',
        '回应简单问句并在课堂、家庭场景中泛化。',
      ].join('\n'),
      courseForm: '个训',
      startEndDate: `${start} - ${end}`,
    },
    {
      domain: '社交互动',
      longGoal: iepData.social.longGoals.join('\n'),
      shortGoal: [
        '在成人提示下完成等待、轮替和简单回应。',
        '小组活动中增加主动发起沟通次数。',
      ].join('\n'),
      courseForm: '集体课',
      startEndDate: `${start} - ${end}`,
    },
    {
      domain: '认知理解',
      longGoal: iepData.cognition.longGoals.join('\n'),
      shortGoal: [
        '完成物品功能、类别配对和简单两步指令。',
        '在桌面任务和生活任务中稳定理解口头指令。',
      ].join('\n'),
      courseForm: '个训',
      startEndDate: `${start} - ${end}`,
    },
    {
      domain: '生活自理',
      longGoal: iepData.selfCare.longGoals.join('\n'),
      shortGoal: [
        '按步骤完成洗手、穿鞋、整理物品等生活流程。',
        '减少成人替代，提升主动参与和任务完成度。',
      ].join('\n'),
      courseForm: '个训',
      startEndDate: `${start} - ${end}`,
    },
  ]
})

const planSheet = computed(() => {
  const { start, end } = defaultPlanDateRange.value
  if (editablePlan.value)
    return editablePlan.value
  const activePlan = streamingPlan.value || generatedPlan.value
  if (activePlan) {
    return createPlanSheetFromPlan(activePlan, { preserveRows: false, model: activePlan.model || (streamingPlan.value ? 'AI生成中' : '') })
  }
  return {
    title: planTitle.value,
    student: {
      name: props.record?.studentName || '',
      gender: props.record?.studentGender || '',
      birthDate: formatDate(props.record?.birthDate) || '',
    },
    meta: {
      planDate: assessmentPlanDate.value || start,
      participant: currentTeacherName.value,
      implementer: currentTeacherName.value,
      startDate: start,
      endDate: end,
    },
    rows: [],
    model: '',
  }
})

const planRows = computed(() => planSheet.value.rows)

const selectedPlanRow = computed(() => {
  const rows = planRows.value || []
  return rows[selectedPlanRowIndex.value] || rows[0] || null
})

const selectedPlanRowTitle = computed(() => {
  const row = selectedPlanRow.value
  if (!row)
    return '未选择目标'
  const sameDomainRows = (planRows.value || []).filter(item => item.domain === row.domain)
  const domainIndex = sameDomainRows.findIndex(item => item === row)
  return `${row.domain || '综合康复'} / 第${Math.max(1, domainIndex + 1)}条短期目标`
})

const planStatusLabel = computed(() => {
  if (generatingExecutionPlan.value)
    return '执行计划生成中'
  if (aiGenerating.value)
    return '生成中'
  if (loadingSavedPlan.value)
    return '加载中'
  if (savedPlanStatus.value === 'confirmed')
    return '已确认'
  if (savedPlanStatus.value === 'draft')
    return '草稿'
  if (editablePlan.value || generatedPlan.value)
    return planSheet.value.model || 'AI草案'
  return '未生成'
})

const aiStreamTail = computed(() => {
  const text = aiStreamText.value.replace(/\s+/g, ' ').trim()
  return text.length > 220 ? text.slice(-220) : text
})

const aiProgressPercent = computed(() => {
  if (!aiGenerating.value && generatedPlan.value)
    return 100
  if (!aiGenerating.value)
    return 0
  return Math.min(96, 12 + Math.floor(aiStreamText.value.length / 80))
})

const executionPlanGeneratingLabel = computed(() => {
  if (executionPlanGeneratingType.value === 'monthly')
    return '月度计划'
  if (executionPlanGeneratingType.value === 'weekly')
    return '周计划'
  return '执行计划'
})

const generationOverlayActive = computed(() => aiGenerating.value || generatingExecutionPlan.value)
const planLoadingOverlayActive = computed(() => showPlanLoadingOverlay.value && loadingSavedPlan.value && hasPlanContent.value && !generationOverlayActive.value)

const generationOverlayTitle = computed(() => {
  if (generatingExecutionPlan.value)
    return `正在AI生成${executionPlanGeneratingLabel.value}`
  return aiStreamStatus.value || '正在生成IEP计划'
})

const generationOverlayDescription = computed(() => {
  if (generatingExecutionPlan.value) {
    if (aiStreamTail.value)
      return aiStreamTail.value
    return `DeepSeek正在根据当前${planTitle.value}${monthlyPlan.value?.rows?.length && executionPlanGeneratingType.value === 'weekly' ? '和月度计划' : ''}生成${executionPlanGeneratingLabel.value}，接收到内容后会同步渲染到A4表格。`
  }
  return aiStreamTail.value || assessmentAdapter.value.generationDescription
})

const generationOverlayWarning = computed(() => {
  if (generatingExecutionPlan.value)
    return '执行计划生成过程中请不要刷新页面或关闭弹窗，否则本次返回结果无法展示。'
  return '生成过程中请不要刷新页面或关闭弹窗；如需离开，请先确认取消本次生成。'
})

const generationSourceText = computed(() => assessmentAdapter.value.generationSourceText || assessmentAdapter.value.generationBasisText)
const fallbackGenerationBasisText = computed(() => assessmentAdapter.value.generationFallbackBasisText || assessmentAdapter.value.generationBasisText)

const generationOverlayPercent = computed(() => {
  if (generatingExecutionPlan.value)
    return Math.min(96, 12 + Math.floor(aiStreamText.value.length / 80))
  return aiProgressPercent.value
})

const generationInlineStatus = computed(() => {
  if (generatingExecutionPlan.value)
    return `正在生成${executionPlanGeneratingLabel.value}`
  return aiStreamStatus.value || 'AI生成中'
})

const generationInlineTail = computed(() => {
  if (generatingExecutionPlan.value)
    return aiStreamTail.value || `DeepSeek正在拆解${planTitle.value}，请等待生成完成。`
  return aiStreamTail.value
})

const planDisplayRows = computed(() => {
  const rows = planRows.value || []
  return rows.map((row, index) => {
    const isFirstInDomain = index === 0 || row.domain !== rows[index - 1]?.domain
    let rowSpan = 1
    if (isFirstInDomain) {
      for (let next = index + 1; next < rows.length && rows[next]?.domain === row.domain; next++)
        rowSpan++
    }
    return {
      ...row,
      sourceIndex: index,
      showGroupCell: isFirstInDomain,
      rowSpan,
    }
  })
})

const selectedExecutionMonthLabel = computed(() => executionMonthLabelForIndex(selectedExecutionMonth.value))
const selectedExecutionMonthRange = computed(() => executionMonthRangeForIndex(selectedExecutionMonth.value))
const selectedExecutionMonthGenerated = computed(() => !!monthlyPlans.value[String(clampExecutionMonth(selectedExecutionMonth.value))])
const executionWeekCount = computed(() => weekCountForRange(selectedExecutionMonthRange.value))
const selectedExecutionWeekValue = computed(() => clampExecutionWeek(selectedExecutionWeek.value))
const selectedExecutionWeekLabel = computed(() => `第${selectedExecutionWeekValue.value}周`)
const selectedExecutionWeekRange = computed(() => executionWeekRangeForIndex(selectedExecutionWeekValue.value))
const executionWeekStorageKey = computed(() => `${clampExecutionMonth(selectedExecutionMonth.value)}-${selectedExecutionWeekValue.value}`)
const selectedExecutionWeekGenerated = computed(() => !!weeklyPlans.value[executionWeekStorageKey.value])

const executionNavigatorMonths = computed(() => {
  const count = Number(planDuration.value) === 6 ? 6 : 3
  return Array.from({ length: count }, (_, index) => {
    const monthIndex = index + 1
    const monthRange = executionMonthRangeForIndex(monthIndex)
    const monthGenerated = !!monthlyPlans.value[String(monthIndex)]
    const monthWeekCount = weekCountForRange(monthRange)
    const weeks = Array.from({ length: monthWeekCount }, (_, weekOffset) => {
      const weekIndex = weekOffset + 1
      const weekRange = executionWeekRangeForMonth(monthIndex, weekIndex)
      const key = `${monthIndex}-${weekIndex}`
      return {
        index: weekIndex,
        label: `第${weekIndex}周`,
        rangeText: `${formatNavigatorDateRange(weekRange.start, weekRange.end)}`,
        generated: !!weeklyPlans.value[key],
      }
    })
    return {
      index: monthIndex,
      label: executionMonthLabelForIndex(monthIndex),
      rangeText: formatNavigatorDateRange(monthRange.start, monthRange.end),
      generated: monthGenerated,
      active: clampExecutionMonth(selectedExecutionMonth.value) === monthIndex,
      weeks,
    }
  })
})

const weeklyDisplayDates = computed(() => {
  const dates = Array.isArray(weeklyPlan.value?.weekDates)
    ? weeklyPlan.value.weekDates.slice(0, 6).map(date => formatMonthDayDate(date))
    : []
  while (dates.length < 6)
    dates.push('')
  return dates
})

const activePreviewTitle = computed(() => {
  if (executionPlanView.value === 'monthly')
    return monthlyPlan.value?.title || `康复教学${selectedExecutionMonthLabel.value}计划`
  if (executionPlanView.value === 'weekly')
    return weeklyPlan.value?.title || `康复教学周计划日记录卡${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}`
  return planSheet.value.title
})

const activePreviewCountText = computed(() => {
  if (executionPlanView.value === 'monthly')
    return `${monthlyDisplayRows.value.length || 0}条训练内容`
  if (executionPlanView.value === 'weekly')
    return `${weeklyPlan.value?.rows?.length || 0}项训练`
  return `${planRows.value.length}条计划`
})

const canExportActivePreview = computed(() => {
  if (generatingExecutionPlan.value)
    return false
  if (executionPlanView.value === 'monthly')
    return !!monthlyPlan.value?.rows?.length
  if (executionPlanView.value === 'weekly')
    return !!weeklyPlan.value?.rows?.length
  return !!planRows.value.length
})

const editingPreviewView = computed(() => editingPlanView.value === executionPlanView.value ? editingPlanView.value : '')
const isIepPreview = computed(() => executionPlanView.value === 'iep')
const isPlanEditable = computed(() => executionPlanView.value === 'iep' && editingPreviewView.value === 'iep' && !!editablePlan.value && !aiGenerating.value && !generatingExecutionPlan.value && !loadingSavedPlan.value)
const isMonthlyPlanEditable = computed(() => executionPlanView.value === 'monthly' && editingPreviewView.value === 'monthly' && !!monthlyPlan.value?.rows?.length && !aiGenerating.value && !generatingExecutionPlan.value && !loadingSavedPlan.value)
const isWeeklyPlanEditable = computed(() => executionPlanView.value === 'weekly' && editingPreviewView.value === 'weekly' && !!weeklyPlan.value?.rows?.length && !aiGenerating.value && !generatingExecutionPlan.value && !loadingSavedPlan.value)
const isAnyPlanEditable = computed(() => isPlanEditable.value || isMonthlyPlanEditable.value || isWeeklyPlanEditable.value)
const modalWidth = computed(() => (isAnyPlanEditable.value ? 1388 : 1180))
const navigationDisabled = computed(() => aiGenerating.value || loadingSavedPlan.value || generatingExecutionPlan.value || savingExecutionPlan.value || isAnyPlanEditable.value)
const canEditActivePreview = computed(() => {
  if (executionPlanView.value === 'monthly')
    return !!monthlyPlan.value?.rows?.length
  if (executionPlanView.value === 'weekly')
    return !!weeklyPlan.value?.rows?.length
  return !!planRows.value.length
})
const activeEditingLabel = computed(() => {
  if (executionPlanView.value === 'monthly')
    return selectedMonthlyEditTitle.value
  if (executionPlanView.value === 'weekly')
    return selectedWeeklyEditTitle.value
  return selectedPlanRowTitle.value
})

const monthlyDisplayRows = computed(() => {
  const rows = monthlyPlan.value?.rows || []
  const result = []
  rows.forEach((row, rowIndex) => {
    const items = normalizeMonthlyTrainingItems(row)
    items.forEach((item, contentIndex) => {
      result.push({
        ...row,
        rowIndex,
        contentIndex,
        contentRowSpan: items.length,
        showTargetCell: contentIndex === 0,
        trainingContent: item.content,
        trainingContentText: `${contentIndex + 1}. ${item.content}`,
        trainingStartEndDate: item.startEndDate,
      })
    })
  })
  return result
})

const selectedMonthlyDisplayRow = computed(() => {
  const rows = monthlyDisplayRows.value || []
  return rows.find(row => row.rowIndex === selectedMonthlyRowIndex.value && row.contentIndex === selectedMonthlyContentIndex.value) || rows[0] || null
})

const selectedMonthlyEditTitle = computed(() => {
  const row = selectedMonthlyDisplayRow.value
  if (!row)
    return `${selectedExecutionMonthLabel.value}计划`
  return `${row.domain || '综合康复'} / 第${row.contentIndex + 1}项训练内容`
})

const selectedWeeklyPlanRow = computed(() => {
  const rows = weeklyPlan.value?.rows || []
  return rows[selectedWeeklyRowIndex.value] || rows[0] || null
})

const selectedWeeklyEditTitle = computed(() => {
  const row = selectedWeeklyPlanRow.value
  if (!row)
    return `${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}周计划`
  return `${row.project || '训练项目'} / 第${Math.max(1, selectedWeeklyRowIndex.value + 1)}项训练`
})

function normalizeMonthlyTrainingItems(row = {}) {
  const rawItems = Array.isArray(row.trainingItems) ? row.trainingItems : []
  const items = rawItems
    .map(item => ({
      content: String(item?.content || '').trim(),
      startEndDate: String(item?.startEndDate || '').trim(),
    }))
  if (!items.length)
    items.push({ content: '', startEndDate: '' })

  const startDate = monthlyPlan.value?.meta?.startDate || defaultPlanDateRange.value.start
  const endDate = monthlyPlan.value?.meta?.endDate || lastDayAfterMonths(startDate, 1)
  return items.map((item, index) => ({
    content: item.content,
    startEndDate: item.startEndDate || dateRangeForMonthlyItem(startDate, endDate, index, items.length) || '',
  }))
}

const executionPlanSourceText = computed(() => {
  if (executionPlanView.value === 'weekly' && monthlyPlan.value)
    return `依据：${planTitle.value} + ${monthlyPlan.value.title} · ${selectedExecutionWeekLabel.value}`
  if (executionPlanView.value === 'weekly')
    return `依据：${planTitle.value} · ${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}`
  if (executionPlanView.value === 'monthly')
    return `依据：${planTitle.value} · ${selectedExecutionMonthLabel.value}`
  return ''
})

const studentMeta = computed(() => {
  const name = props.record?.studentName || '张一鸣'
  const gender = props.record?.studentGender || '男'
  const age = formatAge(props.record) || '3岁1月'
  return `${name} · ${gender} · ${age}`
})

const hasPlanContent = computed(() => {
  return aiGenerating.value || !!editablePlan.value || !!generatedPlan.value || !!streamingPlan.value || planRows.value.length > 0 || Object.keys(monthlyPlans.value).length > 0 || Object.keys(weeklyPlans.value).length > 0
})

const isConfirmedPlan = computed(() => {
  return savedPlanStatus.value === 'confirmed'
})

const modalTitleText = computed(() => {
  if (executionPlanView.value === 'monthly')
    return '查看月度计划'
  if (executionPlanView.value === 'weekly')
    return '查看周计划'
  if (loadingSavedPlan.value && props.record?.iepPlanStatus === 'confirmed')
    return '查看IEP训练计划'
  if (isConfirmedPlan.value)
    return '查看IEP训练计划'
  if (savedPlanStatus.value === 'draft')
    return '编辑IEP草稿'
  return '生成IEP训练计划'
})

const headerPlanMeta = computed(() => {
  if (aiGenerating.value)
    return '正在生成IEP计划'
  if (!hasPlanContent.value)
    return ''
  const date = formatDate(props.record?.assessmentDate)
  return date ? `评估日期：${date}` : ''
})

const periodHint = computed(() => {
  if (aiGenerating.value)
    return '正在按周期生成计划行和起止日期'
  if (isConfirmedPlan.value)
    return '当前周期已确认，另一个周期独立保存'
  if (savedPlanStatus.value === 'draft')
    return '当前周期是草稿，另一个周期独立保存'
  if (hasPlanContent.value)
    return '已生成，可直接编辑表格'
  return '每个周期单独生成和保存'
})

const headerPlanStatusText = computed(() => {
  if (loadingSavedPlan.value)
    return `正在加载${planTitle.value}`
  if (!hasPlanContent.value)
    return `${planTitle.value}未生成`
  return `${planRows.value.length} 条计划 · ${planStatusLabel.value} · 实时表格`
})

function formatDate(value) {
  if (!value)
    return ''
  if (value instanceof Date && !Number.isNaN(value.getTime()))
    return value.toISOString().slice(0, 10)
  return String(value).slice(0, 10)
}

function formatMonthDayDate(value) {
  const text = formatDate(value)
  if (!text)
    return ''
  const match = text.match(/^(\d{4})-(\d{2})-(\d{2})$/)
  if (match)
    return `${match[2]}.${match[3]}`
  return text.replace(/-/g, '.')
}

function addMonths(dateText, months) {
  const date = new Date(`${dateText}T00:00:00`)
  if (Number.isNaN(date.getTime()))
    return ''
  date.setMonth(date.getMonth() + months)
  return date.toISOString().slice(0, 10)
}

function formatLocalDate(date) {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function clampExecutionMonth(value) {
  const count = Number(planDuration.value) === 6 ? 6 : 3
  const month = Number(value || 1)
  if (Number.isNaN(month))
    return 1
  return Math.max(1, Math.min(count, Math.floor(month)))
}

function executionMonthDateForIndex(monthIndex) {
  const startText = defaultPlanDateRange.value.start || formatDate(new Date())
  const date = new Date(`${startText}T00:00:00`)
  if (Number.isNaN(date.getTime()))
    return null
  return new Date(date.getFullYear(), date.getMonth() + clampExecutionMonth(monthIndex) - 1, 1)
}

function executionMonthRangeForIndex(monthIndex) {
  const start = executionMonthDateForIndex(monthIndex)
  if (!start)
    return { start: defaultPlanDateRange.value.start, end: defaultPlanDateRange.value.end }
  const end = new Date(start.getFullYear(), start.getMonth() + 1, 0)
  return {
    start: formatLocalDate(start),
    end: formatLocalDate(end),
  }
}

function weekCountForRange(range = {}) {
  const start = new Date(`${range.start}T00:00:00`)
  const end = new Date(`${range.end}T00:00:00`)
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime()) || end < start)
    return 1
  const days = Math.floor((end.getTime() - start.getTime()) / 86400000) + 1
  return Math.max(1, Math.ceil(days / 6))
}

function clampExecutionWeek(value) {
  const week = Number(value || 1)
  if (Number.isNaN(week))
    return 1
  return Math.max(1, Math.min(executionWeekCount.value, Math.floor(week)))
}

function executionWeekRangeForIndex(weekIndex) {
  return executionWeekRangeForMonth(selectedExecutionMonth.value, weekIndex)
}

function executionWeekRangeForMonth(monthIndex, weekIndex) {
  const monthRange = executionMonthRangeForIndex(monthIndex)
  const start = new Date(`${monthRange.start}T00:00:00`)
  const end = new Date(`${monthRange.end}T00:00:00`)
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime()) || end < start)
    return { start: monthRange.start, end: monthRange.end }
  const index = Math.max(1, Math.min(weekCountForRange(monthRange), Number(weekIndex || 1)))
  const weekStart = new Date(start.getFullYear(), start.getMonth(), start.getDate() + (index - 1) * 6)
  const rawWeekEnd = new Date(weekStart.getFullYear(), weekStart.getMonth(), weekStart.getDate() + 5)
  const weekEnd = rawWeekEnd > end ? end : rawWeekEnd
  return {
    start: formatLocalDate(weekStart),
    end: formatLocalDate(weekEnd),
  }
}

function formatNavigatorDateRange(start, end) {
  const startText = formatMonthDayDate(start)
  const endText = formatMonthDayDate(end)
  if (!startText && !endText)
    return ''
  if (!startText)
    return endText
  if (!endText)
    return startText
  return `${startText}-${endText}`
}

function executionMonthLabelForIndex(monthIndex) {
  const date = executionMonthDateForIndex(monthIndex)
  if (!date)
    return `第${clampExecutionMonth(monthIndex)}月`
  return `${date.getMonth() + 1}月`
}

function firstDayOfMonth(dateText) {
  const date = new Date(`${dateText}T00:00:00`)
  if (Number.isNaN(date.getTime()))
    return dateText
  return formatLocalDate(new Date(date.getFullYear(), date.getMonth(), 1))
}

function lastDayAfterMonths(dateText, months) {
  const date = new Date(`${dateText}T00:00:00`)
  if (Number.isNaN(date.getTime()))
    return addMonths(dateText, months)
  return formatLocalDate(new Date(date.getFullYear(), date.getMonth() + months, 0))
}

function dateRangeForMonthlyItem(startText, endText, itemIndex, itemCount) {
  const start = new Date(`${startText}T00:00:00`)
  const end = new Date(`${endText}T00:00:00`)
  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime()) || end < start)
    return [startText, endText].filter(Boolean).join(' - ')
  const count = Math.max(1, Number(itemCount || 1))
  const index = Math.max(0, Math.min(count - 1, Number(itemIndex || 0)))
  const totalDays = Math.floor((end.getTime() - start.getTime()) / 86400000) + 1
  const offsetStart = Math.floor(index * totalDays / count)
  const offsetEnd = Math.max(offsetStart, Math.floor((index + 1) * totalDays / count) - 1)
  const itemStart = new Date(start.getFullYear(), start.getMonth(), start.getDate() + offsetStart)
  const itemEnd = new Date(start.getFullYear(), start.getMonth(), start.getDate() + offsetEnd)
  return `${formatLocalDate(itemStart)} - ${formatLocalDate(itemEnd)}`
}

function buildStageDateRanges(startDate, durationMonths) {
  const date = new Date(`${startDate}T00:00:00`)
  if (Number.isNaN(date.getTime()))
    return []
  const stageCount = 3
  const baseMonths = Math.floor(durationMonths / stageCount)
  const remainder = durationMonths % stageCount
  const ranges = []
  let current = new Date(date.getFullYear(), date.getMonth(), 1)
  for (let index = 0; index < stageCount; index++) {
    const months = Math.max(1, baseMonths + (index < remainder ? 1 : 0))
    const end = new Date(current.getFullYear(), current.getMonth() + months, 0)
    ranges.push(`${formatLocalDate(current)} - ${formatLocalDate(end)}`)
    current = new Date(end.getFullYear(), end.getMonth(), end.getDate() + 1)
  }
  return ranges
}

async function generateMonthlyPlan(options = {}) {
  if (generatingExecutionPlan.value)
    return
  const forceRegenerate = !!options.forceRegenerate
  if (selectedExecutionMonthGenerated.value && !forceRegenerate) {
    messageService.warning(`${selectedExecutionMonthLabel.value}计划已生成，如需覆盖请进入该月份后点击“重新生成”`)
    return
  }
  const sourcePlan = planPayloadForSave()
  if (!sourcePlan.rows.length) {
    messageService.warning('请先生成或填写IEP计划')
    return
  }
  executionPlanGeneratingType.value = 'monthly'
  executionPlanRequestKey += 1
  const requestKey = executionPlanRequestKey
  const abortController = new AbortController()
  executionPlanStreamAbortController.value = abortController
  generatingExecutionPlan.value = true
  aiStreamStatus.value = `正在准备${selectedExecutionMonthLabel.value}计划生成上下文`
  aiStreamText.value = ''
  monthlyPlan.value = null
  clearSelectedMonthWeeklyPlans()
  executionPlanView.value = 'monthly'
  try {
    const result = await assessmentAdapter.value.generateExecutionPlanStream(
      {
        id: props.record?.id,
        durationMonths: planDuration.value,
        planType: 'monthly',
        targetMonthIndex: selectedExecutionMonth.value,
        sourcePlan,
      },
      {
        onStatus(message) {
          if (requestKey !== executionPlanRequestKey)
            return
          aiStreamStatus.value = message || `正在生成${selectedExecutionMonthLabel.value}计划`
        },
        onDelta(text) {
          if (requestKey !== executionPlanRequestKey)
            return
          aiStreamStatus.value = `正在接收${selectedExecutionMonthLabel.value}计划内容`
          aiStreamText.value += text
          const partialPlan = buildStreamingMonthlyPlanFromText(aiStreamText.value)
          if (partialPlan)
            monthlyPlan.value = partialPlan
        },
        onDone(data) {
          if (requestKey !== executionPlanRequestKey)
            return
          monthlyPlan.value = data
          aiStreamStatus.value = `${selectedExecutionMonthLabel.value}计划生成完成`
        },
      },
      {
        signal: abortController.signal,
      },
    )
    if (requestKey !== executionPlanRequestKey)
      return
    monthlyPlan.value = result
    await persistExecutionPlan('monthly', result)
    executionPlanView.value = 'monthly'
    await scrollPlanViewToTop()
    messageService.success(`AI已生成${selectedExecutionMonthLabel.value}计划`)
  }
  catch (error) {
    if (requestKey !== executionPlanRequestKey || isAbortError(error))
      return
    console.error('generate monthly plan failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '生成月度计划失败')
  }
  finally {
    if (requestKey === executionPlanRequestKey) {
      generatingExecutionPlan.value = false
      executionPlanGeneratingType.value = ''
      executionPlanStreamAbortController.value = null
    }
  }
}

function confirmRegenerateMonthlyPlan() {
  if (!selectedExecutionMonthGenerated.value) {
    generateMonthlyPlan()
    return
  }
  Modal.confirm({
    title: `重新生成${selectedExecutionMonthLabel.value}计划`,
    content: `重新生成会覆盖当前${selectedExecutionMonthLabel.value}计划内容，已导出的Word不受影响。确认要继续吗？`,
    okText: '重新生成',
    cancelText: '保留当前计划',
    closable: true,
    okButtonProps: { danger: true },
    centered: true,
    onOk() {
      runAfterConfirmClosed(() => generateMonthlyPlan({ forceRegenerate: true }))
    },
  })
}

function confirmGenerateMonthlyPlan() {
  Modal.confirm({
    title: `生成${selectedExecutionMonthLabel.value}计划`,
    content: `将基于${generationSourceText.value}和当前IEP总计划，生成${selectedExecutionMonthLabel.value}月度计划。确认要继续吗？`,
    okText: '确认生成',
    cancelText: '先不生成',
    okButtonProps: { type: 'primary' },
    closable: true,
    centered: true,
    onOk() {
      runAfterConfirmClosed(() => generateMonthlyPlan())
    },
  })
}

function confirmRegenerateIepPlan() {
  if (!planRows.value.length) {
    confirmGenerateIepPlan()
    return
  }
  Modal.confirm({
    title: '重新生成IEP计划',
    content: '重新生成会覆盖当前IEP计划内容，已导出的Word不受影响。确认要继续吗？',
    okText: '重新生成',
    cancelText: '保留当前计划',
    okButtonProps: { danger: true },
    closable: true,
    centered: true,
    onOk() {
      runAfterConfirmClosed(() => generateAIPlan())
    },
  })
}

async function resolveGeneratePreConfirm() {
  try {
    return await assessmentAdapter.value.shouldConfirmBeforeGenerate?.(props.record)
  }
  catch (error) {
    console.error('check iep generation context failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '报告解读状态读取失败，请稍后重试')
    return false
  }
}

function showGeneratePreConfirm(confirmConfig, options = {}) {
  Modal.confirm({
    title: confirmConfig.title || '确认生成IEP计划？',
    content: confirmConfig.content || '确认继续生成IEP计划吗？',
    okText: confirmConfig.okText || '确定',
    cancelText: confirmConfig.cancelText || '取消',
    okButtonProps: { type: 'primary' },
    closable: true,
    centered: true,
    onOk() {
      runAfterConfirmClosed(() => generateAIPlan({ skipPrecheck: true, fallbackSource: !!options.fallbackSource }))
    },
  })
}

async function confirmGenerateIepPlan() {
  const preConfirm = await resolveGeneratePreConfirm()
  if (preConfirm === false)
    return
  if (preConfirm) {
    showGeneratePreConfirm(preConfirm, { fallbackSource: true })
    return
  }
  Modal.confirm({
    title: 'AI智能生成IEP计划',
    content: `将基于${generationSourceText.value}生成IEP计划。确认要继续吗？`,
    okText: '确认生成',
    cancelText: '先不生成',
    okButtonProps: { type: 'primary' },
    closable: true,
    centered: true,
    onOk() {
      runAfterConfirmClosed(() => generateAIPlan({ skipPrecheck: true }))
    },
  })
}

function runAfterConfirmClosed(callback) {
  window.setTimeout(() => {
    callback()
  }, 0)
}

function isConfirmDialogDismiss(args) {
  return args.length > 0
}

async function generateWeeklyPlan(skipConfirm = false, options = {}) {
  if (generatingExecutionPlan.value)
    return
  const forceRegenerate = !!options.forceRegenerate
  if (selectedExecutionWeekGenerated.value && !forceRegenerate) {
    messageService.warning(`${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}周计划已生成，如需覆盖请进入该周后点击“重新生成”`)
    return
  }
  const sourcePlan = planPayloadForSave()
  if (!sourcePlan.rows.length) {
    messageService.warning('请先生成或填写IEP计划')
    return
  }
  if (!monthlyPlan.value?.rows?.length && !skipConfirm) {
    Modal.confirm({
      title: '当前还没有月度计划',
      content: `是否直接基于${planTitle.value}生成周计划？`,
      okText: '直接生成周计划',
      cancelText: '先生成月度计划',
      closable: true,
      centered: true,
      onOk() {
        runAfterConfirmClosed(() => generateWeeklyPlan(true))
      },
      onCancel(...args) {
        if (isConfirmDialogDismiss(args))
          return
        runAfterConfirmClosed(() => generateMonthlyPlan())
      },
    })
    return
  }
  executionPlanGeneratingType.value = 'weekly'
  executionPlanRequestKey += 1
  const requestKey = executionPlanRequestKey
  const abortController = new AbortController()
  executionPlanStreamAbortController.value = abortController
  generatingExecutionPlan.value = true
  aiStreamStatus.value = `正在准备${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}周计划生成上下文`
  aiStreamText.value = ''
  weeklyPlan.value = null
  executionPlanView.value = 'weekly'
  try {
    const result = await assessmentAdapter.value.generateExecutionPlanStream(
      {
        id: props.record?.id,
        durationMonths: planDuration.value,
        planType: 'weekly',
        targetMonthIndex: selectedExecutionMonth.value,
        targetWeekIndex: selectedExecutionWeekValue.value,
        sourcePlan,
        monthlyPlan: monthlyPlan.value,
      },
      {
        onStatus(message) {
          if (requestKey !== executionPlanRequestKey)
            return
          aiStreamStatus.value = message || `正在生成${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}周计划`
        },
        onDelta(text) {
          if (requestKey !== executionPlanRequestKey)
            return
          aiStreamStatus.value = `正在接收${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}周计划内容`
          aiStreamText.value += text
          const partialPlan = buildStreamingWeeklyPlanFromText(aiStreamText.value)
          if (partialPlan)
            weeklyPlan.value = partialPlan
        },
        onDone(data) {
          if (requestKey !== executionPlanRequestKey)
            return
          weeklyPlan.value = data
          aiStreamStatus.value = `${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}周计划生成完成`
        },
      },
      {
        signal: abortController.signal,
      },
    )
    if (requestKey !== executionPlanRequestKey)
      return
    weeklyPlan.value = result
    await persistExecutionPlan('weekly', result)
    executionPlanView.value = 'weekly'
    await scrollPlanViewToTop()
    messageService.success(monthlyPlan.value?.rows?.length ? `AI已生成${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}周计划` : `AI已直接基于IEP生成${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}周计划`)
  }
  catch (error) {
    if (requestKey !== executionPlanRequestKey || isAbortError(error))
      return
    console.error('generate weekly plan failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '生成周计划失败')
  }
  finally {
    if (requestKey === executionPlanRequestKey) {
      generatingExecutionPlan.value = false
      executionPlanGeneratingType.value = ''
      executionPlanStreamAbortController.value = null
    }
  }
}

function confirmRegenerateWeeklyPlan() {
  if (!selectedExecutionWeekGenerated.value) {
    generateWeeklyPlan()
    return
  }
  Modal.confirm({
    title: `重新生成${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}周计划`,
    content: `重新生成会覆盖当前${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}周计划内容，已导出的Word不受影响。确认要继续吗？`,
    okText: '重新生成',
    cancelText: '保留当前计划',
    closable: true,
    okButtonProps: { danger: true },
    centered: true,
    onOk() {
      runAfterConfirmClosed(() => generateWeeklyPlan(true, { forceRegenerate: true }))
    },
  })
}

function clearSelectedMonthWeeklyPlans() {
  const monthKey = `${clampExecutionMonth(selectedExecutionMonth.value)}-`
  const nextPlans = { ...weeklyPlans.value }
  Object.keys(nextPlans).forEach((key) => {
    if (key.startsWith(monthKey))
      delete nextPlans[key]
  })
  weeklyPlans.value = nextPlans
}

function selectExecutionNavigatorItem(type, monthIndex, weekIndex) {
  if (aiGenerating.value || loadingSavedPlan.value || generatingExecutionPlan.value || savingExecutionPlan.value)
    return
  if (type === 'iep') {
    finishEditPlan()
    executionPlanView.value = 'iep'
    scrollPlanViewToTop()
    return
  }
  if (typeof monthIndex === 'number')
    selectedExecutionMonth.value = clampExecutionMonth(monthIndex)
  selectedExecutionWeek.value = Math.max(1, Math.min(weekCountForRange(executionMonthRangeForIndex(selectedExecutionMonth.value)), selectedExecutionWeek.value))
  if (type === 'monthly') {
    finishEditPlan()
    executionPlanView.value = 'monthly'
    scrollPlanViewToTop()
    return
  }
  if (typeof weekIndex === 'number')
    selectedExecutionWeek.value = clampExecutionWeek(weekIndex)
  finishEditPlan()
  executionPlanView.value = 'weekly'
  scrollPlanViewToTop()
}

function switchPreviewView(value) {
  const nextView = ['iep', 'monthly', 'weekly'].includes(value) ? value : 'iep'
  finishEditPlan()
  executionPlanView.value = nextView
  scrollPlanViewToTop()
}

async function persistExecutionPlan(planType, plan, options = {}) {
  if (!props.record?.id || !plan)
    return false
  try {
    const response = await assessmentAdapter.value.saveExecutionPlan({
      id: props.record.id,
      durationMonths: planDuration.value,
      planType,
      targetMonthIndex: selectedExecutionMonth.value,
      targetWeekIndex: planType === 'weekly' ? selectedExecutionWeekValue.value : 0,
      monthlyPlan: planType === 'monthly' ? plan : null,
      weeklyPlan: planType === 'weekly' ? plan : null,
      preserveWeeklyPlans: !!options.preserveWeeklyPlans,
    })
    applySavedExecutionPlansData(unwrapResponse(response))
    if (options.successMessage)
      messageService.success(options.successMessage)
    return true
  }
  catch (error) {
    console.error('save execution plan failed', error)
    messageService[options.manual ? 'error' : 'warning'](options.errorMessage || (options.manual ? '保存执行计划失败' : '执行计划已生成，但自动保存失败，请稍后重新生成或导出'))
    return false
  }
}

function applySavedExecutionPlansData(data) {
  const nextMonthlyPlans = {}
  const nextWeeklyPlans = {}
  const monthlyItems = data?.monthlyPlans || []
  const weeklyItems = data?.weeklyPlans || []
  monthlyItems.forEach((item) => {
    if (item?.targetMonthIndex && item.plan)
      nextMonthlyPlans[String(clampExecutionMonth(item.targetMonthIndex))] = applyAssessmentPlanDateToMonthlyPlan(item.plan)
  })
  weeklyItems.forEach((item) => {
    if (item?.targetMonthIndex && item?.targetWeekIndex && item.plan)
      nextWeeklyPlans[`${clampExecutionMonth(item.targetMonthIndex)}-${Math.max(1, Number(item.targetWeekIndex || 1))}`] = item.plan
  })
  monthlyPlans.value = nextMonthlyPlans
  weeklyPlans.value = nextWeeklyPlans
}

async function loadSavedExecutionPlans(durationMonths = planDuration.value) {
  if (!props.record?.id)
    return
  try {
    const response = await assessmentAdapter.value.getExecutionPlans(props.record.id, durationMonths)
    applySavedExecutionPlansData(unwrapResponse(response))
  }
  catch (error) {
    console.error('load execution plans failed', error)
    messageService.warning('已保存的月计划和周计划读取失败')
  }
}

function normalizePlanRows(rows = []) {
  const { start } = defaultPlanDateRange.value
  const stageRanges = buildStageDateRanges(start, Number(planDuration.value || 6))
  const groups = []
  const groupMap = new Map()

  rows.forEach((row) => {
    const rawDomain = String(row?.domain || '').trim()
    const rawLongGoal = String(row?.longGoal || '').trim()
    const rawShortGoal = String(row?.shortGoal || '').trim()
    if (!rawDomain && !rawLongGoal && !rawShortGoal)
      return
    const domain = rawDomain || '综合康复'
    let group = groupMap.get(domain)
    if (!group) {
      group = { domain, longGoals: [], shortGoals: [] }
      groupMap.set(domain, group)
      groups.push(group)
    }
    appendUniqueGoals(group.longGoals, splitGoalLines(rawLongGoal))
    splitGoalLines(rawShortGoal).forEach((shortGoal) => {
      appendUniqueShortGoals(group.shortGoals, {
        goal: shortGoal,
        courseForm: normalizeCourseForm(row?.courseForm),
      })
    })
  })

  return groups.flatMap((group) => {
    const longGoals = ensureLongGoalLines(group.domain, group.longGoals)
    const shortGoals = ensureShortGoalLines(group.domain, group.shortGoals)
    const longGoalText = longGoals.map((item, index) => `${index + 1}. ${item}`).join('\n')
    return shortGoals.map((shortGoal, index) => ({
      domain: group.domain,
      longGoal: longGoalText,
      shortGoal: shortGoal.goal,
      courseForm: shortGoal.courseForm || inferCourseFormFromText(shortGoal.goal, group.domain) || '个训',
      startEndDate: stageDateForGoal(stageRanges, index, shortGoals.length),
    }))
  })
}

function splitGoalLines(text = '') {
  const normalized = String(text)
    .replace(/\r\n/g, '\n')
    .replace(/\r/g, '\n')
    .replace(/[；;]/g, '\n')
    .replace(/(^|\s)(?:\d+[.、．]|[一二三四五六七八九十]+[、.．])\s*/g, '\n')
  return normalized
    .split('\n')
    .map(item => item.trim().replace(/^[，,。.；;、]+|[，,。.；;、]+$/g, ''))
    .filter(Boolean)
}

function appendUniqueGoals(target, additions = []) {
  additions.forEach((item) => {
    if (item && !target.includes(item))
      target.push(item)
  })
}

function appendUniqueShortGoals(target, item) {
  const goal = String(item?.goal || '').trim()
  if (!goal || target.some(existing => existing.goal === goal))
    return
  target.push({
    goal,
    courseForm: normalizeCourseForm(item?.courseForm),
  })
}

function ensureLongGoalLines(domain, goals = []) {
  const result = [...goals]
  appendUniqueGoals(result, [
    `提升${domain}相关核心能力，能在适当提示下稳定参与并完成目标任务。`,
    `将${domain}训练内容泛化到课堂活动、同伴互动和日常生活中，提高主动性、持续性和独立完成度。`,
  ])
  return result.slice(0, 2)
}

function ensureShortGoalLines(domain, goals = []) {
  const result = [...goals]
  const fallbacks = [
    `在教师示范和语言提示下，能参与${domain}相关活动并完成基础目标任务，连续3次课程中达到70%以上。`,
    `在少量提示下，能将${domain}目标应用到对应训练流程中，连续3次课程中达到75%以上。`,
    `在自然活动中，能较稳定完成${domain}目标并减少成人辅助，连续3次课程中达到80%以上。`,
  ]
  fallbacks.forEach((item) => {
    if (result.length < 3)
      appendUniqueShortGoals(result, { goal: item, courseForm: inferCourseFormFromText(item, domain) })
  })
  return result
}

function normalizeCourseForm(value = '') {
  const text = String(value || '').trim()
  if (!text)
    return ''
  if (/一对一|1对1|1v1|个训|个别/i.test(text))
    return '个训'
  if (/集体|小组|团体|融合/.test(text))
    return '集体课'
  return text.length <= 8 ? text : ''
}

function inferCourseFormFromText(...parts) {
  return normalizeCourseForm(parts.filter(Boolean).join(' '))
}

function parseJsonStringLiteral(raw = '') {
  try {
    return JSON.parse(`"${raw}"`)
  }
  catch {
    return raw.replace(/\\"/g, '"').replace(/\\\\/g, '\\')
  }
}

function extractJsonStringField(text, key) {
  const match = text.match(new RegExp(`"${key}"\\s*:\\s*"((?:\\\\.|[^"\\\\])*)"`))
  return match ? parseJsonStringLiteral(match[1]) : ''
}

function extractRowsArrayText(text) {
  const keyIndex = text.indexOf('"rows"')
  if (keyIndex < 0)
    return ''
  const arrayStart = text.indexOf('[', keyIndex)
  if (arrayStart < 0)
    return ''
  return text.slice(arrayStart + 1)
}

function collectCompleteJsonObjects(text) {
  const objects = []
  let start = -1
  let depth = 0
  let inString = false
  let escaped = false
  for (let index = 0; index < text.length; index++) {
    const char = text[index]
    if (inString) {
      if (escaped)
        escaped = false
      else if (char === '\\')
        escaped = true
      else if (char === '"')
        inString = false
      continue
    }
    if (char === '"') {
      inString = true
      continue
    }
    if (char === '{') {
      if (depth === 0)
        start = index
      depth++
      continue
    }
    if (char === '}' && depth > 0) {
      depth--
      if (depth === 0 && start >= 0) {
        objects.push(text.slice(start, index + 1))
        start = -1
      }
    }
  }
  return objects
}

function buildStreamingPlanFromText(text) {
  const content = String(text || '').trim()
  if (!content)
    return null
  try {
    const parsed = JSON.parse(extractCompleteJsonContent(content))
    return parsed?.rows?.length ? normalizeStreamingMonthlyPlanPreview(parsed) : null
  }
  catch {
  }

  const rows = collectCompleteJsonObjects(extractRowsArrayText(content))
    .map((item) => {
      try {
        return JSON.parse(item)
      }
      catch {
        return null
      }
    })
    .filter(Boolean)

  if (!rows.length)
    return null

  return {
    title: extractJsonStringField(content, 'title') || planTitle.value,
    student: {
      name: props.record?.studentName || '',
      gender: props.record?.studentGender || '',
      birthDate: formatDate(props.record?.birthDate) || '',
    },
    meta: {
      planDate: assessmentPlanDate.value || defaultPlanDateRange.value.start,
      participant: currentTeacherName.value,
      implementer: currentTeacherName.value,
      startDate: defaultPlanDateRange.value.start,
      endDate: defaultPlanDateRange.value.end,
    },
    rows,
    model: 'AI生成中',
  }
}

function buildExecutionPreviewWeekDates() {
  const range = selectedExecutionWeekRange.value
  const base = new Date(`${range.start || selectedExecutionMonthRange.value.start || defaultPlanDateRange.value.start || formatDate(new Date())}T00:00:00`)
  const end = new Date(`${range.end || range.start}T00:00:00`)
  const start = Number.isNaN(base.getTime()) ? new Date() : base
  return Array.from({ length: 6 }, (_, index) => {
    const day = new Date(start.getFullYear(), start.getMonth(), start.getDate() + index)
    return !Number.isNaN(end.getTime()) && day > end ? '' : formatLocalDate(day)
  })
}

function buildStreamingMonthlyPlanFromText(text) {
  const content = String(text || '').trim()
  if (!content)
    return null
  try {
    const parsed = JSON.parse(extractCompleteJsonContent(content))
    return parsed?.rows?.length ? normalizeStreamingMonthlyPlanPreview(parsed) : null
  }
  catch {
  }

  const rows = collectCompleteJsonObjects(extractRowsArrayText(content))
    .map((item) => {
      try {
        return JSON.parse(item)
      }
      catch {
        return null
      }
    })
    .filter(Boolean)

  if (!rows.length)
    return null

  return normalizeStreamingMonthlyPlanPreview({
    title: extractJsonStringField(content, 'title') || `康复教学${selectedExecutionMonthLabel.value}计划`,
    student: {
      name: props.record?.studentName || '',
      gender: props.record?.studentGender || '',
      birthDate: formatDate(props.record?.birthDate) || '',
    },
    meta: {
      planDate: assessmentPlanDate.value || defaultPlanDateRange.value.start,
      participant: currentTeacherName.value,
      implementer: currentTeacherName.value,
      startDate: selectedExecutionMonthRange.value.start,
      endDate: selectedExecutionMonthRange.value.end,
      monthLabel: selectedExecutionMonthLabel.value,
      sourceTitle: planTitle.value,
    },
    rows,
    model: 'AI生成中',
  })
}

function normalizeStreamingMonthlyPlanPreview(plan = {}) {
  const range = selectedExecutionMonthRange.value
  return {
    ...plan,
    title: String(plan.title || '').trim() || `康复教学${selectedExecutionMonthLabel.value}计划`,
    student: {
      name: plan.student?.name || props.record?.studentName || '',
      gender: plan.student?.gender || props.record?.studentGender || '',
      birthDate: plan.student?.birthDate || formatDate(props.record?.birthDate) || '',
    },
    meta: {
      ...(plan.meta || {}),
      planDate: assessmentPlanDate.value || plan.meta?.planDate || range.start,
      participant: plan.meta?.participant || currentTeacherName.value,
      implementer: plan.meta?.implementer || currentTeacherName.value,
      startDate: plan.meta?.startDate || range.start,
      endDate: plan.meta?.endDate || range.end,
      monthLabel: plan.meta?.monthLabel || selectedExecutionMonthLabel.value,
      sourceTitle: plan.meta?.sourceTitle || planTitle.value,
    },
    rows: Array.isArray(plan.rows) ? plan.rows : [],
    model: plan.model || 'AI生成中',
  }
}

function buildStreamingWeeklyPlanFromText(text) {
  const content = String(text || '').trim()
  if (!content)
    return null
  try {
    const parsed = JSON.parse(extractCompleteJsonContent(content))
    return parsed?.rows?.length ? normalizeStreamingWeeklyPlanPreview(parsed) : null
  }
  catch {
  }

  const rows = collectCompleteJsonObjects(extractRowsArrayText(content))
    .map((item) => {
      try {
        return JSON.parse(item)
      }
      catch {
        return null
      }
    })
    .filter(Boolean)

  if (!rows.length)
    return null

  const weekDates = buildExecutionPreviewWeekDates()
  return normalizeStreamingWeeklyPlanPreview({
    title: extractJsonStringField(content, 'title') || `康复教学周计划日记录卡${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}`,
    student: {
      name: props.record?.studentName || '',
      gender: props.record?.studentGender || '',
      birthDate: formatDate(props.record?.birthDate) || '',
    },
    teacherName: currentTeacherName.value,
    courseName: extractJsonStringField(content, 'courseName') || (monthlyPlan.value?.title || planTitle.value),
    trainingDate: `${weekDates[0]} 至 ${weekDates[weekDates.length - 1]}`,
    preparation: extractJsonStringField(content, 'preparation') || '正在生成训练前准备内容。',
    weekDates,
    rows: rows.map(row => ({
      ...row,
      completion: Array.isArray(row.completion) ? row.completion : Array.from({ length: weekDates.length }, () => ''),
    })),
    model: 'AI生成中',
  })
}

function normalizeStreamingWeeklyPlanPreview(plan = {}) {
  const weekDates = buildExecutionPreviewWeekDates()
  const visibleDates = weekDates.filter(Boolean)
  const startDate = visibleDates[0] || selectedExecutionWeekRange.value.start
  const endDate = visibleDates[visibleDates.length - 1] || selectedExecutionWeekRange.value.end || startDate
  return {
    ...plan,
    title: String(plan.title || '').trim() || `康复教学周计划日记录卡${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}`,
    student: {
      name: plan.student?.name || props.record?.studentName || '',
      gender: plan.student?.gender || props.record?.studentGender || '',
      birthDate: plan.student?.birthDate || formatDate(props.record?.birthDate) || '',
    },
    teacherName: plan.teacherName || currentTeacherName.value,
    courseName: plan.courseName || (monthlyPlan.value?.title || planTitle.value),
    trainingDate: plan.trainingDate || `${startDate} 至 ${endDate}`,
    preparation: plan.preparation || '正在生成训练前准备内容。',
    weekDates,
    rows: (Array.isArray(plan.rows) ? plan.rows : []).map(row => ({
      ...row,
      completion: Array.isArray(row.completion) ? row.completion : Array.from({ length: weekDates.length }, () => ''),
    })),
    sourceTitle: plan.sourceTitle || (monthlyPlan.value?.title || planTitle.value),
    model: plan.model || 'AI生成中',
  }
}

function extractCompleteJsonContent(text) {
  const start = text.indexOf('{')
  const end = text.lastIndexOf('}')
  if (start >= 0 && end > start)
    return text.slice(start, end + 1)
  return text
}

function stageDateForGoal(stageRanges = [], goalIndex = 0, goalCount = 1) {
  if (!stageRanges.length)
    return ''
  const rangeIndex = Math.min(stageRanges.length - 1, Math.floor(goalIndex * stageRanges.length / Math.max(goalCount, 1)))
  return stageRanges[rangeIndex]
}

function unwrapResponse(res) {
  return res?.data ?? res?.result ?? res
}

function normalizePlanDurationValue(value) {
  return String(value) === '6' ? '6' : '3'
}

function normalizePlanSheetTitle(title = '') {
  const text = String(title || '').trim()
  if (planDuration.value === '3' && (!text || text === '康复教学三个月计划'))
    return planTitle.value
  return text || planTitle.value
}

function deepClone(value) {
  return JSON.parse(JSON.stringify(value || {}))
}

function createPlanSheetFromPlan(plan = {}, options = {}) {
  const { start, end } = defaultPlanDateRange.value
  const rows = options.preserveRows
    ? sanitizeEditablePlanRows(plan.rows)
    : normalizePlanRows(plan.rows)
  return {
    title: normalizePlanSheetTitle(plan.title),
    student: {
      name: plan.student?.name || props.record?.studentName || '',
      gender: plan.student?.gender || props.record?.studentGender || '',
      birthDate: plan.student?.birthDate || formatDate(props.record?.birthDate) || '',
    },
    meta: {
      ...(plan.meta || {}),
      planDate: assessmentPlanDate.value || plan.meta?.planDate || start,
      participant: currentTeacherName.value,
      implementer: plan.meta?.implementer || currentTeacherName.value,
      startDate: plan.meta?.startDate || start,
      endDate: plan.meta?.endDate || end,
    },
    rows,
    model: options.model ?? plan.model ?? '',
  }
}

function createEditablePlanFromPlan(plan = {}, preserveRows = true) {
  return deepClone(createPlanSheetFromPlan(plan, { preserveRows, model: plan.model || 'AI草案' }))
}

function applyAssessmentPlanDateToMonthlyPlan(plan = {}) {
  const nextPlan = deepClone(plan)
  nextPlan.meta = {
    ...(nextPlan.meta || {}),
    planDate: assessmentPlanDate.value || nextPlan.meta?.planDate || defaultPlanDateRange.value.start,
  }
  return nextPlan
}

function sanitizeEditablePlanRows(rows = []) {
  const { start } = defaultPlanDateRange.value
  const stageRanges = buildStageDateRanges(start, Number(planDuration.value || 6))
  return (rows || [])
    .map((row, index) => {
      const domain = String(row?.domain || '').trim()
      const longGoal = String(row?.longGoal || '').trim()
      const shortGoal = String(row?.shortGoal || '').trim()
      if (!shortGoal)
        return null
      return {
        domain: domain || '综合康复',
        longGoal,
        shortGoal,
        courseForm: normalizeCourseForm(row?.courseForm) || '个训',
        startEndDate: String(row?.startEndDate || '').trim() || stageDateForGoal(stageRanges, index, rows.length),
      }
    })
    .filter(Boolean)
}

function ensureEditablePlan() {
  if (!editablePlan.value)
    editablePlan.value = createEditablePlanFromPlan(planSheet.value, true)
  if (!editablePlan.value.rows)
    editablePlan.value.rows = []
  if (!editablePlan.value.meta)
    editablePlan.value.meta = {}
  return editablePlan.value
}

function updatePlanRow(index, patch = {}) {
  const plan = ensureEditablePlan()
  if (!plan.rows[index])
    return
  plan.rows[index] = {
    ...plan.rows[index],
    ...patch,
  }
}

function selectPlanRow(index) {
  if (!isPlanEditable.value)
    return
  const rows = planRows.value || []
  selectedPlanRowIndex.value = Math.max(0, Math.min(index, rows.length - 1))
}

function selectMonthlyPlanRow(rowIndex, contentIndex) {
  if (!isMonthlyPlanEditable.value)
    return
  const rows = monthlyDisplayRows.value || []
  const target = rows.find(row => row.rowIndex === rowIndex && row.contentIndex === contentIndex) || rows[0]
  selectedMonthlyRowIndex.value = target?.rowIndex || 0
  selectedMonthlyContentIndex.value = target?.contentIndex || 0
}

function selectWeeklyPlanRow(index) {
  if (!isWeeklyPlanEditable.value)
    return
  const rows = weeklyPlan.value?.rows || []
  selectedWeeklyRowIndex.value = Math.max(0, Math.min(index, rows.length - 1))
}

function updateSelectedPlanRow(patch = {}) {
  updatePlanRow(selectedPlanRowIndex.value, patch)
}

function updateSelectedPlanDomain(value) {
  const row = selectedPlanRow.value
  if (!row)
    return
  updatePlanGroupDomain(row.domain, value)
}

function updateSelectedPlanLongGoal(value) {
  const row = selectedPlanRow.value
  if (!row)
    return
  updatePlanGroupLongGoal(row.domain, value)
}

function updatePlanGroupDomain(oldDomain, value) {
  const plan = ensureEditablePlan()
  const nextDomain = String(value || '').trim() || '综合康复'
  plan.rows = plan.rows.map((row) => {
    if (row.domain !== oldDomain)
      return row
    return { ...row, domain: nextDomain }
  })
}

function updatePlanGroupLongGoal(domain, value) {
  const plan = ensureEditablePlan()
  plan.rows = plan.rows.map((row) => {
    if (row.domain !== domain)
      return row
    return { ...row, longGoal: value }
  })
}

function appendLongGoal(domain) {
  const plan = ensureEditablePlan()
  const targetRow = plan.rows.find(row => row.domain === domain) || plan.rows[0]
  if (!targetRow)
    return
  const lines = splitGoalLines(targetRow.longGoal)
  const nextLine = `${lines.length + 1}. `
  const nextValue = String(targetRow.longGoal || '').trim()
    ? `${String(targetRow.longGoal).trim()}\n${nextLine}`
    : nextLine
  updatePlanGroupLongGoal(targetRow.domain, nextValue)
}

function removeLongGoalLine(domain) {
  const plan = ensureEditablePlan()
  const targetRow = plan.rows.find(row => row.domain === domain) || plan.rows[0]
  if (!targetRow)
    return
  const lines = splitGoalLines(targetRow.longGoal)
  if (!lines.length)
    return
  lines.pop()
  const nextValue = lines.length ? lines.map((item, index) => `${index + 1}. ${item}`).join('\n') : ''
  updatePlanGroupLongGoal(targetRow.domain, nextValue)
}

function buildBlankShortGoalRow(baseRow = {}) {
  const rows = planRows.value || []
  const fallbackRow = rows[0] || {}
  const { start } = defaultPlanDateRange.value
  const stageRanges = buildStageDateRanges(start, Number(planDuration.value || 6))
  return {
    domain: baseRow.domain || fallbackRow.domain || '综合康复',
    longGoal: baseRow.longGoal || fallbackRow.longGoal || '',
    shortGoal: '',
    courseForm: normalizeCourseForm(baseRow.courseForm || fallbackRow.courseForm) || '个训',
    startEndDate: baseRow.startEndDate || stageDateForGoal(stageRanges, rows.length, rows.length + 1),
  }
}

function addShortGoalAfter(index = -1) {
  const plan = ensureEditablePlan()
  const insertIndex = Number.isInteger(index) && index >= 0 ? index + 1 : plan.rows.length
  const baseRow = plan.rows[Math.max(0, Math.min(index, plan.rows.length - 1))] || plan.rows[plan.rows.length - 1] || {}
  plan.rows.splice(insertIndex, 0, buildBlankShortGoalRow(baseRow))
}

function addShortGoalAfterSelected() {
  const currentIndex = selectedPlanRowIndex.value
  addShortGoalAfter(currentIndex)
  selectedPlanRowIndex.value = Math.min(currentIndex + 1, planRows.value.length - 1)
}

function deleteShortGoal(index) {
  const plan = ensureEditablePlan()
  if (index < 0 || index >= plan.rows.length)
    return
  plan.rows.splice(index, 1)
}

function deleteSelectedShortGoal() {
  const currentIndex = selectedPlanRowIndex.value
  deleteShortGoal(currentIndex)
  selectedPlanRowIndex.value = Math.max(0, Math.min(currentIndex, planRows.value.length - 1))
}

function updateMonthlyPlanStudent(patch = {}) {
  if (!monthlyPlan.value)
    return
  const plan = deepClone(monthlyPlan.value)
  plan.student = {
    ...(plan.student || {}),
    ...patch,
  }
  monthlyPlan.value = plan
}

function updateMonthlyPlanMeta(patch = {}) {
  if (!monthlyPlan.value)
    return
  const plan = deepClone(monthlyPlan.value)
  plan.meta = {
    ...(plan.meta || {}),
    ...patch,
  }
  monthlyPlan.value = plan
}

function updateMonthlyRow(rowIndex, patch = {}) {
  const plan = deepClone(monthlyPlan.value)
  if (!plan?.rows?.[rowIndex])
    return
  plan.rows[rowIndex] = {
    ...plan.rows[rowIndex],
    ...patch,
  }
  monthlyPlan.value = plan
}

function updateMonthlyGroupRows(domain, patch = {}) {
  const plan = deepClone(monthlyPlan.value)
  if (!Array.isArray(plan?.rows))
    return
  plan.rows = plan.rows.map((row) => {
    if (row.domain !== domain)
      return row
    return {
      ...row,
      ...patch,
    }
  })
  monthlyPlan.value = plan
}

function updateMonthlyTrainingItem(rowIndex, contentIndex, patch = {}) {
  const plan = deepClone(monthlyPlan.value)
  const row = plan?.rows?.[rowIndex]
  if (!row)
    return
  if (!Array.isArray(row.trainingItems))
    row.trainingItems = []
  while (row.trainingItems.length <= contentIndex)
    row.trainingItems.push({ content: '', startEndDate: '' })
  row.trainingItems[contentIndex] = {
    ...(row.trainingItems[contentIndex] || {}),
    ...patch,
  }
  monthlyPlan.value = plan
}

function updateWeeklyPlanStudent(patch = {}) {
  if (!weeklyPlan.value)
    return
  const plan = deepClone(weeklyPlan.value)
  plan.student = {
    ...(plan.student || {}),
    ...patch,
  }
  weeklyPlan.value = plan
}

function updateWeeklyPlanField(patch = {}) {
  if (!weeklyPlan.value)
    return
  weeklyPlan.value = {
    ...deepClone(weeklyPlan.value),
    ...patch,
  }
}

function updateWeeklyRow(rowIndex, patch = {}) {
  const plan = deepClone(weeklyPlan.value)
  if (!plan?.rows?.[rowIndex])
    return
  plan.rows[rowIndex] = {
    ...plan.rows[rowIndex],
    ...patch,
  }
  weeklyPlan.value = plan
}

function planPayloadForSave() {
  const plan = createPlanSheetFromPlan(planSheet.value, { preserveRows: true, model: planSheet.value.model || '' })
  plan.rows = sanitizeEditablePlanRows(plan.rows)
  return deepClone(plan)
}

function executionPlanPayloadForSave(planType) {
  const plan = planType === 'monthly' ? monthlyPlan.value : weeklyPlan.value
  if (planType === 'monthly')
    return applyAssessmentPlanDateToMonthlyPlan(plan)
  return deepClone(plan)
}

async function saveActiveExecutionPlan() {
  if (savingExecutionPlan.value || aiGenerating.value || generatingExecutionPlan.value || loadingSavedPlan.value)
    return
  const isMonthly = executionPlanView.value === 'monthly'
  const isWeekly = executionPlanView.value === 'weekly'
  if (isMonthly && !monthlyPlan.value?.rows?.length) {
    messageService.warning('请先生成月度计划')
    return
  }
  if (isWeekly && !weeklyPlan.value?.rows?.length) {
    messageService.warning('请先生成周计划')
    return
  }
  if (!isMonthly && !isWeekly)
    return

  savingExecutionPlan.value = true
  try {
    const planType = isMonthly ? 'monthly' : 'weekly'
    const success = await persistExecutionPlan(planType, executionPlanPayloadForSave(planType), {
      manual: true,
      preserveWeeklyPlans: isMonthly,
      successMessage: isMonthly ? `${selectedExecutionMonthLabel.value}计划修改已保存` : `${selectedExecutionMonthLabel.value}${selectedExecutionWeekLabel.value}周计划修改已保存`,
    })
    return success
  }
  finally {
    savingExecutionPlan.value = false
  }
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
  if (aiGenerating.value) {
    confirmCancelAIGeneration(() => {
      cancelAIGeneration()
      openModal.value = false
    })
    return
  }
  if (generatingExecutionPlan.value) {
    confirmCancelExecutionPlanGeneration(() => {
      cancelExecutionPlanGeneration()
      openModal.value = false
    })
    return
  }
  openModal.value = false
}

function forceCloseModal() {
  cancelAIGeneration()
  cancelExecutionPlanGeneration()
  openModal.value = false
}

function confirmCancelAIGeneration(onConfirm) {
  Modal.confirm({
    title: 'AI计划正在生成中',
    content: '关闭弹窗或离开页面会取消本次生成，当前生成结果不会保存。确认要取消生成吗？',
    okText: '取消生成并关闭',
    cancelText: '继续等待',
    okButtonProps: { danger: true },
    centered: true,
    onOk: onConfirm,
  })
}

function cancelAIGeneration() {
  aiGenerationRequestKey += 1
  if (aiStreamAbortController.value && !aiStreamAbortController.value.signal?.aborted)
    aiStreamAbortController.value.abort()
  aiStreamAbortController.value = null
  aiGenerating.value = false
}

function confirmCancelExecutionPlanGeneration(onConfirm) {
  Modal.confirm({
    title: `${executionPlanGeneratingLabel.value}正在生成中`,
    content: '关闭弹窗会取消本次生成，当前生成结果不会保存。确认要取消生成吗？',
    okText: '取消生成并关闭',
    cancelText: '继续等待',
    okButtonProps: { danger: true },
    centered: true,
    onOk: onConfirm,
  })
}

function cancelExecutionPlanGeneration() {
  executionPlanRequestKey += 1
  if (executionPlanStreamAbortController.value && !executionPlanStreamAbortController.value.signal?.aborted)
    executionPlanStreamAbortController.value.abort()
  executionPlanStreamAbortController.value = null
  generatingExecutionPlan.value = false
  executionPlanGeneratingType.value = ''
}

function isAbortError(error) {
  return error?.name === 'AbortError' || /aborted|abort|取消/.test(String(error?.message || ''))
}

function handleBeforeUnload(event) {
  if (!aiGenerating.value && !generatingExecutionPlan.value)
    return
  event.preventDefault()
  event.returnValue = ''
  return ''
}

function startEditPlan() {
  if (!canEditActivePreview.value || aiGenerating.value || generatingExecutionPlan.value || loadingSavedPlan.value || savingExecutionPlan.value)
    return
  if (executionPlanView.value === 'iep') {
    ensureEditablePlan()
    selectedPlanRowIndex.value = Math.max(0, Math.min(selectedPlanRowIndex.value, planRows.value.length - 1))
  }
  else if (executionPlanView.value === 'monthly') {
    const selected = selectedMonthlyDisplayRow.value || monthlyDisplayRows.value[0]
    selectedMonthlyRowIndex.value = selected?.rowIndex || 0
    selectedMonthlyContentIndex.value = selected?.contentIndex || 0
  }
  else if (executionPlanView.value === 'weekly') {
    const rows = weeklyPlan.value?.rows || []
    selectedWeeklyRowIndex.value = Math.max(0, Math.min(selectedWeeklyRowIndex.value, rows.length - 1))
  }
  editingPlanView.value = executionPlanView.value
}

function finishEditPlan() {
  editingPlanView.value = ''
}

function setPlanDurationWithoutLoading(value) {
  const nextValue = normalizePlanDurationValue(value)
  if (planDuration.value === nextValue)
    return
  ignoreNextPlanDurationWatch = true
  planDuration.value = nextValue
}

function clearDisplayedPlanState() {
  aiStreamStatus.value = ''
  aiStreamText.value = ''
  streamingPlan.value = null
  generatedPlan.value = null
  editablePlan.value = null
  savedPlanStatus.value = ''
  executionPlanView.value = 'iep'
  selectedExecutionMonth.value = 1
  selectedExecutionWeek.value = 1
  monthlyPlans.value = {}
  weeklyPlans.value = {}
  editingPlanView.value = ''
  selectedPlanRowIndex.value = 0
  selectedMonthlyRowIndex.value = 0
  selectedMonthlyContentIndex.value = 0
  selectedWeeklyRowIndex.value = 0
}

function schedulePlanLoadingOverlay(requestKey) {
  if (planLoadingOverlayTimer)
    window.clearTimeout(planLoadingOverlayTimer)
  showPlanLoadingOverlay.value = false
  planLoadingOverlayTimer = window.setTimeout(() => {
    if (requestKey === undefined || requestKey === loadPlanRequestKey)
      showPlanLoadingOverlay.value = true
  }, 180)
}

function hidePlanLoadingOverlay() {
  if (planLoadingOverlayTimer) {
    window.clearTimeout(planLoadingOverlayTimer)
    planLoadingOverlayTimer = 0
  }
  showPlanLoadingOverlay.value = false
}

async function scrollPlanViewToTop() {
  await nextTick()
  const target = iepModalBodyRef.value
  if (!target)
    return
  if (typeof target.scrollTo === 'function') {
    target.scrollTo({ top: 0, left: 0, behavior: 'auto' })
    return
  }
  target.scrollTop = 0
  target.scrollLeft = 0
}

function savedIepPlanHasContent(data) {
  const rows = data?.plan?.rows
  if (!data?.exists || !data.plan || !Array.isArray(rows))
    return false
  return rows.some(row => String(row?.shortGoal || '').trim())
}

function applySavedIepPlanData(data) {
  if (!data?.exists || !data.plan || !savedIepPlanHasContent(data)) {
    savedPlanStatus.value = ''
    generatedPlan.value = null
    editablePlan.value = null
    return false
  }
  savedPlanStatus.value = data.status || 'draft'
  generatedPlan.value = data.plan
  editablePlan.value = createEditablePlanFromPlan(data.plan, true)
  return true
}

function resetIepState() {
  executionPlanRequestKey += 1
  if (executionPlanStreamAbortController.value && !executionPlanStreamAbortController.value.signal?.aborted)
    executionPlanStreamAbortController.value.abort()
  executionPlanStreamAbortController.value = null
  hidePlanLoadingOverlay()
  setPlanDurationWithoutLoading('3')
  exportingWord.value = false
  aiGenerating.value = false
  generatingExecutionPlan.value = false
  executionPlanGeneratingType.value = ''
  loadingSavedPlan.value = false
  savingDraft.value = false
  confirmingPlan.value = false
  savingExecutionPlan.value = false
  clearDisplayedPlanState()
}

async function loadSavedIepPlan(requestKey, durationMonths = planDuration.value, options = {}) {
  if (!props.record?.id)
    return false
  const durationKey = normalizePlanDurationValue(durationMonths)
  loadingSavedPlan.value = true
  if (!options.preserveCurrentState) {
    hidePlanLoadingOverlay()
    clearDisplayedPlanState()
    scrollPlanViewToTop()
  }
  else {
    schedulePlanLoadingOverlay(requestKey)
  }
  try {
    const response = await assessmentAdapter.value.getIepPlan(props.record.id, durationKey)
    if (requestKey !== undefined && requestKey !== loadPlanRequestKey)
      return false
    if (planDuration.value !== durationKey)
      return false
    const data = unwrapResponse(response)
    clearDisplayedPlanState()
    if (applySavedIepPlanData(data))
      await loadSavedExecutionPlans(durationKey)
    await scrollPlanViewToTop()
    return true
  }
  catch (error) {
    if (requestKey !== undefined && requestKey !== loadPlanRequestKey)
      return false
    console.error('load iep plan failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '加载IEP计划失败')
    return false
  }
  finally {
    if (requestKey === undefined || requestKey === loadPlanRequestKey) {
      hidePlanLoadingOverlay()
      loadingSavedPlan.value = false
    }
  }
}

async function loadFirstAvailableIepPlan(requestKey) {
  if (!props.record?.id)
    return
  loadingSavedPlan.value = true
  clearDisplayedPlanState()
  scrollPlanViewToTop()
  try {
    for (const durationKey of ['3', '6']) {
      const response = await assessmentAdapter.value.getIepPlan(props.record.id, durationKey)
      if (requestKey !== undefined && requestKey !== loadPlanRequestKey)
        return
      const data = unwrapResponse(response)
      if (!savedIepPlanHasContent(data))
        continue
      setPlanDurationWithoutLoading(durationKey)
      applySavedIepPlanData(data)
      await loadSavedExecutionPlans(durationKey)
      await scrollPlanViewToTop()
      return
    }
    setPlanDurationWithoutLoading('3')
    clearDisplayedPlanState()
    await scrollPlanViewToTop()
  }
  catch (error) {
    if (requestKey !== undefined && requestKey !== loadPlanRequestKey)
      return
    console.error('load iep plan failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '加载IEP计划失败')
  }
  finally {
    if (requestKey === undefined || requestKey === loadPlanRequestKey)
      loadingSavedPlan.value = false
  }
}

async function persistIepPlan(status, options = {}) {
  if (aiGenerating.value || savingDraft.value || confirmingPlan.value)
    return false
  if (!props.record?.id) {
    messageService.warning('请先选择评估记录')
    return false
  }
  const payloadPlan = planPayloadForSave()
  if (!payloadPlan.rows.length) {
    messageService.warning('请先点击AI生成IEP计划，或填写至少一条计划内容')
    return false
  }
  const isConfirming = status === 'confirmed'
  if (isConfirming)
    confirmingPlan.value = true
  else
    savingDraft.value = true
  try {
    const response = await assessmentAdapter.value.saveIepPlan({
      id: props.record.id,
      durationMonths: planDuration.value,
      status,
      plan: payloadPlan,
    })
    const data = unwrapResponse(response)
    if (data?.durationMonths && normalizePlanDurationValue(data.durationMonths) !== planDuration.value)
      setPlanDurationWithoutLoading(data.durationMonths)
    savedPlanStatus.value = data?.status || status
    if (data?.plan) {
      generatedPlan.value = data.plan
      editablePlan.value = createEditablePlanFromPlan(data.plan, true)
    }
    emit(isConfirming ? 'confirmed' : 'saved', data)
    messageService.success(options.successMessage || (isConfirming ? (options.closeAfterSave ? 'IEP已确认生成' : '修改已保存') : '草稿已保存'))
    if (options.closeAfterSave)
      forceCloseModal()
    return true
  }
  catch (error) {
    console.error('save iep plan failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '保存IEP计划失败')
    return false
  }
  finally {
    if (isConfirming)
      confirmingPlan.value = false
    else
      savingDraft.value = false
  }
}

function saveIepDraft() {
  return persistIepPlan('draft')
}

function confirmIepPlan() {
  return persistIepPlan('confirmed', { closeAfterSave: true })
}

async function saveEditablePlan() {
  if (isMonthlyPlanEditable.value || isWeeklyPlanEditable.value) {
    await saveActiveExecutionPlan()
    return
  }
  if (!isPlanEditable.value)
    return
  const status = savedPlanStatus.value === 'confirmed' ? 'confirmed' : 'draft'
  const success = await persistIepPlan(status, {
    successMessage: status === 'confirmed' ? '修改已保存' : '草稿已保存',
  })
  return success
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
  if (!planRows.value.length) {
    messageService.warning('请先生成或填写IEP计划')
    return
  }
  const studentName = props.record?.studentName || '张一鸣'
  exportingWord.value = true
  try {
    const response = await assessmentAdapter.value.downloadIepPlanWord({
      id: props.record?.id,
      duration: planDuration.value,
      plan: planPayloadForSave(),
    })
    triggerDownload(response, `${studentName}-${planTitle.value}.docx`)
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

async function exportExecutionPlanWord() {
  if (exportingWord.value)
    return
  const isMonthly = executionPlanView.value === 'monthly'
  const isWeekly = executionPlanView.value === 'weekly'
  if (isMonthly && !monthlyPlan.value) {
    messageService.warning('请先生成月度计划')
    return
  }
  if (isWeekly && !weeklyPlan.value) {
    messageService.warning('请先生成周计划')
    return
  }
  exportingWord.value = true
  try {
    const response = await assessmentAdapter.value.downloadExecutionPlanWord({
      id: props.record?.id,
      planType: isMonthly ? 'monthly' : 'weekly',
      monthlyPlan: isMonthly ? monthlyPlan.value : null,
      weeklyPlan: isWeekly ? weeklyPlan.value : null,
    })
    triggerDownload(response, `${props.record?.studentName || '学员'}-${activePreviewTitle.value}.docx`)
    messageService.success('导出成功')
  }
  catch (error) {
    console.error('export execution plan failed', error)
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
    messageService.error(error?.response?.data?.message || error?.message || '导出失败')
  }
  finally {
    exportingWord.value = false
  }
}

async function generateAIPlan(options = {}) {
  if (aiGenerating.value)
    return
  if (!props.record?.id) {
    messageService.warning('请先选择评估记录')
    return
  }
  if (!options.skipPrecheck) {
    const preConfirm = await resolveGeneratePreConfirm()
    if (preConfirm === false)
      return
    if (preConfirm) {
      showGeneratePreConfirm(preConfirm, { fallbackSource: true })
      return
    }
  }
  const generationBasisText = options.fallbackSource ? fallbackGenerationBasisText.value : assessmentAdapter.value.generationBasisText
  const requestKey = aiGenerationRequestKey + 1
  aiGenerationRequestKey = requestKey
  const abortController = new AbortController()
  aiStreamAbortController.value = abortController
  aiGenerating.value = true
  aiStreamStatus.value = `正在准备${generationBasisText}`
  aiStreamText.value = ''
  streamingPlan.value = null
  generatedPlan.value = null
  editablePlan.value = null
  savedPlanStatus.value = ''
  executionPlanView.value = 'iep'
  selectedExecutionMonth.value = 1
  selectedExecutionWeek.value = 1
  monthlyPlans.value = {}
  weeklyPlans.value = {}
  editingPlanView.value = ''
  try {
    const plan = await assessmentAdapter.value.generateIepPlanStream(
      {
        id: Number(props.record.id),
        durationMonths: Number(planDuration.value),
      },
      {
        onStatus(message) {
          if (requestKey !== aiGenerationRequestKey)
            return
          aiStreamStatus.value = message || '正在生成'
        },
        onDelta(text) {
          if (requestKey !== aiGenerationRequestKey)
            return
          aiStreamStatus.value = '正在接收AI生成内容'
          aiStreamText.value += text
          const partialPlan = buildStreamingPlanFromText(aiStreamText.value)
          if (partialPlan)
            streamingPlan.value = partialPlan
        },
        onDone(data) {
          if (requestKey !== aiGenerationRequestKey)
            return
          generatedPlan.value = data
          editablePlan.value = createEditablePlanFromPlan(data, true)
          streamingPlan.value = null
          aiStreamStatus.value = '生成完成'
        },
      },
      {
        signal: abortController.signal,
      },
    )
    if (requestKey !== aiGenerationRequestKey)
      return
    if (!plan) {
      messageService.error('AI生成失败')
      return
    }
    generatedPlan.value = plan
    editablePlan.value = createEditablePlanFromPlan(plan, true)
    aiGenerating.value = false
    aiStreamAbortController.value = null
    await persistIepPlan('draft', { successMessage: 'AI生成成功，已自动保存草稿' })
  }
  catch (error) {
    if (requestKey !== aiGenerationRequestKey || isAbortError(error))
      return
    console.error('generate iep plan failed', error)
    aiStreamStatus.value = '生成失败'
    messageService.error(error?.response?.data?.message || error?.message || 'AI生成失败')
  }
  finally {
    if (requestKey === aiGenerationRequestKey) {
      aiGenerating.value = false
      aiStreamAbortController.value = null
    }
  }
}

watch(
  () => [props.open, props.record?.id],
  async ([open]) => {
    loadPlanRequestKey += 1
    const requestKey = loadPlanRequestKey
    if (!open) {
      resetIepState()
      return
    }
    resetIepState()
    await loadFirstAvailableIepPlan(requestKey)
  },
  { immediate: true },
)

watch(planDuration, async (durationMonths, previousDuration) => {
  if (ignoreNextPlanDurationWatch) {
    ignoreNextPlanDurationWatch = false
    return
  }
  if (!props.open)
    return
  if (aiGenerating.value || generatingExecutionPlan.value) {
    setPlanDurationWithoutLoading(previousDuration || '3')
    return
  }
  loadPlanRequestKey += 1
  const requestKey = loadPlanRequestKey
  const loaded = await loadSavedIepPlan(requestKey, durationMonths, { preserveCurrentState: true })
  if (!loaded && requestKey === loadPlanRequestKey && planDuration.value === durationMonths)
    setPlanDurationWithoutLoading(previousDuration || '3')
})

watch(
  () => planRows.value.length,
  (length) => {
    if (!length) {
      selectedPlanRowIndex.value = 0
      if (editingPlanView.value === 'iep')
        editingPlanView.value = ''
      return
    }
    selectedPlanRowIndex.value = Math.max(0, Math.min(selectedPlanRowIndex.value, length - 1))
  },
)

watch(
  () => monthlyDisplayRows.value.length,
  (length) => {
    if (!length) {
      selectedMonthlyRowIndex.value = 0
      selectedMonthlyContentIndex.value = 0
      if (editingPlanView.value === 'monthly')
        editingPlanView.value = ''
      return
    }
    const current = monthlyDisplayRows.value.find(row => row.rowIndex === selectedMonthlyRowIndex.value && row.contentIndex === selectedMonthlyContentIndex.value)
    if (!current) {
      selectedMonthlyRowIndex.value = monthlyDisplayRows.value[0]?.rowIndex || 0
      selectedMonthlyContentIndex.value = monthlyDisplayRows.value[0]?.contentIndex || 0
    }
  },
)

watch(
  () => weeklyPlan.value?.rows?.length || 0,
  (length) => {
    if (!length) {
      selectedWeeklyRowIndex.value = 0
      if (editingPlanView.value === 'weekly')
        editingPlanView.value = ''
      return
    }
    selectedWeeklyRowIndex.value = Math.max(0, Math.min(selectedWeeklyRowIndex.value, length - 1))
  },
)

onMounted(() => {
  window.addEventListener('beforeunload', handleBeforeUnload)
})

onBeforeUnmount(() => {
  cancelAIGeneration()
  hidePlanLoadingOverlay()
  if (executionPlanStreamAbortController.value && !executionPlanStreamAbortController.value.signal?.aborted)
    executionPlanStreamAbortController.value.abort()
  executionPlanStreamAbortController.value = null
  window.removeEventListener('beforeunload', handleBeforeUnload)
})

</script>

<template>
  <a-modal
    v-model:open="openModal"
    :centered="true"
    :closable="false"
    :footer="null"
    :keyboard="false"
    :mask-closable="false"
    :width="modalWidth"
    :body-style="{ padding: 0 }"
    wrap-class-name="generate-iep-modal-wrap"
  >
    <section class="iep-modal">
      <header class="iep-modal__header">
        <div class="iep-modal__title-block">
          <h2>{{ modalTitleText }}</h2>
          <div class="iep-header-meta">
            <span class="iep-header-meta__student">{{ studentMeta }}</span>
            <span v-if="headerPlanMeta" class="iep-header-meta__plan">{{ headerPlanMeta }}</span>
            <div class="period-switch">
              <span>计划周期</span>
              <a-segmented
                v-model:value="planDuration"
                :options="[
                  { label: '3个月', value: '3' },
                  { label: '6个月', value: '6' },
                ]"
                :disabled="aiGenerating || generatingExecutionPlan || savingDraft || confirmingPlan || savingExecutionPlan"
              />
              <em>{{ periodHint }}</em>
            </div>
            <div class="summary-count">
              {{ headerPlanStatusText }}
            </div>
          </div>
        </div>
        <a-button type="text" class="iep-close-btn" @click="closeModal">
          <template #icon>
            <CloseOutlined />
          </template>
        </a-button>
      </header>

      <div v-if="aiGenerating || aiStreamStatus || generatingExecutionPlan" class="ai-stream-bar">
        <span class="ai-stream-bar__dot" :class="{ 'is-running': aiGenerating || generatingExecutionPlan }" />
        <strong>{{ generationInlineStatus }}</strong>
        <span v-if="generationInlineTail" class="ai-stream-bar__text">{{ generationInlineTail }}</span>
      </div>

      <div class="iep-preview-toolbar">
        <div class="iep-preview-toolbar__info">
          <div class="iep-preview-toolbar__title">
            <span>A4预览</span>
            <strong>{{ activePreviewTitle }}</strong>
          </div>
          <div class="iep-preview-toolbar__meta">
            <span>{{ activePreviewCountText }}</span>
            <span v-if="executionPlanSourceText">{{ executionPlanSourceText }}</span>
            <span v-if="isAnyPlanEditable" class="iep-preview-toolbar__editing">正在编辑：{{ activeEditingLabel }}</span>
          </div>
        </div>
        <div class="iep-preview-toolbar__actions">
          <a-tooltip v-if="isIepPreview && !planRows.length" :title="`基于${generationSourceText}`">
            <span class="iep-toolbar-tooltip-target">
              <a-button
                size="small"
                type="primary"
                :loading="aiGenerating"
                :disabled="generatingExecutionPlan || loadingSavedPlan || savingDraft || confirmingPlan || savingExecutionPlan || isAnyPlanEditable"
                @click="confirmGenerateIepPlan"
              >
                AI智能生成
              </a-button>
            </span>
          </a-tooltip>
          <a-button
            v-if="!isIepPreview && !selectedExecutionMonthGenerated"
            size="small"
            type="primary"
            :loading="generatingExecutionPlan && executionPlanGeneratingType === 'monthly'"
            :disabled="!planRows.length || navigationDisabled"
            @click="confirmGenerateMonthlyPlan"
          >
            生成{{ selectedExecutionMonthLabel }}计划
          </a-button>
          <a-button
            v-if="!isIepPreview && !selectedExecutionWeekGenerated"
            size="small"
            type="primary"
            :loading="generatingExecutionPlan && executionPlanGeneratingType === 'weekly'"
            :disabled="!planRows.length || navigationDisabled"
            @click="() => generateWeeklyPlan()"
          >
            生成{{ selectedExecutionMonthLabel }}{{ selectedExecutionWeekLabel }}计划
          </a-button>
          <a-button
            v-if="isAnyPlanEditable"
            size="small"
            type="primary"
            :loading="savingExecutionPlan || savingDraft || confirmingPlan"
            :disabled="loadingSavedPlan || aiGenerating || generatingExecutionPlan"
            @click="saveEditablePlan"
          >
            保存修改
          </a-button>
          <a-button
            v-if="isAnyPlanEditable"
            size="small"
            :disabled="savingExecutionPlan || savingDraft || confirmingPlan || aiGenerating || generatingExecutionPlan"
            @click="finishEditPlan"
          >
            退出编辑
          </a-button>
          <a-button
            v-else-if="canEditActivePreview"
            size="small"
            :disabled="loadingSavedPlan || aiGenerating || generatingExecutionPlan || savingExecutionPlan"
            @click="startEditPlan"
          >
            编辑计划
          </a-button>
          <a-button
            v-if="isIepPreview && planRows.length"
            size="small"
            danger
            :loading="aiGenerating"
            :disabled="loadingSavedPlan || generatingExecutionPlan || savingDraft || confirmingPlan || savingExecutionPlan || isAnyPlanEditable"
            @click="confirmRegenerateIepPlan"
          >
            重新生成
          </a-button>
          <a-button
            v-if="executionPlanView === 'monthly' && selectedExecutionMonthGenerated"
            size="small"
            danger
            :disabled="loadingSavedPlan || aiGenerating || generatingExecutionPlan || savingExecutionPlan || isAnyPlanEditable"
            @click="confirmRegenerateMonthlyPlan"
          >
            重新生成
          </a-button>
          <a-button
            v-if="executionPlanView === 'weekly' && selectedExecutionWeekGenerated"
            size="small"
            danger
            :disabled="loadingSavedPlan || aiGenerating || generatingExecutionPlan || savingExecutionPlan || isAnyPlanEditable"
            @click="confirmRegenerateWeeklyPlan"
          >
            重新生成
          </a-button>
          <a-button
            size="small"
            :disabled="!canExportActivePreview || loadingSavedPlan || aiGenerating || generatingExecutionPlan || savingExecutionPlan"
            :loading="exportingWord"
            @click="isIepPreview ? exportIepWord() : exportExecutionPlanWord()"
          >
            导出Word
          </a-button>
        </div>
      </div>

      <main ref="iepModalBodyRef" class="iep-modal__body">
        <aside class="iep-plan-navigator">
          <button
            type="button"
            class="iep-plan-nav-item iep-plan-nav-item--root"
            :class="{ 'is-active': executionPlanView === 'iep' }"
            :disabled="navigationDisabled"
            @click="selectExecutionNavigatorItem('iep')"
          >
            <span class="iep-plan-nav-item__status" :class="planRows.length ? 'is-generated' : 'is-empty'" />
            <span class="iep-plan-nav-item__text">
              <strong>IEP总计划</strong>
              <small>{{ planRows.length }}条计划</small>
            </span>
          </button>
          <div v-for="month in executionNavigatorMonths" :key="month.index" class="iep-plan-nav-month">
            <div class="iep-plan-nav-month__head">
              <strong>{{ month.label }}</strong>
              <span>{{ month.rangeText }}</span>
            </div>
            <button
              type="button"
              class="iep-plan-nav-item"
              :class="{
                'is-active': executionPlanView === 'monthly' && selectedExecutionMonth === month.index,
                'is-selected': month.active,
              }"
              :disabled="navigationDisabled"
              @click="selectExecutionNavigatorItem('monthly', month.index)"
            >
              <span class="iep-plan-nav-item__status" :class="month.generated ? 'is-generated' : 'is-empty'" />
              <span class="iep-plan-nav-item__text">
                <strong>{{ month.label }}计划</strong>
                <small>{{ month.generated ? '已生成' : '未生成' }}</small>
              </span>
            </button>
            <button
              v-for="week in month.weeks"
              :key="`${month.index}-${week.index}`"
              type="button"
              class="iep-plan-nav-item iep-plan-nav-item--week"
              :class="{
                'is-active': executionPlanView === 'weekly' && selectedExecutionMonth === month.index && selectedExecutionWeekValue === week.index,
                'is-selected': month.active && selectedExecutionWeekValue === week.index,
              }"
              :disabled="navigationDisabled"
              @click="selectExecutionNavigatorItem('weekly', month.index, week.index)"
            >
              <span class="iep-plan-nav-item__status" :class="week.generated ? 'is-generated' : 'is-empty'" />
              <span class="iep-plan-nav-item__text">
                <strong>{{ week.label }}</strong>
                <small>{{ week.rangeText }} · {{ week.generated ? '已生成' : '未生成' }}</small>
              </span>
            </button>
          </div>
        </aside>
        <div class="a4-workbench">
          <section class="plan-sheet a4-page">
            <h1>{{ activePreviewTitle }}</h1>
            <div v-if="executionPlanView === 'monthly' && !monthlyPlan" class="plan-empty-sheet">
              <strong>{{ selectedExecutionMonthLabel }}计划未生成</strong>
              <span>{{ planTitle }}</span>
            </div>
            <table
              v-else-if="executionPlanView === 'monthly' && monthlyPlan"
              class="plan-sheet-table monthly-plan-table"
              :class="{ 'plan-sheet-table--editing': isMonthlyPlanEditable }"
            >
              <colgroup>
                <col class="monthly-col-1">
                <col class="monthly-col-2">
                <col class="monthly-col-3">
                <col class="monthly-col-4">
                <col class="monthly-col-5">
                <col class="monthly-col-6">
                <col class="monthly-col-7">
                <col class="monthly-col-8">
                <col class="monthly-col-9">
                <col class="monthly-col-10">
                <col class="monthly-col-11">
                <col class="monthly-col-12">
              </colgroup>
              <tbody>
                <tr>
                  <th>姓名</th>
                  <td colspan="2">{{ monthlyPlan.student.name }}</td>
                  <th>性别</th>
                  <td>{{ monthlyPlan.student.gender }}</td>
                  <th colspan="2">出生年月</th>
                  <td colspan="5">{{ monthlyPlan.student.birthDate }}</td>
                </tr>
                <tr>
                  <th>制定<br>日期</th>
                  <td colspan="2">{{ monthlyPlan.meta.planDate }}</td>
                  <th colspan="4">计划参与者</th>
                  <td colspan="5">{{ monthlyPlan.meta.participant }}</td>
                </tr>
                <tr>
                  <th>实施者</th>
                  <td colspan="2">{{ monthlyPlan.meta.implementer }}</td>
                  <th colspan="4">实施起止日期</th>
                  <td colspan="5" class="plan-cell-date">{{ monthlyPlan.meta.startDate }} 至 {{ monthlyPlan.meta.endDate }}</td>
                </tr>
                <tr class="plan-sheet-table__head">
                  <th>康复<br>领域</th>
                  <th colspan="2">长期目标</th>
                  <th colspan="2">短期目标</th>
                  <th colspan="4">训练内容</th>
                  <th>课程<br>形式</th>
                  <th colspan="2" class="plan-cell-date-head">起止日期</th>
                </tr>
                <tr
                  v-for="(row, index) in monthlyDisplayRows"
                  :key="`${row.domain}-${row.rowIndex}-${row.contentIndex}-${index}`"
                  class="plan-data-row"
                  :class="{
                    'plan-data-row--selectable': isMonthlyPlanEditable,
                    'plan-data-row--selected': isMonthlyPlanEditable && selectedMonthlyRowIndex === row.rowIndex && selectedMonthlyContentIndex === row.contentIndex,
                  }"
                  @click="selectMonthlyPlanRow(row.rowIndex, row.contentIndex)"
                >
                  <td v-if="row.showTargetCell" :rowspan="row.contentRowSpan" class="plan-cell-domain">
                    {{ row.domain }}
                  </td>
                  <td v-if="row.showTargetCell" colspan="2" :rowspan="row.contentRowSpan" class="plan-cell-text plan-cell-long">
                    {{ row.longGoal }}
                  </td>
                  <td v-if="row.showTargetCell" colspan="2" :rowspan="row.contentRowSpan" class="plan-cell-text plan-cell-long">
                    {{ row.shortGoal }}
                  </td>
                  <td colspan="4" class="plan-cell-text">
                    {{ row.trainingContentText }}
                  </td>
                  <td v-if="row.showTargetCell" :rowspan="row.contentRowSpan" class="plan-cell-center plan-cell-course">
                    {{ row.courseForm }}
                  </td>
                  <td colspan="2" class="plan-cell-center plan-cell-date monthly-plan-date-cell">
                    {{ row.trainingStartEndDate }}
                  </td>
                </tr>
              </tbody>
            </table>
            <div v-else-if="executionPlanView === 'weekly' && !weeklyPlan" class="plan-empty-sheet">
              <strong>{{ selectedExecutionMonthLabel }}{{ selectedExecutionWeekLabel }}周计划未生成</strong>
              <span>{{ planTitle }}</span>
            </div>
            <table
              v-else-if="executionPlanView === 'weekly' && weeklyPlan"
              class="plan-sheet-table weekly-plan-table"
              :class="{ 'plan-sheet-table--editing': isWeeklyPlanEditable }"
            >
              <colgroup>
                <col class="weekly-col-1">
                <col class="weekly-col-2">
                <col class="weekly-col-3">
                <col class="weekly-col-4">
                <col class="weekly-col-5">
                <col class="weekly-col-6">
                <col class="weekly-col-7">
                <col class="weekly-col-8">
                <col class="weekly-col-9">
                <col class="weekly-col-10">
              </colgroup>
              <tbody>
                <tr>
                  <th>姓名</th>
                  <td>{{ weeklyPlan.student.name }}</td>
                  <th>性别</th>
                  <td>{{ weeklyPlan.student.gender }}</td>
                  <th colspan="2">出生年月</th>
                  <td colspan="4">{{ weeklyPlan.student.birthDate }}</td>
                </tr>
                <tr>
                  <th>任教<br>老师</th>
                  <td>{{ weeklyPlan.teacherName }}</td>
                  <th>课程<br>名称</th>
                  <td>{{ weeklyPlan.courseName }}</td>
                  <th colspan="2">训练日期</th>
                  <td colspan="4" class="plan-cell-date">{{ weeklyPlan.trainingDate }}</td>
                </tr>
                <tr>
                  <th>训练前<br>准备</th>
                  <td colspan="9" class="plan-cell-text">{{ weeklyPlan.preparation }}</td>
                </tr>
                <tr class="weekly-plan-table__main-head">
                  <th rowspan="2">训练项目</th>
                  <th rowspan="2" colspan="3">训练内容</th>
                  <th colspan="6">完成情况</th>
                </tr>
                <tr class="weekly-plan-table__date-head">
                  <th v-for="(date, index) in weeklyDisplayDates" :key="`${date || 'empty'}-${index}`">{{ date }}</th>
                </tr>
                <tr
                  v-for="(row, index) in weeklyPlan.rows"
                  :key="`${row.project}-${index}`"
                  class="plan-data-row"
                  :class="{
                    'plan-data-row--selectable': isWeeklyPlanEditable,
                    'plan-data-row--selected': isWeeklyPlanEditable && selectedWeeklyRowIndex === index,
                  }"
                  @click="selectWeeklyPlanRow(index)"
                >
                  <td class="plan-cell-center">
                    {{ row.project }}
                  </td>
                  <td colspan="3" class="plan-cell-text">
                    {{ row.content }}
                  </td>
                  <td v-for="(_, dayIndex) in weeklyDisplayDates" :key="dayIndex" class="weekly-plan-table__check">
                    {{ row.completion?.[dayIndex] || '' }}
                  </td>
                </tr>
              </tbody>
            </table>
            <table v-else class="plan-sheet-table" :class="{ 'plan-sheet-table--editing': isPlanEditable }">
              <colgroup>
                <col class="plan-col-a">
                <col class="plan-col-b">
                <col class="plan-col-c">
                <col class="plan-col-d">
                <col class="plan-col-e">
                <col class="plan-col-f">
                <col class="plan-col-course">
                <col class="plan-col-date">
              </colgroup>
              <tbody>
                <tr>
                  <th>姓名</th>
                  <td>{{ planSheet.student.name }}</td>
                  <th>性别</th>
                  <td>{{ planSheet.student.gender }}</td>
                  <th>出生年月</th>
                  <td colspan="3">{{ planSheet.student.birthDate }}</td>
                </tr>
                <tr>
                  <th>制定日期</th>
                  <td colspan="3">{{ planSheet.meta.planDate }}</td>
                  <th>计划参与者</th>
                  <td colspan="3">{{ planSheet.meta.participant }}</td>
                </tr>
                <tr class="plan-sheet-table__meta-compact">
                  <th>实施者</th>
                  <td colspan="3">{{ planSheet.meta.implementer }}</td>
                  <th>实施<br>起止日期</th>
                  <td colspan="3" class="plan-cell-date">{{ planSheet.meta.startDate }} 至 {{ planSheet.meta.endDate }}</td>
                </tr>
                <tr class="plan-sheet-table__head">
                  <th>康复<br>领域</th>
                  <th colspan="3">长期目标</th>
                  <th colspan="2">短期目标</th>
                  <th>课程<br>形式</th>
                  <th class="plan-cell-date-head">起止日期</th>
                </tr>
                <tr
                  v-for="(row, index) in planDisplayRows"
                  :key="`${row.domain}-${index}`"
                  class="plan-data-row"
                  :class="{
                    'plan-data-row--selectable': isPlanEditable,
                    'plan-data-row--selected': isPlanEditable && selectedPlanRowIndex === row.sourceIndex,
                  }"
                  @click="selectPlanRow(row.sourceIndex)"
                >
                  <td v-if="row.showGroupCell" :rowspan="row.rowSpan" class="plan-cell-domain">
                    {{ row.domain }}
                  </td>
                  <td v-if="row.showGroupCell" colspan="3" :rowspan="row.rowSpan" class="plan-cell-text plan-cell-long">
                    {{ row.longGoal }}
                  </td>
                  <td colspan="2" class="plan-cell-text">
                    {{ row.shortGoal }}
                  </td>
                  <td class="plan-cell-center plan-cell-course">
                    {{ row.courseForm }}
                  </td>
                  <td class="plan-cell-center plan-cell-date">
                    {{ row.startEndDate }}
                  </td>
                </tr>
                <tr v-if="!planDisplayRows.length" class="plan-empty-row">
                  <td colspan="8">
                    <strong>{{ loadingSavedPlan ? '正在读取IEP计划' : '暂无IEP计划内容' }}</strong>
                    <span>{{ loadingSavedPlan ? '正在读取已保存的草稿或确认计划。' : assessmentAdapter.emptyDescription }}</span>
                  </td>
                </tr>
              </tbody>
            </table>

          </section>
        </div>
        <aside v-if="isAnyPlanEditable" class="iep-edit-panel">
          <div class="iep-edit-panel__head">
            <div>
              <a-tooltip :title="activeEditingLabel">
                <strong class="iep-edit-panel__active-title">{{ activeEditingLabel }}</strong>
              </a-tooltip>
            </div>
          </div>

          <div v-if="isPlanEditable && selectedPlanRow" class="iep-edit-form">
            <section class="iep-edit-section">
              <div class="iep-edit-section__title">
                <span>领域与长期目标</span>
                <div>
                  <a-button size="small" type="text" @click="appendLongGoal(selectedPlanRow.domain)">
                    <template #icon>
                      <PlusOutlined />
                    </template>
                    新增长期目标
                  </a-button>
                  <a-popconfirm title="确认删除最后一条长期目标？" ok-text="删除" cancel-text="取消" @confirm="removeLongGoalLine(selectedPlanRow.domain)">
                    <a-button size="small" type="text" danger>
                      <template #icon>
                        <DeleteOutlined />
                      </template>
                    </a-button>
                  </a-popconfirm>
                </div>
              </div>
              <label class="iep-edit-field">
                <span>康复领域</span>
                <a-input
                  :value="selectedPlanRow.domain"
                  placeholder="请输入康复领域"
                  @update:value="updateSelectedPlanDomain"
                />
              </label>
              <label class="iep-edit-field">
                <span>长期目标</span>
                <a-textarea
                  :value="selectedPlanRow.longGoal"
                  :auto-size="{ minRows: 5, maxRows: 8 }"
                  placeholder="支持按 1. 2. 3. 分条填写"
                  @update:value="updateSelectedPlanLongGoal"
                />
              </label>
            </section>

            <section class="iep-edit-section">
              <div class="iep-edit-section__title">
                <span>短期目标</span>
                <div>
                  <a-button size="small" type="text" @click="addShortGoalAfterSelected">
                    <template #icon>
                      <PlusOutlined />
                    </template>
                    新增短期目标
                  </a-button>
                  <a-popconfirm title="确认删除当前短期目标？" ok-text="删除" cancel-text="取消" @confirm="deleteSelectedShortGoal">
                    <a-button size="small" type="text" danger>
                      <template #icon>
                        <DeleteOutlined />
                      </template>
                    </a-button>
                  </a-popconfirm>
                </div>
              </div>
              <label class="iep-edit-field">
                <span>目标内容</span>
                <a-textarea
                  :value="selectedPlanRow.shortGoal"
                  :auto-size="{ minRows: 4, maxRows: 7 }"
                  placeholder="请输入当前短期目标"
                  @update:value="value => updateSelectedPlanRow({ shortGoal: value })"
                />
              </label>
              <div class="iep-edit-grid">
                <label class="iep-edit-field">
                  <span>课程形式</span>
                  <a-segmented
                    :value="selectedPlanRow.courseForm"
                    :options="courseFormOptions"
                    @update:value="value => updateSelectedPlanRow({ courseForm: value })"
                  />
                </label>
                <label class="iep-edit-field">
                  <span>起止日期</span>
                  <a-input
                    :value="selectedPlanRow.startEndDate"
                    placeholder="YYYY-MM-DD - YYYY-MM-DD"
                    @update:value="value => updateSelectedPlanRow({ startEndDate: value })"
                  />
                </label>
              </div>
            </section>
          </div>

          <div v-else-if="isMonthlyPlanEditable && selectedMonthlyDisplayRow" class="iep-edit-form">
            <section class="iep-edit-section">
              <div class="iep-edit-section__title">
                <span>基本信息</span>
              </div>
              <div class="iep-edit-grid iep-edit-grid--two">
                <label class="iep-edit-field">
                  <span>姓名</span>
                  <a-input
                    :value="monthlyPlan.student.name"
                    @update:value="value => updateMonthlyPlanStudent({ name: value })"
                  />
                </label>
                <label class="iep-edit-field">
                  <span>性别</span>
                  <a-input
                    :value="monthlyPlan.student.gender"
                    @update:value="value => updateMonthlyPlanStudent({ gender: value })"
                  />
                </label>
              </div>
              <label class="iep-edit-field">
                <span>出生年月</span>
                <a-input
                  :value="monthlyPlan.student.birthDate"
                  @update:value="value => updateMonthlyPlanStudent({ birthDate: value })"
                />
              </label>
              <div class="iep-edit-grid iep-edit-grid--two">
                <label class="iep-edit-field">
                  <span>制定日期</span>
                  <a-input
                    :value="monthlyPlan.meta.planDate"
                    @update:value="value => updateMonthlyPlanMeta({ planDate: value })"
                  />
                </label>
                <label class="iep-edit-field">
                  <span>实施者</span>
                  <a-input
                    :value="monthlyPlan.meta.implementer"
                    @update:value="value => updateMonthlyPlanMeta({ implementer: value })"
                  />
                </label>
              </div>
              <label class="iep-edit-field">
                <span>计划参与者</span>
                <a-input
                  :value="monthlyPlan.meta.participant"
                  @update:value="value => updateMonthlyPlanMeta({ participant: value })"
                />
              </label>
              <div class="iep-edit-grid iep-edit-grid--two">
                <label class="iep-edit-field">
                  <span>开始日期</span>
                  <a-input
                    :value="monthlyPlan.meta.startDate"
                    @update:value="value => updateMonthlyPlanMeta({ startDate: value })"
                  />
                </label>
                <label class="iep-edit-field">
                  <span>结束日期</span>
                  <a-input
                    :value="monthlyPlan.meta.endDate"
                    @update:value="value => updateMonthlyPlanMeta({ endDate: value })"
                  />
                </label>
              </div>
            </section>

            <section class="iep-edit-section">
              <div class="iep-edit-section__title">
                <span>领域与目标</span>
              </div>
              <label class="iep-edit-field">
                <span>康复领域</span>
                <a-input
                  :value="selectedMonthlyDisplayRow.domain"
                  @update:value="value => updateMonthlyGroupRows(selectedMonthlyDisplayRow.domain, { domain: value })"
                />
              </label>
              <label class="iep-edit-field">
                <span>长期目标</span>
                <a-textarea
                  :value="selectedMonthlyDisplayRow.longGoal"
                  :auto-size="{ minRows: 5, maxRows: 8 }"
                  @update:value="value => updateMonthlyGroupRows(selectedMonthlyDisplayRow.domain, { longGoal: value })"
                />
              </label>
              <label class="iep-edit-field">
                <span>短期目标</span>
                <a-textarea
                  :value="selectedMonthlyDisplayRow.shortGoal"
                  :auto-size="{ minRows: 4, maxRows: 7 }"
                  @update:value="value => updateMonthlyRow(selectedMonthlyDisplayRow.rowIndex, { shortGoal: value })"
                />
              </label>
            </section>

            <section class="iep-edit-section">
              <div class="iep-edit-section__title">
                <span>训练内容</span>
              </div>
              <label class="iep-edit-field">
                <span>内容</span>
                <a-textarea
                  :value="selectedMonthlyDisplayRow.trainingContent"
                  :auto-size="{ minRows: 4, maxRows: 7 }"
                  @update:value="value => updateMonthlyTrainingItem(selectedMonthlyDisplayRow.rowIndex, selectedMonthlyDisplayRow.contentIndex, { content: value })"
                />
              </label>
              <div class="iep-edit-grid">
                <label class="iep-edit-field">
                  <span>课程形式</span>
                  <a-segmented
                    :value="selectedMonthlyDisplayRow.courseForm"
                    :options="courseFormOptions"
                    @update:value="value => updateMonthlyGroupRows(selectedMonthlyDisplayRow.domain, { courseForm: value })"
                  />
                </label>
                <label class="iep-edit-field">
                  <span>起止日期</span>
                  <a-input
                    :value="selectedMonthlyDisplayRow.trainingStartEndDate"
                    @update:value="value => updateMonthlyTrainingItem(selectedMonthlyDisplayRow.rowIndex, selectedMonthlyDisplayRow.contentIndex, { startEndDate: value })"
                  />
                </label>
              </div>
            </section>
          </div>

          <div v-else-if="isWeeklyPlanEditable && selectedWeeklyPlanRow" class="iep-edit-form">
            <section class="iep-edit-section">
              <div class="iep-edit-section__title">
                <span>基本信息</span>
              </div>
              <div class="iep-edit-grid iep-edit-grid--two">
                <label class="iep-edit-field">
                  <span>姓名</span>
                  <a-input
                    :value="weeklyPlan.student.name"
                    @update:value="value => updateWeeklyPlanStudent({ name: value })"
                  />
                </label>
                <label class="iep-edit-field">
                  <span>性别</span>
                  <a-input
                    :value="weeklyPlan.student.gender"
                    @update:value="value => updateWeeklyPlanStudent({ gender: value })"
                  />
                </label>
              </div>
              <label class="iep-edit-field">
                <span>出生年月</span>
                <a-input
                  :value="weeklyPlan.student.birthDate"
                  @update:value="value => updateWeeklyPlanStudent({ birthDate: value })"
                />
              </label>
              <div class="iep-edit-grid iep-edit-grid--two">
                <label class="iep-edit-field">
                  <span>任教老师</span>
                  <a-input
                    :value="weeklyPlan.teacherName"
                    @update:value="value => updateWeeklyPlanField({ teacherName: value })"
                  />
                </label>
                <label class="iep-edit-field">
                  <span>课程名称</span>
                  <a-input
                    :value="weeklyPlan.courseName"
                    @update:value="value => updateWeeklyPlanField({ courseName: value })"
                  />
                </label>
              </div>
              <label class="iep-edit-field">
                <span>训练日期</span>
                <a-input
                  :value="weeklyPlan.trainingDate"
                  @update:value="value => updateWeeklyPlanField({ trainingDate: value })"
                />
              </label>
              <label class="iep-edit-field">
                <span>训练前准备</span>
                <a-textarea
                  :value="weeklyPlan.preparation"
                  :auto-size="{ minRows: 3, maxRows: 6 }"
                  @update:value="value => updateWeeklyPlanField({ preparation: value })"
                />
              </label>
            </section>

            <section class="iep-edit-section">
              <div class="iep-edit-section__title">
                <span>训练项目</span>
              </div>
              <label class="iep-edit-field">
                <span>项目</span>
                <a-input
                  :value="selectedWeeklyPlanRow.project"
                  @update:value="value => updateWeeklyRow(selectedWeeklyRowIndex, { project: value })"
                />
              </label>
              <label class="iep-edit-field">
                <span>训练内容</span>
                <a-textarea
                  :value="selectedWeeklyPlanRow.content"
                  :auto-size="{ minRows: 4, maxRows: 7 }"
                  @update:value="value => updateWeeklyRow(selectedWeeklyRowIndex, { content: value })"
                />
              </label>
            </section>

          </div>
        </aside>
      </main>

      <div v-if="planLoadingOverlayActive" class="iep-loading-overlay">
        <div class="iep-loading-panel">
          <span class="iep-loading-panel__spinner" />
          <strong>正在切换计划周期</strong>
          <p>{{ headerPlanStatusText }}</p>
        </div>
      </div>

      <div v-if="generationOverlayActive" class="iep-generating-overlay">
        <div class="iep-generating-panel">
          <span class="iep-generating-panel__spinner" />
          <strong>{{ generationOverlayTitle }}</strong>
          <div class="iep-generating-warning">
            {{ generationOverlayWarning }}
          </div>
          <p>{{ generationOverlayDescription }}</p>
          <a-progress
            :percent="generationOverlayPercent"
            :show-info="true"
            status="active"
            size="small"
          />
        </div>
      </div>

      <footer class="iep-modal__footer">
        <div class="footer-hint">
          <template v-if="isIepPreview">
            当前为{{ planTitle }}；{{ isConfirmedPlan ? '已确认计划可继续编辑后保存修改。' : '保存草稿不会改变列表按钮，确认生成后列表显示查看IEP。' }}
          </template>
          <template v-else>
            当前为{{ activePreviewTitle }}预览；{{ isAnyPlanEditable ? '编辑后请保存修改。' : (executionPlanSourceText || `依据：${planTitle}`) }}。
          </template>
        </div>
        <div v-if="isIepPreview" class="footer-actions">
          <a-button @click="closeModal">
            取消
          </a-button>
          <a-button
            v-if="savedPlanStatus !== 'confirmed'"
            :disabled="!planRows.length || loadingSavedPlan || aiGenerating || generatingExecutionPlan || savingExecutionPlan"
            :loading="savingDraft"
            @click="saveIepDraft"
          >
            保存草稿
          </a-button>
          <a-button
            v-if="savedPlanStatus !== 'confirmed'"
            type="primary"
            :disabled="!planRows.length || loadingSavedPlan || aiGenerating || generatingExecutionPlan || savingExecutionPlan"
            :loading="confirmingPlan"
            @click="confirmIepPlan"
          >
            确认生成IEP
          </a-button>
        </div>
        <div v-else class="footer-actions">
          <a-button :disabled="generatingExecutionPlan" @click="switchPreviewView('iep')">
            返回IEP
          </a-button>
          <a-button @click="closeModal">
            取消
          </a-button>
        </div>
      </footer>
    </section>

  </a-modal>
</template>

<style lang="less" scoped>
.iep-modal {
  position: relative;
  overflow: hidden;
  color: #1f2937;
  background: #fff;
  border-radius: 8px;
}

.iep-modal__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 22px 12px;
  border-bottom: 1px solid #edf0f5;

  h2 {
    margin: 0;
    color: #111827;
    font-size: 22px;
    font-weight: 650;
    line-height: 30px;
  }
}

.iep-modal__title-block {
  flex: 1;
  min-width: 0;
}

.iep-header-meta {
  display: flex;
  gap: 14px;
  align-items: center;
  min-width: 0;
  color: #5f6b7a;
  font-size: 13px;
  line-height: 28px;
}

.iep-header-meta__student {
  flex: 0 1 auto;
  min-width: 0;
  overflow: hidden;
  color: #5f6b7a;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.iep-header-meta__plan {
  flex: 0 0 auto;
  color: #5f6b7a;
  white-space: nowrap;
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

.period-switch {
  display: flex;
  flex: 0 0 auto;
  gap: 8px;
  align-items: center;
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

  :deep(.ant-segmented) {
    padding: 2px;
  }

  :deep(.ant-segmented-item) {
    min-height: 28px;
    line-height: 28px;
  }
}

.summary-count {
  flex: 0 0 auto;
  min-width: 0;
  overflow: hidden;
  color: #4b5563;
  font-size: 12px;
  text-align: right;
  text-overflow: ellipsis;
  white-space: nowrap;

  strong {
    color: #1677ff;
    font-weight: 650;
  }
}

.ai-stream-bar {
  display: grid;
  grid-template-columns: auto auto 1fr;
  gap: 10px;
  align-items: center;
  padding: 9px 22px;
  color: #334155;
  font-size: 12px;
  line-height: 20px;
  background: #f8fafc;
  border-bottom: 1px solid #edf0f5;

  strong {
    color: #0f172a;
    font-weight: 650;
    white-space: nowrap;
  }
}

.ai-stream-bar__dot {
  width: 8px;
  height: 8px;
  background: #94a3b8;
  border-radius: 3px;
}

.ai-stream-bar__dot.is-running {
  background: #1677ff;
  animation: ai-stream-pulse 1s ease-in-out infinite;
}

.ai-stream-bar__text {
  min-width: 0;
  overflow: hidden;
  color: #475569;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.iep-preview-toolbar {
  display: grid;
  grid-template-columns: minmax(220px, 1fr) auto;
  gap: 14px;
  align-items: center;
  padding: 10px 22px;
  background: #fff;
  border-bottom: 1px solid #edf0f5;
}

.iep-preview-toolbar__info {
  grid-column: 1;
  min-width: 0;
}

.iep-preview-toolbar__title {
  display: flex;
  gap: 10px;
  align-items: center;
  min-width: 0;
  line-height: 22px;

  > span {
    flex: 0 0 auto;
    color: #8a98ad;
    font-size: 12px;
    font-weight: 650;
    white-space: nowrap;
  }

  strong {
    flex: 1 1 auto;
    min-width: 0;
    overflow: hidden;
    color: #1f2937;
    font-size: 15px;
    font-weight: 650;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.iep-preview-toolbar__meta {
  display: flex;
  gap: 10px;
  align-items: center;
  min-width: 0;
  margin-top: 2px;
  color: #667085;
  font-size: 12px;
  line-height: 18px;

  > span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.iep-preview-toolbar__editing {
  color: #1677ff;
}

.iep-preview-toolbar__actions {
  grid-column: 2;
  display: flex;
  gap: 6px;
  align-items: center;
  justify-self: end;
  min-width: 0;

  :deep(.ant-btn) {
    height: 28px;
    padding: 0 8px;
    font-size: 12px;
    white-space: nowrap;
  }

  :deep(.ant-select-selector) {
    font-size: 12px;
  }
}

.iep-toolbar-tooltip-target {
  display: inline-flex;
}

@keyframes ai-stream-pulse {
  0%,
  100% {
    opacity: 0.45;
  }

  50% {
    opacity: 1;
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
  --iep-side-panel-height: calc(100vh - 286px);

  display: flex;
  gap: 12px;
  align-items: flex-start;
  justify-content: center;
  box-sizing: border-box;
  height: var(--iep-side-panel-height);
  max-height: var(--iep-side-panel-height);
  padding: 12px;
  overflow-x: hidden;
  overflow-y: auto;
  position: relative;
  background: #eef1f5;
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

.iep-plan-navigator {
  position: sticky;
  top: 0;
  display: flex;
  box-sizing: border-box;
  flex: 0 0 206px;
  flex-direction: column;
  gap: 8px;
  height: 100%;
  max-height: 100%;
  padding: 8px;
  overflow-y: auto;
  color: #1f2937;
  background: #fff;
  border: 1px solid #dce3ee;
  border-radius: 8px;
  scrollbar-color: rgba(148, 163, 184, 0.62) transparent;
  scrollbar-width: thin;
}

.iep-plan-navigator::-webkit-scrollbar {
  width: 6px;
}

.iep-plan-navigator::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.5);
  border-radius: 999px;
}

.iep-plan-navigator::-webkit-scrollbar-track {
  background: transparent;
}

.iep-plan-nav-month {
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding-top: 8px;
  border-top: 1px solid #edf1f6;
}

.iep-plan-nav-month__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 22px;
  padding: 0 4px;

  strong {
    color: #111827;
    font-size: 13px;
    font-weight: 650;
    line-height: 20px;
  }

  span {
    color: #8a98ad;
    font-size: 11px;
    line-height: 18px;
    white-space: nowrap;
  }
}

.iep-plan-nav-item {
  display: grid;
  grid-template-columns: 10px minmax(0, 1fr);
  gap: 8px;
  align-items: center;
  width: 100%;
  padding: 8px 10px;
  text-align: left;
  cursor: pointer;
  background: transparent;
  border: 0;
  border-radius: 6px;
  box-shadow: none;
  transition: background-color 0.16s ease, color 0.16s ease;

  &:hover:not(:disabled) {
    background: #f6f8fb;
  }

  &:disabled {
    cursor: not-allowed;
    opacity: 0.62;
  }

  &.is-selected:not(.is-active) {
    background: transparent;
  }

  &.is-active {
    background: #eaf3ff;
    box-shadow: none;
  }
}

.iep-plan-nav-item--root {
  margin-bottom: 2px;
}

.iep-plan-nav-item--week {
  min-height: 34px;
  margin-left: 10px;
  width: calc(100% - 10px);
}

.iep-plan-nav-item__status {
  width: 7px;
  height: 7px;
  border-radius: 50%;
}

.iep-plan-nav-item.is-active .iep-plan-nav-item__status {
  width: 8px;
  height: 8px;
  background: #1677ff;
  box-shadow: 0 0 0 3px rgba(22, 119, 255, 0.12);
}

.iep-plan-nav-item__status.is-generated {
  background: #1677ff;
}

.iep-plan-nav-item__status.is-empty {
  background: #c7d0dd;
}

.iep-plan-nav-item__text {
  display: flex;
  flex-direction: column;
  min-width: 0;

  strong {
    overflow: hidden;
    color: #1f2937;
    font-size: 12px;
    font-weight: 650;
    line-height: 18px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  small {
    overflow: hidden;
    color: #667085;
    font-size: 11px;
    line-height: 16px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.iep-plan-nav-item.is-active .iep-plan-nav-item__text strong,
.iep-plan-nav-item.is-active .iep-plan-nav-item__text small {
  color: #0b5ed7;
}

.iep-plan-nav-item.is-active .iep-plan-nav-item__text strong {
  font-weight: 700;
}

.a4-workbench {
  flex: 0 0 auto;
  display: flex;
  justify-content: flex-start;
  width: 210mm;
  min-width: 210mm;
}

.iep-edit-panel {
  position: sticky;
  top: 0;
  display: flex;
  box-sizing: border-box;
  flex-direction: column;
  flex: 0 0 300px;
  height: 100%;
  max-height: 100%;
  padding: 12px;
  overflow: hidden;
  color: #1f2937;
  background: #fff;
  border: 1px solid #dce3ee;
  border-radius: 8px;
  box-shadow: 0 8px 18px rgba(15, 23, 42, 0.08);
}

.iep-edit-panel__head {
  flex: 0 0 auto;
  display: flex;
  gap: 10px;
  align-items: flex-start;
  justify-content: space-between;
  padding-bottom: 12px;
  border-bottom: 1px solid #edf1f6;

  > div {
    min-width: 0;
    width: 100%;
  }

  strong {
    display: block;
    margin-top: 0;
    color: #111827;
    font-size: 14px;
    font-weight: 650;
    line-height: 21px;
  }
}

.iep-edit-panel__active-title {
  display: block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.iep-edit-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-top: 12px;
  padding-right: 4px;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-color: rgba(148, 163, 184, 0.65) transparent;
  scrollbar-width: thin;
}

.iep-edit-form::-webkit-scrollbar {
  width: 6px;
}

.iep-edit-form::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.52);
  border-radius: 999px;
}

.iep-edit-form::-webkit-scrollbar-track {
  background: transparent;
}

.iep-edit-section {
  padding: 12px;
  background: #fbfcfe;
  border: 1px solid #edf1f7;
  border-radius: 8px;
}

.iep-edit-section__title {
  display: flex;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;

  > span {
    color: #111827;
    font-size: 13px;
    font-weight: 650;
    line-height: 20px;
  }

  > div {
    display: flex;
    gap: 4px;
    align-items: center;
  }

  :deep(.ant-btn) {
    height: 24px;
    padding: 0 6px;
    font-size: 12px;
  }
}

.iep-edit-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 10px;

  &:last-child {
    margin-bottom: 0;
  }

  > span {
    color: #667085;
    font-size: 12px;
    font-weight: 500;
    line-height: 18px;
  }

  :deep(.ant-input),
  :deep(.ant-input-affix-wrapper),
  :deep(.ant-segmented) {
    border-radius: 6px;
  }

  :deep(textarea.ant-input) {
    font-size: 13px;
    line-height: 22px;
  }
}

.iep-edit-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 10px;

  .iep-edit-field {
    margin-bottom: 0;
  }
}

.iep-edit-grid--two {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.iep-generating-overlay {
  position: absolute;
  top: 130px;
  right: 0;
  bottom: 55px;
  left: 0;
  z-index: 20;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  pointer-events: auto;
  background: rgba(238, 241, 245, 0.58);
  backdrop-filter: blur(1.5px);
}

.iep-loading-overlay {
  position: absolute;
  top: 130px;
  right: 0;
  bottom: 55px;
  left: 0;
  z-index: 18;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  pointer-events: auto;
  background: rgba(255, 255, 255, 0.66);
}

.iep-loading-panel {
  display: grid;
  grid-template-columns: auto 1fr;
  column-gap: 10px;
  row-gap: 2px;
  align-items: center;
  width: min(300px, calc(100vw - 96px));
  padding: 13px 15px;
  color: #1f2937;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  box-shadow: 0 14px 30px rgba(15, 23, 42, 0.12);

  strong {
    color: #111827;
    font-size: 13px;
    font-weight: 650;
    line-height: 20px;
  }

  p {
    grid-column: 2;
    min-width: 0;
    margin: 0;
    overflow: hidden;
    color: #6b7280;
    font-size: 12px;
    line-height: 18px;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.iep-loading-panel__spinner {
  display: block;
  grid-row: 1 / span 2;
  width: 16px;
  height: 16px;
  border: 2px solid #dbeafe;
  border-top-color: #1677ff;
  border-radius: 50%;
  animation: iep-loading-spin 0.8s linear infinite;
}

.iep-generating-panel {
  width: min(420px, calc(100vw - 96px));
  padding: 18px 20px 16px;
  color: #1f2937;
  text-align: left;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(210, 218, 230, 0.95);
  border-radius: 8px;
  box-shadow: 0 18px 40px rgba(15, 23, 42, 0.16);

  strong {
    display: block;
    margin: 0 0 6px;
    color: #0f172a;
    font-size: 15px;
    font-weight: 650;
    line-height: 22px;
  }

  .iep-generating-warning {
    margin: 8px 0 10px;
    padding: 8px 10px;
    color: #8a4b00;
    font-size: 12px;
    line-height: 18px;
    background: #fff7e6;
    border: 1px solid #ffd591;
    border-radius: 6px;
  }

  p {
    display: -webkit-box;
    margin: 0 0 12px;
    overflow: hidden;
    color: #5f6b7a;
    font-size: 12px;
    line-height: 20px;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
  }
}

@keyframes iep-loading-spin {
  to {
    transform: rotate(360deg);
  }
}

.iep-generating-panel__spinner {
  float: left;
  width: 18px;
  height: 18px;
  margin: 2px 10px 0 0;
  border: 2px solid #d8e7ff;
  border-top-color: #1677ff;
  border-radius: 50%;
  animation: iep-spin 0.85s linear infinite;
}

@keyframes iep-spin {
  to {
    transform: rotate(360deg);
  }
}

.plan-sheet {
  position: relative;
  color: #111827;
  background: #fff;
  font-family: SimSun, "宋体", serif;

  h1 {
    margin: 0 0 4mm;
    color: #111827;
    font-size: 21px;
    font-weight: 700;
    line-height: 29px;
    text-align: center;
  }
}

.a4-page {
  width: 210mm;
  min-height: 297mm;
  padding: 14mm 13mm 16mm;
  box-sizing: border-box;
  border: 1px solid #d9dee8;
  box-shadow: 0 18px 46px rgba(15, 23, 42, 0.16);
}

.plan-empty-sheet {
  min-height: 160mm;
  padding: 32px 16px;
  color: #667085;
  text-align: center;
  background: #fbfcfe;
  border: 1px dashed #dbe3ee;
  border-radius: 6px;

  strong {
    display: block;
    margin-bottom: 8px;
    color: #111827;
    font-size: 16px;
    font-weight: 650;
    line-height: 24px;
  }

  span {
    display: block;
    font-size: 12px;
    line-height: 20px;
    white-space: pre-wrap;
  }
}

.plan-sheet-table {
  width: 100%;
  table-layout: fixed;
  border-collapse: collapse;
  background: #fff;
  font-family: SimSun, "宋体", serif;

  th,
  td {
    height: 35px;
    padding: 7px 7px;
    color: #000;
    font-size: 12px;
    line-height: 1.55;
    text-align: center;
    white-space: pre-line;
    border: 1px solid #000;
  }

  th {
    color: #000;
    font-weight: 700;
    background: #fff;
  }
}

.plan-col-a {
  width: 10.3%;
}

.plan-col-b {
  width: 14.6%;
}

.plan-col-c {
  width: 6.2%;
}

.plan-col-d {
  width: 8.7%;
}

.plan-col-e {
  width: 12.5%;
}

.plan-col-f {
  width: 15.5%;
}

.plan-col-course {
  width: 9.2%;
}

.plan-col-date {
  width: 23%;
}

.monthly-col-1 {
  width: 8%;
}

.monthly-col-2,
.monthly-col-3 {
  width: 9%;
}

.monthly-col-4,
.monthly-col-5 {
  width: 9.5%;
}

.monthly-col-6,
.monthly-col-7,
.monthly-col-8,
.monthly-col-9 {
  width: 8.75%;
}

.monthly-col-10 {
  width: 8.2%;
}

.monthly-col-11,
.monthly-col-12 {
  width: 5.9%;
}

.weekly-col-1 {
  width: 13%;
}

.weekly-col-2,
.weekly-col-3 {
  width: 13.25%;
}

.weekly-col-4 {
  width: 19.1%;
}

.weekly-col-5,
.weekly-col-6,
.weekly-col-7,
.weekly-col-8,
.weekly-col-9,
.weekly-col-10 {
  width: 6.9%;
}

.plan-sheet-table__head th {
  height: 37px;
  padding-top: 4px;
  padding-bottom: 4px;
  font-size: 12px;
  line-height: 1.35;
  background: #fff;
}

.monthly-plan-table td,
.weekly-plan-table td {
  height: 45px;
}

.monthly-plan-table .plan-cell-date {
  font-size: 12px !important;
}

.monthly-plan-table .monthly-plan-date-cell {
  white-space: normal !important;
  word-break: normal;
}

.weekly-plan-table__main-head th {
  height: 42px;
  font-size: 12px;
  background: #fff;
}

.weekly-plan-table__date-head th {
  height: 34px;
  color: #000;
  font-size: 12px;
  font-weight: 400;
  background: #fff;
  word-break: keep-all;
}

.weekly-plan-table .plan-cell-text {
  vertical-align: middle;
}

.weekly-plan-table__check {
  height: 50px;
  background: #fff;
  white-space: nowrap;
}

.plan-sheet-table--editing .plan-cell-domain,
.plan-sheet-table--editing .plan-cell-text,
.plan-sheet-table--editing .plan-cell-center,
.plan-sheet-table--editing .weekly-plan-table__check {
  vertical-align: middle;
}

.plan-sheet-table__meta-compact th,
.plan-sheet-table__meta-compact td {
  height: 38px;
  padding-top: 5px;
  padding-bottom: 5px;
  line-height: 1.35;
}

.plan-sheet-table--editing td {
  background: #fff;
}

.plan-sheet-table--editing .sheet-input,
.plan-sheet-table--editing .sheet-textarea {
  background: #fff !important;
}

.plan-data-row--selectable {
  cursor: pointer;

  td {
    transition: background-color 0.16s ease, box-shadow 0.16s ease;
  }

  &:hover td {
    background: #f8fbff;
  }
}

.plan-data-row--selected td {
  background: #eef6ff !important;
  box-shadow: inset 0 1px 0 #91caff, inset 0 -1px 0 #91caff;
}

.plan-cell-domain {
  color: #0f172a;
  font-weight: 500;
  vertical-align: middle;
  background: #fbfcfe;
}

.plan-cell-domain,
.plan-cell-text,
.plan-cell-center {
  position: relative;
}

.plan-cell-text {
  text-align: left !important;
  vertical-align: top;
}

.plan-cell-long {
  vertical-align: middle;
}

.plan-cell-center {
  vertical-align: middle;
}

.plan-cell-date {
  padding-right: 4px !important;
  padding-left: 4px !important;
  font-size: 12px !important;
  font-variant-numeric: tabular-nums;
  line-height: 1.35 !important;
  white-space: nowrap !important;
  word-break: keep-all;
}

.plan-cell-date-head {
  white-space: nowrap;
}

.sheet-input,
.sheet-select,
.sheet-textarea {
  width: 100%;
}

.sheet-input,
.sheet-textarea {
  color: #111827 !important;
  font: inherit;
  line-height: inherit;
  background: transparent !important;
  border-color: transparent !important;
  border-radius: 4px;
  box-shadow: none !important;
  transition: border-color 0.16s ease, background-color 0.16s ease;
}

.sheet-select :deep(.ant-select-selector),
.sheet-select :deep(.ant-select-selection-search-input) {
  box-shadow: none !important;
}

.sheet-input:hover,
.sheet-textarea:hover,
.sheet-input:focus,
.sheet-textarea:focus,
.sheet-input:focus-within,
.sheet-textarea:focus-within {
  background: #fbfdff !important;
  border-color: #b7c5d8 !important;
}

.sheet-input--center {
  text-align: center;
}

.sheet-input--date {
  white-space: nowrap;
}

.sheet-input {
  min-height: 26px;
  padding: 2px 5px;
}

.sheet-textarea {
  min-height: 64px;
  padding: 3px 5px;
  resize: none;
}

.plan-cell-editor {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 2px;
  align-items: stretch;
  min-height: 100%;
}

.plan-cell-editor--short {
  gap: 2px;
}

.cell-action-row {
  display: flex;
  gap: 2px;
  justify-content: flex-end;
  min-height: 22px;
  opacity: 0;
  transform: translateY(-2px);
  transition: opacity 0.16s ease, transform 0.16s ease;

  :deep(.ant-btn) {
    height: 22px;
    padding: 0 5px;
    color: #536175;
    font-size: 11px;
    line-height: 20px;
    background: #f8fafc;
    border: 1px solid #e1e7f0;
    border-radius: 4px;
  }

  :deep(.ant-btn-dangerous) {
    color: #d9363e;
    background: #fffafa;
    border-color: #f4d1d5;
  }
}

.plan-cell-editor:hover .cell-action-row,
.plan-cell-editor:focus-within .cell-action-row,
tr:hover .cell-action-row {
  opacity: 1;
  transform: translateY(0);
}

.plan-cell-course {
  padding-right: 3px !important;
  padding-left: 3px !important;
  white-space: nowrap !important;
}

.sheet-select--course {
  min-width: 42px;

  :deep(.ant-select-selector) {
    height: 26px !important;
    min-height: 26px !important;
    padding: 0 3px !important;
    color: #111827;
    text-align: center;
    background: transparent !important;
    border-color: transparent !important;
    border-radius: 4px !important;
    box-shadow: none !important;
  }

  :deep(.ant-select-selection-item) {
    padding-inline-end: 0 !important;
    overflow: visible;
    color: #111827;
    font-size: 12px;
    line-height: 24px !important;
    text-align: center;
  }

  :deep(.ant-select-arrow) {
    display: none;
  }

  &:hover :deep(.ant-select-selector),
  &.ant-select-focused :deep(.ant-select-selector) {
    background: #fbfdff !important;
    border-color: #b7c5d8 !important;
  }
}

.date-range-editor {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  gap: 3px;
  align-items: center;
}

.date-range-editor > span {
  color: #64748b;
  font-size: 11px;
  white-space: nowrap;
}

.plan-empty-row td {
  height: 260px;
  color: #64748b;
  background: #fbfcfe;
  text-align: center !important;

  strong {
    display: block;
    margin-bottom: 6px;
    color: #0f172a;
    font-size: 14px;
    font-weight: 650;
  }

  span {
    font-size: 12px;
  }
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
  gap: 16px;
  align-items: center;
  justify-content: space-between;
  padding: 12px 18px;
  background: #fff;
  border-top: 1px solid #edf0f5;
}

.footer-hint {
  position: relative;
  flex: 1;
  min-width: 0;
  padding-left: 18px;
  overflow: hidden;
  color: #5f6b7a;
  font-size: 12px;
  line-height: 20px;
  text-overflow: ellipsis;
  white-space: nowrap;

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
  flex: 0 0 auto;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
  justify-content: flex-end;
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
