<script setup>
import { PictureOutlined, PlayCircleOutlined } from '@ant-design/icons-vue'
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
// 禁用今天之前的日期
function disabledDate(current) {
  return current && current < dayjs().startOf('day')
}
</script>

<template>
  <div>
    <a-modal
      v-model:open="open" class="afterSchoolTasksModel" :body-style="{ height: '580px', overflowY: 'auto' }"
      width="800px" :title="props.title" destroy-on-close @ok="handleOk"
    >
      <a-form ref="formRef" layout="vertical" :model="formState" v-bind="formItemLayout">
        <a-form-item label="任务标题" name="title" :rules="[{ required: true, message: '请输入任务标题' }]">
          <a-input v-model:value="formState.title" :maxlength="20" placeholder="请输入任务标题，最多20字" />
        </a-form-item>
        <a-form-item label="任务内容" name="content" :rules="[{ required: true, message: '请输入任务内容' }]">
          <a-textarea
            v-model:value="formState.content" :show-count="true" style="height: 98px;min-height: 98px;"
            :maxlength="2000" placeholder="请输入任务内容，最多2000字" :auto-size="{ minRows: 6, maxRows: 6 }"
          />
        </a-form-item>

        <a-form-item>
          <div class="flex flex-col gap-8px">
            <a-upload
              v-model:file-list="formState.imgList" list-type="picture-card" :max-count="12"
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
                  <span>添加图片(0/12)</span>
                </div>
              </a-tooltip>
            </a-upload>

            <a-upload
              v-model:file-list="formState.videoList" list-type="picture-card" :max-count="12"
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
                  <span>添加视频(0/9)</span>
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
          <a-radio-group v-model:value="formState.rule">
            <a-radio :value="1">
              仅本次发布
            </a-radio>
            <a-radio :value="2">
              设置自动任务
            </a-radio>
          </a-radio-group>
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
  /* display: inline-block; */
  padding-bottom: 0;
  text-align: left;
  top: 10px !important;
  vertical-align: middle;
}
</style>

<style scoped lang="less">
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
  // width: 100% !important;
  flex: 1;
  width: 135px !important;
  height: 100% !important;
  display: block;
}

::v-deep(.ant-upload-list-item-container) {
  width: 80px !important;
  height: 80px !important;
}
</style>
