<script setup>
import { RightOutlined } from '@ant-design/icons-vue'

const props = defineProps({
  settingsList: {
    type: Array,
    required: true,
  },
})

const emit = defineEmits(['navigate'])

function navigateToSetting(item) {
  emit('navigate', item)
}
</script>

<template>
  <div class="settings-list">
    <div v-for="(item, index) in settingsList" :key="index" class="settings-item" @click="navigateToSetting(item)">
      <div class="item-content">
        <div class="item-title">
          {{ item.title }}
        </div>
        <div v-if="item.description" class="item-description">
          {{ item.description }}
        </div>
      </div>
      <div v-if="item.status" class="item-status">
        <span class="status-text" :class="[item.status.type]">{{ item.status.text }}</span>
      </div>
      <div v-if="item.hasToggle" class="item-status">
        <ASwitch v-model:checked="item.toggleValue" @click.stop />
      </div>
      <div v-if="!item.hasToggle && !item.status" class="item-arrow">
        <RightOutlined />
      </div>
      <div v-else-if="item.status && item.status.type !== 'enabled'" class="item-arrow">
        <RightOutlined />
      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-list {
  width: 100%;
}

.settings-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
  cursor: pointer;
}

.settings-item:hover {
  background-color: #fafafa;
}

.item-content {
  flex: 1;
}

.item-title {
  font-size: 16px;
  font-weight: 500;
  color: rgba(0, 0, 0, 0.85);
  margin-bottom: 4px;
}

.item-description {
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
}

.item-arrow {
  color: rgba(0, 0, 0, 0.25);
  margin-left: 8px;
}

.item-status {
  display: flex;
  align-items: center;
  margin-left: 8px;
}

.status-text {
  font-size: 12px;
}

.status-text.enabled {
  color: rgba(0, 0, 0, 0.45);
}

.status-text.warning {
  color: #ff4d4f;
}
</style>
