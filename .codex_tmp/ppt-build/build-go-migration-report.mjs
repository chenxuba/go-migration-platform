const { Presentation, PresentationFile } = await import("@oai/artifact-tool");
const fs = await import("node:fs/promises");
const path = await import("node:path");

const ROOT = "/Users/chenrui/Desktop/go-migration-platform";
const WX_DIR = "/Users/chenrui/Library/Containers/com.tencent.xinWeChat/Data/Documents/xwechat_files/wxid_73xsjlqtiqlc21_6ae8/temp/RWTemp/2026-05/e99af8fffe2ab981263e54d73e61756b";
const WX_TEMP = "/Users/chenrui/Library/Containers/com.tencent.xinWeChat/Data/Documents/xwechat_files/wxid_73xsjlqtiqlc21_6ae8/temp";
const OUT_DIR = path.join(ROOT, "output/project-report-ppt");
const SCRATCH_DIR = path.join(ROOT, "tmp/slides/go-migration-report");
const OUT = path.join(OUT_DIR, "output.pptx");

await fs.mkdir(OUT_DIR, { recursive: true });
await fs.mkdir(SCRATCH_DIR, { recursive: true });

const W = 1280;
const H = 720;
const SLIDE_COUNT = 9;
const FONT = "PingFang SC";
const COLORS = {
  ink: "#172033",
  text: "#344054",
  muted: "#778396",
  line: "#E5EAF2",
  bg: "#FBF7F2",
  bg2: "#F4F7FB",
  panel: "#FFFFFF",
  blue: "#2563EB",
  blueSoft: "#EAF2FF",
  green: "#14A06F",
  greenSoft: "#E8F7EF",
  orange: "#F0783D",
  orangeSoft: "#FFF1E8",
  amber: "#F2A51A",
  amberSoft: "#FFF7DE",
  purple: "#7257CE",
  purpleSoft: "#F0ECFF",
  teal: "#0891B2",
  tealSoft: "#E5F8FB",
  red: "#E45E5E",
  redSoft: "#FFF0F0",
};

const images = {
  login: path.join(WX_DIR, "544489a88226d99f9eca05b6656b0ccb.jpg"),
  dashboard: path.join(WX_DIR, "3f8fecc89b76625e3eda6085b2d906f1.jpg"),
  timetable: path.join(WX_DIR, "b08abf8f30dbc756b026c63e33381ae9.jpg"),
  timetableFull: path.join(WX_DIR, "3870637bca37f513d8af8c3cb5e300c7.jpg"),
  scaleLibrary: path.join(WX_DIR, "32db44bf573bc67a5ed57e64a43b2a4e.jpg"),
  assessmentWorkbench: path.join(WX_DIR, "60bf8691642cfcabdce147ff18ca22fb.jpg"),
  reportList: path.join(WX_DIR, "44787044ea3af5af63417217fd50f7c1.jpg"),
  workbenchDetail: path.join(WX_DIR, "522506b9581fafdc295ff1183e7618ec.jpg"),
  loginConfig: path.join(WX_TEMP, "ScreenShot_2026-05-06_101553_735.png"),
  institutionLogin: path.join(WX_TEMP, "ScreenShot_2026-05-06_101633_602.png"),
  governmentAccount: path.join(WX_TEMP, "ScreenShot_2026-05-06_101647_769.png"),
  storageConfig: path.join(WX_TEMP, "ScreenShot_2026-05-06_101710_422.png"),
  iepPlan: path.join(WX_TEMP, "ScreenShot_2026-05-06_101316_203.png"),
  iepPlanEdit: path.join(WX_TEMP, "ScreenShot_2026-05-06_101338_086.png"),
};

for (const [name, imagePath] of Object.entries(images)) {
  await fs.access(imagePath).catch(() => {
    throw new Error(`Missing screenshot ${name}: ${imagePath}`);
  });
}

async function readImageBlob(imagePath) {
  const bytes = await fs.readFile(imagePath);
  return bytes.buffer.slice(bytes.byteOffset, bytes.byteOffset + bytes.byteLength);
}

async function saveBlobToFile(blob, filePath) {
  const bytes = Buffer.from(await blob.arrayBuffer());
  await fs.writeFile(filePath, bytes);
}

function shape(slide, {
  x, y, w, h, fill = COLORS.panel, line = COLORS.line, radius = true,
}) {
  return slide.shapes.add({
    geometry: radius ? "roundRect" : "rect",
    position: { left: x, top: y, width: w, height: h },
    fill,
    line: line === "none" ? { fill: "#FFFFFF00", width: 0 } : { fill: line, width: 1 },
    adjustmentList: radius ? [{ name: "adj", formula: "val 9000" }] : undefined,
  });
}

function text(slide, value, {
  x, y, w, h, size = 18, color = COLORS.text, bold = false,
  align = "left", valign = "top", fill = "#FFFFFF00", line = "none",
  inset = 0, radius = false, autoFit = "shrinkText",
}) {
  const s = slide.shapes.add({
    geometry: radius ? "roundRect" : "rect",
    position: { left: x, top: y, width: w, height: h },
    fill,
    line: line === "none" ? { fill: "#FFFFFF00", width: 0 } : { fill: line, width: 1 },
    adjustmentList: radius ? [{ name: "adj", formula: "val 9000" }] : undefined,
  });
  s.text = value;
  s.text.typeface = FONT;
  s.text.fontSize = size;
  s.text.color = color;
  s.text.bold = bold;
  s.text.alignment = align;
  s.text.verticalAlignment = valign;
  s.text.autoFit = autoFit;
  s.text.insets = typeof inset === "number"
    ? { left: inset, right: inset, top: inset, bottom: inset }
    : inset;
  return s;
}

function title(slide, main, sub) {
  text(slide, main, { x: 56, y: 36, w: 820, h: 42, size: 29, color: COLORS.ink, bold: true });
  if (sub) {
    text(slide, sub, { x: 58, y: 80, w: 880, h: 28, size: 14, color: COLORS.muted });
  }
  shape(slide, { x: 56, y: 114, w: 1168, h: 1.2, fill: COLORS.line, line: "none", radius: false });
}

function footer(slide, idx) {
  text(slide, `教育康复多端业务平台项目汇报 · ${idx}/${SLIDE_COUNT}`, {
    x: 56, y: 684, w: 360, h: 20, size: 11, color: COLORS.muted,
  });
}

function tag(slide, value, x, y, w, fill, color, options = {}) {
  return text(slide, value, {
    x, y, w, h: options.h || 32, size: options.size || 14, color, bold: true,
    align: "center", valign: "middle", fill, radius: true,
    inset: { left: 10, right: 10, top: 4, bottom: 4 },
  });
}

function bullet(slide, items, x, y, w, lineH = 38, size = 16) {
  items.forEach((item, idx) => {
    const yy = y + idx * lineH;
    shape(slide, { x, y: yy + 12, w: 8, h: 8, fill: item.color || COLORS.orange, line: "none", radius: true });
    text(slide, item.text || item, {
      x: x + 20, y: yy, w, h: lineH, size: item.size || size,
      color: item.muted ? COLORS.muted : COLORS.text, bold: !!item.bold,
      valign: "middle",
    });
  });
}

function statCard(slide, { x, y, w, h, title: cardTitle, body, color, soft }) {
  shape(slide, { x, y, w, h, fill: COLORS.panel, line: COLORS.line, radius: true });
  shape(slide, { x: x + 18, y: y + 18, w: 38, h: 38, fill: soft, line: "none", radius: true });
  shape(slide, { x: x + 31, y: y + 31, w: 12, h: 12, fill: color, line: "none", radius: true });
  text(slide, cardTitle, { x: x + 70, y: y + 17, w: w - 94, h: 27, size: 19, color: COLORS.ink, bold: true });
  text(slide, body, { x: x + 70, y: y + 49, w: w - 94, h: h - 60, size: 13.8, color: COLORS.text });
}

async function image(slide, imagePath, x, y, w, h, alt, fit = "contain") {
  const img = slide.images.add({ blob: await readImageBlob(imagePath), fit, alt });
  img.position = { left: x, top: y, width: w, height: h };
  return img;
}

async function imageCard(slide, {
  imagePath, x, y, w, h, caption, alt, fit = "contain", pad = 10,
}) {
  shape(slide, { x, y, w, h, fill: COLORS.panel, line: COLORS.line, radius: true });
  const capH = caption ? 28 : 0;
  await image(slide, imagePath, x + pad, y + pad, w - pad * 2, h - pad * 2 - capH, alt || caption || "项目截图", fit);
  if (caption) {
    text(slide, caption, {
      x: x + pad + 4, y: y + h - capH - 2, w: w - pad * 2 - 8, h: capH,
      size: 12.5, color: COLORS.muted, align: "center", valign: "middle",
    });
  }
}

function moduleCard(slide, { x, y, w, h, name, user, functions, value, color, soft }) {
  shape(slide, { x, y, w, h, fill: COLORS.panel, line: COLORS.line, radius: true });
  shape(slide, { x: x + 16, y: y + 16, w: 44, h: 44, fill: soft, line: "none", radius: true });
  text(slide, name.slice(0, 1), { x: x + 16, y: y + 18, w: 44, h: 40, size: 20, color, bold: true, align: "center", valign: "middle" });
  text(slide, name, { x: x + 70, y: y + 17, w: w - 86, h: 30, size: 19, color: COLORS.ink, bold: true });
  text(slide, `对象：${user}`, { x: x + 22, y: y + 76, w: w - 44, h: 30, size: 13.5, color: COLORS.muted });
  text(slide, "主要功能", { x: x + 22, y: y + 118, w: w - 44, h: 24, size: 14.5, color, bold: true });
  text(slide, functions, { x: x + 22, y: y + 145, w: w - 44, h: 84, size: 13.2, color: COLORS.text });
  text(slide, "解决痛点", { x: x + 22, y: y + 246, w: w - 44, h: 24, size: 14.5, color, bold: true });
  text(slide, value, { x: x + 22, y: y + 273, w: w - 44, h: 86, size: 13.2, color: COLORS.text });
}

const presentation = Presentation.create({
  slideSize: { width: W, height: H },
});

// Slide 1
{
  const slide = presentation.slides.add();
  slide.background.fill = COLORS.bg;
  shape(slide, { x: 0, y: 0, w: W, h: H, fill: COLORS.bg, line: "none", radius: false });
  shape(slide, { x: 730, y: 0, w: 550, h: H, fill: COLORS.orangeSoft, line: "none", radius: false });
  tag(slide, "项目汇报 · 业务价值版", 62, 62, 180, COLORS.greenSoft, COLORS.green, { h: 34, size: 14 });
  text(slide, "教育康复多端业务平台\n项目汇报", {
    x: 62, y: 128, w: 590, h: 116, size: 42, color: COLORS.ink, bold: true,
    inset: { left: 0, right: 0, top: 0, bottom: 0 },
  });
  text(slide, "这套系统把客户开通、机构运营、专业评估、家校协同和监管展示放到同一条业务链路里，让不同角色各用各的入口，但数据能统一沉淀。", {
    x: 64, y: 276, w: 575, h: 88, size: 19, color: COLORS.text,
  });
  const tags = [
    ["解决多系统分散", COLORS.blueSoft, COLORS.blue],
    ["提升机构运营效率", COLORS.orangeSoft, COLORS.orange],
    ["PEP-3 评估闭环", COLORS.purpleSoft, COLORS.purple],
    ["支持监管和家校协同", COLORS.greenSoft, COLORS.green],
  ];
  tags.forEach((t, i) => tag(slide, t[0], 64 + (i % 2) * 214, 402 + Math.floor(i / 2) * 54, 190, t[1], t[2], { h: 38, size: 15 }));
  text(slide, "汇报口径：少讲技术，多讲问题、功能、亮点和业务价值。", {
    x: 64, y: 608, w: 560, h: 28, size: 15, color: COLORS.muted,
  });
  await imageCard(slide, { imagePath: images.dashboard, x: 670, y: 72, w: 520, h: 292, caption: "机构业务首页", alt: "机构业务首页截图" });
  await imageCard(slide, { imagePath: images.loginConfig, x: 665, y: 386, w: 250, h: 174, caption: "总控配置", alt: "总控后台配置截图" });
  await imageCard(slide, { imagePath: images.assessmentWorkbench, x: 940, y: 386, w: 250, h: 174, caption: "PEP-3 评估工作台", alt: "评估工作台截图" });
}

// Slide 2
{
  const slide = presentation.slides.add();
  slide.background.fill = COLORS.bg;
  title(slide, "项目解决了哪些问题", "先用业务语言讲清楚：以前痛在哪里，现在项目把哪些事情变简单了。");
  const pains = [
    ["客户开通慢", "租户、机构、品牌、权限和登录页集中配置，减少重复搭建。", COLORS.orange, COLORS.orangeSoft],
    ["机构运营分散", "招生、课表、上课、收费、报表统一进后台，少在多张表里来回找。", COLORS.blue, COLORS.blueSoft],
    ["评估依赖纸笔", "Pad 现场录入评估过程，报告和计划能在线沉淀。", COLORS.purple, COLORS.purpleSoft],
    ["家长反馈零散", "课表、请假、康复记录和问卷回到系统，方便长期追踪。", COLORS.green, COLORS.greenSoft],
    ["监管统计滞后", "政府端和大屏看汇总数据，减少人工汇报和反复统计。", COLORS.teal, COLORS.tealSoft],
    ["客户差异难管理", "品牌、功能、账号、存储等做成配置，不靠临时改代码。", COLORS.amber, COLORS.amberSoft],
  ];
  pains.forEach((p, i) => {
    const x = 58 + (i % 3) * 398;
    const y = 152 + Math.floor(i / 3) * 176;
    statCard(slide, { x, y, w: 355, h: 132, title: p[0], body: p[1], color: p[2], soft: p[3] });
  });
  shape(slide, { x: 84, y: 548, w: 1112, h: 70, fill: COLORS.greenSoft, line: "none", radius: true });
  text(slide, "一句话：这个项目不是单个后台，而是在把“开客户、管机构、做评估、出结果、看监管”这条链路做成标准化平台。", {
    x: 118, y: 568, w: 1040, h: 30, size: 18, color: COLORS.green, bold: true, align: "center", valign: "middle",
  });
  footer(slide, 2);
}

// Slide 3
{
  const slide = presentation.slides.add();
  slide.background.fill = COLORS.bg;
  title(slide, "项目板块总览", "每个板块服务不同角色，但围绕同一条业务链路协同。");
  const cards = [
    {
      name: "总控后台",
      user: "平台运营 / 客户成功",
      functions: "开通租户、机构、账号、权限、登录页、品牌、版本和基础配置。",
      value: "把客户交付从人工搭建变成后台配置，后续复制更快。",
      color: COLORS.orange,
      soft: COLORS.orangeSoft,
    },
    {
      name: "机构后台",
      user: "校区 / 老师 / 财务",
      functions: "招生、学员、课程、排课、点名、订单、审批和数据看板。",
      value: "把日常经营放到一套后台，减少多系统、多表格管理。",
      color: COLORS.blue,
      soft: COLORS.blueSoft,
    },
    {
      name: "评估 Pad",
      user: "评估老师 / 测试员",
      functions: "选择量表、现场录分、进度保存、报告列表和评估记录。",
      value: "减少纸笔记录和二次录入，评估过程更标准。",
      color: COLORS.purple,
      soft: COLORS.purpleSoft,
    },
    {
      name: "家长端",
      user: "家长 / 照顾者",
      functions: "绑定学员、看课表、请假、查看记录、提交问卷。",
      value: "让家校沟通和家长反馈能沉淀到系统。",
      color: COLORS.green,
      soft: COLORS.greenSoft,
    },
    {
      name: "政府端",
      user: "监管部门 / 区域管理员",
      functions: "机构台账、监管账号、区域汇总、数据大屏和督导任务。",
      value: "把人工汇总变成线上可查、可展示、可追踪。",
      color: COLORS.teal,
      soft: COLORS.tealSoft,
    },
  ];
  cards.forEach((c, i) => moduleCard(slide, { ...c, x: 50 + i * 238, y: 150, w: 216, h: 390 }));
  shape(slide, { x: 122, y: 576, w: 1036, h: 44, fill: COLORS.orangeSoft, line: "none", radius: true });
  text(slide, "汇报时可以按这五个板块讲：谁在用、用来做什么、解决什么痛点。", {
    x: 150, y: 586, w: 980, h: 24, size: 16, color: COLORS.orange, bold: true, align: "center", valign: "middle",
  });
  footer(slide, 3);
}

// Slide 4
{
  const slide = presentation.slides.add();
  slide.background.fill = COLORS.bg;
  title(slide, "平台总控后台", "核心价值：客户开通、品牌差异、账号权限和基础配置统一管理。");
  shape(slide, { x: 58, y: 148, w: 330, h: 424, fill: COLORS.panel, line: COLORS.line, radius: true });
  tag(slide, "板块 1", 88, 178, 88, COLORS.orangeSoft, COLORS.orange, { h: 30, size: 13 });
  text(slide, "总控后台解决的是“客户怎么快速开起来、怎么差异化管理”的问题。", {
    x: 88, y: 226, w: 250, h: 74, size: 21, color: COLORS.ink, bold: true,
  });
  bullet(slide, [
    { text: "给不同客户配置不同品牌和登录页", color: COLORS.orange },
    { text: "开通机构、管理员和监管账号", color: COLORS.blue },
    { text: "统一设置权限、版本、默认角色", color: COLORS.purple },
    { text: "配置上传存储、字典和基础规则", color: COLORS.green },
  ], 90, 330, 235, 44, 15);
  shape(slide, { x: 86, y: 520, w: 260, h: 36, fill: COLORS.orangeSoft, line: "none", radius: true });
  text(slide, "痛点：不再为每家客户单独改一套系统", {
    x: 99, y: 527, w: 236, h: 20, size: 13.5, color: COLORS.orange, bold: true, align: "center", valign: "middle",
  });
  await imageCard(slide, { imagePath: images.loginConfig, x: 420, y: 148, w: 780, h: 248, caption: "登录页配置：品牌、主题、文案和模板可配置" });
  await imageCard(slide, { imagePath: images.institutionLogin, x: 420, y: 418, w: 250, h: 146, caption: "机构独立登录页" });
  await imageCard(slide, { imagePath: images.governmentAccount, x: 685, y: 418, w: 250, h: 146, caption: "政府账号与监管范围" });
  await imageCard(slide, { imagePath: images.storageConfig, x: 950, y: 418, w: 250, h: 146, caption: "云存储配置" });
  footer(slide, 4);
}

// Slide 5
{
  const slide = presentation.slides.add();
  slide.background.fill = COLORS.bg;
  title(slide, "机构运营后台", "核心价值：把校区每天要做的事情放到一套后台里，减少重复录入和信息割裂。");
  shape(slide, { x: 62, y: 150, w: 355, h: 420, fill: COLORS.panel, line: COLORS.line, radius: true });
  tag(slide, "板块 2", 92, 180, 88, COLORS.blueSoft, COLORS.blue, { h: 30, size: 13 });
  text(slide, "机构端主要服务校区管理者、教务老师和财务人员。", {
    x: 92, y: 226, w: 270, h: 56, size: 22, color: COLORS.ink, bold: true,
  });
  bullet(slide, [
    { text: "首页看待办、课程、学员和经营数据", color: COLORS.blue },
    { text: "课表排课、班级安排、老师安排集中管理", color: COLORS.green },
    { text: "上课记录、康复记录、评估结果可追踪", color: COLORS.purple },
    { text: "订单、欠费、报表减少线下对账", color: COLORS.orange },
  ], 94, 316, 250, 44, 15);
  shape(slide, { x: 91, y: 514, w: 278, h: 38, fill: COLORS.blueSoft, line: "none", radius: true });
  text(slide, "痛点：老师、教务、财务不用各管各的数据", {
    x: 104, y: 522, w: 252, h: 20, size: 13, color: COLORS.blue, bold: true, align: "center", valign: "middle",
  });
  await imageCard(slide, { imagePath: images.dashboard, x: 452, y: 150, w: 340, h: 224, caption: "机构首页：日常经营概览" });
  await imageCard(slide, { imagePath: images.timetable, x: 820, y: 150, w: 340, h: 224, caption: "智慧课表：排课与课时安排" });
  await imageCard(slide, { imagePath: images.timetableFull, x: 452, y: 398, w: 708, h: 172, caption: "课表板：班级、一对一、老师和时间统一展示" });
  footer(slide, 5);
}

// Slide 6
{
  const slide = presentation.slides.add();
  slide.background.fill = COLORS.bg;
  title(slide, "PEP-3 评估 Pad", "核心价值：把专业评估从纸笔流程变成可记录、可追踪、可出报告的线上流程。");
  const steps = [
    ["选量表", "从量表库进入评估"],
    ["现场录分", "按题目记录表现"],
    ["保存进度", "支持草稿和继续评估"],
    ["出报告", "沉淀为报告和计划"],
  ];
  steps.forEach((s, i) => {
    const x = 80 + i * 292;
    const color = [COLORS.orange, COLORS.purple, COLORS.green, COLORS.blue][i];
    const soft = [COLORS.orangeSoft, COLORS.purpleSoft, COLORS.greenSoft, COLORS.blueSoft][i];
    shape(slide, { x, y: 150, w: 220, h: 92, fill: soft, line: "none", radius: true });
    text(slide, String(i + 1), { x: x + 16, y: 172, w: 34, h: 34, size: 18, color: "#FFFFFF", bold: true, align: "center", valign: "middle", fill: color, radius: true });
    text(slide, s[0], { x: x + 62, y: 166, w: 120, h: 28, size: 19, color, bold: true });
    text(slide, s[1], { x: x + 62, y: 198, w: 130, h: 26, size: 13.2, color: COLORS.text });
  });
  await imageCard(slide, { imagePath: images.scaleLibrary, x: 60, y: 284, w: 350, h: 250, caption: "量表库：从评估任务进入" });
  await imageCard(slide, { imagePath: images.assessmentWorkbench, x: 465, y: 284, w: 350, h: 250, caption: "评估工作台：题目、进度、录分" });
  await imageCard(slide, { imagePath: images.reportList, x: 870, y: 284, w: 350, h: 250, caption: "报告列表：评估结果集中管理" });
  shape(slide, { x: 92, y: 570, w: 1096, h: 44, fill: COLORS.purpleSoft, line: "none", radius: true });
  text(slide, "亮点：PEP-3 的专业流程已经被做进产品，从“做评估”延伸到“看结果、出报告、做计划”。", {
    x: 118, y: 580, w: 1040, h: 24, size: 16, color: COLORS.purple, bold: true, align: "center", valign: "middle",
  });
  footer(slide, 6);
}

// Slide 7
{
  const slide = presentation.slides.add();
  slide.background.fill = COLORS.bg;
  title(slide, "结果沉淀与共享", "评估结果不只是一次记录，还能继续服务教学计划、家长沟通和监管汇总。");
  await imageCard(slide, { imagePath: images.iepPlan, x: 60, y: 148, w: 570, h: 330, caption: "IEP 计划：把评估结果转成教学 / 康复目标" });
  await imageCard(slide, { imagePath: images.iepPlanEdit, x: 665, y: 148, w: 550, h: 240, caption: "计划详情：目标、周期、人员和执行内容可沉淀" });
  const cards = [
    ["机构", "老师可以按计划安排训练、复盘进展，减少结果只停留在报告里的问题。", COLORS.blue, COLORS.blueSoft],
    ["家长", "家长端可以查看课表、记录、问卷和阶段结果，让沟通有依据。", COLORS.green, COLORS.greenSoft],
    ["监管", "政府端可以看机构、学员、服务和汇总指标，减少线下统计。", COLORS.teal, COLORS.tealSoft],
  ];
  cards.forEach((c, i) => {
    const x = 665 + i * 182;
    shape(slide, { x, y: 420, w: 166, h: 124, fill: c[3], line: "none", radius: true });
    text(slide, c[0], { x: x + 18, y: 438, w: 120, h: 28, size: 20, color: c[2], bold: true });
    text(slide, c[1], { x: x + 18, y: 474, w: 130, h: 56, size: 12.6, color: COLORS.text });
  });
  shape(slide, { x: 90, y: 522, w: 510, h: 54, fill: COLORS.greenSoft, line: "none", radius: true });
  text(slide, "痛点：服务结果可以被持续使用，不再是做完一次评估就结束。", {
    x: 116, y: 536, w: 458, h: 24, size: 16, color: COLORS.green, bold: true, align: "center", valign: "middle",
  });
  footer(slide, 7);
}

// Slide 8
{
  const slide = presentation.slides.add();
  slide.background.fill = COLORS.bg;
  title(slide, "项目亮点", "领导汇报时可以抓住这四个亮点，避免陷入技术细节。");
  const highlights = [
    {
      title: "配置化交付",
      body: "客户品牌、登录页、权限、版本、存储等都能配置，适合多客户复制。",
      color: COLORS.orange,
      soft: COLORS.orangeSoft,
    },
    {
      title: "业务闭环完整",
      body: "从客户开通、机构运营、排课上课，到评估报告和 IEP 计划，链路连起来了。",
      color: COLORS.blue,
      soft: COLORS.blueSoft,
    },
    {
      title: "专业评估数字化",
      body: "PEP-3 流程从量表、录分、进度、报告到计划在线化，体现垂直领域能力。",
      color: COLORS.purple,
      soft: COLORS.purpleSoft,
    },
    {
      title: "多角色协同",
      body: "平台、机构、老师、家长、监管各有入口，数据围绕同一业务对象沉淀。",
      color: COLORS.green,
      soft: COLORS.greenSoft,
    },
  ];
  highlights.forEach((h, i) => {
    const x = 80 + (i % 2) * 580;
    const y = 160 + Math.floor(i / 2) * 190;
    shape(slide, { x, y, w: 520, h: 150, fill: COLORS.panel, line: COLORS.line, radius: true });
    shape(slide, { x: x + 24, y: y + 28, w: 60, h: 60, fill: h.soft, line: "none", radius: true });
    text(slide, String(i + 1).padStart(2, "0"), { x: x + 24, y: y + 38, w: 60, h: 38, size: 21, color: h.color, bold: true, align: "center", valign: "middle" });
    text(slide, h.title, { x: x + 108, y: y + 28, w: 330, h: 32, size: 23, color: COLORS.ink, bold: true });
    text(slide, h.body, { x: x + 108, y: y + 70, w: 360, h: 48, size: 15.2, color: COLORS.text });
  });
  shape(slide, { x: 162, y: 570, w: 956, h: 48, fill: COLORS.orangeSoft, line: "none", radius: true });
  text(slide, "建议表达：这不是“又做了几个页面”，而是把业务交付方式从人工推进变成平台化配置和闭环运营。", {
    x: 190, y: 582, w: 900, h: 24, size: 16, color: COLORS.orange, bold: true, align: "center", valign: "middle",
  });
  footer(slide, 8);
}

// Slide 9
{
  const slide = presentation.slides.add();
  slide.background.fill = COLORS.bg;
  title(slide, "总结与下一步", "汇报最后用一页收住：项目价值、演示顺序、后续可继续增强的地方。");
  shape(slide, { x: 68, y: 152, w: 540, h: 390, fill: COLORS.panel, line: COLORS.line, radius: true });
  text(slide, "建议汇报顺序", { x: 102, y: 184, w: 220, h: 32, size: 24, color: COLORS.ink, bold: true });
  bullet(slide, [
    { text: "先讲痛点：开客户慢、运营散、评估重、数据难汇总", color: COLORS.orange },
    { text: "再讲板块：总控后台、机构后台、评估 Pad、家长端、政府端", color: COLORS.blue },
    { text: "重点演示：登录页配置、课表排课、PEP-3 录分和报告", color: COLORS.purple },
    { text: "最后落到价值：标准化交付、多端协同、结果沉淀", color: COLORS.green },
  ], 108, 240, 405, 56, 16);

  shape(slide, { x: 672, y: 152, w: 540, h: 390, fill: COLORS.panel, line: COLORS.line, radius: true });
  text(slide, "下一步建议", { x: 706, y: 184, w: 220, h: 32, size: 24, color: COLORS.ink, bold: true });
  bullet(slide, [
    { text: "准备一套完整演示数据，覆盖客户开通到出报告", color: COLORS.orange },
    { text: "补齐家长端、政府端的展示截图，汇报更完整", color: COLORS.green },
    { text: "整理现场演示账号和固定演示路径，减少临场切换", color: COLORS.blue },
    { text: "把高频客户差异沉淀成更多可配置项", color: COLORS.purple },
  ], 712, 240, 405, 56, 16);

  shape(slide, { x: 98, y: 590, w: 1084, h: 48, fill: COLORS.greenSoft, line: "none", radius: true });
  text(slide, "一句话结论：项目已经形成“开客户、管机构、做评估、出结果、看监管”的业务闭环，具备继续做产品化交付的基础。", {
    x: 126, y: 602, w: 1028, h: 24, size: 16, color: COLORS.green, bold: true, align: "center", valign: "middle",
  });
  footer(slide, 9);
}

for (let i = 0; i < presentation.slides.count; i += 1) {
  const slide = presentation.slides.getItem(i);
  const png = await presentation.export({ slide, format: "png", scale: 1 });
  await saveBlobToFile(png, path.join(SCRATCH_DIR, `slide-${String(i + 1).padStart(2, "0")}.png`));
}

const pptx = await PresentationFile.exportPptx(presentation);
await pptx.save(OUT);

await fs.writeFile(path.join(SCRATCH_DIR, "verification.json"), JSON.stringify({
  generatedAt: new Date().toISOString(),
  output: OUT,
  slides: presentation.slides.count,
  audience: "业务负责人和管理层",
  narrative: "以痛点、板块功能、项目亮点和业务价值为主，弱化技术细节。",
  screenshotSource: "Only user-provided project screenshots were used as visual evidence.",
}, null, 2));

console.log(OUT);
