import { ref } from 'vue'
import type { ComputedRef, Ref } from 'vue'
import { checkAssistantScheduleAvailabilityApi, checkOneToOneScheduleAvailabilityApi } from '@/api/edu-center/teaching-schedule'
import messageService from '@/utils/messageService'

interface UseSmartTimetableAvailabilityOptions {
  assistantNameById: (id: unknown) => string
  buildAvailabilitySlotKey: (teacherId: unknown, lessonDate: string, startTime: string, endTime: string) => string
  currentModel: Ref<string>
  dataSource: ComputedRef<any[]>
  normalizedSelectedAssistantIds: ComputedRef<string[]>
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

  function buildCurrentOneToOneAvailabilityPayload(oneToOneId: unknown) {
    const schedules: Array<{
      teacherId: string
      lessonDate: string
      startTime: string
      endTime: string
    }> = []

    options.dataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson: any) => {
        if (!lesson.studentId) {
          schedules.push({
            teacherId: String(teacher.teacherId),
            lessonDate: teacher.date,
            startTime: lesson.startTime,
            endTime: lesson.endTime,
          })
        }
      })
    })

    return {
      oneToOneId: String(oneToOneId || ''),
      schedules,
    }
  }

  function buildCurrentAssistantAvailabilityPayload(oneToOneId: unknown) {
    const schedules: Array<{
      lessonDate: string
      startTime: string
      endTime: string
    }> = []

    options.dataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson: any) => {
        if (!lesson.studentId) {
          schedules.push({
            lessonDate: teacher.date,
            startTime: lesson.startTime,
            endTime: lesson.endTime,
          })
        }
      })
    })

    return {
      oneToOneId: String(oneToOneId || ''),
      assistantIds: options.normalizedSelectedAssistantIds.value,
      schedules,
    }
  }

  function buildAssistantConflictReason(issues: any[]) {
    const busyNames = Array.from(
      new Set(
        issues
          .filter(item => item.kind !== 'selection')
          .map(item => String(item.assistantName || '').trim())
          .filter(Boolean),
      ),
    )
    const messageParts = []
    if (busyNames.length)
      messageParts.push(`助教${busyNames.join('、')}该时间段已有安排`)

    const existingSchedules = options.uniqueExistingSchedules(issues.flatMap(item => item.existingSchedules || []))
    const conflictTypes = options.uniqueConflictTypes(['助教', ...issues.flatMap(item => item.conflictTypes || [])])
    return {
      type: existingSchedules.length ? '1v1-api' : '1v1-assistant-selection',
      message: messageParts.join('；') || '所选助教该时间段不可排课',
      conflictTypes,
      existingSchedules,
    }
  }

  function applyServerAvailabilityResult(result: any) {
    options.resetEmptyLessonConflicts('server')
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

  function applyAssistantAvailabilityResult(result: any) {
    options.resetEmptyLessonConflicts('assistant')
    const selectedIds = options.normalizedSelectedAssistantIds.value
    if (!selectedIds.length)
      return

    const invalidItems = (Array.isArray(result?.items) ? result.items : [])
      .filter(item => item?.valid === false)
      .map(item => ({
        assistantId: String(item.assistantId || '').trim(),
        assistantName: String(item.assistantName || options.assistantNameById(item.assistantId) || item.assistantId || '').trim(),
        conflictTypes: options.uniqueConflictTypes(item.conflictTypes || []),
        existingSchedules: Array.isArray(item.existingSchedules) ? item.existingSchedules : [],
      }))

    options.dataSource.value.forEach((teacher) => {
      teacher.lessons.forEach((lesson: any) => {
        if (lesson.studentId)
          return

        const issues: any[] = []

        invalidItems.forEach((item) => {
          if (!item.existingSchedules.length) {
            issues.push({
              ...item,
              kind: 'global',
            })
            return
          }

          const matchedSchedules = item.existingSchedules.filter((schedule: any) => {
            const timeRange = options.parseConflictTimeRange(schedule?.timeText)
            if (!timeRange)
              return false
            return String(schedule?.date || '').trim() === teacher.date
              && isTimeOverlap(
                { start: lesson.startTime, end: lesson.endTime },
                { start: timeRange.startTime, end: timeRange.endTime },
              )
          })

          if (matchedSchedules.length) {
            issues.push({
              ...item,
              kind: 'busy',
              existingSchedules: matchedSchedules,
            })
          }
        })

        lesson.assistantConflict = issues.length > 0
        lesson.assistantConflictReason = issues.length ? buildAssistantConflictReason(issues) : null
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
      const assistantPayload = buildCurrentAssistantAvailabilityPayload(oneToOneId)
      const [res, assistantRes] = await Promise.all([
        checkOneToOneScheduleAvailabilityApi(payload),
        assistantPayload.assistantIds.length
          ? checkAssistantScheduleAvailabilityApi(assistantPayload)
          : Promise.resolve(null),
      ])
      if (seq !== oneToOneAvailabilitySeq)
        return
      if (res.code !== 200 || !res.result)
        throw new Error(res.message || '检测课表空位失败')
      if (assistantRes && (assistantRes.code !== 200 || !assistantRes.result))
        throw new Error(assistantRes.message || '检测助教空闲状态失败')
      applyServerAvailabilityResult(res.result)
      applyAssistantAvailabilityResult(assistantRes?.result)
    }
    catch (error: any) {
      if (seq !== oneToOneAvailabilitySeq)
        return
      console.error('detectOneToOneAvailability failed', error)
      options.resetEmptyLessonConflicts('server')
      options.resetEmptyLessonConflicts('assistant')
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
