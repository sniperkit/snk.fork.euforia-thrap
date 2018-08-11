/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package dockerfile

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/sniperkit/snk.fork.thrap/utils"
)

// DockerIgnoresFile is the docker ignore filename
const DockerIgnoresFile = ".dockerignore"

// ParseIgnoresFile reads and parses the ignores file from the directory
func ParseIgnoresFile(dir string) ([]string, error) {
	fpath := filepath.Join(dir, DockerIgnoresFile)
	if !utils.FileExists(fpath) {
		return []string{}, nil
	}

	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(b), "\n"), nil
}
