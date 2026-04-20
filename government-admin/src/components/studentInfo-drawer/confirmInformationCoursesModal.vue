<script setup>
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
</script>

<template>
  <a-modal
    v-model:open="open" centered wrap-class-name="modal" width="1000px" :body-style="{ padding: 0 }"
    title="请确认以下信息" cancel-text="返回修改" ok-text="确认转出" :keyboard="false" :mask-closable="false"
    @ok="confirm" @cancel="cancel"
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
              全额：
            </span>
            <span>
              233.31 元
            </span>
          </div>
          <div class="mb-2 text-#222222">
            <span>
              有效期至：
            </span>
            <span>
              无限制
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
              赠送金额：
            </span>
            <span>
              0 元
            </span>
          </div>
        </div>
        <div class="font-medium">
          总计学费：¥ 233.31
        </div>
      </div>
    </div>
  </a-modal>
</template>

<style scoped lang="less"></style>
