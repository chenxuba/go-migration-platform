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
	"strings"
	"sync"

	"go-migration-platform/pkg/autismdevscore"
	"go-migration-platform/services/education/internal/repository"
)

var (
	autismDevStaticDataRepoMu  sync.RWMutex
	autismDevStaticDataRepo    *repository.Repository
	errAutismDevStaticDataMiss = errors.New("AutismDev static data is not seeded")
)

func configureAutismDevStaticDataRepository(repo *repository.Repository) {
	autismDevStaticDataRepoMu.Lock()
	defer autismDevStaticDataRepoMu.Unlock()
	autismDevStaticDataRepo = repo
}

func currentAutismDevStaticDataRepository() *repository.Repository {
	autismDevStaticDataRepoMu.RLock()
	defer autismDevStaticDataRepoMu.RUnlock()
	return autismDevStaticDataRepo
}

func (svc *Service) EnsureAutismDevScaleData(ctx context.Context) error {
	if svc == nil || svc.repo == nil {
		return nil
	}
	forceReseed := os.Getenv("AUTISMDEV_STATIC_DATA_RESEED") == "1"
	if !forceReseed {
		hasData, err := hasAutismDevScaleData(ctx, svc.repo)
		if err != nil {
			return err
		}
		if hasData {
			dataset, err := svc.repo.GetAssessmentScaleDataset(ctx, autismDevScaleCode, autismDevScaleVersion)
			if err == nil {
				expectedSources, sourceErr := autismDevExpectedStaticDataSources()
				if sourceErr != nil || sameStringSlice(dataset.Sources, expectedSources) {
					return nil
				}
			}
		}
	}

	dataDir, err := resolveAutismDevDataDir()
	if err != nil {
		return err
	}
	data, err := loadAutismDevStaticDataFromFiles(dataDir)
	if err != nil {
		return err
	}
	if err := svc.repo.ReplaceAssessmentScaleStaticData(ctx, autismDevStaticDataEntity(data), 0); err != nil {
		return err
	}
	resetAutismDevFallbackStaticDataCache()
	return nil
}

func hasAutismDevScaleData(ctx context.Context, repo *repository.Repository) (bool, error) {
	if _, err := repo.GetAssessmentScaleDataset(ctx, autismDevScaleCode, autismDevScaleVersion); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	itemRows, err := repo.ListAssessmentScaleItems(ctx, autismDevScaleCode, autismDevScaleVersion)
	if err != nil {
		return false, err
	}
	domainRows, err := repo.ListAssessmentScaleDomains(ctx, autismDevScaleCode, autismDevScaleVersion)
	if err != nil {
		return false, err
	}
	return len(itemRows) == autismdevscore.ExpectedItemDefinition && len(domainRows) == len(autismdevscore.DomainOrder), nil
}

func loadAutismDevStaticDataFromConfiguredDB() (autismDevStaticData, bool, error) {
	repo := currentAutismDevStaticDataRepository()
	if repo == nil {
		return autismDevStaticData{}, false, nil
	}
	data, err := loadAutismDevStaticDataFromDB(context.Background(), repo)
	if err == nil {
		return data, true, nil
	}
	if isAutismDevStaticDataFallbackError(err) {
		return autismDevStaticData{}, false, nil
	}
	return autismDevStaticData{}, true, err
}

func loadAutismDevStaticDataFromDB(ctx context.Context, repo *repository.Repository) (autismDevStaticData, error) {
	dataset, err := repo.GetAssessmentScaleDataset(ctx, autismDevScaleCode, autismDevScaleVersion)
	if err != nil {
		return autismDevStaticData{}, err
	}
	itemRows, err := repo.ListAssessmentScaleItems(ctx, autismDevScaleCode, autismDevScaleVersion)
	if err != nil {
		return autismDevStaticData{}, err
	}
	domainRows, err := repo.ListAssessmentScaleDomains(ctx, autismDevScaleCode, autismDevScaleVersion)
	if err != nil {
		return autismDevStaticData{}, err
	}
	if len(itemRows) == 0 || len(domainRows) == 0 {
		return autismDevStaticData{}, errAutismDevStaticDataMiss
	}

	items := make([]autismdevscore.ItemDefinition, 0, len(itemRows))
	for _, row := range itemRows {
		var item autismdevscore.ItemDefinition
		if err := json.Unmarshal(row.Raw, &item); err != nil {
			return autismDevStaticData{}, fmt.Errorf("decode AutismDev DB item %d: %w", row.ItemNo, err)
		}
		if item.ItemNo == 0 {
			item.ItemNo = row.ItemNo
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ItemNo < items[j].ItemNo })

	domains := make([]autismDevDomainDefinition, 0, len(domainRows))
	for _, row := range domainRows {
		var domain autismDevDomainDefinition
		if err := json.Unmarshal(row.Raw, &domain); err != nil {
			return autismDevStaticData{}, fmt.Errorf("decode AutismDev DB domain %s: %w", row.DomainCode, err)
		}
		if strings.TrimSpace(domain.ScaleCode) == "" {
			domain.ScaleCode = row.DomainCode
		}
		if domain.SortNo == 0 {
			domain.SortNo = row.SortNo
		}
		domains = append(domains, domain)
	}
	sort.Slice(domains, func(i, j int) bool { return domains[i].SortNo < domains[j].SortNo })

	metadata := autismDevScaleMetadata{
		ScaleCode:    nonEmptyString(dataset.ScaleCode, autismDevScaleCode),
		ScaleName:    autismDevAssessmentName,
		ScaleVersion: nonEmptyString(dataset.ScaleVersion, autismDevScaleVersion),
		ItemCount:    len(items),
		DomainCount:  len(domains),
		DataStatus:   strings.TrimSpace(dataset.DataStatus),
	}
	if metadataRaw := strings.TrimSpace(string(dataset.Metadata)); metadataRaw != "" {
		if err := json.Unmarshal([]byte(metadataRaw), &metadata); err != nil {
			return autismDevStaticData{}, fmt.Errorf("decode AutismDev DB metadata: %w", err)
		}
	}
	metadata.ScaleCode = nonEmptyString(metadata.ScaleCode, autismDevScaleCode)
	metadata.ScaleName = nonEmptyString(metadata.ScaleName, autismDevAssessmentName)
	metadata.ScaleVersion = nonEmptyString(metadata.ScaleVersion, autismDevScaleVersion)
	if metadata.ItemCount == 0 {
		metadata.ItemCount = len(items)
	}
	if metadata.DomainCount == 0 {
		metadata.DomainCount = len(domains)
	}
	dataStatus := strings.TrimSpace(dataset.DataStatus)
	if dataStatus == "" {
		dataStatus = strings.TrimSpace(metadata.DataStatus)
	}
	if dataStatus == "" {
		dataStatus = autismDevDraftDataStatus
	}
	metadata.DataStatus = dataStatus

	return autismDevStaticData{
		metadata:   metadata,
		items:      items,
		domains:    domains,
		sources:    append([]string(nil), dataset.Sources...),
		dataStatus: dataStatus,
	}, nil
}

func loadAutismDevTemplateItemFromConfiguredDB(itemNo int) (autismdevscore.ItemDefinition, bool, error) {
	repo := currentAutismDevStaticDataRepository()
	if repo == nil {
		return autismdevscore.ItemDefinition{}, false, nil
	}
	row, err := repo.GetAssessmentScaleItem(context.Background(), autismDevScaleCode, autismDevScaleVersion, itemNo)
	if err != nil {
		if isAutismDevStaticDataFallbackError(err) {
			return autismdevscore.ItemDefinition{}, false, nil
		}
		return autismdevscore.ItemDefinition{}, true, err
	}
	var item autismdevscore.ItemDefinition
	if err := json.Unmarshal(row.Raw, &item); err != nil {
		return autismdevscore.ItemDefinition{}, true, fmt.Errorf("decode AutismDev DB item %d: %w", row.ItemNo, err)
	}
	if item.ItemNo == 0 {
		item.ItemNo = row.ItemNo
	}
	return item, true, nil
}

func autismDevStaticDataEntity(data autismDevStaticData) repository.AssessmentScaleStaticDataEntity {
	items := make([]repository.AssessmentScaleItemEntity, 0, len(data.items))
	for _, item := range data.items {
		raw, _ := json.Marshal(item)
		items = append(items, repository.AssessmentScaleItemEntity{
			ItemNo: item.ItemNo,
			Raw:    raw,
		})
	}
	domains := make([]repository.AssessmentScaleDomainEntity, 0, len(data.domains))
	for idx, domain := range data.domains {
		raw, _ := json.Marshal(domain)
		sortNo := domain.SortNo
		if sortNo == 0 {
			sortNo = idx + 1
		}
		domains = append(domains, repository.AssessmentScaleDomainEntity{
			DomainCode: domain.ScaleCode,
			SortNo:     sortNo,
			Raw:        raw,
		})
	}
	metadataRaw, _ := json.Marshal(data.metadata)
	return repository.AssessmentScaleStaticDataEntity{
		Dataset: repository.AssessmentScaleDatasetEntity{
			ScaleCode:    autismDevScaleCode,
			ScaleVersion: autismDevScaleVersion,
			DataStatus:   data.dataStatus,
			Sources:      append([]string(nil), data.sources...),
			Metadata:     metadataRaw,
		},
		Items:   items,
		Domains: domains,
	}
}

func autismDevExpectedStaticDataSources() ([]string, error) {
	dataDir, err := resolveAutismDevDataDir()
	if err != nil {
		return nil, err
	}
	var metadata autismDevScaleMetadata
	if err := loadAutismDevJSONFile(filepath.Join(dataDir, autismDevMetadataFile), &metadata); err != nil {
		return nil, err
	}
	sources, _ := autismDevDataSources(metadata.DataStatus)
	return sources, nil
}

func isAutismDevStaticDataFallbackError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, errAutismDevStaticDataMiss) ||
		errors.Is(err, sql.ErrNoRows) ||
		strings.Contains(err.Error(), "database is closed")
}
