import antfu from '@antfu/eslint-config'

export default antfu({
  ignores: [
    'types/auto-imports.d.ts',
    'types/components.d.ts',
    'public',
    'tsconfig.*.json',
    'tsconfig.json',
  ],
}, {
  rules: {
    // Vue 3 / Ant Design Vue：允许 v-model:open 等带参数的 v-model
    'vue/no-v-model-argument': 'off',
    'no-console': 0,
    'style/quote-props': 0,
    'unused-imports/no-unused-vars': 0,
    'ts/no-unused-expressions': 0,
    'style/func-style': 0,
  },
})
