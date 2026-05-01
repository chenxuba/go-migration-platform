# PEP3 题库管理设计图

本文基于当前 PEP3 静态题库结构设计题库管理页面与数据关系。现有数据来源包括：

- `docs/pep3-item-bank-simplified-draft.json`：172 道 PEP3 测试题。
- `docs/pep3-score-domain-map.json`：13 个分量表/领域定义。
- `pkg/pep3template/record_fields.go`：逐题儿童表现记录字段。
- `assessment_scale_*` 表：题库、领域、记录字段、常模的持久化结构。

## 1. 管理端页面结构图

```mermaid
flowchart TB
  A["量表管理 / PEP-3"] --> B["题库管理弹窗/页面"]

  B --> C["顶部概览"]
  C --> C1["量表编码: PEP3"]
  C --> C2["版本: 2025-92题版"]
  C --> C3["题目数: 172"]
  C --> C4["领域数: 13"]
  C --> C5["数据状态 / 来源"]

  B --> D["筛选区"]
  D --> D1["关键词: 题号/项目/材料/方法"]
  D --> D2["分量表: CVP/EL/RL/FM..."]
  D --> D3["题目类型: 发展/行为/照顾者报告"]
  D --> D4["校对状态: OCR草稿/已校对/已发布"]
  D --> D5["来源页范围"]

  B --> E["左侧领域导航"]
  E --> E1["沟通: CVP/EL/RL"]
  E --> E2["运动: FM/GM/VMI"]
  E --> E3["适应不良行为: AE/SR/CMB/CVB"]
  E --> E4["照顾者报告: PB/PSC/AB"]

  B --> F["题目列表"]
  F --> F1["题号"]
  F --> F2["分量表"]
  F --> F3["测试项目"]
  F --> F4["材料"]
  F --> F5["评分选项"]
  F --> F6["记录字段数"]
  F --> F7["来源页"]
  F --> F8["状态"]

  F --> G["题目详情抽屉"]
  G --> G1["基础信息"]
  G --> G2["施测方法"]
  G --> G3["评分标准"]
  G --> G4["2/1/0 评分选项"]
  G --> G5["儿童表现记录字段"]
  G --> G6["来源 PDF / 页码"]
  G --> G7["变更记录"]
```

## 2. 核心数据关系图

```mermaid
erDiagram
  assessment_scale_dataset ||--o{ assessment_scale_domain : "scale_code + scale_version"
  assessment_scale_dataset ||--o{ assessment_scale_item : "scale_code + scale_version"
  assessment_scale_dataset ||--o{ assessment_scale_item_record_field : "scale_code + scale_version"
  assessment_scale_dataset ||--o{ assessment_scale_norm_record : "scale_code + scale_version"
  assessment_scale_domain ||--o{ assessment_scale_item : "domain_code in item_json"
  assessment_scale_item ||--o{ assessment_scale_item_record_field : "item_no"

  assessment_scale_dataset {
    bigint id PK
    varchar scale_code "PEP3"
    varchar scale_version "2025-92题版"
    varchar data_status
    longtext sources_json
    tinyint del_flag
  }

  assessment_scale_domain {
    bigint id PK
    varchar scale_code
    varchar scale_version
    varchar domain_code "CVP/EL/RL/FM..."
    int sort_no
    longtext domain_json
    tinyint del_flag
  }

  assessment_scale_item {
    bigint id PK
    varchar scale_code
    varchar scale_version
    int item_no "1-172"
    longtext item_json
    tinyint del_flag
  }

  assessment_scale_item_record_field {
    bigint id PK
    varchar scale_code
    varchar scale_version
    int item_no
    varchar field_key
    int sort_no
    longtext field_json
    tinyint del_flag
  }

  assessment_scale_norm_record {
    bigint id PK
    varchar scale_code
    varchar scale_version
    varchar record_key
    int sort_no
    longtext norm_json
    tinyint del_flag
  }
```

## 3. 题目 JSON 结构

题目管理的主编辑对象是 `assessment_scale_item.item_json`，当前 PEP3 题目字段如下：

```mermaid
classDiagram
  class ScaleQuestionBankItem {
    int itemNo
    string itemTitle
    string testItem
    string materials
    string method
    string domainCode
    string domainName
    string standard
    ScoreOptionList scoreOptions
    string scoreOptionText
    RecordFieldList recordFields
    string sourcePdf
    IntList sourcePages
    string ocrStatus
    string updatedAt
  }

  class ScoreOption {
    int value
    string label
    string description
  }

  class RecordField {
    string key
    string label
    string fieldType
    string displayType
    bool required
    string placeholder
    FieldOptionList options
  }

  class FieldOption {
    string value
    string label
  }

  ScaleQuestionBankItem "1" --> "0..3" ScoreOption
  ScaleQuestionBankItem "1" --> "0..n" RecordField
  RecordField "1" --> "0..n" FieldOption
```

## 4. 数据流设计图

```mermaid
sequenceDiagram
  participant Seed as JSON/模板种子
  participant Edu as education-service
  participant DB as assessment_scale_* 表
  participant API as 管理端题库 API
  participant Admin as platform-admin 题库管理
  participant Assess as 机构端测评录入

  Seed->>Edu: 启动加载 PEP3 静态数据
  Edu->>DB: ReplaceAssessmentScaleStaticData()
  DB-->>Edu: 题目/领域/记录字段/常模已同步

  Admin->>API: 查询题库版本与题目列表
  API->>DB: 读取 dataset/domain/item/record_field
  DB-->>API: 返回结构化题库
  API-->>Admin: 展示筛选、列表、详情、编辑

  Admin->>API: 保存题目或记录字段变更
  API->>DB: 按 scale_code + scale_version + item_no 更新
  DB-->>API: 保存成功

  Assess->>Edu: GET /api/v1/assessments/pep3/form-template
  Edu->>DB: 读取已发布版本
  Edu-->>Assess: 返回 172 题录入模板
```

## 5. 建议的题库管理功能分区

| 区域 | 功能 | 对应数据 |
|---|---|---|
| 版本概览 | 显示 `PEP3`、`2025-92题版`、题目数、领域数、来源文件、数据状态 | `assessment_scale_dataset` |
| 领域导航 | 按 13 个分量表聚合题目，显示题数和满分 | `assessment_scale_domain.domain_json` |
| 题目列表 | 查看题号、测试项目、材料、评分选项、来源页、校对状态 | `assessment_scale_item.item_json` |
| 题目详情 | 查看/编辑材料、施测方法、评分标准、来源页 | `ScaleQuestionBankItem` |
| 评分选项 | 管理 2/1/0 或 NA 说明 | `scoreOptions` / `scoreOptionText` |
| 记录字段 | 管理文本、数字、单选、多选、优势手/眼等扩展记录 | `assessment_scale_item_record_field.field_json` |
| 常模数据 | 只读查看或单独维护常模换算表 | `assessment_scale_norm_record.norm_json` |
| 发布控制 | OCR 草稿、已校对、已发布、停用 | 建议新增版本状态或复用 `data_status` |

## 6. 编辑与发布状态

```mermaid
stateDiagram-v2
  [*] --> OCRDraft: OCR导入
  OCRDraft --> Reviewing: 人工校对
  Reviewing --> Reviewing: 保存草稿
  Reviewing --> Ready: 校对完成
  Ready --> Published: 发布版本
  Published --> Archived: 新版本替换
  Reviewing --> OCRDraft: 退回重整
  Ready --> Reviewing: 继续修改
```

## 7. PEP3 分量表分组

| 组合 | 分量表 | 题数 |
|---|---|---:|
| 沟通 | CVP 认知（语言/语前）、EL 语言表达、RL 语言理解 | 77 |
| 运动 | FM 小肌肉、GM 大肌肉、VMI 模仿（视觉/动作） | 46 |
| 适应不良行为 | AE 情感表达、SR 社交互动、CMB 行为特征-非语言、CVB 行为特征-语言 | 49 |
| 照顾者报告 | PB 问题行为、PSC 个人自理、AB 适应行为 | 38 |

说明：前 10 个分量表由 172 道现场测评题映射；PB/PSC/AB 是照顾者报告原始分领域，不直接映射现场题号。

## 8. 落地建议

1. 先把平台端题库抽屉从 `pep3-static-question-bank.ts` 的 P1/P3 静态样例，改为调用后端完整题库接口。
2. 后端新增平台管理接口：列表、详情、保存题目、保存记录字段、版本发布。
3. 编辑时保留 `item_json` 的完整 JSON 结构，同时在 API 层展开为前端表单字段，避免前端直接编辑大段 JSON。
4. 对评分标准、施测方法、常模数据建议加“发布版本”控制，机构端测评只读取已发布版本。
