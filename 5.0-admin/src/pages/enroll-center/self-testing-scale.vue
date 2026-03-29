<script setup>
const displayArray = ref([
  'remaining', // 剩余数量
  'recommended', // 是否被推荐
  'age', // 年龄
  'notFollowDays', // 未跟进天数
  'assignTime', // 分配时间
  'nextFollowTime', // 下次跟进时间
  'lastFollowTime', // 最近跟进时间
  'hasSalesPerson', // 是否有销售员
  'salesPerson', // 销售员
  'intention', // 意向度
  'followStatus', // 跟进状态
  'createUser', // 创建人
  'createTime', // 创建时间
  'applyTime', // 申请时间
  'payTime', // 支付日期
  'classEndingTime', // 结课时间
  'classStopTime', // 停课时间
  'intentionCourse', // 意向课程
  'reference', // 推荐人
  'department', // 所属部门（仅在 type='dpt' 时显示）
  'channelStatus', // 渠道状态
  'channelType', // 渠道类型
  'subject', // 科目
  'courseCategory', // 课程类别
  'studentStatus', // 学员状态
  'currentStatus', // 当前状态
  'orNotFenClass', // 是否分班
  'billingMode', // 收费模式
  'isSetExpirationDate', // 是否设置有效期
  'openClassStatus', // 开班状态
  'doYouSchedule', // 是否排课
  'channelCategoryType', // 渠道分类
  'trialPurchaseStatus', // 体验课购买状态
  'channelCategory', // 渠道
  'sex', // 性别
  'birthday', // 生日
  'wxChat', // 微信号
  'school', // 就读学校
  'address', // 家庭住址
  'hobbies', // 兴趣爱好
  'followMethod', // 跟进方式
  'visitStatus', // 回访状态
  'stuStatus', // 学员状态
  'stuPhoneSearch', // 学员/电话
  'recommend', // 推荐人
  'followTime', // 跟进时间
  'teachingMethod', // 授课方式
  'sellStatus', // 售卖状态
  'hasTrialPrice', // 是否有体验价
  'isMicroSchoolSale', // 是否开启微校售卖
  'isMicroSchoolDisplay', // 是否开启微校展示
  'payrollStatus', // 工资条状态
  'className', // 班级名称（多选）
  'enrolledCourse', // 报读课程（多选）
  'validityPeriod', // 有效期至（复用下次跟进时间组件）
  'classTeacher', // 班主任（复用销售员组件）
  'isArrears', // 是否欠费（复用是否有体验价组件）
  'lastClassTime', // 最近上课时间（复用下次跟进时间组件）
])
const arr = ['courseCategory', 'teachingMethod', 'billingMode', 'sellStatus', 'hasTrialPrice', 'isMicroSchoolSale', 'isMicroSchoolDisplay']
</script>

<template>
  <div class="filter-wrap  bg-white  pl-3 pr-3 rounded-4">
    <all-filter :display-array="displayArray" :is-quick-show="true" :is-show-search-stu-phone="true" />
  </div>
</template>

<style scoped lang="less"></style>
