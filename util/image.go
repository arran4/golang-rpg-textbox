package util

import (
	"fmt"
	"golang.org/x/image/draw"
	"image"
	"image/png"
	"log"
	"os"
)

//nolint:golint,unused
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
	i, err := png.Decode(fi)
	if err != nil {
		return nil, fmt.Errorf("png encoding: %w", err)
	}
	return i.(Image), nil
}

func SaveFile(i Image, fn string) error {
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

// Image because image.Image / draw.Image should really have SubImage as part of it.
type Image interface {
	draw.Image
	SubImage(image.Rectangle) image.Image
}
