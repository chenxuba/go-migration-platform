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
