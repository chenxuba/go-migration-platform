import type { NavigationGuardNext, RouteLocationNormalized } from 'vue-router'
import router from '~/router'
import { useAuthorization } from '~/composables/authorization'
import { useLoadingCheck, useScrollToTop } from '~/composables/loading'
import { useMetaTitle } from '~/composables/meta-title'
import { setRouteEmitter } from '~@/utils/route-listener'

const allowList = ['/platform/login', '/platform/login-template-preview', '/platform/error', '/platform/401', '/platform/404', '/platform/403', '/platform/502']
const loginPath = '/platform/login'

let authGuardPromise: Promise<typeof import('./auth-guard')> | undefined

export function preloadAuthGuard() {
  if (!authGuardPromise)
    authGuardPromise = import('./auth-guard')
  return authGuardPromise
}

async function runAuthGuard(
  to: RouteLocationNormalized,
  from: RouteLocationNormalized,
  next: NavigationGuardNext,
) {
  const guard = await preloadAuthGuard()
  return guard.handleAuthGuard(to, from, next)
}

router.beforeEach(async (to, from, next) => {
  setRouteEmitter(to)

  const token = useAuthorization()
  if (!token.value) {
    if (!allowList.includes(to.path) && !to.path.startsWith('/platform/redirect')) {
      next({
        path: loginPath,
        query: {
          redirect: encodeURIComponent(to.fullPath),
        },
      })
      return
    }

    next()
    return
  }

  await runAuthGuard(to, from, next)
})

router.afterEach((to) => {
  useMetaTitle(to)
  useLoadingCheck()
  useScrollToTop()

  const token = useAuthorization()
  if (token.value) {
    void preloadAuthGuard().then(guard => guard.handleAuthAfterEach(to))
  }
})
