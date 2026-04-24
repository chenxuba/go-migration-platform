<script setup lang="ts">
import { Empty, Modal } from 'ant-design-vue'
import type { TableColumnType } from 'ant-design-vue'
import dayjs from 'dayjs'
import type { Dayjs } from 'dayjs'
import { computed, h, onMounted, reactive, ref } from 'vue'
import { type InstConfig, getInstConfigModuleApi, setInstConfigModuleApi } from '~@/api/common/config'
import {
  deleteSchoolHolidayApi,
  listSchoolHolidaysApi,
  resetSchoolHolidaysApi,
  saveSchoolHolidayApi,
  type SchoolHolidayItem,
} from '@/api/business-settings/school-holiday'
import messageService from '~@/utils/messageService'
import TimePeriodSettings from '@/components/business-settings/timePeriodSettings.vue'

interface HolidayFormState {
  id?: number
  source: 'statutory' | 'custom'
  name: string
  startDate: Dayjs | null
  endDate: Dayjs | null
}

type SchoolHolidayRowLike = SchoolHolidayItem | Record<string, any>

const moduleConfig = ref<Partial<InstConfig>>({})
const activeKey = ref('holiday')
const periodGroupCount = ref(0)
const periodSettingsRef = ref<any>(null)
const rowLoadingMap = ref<Record<string, boolean>>({})
const holidayLoading = ref(false)
const holidaySubmitting = ref(false)
const holidayFormOpen = ref(false)
const holidayRows = ref<SchoolHolidayItem[]>([])

const emptyImage = Empty.PRESENTED_IMAGE_SIMPLE
const instConfig = computed<Partial<InstConfig>>(() => moduleConfig.value)
const holidayForm = reactive<HolidayFormState>({
  source: 'custom',
  name: '',
  startDate: null,
  endDate: null,
})

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

function normalizeTeacherSelectionRange(value: unknown) {
  return value === 'all' ? 'all' : 'teacher-only'
}

const holidayEnabled = computed(() => isConfigEnabled(instConfig.value.enableFilterHoliday, false))
const oneToOneLimitEnabled = computed(() => isConfigEnabled(instConfig.value.enableOneToOneScheduleLimit, false))
const allowCampusConflict = computed(() => isConfigEnabled(instConfig.value.enableScheduleConflictContinue, false))
const teacherRange = computed(() => normalizeTeacherSelectionRange(instConfig.value.scheduleTeacherSelectionRange))

const holidayColumns: TableColumnType<SchoolHolidayItem>[] = [
  { title: '类型', dataIndex: 'source', key: 'source', width: 110, align: 'center' },
  { title: '节假日名称', dataIndex: 'name', key: 'name', width: 220, ellipsis: true },
  { title: '开始日期', dataIndex: 'startDate', key: 'startDate', width: 170, align: 'center' },
  { title: '结束日期', dataIndex: 'endDate', key: 'endDate', width: 170, align: 'center' },
  { title: '操作', key: 'action', width: 120, align: 'center', fixed: 'right' },
]

const emptyLocale = {
  emptyText: h(Empty, { image: emptyImage, description: '暂无数据' }),
}

function isRowLoading(key: string) {
  return Boolean(rowLoadingMap.value[key])
}

async function loadCourseConfig() {
  const res = await getInstConfigModuleApi('course')
  moduleConfig.value = res.result || {}
}

async function updateConfigField(field: keyof InstConfig, value: InstConfig[keyof InstConfig], key: string, successText: string) {
  rowLoadingMap.value = {
    ...rowLoadingMap.value,
    [key]: true,
  }
  try {
    const payload = { [field]: value } as Partial<InstConfig>
    await setInstConfigModuleApi('course', payload)
    moduleConfig.value = { ...moduleConfig.value, ...payload }
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

async function handleHolidayToggle(checked: boolean) {
  await updateConfigField('enableFilterHoliday', checked, 'holidayEnabled', checked ? '已开启节假日设置' : '已关闭节假日设置')
}

async function handleOneToOneLimitToggle(checked: boolean) {
  await updateConfigField('enableOneToOneScheduleLimit', checked, 'oneToOneLimit', checked ? '已开启1对1排课数量限制' : '已关闭1对1排课数量限制')
}

async function handleConflictToggle(checked: boolean) {
  await updateConfigField('enableScheduleConflictContinue', checked, 'allowCampusConflict', checked ? '已允许校内冲突继续排课' : '已关闭校内冲突继续排课')
}

async function handleTeacherRangeChange(value: string) {
  await updateConfigField('scheduleTeacherSelectionRange', normalizeTeacherSelectionRange(value), 'teacherRange', '已保存上课教师选择范围')
}

function resetHolidayForm() {
  holidayForm.id = undefined
  holidayForm.source = 'custom'
  holidayForm.name = ''
  holidayForm.startDate = null
  holidayForm.endDate = null
}

function openCreateHolidayModal() {
  resetHolidayForm()
  holidayFormOpen.value = true
}

function openEditHolidayModal(record: SchoolHolidayRowLike) {
  holidayForm.id = Number(record.id)
  holidayForm.source = record.source === 'statutory' ? 'statutory' : 'custom'
  holidayForm.name = String(record.name || '')
  holidayForm.startDate = dayjs(String(record.startDate || ''))
  holidayForm.endDate = dayjs(String(record.endDate || ''))
  holidayFormOpen.value = true
}

async function loadSchoolHolidays() {
  holidayLoading.value = true
  try {
    const res = await listSchoolHolidaysApi()
    holidayRows.value = Array.isArray(res.result) ? res.result : []
  }
  catch (error) {
    console.error('load school holidays failed', error)
    messageService.error('获取节假日配置失败')
  }
  finally {
    holidayLoading.value = false
  }
}

async function submitHolidayForm() {
  const name = holidayForm.name.trim()
  if (!name) {
    messageService.error('请输入节假日名称')
    return
  }
  if (!holidayForm.startDate || !holidayForm.endDate) {
    messageService.error('请选择开始和结束日期')
    return
  }
  if (holidayForm.endDate.isBefore(holidayForm.startDate, 'day')) {
    messageService.error('结束日期不能早于开始日期')
    return
  }

  holidaySubmitting.value = true
  try {
    const res = await saveSchoolHolidayApi({
      id: holidayForm.id,
      name,
      source: holidayForm.source,
      startDate: holidayForm.startDate.format('YYYY-MM-DD'),
      endDate: holidayForm.endDate.format('YYYY-MM-DD'),
    })
    if (res.code !== 200) {
      messageService.error(res.message || '保存节假日失败')
      return
    }
    messageService.success(holidayForm.id ? '节假日更新成功' : '节假日新增成功')
    holidayFormOpen.value = false
    resetHolidayForm()
    await loadSchoolHolidays()
  }
  catch (error) {
    console.error('save school holiday failed', error)
    messageService.error('保存节假日失败')
  }
  finally {
    holidaySubmitting.value = false
  }
}

function handleDeleteHoliday(record: SchoolHolidayRowLike) {
  const holidayName = String(record.name || '')
  const holidayID = Number(record.id)
  Modal.confirm({
    title: '删除节假日',
    centered: true,
    content: `删除后将不再按“${holidayName}”过滤排课日期，是否继续？`,
    async onOk() {
      try {
        const res = await deleteSchoolHolidayApi({ id: holidayID })
        if (res.code !== 200) {
          messageService.error(res.message || '删除节假日失败')
          return
        }
        messageService.success('节假日已删除')
        await loadSchoolHolidays()
      }
      catch (error) {
        console.error('delete school holiday failed', error)
        messageService.error('删除节假日失败')
      }
    },
  })
}

function handleResetStatutoryHolidays() {
  Modal.confirm({
    title: '恢复法定假日',
    centered: true,
    content: '将用国家法定假日重新覆盖当前节假日配置，自定义新增内容会被清空，是否继续？',
    async onOk() {
      try {
        const res = await resetSchoolHolidaysApi()
        if (res.code !== 200) {
          messageService.error(res.message || '恢复法定假日失败')
          return
        }
        messageService.success('已恢复法定假日')
        await loadSchoolHolidays()
      }
      catch (error) {
        console.error('reset school holidays failed', error)
        messageService.error('恢复法定假日失败')
      }
    },
  })
}

onMounted(async () => {
  await loadCourseConfig()
  await loadSchoolHolidays()
})
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
                    <a-switch :checked="holidayEnabled" :loading="isRowLoading('holidayEnabled')" @change="handleHolidayToggle" />
                    <span class="settings-title-control__desc">开启后，系统自动跳过节假日安排日程，修改模板不影响已排课日程。</span>
                  </div>
                </div>
              </template>
              <template #right>
                <div class="schedule-title-actions">
                  <a-button @click="handleResetStatutoryHolidays">
                    恢复法定假日
                  </a-button>
                  <a-button type="primary" @click="openCreateHolidayModal">
                    添加节假日
                  </a-button>
                </div>
              </template>
            </custom-title>

            <div class="settings-subtitle">
              <span class="settings-subtitle__accent" />
              共设置 {{ holidayRows.length }} 条节假日
            </div>

            <a-table
              class="settings-data-table"
              :columns="holidayColumns"
              :data-source="holidayRows"
              :loading="holidayLoading"
              :pagination="false"
              row-key="id"
              :scroll="{ x: 790 }"
              :locale="emptyLocale"
            >
              <template #bodyCell="{ column, record, text }">
                <template v-if="column.key === 'source'">
                  <span class="holiday-type" :class="record.source === 'statutory' ? 'holiday-type--statutory' : 'holiday-type--custom'">
                    {{ record.source === 'statutory' ? '国家法定假日' : '自定义' }}
                  </span>
                </template>
                <template v-else-if="column.key === 'action'">
                  <a-button type="link" size="small" class="settings-link" @click="openEditHolidayModal(record)">
                    编辑
                  </a-button>
                  <a-button type="link" size="small" danger class="settings-link" @click="handleDeleteHoliday(record)">
                    删除
                  </a-button>
                </template>
                <template v-else>
                  {{ text }}
                </template>
              </template>
            </a-table>
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
                    <a-switch :checked="oneToOneLimitEnabled" :loading="isRowLoading('oneToOneLimit')" @change="handleOneToOneLimitToggle" />
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
                  <a-switch :checked="allowCampusConflict" :loading="isRowLoading('allowCampusConflict')" @change="handleConflictToggle" />
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
            <div class="settings-table">
              <div class="settings-row">
                <div class="settings-row__label settings-row__label--fixed">
                  上课教师选择范围
                </div>
                <div class="settings-row__content">
                  <a-radio-group
                    :value="teacherRange"
                    :disabled="isRowLoading('teacherRange')"
                    class="settings-radio-group custom-radio"
                    @change="handleTeacherRangeChange($event.target.value)"
                  >
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
              </div>
            </div>
          </div>
        </section>
      </a-tab-pane>
    </a-tabs>

    <a-modal
      v-model:open="holidayFormOpen"
      :title="holidayForm.id ? '编辑节假日' : '添加节假日'"
      :mask-closable="false"
      :confirm-loading="holidaySubmitting"
      destroy-on-close
      @ok="submitHolidayForm"
      @cancel="resetHolidayForm"
    >
      <a-form layout="vertical">
        <a-form-item label="节假日名称" required>
          <a-input v-model:value="holidayForm.name" placeholder="请输入节假日名称" :maxlength="30" />
        </a-form-item>
        <a-form-item label="开始日期" required>
          <a-date-picker v-model:value="holidayForm.startDate" style="width: 100%;" />
        </a-form-item>
        <a-form-item label="结束日期" required>
          <a-date-picker v-model:value="holidayForm.endDate" style="width: 100%;" />
        </a-form-item>
      </a-form>
    </a-modal>
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

.settings-row__label--fixed {
  width: 280px;
  min-width: 280px;
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

.holiday-type {
  display: inline-flex;
  align-items: center;
  height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 24px;
}

.holiday-type--statutory {
  color: #1668dc;
  background: #e6f4ff;
}

.holiday-type--custom {
  color: #6b7280;
  background: #f3f4f6;
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
