/// <reference types="vitest" />
import { fileURLToPath } from 'node:url'
import * as process from 'node:process'
import { readFileSync } from 'node:fs'
import { loadEnv } from 'vite'
import type { ConfigEnv, PluginOption, UserConfig } from 'vite'
import { createVitePlugins } from './plugins'
import { OUTPUT_DIR } from './plugins/constants'
import { resolve } from 'node:path'

const baseSrc = fileURLToPath(new URL('./src', import.meta.url))

function inlineLoadingScript(): PluginOption {
  return {
    name: 'inline-loading-script',
    enforce: 'pre',
    transformIndexHtml: {
      order: 'pre',
      handler(html) {
        const loadingScript = readFileSync(resolve(process.cwd(), 'public/loading.js'), 'utf-8')
          .replaceAll('</script>', '<\\/script>')
        return html.replace(
          /<script\b(?=[^>]*\bsrc=["'](?:%BASE_URL%|\.\/|\/(?:[^"']*\/)?)loading\.js["'])[^>]*><\/script>/,
          `<script>${loadingScript}</script>`,
        )
      },
    },
  }
}

function normalizeCdnBase(value?: string) {
  const trimmed = value?.trim()
  if (!trimmed)
    return ''
  return trimmed.endsWith('/') ? trimmed : `${trimmed}/`
}

function withCdnVersion(url: string, version?: string) {
  const trimmed = version?.trim()
  if (!trimmed)
    return url
  return `${url}?v=${encodeURIComponent(trimmed)}`
}

// https://vitejs.dev/config/
export default ({ mode }: ConfigEnv): UserConfig => {
  const env = loadEnv(mode, process.cwd())
  const publicBase = env.VITE_APP_PUBLIC_PATH || (mode === 'production' ? '/platform/' : './')
  const cdnBase = normalizeCdnBase(env.VITE_CDN_BASE)
  const cdnVersion = env.VITE_CDN_VERSION
  const proxyObj = {}
  if (mode === 'development'|| mode === 'mylocal' || mode === 'chenlocal') {
    // 获取所有环境变量
    const envKeys = Object.keys(env)
    // 查找所有API和URL配对
    const apiKeys = envKeys.filter(key => key.includes('VITE_BASE'))
    // 循环添加代理配置
    // /api、/sso 直连本地 Go 服务时都需要保留完整路径，与服务注册的路由保持一致。
    const forwardFullPathPrefixes = ['/api', '/sso']
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
    base: publicBase,
    experimental: {
      renderBuiltUrl(filename) {
        if (cdnBase && filename.startsWith('assets/'))
          return withCdnVersion(`${cdnBase}${filename}`, cdnVersion)
      },
    },
    plugins: [
      ...createVitePlugins({ ...env, VITE_APP_PUBLIC_PATH: publicBase }),
      inlineLoadingScript(),
    ],
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
            vue: ['vue'],
            router: ['vue-router', 'pinia'],
            i18n: ['vue-i18n'],
            vueuse: ['@vueuse/core'],
            dayjs: ['dayjs'],
            // lodash: ['loadsh-es'],
          },
        },
      },
    },
    server: {
      port: 6688,
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
