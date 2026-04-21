package git

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type StatusSuite struct {
	suite.Suite
}

func TestStatusSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &StatusSuite{})
}

func (s *StatusSuite) TestBasic() {
	// dir := s.T().TempDir()
	// ctx := context.Background()
	// s.Require().NoError(err)
}
