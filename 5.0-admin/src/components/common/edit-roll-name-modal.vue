<script setup lang="ts">
import { DownOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import type { TeachingRecordDetailResult, TeachingRecordDetailStudent } from '@/api/edu-center/class-record'
import messageService from '@/utils/messageService'
import EditRollNameAddStuModal from './edit-roll-name-add-stu-modal.vue'
import EditRowRollNameModal from './edit-row-roll-name-modal.vue'

const props = withDefaults(defineProps<{
  open: boolean
  detail?: TeachingRecordDetailResult | null
}>(), {
  detail: null,
})

const emit = defineEmits(['update:open', 'updated'])

const openDrawer = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const searchKeyword = ref('')
const editRowRollNameModals = ref(false)
const selectedStudent = ref<TeachingRecordDetailStudent | null>(null)
const editRollNameAddStuModals = ref(false)
const addStudentModalTitle = ref('')

const columns = ref<any[]>([
  {
    title: '',
    dataIndex: 'index',
    fixed: 'left',
    width: 40,
  },
  {
    title: '学员',
    dataIndex: 'name',
    fixed: 'left',
    width: 220,
  },
  {
    title: '上课状态',
    dataIndex: 'status',
    width: 160,
    filters: [
      { text: '未点名', value: '0' },
      { text: '到课', value: '1' },
      { text: '旷课', value: '2' },
      { text: '请假', value: '3' },
      { text: '未记录', value: '4' },
    ],
    onFilter: (value: string | number | boolean, record: TeachingRecordDetailStudent) => String(record.status ?? '') === String(value),
  },
  {
    title: '课消方式',
    dataIndex: 'courseType',
    width: 140,
  },
  {
    title: '上课点名数量',
    dataIndex: 'rollNameCount',
    width: 160,
  },
  {
    title: '对内备注',
    dataIndex: 'innerRemark',
    width: 200,
  },
  {
    title: '对外备注',
    dataIndex: 'outerRemark',
    width: 200,
  },
  {
    title: '操作',
    dataIndex: 'action',
    width: 120,
    fixed: 'right',
  },
])

const totalWidth = computed(() => columns.value.reduce((acc, column) => acc + Number(column.width || 0), 0))
const rawStudents = computed(() => Array.isArray(props.detail?.studentList) ? props.detail?.studentList || [] : [])
const filteredStudents = computed(() => {
  const keyword = searchKeyword.value.trim().toLowerCase()
  if (!keyword)
    return rawStudents.value
  return rawStudents.value.filter((item) => {
    const name = String(item.studentName || '').toLowerCase()
    const phone = String(item.studentPhone || '').toLowerCase()
    return name.includes(keyword) || phone.includes(keyword)
  })
})

function resolveRowKey(record: Partial<TeachingRecordDetailStudent>) {
  const teachingRecordId = String(record.studentTeachingRecordId || '').trim()
  if (teachingRecordId)
    return teachingRecordId
  const studentId = String(record.studentId || '').trim()
  if (studentId)
    return `pending-${studentId}`
  return `pending-${String(record.studentName || '').trim()}`
}

function defaultAvatar(record: Partial<TeachingRecordDetailStudent>) {
  return String(record.avatar || '').trim() || 'https://cdn.schoolpal.cn/schoolpal/next-erp/avator_male.png'
}

function sourceTypeText(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return '临时学员'
  if (type === 3 || type === 7)
    return '补课学员'
  if (type === 4)
    return '试听学员'
  if (type === 6)
    return '1对1学员'
  return '班级学员'
}

function sourceTypeBadge(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return '临时学员'
  if (type === 3 || type === 7)
    return '补课学员'
  if (type === 4)
    return '免费试听'
  return ''
}

function sourceTypeBadgeClass(value?: number) {
  const type = Number(value || 0)
  if (type === 2)
    return 'student-type-badge student-type-badge--temporary'
  if (type === 3 || type === 7)
    return 'student-type-badge student-type-badge--makeup'
  if (type === 4)
    return 'student-type-badge student-type-badge--trial'
  return ''
}

function statusText(value?: number) {
  const status = Number(value || 0)
  if (status === 0)
    return '未点名'
  if (status === 2)
    return '旷课'
  if (status === 3)
    return '请假'
  if (status === 4)
    return '未记录'
  return '到课'
}

function statusClass(value?: number) {
  const status = Number(value || 0)
  if (status === 0)
    return 'status-tag status-tag--pending'
  if (status === 2)
    return 'status-tag status-tag--absent'
  if (status === 3)
    return 'status-tag status-tag--leave'
  if (status === 4)
    return 'status-tag status-tag--unrecorded'
  return 'status-tag status-tag--arrived'
}

function chargingModeText(record: Partial<TeachingRecordDetailStudent>) {
  const mode = Number(record.skuMode || 0)
  if (mode === 2)
    return '按时间'
  if (mode === 3)
    return '按金额'
  if (mode === 1)
    return '按课时'
  return '-'
}

function hasTeachingRecord(record: Partial<TeachingRecordDetailStudent>) {
  return String(record.studentTeachingRecordId || '').trim() !== ''
}

function formatNumber(value?: number) {
  const num = Number(value || 0)
  if (!Number.isFinite(num))
    return '0'
  return Number.isInteger(num) ? String(num) : num.toFixed(2).replace(/\.?0+$/, '')
}

function rollNameCountText(record: Partial<TeachingRecordDetailStudent>) {
  if (!hasTeachingRecord(record))
    return '未记录课时'
  const quantity = Number(record.quantity || 0)
  if (quantity > 0)
    return `${formatNumber(quantity)}课时`
  return '不计课时'
}

function handleEdit(record: Record<string, any>) {
  selectedStudent.value = record as TeachingRecordDetailStudent
  editRowRollNameModals.value = true
}

function handleAddStudent(info: { key: string | number }) {
  const key = String(info.key)
  if (key === '1') {
    messageService.info('补课学员功能暂未开发')
    return
  }
  if (key === '2') {
    addStudentModalTitle.value = '添加临时学员'
  }
  else if (key === '3') {
    addStudentModalTitle.value = '添加试听学员'
  }
  else {
    addStudentModalTitle.value = '添加学员'
  }
  editRollNameAddStuModals.value = true
}

function getRemoveDisabledReason(record: Partial<TeachingRecordDetailStudent>) {
  if (hasTeachingRecord(record))
    return '该学员已点名，不可移出本节'
  return ''
}

function handleRemoveStudent(record: Record<string, any>) {
  const student = record as TeachingRecordDetailStudent
  const name = String(student?.studentName || '').trim() || '当前学员'
  const disabledReason = getRemoveDisabledReason(student)
  if (disabledReason) {
    messageService.warning(disabledReason)
    return
  }
  messageService.info(`“${name}”的移出本节功能暂未接入，当前先保留静态入口`)
}

watch(openDrawer, (value) => {
  if (!value) {
    searchKeyword.value = ''
    selectedStudent.value = null
    editRowRollNameModals.value = false
    editRollNameAddStuModals.value = false
  }
})
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer"
      :push="{ distance: 80 }"
      :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false"
      width="1165px"
      placement="right"
    >
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            编辑点名
          </div>
        </div>
      </template>
      <div class="search px-24px py-12px bg-white">
        <a-input v-model:value="searchKeyword" placeholder="搜索学员" class="h-48px rounded-12px">
          <template #prefix>
            <img
              src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/magnifying.2bcc08ab.svg"
              alt=""
              class="pr-6px mt--2px"
            >
          </template>
        </a-input>
      </div>
      <div class="contenter bg-white">
          <a-table
          :columns="columns"
          :data-source="filteredStudents"
          :pagination="false"
          :scroll="{ x: totalWidth }"
          :row-key="resolveRowKey"
        >
          <template #headerCell="{ column }">
            <div v-if="column.dataIndex === 'courseType'" class="flex items-center">
              {{ column.title }}
              <a-popover title="课消方式">
                <template #content>
                  <div class="w-320px">
                    【按时间】按时间计费<br>
                    【按课时】按课时计费<br>
                    【按金额】按金额计费
                  </div>
                </template>
                <ExclamationCircleOutlined class="text-#333 cursor-pointer ml-4px" />
              </a-popover>
            </div>
          </template>

          <template #bodyCell="{ column, record, index }">
            <div v-if="column.dataIndex === 'index'">
              {{ index + 1 }}
            </div>

            <div v-if="column.dataIndex === 'name'">
              <div class="flex items-center">
                <div class="mr-8px avatar-wrap">
                  <img :src="defaultAvatar(record)" class="w-40px h-40px rounded-full" alt="">
                  <span v-if="sourceTypeBadge(record.sourceType)" :class="sourceTypeBadgeClass(record.sourceType)">
                    {{ sourceTypeBadge(record.sourceType) }}
                  </span>
                </div>
                <div class="flex flex-col">
                  <span class="text-#333">{{ record.studentName || '-' }}</span>
                  <span class="text-#8c8c8c text-12px">{{ sourceTypeText(record.sourceType) }}</span>
                </div>
              </div>
            </div>

            <div v-if="column.dataIndex === 'status'">
              <span :class="statusClass(record.status)">{{ statusText(record.status) }}</span>
            </div>

            <div v-if="column.dataIndex === 'courseType'">
              {{ chargingModeText(record) }}
            </div>

            <div v-if="column.dataIndex === 'rollNameCount'">
              {{ rollNameCountText(record) }}
            </div>

            <div v-if="column.dataIndex === 'innerRemark'">
              {{ record.remark || '-' }}
            </div>

            <div v-if="column.dataIndex === 'outerRemark'">
              {{ record.externalRemark || '-' }}
            </div>

            <div v-if="column.dataIndex === 'action'">
              <a-space :size="20">
                <a @click="handleEdit(record)">编辑</a>
                <a @click="handleRemoveStudent(record)">移出</a>
              </a-space>
            </div>
          </template>
        </a-table>
      </div>
      <template #footer>
        <a-dropdown>
          <template #overlay>
            <a-menu @click="handleAddStudent">
              <a-menu-item key="1">
                添加补课学员
              </a-menu-item>
              <a-menu-item key="2">
                添加临时学员
              </a-menu-item>
              <a-menu-item key="3">
                添加试听学员
              </a-menu-item>
            </a-menu>
          </template>
          <a-button type="primary" ghost class="h-40px w-120px text-16px ml-12px">
            添加学员
            <DownOutlined class="text-12px rotate-icon" />
          </a-button>
        </a-dropdown>
      </template>
    </a-drawer>

    <EditRowRollNameModal
      v-model:open="editRowRollNameModals"
      :student="selectedStudent"
    />
    <EditRollNameAddStuModal
      v-model:open="editRollNameAddStuModals"
      :default-title="addStudentModalTitle"
    />
  </div>
</template>

<style lang="less" scoped>
@keyframes icon-rotate {
  from {
    transform: rotate(0deg);
  }

  to {
    transform: rotate(180deg);
  }
}

.avatar-wrap {
  position: relative;
}

.student-type-badge {
  position: absolute;
  left: -6px;
  bottom: -8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0 6px;
  border-radius: 999px;
  font-size: 10px;
  line-height: 18px;
  white-space: nowrap;
}

.student-type-badge--trial {
  color: #f90;
  background: #fff5e6;
}

.student-type-badge--makeup {
  color: #fff;
  background: #734338;
}

.student-type-badge--temporary {
  color: #fff;
  background: #888;
}

.status-tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 2px 10px;
  border-radius: 999px;
  font-size: 12px;
  line-height: 20px;
}

.status-tag--arrived {
  color: #06f;
  background: #e6f0ff;
}

.status-tag--leave {
  color: #f90;
  background: #fff5e6;
}

.status-tag--absent {
  color: #f33;
  background: #ffe6e6;
}

.status-tag--pending,
.status-tag--unrecorded {
  color: #888;
  background: #f6f7f8;
}

.rotate-icon {
  display: inline-block;
  transition: transform 0.3s ease;
}

.h-40px:hover .rotate-icon {
  transform: rotate(180deg);
}
</style>
