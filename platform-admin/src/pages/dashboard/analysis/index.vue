<script setup>
import { computed, nextTick, ref } from 'vue'
import { CloseOutlined, DeleteOutlined } from '@ant-design/icons-vue'
import checkboxFilter from '~@/components/common/checkbox-filter.vue'

const searchKey = ref(undefined)
const childRef0 = ref(null)
const childRef1 = ref(null)
const childRef2 = ref(null)
const lastUpdated = reactive({})
const conditionOrder = ref([]) // 存储条件类型的添加顺序
// 意向度选项
const customOptions = ref([
  { id: 1, value: '高' },
  { id: 2, value: '中' },
  { id: 3, value: '低' },
  { id: 4, value: '未知' },
])
const selectedValues = ref([])

// 跟进状态选项
const followStatusOptions = ref([
  { id: 1, value: '待跟进' },
  { id: 2, value: '跟进中' },
  { id: 3, value: '未接听' },
  { id: 4, value: '已邀约' },
])
const followStatusVals = ref([])

// 性别选项
const sexOptions = ref([
  { id: 1, value: '男' },
  { id: 2, value: '女' },
  { id: 3, value: '未知' },
])
const sexVals = ref([])

// 创建人选项
const createUserOptions = ref([
  { id: 1, value: '陈瑞生', phone: '17601241636' },
  { id: 2, value: '张晨', phone: '17662072520' },
  { id: 3, value: '陈旭', phone: '15864646629' },
])
const createUserVals = ref(null)

// 创建时间选项
const createTimeVals = ref([])

// 课程列表选项
const courseListOptions = ref([
  { id: 1, value: '初级感统课' },
  { id: 2, value: '初级言语课' },
  { id: 3, value: '交互认知课' },
  { id: 4, value: '高级认知课' },
  { id: 5, value: '高级游戏课' },
])
const selectCourseValues = ref(null)
// 推荐人选项
const stuListOptions = ref([
  { id: 1, name: '张学良', phone: '17601241636' },
  { id: 2, name: '高保庆', phone: '18882327343' },
  { id: 3, name: '谢霆锋', phone: '16723338886' },
  { id: 4, name: '张飞', phone: '17662372329' },
  { id: 5, name: '武松', phone: '19187235172' },
  { id: 6, name: '杜十娘', phone: '15422127865' },
])
const selectStuVals = ref(null)

// 快捷筛选选项（单选）
const quickFilters = ref([
  { id: 1, name: '今日待跟进', count: 1, selected: false },
  { id: 2, name: '本周新增', count: 0, selected: false },
  { id: 3, name: '逾期未回访', count: 0, selected: false },
])

// 处理快捷筛选单选
function selectQuickFilter(selectedFilter) {
  quickFilters.value.forEach((filter) => {
    filter.selected = filter.id === selectedFilter.id ? !filter.selected : false
  })
  console.log('当前快捷筛选:', quickFilters.value.find(q => q.selected)?.name)
}

// 处理常规筛选变化
function handleIntentionChange() {
  nextTick(() => {
    console.log('意向度:', toRaw(selectedValues.value))
  })
}

function handleFollowChange() {
  nextTick(() => {
    console.log('跟进状态:', toRaw(followStatusVals.value))
  })
}

function handleSexChange() {
  nextTick(() => {
    console.log('性别:', toRaw(sexVals.value))
  })
}
function handlecreateUserChange(e) {
  nextTick(() => {
    console.log('创建人:', e)
  })
}
function handleCreateTimeChange(e) {
  nextTick(() => {
    console.log('创建时间:', e)
  })
}
function handleCourseChange(e) {
  nextTick(() => {
    console.log('意向课程:', e)
  })
}
function handleReferenceChange(e) {
  nextTick(() => {
    console.log('推荐人:', e)
  })
}

// 已选条件计算
const selectedConditions = computed(() => {
  const conditions = [
    {
      type: 'quick',
      label: '快捷筛选',
      values: quickFilters.value.filter(q => q.selected).map(q => ({ id: q.id, value: q.name })),
    },
    {
      type: 'intention',
      label: '意向度',
      values: customOptions.value.filter(opt => selectedValues.value.includes(opt.id)),
    },
    {
      type: 'followStatus',
      label: '跟进状态',
      values: followStatusOptions.value.filter(opt => followStatusVals.value.includes(opt.id)),
    },
    {
      type: 'sex',
      label: '性别',
      values: sexOptions.value.filter(opt => sexVals.value.includes(opt.id)),
    },
    {
      type: 'createUser',
      label: '创建人',
      values: createUserOptions.value.filter(opt => opt.id === createUserVals.value),
    },
    {
      type: 'createTime',
      label: '创建时间',
      values: createTimeVals.value.length === 2
        ? [{
            id: 'dateRange',
            value: `${createTimeVals.value[0]} 至 ${createTimeVals.value[1]}`,
          }]
        : [],
    },
    {
      type: 'intentionCourse',
      label: '意向课程',
      values: courseListOptions.value.filter(opt => opt.id === selectCourseValues.value),
    },
    {
      type: 'reference',
      label: '推荐人',
      values: stuListOptions.value.filter(opt => opt.id === selectStuVals.value),
    },
  ]
  return conditions
    .filter(item => item.values.length > 0)
    .sort((a, b) => (lastUpdated[a.type] || 0) - (lastUpdated[b.type] || 0))
})
// 监听各条件变化，更新最后操作时间
watch(selectedValues, () => lastUpdated.intention = Date.now())
watch(followStatusVals, () => lastUpdated.followStatus = Date.now())
watch(sexVals, () => lastUpdated.sex = Date.now())
watch(createUserVals, () => lastUpdated.createUser = Date.now())
watch(createTimeVals, () => lastUpdated.createTime = Date.now())
watch(selectCourseValues, () => lastUpdated.intentionCourse = Date.now())
watch(() => quickFilters.value.map(q => q.selected), () => lastUpdated.quick = Date.now(), { deep: true })
watch(selectStuVals, () => lastUpdated.reference = Date.now())
// 观察筛选条件变化，维护顺序队列
watch(selectedConditions, (newConditions) => {
  const newTypes = newConditions.map(c => c.type)

  // 保留仍然存在的类型
  conditionOrder.value = conditionOrder.value.filter(t => newTypes.includes(t))

  // 添加新增的类型到队列末尾
  newTypes.forEach((t) => {
    if (!conditionOrder.value.includes(t)) {
      conditionOrder.value.push(t)
    }
  })
}, { deep: true })

// 最终使用的排序条件
const orderedConditions = computed(() => {
  return [...selectedConditions.value].sort((a, b) =>
    conditionOrder.value.indexOf(a.type) - conditionOrder.value.indexOf(b.type),
  )
})
// 清空所有筛选
function clearAll() {
  // 重置多选类
  [selectedValues, followStatusVals, sexVals, createTimeVals].forEach(
    ref => ref.value = [],
  )
  // 重置单选类
  quickFilters.value.forEach(q => q.selected = false)
  createUserVals.value = null
  selectCourseValues.value = null
  selectStuVals.value = null
  if (childRef0.value) {
    childRef0.value.resetSearch()
  }
  if (childRef1.value) {
    childRef1.value.resetSearch()
  }
  if (childRef2.value) {
    childRef2.value.resetSearch()
  }
}

// 移除单个条件
function removeCondition(type, id) {
  switch (type) {
    case 'intention':
      selectedValues.value = []
      break
    case 'followStatus':
      followStatusVals.value = []
      break
    case 'sex':
      sexVals.value = []
      break
    case 'quick':
      const filter = quickFilters.value.find(q => q.id === id)
      if (filter)
        filter.selected = false
      break
    case 'createUser': // 新增创建人移除逻辑
      createUserVals.value = null
      break
    case 'createTime': // 新增创建时间移除逻辑
      createTimeVals.value = []
      break
    case 'intentionCourse': // 新增意向课程移除逻辑
      selectCourseValues.value = null
      break
    case 'reference': // 新增推荐人移除逻辑
      selectStuVals.value = null
      break
  }
}

function handleChange(value) {
  console.log(`selected ${value}`)
}
function filterOption(input, option) {
  // 正确访问选项的 label（对应 item.name）和 phone
  const name = option.label || '' // 对应 :label="item.name"
  const phone = option.data?.phone || '' // 通过 data 传递额外字段

  return (
    name.toLowerCase().includes(input.toLowerCase())
    || phone.includes(input)
  )
}
</script>

<template>
  <div class="home flex">
    <div class="flex-1 mr-0">
      <!-- 快捷筛选区域 -->
      <div class="filter-section mb-2">
        <span class="section-title mt-0.5">快捷筛选：</span>
        <div class="quick-filters">
          <a-button
            v-for="filter in quickFilters" :key="filter.id" :type="filter.selected ? 'primary' : 'default'"
            class="filter-btn" @click="selectQuickFilter(filter)"
          >
            {{ filter.name }}（{{ filter.count }}）
          </a-button>
        </div>
      </div>

      <!-- 常规筛选条件 -->
      <div class="filter-section">
        <span class="section-title mt-0.5">筛选条件：</span>
        <div class="standard-filters">
          <checkbox-filter
            v-model:checked-values="selectedValues" :options="customOptions" label="意向度"
            type="checkbox" @change="handleIntentionChange"
          />
          <checkbox-filter
            v-model:checked-values="followStatusVals" :options="followStatusOptions" label="跟进状态"
            type="checkbox" @change="handleFollowChange"
          />
          <checkbox-filter
            v-model:checked-values="sexVals" :options="sexOptions" label="性别" type="checkbox"
            @change="handleSexChange"
          />
          <checkbox-filter
            ref="childRef0" v-model:checked-values="createUserVals" category="teacher" placeholder="请输入创建人"
            :options="createUserOptions" label="创建人" type="radio" @radio-change="handlecreateUserChange"
          />
          <checkbox-filter
            v-model:checked-values="createTimeVals" label="创建时间"
            type="dateTime" @date-picker-change="handleCreateTimeChange"
          />
          <checkbox-filter
            ref="childRef1" v-model:checked-values="selectCourseValues" category="course"
            placeholder="请输入意向课程" :options="courseListOptions" label="意向课程"
            type="radio" @radio-change="handleCourseChange"
          />
          <checkbox-filter
            ref="childRef2" v-model:checked-values="selectStuVals" category="stu" placeholder="请输入推荐人"
            :options="stuListOptions" label="推荐人" type="radio" @radio-change="handleReferenceChange"
          />
        </div>
      </div>

      <!-- 已选条件展示 -->
      <div v-if="orderedConditions.length > 0" class="selected-conditions">
        <span class="section-title">已选条件：</span>
        <div class="condition-tags">
          <a-tag color="red" class="clear-all mb-2" @click="clearAll">
            清空所有
            <DeleteOutlined class="text-3 ml-0.5" />
          </a-tag>

          <a-tag v-for="condition in orderedConditions" :key="condition.type" color="blue" class="condition-tag mb-2">
            <div class="tag-content">
              <span class="condition-label">{{ condition.label }}：</span>
              <div class="condition-values">
                <template v-if="condition.type === 'quick'">
                  <span class="value-item">
                    {{ condition.values[0].value }}
                    <CloseOutlined
                      class="close-icon"
                      @click.stop="removeCondition(condition.type, condition.values[0].id)"
                    />
                  </span>
                </template>
                <template v-else-if="condition.type === 'createTime'">
                  <span class="value-item">
                    {{ condition.values[0].value }}
                    <CloseOutlined class="close-icon" @click.stop="removeCondition(condition.type, 0)" />
                  </span>
                </template>
                <template v-else>
                  <span v-for="(value, index) in condition.values" :key="value.id" class="value-item">
                    {{ value.value ?? value.name }}
                    <CloseOutlined
                      v-if="index === condition.values.length - 1" class="close-icon"
                      @click.stop="removeCondition(condition.type, value.id)"
                    />
                    <span v-else class="separator">、</span>
                  </span>
                </template>
              </div>
            </div>
          </a-tag>
        </div>
      </div>
    </div>
    <div class="w-100">
      <div class="selectBox flex ">
        <div class="label">
          学员/电话
        </div>
        <div>
          <a-select
            v-model:value="searchKey" :filter-option="filterOption" show-search placeholder="搜索姓名/手机号"
            style="width: 240px" option-label-prop="label" @change="handleChange"
          >
            <a-select-option
              v-for="(item) in stuListOptions" :key="item.id" :value="item.id" :data="item"
              :label="item.name"
            >
              <div class="flex flex-center mb-1">
                <div>
                  <img
                    class="w-10 rounded-10"
                    src="https://pcsys.admin.ybc365.com/a369a751-2be5-4929-974d-9ae4439f54c4.png"
                    alt=""
                  >
                </div>
                <div class="ml-2 mr-3">
                  <div class="text-sm text-#666  leading-7">
                    {{ item.name }}
                  </div>
                  <div class="text-xs text-#888">
                    {{ item.phone }}
                  </div>
                </div>
                <div>
                  <a-tag :bordered="false" color="processing">
                    在读学员
                  </a-tag>
                </div>
              </div>
            </a-select-option>
          </a-select>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.selectBox {
  justify-content: flex-end;
  align-items: center;
}

.selectBox .label {
  border: 1px solid #d9d9d9;
  height: 32px;
  padding: 0 10px;
  line-height: 32px;
  text-align: left;
  width: 100px;
  border-radius: 6px;
  border-radius: 8px 0 0 8px !important;
  color: #222;
  font-size: 14px;
  min-width: 104px;
  padding-left: 8px;
  padding-right: 16px;
}

:deep(.selectBox .ant-select-selector) {
  border-left: none !important;
  border-radius: 0 6px 6px 0 !important;
}

.home {
  padding: 24px;
  background: #ffffff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  align-items: flex-start;
}

.debug-panel {
  padding: 16px;
  margin-bottom: 24px;
  background: #f8f9fa;
  border-radius: 6px;
}

.filter-section {
  display: flex;
  align-items: flex-start;
}

.section-title {
  white-space: nowrap;
}

.quick-filters {
  display: flex;
  gap: 8px;
}

.standard-filters {
  display: flex;
  flex-wrap: wrap;
}

.selected-conditions {
  display: flex;
  align-items: flex-start;
}

.condition-tags {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}

.condition-tag {
  display: flex;
  align-items: center;
  border-radius: 4px;
}

.tag-content {
  display: flex;
  align-items: center;
}

.condition-values {
  display: flex;
  align-items: center;
}

.value-item {
  display: inline-flex;
  align-items: center;
}

.close-icon {
  margin-left: 6px;
  font-size: 12px;
  cursor: pointer;
  color: rgba(92, 92, 92, 0.45);
  transition: color 0.3s;
}

.close-icon:hover {
  color: rgba(0, 0, 0, 0.75);
}

.clear-all {
  cursor: pointer;
  display: flex;
  align-items: center;
}

.filter-btn {
  height: 28px;
  padding: 0 12px;
}
</style>
