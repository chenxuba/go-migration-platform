<script setup>
import { computed, onUnmounted, ref } from 'vue'
import { debounce } from 'lodash-es'
import { getRecommenderPageApi } from '~@/api/enroll-center/intention-student'

// Props
const props = defineProps({
  modelValue: {
    type: [String, Number],
    default: undefined,
  },
  placeholder: {
    type: String,
    default: '搜索姓名/手机号',
  },
  width: {
    type: String,
    default: '360px',
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  // 是否允许清除
  allowClear: {
    type: Boolean,
    default: false,
  },
  // 学生状态过滤 0-意向学员 1-在读学员 undefined-全部
  studentStatus: {
    type: Number,
    default: undefined,
  },
})

// Emits
const emit = defineEmits([
  'update:modelValue',
  'change',
  'select',
])

// 学员列表
const stuListOptions = ref([])

// 强制刷新组件的key
const selectKey = ref(0)

// 分页参数
const pagination = ref({
  current: 1,
  pageSize: 10,
  total: 0,
})

const finished = ref(false)
const isLoading = ref(false)

// 创建防抖的搜索函数
const debouncedSearch = debounce((value) => {
  pagination.value.current = 1
  finished.value = false
  isLoading.value = false

  // 如果是搜索操作（有关键词），清空当前列表
  if (value && value.trim()) {
    stuListOptions.value = []
  }

  getStudentListPage({ key: value })
}, 300) // 300ms 的防抖延迟

// 获取学生列表
async function getStudentListPage(params = { key: undefined }) {
  try {
    if (finished.value || isLoading.value) return
    isLoading.value = true

    const res = await getRecommenderPageApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: pagination.value.pageSize,
        pageIndex: pagination.value.current,
        skipCount: 0,
      },
      queryModel: {
        searchKey: params.key,
        studentStatus: props.studentStatus,
      },
      sortModel: {},
    })

    if (res.code === 200) {
      const resultData = Array.isArray(res.result) ? res.result : []
      // 保留首次加载的清空逻辑
      if (pagination.value.current === 1) {
        stuListOptions.value = resultData
      } else {
        // 合并数据时去重，避免重复的 key
        const existingIds = new Set(stuListOptions.value.map(item => item.id))
        const uniqueNewItems = resultData.filter(item => !existingIds.has(item.id))
        stuListOptions.value = [...stuListOptions.value, ...uniqueNewItems]
      }
      pagination.value.total = Number(res.total || 0)
      const reachedEndByTotal = pagination.value.total > 0 && stuListOptions.value.length >= pagination.value.total
      const reachedEndByPage = resultData.length < pagination.value.pageSize
      finished.value = pagination.value.total === 0 || reachedEndByTotal || reachedEndByPage
    }
  } catch (error) {
    console.error('加载学生数据失败:', error)
    // 发生错误时回退页码
    if (pagination.value.current > 1) {
      pagination.value.current -= 1
    }
  } finally {
    isLoading.value = false
  }
}

// 处理下拉框显示状态变化
function handleDropdownVisibleChange(visible) {
  if (visible) {
    // 重置状态
    pagination.value.current = 1
    isLoading.value = false
    // 如果没有数据或需要刷新，则加载数据
    if (stuListOptions.value.length === 0) {
      finished.value = false
      getStudentListPage()
    } else {
      // 如果已经有数据，检查是否已经加载完成
      finished.value = stuListOptions.value.length >= pagination.value.total
    }
  } else {
    setTimeout(() => {
      // 关闭下拉框时强制刷新组件
      selectKey.value += 1
      // 重置所有状态
      pagination.value.current = 1
      isLoading.value = false
      finished.value = false
    }, 300)
  }
}

// 处理滚动加载
function handlePopupScroll(event) {
  const { target } = event
  const { scrollTop, scrollHeight, clientHeight } = target
  // 判断是否滚动到底部
  if (scrollHeight - scrollTop - clientHeight < 10) {
    // 检查是否正在加载且还有更多数据
    if (!isLoading.value && pagination.value.current * pagination.value.pageSize < pagination.value.total) {
      pagination.value.current += 1
      getStudentListPage()
    }
  }
}

// 处理搜索
function handleSearch(value) {
  debouncedSearch(value)
}

// 处理选择变化
function handleChange(value) {
  emit('update:modelValue', value)
  emit('change', value)
  
  // 如果选择了学员，可以获取选中学员的完整信息
  if (value) {
    const selectedStudent = stuListOptions.value.find(item => item.id === value)
    emit('select', selectedStudent)
  }
}

// 获取当前选中的学员信息
const selectedStudent = computed(() => {
  if (!props.modelValue) return null
  return stuListOptions.value.find(item => item.id === props.modelValue)
})

// 根据学生ID查询学生信息（用于外部调用）
async function loadStudentById(studentId) {
  if (!studentId) return null
  
  try {
    const res = await getRecommenderPageApi({
      pageRequestModel: {
        needTotal: true,
        pageSize: 1,
        pageIndex: 1,
        skipCount: 0,
      },
      queryModel: {
        studentId: studentId,
      },
      sortModel: {},
    })

    if (res.code === 200 && res.result && res.result.length > 0) {
      const student = res.result[0]
      // 将学生添加到选项列表中
      if (!stuListOptions.value.find(item => item.id === student.id)) {
        stuListOptions.value.unshift(student)
      }
      return student
    }
  } catch (error) {
    console.error('加载指定学员失败:', error)
  }
  return null
}

// 重置组件状态
function reset() {
  stuListOptions.value = []
  pagination.value.current = 1
  finished.value = false
  isLoading.value = false
  selectKey.value += 1
}

// 组件卸载时取消防抖
onUnmounted(() => {
  debouncedSearch.cancel()
})

// 暴露方法给父组件
defineExpose({
  loadStudentById,
  reset,
  selectedStudent,
})
</script>

<template>
  <a-select
    :key="selectKey"
    :value="modelValue"
    :filter-option="false"
    :disabled="disabled"
    :allow-clear="allowClear"
    show-search
    :placeholder="placeholder"
    :style="{ width }"
    option-label-prop="label"
    @update:value="handleChange"
    @change="handleChange"
    @dropdown-visible-change="handleDropdownVisibleChange"
    @popup-scroll="handlePopupScroll"
    @search="handleSearch"
  >
    <a-select-option 
      v-for="item in stuListOptions" 
      :key="item.id" 
      :value="item.id" 
      :data="item"
      :label="item.stuName"
    >
      <div class="flex flex-center mb-1 justify-between">
        <div class="flex flex-center">
          <div>
            <img 
              class="w-10 rounded-10"
              :src="item.avatarUrl || 'https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png'"
              alt=""
            >
          </div>
          <div class="ml-2 mr-3">
            <div class="text-sm text-#666 leading-7">
              {{ item.stuName }}
            </div>
            <div class="text-xs text-#888">
              {{ item.mobile }}
            </div>
          </div>
        </div>
        <div>
          <a-tag v-if="item.studentStatus === 1" :bordered="false" color="processing">
            在读学员
          </a-tag>
          <a-tag v-else-if="item.studentStatus === 0" :bordered="false" color="orange">
            意向学员
          </a-tag>
          <a-tag v-else-if="item.studentStatus === 2" :bordered="false" color="default">
            历史学员
          </a-tag>
        </div>
      </div>
    </a-select-option>
    
    <!-- 没有更多了 -->
    <a-select-option 
      v-if="finished && stuListOptions.length > 0" 
      key="no-more" 
      :value="-1"
      :label="undefined"
    >
      <div class="text-center text-#999 text-3">
        ～没有更多了～
      </div>
    </a-select-option>
    
    <!-- 加载中 -->
    <a-select-option 
      v-if="isLoading" 
      key="loading" 
      :value="-2" 
      :label="undefined"
    >
      <div class="text-center text-#999 text-3">
        <a-spin />
      </div>
    </a-select-option>
  </a-select>
</template>

<style lang="less" scoped>
// 组件内部样式可以根据需要添加
</style> 
