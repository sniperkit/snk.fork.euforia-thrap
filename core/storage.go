/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package core

import (
	"github.com/sniperkit/snk.fork.thrap/thrapb"
)

// StackStorage is a stack storage interface
type StackStorage interface {
	Get(string) (*thrapb.Stack, error)
	Create(*thrapb.Stack) (*thrapb.Stack, error)
	Update(*thrapb.Stack) (*thrapb.Stack, error)
	Iter(string, func(*thrapb.Stack) error) error
}

// IdentityStorage is a identity storage interface
type IdentityStorage interface {
	// Get returns an identity be the given id
	Get(id string) (*thrapb.Identity, error)
	Create(*thrapb.Identity) (*thrapb.Identity, error)
	Update(*thrapb.Identity) (*thrapb.Identity, error)
	Iter(string, func(*thrapb.Identity) error) error
}
