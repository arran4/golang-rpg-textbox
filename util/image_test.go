package util

import (
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestDrawBox(t *testing.T) {
	w, h := 10, 10
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Fill with white
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	rect := image.Rect(2, 2, 8, 8)
	DrawBox(img, rect)

	// Verify corners and edges are black
	// The box is drawn at x in [s.Min.X, s.Max.X) and y in [s.Min.Y, s.Max.Y)
    // Top edge: y = s.Min.Y, x from s.Min.X to s.Max.X-1
    // Bottom edge: y = s.Max.Y-1, x from s.Min.X to s.Max.X-1
    // Left edge: x = s.Min.X, y from s.Min.Y to s.Max.Y-1
    // Right edge: x = s.Max.X-1, y from s.Min.Y to s.Max.Y-1

    // So for Rect(2, 2, 8, 8):
    // X range: 2 to 7 (inclusive)
    // Y range: 2 to 7 (inclusive)
    // Edges should be black.

	expectedBlack := []image.Point{
		{2, 2}, {7, 2}, // Top corners
		{2, 7}, {7, 7}, // Bottom corners
		{3, 2}, {2, 3}, // Mid points on edges
	}

	for _, p := range expectedBlack {
		c := img.RGBAAt(p.X, p.Y)
		if c.R != 0 || c.G != 0 || c.B != 0 || c.A != 255 {
			t.Errorf("Expected black at %v, got %v", p, c)
		}
	}

    // Verify center is white (not drawn)
    center := image.Point{4, 4}
    c := img.RGBAAt(center.X, center.Y)
    if c.R != 255 || c.G != 255 || c.B != 255 || c.A != 255 {
        t.Errorf("Expected white at %v, got %v", center, c)
    }

    // Verify outside is white
    outside := image.Point{1, 1}
    c = img.RGBAAt(outside.X, outside.Y)
    if c.R != 255 || c.G != 255 || c.B != 255 || c.A != 255 {
        t.Errorf("Expected white at %v, got %v", outside, c)
    }
}

func BenchmarkDrawBox(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	rect := image.Rect(100, 100, 900, 900)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DrawBox(img, rect)
	}
}
