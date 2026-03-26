export interface CourseListInfo {
  name: string
  courseCategory: number
  courseAttribute: number
  type: number
  title: string
  images: string
  description: string
  isShowMicoSchool: boolean
  buyRule: {
    enableBuyLimit: boolean
    isAllowReturningStudent: boolean
    relateProductIds: number[]
    studentStatuses: number[]
  }
  courseProductProperties: {
    coursePropertyId: number
    propertyIdName: string
    coursePropertyValue: number
    propertyValueName: string
  }[]
  courseScope: string
  allowedLessonIds: number[]
}

// 课程属性 获取课程属性列表
export function getCoursePropertyListApi(data: CourseListInfo) {
  return useGet<CourseListInfo>('/api/v1/course-properties', data)
}
// 课程属性 更新课程属性 /instCourseProperty/updateCoursePropertyEnable
export function updateCoursePropertyEnableApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/course-properties/update', data)
}
// 查询课程下拉属性 /instCourseProperty/getCoursePropertyOptions
export function getCoursePropertyOptionsApi(data: CourseListInfo) {
  return useGet<CourseListInfo>('/api/v1/course-property-options', data)
}
// 保存下拉属性 /instCourseProperty/addCoursePropertyOption
export function addCoursePropertyOptionApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/course-property-options/create', data)
}
// 删除下拉属性 /instCoursePropertyOption/deleteOption
export function deleteOptionApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/course-property-options/delete', data)
}
// 更新下拉属性 /instCoursePropertyOption/updateOption
export function updateCoursePropertyOptionApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/course-property-options/update', data)
}
// 批量更新选项排序 /instCoursePropertyOption/updateOptionSort
export function updateOptionSortApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/course-property-options/sort', data)
}
// 查询课程类别 /instCourseCategory/getCourseCategoryPage
export function getCourseCategoryPageApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/course-categories/page', data)
}
// 添加课程类别 /instCourseCategory/addCourseCategory
export function addCourseCategoryApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/course-categories/create', data)
}
// 删除课程类别 /instCourseCategory/deleteCourseCategory
export function deleteCourseCategoryApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/course-categories/delete', data)
}
// 修改课程类别 /instCourseCategory/updateCourseCategory
export function updateCourseCategoryApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/course-categories/update', data)
}
// 分页查询课程列表 /instCourse/getCoursePage
export function getCoursePageApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/courses/page', data)
}
// 新增课程 /instCourse/addCourse
export function addCourseApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/courses/create', data)
}
// 删除课程 /instCourse/batchDelOrResCourse
export function batchDelOrResCourseApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/courses/delete-restore', data)
}
// 批量售卖/停售 /instCourse/batchSaleStatus
export function batchSaleStatusApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/courses/sale-status', data)
}
// 批量开启微校/关闭微校售卖 /instCourse/batchOpenMicroSchoolShow
export function batchOpenMicroSchoolShowApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/courses/micro-school-show', data)
}
// 获取课程详情 /instCourse/getCourseDetail
export function getCourseDetailApi(id: number) {
  return useGet<CourseListInfo>(`/api/v1/courses/detail?id=${id}`)
}
// 更新编辑课程信息 /instCourse/updateCourse
export function updateCourseApi(data: CourseListInfo) {
  return usePost<CourseListInfo>('/api/v1/courses/update', data)
}
