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
	textRect        image.Rectangle
	target          Image
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

func NewSimpleTextBox(th theme.Theme, text string, i Image, options ...Option) (*TextBox, error) {
	tb := &TextBox{
		theme:  th,
		target: i,
	}
	for _, option := range options {
		option.apply(tb)
	}
	if tb.wrapper == nil {
		tb.wrapper = wordwrap.NewSimpleWrapper(text, th.FontFace(), tb.wordwrapOptions...)
	}
	if err := tb.calculateTextRect(); err != nil {
		return nil, err
	}
	for {
		ls, _, err := tb.wrapper.TextToRect(tb.textRect)
		if err != nil {
			return nil, err
		}
		if len(ls) == 0 {
			break
		}
		page := &Page{
			ls: ls,
		}
		tb.pages = append(tb.pages, page)
	}
	if len(tb.pages) == 0 {
		return nil, errors.New("no pages drawn")
	}
	return tb, nil
}

func (tb *TextBox) calculateTextRect() error {
	target := tb.target.Bounds()
	tb.textRect = target
	switch t := tb.theme.(type) {
	case theme.Frame:
		fs := t.Frame().Bounds()
		fc := t.FrameCenter()
		tb.textRect.Min = tb.textRect.Min.Add(fc.Min)
		tb.textRect.Max = tb.textRect.Max.Sub(image.Pt(fs.Max.X-fc.Max.X, fs.Max.Y-fc.Max.Y))
	default:
		return fmt.Errorf("invalid theme, missing a frame drawer")
	}
	return nil
}

func (tb *TextBox) drawFrame() error {
	switch t := tb.theme.(type) {
	case theme.Frame:
		fc := t.FrameCenter()
		ttb := tb.target.Bounds()
		fd := frame.NewFrame(ttb, t.Frame(), fc, frame.Stretched)
		draw.Draw(tb.target, ttb, fd, ttb.Min, draw.Src)
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

func (tb *TextBox) DrawNextFrame() (bool, error) {
	var page *Page
	if tb.page >= len(tb.pages) {
		return false, nil
	}
	page = tb.pages[tb.page]
	tb.page++
	if err := tb.drawFrame(); err != nil {
		return false, err
	}
	subImage := tb.target.SubImage(tb.textRect).(Image)
	err := tb.wrapper.RenderLines(subImage, page.ls, tb.textRect.Min)
	return true, err
}
