const cloneData = value => JSON.parse(JSON.stringify(value))

export const BASE_DATE = '2026-04-21'
export const DEFAULT_PHONE = '17612341636'

export const campusList = [
	{
		id: 'campus-main',
		name: '上海杨浦区以诺儿童培智服务中心 总校区',
		shortName: '上',
		brandName: '以诺儿童培智服务中心',
		logoUrl: ''
	},
	{
		id: 'campus-kongjiang',
		name: '上海杨浦区以诺儿童培智服务中心 控江校区',
		shortName: '控',
		brandName: '以诺儿童培智服务中心',
		logoUrl: ''
	}
]

export const featureList = [
	{ key: 'attendance', title: '上课记录', shortLabel: '记', accent: '#ff8b7b' },
	{ key: 'course', title: '报读课程', shortLabel: '报', accent: '#92c940' },
	{ key: 'leave', title: '请假申请', shortLabel: '假', accent: '#ffc14a' },
	{ key: 'comment', title: '康复记录', shortLabel: '康', accent: '#8e74ff' },
	{ key: 'homework', title: '课后任务', shortLabel: '任', accent: '#ff8e8e' },
	{ key: 'mall', title: '积分商城', shortLabel: '兑', accent: '#ffb34e' },
	{ key: 'growth', title: '康复档案', shortLabel: '档', accent: '#70c442' },
	{ key: 'feedback', title: '意见反馈', shortLabel: '意', accent: '#9679ff' }
]

export const noticeList = [
	{
		id: 'notice-1',
		title: '语言评估报告',
		summary: '第一期 · 2026-04-21 · 已完成语言表达与认知理解评估。'
	},
	{
		id: 'notice-2',
		title: '阶段康复评估',
		summary: '第二期 · 2026-04-18 · 已更新本周康复建议与训练方向。'
	}
]

export const profileMenus = [
	{ key: 'pending', title: '待关注学员', shortLabel: '待', accent: '#2d84ff' },
	{ key: 'coupon', title: '优惠券', shortLabel: '券', accent: '#ff6d6d' },
	{ key: 'notice', title: '系统通知', shortLabel: '通', accent: '#f4b400' },
	{ key: 'orders', title: '我的订单', shortLabel: '单', accent: '#f4b400' }
]

const phoneLookup = {
	[DEFAULT_PHONE]: {
		nickname: '微信用户',
		candidates: [
			{
				id: 'stu-1',
				name: '张时段',
				campusId: 'campus-main',
				campusName: '上海杨浦区以诺儿童培智服务中心 总校区',
				balance: 0,
				relation: '妈妈',
				avatarColor: '#a9c8ff',
				classLabel: '时段课'
			},
			{
				id: 'stu-2',
				name: '张一鸣',
				campusId: 'campus-main',
				campusName: '上海杨浦区以诺儿童培智服务中心 总校区',
				balance: 0,
				relation: '妈妈',
				avatarColor: '#a9c8ff',
				classLabel: '班级感统课'
			}
		]
	}
}

export const scheduleEntries = [
	{
		id: 'schedule-1',
		date: '2026-04-21',
		campusId: 'campus-main',
		studentId: 'stu-1',
		studentName: '张时段',
		startTime: '08:00',
		endTime: '09:00',
		courseName: '时段课',
		className: '张时段-时段课',
		teacherName: '丁海星',
		classroom: '-',
		note: '-',
		statusText: '已下课'
	},
	{
		id: 'schedule-2',
		date: '2026-04-21',
		campusId: 'campus-main',
		studentId: 'stu-2',
		studentName: '张一鸣',
		startTime: '09:00',
		endTime: '10:00',
		courseName: '班级感统课',
		className: '张一鸣-班级感统课1班',
		teacherName: '郭杨',
		classroom: '-',
		note: '-',
		statusText: '已下课'
	},
	{
		id: 'schedule-3',
		date: '2026-04-22',
		campusId: 'campus-main',
		studentId: 'stu-2',
		studentName: '张一鸣',
		startTime: '14:00',
		endTime: '15:00',
		courseName: '认知训练',
		className: '张一鸣-认知训练',
		teacherName: '吴怡',
		classroom: 'B201',
		note: '记得带训练手册',
		statusText: '待上课'
	},
	{
		id: 'schedule-4',
		date: '2026-04-24',
		campusId: 'campus-kongjiang',
		studentId: 'stu-1',
		studentName: '张时段',
		startTime: '10:30',
		endTime: '11:30',
		courseName: '语言训练',
		className: '张时段-语言训练',
		teacherName: '周青',
		classroom: '控江-301',
		note: '-',
		statusText: '待上课'
	}
]

export function maskPhone(phone) {
	const value = `${phone || ''}`.replace(/\D/g, '')
	if (value.length < 7) {
		return phone || ''
	}
	return `${value.slice(0, 3)} **** ${value.slice(-4)}`
}

export function createPhoneSession(phone = DEFAULT_PHONE) {
	const cleanPhone = `${phone || DEFAULT_PHONE}`.replace(/\D/g, '') || DEFAULT_PHONE
	const target = phoneLookup[cleanPhone] || phoneLookup[DEFAULT_PHONE]
	return {
		nickname: target.nickname,
		phone: cleanPhone,
		maskedPhone: maskPhone(cleanPhone),
		candidates: cloneData(target.candidates)
	}
}

export function getCampusById(campusId) {
	return campusList.find(item => item.id === campusId) || campusList[0]
}

export function cloneList(list) {
	return cloneData(list)
}
