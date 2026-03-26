<script setup>
import { ref } from 'vue'

const props = defineProps({
  options: {
    type: Array,
    default: () => ([
      { id: 1, value: '高' },
      { id: 2, value: '中' },
      { id: 3, value: '低' },
      { id: 4, value: '未知' },
    ]),
  },
  label: {
    type: String,
    default: '自定义字段',
  },
  total: {
    type: Number,
    default: 0,
  },
  num: {
    type: Number,
    default: 0,
  },
  checkedValues: {
    type: Array,
  },
})

const emit = defineEmits(['update:checkedValues'])

const visible = ref(false)

const checkedValues = computed({
  get() {
    return props.checkedValues
  },
  set(value) {
    emit('update:checkedValues', value)
  },
})
</script>

<template>
  <a-dropdown v-model:open="visible" :trigger="['click']" placement="bottomLeft" :arrow="false">
    <template #overlay>
      <a-menu class="max-h-80 overflow-auto scrollbar">
        <a-menu-item class="check-item">
          <a-checkbox-group v-model:value="checkedValues" class="vertical-checkbox-group">
            <a-checkbox v-for="item in options" :key="item.id" :disabled="item.disabled" :value="item.id">
              {{ item.value }}
            </a-checkbox>
          </a-checkbox-group>
        </a-menu-item>
      </a-menu>
    </template>
    <a-button class="flex  mr-2 mb-2">
      {{ label }}({{ num }}/{{ total }})
    </a-button>
  </a-dropdown>
</template>

<style lang="less" scoped>
:deep(.ant-dropdown-menu-item.check-item) {
  &:hover {
    background-color: transparent !important;
  }
}

.vertical-checkbox-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>
