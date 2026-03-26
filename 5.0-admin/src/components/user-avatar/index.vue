<script setup lang="ts">
import { LogoutOutlined, ProfileOutlined, UserOutlined } from '@ant-design/icons-vue'

const message = useMessage()
const userStore = useUserStore()
const multiTabStore = useMultiTab()
const layoutMenuStore = useLayoutMenu()
const router = useRouter()
const { avatar, nickname } = storeToRefs(userStore)

async function handleClick({ key }: any) {
  if (key === 'logout') {
    const hide = message.loading('退出登录...', 0)
    try {
      await userStore.logout()
    }
    finally {
      hide()
      message.success('退出登录成功', 3)
      router.push({
        path: '/login',
      }).then(() => {
        multiTabStore.clear()
        layoutMenuStore.clear()
      })
    }
  }
}
</script>

<template>
  <a-dropdown>
    <span flex items-center h-48px px-12px cursor-pointer class="transition-all-300">
      <a-avatar v-if="avatar" :src="avatar" mr-8px size="small" />
      <span
        class="inline-block w-30px h-30px bg-#06f rounded-10 flex justify-center text-#fff font500 mr-6px items-center"
        v-else>{{ nickname?.slice(0, 1) }}</span>
      <span class="anticon">
        {{ nickname }}
      </span>
    </span>
    <template #overlay>
      <a-menu @click="handleClick">
        <a-menu-item key="0">
          <template #icon>
            <UserOutlined />
          </template>
          <RouterLink to="/account/center">
            个人中心
          </RouterLink>
        </a-menu-item>
        <a-menu-item key="1">
          <template #icon>
            <ProfileOutlined />
          </template>
          <RouterLink to="/account/settings">
            个人设置
          </RouterLink>
        </a-menu-item>
        <a-menu-divider />
        <a-menu-item key="logout">
          <template #icon>
            <LogoutOutlined />
          </template>
          退出登录
        </a-menu-item>
      </a-menu>
    </template>
  </a-dropdown>
</template>
