<!-- 新增员工 -->
<script setup>
import { ref, reactive, watch } from 'vue'
import { CloseOutlined } from '@ant-design/icons-vue'
import messageService from '~@/utils/messageService'
import { batchModifyRole } from '@/api/internal-manage/staff-manage'
import { getInstRolePageApi } from '@/api/internal-manage/role-manage'
import selectRole from './selectRole.vue'
import rolesDetailsDrawer from '@/components/common/roles-details-drawer.vue'
import customTitle from '@/components/common/custom-title.vue'
import { debounce } from 'lodash-es'

const { batchUserList } = defineProps({
  batchUserList: {
    type: Array,
    default: () => [],
  },
})
const emit = defineEmits(['success', 'openRole'])
const loading = ref(false)
const selectRoleList = ref([])
const roleInfos = ref([])
const open = defineModel({
  type: Boolean,
  default: false,
})

const formState = reactive({
  roleIds: [],
})

const addEmployeesOpen = ref(false)

const form = ref(null)

function cancel() {
  open.value = false
  selectRoleList.value = []
  form.value?.resetFields()
}

const handleOk = debounce(() => {
  form.value?.validate().then((res) => {
    if (res) {
      loading.value = true
      batchModifyRole({ ...formState, userIds: batchUserList.map(item => item.id) }).then((res) => {
        messageService.success('修改成功')
        emit('success')
      }).finally(() => {
        cancel()
        loading.value = false
      })
    }
    else {
      console.log(res)
    }
  })
}, 500, { leading: true, trailing: false })

function getRoleList(query = { queryModel: {} }) {
  const pages = {
    pageRequestModel: {
      needTotal: true,
      pageSize: 50,
      pageIndex: 1,
      skipCount: 1,
    },
  }
  getInstRolePageApi({ ...pages, ...query }).then(({ result }) => {
    roleInfos.value = result
  })
}

function handleSelectRoleSuccess(roleIds) {
  selectRoleList.value = roleIds
  formState.roleIds = roleIds
}

const roleDetailsOpen = ref(false)
const roleDetailsInfo = ref({})

function roleDetailsFunc(roleInfo) {
  roleDetailsOpen.value = true
  roleDetailsInfo.value = roleInfo
}

watch(open, (val) => {
  if (val) {
    getRoleList()
  }
})
</script>

<template>
  <a-modal
    v-model:open="open" :mask-closable="false" :keyboard="false" width="800px" :destroy-on-close="true"
    :closable="false" :centered="true" :confirm-loading="loading" @cancel="cancel" @ok="handleOk"
  >
    <template #title>
      <div class=" text-5 flex justify-between flex-center">
        <span>批量修改任职角色</span>
        <a-button type="text" class="close-btn" @click="cancel">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <div>
      <div class="bg-#fafafa rd-10px">
        <div class="p-20px">
          员工：共 <span class="text-#06f">{{ batchUserList.length }}</span> 人，<span class="text-#888">{{
            batchUserList.map(item => item.nickName).join('、') }}</span>
        </div>
        <div class="w-100% h-.5px bg-#ddd" />
        <div class="p-20px">
          <a-form ref="form" :model="formState" autocomplete="off">
            <a-form-item
              style="margin: 0;" class="position" name="roleIds"
              :rules="[{ required: true, message: '请设置任职角色' }]" :colon="false"
            >
              <template #label>
                <div class="flex items-center">
                  <div>任职角色：</div>
                  <a-button type="primary" @click="addEmployeesOpen = true">
                    编辑
                  </a-button>
                </div>
              </template>
            </a-form-item>
          </a-form>
          <template v-if="selectRoleList.length !== 0">
            <div class="m-t-20px border-1px border-solid border-gray-200 rounded-10px">
              <div class="p15px">
                <custom-title title="总机构" font-size="16px" font-weight="500" />
              </div>

              <template v-for="roleInfo in roleInfos" :key="roleInfo.id">
                <div
                  v-if="selectRoleList.includes(roleInfo.id)"
                  class="border-t-1px border-#eee border-solid border-0px"
                >
                  <div class="flex justify-between px-25px py-15px">
                    <div class="flex items-start gap-10px">
                      <div>
                        <div class="font-bold">
                          {{ roleInfo.roleName }}
                        </div>
                        <div>
                          <span class="text-#06f font-500">{{ roleInfo.functionalAuthorityCount }}</span>
                          <span>个功能权限,</span>
                          <span class="text-#06f font-500"> {{ roleInfo.dataAuthorityCount }} </span>
                          <span>个数据权限</span>
                        </div>
                        <div class="w-500px overflow-hidden text-ellipsis whitespace-nowrap">
                          <span class="text-#666 text-13px">角色描述：{{ roleInfo.description }}</span>
                        </div>
                      </div>
                    </div>
                    <div
                      class="min-w-95px text-#06f cursor-pointer text-12px px-8px max-h-25px border-#06f border-solid border-1px flex items-center justify-center rounded-15px text-center"
                      @click="roleDetailsFunc(roleInfo)"
                    >
                      角色权限详情
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </template>
        </div>
      </div>
    </div>
    <selectRole
      v-model="addEmployeesOpen" :role-infos="roleInfos" :select-role-list="selectRoleList"
      @success="handleSelectRoleSuccess"
    />
    <roles-details-drawer v-model:open="roleDetailsOpen" :role-id="roleDetailsInfo.id" :details="roleDetailsInfo" @update:open="roleDetailsOpen = $event" />
  </a-modal>
</template>

<style scoped lang="less">
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

.position {
  margin-bottom: 0;
}

:deep(.ant-form-item-explain-error) {
  font-size: 12px;
}

:deep(.ant-modal-body) {
  padding: 12px 24px;
}

:deep(.modal-wrap .ant-modal) {
  top: 30px;
}
</style>
