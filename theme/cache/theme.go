package cache

import (
	_ "embed"
	"github.com/arran4/golang-rpg-textbox/theme"
	"image"
)

type Source interface {
	theme.Frame
	theme.Theme
}

type t struct {
	Source
	chevron image.Image
	frame   image.Image
	avatar  image.Image
}

// New creates a caching theme only caches images
func New(source Source, err error) (*t, error) {
	return &t{
		Source: source,
	}, err
}

var _ theme.Theme = (*t)(nil)
var _ theme.Frame = (*t)(nil)

func (t *t) Chevron() image.Image {
	if t.chevron == nil {
		t.chevron = t.Source.Chevron()
	}
	return t.chevron
}

func (t *t) Frame() image.Image {
	if t.frame == nil {
		t.frame = t.Source.Frame()
	}
	return t.frame
}

func (t *t) FrameCenter() image.Rectangle {
	return t.Source.FrameCenter()
}

func (t *t) Avatar() image.Image {
	if t.avatar == nil {
		t.avatar = t.Source.Avatar()
	}
	return t.avatar
}
