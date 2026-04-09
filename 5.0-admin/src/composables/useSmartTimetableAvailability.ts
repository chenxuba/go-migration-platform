import { ref } from 'vue'
import type { ComputedRef, Ref } from 'vue'
import { validateOneToOneSchedulesApi } from '@/api/edu-center/teaching-schedule'
import messageService from '@/utils/messageService'

interface UseSmartTimetableAvailabilityOptions {
  assistantNameById: (id: unknown) => string
  buildAvailabilitySlotKey: (teacherId: unknown, lessonDate: string, startTime: string, endTime: string) => string
  currentModel: Ref<string>
  dataSource: ComputedRef<any[]>
  normalizedSelectedAssistantIds: ComputedRef<string[]>
  normalizedSelectedClassroomId: ComputedRef<string>
  oneToOneData: Ref<any[]>
  parseConflictTimeRange: (timeText: unknown) => { startTime: string, endTime: string } | null
  resetEmptyLessonConflicts: (scope?: string) => void
  syncLessonConflictState: (lesson: any) => void
  uniqueConflictTypes: (list: any[]) => any[]
  uniqueExistingSchedules: (list: any[]) => any[]
}

export function useSmartTimetableAvailability(options: UseSmartTimetableAvailabilityOptions) {
  const oneToOneAvailabilityLoading = ref(false)

  let oneToOneAvailabilitySeq = 0

  function buildCurrentOneToOneAvailabilityPayload(oneToOneId: unknown) {
    const classroomId = String(options.normalizedSelectedClassroomId.value || '').trim()
    const assistantIds = options.normalizedSelectedAssistantIds.value
    const schedules: Array<{
      teacherId: string
      lessonDate: string
      startTime: string
      endTime: string
      assistantIds?: string[]
      classroomId?: string
    }> = []

    options.dataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson: any) => {
        if (!lesson.studentId) {
          schedules.push({
            teacherId: String(teacher.teacherId),
            lessonDate: teacher.date,
            startTime: lesson.startTime,
            endTime: lesson.endTime,
            assistantIds,
            classroomId: classroomId || undefined,
          })
        }
      })
    })

    return {
      oneToOneId: String(oneToOneId || ''),
      teacherId: '',
      assistantIds,
      classroomId: classroomId || undefined,
      schedules,
    }
  }

  function applyServerAvailabilityResult(result: any) {
    options.resetEmptyLessonConflicts()
    const invalidMap = new Map()
    const items = Array.isArray(result?.items) ? result.items : []
    items.forEach((item) => {
      if (item?.valid === false) {
        invalidMap.set(
          options.buildAvailabilitySlotKey(item.teacherId, item.lessonDate, item.startTime, item.endTime),
          item,
        )
      }
    })

    options.dataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson: any) => {
        if (lesson.studentId)
          return
        const matched = invalidMap.get(
          options.buildAvailabilitySlotKey(teacher.teacherId, teacher.date, lesson.startTime, lesson.endTime),
        )
        lesson.serverConflict = Boolean(matched)
        lesson.serverConflictReason = matched
          ? {
              type: '1v1-api',
              message: matched.message || '该时间段不可排课',
              conflictTypes: matched.conflictTypes || [],
              existingSchedules: matched.existingSchedules || [],
            }
          : null
        options.syncLessonConflictState(lesson)
      })
    })
  }

  function cancelOneToOneAvailabilityCheck() {
    oneToOneAvailabilitySeq += 1
    oneToOneAvailabilityLoading.value = false
  }

  async function detectOneToOneAvailability(value: string | number | undefined) {
    const seq = ++oneToOneAvailabilitySeq
    const oneToOneId = String(value || '').trim()
    if (!oneToOneId || options.currentModel.value !== '1') {
      oneToOneAvailabilityLoading.value = false
      options.resetEmptyLessonConflicts()
      return
    }

    if (!options.oneToOneData.value.some(item => String(item?.id || '').trim() === oneToOneId)) {
      oneToOneAvailabilityLoading.value = false
      options.resetEmptyLessonConflicts()
      return
    }

    const payload = buildCurrentOneToOneAvailabilityPayload(oneToOneId)
    if (!payload.schedules.length) {
      oneToOneAvailabilityLoading.value = false
      options.resetEmptyLessonConflicts()
      return
    }

    oneToOneAvailabilityLoading.value = true
    try {
      const res = await validateOneToOneSchedulesApi(payload)
      if (seq !== oneToOneAvailabilitySeq)
        return
      if (res.code !== 200 || !res.result)
        throw new Error(res.message || '检测课表空位失败')
      applyServerAvailabilityResult(res.result)
    }
    catch (error: any) {
      if (seq !== oneToOneAvailabilitySeq)
        return
      console.error('detectOneToOneAvailability failed', error)
      options.resetEmptyLessonConflicts()
      messageService.error(error?.response?.data?.message || error?.message || '检测课表空位失败')
    }
    finally {
      if (seq === oneToOneAvailabilitySeq)
        oneToOneAvailabilityLoading.value = false
    }
  }

  return {
    cancelOneToOneAvailabilityCheck,
    detectOneToOneAvailability,
    oneToOneAvailabilityLoading,
  }
}
