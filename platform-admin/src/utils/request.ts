import type { AxiosError, AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import axios from 'axios'
import { ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { createVNode } from 'vue'
import { Modal } from 'ant-design-vue'
import { AxiosLoading } from './loading'
import { STORAGE_AUTHORIZE_KEY, useAuthorization } from '~/composables/authorization'
import { ContentTypeEnum, RequestEnum } from '~#/http-enum'
import router from '~/router'
import messageService from '~/utils/messageService'

export interface ResponseBody<T = any> {
  code: number
  result?: T
  message: string
  success?: boolean
  data?: T
  requestId?: string
  total?: number
  current?: number
  size?: number
  pages?: number
  querySource?: string
}

export interface RequestConfigExtra {
  token?: boolean
  customDev?: boolean
  loading?: boolean
  silentError?: boolean
}
const instance: AxiosInstance = axios.create({
  // baseURL: import.meta.env.VITE_APP_BASE_API ?? '/',
  timeout: 60000,
  headers: { 'Content-Type': ContentTypeEnum.JSON },
})

// 添加标志位控制弹窗显示
let isShowingLoginModal = false

// 添加全局401状态标志，用于阻止新请求
let has401Error = false

// 请求管理器：用于跟踪和取消正在进行的请求
class RequestManager {
  private pendingRequests = new Map<string, AbortController>()
  private requestCounter = 0
  
  // 添加请求
  addRequest(config: InternalAxiosRequestConfig): string {
    const controller = new AbortController()
    const baseRequestKey = this.generateRequestKey(config)
    // 为并行请求添加唯一标识，避免相同请求被错误取消
    const requestKey = `${baseRequestKey}_${++this.requestCounter}_${Date.now()}`
    
    this.pendingRequests.set(requestKey, controller)
    config.signal = controller.signal
    
    return requestKey
  }
  
  // 移除请求
  removeRequest(requestKey: string) {
    this.pendingRequests.delete(requestKey)
  }
  
  // 取消所有请求
  cancelAllRequests(reason = 'Request cancelled due to 401 error') {
    this.pendingRequests.forEach((controller) => {
      controller.abort(reason)
    })
    this.pendingRequests.clear()
  }
  
  // 生成请求唯一标识
  private generateRequestKey(config: InternalAxiosRequestConfig): string {
    return `${config.method?.toUpperCase()}_${config.url}_${JSON.stringify(config.params || {})}`
  }
}

const requestManager = new RequestManager()

// 导出请求管理器实例，供外部使用
export { requestManager }

// 重置401状态的函数，供登录成功后调用
export function reset401Status() {
  has401Error = false
  isShowingLoginModal = false
}

const axiosLoading = new AxiosLoading()
async function requestHandler(config: InternalAxiosRequestConfig & RequestConfigExtra): Promise<InternalAxiosRequestConfig> {
  // 如果已经有401错误，直接拒绝新的请求（除了登录接口）
  if (has401Error && !config.url?.includes('/login')) {
    throw new Error('Request blocked due to 401 authentication error')
  }

  // 处理请求前的url
  if (
    import.meta.env.DEV
      && import.meta.env.VITE_APP_BASE_API_DEV
      && import.meta.env.VITE_APP_BASE_URL_DEV
      && config.customDev
  ) {
    //  替换url的请求前缀baseUrl
    config.baseURL = import.meta.env.VITE_APP_BASE_API_DEV
  }
  const token = useAuthorization()
  if (token.value && config.token !== false) {
    config.headers.set(STORAGE_AUTHORIZE_KEY, token.value)
    config.headers.set('Authorization', `Bearer ${token.value}`)
  }

  // 增加多语言的配置
  const { locale } = useI18nLocale()
  config.headers.set('Accept-Language', locale.value ?? 'zh-CN')
  
  // 将请求添加到管理器中
  const requestKey = requestManager.addRequest(config)
  // 将requestKey存储到config中，以便在响应时移除
  ;(config as any).requestKey = requestKey
  
  if (config.loading)
    axiosLoading.addLoading()
  return config
}

function responseHandler(response: any): ResponseBody<any> | AxiosResponse<any> | Promise<any> | any {
  // 从管理器中移除已完成的请求
  const requestKey = response.config?.requestKey
  if (requestKey) {
    requestManager.removeRequest(requestKey)
  }
  return normalizeResponse(response.data)
}

function normalizeResponse(payload: any): ResponseBody<any> {
  if (payload && typeof payload === 'object' && ('success' in payload || 'data' in payload || 'requestId' in payload)) {
    const pageData = payload.data
    const isPageResult = pageData
      && typeof pageData === 'object'
      && Array.isArray(pageData.items)
      && ('total' in pageData)

    return {
      code: payload.success ? 200 : 500,
      result: isPageResult ? pageData.items : payload.data,
      message: payload.message || '',
      success: payload.success,
      data: payload.data,
      requestId: payload.requestId,
      total: isPageResult ? pageData.total : payload.total,
      current: isPageResult ? pageData.current : payload.current,
      size: isPageResult ? pageData.size : payload.size,
      pages: isPageResult ? pageData.pages : payload.pages,
      querySource: isPageResult ? pageData.querySource : payload.querySource,
    }
  }
  return payload
}

function errorHandler(error: AxiosError): Promise<any> {
  const token = useAuthorization()
  const notification = useNotification()
  const silentError = Boolean((error.config as any)?.silentError)

  // 从管理器中移除出错的请求
  const requestKey = (error.config as any)?.requestKey
  if (requestKey) {
    requestManager.removeRequest(requestKey)
  }

      if (error.response) {
    const { data, status, statusText } = error.response as AxiosResponse<ResponseBody>
    if (status === 401) {
      // 设置全局401状态，阻止新请求
      has401Error = true
      
      // 取消所有其他请求
      requestManager.cancelAllRequests('Authentication failed')
      
      // 如果已经在显示弹窗，则直接返回
      if (isShowingLoginModal) {
        return Promise.reject(error)
      }
      
      isShowingLoginModal = true
      Modal.confirm({
        title: '重新登录',
        centered: true,
        icon: createVNode(ExclamationCircleOutlined),
        content: '您的账号已经在别处登录，请注意保护密码。如有问题，请联系机构管理员。',
        onOk() {
          return new Promise<void>((resolve) => {
            token.value = null
            router
              .push({
                path: '/login',
                query: {
                  redirect: router.currentRoute.value.fullPath,
                },
              })
              .then(() => {
                // 跳转完成后重置401状态
                has401Error = false
                resolve()
              })
              .catch((error) => {
                console.error('Navigation failed:', error)
                has401Error = false // 即使失败也要重置状态
                resolve() // Still resolve to close the modal
              })
          })
        },
        onCancel() {
          // 当用户关闭弹窗时，重置标志位
          isShowingLoginModal = false
          has401Error = false
        },
        afterClose() {
          // Modal完全关闭后，重置标志位
          isShowingLoginModal = false
          has401Error = false
        },
      })
    }
    else if (silentError) {
      return Promise.reject(error)
    }
    else if (status === 403) {
      notification?.error({
        message: '403',
        description: data?.message || statusText,
        duration: 3,
      })
    }
    else if (status === 500) {
      notification?.error({
        message: '500',
        description: data?.message || statusText,
        duration: 3,
      })
    }
    else if (status === 400) {
      messageService.error(data?.message || statusText || '请求失败')
    }
    else if (status === 404 || status === 422) {
      notification?.error({
        message: `${status}`,
        description: data?.message || statusText,
        duration: 3,
      })
    }
    else if (status === 502 || status === 503 || status === 504) {
      notification?.error({
        message: '网关错误',
        description: "服务器网关错误，请稍后重试",
        duration: 3,
      })
      // 跳转到502页面
      router.replace({
        path: '/502',
      })
    }
    else {
      notification?.error({
        message: `${status}`,
        description: data?.message || statusText,
        duration: 3,
      })
    }
  }
  return Promise.reject(error)
}
interface AxiosOptions<T> {
  url: string
  params?: T
  data?: T
}
instance.interceptors.request.use(requestHandler)

instance.interceptors.response.use(responseHandler, errorHandler)

export default instance
function instancePromise<R = any, T = any>(options: AxiosOptions<T> & RequestConfigExtra): Promise<ResponseBody<R>> {
  const { loading } = options
  return new Promise((resolve, reject) => {
    instance.request(options).then((res: any) => {
      // console.log(res)
      if (res.code === 500) {
        const notification = useNotification()
        notification?.warning({
          message: '温馨提示',
          description: res?.message,
          duration: 3,
        })
        return resolve(res as any)
      }
      if (res.code === 401) {
        const token = useAuthorization()
        
        // 设置全局401状态，阻止新请求
        has401Error = true
        
        // 取消所有其他请求
        requestManager.cancelAllRequests('Authentication failed')
        
        // 如果已经在显示弹窗，则直接返回
        if (isShowingLoginModal) {
          return resolve(res as any)
        }
        
        isShowingLoginModal = true
        Modal.confirm({
          title: '重新登录',
          centered: true,
          icon: createVNode(ExclamationCircleOutlined),
          content: '您的账号已经在别处登录，请注意保护密码。如有问题，请联系机构管理员。',
          onOk() {
            return new Promise<void>((resolve) => {
              token.value = null
              router
                .push({
                  path: '/login',
                  query: {
                    redirect: router.currentRoute.value.fullPath,
                  },
                })
                .then(() => {
                  // 跳转完成后重置401状态
                  has401Error = false
                  resolve()
                })
                .catch((error) => {
                  console.error('Navigation failed:', error)
                  has401Error = false // 即使失败也要重置状态
                  resolve() // Still resolve to close the modal
                })
            })
          },
          onCancel() {
            // 当用户关闭弹窗时，重置标志位
            isShowingLoginModal = false
            has401Error = false
          },
          afterClose() {
            // Modal完全关闭后，重置标志位
            isShowingLoginModal = false
            has401Error = false
          },
        })
        return resolve(res as any)
      }
      if (res.code !== 200) {
        const notification = useNotification()
        
        // 如果是401相关错误，使用统一的弹窗处理
        if (res.code === 401 || res.message?.includes('401') || res.message?.includes('未授权') || res.message?.includes('登录')) {
          const token = useAuthorization()
          
          // 设置全局401状态，阻止新请求
          has401Error = true
          
          // 取消所有其他请求
          requestManager.cancelAllRequests('Authentication failed')
          
          // 如果已经在显示弹窗，则直接返回
          if (isShowingLoginModal) {
            return resolve(res as any)
          }
          
          isShowingLoginModal = true
          Modal.confirm({
            title: '重新登录',
            centered: true,
            icon: createVNode(ExclamationCircleOutlined),
            content: '您的账号已经在别处登录，请注意保护密码。如有问题，请联系机构管理员。',
            onOk() {
              return new Promise<void>((resolve) => {
                token.value = null
                router
                  .push({
                    path: '/login',
                    query: {
                      redirect: router.currentRoute.value.fullPath,
                    },
                  })
                  .then(() => {
                    // 跳转完成后重置401状态
                    has401Error = false
                    resolve()
                  })
                  .catch((error) => {
                    console.error('Navigation failed:', error)
                    has401Error = false // 即使失败也要重置状态
                    resolve() // Still resolve to close the modal
                  })
              })
            },
            onCancel() {
              // 当用户关闭弹窗时，重置标志位
              isShowingLoginModal = false
              has401Error = false
            },
            afterClose() {
              // Modal完全关闭后，重置标志位
              isShowingLoginModal = false
              has401Error = false
            },
          })
          return resolve(res as any)
        }
        
        notification?.error({
          message: '错误',
          description: res?.message,
          duration: 3,
        })
      }
      resolve(res as any)
    }).catch((e: Error | AxiosError) => {
      // 如果是取消的请求，从管理器中移除
      if (axios.isCancel(e) && 'config' in e && e.config) {
        const requestKey = (e.config as any)?.requestKey
        if (requestKey) {
          requestManager.removeRequest(requestKey)
        }
      }
      reject(e)
    })
      .finally(() => {
        if (loading)
          axiosLoading.closeLoading()
      })
  })
}
export function useGet<R = any, T = any>(url: string, params?: T, config?: AxiosRequestConfig & RequestConfigExtra): Promise<ResponseBody<R>> {
  const options = {
    url,
    params,
    method: RequestEnum.GET,
    ...config,
  }
  return instancePromise<R, T>(options)
}

export function usePost<R = any, T = any>(url: string, data?: T, config?: AxiosRequestConfig & RequestConfigExtra): Promise<ResponseBody<R>> {
  const options = {
    url,
    data,
    method: RequestEnum.POST,
    ...config,
  }
  return instancePromise<R, T>(options)
}

export function usePut<R = any, T = any>(url: string, data?: T, config?: AxiosRequestConfig & RequestConfigExtra): Promise<ResponseBody<R>> {
  const options = {
    url,
    data,
    method: RequestEnum.PUT,
    ...config,
  }
  return instancePromise<R, T>(options)
}

export function useDelete<R = any, T = any>(url: string, data?: T, config?: AxiosRequestConfig & RequestConfigExtra): Promise<ResponseBody<R>> {
  const options = {
    url,
    data,
    method: RequestEnum.DELETE,
    ...config,
  }
  return instancePromise<R, T>(options)
}
