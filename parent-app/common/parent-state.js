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

function createInitialState() {
	return {
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
	parentState.isAuthenticated = true
	parentState.profile = {
		nickname: session.nickname,
		phone: session.phone,
		maskedPhone: session.maskedPhone,
		avatarText: (session.nickname || '家').slice(0, 1)
	}
	parentState.pendingCandidates = session.candidates
	parentState.selectedCandidateIds = session.candidates.map(item => item.id)
	parentState.students = []
	parentState.currentStudentId = ''
	parentState.currentCampusId = session.candidates[0]?.campusId || parentState.currentCampusId
	parentState.bindSuccessVisible = false
	parentState.latestBindStudentName = ''
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
	return getCampusById(parentState.currentCampusId)
}
