<script setup>
import { CloseOutlined, DownOutlined } from '@ant-design/icons-vue'

const emit = defineEmits(['confirm'])

const open = defineModel({
  default: false,
})

const openOrderDetailDrawer = ref(false)

function confirm() {
  emit('confirm', true)
  open.value = false
}
function cancel() {
  emit('confirm', false)
  open.value = false
}
</script>

<template>
  <a-drawer
    v-model:open="open" :mask-closable="false" :keyboard="false" :push="false"
    :body-style="{ padding: '0', background: '#fff' }" :closable="false" width="1165px" placement="right"
  >
    <!-- 自定义头部 -->
    <template #title>
      <div class="custom-header flex justify-between h-4 flex-items-center">
        <div class="text-5">
          订单完成
        </div>
        <a-button type="text" class="close-btn" @click="cancel">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <span class="bg-#f6f7f8 flex h-2 " />
    <div class="flex justify-center border border-b-#eee border-solid border-x-none border-t-none">
      <a-result status="success">
        <template #title>
          <span class="text-8 text-#222" style="font-family: 'DIN alternate', sans-serif;">¥22.00</span>
        </template>
        <template #subTitle>
          <span class="text-4 text-#666">- 订单完成 -</span>
        </template>
        <template #extra>
          <a-button type="primary" @click="openOrderDetailDrawer = true">
            查看订单详情
          </a-button>
          <a-dropdown>
            <template #overlay>
              <a-menu>
                <a-menu-item key="0">
                  打印收据
                </a-menu-item>
                <a-menu-item key="1">
                  下载收据
                </a-menu-item>
                <a-menu-item key="3">
                  发送短信
                </a-menu-item>
              </a-menu>
            </template>
            <a-button>
              查看收据
              <DownOutlined :style="{ fontSize: '10px' }" />
            </a-button>
          </a-dropdown>
        </template>
      </a-result>
    </div>
    <div class="px8 py6">
      <a-descriptions :column="3" :content-style="{ color: '#888' }">
        <a-descriptions-item label="储值账户：">
          niuniuAccount
        </a-descriptions-item>
        <a-descriptions-item label="关联学员：">
          妞妞
        </a-descriptions-item>
        <a-descriptions-item label="收款方式：">
          POS机支付
        </a-descriptions-item>
        <a-descriptions-item label="收款账户：">
          默认账户
        </a-descriptions-item>
        <a-descriptions-item label="支付单号：">
          -
        </a-descriptions-item>
        <a-descriptions-item label="对方账户：">
          -
        </a-descriptions-item>
        <a-descriptions-item label="支付时间：">
          2025-04-15 12:22
        </a-descriptions-item>
        <a-descriptions-item label="账单备注：">
          无
        </a-descriptions-item>
      </a-descriptions>
    </div>
    <order-detail-drawer v-model:open="openOrderDetailDrawer" />
  </a-drawer>
</template>

<style scoped lang="less"></style>
