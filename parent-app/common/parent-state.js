import { reactive } from 'vue'
import {
	DEFAULT_PHONE,
	campusList,
	cloneList,
	createPhoneSession,
	featureList,
	getCampusById,
	noticeList,
	profileMenus
} from '@/common/mock-parent'

const avatarPalette = ['#8fb7ff', '#7cc8ff', '#63d5b2', '#ffb26f', '#f59fb2', '#9a8cff']
const PARENT_SESSION_STORAGE_KEY = 'parent_app_session'
let hasRestoredParentSession = false

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
		scheduleEntries: [],
		scheduleDateMarks: [],
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

tryAutoRestoreParentSession()

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
	persistParentSession()
}

export function setPostAuthPage(url) {
	parentState.postAuthPage = url || '/pages/profile/index'
	persistParentSession()
}

export function confirmStudentBinding() {
	const selectedIdSet = new Set(parentState.selectedCandidateIds)
	const selectedStudents = parentState.pendingCandidates.filter(item => selectedIdSet.has(item.id))
	if (!selectedStudents.length) {
		return ''
	}
	const studentMap = new Map(parentState.students.map(item => [item.id, { ...item }]))
	selectedStudents.forEach(item => {
		studentMap.set(item.id, {
			...item,
			isBound: true
		})
	})
	parentState.students = Array.from(studentMap.values())
	parentState.currentStudentId = selectedStudents[0].id
	parentState.currentCampusId = selectedStudents[0].campusId
	parentState.pendingCandidates = parentState.pendingCandidates.filter(item => !selectedIdSet.has(item.id))
	parentState.selectedCandidateIds = []
	const targetPage = finalizeStudentBinding(selectedStudents[0].name)
	persistParentSession()
	return targetPage
}

export function finalizeStudentBinding(studentName = '') {
	parentState.bindSuccessVisible = true
	parentState.latestBindStudentName = `${studentName || ''}`.trim()
	const targetPage = parentState.postAuthPage || '/pages/profile/index'
	parentState.postAuthPage = '/pages/profile/index'
	persistParentSession()
	return targetPage
}

export function switchCurrentStudent(studentId) {
	const target = parentState.students.find(item => item.id === studentId)
	if (!target) {
		return
	}
	parentState.currentStudentId = target.id
	parentState.currentCampusId = target.campusId
	persistParentSession()
}

export function switchCurrentCampus(campusId) {
	parentState.currentCampusId = campusId
	const campusStudent = parentState.students.find(item => item.campusId === campusId)
	if (campusStudent) {
		parentState.currentStudentId = campusStudent.id
	}
	persistParentSession()
}

export function dismissBindSuccess() {
	parentState.bindSuccessVisible = false
}

export function logoutParent() {
	Object.assign(parentState, createInitialState())
	clearParentSession()
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

export function applyParentStudentLookup(lookup = {}) {
	applyCandidateSnapshot({
		token: parentState.authToken,
		nickname: parentState.profile.nickname || '微信家长',
		phone: `${lookup.phone || parentState.profile.phone || ''}`.trim(),
		maskedPhone: `${lookup.maskedPhone || parentState.profile.maskedPhone || ''}`.trim(),
		candidates: lookup.candidates
	})
}

export function applyParentCampusSummary(summary = {}) {
	const nextCampusList = normalizeCampusList(summary?.items || summary, parentState.pendingCandidates, parentState.students)
	if (!nextCampusList.length) {
		return
	}

	parentState.campusList = nextCampusList
	const hasCurrentCampus = nextCampusList.some(item => item.id === parentState.currentCampusId)
	if (!hasCurrentCampus) {
		parentState.currentCampusId = nextCampusList[0]?.id || ''
	}

	const currentStudent = parentState.students.find(item => item.campusId === parentState.currentCampusId)
	if (currentStudent) {
		parentState.currentStudentId = currentStudent.id
	}
	persistParentSession()
}

export function applyParentScheduleSummary(summary = {}, options = {}) {
	applyScheduleEntries(summary?.items || summary, options)
}

export function applyParentScheduleDateSummary(summary = {}, options = {}) {
	applyScheduleDateMarks(summary?.items || summary, options)
}

export function restoreParentSession() {
	if (hasRestoredParentSession) {
		return parentState.isAuthenticated
	}
	hasRestoredParentSession = true

	const saved = readParentSession()
	if (!saved || !saved.isAuthenticated) {
		return false
	}

	const nextState = createInitialState()
	nextState.authToken = `${saved.authToken || ''}`.trim()
	nextState.isAuthenticated = !!saved.isAuthenticated
	nextState.profile = normalizeProfile(saved.profile)
	nextState.pendingCandidates = normalizeCandidates(saved.pendingCandidates)
	nextState.selectedCandidateIds = normalizeIDList(saved.selectedCandidateIds)
	nextState.students = normalizeCandidates(saved.students)
	nextState.campusList = normalizeCampusList(saved.campusList, nextState.pendingCandidates, nextState.students)
	nextState.currentStudentId = `${saved.currentStudentId || ''}`.trim()
	nextState.currentCampusId = `${saved.currentCampusId || nextState.campusList[0]?.id || campusList[0].id}`.trim()
	nextState.postAuthPage = `${saved.postAuthPage || '/pages/profile/index'}`.trim() || '/pages/profile/index'
	nextState.bindSuccessVisible = false
	nextState.latestBindStudentName = ''
	reconcileRestoredState(nextState)

	Object.assign(parentState, nextState)
	persistParentSession()
	return true
}

function applyAuthSession(session = {}) {
	applyCandidateSnapshot({
		token: session.token,
		nickname: session.nickname,
		phone: session.phone,
		maskedPhone: session.maskedPhone,
		candidates: session.candidates
	})
}

function applyCandidateSnapshot(session = {}) {
	const normalizedCandidates = normalizeCandidates(session.candidates)
	const { boundStudents, pendingCandidates } = splitParentCandidates(normalizedCandidates)
	const nextCampusList = normalizedCandidates.length ? buildCampusList(normalizedCandidates) : cloneList(campusList)
	const nickname = `${session.nickname || '微信家长'}`.trim() || '微信家长'
	const nextToken = session.token === undefined || session.token === null
		? parentState.authToken
		: `${session.token || ''}`.trim()
	const previousStudentId = `${parentState.currentStudentId || ''}`.trim()
	const currentStudent = boundStudents.find(item => item.id === previousStudentId) || boundStudents[0] || null

	parentState.authToken = nextToken
	parentState.isAuthenticated = true
	parentState.profile = {
		nickname,
		phone: `${session.phone || ''}`.trim(),
		maskedPhone: `${session.maskedPhone || ''}`.trim(),
		avatarText: nickname.slice(0, 1)
	}
	parentState.campusList = nextCampusList
	parentState.pendingCandidates = pendingCandidates
	parentState.selectedCandidateIds = pendingCandidates.map(item => item.id)
	parentState.students = boundStudents
	parentState.scheduleEntries = []
	parentState.scheduleDateMarks = []
	parentState.currentStudentId = currentStudent?.id || ''
	parentState.currentCampusId = currentStudent?.campusId || pendingCandidates[0]?.campusId || nextCampusList[0]?.id || parentState.currentCampusId
	parentState.bindSuccessVisible = false
	parentState.latestBindStudentName = ''
	persistParentSession()
}

function splitParentCandidates(candidates = []) {
	const boundStudents = []
	const pendingCandidates = []

	;(Array.isArray(candidates) ? candidates : []).forEach(item => {
		const nextItem = { ...item }
		if (nextItem.isBound) {
			boundStudents.push(nextItem)
			return
		}
		pendingCandidates.push(nextItem)
	})

	return {
		boundStudents,
		pendingCandidates
	}
}

function normalizeCandidates(candidates = []) {
	return (Array.isArray(candidates) ? candidates : []).map((item, index) => {
		const rawID = `${item?.id ?? `candidate-${index + 1}`}`
		const campusId = `${item?.campusId || (item?.instId ? `inst-${item.instId}` : `campus-${index + 1}`)}`
		return {
			id: rawID,
			rawId: item?.id ?? rawID,
			name: `${item?.name || item?.stuName || '-'}`.trim() || '-',
			avatarUrl: `${item?.avatarUrl || ''}`.trim(),
			campusId,
			campusName: `${item?.campusName || item?.institutionName || '-'}`.trim() || '-',
			campusLogoUrl: `${item?.campusLogoUrl || item?.institutionLogo || item?.logoUrl || ''}`.trim(),
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

function normalizeScheduleEntries(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		id: `${item?.id || item?.scheduleId || `schedule-${index + 1}`}`.trim(),
		scheduleId: `${item?.scheduleId || item?.id || ''}`.trim(),
		instId: Number(item?.instId || 0),
		date: `${item?.date || item?.lessonDate || ''}`.trim(),
		campusId: `${item?.campusId || (item?.instId ? `inst-${item.instId}` : '')}`.trim(),
		campusName: `${item?.campusName || item?.institutionName || '-'}`.trim() || '-',
		studentId: `${item?.studentId || ''}`.trim(),
		studentName: `${item?.studentName || '-'}`.trim() || '-',
		studentAvatarUrl: `${item?.studentAvatarUrl || item?.avatarUrl || ''}`.trim(),
		startTime: `${item?.startTime || ''}`.trim() || '-',
		endTime: `${item?.endTime || ''}`.trim() || '-',
		courseName: `${item?.courseName || item?.lessonName || '-'}`.trim() || '-',
		className: `${item?.className || item?.teachingClassName || '-'}`.trim() || '-',
		teacherName: `${item?.teacherName || '-'}`.trim() || '-',
		classroom: `${item?.classroom || item?.classroomName || '-'}`.trim() || '-',
		note: `${item?.note || '-'}`.trim() || '-',
		statusText: `${item?.statusText || '-'}`.trim() || '-',
		callStatus: Number(item?.callStatus || 0),
		callStatusText: `${item?.callStatusText || ''}`.trim()
	}))
}

function normalizeScheduleDateMarks(list = []) {
	return (Array.isArray(list) ? list : []).map((item, index) => ({
		id: `${item?.campusId || item?.instId || 'campus'}-${item?.date || index + 1}`,
		instId: Number(item?.instId || 0),
		campusId: `${item?.campusId || (item?.instId ? `inst-${item.instId}` : '')}`.trim(),
		campusName: `${item?.campusName || '-'}`.trim() || '-',
		date: `${item?.date || ''}`.trim(),
		scheduleCount: Number(item?.scheduleCount || 0)
	})).filter(item => item.campusId && item.date)
}

function applyScheduleEntries(list = [], options = {}) {
	const nextItems = normalizeScheduleEntries(list)
	const targetDate = `${options?.date || ''}`.trim()
	if (!targetDate) {
		parentState.scheduleEntries = nextItems
		return
	}

	const preserved = parentState.scheduleEntries.filter(item => item.date !== targetDate)
	parentState.scheduleEntries = [...preserved, ...nextItems]
}

function applyScheduleDateMarks(list = [], options = {}) {
	const nextItems = normalizeScheduleDateMarks(list)
	const startDate = `${options?.startDate || ''}`.trim()
	const endDate = `${options?.endDate || ''}`.trim()
	if (!startDate || !endDate) {
		parentState.scheduleDateMarks = nextItems
		return
	}

	const preserved = parentState.scheduleDateMarks.filter(item => item.date < startDate || item.date > endDate)
	parentState.scheduleDateMarks = [...preserved, ...nextItems]
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
			brandName,
			logoUrl: `${item.campusLogoUrl || ''}`.trim()
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

function normalizeProfile(profile = {}) {
	const nickname = `${profile?.nickname || '微信家长'}`.trim() || '微信家长'
	return {
		nickname,
		phone: `${profile?.phone || ''}`.trim(),
		maskedPhone: `${profile?.maskedPhone || ''}`.trim(),
		avatarText: `${profile?.avatarText || nickname.slice(0, 1) || '家'}`.trim() || '家'
	}
}

function normalizeIDList(values = []) {
	return (Array.isArray(values) ? values : []).map(item => `${item}`)
}

function normalizeCampusList(list = [], pending = [], students = []) {
	const normalized = (Array.isArray(list) ? list : [])
		.map(item => ({
			id: `${item?.id || ''}`.trim(),
			name: `${item?.name || ''}`.trim(),
			shortName: `${item?.shortName || ''}`.trim(),
			brandName: `${item?.brandName || ''}`.trim(),
			logoUrl: `${item?.logoUrl || ''}`.trim()
		}))
		.filter(item => item.id && item.name)

	if (normalized.length) {
		return normalized
	}
	return buildCampusList([...(Array.isArray(pending) ? pending : []), ...(Array.isArray(students) ? students : [])])
}

function reconcileRestoredState(state) {
	const studentList = Array.isArray(state.students) ? state.students : []
	const campusIdSet = new Set((Array.isArray(state.campusList) ? state.campusList : []).map(item => item.id))
	const currentStudent = studentList.find(item => item.id === state.currentStudentId) || studentList[0] || null

	if (!currentStudent) {
		state.currentStudentId = ''
	} else {
		state.currentStudentId = currentStudent.id
	}

	if (currentStudent?.campusId) {
		state.currentCampusId = currentStudent.campusId
		return
	}

	if (campusIdSet.has(state.currentCampusId)) {
		return
	}

	state.currentCampusId = state.pendingCandidates[0]?.campusId || state.campusList[0]?.id || campusList[0].id
}

function serializeParentSession() {
	return {
		authToken: parentState.authToken,
		isAuthenticated: parentState.isAuthenticated,
		profile: {
			...parentState.profile
		},
		campusList: parentState.campusList.map(item => ({ ...item })),
		pendingCandidates: parentState.pendingCandidates.map(item => ({ ...item })),
		selectedCandidateIds: [...parentState.selectedCandidateIds],
		students: parentState.students.map(item => ({ ...item })),
		currentStudentId: parentState.currentStudentId,
		currentCampusId: parentState.currentCampusId,
		postAuthPage: parentState.postAuthPage
	}
}

function persistParentSession() {
	try {
		uni.setStorageSync(PARENT_SESSION_STORAGE_KEY, serializeParentSession())
	} catch (error) {
		console.warn('persist parent session failed', error)
	}
}

function readParentSession() {
	try {
		return uni.getStorageSync(PARENT_SESSION_STORAGE_KEY)
	} catch (error) {
		console.warn('read parent session failed', error)
		return null
	}
}

function clearParentSession() {
	try {
		uni.removeStorageSync(PARENT_SESSION_STORAGE_KEY)
	} catch (error) {
		console.warn('clear parent session failed', error)
	}
}

function tryAutoRestoreParentSession() {
	try {
		restoreParentSession()
	} catch (error) {
		hasRestoredParentSession = false
		console.warn('auto restore parent session failed', error)
	}
}
