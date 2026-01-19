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
	chevron  image.Image
	frame    image.Image
	avatar   image.Image
}

// New creates a new theme from a directory location, it assumes all files are PNG.
func New(dir string, fontFace font.Face) (*t, error) {
	chevron, err := util.LoadImageFile(filepath.Join(dir, "chevron.png"))
	if err != nil {
		return nil, err
	}
	frame, err := util.LoadImageFile(filepath.Join(dir, "frame.png"))
	if err != nil {
		return nil, err
	}
	avatar, err := util.LoadImageFile(filepath.Join(dir, "avatar.png"))
	if err != nil {
		return nil, err
	}
	return &t{
		dir:      dir,
		fontFace: fontFace,
		chevron:  chevron,
		frame:    frame,
		avatar:   avatar,
	}, nil
}

var _ theme.Theme = (*t)(nil)
var _ theme.Frame = (*t)(nil)

func (t *t) Chevron() image.Image {
	return t.chevron
}

func (t *t) Frame() image.Image {
	return t.frame
}

func (t *t) FrameCenter() image.Rectangle {
	return image.Rect(35, 34, 63, 58)
}

func (t *t) Avatar() image.Image {
	return t.avatar
}

func (t *t) FontFace() font.Face {
	return t.fontFace
}

func (t *t) FontDrawer() *font.Drawer {
	return &font.Drawer{
		Src:  image.NewUniform(image.Black),
		Face: t.FontFace(),
	}
}
