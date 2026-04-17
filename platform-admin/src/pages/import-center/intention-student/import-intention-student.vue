<script setup>
import { useRouter } from 'vue-router'
import { useUserStore } from '~@/stores/user'
import { computed } from 'vue'

const router = useRouter()
const userStore = useUserStore()

// 从 Pinia store 中获取机构名称
const schoolName = computed(() => userStore.userInfo?.orgName || '总机构')

function replaceToWorkbench() {
  window.location.replace(`${window.location.origin}${window.location.pathname}#/enroll-center/intention-student`)
}

function handleBack() {
  replaceToWorkbench()
}
function handleConfirm() {
  router.push('/import-center/import-intention-student-starter')
}
</script>
<template>
  <div class="page">
    <div class="confirm-school-box">
      <div class="import-logo"></div>
      <div class="import-title">
        你即将进入「意向学员导入」流程
      </div>
      <div class="school-name">
        「{{ schoolName }}」
      </div>
      <div class="import-btn">
        <a-button type="primary" @click="handleConfirm" class="w-full h-40px bg-#06f font500 text-16px">
          确认机构并进入
        </a-button>
      </div>
      <div class="back">
        如需调整请 <span @click="handleBack">返回工作台</span>
      </div>
    </div>
  </div>
</template>



<style lang="less" scoped>
.page {
  width: 100%;
  height: 100%;
  background-color: #F7f7fd;
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
    box-shadow: 0 0 32px 0 rgba(0, 0, 0, .08);
    display: flex;
    flex-direction: column;
    align-items: center;

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
}
</style>
