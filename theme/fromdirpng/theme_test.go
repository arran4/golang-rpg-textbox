package fromdirpng

import (
	"path/filepath"
	"testing"
)

func BenchmarkChevron(b *testing.B) {
	// Use the sibling "simple" directory which contains the required png files
	dir := filepath.Join("..", "simple")

	t, err := New(dir, nil)
	if err != nil {
		b.Fatalf("Failed to create theme: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = t.Chevron()
	}
}

func TestChevron_MissingFiles(t *testing.T) {
	dir := t.TempDir() // Empty directory
	theme, err := New(dir, nil)
	if err != nil {
		t.Fatalf("Expected no error when creating theme from empty directory (lazy loading), got %v", err)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when accessing missing file, got none")
		}
	}()
	_ = theme.Chevron()
}
