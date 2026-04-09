import dayjs from 'dayjs'
import { ref } from 'vue'
import type { ComputedRef } from 'vue'
import { getGroupClassDetailApi, listGroupClassStudentsByClassIdsApi, pageGroupClassesApi } from '@/api/edu-center/group-class'

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
  allDataSource: ComputedRef<any[]>
  dataSource: ComputedRef<any[]>
  getLessonIndex: (startTime: string) => string | number
  resetEmptyLessonConflicts: (scope?: string) => void
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
    classroomId: String(value.classroomId || '').trim(),
    classroomName: String(value.classroomName || '').trim(),
    teacherIds: Array.isArray(value.teacherIds) ? value.teacherIds.map(item => String(item || '').trim()).filter(Boolean) : [],
    detailLoaded: value.detailLoaded === true,
  }
}

export function useSmartTimetableClassMode(options: UseSmartTimetableClassModeOptions) {
  const classData = ref<ClassInfo[]>([])
  const classListLoading = ref(false)
  const classDetailLoading = ref(false)
  const pendingClassLoads = new Map<string, Promise<ClassInfo | null>>()

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
          classroomId: String(detail?.classroomId ?? ''),
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

  function resolveSelectedClassTarget(value: unknown) {
    const selectedClass = findClassInfo(value)
    return {
      modeLabel: '班课',
      targetLabel: '排课班级',
      targetValue: selectedClass?.name || '未选择班级',
      courseName: selectedClass?.courseName || '未选择课程',
    }
  }

  function checkClassCrossTimeConflicts(classInfo: ClassInfo) {
    options.resetEmptyLessonConflicts()

    const classExistingLessons: Array<{
      date: string
      endTime: string
      lessonIndex: string | number
      startTime: string
      teacherId: string
      teacherName: string
    }> = []

    options.allDataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson: any) => {
        if (lesson.classId === classInfo.id) {
          classExistingLessons.push({
            date: teacher.date,
            startTime: lesson.startTime,
            endTime: lesson.endTime,
            teacherName: teacher.name,
            teacherId: teacher.teacherId,
            lessonIndex: options.getLessonIndex(lesson.startTime),
          })
        }
      })
    })

    options.dataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson: any, lessonIndex: number) => {
        if (lesson.studentId)
          return

        const currentTime = {
          date: teacher.date,
          startTime: lesson.startTime,
          endTime: lesson.endTime,
        }

        let hasConflict = false
        let conflictReason = null

        const classTimeConflict = classExistingLessons.find(existingLesson =>
          existingLesson.date === currentTime.date
          && (existingLesson.startTime !== currentTime.startTime
            || existingLesson.endTime !== currentTime.endTime)
          && isTimeOverlap(
            { start: existingLesson.startTime, end: existingLesson.endTime },
            { start: currentTime.startTime, end: currentTime.endTime },
          ),
        )

        if (classTimeConflict) {
          hasConflict = true

          const month = dayjs(classTimeConflict.date).format('M')
          const day = dayjs(classTimeConflict.date).format('D')
          const conflictGroup = options.activeGroupLabel.value

          conflictReason = {
            type: '班级时间段交叉冲突',
            className: classInfo.name,
            date: `${month}月${day}日`,
            lessonIndex: classTimeConflict.lessonIndex,
            teacherName: classTimeConflict.teacherName,
            group: conflictGroup,
            time: `${classTimeConflict.startTime}-${classTimeConflict.endTime}`,
          }
        }

        if (!hasConflict) {
          const teacherOtherLesson = teacher.lessons.find((item: any, idx: number) =>
            idx !== lessonIndex
            && item.courseType === 2
            && item.classId !== classInfo.id
            && isTimeOverlap(
              { start: item.startTime, end: item.endTime },
              { start: currentTime.startTime, end: currentTime.endTime },
            ),
          )

          if (teacherOtherLesson) {
            hasConflict = true

            const month = dayjs(teacher.date).format('M')
            const day = dayjs(teacher.date).format('D')
            conflictReason = {
              type: '教师班课冲突',
              teacherName: teacher.name,
              date: `${month}月${day}日`,
              lessonIndex: options.getLessonIndex(currentTime.startTime),
              className: teacherOtherLesson.className,
              courseName: teacherOtherLesson.courseName,
              time: `${teacherOtherLesson.startTime}-${teacherOtherLesson.endTime}`,
            }
          }
        }

        if (!hasConflict && classInfo.studentIds.length > 0) {
          for (const teacherRow of options.allDataSource.value) {
            if (teacherRow.date !== currentTime.date)
              continue

            const sameTimeLessons = teacherRow.lessons.filter((item: any) =>
              item.studentId
              && isTimeOverlap(
                { start: item.startTime, end: item.endTime },
                { start: currentTime.startTime, end: currentTime.endTime },
              ),
            )

            let matchedStudentConflict = false
            for (const sameTimeLesson of sameTimeLessons) {
              if (sameTimeLesson.classId === classInfo.id)
                continue

              for (const sid of classInfo.studentIds) {
                if (sameTimeLesson.studentId?.includes?.(sid)) {
                  hasConflict = true

                  const studentIndex = classInfo.studentIds.indexOf(sid)
                  const studentName = studentIndex >= 0 ? classInfo.studentNames[studentIndex] : '未知学生'
                  const month = dayjs(teacherRow.date).format('M')
                  const day = dayjs(teacherRow.date).format('D')
                  const conflictGroup = options.activeGroupLabel.value

                  conflictReason = {
                    type: '学生课程冲突',
                    studentName,
                    date: `${month}月${day}日`,
                    lessonIndex: options.getLessonIndex(sameTimeLesson.startTime),
                    teacherName: teacherRow.name,
                    courseName: sameTimeLesson.courseName,
                    className: sameTimeLesson.className,
                    group: conflictGroup,
                    time: `${sameTimeLesson.startTime}-${sameTimeLesson.endTime}`,
                  }

                  matchedStudentConflict = true
                  break
                }
              }

              if (matchedStudentConflict)
                break
            }

            if (matchedStudentConflict)
              break
          }
        }

        lesson.conflict = hasConflict
        lesson.conflictReason = conflictReason
      })
    })
  }

  async function handleClass(value: unknown) {
    if (!value) {
      options.resetEmptyLessonConflicts()
      return null
    }

    const classInfo = await ensureClassLoaded(value)
    if (!classInfo) {
      options.resetEmptyLessonConflicts()
      return null
    }

    checkClassCrossTimeConflicts(classInfo)
    return classInfo
  }

  function resolveClassConflictMessage(reason: any) {
    if (!reason)
      return ''

    const groupInfo = reason.group ? `(${reason.group})` : ''
    const timeInfo = reason.time ? `[${reason.time}]` : ''

    if (reason.type === '教师班课冲突')
      return `该时间段${reason.teacherName}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.className}的${reason.courseName}班课安排，无法排课`
    if (reason.type === '学生课程冲突')
      return `该时间段${reason.studentName}在${reason.date}第${reason.lessonIndex}节课${timeInfo}已有${reason.teacherName}${groupInfo}的${reason.courseName || (`${reason.className}班课`)}课程安排，无法排课`
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
    classDetailLoading,
    classListLoading,
    ensureClassLoaded,
    findClassInfo,
    handleClass,
    loadClassOptions,
    resolveClassConflictMessage,
    resolveSelectedClassTarget,
  }
}
