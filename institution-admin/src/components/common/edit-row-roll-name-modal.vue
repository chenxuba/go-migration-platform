<script setup lang="ts">
import { CloseOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { updateStudentTeachingRecordApi, type TeachingRecordDetailStudent } from '@/api/edu-center/class-record'
import { listTuitionAccountsByStudentAndLessonApi } from '@/api/edu-center/one-to-one'
import { getRollCallStudentTuitionAccountsApi } from '@/api/edu-center/roll-call'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  open: boolean
  student?: TeachingRecordDetailStudent | null
  teachingRecordId?: string
  lessonId?: string
  fallbackChargingMode?: number
  defaultQuantity?: number
}>(), {
  student: null,
  teachingRecordId: '',
  lessonId: '',
  fallbackChargingMode: 0,
  defaultQuantity: 0,
})

const emit = defineEmits(['update:open', 'saved'])
const formRef = ref()
const submitting = ref(false)
const loadingTuitionAccount = ref(false)
const syncingFormState = ref(false)
const previousStatus = ref(1)
const resolvedStudent = ref<TeachingRecordDetailStudent | null>(null)
let tuitionAccountLoadSeq = 0

const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const currentStudent = computed(() => resolvedStudent.value || props.student || null)

const formState = reactive({
  status: 1,
  editRecord: undefined as number | undefined,
  internalNote: '',
  externalNote: '',
})

const avatarUrl = computed(() => String(currentStudent.value?.avatar || '').trim() || 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png')
const studentName = computed(() => String(currentStudent.value?.studentName || '').trim() || '-')
const chargingModeText = computed(() => {
  const mode = Number(currentStudent.value?.skuMode || props.fallbackChargingMode || 0)
  if (mode === 2)
    return '按时间'
  if (mode === 3)
    return '按金额'
  if (mode === 1)
    return '按课时'
  return '-'
})
const sourceTypeText = computed(() => {
  const type = Number(currentStudent.value?.sourceType || 0)
  if (type === 2)
    return '临时学员'
  if (type === 3 || type === 7)
    return '补课学员'
  if (type === 4)
    return '试听学员'
  if (type === 6)
    return '1对1学员'
  return '班级学员'
})
const tuitionAccountText = computed(() => {
  if (loadingTuitionAccount.value)
    return '加载中...'
  return String(currentStudent.value?.tuitionAccountName || '').trim() || '-'
})
const remainingLabelText = computed(() => {
  const mode = Number(currentStudent.value?.skuMode || props.fallbackChargingMode || 0)
  if (mode === 2)
    return '剩余时段'
  if (mode === 3)
    return '剩余金额'
  return '剩余课时'
})
const leftQuantityText = computed(() => {
  if (loadingTuitionAccount.value)
    return '加载中...'
  const quantity = Number(currentStudent.value?.leftQuantity || 0)
  if (!Number.isFinite(quantity))
    return '-'
  const text = Number.isInteger(quantity) ? String(quantity) : quantity.toFixed(2).replace(/\.?0+$/, '')
  const mode = Number(currentStudent.value?.skuMode || props.fallbackChargingMode || 0)
  if (mode === 2)
    return `${text}天`
  if (mode === 3)
    return `${text}元`
  return `${text}课时`
})
const isFreeTrialStudent = computed(() => Number(currentStudent.value?.sourceType || 0) === 4)
const showTuitionAccountInfo = computed(() => !isFreeTrialStudent.value)
const quantityDisabled = computed(() => Number(currentStudent.value?.sourceType || 0) === 4 || Number(currentStudent.value?.skuMode || 0) === 2)
const effectiveQuantityDisabled = computed(() => quantityDisabled.value)
const showUnrecordedOption = computed(() => Number(currentStudent.value?.status || 0) === 4 || Number(formState.status || 0) === 4)
const showEditRecord = computed(() => Number(formState.status || 1) !== 4 && !isFreeTrialStudent.value)
const quantityHintText = computed(() => {
  if (quantityDisabled.value)
    return '当前学员课消方式不记录课时'
  return ''
})
const editRecordRules = computed(() => {
  if (!showEditRecord.value)
    return []
  return [{
    validator: async (_rule: unknown, value: number | undefined) => {
      if (value === undefined || value === null || Number.isNaN(Number(value)))
        throw new Error('请填写编辑记录')
    },
    trigger: 'change',
  }]
})

function normalizeQuantity(value?: number) {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return 0
  return num
}

function resolveDefaultQuantity() {
  const defaultQuantity = normalizeQuantity(props.defaultQuantity)
  if (defaultQuantity > 0)
    return defaultQuantity
  const currentQuantity = normalizeQuantity(currentStudent.value?.quantity)
  if (currentQuantity > 0)
    return currentQuantity
  return 1
}

function resolveInitialEditRecord(status: number) {
  if (isFreeTrialStudent.value)
    return 0
  const currentQuantity = normalizeQuantity(currentStudent.value?.quantity)
  if (status === 1) {
    if (currentQuantity > 0)
      return currentQuantity
    return resolveDefaultQuantity()
  }
  if (String(currentStudent.value?.studentTeachingRecordId || '').trim())
    return currentQuantity
  return 0
}

function syncFormState() {
  const student = currentStudent.value
  syncingFormState.value = true
  const status = Number(student?.status ?? 1)
  formState.status = [1, 2, 3, 4].includes(status) ? status : 1
  formState.editRecord = resolveInitialEditRecord(formState.status)
  formState.internalNote = String(student?.remark || '')
  formState.externalNote = String(student?.externalRemark || '')
  previousStatus.value = formState.status
  nextTick(() => {
    syncingFormState.value = false
  })
}

function clearFormValidateState() {
  nextTick(() => {
    formRef.value?.clearValidate?.()
  })
}

function normalizeAccountChargingMode(mode?: number) {
  const parsed = Number(mode || 0)
  if (parsed === 4)
    return 3
  return parsed
}

function effectiveAccountChargingMode(acc: Record<string, any>) {
  const mode = normalizeAccountChargingMode(acc?.lessonChargingMode)
  if (mode > 0)
    return mode
  const totalQty = Number(acc?.totalQuantity || 0) + Number(acc?.totalFreeQuantity || 0)
  const remainQty = Number(acc?.quantity || 0) + Number(acc?.freeQuantity || 0)
  if ((totalQty > 0 || remainQty > 0) && acc?.enableExpireTime)
    return 2
  if (totalQty > 0 || remainQty > 0)
    return 1
  if (Number(acc?.totalTuition || 0) > 0 || Number(acc?.tuition || 0) > 0)
    return 3
  return 0
}

function pickStudentTuitionAccount(list: Array<Record<string, any>>, student?: TeachingRecordDetailStudent | null) {
  const currentId = String(student?.tuitionAccountId || '').trim()
  if (currentId) {
    const matched = list.find(acc => String(acc?.id || '').trim() === currentId)
    if (matched)
      return matched
  }
  return list.find(acc => Boolean(acc?.isTuitionAccountActive)) || list[0] || null
}

async function loadStudentTuitionAccountList(studentId: string, lessonId: string) {
  const rollCallRes = await getRollCallStudentTuitionAccountsApi({ studentId, lessonId })
  if (rollCallRes.code !== 200)
    throw new Error(rollCallRes.message || '加载扣费课程账户失败')
  const rollCallList = Array.isArray(rollCallRes.result?.list) ? rollCallRes.result.list : []
  if (rollCallList.length > 0)
    return rollCallList

  const commonRes = await listTuitionAccountsByStudentAndLessonApi({ studentId, lessonId })
  if (commonRes.code !== 200)
    throw new Error(commonRes.message || '加载扣费课程账户失败')
  return Array.isArray(commonRes.result?.list) ? commonRes.result.list : []
}

async function loadStudentTuitionAccount() {
  const student = currentStudent.value
  const studentId = String(student?.studentId || '').trim()
  const lessonId = String(props.lessonId || '').trim()
  if (!openModal.value || !studentId || !lessonId || isFreeTrialStudent.value)
    return
  if (String(student?.studentTeachingRecordId || '').trim())
    return
  const seq = ++tuitionAccountLoadSeq
  loadingTuitionAccount.value = true
  try {
    const list = await loadStudentTuitionAccountList(studentId, lessonId)
    if (seq !== tuitionAccountLoadSeq)
      return
    const selectedAccount = pickStudentTuitionAccount(list, student)
    if (!selectedAccount) {
      resolvedStudent.value = {
        ...(resolvedStudent.value || props.student || {}),
        tuitionAccountId: '',
        tuitionAccountName: '',
        isTuitionAccountActive: false,
        leftQuantity: 0,
        skuMode: Number(props.student?.skuMode || props.fallbackChargingMode || 0),
      } as TeachingRecordDetailStudent
      return
    }
    const mode = effectiveAccountChargingMode(selectedAccount)
    const remainQuantity = mode === 3
      ? Number(selectedAccount?.tuition || 0)
      : Number(selectedAccount?.quantity || 0) + Number(selectedAccount?.freeQuantity || 0)
    resolvedStudent.value = {
      ...(resolvedStudent.value || props.student || {}),
      tuitionAccountId: String(selectedAccount?.id || ''),
      tuitionAccountName: String(selectedAccount?.productName || selectedAccount?.lessonName || ''),
      isTuitionAccountActive: Boolean(selectedAccount?.isTuitionAccountActive),
      leftQuantity: remainQuantity,
      skuMode: Number(mode || props.fallbackChargingMode || 0),
    } as TeachingRecordDetailStudent
  }
  catch (error: any) {
    console.error('load edit roll call tuition account failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '加载扣费课程账户失败')
  }
  finally {
    if (seq === tuitionAccountLoadSeq)
      loadingTuitionAccount.value = false
  }
}

watch(
  () => props.student,
  (student) => {
    resolvedStudent.value = student ? { ...student } : null
  },
  { immediate: true },
)

watch(
  () => [openModal.value, props.student] as const,
  ([open]) => {
    if (!open)
      return
    syncFormState()
    clearFormValidateState()
    loadStudentTuitionAccount()
  },
  { immediate: true },
)

watch(
  () => openModal.value,
  (open) => {
    if (open)
      return
    clearFormValidateState()
  },
)

watch(
  () => formState.status,
  (status) => {
    if (!openModal.value || syncingFormState.value)
      return
    const nextStatus = Number(status || 1)
    if (isFreeTrialStudent.value) {
      formState.editRecord = 0
      previousStatus.value = nextStatus
      return
    }
    if (nextStatus === 4)
      formState.editRecord = 0
    if (nextStatus === 1 && previousStatus.value !== 1 && normalizeQuantity(formState.editRecord) <= 0)
      formState.editRecord = resolveDefaultQuantity()
    if ((nextStatus === 2 || nextStatus === 3) && previousStatus.value === 1)
      formState.editRecord = 0
    previousStatus.value = nextStatus
  },
)

async function handleSubmit() {
  const studentTeachingRecordId = String(currentStudent.value?.studentTeachingRecordId || '').trim()
  if (submitting.value)
    return
  try {
    await formRef.value?.validate()
    submitting.value = true
    const res = await updateStudentTeachingRecordApi({
      studentTeachingRecordId: studentTeachingRecordId || undefined,
      teachingRecordId: String(props.teachingRecordId || '').trim() || undefined,
      studentId: String(currentStudent.value?.studentId || '').trim() || undefined,
      sourceType: Number(currentStudent.value?.sourceType || 0) || undefined,
      status: Number(formState.status || 1),
      quantity: effectiveQuantityDisabled.value ? 0 : Number(formState.editRecord || 0),
      remark: String(formState.internalNote || '').trim(),
      externalRemark: String(formState.externalNote || '').trim(),
    })
    if (res.code !== 200 || res.result !== true)
      throw new Error(res.message || '编辑点名保存失败')
    messageService.success('编辑点名成功')
    emit('saved')
    openModal.value = false
  }
  catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '编辑点名保存失败')
  }
  finally {
    submitting.value = false
  }
}

function closeFun() {
  if (submitting.value)
    return
  openModal.value = false
}
</script>

<template>
  <a-modal
    v-model:open="openModal"
    centered
    class="modal-content-box"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="800"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>编辑点名</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="rounded-12px bg-#fafafa mx-24px">
      <div class="contenter px6 py3" style="margin-bottom: 0;">
        <div class="avatar flex flex-items-center">
          <img :src="avatarUrl" class="w-40px h-40px rounded-full mr-8px" alt="">
          <span class="text-5 text-#222 font-800">{{ studentName }}</span>
        </div>
        <div class="text-14px text-#222 flex mt-12px flex-items-center gap-x-60px whitespace-nowrap overflow-x-auto">
          <div class="flex flex-items-center flex-none">
            <span class="text-#888">学员身份：</span>
            <span>{{ sourceTypeText }}</span>
          </div>
          <div class="flex flex-items-center flex-none">
            <span class="text-#888">课消方式：</span>
            <span>{{ chargingModeText }}</span>
          </div>
          <div v-if="showTuitionAccountInfo" class="flex flex-items-center flex-none">
            <span class="text-#888">扣费课程账户：</span>
            <span>{{ tuitionAccountText }}</span>
          </div>
          <div v-if="showTuitionAccountInfo" class="flex flex-items-center flex-none">
            <span class="text-#888">{{ remainingLabelText }}：</span>
            <span>{{ leftQuantityText }}</span>
          </div>
        </div>
      </div>
      <div class="w-752px">
        <a-divider class="my-0" />
      </div>
      <div class="contenter scrollbar" style="margin: 0 24px 24px 24px;">
        <a-form ref="formRef" :model="formState" :label-col="{ span: 4 }" :wrapper-col="{ span: 18 }">
          <a-form-item label="编辑状态" name="status" :rules="[{ required: true, message: '请选择编辑状态' }]">
            <div class="flex flex-items-center flex-wrap gap-y-8px">
              <a-radio-group v-model:value="formState.status" class="custom-radio">
                <a-radio :value="1">
                  到课
                </a-radio>
                <a-radio :value="3">
                  请假
                </a-radio>
                <a-radio :value="2">
                  旷课
                </a-radio>
                <a-radio v-if="showUnrecordedOption" :value="4">
                  <span class="inline-flex items-center">
                    <span>未记录</span>
                    <a-popover title="未记录" trigger="hover">
                      <template #content>
                        <div class="w-220px leading-22px">
                          学员为“未记录”状态时，无法记录课时，也不会发送家长端消息提醒。
                        </div>
                      </template>
                      <ExclamationCircleOutlined class="ml-6px text-#999 cursor-pointer" @click.stop />
                    </a-popover>
                  </span>
                </a-radio>
              </a-radio-group>
              <span v-if="isFreeTrialStudent" class="ml-12px text-12px text-#888 leading-20px">
                免费试听学员，不支持记录课时
              </span>
            </div>
          </a-form-item>

          <a-form-item v-if="showEditRecord" label="编辑记录" name="editRecord" :rules="editRecordRules">
            <div class="flex flex-items-center">
              <a-input-number v-model:value="formState.editRecord" :precision="2" :min="0" :max="100" :disabled="effectiveQuantityDisabled" />
              <span class="ml-4px">课时</span>
            </div>
            <div v-if="quantityHintText" class="text-12px text-#888 mt-6px">
              {{ quantityHintText }}
            </div>
          </a-form-item>

          <a-form-item label="编辑对内备注" name="internalNote">
            <a-input v-model:value="formState.internalNote" placeholder="选填（200字以内）" :maxlength="200" />
          </a-form-item>

          <a-form-item label="编辑对外备注" name="externalNote">
            <a-input v-model:value="formState.externalNote" placeholder="选填（200字以内）" :maxlength="200" />
          </a-form-item>
        </a-form>
      </div>
    </div>
    <template #footer>
      <a-button danger ghost :disabled="submitting" @click="closeFun">
        取消
      </a-button>
      <a-button type="primary" ghost :loading="submitting" @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
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

.contenter {
  padding: 24px;
  background: #fafafa;
  margin: 24px 0px;
}
</style>

<style>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
