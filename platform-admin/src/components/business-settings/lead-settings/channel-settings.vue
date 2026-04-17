<script setup>
import {
  CheckOutlined,
  CloseOutlined,
  DownOutlined,
  EllipsisOutlined,
  PlusCircleOutlined,
  PlusOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons-vue'
import { messageService } from '@/utils/messageService'

import { adjustChannelApi, createChannelApi, deleteCategoryApi, getChannelListWithChannelsApi, getDefaultChannelList, saveCategoryApi, updateCategoryApi, updateChannelApi, updateChannelStatusApi } from '~@/api/enroll-center/intention-student'

// 系统默认渠道列表
const defaultChannelList = ref([])
// 渠道列表
const channelList = ref([])
// 分类列表
const categoryList = ref([])
// 创建渠道弹窗
const createChannelOpen = ref(false)

// 在变量定义区域添加（约在第20行附近）
// 调整弹窗状态
const adjustOpen = ref(false)
// 选择的分类
const selectedCategory = ref(null)

// 批量调整弹窗状态
const batchAdjustOpen = ref(false)
// 选中的渠道列表
const selectedChannels = ref([])

onMounted(() => {
  init()
})

const activeKey = ref([])

// 抽屉状态
const open = ref(false)
// modal状态
const modalOpen = ref(false)
// 抽屉类型
const drawerType = ref('')

// 分类名称
const categoryName = ref('')
const categoryItem = ref({})

// 渠道名称
const channelName = ref('')
const channelItem = ref({})
// 渠道备注
const channelRemark = ref('')

function init() {
  // 获取系统默认渠道列表
  getDefaultChannelList().then(({ result }) => {
    defaultChannelList.value = result
  })
  // 获取分类和渠道列表
  getChannelListWithChannelsApi().then(({ result }) => {
    categoryList.value = result.filter(item => item.type === 1).map(item => ({ ...item, active: [''] }))
    channelList.value = result.filter(item => item.type === 0)
  })
}

const customizeNum = computed(() => {
  const num = categoryList.value.reduce((pre, cur) => {
    return pre + cur.channelList.length
  }, 0)
  return num + channelList.value.length
})

function showDrawer(type, item) {
  open.value = true
  drawerType.value = type
  if (type === '分类') {
    categoryItem.value = item
  }

  if (type === '渠道') {
    channelItem.value = item
  }
}

function onClose() {
  open.value = false
  channelItem.value = {}
  categoryItem.value = {}
  drawerType.value = ''
}

function closeCreateChannel() {
  createChannelOpen.value = false
  channelName.value = ''
  channelRemark.value = ''
}

// 打开新增弹窗
function handleCreateCategory() {
  open.value = false
  modalOpen.value = true
}

function handleCreateChannel(item) {
  categoryItem.value = item
  open.value = false
  createChannelOpen.value = true
}

function modalCancel() {
  modalOpen.value = false
  categoryName.value = ''
}
// 新增分类或编辑
function handleSaveCategory() {
  if (!categoryName.value || categoryName.value.trim() === '') {
    messageService.error('分类名称不能为空')
    return
  }

  if (categoryItem.value.id) {
    updateCategoryApi({ ...categoryItem.value, categoryName: categoryName.value.trim() }).then(({ result }) => {
      modalOpen.value = false
      categoryName.value = ''
      categoryItem.value = {}
    }).finally(() => {
      init()
    })
  }
  else {
    saveCategoryApi({ categoryName: categoryName.value.trim() }).then(({ result }) => {
      modalOpen.value = false
      categoryName.value = ''
    }).finally(() => {
      init()
    })
  }
}
// 新增渠道
function submitCreateChannel() {
  if (!channelName.value || channelName.value.trim() === '') {
    messageService.error('渠道名称不能为空')
    return
  }

  if (channelItem.value.id) {
    // 编辑渠道
    updateChannelApi({
      id: channelItem.value.id,
      uuid: channelItem.value.uuid,
      version: channelItem.value.version,
      channelName: channelName.value.trim(),
      remark: channelRemark.value,
    }).then(({ result }) => {
      createChannelOpen.value = false
      channelName.value = ''
      channelRemark.value = ''
      channelItem.value = {}
      messageService.success('编辑成功')
    }).finally(() => {
      init()
    })
  }
  else {
    // 新增渠道
    createChannelApi({
      categoryId: categoryItem.value.id,
      channelName: channelName.value.trim(),
      remark: channelRemark.value,
    }).then(({ result }) => {
      createChannelOpen.value = false
      channelName.value = ''
      channelRemark.value = ''
      categoryItem.value = {}
      messageService.success('创建成功')
    }).finally(() => {
      init()
    })
  }
}

// 打开批量调整弹窗
function openBatchAdjust() {
  batchAdjustOpen.value = true
}

// 关闭批量调整弹窗
function closeBatchAdjust() {
  batchAdjustOpen.value = false
  selectedChannels.value = []
  selectedCategory.value = null
}

// 选择或取消选择渠道
function toggleChannelSelection(channelId) {
  const index = selectedChannels.value.indexOf(channelId)
  if (index === -1) {
    selectedChannels.value.push(channelId)
  }
  else {
    selectedChannels.value.splice(index, 1)
  }
}

const selectedChannelsLength = computed(() => {
  return [...categoryList.value.flatMap(item => item.channelList.map(channel => channel.id)), ...channelList.value.map(item => item.id)].length
})

function toggleAllSelection() {
  const len = selectedChannels.value.length
  if (len !== selectedChannelsLength.value) {
    selectedChannels.value = [...categoryList.value.flatMap(item => item.channelList.map(channel => channel.id)), ...channelList.value.map(item => item.id)]
  }
  else {
    selectedChannels.value = []
  }
}

// 添加 handleAdjust 函数（约在第150行附近，其他函数之后）
// 调整渠道分类
function handleAdjust() {
  open.value = false
  batchAdjustOpen.value = false
  adjustOpen.value = true
}

// 关闭调整弹窗
function closeAdjustDrawer() {
  adjustOpen.value = false
  selectedCategory.value = null
}

// 确认调整
function confirmAdjust() {
  if (!selectedCategory.value)
    return

  // 这里添加调整渠道分类的 API 调用
  adjustChannelApi({ channelIds: [channelItem.value.id, ...selectedChannels.value], categoryId: selectedCategory.value })
    .then(() => {
      adjustOpen.value = false
      selectedCategory.value = null
      selectedChannels.value = []
    })
    .finally(() => {
      init()
    })

  // 临时模拟 API 调用
  adjustOpen.value = false
  selectedCategory.value = null
  init()
}

// 删除分类
function handleDeleteCategory() {
  deleteCategoryApi(categoryItem.value).then(() => {
    categoryItem.value = {}
    open.value = false
    init()
  })
}

// 启用/禁用渠道
function handleDisable() {
  updateChannelStatusApi({ ...channelItem.value, isDisabled: !channelItem.value.isDisabled }).then(() => {
    open.value = false
    // 提示启用成功或者停用成功
    messageService.success(channelItem.value.isDisabled ? '启用成功' : '停用成功')
  }).finally(() => {
    init()
  })
}

// 编辑分类
function handleEditCategory() {
  categoryName.value = categoryItem.value.name
  open.value = false
  modalOpen.value = true
}

// 编辑渠道
function handleEditChannel() {
  channelName.value = channelItem.value.name
  channelRemark.value = channelItem.value.remark
  open.value = false
  createChannelOpen.value = true
}
</script>

<template>
  <div
    class="registration-settings"
    :class="{ 'overflow-hidden': batchAdjustOpen || open || createChannelOpen || adjustOpen }"
  >
    <!-- 渠道设置 -->
    <div class="py-10px px-15px text-12px text-#333">
      系统默认渠道，不支持删除编辑等操作
    </div>
    <div
      v-for="item in defaultChannelList" :key="item.id"
      class="flex items-center justify-between px-20px p-10px bg-#fff border-0px border-b-1px border-#f6f6f6 border-solid"
    >
      <div class="flex items-center gap-3px">
        <div class="text-14px text-#222">
          {{ item.name }}
        </div>
        <span class="text-12px text-#06f">
          <a-tooltip>
            <template #title>{{ item.remark }}</template>
            <QuestionCircleOutlined />
          </a-tooltip>
        </span>
      </div>
      <div class="rounded-8px px-8px py-2px bg-#f6f7f8 text-#888 text-12px">
        系统默认
      </div>
    </div>

    <div class="bg-#f6f7f8 px-15px py-5px flex items-center justify-between">
      <div class="flex items-center gap-3px">
        <span class="text-12px text-#333">当前共计 {{ customizeNum }} 个自定义渠道</span>
        <a-tooltip>
          <template #title>
            <p> 为了便于管理，可将具备某些共同属性的渠道归属于同一类别中，如"线上渠道"、"线下渠道"等 </p>
            <p> 渠道可以单独创建，也可归属于某个渠道分类中，如"儿童节银泰城地推"、"双十一异业合作"等 </p>
          </template>
          <QuestionCircleOutlined class="text-12px text-#06f" />
        </a-tooltip>
      </div>
      <a-button type="link" class="text-12px text-#06f " @click="openBatchAdjust">
        批量调整渠道
      </a-button>
    </div>

    <template v-for="item in categoryList" :key="item.id">
      <div class="flex justify-between items-center py-7px bg-white">
        <a-collapse v-model:active-key="item.active" ghost class="flex-1">
          <a-collapse-panel key="1" :show-arrow="false">
            <template #header>
              <div class="px-15px py-5px flex items-center justify-between">
                <div class="text-14px text-#666 flex items-center gap-10px">
                  <DownOutlined
                    :class="item.active.includes('1') ? '' : 'rotate-180'"
                    class="transition-transform duration-300 text-10px"
                  />
                  <span>分类：{{ item.name }}<span class="text-#ccc text-14px">（{{ item.channelList.length
                  }}）</span></span>
                </div>
                <div>
                  <EllipsisOutlined class="text-20px cursor-pointer" @click.stop="showDrawer('分类', item)" />
                </div>
              </div>
            </template>
            <div
              v-for="channel in item.channelList" :key="channel.id"
              class="px-15px pt-15px pb-5px flex items-center justify-between bg-white  py-12px mt-10px border-t-1px border-b-0 border-l-0 border-r-0 border-#f6f6f6 border-solid"
            >
              <div class="text-13px text-#333 flex items-center gap-10px pl-21px">
                <span>渠道：{{ channel.name }}</span>
              </div>
              <div class="flex items-center gap-10px">
                <div v-if="channel.isDisabled" class="text-#ff3333 bg-#ffe6e6 px-10px py-1px rounded-15px text-11px">
                  已停用
                </div>
                <EllipsisOutlined class="text-20px cursor-pointer" @click.stop="showDrawer('渠道', channel)" />
              </div>
            </div>
            <div
              class="ml-20px px-15px py-5px flex items-center justify-between bg-white  py-12px mt-10px border-t-1px border-b-0 border-l-0 border-r-0 border-#eee border-solid"
            >
              <div class="text-14px text-#333 flex items-center gap-5px ">
                <PlusCircleOutlined class="text-12px text-#06f" />
                <span type="text" class="text-12px cursor-pointer text-#06f" @click.stop="handleCreateChannel(item)">
                  新增渠道
                </span>
              </div>
            </div>
          </a-collapse-panel>
        </a-collapse>
      </div>
      <div class="h-6px bg-#f6f6f6" />
    </template>

    <template v-for="item in channelList" :key="item.id">
      <div class="mt-10px flex justify-between items-center  py-10px px-15px bg-white">
        <div class="text-14px text-#333 flex items-center gap-10px ">
          <span>渠道：{{ item.name }}</span>
        </div>
        <div class="flex items-center gap-10px">
          <div v-if="item.isDisabled" class="text-#ff3333 bg-#ffe6e6 px-10px py-1px rounded-15px text-11px">
            已停用
          </div>
          <EllipsisOutlined class="text-20px cursor-pointer" @click.stop="showDrawer('渠道', item)" />
        </div>
      </div>
      <div class="h-6px bg-#f6f6f6" />
    </template>

    <!-- 在组件末尾添加新增按钮 -->
    <div class="add-button" @click="showDrawer('新增')">
      <PlusOutlined />
    </div>
    <!-- 操作菜单 -->
    <a-drawer
      placement="bottom" :closable="false" :open="open" :get-container="false"
      @close="onClose"
    >
      <div class="channel-action-menu">
        <template v-if="drawerType === '渠道'">
          <div class="menu-item" @click="handleEditChannel">
            <span>编辑</span>
          </div>
          <div class="menu-item" @click="handleAdjust">
            <span>调整</span>
          </div>
          <div class="menu-item" :class="channelItem.isDisabled ? 'info' : 'danger'" @click="handleDisable">
            <span>{{ channelItem.isDisabled ? '启用' : '停用' }}</span>
          </div>
        </template>

        <template v-if="drawerType === '新增'">
          <div class="menu-item" @click="handleCreateCategory">
            <span>创建分类</span>
          </div>
          <div class="menu-item" @click="handleCreateChannel">
            <span>创建渠道</span>
          </div>
        </template>
        <template v-if="drawerType === '分类'">
          <div class="menu-item" @click="handleEditCategory">
            <span>编辑</span>
          </div>
          <div class="menu-item danger" @click="handleDeleteCategory">
            <span>删除</span>
          </div>
        </template>
        <div class="menu-item cancel" @click="onClose">
          <span>取消</span>
        </div>
      </div>
    </a-drawer>

    <!-- 创建分类 -->
    <a-modal v-model:open="modalOpen" :centered="true" :mask-closable="false" :closable="false" width="300px">
      <template #title>
        <div>创建分类</div>
      </template>
      <a-input v-model:value="categoryName" placeholder="请填写分类名称" />
      <template #footer>
        <div class="flex items-center justify-between">
          <a-button ghost class="flex-1" type="text" @click="modalCancel">
            取消
          </a-button>
          <div class="w-1px bg-gray-300 h-20px" />
          <a-button ghost class="flex-1 text-#06f" type="text" @click="handleSaveCategory">
            确定
          </a-button>
        </div>
      </template>
    </a-modal>

    <!-- 创建渠道 -->
    <a-drawer
      root-class-name="create-drawer" :closable="false" :mask-closable="false" placement="bottom"
      :open="createChannelOpen" :get-container="false"
    >
      <div class="create-channel-form">
        <div class="drawer-header">
          <span class="title">创建渠道</span>
          <span class="close-icon" @click="closeCreateChannel">
            <CloseOutlined />
          </span>
        </div>
        <div class="form-content">
          <div class="form-item">
            <div class="label">
              <span class="required">*</span>
              渠道名称
            </div>
            <a-input v-model:value="channelName" placeholder="请填写渠道名称" class="channel-input" />
          </div>
          <div class="form-item">
            <div class="label">
              备注
            </div>
            <a-textarea
              v-model:value="channelRemark" placeholder="点此输入备注，建议200字以内" class="channel-textarea"
              :rows="4"
            />
          </div>
          <a-button type="primary" class="submit-btn" @click="submitCreateChannel">
            确定
          </a-button>
        </div>
      </div>
    </a-drawer>

    <!-- 调整弹窗 -->
    <a-drawer
      root-class-name="adjust-drawer" :closable="false" :mask-closable="false" placement="bottom"
      :open="adjustOpen" :get-container="false"
    >
      <div class="adjust-form h-100% flex flex-col">
        <div class="drawer-header">
          <span class="title">请选择</span>
          <span class="close-icon" @click="closeAdjustDrawer">
            <CloseOutlined />
          </span>
        </div>
        <div class=" flex-1 flex flex-col justify-between">
          <div class="">
            <div class="px-15px  py-8px cursor-pointer" @click="selectedCategory = '0'">
              移出所有分类
              <CheckOutlined v-if="selectedCategory === '0'" class="text-#06f ml-5px text-18px" />
            </div>
            <div
              v-for="category in categoryList" :key="category.id"
              class="cursor-pointer px-15px py-8px border-t-1px border-b-0 border-l-0 border-r-0 border-#eee border-solid"
              @click="selectedCategory = category.id"
            >
              <span>分类：{{ category.name }}</span>
              <CheckOutlined v-if="selectedCategory === category.id" class="text-#06f ml-5px text-18px" />
            </div>
          </div>
          <div class="w-100% p-20px">
            <a-button class="w-100% h-34px rounded-15px" type="primary" @click="confirmAdjust">
              确认调整
            </a-button>
          </div>
        </div>
      </div>
    </a-drawer>

    <!-- 批量调整弹窗 -->
    <a-drawer
      root-class-name="batch-adjust-drawer" :closable="false" :mask-closable="false" placement="bottom"
      :open="batchAdjustOpen" :get-container="false"
    >
      <div class="batch-adjust-form h-100% flex flex-col">
        <div class="drawer-header position-sticky top-0 bg-#fff">
          <span class="title">批量调整渠道</span>
          <span class="close-icon" @click="closeBatchAdjust">
            <CloseOutlined />
          </span>
        </div>
        <div class="flex-1 flex flex-col justify-between">
          <div class="channel-list">
            <div v-for="category in categoryList" :key="category.id" class="category-section">
              <div
                v-for="channel in category.channelList" :key="channel.id"
                class="channel-item cursor-pointer px-15px py-8px border-t-1px border-b-0 border-l-0 border-r-0 border-#eee border-solid"
                @click="toggleChannelSelection(channel.id)"
              >
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-5px">
                    <div
                      class="flex items-center justify-center w-18px h-18px rounded-50% border-1px border-solid border-#ccc"
                    >
                      <CheckOutlined
                        :class="selectedChannels.includes(channel.id) ? 'opacity-100' : 'opacity-0'"
                        class="text-#06f text-13px"
                      />
                    </div>
                    <span>渠道：{{ channel.name }}</span>
                  </div>
                  <div class="flex items-center">
                    <div
                      v-if="channel.isDisabled"
                      class="text-#ff3333 bg-#ffe6e6 px-10px py-1px rounded-15px text-11px"
                    >
                      已停用
                    </div>
                  </div>
                </div>
              </div>
            </div>
            <div
              v-for="channel in channelList" :key="channel.id"
              class="channel-item cursor-pointer px-15px py-8px border-t-1px border-b-0 border-l-0 border-r-0 border-#eee border-solid"
              @click="toggleChannelSelection(channel.id)"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-5px">
                  <div
                    class="flex items-center justify-center w-18px h-18px rounded-50% border-1px border-solid border-#ccc"
                  >
                    <CheckOutlined
                      :class="selectedChannels.includes(channel.id) ? 'opacity-100' : 'opacity-0'"
                      class="text-#06f text-13px"
                    />
                  </div>
                  渠道：{{ channel.name }}
                </div>
                <div class="flex items-center">
                  <div v-if="channel.isDisabled" class="text-#ff3333 bg-#ffe6e6 px-10px py-1px rounded-15px text-11px">
                    已停用
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="w-100% px-15px pb-20px flex items-center gap-10px position-sticky bottom-0 bg-#fff">
          <div class="flex items-center justify-between gap-10px">
            <div
              class="cursor-pointer flex items-center justify-center w-18px h-18px rounded-50% border-1px border-solid border-#ccc"
              @click="toggleAllSelection"
            >
              <CheckOutlined
                :class="selectedChannels.length === selectedChannelsLength ? 'opacity-100' : 'opacity-0'"
                class="text-#06f text-13px"
              />
            </div>
            <span class="text-14px text-#666">已选：{{ selectedChannels.length }}</span>
          </div>
          <a-button class="flex-1 h-34px rounded-15px" type="primary" @click="handleAdjust">
            选择调整到分类
          </a-button>
        </div>
      </div>
    </a-drawer>
  </div>
</template>

<style lang="less" scoped>
.registration-settings {
  padding: 5px 0px;
  background-color: #f6f7f8;
  height: 100%;
  position: relative; // 添加相对定位，作为按钮的定位参考

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

// 添加新增按钮样式
.add-button {
  position: sticky;
  left: 300px;
  bottom: 10px;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background-color: #0066ff;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  cursor: pointer;
  z-index: 10;
  transition: all 0.3s;

  &:hover {
    transform: scale(1.05);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }

  &:active {
    transform: scale(0.95);
  }
}

.channel-action-menu {
  border-radius: 12px 12px 0 0;
  overflow: hidden;
  box-shadow: 0 -4px 10px rgba(0, 0, 0, 0.05);
  background-color: #fff;
  display: flex;
  flex-direction: column;

  .menu-item {
    height: 54px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 16px;
    color: #333;
    border-bottom: 1px solid #f0f0f0;
    cursor: pointer;
    transition: all 0.2s ease;
    position: relative;

    .menu-icon {
      margin-right: 8px;
      font-size: 18px;
    }

    &:active {
      background-color: #f0f0f0;
    }

    &:hover {
      background-color: #f9f9f9;
    }

    &.danger {
      color: #ff4d4f;

      &:hover {
        background-color: rgba(255, 77, 79, 0.05);
      }
    }

    &.info {
      color: #1890ff;

      &:hover {
        background-color: rgba(24, 144, 255, 0.05);
      }
    }

    &.cancel {
      margin-top: 8px;
      color: #666;
      border-top: none;
      border-bottom: none;
      background-color: #fff;
      font-weight: 500;

      &:hover {
        background-color: #f9f9f9;
      }

      &:before {
        content: '';
        position: absolute;
        top: -8px;
        left: 0;
        right: 0;
        height: 8px;
        background-color: #f6f7f8;
      }
    }
  }
}

.create-channel-form {
  .drawer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;

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

  .form-content {
    padding: 20px;

    .form-item {
      margin-bottom: 16px;

      .label {
        margin-bottom: 8px;
        font-size: 14px;
        color: #333;

        .required {
          color: #ff4d4f;
          margin-right: 4px;
        }
      }

      .channel-input {
        border-radius: 4px;
      }

      .channel-textarea {
        border-radius: 4px;
        resize: none;
      }
    }

    .submit-btn {
      width: 100%;
      height: 40px;
      border-radius: 4px;
      font-size: 14px;
      margin-top: 8px;
    }
  }
}

::v-deep(.create-drawer .ant-drawer-content-wrapper) {
  height: 400px !important;
  bottom: -45px !important;
}

::v-deep(.create-drawer .ant-drawer-content) {
  border-radius: 16px 16px 0 0;
  overflow: hidden;
}

::v-deep(.create-drawer .ant-drawer-body) {
  padding: 0 !important;
}

::v-deep(.ant-drawer-content) {
  border-radius: 16px 16px 0 0;
  overflow: hidden;
}

::v-deep(.ant-drawer-header) {
  display: none;
}

::v-deep(.ant-drawer-body) {
  padding: 0 !important;
}

::v-deep(.ant-drawer-content-wrapper) {
  height: fit-content !important;
  max-height: 80vh !important;
  bottom: 0px !important;
}

::v-deep(.ant-drawer-content) {
  border-radius: 16px 16px 0 0;
  overflow: hidden;
  height: 100%;
}

::v-deep(.ant-drawer-body) {
  padding: 0 !important;
  height: 100%;
  display: flex;
  flex-direction: column;
}

.adjust-form {
  .drawer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid #f0f0f0;

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

  .form-content {
    padding: 20px;

    .option-title {
      font-size: 14px;
      color: #333;
      margin-bottom: 12px;
    }

    .category-option {
      padding: 12px 16px;
      background-color: #fff;
      border-radius: 4px;
      margin-bottom: 16px;
      cursor: pointer;
      border: 1px solid #f0f0f0;

      &:hover {
        background-color: #f9f9f9;
      }

      .category-name {
        font-size: 14px;
        color: #333;
      }
    }
  }
}

// 调整抽屉样式
:deep(.adjust-drawer .ant-drawer-content-wrapper) {
  height: 500px !important;
  bottom: 0px !important;
}

:deep(.batch-adjust-drawer .ant-drawer-content-wrapper) {
  height: 500px !important;
  bottom: 0px !important;
}

.batch-adjust-form {
  .drawer-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    // border-bottom: 1px solid #f0f0f0;

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

  .channel-list {
    overflow-y: auto;
    flex: 1;
  }

  .channel-item {
    &:hover {
      background-color: #f9f9f9;
    }
  }

  // .category-section {
  //   // margin-bottom: 10px;
  // }
}
</style>
