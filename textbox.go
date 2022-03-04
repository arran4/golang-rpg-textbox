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
	"image/color"
	"log"
	"time"
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

func BoxTextBox() *boxTextBox {
	return &boxTextBox{}
}

type boxTextBox struct{}

var _ PostDrawer = (*boxTextBox)(nil)

func (btb *boxTextBox) PostDraw(target wordwrap.Image, layout *SimpleLayout, ls []wordwrap.Line, options ...wordwrap.DrawOption) error {
	util.DrawBox(target, layout.textRect)
	return nil
}

func (btb *boxTextBox) apply(box *TextBox) {
	box.postDraw = append(box.postDraw, btb)
}

type PostDrawer interface {
	PostDraw(target wordwrap.Image, layout *SimpleLayout, ls []wordwrap.Line, options ...wordwrap.DrawOption) error
}

type TextBox struct {
	avatarLocation      AvatarLocations
	moreChevronLocation MoreChevronLocations
	theme               theme.Theme
	wordwrapOptions     []wordwrap.WrapperOption
	wrapper             *wordwrap.SimpleWrapper
	name                Name
	nextPage            int
	pages               []*Page
	avatarFit           AvatarFit
	avatar              *avatar
	postDraw            []PostDrawer
	animation           AnimationMode
}

type Option interface {
	apply(*TextBox)
}

type Name string

func (n Name) apply(box *TextBox) {
	box.name = n
}

type avatar struct {
	wordwrap.Image
}

func Avatar(i wordwrap.Image) Option {
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
		tb.wordwrapOptions = append(tb.wordwrapOptions, wordwrap.NewPageBreakBox(wordwrap.NewImageBox(th.Chevron(), wordwrap.ImageBoxMetricCenter(th.FontDrawer()))))
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
		ydiff := util.Max(l.chevronRect.Dy()-(destRect.Max.Y-l.textRect.Max.Y), 0)
		l.textRect.Max.Y -= ydiff
		l.chevronRect = l.chevronRect.Sub(image.Pt(0, ydiff))
	default:
		return nil, fmt.Errorf("unknown more chevron location %v", tb.moreChevronLocation)
	}
	switch tb.moreChevronLocation {
	case NoMoreChevron, TextEndChevron:
	case CenterBottomInsideTextFrame, CenterBottomOnFrameTextFrame:
		l.chevronRect = l.chevronRect.Add(image.Pt(l.textRect.Min.X+(l.textRect.Dx()-l.chevronRect.Dx())/2, l.textRect.Max.Y))
	case CenterBottomInsideFrame, CenterBottomOnFrameFrame:
		l.chevronRect = l.chevronRect.Add(image.Pt(l.centerRect.Min.X+(l.centerRect.Dx()-l.chevronRect.Dx())/2, l.textRect.Max.Y))
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

func drawFrame(t theme.Theme, target wordwrap.Image, options ...wordwrap.DrawOption) error {
	switch t := t.(type) {
	case theme.Frame:
		fc := t.FrameCenter()
		f := t.Frame()
		for _, option := range options {
			switch option := option.(type) {
			case wordwrap.SourceImageMapper:
				f = option(f)
			}
		}
		ttb := target.Bounds()
		fd := frame.NewFrame(ttb, f, fc, frame.Stretched)
		draw.Draw(target, ttb, fd, fd.Bounds().Min, draw.Over)
	default:
		return fmt.Errorf("invalid theme, missing a frame drawer")
	}
	return nil
}

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

type AnimationMode interface {
	DrawOption(target wordwrap.Image) (lastPage bool, userInputAccepted bool, wait time.Duration, err error)
	Done() bool
	Option
}

type FadeState int

const (
	FadeIn FadeState = iota
	FadeOut
)

type FadeAnimation struct {
	tb        *TextBox
	fadeState FadeState
	duration  time.Duration
	steps     int
	step      int
	layout    *SimpleLayout
	page      *Page
}

func (f *FadeAnimation) DrawOption(target wordwrap.Image) (finished bool, userInputAccepted bool, waitTime time.Duration, err error) {
	log.Printf("Step: %d/%d state: %d nextPage: %d", f.step, f.steps, f.fadeState, f.tb.nextPage)
	if f.layout == nil {
		f.layout, f.page, err = f.tb.getNextPage(target.Bounds())
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		if f.layout == nil || f.page == nil {
			log.Printf("Error: Layout of page empty")
			finished = true
			waitTime = -1
			return
		}
	}
	var done bool
	var opts []wordwrap.DrawOption
	doDraw := true
	if (f.step == 0 && f.fadeState == FadeIn) || (f.steps == f.step && f.fadeState == FadeOut) {
		doDraw = false
	} else if (f.step != 0 && f.fadeState == FadeOut) || (f.steps != f.step && f.fadeState == FadeIn) {
		multiplier := float64(f.step) / float64(f.steps)
		if f.fadeState == FadeOut {
			multiplier = float64(f.steps-f.step-1) / float64(f.steps)
		}
		opts = append(opts, wordwrap.SourceImageMapper(func(i image.Image) image.Image {
			return NewAlphaSourceImageMapper(i, multiplier)
		}))
	}

	if doDraw {
		done, err = f.tb.drawFrame(target, f.layout, f.page, opts...)
	}
	waitTime = f.duration / time.Duration(f.steps)
	finished = done && f.step == f.steps && f.fadeState == FadeOut
	userInputAccepted = f.fadeState == FadeIn && f.step == f.steps
	if userInputAccepted {
		f.fadeState = FadeOut
		f.step = 0
	} else if f.step == f.steps && f.fadeState == FadeOut {
		f.fadeState = FadeIn
		f.step = 0
		f.layout = nil
	} else {
		f.step++
	}
	return
}

type AlphaSourceImageMapper struct {
	image.Image
	Multiplier float64
}

func (asim *AlphaSourceImageMapper) At(x, y int) color.Color {
	c := asim.Image.At(x, y)
	r, g, b, a := c.RGBA()
	return color.RGBA64{
		A: uint16(a),
		R: uint16(asim.Multiplier * float64(r)),
		G: uint16(asim.Multiplier * float64(g)),
		B: uint16(asim.Multiplier * float64(b)),
	}
}

func NewAlphaSourceImageMapper(i image.Image, multiplier float64) image.Image {
	return &AlphaSourceImageMapper{
		i,
		multiplier,
	}
}

func (f *FadeAnimation) Done() bool {
	if f.step == f.steps && f.fadeState == FadeOut {
		return !f.tb.HasNext()
	}
	return false
}

func (f *FadeAnimation) apply(box *TextBox) {
	f.tb = box
	box.animation = f
}

var _ AnimationMode = (*FadeAnimation)(nil)

func NewFadeAnimation() *FadeAnimation {
	duration := 2 * time.Second
	return &FadeAnimation{
		duration: duration,
		steps:    int(duration * 10 / time.Second),
	}
}

func (tb *TextBox) DrawNextFrame(target wordwrap.Image) (lastPage bool, userInputAccepted bool, wait time.Duration, err error) {
	if tb.animation == nil {
		next, err := tb.DrawNextPageFrame(target)
		return next, true, 0, err
	}
	return tb.animation.DrawOption(target)
}

func (tb *TextBox) DrawNextPageFrame(target wordwrap.Image, opts ...wordwrap.DrawOption) (bool, error) {
	layout, page, err := tb.getNextPage(target.Bounds())
	if err != nil {
		return false, err
	}
	if layout == nil || page == nil {
		return false, nil
	}
	return tb.drawFrame(target, layout, page, opts...)
}

func (tb *TextBox) drawFrame(target wordwrap.Image, layout *SimpleLayout, page *Page, opts ...wordwrap.DrawOption) (bool, error) {
	if err := drawFrame(tb.theme, target, opts...); err != nil {
		return false, err
	}
	subImage := target.SubImage(layout.TextRect()).(wordwrap.Image)
	tb.drawAvatar(target, layout, opts...)
	if tb.HasNext() {
		tb.drawMoreChevron(target, layout, opts...)
	}
	if err := tb.wrapper.RenderLines(subImage, page.ls, layout.TextRect().Min, opts...); err != nil {
		return false, err
	}
	for _, postDrawer := range tb.postDraw {
		if err := postDrawer.PostDraw(target, layout, page.ls, opts...); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (tb *TextBox) getNextPage(bounds image.Rectangle) (*SimpleLayout, *Page, error) {
	layout, err := NewSimpleLayout(tb, bounds)
	if err != nil {
		return nil, nil, err
	}
	if tb.nextPage == len(tb.pages) {
		n, err := tb.calculateNextFrame(layout)
		if err != nil {
			return nil, nil, err
		}
		if !n {
			return nil, nil, nil
		}
	}
	if tb.nextPage >= len(tb.pages) {
		return nil, nil, nil
	}
	page := tb.pages[tb.nextPage]
	tb.nextPage++
	return layout, page, nil
}

func (tb *TextBox) drawMoreChevron(target wordwrap.Image, layout Layout, options ...wordwrap.DrawOption) {
	cti := tb.theme.Chevron()
	for _, option := range options {
		switch option := option.(type) {
		case wordwrap.SourceImageMapper:
			cti = option(cti)
		}
	}
	ctr := cti.Bounds()
	switch tb.moreChevronLocation {
	case NoMoreChevron, TextEndChevron:
	default:
		draw.Draw(target.SubImage(layout.ChevronRect()).(wordwrap.Image), layout.ChevronRect(), cti, ctr.Min, draw.Over)
	}
}

func (tb *TextBox) drawAvatar(target wordwrap.Image, layout Layout, options ...wordwrap.DrawOption) {
	switch tb.avatarLocation {
	case RightAvatar, LeftAvatar:
		avatarImg := tb.Avatar()
		for _, option := range options {
			switch option := option.(type) {
			case wordwrap.SourceImageMapper:
				avatarImg = option(avatarImg)
			}
		}
		air := avatarImg.Bounds()
		atr := layout.AvatarRect()
		switch tb.avatarFit {
		case NearestNeighbour:
			draw.NearestNeighbor.Scale(target.SubImage(layout.AvatarRect()).(wordwrap.Image), layout.AvatarRect(), avatarImg, air, draw.Over, nil)
		case ApproxBiLinear:
			draw.ApproxBiLinear.Scale(target.SubImage(layout.AvatarRect()).(wordwrap.Image), layout.AvatarRect(), avatarImg, air, draw.Over, nil)
		case NoAvatarFit:
			draw.Draw(target.SubImage(layout.AvatarRect()).(wordwrap.Image), layout.AvatarRect(), avatarImg, air.Min, draw.Over)
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
			draw.Draw(target.SubImage(layout.AvatarRect()).(wordwrap.Image), layout.AvatarRect(), avatarImg, air.Min, draw.Over)
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
	if len(tb.pages) > tb.nextPage {
		return true
	}
	return tb.wrapper.HasNext()
}
