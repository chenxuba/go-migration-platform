// 使用 useColorMode 来更好地控制主题，不跟随系统偏好
import { useColorMode } from '@vueuse/core'
import { computed } from 'vue'

const colorMode = useColorMode({
  initialValue: 'light', // 默认为浅色模式
  modes: {
    light: 'light',
    dark: 'dark'
  },
  storageKey: 'vueuse-color-scheme',
})

// 保持与原有 API 兼容
export const isDark = computed(() => colorMode.value === 'dark')
export const toggleDark = () => {
  colorMode.value = colorMode.value === 'dark' ? 'light' : 'dark'
}

// 新增：直接设置主题模式
export const setDark = (dark: boolean) => {
  colorMode.value = dark ? 'dark' : 'light'
}

// 导出 colorMode 以便外部直接使用
export { colorMode }
