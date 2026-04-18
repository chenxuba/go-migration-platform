<script setup>
import { QuestionCircleOutlined } from '@ant-design/icons-vue'

const emit = defineEmits(['confirm'])

const open = defineModel({
  default: false,
})

function confirm() {
  emit('confirm', true)
  open.value = false
}
function cancel() {
  emit('confirm', false)
  open.value = false
}
const value = ref('')
</script>

<template>
  <a-modal
    v-model:open="open" centered wrap-class-name="modal" width="800px" :body-style="{ padding: 0 }"
    title="请确认以下信息" cancel-text="返回修改" ok-text="确认转出" :keyboard="false" :mask-closable="false" @ok="confirm"
    @cancel="cancel"
  >
    <div class="modal-wrap p-25px">
      <!-- 转出课程信息 -->
      <div class="position-relative mb-20px p20px bg-#fafafa rounded-15px">
        <div
          style="border-radius:0 15px 0 15px ;"
          class="text-white text-12px bg-#666666 px-15px py-2px position-absolute top-0px right-0px"
        >
          转出
        </div>
        <div class="text-lg font-medium mb-3">
          认知课
        </div>
        <div class="flex gap-2 mb-3">
          <span class="bg-#e6f0ff text-#06f px-2  rounded-lg text-12px">班级课程</span>
          <span class="bg-#e6f0ff text-#06f px-2  rounded-lg text-12px">课时</span>
        </div>
        <div>
          <div class="mb-10px">
            <span class="text-#222222">课时：</span>
            <span class="text-#969696">1 课时</span>
          </div>
          <div>
            <span class="text-#222222">有效期至：</span>
            <span class="text-#969696">-</span>
          </div>
        </div>
        <div class="font-medium mt-10px">
          总计学费：¥ 233.31
        </div>
      </div>

      <!-- 转入课程信息 -->
      <div class="position-relative border-solid border-2px rounded-15px p-20px border-#06f">
        <div
          style="border-radius:0 15px 0 15px ;"
          class="text-white text-12px bg-#06f px-15px py-2px position-absolute top-0px right--2px"
        >
          转入
        </div>
        <div class="flex items-center justify-between mb-3">
          <div class="text-lg font-medium">
            视知觉训练
          </div>
        </div>
        <div class="text-gray-600">
          <div class="mb-2 text-#222222">
            <span>
              天数：
            </span>
            <span>
              31天
            </span>
          </div>
          <div class="mb-2 text-#222222">
            <span>
              开始时间：
            </span>
            <span>
              2025-05-15 ~ 2025-05-16
            </span>
          </div>
          <div class="mb-2 text-#222222">
            <span>
              转入班级：
            </span>
            <span>
              -
            </span>
          </div>
          <div class="mb-2 text-#222222">
            <span>
              赠送天数：
            </span>
            <span>
              0 天
            </span>
          </div>
          <div class="mb-2 text-#222222">
            <span>
              单课优惠：
            </span>
            <span>
              -￥0.00
            </span>
          </div>
          <div class="mb-2 text-#222222">
            <span>
              优惠卷：
            </span>
            <span>
              无
            </span>
          </div>
          <div class="mb-2 text-#222222">
            <span>
              总计优惠：
            </span>
            <span>
              -￥0.00
            </span>
          </div>
        </div>
        <div class="font-medium">
          总计学费：¥ 233.31
        </div>
      </div>
      <!-- 使用储值账户 -->
      <div class="flex items-center justify-between mt-20px">
        <div class="flex items-center rounded-15px p-20px bg-#f6f7f8">
          <div>使用储值账户：</div>
          <div class="mr-20px">
            <div class="flex items-center">
              <span>本次使用充值余额：</span>
              <a-input-search v-model:value="value" style="width: 100px;" placeholder="">
                <template #enterButton>
                  <a-button>元</a-button>
                </template>
              </a-input-search>
            </div>
            <span class="text-#a7a8a8 text-13px">可用充值余额：¥ 0.00</span>
          </div>
          <div>
            <div class="flex items-center">
              <span>使用赠送余额：</span>
              <a-input-search v-model:value="value" style="width: 100px;" placeholder="">
                <template #enterButton>
                  <a-button>元</a-button>
                </template>
              </a-input-search>
            </div>
            <span class="text-#a7a8a8 text-13px">可用充值余额：¥ 0.00</span>
          </div>
        </div>
        <div class="flex flex-col items-end justify-between">
          <div class="flex items-center gap-5px">
            <span class="text-#797979">应收金额</span>
            <a-popover color="#fff" title="字段说明">
              <template #content>
                应收金额 = 转入课程金额 - 单课优惠 - 优惠券 - 转出课程金额
              </template>
              <QuestionCircleOutlined class="text-#888 font-size-14px" />
            </a-popover>
          </div>
          <div class="text-26px font-700">
            ￥2570.00
          </div>
        </div>
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less"></style>
