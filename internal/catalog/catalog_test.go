package catalog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectTokensOverrideAndAdd(t *testing.T) {
	dir := t.TempDir()
	project := filepath.Join(dir, "tokens.json")
	if err := os.WriteFile(project, []byte(`{"color":{"brand":{"$type":"color","$value":"#7A4B16"},"accent":{"$type":"color","$value":"#123456"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	base, err := Load("../../catalog")
	if err != nil {
		t.Fatal(err)
	}
	c, err := LoadWith("../../catalog", project)
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Tokens["color.brand"].Value; got != "#7A4B16" {
		t.Errorf("project value must win, got %v", got)
	}
	if _, ok := c.Tokens["color.accent"]; !ok {
		t.Error("new project token must be added")
	}
	if c.Tokens["color.surface"].Value != base.Tokens["color.surface"].Value {
		t.Error("tokens the project does not set must keep the base value")
	}
	if len(c.Components) != len(base.Components) {
		t.Error("components must still validate against the merged tokens")
	}
}

func TestProjectTokensMissingFile(t *testing.T) {
	if _, err := LoadWith("../../catalog", "/nonexistent/tokens.json"); err == nil {
		t.Error("a missing project token file must be an error, not silently ignored")
	}
}

func TestCheckUsageIgnoresTokenReferences(t *testing.T) {
	c, err := Load("../../catalog")
	if err != nil {
		t.Fatal(err)
	}
	clean := "FButton(\"Save\") {}\nText(\"x\").padding(FoundationTokens.Space.s3)\n// #FF0000 in a comment is fine"
	if f := c.CheckUsage(clean); len(f) != 0 {
		t.Errorf("clean source reported findings: %+v", f)
	}
}

func TestCheckUsageAcceptsTypesOfCatalogComponents(t *testing.T) {
	c, err := Load("../../catalog")
	if err != nil {
		t.Fatal(err)
	}
	src := "FButton(title: 'x', variant: FButtonVariant.primary)"
	if f := c.CheckUsage(src); len(f) != 0 {
		t.Errorf("a type belonging to a catalog component must not be flagged: %+v", f)
	}
	if f := c.CheckUsage("FMagicThing()"); len(f) != 1 {
		t.Errorf("a component outside the catalog must still be flagged, got %+v", f)
	}
}

// .Type is reserved in Swift: a nested `enum Type` makes FoundationTokens.Type the metatype.
func TestSwiftExportAvoidsReservedNames(t *testing.T) {
	c, err := Load("../../catalog")
	if err != nil {
		t.Fatal(err)
	}
	out := c.Export("swiftui", "")
	if strings.Contains(out, "enum Type") {
		t.Error("generated Swift declares a nested enum named Type")
	}
	if !strings.Contains(out, "enum Typography") {
		t.Error("generated Swift has no Typography enum")
	}
}

func TestCatalogSourcesMatchTheGeneratedTokenNames(t *testing.T) {
	c, err := Load("../../catalog")
	if err != nil {
		t.Fatal(err)
	}
	for _, comp := range c.Components {
		src, err := c.Source(comp, "swiftui")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(src, "FoundationTokens.Type.") {
			t.Errorf("%s uses FoundationTokens.Type", comp.Name)
		}
	}
}
