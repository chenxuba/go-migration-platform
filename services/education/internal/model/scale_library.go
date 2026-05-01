package model

type ScaleLibraryQuery struct {
	Keyword  string `json:"keyword,omitempty"`
	Category string `json:"category,omitempty"`
	Scenario string `json:"scenario,omitempty"`
	Status   string `json:"status,omitempty"`
	AgeScope string `json:"ageScope,omitempty"`
	Duration string `json:"duration,omitempty"`
}

type ScaleLibraryTextResource struct {
	ID      int64  `json:"id"`
	ScaleID int64  `json:"scaleId"`
	Content string `json:"content"`
	Sort    int    `json:"sort"`
}

type ScaleLibraryItem struct {
	ID                 int64                      `json:"id"`
	Name               string                     `json:"name"`
	Code               string                     `json:"code"`
	Category           string                     `json:"category"`
	Scenario           string                     `json:"scenario"`
	AgeRange           string                     `json:"ageRange"`
	AgeMinMonths       int                        `json:"ageMinMonths"`
	AgeMaxMonths       int                        `json:"ageMaxMonths"`
	Duration           string                     `json:"duration"`
	DurationMinMinutes int                        `json:"durationMinMinutes"`
	DurationMaxMinutes int                        `json:"durationMaxMinutes"`
	CurrentVersion     string                     `json:"currentVersion"`
	ItemCount          int                        `json:"itemCount"`
	DomainCount        int                        `json:"domainCount"`
	InstitutionCount   int                        `json:"institutionCount"`
	MonthUsage         int                        `json:"monthUsage"`
	UsageCount         int                        `json:"usageCount"`
	LatestUse          string                     `json:"latestUse"`
	DataStatus         string                     `json:"dataStatus"`
	Status             string                     `json:"status"`
	StatusText         string                     `json:"statusText"`
	UpdatedAt          string                     `json:"updatedAt"`
	Summary            string                     `json:"summary"`
	PosterURL          string                     `json:"posterUrl"`
	ExecutionEntry     string                     `json:"executionEntry"`
	APIPackage         string                     `json:"apiPackage"`
	References         []ScaleLibraryTextResource `json:"references"`
	Acknowledgements   []ScaleLibraryTextResource `json:"acknowledgements"`
	AuthReserved       bool                       `json:"authReserved"`
	AuthActionEnabled  bool                       `json:"authActionEnabled"`
}

type ScaleLibrarySummary struct {
	Total         int `json:"total"`
	Available     int `json:"available"`
	Unavailable   int `json:"unavailable"`
	MonthUsage    int `json:"monthUsage"`
	UsageCount    int `json:"usageCount"`
	ReservedAuths int `json:"reservedAuths"`
}

type ScaleLibraryFilterOptions struct {
	Categories     []string       `json:"categories"`
	CategoryCounts map[string]int `json:"categoryCounts"`
	Scenarios      []string       `json:"scenarios"`
	Statuses       []string       `json:"statuses"`
	AgeScopes      []string       `json:"ageScopes"`
	Durations      []string       `json:"durations"`
}

type ScaleLibraryVO struct {
	Items         []ScaleLibraryItem        `json:"items"`
	Summary       ScaleLibrarySummary       `json:"summary"`
	FilterOptions ScaleLibraryFilterOptions `json:"filterOptions"`
}
