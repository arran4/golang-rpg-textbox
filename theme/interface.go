package theme

import (
	"golang.org/x/image/font"
	"image"
)

type Theme interface {
	Chevron() image.Image
	Avatar() image.Image
	FontFace() font.Face
	FontDrawer() *font.Drawer
}

type Frame interface {
	Frame() image.Image
	FrameCenter() image.Rectangle
}
