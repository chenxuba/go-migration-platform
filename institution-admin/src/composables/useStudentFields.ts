import { ref } from 'vue'
import type { FieldInfo } from '~@/api/edu-center/student-list'
import { getStuCustomFieldApi, getStuDefaultFieldApi } from '~@/api/edu-center/student-list'

interface ExtendedFieldInfo extends FieldInfo {
  isDisplay: boolean
  optionsList?: Array<{
    value: string
    id: string
  }>
}

interface ApiResponse {
  result: ExtendedFieldInfo[]
  [key: string]: any
}

export function useStudentFields() {
  const systemDefaultIsDisplayList = ref<ExtendedFieldInfo[]>([])
  const customIsDisplayList = ref<ExtendedFieldInfo[]>([])
  const customIsDisplaySearchList = ref<ExtendedFieldInfo[]>([])

  // 生成唯一ID的函数
  const generateId = () => {
    return `_${Math.random().toString(36).substr(2, 9)}`
  }

  // 创建过滤函数
  const filterFields = (items, filter) => {
    const baseCondition = item => item.isDisplay

    if (filter === 1) {
      return items.filter(item => baseCondition(item) && item.searched)
    }
    else if (filter === 2) {
      return items.filter(item => baseCondition(item) && !item.searched)
    }
    return items.filter(baseCondition)
  }

  // 获取所有学员字段（系统默认和自定义）
  const getAllStuFields = async (filter: { filter: number } = { filter: 1 }) => {
    try {
      // 并行请求两个API
      const [defaultRes, customRes] = await Promise.all([
        getStuDefaultFieldApi({} as FieldInfo),
        getStuCustomFieldApi({} as FieldInfo),
      ]) as unknown as [ApiResponse, ApiResponse]

      // 应用过滤条件
      systemDefaultIsDisplayList.value = filterFields(defaultRes.result, filter.filter)
      customIsDisplayList.value = filterFields(customRes.result, filter.filter)
      customIsDisplaySearchList.value = filterFields(customRes.result, 1)

      // 把 optionsJson: "自闭症,发育迟缓,言语障碍,肢体障碍" 转成数组
      customIsDisplayList.value.forEach((item) => {
        if (item.optionsJson) {
          item.optionsList = item.optionsJson.split(',').map(value => ({
            value,
            id: generateId(),
          }))
        }
      })
    }
    catch (error: any) {
      // 特别处理请求取消错误
      if (error.name === 'CanceledError' || error.code === 'ERR_CANCELED') {
        console.log('请求被取消，这通常是由于组件卸载或重复请求导致的')
        return // 静默处理取消错误，不显示错误提示
      }
      
      console.log('获取学员字段失败:', error)
      // 可以在这里添加错误处理，比如显示错误提示等
    }
  }
  // 获取自定义字段，只取isDisplay为true的
  const getCustomField = async () => {
    const res = await getStuCustomFieldApi({} as FieldInfo)
    return res.result.filter((item: { isDisplay: any }) => item.isDisplay)
  }

  // onMounted(() => {
  //   getAllStuFields()
  // })

  return {
    systemDefaultIsDisplayList,
    customIsDisplayList,
    customIsDisplaySearchList,
    getAllStuFields,
    getCustomField,
  }
}
