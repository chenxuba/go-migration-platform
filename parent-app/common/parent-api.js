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

function buildUploadHeaders(token = '') {
	const headers = {
		'X-Tenant-ID': PARENT_TENANT_ID
	}
	if (token) {
		headers.Authorization = `Bearer ${token}`
	}
	return headers
}

function createRequestError(message = '', options = {}) {
	const error = new Error(`${message || ''}`.trim() || '请求失败')
	error.statusCode = Number(options?.statusCode || 0)
	error.requestId = `${options?.requestId || ''}`.trim()
	error.code = `${options?.code || ''}`.trim()
	if (!error.code && error.statusCode === 401) {
		error.code = 'UNAUTHORIZED'
	}
	return error
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
				reject(createRequestError(payload.message || '请求失败', {
					statusCode: response?.statusCode,
					requestId: payload?.requestId
				}))
			},
			fail(error) {
				reject(createRequestError(error?.errMsg || '网络请求失败'))
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

export function refreshParentWeChatIdentity(token, payload) {
	return request({
		url: '/api/v1/parent/auth/wechat/identity',
		method: 'POST',
		data: payload,
		token
	})
}

export function getWeChatOfficialBindTicketPreview(payload) {
	return request({
		url: '/api/v1/wechat/official/bind-ticket/preview',
		method: 'POST',
		data: payload
	})
}

export function listWeChatOfficialBindStudents(payload) {
	return request({
		url: '/api/v1/wechat/official/bind-ticket/students',
		method: 'POST',
		data: payload
	})
}

export function confirmWeChatOfficialStudentBinding(payload) {
	return request({
		url: '/api/v1/wechat/official/bind-ticket/confirm',
		method: 'POST',
		data: payload
	})
}

export function getParentWeChatOfficialStatus(token) {
	return request({
		url: '/api/v1/parent/wechat/official/status',
		method: 'GET',
		token
	})
}

export function lookupParentStudentsByPhone(token) {
	return request({
		url: '/api/v1/parent/students/by-phone',
		method: 'GET',
		token
	})
}

export function listParentBoundStudents(token) {
	return request({
		url: '/api/v1/parent/students/bound',
		method: 'GET',
		token
	})
}

export function listParentPendingStudents(token) {
	return request({
		url: '/api/v1/parent/students/pending',
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

export function cancelParentAccount(token, payload) {
	return request({
		url: '/api/v1/parent/account/cancel',
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

export function listParentLeaves(token, query = {}) {
	return request({
		url: `/api/v1/parent/leaves${buildQueryString({
			studentId: query.studentId,
			pageIndex: query.pageIndex,
			pageSize: query.pageSize
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

export function getParentClassRecordDetail(token, query = {}) {
	return request({
		url: `/api/v1/parent/class-records/detail${buildQueryString({
			studentId: query.studentId,
			studentTeachingRecordId: query.studentTeachingRecordId
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

export function getParentCourseArrears(token, query = {}) {
	return request({
		url: `/api/v1/parent/course-arrears${buildQueryString({
			studentId: query.studentId,
			lessonId: query.lessonId,
			chargingMode: query.chargingMode
		})}`,
		method: 'GET',
		token
	})
}

export function listParentRehabRecords(token, query = {}) {
	return request({
		url: `/api/v1/parent/rehab-records${buildQueryString({
			studentId: query.studentId,
			pageIndex: query.pageIndex,
			pageSize: query.pageSize
		})}`,
		method: 'GET',
		token
	})
}

export function getParentRehabRecordDetail(token, query = {}) {
	return request({
		url: `/api/v1/parent/rehab-records/detail${buildQueryString({
			studentId: query.studentId,
			studentTeachingRecordId: query.studentTeachingRecordId
		})}`,
		method: 'GET',
		token
	})
}

export function uploadParentRehabSignature(token, filePath = '') {
	return new Promise((resolve, reject) => {
		try {
			validateParentAPIBaseURL()
		} catch (error) {
			reject(error)
			return
		}

		const normalizedPath = `${filePath || ''}`.trim()
		if (!normalizedPath) {
			reject(new Error('签名文件不能为空'))
			return
		}

		uni.uploadFile({
			url: `${PARENT_API_BASE_URL}/api/v1/parent/rehab-records/signature-upload`,
			filePath: normalizedPath,
			name: 'file',
			header: buildUploadHeaders(token),
			success(response) {
				let payload = {}
				try {
					payload = typeof response?.data === 'string'
						? JSON.parse(response.data || '{}')
						: (response?.data || {})
				} catch (error) {
					reject(new Error('签名上传响应解析失败'))
					return
				}
				if (response.statusCode >= 200 && response.statusCode < 300 && payload.success !== false) {
					resolve(payload.data)
					return
				}
				reject(createRequestError(payload.message || '签名上传失败', {
					statusCode: response?.statusCode,
					requestId: payload?.requestId
				}))
			},
			fail(error) {
				reject(createRequestError(error?.errMsg || '签名上传失败'))
			}
		})
	})
}

export function saveParentRehabFeedback(token, payload = {}) {
	return request({
		url: '/api/v1/parent/rehab-records/feedback',
		method: 'POST',
		data: payload,
		token
	})
}

export function getPEP3CaregiverReportTemplate(params = {}) {
	const query = typeof params === 'string' ? { token: params } : params
	return request({
		url: `/api/v1/parent/pep3/caregiver-report/template${buildQueryString(query)}`,
		method: 'GET'
	})
}

export function submitPEP3CaregiverReport(payload = {}) {
	return request({
		url: '/api/v1/parent/pep3/caregiver-report/submit',
		method: 'POST',
		data: payload
	})
}
