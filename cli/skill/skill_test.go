package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTargetDirectory(t *testing.T) {
	// Project scope
	dir, err := GetTargetDirectory("project", "", "myskill")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := filepath.Join(".agents", "skills", "myskill")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}

	// User scope with specific agent
	home, _ := os.UserHomeDir()
	dir, err = GetTargetDirectory("user", "cursor", "myskill")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected = filepath.Join(home, ".cursor", "skills", "myskill")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestLocalInstallAndRemove(t *testing.T) {
	// Setup mock source
	sourceDir := t.TempDir()
	skillDir := filepath.Join(sourceDir, "testskill")
	os.MkdirAll(skillDir, 0755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Test Skill"), 0644)

	// Change working directory to a temp dir so "project" scope installs there
	cwd, _ := os.Getwd()
	workDir := t.TempDir()
	os.Chdir(workDir)
	defer os.Chdir(cwd)

	// Test Install
	err := Install(skillDir, "project", "", []string{})
	if err != nil {
		t.Fatalf("failed to install local skill: %v", err)
	}

	targetDir, _ := GetTargetDirectory("project", "", "testskill")
	if _, err := os.Stat(filepath.Join(targetDir, "SKILL.md")); os.IsNotExist(err) {
		t.Errorf("SKILL.md was not installed to %s", targetDir)
	}

	meta, err := ReadMetadata(targetDir)
	if err != nil {
		t.Fatalf("failed to read metadata: %v", err)
	}
	if meta.Name != "testskill" {
		t.Errorf("expected name testskill, got %s", meta.Name)
	}

	// Test Update
	err = Update("testskill", false, false)
	if err != nil {
		t.Fatalf("failed to update local skill: %v", err)
	}

	// Test List
	err = List("text")
	if err != nil {
		t.Fatalf("failed to list skills: %v", err)
	}

	// Test Inspect
	err = Inspect("testskill")
	if err != nil {
		t.Fatalf("failed to inspect skill: %v", err)
	}

	// Test Remove
	err = Remove("testskill", false)
	if err != nil {
		t.Fatalf("failed to remove local skill: %v", err)
	}

	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		t.Errorf("target dir was not removed: %s", targetDir)
	}
}
