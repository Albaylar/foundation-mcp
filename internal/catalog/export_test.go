package catalog

import (
	"strings"
	"testing"
)

// projectFixture is a project token file that exercises what the base tokens never
// produce: a dark value, a color without one, another text style, a size with no text style.
const projectFixture = "testdata/project-tokens.json"

func loadFixture(t *testing.T) *Catalog {
	t.Helper()
	c, err := LoadWith("../../catalog", projectFixture)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestDarkValueIsReadFromColorTokens(t *testing.T) {
	c := loadFixture(t)
	if got := c.Tokens["color.surface"].Dark; got != "#0E0C0A" {
		t.Errorf("color.surface dark = %q", got)
	}
	if got := c.Tokens["color.surface"].Value; got != "#FFFFFF" {
		t.Errorf("the base value must stay the light value, got %v", got)
	}
	if got := c.Tokens["color.brand"].Dark; got != "" {
		t.Errorf("a color without a dark value must have none, got %q", got)
	}
	if got := c.Tokens["space.1"].Dark; got != "" {
		t.Errorf("only color tokens have a dark value, got %q on space.1", got)
	}
}

func TestSwiftExportAdaptsColorsWithADarkValue(t *testing.T) {
	out := loadFixture(t).Export("swiftui", "")
	for _, want := range []string{
		`static let surface = SwiftUI.Color(light: "#FFFFFF", dark: "#0E0C0A")`,
		`static let brand = SwiftUI.Color(hex: "#7A4B16")`,
		"init(light: String, dark: String)",
		"userInterfaceStyle == .dark",
		"import UIKit\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("swift export has no %q:\n%s", want, out)
		}
	}
}

func TestSwiftExportWithoutDarkValuesStaysAsItWas(t *testing.T) {
	c, err := Load("../../catalog")
	if err != nil {
		t.Fatal(err)
	}
	out := c.Export("swiftui", "")
	if !strings.Contains(out, `static let brand = SwiftUI.Color(hex: "#0B6E99")`) {
		t.Errorf("fixed color missing:\n%s", out)
	}
	for _, unwanted := range []string{"light:", "UIKit", "UIColor"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("no token has a dark value, yet the export contains %q:\n%s", unwanted, out)
		}
	}
}

func TestSwiftExportMapsBaseTypeToScalingTextStyles(t *testing.T) {
	c, err := Load("../../catalog")
	if err != nil {
		t.Fatal(err)
	}
	out := c.Export("swiftui", "")
	for _, want := range []string{
		"static let title = Font.title.weight(.bold)\n",
		"static let body = Font.body\n",
		"static let label = Font.subheadline.weight(.semibold)\n",
		"static let caption = Font.footnote\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("swift export has no %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "Font.system(size:") {
		t.Errorf("base sizes all match a text style, so none may be fixed:\n%s", out)
	}
	if w := c.ExportWarnings("swiftui", ""); len(w) != 0 {
		t.Errorf("base tokens must not warn: %v", w)
	}
}

func TestSwiftExportFollowsAProjectSizeToItsTextStyle(t *testing.T) {
	out := loadFixture(t).Export("swiftui", "")
	if !strings.Contains(out, "static let title = Font.largeTitle.weight(.bold)\n") {
		t.Errorf("34 pt is the default of Large Title, not Title 1:\n%s", out)
	}
}

// A project's size is its design: it is never swapped for the nearest style, and the
// loss of scaling is reported in the generated file and to whoever runs the export.
func TestSwiftExportKeepsASizeWithNoTextStyleAndSaysSo(t *testing.T) {
	c := loadFixture(t)
	out := c.Export("swiftui", "")
	if !strings.Contains(out, "static let body = Font.system(size: 18, weight: .regular)\n") {
		t.Errorf("the project's size must be kept:\n%s", out)
	}
	if !strings.Contains(out, `#warning("type.body`) {
		t.Errorf("the generated file must warn at build time:\n%s", out)
	}
	w := c.ExportWarnings("swiftui", "")
	if len(w) != 1 || !strings.Contains(w[0], "type.body") || !strings.Contains(w[0], "Dynamic Type") {
		t.Errorf("want one warning naming type.body and Dynamic Type, got %v", w)
	}
	if w := c.ExportWarnings("swiftui", "color"); len(w) != 0 {
		t.Errorf("a color export has no fonts to warn about, got %v", w)
	}
	if w := c.ExportWarnings("flutter", ""); len(w) != 0 {
		t.Errorf("Flutter text scales at any size, got %v", w)
	}
}

func TestFlutterExportStaysLightOnlyAndSaysSo(t *testing.T) {
	out := loadFixture(t).Export("flutter", "")
	if !strings.Contains(out, "static const Color colorSurface = Color(0xFFFFFFFF)") {
		t.Errorf("flutter must keep the light value:\n%s", out)
	}
	if strings.Contains(out, "0E0C0A") {
		t.Errorf("flutter must not export the dark value:\n%s", out)
	}
	if !strings.Contains(out, "// Light values only") {
		t.Errorf("a dropped dark value must be stated in the generated file:\n%s", out)
	}
	c, err := Load("../../catalog")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(c.Export("flutter", ""), "// Light values only") {
		t.Error("nothing is dropped when no token has a dark value")
	}
}

func TestOnDangerKeepsTodaysLookInTheBaseTokens(t *testing.T) {
	c, err := Load("../../catalog")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := c.Tokens["color.on-danger"].Value, c.Tokens["color.on-brand"].Value; got != want || got == nil {
		t.Errorf("base color.on-danger = %v, want the base color.on-brand %v", got, want)
	}
}

func TestDestructiveButtonUsesOnDanger(t *testing.T) {
	c, err := Load("../../catalog")
	if err != nil {
		t.Fatal(err)
	}
	comp, ok := c.Get("button")
	if !ok {
		t.Fatal("no Button")
	}
	if !contains(comp.Tokens, "color.on-danger") {
		t.Errorf("Button manifest does not list color.on-danger: %v", comp.Tokens)
	}
	for platform, want := range map[string]string{
		"swiftui": "case .destructive: FoundationTokens.Color.onDanger",
		"flutter": "FButtonVariant.destructive => FoundationTokens.colorOnDanger",
	} {
		src, err := c.Source(comp, platform)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(src, want) {
			t.Errorf("%s Button has no %q", platform, want)
		}
	}
}
