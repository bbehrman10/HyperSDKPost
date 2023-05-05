package actions

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/bbehrman10/energyavavm/auth"
	"github.com/bbehrman10/energyavavm/storage"
)

var _ chain.Action = (*CloseEnergyOrder)(nil)

type CloseEnergyOrder struct {
	// [Order] is the OrderID you wish to close.
	Order ids.ID `json:"order"`

	// [Out] is the asset locked up in the order. We need to provide this to
	// populate [StateKeys].
	Out ids.ID `json:"out"`
}

func (c *CloseEnergyOrder) StateKeys(rauth chain.Auth, _ ids.ID) [][]byte {
	actor := auth.GetActor(rauth)
	return [][]byte{
		storage.PrefixEnergyOrderKey(c.Order),
		storage.PrefixBalanceKey(actor, c.Out),
	}
}

func (c *CloseEnergyOrder) Execute(
	ctx context.Context,
	r chain.Rules,
	db chain.Database,
	_ int64,
	rauth chain.Auth,
	_ ids.ID,
	_ bool,
) (*chain.Result, error) {
	actor := auth.GetActor(rauth)
	unitsUsed := c.MaxUnits(r) // max units == units
	exists, _, _, out, _, remaining, owner, err := storage.GetEnergyOrder(ctx, db, c.Order)
	if err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	if !exists {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputOrderMissing}, nil
	}
	if owner != actor {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputUnauthorized}, nil
	}
	if out != c.Out {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputWrongOut}, nil
	}
	if err := storage.DeleteOrder(ctx, db, c.Order); err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	if err := storage.AddBalance(ctx, db, actor, c.Out, remaining); err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	return &chain.Result{Success: true, Units: unitsUsed}, nil
}

func (*CloseEnergyOrder) MaxUnits(chain.Rules) uint64 {
	return consts.IDLen * 2
}

func (c *CloseEnergyOrder) Marshal(p *codec.Packer) {
	p.PackID(c.Order)
	p.PackID(c.Out)
}

func UnmarshalCloseOrder(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var cl CloseEnergyOrder
	p.UnpackID(true, &cl.Order)
	p.UnpackID(false, &cl.Out) // empty ID is the native asset
	return &cl, p.Err()
}

func (*CloseEnergyOrder) ValidRange(chain.Rules) (int64, int64) {
	return -1, -1
}
