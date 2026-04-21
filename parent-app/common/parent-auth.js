import { wechatParentLogin } from '@/common/parent-api'
import { applyParentAuthSession, setAuthLoading, setPostAuthPage } from '@/common/parent-state'

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
	if (message.length > 22) {
		return '授权登录失败，请检查接口配置'
	}
	return message
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

		const session = await wechatParentLogin({
			loginCode,
			phoneCode
		})
		applyParentAuthSession(session)

		if (Array.isArray(session?.candidates) && session.candidates.length) {
			await navigateToStudentSelection()
			return true
		}

		uni.showToast({
			title: '当前手机号未匹配到学员',
			icon: 'none'
		})
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
