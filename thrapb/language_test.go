/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package thrapb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_LanguageID(t *testing.T) {
	lang := LanguageID("")
	err := lang.Validate()
	assert.Equal(t, errLangNotSpecified, err)

	lang = "foo"
	err = lang.Validate()
	assert.Nil(t, err)
	assert.Equal(t, "foo", lang.Lang())

	lang = "foo:1.2"
	assert.Equal(t, "foo", lang.Lang())
	assert.Equal(t, "1.2", lang.Version())

	lang = LanguageID("go:bs")
	err = lang.Validate()
	assert.NotNil(t, err)
}
