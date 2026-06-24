package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func ReadJSON(path string, value any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read json file: %w", err)
	}
	if err := json.Unmarshal(raw, value); err != nil {
		return fmt.Errorf("parse json file: %w", err)
	}
	return nil
}

func WriteJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create json file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return fmt.Errorf("write json file: %w", err)
	}

	return nil
}
