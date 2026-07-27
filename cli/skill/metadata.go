package skill

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type SkillMetadata struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Revision    string `json:"revision"`
	Scope       string `json:"scope"`
	InstallTime string `json:"install_time"`
	Agent       string `json:"agent,omitempty"`
}

func ReadMetadata(dir string) (*SkillMetadata, error) {
	data, err := os.ReadFile(filepath.Join(dir, ".skill-metadata.json"))
	if err != nil {
		return nil, err
	}
	var meta SkillMetadata
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func WriteMetadata(dir string, meta *SkillMetadata) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".skill-metadata.json"), data, 0644)
}
