package storage

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"

	"github.com/ava-labs/avalanchego/database"
	"github.com/ava-labs/avalanchego/ids"
	smath "github.com/ava-labs/avalanchego/utils/math"
	"github.com/ava-labs/hypersdk/chain"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/crypto"
	"github.com/bbehrman10/energyavavm/utils"
)

type ReadState func(context.Context, [][]byte) ([][]byte, []error)

const (
	txPrefix = 0x0

	balancePrefix      = 0x1
	assetPrefix        = 0x2
	energyOrderPrefix  = 0x3
	creditPrefix       = 0x4
	heightPrefix       = 0x5
	incomingWarpPrefix = 0x6
	outgoingWarpPrefix = 0x7
)

var (
	failureByte = byte(0x0)
	successByte = byte(0x1)
	heightKey   = []byte{heightPrefix}

	balancePrefixPool = sync.Pool{
		New: func() any {
			return make([]byte, 1+crypto.PublicKeyLen+consts.IDLen)
		},
	}
)

func PrefixTxKey(id ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen)
	k[0] = txPrefix
	copy(k[1:], id[:])
	return
}

func StoreTransaction(
	_ context.Context,
	db database.KeyValueWriter,
	id ids.ID,
	t int64,
	success bool,
	units uint64,
) error {
	k := PrefixTxKey(id)
	v := make([]byte, consts.Uint64Len+1+consts.Uint64Len)
	binary.BigEndian.PutUint64(v, uint64(t))
	if success {
		v[consts.Uint64Len] = successByte
	} else {
		v[consts.Uint64Len] = failureByte
	}
	binary.BigEndian.PutUint64((v[consts.Uint64Len+1:]), units)
	return db.Put(k, v)
}

func GetTransaction(
	_ context.Context,
	db database.KeyValueReader,
	id ids.ID,
) (bool, int64, bool, uint64, error) {
	k := PrefixTxKey(id)
	v, err := db.Get(k)
	if errors.Is(err, database.ErrNotFound) {
		return false, 0, false, 0, nil
	}
	if err != nil {
		return false, 0, false, 0, err
	}
	t := int64(binary.BigEndian.Uint64(v))
	success := true
	if v[consts.Uint64Len] == failureByte {
		success = false
	}
	units := binary.BigEndian.Uint64(v[consts.Uint64Len+1:])
	return true, t, success, units, nil
}

func PrefixBalanceKey(pk crypto.PublicKey, asset ids.ID) (k []byte) {
	k = balancePrefixPool.Get().([]byte)
	k[0] = balancePrefix
	copy(k[1:], pk[:])
	copy(k[1+crypto.PublicKeyLen:], asset[:])
	return
}

func GetBalance(
	ctx context.Context,
	db chain.Database,
	pk crypto.PublicKey,
	asset ids.ID,
) (uint64, error) {
	dbKey, bal, err := getBalance(ctx, db, pk, asset)
	balancePrefixPool.Put(dbKey)
	return bal, err
}

func getBalance(
	ctx context.Context,
	db chain.Database,
	pk crypto.PublicKey,
	asset ids.ID,
) ([]byte, uint64, error) {
	k := PrefixBalanceKey(pk, asset)
	bal, err := innerGetBalance(db.GetValue(ctx, k))
	return k, bal, err
}

// Used to serve RPC queries
func GetBalanceFromState(
	ctx context.Context,
	f ReadState,
	pk crypto.PublicKey,
	asset ids.ID,
) (uint64, error) {
	k := PrefixBalanceKey(pk, asset)
	values, errs := f(ctx, [][]byte{k})
	bal, err := innerGetBalance(values[0], errs[0])
	balancePrefixPool.Put(k)
	return bal, err
}

func innerGetBalance(
	v []byte,
	err error,
) (uint64, error) {
	if errors.Is(err, database.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(v), nil
}

func SetBalance(
	ctx context.Context,
	db chain.Database,
	pk crypto.PublicKey,
	asset ids.ID,
	balance uint64,
) error {
	k := PrefixBalanceKey(pk, asset)
	return setBalance(ctx, db, k, balance)
}

func setBalance(
	ctx context.Context,
	db chain.Database,
	dbKey []byte,
	balance uint64,
) error {
	return db.Insert(ctx, dbKey, binary.BigEndian.AppendUint64(nil, balance))
}

func DeleteBalance(
	ctx context.Context,
	db chain.Database,
	pk crypto.PublicKey,
	asset ids.ID,
) error {
	return db.Remove(ctx, PrefixBalanceKey(pk, asset))
}

func AddBalance(
	ctx context.Context,
	db chain.Database,
	pk crypto.PublicKey,
	asset ids.ID,
	amount uint64,
) error {
	dbKey, bal, err := getBalance(ctx, db, pk, asset)
	if err != nil {
		return err
	}
	nbal, err := smath.Add64(bal, amount)
	if err != nil {
		return fmt.Errorf(
			"%w: could not add to balance: asset=%s, balance=%d, address=%v, amount=%d",
			ErrInvalidBalance,
			asset,
			bal,
			utils.Address(pk),
			amount,
		)
	}
	return setBalance(ctx, db, dbKey, nbal)
}

func SubBalance(
	ctx context.Context,
	db chain.Database,
	pk crypto.PublicKey,
	asset ids.ID,
	amount uint64,
) error {
	dbKey, bal, err := getBalance(ctx, db, pk, asset)
	if err != nil {
		return err
	}
	nbal, err := smath.Sub(bal, amount)
	if err != nil {
		return fmt.Errorf(
			"%w: could not subtract from balance: asset=%s, balance=%d, address=%v, amount=%d",
			ErrInvalidBalance,
			asset,
			bal,
			utils.Address(pk),
			amount,
		)
	}
	if nbal == 0 {
		return db.Remove(ctx, dbKey)
	}
	return setBalance(ctx, db, dbKey, nbal)
}

func PrefixAssetKey(asset ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen)
	k[0] = assetPrefix
	copy(k[1:], asset[:])
	return
}

func GetAssetFromState(
	ctx context.Context,
	f ReadState,
	asset ids.ID,
) (bool, []byte, uint64, crypto.PublicKey, bool, error) {
	values, errs := f(ctx, [][]byte{PrefixAssetKey(asset)})
	return innerGetAsset(values[0], errs[0])
}

func GetAsset(
	ctx context.Context,
	db chain.Database,
	asset ids.ID,
) (bool, []byte, uint64, crypto.PublicKey, bool, error) {
	k := PrefixAssetKey(asset)
	return innerGetAsset(db.GetValue(ctx, k))
}

func innerGetAsset(
	v []byte,
	err error,
) (bool, []byte, uint64, crypto.PublicKey, bool, error) {
	if errors.Is(err, database.ErrNotFound) {
		return false, nil, 0, crypto.EmptyPublicKey, false, nil
	}
	if err != nil {
		return false, nil, 0, crypto.EmptyPublicKey, false, err
	}
	metadataLen := binary.BigEndian.Uint16(v)
	metadata := v[consts.Uint16Len : consts.Uint16Len+metadataLen]
	supply := binary.BigEndian.Uint64(v[consts.Uint16Len+metadataLen:])
	var pk crypto.PublicKey
	copy(pk[:], v[consts.Uint16Len+metadataLen+consts.Uint64Len:])
	warp := v[consts.Uint16Len+metadataLen+consts.Uint64Len+crypto.PublicKeyLen] == 0x1
	return true, metadata, supply, pk, warp, nil
}

func SetAsset(
	ctx context.Context,
	db chain.Database,
	asset ids.ID,
	metadata []byte,
	supply uint64,
	owner crypto.PublicKey,
	warp bool,
) error {
	k := PrefixAssetKey(asset)
	metadataLen := len(metadata)
	v := make([]byte, consts.Uint16Len+metadataLen+consts.Uint64Len+crypto.PublicKeyLen+1)
	binary.BigEndian.PutUint16(v, uint16(metadataLen))
	copy(v[consts.Uint16Len:], metadata)
	binary.BigEndian.PutUint64(v[consts.Uint16Len+metadataLen:], supply)
	copy(v[consts.Uint16Len+metadataLen+consts.Uint64Len:], owner[:])
	b := byte(0x0)
	if warp {
		b = 0x1
	}
	v[consts.Uint16Len+metadataLen+consts.Uint64Len+crypto.PublicKeyLen] = b
	return db.Insert(ctx, k, v)
}

func DeleteAsset(ctx context.Context, db chain.Database, asset ids.ID) error {
	return db.Remove(ctx, PrefixAssetKey(asset))
}

func PrefixEnergyOrderKey(txID ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen)
	k[0] = energyOrderPrefix
	copy(k[1:], txID[:])
	return
}

func SetEnergyOrder(
	ctx context.Context,
	db chain.Database,
	tdID ids.ID,
	in ids.ID,
	inTick uint64,
	out ids.ID,
	outTick uint64,
	supply uint64,
	owner crypto.PublicKey,
) error {
	k := PrefixEnergyOrderKey(tdID)
	v := make([]byte, consts.IDLen*2+consts.Uint64Len*3+crypto.PublicKeyLen)
	copy(v, in[:])
	binary.BigEndian.PutUint64(v[consts.IDLen:], inTick)
	copy(v[consts.IDLen+consts.Uint64Len:], out[:])
	binary.BigEndian.PutUint64(v[consts.IDLen*2+consts.Uint64Len:], outTick)
	binary.BigEndian.PutUint64(v[consts.IDLen*2+consts.Uint64Len*2:], supply)
	copy(v[consts.IDLen*2+consts.Uint64Len*3:], owner[:])
	return db.Insert(ctx, k, v)
}

func GetEnergyOrder(
	ctx context.Context,
	db chain.Database,
	order ids.ID,
) (
	bool,
	ids.ID,
	uint64,
	ids.ID,
	uint64,
	uint64,
	crypto.PublicKey,
	error,
) {
	k := PrefixEnergyOrderKey(order)
	v, err := db.GetValue(ctx, k)
	if errors.Is(err, database.ErrNotFound) {
		return false, ids.Empty, 0, ids.Empty, 0, 0, crypto.EmptyPublicKey, nil
	}
	if err != nil {
		return false, ids.Empty, 0, ids.Empty, 0, 0, crypto.EmptyPublicKey, err
	}
	var in ids.ID
	copy(in[:], v[:consts.IDLen])
	inTick := binary.BigEndian.Uint64(v[consts.IDLen:])
	var out ids.ID
	copy(out[:], v[consts.IDLen+consts.Uint64Len:consts.IDLen*2+consts.Uint64Len])
	outTick := binary.BigEndian.Uint64(v[consts.IDLen*2+consts.Uint64Len:])
	supply := binary.BigEndian.Uint64(v[consts.IDLen*2+consts.Uint64Len*2:])
	var owner crypto.PublicKey
	copy(owner[:], v[consts.IDLen*2+consts.Uint64Len*3:])
	return true, in, inTick, out, outTick, supply, owner, nil
}

func DeleteOrder(ctx context.Context, db chain.Database, order ids.ID) error {
	return db.Remove(ctx, PrefixEnergyOrderKey(order))
}

func CreditPrefixKey(asset ids.ID, destination ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen*2)
	k[0] = creditPrefix
	copy(k[1:], asset[:])
	copy(k[1+consts.IDLen:], destination[:])
	return
}

func GetCreditFromState(
	ctx context.Context,
	f ReadState,
	asset ids.ID,
	destination ids.ID,
) (uint64, error) {
	values, errs := f(ctx, [][]byte{CreditPrefixKey(asset, destination)})
	return innerGetCredit(values[0], errs[0])
}

func innerGetCredit(v []byte, err error) (uint64, error) {
	if errors.Is(err, database.ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(v), nil
}

func GetCredit(
	ctx context.Context,
	db chain.Database,
	asset ids.ID,
	destination ids.ID,
) (uint64, error) {
	k := CreditPrefixKey(asset, destination)
	v, err := db.GetValue(ctx, k)
	return innerGetCredit(v, err)
}

func SetCredit(
	ctx context.Context,
	db chain.Database,
	asset ids.ID,
	destination ids.ID,
	amount uint64,
) error {
	k := CreditPrefixKey(asset, destination)
	return db.Insert(ctx, k, binary.BigEndian.AppendUint64(nil, amount))
}

func AddCredit(
	ctx context.Context,
	db chain.Database,
	asset ids.ID,
	destination ids.ID,
	amount uint64,
) error {
	credit, err := GetCredit(ctx, db, asset, destination)
	if err != nil {
		return err
	}
	ncredit, err := smath.Add64(credit, amount)
	if err != nil {
		return fmt.Errorf(
			"%w: could not add to credit: asset=%s, destination=%s, amount=%d",
			ErrInvalidBalance,
			asset,
			destination,
			amount,
		)
	}
	if ncredit == 0 {
		return db.Remove(ctx, CreditPrefixKey(asset, destination))
	}
	return SetCredit(ctx, db, asset, destination, ncredit)
}

func SubCredit(
	ctx context.Context,
	db chain.Database,
	asset ids.ID,
	destination ids.ID,
	amount uint64,
) error {
	credit, err := GetCredit(ctx, db, asset, destination)
	if err != nil {
		return err
	}
	ncredit, err := smath.Sub(credit, amount)
	if err != nil {
		return fmt.Errorf(
			"%w: could not subtract from credit: asset=%s, destination=%s, amount=%d",
			ErrInvalidBalance,
			asset,
			destination,
			amount,
		)
	}
	if ncredit == 0 {
		return db.Remove(ctx, CreditPrefixKey(asset, destination))
	}
	return SetCredit(ctx, db, asset, destination, ncredit)
}

func HeightKey() (k []byte) {
	return heightKey
}

func IncomingWarpKeyPrefix(sourceChainID ids.ID, msgID ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen*2)
	k[0] = incomingWarpPrefix
	copy(k[1:], sourceChainID[:])
	copy(k[1+consts.IDLen:], msgID[:])
	return k
}

func OutgoingWarpPrefix(txID ids.ID) (k []byte) {
	k = make([]byte, 1+consts.IDLen)
	k[0] = outgoingWarpPrefix
	copy(k[1:], txID[:])
	return k
}
