<script setup lang="ts">
import type { FormInstance, Rule } from 'ant-design-vue/es/form'
import type { InstitutionMutationPayload } from '@/api/platform/institutions'
import {
  createInstitutionApi,
  getInstitutionDetailApi,
  updateInstitutionApi,
} from '@/api/platform/institutions'
import messageService from '@/utils/messageService'

const props = defineProps<{
  open: boolean
  institutionId?: number | null
}>()

const emit = defineEmits<{
  (e: 'update:open', value: boolean): void
  (e: 'saved'): void
}>()

interface InstitutionFormState {
  id?: number
  organCode: string
  organName: string
  loginName: string
  mobile: string
  principal: string
  province: string
  city: string
  region: string
  address: string
  concatPhone: string
  fixedPhone: string
  remark: string
  logo: string
  enabled: boolean
}

const formRef = ref<FormInstance>()
const detailLoading = ref(false)
const submitting = ref(false)

const formState = reactive<InstitutionFormState>({
  organCode: '',
  organName: '',
  loginName: '',
  mobile: '',
  principal: '',
  province: '',
  city: '',
  region: '',
  address: '',
  concatPhone: '',
  fixedPhone: '',
  remark: '',
  logo: '',
  enabled: true,
})

const isEdit = computed(() => Number(props.institutionId || 0) > 0)
const drawerTitle = computed(() => (isEdit.value ? '编辑机构' : '新建机构'))

const rules: Record<string, Rule[]> = {
  organName: [{ required: true, message: '请输入机构名称', trigger: 'blur' }],
  loginName: [{ required: true, message: '请输入登录账号', trigger: 'blur' }],
  mobile: [
    { required: true, message: '请输入联系电话', trigger: 'blur' },
    { pattern: /^\d{11}$/, message: '联系电话需为 11 位手机号', trigger: 'blur' },
  ],
  province: [{ required: true, message: '请输入省份', trigger: 'blur' }],
  city: [{ required: true, message: '请输入城市', trigger: 'blur' }],
}

function resetForm() {
  formState.id = undefined
  formState.organCode = ''
  formState.organName = ''
  formState.loginName = ''
  formState.mobile = ''
  formState.principal = ''
  formState.province = ''
  formState.city = ''
  formState.region = ''
  formState.address = ''
  formState.concatPhone = ''
  formState.fixedPhone = ''
  formState.remark = ''
  formState.logo = ''
  formState.enabled = true
  nextTick(() => {
    formRef.value?.clearValidate?.()
  })
}

function closeDrawer() {
  emit('update:open', false)
}

async function loadInstitutionDetail(id: number) {
  detailLoading.value = true
  try {
    const res = await getInstitutionDetailApi({ id })
    if (res.code !== 200 || !res.result) {
      messageService.error(res.message || '获取机构详情失败')
      return
    }

    const detail = res.result
    formState.id = detail.id
    formState.organCode = String(detail.organCode || '')
    formState.organName = String(detail.organName || '')
    formState.loginName = String(detail.loginName || '')
    formState.mobile = String(detail.mobile || '')
    formState.principal = String(detail.principal || '')
    formState.province = String(detail.province || '')
    formState.city = String(detail.city || '')
    formState.region = String(detail.region || '')
    formState.address = String(detail.address || '')
    formState.concatPhone = String(detail.concatPhone || '')
    formState.fixedPhone = String(detail.fixedPhone || '')
    formState.remark = String(detail.remark || '')
    formState.logo = String(detail.logo || '')
    formState.enabled = !!detail.enabled
  }
  catch (error: any) {
    console.error('load institution detail failed', error)
    messageService.error(error?.message || '获取机构详情失败')
  }
  finally {
    detailLoading.value = false
  }
}

watch(
  () => [props.open, props.institutionId] as const,
  ([open, institutionId]) => {
    if (!open) {
      resetForm()
      return
    }

    if (institutionId) {
      void loadInstitutionDetail(Number(institutionId))
      return
    }

    resetForm()
  },
  { immediate: true },
)

function buildPayload(): InstitutionMutationPayload {
  return {
    organName: formState.organName.trim(),
    loginName: formState.loginName.trim(),
    mobile: formState.mobile.trim(),
    principal: formState.principal.trim() || undefined,
    province: formState.province.trim(),
    city: formState.city.trim(),
    region: formState.region.trim() || undefined,
    address: formState.address.trim() || undefined,
    concatPhone: formState.concatPhone.trim() || undefined,
    fixedPhone: formState.fixedPhone.trim() || undefined,
    remark: formState.remark.trim() || undefined,
    logo: formState.logo.trim() || undefined,
    enabled: !!formState.enabled,
  }
}

async function submitForm() {
  try {
    await formRef.value?.validate()
  }
  catch {
    return
  }

  submitting.value = true
  try {
    const payload = buildPayload()
    const res = isEdit.value && formState.id
      ? await updateInstitutionApi({ id: formState.id, ...payload })
      : await createInstitutionApi(payload)

    if (res.code !== 200) {
      messageService.error(res.message || (isEdit.value ? '更新机构失败' : '新增机构失败'))
      return
    }

    messageService.success(isEdit.value ? '机构更新成功' : '机构新增成功')
    emit('saved')
    closeDrawer()
  }
  catch (error: any) {
    console.error('submit institution failed', error)
    messageService.error(error?.message || (isEdit.value ? '更新机构失败' : '新增机构失败'))
  }
  finally {
    submitting.value = false
  }
}
</script>

<template>
  <a-drawer
    :open="open"
    :title="drawerTitle"
    :width="640"
    :mask-closable="false"
    destroy-on-close
    @close="closeDrawer"
    @update:open="emit('update:open', $event)"
  >
    <a-spin :spinning="detailLoading">
      <a-form ref="formRef" :model="formState" :rules="rules" layout="vertical" class="institution-form">
        <div class="institution-form__grid">
          <a-form-item label="机构名称" name="organName">
            <a-input v-model:value="formState.organName" :maxlength="64" placeholder="请输入机构名称" />
          </a-form-item>

          <a-form-item label="登录账号" name="loginName">
            <a-input v-model:value="formState.loginName" :maxlength="22" placeholder="请输入登录账号" />
          </a-form-item>

          <a-form-item label="联系电话" name="mobile">
            <a-input v-model:value="formState.mobile" :maxlength="11" placeholder="请输入 11 位手机号" />
          </a-form-item>

          <a-form-item label="负责人">
            <a-input v-model:value="formState.principal" :maxlength="64" placeholder="请输入负责人姓名" />
          </a-form-item>

          <a-form-item label="省份" name="province">
            <a-input v-model:value="formState.province" :maxlength="32" placeholder="例如：四川省" />
          </a-form-item>

          <a-form-item label="城市" name="city">
            <a-input v-model:value="formState.city" :maxlength="32" placeholder="例如：成都市" />
          </a-form-item>

          <a-form-item label="区县">
            <a-input v-model:value="formState.region" :maxlength="32" placeholder="例如：高新区" />
          </a-form-item>

          <a-form-item label="机构电话">
            <a-input v-model:value="formState.concatPhone" :maxlength="40" placeholder="请输入机构电话" />
          </a-form-item>
        </div>

        <a-form-item label="机构编码">
          <a-input :value="isEdit ? formState.organCode : '保存后自动生成'" disabled />
        </a-form-item>

        <a-form-item label="详细地址">
          <a-input v-model:value="formState.address" :maxlength="256" placeholder="请输入详细地址" />
        </a-form-item>

        <a-form-item label="固定电话">
          <a-input v-model:value="formState.fixedPhone" :maxlength="255" placeholder="请输入固定电话" />
        </a-form-item>

        <a-form-item label="Logo 地址">
          <a-input v-model:value="formState.logo" :maxlength="2000" placeholder="可选，支持 http/https 图片地址" />
        </a-form-item>

        <a-form-item label="备注">
          <a-textarea v-model:value="formState.remark" :maxlength="255" :rows="4" placeholder="补充说明信息" />
        </a-form-item>

        <div class="institution-form__switch">
          <span class="institution-form__switch-label">启用状态</span>
          <a-switch v-model:checked="formState.enabled" checked-children="启用" un-checked-children="停用" />
        </div>
      </a-form>
    </a-spin>

    <template #footer>
      <div class="institution-form__footer">
        <a-button @click="closeDrawer">
          取消
        </a-button>
        <a-button type="primary" :loading="submitting" @click="submitForm">
          保存
        </a-button>
      </div>
    </template>
  </a-drawer>
</template>

<style scoped>
.institution-form {
  padding-bottom: 20px;
}

.institution-form__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 16px;
}

.institution-form__switch {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  border: 1px solid #eaecf0;
  border-radius: 14px;
  background: #f8fafc;
}

.institution-form__switch-label {
  color: #344054;
  font-weight: 600;
}

.institution-form__footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

@media (max-width: 768px) {
  .institution-form__grid {
    grid-template-columns: 1fr;
  }
}
</style>
