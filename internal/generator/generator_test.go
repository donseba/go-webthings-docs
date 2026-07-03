package generator

import (
	"bytes"
	"image/png"
	"testing"
)

func TestPNGGeneratesLogo(t *testing.T) {
	body, err := PNG("Partial")
	if err != nil {
		t.Fatal(err)
	}
	if len(body) == 0 {
		t.Fatal("expected PNG bytes")
	}

	img, err := png.Decode(bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dy() != 70 {
		t.Fatalf("expected logo height 70, got %d", img.Bounds().Dy())
	}
	if img.Bounds().Dx() <= 81+11 {
		t.Fatalf("expected logo width to include dynamic middle, got %d", img.Bounds().Dx())
	}
}

func TestShortLogoStaysTight(t *testing.T) {
	img, err := Image("Web")
	if err != nil {
		t.Fatal(err)
	}
	if got := img.Bounds().Dx(); got > 170 {
		t.Fatalf("expected short generated logo to stay tight, got width %d", got)
	}
}

func TestSuffixTextStaysInsideBar(t *testing.T) {
	img, err := Image("Partial")
	if err != nil {
		t.Fatal(err)
	}

	barTop := 41
	barBottom := 68
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if a < 0x8000 {
				continue
			}
			greenText := g > 0x9000 && r < 0x7000 && b < 0x7000
			if greenText && (y < barTop || y > barBottom) {
				t.Fatalf("green suffix text escaped bar at %d,%d", x, y)
			}
		}
	}
}
