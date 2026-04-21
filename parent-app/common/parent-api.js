import { PARENT_API_BASE_URL, PARENT_TENANT_ID } from '@/common/parent-config'

function validateParentAPIBaseURL() {
	const baseURL = `${PARENT_API_BASE_URL || ''}`.trim()
	if (!baseURL) {
		throw new Error('家长端接口地址未配置')
	}

	// #ifdef MP-WEIXIN
	const platform = `${uni.getSystemInfoSync()?.platform || ''}`.toLowerCase()
	const isRealDevice = platform && platform !== 'devtools'
	if (isRealDevice) {
		if (/^http:\/\/(127\.0\.0\.1|localhost)(:\d+)?$/i.test(baseURL)) {
			throw new Error('真机不能访问 127.0.0.1，请改成可访问的 HTTPS 接口域名')
		}
	}
	// #endif
}

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
		try {
			validateParentAPIBaseURL()
		} catch (error) {
			reject(error)
			return
		}

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

function buildQueryString(params = {}) {
	const entries = Object.entries(params)
		.map(([key, value]) => [key, `${value ?? ''}`.trim()])
		.filter(([, value]) => value)

	if (!entries.length) {
		return ''
	}

	return `?${entries.map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`).join('&')}`
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

export function confirmParentStudents(token, payload) {
	return request({
		url: '/api/v1/parent/students/confirm',
		method: 'POST',
		data: payload,
		token
	})
}

export function listParentCampuses(token) {
	return request({
		url: '/api/v1/parent/campuses',
		method: 'GET',
		token
	})
}

export function listParentSchedules(token, query = {}) {
	return request({
		url: `/api/v1/parent/schedules${buildQueryString({
			startDate: query.startDate,
			endDate: query.endDate
		})}`,
		method: 'GET',
		token
	})
}

export function listParentScheduleDates(token, query = {}) {
	return request({
		url: `/api/v1/parent/schedule-dates${buildQueryString({
			startDate: query.startDate,
			endDate: query.endDate
		})}`,
		method: 'GET',
		token
	})
}

export function listParentClassRecords(token, query = {}) {
	return request({
		url: `/api/v1/parent/class-records${buildQueryString({
			studentId: query.studentId,
			pageIndex: query.pageIndex,
			pageSize: query.pageSize
		})}`,
		method: 'GET',
		token
	})
}

export function listParentCourseEnrollments(token, query = {}) {
	return request({
		url: `/api/v1/parent/course-enrollments${buildQueryString({
			studentId: query.studentId
		})}`,
		method: 'GET',
		token
	})
}

export function getParentCourseEnrollmentDetail(token, query = {}) {
	return request({
		url: `/api/v1/parent/course-enrollments/detail${buildQueryString({
			studentId: query.studentId,
			lessonId: query.lessonId,
			chargingMode: query.chargingMode,
			pageIndex: query.pageIndex,
			pageSize: query.pageSize
		})}`,
		method: 'GET',
		token
	})
}
