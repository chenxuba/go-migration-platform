import dayjs from 'dayjs'
import { ref } from 'vue'
import type { ComputedRef } from 'vue'
import { getGroupClassDetailApi, listGroupClassStudentsByClassIdsApi, pageGroupClassesApi } from '@/api/edu-center/group-class'
import { type TeachingScheduleValidationResult, validateGroupClassSchedulesApi } from '@/api/edu-center/teaching-schedule'

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
  hasScheduledLesson: (lesson: any) => boolean
  normalizedSelectedAssistantIds: ComputedRef<string[]>
  queryDateRange: ComputedRef<{ startDate: string, endDate: string }>
  resetEmptyLessonConflicts: (scope?: string) => void
  selectedClassroomId?: ComputedRef<string>
  resolveClassroomName?: (id: unknown) => string
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
  const pendingClassConflictLoads = new Map<string, Promise<TeachingScheduleValidationResult | null>>()
  const classConflictCache = new Map<string, TeachingScheduleValidationResult | null>()
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
      [...options.normalizedSelectedAssistantIds.value].sort().join(','),
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

  async function loadClassConflictValidation(classInfo: ClassInfo) {
    const cacheKey = buildClassConflictCacheKey(classInfo)
    if (classConflictCache.has(cacheKey))
      return classConflictCache.get(cacheKey)!

    const pending = pendingClassConflictLoads.get(cacheKey)
    if (pending)
      return pending

    const request = (async () => {
      const cacheVersion = classConflictCacheVersion
      const effectiveClassroom = resolveEffectiveClassroom(classInfo)
      const schedules = options.dataSource.value.flatMap((teacher) => {
        return (Array.isArray(teacher?.lessons) ? teacher.lessons : [])
          .filter((lesson: any) => !options.hasScheduledLesson(lesson))
          .map((lesson: any) => {
            const assignment = buildClassScheduleAssignment(
              classInfo,
              teacher?.teacherId,
              options.normalizedSelectedAssistantIds.value,
            )
            return {
              lessonDate: String(teacher?.date || '').trim(),
              startTime: String(lesson?.startTime || '').trim(),
              endTime: String(lesson?.endTime || '').trim(),
              teacherId: assignment.teacherId,
              assistantIds: assignment.assistantIds,
              classroomId: effectiveClassroom.id || undefined,
            }
          })
          .filter(item => item.lessonDate && item.startTime && item.endTime && item.teacherId)
      })

      if (!schedules.length) {
        if (cacheVersion === classConflictCacheVersion)
          classConflictCache.set(cacheKey, null)
        return null
      }

      const res = await validateGroupClassSchedulesApi({
        groupClassId: classInfo.id,
        teacherId: '',
        classroomId: effectiveClassroom.id || undefined,
        schedules,
      })
      if (res.code !== 200)
        throw new Error(res.message || '检测班课空位失败')
      const validation = res.result || null
      if (cacheVersion === classConflictCacheVersion)
        classConflictCache.set(cacheKey, validation)
      return validation
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

  function buildValidationItemKey(teacherId: unknown, lessonDate: unknown, startTime: unknown, endTime: unknown) {
    return [
      String(teacherId || '').trim(),
      String(lessonDate || '').trim(),
      String(startTime || '').trim(),
      String(endTime || '').trim(),
    ].join('|')
  }

  function applyClassValidationResult(classInfo: ClassInfo, validation: TeachingScheduleValidationResult | null) {
    options.resetEmptyLessonConflicts()
    const invalidMap = new Map(
      (Array.isArray(validation?.items) ? validation.items : [])
        .filter(item => item?.valid === false)
        .map(item => [
          buildValidationItemKey(item.teacherId, item.lessonDate, item.startTime, item.endTime),
          item,
        ]),
    )

    options.dataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson: any) => {
        if (options.hasScheduledLesson(lesson))
          return
        const matched = invalidMap.get(
          buildValidationItemKey(teacher?.teacherId, teacher?.date, lesson?.startTime, lesson?.endTime),
        )
        lesson.conflict = Boolean(matched)
        lesson.conflictReason = matched
          ? {
              type: 'group-class-api',
              className: classInfo.name,
              lessonIndex: options.getLessonIndex(lesson?.startTime),
              date: teacher?.date ? `${dayjs(teacher.date).format('M')}月${dayjs(teacher.date).format('D')}日` : '',
              time: `${String(lesson?.startTime || '').trim()}-${String(lesson?.endTime || '').trim()}`,
              message: matched.message || '当前空位不可排课',
              conflictTypes: Array.isArray(matched.conflictTypes) ? matched.conflictTypes : [],
              existingSchedules: Array.isArray(matched.existingSchedules) ? matched.existingSchedules : [],
              conflictingStudentNames: Array.isArray(matched.conflictingStudentNames) ? matched.conflictingStudentNames : [],
            }
          : null
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

      const validation = await loadClassConflictValidation(classInfo)
      if (seq !== classConflictSeq)
        return classInfo

      applyClassValidationResult(classInfo, validation)
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
