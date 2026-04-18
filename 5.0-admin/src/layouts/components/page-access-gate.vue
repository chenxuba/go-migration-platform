<script setup lang="ts">
import { LockOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import type { RouteLocationNormalizedLoaded } from 'vue-router'
import { getMenuByCodeApi } from '@/api/common/menu'
import { resolveRoutePageAccess } from '~@/router/access-meta'

const props = defineProps<{
  route: RouteLocationNormalizedLoaded
}>()

const { hasAccess } = useAccess()
const accessDeniedImageCache = new Map<string, string>()

const pageTitle = computed(() => String(props.route.meta?.title || '当前页面'))
const pageAccess = computed(() => resolveRoutePageAccess(props.route))
const canUsePage = computed(() => pageAccess.value.length === 0 || hasAccess(pageAccess.value))
const pagePermissionCode = computed(() => String(pageAccess.value[0] || '').trim())
const accessDeniedImage = ref('')

function handleApply() {
  message.info(`请联系管理员开通“${pageTitle.value}”页面功能权限`)
}

async function loadAccessDeniedImage(permissionCode: string) {
  const code = String(permissionCode || '').trim()
  if (!code) {
    accessDeniedImage.value = ''
    return
  }

  if (accessDeniedImageCache.has(code)) {
    accessDeniedImage.value = accessDeniedImageCache.get(code) || ''
    return
  }

  try {
    const res = await getMenuByCodeApi({
      menuCode: code,
      ownType: 2,
    }, {
      silentError: true,
    })
    const imageUrl = String(res.result?.accessDeniedImage || '').trim()
    accessDeniedImageCache.set(code, imageUrl)
    if (pagePermissionCode.value === code && !canUsePage.value)
      accessDeniedImage.value = imageUrl
  }
  catch {
    accessDeniedImageCache.set(code, '')
    if (pagePermissionCode.value === code)
      accessDeniedImage.value = ''
  }
}

watch(
  [canUsePage, pagePermissionCode],
  ([allowed, permissionCode]) => {
    if (allowed || !permissionCode) {
      accessDeniedImage.value = ''
      return
    }
    void loadAccessDeniedImage(permissionCode)
  },
  { immediate: true },
)
</script>

<template>
  <slot v-if="canUsePage" />
  <div v-else class="page-access-state" :class="{ 'page-access-state--poster': !!accessDeniedImage }">
    <div v-if="accessDeniedImage" class="page-access-poster-card">
      <img class="page-access-poster-card__image" :src="accessDeniedImage" :alt="pageTitle">
    </div>

    <div v-if="accessDeniedImage" class="page-access-floating-action">
      <a-button type="primary" class="page-access-poster-button" @click="handleApply">
        申请使用
      </a-button>
    </div>

    <div v-else class="page-access-card">
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

.page-access-state--poster {
  min-height: auto;
  display: block;
  padding: 0;
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

.page-access-poster-card {
  width: 100%;
  overflow: hidden;
}

.page-access-poster-card__image {
  width: 100%;
  height: auto;
  display: block;
  object-fit: contain;
  background: #f5f7fa;
}

.page-access-floating-action {
  position: fixed;
  left: 50%;
  bottom: 28px;
  z-index: 120;
  transform: translateX(-50%);
  display: flex;
  justify-content: center;
  pointer-events: none;
}

.page-access-poster-button {
  pointer-events: auto;
  width: 200px;
  height: 58px;
  padding: 0;
  border: none;
  border-radius: 32px;
  color: #fff;
  font-size: 24px;
  font-weight: 600;
  cursor: pointer;
  background: linear-gradient(#737bff 0%, #5039fe 100%);
  box-shadow: 0 8px 20px #5b4ffe66;
  animation: page-access-breathe 1.45s cubic-bezier(0.4, 0, 0.2, 1) infinite;
}

.page-access-poster-button:hover,
.page-access-poster-button:focus {
  background: linear-gradient(#737bff 0%, #5039fe 100%);
  color: #fff;
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

@keyframes page-access-breathe {
  0%,
  100% {
    transform: translateY(0) scale(1);
    box-shadow:
      0 8px 20px rgba(91, 79, 254, 0.4),
      0 0 0 0 rgba(91, 79, 254, 0.26);
  }

  50% {
    transform: translateY(-3px) scale(1.08);
    box-shadow:
      0 18px 34px rgba(91, 79, 254, 0.56),
      0 0 0 16px rgba(91, 79, 254, 0);
  }
}

@media (max-width: 768px) {
  .page-access-state--poster {
    padding: 8px 0 92px;
  }

  .page-access-floating-action {
    bottom: 18px;
  }

  .page-access-poster-button {
    width: 200px;
    height: 58px;
    font-size: 24px;
  }
}
</style>
