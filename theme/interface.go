package theme

import (
	"golang.org/x/image/font"
	"image"
)

// Theme basics for a theme. Selects avatar, more text chevron and font to use needs to be combined with Frame
type Theme interface {
	Chevron() image.Image
	Avatar() image.Image
	FontFace() font.Face
	FontDrawer() *font.Drawer
}

// Frame defined for a theme using github.com/arran4/golang-frame
type Frame interface {
	Frame() image.Image
	FrameCenter() image.Rectangle
}
