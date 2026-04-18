<script setup>
import Icon, { CloseOutlined, FormOutlined, PlusOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import chooseCoursesModal from './chooseCoursesModal.vue'
import confirmInformationCoursesModal from './confirmInformationCoursesModal.vue'
import orderSettingsConfirmationModal from './orderSettingsConfirmationModal.vue'
import orderCompletedDrawer from './orderCompletedDrawer.vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open'])

// 处理双向绑定
const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
// 切换tab
const activeKey = ref('me')
// 步骤状态
const stepsStatus = ref(1)
// 学生列表
const stuListOptions = ref([
  { id: 1, name: '张学良', phone: '17601241636' },
  { id: 2, name: '高保庆', phone: '18882327343' },
  { id: 3, name: '谢霆锋', phone: '16723338886' },
  { id: 4, name: '张飞', phone: '17662372329' },
  { id: 5, name: '武松', phone: '19187235172' },
  { id: 6, name: '杜十娘', phone: '15422127865' },
])
const fromRef = ref()
// 转出课程form
const formState = reactive({
  transferOutAmount: undefined, // 转出数量
})
// 全部转出
function handleAllOut() {
  formState.transferOutAmount = 1
}

// 选择课程弹窗
const showChooseCourses = ref(false)
// 判断是否点了选择课程按钮
const flgCourse = ref(false)
const classOptions = [
  {
    label: '阳光学院',
    value: '1',
  },
]
const courseOptions = [
  { value: 1, label: '初级言语课' },
  { value: 2, label: '初级认知课' },
  { value: 3, label: '初级感统课' },
]
// 课程内容表单
const courseForm = reactive({
  courseName: undefined, // 课程名称
  className: undefined, // 班级名称
  type: '1', // 收费方式
  intoCourse: '', // 转入课时
  giveCourse: '', // 赠送课时
  validity: '1', // 有效期
  validityDate: undefined, // 有效期日期
  orderLabel: undefined, // 订单标签
  internalNotes: '', // 对内备注
  externalRemarks: '', // 对外备注
  intoDays: '', // 转入天数
  giveDays: '', // 赠送天数
  startDate: '', // 开始日期
  giveMoney: '', // 赠送金额
})

const courseFormRef = ref()

// 转给他人课程内容表单
const transferForm = reactive({
  courseName: '', // 课程名称
  studentName: '', // 学生姓名
  className: '', // 班级名称
  type: '1', // 收费方式
  intoCourse: '', // 转入课时
  validity: '1', // 有效期
  validityDate: '', // 有效期日期
  orderLabel: '', // 订单标签
  internalNotes: '', // 对内备注
  externalRemarks: '', // 对外备注
})

const transferFormRef = ref()

// 二次确认弹窗
const showConfirmInformationCourses = ref(false)
const needAdjust = ref(false)
// 多退少补表单
const adjustForm = reactive({
  newCourseType: undefined, // 课程类型
  className: '选择班级', // 班级名称
  buyCount: '', // 购买数量
  giveDays: '', // 赠送天数
  startDate: '', // 开始日期
  singleCourseDiscount: '1', // 单课优惠
  singleCourseDiscountRate: '', // 单课优惠率
})

const adjustFormRef = ref()

// 是否保存多退少补表单
const saveAdjustForm = ref(false)
// 保存多退少补表单
function handleAdjust() {
  if (!saveAdjustForm.value) {
    adjustFormRef.value.validate().then(() => {
      console.log('adjustFormRef验证成功')
      saveAdjustForm.value = true
    })
  }
}
// 订单设置表单
const orderSettingForm = reactive({

})
// 订单设置弹窗
const showOrderSettingsConfirmation = ref(false)

// 初始化表单状态的函数
function initFormState() {
  return {
    model: '1',
    // 退款方式
    payType: 1,
    // 账单备注
    billRemarks: '',
    fileList: [
      {
        uid: '-1',
        name: 'image.png',
        status: 'done',
        url: 'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png',
      },
      {
        uid: '-2',
        name: 'image.png',
        status: 'done',
        url: 'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png',
      },
      {
        uid: '-3',
        name: 'image.png',
        status: 'done',
        url: 'https://zos.alipayobjects.com/rmsportal/jkjgkEfvpUPVyRjUImniVslZfWPnJuuZ.png',
      },
    ],
  }
}
// 订单完成弹窗
const showOrderCompletedDrawer = ref(false)

const paymentForm = reactive(initFormState())
// 订单标签
const orderLabelOptions = ref([
  { id: '1', name: '内荐' },
  { id: '2', name: '转介绍' },
  { id: '3', name: '双11订单' },
  { id: '4', name: '会员日订单' },
])
// 订单销售员
const salespersonOptions = ref([
  { id: '1', name: '陈瑞生', phone: '17601241636' },
  { id: '2', name: '刘明', phone: '18876552232' },
  { id: '3', name: '张望名', phone: '17601241636' },
  { id: '4', name: '李元芳', phone: '17601241636' },
])
function orderLabelFilterOption(input, option) {
  const keyword = input.toLowerCase()
  const nameMatch = option.name.toLowerCase().includes(keyword)
  return nameMatch
}

function getBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.readAsDataURL(file)
    reader.onload = () => resolve(reader.result)
    reader.onerror = error => reject(error)
  })
}
function handleCancelImg() {
  previewVisible.value = false
  previewTitle.value = ''
}
const checkOptions = reactive([
  {
    id: 1,
    label: '微信',
    img: 'https://pcsys.admin.ybc365.com/e068b5e2-e27e-4228-8437-fae315326ced.png',
  },
  {
    id: 2,
    label: '支付宝',
    img: 'https://pcsys.admin.ybc365.com/3f396285-ac3a-43fe-94a8-adfef395d47d.png',
  },
  {
    id: 3,
    label: '银行转帐',
    img: 'https://pcsys.admin.ybc365.com/b05c2b92-6c09-46a0-8123-42446727d480.png',
  },
  {
    id: 4,
    label: 'POS机',
    img: 'https://pcsys.admin.ybc365.com/2a6137fb-3e35-470f-b368-2fe71e5439f1.png',
  },
  {
    id: 5,
    label: '现金',
    img: 'https://pcsys.admin.ybc365.com/bab59869-17ea-42c9-9354-d88a59ecf18a.png',
  },
  {
    id: 6,
    label: '其他',
    img: 'https://pcsys.admin.ybc365.com/6d6faf00-aaea-4f6c-b3ab-e581a9bcf7f6.png',
  },
  {
    id: 7,
    label: '储值账户',
    img: 'https://pcsys.admin.ybc365.com/bca6bad7-3180-4b01-b766-4b05013347f7.png',
  },

])
// 点击下一步
function handleNext() {
  switch (stepsStatus.value) {
    case 1:
      fromRef.value.validate().then(() => {
        console.log('fromRef验证成功')
        stepsStatus.value = 2
      })
      break
    case 2:
      if (activeKey.value === 'me') {
        courseFormRef.value.validate().then(() => {
          console.log('courseFormRef验证成功', courseForm)
          // stepsStatus.value = 3
          showConfirmInformationCourses.value = true
        })
      }
      else if (activeKey.value === 'he') {
        transferFormRef.value.validate().then(() => {
          console.log('transferFormRef验证成功')
          // stepsStatus.value = 3
          showConfirmInformationCourses.value = true
        })
      }
      break
    case 3:
      showOrderCompletedDrawer.value = true
      drawerClose()
      break
  }
}
// 点击上一步
function handleCancel() {
  if (stepsStatus.value > 1) {
    stepsStatus.value -= 1
  }
  else {
    drawerClose()
  }
}
// 选择课程后回显
function handleChooseCoursesChange(courses) {
  flgCourse.value = true
  courseForm.courseName = courses.id
  console.log('选择的课程：', courses)
}
// 二次确认弹窗返回
function handleConfirmInformationCoursesConfirm(value) {
  console.log('确认转入转出', value)
  if (value) {
    drawerClose()
  }
}
// 二次确认设置弹窗返回
function handleOrderSettingsConfirmationConfirm(value) {
  if (value) {
    console.log('订单设置弹窗返回：', value)
    needAdjust.value = false
    saveAdjustForm.value = false
    stepsStatus.value = 3
  }
}

// 关闭弹窗 回掉
function drawerClose() {
  // 重置各种状态
  stepsStatus.value = 1 // 步骤状态
  activeKey.value = 'me' //  重置切换tab
  openDrawer.value = false // 重置弹窗状态
  flgCourse.value = false // 重置选择课程按钮的状态
  saveAdjustForm.value = false // 重置多退少补表单的状态
  needAdjust.value = false // 重置多退少补表单的状态
  showChooseCourses.value = false // 重置选择课程弹窗状态
  showConfirmInformationCourses.value = false // 重置二次确认弹窗状态
  // 重置表单
  // fromRef.value.resetFields()
  // courseFormRef.value.resetFields()
  // transferFormRef.value.resetFields()
  // adjustFormRef.value.resetFields()
}
// 禁用今天之前的日期
function disabledDate(current) {
  return current && current < dayjs().startOf('day')
}
// 禁用今天之后的日期
function disabledEndDate(current) {
  return current && current > dayjs().endOf('day')
}
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer" :destroy-on-close="true" :keyboard="false" :mask-closable="false" :push="false"
      :body-style="{ padding: '0', background: '#f7f7fd' }" :closable="false" width="1165px" placement="right"
      @close="drawerClose"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            转课
          </div>
          <a-button type="text" class="close-btn" @click="drawerClose">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <!-- 自定义内容 -->
      <div class="mx-30px my-20px">
        <!-- tabs -->
        <div class="mb-20px">
          <div class="flex justify-center mb-20px">
            <a-radio-group v-model:value="activeKey" size="large" button-style="solid" @change="() => stepsStatus = 1">
              <a-radio-button value="me" class="px-8">
                转给自己
              </a-radio-button>
              <a-radio-button value="he" class="px-8">
                转给他人
              </a-radio-button>
            </a-radio-group>
          </div>
          <!-- 步骤图标 -->
          <div class="bg-white p-6 rounded-lg flex gap-10px justify-center items-center">
            <div class="flex gap-10px items-center flex-1 max-w-25%">
              <div class="flex  items-center justify-center">
                <Icon :style="{ color: 'hotpink' }">
                  <template #component>
                    <svg
                      width="33px" height="32px" viewBox="0 0 33 32" version="1.1" xmlns="http://www.w3.org/2000/svg"
                      xmlns:xlink="http://www.w3.org/1999/xlink"
                    >
                      <title>编组备份 9</title>
                      <g id="页面-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                        <g id="转课" transform="translate(-304.000000, -270.000000)">
                          <g id="组件/步骤条备份-7" transform="translate(304.000000, 270.000000)">
                            <g id="编组" transform="translate(0.500000, 0.000000)">
                              <rect id="矩形备份" x="0" y="0" width="32" height="32" />
                              <path
                                id="形状结合"
                                d="M21.6604298,15.4568657 C21.9633694,14.9886864 22.5907143,14.8587428 23.0547002,15.1680667 L23.0547002,15.1680667 L26.0547002,17.1680667 C26.5097843,17.4714561 26.6366976,18.0841713 26.3395702,18.5433683 L26.3395702,18.5433683 L20.8395702,27.0433683 C20.7162133,27.2340107 20.5316467,27.376994 20.3162278,27.4488003 L20.3162278,27.4488003 L17.3162278,28.4488003 C16.6686967,28.664644 16,28.1826747 16,27.500117 L16,27.500117 L16,24.500117 C16,24.3073557 16.055712,24.1187024 16.1604298,23.9568657 L16.1604298,23.9568657 Z M24,5 C25.5976809,5 26.9036609,6.24891996 26.9949073,7.82372721 L27,8 L27,13 C27,13.5522847 26.5522847,14 26,14 C25.4477153,14 25,13.5522847 25,13 L25,8 C25,7.48716416 24.6139598,7.06449284 24.1166211,7.00672773 L24,7 L8,7 C7.48716416,7 7.06449284,7.38604019 7.00672773,7.88337887 L7,8 L7,24 C7,24.5128358 7.38604019,24.9355072 7.88337887,24.9932723 L8,25 L13,25 C13.5522847,25 14,25.4477153 14,26 C14,26.5522847 13.5522847,27 13,27 L8,27 C6.40231912,27 5.09633912,25.75108 5.00509269,24.1762728 L5,24 L5,8 C5,6.40231912 6.24891996,5.09633912 7.82372721,5.00509269 L8,5 L24,5 Z M22.788,17.394117 L18,24.794117 L18,26.112117 L19.35,25.662117 L24.124,18.285117 L22.788,17.394117 Z M13,18 C13.5522847,18 14,18.4477153 14,19 C14,19.5522847 13.5522847,20 13,20 L10,20 C9.44771525,20 9,19.5522847 9,19 C9,18.4477153 9.44771525,18 10,18 L13,18 Z M13,14 C13.5522847,14 14,14.4477153 14,15 C14,15.5522847 13.5522847,16 13,16 L10,16 C9.44771525,16 9,15.5522847 9,15 C9,14.4477153 9.44771525,14 10,14 L13,14 Z M16,10 C16.5522847,10 17,10.4477153 17,11 C17,11.5522847 16.5522847,12 16,12 L10,12 C9.44771525,12 9,11.5522847 9,11 C9,10.4477153 9.44771525,10 10,10 L16,10 Z"
                                fill="#0066FF"
                              />
                            </g>
                          </g>
                        </g>
                      </g>
                    </svg>
                  </template>
                </Icon>
                <span class="icon-text">转出课程</span>
              </div>
              <div class="h-1px  flex-1" :class="stepsStatus >= 2 ? 'bg-#06f' : 'bg-gray-200'" />
            </div>
            <div :class="{ 'max-w-88px': activeKey === 'he' }" class="flex gap-10px items-center flex-1">
              <div class=" flex items-center justify-center">
                <Icon :style="{ color: 'hotpink' }">
                  <template #component>
                    <svg
                      v-if="stepsStatus === 1" width="32px" height="32px" viewBox="0 0 32 32" version="1.1"
                      xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
                    >
                      <title>编组备份 8</title>
                      <g id="页面-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                        <g id="转课" transform="translate(-625.000000, -270.000000)">
                          <g id="组件/步骤条备份-7" transform="translate(304.000000, 270.000000)">
                            <g id="编组" transform="translate(321.000000, 0.000000)">
                              <rect id="矩形备份" x="0" y="0" width="32" height="32" />
                              <path
                                id="形状结合"
                                d="M22,15 C25.8659932,15 29,18.1340068 29,22 C29,25.8659932 25.8659932,29 22,29 C18.1340068,29 15,25.8659932 15,22 C15,18.1340068 18.1340068,15 22,15 Z M24,5 C25.5976809,5 26.9036609,6.24891996 26.9949073,7.82372721 L27,8 L27,13 C27,13.5522847 26.5522847,14 26,14 C25.4477153,14 25,13.5522847 25,13 L25,8 C25,7.48716416 24.6139598,7.06449284 24.1166211,7.00672773 L24,7 L8,7 C7.48716416,7 7.06449284,7.38604019 7.00672773,7.88337887 L7,8 L7,24 C7,24.5128358 7.38604019,24.9355072 7.88337887,24.9932723 L8,25 L13,25 C13.5522847,25 14,25.4477153 14,26 C14,26.5522847 13.5522847,27 13,27 L8,27 C6.40231912,27 5.09633912,25.75108 5.00509269,24.1762728 L5,24 L5,8 C5,6.40231912 6.24891996,5.09633912 7.82372721,5.00509269 L8,5 L24,5 Z M22,17 C19.2385763,17 17,19.2385763 17,22 C17,24.7614237 19.2385763,27 22,27 C24.7614237,27 27,24.7614237 27,22 C27,19.2385763 24.7614237,17 22,17 Z M25.1585046,19.7474233 C25.5421692,20.0831298 25.607569,20.64726 25.3293145,21.0589723 L25.2525767,21.1585046 L21.7525767,25.1585046 C21.3988641,25.5627476 20.7961308,25.6108497 20.3843337,25.2881619 L20.2928932,25.2071068 L18.7928932,23.7071068 C18.4023689,23.3165825 18.4023689,22.6834175 18.7928932,22.2928932 C19.1533772,21.9324093 19.7206082,21.9046797 20.1128994,22.2097046 L20.2071068,22.2928932 L20.951,23.036 L23.7474233,19.8414954 C24.1111054,19.4258588 24.742868,19.3837413 25.1585046,19.7474233 Z M13,18 C13.5522847,18 14,18.4477153 14,19 C14,19.5522847 13.5522847,20 13,20 L10,20 C9.44771525,20 9,19.5522847 9,19 C9,18.4477153 9.44771525,18 10,18 L13,18 Z M13,14 C13.5522847,14 14,14.4477153 14,15 C14,15.5522847 13.5522847,16 13,16 L10,16 C9.44771525,16 9,15.5522847 9,15 C9,14.4477153 9.44771525,14 10,14 L13,14 Z M16,10 C16.5522847,10 17,10.4477153 17,11 C17,11.5522847 16.5522847,12 16,12 L10,12 C9.44771525,12 9,11.5522847 9,11 C9,10.4477153 9.44771525,10 10,10 L16,10 Z"
                                fill="#CCCCCC"
                              />
                            </g>
                          </g>
                        </g>
                      </g>
                    </svg>
                    <svg
                      v-if="stepsStatus >= 2" width="32px" height="32px" viewBox="0 0 32 32" version="1.1"
                      xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
                    >
                      <title>编组备份 8</title>
                      <g id="页面-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                        <g id="转课2" transform="translate(-606.000000, -397.000000)">
                          <g id="编组-3" transform="translate(94.000000, 273.000000)">
                            <g id="组件/步骤条备份" transform="translate(191.000000, 124.000000)">
                              <g id="编组" transform="translate(321.000000, 0.000000)">
                                <rect id="矩形备份" x="0" y="0" width="32" height="32" />
                                <path
                                  id="形状结合"
                                  d="M22,15 C25.8659932,15 29,18.1340068 29,22 C29,25.8659932 25.8659932,29 22,29 C18.1340068,29 15,25.8659932 15,22 C15,18.1340068 18.1340068,15 22,15 Z M24,5 C25.5976809,5 26.9036609,6.24891996 26.9949073,7.82372721 L27,8 L27,13 C27,13.5522847 26.5522847,14 26,14 C25.4477153,14 25,13.5522847 25,13 L25,8 C25,7.48716416 24.6139598,7.06449284 24.1166211,7.00672773 L24,7 L8,7 C7.48716416,7 7.06449284,7.38604019 7.00672773,7.88337887 L7,8 L7,24 C7,24.5128358 7.38604019,24.9355072 7.88337887,24.9932723 L8,25 L13,25 C13.5522847,25 14,25.4477153 14,26 C14,26.5522847 13.5522847,27 13,27 L8,27 C6.40231912,27 5.09633912,25.75108 5.00509269,24.1762728 L5,24 L5,8 C5,6.40231912 6.24891996,5.09633912 7.82372721,5.00509269 L8,5 L24,5 Z M22,17 C19.2385763,17 17,19.2385763 17,22 C17,24.7614237 19.2385763,27 22,27 C24.7614237,27 27,24.7614237 27,22 C27,19.2385763 24.7614237,17 22,17 Z M25.1585046,19.7474233 C25.5421692,20.0831298 25.607569,20.64726 25.3293145,21.0589723 L25.2525767,21.1585046 L21.7525767,25.1585046 C21.3988641,25.5627476 20.7961308,25.6108497 20.3843337,25.2881619 L20.2928932,25.2071068 L18.7928932,23.7071068 C18.4023689,23.3165825 18.4023689,22.6834175 18.7928932,22.2928932 C19.1533772,21.9324093 19.7206082,21.9046797 20.1128994,22.2097046 L20.2071068,22.2928932 L20.951,23.036 L23.7474233,19.8414954 C24.1111054,19.4258588 24.742868,19.3837413 25.1585046,19.7474233 Z M13,18 C13.5522847,18 14,18.4477153 14,19 C14,19.5522847 13.5522847,20 13,20 L10,20 C9.44771525,20 9,19.5522847 9,19 C9,18.4477153 9.44771525,18 10,18 L13,18 Z M13,14 C13.5522847,14 14,14.4477153 14,15 C14,15.5522847 13.5522847,16 13,16 L10,16 C9.44771525,16 9,15.5522847 9,15 C9,14.4477153 9.44771525,14 10,14 L13,14 Z M16,10 C16.5522847,10 17,10.4477153 17,11 C17,11.5522847 16.5522847,12 16,12 L10,12 C9.44771525,12 9,11.5522847 9,11 C9,10.4477153 9.44771525,10 10,10 L16,10 Z"
                                  fill="#0066FF"
                                />
                              </g>
                            </g>
                          </g>
                        </g>
                      </g>
                    </svg>
                  </template>
                </Icon>
                <span class="">转入课程</span>
              </div>
              <div
                v-if="activeKey === 'me'" class="h-1px flex-1"
                :class="stepsStatus >= 3 ? 'bg-#06f' : 'bg-gray-200'"
              />
            </div>

            <div v-if="activeKey === 'me'" class="max-w-25% flex gap-10px items-center flex-1">
              <div class=" flex items-center justify-center">
                <Icon :style="{ color: 'hotpink' }">
                  <template #component>
                    <svg
                      v-if="stepsStatus === 2" width="33px" height="32px" viewBox="0 0 33 32" version="1.1"
                      xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
                    >
                      <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                        <g id="转课-转给自己" transform="translate(-723.000000, -267.000000)">
                          <g id="Steps/_Resource/IconStep+Normal" transform="translate(723.000000, 267.000000)">
                            <g id="编组" transform="translate(0.500000, 0.000000)">
                              <rect id="矩形" x="0" y="0" width="32" height="32" />
                              <path
                                id="形状结合"
                                d="M26.3101562,7 C27.240625,7 27.9976562,7.75754803 28,8.69389622 L28,8.69389622 L28,23.3061038 C28,24.242452 27.2453125,25 26.3125,25 L26.3125,25 L5.6875,25 C4.7546875,25 4,24.242452 4,23.3061038 L4,23.3061038 L4,8.69389622 C4,7.75754803 4.7546875,7 5.6875,7 L5.6875,7 Z M26,13 L6,13 L6,23 L26,23 L26,13 Z M9,18 L9.1166189,18.0067279 C9.61394944,18.0644944 10,18.4871753 10,19 L10,19 L9.9932721,19.1166189 C9.93550557,19.6139494 9.51282468,20 9,20 L9,20 L8.8833811,19.9932721 C8.38605056,19.9355056 8,19.5128247 8,19 L8,19 L8.0067279,18.8833811 C8.06449443,18.3860506 8.48717532,18 9,18 L9,18 Z M13,18 L13.1166189,18.0067279 C13.6139494,18.0644944 14,18.4871753 14,19 L14,19 L13.9932721,19.1166189 C13.9355056,19.6139494 13.5128247,20 13,20 L13,20 L12.8833811,19.9932721 C12.3860506,19.9355056 12,19.5128247 12,19 L12,19 L12.0067279,18.8833811 C12.0644944,18.3860506 12.4871753,18 13,18 L13,18 Z M26,9 L6,9 L6,11 L26,11 L26,9 Z"
                                fill="#CCCCCC"
                              />
                            </g>
                          </g>
                        </g>
                      </g>
                    </svg>
                    <svg
                      v-if="stepsStatus >= 3" width="33px" height="32px" viewBox="0 0 33 32" version="1.1"
                      xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink"
                    >
                      <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                        <g id="转课-转给自己" transform="translate(-723.000000, -267.000000)">
                          <g id="Steps/_Resource/IconStep+Normal" transform="translate(723.000000, 267.000000)">
                            <g id="编组" transform="translate(0.500000, 0.000000)">
                              <rect id="矩形" x="0" y="0" width="32" height="32" />
                              <path
                                id="形状结合"
                                d="M26.3101562,7 C27.240625,7 27.9976562,7.75754803 28,8.69389622 L28,8.69389622 L28,23.3061038 C28,24.242452 27.2453125,25 26.3125,25 L26.3125,25 L5.6875,25 C4.7546875,25 4,24.242452 4,23.3061038 L4,23.3061038 L4,8.69389622 C4,7.75754803 4.7546875,7 5.6875,7 L5.6875,7 Z M26,13 L6,13 L6,23 L26,23 L26,13 Z M9,18 L9.1166189,18.0067279 C9.61394944,18.0644944 10,18.4871753 10,19 L10,19 L9.9932721,19.1166189 C9.93550557,19.6139494 9.51282468,20 9,20 L9,20 L8.8833811,19.9932721 C8.38605056,19.9355056 8,19.5128247 8,19 L8,19 L8.0067279,18.8833811 C8.06449443,18.3860506 8.48717532,18 9,18 L9,18 Z M13,18 L13.1166189,18.0067279 C13.6139494,18.0644944 14,18.4871753 14,19 L14,19 L13.9932721,19.1166189 C13.9355056,19.6139494 13.5128247,20 13,20 L13,20 L12.8833811,19.9932721 C12.3860506,19.9355056 12,19.5128247 12,19 L12,19 L12.0067279,18.8833811 C12.0644944,18.3860506 12.4871753,18 13,18 L13,18 Z M26,9 L6,9 L6,11 L26,11 L26,9 Z"
                                fill="#0066FF"
                              />
                            </g>
                          </g>
                        </g>
                      </g>
                    </svg>
                  </template>
                </Icon>
                <span class="icon-text">订单支付</span>
              </div>
              <div class="h-1px bg-gray-200 flex-1" />
            </div>

            <div v-if="activeKey === 'me'" class="max-w-25% flex  items-center">
              <div class="step-icon flex items-center justify-center">
                <Icon :style="{ color: 'hotpink' }">
                  <template #component>
                    <svg
                      width="32px" height="32px" viewBox="0 0 32 32" version="1.1" xmlns="http://www.w3.org/2000/svg"
                      xmlns:xlink="http://www.w3.org/1999/xlink"
                    >
                      <g id="页面-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                        <g id="转课-转给自己" transform="translate(-1043.000000, -267.000000)">
                          <g id="编组" transform="translate(1043.000000, 267.000000)">
                            <rect id="矩形备份" x="0" y="0" width="32" height="32" />
                            <path
                              id="形状结合"
                              d="M26.3101562,7 C27.191653,7 27.9174883,7.67990461 27.993443,8.54762202 L28,8.69389622 L28.0015996,16.1026501 C29.2378445,17.3650451 30,19.093517 30,21 C30,24.8659932 26.8659932,28 23,28 C20.6218364,28 18.5206629,26.8140641 17.255617,25.0013298 L5.6875,25 C4.80378289,25 4.07993421,24.3200954 4.00618758,23.452378 L4,23.3061038 L4,8.69389622 C4,7.80682952 4.67733726,7.08023719 5.54177814,7.00621104 L5.6875,7 L26.3101562,7 Z M26,13 L6,13 L6,23 L16.2901893,23.0010432 C16.1013887,22.3670274 16,21.695354 16,21 C16,17.1340068 19.1340068,14 23,14 C24.073887,14 25.0912936,14.2418215 26.0007361,14.6739809 L26,13 Z M9,18 L9.1166189,18.0067279 C9.61394944,18.0644944 10,18.4871753 10,19 L10,19 L9.9932721,19.1166189 C9.93550557,19.6139494 9.51282468,20 9,20 L9,20 L8.8833811,19.9932721 C8.38605056,19.9355056 8,19.5128247 8,19 L8,19 L8.0067279,18.8833811 C8.06449443,18.3860506 8.48717532,18 9,18 L9,18 Z M13,18 L13.1166189,18.0067279 C13.6139494,18.0644944 14,18.4871753 14,19 L14,19 L13.9932721,19.1166189 C13.9355056,19.6139494 13.5128247,20 13,20 L13,20 L12.8833811,19.9932721 C12.3860506,19.9355056 12,19.5128247 12,19 L12,19 L12.0067279,18.8833811 C12.0644944,18.3860506 12.4871753,18 13,18 L13,18 Z M26,9 L6,9 L6,11 L26,11 L26,9 Z"
                              fill="#CCCCCC"
                            />
                            <path
                              id="路径"
                              d="M23,16 C25.7614237,16 28,18.2385763 28,21 C28,23.7614237 25.7614237,26 23,26 C20.2385763,26 18,23.7614237 18,21 C18,18.2385763 20.2385763,16 23,16 Z"
                              fill="#FEFEFE" fill-rule="nonzero"
                            />
                            <path
                              id="路径-2"
                              d="M24.7474233,18.8414954 C25.1111054,18.4258588 25.742868,18.3837413 26.1585046,18.7474233 C26.5421692,19.0831298 26.607569,19.64726 26.3293145,20.0589723 L26.2525767,20.1585046 L22.7525767,24.1585046 C22.3988641,24.5627476 21.7961308,24.6108497 21.3843337,24.2881619 L21.2928932,24.2071068 L19.7928932,22.7071068 C19.4023689,22.3165825 19.4023689,21.6834175 19.7928932,21.2928932 C20.1533772,20.9324093 20.7206082,20.9046797 21.1128994,21.2097046 L21.2071068,21.2928932 L21.951,22.036 L24.7474233,18.8414954 Z"
                              fill="#CCCCCC"
                            />
                          </g>
                        </g>
                      </g>
                    </svg>
                  </template>
                </Icon>
                <span class="icon-text">订单完成</span>
              </div>
            </div>
          </div>
        </div>
        <!-- 表单 -->
        <div v-if="stepsStatus === 1" class="bg-white  rounded-lg">
          <div class="p-4">
            <div class="text-lg font-medium mb-10px">
              12
            </div>
            <div class="mb-20px flex items-center gap-5px">
              <span class="bg-#e6f0ff text-#06f px-8px py-2px font-size-12px rounded-8px">
                班级课程
              </span>
              <span class="bg-#e6f0ff text-#06f px-8px py-2px font-size-12px rounded-8px">
                课时
              </span>
            </div>
            <div class="mb-4">
              <div class=" flex mb-20px">
                <span>剩余课时：</span>
                <span class="text-gray-400">1 课时</span>
              </div>
              <div class=" flex mb-2">
                <span class="">有效期至：</span>
                <span class="text-gray-400">无</span>
              </div>
            </div>
          </div>
          <div class="h-1px bg-gray-200" />
          <div class="p-4 mb-4">
            <div class="flex items-start">
              <a-form ref="fromRef" :model="formState" name="basic" autocomplete="off">
                <a-form-item label="转出数量" name="transferOutAmount" :rules="[{ required: true, message: '请输入' }]">
                  <a-input-number
                    v-model:value="formState.transferOutAmount" placeholder="请输入" :controls="false"
                    :precision="2" :min="0"
                  />
                </a-form-item>
              </a-form>
              <a-button type="link" class="ml-2" @click="handleAllOut">
                全部转出
              </a-button>
            </div>
          </div>
        </div>
        <!-- 课程内容 -->
        <div v-if="stepsStatus === 2" class="bg-white p-20px rounded-lg">
          <custom-title v-if="activeKey === 'me'" title="课程内容" font-size="18px" font-weight="800">
            <template v-if="flgCourse" #right>
              <div class="flex items-center gap-10px">
                <span class="text-13px font-400 text-#666">需要多退少补:</span>
                <a-switch v-model:checked="needAdjust" />
              </div>
            </template>
          </custom-title>
          <!-- 选择课程 -->
          <div
            v-if="!flgCourse && activeKey === 'me'"
            class="mt-20px flex flex-col items-center justify-center gap-15px"
          >
            <Icon :style="{ color: 'hotpink' }">
              <template #component>
                <svg width="56px" height="73px" viewBox="0 0 56 73">
                  <g id="\u9875\u9762-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                    <g id="\u753B\u677F\u5907\u4EFD-36" transform="translate(-605.000000, -1087.000000)">
                      <g id="\u7F16\u7EC4-10\u5907\u4EFD" transform="translate(25.000000, 996.000000)">
                        <g id="\u7F16\u7EC4-9" transform="translate(528.000000, 92.000000)">
                          <g id="\u7F16\u7EC4-17" transform="translate(52.000000, -0.000000)">
                            <path
                              id="\u5F62\u72B6\u7ED3\u5408"
                              d="M44,0 L56,12 L56,68 C56,70.209139 54.209139,72 52,72 L4,72 C1.790861,72 -7.10542736e-15,70.209139 -7.10542736e-15,68 L-7.10542736e-15,4 C-7.10542736e-15,1.790861 1.790861,0 4,0 L44,0 Z M34,48 L12,48 L11.85554,48.0068666 C11.0948881,48.0795513 10.5,48.7203039 10.5,49.5 C10.5,50.3284271 11.1715729,51 12,51 L12,51 L34,51 L34.14446,50.9931334 C34.9051119,50.9204487 35.5,50.2796961 35.5,49.5 C35.5,48.6715729 34.8284271,48 34,48 L34,48 Z M12,36 L11.85554,36.0068666 C11.0948881,36.0795513 10.5,36.7203039 10.5,37.5 C10.5,38.3284271 11.1715729,39 12,39 L12,39 L44,39 L44.14446,38.9931334 C44.9051119,38.9204487 45.5,38.2796961 45.5,37.5 C45.5,36.6715729 44.8284271,36 44,36 L44,36 L12,36 Z M12,24 L11.85554,24.0068666 C11.0948881,24.0795513 10.5,24.7203039 10.5,25.5 C10.5,26.3284271 11.1715729,27 12,27 L12,27 L44,27 L44.14446,26.9931334 C44.9051119,26.9204487 45.5,26.2796961 45.5,25.5 C45.5,24.6715729 44.8284271,24 44,24 L44,24 L12,24 Z"
                              fill="#B3D1FF"
                            />
                            <path
                              id="\u77E9\u5F62"
                              d="M44,-1.77635684e-15 L56,12 L48,12 C45.790861,12 44,10.209139 44,8 L44,-1.77635684e-15 Z"
                              fill="#E6F0FF"
                            />
                            <polygon
                              id="\u77E9\u5F62" fill="#0066FF"
                              transform="translate(15.000000, 8.000000) rotate(-360.000000) translate(-15.000000, -8.000000) "
                              points="10 0 20 0 20 16 15 12 10 16"
                            />
                          </g>
                        </g>
                      </g>
                    </g>
                  </g>
                </svg>
              </template>
            </Icon>
            <a-button type="primary" @click="showChooseCourses = true">
              选择课程
            </a-button>
          </div>
          <!-- 课程内容表单 -->
          <div v-else class="py-20px">
            <a-form
              v-if="activeKey === 'me' && !needAdjust" ref="courseFormRef" :model="courseForm"
              :label-col="{ style: { width: '100px' } }" :wrapper-col="{ span: 14 }"
            >
              <a-form-item label="转入课程" name="courseName" :rules="[{ required: true, message: '请选择转入课程' }]">
                <a-select
                  v-model:value="courseForm.courseName" placeholder="请选择课程" style="width: 100%"
                  :options="courseOptions"
                />
              </a-form-item>
              <a-form-item label="选择班级">
                <a-select
                  v-model:value="courseForm.className" placeholder="请选择班级" style="width: 100%"
                  :options="classOptions"
                />
              </a-form-item>
              <a-form-item label="收费方式" :required="true">
                <a-radio-group v-model:value="courseForm.type" button-style="solid">
                  <a-radio-button value="1">
                    按课时
                  </a-radio-button>
                  <a-radio-button value="2">
                    按时段
                  </a-radio-button>
                  <a-radio-button value="3">
                    按金额
                  </a-radio-button>
                </a-radio-group>
              </a-form-item>
              <template v-if="courseForm.type === '1'">
                <a-form-item label="转入课时" name="intoCourse" :rules="[{ required: true, message: '请输入转入课时' }]">
                  <div class="flex items-center">
                    <a-input-number
                      v-model:value="courseForm.intoCourse" :controls="false" :precision="2" :min="0"
                      placeholder="请输入" style="width: 50%"
                    />
                    <span>（转出：{{ formState.transferOutAmount?.toFixed(2) }}课时）</span>
                  </div>
                </a-form-item>
                <a-form-item label="赠送课时">
                  <a-input-number
                    v-model:value="courseForm.giveCourse" :controls="false" :precision="2" :min="0"
                    placeholder="请输入" style="width: 50%"
                  />
                </a-form-item>
                <a-space>
                  <a-form-item label="有效期至">
                    <a-radio-group v-model:value="courseForm.validity" style="width: 210px;" name="radioGroup">
                      <a-radio value="1">
                        不限制
                      </a-radio>
                      <a-radio value="2">
                        设置有效期至
                      </a-radio>
                    </a-radio-group>
                  </a-form-item>
                  <a-form-item
                    v-if="courseForm.validity === '2'" name="validityDate"
                    :rules="[{ required: true, message: '请选择有效期' }]"
                  >
                    <a-date-picker
                      v-model:value="courseForm.validityDate"
                      :disabled-date="disabledDate" style="width: 150px"
                      value-format="YYYY-MM-DD" placeholder="请选择"
                    />
                  </a-form-item>
                </a-space>
              </template>
              <template v-if="courseForm.type === '2'">
                <a-form-item name="intoDays" label="转入天数" :rules="[{ required: true, message: '请输入' }]">
                  <a-input-number
                    v-model:value="courseForm.intoDays" :min="0" :controls="false" :precision="0"
                    placeholder="转入天数" style="width: 50%"
                  />
                </a-form-item>
                <a-form-item label="赠送天数">
                  <a-input-number
                    v-model:value="courseForm.giveDays" :min="0" :controls="false" :precision="0"
                    placeholder="请输入" style="width: 50%"
                  />
                </a-form-item>
                <a-form-item label="开始时间" name="startDate" :rules="[{ required: true, message: '请输入' }]">
                  <a-date-picker
                    v-model:value="courseForm.startDate" style="width: 50%" value-format="YYYY-MM-DD"
                    placeholder="请选择"
                  />
                </a-form-item>
              </template>
              <template v-if="courseForm.type === '3'">
                <a-form-item label="转入金额" :required="true">
                  <span>233元（按金额模式，不可更改金额）</span>
                </a-form-item>
                <a-form-item label="赠送金额">
                  <a-input-number
                    v-model:value="courseForm.giveMoney" :controls="false" :min="0" :precision="2"
                    placeholder="请输入" style="width: 50%"
                  />
                </a-form-item>
                <a-form-item label="有效期至">
                  <a-radio-group v-model:value="courseForm.validity" name="radioGroup">
                    <a-radio value="1">
                      不限制
                    </a-radio>
                    <a-radio value="2">
                      设置有效期至
                    </a-radio>
                  </a-radio-group>
                </a-form-item>
              </template>
              <a-form-item label="订单标签">
                <a-select
                  v-model:value="courseForm.orderLabel" placeholder="请选择订单标签(最多可选5个)" style="width: 100%"
                  :options="[]"
                />
              </a-form-item>
              <a-form-item label="对内备注">
                <a-input v-model:value="courseForm.internalNotes" placeholder="此备注仅内部员工可见" />
              </a-form-item>
              <a-form-item label="对外备注">
                <a-input v-model:value="courseForm.externalRemarks" placeholder="此备注打印时将显示" />
              </a-form-item>
            </a-form>
            <!-- 需要多退少补内容 -->
            <div v-if="activeKey === 'me' && needAdjust" class="flex flex-col gap-15px">
              <div>
                <div
                  style="border-top-right-radius: 15px; border-top-left-radius: 15px; "
                  class="bg-#f0f5fe flex items-center justify-between p-10px"
                >
                  <div class="flex items-center gap-10px">
                    <span class="text-15px ">认知课程</span>
                    <span class="bg-#e6f0ff  text-13px text-#06f px-10px py-4px rounded-15px">班级授课</span>
                    <span class="bg-#e6f0ff  text-13px text-#06f px-10px py-4px rounded-15px">时段</span>
                    <a-select
                      v-model:value="adjustForm.newCourseType" placeholder="无"
                      :options="[{ label: '无', value: 1 }]" style="width: 100px"
                    />
                    <a-select
                      v-model:value="adjustForm.className" placeholder="选择班级"
                      :options="[{ label: '选择班级', value: 1 }]" style="width: 100px"
                    />
                  </div>
                  <div>
                    <span>合计：￥0.00</span>
                    <a-button type="link" style="margin-left: 10px">
                      取消选择
                    </a-button>
                  </div>
                </div>
                <div
                  style="border-bottom-right-radius: 15px; border-bottom-left-radius: 15px;"
                  class="bg-#fafafa flex items-center justify-between p-10px"
                >
                  <a-form ref="adjustFormRef" :model="adjustForm" layout="vertical" style="overflow: auto;">
                    <a-space style="gap:45px;">
                      <a-form-item label="报价单" name="priceListId" :rules="[{ required: true, message: '请选择报价单' }]">
                        <a-select
                          v-if="!saveAdjustForm" v-model:value="adjustForm.priceListId" placeholder="请选择报价单"
                          style="width: 180px;" :options="[{ label: '10课时|333元', value: 1 }]"
                        />
                        <div v-else style="width: 180px;">
                          {{ '10课时|333元' }}
                        </div>
                      </a-form-item>
                      <a-form-item label="购买份数" name="buyCount" :rules="[{ required: true, message: '请输入购买份数' }]">
                        <a-input-number
                          v-if="!saveAdjustForm" v-model:value="adjustForm.buyCount" style="width: 120px;"
                          :precision="0" :min="0" placeholder="请输入"
                        />
                        <div v-else style="width: 120px;">
                          1
                        </div>
                      </a-form-item>
                      <a-form-item label="赠送天数">
                        <a-input-number
                          v-if="!saveAdjustForm" v-model:value="adjustForm.giveDays" style="width: 120px;"
                          :precision="0" :min="0" placeholder="请输入"
                        />
                        <div v-else style="width: 120px;">
                          1
                        </div>
                      </a-form-item>
                      <a-form-item label="开始时间">
                        <a-date-picker
                          v-if="!saveAdjustForm" v-model:value="adjustForm.startDate" style="width: 140px;"
                          value-format="YYYY-MM-DD"
                        />
                        <div v-else style="width: 140px;">
                          {{ adjustForm.startDate }}
                        </div>
                      </a-form-item>
                      <a-form-item label="结束时间">
                        <div class="flex items-center w-140px ">
                          <span>2025-12-31（周五）</span>
                          <FormOutlined class="text-#06f" />
                        </div>
                      </a-form-item>
                      <a-form-item label="总天数（含增）">
                        <div class="flex items-center w-140px">
                          31
                        </div>
                      </a-form-item>
                      <a-form-item label="单课程优惠">
                        <div v-if="!saveAdjustForm" class="flex items-center w-300px">
                          <a-radio-group v-model:value="adjustForm.singleCourseDiscount">
                            <a-radio value="1">
                              无
                            </a-radio>
                            <a-radio value="2">
                              金额
                            </a-radio>
                            <a-radio value="3">
                              折扣
                            </a-radio>
                          </a-radio-group>
                          <template v-if="adjustForm.singleCourseDiscount !== '1'">
                            <a-input-number
                              v-model:value="adjustForm.singleCourseDiscountRate" class="mr-10px"
                              :precision="2" :min="0" placeholder="请输入" style="width:80px"
                            />
                            <span v-if="adjustForm.singleCourseDiscount === '2'">元</span>
                            <span v-if="adjustForm.singleCourseDiscount === '3'">折</span>
                          </template>
                        </div>
                        <div v-else class="w-300px">
                          无
                        </div>
                      </a-form-item>
                    </a-space>
                  </a-form>
                </div>
              </div>
              <div class="flex justify-end">
                <a-button type="primary" :ghost="saveAdjustForm" @click="handleAdjust">
                  {{ saveAdjustForm ? '编辑内容' : '保存' }}
                </a-button>
              </div>
            </div>
            <a-form
              v-if="activeKey === 'he'" ref="transferFormRef" :model="transferForm"
              :label-col="{ style: { width: '100px' } }" :wrapper-col="{ span: 14 }"
            >
              <a-form-item label="转入课程" :required="true">
                <span>认知课（不可更改）</span>
              </a-form-item>
              <a-form-item label="转入学员" name="studentName" :rules="[{ required: true, message: '请选择转入学员' }]">
                <a-select
                  v-model:value="transferForm.studentName" allow-clear show-search placeholder="搜索姓名/手机号"
                  style="width: 50%" option-label-prop="label"
                >
                  <a-select-option
                    v-for="item in stuListOptions" :key="item.id" :value="item.id" :data="item"
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
                        <div class="text-sm text-#666 leading-7">
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
              </a-form-item>
              <a-form-item label="选择班级">
                <a-select v-model:value="transferForm.className" placeholder="请选择班级" style="width: 50%" :options="[]" />
              </a-form-item>
              <a-form-item label="收费方式" :required="true">
                <span>按课时（不可更改）</span>
              </a-form-item>
              <a-form-item label="转入课时" :required="true">
                <span>1.00课时（不可更改）</span>
              </a-form-item>

              <a-space>
                <a-form-item label="有效期至">
                  <a-radio-group v-model:value="transferForm.validity" style="width: 210px;" name="radioGroup">
                    <a-radio value="1">
                      不限制
                    </a-radio>
                    <a-radio value="2">
                      设置有效期至
                    </a-radio>
                  </a-radio-group>
                </a-form-item>
                <a-form-item
                  v-if="transferForm.validity === '2'" name="validityDate"
                  :rules="[{ required: true, message: '请选择有效期' }]"
                >
                  <a-date-picker
                    v-model:value="transferForm.validityDate"
                    :disabled-date="disabledDate" style="width: 150px"
                    value-format="YYYY-MM-DD" placeholder="请选择"
                  />
                </a-form-item>
              </a-space>

              <a-form-item label="订单标签">
                <a-select
                  v-model:value="transferForm.orderLabel" placeholder="请选择订单标签(最多可选5个)" style="width: 100%"
                  :options="[]"
                />
              </a-form-item>
              <a-form-item label="对内备注">
                <a-input v-model:value="transferForm.internalNotes" placeholder="此备注仅内部员工可见" />
              </a-form-item>
              <a-form-item label="对外备注">
                <a-input v-model:value="transferForm.externalRemarks" placeholder="此备注打印时将显示" />
              </a-form-item>
            </a-form>
          </div>
        </div>
        <!-- 订单设置 -->
        <div v-if="activeKey === 'me' && needAdjust && stepsStatus === 2" class="bg-white p-20px rounded-lg mt-20px">
          <custom-title title="订单设置" font-size="18px" font-weight="800" />
          <div v-if="saveAdjustForm" class="mt-20px">
            <a-form :model="orderSettingForm" :label-col="{ style: { width: '100px' } }" :wrapper-col="{ span: 14 }">
              <a-form-item label="优惠卷">
                <a-button type="primary" ghost>
                  输入优惠卷码
                </a-button>
              </a-form-item>
              <a-form-item label="经办日期">
                <a-date-picker
                  v-model:value="orderSettingForm.handleDate"
                  :disabled-date="disabledEndDate" style="width: 50%"
                  value-format="YYYY-MM-DD"
                />
              </a-form-item>
              <a-form-item label="订单销售员">
                <a-select
                  v-model:value="orderSettingForm.salesName" placeholder="请选择" style="width: 50%"
                  :options="[]"
                />
              </a-form-item>
              <a-form-item label="订单标签">
                <a-select
                  v-model:value="orderSettingForm.orderLabel" placeholder="请选择订单标签(最多可选5个)" style="width: 100%"
                  :options="[]"
                />
              </a-form-item>
              <a-form-item label="对内备注">
                <a-input v-model:value="orderSettingForm.internalNotes" placeholder="此备注仅内部员工可见" />
              </a-form-item>
              <a-form-item label="对外备注">
                <a-input v-model:value="orderSettingForm.externalRemarks" placeholder="此备注打印时将显示" />
              </a-form-item>
            </a-form>
          </div>
        </div>
        <!-- 订单支付 -->
        <div v-if="stepsStatus === 3" class="bg-white p-20px rounded-lg mt-20px">
          <a-form :model="paymentForm" layout="vertical">
            <a-form-item label="实收金额" :required="true">
              <div class="flex items-center">
                <span class="text-48px font-700">￥3313</span>
                <span class="text-16px text-gray-500 mt-17px ml-16px">应收金额：3313.00</span>
              </div>
              <a-radio-group v-model:value="paymentForm.model" button-style="solid">
                <a-radio-button value="1">
                  已收款只记账
                </a-radio-button>
                <a-radio-button value="2">
                  面对面收款
                </a-radio-button>
              </a-radio-group>
            </a-form-item>
            <!-- 退款方式 -->
            <a-form-item name="payType">
              <div class="payList mt-4 text-#666">
                <div>
                  <span class=" mr-2px">*</span>收款方式
                  <span class="payList-tip ml-2">请选择</span>
                </div>
                <div class="pay">
                  <a-radio-group v-model:value="paymentForm.payType" class="custom-radio">
                    <a-space :size="16" class="flex-wrap">
                      <template v-for="(item, index) in checkOptions" :key="index">
                        <label class="pay-box" :class="{ active: paymentForm.payType === item.id }">
                          <span> <img :src="item.img" alt=""> {{ item.label }}</span>
                          <a-radio :value="item.id" />
                        </label>
                      </template>
                    </a-space>
                  </a-radio-group>
                </div>
              </div>
            </a-form-item>
            <a-form-item>
              <div class="text-#666 flex flex-items-center mb-6px">
                账单备注：
              </div>
              <a-textarea
                v-model:value="paymentForm.billRemarks" placeholder="请输入内容，最多100字"
                :auto-size="{ minRows: 2, maxRows: 5 }"
              />
            </a-form-item>
            <a-form-item class="w-80%">
              <a-form-item-rest>
                <div class="mt--10px">
                  <a-upload
                    v-model:file-list="paymentForm.fileList"
                    action="https://www.mocky.io/v2/5cc8019d300000980a055e76" list-type="picture-card"
                    accept=".jpg,.jpeg,.png" @preview="handlePreview"
                  >
                    <div v-if="paymentForm.fileList.length < 3">
                      <!-- <plus-outlined class="text-20px" /> -->
                      <PlusOutlined class="text-24px" />
                    </div>
                  </a-upload>
                  <span class="text-#888 text-12px">最多上传3张，支持JPG、JPEG、PNG，单张图片不超过 4 MB</span>
                  <a-modal :open="previewVisible" :title="previewTitle" :footer="null" @cancel="handleCancelPreview">
                    <img alt="example" style="width: 100%" :src="previewImage">
                  </a-modal>
                </div>
              </a-form-item-rest>
            </a-form-item>
          </a-form>
          <a-form layout="vertical" :model="paymentForm" />
        </div>
      </div>

      <!-- 底部按钮 -->
      <template #footer>
        <div class="flex gap-20px justify-between items-center px-20px py-10px">
          <div>
            <a-button v-if="stepsStatus > 1" size="large" type="primary" ghost @click="handleCancel">
              返回上一步
            </a-button>
          </div>
          <div v-if="!needAdjust" class="flex gap-20px justify-between items-center">
            <div v-if="[1, 2].includes(stepsStatus)" class="flex flex-col items-end">
              <div class="text-lg font-medium">
                总计学费：￥0.00
              </div>
              <div v-if="stepsStatus === 1" class="text-sm text-gray-500">
                转出 {{ formState.transferOutAmount }}课时，转赠 0 课时（赠送不计入总计）
              </div>
              <div v-if="stepsStatus === 2" class="text-sm text-gray-500">
                转入0，赠送0
              </div>
            </div>
            <a-button
              type="primary" size="large" style="width: 140px;height: 48px;font-size: 20px;"
              @click="handleNext"
            >
              {{ stepsStatus === 3 ? '提交' : '下一步' }}
            </a-button>
          </div>
          <!-- 确定订单设置 -->
          <div v-if="saveAdjustForm" class="flex items-center gap-10px">
            <div class="flex flex-col items-end">
              <div class="text-lg font-medium">
                总计学费：￥0.00
              </div>
            </div>
            <a-button
              style="width: 140px;height: 48px;font-size: 20px;" type="primary" size="large"
              @click="showOrderSettingsConfirmation = true"
            >
              确定
            </a-button>
          </div>
        </div>
      </template>
    </a-drawer>
    <chooseCoursesModal v-model="showChooseCourses" @change="handleChooseCoursesChange" />
    <confirmInformationCoursesModal
      v-model="showConfirmInformationCourses"
      @confirm="handleConfirmInformationCoursesConfirm"
    />
    <orderSettingsConfirmationModal
      v-model="showOrderSettingsConfirmation"
      @confirm="handleOrderSettingsConfirmationConfirm"
    />
    <orderCompletedDrawer v-model="showOrderCompletedDrawer" />
  </div>
</template>

<style lang="less" scoped>
:deep(.ant-form-item) {
  margin-bottom: 16px;
}

:deep(.ant-space-item) {
  flex-grow: 1;
}

/* 添加旋转动画 */
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

.payList {
  span {
    color: red;
  }

  span.payList-tip {
    color: var(--pro-ant-color-primary);
  }
}

.pay {
  margin-top: 10px;

  .pay-box {
    max-width: 200px;
    border: 1px solid #eee;
    padding: 12px 16px;
    display: flex;
    align-items: center;
    border-radius: 6px;
    margin-right: 6px;
    user-select: none;
    cursor: pointer;

    &:hover {
      border-color: var(--pro-ant-color-primary);
    }

    span {
      color: #000;
      margin-right: 20px;
      display: flex;
      align-items: center;

      img {
        width: 20px;
        height: 20px;
        margin-right: 6px;
      }
    }
  }

  .active {
    border-color: var(--pro-ant-color-primary);
  }
}

.payPrice {
  :deep(.ant-input-number-input) {
    height: 80px;
    line-height: 80px;
    font-family: "DIN alternate", sans-serif;
  }

  :deep(.ant-input-number-group-addon) {
    background: transparent;
    border: none;
  }

}
</style>
