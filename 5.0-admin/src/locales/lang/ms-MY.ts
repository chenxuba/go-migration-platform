import msMY from 'ant-design-vue/es/locale/ms_MY'

const msMYModules = import.meta.glob([
  '~/locales/lang/**/ms-MY.ts',
  '~/pages/**/locales/ms-MY.ts',
], {
  eager: true,
})

const messages = {}

for (const item in msMYModules) {
  const locale = (msMYModules[item] as any)?.default
  if (locale)
    Object.assign(messages, locale)
}

export default {
  ...messages,
  antd: msMY,
}
