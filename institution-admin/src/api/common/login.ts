export interface LoginParams {
  username: string
  password: string
  loginType?: number
  type?: 'account'
  institutionId?: number
  userId?: number
}

export interface LoginMobileParams {
  mobile: string
  code: string
  loginType?: number
  type: 'mobile'
  institutionId?: number
  userId?: number
}

export interface LoginInstitutionOptionParams {
  identifier: string
  loginType?: number
}

export interface LoginInstitutionOptionModel {
  userId: number
  instId: number
  orgName: string
  loginName: string
  nickName: string
  mobile: string
  logo?: string
  admin: boolean
  institutionStatus?: string
  institutionReadonly?: boolean
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
    silentError: true,
  })
}

export function loginInstitutionOptionsApi(params: LoginInstitutionOptionParams) {
  return usePost<LoginInstitutionOptionModel[], LoginInstitutionOptionParams>('/sso/sso/loginInstitutions', {
    ...params,
    loginType: params.loginType ?? 2,
  }, {
    token: false,
    loading: false,
    silentError: true,
  })
}

export function logoutApi() {
  return useGet('/sso/sso/logout')
}
