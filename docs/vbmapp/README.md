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
| `milestone-response-schemas.json` | 170 条里程碑逐题证据记录、素材策略和智能完成规则 |
| `barrier-response-schemas.json` | 24 条障碍评估行为证据记录规则 |
| `transition-response-schemas.json` | 18 条转衔评估系统建议、人工确认和安置证据规则 |
| `response-field-templates.json` | Pad 端可复用的证据字段/控件模板 |
| `response-material-profiles.json` | VB-MAPP 评估素材准备策略 |
| `response-schema-summary.json` | 证据记录 schema 统计 |
| `smart-assessment-design.md` | 智能版 VB-MAPP 测评产品设计说明 |
| `upper-lower-ocr-analysis.md` | 上册/下册 OCR 分析和来源核对记录 |
| `ocr-source-audit.md` | 上册/下册 OCR 与当前题库的差异审计 |
| `milestone-source-corrections.json` | 已识别的安全修正和来源冲突记录 |

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

逐题证据记录 schema 可用以下命令重新生成：

```bash
python3 scripts/generate_vbmapp_response_schemas.py
```

OCR 来源审计可用以下命令复跑；只会在当前字段为空且来源一致时应用安全补全：

```bash
python3 scripts/audit_vbmapp_ocr_sources.py
python3 scripts/audit_vbmapp_ocr_sources.py --apply
```

## 后端接口

- `GET /api/v1/assessments/vbmapp/schema`：返回 170 个里程碑、24 个障碍、18 个转衔项目及逐题 response schema、字段模板和素材策略。
- `POST /api/v1/assessments/vbmapp/drafts/item/save`：除 `score` 外，支持保存 `evidence`、`suggestedScore`、`teacherConfirmed`、`overrideReason`、`recordStatus`，用于 Pad 端把每题证据写入草稿 JSON。

## 上册/下册 OCR

上册、下册 PDF 是扫描图像型 PDF，普通文本抽取会接近空文本。可用 macOS Vision OCR 脚本识别：

```bash
swiftc scripts/ocr_pdf_vision.swift -o /tmp/ocr_pdf_vision
/tmp/ocr_pdf_vision "/Users/chenrui/Downloads/vb-mapp 语言行为里程碑评估及安置计划 上册 指南 第2版.pdf" 1 242 2.0 > /tmp/vbmapp_ocr/upper_guide_ocr.txt
/tmp/ocr_pdf_vision "/Users/chenrui/Downloads/vbmapp下册概况（第二版）.pdf" 1 69 2.0 > /tmp/vbmapp_ocr/lower_overview_ocr.txt
```

完整 OCR 文本不提交到仓库；项目只保存由 OCR 推导出的结构化数据、差异核对和产品设计结论。
