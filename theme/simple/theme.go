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

type t struct {
	chevron image.Image
	frame   image.Image
	person  image.Image
}

func New() (*t, error) {
	return &t{}, nil
}

var _ theme.Theme = (*t)(nil)
var _ theme.Frame = (*t)(nil)

func (t *t) Chevron() image.Image {
	if t.chevron == nil {
		var err error
		t.chevron, err = png.Decode(bytes.NewReader(ChevronBytes))
		if err != nil {
			panic(err)
		}
	}
	return t.chevron
}

func (t *t) Frame() image.Image {
	if t.frame == nil {
		var err error
		t.frame, err = png.Decode(bytes.NewReader(FrameBytes))
		if err != nil {
			panic(err)
		}
	}
	return t.frame
}

func (t *t) FrameCenter() image.Rectangle {
	return image.Rect(35, 34, 63, 58)
}

func (t *t) Avatar() image.Image {
	if t.person == nil {
		var err error
		t.person, err = png.Decode(bytes.NewReader(AvatarBytes))
		if err != nil {
			panic(err)
		}
	}
	return t.person
}

func (t *t) FontFace() font.Face {
	f, err := truetype.Parse(goregular.TTF)
	if err != nil {
		panic(err)
	}
	return truetype.NewFace(f, &truetype.Options{
		Size: 16,
		DPI:  75,
	})
}

func (t *t) FontDrawer() *font.Drawer {
	return &font.Drawer{
		Src:  nil,
		Face: t.FontFace(),
	}
}
