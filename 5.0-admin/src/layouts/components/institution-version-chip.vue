<script setup lang="ts">
import { useUserStore } from '~@/stores/user'

const openTypeLabelMap: Record<number, string> = {
  1: '体验版',
  2: '基础版',
  3: '高级版',
  4: '旗舰版',
}

const userStore = useUserStore()

const versionLabel = computed(() => {
  const rawName = String(userStore.userInfo?.versionName || '').trim()
  if (rawName)
    return rawName

  const openType = Number(userStore.userInfo?.openType || 0)
  return openTypeLabelMap[openType] || ''
})

const versionClass = computed(() => {
  switch (versionLabel.value) {
    case '体验版':
      return 'inst-version-chip--trial'
    case '基础版':
      return 'inst-version-chip--basic'
    case '高级版':
      return 'inst-version-chip--advanced'
    case '旗舰版':
      return 'inst-version-chip--flagship'
    default:
      return 'inst-version-chip--default'
  }
})
</script>

<template>
  <span
    v-if="versionLabel"
    class="inst-version-chip"
    :class="versionClass"
    :title="`当前版本：${versionLabel}`"
  >
    <span class="inst-version-chip__icon" aria-hidden="true">
      <span class="inst-version-chip__icon-core" />
    </span>
    <span class="inst-version-chip__value">{{ versionLabel }}</span>
  </span>
</template>

<style scoped lang="less">
.inst-version-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  height: 32px;
  padding: 0 12px 0 8px;
  border: 1px solid transparent;
  border-radius: 16px;
  background: linear-gradient(180deg, var(--inst-version-bg-top), var(--inst-version-bg-bottom));
  box-shadow:
    0 10px 24px var(--inst-version-shadow),
    inset 0 1px 0 rgba(255, 255, 255, 0.82);
  white-space: nowrap;
  --inst-version-bg-top: #ffffff;
  --inst-version-bg-bottom: #f8fafc;
  --inst-version-border: #d0d5dd;
  --inst-version-label: #98a2b3;
  --inst-version-value: #344054;
  --inst-version-icon-from: #cbd5e1;
  --inst-version-icon-to: #94a3b8;
  --inst-version-shadow: rgba(15, 23, 42, 0.08);
}

.inst-version-chip__icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border-radius: 999px;
  background: linear-gradient(135deg, var(--inst-version-icon-from), var(--inst-version-icon-to));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.52),
    0 4px 10px rgba(15, 23, 42, 0.12);
}

.inst-version-chip__icon-core {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.95);
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.2);
}

.inst-version-chip__value {
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
  color: var(--inst-version-value);
}

.inst-version-chip--trial {
  border-color: #f5df9b;
  --inst-version-bg-top: #fffdf5;
  --inst-version-bg-bottom: #fff5d9;
  --inst-version-value: #9a6700;
  --inst-version-icon-from: #f8d568;
  --inst-version-icon-to: #e6a700;
  --inst-version-shadow: rgba(234, 179, 8, 0.16);
}

.inst-version-chip--basic {
  border-color: #cfe0ff;
  --inst-version-bg-top: #fbfdff;
  --inst-version-bg-bottom: #edf4ff;
  --inst-version-value: #155eef;
  --inst-version-icon-from: #8bb8ff;
  --inst-version-icon-to: #2563eb;
  --inst-version-shadow: rgba(37, 99, 235, 0.14);
}

.inst-version-chip--advanced {
  border-color: #bcebdc;
  --inst-version-bg-top: #fbfffe;
  --inst-version-bg-bottom: #ebfbf5;
  --inst-version-value: #0f766e;
  --inst-version-icon-from: #5eead4;
  --inst-version-icon-to: #0f766e;
  --inst-version-shadow: rgba(15, 118, 110, 0.16);
}

.inst-version-chip--flagship {
  border-color: #ddceff;
  --inst-version-bg-top: #fdfcff;
  --inst-version-bg-bottom: #f4efff;
  --inst-version-value: #7c3aed;
  --inst-version-icon-from: #c4b5fd;
  --inst-version-icon-to: #7c3aed;
  --inst-version-shadow: rgba(124, 58, 237, 0.15);
}

.inst-version-chip--default {
  border-color: #d0d5dd;
  --inst-version-value: #475467;
  --inst-version-icon-from: #cbd5e1;
  --inst-version-icon-to: #94a3b8;
}

:global([data-theme='dark']) .inst-version-chip {
  border-color: rgba(255, 255, 255, 0.08);
  --inst-version-bg-top: rgba(255, 255, 255, 0.08);
  --inst-version-bg-bottom: rgba(255, 255, 255, 0.04);
  --inst-version-label: rgba(255, 255, 255, 0.52);
  --inst-version-shadow: rgba(0, 0, 0, 0.22);
}

:global([data-theme='dark']) .inst-version-chip--trial {
  border-color: rgba(251, 191, 36, 0.26);
  --inst-version-value: #fbbf24;
}

:global([data-theme='dark']) .inst-version-chip--basic {
  border-color: rgba(96, 165, 250, 0.24);
  --inst-version-value: #93c5fd;
}

:global([data-theme='dark']) .inst-version-chip--advanced {
  border-color: rgba(16, 185, 129, 0.24);
  --inst-version-value: #6ee7b7;
}

:global([data-theme='dark']) .inst-version-chip--flagship {
  border-color: rgba(168, 85, 247, 0.24);
  --inst-version-value: #c4b5fd;
}

:global([data-theme='dark']) .inst-version-chip--default {
  --inst-version-value: rgba(255, 255, 255, 0.82);
}
</style>
