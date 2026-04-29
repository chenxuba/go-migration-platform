# PEP-3 自动计分模块说明

代码位置：`pkg/pep3score`

## 模块职责

这个包只做计算，不依赖数据库和接口层：

- 按题目得分汇总各分量表原始分。
- 按儿童实足年龄选择常模年龄组。
- 查发展年龄、百分比级数、标准分。
- 计算沟通、体能、适应不良行为三个合成分数。
- 输出发展/适应程度。

## 输入数据

需要三份基础资料：

- 题库：`docs/pep3-item-bank-simplified-draft.json`
- 分量表映射：`docs/pep3-score-domain-map.json`
- 常模换算：`docs/pep3-norm-conversion-ocr-draft.json`
- 人工校对修正：`docs/pep3-norm-manual-corrections.json`

示例：

```go
items, _ := pep3score.LoadItemDefinitionsFile("docs/pep3-item-bank-simplified-draft.json")
domains, _ := pep3score.LoadDomainDefinitionsFile("docs/pep3-score-domain-map.json")
norms, _ := pep3score.LoadMergedNormRecordsFiles(
    "docs/pep3-norm-conversion-ocr-draft.json",
    "docs/pep3-norm-manual-corrections.json",
)

engine, _ := pep3score.NewEngine(items, domains, norms)
```

## 按题目分数计分

```go
result, err := engine.Score(pep3score.AssessmentInput{
    BirthDate:      time.Date(2020, 1, 10, 0, 0, 0, 0, time.UTC),
    AssessmentDate: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC),
    ItemScores: map[int]int{
        1: 2,
        2: 1,
        3: 0,
        // ...
    },
})
```

## 按原始分计分

适合导入旧记录或校验手册样例：

```go
result, err := engine.Score(pep3score.AssessmentInput{
    BirthDate:      time.Date(2000, 10, 29, 0, 0, 0, 0, time.UTC),
    AssessmentDate: time.Date(2004, 2, 10, 0, 0, 0, 0, time.UTC),
    RawScores: map[string]int{
        "CVP": 16,
        "EL":  18,
        "RL":  12,
        "FM":  34,
        "GM":  27,
        "VMI": 11,
        "AE":  3,
        "SR":  6,
        "CMB": 7,
        "CVB": 10,
    },
})
```

## 输出结构

`AssessmentResult` 包含：

- `Age`：按手册口径计算的实足年龄，不四舍五入。
- `Scales`：各分量表结果，包括原始分、发展年龄、百分位、标准分、程度。
- `Composites`：沟通、体能、适应不良行为合成结果。
- `Warnings`：数据缺失或换算表未匹配时的警告。

## 当前风险

常模表来自扫描 PDF OCR，当前只能用于开发联调。`pep3-norm-manual-corrections.json` 用于覆盖已经人工校对的格子；正式测评软件上线前，必须完成全量校对，再把校对结果固化为 verified 版本。
