<script setup>
import {
  DeleteOutlined,
  FormOutlined,
  MinusCircleFilled,
  PlusCircleFilled,
} from '@ant-design/icons-vue'
import { onActivated, onMounted, ref } from 'vue'
import AddStudentCodeModal from './add-student-code-modal.vue'
import messageService from '~@/utils/messageService'
import { addStuCustomFieldApi, deleteStuCustomFieldApi, getStuCustomFieldApi, getStuDefaultFieldApi, updateStuCustomFieldApi, updateStuDisplayStatusApi } from '~@/api/edu-center/student-list'

const systemDefaultIsDisplayList = ref([])
const customIsDisplayList = ref([])
const systemDefaultIsDisplayNoneList = ref([])
const customIsDisplayNoneList = ref([])
const openModal = ref(false)
const loading = ref(true)
const isEditMode = ref(false)
const currentEditItem = ref(null)
const isSystemDefault = ref(false)

function addStudentCodeModal() {
  isEditMode.value = false
  currentEditItem.value = null
  isSystemDefault.value = false
  openModal.value = true
}

// 编辑学员属性
function editStudentCode(item, isSystem) {
  isEditMode.value = true
  isSystemDefault.value = isSystem
  currentEditItem.value = { ...item }
  openModal.value = true
}

// 获取所有学员字段（系统默认和自定义）
async function getAllStuFields() {
  try {
    loading.value = true

    // 获取系统默认字段
    const defaultRes = await getStuDefaultFieldApi()
    systemDefaultIsDisplayList.value = defaultRes.result.filter(item => item.isDisplay)
    systemDefaultIsDisplayNoneList.value = defaultRes.result.filter(item => !item.isDisplay)

    // 获取自定义字段
    const customRes = await getStuCustomFieldApi()
    customIsDisplayList.value = customRes.result.filter(item => item.isDisplay)
    customIsDisplayNoneList.value = customRes.result.filter(item => !item.isDisplay)
  }
  finally {
    checkLoading()
  }
}

// 更新学员属性显示状态
async function updateStuDisplayStatus(item) {
  loading.value = true
  try {
    item.isDisplay = !item.isDisplay
    await updateStuDisplayStatusApi(item)
    getAllStuFields()
  }
  catch (error) {
    console.log(error)
  }
  finally {
  }
}
// 新增或更新学员属性
async function handleSubmitStuCode(item) {
  try {
    if (isEditMode.value) {
      await updateStuCustomFieldApi(item)
      messageService.success('更新学员属性成功')
    }
    else {
      await addStuCustomFieldApi(item)
      messageService.success('新增学员属性成功')
    }
    getAllStuFields()
    openModal.value = false
  }
  catch (error) {
    console.log(error)
  }
  finally {
    checkLoading()
  }
}
// 删除自定义学员属性
async function deleteStuCustomField(item) {
  try {
    loading.value = true
    await deleteStuCustomFieldApi(item)
    messageService.success('删除学员属性成功')
    getAllStuFields()
  }
  catch (error) {
    console.log(error)
  }
}
// 检查loading状态
function checkLoading() {
  if (systemDefaultIsDisplayList.value.length > 0 || customIsDisplayList.value.length > 0) {
    loading.value = false
  }
}

onMounted(() => {
  getAllStuFields()
})

onActivated(() => {
  getAllStuFields()
})
</script>

<template>
  <div>
    <a-spin :spinning="loading">
      <div class="student-code bg-white p-4 rounded-b-4">
        <div class="flex flex-items-center justify-between">
          <div class="flex ">
            <custom-title
              :title="`已选属性 (${systemDefaultIsDisplayList.length + customIsDisplayList.length})`"
              font-size="18px" font-weight="600"
            /> <span
              class="mt-1 flex flex-items-center text-3 ml-4 text-#888"
            >带 <span
              class="text-red text-4 ml-1 mr-1 mt-1 flex"
            >*</span>
              为学员必填属性</span>
          </div>
          <div>
            <a-button class="mr-4">
              自动升年级
            </a-button>
            <a-button type="primary" class="font-800" @click="addStudentCodeModal">
              新增学员属性
            </a-button>
          </div>
        </div>
        <div class="content pl-2.5 mt-2 mb-4">
          <div class="subTitle text-3.5 text-#222 font-500 mb-3">
            系统默认（{{ systemDefaultIsDisplayList.length }}）
          </div>
          <a-space v-if="loading" :size="12" class="flex flex-wrap">
            <div v-for="i in 11" :key="i" class="item bg-#f0f0f0 w-60 h-10 rounded-1.5 skeleton-item" />
          </a-space>
          <a-space v-else :size="12" class="flex flex-wrap">
            <div
              v-for="(item, index) in systemDefaultIsDisplayList" :key="index"
              class="item bg-#f6f7f8 w-60 h-10 rounded-1.5 text-3.5 text-#222 flex flex-items-center pl-3 pr-3 justify-between"
            >
              <div class="flex-center">
                <span :class="item.required ? 'text-red' : 'text-#f6f7f8'" class="text-5 mt-1.5 mr-1">*</span><span>{{
                  item.fieldKey }}</span>
              </div>
              <div>
                <FormOutlined v-if="item.canEdit" class="icon cursor-pointer" @click="editStudentCode(item, true)" />
                <MinusCircleFilled
                  v-if="item.canDelete" class="ml-3 cursor-pointer"
                  style="color: red;" @click="updateStuDisplayStatus(item)"
                />
              </div>
            </div>
          </a-space>
        </div>
        <div v-if="customIsDisplayList.length > 0" class="content pl-2.5 mt-2 mb-4">
          <div class="subTitle text-3.5 text-#222 font-500 mb-3">
            自定义（{{ customIsDisplayList.length }}）
          </div>
          <a-space v-if="loading" :size="12" class="flex flex-wrap">
            <div v-for="i in 3" :key="i" class="item bg-#f0f0f0 w-60 h-10 rounded-1.5 skeleton-item" />
          </a-space>
          <a-space v-else :size="12" class="flex flex-wrap">
            <div
              v-for="(item, index) in customIsDisplayList" :key="index"
              class="item bg-#f6f7f8 w-60 h-10 rounded-1.5 text-3.5 text-#222 flex flex-items-center pl-3 pr-3 justify-between"
            >
              <div class="flex-center">
                <span :class="item.required ? 'text-red' : 'text-#f6f7f8'" class="text-5 mt-1.5 mr-1">*</span><span class="flex">
                  <clamped-text :text="item.fieldKey" :lines="1" class="w-60% line-clamp-1" />
                  <span class="text-#ccc whitespace-nowrap">( {{ item.fieldType === 1 ? '文本' : item.fieldType === 2
                    ? '数字' : item.fieldType === 3 ? '日期' : '选项' }} )</span></span>
              </div>
              <div class="flex">
                <FormOutlined v-if="item.canEdit" class="icon cursor-pointer" @click="editStudentCode(item, false)" />
                <MinusCircleFilled
                  v-if="item.canDelete" class="ml-3 cursor-pointer"
                  style="color: red;" @click="updateStuDisplayStatus(item)"
                />
              </div>
            </div>
          </a-space>
        </div>
      </div>
      <div class="student-code bg-white p-4 rounded-4 mt-4">
        <div class="flex flex-items-center justify-between">
          <div class="flex ">
            <custom-title
              :title="`未选属性 (${systemDefaultIsDisplayNoneList.length + customIsDisplayNoneList.length})`"
              font-size="18px" font-weight="600"
            /> <span
              class="mt-1 flex flex-items-center text-3 ml-4 text-#888"
            >带 <span
              class="text-red text-4 ml-1 mr-1 mt-1 flex"
            >*</span>
              为学员必填属性</span>
          </div>
        </div>
        <div v-if="systemDefaultIsDisplayNoneList.length > 0" class="content pl-2.5 mt-2 mb-4">
          <div class="subTitle text-3.5 text-#222 font-500 mb-3">
            系统默认（{{ systemDefaultIsDisplayNoneList.length }}）
          </div>
          <a-space v-if="loading" :size="12" class="flex flex-wrap">
            <div v-for="i in 4" :key="i" class="item bg-#f0f0f0 w-60 h-10 rounded-1.5 skeleton-item" />
          </a-space>
          <a-space v-else :size="12" class="flex flex-wrap">
            <div
              v-for="(item, index) in systemDefaultIsDisplayNoneList" :key="index"
              class="item bg-#f6f7f8 w-60 h-10 rounded-1.5 text-3.5 text-#222 flex flex-items-center pl-3 pr-3 justify-between"
            >
              <div class="flex-center">
                <span :class="item.required ? 'text-red' : 'text-#f6f7f8'" class="text-5 mt-1.5 mr-1">*</span><span>{{
                  item.fieldKey }}</span>
              </div>
              <div>
                <FormOutlined v-if="item.canEdit" class="icon cursor-pointer" @click="editStudentCode(item, true)" />
                <PlusCircleFilled
                  v-if="item.canDelete" class="ml-3 cursor-pointer icon"
                  @click="updateStuDisplayStatus(item)"
                />
              </div>
            </div>
          </a-space>
        </div>
        <div v-if="customIsDisplayNoneList.length > 0" class="content pl-2.5 mt-2 mb-4">
          <div class="subTitle text-3.5 text-#222 font-500 mb-3">
            自定义（{{ customIsDisplayNoneList.length }}）
          </div>
          <a-space v-if="loading" :size="12" class="flex flex-wrap">
            <div v-for="i in 4" :key="i" class="item bg-#f0f0f0 w-60 h-10 rounded-1.5 skeleton-item" />
          </a-space>
          <a-space v-else :size="12" class="flex flex-wrap">
            <div
              v-for="(item, index) in customIsDisplayNoneList" :key="index"
              class="item bg-#f6f7f8 w-60 h-10 rounded-1.5 text-3.5 text-#222 flex flex-items-center pl-3 pr-3 justify-between"
            >
              <div class="flex-center">
                <span :class="item.required ? 'text-red' : 'text-#f6f7f8'" class="text-5 mt-1.5 mr-1">*</span><span class="flex">
                  <clamped-text :text="item.fieldKey" :lines="1" class="w-60% line-clamp-1" />
                  <span class="text-#ccc whitespace-nowrap">( {{ item.fieldType === 1 ? '文本' : item.fieldType === 2
                    ? '数字' : item.fieldType === 3 ? '日期' : '选项' }} )</span> </span>
              </div>
              <div class="flex">
                <a-popconfirm
                  title="删除后不可撤销，请谨慎操作?" ok-text="确定" cancel-text="取消"
                  @confirm="deleteStuCustomField(item)"
                >
                  <DeleteOutlined class="cursor-pointer" style="color:red;" />
                </a-popconfirm>
                <FormOutlined v-if="item.canEdit" class="icon cursor-pointer mx-3" @click="editStudentCode(item, false)" />
                <PlusCircleFilled
                  v-if="item.canDelete" class="cursor-pointer icon"
                  @click="updateStuDisplayStatus(item)"
                />
              </div>
            </div>
          </a-space>
        </div>
      </div>
    </a-spin>
    <AddStudentCodeModal v-model:open="openModal" :edit-mode="isEditMode" :edit-data="currentEditItem" :is-system-default="isSystemDefault" @handle-submit-stu-code="handleSubmitStuCode" />
  </div>
</template>

<style lang="less" scoped>
.icon {
  color: var(--pro-ant-color-primary);
}

.skeleton-item {
  position: relative;
  overflow: hidden;

  &::after {
    content: '';
    position: absolute;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    transform: translateX(-100%);
    background-image: linear-gradient(90deg,
        rgba(255, 255, 255, 0) 0,
        rgba(255, 255, 255, 0.2) 20%,
        rgba(255, 255, 255, 0.5) 60%,
        rgba(255, 255, 255, 0));
    animation: shimmer 1.5s infinite;
  }
}

@keyframes shimmer {
  100% {
    transform: translateX(100%);
  }
}
</style>
