import dayjs from 'dayjs'
import { ref } from 'vue'
import type { ComputedRef } from 'vue'
import { getGroupClassDetailApi, listGroupClassStudentsByClassIdsApi, pageGroupClassesApi } from '@/api/edu-center/group-class'
import { type TeachingScheduleItem, listTeachingSchedulesApi } from '@/api/edu-center/teaching-schedule'

interface ClassInfo {
  id: string
  name: string
  studentIds: string[]
  studentNames: string[]
  courseId: string
  courseName: string
  mainTeacherId: string
  mainTeacherName: string
  classroomId: string
  classroomName: string
  teacherIds: string[]
  detailLoaded?: boolean
}

interface UseSmartTimetableClassModeOptions {
  activeGroupLabel: ComputedRef<string>
  dataSource: ComputedRef<any[]>
  getLessonIndex: (startTime: string) => string | number
  queryDateRange: ComputedRef<{ startDate: string, endDate: string }>
  resetEmptyLessonConflicts: (scope?: string) => void
  selectedClassroomId?: ComputedRef<string>
  resolveClassroomName?: (id: unknown) => string
}

interface ClassConflictSnapshot {
  classSchedules: TeachingScheduleItem[]
  classroomSchedules: TeachingScheduleItem[]
  studentSchedulesById: Map<string, TeachingScheduleItem[]>
}

interface ConflictExistingScheduleItem {
  name: string
  classTypeText: string
  date: string
  week: string
  timeText: string
  teacherId?: string
  teacherName: string
  assistantNames?: string[]
  classroomName?: string
  studentNames?: string[]
  conflictTypes?: string[]
}

function isTimeOverlap(time1: { start: string, end: string }, time2: { start: string, end: string }) {
  const timeToMinutes = (timeStr: string) => {
    const [hours, minutes] = timeStr.split(':').map(Number)
    return hours * 60 + minutes
  }

  const start1 = timeToMinutes(time1.start)
  const end1 = timeToMinutes(time1.end)
  const start2 = timeToMinutes(time2.start)
  const end2 = timeToMinutes(time2.end)

  return start1 < end2 && start2 < end1
}

function formatWeek(dateText: string) {
  const day = dayjs(dateText).day()
  const weekMap = ['日', '一', '二', '三', '四', '五', '六']
  return `周${weekMap[day] || ''}`
}

function normalizeOptionalClassroomId(value: unknown) {
  const normalized = String(value ?? '').trim()
  if (!normalized || normalized === '0' || normalized.toLowerCase() === 'null' || normalized.toLowerCase() === 'undefined')
    return ''
  return normalized
}

function normalizeClassInfo(value: Partial<ClassInfo> & { id: string, name: string }): ClassInfo {
  return {
    id: String(value.id || '').trim(),
    name: String(value.name || '').trim(),
    studentIds: Array.isArray(value.studentIds) ? value.studentIds.map(item => String(item || '').trim()).filter(Boolean) : [],
    studentNames: Array.isArray(value.studentNames) ? value.studentNames.map(item => String(item || '').trim()).filter(Boolean) : [],
    courseId: String(value.courseId || '').trim(),
    courseName: String(value.courseName || '').trim(),
    mainTeacherId: String(value.mainTeacherId || '').trim(),
    mainTeacherName: String(value.mainTeacherName || '').trim(),
    classroomId: normalizeOptionalClassroomId(value.classroomId),
    classroomName: String(value.classroomName || '').trim(),
    teacherIds: Array.isArray(value.teacherIds) ? value.teacherIds.map(item => String(item || '').trim()).filter(Boolean) : [],
    detailLoaded: value.detailLoaded === true,
  }
}

export function useSmartTimetableClassMode(options: UseSmartTimetableClassModeOptions) {
  const classData = ref<ClassInfo[]>([])
  const classListLoading = ref(false)
  const classDetailLoading = ref(false)
  const classConflictLoading = ref(false)
  const pendingClassLoads = new Map<string, Promise<ClassInfo | null>>()
  const pendingClassConflictLoads = new Map<string, Promise<ClassConflictSnapshot>>()
  const classConflictCache = new Map<string, ClassConflictSnapshot>()
  let classConflictCacheVersion = 0
  let classConflictSeq = 0

  function upsertClassInfo(next: ClassInfo) {
    const currentMap = new Map(classData.value.map(item => [item.id, item]))
    const existing = currentMap.get(next.id)
    currentMap.set(next.id, normalizeClassInfo({
      ...existing,
      ...next,
      studentIds: next.studentIds.length ? next.studentIds : existing?.studentIds,
      studentNames: next.studentNames.length ? next.studentNames : existing?.studentNames,
      teacherIds: next.teacherIds.length ? next.teacherIds : existing?.teacherIds,
      detailLoaded: next.detailLoaded || existing?.detailLoaded,
    }))
    classData.value = [...currentMap.values()]
    return currentMap.get(next.id) || null
  }

  function mapClassListItem(item: any): ClassInfo {
    return normalizeClassInfo({
      id: String(item?.id ?? ''),
      name: String(item?.name || item?.id || '').trim(),
      courseId: String(item?.lessonId ?? ''),
      courseName: String(item?.lessonName || '').trim(),
      mainTeacherId: String(item?.defaultTeacherId ?? ''),
      mainTeacherName: String(item?.defaultTeacherName || '').trim(),
      classroomName: String(item?.classRoomName || '').trim(),
      teacherIds: Array.isArray(item?.teachers) ? item.teachers.map((teacher: any) => teacher?.id) : [],
    })
  }

  async function loadClassOptions(searchKey = '') {
    classListLoading.value = true
    try {
      const res = await pageGroupClassesApi({
        pageRequestModel: {
          needTotal: true,
          pageSize: 50,
          pageIndex: 1,
          skipCount: 0,
        },
        queryModel: {
          className: String(searchKey || '').trim() || undefined,
          statues: [1],
        },
      })
      if (res.code !== 200)
        return
      const list = Array.isArray(res.result?.list) ? res.result.list : []
      list.map(mapClassListItem).forEach(item => upsertClassInfo(item))
    }
    catch (error) {
      console.error('load class options failed', error)
    }
    finally {
      classListLoading.value = false
    }
  }

  function findClassInfo(value: unknown) {
    const normalized = String(value || '').trim()
    if (!normalized)
      return null
    return classData.value.find(item => item.id === normalized) || null
  }

  async function ensureClassLoaded(value: unknown) {
    const classID = String(value || '').trim()
    if (!classID)
      return null

    const existing = findClassInfo(classID)
    if (existing?.detailLoaded)
      return existing

    const pending = pendingClassLoads.get(classID)
    if (pending)
      return pending

    const request = (async () => {
      classDetailLoading.value = true
      try {
        const [detailRes, studentsRes] = await Promise.all([
          getGroupClassDetailApi({ id: classID }),
          listGroupClassStudentsByClassIdsApi({ classIds: [classID] }),
        ])
        if (detailRes.code !== 200)
          throw new Error(detailRes.message || '获取班级详情失败')
        if (studentsRes.code !== 200)
          throw new Error(studentsRes.message || '获取班级学员失败')

        const detail: any = detailRes.result || {}
        const studentBucket = (Array.isArray(studentsRes.result) ? studentsRes.result : [])
          .find(bucket => String(bucket?.classId || '') === classID)
        const students = Array.isArray(studentBucket?.students) ? studentBucket.students : []

        const next = normalizeClassInfo({
          ...mapClassListItem(detail),
          id: classID,
          name: String(detail?.name || existing?.name || classID).trim(),
          classroomId: normalizeOptionalClassroomId(detail?.classroomId),
          classroomName: String(detail?.classroomName || detail?.classRoomName || existing?.classroomName || '').trim(),
          studentIds: students.map(student => String(student?.id || '').trim()).filter(Boolean),
          studentNames: students.map(student => String(student?.name || '').trim()).filter(Boolean),
          detailLoaded: true,
        })

        return upsertClassInfo(next)
      }
      catch (error) {
        console.error('load class detail failed', error)
        return existing || null
      }
      finally {
        classDetailLoading.value = false
        pendingClassLoads.delete(classID)
      }
    })()

    pendingClassLoads.set(classID, request)
    return request
  }

  function clearClassConflictCache() {
    classConflictCacheVersion += 1
    pendingClassConflictLoads.clear()
    classConflictCache.clear()
  }

  function buildClassConflictCacheKey(classInfo: ClassInfo) {
    const effectiveClassroom = resolveEffectiveClassroom(classInfo)
    const { startDate, endDate } = options.queryDateRange.value
    return [
      classInfo.id,
      effectiveClassroom.id,
      startDate,
      endDate,
      [...classInfo.studentIds].sort().join(','),
    ].join('|')
  }

  function resolveEffectiveClassroom(classInfo: ClassInfo) {
    const selectedClassroomId = normalizeOptionalClassroomId(options.selectedClassroomId?.value)
    const selectedClassroomName = selectedClassroomId
      ? String(options.resolveClassroomName?.(selectedClassroomId) || '').trim()
      : ''
    if (selectedClassroomId) {
      return {
        id: selectedClassroomId,
        name: selectedClassroomName || String(classInfo.classroomName || '').trim(),
      }
    }
    return {
      id: normalizeOptionalClassroomId(classInfo.classroomId),
      name: String(classInfo.classroomName || '').trim(),
    }
  }

  async function listTeachingSchedulesSafe(params: Parameters<typeof listTeachingSchedulesApi>[0]) {
    try {
      const res = await listTeachingSchedulesApi(params)
      return res.code === 200 && Array.isArray(res.result) ? res.result : []
    }
    catch (error) {
      console.error('load class related schedules failed', error)
      return []
    }
  }

  function dedupeTeachingSchedules(items: TeachingScheduleItem[]) {
    const map = new Map<string, TeachingScheduleItem>()
    items.forEach((item) => {
      const key = [
        String(item?.id || '').trim(),
        String(item?.teachingClassId || '').trim(),
        String(item?.lessonDate || '').trim(),
        String(item?.startAt || '').trim(),
        String(item?.endAt || '').trim(),
      ].join('|')
      if (key && !map.has(key))
        map.set(key, item)
    })
    return [...map.values()]
  }

  function dedupeExistingSchedules(items: ConflictExistingScheduleItem[]) {
    const map = new Map<string, ConflictExistingScheduleItem>()
    items.forEach((item) => {
      const key = [
        String(item?.teacherId || '').trim(),
        String(item?.date || '').trim(),
        String(item?.timeText || '').trim(),
        String(item?.name || '').trim(),
      ].join('|')
      if (key && !map.has(key))
        map.set(key, item)
    })
    return [...map.values()]
  }

  function extractScheduleDate(value: Partial<TeachingScheduleItem>) {
    const lessonDate = String(value?.lessonDate || '').trim()
    if (lessonDate)
      return lessonDate
    const startAt = String(value?.startAt || '').trim()
    return startAt ? dayjs(startAt).format('YYYY-MM-DD') : ''
  }

  function extractScheduleHHMM(value: unknown) {
    const text = String(value || '').trim()
    if (!text)
      return ''
    const matched = text.match(/(\d{2}:\d{2})/)
    if (matched?.[1])
      return matched[1]
    const parsed = dayjs(text)
    return parsed.isValid() ? parsed.format('HH:mm') : ''
  }

  function scheduleDisplayName(schedule: Partial<TeachingScheduleItem>) {
    const className = String(schedule?.teachingClassName || '').trim()
    const courseName = String(schedule?.lessonName || '').trim()
    if (className && courseName && className !== courseName)
      return `${className}·${courseName}`
    return className || courseName || '课程'
  }

  function scheduleClassTypeText(schedule: Partial<TeachingScheduleItem>) {
    return Number(schedule?.classType) === 2 ? '1对1日程' : '班课日程'
  }

  function normalizeStudentNames(schedule: Partial<TeachingScheduleItem>) {
    const raw = String(schedule?.studentName || '').trim()
    if (!raw)
      return []
    return raw.split(/[、,，]/).map(item => item.trim()).filter(Boolean)
  }

  function buildScheduleMeta(schedule: Partial<TeachingScheduleItem>) {
    const lessonDate = extractScheduleDate(schedule)
    const startTime = extractScheduleHHMM(schedule?.startAt)
    const endTime = extractScheduleHHMM(schedule?.endAt)
    return {
      classId: String(schedule?.teachingClassId || '').trim(),
      className: String(schedule?.teachingClassName || '').trim(),
      classroomName: String(schedule?.classroomName || '').trim(),
      courseName: String(schedule?.lessonName || '').trim(),
      displayName: scheduleDisplayName(schedule),
      lessonDate,
      startTime,
      endTime,
      teacherName: String(schedule?.teacherName || '').trim() || '未知老师',
      dateLabel: lessonDate ? `${dayjs(lessonDate).format('M')}月${dayjs(lessonDate).format('D')}日` : '',
      lessonIndex: startTime ? options.getLessonIndex(startTime) : '',
      timeLabel: startTime && endTime ? `${startTime}-${endTime}` : '',
    }
  }

  function buildExistingScheduleFromTeachingSchedule(
    schedule: Partial<TeachingScheduleItem>,
    conflictTypes: string[] = [],
  ): ConflictExistingScheduleItem {
    const meta = buildScheduleMeta(schedule)
    return {
      name: meta.displayName,
      classTypeText: scheduleClassTypeText(schedule),
      date: meta.lessonDate,
      week: meta.lessonDate ? formatWeek(meta.lessonDate) : '',
      timeText: meta.timeLabel,
      teacherId: String(schedule?.teacherId || '').trim(),
      teacherName: meta.teacherName,
      assistantNames: Array.isArray(schedule?.assistantNames) ? schedule.assistantNames.filter(Boolean) : [],
      classroomName: meta.classroomName,
      studentNames: normalizeStudentNames(schedule),
      conflictTypes,
    }
  }

  function buildExistingScheduleFromMatrixLesson(
    teacher: any,
    lesson: any,
    conflictTypes: string[] = [],
  ): ConflictExistingScheduleItem {
    const studentNames = Array.isArray(lesson?.studentNames)
      ? lesson.studentNames.map((item: any) => String(item?.name || '').trim()).filter(Boolean)
      : []
    const assistantNames = String(lesson?.assistantText || '').trim() && String(lesson?.assistantText || '').trim() !== '未安排'
      ? String(lesson.assistantText).split('、').map(item => item.trim()).filter(Boolean)
      : []
    return {
      name: String(lesson?.className || lesson?.courseName || '').trim() || '课程',
      classTypeText: Number(lesson?.courseType) === 1 ? '1对1日程' : '班课日程',
      date: String(teacher?.date || '').trim(),
      week: String(teacher?.date || '').trim() ? formatWeek(String(teacher.date).trim()) : '',
      timeText: `${String(lesson?.startTime || '').trim()}-${String(lesson?.endTime || '').trim()}`,
      teacherId: String(teacher?.teacherId || '').trim(),
      teacherName: String(teacher?.name || '').trim() || '未知老师',
      assistantNames,
      classroomName: String(lesson?.classroomName || '').trim(),
      studentNames,
      conflictTypes,
    }
  }

  async function loadClassConflictSnapshot(classInfo: ClassInfo) {
    const cacheKey = buildClassConflictCacheKey(classInfo)
    if (classConflictCache.has(cacheKey))
      return classConflictCache.get(cacheKey)!

    const pending = pendingClassConflictLoads.get(cacheKey)
    if (pending)
      return pending

    const request = (async () => {
      const cacheVersion = classConflictCacheVersion
      const { startDate, endDate } = options.queryDateRange.value
      const uniqueStudentIds = Array.from(new Set(classInfo.studentIds))
      const effectiveClassroom = resolveEffectiveClassroom(classInfo)

      const [classSchedules, classroomSchedules, studentEntries] = await Promise.all([
        listTeachingSchedulesSafe({
          startDate,
          endDate,
          groupClassIds: classInfo.id,
        }),
        effectiveClassroom.id
          ? listTeachingSchedulesSafe({
              startDate,
              endDate,
              classroomIds: effectiveClassroom.id,
            })
          : Promise.resolve([]),
        Promise.all(uniqueStudentIds.map(async (studentId) => {
          const rows = await listTeachingSchedulesSafe({
            startDate,
            endDate,
            studentId,
          })
          return [studentId, dedupeTeachingSchedules(rows)] as const
        })),
      ])

      const snapshot: ClassConflictSnapshot = {
        classSchedules: dedupeTeachingSchedules(classSchedules),
        classroomSchedules: dedupeTeachingSchedules(classroomSchedules),
        studentSchedulesById: new Map(studentEntries),
      }
      if (cacheVersion === classConflictCacheVersion)
        classConflictCache.set(cacheKey, snapshot)
      return snapshot
    })()

    pendingClassConflictLoads.set(cacheKey, request)
    try {
      return await request
    }
    finally {
      if (pendingClassConflictLoads.get(cacheKey) === request)
        pendingClassConflictLoads.delete(cacheKey)
    }
  }

  function resolveSelectedClassTarget(value: unknown) {
    const selectedClass = findClassInfo(value)
    return {
      modeLabel: '班课',
      targetLabel: '排课班级',
      targetValue: selectedClass?.name || '未选择班级',
      courseName: selectedClass?.courseName || '未选择课程',
    }
  }

  function buildCombinedConflictReason(reasons: any[]) {
    const normalized = (Array.isArray(reasons) ? reasons : []).filter(Boolean)
    if (!normalized.length)
      return null
    if (normalized.length === 1)
      return normalized[0]

    const conflictTypes = Array.from(new Set(
      normalized
        .flatMap(item => Array.isArray(item?.conflictTypes) ? item.conflictTypes : [])
        .map(item => String(item || '').trim())
        .filter(Boolean),
    ))
    const messages = Array.from(new Set(
      normalized
        .map(item => String(item?.message || '').trim())
        .filter(Boolean),
    ))
    const existingSchedules = dedupeExistingSchedules(
      normalized.flatMap(item => Array.isArray(item?.existingSchedules) ? item.existingSchedules : []),
    )
    const conflictingStudentNames = Array.from(new Set(
      normalized
        .flatMap((item) => {
          if (Array.isArray(item?.conflictingStudentNames))
            return item.conflictingStudentNames
          return item?.studentName ? [item.studentName] : []
        })
        .map(item => String(item || '').trim())
        .filter(Boolean),
    ))

    return {
      type: '班课全局冲突',
      conflictTypes,
      existingSchedules,
      conflictingStudentNames,
      message: messages.length
        ? `该时间段存在${conflictTypes.join('、')}冲突：${messages.join('；')}`
        : '该时间段存在冲突，无法排课',
    }
  }

  function checkClassCrossTimeConflicts(classInfo: ClassInfo, snapshot: ClassConflictSnapshot) {
    options.resetEmptyLessonConflicts()
    const effectiveClassroom = resolveEffectiveClassroom(classInfo)

    options.dataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson: any, lessonIndex: number) => {
        if (lesson.studentId)
          return

        const currentTime = {
          date: teacher.date,
          startTime: lesson.startTime,
          endTime: lesson.endTime,
        }

        const conflictReasons: any[] = []

        const classTimeConflict = snapshot.classSchedules.find((schedule) => {
          const meta = buildScheduleMeta(schedule)
          return meta.lessonDate === currentTime.date
            && isTimeOverlap(
              { start: meta.startTime, end: meta.endTime },
              { start: currentTime.startTime, end: currentTime.endTime },
            )
        })

        if (classTimeConflict) {
          const meta = buildScheduleMeta(classTimeConflict)
          const exactMatch = meta.startTime === currentTime.startTime && meta.endTime === currentTime.endTime
          const reasonType = exactMatch ? '班级已有安排' : '班级时间段交叉冲突'
          const baseMessage = exactMatch
            ? `该时间段${classInfo.name}在${meta.dateLabel}第${meta.lessonIndex}节课[${meta.timeLabel}]已有${meta.teacherName}的${meta.displayName}安排，无法重复排课`
            : `该时间段${classInfo.name}在${meta.dateLabel}第${meta.lessonIndex}节课[${meta.timeLabel}]已有${meta.teacherName}的课程安排，不支持交叉时间段排课`
          conflictReasons.push({
            type: reasonType,
            className: classInfo.name,
            date: meta.dateLabel,
            lessonIndex: meta.lessonIndex,
            teacherName: meta.teacherName,
            group: options.activeGroupLabel.value,
            time: meta.timeLabel,
            conflictTypes: ['班级'],
            existingSchedules: [buildExistingScheduleFromTeachingSchedule(classTimeConflict, ['班级'])],
            message: baseMessage,
          })
        }

        const teacherOtherLesson = teacher.lessons.find((item: any, idx: number) =>
          idx !== lessonIndex
          && item.studentId
          && isTimeOverlap(
            { start: item.startTime, end: item.endTime },
            { start: currentTime.startTime, end: currentTime.endTime },
          ),
        )

        if (teacherOtherLesson) {
          const month = dayjs(teacher.date).format('M')
          const day = dayjs(teacher.date).format('D')
          const timeText = `${teacherOtherLesson.startTime}-${teacherOtherLesson.endTime}`
          conflictReasons.push({
            type: '教师课程冲突',
            teacherName: teacher.name,
            date: `${month}月${day}日`,
            lessonIndex: options.getLessonIndex(teacherOtherLesson.startTime),
            className: teacherOtherLesson.className,
            courseName: teacherOtherLesson.courseName,
            time: timeText,
            conflictTypes: ['老师'],
            existingSchedules: [buildExistingScheduleFromMatrixLesson(teacher, teacherOtherLesson, ['老师'])],
            message: `该时间段${teacher.name}在${month}月${day}日第${options.getLessonIndex(teacherOtherLesson.startTime)}节课[${timeText}]已有${teacherOtherLesson.className || teacherOtherLesson.courseName || '课程'}安排，无法排课`,
          })
        }

        if (effectiveClassroom.id) {
          const classroomConflict = snapshot.classroomSchedules.find((schedule) => {
            const meta = buildScheduleMeta(schedule)
            return meta.classId !== classInfo.id
              && meta.lessonDate === currentTime.date
              && isTimeOverlap(
                { start: meta.startTime, end: meta.endTime },
                { start: currentTime.startTime, end: currentTime.endTime },
              )
          })

          if (classroomConflict) {
            const meta = buildScheduleMeta(classroomConflict)
            conflictReasons.push({
              type: '教室冲突',
              classroomName: effectiveClassroom.name || meta.classroomName,
              date: meta.dateLabel,
              lessonIndex: meta.lessonIndex,
              teacherName: meta.teacherName,
              className: meta.className,
              courseName: meta.courseName,
              time: meta.timeLabel,
              conflictTypes: ['教室'],
              existingSchedules: [buildExistingScheduleFromTeachingSchedule(classroomConflict, ['教室'])],
              message: `该时间段教室${effectiveClassroom.name || meta.classroomName || '-'}在${meta.dateLabel}第${meta.lessonIndex}节课[${meta.timeLabel}]已有${meta.teacherName}的${meta.displayName}安排，无法排课`,
            })
          }
        }

        for (const sid of classInfo.studentIds) {
          const matchedSchedules = snapshot.studentSchedulesById.get(sid) || []
          const studentConflict = matchedSchedules.find((schedule) => {
            const meta = buildScheduleMeta(schedule)
            return meta.classId !== classInfo.id
              && meta.lessonDate === currentTime.date
              && isTimeOverlap(
                { start: meta.startTime, end: meta.endTime },
                { start: currentTime.startTime, end: currentTime.endTime },
              )
          })

          if (!studentConflict)
            continue

          const meta = buildScheduleMeta(studentConflict)
          const studentIndex = classInfo.studentIds.indexOf(sid)
          const studentName = studentIndex >= 0 ? classInfo.studentNames[studentIndex] : '未知学生'
          conflictReasons.push({
            type: '学生课程冲突',
            studentName,
            conflictingStudentNames: [studentName],
            date: meta.dateLabel,
            lessonIndex: meta.lessonIndex,
            teacherName: meta.teacherName,
            courseName: meta.courseName,
            className: meta.className,
            group: options.activeGroupLabel.value,
            time: meta.timeLabel,
            conflictTypes: ['学员'],
            existingSchedules: [buildExistingScheduleFromTeachingSchedule(studentConflict, ['学员'])],
            message: `该时间段${studentName}在${meta.dateLabel}第${meta.lessonIndex}节课[${meta.timeLabel}]已有${meta.teacherName}的${meta.displayName}安排，无法排课`,
          })
        }

        lesson.conflict = conflictReasons.length > 0
        lesson.conflictReason = buildCombinedConflictReason(conflictReasons)
      })
    })
  }

  async function handleClass(value: unknown) {
    const seq = ++classConflictSeq
    if (!value) {
      classConflictLoading.value = false
      options.resetEmptyLessonConflicts()
      return null
    }

    classConflictLoading.value = true
    try {
      const classInfo = await ensureClassLoaded(value)
      if (seq !== classConflictSeq)
        return classInfo

      if (!classInfo) {
        options.resetEmptyLessonConflicts()
        return null
      }

      const snapshot = await loadClassConflictSnapshot(classInfo)
      if (seq !== classConflictSeq)
        return classInfo

      checkClassCrossTimeConflicts(classInfo, snapshot)
      return classInfo
    }
    catch (error) {
      console.error('check class schedule conflicts failed', error)
      if (seq === classConflictSeq)
        options.resetEmptyLessonConflicts()
      return findClassInfo(value)
    }
    finally {
      if (seq === classConflictSeq)
        classConflictLoading.value = false
    }
  }

  function resolveClassConflictMessage(reason: any) {
    if (!reason)
      return ''

    if (reason.message)
      return reason.message

    const groupInfo = reason.group ? `(${reason.group})` : ''
    const timeInfo = reason.time ? `[${reason.time}]` : ''

    if (reason.type === '教师课程冲突')
      return `该时间段${reason.teacherName}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.className || reason.courseName || '课程'}安排，无法排课`
    if (reason.type === '教师班课冲突')
      return `该时间段${reason.teacherName}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.className}的${reason.courseName}班课安排，无法排课`
    if (reason.type === '学生课程冲突')
      return `该时间段${reason.studentName}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.teacherName}${groupInfo}的${reason.courseName || (`${reason.className}班课`)}课程安排，无法排课`
    if (reason.type === '教室冲突')
      return `该时间段教室${reason.classroomName || '-'}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.teacherName}的${reason.className || reason.courseName || '课程'}安排，无法排课`
    if (reason.type === '班级已有安排')
      return `该时间段${reason.className}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.teacherName}的课程安排，无法重复排课`
    if (reason.type === '班级时间段交叉冲突')
      return `该时间段${reason.className}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.teacherName}${groupInfo}的课程安排，不支持交叉时间段排课`
    return ''
  }

  function buildClassScheduleAssignment(classInfo: ClassInfo, teacherId: unknown, assistantIds: unknown = []) {
    const targetTeacherId = String(teacherId || '').trim()
    const effectiveTeacherId = targetTeacherId || String(classInfo?.mainTeacherId || '').trim()
    const rawAssistantIds = (Array.isArray(assistantIds) ? assistantIds : [])
      .map(item => String(item || '').trim())
      .filter(Boolean)
    const removedAssistantIds = Array.from(new Set(rawAssistantIds.filter(item => item === effectiveTeacherId)))
    const normalizedAssistantIds = Array.from(new Set(rawAssistantIds.filter(item => item !== effectiveTeacherId)))
    return {
      teacherId: effectiveTeacherId,
      assistantIds: normalizedAssistantIds,
      removedAssistantIds,
      isMainTeacher: true,
    }
  }

  return {
    buildClassScheduleAssignment,
    classData,
    classConflictLoading,
    classDetailLoading,
    classListLoading,
    clearClassConflictCache,
    ensureClassLoaded,
    findClassInfo,
    handleClass,
    loadClassOptions,
    resolveClassConflictMessage,
    resolveSelectedClassTarget,
  }
}
