package main

import (
	_ "embed"
	"flag"
	"github.com/arran4/golang-rpg-textbox"
	"github.com/arran4/golang-rpg-textbox/theme/simple"
	"github.com/arran4/golang-rpg-textbox/util"
	"golang.org/x/image/draw"
	"image"
	"image/color/palette"
	"image/gif"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	width      = flag.Int("width", 600, "Doc width")
	height     = flag.Int("height", 150, "Doc height")
	textsource = flag.String("text", "", "File in, or - for std input")
	outdir     = flag.String("outdir", "images/", "directory to save samples to")
	//go:embed "sample1.txt"
	embeddedtext []byte
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

type TextBox struct {
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

func main() {
	flag.Parse()
	log.Printf("Starting")
	textBoxSize := image.Pt(*width, *height)
	var text string
	if *textsource == "" {
		text = string(embeddedtext)
	} else {
		var err error
		text, err = util.GetText(*textsource)
		if err != nil {
			log.Panicf("Text fetch error: %s", err)
		}
	}
	t, err := simple.New()
	if err != nil {
		log.Panicf("Theme fetch error: %s", err)
	}
	var points []*TextBox
	var maxPages int
	addTextBox := func(filename string, tb *rpgtextbox.TextBox) {
		pages, t := NewTextBox(filename, tb, textBoxSize)
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
		go func(tb *TextBox) {
			defer wg.Done()
			tb.Render()
		}(points[i])
	}
	wg.Wait()
	log.Printf("Done")
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

func NewTextBox(filename string, tb *rpgtextbox.TextBox, textBoxSize image.Point) (int, *TextBox) {
	pages, err := tb.CalculateAllPages(textBoxSize)
	if err != nil {
		log.Panicf("Text fetch error: %s", err)
	}
	t := &TextBox{
		Filename: filename,
		rtb:      tb,
		pages:    pages,
	}
	return pages, t
}

func (tb *TextBox) Render() {
	gifo := &gif.GIF{}
	f := 0
	page := 0
	for {
		i := image.NewRGBA(image.Rect(0, 0, *width, *height))
		if done, ui, w, err := tb.rtb.DrawNextFrame(i); err != nil {
			log.Panicf("Draw next frame error: %s", err)
		} else if done && !ui && w <= 0 {
			break
			//} else if ui {
			//	break
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
	ofn := filepath.Join(*outdir, tb.Filename)
	log.Printf("Saving %s", ofn)
	if err := util.SaveGifFile(ofn, gifo); err != nil {
		log.Panicf("Error with saving file: %s", err)
	}
	log.Printf("Saved %s", ofn)
}
