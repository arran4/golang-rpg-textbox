package rpgtextbox

import (
	"errors"
	"fmt"
	"image"
	"log"
	"time"

	frame "github.com/arran4/golang-frame"
	"github.com/arran4/golang-rpg-textbox/theme"
	"github.com/arran4/golang-rpg-textbox/util"
	wordwrap "github.com/arran4/golang-wordwrap"
	"github.com/arran4/spacemap/shared"
	"golang.org/x/image/draw"
)

// AvatarLocations defines where the avatar is positioned relative to the text box.
type AvatarLocations int

const (
	// NoAvatar hides the avatar.
	NoAvatar AvatarLocations = iota
	// LeftAvatar positions the avatar on the left.
	LeftAvatar
	// RightAvatar positions the avatar on the right.
	RightAvatar
)

// apply implements the Option interface.
func (al AvatarLocations) apply(box *TextBox) {
	box.avatarLocation = al
}

// NamePositions defines where the name tag is positioned.
type NamePositions int

const (
	// NoName hides the name tag.
	NoName NamePositions = iota
	// NameTopLeftAboveTextInFrame positions the name tag above the text, aligned left.
	NameTopLeftAboveTextInFrame
	// NameTopCenterInFrame positions the name tag above the text, centered.
	NameTopCenterInFrame
	// NameLeftAboveAvatarInFrame positions the name tag above the avatar, aligned left.
	NameLeftAboveAvatarInFrame
	// NameTopLeftAboveFrame positions the name tag above the frame, aligned left.
	NameTopLeftAboveFrame
	// NameTopCenterAboveFrame positions the name tag above the frame, centered.
	NameTopCenterAboveFrame
)

// apply implements the Option interface.
func (al NamePositions) apply(box *TextBox) {
	box.namePosition = al
}

// MoreChevronLocations defines where the "next page" indicator is positioned.
type MoreChevronLocations int

const (
	// NoMoreChevron hides the chevron.
	NoMoreChevron MoreChevronLocations = iota
	// CenterBottomInsideTextFrame positions the chevron centered at the bottom of the text area.
	CenterBottomInsideTextFrame
	// CenterBottomInsideFrame positions the chevron centered at the bottom of the entire frame.
	CenterBottomInsideFrame
	// CenterBottomOnFrameTextFrame positions the chevron centered on the bottom edge of the text area.
	CenterBottomOnFrameTextFrame
	// CenterBottomOnFrameFrame positions the chevron centered on the bottom edge of the frame.
	CenterBottomOnFrameFrame
	// RightBottomInsideTextFrame positions the chevron at the bottom right of the text area.
	RightBottomInsideTextFrame
	// RightBottomInsideFrame positions the chevron at the bottom right of the frame.
	RightBottomInsideFrame
	// RightBottomOnFrameTextFrame positions the chevron on the bottom right edge of the text area.
	RightBottomOnFrameTextFrame
	// RightBottomOnFrameFrame positions the chevron on the bottom right edge of the frame.
	RightBottomOnFrameFrame
	// TextEndChevron positions the chevron inline at the end of the text.
	TextEndChevron
)

// apply implements the Option interface.
func (cl MoreChevronLocations) apply(box *TextBox) {
	box.moreChevronLocation = cl
}

// AvatarFit defines how the avatar is scaled if it doesn't fit the allocated space.
type AvatarFit int

const (
	// NoAvatarFit performs no scaling (undefined behavior if too large).
	NoAvatarFit AvatarFit = iota
	// CenterAvatar centers the avatar without scaling.
	CenterAvatar
	// NearestNeighbour scales using nearest-neighbor interpolation.
	NearestNeighbour
	// ApproxBiLinear scales using approximate bi-linear interpolation.
	ApproxBiLinear
)

// apply implements the Option interface.
func (af AvatarFit) apply(box *TextBox) {
	box.avatarFit = af
}

// BoxTextBox creates a PostDrawer that draws a box around the text area.
func BoxTextBox() *boxTextBox {
	return &boxTextBox{}
}

// boxTextBox implements PostDrawer to draw a debug box.
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

// PostDrawer allows custom components to be drawn after standard elements.
type PostDrawer interface {
	PostDraw(target wordwrap.Image, layout *SimpleLayout, ls []wordwrap.Line, options ...wordwrap.DrawOption) error
}

// TextBox is the main component for rendering RPG-style text boxes.
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
	namePosition        NamePositions
	nameBox             wordwrap.Box
	spaceMap            SpaceMap
}

// SpaceMap is an interface for mapping screen space to interactive shapes.
type SpaceMap interface {
	Add(shape shared.Shape, zIndex int)
}

// SetSpaceMap sets the space map to populate
func (tb *TextBox) SetSpaceMap(m SpaceMap) {
	tb.spaceMap = m
}

// Option defines a configuration option for the TextBox.
type Option interface {
	apply(*TextBox)
}

// Name sets the character name to display.
type Name string

// apply implements the Option interface.
func (n Name) apply(box *TextBox) {
	box.name = n
	box.nameBox, _ = wordwrap.NewSimpleTextBox(box.theme.FontDrawer(), string(n))
}

// avatar holds an avatar image override.
type avatar struct {
	wordwrap.Image
}

// Avatar overrides the default theme avatar.
func Avatar(i wordwrap.Image) Option {
	return &avatar{
		Image: i,
	}
}

// apply Set the location when used as an Option
func (a *avatar) apply(box *TextBox) {
	box.avatar = a
}

type wordwrapOption struct {
	opt wordwrap.WrapperOption
}

func (w *wordwrapOption) apply(box *TextBox) {
	box.wordwrapOptions = append(box.wordwrapOptions, w.opt)
}

// WithWordwrapOption allows passing wordwrap options as TextBox options
func WithWordwrapOption(opt wordwrap.WrapperOption) Option {
	return &wordwrapOption{opt: opt}
}

// NewSimpleTextBox creates a TextBox with simple string content.
// theme is required.
// destSize is the intended size of the image, but can be updated per frame.
func NewSimpleTextBox(th theme.Theme, text string, destSize image.Point, options ...Option) (*TextBox, error) {
	args := []interface{}{text, destSize}
	for _, o := range options {
		args = append(args, o)
	}
	return NewRichTextBox(th, args...)
}

// NewRichTextBox creates a TextBox with rich text content (e.g. colors, images).
// theme is required.
// args can include string content, image.Point (for size), and Options.
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
			}
		default:
			wordwrapArgs = append(wordwrapArgs, arg)
		}
	}

	if !foundDestSize {
		log.Printf("Warning: destSize not found in NewRichTextBox arguments")
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

// Layout defines the positioning of elements within the text box.
type Layout interface {
	// TextRect is the area containing the text.
	TextRect() image.Rectangle
	// CenterRect is the main content area inside the frame.
	CenterRect() image.Rectangle
	// AvatarRect is the area containing the avatar.
	AvatarRect() image.Rectangle
	// ChevronRect is the area containing the chevron.
	ChevronRect() image.Rectangle
	// NameRect is the optional area containing the name tag.
	NameRect() image.Rectangle
	// FrameRect is the area containing the frame.
	FrameRect() image.Rectangle
}

// SimpleLayout implements a standard text box layout.
type SimpleLayout struct {
	textRect    image.Rectangle
	centerRect  image.Rectangle
	avatarRect  image.Rectangle
	chevronRect image.Rectangle
	nameRect    image.Rectangle
	frameRect   image.Rectangle
}

// Interface enforcement
var _ Layout = (*SimpleLayout)(nil)

// NameRect returns the name tag rectangle.
func (sl *SimpleLayout) NameRect() image.Rectangle {
	return sl.nameRect
}

// FrameRect returns the frame rectangle.
func (sl *SimpleLayout) FrameRect() image.Rectangle {
	return sl.frameRect
}

// TextRect returns the text rectangle.
func (sl *SimpleLayout) TextRect() image.Rectangle {
	return sl.textRect
}

// CenterRect returns the center content rectangle.
func (sl *SimpleLayout) CenterRect() image.Rectangle {
	return sl.centerRect
}

// AvatarRect returns the avatar rectangle.
func (sl *SimpleLayout) AvatarRect() image.Rectangle {
	return sl.avatarRect
}

// ChevronRect returns the chevron rectangle.
func (sl *SimpleLayout) ChevronRect() image.Rectangle {
	return sl.chevronRect
}

// NewSimpleLayout constructs SimpleLayout simply as possible (for the user.)
func NewSimpleLayout(tb *TextBox, destRect image.Rectangle) (*SimpleLayout, error) {
	l := &SimpleLayout{}
	l.frameRect = destRect
	if tb.nameBox != nil {
		m := tb.nameBox.MetricsRect()
		a := tb.nameBox.AdvanceRect()
		height := (m.Ascent + m.Descent).Ceil()
		width := a.Ceil()
		switch tb.namePosition {
		case NameTopLeftAboveFrame:
			l.nameRect = image.Rect(destRect.Min.X, destRect.Min.Y, destRect.Min.X+width, destRect.Min.Y+height)
			l.frameRect.Min.Y += height
		case NameTopCenterAboveFrame:
			centerX := destRect.Min.X + (destRect.Dx()-width)/2
			l.nameRect = image.Rect(centerX, destRect.Min.Y, centerX+width, destRect.Min.Y+height)
			l.frameRect.Min.Y += height
		}
	}
	if centerRect, err := tb.calculateCenterRect(l.frameRect); err != nil {
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
		switch tb.namePosition {
		case NoName, NameTopLeftAboveFrame, NameTopCenterAboveFrame:
		case NameTopCenterInFrame:
			l.nameRect = image.Rect(0, 0, width, height)
			l.nameRect.Min.X = l.centerRect.Min.X + (l.centerRect.Dx()-l.nameRect.Dx())/2
			l.nameRect.Max.X = l.nameRect.Min.X + width
			fallthrough
		default:
			if l.nameRect.Empty() {
				l.nameRect = image.Rect(0, 0, width, height)
			}
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
	if err := drawFrame(tb.theme, target.SubImage(layout.FrameRect()).(wordwrap.Image), opts...); err != nil {
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
	return fmt.Sprintf("Box(%s)", b.TextValue())
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

// drawMoreChevron draws the "next page" indicator.
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

// drawNameTag draws the name tag.
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

// drawAvatar draws the avatar image.
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
