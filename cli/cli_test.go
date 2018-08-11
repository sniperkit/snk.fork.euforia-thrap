/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_CLI(t *testing.T) {
	app := NewCLI("version")
	err := app.Run([]string{"thrap", "version"})
	assert.Nil(t, err)
}
