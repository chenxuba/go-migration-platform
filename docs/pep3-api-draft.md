# PEP-3 评分接口草稿

## 接口

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

报告接口会读取已保存记录的评分结果，返回适合前端预览、打印页或后续导出使用的结构化报告：

- `basicInfo`：儿童、测试员、出生日期、测试日期、实足年龄、常模月龄
- `developmentRows`：发展量表分量表结果
- `behaviorRows`：适应不良行为分量表结果
- `caregiverReportRows`：照顾者报告分量表结果
- `compositeRows`：沟通、体能、适应不良行为合成结果
- `summary`：按合成分生成的简要文字摘要
- `warnings`：缺失换算值、OCR 草稿等需要复核的提示

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

当前建表入口是教育服务启动时的 `EnsureInfrastructureTables`，表名为 `assessment_record`。后续如果要支持多量表，建议沿用这张通用记录表，再按量表编码调用各自 scorer。
