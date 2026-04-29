package pep3score

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func LoadItemDefinitions(r io.Reader) ([]ItemDefinition, error) {
	var items []ItemDefinition
	if err := json.NewDecoder(r).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode item definitions: %w", err)
	}
	return items, nil
}

func LoadItemDefinitionsFile(path string) ([]ItemDefinition, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadItemDefinitions(f)
}

func LoadDomainDefinitions(r io.Reader) ([]DomainDefinition, error) {
	var domains []DomainDefinition
	if err := json.NewDecoder(r).Decode(&domains); err != nil {
		return nil, fmt.Errorf("decode domain definitions: %w", err)
	}
	return domains, nil
}

func LoadDomainDefinitionsFile(path string) ([]DomainDefinition, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadDomainDefinitions(f)
}

func LoadNormRecords(r io.Reader) ([]NormRecord, error) {
	var records []NormRecord
	if err := json.NewDecoder(r).Decode(&records); err != nil {
		return nil, fmt.Errorf("decode norm records: %w", err)
	}
	return records, nil
}

func LoadNormRecordsFile(path string) ([]NormRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadNormRecords(f)
}

func LoadMergedNormRecordsFiles(paths ...string) ([]NormRecord, error) {
	var groups [][]NormRecord
	for _, path := range paths {
		records, err := LoadNormRecordsFile(path)
		if err != nil {
			return nil, err
		}
		groups = append(groups, records)
	}
	return MergeNormRecords(groups...), nil
}

func MergeNormRecords(groups ...[]NormRecord) []NormRecord {
	merged := make([]NormRecord, 0)
	indexByKey := make(map[string]int)
	for _, group := range groups {
		for _, record := range group {
			key := normRecordKey(record)
			if key == "" {
				merged = append(merged, record)
				continue
			}
			if idx, ok := indexByKey[key]; ok {
				merged[idx] = record
				continue
			}
			indexByKey[key] = len(merged)
			merged = append(merged, record)
		}
	}
	return merged
}

func normRecordKey(record NormRecord) string {
	parts := []string{record.TableType}
	switch record.TableType {
	case TableDevelopmentAge:
		parts = append(parts, record.ScaleCode, ptrIntKey(record.RawScoreMin), ptrIntKey(record.RawScoreMax))
	case TablePercentile, TableScaledScore:
		parts = append(parts, ptrIntKey(record.AgeMinMonths), ptrIntKey(record.AgeMaxMonths), record.ScaleCode, ptrIntKey(record.RawScore))
	case TableComposite:
		parts = append(parts, record.CompositeCode, ptrIntKey(record.StandardScoreSum), record.StandardScoreSumLabel)
	default:
		return ""
	}
	return strings.Join(parts, "|")
}

func ptrIntKey(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}
