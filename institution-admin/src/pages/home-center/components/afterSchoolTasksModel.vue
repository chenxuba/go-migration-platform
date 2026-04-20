<script setup>
import { PictureOutlined, PlayCircleOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'

const props = defineProps({
  title: {
    type: String,
    default: '新建课后任务',
  },
})
const open = defineModel({
  type: Boolean,
  default: false,
})

const confirmLoading = ref(false)
const formRef = ref(null)
const formState = reactive({
  title: '',
  content: '',
  rule: 1,
  students: [],
  publishAt: undefined,
  deadlineAt: undefined,
  dateRange: [],
  time: undefined,
  weeks: [],
  imgList: [],
  videoList: [],
  audioList: [],
})

const activeFile = ref(undefined)

const weeks = [{ label: '星期一', value: 1 }, { label: '星期二', value: 2 }, { label: '星期三', value: 3 }, { label: '星期四', value: 4 }, { label: '星期五', value: 5 }, { label: '星期六', value: 6 }, { label: '星期日', value: 7 }]
const dateOptions = [{ label: '00:00', value: '00:00' }, { label: '01:00', value: '01:00' }, { label: '02:00', value: '02:00' }, { label: '03:00', value: '03:00' }, { label: '04:00', value: '04:00' }, { label: '05:00', value: '05:00' }, { label: '06:00', value: '06:00' }, { label: '07:00', value: '07:00' }, { label: '08:00', value: '08:00' }, { label: '09:00', value: '09:00' }, { label: '10:00', value: '10:00' }, { label: '11:00', value: '11:00' }, { label: '12:00', value: '12:00' }, { label: '13:00', value: '13:00' }, { label: '14:00', value: '14:00' }, { label: '15:00', value: '15:00' }, { label: '16:00', value: '16:00' }, { label: '17:00', value: '17:00' }, { label: '18:00', value: '18:00' }, { label: '19:00', value: '19:00' }, { label: '20:00', value: '20:00' }, { label: '21:00', value: '21:00' }, { label: '22:00', value: '22:00' }, { label: '23:00', value: '23:00' }]

function handleWeek(value) {
  const index = formState.weeks.indexOf(value)
  if (index === -1) {
    formState.weeks.push(value)
  }
  else {
    formState.weeks.splice(index, 1)
  }
}

// 图片预览
function handlePreview(file) {
  console.log(file)
}

// 鼠标悬停样式处理
function handleOpenChange(value, show) {
  if (show) {
    activeFile.value = value
  }
  else {
    activeFile.value = undefined
  }
}

function handleOk() {
  formRef.value.validate().then(() => {
    console.log('验证通过')
  })
}

function buildTimeRange(start, end) {
  return Array.from({ length: Math.max(end - start, 0) }, (_, index) => start + index)
}

function getPublishAtDate() {
  if (!formState.publishAt)
    return null
  const publishAt = dayjs(formState.publishAt)
  return publishAt.isValid() ? publishAt : null
}

// 禁用今天之前的日期
function disabledDate(current) {
  return current && current < dayjs().startOf('day')
}

function disabledDeadlineDate(current) {
  if (!current)
    return false

  const publishAt = getPublishAtDate()
  if (publishAt)
    return !dayjs(current).isAfter(publishAt, 'day')

  return dayjs(current).isBefore(dayjs().startOf('day'), 'day')
}

function disabledDeadlineTime(current) {
  const publishAt = getPublishAtDate()
  if (!publishAt || !current || !dayjs(current).isSame(publishAt, 'day'))
    return {}

  return {
    disabledHours: () => buildTimeRange(0, publishAt.hour()),
    disabledMinutes: selectedHour => selectedHour === publishAt.hour() ? buildTimeRange(0, publishAt.minute()) : [],
  }
}

watch(() => formState.publishAt, (publishAtValue) => {
  if (!publishAtValue || !formState.deadlineAt)
    return

  const publishAt = dayjs(publishAtValue)
  const deadlineAt = dayjs(formState.deadlineAt)
  if (publishAt.isValid() && deadlineAt.isValid() && !deadlineAt.isAfter(publishAt, 'day'))
    formState.deadlineAt = undefined
})
</script>

<template>
  <div>
    <a-modal
      v-model:open="open"
      centered
      class="afterSchoolTasksModel"
      :body-style="{ height: '580px', overflowY: 'auto' }"
      width="800px"
      :title="props.title"
      destroy-on-close
      @ok="handleOk"
    >
      <a-form ref="formRef" layout="vertical" :model="formState" v-bind="formItemLayout">
        <a-form-item label="任务标题" name="title" :rules="[{ required: true, message: '请输入任务标题' }]">
          <a-input v-model:value="formState.title" :maxlength="20" placeholder="请输入任务标题，最多20字" />
        </a-form-item>
        <a-form-item
          label="任务内容"
          name="content"
          class="afterSchoolTasksModel__content-item"
          :rules="[{ required: true, message: '请输入任务内容' }]"
        >
          <a-textarea
            class="afterSchoolTasksModel__content-textarea"
            v-model:value="formState.content"
            :show-count="true"
            style="height: 66px; min-height: 66px;"
            :maxlength="2000"
            placeholder="请输入任务内容，最多2000字"
            :auto-size="{ minRows: 4, maxRows: 4 }"
          />
        </a-form-item>

        <a-form-item class="afterSchoolTasksModel__upload-form-item">
          <div class="afterSchoolTasksModel__upload-actions flex flex-wrap items-center gap-8px">
            <a-upload
              v-model:file-list="formState.imgList"
              class="afterSchoolTasksModel__upload"
              list-type="picture-card"
              :max-count="12"
              action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
            >
              <a-tooltip placement="right" @open-change="(show) => handleOpenChange(1, show)">
                <template #title>
                  限制单张 9 M
                </template>
                <div
                  :class="{ 'bg-#06f!important': activeFile === 1, 'text-#fff!important': activeFile === 1 }"
                  class="w-135px cursor-pointer flex items-center gap-5px bg-#f6f7f8 px-10px py-3px rounded-12px"
                >
                  <PictureOutlined :class="activeFile === 1 ? 'text-#fff' : 'text-#06f'" />
                  <span>添加图片({{ formState.imgList.length }}/12)</span>
                </div>
              </a-tooltip>
            </a-upload>

            <a-upload
              v-model:file-list="formState.videoList"
              class="afterSchoolTasksModel__upload"
              list-type="picture-card"
              :max-count="9"
              action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
            >
              <a-tooltip placement="right" @open-change="(show) => handleOpenChange(2, show)">
                <template #title>
                  限制每个 500 M
                </template>
                <div
                  :class="{ 'bg-#06f!important': activeFile === 2, 'text-#fff!important': activeFile === 2 }"
                  class="w-135px cursor-pointer flex items-center gap-5px bg-#f6f7f8 px-10px py-3px rounded-12px"
                >
                  <PlayCircleOutlined :class="activeFile === 2 ? 'text-#fff' : 'text-#06f'" />
                  <span>添加视频({{ formState.videoList.length }}/9)</span>
                </div>
              </a-tooltip>
            </a-upload>
            <!-- <a-upload
              v-model:file-list="formState.audioList" list-type="picture-card" :max-count="12"
              action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
            >
              <a-tooltip placement="right" @open-change="(show) => handleOpenChange(3, show)">
                <template #title>
                  限制每个 10 M
                </template>
                <div
                  :class="{ 'bg-#06f!important': activeFile === 3, 'text-#fff!important': activeFile === 3 }"
                  class="w-135px cursor-pointer flex items-center gap-5px bg-#f6f7f8 px-10px py-3px rounded-12px"
                >
                  <AudioOutlined :class="activeFile === 3 ? 'text-#fff' : 'text-#06f'" />
                  <span>添加音频(0/10)</span>
                </div>
              </a-tooltip>
            </a-upload> -->
          </div>
        </a-form-item>

        <a-form-item label="选择班级/学员" name="students" :rules="[{ required: true, message: '情选择班级/学员' }]">
          <a-button type="primary" ghost>
            选择班级/学员
          </a-button>
        </a-form-item>
        <a-form-item label="发布规则" :required="true">
          <a-radio-group v-model:value="formState.rule" class="custom-radio">
            <a-radio :value="1">
              仅本次发布
            </a-radio>
            <a-radio :value="2">
              设置自动任务
            </a-radio>
          </a-radio-group>
        </a-form-item>
        <a-form-item v-if="formState.rule === 1">
          <div class="afterSchoolTasksModel__rule-card">
            <div class="afterSchoolTasksModel__rule-card-title">
              设置本次发布时间（非必填）
            </div>
            <a-row :gutter="[20, 16]">
              <a-col :xs="24" :sm="12">
                <div class="afterSchoolTasksModel__rule-field">
                  <div class="afterSchoolTasksModel__rule-label">
                    <span>定时发布日期</span>
                    <a-popover
                      color="#fff"
                      placement="topLeft"
                      title="定时发布日期"
                    >
                      <template #content>
                        <div class="afterSchoolTasksModel__rule-popover">
                          设置后，任务创建完成，会按设置的时间点发送，如果不设置，任务创建完成会立即发送
                        </div>
                      </template>
                      <QuestionCircleOutlined class="afterSchoolTasksModel__rule-tip" />
                    </a-popover>
                    <span>:</span>
                  </div>
                  <a-date-picker
                    v-model:value="formState.publishAt"
                    class="w-full"
                    :show-time="{ format: 'HH:mm' }"
                    value-format="YYYY-MM-DD HH:mm"
                    format="YYYY-MM-DD HH:mm"
                    placeholder="请选择日期时间"
                    :disabled-date="disabledDate"
                  />
                </div>
              </a-col>
              <a-col :xs="24" :sm="12">
                <div class="afterSchoolTasksModel__rule-field">
                  <div class="afterSchoolTasksModel__rule-label">
                    <span>设置任务截止日期</span>
                    <a-popover
                      color="#fff"
                      placement="topLeft"
                      title="任务截止日期"
                    >
                      <template #content>
                        <div class="afterSchoolTasksModel__rule-popover">
                          设置后，学员仍可以上传任务，超时提交的学员，在系统上为学员打上超时提交标签
                        </div>
                      </template>
                      <QuestionCircleOutlined class="afterSchoolTasksModel__rule-tip" />
                    </a-popover>
                    <span>:</span>
                  </div>
                  <a-date-picker
                    v-model:value="formState.deadlineAt"
                    class="w-full"
                    :show-time="{ format: 'HH:mm' }"
                    value-format="YYYY-MM-DD HH:mm"
                    format="YYYY-MM-DD HH:mm"
                    placeholder="请选择日期时间"
                    :disabled-date="disabledDeadlineDate"
                    :disabled-time="disabledDeadlineTime"
                  />
                </div>
              </a-col>
            </a-row>
          </div>
        </a-form-item>
        <a-form-item v-if="formState.rule === 2">
          <div class="border border-gray-200 rounded-8px border-solid">
            <div class="bg-#fafafa  px-15px py-10px">
              设置自动任务周期
            </div>
            <div class="flex items-center gap-30px p-15px">
              <div v-for="(week, index) in weeks" :key="index" class="flex flex-col items-center gap-5px">
                <div class="text-#888 text-12px">
                  {{ week.label }}
                </div>
                <div
                  class="week-day" :class="{ 'week-active': formState.weeks.includes(week.value) }"
                  @click="handleWeek(week.value)"
                />
              </div>
            </div>
            <div class="flex items-center  justify-between gap-10px p-15px">
              <a-form-item label="任务日期范围" name="dateRange" :rules="[{ required: true, message: '请选择周期' }]">
                <a-range-picker v-model:value="formState.dateRange" :disabled-date="disabledDate" />
              </a-form-item>
              <a-form-item
                class=" w-340px" label="任务推送时间：" name="time"
                :rules="[{ required: true, message: '请选择任务推送时间' }]"
              >
                <a-select v-model:value="formState.time" placeholder="请选择" :options="dateOptions" />
              </a-form-item>
            </div>
          </div>
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<style>
.afterSchoolTasksModel {
  padding-bottom: 0;
  text-align: left;
}
</style>

<style scoped lang="less">
.afterSchoolTasksModel__upload-form-item {
  margin-top: -14px;
  margin-bottom: 12px;
}

.afterSchoolTasksModel__content-item {
  margin-bottom: 10px;
}

.afterSchoolTasksModel__upload-actions {
  align-items: flex-start;
}

.afterSchoolTasksModel__upload {
  display: inline-flex;
  width: auto;
  flex: none;
  margin-bottom: 0;
}

.afterSchoolTasksModel__rule-card {
  border: 1px solid #e8e8e8;
  border-radius: 10px;
  background: #fff;
  padding: 16px;
}

.afterSchoolTasksModel__rule-card-title {
  margin-bottom: 18px;
  color: #262626;
  font-size: 14px;
  line-height: 22px;
  font-weight: 500;
}

.afterSchoolTasksModel__rule-field {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.afterSchoolTasksModel__rule-label {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  color: #262626;
  font-size: 14px;
  line-height: 22px;
}

.afterSchoolTasksModel__rule-tip {
  color: #999;
  font-size: 14px;
  cursor: pointer;
}

.afterSchoolTasksModel__rule-popover {
  max-width: 360px;
  color: #595959;
  font-size: 14px;
  line-height: 22px;
}

:deep(.afterSchoolTasksModel__content-item .ant-input-textarea-show-count::after) {
  margin-top: 2px;
}

:deep(.afterSchoolTasksModel__content-item .ant-form-item-control-input + div) {
  min-height: auto;
  margin-top: -24px;
  margin-bottom: 5px;
}

:deep(.afterSchoolTasksModel__content-textarea.ant-input-textarea-show-count::after) {
  margin-top: 2px;
}

.week-day {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  background-color: #eee;
  background-image: url('https://pcsys.admin.ybc365.com/64344ed6-b8db-43a2-8488-4c18a6095a50.png');
  background-repeat: no-repeat;
  background-position: center;
  background-size: 24px;
  cursor: pointer;
}

.week-active {
  background-color: #06f;
}

::v-deep(.ant-upload-select) {
  border: none !important;
  flex: 1;
  width: 135px !important;
  height: 100% !important;
  display: block;
}

::v-deep(.afterSchoolTasksModel__upload .ant-upload-list) {
  display: inline-flex;
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 12px;
}

::v-deep(.ant-upload-list-item-container) {
  width: 80px !important;
  height: 80px !important;
}
</style>
