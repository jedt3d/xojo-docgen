package main

import (
	"strings"
	"testing"
)

func TestParsePrimaryColor(t *testing.T) {
	color, err := parsePrimaryColor("11, 99, 56")
	if err != nil {
		t.Fatal(err)
	}
	if color != (RGBColor{R: 11, G: 99, B: 56}) {
		t.Fatalf("color = %#v", color)
	}
	if color.hex() != "#0B6338" {
		t.Fatalf("hex = %s", color.hex())
	}
}

func TestParsePrimaryColorRejectsInvalidValues(t *testing.T) {
	for _, value := range []string{"#0B6338", "11,99", "11,99,256", "11,green,56"} {
		if _, err := parsePrimaryColor(value); err == nil {
			t.Errorf("parsePrimaryColor(%q) succeeded", value)
		}
	}
}

func TestGeneratedPaletteUsesPrimaryAndAccessibleAccents(t *testing.T) {
	primary := RGBColor{R: 230, G: 230, B: 20}
	css := renderPrimaryColorCSS(primary)
	if !strings.Contains(css, "--xojo-primary: #E6E614;") {
		t.Fatalf("palette does not contain primary:\n%s", css)
	}

	lightAccent := accessibleVariant(primary, RGBColor{R: 255, G: 255, B: 255}, RGBColor{})
	if contrastRatio(lightAccent, RGBColor{R: 255, G: 255, B: 255}) < 4.5 {
		t.Fatalf("light accent contrast = %.2f", contrastRatio(lightAccent, RGBColor{R: 255, G: 255, B: 255}))
	}
	darkAccent := accessibleVariant(primary, RGBColor{R: 16, G: 21, B: 18}, RGBColor{R: 255, G: 255, B: 255})
	if contrastRatio(darkAccent, RGBColor{R: 16, G: 21, B: 18}) < 4.5 {
		t.Fatalf("dark accent contrast = %.2f", contrastRatio(darkAccent, RGBColor{R: 16, G: 21, B: 18}))
	}
}
