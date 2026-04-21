import { PARENT_API_BASE_URL, PARENT_TENANT_ID } from '@/common/parent-config'

function buildHeaders(token = '') {
	const headers = {
		'Content-Type': 'application/json',
		'X-Tenant-ID': PARENT_TENANT_ID
	}
	if (token) {
		headers.Authorization = `Bearer ${token}`
	}
	return headers
}

function request({ url, method = 'GET', data, token = '' }) {
	return new Promise((resolve, reject) => {
		uni.request({
			url: `${PARENT_API_BASE_URL}${url}`,
			method,
			data,
			header: buildHeaders(token),
			success(response) {
				const payload = response?.data || {}
				if (response.statusCode >= 200 && response.statusCode < 300 && payload.success !== false) {
					resolve(payload.data)
					return
				}
				reject(new Error(payload.message || '请求失败'))
			},
			fail(error) {
				reject(new Error(error?.errMsg || '网络请求失败'))
			}
		})
	})
}

export function wechatParentLogin(payload) {
	return request({
		url: '/api/v1/parent/auth/wechat/login',
		method: 'POST',
		data: payload
	})
}

export function lookupParentStudentsByPhone(token) {
	return request({
		url: '/api/v1/parent/students/by-phone',
		method: 'GET',
		token
	})
}
