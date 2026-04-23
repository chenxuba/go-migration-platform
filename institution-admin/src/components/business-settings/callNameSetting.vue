<script setup lang="ts">
import { InputNumber, Modal } from 'ant-design-vue'
import { computed, h, onMounted, ref } from 'vue'
import { type InstConfig, setInstConfigApi } from '~@/api/common/config'
import { getCourseRollCallDeductRulePageApi } from '~@/api/edu-center/course-list'
import RollCallDeductRuleModal from '~@/components/business-settings/roll-call-deduct-rule-modal.vue'
import RollCallDefaultClassTimeModal from '~@/components/business-settings/roll-call-default-class-time-modal.vue'
import { useUserStore } from '~@/stores/user'
import messageService from '~@/utils/messageService'

const userStore = useUserStore()
const activeKey = ref('deduct-order')
const rowLoadingMap = ref<Record<string, boolean>>({})

const courseDeductOrder = ref('oldest')

const instConfig = computed<Partial<InstConfig>>(() => userStore.instConfig ?? {})

function isConfigEnabled(value: unknown, defaultValue = false) {
  if (typeof value === 'boolean')
    return value
  if (typeof value === 'number')
    return value !== 0
  if (typeof value === 'string') {
    const normalized = value.trim().toLowerCase()
    if (normalized === '1' || normalized === 'true')
      return true
    if (normalized === '0' || normalized === 'false')
      return false
  }
  return defaultValue
}

function normalizePositiveInteger(value: unknown, fallback = 1) {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0)
    return fallback
  return Math.floor(parsed)
}

function normalizeNumberText(value: unknown, fallback: string) {
  if (value == null)
    return fallback
  const text = String(value).trim()
  return text || fallback
}

const consumeOverEnabled = computed(() => isConfigEnabled(instConfig.value.enabledArrearsRollcall, false))
const autoRollCallEnabled = computed(() => isConfigEnabled(instConfig.value.enableByAutoTeaching, false))
const limitSingleOrderEnabled = computed(() => isConfigEnabled(instConfig.value.enableLimitSingleOrderArrearsDeduct, false))
const leaveNormalByHourEnabled = computed(() => isConfigEnabled(instConfig.value.enableHourLeaveNormalRecord, false))
const absentNormalByHourEnabled = computed(() => isConfigEnabled(instConfig.value.enableHourTruancyNormalRecord, false))
const periodMakeupEnabled = computed(() => isConfigEnabled(instConfig.value.enablePeriodMakeup, false))
const periodAutoEndEnabled = computed(() => isConfigEnabled(instConfig.value.enablePeriodAutoFinishWhenZero, false))
const amountLeaveNormalEnabled = computed(() => isConfigEnabled(instConfig.value.enablePriceLeaveNormalRecord, false))
const amountAbsentNormalEnabled = computed(() => isConfigEnabled(instConfig.value.enablePriceTruancyNormalRecord, false))
const amountMakeupEnabled = computed(() => isConfigEnabled(instConfig.value.enablePriceMakeup, false))
const faceDeductEnabled = computed(() => isConfigEnabled(instConfig.value.enableFaceAttendanceRelateTeaching, false))
const faceSignInNoticeEnabled = computed(() => isConfigEnabled(instConfig.value.enableFaceAttendanceCheckInNotice, false))
const faceSignOutNoticeEnabled = computed(() => isConfigEnabled(instConfig.value.enableFaceAttendanceCheckOutNotice, false))
const voicePromptEnabled = computed(() => isConfigEnabled(instConfig.value.enableByVoiceTips, false))
const faceAdminNoticeEnabled = computed(() => isConfigEnabled(instConfig.value.enableSendFaceAttendNoticeToAdmin, false))
const faceAttendanceIntervalMinutes = computed(() => normalizePositiveInteger(instConfig.value.faceAttendanceInterval, 1))
const defaultClassTimeRecordMode = computed(() => normalizePositiveInteger(instConfig.value.defaultClassTimeRecordMode, 1))
const defaultStudentClassTimeText = computed(() => normalizeNumberText(instConfig.value.defaultStudentClassTime, '1'))
const defaultTeacherClassTimeText = computed(() => normalizeNumberText(instConfig.value.defaultTeacherClassTime, '0'))
const chargeByPriceDefaultPriceText = computed(() => normalizeNumberText(instConfig.value.chargeByPriceDefaultPrice, '100'))
const defaultClassTimeModalOpen = ref(false)
const deductRuleModalOpen = ref(false)
const amountDeductSingleCourseCount = ref(0)

const orderExampleRows = [
  { no: 1, order: '订单 A', validUntil: '2025-06-08', accountType: '正价', createdAt: '2025-05-05 10:10' },
  { no: 2, order: '订单 B', validUntil: '2025-08-08', accountType: '正价', createdAt: '2025-06-05 10:10' },
  { no: 3, order: '订单 B', validUntil: '2025-08-08', accountType: '赠送', createdAt: '2025-06-05 10:10' },
  { no: 4, order: '订单 C', validUntil: '不限制', accountType: '正价', createdAt: '2025-07-05 10:10' },
]

function isRowLoading(key: string) {
  return Boolean(rowLoadingMap.value[key])
}

async function ensureInstConfigLoaded() {
  if (!userStore.instConfig)
    await userStore.getInstConfig()
}

function extractPagedItems(res: any) {
  const list = res?.result?.items ?? res?.result ?? res?.data?.items ?? res?.data ?? []
  return Array.isArray(list) ? list : []
}

async function refreshAmountDeductRuleSummary() {
  try {
    const res = await getCourseRollCallDeductRulePageApi({
      pageRequestModel: {
        needTotal: true,
        pageIndex: 1,
        pageSize: 200,
      },
      sortModel: {
        byUpdateTime: -1,
        byTotalSales: 0,
      },
      queryModel: {
        chargeTypes: [3],
        delFlag: false,
      },
    })
    const rows = extractPagedItems(res)
    amountDeductSingleCourseCount.value = rows.filter((item: any) => item?.rollCallDeductPrice != null && item?.rollCallDeductPrice !== '').length
  }
  catch (error) {
    console.error('refresh amount deduct rule summary failed', error)
    amountDeductSingleCourseCount.value = 0
  }
}

async function updateConfigField(field: keyof InstConfig, value: InstConfig[keyof InstConfig], key: string, successText: string) {
  rowLoadingMap.value = {
    ...rowLoadingMap.value,
    [key]: true,
  }
  try {
    await setInstConfigApi({
      ...(instConfig.value as InstConfig),
      [field]: value,
    })
    await userStore.getInstConfig()
    messageService.success(successText)
  }
  catch (error) {
    console.error(`update ${String(field)} failed`, error)
    messageService.error('保存失败，请稍后重试')
  }
  finally {
    rowLoadingMap.value = {
      ...rowLoadingMap.value,
      [key]: false,
    }
  }
}

async function handleConsumeOverToggle(checked: boolean) {
  await updateConfigField('enabledArrearsRollcall', checked, 'consumeOver', checked ? '已开启消超点名' : '已关闭消超点名')
}

async function handleAutoRollCallToggle(checked: boolean) {
  await updateConfigField('enableByAutoTeaching', checked, 'autoRollCall', checked ? '已开启自动点名' : '已关闭自动点名')
}

async function handleLimitSingleOrderToggle(checked: boolean) {
  await updateConfigField('enableLimitSingleOrderArrearsDeduct', checked, 'limitSingleOrder', checked ? '已开启限制订单欠费课消' : '已关闭限制订单欠费课消')
}

async function handleLeaveNormalByHourToggle(checked: boolean) {
  await updateConfigField('enableHourLeaveNormalRecord', checked, 'leaveNormalByHour', checked ? '已开启学员请假正常记录课时' : '已关闭学员请假正常记录课时')
}

async function handleAbsentNormalByHourToggle(checked: boolean) {
  await updateConfigField('enableHourTruancyNormalRecord', checked, 'absentNormalByHour', checked ? '已开启学员旷课正常记录课时' : '已关闭学员旷课正常记录课时')
}

async function handlePeriodMakeupToggle(checked: boolean) {
  await updateConfigField('enablePeriodMakeup', checked, 'periodMakeup', checked ? '已开启按时段收费学员补课' : '已关闭按时段收费学员补课')
}

async function handlePeriodAutoEndToggle(checked: boolean) {
  await updateConfigField('enablePeriodAutoFinishWhenZero', checked, 'periodAutoEnd', checked ? '已开启课消为0天后自动结课' : '已关闭课消为0天后自动结课')
}

async function handleAmountLeaveNormalToggle(checked: boolean) {
  await updateConfigField('enablePriceLeaveNormalRecord', checked, 'amountLeaveNormal', checked ? '已开启学员请假正常记录金额' : '已关闭学员请假正常记录金额')
}

async function handleAmountAbsentNormalToggle(checked: boolean) {
  await updateConfigField('enablePriceTruancyNormalRecord', checked, 'amountAbsentNormal', checked ? '已开启学员旷课正常记录金额' : '已关闭学员旷课正常记录金额')
}

async function handleAmountMakeupToggle(checked: boolean) {
  await updateConfigField('enablePriceMakeup', checked, 'amountMakeup', checked ? '已开启按金额收费学员补课' : '已关闭按金额收费学员补课')
}

async function handleFaceDeductToggle(checked: boolean) {
  await updateConfigField('enableFaceAttendanceRelateTeaching', checked, 'faceDeduct', checked ? '已开启人脸考勤关联点名课消' : '已关闭人脸考勤关联点名课消')
}

async function handleFaceSignInNoticeToggle(checked: boolean) {
  await updateConfigField('enableFaceAttendanceCheckInNotice', checked, 'faceSignInNotice', checked ? '已开启学员考勤签到通知' : '已关闭学员考勤签到通知')
}

async function handleFaceSignOutNoticeToggle(checked: boolean) {
  await updateConfigField('enableFaceAttendanceCheckOutNotice', checked, 'faceSignOutNotice', checked ? '已开启学员考勤签退通知' : '已关闭学员考勤签退通知')
}

async function handleVoicePromptToggle(checked: boolean) {
  await updateConfigField('enableByVoiceTips', checked, 'voicePrompt', checked ? '已开启语音提示' : '已关闭语音提示')
}

async function handleFaceAdminNoticeToggle(checked: boolean) {
  await updateConfigField('enableSendFaceAttendNoticeToAdmin', checked, 'faceAdminNotice', checked ? '已开启学员刷脸成功提醒至管理员' : '已关闭学员刷脸成功提醒至管理员')
}

function handleEditFaceAttendanceInterval() {
  let nextValue = faceAttendanceIntervalMinutes.value
  Modal.confirm({
    title: '编辑刷脸间隔',
    centered: true,
    content: h('div', { style: 'display:flex;align-items:center;gap:8px;margin-top:8px;' }, [
      h('span', null, '刷脸间隔'),
      h(InputNumber, {
        min: 1,
        max: 1440,
        value: nextValue,
        precision: 0,
        style: 'width: 120px;',
        'onUpdate:value': (value: number | null) => {
          nextValue = normalizePositiveInteger(value, faceAttendanceIntervalMinutes.value)
        },
      }),
      h('span', null, '分钟'),
    ]),
    async onOk() {
      await updateConfigField('faceAttendanceInterval', String(normalizePositiveInteger(nextValue, 1)), 'faceAttendanceInterval', '已保存刷脸间隔设置')
    },
  })
}

async function handleDeductRuleSaved() {
  await refreshAmountDeductRuleSummary()
}

onMounted(async () => {
  await ensureInstConfigLoaded()
  await refreshAmountDeductRuleSummary()
})
</script>

<template>
  <div class="roll-call-settings">
    <a-tabs v-model:active-key="activeKey" destroy-inactive-tab-pane class="roll-call-settings__tabs">
      <a-tab-pane key="deduct-order" tab="点名课消顺序">
        <section class="settings-section">
          <div class="settings-section__title">
            点名课消顺序
          </div>

          <div class="settings-table">
            <div class="settings-row">
              <div class="settings-row__label">
                课程课消顺序
              </div>
              <div class="settings-row__content">
                <a-radio-group v-model:value="courseDeductOrder" class="settings-radio-group custom-radio">
                  <a-radio value="oldest">
                    先进先出
                  </a-radio>
                  <a-radio value="account-first">
                    优先消通用账户
                  </a-radio>
                  <a-radio value="normal-first">
                    优先消普通账户
                  </a-radio>
                </a-radio-group>
                <div class="settings-desc">
                  按学员的学费账户生成时间优先消耗，适用于多数常规课消场景。
                </div>
              </div>
            </div>

            <div class="settings-row settings-row--top">
              <div class="settings-row__label">
                订单课消顺序
              </div>
              <div class="settings-row__content">
                <div class="settings-inline">
                  <span>智能排序：</span>
                  <span class="text-primary">先进先出</span>
                  <a-button type="link"  class="settings-link">
                    编辑
                  </a-button>
                </div>

                <div class="example-card">
                  <div class="example-card__title">
                    场景举例：
                  </div>
                  <div class="example-card__desc">
                    学员报读同课程有多个订单时，系统将按规则自动选择优先课消的订单。
                  </div>
                  <table class="example-table">
                    <thead>
                      <tr>
                        <th>序号</th>
                        <th>订单</th>
                        <th>有效期至</th>
                        <th>账户属性</th>
                        <th>生成时间</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="row in orderExampleRows" :key="row.no">
                        <td>{{ row.no }}</td>
                        <td>{{ row.order }}</td>
                        <td>{{ row.validUntil }}</td>
                        <td>{{ row.accountType }}</td>
                        <td>{{ row.createdAt }}</td>
                      </tr>
                    </tbody>
                  </table>
                  <div class="example-card__footer">
                    若设置智能排序规则：近有效期 &gt; 赠送 &gt; 先进先出，则订单课消顺序为：1 &gt; 3 &gt; 2 &gt; 4
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </a-tab-pane>

      <a-tab-pane key="deduct-rule" tab="点名课消规则">
        <section class="settings-section">
          <div class="settings-section__title">
            点名课消规则
          </div>

          <div class="settings-table">
            <div class="settings-row">
              <div class="settings-row__label">
                消超点名
              </div>
              <div class="settings-row__content">
                <a-switch :checked="consumeOverEnabled" :loading="isRowLoading('consumeOver')" @change="handleConsumeOverToggle" />
                <div class="settings-desc">
                  开启后，学员的上课点名支持消超记录（支持欠费点名）。
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                自动点名
              </div>
              <div class="settings-row__content">
                <a-switch :checked="autoRollCallEnabled" :loading="isRowLoading('autoRollCall')" @change="handleAutoRollCallToggle" />
                <div class="settings-desc">
                  开启后，未点名且未触发人脸考勤关联课消的课程，在结束后 35 分钟内，系统将自动点名并记录课消，作为兜底处理。
                </div>
              </div>
            </div>

            <div class="settings-row settings-row--top">
              <div class="settings-row__label">
                限制订单欠费课消
              </div>
              <div class="settings-row__content">
                <a-switch :checked="limitSingleOrderEnabled" :loading="isRowLoading('limitSingleOrder')" @change="handleLimitSingleOrderToggle" />
                <div class="settings-desc">
                  开启后，学员订单对应的费用耗完后，限制继续课消。
                </div>
                <div class="example-card">
                  <div class="example-card__title">
                    场景举例：
                  </div>
                  <div class="example-card__desc">
                    开启限制订单欠费课消后，以下情况学员无法正常课消：
                  </div>
                  <ol class="example-list">
                    <li>学员报名 10 课时共计 100 元，实际支付 50 元，欠费 50 元，消耗 5 课时后，无法继续课消。</li>
                    <li>学员报名课程支付 0 元，如果是欠费报名，则无法课消。</li>
                    <li>学员报名的课程 A 有三笔订单，订单 1 的实收学费不足时，可手动调整课消顺序继续课消。</li>
                  </ol>
                </div>
              </div>
            </div>

            <div class="settings-row settings-row--top">
              <div class="settings-row__label">
                按课时收费
              </div>
              <div class="settings-row__content">
                <div class="settings-inline settings-inline--heading">
                  按“课时购买数”定价，以“课时”为单位计费
                </div>
                <div class="rule-box">
                  <div class="settings-inline">
                    <span><span class="settings-switch-line__label">默认记录课时：</span>创建班级 / 1 对 1 时仍可编辑调整</span>
                  </div>
                  <div class="settings-inline settings-inline--muted">
                    <span>
                      {{ Number(defaultClassTimeRecordMode) === 2 ? '按上课时长记录' : '按固定课时记录' }}：
                      默认记录学员 <span class="text-primary">{{ defaultStudentClassTimeText }}</span> 课时，教师 <span class="text-primary">{{ defaultTeacherClassTimeText }}</span> 课时
                    </span>
                    <a-button type="link" class="settings-link" @click="defaultClassTimeModalOpen = true">
                      编辑
                    </a-button>
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员请假正常记录：</span>
                    <a-switch :checked="leaveNormalByHourEnabled" :loading="isRowLoading('leaveNormalByHour')" @change="handleLeaveNormalByHourToggle" />
                  </div>
                  <div class="settings-desc">
                    开启后，学员请假正常记录课时
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员旷课正常记录：</span>
                    <a-switch :checked="absentNormalByHourEnabled" :loading="isRowLoading('absentNormalByHour')" @change="handleAbsentNormalByHourToggle" />
                  </div>
                  <div class="settings-desc">
                    开启后，学员旷课正常记录课时
                  </div>
                </div>
              </div>
            </div>

            <div class="settings-row settings-row--top">
              <div class="settings-row__label">
                按时段收费
              </div>
              <div class="settings-row__content">
                <div class="settings-inline">
                  按“天 / 自然月 / 自然年”定价，以“天”为单位计费，每日自动课消。
                </div>
                <div class="rule-box">
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员补课：</span>
                    <a-switch :checked="periodMakeupEnabled" :loading="isRowLoading('periodMakeup')" @change="handlePeriodMakeupToggle" />
                  </div>
                  <div class="settings-desc">
                    开启后，按时段收费的课程，支持缺课学员补课
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">课消为 0 天后自动结课：</span>
                    <a-switch :checked="periodAutoEndEnabled" :loading="isRowLoading('periodAutoEnd')" @change="handlePeriodAutoEndToggle" />
                  </div>
                  <div class="settings-desc">
                    开启后，当课程账户每日自动课消剩余为 0 天后，次日自动结课
                  </div>
                </div>
              </div>
            </div>

            <div class="settings-row settings-row--top">
              <div class="settings-row__label">
                按金额收费
              </div>
              <div class="settings-row__content">
                <div class="settings-inline settings-inline--heading">
                  按“充值金额”定价，每次点名扣除对应金额数
                </div>
                <div class="rule-box">
                  <div class="settings-inline">
                    <span class="settings-switch-line__label">扣费规则：</span>
                    <a-button type="link" class="settings-link" @click="deductRuleModalOpen = true">
                      编辑
                    </a-button>
                  </div>
                  <div class="settings-desc">
                    默认扣费：<span class="text-primary">{{ chargeByPriceDefaultPriceText }}</span> 元（仅对未设置单课扣费的课程有效）
                  </div>
                  <div class="settings-desc">
                    单课扣费：已设置 <span class="text-primary">{{ amountDeductSingleCourseCount }}</span> 门课程，点此上方编辑操作对单个课程进行扣费金额设置
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员请假正常记录：</span>
                    <a-switch :checked="amountLeaveNormalEnabled" :loading="isRowLoading('amountLeaveNormal')" @change="handleAmountLeaveNormalToggle" />
                  </div>
                  <div class="settings-desc">
                    开启后，学员请假正常记录金额
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员旷课正常记录：</span>
                    <a-switch :checked="amountAbsentNormalEnabled" :loading="isRowLoading('amountAbsentNormal')" @change="handleAmountAbsentNormalToggle" />
                  </div>
                  <div class="settings-desc">
                    开启后，学员旷课正常记录金额
                  </div>
                  <div class="settings-switch-line">
                    <span class="settings-switch-line__label">学员补课：</span>
                    <a-switch :checked="amountMakeupEnabled" :loading="isRowLoading('amountMakeup')" @change="handleAmountMakeupToggle" />
                  </div>
                  <div class="settings-desc">
                    开启后，按金额收费的课程，支持缺课学员补课
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </a-tab-pane>

      <a-tab-pane key="face-attendance" tab="人脸考勤设置">
        <section class="settings-section">
          <div class="settings-section__title">
            人脸考勤设置
          </div>

          <div class="settings-table">
            <div class="settings-row">
              <div class="settings-row__label">
                人脸考勤关联点名课消
              </div>
              <div class="settings-row__content">
                <a-switch :checked="faceDeductEnabled" :loading="isRowLoading('faceDeduct')" @change="handleFaceDeductToggle" />
                <div class="settings-inline settings-inline--block">
                  <span class="settings-switch-line__label">人脸考勤课消规则：</span>
                  <span>按学员人脸签到后的当日日程自动课消</span>
                </div>
                <div class="settings-desc">
                  开启后，系统将匹配学员当日已关联的班课 / 1 对 1 日程；对结束时间晚于签到时间的课程，在下课 30 分钟后自动点名课消，已有点名记录的不重复课消。
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                学员考勤签到通知
              </div>
              <div class="settings-row__content">
                <a-switch :checked="faceSignInNoticeEnabled" :loading="isRowLoading('faceSignInNotice')" @change="handleFaceSignInNoticeToggle" />
                <div class="settings-desc">
                  开启后，学员完成考勤签到或手动标记签到将自动发送微信通知给家长。
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                学员考勤签退通知
              </div>
              <div class="settings-row__content">
                <a-switch :checked="faceSignOutNoticeEnabled" :loading="isRowLoading('faceSignOutNotice')" @change="handleFaceSignOutNoticeToggle" />
                <div class="settings-desc">
                  开启后，学员完成考勤签退或手动标记签退将自动发送微信通知给家长。
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                语音提示
              </div>
              <div class="settings-row__content">
                <a-switch :checked="voicePromptEnabled" :loading="isRowLoading('voicePrompt')" @change="handleVoicePromptToggle" />
                <div class="settings-desc">
                  开启后，系统实时播报学员是否人脸考勤成功、签到签退成功。
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                刷脸间隔设置
              </div>
              <div class="settings-row__content">
                <div class="settings-inline">
                  <span>刷脸间隔 <span class="text-primary">{{ faceAttendanceIntervalMinutes }}</span> 分钟</span>
                  <a-button type="link" class="settings-link" :loading="isRowLoading('faceAttendanceInterval')" @click="handleEditFaceAttendanceInterval">
                    编辑
                  </a-button>
                </div>
                <div class="settings-desc">
                  学员刷脸考勤成功后，允许下次刷脸考勤间隔时长。
                </div>
              </div>
            </div>

            <div class="settings-row">
              <div class="settings-row__label">
                学员刷脸成功提醒至管理员
              </div>
              <div class="settings-row__content">
                <a-switch :checked="faceAdminNoticeEnabled" :loading="isRowLoading('faceAdminNotice')" @change="handleFaceAdminNoticeToggle" />
                <div class="settings-desc">
                  开启后，校区管理员将在 App 消息中心收到学员刷脸成功的推送提醒。
                </div>
              </div>
            </div>
          </div>
        </section>
      </a-tab-pane>
    </a-tabs>

    <RollCallDefaultClassTimeModal
      v-model:open="defaultClassTimeModalOpen"
      :inst-config="instConfig"
    />
    <RollCallDeductRuleModal
      v-model:open="deductRuleModalOpen"
      :inst-config="instConfig"
      @saved="handleDeductRuleSaved"
    />
  </div>
</template>

<style lang="less" scoped>
.roll-call-settings {
  min-height: 420px;
}

.roll-call-settings__tabs {
  :deep(.ant-tabs-nav) {
    margin: 0;
    background: #fff;
    border-radius: 0 0 16px 16px !important;

    &::before {
      display: none;
    }
  }

  :deep(.ant-tabs-nav-wrap) {
    padding-left: 10px;
    margin: 6px 0;
  }

  :deep(.ant-tabs-tab) {
    margin: 0 8px 0 0;
    padding: 6px 14px !important;
    font-size: 14px !important;
  }

  :deep(.ant-tabs-tab .ant-tabs-tab-btn) {
    font-size: 14px !important;
    line-height: 22px;
  }

  :deep(.ant-tabs-tab-active) {
    background: #e6f0ff;
    border-radius: 8px;
  }

  :deep(.ant-tabs-tab-active .ant-tabs-tab-btn) {
    color: var(--pro-ant-color-primary, #1677ff);
    font-size: 14px !important;
    font-weight: 500;
  }

  :deep(.ant-tabs-ink-bar) {
    display: none;
  }

  :deep(.ant-tabs-content-holder) {
    background: transparent;
  }
}

.settings-section {
  margin-top: 10px;
  padding: 16px 16px 22px;
  background: #fff;
  border-radius: 14px;
}

.settings-section__title {
  position: relative;
  padding-left: 12px;
  margin-bottom: 14px;
  color: #1f2937;
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;

  &::before {
    position: absolute;
    top: 5px;
    left: 0;
    width: 3px;
    height: 14px;
    border-radius: 3px;
    background: var(--pro-ant-color-primary, #1677ff);
    content: '';
  }
}

.settings-table {
  overflow: hidden;
  border: 1px solid #edf0f5;
  background: #fff;
}

.settings-row {
  display: grid;
  grid-template-columns: 200px minmax(0, 1fr);
  min-height: 84px;
  border-bottom: 1px solid #edf0f5;

  &:last-child {
    border-bottom: 0;
  }
}

.settings-row--top {
  align-items: stretch;
}

.settings-row__label {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 18px 14px;
  border-right: 1px solid #edf0f5;
  color: #1f2937;
  font-size: 14px;
  font-weight: 500;
  text-align: center;
}

.settings-row__content {
  min-width: 0;
  padding: 18px 18px;
  color: #1f2937;
  font-size: 14px;
}

.settings-radio-group {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 18px;
}

.custom-radio :deep(.ant-radio-wrapper:hover .ant-radio),
.custom-radio :deep(.ant-radio:hover .ant-radio-inner),
.custom-radio :deep(.ant-radio-input:focus + .ant-radio-inner) {
  border-color: var(--pro-ant-color-primary, #1677ff);
}

.custom-radio :deep(.ant-radio-inner) {
  background-color: transparent;
  border-color: #d9d9d9;
}

.custom-radio :deep(.ant-radio-checked .ant-radio-inner) {
  background-color: transparent;
  border-color: var(--pro-ant-color-primary, #1677ff);
}

.custom-radio :deep(.ant-radio-inner::after) {
  background-color: var(--pro-ant-color-primary, #1677ff);
  transform: scale(0.5);
}

.settings-desc {
  margin-top: 8px;
  color: #333;
  font-size: 14px;
  line-height: 20px;
}

.settings-inline {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  color: #1f2937;
}

.settings-inline--muted {
  margin-top: 8px;
  color: #333;
}

.settings-inline--heading {
  margin-bottom: 12px;
  color: #1f2937;
  font-size: 14px;
  line-height: 22px;
}

.settings-inline--block {
  margin-top: 12px;
}

.settings-switch-line {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  color: #333;
  font-size: 14px;
}

.settings-switch-line__label {
  color: #1f2937;
  font-weight: 600;
}

.settings-link {
  padding: 0 4px;
  font-size: 14px;
}

.text-primary {
  color: var(--pro-ant-color-primary, #1677ff);
}

.example-card {
  margin-top: 14px;
  padding: 14px;
  border: 1px solid #edf0f5;
  background: #fafafa;
  border-radius: 8px;
}

.rule-box {
  margin-top: 14px;
  padding: 12px 14px;
  border: 1px solid #edf0f5;
  background: #fafafa;
  border-radius: 8px;
}

.example-card__title {
  color: #1f2937;
  font-size: 14px;
  font-weight: 600;
}

.example-card__desc,
.example-card__footer {
  margin-top: 8px;
  color: #333;
  font-size: 14px;
  line-height: 20px;
}

.example-table {
  width: 100%;
  margin-top: 12px;
  overflow: hidden;
  border-spacing: 0;
  border-collapse: separate;
  border: 1px solid #edf0f5;

  th,
  td {
    height: 44px;
    padding: 0 12px;
    border-right: 1px solid #edf0f5;
    border-bottom: 1px solid #edf0f5;
    color: #4b5563;
    font-size: 14px;
    text-align: center;

    &:last-child {
      border-right: 0;
    }
  }

  th {
    color: #1f2937;
    font-weight: 600;
  }

  tbody tr:last-child td {
    border-bottom: 0;
  }
}

.example-list {
  margin: 8px 0 0;
  padding-left: 18px;
  color: #333;
  font-size: 14px;
  line-height: 24px;
}

@media (max-width: 768px) {
  .settings-row {
    grid-template-columns: 1fr;
  }

  .settings-row__label {
    justify-content: flex-start;
    border-right: 0;
    border-bottom: 1px solid #edf0f5;
  }
}
</style>
