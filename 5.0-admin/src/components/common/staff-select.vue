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
      v-if="selectedStaffDisplay && !staffOptions.find(item => sameStaffId(item.id, selectedStaffDisplay.id))"
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
import { ref, watch } from 'vue'
import { debounce } from 'lodash-es'
import { getUserListApi } from '@/api/internal-manage/staff-manage'
import { getStaffSummariesApi } from '@/api/finance-center/approval-manage'
import { findCachedStaff, getCachedInitialStaffList, mergeCachedStaff } from '@/composables/staff-select-cache'

function sameStaffId(a, b) {
  return a != null && b != null && String(a) === String(b)
}

// Props
const props = defineProps({
  modelValue: {
    type: [String, Number, Array],
    default: undefined,
  },
  /** 外部已知的员工展示信息（如班级详情 teachers），用于多选标签立刻显示姓名、避免仅显示 ID */
  presetStaff: {
    type: Array,
    default: () => [],
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
  includeDisabled: {
    type: Boolean,
    default: false,
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

function normalizeStaffItems(res) {
  return props.fetchType === 'approval'
    ? (res.result?.list || []).map(item => ({
        id: item.id,
        nickName: item.name,
        mobile: item.phone,
        status: item.status,
      }))
    : (res.result || [])
}

function effectiveStatus() {
  return props.includeDisabled ? undefined : props.status
}

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

/** 预设/班级接口只有姓名无手机号时，用员工分页里的同 id 记录补齐 mobile */
function mergeMobileFromApiIntoOptions(apiItems) {
  if (!apiItems?.length)
    return
  const byId = new Map(apiItems.map(item => [String(item.id), item]))
  staffOptions.value = staffOptions.value.map((opt) => {
    const src = byId.get(String(opt.id))
    const srcM = src && String(src.mobile ?? '').trim()
    const optM = String(opt.mobile ?? '').trim()
    if (srcM && !optM)
      return { ...opt, mobile: srcM }
    return opt
  })
}

// 获取员工列表
async function getStaffList(params = { searchKey: undefined }) {
  try {
    if (finished.value) return

    let res
    const searchKey = `${params.searchKey || ''}`.trim()
    const status = effectiveStatus()
    const loadStaffPage = async () => {
      if (props.fetchType === 'approval') {
        const response = await getStaffSummariesApi({
          queryModel: {
            searchKey,
          },
          pageRequestModel: {
            needTotal: true,
            pageSize: pagination.value.pageSize,
            pageIndex: pagination.value.current,
            skipCount: 0,
          },
        })
        return {
          res: response,
          items: normalizeStaffItems(response),
          total: response.result?.total || 0,
        }
      }

      const response = await getUserListApi({
        pageRequestModel: {
          needTotal: true,
          pageSize: pagination.value.pageSize,
          pageIndex: pagination.value.current,
          skipCount: 0,
        },
        queryModel: {
          status,
          searchKey,
        },
      })
      return {
        res: response,
        items: normalizeStaffItems(response),
        total: response.total || 0,
      }
    }

    let resultItems = []
    let total = 0
    if (!searchKey && pagination.value.current === 1) {
      const cached = await getCachedInitialStaffList(
        props.fetchType,
        status,
        async () => {
          const loaded = await loadStaffPage()
          return {
            items: loaded.items,
            total: loaded.total,
          }
        },
      )
      resultItems = cached.items
      total = cached.total
      res = { code: 200 }
    } else {
      const loaded = await loadStaffPage()
      res = loaded.res
      resultItems = loaded.items
      total = loaded.total
      if (!searchKey && pagination.value.current === 1) {
        mergeCachedStaff(props.fetchType, status, resultItems, total)
      }
    }

    if (res.code === 200) {
      hasLoadedList.value = true // 标记已加载过员工列表
      
      // 保留首次加载的清空逻辑
      if (pagination.value.current === 1) {
        let existingSelected = []
        if (props.multiple && Array.isArray(props.modelValue)) {
          existingSelected = props.modelValue
            .map(v => staffOptions.value.find(item => sameStaffId(item.id, v)))
            .filter(Boolean)
          for (const v of props.modelValue) {
            if (existingSelected.some(item => sameStaffId(item.id, v)))
              continue
            const p = props.presetStaff?.find(x => sameStaffId(x.id, v))
            if (p && (p.nickName || p.name)) {
              existingSelected.push({
                id: p.id,
                nickName: p.nickName || p.name || '',
                mobile: p.mobile || '',
              })
            }
          }
        }
        else if (selectedStaffDisplay.value && sameStaffId(props.modelValue, selectedStaffDisplay.value.id)) {
          existingSelected = [selectedStaffDisplay.value]
        }
        const newItems = resultItems
        // 合并已选中的员工和新加载的员工，去重
        const existingIds = new Set(existingSelected.map(item => String(item.id)))
        const uniqueNewItems = newItems.filter(item => !existingIds.has(String(item.id)))
        staffOptions.value = [...existingSelected, ...uniqueNewItems]
        mergeMobileFromApiIntoOptions(newItems)
      } else {
        // 合并数据时去重，避免重复的 key
        const newItems = resultItems
        const existingIds = new Set(staffOptions.value.map(item => String(item.id)))
        const uniqueNewItems = newItems.filter(item => !existingIds.has(String(item.id)))
        staffOptions.value = [...staffOptions.value, ...uniqueNewItems]
        mergeMobileFromApiIntoOptions(newItems)
      }
      pagination.value.total = total
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
    const selectedStaffs = filteredValue.map(v => staffOptions.value.find(item => sameStaffId(item.id, v))).filter(Boolean)
    emit('change', filteredValue, selectedStaffs)
  } else {
    // 单选模式：过滤掉特殊值
    if (value === -1 || value === -2 || value === -3 || value === -4) {
      return
    }
    emit('update:modelValue', value)
    emit('change', value, staffOptions.value.find(item => sameStaffId(item.id, value)))
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
      // 多选：必须保留当前选中项在 options 里，否则标签会退回显示原始 value（ID）
      let preserved = []
      if (props.multiple && Array.isArray(props.modelValue) && props.modelValue.length) {
        const mv = props.modelValue.filter(Boolean)
        const seen = new Set()
        for (const v of mv) {
          let row = staffOptions.value.find(item => sameStaffId(item.id, v))
          if (!row && props.presetStaff?.length) {
            const p = props.presetStaff.find(x => sameStaffId(x.id, v))
            if (p && (p.nickName || p.name)) {
              row = {
                id: p.id,
                nickName: p.nickName || p.name || '',
                mobile: p.mobile || '',
              }
            }
          }
          if (row) {
            const k = String(row.id)
            if (!seen.has(k)) {
              seen.add(k)
              preserved.push(row)
            }
          }
        }
      }
      else if (selectedStaffDisplay.value && sameStaffId(props.modelValue, selectedStaffDisplay.value.id)) {
        const sd = selectedStaffDisplay.value
        // 勿把「加载中」占位写回 options，否则与 value 同 id 永远匹配到占位，真实姓名永远出不来
        if (String(sd.nickName || "") === "加载中...")
          preserved = []
        else
          preserved = [sd]
      }

      pagination.value.current = 1
      initialLoading.value = false
      scrollLoading.value = false
      finished.value = false
      hasLoadedList.value = false
      staffOptions.value = preserved
    }, 300)
  }
}

// 根据员工ID获取员工信息（用于初始化显示）
async function getStaffById(staffId) {
  if (!staffId) return null
  const status = effectiveStatus()

  const cachedStaff = findCachedStaff(props.fetchType, status, staffId)
  if (cachedStaff) {
    return cachedStaff
  }

  try {
    const cached = await getCachedInitialStaffList(
      props.fetchType,
      status,
      async () => {
        if (props.fetchType === 'approval') {
          const response = await getStaffSummariesApi({
            queryModel: {},
            pageRequestModel: {
              needTotal: true,
              pageSize: 100,
              pageIndex: 1,
              skipCount: 0,
            },
          })
          return {
            items: normalizeStaffItems(response),
            total: response.result?.total || 0,
          }
        }

        const response = await getUserListApi({
          pageRequestModel: {
            needTotal: true,
            pageSize: 20,
            pageIndex: 1,
            skipCount: 0,
          },
          queryModel: {
            status,
          },
        })
        return {
          items: normalizeStaffItems(response),
          total: response.total || 0,
        }
      },
    )
    const sharedMatched = cached.items.find(item => `${item.id}` === `${staffId}`)
    if (sharedMatched) {
      return sharedMatched
    }
  } catch (error) {
    console.error('从共享缓存获取员工失败:', error)
  }
  
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
          status,
        },
      })
    }

    if (res.code === 200) {
      const sourceList = normalizeStaffItems(res)
      const staff = sourceList.find(item => sameStaffId(item.id, staffId))
      if (staff) {
        mergeCachedStaff(props.fetchType, status, [staff])
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

function applyPresetStaffToOptions() {
  if (!props.multiple || !Array.isArray(props.modelValue) || !props.presetStaff?.length)
    return
  const added = []
  for (const v of props.modelValue) {
    if (!v || staffOptions.value.find(item => sameStaffId(item.id, v)))
      continue
    const p = props.presetStaff.find(x => sameStaffId(x.id, v))
    if (p && (p.nickName || p.name)) {
      added.push({
        id: p.id,
        nickName: p.nickName || p.name || '',
        mobile: p.mobile || '',
      })
    }
  }
  if (added.length)
    staffOptions.value = [...added, ...staffOptions.value]
}

// 监听 modelValue / 预设，初始化显示
watch(() => [props.modelValue, props.presetStaff], async ([newValue]) => {
  if (props.multiple && Array.isArray(newValue)) {
    applyPresetStaffToOptions()

    const targetIds = newValue.filter(Boolean)
    if (targetIds.length === 0) {
      return
    }

    const missingIds = targetIds.filter(value => !staffOptions.value.find(item => sameStaffId(item.id, value)))
    if (missingIds.length === 0) {
      return
    }

    // 多选模式下避免按已选人数逐个请求，先拉一页员工列表再本地匹配
    if (!hasLoadedList.value && !initialLoading.value && !scrollLoading.value) {
      initialLoading.value = true
      finished.value = false
      pagination.value.current = 1
      await getStaffList()
    }

    const stillMissingIds = targetIds.filter(value => !staffOptions.value.find(item => sameStaffId(item.id, value)))
    for (const value of stillMissingIds) {
      const staffInfo = await getStaffById(value)
      if (staffInfo && !staffOptions.value.find(item => sameStaffId(item.id, staffInfo.id))) {
        staffOptions.value = [staffInfo, ...staffOptions.value]
      }
    }
  } else if (newValue && !props.multiple) {
    function rowFromPresetOrOptions() {
      let row = staffOptions.value.find(item => sameStaffId(item.id, newValue))
      if (row && String(row.nickName || "") !== "加载中...")
        return row
      if (row && String(row.nickName || "") === "加载中...")
        row = null
      const preset = props.presetStaff?.find(x => sameStaffId(x.id, newValue))
      if (preset && (preset.nickName || preset.name)) {
        const built = {
          id: preset.id,
          nickName: preset.nickName || preset.name || "",
          mobile: preset.mobile || "",
        }
        staffOptions.value = [
          built,
          ...staffOptions.value.filter(item => !sameStaffId(item.id, built.id)),
        ]
        return built
      }
      return row
    }

    const existing = rowFromPresetOrOptions()
    if (existing) {
      selectedStaffDisplay.value = existing
    }
    else {
      selectedStaffDisplay.value = {
        id: newValue,
        nickName: "加载中...",
        mobile: "",
      }
      try {
        const staffInfo = await getStaffById(newValue)
        if (staffInfo) {
          selectedStaffDisplay.value = staffInfo
          staffOptions.value = [
            staffInfo,
            ...staffOptions.value.filter(item => !sameStaffId(item.id, staffInfo.id)),
          ]
        }
        else {
          selectedStaffDisplay.value = {
            id: newValue,
            nickName: String(newValue),
            mobile: "",
          }
        }
      }
      catch {
        selectedStaffDisplay.value = {
          id: newValue,
          nickName: String(newValue),
          mobile: "",
        }
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
      return staffOptions.value.filter(item =>
        props.modelValue.some(v => sameStaffId(item.id, v)),
      )
    }
    return staffOptions.value.find(item => sameStaffId(item.id, props.modelValue)) || selectedStaffDisplay.value
  },
})
</script>

<style lang="less" scoped>
// 组件样式可以根据需要添加
</style> 
