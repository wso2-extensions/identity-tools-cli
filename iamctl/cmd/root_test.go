package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	t.Run("Test whether the attributes of the root command are correctly parsed", func(t *testing.T) {
		cmd := RootCmd
		assert.Equal(t, "iamctl", cmd.Use)
		assert.Contains(t, cmd.Short, "IAM Command Line Interface (CLI) Tool")
		assert.Contains(t, cmd.Long, "IAM Command Line Interface (CLI) Tool to manage and configure WSO2 Identity Server and Asgardeo.")
		assert.NotNil(t, cmd.Run)
	})

	t.Run("Thest whether the ub commands are added correctly", func(t *testing.T) {
		cmds := RootCmd.Commands()
		assert.Greater(t, len(cmds), 0, "Expected at least one sub-command to be present")
		for _, subCmd := range cmds {
			assert.NotNil(t, subCmd.Use, "Sub-command 'Use' should not be nil")
		}
	})
}
