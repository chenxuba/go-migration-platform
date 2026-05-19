<script setup>
import dayjs from 'dayjs'
import messageService from '@/utils/messageService'
import {
  getPEP3AssessmentRecordDetailApi,
  updatePEP3AssessmentRecordConfigApi,
} from '@/api/edu-center/pep3-assessment'
import {
  getERXinAssessmentRecordDetailApi,
  updateERXinAssessmentRecordConfigApi,
} from '@/api/edu-center/erxin-assessment'
import {
  getAutismDevAssessmentRecordDetailApi,
  updateAutismDevAssessmentRecordConfigApi,
} from '@/api/edu-center/autismdev-assessment'
import {
  getShuangxiAAssessmentRecordDetailApi,
  updateShuangxiAAssessmentRecordConfigApi,
} from '@/api/edu-center/shuangxi-assessment'
import { getUserListApi } from '@/api/internal-manage/staff-manage'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  record: {
    type: Object,
    default: null,
  },
})

const emit = defineEmits(['update:open', 'saved'])

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const submitting = ref(false)
const detailLoading = ref(false)
const teacherOptions = ref([])
const teacherLoading = ref(false)
let teacherSearchTimer = 0

const form = reactive({
  examinerNames: [],
  assessmentDate: undefined,
})

const original = reactive({
  examinerName: '',
  assessmentDate: '',
})

function unwrap(res) {
  return res?.data ?? res?.result ?? res
}

function getErrorMessage(error, fallback) {
  return error?.response?.data?.message || error?.message || fallback
}

function normalizeDateValue(value) {
  if (!value)
    return ''
  const parsed = dayjs(value)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD') : ''
}

function isERXinRecord(record) {
  const source = String(record?._recordSource || record?.assessmentCode || '').trim().toUpperCase()
  return source === 'ERXIN' || source.startsWith('ERXIN')
}

function isAutismDevRecord(record) {
  const source = String(record?._recordSource || record?.assessmentCode || '').trim().toUpperCase()
  return source === 'AUTISMDEV'
}

function isShuangxiARecord(record) {
  const source = String(record?._recordSource || record?.assessmentCode || '').trim().toUpperCase()
  return source === 'SHUANGXI_A' || source === 'SHUANGXIA' || source.startsWith('SHUANGXI')
}

function recordSourceType(record) {
  if (isShuangxiARecord(record))
    return 'SHUANGXI_A'
  if (isAutismDevRecord(record))
    return 'AUTISMDEV'
  if (isERXinRecord(record))
    return 'ERXIN'
  return 'PEP3'
}

function splitExaminerNames(value) {
  return String(value || '')
    .split(/[、,，]/)
    .map(item => item.trim())
    .filter(Boolean)
}

function uniqueExaminerNames(names) {
  const seen = new Set()
  return (Array.isArray(names) ? names : [])
    .map(item => String(item || '').trim())
    .filter((name) => {
      if (!name || seen.has(name))
        return false
      seen.add(name)
      return true
    })
}

function joinExaminerNames(names) {
  return uniqueExaminerNames(names).join('、')
}

function getInputField(input, key) {
  const value = input?.[key]
  if (value === undefined || value === null)
    return ''
  return String(value).trim()
}

function ensureTeacherOptions(names = []) {
  const merged = new Map(teacherOptions.value.map(item => [String(item.value), item]))
  uniqueExaminerNames(names).forEach((name) => {
    if (!merged.has(name))
      merged.set(name, { label: name, value: name })
  })
  teacherOptions.value = Array.from(merged.values())
}

function normalizeTeacherOptions(res, extraNames = []) {
  const rows = Array.isArray(res?.result)
    ? res.result
    : Array.isArray(res?.data)
      ? res.data
      : []
  const merged = new Map()
  rows.forEach((row) => {
    const name = String(row?.nickName || row?.name || '').trim()
    if (name && !merged.has(name)) {
      merged.set(name, {
        label: name,
        value: name,
      })
    }
  })
  uniqueExaminerNames(extraNames).forEach((name) => {
    if (!merged.has(name))
      merged.set(name, { label: name, value: name })
  })
  return Array.from(merged.values())
}

async function fetchTeacherOptions(searchKey = '') {
  teacherLoading.value = true
  try {
    const keepNames = [
      ...form.examinerNames,
      ...splitExaminerNames(original.examinerName),
    ]
    const res = await getUserListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 50,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        status: 0,
        searchKey: String(searchKey || '').trim(),
      },
    })
    teacherOptions.value = normalizeTeacherOptions(res, keepNames)
  }
  catch (error) {
    messageService.error(getErrorMessage(error, '获取评估老师失败'))
    ensureTeacherOptions([...form.examinerNames, ...splitExaminerNames(original.examinerName)])
  }
  finally {
    teacherLoading.value = false
  }
}

function handleTeacherSearch(value) {
  if (teacherSearchTimer)
    window.clearTimeout(teacherSearchTimer)
  teacherSearchTimer = window.setTimeout(() => {
    fetchTeacherOptions(value)
    teacherSearchTimer = 0
  }, 260)
}

function handleTeacherDropdownVisibleChange(visible) {
  if (visible)
    fetchTeacherOptions()
}

async function initializeConfig() {
  const row = props.record
  if (!row?.id)
    return
  original.examinerName = ''
  original.assessmentDate = ''
  form.examinerNames = uniqueExaminerNames(splitExaminerNames(row.examinerName))
  form.assessmentDate = normalizeDateValue(row.assessmentDate) ? dayjs(row.assessmentDate) : undefined
  ensureTeacherOptions(form.examinerNames)
  fetchTeacherOptions()

  const recordId = row.id
  const recordSource = recordSourceType(row)
  detailLoading.value = true
  try {
    const detail = unwrap(await (recordSource === 'AUTISMDEV'
      ? getAutismDevAssessmentRecordDetailApi(recordId)
      : recordSource === 'SHUANGXI_A'
        ? getShuangxiAAssessmentRecordDetailApi(recordId)
      : recordSource === 'ERXIN'
        ? getERXinAssessmentRecordDetailApi(recordId)
        : getPEP3AssessmentRecordDetailApi(recordId)))
    if (props.record?.id !== recordId || !props.open || recordSourceType(props.record) !== recordSource)
      return
    const originalExaminerName = getInputField(detail?.input, 'examinerName') || row.examinerName || ''
    const originalAssessmentDate = normalizeDateValue(getInputField(detail?.input, 'assessmentDate')) || normalizeDateValue(row.assessmentDate)
    original.examinerName = originalExaminerName
    original.assessmentDate = originalAssessmentDate
    ensureTeacherOptions([
      ...form.examinerNames,
      ...splitExaminerNames(originalExaminerName),
    ])
  }
  catch (error) {
    messageService.error(getErrorMessage(error, '获取原评估信息失败'))
    original.examinerName = row.examinerName || ''
    original.assessmentDate = normalizeDateValue(row.assessmentDate)
  }
  finally {
    if (props.record?.id === recordId)
      detailLoading.value = false
  }
}

function closeModal() {
  if (submitting.value)
    return
  modalOpen.value = false
}

function restoreOriginalExaminer() {
  const names = splitExaminerNames(original.examinerName)
  if (!names.length) {
    messageService.warning('暂无原评估老师可恢复')
    return
  }
  form.examinerNames = uniqueExaminerNames(names)
  ensureTeacherOptions(form.examinerNames)
}

function restoreOriginalAssessmentDate() {
  const originalDate = normalizeDateValue(original.assessmentDate)
  if (!originalDate) {
    messageService.warning('暂无原评估日期可恢复')
    return
  }
  form.assessmentDate = dayjs(originalDate)
}

async function saveConfig() {
  const row = props.record
  if (!row?.id)
    return
  const examinerName = joinExaminerNames(form.examinerNames)
  if (!examinerName) {
    messageService.warning('请选择评估老师')
    return
  }
  if (!form.assessmentDate || !dayjs(form.assessmentDate).isValid()) {
    messageService.warning('请选择评估日期')
    return
  }
  submitting.value = true
  try {
    const payload = {
      id: row.id,
      examinerName,
      assessmentDate: dayjs(form.assessmentDate).format('YYYY-MM-DD'),
    }
    if (isAutismDevRecord(row))
      await updateAutismDevAssessmentRecordConfigApi(payload)
    else if (isShuangxiARecord(row))
      await updateShuangxiAAssessmentRecordConfigApi(payload)
    else if (isERXinRecord(row))
      await updateERXinAssessmentRecordConfigApi(payload)
    else
      await updatePEP3AssessmentRecordConfigApi(payload)
    messageService.success('评估配置已保存')
    modalOpen.value = false
    emit('saved')
  }
  catch (error) {
    messageService.error(getErrorMessage(error, '保存评估配置失败'))
  }
  finally {
    submitting.value = false
  }
}

watch(
  () => [props.open, props.record?.id],
  ([open]) => {
    if (open)
      initializeConfig()
  },
)

onBeforeUnmount(() => {
  if (teacherSearchTimer) {
    window.clearTimeout(teacherSearchTimer)
    teacherSearchTimer = 0
  }
})
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    width="520px"
    :centered="true"
    :confirm-loading="submitting"
    :mask-closable="!submitting"
    ok-text="保存"
    cancel-text="取消"
    wrap-class-name="assessment-record-config-modal"
    @ok="saveConfig"
    @cancel="closeModal"
  >
    <template #title>
      <div class="config-modal-title">
        <span>配置评估记录</span>
        <small>{{ record?.studentName || '-' }} / {{ record?.assessmentName || '评估记录' }}</small>
      </div>
    </template>
    <a-spin :spinning="detailLoading">
      <div class="config-form">
        <a-alert
          class="config-alert"
          type="info"
          show-icon
          message="仅同步评估老师、评估日期等表头信息；IEP目标内容不重新生成、不改动。已生成IEP的计划参与者/实施者会随保存同步。"
        />
        <a-form layout="vertical">
          <a-form-item label="评估老师" required>
            <a-select
              v-model:value="form.examinerNames"
              mode="multiple"
              show-search
              allow-clear
              :filter-option="false"
              :loading="teacherLoading"
              :options="teacherOptions"
              :max-tag-count="2"
              placeholder="请选择评估老师"
              @search="handleTeacherSearch"
              @dropdown-visible-change="handleTeacherDropdownVisibleChange"
            />
            <div class="config-original-line">
              <span>原评估老师：{{ original.examinerName || '-' }}</span>
              <a-button type="link" size="small" :disabled="!original.examinerName" @click="restoreOriginalExaminer">
                恢复原评估老师
              </a-button>
            </div>
          </a-form-item>

          <a-form-item label="评估日期" required>
            <a-date-picker
              v-model:value="form.assessmentDate"
              class="config-date-picker"
              :allow-clear="false"
              placeholder="请选择评估日期"
            />
            <div class="config-original-line">
              <span>原评估日期：{{ original.assessmentDate || '-' }}</span>
              <a-button type="link" size="small" :disabled="!original.assessmentDate" @click="restoreOriginalAssessmentDate">
                恢复原评估日期
              </a-button>
            </div>
          </a-form-item>
        </a-form>
      </div>
    </a-spin>
  </a-modal>
</template>

<style lang="less" scoped>
.config-modal-title {
  display: flex;
  flex-direction: column;
  gap: 2px;

  span {
    color: #1f2937;
    font-size: 18px;
    font-weight: 600;
    line-height: 26px;
  }

  small {
    color: #8a94a6;
    font-size: 12px;
    font-weight: 400;
    line-height: 18px;
  }
}

.config-form {
  padding: 10px 0 2px;
}

.config-alert {
  margin-bottom: 16px;
}

.config-date-picker {
  width: 100%;
}

.config-original-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-height: 24px;
  margin-top: 6px;
  color: #7a8494;
  font-size: 12px;
  line-height: 20px;

  span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  :deep(.ant-btn-link) {
    flex: 0 0 auto;
    height: 24px;
    padding: 0;
    font-size: 12px;
  }
}
</style>
