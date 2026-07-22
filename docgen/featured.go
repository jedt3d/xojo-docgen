package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

const (
	xojoGreen       = "#87B946" // primary
	xojoGreenDark   = "#5E8A2E" // accent
	xojoGreenLight  = "#A8D073" // tint
	xojoGreenLight2 = "#D4E8BF" // pale tint
)

// generateFeaturedPNG writes a 1200x630 landscape green banner PNG to assetsPath.
// Uses only stdlib (image, image/png). The banner is a vertical gradient from
// the base green to the darker accent, with a lighter band — a clean,
// intentional placeholder.
func generateFeaturedPNG(assetsPath string) error {
	return drawFeaturedBanner(assetsPath, 1200, 630)
}

// generateFeaturedPortraitPNG writes an 800x1000 portrait green banner PNG to
// assetsPath — a smaller, taller variant for phones and portrait tablets. The
// landing page uses <picture> art direction to swap it in below ~768px width.
// Same gradient recipe as the landscape banner, just a 4:5 aspect ratio.
func generateFeaturedPortraitPNG(assetsPath string) error {
	return drawFeaturedBanner(assetsPath, 800, 1000)
}

// drawFeaturedBanner is the shared gradient renderer for both orientations.
// The banner is a vertical gradient from the base green to the darker accent,
// with a lighter band — a clean, intentional placeholder.
func drawFeaturedBanner(assetsPath string, w, h int) error {
	if err := os.MkdirAll(filepath.Dir(assetsPath), 0o755); err != nil {
		return err
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	base := parseHex(xojoGreen)
	dark := parseHex(xojoGreenDark)
	light := parseHex(xojoGreenLight)
	for y := 0; y < h; y++ {
		t := float64(y) / float64(h-1)
		// Top 55% is a gradient from base to dark; a lighter band around 55-62%; below is dark.
		bandStart := float64(h) * 0.55
		bandEnd := float64(h) * 0.62
		var c color.RGBA
		switch {
		case y < int(bandStart):
			c = lerpColor(base, dark, t/0.55)
		case y < int(bandEnd):
			c = lerpColor(dark, light, (float64(y)-bandStart)/(bandEnd-bandStart))
		default:
			c = lerpColor(light, dark, (float64(y)-bandEnd)/(float64(h)-bandEnd))
		}
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	f, err := os.Create(assetsPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

// parseHex parses "#RRGGBB" into a color.RGBA (alpha 255).
func parseHex(s string) color.RGBA {
	s = trimPrefix(s, "#")
	if len(s) != 6 {
		return color.RGBA{0x87, 0xB9, 0x46, 0xFF}
	}
	r := hexByte(s[0:2])
	g := hexByte(s[2:4])
	b := hexByte(s[4:6])
	return color.RGBA{r, g, b, 0xFF}
}

func trimPrefix(s, p string) string {
	for i := 0; i < len(p) && i < len(s); i++ {
		if s[i] != p[i] {
			return s
		}
	}
	if len(s) >= len(p) {
		return s[len(p):]
	}
	return s
}

func hexByte(s string) uint8 {
	var v uint8
	for i := 0; i < len(s); i++ {
		v <<= 4
		switch {
		case s[i] >= '0' && s[i] <= '9':
			v |= s[i] - '0'
		case s[i] >= 'a' && s[i] <= 'f':
			v |= s[i] - 'a' + 10
		case s[i] >= 'A' && s[i] <= 'F':
			v |= s[i] - 'A' + 10
		}
	}
	return v
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: 0xFF,
	}
}
