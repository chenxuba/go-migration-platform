<script setup>
const displayArray = ref(['intention', 'followStatus', 'sex', 'createUser', 'createTime', 'intentionCourse', 'reference'])
const dataSource = ref([
  {
    key: '1',
    name: '胡彦斌',
    phone: 17601241636,
    intentionCourse: '初级言语课、高级感统课、中级认知课',
    channelType: '外部渠道',
    channel: '抖音',
    teacher: '张晨',
    status: '已邀约',
    followed: '2025-03-31 17:09',
    nextTime: '2025-03-31 17:09',
    createTime: '2025-03-31 17:09',
    createUser: '张晨',
    putType: '否',
    putPeo: '-',
    birthday: '2022-09-23',
    wxchat: '1115009958',
    grade: '一年级',
    school: '上海市第一人民小学',
    address: '上海市杨浦区纪念路8号财大科技园区5号楼102A',
    IDcard1: 'CL202209229932',
    IDcard2: '37292520220922883X',
  },
])
const allColumns = ref([
  {
    title: '学员/性别/年龄',
    dataIndex: 'name',
    key: 'name',
    fixed: 'left',
    width: 120,
    required: true, // 新增必选标识
  },
  {
    title: '联系电话',
    dataIndex: 'phone',
    width: 120,
    key: 'phone',
  },
  {
    title: '学员状态',
    dataIndex: 'studentStatus',
    key: 'studentStatus',
    width: 120,
  },
  {
    title: '家校云',
    dataIndex: 'cloud',
    key: 'cloud',
    width: 120,
  },
])

const rowSelection = {
  onChange: (selectedRowKeys, selectedRows) => {
    console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', selectedRows)
  },
}
// 从本地存储读取已保存的列配置
const savedSelected = localStorage.getItem('waitFocus')
const keysArray = allColumns.value
  .map(column => column?.key) // 可选链操作符
  .filter(key => typeof key !== 'undefined') // 过滤未定义的值
const initialSelectedValues = savedSelected
  ? JSON.parse(savedSelected)
  : keysArray

// 选中的列（初始化包含重要字段）
const selectedValues = ref(initialSelectedValues)
// 生成字段选择选项（排除操作列）
const columnOptions = computed(() =>
  allColumns.value
    .filter(col => col.key !== 'action')
    .map(col => ({
      id: col.key,
      value: col.title,
      disabled: col.required, // 禁用必选字段
    })),
)
// 过滤后的列（自动包含必选列）
const filteredColumns = computed(() => {
  const requiredColumns = allColumns.value.filter(col => col.required)
  const optionalColumns = allColumns.value
    .filter(col =>
      selectedValues.value.includes(col.key)
      && !col.required,
    )

  // 保持固定列顺序：left -> normal -> right
  return [
    ...requiredColumns.filter(col => col.fixed === 'left'),
    ...optionalColumns,
    ...requiredColumns.filter(col => col.fixed === 'right'),
  ]
})
// 强制包含必选字段的监听
watch(selectedValues, (newVal) => {
  const requiredKeys = allColumns.value
    .filter(col => col.required)
    .map(col => col.key)

  // 自动补全必选字段
  if (!requiredKeys.every(k => newVal.includes(k))) {
    selectedValues.value = Array.from(new Set([
      ...newVal.filter(v => !requiredKeys.includes(v)),
      ...requiredKeys,
    ]))
  }
}, { deep: true })
// 自动保存列配置到本地存储
watch(selectedValues, (newVal) => {
  localStorage.setItem('waitFocus', JSON.stringify(newVal))
}, { deep: true })
// 表格总宽度计算
const totalWidth = computed(() =>
  filteredColumns.value.reduce((acc, column) => acc + (column.width || 0), 0),
)
</script>

<template>
  <div>
    <!-- 学员筛选条件 -->
    <div class="filter-wrap mt-2 bg-white  pl-3 pr-3 rounded-4">
      <all-filter :display-array="displayArray" :is-quick-show="false" :is-show-search-stu-phone="false" />
    </div>
    <div class="student-list mt-2 pt-3 pb-3 pl-6 pr-6 bg-white rounded-4">
      <div class="tab-table">
        <div class="table-title flex justify-between">
          <div class="total">
            当前共{{ dataSource.length }}名学员
          </div>
          <div class="edit flex">
            <a-button class="mr-2">
              二维码邀请
            </a-button>
            <a-button class="mr-2" type="primary">
              短信邀请
            </a-button>
            <!-- 自定义字段 -->
            <customize-code
              v-model:checked-values="selectedValues" :options="columnOptions" :total="allColumns.length"
              :num="selectedValues.length"
            />
          </div>
        </div>
        <div class="table-content mt-2">
          <a-table
            :data-source="dataSource" :pagination="dataSource.length > 10" :columns="filteredColumns" :row-selection="rowSelection"
            :scroll="{ x: totalWidth }" size="small"
          >
            <template #bodyCell="{ column, record }">
              <template v-if="column.key === 'name'">
                <student-avatar
                  name="龙龙"
                  gender="男"
                  :show-age="false"
                  default-active-key="0"
                />
              </template>
              <template v-if="column.key === 'phone'">
                <div class="name">
                  <div class="text-#222">
                    爸爸
                  </div>
                  <div class="text-3 text-#666">
                    176****1636
                  </div>
                </div>
              </template>
              <template v-if="column.key === 'studentStatus'">
                <div class="flex flex-items-center studentStatus">
                  <span class="dot" />
                  <span>在读学员{{ record.a }}</span>
                </div>
              </template>
              <template v-if="column.key === 'cloud'">
                <a-tooltip placement="right">
                  <template #title>
                    <span>点击邀请关注</span>
                  </template>
                  <div class="flex flex-items-center cursor-pointer">
                    <span class="whitespace-nowrap text-#ccc">
                      未关注
                    </span>
                    <svg width="16px" height="16px" class="ml-2" viewBox="0 0 16 16">
                      <g id="\u9875\u9762-2" stroke="none" stroke-width="1" fill="none" fill-rule="evenodd">
                        <g
                          id="\u753B\u677F\u5907\u4EFD-21" transform="translate(-474.000000, -608.000000)"
                          fill="#CCCCCC"
                        >
                          <g id="Rectangle-2\u5907\u4EFD-89" transform="translate(398.000000, 580.000000)">
                            <g id="\u7F16\u7EC4" transform="translate(76.000000, 21.600000)">
                              <g id="\u7F16\u7EC4" transform="translate(0.000000, 6.400000)">
                                <path
                                  id="\u8DEF\u5F84"
                                  d="M12.5488957,14.2844713 L11.5010486,14.2844713 C11.1341596,14.280754 10.8398076,13.9883197 10.843536,13.6312425 C10.8398076,13.2741654 11.1341596,12.9817311 11.5010486,12.9780138 L12.5488957,12.9780138 C13.1929132,12.9707828 13.7094253,12.457622 13.7035882,11.8308133 L13.7035882,5.51659915 C13.7049584,5.07149643 13.4426881,4.66546656 13.0299588,4.47372986 L8.49973266,2.41098807 C8.19497588,2.2717625 7.84236314,2.2717625 7.53760636,2.41098807 L3.00725203,4.47372986 C2.59455747,4.66549483 2.33231941,5.07151473 2.33368895,5.51659915 L2.33368895,8.11331051 C2.33741739,8.47038769 2.04306536,8.76282195 1.67617635,8.76653928 C1.30928733,8.76282195 1.0149353,8.47038769 1.01862871,8.11331051 L1.01862871,5.51659915 C1.01573797,4.56462047 1.57664141,3.69619985 2.4593492,3.28605311 L6.98970297,1.22331132 C7.64157303,0.925562892 8.39577156,0.925562892 9.04764162,1.22331132 L13.5778672,3.28605311 C14.460609,3.69617215 15.0215439,4.56460257 15.0186287,5.51659915 L15.0186287,11.8308133 C15.0186287,13.1837748 13.9107309,14.2844713 12.5488957,14.2844713 Z"
                                />
                                <path
                                  id="\u8DEF\u5F84"
                                  d="M1.56733162,10.2194036 C1.40127346,10.2195233 1.23544109,10.2313173 1.07112909,10.2546935 C1.02383678,10.4730961 1,10.6956916 1,10.9188916 C1,11.7700045 1.34739282,12.5862583 1.96575882,13.1880863 C2.58412481,13.7899143 3.42280952,14.1280178 4.29731162,14.1280178 C4.51607755,14.1280178 4.73430678,14.1069519 4.94880175,14.0650857 C4.97598308,13.8947261 4.98963678,13.7225811 4.98963678,13.5501797 C4.98963678,12.666803 4.6290758,11.8196066 3.98726883,11.1949647 C3.34546185,10.5703228 2.4749842,10.2194036 1.56733162,10.2194036 Z"
                                />
                                <path
                                  id="\u8DEF\u5F84"
                                  d="M4.04965057,14.1112242 C4.36580361,14.1624804 4.68574686,14.1883875 5.00625618,14.1886844 C6.55014367,14.1886844 8.03079835,13.5917848 9.12249298,12.529291 C10.2141876,11.4667972 10.8274979,10.0257453 10.8274979,8.5231503 C10.8271446,8.25723104 10.8076999,7.99166078 10.7693057,7.72837983 C10.4531526,7.67712408 10.1332094,7.65121719 9.81270009,7.65092023 C8.26881275,7.65092023 6.78815822,8.24781981 5.69646369,9.3103135 C4.60476916,10.3728072 3.99145897,11.813859 3.99145897,13.3164538 C3.99181199,13.582373 4.01125654,13.8479433 4.04965057,14.1112242 Z"
                                />
                              </g>
                            </g>
                          </g>
                        </g>
                      </g>
                    </svg>
                  </div>
                </a-tooltip>
              </template>
            </template>
          </a-table>
        </div>
      </div>
    </div>
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
