<script setup>
import { message } from 'ant-design-vue'
import { ref } from 'vue'
import SettingsNavigator from './components/SettingsNavigator.vue'
// 导入二级页面组件
import CourseOrEndSetting from './courseOrEndSetting.vue'
import RegistrationSettings from './registrationSettings.vue'
import StudentAttributeSettings from './studentAttributeSettings.vue'
import LeadSettings from './leadSettings.vue'
import channelSettings from './lead-settings/channel-settings.vue'
import CampusDataClear from './campusDataClear.vue'

import { useModalStore } from '~/stores/modal'

const modalStore = useModalStore()

// 页面导航状态
const navigation = ref({
  level1: true, // 是否在一级页面
  level2Path: '', // 二级页面路径
  level3Path: '', // 三级页面路径
})

// 设置列表数据
const settingsList = ref([
  {
    title: '课程与课消设置',
    description: '授课方式、收款方式、课消规则等相关设置',
    path: '/course-settings',
  },
  {
    title: '报名/退课设置',
    description: '支持在报名/退票/退费场景时，各项管理规则的设置',
    path: '/registration-settings',
  },
  {
    title: '人脸考勤设置',
    description: '人脸考勤关联点名课消、签到签退通知等相关设置',
    path: '/facial-attendance-settings',
  },
  {
    title: '点名设置',
    description: '自动点名相关设置',
    path: '/roll-call-settings',
  },
  {
    title: '成长档案类型设置',
    description: '支持设置自定义或校园档案类型',
    path: '/profile-type-settings',
  },
  {
    title: '课堂点评设置',
    description: '家长反馈、老师点评等相关设置',
    path: '/evaluation-settings',
  },
  {
    title: '收款账户设置',
    description: '在线收款、关联收款账户等相关设置',
    path: '/payment-account-settings',
  },
  {
    title: '线索设置',
    description: '意向学员的来源渠道分类、待分配学员规则等相关设置',
    path: '/lead-settings',
  },
  {
    title: '学员属性设置',
    path: '/student-attribute-settings',
  },
  {
    title: '班级设置',
    path: '/class-settings',
  },
  {
    title: '升期设置',
    status: { type: 'enabled', text: '已开启' },
    path: '/term-settings',
  },
  {
    title: '教室设置',
    path: '/classroom-settings',
  },
  {
    title: '短信设置',
    status: { type: 'warning', text: '短信余额不足' },
    path: '/sms-settings',
  },
  {
    title: '试听设置',
    path: '/trial-settings',
  },
  {
    title: '约课设置',
    path: '/appointment-settings',
  },
  {
    title: '家校设置',
    path: '/home-school-settings',
  },
  {
    title: '公众号设置',
    description: '对接机构公众号，家校服务再升级',
    path: '/wechat-settings',
  },
  {
    title: '出入库管理设置',
    path: '/inventory-settings',
  },
  {
    title: '订单收据显示机构',
    description: '开启后，订单在查看/打印/下载收据时会同时显示机构机构信息；关闭后，仅显示机构，不显示机构',
    hasToggle: true,
    toggleValue: true,
    path: '/order-receipt-settings',
  },
  {
    title: '同行资讯及服务管理',
    description: '开启后，App 端首页会显示同行资讯；同行都在看等消息提示；PC 端首页会显示信息服务快捷入口',
    hasToggle: true,
    toggleValue: true,
    path: '/industry-news-settings',
  },
  {
    title: '机构数据清空',
    path: '/campus-data-clear', 
  },
  {
    title: '分配业绩',
    description: '自动分配业绩相关设置',
    path: '/performance-allocation',
  },

])

// 获取当前应该显示的组件
const getCurrentComponent = computed(() => {
  console.log(navigation.value)
  // 如果有三级路径
  if (navigation.value.level3Path) {
    // 返回对应的三级页面组件
    return getLevel3Component(navigation.value.level2Path, navigation.value.level3Path)
  }

  // 如果有二级路径
  if (navigation.value.level2Path) {
    // 返回对应的二级页面组件
    return getLevel2Component(navigation.value.level2Path)
  }

  // 默认返回null，显示一级页面
  return null
})

// 获取二级页面组件
function getLevel2Component(path) {
  switch (path) {
    case '/course-settings': return CourseOrEndSetting // 课程与课消设置
    case '/registration-settings': return RegistrationSettings // 报名/退课设置
    case '/lead-settings': return LeadSettings // 线索设置
    case '/student-attribute-settings' : return StudentAttributeSettings // 学员属性设置
    case '/campus-data-clear': return CampusDataClear // 机构数据清空
    // ... 其他二级页面
    default: return null
  }
}

// 获取三级页面组件
function getLevel3Component(level2Path, level3Path) {
  const fullPath = `${level2Path}${level3Path}`
  // 根据完整路径返回对应的三级页面组件
  switch (fullPath) {
    // 示例：课程设置下的授课方式设置
    case '/lead-settings/channel-settings':
      return channelSettings
    // ... 其他三级页面
    default: return null
  }
}

// 导航到二级页面
function navigateToSetting(item) {
  navigation.value.level1 = false
  navigation.value.level2Path = item.path
  navigation.value.level3Path = ''

  // 更新模态框标题
  modalStore.updateModalTitle(item.title)
  modalStore.updateModalBackButton(true)
}

// 导航到三级页面
function navigateToSubSetting(subPath, title) {
  navigation.value.level3Path = subPath

  // 更新模态框标题，添加层级指示
  const currentTitle = modalStore.modalTitle
  modalStore.updateModalTitle(`${title}`)
}

// 返回上一级
function goBack() {
  // 如果在三级页面，返回二级页面
  if (navigation.value.level3Path) {
    navigation.value.level3Path = ''

    // 恢复二级页面标题
    const currentItem = settingsList.value.find(item => item.path === navigation.value.level2Path)
    if (currentItem) {
      modalStore.updateModalTitle(currentItem.title)
    }
    return
  }

  // 如果在二级页面，返回一级页面
  navigation.value.level1 = true
  navigation.value.level2Path = ''
  modalStore.updateModalBackButton(false)
  modalStore.updateModalTitle('业务设置')
}

// 暴露方法给父组件
defineExpose({ goBack })

function handleSave() {
  message.success('保存成功')
  modalStore.closeModal()
}

function handleCancel() {
  modalStore.closeModal()
}
</script>

<template>
  <div class="business-settings scrollbar">
    <!-- 一级页面：设置列表 -->
    <SettingsNavigator
      v-if="navigation.level1"
      :settings-list="settingsList"
      @navigate="navigateToSetting"
    />

    <!-- 二级和三级页面：动态组件 -->
    <component
      :is="getCurrentComponent"
      v-else
      @navigate-to-sub="navigateToSubSetting"
    />
  </div>
</template>

<style lang="less" scoped>
.business-settings {
  height: calc(100vh - 140px);
  background-color: #fff;
  overflow-y: auto;
  border-top-left-radius: 0;
  border-top-right-radius: 0;
  border-bottom-left-radius: 10px;
  border-bottom-right-radius: 10px;
  // padding-bottom: 50px;
}
</style>
