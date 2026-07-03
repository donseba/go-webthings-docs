package generator

import (
	"bytes"
	"embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"
	"unicode"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed img/*.png
//go:embed font/*.ttf
var assetFS embed.FS

var (
	goBlue  = color.NRGBA{R: 0x00, G: 0xc0, B: 0xfd, A: 0xff}
	textGre = color.NRGBA{R: 0x63, G: 0xd8, B: 0x1f, A: 0xff}
)

const activeFontProfile = "press-start-2p"

type fontProfile struct {
	File           string
	Size           float64
	TextInset      int
	RightPad       int
	BarTop         int
	BarBottom      int
	BaselineOffset int
}

var fontProfiles = map[string]fontProfile{
	"upheavtt": {
		File:           "font/upheavtt.ttf",
		Size:           26,
		TextInset:      6,
		RightPad:       -5,
		BarTop:         41,
		BarBottom:      68,
		BaselineOffset: 1,
	},
	"press-start-2p": {
		File:           "font/PressStart2P.ttf",
		Size:           15,
		TextInset:      7,
		RightPad:       -3,
		BarTop:         41,
		BarBottom:      68,
		BaselineOffset: 3,
	},
}

type assets struct {
	start  image.Image
	middle image.Image
	end    image.Image
}

func PNG(input string) ([]byte, error) {
	img, err := Image(input)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Image(input string) (image.Image, error) {
	parts, err := loadAssets()
	if err != nil {
		return nil, err
	}

	suffix := normalize(input)
	prefix := "Go-"
	profile := fontProfiles[activeFontProfile]
	face, err := logoFace(profile)
	if err != nil {
		return nil, err
	}
	defer face.Close()

	textWidth := textInkPixels(face, prefix+suffix)
	contentWidth := profile.TextInset + textWidth + profile.RightPad
	middleWidth := max(parts.middle.Bounds().Dx(), contentWidth-parts.start.Bounds().Dx())
	height := parts.start.Bounds().Dy()
	width := parts.start.Bounds().Dx() + middleWidth + parts.end.Bounds().Dx()

	out := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(out, parts.start.Bounds(), parts.start, parts.start.Bounds().Min, draw.Over)

	middleStart := parts.start.Bounds().Dx()
	for x := 0; x < middleWidth; x += parts.middle.Bounds().Dx() {
		dst := image.Rect(middleStart+x, 0, min(middleStart+x+parts.middle.Bounds().Dx(), middleStart+middleWidth), height)
		draw.Draw(out, dst, parts.middle, image.Point{}, draw.Over)
	}

	endX := middleStart + middleWidth
	draw.Draw(out, image.Rect(endX, 0, endX+parts.end.Bounds().Dx(), height), parts.end, image.Point{}, draw.Over)
	drawLabel(out, face, profile, prefix, suffix)

	return out, nil
}

func drawLabel(dst *image.NRGBA, face font.Face, profile fontProfile, prefix, suffix string) {
	metrics := face.Metrics()
	textHeight := (metrics.Ascent + metrics.Descent).Ceil()
	baseline := profile.BarTop + (profile.BarBottom-profile.BarTop-textHeight)/2 + metrics.Ascent.Ceil() + profile.BaselineOffset

	x := profile.TextInset
	x += drawSolidString(dst, face, prefix, x, baseline, goBlue)
	drawSolidString(dst, face, suffix, x, baseline, textGre)
}

func drawSolidString(dst *image.NRGBA, face font.Face, text string, x, baseline int, c color.NRGBA) int {
	mask := image.NewAlpha(dst.Bounds())
	d := font.Drawer{
		Dst:  mask,
		Src:  image.White,
		Face: face,
		Dot:  fixed.P(x, baseline),
	}
	d.DrawString(text)

	// The source logo is crisp pixel art. Threshold the rasterized TTF mask so
	// the generated wordmark has matching hard edges instead of blurry AA.
	for y := mask.Bounds().Min.Y; y < mask.Bounds().Max.Y; y++ {
		for x := mask.Bounds().Min.X; x < mask.Bounds().Max.X; x++ {
			if mask.AlphaAt(x, y).A > 96 {
				dst.SetNRGBA(x, y, c)
			}
		}
	}

	return textPixels(face, text)
}

func textPixels(face font.Face, text string) int {
	return font.MeasureString(face, text).Ceil()
}

func textInkPixels(face font.Face, text string) int {
	bounds, _ := font.BoundString(face, text)
	return (bounds.Max.X - bounds.Min.X).Ceil()
}

func normalize(input string) string {
	input = strings.TrimSpace(input)
	input = strings.TrimPrefix(input, "Go-")
	input = strings.TrimPrefix(input, "go-")
	var out []rune
	for _, r := range input {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			out = append(out, r)
		case r == ' ', r == '-', r == '_':
			out = append(out, r)
		}
		if len(out) >= 28 {
			break
		}
	}
	clean := strings.TrimSpace(string(out))
	if clean == "" {
		return "WebThings"
	}
	return clean
}

func loadAssets() (assets, error) {
	start, err := loadPNG("img/logo_start.png")
	if err != nil {
		return assets{}, err
	}
	middle, err := loadPNG("img/logo_middle.png")
	if err != nil {
		return assets{}, err
	}
	end, err := loadPNG("img/logo_end.png")
	if err != nil {
		return assets{}, err
	}
	return assets{start: start, middle: middle, end: end}, nil
}

func loadPNG(path string) (image.Image, error) {
	file, err := assetFS.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return png.Decode(file)
}

func logoFace(profile fontProfile) (font.Face, error) {
	body, err := assetFS.ReadFile(profile.File)
	if err != nil {
		return nil, err
	}
	ttf, err := opentype.Parse(body)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(ttf, &opentype.FaceOptions{
		Size:    profile.Size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}
