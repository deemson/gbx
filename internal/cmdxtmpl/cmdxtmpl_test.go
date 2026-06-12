package cmdxtmpl_test

import (
	"errors"
	"testing"

	"github.com/deemson/gbx/internal/cmdxtmpl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mapResolve resolves a tag to its map value, erroring on an unknown tag.
func mapResolve(m map[string]string) func(string) (string, error) {
	return func(tag string) (string, error) {
		v, ok := m[tag]
		if !ok {
			return "", errors.New("unknown tag")
		}
		return v, nil
	}
}

func TestInterpolate(t *testing.T) {
	tests := []struct {
		name string
		args []string
		vals map[string]string
		want []string
	}{
		{
			name: "no tags pass through",
			args: []string{"lazygit", "--all"},
			vals: nil,
			want: []string{"lazygit", "--all"},
		},
		{
			name: "whole element is a tag",
			args: []string{"{{shell}}"},
			vals: map[string]string{"shell": "/bin/zsh"},
			want: []string{"/bin/zsh"},
		},
		{
			name: "embedded tag with prefix and suffix",
			args: []string{"pre-{{x}}-post"},
			vals: map[string]string{"x": "MID"},
			want: []string{"pre-MID-post"},
		},
		{
			name: "value with space stays one element",
			args: []string{"{{cmd}}"},
			vals: map[string]string{"cmd": "/bin/zsh -l"},
			want: []string{"/bin/zsh -l"},
		},
		{
			name: "empty result is preserved in place",
			args: []string{"a", "{{empty}}", "b"},
			vals: map[string]string{"empty": ""},
			want: []string{"a", "", "b"},
		},
		{
			name: "surrounding whitespace is trimmed before resolve",
			args: []string{"{{ env.SHELL }}"},
			vals: map[string]string{"env.SHELL": "/bin/zsh"},
			want: []string{"/bin/zsh"},
		},
		{
			name: "padded and tight tags are equivalent",
			args: []string{"{{ x }}", "{{x}}", "{{\tx\t}}"},
			vals: map[string]string{"x": "v"},
			want: []string{"v", "v", "v"},
		},
		{
			name: "multiple elements and tags",
			args: []string{"{{a}}", "lit", "{{b}}/{{a}}"},
			vals: map[string]string{"a": "1", "b": "2"},
			want: []string{"1", "lit", "2/1"},
		},
		{
			name: "empty input yields empty output",
			args: []string{},
			vals: nil,
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cmdxtmpl.Interpolate(tt.args, mapResolve(tt.vals))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestInterpolateDoesNotMutateInput(t *testing.T) {
	args := []string{"{{x}}", "lit"}
	_, err := cmdxtmpl.Interpolate(args, mapResolve(map[string]string{"x": "v"}))
	require.NoError(t, err)
	assert.Equal(t, []string{"{{x}}", "lit"}, args)
}

func TestInterpolateFailsFastWithContext(t *testing.T) {
	args := []string{"{{ok}}", "lit", "{{ bad }}"}
	got, err := cmdxtmpl.Interpolate(args, mapResolve(map[string]string{"ok": "v"}))
	assert.Nil(t, got)
	require.Error(t, err)
	assert.ErrorContains(t, err, "arg 2")
	assert.ErrorContains(t, err, `tag "bad"`)
	assert.ErrorContains(t, err, "unknown tag")
}
