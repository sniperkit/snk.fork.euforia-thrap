/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package utils

import (
	"errors"
)

// FlattenErrors flattens a map of errors into a new line delimited
// string and returns tha single error
func FlattenErrors(errs map[string]error) error {
	var out string
	for k, v := range errs {
		out += k + ":" + v.Error() + "\n"
	}
	return errors.New(out)
}
