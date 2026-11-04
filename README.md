# foundation-mcp

**A Go program that serves a design system as an MCP server.** A button, a text field, a card are designed **once**; their tokens (color, type, spacing, radius, motion), variants and SwiftUI / Flutter / Compose code live here. Every mobile app built with Claude Code pulls these components through MCP tools. Claude does not design a new button on every screen, it uses what exists; apps stay consistent and use fewer tokens.

| | |
|---|---|
| Version | 0.3.0 (2026-11-04) |
| Language | Go 1.25+, official [`modelcontextprotocol/go-sdk`](https://github.com/modelcontextprotocol/go-sdk) v1.8.0 |
| Transport | stdio (default, for `.mcp.json`) and streamable HTTP (shared server) |
| Repository | [Albaylar/foundation-mcp](https://github.com/Albaylar/foundation-mcp) |
| Used by | [Mobile Development Kit](https://github.com/Albaylar/Mobile-Development-Kit): `/foundation setup` clones, builds and connects this repository |

---

## Contents
1. [Why it exists](#1-why-it-exists)
2. [How it works](#2-how-it-works)
3. [Install](#3-install)
4. [Connect to Claude Code](#4-connect-to-claude-code)
5. [Tools in detail](#5-tools-in-detail)
6. [Catalog layout](#6-catalog-layout)
7. [Tokens and the project's brand](#7-tokens-and-the-projects-brand)
8. [Loading a design](#8-loading-a-design)
9. [Adding a component](#9-adding-a-component)
10. [Tests, CI and quality](#10-tests-ci-and-quality)
11. [Inside the code](#11-inside-the-code)
12. [Limits and roadmap](#12-limits-and-roadmap)
13. [FAQ](#13-faq)
14. [Working rules](#14-working-rules)

---

## 1. Why it exists

Without the server an AI coding agent does this on every screen: reads the design-system file from scratch, pulls layer data from Figma, and re-derives the button. The result:

- **Inconsistency.** Three screens, three buttons: slightly different height, radius, pressed state.
- **Wasted tokens.** Hundreds of lines of design file and Figma output enter the context again for every screen.
- **Forgotten details.** Loading state, VoiceOver label, 44 pt touch target have to be re-thought every time.

With the server a screen costs one `list_components` (a short table), one `get_component` per used component (3–6, each returns ready source code) and one `get_tokens`. The code is **pasted, not invented**. Claude Code loads MCP tool schemas lazily; the six tools cost almost nothing until called.

## 2. How it works

```
 Figma variables / Claude Design / design-system.md
                │   /foundation setup extracts the tokens
                ▼
 app/docs/design/tokens.json          the brand's values (only what differs)
                │   --tokens
                ▼
 foundation-mcp/catalog/tokens.json   base tokens (names are fixed)
 foundation-mcp/catalog/components/   Button, TextField, Card … (SwiftUI, Flutter)
                │
                ▼
          foundation (Go binary, MCP server)
                │   stdio
                ▼
          Claude Code: /design, /ui, /qa
```

**Components are shared, values belong to the project.** Components reference token **names** (`color.brand`, `space.4`). Each app passes its own `tokens.json`; the server merges it over the base tokens. The same `FButton` is brown in a bakery app and blue in a tide app; the code is identical.

## 3. Install

Requires Go 1.25 or newer (`brew install go`).

```bash
git clone https://github.com/Albaylar/foundation-mcp.git ~/Developer/foundation-mcp
```
```bash
cd ~/Developer/foundation-mcp && go build -o foundation ./cmd/foundation
```
```bash
cd ~/Developer/foundation-mcp && ./foundation --catalog ./catalog --check
```
Expected output: `ok: 25 tokens (catalog tokens), 4 components`.

| Flag | Meaning |
|---|---|
| `--catalog <dir>` | shared catalog (default `catalog`) |
| `--tokens <file>` | the project's token file, merged over the base tokens |
| `--http :8765` | streamable HTTP instead of stdio (shared team server) |
| `--check` | load and validate the catalog, then exit (setup and CI check); warnings about the SwiftUI export go to stderr |
| `--export swiftui\|flutter` | print the generated `FoundationTokens` file (with `--tokens`: the project's values), then exit; warnings go to stderr |
| `--version` | print the version |

With the Mobile Development Kit, `/foundation setup` does all of this.

## 4. Connect to Claude Code

In the app project's root, `.mcp.json` (paths must be **absolute**):

```json
{
  "mcpServers": {
    "foundation": {
      "command": "/Users/YOU/Developer/foundation-mcp/foundation",
      "args": [
        "--catalog", "/Users/YOU/Developer/foundation-mcp/catalog",
        "--tokens", "/Users/YOU/Developer/my-app/docs/design/tokens.json"
      ]
    }
  }
}
```

Claude Code loads MCP servers **when a session starts**: after writing the file, restart Claude Code in that folder (or reconnect in `/mcp`) and approve the `foundation` server once. The tools appear as `mcp__foundation__list_components` and so on.

One server for a team: `./foundation --catalog ./catalog --http :8765` and `"url": "http://host:8765"` in `.mcp.json`. The HTTP transport has no authentication yet; use it only on a trusted network.

## 5. Tools in detail

All tools return structured (JSON) output. On failure they return `isError: true` with a readable message; it is not a protocol error, so the agent can read it and continue.

### `list_components`
Lists the catalog. **Called before designing any control.**

| Input | Values | |
|---|---|---|
| `platform` | `swiftui` / `flutter` / `compose` | optional; components with source for that platform |
| `tag` | text | optional, e.g. `form` |

Output: `components[]` → `name`, `purpose`, `status`, `version`, `variants[]`, `platforms[]`, `tags[]`.

### `get_component`
Everything about one component: the manifest (tokens, variants, accessibility note, Figma node) and the **source code** for the requested platform.

| Input | | Example |
|---|---|---|
| `name` | case-insensitive | `button` |
| `platform` | required | `flutter` |

Unknown component, or no source for that platform → error naming the available platforms.

### `get_tokens`
The token list and, on request, a **generated source file**.

| Input | Values | |
|---|---|---|
| `set` | `color` / `type` / `space` / `radius` / `motion` | empty = all |
| `platform` | `swiftui` / `flutter` | if given, `export` holds the `FoundationTokens` file |

SwiftUI output: `FoundationTokens.Color.brand`, `.Space.s4`, `.Typography.label`, `.Radius.field`, `.Motion.fast`. Flutter output: `FoundationTokens.colorBrand`, `space4`, `typeLabel`. These files are **never edited by hand**; they are regenerated whenever tokens change.

Each token carries `dark` when it has a dark value. `warnings` lists what the export cannot honor (see [Dynamic Type](#dynamic-type)).

### `search`
Search by purpose: `"destructive confirmation"`, `"input with error"`. Every word is matched against name, purpose, tags and variant names; the best match comes first. Tokens are searched too.

### `check_usage`
Scans a screen's source code and returns three kinds of findings:

| Kind | Catches | Advice |
|---|---|---|
| `hex-color` | `#FF0000`, `Color(red:`, `0xFF…` | use a color token |
| `literal-spacing` | `.padding(12)`, `EdgeInsets.all(8)`, `spacing: 16` (when no token is used) | use a space token |
| `unknown-component` | an `F…` component not in the catalog (`FToast`) | register it through `/design` first |

Comment lines are skipped. v1 is regex-based: cheap and platform-agnostic. A false positive is explained in one line and skipped.

### `register_component`
Adds an approved new component to the catalog: manifest + platform sources. **Only the design step calls it** (`/design` in the kit). It rejects unknown tokens, existing names, components without source and unknown platforms. On success it writes `catalog/components/<name>/`; the change is committed in this repository and merged by pull request (see §9).

### Resource and prompt
- `foundation://tokens.json`: all merged tokens.
- `use_foundation` prompt: the rules a UI-building agent follows (search first; paste the code, never rewrite it; no literal values, only tokens; no component outside the catalog; `check_usage` before reporting).

## 6. Catalog layout

```
catalog/
  tokens.json                         base tokens (W3C Design Tokens)
  components/
    button/
      manifest.json                   the component's identity
      swiftui/Button.swift            SwiftUI implementation + #Preview
      flutter/button.dart             Flutter implementation
    text_field/  …
    card/        …
```

`manifest.json` fields:

| Field | Meaning |
|---|---|
| `name` | name in the catalog (`Button`); in code with the `F` prefix (`FButton`) |
| `purpose` | when to use it; `search` matches it |
| `tags` | `form`, `action`, `container` … |
| `status` | `stable` / `draft` / `deprecated` |
| `version` | semver; bumped when behavior changes |
| `variants` | `[{name, props}]`, e.g. `primary`, `secondary`, `destructive` |
| `tokens` | token paths it uses; all must exist in `tokens.json` |
| `accessibility` | label, minimum size, announced states |
| `figma_node` | the component's node id in Figma (for Code Connect) |
| `platforms` | `{"swiftui": "swiftui/Button.swift", …}` |

Every manifest is validated on load: a missing token or a missing source file and the server does not start. A broken catalog never reaches an app.

In the catalog today:

| Component | Variants | SwiftUI | Flutter |
|---|---|---|---|
| **Button** | primary, secondary, destructive; loading state | `FButton` | `FButton` |
| **TextField** | default (label, helper, error), search | `FTextField` | `FTextField` |
| **Card** | default, tappable | `FCard` | `FCard` |
| **EmptyState** | default (title, message, icon) | `FEmptyState` | none |

## 7. Tokens and the project's brand

`tokens.json` uses the [W3C Design Tokens](https://design-tokens.github.io/community-group/format/) format:

```json
{ "color": { "brand": { "$type": "color", "$value": "#0B6E99", "$description": "Primary accent" } } }
```

The server flattens it to paths like `color.brand`. Groups: `color` (brand, brand-strong, surface, surface-2, text-1, text-2, line, danger, on-brand, on-danger), `type` (title, body, label, caption), `space` (1–6 = 4–32), `radius` (field, card, pill), `motion` (fast, base).

**Project tokens (`--tokens`).** The project file is merged over the base file:
- same path → the project's value wins,
- new path → added,
- path the project does not set → the base value stays.

Write **only what differs** into the project file:

```json
{ "color": { "brand": { "$type": "color", "$value": "#7A4B16" } } }
```

Rule: **token names never change.** Inventing `color.primary` in one project means every component is blind to that color. A missing project file is not silently ignored; the server reports an error and does not start.

A project that sets `color.danger` usually sets `color.on-danger` too: it is the text on a destructive Button, and it has to stay readable on that project's `color.danger` in every mode it supports.

### Dark mode

A color token may carry its dark value next to the base value, which is the light value:

```json
{ "color": { "surface": { "$type": "color", "$value": "#FFFFFF", "$extensions": { "mode": { "dark": "#0E0C0A" } } } } }
```

| Export | With a dark value | Without one |
|---|---|---|
| SwiftUI | `SwiftUI.Color(light: "#FFFFFF", dark: "#0E0C0A")`, follows the system appearance (iOS; the file then imports UIKit) | `SwiftUI.Color(hex: "#FFFFFF")`, the same in both modes |
| Flutter | the light value only; the file says so in a comment | the value |

A project token replaces the whole base token, dark value included, so a value set without a dark value has none.

### Dynamic Type

The SwiftUI export maps a type token by its `fontSize` to the iOS text style with that default size, and applies the token's weight. Project sizes follow the same table, so a project's `type.title` at 34 becomes Large Title.

| `fontSize` | 34 | 28 | 22 | 20 | 17 | 16 | 15 | 13 | 12 | 11 |
|---|---|---|---|---|---|---|---|---|---|---|
| `Font.` | `largeTitle` | `title` | `title2` | `title3` | `body` | `callout` | `subheadline` | `footnote` | `caption` | `caption2` |

The base tokens give `type.title` = `Font.title.weight(.bold)`, `type.body` = `Font.body`, `type.label` = `Font.subheadline.weight(.semibold)`, `type.caption` = `Font.footnote`.

**Any other size stays exactly as the project set it, and does not scale**, because SwiftUI has no scaling system font at an arbitrary size. That is never silent: the generated file has a `#warning`, so Xcode shows it on every build, and `--check`, `--export swiftui` and `get_tokens` report it. To scale, use a size from the table.

The `fontFamily` and `lineHeight` of a type token are not used by the SwiftUI export. In Flutter the text scale applies at any size, so nothing changes there.

## 8. Loading a design

The kit's `/foundation setup` does this. By hand:

**Figma.** With the Figma MCP server connected, call `get_variable_defs` on the library page or on a frame that uses every style (the desktop server needs Dev Mode in the file). Map each Figma variable to a base path (`Brand/Primary` → `color.brand`) and write the values to the project's `tokens.json`. Put the component's Figma node id in `figma_node`. If Figma Code Connect maps the component to the source in this repository, Figma's `get_design_context` returns `FButton` instead of raw layers (the Code Connect label is documented for SwiftUI today).

**Claude Design.** Take the color, type, spacing and radius values from the project's Design System artifact and write them to the same paths.

**design-system.md.** Take hex colors, the type scale, spacing and radii.

Always validate at the end:
```bash
cd ~/Developer/foundation-mcp && ./foundation --catalog ./catalog --tokens /path/to/my-app/docs/design/tokens.json --check
```

## 9. Adding a component

1. The component exists in the **design** first (a Figma component or the Claude Design System artifact). It is never born in code and added to the design later.
2. The kit's `/design` step calls `register_component`: manifest + SwiftUI and/or Flutter source. The server validates tokens and the name.
3. The change stays in this repository: branch `feat/component-<name>`, commit, pull request to `Albaylar/foundation-mcp`; open the pull request against `main`.
4. After the merge every app picks it up with `/foundation sync`.

By hand: `catalog/components/<name>/manifest.json` plus the platform folders, then `go test ./...` and `--check`.

Source rules: no literal values, only `FoundationTokens`; minimum 44 pt / 48 dp; an accessibility label; loading and disabled states; a `#Preview` in SwiftUI.

## 10. Tests, CI and quality

```bash
cd ~/Developer/foundation-mcp && go vet ./... && go test ./...
```

| Test | What it proves |
|---|---|
| `TestToolsAreRegistered` | all six tools are registered |
| `TestListAndGet` | platform filter, source returned, unknown component = tool error (not a protocol error) |
| `TestTokensExport` | correct values in the SwiftUI and Dart output |
| `TestSearch` | "destructive" → Button first |
| `TestCheckUsage` | one finding of each kind is caught, token usage counts as clean |
| `TestRegisterRejectsUnknownToken` | an unknown token is rejected |
| `TestProjectTokensOverrideAndAdd` | the project value wins, a new token is added, an unset base value stays |
| `TestProjectTokensMissingFile` | a missing project file is not swallowed |
| `TestCheckUsageIgnoresTokenReferences` | clean source produces no findings |
| `TestCheckUsageAcceptsTypesOfCatalogComponents` | `FButtonVariant` belongs to `FButton` and is not flagged; `FToast` still is |
| `TestSwiftExportAvoidsReservedNames` | no nested `enum Type` in generated Swift |
| `TestCatalogSourcesMatchTheGeneratedTokenNames` | catalog sources use the generated names |
| `TestTokensExportReportsFixedSizes` | `get_tokens` returns dark values and warnings |
| `TestDarkValueIsReadFromColorTokens` | the dark value is read on color tokens only; the base value stays the light value |
| `TestSwiftExportAdaptsColorsWithADarkValue` | a dark value gives an adaptive color, a color without one stays fixed |
| `TestSwiftExportWithoutDarkValuesStaysAsItWas` | no dark values: no UIKit, no adaptive code |
| `TestSwiftExportMapsBaseTypeToScalingTextStyles` | the four base type tokens map to their text styles, with no warning |
| `TestSwiftExportFollowsAProjectSizeToItsTextStyle` | a project size follows the table, not the token's name |
| `TestSwiftExportKeepsASizeWithNoTextStyleAndSaysSo` | another size is kept and reported in the file and in `ExportWarnings` |
| `TestFlutterExportStaysLightOnlyAndSaysSo` | Flutter keeps the light value and states it |
| `TestOnDangerKeepsTodaysLookInTheBaseTokens` | base `color.on-danger` equals base `color.on-brand` |
| `TestDestructiveButtonUsesOnDanger` | both Button sources and the manifest use `color.on-danger` |

The tool tests drive the server through an **in-memory MCP transport** with a real client, so they run end to end including schema validation. New tests are verified by sabotage: break the behavior and the test must turn red (the project-token tests were checked this way).

CI (`.github/workflows/ci.yml`) runs `go vet`, `go test` and `--check`, and **really compiles the catalog**: `scripts/compile-catalog.sh` generates the token file and typechecks every SwiftUI component against the iOS SDK with Swift 6, once with the base tokens and once with `internal/catalog/testdata/project-tokens.json` (the base tokens never produce adaptive colors or fixed sizes), and runs `flutter analyze` on every Flutter component. String comparisons are not enough here: an early version generated a Swift enum named `Type`, which is reserved, and every test passed while nothing compiled.

## 11. Inside the code

```
cmd/foundation/main.go        flags, server setup, stdio / HTTP
internal/catalog/catalog.go   loading, merging, validation, list/get/search/check/register (no MCP, plain Go)
internal/catalog/export.go    token → SwiftUI / Dart source generation
internal/tools/tools.go       MCP tools, resource, prompt (thin adapter over catalog)
internal/*/…_test.go          tests
```

Tool inputs are typed Go structs; their `jsonschema` tags become the tool schema (`mcp.AddTool`). Optional fields are `omitempty`: the SDK also validates the output schema, so an empty map or slice sent as `null` makes the call fail with a protocol error (this happened once; the tests catch it now).

## 12. Limits and roadmap

Not done yet:
- Compose sources (SwiftUI and Flutter only)
- dark mode in Flutter: the export is `static const` values read directly by the components, so adapting needs a theme-based design; light values only until a Flutter project needs it
- an automatic token import command from the Figma REST API
- authentication for the HTTP transport
- a versioned changelog resource
- a platform-specific parser for `check_usage` (regex today)

Recommended next: Compose sources (only once a native Android project exists), then the Figma import.

## 13. FAQ

**A separate catalog per app?** No. Components are shared, brands are separate through `--tokens`. If an app seems to need its own component, ask whether it is really specific; usually another app can use it too.

**Claude does not see the tools.** Are the paths in `.mcp.json` absolute? Did you restart Claude Code in that folder? Is `foundation` connected in `/mcp`? Does `./foundation … --check` fail? In the kit, `/foundation status` checks all of this at once.

**I changed a token and the app still shows the old color.** Regenerate the `FoundationTokens` file (`./foundation --catalog … --tokens … --export swiftui`, `get_tokens platform=…`, or `/foundation sync` in the kit).

**Why Go?** One static binary, no runtime, fast start; the official MCP SDK for Go is stable.

## 14. Working rules

- Branches `feat/component-<name>`, `fix/<name>`; no direct commits to `main`, only pull requests.
- Test first (red), then code (green); 0 tests run counts as failed.
- No literal values in the catalog; every component uses tokens only.

## License

[MIT](LICENSE), © 2026 Furkan Deniz Albaylar.
