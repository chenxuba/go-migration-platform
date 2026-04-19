<script setup>
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

function selectOption(item) {
  selectedInstitutionKey.value = getInstitutionOptionKey(item)
}
</script>

<template>
  <a-modal
    v-model:open="openProxy"
    centered
    :mask-closable="false"
    :width="600"
    wrap-class-name="institution-picker-modal"
    @cancel="handleCancel"
  >
    <template #title>
      <span class="ipm-title">{{ t('pages.login.institutionPicker.title', '选择登录机构') }}</span>
    </template>

    <div class="ipm">
      <header class="ipm-header">
        <p class="ipm-header__line">
          {{ t('pages.login.institutionPicker.summary', '该登录账号关联多个机构') }}
        </p>
        <p class="ipm-header__meta">
          <span>{{ t('pages.login.institutionPicker.description', '请选择本次要进入的机构，确认后将直接登录对应后台。') }}</span>
          <span class="ipm-header__count" aria-live="polite">
            {{ options.length }}{{ t('pages.login.institutionPicker.countUnit', '个机构') }}
          </span>
        </p>
      </header>

      <div
        class="ipm-list"
        role="listbox"
        :aria-label="t('pages.login.institutionPicker.title', '选择登录机构')"
      >
        <button
          v-for="item in options"
          :key="getInstitutionOptionKey(item)"
          type="button"
          class="ipm-item"
          :class="{ 'is-selected': selectedInstitutionKey === getInstitutionOptionKey(item) }"
          :aria-selected="selectedInstitutionKey === getInstitutionOptionKey(item)"
          @click="selectOption(item)"
        >
          <span class="ipm-item__radio" aria-hidden="true" />
          <span class="ipm-item__avatar" aria-hidden="true">
            <img
              v-if="item.logo"
              :src="item.logo"
              :alt="item.orgName || t('pages.login.institutionPicker.unknownOrg', '机构 logo')"
            >
            <span v-else>{{ getOrgInitial(item.orgName) }}</span>
          </span>
          <span class="ipm-item__body">
            <span class="ipm-item__row">
              <span class="ipm-item__org">{{ item.orgName || t('pages.login.institutionPicker.unknownOrg', '未命名机构') }}</span>
              <span v-if="item.admin" class="ipm-item__tag">{{ t('pages.login.institutionPicker.admin', '超级管理员') }}</span>
            </span>
            <span class="ipm-item__sub">{{ item.nickName || t('pages.login.institutionPicker.noName', '未设置姓名') }}</span>
            <span class="ipm-item__detail">
              <span>{{ item.loginName || '--' }}</span>
              <span class="ipm-item__sep" aria-hidden="true" />
              <span>{{ maskMobile(item.mobile) || '--' }}</span>
            </span>
          </span>
        </button>
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
  --ipm-text: rgba(0, 0, 0, 0.88);
  --ipm-secondary: rgba(0, 0, 0, 0.45);
  --ipm-tertiary: rgba(0, 0, 0, 0.35);
  --ipm-border: rgba(0, 0, 0, 0.06);
  --ipm-fill: rgba(0, 0, 0, 0.02);
  --ipm-primary: var(--pro-ant-color-primary, #1677ff);

  .ant-modal-content {
    padding: 0;
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 3px 6px -4px rgba(0, 0, 0, 0.12);
  }

  .ant-modal-header {
    margin: 0;
    padding: 20px 24px 16px;
    border-bottom: 1px solid var(--ipm-border);
  }

  .ant-modal-title {
    margin: 0;
  }

  .ant-modal-close {
    top: 18px;
    inset-inline-end: 18px;
    color: var(--ipm-secondary);
  }

  .ant-modal-body {
    padding: 0;
  }

  .ant-modal-footer {
    margin: 0;
    padding: 12px 24px 16px;
    border-top: 1px solid var(--ipm-border);
  }

  .ipm-title {
    font-size: 16px;
    font-weight: 600;
    line-height: 24px;
    color: var(--ipm-text);
  }

  .ipm-header {
    padding: 16px 24px 0;
  }

  .ipm-header__line {
    margin: 0;
    font-size: 14px;
    line-height: 22px;
    color: var(--ipm-text);
  }

  .ipm-header__meta {
    margin: 8px 0 0;
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 8px 16px;
    font-size: 13px;
    line-height: 20px;
    color: var(--ipm-secondary);
  }

  .ipm-header__meta > span:first-child {
    flex: 1 1 200px;
    min-width: 0;
  }

  .ipm-header__count {
    flex-shrink: 0;
    font-variant-numeric: tabular-nums;
    color: var(--ipm-tertiary);
    font-size: 12px;
  }

  .ipm-list {
    padding: 16px 24px 20px;
    max-height: 360px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .ipm-item {
    appearance: none;
    box-sizing: border-box;
    display: flex;
    align-items: flex-start;
    gap: 12px;
    width: 100%;
    margin: 0;
    padding: 12px 14px;
    font: inherit;
    text-align: left;
    color: inherit;
    cursor: pointer;
    border: 1px solid var(--ipm-border);
    border-radius: 8px;
    background: #fff;
    transition: border-color 0.15s ease, background 0.15s ease;
  }

  .ipm-item:hover {
    border-color: rgba(0, 0, 0, 0.12);
    background: var(--ipm-fill);
  }

  .ipm-item.is-selected {
    border-color: var(--ipm-primary);
    background: rgba(22, 119, 255, 0.04);
  }

  .ipm-item:focus-visible {
    outline: 2px solid var(--ipm-primary);
    outline-offset: 1px;
  }

  .ipm-item__radio {
    flex-shrink: 0;
    width: 16px;
    height: 16px;
    margin-top: 3px;
    border-radius: 50%;
    border: 1px solid rgba(0, 0, 0, 0.25);
    position: relative;
  }

  .ipm-item.is-selected .ipm-item__radio {
    border-color: var(--ipm-primary);
  }

  .ipm-item.is-selected .ipm-item__radio::after {
    content: "";
    position: absolute;
    top: 50%;
    left: 50%;
    width: 8px;
    height: 8px;
    margin: -4px 0 0 -4px;
    border-radius: 50%;
    background: var(--ipm-primary);
  }

  .ipm-item__avatar {
    flex-shrink: 0;
    width: 40px;
    height: 40px;
    border-radius: 8px;
    overflow: hidden;
    background: var(--ipm-fill);
    border: 1px solid var(--ipm-border);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    font-weight: 600;
    color: var(--ipm-secondary);
  }

  .ipm-item__avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .ipm-item__body {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .ipm-item__row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .ipm-item__org {
    font-size: 14px;
    font-weight: 500;
    line-height: 22px;
    color: var(--ipm-text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .ipm-item__tag {
    flex-shrink: 0;
    font-size: 12px;
    line-height: 20px;
    color: var(--ipm-secondary);
  }

  .ipm-item__sub {
    font-size: 13px;
    line-height: 20px;
    color: var(--ipm-secondary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .ipm-item__detail {
    margin-top: 2px;
    font-size: 12px;
    line-height: 18px;
    color: var(--ipm-tertiary);
    font-variant-numeric: tabular-nums;
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0 6px;
    min-width: 0;
  }

  .ipm-item__detail > span:first-child,
  .ipm-item__detail > span:last-child {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 100%;
  }

  .ipm-item__sep {
    width: 3px;
    height: 3px;
    border-radius: 50%;
    background: var(--ipm-tertiary);
    opacity: 0.5;
    flex-shrink: 0;
  }
}

@media (max-width: 576px) {
  .institution-picker-modal {
    .ant-modal {
      max-width: calc(100vw - 32px);
    }

    .ipm-header,
    .ipm-list {
      padding-left: 16px;
      padding-right: 16px;
    }

    .ant-modal-header,
    .ant-modal-footer {
      padding-left: 16px;
      padding-right: 16px;
    }
  }
}
</style>
