package main

import (
	_ "embed"
	"errors"
	"flag"
	"fmt"
	"github.com/arran4/golang-rpg-textbox"
	"github.com/arran4/golang-rpg-textbox/theme/simple"
	"github.com/arran4/golang-rpg-textbox/util"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
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
		text, err = GetText(*textsource)
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
		pages, t := NewTextBox(filename, tb, err, textBoxSize)
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
				oas = append(oas, o2.Description)
				addTextBox(strings.Join(oas, "+")+".png", Must(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, oa...)))
			}
		}
	}

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

func NewTextBox(filename string, tb *rpgtextbox.TextBox, err error, textBoxSize image.Point) (int, *TextBox) {
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
			if _, err := tb.rtb.DrawNextPageFrame(i.SubImage(target).(util.Image)); err != nil {
				log.Panicf("Draw next frame error: %s", err)
			}
		}
	}
	ofn := filepath.Join(*outdir, tb.Filename)
	if err := util.SaveFile(i, ofn); err != nil {
		log.Panicf("Error with saving file: %s", err)
	}
	log.Printf("Saving %s", ofn)
}

func GetText(fn string) (string, error) {
	if fn == "" {
		return "", errors.New("no input file specified")
	}
	if fn == "-" {
		log.Printf("STDIN mode, enter text plaese")
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", fmt.Errorf("reading file %s: %w", fn, err)
	}
	return string(b), nil
}
