package simple

import (
	"bytes"
	_ "embed"
	"sync"

	"github.com/arran4/golang-rpg-textbox/theme"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"image"
	"image/png"
)

var (
	//go:embed "chevron.png"
	ChevronBytes []byte
	//go:embed "frame.png"
	FrameBytes []byte
	//go:embed "avatar.png"
	AvatarBytes []byte

	fontFaceOnce sync.Once
	fontFace     font.Face

	chevronOnce sync.Once
	chevronImg  image.Image

	frameOnce sync.Once
	frameImg  image.Image

	avatarOnce sync.Once
	avatarImg  image.Image
)

type t struct{}

// New of my simple drawn theme. Provided for development purposes please substitute with your own
func New() (*t, error) {
	return &t{}, nil
}

var _ theme.Theme = (*t)(nil)
var _ theme.Frame = (*t)(nil)

func (t *t) Chevron() image.Image {
	chevronOnce.Do(func() {
		var err error
		chevronImg, err = png.Decode(bytes.NewReader(ChevronBytes))
		if err != nil {
			panic(err)
		}
	})
	return chevronImg
}

func (t *t) Frame() image.Image {
	frameOnce.Do(func() {
		var err error
		frameImg, err = png.Decode(bytes.NewReader(FrameBytes))
		if err != nil {
			panic(err)
		}
	})
	return frameImg
}

func (t *t) FrameCenter() image.Rectangle {
	return image.Rect(35, 34, 63, 58)
}

func (t *t) Avatar() image.Image {
	avatarOnce.Do(func() {
		var err error
		avatarImg, err = png.Decode(bytes.NewReader(AvatarBytes))
		if err != nil {
			panic(err)
		}
	})
	return avatarImg
}

func (t *t) FontFace() font.Face {
	fontFaceOnce.Do(func() {
		f, err := truetype.Parse(goregular.TTF)
		if err != nil {
			panic(err)
		}
		fontFace = truetype.NewFace(f, &truetype.Options{
			Size: 16,
			DPI:  75,
		})
	})
	return fontFace
}

func (t *t) FontDrawer() *font.Drawer {
	return &font.Drawer{
		Src:  image.NewUniform(image.Black),
		Face: t.FontFace(),
	}
}
