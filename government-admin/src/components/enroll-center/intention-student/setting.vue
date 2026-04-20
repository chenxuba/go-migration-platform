<script setup>
import { ExclamationCircleFilled, FormOutlined, RightOutlined } from '@ant-design/icons-vue'
import { computed, onMounted, ref } from 'vue'
import { setInstConfigApi } from '~@/api/common/config'
import { useUserStore } from '~@/stores/user'

const emit = defineEmits(['diaplayPublicData'])
const openModel = ref(false)
const intentionLoading = ref(false)
const importLoading = ref(false)
const intentionWxLoading = ref(false)
const importWxLoading = ref(false)
const enablePublicPoolLoading = ref(false)
const tempUnfollowedTime = ref(1)

// Get the user store
const userStore = useUserStore()
const instConfig = ref({})

function changeSwitch(e) {
  emit('diaplayPublicData', e)
}

/**
 * Sets loading state, calls API, and refreshes config
 */
async function updateConfig(loadingRef, configUpdates = null) {
  try {
    if (loadingRef)
      loadingRef.value = true

    // Apply any updates to the config before saving
    if (configUpdates) {
      Object.assign(instConfig.value, configUpdates)
    }

    await setInstConfigApi(instConfig.value)
    await userStore.getInstConfig()
    instConfig.value = userStore.instConfig

    return true
  }
  catch (error) {
    console.error('Failed to update configuration:', error)
    return false
  }
  finally {
    if (loadingRef)
      loadingRef.value = false
  }
}

async function handleSetConfig(e) {
  // Determine which setting was changed and set its loading state
  let loadingRef = null
  const name = e.target?.name

  if (name === 'radioGroup1') {
    loadingRef = intentionLoading
  }
  else if (name === 'radioGroup2') {
    loadingRef = importLoading
  }
  else if (name === 'intentionWx') {
    loadingRef = intentionWxLoading
  }
  else if (name === 'importWx') {
    loadingRef = importWxLoading
  }

  await updateConfig(loadingRef)
}

async function handleSetPublicConfig(e) {
  if (!instConfig.value.enablePublicPool) {
    openModel.value = true
    tempUnfollowedTime.value = 1
    return
  }

  await updateConfig(enablePublicPoolLoading, {
    enablePublicPool: false,
    unfollowedTime: 0,
  })

  changeSwitch(e)
}

function handleSetting() {
  openModel.value = true
  tempUnfollowedTime.value = instConfig.value.unfollowedTime || 1
}

async function handleOkModal() {
  const success = await updateConfig(enablePublicPoolLoading, {
    enablePublicPool: true,
    unfollowedTime: tempUnfollowedTime.value,
  })

  if (success) {
    changeSwitch(instConfig.value.enablePublicPool)
    openModel.value = false
  }
}

const warningText = computed(() => {
  switch (instConfig.value.addIntentionStudentRule) {
    case 2:
      return '手机号重复的意向学员信息，均不允许录入！'
    case 3:
      return '姓名重复的意向学员信息，均不允许录入！'
    case 1:
    default:
      return '姓名 + 手机号都重复的意向学员信息，均不允许录入！'
  }
})

onMounted(async () => {
  try {
    if (!userStore.instConfig) {
      await userStore.getInstConfig()
    }

    instConfig.value = userStore.instConfig || {}
    tempUnfollowedTime.value = instConfig.value.unfollowedTime || 1
  }
  catch (error) {
    console.error('Failed to load configuration:', error)
  }
})
</script>

<template>
  <div class="tab-content mt-2">
    <!-- 意向学员录入设置 -->
    <div class="setting">
      <div class="title">
        意向学员录入设置
      </div>
      <div class="tips mt-2">
        <ExclamationCircleFilled class="mr-2" /> {{ warningText }}
      </div>
      <div class="table-wrap">
        <table border>
          <tbody>
            <tr>
              <td class="td1" rowspan="2">
                新增意向学员
              </td>
              <td>
                <a-radio-group
                  v-model:value="instConfig.addIntentionStudentRule" class="custom-radio"
                  name="radioGroup1" @change="handleSetConfig"
                >
                  <a-spin :spinning="intentionLoading">
                    <a-space class="flex flex-wrap">
                      <a-radio :value="1">
                        限制录入手机号和姓名同时相同的学员
                      </a-radio>
                      <a-radio :value="2">
                        限制录入手机号相同的学员
                      </a-radio>
                      <a-radio :value="3">
                        限制录入姓名相同的学员
                      </a-radio>
                    </a-space>
                  </a-spin>
                </a-radio-group>
              </td>
            </tr>
            <tr>
              <td>
                <a-spin :spinning="intentionWxLoading" style="width: 200px;">
                  <a-checkbox
                    v-model:checked="instConfig.limitSameWeChat" name="intentionWx"
                    @change="handleSetConfig"
                  >
                    限制录入微信号相同的学员
                  </a-checkbox>
                </a-spin>
              </td>
            </tr>
            <tr>
              <td class="td1" rowspan="2">
                导入学员
              </td>
              <td>
                <a-radio-group
                  v-model:value="instConfig.addImportStudentRule" class="custom-radio"
                  name="radioGroup2" @change="handleSetConfig"
                >
                  <a-spin :spinning="importLoading">
                    <a-space class="flex flex-wrap">
                      <a-radio :value="1">
                        限制录入手机号和姓名同时相同的学员
                      </a-radio>
                      <a-radio :value="2">
                        限制录入手机号相同的学员
                      </a-radio>
                      <a-radio :value="3">
                        限制录入姓名相同的学员
                      </a-radio>
                    </a-space>
                  </a-spin>
                </a-radio-group>
              </td>
            </tr>
            <tr>
              <td>
                <a-spin :spinning="importWxLoading" style="width: 200px;">
                  <a-checkbox
                    v-model:checked="instConfig.limitImportSameWeChat" name="importWx"
                    @change="handleSetConfig"
                  >
                    限制录入微信号相同的学员
                  </a-checkbox>
                </a-spin>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
  <div class="setting before mt-2">
    <div class="title">
      售前人员设置
    </div>
    <div class="table-wrap">
      <a class="font-800">去设置
        <RightOutlined />
      </a>
    </div>
  </div>
  <div class="tab-content mt-2">
    <!-- 意向学员录入设置 -->
    <div class="setting">
      <div class="title mb-2.5">
        公有池设置
      </div>
      <div class="table-wrap">
        <table border>
          <tbody>
            <tr>
              <td class="td1" rowspan="2">
                公有池
              </td>
              <td>
                <a-spin :spinning="enablePublicPoolLoading">
                  <div class="checked">
                    <span class="flex flex-start justify-start text-#666">
                      <a-switch
                        class="mr-2" :checked="instConfig.enablePublicPool" checked-children="开"
                        un-checked-children="关" @click="handleSetPublicConfig"
                      />
                      开启后，系统会自动将没有销售员的意向学员汇总到公有池，方便管理和再次分配</span>
                  </div>
                  <div v-if="instConfig.enablePublicPool" class="tip mt-5 pl-13">
                    未跟进时间超过 <span class="day">{{
                      instConfig.unfollowedTime }}</span>
                    天的意向学员将自动进入公有池
                    <FormOutlined class="icon" @click="handleSetting" />
                  </div>
                </a-spin>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
  <!-- 未跟进天数进公有池设置model -->
  <a-modal v-model:open="openModel" centered title="未跟进天数设置" width="437px" @ok="handleOkModal">
    <div class="setting-wrap">
      <div class="settingPoolCont flex flex-center justify-start">
        未跟进<a-input-number
          v-model:value="tempUnfollowedTime"
          class="w-34 ml-2 mr-2" :min="1"
        /> 天
      </div>
      <div class="setting-tip mt-3 text-#888">
        设置后，超过未跟进时间的学员进入公有池
      </div>
    </div>
  </a-modal>
</template>

<style lang="less" scoped>
.tab-content {
  background: #fff;
  border-radius: 12px;
  padding: 12px 20px;

  /* 自定义镂空样式 */
  .custom-radio ::v-deep(.ant-radio-wrapper:hover .ant-radio),
  .custom-radio ::v-deep(.ant-radio:hover .ant-radio-inner),
  .custom-radio ::v-deep(.ant-radio-input:focus + .ant-radio-inner) {
    border-color: var(--pro-ant-color-primary);
  }

  .custom-radio ::v-deep(.ant-radio-inner) {
    background-color: transparent;
    border-color: #d9d9d9;
  }

  .custom-radio ::v-deep(.ant-radio-checked .ant-radio-inner) {
    background-color: transparent;
    border-color: var(--pro-ant-color-primary);
  }

  .custom-radio ::v-deep(.ant-radio-inner::after) {
    background-color: var(--pro-ant-color-primary);
    transform: scale(0.5);
  }

  .setting {

    .title {
      font-size: 18px;
      color: #222;
      font-weight: 600;
      display: flex;
      align-items: center;

      &::before {
        display: inline-block;
        content: '';
        width: 4px;
        height: 14px;
        background: var(--pro-ant-color-primary);
        border-radius: 100px;
        margin-right: 6px;
      }
    }

    .tips {
      height: 40px;
      display: flex;
      align-items: center;
      background: #e6f0ff;
      color: #06f;
      padding: 0 24px;
      font-size: 14px;
      margin-bottom: 8px;
      border-radius: 4px;
      justify-content: flex-start;
    }

    .table-wrap {
      a {
        display: flex;
        align-items: center;
      }

      table {
        width: 100%;
        border-collapse: collapse;
        border: 1px solid #eee;
        border-radius: 8px;

        tr,
        td {
          border: 1px solid #eee;
        }

        td {
          padding: 18px 24px;
        }

        .td1 {
          width: 180px;
          text-align: center;
          font-size: 14px;
          font-family: PingFangSC-Regular, PingFang SC;
          font-weight: 400;
          color: #222;
        }
      }

      .day {
        color: var(--pro-ant-color-primary);
        font-weight: bold;
        font-size: 14px;
      }

      .icon {
        color: var(--pro-ant-color-primary);
        margin-left: 4px;
        cursor: pointer;
      }
    }
  }
}

.before {
  background: #fff;
  display: flex;
  justify-content: space-between;
  padding: 18px 24px;
  border-radius: 12px;
  align-items: center;

  .title {
    font-size: 18px;
    color: #222;
    font-weight: 600;
    display: flex;
    align-items: center;

    &::before {
      display: inline-block;
      content: '';
      width: 4px;
      height: 14px;
      background: var(--pro-ant-color-primary);
      border-radius: 100px;
      margin-right: 6px;
    }
  }

  .table-wrap {
    font-size: 14px;

    a {
      color: var(--pro-ant-color-primary);
    }
  }
}

.setting-wrap {
  background: #f6f7f8;
  border-radius: 8px;
  padding: 16px;
}
</style>
