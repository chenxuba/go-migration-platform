package service

import (
	"os"
	"path/filepath"
	"testing"

	"go-migration-platform/services/education/internal/model"
)

func TestPreviewDOCXToPDFExports(t *testing.T) {
	if os.Getenv("RUN_DOCX_PDF_PREVIEW") != "1" {
		t.Skip("set RUN_DOCX_PDF_PREVIEW=1 to generate local preview PDFs")
	}
	defer stopDefaultSofficeDaemonForTest(t)

	outputDir := filepath.Join(os.TempDir(), "iep-pdf-preview")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		t.Fatalf("mkdir output dir: %v", err)
	}

	writePreviewPDF := func(name, sourceFileName string, docxBytes []byte) {
		t.Helper()
		pdfBytes, err := convertDOCXToPDFBytes(docxBytes, sourceFileName)
		if err != nil {
			t.Fatalf("convert %s to pdf: %v", name, err)
		}
		pdfPath := filepath.Join(outputDir, name+".pdf")
		if err := os.WriteFile(pdfPath, pdfBytes, 0o644); err != nil {
			t.Fatalf("write %s: %v", pdfPath, err)
		}
		t.Logf("%s", pdfPath)
	}

	totalDocx, err := buildPEP3IEPPlanWordDocx(buildPEP3IEPPlanWordExportFromAIResult(previewIEPPlanResult(), 6))
	if err != nil {
		t.Fatalf("build total plan docx: %v", err)
	}
	writePreviewPDF("pep3-total-plan-preview", "pep3-total-plan-preview.docx", totalDocx)

	monthlyDocx, err := buildPEP3MonthlyPlanWordDocx(previewMonthlyPlanResult())
	if err != nil {
		t.Fatalf("build monthly plan docx: %v", err)
	}
	writePreviewPDF("pep3-monthly-plan-preview", "pep3-monthly-plan-preview.docx", monthlyDocx)

	weeklyDocx, err := buildPEP3WeeklyPlanWordDocx(previewWeeklyPlanResult())
	if err != nil {
		t.Fatalf("build weekly plan docx: %v", err)
	}
	writePreviewPDF("pep3-weekly-plan-preview", "pep3-weekly-plan-preview.docx", weeklyDocx)
}

func previewIEPPlanResult() model.PEP3IEPPlanAIResult {
	return model.PEP3IEPPlanAIResult{
		Title: "康复教学计划",
		Student: model.PEP3IEPPlanStudent{
			Name:      "陈旭",
			Gender:    "男",
			BirthDate: "2021-07-18",
		},
		Meta: model.PEP3IEPPlanMeta{
			PlanDate:    "2026-05-09",
			Participant: "原评估老师、班主任",
			Implementer: "原评估老师",
			StartDate:   "2026-05-04",
			EndDate:     "2026-10-31",
		},
		Rows: []model.PEP3IEPPlanRow{
			{
				Domain:       "大肌肉",
				LongGoal:     "1. 提升平衡与下肢协调能力。\n2. 能连续向前跳跃并保持动作稳定。",
				ShortGoal:    "能在口头提示下完成双脚连续向前跳跃 3 次，并在落地后保持身体平衡。",
				CourseForm:   "个训",
				StartEndDate: "2026-05-04 - 2026-06-30",
			},
			{
				Domain:       "大肌肉",
				LongGoal:     "1. 提升平衡与下肢协调能力。\n2. 能连续向前跳跃并保持动作稳定。",
				ShortGoal:    "能在示范辅助下沿地面标记完成跨步、跳圈等组合动作，并减少停顿。",
				CourseForm:   "个训",
				StartEndDate: "2026-07-01 - 2026-08-31",
			},
			{
				Domain:       "情感表达",
				LongGoal:     "1. 能在训练中主动表达需要与拒绝。\n2. 在互动活动中使用简单语句回应他人。",
				ShortGoal:    "在高频课堂情境中，能使用“我要”“不要”“帮我”等词句主动表达需求。",
				CourseForm:   "个训",
				StartEndDate: "2026-05-04 - 2026-07-31",
			},
			{
				Domain:       "情感表达",
				LongGoal:     "1. 能在训练中主动表达需要与拒绝。\n2. 在互动活动中使用简单语句回应他人。",
				ShortGoal:    "在轮替互动中，能在语言提示下回应教师提问，并维持 1 到 2 轮对答。",
				CourseForm:   "集体课",
				StartEndDate: "2026-08-01 - 2026-10-31",
			},
		},
	}
}

func previewMonthlyPlanResult() model.PEP3MonthlyPlanAIResult {
	return model.PEP3MonthlyPlanAIResult{
		Title: "康复教学5月计划",
		Student: model.PEP3IEPPlanStudent{
			Name:      "陈旭",
			Gender:    "男",
			BirthDate: "2021-07-18",
		},
		Meta: model.PEP3MonthlyPlanMeta{
			PlanDate:    "2026-05-09",
			Participant: "原评估老师、班主任",
			Implementer: "原评估老师",
			StartDate:   "2026-05-04",
			EndDate:     "2026-05-31",
			MonthLabel:  "2026年5月",
			SourceTitle: "康复教学计划",
		},
		Rows: []model.PEP3MonthlyPlanRow{
			{
				Domain:     "大肌肉",
				LongGoal:   "1. 提升平衡与下肢协调能力。\n2. 能连续向前跳跃并保持动作稳定。",
				ShortGoal:  "能在口头提示下完成双脚连续向前跳跃 3 次，并在落地后保持身体平衡。",
				CourseForm: "个训",
				TrainingItems: []model.PEP3MonthlyTrainingItem{
					{Content: "训练项目：连续跳圈。\n训练内容：按照地面脚印与圆圈提示完成双脚向前连续跳跃，练习落地稳定与身体前移控制。", StartEndDate: "2026-05-04 - 2026-05-10"},
					{Content: "训练项目：跨步跳组合。\n训练内容：结合跨步、跳圈和折返路径，提升动作计划与节奏衔接。", StartEndDate: "2026-05-11 - 2026-05-17"},
				},
			},
			{
				Domain:     "大肌肉",
				LongGoal:   "1. 提升平衡与下肢协调能力。\n2. 能连续向前跳跃并保持动作稳定。",
				ShortGoal:  "能在示范辅助下沿地面标记完成跨步、跳圈等组合动作，并减少停顿。",
				CourseForm: "个训",
				TrainingItems: []model.PEP3MonthlyTrainingItem{
					{Content: "训练项目：障碍跨越。\n训练内容：在不同高度与间距的软障碍中完成跨越、跳跃和转身。", StartEndDate: "2026-05-18 - 2026-05-24"},
					{Content: "训练项目：平衡路径。\n训练内容：在平衡垫、地胶线和定点标识间完成连续动作并减少外界提示。", StartEndDate: "2026-05-25 - 2026-05-31"},
				},
			},
			{
				Domain:     "情感表达",
				LongGoal:   "1. 能在训练中主动表达需要与拒绝。\n2. 在互动活动中使用简单语句回应他人。",
				ShortGoal:  "在高频课堂情境中，能使用“我要”“不要”“帮我”等词句主动表达需求。",
				CourseForm: "个训",
				TrainingItems: []model.PEP3MonthlyTrainingItem{
					{Content: "训练项目：选择表达。\n训练内容：在玩具、食物与活动选择中主动说出偏好或拒绝。", StartEndDate: "2026-05-04 - 2026-05-17"},
					{Content: "训练项目：请求帮助。\n训练内容：创设够不到、打不开、需要协助的任务，练习主动表达“帮我”。", StartEndDate: "2026-05-18 - 2026-05-31"},
				},
			},
		},
	}
}

func previewWeeklyPlanResult() model.PEP3WeeklyPlanAIResult {
	return model.PEP3WeeklyPlanAIResult{
		Title:        "康复教学周计划日记录卡5月第2周",
		Student:      model.PEP3IEPPlanStudent{Name: "陈旭"},
		TeacherName:  "原评估老师",
		CourseName:   "按大小、颜色、形状三维度分类物品",
		TrainingDate: "2026年5月第2周",
		Preparation:  "分类盒、不同颜色积木、大小圆片、形状卡片、奖励贴纸。根据孩子状态调整材料数量，先做熟悉任务再进入新任务。",
		WeekDates:    []string{"5.11", "5.12", "5.13", "5.14", "5.15", "5.16"},
		Rows: []model.PEP3WeeklyPlanRow{
			{
				Project: "训练前准备",
				Content: "通过问候、视觉日程和坐姿调整完成上课准备；回顾上节课完成情况，建立进入课堂的稳定状态。",
			},
			{
				Project:    "训练项目",
				Content:    "按大小、颜色、形状三维度分类物品。先进行单一维度配对，再过渡到双维度选择，最后在教师语言提示下完成三维度分类。",
				Completion: []string{"", "", "", "", "", ""},
			},
			{
				Project:    "训练内容",
				Content:    "1. 颜色分类：红黄蓝三色积木按篮筐投放。\n2. 大小配对：根据示例将大圆片、小圆片分别放入对应区域。\n3. 形状辨识：在圆形、三角形、正方形中完成命名与配对。\n4. 综合练习：听口令完成“把大的红色圆形放到这里”。",
				Completion: []string{"√", "S", "G", "", "", ""},
			},
			{
				Project:    "当天训练情况记录",
				Content:    "观察是否能保持注意、是否主动表达需要，以及在视觉提示撤除后是否仍能完成分类任务。",
				Completion: []string{"√", "G", "M", "", "", ""},
			},
		},
	}
}
