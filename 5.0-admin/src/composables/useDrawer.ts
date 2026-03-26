import { computed } from 'vue'

export function useDrawer(props: { open: boolean }, emit: (event: 'update:open', value: boolean) => void) {
  // 处理双向绑定
  const openDrawer = computed({
    get: () => props.open,
    set: value => emit('update:open', value),
  })

  return {
    openDrawer,
  }
}
