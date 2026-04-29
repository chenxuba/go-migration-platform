# PEP-3 常模换算表索引（OCR草稿）

- 来源文件：`/Users/chenrui/Desktop/教学教研/教研/pep3/PEP-3常模(2025).pdf`
- 输出数据：`pep3-norm-conversion-ocr-draft.csv/json`
- 处理方式：扫描页 OCR + 表格坐标归位；附表4为人工转录草稿。
- 注意：常模数据会直接影响测评结论，正式上线前必须逐页人工校对。

## 表结构

| 附表 | 用途 | 页码 | 软件用途 |
|---|---|---|---|
| 附表1 | 发展及行为副测验原始分与相应发展年龄 | 2-3 | 原始分换发展年龄/发展月龄 |
| 附表2.1-2.10 | 各年龄组副测验原始分与百分比级数 | 5-24 | 原始分 + 实足年龄组 -> 百分位 |
| 附表3.1-3.10 | 各年龄组副测验原始分与标准分数 | 26-45 | 原始分 + 实足年龄组 -> 标准分，用于合成分数 |
| 附表4.1 | 合成分数标准分数总和与百分比级数 | 47-48 | 标准分总和 -> 合成百分位 |

## 年龄组

| 表号 | 年龄组 | 月龄范围 | 附表2页码 | 附表3页码 |
|---|---|---:|---|---|
| 2.1/3.1 | 2岁0个月-2岁5个月 | 24-29 | 5-6 | 26-27 |
| 2.2/3.2 | 2岁6个月-2岁11个月 | 30-35 | 7-8 | 28-29 |
| 2.3/3.3 | 3岁0个月-3岁5个月 | 36-41 | 9-10 | 30-31 |
| 2.4/3.4 | 3岁6个月-3岁11个月 | 42-47 | 11-12 | 32-33 |
| 2.5/3.5 | 4岁0个月-4岁5个月 | 48-53 | 13-14 | 34-35 |
| 2.6/3.6 | 4岁6个月-4岁11个月 | 54-59 | 15-16 | 36-37 |
| 2.7/3.7 | 5岁0个月-5岁5个月 | 60-65 | 17-18 | 38-39 |
| 2.8/3.8 | 5岁6个月-5岁11个月 | 66-71 | 19-20 | 40-41 |
| 2.9/3.9 | 6岁0个月-6岁5个月 | 72-77 | 21-22 | 42-43 |
| 2.10/3.10 | 6岁6个月-7岁5个月 | 78-89 | 23-24 | 44-45 |

## 记录数量

- `development_age_raw_score_range`：183 条
- `raw_score_to_percentile_rank`：1613 条
- `raw_score_to_scaled_score`：1621 条
- `composite_standard_score_sum_to_percentile_rank`：141 条

## OCR 诊断

| 页码 | 行定位 | 行锚点数 | 识别到的单元格数 |
|---:|---|---:|---:|
| 5 | raw_score_fit | 29 | 301 |
| 6 | geometry_fallback | 0 | 21 |
| 7 | geometry_fallback | 0 | 95 |
| 8 | raw_score_fit | 33 | 56 |
| 9 | geometry_fallback | 0 | 65 |
| 10 | raw_score_fit | 31 | 22 |
| 11 | raw_score_fit | 30 | 269 |
| 12 | geometry_fallback | 0 | 23 |
| 13 | raw_score_fit | 31 | 257 |
| 14 | raw_score_fit | 33 | 56 |
| 15 | geometry_fallback | 0 | 55 |
| 16 | raw_score_fit | 32 | 39 |
| 17 | geometry_fallback | 0 | 78 |
| 18 | geometry_fallback | 0 | 17 |
| 19 | geometry_fallback | 0 | 92 |
| 20 | raw_score_fit | 33 | 23 |
| 21 | geometry_fallback | 0 | 16 |
| 22 | raw_score_fit | 33 | 46 |
| 23 | geometry_fallback | 0 | 64 |
| 24 | geometry_fallback | 0 | 18 |
| 26 | raw_score_fit | 34 | 266 |
| 27 | raw_score_fit | 33 | 56 |
| 28 | geometry_fallback | 0 | 87 |
| 29 | geometry_fallback | 0 | 16 |
| 30 | raw_score_fit | 9 | 105 |
| 31 | geometry_fallback | 0 | 20 |
| 32 | geometry_fallback | 0 | 86 |
| 33 | raw_score_fit | 8 | 18 |
| 34 | raw_score_fit | 28 | 208 |
| 35 | raw_score_fit | 33 | 56 |
| 36 | raw_score_fit | 34 | 206 |
| 37 | geometry_fallback | 0 | 19 |
| 38 | geometry_fallback | 0 | 65 |
| 39 | raw_score_fit | 32 | 37 |
| 40 | raw_score_fit | 27 | 175 |
| 41 | raw_score_fit | 13 | 26 |
| 42 | geometry_fallback | 0 | 35 |
| 43 | geometry_fallback | 0 | 30 |
| 44 | raw_score_fit | 11 | 77 |
| 45 | geometry_fallback | 0 | 33 |
