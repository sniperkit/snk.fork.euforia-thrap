/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package orchestrator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Orchestrator(t *testing.T) {
	conf := &Config{Provider: "nomad"}
	orch, err := New(conf)
	assert.Nil(t, err)
	o := orch.(*nomadOrchestrator)
	assert.NotNil(t, o.client)

	conf.Provider = "foo"
	_, err = New(conf)
	assert.Contains(t, err.Error(), "unsupported")
}
