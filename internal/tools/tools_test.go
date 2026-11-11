package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Albaylar/foundation-mcp/internal/catalog"
)

// connect starts the server over in-memory transports and returns a client session.
func connect(t *testing.T) *mcp.ClientSession {
	t.Helper()
	return connectWith(t, "")
}

// connectWith is connect with a project token file merged over the catalog's tokens.
func connectWith(t *testing.T, projectTokens string) *mcp.ClientSession {
	t.Helper()
	c, err := catalog.LoadWith("../../catalog", projectTokens)
	if err != nil {
		t.Fatal(err)
	}
	server := mcp.NewServer(&mcp.Implementation{Name: "foundation", Version: "test"}, nil)
	Register(server, c)
	st, ct := mcp.NewInMemoryTransports()
	if _, err := server.Connect(context.Background(), st, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(context.Background(), ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

func call(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) map[string]any {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if res.IsError {
		t.Fatalf("%s returned error: %v", name, res.Content[0].(*mcp.TextContent).Text)
	}
	b, _ := json.Marshal(res.StructuredContent)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out
}

func TestToolsAreRegistered(t *testing.T) {
	cs := connect(t)
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"list_components": false, "get_component": false, "get_tokens": false, "search": false, "check_usage": false, "register_component": false}
	for _, tool := range res.Tools {
		want[tool.Name] = true
	}
	for name, ok := range want {
		if !ok {
			t.Errorf("tool %s missing", name)
		}
	}
}

func TestListAndGet(t *testing.T) {
	cs := connect(t)
	out := call(t, cs, "list_components", map[string]any{"platform": "swiftui"})
	comps := out["components"].([]any)
	if len(comps) != 11 {
		t.Fatalf("want 11 swiftui components, got %d", len(comps))
	}
	got := call(t, cs, "get_component", map[string]any{"name": "button", "platform": "flutter"})
	if !strings.Contains(got["source"].(string), "class FButton") {
		t.Errorf("flutter source not returned: %v", got["source"])
	}
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "get_component", Arguments: map[string]any{"name": "MagicThing", "platform": "swiftui"}})
	if err != nil {
		t.Fatalf("error result must not be a protocol error: %v", err)
	}
	if !res.IsError {
		t.Error("unknown component must be a tool error")
	}
}

func TestTokensExport(t *testing.T) {
	cs := connect(t)
	out := call(t, cs, "get_tokens", map[string]any{"set": "color", "platform": "swiftui"})
	export := out["export"].(string)
	if !strings.Contains(export, "static let brand = SwiftUI.Color(hex: \"#0B6E99\")") {
		t.Errorf("swift export wrong:\n%s", export)
	}
	out = call(t, cs, "get_tokens", map[string]any{"platform": "flutter"})
	if !strings.Contains(out["export"].(string), "static const Color colorBrand = Color(0xFF0B6E99)") {
		t.Errorf("dart export wrong:\n%s", out["export"])
	}
}

func TestTokensExportReportsFixedSizes(t *testing.T) {
	cs := connectWith(t, "../catalog/testdata/project-tokens.json")
	out := call(t, cs, "get_tokens", map[string]any{"platform": "swiftui"})
	warnings, _ := out["warnings"].([]any)
	if len(warnings) != 1 || !strings.Contains(warnings[0].(string), "type.body") {
		t.Errorf("want one warning naming type.body, got %v", out["warnings"])
	}
	found := false
	for _, tok := range out["tokens"].([]any) {
		m := tok.(map[string]any)
		if m["path"] == "color.surface" {
			found = m["dark"] == "#0E0C0A"
		}
	}
	if !found {
		t.Errorf("get_tokens must return the dark value of color.surface: %v", out["tokens"])
	}
}

func TestSearch(t *testing.T) {
	cs := connect(t)
	out := call(t, cs, "search", map[string]any{"query": "destructive"})
	comps := out["components"].([]any)
	if len(comps) == 0 || comps[0].(map[string]any)["name"] != "Button" {
		t.Errorf("search destructive should find Button first: %v", comps)
	}
}

func TestCheckUsage(t *testing.T) {
	cs := connect(t)
	src := "VStack {\n  FButton(\"Save\") {}\n  FMagicThing(\"hi\")\n  Rectangle().fill(Color(hex: \"#FF0000\"))\n  Text(\"x\").padding(12)\n  Text(\"y\").padding(FoundationTokens.Space.s3)\n}"
	out := call(t, cs, "check_usage", map[string]any{"source": src})
	findings := out["findings"].([]any)
	kinds := map[string]int{}
	for _, f := range findings {
		kinds[f.(map[string]any)["kind"].(string)]++
	}
	if kinds["unknown-component"] != 1 || kinds["hex-color"] != 1 || kinds["literal-spacing"] != 1 {
		t.Errorf("want 1 of each finding, got %v", kinds)
	}
	if out["clean"].(bool) {
		t.Error("must not be clean")
	}
}

func TestRegisterRejectsUnknownToken(t *testing.T) {
	cs := connect(t)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "register_component", Arguments: map[string]any{
		"name": "Badge", "purpose": "x", "tokens": []string{"color.nope"}, "sources": map[string]string{"swiftui": "struct FBadge {}"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError || !strings.Contains(res.Content[0].(*mcp.TextContent).Text, "color.nope") {
		t.Errorf("unknown token must be rejected: %+v", res)
	}
}
