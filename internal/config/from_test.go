package config_test

import (
	"strings"
	"testing"

	"github.com/deemson/gbx/internal/config"
	"github.com/stretchr/testify/suite"
)

type FromBytesSuite struct {
	suite.Suite
}

func TestFromBytesSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &FromBytesSuite{})
}

func (s *FromBytesSuite) TestBadTOML() {
	testCases := map[string]struct {
		cfg []string
		err string
	}{
		"bad number": {
			cfg: []string{"key = value"},
			err: "line 1, column 7: incomplete number",
		},
		"non-terminated string": {
			cfg: []string{`key = "value`},
			err: `line 1, column 1: basic string not terminated by "`,
		},
		"unterminated table header": {
			cfg: []string{"[key"},
			err: "line 1, column 1: expected character ] but the document ended here",
		},
	}
	for name, testCase := range testCases {
		s.T().Run(name, func(t *testing.T) {
			_, err := config.FromBytes([]byte(strings.Join(testCase.cfg, "\n")))
			s.Require().Error(err)
			var tomlErr *config.TOMLError
			if s.Assert().ErrorAs(err, &tomlErr) {
				s.Assert().Equal(testCase.err, tomlErr.Error())
			}
		})
	}
}

func (s *FromBytesSuite) TestValidation() {
	testCases := map[string]struct {
		cfg []string
		err []string
	}{
		"unknown field at root": {
			cfg: []string{
				"unknown = 42",
			},
			err: []string{
				"actions: required field is missing",
				"unknown: unknown field",
			},
		},
		"unknown field in action": {
			cfg: []string{
				"[[actions]]",
				`label = "x"`,
				`command = ["x"]`,
				"unknown = 42",
			},
			err: []string{
				"actions.0.unknown: unknown field",
			},
		},
		"missing action fields": {
			cfg: []string{
				"[[actions]]",
				`label = "x"`,
			},
			err: []string{
				"actions.0.command: required field is missing",
			},
		},
		"wrong type": {
			cfg: []string{
				`actions = "lazygit"`,
			},
			err: []string{
				"actions: expected array, got string",
			},
		},
		"too many actions": {
			cfg: []string{
				`[[actions]]`, `label="1"`, `command=["a"]`,
				`[[actions]]`, `label="2"`, `command=["a"]`,
				`[[actions]]`, `label="3"`, `command=["a"]`,
				`[[actions]]`, `label="4"`, `command=["a"]`,
				`[[actions]]`, `label="5"`, `command=["a"]`,
				`[[actions]]`, `label="6"`, `command=["a"]`,
				`[[actions]]`, `label="7"`, `command=["a"]`,
				`[[actions]]`, `label="8"`, `command=["a"]`,
				`[[actions]]`, `label="9"`, `command=["a"]`,
				`[[actions]]`, `label="10"`, `command=["a"]`,
			},
			err: []string{
				"actions: must have at most 9 items",
			},
		},
		"empty actions": {
			cfg: []string{
				`actions = []`,
			},
			err: []string{
				"actions: must have at least 1 items",
			},
		},
	}
	for name, testCase := range testCases {
		s.T().Run(name, func(t *testing.T) {
			_, err := config.FromBytes([]byte(strings.Join(testCase.cfg, "\n")))
			s.Require().Error(err)
			var valErr *config.ValidationError
			if s.Assert().ErrorAs(err, &valErr) {
				s.Assert().Equal(strings.Join(testCase.err, "\n"), valErr.Error())
			}
		})
	}
}
