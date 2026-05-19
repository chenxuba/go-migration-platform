# VB-MAPP 数据包

本目录保存 VB-MAPP 模块首版开发用的结构化数据。数据由 `scripts/extract_vbmapp_data.py` 从本机资料抽取生成。

## 文件

| 文件 | 内容 |
|---|---|
| `domains.json` | 16 个能力领域、排序、阶段覆盖和最大分 |
| `milestone-items.json` | 170 条 Milestones 里程碑题项 |
| `milestone-scoring-rules.json` | 170 条里程碑 `1分 / 0.5分` 判定规则 |
| `barriers.json` | 24 项 Barriers 障碍评估，0-4 分选项 |
| `transition.json` | 18 项 Transition 转衔评估，1-5 分选项和安置建议 |
| `extraction-summary.json` | 抽取数量、领域分布和校验结果 |

## 当前版本

```text
VBMAPP_CN_2ND_DRAFT_2026_05
```

这是开发草稿版本。正式上线前需要做两件事：

- 逐项人工核对题项、计分规则、障碍项和转衔项。
- 确认资料数字化和商用授权边界。

## 重新生成

```bash
/Users/chenrui/.cache/codex-runtimes/codex-primary-runtime/dependencies/python/bin/python3 scripts/extract_vbmapp_data.py
```

抽取脚本依赖本机 `/Users/chenrui/Downloads` 下的 VB-MAPP 资料文件。若资料文件名变化，需要同步更新脚本中的路径。
