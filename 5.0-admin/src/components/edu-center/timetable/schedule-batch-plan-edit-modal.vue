<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { type BatchPlanModalPreset, inferBatchPlanPreset } from './batch-plan-preset'
import GroupClassScheduleModal from './group-class-schedule-modal.vue'
import { type GroupClassBatchPlanModalPreset, inferGroupClassBatchPlanPreset } from './group-class-batch-plan-preset'
import OneToOneScheduleModal from './one-to-one-schedule-modal.vue'
import type { TeachingScheduleItem } from '@/api/edu-center/teaching-schedule'
import { getTeachingScheduleBatchDetailApi } from '@/api/edu-center/teaching-schedule'
import messageService from '@/utils/messageService'

type ScheduleBatchPlanEditScope = 'batch' | 'current'

const props = withDefaults(defineProps<{
  open: boolean
  schedule?: TeachingScheduleItem | null
  scope?: ScheduleBatchPlanEditScope
}>(), {
  scope: 'batch',
})

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'updated'): void
}>()

const modalOpen = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const loading = ref(false)
const preset = ref<BatchPlanModalPreset | null>(null)
const groupClassPreset = ref<GroupClassBatchPlanModalPreset | null>(null)

let loadSeq = 0

async function loadPreset() {
  const current = props.schedule
  if (!modalOpen.value || !current)
    return

  const seq = ++loadSeq
  loading.value = true
  preset.value = null
  groupClassPreset.value = null

  try {
    const res = await getTeachingScheduleBatchDetailApi({
      batchNo: props.scope === 'current' ? undefined : current.batchNo,
      ids: props.scope === 'current' ? [current.id] : undefined,
      id: props.scope === 'current' ? undefined : (current.batchNo ? undefined : current.id),
    })
    if (seq !== loadSeq)
      return
    if (res.code !== 200 || !res.result)
      throw new Error(res.message || '加载批次规则失败')
    if (Number(res.result.classType) === 2) {
      preset.value = inferBatchPlanPreset(res.result, current.id)
      return
    }
    if (Number(res.result.classType) === 1) {
      groupClassPreset.value = {
        ...inferGroupClassBatchPlanPreset(res.result, current.id),
        editScope: props.scope,
      }
      return
    }
    throw new Error('当前日程类型暂不支持编辑')
  }
  catch (error: any) {
    if (seq !== loadSeq)
      return
    console.error('load batch plan preset failed', error)
    messageService.error(error?.response?.data?.message || error?.message || '加载批次规则失败')
    modalOpen.value = false
  }
  finally {
    if (seq === loadSeq)
      loading.value = false
  }
}

watch(
  () => [modalOpen.value, props.scope, props.schedule?.batchNo, props.schedule?.id].join('|'),
  async () => {
    if (!modalOpen.value) {
      preset.value = null
      groupClassPreset.value = null
      loading.value = false
      return
    }
    await loadPreset()
  },
  { immediate: true },
)

function handleUpdated() {
  emit('updated')
}
</script>

<template>
  <OneToOneScheduleModal
    v-if="modalOpen && preset"
    v-model:open="modalOpen"
    mode="editBatch"
    :batch-plan-preset="preset"
    @updated="handleUpdated"
  />

  <GroupClassScheduleModal
    v-else-if="modalOpen && groupClassPreset"
    v-model:open="modalOpen"
    mode="editBatch"
    :batch-plan-preset="groupClassPreset"
    @updated="handleUpdated"
  />

  <a-modal
    v-else
    v-model:open="modalOpen"
    centered
    :footer="null"
    :keyboard="false"
    :closable="false"
    :mask-closable="false"
    :width="420"
  >
    <div class="batch-plan-loading">
      <a-spin :spinning="loading">
        <div class="batch-plan-loading__title">
          正在准备批次规则
        </div>
        <div class="batch-plan-loading__desc">
          正在回显这批课程的生成条件，请稍候。
        </div>
      </a-spin>
    </div>
  </a-modal>
</template>

<style scoped lang="less">
.batch-plan-loading {
  padding: 28px 8px;
  text-align: center;
}

.batch-plan-loading__title {
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
}

.batch-plan-loading__desc {
  margin-top: 8px;
  color: #6b7280;
  font-size: 13px;
  line-height: 1.6;
}
</style>
