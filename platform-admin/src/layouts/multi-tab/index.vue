<script lang="ts" setup>
import type { CSSProperties } from 'vue'
import { CloseOutlined, MenuFoldOutlined, MenuUnfoldOutlined, MoreOutlined, ReloadOutlined } from '@ant-design/icons-vue'
import type { RouteLocationNormalized } from 'vue-router'
import { listenerRouteChange, removeRouteListener } from '~@/utils/route-listener'
import { useLayoutState } from '~/layouts/basic-layout/context'

const multiTabStore = useMultiTab()
const { list, activeKey } = storeToRefs(multiTabStore)
const { layoutSetting } = storeToRefs(useAppStore())
const {
  layout,
  collapsed,
  isMobile,
  handleCollapsed,
} = useLayoutState()
const tabStyle = computed<CSSProperties>(() => {
  const style: CSSProperties = {}
  if (layoutSetting.value.multiTabFixed) {
    style.position = 'sticky'
    style.top = `${layoutSetting.value.headerHeight}px`
    style.zIndex = 3
    style.right = 0
  }
  // bugfix https://github.com/antdv-pro/antdv-pro/issues/173
  if (layoutSetting.value.header === false || (layout.value !== 'mix' && layoutSetting.value.fixedHeader === false))
    style.top = '0px'

  return style
})
const tabsRef = shallowRef()

function handleSwitch({ key }: any, current: string) {
  if (key === 'closeCurrent')
    multiTabStore.close(activeKey.value)
  else if (key === 'closeLeft')
    multiTabStore.closeLeft(current)
  else if (key === 'closeRight')
    multiTabStore.closeRight(current)
  else if (key === 'closeOther')
    multiTabStore.closeOther(current)
  else if (key === 'refresh')
    multiTabStore.refresh(activeKey.value)
}

const isCurrentDisabled = computed(() => {
  return (
    list.value.length === 1 || list.value.filter(v => !v.affix).length <= 1
  )
})

function leftDisabled(key: string) {
  // 判断左侧是否还有可关闭的
  const index = list.value.findIndex(v => v.fullPath === key)
  return index === 0 || list.value.filter(v => !v.affix).length <= 1
}

function rightDisabled(key: string) {
  // 判断右侧是否还有可关闭的
  const index = list.value.findIndex(v => v.fullPath === key)
  return (
    index === list.value.length - 1
    || list.value.filter(v => !v.affix).length <= 1
  )
}
const otherDisabled = computed(() => {
  return (
    list.value.length === 1 || list.value.filter(v => !v.affix).length <= 1
  )
})
listenerRouteChange((route: RouteLocationNormalized) => {
  if (route.fullPath.startsWith('/redirect'))
    return
  const item = list.value.find(item => item.fullPath === route.fullPath)

  if (route.fullPath === activeKey.value && !item?.loading)
    return
  activeKey.value = route.fullPath
  multiTabStore.addItem(route)
}, true)
onUnmounted(() => {
  removeRouteListener()
})
</script>

<template>
  <!-- bg-white -->
  <a-tabs
    ref="tabsRef" :active-key="activeKey" :style="tabStyle"
    class="  dark:bg-#242525 w-100% pro-ant-multi-tab" pt-10px type="card" size="small" :tab-bar-gutter="5"
    @update:active-key="multiTabStore.switchTab"
  >
    <a-tab-pane v-for="item in list" :key="item.fullPath">
      <template #tab>
        <a-dropdown :trigger="['contextmenu']">
          <div>
            {{ item.locale ? $t(item.locale) : item.title }}
            <button
              v-if="activeKey === item.fullPath" class="ant-tabs-tab-remove" style="margin: 0"
              @click.stop="multiTabStore.refresh(item.fullPath)"
            >
              <ReloadOutlined :spin="item.loading" />
            </button>
            <button
              v-if="!item.affix && list.length > 1" class="ant-tabs-tab-remove" style="margin: 0"
              @click.stop="multiTabStore.close(item.fullPath)"
            >
              <CloseOutlined />
            </button>
          </div>
          <template #overlay>
            <a-menu @click="handleSwitch($event, item.fullPath)">
              <a-menu-item key="closeCurrent" :disabled="isCurrentDisabled || activeKey !== item.fullPath">
                <!-- 关闭当前 -->
                {{ $t("app.multiTab.closeCurrent") }}
              </a-menu-item>
              <a-menu-item key="closeLeft" :disabled="isCurrentDisabled || leftDisabled(item.fullPath)">
                <!-- 关闭左侧 -->
                {{ $t("app.multiTab.closeLeft") }}
              </a-menu-item>
              <a-menu-item key="closeRight" :disabled="isCurrentDisabled || rightDisabled(item.fullPath)">
                <!-- 关闭右侧 -->
                {{ $t("app.multiTab.closeRight") }}
              </a-menu-item>
              <a-menu-item key="closeOther" :disabled="isCurrentDisabled || otherDisabled">
                <!-- 关闭其他 -->
                {{ $t("app.multiTab.closeOther") }}
              </a-menu-item>
              <a-menu-item key="refresh" :disabled="!isCurrentDisabled">
                <!-- 刷新当前 -->
                {{ $t("app.multiTab.refresh") }}
              </a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </template>
    </a-tab-pane>
    <template #leftExtra>
      <div v-if="isMobile" class="w-12px" />
      <div v-if="!isMobile" class="w-24px" />
      <div v-if="!isMobile" class="collapsedIcon" @click="handleCollapsed?.(!collapsed)">
        <MenuUnfoldOutlined v-if="collapsed" />
        <MenuFoldOutlined v-else />
      </div>
    </template>
    <template #rightExtra>
      <div class="w-48px flex item-center justify-center">
        <a-dropdown :trigger="['hover']">
          <MoreOutlined class="text-16px" />
          <template #overlay>
            <a-menu @click="handleSwitch($event, activeKey)">
              <a-menu-item key="closeOther" :disabled="isCurrentDisabled || otherDisabled">
                <!-- 关闭其他 -->
                {{ $t("app.multiTab.closeOther") }}
              </a-menu-item>
              <a-menu-item key="refresh">
                <!-- 刷新当前 -->
                {{ $t("app.multiTab.refresh") }}
              </a-menu-item>
            </a-menu>
          </template>
        </a-dropdown>
      </div>
    </template>
  </a-tabs>
</template>

<style lang="less">
.ant-tabs-dropdown-placement-bottomRight{
  .ant-tabs-dropdown-menu{
    width: 110px;
  }
    .ant-tabs-tab-remove{
      display: none !important;
    }
  }
.pro-ant-multi-tab {
  transition: all .3s;
  background: #f7f7fb;
  // .ant-tabs-nav-operations {
  //   display: none !important;
  // }

  .ant-tabs-nav-more{
    padding: 0 16px !important;
  }

  .ant-tabs-tab {
  border-radius: 5px !important;
  height: 32px;
  padding-right: 6px !important;
  background-color: #eceef7 !important;
  border: none !important;
  font-size: 12px !important;
  color: #666;
}

.ant-tabs-tab-active {
  background-color: #fff !important;

  .ant-dropdown-trigger {
    font-size: 14px;
    font-weight: 600;
  }
}

.ant-tabs-nav {
  margin-bottom: 12px !important;
  &::before {
    border-bottom: none !important;
  }
}

.collapsedIcon {
  width: 32px;
  height: 32px;
  background: #fff;
  box-shadow: 2px 2px 17px 0 rgba(211, 222, 241, .2);
  border-radius: 8px;
  cursor: pointer;
  user-select: none;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 10px 0 12px;
}
}
</style>
