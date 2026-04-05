<script setup>
import { LeftOutlined, RightOutlined } from '@ant-design/icons-vue'
import { computed } from 'vue'
import CreateSchedulePopover from './create-schedule-popover.vue'

const props = defineProps({
  currentModel: {
    type: String,
    required: true,
  },
  oneToOneRecordId: {
    type: [String, Number],
    default: undefined,
  },
  oneToOnePickerOpen: {
    type: Boolean,
    default: false,
  },
  oneToOneListLoading: {
    type: Boolean,
    default: false,
  },
  oneToOneData: {
    type: Array,
    default: () => [],
  },
  renderOneToOneDropdown: {
    type: Function,
    required: true,
  },
  filterOneToOneOption: {
    type: Function,
    required: true,
  },
  classId: {
    type: [String, Number],
    default: null,
  },
  classData: {
    type: Array,
    default: () => [],
  },
  currentTime: {
    type: String,
    required: true,
  },
  timeViewOptions: {
    type: Array,
    default: () => [],
  },
  currentWeek: {
    type: Object,
    required: true,
  },
  formatDateRange: {
    type: Function,
    required: true,
  },
  isWeekLikeView: {
    type: Boolean,
    default: false,
  },
  currentGroup: {
    type: String,
    required: true,
  },
  groupOptions: {
    type: Array,
    default: () => [],
  },
  onOneToOneChange: {
    type: Function,
    required: true,
  },
  onOneToOneDropdownVisibleChange: {
    type: Function,
    required: true,
  },
  onClassChange: {
    type: Function,
    required: true,
  },
  onPrev: {
    type: Function,
    required: true,
  },
  onNext: {
    type: Function,
    required: true,
  },
  onThisWeek: {
    type: Function,
    required: true,
  },
})

const emit = defineEmits([
  'update:currentModel',
  'update:oneToOneRecordId',
  'update:oneToOnePickerOpen',
  'update:classId',
  'update:currentTime',
  'update:currentWeek',
  'update:currentGroup',
])

const currentModelValue = computed({
  get: () => props.currentModel,
  set: value => emit('update:currentModel', value),
})

const oneToOneValue = computed({
  get: () => props.oneToOneRecordId,
  set: value => emit('update:oneToOneRecordId', value),
})

const oneToOneOpenValue = computed({
  get: () => props.oneToOnePickerOpen,
  set: value => emit('update:oneToOnePickerOpen', value),
})

const classIdValue = computed({
  get: () => props.classId,
  set: value => emit('update:classId', value),
})

const currentTimeValue = computed({
  get: () => props.currentTime,
  set: value => emit('update:currentTime', value),
})

const currentWeekValue = computed({
  get: () => props.currentWeek,
  set: value => emit('update:currentWeek', value),
})

const currentGroupValue = computed({
  get: () => props.currentGroup,
  set: value => emit('update:currentGroup', value),
})
</script>

<template>
  <div class="time-template mt2 bg-white py3 px5 rounded-4 rounded-lb-0 rounded-rb-0">
    <div class="top-filter st-top-filter-bar flex flex-nowrap items-center gap-1 overflow-x-auto">
      <div class="shrink-0">
        <a-radio-group v-model:value="currentModelValue" button-style="solid">
          <a-radio-button value="1">
            1v1
          </a-radio-button>
          <a-radio-button value="2">
            班课
          </a-radio-button>
        </a-radio-group>
      </div>

      <div class="shrink-0">
        <div v-if="currentModel === '1'" class="flex items-center shrink-0 gap-1">
          <span class="whitespace-nowrap w-71px text-right">选择1v1：</span>
          <a-select
            v-model:value="oneToOneValue"
            v-model:open="oneToOneOpenValue"
            allow-clear
            show-search
            :loading="oneToOneListLoading"
            :dropdown-match-select-width="false"
            :dropdown-style="{ width: '520px' }"
            :dropdown-render="renderOneToOneDropdown"
            :filter-option="filterOneToOneOption"
            placeholder="搜索/选择"
            class="st-top-1v1-select"
            popup-class-name="st-top-1v1-select-dropdown"
            option-label-prop="label"
            @dropdown-visible-change="onOneToOneDropdownVisibleChange"
            @change="onOneToOneChange"
          >
            <a-select-option
              v-for="item in oneToOneData"
              :key="item.id"
              :value="item.id"
              :label="item.name"
            >
              <div>{{ item.name }}</div>
            </a-select-option>
          </a-select>
        </div>

        <div v-if="currentModel === '2'" class="flex items-center">
          <span class="w-75px">选择班级：</span>
          <a-select
            v-model:value="classIdValue"
            allow-clear
            placeholder="请搜索/选择班级"
            class="st-top-class-select"
            option-label-prop="label"
            @change="onClassChange"
          >
            <a-select-option
              v-for="item in classData"
              :key="item.id"
              :value="item.id"
              :data="item"
              :label="item.name"
            >
              <div>{{ item.name }}</div>
              <div class="text-3 text-#666">
                主教：{{ item.mainTeacherName }}
              </div>
            </a-select-option>
          </a-select>
        </div>
      </div>

      <div class="time-selector flex items-center shrink-0 st-time-selector--after-filters">
        <a-select
          v-model:value="currentTimeValue"
          :options="timeViewOptions"
          class="st-time-view-select"
        />
        <div
          class="text-#0061ff font-800 text-5 flex items-center shrink-0 st-date-nav"
          :class="
            currentTime === 'day'
              ? 'st-date-nav--day'
              : isWeekLikeView
                ? 'st-date-nav--week'
                : 'st-date-nav--month'
          "
        >
          <a-popover trigger="hover">
            <template #content>
              {{ currentTime === 'day' ? '前一天' : isWeekLikeView ? '上一周' : '上个月' }}
            </template>
            <span
              class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 shrink-0 hover-text-#06f hover-bg-#e6f0ff"
              @click="onPrev"
            >
              <LeftOutlined />
            </span>
          </a-popover>
          <span class="mx-1 min-w-0 flex-1 st-date-nav__mid">
            <div class="relative cursor-pointer whitespace-nowrap text-center st-date-nav__text">
              {{ formatDateRange(currentWeek) }}
              <a-date-picker
                v-if="currentTime === 'day'"
                v-model:value="currentWeekValue"
                class="absolute top-0 left-0 right-0 bottom-0 z-10 opacity-0"
                :allow-clear="false"
                :bordered="false"
                :format="formatDateRange"
                style="cursor:pointer;"
              />
              <a-date-picker
                v-else-if="isWeekLikeView"
                v-model:value="currentWeekValue"
                class="absolute top-0 left-0 right-0 bottom-0 z-10 opacity-0"
                picker="week"
                :allow-clear="false"
                :bordered="false"
                :format="formatDateRange"
                style="cursor:pointer;"
              />
              <a-date-picker
                v-else
                v-model:value="currentWeekValue"
                class="absolute top-0 left-0 right-0 bottom-0 z-10 opacity-0"
                picker="month"
                :allow-clear="false"
                :bordered="false"
                :format="formatDateRange"
                style="cursor:pointer;"
              />
            </div>
          </span>
          <a-popover trigger="hover">
            <template #content>
              {{ currentTime === 'day' ? '后一天' : isWeekLikeView ? '下一周' : '下个月' }}
            </template>
            <span
              class="cursor-pointer text-3 text-#888 flex w6 h6 bg-#eee rounded-10 flex-center font-500 shrink-0 hover-text-#06f hover-bg-#e6f0ff"
              @click="onNext"
            >
              <RightOutlined />
            </span>
          </a-popover>
        </div>
        <a-button size="small" class="shrink-0 st-this-week-btn" @click="onThisWeek">
          本周
        </a-button>
      </div>

      <div class="ml-auto flex shrink-0 items-center gap-2">
        <a-radio-group v-model:value="currentGroupValue" button-style="solid">
          <a-radio-button v-for="opt in groupOptions" :key="opt.key" :value="opt.key">
            {{ opt.label }}
          </a-radio-button>
        </a-radio-group>
        <a-space>
          <CreateSchedulePopover />
          <a-button>导出课表</a-button>
        </a-space>
      </div>
    </div>
  </div>
</template>

<style scoped lang="less">
.st-top-1v1-select {
  width: 180px;
  max-width: 180px;
}

.st-top-class-select {
  width: 180px;
  max-width: 180px;
}

.st-time-view-select {
  width: 112px;
  min-width: 112px;
  flex-shrink: 0;
}

.st-date-nav {
  box-sizing: border-box;
}

.st-date-nav--day {
  width: 300px;
  min-width: 300px;
  max-width: 300px;
}

.st-date-nav--week {
  width: 300px;
  min-width: 300px;
  max-width: 300px;
}

.st-date-nav--month {
  width: 180px;
  min-width: 180px;
  max-width: 180px;
}

.st-date-nav__mid {
  overflow: hidden;
}

.st-date-nav__text {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
}

.st-top-filter-bar {
  scrollbar-width: thin;
  -webkit-overflow-scrolling: touch;
}

.time-selector {
  font-family: DINAlternate-Bold, DINAlternate;
  gap: 6px;

  .ant-radio-button-wrapper {
    padding: 0 16px;
  }
}

.st-time-selector--after-filters {
  margin-left: 8px;
}

.st-this-week-btn {
  padding: 0 10px;
  height: 28px;
  line-height: 26px;
  border-radius: 8px;
}
</style>
