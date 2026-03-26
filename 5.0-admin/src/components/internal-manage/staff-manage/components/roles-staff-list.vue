<script setup>
import { QuestionCircleOutlined } from '@ant-design/icons-vue'
import dayjs from 'dayjs'
import { getStaffListByRoleId } from '@/api/internal-manage/role-manage'

const props = defineProps({
  roleId: {
    type: [String, Number],
    default: '',
  },
})

const columns = [
  {
    title: '员工姓名',
    dataIndex: 'nickName',
    key: 'nickName',
  },
  {
    title: '账号状态',
    dataIndex: 'disabled',
    key: 'disabled',
  },
  {
    title: '创建日期',
    dataIndex: 'createTime',
    key: 'createTime',
  },
  // {
  //   title: "操作",
  //   key: "action",
  //   width: 100,
  // },
]

const tableData = ref([])
const loading = ref(false)

// 获取任职员工列表
async function getStaffList() {
  if (!props.roleId) {
    return
  }

  try {
    loading.value = true
    const { result } = await getStaffListByRoleId({ roleId: props.roleId })
    tableData.value = result || []
  }
  catch (error) {
    console.error('获取员工列表失败:', error)
    tableData.value = []
  }
  finally {
    loading.value = false
  }
}

// 员工详情
function handleDetail(record) {
  // 这里可以触发详情弹窗或跳转详情页
  console.log('查看员工详情:', record)
}

// 监听角色ID变化
watch(
  () => props.roleId,
  () => {
    getStaffList()
  },
  { immediate: true },
)

onMounted(() => {
  getStaffList()
})
</script>

<template>
  <div class="roles-staff-list px-16px">
    <a-table
      :columns="columns"
      :data-source="tableData"
      :pagination="false"
      :loading="loading"
      :scroll="{ x: 600 }"
    >
      <template #bodyCell="{ column, record }">
        <!-- 员工姓名列 -->
        <template v-if="column.key === 'nickName'">
          <div class="roleStaffBox">
            <div class="avatar">
              <div class="imgSpace">
                <div>
                  {{ record.nickName ? record.nickName.slice(0, 1) : "U" }}
                </div>
              </div>
            </div>
            <div class="staffName">
              <div class="staffNameText">
                <a-tooltip placement="top">
                  <template #title>
                    <span>{{ record.nickName }}</span>
                  </template>
                  {{ record.nickName }}
                </a-tooltip>
              </div>
              <div class="phoneBox">
                <span>{{ record.mobile }}</span>
                <a-popover placement="top">
                  <template #title>
                    <div style="border-bottom:1px solid #eee;padding-bottom: 6px;">
                      说明
                    </div>
                  </template>
                  <template #content>
                    <div class="w-400px">
                      未激活：当前员工未激活登录过系统，如手机号不正确，员工自己可登录 App 点击“我的-个人信息-账号与安全”修改手机号码或超级管理员修改员工手机号码。
                    </div>
                  </template>
                  <span v-if="!record.activatedStatus" class="cursor-pointer text-12px" style="color: #ff3333">未激活 <QuestionCircleOutlined style="color: #ff3333;" />
                  </span>
                </a-popover>

                <!-- 如果是超级管理员，显示标签 -->
                <span
                  v-if="
                    (record.isAdmin
                      && record.manage
                      && record.activatedStatus)
                      || record.roleName?.includes('超级管理员')
                  "
                  class="admin-tag"
                >
                  超级管理员
                </span>
              </div>
            </div>
          </div>
        </template>

        <!-- 账号状态列 -->
        <template v-else-if="column.key === 'disabled'">
          <span
            :class="
              record.disabled
                ? 'text-#ff3333 bg-#ffe6e6'
                : 'text-#06f bg-#e6f0ff'
            "
            class="status-tag"
          >
            {{ record.disabled ? "已离职" : "在职中" }}
          </span>
        </template>

        <!-- 创建日期列 -->
        <template v-else-if="column.key === 'createTime'">
          <span>
            {{
              record.createTime
                ? dayjs(record.createTime).format("YYYY-MM-DD HH:mm")
                : "-"
            }}
          </span>
        </template>

        <!-- 操作列 -->
        <!-- <template v-else-if="column.key === 'action'">
          <a-button type="link" size="small" @click="handleDetail(record)">
            详情
          </a-button>
        </template> -->
      </template>
    </a-table>
  </div>
</template>

<style lang="less" scoped>
.roles-staff-list {
  .roleStaffBox {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .avatar {
    .imgSpace {
      position: relative;
      box-sizing: content-box;
      width: 40px;
      height: 40px;
      overflow: hidden;
      border-radius: 50%;

      div {
        width: 40px;
        height: 40px;
        line-height: 40px;
        text-align: center;
        background-color: #005ce6;
        font-size: 16px;
        color: #fff;
        border-radius: 50%;
      }
    }
  }

  .staffName {
    flex: 1;
    min-width: 0;

    .staffNameText {
      font-size: 14px;
      font-weight: 400;
      color: #222;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      line-height: 20px;
    }

    .phoneBox {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-top: 2px;

      span {
        font-size: 12px;
        color: #666;
        line-height: 16px;
      }

      .admin-tag {
        background: #fff7e6;
        color: #fa8c16;
        padding: 2px 6px;
        border-radius: 4px;
        font-size: 12px;
        border: 1px solid #ffd591;
      }
    }
  }

  .status-tag {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 10px;
    font-size: 12px;
    font-weight: 500;
    text-align: center;
  }
}
</style>
