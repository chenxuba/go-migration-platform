package service

import (
	"context"
	"errors"
	"strings"

	"go-migration-platform/pkg/autismdevscore"
	"go-migration-platform/services/education/internal/model"
)

type autismDevIEPPlanPromptPayload struct {
	Student        pep3IEPPlanPromptStudent              `json:"student"`
	Assessment     autismDevResultAnalysisAssessment     `json:"assessment"`
	Domains        []autismDevResultAnalysisPromptDomain `json:"domains"`
	Analysis       *model.AutismDevResultAnalysisVO      `json:"analysis,omitempty"`
	Interpretation *model.ERXinReportInterpretationVO    `json:"interpretation,omitempty"`
	RehabRecords   []pep3IEPPlanPromptRehabRecord        `json:"rehabRecords,omitempty"`
	OutputRequest  pep3IEPPlanPromptOutput               `json:"outputRequest"`
}

func (svc *Service) GenerateAutismDevIEPPlanWithAI(userID int64, recordID int64, durationMonths int) (model.PEP3IEPPlanAIResult, error) {
	if recordID <= 0 {
		return model.PEP3IEPPlanAIResult{}, errors.New("invalid assessment record id")
	}
	if durationMonths <= 0 {
		durationMonths = 6
	}
	ctx := context.Background()
	record, score, data, itemScores, analysis, interpretation, rehabRows, err := svc.prepareAutismDevIEPPlanSource(ctx, userID, recordID)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	currentTeacherName := svc.currentIEPPlanTeacherName(ctx, userID)
	payload := buildAutismDevIEPPlanPromptPayload(record, score, data, itemScores, analysis, interpretation, rehabRows, durationMonths)
	result, err := callDeepSeekIEPPlanWithPrompt(ctx, payload, autismDevIEPPlanSystemPrompt())
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	return normalizePEP3IEPPlanAIResult(result, record, rehabRows, currentTeacherName, durationMonths), nil
}

func (svc *Service) GenerateAutismDevIEPPlanWithAIStream(ctx context.Context, userID int64, recordID int64, durationMonths int, onDelta func(string) error) (model.PEP3IEPPlanAIResult, *model.DeepSeekUsageVO, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if recordID <= 0 {
		return model.PEP3IEPPlanAIResult{}, nil, errors.New("invalid assessment record id")
	}
	if durationMonths <= 0 {
		durationMonths = 6
	}
	record, score, data, itemScores, analysis, interpretation, rehabRows, err := svc.prepareAutismDevIEPPlanSource(ctx, userID, recordID)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, nil, err
	}
	currentTeacherName := svc.currentIEPPlanTeacherName(ctx, userID)
	payload := buildAutismDevIEPPlanPromptPayload(record, score, data, itemScores, analysis, interpretation, rehabRows, durationMonths)
	result, usage, err := callDeepSeekIEPPlanStreamWithPrompt(ctx, payload, autismDevIEPPlanSystemPrompt(), onDelta)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, usage, err
	}
	return normalizePEP3IEPPlanAIResult(result, record, rehabRows, currentTeacherName, durationMonths), usage, nil
}

func (svc *Service) SaveAutismDevIEPPlan(userID int64, req model.PEP3IEPPlanSaveRequest) (model.PEP3IEPPlanSavedVO, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, req.ID); err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	return svc.SavePEP3IEPPlan(userID, req)
}

func (svc *Service) GetAutismDevIEPPlan(userID, recordID int64, durationMonths int) (model.PEP3IEPPlanSavedVO, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, recordID); err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	return svc.GetPEP3IEPPlan(userID, recordID, durationMonths)
}

func (svc *Service) ExportAutismDevIEPPlanWord(userID int64, recordID int64, durationMonths int) (string, string, []byte, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, recordID); err != nil {
		return "", "", nil, err
	}
	saved, err := svc.GetPEP3IEPPlan(userID, recordID, durationMonths)
	if err != nil {
		return "", "", nil, err
	}
	if !saved.Exists || saved.Plan == nil || len(saved.Plan.Rows) == 0 {
		return "", "", nil, errors.New("暂无可导出的IEP计划")
	}
	return svc.ExportPEP3IEPPlanWordFromAIResult(userID, recordID, *saved.Plan, durationMonths)
}

func (svc *Service) ExportAutismDevIEPPlanWordFromAIResult(userID int64, recordID int64, planResult model.PEP3IEPPlanAIResult, durationMonths int) (string, string, []byte, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, recordID); err != nil {
		return "", "", nil, err
	}
	return svc.ExportPEP3IEPPlanWordFromAIResult(userID, recordID, planResult, durationMonths)
}

func (svc *Service) ExportAutismDevIEPPlanPDF(userID int64, recordID int64, durationMonths int) (string, string, []byte, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, recordID); err != nil {
		return "", "", nil, err
	}
	saved, err := svc.GetPEP3IEPPlan(userID, recordID, durationMonths)
	if err != nil {
		return "", "", nil, err
	}
	if !saved.Exists || saved.Plan == nil || len(saved.Plan.Rows) == 0 {
		return "", "", nil, errors.New("暂无可打印的IEP计划")
	}
	return svc.ExportPEP3IEPPlanPDFFromAIResult(userID, recordID, *saved.Plan, durationMonths)
}

func (svc *Service) ExportAutismDevIEPPlanPDFFromAIResult(userID int64, recordID int64, planResult model.PEP3IEPPlanAIResult, durationMonths int) (string, string, []byte, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, recordID); err != nil {
		return "", "", nil, err
	}
	return svc.ExportPEP3IEPPlanPDFFromAIResult(userID, recordID, planResult, durationMonths)
}

func (svc *Service) GenerateAutismDevExecutionPlanWithAI(ctx context.Context, userID int64, req model.PEP3ExecutionPlanGenerateRequest) (any, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, req.ID); err != nil {
		return nil, err
	}
	return svc.GeneratePEP3ExecutionPlanWithAI(ctx, userID, req)
}

func (svc *Service) GenerateAutismDevExecutionPlanWithAIStream(ctx context.Context, userID int64, req model.PEP3ExecutionPlanGenerateRequest, onDelta func(string) error) (any, *model.DeepSeekUsageVO, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, req.ID); err != nil {
		return nil, nil, err
	}
	return svc.GeneratePEP3ExecutionPlanWithAIStream(ctx, userID, req, onDelta)
}

func (svc *Service) SaveAutismDevExecutionPlan(userID int64, req model.PEP3ExecutionPlanSaveRequest) (model.PEP3ExecutionPlanSavedVO, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, req.ID); err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	return svc.SavePEP3ExecutionPlan(userID, req)
}

func (svc *Service) GetAutismDevExecutionPlans(userID, recordID int64, durationMonths int) (model.PEP3ExecutionPlanSavedVO, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, recordID); err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	return svc.GetPEP3ExecutionPlans(userID, recordID, durationMonths)
}

func (svc *Service) ExportAutismDevExecutionPlanWord(userID int64, req model.PEP3ExecutionPlanWordExportRequest) (string, string, []byte, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, req.ID); err != nil {
		return "", "", nil, err
	}
	return svc.ExportPEP3ExecutionPlanWord(userID, req)
}

func (svc *Service) ExportAutismDevExecutionPlanPDF(userID int64, req model.PEP3ExecutionPlanWordExportRequest) (string, string, []byte, error) {
	if _, err := svc.validateAutismDevIEPPlanRecord(userID, req.ID); err != nil {
		return "", "", nil, err
	}
	return svc.ExportPEP3ExecutionPlanPDF(userID, req)
}

func (svc *Service) prepareAutismDevIEPPlanSource(ctx context.Context, userID, recordID int64) (model.AssessmentRecordDetailVO, autismdevscore.AssessmentResult, autismDevStaticData, map[int]string, model.AutismDevResultAnalysisVO, model.ERXinReportInterpretationVO, []pep3IEPPlanPromptRehabRecord, error) {
	_, record, score, data, itemScores, err := svc.autismDevResultAnalysisContext(userID, recordID)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, autismdevscore.AssessmentResult{}, autismDevStaticData{}, nil, model.AutismDevResultAnalysisVO{}, model.ERXinReportInterpretationVO{}, nil, err
	}
	analysis, err := svc.GetAutismDevResultAnalysis(userID, recordID)
	if err != nil || autismDevResultAnalysisIsEmpty(analysis) {
		analysis = buildRuleBasedAutismDevResultAnalysis(record, score, data, itemScores)
	}
	interpretation, err := svc.GetAutismDevReportInterpretation(userID, recordID)
	if err != nil {
		interpretation = model.ERXinReportInterpretationVO{}
	}
	rehabRows, err := svc.autismDevIEPPlanPromptRehabRecords(ctx, userID, record)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, autismdevscore.AssessmentResult{}, autismDevStaticData{}, nil, model.AutismDevResultAnalysisVO{}, model.ERXinReportInterpretationVO{}, nil, err
	}
	return record, score, data, itemScores, analysis, interpretation, rehabRows, nil
}

func (svc *Service) validateAutismDevIEPPlanRecord(userID, recordID int64) (model.AssessmentRecordDetailVO, error) {
	if recordID <= 0 {
		return model.AssessmentRecordDetailVO{}, errors.New("invalid assessment record id")
	}
	return svc.GetAutismDevAssessmentRecord(userID, recordID)
}

func (svc *Service) autismDevIEPPlanPromptRehabRecords(ctx context.Context, userID int64, record model.AssessmentRecordDetailVO) ([]pep3IEPPlanPromptRehabRecord, error) {
	if record.StudentID <= 0 {
		return nil, nil
	}
	instID, err := svc.pep3AssessmentInstID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := svc.repo.ListRecentPublishedRehabRecordRows(ctx, instID, record.StudentID, 12)
	if err != nil {
		return nil, err
	}
	return buildPEP3IEPPlanPromptRehabRecords(rows), nil
}

func buildAutismDevIEPPlanPromptPayload(record model.AssessmentRecordDetailVO, score autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string, analysis model.AutismDevResultAnalysisVO, interpretation model.ERXinReportInterpretationVO, rehabRecords []pep3IEPPlanPromptRehabRecord, durationMonths int) autismDevIEPPlanPromptPayload {
	payload := autismDevIEPPlanPromptPayload{
		Student: pep3IEPPlanPromptStudent{
			Name:      strings.TrimSpace(record.StudentName),
			Gender:    strings.TrimSpace(record.StudentGender),
			BirthDate: formatIEPPlanDate(record.BirthDate),
			Age:       formatIEPPlanAge(record.AgeYears, record.AgeMonths, record.AgeDays),
		},
		Assessment: autismDevResultAnalysisAssessment{
			Date:         formatIEPPlanDate(record.AssessmentDate),
			ScaleVersion: strings.TrimSpace(record.ScaleVersion),
			DataStatus:   strings.TrimSpace(record.DataStatus),
			Remark:       strings.TrimSpace(record.Remark),
		},
		Domains:      autismDevIEPPlanPromptDomains(score, data, itemScores),
		RehabRecords: rehabRecords,
		OutputRequest: pep3IEPPlanPromptOutput{
			Title:          iepPlanTitle(durationMonths),
			DurationMonths: durationMonths,
			RequiredSchema: "只输出JSON：title, student{name,gender,birthDate}, meta{planDate,participant,implementer,startDate,endDate}, rows[{domain,longGoal,shortGoal,courseForm,startEndDate}]。rows必须沿用PEP3 IEP模板的表格结构；不要输出家庭干预计划。domain优先使用孤独症儿童发展评估8大项中文领域名；每个需要训练的主要领域至少3行rows；每行shortGoal只能放1条短期目标；同一领域longGoal必须完全相同，写成至少2条编号长期目标并用\\n分隔；courseForm常见值为个训、集体课；startEndDate按自然月份阶段填写，不能每行都写整个计划周期。",
		},
	}
	if !autismDevResultAnalysisIsEmpty(analysis) {
		payload.Analysis = &analysis
	}
	if !erxinReportInterpretationIsEmpty(interpretation) {
		payload.Interpretation = &interpretation
	}
	return payload
}

func autismDevIEPPlanPromptDomains(result autismdevscore.AssessmentResult, data autismDevStaticData, itemScores map[int]string) []autismDevResultAnalysisPromptDomain {
	byCode := make(map[string]autismdevscore.DomainResult, len(result.Domains))
	for _, domain := range result.Domains {
		byCode[strings.TrimSpace(domain.DomainCode)] = domain
	}
	itemsByDomain := autismDevItemsByDomain(data.items)
	out := make([]autismDevResultAnalysisPromptDomain, 0, len(autismdevscore.DomainOrder))
	for _, code := range autismdevscore.DomainOrder {
		domain, ok := byCode[code]
		if !ok || strings.TrimSpace(domain.DomainName) == "" || domain.AnsweredItemCount <= 0 {
			continue
		}
		passedItems, emergingItems, failedItems := autismDevIEPPlanScoredItems(itemsByDomain[code], itemScores)
		targetItems := autismDevResultAnalysisTargetItems(emergingItems, failedItems)
		out = append(out, autismDevResultAnalysisPromptDomain{
			DomainCode:        strings.TrimSpace(domain.DomainCode),
			DomainName:        strings.TrimSpace(domain.DomainName),
			ScoreType:         strings.TrimSpace(domain.ScoreType),
			ItemCount:         domain.ItemCount,
			AnsweredItemCount: domain.AnsweredItemCount,
			MissingItemCount:  domain.MissingItemCount,
			Complete:          domain.Complete,
			PCount:            domain.PCount,
			ECount:            domain.ECount,
			FCount:            domain.FCount,
			XCount:            domain.XCount,
			PECount:           domain.PECount,
			RawScore:          domain.RawScore,
			ScorableItemCount: domain.ScorableItemCount,
			ScoreRate:         domain.ScoreRate,
			ACount:            domain.ACount,
			MCount:            domain.MCount,
			SCount:            domain.SCount,
			AdaptiveCount:     domain.AdaptiveCount,
			AbnormalCount:     domain.AbnormalCount,
			DevelopmentLevel:  autismDevResultAnalysisDevelopmentLevel(passedItems),
			PassedItems:       autismDevResultAnalysisSelectItems(passedItems, 8, true),
			EmergingItems:     autismDevResultAnalysisSelectItems(emergingItems, 8, false),
			FailedItems:       autismDevResultAnalysisSelectItems(failedItems, 8, false),
			TargetItems:       autismDevResultAnalysisSelectItems(targetItems, autismDevResultAnalysisPromptTargetLimit(targetItems), false),
			Warnings:          append([]string(nil), domain.Warnings...),
		})
	}
	return out
}

func autismDevIEPPlanScoredItems(items []autismdevscore.ItemDefinition, itemScores map[int]string) ([]autismDevResultAnalysisPromptItem, []autismDevResultAnalysisPromptItem, []autismDevResultAnalysisPromptItem) {
	passed := make([]autismDevResultAnalysisPromptItem, 0)
	emerging := make([]autismDevResultAnalysisPromptItem, 0)
	failed := make([]autismDevResultAnalysisPromptItem, 0)
	for _, item := range items {
		score := normalizeAutismDevScore(itemScores[item.ItemNo])
		if score == "" || score == autismdevscore.ScoreX {
			continue
		}
		promptItem := autismDevResultAnalysisPromptItemFromDefinition(item, score)
		switch score {
		case autismdevscore.ScoreP, autismdevscore.ScoreA:
			passed = append(passed, promptItem)
		case autismdevscore.ScoreE, autismdevscore.ScoreM:
			emerging = append(emerging, promptItem)
		case autismdevscore.ScoreF, autismdevscore.ScoreS:
			failed = append(failed, promptItem)
		}
	}
	return passed, emerging, failed
}

func autismDevIEPPlanSystemPrompt() string {
	return strings.Join([]string{
		"你是儿童康复机构的IEP计划生成助手。",
		"根据孤独症儿童发展评估表8大项记录、具体题目得分、报告解读、结果分析和近期训练记录，生成可落地的康复教学计划；如果没有报告解读，就依据测评结果、结果分析和训练记录生成。",
		"必须输出严格JSON，不要Markdown，不要代码块，不要解释。",
		"输出模板必须与PEP3 IEP计划一致：康复领域、长期目标、短期目标、课程形式、起止日期。",
		"不得更改P/E/F/X/A/M/S计数、题目、年龄段、日期等测评事实；不得做医学诊断；目标要转化为可训练、可观察、可执行的教学目标。",
		"康复领域优先围绕孤独症儿童发展评估8大项：感知觉、大运动、精细动作、语言与沟通、认知、社会交往、生活自理、情绪与行为；可根据结果分析和近期训练记录合并或补充综合康复领域。",
		"优先从emergingItems和failedItems中选择近期关键目标，passedItems只作为优势和泛化基础；情绪与行为领域要写成可观察的适应行为或替代行为训练目标。",
		"不要输出家庭干预计划。每个主要康复领域至少输出3行短期目标，一行只能放1条短期目标。",
		"同一康复领域的longGoal要写成同一个字符串，至少包含2条长期目标，用\\n分隔并编号；同领域每行longGoal保持完全相同，便于合并成一个单元格。",
		"courseForm要根据目标场景和实际干预组织方式判断；一对一、个训、个别训练写个训；集体、小组、融合、团体场景写集体课。",
		"起止日期要按自然月份阶段划分：3个月计划分3段，每段1个月；6个月计划分3段，每段2个月；不要把每条短期目标都写成生成当天到半年后。",
	}, "\n")
}
