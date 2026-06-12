package tui

import (
	"fmt"
	"os"
	"strings"
)

// resolveTag resolves a {{ }} tag in an action command vector. Only the env.*
// namespace is supported: env.NAME yields the environment variable NAME. An
// unknown namespace or an unset variable is a hard error, so a typo aborts the
// action before it launches rather than running the wrong thing. The cursored
// repo's directory is supplied to exec directly (cmd.Dir), not via a tag.
func resolveTag(tag string) (string, error) {
	name, ok := strings.CutPrefix(tag, "env.")
	if !ok {
		return "", fmt.Errorf("unknown tag %q (only env.* is supported)", tag)
	}
	val, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("environment variable %s is not set", name)
	}
	return val, nil
}
