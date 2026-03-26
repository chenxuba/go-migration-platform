<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '~@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const schoolName = computed(() => userStore.userInfo?.orgName || '总机构')

function replaceToWorkbench() {
  window.location.replace(`${window.location.origin}${window.location.pathname}#/finance-center/order-list`)
}

function handleBack() {
  replaceToWorkbench()
}

function handleConfirm() {
  router.push('/import-center/import-order-starter')
}
</script>

<template>
  <div class="page">
    <div class="confirm-school-box">
      <div class="import-logo" />
      <div class="import-title">
        你即将进入「学员订单导入」流程
      </div>
      <div class="school-name">
        「{{ schoolName }}」
      </div>
      <div class="import-btn">
        <a-button type="primary" class="w-full h-40px bg-#06f font500 text-16px" @click="handleConfirm">
          确认机构并进入
        </a-button>
      </div>
      <div class="back">
        如需切换机构请 <span @click="handleBack">返回工作台</span>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.page {
  width: 100%;
  height: 100%;
  background-color: #f7f7fd;
  background-image: url("https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12488/static/import-starter-bg.909674eb.png");
  background-size: contain;
  background-position: 100%;
  background-repeat: no-repeat;
  display: flex;
  justify-content: center;
  align-items: center;

  .confirm-school-box {
    box-sizing: border-box;
    padding: 89px 40px 40px;
    width: 440px;
    height: 560px;
    background-color: #fff;
    border-radius: 24px;
    box-shadow: 0 0 32px 0 rgba(0, 0, 0, 0.08);
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .import-logo {
    width: 236px;
    height: 115px;
    background-image: url("https://pcsys.admin.ybc365.com/eb0adc86-db89-4ecb-9e15-c96cde83ab24.png");
    background-size: contain;
    background-position: 50%;
    background-repeat: no-repeat;
  }

  .import-title {
    padding-top: 48px;
    font-size: 22px;
    font-weight: 600;
    color: #222;
    line-height: 40px;
  }

  .school-name {
    padding-top: 8px;
    font-size: 22px;
    font-weight: 600;
    color: #06f;
    line-height: 40px;
    text-align: center;
  }

  .import-btn {
    padding-top: 48px;
    width: 300px;
  }

  .back {
    padding-top: 24px;
    font-size: 16px;
    color: #999;
    line-height: 22px;

    span {
      color: #06f;
      cursor: pointer;
    }
  }
}
</style>
