interface CompactableScheduleRow {
  lessonDate: string
  startTime: string
  endTime: string
  teacherId?: string
  assistantIds?: Array<string | number>
  classroomId?: string
}

function normalizeStringId(value?: string | number | null) {
  return String(value ?? '').trim()
}

function normalizeStringIds(values?: Array<string | number>) {
  const result: string[] = []
  const seen = new Set<string>()
  ;(values || []).forEach((value) => {
    const normalized = normalizeStringId(value)
    if (!normalized || seen.has(normalized))
      return
    seen.add(normalized)
    result.push(normalized)
  })
  return result
}

function commonScalarValue(values: string[]) {
  if (!values.length)
    return ''
  const first = values[0]
  return values.every(value => value === first) ? first : ''
}

function assistantSignature(values: string[]) {
  return [...values].sort().join('|')
}

function commonAssistantIds(rows: Array<{ assistantIds: string[] }>) {
  if (!rows.length)
    return []
  const first = rows[0].assistantIds
  const signature = assistantSignature(first)
  return rows.every(row => assistantSignature(row.assistantIds) === signature) ? first : []
}

function sameAssistantIds(left: string[], right: string[]) {
  return assistantSignature(left) === assistantSignature(right)
}

export function compactTeachingScheduleAssignments<T extends CompactableScheduleRow>(rows: T[]) {
  const normalizedRows = rows.map(row => ({
    ...row,
    teacherId: normalizeStringId(row.teacherId) || undefined,
    assistantIds: normalizeStringIds(row.assistantIds),
    classroomId: normalizeStringId(row.classroomId) || undefined,
  }))

  const teacherId = commonScalarValue(normalizedRows.map(row => row.teacherId || ''))
  const classroomId = commonScalarValue(normalizedRows.map(row => row.classroomId || ''))
  const assistantIds = commonAssistantIds(normalizedRows)

  return {
    teacherId,
    assistantIds,
    classroomId,
    schedules: normalizedRows.map((row) => {
      return {
        ...row,
        teacherId: row.teacherId === teacherId ? undefined : row.teacherId,
        assistantIds: sameAssistantIds(row.assistantIds, assistantIds) ? undefined : row.assistantIds,
        classroomId: (row.classroomId || '') === classroomId ? undefined : row.classroomId,
      }
    }),
  }
}
