package dynamic

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/arran4/go-pattern/dsl"
	pattern_cli "github.com/arran4/go-pattern/pkg/pattern-cli"
	"github.com/arran4/golang-frame/frames"
	"github.com/arran4/golang-rpg-textbox/theme"
	"github.com/arran4/golang-rpg-textbox/theme/cache"
	"golang.org/x/image/font"
)

type t struct {
	cache.Source
	frameName  string
	patternStr string
}

func New(source cache.Source, frameName, patternStr string) *t {
	return &t{
		Source:     source,
		frameName:  frameName,
		patternStr: patternStr,
	}
}

var _ theme.Theme = (*t)(nil)
var _ theme.Frame = (*t)(nil)

func (t *t) Frame() image.Image {
	var fImg image.Image
	if t.frameName != "" {
		if def, ok := frames.ByName[t.frameName]; ok {
			fImg = def.Image
		} else {
			panic(fmt.Sprintf("unknown frame: %s", t.frameName))
		}
	} else {
		fImg = t.Source.Frame()
	}

	if t.patternStr != "" {
		p, err := dsl.Parse(t.patternStr)
		if err != nil {
			panic(fmt.Errorf("invalid pattern %q: %w", t.patternStr, err))
		}
		fm := make(dsl.FuncMap)
		pattern_cli.RegisterGeneratedCommands(fm)

		bounds := fImg.Bounds()
		bgImg, err := p.Execute(fm, image.NewRGBA(bounds))
		if err != nil {
			panic(fmt.Errorf("failed to execute pattern: %w", err))
		}

		result := image.NewRGBA(bounds)
		// Draw pattern
		draw.Draw(result, bounds, bgImg, bgImg.Bounds().Min, draw.Src)
		// Overlay frame
		draw.Draw(result, bounds, fImg, bounds.Min, draw.Over)
		return result
	}

	return fImg
}

func (t *t) FrameCenter() image.Rectangle {
	if t.frameName != "" {
		if def, ok := frames.ByName[t.frameName]; ok {
			return def.Middle
		}
	}
	return t.Source.FrameCenter()
}
func (t *t) FontDrawer() *font.Drawer {
	fd := t.Source.FontDrawer()
	// Change font color for better readability on generated pattern backgrounds
	if t.patternStr != "" {
		fd.Src = image.NewUniform(color.RGBA{240, 240, 240, 255})
	}
	return fd
}
