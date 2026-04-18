<script setup lang="ts">
import { LockOutlined } from '@ant-design/icons-vue'
import { message } from 'ant-design-vue'
import type { RouteLocationNormalizedLoaded } from 'vue-router'
import { getMenuAccessCheckApi, getMenuByCodeApi } from '@/api/common/menu'
import { resolveRoutePageAccess } from '~@/router/access-meta'

const props = defineProps<{
  route: RouteLocationNormalizedLoaded
}>()

const { hasAccess } = useAccess()

const pageTitle = computed(() => String(props.route.meta?.title || '当前页面'))
const pageAccess = computed(() => resolveRoutePageAccess(props.route))
const localCanUsePage = computed(() => pageAccess.value.length === 0 || hasAccess(pageAccess.value))
const pagePermissionCode = computed(() => String(pageAccess.value[0] || '').trim())
const accessDeniedImage = ref('')
const canUsePage = ref(true)
const isCheckingPageAccess = ref(false)
const isResolvingAccessDeniedImage = ref(false)
let pageAccessRequestSeed = 0

function handleApply() {
  message.info(`请联系管理员开通“${pageTitle.value}”页面功能权限`)
}

function preloadImage(src: string) {
  return new Promise<string>((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(src)
    image.onerror = () => reject(new Error('load access denied image failed'))
    image.src = src
  })
}

async function loadAccessDeniedImage(permissionCode: string, requestId: number) {
  const code = String(permissionCode || '').trim()
  if (!code) {
    accessDeniedImage.value = ''
    isResolvingAccessDeniedImage.value = false
    return
  }

  accessDeniedImage.value = ''
  isResolvingAccessDeniedImage.value = true
  try {
    const res = await getMenuByCodeApi({
      menuCode: code,
      ownType: 2,
    }, {
      silentError: true,
    })
    const imageUrl = String(res.result?.accessDeniedImage || '').trim()
    const readyImageUrl = imageUrl ? await preloadImage(imageUrl) : ''
    if (pageAccessRequestSeed === requestId && pagePermissionCode.value === code && !canUsePage.value)
      accessDeniedImage.value = readyImageUrl
  }
  catch {
    if (pageAccessRequestSeed === requestId && pagePermissionCode.value === code)
      accessDeniedImage.value = ''
  }
  finally {
    if (pageAccessRequestSeed === requestId)
      isResolvingAccessDeniedImage.value = false
  }
}

watch(
  [() => props.route.fullPath, pagePermissionCode],
  async ([, permissionCode]) => {
    const requestId = ++pageAccessRequestSeed
    canUsePage.value = false
    accessDeniedImage.value = ''
    isResolvingAccessDeniedImage.value = false

    if (!permissionCode) {
      canUsePage.value = true
      isCheckingPageAccess.value = false
      return
    }

    isCheckingPageAccess.value = true

    try {
      const res = await getMenuAccessCheckApi({
        menuCode: permissionCode,
        ownType: 2,
      }, {
        silentError: true,
      })
      if (pageAccessRequestSeed !== requestId)
        return

      const allowed = !!res.result?.allowed
      canUsePage.value = allowed
      if (allowed) {
        accessDeniedImage.value = ''
        return
      }

      await loadAccessDeniedImage(permissionCode, requestId)
    }
    catch {
      if (pageAccessRequestSeed !== requestId)
        return

      const allowed = localCanUsePage.value
      canUsePage.value = allowed
      if (!allowed)
        await loadAccessDeniedImage(permissionCode, requestId)
    }
    finally {
      if (pageAccessRequestSeed !== requestId)
        return
      isCheckingPageAccess.value = false
    }
  },
  { immediate: true },
)

const shouldShowPosterLayout = computed(() => {
  return isCheckingPageAccess.value || isResolvingAccessDeniedImage.value || !!accessDeniedImage.value
})

const shouldShowLoadingState = computed(() => {
  return isCheckingPageAccess.value || isResolvingAccessDeniedImage.value
})
</script>

<template>
  <slot v-if="canUsePage" />
  <div
    v-else
    class="page-access-state"
    :class="{ 'page-access-state--poster': shouldShowPosterLayout }"
  >
    <div v-if="accessDeniedImage" class="page-access-poster-card">
      <img class="page-access-poster-card__image" :src="accessDeniedImage" :alt="pageTitle">
    </div>

    <div v-if="accessDeniedImage" class="page-access-floating-action">
      <a-button type="primary" class="page-access-poster-button" @click="handleApply">
        申请使用
      </a-button>
    </div>

    <div v-else-if="shouldShowLoadingState" class="page-access-poster-loading" aria-hidden="true">
      <div class="page-access-poster-loading__shimmer" />
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

.page-access-poster-loading {
  position: relative;
  width: 100%;
  min-height: calc(100vh - 140px);
  overflow: hidden;
  background:
    linear-gradient(180deg, rgba(238, 243, 255, 0.9) 0%, rgba(246, 248, 255, 0.92) 100%);
}

.page-access-poster-loading__shimmer {
  position: absolute;
  inset: 0;
  background:
    linear-gradient(90deg,
      rgba(255, 255, 255, 0) 0%,
      rgba(255, 255, 255, 0.62) 34%,
      rgba(255, 255, 255, 0) 68%);
  transform: translateX(-100%);
  animation: page-access-poster-shimmer 1.2s ease-in-out infinite;
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

@keyframes page-access-poster-shimmer {
  100% {
    transform: translateX(100%);
  }
}

@media (max-width: 768px) {
  .page-access-state--poster {
    padding: 8px 0 92px;
  }

  .page-access-poster-loading {
    min-height: calc(100vh - 120px);
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
