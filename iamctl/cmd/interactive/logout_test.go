package interactive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogout(t *testing.T) {
	t.Run("Test whether the attributes of logout command are correctly parsed", func(t *testing.T) {
		cmd := logoutCmd
		assert.Equal(t, "logout", cmd.Use)
		assert.Contains(t, cmd.Short, "Logout and disconnect from the Server")
		assert.Contains(t, cmd.Long, "Logout and disconnect from the Server by removing the stored credentials.")
		assert.NotNil(t, cmd.Run)
	})
}
