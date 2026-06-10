package config_test

import (
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/deemson/gbx/internal/config"
	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/suite"
)

type ConfigSuite struct {
	suite.Suite
}

func TestConfigSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &ConfigSuite{})
}

func (s *ConfigSuite) TestBadTOML() {
	testCases := map[string]struct {
		cfg []string
		err []string
	}{
		"bad number": {
			cfg: []string{
				"key = value",
			},
			err: []string{
				"1| key = value",
				" |       ~~~~~ incomplete number",
			},
		},
		"non-terminated string": {
			cfg: []string{
				`key = "value`,
			},
			err: []string{
				`1| key = "value`,
				` |  basic string not terminated by "`,
			},
		},
		"asd": {
			cfg: []string{
				"[key",
			},
			err: []string{
				"1| [key",
				" |  expected character ] but the document ended here",
			},
		},
	}
	for name, testCase := range testCases {
		s.T().Run(name, func(t *testing.T) {
			_, err := config.Load([]byte(strings.Join(testCase.cfg, "\n")))
			s.Require().Error(err)
			var tomlErr *toml.DecodeError
			if s.Assert().ErrorAs(err, &tomlErr) {
				s.Assert().Equal(strings.Join(testCase.err, "\n"), tomlErr.String())
			}
		})
	}
}

func (s *ConfigSuite) TestValidation() {
	testCases := map[string]struct {
		cfg []string
	}{
		"unknown field at root": {
			cfg: []string{
				"unknown = 42",
			},
		},
		"unknown field in actions": {
			cfg: []string{
				"[actions]",
				"unknown = 42",
			},
		},
	}
	for name, testCase := range testCases {
		s.T().Run(name, func(t *testing.T) {
			_, err := config.Load([]byte(strings.Join(testCase.cfg, "\n")))
			if s.Assert().Error(err) {
				spew.Dump(err)
			}
		})
	}
}
