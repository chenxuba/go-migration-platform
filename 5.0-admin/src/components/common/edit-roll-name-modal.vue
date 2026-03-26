<script setup>
import { CloseOutlined, DownOutlined, ExclamationCircleOutlined } from '@ant-design/icons-vue'
import editRollNameAddStuModal from './edit-roll-name-add-stu-modal.vue'

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

// defineEmits(['update:open']);

const userName = ref('')
const columns = ref([
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
    width: 200,
  },
  {
    title: '上课状态',
    dataIndex: 'status',
    width: 160,
    filters: [
      {
        text: '未记录',
        value: '0',
      },
      {
        text: '到课',
        value: '1',
      },
      {
        text: '请假',
        value: '2',
      },
      {
        text: '旷课',
        value: '3',
      },
    ],
    onFilter: (value, record) => {
      return record.status === value
    },
  },
  {
    title: '课消方式',
    dataIndex: 'courseType',
    width: 140,
  },
  {
    title: '上课点名数量',
    dataIndex: 'rollNameCount',
    width: 140,
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
const data = ref([
  {
    index: 1,
    name: '张三',
    status: '0',
    courseType: '按课时',
    rollNameCount: '',
    innerRemark: '-',
    outerRemark: '-',
    type: '1',
  },
  {
    index: 2,
    name: '李四',
    status: '1',
    courseType: '按课时',
    rollNameCount: 1,
    innerRemark: '-',
    outerRemark: '-',
    type: '2',
  },
  {
    index: 3,
    name: '王五',
    status: '2',
    courseType: '按课时',
    rollNameCount: 0,
    innerRemark: '-',
    outerRemark: '-',
    type: '3',
  },
  {
    index: 4,
    name: '赵六',
    status: '3',
    courseType: '按课时',
    rollNameCount: 0,
    innerRemark: '-',
    outerRemark: '-',
    type: '4',
  },
])
// totalWidth 计算属性
const totalWidth = computed(() => {
  return columns.value.reduce((acc, column) => acc + column.width, 0)
})
// editRollNameAddStuModal
const editRollNameAddStuModals = ref(false)
const editRowRollNameModals = ref(false)
const defaultTitle = ref('')
// 添加学员
function handleAddStudent({ key }) {
  if (key === '1') {
    defaultTitle.value = '添加补课学员'
  }
  else if (key === '2') {
    defaultTitle.value = '添加临时学员'
  }
  else if (key === '3') {
    defaultTitle.value = '添加试听学员'
  }
  else if (key === '4') {
    defaultTitle.value = '添加班级学员'
  }
  editRollNameAddStuModals.value = true
}
// 编辑
function handleEdit(record) {
  editRowRollNameModals.value = true
}
</script>

<template>
  <div>
    <a-drawer
      v-model:open="openDrawer" :push="{ distance: 80 }" :body-style="{ padding: '0', background: '#f7f7fd' }"
      :closable="false" width="1165px" placement="right"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            编辑点名
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <div class="search px-24px py-12px">
        <a-input v-model:value="userName" placeholder="搜索学员" class="h-48px rounded-12px">
          <template #prefix>
            <img
              src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12181/static/magnifying.2bcc08ab.svg"
              alt="" class="pr-6px mt--2px"
            >
          </template>
        </a-input>
      </div>
      <div class="contenter">
        <!-- table -->
        <a-table
          :columns="columns" :data-source="data" row-key="id" :pagination="false"
          :scroll="{ x: totalWidth }"
        >
          <template #headerCell="{ column }">
            <div v-if="column.dataIndex === 'courseType'" class="flex items-center">
              {{ column.title }}
              <a-popover title="课消方式">
                <template #content>
                  <div class="w-320px">
                    【按时间】按天数计费<br>
                    【按课时】按课时计费<br>
                    【按金额】按金额计费<br>
                    【提示】按时间、按金额计费，开启记录课时后仅作为「记录」，课时增减不产生学费变动。
                  </div>
                </template>
                <ExclamationCircleOutlined class="text-#333 cursor-pointer ml-4px" />
              </a-popover>
            </div>
          </template>
          <!-- index -->
          <template #bodyCell="{ column, record, index }">
            <div v-if="column.dataIndex === 'index'">
              {{ index + 1 }}
            </div>
            <!-- 学员 -->
            <div v-if="column.dataIndex === 'name'">
              <div class="flex items-center">
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
                <span>{{ record.name }}</span>
              </div>
            </div>
            <!-- 上课状态 -->
            <div v-if="column.dataIndex === 'status'">
              <span v-if="record.status === '0'" class="bg-#f6f7f8 text-#888 text-3 px2 py1 rounded-10 ml2">未记录</span>
              <span v-if="record.status === '1'" class="bg-#e6f0ff text-#06f text-3 px2 py1 rounded-10 ml2">到课</span>
              <span v-if="record.status === '2'" class="bg-#fff5e6 text-#f90 text-3 px2 py1 rounded-10 ml2">请假</span>
              <span v-if="record.status === '3'" class="bg-#ffe6e6 text-#f33 text-3 px2 py1 rounded-10 ml2">旷课</span>
            </div>
            <!-- 课消方式 -->
            <div v-if="column.dataIndex === 'courseType'">
              {{ record.courseType }}
            </div>
            <!-- 上课点名数量 -->
            <div v-if="column.dataIndex === 'rollNameCount'">
              <span v-if="record.rollNameCount === ''">未记录课时</span>
              <span v-if="record.rollNameCount > 0">{{ record.rollNameCount }} 课时</span>
              <span v-if="record.rollNameCount === 0">不计课时</span>
            </div>
            <!-- 对内备注 -->
            <div v-if="column.dataIndex === 'innerRemark'">
              {{ record.innerRemark }}
            </div>
            <!-- 对外备注 -->
            <div v-if="column.dataIndex === 'outerRemark'">
              {{ record.outerRemark }}
            </div>
            <!-- 操作 -->
            <div v-if="column.dataIndex === 'action'">
              <a-space :size="20">
                <a @click="handleEdit(record)">编辑</a>
                <a>移出</a>
              </a-space>
            </div>
          </template>
        </a-table>
      </div>
      <!-- 自定义footer -->
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
              <a-menu-item key="4">
                添加班级学员
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
    <edit-roll-name-add-stu-modal v-model:open="editRollNameAddStuModals" :default-title="defaultTitle" />
    <edit-row-roll-name-modal v-model:open="editRowRollNameModals" />
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
