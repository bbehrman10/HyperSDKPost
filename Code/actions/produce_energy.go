package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	smath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/bbehrman10/energyavavm/auth"
	"github.com/bbehrman10/energyavavm/storage"
)

var _ chain.Action = (*ProduceEnergy)(nil)

type ProduceEnergy struct {
	// To is the recipient of the [Value].
	To crypto.PublicKey `json:"to"`

	// Asset is the [TxID] that was used when the asset was initialized
	// Eventually, this will be dynamically provided from chain initialization
	Asset ids.ID `json:"asset"`

	// Number of KiloWattHours to produce
	Value uint64 `json:"value"`
}

func (m *ProduceEnergy) StateKeys(chain.Auth, ids.ID) [][]byte {
	return [][]byte{
		storage.PrefixAssetKey(m.Asset),
		storage.PrefixBalanceKey(m.To, m.Asset),
	}
}

func (m *ProduceEnergy) Execute(
	ctx context.Context,
	r chain.Rules,
	db chain.Database,
	_ int64,
	rauth chain.Auth,
	_ ids.ID,
	_ bool,
) (*chain.Result, error) {
	actor := auth.GetActor(rauth)
	unitsUsed := m.MaxUnits(r)
	if m.Asset == ids.Empty {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputAssetIsNative}, nil
	}
	if m.Value == 0 {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputValueZero}, nil
	}
	exists, metadata, supply, owner, isWarp, err := storage.GetAsset(ctx, db, m.Asset)
	if err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	if !exists {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputAssetMissing}, nil
	}
	if isWarp {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputWarpAsset}, nil
	}
	if owner != actor {
		return &chain.Result{
			Success: false,
			Units:   unitsUsed,
			Output:  OutputWrongOwner,
		}, nil
	}
	newSupply, err := smath.Add64(supply, m.Value)
	if err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	if err := storage.SetAsset(ctx, db, m.Asset, metadata, newSupply, actor, isWarp); err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	if err := storage.AddBalance(ctx, db, m.To, m.Asset, m.Value); err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	return &chain.Result{Success: true, Units: unitsUsed}, nil
}

func (*ProduceEnergy) MaxUnits(chain.Rules) uint64 {
	return crypto.PublicKeyLen + consts.IDLen + consts.Uint64Len
}

func (m *ProduceEnergy) Marshal(p *codec.Packer) {
	p.PackPublicKey(m.To)
	p.PackID(m.Asset)
	p.PackUint64(m.Value)
}

func UnmarshalProduceEnergy(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var produce ProduceEnergy
	p.UnpackPublicKey(true, &produce.To) // cannot produce to nothing
	p.UnpackID(true, &produce.Asset)     // empty ID is the native asset
	produce.Value = p.UnpackUint64(true)
	return &produce, p.Err()
}

func (*ProduceEnergy) ValidRange(chain.Rules) (int64, int64) {
	return -1, -1
}
