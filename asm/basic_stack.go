/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package asm

import (
	"strings"

	"github.com/sniperkit/snk.fork.thrap/consts"
	"github.com/sniperkit/snk.fork.thrap/packs"
	"github.com/sniperkit/snk.fork.thrap/thrapb"
	"github.com/sniperkit/snk.fork.thrap/vars"
)

// BasicStackConfig holds a configurations to build canned stacks.  Currently
// supports a dev, datastore and web component
type BasicStackConfig struct {
	Name      string
	Language  thrapb.LanguageID
	DataStore string
	WebServer string
}

// NewBasicStack builds a skeleton stack inferring as much information as
// possible to assemble the stack.  It only inits the object with the minimal
// valid defaults
func NewBasicStack(c *BasicStackConfig, pks *packs.Packs) (*thrapb.Stack, error) {

	stack := thrapb.Stack{
		ID:   c.Name,
		Name: c.Name,
	}

	devComp := makeDevComp(consts.DefaultAPICompID, c.Language)
	comps := map[string]*thrapb.Component{
		consts.DefaultAPICompID: devComp,
	}

	var dsComp *thrapb.Component
	if c.DataStore != "none" {
		dspacks := pks.Datastore()
		ds, err := dspacks.Load(c.DataStore)
		if err != nil {
			return nil, err
		}

		dsComp = thrapb.NewComponent(ds.Image, ds.DefaultVersion, thrapb.CompTypeDatastore)
		comps[consts.DefaultDSCompID] = dsComp

		// Add env var for host, port and addr to dev component
		ev := defaultCompEnvVars(consts.DefaultDSCompID)
		for k, v := range ev {
			devComp.Env.Vars[k] = v
		}
	}

	var wsComp *thrapb.Component
	if c.WebServer != "none" {
		wpacks := pks.Web()
		ws, err := wpacks.Load(c.WebServer)
		if err != nil {
			return nil, err
		}
		wsComp = thrapb.NewComponent(ws.Image, ws.DefaultVersion, thrapb.CompTypeWeb)
		comps[consts.DefaultWebCompID] = wsComp

		// Add env var for host, port and addr of dev to web component (the just created)
		wsComp.Env = &thrapb.Envionment{
			Vars: defaultCompEnvVars(consts.DefaultAPICompID),
		}
	}

	stack.Components = comps

	return &stack, nil
}

func defaultCompEnvVars(pre string) map[string]string {
	upre := strings.ToUpper(pre)
	return map[string]string{
		upre + "_CONTAINER_IP": "${" + consts.CompVarPrefixKey + "." + pre + ".container.ip}",
	}
}

func makeDevComp(compID string, lang thrapb.LanguageID) *thrapb.Component {
	return &thrapb.Component{
		ID: compID,
		// registry info and parent are dynamically prepended based on profile
		Name:     compID,
		Type:     thrapb.CompTypeAPI,
		Language: lang,
		Head:     true,
		Build: &thrapb.Build{
			Dockerfile: compID + "." + consts.DefaultDockerFile,
		},
		Env: &thrapb.Envionment{
			File: consts.DefaultEnvFile,
			Vars: map[string]string{
				consts.EnvVarVersion: "${" + vars.StackVersion + "}",
			},
		},
		Secrets: &thrapb.Secrets{
			Destination: consts.DefaultSecretsFile,
		},
	}
}
