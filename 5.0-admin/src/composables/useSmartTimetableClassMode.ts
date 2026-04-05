import dayjs from 'dayjs'
import { ref } from 'vue'
import type { ComputedRef } from 'vue'

interface ClassInfo {
  id: string
  name: string
  studentIds: string[]
  studentNames: string[]
  courseId: string
  courseName: string
  mainTeacherId: string
  mainTeacherName: string
}

interface UseSmartTimetableClassModeOptions {
  activeGroupLabel: ComputedRef<string>
  allDataSource: ComputedRef<any[]>
  dataSource: ComputedRef<any[]>
  getLessonIndex: (startTime: string) => string | number
  resetEmptyLessonConflicts: (scope?: string) => void
}

export function useSmartTimetableClassMode(options: UseSmartTimetableClassModeOptions) {
  const classData = ref<ClassInfo[]>([
    {
      id: 'C-01',
      name: '苹果基础班',
      studentIds: ['589250903194799104', '5892509031876223323', '10001'],
      studentNames: ['陈陈', '晨晨', '张三'],
      courseId: '589251114063479808',
      courseName: '初级认知课',
      mainTeacherId: 't001',
      mainTeacherName: '张老师',
    },
    {
      id: 'C-02',
      name: '橙子基础班',
      studentIds: ['20004', '20009', '5892509031876223323'],
      studentNames: ['张四', '王九', '晨晨'],
      courseId: '589251121574791084',
      courseName: '初级认知课',
      mainTeacherId: 't003',
      mainTeacherName: '李老师',
    },
  ])

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

  function findClassInfo(value: unknown) {
    const normalized = String(value || '').trim()
    if (!normalized)
      return null
    return classData.value.find(item => item.id === normalized) || null
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
    console.log('运行班课冲突检测', classInfo)

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

    console.log('班级已排课时间段', classExistingLessons)

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
          console.log('班级跨组交叉时段冲突', classInfo.name, currentTime.date, currentTime.startTime)
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
            console.log('教师已有其他班级课程', teacher.name, currentTime.startTime)
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

        if (!hasConflict && classInfo.studentIds?.length > 0) {
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
                if (sameTimeLesson.studentId?.includes(sid)) {
                  console.log('学生时间冲突', currentTime.date, currentTime.startTime, sameTimeLesson.startTime)
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

  function checkClassExistingTeacherRole(classId: string, teacherId: string, startTime: string, endTime: string) {
    console.log('检查班级主教/辅教角色', classId, teacherId, startTime)

    const classInfo = findClassInfo(classId)
    if (!classInfo) {
      console.log('未找到班级信息，默认设置为主教')
      return { hasExistingArrangement: false, isMainTeacher: true }
    }

    const isMainTeacher = classInfo.mainTeacherId === teacherId

    console.log('根据班级配置判断角色:', isMainTeacher ? '主教' : '辅教')
    console.log('班级配置的主教ID:', classInfo.mainTeacherId, '当前老师ID:', teacherId)

    let hasExistingArrangement = false

    options.allDataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson: any) => {
        if (lesson.startTime === startTime && lesson.endTime === endTime && lesson.classId === classId)
          hasExistingArrangement = true
      })
    })

    console.log('是否已有该班级课程安排:', hasExistingArrangement)
    console.log('最终角色设置:', isMainTeacher ? '主教' : '辅教')
    return { hasExistingArrangement, isMainTeacher }
  }

  function handleClass(value: unknown) {
    if (!value) {
      options.resetEmptyLessonConflicts()
      return
    }

    const classInfo = findClassInfo(value)
    if (!classInfo)
      return

    console.log('选择班级', classInfo.name)
    checkClassCrossTimeConflicts(classInfo)
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

  function applyClassSchedule(params: {
    classInfo: ClassInfo
    column: any
    record: any
  }) {
    const { classInfo, column, record } = params
    const targetTeacher = options.dataSource.value.find(
      teacher => teacher.teacherId === record.teacherId && teacher.date === record.date,
    )

    if (!targetTeacher)
      return false

    const columnIndex = column?.dataIndex?.[1]
    const targetLesson = targetTeacher.lessons?.[columnIndex]
    if (!targetLesson)
      return false

    const { isMainTeacher } = checkClassExistingTeacherRole(
      classInfo.id,
      String(record.teacherId || ''),
      targetLesson.startTime,
      targetLesson.endTime,
    )

    Object.assign(targetLesson, {
      classId: classInfo.id,
      className: classInfo.name,
      courseName: classInfo.courseName,
      courseType: 2,
      isMain: isMainTeacher,
      studentNames: classInfo.studentNames.map(name => ({ name })),
      studentId: classInfo.studentIds,
      conflict: false,
      conflictReason: null,
      serverConflict: false,
      serverConflictReason: null,
    })

    console.log('更新课程信息完成', targetLesson)
    checkClassCrossTimeConflicts(classInfo)
    return true
  }

  return {
    applyClassSchedule,
    classData,
    findClassInfo,
    handleClass,
    resolveClassConflictMessage,
    resolveSelectedClassTarget,
  }
}
