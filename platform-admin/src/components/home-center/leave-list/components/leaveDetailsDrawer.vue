<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'

const open = defineModel({
  type: Boolean,
  default: false,
})

// 这里可以添加请假详情的数据，实际使用时可能需要从props或API获取
const leaveInfo = reactive({
  studentName: '浩楠',
  studentPhone: '159****0001',
  startTime: '2025-05-20 12:42',
  endTime: '2025-05-21 23:59',
  leaveTitle: '请假课节 (1)',
  leaveClasses: [
    {
      time: '11:00 ~ 11:45',
      className: '浩楠-高级音语课',
      courseName: '高级音语课',
      teacherName: '何红武',
    },
  ],
  leaveProcess: [
    {
      name: '何红武 (代办)',
      status: '发起',
      time: '2025-05-20 10:43',
    },
    {
      name: '系统自动执行',
      status: '已通过',
      time: '2025-05-20 10:43',
    },
  ],
})
</script>

<template>
  <a-drawer
    v-model:open="open" :body-style="{ padding: '0', background: '#fff' }" :keyboard="false"
    :mask-closable="false" :closable="false" width="800px"
  >
    <!-- 自定义头部 -->
    <template #title>
      <div class="custom-header flex justify-between h-4 flex-items-center">
        <div class="text-5">
          请假详情
        </div>
        <a-button type="text" class="close-btn" @click="open = false">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <!-- 自定义内容 -->
    <div class="custom-content py-20px px-30px">
      <!-- 学生信息区域 -->
      <div class="student-info flex items-center justify-between mb-10px">
        <div class="flex items-center">
          <div class="avatar-container mr-15px">
            <div class="avatarBox w-16 h-16 relative">
              <img
                width="64" height="64" class=" rounded-100"
                src="https://cdn.schoolpal.cn/schoolpal/next-erp/avator_female.png?x-oss-process=image/resize,w_192"
                alt=""
              >
              <svg width="24px" height="24px" viewBox="0 0 24 24" style="position: absolute; bottom: 0px; right: 0px;">
                <g id="\u9875\u9762-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                  <g id="\u753B\u677F\u5907\u4EFD-24" transform="translate(-1529.000000, -266.000000)">
                    <g id="\u7F16\u7EC4" transform="translate(1529.000000, 266.000000)">
                      <g id="\u7537" fill="#FF3333">
                        <circle id="\u692D\u5706\u5F62" cx="12" cy="12" r="12" />
                      </g>
                      <g id="\u65B0/\u7B26\u53F7/\u5973\u751F" transform="translate(4.000000, 3.722662)" fill="#FFFFFF">
                        <path
                          id="\u5F62\u72B6\u7ED3\u5408"
                          d="M14.0392198,1.98464807 C14.3104172,2.25584549 14.3104172,2.6955434 14.0392198,2.96674082 L12.7581282,4.24782197 L13.8356011,5.32529291 C14.1067985,5.59649034 14.1067985,6.03618824 13.8356011,6.30738566 C13.5644037,6.57858309 13.1247058,6.57858309 12.8535084,6.30738566 L11.7760417,5.22990846 L11.0422043,5.9637881 C11.6729795,6.80471314 12.0529934,7.84453862 12.0712783,8.97229681 L12.0719697,9.05764313 C12.0719697,11.9079764 9.76131807,14.218628 6.91098485,14.218628 C4.06065162,14.218628 1.75,11.9079764 1.75,9.05764313 C1.75,6.20730991 4.06065162,3.89665829 6.91098485,3.89665829 C8.09991052,3.89665829 9.19493833,4.29868287 10.067543,4.9742068 L10.7939552,4.24782197 L9.9295504,3.3834277 C9.65835297,3.11223027 9.65835297,2.67253237 9.9295504,2.40133495 C10.2007478,2.13013752 10.6404457,2.13013752 10.9116432,2.40133495 L11.7760417,3.26573548 L13.0571271,1.98464807 C13.3283245,1.71345064 13.7680224,1.71345064 14.0392198,1.98464807 Z M6.91098485,5.28554717 C4.82771378,5.28554717 3.13888889,6.97437206 3.13888889,9.05764313 C3.13888889,11.1409142 4.82771378,12.8297391 6.91098485,12.8297391 C8.97342321,12.8297391 10.6492649,11.1745218 10.6825755,9.12002164 L10.6830492,9.06119427 L10.6824968,8.99148152 L10.6821047,8.97066975 C10.664377,8.18243978 10.4051064,7.43717494 9.94528363,6.81616198 L9.93114482,6.79719028 L9.62321654,6.38666775 L9.21733473,6.07245915 C8.5679879,5.56976952 7.7753284,5.29115293 6.93642036,5.28563087 L6.91098485,5.28554717 Z"
                          transform="translate(7.996309, 7.999939) scale(-1, -1) translate(-7.996309, -7.999939) "
                        />
                      </g>
                    </g>
                  </g>
                </g>
              </svg>
            </div>
          </div>
          <div class="info-container">
            <div class="name text-20px font-medium mb-5px">
              {{ leaveInfo.studentName }}
            </div>
            <div class="phone text-16px font-450">
              {{ leaveInfo.studentPhone }}
            </div>
          </div>
        </div>
        <div>
          <img
            width="96" height="96" src="https://pcsys.admin.ybc365.com/5efd9ae5-3728-47df-86bc-0f6493f3bfaa.png"
            alt=""
          >
        </div>
      </div>

      <!-- 请假时间信息 -->
      <div class="time-info flex items-center gap-25% mb-25px">
        <div class="flex items-center gap-1 text-14px">
          <div class="label ">
            开始时间：
          </div>
          <div class="value text-#999">
            {{ leaveInfo.startTime }}
          </div>
        </div>
        <div class="flex items-center gap-1 .dark:">
          <div class="label">
            结束时间：
          </div>
          <div class="value text-#999">
            {{ leaveInfo.endTime }}
          </div>
        </div>
      </div>

      <!-- 请假课节 -->
      <div class=" mb-20px">
        <div class="text-20px font-700 mb-15px">
          {{ leaveInfo.leaveTitle }}
        </div>

        <div class="bg-#fafafa rounded-8px p-15px">
          <div v-for="(item, index) in leaveInfo.leaveClasses" :key="index" class="class-item">
            <div class="time-label flex items-center mb-10px">
              <div class="dot w-6px h-6px rounded-full bg-#0066ff mr-8px" />
              <div class="time text-15px font-medium">
                {{ item.time }}
              </div>
              <div class="class-name text-14px font-medium ml-10px">
                {{ item.className }}
              </div>
            </div>

            <div class="pl-2px">
              <div class="px-10px border-0px border-l-1px border-l-#e5e6eb border-solid">
                <div class="course mb-8px text-#999">
                  <span class="label text-14px ">上课课程：</span>
                  <span class="value text-14px ml-5px">{{ item.courseName }}</span>
                </div>
                <div class="text-#999">
                  <span class="label text-14px">上课教师：</span>
                  <span class="value text-14px ml-5px">{{ item.teacherName }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 请假流程 -->
      <div class="">
        <div class="text-20px font-700 mb-15px">
          请假流程
        </div>

        <div class="process-timeline">
          <div
            v-for="(item, index) in leaveInfo.leaveProcess" :key="index" class="process-item flex relative pb-20px"
            :class="{ 'last-item': index === leaveInfo.leaveProcess.length - 1 }"
          >
            <div class="flex items-center gap-20px flex-1">
              <div
                :class="{ 'bg-#0066ff line': index === 0, 'bg-#00cc66': index !== 0 }"
                class=" p-8px text-12px text-#fff  rounded-50% w-28px h-28px flex items-center justify-center"
              >
                {{ index === 0 ? '发' : '批' }}
              </div>
              <div class="flex items-center justify-between bg-#f6f6f6 rounded-8px p-15px py-20px flex-1">
                <div class="flex items-center gap-20px">
                  <span class="text-16px font-500">{{ item.name }}</span>
                  <div class="text-12px font-400 py-3px px-8px rounded-8px " :class="index === 0 ? 'info' : 'success'">
                    {{ item.status }}
                  </div>
                </div>
                <div class="text-#999">
                  {{ index === 0 ? '发起时间' : '处理时间' }}：{{ item.time }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </a-drawer>
</template>

<style lang="less" scoped>
.close-btn {
  &:hover {
    background: transparent;
  }
}

.custom-content {
  height: calc(100vh - 56px);
  overflow-y: auto;
}

.process-item {
  &.last-item {
    padding-bottom: 0;
  }
}

.info {
  color: #0066ff;
  background-color: #e6f0ff;
}

.success {
  color: #00cc66;
  background-color: #e6ffed;
}

.line {
  position: relative;
  &::before {
    content: '';
    display: block;
    position: absolute;
    max-height: 80px;
    height: 40px;
    width: 1px;
    background-color: #ccc;
    // bottom: 0;
    top:33px;
  }
}
</style>
