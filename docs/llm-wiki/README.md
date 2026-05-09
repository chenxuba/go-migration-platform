# LLM Wiki

本目录是项目内嵌的 LLM 可读知识库，用来沉淀架构、业务规则、接口约定、迁移决策和常见排障路径。

设计参考 Andrej Karpathy 的 LLM wiki 思路：把原始资料和面向 LLM 的整理稿分开，让代理先读稳定索引，再按需追溯证据。

参考来源：

- https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f

## 目录结构

```text
docs/llm-wiki/
  README.md                 # 本说明
  AGENTS.md                 # Codex 维护规则
  sources.md                # 外部资料和项目资料登记
  wiki/
    index.md                # LLM 优先阅读入口
    architecture.md         # 项目架构地图
    business-domains.md     # 业务域地图
    operating-playbooks.md  # 常见任务操作手册
  raw/
    README.md               # 原始资料投放规则
  schemas/
    wiki-entry.schema.json  # wiki 条目元数据 schema
  notes/
    changelog.md            # wiki 维护日志
```

## 使用方式

给 Codex 或其它 LLM 上下文时，优先提供：

1. `docs/llm-wiki/wiki/index.md`
2. 与当前任务相关的 wiki 页面
3. 必要时再提供 `docs/llm-wiki/sources.md` 或 `docs/llm-wiki/raw/` 中的原始资料

不要默认把 `raw/` 全量塞进上下文。原始资料用于追溯和更新，不是默认阅读入口。

## 更新规则

- 新增稳定知识时，先更新 `wiki/`，再在 `sources.md` 登记证据来源。
- 新增外部资料时，放入 `raw/` 或登记 URL，避免直接把大段外部文本复制到 wiki 正文。
- 业务规则必须带来源路径、接口、页面或提交背景。
- 不确定的推断要标注为 `待验证`，不要写成事实。
- 大改后更新 `notes/changelog.md`。
