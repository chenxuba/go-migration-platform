<script setup>
import { LeftOutlined } from '@ant-design/icons-vue'
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  buildAmountOrderImportTemplateApi,
  buildLessonHourOrderImportTemplateApi,
  buildTimeSlotOrderImportTemplateApi,
  submitOrderImportTaskApi,
  uploadOrderImportApi,
} from '~@/api/finance-center/order-import'
import { useUserStore } from '~@/stores/user'
import messageService from '~@/utils/messageService'

const router = useRouter()
const userStore = useUserStore()

const schoolName = computed(() => userStore.userInfo?.orgName || '总机构')
const fileList = ref([])
const downloadLoading = ref(false)
const uploadLoading = ref(false)
const handleChange = () => {}

const templateLinks = [
  '下载按课时订单模板',
  '下载按时段订单模板',
  '下载按金额订单模板',
  '下载学杂费订单模板',
  '下载教材订单模板',
]

function replaceToWorkbench() {
  window.location.replace(`${window.location.origin}${window.location.pathname}#/finance-center/order-list`)
}

function handleDrop(e) {
  console.log(e)
}

function goBack() {
  replaceToWorkbench()
}

function handleImportRecord() {
  router.push('/import-center/import-order/record')
}

async function handleDownloadTemplate(label) {
  downloadLoading.value = true
  try {
    let res
    switch (label) {
      case '下载按课时订单模板':
        res = await buildLessonHourOrderImportTemplateApi()
        break
      case '下载按时段订单模板':
        res = await buildTimeSlotOrderImportTemplateApi()
        break
      case '下载按金额订单模板':
        res = await buildAmountOrderImportTemplateApi()
        break
      default:
        messageService.info(`${label}待接入`)
        return
    }
    const url = res.result || res.data
    if (!url) {
      throw new Error('模板下载链接生成失败')
    }
    window.open(url, '_blank')
  }
  catch (error) {
    console.error('download order template failed', error)
    messageService.error(error?.message || '模板下载失败，请稍后重试')
  }
  finally {
    downloadLoading.value = false
  }
}

async function handleCustomUpload({ file, onSuccess, onError }) {
  if (uploadLoading.value) {
    return
  }
  uploadLoading.value = true
  try {
    if (!file) {
      throw new Error('请选择文件')
    }
    const formData = new FormData()
    formData.append('file', file)
    const uploadRes = await uploadOrderImportApi(formData)
    const uploadResult = uploadRes.result || uploadRes.data
    if (!uploadResult?.fileUrl) {
      throw new Error(uploadRes.message || '导入文件上传失败')
    }

    const submitRes = await submitOrderImportTaskApi({
      fileUrl: uploadResult.fileUrl,
      fileName: uploadResult.fileName || file.name,
    })
    const taskId = submitRes.result || submitRes.data
    if (!taskId) {
      throw new Error(submitRes.message || '导入任务创建失败')
    }

    fileList.value = [file]
    onSuccess?.({ taskId })
    router.push(`/import-center/import-order/edit/${taskId}`)
  }
  catch (error) {
    console.error('parse order import failed', error)
    messageService.error(error?.message || '导入文件解析失败')
    onError?.(error)
  }
  finally {
    uploadLoading.value = false
  }
}
</script>

<template>
  <div class="import-center-layout w-full h-full bg-#f7f7fd">
    <div class="work-top flex justify-between items-center h-56px bg-#fff">
      <div class="work-top-left flex items-center">
        <div class="import-header-logo" title="导入中心" aria-hidden="true" />
        <span class="text-20px text-#06f font500 ml-16px flex items-center cursor-pointer" @click="goBack">
          <LeftOutlined class="mt--1px" /> 返回
        </span>
      </div>
      <div class="work-top-right pr-20px text-16px text-#000 font500">
        当前机构：{{ schoolName }}
      </div>
    </div>
    <div class="work-main">
      <div class="work-main-card">
        <div class="flex justify-between items-center">
          <span class="text-24px text-#000 font500">导入学员订单</span>
          <a-button @click="handleImportRecord">导入记录</a-button>
        </div>
        <div class=" text-20px text-#666 font500">
          没做好准备？<span class="text-#06f cursor-pointer" @click="goBack">返回工作台</span>
        </div>
        <div class="flex justify-between mt16px">
          <div class="bg-#0064ff0a flex-1 rounded-6px px-24px py-12px flex  flex-col justify-center">
            <div class="text-16px text-#222 font500">✅ 正确做法：</div>
            <div class="text-16px text-#222">请务必先创建「<span class="font500">即将导入的学员</span>」在「<span class="font500">当前机构内</span>」的课程、学杂费或教材商品</div>
            <div class="text-16px text-#222">如需「<span class="font500">分班</span>」，则需要先创建班级，再下载「<span class="font500">学员订单模板</span>」，按模板格式要求填写内容</div>
          </div>
          <div class="bg-#ff32320a flex-1 h-136px rounded-6px ml-16px px-24px flex flex-col justify-center">
            <div class="text-16px text-#222 font500">❌ 错误做法：</div>
            <div class="text-16px text-#222">未完成导入学员的课程和班级创建</div>
            <div class="text-16px text-#222">未按要求填写模板的表格内容</div>
            <div class="text-16px text-#222">重复上传相同表格，或上传空表格</div>
          </div>
        </div>
        <div class="upload-box mt20px h-350px">
          <a-upload-dragger
            v-model:fileList="fileList"
            name="file"
            :multiple="false"
            class=" w-full h-full"
            :custom-request="handleCustomUpload"
            :show-upload-list="false"
            :disabled="uploadLoading"
            @change="handleChange"
            @drop="handleDrop"
          >
            <div class="flex flex-col items-center justify-center w-full h-full">
              <div class="ant-upload-drag-icon flex flex-col items-center">
                <img src="https://pcsys.admin.ybc365.com/e8183085-4314-4fdf-a9b1-f1934defad7c.png" class="h-50px" alt="">
                <a-button type="primary" class="mt-16px rounded-10 w-120px h-30px text-12px font500" :loading="uploadLoading">
                  本地上传
                </a-button>
                <div class="mt-18px text-16px text-#888 font500">当前仅支持上传扩展名为 .xls .xlsx 的文件（每次最多支持导入1000条数据）</div>
                <div class="text-16px text-#888 font500">请务必按照模板内容填写学员数据，否则可能会无法正常导入</div>
                <div class="mt-12px flex flex-wrap items-center justify-center gap-x-14px gap-y-6px">
                  <a-button
                    v-for="label in templateLinks"
                    :key="label"
                    type="link"
                    class="text-16px text-#06f font-500 px-0!"
                    :loading="downloadLoading"
                    @click.stop="handleDownloadTemplate(label)"
                  >
                    {{ label }}
                  </a-button>
                </div>
              </div>
            </div>
          </a-upload-dragger>
        </div>
      </div>
    </div>
    <div class="work-footer">
      <div class="beian-info flex items-center text-12px">
        <a href="https://beian.mps.gov.cn/#/query/webSearch" target="_blank" class="beian-link flex items-center ">
          <img src="https://pcsys.admin.ybc365.com/a0b0315a-d432-46fa-8d86-2d466b650271.png" alt="公安备案" class="beian-icon">
          <span class="beian-text text-14px">沪ICP备15044463号-1 </span>
        </a>
        <span class="ml-6px">已通过 ISO27001:2013 信息安全认证</span>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
/* 左上角品牌徽标：纯 CSS，无外链图 */
.import-header-logo {
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.22) 0%, transparent 42%),
    linear-gradient(145deg, #2b8cff 0%, #0066ff 45%, #0050d8 100%);
  box-shadow:
    0 1px 2px rgba(0, 80, 200, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
  position: relative;
  overflow: hidden;

  /* 表格式横线（白底小表 + 行线，示意导入表格） */
  &::before {
    content: '';
    position: absolute;
    left: 12px;
    top: 15px;
    width: 32px;
    height: 26px;
    background-color: rgba(255, 255, 255, 0.94);
    background-image:
      linear-gradient(rgba(0, 102, 255, 0.22), rgba(0, 102, 255, 0.22)),
      linear-gradient(rgba(0, 102, 255, 0.18), rgba(0, 102, 255, 0.18)),
      linear-gradient(rgba(0, 102, 255, 0.14), rgba(0, 102, 255, 0.14));
    background-size: 24px 2px, 18px 2px, 22px 2px;
    background-position: 4px 8px, 4px 14px, 4px 20px;
    background-repeat: no-repeat;
    box-shadow: 0 1px 2px rgba(0, 60, 180, 0.15);
  }

  
}

.work-main {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 1366px;
  min-height: calc(100% - 110px);
  padding: 40px 0px;
  background-color: #f7f7fd;

  .work-main-card {
    position: relative;
    box-sizing: border-box;
    width: 1300px;
    padding: 50px 80px 48px;
    border-radius: 24px;
    background: rgb(255, 255, 255);
    box-shadow: rgba(0, 0, 0, 0.08) 0px 0px 32px 0px;
  }
}

.work-footer {
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f7f7fd;

  .beian-info {
    .beian-link {
      .beian-icon {
        width: 16px;
        height: 16px;
        margin-right: 6px;
      }

    }
  }
}
</style>
