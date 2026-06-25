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

# Where the throwaway demo repos are built (regenerated on every run).
demo_fixture_dir := "/tmp/gbx-demo-fixture"

# Build the throwaway tree of git repos the demo films against.
gen-demo-fixture:
    rm -rf {{demo_fixture_dir}} {{demo_fixture_dir}}-remotes
    GBX_FIXTURE_DIR={{demo_fixture_dir}} go test -tags fixture -run TestGenerateDemoFixture ./internal/demo/ -count=1

# Each clip regenerates the throwaway fixture first, so the mutating pull clip
# always has its behind-repo and the order clips are filmed in never matters.
# Record demo GIFs into assets/ — no arg films every demos/*.tape, a name (e.g.
# `just demo filter`) films only demos/<name>.tape.
demo name="":
    #!/usr/bin/env bash
    set -euo pipefail
    cd {{justfile_directory()}}
    go build -o gbx .
    mkdir -p assets
    if [ -n "{{name}}" ]; then
      tapes=("demos/{{name}}.tape")
    else
      tapes=(demos/*.tape)
    fi
    for tape in "${tapes[@]}"; do
      tname=$(basename "$tape" .tape)
      echo "==> filming $tname"
      just gen-demo-fixture
      ( cd {{demo_fixture_dir}} && PATH="{{justfile_directory()}}:$PATH" \
        vhs --output "{{justfile_directory()}}/assets/$tname.gif" "{{justfile_directory()}}/$tape" )
    done

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
