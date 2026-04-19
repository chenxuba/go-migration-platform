<script setup>
import { BankOutlined } from '@ant-design/icons-vue'
import { useMessage } from '@/composables/global-config'

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  options: {
    type: Array,
    default: () => [],
  },
  confirmLoading: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['update:open', 'confirm', 'cancel'])

const { t } = useI18nLocale()
const message = useMessage()
const selectedInstitutionKey = ref('')

const openProxy = computed({
  get: () => props.open,
  set: value => emit('update:open', value),
})

const selectedInstitutionOption = computed(() => {
  return props.options.find(item => getInstitutionOptionKey(item) === selectedInstitutionKey.value) || null
})

watch(
  () => [props.open, props.options],
  () => {
    if (!props.open) {
      selectedInstitutionKey.value = ''
      return
    }
    selectedInstitutionKey.value = props.options.length > 0 ? getInstitutionOptionKey(props.options[0]) : ''
  },
  { immediate: true, deep: true },
)

function getInstitutionOptionKey(item) {
  return `${item.instId}-${item.userId}`
}

function maskMobile(mobile) {
  const value = String(mobile || '').trim()
  if (!/^1\d{10}$/.test(value))
    return value
  return `${value.slice(0, 3)}****${value.slice(-4)}`
}

function getOrgInitial(orgName) {
  const value = String(orgName || '').trim()
  if (!value)
    return '校'
  return value.slice(0, 1).toUpperCase()
}

function handleCancel() {
  emit('update:open', false)
  emit('cancel')
}

function handleConfirm() {
  if (!selectedInstitutionOption.value) {
    message.warning(t('pages.login.institutionPicker.required', '请选择要登录的机构'))
    return
  }
  emit('confirm', selectedInstitutionOption.value)
}
</script>

<template>
  <a-modal
    v-model:open="openProxy"
    centered
    :mask-closable="false"
    :width="680"
    wrap-class-name="institution-picker-modal"
    @cancel="handleCancel"
  >
    <template #title>
      <div class="institution-picker-title">
        {{ t('pages.login.institutionPicker.title', '选择登录机构') }}
      </div>
    </template>
    <div class="institution-picker-shell">
      <div class="institution-picker-summary">
        <div class="institution-picker-summary__accent" />
        <div class="institution-picker-summary__content">
          <div class="institution-picker-summary__eyebrow">
            {{ t('pages.login.institutionPicker.eyebrow', '多机构登录识别') }}
          </div>
          <div class="institution-picker-summary__title">
            {{ t('pages.login.institutionPicker.summary', '该登录账号关联多个机构') }}
          </div>
          <div class="institution-picker-summary__desc">
            {{ t('pages.login.institutionPicker.description', '请选择本次要进入的机构，确认后将直接登录对应后台。') }}
          </div>
        </div>
        <div class="institution-picker-summary__badge">
          <strong>{{ options.length }}</strong>
          <span>{{ t('pages.login.institutionPicker.countUnit', '个机构') }}</span>
        </div>
      </div>
      <div class="institution-picker-list">
        <div
          v-for="item in options"
          :key="getInstitutionOptionKey(item)"
          class="institution-picker-card"
          :class="{ 'is-active': selectedInstitutionKey === getInstitutionOptionKey(item) }"
          @click="selectedInstitutionKey = getInstitutionOptionKey(item)"
        >
          <div class="institution-picker-card__top">
            <div class="institution-picker-card__identity">
              <div class="institution-picker-card__logo">
                <img
                  v-if="item.logo"
                  :src="item.logo"
                  :alt="item.orgName || t('pages.login.institutionPicker.unknownOrg', '机构 logo')"
                >
                <span v-else>{{ getOrgInitial(item.orgName) }}</span>
              </div>
              <div class="institution-picker-card__heading">
                <div class="institution-picker-card__title-row">
                  <div class="institution-picker-card__name">
                    <BankOutlined />
                    <span>{{ item.orgName || t('pages.login.institutionPicker.unknownOrg', '未命名机构') }}</span>
                  </div>
                  <a-tag v-if="item.admin">
                    {{ t('pages.login.institutionPicker.admin', '超级管理员') }}
                  </a-tag>
                </div>
                <div class="institution-picker-card__nick">
                  {{ item.nickName || t('pages.login.institutionPicker.noName', '未设置姓名') }}
                </div>
              </div>
            </div>
            <div class="institution-picker-card__status">
              <div class="institution-picker-card__radio">
                <span class="institution-picker-card__radio-dot" />
              </div>
              <span class="institution-picker-card__status-text">
                {{ selectedInstitutionKey === getInstitutionOptionKey(item) ? t('pages.login.institutionPicker.selected', '当前进入') : t('pages.login.institutionPicker.select', '点击选择') }}
              </span>
            </div>
          </div>
          <div class="institution-picker-card__meta-row">
            <div class="institution-picker-card__meta-item">
              <label>{{ t('pages.login.institutionPicker.loginName', '登录账号') }}</label>
              <span>{{ item.loginName || '--' }}</span>
            </div>
            <div class="institution-picker-card__meta-divider" />
            <div class="institution-picker-card__meta-item">
              <label>{{ t('pages.login.institutionPicker.mobile', '手机号') }}</label>
              <span>{{ maskMobile(item.mobile) || '--' }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <template #footer>
      <a-button @click="handleCancel">
        {{ t('common.cancel', '取消') }}
      </a-button>
      <a-button type="primary" :loading="confirmLoading" @click="handleConfirm">
        {{ t('pages.login.institutionPicker.confirm', '确认登录') }}
      </a-button>
    </template>
  </a-modal>
</template>

<style lang="less">
.institution-picker-modal {
  .ant-modal-content {
    border-radius: 28px;
    padding: 0;
    overflow: hidden;
    background:
      linear-gradient(180deg, rgba(248, 251, 255, 0.95) 0%, #ffffff 14%),
      #fff;
    box-shadow: 0 28px 88px rgba(15, 23, 42, 0.18);
  }

  .ant-modal-header {
    padding: 26px 30px 0;
    margin-bottom: 0;
    border-bottom: none;
    background: transparent;
  }

  .ant-modal-close {
    top: 18px;
    inset-inline-end: 22px;
    color: #94a3b8;
    border-radius: 14px;
    transition: all 0.2s ease;
  }

  .ant-modal-close:hover {
    color: #334155;
    background: rgba(148, 163, 184, 0.12);
  }

  .ant-modal-body {
    padding: 20px 30px 12px;
  }

  .ant-modal-footer {
    padding: 18px 30px 26px;
    border-top: 1px solid rgba(148, 163, 184, 0.14);
    background: #fff;
  }

  .ant-tag {
    margin-inline-end: 0;
    border-radius: 999px;
    border-color: rgba(59, 130, 246, 0.10);
    background: rgba(59, 130, 246, 0.05);
    color: #2563eb;
    font-weight: 500;
  }

  .ant-btn {
    height: 44px;
    min-width: 116px;
    padding: 0 24px;
    border-radius: 16px;
    font-weight: 600;
  }

  .ant-btn-default {
    border-color: rgba(148, 163, 184, 0.24);
    color: #334155;
    box-shadow: none;
  }

  .ant-btn-primary {
    box-shadow: 0 12px 22px rgba(37, 99, 235, 0.18);
  }

  .institution-picker-title {
    font-size: 25px;
    font-weight: 700;
    color: #0f172a;
    line-height: 36px;
  }

  .institution-picker-shell {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .institution-picker-summary {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 18px;
    padding: 18px 20px;
    border-radius: 22px;
    border: 1px solid rgba(59, 130, 246, 0.10);
    background:
      linear-gradient(135deg, rgba(248, 251, 255, 1) 0%, rgba(255, 255, 255, 1) 68%),
      #fff;
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
  }

  .institution-picker-summary__accent {
    width: 4px;
    align-self: stretch;
    border-radius: 999px;
    background: linear-gradient(180deg, #60a5fa 0%, #2563eb 100%);
    box-shadow: 0 0 0 6px rgba(96, 165, 250, 0.08);
    flex: 0 0 auto;
  }

  .institution-picker-summary__content {
    flex: 1 1 auto;
    min-width: 0;
  }

  .institution-picker-summary__eyebrow {
    font-size: 12px;
    line-height: 20px;
    font-weight: 600;
    color: #3b82f6;
    letter-spacing: 0.08em;
  }

  .institution-picker-summary__title {
    margin-top: 4px;
    font-size: 18px;
    line-height: 28px;
    font-weight: 600;
    color: #0f172a;
  }

  .institution-picker-summary__desc {
    margin-top: 6px;
    max-width: 400px;
    font-size: 13px;
    line-height: 22px;
    color: #64748b;
  }

  .institution-picker-summary__badge {
    flex: 0 0 auto;
    min-width: 96px;
    padding: 12px 14px;
    border-radius: 18px;
    background: rgba(255, 255, 255, 0.92);
    border: 1px solid rgba(148, 163, 184, 0.14);
    text-align: center;
    box-shadow: 0 10px 28px rgba(15, 23, 42, 0.05);
  }

  .institution-picker-summary__badge strong {
    display: block;
    font-size: 24px;
    line-height: 30px;
    font-weight: 700;
    color: #0f172a;
  }

  .institution-picker-summary__badge span {
    display: block;
    margin-top: 4px;
    font-size: 11px;
    line-height: 18px;
    color: #64748b;
  }

  .institution-picker-list {
    max-height: 392px;
    overflow-y: auto;
    padding-right: 6px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .institution-picker-list::-webkit-scrollbar {
    width: 8px;
  }

  .institution-picker-list::-webkit-scrollbar-thumb {
    border-radius: 999px;
    background: rgba(148, 163, 184, 0.48);
    border: 2px solid transparent;
    background-clip: padding-box;
  }

  .institution-picker-list::-webkit-scrollbar-track {
    background: transparent;
  }

  .institution-picker-card {
    position: relative;
    padding: 16px 18px 14px;
    border-radius: 22px;
    border: 1px solid rgba(226, 232, 240, 0.9);
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 1) 0%, rgba(251, 252, 254, 1) 100%);
    cursor: pointer;
    transition: all 0.24s ease;
    box-shadow: 0 10px 24px rgba(15, 23, 42, 0.05);
  }

  .institution-picker-card:hover {
    transform: translateY(-1px);
    border-color: rgba(59, 130, 246, 0.18);
    box-shadow: 0 14px 30px rgba(15, 23, 42, 0.08);
  }

  .institution-picker-card.is-active {
    border-color: rgba(59, 130, 246, 0.28);
    background:
      linear-gradient(180deg, rgba(255, 255, 255, 1) 0%, rgba(247, 250, 255, 1) 100%);
    box-shadow: 0 16px 34px rgba(59, 130, 246, 0.10);
  }

  .institution-picker-card.is-active::before {
    position: absolute;
    inset: 0;
    border-radius: 22px;
    padding: 1px;
    background: linear-gradient(180deg, rgba(96, 165, 250, 0.36), rgba(37, 99, 235, 0.14));
    content: "";
    -webkit-mask:
      linear-gradient(#fff 0 0) content-box,
      linear-gradient(#fff 0 0);
    -webkit-mask-composite: xor;
    mask-composite: exclude;
    pointer-events: none;
  }

  .institution-picker-card__top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }

  .institution-picker-card__identity {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
    flex: 1 1 auto;
  }

  .institution-picker-card__logo {
    width: 48px;
    height: 48px;
    border-radius: 16px;
    background: linear-gradient(135deg, #2563eb 0%, #60a5fa 100%);
    box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.24);
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    color: #fff;
    font-size: 18px;
    font-weight: 700;
    flex: 0 0 auto;
  }

  .institution-picker-card__logo img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .institution-picker-card__heading {
    min-width: 0;
    flex: 1 1 auto;
  }

  .institution-picker-card__title-row {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex-wrap: wrap;
  }

  .institution-picker-card__name {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    font-size: 15px;
    font-weight: 600;
    color: #0f172a;
    line-height: 24px;
  }

  .institution-picker-card__name span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .institution-picker-card__nick {
    margin-top: 5px;
    min-width: 0;
    font-size: 13px;
    color: #7c8aa5;
    line-height: 20px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .institution-picker-card__status {
    flex: 0 0 auto;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .institution-picker-card__radio {
    position: relative;
    width: 22px;
    height: 22px;
    border-radius: 999px;
    border: 1.5px solid rgba(148, 163, 184, 0.32);
    background: #fff;
    flex: 0 0 auto;
  }

  .institution-picker-card__radio-dot {
    position: absolute;
    inset: 4px;
    border-radius: 999px;
    background: transparent;
    transition: all 0.2s ease;
  }

  .institution-picker-card__status-text {
    font-size: 12px;
    line-height: 18px;
    font-weight: 600;
    color: #64748b;
  }

  .institution-picker-card.is-active .institution-picker-card__radio {
    border-color: rgba(59, 130, 246, 0.42);
    box-shadow: 0 0 0 4px rgba(96, 165, 250, 0.10);
  }

  .institution-picker-card.is-active .institution-picker-card__radio-dot {
    background: linear-gradient(180deg, #60a5fa 0%, #2563eb 100%);
  }

  .institution-picker-card.is-active .institution-picker-card__status-text {
    color: #2563eb;
  }

  .institution-picker-card__meta-row {
    display: flex;
    align-items: stretch;
    gap: 0;
    margin-top: 14px;
    border-radius: 16px;
    background: rgba(248, 250, 252, 0.9);
    border: 1px solid rgba(226, 232, 240, 0.72);
    overflow: hidden;
  }

  .institution-picker-card__meta-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
    flex: 1 1 0;
    padding: 11px 14px 12px;
  }

  .institution-picker-card__meta-divider {
    width: 1px;
    background: rgba(226, 232, 240, 0.92);
    flex: 0 0 auto;
  }

  .institution-picker-card__meta-item label {
    font-size: 12px;
    line-height: 18px;
    color: #64748b;
  }

  .institution-picker-card__meta-item span {
    min-width: 0;
    font-size: 14px;
    line-height: 22px;
    font-weight: 600;
    color: #1e293b;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

@media (max-width: 768px) {
  .institution-picker-modal {
    .ant-modal {
      max-width: calc(100vw - 24px);
      margin: 0 auto;
    }

    .ant-modal-header,
    .ant-modal-body,
    .ant-modal-footer {
      padding-left: 18px;
      padding-right: 18px;
    }

    .institution-picker-title {
      font-size: 22px;
      line-height: 32px;
    }

    .institution-picker-summary {
      flex-direction: column;
      gap: 14px;
      padding: 18px;
      align-items: flex-start;
    }

    .institution-picker-summary__accent {
      width: 100%;
      height: 4px;
      min-height: 4px;
      box-shadow: 0 0 0 4px rgba(96, 165, 250, 0.06);
    }

    .institution-picker-summary__badge {
      width: 100%;
      display: flex;
      align-items: baseline;
      justify-content: space-between;
    }

    .institution-picker-card__top {
      align-items: flex-start;
      flex-direction: column;
    }

    .institution-picker-card__status {
      align-self: flex-end;
    }

    .institution-picker-card__meta-row {
      flex-direction: column;
    }

    .institution-picker-card__meta-divider {
      width: auto;
      height: 1px;
    }
  }
}

@media (max-width: 480px) {
  .institution-picker-modal {
    .ant-modal-header {
      padding-top: 20px;
    }

    .ant-modal-footer {
      padding-bottom: 18px;
    }

    .institution-picker-summary__title {
      font-size: 16px;
      line-height: 24px;
    }

    .institution-picker-card {
      padding: 16px 14px 14px;
    }

    .institution-picker-card__identity {
      gap: 12px;
    }

    .institution-picker-card__logo {
      width: 44px;
      height: 44px;
      border-radius: 14px;
      font-size: 18px;
    }

    .institution-picker-card__name {
      font-size: 15px;
      line-height: 24px;
    }

    .institution-picker-card__status {
      width: 100%;
      justify-content: flex-end;
    }

    .institution-picker-card__meta-item span {
      font-size: 14px;
    }
  }
}
</style>
