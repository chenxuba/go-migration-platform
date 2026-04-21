import { reactive } from 'vue'
import {
	DEFAULT_PHONE,
	campusList,
	cloneList,
	createPhoneSession,
	featureList,
	getCampusById,
	noticeList,
	profileMenus,
	scheduleEntries
} from '@/common/mock-parent'

const avatarPalette = ['#8fb7ff', '#7cc8ff', '#63d5b2', '#ffb26f', '#f59fb2', '#9a8cff']

function createInitialState() {
	return {
		authToken: '',
		authLoading: false,
		isAuthenticated: false,
		profile: {
			nickname: '请登录',
			phone: '',
			maskedPhone: '',
			avatarText: '家'
		},
		campusList: cloneList(campusList),
		featureList: cloneList(featureList),
		noticeList: cloneList(noticeList),
		profileMenus: cloneList(profileMenus),
		scheduleEntries: cloneList(scheduleEntries),
		pendingCandidates: [],
		selectedCandidateIds: [],
		students: [],
		currentStudentId: '',
		currentCampusId: campusList[0].id,
		postAuthPage: '/pages/profile/index',
		bindSuccessVisible: false,
		latestBindStudentName: ''
	}
}

export const parentState = reactive(createInitialState())

export function authorizeByPhone(phone = DEFAULT_PHONE) {
	const session = createPhoneSession(phone)
	applyAuthSession({
		token: '',
		nickname: session.nickname,
		phone: session.phone,
		maskedPhone: session.maskedPhone,
		candidates: session.candidates
	})
	return session
}

export function setPendingSelection(ids) {
	parentState.selectedCandidateIds = Array.isArray(ids) ? [...ids] : []
}

export function setPostAuthPage(url) {
	parentState.postAuthPage = url || '/pages/profile/index'
}

export function confirmStudentBinding() {
	const selectedIdSet = new Set(parentState.selectedCandidateIds)
	const selectedStudents = parentState.pendingCandidates.filter(item => selectedIdSet.has(item.id))
	if (!selectedStudents.length) {
		return ''
	}
	parentState.students = selectedStudents.map(item => ({ ...item }))
	parentState.currentStudentId = selectedStudents[0].id
	parentState.currentCampusId = selectedStudents[0].campusId
	parentState.pendingCandidates = parentState.pendingCandidates.filter(item => !selectedIdSet.has(item.id))
	parentState.selectedCandidateIds = []
	parentState.bindSuccessVisible = true
	parentState.latestBindStudentName = selectedStudents[0].name
	const targetPage = parentState.postAuthPage || '/pages/profile/index'
	parentState.postAuthPage = '/pages/profile/index'
	return targetPage
}

export function switchCurrentStudent(studentId) {
	const target = parentState.students.find(item => item.id === studentId)
	if (!target) {
		return
	}
	parentState.currentStudentId = target.id
	parentState.currentCampusId = target.campusId
}

export function switchCurrentCampus(campusId) {
	parentState.currentCampusId = campusId
	const campusStudent = parentState.students.find(item => item.campusId === campusId)
	if (campusStudent) {
		parentState.currentStudentId = campusStudent.id
	}
}

export function dismissBindSuccess() {
	parentState.bindSuccessVisible = false
}

export function logoutParent() {
	Object.assign(parentState, createInitialState())
}

export function getCurrentCampus() {
	return parentState.campusList.find(item => item.id === parentState.currentCampusId) || getCampusById(parentState.currentCampusId)
}

export function setAuthLoading(loading) {
	parentState.authLoading = !!loading
}

export function applyParentAuthSession(session = {}) {
	applyAuthSession(session)
}

function applyAuthSession(session = {}) {
	const normalizedCandidates = normalizeCandidates(session.candidates)
	const nextCampusList = normalizedCandidates.length ? buildCampusList(normalizedCandidates) : cloneList(campusList)
	const nickname = `${session.nickname || '微信家长'}`.trim() || '微信家长'

	parentState.authToken = `${session.token || ''}`.trim()
	parentState.isAuthenticated = true
	parentState.profile = {
		nickname,
		phone: `${session.phone || ''}`.trim(),
		maskedPhone: `${session.maskedPhone || ''}`.trim(),
		avatarText: nickname.slice(0, 1)
	}
	parentState.campusList = nextCampusList
	parentState.pendingCandidates = normalizedCandidates
	parentState.selectedCandidateIds = normalizedCandidates.map(item => item.id)
	parentState.students = []
	parentState.currentStudentId = ''
	parentState.currentCampusId = normalizedCandidates[0]?.campusId || nextCampusList[0]?.id || parentState.currentCampusId
	parentState.bindSuccessVisible = false
	parentState.latestBindStudentName = ''
}

function normalizeCandidates(candidates = []) {
	return (Array.isArray(candidates) ? candidates : []).map((item, index) => {
		const rawID = `${item?.id ?? `candidate-${index + 1}`}`
		const campusId = `${item?.campusId || (item?.instId ? `inst-${item.instId}` : `campus-${index + 1}`)}`
		return {
			id: rawID,
			rawId: item?.id ?? rawID,
			name: `${item?.name || item?.stuName || '-'}`.trim() || '-',
			campusId,
			campusName: `${item?.campusName || item?.institutionName || '-'}`.trim() || '-',
			balance: Number(item?.balance || 0),
			relation: `${item?.relation || item?.relationText || '-'}`.trim() || '-',
			avatarColor: item?.avatarColor || buildAvatarColor(rawID),
			classLabel: `${item?.classLabel || item?.studentStatusText || '已匹配'}`.trim() || '已匹配',
			studentStatus: Number(item?.studentStatus || 0),
			isBound: !!item?.isBound,
			mobile: `${item?.mobile || ''}`.trim(),
			maskedMobile: `${item?.maskedMobile || ''}`.trim(),
			instId: Number(item?.instId || 0)
		}
	})
}

function buildCampusList(candidates = []) {
	const result = []
	const seen = new Set()

	candidates.forEach(item => {
		if (!item?.campusId || seen.has(item.campusId)) {
			return
		}
		seen.add(item.campusId)
		const campusName = `${item.campusName || ''}`.trim() || '机构'
		const brandName = deriveBrandName(campusName)
		result.push({
			id: item.campusId,
			name: campusName,
			shortName: brandName.slice(0, 1) || '校',
			brandName
		})
	})

	return result.length ? result : cloneList(campusList)
}

function deriveBrandName(name = '') {
	const text = `${name || ''}`.trim()
	if (!text) {
		return '机构'
	}
	return text.replace(/\s*(总校区|控江校区|校区|分校|院区)\s*$/u, '').trim() || text
}

function buildAvatarColor(seed = '') {
	const text = `${seed || ''}`
	let total = 0
	for (let index = 0; index < text.length; index += 1) {
		total += text.charCodeAt(index)
	}
	return avatarPalette[total % avatarPalette.length]
}
