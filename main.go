package main

import (
	"runtime/debug"

	"github.com/deemson/gbx/internal/cmd"
)

// version is set at build time via -ldflags "-X main.version=v1.2.3"; empty in
// plain `go build`, where resolveVersion falls back to the embedded commit hash
// (or, failing that, the TUI's "dev" default).
var version string

func main() {
	cmd.Main(resolveVersion())
}

// resolveVersion honours the ldflags version if set, otherwise falls back to the
// abbreviated commit hash Go embeds in the binary's build info. This means
// `go install`d builds — which never see our release ldflags — still report an
// identifiable revision instead of "dev". Returns "" when neither is available.
func resolveVersion() string {
	if version != "" {
		return version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, s := range info.Settings {
		if s.Key == "vcs.revision" && len(s.Value) >= 7 {
			return s.Value[:7]
		}
	}
	return ""
}
