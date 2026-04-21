<script setup lang="ts">
import { GlobalOutlined } from '@ant-design/icons-vue'

const { locale, setLocale } = useI18nLocale()
function handleClick({ key }: any) {
  setLocale(key)
}
</script>

<template>
  <a-dropdown>
    <span flex items-center h-48px px-12px text-16px cursor-pointer class="transition-all-300">
      <GlobalOutlined class="anticon" />
    </span>
    <template #overlay>
      <a-menu :selected-keys="[locale]" @click="handleClick">
        <a-menu-item key="zh-CN">
          <template #icon>
            <span>
              🇨🇳
            </span>
          </template>
          简体中文
        </a-menu-item>
        <a-menu-item key="en-US">
          <template #icon>
            <span>
              🇺🇸
            </span>
          </template>
          English
        </a-menu-item>
        <a-menu-item key="ms-MY">
          <template #icon>
            <span>
              🇲🇾
            </span>
          </template>
          Bahasa Melayu
        </a-menu-item>
      </a-menu>
    </template>
  </a-dropdown>
</template>
