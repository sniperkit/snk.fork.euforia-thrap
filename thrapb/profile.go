/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package thrapb

// DefaultProfile returns the default local profile
func DefaultProfile() *Profile {
	return &Profile{
		ID:           "local",
		Orchestrator: "docker",
		Registry:     "docker",
	}
}
