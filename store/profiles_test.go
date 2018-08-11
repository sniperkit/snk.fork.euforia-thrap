/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Profiles(t *testing.T) {
	db, err := parseProfiles("../test-fixtures/profiles.hcl")
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, db.Profiles["local"])
	assert.Equal(t, "local", db.Default)
	assert.Equal(t, "docker", db.Profiles["local"].Orchestrator)
	assert.Equal(t, "docker", db.Profiles["local"].Registry)
}
