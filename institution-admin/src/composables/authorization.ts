export const STORAGE_AUTHORIZE_KEY = 'ybcToken'

export const useAuthorization = createGlobalState(() => useStorage<null | string>(STORAGE_AUTHORIZE_KEY, null))
