<script setup lang="ts">
import { ArrowLeftOutlined } from '@ant-design/icons-vue'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import ScaleQuestionBankLoader from './scale-question-bank-loader.vue'

const route = useRoute()
const router = useRouter()

const scaleCode = computed(() => String(route.query.scaleCode || 'PEP3').trim().toUpperCase())
const scaleVersion = computed(() => String(route.query.scaleVersion || '').trim())

function backToScales() {
  router.push({ name: 'PlatformScales' })
}
</script>

<template>
  <div class="scale-question-bank-page">
    <div class="scale-question-bank-page__bar">
      <a-button class="scale-question-bank-page__back" @click="backToScales">
        <template #icon>
          <ArrowLeftOutlined />
        </template>
        返回
      </a-button>
      <div class="scale-question-bank-page__meta">
        <span>{{ scaleCode }}</span>
        <span v-if="scaleVersion">{{ scaleVersion }}</span>
      </div>
    </div>

    <ScaleQuestionBankLoader :scale-code="scaleCode" :scale-version="scaleVersion" />
  </div>
</template>

<style scoped lang="less">
.scale-question-bank-page {
  min-height: 100%;
}

.scale-question-bank-page__bar {
  display: flex;
  align-items: center;
  gap: 12px;
  min-height: 44px;
  margin-bottom: 12px;
}

.scale-question-bank-page__back {
  height: 32px;
  padding: 0 12px;
  border-radius: 6px;
}

.scale-question-bank-page__meta {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #667085;
  font-size: 13px;
  line-height: 20px;
}

.scale-question-bank-page__meta span + span::before {
  content: '·';
  margin-right: 8px;
  color: #c8cdd6;
}
</style>
