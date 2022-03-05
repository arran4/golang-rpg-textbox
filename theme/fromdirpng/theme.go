package fromdirpng

import (
	_ "embed"
	"github.com/arran4/golang-rpg-textbox/theme"
	"github.com/arran4/golang-rpg-textbox/util"
	"golang.org/x/image/font"
	"image"
	"path/filepath"
)

type t struct {
	dir      string
	fontFace font.Face
}

// New creates a new theme from a directory location, it assumes all files are PNG.
func New(dir string, fontFace font.Face) (*t, error) {
	return &t{
		dir:      dir,
		fontFace: fontFace,
	}, nil
}

var _ theme.Theme = (*t)(nil)
var _ theme.Frame = (*t)(nil)

func (t *t) Chevron() image.Image {
	chevron, err := util.LoadImageFile(filepath.Join(t.dir, "chevron.png"))
	if err != nil {
		panic(err)
	}
	return chevron
}

func (t *t) Frame() image.Image {
	frame, err := util.LoadImageFile(filepath.Join(t.dir, "frame.png"))
	if err != nil {
		panic(err)
	}
	return frame
}

func (t *t) FrameCenter() image.Rectangle {
	return image.Rect(35, 34, 63, 58)
}

func (t *t) Avatar() image.Image {
	person, err := util.LoadImageFile(filepath.Join(t.dir, "avatar.png"))
	if err != nil {
		panic(err)
	}
	return person
}

func (t *t) FontFace() font.Face {
	return t.fontFace
}

func (t *t) FontDrawer() *font.Drawer {
	return &font.Drawer{
		Src:  nil,
		Face: t.FontFace(),
	}
}
