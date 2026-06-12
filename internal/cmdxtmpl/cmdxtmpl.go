// Package cmdxtmpl expands fasttemplate tags ({{ }}) in command argument
// vectors, delegating each tag to a caller-supplied resolver. It builds and
// runs nothing: strings in, strings out.
package cmdxtmpl

import (
	"fmt"
	"io"
	"strings"

	"github.com/valyala/fasttemplate"
)

const (
	startTag = "{{"
	endTag   = "}}"
)

// Interpolate returns a new slice the same length as args, with every {{ }} tag
// in each element replaced by the result of resolve. Whitespace surrounding the
// inner text is trimmed before resolve sees it, so {{ env.SHELL }} and
// {{env.SHELL}} are equivalent. Element count is invariant: tags are never split
// on whitespace and empty results are preserved in place.
//
// The first resolve error aborts the call and is returned wrapped with the
// element index and tag; on success the input is never mutated.
func Interpolate(args []string, resolve func(tag string) (string, error)) ([]string, error) {
	out := make([]string, len(args))
	for i, arg := range args {
		s, err := fasttemplate.ExecuteFuncStringWithErr(arg, startTag, endTag, func(w io.Writer, tag string) (int, error) {
			tag = strings.TrimSpace(tag)
			val, err := resolve(tag)
			if err != nil {
				return 0, fmt.Errorf("arg %d: tag %q: %w", i, tag, err)
			}
			return w.Write([]byte(val))
		})
		if err != nil {
			return nil, err
		}
		out[i] = s
	}
	return out, nil
}
