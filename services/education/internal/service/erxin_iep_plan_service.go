package service

import (
	"context"
	"errors"
	"strings"

	"go-migration-platform/services/education/internal/model"
)

type erxinIEPPlanPromptPayload struct {
	Student        pep3IEPPlanPromptStudent           `json:"student"`
	Assessment     erxinIEPPlanPromptAssessment       `json:"assessment"`
	ReportSummary  model.ERXinReportSummary           `json:"reportSummary"`
	DomainRows     []model.ERXinReportDomainRow       `json:"domainRows"`
	Interpretation *model.ERXinReportInterpretationVO `json:"interpretation,omitempty"`
	RehabRecords   []pep3IEPPlanPromptRehabRecord     `json:"rehabRecords,omitempty"`
	OutputRequest  pep3IEPPlanPromptOutput            `json:"outputRequest"`
}

type erxinIEPPlanPromptAssessment struct {
	Date           string   `json:"date,omitempty"`
	Version        string   `json:"version,omitempty"`
	SourceStandard string   `json:"sourceStandard,omitempty"`
	DataStatus     string   `json:"dataStatus,omitempty"`
	Remark         string   `json:"remark,omitempty"`
	Warnings       []string `json:"warnings,omitempty"`
}

func (svc *Service) GenerateERXinIEPPlanWithAI(userID int64, recordID int64, durationMonths int) (model.PEP3IEPPlanAIResult, error) {
	if recordID <= 0 {
		return model.PEP3IEPPlanAIResult{}, errors.New("invalid assessment record id")
	}
	if durationMonths <= 0 {
		durationMonths = 6
	}
	ctx := context.Background()
	record, report, interpretation, err := svc.prepareERXinIEPPlanSource(ctx, userID, recordID)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	currentTeacherName := svc.currentIEPPlanTeacherName(ctx, userID)
	rehabRows, err := svc.erxinIEPPlanPromptRehabRecords(ctx, userID, record)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	payload := buildERXinIEPPlanPromptPayload(record, report, interpretation, rehabRows, durationMonths)
	result, err := callDeepSeekIEPPlanWithPrompt(ctx, payload, erxinIEPPlanSystemPrompt())
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, err
	}
	return normalizePEP3IEPPlanAIResult(result, record, rehabRows, currentTeacherName, durationMonths), nil
}

func (svc *Service) GenerateERXinIEPPlanWithAIStream(ctx context.Context, userID int64, recordID int64, durationMonths int, onDelta func(string) error) (model.PEP3IEPPlanAIResult, *model.DeepSeekUsageVO, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if recordID <= 0 {
		return model.PEP3IEPPlanAIResult{}, nil, errors.New("invalid assessment record id")
	}
	if durationMonths <= 0 {
		durationMonths = 6
	}
	record, report, interpretation, err := svc.prepareERXinIEPPlanSource(ctx, userID, recordID)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, nil, err
	}
	currentTeacherName := svc.currentIEPPlanTeacherName(ctx, userID)
	rehabRows, err := svc.erxinIEPPlanPromptRehabRecords(ctx, userID, record)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, nil, err
	}
	payload := buildERXinIEPPlanPromptPayload(record, report, interpretation, rehabRows, durationMonths)
	result, usage, err := callDeepSeekIEPPlanStreamWithPrompt(ctx, payload, erxinIEPPlanSystemPrompt(), onDelta)
	if err != nil {
		return model.PEP3IEPPlanAIResult{}, usage, err
	}
	return normalizePEP3IEPPlanAIResult(result, record, rehabRows, currentTeacherName, durationMonths), usage, nil
}

func (svc *Service) SaveERXinIEPPlan(userID int64, req model.PEP3IEPPlanSaveRequest) (model.PEP3IEPPlanSavedVO, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, req.ID); err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	return svc.SavePEP3IEPPlan(userID, req)
}

func (svc *Service) GetERXinIEPPlan(userID, recordID int64, durationMonths int) (model.PEP3IEPPlanSavedVO, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, recordID); err != nil {
		return model.PEP3IEPPlanSavedVO{}, err
	}
	return svc.GetPEP3IEPPlan(userID, recordID, durationMonths)
}

func (svc *Service) ExportERXinIEPPlanWord(userID int64, recordID int64, durationMonths int) (string, string, []byte, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, recordID); err != nil {
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

func (svc *Service) ExportERXinIEPPlanWordFromAIResult(userID int64, recordID int64, planResult model.PEP3IEPPlanAIResult, durationMonths int) (string, string, []byte, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, recordID); err != nil {
		return "", "", nil, err
	}
	return svc.ExportPEP3IEPPlanWordFromAIResult(userID, recordID, planResult, durationMonths)
}

func (svc *Service) ExportERXinIEPPlanPDF(userID int64, recordID int64, durationMonths int) (string, string, []byte, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, recordID); err != nil {
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

func (svc *Service) ExportERXinIEPPlanPDFFromAIResult(userID int64, recordID int64, planResult model.PEP3IEPPlanAIResult, durationMonths int) (string, string, []byte, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, recordID); err != nil {
		return "", "", nil, err
	}
	return svc.ExportPEP3IEPPlanPDFFromAIResult(userID, recordID, planResult, durationMonths)
}

func (svc *Service) GenerateERXinExecutionPlanWithAI(ctx context.Context, userID int64, req model.PEP3ExecutionPlanGenerateRequest) (any, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, req.ID); err != nil {
		return nil, err
	}
	return svc.GeneratePEP3ExecutionPlanWithAI(ctx, userID, req)
}

func (svc *Service) GenerateERXinExecutionPlanWithAIStream(ctx context.Context, userID int64, req model.PEP3ExecutionPlanGenerateRequest, onDelta func(string) error) (any, *model.DeepSeekUsageVO, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, req.ID); err != nil {
		return nil, nil, err
	}
	return svc.GeneratePEP3ExecutionPlanWithAIStream(ctx, userID, req, onDelta)
}

func (svc *Service) SaveERXinExecutionPlan(userID int64, req model.PEP3ExecutionPlanSaveRequest) (model.PEP3ExecutionPlanSavedVO, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, req.ID); err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	return svc.SavePEP3ExecutionPlan(userID, req)
}

func (svc *Service) GetERXinExecutionPlans(userID, recordID int64, durationMonths int) (model.PEP3ExecutionPlanSavedVO, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, recordID); err != nil {
		return model.PEP3ExecutionPlanSavedVO{}, err
	}
	return svc.GetPEP3ExecutionPlans(userID, recordID, durationMonths)
}

func (svc *Service) ExportERXinExecutionPlanWord(userID int64, req model.PEP3ExecutionPlanWordExportRequest) (string, string, []byte, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, req.ID); err != nil {
		return "", "", nil, err
	}
	return svc.ExportPEP3ExecutionPlanWord(userID, req)
}

func (svc *Service) ExportERXinExecutionPlanPDF(userID int64, req model.PEP3ExecutionPlanWordExportRequest) (string, string, []byte, error) {
	if _, err := svc.validateERXinIEPPlanRecord(userID, req.ID); err != nil {
		return "", "", nil, err
	}
	return svc.ExportPEP3ExecutionPlanPDF(userID, req)
}

func (svc *Service) prepareERXinIEPPlanSource(ctx context.Context, userID, recordID int64) (model.AssessmentRecordDetailVO, model.ERXinReportVO, model.ERXinReportInterpretationVO, error) {
	record, err := svc.validateERXinIEPPlanRecord(userID, recordID)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, model.ERXinReportVO{}, model.ERXinReportInterpretationVO{}, err
	}
	report, err := svc.GetERXinAssessmentReport(userID, recordID)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, model.ERXinReportVO{}, model.ERXinReportInterpretationVO{}, err
	}
	interpretation, err := svc.GetERXinReportInterpretation(userID, recordID)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, model.ERXinReportVO{}, model.ERXinReportInterpretationVO{}, err
	}
	return record, report, interpretation, nil
}

func (svc *Service) validateERXinIEPPlanRecord(userID, recordID int64) (model.AssessmentRecordDetailVO, error) {
	if recordID <= 0 {
		return model.AssessmentRecordDetailVO{}, errors.New("invalid assessment record id")
	}
	return svc.GetERXinAssessmentRecord(userID, recordID)
}

func (svc *Service) erxinIEPPlanPromptRehabRecords(ctx context.Context, userID int64, record model.AssessmentRecordDetailVO) ([]pep3IEPPlanPromptRehabRecord, error) {
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

func buildERXinIEPPlanPromptPayload(record model.AssessmentRecordDetailVO, report model.ERXinReportVO, interpretation model.ERXinReportInterpretationVO, rehabRecords []pep3IEPPlanPromptRehabRecord, durationMonths int) erxinIEPPlanPromptPayload {
	payload := erxinIEPPlanPromptPayload{
		Student: pep3IEPPlanPromptStudent{
			Name:      strings.TrimSpace(record.StudentName),
			Gender:    strings.TrimSpace(record.StudentGender),
			BirthDate: formatIEPPlanDate(record.BirthDate),
			Age:       formatIEPPlanAge(record.AgeYears, record.AgeMonths, record.AgeDays),
		},
		Assessment: erxinIEPPlanPromptAssessment{
			Date:           formatIEPPlanDate(record.AssessmentDate),
			Version:        strings.TrimSpace(report.ScaleVersion),
			SourceStandard: strings.TrimSpace(report.SourceStandard),
			DataStatus:     strings.TrimSpace(report.DataStatus),
			Remark:         strings.TrimSpace(record.Remark),
			Warnings:       append([]string(nil), report.Warnings...),
		},
		ReportSummary: report.Summary,
		DomainRows:    append([]model.ERXinReportDomainRow(nil), report.DomainRows...),
		RehabRecords:  rehabRecords,
		OutputRequest: pep3IEPPlanPromptOutput{
			Title:          iepPlanTitle(durationMonths),
			DurationMonths: durationMonths,
			RequiredSchema: "只输出JSON：title, student{name,gender,birthDate}, meta{planDate,participant,implementer,startDate,endDate}, rows[{domain,longGoal,shortGoal,courseForm,startEndDate}]。rows必须沿用PEP3 IEP模板的表格结构；不要输出家庭干预计划。domain优先使用儿心五大能区名称；每个主要能区至少3行rows；每行shortGoal只能放1条短期目标；同一领域longGoal必须完全相同，写成至少2条编号长期目标并用\\n分隔；courseForm常见值为个训、集体课；startEndDate按自然月份阶段填写，不能每行都写整个计划周期。",
		},
	}
	if !erxinReportInterpretationIsEmpty(interpretation) {
		payload.Interpretation = &interpretation
	}
	return payload
}

func erxinIEPPlanSystemPrompt() string {
	return strings.Join([]string{
		"你是儿童康复机构的IEP计划生成助手。",
		"根据儿心量表-II测评记录结果、五大能区评分以及可用的报告解读，生成可落地的康复教学计划；如果没有报告解读，就只依据测评结果和评分数据生成。",
		"必须输出严格JSON，不要Markdown，不要代码块，不要解释。",
		"输出模板必须与PEP3 IEP计划一致：康复领域、长期目标、短期目标、课程形式、起止日期。",
		"不得更改DQ、智龄、等级、日期等测评事实；不得做医学诊断；目标要转化为可训练、可观察、可执行的教学目标。",
		"康复领域优先围绕儿心五大能区：大运动、精细动作、适应能力、语言、个人-社交；可根据报告解读合并或补充综合康复领域。",
		"不要输出家庭干预计划。每个主要康复领域至少输出3行短期目标，一行只能放1条短期目标。",
		"同一康复领域的longGoal要写成同一个字符串，至少包含2条长期目标，用\\n分隔并编号；同领域每行longGoal保持完全相同，便于合并成一个单元格。",
		"courseForm要根据目标场景和实际干预组织方式判断；一对一、个训、个别训练写个训；集体、小组、融合、团体场景写集体课。",
		"起止日期要按自然月份阶段划分：3个月计划分3段，每段1个月；6个月计划分3段，每段2个月；不要把每条短期目标都写成生成当天到半年后。",
	}, "\n")
}
