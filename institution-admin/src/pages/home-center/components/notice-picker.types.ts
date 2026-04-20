export interface NoticePickerSelection {
  sourceType: 'class' | 'one_to_one'
  sourceId: string
  sourceName: string
  studentId: string
  studentName: string
  tuitionAccountId?: string
  isBind: boolean
  selectionType?: 'student' | 'source'
}

export interface NoticePickerSource {
  sourceType: 'class' | 'one_to_one'
  sourceId: string
  sourceName: string
  students: NoticePickerSelection[]
}

export interface NoticePickerCompletePayload {
  selectedStudents: NoticePickerSelection[]
  selectedSources: NoticePickerSource[]
}
