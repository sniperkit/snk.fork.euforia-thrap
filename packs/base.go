/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package packs

import (
	"io/ioutil"
	"path/filepath"

	"github.com/hashicorp/hcl"
	"github.com/sniperkit/snk.fork.thrap/thrapb"
)

type BasePack struct {
	dir string
	*thrapb.PackManifest
}

type BasePacks struct {
	*basePackSet
	packs map[string]*BasePack
}

func NewBasePacks(dir string) *BasePacks {
	return &BasePacks{
		basePackSet: &basePackSet{"web", dir},
		packs:       make(map[string]*BasePack),
	}
}

func (packs *BasePacks) Load(packID string) (*BasePack, error) {
	if val, ok := packs.packs[packID]; ok {
		return val, nil
	}
	pack, err := LoadBasePack(packID, packs.dir)
	if err == nil {
		packs.packs[packID] = pack
	}
	return pack, err
}

func LoadBasePack(packID, dir string) (*BasePack, error) {
	pdir := filepath.Join(dir, packID)
	b, err := ioutil.ReadFile(filepath.Join(pdir, packManfiestFile))
	if err != nil {
		return nil, err
	}

	wp := &BasePack{dir: pdir}
	var conf thrapb.PackManifest
	err = hcl.Unmarshal(b, &conf)
	if err == nil {
		wp.PackManifest = &conf
	}
	return wp, err
}
