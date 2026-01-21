package rpgtextbox

import (
	"fmt"
	"image"
	"testing"

	"github.com/arran4/golang-rpg-textbox/theme/simple"
	"github.com/arran4/golang-rpg-textbox/util"
)

func TestNamePositioning(t *testing.T) {
	theme, err := simple.New()
	if err != nil {
		t.Fatalf("Failed to create simple theme: %v", err)
	}

	tests := []struct {
		name     string
		position NamePositions
	}{
		{"NameTopLeftAboveTextInFrame", NameTopLeftAboveTextInFrame},
		{"NameTopCenterInFrame", NameTopCenterInFrame},
		{"NameTopLeftAboveFrame", NameTopLeftAboveFrame},
		{"NameTopCenterAboveFrame", NameTopCenterAboveFrame},
	}

	width, height := 600, 150
	textBoxSize := image.Pt(width, height)
	text := "Hello, this is a test text to verify name tag positioning."
	name := Name("Test Name")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb, err := NewSimpleTextBox(theme, text, textBoxSize, name, tt.position)
			if err != nil {
				t.Fatalf("Error creating text box: %v", err)
			}

			pages, err := tb.CalculateAllPages(textBoxSize)
			if err != nil {
				t.Fatalf("Calculate pages error: %v", err)
			}
			if pages == 0 {
				t.Fatal("Expected at least one page")
			}

			i := image.NewRGBA(image.Rect(0, 0, width, height))
			if _, err := tb.DrawNextPageFrame(i); err != nil {
				t.Fatalf("Draw next frame error: %v", err)
			}

			filename := fmt.Sprintf("test_output/%s.png", tt.name)
			if err := util.SavePngFile(i, filename); err != nil {
				t.Fatalf("Error saving file %s: %v", filename, err)
			}
			t.Logf("Saved output to %s", filename)
		})
	}
}
