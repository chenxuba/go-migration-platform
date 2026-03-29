<script setup>
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { computed, createVNode, ref } from 'vue'
import { CAMPUS_DATA_CLEAR_SCOPE_BUSINESS_ONLY, clearCampusDataApi } from '@/api/business-settings/campus-data-clear'
import messageService from '@/utils/messageService'

const acknowledged = ref(false)
const submitting = ref(false)

const recommendedScope = {
  value: CAMPUS_DATA_CLEAR_SCOPE_BUSINESS_ONLY,
  title: '推荐范围：只清业务数据',
  description: '适合机构重新起盘、演示环境回收、历史业务重置，不影响基础配置继续使用。',
  includes: [
    '学员主档数据',
    '学员自定义字段值与学员变更记录',
    '跟进记录',
    '课程、课程详情、课程报价、课程属性结果',
    '套餐、套餐内商品、套餐属性结果',
    '订单、订单明细、支付记录',
    '审批记录与审批历史',
    '学费账户',
    '意向学员导入记录',
    '课程销量统计归零',
    '班级与 1 对 1（含班员、教师关联）',
  ],
  excludes: [
    '员工、角色、部门',
    '机构业务设置',
    '渠道与渠道分类',
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
    return result?.intentStudentIndexMessage || '机构业务数据已清空'

  const summary = `已清空学员 ${cleared.students || 0} 条、课程 ${cleared.courses || 0} 条、套餐 ${cleared.productPackages || 0} 条、订单 ${cleared.orders || 0} 条、1对1/班级 ${cleared.teachingClasses || 0} 条`
  return result?.intentStudentIndexMessage ? `${summary}，${result.intentStudentIndexMessage}` : summary
}

function handleClearClick() {
  if (!acknowledged.value) {
    messageService.warning('请先确认已知晓清空风险与保留范围')
    return
  }

  Modal.confirm({
    title: '确认清空当前机构业务数据？',
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
  <div class="campus-data-clear">
    <a-alert
      type="warning"
      show-icon
      class="mb-16px"
      message="操作前请仔细阅读"
      description="机构数据清空将删除或重置当前机构下的业务数据，该操作通常不可恢复。请确认已由管理员授权后再执行。"
    >
      <template #icon>
        <ExclamationCircleOutlined />
      </template>
    </a-alert>
    <div class="settings-panel">
      <div class="panel-title">
        清空数据范围建议
      </div>
      <div class="scope-card">
        <div class="scope-card__header">
          <div>
            <div class="scope-card__title">
              {{ recommendedScope.title }}
            </div>
            <div class="scope-card__desc">
              {{ recommendedScope.description }}
            </div>
          </div>
          <a-tag color="orange">
            风险最低
          </a-tag>
        </div>

        <div class="scope-section">
          <div class="scope-section__title">
            本次会清空
          </div>
          <ul class="panel-list">
            <li v-for="item in recommendedScope.includes" :key="item">
              {{ item }}
            </li>
          </ul>
        </div>

        <div class="scope-section scope-section--safe">
          <div class="scope-section__title">
            本次会保留
          </div>
          <ul class="panel-list">
            <li v-for="item in recommendedScope.excludes" :key="item">
              {{ item }}
            </li>
          </ul>
        </div>
      </div>

      <a-checkbox v-model:checked="acknowledged" class="risk-check">
        <span class="risk-check__text">我已确认本次仅清空业务数据，且已获得管理员授权</span>
      </a-checkbox>

      <div class="panel-actions">
        <a-button type="primary" danger block :loading="submitting" @click="handleClearClick">
          确认清空
        </a-button>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.campus-data-clear {
  padding: 12px 16px;
  background-color: #f6f7f8;
  min-height: 100%;
}

.settings-panel {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  border: 1px solid #f0f0f0;
}

.panel-title {
  font-size: 15px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.85);
  margin-bottom: 12px;
}

.scope-card {
  border: 1px solid #ffd591;
  background: linear-gradient(180deg, #fffaf0 0%, #fff 100%);
  border-radius: 10px;
  padding: 16px;
}

.scope-card__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.scope-card__title {
  font-size: 16px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.88);
}

.scope-card__desc {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.7;
  color: rgba(0, 0, 0, 0.65);
}

.scope-section {
  margin-top: 16px;
}

.scope-section--safe {
  padding-top: 16px;
  border-top: 1px dashed #f0f0f0;
}

.scope-section__title {
  font-size: 13px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.78);
  margin-bottom: 8px;
}

.panel-list {
  margin: 0;
  padding-left: 20px;
  color: rgba(0, 0, 0, 0.65);
  font-size: 14px;
}

/* 勿用 display:block：会破坏 ant-checkbox-wrapper 内联布局，导致勾选框与文案上下错位 */
.risk-check {
  margin-top: 16px;
  width: 100%;
}


.risk-check__text {
  flex: 1;
  min-width: 0;
  white-space: normal;
}

.panel-actions {
  margin-top: 20px;
  width: 100%;
}
</style>
