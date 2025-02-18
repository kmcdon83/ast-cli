// +build !integration

package commands

import (
	"testing"

	"gotest.tools/assert"
)

func TestSastWebAppHealthCheck(t *testing.T) {
	cmd := createASTTestCommand()
	err := executeTestCommand(cmd, "utils", "health-check", "--role", "SAST_ALL_IN_ONE")
	assert.NilError(t, err)
}

func TestScaHealthCheck(t *testing.T) {
	cmd := createASTTestCommand()
	err := executeTestCommand(cmd, "utils", "health-check", "--role", "SCA_AGENT")
	assert.NilError(t, err)
}
