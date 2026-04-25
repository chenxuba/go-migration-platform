<template>
  <div class="mini-metric" :class="color">
    <span>{{ label }}</span>
    <strong>{{ value }}</strong>
    <small>已同步 {{ color === 'cyan' ? '6,065' : '9,842' }} 次</small>
  </div>
</template>

<script setup lang="ts">
defineProps<{ label: string, value: string, color: 'cyan' | 'green' }>()
</script>
