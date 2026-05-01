<script setup lang="ts">
import {
  ControlOutlined,
  ProfileOutlined,
  SettingOutlined,
  TeamOutlined,
  WarningOutlined,
} from '@ant-design/icons-vue'
import { isFunction } from '@v-c/utils'
import type { VNodeChild } from 'vue'

const props = defineProps<{
  icon: string | ((...args: any[]) => VNodeChild)
}>()

const menuIcons = {
  ControlOutlined,
  ProfileOutlined,
  SettingOutlined,
  TeamOutlined,
  WarningOutlined,
}

const Comp = computed(() => {
  if (isFunction(props.icon)) {
    const node = props.icon()
    if (node)
      return node
  }
  else {
    return menuIcons[props.icon as keyof typeof menuIcons]
  }
  return undefined
})
</script>

<template>
  <component :is="Comp" v-if="icon" />
</template>
