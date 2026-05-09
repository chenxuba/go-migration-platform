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

function noCacheOptimizedDeps(): PluginOption {
  return {
    name: 'no-cache-optimized-deps',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        if (req.url?.startsWith('/node_modules/.vite/deps/')) {
          const setHeader = res.setHeader.bind(res)
          res.setHeader = (name, value) => {
            if (String(name).toLowerCase() === 'cache-control')
              return setHeader(name, 'no-store')
            return setHeader(name, value)
          }
          res.setHeader('Cache-Control', 'no-store')
        }
        next()
      })
    },
  }
}

// https://vitejs.dev/config/
export default ({ mode }: ConfigEnv): UserConfig => {
  const env = loadEnv(mode, process.cwd())
  const publicBase = env.VITE_APP_PUBLIC_PATH || (mode === 'production' ? '/institution/' : './')
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
    base: publicBase,
    plugins: [
      inlineLoadingScript(),
      noCacheOptimizedDeps(),
      ...createVitePlugins({ ...env, VITE_APP_PUBLIC_PATH: publicBase }),
    ],
    optimizeDeps: {
      include: ['dayjs/locale/zh-cn'],
    },
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
      strictPort: true,
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
