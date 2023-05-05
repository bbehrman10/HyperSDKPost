package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/bbehrman10/energyavavm/auth"
	"github.com/bbehrman10/energyavavm/storage"
)

var _ chain.Action = (*InitializeEnergyAsset)(nil)

type InitializeEnergyAsset struct {
	Metadata []byte `json:"metadata"`
}

func (*InitializeEnergyAsset) StateKeys(_ chain.Auth, txID ids.ID) [][]byte {
	return [][]byte{storage.PrefixAssetKey(txID)}
}

func (c *InitializeEnergyAsset) Execute(
	ctx context.Context,
	r chain.Rules,
	db chain.Database,
	_ int64,
	rauth chain.Auth,
	txID ids.ID,
	_ bool,
) (*chain.Result, error) {
	actor := auth.GetActor(rauth)
	unitsUsed := c.MaxUnits(r)
	if len(c.Metadata) > MaxMetadataSize {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputMetadataTooLarge}, nil
	}
	if err := storage.SetAsset(ctx, db, txID, c.Metadata, 0, actor, false); err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	return &chain.Result{Success: true, Units: unitsUsed}, nil
}

func (c *InitializeEnergyAsset) MaxUnits(chain.Rules) uint64 {
	return uint64(len(c.Metadata))
}

func (c *InitializeEnergyAsset) Marshal(p *codec.Packer) {
	p.PackBytes(c.Metadata)
}

func UnmarshalCreateAsset(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var create InitializeEnergyAsset
	p.UnpackBytes(MaxMetadataSize, false, &create.Metadata)
	return &create, p.Err()
}

func (*InitializeEnergyAsset) ValidRange(chain.Rules) (int64, int64) {
	return -1, -1
}
