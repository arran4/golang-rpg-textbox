package simple

import (
	"bytes"
	_ "embed"
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
)

type Theme struct {
	chevron image.Image
	frame   image.Image
	person  image.Image
}

var _ theme.Theme = (*Theme)(nil)
var _ theme.Frame = (*Theme)(nil)

func (t *Theme) Chevron() image.Image {
	if t.chevron == nil {
		var err error
		t.chevron, err = png.Decode(bytes.NewReader(ChevronBytes))
		if err != nil {
			panic(err)
		}
	}
	return t.chevron
}

func (t *Theme) Frame() image.Image {
	if t.frame == nil {
		var err error
		t.frame, err = png.Decode(bytes.NewReader(FrameBytes))
		if err != nil {
			panic(err)
		}
	}
	return t.frame
}

func (t *Theme) FrameCenter() image.Rectangle {
	return image.Rect(34, 34, 63, 58)
}

func (t *Theme) Avatar() image.Image {
	if t.person == nil {
		var err error
		t.person, err = png.Decode(bytes.NewReader(AvatarBytes))
		if err != nil {
			panic(err)
		}
	}
	return t.person
}

func (t *Theme) FontFace() font.Face {
	f, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}
	return truetype.NewFace(f, &truetype.Options{
		Size: 16,
		DPI:  75,
	})
}

func (t *Theme) FontDrawer() *font.Drawer {
	return &font.Drawer{
		Src:  nil,
		Face: t.FontFace(),
	}
}
