<script setup>
const props = defineProps({
  spinning: {
    type: Boolean,
    default: false,
  },
  tableDataSource: {
    type: Array,
    default: () => [],
  },
  columns: {
    type: Array,
    default: () => [],
  },
  isSwapTimeGrid: {
    type: Boolean,
    default: false,
  },
  focusedScheduleCellKey: {
    type: String,
    default: '',
  },
  draggingScheduleCellKey: {
    type: String,
    default: '',
  },
  isScheduleColumn: {
    type: Function,
    required: true,
  },
  scheduleCellKey: {
    type: Function,
    required: true,
  },
  scheduleCellStartTime: {
    type: Function,
    required: true,
  },
  scheduleCellEndTime: {
    type: Function,
    required: true,
  },
  scheduleCellContextColumn: {
    type: Function,
    required: true,
  },
  scheduleCellContextRecord: {
    type: Function,
    required: true,
  },
  openScheduledLessonDetail: {
    type: Function,
    required: true,
  },
  openScheduledConflictDetail: {
    type: Function,
    required: true,
  },
  handleConflictClick: {
    type: Function,
    required: true,
  },
  handleScheduleClick: {
    type: Function,
    required: true,
  },
  handleSchedulePointerDown: {
    type: Function,
    required: true,
  },
  isScheduleDraggable: {
    type: Function,
    required: true,
  },
  draggingScheduleStyle: {
    type: Object,
    default: () => ({}),
  },
  emptyLessonDragState: {
    type: Function,
    required: true,
  },
  emptyLessonStatusText: {
    type: Function,
    required: true,
  },
  teacherLessonCountLabel: {
    type: Function,
    required: true,
  },
  formatWeek: {
    type: Function,
    required: true,
  },
  formatDate: {
    type: Function,
    required: true,
  },
})
</script>

<template>
  <a-spin :spinning="spinning">
    <a-table
      :scroll="{ x: 1300 }"
      :sticky="{ offsetHeader: 100 }"
      size="small"
      :pagination="false"
      bordered
      :data-source="tableDataSource"
      :columns="columns"
    >
      <template #headerCell="{ column }">
        <template v-if="!isSwapTimeGrid && column.startTime && column.endTime">
          <div>{{ column.title }}</div>
          <div class="text-12px text-#666 line-height-2">
            {{ column.startTime }}-{{ column.endTime }}
          </div>
        </template>
        <template v-else-if="isSwapTimeGrid && column.date">
          <div>{{ column.title }}</div>
          <div class="text-12px text-#666 line-height-2">
            {{ column.dateText }}
          </div>
        </template>
        <template v-else>
          {{ column.title }}
        </template>
      </template>

      <template #bodyCell="{ column, record, text }">
        <template v-if="isScheduleColumn(column)">
          <div
            v-if="text.studentId"
            :data-schedule-cell-key="scheduleCellKey(column, record)"
            class="st-schedule-cell flex flex-col bg-#4e6dff1f h-11 rounded-1 text-3 text-#fff cursor-pointer"
            :class="{
              'st-schedule-cell--focused': focusedScheduleCellKey === scheduleCellKey(column, record),
              'st-schedule-cell--conflict': text.scheduledConflict,
              'st-schedule-cell--draggable': isScheduleDraggable(text),
              'st-schedule-cell--drag-disabled': !isScheduleDraggable(text) && text.courseType === 1 && text.isMain === false,
              'st-schedule-cell--floating': draggingScheduleCellKey === scheduleCellKey(column, record),
              'st-schedule-cell--dragging': draggingScheduleCellKey === scheduleCellKey(column, record),
            }"
            :title="!isScheduleDraggable(text) && text.courseType === 1 && text.isMain === false ? '助教课暂不支持拖拽调课，请在主教老师所在行操作' : undefined"
            :style="draggingScheduleCellKey === scheduleCellKey(column, record) ? draggingScheduleStyle : undefined"
            @click="openScheduledLessonDetail(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record))"
            @mousedown.left="handleSchedulePointerDown($event, text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record))"
          >
            <div class="pl1 bg-#06f rounded-1 rounded-lb-0 rounded-rb-0 flex relative h-5">
              {{ scheduleCellStartTime(column, record) }}-{{ scheduleCellEndTime(column, record) }}
              <span
                class="absolute right-0 pl-2 pr-1 h-4 text-#fff text-2.5 font-500 rounded-rt-1 rounded-lb-2"
                :class="text.scheduledConflict ? 'st-schedule-cell__badge--conflict' : 'bg-#00000080'"
                :style="{ cursor: text.scheduledConflict ? 'pointer' : 'default' }"
                @click.stop="text.scheduledConflict ? openScheduledConflictDetail(text) : undefined"
              >
                <span v-if="text.scheduledConflict">冲突</span>
                <span v-else-if="text.courseType === 1">1v1</span>
                <span v-else-if="text.courseType === 2">班课(<span>{{ text.isMain ? '主教' : '辅教' }}</span>)</span>
              </span>
            </div>

            <div v-if="!text.classId" class="flex pl-1 flex-1 text-#002cfd flex-items-center">
              <span v-for="(item, index) in text.studentNames" :key="index">
                <div class="flex">{{ item.name }}{{ index !== text.studentNames.length - 1 ? '、' : '' }}-{{ text.courseName }}</div>
              </span>
            </div>

            <div v-else class="flex pl-1 flex-1 text-#002cfd line-height-4 flex-items-center">
              <div class="flex">
                {{ text.className }}-{{ text.courseName }}
              </div>
            </div>
          </div>

          <div
            v-else
            :data-empty-schedule-cell-key="scheduleCellKey(column, record)"
            :data-drag-target-teacher-id="record?.teacherId || ''"
            :data-drag-target-teacher-name="record?.name || ''"
            :data-drag-target-lesson-date="scheduleCellContextRecord(column, record)?.date || ''"
            :data-drag-target-start-time="scheduleCellContextColumn(column, record)?.startTime || ''"
            :data-drag-target-end-time="scheduleCellContextColumn(column, record)?.endTime || ''"
            class="st-empty-cell h-11 rounded-1 text-3 flex-center cursor-pointer"
            :title="emptyLessonDragState(column, record)?.message || undefined"
            :class="[
              emptyLessonDragState(column, record)?.checking
                ? 'st-empty-cell--drag-checking'
                : emptyLessonDragState(column, record)?.valid === true
                  ? 'st-empty-cell--drag-valid'
                  : emptyLessonDragState(column, record)?.valid === false
                    ? 'st-empty-cell--drag-invalid'
                    : (text.conflict ? 'bg-#ffe6e6 text-#a31616' : 'bg-#e6ffe6 text-#16a34a'),
            ]"
            @click="text.conflict ? handleConflictClick(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record)) : handleScheduleClick(text, scheduleCellContextColumn(column, record), scheduleCellContextRecord(column, record))"
          >
            {{ emptyLessonDragState(column, record)?.label || emptyLessonStatusText(text) }}
          </div>
        </template>

        <template v-if="column.key === 'date'">
          <div class="text-3.5 ">
            {{ formatWeek(text) }}
          </div>
          <div class="text-3 font-500 line-height-3 text-#666">
            {{ formatDate(text) }}
          </div>
        </template>

        <template v-if="column.key === 'name'">
          <div>{{ text }}</div>
          <div class="text-3 text-#666 leading-snug">
            {{ teacherLessonCountLabel(record.teacherId) }}
          </div>
        </template>

        <template v-if="column.key === 'slot'">
          <div class="text-3.5">
            {{ text }}
          </div>
          <div class="text-3 font-500 line-height-3 text-#666">
            {{ record.startTime }}-{{ record.endTime }}
          </div>
        </template>
      </template>
    </a-table>
  </a-spin>
</template>

<style scoped lang="less">
:deep(td.ant-table-cell.ant-table-cell-row-hover) {
  background-color: rgb(231, 236, 255) !important;
}

:deep(td.ant-table-cell) {
  padding: 4px !important;
}

.st-schedule-cell {
  position: relative;
  transition: box-shadow 0.25s ease, transform 0.25s ease, opacity 0.2s ease;
}

.st-schedule-cell--conflict {
  box-shadow: 0 0 0 2px rgba(255, 77, 79, 0.4);
}

.st-schedule-cell--draggable {
  cursor: grab;
}

.st-schedule-cell--drag-disabled {
  cursor: not-allowed;
}

.st-schedule-cell--dragging {
  cursor: grabbing;
}

.st-schedule-cell--floating {
  margin: 0 !important;
  pointer-events: none;
  box-shadow:
    0 18px 40px rgba(31, 35, 41, 0.2),
    0 8px 18px rgba(31, 35, 41, 0.12);
  transform: rotate(-1deg);
  z-index: 1200;
}

.st-schedule-cell__badge--conflict {
  background: #ff4d4f;
}

.st-schedule-cell--focused {
  animation: st-schedule-cell-flash 0.5s ease-in-out 6;
  box-shadow:
    0 0 0 3px rgba(255, 77, 79, 0.98),
    0 0 0 8px rgba(255, 77, 79, 0.26),
    0 0 20px rgba(255, 77, 79, 0.5);
  transform: scale(1.015);
  z-index: 2;
}

.st-empty-cell {
  transition:
    background-color 0.18s ease,
    color 0.18s ease,
    box-shadow 0.18s ease,
    transform 0.18s ease;
}

.st-empty-cell--drag-checking {
  background: #fff7e6;
  color: #d48806;
  box-shadow: inset 0 0 0 1px rgba(250, 173, 20, 0.28);
}

.st-empty-cell--drag-valid {
  background: #fff7e6;
  color: #d48806;
  box-shadow:
    inset 0 0 0 1px rgba(250, 173, 20, 0.34),
    0 0 0 2px rgba(250, 173, 20, 0.12);
}

.st-empty-cell--drag-invalid {
  background: #ffe1df;
  color: #cf1322;
  box-shadow:
    inset 0 0 0 1px rgba(255, 77, 79, 0.4),
    0 0 0 2px rgba(255, 77, 79, 0.16);
}

@keyframes st-schedule-cell-flash {
  0%,
  100% {
    box-shadow:
      0 0 0 3px rgba(255, 77, 79, 0.98),
      0 0 0 8px rgba(255, 77, 79, 0.26),
      0 0 20px rgba(255, 77, 79, 0.5);
  }

  50% {
    box-shadow:
      0 0 0 3px rgba(255, 77, 79, 0.25),
      0 0 0 8px rgba(255, 77, 79, 0.08),
      0 0 10px rgba(255, 77, 79, 0.18);
  }
}
</style>
