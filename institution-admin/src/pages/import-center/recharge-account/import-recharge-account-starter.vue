<script setup>
import { LeftOutlined, QuestionCircleOutlined } from '@ant-design/icons-vue'
import { computed, ref } from 'vue'
import {
  buildRechargeAccountImportByAccountTemplateApi,
  buildRechargeAccountImportByStudentTemplateApi,
  submitRechargeAccountImportTaskApi,
  uploadRechargeAccountImportApi,
} from '@/api/finance-center/recharge-account'
import { useUserStore } from '~@/stores/user'
import messageService from '~@/utils/messageService'
import { useRouter } from 'vue-router'

const router = useRouter()
const userStore = useUserStore()

const schoolName = computed(() => userStore.userInfo?.orgName || '总机构')
const fileList = ref([])
const downloadLoading = ref(false)
const uploadLoading = ref(false)

function replaceToWorkbench() {
  window.location.replace(`${window.location.origin}${window.location.pathname}#/finance-center/recharge-account`)
}

function goBack() {
  replaceToWorkbench()
}

function handleImportRecord() {
  router.push('/import-center/import-recharge-account/record')
}

async function handleDownloadTemplate(type) {
  downloadLoading.value = true
  try {
    const res = type === 'byAccount'
      ? await buildRechargeAccountImportByAccountTemplateApi()
      : await buildRechargeAccountImportByStudentTemplateApi()
    const url = res.result || res.data
    if (!url) {
      throw new Error('模板下载链接生成失败')
    }
    window.open(url, '_blank')
  }
  catch (error) {
    console.error('download recharge account import template failed', error)
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
    const uploadRes = await uploadRechargeAccountImportApi(formData)
    const uploadResult = uploadRes.result || uploadRes.data
    if (!uploadResult?.fileUrl) {
      throw new Error(uploadRes.message || '导入文件上传失败')
    }
    const submitRes = await submitRechargeAccountImportTaskApi({
      fileUrl: uploadResult.fileUrl,
      fileName: uploadResult.fileName || file.name,
    })
    const taskId = submitRes.result || submitRes.data
    if (!taskId) {
      throw new Error(submitRes.message || '导入任务创建失败')
    }

    fileList.value = [file]
    onSuccess?.({ taskId })
    router.push(`/import-center/import-recharge-account/edit/${taskId}`)
  }
  catch (error) {
    console.error('parse recharge account import failed', error)
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
        当前校区：{{ schoolName }}
      </div>
    </div>

    <div class="work-main">
      <div class="work-main-card">
        <div class="flex justify-between items-center">
          <div class="flex items-center">
            <span class="text-24px text-#000 font500">导入储值账户</span>
            <QuestionCircleOutlined class="ml-8px text-#06f text-18px" />
          </div>
          <a-button @click="handleImportRecord">导入记录</a-button>
        </div>

        <div class="text-20px text-#666 font500">
          没做好准备？<span class="text-#06f cursor-pointer" @click="goBack">返回工作台</span>
        </div>

        <div class="flex justify-between mt16px">
          <div class="bg-#0064ff0a flex-1 rounded-6px px-24px py-12px flex flex-col justify-center">
            <div class="text-16px text-#222 font500">✅ 正确做法：</div>
            <div class="text-16px text-#222">
              <span class="text-#ff4d4f font500">【重点】</span>按关联学员导入时，请确保所「<span class="font500">关联的学员</span>」
            </div>
            <div class="text-16px text-#222">已存在。如学员不存在，系统将自动为您「创建新学员」。</div>
            <div class="text-16px text-#222">按储值账户导入时，请确保「<span class="font500">储值账户号</span>」已存在。</div>
          </div>
          <div class="bg-#ff32320a flex-1 rounded-6px ml-16px px-24px py-12px flex flex-col justify-center">
            <div class="text-16px text-#222 font500">❌ 错误做法：</div>
            <div class="text-16px text-#222">未按要求填写模板的表格内容</div>
            <div class="text-16px text-#222">重复上传相同表格，或上传空表格</div>
          </div>
        </div>

        <div class="upload-box mt20px h-350px">
          <a-upload-dragger
            v-model:fileList="fileList"
            name="file"
            :multiple="false"
            class="w-full h-full"
            :custom-request="handleCustomUpload"
            :show-upload-list="false"
            :disabled="uploadLoading"
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
                    type="link"
                    class="text-16px text-#06f font-500 px-0!"
                    :loading="downloadLoading"
                    @click.stop="handleDownloadTemplate('byStudent')"
                  >
                    下载按关联学员导入模板
                  </a-button>
                  <a-button
                    type="link"
                    class="text-16px text-#06f font-500 px-0!"
                    :loading="downloadLoading"
                    @click.stop="handleDownloadTemplate('byAccount')"
                  >
                    下载按储值账户导入模板
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
        <a href="https://beian.mps.gov.cn/#/query/webSearch" target="_blank" class="beian-link flex items-center">
          <img src="https://pcsys.admin.ybc365.com/a0b0315a-d432-46fa-8d86-2d466b650271.png" alt="公安备案" class="beian-icon">
          <span class="beian-text text-14px">沪ICP备15044463号-1 </span>
        </a>
        <span class="ml-6px">已通过 ISO27001:2013 信息安全认证</span>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
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

  &::after {
    content: '';
    position: absolute;
    left: 18px;
    top: 21px;
    width: 20px;
    height: 2px;
    background: rgba(0, 102, 255, 0.18);
    transform: rotate(-18deg);
    transform-origin: left center;
    opacity: 0.5;
  }
}

.import-center-layout {
  min-height: 100%;
}

.work-main {
  padding: 40px 0 32px;
  display: flex;
  justify-content: center;
}

.work-main-card {
  width: 1300px;
  background: #fff;
  border-radius: 24px;
  box-shadow: 0 0 32px rgba(0, 0, 0, 0.05);
  padding: 28px 32px 32px;
}

.upload-box {
  border: 1px dashed #e7e7f2;
  border-radius: 12px;
  overflow: hidden;
}

.work-footer {
  display: flex;
  justify-content: center;
  padding: 12px 0 18px;
  color: #999;
}

.beian-link {
  color: inherit;
  text-decoration: none;
}

.beian-icon {
  width: 14px;
  height: 14px;
  margin-right: 4px;
}
</style>
