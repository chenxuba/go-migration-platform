<script setup lang="ts">
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { computed, createVNode, onMounted, ref } from 'vue'
import { type InstConfig, setInstConfigApi } from '~@/api/common/config'
import { useUserStore } from '~@/stores/user'
import messageService from '~@/utils/messageService'

const userStore = useUserStore()
const rowLoadingMap = ref<Record<string, boolean>>({})

const instConfig = computed<Partial<InstConfig>>(() => userStore.instConfig ?? {})

function isConfigEnabled(value: unknown) {
  if (typeof value === 'boolean')
    return value
  if (typeof value === 'number')
    return value !== 0
  if (typeof value === 'string')
    return value === '1' || value.toLowerCase() === 'true'
  return false
}

const teachingMethodRows = computed(() => [
  {
    key: 'classroom',
    configField: 'enableClassroomTeaching' as const,
    label: '班级授课',
    mode: 'status',
    enabled: isConfigEnabled(instConfig.value.enableClassroomTeaching),
    description: '开启后机构可开设班级授课，支持自定义学员数量，允许上课老师给多名学员上课',
  },
  {
    key: 'oneToOne',
    configField: 'enabledOne2one' as const,
    label: '1对1授课',
    mode: 'status',
    enabled: isConfigEnabled(instConfig.value.enabledOne2one),
    description: '开启后机构可开设1对1课程，学员数量有限制，只允许上课老师给1名学员上课',
  },
  {
    key: 'multiCourse',
    configField: 'enableComposeLesson' as const,
    label: '一班多课',
    mode: 'switch',
    enabled: isConfigEnabled(instConfig.value.enableComposeLesson),
    description: '开启后，机构可设置组合课程，创建班级时支持关联组合课程，实现报名不同课程的学员在同一个班级中上课',
    actionText: '查看组合课程列表',
  },
])

async function ensureInstConfigLoaded() {
  if (!userStore.instConfig)
    await userStore.getInstConfig()
}

function isRowLoading(key: string) {
  return Boolean(rowLoadingMap.value[key])
}

async function updateConfigField(field: keyof InstConfig, value: boolean, key: string) {
  try {
    rowLoadingMap.value = {
      ...rowLoadingMap.value,
      [key]: true,
    }

    await setInstConfigApi({
      ...(instConfig.value as InstConfig),
      [field]: value,
    })
    await userStore.getInstConfig()
    messageService.success(value ? '开启成功' : '关闭成功')
  }
  catch (error) {
    console.error(`update ${String(field)} failed`, error)
  }
  finally {
    rowLoadingMap.value = {
      ...rowLoadingMap.value,
      [key]: false,
    }
  }
}

function handleEnableRow(row: { key: string, label: string, configField: keyof InstConfig }) {
  Modal.confirm({
    title: `开启${row.label}`,
    centered: true,
    icon: createVNode(ExclamationCircleOutlined),
    content: '开启后无法关闭，请谨慎操作',
    async onOk() {
      await updateConfigField(row.configField, true, row.key)
    },
  })
}

async function handleComposeLessonChange(checked: boolean) {
  await updateConfigField('enableComposeLesson', checked, 'multiCourse')
}

onMounted(async () => {
  await ensureInstConfigLoaded()
})
</script>

<template>
  <div class="tab-content">
    <div class="setting">
      <custom-title title="授课方式设置" font-size="18px" font-weight="800" before-height="14px" />

      <div class="table-wrap mt-2">
        <table border>
          <tbody>
            <tr v-for="row in teachingMethodRows" :key="row.key">
              <td class="td1">
                {{ row.label }}
              </td>
              <td>
                <div v-if="row.mode === 'status'" class="status-line">
                  <span class="status-dot" :class="{ 'status-dot--disabled': !row.enabled }" />
                  <span class="status-text" :class="{ 'status-text--disabled': !row.enabled }">{{ row.enabled ? '已开启' : '已关闭' }}</span>
                  <a-button
                    v-if="!row.enabled"
                    type="link"
                    size="small"
                    class="status-enable-btn"
                    :loading="isRowLoading(row.key)"
                    @click="handleEnableRow(row)"
                  >
                    开启
                  </a-button>
                </div>

                <div v-else class="switch-line">
                  <a-switch :checked="row.enabled" :loading="isRowLoading(row.key)" @change="handleComposeLessonChange" />
                </div>

                <div class="desc">
                  {{ row.description }}
                </div>

                <div v-if="row.actionText && row.enabled" class="action-wrap">
                  <a-button type="primary" ghost>
                    {{ row.actionText }}
                  </a-button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped lang="less">
.tab-content {
  background: #fff;
  border-radius: 12px;
  padding: 18px 20px 12px;
}

.setting {
  .table-wrap {
    table {
      width: 100%;
      border: 1px solid #eee;
      border-collapse: collapse;
      border-radius: 8px;
    }

    tr,
    td {
      border: 1px solid #eee;
    }

    td {
      padding: 18px 24px;
      vertical-align: middle;
    }

    .td1 {
      width: 180px;
      color: #333;
      font-size: 14px;
      font-weight: 500;
      text-align: center;
    }
  }
}

.status-line {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.status-dot {
  width: 8px;
  height: 8px;
  background: #1fbe4f;
  border-radius: 999px;
  flex-shrink: 0;
}

.status-dot--disabled {
  background: #c7cbd3;
}

.status-text {
  color: #333;
  font-size: 14px;
  font-weight: 500;
}

.status-text--disabled {
  color: #999;
}

.status-enable-btn {
  height: auto;
  padding: 0;
  margin-left: 8px;
  font-size: 14px;
}

.switch-line {
  display: inline-flex;
  align-items: center;
}

.desc {
  color: #222;
  font-size: 14px;
  line-height: 1.75;
}

.action-wrap {
  margin-top: 12px;
}
</style>
