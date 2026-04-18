<script setup>
import { RightOutlined } from '@ant-design/icons-vue'
import { ref } from 'vue'
import { useUserStore } from '~/stores/user'

const emit = defineEmits(['navigateToSub'])

const userStore = useUserStore()
const { isAdmin } = userStore

const toggleValue1 = ref(false)
const toggleValue2 = ref(false)
const toggleValue3 = ref(false)
const toggleValue4 = ref(false)
const toggleValue5 = ref(false)

function handleClick() {
  emit('navigateToSub', '/channel-settings', '渠道管理')
}
</script>

<template>
  <div class="registration-settings">
    <!-- 渠道设置 -->
    <div
      class="mt-8px bg-white px-15px py-10px  border-t-0 border-b-1px border-l-0 border-r-0 border-#eee border-solid
      flex justify-between items-center cursor-pointer"
      @click="handleClick"
    >
      <div>
        <div class="text-16px">
          渠道设置
        </div>
        <div class="text-#999 text-12px mt-2px">
          意向学员的来源渠道分类相关设置
        </div>
      </div>
      <div class="item-status ml-20px">
        <RightOutlined />
      </div>
    </div>
    <template v-if="isAdmin">
      <!-- 是否开启公有池 -->
      <div
        class="mt-8px bg-white px-15px py-10px  border-t-0 border-b-1px border-l-0 border-r-0 border-#eee border-solid
      flex justify-between items-center"
      >
        <div>
          <div class="text-16px">
            是否开启公有池
          </div>
          <div class="text-#999 text-12px mt-2px">
            开启后，系统会自动将没有销售的意向学员汇总到公有池，方便管理和再次分配
          </div>
        </div>
        <div class="item-status ml-20px">
          <a-switch v-model:checked="toggleValue1" />
        </div>
      </div>
      <!-- 未跟进时间超过3天的学员 -->
      <div
        class="mt-8px bg-white px-15px py-10px  border-t-0 border-b-1px border-l-0 border-r-0 border-#eee border-solid
      flex justify-between items-center"
      >
        <div>
          <div class="text-16px">
            未跟进时间超过 <span class="text-#06f"> 3 </span> 天的学员
          </div>
          <div class="text-#999 text-12px mt-2px">
            超过以上时间的意向学员将自动进入公有池
          </div>
        </div>
        <div class="item-status ml-20px">
          <a-button type="link">
            编辑
          </a-button>
        </div>
      </div>
    </template>
  </div>
</template>

<style lang="less" scoped>
.registration-settings {
  padding: 5px 0px;
  background-color: #f6f7f8;
  // height: 100%;

  :deep(.ant-collapse-header) {
    padding: 0 !important;
  }

  :deep(.ant-collapse-expand-icon) {
    position: absolute;
    right: 12px;
    padding-inline-end: 0px !important;
  }

  :deep(.ant-collapse-content-box) {
    padding: 0px !important;
  }

  .settings-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 12px 12px 16px;
    border-top: 1px solid #f0f0f0;
    cursor: pointer;
  }

  .item-title {
    font-size: 14px;
    color: rgba(0, 0, 0, 0.85);
  }

  .item-description {
    font-size: 12px;
    color: rgba(0, 0, 0, 0.45);
  }
}
</style>
