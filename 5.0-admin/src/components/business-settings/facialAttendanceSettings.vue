<script setup>
import { QuestionCircleOutlined } from '@ant-design/icons-vue'

const toggleValue1 = ref(false)
const toggleValue2 = ref(false)
const toggleValue3 = ref(false)
const toggleValue4 = ref(false)
const toggleValue5 = ref(false)
</script>

<template>
  <div class="registration-settings">
    <!-- 人脸考勤关联点名课消 -->
    <div
      class="mt-12px bg-white px-16px py-10px  border-t-0 border-b-1px border-l-0 border-r-0 border-#eee border-solid
      flex justify-between items-center"
    >
      <div>
        <div class="text-16px">
          人脸考勤关联点名课消
        </div>
        <div class="text-#999 text-12px mt-2px">
          开启后，可设置每日考勤时段及人脸考勤课消规则
        </div>
      </div>
      <div class="item-status ml-20px">
        <a-switch v-model:checked="toggleValue1" @click.stop />
      </div>
    </div>
    <!-- 每日考勤时段 -->
    <div
      class="bg-white px-16px py-10px flex justify-between items-center border-t-0 border-b-1px border-l-0 border-r-0 border-#eee border-solid"
    >
      <div>
        <div class="flex items-center gap-2">
          <span class="text-16px">每日考勤时段</span>
          <span>
            <QuestionCircleOutlined class="text-14px text-#06f cursor-pointer" />
          </span>
        </div>
        <div class="text-#999 text-12px">
          00:00 ~ 24:00
        </div>
      </div>
      <div class="item-arrow">
        <span class="text-#06f text-13px cursor-pointer">设置</span>
      </div>
    </div>
    <!--  人脸考勤课消规则  -->
    <div class="bg-white px-16px py-10px flex justify-between items-center ">
      <div style="flex: 4;">
        <div class="flex items-center gap-2">
          <span class="text-16px"> 人脸考勤课消规则 </span>
          <span class="text-#06f">按序课消</span>
        </div>
        <div class="text-#999 text-12px ">
          同一考勤时段内，若只有一个日程将自动课消，若有多个日程将按选定规则执行
        </div>
      </div>
      <div class="item-arrow text-right" style="flex: 1;">
        <span class="text-#06f text-13px cursor-pointer">切换</span>
      </div>
    </div>
    <!--  学员考勤签到通知  -->
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          学员考勤签到通知
        </div>
        <div class="item-description line-height-16px mt-4px">
          开启后，学员完成考勤签到或手动标记签到将会自动发送微信通知给家长
        </div>
      </div>
      <div class="item-status ml-20px">
        <a-switch v-model:checked="toggleValue2" @click.stop />
      </div>
    </div>
    <!--  学员考勤签退通知  -->
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          学员考勤签退通知
        </div>
        <div class="item-description line-height-16px mt-4px">
          开启后，学员完成考勤签退或手动标记签退将会自动发送微信通知给家长
        </div>
      </div>
      <div class="item-status ml-20px">
        <a-switch v-model:checked="toggleValue3" @click.stop />
      </div>
    </div>
    <!--  语音提示  -->
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          语音提示
        </div>
        <div class="item-description line-height-16px mt-4px">
          开启后，系统将实时播报学员是否人脸考勤成功、签到签退成功
        </div>
      </div>
      <div class="item-status ml-20px">
        <a-switch v-model:checked="toggleValue4" @click.stop />
      </div>
    </div>
    <!--  刷脸间隔时间  -->
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          刷脸间隔 <span class="text-#06f">10</span> 分钟
        </div>
        <div class="item-description line-height-16px mt-4px">
          学员刷脸考勤成功后，允许下次刷脸考勤间隔时长
        </div>
      </div>
      <div class="item-status ml-20px">
        <span class="text-#06f text-13px cursor-pointer">编辑</span>
      </div>
    </div>
    <!--  学员刷脸成功提醒至管理员  -->
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          学员刷脸成功提醒至管理员
        </div>
        <div class="item-description line-height-16px mt-4px">
          开启后，机构管理员将在App 端消息中心收到学员刷脸成功的推送提醒
        </div>
      </div>
      <div class="item-status ml-20px">
        <a-switch v-model:checked="toggleValue5" @click.stop />
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
  .registration-settings {
    padding: 12px 0px;
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
