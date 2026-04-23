<script setup lang="ts">
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { Modal, message } from 'ant-design-vue'
import { computed, createVNode, onMounted, ref } from 'vue'
import { type InstConfig, setInstConfigApi } from '~@/api/common/config'
import { useUserStore } from '~@/stores/user'

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

const billingRows = computed(() => [
  {
    key: 'byHours',
    configField: 'enableChargeByHours' as const,
    label: '按课时收费',
    enabled: isConfigEnabled(instConfig.value.enableChargeByHours),
    description: '按“课时购买数”定价，以“课时”为单位计费',
  },
  {
    key: 'byPeriod',
    configField: 'enableByDateLesson' as const,
    label: '按时段收费',
    enabled: isConfigEnabled(instConfig.value.enableByDateLesson),
    description: '按“天/自然月/自然年”定价，以“天”为单位计费',
  },
  {
    key: 'byAmount',
    configField: 'enableChargeByPrice' as const,
    label: '按金额收费',
    enabled: isConfigEnabled(instConfig.value.enableChargeByPrice),
    description: '开启后机构支持按“充值金额”定价，每次点名扣除对应金额数',
  },
])

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
    message.success(value ? '开启成功' : '关闭成功')
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

onMounted(async () => {
  if (!userStore.instConfig)
    await userStore.getInstConfig()
})
</script>

<template>
  <div class="tab-content">
    <div class="setting">
      <custom-title title="收费方式设置" font-size="18px" font-weight="800" before-height="14px" />

      <div class="table-wrap mt-2">
        <table border>
          <tbody>
            <tr v-for="row in billingRows" :key="row.key">
              <td class="td1">
                {{ row.label }}
              </td>
              <td>
                <div class="status-line">
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

                <div class="desc">
                  {{ row.description }}
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

.desc {
  color: #222;
  font-size: 14px;
  line-height: 1.75;
}
</style>
