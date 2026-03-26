<script setup>
import {
  ExclamationCircleOutlined,
} from '@ant-design/icons-vue'
import classReviewDrawer from './classReviewDrawer.vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref([
  'intention',
  'followStatus',
  'sex',
  'createUser',
  'createTime',
  'intentionCourse',
  'reference',
  'studentStatus',
  'classEndingTime',
  'classStopTime',
])
const dataSource = ref([{ key: 1 }, { key: 2 }])
const allColumns = ref([
  {
    title: '上课日期/时段',
    dataIndex: 'classDateTime',
    key: 'classDateTime',
    fixed: 'left',
    width: 160,
    // 排序 ，默认倒序
    sorter: {
      compare: (a, b) => a.classDateTime - b.classDateTime,
    },
    defaultSortOrder: 'descend', // 设置默认排序顺序为降序
    required: true, // 新增必选标识
  },
  {
    title: '类型',
    dataIndex: 'type',
    key: 'type',
    width: 120,
  },
  {
    title: '所属班级/1对1',
    key: 'linkClassOr1v1',
    dataIndex: 'linkClassOr1v1',
    width: 160,
  },
  {
    title: '所属课程',
    key: 'linkCourse',
    dataIndex: 'linkCourse',
    width: 160,
  },
  {
    title: '上课老师',
    key: 'teacher',
    dataIndex: 'teacher',
    width: 130,
  },
  {
    title: '点评统计',
    key: 'commentStatistics',
    dataIndex: 'commentStatistics',
    width: 130,
  },
  {
    title: '已读/未读',
    dataIndex: 'readOrUnread',
    key: 'readOrUnread',
    width: 150,
  },
  {
    title: '上课助教',
    dataIndex: 'subTeacher',
    key: 'subTeacher',
    width: 130,
  },
  {
    title: '上课教室',
    dataIndex: 'classRoom',
    key: 'classRoom',
    width: 130,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 140,
  },
])
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'class-comment-list', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })

const openDrawer = ref(false)
function handleSeeStuData() {
  openDrawer.value = true
}
const openOrderDetailDrawer = ref(false)
function handleOrderDetail() {
  openOrderDetailDrawer.value = true
}

const openClassReviewDrawer = ref(false)
const classReviewDrawerType = ref(1)

function handelComment(type) {
  classReviewDrawerType.value = type
  openClassReviewDrawer.value = true
}
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter
        :display-array="displayArray"
        :is-quick-show="false"
      />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ dataSource.length }} 条数据
          </div>
          <div class="edit flex">
            <!-- 自定义字段 -->
            <!-- <customize-code
              v-model:checkedValues="selectedValues"
              :options="columnOptions"
              :total="allColumns.length - 1"
              :num="selectedValues.length - 1"
            /> -->
          </div>
        </div>
        <div class="table-content mt-2">
          <!-- <div class="tip mb2 text-#06f"> <ExclamationCircleFilled />  </div> -->
          <a-alert
            class="mb2 text-#06f"
            message="家长分享课评，机构就可获得转介绍线索"
            type="info"
            show-icon
            closable
          />
          <a-table
            :data-source="dataSource"
            :pagination="dataSource.length > 10"
            :columns="filteredColumns"
            :scroll="{ x: totalWidth }"
            size="small"
          >
            <template #headerCell="{ column }">
              <template v-if="column.key === 'commentStatistics'">
                <span class="mr-1">{{ column.title }}</span>
                <a-tooltip color="#666">
                  <template #title>
                    已点评人数/应点评人数
                  </template>
                  <ExclamationCircleOutlined />
                </a-tooltip>
              </template>
            </template>
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'classDateTime'">
                <div class="text-#222">
                  2025-04-01（周二）
                </div>
                <div class="text-#222">
                  15:00～16:00
                </div>
              </template>
              <template v-if="column.key === 'type'">
                <div class="justify-between flex-center">
                  <span>一对一</span>
                  <img
                    height="45"
                    src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12083/static/one2one-tag.03fd85df.svg"
                    alt=""
                  >
                </div>
              </template>
              <template v-if="column.key === 'linkClassOr1v1'">
                <div class="text-#222">
                  妞妞-一对一认知课
                </div>
              </template>
              <template v-if="column.key === 'linkCourse'">
                <div class="text-#222">
                  一对一认知课
                </div>
              </template>
              <template v-if="column.key === 'teacher'">
                <div class="text-#222">
                  商老师
                </div>
              </template>
              <template v-if="column.key === 'commentStatistics'">
                <div class="text-#222">
                  0/1
                </div>
              </template>
              <template v-if="column.key === 'readOrUnread'">
                <div class="text-#222">
                  已读0人
                </div>
                <div class="text-#888 text-3">
                  未读2人
                </div>
              </template>
              <template v-if="column.key === 'subTeacher'">
                <div class="text-#222">
                  何红武
                </div>
              </template>
              <template v-if="column.key === 'classRoom'">
                <div class="text-#222">
                  -
                </div>
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="14">
                  <a class="font500" @click="handelComment('1')">去点评{{ record.a }}</a>
                  <a class="font500" @click="handelComment('0')">查看{{ record.a }}</a>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <student-info-drawer v-model:open="openDrawer" />
    <order-detail-drawer
      v-model:open="openOrderDetailDrawer"
    />
    <classReviewDrawer v-model="openClassReviewDrawer" :type="classReviewDrawerType" />
  </div>
</template>

  <style lang="less" scoped>
.total {
  position: relative;
  padding-left: 10px;
  color: #222;
  display: flex;
  align-items: center;

  &::before {
    display: inline-block;
    background: var(--pro-ant-color-primary);
    border-radius: 2px;
    content: "";
    height: 12px;
    left: 0;
    position: absolute;
    width: 4px;
  }
}

.tip {
  padding: 10px 24px 10px 14px;
  background: #e6f0ff;
}

.upNew {
  position: relative;

  &::before {
    position: absolute;
    top: -12px;
    left: -22px;
    z-index: 999;
    width: 39px;
    height: 22px;
    background: url(data:image/png;base64,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);
    background-size: contain;
    content: "";
  }
}
</style>
