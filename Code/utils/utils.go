package utils

import (
	"github.com/ava-labs/hypersdk/crypto"

	"github.com/bbehrman10/energyavavm/consts"
)

func Address(pk crypto.PublicKey) string {
	return crypto.Address(consts.HRP, pk)
}

func ParseAddress(s string) (crypto.PublicKey, error) {
	return crypto.ParseAddress(consts.HRP, s)
}
