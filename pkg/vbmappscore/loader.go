package vbmappscore

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func LoadDomainDefinitions(r io.Reader) ([]DomainDefinition, error) {
	var domains []DomainDefinition
	if err := json.NewDecoder(r).Decode(&domains); err != nil {
		return nil, fmt.Errorf("decode vbmapp domain definitions: %w", err)
	}
	return domains, nil
}

func LoadDomainDefinitionsFile(path string) ([]DomainDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadDomainDefinitions(file)
}

func LoadMilestoneItemDefinitions(r io.Reader) ([]MilestoneItemDefinition, error) {
	var items []MilestoneItemDefinition
	if err := json.NewDecoder(r).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode vbmapp milestone item definitions: %w", err)
	}
	return items, nil
}

func LoadMilestoneItemDefinitionsFile(path string) ([]MilestoneItemDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadMilestoneItemDefinitions(file)
}

func LoadMilestoneScoringRules(r io.Reader) ([]MilestoneScoringRule, error) {
	var rules []MilestoneScoringRule
	if err := json.NewDecoder(r).Decode(&rules); err != nil {
		return nil, fmt.Errorf("decode vbmapp milestone scoring rules: %w", err)
	}
	return rules, nil
}

func LoadMilestoneScoringRulesFile(path string) ([]MilestoneScoringRule, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadMilestoneScoringRules(file)
}

func LoadBarrierDefinitions(r io.Reader) ([]BarrierDefinition, error) {
	var barriers []BarrierDefinition
	if err := json.NewDecoder(r).Decode(&barriers); err != nil {
		return nil, fmt.Errorf("decode vbmapp barrier definitions: %w", err)
	}
	return barriers, nil
}

func LoadBarrierDefinitionsFile(path string) ([]BarrierDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadBarrierDefinitions(file)
}

func LoadTransitionDefinitions(r io.Reader) ([]TransitionDefinition, error) {
	var transitions []TransitionDefinition
	if err := json.NewDecoder(r).Decode(&transitions); err != nil {
		return nil, fmt.Errorf("decode vbmapp transition definitions: %w", err)
	}
	return transitions, nil
}

func LoadTransitionDefinitionsFile(path string) ([]TransitionDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadTransitionDefinitions(file)
}

func LoadResponseFieldTemplates(r io.Reader) (map[string]ResponseFieldTemplate, error) {
	var templates map[string]ResponseFieldTemplate
	if err := json.NewDecoder(r).Decode(&templates); err != nil {
		return nil, fmt.Errorf("decode vbmapp response field templates: %w", err)
	}
	return templates, nil
}

func LoadResponseFieldTemplatesFile(path string) (map[string]ResponseFieldTemplate, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadResponseFieldTemplates(file)
}

func LoadResponseMaterialProfiles(r io.Reader) (map[string]ResponseMaterialProfile, error) {
	var profiles map[string]ResponseMaterialProfile
	if err := json.NewDecoder(r).Decode(&profiles); err != nil {
		return nil, fmt.Errorf("decode vbmapp response material profiles: %w", err)
	}
	return profiles, nil
}

func LoadResponseMaterialProfilesFile(path string) (map[string]ResponseMaterialProfile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadResponseMaterialProfiles(file)
}

func LoadMilestoneResponseSchemas(r io.Reader) ([]MilestoneResponseSchema, error) {
	var schemas []MilestoneResponseSchema
	if err := json.NewDecoder(r).Decode(&schemas); err != nil {
		return nil, fmt.Errorf("decode vbmapp milestone response schemas: %w", err)
	}
	return schemas, nil
}

func LoadMilestoneResponseSchemasFile(path string) ([]MilestoneResponseSchema, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadMilestoneResponseSchemas(file)
}

func LoadBarrierResponseSchemas(r io.Reader) ([]BarrierResponseSchema, error) {
	var schemas []BarrierResponseSchema
	if err := json.NewDecoder(r).Decode(&schemas); err != nil {
		return nil, fmt.Errorf("decode vbmapp barrier response schemas: %w", err)
	}
	return schemas, nil
}

func LoadBarrierResponseSchemasFile(path string) ([]BarrierResponseSchema, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadBarrierResponseSchemas(file)
}

func LoadTransitionResponseSchemas(r io.Reader) ([]TransitionResponseSchema, error) {
	var schemas []TransitionResponseSchema
	if err := json.NewDecoder(r).Decode(&schemas); err != nil {
		return nil, fmt.Errorf("decode vbmapp transition response schemas: %w", err)
	}
	return schemas, nil
}

func LoadTransitionResponseSchemasFile(path string) ([]TransitionResponseSchema, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadTransitionResponseSchemas(file)
}
