<template>
  <div class="ring-metric">
    <div class="ring"><span>{{ value }}</span></div>
    <p>{{ text1 }}</p>
    <p>{{ text2 }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{ value: string, text1: string, text2: string }>()
</script>
