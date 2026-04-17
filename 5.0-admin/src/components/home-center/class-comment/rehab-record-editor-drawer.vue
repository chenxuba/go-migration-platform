<script setup lang="ts">
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { computed, reactive, ref, watch } from 'vue'
import messageService from '@/utils/messageService'

interface RehabRecordStudent {
  id?: string
  name?: string
  avatar?: string
  status?: string
  gender?: string | number
  sex?: string | number
  birthday?: string
  birthDate?: string
}

interface RehabRecordSession {
  sourceName?: string
  lessonName?: string
  teacherName?: string
  classRoomName?: string
  startTime?: string
  endTime?: string
}

interface RehabRecordParentFeedback {
  content?: string
  parentFeedbackContent?: string
  signature?: string
  parentSignature?: string
  date?: string
  feedbackDate?: string
}

const props = withDefaults(defineProps<{
  student?: Partial<RehabRecordStudent> | null
  session?: Partial<RehabRecordSession> | null
  parentFeedback?: Partial<RehabRecordParentFeedback> | null
}>(), {
  student: null,
  session: null,
  parentFeedback: null,
})

const open = defineModel<boolean>({
  default: false,
})

const genderOptions = [
  { label: '男', value: '男' },
  { label: '女', value: '女' },
]

interface TrainingModuleItem {
  id: number
  title: string
  content: string
}

const trainingModuleSeed = ref(2)
const trainingModules = ref<TrainingModuleItem[]>([
  {
    id: 1,
    title: '',
    content: '',
  },
  {
    id: trainingModuleSeed.value,
    title: '',
    content: '',
  },
])

const formModel = reactive({
  studentName: '',
  gender: undefined as string | undefined,
  birthDate: undefined as string | undefined,
  className: '',
  teacherName: '',
  trainingDate: undefined as string | undefined,
  trainingTarget: '',
  performance: '',
  suggestion: '',
  parentFeedback: '',
  parentSignature: '',
  feedbackDate: undefined as string | undefined,
})

function normalizeTextValue(value?: string | number | null) {
  return String(value ?? '').trim()
}

function resolveGenderValue() {
  const rawValue = props.student?.gender ?? props.student?.sex
  if (rawValue === undefined || rawValue === null || rawValue === '')
    return undefined

  if (rawValue === 1 || rawValue === '1' || rawValue === '男')
    return '男'

  if (rawValue === 2 || rawValue === '2' || rawValue === '女')
    return '女'

  return normalizeTextValue(rawValue) || undefined
}

function normalizeDateValue(value?: string | null) {
  if (!value)
    return undefined
  const date = dayjs(value)
  return date.isValid() ? date.format('YYYY-MM-DD') : undefined
}

function hydrateBasicInfo() {
  formModel.studentName = normalizeTextValue(props.student?.name)
  formModel.gender = resolveGenderValue()
  formModel.birthDate = normalizeDateValue(props.student?.birthDate || props.student?.birthday)
  formModel.className = normalizeTextValue(props.session?.sourceName)
  formModel.teacherName = normalizeTextValue(props.session?.teacherName)
  formModel.trainingDate = normalizeDateValue(props.session?.startTime)
}

function hydrateParentFeedback() {
  formModel.parentFeedback = normalizeTextValue(props.parentFeedback?.content || props.parentFeedback?.parentFeedbackContent)
  formModel.parentSignature = normalizeTextValue(props.parentFeedback?.signature || props.parentFeedback?.parentSignature)
  formModel.feedbackDate = normalizeDateValue(props.parentFeedback?.date || props.parentFeedback?.feedbackDate)
}

const showParentFeedbackSection = computed(() => {
  return Boolean(
    normalizeTextValue(props.parentFeedback?.content || props.parentFeedback?.parentFeedbackContent)
    || normalizeTextValue(props.parentFeedback?.signature || props.parentFeedback?.parentSignature)
    || normalizeDateValue(props.parentFeedback?.date || props.parentFeedback?.feedbackDate),
  )
})

watch(() => open.value, (value) => {
  if (value) {
    hydrateBasicInfo()
    hydrateParentFeedback()
  }
})

watch([() => props.student, () => props.session, () => props.parentFeedback], () => {
  if (open.value) {
    hydrateBasicInfo()
    hydrateParentFeedback()
  }
}, { deep: true })

function handleAddTrainingModule() {
  trainingModuleSeed.value += 1
  trainingModules.value.push({
    id: trainingModuleSeed.value,
    title: '',
    content: '',
  })
}

function handleRemoveTrainingModule(id: number) {
  if (trainingModules.value.length <= 1)
    return
  trainingModules.value = trainingModules.value.filter(item => item.id !== id)
}

function getTrainingModuleColSpan(index: number) {
  const count = trainingModules.value.length
  if (count === 1)
    return 24
  if (count % 2 === 1 && index === count - 1)
    return 24
  return 12
}

function handlePreviewAction(action: string) {
  messageService.info(`${action}暂未开发，当前为静态页面预览`)
}
</script>

<template>
  <a-drawer
    v-model:open="open"
    :body-style="{ padding: '0', background: '#f7f7fd', display: 'flex', flexDirection: 'column' }"
    :closable="false"
    width="980px"
    placement="right"
    :z-index="1100"
  >
    <template #title>
      <div class="custom-header flex justify-between h-4 flex-items-center">
        <div class="text-5">
          康复训练记录
        </div>
        <a-button type="text" class="close-btn" @click="open = false">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>

    <div class="h-full flex flex-col min-h-0">
      <div class="flex-1 min-h-0 overflow-auto p-12px">
        <div class="flex flex-col gap-16px">
          <div class="rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
            <div class="text-15px leading-22px font-600 text-#222">
              基础信息
            </div>

            <a-row :gutter="[12, 12]" class="mt-14px">
              <a-col :xs="24" :sm="12" :lg="8">
                <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                  <div class="mb-10px text-12px leading-18px text-#8c8c8c">姓名</div>
                  <a-input v-model:value="formModel.studentName" placeholder="请输入学员姓名" />
                </div>
              </a-col>
              <a-col :xs="24" :sm="12" :lg="8">
                <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                  <div class="mb-10px text-12px leading-18px text-#8c8c8c">性别</div>
                  <a-select
                    v-model:value="formModel.gender"
                    :options="genderOptions"
                    allow-clear
                    style="width: 100%;"
                    placeholder="请选择性别"
                  />
                </div>
              </a-col>
              <a-col :xs="24" :sm="12" :lg="8">
                <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                  <div class="mb-10px text-12px leading-18px text-#8c8c8c">出生年月</div>
                  <a-date-picker
                    v-model:value="formModel.birthDate"
                    value-format="YYYY-MM-DD"
                    style="width: 100%;"
                    placeholder="请选择出生年月"
                  />
                </div>
              </a-col>
              <a-col :xs="24" :sm="12" :lg="8">
                <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                  <div class="mb-10px text-12px leading-18px text-#8c8c8c">班别</div>
                  <a-input v-model:value="formModel.className" placeholder="请输入班别" />
                </div>
              </a-col>
              <a-col :xs="24" :sm="12" :lg="8">
                <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                  <div class="mb-10px text-12px leading-18px text-#8c8c8c">任教老师</div>
                  <a-input v-model:value="formModel.teacherName" placeholder="请输入任教老师" />
                </div>
              </a-col>
              <a-col :xs="24" :sm="12" :lg="8">
                <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                  <div class="mb-10px text-12px leading-18px text-#8c8c8c">训练日期</div>
                  <a-date-picker
                    v-model:value="formModel.trainingDate"
                    value-format="YYYY-MM-DD"
                    style="width: 100%;"
                    placeholder="请选择训练日期"
                  />
                </div>
              </a-col>
            </a-row>
          </div>

          <div class="rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
            <div class="text-15px leading-22px font-600 text-#222">
              训练目标
            </div>
            <div class="mt-14px">
              <a-textarea
                v-model:value="formModel.trainingTarget"
                :auto-size="{ minRows: 4, maxRows: 6 }"
                placeholder="请输入本次训练目标"
              />
            </div>
          </div>

          <div class="rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
            <div class="flex items-center justify-between gap-12px">
              <div class="text-15px leading-22px font-600 text-#222">
                训练项目
              </div>
              <a-button type="dashed" size="small" @click="handleAddTrainingModule">
                新增项目
              </a-button>
            </div>

            <a-row :gutter="[12, 12]" class="mt-14px">
              <a-col
                v-for="(item, index) in trainingModules"
                :key="item.id"
                :xs="24"
                :lg="getTrainingModuleColSpan(index)"
              >
                <div class="h-full rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-16px">
                  <div class="flex items-center justify-between gap-12px mb-12px">
                    <a-input
                      v-model:value="item.title"
                      class="flex-1"
                      placeholder="请输入训练项目名称"
                    />
                    <a-button
                      v-if="trainingModules.length > 1"
                      type="link"
                      danger
                      class="px-0"
                      @click="handleRemoveTrainingModule(item.id)"
                    >
                      删除
                    </a-button>
                  </div>

                  <a-textarea
                    v-model:value="item.content"
                    :auto-size="{ minRows: 5, maxRows: 8 }"
                    placeholder="请输入训练项目内容"
                  />
                </div>
              </a-col>
            </a-row>
          </div>

          <a-row :gutter="[16, 16]">
            <a-col :xs="24" :lg="12">
              <div class="h-full rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
                <div class="text-15px leading-22px font-600 text-#222">
                  学生综合表现
                </div>
                <div class="mt-14px">
                  <a-textarea
                    v-model:value="formModel.performance"
                    :auto-size="{ minRows: 7, maxRows: 10 }"
                    placeholder="请输入学生综合表现"
                  />
                </div>
              </div>
            </a-col>

            <a-col :xs="24" :lg="12">
              <div class="h-full rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
                <div class="text-15px leading-22px font-600 text-#222">
                  康复建议
                </div>
                <div class="mt-14px">
                  <a-textarea
                    v-model:value="formModel.suggestion"
                    :auto-size="{ minRows: 7, maxRows: 10 }"
                    placeholder="请输入康复建议"
                  />
                </div>
              </div>
            </a-col>
          </a-row>

          <div v-if="showParentFeedbackSection" class="rounded-14px border border-solid border-#e9edf5 bg-white p-18px">
            <div class="text-15px leading-22px font-600 text-#222">
              家长意见反馈
            </div>

            <div class="mt-14px">
              <a-textarea
                v-model:value="formModel.parentFeedback"
                :auto-size="{ minRows: 4, maxRows: 6 }"
                placeholder="请输入家长意见反馈"
              />
            </div>

            <a-row :gutter="[12, 12]" class="mt-16px">
              <a-col :xs="24" :lg="12">
                <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                  <div class="mb-10px text-12px leading-18px text-#8c8c8c">家长签名</div>
                  <a-input
                    v-model:value="formModel.parentSignature"
                    placeholder="请输入家长签名"
                  />
                </div>
              </a-col>
              <a-col :xs="24" :lg="12">
                <div class="rounded-12px border border-solid border-#edf0f5 bg-#fafbfc p-14px">
                  <div class="mb-10px text-12px leading-18px text-#8c8c8c">日期</div>
                  <a-date-picker
                    v-model:value="formModel.feedbackDate"
                    value-format="YYYY-MM-DD"
                    style="width: 100%;"
                    placeholder="请选择日期"
                  />
                </div>
              </a-col>
            </a-row>
          </div>
          </div>
        </div>
      </div>

      <div class="flex justify-end px-20px pt-12px pb-16px bg-white border-t border-solid border-#eef0f5">
        <a-space>
          <a-button @click="open = false">
            取消
          </a-button>
          <a-button @click="handlePreviewAction('保存草稿')">
            保存草稿
          </a-button>
          <a-button type="primary" @click="handlePreviewAction('发布记录')">
            发布记录
          </a-button>
        </a-space>
      </div>
  </a-drawer>
</template>

<style lang="less" scoped>
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
</style>
