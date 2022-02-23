package rpgtextbox

import (
	"errors"
	"fmt"
	frame "github.com/arran4/golang-frame"
	"github.com/arran4/golang-rpg-textbox/theme"
	"github.com/arran4/golang-wordwrap"
	"image"
	"image/draw"
)

type AvatarLocations int

const (
	NoAvatar AvatarLocations = iota
	CenterLeft
)

func (al AvatarLocations) apply(box *TextBox) {
	box.avatarLocation = al
}

type MoreChevronLocations int

const (
	NoMoreChevron MoreChevronLocations = iota
	CenterBottom
)

func (cl MoreChevronLocations) apply(box *TextBox) {
	box.chevronLocation = cl
}

type TextBox struct {
	avatarLocation  AvatarLocations
	chevronLocation MoreChevronLocations
	theme           theme.Theme
	wordwrapOptions []wordwrap.WrapperOption
	wrapper         *wordwrap.SimpleWrapper
	name            Name
	page            int
	pages           []*Page
}

type Option interface {
	apply(*TextBox)
}

type Name string

func (n Name) apply(box *TextBox) {
	box.name = n
}

// Image because image.Image / draw.Image should really have SubImage as part of it.
type Image interface {
	draw.Image
	SubImage(image.Rectangle) image.Image
}

func NewSimpleTextBox(th theme.Theme, text string, destSize image.Point, options ...Option) (*TextBox, error) {
	tb := &TextBox{
		theme: th,
	}
	for _, option := range options {
		option.apply(tb)
	}
	if tb.wrapper == nil {
		tb.wrapper = wordwrap.NewSimpleWrapper(text, th.FontFace(), tb.wordwrapOptions...)
	}
	destRect := image.Rectangle{
		Min: image.Point{},
		Max: destSize,
	}
	tr, err := tb.calculateTextRect(destRect)
	if err != nil {
		return nil, err
	}
	found, err := tb.calculateNextFrame(tr)
	if err != nil {
		return nil, err
	}
	if !found || len(tb.pages) == 0 {
		return nil, errors.New("no pages drawn")
	}
	return tb, nil
}

func (tb *TextBox) calculateNextFrame(textRect image.Rectangle) (bool, error) {
	ls, _, err := tb.wrapper.TextToRect(textRect)
	if err != nil {
		return false, err
	}
	if len(ls) == 0 {
		return false, nil
	}
	page := &Page{
		ls: ls,
	}
	tb.pages = append(tb.pages, page)
	return true, nil
}

func (tb *TextBox) calculateTextRect(destRect image.Rectangle) (image.Rectangle, error) {
	textRect := destRect
	switch t := tb.theme.(type) {
	case theme.Frame:
		fs := t.Frame().Bounds()
		fc := t.FrameCenter()
		textRect.Min = textRect.Min.Add(fc.Min)
		textRect.Max = textRect.Max.Sub(image.Pt(fs.Max.X-fc.Max.X, fs.Max.Y-fc.Max.Y))
	default:
		return textRect, fmt.Errorf("invalid theme, missing a frame drawer")
	}
	return textRect, nil
}

func drawFrame(t theme.Theme, target Image) error {
	switch t := t.(type) {
	case theme.Frame:
		fc := t.FrameCenter()
		ttb := target.Bounds()
		fd := frame.NewFrame(ttb, t.Frame(), fc, frame.Stretched)
		draw.Draw(target, ttb, fd, ttb.Min, draw.Src)
	default:
		return fmt.Errorf("invalid theme, missing a frame drawer")
	}
	return nil
}

var (
	ErrEOF = errors.New("end of pages")
)

type Page struct {
	ls []wordwrap.Line
}

func (tb *TextBox) CalculateAllPages(destSize image.Point) (int, error) {
	destRect := image.Rectangle{
		Min: image.Point{},
		Max: destSize,
	}
	tr, err := tb.calculateTextRect(destRect)
	if err != nil {
		return len(tb.pages), err
	}
	for {
		found, err := tb.calculateNextFrame(tr)
		if err != nil {
			return len(tb.pages), err
		}
		if !found {
			return len(tb.pages), nil
		}
	}
}

func (tb *TextBox) DrawNextPageFrame(target Image) (bool, error) {
	var page *Page
	if tb.page == len(tb.pages) {
		n, err := tb.calculateNextFrame(target.Bounds())
		if err != nil {
			return false, err
		}
		if !n {
			return false, nil
		}
	}
	if tb.page >= len(tb.pages) {
		return false, nil
	}
	page = tb.pages[tb.page]
	tb.page++
	if err := drawFrame(tb.theme, target); err != nil {
		return false, err
	}
	textRect, err := tb.calculateTextRect(target.Bounds())
	if err != nil {
		return false, err
	}
	subImage := target.SubImage(textRect).(Image)
	if err := tb.wrapper.RenderLines(subImage, page.ls, textRect.Min); err != nil {
		return false, err
	}
	return true, nil
}
