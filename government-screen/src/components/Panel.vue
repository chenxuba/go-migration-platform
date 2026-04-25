<template>
  <section class="screen-panel">
    <div class="panel-header">
      <h2>{{ title }}</h2>
      <slot name="extra" />
    </div>
    <div class="panel-body">
      <slot />
    </div>
  </section>
</template>

<script setup lang="ts">
defineProps<{ title: string }>()
</script>
