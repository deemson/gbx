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

# The fixture is regenerated per clip, so the mutating pull clip always has its
# behind-repo and the order clips are filmed in never matters.

# Film demos/<name>.tape into assets/<name>.gif.
demo name:
    go build -o gbx .
    mkdir -p assets
    just gen-demo-fixture
    cd {{demo_fixture_dir}} && PATH="{{justfile_directory()}}:$PATH" vhs --output "{{justfile_directory()}}/assets/{{name}}.gif" "{{justfile_directory()}}/demos/{{name}}.tape"

# Film every demos/*.tape.
demo-all:
    for tape in demos/*.tape; do just demo "$(basename "$tape" .tape)"; done
