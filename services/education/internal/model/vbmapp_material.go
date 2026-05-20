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
