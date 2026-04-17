<script setup>
import { CloseOutlined, DeleteOutlined, FormOutlined, InfoCircleOutlined, MinusCircleOutlined, PlusCircleOutlined, UnorderedListOutlined } from '@ant-design/icons-vue'
import { useThrottleFn } from '@vueuse/core'
import { addStuCustomFieldApi, deleteStuCustomFieldApi, getStuCustomFieldApi, getStuDefaultFieldApi, updateStuCustomFieldApi, updateStuDisplayStatusApi } from '@/api/edu-center/student-list'
import { messageService } from '~@/utils/messageService'

const activeKey = ref()
const scroll = ref()
const studentAttributeDrawerVisible = ref(false)
const fieldItem = ref({})

const fieldFrom = ref({
  fieldKey: undefined,
  fieldType: 1,
  required: false,
  searched: false,
  optionsJson: '',
})

const options = ref([
  {
    label: '选项名称',
    value: '',
  },
  {
    label: '选项名称',
    value: '',
  },
  {
    label: '选项名称',
    value: '',
  },
])

// ... existing code ...

// 拖拽相关变量
const dragIndex = ref(-1)
const dragOverIndex = ref(-1)
const optionsContainer = ref(null)

// 开始拖拽时触发
function handleDragStart(index) {
  dragIndex.value = index
}

// 拖拽经过其他元素时触发
function handleDragOver(index) {

}

// 放置拖拽元素时触发
function handleDrop(index) {
  if (dragIndex.value === index)
    return

  // 更新数组顺序
  const draggedItem = options.value.splice(dragIndex.value, 1)[0]
  options.value.splice(index, 0, draggedItem)

  // 重置拖拽索引
  dragIndex.value = -1
  dragOverIndex.value = -1
}

// 拖拽结束时触发（可选，用于清理状态）
function handleDragEnd() {
  // 重置拖拽索引
  dragIndex.value = -1
  dragOverIndex.value = -1
}
// ... existing code ...

// 系统默认属性
const defaultField = ref([])
// 自定义属性
const customField = ref([])
// 选中的系统默认属性
const selectDefaultField = computed(() => {
  return defaultField.value.filter(item => item.isDisplay)
})
// 选中的自定义属性
const selectCustomField = computed(() => {
  return customField.value.filter(item => item.isDisplay)
})

// 可选系统默认属性
const otherDefaultField = computed(() => {
  return defaultField.value.filter(item => !item.isDisplay)
})

// 可选的自定义属性
const otherCustomField = computed(() => {
  return customField.value.filter(item => !item.isDisplay)
})

onMounted(() => {
  init()
})

const editDefaultField = useThrottleFn((item, type) => {
  const field = [...defaultField.value, ...customField.value].find(i => i.id === item.id)
  if (type === 'isDisplay') {
    updateStuDisplayStatusApi({ ...field, isDisplay: !field.isDisplay }).finally(() => {
      init()
    })
  }
}, 1000)

function addOptions() {
  if (options.value.length >= 40) {
    return
  }
  options.value.push({
    label: '选项名称',
    value: '',
  })
}

function editField(item) {
  const sub = JSON.parse(JSON.stringify(item))
  fieldItem.value = sub
  const { id, uuid, version, ...rest } = sub
  fieldFrom.value = { ...rest }
  if (sub.optionsJson) {
    options.value = item.optionsJson.split(',').map((item, index) => {
      return {
        label: `选项名称`,
        value: item,
      }
    })
  }
  studentAttributeDrawerVisible.value = true
}
// 新增自定义属性
function addStuCustomField() {
  // 名称非空校验
  if (!fieldFrom.value.fieldKey?.trim()) {
    messageService.error('请输入属性名称')
    return
  }

  // 选项类型校验
  if (fieldFrom.value.fieldType === 4) {
    // 检查是否有选项
    if (options.value.length === 0) {
      messageService.error('请至少添加一个选项')
      return
    }

    // 检查选项值是否都已填写
    const emptyOptions = options.value.filter(item => !item.value?.trim())
    if (emptyOptions.length > 0) {
      messageService.error('请填写所有选项的名称')
      return
    }

    // 检查选项值是否重复
    const values = options.value.map(item => item.value.trim())
    if (new Set(values).size !== values.length) {
      messageService.error('选项名称不能重复')
      return
    }
  }

  fieldFrom.value.optionsJson = options.value.map(item => item.value.trim()).join(',')
  const isEdit = Boolean(fieldItem.value.id)
  const method = isEdit ? updateStuCustomFieldApi : addStuCustomFieldApi
  method({ ...fieldItem.value, ...fieldFrom.value }).then(() => {
    init()
    studentAttributeDrawerVisible.value = false
    // 重置表单
    resetForm()
    messageService.success(isEdit ? '编辑成功' : '添加成功')
  })
}

// 重置表单方法
function resetForm() {
  fieldFrom.value.fieldKey = undefined
  fieldFrom.value.fieldType = 1
  fieldFrom.value.required = false
  fieldFrom.value.searched = false
  fieldFrom.value.optionsJson = ''
  options.value = [
    {
      label: '选项名称',
      value: '',
    },
    {
      label: '选项名称',
      value: '',
    },
    {
      label: '选项名称',
      value: '',
    },
  ]
  fieldItem.value = {}
}

function closeDrawer() {
  studentAttributeDrawerVisible.value = false
  fieldItem.value = {}
  fieldFrom.value.fieldKey = undefined
  fieldFrom.value.fieldType = 1
  fieldFrom.value.required = false
  fieldFrom.value.searched = false
  fieldFrom.value.optionsJson = ''
  options.value = [
    {
      label: '选项名称',
      value: '',
    },
    {
      label: '选项名称',
      value: '',
    },
    {
      label: '选项名称',
      value: '',
    },
  ]
}

// 删除自定义属性
function deleteStuCustomField(item) {
  deleteStuCustomFieldApi(item).then(() => {
    init()
    messageService.success('删除成功')
  })
}

// 获取系统默认属性和自定义属性
function init() {
  getStuDefaultFieldApi().then(({ result }) => {
    defaultField.value = result
  })
  getStuCustomFieldApi().then(({ result }) => {
    customField.value = result
  })
}

// 监听抽屉显示状态
watch(studentAttributeDrawerVisible, (val) => {
  if (val) {
    // 打开抽屉时禁用父级滚动
    document.body.style.overflow = 'hidden'
  }
  else {
    // 关闭抽屉时恢复父级滚动
    document.body.style.overflow = ''
  }
})
</script>

<template>
  <div ref="scroll" class="student-attribute-settings scrollbar" :class="{ 'overflow-hidden': studentAttributeDrawerVisible }">
    <a-affix :target="() => scroll">
      <div class="bg-#e6f0ff p-16px flex items-center">
        <InfoCircleOutlined class="text-#06f mr-8px text-14px" />
        <span class="text-#06f text-14px cursor-pointer">调整后，学员属性相关页即时生效</span>
      </div>
      <div class="bg-#f6f7f8 px-16px py-10px">
        <span class="text-#999 text-12px cursor-pointer"> 已选属性 {{ selectDefaultField.length }} 个（带
          <span class="text-red-500">*</span> 为学员必填属性）
        </span>
      </div>
    </a-affix>
    <div class="flex items-center bg-white">
      <a-collapse v-model:active-key="activeKey" ghost class="flex-1">
        <a-collapse-panel key="1">
          <template #header>
            <span class="text-14px text-#666">系统默认属性（{{ selectDefaultField.length }}个）</span>
          </template>
          <div>
            <div
              v-for="(item, index) in selectDefaultField" :key="index"
              class="flex items-center justify-between border-t-1px border-b-0px border-r-0px border border-#eee border-solid px-16px py-12px"
            >
              <div class="flex items-center">
                <div
                  class="cursor-pointer" :class="{ 'opacity-0': !item.canDelete }"
                  @click="editDefaultField(item, 'isDisplay')"
                >
                  <MinusCircleOutlined class="text-red-400" />
                </div>
                <div :class="{ 'opacity-0': !item.required }" class="mx-5px">
                  <span class="text-red-500">*</span>
                </div>
                <div>{{ item.fieldKey }}</div>
              </div>
              <div class="cursor-pointer" :class="{ 'opacity-0': !item.canEdit }" @click="editField(item)">
                <FormOutlined class="text-#06f" />
              </div>
            </div>
          </div>
        </a-collapse-panel>
      </a-collapse>
    </div>

    <!-- 自定选择义属性 -->
    <div v-if="selectCustomField.length" class="bg-white m-t-10px">
      <div
        v-for="(item, index) in selectCustomField" :key="index"
        class="flex items-center justify-between border-t-1px border-b-0px border-r-0px border border-#eee border-solid px-16px py-12px"
      >
        <div class="flex items-center">
          <div
            class="cursor-pointer" :class="{ 'opacity-0': !item.canDelete }"
            @click="editDefaultField(item, 'isDisplay')"
          >
            <MinusCircleOutlined class="text-red-400" />
          </div>
          <div :class="{ 'opacity-0': !item.required }" class="mx-5px">
            <span class="text-red-500">*</span>
          </div>
          <div>{{ item.fieldKey }}</div>
        </div>
        <div class="flex items-center gap-10px">
          <!-- <DeleteOutlined :class="{ 'opacity-0': !item.canDelete }" class="text-red-400" /> -->
          <FormOutlined :class="{ 'opacity-0': !item.canEdit }" class="text-#06f" @click="editField(item)" />
        </div>
      </div>
    </div>

    <!-- 系统默认属性 -->
    <template v-if="otherDefaultField.length + otherCustomField.length">
      <div class="bg-#f6f7f8 px-16px py-10px">
        <span class="text-#999 text-12px cursor-pointer"> 可选属性 {{ otherDefaultField.length + otherCustomField.length }}
          个</span>
      </div>
      <div v-if="otherDefaultField.length" class="bg-white">
        <div
          v-for="(item, index) in otherDefaultField" :key="index"
          class="flex items-center justify-between border-t-1px border-b-0px border-r-0px border border-#eee border-solid px-16px py-12px"
        >
          <div class="flex items-center">
            <div
              class="cursor-pointer" :class="{ 'opacity-0': !item.canDelete }"
              @click="editDefaultField(item, 'isDisplay')"
            >
              <PlusCircleOutlined class="text-#06f" />
            </div>
            <div :class="{ 'opacity-0': !item.required }" class="mx-5px">
              <span class="text-red-500">*</span>
            </div>
            <div>{{ item.fieldKey }}</div>
          </div>
          <div class="flex items-center gap-10px">
            <!-- <DeleteOutlined :class="{ 'opacity-0': !item.canDelete }" class="text-red-400" /> -->
            <FormOutlined :class="{ 'opacity-0': !item.canEdit }" class="text-#06f" @click="editField(item)" />
          </div>
        </div>
      </div>
    </template>
    <!-- 自定义属性 -->
    <template v-if="otherCustomField.length">
      <div class="bg-white">
        <div
          v-for="(item, index) in otherCustomField" :key="index"
          class="flex items-center justify-between border-t-1px border-b-0px border-r-0px border border-#eee border-solid px-16px py-12px"
        >
          <div class="flex items-center">
            <div
              class="cursor-pointer" :class="{ 'opacity-0': !item.canDelete }"
              @click="editDefaultField(item, 'isDisplay')"
            >
              <PlusCircleOutlined class="text-#06f" />
            </div>
            <div :class="{ 'opacity-0': !item.required }" class="mx-5px">
              <span class="text-red-500">*</span>
            </div>
            <div>{{ item.fieldKey }}</div>
          </div>
          <div class="flex items-center gap-10px">
            <a-popconfirm title="删除后不可撤销，请谨慎操作" ok-text="删除" cancel-text="再想想" @confirm="deleteStuCustomField(item)">
              <DeleteOutlined :class="{ 'opacity-0': !item.canDelete }" class="text-red-400 cursor-pointer" />
            </a-popconfirm>

            <FormOutlined
              :class="{ 'opacity-0': !item.canEdit }" class="text-#06f cursor-pointer"
              @click="editField(item)"
            />
          </div>
        </div>
      </div>
    </template>
    <div class="p-16px position-absolute bottom-0 left-0 right-0 bottom-0 bg-white mt-10px">
      <a-button
        class="w-100% h-34px rounded-15px bg-#06f text-white" type="primary"
        @click="() => studentAttributeDrawerVisible = true"
      >
        新增学员属性
      </a-button>
    </div>
    <!-- 编辑学员属性弹窗 -->
    <a-drawer
      root-class-name="student-drawer" :closable="false" :mask="false" :mask-closable="false" placement="bottom"
      :open="studentAttributeDrawerVisible" :get-container="false"
    >
      <div class="adjust-form scrollbar h-100% overflow-y-auto flex flex-col">
        <div class="drawer-header bg-white rounded-lt-16px rounded-rt-16px">
          <span class="title">{{ fieldItem.id ? '编辑' : '新增' }}学员属性</span>
          <span class="close-icon" @click="closeDrawer">
            <CloseOutlined />
          </span>
        </div>
        <div class="pt-15px flex-1 flex flex-col justify-between">
          <div class="bg-#f6f7f8 flex-1">
            <div class="mb-10px bg-#fff h-50px px-5px">
              <a-input
                v-model:value="fieldFrom.fieldKey" :disabled="fieldItem.isDefault"
                class="h-34px text-18px font-bold" :bordered="false" placeholder="名称，最多20字" :maxlength="20"
              />
            </div>

            <div class="">
              <div class="px-15px text-#666 text-12px mb-5px">
                格式类型
              </div>
              <div class="p-15px flex gap-10px bg-#fff ">
                <template v-for="(item, index) in ['文本', '数字', '日期', '选项']" :key="index">
                  <template v-if="fieldItem.id">
                    <a-button
                      v-if="fieldFrom.fieldType === index + 1"
                      :class="{ 'bg-#e6f0ff!important text-#06f': fieldFrom.fieldType === index + 1 }"
                      class="w-25% rounded-15px border-none bg-#f6f7f8" @click="() => fieldFrom.fieldType = index + 1"
                    >
                      {{ item }}
                    </a-button>
                  </template>
                  <template v-else>
                    <a-button
                      :class="{ 'bg-#e6f0ff!important text-#06f': fieldFrom.fieldType === index + 1 }"
                      class="w-25% rounded-15px border-none bg-#f6f7f8" @click="() => fieldFrom.fieldType = index + 1"
                    >
                      {{ item }}
                    </a-button>
                  </template>
                </template>
              </div>
            </div>

            <div class="mb-10px bg-#fff px-15px ">
              <div class="text-#999 mb-5px py-8px  text-13px border-0px border-t-1px border-solid border-#eee">
                <span v-if="fieldFrom.fieldType === 1">文本内容，限 100 字</span>
                <span v-if="fieldFrom.fieldType === 2">仅限数字</span>
                <span v-if="fieldFrom.fieldType === 3">选择格式：年-月-日</span>
                <div v-if="fieldFrom.fieldType === 4">
                  <div ref="optionsContainer">
                    <div
                      v-for="(item, index) in options" :key="index" draggable="true"
                      class="py-4px flex items-center gap-8px mb-8px border-0px border-b-1px border-solid border-#eee"
                      @dragstart="handleDragStart(index)" @dragover.prevent="handleDragOver(index)" @dragenter.prevent
                      @drop="handleDrop(index)"
                    >
                      <div class="flex items-center justify-center gap-8px">
                        <MinusCircleOutlined
                          class="text-14px text-red-400 cursor-pointer"
                          @click="() => options.splice(index, 1)"
                        />
                        <span>{{ item.label }}</span>
                      </div>
                      <a-input
                        v-model:value="item.value" :bordered="false" class="flex-1" placeholder="最多20字"
                        :maxlength="20"
                      />
                      <div class="flex items-center justify-center cursor-move drag-handle">
                        <UnorderedListOutlined class="text-#999" />
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center text-#06f cursor-pointer mt-10px" @click="addOptions">
                    <PlusCircleOutlined class="mr-5px text-14px" />
                    <span>添加({{ options.length }}/40)</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="bg-#fff px-15px">
              <div class="flex items-center justify-between mb-10px py-10px">
                <div class="flex items-center gap-5px">
                  <span class="text-#666">是否必填</span>
                  <a-tooltip>
                    <template #title>
                      开启必填后，报名创建/编辑学员、填写招生表单等相关学员属性填写业务时必填，否则无法保存
                    </template>
                    <InfoCircleOutlined class="text-#06f text-14px" />
                  </a-tooltip>
                </div>
                <a-switch v-model:checked="fieldFrom.required" />
              </div>

              <div
                class="flex items-center justify-between mb-10px py-10px border-0px border-t-1px border-solid border-#eee"
              >
                <div class="flex items-center gap-5px">
                  <span class="text-#666">支持搜索</span>
                  <a-tooltip>
                    <template #title>
                      开启支持筛选/搜索后，学员管理的"在读学员""意向学员"等相关页面将支持此属性的筛选/搜索功能
                    </template>
                    <InfoCircleOutlined class="text-#06f text-14px" />
                  </a-tooltip>
                </div>
                <a-switch v-model:checked="fieldFrom.searched" />
              </div>
            </div>
          </div>
          <div class="w-100% px-20px bg-#fff flex items-center justify-center">
            <a-button class="w-100% h-34px rounded-15px" type="primary" @click="addStuCustomField">
              保存并选择
            </a-button>
          </div>
        </div>
      </div>
    </a-drawer>
  </div>
</template>

<style lang="less" scoped>
.student-attribute-settings {
  background-color: #f6f7f8;
  height: 100%;
  overflow-y: auto;
  position: relative;

  &.overflow-hidden {
    overflow: hidden !important;
  }
}

:deep(.ant-collapse-content-box) {
  padding: 0px !important;
}

.student-drawer {
  .drawer-header {
    position: sticky;
    top: 0;
    z-index: 10;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;
    background: #fff;

    .title {
      font-size: 16px;
      font-weight: 500;
      color: #333;
    }

    .close-icon {
      cursor: pointer;
      color: #999;
      font-size: 16px;
    }
  }
}

.p-16px.position-absolute {
  position: sticky;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 100;
  background: #fff;
  margin-top: auto;
  border-top: 1px solid #f0f0f0;
}

::v-deep(.ant-drawer-content-wrapper) {
  height: 95% !important;
}

::v-deep(.ant-drawer-header) {
  display: none;
}

::v-deep(.ant-drawer-body) {
  padding: 0 !important;
  height: 100%;
  overflow: hidden;
}

::v-deep(.ant-drawer-content) {
  border-radius: 16px 16px 0 0;
  overflow: hidden;
  height: 100%;
}

::v-deep(.ant-drawer-open) {
  .student-attribute-settings {
    overflow: hidden;
  }
}

.w-100\%.px-20px {
  position: sticky;
  bottom: 0;
  background: #fff;
  padding: 16px 20px;
  border-top: 1px solid #f0f0f0;
  z-index: 10;
}
</style>
