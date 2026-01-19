package util

import (
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/color"
	"image/gif"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
)

func LoadImageFile(fn string) (Image, error) {
	fi, err := os.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("file create: %w", err)
	}
	defer func() {
		if err := fi.Close(); err != nil {
			log.Printf("File close error: %s", err)
		}
	}()
	i, _, err := image.Decode(fi)
	if err != nil {
		return nil, fmt.Errorf("image encoding: %w", err)
	}
	return i.(Image), nil
}

func SavePngFile(i Image, fn string) error {
	_ = os.MkdirAll("images", 0755)
	fi, err := os.Create(fn)
	if err != nil {
		return fmt.Errorf("file create: %w", err)
	}
	defer func() {
		if err := fi.Close(); err != nil {
			log.Printf("File close error: %s", err)
		}
	}()
	if err := png.Encode(fi, i); err != nil {
		return fmt.Errorf("png encoding: %w", err)
	}
	return nil
}

func SaveGifFile(fn string, options *gif.GIF) error {
	_ = os.MkdirAll("images", 0755)
	fi, err := os.Create(fn)
	if err != nil {
		return fmt.Errorf("file create: %w", err)
	}
	defer func() {
		if err := fi.Close(); err != nil {
			log.Printf("File close error: %s", err)
		}
	}()
	if err := gif.EncodeAll(fi, options); err != nil {
		return fmt.Errorf("png encoding: %w", err)
	}
	return nil
}

// Image because image.Image / draw.Image should really have SubImage as part of it.
type Image interface {
	draw.Image
	SubImage(image.Rectangle) image.Image
}

func DrawBox(i draw.Image, s image.Rectangle) {
	// Top
	draw.Draw(i, image.Rect(s.Min.X, s.Min.Y, s.Max.X, s.Min.Y+1), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	// Bottom
	draw.Draw(i, image.Rect(s.Min.X, s.Max.Y-1, s.Max.X, s.Max.Y), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	// Left
	draw.Draw(i, image.Rect(s.Min.X, s.Min.Y, s.Min.X+1, s.Max.Y), &image.Uniform{color.Black}, image.Point{}, draw.Src)
	// Right
	draw.Draw(i, image.Rect(s.Max.X-1, s.Min.Y, s.Max.X, s.Max.Y), &image.Uniform{color.Black}, image.Point{}, draw.Src)
}
