package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Persistence interface {
	Load() (map[string]CacheEntry, error)
	Save(data map[string]CacheEntry) error
}

type FilePersistence struct {
	filePath string
}

func newFilePersistence(filePath string) *FilePersistence {
	return &FilePersistence{filePath: filePath}
}

func (fp *FilePersistence) Load() (map[string]CacheEntry, error) {
	file, err := os.Open(fp.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache data: %w", err)
	}

	var cacheData map[string]CacheEntry
	err = json.Unmarshal(data, &cacheData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	return cacheData, nil
}

func (p *FilePersistence) Save(data map[string]CacheEntry) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	file, err := os.Create(p.filePath)
	if err != nil {
		return fmt.Errorf("failed to create cache file: %w", err)
	}
	defer file.Close()

	_, err = io.WriteString(file, string(jsonData))
	if err != nil {
		return fmt.Errorf("failed to write cache data: %w", err)
	}

	return nil
}
