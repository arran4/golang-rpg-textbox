package cli

import (
	"fmt"
	"github.com/arran4/golang-rpg-textbox"
	"github.com/arran4/golang-rpg-textbox/theme/cache"
	"github.com/arran4/golang-rpg-textbox/theme/fromdirpng"
	"github.com/arran4/golang-rpg-textbox/util"
	"golang.org/x/image/draw"
	"image"
	"image/color/palette"
	"image/gif"
	"log"
	"strconv"
	"time"
)

// GenerateTextBox is a subcommand `rpgtextbox generate`
//
// Flags:
//   width:       --width       (default: 600)         Doc width
//   height:      --height      (default: 150)         Doc height
//   themeDir:    --themedir    (default: "./theme")   Directory to find the theme
//   fontName:    --font        (default: "goregular") Text font
//   dpi:         --dpi         (default: "75")        Doc dpi
//   fontSize:    --size        (default: "16")        font size
//   textSource:  --text        (default: "")          File in, or - for std input
//   outPrefix:   --out         (default: "out-")      Prefix of filename to output
//   chevronLoc:  --chevron     (default: "")          Use help for list
//   avatarPos:   --avatar-pos  (default: "")          Use help for list
//   avatarScale: --avatar-scale (default: "")         Use help for list
//   animation:   --animation   (default: "")          Use help for list
func GenerateTextBox(width, height int, themeDir, fontName string, dpi, fontSize string, textSource, outPrefix, chevronLoc, avatarPos, avatarScale, animation string) error {
	log.Printf("Starting")
	textBoxSize := image.Pt(width, height)
	var text string
	var err error
	text, err = util.GetText(textSource)
	if err != nil {
		return fmt.Errorf("Text fetch error: %w", err)
	}
	gr, err := util.OpenFont(fontName)
	if err != nil {
		return fmt.Errorf("Error opening font %s: %w", fontName, err)
	}
	fFontSize, err := strconv.ParseFloat(fontSize, 64)
	if err != nil {
		return fmt.Errorf("invalid float value for font size: %s", fontSize)
	}
	fDpi, err := strconv.ParseFloat(dpi, 64)
	if err != nil {
		return fmt.Errorf("invalid float value for dpi: %s", dpi)
	}
	grf := util.GetFontFace(fFontSize, fDpi, gr)
	t, err := cache.New(fromdirpng.New(themeDir, grf))
	if err != nil {
		return fmt.Errorf("Theme fetch error: %w", err)
	}
	var ops []rpgtextbox.Option
	chevronLocs := map[string][]rpgtextbox.Option{
		"center-bottom-chevron":               []rpgtextbox.Option{rpgtextbox.CenterBottomInsideTextFrame},
		"center-bottom-inside-chevron":        []rpgtextbox.Option{rpgtextbox.CenterBottomInsideFrame},
		"center-bottom-on-frame-text-chevron": []rpgtextbox.Option{rpgtextbox.CenterBottomOnFrameTextFrame},
		"center-bottom-on-frame-chevron":      []rpgtextbox.Option{rpgtextbox.CenterBottomOnFrameFrame},
		"right-bottom-inside-text-chevron":    []rpgtextbox.Option{rpgtextbox.RightBottomInsideTextFrame},
		"right-bottom-inside-chevron":         []rpgtextbox.Option{rpgtextbox.RightBottomInsideFrame},
		"right-bottom-on-frame-text-chevron":  []rpgtextbox.Option{rpgtextbox.RightBottomOnFrameTextFrame},
		"right-bottom-on-frame-chevron":       []rpgtextbox.Option{rpgtextbox.RightBottomOnFrameFrame},
		"end-of-text-chevron":                 []rpgtextbox.Option{rpgtextbox.TextEndChevron},
	}
	ext := "png"
	animated := false
	if len(chevronLoc) > 0 {
		help := chevronLoc == "help"
		if os, ok := chevronLocs[chevronLoc]; ok {
			ops = append(ops, os...)
		} else {
			help = true
		}
		if help {
			for k := range chevronLocs {
				log.Printf("%s", k)
			}
			return nil
		}
	}
	avatarPoss := map[string][]rpgtextbox.Option{
		"left-avatar":  []rpgtextbox.Option{rpgtextbox.LeftAvatar},
		"right-avatar": []rpgtextbox.Option{rpgtextbox.RightAvatar},
	}
	if len(avatarPos) > 0 {
		help := avatarPos == "help"
		if os, ok := avatarPoss[avatarPos]; ok {
			ops = append(ops, os...)
		} else {
			help = true
		}
		if help {
			for k := range avatarPoss {
				log.Printf("%s", k)
			}
			return nil
		}
	}
	avatarScales := map[string][]rpgtextbox.Option{
		"center-avatar":     []rpgtextbox.Option{rpgtextbox.CenterAvatar},
		"nearest-neighbour": []rpgtextbox.Option{rpgtextbox.NearestNeighbour},
		"approx-biLinear":   []rpgtextbox.Option{rpgtextbox.ApproxBiLinear},
	}
	if len(avatarScale) > 0 {
		help := avatarScale == "help"
		if os, ok := avatarScales[avatarScale]; ok {
			ops = append(ops, os...)
		} else {
			help = true
		}
		if help {
			for k := range avatarScales {
				log.Printf("%s", k)
			}
			return nil
		}
	}
	animations := map[string][]rpgtextbox.Option{
		"fade-animation":             []rpgtextbox.Option{rpgtextbox.NewFadeAnimation()},
		"box-by-box-animation":       []rpgtextbox.Option{rpgtextbox.NewBoxByBoxAnimation()},
		"letter-by-letter-animation": []rpgtextbox.Option{rpgtextbox.NewLetterByLetterAnimation()},
	}
	if len(animation) > 0 {
		help := animation == "help"
		if os, ok := animations[animation]; ok {
			ops = append(ops, os...)
			ext = "gif"
			animated = true
		} else {
			help = true
		}
		if help {
			for k := range animations {
				log.Printf("%s", k)
			}
			return nil
		}
	}

	tb, err := rpgtextbox.NewSimpleTextBox(t, text, textBoxSize, ops...)
	if err != nil {
		return fmt.Errorf("Error %w", err)
	}

	pages, err := tb.CalculateAllPages(textBoxSize)
	if err != nil {
		return fmt.Errorf("Text fetch error: %w", err)
	}

	if animated {
		gifo := &gif.GIF{}
		f := 0
		page := 0
		ofn := fmt.Sprintf("%s-animated.%s", outPrefix, ext)
		for {
			i := image.NewRGBA(image.Rect(0, 0, width, height))
			if done, ui, w, err := tb.DrawNextFrame(i); err != nil {
				return fmt.Errorf("Draw next frame error: %w", err)
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
				log.Printf("%s: Adding frame %d for page %d", ofn, f, page)
				bounds := i.Bounds()
				palettedImage := image.NewPaletted(bounds, palette.Plan9)
				draw.Draw(palettedImage, palettedImage.Rect, i, bounds.Min, draw.Over)
				gifo.Image = append(gifo.Image, palettedImage)
				gifo.Delay = append(gifo.Delay, int(w/(time.Second/100)))
			}
		}
		log.Printf("Saving %s", ofn)
		if err := util.SaveGifFile(ofn, gifo); err != nil {
			return fmt.Errorf("Error with saving file: %w", err)
		}
		log.Printf("Saved %s", ofn)

	} else {
		for page := 0; page < pages; page++ {
			i := image.NewRGBA(image.Rect(0, 0, width, height))
			if _, err := tb.DrawNextPageFrame(i); err != nil {
				return fmt.Errorf("Draw next frame error: %w", err)
			}
			ofn := fmt.Sprintf("%s-%02d.%s", outPrefix, page+1, ext)
			if err := util.SavePngFile(i, ofn); err != nil {
				return fmt.Errorf("Error with saving file: %w", err)
			}
			log.Printf("Saving %s", ofn)
		}
	}
	log.Printf("Done")
	return nil
}
