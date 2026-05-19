# VB-MAPP Pad 端测评页面设计

## 结论

VB-MAPP 不应该 170 道题各做一套完全不同页面。正确做法是：

```text
统一测评工作台骨架
+ 模块导航
+ 领域/阶段导航
+ 通用题项展示
+ 按题型切换录入组件
```

也就是说，页面整体保持一致，差异放在“题型组件”和“材料提示”里。老师在同一个工作台里连续完成里程碑、障碍、转衔、任务分析等模块，不是进入几套互不相干的页面。

前面生成的 3 张图只能理解为同一个工作台的 3 个模块状态：

- Milestones 状态：中央区域显示 `0 / 0.5 / 1` 评分。
- Barriers 状态：中央区域切换为 `0-4` 障碍程度评分。
- Transition 状态：中央区域切换为 `1-5` 转衔评分和安置建议。

实际产品必须能互相跳转：左侧模块导航负责跨模块跳转，底部上一题/下一题负责按全局测评顺序连续推进。

## 页面骨架

建议 Pad 横屏优先，适配 11 寸和 12.9 寸。

```text
┌────────────────────────────────────────────────────────────────────────────┐
│ 顶栏：VB-MAPP测评工作台 / 儿童 / 年龄 / 日期 / 评估师 / 自动保存 / 提交       │
├──────────────┬──────────────────────────────────────────────┬──────────────┤
│ 左侧导航      │ 中央题项工作区                                │ 右侧辅助栏     │
│              │                                              │              │
│ Milestones   │ 阶段 + 领域 + 题号                              │ 总分/进度      │
│ Level 1      │ 题目描述                                      │ 当前领域得分   │
│ Level 2      │ 评分标准                                      │ 未完成项       │
│ Level 3      │ 材料与操作提示                                  │ 高风险障碍     │
│ Barriers     │ 分数选择 / 观察记录 / 备注                       │ 历史对比       │
│ Transition   │                                              │ IEP候选目标    │
│ Task Skills  │                                              │              │
├──────────────┴──────────────────────────────────────────────┴──────────────┤
│ 底栏：上一题 / 题号进度 / 下一题 / 跳转 / 自动下一题 / 保存草稿               │
└────────────────────────────────────────────────────────────────────────────┘
```

## 暖色视觉基准

VB-MAPP 页面应沿用当前 Pad 端已有的暖色调，不采用冷灰蓝 SaaS 风格。现有 Pad 端可复用的视觉基准：

- 页面底色：`#FFF7EE`
- 主文字：`#432B22`
- 正文文字：`#7F665A`
- 弱文字：`#BBA99C`
- 主操作橙色：`#E96F43`
- 深橙强调：`#C95735`
- 分割线：`#F0DACB`
- 软分割线：`#F6E7DC`

新的 VB-MAPP 页面应和现有 PEP-3 Pad 页面保持同一套温暖、清晰、低干扰的测评工具感。

## 锁定现有页头页脚

VB-MAPP 的顶部和底部必须继承现有 Pad 测评页，不单独设计一套壳。这里的“保持一致”只限定 Header 和 Footer；中间主体区域按 VB-MAPP 重新设计。

### 顶部 Header

沿用现有结构：

```text
返回按钮
量表测评工作台标题
儿童 / 年龄 / 日期 / 施测者
已恢复最新草稿
保存草稿
提交记录
```

VB-MAPP 只替换标题文本，例如：

- `VB-MAPP里程碑评估 测评工作台`
- `VB-MAPP障碍评估 测评工作台`
- `VB-MAPP转衔评估 测评工作台`

顶部按钮、间距、字号、圆角、状态提示和橙色主按钮都跟现有页面保持一致。

### 底部 Footer

沿用现有固定底栏：

```text
上一题 | 当前题号 / 总题数 | 下一题 | 跳到缺题 | 自动下一题
```

不要再使用单独的 `评估记录 / 保存草稿 / 下一项` 底栏样式。保存和提交放回顶部，底部只负责连续测评导航。

### 页面主体

中间主体不照搬现有量表页面，按 VB-MAPP 模块重新设计：

- 里程碑：`0 / 0.5 / 1`。
- 障碍：`0 / 1 / 2 / 3 / 4`。
- 转衔：`1 / 2 / 3 / 4 / 5`。

主体建议保持 Pad 测评页适合横屏操作的三栏工作区，但内容组织由 VB-MAPP 决定：左侧题项/模块导航，中间评分与证据录入，右侧进度、历史对比、备注、IEP 候选目标。

## 统一跳转模型

这 3 类页面不是孤立页面，而是同一个 `assessment_draft` 的不同工作区状态。

```text
全局测评序列
Milestones Level 1
-> Milestones Level 2
-> Milestones Level 3
-> Barriers
-> Transition
-> Task Analysis
-> 生成报告 / IEP
```

### 左侧模块导航

左侧导航负责“跳模块”和“跳领域”：

- 点 `里程碑`：进入 Milestones，展开 Level 1/2/3 和 16 个领域。
- 点 `障碍评估`：进入 Barriers，中央区域切换为 24 个障碍项。
- 点 `转衔评估`：进入 Transition，中央区域切换为 18 个转衔项。
- 点 `任务分析`：进入低分里程碑对应的技能拆解列表。

跳转前自动保存当前题项。模块旁显示完成数、低分提醒、是否有未保存改动。

### 底部上一题/下一题

底部按钮负责“按测评顺序连续走”：

- 当前题不是模块最后一题：`下一题` 进入同模块下一题。
- 当前题是模块最后一题：`下一题` 自动进入下一个模块第一题。
- 当前题是模块第一题：`上一题` 回到上一个模块最后一题。
- 用户开启“自动下一题”时，选择分数后延迟 300-500ms 推进，避免误触。

因此从 Milestones 可以自然进入 Barriers，再进入 Transition；老师不需要退出页面重新打开另一个测评。

### 顶部和右侧连续状态

顶部、右侧栏在所有模块保持一致：

- 顶部显示儿童、年龄、日期、评估师、草稿保存状态。
- 右侧显示总进度、当前模块进度、历史同题得分、低分项和 IEP 候选目标。
- 切换模块时右侧只更新统计内容，不改变整体布局。

### 历史评估对比

历史对比不是单独页面，先作为右侧辅助栏能力：

- 当前题显示上次得分和本次得分差异。
- 领域维度显示上次总分、本次总分、变化趋势。
- 报告页再生成完整对比图表。

## 顶栏

顶栏不要做得太重，测评时老师最需要稳定看到：

- 量表名：`VB-MAPP 语言行为里程碑评估及安置计划`
- 儿童姓名、年龄、测评日期、评估师
- 当前状态：草稿、已保存、未保存、提交中
- 操作：返回、保存草稿、提交正式记录

注意：不要压缩标题字体；如果空间不够，把低优先级状态放到第二行或右侧，不要把量表名和儿童信息缩小到难读。

## 左侧导航

左侧导航建议分两层：

### 模块层

- 里程碑
- 障碍评估
- 转衔评估
- 任务分析
- 生活自理，可选

### 里程碑内部

里程碑模块按阶段和领域组织：

- Level 1：0-18 个月
- Level 2：18-30 个月
- Level 3：30-48 个月

每个领域显示：

- 领域名
- 已完成题数
- 当前领域得分
- 是否存在 0 分或 0.5 分

## 中央题项工作区

中央区域是核心，不同题目共用同一结构：

1. 题项定位
   - `Level 1 / 提要求 / 1-M`
   - 评估方式标签：`T`、`O`、`E`、`TO`

2. 题目描述
   - 展示题项主文本。
   - 如果题目较长，允许换行，不要省略。

3. 评分区
   - Milestones：`0分 / 0.5分 / 1分`
   - Barriers：`0 / 1 / 2 / 3 / 4`
   - Transition：`1 / 2 / 3 / 4 / 5`
   - Task Analysis：`未开始 / 辅助中 / 部分独立 / 独立 / 泛化`

4. 评分标准
   - 左侧显示 `1分标准`
   - 右侧显示 `0.5分标准`
   - 未达标时默认可选 `0分`

5. 材料与操作提示
   - 来自材料提示手册。
   - 显示建议材料、环境、测试方式。
   - 对 `TO` 题显示计时器。

6. 观察记录
   - 老师记录儿童表现。
   - 支持常用语快捷短语。
   - 支持语音转文字可作为后续增强。

## 右侧辅助栏

右侧辅助栏不要占据主操作，只做辅助决策：

- 当前总分：Milestones、Barriers、Transition
- 当前领域进度
- 低分项提醒
- 高障碍项提醒
- 上次测评同题得分
- 可能的 IEP 候选目标
- 材料准备清单

右侧栏可以做成可折叠，老师在连续打分时能扩大中央区域。

## 底栏

底栏保持固定，不随内容跳动：

- 上一题
- 当前题号：`23 / 170`
- 下一题
- 跳题
- 自动下一题开关
- 保存草稿

如果选择分数后开启“自动下一题”，系统可以延迟 300-500ms 跳转，避免误触。

## 题型组件

VB-MAPP 题目不是每题一套页面，而是按题型复用组件。

### 1. 标准里程碑题

适合大多数题项。

组件：

- 题目描述
- `0 / 0.5 / 1` 分段按钮
- `1分 / 0.5分` 判定规则
- 备注输入

### 2. 计时观察题

适合 `TO:30分钟`、`TO:60分钟` 等题项。

组件：

- 计时器
- 观察次数计数器
- 事件记录按钮，例如“发生一次”
- 最终分数选择

### 3. 材料测试题

适合命名、听者反应、配对、LRFFC 等。

组件：

- 材料清单
- 试次记录
- 正确/错误/辅助
- 自动统计达到多少个
- 最终分数选择

### 4. 多试次/数量阈值题

适合“达到 5 次”“20 个物品”“10 个不同要求”等。

组件：

- 目标数量
- 已达成数量
- 快捷加减
- 根据阈值提示建议得分

注意：建议得分只是辅助，最终仍由老师确认。

### 5. Barriers 障碍题

组件：

- 0-4 分纵向选项
- 每个分值一句判定说明
- 行为例子备注
- 高分时提示“需进入 IEP 风险考虑”

### 6. Transition 转衔题

组件：

- 1-5 分选项
- 可由 Milestones / Barriers 自动预填的项目显示“系统建议”
- 安置建议文本
- 老师确认/调整

### 7. Task Analysis 技能追踪题

组件：

- 当前里程碑下的细分技能列表
- 每个技能一个状态
- 可批量标记
- 低分里程碑自动展开

## 典型页面状态

### Milestones 题项页

```text
Level 1 / 提要求 / 1-M                         T/O/E/TO 标签

发出2个话语、手语或图片交换沟通系统……

得分：
[ 0分 ] [ 0.5分 ] [ 1分 ]

1分标准：2个
0.5分标准：1个

材料提示：
准备儿童喜欢但不能直接获得的强化物……

观察记录：
[ 输入儿童表现、辅助情况、例子 ]
```

### Barriers 题项页

```text
障碍评估 / B01 负面行为

请选择最符合当前儿童表现的分值：

0 未见明显障碍
1 偶发轻微行为问题
2 每天出现小的负面行为
3 更严重且更频繁
4 每天多次且可能造成危险

案例备注：
[ 输入触发条件、频率、处理方式 ]
```

### Transition 题项页

```text
转衔评估 / T01 VB-MAPP里程碑评估总分

系统建议：3分
依据：Milestones 总分 76

请选择确认分值：
1 0-25分
2 26-50分
3 51-100分
4 101-135分
5 136-170分

安置建议：
开始受益于集体教学、自然环境……
```

## 是否需要图像模型

可以用图像模型生成视觉稿，但不要按 170 道题分别生成页面。建议只生成两类图：

1. 一个“统一工作台总览图”，重点表达模块之间如何跳转。
2. 少量模块状态图，例如 Milestones、Barriers、Transition。

图像模型只做视觉方向稿。实际 Flutter 页面应使用结构化组件和真实接口数据渲染，不依赖图中的文字细节。

已生成但偏冷的方向稿，暂不作为主方向：

- `output/imagegen/vbmapp-pad-milestones.png`
- `output/imagegen/vbmapp-pad-barriers.png`
- `output/imagegen/vbmapp-pad-transition.png`

已生成的暖色方向稿，更接近当前 Pad 端视觉：

- `output/imagegen/vbmapp-locked-shell-custom-body/vbmapp-locked-shell-milestones.png`
- `output/imagegen/vbmapp-locked-shell-custom-body/vbmapp-locked-shell-barriers.png`
- `output/imagegen/vbmapp-locked-shell-custom-body/vbmapp-locked-shell-transition.png`
- `output/imagegen/vbmapp-existing-shell-set/vbmapp-existing-shell-milestones.png`
- `output/imagegen/vbmapp-existing-shell-set/vbmapp-existing-shell-barriers.png`
- `output/imagegen/vbmapp-existing-shell-set/vbmapp-existing-shell-transition.png`
- `output/imagegen/vbmapp-final-style-set/vbmapp-final-style-milestones.png`
- `output/imagegen/vbmapp-final-style-set/vbmapp-final-style-barriers.png`
- `output/imagegen/vbmapp-final-style-set/vbmapp-final-style-transition.png`
- `output/imagegen/vbmapp-pad-connected-workbench-warm.png`
- `output/imagegen/vbmapp-pad-milestones-warm.png`
- `output/imagegen/vbmapp-pad-barriers-warm.png`
- `output/imagegen/vbmapp-pad-transition-warm.png`
- `output/imagegen/vbmapp-redesign-batch/vbmapp-pad-redesign-light-workbench.png`
- `output/imagegen/vbmapp-redesign-batch/vbmapp-pad-redesign-clinical-efficient.png`

其中 `vbmapp-locked-shell-custom-body/` 这一组三张图作为当前主方向：顶部和底部严格继承现有 Pad 测评页，中间主体按 VB-MAPP 的里程碑、障碍、转衔重新设计。`vbmapp-pad-connected-workbench-warm.png` 只保留“模块之间是连续流程”的结构参考，不作为视觉主方向。

## 给图像模型的提示词草案

```text
Use case: ui-mockup
Asset type: iPad landscape assessment app screen
Primary request: Design a polished VB-MAPP assessment workbench for therapists.
Style/medium: modern clinical education SaaS UI, calm, dense, professional.
Composition/framing: 12.9-inch iPad landscape screen, top header, left navigation, central assessment item workspace, right progress rail, fixed footer.
Color palette: warm Pad app palette: #FFF7EE page background, #432B22 ink, #7F665A body text, #E96F43 primary orange, #C95735 deep orange, #F0DACB lines.
Text: Chinese UI labels, including "VB-MAPP测评工作台", "里程碑", "障碍评估", "转衔评估", "任务分析", "Level 1 / 提要求 / 1-M", "0分", "0.5分", "1分", "上一题", "下一题", "自动保存".
Constraints: no marketing hero, no decorative illustration, no nested cards, no oversized text, no text overlap, practical admin tool layout.
Avoid: flashy gradients, cartoon style, stock photo background, cluttered dashboard, tiny unreadable text.
```

## 开发建议

Flutter 文件结构建议沿用现有 Pad 测评拆分方式：

- `vbmapp_assessment_page.dart`
- `vbmapp_assessment_chrome.dart`
- `vbmapp_assessment_navigation.dart`
- `vbmapp_assessment_workspace.dart`
- `vbmapp_assessment_right_rail.dart`
- `vbmapp_assessment_footer.dart`
- `vbmapp_assessment_state_actions.dart`
- `vbmapp_assessment_loading.dart`
- `vbmapp_assessment_support.dart`

首版先实现 Milestones 页面，再扩展 Barriers 和 Transition。不要一开始把任务分析、生活自理、IEP 全塞进同一个工作区。
