<script setup lang="ts">
import { PlusOutlined } from '@ant-design/icons-vue'
import type { TableColumnType } from 'ant-design-vue'
import {
  type ClassroomItem,
  createClassroomApi,
  listClassroomsApi,
  updateClassroomApi,
  updateClassroomStatusApi,
} from '@/api/business-settings/classroom'
import messageService from '@/utils/messageService'

interface ClassroomFormState {
  id?: number
  name: string
  address: string
  enabled: boolean
}

type ClassroomRowLike = ClassroomItem | Record<string, any>

const loading = ref(false)
const submitting = ref(false)
const formModalOpen = ref(false)
const classrooms = ref<ClassroomItem[]>([])

const formState = reactive<ClassroomFormState>({
  name: '',
  address: '',
  enabled: true,
})

const columns: TableColumnType<ClassroomItem>[] = [
  { title: '教室名称', dataIndex: 'name', key: 'name', ellipsis: true },
  { title: '教室状态', key: 'status', width: 120 },
  { title: '操作', key: 'action', width: 180 },
]

async function loadClassrooms() {
  loading.value = true
  try {
    const res = await listClassroomsApi()
    if (res.code === 200) {
      classrooms.value = Array.isArray(res.result) ? res.result : []
      return
    }
    messageService.error(res.message || '获取教室列表失败')
  }
  catch (error: any) {
    console.error('load classrooms failed', error)
    messageService.error(error?.message || '获取教室列表失败')
  }
  finally {
    loading.value = false
  }
}

onMounted(() => {
  loadClassrooms()
})

function resetForm() {
  formState.id = undefined
  formState.name = ''
  formState.address = ''
  formState.enabled = true
}

function openCreateModal() {
  resetForm()
  formModalOpen.value = true
}

function openEditModal(item: ClassroomRowLike) {
  formState.id = Number(item.id)
  formState.name = String(item.name || '')
  formState.address = String(item.address || '')
  formState.enabled = !!item.enabled
  formModalOpen.value = true
}

function closeFormModal() {
  formModalOpen.value = false
}

function setFormModalOpen(v: boolean) {
  formModalOpen.value = v
}

watch(formModalOpen, (open) => {
  if (!open)
    resetForm()
})

async function submitForm() {
  if (!formState.name.trim()) {
    messageService.error('请输入教室名称')
    return
  }
  submitting.value = true
  try {
    const payload = {
      id: formState.id,
      name: formState.name.trim(),
      address: formState.address.trim(),
      enabled: !!formState.enabled,
    }
    const res = formState.id
      ? await updateClassroomApi(payload)
      : await createClassroomApi(payload)
    if (res.code !== 200) {
      messageService.error(res.message || (formState.id ? '更新教室失败' : '新增教室失败'))
      return
    }
    messageService.success(formState.id ? '教室更新成功' : '教室新增成功')
    formModalOpen.value = false
    await loadClassrooms()
  }
  catch (error: any) {
    console.error('submit classroom failed', error)
    messageService.error(error?.message || (formState.id ? '更新教室失败' : '新增教室失败'))
  }
  finally {
    submitting.value = false
  }
}

async function toggleClassroomStatus(item: ClassroomRowLike, checked: boolean) {
  try {
    const res = await updateClassroomStatusApi({
      id: Number(item.id),
      enabled: checked,
    })
    if (res.code !== 200) {
      messageService.error(res.message || '更新教室状态失败')
      return
    }
    item.enabled = checked
    messageService.success(checked ? '已启用该教室' : '已停用该教室')
  }
  catch (error: any) {
    console.error('update classroom status failed', error)
    messageService.error(error?.message || '更新教室状态失败')
  }
}
</script>

<template>
  <div class="classroom-settings scrollbar">
    <div class="classroom-settings__panel">
      <div class="classroom-panel__head">
        <div class="classroom-panel__summary">
          <span class="classroom-panel__accent" aria-hidden="true" />
          <span class="classroom-panel__summary-text">当前共计 {{ classrooms.length }} 个教室</span>
        </div>
        <a-button type="primary" class="classroom-panel__create" @click="openCreateModal">
          <template #icon>
            <PlusOutlined />
          </template>
          创建教室
        </a-button>
      </div>

      <a-spin :spinning="loading">
        <a-table
          class="classroom-table"
          :columns="columns"
          :data-source="classrooms"
          :pagination="false"
          row-key="id"
        >
          <template #bodyCell="{ column, text, record }">
            <template v-if="column.key === 'status'">
              <span
                class="classroom-status"
                :class="record.enabled ? 'classroom-status--on' : 'classroom-status--off'"
              >
                {{ record.enabled ? '启用' : '停用' }}
              </span>
            </template>
            <template v-else-if="column.key === 'action'">
              <a-button type="link" size="small" class="classroom-action classroom-action--edit" @click="openEditModal(record)">
                编辑
              </a-button>
              <a-button
                v-if="record.enabled"
                type="link"
                size="small"
                danger
                class="classroom-action"
                @click="toggleClassroomStatus(record, false)"
              >
                停用
              </a-button>
              <a-button
                v-else
                type="link"
                size="small"
                class="classroom-action classroom-action--enable"
                @click="toggleClassroomStatus(record, true)"
              >
                启用
              </a-button>
            </template>
            <template v-else>
              {{ text }}
            </template>
          </template>
        </a-table>
      </a-spin>
    </div>

    <a-modal
      :open="formModalOpen"
      class="classroom-form-modal"
      :title="formState.id ? '编辑教室' : '创建教室'"
      :width="520"
      :mask-closable="false"
      destroy-on-close
      :footer="null"
      @update:open="setFormModalOpen"
      @cancel="closeFormModal"
    >
      <div class="classroom-modal__body">
        <div class="classroom-form-block">
          <span class="classroom-form-block__label required">教室名称</span>
          <a-input
            v-model:value="formState.name"
            :maxlength="50"
            placeholder="例如：个训室 01"
          />
        </div>

        <div class="classroom-form-block">
          <span class="classroom-form-block__label">教室位置</span>
          <a-input
            v-model:value="formState.address"
            :maxlength="100"
            placeholder="例如：二楼东侧"
          />
        </div>

        <div class="classroom-form-block">
          <div class="classroom-form-block__switch">
            <span class="classroom-form-block__label">是否启用</span>
            <a-switch v-model:checked="formState.enabled" />
          </div>
        </div>
      </div>

      <div class="classroom-modal__footer">
        <a-button @click="closeFormModal">
          取消
        </a-button>
        <a-button type="primary" :loading="submitting" @click="submitForm">
          保存
        </a-button>
      </div>
    </a-modal>
  </div>
</template>

<style scoped lang="less">
.classroom-settings {
  position: relative;
  height: 100%;
  overflow-y: auto;
  background: #f2f4f7;
}

.classroom-settings__panel {
  margin: 12px 16px 20px;
  padding: 18px 20px 12px;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 1px 4px rgb(15 23 42 / 6%);
}

.classroom-panel__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
}

.classroom-panel__summary {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}

.classroom-panel__accent {
  display: inline-block;
  width: 4px;
  height: 16px;
  flex-shrink: 0;
  border-radius: 2px;
  background: #1677ff;
}

.classroom-panel__summary-text {
  font-size: 15px;
  font-weight: 500;
  color: #1f2329;
  line-height: 1.4;
}

.classroom-panel__create {
  flex-shrink: 0;
  border-radius: 6px;
}

.classroom-table {
  :deep(.ant-table) {
    background: transparent;
  }

  :deep(.ant-table-thead > tr > th) {
    padding: 12px 16px;
    font-weight: 500;
    color: #262626;
    background: #fafafa !important;
    border-bottom: 1px solid #f0f0f0;
  }

  :deep(.ant-table-thead > tr > th::before) {
    display: none;
  }

  :deep(.ant-table-tbody > tr > td) {
    padding: 14px 16px;
    border-bottom: 1px solid #f5f5f5;
    background: #fff;
  }

  :deep(.ant-table-tbody > tr:last-child > td) {
    border-bottom: none;
  }
}

.classroom-status {
  font-size: 14px;
  font-weight: 500;
}

.classroom-status--on {
  color: #52c41a;
}

.classroom-status--off {
  color: #8c8c8c;
  font-weight: 400;
}

.classroom-action {
  padding: 0 4px !important;
  height: auto !important;
}

.classroom-action--edit,
.classroom-action--enable {
  color: #1677ff !important;
}

.classroom-action + .classroom-action {
  margin-left: 4px;
}

.classroom-modal__body {
  max-height: min(70vh, 520px);
  overflow-y: auto;
  padding: 4px 0 8px;
}

.classroom-form-block {
  padding: 14px;
  border-radius: 18px;
  background: #fff;
}

.classroom-form-block__label {
  display: block;
  margin-bottom: 10px;
  color: #4b5563;
  font-size: 13px;
  font-weight: 600;
}

.classroom-form-block__label.required::before {
  margin-right: 4px;
  color: #ff4d4f;
  content: '*';
}

.classroom-form-block__switch {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.classroom-form-block__switch .classroom-form-block__label {
  margin-bottom: 0;
}

.classroom-modal__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}
</style>
