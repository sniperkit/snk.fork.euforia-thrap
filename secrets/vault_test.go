/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Secrets(t *testing.T) {
	conf := &Config{Provider: "vault", Conf: map[string]interface{}{"prefix": "/"}}
	sec, err := New(conf)
	assert.Nil(t, err)
	s := sec.(*vaultSecrets)
	assert.NotNil(t, s.client)

	conf.Provider = "foo"
	_, err = New(conf)
	assert.NotNil(t, err)
}

func Test_vault(t *testing.T) {
	conf := &Config{Provider: "vault"}
	sec, _ := New(conf)
	err := sec.Init(map[string]interface{}{
		// "addr":   "http://localhost:8200",
		"prefix": "/thrap/db",
		"token":  "myroot",
	})
	assert.Nil(t, err)

	err = sec.Set(map[string]interface{}{"foo": "bar"})
	assert.Nil(t, err)

	kvs, err := sec.Get()
	assert.Nil(t, err)
	_, ok := kvs["foo"]
	assert.True(t, ok)
}
