<script setup lang="ts">
import { LockOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import type { RouteLocationNormalizedLoaded } from 'vue-router'
import { resolveRoutePageAccess } from '~@/router/access-meta'

const props = defineProps<{
  route: RouteLocationNormalizedLoaded
}>()

const { hasAccess } = useAccess()

const pageTitle = computed(() => String(props.route.meta?.title || '当前页面'))
const pageAccess = computed(() => resolveRoutePageAccess(props.route))
const canUsePage = computed(() => pageAccess.value.length === 0 || hasAccess(pageAccess.value))

function handleApply() {
  message.info(`请联系管理员开通“${pageTitle.value}”页面功能权限`)
}
</script>

<template>
  <slot v-if="canUsePage" />
  <div v-else class="page-access-state">
    <div class="page-access-card">
      <div class="page-access-icon">
        <LockOutlined />
      </div>
      <div class="page-access-title">
        {{ pageTitle }}
      </div>
      <div class="page-access-desc">
        当前已分配菜单访问权限，但暂未开通该页面功能。
      </div>
      <div class="page-access-tip">
        开通后即可正常使用页面内的业务能力；当前仅支持查看并申请使用。
      </div>
      <a-button type="primary" class="page-access-button" @click="handleApply">
        申请使用
      </a-button>
    </div>
  </div>
</template>

<style scoped lang="less">
.page-access-state {
  min-height: calc(100vh - 220px);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.page-access-card {
  width: min(520px, 100%);
  padding: 40px 36px;
  background: #fff;
  border: 1px solid #eef1f6;
  border-radius: 16px;
  box-shadow: 0 12px 32px rgba(15, 35, 95, 0.06);
  text-align: center;
}

.page-access-icon {
  width: 56px;
  height: 56px;
  margin: 0 auto 18px;
  border-radius: 16px;
  background: #edf4ff;
  color: #1677ff;
  font-size: 24px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.page-access-title {
  color: #1f2329;
  font-size: 22px;
  font-weight: 600;
  line-height: 30px;
}

.page-access-desc {
  margin-top: 12px;
  color: #4e5969;
  font-size: 14px;
  line-height: 22px;
}

.page-access-tip {
  margin-top: 8px;
  color: #86909c;
  font-size: 13px;
  line-height: 21px;
}

.page-access-button {
  margin-top: 24px;
  min-width: 120px;
}
</style>
