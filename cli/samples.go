package cli

import (
	_ "embed"
	"fmt"
	"image"
	"image/color/palette"
	"image/gif"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	rpgtextbox "github.com/arran4/golang-rpg-textbox"
	"github.com/arran4/golang-rpg-textbox/theme/cache"
	"github.com/arran4/golang-rpg-textbox/theme/simple"
	"github.com/arran4/golang-rpg-textbox/util"
	wordwrap "github.com/arran4/golang-wordwrap"
	"golang.org/x/image/draw"
)

//go:embed "sample1.txt"
var embeddedtext []byte

type SampleTextBox struct {
	Filename string
	rtb      *rpgtextbox.TextBox
	pages    int
}

func Must(tb *rpgtextbox.TextBox, err error) *rpgtextbox.TextBox {
	if err != nil {
		log.Panicf("Text fetch error: %s", err)
	}
	return tb
}

func NewSampleTextBox(filename string, tb *rpgtextbox.TextBox, textBoxSize image.Point) (int, *SampleTextBox) {
	pages, err := tb.CalculateAllPages(textBoxSize)
	if err != nil {
		log.Panicf("Text fetch error: %s", err)
	}
	t := &SampleTextBox{
		Filename: filename,
		rtb:      tb,
		pages:    pages,
	}
	return pages, t
}

func (tb *SampleTextBox) RenderAnimation(width, height int, outdir string) {
	gifo := &gif.GIF{}
	f := 0
	page := 0
	for {
		i := image.NewRGBA(image.Rect(0, 0, width, height))
		if done, ui, w, err := tb.rtb.DrawNextFrame(i); err != nil {
			log.Panicf("Draw next frame error: %s", err)
		} else if done && !ui && w <= 0 {
			break
		} else {
			if w <= 0 {
				w = time.Second / 2
			}
			f++
			if ui && w <= 0 {
				page++
			}
			log.Printf("%s: Adding frame %d for page %d", tb.Filename, f, page)
			bounds := i.Bounds()
			palettedImage := image.NewPaletted(bounds, palette.Plan9)
			draw.Draw(palettedImage, palettedImage.Rect, i, bounds.Min, draw.Over)
			gifo.Image = append(gifo.Image, palettedImage)
			gifo.Delay = append(gifo.Delay, int(w/(time.Second/100)))
		}
	}
	ofn := filepath.Join(outdir, tb.Filename)
	log.Printf("Saving %s", ofn)
	if err := util.SaveGifFile(ofn, gifo); err != nil {
		log.Panicf("Error with saving file: %s", err)
	}
	log.Printf("Saved %s", ofn)
}

func (tb *SampleTextBox) RenderStatic(maxPages int, pos image.Rectangle, width, height int, outdir string) {
	i := image.NewRGBA(image.Rect(0, 0, width, height*maxPages))
	for page := 0; page < maxPages; page++ {
		if tb.pages > page {
			target := pos.Add(image.Pt(0, height*page))
			if _, err := tb.rtb.DrawNextPageFrame(i.SubImage(target).(wordwrap.Image)); err != nil {
				log.Panicf("Draw next frame error: %s", err)
			}
		}
	}
	ofn := filepath.Join(outdir, tb.Filename)
	if err := util.SavePngFile(i, ofn); err != nil {
		log.Panicf("Error with saving file: %s", err)
	}
	log.Printf("Saving %s", ofn)
}

type OptionDescription struct {
	Options     []rpgtextbox.Option
	Description string
}

func OptionDescriptionBuild(addTextBox func(oas []string, oa []rpgtextbox.Option), oas []string, oa []rpgtextbox.Option, ods ...[]*OptionDescription) {
	if len(ods) == 0 {
		addTextBox(oas, oa)
		return
	}
	oap := len(oa)
	oasp := len(oas)
	for _, o := range ods[0] {
		oa = oa[:oap]
		oas = oas[:oasp]
		oa = append(oa, o.Options...)
		oas = append(oas, o.Description)
		OptionDescriptionBuild(addTextBox, oas[:], oa[:], ods[1:]...)
	}
}

// GenerateAnimationSamples is a subcommand `rpgtextbox samples animation`
//
// Flags:
//
//	width:      --width      (default: 600)      Doc width
//	height:     --height     (default: 150)      Doc height
//	textSource: --text       (default: "")       File in, or - for std input
//	outDir:     --outdir     (default: "images/") directory to save samples to
func GenerateAnimationSamples(width, height int, textSource, outDir string) error {
	log.Printf("Starting")
	textBoxSize := image.Pt(width, height)
	var text string
	if textSource == "" {
		text = string(embeddedtext)
	} else {
		var err error
		text, err = util.GetText(textSource)
		if err != nil {
			return fmt.Errorf("text fetch error: %w", err)
		}
	}
	t, err := cache.New(simple.New())
	if err != nil {
		return fmt.Errorf("theme fetch error: %w", err)
	}
	var points []*SampleTextBox
	var maxPages int
	addTextBox := func(filename string, tb *rpgtextbox.TextBox) {
		pages, t := NewSampleTextBox(filename, tb, textBoxSize)
		if pages > maxPages {
			maxPages = pages
		}
		points = append(points, t)
	}
	chevronLocs := []*OptionDescription{
		{
			Options:     []rpgtextbox.Option{rpgtextbox.TextEndChevron},
			Description: "end-of-text-chevron",
		},
	}
	avatarPos := []*OptionDescription{
		{
			Options:     []rpgtextbox.Option{rpgtextbox.LeftAvatar},
			Description: "left-avatar",
		},
	}
	avatarScale := []*OptionDescription{
		{
			Options:     []rpgtextbox.Option{rpgtextbox.CenterAvatar},
			Description: "center-avatar",
		},
	}
	animations := []*OptionDescription{
		{
			Options:     []rpgtextbox.Option{rpgtextbox.NewFadeAnimation()},
			Description: "fade-animation",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.NewBoxByBoxAnimation()},
			Description: "box-by-box-animation",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.NewLetterByLetterAnimation()},
			Description: "letter-by-letter-animation",
		},
	}
	OptionDescriptionBuild(func(oas []string, oa []rpgtextbox.Option) {
		addTextBox(strings.Join(oas, "+")+".gif", Must(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, oa...)))
	}, []string{}, []rpgtextbox.Option{}, chevronLocs, avatarPos, avatarScale, animations)
	wg := sync.WaitGroup{}
	for i := range points {
		wg.Add(1)
		go func(tb *SampleTextBox) {
			defer wg.Done()
			tb.RenderAnimation(width, height, outDir)
		}(points[i])
	}
	wg.Wait()
	log.Printf("Done")
	return nil
}

// GenerateSamples is a subcommand `rpgtextbox samples static`
//
// Flags:
//
//	width:      --width      (default: 600)      Doc width
//	height:     --height     (default: 150)      Doc height
//	textSource: --text       (default: "")       File in, or - for std input
//	outDir:     --outdir     (default: "images/") directory to save samples to
func GenerateSamples(width, height int, textSource, outDir string) error {
	log.Printf("Starting")
	textBoxSize := image.Pt(width, height)
	var text string
	if textSource == "" {
		text = string(embeddedtext)
	} else {
		var err error
		text, err = util.GetText(textSource)
		if err != nil {
			return fmt.Errorf("text fetch error: %w", err)
		}
	}
	t, err := cache.New(simple.New())
	if err != nil {
		return fmt.Errorf("theme fetch error: %w", err)
	}
	var points []*SampleTextBox
	var maxPages int
	addTextBox := func(filename string, tb *rpgtextbox.TextBox) {
		pages, t := NewSampleTextBox(filename, tb, textBoxSize)
		if pages > maxPages {
			maxPages = pages
		}
		points = append(points, t)
	}
	chevronLocs := []*OptionDescription{
		{
			Options:     nil,
			Description: "no-chevron",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.CenterBottomInsideTextFrame},
			Description: "center-bottom-chevron",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.CenterBottomInsideFrame},
			Description: "center-bottom-inside-chevron",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.CenterBottomOnFrameTextFrame},
			Description: "center-bottom-on-frame-text-chevron",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.CenterBottomOnFrameFrame},
			Description: "center-bottom-on-frame-chevron",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.RightBottomInsideTextFrame},
			Description: "right-bottom-inside-text-chevron",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.RightBottomInsideFrame},
			Description: "right-bottom-inside-chevron",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.RightBottomOnFrameTextFrame},
			Description: "right-bottom-on-frame-text-chevron",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.RightBottomOnFrameFrame},
			Description: "right-bottom-on-frame-chevron",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.TextEndChevron},
			Description: "end-of-text-chevron",
		},
	}
	avatarPos := []*OptionDescription{
		{
			Options:     nil,
			Description: "no-avatar",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.LeftAvatar},
			Description: "left-avatar",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.RightAvatar},
			Description: "right-avatar",
		},
	}
	avatarScale := []*OptionDescription{
		{
			Options:     nil,
			Description: "no-scaling",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.CenterAvatar},
			Description: "center-avatar",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.NearestNeighbour},
			Description: "nearest-neighbour",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.ApproxBiLinear},
			Description: "approx-biLinear",
		},
	}
	namePos := []*OptionDescription{
		{
			Options:     []rpgtextbox.Option{rpgtextbox.Name("Player 1"), rpgtextbox.NameTopLeftAboveTextInFrame},
			Description: "name-top-left-text",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.Name("Player 1"), rpgtextbox.NameTopCenterInFrame},
			Description: "name-top-center",
		},
		{
			Options:     []rpgtextbox.Option{rpgtextbox.Name("Player 1"), rpgtextbox.NameLeftAboveAvatarInFrame},
			Description: "name-left-above-avatar",
		},
	}
	OptionDescriptionBuild(func(oas []string, oa []rpgtextbox.Option) {
		addTextBox(strings.Join(oas, "+")+".png", Must(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, oa...)))
	}, []string{}, []rpgtextbox.Option{}, chevronLocs, avatarPos, avatarScale)
	OptionDescriptionBuild(func(oas []string, oa []rpgtextbox.Option) {
		addTextBox(strings.Join(oas, "+")+".png", Must(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, oa...)))
	}, []string{}, []rpgtextbox.Option{}, chevronLocs[len(chevronLocs)-2:], avatarPos, avatarScale[1:2], namePos)
	pos := image.Rect(0, 0, width, height)
	wg := sync.WaitGroup{}
	for i := range points {
		wg.Add(1)
		go func(tb *SampleTextBox) {
			defer wg.Done()
			tb.RenderStatic(maxPages, pos, width, height, outDir)
		}(points[i])
	}
	wg.Wait()
	log.Printf("Done")
	return nil
}
