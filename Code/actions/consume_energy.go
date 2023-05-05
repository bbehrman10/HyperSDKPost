package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	smath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"

	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/bbehrman10/energyavavm/auth"
	"github.com/bbehrman10/energyavavm/storage"
)

var _ chain.Action = (*ConsumeEnergy)(nil)

type ConsumeEnergy struct {
	// Asset is the [TxID] that created the asset.
	// Like with producing, this will eventually be dynamically provided
	Asset ids.ID `json:"asset"`

	// number of kilowatt hours to consume
	Value uint64 `json:"value"`
}

func (b *ConsumeEnergy) StateKeys(rauth chain.Auth, _ ids.ID) [][]byte {
	actor := auth.GetActor(rauth)
	return [][]byte{
		storage.PrefixAssetKey(b.Asset),
		storage.PrefixBalanceKey(actor, b.Asset),
	}
}

func (b *ConsumeEnergy) Execute(
	ctx context.Context,
	r chain.Rules,
	db chain.Database,
	_ int64,
	rauth chain.Auth,
	_ ids.ID,
	_ bool,
) (*chain.Result, error) {
	actor := auth.GetActor(rauth)
	unitsUsed := b.MaxUnits(r)
	if b.Value == 0 {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputValueZero}, nil
	}
	if err := storage.SubBalance(ctx, db, actor, b.Asset, b.Value); err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	exists, metadata, supply, owner, warp, err := storage.GetAsset(ctx, db, b.Asset)
	if err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	if !exists {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputAssetMissing}, nil
	}
	newSupply, err := smath.Sub(supply, b.Value)
	if err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	if err := storage.SetAsset(ctx, db, b.Asset, metadata, newSupply, owner, warp); err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	return &chain.Result{Success: true, Units: unitsUsed}, nil
}

func (*ConsumeEnergy) MaxUnits(chain.Rules) uint64 {
	return consts.IDLen + consts.Uint64Len
}

func (b *ConsumeEnergy) Marshal(p *codec.Packer) {
	p.PackID(b.Asset)
	p.PackUint64(b.Value)
}

func UnmarshalConsumeEnergy(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var consume ConsumeEnergy
	p.UnpackID(false, &consume.Asset)
	consume.Value = p.UnpackUint64(true)
	return &consume, p.Err()
}

func (*ConsumeEnergy) ValidRange(chain.Rules) (int64, int64) {
	return -1, -1
}
