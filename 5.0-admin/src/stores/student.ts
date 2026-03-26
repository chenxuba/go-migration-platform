import { defineStore } from 'pinia'

export const useStudentStore = defineStore('student', {
  state: () => ({
    studentId: null as string | null,
  }),
  actions: {
    setStudentId(id: string) {
      this.studentId = id
    },
    clearStudentId() {
      this.studentId = null
    },
  },
})
