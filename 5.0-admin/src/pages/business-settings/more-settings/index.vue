<script setup lang="ts">
import { ref } from 'vue'
import type { InstConfig } from '~@/api/common/config'
import { setInstConfigApi } from '~@/api/common/config'
import CampusDataClear from '~/components/business-settings/campusDataClear.vue'
import { useUserStore } from '~@/stores/user'
import messageService from '@/utils/messageService'

type MainTabKey = 'risk-warning' | 'inventory' | 'system' | 'print' | 'campus-info'
type SystemTabKey = 'peer-info' | 'campus-data-clear'

const userStore = useUserStore()

const activeMainTab = ref<MainTabKey>('system')
const activeSystemTab = ref<SystemTabKey>('campus-data-clear')
const peerInfoLoading = ref(false)
const instConfig = ref<Partial<InstConfig>>({})

const placeholderMap: Record<Exclude<MainTabKey, 'system'>, { title: string, description: string }> = {
  'risk-warning': {
    title: '风险预警设置',
    description: '预警规则与提醒能力后续可统一收敛到这里配置。',
  },
  'inventory': {
    title: '出入库管理',
    description: '库存与出入库相关的业务设置入口已预留，后续可继续补充。',
  },
  'print': {
    title: '打印设置',
    description: '单据、票据与打印模板设置入口已预留，方便后续接入。',
  },
  'campus-info': {
    title: '校区信息设置',
    description: '校区基础资料与展示信息入口已预留，后续可在这里补全。',
  },
}

function isSwitchEnabled(value: unknown) {
  return value === true || value === 1 || value === '1' || value === 'true'
}

async function ensureInstConfig() {
  if (!userStore.instConfig) {
    await userStore.getInstConfig()
  }
  instConfig.value = { ...(userStore.instConfig ?? {}) }
}

async function handlePeerInfoToggle(checked: boolean) {
  if (!Object.keys(instConfig.value).length) {
    await ensureInstConfig()
  }

  const previousValue = instConfig.value.enablePeerInfoAndServiceManagement
  instConfig.value = {
    ...instConfig.value,
    enablePeerInfoAndServiceManagement: checked ? '1' : '0',
  }

  peerInfoLoading.value = true
  try {
    await setInstConfigApi(instConfig.value as InstConfig)
    await userStore.getInstConfig()
    instConfig.value = { ...(userStore.instConfig ?? {}) }
    messageService.success(checked ? '已开启同行资讯与服务管理' : '已关闭同行资讯与服务管理')
  }
  catch (error) {
    console.error('Failed to update peer info setting:', error)
    instConfig.value = {
      ...instConfig.value,
      enablePeerInfoAndServiceManagement: previousValue,
    }
    messageService.error('保存失败，请稍后重试')
  }
  finally {
    peerInfoLoading.value = false
  }
}

function getPlaceholder(key: Exclude<MainTabKey, 'system'>) {
  return placeholderMap[key]
}

onMounted(() => {
  ensureInstConfig()
})
</script>

<template>
  <div class="more-settings-page">
    <a-tabs v-model:active-key="activeMainTab" :animated="false" class="more-settings-page__tabs">
      <a-tab-pane key="risk-warning" tab="风险预警设置">
        <div class="more-settings-page__pane">
          <div class="settings-placeholder-card">
            <div class="settings-placeholder-card__title">
              {{ getPlaceholder('risk-warning').title }}
            </div>
            <div class="settings-placeholder-card__desc">
              {{ getPlaceholder('risk-warning').description }}
            </div>
            <a-empty description="功能建设中" />
          </div>
        </div>
      </a-tab-pane>

      <a-tab-pane key="inventory" tab="出入库管理">
        <div class="more-settings-page__pane">
          <div class="settings-placeholder-card">
            <div class="settings-placeholder-card__title">
              {{ getPlaceholder('inventory').title }}
            </div>
            <div class="settings-placeholder-card__desc">
              {{ getPlaceholder('inventory').description }}
            </div>
            <a-empty description="功能建设中" />
          </div>
        </div>
      </a-tab-pane>

      <a-tab-pane key="system" tab="系统设置">
        <div class="more-settings-page__pane">
          <div class="system-settings">
            <a-tabs v-model:active-key="activeSystemTab" :animated="false" class="system-settings__tabs">
              <a-tab-pane key="peer-info" tab="同行资讯与服务管理">
                <div class="settings-card">
                  <div class="settings-card__header">
                    <div class="settings-card__title-row">
                      <span class="settings-card__marker" />
                      <span class="settings-card__title">同行资讯与服务管理</span>
                    </div>
                    <a-switch
                      :checked="isSwitchEnabled(instConfig.enablePeerInfoAndServiceManagement)"
                      :loading="peerInfoLoading"
                      @change="handlePeerInfoToggle"
                    />
                  </div>
                  <div class="settings-card__desc">
                    开启后，App 端首页会显示同行在使用、同行在看等热门资讯；PC 端会展示相关服务快捷入口。
                  </div>
                  <div class="settings-card__tips">
                    关闭后仅隐藏资讯与服务入口，不影响校区其他业务配置和日常使用。
                  </div>
                </div>
              </a-tab-pane>

              <a-tab-pane key="campus-data-clear" tab="校区数据清空">
                <CampusDataClear embedded />
              </a-tab-pane>
            </a-tabs>
          </div>
        </div>
      </a-tab-pane>

      <a-tab-pane key="print" tab="打印设置">
        <div class="more-settings-page__pane">
          <div class="settings-placeholder-card">
            <div class="settings-placeholder-card__title">
              {{ getPlaceholder('print').title }}
            </div>
            <div class="settings-placeholder-card__desc">
              {{ getPlaceholder('print').description }}
            </div>
            <a-empty description="功能建设中" />
          </div>
        </div>
      </a-tab-pane>

      <a-tab-pane key="campus-info" tab="校区信息设置">
        <div class="more-settings-page__pane">
          <div class="settings-placeholder-card">
            <div class="settings-placeholder-card__title">
              {{ getPlaceholder('campus-info').title }}
            </div>
            <div class="settings-placeholder-card__desc">
              {{ getPlaceholder('campus-info').description }}
            </div>
            <a-empty description="功能建设中" />
          </div>
        </div>
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<style scoped lang="less">
.more-settings-page {
  width: 100%;
  min-height: calc(100vh - 120px);
  color: #666;
}

.more-settings-page__tabs {
  :deep(.ant-tabs-nav) {
    margin: 0;
    padding: 0 12px;
    background: #fff;
    border-radius: 16px 16px 0 0;
  }

  :deep(.ant-tabs-nav-wrap) {
    padding-left: 36px;
  }

  :deep(.ant-tabs-tab) {
    padding: 12px 0;
    font-size: 14px;
    font-weight: 500;
  }

  :deep(.ant-tabs-tab-btn) {
    color: #262626;
  }

  :deep(.ant-tabs-tab-active .ant-tabs-tab-btn) {
    font-weight: 500;
  }

  :deep(.ant-tabs-ink-bar) {
    height: 9px !important;
    background: transparent !important;
    bottom: 1px !important;

    &::after {
      position: absolute;
      top: 0;
      left: calc(50% - 12px);
      width: 24px !important;
      height: 4px !important;
      border-radius: 2px;
      background-color: var(--pro-ant-color-primary, #1677ff);
      content: '';
    }
  }

  :deep(.ant-tabs-content-holder) {
    background: transparent;
  }
}

.more-settings-page__pane {
  min-height: 480px;
  padding-top: 0;
}

.system-settings {
  min-height: calc(100vh - 200px);
  background: #fff;
  border-radius: 0 0 16px 16px;
  overflow: hidden;
}

.system-settings__tabs {
  :deep(.ant-tabs-nav) {
    margin: 0;
    padding: 12px 24px 8px 36px;
    background: #fff;
  }

  :deep(.ant-tabs-nav::before) {
    border-bottom: 1px solid #f0f0f0;
  }

  :deep(.ant-tabs-nav-wrap) {
    padding-left: 0;
  }

  :deep(.ant-tabs-tab) {
    margin: 0 4px 0 0 !important;
    padding: 6px 12px !important;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    transition: none;
  }

  :deep(.ant-tabs-tab + .ant-tabs-tab) {
    margin-left: 0 !important;
  }

  :deep(.ant-tabs-tab-active) {
    background: #e8f1ff;
  }

  :deep(.ant-tabs-tab-active .ant-tabs-tab-btn) {
    color: var(--pro-ant-color-primary, #1677ff);
    font-weight: 500;
  }

  :deep(.ant-tabs-ink-bar) {
    display: none;
  }

  :deep(.ant-tabs-content-holder) {
    padding: 16px 24px 24px;
    background: transparent;
  }
}

.settings-card,
.settings-placeholder-card {
  background: #fff;
  border: 1px solid #edf0f5;
  padding: 24px;
}

.settings-card {
  border-radius: 16px;
}

.settings-placeholder-card {
  border-top: 0;
  border-radius: 0 0 16px 16px;
}

.settings-card__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}

.settings-card__title-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.settings-card__marker {
  width: 6px;
  height: 24px;
  border-radius: 999px;
  background: var(--pro-ant-color-primary, #1677ff);
  flex-shrink: 0;
}

.settings-card__title,
.settings-placeholder-card__title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2329;
}

.settings-card__desc,
.settings-placeholder-card__desc {
  margin-top: 18px;
  font-size: 14px;
  line-height: 1.8;
  color: rgba(0, 0, 0, 0.68);
}

.settings-card__tips {
  margin-top: 12px;
  padding: 12px 14px;
  border-radius: 12px;
  background: #f7f9fc;
  font-size: 13px;
  line-height: 1.7;
  color: rgba(0, 0, 0, 0.55);
}

.settings-placeholder-card {
  min-height: 360px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}
</style>
