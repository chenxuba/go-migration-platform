<script setup>
import { computed, onBeforeUnmount, ref, toRaw, watch } from 'vue'
import { DownOutlined } from '@ant-design/icons-vue'
import { debounce } from 'lodash-es'
import dayjs from 'dayjs'
import { Cascader, Empty } from 'ant-design-vue'
const openFlag = ref(false)

const props = defineProps({
  options: {
    type: Array,
    default: () => ([
      { id: 1, value: '高' },
      { id: 2, value: '中' },
      { id: 3, value: '低' },
      { id: 4, value: '未知' },
    ]),
  },
  label: {
    type: String,
    default: '意向度',
  },
  type: {
    type: String,
    default: 'radio',
  },
  checkedValues: {
    type: [Array, Number, String, Object],
  },
  placeholder: {
    type: String,
    default: '请输入关键字',
  },
  category: {
    type: String,
    default: 'stu',
  },
  showSearch: {
    type: Boolean,
    default: false,
  },
  id: {
    type: [Number, String],
    default: null,
  },
  finished: {
    type: Boolean,
    default: false,
  },
  cascaderExpandTrigger: {
    type: String,
    default: 'click',
  },
})

const emit = defineEmits(['update:checkedValues', 'change', 'radioChange', 'datePickerChange', 'inputChange', 'dropdownVisibleChange', 'onDropdownVisibleChange', 'onSearch', 'loadMore'])

const simpleImage = Empty.PRESENTED_IMAGE_SIMPLE

// 将 specifiedDate 改为响应式变量，每个组件实例都有自己的
const specifiedDate = ref(null)

// 禁用条件：日期在 specifiedDate 之前 或 今天之后
function disabledDate(current) {
  const today = new Date()
  today.setHours(23, 59, 59, 999) // 设置到当天的最后一刻

  // 克隆当前日期并清除时间部分
  const currentDate = new Date(current)
  currentDate.setHours(0, 0, 0, 0) // 设置到当天的开始

  // 如果有指定日期，则禁用指定日期之前的日期；同时始终禁用今天之后的日期
  // 确保 specifiedDate 也是 00:00:00
  const specDate = specifiedDate.value
    ? new Date(specifiedDate.value.setHours(0, 0, 0, 0))
    : null

  return (
    (specDate && currentDate < specDate)
    || currentDate > today
  )
}
// 禁用条件：日期在 specifiedDate 之前
function disabledDateBefore(current) {
  const today = new Date()
  today.setHours(23, 59, 59, 999) // 设置到当天的最后一刻
  // 克隆当前日期并清除时间部分
  const currentDate = new Date(current)
  currentDate.setHours(0, 0, 0, 0) // 设置到当天的开始
  // 禁用条件：日期在 specifiedDate 之前
  const specDate = specifiedDate.value
    ? new Date(specifiedDate.value.setHours(0, 0, 0, 0))
    : null
  return (specDate && currentDate < specDate)
}

// 含未来时间
const rangePresets = ref([
  {
    label: '本周',
    // 周一到周日（含未来日期）
    value: [dayjs().startOf('week'), dayjs().endOf('week')],
  },
  {
    label: '本月',
    // 本月1号到本月最后一天（含未来日期）
    value: [dayjs().startOf('month'), dayjs().endOf('month')],
  },
  {
    label: '上周',
    value: [
      dayjs().subtract(1, 'week').startOf('week'),
      dayjs().subtract(1, 'week').endOf('week'),
    ],
  },
  {
    label: '上月',
    value: [
      dayjs().subtract(1, 'month').startOf('month'),
      dayjs().subtract(1, 'month').endOf('month'),
    ],
  },
  {
    label: '截止今日',
    // 固定起始日期到当前日期（不含未来）
    value: [dayjs('2020-01-01'), dayjs()],
  },
])
// 不含未来时间
const rangePresetsNot = ref([
  {
    label: '本周',
    // 周一到当前时间（不含未来）
    value: [dayjs().startOf('week'), dayjs()],
  },
  {
    label: '本月',
    // 本月1号到当前时间（不含未来）
    value: [dayjs().startOf('month'), dayjs()],
  },
  {
    label: '上周',
    // 上周时间范围（自动排除未来，因上周已过去）
    value: [
      dayjs().subtract(1, 'week').startOf('week'),
      dayjs().subtract(1, 'week').endOf('week'),
    ],
  },
  {
    label: '上月',
    // 上月时间范围（自动排除未来，因上月已过去）
    value: [
      dayjs().subtract(1, 'month').startOf('month'),
      dayjs().subtract(1, 'month').endOf('month'),
    ],
  },
  {
    label: '截止今日',
    // 固定起始日期到当前时间（不含未来）
    value: [dayjs('2020-01-01'), dayjs()],
  },
])
function calendarChangeFun(e) {
  specifiedDate.value = new Date(e[0])
}
const searchPeo = ref('')
const searchChannelCategory = ref('')
const searchInput = ref('')
const selectDates = ref([])
const pickerKey = ref(0)

function filter(inputValue, path) {
  return path.some(option => option.name.toLowerCase().includes(inputValue.toLowerCase()))
}
const spinning = ref(false)
// 监听输入变化
watch(searchPeo, (newVal) => {
  spinning.value = true
  debouncedSearch(newVal)
})
// 监听输入变化
watch(searchChannelCategory, (newVal) => {
  debouncedSearchCategory(newVal)
})
function handleDropdownVisible(visible) {
  if (visible && props.category == 'stu') {
    // console.log("下拉框展开");
    emit('onDropdownVisibleChange', searchPeo.value)
  }
  if (visible && (props.category == 'course' || props.category == 'teacher')) {
    emit('onDropdownVisibleChange', searchPeo.value)
  }
  if (visible && props.category == 'courseAttribute') {
    searchPeo.value = ''
    emit('onDropdownVisibleChange', searchPeo.value, props.id)
  }
}
// 暴露清空方法给父组件
function resetSearch() {
  searchPeo.value = ''
  searchInputVals.value = ''
  searchNumberInputVals.value = ''
  if (props.type === 'tree') {
    treeSelectedValue.value = null
  }
}
function resetSpinning() {
  spinning.value = false
}
function openSpinning() {
  spinning.value = true
}
defineExpose({
  resetSearch,
  resetSpinning,
  openSpinning,
  closeDropdown: () => { visible.value = false }, // 新增暴露方法
})
// 实际搜索逻辑
function doSearch() {
  console.log('执行搜索1:', searchPeo.value)
  // 这里替换为真实的搜索逻辑
  emit('onSearch', searchPeo.value)
}
// 实际搜索逻辑
function doSearchCategory() {
  console.log('执行搜索2:', searchChannelCategory.value)
  // 如果是课程类别，触发远程搜索
  if (props.category === 'course') {
    emit('onSearch', searchChannelCategory.value)
  }
  // 其他情况使用本地过滤（在 computed 中处理）
}
// 创建防抖函数（500ms延迟）
const debouncedSearch = debounce(doSearch, 300)
const debouncedSearchCategory = debounce(doSearchCategory, 500)
// 过滤后的选项列表（使用 computed 实现）
const filteredOptions = computed(() => {
  // 如果是课程类别，直接返回 options（已经是远程搜索的结果）
  if (props.category === 'course') {
    return props.options
  }

  const searchText = searchChannelCategory.value.trim().toLowerCase()

  if (!searchText)
    return props.options

  return props.options.filter(option =>
    option.value.toLowerCase().includes(searchText),
  )
})
// 组件卸载前清理
onBeforeUnmount(() => {
  debouncedSearch.cancel()
})
const visible = ref(false)
const treeSelectKey = ref(0) // 用于强制重新渲染

// 为树形选择器使用独立的 ref
const treeSelectedValue = ref(null)

const checkedValues = computed({
  get() {
    return props.checkedValues
  },
  set(value) {
    emit('update:checkedValues', value)
  },
})
const searchInputVals = ref('')
const searchNumberInputVals = ref('')
function handleChange(type) {
  if (type === 'inputType') {
    checkedValues.value = searchInputVals.value
  }
  else if (type === 'numberInputType') {
    checkedValues.value = searchNumberInputVals.value
  }
  emit('change', checkedValues.value)
}

function handleCheckboxGroupChange(values) {
  emit('change', values)
}
function handleDropdownVisibleChange(visible) {
  // 如果为true emit自定义事件 让父组件请求接口获取渠道树
  if (visible) {
    emit('dropdownVisibleChange')
  }
}
function handleRadioChange(id) {
  emit('radioChange', id)
  checkedValues.value = id
  visible.value = false
}

function handleTreeChange(value) {
  // 更新本地值
  treeSelectedValue.value = value
  // 发出 update 事件
  emit('update:checkedValues', value)
  emit('change', value)
  // visible.value = false
  openFlag.value = true
}

function handleTreeDropdownChange(open) {
  // 当下拉框关闭时，重置 openFlag
  if (!open) {
    openFlag.value = false
  }
}
function handleRangePicker(dates) {
  checkedValues.value = dates
  selectDates.value = []
  // 转为普通数组
  const rawDates = toRaw(dates)
  // 或浅拷贝
  // const rawDates = [...dates];
  emit('datePickerChange', rawDates)
  visible.value = false
  setTimeout(() => {
    pickerKey.value++ // 强制重新渲染
  }, 400)
}
function resetPicker() {
  specifiedDate.value = null
  selectDates.value = []
  pickerKey.value++ // 强制重新渲染
}
function handleReset() {
  if (props.type === 'inputType') {
    searchInputVals.value = ''
  }
  else if (props.type === 'numberInputType') {
    searchNumberInputVals.value = ''
  }
}

// 监听下拉框的显示/隐藏状态
watch(visible, (newVal) => {
  if (newVal) {
    // 当下拉框打开时，触发事件（用于加载数据）
    if (props.category === 'course' || props.category === 'teacher' || props.category === 'stu') {
      emit('onDropdownVisibleChange', searchPeo.value)
    }
  } else {
    // 当下拉框关闭时，重置所有状态
    searchPeo.value = ''
    searchChannelCategory.value = ''
    specifiedDate.value = null
    selectDates.value = []
    pickerKey.value++ // 加上这个不会出现bug 但会没有隐藏动画
  }
})

// 监听 checkedValues 的变化，同步到树形选择器
watch(() => props.checkedValues, (newVal, oldVal) => {
  if (newVal !== oldVal && props.type === 'tree') {
    console.log('props.checkedValues changed from', oldVal, 'to', newVal)
    treeSelectedValue.value = newVal
    treeSelectKey.value++
  }
}, { immediate: true })

const isYearOrMonthPanel = ref(false)

// 添加面板变化处理函数
function handlePanelChange(value, mode) {
  // mode 是一个数组，包含左右两个面板的模式
  isYearOrMonthPanel.value = mode.includes('year') || mode.includes('month')

  // 如果需要在面板切换时隐藏预设选项，可以通过添加类名来控制
  const presets = document.querySelector('.ant-picker-presets')
  if (presets) {
    if (isYearOrMonthPanel.value) {
      presets.style.display = 'none'
    }
    else {
      presets.style.display = ''
    }
  }
}

const scrollContainer = ref(null)

function handleScroll(e) {
  const { scrollTop, scrollHeight, clientHeight } = e.target
  // 当滚动到距离底部20px时触发加载
  if (scrollHeight - scrollTop - clientHeight < 5) {
    emit('loadMore')
  }
}
</script>

<template>
  <a-dropdown v-if="type == 'checkbox'" v-model:open="visible" :trigger="['click']" placement="bottomLeft"
    :arrow="true">
    <template #overlay>
      <a-menu>
        <a-menu-item v-if="showSearch" class="top-item">
          <!-- 搜索栏 -->
          <a-input v-model:value="searchChannelCategory" class="mt-1 mb-2 w-35" autofocus :placeholder="placeholder" 
            @input="debouncedSearchCategory" />
        </a-menu-item>
        <a-menu-item v-if="filteredOptions.length > 0" class="check-item ">
          <div ref="scrollContainer" class="list scrollbar" @scroll="handleScroll">
            <a-spin :spinning="spinning">
              <a-checkbox-group v-model:value="checkedValues" class="vertical-checkbox-group " @change="handleCheckboxGroupChange">
                <a-checkbox v-for="item in filteredOptions" :key="item.id" :value="item.id">
                  {{ item.value }}
                </a-checkbox>
              </a-checkbox-group>
              <div v-if="finished && filteredOptions.length > 0" class="no-more-data">
                没有更多了
              </div>
            </a-spin>
          </div>
        </a-menu-item>
        <a-menu-item v-if="filteredOptions.length == 0">
          <a-empty :image="simpleImage" />
        </a-menu-item>
      </a-menu>
    </template>

    <a-button class="h-28px flex filter-btn mr-2 mb-2" :class="{ 'active-blue': visible }">
      {{ label }}
      <div v-if="checkedValues.length > 0" class="num">
        {{ checkedValues.length }}
      </div>
      <DownOutlined v-else :style="{ fontSize: '10px' }" />
    </a-button>
  </a-dropdown>
  <a-dropdown v-if="type == 'radio'" v-model:open="visible" :trigger="['click']" placement="bottomLeft" :arrow="true"
    @open-change="handleDropdownVisible">
    <template #overlay>
      <a-menu>
        <a-menu-item class="top-item">
          <!-- 搜索栏 -->
          <div v-if="category == 'course' || category == 'teacher'" class="flex justify-center w-120px px-8px">
            <a-input v-model:value="searchPeo" class="mt-1 mb-2 w-100%" allow-clear :placeholder="placeholder" />
          </div>
          <div v-if="category == 'stu'" class="flex justify-center px-8px">
            <a-input v-model:value="searchPeo" class="mt-1 mb-2 w-100% " :placeholder="placeholder" allow-clear />
          </div>
        </a-menu-item>
        <div v-if="category == 'course'" class="max-h-80 overflow-auto scrollbar">
          <a-spin :spinning="spinning">
            <a-menu-item v-for="item in options" :key="item.id"
              :class="checkedValues == item.id ? 'menu-item active' : 'menu-item'" :value="item.id"
              @click="handleRadioChange(item.id)">
              <div class="text-sm text-#666  leading-7">
                {{ item.value ?? item.name ?? item.stuName ?? item.nickName }}
              </div>
              <div class="text-xs text-#888">
                {{ item.mobile ?? item.phone ?? '' }}
              </div>
            </a-menu-item>
            <div v-if="options.length == 0" class="flex justify-center">
              <a-empty :image-style="{ width: '80px' }" :image="simpleImage" />
            </div>
            <div v-if="options.length > 0 && finished" class="no-more-data">
              没有更多了
            </div>
          </a-spin>
        </div>
        <div v-if="category == 'teacher'" ref="scrollContainer" class="max-h-80 overflow-auto scrollbar"
          @scroll="handleScroll">
          <a-spin :spinning="spinning">
            <a-menu-item v-for="item in options" :key="item.id"
              :class="checkedValues == item.id ? 'menu-item active' : 'menu-item'" :value="item.id"
              @click="handleRadioChange(item.id)">
              <div class="text-sm text-#666  leading-7">
                {{ item.value ?? item.name ?? item.stuName ?? item.nickName }}
              </div>
              <div class="text-xs text-#888">
                {{ item.mobile ?? item.phone ?? '' }}
              </div>
            </a-menu-item>
            <div v-if="options.length == 0" class="flex justify-center">
              <a-empty :image-style="{ width: '80px' }" :image="simpleImage" />
            </div>
            <div v-if="options.length > 0 && finished" class="no-more-data">
              没有更多了
            </div>
          </a-spin>
        </div>
        <div v-if="category == 'noSearchRadio'" class="max-h-80 overflow-auto scrollbar">
          <a-menu-item v-for="item in options" :key="item.id"
            :class="checkedValues == item.id ? 'menu-item active' : 'menu-item'" :value="item.id"
            @click="handleRadioChange(item.id)">
            <div class="text-sm text-#666  leading-7">
              {{ item.value ?? item.name }}
            </div>
          </a-menu-item>
          <div v-if="options.length == 0" class="flex justify-center">
            <a-empty :image-style="{ width: '80px' }" :image="simpleImage" />
          </div>
        </div>
        <div v-if="category == 'stu'" ref="scrollContainer" class="max-h-70 overflow-auto scrollbar"
          @scroll="handleScroll">
          <a-spin :spinning="spinning">
            <a-menu-item v-for="item in options" :key="item.id"
              :class="checkedValues == item.id ? 'menu-item active' : 'menu-item'" :value="item.id"
              @click="handleRadioChange(item.id)">
              <div class="flex flex-center mb-2">
                <div>
                  <img class="w-10 rounded-10" :src="item.avatarUrl" alt="">
                </div>
                <div class="ml-2 mr-3">
                  <div class="text-sm text-#666  leading-7">
                    {{ item.value ?? item.name ?? item.stuName }}
                  </div>
                  <div class="text-xs text-#888">
                    {{ item.mobile ?? '' }}
                  </div>
                </div>
                <div>
                  <a-tag v-if="item.studentStatus == 1" :bordered="false" color="processing">
                    在读学员
                  </a-tag>
                  <a-tag v-if="item.studentStatus == 0" :bordered="false" color="orange">
                    意向学员
                  </a-tag>
                </div>
              </div>
            </a-menu-item>

            <div v-if="options.length == 0" class="flex justify-center">
              <a-empty :image-style="{ width: '80px' }" :image="simpleImage" />
            </div>

            <div v-if="options.length > 0 && finished" class="no-more-data">
              没有更多了
            </div>
          </a-spin>
        </div>
      </a-menu>
    </template>
    <a-button class="h-28px flex filter-btn mr-2 mb-2" :class="{ 'active-blue': visible }">
      {{ label }}
      <div v-if="checkedValues !== undefined && checkedValues !== null && checkedValues !== ''" class="num">
        1
      </div>
      <DownOutlined v-else :style="{ fontSize: '10px' }" />
    </a-button>
  </a-dropdown>
  <a-dropdown v-if="type == 'dateTime'" v-model:open="visible" :trigger="['click']" placement="bottomLeft"
    :arrow="true">
    <a-button style="position: relative;" class="h-28px flex filter-btn mr-2 mb-2" :class="{ 'active-blue': visible }">
      {{ label }}
      <div v-if="checkedValues.length > 0" class="num">
        1
      </div>
      <DownOutlined v-else :style="{ fontSize: '10px' }" />
      <a-range-picker :key="pickerKey" v-model:value="selectDates" value-format="YYYY-MM-DD"
        :disabled-date="disabledDate" popup-class-name="picker-wrapper dateTimeQuick" :open="visible"
        :presets="rangePresetsNot" @calendar-change="calendarChangeFun" @change="handleRangePicker"
        @panel-change="handlePanelChange">
        <template #renderExtraFooter>
          <div v-if="!isYearOrMonthPanel" class="pl-3.5">
            <a-tag color="pink">
              本周
            </a-tag>
            <a-tag color="red">
              上周
            </a-tag>
            <a-tag color="orange">
              本月
            </a-tag>
            <a-tag color="green">
              上月
            </a-tag>
            <a-tag color="cyan">
              截至昨日
            </a-tag>
            <a-tag color="#e6f4ff" class="cursor-pointer ml-8 reset-btn" @click="resetPicker">
              重置
            </a-tag>
          </div>
        </template>
      </a-range-picker>
    </a-button>
  </a-dropdown>
  <a-dropdown v-if="type == 'dateTimeQuick'" v-model:open="visible" :trigger="['click']" placement="bottomLeft"
    :arrow="true">
    <a-button style="position: relative;" class="h-28px flex filter-btn mr-2 mb-2" :class="{ 'active-blue': visible }">
      {{ label }}
      <div v-if="checkedValues.length > 0" class="num">
        1
      </div>
      <DownOutlined v-else :style="{ fontSize: '10px' }" />
      <a-range-picker :key="pickerKey" v-model:value="selectDates" value-format="YYYY-MM-DD"
        :disabled-date="disabledDateBefore" popup-class-name="picker-wrapper dateTimeQuick" :open="visible"
        :presets="rangePresets" @calendar-change="calendarChangeFun" @change="handleRangePicker"
        @panel-change="handlePanelChange">
        <template #renderExtraFooter>
          <div v-if="!isYearOrMonthPanel" class="pl-3.5">
            <a-tag color="pink">
              本周
            </a-tag>
            <a-tag color="red">
              上周
            </a-tag>
            <a-tag color="orange">
              本月
            </a-tag>
            <a-tag color="green">
              上月
            </a-tag>
            <a-tag color="cyan">
              截至昨日
            </a-tag>
            <a-tag color="#e6f4ff" class="cursor-pointer ml-8 reset-btn" @click="resetPicker">
              重置
            </a-tag>
          </div>
        </template>
      </a-range-picker>
    </a-button>
  </a-dropdown>
  <a-dropdown v-if="type == 'tree'" v-model:open="visible" :trigger="['click']" placement="bottomLeft" :arrow="true"
    @open-change="handleTreeDropdownChange">
    <template #overlay>
      <a-menu>
        <a-tree-select popupClassName="tree-select" @click="openFlag = true" :key="treeSelectKey"
          v-model:value="treeSelectedValue" style="width: 200px"
          :dropdown-style="{ maxHeight: '400px', overflow: 'auto' }" placeholder="请选择部门" tree-default-expand-all
          :tree-data="options" :field-names="{ children: 'children', label: 'name', value: 'id' }"
          @change="handleTreeChange" :open="openFlag">
          <template #title="{ value: id, name }">
            <b v-if="id === 1" style="color: #08c"></b>
            <template v-else>
              {{ name }}
            </template>
          </template>
        </a-tree-select>
      </a-menu>
    </template>
    <a-button class="h-28px flex filter-btn mr-2 mb-2" :class="{ 'active-blue': visible }">
      {{ label }}
      <div v-if="checkedValues !== undefined && checkedValues !== null && checkedValues !== ''" class="num">
        1
      </div>
      <DownOutlined v-else :style="{ fontSize: '10px' }" />
    </a-button>
  </a-dropdown>
  <a-dropdown v-if="type == 'cascader'" v-model:open="visible" :trigger="['click']" placement="bottomLeft"
    :arrow="true">
    <template #overlay>
      <a-menu>
        <a-cascader v-model:value="checkedValues" :field-names="{ children: 'channelList', label: 'name', value: 'id' }"
          :expand-trigger="cascaderExpandTrigger"
          :show-search="{ filter }" :show-checked-strategy="Cascader.SHOW_CHILD" multiple max-tag-count="responsive"
          :options="options" :placeholder="placeholder || '搜索渠道'" @change="handleChange"
          @dropdown-visible-change="handleDropdownVisibleChange" />
      </a-menu>
    </template>
    <a-button class="h-28px flex filter-btn mr-2 mb-2" :class="{ 'active-blue': visible }">
      {{ label }}
      <div v-if="checkedValues.length > 0" class="num">
        {{ checkedValues.length }}
      </div>
      <DownOutlined v-else :style="{ fontSize: '10px' }" />
    </a-button>
  </a-dropdown>
  <a-dropdown v-if="type == 'custom'" v-model:open="visible" :trigger="['click']" placement="bottomLeft" :arrow="true">
    <template #overlay>
      <a-menu>
        <div class="custom-filter">
          <slot name="custom" />
        </div>
      </a-menu>
    </template>
    <a-button class="h-28px flex filter-btn mr-2 mb-2" :class="{ 'active-blue': visible }">
      {{ label }}
      <div v-if="checkedValues.length > 0" class="num">
        1
      </div>
      <DownOutlined v-else :style="{ fontSize: '10px' }" />
    </a-button>
  </a-dropdown>
  <a-dropdown v-if="type == 'inputType'" v-model:open="visible" :trigger="['click']" placement="bottomLeft"
    :arrow="true">
    <template #overlay>
      <a-menu>
        <div class="custom-filter p-2">
          <a-input v-model:value="searchInputVals" :placeholder="placeholder" />
          <!-- 分割线 -->
          <div class="w-full h-1px bg-#e5e5e5 my-2" />
          <!-- 重置  确定 -->
          <div class="flex justify-between">
            <a-button type="primary" :disabled="!searchInputVals" size="small" @click="handleReset">
              <span class="text-12px w-40px">重置</span>
            </a-button>
            <a-button type="primary" size="small" @click="handleChange('inputType')">
              <span class="text-12px w-40px">确定</span>
            </a-button>
          </div>
        </div>
      </a-menu>
    </template>
    <a-button class="h-28px flex filter-btn mr-2 mb-2" :class="{ 'active-blue': visible }">
      {{ label }}
      <div v-if="checkedValues !== undefined && checkedValues !== null && checkedValues !== ''" class="num">
        1
      </div>
      <DownOutlined v-else :style="{ fontSize: '10px' }" />
    </a-button>
  </a-dropdown>
  <a-dropdown v-if="type == 'numberInputType'" v-model:open="visible" :trigger="['click']" placement="bottomLeft"
    :arrow="true">
    <template #overlay>
      <a-menu>
        <div class="custom-filter p-2">
          <a-input-number v-model:value="searchNumberInputVals" class="w-180px" :placeholder="placeholder" />
          <!-- 分割线 -->
          <div class="w-full h-1px bg-#e5e5e5 my-2" />
          <!-- 重置  确定 -->
          <div class="flex justify-between">
            <a-button type="primary" :disabled="!searchNumberInputVals" size="small" @click="handleReset">
              <span class="text-12px w-40px">重置</span>
            </a-button>
            <a-button type="primary" size="small" @click="handleChange('numberInputType')">
              <span class="text-12px w-40px">确定</span>
            </a-button>
          </div>
        </div>
      </a-menu>
    </template>
    <a-button class="h-28px flex filter-btn mr-2 mb-2" :class="{ 'active-blue': visible }">
      {{ label }}
      <div v-if="checkedValues !== undefined && checkedValues !== null && checkedValues !== ''" class="num">
        1
      </div>
      <DownOutlined v-else :style="{ fontSize: '10px' }" />
    </a-button>
  </a-dropdown>
  <a-dropdown v-if="type == 'dateSelectType'" v-model:open="visible" :trigger="['click']" placement="bottomLeft"
    :arrow="true">
    <a-button style="position: relative;" class="h-28px flex filter-btn mr-2 mb-2" :class="{ 'active-blue': visible }">
      {{ label }}
      <div v-if="Array.isArray(checkedValues) ? checkedValues.length > 0 : checkedValues !== undefined && checkedValues !== null && checkedValues !== ''" class="num">
        1
      </div>
      <DownOutlined v-else :style="{ fontSize: '10px' }" />
      <a-range-picker :key="pickerKey" v-model:value="selectDates" value-format="YYYY-MM-DD"
        :disabled-date="disabledDateBefore" popup-class-name="picker-wrapper dateTimeQuick" :open="visible"
        :presets="rangePresets" @calendar-change="calendarChangeFun" @change="handleRangePicker">
        <template #renderExtraFooter>
          <div class="pl-3.5">
            <a-tag color="pink">
              本周
            </a-tag>
            <a-tag color="red">
              上周
            </a-tag>
            <a-tag color="orange">
              本月
            </a-tag>
            <a-tag color="green">
              上月
            </a-tag>
            <a-tag color="cyan">
              截至昨日
            </a-tag>
            <a-tag color="#e6f4ff" class="cursor-pointer ml-8 reset-btn" @click="resetPicker">
              重置
            </a-tag>
          </div>
        </template>
      </a-range-picker>
    </a-button>
  </a-dropdown>
  <a-dropdown v-if="type == 'radioType'" v-model:open="visible" :trigger="['click']" placement="bottomLeft"
    :arrow="true" @open-change="handleDropdownVisible">
    <template #overlay>
      <a-menu>
        <a-menu-item class="top-item">
          <!-- 搜索栏 -->
          <a-input v-model:value="searchPeo" class="mt-1 mb-2 w-130px" autofocus :placeholder="placeholder" />
        </a-menu-item>
        <div class="max-h-80 overflow-auto scrollbar">
          <a-spin :spinning="spinning">
            <a-menu-item v-for="item in options" :key="item.id"
              :class="checkedValues == item.id ? 'menu-item active' : 'menu-item'" :value="item.id"
              @click="handleRadioChange(item.id)">
              <div class="text-sm text-#666  leading-7">
                {{ item.value ?? item.name }}
              </div>
              <div class="text-xs text-#888">
                {{ item.phone ?? '' }}
              </div>
            </a-menu-item>

            <div v-if="options.length == 0" class="flex justify-center">
              <a-empty :image-style="{ width: '80px' }" :image="simpleImage" />
            </div>
          </a-spin>
        </div>
      </a-menu>
    </template>

    <a-button class="h-28px flex filter-btn mr-2 mb-2" :class="{ 'active-blue': visible }">
      {{ label }}
      <div v-if="checkedValues !== undefined && checkedValues !== null && checkedValues !== ''" class="num">
        1
      </div>
      <DownOutlined v-else :style="{ fontSize: '10px' }" />
    </a-button>
  </a-dropdown>
</template>

<!-- 保留原有样式或根据需要添加 scoped 样式 -->
<style lang="less" scoped>
@import url(~@/assets/styles/common.css);

.active-blue {
  border: 1px solid #1677ff !important;
  color: #1677ff !important;
}

.num {
  background: #f33;
  color: #fff;
  font-size: 12px;
  height: 16px;
  line-height: 16px;
  margin-left: 6px;
  display: flex;
  justify-content: center;
  width: 16px;
  border-radius: 100px;
  font-weight: 600;
}

/* 或者使用 flex 布局 */
.vertical-checkbox-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

:deep(.ant-dropdown-menu-item.check-item) {
  .list {
    // padding-right: 24px;
    max-height: 300px;
    overflow-y: auto;
    overflow-x: hidden;

    &::-webkit-scrollbar {
      width: 0px;
    }
  }

  &:hover {
    background-color: transparent !important;
  }
}

:deep(.ant-dropdown-menu-item.menu-item) {
  &:hover {
    background-color: rgba(230, 240, 255, .5) !important;
  }
}

:deep(.active) {
  background-color: rgba(230, 240, 255, .5) !important;

  div {
    color: var(--pro-ant-color-primary) !important;
  }
}

:deep(.top-item) {
  padding: 0 !important;
  text-align: center;
}

:deep(.ant-picker-range) {
  pointer-events: none;
  opacity: 0;
  position: absolute;
  left: -10px !important;
}

:deep(.ant-select-tree-node-selected) {
  width: 140px !important;
}

.reset-btn {
  position: relative;
  top: -2px;
  left: -4px;
  height: 22.85px;
  border-color: #7abdff;
  color: #06f;
  border-radius: 4px;
}

.no-more-data {
  text-align: center;
  color: #999;
  font-size: 12px;
  padding: 8px 0;
  border-top: 1px dashed #eee;
  margin-top: 14px;
}
</style>

<style lang="less">
.picker-wrapper {

  .ant-picker-range-arrow,
  .ant-picker-range-arrow {
    display: none !important;
    // left: 0 !important;
  }
}

.ant-dropdown-show-arrow {
  z-index: 9999 !important;
}

.dateTimeQuick {
  // .ant-picker-panel-layout {}

  .ant-picker-presets {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    width: 100%;
    height: 30px;
    max-width: 55% !important;
    background: #fff;

    ul {
      overflow: hidden !important;
      display: flex;
      align-items: center;
      padding: 0 !important;
      margin-top: -10px !important;
      padding: 0 0 0 12px !important;
      border: none !important;

      li {
        margin: 0 !important;
        margin-right: 12px !important;
        background: #e6f4ff;
        color: #06f;
        cursor: pointer;
        font-size: 12px;
        border: 1px solid #f0f0f0 !important;
        border-color: #7abdff !important;
        border-radius: 4px !important;
        padding: 1px 8px !important;

        &:hover {
          background: #e6f4ff !important;
        }
      }
    }
  }
}
.tree-select .ant-select-tree-title {
  display: block;
  white-space: nowrap !important;
  max-width: 150px !important;
  /* 展示省略号 */
  overflow: hidden !important;
  text-overflow: ellipsis !important;

}
</style>
