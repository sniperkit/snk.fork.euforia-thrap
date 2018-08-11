/*
Sniperkit-Bot
- Date: 2018-08-11 22:25:29.898780201 +0200 CEST m=+0.118184110
- Status: analyzed
*/

package thrap

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"math/big"
)

func makePubKeyFromBytes(curve elliptic.Curve, pubkey []byte) *ecdsa.PublicKey {
	x := big.Int{}
	y := big.Int{}
	keyLen := len(pubkey)
	x.SetBytes(pubkey[:(keyLen / 2)])
	y.SetBytes(pubkey[(keyLen / 2):])

	return &ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
}
