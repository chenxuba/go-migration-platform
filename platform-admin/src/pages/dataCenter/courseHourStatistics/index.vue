<!-- 课时统计 -->
<script setup>
  import { QuestionCircleOutlined } from '@ant-design/icons-vue'
  import AllFilter from '@/components/common/all-filter.vue'
  
  const activeKey = ref('1')
  
  // 筛选条件配置
  const displayArray = ref([
    'stuPhoneSearch', // 学员/电话搜索
    'intentionCourse', // 意向课程（可用于课程筛选）
    'nextFollowTime', //下次跟进时间
  ])
  
  // 筛选值
  const stuPhoneSearchFilter = ref('')
  const intentionCourseFilter = ref([])
  
  // 搜索条件（兼容原有逻辑）
  const searchForm = computed(() => {
    let courseName = ''
    if (intentionCourseFilter.value && intentionCourseFilter.value.length > 0) {
      // intentionCourseFilter 可能是数组，包含 { label, value } 对象
      const firstCourse = intentionCourseFilter.value[0]
      courseName = typeof firstCourse === 'string' ? firstCourse : (firstCourse?.label || firstCourse?.value || '')
    }
    return {
      studentName: stuPhoneSearchFilter.value || '',
      courseName,
    }
  })
  
  // 数据简报
  const reportList = computed(() => {
    const totalStudents = dataSource.value.length
    const totalCourses = dataSource.value.reduce((sum, student) => sum + (student.courses?.length || 0), 0)
    const totalPurchased = dataSource.value.reduce((sum, student) => {
      return sum + (student.courses?.reduce((s, course) => s + (course.purchasedHours || 0), 0) || 0)
    }, 0)
    const totalConsumed = dataSource.value.reduce((sum, student) => {
      return sum + (student.courses?.reduce((s, course) => {
        // 优先使用 totalConsumed，如果没有则计算 normalConsumed + leaveDeducted
        return s + (course.totalConsumed || (course.normalConsumed || 0) + (course.leaveDeducted || 0) || course.consumedHours || 0)
      }, 0) || 0)
    }, 0)
    const totalRemaining = dataSource.value.reduce((sum, student) => {
      return sum + (student.courses?.reduce((s, course) => s + (course.remainingHours || 0), 0) || 0)
    }, 0)
  
    return [
      {
        title: '统计学生数量',
        value: totalStudents,
        briefing: true,
        popover_title: '统计学生数量',
        popover_content: '统计范围内参与课程的学生总数',
        chain: '',
        onYear: '',
      },
      {
        title: '统计课程数量',
        value: totalCourses,
        briefing: true,
        popover_title: '统计课程数量',
        popover_content: '统计范围内学生购买的课程总数',
        chain: '',
        onYear: '',
      },
      {
        title: '总购买课时',
        value: totalPurchased,
        briefing: true,
        popover_title: '总购买课时',
        popover_content: '所有学生购买的所有课程课时总数',
        chain: '',
        onYear: '',
      },
      {
        title: '总消耗课时',
        value: totalConsumed,
        briefing: true,
        popover_title: '总消耗课时',
        popover_content: '所有学生已消耗的课时总数',
        chain: '',
        onYear: '',
      },
      {
        title: '总剩余课时',
        value: totalRemaining,
        briefing: true,
        popover_title: '总剩余课时',
        popover_content: '所有学生剩余的课时总数',
        chain: '',
        onYear: '',
      },
      {
        title: '课时消耗率',
        value: totalPurchased > 0 ? `${((totalConsumed / totalPurchased) * 100).toFixed(1)}%` : '0%',
        briefing: true,
        popover_title: '课时消耗率',
        popover_content: '已消耗课时占购买课时的百分比',
        chain: '',
        onYear: '',
      },
    ]
  })
  
  
  // 表格列配置
  const columns = [
    {
      title: '学生姓名',
      dataIndex: 'studentName',
      key: 'studentName',
      width: 150,
      sorter: (a, b) => a.studentName.localeCompare(b.studentName),
    },
    {
      title: '课程数量',
      dataIndex: 'courseCount',
      key: 'courseCount',
      width: 100,
    },
    {
      title: '购买课时合计',
      dataIndex: 'totalPurchased',
      key: 'totalPurchased',
      width: 130,
      align: 'right',
    },
    {
      title: '消耗课时合计',
      dataIndex: 'totalConsumed',
      key: 'totalConsumed',
      width: 130,
      align: 'right',
    },
    {
      title: '剩余课时合计',
      dataIndex: 'totalRemaining',
      key: 'totalRemaining',
      width: 130,
      align: 'right',
    },
  ]
  
  // 课程详情表格列
  const courseColumns = [
    {
      title: '课程名称',
      dataIndex: 'courseName',
      key: 'courseName',
      width: 160,
    },
    {
      title: '课程类型',
      dataIndex: 'courseType',
      key: 'courseType',
      width: 100,
      align: 'center',
    },
    {
      title: '购买课时',
      dataIndex: 'purchasedHours',
      key: 'purchasedHours',
      width: 90,
      align: 'center',
    },
    {
      title: '正常消耗',
      dataIndex: 'normalConsumed',
      key: 'normalConsumed',
      width: 90,
      align: 'center',
    },
    {
      title: '请假扣课时',
      dataIndex: 'leaveDeducted',
      key: 'leaveDeducted',
      width: 100,
      align: 'center',
    },
    {
      title: '总消耗课时',
      dataIndex: 'totalConsumed',
      key: 'totalConsumed',
      width: 100,
      align: 'center',
    },
    {
      title: '剩余课时',
      dataIndex: 'remainingHours',
      key: 'remainingHours',
      width: 90,
      align: 'center',
    },
    {
      title: '有效期',
      dataIndex: 'expiryDate',
      key: 'expiryDate',
      width: 110,
      align: 'center',
    },
  ]
  
  // 表格数据
  const dataSource = ref([
    {
      key: '1',
      studentName: '张三',
      studentNo: 'STU001',
      phone: '13800138001',
      courses: [
        {
          courseName: '英语基础课程',
          courseType: '语言类',
          purchasedHours: 50,
          normalConsumed: 25,
          leaveDeducted: 5,
          totalConsumed: 30,
          remainingHours: 20,
          purchaseDate: '2024-01-15',
          expiryDate: '2025-01-15',
        },
        {
          courseName: '数学提高班',
          courseType: '学科类',
          purchasedHours: 40,
          normalConsumed: 25,
          leaveDeducted: 0,
          totalConsumed: 25,
          remainingHours: 15,
          purchaseDate: '2024-02-10',
          expiryDate: '',
        },
        {
          courseName: '钢琴启蒙课',
          courseType: '艺术类',
          purchasedHours: 30,
          normalConsumed: 8,
          leaveDeducted: 2,
          totalConsumed: 10,
          remainingHours: 20,
          purchaseDate: '2024-03-01',
          expiryDate: '2025-03-01',
        },
      ],
    },
    {
      key: '2',
      studentName: '李四',
      studentNo: 'STU002',
      phone: '13800138002',
      courses: [
        {
          courseName: '英语基础课程',
          courseType: '语言类',
          purchasedHours: 60,
          normalConsumed: 40,
          leaveDeducted: 5,
          totalConsumed: 45,
          remainingHours: 15,
          purchaseDate: '2023-12-20',
          expiryDate: '2024-12-20',
        },
        {
          courseName: '编程入门课',
          courseType: '技能类',
          purchasedHours: 48,
          normalConsumed: 30,
          leaveDeducted: 2,
          totalConsumed: 32,
          remainingHours: 16,
          purchaseDate: '2024-01-05',
          expiryDate: '',
        },
      ],
    },
    {
      key: '3',
      studentName: '王五',
      studentNo: 'STU003',
      phone: '13800138003',
      courses: [
        {
          courseName: '数学提高班',
          courseType: '学科类',
          purchasedHours: 40,
          normalConsumed: 38,
          leaveDeducted: 2,
          totalConsumed: 40,
          remainingHours: 0,
          purchaseDate: '2023-11-10',
          expiryDate: '2024-11-10',
        },
        {
          courseName: '英语基础课程',
          courseType: '语言类',
          purchasedHours: 50,
          normalConsumed: 18,
          leaveDeducted: 2,
          totalConsumed: 20,
          remainingHours: 30,
          purchaseDate: '2024-02-01',
          expiryDate: '',
        },
        {
          courseName: '美术创作课',
          courseType: '艺术类',
          purchasedHours: 36,
          normalConsumed: 10,
          leaveDeducted: 2,
          totalConsumed: 12,
          remainingHours: 24,
          purchaseDate: '2024-02-20',
          expiryDate: '2025-02-20',
        },
      ],
    },
    {
      key: '4',
      studentName: '赵六',
      studentNo: 'STU004',
      phone: '13800138004',
      courses: [
        {
          courseName: '编程入门课',
          courseType: '技能类',
          purchasedHours: 48,
          normalConsumed: 13,
          leaveDeducted: 2,
          totalConsumed: 15,
          remainingHours: 33,
          purchaseDate: '2024-01-20',
          expiryDate: '',
        },
      ],
    },
    {
      key: '5',
      studentName: '钱七',
      studentNo: 'STU005',
      phone: '13800138005',
      courses: [
        {
          courseName: '英语基础课程',
          courseType: '语言类',
          purchasedHours: 50,
          normalConsumed: 48,
          leaveDeducted: 2,
          totalConsumed: 50,
          remainingHours: 0,
          purchaseDate: '2023-10-15',
          expiryDate: '2024-10-15',
        },
        {
          courseName: '钢琴启蒙课',
          courseType: '艺术类',
          purchasedHours: 30,
          normalConsumed: 6,
          leaveDeducted: 2,
          totalConsumed: 8,
          remainingHours: 22,
          purchaseDate: '2024-03-05',
          expiryDate: '',
        },
      ],
    },
  ])
  
  // 过滤后的数据
  const filteredDataSource = computed(() => {
    let result = dataSource.value
  
    if (searchForm.value.studentName) {
      result = result.filter(item =>
        item.studentName.includes(searchForm.value.studentName) ||
        item.phone?.includes(searchForm.value.studentName),
      )
    }
  
    if (searchForm.value.courseName) {
      result = result.map(student => {
        const filteredCourses = student.courses?.filter(course =>
          course.courseName.includes(searchForm.value.courseName),
        )
        if (filteredCourses && filteredCourses.length > 0) {
          return { ...student, courses: filteredCourses }
        }
        return null
      }).filter(Boolean)
    }
  
    return result
  })
  
  </script>
  
  <template>
    <div class="home">
      <div class="tabs">
        <a-tabs
          v-model:active-key="activeKey"
          :tab-bar-style="{
            'border-bottom-left-radius': '0px',
            'border-bottom-right-radius': '0px',
          }"
        >
          <a-tab-pane key="1" tab="按学生">
            <div class="content-wrapper">
              <!-- 筛选区域 -->
              <div class="filter-wrap bg-white pl-3 pr-3 rounded-4">
                <all-filter
                  :display-array="displayArray"
                  @update:stuPhoneSearchFilter="(val) => { stuPhoneSearchFilter = val }"
                  @update:intentionCourseFilter="(val) => { intentionCourseFilter = val }"
                />
              </div>
  
              <!-- 数据简报 -->
              <div class="card-white">
                <div class="flex justify-between ml-2">
                  <div class="total font-bold">
                    数据简报
                  </div>
                </div>
                <div class="flex align-center p-12px gap-24px nowrap">
                  <div
                    v-for="(item, index) in reportList"
                    :key="index"
                    class="flex-1 p-12px bg-#fbfcff border-radius-12px"
                  >
                    <div class="block_top flex align-center gap-1">
                      <span class="text-#888 font-size-12px whitespace-nowrap">{{ item.title }}</span>
                      <a-popover color="#fff" :title="item.popover_title">
                        <template #content>
                          <div v-html="item.popover_content" />
                        </template>
                        <QuestionCircleOutlined class="text-#888 font-size-12px" />
                      </a-popover>
                    </div>
                    <div class="block_bottom">
                      <div class="font-size-24px font-bold" style="font-family: 'DIN Alternate'">
                        {{ item.value }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 数据明细 -->
              <div class="card-white" style="padding-left: 24px;">
                <div class="flex justify-between align-center py-8px">
                  <span>共计{{ filteredDataSource.length }}条数据</span>
                  <a-button ghost type="primary">
                    下载报表
                  </a-button>
                </div>
                <div class="table-content mt-2">
                  <a-table
                    :data-source="filteredDataSource"
                    :columns="columns"
                    :pagination="{ pageSize: 10, showTotal: (total) => `共 ${total} 条` }"
                    :scroll="{ x: 1000 }"
                    size="small"
                    :expandable="{
                      rowExpandable: (record) => record.courses && record.courses.length > 0,
                    }"
                  >
                    <template #expandedRowRender="{ record }">
                      <div class="expanded-row-content">
                        <div class="expanded-row-header">
                          <span class="expanded-row-title">{{ record.studentName }}的课程详情</span>
                          <span class="expanded-row-count">共 {{ record.courses?.length || 0 }} 门课程</span>
                        </div>
                        <a-table
                          :columns="courseColumns"
                          :data-source="record.courses || []"
                          :pagination="false"
                          size="small"
                          :bordered="true"
                          class="expanded-table"
                        >
                          <template #bodyCell="{ column, record: course }">
                            <template v-if="column.key === 'expiryDate'">
                              {{ course.expiryDate || '-' }}
                            </template>
                            <template v-else-if="column.key === 'totalConsumed'">
                              <span style="font-weight: 600; color: #1890ff;">
                                {{ course.totalConsumed || (course.normalConsumed || 0) + (course.leaveDeducted || 0) }}
                              </span>
                            </template>
                            <template v-else-if="column.key === 'normalConsumed'">
                              {{ course.normalConsumed || 0 }}
                            </template>
                            <template v-else-if="column.key === 'leaveDeducted'">
                              {{ course.leaveDeducted || 0 }}
                            </template>
                           
                          </template>
                        </a-table>
                      </div>
                    </template>
                    <template #bodyCell="{ column, record }">
                      <template v-if="column.key === 'courseCount'">
                        {{ record.courses?.length || 0 }}
                      </template>
                      <template v-else-if="column.key === 'totalPurchased'">
                        {{ record.courses?.reduce((sum, course) => sum + (course.purchasedHours || 0), 0) || 0 }}
                      </template>
                      <template v-else-if="column.key === 'totalConsumed'">
                        {{ record.courses?.reduce((sum, course) => {
                          return sum + (course.totalConsumed || (course.normalConsumed || 0) + (course.leaveDeducted || 0) || course.consumedHours || 0)
                        }, 0) || 0 }}
                      </template>
                      <template v-else-if="column.key === 'totalRemaining'">
                        {{ record.courses?.reduce((sum, course) => sum + (course.remainingHours || 0), 0) || 0 }}
                      </template>
                    </template>
                    <template #summary>
                      <a-table-summary fixed>
                        <a-table-summary-row class="summary-row">
                          <a-table-summary-cell :index="0">
                          </a-table-summary-cell>
                          <a-table-summary-cell :index="1">
                            总计
                          </a-table-summary-cell>
                          <a-table-summary-cell :index="2">
                            {{ filteredDataSource.reduce((sum, record) => sum + (record.courses?.length || 0), 0) }}
                          </a-table-summary-cell>
                          <a-table-summary-cell :index="3" align="right">
                            {{ filteredDataSource.reduce((sum, record) => {
                              return sum + (record.courses?.reduce((s, course) => s + (course.purchasedHours || 0), 0) || 0)
                            }, 0) }}
                          </a-table-summary-cell>
                          <a-table-summary-cell :index="4" align="right">
                            {{ filteredDataSource.reduce((sum, record) => {
                              return sum + (record.courses?.reduce((s, course) => {
                                return s + (course.totalConsumed || (course.normalConsumed || 0) + (course.leaveDeducted || 0) || course.consumedHours || 0)
                              }, 0) || 0)
                            }, 0) }}
                          </a-table-summary-cell>
                          <a-table-summary-cell :index="5" align="right">
                            {{ filteredDataSource.reduce((sum, record) => {
                              return sum + (record.courses?.reduce((s, course) => s + (course.remainingHours || 0), 0) || 0)
                            }, 0) }}
                          </a-table-summary-cell>
                        </a-table-summary-row>
                      </a-table-summary>
                    </template>
                  </a-table>
                </div>
              </div>
            </div>
          </a-tab-pane>
          <a-tab-pane key="2" tab="按老师">
            按老师
          </a-tab-pane>
          <a-tab-pane key="3" tab="按课程">
            按课程
          </a-tab-pane>
        </a-tabs>
      </div>
    </div>
  </template>
  
  <style scoped lang="less">
  .home {
    color: #666;
  
    .tabs {
      width: 100%;
      border-radius: 10px;
      line-height: 40px;
  
      :deep(.ant-tabs-nav) {
        background: #fff;
        border-radius: 16px;
        margin: 0;
      }
  
      :deep(.ant-tabs-nav-wrap) {
        padding-left: 36px;
      }
  
      :deep(.ant-tabs-ink-bar) {
        text-align: center;
        height: 9px !important;
        background: transparent;
        bottom: 1px !important;
  
        &::after {
          position: absolute;
          top: 0;
          left: calc(50% - 12px);
          width: 24px !important;
          height: 4px !important;
          border-radius: 2px;
          background-color: var(--pro-ant-color-primary);
          content: "";
        }
      }
  
      .twoTab {
        padding: 0;
  
        :deep(.ant-tabs-nav-wrap) {
          padding-left: 10px;
          margin: 6px 0;
        }
  
        :deep(.ant-tabs-nav) {
          &::before {
            display: none;
          }
        }
  
        :deep(.ant-tabs-tab) {
          padding: 6px 14px !important;
        }
  
        :deep(.ant-tabs-tab-active) {
          background: #e6f0ff;
          border-radius: 8px;
        }
  
        :deep(.ant-tabs-ink-bar) {
          display: none;
        }
      }
    }

    .content-wrapper {
      background: #f5f5f5;
      min-height: calc(100vh - 200px);
    }
  }

  .card-white {
    background: #fff;
    margin-top: 8px;
    padding: 12px;
    border-radius: 12px;

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
  }

  .filter-wrap {
    margin-top: 8px;
    border-radius: 12px;
  }

  .nowrap {
    flex-wrap: nowrap !important;
    overflow-x: auto;
  }

  .expanded-row-content {
    padding: 20px 24px;
    background: #fafbfc;
    border-left: 3px solid var(--pro-ant-color-primary);
  }

  .expanded-row-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    padding-bottom: 12px;
    border-bottom: 1px solid #e8e8e8;
  }

  .expanded-row-title {
    font-size: 14px;
    font-weight: 600;
    color: #222;
  }

  .expanded-row-count {
    font-size: 12px;
    color: #666;
  }

  .expanded-table {
    :deep(.ant-table) {
      background: #fff;
    }

    :deep(.ant-table-thead > tr > th) {
      background: #f5f7fa;
      font-weight: 500;
      color: #333;
    }

    :deep(.ant-table-tbody > tr:hover > td) {
      background: #f5f7fa;
    }
  }

  .table-content {
    :deep(.summary-row) {
      background-color: #f0f7ff;
      font-weight: bold;

      td,
      th {
        background-color: #f0f7ff !important;
      }
    }
  }
  </style>