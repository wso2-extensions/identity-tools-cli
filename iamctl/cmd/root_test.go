package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCommand(t *testing.T) {
	t.Run("Test whether the attributes of the root command are correctly parsed", func(t *testing.T) {
		cmd := RootCmd
		assert.Equal(t, "iamctl", cmd.Use)
		assert.Contains(t, cmd.Short, "A CLI tool to manage Identity and Access Management tasks")
		assert.Contains(t, cmd.Long, "A CLI tool to manage Identity and Access Management tasks for WSO2 Identity Server and Asgardeo.")
		assert.NotNil(t, cmd.Run)
	})
}
