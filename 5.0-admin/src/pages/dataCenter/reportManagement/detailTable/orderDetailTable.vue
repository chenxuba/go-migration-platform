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
// 默认维度指标
const defaultDimension = [
  { label: '机构名称', value: '1' },
  { label: '订单编号', value: '2' },
  { label: '订单属性', value: '3', hasSubDimension: true, subDimensionKey: 'orderAttribute' },
  { label: '学员姓名', value: '4' },
  { label: '学员手机号', value: '5' },
  { label: '学员基本信息', value: '6', hasSubDimension: true, subDimensionKey: 'studentInfo' },
  { label: '学员自定义属性', value: '7', hasSubDimension: true, subDimensionKey: 'studentCustom' },
  { label: '学员关联人员信息', value: '8' },
  { label: '子订单ID', value: '9' },
  { label: '报读类型', value: '10' },
  { label: '购买商品名', value: '11' },
  { label: '授课方式', value: '12' },
  { label: '课程模板', value: '13' },
  { label: '商品类型', value: '14' },
  { label: '课程类别', value: '15' },
  { label: '课程自定义属性', value: '16', hasSubDimension: true, subDimensionKey: 'courseCustom' },
  { label: '科目', value: '17' },
  { label: '报名选择班级', value: '18' },
  { label: '报名分班状态', value: '19' },
  { label: '报名选择班级当前班主任', value: '20' },
  { label: '购买报价单名称', value: '21' },
  { label: '购买报价单明细', value: '22' },
  { label: '购买份数', value: '23' },
  { label: '购买数量', value: '24', tooltip: '不包含按金额课程模式商品的订单明细' },
  { label: '赠送数量', value: '25', tooltip: '不包含按金额课程模式商品的订单明细' },
  { label: '单课优惠名称', value: '26' },
  { label: '单课优惠明细', value: '27' },
  { label: '单课优惠金额', value: '28' },
  { label: '分摊整单优惠金额', value: '29' },
  { label: '储值账户赠送金额变动分摊', value: '30' },
  { label: '储值账户正价金额变动分摊', value: '31' },
  { label: '优惠券分摊金额', value: '32' },
  { label: '商品应收金额', value: '33' },
  { label: '分摊后实付金额', value: '34' },
]

// 订单属性指标
const orderAttributeDimension = [
  { label: '订单类型', value: '3_1' },
  { label: '订单来源', value: '3_2' },
  { label: '订单标签', value: '3_3' },
  { label: '订单状态', value: '3_4' },
  { label: '订单销售员', value: '3_5' },
  { label: '订单销售员在职状态', value: '3_6' },
  { label: '订单销售员所属部门', value: '3_7' },
  { label: '订单经办时间', value: '3_8' },
  { label: '订单完成时间', value: '3_9' },
  { label: '订单对内备注', value: '3_10' },
  { label: '订单对外备注', value: '3_11' },
  { label: '订单经办人', value: '3_12' },
  { label: '整单优惠名称', value: '3_13' },
  { label: '整单优惠金额', value: '3_14' },
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
const studentCustomDimension = [
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
  orderAttribute: orderAttributeDimension,
  studentInfo: studentInfoDimension,
  studentCustom: studentCustomDimension,
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
  orderAttribute: {
    indeterminate: true,
    checkAll: false,
    checkedList: ['3_1', '3_2'],
    visible: false,
  },
  studentInfo: {
    indeterminate: true,
    checkAll: false,
    checkedList: [],
    visible: false,
  },
  studentCustom: {
    indeterminate: true,
    checkAll: false,
    checkedList: [],
    visible: false,
  },
  courseCustom: {
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
          <div v-if="subDimensionStates.orderAttribute.visible" class="flex flex-col gap-15px">
            <div
              class="text-#0066ff flex justify-between py-5px px-15px items-center gap-1 border-#0066ff border-1px rounded-15px border-solid w-120px"
            >
              <div>订单属性</div>
              <CloseOutlined @click="closeSubDimension('orderAttribute')" />
            </div>
            <div class="flex items-start ">
              <a-checkbox
                v-model:checked="subDimensionStates.orderAttribute.checkAll" class="w-90px"
                :indeterminate="subDimensionStates.orderAttribute.indeterminate"
                @change="(e) => onSubDimensionCheckAllChange('orderAttribute', e)"
              >
                全选
              </a-checkbox>
              <a-checkbox-group
                v-model:value="subDimensionStates.orderAttribute.checkedList" class="gap-10px"
                :options="dimensionMap.orderAttribute"
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
          <!--  -->
          <div v-if="subDimensionStates.studentInfo.visible" class="flex flex-col gap-15px">
            <div
              class="text-#0066ff flex justify-between py-5px px-15px items-center gap-1 border-#0066ff border-1px rounded-15px border-solid w-150px"
            >
              <div>学员基本信息</div>
              <CloseOutlined @click="closeSubDimension('studentInfo')" />
            </div>
            <div class="flex items-start ">
              <a-checkbox
                v-model:checked="subDimensionStates.studentInfo.checkAll" class="w-90px"
                :indeterminate="subDimensionStates.studentInfo.indeterminate"
                @change="(e) => onSubDimensionCheckAllChange('studentInfo', e)"
              >
                全选
              </a-checkbox>
              <a-checkbox-group
                v-model:value="subDimensionStates.studentInfo.checkedList" class="gap-10px"
                :options="dimensionMap.studentInfo"
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
          <div v-if="subDimensionStates.studentCustom.visible" class="flex flex-col gap-15px">
            <div
              class="text-#0066ff flex justify-between py-5px px-15px items-center gap-1 border-#0066ff border-1px rounded-15px border-solid w-180px"
            >
              <div>学员关联人员信息</div>
              <CloseOutlined @click="closeSubDimension('studentCustom')" />
            </div>
            <div class="flex items-start ">
              <a-checkbox
                v-model:checked="subDimensionStates.studentCustom.checkAll" class="w-90px"
                :indeterminate="subDimensionStates.studentCustom.indeterminate"
                @change="(e) => onSubDimensionCheckAllChange('studentCustom', e)"
              >
                全选
              </a-checkbox>
              <a-checkbox-group
                v-model:value="subDimensionStates.studentCustom.checkedList" class="gap-10px"
                :options="dimensionMap.studentCustom"
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
          <div v-if="subDimensionStates.courseCustom.visible" class="flex flex-col gap-15px">
            <div
              class="text-#0066ff flex justify-between py-5px px-15px items-center gap-1 border-#0066ff border-1px rounded-15px border-solid w-150px"
            >
              <div>课程自定义属性</div>
              <CloseOutlined @click="closeSubDimension('courseCustom')" />
            </div>
            <div class="flex items-start ">
              <a-checkbox
                v-model:checked="subDimensionStates.courseCustom.checkAll" class="w-90px"
                :indeterminate="subDimensionStates.courseCustom.indeterminate"
                @change="(e) => onSubDimensionCheckAllChange('courseCustom', e)"
              >
                全选
              </a-checkbox>
              <a-checkbox-group
                v-model:value="subDimensionStates.courseCustom.checkedList" class="gap-10px"
                :options="dimensionMap.courseCustom"
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
