package rpgtextbox

import (
	"errors"
	"fmt"
	"github.com/arran4/golang-frame"
	"github.com/arran4/golang-rpg-textbox/theme"
	"github.com/arran4/golang-rpg-textbox/util"
	"github.com/arran4/golang-wordwrap"
	"golang.org/x/image/draw"
	"image"
)

type AvatarLocations int

const (
	NoAvatar AvatarLocations = iota
	LeftAvatar
	RightAvatar
)

func (al AvatarLocations) apply(box *TextBox) {
	box.avatarLocation = al
}

type MoreChevronLocations int

const (
	NoMoreChevron MoreChevronLocations = iota
	CenterBottomInsideTextFrame
	CenterBottomInsideFrame
	CenterBottomOnFrameTextFrame
	CenterBottomOnFrameFrame
	RightBottomInsideTextFrame
	RightBottomInsideFrame
	RightBottomOnFrameTextFrame
	RightBottomOnFrameFrame
	TextEndChevron
)

func (cl MoreChevronLocations) apply(box *TextBox) {
	box.moreChevronLocation = cl
}

type AvatarFit int

const (
	NoAvatarFit AvatarFit = iota
	CenterAvatar
	NearestNeighbour
	ApproxBiLinear
)

func (af AvatarFit) apply(box *TextBox) {
	box.avatarFit = af
}

var BoxTextBox = &boxTextBox{}

type boxTextBox struct{}

func (btb *boxTextBox) PostDraw(target util.Image, layout *SimpleLayout, ls []wordwrap.Line) error {
	util.DrawBox(target, layout.textRect)
	return nil
}

func (btb *boxTextBox) apply(box *TextBox) {
	box.postDraw = append(box.postDraw, btb)
}

type PostDrawer interface {
	PostDraw(target util.Image, layout *SimpleLayout, ls []wordwrap.Line) error
}

type TextBox struct {
	avatarLocation      AvatarLocations
	moreChevronLocation MoreChevronLocations
	theme               theme.Theme
	wordwrapOptions     []wordwrap.WrapperOption
	wrapper             *wordwrap.SimpleWrapper
	name                Name
	page                int
	pages               []*Page
	avatarFit           AvatarFit
	avatar              *avatar
	postDraw            []PostDrawer
}

type Option interface {
	apply(*TextBox)
}

type Name string

func (n Name) apply(box *TextBox) {
	box.name = n
}

type avatar struct {
	util.Image
}

func Avatar(i util.Image) Option {
	return &avatar{
		Image: i,
	}
}

func (a *avatar) apply(box *TextBox) {
	box.avatar = a
}

func NewSimpleTextBox(th theme.Theme, text string, destSize image.Point, options ...Option) (*TextBox, error) {
	tb := &TextBox{
		theme: th,
	}
	for _, option := range options {
		option.apply(tb)
	}
	if tb.moreChevronLocation == TextEndChevron {
		tb.wordwrapOptions = append(tb.wordwrapOptions, wordwrap.NewPageBreakBox(wordwrap.NewImageBox(th.Chevron())))
	}
	if tb.wrapper == nil {
		tb.wrapper = wordwrap.NewSimpleWrapper(text, th.FontFace(), tb.wordwrapOptions...)
	}
	destRect := image.Rectangle{
		Min: image.Point{},
		Max: destSize,
	}
	layout, err := NewSimpleLayout(tb, destRect)
	if err != nil {
		return nil, err
	}
	found, err := tb.calculateNextFrame(layout)
	if err != nil {
		return nil, err
	}
	if !found || len(tb.pages) == 0 {
		return nil, errors.New("no pages drawn")
	}
	return tb, nil
}

func (tb *TextBox) calculateNextFrame(layout Layout) (bool, error) {
	ls, _, err := tb.wrapper.TextToRect(layout.TextRect())
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

type Layout interface {
	TextRect() image.Rectangle
	CenterRect() image.Rectangle
	AvatarRect() image.Rectangle
	ChevronRect() image.Rectangle
}

type SimpleLayout struct {
	textRect    image.Rectangle
	centerRect  image.Rectangle
	avatarRect  image.Rectangle
	chevronRect image.Rectangle
}

func (sl *SimpleLayout) TextRect() image.Rectangle {
	return sl.textRect
}

func (sl *SimpleLayout) CenterRect() image.Rectangle {
	return sl.centerRect
}

func (sl *SimpleLayout) AvatarRect() image.Rectangle {
	return sl.avatarRect
}

func (sl *SimpleLayout) ChevronRect() image.Rectangle {
	return sl.chevronRect
}

func NewSimpleLayout(tb *TextBox, destRect image.Rectangle) (*SimpleLayout, error) {
	l := &SimpleLayout{}
	if centerRect, err := tb.calculateCenterRect(destRect); err != nil {
		return nil, err
	} else {
		l.centerRect = centerRect
		l.textRect = centerRect
	}
	l.avatarRect = tb.Avatar().Bounds()
	switch tb.avatarFit {
	case CenterAvatar:
	case NearestNeighbour, ApproxBiLinear:
		dx := float64(l.avatarRect.Dx()) / float64(l.centerRect.Dx())
		dy := float64(l.avatarRect.Dy()) / float64(l.centerRect.Dy())
		if dy > 1 && dy > dx {
			l.avatarRect.Max = image.Point{
				X: l.avatarRect.Min.X + int(float64(l.avatarRect.Dx())/dy),
				Y: l.avatarRect.Min.Y + int(float64(l.avatarRect.Dy())/dy),
			}
		} else if dx > 1 && dx > dy {
			l.avatarRect.Max = image.Point{
				X: l.avatarRect.Min.X + int(float64(l.avatarRect.Dx())/dx),
				Y: l.avatarRect.Min.Y + int(float64(l.avatarRect.Dy())/dx),
			}
		}
	}
	switch tb.avatarLocation {
	case NoAvatar:
	case LeftAvatar:
		l.textRect.Min.X += l.avatarRect.Dx()
		l.avatarRect = image.Rectangle{
			Min: l.centerRect.Min,
			Max: image.Point{
				X: l.textRect.Min.X,
				Y: l.centerRect.Max.Y,
			},
		}
	case RightAvatar:
		l.textRect.Max.X -= l.avatarRect.Dx()
		l.avatarRect = image.Rectangle{
			Min: image.Pt(l.textRect.Max.X, l.textRect.Min.Y),
			Max: l.centerRect.Max,
		}
	default:
		return nil, fmt.Errorf("unknown avatar location %v", tb.avatarLocation)
	}
	l.chevronRect = tb.theme.Chevron().Bounds()
	switch tb.moreChevronLocation {
	case NoMoreChevron, TextEndChevron:
	case CenterBottomInsideTextFrame, CenterBottomInsideFrame, RightBottomInsideTextFrame, RightBottomInsideFrame:
		l.textRect.Max.Y -= l.chevronRect.Dy()
	case CenterBottomOnFrameTextFrame, CenterBottomOnFrameFrame, RightBottomOnFrameTextFrame, RightBottomOnFrameFrame:
		l.textRect.Max.Y -= util.Max(l.chevronRect.Dy()-destRect.Dy(), 0)
	default:
		return nil, fmt.Errorf("unknown more chevron location %v", tb.moreChevronLocation)
	}
	switch tb.moreChevronLocation {
	case NoMoreChevron, TextEndChevron:
	case CenterBottomInsideTextFrame, CenterBottomOnFrameTextFrame:
		l.chevronRect = l.chevronRect.Add(image.Pt(l.textRect.Min.X+(l.textRect.Dx()+l.chevronRect.Dx())/2, l.textRect.Max.Y))
	case CenterBottomInsideFrame, CenterBottomOnFrameFrame:
		l.chevronRect = l.chevronRect.Add(image.Pt(l.centerRect.Min.X+(l.centerRect.Dx()+l.chevronRect.Dx())/2, l.textRect.Max.Y))
	case RightBottomInsideTextFrame, RightBottomInsideFrame:
		l.chevronRect = l.chevronRect.Add(image.Pt(l.textRect.Max.X-l.chevronRect.Dx(), l.textRect.Max.Y))
	case RightBottomOnFrameTextFrame, RightBottomOnFrameFrame:
		l.chevronRect = l.chevronRect.Add(image.Pt(l.centerRect.Max.X-l.chevronRect.Dx(), l.textRect.Max.Y))
	default:
		return nil, fmt.Errorf("unknown more chevron location %v", tb.moreChevronLocation)
	}
	return l, nil
}

func (tb *TextBox) calculateCenterRect(destRect image.Rectangle) (image.Rectangle, error) {
	textRect := destRect
	switch t := tb.theme.(type) {
	case theme.Frame:
		fc := t.FrameCenter()
		fd := frame.NewFrame(destRect, t.Frame(), fc)
		textRect = fd.MiddleRect()
	default:
		return textRect, fmt.Errorf("invalid theme, missing a frame drawer")
	}
	return textRect, nil
}

func drawFrame(t theme.Theme, target util.Image) error {
	switch t := t.(type) {
	case theme.Frame:
		fc := t.FrameCenter()
		ttb := target.Bounds()
		fd := frame.NewFrame(ttb, t.Frame(), fc, frame.Stretched)
		draw.Draw(target, ttb, fd, fd.Bounds().Min, draw.Src)
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
	l, err := NewSimpleLayout(tb, destRect)
	if err != nil {
		return len(tb.pages), err
	}
	for {
		found, err := tb.calculateNextFrame(l)
		if err != nil {
			return len(tb.pages), err
		}
		if !found {
			return len(tb.pages), nil
		}
	}
}

func (tb *TextBox) DrawNextPageFrame(target util.Image) (bool, error) {
	var page *Page
	layout, err := NewSimpleLayout(tb, target.Bounds())
	if err != nil {
		return false, err
	}
	if tb.page == len(tb.pages) {
		n, err := tb.calculateNextFrame(layout)
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
	defer func() { tb.page++ }()
	if err := drawFrame(tb.theme, target); err != nil {
		return false, err
	}
	subImage := target.SubImage(layout.TextRect()).(util.Image)
	tb.drawAvatar(target, layout)
	if tb.HasNext() {
		tb.drawMoreChevron(target, layout)
	}
	if err := tb.wrapper.RenderLines(subImage, page.ls, layout.TextRect().Min); err != nil {
		return false, err
	}
	for _, postDrawer := range tb.postDraw {
		if err := postDrawer.PostDraw(target, layout, page.ls); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (tb *TextBox) drawMoreChevron(target util.Image, layout Layout) {
	cti := tb.theme.Chevron()
	ctr := cti.Bounds()
	switch tb.moreChevronLocation {
	case NoMoreChevron, TextEndChevron:
	default:
		draw.Draw(target.SubImage(layout.ChevronRect()).(util.Image), layout.ChevronRect(), cti, ctr.Min, draw.Over)
	}
}

func (tb *TextBox) drawAvatar(target util.Image, layout Layout) {
	switch tb.avatarLocation {
	case RightAvatar, LeftAvatar:
		avatarImg := tb.Avatar()
		air := avatarImg.Bounds()
		atr := layout.AvatarRect()
		switch tb.avatarFit {
		case NearestNeighbour:
			draw.NearestNeighbor.Scale(target.SubImage(layout.AvatarRect()).(util.Image), layout.AvatarRect(), avatarImg, air, draw.Over, nil)
		case ApproxBiLinear:
			draw.ApproxBiLinear.Scale(target.SubImage(layout.AvatarRect()).(util.Image), layout.AvatarRect(), avatarImg, air, draw.Over, nil)
		case NoAvatarFit:
			draw.Draw(target.SubImage(layout.AvatarRect()).(util.Image), layout.AvatarRect(), avatarImg, air.Min, draw.Over)
		case CenterAvatar:
			dx := air.Dx() - atr.Dx()
			dy := air.Dy() - atr.Dy()
			if dy < 0 {
				dy = 0
			}
			if dx < 0 {
				dx = 0
			}
			air = air.Add(image.Pt(dx/2, dy/2))
			draw.Draw(target.SubImage(layout.AvatarRect()).(util.Image), layout.AvatarRect(), avatarImg, air.Min, draw.Over)
		}
	}
}

func (tb *TextBox) Avatar() image.Image {
	if tb.avatar != nil {
		return tb.avatar
	}
	return tb.theme.Avatar()
}

func (tb *TextBox) HasNext() bool {
	if len(tb.pages) > tb.page+1 {
		return true
	}
	return tb.wrapper.HasNext()
}
