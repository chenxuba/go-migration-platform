<script setup>
import createNoticeModel from './components/createNoticeModel.vue'
import { useTableColumns } from '@/composables/useTableColumns'

const createNotice = ref(false)

function handelCreateNotice() {
  createNotice.value = true
}

const displayArray = ref([
  'intention',
  'followStatus',
  'sex',
  'createUser',
  'applyTime',
  'intentionCourse',
  'reference',
  'studentStatus',
  'classEndingTime',
  'classStopTime',
])
const dataSource = ref([{ key: 1 }, { key: 2 }])
// const defaultCreateTimeVals = ref(["2025-04-01", "2025-04-13"]);
const allColumns = ref([
  {
    title: '通知标题',
    dataIndex: 'noticeTitle',
    key: 'noticeTitle',
    width: 180,
  },
  {
    title: '摘要内容',
    dataIndex: 'abstractContent',
    key: 'abstractContent',
    width: 180,
  },
  {
    title: '通知范围',
    dataIndex: 'noticeScope',
    key: 'noticeScope',
    width: 150,
  },

  {
    title: '公告状态',
    dataIndex: 'noticeStatus',
    key: 'noticeStatus',
    width: 150,
  },
  {
    title: '家长端已读',
    dataIndex: 'parentRead',
    key: 'parentRead',
    width: 150,
  },
  {
    title: '创建人',
    dataIndex: 'createUser',
    key: 'createUser',
    width: 150,
  },
  {
    title: '发布时间',
    dataIndex: 'publishTime',
    key: 'publishTime',
    width: 150,
  },
  {
    title: '操作',
    dataIndex: 'action',
    key: 'action',
    fixed: 'right',
    width: 130,
  },
])
const { selectedValues, columnOptions, filteredColumns, totalWidth }
  = useTableColumns({
    storageKey: 'notice-list', // 本地存储键名
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
</script>

<template>
  <div>
    <div class="bg-white px5 py4 rounded-4">
      <div class="t flex-center justify-between">
        <div>
          <span class="text-5 text-#000 font800">通知公告模板</span>
          <span class="text-3.5 text-#666 ml3">多种模板，一键群发，已读未读及时跟进</span>
        </div>
        <div class="bg-#0066ff14 py2 px2.5 rounded-2 after">
          <img
            width="14"
            src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAB4AAAAgCAYAAAAFQMh/AAAAAXNSR0IArs4c6QAABIJJREFUWEe1ln1sE3UYx7/PvdDCnC8Mx0RM3DSRoBIQM9Z2aFG2dajbnBkxkQjEdyEmKo46iXbRZB3E+DoTjFETEqIh4rbo2g0iZLNtiJk6/yC+ZmIMOpYpDJm99nqPuaZX29rrVan31909z/P93O95nt/vHoLFxbsGq6HyMBgT6G3ZSCDODJnyco1IWEEaTgEor+ilYStN3U5WTrxz4AMQ2pN+zPdQb9v+zJgZLzcyYMDeXeSnrVaalmDuHKyDwJG0EGMS85Vl5NsYM96d8vIqAfg8Jbanwk+d5wVm3xEJ0dmjAFw5Qp3kb91jvJvu4nWkYSsxpjQBc5f20HP/GcxdgzdA41cA1OcR0UDcB2jd1NM+Yw5hcnrijzAJ1xOJ/vAQncj0zaox7zhYCUncDeBey/ozTkPA87B98TL5fFqmaEcHiydn428zka6jN8eJi0i+JhAgxfDLBj89MAJGQzGp+rvuvJ162/qM5+Zmtp3R1PdAaMvRcYaDcrpfssHegb0AHvwXYA2MFupt/ViPcbv5gphdHQBwS6YGA+OROqkWPkpnJhvs8wlQVm4G0wsAllh8wCigPUH+O8d1P0cTLyRSAwBqs+Pox7ggOj4bol9Na2wYeMdwGWRlF5i9eeAzAD9A/rYPDZurgZewqI4AuDbHf4aQcIWC9m9ydQoeIOztfwegLVlBRHdQT8tHxrv626I1iYR0mMDVOeJ/sob1kRE5nC9zhcHP9F+BBH0LwJ4KHiV/682GkLNRuQ6CoK/0shxxjYnuigSkfrNyWR+Z3gF9ez2VFCCuo562Y/pt/YbYGk0jvaaX/FOct4WD894o1CPWYL3eYvQlgCeMbVPXFF9PhH4CynLFGeiJBOUuq51hCc4VcHnUdgbrPwpbnpXuCwflzXpqSgp2emJbAHoLgJinSw9Hp6UN4+MUt4Imq1aMk+7j8sQeZ9CLeWMYX5JNuik0SGeL1SsK7GiKdxPh2fyifEJk2TE2TL8Y9oe/P1vJotDNxGOCBpGJnKwp3jevWnjG8LEAMzmaE68S83aTlfzGmuaKjNi+zrQ/OvlHVQJ0Mp0dgsoJZVFRYLebpZg9rh8gm0ygUQIaQkH503z2hybP/QCgJmWb2FtdtjLTL++K3W62x+zq+wBaTKAagTpCQelgPvu2n2cr1Lg4ldGEytz8BRX7quicaapdLVzOseQfZp1poxA/Fg7Me83Mfv93c0tFiV9P1pghMsgpyYn7+pZemB4cslZceytXSLIaBHCjORS7wwF5Z7Hda+aXBjtu58uhqiMELDd1Jt4fCsibijkgrD4sCXY0Ra8mkg4BfGWBgE9Ol0vNxw9QesK0Ei9kpzUeZYUEYZiBqgJp+Uolae2xAM2eDyyrq52euD4TryrQSD+RKjtCh5L7smSXPoLqA1idieLvImn1YwHb8ZIRU0LkaFSWkSgcBWNxjrgCAY3hIXm01FBdL9lca5uV5QkWjgCoTEE0Bt0dCUoH/g9oGqzfpMYYfaJYDOInCx0QpfiYrANk9WqWF1Ti4rEATZdCvJDGX55jnDAUsnPxAAAAAElFTkSuQmCC"
            alt=""
          >
          更多模板
        </div>
      </div>
      <div class="overflow-auto">
        <a-space :size="14" class="mt3">
          <div v-for="(item, index) in 18" :key="index" class="templateItem w-39">
            <div class="h-39 bgimg">
              <div class="mask">
                <a-button type="primary">
                  使用
                </a-button>
                <div class="eye">
                  <svg width="17px" height="12px" viewBox="0 0 17 12">
                    <title>编组</title>
                    <g id="\u9875\u9762-1" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                      <g id="\u901A\u77E5\u516C\u544A" transform="translate(-1825.000000, -377.000000)">
                        <g id="\u7F16\u7EC4" transform="translate(1816.000000, 367.000000)">
                          <g transform="translate(9.000000, 10.000000)">
                            <path
                              id="\u5F62\u72B6"
                              d="M8.00044664,12 C3.58904385,12 0,9.30840603 0,6 C0,2.69159397 3.58922251,0 8.00044664,0 C12.4107775,0 16,2.69159397 16,6 C16,9.30840603 12.4107775,12 8.00044664,12 L8.00044664,12 Z M8.00044664,1.69822139 C4.55254196,1.69822139 1.64025146,3.66832095 1.64025146,6 C1.64025146,8.33167905 4.55254196,10.3017786 8.00044664,10.3017786 C11.447458,10.3017786 14.3595699,8.33167905 14.3595699,6 C14.3595699,3.66832095 11.4472794,1.69822139 8.00044664,1.69822139 Z"
                              fill="currentColor"
                            />
                            <path
                              id="\u8DEF\u5F84"
                              d="M5.44993691,5.99981505 C5.44993691,6.94299358 5.93599319,7.81452648 6.7250131,8.28611575 C7.51403302,8.75770502 8.48614564,8.75770502 9.27516555,8.28611575 C10.0641855,7.81452648 10.5502417,6.94299358 10.5502417,5.99981505 C10.5502417,5.05663652 10.0641855,4.18510361 9.27516555,3.71351434 C8.48614564,3.24192507 7.51403302,3.24192507 6.7250131,3.71351434 C5.93599319,4.18510361 5.44993691,5.05663652 5.44993691,5.99981505 L5.44993691,5.99981505 Z"
                              fill="currentColor"
                            />
                          </g>
                        </g>
                      </g>
                    </g>
                  </svg>
                </div>
              </div>
            </div>
            <div class="templateItemTitle text-#000 text-3.5 mt1 font-500">
              春季班课程安排
            </div>
          </div>
        </a-space>
      </div>
    </div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap bg-white pl-3 pr-3 rounded-4 mt3">
      <all-filter :display-array="displayArray" />
    </div>
    <div class="student-list mt-3 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共计 {{ dataSource.length }} 条通知公告
          </div>
          <div class="edit flex">
            <a-space>
              <a-button type="primary" @click="handelCreateNotice">
                创建通知
              </a-button>
            </a-space>
            <!-- 自定义字段 -->
            <!-- <customize-code v-model:checkedValues="selectedValues" :options="columnOptions"
                  :total="allColumns.length - 1" :num="selectedValues.length - 1" /> -->
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns"
            :scroll="{ x: totalWidth }" size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'noticeTitle'">
                <div class="text-#222">
                  51放假通知
                </div>
              </template>
              <template v-if="column.key === 'abstractContent'">
                <div class="text-#222 w-60%">
                  <clamped-text :lines="2" text="完成【静夜思】诗词抄写10遍，并完成背诵，家长录制背诵视频上传" />
                </div>
              </template>
              <template v-if="column.key === 'noticeScope'">
                <div class="text-#222">
                  全体学生
                </div>
              </template>
              <template v-if="column.key === 'noticeStatus'">
                <div class="text-#222">
                  已发布
                </div>
              </template>
              <template v-if="column.key === 'parentRead'">
                <div class="text-#222">
                  已读
                </div>
              </template>
              <template v-if="column.key === 'createUser'">
                <div class="text-#222">
                  陈瑞
                </div>
              </template>
              <template v-if="column.key === 'publishTime'">
                <div class="text-#222">
                  2025-04-16 (周三)
                </div>
                <div class="text-#888 text-3">
                  18:12
                </div>
              </template>
              <template v-if="column.key === 'action'">
                <a-space :size="14">
                  <a class="font500">编辑{{ record.a }}</a>
                  <a class="font500">复制</a>
                  <a class="font500">删除</a>
                </a-space>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
    <student-info-drawer v-model:open="openDrawer" />
    <order-detail-drawer v-model:open="openOrderDetailDrawer" />
    <createNoticeModel v-model="createNotice" />
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
  background: #0c3;
}

.tip {
  padding: 10px 24px 10px 14px;
  background: #e6f0ff;
  color: #333;

  a {
    color: var(--pro-ant-color-primary);
  }
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

.hover {
  &:hover {
    .name {
      color: var(--pro-ant-color-primary);
    }
  }
}

.after {
  position: relative;
  cursor: pointer;

  &::after {
    content: "";
    position: absolute;
    top: 7px;
    right: 4px;
    width: 8px;
    height: 8px;
    background: #ee1625;
    border: 1px solid #fff;
    border-radius: 50%;
  }
}

.templateItem {
  .bgimg {
    width: 100%;
    background: url("https://prod-cdn.schoolpal.cn/training/next-erp/h5/static/images/notice/cjbk2025.png");
    background-size: 100% 100%;
    border-radius: 8px;

    .mask {
      width: 100%;
      height: 100%;
      opacity: 0;
      display: flex;
      justify-content: space-between;
      padding: 12px;
      background: rgba(0, 0, 0, 0.1);
      border-radius: 8px;

      .eye {
        width: 34px;
        height: 32px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #666;
        cursor: pointer;
        background: #fff;
        border-radius: 8px;
        transition: all 0.3s ease;
        box-sizing: border-box;
        border: 1px solid #fff;

        &:hover {
          color: #06f;
          border: 1px solid #06f;
        }
      }
    }

    &:hover {
      cursor: pointer;

      .mask {
        opacity: 1;
      }
    }
  }
}

.overflow-auto {
  padding-bottom: 4px;

  &::-webkit-scrollbar {
    width: 10px;
    height: 6px;
    background: #eee;
  }

  /* 滚动条轨道背景 */
  &::-webkit-scrollbar-track {
    background: #eee;
    border-radius: 2px;
  }

  /* 滚动条滑块 */
  &::-webkit-scrollbar-thumb {
    background: #aaa;
    border-radius: 5px;
    border: 6px solid transparent;
    /* 增加留白效果 */
    background-clip: content-box;
  }

  /* 滑块悬停效果 */
  &::-webkit-scrollbar-thumb:hover {
    background: #ccc;
  }

  /* 滑块点击效果 */
  &::-webkit-scrollbar-thumb:active {
    background: #666;
  }

}
</style>
