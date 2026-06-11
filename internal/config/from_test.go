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
		"unknown field in actions": {
			cfg: []string{
				"[actions]",
				"unknown = 42",
			},
			err: []string{
				"actions.enter: required field is missing",
				"actions.shift-enter: required field is missing",
				"actions.unknown: unknown field",
			},
		},
		"wrong type": {
			cfg: []string{
				"[actions]",
				`enter = "lazygit"`,
				`shift-enter = ["bash"]`,
			},
			err: []string{
				"actions.enter: expected array, got string",
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
