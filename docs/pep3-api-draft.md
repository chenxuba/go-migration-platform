# PEP-3 评分接口草稿

机构端前端已提供 TypeScript 类型和调用封装：`institution-admin/src/api/edu-center/pep3-assessment.ts`。

## 接口

获取测评录入表模板：

`GET /api/v1/assessments/pep3/form-template`

这个接口给前端录入页面使用，返回 172 道测试题、每题材料/方法/评分标准、拆分后的 `2/1/0` 选项、分量表映射和原始分录入字段。前端不需要把题库或评分标准写死在页面里。

该接口需要登录态 token，因为响应包含完整题库和评分标准。

核心字段：

- `templateCode`：录入表模板编码，目前为 `PEP3_ASSESSMENT_FORM`
- `itemCount`：题目数量，当前为 `172`
- `basicFields`：保存测评记录所需的基础字段
- `domains`：13 个分量表定义，包括发展、行为和照顾者报告分量表
- `rawScoreFields`：原始分字段；发展/行为副测验可由逐题分自动汇总，照顾者报告当前按原始分录入
- `itemGroups`：按测试员记录册页面分组的题目列表
- `submitContract`：前端提交到计分接口或保存接口时使用的字段约定

题目片段示例：

```json
{
  "templateCode": "PEP3_ASSESSMENT_FORM",
  "itemCount": 172,
  "scoreOptions": [
    { "value": 2, "label": "2分", "description": "通过 / 恰当" },
    { "value": 1, "label": "1分", "description": "部分通过 / 轻微" },
    { "value": 0, "label": "0分", "description": "未能通过 / 严重" }
  ],
  "itemGroups": [
    {
      "groupCode": "booklet_page_2",
      "title": "记录册第2页题目（1-14）",
      "items": [
        {
          "itemNo": 1,
          "itemTitle": "（1） 旋开瓶盖",
          "domainCode": "FM",
          "materials": "肥皂泡液",
          "method": "测试员把瓶放在桌子上...",
          "standard": "2- 能自行旋开瓶盖\n1- ...\n0- ...",
          "scoreOptions": [
            { "value": 2, "label": "2分", "description": "能自行旋开瓶盖" }
          ]
        }
      ]
    }
  ]
}
```

计分：

`POST /api/v1/assessments/pep3/score`

这个评分接口只负责评分计算，不保存测评记录，不生成报告。需要落库时使用下方“保存测评记录”接口。

## 请求

可以传每题得分，也可以直接传分量表原始分。日期使用 `YYYY-MM-DD`。

```json
{
  "birthDate": "2000-10-29",
  "assessmentDate": "2004-02-10",
  "rawScores": {
    "CVP": 16,
    "EL": 18,
    "RL": 12,
    "FM": 34,
    "GM": 27,
    "VMI": 11,
    "AE": 3,
    "SR": 6,
    "CMB": 7,
    "CVB": 10,
    "PB": 7,
    "PSC": 7,
    "AB": 10
  }
}
```

每题得分写法：

```json
{
  "birthDate": "2020-01-01",
  "assessmentDate": "2024-01-01",
  "allowMissingItems": true,
  "itemScoreList": [
    { "itemNo": 1, "score": 2 },
    { "itemNo": 2, "score": 1 }
  ]
}
```

## 响应

响应会返回：

- `scaleCode`：量表编码，目前为 `PEP3`
- `scaleVersion`：当前数据版本，目前为 `2025-draft`
- `dataStatus`：数据状态说明
- `sources`：本次评分加载的数据文件
- `result`：年龄、分量表原始分、发展年龄、百分比级数、标准分和合成分结果

## 数据来源

当前接口从本地整理文件加载：

- `docs/pep3-item-bank-simplified-draft.json`
- `docs/pep3-score-domain-map.json`
- `docs/pep3-norm-conversion-ocr-draft.json`
- `docs/pep3-norm-manual-corrections.json`

可以通过环境变量 `PEP3_DATA_DIR` 指定这些文件所在目录。常模文件仍是 OCR 草稿，正式使用前要生成全量人工核验版。

## 测评草稿

草稿用于前端边测边保存，和正式测评记录分表存储。草稿不会生成评分结果，也不会进入解释性报告或记录册；完成录入后，前端仍然调用 `records/create` 生成正式记录。

保存草稿：

`POST /api/v1/assessments/pep3/drafts/save`

请求可以只提交部分题目，`birthDate` 和 `assessmentDate` 在草稿阶段也可以为空。再次保存时带上 `id` 即可覆盖同一草稿。

```json
{
  "id": 1,
  "studentId": 1001,
  "studentName": "李东尼",
  "birthDate": "2020-01-01",
  "assessmentDate": "2024-01-01",
  "allowMissingItems": true,
  "itemScoreList": [
    { "itemNo": 1, "score": 2 },
    { "itemNo": 2, "score": 1 }
  ],
  "rawScoreList": [
    { "scaleCode": "PB", "rawScore": 7 }
  ],
  "remark": "测到第2页"
}
```

响应会返回完整草稿详情：

- `input`：前端下次继续测评时可直接回填的输入快照
- `progress.itemCount`：题库题目数，当前为 `172`
- `progress.answeredItemCount`：已保存逐题得分数量
- `progress.missingItemNos`：尚未保存逐题得分的题号
- `progress.domainProgress`：按分量表统计的完成情况
- `progress.canScore`：当前草稿是否已有足够数据调用计分接口
- `status`：`draft`、`ready_to_score` 或 `complete`

查询草稿详情：

`GET /api/v1/assessments/pep3/drafts/detail?id=1`

分页查询草稿：

`POST /api/v1/assessments/pep3/drafts/page`

```json
{
  "pageRequestModel": {
    "pageIndex": 1,
    "pageSize": 20
  },
  "queryModel": {
    "studentId": 1001,
    "searchKey": "李东尼",
    "status": "draft"
  }
}
```

提交草稿生成正式测评记录：

`POST /api/v1/assessments/pep3/drafts/submit`

```json
{
  "id": 1
}
```

提交接口会读取草稿中的 `input`，完成计分并写入 `assessment_record`，然后把草稿标记为 `submitted`，响应中返回正式记录：

```json
{
  "draftId": 1,
  "recordId": 10,
  "draftStatus": "submitted",
  "record": {
    "id": 10,
    "assessmentCode": "PEP3",
    "assessmentName": "PEP-3儿童心理教育评核"
  }
}
```

删除草稿：

`POST /api/v1/assessments/pep3/drafts/delete`

```json
{
  "id": 1
}
```

## 保存测评记录

`POST /api/v1/assessments/pep3/records/create`

这个接口会完成评分并保存测评记录。记录会绑定：

- 机构 ID
- 学员信息
- 测试员信息
- PEP-3 量表编码与数据版本
- 出生日期、测试日期和计算年龄
- 本次输入 JSON
- 本次评分结果 JSON

请求示例：

```json
{
  "studentId": 1001,
  "studentName": "李东尼",
  "birthDate": "2000-10-29",
  "assessmentDate": "2004-02-10",
  "remark": "首次评估",
  "rawScores": {
    "CVP": 16,
    "EL": 18,
    "RL": 12,
    "FM": 34,
    "GM": 27,
    "VMI": 11,
    "AE": 3,
    "SR": 6,
    "CMB": 7,
    "CVB": 10,
    "PB": 7,
    "PSC": 7,
    "AB": 10
  }
}
```

查询详情：

`GET /api/v1/assessments/pep3/records/detail?id=1`

生成报告数据：

`GET /api/v1/assessments/pep3/records/report?id=1`

报告接口会读取已保存记录的评分结果，返回适合前端预览、打印页或后续导出使用的结构化报告。

为了后续前端可以直接填充页面，响应只暴露模板结构和必要元数据，不再保留旧的顶层行数据字段：

- `templateCode`：报告模板编码，目前为 `PEP3_EXPLANATORY_REPORT`
- `templateVersion`：模板/量表数据版本
- `title`：报告标题
- `sections`：前端直接渲染用的区块数组，每个区块包含 `sectionCode`、`title`、`type`、`fields`、`table` 或 `textItems`

旧版 `basicInfo`、`developmentRows`、`behaviorRows`、`caregiverReportRows`、`compositeRows`、`summary`、`warnings` 顶层字段已删除，对应内容都放在 `sections` 中。

`sections` 示例：

```json
{
  "templateCode": "PEP3_EXPLANATORY_REPORT",
  "title": "PEP-3解释性报告",
  "sections": [
    {
      "sectionCode": "basic_info",
      "title": "基本资料",
      "type": "field_grid",
      "fields": [
        { "key": "studentName", "label": "儿童姓名", "value": "李东尼", "rawValue": "李东尼" },
        { "key": "ageText", "label": "实足年龄", "value": "3岁3个月11天", "rawValue": "3岁3个月11天" }
      ]
    },
    {
      "sectionCode": "development_scores",
      "title": "发展量表",
      "type": "score_table",
      "table": {
        "columns": [
          { "key": "scaleCode", "label": "编码" },
          { "key": "scaleName", "label": "副测验" },
          { "key": "rawScore", "label": "原始分" }
        ],
        "rows": [
          { "scaleCode": "CVP", "scaleName": "认知（语言/语前）", "rawScore": 16 }
        ]
      }
    }
  ]
}
```

生成测试员记录册页面数据：

`GET /api/v1/assessments/pep3/records/booklet?id=1`

这个接口专门对应 `测试员记录册彩(1).pdf`。返回的是 14 个扫描 PDF 页面的页面数据，前端后续可以按 `pages[].sections[]` 直接填充记录册样式：

- `templateCode`：记录册模板编码，目前为 `PEP3_RECORD_BOOKLET`
- `sourcePdf`：对应的记录册源文件
- `pages`：PDF 页数组，`sourcePdfPageNo` 与扫描 PDF 页码一致
- `sections[].layout`：区块在扫描页上的版面区域，常见为 `left`、`right`、`full`
- `sections[].type`：区块类型，例如 `field_grid`、`score_summary_table`、`composite_score_table`、`item_grid`、`page_tally`、`domain_score_table`
- `item_grid.table.rows`：逐题记录行，包含 `itemNo`、`itemTitle`、`domainCode`、`score` 以及 `CVP/EL/RL/FM/GM/VMI/AE/SR/CMB/CVB` 对应得分格
- `page_tally.table.rows`：每页按 2/1/0 与原始分小计生成的分领域汇总
- `domain_score_table.table.rows`：按领域归集的原始分计算表

如果保存记录时只提交了 `rawScores`，没有提交 `itemScores` 或 `itemScoreList`，记录册仍会返回完整题目行和总分换算，但逐题格会为空，并在 `warnings` 中提示“该记录未保存逐题得分”。

记录册片段示例：

```json
{
  "templateCode": "PEP3_RECORD_BOOKLET",
  "title": "PEP-3测试员记录册",
  "sourcePdf": "测试员记录册彩(1).pdf",
  "pages": [
    {
      "pageNo": 2,
      "sourcePdfPageNo": 2,
      "pageType": "spread",
      "sections": [
        {
          "sectionCode": "page_2_item_grid",
          "title": "第2页 儿童表现记录",
          "type": "item_grid",
          "layout": "full",
          "table": {
            "rows": [
              {
                "itemNo": 1,
                "itemTitle": "（1） 旋开瓶盖",
                "domainCode": "FM",
                "score": 2,
                "FM": 2
              }
            ]
          }
        }
      ]
    }
  ]
}
```

分页查询：

`POST /api/v1/assessments/pep3/records/page`

```json
{
  "pageRequestModel": {
    "pageIndex": 1,
    "pageSize": 20
  },
  "queryModel": {
    "studentId": 1001,
    "searchKey": "李东尼",
    "assessmentDateBegin": "2026-01-01",
    "assessmentDateEnd": "2026-12-31"
  }
}
```

当前建表入口是教育服务启动时的 `EnsureInfrastructureTables`，正式记录表为 `assessment_record`，测评草稿表为 `assessment_draft`。后续如果要支持多量表，建议沿用这两张通用表，再按量表编码调用各自 scorer。
