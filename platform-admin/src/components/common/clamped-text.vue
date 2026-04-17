<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

const props = defineProps({
  text: String,
  lines: {
    type: Number,
    default: 2,
    validator: val => val >= 1,
  },
})

const contentRef = ref(null)
const isOverflow = ref(false)
let rafId = null
let observer = null

const clampStyle = computed(() => ({
  display: '-webkit-box',
  WebkitLineClamp: props.lines,
  WebkitBoxOrient: 'vertical',
  overflow: 'hidden',
  lineClamp: props.lines,
}))

async function checkOverflow() {
  await nextTick()
  const el = contentRef.value
  if (!el)
    return

  const originalStyle = {
    display: el.style.display,
    webkitLineClamp: el.style.webkitLineClamp,
  }

  el.style.display = '-webkit-box'
  el.style.webkitLineClamp = 'unset'
  void el.offsetHeight

  const lineHeight = Number.parseInt(getComputedStyle(el).lineHeight, 10) || 20
  const actualLines = Math.round(el.scrollHeight / lineHeight)

  el.style.display = originalStyle.display
  el.style.webkitLineClamp = originalStyle.webkitLineClamp

  isOverflow.value = actualLines > props.lines
}

function debouncedCheck() {
  if (rafId)
    cancelAnimationFrame(rafId)
  rafId = requestAnimationFrame(() => {
    checkOverflow()
  })
}

watch([() => props.text, () => props.lines], debouncedCheck)

onMounted(() => {
  observer = new ResizeObserver(debouncedCheck)
  if (contentRef.value) {
    observer.observe(contentRef.value)
  }
  setTimeout(debouncedCheck, 60)
})

onBeforeUnmount(() => {
  observer?.disconnect()
  if (rafId)
    cancelAnimationFrame(rafId)
})
</script>

<template>
  <a-tooltip v-if="isOverflow">
    <template #title>
      <div class="whitespace-pre-wrap max-h-60 overflow-y-auto scrollbar">
        {{ text }}
      </div>
    </template>
    <div ref="contentRef" :style="clampStyle">
      {{ text }}
    </div>
  </a-tooltip>
  <div v-else ref="contentRef" :style="clampStyle">
    {{ text }}
  </div>
</template>
