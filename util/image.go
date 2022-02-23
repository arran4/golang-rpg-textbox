package util

import (
	"fmt"
	"github.com/arran4/golang-rpg-textbox"
	"image/png"
	"log"
	"os"
)

//nolint:golint,unused
func LoadImageFile(fn string) (rpgtextbox.Image, error) {
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
	return i.(rpgtextbox.Image), nil
}
