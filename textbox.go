package rpgtextbox

import (
	"errors"
	"fmt"
	"image"
	"time"

	frame "github.com/arran4/golang-frame"
	"github.com/arran4/golang-rpg-textbox/theme"
	"github.com/arran4/golang-rpg-textbox/util"
	wordwrap "github.com/arran4/golang-wordwrap"
	"github.com/arran4/spacemap/shared"
	"golang.org/x/image/draw"
)

// AvatarLocations Positioning locations for avatars
type AvatarLocations int

const (
	// NoAvatar default show no avatar
	NoAvatar AvatarLocations = iota
	// LeftAvatar Avatar on the left
	LeftAvatar
	// RightAvatar on the right
	RightAvatar
)

// apply Set the location when used as an Option
func (al AvatarLocations) apply(box *TextBox) {
	box.avatarLocation = al
}

// NamePositions Positioning locations for the name tag
type NamePositions int

const (
	// NoName default show no name tag
	NoName NamePositions = iota
	// NameTopLeftAboveTextInFrame on the left
	NameTopLeftAboveTextInFrame
	// NameTopCenterInFrame on the left
	NameTopCenterInFrame
	// NameLeftAboveAvatarInFrame on the right
	NameLeftAboveAvatarInFrame
)

// apply Set the location when used as an Option
func (al NamePositions) apply(box *TextBox) {
	box.namePosition = al
}

// MoreChevronLocations Position to put the "more text" marker
type MoreChevronLocations int

const (
	// NoMoreChevron default. Don't show an avatar
	NoMoreChevron MoreChevronLocations = iota
	// CenterBottomInsideTextFrame Name says it all. See readme or samples for more details.
	CenterBottomInsideTextFrame
	// CenterBottomInsideFrame Name says it all. See readme or samples for more details.
	CenterBottomInsideFrame
	// CenterBottomOnFrameTextFrame Name says it all. See readme or samples for more details.
	CenterBottomOnFrameTextFrame
	// CenterBottomOnFrameFrame Name says it all. See readme or samples for more details.
	CenterBottomOnFrameFrame
	// RightBottomInsideTextFrame Name says it all. See readme or samples for more details.
	RightBottomInsideTextFrame
	// RightBottomInsideFrame Name says it all. See readme or samples for more details.
	RightBottomInsideFrame
	// RightBottomOnFrameTextFrame Name says it all. See readme or samples for more details.
	RightBottomOnFrameTextFrame
	// RightBottomOnFrameFrame Name says it all. See readme or samples for more details.
	RightBottomOnFrameFrame
	// TextEndChevron Puts the reader marker at the end of text inline as though it were a character
	TextEndChevron
)

// apply Set the location when used as an Option
func (cl MoreChevronLocations) apply(box *TextBox) {
	box.moreChevronLocation = cl
}

// AvatarFit is an enum of how to handle the avatar not fitting
type AvatarFit int

const (
	// NoAvatarFit Don't attempt to do anything, will have undefined behavior
	NoAvatarFit AvatarFit = iota
	// CenterAvatar as the name suggests
	CenterAvatar
	// NearestNeighbour Resizes the avatar using the NearestNeighbour algorithm from experimental go draw package
	NearestNeighbour
	// ApproxBiLinear Resizes the avatar using the ApproxBiLinear algorithm from experimental go draw package
	ApproxBiLinear
)

// apply Set the location when used as an Option
func (af AvatarFit) apply(box *TextBox) {
	box.avatarFit = af
}

// BoxTextBox Creates a function to draw a box around the area where the text box will be
func BoxTextBox() *boxTextBox {
	return &boxTextBox{}
}

// BoxTextBox Creates a function to draw a box around the area where the text box will be
type boxTextBox struct{}

// Interface enforcement
var _ PostDrawer = (*boxTextBox)(nil)

// PostDraw applies the PostDrawer interface and will execute after all the other elements have been drawn
func (btb *boxTextBox) PostDraw(target wordwrap.Image, layout *SimpleLayout, ls []wordwrap.Line, options ...wordwrap.DrawOption) error {
	util.DrawBox(target, layout.textRect)
	return nil
}

// apply Set the location when used as an Option
func (btb *boxTextBox) apply(box *TextBox) {
	box.postDraw = append(box.postDraw, btb)
}

// PostDrawer allows custom components to be drawn after all other elements have
type PostDrawer interface {
	PostDraw(target wordwrap.Image, layout *SimpleLayout, ls []wordwrap.Line, options ...wordwrap.DrawOption) error
}

// TextBox is the core class
type TextBox struct {
	// avatarLocation the location of the avatars
	avatarLocation AvatarLocations
	// moreChevronLocation More text location
	moreChevronLocation MoreChevronLocations
	// theme the theme to use
	theme theme.Theme
	// wordwrapOptions The options to pass to the wordwrapper
	wordwrapOptions []wordwrap.WrapperOption
	// wrapper The word wrapper itself
	wrapper *wordwrap.SimpleWrapper
	// name the name to show above the text box (TODO)
	name Name
	// nextPage the next page after the one we are on
	nextPage int
	// pages a cache of the pages
	pages []*Page
	// avatarFit Avatar scaling algorithm
	avatarFit AvatarFit
	// avatar Avatar image override
	avatar *avatar
	// postDraw Things to draw after all the other elements have been drawn
	postDraw []PostDrawer
	// animation the selected animation style
	animation AnimationMode
	// namePosition the location of the name tag
	namePosition NamePositions
	// nameBox is the image of the name tag
	// nameBox is the image of the name tag
	nameBox wordwrap.Box
	// spaceMap is the space map to populate
	spaceMap SpaceMap
}

// SpaceMap interface for space mapping
type SpaceMap interface {
	Add(shape shared.Shape, zIndex int)
}

// SetSpaceMap sets the space map to populate
func (tb *TextBox) SetSpaceMap(m SpaceMap) {
	tb.spaceMap = m
}

// Option are the configuration arguments
type Option interface {
	apply(*TextBox)
}

// Name the name of the character to show
type Name string

// apply Set the location when used as an Option
func (n Name) apply(box *TextBox) {
	box.name = n
	box.nameBox, _ = wordwrap.NewSimpleTextBox(box.theme.FontDrawer(), string(n))
}

// avatar An avatar image overwrite of the theme default
type avatar struct {
	wordwrap.Image
}

// Avatar override the default theme avatar
func Avatar(i wordwrap.Image) Option {
	return &avatar{
		Image: i,
	}
}

// apply Set the location when used as an Option
func (a *avatar) apply(box *TextBox) {
	box.avatar = a
}

// NewSimpleTextBox as simple as possible constructor for the RPG TextBox,
// Theme is required
// destSize can be modified on a per frame basis but is the intended size of hte image
// options see readme
func NewSimpleTextBox(th theme.Theme, text string, destSize image.Point, options ...Option) (*TextBox, error) {
	args := []interface{}{text, destSize}
	for _, o := range options {
		args = append(args, o)
	}
	return NewRichTextBox(th, args...)
}

// NewRichTextBox constructor for rich text arguments
// Theme is required
// destSize can be modified on a per frame basis but is the intended size of hte image
// options see readme
func NewRichTextBox(th theme.Theme, args ...interface{}) (*TextBox, error) {
	tb := &TextBox{
		theme: th,
	}
	var wordwrapArgs []interface{}
	// Prepend font face from theme if not provided?
	// NewRichWrapper handles arguments.
	// We need to extract TextBox Options and DestSize.
	destSize := image.Point{}
	foundDestSize := false

	// Iterate args to find TextBox options and DestSize
	for _, arg := range args {
		switch a := arg.(type) {
		case Option:
			a.apply(tb)
		case image.Point:
			if !foundDestSize {
				destSize = a
				foundDestSize = true
			} else {
				// Treat second point as...? Or just error?
				// For now assume only one point is passed which is destSize
				// But wordwrap might accept points? No.
			}
		default:
			wordwrapArgs = append(wordwrapArgs, arg)
		}
	}

	if !foundDestSize {
		// Default or error?
		// Existing code assumed destSize was passed.
		// Let's assume 0,0 is valid if not passed, or caller checks.
	}

	// Add Theme Font Face to wordwrap args if not present?
	// NewRichWrapper takes variadic args.
	// It expects: contents, fontDrawer (or Face), wrapperOptions, boxerOptions...
	// We should pass th.FontFace() as default font if user didn't supply one in ops?
	// wordwrap.ProcessRichArgs parses them.
	// Let's prepend th.FontFace() to wordwrapArgs to ensure a default font is available.
	// But ProcessRichArgs might take the *last* font?
	// Let's append it? No, if we append it, it might override user provided?
	// ProcessRichArgs: if arg matches type, it sets it.
	// If multiple fonts passed, last one wins?
	// Let's check ProcessRichArgs (impl detail of wordwrap).
	// Assuming providing it as one of the args is good.
	wordwrapArgs = append([]interface{}{th.FontFace()}, wordwrapArgs...)

	// We also need to pass tb.wordwrapOptions which might have been populated by TextBox options (like Chevron)
	if tb.moreChevronLocation == TextEndChevron {
		// This uses wordwrap types which are internal to wordwrap package unless exported.
		// NewPageBreakBox is exported.
		pb := wordwrap.NewPageBreakBox(wordwrap.NewImageBox(th.Chevron(), wordwrap.ImageBoxMetricCenter(th.FontDrawer())))
		wordwrapArgs = append(wordwrapArgs, pb)
	}

	// Add any wordwrap options stored in tb (e.g. from Option that modified tb.wordwrapOptions)
	for _, opt := range tb.wordwrapOptions {
		wordwrapArgs = append(wordwrapArgs, opt)
	}

	if tb.wrapper == nil {
		tb.wrapper = wordwrap.NewRichWrapper(wordwrapArgs...)
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

// calculateNextFrame calculates the next frame box layout (positions everything, and reads from the word wrapper)
func (tb *TextBox) calculateNextFrame(layout Layout) (bool, error) {
	ls, _, err := tb.wrapper.TextToRect(layout.TextRect())
	if err != nil {
		return false, err
	}
	if len(ls) == 0 {
		return false, nil
	}
	boxCount := 0
	for _, l := range ls {
		boxCount += len(l.Boxes())
	}
	page := &Page{
		ls:       ls,
		boxCount: boxCount,
	}
	tb.pages = append(tb.pages, page)
	return true, nil
}

// Layout the positioning algorithm output for the layout of the elements on the page
type Layout interface {
	// TextRect the area the text will be boxed into
	TextRect() image.Rectangle
	// CenterRect The midsection basically everything inside the frame
	CenterRect() image.Rectangle
	// AvatarRect the location the avatar is boxed into
	AvatarRect() image.Rectangle
	// ChevronRect the position of the chevron if given the correct position options
	ChevronRect() image.Rectangle
	// NameRect the position of the name if given the correct position options
	NameRect() image.Rectangle
}

// SimpleLayout simple as possible layout. Keeps it mostly what most use-cases would suggest
type SimpleLayout struct {
	// TextRect the area the text will be boxed into
	textRect image.Rectangle
	// CenterRect The midsection basically everything inside the frame
	centerRect image.Rectangle
	// AvatarRect the location the avatar is boxed into
	avatarRect image.Rectangle
	// ChevronRect the position of the chevron if given the correct position options
	chevronRect image.Rectangle
	// nameRect optional position of the name rect
	nameRect image.Rectangle
}

// Interface enforcement
var _ Layout = (*SimpleLayout)(nil)

// NameRect where the name rect belongs
func (sl *SimpleLayout) NameRect() image.Rectangle {
	return sl.nameRect
}

// TextRect the area the text will be boxed into
func (sl *SimpleLayout) TextRect() image.Rectangle {
	return sl.textRect
}

// CenterRect The midsection basically everythign inside the frame
func (sl *SimpleLayout) CenterRect() image.Rectangle {
	return sl.centerRect
}

// AvatarRect the location the avatar is boxed into
func (sl *SimpleLayout) AvatarRect() image.Rectangle {
	return sl.avatarRect
}

// ChevronRect the position of the chevron if given the correct position options
func (sl *SimpleLayout) ChevronRect() image.Rectangle {
	return sl.chevronRect
}

// NewSimpleLayout constructs SimpleLayout simply as possible (for the user.)
func NewSimpleLayout(tb *TextBox, destRect image.Rectangle) (*SimpleLayout, error) {
	l := &SimpleLayout{}
	if centerRect, err := tb.calculateCenterRect(destRect); err != nil {
		return nil, err
	} else {
		l.centerRect = centerRect
		l.textRect = centerRect
	}
	if tb.nameBox != nil {
		m := tb.nameBox.MetricsRect()
		a := tb.nameBox.AdvanceRect()
		height := (m.Ascent + m.Descent).Ceil()
		width := a.Ceil()
		l.nameRect = image.Rect(0, 0, width, height)
		switch tb.namePosition {
		case NoName:
		case NameTopCenterInFrame:
			l.nameRect.Min.X = l.centerRect.Min.X + (l.centerRect.Dx()-l.nameRect.Dx())/2
			l.nameRect.Max.X = l.nameRect.Min.X + width
			fallthrough
		default:
			l.nameRect.Min.Y = l.centerRect.Min.Y
			l.nameRect.Max.Y = l.centerRect.Min.Y + height
			l.centerRect.Min.Y += height
			l.textRect.Min = l.textRect.Min.Add(image.Pt(0, height))
		}
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
	if tb.nameBox != nil {
		switch tb.namePosition {
		case NameLeftAboveAvatarInFrame:
			if tb.avatarLocation != NoAvatar {
				l.nameRect = l.nameRect.Add(image.Pt(l.avatarRect.Min.X, 0))
				break
			}
			fallthrough
		case NameTopLeftAboveTextInFrame:
			l.nameRect = l.nameRect.Add(image.Pt(l.textRect.Min.X, 0))
		}
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

// calculateCenterRect calculates the size of the center rectangle based on the provided theme
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

// drawFrame Draws the theme's frame
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

// Page where a page of lines is stored with some statistical information. A page being what is visible in the provided
// rectangle
type Page struct {
	ls       []wordwrap.Line
	boxCount int
}

// CalculateAllPages calculates the box positioning of all remain pages in advance
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

// DrawNextFrame draws the next frame, if there is an animation it will draw the animation frame.
// lastPage is true if you're on the last page
// userInputAccepted is if it's at the stage where you would typically accept user input (ie the animation is waiting
// user input, doesn't imply anything to do with the animation
// wait is either 0 or less, or the amount of time before the next animation phase
// err is err
// To determine if you're at the end the only way of doing it as of writing is to wait for; lastPage = true,
// userInputAccepted = false, wait = -1
func (tb *TextBox) DrawNextFrame(target wordwrap.Image) (lastPage bool, userInputAccepted bool, wait time.Duration, err error) {
	if tb.animation == nil {
		next, err := tb.DrawNextPageFrame(target)
		return next, true, 0, err
	}
	return tb.animation.DrawOption(target)
}

// DrawNextPageFrame Draws the next frame ignores animation. Please use either function but be very careful if you use
// both, if it's supported is at an animation level
func (tb *TextBox) DrawNextPageFrame(target wordwrap.Image, opts ...wordwrap.DrawOption) (bool, error) {
	layout, page, err := tb.getNextPage(target.Bounds())
	if err != nil {
		return false, err
	}
	if layout == nil || page == nil {
		return false, nil
	}
	return tb.drawPage(target, layout, page, opts...)
}

// drawPage draws the entire page.
func (tb *TextBox) drawPage(target wordwrap.Image, layout *SimpleLayout, page *Page, opts ...wordwrap.DrawOption) (bool, error) {
	if err := drawFrame(tb.theme, target, opts...); err != nil {
		return false, err
	}
	subImage := target.SubImage(layout.TextRect()).(wordwrap.Image)
	tb.drawAvatar(target, layout, opts...)
	if tb.HasNext() {
		tb.drawMoreChevron(target, layout, opts...)
	}
	if tb.name != "" {
		tb.drawNameTag(target, layout, opts...)
	}
	if tb.name != "" {
		tb.drawNameTag(target, layout, opts...)
	}
	if tb.spaceMap != nil {
		opts = append(opts, wordwrap.BoxRecorder(func(box wordwrap.Box, min, max image.Point, bps *wordwrap.BoxPositionStats) {
			tb.spaceMap.Add(&BoxShape{
				Box:  box,
				Rect: image.Rectangle{Min: min, Max: max},
			}, 0)
		}))
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

type BoxShape struct {
	wordwrap.Box
	Rect image.Rectangle
}

func (b *BoxShape) Bounds() image.Rectangle {
	return b.Rect
}

func (b *BoxShape) PointIn(x, y int) bool {
	return b.Rect.Min.X <= x && x < b.Rect.Max.X && b.Rect.Min.Y <= y && y < b.Rect.Max.Y
}

func (b *BoxShape) String() string {
	return fmt.Sprintf("Box(%s)", b.Box.TextValue())
}

func (b *BoxShape) ID() interface{} {
	if i, ok := b.Box.(wordwrap.Identifier); ok {
		return i.ID()
	}
	return nil
}

// getNextPage calculates the next page and updates various info. If all results are nil then it means there is nothing
// left
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

// drawMoreChevron as expected
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

// drawNameTag as expected
func (tb *TextBox) drawNameTag(target wordwrap.Image, layout Layout, options ...wordwrap.DrawOption) {
	if tb.nameBox != nil {
		switch tb.namePosition {
		case NoName:
		default:
			m := tb.nameBox.MetricsRect()
			config := wordwrap.NewDrawConfig(options...)
			var bb = tb.nameBox
			if config.BoxDrawMap != nil {
				bb = config.ApplyMap(bb, &wordwrap.BoxPositionStats{
					LinePositionStats: &wordwrap.LinePositionStats{},
				})
			}
			if bb != nil {
				bb.DrawBox(target.SubImage(layout.NameRect()).(wordwrap.Image), m.Ascent, config)
			}
		}
	}
}

// drawAvatar as expected
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

// Avatar returns the correct avatar (if you have overwritten the theme etc.)
func (tb *TextBox) Avatar() image.Image {
	if tb.avatar != nil {
		return tb.avatar
	}
	return tb.theme.Avatar()
}

// HasNext returns if there is a next page, this doesn't take into consideration if there is an animation
func (tb *TextBox) HasNext() bool {
	if len(tb.pages) > tb.nextPage {
		return true
	}
	return tb.wrapper.HasNext()
}
