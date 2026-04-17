import { ref } from 'vue'
import type { CourseListInfo } from '~@/api/edu-center/course-list'
import { getCoursePropertyListApi } from '~@/api/edu-center/course-list'

interface CoursePropertyInfo {
  id: number
  name: string
  enable: boolean
  version?: number
  [key: string]: any
}

interface ApiResponse {
  result: CoursePropertyInfo[]
  code: number
  message?: string
  [key: string]: any
}

export function useCourseAttribute() {
  const allCourseProperties = ref<CoursePropertyInfo[]>([])
  const enabledCourseProperties = ref<CoursePropertyInfo[]>([])

  // 获取全部课程属性数组
  const getAllCourseProperties = async () => {
    try {
      const res = await getCoursePropertyListApi({} as CourseListInfo) as unknown as ApiResponse

      if (res.code === 200) {
        allCourseProperties.value = res.result
        // 过滤出已开启的课程属性
        enabledCourseProperties.value = res.result.filter(item => item.enable)
      }
      else {
        console.error('获取课程属性失败:', res.message)
      }
    }
    catch (error) {
      console.error('获取课程属性失败:', error)
    }
  }

  // 获取已开启的课程属性数组（独立调用）
  const getEnabledCourseProperties = async () => {
    try {
      const res = await getCoursePropertyListApi({} as CourseListInfo) as unknown as ApiResponse

      if (res.code === 200) {
        enabledCourseProperties.value = res.result.filter(item => item.enable)
      }
      else {
        console.error('获取已开启课程属性失败:', res.message)
      }
    }
    catch (error) {
      console.error('获取已开启课程属性失败:', error)
    }
  }

  // 刷新课程属性列表
  const refreshCourseProperties = async () => {
    await getAllCourseProperties()
  }

  return {
    allCourseProperties,
    enabledCourseProperties,
    getAllCourseProperties,
    getEnabledCourseProperties,
    refreshCourseProperties,
  }
}
