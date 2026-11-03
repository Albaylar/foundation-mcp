// foundation serves the design-system catalog over MCP.
//
//	foundation --catalog ./catalog                                   stdio (for .mcp.json)
//	foundation --catalog ./catalog --tokens ./app/design/tokens.json  stdio, with the app's brand tokens
//	foundation --catalog ./catalog --http :8765                       streamable HTTP (shared instance)
//	foundation --catalog ./catalog --tokens … --check                 validate and exit
//	foundation --catalog ./catalog --tokens … --export swiftui        print the generated token file
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/Albaylar/foundation-mcp/internal/catalog"
	"github.com/Albaylar/foundation-mcp/internal/tools"
)

const version = "0.2.1"

func main() {
	root := flag.String("catalog", "catalog", "path to the shared catalog directory")
	projectTokens := flag.String("tokens", "", "optional project tokens.json merged over the catalog's tokens (the app's brand)")
	addr := flag.String("http", "", "serve streamable HTTP on this address instead of stdio, e.g. :8765")
	check := flag.Bool("check", false, "load and validate the catalog, then exit")
	showVersion := flag.Bool("version", false, "print the version and exit")
	export := flag.String("export", "", "print the generated FoundationTokens source for swiftui or flutter, then exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	c, err := catalog.LoadWith(*root, *projectTokens)
	if err != nil {
		fmt.Fprintln(os.Stderr, "catalog:", err)
		os.Exit(1)
	}
	if *export != "" {
		out := c.Export(*export, "")
		if out == "" {
			fmt.Fprintf(os.Stderr, "no export for platform %q (use swiftui or flutter)\n", *export)
			os.Exit(1)
		}
		fmt.Print(out)
		return
	}
	if *check {
		src := "catalog tokens"
		if *projectTokens != "" {
			src = "catalog tokens + " + *projectTokens
		}
		fmt.Printf("ok: %d tokens (%s), %d components\n", len(c.Tokens), src, len(c.Components))
		return
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "foundation", Version: version}, nil)
	tools.Register(server, c)

	if *addr != "" {
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
		log.Printf("foundation %s serving %s on %s", version, *root, *addr)
		log.Fatal(http.ListenAndServe(*addr, handler))
	}
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
