<template>
  <a-select
    :value="modelValue"
    :placeholder="placeholder"
    :style="{ width: width }"
    :filter-option="false"
    :allow-clear="allowClear"
    :loading="initialLoading"
    :mode="multiple ? 'multiple' : undefined"
    show-search
    option-label-prop="label"
    @update:value="handleValueChange"
    @dropdown-visible-change="handleDropdownVisibleChange"
    @popup-scroll="handlePopupScroll"
    @search="handleSearch"
  >
    <!-- 初始加载状态 -->
    <a-select-option 
      v-if="initialLoading" 
      key="initial-loading" 
      :value="-3" 
      :label="undefined"
    >
      <div class="text-center text-#999 text-3 py-2">
        <a-spin size="small" />
        <span class="ml-2">加载中...</span>
      </div>
    </a-select-option>
    
    <!-- 正常选项 -->
    <a-select-option
      v-for="item in staffOptions"
      :key="item.id"
      :value="item.id"
      :data="item"
      :label="item.nickName"
      :disabled="props.fetchType === 'approval' && item.status === 2"
    >
      <div class="flex justify-between flex-items-center">
        <span :class="{ 'text-#bbb': props.fetchType === 'approval' && item.status === 2 }">{{ item.nickName }}</span>
        <span class="text-#999 text-3">{{ item.mobile }}</span>
      </div>
    </a-select-option>
    
    <!-- 当前选中但不在选项中的员工 -->
    <a-select-option
      v-if="selectedStaffDisplay && !staffOptions.find(item => item.id === selectedStaffDisplay.id)"
      :key="selectedStaffDisplay.id"
      :value="selectedStaffDisplay.id"
      :data="selectedStaffDisplay"
      :label="selectedStaffDisplay.nickName"
    >
      <div class="flex justify-between flex-items-center">
        <span>{{ selectedStaffDisplay.nickName }}</span>
        <span class="text-#999 text-3">{{ selectedStaffDisplay.mobile }}</span>
      </div>
    </a-select-option>
    
    <!-- 滚动加载中 -->
    <a-select-option 
      v-if="scrollLoading && staffOptions.length > 0" 
      key="scroll-loading" 
      :value="-2" 
      :label="undefined"
    >
      <div class="text-center text-#999 text-3 py-2">
        <a-spin size="small" />
        <span class="ml-2">加载更多...</span>
      </div>
    </a-select-option>
    
    <!-- 没有更多了 -->
    <a-select-option 
      v-if="finished && staffOptions.length > 0 && !scrollLoading" 
      key="no-more" 
      :value="-1"
      :label="undefined"
    >
      <div class="text-center text-#999 text-3 py-1">
        ～没有更多了～
      </div>
    </a-select-option>
    
    <!-- 暂无数据 -->
    <a-select-option 
      v-if="!initialLoading && !scrollLoading && staffOptions.length === 0" 
      key="no-data" 
      :value="-4"
      :label="undefined"
    >
      <div class="text-center text-#999 text-3 py-2">
        暂无数据
      </div>
    </a-select-option>
  </a-select>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { debounce } from 'lodash-es'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import { getStaffSummariesApi } from '@/api/finance-center/approval-manage'
import { useUserStore } from '@/stores/user'
const userStore = useUserStore()
const userInfo = computed(() => {
  return userStore.userInfo
})
// Props
const props = defineProps({
  modelValue: {
    type: [String, Number, Array],
    default: undefined,
  },
  placeholder: {
    type: String,
    default: '请选择员工',
  },
  width: {
    type: String,
    default: '240px',
  },
  allowClear: {
    type: Boolean,
    default: true,
  },
  // 员工状态筛选，0-在职，1-离职，undefined-全部
  status: {
    type: Number,
    default: 0,
  },
  // 每页数量
  pageSize: {
    type: Number,
    default: 10,
  },
  // 是否多选
  multiple: {
    type: Boolean,
    default: false,
  },
  // 数据来源
  fetchType: {
    type: String,
    default: 'inst-user',
  },
})

// Emits
const emit = defineEmits(['update:modelValue', 'change'])

// 响应式数据
const staffOptions = ref([])
const pagination = ref({
  current: 1,
  pageSize: props.pageSize,
  total: 0,
})
const finished = ref(false)
const initialLoading = ref(false) // 初始加载状态
const scrollLoading = ref(false) // 滚动加载状态
const hasLoadedList = ref(false) // 是否已经加载过员工列表

// 当前选中员工的显示信息
const selectedStaffDisplay = ref(null)

// 搜索防抖函数
const debouncedSearch = debounce((value) => {
  pagination.value.current = 1
  finished.value = false
  scrollLoading.value = false
  initialLoading.value = true

  // 如果是搜索操作（有关键词），清空当前列表
  if (value && value.trim()) {
    staffOptions.value = []
  }

  getStaffList({ searchKey: value })
}, 300)

// 获取员工列表
async function getStaffList(params = { searchKey: undefined }) {
  try {
    if (finished.value) return

    let res
    if (props.fetchType === 'approval') {
      res = await getStaffSummariesApi({
        queryModel: {
          searchKey: params.searchKey,
        },
        pageRequestModel: {
          needTotal: true,
          pageSize: pagination.value.pageSize,
          pageIndex: pagination.value.current,
          skipCount: 0,
        },
      })
    }
    else {
      const requestParams = {
        pageRequestModel: {
          needTotal: true,
          pageSize: pagination.value.pageSize,
          pageIndex: pagination.value.current,
          skipCount: 0,
        },
        queryModel: {
          status: props.status,
          searchKey: params.searchKey,
        },
      }
      res = await getUserListApi(requestParams)
    }

    if (res.code === 200) {
      const resultItems = props.fetchType === 'approval'
        ? (res.result?.list || []).map(item => ({
            id: item.id,
            nickName: item.name,
            mobile: item.phone,
            status: item.status,
          }))
        : (res.result || [])
      hasLoadedList.value = true // 标记已加载过员工列表
      
      // 保留首次加载的清空逻辑
      if (pagination.value.current === 1) {
        const existingSelected = selectedStaffDisplay.value && props.modelValue === selectedStaffDisplay.value.id ? [selectedStaffDisplay.value] : []
        const newItems = resultItems
        // 合并已选中的员工和新加载的员工，去重
        const existingIds = new Set(existingSelected.map(item => item.id))
        const uniqueNewItems = newItems.filter(item => !existingIds.has(item.id))
        staffOptions.value = [...existingSelected, ...uniqueNewItems]
      } else {
        // 合并数据时去重，避免重复的 key
        const newItems = resultItems
        const existingIds = new Set(staffOptions.value.map(item => item.id))
        const uniqueNewItems = newItems.filter(item => !existingIds.has(item.id))
        staffOptions.value = [...staffOptions.value, ...uniqueNewItems]
      }
      pagination.value.total = props.fetchType === 'approval' ? (res.result?.total || 0) : (res.total || 0)
      if (staffOptions.value.length >= pagination.value.total) {
        finished.value = true
      }
    }
  } catch (error) {
    console.error('获取员工列表失败:', error)
    // 发生错误时回退页码
    if (pagination.value.current > 1) {
      pagination.value.current -= 1
    }
  } finally {
    initialLoading.value = false
    scrollLoading.value = false
  }
}

// 处理值变化
function handleValueChange(value) {
  // 过滤掉特殊值（加载状态和没有更多的占位项）
  if (props.multiple && Array.isArray(value)) {
    // 多选模式：过滤掉特殊值
    const filteredValue = value.filter(v => v !== -1 && v !== -2 && v !== -3 && v !== -4)
    emit('update:modelValue', filteredValue)
    const selectedStaffs = filteredValue.map(v => staffOptions.value.find(item => item.id === v)).filter(Boolean)
    emit('change', filteredValue, selectedStaffs)
  } else {
    // 单选模式：过滤掉特殊值
    if (value === -1 || value === -2 || value === -3 || value === -4) {
      return
    }
    emit('update:modelValue', value)
    emit('change', value, staffOptions.value.find(item => item.id === value))
  }
}

// 处理下拉菜单打开状态变化
function handleDropdownVisibleChange(visible) {
  if (visible) {
    // 重置状态
    pagination.value.current = 1
    scrollLoading.value = false
    // 如果没有加载过员工列表，则加载数据
    if (!hasLoadedList.value) {
      finished.value = false
      initialLoading.value = true
      getStaffList()
    } else {
      // 如果已经有数据，检查是否已经加载完成
      finished.value = staffOptions.value.length >= pagination.value.total
    }
  } else {
    setTimeout(() => {
      // 关闭下拉框时强制刷新组件
      // 重置所有状态
      staffOptions.value = []
      pagination.value.current = 1
      initialLoading.value = false
      scrollLoading.value = false
      finished.value = false
      hasLoadedList.value = false
      // 保留当前选中的员工信息用于显示
      if (selectedStaffDisplay.value && props.modelValue === selectedStaffDisplay.value.id) {
        staffOptions.value = [selectedStaffDisplay.value]
      }
    }, 300)
  }
}

// 根据员工ID获取员工信息（用于初始化显示）
async function getStaffById(staffId) {
  if (!staffId) return null
  
  try {
    let res
    if (props.fetchType === 'approval') {
      res = await getStaffSummariesApi({
        queryModel: {},
        pageRequestModel: {
          needTotal: true,
          pageSize: 100,
          pageIndex: 1,
          skipCount: 0,
        },
      })
    }
    else {
      res = await getUserListApi({
        pageRequestModel: {
          needTotal: true,
          pageSize: 20,
          pageIndex: 1,
          skipCount: 0,
        },
        queryModel: {
          status: props.status,
        },
      })
    }

    if (res.code === 200) {
      const sourceList = props.fetchType === 'approval'
        ? (res.result?.list || []).map(item => ({
            id: item.id,
            nickName: item.name,
            mobile: item.phone,
            status: item.status,
          }))
        : (res.result || [])
      const staff = sourceList.find(item => item.id === staffId)
      if (staff) {
        return staff
      }
    }
  } catch (error) {
    console.error('获取员工信息失败:', error)
  }
  return null
}

// 处理滚动加载更多
function handlePopupScroll(event) {
  const { target } = event
  const { scrollTop, scrollHeight, clientHeight } = target
  // 判断是否滚动到底部
  if (scrollHeight - scrollTop - clientHeight < 50) {
    // 检查是否正在加载且还有更多数据
    if (!scrollLoading.value && !initialLoading.value && pagination.value.current * pagination.value.pageSize < pagination.value.total) {
      scrollLoading.value = true
      pagination.value.current += 1
      getStaffList()
    }
  }
}

// 搜索
function handleSearch(value) {
  debouncedSearch(value)
}

// 监听 modelValue 变化，初始化显示
watch(() => props.modelValue, async (newValue) => {
  if (props.multiple && Array.isArray(newValue)) {
    // 多选模式：处理每个选中的值
    for (const value of newValue) {
      if (value && !staffOptions.value.find(item => item.id === value)) {
        const staffInfo = await getStaffById(value)
        if (staffInfo && !hasLoadedList.value && !staffOptions.value.find(item => item.id === staffInfo.id)) {
          staffOptions.value = [staffInfo, ...staffOptions.value]
        }
      }
    }
  } else if (newValue && !props.multiple) {
    // 单选模式：保留原有逻辑
    if (!staffOptions.value.find(item => item.id === newValue)) {
      // 如果有值但找不到对应的员工信息，先设置一个占位符
      selectedStaffDisplay.value = {
        id: newValue,
        nickName: '加载中...',
        mobile: ''
      }
      
      // 尝试获取员工信息
      const staffInfo = await getStaffById(newValue)
      console.log(staffInfo);
      
      if (staffInfo) {
        selectedStaffDisplay.value = staffInfo
        // 只在还没有加载员工列表时才添加到选项中
        if (!hasLoadedList.value && !staffOptions.value.find(item => item.id === staffInfo.id)) {
          staffOptions.value = [staffInfo, ...staffOptions.value]
        }
      } else {
        // 如果获取失败，显示ID
        selectedStaffDisplay.value = {
          id: newValue,
          nickName: userInfo.value.nickName,
          mobile: userInfo.value.mobile
        }
      }
    } else {
      // 如果在现有选项中找到了，更新显示
      const found = staffOptions.value.find(item => item.id === newValue)
      if (found) {
        selectedStaffDisplay.value = found
      }
    }
  } else if (!newValue) {
    selectedStaffDisplay.value = null
  }
}, { immediate: true })

// 暴露方法供父组件调用
defineExpose({
  refresh: () => {
    staffOptions.value = []
    pagination.value.current = 1
    finished.value = false
    initialLoading.value = true
    scrollLoading.value = false
    hasLoadedList.value = false
    getStaffList()
  },
  getSelectedStaff: () => {
    if (props.multiple && Array.isArray(props.modelValue)) {
      return staffOptions.value.filter(item => props.modelValue.includes(item.id))
    }
    return staffOptions.value.find(item => item.id === props.modelValue) || selectedStaffDisplay.value
  },
})
</script>

<style lang="less" scoped>
// 组件样式可以根据需要添加
</style> 
