<script setup lang="ts">
import { EditOutlined, InfoCircleFilled } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { computed, reactive, ref } from 'vue'
import { type InstConfig, getInstConfigModuleApi, setInstConfigModuleApi } from '~@/api/common/config'
import messageService from '~@/utils/messageService'

const activeKey = ref('message')
const activeMessageKey = ref('sms')
const rowLoadingMap = ref<Record<string, boolean>>({})

const birthdayModalOpen = ref(false)
const birthdayEnableAfterConfirm = ref(false)
const leaveCountModalOpen = ref(false)
const leaveCountEnableAfterConfirm = ref(false)
const leaveTimeModalOpen = ref(false)
const leaveTimeEnableAfterConfirm = ref(false)
const smsClassReminderModalOpen = ref(false)
const classReminderModalOpen = ref(false)
const renewalModalOpen = ref(false)

const birthdayReceivers = ref(['current', 'history'])
const birthdayDraft = ref<string[]>([])
const smsClassReminderHourDraft = ref('19:00')
const classReminderHourDraft = ref('19:00')
const classReminderHourOptions = Array.from({ length: 24 }, (_, hour) => {
  const value = `${String(hour).padStart(2, '0')}:00`
  return { label: value, value }
})
const renewalDraft = reactive({
  classNumEnabled: true,
  classNum: 5,
  validityDayEnabled: true,
  validityDay: 15,
  priceEnabled: false,
  price: 500,
})
const leaveLimitDraft = reactive({ cycle: '3', count: 2, type: '1' })
const leaveTimeDraft = ref<number | null>(1)
const leaveTimeError = ref('')

const leaveCycleValueMap: Record<string, string> = {
  none: '0',
  day: '1',
  week: '2',
  month: '3',
  quarter: '4',
  year: '5',
}

const leaveCycleTextMap: Record<string, string> = {
  0: '不限制',
  1: '每天',
  2: '每周',
  3: '每月',
  4: '每季度',
  5: '每年',
}

const leaveTypeValueMap: Record<string, string> = {
  course: '1',
  student: '2',
}

const leaveTypeTextMap: Record<string, string> = {
  1: '按课程',
  2: '按学员',
}

const moduleConfig = ref<Partial<InstConfig>>({})
const instConfig = computed<Partial<InstConfig>>(() => moduleConfig.value)

function isSwitchEnabled(value: unknown, defaultValue = false) {
  if (typeof value === 'boolean')
    return value
  if (typeof value === 'number')
    return value !== 0
  if (typeof value === 'string') {
    const normalized = value.trim().toLowerCase()
    if (['1', 'true', 'yes'].includes(normalized))
      return true
    if (['0', 'false', 'no'].includes(normalized))
      return false
  }
  return defaultValue
}

function getNumberText(value: unknown, fallback: string) {
  const text = value == null ? '' : String(value).trim()
  return text || fallback
}

function normalizeLeaveCycleValue(value: unknown) {
  const text = getNumberText(value, '3')
  return leaveCycleValueMap[text] || (leaveCycleTextMap[text] ? text : '3')
}

function normalizeLeaveTypeValue(value: unknown) {
  const text = getNumberText(value, '1')
  return leaveTypeValueMap[text] || (leaveTypeTextMap[text] ? text : '1')
}

function isRowLoading(key: string) {
  return Boolean(rowLoadingMap.value[key])
}

async function loadHomeSchoolConfig() {
  const res = await getInstConfigModuleApi('homeSchool')
  moduleConfig.value = res.result || {}
}

function mergeHomeSchoolConfig(patch: Partial<InstConfig>) {
  moduleConfig.value = {
    ...moduleConfig.value,
    ...patch,
  }
}

async function updateConfigField(field: keyof InstConfig, value: InstConfig[keyof InstConfig], key: string, successText: string) {
  rowLoadingMap.value = {
    ...rowLoadingMap.value,
    [key]: true,
  }
  try {
    const payload = { [field]: value } as Partial<InstConfig>
    if (field === 'leaveApplyCycleLimit')
      payload.leaveApplyCycleLimit = normalizeLeaveCycleValue(value)
    if (field === 'leaveApplyTypeLimit')
      payload.leaveApplyTypeLimit = normalizeLeaveTypeValue(value)
    await setInstConfigModuleApi('homeSchool', payload)
    mergeHomeSchoolConfig(payload)
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

function openBirthdayModal(enableAfterConfirm = false) {
  birthdayEnableAfterConfirm.value = enableAfterConfirm
  birthdayDraft.value = [...birthdayReceivers.value]
  birthdayModalOpen.value = true
}

async function saveBirthdayReceivers() {
  birthdayReceivers.value = [...birthdayDraft.value]
  if (birthdayEnableAfterConfirm.value)
    await updateConfigField('autoSendBirthdayMessage', true as InstConfig[keyof InstConfig], 'birthday', '已开启生日祝福')
  birthdayEnableAfterConfirm.value = false
  birthdayModalOpen.value = false
  messageService.success('发送学员已保存')
}

function openSmsClassReminderModal() {
  smsClassReminderHourDraft.value = getNumberText(instConfig.value.sendClassReminderSmsHour, '19:00')
  smsClassReminderModalOpen.value = true
}

async function saveSmsClassReminderHour() {
  await updateConfigField('sendClassReminderSmsHour', smsClassReminderHourDraft.value as InstConfig[keyof InstConfig], 'smsClassReminderHour', '上课提醒短信发送时间已保存')
  await updateConfigField('enabledClassReminder', true as InstConfig[keyof InstConfig], 'classSms', '已开启上课提醒')
  smsClassReminderModalOpen.value = false
}

function openClassReminderModal() {
  classReminderHourDraft.value = getNumberText(instConfig.value.sendClassReminderMsgHour, '19:00')
  classReminderModalOpen.value = true
}

async function saveClassReminderHour() {
  await updateConfigField('sendClassReminderMsgHour', classReminderHourDraft.value as InstConfig[keyof InstConfig], 'classReminderHour', '上课提醒时间已保存')
  classReminderModalOpen.value = false
}

function openRenewalModal() {
  renewalDraft.classNumEnabled = isSwitchEnabled(instConfig.value.enableRenewClassNum, true)
  renewalDraft.classNum = Number(getNumberText(instConfig.value.renewClassNum, '5')) || 5
  renewalDraft.validityDayEnabled = isSwitchEnabled(instConfig.value.enableRenewValidityDay, true)
  renewalDraft.validityDay = Number(getNumberText(instConfig.value.renewValidityDay, '15')) || 15
  renewalDraft.priceEnabled = isSwitchEnabled(instConfig.value.enableRenewPrice, false)
  renewalDraft.price = Number(getNumberText(instConfig.value.renewPrice, '500')) || 500
  renewalModalOpen.value = true
}

async function saveRenewalCondition() {
  await updateConfigField('enableRenewClassNum', renewalDraft.classNumEnabled as InstConfig[keyof InstConfig], 'renewConditionClassEnabled', '续费提醒条件已保存')
  await updateConfigField('renewClassNum', String(renewalDraft.classNum) as InstConfig[keyof InstConfig], 'renewConditionClass', '续费提醒条件已保存')
  await updateConfigField('enableRenewValidityDay', renewalDraft.validityDayEnabled as InstConfig[keyof InstConfig], 'renewConditionDayEnabled', '续费提醒条件已保存')
  await updateConfigField('renewValidityDay', String(renewalDraft.validityDay) as InstConfig[keyof InstConfig], 'renewConditionDay', '续费提醒条件已保存')
  await updateConfigField('enableRenewPrice', renewalDraft.priceEnabled as InstConfig[keyof InstConfig], 'renewConditionPriceEnabled', '续费提醒条件已保存')
  await updateConfigField('renewPrice', String(renewalDraft.price) as InstConfig[keyof InstConfig], 'renewConditionPrice', '续费提醒条件已保存')
  renewalModalOpen.value = false
}

function openLeaveCountModal(enableAfterConfirm = false) {
  leaveCountEnableAfterConfirm.value = enableAfterConfirm
  leaveLimitDraft.cycle = normalizeLeaveCycleValue(instConfig.value.leaveApplyCycleLimit)
  leaveLimitDraft.count = Number(getNumberText(instConfig.value.leaveApplyNumberLimit, '2')) || 2
  leaveLimitDraft.type = normalizeLeaveTypeValue(instConfig.value.leaveApplyTypeLimit)
  leaveCountModalOpen.value = true
}

async function saveLeaveCountLimit() {
  rowLoadingMap.value = {
    ...rowLoadingMap.value,
    leaveCycle: true,
    leaveCount: true,
    leaveType: true,
    leaveNumberLimit: true,
  }
  try {
    const payload = {
      leaveApplyCycleLimit: normalizeLeaveCycleValue(leaveLimitDraft.cycle),
      leaveApplyNumberLimit: String(leaveLimitDraft.count),
      leaveApplyTypeLimit: normalizeLeaveTypeValue(leaveLimitDraft.type),
      enableLeaveApplyNumberLimit: leaveCountEnableAfterConfirm.value ? true : instConfig.value.enableLeaveApplyNumberLimit,
    }
    await setInstConfigModuleApi('homeSchool', payload as Partial<InstConfig>)
    mergeHomeSchoolConfig(payload as Partial<InstConfig>)
    messageService.success(leaveCountEnableAfterConfirm.value ? '已开启请假次数限制' : '请假次数限制已保存')
    leaveCountEnableAfterConfirm.value = false
    leaveCountModalOpen.value = false
  }
  catch (error) {
    console.error('save leave count limit failed', error)
    messageService.error('保存失败，请稍后重试')
  }
  finally {
    rowLoadingMap.value = {
      ...rowLoadingMap.value,
      leaveCycle: false,
      leaveCount: false,
      leaveType: false,
      leaveNumberLimit: false,
    }
  }
}

function openLeaveTimeModal(enableAfterConfirm = false) {
  leaveTimeEnableAfterConfirm.value = enableAfterConfirm
  leaveTimeError.value = ''
  leaveTimeDraft.value = enableAfterConfirm ? null : Number(getNumberText(instConfig.value.leaveApplyTimeLimit, '1.0')) || 1
  leaveTimeModalOpen.value = true
}

async function saveLeaveTimeLimit() {
  if (leaveTimeDraft.value == null || Number.isNaN(Number(leaveTimeDraft.value))) {
    leaveTimeError.value = '内容不可为空'
    return
  }
  await updateConfigField('leaveApplyTimeLimit', leaveTimeDraft.value as InstConfig[keyof InstConfig], 'leaveTime', '请假时间限制已保存')
  if (leaveTimeEnableAfterConfirm.value)
    await updateConfigField('enableLeaveApplyTimeLimit', true as InstConfig[keyof InstConfig], 'leaveTimeLimit', '已开启请假时间限制')
  leaveTimeEnableAfterConfirm.value = false
  leaveTimeModalOpen.value = false
}

function handleLeaveCountToggle(checked: boolean) {
  if (checked) {
    openLeaveCountModal(true)
    return
  }
  Modal.confirm({
    title: '关闭请假次数限制',
    content: '关闭后，学员在公众号内请假将不受次数限制；每个学员单独设置的请假次数将清空，无法恢复',
    okText: '关闭',
    cancelText: '我再想想',
    centered: true,
    async onOk() {
      await updateConfigField('enableLeaveApplyNumberLimit', false as InstConfig[keyof InstConfig], 'leaveNumberLimit', '已关闭请假次数限制')
    },
  })
}

function handleLeaveTimeToggle(checked: boolean) {
  if (checked) {
    openLeaveTimeModal(true)
    return
  }
  Modal.confirm({
    title: '关闭请假时间限制',
    content: '关闭后，学员在公众号内课前请假将不受时间限制',
    okText: '关闭',
    cancelText: '我再想想',
    centered: true,
    async onOk() {
      await updateConfigField('enableLeaveApplyTimeLimit', false as InstConfig[keyof InstConfig], 'leaveTimeLimit', '已关闭请假时间限制')
    },
  })
}

function receiverText() {
  const names: Record<string, string> = {
    current: '在读学员',
    prospect: '意向学员',
    history: '历史学员',
  }
  return birthdayReceivers.value.map(item => names[item]).filter(Boolean).join('、') || '未选择'
}

const smsRows = computed(() => [
  {
    key: 'birthday',
    title: '生日祝福',
    field: 'autoSendBirthdayMessage' as keyof InstConfig,
    desc: '开启后，学员生日当天系统自动发送祝福短信',
    extra: `发送给：${receiverText()}`,
    editable: true,
    onEdit: () => openBirthdayModal(false),
  },
  {
    key: 'recharge',
    title: '储值账户变动记录提醒',
    field: 'enableRechargeAccountChangeMessage' as keyof InstConfig,
    desc: '开启后，储值账户发生的变动记录（包含充值、消费、退款）都会给家长自动发送提醒短信',
  },
  {
    key: 'classSms',
    title: '上课提醒',
    field: 'enabledClassReminder' as keyof InstConfig,
    desc: '开启后，学员上课前会发送短信提醒：上课时间及相关的课程内容',
    extra: `上课前一日 ${getNumberText(instConfig.value.sendClassReminderSmsHour, '19:00')} 发送短信至家长`,
    editable: true,
    onEdit: openSmsClassReminderModal,
    sample: '短信内容举例：\n【校宝】上课提醒：西西明天（05-17 11:00）在校宝艺木（黄龙校区）有1节课：美术课，请提前安排好日程，准时来上课哦～',
  },
  {
    key: 'consume',
    title: '课消提醒',
    field: 'enabledClassConsumptionReminder' as keyof InstConfig,
    desc: '开启后，老师上课点名后，将会给学员发送对应的课消短信',
  },
  {
    key: 'audition',
    title: '试听短信提醒',
    field: 'enableAuditionSmsRemind' as keyof InstConfig,
    desc: '开启后，创建的试听课程，将在开课前 2 小时以短信方式提醒到学员联系人',
  },
  {
    key: 'coupon',
    title: '优惠券领取/优惠券到期提醒',
    field: 'enableSendCouponRemindSms' as keyof InstConfig,
    desc: '开启后，将会发送给学员：优惠券领取成功、优惠券到期（过期当天 10:00）短信提醒',
  },
  {
    key: 'deleteStudent',
    title: '删除学员短信验证提示',
    field: 'enableSendChildBindNoticeToAdmin' as keyof InstConfig,
    desc: '开启后，当日删除学员超 50 人后再次删除学员会触发超级管理员的短信验证',
  },
])

const remindRows = computed(() => [
  {
    key: 'bill',
    title: '报名续费交易提醒',
    field: 'enableTeachingBillRemindSms' as keyof InstConfig,
    desc: '开启后，家长将在家校公众号收到报名续费成功的推送提醒',
  },
  {
    key: 'classMsg',
    title: '上课提醒',
    field: 'enabledClassReminder' as keyof InstConfig,
    desc: '开启后，家长将在家校公众号接收学员明日的上课安排',
    extra: `上课前一日 ${getNumberText(instConfig.value.sendClassReminderMsgHour, '19:00')} 发送公众号消息至家长`,
    editable: true,
    onEdit: openClassReminderModal,
  },
  {
    key: 'consumeMsg',
    title: '课消提醒',
    field: 'enabledClassConsumptionReminder' as keyof InstConfig,
    desc: '开启后，当老师完成点名，家长会实时接收学员的点名记录',
  },
  {
    key: 'correct',
    title: '纠正点名记录提醒',
    field: 'studentAbsentClassSwitch' as keyof InstConfig,
    desc: '开启后，纠正点名会通过系统自动发送微信提醒给家长',
  },
  {
    key: 'renew',
    title: '续费提醒',
    field: 'enabledRenewReminder' as keyof InstConfig,
    desc: '开启后，家长将会在剩余课时或有效期或金额达到提醒条件时收到续费提醒',
    extra: `提醒条件：课时数不足 ${getNumberText(instConfig.value.renewClassNum, '5')} 课时/有效期不足 ${getNumberText(instConfig.value.renewValidityDay, '15')} 天`,
    editable: true,
    onEdit: openRenewalModal,
  },
  {
    key: 'arrears',
    title: '消超消息提醒',
    field: 'enableArrearagedSendMessage' as keyof InstConfig,
    desc: '开启后，家长端公众号将收到推送提醒，如“xx学员学费不足，本次欠费xx课时”',
  },
  {
    key: 'liquidation',
    title: '清算提醒',
    field: 'enableLiquidationRemindMessage' as keyof InstConfig,
    desc: '开启后，清算成功后会自动发送消息给家长',
  },
  {
    key: 'point',
    title: '积分变动提醒通知',
    field: 'enablePointChangeRemindMessage' as keyof InstConfig,
    desc: '开启后，学员积分发生变动会及时提醒通知家长',
  },
  {
    key: 'bindAdmin',
    title: '家长绑定学员提醒至管理员',
    field: 'enableOrgSendChildBindNoticeToAdmin' as keyof InstConfig,
    desc: '开启后，校区管理员将在App端消息中心收到家长绑定学员的推送提醒',
  },
])

function shouldShowRowExtra(row: { field: keyof InstConfig, extra?: string }) {
  return Boolean(row.extra) && isSwitchEnabled(instConfig.value[row.field])
}

async function toggleRow(row: { field: keyof InstConfig, key: string, title: string }, checked: boolean) {
  if (row.key === 'birthday' && checked) {
    openBirthdayModal(true)
    return
  }
  if (row.key === 'classSms' && checked) {
    openSmsClassReminderModal()
    return
  }
  await updateConfigField(row.field, checked as InstConfig[keyof InstConfig], row.key, checked ? `已开启${row.title}` : `已关闭${row.title}`)
}

const leaveCycleText = computed(() => {
  return leaveCycleTextMap[normalizeLeaveCycleValue(instConfig.value.leaveApplyCycleLimit)] || '每月'
})

const leaveTypeText = computed(() => leaveTypeTextMap[normalizeLeaveTypeValue(instConfig.value.leaveApplyTypeLimit)] || '按课程')

onMounted(() => {
  void loadHomeSchoolConfig()
})
</script>

<template>
  <div class="home-school-settings-page">
    <a-tabs v-model:active-key="activeKey" :animated="false" class="home-school-settings-page__tabs">
      <a-tab-pane key="message" tab="消息设置">
        <div class="home-school-settings-page__pane">
          <a-tabs v-model:active-key="activeMessageKey" :animated="false" class="home-school-settings-page__sub-tabs">
            <a-tab-pane key="sms" tab="短信设置">
              <section class="settings-section">
                <div class="settings-section__title">短信设置</div>
                <div class="settings-table">
                  <div v-for="row in smsRows" :key="row.key" class="settings-row">
                    <div class="settings-row__label">{{ row.title }}</div>
                    <div class="settings-row__content">
                      <a-switch
                        :checked="isSwitchEnabled(instConfig[row.field])"
                        :loading="isRowLoading(row.key)"
                        @change="checked => toggleRow(row, Boolean(checked))"
                      />
                      <div class="settings-desc">{{ row.desc }}</div>
                      <div v-if="shouldShowRowExtra(row)" class="settings-extra">
                        {{ row.extra }}
                        <a-button v-if="row.editable" type="link" size="small" class="settings-link" @click="row.onEdit?.()">
                          <template #icon><EditOutlined /></template>
                          编辑
                        </a-button>
                      </div>
                      <div v-if="row.sample && isSwitchEnabled(instConfig[row.field])" class="sms-sample-card">
                        <div v-for="line in row.sample.split('\n')" :key="line">{{ line }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </section>
            </a-tab-pane>

            <a-tab-pane key="remind" tab="消息提醒设置">
              <section class="settings-section">
                <div class="settings-section__title">消息提醒设置</div>
                <div class="settings-alert"><InfoCircleFilled />以下消息均通过公众号发送，请尽快让家长关注学员和公众号</div>
                <div class="settings-table">
                  <div v-for="row in remindRows" :key="row.key" class="settings-row">
                    <div class="settings-row__label">{{ row.title }}</div>
                    <div class="settings-row__content">
                      <a-switch
                        :checked="isSwitchEnabled(instConfig[row.field])"
                        :loading="isRowLoading(row.key)"
                        @change="checked => toggleRow(row, Boolean(checked))"
                      />
                      <div class="settings-desc">{{ row.desc }}</div>
                      <div v-if="shouldShowRowExtra(row)" class="settings-extra">
                        {{ row.extra }}
                        <a-button v-if="row.editable" type="link" size="small" class="settings-link" @click="row.onEdit?.()">
                          <template #icon><EditOutlined /></template>
                          编辑
                        </a-button>
                      </div>
                    </div>
                  </div>
                </div>
              </section>
            </a-tab-pane>
          </a-tabs>
        </div>
      </a-tab-pane>

      <a-tab-pane key="leave" tab="请假设置">
        <div class="home-school-settings-page__pane">
          <section class="settings-section">
            <div class="settings-section__title">请假设置</div>
            <div class="settings-table settings-table--leave">
              <div class="settings-row settings-row--top" :class="{ 'settings-row--large': isSwitchEnabled(instConfig.enableLeaveApplyNumberLimit) }">
                <div class="settings-row__label">请假次数限制</div>
                <div class="settings-row__content">
                  <a-switch
                    :checked="isSwitchEnabled(instConfig.enableLeaveApplyNumberLimit)"
                    :loading="isRowLoading('leaveNumberLimit')"
                    @change="checked => handleLeaveCountToggle(Boolean(checked))"
                  />
                  <div class="settings-desc">开启后，学员发起请假会受到相应次数限制</div>
                  <div v-if="isSwitchEnabled(instConfig.enableLeaveApplyNumberLimit)" class="settings-extra">
                    请假次数限制设置：
                    <a-button type="link" size="small" class="settings-link" @click="openLeaveCountModal(false)">
                      <template #icon><EditOutlined /></template>
                      编辑
                    </a-button>
                  </div>
                  <div v-if="isSwitchEnabled(instConfig.enableLeaveApplyNumberLimit)" class="leave-rule-card">
                    <p>限制周期：{{ leaveCycleText }}</p>
                    <p>可请假次数：{{ getNumberText(instConfig.leaveApplyNumberLimit, '2') }}</p>
                    <p>请假类型：{{ leaveTypeText }}</p>
                    <ol>
                      <li>“可请假次数”指学员在每门课程（包括通用课下所有对应课程）能请假的次数，超过限制后学员无法发起请假，但不影响机构发起请假代办</li>
                      <li>学员请假1节课即消耗1次请假次数</li>
                      <li>仅当设置开启时，学员请假才会计入请假次数，如本月15日开启设置，则15日前学员的请假记录不统计在内</li>
                      <li>修改请假类型、限制周期，每个学员单独设置的请假次数将清空，无法恢复</li>
                    </ol>
                  </div>
                </div>
              </div>

              <div class="settings-row">
                <div class="settings-row__label">请假时间限制</div>
                <div class="settings-row__content">
                  <a-switch
                    :checked="isSwitchEnabled(instConfig.enableLeaveApplyTimeLimit)"
                    :loading="isRowLoading('leaveTimeLimit')"
                    @change="checked => handleLeaveTimeToggle(Boolean(checked))"
                  />
                  <div class="settings-desc">开启后，学员需在规定时间范围内发起请假</div>
                  <div v-if="isSwitchEnabled(instConfig.enableLeaveApplyTimeLimit)" class="settings-extra">
                    学员需至少在开课前 <span class="primary-text">{{ getNumberText(instConfig.leaveApplyTimeLimit, '1') }}</span> 小时发起请假
                    <a-button type="link" size="small" class="settings-link" @click="openLeaveTimeModal(false)">
                      <template #icon><EditOutlined /></template>
                      编辑
                    </a-button>
                  </div>
                </div>
              </div>
            </div>
          </section>
        </div>
      </a-tab-pane>

    </a-tabs>

    <a-modal v-model:open="birthdayModalOpen" title="发送学员" centered :width="400" wrap-class-name="home-school-settings-modal" @ok="saveBirthdayReceivers">
      <a-checkbox-group v-model:value="birthdayDraft" class="inline-checkboxes">
        <a-checkbox value="current">在读学员</a-checkbox>
        <a-checkbox value="prospect">意向学员</a-checkbox>
        <a-checkbox value="history">历史学员</a-checkbox>
      </a-checkbox-group>
    </a-modal>

    <a-modal v-model:open="smsClassReminderModalOpen" title="编辑上课提醒短信发送时间" centered :width="450" wrap-class-name="home-school-settings-modal class-reminder-time-modal" @ok="saveSmsClassReminderHour">
      <div class="compact-form-row">
        <span>开启后，上课前一日</span>
        <a-select
          v-model:value="smsClassReminderHourDraft"
          :options="classReminderHourOptions"
          class="class-reminder-time-select"
          popup-class-name="class-reminder-time-dropdown"
        />
        <span>发送短信至家长</span>
      </div>
    </a-modal>

    <a-modal v-model:open="classReminderModalOpen" title="编辑上课提醒时间" centered :width="450" wrap-class-name="home-school-settings-modal class-reminder-time-modal" @ok="saveClassReminderHour">
      <div class="compact-form-row">
        <span>开启后，上课前一日</span>
        <a-select
          v-model:value="classReminderHourDraft"
          :options="classReminderHourOptions"
          class="class-reminder-time-select"
          popup-class-name="class-reminder-time-dropdown"
        />
        <span>发送公众号消息至家长</span>
      </div>
    </a-modal>

    <a-modal v-model:open="renewalModalOpen" title="续费提醒" centered :width="552" wrap-class-name="home-school-settings-modal renewal-reminder-modal" @ok="saveRenewalCondition">
      <div class="renewal-reminder">
        <div class="renewal-reminder__tip">达到以下设置的条件后家长将会收到续费提醒</div>
        <div class="renewal-reminder__card">
          <div class="renewal-reminder__item">
            <div class="renewal-reminder__title">
              <span>续费提醒天数：</span>
              <a-switch v-model:checked="renewalDraft.classNumEnabled" />
            </div>
            <div v-if="renewalDraft.classNumEnabled" class="renewal-reminder__condition">
              <span>剩余课时数 ＜</span>
              <a-input-number v-model:value="renewalDraft.classNum" :min="0" class="renewal-reminder__input" />
              <span>课时</span>
            </div>
          </div>

          <div class="renewal-reminder__item">
            <div class="renewal-reminder__title">
              <span>有效期不足：</span>
              <a-switch v-model:checked="renewalDraft.validityDayEnabled" />
            </div>
            <div v-if="renewalDraft.validityDayEnabled" class="renewal-reminder__condition">
              <span>有效期天数 ＜</span>
              <a-input-number v-model:value="renewalDraft.validityDay" :min="0" class="renewal-reminder__input" />
              <span>天</span>
            </div>
          </div>

          <div class="renewal-reminder__item">
            <div class="renewal-reminder__title renewal-reminder__title--only">
              <span>金额不足：</span>
              <a-switch v-model:checked="renewalDraft.priceEnabled" />
            </div>
            <div v-if="renewalDraft.priceEnabled" class="renewal-reminder__condition">
              <span>金额 ＜</span>
              <a-input-number v-model:value="renewalDraft.price" :min="0" class="renewal-reminder__input" />
              <span>元</span>
            </div>
          </div>
        </div>
      </div>
    </a-modal>

    <a-modal v-model:open="leaveCountModalOpen" title="请假次数限制设置" centered :width="800" wrap-class-name="home-school-settings-modal" @ok="saveLeaveCountLimit">
      <a-form class="leave-limit-form" :label-col="{ style: { width: '96px' } }" :wrapper-col="{ flex: 1 }">
        <a-form-item label="限制周期" required>
          <a-radio-group v-model:value="leaveLimitDraft.cycle" class="custom-radio">
            <a-radio value="0">不限制</a-radio>
            <a-radio value="1">每天</a-radio>
            <a-radio value="2">每周</a-radio>
            <a-radio value="3">每月</a-radio>
            <a-radio value="4">每季度</a-radio>
            <a-radio value="5">每年</a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item label="可请假次数" required>
          <a-input-number v-model:value="leaveLimitDraft.count" :min="0" class="compact-input" />
        </a-form-item>
        <a-form-item label="请假类型" required>
          <a-radio-group v-model:value="leaveLimitDraft.type" class="custom-radio">
            <a-radio value="1">按课程</a-radio>
            <a-radio value="2">按学员</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
      <ol class="modal-tips">
        <li>“可请假次数”指学员在每门课程（包括通用课下所有对应课程）能请假的次数，超过限制后学员无法发起请假，但不影响机构发起请假代办</li>
        <li>学员请假1节课即消耗1次请假次数</li>
        <li>仅当设置开启时，学员请假才会计入请假次数，如本月15日开启设置，则15日前学员的请假记录不统计在内</li>
        <li>修改请假类型、限制周期，每个学员单独设置的请假次数将清空，无法恢复</li>
      </ol>
    </a-modal>

    <a-modal v-model:open="leaveTimeModalOpen" title="请假时间设置" centered :width="450" wrap-class-name="home-school-settings-modal" @ok="saveLeaveTimeLimit">
      <div class="compact-form-row">
        <span>学员至少在开课前</span>
        <div class="field-error-wrap">
          <a-input-number v-model:value="leaveTimeDraft" :min="0" :step="0.5" placeholder="请输入" class="compact-input" @change="leaveTimeError = ''" />
          <div v-if="leaveTimeError" class="field-error-text">{{ leaveTimeError }}</div>
        </div>
        <span>小时发起请假申请</span>
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.home-school-settings-page {
  width: 100%;
  min-height: calc(100vh - 120px);
  color: #262626;
}

.home-school-settings-page__tabs {
  :deep(.ant-tabs-nav) {
    margin: 0;
    padding: 0 12px;
    background: #fff;
    border-radius: 16px 16px 0 0;
  }

  :deep(.ant-tabs-nav-wrap) {
    padding-left: 24px;
  }

  :deep(.ant-tabs-tab) {
    padding: 12px 0;
    font-size: 15px;
  }

  :deep(.ant-tabs-ink-bar) {
    height: 9px !important;
    background: transparent !important;
    bottom: 1px !important;

    &::after {
      position: absolute;
      top: 0;
      left: calc(50% - 12px);
      width: 24px !important;
      height: 4px !important;
      border-radius: 2px;
      background-color: var(--pro-ant-color-primary, #1677ff);
      content: '';
    }
  }

  :deep(.ant-tabs-content-holder) {
    background: transparent;
  }
}

.home-school-settings-page__pane {
  min-height: 480px;
}

.home-school-settings-page__sub-tabs {
  padding: 0;

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

.settings-alert {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 44px;
  margin-bottom: 10px;
  padding: 0 16px;
  color: #1677ff;
  background: #e8f1ff;
  font-size: 14px;
}

.settings-table {
  overflow: hidden;
  border: 1px solid #edf0f5;
  background: #fff;
}

.settings-row {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  min-height: 84px;
  border-bottom: 1px solid #edf0f5;

  &:last-child {
    border-bottom: 0;
  }
}

.settings-row--large {
  min-height: 274px;
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

  :deep(.ant-switch) {
    display: inline-flex;
    flex: 0 0 auto;
    width: 44px;
    min-width: 44px;
  }
}

.settings-desc {
  margin-top: 8px;
  color: #222;
  font-size: 14px;
  line-height: 20px;
}

.settings-extra {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 12px;
  color: #222;
  font-size: 14px;
  line-height: 20px;
}

.settings-link {
  padding: 0 4px;
  font-size: 14px;
}

.sms-sample-card {
  max-width: 860px;
  margin-top: 12px;
  padding: 14px 16px;
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  color: #595959;
  background: #fff;
  font-size: 14px;
  line-height: 24px;
}

.leave-rule-card {
  max-width: 1000px;
  margin-top: 14px;
  padding: 14px 36px;
  border: 1px solid #edf0f5;
  border-radius: 8px;
  color: #222;
  background: #fafafa;
  line-height: 24px;

  p {
    margin: 0 0 6px;
  }

  ol {
    margin: 2px 0 0;
    padding-left: 0;
    color: #595959;
    list-style-position: inside;
  }
}

.primary-text {
  color: var(--pro-ant-color-primary, #1677ff);
}

.inline-checkboxes {
  display: flex;
  gap: 18px;
  padding: 4px 0 8px;
}

.compact-form-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  color: #595959;
}

.class-reminder-time-select {
  width: 100px;
}

.compact-input {
  width: 88px;
}

.field-error-wrap {
  position: relative;
  display: inline-flex;
  flex-direction: column;
}

.field-error-text {
  position: absolute;
  top: 34px;
  left: 0;
  color: #ff4d4f;
  font-size: 12px;
  line-height: 18px;
  white-space: nowrap;
}

.leave-limit-form {
  :deep(.ant-form-item) {
    margin-bottom: 12px;
  }

  :deep(.ant-form-item-label) {
    text-align: right;
  }

  :deep(.ant-form-item-label > label) {
    justify-content: flex-end;
    width: 100%;
  }

  :deep(.ant-form-item-control) {
    min-width: 0;
  }
}

.modal-tips {
  margin: 4px 0 0;
  color: #595959;
  line-height: 24px;
}

.renewal-reminder__tip {
  margin-bottom: 10px;
  color: #595959;
  font-size: 14px;
  line-height: 22px;
}

.renewal-reminder__card {
  overflow: hidden;
  border: 1px solid #d9d9d9;
  border-radius: 8px;
  background: #fff;
}

.renewal-reminder__item {
  padding: 16px;

  & + & {
    border-top: 1px solid #f0f0f0;
  }
}

.renewal-reminder__title {
  display: flex;
  align-items: center;
  gap: 10px;
  color: #1f1f1f;
  font-size: 14px;
  font-weight: 600;
  line-height: 22px;

  :deep(.ant-switch) {
    width: 44px;
    min-width: 44px;
  }
}

.renewal-reminder__title--only {
  min-height: 24px;
}

.renewal-reminder__condition {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 12px;
  color: #595959;
  font-size: 14px;
  line-height: 22px;
}

.renewal-reminder__input {
  width: 90px;
}

@media (max-width: 768px) {
  .home-school-settings-page__tabs :deep(.ant-tabs-nav-wrap) {
    padding-left: 12px;
  }

  .settings-row {
    grid-template-columns: 1fr;
  }

  .settings-row__label {
    justify-content: flex-start;
    border-right: 0;
    border-bottom: 1px solid #edf0f5;
  }

  .settings-section {
    padding: 16px 12px;
  }
}
</style>

<style lang="less">
.home-school-settings-modal {
  .ant-modal-content {
    border-radius: 14px;
    overflow: hidden;
  }

  .ant-modal-header {
    padding: 18px 24px 12px;
    margin-bottom: 0;
  }

  .ant-modal-title {
    font-size: 18px;
    font-weight: 600;
  }

  .ant-modal-body {
    padding: 24px;
  }

  .ant-modal-footer {
    padding: 12px 24px;
    border-top: 1px solid #f0f0f0;
  }
}

.renewal-reminder-modal {
  .ant-modal-body {
    padding: 24px 24px 22px;
  }
}

.class-reminder-time-modal {
  .ant-modal-body {
    padding: 24px;
  }
}

.class-reminder-time-dropdown {
  width: 100px !important;
  min-width: 100px !important;

  .ant-select-item {
    min-height: 32px;
    padding: 5px 12px;
  }
}
</style>
