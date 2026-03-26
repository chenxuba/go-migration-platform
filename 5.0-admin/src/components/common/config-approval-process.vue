<script setup>
import { CloseOutlined } from '@ant-design/icons-vue'
import { computed, ref, watch } from 'vue'
import messageService from '../../utils/messageService'
import StaffSelect from './staff-select.vue'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: '配置审批流程',
  },
  flowModels: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['update:open', 'save'])

// Add form reference
const formRef = ref()

// 处理双向绑定
const openModal = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

// 初始表单状态
const initialState = {
  approvalLevels: [
    {
      title: '一级审批人',
      step: 1,
      approvers: [],
      approverNames: [],
    },
  ],
}

// 定义表单状态
const formState = ref({ ...initialState })

// 监听弹窗打开状态，重置表单
watch(() => props.open, (newVal) => {
  if (newVal) {
    if (Array.isArray(props.flowModels) && props.flowModels.length > 0) {
      formState.value = {
        approvalLevels: props.flowModels.map(item => ({
          title: getLevelTitle(item.step),
          step: item.step,
          approvers: Array.isArray(item.staffIds) ? item.staffIds.map(id => String(id)) : [],
          approverNames: Array.isArray(item.staffNames) ? [...item.staffNames] : [],
        })),
      }
    }
    else {
      formState.value = JSON.parse(JSON.stringify(initialState))
    }
  }
})

function closeFun() {
  openModal.value = false
}

// 添加审批层级
function addLevel() {
  if (formState.value.approvalLevels.length < 10) {
    const levelNumber = formState.value.approvalLevels.length + 1
    const levelTitle = getLevelTitle(levelNumber)

    formState.value.approvalLevels.push({
      title: levelTitle,
      step: levelNumber,
      approvers: [],
      approverNames: [],
    })
  }
  else {
    // 使用自定义消息服务显示警告
    messageService.warning('最多可添加10级审批人', {
      duration: 3000,
    })

    // 添加抖动效果到添加按钮
    const addButton = document.querySelector('.stepItemAdd')
    if (addButton) {
      // 先移除动画类以便可以重新触发
      addButton.classList.remove('shake-animation')
      // 触发浏览器重绘
      void addButton.offsetWidth
      // 添加动画类
      addButton.classList.add('shake-animation')
    }
  }
}

// 删除审批层级
function removeLevel(index) {
  formState.value.approvalLevels.splice(index, 1)
  // 更新所有层级的标题
  formState.value.approvalLevels.forEach((level, idx) => {
    level.step = idx + 1
    level.title = getLevelTitle(idx + 1)
  })
}

// 获取层级标题
function getLevelTitle(level) {
  const titles = ['一', '二', '三', '四', '五', '六', '七', '八', '九', '十']
  return `${titles[level - 1]}级审批人`
}

function handleSubmit() {
  // 使用表单验证
  formRef.value.validate().then(() => {
    emit('save', formState.value.approvalLevels.map((level) => {
      return {
        step: level.step,
        staffIds: [...level.approvers],
        staffNames: [...(level.approverNames || [])],
      }
    }))

    // 关闭弹窗（会自动触发watch重置表单）
    closeFun()
  }).catch((errors) => {
    // 验证失败，错误信息在errors中
    console.log('表单验证失败:', errors)

    // 只查找第一个未配置的审批人
    let firstInvalidLevel = null
    for (let i = 0; i < formState.value.approvalLevels.length; i++) {
      const level = formState.value.approvalLevels[i]
      if (!level.approvers || level.approvers.length === 0) {
        firstInvalidLevel = level.title
        break
      }
    }

    if (firstInvalidLevel) {
      messageService.warning(`${firstInvalidLevel}未配置`)
    }
  })
}

function handleApproverChange(index, _value, selectedStaffs) {
  const list = Array.isArray(selectedStaffs) ? selectedStaffs : []
  formState.value.approvalLevels[index].approverNames = list.map(item => item?.nickName || item?.name).filter(Boolean)
}
</script>

<template>
  <a-modal
    v-model:open="openModal" centered class="modal-content-box" :keyboard="false" :closable="false"
      :mask-closable="false" :width="560"
  >
    <template #title>
      <div class="text-5 flex justify-between flex-center">
        <span>{{ title }}</span>
        <a-button type="text" class="close-btn" @click="closeFun">
          <template #icon>
            <CloseOutlined class="text-5 close-icon" />
          </template>
        </a-button>
      </div>
    </template>
    <a-form ref="formRef" :model="formState">
      <div class="contenter scrollbar">
        <div class="stepList">
          <div class="stepBody">
            <div class="stepTitle">
              申请人触发审批
            </div>
            <!-- 循环主体 start -->
            <div v-for="(level, index) in formState.approvalLevels" :key="index">
              <div class="stepItem">
                <div class="itemTitle">
                  {{ level.title }}
                  <a-button v-if="index > 0" type="link" @click="removeLevel(index)">
                    删除
                  </a-button>
                </div>
                <div class="itemDropBox">
                  <a-form-item
                    :name="['approvalLevels', index, 'approvers']"
                    :rules="[{ required: true, message: '请选择审批人' }]"
                  >
                    <!-- 多选 -->
                    <StaffSelect
                      v-model="level.approvers"
                      :multiple="true"
                      placeholder="请选择/输入审批人姓名（多选）"
                      width="454px"
                      fetch-type="approval"
                      @change="(...args) => handleApproverChange(index, ...args)"
                    />
                  </a-form-item>
                </div>
              </div>
            </div>
            <!-- 循环主体 end -->
            <div>
              <div class="stepItemAdd" @click="addLevel">
                <img
                  class="addIcon" src="https://pcsys.admin.ybc365.com/b2fbdf31-a670-4a61-a2cb-cc6a01034f92.png"
                  alt=""
                >
                <div class="addText">
                  添加审批层级（{{ formState.approvalLevels.length }}/10）
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="bottomTips">
          <div class="tipsIcon">
            小贴士
          </div>
          <div class="tipsText">
            <p>1. 审批人均为【或批】，即相同审批层级设置多个审批人后，只要任一审批人通过，即可完成此层级审批。</p>
            <p>2. 若申请人与审批人一致，该层级审批将会自动通过。</p>
            <p>3. 若为多级审批，且本级审批人在此审批单之前操作过通过的，本级审批会自动通过。</p>
          </div>
        </div>
      </div>
    </a-form>
    <template #footer>
      <a-button danger ghost @click="closeFun">
        取消
      </a-button>
      <a-button type="primary" ghost @click="handleSubmit">
        确定
      </a-button>
    </template>
  </a-modal>
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

.contenter {
  padding: 24px;
  max-height: calc(100vh - 220px);
  overflow-y: auto;

  .stepList {
    .stepBody {
      position: relative;
      padding-left: 26px;

      &::after {
        position: absolute;
        content: "";
        top: 16px;
        left: 7.5px;
        bottom: 18px;
        width: 1px;
        background: #d0e4ff;
      }

      .stepTitle {
        position: relative;
        font-weight: 500;
        font-size: 16px;
        color: #333;

        &::after {
          position: absolute;
          content: "";
          top: 3px;
          left: -26px;
          width: 16px;
          height: 17px;
          background: url("https://pcsys.admin.ybc365.com/84195e8a-c21c-4a3d-b2d6-c08dd3505a6f.png") no-repeat;
          background-size: contain;
        }
      }

      .stepItem {
        width: 100%;
        margin-top: 16px;
        background: #fafafa;
        border-radius: 8px;

        .itemTitle {
          position: relative;
          display: flex;
          align-items: center;
          justify-content: space-between;
          height: 54px;
          font-weight: 500;
          font-size: 16px;
          color: #333;
          padding: 0 16px;
          border-bottom: 1px solid #eee;

          &::after {
            position: absolute;
            z-index: 2;
            content: "";
            top: 19px;
            left: -24px;
            width: 12px;
            height: 13px;
            background: url("https://pcsys.admin.ybc365.com/d7fc84ec-a103-4e8a-a737-3ce23dd7e472.png") no-repeat;
            background-size: contain;
          }
        }

        .itemDropBox {
          padding: 16px;
          :deep(.ant-form-item) {
            margin-bottom: 0;
          }
        }
      }

      .stepItemAdd {
        position: relative;
        display: inline-flex;
        align-items: center;
        margin-top: 16px;
        cursor: pointer;

        .addIcon {
          width: 16px;
          height: 16px;
          margin-right: 8px;
          cursor: pointer;
          object-fit: contain;
        }

        .addText {
          color: #06f;

          &::after {
            position: absolute;
            content: "";
            top: 2px;
            left: -26px;
            width: 16px;
            height: 17px;
            background: url("https://pcsys.admin.ybc365.com/84195e8a-c21c-4a3d-b2d6-c08dd3505a6f.png") no-repeat;
            background-size: contain;
          }
        }
      }
    }
  }

  .bottomTips {
    position: relative;
    display: flex;
    align-items: flex-start;
    margin-top: 32px;
    background: #f5f9ff;
    border-radius: 8px 8px 0 8px;
    padding: 16px 0;

    &::before {
      content: "";
      position: absolute;
      right: -7px;
      bottom: -2px;
      width: 0;
      height: 0;
      border-left: 10px solid transparent;
      border-right: 10px solid transparent;
      border-top: 10px solid #fff;
      transform: rotate(-45deg);
    }

    &::after {
      content: "";
      position: absolute;
      right: 0;
      bottom: 6px;
      width: 0;
      height: 0;
      border-left: 10px solid transparent;
      border-right: 10px solid transparent;
      border-top: 10px solid #b3d1ff;
      transform: rotate(135deg);
    }

    .tipsIcon {
      display: flex;
      justify-content: center;
      width: 68px;
      padding-top: 6px;
      font-weight: 600;
      font-size: 12px;
      color: #8aafe9;

    }

    .tipsText {
      flex: 1 1;
      border-left: 1px dashed rgba(179, 209, 255, .6);
      padding: 0 16px;

      p {
        font-weight: 400;
        font-size: 12px;
        color: #99999a;
        margin: 0;
        line-height: 1.8;
      }
    }
  }
}

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
</style>

<style>
.modal-content-box .ant-modal-header {
  padding: 10px 16px !important;
  margin-bottom: 0;
}

.modal-content-box .ant-modal-body {
  padding: 0 !important;
}
</style>
