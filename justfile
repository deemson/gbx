[default]
_default:
	@just --list

# Run golangci-lint over the module.
lint:
	golangci-lint run

# Run the test suite.
test:
	go test ./...

# Lint and test, as CI does.
check: lint test

# Destination for the bundled upstream examples.
examples_dir := ".claude/skills/charm-tui/examples"

# Pull Bubble Tea / Lipgloss examples from upstream, version-matched to go.mod.
sync-charm-land-examples:
    #!/usr/bin/env bash
    set -euo pipefail
    dest="{{examples_dir}}"
    rm -rf "$dest"
    mkdir -p "$dest"
    : > "$dest/VERSIONS.txt"

    sync() {
      local module="$1" repo="$2"; shift 2
      local version
      version=$(go list -m "$module" | awk '{print $2}')
      echo "==> $repo @ $version"
      local tmp
      tmp=$(mktemp -d)
      # Drop git's benign "refs/tags/X is not a commit!" warning, emitted when
      # shallow-cloning an annotated tag; keep every other stderr line.
      git -c advice.detachedHead=false clone --quiet --depth 1 --branch "$version" \
        "https://github.com/charmbracelet/$repo" "$tmp" \
        2> >(grep -v 'is not a commit!' >&2 || true)
      local got=0
      for sub in "$@"; do
        if [ -d "$tmp/$sub" ]; then
          mkdir -p "$dest/$repo"
          cp -R "$tmp/$sub" "$dest/$repo/"
          echo "    + $sub"
          got=1
        else
          echo "    - $sub (absent, skipped)"
        fi
      done
      [ "$got" -eq 1 ] && echo "$repo $version" >> "$dest/VERSIONS.txt"
      rm -rf "$tmp"
    }

    sync charm.land/bubbletea/v2 bubbletea examples tutorials
    sync charm.land/lipgloss/v2 lipgloss examples

    echo "Done -> $dest"
