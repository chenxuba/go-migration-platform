<script setup>
import { computed } from 'vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  forcing: {
    type: Boolean,
    default: false,
  },
  locating: {
    type: Boolean,
    default: false,
  },
  conflictDetailState: {
    type: Object,
    default: () => ({
      summary: '',
      attempted: null,
      items: [],
    }),
  },
})

const emit = defineEmits(['update:open', 'force', 'jump'])

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const summaryCount = computed(() => (
  Array.isArray(props.conflictDetailState?.items)
    ? props.conflictDetailState.items.length
    : 0
))

const summaryConflictTypes = computed(() => {
  const attemptedTypes = Array.isArray(props.conflictDetailState?.attempted?.conflictTypes)
    ? props.conflictDetailState.attempted.conflictTypes
    : []
  const itemTypes = Array.isArray(props.conflictDetailState?.items)
    ? props.conflictDetailState.items.flatMap(item => Array.isArray(item?.conflictTypes) ? item.conflictTypes : [])
    : []
  const normalized = [...attemptedTypes, ...itemTypes]
    .map(item => String(item || '').trim())
    .filter(Boolean)
  return Array.from(new Set(normalized))
})

const summaryTitle = computed(() => {
  if (summaryConflictTypes.value.length > 1)
    return `当前空位存在 ${summaryConflictTypes.value.join('、')} 冲突`
  if (summaryConflictTypes.value.length === 1)
    return `当前空位存在 ${summaryConflictTypes.value[0]} 冲突`
  return '当前空位存在冲突'
})

const summaryHint = computed(() => {
  if (props.conflictDetailState?.attempted?.forceAllowed)
    return '下方列出具体冲突课程，可先定位查看；纯学员冲突支持仍要排课。'
  return '下方列出具体冲突课程，可直接定位查看，再决定调整时间、老师或排课对象。'
})

function normalizeNameList(value) {
  if (Array.isArray(value))
    return value.map(item => String(item || '').trim()).filter(Boolean)
  const text = String(value || '').trim()
  if (!text || text === '未安排' || text === '-')
    return []
  return text.split(/[、,，]/).map(item => item.trim()).filter(Boolean)
}

function extractAssistantNames(source) {
  if (!source)
    return []
  if (Array.isArray(source.assistantNames) && source.assistantNames.length)
    return normalizeNameList(source.assistantNames)
  return normalizeNameList(source.assistantText)
}

function hasAssistantConflict(source) {
  if (!source)
    return false
  if (typeof source.hasAssistantConflict === 'boolean')
    return source.hasAssistantConflict
  return Array.isArray(source.conflictTypes) && source.conflictTypes.includes('助教')
}

const attemptedAssistantNames = computed(() => extractAssistantNames(props.conflictDetailState?.attempted))
const attemptedAssistantNameSet = computed(() => new Set(attemptedAssistantNames.value))

const conflictingAttemptedAssistantNames = computed(() => {
  const names = Array.isArray(props.conflictDetailState?.items)
    ? props.conflictDetailState.items
        .filter(item => hasAssistantConflict(item))
        .flatMap(item => extractAssistantNames(item).filter(name => attemptedAssistantNameSet.value.has(name)))
    : []
  return new Set(names)
})

function assistantNamesFor(source) {
  return extractAssistantNames(source)
}

function isAttemptedAssistantDanger(name) {
  return conflictingAttemptedAssistantNames.value.has(name)
}

function isConflictAssistantDanger(item, name) {
  return hasAssistantConflict(item) && attemptedAssistantNameSet.value.has(name)
}

function studentNamesFor(source) {
  if (!source)
    return []
  if (Array.isArray(source.studentNames) && source.studentNames.length)
    return normalizeNameList(source.studentNames)
  return normalizeNameList(source.studentText)
}

function conflictingStudentNameSetFor(source) {
  if (!source)
    return new Set()
  if (Array.isArray(source.conflictingStudentNames) && source.conflictingStudentNames.length)
    return new Set(normalizeNameList(source.conflictingStudentNames))
  return new Set()
}

function isConflictStudentDanger(item, name) {
  if (!item?.hasStudentConflict)
    return false
  const conflictingSet = conflictingStudentNameSetFor(item)
  if (!conflictingSet.size)
    return true
  return conflictingSet.has(name)
}
</script>

<template>
  <a-modal
    v-model:open="modalOpen"
    title="冲突详情"
    class="st-conflict-detail-modal"
    :footer="null"
    width="760px"
    centered
    :body-style="{ paddingTop: '0px' }"
  >
    <div class="st-conflict-modal">
      <div class="st-conflict-summary">
        <div class="st-conflict-summary__headline">
          <div class="st-conflict-summary__title">
            {{ summaryTitle }}
          </div>
          <div class="st-conflict-summary__count">
            共 {{ summaryCount }} 条冲突日程
          </div>
        </div>
        <div v-if="summaryConflictTypes.length" class="st-conflict-summary__chips">
          <span class="st-conflict-summary__chips-label">冲突类型</span>
          <span
            v-for="type in summaryConflictTypes"
            :key="type"
            class="st-conflict-summary__chip"
          >
            {{ type }}冲突
          </span>
        </div>
        <div class="st-conflict-summary__hint">
          {{ summaryHint }}
        </div>
      </div>

      <a-spin :spinning="locating" tip="定位中..." class="st-conflict-content-spin">
        <div class="st-conflict-content">
          <div v-if="conflictDetailState.attempted" class="st-conflict-attempt">
            <div class="st-conflict-section-title">
              你正在选择的空位
            </div>
            <div class="st-conflict-attempt__card">
              <div class="st-conflict-attempt__headline">
                <div class="st-conflict-attempt__headline-main">
                  <span class="st-conflict-attempt__badge">{{ conflictDetailState.attempted.modeLabel }}</span>
                  <span>待排课程信息</span>
                </div>
                <a-button
                  v-if="conflictDetailState.attempted?.forceAllowed"
                  type="primary"
                  ghost
                  danger
                  :loading="forcing"
                  @click="$emit('force')"
                >
                  仍要排课
                </a-button>
                <div
                  v-else-if="conflictDetailState.attempted?.forceDisabledReason"
                  class="st-conflict-attempt__force-tip"
                >
                  {{ conflictDetailState.attempted.forceDisabledReason }}
                </div>
              </div>
              <div class="st-conflict-attempt__meta st-conflict-attempt__meta--time">
                {{ conflictDetailState.attempted.date }} {{ conflictDetailState.attempted.week }}
                第{{ conflictDetailState.attempted.lessonIndex }}节
              </div>
              <div class="st-conflict-attempt__target">
                <div class="st-conflict-attempt__target-label">
                  <span>{{ conflictDetailState.attempted.targetLabel }}</span>
                </div>
                <strong class="st-conflict-attempt__target-value">{{ conflictDetailState.attempted.targetValue }}</strong>
              </div>
              <div class="st-conflict-attempt__facts">
                <div class="st-conflict-attempt__fact">
                  <span class="st-conflict-attempt__fact-label">上课课程</span>
                  <strong class="st-conflict-attempt__fact-value">{{ conflictDetailState.attempted.courseName }}</strong>
                </div>
                <div class="st-conflict-attempt__fact">
                  <span class="st-conflict-attempt__fact-label">上课时间</span>
                  <strong class="st-conflict-attempt__fact-value">{{ conflictDetailState.attempted.timeText }}</strong>
                </div>
                <div class="st-conflict-attempt__fact">
                  <span class="st-conflict-attempt__fact-label">上课老师</span>
                  <strong
                    class="st-conflict-attempt__fact-value"
                    :class="{ 'st-conflict-attempt__fact-value--danger': (conflictDetailState.attempted?.conflictTypes || []).includes('老师') }"
                  >
                    {{ conflictDetailState.attempted.teacherName }}
                  </strong>
                </div>
                <div class="st-conflict-attempt__fact">
                  <span class="st-conflict-attempt__fact-label">上课助教</span>
                  <strong
                    class="st-conflict-attempt__fact-value"
                  >
                    <template v-if="assistantNamesFor(conflictDetailState.attempted).length">
                      <template
                        v-for="(name, index) in assistantNamesFor(conflictDetailState.attempted)"
                        :key="`${name}-${index}`"
                      >
                        <span :class="{ 'st-conflict-attempt__fact-value--danger': isAttemptedAssistantDanger(name) }">{{ name }}</span>
                        <span v-if="index < assistantNamesFor(conflictDetailState.attempted).length - 1">、</span>
                      </template>
                    </template>
                    <template v-else>
                      未安排
                    </template>
                  </strong>
                </div>
                <div
                  v-if="conflictDetailState.attempted?.studentText"
                  class="st-conflict-attempt__fact"
                >
                  <span class="st-conflict-attempt__fact-label">{{ conflictDetailState.attempted.studentLabel || '上课学员' }}</span>
                  <strong class="st-conflict-attempt__fact-value">
                    {{ conflictDetailState.attempted.studentText }}
                  </strong>
                </div>
                <div
                  v-if="conflictDetailState.attempted?.conflictStudentText"
                  class="st-conflict-attempt__fact"
                >
                  <span class="st-conflict-attempt__fact-label">{{ conflictDetailState.attempted.conflictStudentLabel || '冲突学员' }}</span>
                  <strong
                    class="st-conflict-attempt__fact-value st-conflict-attempt__fact-value--danger"
                  >
                    {{ conflictDetailState.attempted.conflictStudentText }}
                  </strong>
                </div>
                <div
                  v-if="conflictDetailState.attempted?.modeLabel === '班课' || conflictDetailState.attempted?.classroomName || (conflictDetailState.attempted?.conflictTypes || []).includes('教室')"
                  class="st-conflict-attempt__fact"
                >
                  <span class="st-conflict-attempt__fact-label">上课教室</span>
                  <strong
                    class="st-conflict-attempt__fact-value"
                    :class="{ 'st-conflict-attempt__fact-value--danger': (conflictDetailState.attempted?.conflictTypes || []).includes('教室') }"
                  >
                    {{
                      conflictDetailState.attempted.classroomName
                        || '未设置教室'
                    }}
                  </strong>
                </div>
                <div class="st-conflict-attempt__fact">
                  <span class="st-conflict-attempt__fact-label">所在组别</span>
                  <strong class="st-conflict-attempt__fact-value">{{ conflictDetailState.attempted.groupLabel }}</strong>
                </div>
              </div>
              <div v-if="(conflictDetailState.attempted?.conflictTypes || []).length" class="st-conflict-attempt__reason-row">
                <span class="st-conflict-attempt__reason-label">当前冲突：</span>
                <span
                  v-for="type in conflictDetailState.attempted.conflictTypes || []"
                  :key="type"
                  class="st-conflict-attempt__reason-chip"
                >
                  {{ type }}冲突
                </span>
              </div>
              <div v-if="conflictDetailState.attempted?.warningText" class="st-conflict-attempt__warning">
                {{ conflictDetailState.attempted.warningText }}
              </div>
              <div class="st-conflict-attempt__meta">
                系统正在校验这条待排课信息与课表中的已有日程是否冲突。
              </div>
            </div>
          </div>

          <div class="st-conflict-section-title">
            冲突课程
          </div>
          <div class="st-conflict-list">
            <div v-for="item in conflictDetailState.items" :key="item.key" class="st-conflict-item">
              <div class="st-conflict-item__main">
                <div class="st-conflict-item__headline">
                  <span>{{ item.name }}</span>
                  <a-tag color="blue" :bordered="false">
                    {{ item.classTypeText }}
                  </a-tag>
                  <a-tag color="orange" :bordered="false">
                    {{ item.groupLabel }}
                  </a-tag>
                </div>
                <div class="st-conflict-item__meta">
                  {{ item.date }} {{ item.week }} · {{ item.timeText }}
                </div>
                <div class="st-conflict-item__meta">
                  教师：
                  <span :class="{ 'st-conflict-item__value--danger': item.hasTeacherConflict }">{{ item.teacherName }}</span>
                  <template v-if="item.assistantText && item.assistantText !== '-'">
                    <span class="st-conflict-item__sep">｜</span>
                    助教：
                    <span>
                      <template
                        v-for="(name, index) in assistantNamesFor(item)"
                        :key="`${item.key}-assistant-${name}-${index}`"
                      >
                        <span :class="{ 'st-conflict-item__value--danger': isConflictAssistantDanger(item, name) }">{{ name }}</span>
                        <span v-if="index < assistantNamesFor(item).length - 1">、</span>
                      </template>
                    </span>
                  </template>
                  <span class="st-conflict-item__sep">｜</span>
                  学员：
                  <span>
                    <template v-if="studentNamesFor(item).length">
                      <template
                        v-for="(name, index) in studentNamesFor(item)"
                        :key="`${item.key}-student-${name}-${index}`"
                      >
                        <span :class="{ 'st-conflict-item__value--danger': isConflictStudentDanger(item, name) }">{{ name }}</span>
                        <span v-if="index < studentNamesFor(item).length - 1">、</span>
                      </template>
                    </template>
                    <template v-else>
                      <span :class="{ 'st-conflict-item__value--danger': item.hasStudentConflict }">{{ item.studentText }}</span>
                    </template>
                  </span>
                  <template v-if="item.hasClassroomConflict || (item.classroomName && item.classroomName !== '-')">
                    <span class="st-conflict-item__sep">｜</span>
                    教室：
                    <span :class="{ 'st-conflict-item__value--danger': item.hasClassroomConflict }">{{ item.classroomName || '未设置教室' }}</span>
                  </template>
                </div>
                <div class="st-conflict-item__meta st-conflict-item__meta--reasons">
                  <span>冲突原因：</span>
                  <span v-if="!(item.conflictTypes || []).length" class="st-conflict-item__reason-chip st-conflict-item__reason-chip--danger">
                    时间冲突
                  </span>
                  <template v-else>
                    <span
                      v-for="type in item.conflictTypes || []"
                      :key="type"
                      class="st-conflict-item__reason-chip st-conflict-item__reason-chip--danger"
                    >
                      {{ type }}冲突
                    </span>
                  </template>
                </div>
              </div>
              <div class="st-conflict-item__side">
                <a-button
                  type="primary"
                  ghost
                  :disabled="!item.jumpCellKey"
                  @click="$emit('jump', item)"
                >
                  定位到课程
                </a-button>
              </div>
            </div>
          </div>
        </div>
      </a-spin>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
.st-conflict-modal {
  display: flex;
  flex-direction: column;
}

.st-conflict-content {
  display: flex;
  flex-direction: column;
}

:deep(.st-conflict-detail-modal .ant-modal-body) {
  padding-top: 0 !important;
}

.st-conflict-summary {
  padding: 14px 16px;
  border-radius: 12px;
  background: #fff7e6;
  border: 1px solid #ffe7ba;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.st-conflict-summary__headline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.st-conflict-summary__title {
  color: #ad6800;
  font-size: 16px;
  font-weight: 700;
  line-height: 1.5;
}

.st-conflict-summary__count {
  color: #8c8c8c;
  font-size: 13px;
  font-weight: 600;
  line-height: 20px;
}

.st-conflict-summary__chips {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.st-conflict-summary__chips-label {
  color: #8c8c8c;
  font-size: 13px;
  line-height: 20px;
}

.st-conflict-summary__chip {
  display: inline-flex;
  align-items: center;
  min-height: 26px;
  padding: 0 10px;
  border-radius: 999px;
  background: #fff1f0;
  color: #ff4d4f;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
}

.st-conflict-summary__hint {
  color: #8c8c8c;
  font-size: 13px;
  line-height: 1.7;
}

.st-conflict-content-spin {
  min-height: 280px;
}

.st-conflict-section-title {
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
  margin: 10px 0;
}

.st-conflict-attempt__card,
.st-conflict-item {
  border: 1px solid #edf2f7;
  border-radius: 14px;
  background: #fff;
}

.st-conflict-attempt__card {
  padding: 14px 16px;
}

.st-conflict-attempt__warning {
  margin-top: 12px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #fff7e6;
  border: 1px solid #ffd591;
  color: #ad6800;
  font-size: 13px;
  line-height: 1.7;
}

.st-conflict-attempt__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 46px;
  height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  background: #1677ff;
  color: #fff;
  font-size: 12px;
  font-weight: 700;
}

.st-conflict-attempt__headline,
.st-conflict-item__headline {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  color: #1f2329;
  font-size: 15px;
  font-weight: 700;
}

.st-conflict-attempt {
  margin-bottom: 10px;
}

.st-conflict-attempt__headline-main {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.st-conflict-attempt__force-tip {
  max-width: 420px;
  padding: 10px 14px;
  border-radius: 12px;
  border: 1px solid #ffd8bf;
  background: #fff2e8;
  text-align: left;
  color: #cf1322;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.4;
  white-space: nowrap;
}

.st-conflict-attempt__meta,
.st-conflict-item__meta {
  margin-top: 6px;
  color: #4b5563;
  font-size: 13px;
  line-height: 1.7;
}

.st-conflict-attempt__meta--time {
  color: #1677ff;
  font-weight: 700;
}

.st-conflict-attempt__target {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
  padding: 12px 14px;
  border-radius: 12px;
  background: #f8fafc;
  border: 1px solid #edf2f7;
}

.st-conflict-attempt__target-label {
  flex-shrink: 0;
  color: #8c8c8c;
  font-size: 13px;
  line-height: 20px;
}

.st-conflict-attempt__target-value {
  color: #1f2329;
  font-size: 13px;
  font-weight: 700;
  line-height: 22px;
}

.st-conflict-attempt__facts {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
}

.st-conflict-attempt__fact {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 38px;
  padding: 8px 12px;
  border-radius: 12px;
  background: #f8fafc;
  border: 1px solid #edf2f7;
  font-size: 13px;
  line-height: 20px;
}

.st-conflict-attempt__fact-label {
  color: #8c8c8c;
}

.st-conflict-attempt__fact-value {
  color: #1f2329;
  font-weight: 700;
  word-break: break-word;
}

.st-conflict-attempt__fact-value--danger {
  color: #ff4d4f;
}

.st-conflict-attempt__reason-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 12px;
}

.st-conflict-attempt__reason-label {
  color: #8c8c8c;
  font-size: 13px;
  line-height: 20px;
}

.st-conflict-attempt__reason-chip {
  display: inline-flex;
  align-items: center;
  min-height: 28px;
  padding: 0 12px;
  border-radius: 999px;
  background: #fff1f0;
  color: #ff4d4f;
  font-size: 12px;
  font-weight: 700;
  line-height: 1;
}

.st-conflict-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.st-conflict-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 16px;
}

.st-conflict-item__main {
  min-width: 0;
  flex: 1;
}

.st-conflict-item__sep {
  margin: 0 4px;
  color: #d9d9d9;
}

.st-conflict-item__value--danger {
  color: #ff4d4f;
  font-weight: 700;
}

.st-conflict-item__meta--reasons {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.st-conflict-item__reason-chip {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 24px;
  padding: 0 10px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 700;
  line-height: 24px;
}

.st-conflict-item__reason-chip--danger {
  background: #fff1f0;
  color: #ff4d4f;
}

.st-conflict-item__side {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}
</style>
