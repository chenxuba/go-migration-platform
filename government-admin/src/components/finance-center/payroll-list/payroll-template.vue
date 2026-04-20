<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import { useTableColumns } from '@/composables/useTableColumns'

const displayArray = ref([
  'intention',
  'intentionCourse',
  'reference',
  'studentStatus',
  'classEndingTime',
  'classStopTime',
])
const dataSource = ref([])
const allColumns = ref([
  {
    title: '模板名称',
    dataIndex: 'titleOrDate',
    key: 'titleOrDate',
    width: 180,
  },
  {
    title: '运用工资项目',
    dataIndex: 'user',
    key: 'user',
    width: 260,
  },
  {
    title: '适用员工',
    dataIndex: 'applicableStaff',
    key: 'applicableStaff',
    width: 120,
  },
  {
    title: '创建人',
    dataIndex: 'createUser',
    key: 'createUser',
    width: 100,
  },
  {
    title: '创建日期',
    dataIndex: 'createTime',
    key: 'createTime',
    width: 140,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    width: 140,
    fixed: 'right',
  },
])
const payrollSource = ref([{}, {}, {}, {}])
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'payroll-template', // 本地存储键名
    allColumns, // 原始列配置
    excludeKeys: ['action'], // 需要排除的列键
  })
const openSalaryProjectManagement = ref(false)
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-lb-4 rounded-rb-4">
      <all-filter :display-array="displayArray" :is-quick-show="false" :is-show-search-input="true"
        search-label="模版名称" />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            共 {{ dataSource.length }} 个工资构成模板
          </div>
          <div class="edit flex">
            <a-button class="mr-2" @click="openSalaryProjectManagement = true">
              工资项目管理
            </a-button>
            <a-button type="primary" class="mr-2">
              新建模版
            </a-button>
            <!-- 自定义字段 -->
            <!-- <customize-code v-model:checkedValues="selectedValues" :options="columnOptions"
              :total="allColumns.length - 1" :num="selectedValues.length - 1" /> -->
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns"
            :scroll="{ x: totalWidth }">
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'titleOrDate'">
                <div class="text-#222">
                  {{ record.titleOrDate }}
                </div>
              </template>
              <template v-if="column.key === 'user'">
                {{ record.user }}
              </template>
              <template v-if="column.key === 'applicableStaff'">
                {{ record.applicableStaff }}名
              </template>
              <template v-if="column.key === 'createUser'">
                {{ record.createUser }}
              </template>
              <template v-if="column.key === 'createTime'">
                {{ record.createTime }}
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="16">
                  <a class=" whitespace-nowrap">编辑</a>
                  <a class=" whitespace-nowrap">复制</a>
                  <a class=" whitespace-nowrap">删除</a>
                </a-space>
              </template>
            </template>
            <template #emptyText>
              <div class="empty-state">
                <img
                  src="https://prod-tbu-next-erp-cdn.schoolpal.cn/next-pc-static/static/12585/static/empty2x.6a92b2d7.png"
                  alt="暂无工资模板" class="empty-image" />
                <div class="empty-title">暂无工资模板，你可以参照以下步骤进行配置</div>
                <div class="steps-container">
                  <div class="step-item">
                    <div class="step-number">1</div>
                    <div class="step-content">
                      <div class="step-title">配置工资项目</div>
                      <a-button type="primary" ghost class="step-button" @click="openSalaryProjectManagement = true">
                        去配置
                      </a-button>
                    </div>
                  </div>
                  <div class="step-connector"></div>
                  <div class="step-item">
                    <div class="step-number">2</div>
                    <div class="step-content">
                      <div class="step-title">用工资项目构成模板</div>
                      <a-button type="primary" ghost class="step-button">
                        去配置
                      </a-button>
                    </div>
                  </div>
                </div>
              </div>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <salary-project-management v-model:open="openSalaryProjectManagement" />
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

span.dot {
  border-radius: 50%;
  display: inline-block;
  height: 6px;
  position: relative;
  vertical-align: middle;
  width: 6px;
  margin-right: 4px;
  background: #06f;
}

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

/* 空状态样式 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;

  .empty-image {
    width: 120px;
    height: auto;
    margin-bottom: 24px;
    opacity: 0.8;
  }

  .empty-title {
    font-size: 14px;
    color: #666;
    margin-bottom: 32px;
    font-weight: 500;
    line-height: 1.5;
  }

  .steps-container {
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    margin-left: 25px;

    .step-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      text-align: center;
      position: relative;
      z-index: 2;

      .step-number {
        width: 30px;
        height: 30px;
        border-radius: 50%;
        border: 1px solid #06f;
        color: #06f;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 16px;
        font-weight: 500;
        margin-bottom: 12px;
      }

      .step-content {
        display: flex;
        flex-direction: column;
        align-items: center;

        .step-title {
          font-size: 16px;
          color: #333;
          margin-bottom: 12px;
          font-weight: 500;
        }

        .step-button {
          border-radius: 16px;
          height: 32px;
          padding: 0 16px;
          font-size: 14px;
        }
      }
    }

    .step-connector {
      width: 100px;
      height: 1px;
      background: #d9d9d9;
      position: relative;
      z-index: 1;
      margin-left: 10px;
      margin-bottom: 80px; // 调整连接线位置，使其与步骤编号对齐
    }
  }
}
</style>
