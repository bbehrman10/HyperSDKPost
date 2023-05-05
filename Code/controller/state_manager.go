package controller

import (
	"github.com/ava-labs/avalanchego/ids"
	"github.com/bbehrman10/energyavavm/storage"
)

type StateManager struct{}

func (*StateManager) HeightKey() []byte {
	return storage.HeightKey()
}

func (*StateManager) IncomingWarpKey(sourceChainID ids.ID, msgID ids.ID) []byte {
	return storage.IncomingWarpKeyPrefix(sourceChainID, msgID)
}

func (*StateManager) OutgoingWarpKey(txID ids.ID) []byte {
	return storage.OutgoingWarpPrefix(txID)
}
