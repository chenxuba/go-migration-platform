package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"go-migration-platform/pkg/pep3score"
	"go-migration-platform/services/education/internal/model"
	"go-migration-platform/services/education/internal/repository"
)

var (
	pep3StaticDataRepoMu  sync.RWMutex
	pep3StaticDataRepo    *repository.Repository
	errPEP3StaticDataMiss = errors.New("PEP-3 static data is not seeded")
)

type pep3StaticData struct {
	itemRows     []json.RawMessage
	formItems    []pep3FormItemDefinition
	scoreItems   []pep3score.ItemDefinition
	domains      []pep3score.DomainDefinition
	recordFields map[int][]model.PEP3ItemRecordField
	norms        []pep3score.NormRecord
	sources      []string
	dataStatus   string
}

func configurePEP3StaticDataRepository(repo *repository.Repository) {
	pep3StaticDataRepoMu.Lock()
	defer pep3StaticDataRepoMu.Unlock()
	pep3StaticDataRepo = repo
}

func currentPEP3StaticDataRepository() *repository.Repository {
	pep3StaticDataRepoMu.RLock()
	defer pep3StaticDataRepoMu.RUnlock()
	return pep3StaticDataRepo
}

func (svc *Service) EnsurePEP3ScaleData(ctx context.Context) error {
	if svc == nil || svc.repo == nil {
		return nil
	}
	forceReseed := os.Getenv("PEP3_STATIC_DATA_RESEED") == "1"
	if !forceReseed {
		hasData, err := svc.repo.HasAssessmentScaleStaticData(ctx, pep3ScaleCode, pep3ScaleVersion)
		if err != nil {
			return err
		}
		if hasData {
			return nil
		}
	}
	data, err := loadPEP3StaticDataFromFiles()
	if err != nil {
		return err
	}
	if err := svc.repo.ReplaceAssessmentScaleStaticData(ctx, pep3StaticDataEntity(data), 0); err != nil {
		return err
	}
	pep3EngineOnce = sync.Once{}
	pep3Engine = nil
	pep3EngineInfo = PEP3ScoreDataInfo{}
	pep3EngineLoadErr = nil
	return nil
}

func loadPEP3StaticData() (pep3StaticData, error) {
	if repo := currentPEP3StaticDataRepository(); repo != nil {
		data, err := loadPEP3StaticDataFromDB(context.Background(), repo)
		if err == nil {
			return data, nil
		}
		if !isPEP3StaticDataFallbackError(err) {
			return pep3StaticData{}, err
		}
	}
	return loadPEP3StaticDataFromFiles()
}

func isPEP3StaticDataFallbackError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, errPEP3StaticDataMiss) ||
		errors.Is(err, sql.ErrNoRows) ||
		strings.Contains(err.Error(), "database is closed")
}

func loadPEP3StaticDataFromFiles() (pep3StaticData, error) {
	dataDir, err := resolvePEP3DataDir()
	if err != nil {
		return pep3StaticData{}, err
	}
	formItems, err := loadPEP3FormItems(dataDir)
	if err != nil {
		return pep3StaticData{}, err
	}
	domains, err := pep3score.LoadDomainDefinitionsFile(filepath.Join(dataDir, pep3DomainMapFile))
	if err != nil {
		return pep3StaticData{}, fmt.Errorf("load PEP-3 domain map: %w", err)
	}
	normPaths := []string{filepath.Join(dataDir, pep3NormFile)}
	if correctionPath := filepath.Join(dataDir, pep3CorrectionFile); fileExists(correctionPath) {
		normPaths = append(normPaths, correctionPath)
	}
	norms, err := pep3score.LoadMergedNormRecordsFiles(normPaths...)
	if err != nil {
		return pep3StaticData{}, fmt.Errorf("load PEP-3 norm records: %w", err)
	}
	itemRows, err := marshalPEP3StaticItems(formItems)
	if err != nil {
		return pep3StaticData{}, err
	}
	sources, dataStatus := pep3DataSources(dataDir)
	return pep3StaticData{
		itemRows:     itemRows,
		formItems:    formItems,
		scoreItems:   pep3ScoreItemsFromFormItems(formItems),
		domains:      domains,
		recordFields: pep3AllItemRecordFields(),
		norms:        norms,
		sources:      sources,
		dataStatus:   dataStatus,
	}, nil
}

func loadPEP3StaticDataFromDB(ctx context.Context, repo *repository.Repository) (pep3StaticData, error) {
	dataset, err := repo.GetAssessmentScaleDataset(ctx, pep3ScaleCode, pep3ScaleVersion)
	if err != nil {
		return pep3StaticData{}, err
	}
	itemRows, err := repo.ListAssessmentScaleItems(ctx, pep3ScaleCode, pep3ScaleVersion)
	if err != nil {
		return pep3StaticData{}, err
	}
	domainRows, err := repo.ListAssessmentScaleDomains(ctx, pep3ScaleCode, pep3ScaleVersion)
	if err != nil {
		return pep3StaticData{}, err
	}
	normRows, err := repo.ListAssessmentScaleNormRecords(ctx, pep3ScaleCode, pep3ScaleVersion)
	if err != nil {
		return pep3StaticData{}, err
	}
	recordFieldRows, err := repo.ListAssessmentScaleItemRecordFields(ctx, pep3ScaleCode, pep3ScaleVersion)
	if err != nil && !strings.Contains(err.Error(), "assessment_scale_item_record_field") {
		return pep3StaticData{}, err
	}
	if len(itemRows) == 0 || len(domainRows) == 0 || len(normRows) == 0 {
		return pep3StaticData{}, errPEP3StaticDataMiss
	}

	rawItems := make([]json.RawMessage, 0, len(itemRows))
	formItems := make([]pep3FormItemDefinition, 0, len(itemRows))
	for _, row := range itemRows {
		var item pep3FormItemDefinition
		if err := json.Unmarshal(row.Raw, &item); err != nil {
			return pep3StaticData{}, fmt.Errorf("decode PEP-3 DB item %d: %w", row.ItemNo, err)
		}
		rawItems = append(rawItems, append(json.RawMessage(nil), row.Raw...))
		formItems = append(formItems, item)
	}
	sort.Slice(formItems, func(i, j int) bool { return formItems[i].ItemNo < formItems[j].ItemNo })

	domains := make([]pep3score.DomainDefinition, 0, len(domainRows))
	for _, row := range domainRows {
		var domain pep3score.DomainDefinition
		if err := json.Unmarshal(row.Raw, &domain); err != nil {
			return pep3StaticData{}, fmt.Errorf("decode PEP-3 DB domain %s: %w", row.DomainCode, err)
		}
		domains = append(domains, domain)
	}

	norms := make([]pep3score.NormRecord, 0, len(normRows))
	for _, row := range normRows {
		var record pep3score.NormRecord
		if err := json.Unmarshal(row.Raw, &record); err != nil {
			return pep3StaticData{}, fmt.Errorf("decode PEP-3 DB norm record %s: %w", row.RecordKey, err)
		}
		norms = append(norms, record)
	}

	recordFields := pep3RecordFieldsFromRows(recordFieldRows)
	if len(recordFields) == 0 {
		recordFields = pep3AllItemRecordFields()
	}

	return pep3StaticData{
		itemRows:     rawItems,
		formItems:    formItems,
		scoreItems:   pep3ScoreItemsFromFormItems(formItems),
		domains:      domains,
		recordFields: recordFields,
		norms:        norms,
		sources:      dataset.Sources,
		dataStatus:   dataset.DataStatus,
	}, nil
}

func pep3StaticDataEntity(data pep3StaticData) repository.AssessmentScaleStaticDataEntity {
	items := make([]repository.AssessmentScaleItemEntity, 0, len(data.formItems))
	for idx, item := range data.formItems {
		raw := json.RawMessage(nil)
		if idx < len(data.itemRows) {
			raw = data.itemRows[idx]
		}
		if len(raw) == 0 {
			if marshaled, err := json.Marshal(item); err == nil {
				raw = marshaled
			}
		}
		items = append(items, repository.AssessmentScaleItemEntity{ItemNo: item.ItemNo, Raw: raw})
	}
	domains := make([]repository.AssessmentScaleDomainEntity, 0, len(data.domains))
	for idx, domain := range data.domains {
		raw, _ := json.Marshal(domain)
		domains = append(domains, repository.AssessmentScaleDomainEntity{
			DomainCode: domain.ScaleCode,
			SortNo:     idx + 1,
			Raw:        raw,
		})
	}
	recordFields := make([]repository.AssessmentScaleItemRecordFieldEntity, 0)
	for itemNo, fields := range data.recordFields {
		for idx, field := range fields {
			if strings.TrimSpace(field.Key) == "" {
				continue
			}
			raw, _ := json.Marshal(field)
			recordFields = append(recordFields, repository.AssessmentScaleItemRecordFieldEntity{
				ItemNo:   itemNo,
				FieldKey: field.Key,
				SortNo:   idx + 1,
				Raw:      raw,
			})
		}
	}
	norms := make([]repository.AssessmentScaleNormRecordEntity, 0, len(data.norms))
	for idx, record := range data.norms {
		raw, _ := json.Marshal(record)
		norms = append(norms, repository.AssessmentScaleNormRecordEntity{
			RecordKey: fmt.Sprintf("%06d", idx+1),
			SortNo:    idx + 1,
			Raw:       raw,
		})
	}
	return repository.AssessmentScaleStaticDataEntity{
		Dataset: repository.AssessmentScaleDatasetEntity{
			ScaleCode:    pep3ScaleCode,
			ScaleVersion: pep3ScaleVersion,
			DataStatus:   data.dataStatus,
			Sources:      data.sources,
		},
		Items:        items,
		Domains:      domains,
		RecordFields: recordFields,
		NormRecords:  norms,
	}
}

func pep3RecordFieldsFromRows(rows []repository.AssessmentScaleItemRecordFieldEntity) map[int][]model.PEP3ItemRecordField {
	out := make(map[int][]model.PEP3ItemRecordField)
	for _, row := range rows {
		if row.ItemNo <= 0 || len(row.Raw) == 0 {
			continue
		}
		var field model.PEP3ItemRecordField
		if err := json.Unmarshal(row.Raw, &field); err != nil {
			continue
		}
		if strings.TrimSpace(field.Key) == "" {
			field.Key = strings.TrimSpace(row.FieldKey)
		}
		if strings.TrimSpace(field.Key) == "" {
			continue
		}
		out[row.ItemNo] = append(out[row.ItemNo], field)
	}
	return out
}

func marshalPEP3StaticItems(items []pep3FormItemDefinition) ([]json.RawMessage, error) {
	out := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		raw, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("marshal PEP-3 item %d: %w", item.ItemNo, err)
		}
		out = append(out, raw)
	}
	return out, nil
}

func pep3ScoreItemsFromFormItems(items []pep3FormItemDefinition) []pep3score.ItemDefinition {
	out := make([]pep3score.ItemDefinition, 0, len(items))
	for _, item := range items {
		if item.ItemNo <= 0 {
			continue
		}
		out = append(out, pep3score.ItemDefinition{
			ItemNo:    item.ItemNo,
			ItemTitle: item.ItemTitle,
			ScaleCode: strings.ToUpper(strings.TrimSpace(item.DomainCode)),
			ScaleName: item.Domain,
		})
	}
	return out
}

func unmarshalPEP3BookletItems(rows []json.RawMessage) ([]pep3BookletItemDefinition, error) {
	items := make([]pep3BookletItemDefinition, 0, len(rows))
	for idx, raw := range rows {
		var item pep3BookletItemDefinition
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, fmt.Errorf("decode PEP-3 booklet item row %s: %w", strconv.Itoa(idx+1), err)
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ItemNo < items[j].ItemNo })
	return items, nil
}
