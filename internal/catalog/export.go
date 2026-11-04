package catalog

import (
	"fmt"
	"sort"
	"strings"
)

// Export renders the tokens of a set as source for one platform, so a project
// can paste a generated FoundationTokens file instead of typing values.
// Unsupported platform or set → empty string (the JSON list is still returned by the tool).
func (c *Catalog) Export(platform, set string) string {
	toks := c.TokensIn(set)
	switch platform {
	case "swiftui":
		return exportSwift(toks)
	case "flutter":
		return exportDart(toks)
	}
	return ""
}

// ExportWarnings lists what the export for a platform cannot honor, for whoever runs it.
func (c *Catalog) ExportWarnings(platform, set string) []string {
	if platform != "swiftui" {
		return nil
	}
	var out []string
	for _, t := range c.TokensIn(set) {
		if strings.HasPrefix(t.Path, "type.") {
			if w := fixedSizeWarning(t); w != "" {
				out = append(out, w)
			}
		}
	}
	return out
}

func exportSwift(toks []Token) string {
	adaptive := hasDarkValue(toks)
	var b strings.Builder
	b.WriteString("import SwiftUI\n")
	if adaptive {
		b.WriteString("import UIKit\n")
	}
	b.WriteString("\n// Generated from tokens.json by foundation-mcp. Do not edit.\npublic enum FoundationTokens {\n")
	groups := group(toks)
	for _, g := range sortedKeys(groups) {
		switch g {
		case "color":
			b.WriteString("    public enum Color {\n")
			for _, t := range groups[g] {
				if t.Dark != "" {
					fmt.Fprintf(&b, "        public static let %s = SwiftUI.Color(light: %q, dark: %q)\n", camel(t.Path), t.Value, t.Dark)
				} else {
					fmt.Fprintf(&b, "        public static let %s = SwiftUI.Color(hex: %q)\n", camel(t.Path), t.Value)
				}
			}
			b.WriteString("    }\n")
		case "space", "radius":
			name := map[string]string{"space": "Space", "radius": "Radius"}[g]
			fmt.Fprintf(&b, "    public enum %s {\n", name)
			for _, t := range groups[g] {
				n := camel(t.Path)
				if g == "space" {
					n = "s" + n
				}
				fmt.Fprintf(&b, "        public static let %s: CGFloat = %v\n", n, t.Value)
			}
			b.WriteString("    }\n")
		case "type":
			b.WriteString("    public enum Typography {\n")
			for _, t := range groups[g] {
				if w := fixedSizeWarning(t); w != "" {
					fmt.Fprintf(&b, "        #warning(%q)\n", w)
				}
				fmt.Fprintf(&b, "        public static let %s = %s\n", camel(t.Path), swiftFont(t))
			}
			b.WriteString("    }\n")
		case "motion":
			b.WriteString("    public enum Motion {\n")
			for _, t := range groups[g] {
				fmt.Fprintf(&b, "        public static let %s: Double = %s\n", camel(t.Path), seconds(fmt.Sprint(t.Value)))
			}
			b.WriteString("    }\n")
		}
	}
	b.WriteString("}\n\nextension SwiftUI.Color {\n    init(hex: String) {\n        var v: UInt64 = 0\n        Scanner(string: String(hex.dropFirst())).scanHexInt64(&v)\n        self.init(red: Double((v >> 16) & 0xFF) / 255, green: Double((v >> 8) & 0xFF) / 255, blue: Double(v & 0xFF) / 255)\n    }\n")
	if adaptive {
		b.WriteString("\n    init(light: String, dark: String) {\n        let light = UIColor(SwiftUI.Color(hex: light))\n        let dark = UIColor(SwiftUI.Color(hex: dark))\n        self.init(uiColor: UIColor { $0.userInterfaceStyle == .dark ? dark : light })\n    }\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func exportDart(toks []Token) string {
	var b strings.Builder
	b.WriteString("import 'package:flutter/material.dart';\n\n// Generated from tokens.json by foundation-mcp. Do not edit.\n")
	if hasDarkValue(toks) {
		b.WriteString("// Light values only: the dark values in the token file are not exported for Flutter.\n")
	}
	b.WriteString("abstract final class FoundationTokens {\n")
	groups := group(toks)
	for _, g := range sortedKeys(groups) {
		for _, t := range groups[g] {
			n := g + pascal(t.Path)
			switch g {
			case "color":
				fmt.Fprintf(&b, "  static const Color %s = Color(0xFF%s);\n", n, strings.TrimPrefix(fmt.Sprint(t.Value), "#"))
			case "space", "radius":
				fmt.Fprintf(&b, "  static const double %s = %v;\n", n, t.Value)
			case "type":
				m, _ := t.Value.(map[string]any)
				fmt.Fprintf(&b, "  static const TextStyle %s = TextStyle(fontSize: %v, fontWeight: FontWeight.w%v, height: %s);\n", n, m["fontSize"], m["fontWeight"], ratio(fmt.Sprint(m["lineHeight"]), fmt.Sprint(m["fontSize"])))
			case "motion":
				fmt.Fprintf(&b, "  static const Duration %s = Duration(milliseconds: %s);\n", n, strings.TrimSuffix(fmt.Sprint(t.Value), "ms"))
			}
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func hasDarkValue(toks []Token) bool {
	for _, t := range toks {
		if t.Dark != "" {
			return true
		}
	}
	return false
}

func group(toks []Token) map[string][]Token {
	out := map[string][]Token{}
	for _, t := range toks {
		g := strings.SplitN(t.Path, ".", 2)[0]
		out[g] = append(out[g], t)
	}
	return out
}

func sortedKeys(m map[string][]Token) []string {
	var ks []string
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// camel: "color.brand-strong" → "brandStrong"; "space.4" → "4".
func camel(path string) string {
	p := pascal(path)
	if p == "" {
		return p
	}
	return strings.ToLower(p[:1]) + p[1:]
}

// pascal: "color.brand-strong" → "BrandStrong".
func pascal(path string) string {
	name := path[strings.Index(path, ".")+1:]
	var b strings.Builder
	up := true
	for _, r := range name {
		if r == '-' || r == '_' || r == '.' {
			up = true
			continue
		}
		if up {
			b.WriteString(strings.ToUpper(string(r)))
			up = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// swiftTextStyles maps the default size of each iOS text style to its SwiftUI font.
var swiftTextStyles = map[string]string{
	"34": "largeTitle",
	"28": "title",
	"22": "title2",
	"20": "title3",
	"17": "body",
	"16": "callout",
	"15": "subheadline",
	"13": "footnote",
	"12": "caption",
	"11": "caption2",
}

// swiftFont scales with Dynamic Type when the size is the default of a text style.
// SwiftUI has no scaling system font at any other size, so those keep the project's
// size and stay fixed; fixedSizeWarning reports them.
func swiftFont(t Token) string {
	m, _ := t.Value.(map[string]any)
	weight := swiftWeight(fmt.Sprint(m["fontWeight"]))
	style, ok := swiftTextStyles[fmt.Sprint(m["fontSize"])]
	if !ok {
		return fmt.Sprintf("Font.system(size: %v, weight: %s)", m["fontSize"], weight)
	}
	if weight == ".regular" {
		return "Font." + style
	}
	return fmt.Sprintf("Font.%s.weight(%s)", style, weight)
}

func fixedSizeWarning(t Token) string {
	m, _ := t.Value.(map[string]any)
	if _, ok := swiftTextStyles[fmt.Sprint(m["fontSize"])]; ok {
		return ""
	}
	return fmt.Sprintf("%s: size %v is no iOS text style default (11, 12, 13, 15, 16, 17, 20, 22, 28, 34), so its SwiftUI font stays fixed and ignores Dynamic Type", t.Path, m["fontSize"])
}

func swiftWeight(w string) string {
	switch w {
	case "700":
		return ".bold"
	case "600":
		return ".semibold"
	case "500":
		return ".medium"
	default:
		return ".regular"
	}
}

func seconds(ms string) string {
	var n float64
	fmt.Sscanf(strings.TrimSuffix(ms, "ms"), "%f", &n)
	return fmt.Sprintf("%g", n/1000)
}

func ratio(lineHeight, fontSize string) string {
	var lh, fs float64
	fmt.Sscanf(lineHeight, "%f", &lh)
	fmt.Sscanf(fontSize, "%f", &fs)
	if fs == 0 {
		return "1.2"
	}
	return fmt.Sprintf("%.2f", lh/fs)
}
