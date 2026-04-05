<script setup>
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { computed, createVNode, ref } from 'vue'
import { CAMPUS_DATA_CLEAR_SCOPE_BUSINESS_ONLY, clearCampusDataApi } from '@/api/business-settings/campus-data-clear'
import messageService from '@/utils/messageService'

const props = defineProps({
  embedded: {
    type: Boolean,
    default: false,
  },
})

const acknowledged = ref(false)
const submitting = ref(false)

const recommendedScope = {
  value: CAMPUS_DATA_CLEAR_SCOPE_BUSINESS_ONLY,
  title: '只清校区业务数据',
  description: '适合校区重新起盘、演示环境回收、历史业务重置等场景，不影响基础配置继续使用。',
  includes: [
    '学员主档、学员自定义字段值与学员变更记录',
    '跟进记录',
    '订单、订单课程明细与支付记录',
    '审批记录与审批历史',
    '学费账户、充值账户及相关流水、业务台账',
    '意向学员/订单导入记录与学员导出记录',
    '套餐、套餐内商品、套餐属性结果',
    '班级与 1 对 1（含班员、教师关联）',
  ],
  excludes: [
    '员工、角色、部门等组织信息',
    '校区信息',
    '功能配置与业务设置',
    '渠道与渠道分类',
    '课程与课程配置（详情、报价、属性结果、销量）',
    '课程分类、课程属性、课程属性选项',
    '订单标签',
    '学员自定义字段定义',
    '审批模板与审批流配置',
  ],
}

const confirmContent = computed(() => `本次将按“${recommendedScope.title}”执行。

会清空：${recommendedScope.includes.join('、')}。

会保留：${recommendedScope.excludes.join('、')}。

清空后相关业务数据通常无法恢复，请确认已获得管理员授权。`)

function buildSuccessMessage(result) {
  const cleared = result?.cleared
  if (!cleared)
    return result?.intentStudentIndexMessage || '校区业务数据已清空'

  const summary = `已清空学员 ${cleared.students || 0} 条、订单 ${cleared.orders || 0} 条、审批 ${cleared.approvalRecords || 0} 条、1对1/班级 ${cleared.teachingClasses || 0} 条`
  return result?.intentStudentIndexMessage ? `${summary}，${result.intentStudentIndexMessage}` : summary
}

function handleClearClick() {
  if (!acknowledged.value) {
    messageService.warning('请先确认已知晓清空风险与保留范围')
    return
  }

  Modal.confirm({
    title: '确认清空当前校区业务数据？',
    icon: createVNode(ExclamationCircleOutlined),
    centered: true,
    okText: '确认清空',
    okType: 'danger',
    cancelText: '取消',
    content: confirmContent.value,
    async onOk() {
      submitting.value = true
      try {
        const res = await clearCampusDataApi({ scope: recommendedScope.value })
        if (res.code === 200) {
          messageService.success(buildSuccessMessage(res.data))
          return
        }
        messageService.error(res.message || '操作失败')
        return Promise.reject(new Error(res.message || '操作失败'))
      }
      catch (e) {
        console.error(e)
        messageService.error('请求失败，请稍后重试')
        return Promise.reject(e)
      }
      finally {
        submitting.value = false
      }
    },
  })
}
</script>

<template>
  <div class="campus-data-clear" :class="{ 'campus-data-clear--embedded': props.embedded }">
    <div class="settings-panel">
      <div class="panel-title">
        <span class="panel-title__marker" />
        <span>校区数据清空</span>
      </div>
      <div class="warning-banner">
        <ExclamationCircleOutlined class="warning-banner__icon" />
        <span>校区数据清空后，无法进行恢复，请谨慎操作！</span>
      </div>

      <div class="summary-text">
        清空数据后，系统仍会保留员工信息、校区信息、功能配置等基础信息，渠道、课程配置等资料也会继续保留，方便后续继续使用。
      </div>

      <div class="scope-grid">
        <div class="scope-card">
          <div class="scope-card__title">
            <span class="scope-card__dot" />
            <span>本次会清空</span>
          </div>
          <div class="scope-card__desc">
            {{ recommendedScope.description }}
          </div>
          <ul class="panel-list">
            <li v-for="item in recommendedScope.includes" :key="item">
              {{ item }}
            </li>
          </ul>
        </div>

        <div class="scope-card scope-card--safe">
          <div class="scope-card__title">
            <span class="scope-card__dot scope-card__dot--safe" />
            <span>本次会保留</span>
          </div>
          <div class="scope-card__desc">
            清空后依旧可继续沿用的基础资料与配置项。
          </div>
          <ul class="panel-list">
            <li v-for="item in recommendedScope.excludes" :key="item">
              {{ item }}
            </li>
          </ul>
        </div>
      </div>

      <a-checkbox v-model:checked="acknowledged" class="risk-check">
        <span class="risk-check__text">我已知晓清空后无法恢复，且已确认保留项符合当前校区使用需求</span>
      </a-checkbox>

      <div class="panel-actions">
        <a-button type="primary" :loading="submitting" @click="handleClearClick">
          确认清空
        </a-button>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.campus-data-clear {
  padding: 12px 0 0;
  background-color: #f6f7f8;
  min-height: 100%;
}

.campus-data-clear--embedded {
  padding: 0;
  background: transparent;
}

.settings-panel {
  background: #fff;
  border-radius: 16px;
  padding: 24px;
  border: 1px solid #edf0f5;
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 15px;
  font-weight: 600;
  color: #1f2329;
  margin-bottom: 18px;
}

.panel-title__marker {
  width: 6px;
  height: 24px;
  border-radius: 999px;
  background: var(--pro-ant-color-primary, #1677ff);
  flex-shrink: 0;
}

.warning-banner {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 18px 20px;
  border-radius: 14px;
  background: #fff7e8;
  border: 1px solid #ffe7ba;
  color: #ff8f1f;
  font-size: 16px;
  font-weight: 500;
}

.warning-banner__icon {
  font-size: 20px;
  flex-shrink: 0;
}

.summary-text {
  margin-top: 22px;
  font-size: 15px;
  line-height: 1.7;
  color: rgba(0, 0, 0, 0.78);
}

.scope-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 16px;
  margin-top: 22px;
}

.scope-card {
  border: 1px solid #ffe0b0;
  background: linear-gradient(180deg, #fffaf1 0%, #fff 100%);
  border-radius: 14px;
  padding: 18px;
}

.scope-card--safe {
  border-color: #dbeec7;
  background: linear-gradient(180deg, #fbfff7 0%, #fff 100%);
}

.scope-card__title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
}

.scope-card__dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #ff9c1a;
  flex-shrink: 0;
}

.scope-card__dot--safe {
  background: #53c41a;
}

.scope-card__desc {
  margin-top: 10px;
  font-size: 13px;
  line-height: 1.7;
  color: rgba(0, 0, 0, 0.56);
}

.panel-list {
  margin: 14px 0 0;
  padding-left: 20px;
  color: rgba(0, 0, 0, 0.68);
  font-size: 14px;
  display: grid;
  gap: 10px;
}

/* 勿用 display:block：会破坏 ant-checkbox-wrapper 内联布局，导致勾选框与文案上下错位 */
.risk-check {
  margin-top: 20px;
  width: 100%;
}

.risk-check__text {
  line-height: 1.75;
  white-space: normal;
}

.panel-actions {
  margin-top: 24px;
}

.panel-actions :deep(.ant-btn) {
  min-width: 144px;
  height: 44px;
  border-radius: 12px;
  font-size: 16px;
}
</style>
