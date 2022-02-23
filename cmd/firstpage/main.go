package main

import (
	"errors"
	"flag"
	"fmt"
	rpgtextbox "github.com/arran4/golang-rpg-textbox"
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
	rtb, err := rpgtextbox.NewSimpleTextBox(t, text, textBoxSize)
	if err != nil {
		log.Panicf("Text fetch error: %s", err)
	}
	pages, err := rtb.CalculateAllPages(textBoxSize)
	if err != nil {
		log.Panicf("Text calculate error: %s", err)
	}

	i := image.NewRGBA(image.Rect(0, 0, *width, *height*pages))
	pos := image.Rect(0, 0, *width, *height)
	for page := 0; page < pages; page++ {
		log.Printf("Rendering page %d or %d at %s", page, pages, pos)
		if _, err := rtb.DrawNextPageFrame(i.SubImage(pos).(rpgtextbox.Image)); err != nil {
			log.Panicf("Draw next frame error: %s", err)
		}
		pos = pos.Add(image.Pt(0, textBoxSize.Y))
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
