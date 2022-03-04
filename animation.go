package rpgtextbox

import (
	"github.com/arran4/golang-wordwrap"
	"image"
	"image/color"
	"time"
)

type AnimationMode interface {
	DrawOption(target wordwrap.Image) (lastPage bool, userInputAccepted bool, wait time.Duration, err error)
	Option
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

type BoxByBoxAnimation struct {
	tb           *TextBox
	boxNumber    int
	layout       *SimpleLayout
	page         *Page
	WaitTimeFunc func(*BoxByBoxAnimation) time.Duration
}

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

	done, err = byb.tb.drawFrame(target, byb.layout, byb.page, opts...)

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

func (byb *BoxByBoxAnimation) apply(box *TextBox) {
	byb.tb = box
	box.animation = byb
}

var _ AnimationMode = (*BoxByBoxAnimation)(nil)

func NewBoxByBoxAnimation() *BoxByBoxAnimation {
	return &BoxByBoxAnimation{
		WaitTimeFunc: func(byb *BoxByBoxAnimation) time.Duration {
			return time.Second / 10
		},
	}
}
