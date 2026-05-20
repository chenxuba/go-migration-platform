package model

import "time"

type VBMAPPMaterialProfile struct {
	ProfileID         string               `json:"profileId"`
	Label             string               `json:"label"`
	SourceLogic       string               `json:"sourceLogic,omitempty"`
	SuggestedTypes    []string             `json:"suggestedTypes,omitempty"`
	PreparationChecks []string             `json:"preparationChecks,omitempty"`
	Materials         []VBMAPPMaterialItem `json:"materials,omitempty"`
}

type VBMAPPMaterialItem struct {
	ID           int64      `json:"id,omitempty"`
	ScaleVersion string     `json:"scaleVersion,omitempty"`
	LibraryScope string     `json:"libraryScope,omitempty"`
	InstID       int64      `json:"instId,omitempty"`
	ProfileID    string     `json:"profileId"`
	MaterialCode string     `json:"materialCode,omitempty"`
	MaterialName string     `json:"materialName"`
	MaterialType string     `json:"materialType,omitempty"`
	SortNo       int        `json:"sortNo,omitempty"`
	Status       string     `json:"status,omitempty"`
	CreatedTime  *time.Time `json:"createdTime,omitempty"`
	UpdatedTime  *time.Time `json:"updatedTime,omitempty"`
}

type VBMAPPMaterialCatalog struct {
	ScaleCode      string                      `json:"scaleCode"`
	ScaleVersion   string                      `json:"scaleVersion"`
	AssessmentName string                      `json:"assessmentName"`
	DataStatus     string                      `json:"dataStatus,omitempty"`
	Sources        []string                    `json:"sources,omitempty"`
	Items          []VBMAPPMaterialCatalogItem `json:"items"`
}

type VBMAPPMaterialCatalogItem struct {
	ModuleCode           string                     `json:"moduleCode"`
	ItemCode             string                     `json:"itemCode"`
	Label                string                     `json:"label,omitempty"`
	Title                string                     `json:"title,omitempty"`
	DomainCode           string                     `json:"domainCode,omitempty"`
	DomainName           string                     `json:"domainName,omitempty"`
	Level                int                        `json:"level,omitempty"`
	AgeBand              string                     `json:"ageBand,omitempty"`
	AssessmentMode       string                     `json:"assessmentMode,omitempty"`
	MaterialProfileID    string                     `json:"materialProfileId"`
	MaterialProfileLabel string                     `json:"materialProfileLabel,omitempty"`
	WhyRecord            string                     `json:"whyRecord,omitempty"`
	QualityChecks        []string                   `json:"qualityChecks,omitempty"`
	PreparationChecks    []string                   `json:"preparationChecks,omitempty"`
	SuggestedTypes       []string                   `json:"suggestedTypes,omitempty"`
	RecommendedMaterials []VBMAPPMaterialSuggestion `json:"recommendedMaterials,omitempty"`
	QuickPicksByField    map[string][]string        `json:"quickPicksByField,omitempty"`
}

type VBMAPPMaterialSuggestion struct {
	MaterialCode string `json:"materialCode,omitempty"`
	MaterialName string `json:"materialName"`
	MaterialType string `json:"materialType,omitempty"`
}
