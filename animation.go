package rpgtextbox

import (
	"github.com/arran4/golang-wordwrap"
	"image"
	"image/color"
	"time"
)

// AnimationMode interface defining the optional animation. Used as an Option
type AnimationMode interface {
	// DrawOption draws with options.. Controls the drawing process to add extra frames and a delay to create an animation
	// finished is true if you're on the last page
	// userInputAccepted is if it's at the stage where you would typically accept user input (ie the animation is waiting
	// user input, doesn't imply anything to do with the animation
	// wait is either 0 or less, or the amount of time before the next animation phase
	// err is err
	// To determine if you're at the end the only way of doing it as of writing is to wait for; lastPage = true,
	// userInputAccepted = false, wait = -1
	DrawOption(target wordwrap.Image) (lastPage bool, userInputAccepted bool, wait time.Duration, err error)
	Option
}

// AlphaSourceImageMapper is a draw.Image compatible source image, that allows an image to fade.
type AlphaSourceImageMapper struct {
	// original image
	image.Image
	// Multiplier How much to "fade" it by
	Multiplier float64
}

// At a possible wrong implementation of fading a text box
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

// NewAlphaSourceImageMapper Creates a proxy image which will provide a source
func NewAlphaSourceImageMapper(i image.Image, multiplier float64) image.Image {
	return &AlphaSourceImageMapper{
		i,
		multiplier,
	}
}

// FadeState is the current picture fading in or out
type FadeState int

const (
	// FadeIn picture is going to fade in or is fading in
	FadeIn FadeState = iota
	// FadeOut picture is going to fade out or is fading out
	FadeOut
)

// FadeAnimation The animation for fading.
type FadeAnimation struct {
	tb        *TextBox
	fadeState FadeState
	duration  time.Duration
	steps     int
	step      int
	layout    *SimpleLayout
	page      *Page
}

// DrawOption draws with options.. Controls the drawing process to add extra frames, a wait time and more
// finished is true if you're on the last page
// userInputAccepted is if it's at the stage where you would typically accept user input (ie the animation is waiting
// user input, doesn't imply anything to do with the animation
// wait is either 0 or less, or the amount of time before the next animation phase
// err is err
// To determine if you're at the end the only way of doing it as of writing is to wait for; lastPage = true,
// userInputAccepted = false, wait = -1
func (f *FadeAnimation) DrawOption(target wordwrap.Image) (finished bool, userInputAccepted bool, waitTime time.Duration, err error) {
	if f.layout == nil {
		f.layout, f.page, err = f.tb.getNextPage(target.Bounds())
		if err != nil {
			return
		}
		if f.layout == nil || f.page == nil {
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
		done, err = f.tb.drawPage(target, f.layout, f.page, opts...)
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

// apply Set the location when used as an Option
func (f *FadeAnimation) apply(box *TextBox) {
	f.tb = box
	box.animation = f
}

// forces implementation
var _ AnimationMode = (*FadeAnimation)(nil)

// NewFadeAnimation constructs FadeAnimation
func NewFadeAnimation() *FadeAnimation {
	duration := 2 * time.Second
	return &FadeAnimation{
		duration: duration,
		steps:    int(duration * 10 / time.Second),
	}
}

// BoxByBoxAnimation is an animation style in which each non-whitespace box comes into visibility one by one
type BoxByBoxAnimation struct {
	tb        *TextBox
	boxNumber int
	layout    *SimpleLayout
	page      *Page
	// The function to calculate the wait time between each box
	WaitTimeFunc func(*BoxByBoxAnimation) time.Duration
}

// DrawOption draws with options.. Controls the drawing process to add extra frames, a wait time and more
// finished is true if you're on the last page
// userInputAccepted is if it's at the stage where you would typically accept user input (ie the animation is waiting
// user input, doesn't imply anything to do with the animation
// wait is either 0 or less, or the amount of time before the next animation phase
// err is err
// To determine if you're at the end the only way of doing it as of writing is to wait for; lastPage = true,
// userInputAccepted = false, wait = -1
func (byb *BoxByBoxAnimation) DrawOption(target wordwrap.Image) (finished bool, userInputAccepted bool, waitTime time.Duration, err error) {
	if byb.layout == nil {
		byb.layout, byb.page, err = byb.tb.getNextPage(target.Bounds())
		if err != nil {
			return
		}
		if byb.layout == nil || byb.page == nil {
			finished = true
			waitTime = -1
			return
		}
	}
	var done bool
	var opts []wordwrap.DrawOption

	if byb.boxNumber != byb.page.boxCount {
		opts = append(opts, wordwrap.BoxDrawMap(func(box wordwrap.Box, drawConfig *wordwrap.DrawConfig, stats *wordwrap.BoxPositionStats) wordwrap.Box {
			if stats.PageBoxOffset == byb.boxNumber {
				if box.Whitespace() || box.Len() == 0 {
					byb.boxNumber++
				}
			}
			if stats.PageBoxOffset < byb.boxNumber {
				return box
			}
			return nil
		}))
	}

	done, err = byb.tb.drawPage(target, byb.layout, byb.page, opts...)

	if byb.boxNumber > byb.page.boxCount {
		finished = done
		userInputAccepted = true
		byb.boxNumber = 0
		byb.layout = nil
		waitTime = -1
	} else {
		if byb.WaitTimeFunc != nil {
			waitTime = byb.WaitTimeFunc(byb)
		} else {
			waitTime = time.Second / 10
		}
		byb.boxNumber++
	}
	return
}

// apply Set the location when used as an Option
func (byb *BoxByBoxAnimation) apply(box *TextBox) {
	byb.tb = box
	box.animation = byb
}

// Enforce the interface
var _ AnimationMode = (*BoxByBoxAnimation)(nil)

// NewBoxByBoxAnimation creates an animation style where one block comes on at one time. Use WaitTimeFunc to create
// your own timing for each block
func NewBoxByBoxAnimation() *BoxByBoxAnimation {
	return &BoxByBoxAnimation{
		WaitTimeFunc: func(byb *BoxByBoxAnimation) time.Duration {
			return time.Second / 10
		},
	}
}
