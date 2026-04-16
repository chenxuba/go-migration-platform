<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { getTeachingRecordChangeLogPagedListApi, type TeachingRecordChangeLogItem } from '@/api/edu-center/class-record'
import messageService from '@/utils/messageService'

const props = withDefaults(defineProps<{
  teachingRecordId?: string
  refreshToken?: number
}>(), {
  teachingRecordId: '',
  refreshToken: 0,
})

const dataSource = ref<TeachingRecordChangeLogItem[]>([])
const loading = ref(false)
const total = ref(0)

const allColumns = ref<any[]>([
  {
    title: '变更时间',
    dataIndex: 'changeTime',
    key: 'changeTime',
    width: 170,
  },
  {
    title: '变更人',
    dataIndex: 'changeUser',
    key: 'changeUser',
    width: 100,
  },
  {
    title: '变更内容',
    key: 'changeContent',
    dataIndex: 'changeContent',
  },
])

const savedSelected = localStorage.getItem('call-name-change-details')
const keysArray = allColumns.value
  .map(column => column?.key)
  .filter(key => typeof key !== 'undefined')
const initialSelectedValues = savedSelected
  ? JSON.parse(savedSelected)
  : keysArray

const selectedValues = ref(initialSelectedValues)
const columnOptions = computed(() =>
  allColumns.value
    .filter(col => col.key !== 'action')
    .map(col => ({
      id: col.key,
      value: col.title,
      disabled: col.required,
    })),
)

const filteredColumns = computed(() => {
  return allColumns.value.filter((col: any) =>
    col.required || selectedValues.value.includes(col.key),
  )
})

watch(selectedValues, (newVal) => {
  const requiredKeys = allColumns.value
    .filter(col => col.required)
    .map(col => col.key)

  if (!requiredKeys.every(k => newVal.includes(k))) {
    selectedValues.value = Array.from(new Set([
      ...newVal.filter(v => !requiredKeys.includes(v)),
      ...requiredKeys,
    ]))
  }
}, { deep: true })

watch(selectedValues, (newVal) => {
  localStorage.setItem('call-name-change-details', JSON.stringify(newVal))
}, { deep: true })

const totalWidth = computed(() =>
  filteredColumns.value.reduce((acc: number, column: any) => acc + Number(column.width || 0), 0),
)

const tablePagination = computed(() => (dataSource.value.length > 10 ? { hideOnSinglePage: true } : false) as any)

async function loadChangeLogs() {
  const teachingRecordId = String(props.teachingRecordId || '').trim()
  if (!teachingRecordId) {
    dataSource.value = []
    total.value = 0
    return
  }

  loading.value = true
  try {
    const res = await getTeachingRecordChangeLogPagedListApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 1000,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        teachingRecordId,
      },
    })

    if (res.code !== 200) {
      throw new Error(res.message || '加载点名变更记录失败')
    }

    dataSource.value = Array.isArray(res.result?.list) ? res.result.list : []
    total.value = Number(res.result?.total || 0)
  }
  catch (error: any) {
    dataSource.value = []
    total.value = 0
    messageService.error(error?.response?.data?.message || error?.message || '加载点名变更记录失败')
  }
  finally {
    loading.value = false
  }
}

watch(
  () => [props.teachingRecordId, props.refreshToken] as const,
  () => {
    loadChangeLogs()
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 <span class="text-4 mx-2 text-#06f">{{ total }}</span> 条记录
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource"
            :loading="loading"
            :pagination="tablePagination"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'changeUser'">
                <div class="text-#666">
                  {{ record.changeUser || '-' }}
                </div>
              </template>
              <template v-if="column.key === 'changeTime'">
                <div class="text-#666">
                  {{ record.changeTime || '-' }}
                </div>
              </template>
              <template v-if="column.key === 'changeContent'">
                <span class="text-#666">{{ record.changeContent || '-' }}</span>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.total {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    display: inline-block;
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

.studentStatus {

  span.dot {
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    position: relative;
    vertical-align: middle;
    width: 6px;
    margin-right: 4px;
    background: var(--pro-ant-color-primary);
  }
}

.tip {
  padding: 10px 24px 10px 14px;
  background: #e6f0ff;
  color: #333;

  a {
    color: var(--pro-ant-color-primary);
  }
}

.upNew {
  position: relative;

  &::before {
    position: absolute;
    top: -12px;
    left: -22px;
    z-index: 999;
    width: 39px;
    height: 22px;
    background: url(data:image/png;base64,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);
    background-size: contain;
    content: "";
  }
}

.hover {
  &:hover {
    .name {
      color: var(--pro-ant-color-primary);
    }
  }
}
</style>
