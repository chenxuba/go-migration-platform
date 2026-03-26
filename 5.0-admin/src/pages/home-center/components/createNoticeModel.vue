<script setup>
import { QuillEditor } from '@vueup/vue-quill'
import '@vueup/vue-quill/dist/vue-quill.snow.css'
import { CloseOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'

const open = defineModel({
  type: Boolean,
  default: false,
})

const openDrawer = ref(false)

const formState = reactive({
  title: '',
  content: '',
  rule: 1,
  students: [],
})

const settingsForm = reactive({
  rule: '1',
})

// 定义自定义工具栏的引用
const toolbarRef = ref(null)

const editorOption = {
  modules: {
    toolbar: {
      container: '#toolbar', // 指向自定义工具栏的ID
    },
  },
  theme: 'snow',
}

function handleOk() {
  open.value = false
}
</script>

<template>
  <div>
    <a-modal
      v-model:open="open" :after-close="() => openDrawer = true" :keyboard="false" :mask-closable="false"
      class="noticeModel" width="800px" title="通知公告规则" destroy-on-close
    >
      <div class="rulesWrapper">
        <div class="rulesTitle">
          一、内容规范
        </div>
        <div class="rulesContent">
          用户使用微信公众平台通知服务，须遵守平台相关运营规范。为避免发送的内容引起学员家长投诉导致微信封禁，请勿发布以下违规内容。<br>1.
          发送内容与服务场景不一致（含标题、关键词）的模板消息。<br>2. 在文字或图片中含有广告营销类内容。<br>如：报课优惠类通知、报课返利类通知、课程降价类通知等一些涉及消费的营销类通知。<br>3.
          发送红包、卡券、优惠券、代金券、会员卡类。<br>如：报课领红包、参加活动领优惠券、预存金额送代金券等。<br>4.
          频繁发送相同内容或性质的消息，对用户造成骚扰。原则上，仅支持一个自然日对同一用户发送一次消息。<br>如：频率过高的到期提醒类通知、频率过高的缴费提醒类通知、频率过高的留言提醒类通知、订阅提醒类通知等。<br>处罚规则：<br>一经发现将根据违规程度采取阶梯性封禁通知公告功能等措施。<br>更多运营规范内容可参考：<a
            target="_blank" href="https://mp.weixin.qq.com/mp/opshowpage?action=newoplaw#t3-3-9"
            rel="noreferrer"
          >《微信公众平台运营规范》</a>
        </div><br>
        <div>
          <div class="rulesTitle">
            二、审核注意事项
          </div>
          <div class="rulesContent">
            为避免过度打扰，仅支持一个自然日内对同一学员发送一条通知公告。如果某学员已经被发送过通知公告，则当日发送给该学员的其他通知公告将发送失败。<br>通知公告内容审核通过则立即发送，如有疑问，可咨询校宝客户经理。
          </div>
        </div>
      </div>
      <template #footer>
        <a-button key="submit" type="primary" @click="handleOk">
          知道了
        </a-button>
      </template>
    </a-modal>
    <a-drawer
      v-model:open="openDrawer" width="800px" :keyboard="false" :mask-closable="false" placement="right"
      :body-style="{ padding: '0', background: '#f7f7fd' }" :closable="false"
    >
      <!-- 自定义头部 -->
      <template #title>
        <div class="custom-header flex justify-between h-4 flex-items-center">
          <div class="text-5">
            创建通知
          </div>
          <a-button type="text" class="close-btn" @click="openDrawer = false">
            <template #icon>
              <CloseOutlined class="text-5 close-icon" />
            </template>
          </a-button>
        </div>
      </template>
      <!-- 自定义内容 -->
      <div class="p-20px">
        <!-- 内容 -->
        <div class="px-15px py-20px bg-white rounded-12px">
          <custom-title title="内容" font-size="20px" font-weight="550">
            <template #left>
              <div class="flex justify-between items-center gap-5px">
                <span>内容</span>
                <div
                  class="text-14px text-#999 flex items-center gap-5px cursor-pointer hover:text-#06f"
                  @click="open = true"
                >
                  <QuestionCircleOutlined />
                  <span>规则</span>
                </div>
              </div>
            </template>
            <template #right>
              <a-button type="link" size="small" class="text-12px" @click="handleCancel">
                预览
              </a-button>
            </template>
          </custom-title>
          <a-form
            class="mt-20px" :model="formState" name="basic" :label-col="{ span: 3 }" :wrapper-col="{ span: 21 }"
            autocomplete="off"
          >
            <a-form-item label="通知标题" name="title" :rules="[{ required: true, message: '请输入通知标题' }]">
              <a-input v-model:value="formState.title" class="w-240px" placeholder="请输入(最多20字)" />
            </a-form-item>
            <a-form-item label="通知内容" name="content" :required="true">
              <!-- 自定义工具栏 -->
              <div id="toolbar" ref="toolbarRef">
                <!-- 标题 -->
                <select class="ql-header">
                  <option value="2">
                    标题2
                  </option>
                  <option value="3">
                    标题3
                  </option>
                  <option value="false" selected>
                    正文
                  </option>
                </select>
                <!-- 加粗、斜体、下划线、删除线 -->
                <button class="ql-bold" />
                <button class="ql-italic" />
                <button class="ql-underline" />
                <button class="ql-strike" />
                <!-- 字体颜色、背景颜色 -->
                <select class="ql-color" />
                <select class="ql-background" />
                <!-- 字体大小 -->
                <select class="ql-size">
                  <option value="small">
                    字号10
                  </option>
                  <option selected>
                    默认字号
                  </option>
                  <option value="large">
                    字号18
                  </option>
                  <option value="huge">
                    字号32
                  </option>
                </select>

                <!-- 图片 -->
                <button class="ql-image" />
              </div>
              <QuillEditor
                v-model:content="formState.content" placeholder="在编辑通知公告时，请注意：
1.直接从外部复制的图片可能无法在微信小程序中正常显示，请尽量使用编辑器的图片上传功能来上传您的图片。
2.文本样式可能需要调整以保证在微信小程序中的显示效果。
发布前，建议预览内容以确保一切显示正常。" style="height: 500px;" :options="editorOption"
              />
            </a-form-item>
          </a-form>
        </div>
        <!-- 发布设置 -->
        <div class="px-15px py-20px bg-white rounded-12px mt-15px">
          <custom-title title="发布设置" font-size="20px" font-weight="550" />
          <a-form
            class="mt-20px" :model="settingsForm" name="basic" :label-col="{ span: 4 }" :wrapper-col="{ span: 20 }"
            autocomplete="off"
          >
            <a-form-item name="rule" :rules="[{ required: true, message: '请选择通知方式' }]">
              <template #label>
                <span class="mr-2px">通知标题</span>
                <a-tooltip title="未关注家校平台的学员家长，无法接收通知公告">
                  <QuestionCircleOutlined />
                </a-tooltip>
              </template>
              <a-radio-group v-model:value="settingsForm.rule" class="custom-radio">
                <a-radio value="1">
                  全校群发 (推送1人/共4人)
                </a-radio>
                <a-radio value="2">
                  选择班级/1v1
                </a-radio>
              </a-radio-group>
            </a-form-item>
            <a-form-item label="发布时间" name="time" :rules="[{ required: true, message: '请选择发布时间' }]">
              <a-radio-group v-model:value="settingsForm.time" class="custom-radio">
                <a-radio value="1">
                  立即发布
                </a-radio>
                <a-radio value="2">
                  定时发布
                </a-radio>
              </a-radio-group>
            </a-form-item>
            <a-form-item label="需家长确认" name="students">
              <div class="flex items-center">
                <a-switch v-model:checked="checked" />
                <span class="ml-5px text-13px text-gray">开启后，家长需点击按钮来确认收到通知</span>
              </div>
            </a-form-item>
          </a-form>
        </div>
      </div>
      <!-- 自定义底部 -->
      <template #footer>
        <div class="flex justify-end">
          <a-button style="font-size: 20px; height: 48px;min-width: 140px;" size="large" type="primary" @click="handleSubmit">
            发布
          </a-button>
        </div>
      </template>
    </a-drawer>
  </div>
</template>

<style>
.noticeModel {
  /* display: inline-block; */
  padding-bottom: 0;
  text-align: left;
  top: 35px !important;
  vertical-align: middle;
}
</style>

<style scoped lang="less">
.close-btn {
  &:hover {
    background: transparent;
  }
}

.rulesWrapper {
  width: 100%;
  min-height: 430px;

  .rulesTitle {
    font-weight: 500;
    font-size: 14px;
    color: #222;
    line-height: 22px;
    margin-bottom: 8px
  }

  .rulesContent {
    font-weight: 400;
    font-size: 14px;
    color: #666;
    line-height: 22px
  }
}

::v-deep(.ql-toolbar) {
  border-radius: 8px 8px 0 0;
}

::v-deep(.ql-container) {
  border-radius: 0 0 8px 8px;
}
</style>
