<script setup>
import { CaretDownOutlined, DownOutlined } from '@ant-design/icons-vue'
import CreateClassModal from '@/components/common/create-class-modal.vue'
import ClassListDrawer from '@/components/edu-center/class-list/class-list-drawer.vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref(['openClassStatus', 'doYouSchedule', 'billingMode', 'createUser', 'createTime', 'intentionCourse', 'reference', 'classEndingTime', 'classStopTime'])
const dataSource = ref([{}, {}])
const allColumns = ref([
  {
    title: '班级名称',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 200,
    required: true, // 新增必选标识
  },
  {
    title: '关联课程',
    key: 'linkCourse',
    dataIndex: 'linkCourse',
    width: 140,
  },
  {
    title: '学员数',
    key: 'studentNum',
    dataIndex: 'studentNum',
    width: 110,
  },
  {
    title: '班主任',
    key: 'headTeacher',
    dataIndex: 'headTeacher',
    width: 110,

  },
  {
    title: '上课教室',
    dataIndex: 'classRoom',
    key: 'classRoom',
    width: 120,
  },
  {
    title: '上课时间',
    dataIndex: 'classTime',
    key: 'classTime',
    width: 300,
  },
  {
    title: '是否排课',
    dataIndex: 'doYouSchedule',
    key: 'doYouSchedule',
    width: 120,
  },
  {
    title: '已上/日程总数',
    dataIndex: 'alreadyOnOrtotal',
    key: 'alreadyOnOrtotal',
    width: 150,
  },
  {
    title: '状态',
    dataIndex: 'openClassStatus',
    key: 'openClassStatus',
    width: 120,

  },
  {
    title: '创建时间',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 180,

  },
  {
    title: '创建人',
    key: 'createUser',
    dataIndex: 'createUser',
    width: 100,
  },
  // 备注
  {
    title: '备注',
    key: 'remark',
    dataIndex: 'remark',
    width: 120,
  },
  // 结班日期
  {
    title: '结班日期',
    key: 'classEndingTime',
    dataIndex: 'classEndingTime',
    width: 120,
  },

  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 220,
  },

])
const rowSelection = {
  onChange: (selectedRowKeys, selectedRows) => {
    console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', selectedRows)
  },
}
const defaultOpenClassStatus = ref(1)
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'class-list', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })
const createClassModal = ref(false)
function createClass() {
  createClassModal.value = true
}
function onClickMenu(key) {
  console.log(key)
}
const classListDrawerFlag = ref(false)
function openClassListDrawer() {
  classListDrawerFlag.value = true
}
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap  bg-white  pl-3 pr-3 rounded-4">
      <all-filter
        :default-open-class-status="defaultOpenClassStatus"
        :display-array="displayArray" :is-quick-show="false" :is-show-search-stu-phone="false" :is-show-clsss-or-course-search="true" search-label="班级名称"
      />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            总计{{ dataSource.length }}个班级 ，2 个开班中，在读学员 4 人，在读人次 5 人
          </div>
          <div class="edit flex">
            <a-button class="mr-2">
              批量结班
            </a-button>
            <a-dropdown class="mr-2">
              <template #overlay>
                <a-menu>
                  <a-menu-item key="0">
                    导入班级
                  </a-menu-item>
                  <a-menu-item key="1">
                    批量导出
                  </a-menu-item>
                  <a-menu-item key="2">
                    导出记录
                  </a-menu-item>
                </a-menu>
              </template>
              <a-button>
                导出数据
                <DownOutlined :style="{ fontSize: '10px' }" />
              </a-button>
            </a-dropdown>
            <a-button type="primary" class="mr-2" @click="createClass">
              创建班级
            </a-button>
            <!-- 自定义字段 -->
            <customize-code
              v-model:checked-values="selectedValues" :options="columnOptions" :total="allColumns.length - 1"
              :num="selectedValues.length - 1"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns"
            :row-selection="rowSelection" :scroll="{ x: totalWidth }" size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <a-button type="link" class="text-#222" @click="openClassListDrawer">
                  视知觉康复班级
                </a-button>
              </template>
              <template v-if="column.key === 'linkCourse'">
                <div class="text-#222">
                  奥夫音乐课{{ record.a }}
                </div>
                <div class="text-3 text-#888 flex flex-items-center">
                  课程
                </div>
              </template>
              <template v-if="column.key === 'studentNum'">
                0
              </template>
              <template v-if="column.key === 'headTeacher'">
                张晨
              </template>
              <template v-if="column.key === 'classRoom'">
                -
              </template>
              <template v-if="column.key === 'classTime'">
                隔天 10:00 ～ 10:30
              </template>
              <template v-if="column.key === 'doYouSchedule'">
                <div class="studentStatus">
                  <span class="dot" />
                  <span>已排课</span>
                </div>
              </template>
              <template v-if="column.key === 'alreadyOnOrtotal'">
                0/4节
              </template>
              <template v-if="column.key === 'openClassStatus'">
                <div class="text-#06f bg-#e6f0ff rounded-2.5 inline-block text-3 pt-0.5 pb-0.5 pl-2 pr-2">
                  开班中
                </div>
              </template>
              <template v-if="column.key === 'createTime'">
                2025-03-24 08:01
              </template>
              <template v-if="column.key === 'createUser'">
                龙钊
              </template>
              <template v-if="column.key === 'remark'">
                -
              </template>
              <template v-if="column.key === 'classEndingTime'">
                2023-12-12
              </template>
              <template v-else-if="column.key === 'action'">
                <span class="flex action">
                  <a class="mr-3">排课</a>
                  <a class="mr-3">添加学员</a>
                  <div style="cursor: pointer;">
                    <a-dropdown :trigger="['click']" placement="bottom">
                      <a @click.prevent>
                        <div class="intention">更多<CaretDownOutlined
                          class=" text-#1677ff"
                          :style="{ 'font-size': '12px' }"
                        />
                        </div>
                      </a>
                      <template #overlay>
                        <a-menu style="text-align: center;width: 120px;" @click="onClickMenu">
                          <a-menu-item key="1">
                            上课点名
                          </a-menu-item>
                          <a-menu-item key="2">
                            未排课点名
                          </a-menu-item>
                          <a-menu-item key="3">
                            编辑班级
                          </a-menu-item>
                          <a-menu-item key="4">
                            结班
                          </a-menu-item>
                        </a-menu>
                      </template>
                    </a-dropdown>
                  </div>
                </span>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <CreateClassModal v-model:open="createClassModal" />
    <ClassListDrawer v-model:open="classListDrawerFlag" />
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

.studentStatus {
  display: flex;
  align-items: center;
  span.dot {
    border-radius: 50%;
    display: inline-block;
    height: 6px;
    position: relative;
    vertical-align: middle;
    width: 6px;
    margin-right: 4px;
    background: var(--pro-ant-color-primary);
  }
}
</style>
