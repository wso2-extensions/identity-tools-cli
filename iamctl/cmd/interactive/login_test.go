package interactive

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	t.Run("Test whether the attributes of the commands are correctly parsed", func(t *testing.T) {
		cmd := loginCmd
		assert.Equal(t, "login", cmd.Use)
		assert.Contains(t, cmd.Short, "Login and connect with the Server")
		assert.Contains(t, cmd.Long, fmt.Sprintf(`Login and connect with the Server using Client ID, Client Secret and Organization Name (Asgardeo) or Tenant Domain (Identity Server). 
You will be asked to select the server type (Asgardeo/Identity Server) and provide the required details to login. 
You can provide the Client ID and Organization Name (Asgardeo) or Tenant Domain (Identity Server) as flags, or you will be prompted to enter them interactively. 
You can also provide the Client Secret as a flag. If not provided, you will be prompted to enter it securely. 
We recommend using flags for non-interactive usage (Automation) and secure prompts for interactive usage.`))
		assert.NotNil(t, cmd.Run)
	})

	t.Run("Test whether flags are correctly defined", func(t *testing.T) {
		cmd := loginCmd
		flags := cmd.Flags()

		assert.NotNil(t, flags.Lookup("client-id"))
		assert.Equal(t, "Client ID of the M2M application", flags.Lookup("client-id").Usage)

		assert.NotNil(t, flags.Lookup("org"))
		assert.Equal(t, "Name of the Organization (Asgardeo) or Tenant Domain (Identity Server)", flags.Lookup("org").Usage)

		assert.NotNil(t, flags.Lookup("client-secret"))
		assert.Equal(t, "Client Secret of the M2M application", flags.Lookup("client-secret").Usage)

		assert.NotNil(t, flags.Lookup("server-type"))
		assert.Equal(t, "Type of the server to connect to (Asgardeo/Identity Server)", flags.Lookup("server-type").Usage)

		assert.NotNil(t, flags.Lookup("identity-server-url"))
		assert.Equal(t, "URL of the Identity Server in the format [<protocol>://<host>] (Example https://localhost:9443)", flags.Lookup("identity-server-url").Usage)
	})

}
