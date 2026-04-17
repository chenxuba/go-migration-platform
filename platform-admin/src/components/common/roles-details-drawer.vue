<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { useDrawer } from '@/composables/useDrawer'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  roleId: {
    type: Number,
    default: null,
  },
  details: {
    type: Object,
    default: {},
  },
})
const emit = defineEmits(['update:open', 'onEditSuccess'])
const activeKey = ref('1')
const { openDrawer } = useDrawer(props, emit)

// 响应式宽度
const drawerWidth = ref('800px')

// 计算响应式宽度
const updateDrawerWidth = () => {
  const windowWidth = window.innerWidth
  if (windowWidth <= 768) {
    // 移动端 - 全屏
    drawerWidth.value = '100vw'
  } else if (windowWidth <= 1024) {
    // 平板 - 80%宽度，最大600px
    drawerWidth.value = Math.min(windowWidth * 0.8, 600) + 'px'
  } else if (windowWidth <= 1440) {
    // 小桌面 - 60%宽度，最大700px
    drawerWidth.value = Math.min(windowWidth * 0.6, 700) + 'px'
  } else {
    // 大桌面 - 固定800px
    drawerWidth.value = '800px'
  }
}

// 组件挂载时计算初始宽度并添加监听器
onMounted(() => {
  updateDrawerWidth()
  window.addEventListener('resize', updateDrawerWidth)
})

// 组件卸载时移除监听器
onUnmounted(() => {
  window.removeEventListener('resize', updateDrawerWidth)
})
// 获取用户信息
const userStore = useUserStore()
// 计算属性获取机构名称
const orgName = computed(() => userStore.userInfo?.orgName || '总机构')

// 获取 functions-permiss 组件的引用
const functionsPermissRef = ref(null)

// 编辑角色抽屉控制
const createRolesDrawerOpen = ref(false)

// 本地角色详情数据，用于在编辑后更新显示
const localDetails = ref({})

// 监听 props.details 变化，同步到本地数据
watch(
  () => props.details,
  (newDetails) => {
    if (newDetails) {
      localDetails.value = { ...newDetails }
    }
  },
  { immediate: true, deep: true },
)


// 编辑角色按钮点击事件
function handleEditRole() {
  createRolesDrawerOpen.value = true
}

// 编辑成功回调
function handleEditSuccess(data) {
  createRolesDrawerOpen.value = false

  // 更新本地数据
  if (data) {
    localDetails.value = {
      ...data,
      updateName: data.updateName || localDetails.value.updateName,
      updateTime: data.updateTime || localDetails.value.updateTime
    }
  }

  // 刷新权限组件数据
  nextTick(() => {
    if (functionsPermissRef.value && functionsPermissRef.value.refreshPermissions) {
      functionsPermissRef.value.refreshPermissions()
    }
  })

  // 向父组件发送编辑成功事件，需要刷新角色列表数据以更新权限个数
  emit('onEditSuccess', data)
}

// 监听抽屉打开关闭事件
watch(
  () => openDrawer.value,
  (newValue, oldValue) => {
    activeKey.value = '1'
    // 当抽屉打开时，刷新权限数据
    if (newValue === true) {
      nextTick(() => {
        if (functionsPermissRef.value && functionsPermissRef.value.refreshPermissions) {
          functionsPermissRef.value.refreshPermissions()
        }
      })
    }
    // 当抽屉从打开状态变为关闭状态时
    if (oldValue === true && newValue === false) {
      // 重置权限节点展开状态
      if (functionsPermissRef.value && functionsPermissRef.value.resetExpandState) {
        functionsPermissRef.value.resetExpandState()
      }
    }
  },
)
</script>

<template>
  <div>
    <a-drawer v-model:open="openDrawer" :push="{ distance: 80 }" :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false" :width="drawerWidth" placement="right">
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            角色详情
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="contenter">
        <div class="p-24px pb-0">
          <div class="flex justify-between items-center mb-10px">
            <span class="text-20px font500">{{ localDetails.roleName }}</span>
            <a-button v-if="localDetails.isDefault != null" type="primary" :disabled="localDetails.isDefault"
              @click="handleEditRole">
              编辑
            </a-button>
            <a-skeleton-button v-else :active="true" shape="default" :block="false" />
          </div>
          <div class="flex justify-between items-center mt-8px text-#666 mb-10px">
            <span>所属机构：<span class="text-#222">{{ orgName }}</span>
            </span>
            <span v-if="!localDetails.isDefault && localDetails.updateName">最近编辑时间：<span class="text-#222">{{
              dayjs(localDetails.updateTime).format("YYYY-MM-DD HH:mm") }}
                {{ localDetails.updateName || "-" }}</span></span>
          </div>
          <div class="text-#666 mb-10px">
            <span class="flex">
              <span class="whitespace-nowrap">角色描述：</span>
              <clamped-text class="text-#222" :text="localDetails.description || '-'" :lines="1" /></span>
          </div>
        </div>
        <!-- tabs 功能与权限 任职员工 -->
        <div class="tabs">
          <a-tabs v-model:active-key="activeKey">
            <a-tab-pane key="1" tab="功能与权限">
              <functions-permiss ref="functionsPermissRef" :role-id="roleId" v-model:details="localDetails" />
            </a-tab-pane>
            <a-tab-pane key="2" tab="任职员工">
              <roles-staff-list :role-id="roleId" />
            </a-tab-pane>
          </a-tabs>
        </div>
      </div>
    </a-drawer>

    <!-- 编辑角色抽屉 -->
    <create-roles-drawer v-model:open="createRolesDrawerOpen" :role-id="roleId" @on-success="handleEditSuccess" />
  </div>
</template>

<style lang="less" scoped>
.contenter {
  background: #fff;
}

.tabs {
  width: 100%;
  border-radius: 10px;
  line-height: 40px;

  :deep(.ant-tabs-nav) {
    border-radius: 16px;
    padding: 0 24px;
  }

  :deep(.ant-tabs-tab) {
    font-size: 16px;
  }

  :deep(.ant-tabs-ink-bar) {
    text-align: center;
    height: 9px !important;
    background: transparent;
    bottom: 1px !important;

    &::after {
      position: absolute;
      top: 0;
      left: calc(50% - 12px);
      width: 24px !important;
      height: 4px !important;
      border-radius: 2px;
      background-color: var(--pro-ant-color-primary);
      content: "";
    }
  }

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
</style>
