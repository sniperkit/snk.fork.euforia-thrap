/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetLocalPath(t *testing.T) {
	cwd, _ := os.Getwd()

	p, err := GetLocalPath("")
	assert.Nil(t, err)
	assert.Equal(t, cwd, p)

	p, _ = GetLocalPath("utils/foobar")
	assert.Equal(t, cwd+"/utils/foobar", p)
}

func Test_GetAbsPath(t *testing.T) {
	cwd, _ := os.Getwd()

	p, _ := GetAbsPath("")
	assert.Equal(t, cwd, p)

	p, _ = GetAbsPath("foo")
	assert.Equal(t, filepath.Join(cwd, "foo"), p)

	p, _ = GetAbsPath("~/bar")
	assert.True(t, strings.HasPrefix(p, "/"))
}
