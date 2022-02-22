package theme

import (
	"golang.org/x/image/font"
	"image"
)

type Theme interface {
	Chevron() image.Image
	Frame() image.Image
	FrameCenter() image.Rectangle
	Person() image.Image
	FontFace() font.Face
	FontDrawer() *font.Drawer
}
