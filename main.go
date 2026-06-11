package main

import "github.com/deemson/gbx/internal/cmd"

// version is set at build time via -ldflags "-X main.version=v1.2.3"; empty in
// plain `go build`, where the TUI falls back to "dev".
var version string

func main() {
	cmd.Main(version)
}
