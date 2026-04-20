<script setup>
import { computed } from 'vue'

const defaultScheduleLegend = [
  {
    key: 'unsigned',
    label: '未点名（教师/课程）',
    type: 'bar',
    color: 'linear-gradient(90deg, #39b8ff 0%, #6c5cff 50%, #74d87f 100%)',
  },
  {
    key: 'signed',
    label: '已点名',
    type: 'bar',
    color: '#b7bec8',
  },
  {
    key: 'partial',
    label: '部分点名',
    type: 'bar',
    color: '#f59e0b',
  },
  {
    key: 'trial',
    label: '含试听学员',
    type: 'icon',
  },
  {
    key: 'conflict',
    label: '日程冲突',
    type: 'icon-danger',
  },
]

const props = defineProps({
  total: {
    type: Number,
    default: 0,
  },
  unsignedCount: {
    type: Number,
    default: 0,
  },
  totalUnit: {
    type: String,
    default: '个日程',
  },
  unsignedUnit: {
    type: String,
    default: '个日程',
  },
  legends: {
    type: Array,
    default: undefined,
  },
})

const summaryLegends = computed(() => {
  if (Array.isArray(props.legends) && props.legends.length)
    return props.legends
  return defaultScheduleLegend
})
</script>

<template>
  <div class="timetable-summary">
    <div class="timetable-summary__left">
      <span class="timetable-summary__accent" />
      <span>
        共 {{ total }} {{ totalUnit }}（未点名 {{ unsignedCount }} {{ unsignedUnit }}）
      </span>
    </div>

    <div class="timetable-summary__right">
      <span
        v-for="item in summaryLegends"
        :key="item.key"
        class="timetable-summary__legend"
      >
        <span
          v-if="item.type === 'bar'"
          class="timetable-summary__legend-bar"
          :style="{ background: item.color }"
        />
        <span
          v-else-if="item.type === 'icon'"
          class="timetable-summary__legend-icon timetable-summary__legend-icon--trial"
        />
        <span
          v-else
          class="timetable-summary__legend-icon timetable-summary__legend-icon--danger"
        />
        {{ item.label }}
      </span>
    </div>
  </div>
</template>

<style scoped lang="less">
.timetable-summary {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 12px 16px 10px;
  border-bottom: 1px solid #edf2f7;
  background: rgb(255 255 255 / 98%);
  backdrop-filter: blur(12px);
}

.timetable-summary__left,
.timetable-summary__right {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.timetable-summary__left {
  color: #1f2937;
  font-size: 13px;
  font-weight: 600;
}

.timetable-summary__accent {
  width: 4px;
  height: 16px;
  border-radius: 999px;
  background: #1677ff;
}

.timetable-summary__legend {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: #4b5563;
  font-size: 12px;
}

.timetable-summary__legend-bar {
  display: inline-block;
  width: 18px;
  height: 4px;
  border-radius: 999px;
}

.timetable-summary__legend-icon {
  position: relative;
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 1px solid #cbd5e1;
  border-radius: 3px;
  background: #fff;
}

.timetable-summary__legend-icon--trial::after {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 6px;
  height: 6px;
  border-radius: 1px;
  background: #b5bfcf;
  content: "";
}

.timetable-summary__legend-icon--danger {
  border-color: #ff7875;
}

.timetable-summary__legend-icon--danger::after {
  position: absolute;
  top: 50%;
  right: 1px;
  left: 1px;
  height: 2px;
  background: #ff4d4f;
  transform: translateY(-50%);
  content: "";
}

@media (max-width: 768px) {
  .timetable-summary {
    align-items: flex-start;
    padding-right: 12px;
    padding-left: 12px;
  }
}
</style>
