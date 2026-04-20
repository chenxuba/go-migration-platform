<script setup>
import { CloseOutlined, ExclamationCircleFilled } from '@ant-design/icons-vue'
import { computed, ref, watch } from 'vue'
import { getApprovalTemplatesApi, saveApprovalTemplatesApi } from '@/api/finance-center/approval-manage'
import messageService from '@/utils/messageService'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },

})
const emit = defineEmits(['update:open'])
// 处理双向绑定
const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const registrationRenewalChecked = ref(false) // 报名续费
const transferCourseChecked = ref(false) // 转课
const refundCourseChecked = ref(false) // 退课
const rechargeValueChecked = ref(false) // 储值充值
const rechargeRefundChecked = ref(false) // 储值退费
const tuitionRefundChecked = ref(false) // 退学杂教材费

const loading = ref(false)
const saving = ref(false)
const templates = ref([])
const submitAttempted = ref(0)

const templateDefs = [
  { type: 1, key: 'registrationRenewalChecked' },
  { type: 2, key: 'transferCourseChecked' },
  { type: 3, key: 'refundCourseChecked' },
  { type: 4, key: 'rechargeValueChecked' },
  { type: 5, key: 'rechargeRefundChecked' },
  { type: 6, key: 'tuitionRefundChecked' },
]

const latestUpdateText = computed(() => {
  const list = (templates.value || []).filter(item => item.updatedTime)
  if (list.length === 0)
    return '-'
  const latest = [...list].sort((a, b) => new Date(b.updatedTime).getTime() - new Date(a.updatedTime).getTime())[0]
  return `${latest.updatedStaffName || '-'} ${String(latest.updatedTime).replace('T', ' ').slice(0, 16)}`
})

function getTemplate(type) {
  return templates.value.find(item => item.type === type) || {
    id: '0',
    type,
    enable: false,
    ruleJson: '',
    flowModels: [],
  }
}

function setCheckedByType(type, checked) {
  if (type === 1) registrationRenewalChecked.value = checked
  if (type === 2) transferCourseChecked.value = checked
  if (type === 3) refundCourseChecked.value = checked
  if (type === 4) rechargeValueChecked.value = checked
  if (type === 5) rechargeRefundChecked.value = checked
  if (type === 6) tuitionRefundChecked.value = checked
}

async function loadTemplates() {
  try {
    loading.value = true
    const res = await getApprovalTemplatesApi()
    if (res.code === 200) {
      templates.value = Array.isArray(res.result) ? res.result : []
      templateDefs.forEach(item => {
        setCheckedByType(item.type, !!getTemplate(item.type).enable)
      })
      return
    }
    messageService.error(res.message || '获取审批规则失败')
  }
  catch (error) {
    console.error('获取审批规则失败:', error)
    messageService.error('获取审批规则失败')
  }
  finally {
    loading.value = false
  }
}

async function handleSubmit() {
  submitAttempted.value += 1

  const registrationTemplate = getTemplate(1)
  const registrationRuleState = registrationTemplate?.__ruleState || {}
  const registrationHasAnyRule = [
    registrationRuleState.giveClassTimeChecked,
    registrationRuleState.givePriceChecked,
    registrationRuleState.giveDaysChecked,
    registrationRuleState.discountRateChecked,
    registrationRuleState.discountAmountChecked,
  ].some(Boolean)
  const registrationHasInvalidValue = (
    (registrationRuleState.giveClassTimeChecked && !(Number(registrationRuleState.giveClassTime) > 0))
    || (registrationRuleState.givePriceChecked && !(Number(registrationRuleState.givePrice) > 0))
    || (registrationRuleState.giveDaysChecked && !(Number(registrationRuleState.giveDays) > 0))
    || (registrationRuleState.discountRateChecked && !(Number(registrationRuleState.discountRate) > 0))
    || (registrationRuleState.discountAmountChecked && !(Number(registrationRuleState.discountAmount) > 0))
  )

  if (registrationTemplate.enable && registrationTemplate.__approvalCriteria === '2' && !registrationHasAnyRule) {
    messageService.error('请至少选择1项【报名续费】的限制条件')
    return
  }

  if (registrationTemplate.enable && registrationTemplate.__approvalCriteria === '2' && registrationHasInvalidValue) {
    messageService.error('请完善【报名续费】的限制条件')
    return
  }

  const invalidRuleTemplate = templateDefs
    .map(item => getTemplate(item.type))
    .find(item => item.enable && item.type === 1 && item.__approvalCriteria === '2' && !String(item.ruleJson || '').trim())

  if (invalidRuleTemplate) {
    messageService.error(`请至少选择1项【${invalidRuleTemplate.name || '报名续费'}】的限制条件`)
    return
  }

  const invalidTemplate = templateDefs
    .map(item => getTemplate(item.type))
    .find(item => item.enable && (!Array.isArray(item.flowModels) || item.flowModels.length === 0))

  if (invalidTemplate) {
    messageService.error(`请配置【${invalidTemplate.name || '当前办理类型'}】审批流程`)
    return
  }

  try {
    saving.value = true
    const payload = {
      approveTemplateRequests: templateDefs.map(item => {
        const current = getTemplate(item.type)
        return {
          id: Number(current.id || 0),
          type: item.type,
          enable: !!current.enable,
          ruleJson: current.ruleJson || '',
          flowRequestModels: current.enable
            ? (current.flowModels || []).map(flow => ({
              step: flow.step,
              staffIds: (flow.staffIds || []).map(id => Number(id)),
            }))
            : [],
        }
      }),
    }
    const res = await saveApprovalTemplatesApi(payload)
    if (res.code === 200) {
      messageService.success('审批规则保存成功')
      await loadTemplates()
      openDrawer.value = false
      return
    }
    messageService.error(res.message || '保存审批规则失败')
  }
  catch (error) {
    console.error('保存审批规则失败:', error)
    messageService.error('保存审批规则失败')
  }
  finally {
    saving.value = false
  }
}

function updateTemplate(type, patch) {
  const index = templates.value.findIndex(item => item.type === type)
  const current = getTemplate(type)
  const nextValue = {
    ...current,
    ...patch,
  }
  if (index >= 0) {
    templates.value[index] = nextValue
  }
  else {
    templates.value.push(nextValue)
  }
}

function handleCheckedChange(type, checked) {
  updateTemplate(type, { enable: checked })
}

watch(() => props.open, (open) => {
  if (open) {
    loadTemplates()
  }
}, { immediate: true })
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer" :push="{ distance: 80 }" :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false" width="800px" placement="right"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            配置审批规则
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <a-alert type="info" class="rounded-0 border-none">
        <template #message>
          <div class="text-#06f">
            <div>
              <ExclamationCircleFilled class="text-#06f" />
              以下办理类型开启审批后 ，满足规则的订单将会生成审批单，审批通过后订单才会生效。
            </div>
            <div class="pl-17px">
              当前规则配置更新于：{{ latestUpdateText }}
            </div>
          </div>
        </template>
      </a-alert>
      <div class="contenter">
        <a-space class="flex flex-col items-start" :size="24">
          <!-- 报名续费 -->
          <registration-renewal
            v-model:checked="registrationRenewalChecked"
            :template-data="getTemplate(1)"
            :submit-attempted="submitAttempted"
            @update:checked="handleCheckedChange(1, $event)"
            @update:template-data="updateTemplate(1, $event)"
          />
          <!-- 转课 -->
          <transfer-course
            v-model:checked="transferCourseChecked"
            :template-data="getTemplate(2)"
            @update:checked="handleCheckedChange(2, $event)"
            @update:template-data="updateTemplate(2, $event)"
          />
          <!-- 退课 -->
          <refund-course
            v-model:checked="refundCourseChecked"
            :template-data="getTemplate(3)"
            @update:checked="handleCheckedChange(3, $event)"
            @update:template-data="updateTemplate(3, $event)"
          />
          <!-- 储值充值 -->
          <recharge-value
            v-model:checked="rechargeValueChecked"
            :template-data="getTemplate(4)"
            @update:checked="handleCheckedChange(4, $event)"
            @update:template-data="updateTemplate(4, $event)"
          />
          <!-- 储值退费 -->
          <recharge-refund
            v-model:checked="rechargeRefundChecked"
            :template-data="getTemplate(5)"
            @update:checked="handleCheckedChange(5, $event)"
            @update:template-data="updateTemplate(5, $event)"
          />
          <!-- 退学杂教材费 -->
          <tuition-refund
            v-model:checked="tuitionRefundChecked"
            :template-data="getTemplate(6)"
            @update:checked="handleCheckedChange(6, $event)"
            @update:template-data="updateTemplate(6, $event)"
          />
        </a-space>
      </div>
      <template #footer>
        <div class="drawer-footer">
          <a-button class="submit-btn" type="primary" :loading="saving" @click="handleSubmit">
            确定
          </a-button>
        </div>
      </template>
    </a-drawer>
  </div>
</template>

<style lang="less" scoped>
/* 添加旋转动画 */
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.contenter {
  padding: 10px 24px;
  background: #fff;
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  height: 60px;
  padding: 0 24px;
  background: #fff;
}

.submit-btn {
  width: 140px;
  height: 48px;
  border-radius: 12px;
  font-size: 18px;
  font-weight: 600;
}
</style>
