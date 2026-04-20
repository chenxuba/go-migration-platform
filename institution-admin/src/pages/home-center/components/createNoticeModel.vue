<script setup lang="ts">
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import { CloseOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import dayjs from 'dayjs'
import { computed, nextTick, reactive, ref, watch } from 'vue'
import NoticePhonePreviewModal from './noticePhonePreviewModal.vue'
import NoticeScopePickerModel from './noticeScopePickerModel.vue'
import NoticeStudentPickerModel from './noticeStudentPickerModel.vue'
import type { NoticePickerCompletePayload, NoticePickerSelection, NoticePickerSource } from './notice-picker.types'
import {
  checkNoticeFilterWordsApi,
  checkRepeatNoticeStudentApi,
  createNoticeApi,
  type NoticeTemplateItem,
} from '@/api/home-center/notice'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  selectedTemplate?: NoticeTemplateItem | null
}>(), {
  selectedTemplate: null,
})

const emit = defineEmits<{
  (e: 'success'): void
}>()

const open = defineModel<boolean>({
  default: false,
})

const formRef = ref()
const noticeEditorRef = ref<any>()
const submitting = ref(false)
const openDrawer = ref(false)
const rulesModalOpen = ref(false)
const scopePickerOpen = ref(false)
const studentPickerOpen = ref(false)
const previewOpen = ref(false)
const pendingTemplateContent = ref('')

const formState = reactive({
  title: '',
  content: '',
  selectedSources: [] as NoticePickerSource[],
  selectedStudents: [] as NoticePickerSelection[],
})

const settingsForm = reactive({
  scope: '1',
  publishType: '1',
  publishAt: undefined as string | undefined,
  isConfirm: false,
})

const editorOption = {
  modules: {
    toolbar: {
      container: '#notice-toolbar',
    },
  },
  theme: 'snow',
}

const selectedTargetButtonText = computed(() => formState.selectedStudents.length > 0 ? `已选班级/学员（${formState.selectedStudents.length}）` : '选择班级/学员')
const previewMetaPrimary = computed(() => String(props.selectedTemplate?.tag || '').trim() || '通知预览')
const previewPublishText = computed(() => dayjs().format('MM-DD HH:mm'))
const selectedStudentPreviewText = computed(() => {
  const names = Array.from(new Set(formState.selectedStudents.map(item => String(item.studentName || '').trim()).filter(Boolean)))
  if (names.length <= 8)
    return names.join('、')
  return `${names.slice(0, 8).join('、')} 等 ${names.length} 人`
})

function applySelectedTemplateToForm() {
  formState.title = String(props.selectedTemplate?.title || '').trim()
  const rawContent = String(props.selectedTemplate?.content || '').trim()
  pendingTemplateContent.value = rawContent
  formState.content = parseTemplateDeltaContent(rawContent) ? '' : rawContent
}

function parseTemplateDeltaContent(rawContent?: string | null) {
  const text = String(rawContent || '').trim()
  if (!text || (!text.startsWith('[') && !text.startsWith('{') && !text.startsWith('"')))
    return null

  const normalizeParsed = (parsed: any) => {
    if (Array.isArray(parsed))
      return { ops: parsed }
    if (parsed && typeof parsed === 'object' && Array.isArray(parsed.ops))
      return parsed
    return null
  }

  try {
    const parsed = JSON.parse(text)
    const normalized = normalizeParsed(parsed)
    if (normalized)
      return normalized
    if (typeof parsed === 'string')
      return normalizeParsed(JSON.parse(parsed))
  }
  catch {
    return null
  }

  return null
}

function getNoticeEditorInstance() {
  const editor = noticeEditorRef.value
  return editor?.getQuill?.() || editor?.$?.exposed?.getQuill?.() || editor?.quill || null
}

async function syncTemplateContentToEditor() {
  const rawContent = String(pendingTemplateContent.value || '').trim()
  const quill = getNoticeEditorInstance()
  if (!quill) {
    formState.content = parseTemplateDeltaContent(rawContent) ? '' : rawContent
    return
  }

  if (!rawContent) {
    quill.setText?.('')
    formState.content = ''
    return
  }

  const parsedDelta = parseTemplateDeltaContent(rawContent)
  if (!parsedDelta) {
    formState.content = rawContent
    return
  }

  quill.setContents?.(parsedDelta)
  const html = String(quill.root?.innerHTML || '').trim()
  formState.content = html === '<p><br></p>' ? '' : html
}

function resetFormState() {
  formState.title = ''
  formState.content = ''
  formState.selectedSources = []
  formState.selectedStudents = []
  settingsForm.scope = '1'
  settingsForm.publishType = '1'
  settingsForm.publishAt = undefined
  settingsForm.isConfirm = false
}

async function initializeFormState() {
  resetFormState()
  applySelectedTemplateToForm()
  await nextTick()
  formRef.value?.clearValidate?.()
}

function handleRuleConfirm() {
  applySelectedTemplateToForm()
  rulesModalOpen.value = false
  openDrawer.value = true
}

function handleDrawerClose() {
  openDrawer.value = false
  scopePickerOpen.value = false
  studentPickerOpen.value = false
  previewOpen.value = false
  open.value = false
}

function handlePreview() {
  previewOpen.value = true
}

async function syncSelectedRange(payload: NoticePickerCompletePayload) {
  formState.selectedSources = payload.selectedSources
  formState.selectedStudents = payload.selectedStudents
  await nextTick()
  formRef.value?.clearValidate?.(['selectedStudents'])
}

function handleScopePickerComplete(payload: NoticePickerCompletePayload) {
  void syncSelectedRange(payload)
}

function handleStudentPickerComplete(payload: NoticePickerCompletePayload) {
  void syncSelectedRange(payload)
}

function openStudentPicker() {
  if (!formState.selectedSources.length) {
    messageService.warning('请先选择班级/1v1')
    return
  }
  studentPickerOpen.value = true
}

function stripHtml(text: string) {
  return String(text || '')
    .replace(/<[^>]*>/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}

function buildSummaryText() {
  const plain = stripHtml(formState.content)
  if (plain)
    return plain.slice(0, 120)
  return String(formState.title || '').trim().slice(0, 120)
}

function buildNoticeReceivers() {
  if (settingsForm.scope === '1') {
    return {
      classIds: [] as string[],
      studentIds: [] as string[],
    }
  }

  const classIds: string[] = []
  const studentIds: string[] = []
  const selectedBySource = new Map<string, NoticePickerSelection[]>()

  formState.selectedStudents.forEach((student) => {
    const sourceKey = `${student.sourceType}:${student.sourceId}`
    if (!selectedBySource.has(sourceKey))
      selectedBySource.set(sourceKey, [])
    selectedBySource.get(sourceKey)?.push(student)
  })

  formState.selectedSources.forEach((source) => {
    const sourceKey = `${source.sourceType}:${source.sourceId}`
    const selectedStudents = selectedBySource.get(sourceKey) || []
    if (!selectedStudents.length)
      return

    if (source.sourceType === 'one_to_one') {
      classIds.push(source.sourceId)
      return
    }

    if (selectedStudents.length === source.students.length && source.students.length > 0) {
      classIds.push(source.sourceId)
      return
    }

    selectedStudents.forEach((student) => {
      studentIds.push(student.studentId)
    })
  })

  return {
    classIds: Array.from(new Set(classIds.filter(Boolean))),
    studentIds: Array.from(new Set(studentIds.filter(Boolean))),
  }
}

function disabledDate(current: dayjs.Dayjs) {
  return current && current < dayjs().startOf('day')
}

async function submitNotice(forceRepeat = false) {
  try {
    await formRef.value?.validate?.()
  }
  catch {
    return
  }

  if (!String(formState.content || '').trim()) {
    messageService.warning('请输入通知内容')
    return
  }

  if (settingsForm.scope === '2') {
    if (!formState.selectedSources.length) {
      messageService.warning('请选择班级/1v1')
      return
    }
    if (!formState.selectedStudents.length) {
      messageService.warning('请选择学员')
      return
    }
  }

  if (settingsForm.publishType === '2' && !settingsForm.publishAt) {
    messageService.warning('请选择发布时间')
    return
  }

  const summary = buildSummaryText()
  const noticeTitle = String(formState.title || '').trim()
  const noticeContent = String(formState.content || '').trim()

  submitting.value = true
  try {
    const filterRes = await checkNoticeFilterWordsApi({
      title: noticeTitle,
      content: noticeContent,
      summary,
    })

    if (filterRes.code !== 200) {
      messageService.error(filterRes.message || '通知内容校验失败')
      return
    }

    const hitFilterWords = Array.isArray(filterRes.result) ? filterRes.result.filter(Boolean) : []
    if (hitFilterWords.length > 0) {
      Modal.warning({
        title: '命中过滤词',
        content: `检测到敏感词：${hitFilterWords.join('、')}`,
        okText: '知道了',
      })
      return
    }

    const receivers = buildNoticeReceivers()
    const checkStudentIds = Array.from(new Set(formState.selectedStudents.map(item => item.studentId).filter(Boolean)))

    if (settingsForm.scope === '2' && receivers.classIds.length === 0 && receivers.studentIds.length === 0) {
      messageService.warning('请选择有效的通知对象')
      return
    }

    if (!forceRepeat && settingsForm.scope === '2' && checkStudentIds.length > 0) {
      const repeatRes = await checkRepeatNoticeStudentApi({ studentIds: checkStudentIds })
      if (repeatRes.code !== 200) {
        messageService.error(repeatRes.message || '学员重复校验失败')
        return
      }

      if (repeatRes.result) {
        Modal.confirm({
          title: '发送提示',
          content: '部分学员今日已接收过通知，继续发布可能导致发送失败，是否继续发布？',
          okText: '继续发布',
          cancelText: '返回修改',
          onOk: async () => {
            await submitNotice(true)
          },
        })
        return
      }
    }

    const publishMoment = settingsForm.publishType === '2' && settingsForm.publishAt
      ? dayjs(settingsForm.publishAt)
      : null

    const payload = {
      noticeTemplateId: props.selectedTemplate?.id || undefined,
      title: noticeTitle,
      content: noticeContent,
      isAllSchool: settingsForm.scope === '1',
      isDelaySend: settingsForm.publishType === '2',
      publishDate: publishMoment?.isValid() ? publishMoment.format('YYYY-MM-DD') : undefined,
      hour: publishMoment?.isValid() ? publishMoment.hour() : undefined,
      isConfirm: settingsForm.isConfirm,
      summary,
      classIds: settingsForm.scope === '1' ? [] : receivers.classIds,
      studentIds: settingsForm.scope === '1' ? [] : receivers.studentIds,
    }

    const res = await createNoticeApi(payload)
    if (res.code !== 200) {
      messageService.error(res.message || '创建通知失败')
      return
    }

    messageService.success('通知创建成功')
    handleDrawerClose()
    emit('success')
  }
  catch (error: any) {
    console.error('create notice failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '创建通知失败')
  }
  finally {
    submitting.value = false
  }
}

watch(() => open.value, (visible) => {
  if (!visible) {
    openDrawer.value = false
    rulesModalOpen.value = false
    scopePickerOpen.value = false
    studentPickerOpen.value = false
    return
  }

  void initializeFormState()
  rulesModalOpen.value = true
})

watch(() => props.selectedTemplate, () => {
  if (!open.value)
    return
  applySelectedTemplateToForm()
})

watch(() => openDrawer.value, async (visible) => {
  if (!visible)
    return
  await nextTick()
  await syncTemplateContentToEditor()
})
</script>

<template>
  <div>
    <a-modal
      v-model:open="rulesModalOpen"
      :keyboard="false"
      :mask-closable="false"
      class="noticeModel"
      width="800px"
      title="通知公告规则"
      destroy-on-close
    >
      <div class="rulesWrapper">
        <div class="rulesTitle">
          一、内容规范
        </div>
        <div class="rulesContent">
          用户使用微信公众平台通知服务，须遵守平台相关运营规范。为避免发送的内容引起学员家长投诉导致微信封禁，请勿发布以下违规内容。<br>1.
          发送内容与服务场景不一致（含标题、关键词）的模板消息。<br>2. 在文字或图片中含有广告营销类内容。<br>如：报课优惠类通知、报课返利类通知、课程降价类通知等一些涉及消费的营销类通知。<br>3.
          发送红包、卡券、优惠券、代金券、会员卡类。<br>如：报课领红包、参加活动领优惠券、预存金额送代金券等。<br>4.
          频繁发送相同内容或性质的消息，对用户造成骚扰。原则上，仅支持一个自然日对同一用户发送一次消息。<br>如：频率过高的到期提醒类通知、频率过高的缴费提醒类通知、频率过高的留言提醒类通知、订阅提醒类通知等。<br>处罚规则：<br>一经发现将根据违规程度采取阶梯性封禁通知公告功能等措施。<br>更多运营规范内容可参考：<a
            target="_blank"
            href="https://mp.weixin.qq.com/mp/opshowpage?action=newoplaw#t3-3-9"
            rel="noreferrer"
          >《微信公众平台运营规范》</a>
        </div><br>
        <div>
          <div class="rulesTitle">
            二、审核注意事项
          </div>
          <div class="rulesContent">
            为避免过度打扰，仅支持一个自然日内对同一学员发送一条通知公告。如果某学员已经被发送过通知公告，则当日发送给该学员的其他通知公告将发送失败。<br>通知公告内容审核通过则立即发送，如有疑问，可咨询云宝客户经理。
          </div>
        </div>
      </div>
      <template #footer>
        <a-button key="submit" type="primary" @click="handleRuleConfirm">
          知道了
        </a-button>
      </template>
    </a-modal>

    <a-drawer
      v-model:open="openDrawer"
      width="800px"
      :keyboard="false"
      :mask-closable="false"
      placement="right"
      :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false"
      destroy-on-close
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            创建通知
          </div>
          <a-button type="text" class="close-btn" @click="handleDrawerClose">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>

      <div class="p-20px">
        <div class="px-15px py-20px bg-white rounded-12px">
          <custom-title title="内容" font-size="20px" font-weight="550">
            <template #left>
              <div class="flex justify-between items-center gap-5px">
                <span>内容</span>
                <div
                  class="text-14px text-#999 flex items-center gap-5px cursor-pointer hover:text-#06f"
                  @click="rulesModalOpen = true"
                >
                  <QuestionCircleOutlined />
                  <span>规则</span>
                </div>
              </div>
            </template>
            <template #right>
              <a-button type="link" size="small" class="text-12px" @click="handlePreview">
                预览
              </a-button>
            </template>
          </custom-title>

          <a-form
            ref="formRef"
            class="mt-20px"
            :model="formState"
            name="noticeForm"
            :label-col="{ span: 3 }"
            :wrapper-col="{ span: 21 }"
            autocomplete="off"
          >
            <a-form-item label="通知标题" name="title" :rules="[{ required: true, message: '请输入通知标题' }]">
              <a-input v-model:value="formState.title" :maxlength="20" class="w-240px" placeholder="请输入(最多20字)" />
            </a-form-item>

            <a-form-item label="通知内容" name="content" :rules="[{ required: true, message: '请输入通知内容' }]">
              <div id="notice-toolbar">
                <select class="ql-header">
                  <option value="2">
                    标题2
                  </option>
                  <option value="3">
                    标题3
                  </option>
                  <option value="false" selected>
                    正文
                  </option>
                </select>
                <button class="ql-bold" />
                <button class="ql-italic" />
                <button class="ql-underline" />
                <button class="ql-strike" />
                <select class="ql-color" />
                <select class="ql-background" />
                <select class="ql-size">
                  <option value="small">
                    字号10
                  </option>
                  <option selected>
                    默认字号
                  </option>
                  <option value="large">
                    字号18
                  </option>
                  <option value="huge">
                    字号32
                  </option>
                </select>
                <button class="ql-image" />
              </div>
              <QuillEditor
                ref="noticeEditorRef"
                v-model:content="formState.content"
                content-type="html"
                placeholder="在编辑通知公告时，请注意：
1.直接从外部复制的图片可能无法在微信小程序中正常显示，请尽量使用编辑器的图片上传功能来上传您的图片。
2.文本样式可能需要调整以保证在微信小程序中的显示效果。
发布前，建议预览内容以确保一切显示正常。"
                style="height: 500px;"
                :options="editorOption"
              />
            </a-form-item>
          </a-form>
        </div>

        <div class="px-15px py-20px bg-white rounded-12px mt-15px">
          <custom-title title="发布设置" font-size="20px" font-weight="550" />
          <a-form
            class="mt-20px"
            :model="settingsForm"
            name="noticeSettingForm"
            :label-col="{ span: 4 }"
            :wrapper-col="{ span: 20 }"
            autocomplete="off"
          >
            <a-form-item name="scope" :rules="[{ required: true, message: '请选择通知范围' }]">
              <template #label>
                <span class="mr-2px">通知范围</span>
                <a-tooltip title="未关注家校平台的学员家长，无法接收通知公告">
                  <QuestionCircleOutlined />
                </a-tooltip>
              </template>
              <a-radio-group v-model:value="settingsForm.scope" class="custom-radio">
                <a-radio value="1">
                  全校群发
                </a-radio>
                <a-radio value="2">
                  选择班级/1v1
                </a-radio>
              </a-radio-group>
            </a-form-item>

            <a-form-item
              v-if="settingsForm.scope === '2'"
              label="通知对象"
              name="selectedStudents"
              :rules="[{ required: true, message: '请选择通知对象' }]"
            >
              <div class="notice-scope-selector">
                <a-button type="primary" ghost @click="scopePickerOpen = true">
                  {{ selectedTargetButtonText }}
                </a-button>
              </div>
              <div
                v-if="selectedStudentPreviewText"
                class="notice-scope-preview"
                @click="openStudentPicker"
              >
                {{ selectedStudentPreviewText }}
              </div>
            </a-form-item>

            <a-form-item label="发布时间" name="publishType" :rules="[{ required: true, message: '请选择发布时间' }]">
              <a-radio-group v-model:value="settingsForm.publishType" class="custom-radio">
                <a-radio value="1">
                  立即发布
                </a-radio>
                <a-radio value="2">
                  定时发布
                </a-radio>
              </a-radio-group>
            </a-form-item>

            <a-form-item v-if="settingsForm.publishType === '2'" label="发布时刻" name="publishAt" :rules="[{ required: true, message: '请选择定时发布时间' }]">
              <a-date-picker
                v-model:value="settingsForm.publishAt"
                class="w-260px"
                :show-time="{ format: 'HH:mm' }"
                value-format="YYYY-MM-DD HH:mm"
                format="YYYY-MM-DD HH:mm"
                placeholder="请选择日期时间"
                :disabled-date="disabledDate"
              />
            </a-form-item>

            <a-form-item label="需家长确认" name="isConfirm">
              <div class="flex items-center">
                <a-switch v-model:checked="settingsForm.isConfirm" />
                <span class="ml-5px text-13px text-gray">开启后，家长需点击按钮来确认收到通知</span>
              </div>
            </a-form-item>
          </a-form>
        </div>
      </div>

      <template #footer>
        <div class="flex justify-end">
          <a-button style="font-size: 20px; height: 48px;min-width: 140px;" size="large" type="primary" :loading="submitting" @click="submitNotice()">
            发布
          </a-button>
        </div>
      </template>
    </a-drawer>

    <NoticeScopePickerModel
      v-model="scopePickerOpen"
      :selected-students="formState.selectedStudents"
      :selected-sources="formState.selectedSources"
      @complete="handleScopePickerComplete"
    />

    <NoticeStudentPickerModel
      v-model="studentPickerOpen"
      :selected-students="formState.selectedStudents"
      :selected-sources="formState.selectedSources"
      @complete="handleStudentPickerComplete"
    />

    <NoticePhonePreviewModal
      v-model="previewOpen"
      :title="formState.title"
      :content-html="formState.content"
      :cover-url="props.selectedTemplate?.coverUrl || ''"
      :primary-text="previewMetaPrimary"
      :publish-text="previewPublishText"
    />
  </div>
</template>

<style scoped lang="less">
.noticeModel {
  padding-bottom: 0;
  text-align: left;
  top: 35px !important;
  vertical-align: middle;
}

.close-btn {
  &:hover {
    background: transparent;
  }
}

.rulesWrapper {
  width: 100%;
  min-height: 430px;

  .rulesTitle {
    font-weight: 500;
    font-size: 14px;
    color: #222;
    line-height: 22px;
    margin-bottom: 8px;
  }

  .rulesContent {
    font-weight: 400;
    font-size: 14px;
    color: #666;
    line-height: 22px;
  }
}

.notice-scope-selector {
  display: flex;
  align-items: center;
  gap: 10px;
}

.notice-scope-preview {
  margin-top: 8px;
  width: 100%;
  min-height: 42px;
  padding: 9px 14px;
  border: 1px solid #d9d9d9;
  border-radius: 8px;
  background: #fafafa;
  color: #606266;
  font-size: 14px;
  line-height: 20px;
  word-break: break-all;
  cursor: pointer;
}

::v-deep(.ql-toolbar) {
  border-radius: 8px 8px 0 0;
}

::v-deep(.ql-container) {
  border-radius: 0 0 8px 8px;
}
</style>
