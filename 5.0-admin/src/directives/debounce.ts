/* eslint-disable ts/no-unsafe-function-type */
import type { App, Directive } from 'vue'

function debounce(fn: Function, delay: number) {
  let timer: any = null
  return function (this: any, ...args: any[]) {
    clearTimeout(timer)
    timer = setTimeout(() => fn.apply(this, args), delay)
  }
}

export const debounceDirective: Directive = {
  mounted(el, binding) {
    const delay = Number(binding.arg) || 300
    const event = Object.keys(binding.modifiers)[0] || 'click'
    const handler = debounce(binding.value, delay)
    el._debounceHandler = handler
    el.addEventListener(event, handler)
  },
  unmounted(el, binding) {
    const event = Object.keys(binding.modifiers)[0] || 'click'
    el.removeEventListener(event, el._debounceHandler)
  },
}

export function setupDebounceDirective(app: App) {
  app.directive('debounce', debounceDirective)
}
