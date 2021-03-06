/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package vars

import (
	"github.com/euforia/pseudo/scope"
)

// MergeScopeVars merges the 2 variable sets. The last one take precedence
func MergeScopeVars(base, add scope.Variables) scope.Variables {
	if base == nil {
		return add
	} else if add == nil {
		return base
	}

	for k, v := range add {
		base[k] = v
	}
	return base
}
