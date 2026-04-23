<script setup lang="ts">
import { Empty } from 'ant-design-vue'
import type { TableColumnType } from 'ant-design-vue'
import { h, reactive, ref } from 'vue'
import TimePeriodSettings from '@/components/business-settings/timePeriodSettings.vue'

type HolidayRow = {
  id: string
  period: string
  name: string
  startDate: string
  endDate: string
}

const activeKey = ref('holiday')
const teacherRange = ref('teacher-only')
const periodGroupCount = ref(0)
const periodSettingsRef = ref<any>(null)

const switches = reactive({
  holidayEnabled: true,
  oneToOneLimit: false,
  allowCampusConflict: true,
})

const emptyImage = Empty.PRESENTED_IMAGE_SIMPLE

const holidayColumns: TableColumnType<HolidayRow>[] = [
  { title: '时段', dataIndex: 'period', key: 'period', width: 140 },
  { title: '节假日时段名称', dataIndex: 'name', key: 'name' },
  { title: '开始日期', dataIndex: 'startDate', key: 'startDate', width: 180 },
  { title: '结束日期', dataIndex: 'endDate', key: 'endDate', width: 180 },
]

const holidayRows: HolidayRow[] = []

const emptyLocale = {
  emptyText: h(Empty, { image: emptyImage, description: '暂无数据' }),
}
</script>

<template>
  <div class="schedule-settings">
    <a-tabs v-model:active-key="activeKey" destroy-inactive-tab-pane class="schedule-settings__tabs">
      <a-tab-pane key="holiday" tab="节假日设置">
        <section class="tab-content">
          <div class="setting">
            <custom-title font-size="18px" font-weight="800" before-height="14px">
              <template #left>
                <div class="settings-title-inline">
                  <span>节假日设置</span>
                  <div class="settings-title-control">
                    <a-switch v-model:checked="switches.holidayEnabled" />
                    <span class="settings-title-control__desc">开启后，系统自动跳过节假日安排日程，修改模板不影响已排课日程。</span>
                  </div>
                </div>
              </template>
            </custom-title>

            <div class="settings-subtitle">
              <span class="settings-subtitle__accent" />
              共设置 0 条节假日
            </div>

            <a-table
              class="settings-data-table"
              :columns="holidayColumns"
              :data-source="holidayRows"
              :pagination="false"
              row-key="id"
              :locale="emptyLocale"
            />
          </div>
        </section>
      </a-tab-pane>

      <a-tab-pane key="period" tab="上课时段设置">
        <section class="tab-content">
          <div class="setting">
            <custom-title font-size="18px" font-weight="800" before-height="14px">
              <template #left>
                <div class="schedule-title-inline">
                  <span>上课时段设置</span>
                  <span class="schedule-title-inline__meta">当前共计 {{ periodGroupCount }} 个时段组（逐行编辑，可随时添加）</span>
                </div>
              </template>
              <template #right>
                <div class="schedule-title-actions">
                  <a-button :loading="Boolean(periodSettingsRef?.repairing)" @click="periodSettingsRef?.repairPeriodVersions?.()">
                    一键修复
                  </a-button>
                  <a-button type="primary" @click="periodSettingsRef?.openCreateGroup?.()">
                    添加时段组
                  </a-button>
                </div>
              </template>
            </custom-title>
            <div class="schedule-settings__period-pane">
              <TimePeriodSettings ref="periodSettingsRef" embedded @summary-change="periodGroupCount = $event" />
            </div>
          </div>
        </section>
      </a-tab-pane>

      <a-tab-pane key="one-to-one-limit" tab="1对1排课数量限制">
        <section class="tab-content">
          <div class="setting">
            <custom-title font-size="18px" font-weight="800" before-height="14px">
              <template #left>
                <div class="settings-title-inline">
                  <span>1对1排课数量限制</span>
                  <div class="settings-title-control">
                    <a-switch v-model:checked="switches.oneToOneLimit" />
                    <span class="settings-title-control__desc">开启后，如果 1 对 1 排课预计花费课时或金额大于学员剩余数量，将无法排课；按时段收费的课程不受限制。</span>
                  </div>
                </div>
              </template>
            </custom-title>
          </div>
        </section>
      </a-tab-pane>

      <a-tab-pane key="conflict" tab="课程冲突设置">
        <section class="tab-content">
          <div class="setting">
            <custom-title title="课程冲突设置" font-size="18px" font-weight="800" before-height="14px" />
            <div class="settings-table">
              <div class="settings-row">
                <div class="settings-row__label">
                  校内冲突是否允许仍然排课
                </div>
                <div class="settings-row__content">
                  <a-switch v-model:checked="switches.allowCampusConflict" />
                  <div class="settings-desc">
                    开启后，本校区的冲突日程可以选择仍然继续排课。
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </a-tab-pane>

      <a-tab-pane key="teacher-range" tab="上课教师选择范围">
        <section class="tab-content">
          <div class="setting">
            <custom-title title="上课教师选择范围" font-size="18px" font-weight="800" before-height="14px" />
            <a-radio-group v-model:value="teacherRange" class="settings-radio-group custom-radio">
              <a-radio value="all">
                全部员工
              </a-radio>
              <a-radio value="teacher-only">
                仅教师
              </a-radio>
            </a-radio-group>

            <div class="settings-desc">
              排课时上课教师仅可以选择有教师身份的员工，可前往员工信息中编辑教师身份。
            </div>
          </div>
        </section>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<style scoped lang="less">
.schedule-settings {
  min-height: 420px;
}

.schedule-settings__tabs {
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

.tab-content {
  margin-top: 10px;
  background: #fff;
  border-radius: 12px;
  padding: 18px 20px 12px;
}

.setting {
  min-height: 180px;
}

.schedule-settings__period-pane {
  margin-top: 8px;
  min-height: calc(100vh - 300px);
  overflow: hidden;
}

.schedule-title-inline {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.schedule-title-inline__meta {
  color: #8c8c8c;
  font-size: 13px;
  font-weight: 400;
  line-height: 22px;
  white-space: nowrap;
}

.schedule-title-actions {
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}

.settings-title-inline {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.settings-title-control {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  justify-content: flex-start;
  max-width: 820px;
}

.settings-title-control__desc {
  color: #666;
  font-size: 14px;
  font-weight: 400;
  line-height: 22px;
  text-align: left;
}

.settings-desc {
  margin-top: 8px;
  color: #333;
  font-size: 14px;
  line-height: 22px;
}

.setting :deep(.title) {
  margin-bottom: 12px;
}

.settings-subtitle {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 22px 0 14px;
  color: #1f2937;
  font-size: 14px;
  font-weight: 600;
}

.settings-subtitle__accent {
  width: 3px;
  height: 14px;
  border-radius: 3px;
  background: var(--pro-ant-color-primary, #1677ff);
}

.settings-table {
  overflow: hidden;
  border: 1px solid #edf0f5;
  background: #fff;
}

.settings-row {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  min-height: 84px;
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

.settings-data-table {
  :deep(.ant-table) {
    border: 1px solid #edf0f5;
    border-radius: 8px;
    overflow: hidden;
    background: #fff;
  }

  :deep(.ant-table-container) {
    border-inline-start: 0 !important;
  }

  :deep(.ant-table-thead > tr > th) {
    background: #fff;
    color: #1f2937;
    font-weight: 600;
  }

  :deep(.ant-table-thead > tr > th),
  :deep(.ant-table-tbody > tr > td) {
    padding: 14px 16px;
    border-color: #edf0f5;
    font-size: 14px;
  }

  :deep(.ant-table-placeholder .ant-table-cell) {
    padding: 42px 16px;
  }
}

.settings-link {
  padding: 0 4px;
  font-size: 14px;
}

@media (max-width: 768px) {
  .settings-title-inline {
    display: flex;
    align-items: flex-start;
  }

  .settings-title-control {
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-start;
  }

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
