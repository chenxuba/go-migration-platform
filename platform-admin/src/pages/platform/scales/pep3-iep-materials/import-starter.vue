<script setup lang="ts">
import { CheckCircleFilled, CloseCircleFilled, InboxOutlined, LeftOutlined } from '@ant-design/icons-vue'
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  buildPlatformPEP3IEPMaterialImportTemplateApi,
  downloadPlatformPEP3IEPMaterialImportTemplateFileApi,
  submitPlatformPEP3IEPMaterialImportTaskApi,
  uploadPlatformPEP3IEPMaterialImportApi,
} from '@/api/platform/scales'
import messageService from '@/utils/messageService'

const router = useRouter()
const fileList = ref<any[]>([])
const downloadLoading = ref(false)
const uploadLoading = ref(false)
const lastUploadSignature = ref('')
const lastUploadAt = ref(0)

function unwrap<T>(res: any): T {
  return (res?.result ?? res?.data ?? res) as T
}

function extractFileName(disposition?: string) {
  const header = `${disposition || ''}`
  const encodedMatch = header.match(/filename\*=UTF-8''([^;]+)/i)
  const plainMatch = header.match(/filename="?([^";]+)"?/i)
  const raw = encodedMatch?.[1] || plainMatch?.[1] || ''
  try {
    return decodeURIComponent(raw)
  } catch {
    return raw
  }
}

function goBack() {
  router.replace('/platform/scales/pep3-iep-materials')
}

async function handleDownloadTemplate() {
  downloadLoading.value = true
  try {
    const url = unwrap<string>(await buildPlatformPEP3IEPMaterialImportTemplateApi())
    if (!url) {
      messageService.error('模板下载链接生成失败')
      return
    }
    const response = await downloadPlatformPEP3IEPMaterialImportTemplateFileApi(url)
    const blob = new Blob([response.data], { type: response.headers?.['content-type'] || 'application/octet-stream' })
    const downloadUrl = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = downloadUrl
    link.download = extractFileName(response.headers?.['content-disposition']) || 'PEP3-IEP素材库导入模板.xlsx'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(downloadUrl)
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '模板下载失败，请稍后重试')
  } finally {
    downloadLoading.value = false
  }
}

async function handleCustomUpload({ file, onSuccess, onError }: any) {
  const fileSignature = `${file?.name || ''}_${file?.size || 0}_${file?.lastModified || 0}`
  const now = Date.now()
  if (uploadLoading.value)
    return
  if (lastUploadSignature.value === fileSignature && now - lastUploadAt.value < 3000)
    return

  uploadLoading.value = true
  lastUploadSignature.value = fileSignature
  lastUploadAt.value = now
  try {
    const formData = new FormData()
    formData.append('file', file)
    const uploadResult = unwrap<{ fileUrl: string, fileName: string }>(await uploadPlatformPEP3IEPMaterialImportApi(formData))
    if (!uploadResult?.fileUrl)
      throw new Error('导入文件上传失败')

    const taskId = unwrap<string>(await submitPlatformPEP3IEPMaterialImportTaskApi({
      fileUrl: uploadResult.fileUrl,
      fileName: uploadResult.fileName || file.name,
    }))
    if (!taskId)
      throw new Error('导入任务创建失败')

    fileList.value = [file]
    onSuccess?.({ taskId })
    router.push(`/platform/scales/pep3-iep-materials/import/edit/${taskId}`)
  } catch (error: any) {
    messageService.error(error?.response?.data?.message || error?.message || '导入文件解析失败')
    onError?.(error)
  } finally {
    uploadLoading.value = false
  }
}
</script>

<template>
  <div class="pep3-import-layout">
    <div class="work-top">
      <div class="work-top-left">
        <div class="import-header-logo" title="导入中心" aria-hidden="true" />
        <span class="back-link" @click="goBack">
          <LeftOutlined /> 返回
        </span>
      </div>
      <div class="work-top-right">
        当前端：平台总控
      </div>
    </div>

    <div class="work-main">
      <div class="work-main-card">
        <div class="card-header">
          <div>
            <div class="page-title">
              导入PEP3 IEP素材库
            </div>
            <div class="page-subtitle">
              按模板维护题目选项、长期目标、短期目标和训练内容。
            </div>
          </div>
          <a-button @click="router.push('/platform/scales/pep3-iep-materials/import/record')">
            导入记录
          </a-button>
        </div>

        <div class="hint-row">
          <div class="hint-card hint-card--ok">
            <div class="hint-title">
              <CheckCircleFilled class="hint-icon hint-icon--ok" />
              正确做法
            </div>
            <div>先下载最新模板，领域、题目、选项、课程形式、状态都从下拉项选择。</div>
            <div>每行至少填写「领域、题目、选项、长期目标」，短期目标和训练内容可按需补充。</div>
          </div>
          <div class="hint-card hint-card--danger">
            <div class="hint-title">
              <CloseCircleFilled class="hint-icon hint-icon--danger" />
              错误做法
            </div>
            <div>不要修改表头名称，不要手输题号或把题目复制成非模板下拉文案。</div>
            <div>不要只填训练内容不填短期目标，训练项目和训练内容必须成对出现。</div>
          </div>
        </div>

        <div class="upload-box">
          <a-upload-dragger
            v-model:fileList="fileList"
            name="file"
            :multiple="false"
            :custom-request="handleCustomUpload"
            :show-upload-list="false"
            :disabled="uploadLoading"
          >
            <div class="upload-inner">
              <InboxOutlined class="upload-icon" />
              <a-button type="primary" class="upload-btn" :loading="uploadLoading">
                本地上传
              </a-button>
              <div class="upload-desc">
                当前仅支持上传 .xls .xlsx 文件，每次最多导入1000条数据
              </div>
              <div class="upload-desc">
                请按照模板下拉选项填写，否则会进入异常数据处理
              </div>
              <a-button type="link" class="template-link" :loading="downloadLoading" @click.stop="handleDownloadTemplate">
                下载PEP3素材导入模板
              </a-button>
            </div>
          </a-upload-dragger>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="less" scoped>
.pep3-import-layout {
  height: 100vh;
  min-height: 100vh;
  overflow: hidden;
  background: #f7f7fd;
}

.work-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 52px;
  background: #fff;
}

.work-top-left {
  display: flex;
  align-items: center;
}

.work-top-right {
  padding-right: 24px;
  color: #000;
  font-size: 15px;
  font-weight: 500;
}

.import-header-logo {
  position: relative;
  width: 52px;
  height: 52px;
  flex-shrink: 0;
  overflow: hidden;
  background: linear-gradient(145deg, #2b8cff 0%, #0066ff 45%, #0050d8 100%);
}

.import-header-logo::before {
  position: absolute;
  top: 14px;
  left: 11px;
  width: 30px;
  height: 24px;
  background-color: rgba(255, 255, 255, 0.94);
  background-image:
    linear-gradient(rgba(0, 102, 255, 0.22), rgba(0, 102, 255, 0.22)),
    linear-gradient(rgba(0, 102, 255, 0.18), rgba(0, 102, 255, 0.18)),
    linear-gradient(rgba(0, 102, 255, 0.14), rgba(0, 102, 255, 0.14));
  background-position: 4px 8px, 4px 14px, 4px 20px;
  background-repeat: no-repeat;
  background-size: 24px 2px, 18px 2px, 22px 2px;
  content: '';
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: 14px;
  color: #06f;
  font-size: 18px;
  font-weight: 500;
  cursor: pointer;
}

.work-main {
  box-sizing: border-box;
  display: flex;
  justify-content: center;
  height: calc(100vh - 52px);
  min-width: 1040px;
  padding: 24px 40px;
}

.work-main-card {
  box-sizing: border-box;
  display: flex;
  width: min(1260px, 100%);
  height: 100%;
  min-height: 0;
  flex-direction: column;
  padding: 28px 48px 30px;
  border-radius: 16px;
  background: #fff;
  box-shadow: 0 12px 32px rgba(15, 35, 80, 0.08);
}

.card-header,
.hint-row {
  display: flex;
  justify-content: space-between;
}

.card-header {
  align-items: flex-start;
}

.page-title {
  color: #111827;
  font-size: 22px;
  font-weight: 600;
  line-height: 1.3;
}

.page-subtitle {
  margin-top: 6px;
  color: #667085;
  font-size: 14px;
}

.hint-row {
  gap: 16px;
  margin-top: 18px;
}

.hint-card {
  flex: 1;
  min-height: 94px;
  padding: 16px 20px;
  border-radius: 8px;
  color: #222;
  font-size: 14px;
  line-height: 1.75;
}

.hint-card--ok {
  background: rgba(0, 100, 255, 0.05);
}

.hint-card--danger {
  background: rgba(255, 50, 50, 0.05);
}

.hint-title {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 15px;
  font-weight: 600;
}

.hint-icon {
  font-size: 18px;
}

.hint-icon--ok {
  color: #12b76a;
}

.hint-icon--danger {
  color: #f04438;
}

.upload-box {
  min-height: 250px;
  flex: 1;
  margin-top: 18px;
}

.upload-box :deep(.ant-upload),
.upload-box :deep(.ant-upload-drag) {
  width: 100%;
  height: 100%;
}

.upload-inner {
  display: flex;
  height: 100%;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.upload-icon {
  color: #1677ff;
  font-size: 44px;
}

.upload-btn {
  width: 120px;
  height: 32px;
  margin-top: 12px;
  border-radius: 8px;
}

.upload-desc {
  margin-top: 8px;
  color: #888;
  font-size: 14px;
}

.template-link {
  margin-top: 6px;
  font-size: 14px;
}

@media (max-height: 820px) {
  .work-main {
    padding: 18px 36px;
  }

  .work-main-card {
    padding: 24px 44px 26px;
  }

  .hint-row {
    margin-top: 14px;
  }

  .hint-card {
    min-height: 84px;
    padding: 14px 18px;
    line-height: 1.65;
  }

  .upload-box {
    min-height: 220px;
    margin-top: 14px;
  }
}
</style>
