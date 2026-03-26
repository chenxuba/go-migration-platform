/// <reference types="vitest" />
import { fileURLToPath } from 'node:url'
import * as process from 'node:process'
import { loadEnv } from 'vite'
import type { ConfigEnv, UserConfig } from 'vite'
import { createVitePlugins } from './plugins'
import { OUTPUT_DIR } from './plugins/constants'
import { resolve } from 'node:path'

const baseSrc = fileURLToPath(new URL('./src', import.meta.url))
// https://vitejs.dev/config/
export default ({ mode }: ConfigEnv): UserConfig => {
  const env = loadEnv(mode, process.cwd())
  const proxyObj = {}
  if (mode === 'development'|| mode === 'mylocal' || mode === 'chenlocal') {
    // 获取所有环境变量
    const envKeys = Object.keys(env)
    // 查找所有API和URL配对
    const apiKeys = envKeys.filter(key => key.includes('VITE_BASE'))
    // 循环添加代理配置
    // /api 指向 education 时保留完整路径，与 Go 注册的 /api/v1/... 一致；/sso 等仍去掉前缀。
    const forwardFullPathPrefixes = ['/api']
    apiKeys.forEach((apiKey) => {
      const apiValue = env[apiKey]
      // 构造对应的URL键名
      const urlKey = apiKey.replace('BASE', 'URL')
      const urlValue = env[urlKey]
      if (apiValue && urlValue) {
        const keepPath = forwardFullPathPrefixes.some(prefix => apiValue === prefix)
        proxyObj[apiValue] = {
          target: urlValue,
          changeOrigin: true,
          rewrite: keepPath ? (path => path) : (path => path.replace(new RegExp(`^${apiValue}`), '')),
        }
      }
    })
  }
  return {
    base: './',
    plugins: createVitePlugins(env),
    resolve: {
      alias: [
        {
          find: 'dayjs',
          replacement: 'dayjs/esm',
        },
        {
          find: /^dayjs\/locale/,
          replacement: 'dayjs/esm/locale',
        },
        {
          find: /^dayjs\/plugin/,
          replacement: 'dayjs/esm/plugin',
        },
        {
          find: 'vue-i18n',
          replacement: mode === 'development' ? 'vue-i18n/dist/vue-i18n.esm-browser.js' : 'vue-i18n/dist/vue-i18n.esm-bundler.js',
        },
        {
          find: /^ant-design-vue\/es$/,
          replacement: 'ant-design-vue/es',
        },
        {
          find: /^ant-design-vue\/dist$/,
          replacement: 'ant-design-vue/dist',
        },
        {
          find: /^ant-design-vue\/lib$/,
          replacement: 'ant-design-vue/es',
        },
        {
          find: /^ant-design-vue$/,
          replacement: 'ant-design-vue/es',
        },
        {
          find: 'lodash',
          replacement: 'lodash-es',
        },
        {
          find: '~@',
          replacement: baseSrc,
        },
        {
          find: '~',
          replacement: baseSrc,
        },
        {
          find: '@',
          replacement: baseSrc,
        },
        {
          find: '~#',
          replacement: resolve(baseSrc, './enums'),
        },
      ],
    },
    build: {
      chunkSizeWarningLimit: 4096,
      outDir: OUTPUT_DIR,
      rollupOptions: {
        output: {
          manualChunks: {
            vue: ['vue', 'vue-router', 'pinia', 'vue-i18n', '@vueuse/core'],
            antd: ['ant-design-vue', '@ant-design/icons-vue', 'dayjs'],
            // lodash: ['loadsh-es'],
          },
        },
      },
    },
    server: {
      port: 6678,
      proxy: {
        ...proxyObj,
        // [env.VITE_APP_BASE_API]: {
        //   target: env.VITE_APP_BASE_URL,
        // //   如果你是https接口，需要配置这个参数
        // //   secure: false,
        //   changeOrigin: true,
        //   rewrite: path => path.replace(new RegExp(`^${env.VITE_APP_BASE_API}`), ''),
        // },
      },
    },
    test: {
      globals: true,
      environment: 'jsdom',
    },
  }
}
