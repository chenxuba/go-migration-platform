<script setup lang="ts">
import {
  CloseOutlined,
  DeleteOutlined,
  EditOutlined,
  EnvironmentOutlined,
  PlusOutlined,
  SearchOutlined,
  TeamOutlined,
} from '@ant-design/icons-vue'
import { Modal } from 'ant-design-vue'
import { debounce } from 'lodash-es'
import {
  type ClassroomItem,
  createClassroomApi,
  deleteClassroomApi,
  listClassroomsApi,
  updateClassroomApi,
  updateClassroomStatusApi,
} from '@/api/business-settings/classroom'
import messageService from '@/utils/messageService'

interface ClassroomFormState {
  id?: number
  name: string
  address: string
  capacity: number | undefined
  enabled: boolean
  remark: string
  sort: number | undefined
}

const loading = ref(false)
const submitting = ref(false)
const drawerOpen = ref(false)
const searchKey = ref('')
const classrooms = ref<ClassroomItem[]>([])
const editingItem = ref<ClassroomItem | null>(null)

const formState = reactive<ClassroomFormState>({
  name: '',
  address: '',
  capacity: undefined,
  enabled: true,
  remark: '',
  sort: 0,
})

const enabledCount = computed(() => classrooms.value.filter(item => item.enabled).length)
const disabledCount = computed(() => classrooms.value.filter(item => !item.enabled).length)

const summaryItems = computed(() => [
  { label: '全部教室', value: classrooms.value.length, tone: 'primary' },
  { label: '启用中', value: enabledCount.value, tone: 'success' },
  { label: '停用中', value: disabledCount.value, tone: 'neutral' },
])

async function loadClassrooms() {
  loading.value = true
  try {
    const res = await listClassroomsApi({
      searchKey: searchKey.value.trim() || undefined,
    })
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

const debouncedLoadClassrooms = debounce(() => {
  loadClassrooms()
}, 300)

watch(searchKey, () => {
  debouncedLoadClassrooms()
})

onMounted(() => {
  loadClassrooms()
})

function resetForm() {
  editingItem.value = null
  formState.id = undefined
  formState.name = ''
  formState.address = ''
  formState.capacity = undefined
  formState.enabled = true
  formState.remark = ''
  formState.sort = 0
}

function openCreateDrawer() {
  resetForm()
  drawerOpen.value = true
}

function openEditDrawer(item: ClassroomItem) {
  editingItem.value = item
  formState.id = item.id
  formState.name = item.name || ''
  formState.address = item.address || ''
  formState.capacity = Number.isFinite(Number(item.capacity)) ? Number(item.capacity) : undefined
  formState.enabled = !!item.enabled
  formState.remark = item.remark || ''
  formState.sort = Number.isFinite(Number(item.sort)) ? Number(item.sort) : 0
  drawerOpen.value = true
}

function closeDrawer() {
  drawerOpen.value = false
  resetForm()
}

async function submitForm() {
  if (!formState.name.trim()) {
    messageService.error('请输入教室名称')
    return
  }
  if ((formState.capacity ?? 0) < 0) {
    messageService.error('容纳人数不能小于0')
    return
  }
  submitting.value = true
  try {
    const payload = {
      id: formState.id,
      name: formState.name.trim(),
      address: formState.address.trim(),
      capacity: Number(formState.capacity || 0),
      enabled: !!formState.enabled,
      remark: formState.remark.trim(),
      sort: Number(formState.sort || 0),
    }
    const res = formState.id
      ? await updateClassroomApi(payload)
      : await createClassroomApi(payload)
    if (res.code !== 200) {
      messageService.error(res.message || (formState.id ? '更新教室失败' : '新增教室失败'))
      return
    }
    messageService.success(formState.id ? '教室更新成功' : '教室新增成功')
    closeDrawer()
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

async function toggleClassroomStatus(item: ClassroomItem, checked: boolean) {
  try {
    const res = await updateClassroomStatusApi({
      id: item.id,
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

function confirmDelete(item: ClassroomItem) {
  Modal.confirm({
    title: '删除教室',
    centered: true,
    content: `删除后，该教室将不再出现在新增排课和建班选择里。确认删除“${item.name}”吗？`,
    okText: '删除',
    okType: 'danger',
    cancelText: '取消',
    async onOk() {
      try {
        const res = await deleteClassroomApi({ id: item.id })
        if (res.code !== 200) {
          messageService.error(res.message || '删除教室失败')
          return
        }
        messageService.success('教室删除成功')
        await loadClassrooms()
      }
      catch (error: any) {
        console.error('delete classroom failed', error)
        messageService.error(error?.message || '删除教室失败')
      }
    },
  })
}
</script>

<template>
  <div class="classroom-settings scrollbar" :class="{ 'overflow-hidden': drawerOpen }">
    <div class="classroom-settings__hero">
      <div class="classroom-settings__hero-title">
        管理机构教室，统一维护排课和班级创建时可选的教室范围。
      </div>
      <div class="classroom-settings__hero-stats">
        <div
          v-for="item in summaryItems"
          :key="item.label"
          class="classroom-stat"
          :class="`classroom-stat--${item.tone}`"
        >
          <strong>{{ item.value }}</strong>
          <span>{{ item.label }}</span>
        </div>
      </div>
    </div>

    <div class="classroom-toolbar">
      <a-input
        v-model:value="searchKey"
        allow-clear
        placeholder="搜索教室名称/位置/备注"
      >
        <template #prefix>
          <SearchOutlined />
        </template>
      </a-input>
      <a-button type="primary" class="classroom-toolbar__button" @click="openCreateDrawer">
        <template #icon>
          <PlusOutlined />
        </template>
        新增教室
      </a-button>
    </div>

    <a-spin :spinning="loading">
      <div v-if="classrooms.length" class="classroom-list">
        <div
          v-for="item in classrooms"
          :key="item.id"
          class="classroom-card"
        >
          <div class="classroom-card__head">
            <div class="classroom-card__title-wrap">
              <div class="classroom-card__title">
                {{ item.name }}
              </div>
              <a-tag :color="item.enabled ? 'processing' : 'default'" :bordered="false">
                {{ item.enabled ? '启用中' : '已停用' }}
              </a-tag>
            </div>
            <div class="classroom-card__actions">
              <a-button type="text" @click="openEditDrawer(item)">
                <template #icon>
                  <EditOutlined />
                </template>
              </a-button>
              <a-button type="text" danger @click="confirmDelete(item)">
                <template #icon>
                  <DeleteOutlined />
                </template>
              </a-button>
            </div>
          </div>

          <div class="classroom-card__meta">
            <div class="classroom-card__meta-item">
              <EnvironmentOutlined />
              <span>{{ item.address || '未填写位置' }}</span>
            </div>
            <div class="classroom-card__meta-item">
              <TeamOutlined />
              <span>容纳 {{ item.capacity || 0 }} 人</span>
            </div>
          </div>

          <div v-if="item.remark" class="classroom-card__remark">
            {{ item.remark }}
          </div>

          <div class="classroom-card__foot">
            <span class="classroom-card__sort">排序 {{ item.sort || 0 }}</span>
            <a-switch
              :checked="item.enabled"
              checked-children="启用"
              un-checked-children="停用"
              @change="(checked: boolean) => toggleClassroomStatus(item, checked)"
            />
          </div>
        </div>
      </div>

      <a-empty v-else class="classroom-empty" :description="searchKey ? '没有匹配的教室' : '还没有教室，先创建一个吧'">
        <a-button type="primary" @click="openCreateDrawer">
          新增教室
        </a-button>
      </a-empty>
    </a-spin>

    <a-drawer
      root-class-name="classroom-setting-drawer"
      :closable="false"
      :mask="false"
      :mask-closable="false"
      placement="bottom"
      :open="drawerOpen"
      :get-container="false"
    >
      <div class="classroom-drawer">
        <div class="classroom-drawer__header">
          <span class="classroom-drawer__title">{{ formState.id ? '编辑教室' : '新增教室' }}</span>
          <span class="classroom-drawer__close" @click="closeDrawer">
            <CloseOutlined />
          </span>
        </div>

        <div class="classroom-drawer__body">
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

          <div class="classroom-form-grid">
            <div class="classroom-form-block">
              <span class="classroom-form-block__label">容纳人数</span>
              <a-input-number
                v-model:value="formState.capacity"
                :min="0"
                :precision="0"
                placeholder="0"
                class="classroom-form-block__control"
              />
            </div>

            <div class="classroom-form-block">
              <span class="classroom-form-block__label">排序值</span>
              <a-input-number
                v-model:value="formState.sort"
                :min="0"
                :precision="0"
                placeholder="0"
                class="classroom-form-block__control"
              />
            </div>
          </div>

          <div class="classroom-form-block">
            <div class="classroom-form-block__switch">
              <span class="classroom-form-block__label">是否启用</span>
              <a-switch v-model:checked="formState.enabled" />
            </div>
          </div>

          <div class="classroom-form-block">
            <span class="classroom-form-block__label">备注</span>
            <a-textarea
              v-model:value="formState.remark"
              :maxlength="200"
              :auto-size="{ minRows: 3, maxRows: 5 }"
              placeholder="可填写适用说明、设备信息等"
            />
          </div>
        </div>

        <div class="classroom-drawer__footer">
          <a-button @click="closeDrawer">
            取消
          </a-button>
          <a-button type="primary" :loading="submitting" @click="submitForm">
            保存
          </a-button>
        </div>
      </div>
    </a-drawer>
  </div>
</template>

<style scoped lang="less">
.classroom-settings {
  position: relative;
  height: 100%;
  overflow-y: auto;
  background: #f6f7f8;

  &.overflow-hidden {
    overflow: hidden !important;
  }
}

.classroom-settings__hero {
  padding: 14px 16px 10px;
}

.classroom-settings__hero-title {
  color: #8c8c8c;
  font-size: 12px;
  line-height: 1.6;
}

.classroom-settings__hero-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
  margin-top: 12px;
}

.classroom-stat {
  padding: 10px 12px;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 6px 18px rgb(15 23 42 / 4%);
}

.classroom-stat strong {
  display: block;
  color: #1f2329;
  font-size: 18px;
  font-weight: 700;
  line-height: 1.1;
}

.classroom-stat span {
  display: block;
  margin-top: 4px;
  color: #8c8c8c;
  font-size: 12px;
  line-height: 1.4;
}

.classroom-stat--primary {
  background: linear-gradient(180deg, #ffffff 0%, #f7fbff 100%);
}

.classroom-stat--success {
  background: linear-gradient(180deg, #ffffff 0%, #f6fffb 100%);
}

.classroom-toolbar {
  display: flex;
  gap: 10px;
  padding: 0 16px 12px;
}

.classroom-toolbar :deep(.ant-input-affix-wrapper) {
  border-radius: 14px;
}

.classroom-toolbar__button {
  flex-shrink: 0;
  border-radius: 14px;
}

.classroom-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 0 16px 20px;
}

.classroom-card {
  padding: 14px;
  border-radius: 18px;
  background: #fff;
  box-shadow: 0 10px 28px rgb(15 23 42 / 6%);
}

.classroom-card__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.classroom-card__title-wrap {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
}

.classroom-card__title {
  min-width: 0;
  color: #1f2329;
  font-size: 16px;
  font-weight: 700;
  line-height: 1.4;
}

.classroom-card__actions {
  display: flex;
  align-items: center;
}

.classroom-card__meta {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 12px;
}

.classroom-card__meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #6b7280;
  font-size: 13px;
  line-height: 1.5;
}

.classroom-card__remark {
  margin-top: 10px;
  padding: 10px 12px;
  border-radius: 12px;
  background: #f8fafc;
  color: #6b7280;
  font-size: 12px;
  line-height: 1.6;
}

.classroom-card__foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid #f0f2f5;
}

.classroom-card__sort {
  color: #9aa4b2;
  font-size: 12px;
}

.classroom-empty {
  padding: 40px 16px 20px;
}

.classroom-drawer {
  display: flex;
  height: 100%;
  flex-direction: column;
  background: #f6f7f8;
}

.classroom-drawer__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: #fff;
}

.classroom-drawer__title {
  color: #1f2329;
  font-size: 18px;
  font-weight: 700;
}

.classroom-drawer__close {
  cursor: pointer;
  color: #8c8c8c;
  font-size: 18px;
}

.classroom-drawer__body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 16px 24px;
}

.classroom-form-block {
  margin-bottom: 12px;
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

.classroom-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.classroom-form-block__control {
  width: 100%;
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

.classroom-drawer__footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 14px 16px calc(env(safe-area-inset-bottom) + 14px);
  border-top: 1px solid #eef2f6;
  background: #fff;
}

:deep(.classroom-setting-drawer .ant-drawer-content-wrapper) {
  height: 86% !important;
}

:deep(.classroom-setting-drawer .ant-drawer-content) {
  border-radius: 20px 20px 0 0;
  overflow: hidden;
}

:deep(.classroom-setting-drawer .ant-drawer-body) {
  padding: 0;
}
</style>
