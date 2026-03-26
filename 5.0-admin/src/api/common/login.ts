export interface LoginParams {
  username: string
  password: string
  loginType?: number
  type?: 'account'
}

export interface LoginMobileParams {
  mobile: string
  code: string
  loginType?: number
  type: 'mobile'
}

export interface LoginResultModel {
  token: string
}

export function loginApi(params: LoginParams | LoginMobileParams) {
  return usePost<LoginResultModel, LoginParams | LoginMobileParams>('/sso/sso/doLogin', { ...params, loginType: 2 }, {
    // 设置为false的时候不会携带token
    token: false,
    // 开发模式下使用自定义的接口
    // customDev: true,
    // 是否开启全局请求loading
    loading: true,
  })
}

export function logoutApi() {
  return useGet('/sso/sso/logout')
}
