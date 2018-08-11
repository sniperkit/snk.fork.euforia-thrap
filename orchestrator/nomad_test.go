/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/sniperkit/snk.fork.thrap/manifest"
	"github.com/stretchr/testify/assert"
)

func Test_nomad_env(t *testing.T) {
	conf := &Config{Provider: "nomad"}
	orch, err := New(conf)
	assert.Nil(t, err)

	norch := orch.(*nomadOrchestrator)
	assert.Equal(t, norch.client.Address(), os.Getenv("NOMAD_ADDR"))
}

func Test_nomad_dryrun(t *testing.T) {
	conf := &Config{Provider: "nomad", Conf: map[string]interface{}{
		"addr": os.Getenv("NOMAD_ADDR"),
		// "addr": "http://127.0.0.1:4646",
	}}
	orch, err := New(conf)
	assert.Nil(t, err)

	st, err := manifest.LoadManifest("../thrap.yml")
	if err != nil {
		t.Fatal(err)
	}
	st.Validate()

	ctx := context.Background()
	_, ijob, err := orch.Deploy(ctx, st, RequestOptions{Dryrun: true})
	if err != nil {
		t.Fatal(err)
	}

	// ijob.(*api.Job).Canonicalize()
	b, _ := json.MarshalIndent(ijob, "", "  ")
	fmt.Printf("%s\n", b)
}
