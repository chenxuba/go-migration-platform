<script setup>
import { CloseOutlined, DownOutlined, ExclamationCircleFilled, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import { onMounted, onUnmounted } from 'vue'
import EditClassInfoModal from './edit-class-info-modal.vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open'])
// 处理双向绑定
const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
// 数据源改为ref，以便响应式更新
const data = ref([
  {
    id: '1',
    type: '1', // 上课学员
    studentAccount: '张三',
    attended: true,
    absent: false,
    leave: false,
    unrecorded: false,
    consumptionMethod: '1', // 按课时
    recordAttendance: true,
    attendanceCount: 1,
    internalNote: '',
    externalNote: '',
  },
  {
    id: '2',
    type: '2', // 临时学员
    studentAccount: '李四',
    attended: true,
    absent: false,
    leave: false,
    unrecorded: false,
    consumptionMethod: '2', // 按时段
    recordAttendance: false,
    attendanceCount: 1,
    internalNote: '',
    externalNote: '',
  },
  {
    id: '3',
    type: '3', // 试听学员
    studentAccount: '王五',
    attended: true,
    absent: false,
    leave: false,
    unrecorded: false,
    consumptionMethod: '3', // 按金额
    recordAttendance: false,
    attendanceCount: 1,
    internalNote: '',
    externalNote: '',
  },
  {
    id: '4',
    type: '4', // 补课学员
    studentAccount: '赵六',
    attended: true,
    absent: false,
    leave: false,
    unrecorded: false,
    consumptionMethod: '1', // 按课时
    recordAttendance: false,
    attendanceCount: 1,
    internalNote: '',
    externalNote: '',
  },

])
// 定义列
const columns = ref(
  [
    {
      title: '',
      dataIndex: 'index',
      fixed: 'left',
      width: 30,
      key: 'index',
    },
    {
      title: '学员/扣费课程账户',
      dataIndex: 'studentAccount',
      fixed: 'left',
      width: 220,
      key: 'studentAccount',
    },
    {
      title: '到课',
      dataIndex: 'attended',
      width: 95,
      key: 'attended',
    },
    {
      title: '旷课',
      dataIndex: 'absent',
      width: 95,
      key: 'absent',
    },
    {
      title: '请假',
      dataIndex: 'leave',
      width: 95,
      key: 'leave',
    },
    {
      title: '未记录',
      dataIndex: 'unrecorded',
      width: 130,
      key: 'unrecorded',
    },
    {
      title: '课消方式',
      dataIndex: 'consumptionMethod',
      width: 120,
      key: 'consumptionMethod',
    },
    {
      title: '上课点名数量',
      dataIndex: 'attendanceCount',
      width: 150,
      key: 'attendanceCount',
    },
    {
      title: '对内备注',
      dataIndex: 'internalNote',
      width: 150,
      key: 'internalNote',
    },
    {
      title: '对外备注',
      dataIndex: 'externalNote',
      width: 150,
      key: 'externalNote',
    },
    {
      title: '操作',
      dataIndex: 'action',
      fixed: 'right',
      width: 80,
      key: 'action',
    },
  ],
)
// 计算表格总宽度
const totalWidth = computed(() =>
  columns.value.reduce((acc, col) => acc + (col.width || 0), 0),
)
// 计算表格高度 500 换成动态的值
const customScrollY = ref(window.innerHeight - 500)
const userName = ref('')
// 修改为独立的状态控制
const headerStatus = ref('attended') // 默认为到课
// 根据表头状态设置每个学生的状态
function setAllStudentStatus(status) {
  data.value.forEach((item) => {
    // 重置所有状态
    item.attended = false
    item.absent = false
    item.leave = false
    item.unrecorded = false
    // 设置选中的状态
    if (status) {
      item[status] = true
    }
  })
}
// 计算属性监听表头状态变化
watch(() => headerStatus.value, (newStatus) => {
  // 当表头状态变化时，更新所有学生的状态
  if (newStatus === '') {
    return
  }
  setAllStudentStatus(newStatus)
})
// 处理表头批量操作
function handleHeaderStatusChange(status) {
  if (headerStatus.value === status) {
    // 如果点击当前选中的状态，则取消选择
    headerStatus.value = ''
    setAllStudentStatus('')
  }
  else {
    // 否则切换到新状态
    headerStatus.value = status
    setAllStudentStatus(status)
  }
}
// 处理单个学生状态变更
function handleStudentStatusChange(record, status) {
  // 判断是否是取消选择当前状态
  if (record[status]) {
    // 如果当前状态已经选中，则取消选择
    record[status] = false
  }
  else {
    // 重置该学生的所有状态
    record.attended = false
    record.absent = false
    record.leave = false
    record.unrecorded = false

    // 设置新状态
    record[status] = true
  }

  // 检查表头状态，只有全部学生选中同一状态时才改变表头状态
  const allSelected = data.value.every(item => item[status])
  const anySelected = data.value.some(item => item[status])

  if (allSelected) {
    headerStatus.value = status
  }
  else if (!anySelected && headerStatus.value === status) {
    // 如果没有学生选中此状态，且表头状态为此状态，则清除表头状态
    headerStatus.value = ''
  }
  else if (!anySelected) {
    headerStatus.value = ''
  }
  else {
    headerStatus.value = ''
  }
}
// 计算出席统计
const attendanceStats = computed(() => {
  return {
    attended: data.value.filter(student => student.attended).length,
    absent: data.value.filter(student => student.absent).length,
    leave: data.value.filter(student => student.leave).length,
    unrecorded: data.value.filter(student => student.unrecorded).length,
  }
})
// 编辑上课信息
const editClassInfoModal = ref(false)
function handleEditClassInfo() {
  editClassInfoModal.value = true
}

// 监听窗口大小变化
function onResize() {
  customScrollY.value = window.innerHeight - 500
}

// 组件挂载时添加监听
onMounted(() => {
  window.addEventListener('resize', onResize)
})

// 组件卸载时移除监听
onUnmounted(() => {
  window.removeEventListener('resize', onResize)
})
// 添加学员modal
const addStudentModal = ref(false)
const addStudentModalTitle = ref('')
// 添加学员
function handleAddStudent({ key }) {
  if (key === '1') {
    // 补课学员 跳转补课学员页面
    addStudentModalTitle.value = '添加补课学员'
  }
  else if (key === '2') {
    // 临时学员 跳转临时学员页面
    addStudentModalTitle.value = '添加临时学员'
  }
  else if (key === '3') {
    // 试听学员 跳转试听学员页面
    addStudentModalTitle.value = '添加试听学员'
  }
  addStudentModal.value = true
}
// 批量编辑modal
const batchEditModal = ref(false)
function handleBatchEdit() {
  batchEditModal.value = true
}
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer" :push="{ distance: 80 }" :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false" width="1244px" placement="right"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            上课点名
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="tips bg-#e6f0ff py-12px px-16px text-#06f">
        <ExclamationCircleFilled /> 未关注家校平台的学员家长，无法接收课消提醒
      </div>
      <div class="contenter flex flex-center bg-white px6 py3">
        <div class="avatarBox w-16 h-16 relative">
          <img
            width="64" height="64" src="https://pcsys.admin.ybc365.com/83b8fd68-2f9b-4a35-979f-1fd0ea349889.png"
            alt=""
          >
        </div>
        <div class="info flex flex-1 ml-4 flex-col">
          <div class="top flex justify-between flex-center flex-1">
            <a-space>
              <div class="name text-5 font-800">
                陈陈-初级感统课
              </div>
            </a-space>
          </div>
          <div class="bottom flex-1 flex flex-items-center mt-2">
            <div class="birthday flex-center">
              <span class="text-4 text-#222">2025-04-14(周一)10:00 ~ 10:30</span>
              <span class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">30分钟</span>
            </div>
          </div>
        </div>
      </div>
      <div class="desc pt-4 bg-white px6 py3 pb0">
        <a-descriptions :column="3" size="small" :content-style="{ color: '#333' }">
          <a-descriptions-item label="上课教师">
            张晨
          </a-descriptions-item>
          <a-descriptions-item label="上课助教">
            陈瑞生
          </a-descriptions-item>
          <a-descriptions-item label="上课教室">
            -
          </a-descriptions-item>
          <a-descriptions-item label="本次上课">
            教师记录 <span class="text-#f03 mx-1">1</span> 课时
          </a-descriptions-item>
          <a-descriptions-item>
            <span class="text-#06f cursor-pointer" @click="handleEditClassInfo">编辑上课信息</span>
          </a-descriptions-item>
        </a-descriptions>
      </div>
      <div class="tables bg-#fff pt-16px px-24px pb-30px">
        <a-input v-model:value="userName" placeholder="搜索学员" class="h-48px rounded-12px">
          <template #prefix>
            <img
              src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/magnifying.2bcc08ab.svg"
              alt="" class="pr-6px mt--2px"
            >
          </template>
        </a-input>
        <!-- 用a-table 学员/扣费课程账户	到课  旷课  请假  未记录  课消方式 	上课点名数量	对内备注	对外备注	操作 -->
        <!-- 带序号 -->
        <a-table
          :columns="columns" :data-source="data" row-key="id" class="mt-12px" :pagination="false"
          :scroll="{ x: totalWidth, y: customScrollY }"
        >
          <template #headerCell="{ column }">
            <div v-if="column.dataIndex === 'studentAccount'">
              <div class="text-#333 font-800">
                {{ column.title }}
                <a-popover title="学员/扣费课程账户">
                  <template #content>
                    <div class="w-450px">
                      【扣费课程账户】当前点名所消耗的课程账户，报读相同课程且课消方式相同时，会合并计算为一个扣费课程账户，支持切换课程账户课消
                      <br>
                      【剩余数量】对应课程账户相关的剩余数量（课时/金额/天数）
                    </div>
                  </template>
                  <ExclamationCircleOutlined class="text-#06f cursor-pointer ml-4px" />
                </a-popover>
              </div>
            </div>
            <div v-if="column.dataIndex === 'attended'">
              <a-tooltip>
                <template #title>
                  {{ headerStatus === 'attended' ? '取消批量到课' : '批量到课' }}
                </template>
                <div class="text-#333 font-800">
                  <a-checkbox
                    :checked="headerStatus === 'attended'" class="status-checkbox attended-checkbox"
                    :class="{ 'active-checkbox': headerStatus === 'attended' }"
                    @click="() => handleHeaderStatusChange('attended')"
                  >
                    {{
                      column.title }}
                  </a-checkbox>
                </div>
              </a-tooltip>
            </div>
            <div v-if="column.dataIndex === 'absent'">
              <a-tooltip>
                <template #title>
                  {{ headerStatus === 'absent' ? '取消批量旷课' : '批量旷课' }}
                </template>
                <div class="text-#333 font-800">
                  <a-checkbox
                    :checked="headerStatus === 'absent'" class="status-checkbox absent-checkbox"
                    :class="{ 'active-checkbox': headerStatus === 'absent' }"
                    @click="() => handleHeaderStatusChange('absent')"
                  >
                    {{
                      column.title }}
                  </a-checkbox>
                </div>
              </a-tooltip>
            </div>
            <div v-if="column.dataIndex === 'leave'">
              <a-tooltip>
                <template #title>
                  {{ headerStatus === 'leave' ? '取消批量请假' : '批量请假' }}
                </template>
                <div class="text-#333 font-800">
                  <a-checkbox
                    :checked="headerStatus === 'leave'" class="status-checkbox leave-checkbox"
                    :class="{ 'active-checkbox': headerStatus === 'leave' }"
                    @click="() => handleHeaderStatusChange('leave')"
                  >
                    {{
                      column.title }}
                  </a-checkbox>
                </div>
              </a-tooltip>
            </div>
            <div v-if="column.dataIndex === 'unrecorded'">
              <a-popover title="字段说明">
                <template #content>
                  <div class="w-300px">
                    学员为"未记录"状态时，无法记录课时，也不会发送家长端消息提醒。
                  </div>
                </template>
                <div class="text-#333 font-800">
                  <a-checkbox
                    :checked="headerStatus === 'unrecorded'"
                    class="status-checkbox unrecorded-checkbox" :class="{ 'active-checkbox': headerStatus === 'unrecorded' }"
                    @click="() => handleHeaderStatusChange('unrecorded')"
                  >
                    {{
                      column.title }}
                    <ExclamationCircleOutlined class="cursor-pointer mr-4px" />
                  </a-checkbox>
                </div>
              </a-popover>
            </div>
            <div v-if="column.dataIndex === 'consumptionMethod'">
              <a-popover title="课消方式">
                <template #content>
                  <div class="w-300px">
                    【按时间】按天数计费 <br>
                    【按课时】按课时计费 <br>
                    【按金额】按金额计费 <br>
                    【提示】按时间、按金额计费，开启记录课时后仅作为「记录」，课时增减不产生学费变动。
                  </div>
                </template>
                <div class="text-#333 font-800">
                  {{ column.title }}
                  <ExclamationCircleOutlined class="cursor-pointer mr-4px" />
                </div>
              </a-popover>
            </div>
          </template>
          <template #bodyCell="{ column, record, index }">
            <div v-if="column.dataIndex === 'index'">
              {{ index + 1 }} {{ record.name }}
            </div>
            <div v-if="column.dataIndex === 'studentAccount'">
              <div class="flex flex-items-center text-3">
                <div class=" mr-4px">
                  <img
                    src="https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png"
                    class="w-40px h-40px rounded-full mr-6px" alt=""
                  >
                  <span
                    v-if="record.type === '3'"
                    class=" flex bg-#fff5e6 text-#f90 w-120% justify-center ml--8px text-10px rounded-10"
                  >免费试听</span>
                  <span
                    v-if="record.type === '4'"
                    class=" flex bg-#734338 text-#fff w-120% justify-center ml--8px text-10px rounded-10"
                  >补课学员</span>
                  <span
                    v-if="record.type === '2'"
                    class=" flex bg-#888 text-#fff w-120% justify-center ml--8px text-10px rounded-10"
                  >临时学员</span>
                </div>
                <div class="text-#888">
                  <div class="text-14px text-#333 mb-2px">
                    {{ record.studentAccount }} <span
                      class="bg-#e6f0ff text-#06f text-3 px2 py2px rounded-10 ml2px"
                    >已关注</span>
                  </div>
                  <div>高级言语课</div>
                  <div>剩余课时：22</div>
                  <div class="text-#f90">
                    已请假：2次
                  </div>
                </div>
              </div>
            </div>
            <div v-if="column.dataIndex === 'attended'">
              <div
                class="text-#333 attended-status"
                :class="{ 'active-status': record.attended, 'inactive-status': !record.attended }"
              >
                <a-checkbox
                  :checked="record.attended" class="status-checkbox attended-checkbox"
                  :class="{ 'active-checkbox': record.attended }" @click="() => handleStudentStatusChange(record, 'attended')"
                >
                  到课
                </a-checkbox>
              </div>
            </div>
            <div v-if="column.dataIndex === 'absent'">
              <div
                class="text-#333 cursor-pointer absent-status"
                :class="{ 'active-status': record.absent, 'inactive-status': !record.absent }"
              >
                <a-checkbox
                  :checked="record.absent" class="status-checkbox absent-checkbox"
                  :class="{ 'active-checkbox': record.absent }" @click="() => handleStudentStatusChange(record, 'absent')"
                >
                  旷课
                </a-checkbox>
              </div>
            </div>
            <div v-if="column.dataIndex === 'leave'">
              <div
                class="text-#333 cursor-pointer leave-status"
                :class="{ 'active-status': record.leave, 'inactive-status': !record.leave }"
              >
                <a-checkbox
                  :checked="record.leave" class="status-checkbox leave-checkbox"
                  :class="{ 'active-checkbox': record.leave }" @click="() => handleStudentStatusChange(record, 'leave')"
                >
                  请假
                </a-checkbox>
              </div>
            </div>
            <div v-if="column.dataIndex === 'unrecorded'">
              <div
                class="text-#333 cursor-pointer unrecorded-status"
                :class="{ 'active-status': record.unrecorded, 'inactive-status': !record.unrecorded }"
              >
                <a-checkbox
                  :checked="record.unrecorded" class="status-checkbox unrecorded-checkbox"
                  :class="{ 'active-checkbox': record.unrecorded }"
                  @click="() => handleStudentStatusChange(record, 'unrecorded')"
                >
                  未记录
                </a-checkbox>
              </div>
            </div>
            <div v-if="column.dataIndex === 'consumptionMethod'">
              <span>{{ record.consumptionMethod === '1' ? '按课时' : record.consumptionMethod === '2' ? '按时间' : '按金额'
              }}</span>
              <!-- 分割线 -->
              <span class="flex w-47px">
                <a-divider
                  v-if="record.consumptionMethod !== '1' && !record.unrecorded && record.type !== '3'"
                  class="my-2px"
                />
              </span>
              <div
                v-if="record.consumptionMethod !== '1' && !record.unrecorded && record.type !== '3'"
                class="text-#888 text-3 flex flex-col"
              >
                <span>记录课时</span>
                <a-switch v-model:checked="record.recordAttendance" class="w-35px" />
              </div>
            </div>
            <div v-if="column.dataIndex === 'attendanceCount'">
              <span v-if="record.recordAttendance && !record.unrecorded" class="flex flex-items-center"><a-input-number
                v-model:value="record.attendanceCount" :min="0" :precision="2" class="w-80px mr-4px"
              />课时</span>
              <!-- 当是未记录时，展示不计课时，不发送家长端消息提示 -->
              <span v-else-if="record.unrecorded && record.type !== '3'" class="flex flex-col">
                <span>不计课时</span>
                <span class="text-3 text-#999">不发送家长端 <br> 消息提示</span>
              </span>
              <span v-else-if="record.type === '3'">
                <div class="text-#888">免费试听学员</div>
                <span class="text-#999">不支持记课时</span>
              </span>
              <span v-else class="text-#888">不计课时</span>
            </div>
            <div v-if="column.dataIndex === 'internalNote'">
              <a-input v-model:value="record.internalNote" class="w-100px" placeholder="请输入" />
            </div>
            <div v-if="column.dataIndex === 'externalNote'">
              <a-input v-model:value="record.externalNote" class="w-100px" placeholder="请输入" />
            </div>
            <div v-else-if="column.dataIndex === 'action'">
              <a-space>
                <a>移出</a>
              </a-space>
            </div>
          </template>
        </a-table>
      </div>
      <!-- 自定义footer -->
      <template #footer>
        <div class="h-60px flex flex-items-center justify-between px-24px">
          <a-space :size="20">
            <a-dropdown>
              <template #overlay>
                <a-menu @click="handleAddStudent">
                  <a-menu-item key="1">
                    补课学员
                  </a-menu-item>
                  <a-menu-item key="2">
                    临时学员
                  </a-menu-item>
                  <a-menu-item key="3">
                    试听学员
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button type="primary" ghost class="h-40px text-16px">
                添加学员
                <DownOutlined class="text-12px rotate-icon" />
              </a-button>
            </a-dropdown>
            <a-button type="primary" ghost class="h-40px text-16px" @click="handleBatchEdit">
              批量编辑点名数量
            </a-button>
          </a-space>
          <a-space :size="20">
            <div class="flex flex-col text-#222 text-16px font-500">
              <span class="mb-4px">共{{ data.length }}名学员</span>
              <span>到课{{ attendanceStats.attended }}人，请假{{ attendanceStats.leave }}人，旷课{{ attendanceStats.absent
              }}人，未记录{{
                attendanceStats.unrecorded }}人</span>
            </div>
            <a-button type="primary" class="h-48px text-18px w-140px font500">
              确认点名
            </a-button>
          </a-space>
        </div>
      </template>
    </a-drawer>
    <!-- 编辑上课信息 -->
    <EditClassInfoModal v-model:open="editClassInfoModal" />
    <!-- 添加学员modal -->
    <roll-call-add-student-modal v-model:open="addStudentModal" :title="addStudentModalTitle" />
    <!-- 批量编辑 -->
    <roll-call-batch-edit-modal v-model:open="batchEditModal" />
  </div>
</template>

<style lang="less" scoped>
/* 添加旋转动画 */
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.close-btn {
  &:hover {
    background: transparent;

    .close-icon {
      animation: icon-rotate 0.3s linear;
    }
  }
}

/* 添加选中状态样式 */
.active-checkbox {
  &.attended-checkbox {
    :deep(.ant-checkbox-inner) {
      background-color: #1890ff !important;
      border-color: #1890ff !important;
    }
  }

  &.absent-checkbox {
    :deep(.ant-checkbox-inner) {
      background-color: #f33 !important;
      border-color: #f33 !important;
    }
  }

  &.leave-checkbox {
    :deep(.ant-checkbox-inner) {
      background-color: #f90 !important;
      border-color: #f90 !important;
    }
  }

  &.unrecorded-checkbox {
    :deep(.ant-checkbox-inner) {
      background-color: #888 !important;
      border-color: #888 !important;
    }
  }
}

.active-status {
  &.attended-status {
    color: #1890ff !important;
  }

  &.absent-status {
    color: #f33 !important;
  }

  &.leave-status {
    color: #f90 !important;
  }

  &.unrecorded-status {
    color: #888 !important;
  }

  opacity: 1 !important;
}

.inactive-status {
  opacity: 0.4;
  transition: opacity 0.3s, color 0.3s;

  &:hover {
    opacity: 1;
  }

  &.attended-status:hover {
    color: #1890ff !important;
  }

  &.absent-status:hover {
    color: #f33 !important;
  }

  &.leave-status:hover {
    color: #f90 !important;
  }

  &.unrecorded-status:hover {
    color: #888 !important;
  }

  /* 添加复选框hover效果 */
  &.attended-status:hover {
    :deep(.ant-checkbox-inner) {
      border-color: #1890ff !important;
    }
  }

  &.absent-status:hover {
    :deep(.ant-checkbox-inner) {
      border-color: #f33 !important;
    }
  }

  &.leave-status:hover {
    :deep(.ant-checkbox-inner) {
      border-color: #f90 !important;
    }
  }

  &.unrecorded-status:hover {
    :deep(.ant-checkbox-inner) {
      border-color: #888 !important;
    }
  }
}

/* 添加旋转过渡效果 */
.rotate-icon {
  display: inline-block;
  transition: transform 0.3s ease;
}

/* 当按钮悬停时旋转图标 */
.h-40px:hover .rotate-icon {
  transform: rotate(180deg);
}
</style>
