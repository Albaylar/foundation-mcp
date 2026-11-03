// Package tools registers the MCP tools, resources and the prompt on a server.
// Each tool is a thin adapter over internal/catalog.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Albaylar/foundation-mcp/internal/catalog"
)

// Register wires every tool, resource and prompt onto s.
func Register(s *mcp.Server, c *catalog.Catalog) {
	mcp.AddTool(s, &mcp.Tool{Name: "list_components", Description: "List the design-system components, optionally for one platform (swiftui, flutter, compose) or tag. Call this before designing any control."},
		func(ctx context.Context, req *mcp.CallToolRequest, in ListIn) (*mcp.CallToolResult, ListOut, error) {
			return nil, ListOut{Components: c.List(in.Platform, in.Tag)}, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "get_component", Description: "Get one component: tokens, variants, accessibility notes and the implementation source for a platform. Paste the source; do not rewrite it."},
		func(ctx context.Context, req *mcp.CallToolRequest, in GetIn) (*mcp.CallToolResult, GetOut, error) {
			comp, ok := c.Get(in.Name)
			if !ok {
				return toolError(fmt.Sprintf("no component %q; try search", in.Name)), GetOut{}, nil
			}
			src, err := c.Source(comp, in.Platform)
			if err != nil {
				return toolError(err.Error()), GetOut{}, nil
			}
			return nil, GetOut{Manifest: comp.Manifest, Platform: in.Platform, Source: src}, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "get_tokens", Description: "Get design tokens of one set (color, type, space, radius, motion) or all, plus a generated FoundationTokens source file for the platform."},
		func(ctx context.Context, req *mcp.CallToolRequest, in TokensIn) (*mcp.CallToolResult, TokensOut, error) {
			return nil, TokensOut{Tokens: c.TokensIn(in.Set), Export: c.Export(in.Platform, in.Set)}, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "search", Description: "Find components and tokens by purpose, name, tag or variant (e.g. 'destructive confirmation', 'input with error')."},
		func(ctx context.Context, req *mcp.CallToolRequest, in SearchIn) (*mcp.CallToolResult, SearchOut, error) {
			comps, toks := c.Search(in.Query)
			return nil, SearchOut{Components: comps, Tokens: toks}, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "check_usage", Description: "Scan screen source for hard-coded colors and spacing and for F-prefixed components that are not in the catalog. Run on every new screen before reporting."},
		func(ctx context.Context, req *mcp.CallToolRequest, in CheckIn) (*mcp.CallToolResult, CheckOut, error) {
			f := c.CheckUsage(in.Source)
			return nil, CheckOut{Findings: f, Clean: len(f) == 0}, nil
		})

	mcp.AddTool(s, &mcp.Tool{Name: "register_component", Description: "Add a new, approved component to the catalog with its manifest and platform sources. Only /design calls this, after the component exists in the design source. Fails if a token is unknown or the name exists."},
		func(ctx context.Context, req *mcp.CallToolRequest, in RegisterIn) (*mcp.CallToolResult, RegisterOut, error) {
			m := catalog.Manifest{Name: in.Name, Purpose: in.Purpose, Tags: in.Tags, Status: orDefault(in.Status, "draft"), Version: orDefault(in.Version, "0.1.0"), Variants: in.Variants, Tokens: in.Tokens, Accessibility: in.Accessibility, FigmaNode: in.FigmaNode}
			if err := c.Register(m, in.Sources); err != nil {
				return toolError(err.Error()), RegisterOut{}, nil
			}
			return nil, RegisterOut{Registered: in.Name, Platforms: keysOf(in.Sources)}, nil
		})

	s.AddResource(&mcp.Resource{URI: "foundation://tokens.json", Name: "tokens", MIMEType: "application/json", Description: "All design tokens (W3C Design Tokens format, flattened)."},
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			b, _ := json.MarshalIndent(c.TokensIn(""), "", "  ")
			return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: req.Params.URI, MIMEType: "application/json", Text: string(b)}}}, nil
		})

	s.AddPrompt(&mcp.Prompt{Name: "use_foundation", Description: "Instructions for building a screen with the design system."},
		func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
			return &mcp.GetPromptResult{Messages: []*mcp.PromptMessage{{Role: "user", Content: &mcp.TextContent{Text: usePrompt}}}}, nil
		})
}

const usePrompt = `Build the screen with the Foundation design system only.
1. Call search or list_components before designing any control. Use get_component and paste its source; never rewrite a component.
2. Colors, type, spacing, radius and motion come only from get_tokens (the generated FoundationTokens file). No hex values, no literal paddings.
3. A control that does not exist in the catalog is not written: report it; it must be designed and registered through /design first.
4. Before reporting, run check_usage on every new screen file and fix every finding.`

// Input and output types. jsonschema tags become the tool schema.
type ListIn struct {
	Platform string `json:"platform,omitempty" jsonschema:"swiftui, flutter or compose; empty = all"`
	Tag      string `json:"tag,omitempty" jsonschema:"filter by tag, e.g. form"`
}
type ListOut struct {
	Components []catalog.Summary `json:"components"`
}

type GetIn struct {
	Name     string `json:"name" jsonschema:"component name, e.g. Button"`
	Platform string `json:"platform" jsonschema:"swiftui, flutter or compose"`
}
type GetOut struct {
	catalog.Manifest
	Platform string `json:"platform"`
	Source   string `json:"source"`
}

type TokensIn struct {
	Set      string `json:"set,omitempty" jsonschema:"color, type, space, radius or motion; empty = all"`
	Platform string `json:"platform,omitempty" jsonschema:"swiftui or flutter for a generated source file"`
}
type TokensOut struct {
	Tokens []catalog.Token `json:"tokens"`
	Export string          `json:"export,omitempty"`
}

type SearchIn struct {
	Query string `json:"query" jsonschema:"words describing the need"`
}
type SearchOut struct {
	Components []catalog.Summary `json:"components"`
	Tokens     []catalog.Token   `json:"tokens"`
}

type CheckIn struct {
	Source string `json:"source" jsonschema:"the screen source code to check"`
}
type CheckOut struct {
	Findings []catalog.Finding `json:"findings"`
	Clean    bool              `json:"clean"`
}

type RegisterIn struct {
	Name          string            `json:"name"`
	Purpose       string            `json:"purpose"`
	Tags          []string          `json:"tags,omitempty"`
	Status        string            `json:"status,omitempty" jsonschema:"stable, draft or deprecated; default draft"`
	Version       string            `json:"version,omitempty"`
	Variants      []catalog.Variant `json:"variants,omitempty"`
	Tokens        []string          `json:"tokens" jsonschema:"token paths used, e.g. color.brand"`
	Accessibility string            `json:"accessibility,omitempty"`
	FigmaNode     string            `json:"figma_node,omitempty"`
	Sources       map[string]string `json:"sources" jsonschema:"platform -> implementation source"`
}
type RegisterOut struct {
	Registered string   `json:"registered"`
	Platforms  []string `json:"platforms,omitempty"`
}

func toolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: msg}}}
}

func orDefault(v, d string) string {
	if strings.TrimSpace(v) == "" {
		return d
	}
	return v
}

func keysOf(m map[string]string) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}
