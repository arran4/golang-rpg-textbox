package skill

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Install(source string, scope string, agent string, args []string) error {
	if source == "" {
		return fmt.Errorf("source is required")
	}

	name := ""
	if len(args) > 0 {
		name = args[0]
	} else {
		parts := strings.Split(source, "/")
		name = parts[len(parts)-1]
		if name == "" && len(parts) > 1 {
			name = parts[len(parts)-2]
		}
	}

	targetDir, err := GetTargetDirectory(scope, agent, name)
	if err != nil {
		return err
	}

	fmt.Printf("Installing skill %s from %s to %s\n", name, source, targetDir)

	if _, err := os.Stat(targetDir); err == nil {
		fmt.Printf("Removing existing skill directory...\n")
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	revision, err := FetchSkill(source, targetDir)
	if err != nil {
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("failed to fetch skill: %w", err)
	}

	if _, err := os.Stat(filepath.Join(targetDir, "SKILL.md")); os.IsNotExist(err) {
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("invalid skill: SKILL.md not found")
	}

	meta := &SkillMetadata{
		Name:        name,
		Source:      source,
		Revision:    revision,
		Scope:       scope,
		Agent:       agent,
		InstallTime: time.Now().Format(time.RFC3339),
	}
	if err := WriteMetadata(targetDir, meta); err != nil {
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	fmt.Printf("Successfully installed skill %s\n", name)
	return nil
}

func Update(name string, all bool, force bool) error {
	if !all && name == "" {
		return fmt.Errorf("skill name is required unless --all is specified")
	}

	skillsToUpdate := []string{}

	scopes := []string{"user", "project"}
	agents := []string{"", "cursor", "copilot"}

	if all {
		for _, scope := range scopes {
			for _, agent := range agents {
				baseDir, err := GetTargetDirectory(scope, agent, "")
				if err != nil {
					continue
				}
				entries, err := os.ReadDir(baseDir)
				if err != nil {
					continue
				}
				for _, entry := range entries {
					if entry.IsDir() {
						skillsToUpdate = append(skillsToUpdate, entry.Name())
					}
				}
			}
		}
	} else {
		skillsToUpdate = append(skillsToUpdate, name)
	}

	if len(skillsToUpdate) == 0 {
		fmt.Println("No skills found to update")
		return nil
	}

	// Deduplicate
	uniqueSkills := make(map[string]bool)
	for _, s := range skillsToUpdate {
		uniqueSkills[s] = true
	}

	for s := range uniqueSkills {
		fmt.Printf("Checking for updates for %s...\n", s)

		var targetDir string
		var meta *SkillMetadata
		var err error

		// Find the skill
		for _, scope := range scopes {
			for _, agent := range agents {
				dir, err2 := GetTargetDirectory(scope, agent, s)
				if err2 != nil {
					continue
				}
				if m, err3 := ReadMetadata(dir); err3 == nil {
					targetDir = dir
					meta = m
					break
				}
			}
			if targetDir != "" {
				break
			}
		}

		if targetDir == "" {
			if !all {
				return fmt.Errorf("skill %s not found or missing metadata", s)
			}
			continue
		}

		fmt.Printf("Updating %s from %s...\n", s, meta.Source)

		tempDir, err := os.MkdirTemp("", "rpgtextbox-skill-update-*")
		if err != nil {
			return err
		}

		revision, err := FetchSkill(meta.Source, tempDir)
		if err != nil {
			_ = os.RemoveAll(tempDir)
			fmt.Printf("Failed to check for updates for %s: %v\n", s, err)
			continue
		}

		if revision == meta.Revision && !force {
			fmt.Printf("Skill %s is already up to date (revision: %s)\n", s, meta.Revision)
			_ = os.RemoveAll(tempDir)
			continue
		}

		if _, err := os.Stat(filepath.Join(tempDir, "SKILL.md")); os.IsNotExist(err) {
			_ = os.RemoveAll(tempDir)
			fmt.Printf("Failed to update %s: new version is missing SKILL.md\n", s)
			continue
		}

		if err := os.RemoveAll(targetDir); err != nil {
			fmt.Printf("Failed to remove old skill %s: %v\n", s, err)
			_ = os.RemoveAll(tempDir)
			continue
		}
		if err := os.Rename(tempDir, targetDir); err != nil {
			fmt.Printf("Failed to rename updated skill %s: %v\n", s, err)
			_ = os.RemoveAll(tempDir)
			continue
		}

		meta.Revision = revision
		meta.InstallTime = time.Now().Format(time.RFC3339)
		if err := WriteMetadata(targetDir, meta); err != nil {
			fmt.Printf("Failed to write metadata for skill %s: %v\n", s, err)
			continue
		}

		fmt.Printf("Successfully updated skill %s to revision %s\n", s, revision)
	}

	return nil
}

func Remove(name string, force bool) error {
	if name == "" {
		return fmt.Errorf("skill name is required")
	}

	scopes := []string{"user", "project"}
	agents := []string{"", "cursor", "copilot"}
	removed := false

	for _, scope := range scopes {
		for _, agent := range agents {
			targetDir, err := GetTargetDirectory(scope, agent, name)
			if err != nil {
				continue
			}
			if _, err := os.Stat(targetDir); err == nil {
				if _, err := os.Stat(filepath.Join(targetDir, "SKILL.md")); err == nil {
					fmt.Printf("Removing skill %s from %s\n", name, targetDir)
					if err := os.RemoveAll(targetDir); err != nil {
						return fmt.Errorf("failed to remove skill: %w", err)
					}
					removed = true
				}
			}
		}
	}

	if !removed {
		return fmt.Errorf("skill %s not found", name)
	}
	fmt.Printf("Successfully removed skill %s\n", name)
	return nil
}

func List(format string) error {
	scopes := []string{"user", "project"}
	agents := []string{"", "cursor", "copilot"}

	var allSkills []*SkillMetadata
	seen := make(map[string]bool)

	for _, scope := range scopes {
		for _, agent := range agents {
			baseDir, err := GetTargetDirectory(scope, agent, "")
			if err != nil {
				continue
			}

			entries, err := os.ReadDir(baseDir)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				if entry.IsDir() {
					dir := filepath.Join(baseDir, entry.Name())
					if meta, err := ReadMetadata(dir); err == nil {
						key := dir
						if !seen[key] {
							allSkills = append(allSkills, meta)
							seen[key] = true
						}
					}
				}
			}
		}
	}

	if format == "json" {
		data, err := json.MarshalIndent(allSkills, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if len(allSkills) == 0 {
		fmt.Println("No skills installed.")
		return nil
	}

	fmt.Printf("%-20s %-15s %-15s %s\n", "NAME", "SCOPE", "AGENT", "SOURCE")
	fmt.Println(strings.Repeat("-", 80))
	for _, s := range allSkills {
		agentStr := s.Agent
		if agentStr == "" {
			agentStr = "default"
		}
		fmt.Printf("%-20s %-15s %-15s %s\n", s.Name, s.Scope, agentStr, s.Source)
	}

	return nil
}

func Inspect(name string) error {
	if name == "" {
		return fmt.Errorf("skill name is required")
	}

	scopes := []string{"user", "project"}
	agents := []string{"", "cursor", "copilot"}

	var targetDir string
	var meta *SkillMetadata

	for _, scope := range scopes {
		for _, agent := range agents {
			dir, err := GetTargetDirectory(scope, agent, name)
			if err != nil {
				continue
			}
			if m, err := ReadMetadata(dir); err == nil {
				targetDir = dir
				meta = m
				break
			}
		}
		if targetDir != "" {
			break
		}
	}

	if targetDir == "" {
		return fmt.Errorf("skill %s not found", name)
	}

	fmt.Printf("Skill: %s\n", meta.Name)
	fmt.Printf("Source: %s\n", meta.Source)
	fmt.Printf("Revision: %s\n", meta.Revision)
	fmt.Printf("Scope: %s\n", meta.Scope)
	fmt.Printf("Agent: %s\n", meta.Agent)
	fmt.Printf("Install Time: %s\n", meta.InstallTime)
	fmt.Printf("Path: %s\n", targetDir)
	fmt.Println("\n--- SKILL.md ---")

	content, err := os.ReadFile(filepath.Join(targetDir, "SKILL.md"))
	if err != nil {
		fmt.Println("(Failed to read SKILL.md)")
	} else {
		// print first 500 chars max
		str := string(content)
		if len(str) > 500 {
			str = str[:500] + "\n... (truncated)"
		}
		fmt.Println(str)
	}

	return nil
}
