/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package manifest

import (
	"io/ioutil"

	"github.com/hashicorp/hcl"
	pb "github.com/sniperkit/snk.fork.thrap/thrapb"
	"gopkg.in/yaml.v2"
)

// ParseYAML parses a manifest yaml file
func ParseYAML(file string) (mf *pb.Stack, err error) {
	var in []byte
	in, err = ioutil.ReadFile(file)
	if err == nil {
		mf, err = ParseYAMLBytes(in)
	}
	return
}

// ParseYAMLBytes reads a stack configuration into a stack struct
func ParseYAMLBytes(in []byte) (m *pb.Stack, err error) {
	var stack pb.Stack
	err = yaml.Unmarshal(in, &stack)
	if err == nil {
		m = &stack
	}
	return
}

// ParseHCL parses a hcl manifest file
func ParseHCL(file string) (mf *pb.Stack, err error) {
	var in []byte
	in, err = ioutil.ReadFile(file)
	if err == nil {
		mf, err = ParseHCLBytes(in)
	}
	return
}

// ParseHCLBytes parse a byte slice to a Stack struct
func ParseHCLBytes(in []byte) (*pb.Stack, error) {
	// wrapper map
	var ms map[string]map[string]*pb.Stack
	err := hcl.Unmarshal(in, &ms)
	if err != nil {
		return nil, err
	}

	var stack *pb.Stack
	for k, v := range ms["manifest"] {
		stack = v
		stack.ID = k
		break
	}
	return stack, nil
}
