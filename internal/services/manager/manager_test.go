package manager

import (
	"github.com/blkmlk/file-storage/deps"
	"github.com/stretchr/testify/suite"
	"go.uber.org/dig"
)

type testSuite struct {
	suite.Suite
	manager Manager
}

func (t *testSuite) SetupSuite() {
	container := dig.New()

	container.Provide(deps.NewDB)
	container.Provide(New)
	container.Provide(NewMockedClientFactory)

	err := container.Invoke(func(m Manager) {
		t.manager = m
	})
	t.Require().NoError(err)
}
