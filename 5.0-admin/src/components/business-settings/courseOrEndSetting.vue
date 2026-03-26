<script setup>
import { QuestionCircleOutlined, RightOutlined } from '@ant-design/icons-vue'

const activeKey = ref('')
const activeKey2 = ref('')
const activeKey3 = ref('')
const activeKey4 = ref('')
const toggleValue = ref(false)
const toggleValue2 = ref(false)
const toggleValue3 = ref(true)
const toggleValue4 = ref(false)
const toggleValue5 = ref(false)
const toggleValue6 = ref(true)
const toggleValue7 = ref(true)
const toggleValue8 = ref(true)
const toggleValue9 = ref(true)
const toggleValue10 = ref(true)
</script>

<template>
  <div class="business-settings-course-or-end-setting">
    <div class="flex justify-between items-center px-16px">
      <span class="text-#666 text-12px">授课方式</span>
      <span>
        <QuestionCircleOutlined class="text-#06f cursor-pointer" />
      </span>
    </div>
    <div
      class="mt-12px bg-white px-16px py-10px  border-t-0 border-b-1px border-l-0 border-r-0 border-#eee border-solid">
      <div class="flex justify-between items-center">
        <span class="text-16px">班级授课</span>
        <span class="text-#ccc text-13px">已开启</span>
      </div>
      <div class="text-#999 text-12px mt-2px">
        可自定义学员数量，允许上课老师给多名学员上课
      </div>
    </div>
    <div class=" bg-white px-16px py-10px">
      <div class="flex justify-between items-center">
        <span class="text-16px">1v1授课</span>
        <span class="text-#ccc text-13px">已开启</span>
      </div>
      <div class="text-#999 text-12px mt-2px">
        对学员数量有限制，只允许上课老师给1名学员上课
      </div>
    </div>
    <!-- 收费方式 -->
    <div class="flex justify-between items-center px-16px pt-10px">
      <span class="text-#666 text-12px">收费方式</span>
      <span>
        <QuestionCircleOutlined class="text-#06f cursor-pointer" />
      </span>
    </div>
    <!-- 按课时收费 -->
    <div
      class="mt-10px bg-white px-16px py-10px  border-t-0 border-b-1px border-l-0 border-r-0 border-#eee border-solid">
      <div class="flex justify-between items-center">
        <span class="text-16px">按课时收费</span>
        <span class="text-#ccc text-13px">已开启</span>
      </div>
      <div class="text-#999 text-12px mt-2px">
        按“课时购买数”定价，以“课时”为单位计费
      </div>
    </div>
    <div class="flex justify-between items-center  py-10px bg-white">
      <a-collapse v-model:active-key="activeKey" ghost class="flex-1">
        <a-collapse-panel key="1">
          <template #header>
            <span class="text-12px text-#666 px-16px">关联项设置</span>
          </template>
          <div
            class=" bg-white  py-12px mt-10px border-t-1px border-b-0 border-l-0 border-r-0 border-#eee border-solid">
            <div class="flex justify-between items-center px-16px">
              <span class="text-14px">每次上课，默认记录学员 <span class="text-#06f">1</span> 课时，教师 <span
                  class="text-#06f">0</span> 课时
              </span>
              <span class="text-#06f text-13px cursor-pointer">编辑</span>
            </div>
            <div class="text-#999 text-12px mt-2px px-16px">
              创建班级/一对一时仍然可以进行编辑调整
            </div>
          </div>
          <div class="settings-item">
            <div class="item-content">
              <div class="item-title">
                学员请假正常记录
              </div>
              <div class="item-description">
                开启后，学员请假正常记录课时
              </div>
            </div>
            <div class="item-status">
              <a-switch v-model:checked="toggleValue" @click.stop />
            </div>
          </div>
          <div class="settings-item">
            <div class="item-content">
              <div class="item-title">
                学员旷课正常记录
              </div>
              <div class="item-description">
                开启后，学员旷课正常记录课时
              </div>
            </div>
            <div class="item-status">
              <a-switch v-model:checked="toggleValue2" @click.stop />
            </div>
          </div>
        </a-collapse-panel>
      </a-collapse>
    </div>
    <!-- 按时段收费 -->
    <div
      class="mt-10px bg-white px-16px py-10px  border-t-0 border-b-1px border-l-0 border-r-0 border-#eee border-solid">
      <div class="flex justify-between items-center">
        <span class="text-16px">按时段收费</span>
        <span class="text-#ccc text-13px">已开启</span>
      </div>
      <div class="text-#999 text-12px mt-2px">
        按“天/自然月/自然年”定价，以“天”为单位计费
      </div>
    </div>
    <div class="flex justify-between items-center  py-10px bg-white">
      <a-collapse v-model:active-key="activeKey2" ghost class="flex-1">
        <a-collapse-panel key="1">
          <template #header>
            <span class="text-12px text-#666 px-16px">关联项设置</span>
          </template>
          <div class="settings-item mt-10px">
            <div class="item-content">
              <div class="item-title">
                学员补课
              </div>
              <div class="item-description">
                开启后，按时段收费的课程，支持缺课学员补课
              </div>
            </div>
            <div class="item-status">
              <a-switch v-model:checked="toggleValue3" @click.stop />
            </div>
          </div>
        </a-collapse-panel>
      </a-collapse>
    </div>
    <!-- 按金额收费 -->
    <div
      class="mt-10px bg-white px-16px py-10px  border-t-0 border-b-1px border-l-0 border-r-0 border-#eee border-solid">
      <div class="flex justify-between items-center">
        <span class="text-16px">按金额收费</span>
        <span class="text-#ccc text-13px">已开启</span>
      </div>
      <div class="text-#999 text-12px mt-2px">
        按“充值金额”定价，以“课次”为单位计费
      </div>
    </div>
    <div class="flex justify-between items-center  py-10px bg-white">
      <a-collapse v-model:active-key="activeKey3" ghost class="flex-1">
        <a-collapse-panel key="1">
          <template #header>
            <span class="text-12px text-#666 px-16px">关联项设置</span>
          </template>
          <div
            class=" bg-white  py-12px mt-10px border-t-1px border-b-0 border-l-0 border-r-0 border-#eee border-solid">
            <div class="flex justify-between items-center px-16px">
              <span class="text-14px">扣费规则</span>
              <span class="text-#06f text-13px cursor-pointer">编辑</span>
            </div>
            <div class="text-#333 text-12px mt-2px px-16px">
              <div>默认扣费：<span class="text-#06f">100</span> 元 <span class="text-#999">（仅对未设置单课扣费的课程有效）</span></div>
              <div>单课扣费：已设置 <span class="text-#06f">0</span> 门课</div>
            </div>
          </div>
          <div class="settings-item">
            <div class="item-content">
              <div class="item-title">
                学员请假正常记录
              </div>
              <div class="item-description">
                开启后，学员请假正常记录课时
              </div>
            </div>
            <div class="item-status">
              <a-switch v-model:checked="toggleValue4" @click.stop />
            </div>
          </div>
          <div class="settings-item">
            <div class="item-content">
              <div class="item-title">
                学员旷课正常记录
              </div>
              <div class="item-description">
                开启后，学员旷课正常记录课时
              </div>
            </div>
            <div class="item-status">
              <a-switch v-model:checked="toggleValue5" @click.stop />
            </div>
          </div>
          <div class="settings-item mt-10px">
            <div class="item-content">
              <div class="item-title">
                学员补课
              </div>
              <div class="item-description">
                开启后，按金额收费的课程，支持缺课学员补课
              </div>
            </div>
            <div class="item-status">
              <a-switch v-model:checked="toggleValue6" @click.stop />
            </div>
          </div>
        </a-collapse-panel>
      </a-collapse>
    </div>
    <div class=" bg-white  py-12px mt-10px border-t-1px border-b-0 border-l-0 border-r-0 border-#eee border-solid">
      <div class="flex justify-between items-center px-16px">
        <span class="text-16px">课消顺序设置</span>
        <span class="text-#06f text-13px cursor-pointer">编辑</span>
      </div>
      <div class="text-#666 text-12px mt-2px px-16px">
        当前：先进先出
      </div>
    </div>
    <!-- 通用课 -->
    <div class="flex justify-between items-center px-16px mt-12px">
      <span class="text-#666 text-12px">通用课</span>
    </div>
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          通用课
        </div>
        <div class="item-description line-height-16px mt-4px">
          开启后机构可开设通用课程，允许报读学员在设置的课程范围内上多门课
        </div>
      </div>
      <div class="item-status ml-20px">
        <a-switch v-model:checked="toggleValue7" @click.stop />
      </div>
    </div>
    <!-- 一班多课 -->
    <div class="flex justify-between items-center px-16px mt-12px">
      <span class="text-#666 text-12px">一班多课</span>
    </div>
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          一班多课
        </div>
        <div class="item-description line-height-16px mt-4px">
          开启后，机构可设置组合课程，创建班级时支持关联组合课程，实现报名不同课程的学员在同一个班级中上课
        </div>
      </div>
      <div class="item-status">
        <a-switch v-model:checked="toggleValue8" @click.stop />
      </div>
    </div>
    <!-- 查看组合课程列表 -->
    <div class="settings-item bg-white">
      <div class="item-content">
        <div class="item-title">
          查看组合课程列表
        </div>
      </div>
      <div class="item-arrow">
        <RightOutlined />
      </div>
    </div>
    <!-- 节假日设置 -->
    <div class="flex justify-between items-center px-16px mt-12px">
      <span class="text-#666 text-12px">节假日设置</span>
    </div>
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          过滤节假日
        </div>
        <div class="item-description line-height-16px mt-4px">
          开启后，可在排课时选择过滤节假日，系统会根据节假日设置计算日程排布
        </div>
      </div>
      <div class="item-status ml-12px">
        <a-switch v-model:checked="toggleValue9" @click.stop />
      </div>
    </div>
    <!-- 上课时段设置 -->
    <div class="flex justify-between items-center px-16px mt-12px">
      <span class="text-#666 text-12px">上课时段设置</span>
    </div>
    <div class="bg-white settings-item mt-10px">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          快捷选择学校统一时段
        </div>
        <div class="item-description ">
          开启后，可在排课时快捷选择学校统一上课时段
        </div>
      </div>
      <div class="item-status ml-12px">
        <a-switch v-model:checked="toggleValue10" @click.stop />
      </div>
    </div>
    <!-- 学校统一上课时间段 -->
    <div class="settings-item bg-white">
      <div class="item-content">
        <div class="item-title" style="font-size: 16px;">
          学校统一上课时间段
        </div>
      </div>
      <div class="item-arrow">
        <span class="text-#06f text-13px cursor-pointer">编辑</span>
      </div>
    </div>
    <!-- 时段列表 -->
    <a-collapse v-model:active-key="activeKey4" ghost
      class="schedule-collapse bg-white border-t-1px border-b-0 border-l-0 border-r-0 border-#eee border-solid">
      <a-collapse-panel key="1" class="schedule-panel">
        <template #header>
          <div class="schedule-header">
            <div class="schedule-badge schedule-badge-a">A</div>
            <div class="schedule-info">
              <span class="schedule-title">A时段</span>
              <span class="schedule-count">共12节课</span>
            </div>
          </div>
        </template>
        <div class="schedule-content">
          <div class="schedule-grid">
            <div class="schedule-item" v-for="(item, index) in 12" :key="index">
              <div class="schedule-number">{{ index + 1 }}</div>
              <div class="schedule-time">
                {{ String(8 + index).padStart(2, '0') }}:00-{{ String(9 + index).padStart(2, '0') }}:00
              </div>
              <div class="schedule-indicator"></div>
            </div>
          </div>
        </div>
      </a-collapse-panel>
      <a-collapse-panel key="2" class="schedule-panel">
        <template #header>
          <div class="schedule-header">
            <div class="schedule-badge schedule-badge-b">B</div>
            <div class="schedule-info">
              <span class="schedule-title">B时段</span>
              <span class="schedule-count">共12节课</span>
            </div>
          </div>
        </template>
        <div class="schedule-content">
          <div class="schedule-grid">
            <div class="schedule-item" v-for="(item, index) in 12" :key="index">
              <div class="schedule-number">{{ index + 1 }}</div>
              <div class="schedule-time">
                {{ String(8 + index).padStart(2, '0') }}:00-{{ String(9 + index).padStart(2, '0') }}:00
              </div>
              <div class="schedule-indicator"></div>
            </div>
          </div>
        </div>
      </a-collapse-panel>
    </a-collapse>
  </div>
</template>

<style lang="less" scoped>
.business-settings-course-or-end-setting {
  padding: 12px 0px;
  background-color: #f6f7f8;

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

  .schedule-collapse {
    margin-top: 10px;
    background-color: #fff;
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
    border: 1px solid #e8e9ea;

    :deep(.ant-collapse-header) {
      padding: 0 !important;
      background: linear-gradient(135deg, #f8f9fa 0%, #f1f3f4 100%);
      border-bottom: 1px solid #e8e9ea;
      transition: all 0.3s ease;
    }

    :deep(.ant-collapse-header:hover) {
      background: linear-gradient(135deg, #e6f7ff 0%, #d6e7f0 100%);
    }

    :deep(.ant-collapse-expand-icon) {
      position: absolute;
      right: 16px;
      top: 50%;
      transform: translateY(-50%);
      padding-inline-end: 0px !important;
      color: #666;
      transition: color 0.3s ease;
    }

    :deep(.ant-collapse-content-box) {
      padding: 0px !important;
      background-color: #fafafa;
    }

    .schedule-panel {
      border-bottom: 1px solid #e8e9ea;

      &:last-child {
        border-bottom: none;
      }
    }

    .schedule-header {
      display: flex;
      align-items: center;
      padding: 16px 20px;
      cursor: pointer;
      transition: all 0.3s ease;
    }

    .schedule-badge {
      width: 32px;
      height: 32px;
      line-height: 32px;
      text-align: center;
      border-radius: 50%;
      font-size: 14px;
      font-weight: bold;
      color: #fff;
      margin-right: 16px;
      flex-shrink: 0;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
    }

    .schedule-badge-a {
      background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%);
    }

    .schedule-badge-b {
      background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
    }

    .schedule-info {
      display: flex;
      flex-direction: column;
      gap: 2px;
    }

    .schedule-title {
      font-size: 16px;
      color: #262626;
      font-weight: 600;
      line-height: 1.4;
    }

    .schedule-count {
      font-size: 12px;
      color: #8c8c8c;
      font-weight: normal;
    }

    .schedule-content {
      padding: 20px;
      background-color: #fff;
    }

    .schedule-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
      gap: 12px;
    }

    .schedule-item {
      display: flex;
      align-items: center;
      padding: 12px 16px;
      background: linear-gradient(135deg, #f8f9fa 0%, #ffffff 100%);
      border: 1px solid #e8e9ea;
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;

      &:hover {
        background: linear-gradient(135deg, #e6f7ff 0%, #f0f9ff 100%);
        border-color: #91d5ff;
        transform: translateY(-2px);
        box-shadow: 0 6px 16px rgba(24, 144, 255, 0.15);
      }

      &:before {
        content: '';
        position: absolute;
        left: 0;
        top: 0;
        bottom: 0;
        width: 4px;
        background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%);
        opacity: 0;
        transition: opacity 0.3s ease;
      }

      &:hover:before {
        opacity: 1;
      }
    }

    .schedule-number {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: linear-gradient(135deg, #1890ff 0%, #40a9ff 100%);
      color: #fff;
      border-radius: 50%;
      font-size: 12px;
      font-weight: bold;
      margin-right: 12px;
      flex-shrink: 0;
      box-shadow: 0 2px 6px rgba(24, 144, 255, 0.3);
    }

    .schedule-time {
      font-size: 14px;
      color: #595959;
      font-weight: 500;
      margin-right: auto;
      font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
    }

    .schedule-indicator {
      width: 10px;
      height: 10px;
      background: linear-gradient(135deg, #52c41a 0%, #73d13d 100%);
      border-radius: 50%;
      flex-shrink: 0;
      box-shadow: 0 0 0 3px rgba(82, 196, 26, 0.2);
      animation: pulse 2s infinite;
    }

    @keyframes pulse {
      0% {
        box-shadow: 0 0 0 0 rgba(82, 196, 26, 0.4);
      }

      70% {
        box-shadow: 0 0 0 6px rgba(82, 196, 26, 0);
      }

      100% {
        box-shadow: 0 0 0 0 rgba(82, 196, 26, 0);
      }
    }
  }
}
</style>
