//go:build mage
// +build mage

package main

import (
	"os"

	"github.com/aserto-dev/mage-loot/common"
	"github.com/aserto-dev/mage-loot/deps"
)

func init() {
	os.Setenv("GO_VERSION", "1.19")
	os.Setenv("DOCKER_BUILDKIT", "1")
}

// Lint runs linting for the entire project.
func Lint() error {
	return common.Lint()
}

func Deps() {
	deps.GetAllDeps()
}
