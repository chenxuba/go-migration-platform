package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

const (
	pep3ScaleCode       = "PEP3"
	pep3ScaleVersion    = "2025-92mo-draft"
	pep3LegacyVersion   = "2025-draft"
	pep3NormSourcePDF   = "PEP-3常模(2025).pdf"
	pep3ItemBankFile    = "pep3-item-bank-simplified-draft.json"
	pep3DomainMapFile   = "pep3-score-domain-map.json"
	pep3NormFile        = "pep3-norm-conversion-ocr-draft.json"
	pep3CorrectionFile  = "pep3-norm-manual-corrections.json"
	pep3DraftDataStatus = "题库为简体整理稿；常模为OCR草稿并叠加人工校对修正，投产前需完成全表核验"
)

type PEP3ScoreDataInfo struct {
	ScaleCode    string `json:"scaleCode"`
	ScaleVersion string `json:"scaleVersion"`
	model.PEP3NormDataInfo
	DataStatus string   `json:"dataStatus"`
	Sources    []string `json:"sources"`
}

type PEP3ScoreResponse struct {
	PEP3ScoreDataInfo
	Result pep3score.AssessmentResult `json:"result"`
}

type PEP3AssessmentRecordSaveInput struct {
	StudentID     int64
	StudentName   string
	ExaminerName  string
	Remark        string
	ScoreInput    pep3score.AssessmentInput
	InputSnapshot any
}

var (
	pep3EngineOnce    sync.Once
	pep3Engine        *pep3score.Engine
	pep3EngineInfo    PEP3ScoreDataInfo
	pep3EngineLoadErr error
)

func (svc *Service) ScorePEP3(input pep3score.AssessmentInput) (PEP3ScoreResponse, error) {
	engine, info, err := loadPEP3Engine()
	if err != nil {
		return PEP3ScoreResponse{}, err
	}
	result, err := engine.Score(input)
	if err != nil {
		return PEP3ScoreResponse{}, err
	}
	return PEP3ScoreResponse{
		PEP3ScoreDataInfo: info,
		Result:            result,
	}, nil
}

func (svc *Service) CreatePEP3AssessmentRecord(userID int64, input PEP3AssessmentRecordSaveInput) (model.AssessmentRecordDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentRecordDetailVO{}, errors.New("no institution context")
		}
		return model.AssessmentRecordDetailVO{}, err
	}
	examinerID, err := svc.repo.FindInstUserIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentRecordDetailVO{}, errors.New("no institution user context")
		}
		return model.AssessmentRecordDetailVO{}, err
	}
	examinerName := strings.TrimSpace(input.ExaminerName)
	if examinerName == "" {
		examinerName = svc.repo.GetStaffNameByID(context.Background(), &examinerID)
	}

	scoreResult, err := svc.ScorePEP3(input.ScoreInput)
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	recordID, err := svc.repo.CreateAssessmentRecord(context.Background(), repository.AssessmentRecordEntity{
		InstID:         instID,
		StudentID:      input.StudentID,
		StudentName:    input.StudentName,
		AssessmentCode: pep3ScaleCode,
		AssessmentName: "PEP-3儿童心理教育评核",
		ScaleVersion:   scoreResult.ScaleVersion,
		BirthDate:      input.ScoreInput.BirthDate,
		AssessmentDate: input.ScoreInput.AssessmentDate,
		AgeYears:       scoreResult.Result.Age.Years,
		AgeMonths:      scoreResult.Result.Age.Months,
		AgeDays:        scoreResult.Result.Age.Days,
		NormAgeMonths:  scoreResult.Result.Age.TotalMonthsForNorm,
		ExaminerID:     examinerID,
		ExaminerName:   examinerName,
		Input:          input.InputSnapshot,
		Result:         scoreResult,
		DataStatus:     scoreResult.DataStatus,
		Remark:         input.Remark,
	})
	if err != nil {
		return model.AssessmentRecordDetailVO{}, err
	}
	return svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
}

func (svc *Service) GetPEP3AssessmentRecord(userID, recordID int64) (model.AssessmentRecordDetailVO, error) {
	if svc.repo == nil {
		return model.AssessmentRecordDetailVO{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.AssessmentRecordDetailVO{}, errors.New("no institution context")
		}
		return model.AssessmentRecordDetailVO{}, err
	}
	return svc.repo.GetAssessmentRecord(context.Background(), instID, recordID)
}

func (svc *Service) PagePEP3AssessmentRecords(userID int64, query model.AssessmentRecordPageQueryDTO) (model.PageResult[model.AssessmentRecordSummaryVO], error) {
	if svc.repo == nil {
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, errors.New("assessment repository is not configured")
	}
	instID, err := svc.repo.FindInstIDByUserID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PageResult[model.AssessmentRecordSummaryVO]{}, errors.New("no institution context")
		}
		return model.PageResult[model.AssessmentRecordSummaryVO]{}, err
	}
	query.QueryModel.AssessmentCode = pep3ScaleCode
	return svc.repo.PageAssessmentRecords(context.Background(), instID, query.QueryModel, query.PageRequestModel.PageIndex, query.PageRequestModel.PageSize)
}

func loadPEP3Engine() (*pep3score.Engine, PEP3ScoreDataInfo, error) {
	pep3EngineOnce.Do(func() {
		pep3Engine, pep3EngineInfo, pep3EngineLoadErr = buildPEP3Engine()
	})
	if pep3EngineLoadErr != nil {
		return nil, PEP3ScoreDataInfo{}, pep3EngineLoadErr
	}
	return pep3Engine, pep3EngineInfo, nil
}

func buildPEP3Engine() (*pep3score.Engine, PEP3ScoreDataInfo, error) {
	dataDir, err := resolvePEP3DataDir()
	if err != nil {
		return nil, PEP3ScoreDataInfo{}, err
	}

	itemPath := filepath.Join(dataDir, pep3ItemBankFile)
	domainPath := filepath.Join(dataDir, pep3DomainMapFile)
	normPath := filepath.Join(dataDir, pep3NormFile)
	correctionPath := filepath.Join(dataDir, pep3CorrectionFile)

	items, err := pep3score.LoadItemDefinitionsFile(itemPath)
	if err != nil {
		return nil, PEP3ScoreDataInfo{}, fmt.Errorf("load PEP-3 item bank: %w", err)
	}
	domains, err := pep3score.LoadDomainDefinitionsFile(domainPath)
	if err != nil {
		return nil, PEP3ScoreDataInfo{}, fmt.Errorf("load PEP-3 domain map: %w", err)
	}

	normPaths := []string{normPath}
	sources, dataStatus := pep3DataSources(dataDir)
	if fileExists(correctionPath) {
		normPaths = append(normPaths, correctionPath)
	}
	norms, err := pep3score.LoadMergedNormRecordsFiles(normPaths...)
	if err != nil {
		return nil, PEP3ScoreDataInfo{}, fmt.Errorf("load PEP-3 norm records: %w", err)
	}
	normDataInfo := pep3NormDataInfoFromRecords(norms)
	engine, err := pep3score.NewEngine(items, domains, norms)
	if err != nil {
		return nil, PEP3ScoreDataInfo{}, fmt.Errorf("build PEP-3 score engine: %w", err)
	}

	return engine, PEP3ScoreDataInfo{
		ScaleCode:        pep3ScaleCode,
		ScaleVersion:     pep3ScaleVersion,
		PEP3NormDataInfo: normDataInfo,
		DataStatus:       dataStatus,
		Sources:          sources,
	}, nil
}

func pep3DefaultNormDataInfo() model.PEP3NormDataInfo {
	return model.PEP3NormDataInfo{
		NormVersion:             pep3ScaleVersion,
		DevelopmentAgeMaxMonths: 92,
		NormAgeBandMaxMonths:    89,
		NormSourcePDF:           pep3NormSourcePDF,
	}
}

func pep3NormDataInfoFromRecords(records []pep3score.NormRecord) model.PEP3NormDataInfo {
	info := pep3DefaultNormDataInfo()
	for _, record := range records {
		if record.SourcePDF != "" {
			info.NormSourcePDF = record.SourcePDF
		}
		switch record.TableType {
		case pep3score.TableDevelopmentAge:
			if record.DevelopmentAgeMonths != nil && *record.DevelopmentAgeMonths > info.DevelopmentAgeMaxMonths {
				info.DevelopmentAgeMaxMonths = *record.DevelopmentAgeMonths
			}
		case pep3score.TablePercentile, pep3score.TableScaledScore:
			if record.AgeMaxMonths != nil && *record.AgeMaxMonths > info.NormAgeBandMaxMonths {
				info.NormAgeBandMaxMonths = *record.AgeMaxMonths
			}
		}
	}
	return info
}

func pep3NormalizeNormDataInfo(info model.PEP3NormDataInfo) model.PEP3NormDataInfo {
	defaultInfo := pep3DefaultNormDataInfo()
	if strings.TrimSpace(info.NormVersion) == "" || strings.TrimSpace(info.NormVersion) == pep3LegacyVersion {
		info.NormVersion = defaultInfo.NormVersion
	}
	if info.DevelopmentAgeMaxMonths == 0 {
		info.DevelopmentAgeMaxMonths = defaultInfo.DevelopmentAgeMaxMonths
	}
	if info.NormAgeBandMaxMonths == 0 {
		info.NormAgeBandMaxMonths = defaultInfo.NormAgeBandMaxMonths
	}
	if strings.TrimSpace(info.NormSourcePDF) == "" {
		info.NormSourcePDF = defaultInfo.NormSourcePDF
	}
	return info
}

func pep3NormalizeScoreDataInfo(info PEP3ScoreDataInfo) PEP3ScoreDataInfo {
	if strings.TrimSpace(info.ScaleCode) == "" {
		info.ScaleCode = pep3ScaleCode
	}
	if strings.TrimSpace(info.ScaleVersion) == "" || strings.TrimSpace(info.ScaleVersion) == pep3LegacyVersion {
		info.ScaleVersion = pep3ScaleVersion
	}
	info.PEP3NormDataInfo = pep3NormalizeNormDataInfo(info.PEP3NormDataInfo)
	return info
}

func resolvePEP3DataDir() (string, error) {
	if raw := os.Getenv("PEP3_DATA_DIR"); raw != "" {
		if err := requirePEP3DataFiles(raw); err != nil {
			return "", err
		}
		return raw, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for dir := cwd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "docs")
		if requirePEP3DataFiles(candidate) == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return "", fmt.Errorf("PEP-3 data files not found; set PEP3_DATA_DIR to the directory containing %s", pep3ItemBankFile)
}

func requirePEP3DataFiles(dir string) error {
	for _, name := range []string{pep3ItemBankFile, pep3DomainMapFile, pep3NormFile} {
		if !fileExists(filepath.Join(dir, name)) {
			return fmt.Errorf("PEP-3 data file %s not found in %s", name, dir)
		}
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
