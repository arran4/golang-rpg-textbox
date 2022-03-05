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
	type OptionDescription struct {
		Options     []rpgtextbox.Option
		Description string
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
	for _, o1 := range chevronLocs {
		oa := append([]rpgtextbox.Option{}, o1.Options...)
		oas := append([]string{}, o1.Description)
		o1p := len(oa)
		o1sp := len(oas)
		for _, o2 := range avatarPos {
			oa = oa[:o1p]
			oas = oas[:o1sp]
			oa = append(oa, o2.Options...)
			oas = append(oas, o2.Description)
			o2p := len(oa)
			o2sp := len(oas)
			for _, o3 := range avatarScale {
				oa = oa[:o2p]
				oas = oas[:o2sp]
				oa = append(oa, o3.Options...)
				oas = append(oas, o3.Description)
				o3p := len(oa)
				o3sp := len(oas)
				for _, o4 := range animations {
					oa = oa[:o3p]
					oas = oas[:o3sp]
					oa = append(oa, o4.Options...)
					oas = append(oas, o4.Description)
					addTextBox(strings.Join(oas, "+")+".gif", Must(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, oa...)))
				}
			}
		}
	}

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
