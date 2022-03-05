package main

import (
	_ "embed"
	"flag"
	"github.com/arran4/golang-rpg-textbox"
	"github.com/arran4/golang-rpg-textbox/theme/cache"
	"github.com/arran4/golang-rpg-textbox/theme/simple"
	"github.com/arran4/golang-rpg-textbox/util"
	wordwrap "github.com/arran4/golang-wordwrap"
	"image"
	"log"
	"path/filepath"
	"strings"
	"sync"
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
	t, err := cache.New(simple.New())
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
	OptionDescriptionBuild(func(oas []string, oa []rpgtextbox.Option) {
		addTextBox(strings.Join(oas, "+")+".png", Must(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, oa...)))
	}, []string{}, []rpgtextbox.Option{}, chevronLocs, avatarPos, avatarScale)
	pos := image.Rect(0, 0, *width, *height)
	wg := sync.WaitGroup{}
	for i := range points {
		wg.Add(1)
		go func(tb *TextBox) {
			defer wg.Done()
			tb.Render(maxPages, pos)
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

func (tb *TextBox) Render(maxPages int, pos image.Rectangle) {
	i := image.NewRGBA(image.Rect(0, 0, *width, *height*maxPages))
	for page := 0; page < maxPages; page++ {
		if tb.pages > page {
			target := pos.Add(image.Pt(0, *height*page))
			if _, err := tb.rtb.DrawNextPageFrame(i.SubImage(target).(wordwrap.Image)); err != nil {
				log.Panicf("Draw next frame error: %s", err)
			}
		}
	}
	ofn := filepath.Join(*outdir, tb.Filename)
	if err := util.SavePngFile(i, ofn); err != nil {
		log.Panicf("Error with saving file: %s", err)
	}
	log.Printf("Saving %s", ofn)
}
