# VB-MAPP 产品与开发方案

## 目标

把 VB-MAPP 做成平台内的完整测评与干预计划模块，优先服务机构老师的日常评估、复评、报告和 IEP 制定。

首版目标：

- 支持里程碑评估录入、草稿保存、正式提交和自动算分。
- 支持障碍评估、转衔评估录入和解释。
- 支持 PDF / Word 报告导出。
- 支持历史评估对比。
- 基于低分项、高障碍项、任务分析和课程逻辑生成 IEP 初稿。
- 支持家庭干预建议和生活自理能力作为扩展内容。

## 资料分工

当前资料已经可以覆盖从测评录入到 IEP 初稿的大部分链路：

| 资料 | 主要用途 |
|---|---|
| 三阶段 Word 题项文档 | 170 条 Milestones 题库、领域、阶段、题项描述 |
| 里程碑评估材料提示手册 | 评估材料、操作提示、测试方式、素材建议 |
| 里程碑评估 PDF | 0 / 0.5 / 1 计分阈值 |
| 障碍评估 PDF | 24 项 Barriers 0-4 计分说明 |
| 转衔评估 PDF | 18 项 Transition 1-5 计分、安置建议 |
| 上册 Guide | 评分解释、阶段解读、IEP 和安置逻辑复核 |
| 下册 Protocol | 录入表、汇总图、任务分析表和报告版式参考 |
| 任务分析 `.doc` | 细分技能、前置能力、技能追踪 |
| 课程逻辑 PDF | 里程碑到课程目标、IEP 目标的映射 |
| 300 名词表 | 命名、听者反应、词汇素材库 |
| 家庭干预教学指南 | 家庭训练目标和家长建议 |
| 生活自理检核表 | 扩展生活自理评估和 IEP 补充目标 |

这些资料涉及版权和机构来源。开发时可以先做内部授权版本；如果要商业化销售，需要先明确数字化展示、复制、改写和报告生成的授权边界。

## 产品模块

### 1. 测评工作台

面向老师或评估师，主流程为：

1. 选择学员，读取生日、性别、历史测评。
2. 创建 VB-MAPP 测评草稿。
3. 按模块录入：里程碑、障碍、转衔、任务分析、生活自理。
4. 自动保存草稿，显示完成进度。
5. 提交正式测评，生成结构化结果。
6. 进入报告、历史对比和 IEP 生成。

工作台建议布局：

- 顶部：学员、年龄、评估日期、评估师、保存状态、提交按钮。
- 左侧：模块导航，按 Level、领域、Barriers、Transition 分组。
- 中间：题项录入表，支持分数、备注、观察记录、材料提示。
- 右侧：实时得分概览、未完成项、高风险障碍、IEP 候选目标。

### 2. Milestones 里程碑评估

录入结构：

- 3 个阶段。
- 16 个左右能力领域，按 VB-MAPP 结构展示。
- 170 个里程碑题项。
- 每题支持 `0`、`0.5`、`1`，并保存评估方式、备注和证据说明。

自动计算：

- 每个领域小计。
- 每个 Level 小计。
- 总分，满分 170。
- 通过项、半通过项、未通过项。
- 优势领域、优先干预领域。

### 3. Barriers 障碍评估

录入结构：

- 24 类障碍。
- 每项 0-4 分，分数越高表示障碍越明显。
- 支持备注和案例记录。

自动分析：

- 障碍总分。
- 高风险障碍清单。
- 与里程碑低分项联动，例如提要求弱 + 负面行为高时，提高功能沟通目标优先级。
- 在报告中生成干预注意事项。

### 4. Transition 转衔评估

录入结构：

- 18 项转衔指标。
- 每项 1-5 分。
- 部分指标可由 Milestones / Barriers 自动预填，老师确认后保存。

自动分析：

- 转衔总分。
- 安置建议：强化教学、部分融合、较少限制环境等。
- 支持报告中解释“为什么建议当前安置方式”。

### 5. Task Analysis 技能追踪

任务分析不是单纯算总分，而是服务训练计划：

- 每个里程碑下挂 `a/b/c/d/M` 等细分技能。
- 每个细分技能支持未开始、辅助中、部分独立、独立、泛化。
- 低分里程碑自动展开对应前置技能，形成训练目标候选。
- 支持复评对比时展示细分技能进步。

### 6. IEP 自动生成

IEP 生成不要只调用大模型直接写文本，应先做规则候选，再由 AI 整理成报告语言。

候选目标来源：

- Milestones 中 `0` 和 `0.5` 的题项。
- Barriers 中 3-4 分的障碍项。
- Transition 中低分项。
- Task Analysis 中未掌握的前置技能。
- 课程逻辑 PDF 中对应的课程目标。
- 家庭干预和生活自理检核表中的可训练目标。

生成策略：

- 优先选择影响沟通、配合、教学控制、社交参与的目标。
- 每个领域控制目标数量，避免 IEP 过满。
- 每个目标包含长期目标、短期目标、评估标准、训练形式、材料建议、家庭建议。
- AI 输出必须可编辑，老师确认后才进入正式 IEP。

### 7. 报告与导出

报告内容建议：

- 基本资料。
- 里程碑总览和各领域得分。
- Level 1/2/3 得分图。
- Barriers 障碍评估结果。
- Transition 转衔评估和安置建议。
- 历史评估对比。
- 优势与弱项摘要。
- IEP 目标建议。
- 家庭干预建议。
- 附录：任务分析、生活自理检核。

导出格式：

- PDF：用于机构正式出具、家长查看。
- Word：用于老师二次编辑。
- 可选 Excel：用于题项明细和历史数据分析。

## 技术落点

### 后端服务

沿用现有 `education-service` 的测评框架：

- 草稿：`assessment_draft`
- 正式记录：`assessment_record`
- 量表静态数据：`assessment_scale_dataset`、`assessment_scale_item`、`assessment_scale_domain`
- 题项记录字段：`assessment_scale_item_record_field`
- 报告解释缓存：`assessment_report_interpretation`
- IEP 计划：沿用现有 `assessment_iep_plan` 相关能力

新增建议：

- `pkg/vbmappscore`：VB-MAPP 计分引擎。
- `services/education/internal/service/vbmapp_static_data_service.go`
- `services/education/internal/service/vbmapp_assessment_service.go`
- `services/education/internal/service/vbmapp_draft_service.go`
- `services/education/internal/service/vbmapp_report_service.go`
- `services/education/internal/service/vbmapp_iep_plan_service.go`
- `services/education/internal/handler/vbmapp_assessment_handler.go`

### 静态数据文件

建议先整理成 JSON 数据包，人工校验后再种入数据库：

- `docs/vbmapp/milestone-items.json`
- `docs/vbmapp/milestone-scoring-rules.json`
- `docs/vbmapp/domains.json`
- `docs/vbmapp/barriers.json`
- `docs/vbmapp/transition.json`
- `docs/vbmapp/task-analysis.json`
- `docs/vbmapp/material-prompts.json`
- `docs/vbmapp/curriculum-map.json`
- `docs/vbmapp/home-intervention.json`
- `docs/vbmapp/self-care-checklist.json`

静态数据版本建议使用：

```text
VBMAPP_CN_2ND_DRAFT_2026_05
```

等人工核验和授权确认后，再切到正式版本。

### 输入快照

正式记录保存时，`input_json` 建议保留完整作答快照：

```json
{
  "studentId": 1001,
  "birthDate": "2020-01-01",
  "assessmentDate": "2026-05-19",
  "milestoneScores": [],
  "barrierScores": [],
  "transitionScores": [],
  "taskSkillChecks": [],
  "selfCareChecks": [],
  "remark": ""
}
```

### 评分结果

`result_json` 建议输出：

```json
{
  "scaleCode": "VBMAPP",
  "scaleVersion": "VBMAPP_CN_2ND_DRAFT_2026_05",
  "milestones": {
    "totalScore": 0,
    "maxScore": 170,
    "levelScores": [],
    "domainScores": [],
    "itemResults": []
  },
  "barriers": {
    "totalScore": 0,
    "highRiskItems": [],
    "itemResults": []
  },
  "transition": {
    "totalScore": 0,
    "placementLevel": "",
    "itemResults": []
  },
  "iepCandidates": [],
  "warnings": []
}
```

## API 草案

建议路径与 PEP-3 保持一致：

- `GET /api/v1/assessments/vbmapp/form-template`
- `GET /api/v1/assessments/vbmapp/form-template/summary`
- `GET /api/v1/assessments/vbmapp/form-template/item?itemNo=`
- `POST /api/v1/assessments/vbmapp/score`
- `POST /api/v1/assessments/vbmapp/drafts/save`
- `POST /api/v1/assessments/vbmapp/drafts/item/save`
- `GET /api/v1/assessments/vbmapp/drafts/detail?id=`
- `POST /api/v1/assessments/vbmapp/drafts/page`
- `POST /api/v1/assessments/vbmapp/drafts/submit`
- `POST /api/v1/assessments/vbmapp/records/create`
- `POST /api/v1/assessments/vbmapp/records/update`
- `GET /api/v1/assessments/vbmapp/records/detail?id=`
- `POST /api/v1/assessments/vbmapp/records/page`
- `GET /api/v1/assessments/vbmapp/records/report?id=`
- `GET /api/v1/assessments/vbmapp/records/history?studentId=`
- `POST /api/v1/assessments/vbmapp/iep/generate`
- `GET /api/v1/assessments/vbmapp/records/report/export-word?id=`
- `GET /api/v1/assessments/vbmapp/records/report/export-pdf?id=`

## 前端页面

建议新增：

- `institution-admin/src/api/edu-center/vbmapp-assessment.ts`
- `institution-admin/src/pages/teacherCenter/vbmapp-assessment-workbench.vue`
- `institution-admin/src/components/assessment/vbmapp/VBMAPPScoreGrid.vue`
- `institution-admin/src/components/assessment/vbmapp/VBMAPPDomainNavigator.vue`
- `institution-admin/src/components/assessment/vbmapp/VBMAPPReportPreview.vue`
- `institution-admin/src/components/assessment/vbmapp/VBMAPPHistoryCompare.vue`
- `institution-admin/src/components/assessment/vbmapp/VBMAPPIEPCandidates.vue`

设计风格应保持后台工具属性，重点是密度、可扫描性和快速录入，不做宣传页。

## 开发阶段

### 第 0 阶段：合规与数据整理

- 明确授权边界：内部使用、机构使用、商业销售分别确认。
- 将资料抽取为 JSON。
- 建立人工校验表，题项、分值、领域、阶段必须逐项核对。

### 第 1 阶段：Milestones MVP

- 170 题题库。
- 0 / 0.5 / 1 录入和自动算分。
- 草稿保存、正式提交。
- 基础报告和历史对比。

### 第 2 阶段：Barriers + Transition

- 24 项障碍评估。
- 18 项转衔评估。
- 安置建议。
- 报告扩展。

### 第 3 阶段：Task Analysis + IEP

- 细分技能追踪。
- 课程逻辑映射。
- IEP 候选目标。
- AI 生成 IEP 初稿和老师编辑确认。

### 第 4 阶段：家庭干预与生活自理

- 家长版建议。
- 生活自理检核。
- Word/PDF 报告优化。
- 家长端查看或分享。

## 验收标准

首版可验收条件：

- 同一份人工评分样例，系统 Milestones 总分、领域分、Level 分一致。
- Barriers 和 Transition 总分及解释一致。
- 草稿可中断恢复，提交后不可误覆盖历史记录。
- PDF/Word 包含完整基本信息、分数表、图表和建议。
- 历史对比至少支持最近两次和任意两次对比。
- IEP 初稿能追溯到来源题项、障碍项或任务分析项。

## 主要风险

- 授权风险：资料可读不等于可商用数字化。
- 数据质量风险：扫描资料可能有 OCR 错误，必须人工校验。
- 专业风险：IEP 自动生成只能作为初稿，必须由老师确认。
- 版本风险：后续若切换正式授权数据包，需要保留旧记录的版本快照。
- 前端复杂度风险：VB-MAPP 模块多，首版必须先做 Milestones 主链路，不要同时铺开所有扩展模块。
