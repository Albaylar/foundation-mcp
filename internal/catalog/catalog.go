// Package catalog loads the design-system catalog (tokens.json plus one
// manifest.json per component) and answers the questions the MCP tools ask.
// It knows nothing about MCP; everything here is plain Go and testable.
package catalog

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Platforms the catalog may carry an implementation for.
var Platforms = []string{"swiftui", "flutter", "compose"}

// Token is one entry of the W3C Design Tokens file, flattened to "group.name".
type Token struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	Value       any    `json:"value"`
	Description string `json:"description,omitempty"`
}

// Variant is one named configuration of a component.
type Variant struct {
	Name  string            `json:"name"`
	Props map[string]string `json:"props"`
}

// Manifest is the on-disk description of a component (manifest.json).
type Manifest struct {
	Name          string            `json:"name"`
	Purpose       string            `json:"purpose"`
	Tags          []string          `json:"tags,omitempty"`
	Status        string            `json:"status"`
	Version       string            `json:"version"`
	Variants      []Variant         `json:"variants,omitempty"`
	Tokens        []string          `json:"tokens,omitempty"`
	Accessibility string            `json:"accessibility"`
	FigmaNode     string            `json:"figma_node"`
	Platforms     map[string]string `json:"platforms,omitempty"` // platform -> source path relative to the component dir
}

// Component is a loaded manifest plus where it lives.
type Component struct {
	Manifest
	Dir string `json:"-"`
}

// Catalog is the loaded design system.
type Catalog struct {
	Root       string
	Tokens     map[string]Token
	Components map[string]*Component // keyed by lower-case name
}

// Load reads tokens.json and every components/*/manifest.json under root.
func Load(root string) (*Catalog, error) {
	return LoadWith(root, "")
}

// LoadWith is Load plus an optional project token file. The project file is
// merged over the base tokens: same path = the project's value wins, new paths
// are added. That is how one shared component catalog serves apps with
// different brands: components reference token names, projects own the values.
func LoadWith(root, projectTokens string) (*Catalog, error) {
	c := &Catalog{Root: root, Tokens: map[string]Token{}, Components: map[string]*Component{}}
	if err := readTokens(filepath.Join(root, "tokens.json"), c.Tokens); err != nil {
		return nil, err
	}
	if projectTokens != "" {
		if err := readTokens(projectTokens, c.Tokens); err != nil {
			return nil, err
		}
	}

	dirs, err := filepath.Glob(filepath.Join(root, "components", "*", "manifest.json"))
	if err != nil {
		return nil, err
	}
	for _, m := range dirs {
		comp, err := loadManifest(m)
		if err != nil {
			return nil, err
		}
		if err := c.validate(comp); err != nil {
			return nil, fmt.Errorf("%s: %w", comp.Name, err)
		}
		c.Components[strings.ToLower(comp.Name)] = comp
	}
	return c, nil
}

func readTokens(path string, into map[string]Token) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("tokens: %w", err)
	}
	var tree map[string]any
	if err := json.Unmarshal(raw, &tree); err != nil {
		return fmt.Errorf("tokens %s: %w", path, err)
	}
	flattenTokens("", tree, into)
	return nil
}

func loadManifest(path string) (*Component, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &Component{Manifest: m, Dir: filepath.Dir(path)}, nil
}

// flattenTokens turns {"color":{"brand":{"$value":…}}} into "color.brand".
func flattenTokens(prefix string, node map[string]any, out map[string]Token) {
	if v, ok := node["$value"]; ok {
		t := Token{Path: prefix, Value: v}
		if s, ok := node["$type"].(string); ok {
			t.Type = s
		}
		if s, ok := node["$description"].(string); ok {
			t.Description = s
		}
		out[prefix] = t
		return
	}
	for k, v := range node {
		if strings.HasPrefix(k, "$") {
			continue
		}
		child, ok := v.(map[string]any)
		if !ok {
			continue
		}
		p := k
		if prefix != "" {
			p = prefix + "." + k
		}
		flattenTokens(p, child, out)
	}
}

// validate is the rule set register_component also enforces: every token
// exists, at least one platform source exists, status is known.
func (c *Catalog) validate(comp *Component) error {
	if comp.Name == "" {
		return errors.New("name is empty")
	}
	switch comp.Status {
	case "stable", "draft", "deprecated":
	default:
		return fmt.Errorf("status %q must be stable, draft or deprecated", comp.Status)
	}
	for _, t := range comp.Tokens {
		if _, ok := c.Tokens[t]; !ok {
			return fmt.Errorf("token %q is not in tokens.json", t)
		}
	}
	if len(comp.Platforms) == 0 {
		return errors.New("no platform implementation")
	}
	for p, src := range comp.Platforms {
		if !contains(Platforms, p) {
			return fmt.Errorf("unknown platform %q", p)
		}
		if _, err := os.Stat(filepath.Join(comp.Dir, src)); err != nil {
			return fmt.Errorf("platform %s: source %s missing", p, src)
		}
	}
	return nil
}

// Summary is the row list_components returns.
type Summary struct {
	Name      string   `json:"name"`
	Purpose   string   `json:"purpose"`
	Status    string   `json:"status"`
	Version   string   `json:"version"`
	Variants  []string `json:"variants"`
	Platforms []string `json:"platforms"`
	Tags      []string `json:"tags"`
}

// List returns every component, optionally filtered by platform and tag, sorted by name.
func (c *Catalog) List(platform, tag string) []Summary {
	var out []Summary
	for _, comp := range c.Components {
		if platform != "" {
			if _, ok := comp.Platforms[platform]; !ok {
				continue
			}
		}
		if tag != "" && !contains(comp.Tags, tag) {
			continue
		}
		out = append(out, summarize(comp))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func summarize(comp *Component) Summary {
	s := Summary{Name: comp.Name, Purpose: comp.Purpose, Status: comp.Status, Version: comp.Version, Tags: comp.Tags}
	for _, v := range comp.Variants {
		s.Variants = append(s.Variants, v.Name)
	}
	for p := range comp.Platforms {
		s.Platforms = append(s.Platforms, p)
	}
	sort.Strings(s.Platforms)
	return s
}

// Get returns a component by name (case-insensitive).
func (c *Catalog) Get(name string) (*Component, bool) {
	comp, ok := c.Components[strings.ToLower(name)]
	return comp, ok
}

// Source reads the implementation of a component for one platform.
func (c *Catalog) Source(comp *Component, platform string) (string, error) {
	rel, ok := comp.Platforms[platform]
	if !ok {
		return "", fmt.Errorf("%s has no %s implementation (available: %s)", comp.Name, platform, strings.Join(keys(comp.Platforms), ", "))
	}
	b, err := os.ReadFile(filepath.Join(comp.Dir, rel))
	return string(b), err
}

// Search matches query words against name, purpose, tags and variant names.
// Every word must match somewhere; result is sorted by number of hits.
func (c *Catalog) Search(query string) ([]Summary, []Token) {
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return nil, nil
	}
	type scored struct {
		s     Summary
		score int
	}
	var comps []scored
	for _, comp := range c.Components {
		hay := strings.ToLower(comp.Name + " " + comp.Purpose + " " + strings.Join(comp.Tags, " "))
		for _, v := range comp.Variants {
			hay += " " + strings.ToLower(v.Name)
		}
		score := 0
		for _, w := range words {
			if strings.Contains(hay, w) {
				score++
			}
		}
		if score > 0 {
			comps = append(comps, scored{summarize(comp), score})
		}
	}
	sort.Slice(comps, func(i, j int) bool {
		if comps[i].score != comps[j].score {
			return comps[i].score > comps[j].score
		}
		return comps[i].s.Name < comps[j].s.Name
	})
	var out []Summary
	for _, s := range comps {
		out = append(out, s.s)
	}
	var toks []Token
	for _, t := range c.Tokens {
		hay := strings.ToLower(t.Path + " " + t.Description)
		for _, w := range words {
			if strings.Contains(hay, w) {
				toks = append(toks, t)
				break
			}
		}
	}
	sort.Slice(toks, func(i, j int) bool { return toks[i].Path < toks[j].Path })
	return out, toks
}

// TokensIn returns the tokens of one set ("color", "type", …) or all.
func (c *Catalog) TokensIn(set string) []Token {
	var out []Token
	for _, t := range c.Tokens {
		if set == "" || strings.HasPrefix(t.Path, set+".") {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// Finding is one hard-coded value or unknown component check_usage reports.
type Finding struct {
	Line    int    `json:"line"`
	Kind    string `json:"kind"` // hex-color, literal-spacing, unknown-component
	Snippet string `json:"snippet"`
	Hint    string `json:"hint"`
}

var (
	reHex     = regexp.MustCompile(`(?i)#[0-9a-f]{6}\b|Color\(\s*(?:red|0x)|\b0xFF[0-9A-F]{6}\b`)
	reSpacing = regexp.MustCompile(`\.padding\(\s*\d+|EdgeInsets\.(?:all|symmetric|only)\([^)]*\d|padding\(\s*(?:horizontal|vertical)?\s*:?\s*\d+|spacing:\s*\d+`)
	reComp    = regexp.MustCompile(`\bF([A-Z][A-Za-z]+)\b`)
)

// CheckUsage scans screen source for values that should be tokens or components
// that are not in the catalog. Regex-based on purpose (v1): cheap and platform-agnostic.
func (c *Catalog) CheckUsage(source string) []Finding {
	var out []Finding
	for i, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "///") {
			continue
		}
		if m := reHex.FindString(line); m != "" {
			out = append(out, Finding{i + 1, "hex-color", trimmed, "use a color token (get_tokens set=color)"})
		}
		if m := reSpacing.FindString(line); m != "" && !strings.Contains(line, "Tokens.") {
			out = append(out, Finding{i + 1, "literal-spacing", trimmed, "use a space token (get_tokens set=space)"})
		}
		for _, m := range reComp.FindAllStringSubmatch(line, -1) {
			if _, ok := c.Get(m[1]); !ok {
				out = append(out, Finding{i + 1, "unknown-component", trimmed, "F" + m[1] + " is not in the catalog; register it through /design"})
			}
		}
	}
	return out
}

// Register validates and writes a new component into the catalog.
// sources maps platform -> file content; the file name is derived from the component name.
func (c *Catalog) Register(m Manifest, sources map[string]string) error {
	if _, exists := c.Get(m.Name); exists {
		return fmt.Errorf("%s already exists; bump its version instead", m.Name)
	}
	dir := filepath.Join(c.Root, "components", strings.ToLower(strings.ReplaceAll(m.Name, " ", "_")))
	if m.Platforms == nil {
		m.Platforms = map[string]string{}
	}
	ext := map[string]string{"swiftui": ".swift", "flutter": ".dart", "compose": ".kt"}
	for p, src := range sources {
		if !contains(Platforms, p) {
			return fmt.Errorf("unknown platform %q", p)
		}
		rel := filepath.Join(p, m.Name+ext[p])
		m.Platforms[p] = rel
		if err := os.MkdirAll(filepath.Join(dir, p), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(src), 0o644); err != nil {
			return err
		}
	}
	comp := &Component{Manifest: m, Dir: dir}
	if err := c.validate(comp); err != nil {
		_ = os.RemoveAll(dir)
		return err
	}
	raw, _ := json.MarshalIndent(m, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), raw, 0o644); err != nil {
		return err
	}
	c.Components[strings.ToLower(m.Name)] = comp
	return nil
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func keys(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
