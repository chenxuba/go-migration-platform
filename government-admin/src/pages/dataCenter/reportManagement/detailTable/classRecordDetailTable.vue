<script setup>
import { CloseOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
// 时间范围
const dateRange = ref([])
// 快捷日期范围选项
const dateRangeOptions = [
  { label: '本周', value: 'week' },
  { label: '上周', value: 'lastWeek' },
  { label: '本月', value: 'month' },
  { label: '上月', value: 'lastMonth' },
]
// 快捷日期选项
const activeDateType = ref('')

// 监听日期选项变化
watch(activeDateType, (newVal, oldVal) => {
  switch (newVal) {
    case 'week':
      dateRange.value = [dayjs().startOf('week'), dayjs().endOf('week')]
      break
    case 'lastWeek':
      dateRange.value = [dayjs().subtract(1, 'week').startOf('week'), dayjs().subtract(1, 'week').endOf('week')]
      break
    case 'month':
      dateRange.value = [dayjs().startOf('month'), dayjs().endOf('month')]
      break
    case 'lastMonth':
      dateRange.value = [dayjs().subtract(1, 'month').startOf('month'), dayjs().subtract(1, 'month').endOf('month')]
      break
    default:
      break
  }
})

// 维度选项
const dimensionOptions = [
  { label: '默认维度', value: 'default' },
]

// 维度
const dimension = ref('default')
// 默认一级维度指标
const defaultDimension = [
  { label: '所属机构', value: '1' },
  { label: '上课记录ID', value: '2' },
  { label: '上课日期', value: '3' },
  { label: '上课时段', value: '4' },
  { label: '上课学员姓名', value: '5' },
  { label: '学员基本信息', value: '6', hasSubDimension: true, subDimensionKey: 'studentInfo' },
  { label: '学员自定义属性', value: '7' },
  { label: '学员关联人员信息', value: '8', hasSubDimension: true, subDimensionKey: 'studentRelation' },
  { label: '上课老师', value: '9' },
  { label: '上课老师是否兼职', value: '10' },
  { label: '上课助教', value: '11' },
  { label: '上课老师所属部门', value: '12' },
  { label: '上课班级', value: '13' },
  { label: '上课课程', value: '14' },
  { label: '课程类别', value: '15' },
  { label: '课程自定义属性', value: '16', hasSubDimension: true, subDimensionKey: 'courseCustom' },
  { label: '科目', value: '17' },
  { label: '课程模板', value: '18' },
  { label: '授课方式', value: '19' },
  { label: '日程类型', value: '20' },
  { label: '学员身份', value: '21' },
  { label: '扣费课程账户名', value: '22' },
  { label: '课消方式', value: '23' },
  { label: '应扣金额数量', value: '24' },
  { label: '实扣金额数量', value: '25' },
  { label: '应扣课时数量', value: '26' },
  { label: '实扣课时数量', value: '27' },
  { label: '扣除数量对应课消金额（元）', value: '28' },
  { label: '拖欠课时数量', value: '29' },
  { label: '拖欠金额数量', value: '30' },
  { label: '点名更新时间', value: '31' },
  { label: '点名更新操作人', value: '32' },
  { label: '点名创建时间', value: '33' },
  { label: '点名创建人', value: '34' },
  { label: '对内备注', value: '35' },
  { label: '对外备注', value: '36' },
]

// 学员基本信息指标
const studentInfoDimension = [
  { label: '手机号归属人', value: '6_1' },
  { label: '微信号', value: '6_2' },
  { label: '渠道', value: '6_3' },
  { label: '性别', value: '6_4' },
  { label: '学员年龄', value: '6_5' },
  { label: '生日', value: '6_6' },
  { label: '创建人', value: '6_7' },
  { label: '转介绍推荐人', value: '6_8' },
  { label: '学员备注', value: '6_9' },
  { label: '学员状态', value: '6_10' },
  { label: '创建时间', value: '6_11' },
  { label: '首次报读时间', value: '6_12' },
  { label: '最近跟进时间', value: '6_13' },
  { label: '关联储值账户余额', value: '6_14' },
  { label: '关联储值账户赠送余额', value: '6_15' },
  { label: '剩余积分数量', value: '6_16' },
]

// 学员关联人员信息指标
const studentRelationDimension = [
  { label: '电话销售', value: '7_1' },
  { label: '副销售', value: '7_2' },
  { label: '前台', value: '7_3' },
  { label: '采单员', value: '7_4' },
  { label: '顾问', value: '7_5' },
  { label: '学管师', value: '7_6' },
  { label: '班主任', value: '7_7' },
  { label: '销售员', value: '7_8' },
]

// 课程自定义属性指标
const courseCustomDimension = [
  { label: '适合年龄', value: '16_1' },
  { label: '课程难度', value: '16_2' },
  { label: '教材版本', value: '16_3' },
]

// 维度映射
const dimensionMap = {
  default: defaultDimension,
  studentInfo: studentInfoDimension,
  studentRelation: studentRelationDimension,
  courseCustom: courseCustomDimension,
}

// 一级维度映射
const plainOptions = computed(() => {
  return dimensionMap[dimension.value]
})

// 一级多选状态
const state = reactive({
  indeterminate: true,
  checkAll: false,
  checkedList: ['1', '2'], // 修改为对应的value值
})

// 二级指标状态集合
const subDimensionStates = reactive({
  studentInfo: {
    title: '学员基本信息',
    indeterminate: true,
    checkAll: false,
    checkedList: [],
    visible: false,
  },
  studentRelation: {
    title: '学员关联人员信息',
    indeterminate: true,
    checkAll: false,
    checkedList: [],
    visible: false,
  },
  courseCustom: {
    title: '课程自定义属性',
    indeterminate: true,
    checkAll: false,
    checkedList: [],
    visible: false,
  },
})

// 监听一级指标选择变化，控制二级指标显示
watch(() => state.checkedList, (newVal) => {
  // 重置所有二级指标的可见性
  Object.keys(subDimensionStates).forEach((key) => {
    subDimensionStates[key].visible = false
  })

  // 检查选中的一级指标，显示对应的二级指标
  defaultDimension.forEach((item) => {
    if (item.hasSubDimension && newVal.includes(item.value)) {
      subDimensionStates[item.subDimensionKey].visible = true
    }
  })
  // 更新一级指标的indeterminate状态
  state.indeterminate = newVal.length > 0 && newVal.length < plainOptions.value.length
  state.checkAll = newVal.length === plainOptions.value.length
}, { deep: true })

// 监听各个二级指标选择变化，更新indeterminate状态
Object.keys(subDimensionStates).forEach((key) => {
  watch(() => subDimensionStates[key].checkedList, (newVal) => {
    // 更新二级指标的indeterminate状态
    const totalOptions = dimensionMap[key].length
    subDimensionStates[key].indeterminate = newVal.length > 0 && newVal.length < totalOptions
    subDimensionStates[key].checkAll = newVal.length === totalOptions
  }, { deep: true })
})

// 一级全选
function onCheckAllChange(e) {
  Object.assign(state, {
    checkedList: e.target.checked ? plainOptions.value.map(option => option.value) : [],
    indeterminate: false,
  })
};

// 二级指标全选
function onSubDimensionCheckAllChange(dimensionKey, e) {
  Object.assign(subDimensionStates[dimensionKey], {
    checkedList: e.target.checked ? dimensionMap[dimensionKey].map(option => option.value) : [],
    indeterminate: false,
  })
};

// 关闭二级指标面板
function closeSubDimension(dimensionKey) {
  subDimensionStates[dimensionKey].visible = false
  // 从一级指标中移除对应的选项
  const parentItem = defaultDimension.find(item => item.subDimensionKey === dimensionKey)
  if (parentItem) {
    const index = state.checkedList.indexOf(parentItem.value)
    if (index > -1) {
      state.checkedList.splice(index, 1)
    }
  }
}

// 获取选中的指标 包括二级
const selectedDimensions = computed(() => {
  const dimensions = []
  // 处理一级指标
  state.checkedList.forEach((value) => {
    const item = plainOptions.value.find(option => option.value === value)
    if (item) {
      dimensions.push(item)
    }
  })

  // 处理二级指标
  Object.keys(subDimensionStates).forEach((key) => {
    if (subDimensionStates[key].visible) {
      dimensions.push(...dimensionMap[key].filter(option => subDimensionStates[key].checkedList.includes(option.value)))
    }
  })

  return dimensions
})
</script>

<template>
  <div>
    <div class="card-white flex flex-col gap-20px">
      <!-- 时间 -->
      <div class="flex items-center gap-5px">
        <div class="label text-gray flex items-center gap-1">
          <span class="text-red">*</span>
          <span>订单经办时间：</span>
        </div>
        <div class="value">
          <a-range-picker v-model:value="dateRange" class="mr-10px" value-format="YYYY-MM-DD" />
          <a-radio-group v-model:value="activeDateType" button-style="solid">
            <a-radio-button v-for="(item, index) in dateRangeOptions" :key="index" :value="item.value">
              {{ item.label }}
            </a-radio-button>
          </a-radio-group>
        </div>
      </div>
      <!-- 指标 -->
      <div class="flex items-start gap-5px">
        <div class="label text-gray flex items-center gap-1 min-w-55px">
          <span class="text-red">*</span>
          <span>指标：</span>
        </div>
        <div class="value flex items-start">
          <div class="min-w-65px">
            <a-checkbox
              v-model:checked="state.checkAll" :indeterminate="state.indeterminate"
              @change="onCheckAllChange"
            >
              全选
            </a-checkbox>
          </div>

          <a-checkbox-group v-model:value="state.checkedList" class="gap-10px" :options="plainOptions">
            <template #label="{ label, tooltip }">
              <div class="flex items-center">
                <span class="mr-5px">{{ label }}</span>
                <a-popover v-if="tooltip" color="#fff" title="字段说明">
                  <template #content>
                    <div v-html="tooltip" />
                  </template>
                  <ExclamationCircleOutlined class="text-#888 font-size-14px" />
                </a-popover>
              </div>
            </template>
          </a-checkbox-group>
        </div>
      </div>
      <!-- 二级指标区域 -->
      <div>
        <!-- 订单属性二级指标 -->
        <div
          v-if="Object.values(subDimensionStates).some(state => state.visible)"
          class="flex flex-col items-start gap-25px rounded-8px border-1px border-solid border-gray-300 p-15px bg-#f8f9fb"
        >
          <template v-for="(dimensionState, key) in subDimensionStates" :key="key">
            <div v-if="dimensionState.visible" class="flex flex-col gap-15px">
              <div
                class="max-w-170px text-#0066ff flex justify-between py-5px px-15px items-center gap-1 border-#0066ff border-1px rounded-15px border-solid"
              >
                <div>{{ dimensionState.title }}</div>
                <CloseOutlined @click="closeSubDimension(key)" />
              </div>
              <div class="flex items-start ">
                <a-checkbox
                  v-model:checked="dimensionState.checkAll" class="w-90px" :indeterminate="dimensionState.indeterminate"
                  @change="(e) => onSubDimensionCheckAllChange(key, e)"
                >
                  全选
                </a-checkbox>
                <a-checkbox-group
                  v-model:value="dimensionState.checkedList" class="gap-10px"
                  :options="dimensionMap[key]"
                >
                  <template #label="{ label, tooltip }">
                    <div class="flex items-center">
                      <span class="mr-5px">{{ label }}</span>
                      <a-popover v-if="tooltip" color="#fff" title="字段说明">
                        <template #content>
                          <div v-html="tooltip" />
                        </template>
                        <ExclamationCircleOutlined class="text-#888 font-size-14px" />
                      </a-popover>
                    </div>
                  </template>
                </a-checkbox-group>
              </div>
            </div>
          </template>
        </div>
      </div>

      <!-- 按钮 -->
      <div class="flex justify-end gap-20px">
        <a-button type="default">
          历史报表
        </a-button>
        <a-button type="primary">
          生成报表
        </a-button>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.card-white {
  background: #fff;
  // margin-top: 8px;
  padding: 12px;
  border-bottom-left-radius: 12px;
  border-bottom-right-radius: 12px;
}
</style>
