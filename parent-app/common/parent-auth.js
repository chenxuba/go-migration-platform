import { listParentBoundStudents, listParentPendingStudents, wechatParentLogin } from '@/common/parent-api'
import {
	applyParentAuthSession,
	applyParentBoundStudentSummary,
	applyParentPendingStudentSummary,
	parentState,
	setAuthLoading,
	setPostAuthPage
} from '@/common/parent-state'

function loginWithWeChat() {
	return new Promise((resolve, reject) => {
		uni.login({
			provider: 'weixin',
			success: resolve,
			fail: reject
		})
	})
}

function navigateToStudentSelection() {
	return new Promise(resolve => {
		uni.navigateTo({
			url: '/pages/select-student/index',
			success() {
				resolve(true)
			},
			fail() {
				resolve(false)
			}
		})
	})
}

function normalizeErrorMessage(error) {
	const message = `${error?.message || error || ''}`.trim()
	if (!message) {
		return '授权登录失败，请稍后重试'
	}

	if (message.includes('127.0.0.1')) {
		return '真机不能访问本地接口，请改成 HTTPS 域名'
	}
	if (message.includes('HTTP')) {
		return '真机不能请求 HTTP，请改成 HTTPS 域名'
	}
	if (message.includes('url not in domain list') || message.includes('合法域名')) {
		return '接口域名未加入小程序白名单'
	}
	if (message.includes('ssl') || message.includes('certificate')) {
		return '接口证书异常，请检查 HTTPS 配置'
	}
	if (message.includes('request:fail')) {
		return '接口请求失败，请检查域名和网络'
	}
	if (message.length > 30) {
		return message.slice(0, 30)
	}
	return message
}

function isMockLoginCode(code = '') {
	const value = `${code || ''}`.trim().toLowerCase()
	return value === 'the code is a mock one' || value.includes('mock')
}

function hasPendingCandidates(candidates = []) {
	return (Array.isArray(candidates) ? candidates : []).some(item => !item?.isBound)
}

export async function authorizeByWechatPhone(event, options = {}) {
	const postAuthPage = options.postAuthPage || '/pages/profile/index'
	const phoneCode = `${event?.detail?.code || ''}`.trim()
	const errMsg = `${event?.detail?.errMsg || ''}`.trim()

	if (errMsg && !errMsg.includes('ok')) {
		uni.showToast({
			title: '你已取消手机号授权',
			icon: 'none'
		})
		return false
	}
	if (!phoneCode) {
		uni.showToast({
			title: '未获取到手机号授权凭证',
			icon: 'none'
		})
		return false
	}

	setPostAuthPage(postAuthPage)
	setAuthLoading(true)

	try {
		const loginResult = await loginWithWeChat()
		const loginCode = `${loginResult?.code || ''}`.trim()
		if (!loginCode) {
			throw new Error('未获取到微信登录凭证')
		}
		if (isMockLoginCode(loginCode)) {
			throw new Error('当前拿到的是模拟登录 code，请用真实小程序 AppID 重新编译后再试')
		}

		const session = await wechatParentLogin({
			loginCode,
			phoneCode,
			bindTicket: `${parentState.officialBindTicket || ''}`.trim()
		})
		applyParentAuthSession(session)

		const [boundSummary, pendingSummary] = await Promise.all([
			listParentBoundStudents(session?.token || ''),
			listParentPendingStudents(session?.token || '')
		])
		applyParentBoundStudentSummary(boundSummary)
		applyParentPendingStudentSummary(pendingSummary)

		if (hasPendingCandidates(pendingSummary?.candidates)) {
			await navigateToStudentSelection()
			return true
		}

		if (!Array.isArray(boundSummary?.students) || !boundSummary.students.length) {
			uni.showToast({
				title: '当前手机号未匹配到学员',
				icon: 'none'
			})
		}
		return true
	} catch (error) {
		uni.showToast({
			title: normalizeErrorMessage(error),
			icon: 'none'
		})
		return false
	} finally {
		setAuthLoading(false)
	}
}
