<script setup>
import { CloseOutlined, QuestionCircleOutlined, SwapLeftOutlined } from '@ant-design/icons-vue'
import { ref } from 'vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
})
const emit = defineEmits(['update:open'])
// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})
const manualSort = ref(false)

function closeFun() {
  openModal.value = false
}

// 更新展示顺序
function updateDisplayOrder() {
  data.value.forEach((item, index) => {
    item.displayOrder = index + 1
  })
}

// 调整顺序
function moveUp(index) {
  if (index > 0) {
    const temp = data.value[index]
    data.value[index] = data.value[index - 1]
    data.value[index - 1] = temp
    updateDisplayOrder()
  }
}

function moveDown(index) {
  if (index < data.value.length - 1) {
    const temp = data.value[index]
    data.value[index] = data.value[index + 1]
    data.value[index + 1] = temp
    updateDisplayOrder()
  }
}

// 拖拽相关
function onDragStart(event, index) {
  if (!manualSort.value) {
    event.preventDefault()
    return
  }
  event.dataTransfer.setData('text/plain', index)
  event.dataTransfer.effectAllowed = 'move'
}

function onDragOver(event) {
  if (!manualSort.value) {
    return
  }
  event.preventDefault()
  event.dataTransfer.dropEffect = 'move'
}

function onDrop(event, dropIndex) {
  if (!manualSort.value) {
    return
  }
  event.preventDefault()
  const dragIndex = parseInt(event.dataTransfer.getData('text/plain'))
  if (dragIndex !== dropIndex && dragIndex !== undefined && !isNaN(dragIndex)) {
    const dragItem = data.value[dragIndex]
    data.value.splice(dragIndex, 1)
    data.value.splice(dropIndex, 0, dragItem)
    updateDisplayOrder()
  }
}

const columns = [
  {
    title: '展示顺序',
    dataIndex: 'displayOrder',
    width: 100,
  },
  {
    title: '剩余课时',
    dataIndex: 'remainingHours',
    width: 120,
  },
  {
    title: '来源',
    dataIndex: 'source',
    width: 120,
  },
  {
    title: '生成时间',
    dataIndex: 'createTime',
    width: 180,
  },
  {
    title: '有效期至',
    dataIndex: 'validity',
    width: 200,
  },
  {
    title: '状态',
    dataIndex: 'status',
    width: 100,
  },
  {
    title: '操作',
    dataIndex: 'action',
    width: 100,
    fixed: 'right',
  },
]
const data = ref([
  {
    displayOrder: 1,
    remainingHours: '1课时',
    source: '报名/续费',
    createTime: '2025-04-13 10:00',
    validity: '',
    status: '未过期',
  },
  {
    displayOrder: 2,
    remainingHours: '2课时',
    source: '报名/续费',
    createTime: '2025-04-13 10:00',
    validity: '',
    status: '未过期',
  },
  {
    displayOrder: 3,
    remainingHours: '3课时',
    source: '报名/续费',
    createTime: '2025-04-13 10:00',
    validity: '',
    status: '未过期',
  },
])
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false"
    :closable="false" :mask-closable="false" :width="800"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>剩余详情</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div class="contenter scrollbar">
     <div class="flex justify-between items-start">
      <div>
        <div class="text-16px text-#444 mb-4px">
        课程名称：认知课
      </div>
      <div class="text-14px text-#888 mb-14px">
        学费变动（课消、转课或退课）会按照当前顺序扣减
      </div>
      </div>
      <div class="flex items-center">
        <span class="text-14px text-#888 mr-8px">
          手动排序
          <a-tooltip overlay-class-name="manual-sort-tooltip">
            <template #title>
              <div class="tooltip-content">
                <div>开启手动排序后,可自定义课消顺序;</div>
                <div>当前智能排序:先进先出</div>
              </div>
            </template>
            <QuestionCircleOutlined class="cursor-pointer" />
          </a-tooltip>
        </span>
        <a-switch v-model:checked="manualSort" />
      </div>
     </div>
      <!-- 剩余课时	来源	生成时间	有效期至	状态	详情 -->
      <a-table 
        :columns="columns" 
        :data-source="data" 
        :pagination="false" 
        row-key="displayOrder"
        :scroll="{ x: 940 }"
        :custom-row="(record, index) => ({
          draggable: manualSort,
          class: manualSort ? 'draggable-row' : '',
          onDragstart: (e) => onDragStart(e, index),
          onDragover: onDragOver,
          onDrop: (e) => onDrop(e, index),
        })"
      >
        <template #bodyCell="{ column, record, index }">
          <!-- 展示顺序 -->
          <template v-if="column.dataIndex === 'displayOrder'">
            <div class="flex items-center gap-1">
              <span>{{ record.displayOrder }}</span>
              <div v-if="manualSort" class="flex">
                <a-tooltip title="长按拖拽调整顺序">
                  <SwapLeftOutlined
                    class="text-18px cursor-pointer transition-colors rotate--90 font-800"
                    :class="index === data.length - 1 ? 'text-#ccc cursor-not-allowed' : 'text-#06f hover:text-#1890ff'"
                    @click="index < data.length - 1 && moveDown(index)"
                  />
                </a-tooltip>
                <a-tooltip title="长按拖拽调整顺序">
                  <SwapLeftOutlined
                    class="text-18px cursor-pointer transition-colors rotate-90 -ml-2 font-800"
                    :class="index === 0 ? 'text-#ccc cursor-not-allowed' : 'text-#06f hover:text-#1890ff'"
                    @click="index > 0 && moveUp(index)"
                  />
                </a-tooltip>
              </div>
            </div>
          </template>
          <!-- 有效期 -->
          <template v-if="column.dataIndex === 'validity'">
            <span v-if="record.validity === ''">
              不限制 <a-button type="link" class="text-#06f">修改有效期</a-button>
            </span>
            <span v-else>
              {{ record.validity }}<a-button type="link" class="text-#06f">修改有效期</a-button>
            </span>
          </template>
          <!-- 详情 -->
          <template v-if="column.dataIndex === 'action'">
            <a-button type="link" class="text-#06f">
              订单详情
            </a-button>
          </template>
        </template>
      </a-table>
    </div>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        关闭
      </a-button>
    </template>
  </a-modal>
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
.contenter{
  padding: 24px;
}

/* 手动排序提示框样式 */
:deep(.manual-sort-tooltip) {
  .ant-tooltip-inner {
    background-color: #4a4a4a !important;
    color: #fff !important;
    padding: 8px 12px !important;
    border-radius: 4px !important;
    font-size: 12px !important;
    line-height: 1.5 !important;
  }

  .ant-tooltip-arrow::before {
    background-color: #4a4a4a !important;
  }
}

.tooltip-content {
  div {
    white-space: nowrap;
    
    &:first-child {
      margin-bottom: 4px;
    }
  }
}

/* 拖拽排序样式 */
:deep(.draggable-row) {
  cursor: move;
  transition: all 0.2s;
  
  &:hover {
    background-color: #f5f5f5;
  }
  
  &:active {
    cursor: grabbing;
  }
}

:deep(.ant-table-tbody tr.draggable-row[draggable="true"]) {
  user-select: none;
}
</style>

<style>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
