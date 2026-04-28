import type { PluginOption, ResolvedConfig } from 'vite'

interface PreloadAllOptions {
  includeJs?: boolean
  includeCss?: boolean
}

const DEFAULT_OPTIONS: Required<PreloadAllOptions> = {
  includeJs: true,
  includeCss: true,
}

function withBase(base: string, fileName: string) {
  const cleanFileName = fileName.replace(/^\/+/, '')
  if (!base || base === './')
    return cleanFileName
  return `${base.replace(/\/?$/, '/')}${cleanFileName}`
}

function collectExistingPaths(html: string) {
  const paths = new Set<string>()
  const attrRegExp = /\s(?:href|src)=["']([^"']+)["']/g
  let match: RegExpExecArray | null

  while ((match = attrRegExp.exec(html)) !== null)
    paths.add(match[1])

  return paths
}

export default function preloadAll(options: PreloadAllOptions = {}): PluginOption {
  const mergedOptions = { ...DEFAULT_OPTIONS, ...options }
  let viteConfig: ResolvedConfig

  return {
    name: 'local:preload-all',
    enforce: 'post',
    apply: 'build',
    configResolved(config) {
      viteConfig = config
    },
    transformIndexHtml: {
      order: 'post',
      handler(html, ctx) {
        if (!ctx.bundle)
          return html

        const existingPaths = collectExistingPaths(html)
        const tags = Object.values(ctx.bundle)
          .map((bundle) => {
            const href = withBase(viteConfig.base ?? '', bundle.fileName)
            if (existingPaths.has(href))
              return null

            if (mergedOptions.includeJs && bundle.type === 'chunk' && /-[^/]+\.js$/.test(bundle.fileName)) {
              return {
                tag: 'link',
                attrs: {
                  rel: 'modulepreload',
                  href,
                },
                injectTo: 'head' as const,
              }
            }

            if (mergedOptions.includeCss && bundle.type === 'asset' && /-[^/]+\.css$/.test(bundle.fileName)) {
              return {
                tag: 'link',
                attrs: {
                  rel: 'stylesheet',
                  href,
                },
                injectTo: 'head' as const,
              }
            }

            return null
          })
          .filter((tag): tag is NonNullable<typeof tag> => Boolean(tag))
          .sort((a, b) => String(a.attrs.href).localeCompare(String(b.attrs.href)))

        return {
          html,
          tags,
        }
      },
    },
  }
}
