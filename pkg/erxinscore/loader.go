package erxinscore

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func LoadItemDefinitions(r io.Reader) ([]ItemDefinition, error) {
	var items []ItemDefinition
	if err := json.NewDecoder(r).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode erxin item definitions: %w", err)
	}
	return items, nil
}

func LoadItemDefinitionsFile(path string) ([]ItemDefinition, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return LoadItemDefinitions(file)
}
