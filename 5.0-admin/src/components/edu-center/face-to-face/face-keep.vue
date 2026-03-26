<script setup>
import { InfoCircleFilled, RightOutlined } from '@ant-design/icons-vue'

const router = useRouter()
// 在组件作用域外定义窗口引用（避免被Vue响应式代理）
let faceWindow = null
function handleFaceSign(type) {
  const newUrl = router.resolve({
    path: '/pc/face',
    query: { type }, // 确保使用最新的 type
  }).href

  // 判断窗口是否存在且未关闭
  if (faceWindow && !faceWindow.closed) {
    // 比较当前 URL 与新 URL
    if (faceWindow.location.href !== newUrl) {
      // URL 不同时先关闭旧窗口
      faceWindow.close()
      faceWindow = window.open(newUrl, '_blank')
    }
    else {
      // 相同则直接聚焦
      faceWindow.focus()
    }
  }
  else {
    // 打开新窗口并保存引用
    faceWindow = window.open(newUrl, '_blank')
  }

  // 处理可能被拦截的情况
  if (!faceWindow || faceWindow.closed) {
    faceWindow = null
    // 可选：添加重试逻辑或用户提示
  }
}
</script>

<template>
  <div class="bg-white  h-80vh flex justify-center flex-wrap items-center ">
    <div class="flex">
      <div
        class="h-56 mr-6 hover-shadow w-55 flex  items-center  flex-col border-1 border-color-#0000000d border-solid rounded-2 cursor-pointer"
        @click="handleFaceSign(1)"
      >
        <img
          width="120" height="120" class="mb-2"
          src="https://pcsys.admin.ybc365.com//e7cb1394-1c75-47ec-b37c-ad95f6863504.png" alt=""
        >
        <div class="text-4 text-#222 font-500 mb-1">
          人脸采集
        </div>
        <div class="text-3 text-#222">
          先采集，再考勤
        </div>
      </div>
      <div
        class="h-56 mr-6 hover-shadow w-55 flex items-center flex-col border-1 border-color-#0000000d border-solid rounded-2 cursor-pointer"
        @click="handleFaceSign(2)"
      >
        <img
          width="120" height="120" class="mb-2"
          src="https://pcsys.admin.ybc365.com//1300b671-9022-4b3f-9cc6-a1deec75d52e.png" alt=""
        >
        <div class="text-4 text-#222 font-500 mb-1">
          人脸考勤
        </div>
        <div class="text-3 text-#222">
          识别自动记录，支持多人识别
        </div>
      </div>
      <div
        class="h-56 mr-6 hover-shadow w-55 flex-center flex-col border-1 border-color-#0000000d border-solid rounded-2 cursor-pointer"
      >
        <img
          width="73" height="73" class="mb-2"
          src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/more-dian-icon.4505caef.png"
          alt=""
        >
        <div class="text-4 text-#222 font-500 mb-1">
          人脸考勤设备
        </div>
        <div class="text-3 text-#06f">
          未激活设备，点此激活 >
        </div>
      </div>
    </div>
    <div class="w-88 h-83 total ">
      <div class="t flex justify-between">
        <span>今日考勤统计</span>
        <InfoCircleFilled class="text-#c5cee0 cursor-pointer hover-text-#06f " />
      </div>
      <div class="mt-8">
        <div class="mb-3.5 bg-white rounded-2 py-4 px-3 flex justify-between flex-center cursor-pointer items">
          <span>待考勤</span>
          <span class="num flex flex-items-center  ">2
            <RightOutlined class="text-3 text-#ccc ml-1" />
          </span>
        </div>
        <div class="mb-3.5 bg-white rounded-2 py-4 px-3 flex justify-between flex-center cursor-pointer items">
          <span>考勤成功</span>
          <span class="num flex flex-items-center  ">0
            <RightOutlined class="text-3 text-#ccc ml-1" />
          </span>
        </div>
        <div class=" bg-white rounded-2 py-4 px-3 flex justify-between flex-center cursor-pointer items">
          <span>考勤成功未点名</span>
          <span class="num flex flex-items-center  ">0
            <RightOutlined class="text-3 text-#ccc ml-1 icon" />
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.total {
  background: url('https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12087/static/today-data.82d0298c.png');
  background-size: contain;
  padding: 24px 48px 24px;
  border-radius: 16px;

  .t {
    font-family: PingFangSC-Medium, PingFang SC, sans-serif;
    font-size: 18px;
    font-weight: 500;
    color: #222;
  }

  .num {
    font-family: DINAlternate-Bold, DINAlternate, sans-serif;
    font-size: 20px;
    font-weight: bold;
  }

  .items {
    &:hover {
      .num {
        color: var(--pro-ant-color-primary);

        :deep(svg) {
          color: var(--pro-ant-color-primary) !important;
        }
      }
    }
  }
}
</style>
