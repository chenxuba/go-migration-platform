<script setup>
import { BankOutlined, CheckOutlined } from '@ant-design/icons-vue'
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
    :width="720"
    wrap-class-name="institution-picker-modal"
    @cancel="handleCancel"
  >
    <template #title>
      <div class="ipm-head">
        <div class="ipm-head__title">
          {{ t('pages.login.institutionPicker.title', '选择登录机构') }}
        </div>
        <p class="ipm-head__sub">
          {{ t('pages.login.institutionPicker.summary', '该登录账号关联多个机构') }}
        </p>
      </div>
    </template>

    <div class="ipm-body">
      <section class="ipm-intro" aria-labelledby="ipm-intro-heading">
        <div class="ipm-intro__main">
          <span id="ipm-intro-heading" class="ipm-intro__kicker">
            {{ t('pages.login.institutionPicker.eyebrow', '多机构登录识别') }}
          </span>
          <p class="ipm-intro__text">
            {{ t('pages.login.institutionPicker.description', '请选择本次要进入的机构，确认后将直接登录对应后台。') }}
          </p>
        </div>
        <div class="ipm-intro__stat" aria-live="polite">
          <span class="ipm-intro__stat-num">{{ options.length }}</span>
          <span class="ipm-intro__stat-label">{{ t('pages.login.institutionPicker.countUnit', '个机构') }}</span>
        </div>
      </section>

      <div class="ipm-list" role="listbox" :aria-label="t('pages.login.institutionPicker.title', '选择登录机构')">
        <button
          v-for="item in options"
          :key="getInstitutionOptionKey(item)"
          type="button"
          class="ipm-card"
          :class="{ 'is-selected': selectedInstitutionKey === getInstitutionOptionKey(item) }"
          :aria-selected="selectedInstitutionKey === getInstitutionOptionKey(item)"
          @click="selectOption(item)"
        >
          <span class="ipm-card__accent" aria-hidden="true" />
          <div class="ipm-card__row">
            <div class="ipm-card__avatar" aria-hidden="true">
              <img
                v-if="item.logo"
                :src="item.logo"
                :alt="item.orgName || t('pages.login.institutionPicker.unknownOrg', '机构 logo')"
              >
              <span v-else class="ipm-card__avatar-fallback">{{ getOrgInitial(item.orgName) }}</span>
            </div>
            <div class="ipm-card__core">
              <div class="ipm-card__title-line">
                <span class="ipm-card__org">
                  <BankOutlined class="ipm-card__org-icon" />
                  <span class="ipm-card__org-name">{{ item.orgName || t('pages.login.institutionPicker.unknownOrg', '未命名机构') }}</span>
                </span>
                <span v-if="item.admin" class="ipm-card__role">
                  {{ t('pages.login.institutionPicker.admin', '超级管理员') }}
                </span>
              </div>
              <div class="ipm-card__user">
                {{ item.nickName || t('pages.login.institutionPicker.noName', '未设置姓名') }}
              </div>
              <div class="ipm-card__chips">
                <span class="ipm-chip">
                  <span class="ipm-chip__label">{{ t('pages.login.institutionPicker.loginName', '登录账号') }}</span>
                  <span class="ipm-chip__value">{{ item.loginName || '--' }}</span>
                </span>
                <span class="ipm-chip">
                  <span class="ipm-chip__label">{{ t('pages.login.institutionPicker.mobile', '手机号') }}</span>
                  <span class="ipm-chip__value">{{ maskMobile(item.mobile) || '--' }}</span>
                </span>
              </div>
            </div>
            <div class="ipm-card__pick">
              <span
                class="ipm-card__pick-ring"
                :class="{ 'is-on': selectedInstitutionKey === getInstitutionOptionKey(item) }"
              >
                <CheckOutlined v-if="selectedInstitutionKey === getInstitutionOptionKey(item)" class="ipm-card__pick-check" />
              </span>
              <span class="ipm-card__pick-label">
                {{ selectedInstitutionKey === getInstitutionOptionKey(item) ? t('pages.login.institutionPicker.selected', '当前进入') : t('pages.login.institutionPicker.select', '点击选择') }}
              </span>
            </div>
          </div>
        </button>
      </div>
    </div>

    <template #footer>
      <div class="ipm-footer">
        <a-button class="ipm-footer__btn" size="large" @click="handleCancel">
          {{ t('common.cancel', '取消') }}
        </a-button>
        <a-button
          class="ipm-footer__btn ipm-footer__btn--primary"
          type="primary"
          size="large"
          :loading="confirmLoading"
          @click="handleConfirm"
        >
          {{ t('pages.login.institutionPicker.confirm', '确认登录') }}
        </a-button>
      </div>
    </template>
  </a-modal>
</template>

<style lang="less">
/* Modern institution picker — self-contained under wrap class */
.institution-picker-modal {
  --ipm-bg: #ffffff;
  --ipm-surface: #ffffff;
  --ipm-ink: #1f2329;
  --ipm-muted: #86909c;
  --ipm-line: #e5e6eb;
  --ipm-subtle: #f7f8fa;
  --ipm-accent: var(--pro-ant-color-primary, #1677ff);
  --ipm-accent-soft: rgba(22, 119, 255, 0.08);
  --ipm-accent-line: rgba(22, 119, 255, 0.22);
  --ipm-radius: 18px;
  --ipm-radius-sm: 12px;

  .ant-modal-content {
    padding: 0;
    overflow: hidden;
    border-radius: var(--ipm-radius);
    border: 1px solid var(--ipm-line);
    background: var(--ipm-bg);
    box-shadow: 0 18px 48px rgba(31, 35, 41, 0.12);
  }

  .ant-modal-header {
    margin: 0;
    padding: 24px 24px 18px;
    border-bottom: 1px solid var(--ipm-line);
    background: var(--ipm-surface);
  }

  .ant-modal-title {
    width: 100%;
  }

  .ant-modal-close {
    top: 16px;
    inset-inline-end: 16px;
    color: #8c8c8c;
    width: 32px;
    height: 32px;
    border-radius: 8px;
    transition: color 0.2s ease, background 0.2s ease;
  }

  .ant-modal-close:hover {
    color: #4e5969;
    background: var(--ipm-subtle);
  }

  .ant-modal-body {
    padding: 20px 24px 8px;
    background: var(--ipm-surface);
  }

  .ant-modal-footer {
    margin: 0;
    padding: 16px 24px 20px;
    border-top: 1px solid var(--ipm-line);
    background: var(--ipm-surface);
  }

  .ant-modal-footer .ant-btn + .ant-btn {
    margin-inline-start: 12px;
  }

  /* —— Header —— */
  .ipm-head__title {
    font-size: 20px;
    font-weight: 600;
    color: var(--ipm-ink);
    line-height: 1.4;
    padding-right: 28px;
  }

  .ipm-head__sub {
    margin: 6px 0 0;
    max-width: 520px;
    font-size: 13px;
    line-height: 1.6;
    color: var(--ipm-muted);
    font-weight: 400;
  }

  /* —— Body —— */
  .ipm-intro {
    display: flex;
    align-items: stretch;
    justify-content: space-between;
    gap: 16px;
    padding: 14px 16px;
    margin-bottom: 14px;
    border-radius: var(--ipm-radius-sm);
    background: var(--ipm-subtle);
    border: 1px solid #f0f0f0;
  }

  .ipm-intro__main {
    flex: 1;
    min-width: 0;
  }

  .ipm-intro__kicker {
    display: inline-block;
    font-size: 12px;
    font-weight: 600;
    letter-spacing: 0;
    color: #4e5969;
    line-height: 1.4;
  }

  .ipm-intro__text {
    margin: 6px 0 0;
    font-size: 13px;
    line-height: 1.6;
    color: var(--ipm-muted);
  }

  .ipm-intro__stat {
    flex-shrink: 0;
    align-self: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-width: 88px;
    padding: 10px 14px;
    border-radius: 12px;
    background: #fff;
    border: 1px solid var(--ipm-line);
  }

  .ipm-intro__stat-num {
    font-size: 26px;
    font-weight: 700;
    line-height: 1;
    color: var(--ipm-ink);
    font-variant-numeric: tabular-nums;
  }

  .ipm-intro__stat-label {
    margin-top: 4px;
    font-size: 12px;
    font-weight: 500;
    color: var(--ipm-muted);
  }

  /* —— List —— */
  .ipm-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
    max-height: 380px;
    overflow-y: auto;
    padding-bottom: 4px;
    scrollbar-width: thin;
    scrollbar-color: rgba(100, 116, 139, 0.45) transparent;
  }

  .ipm-list::-webkit-scrollbar {
    width: 6px;
  }

  .ipm-list::-webkit-scrollbar-thumb {
    border-radius: 999px;
    background: rgba(100, 116, 139, 0.35);
  }

  /* —— Card (button) —— */
  .ipm-card {
    appearance: none;
    box-sizing: border-box;
    display: block;
    margin: 0;
    font: inherit;
    width: 100%;
    position: relative;
    text-align: left;
    cursor: pointer;
    border-radius: var(--ipm-radius-sm);
    background: var(--ipm-surface);
    border: 1px solid #eaedf1;
    transition:
      border-color 0.2s ease,
      box-shadow 0.2s ease,
      background 0.2s ease;
  }

  .ipm-card:focus-visible {
    outline: 2px solid var(--ipm-accent);
    outline-offset: 2px;
  }

  .ipm-card__accent {
    position: absolute;
    left: 0;
    top: 12px;
    bottom: 12px;
    width: 2px;
    border-radius: 999px;
    background: transparent;
    transition: background 0.2s ease;
  }

  .ipm-card:hover {
    border-color: #d6e4ff;
    background: #fff;
  }

  .ipm-card.is-selected {
    border-color: var(--ipm-accent-line);
    background: #f8fbff;
    box-shadow: 0 0 0 3px var(--ipm-accent-soft);
  }

  .ipm-card.is-selected .ipm-card__accent {
    background: var(--ipm-accent);
  }

  .ipm-card__row {
    display: flex;
    align-items: flex-start;
    gap: 14px;
    padding: 16px 16px 16px 18px;
  }

  .ipm-card__avatar {
    flex-shrink: 0;
    width: 48px;
    height: 48px;
    border-radius: 12px;
    overflow: hidden;
    background: #eef4ff;
    display: flex;
    align-items: center;
    justify-content: center;
    border: 1px solid #deebff;
  }

  .ipm-card__avatar img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .ipm-card__avatar-fallback {
    font-size: 18px;
    font-weight: 600;
    color: var(--ipm-accent);
  }

  .ipm-card__core {
    flex: 1;
    min-width: 0;
  }

  .ipm-card__title-line {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .ipm-card__org {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    font-size: 15px;
    font-weight: 600;
    color: var(--ipm-ink);
    line-height: 1.35;
  }

  .ipm-card__org-icon {
    flex-shrink: 0;
    font-size: 14px;
    color: #94a3b8;
  }

  .ipm-card__org-name {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .ipm-card__role {
    flex-shrink: 0;
    font-size: 12px;
    font-weight: 500;
    padding: 2px 8px;
    border-radius: 999px;
    color: #4e5969;
    background: #f7f8fa;
    border: 1px solid var(--ipm-line);
  }

  .ipm-card__user {
    margin-top: 4px;
    font-size: 13px;
    color: var(--ipm-muted);
    line-height: 1.4;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .ipm-card__chips {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 12px;
  }

  .ipm-chip {
    display: inline-flex;
    align-items: baseline;
    gap: 6px;
    padding: 7px 10px;
    border-radius: 10px;
    background: #fff;
    border: 1px solid #eff0f1;
    max-width: 100%;
  }

  .ipm-chip__label {
    font-size: 12px;
    font-weight: 500;
    color: var(--ipm-muted);
  }

  .ipm-chip__value {
    font-size: 13px;
    font-weight: 600;
    color: var(--ipm-ink);
    font-variant-numeric: tabular-nums;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 180px;
  }

  .ipm-card__pick {
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    padding-top: 2px;
    min-width: 72px;
  }

  .ipm-card__pick-ring {
    width: 22px;
    height: 22px;
    border-radius: 999px;
    border: 1.5px solid #d0d7de;
    background: #fff;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: border-color 0.2s ease, background 0.2s ease, box-shadow 0.2s ease;
  }

  .ipm-card__pick-ring.is-on {
    border-color: var(--ipm-accent);
    background: var(--ipm-accent-soft);
    box-shadow: none;
  }

  .ipm-card__pick-check {
    font-size: 12px;
    color: var(--ipm-accent);
  }

  .ipm-card__pick-label {
    font-size: 11px;
    font-weight: 500;
    color: var(--ipm-muted);
    text-align: center;
    line-height: 1.3;
    max-width: 72px;
  }

  .ipm-card.is-selected .ipm-card__pick-label {
    color: #4e5969;
  }

  /* —— Footer —— */
  .ipm-footer {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    gap: 12px;
  }

  .ipm-footer__btn.ant-btn {
    min-width: 104px;
    height: 40px;
    padding: 0 22px;
    border-radius: 10px;
    font-weight: 600;
    font-size: 14px;
  }

  .ipm-footer__btn.ant-btn-default {
    border-color: rgba(15, 23, 42, 0.12);
    color: #475569;
    background: #fff;
  }

  .ipm-footer__btn--primary.ant-btn-primary {
    box-shadow: none;
  }
}

@media (max-width: 768px) {
  .institution-picker-modal {
    .ant-modal {
      max-width: calc(100vw - 20px);
      margin: 12px auto;
    }

    .ant-modal-header {
      padding: 20px 18px 16px;
    }

    .ipm-head__title {
      font-size: 19px;
      padding-right: 28px;
    }

    .ipm-head__sub {
      font-size: 13px;
    }

    .ant-modal-body {
      padding: 16px 16px 6px;
    }

    .ant-modal-footer {
      padding: 14px 16px 18px;
    }

    .ipm-intro {
      flex-direction: column;
      align-items: stretch;
    }

    .ipm-intro__stat {
      flex-direction: row;
      justify-content: space-between;
      align-items: center;
      min-width: unset;
    }

    .ipm-intro__stat-num {
      font-size: 24px;
    }

    .ipm-card__row {
      flex-wrap: wrap;
    }

    .ipm-card__pick {
      width: 100%;
      flex-direction: row;
      justify-content: flex-end;
      padding-top: 0;
      min-width: unset;
    }

    .ipm-card__pick-label {
      max-width: none;
      text-align: right;
    }

    .ipm-chip__value {
      max-width: none;
    }
  }
}

@media (max-width: 480px) {
  .institution-picker-modal {
    .ipm-footer {
      flex-direction: column-reverse;
      align-items: stretch;
    }

    .ipm-footer__btn.ant-btn {
      width: 100%;
    }

    .ant-modal-footer .ant-btn + .ant-btn {
      margin-inline-start: 0;
    }
  }
}
</style>
