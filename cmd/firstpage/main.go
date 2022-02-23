package main

import (
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
)

var (
	width       = flag.Int("width", 600, "Doc width")
	height      = flag.Int("height", 150, "Doc height")
	textsource  = flag.String("text", "-", "File in, or - for std input")
	outfilename = flag.String("out", "out.png", "file to write to, in some cases this is ignored")
)

func init() {
	log.SetFlags(log.Flags() | log.Lshortfile)
}

func main() {
	flag.Parse()
	log.Printf("Starting")
	textBoxSize := image.Pt(*width, *height)
	text, err := GetText(*textsource)
	if err != nil {
		log.Panicf("Text fetch error: %s", err)
	}
	t, err := simple.New()
	if err != nil {
		log.Panicf("Theme fetch error: %s", err)
	}
	type TextBox struct {
		rtb   *rpgtextbox.TextBox
		pages int
	}
	var points []*TextBox
	var maxPages int
	addPoint := func(tb *rpgtextbox.TextBox, err error) {
		if err != nil {
			log.Panicf("Text fetch error: %s", err)
		}
		pages, err := tb.CalculateAllPages(textBoxSize)
		if err != nil {
			log.Panicf("Text fetch error: %s", err)
		}
		if pages > maxPages {
			maxPages = pages
		}
		points = append(points, &TextBox{
			rtb:   tb,
			pages: pages,
		})
	}
	addPoint(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize))
	addPoint(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, rpgtextbox.CenterLeft, rpgtextbox.CenterAvatar))
	addPoint(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, rpgtextbox.CenterRight))
	addPoint(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, rpgtextbox.CenterLeft))
	addPoint(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, rpgtextbox.CenterRight, rpgtextbox.NearestNeighbour))
	addPoint(rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, rpgtextbox.CenterLeft, rpgtextbox.ApproxBiLinear))

	const columns = 2

	i := image.NewRGBA(image.Rect(0, 0, *width*columns, (*height)*maxPages*(len(points)/columns+(columns-1+len(points)%columns)/columns)))
	pos := image.Rect(0, 0, *width, *height)
	for page := 0; page < maxPages; page++ {
		for tbi, tb := range points {
			if tb.pages > page {
				target := pos.Add(image.Pt(*width*(tbi%columns), *height*(maxPages*(tbi/columns)+page)))
				if _, err := tb.rtb.DrawNextPageFrame(i.SubImage(target).(rpgtextbox.Image)); err != nil {
					log.Panicf("Draw next frame error: %s", err)
				}
			}
		}
	}
	if err := util.SaveFile(i, *outfilename); err != nil {
		log.Panicf("Error with saving file: %s", err)
	}
	log.Printf("Done as %s", *outfilename)
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
