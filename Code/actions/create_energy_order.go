package actions

import (
	"context"
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/warp"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/utils"
	"github.com/bbehrman10/energyavavm/auth"
	"github.com/bbehrman10/energyavavm/storage"
)

var _ chain.Action = (*CreateEnergyOrder)(nil)

type CreateEnergyOrder struct {
	// In represents the energy in kilowatt-hours (kWh) the seller wants to sell.
	In ids.ID `json:"in"`

	// InTick represents the amount of energy in kWh per unit of currency.
	InTick uint64 `json:"inTick"`

	// Out represents the currency the seller accepts in exchange for energy.
	Out ids.ID `json:"out"`

	// OutTick represents the amount of currency the seller receives per unit of energy.
	OutTick uint64 `json:"outTick"`

	// Supply represents the total amount of energy in kWh the seller is willing to sell.
	Supply uint64 `json:"supply"`
}

func (c *CreateEnergyOrder) StateKeys(rauth chain.Auth, txID ids.ID) [][]byte {
	actor := auth.GetActor(rauth)
	return [][]byte{
		storage.PrefixBalanceKey(actor, c.Out),
		storage.PrefixEnergyOrderKey(txID),
	}
}

func (c *CreateEnergyOrder) Execute(
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
	if c.In == c.Out {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputSameInOut}, nil
	}
	if c.InTick == 0 {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputInTickZero}, nil
	}
	if c.OutTick == 0 {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputOutTickZero}, nil
	}
	if c.Supply == 0 {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputSupplyZero}, nil
	}
	if c.Supply%c.OutTick != 0 {
		return &chain.Result{Success: false, Units: unitsUsed, Output: OutputSupplyMisaligned}, nil
	}
	if err := storage.SubBalance(ctx, db, actor, c.Out, c.Supply); err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	if err := storage.SetEnergyOrder(ctx, db, txID, c.In, c.InTick, c.Out, c.OutTick, c.Supply, actor); err != nil {
		return &chain.Result{Success: false, Units: unitsUsed, Output: utils.ErrBytes(err)}, nil
	}
	return &chain.Result{Success: true, Units: unitsUsed}, nil
}

func (*CreateEnergyOrder) MaxUnits(chain.Rules) uint64 {
	return consts.IDLen*2 + consts.Uint64Len*3
}

func (c *CreateEnergyOrder) Marshal(p *codec.Packer) {
	p.PackID(c.In)
	p.PackUint64(c.InTick)
	p.PackID(c.Out)
	p.PackUint64(c.OutTick)
	p.PackUint64(c.Supply)
}

func UnmarshalCreateEnergyOrder(p *codec.Packer, _ *warp.Message) (chain.Action, error) {
	var create CreateEnergyOrder
	p.UnpackID(false, &create.In)
	create.InTick = p.UnpackUint64(true)
	p.UnpackID(false, &create.Out)
	create.OutTick = p.UnpackUint64(true)
	create.Supply = p.UnpackUint64(true)
	return &create, p.Err()
}

func (*CreateEnergyOrder) ValidRange(chain.Rules) (int64, int64) {
	return -1, -1
}

func PairID(in ids.ID, out ids.ID) string {
	return fmt.Sprintf("%s-%s", in.String(), out.String())
}
