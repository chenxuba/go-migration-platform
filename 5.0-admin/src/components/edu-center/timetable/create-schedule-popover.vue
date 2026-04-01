<script setup lang="ts">
import { ref } from 'vue'
import addSchedulePopoverBg from '../../../assets/images/timetable/add-schedule-pop-bg.png'
import scheduleClassIcon from '../../../assets/images/timetable/schedule-class.png'
import scheduleFreeIcon from '../../../assets/images/timetable/schedule-free.png'
import scheduleOneToOneIcon from '../../../assets/images/timetable/schedule-one2one.png'
import OneToOneScheduleModal from './one-to-one-schedule-modal.vue'

type ScheduleType = 'class' | 'oneToOne' | 'trial'

const emit = defineEmits<{
  (e: 'select', type: ScheduleType): void
}>()

const scheduleOptions = [
  {
    key: 'class' as const,
    title: '班级',
    description: '创建班级日程',
    icon: scheduleClassIcon,
  },
  {
    key: 'oneToOne' as const,
    title: '1对1',
    description: '创建1对1日程',
    icon: scheduleOneToOneIcon,
  },
  {
    key: 'trial' as const,
    title: '试听',
    description: '创建试听日程',
    icon: scheduleFreeIcon,
  },
]

const panelStyle = {
  backgroundImage: `url(${addSchedulePopoverBg})`,
}

const overlayInnerStyle = {
  padding: '0px',
}

const oneToOneModalOpen = ref(false)

function handleSelect(type: ScheduleType) {
  if (type === 'oneToOne') {
    oneToOneModalOpen.value = true
  }
  emit('select', type)
}
</script>

<template>
  <a-popover
    trigger="hover"
    placement="bottom"
    overlay-class-name="create-schedule-popover-overlay"
    :overlay-inner-style="overlayInnerStyle"
    :mouse-enter-delay="0.08"
  >
    <template #content>
      <div class="create-schedule-popover-panel" :style="panelStyle">
        <div class="create-schedule-popover-panel__title">
          请选择
        </div>

        <div class="create-schedule-popover-panel__grid">
          <button
            v-for="item in scheduleOptions"
            :key="item.key"
            type="button"
            class="create-schedule-card"
            :class="`create-schedule-card--${item.key}`"
            @click="handleSelect(item.key)"
          >
            <span class="create-schedule-card__icon-shell">
              <img
                :src="item.icon"
                :alt="item.title"
                class="create-schedule-card__icon-image"
              >
            </span>

            <span class="create-schedule-card__title">{{ item.title }}</span>
            <span class="create-schedule-card__desc">{{ item.description }}</span>
          </button>
        </div>
      </div>
    </template>

    <slot>
      <a-button type="primary">
        创建日程
      </a-button>
    </slot>
  </a-popover>

  <one-to-one-schedule-modal v-model:open="oneToOneModalOpen" />
</template>

<style scoped lang="less">
:deep(.create-schedule-popover-overlay) {
  padding-top: 10px;
}

:deep(.create-schedule-popover-overlay .ant-popover-inner) {
  padding: 0;
  box-shadow: 0 18px 38px rgb(15 23 42 / 14%);
}

:deep(.create-schedule-popover-overlay .ant-popover-inner-content) {
  padding: 0;
}

.create-schedule-popover-panel {
  width: 289px;
  min-height: 166px;
  padding: 12px 12px 10px;
  background-color: #fff;
  background-repeat: no-repeat;
  background-position: center top;
  background-size: 100% 100%;
  border-radius: 10px;

}

.create-schedule-popover-panel__title {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  margin-bottom: 10px;
  color: #b7bfce;
  font-size: 11px;
  font-weight: 600;
  line-height: 1;
}

.create-schedule-popover-panel__title::before,
.create-schedule-popover-panel__title::after {
  content: '';
  width: 62px;
  height: 1px;
  background: linear-gradient(90deg, rgb(224 231 239 / 0%) 0%, rgb(229 234 242 / 80%) 100%);
}

.create-schedule-popover-panel__title::after {
  transform: rotate(180deg);
}

.create-schedule-popover-panel__grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 4px;
}

.create-schedule-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 6px 2px 4px;
  border: 0;
  border-radius: 12px;
  background: transparent;
  cursor: pointer;
  transition:
    transform 0.16s ease,
    background-color 0.16s ease;
}

.create-schedule-card:hover {
  background: rgb(255 255 255 / 32%);
  transform: translateY(-1px);
}

.create-schedule-card__icon-shell {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 52px;
  height: 52px;
}

.create-schedule-card__icon-image {
  width: 52px;
  height: 52px;
  object-fit: contain;
}

.create-schedule-card__title {
  color: #1f2a44;
  font-size: 12px;
  font-weight: 700;
  line-height: 1.15;
}

.create-schedule-card__desc {
  color: #c0c8d6;
  font-size: 8px;
  line-height: 1.15;
  white-space: nowrap;
}

@media (max-width: 640px) {
  .create-schedule-popover-panel {
    width: min(289px, calc(100vw - 24px));
  }

  .create-schedule-popover-panel__title::before,
  .create-schedule-popover-panel__title::after {
    width: 62px;
  }

  .create-schedule-popover-panel__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 4px;
  }

  .create-schedule-card {
    padding: 6px 2px 4px;
  }
}
</style>
