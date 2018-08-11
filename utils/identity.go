/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package utils

import (
	"io/ioutil"

	"github.com/euforia/hclencoder"
	"github.com/hashicorp/hcl"
	"github.com/sniperkit/snk.fork.thrap/thrapb"
)

func LoadIdentities(filename string) (map[string]*thrapb.Identity, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var out map[string]*thrapb.Identity
	err = hcl.Unmarshal(b, &out)

	return out, err
}

func WriteIdentities(filename string, idents map[string]*thrapb.Identity) error {
	b, err := hclencoder.Encode(idents)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, b, 0644)
}
