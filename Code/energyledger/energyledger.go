package energyledger

import (
	"sync"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/crypto"
	"github.com/ava-labs/hypersdk/heap"
	"go.uber.org/zap"
)

const (
	initialPairCapacity = 128
	allPairs            = "*"
)

type EnergyOrder struct {
	ID           ids.ID `json:"id"`
	Producer     string `json:"producer"`
	EnergyAmount uint64 `json:"energyAmount"`
	TokensPaid   uint64 `json:"tokensPaid"`
	Remaining    uint64 `json:"remaining"`

	producer crypto.PrivateKey
}

type EnergyLedger struct {
	c *Controller

	orders      map[string]*heap.Heap[*EnergyOrder, float64]
	orderToPair map[ids.ID]string
	l           sync.Mutex

	trackAll bool
}

func NewEnergyLedger(c *Controller, trackedPairs []string) *EnergyLedger {
	m := map[string]*heap.Heap[*EnergyOrder, float64]{}
	trackAll := false
	if len(trackedPairs) == 1 && trackedPairs[0] == allPairs {
		trackAll = true
		c.Logger().Info("tracking all energy ledgers")
	} else {
		for _, pair := range trackedPairs {
			//Must use a max heap to return best rates
			m[pair] = heap.New[*EnergyOrder, float64](initialPairCapacity, true)
			c.Logger().Info("tracking energy ledger", zap.String("pair", pair))
		}
	}
	return &EnergyLedger{
		c:           c,
		orders:      m,
		orderToPair: map[ids.ID]string{},
		trackAll:    trackAll,
	}
}

func (o *EnergyLedger) Add(pair string, order *EnergyOrder) {
	o.l.Lock()
	defer o.l.Unlock()
	h, ok := o.orders[pair]
	switch {
	case !ok && !o.trackAll:
		return
	case !ok && o.trackAll:
		o.c.Logger().Info("tracking energy ledger", zap.String("pair", pair))
		h = heap.New[*EnergyOrder, float64](initialPairCapacity, true)
		o.orders[pair] = h
		h.Push(&heap.Entry[*EnergyOrder, float64]{
			ID:    order.ID,
			Val:   float64(order.EnergyAmount) / float64(order.TokensPaid),
			Item:  order,
			Index: h.Len(),
		})
		o.orderToPair[order.ID] = pair
	}
}

func (o *EnergyLedger) Remove(id ids.ID) {
	o.l.Lock()
	defer o.l.Unlock()
	pair, ok := o.orderToPair[id]
	if !ok {
		return
	}
	delete(o.orderToPair, id)
	h, ok := o.orders[pair]
	if !ok {
		//should never happen
		return
	}
	entry, ok := h.Get(id) // 0(log 1)
	if !ok {
		//should never happen
		return
	}
	h.Remove(entry.Index) // 0(log n)
}

func (o *EnergyLedger) UpdateRemaining(id ids.ID, remaining uint64) {
	o.l.Lock()
	defer o.l.Unlock()
	pair, ok := o.orderToPair[id]
	if !ok {
		return
	}
	h, ok := o.orders[pair]
	if !ok {
		return
	}
	entry, ok := h.Get(id)
	if !ok {
		return
	}
	entry.Item.Remaining = remaining
}

func (o *EnergyLedger) Orders(pair string, limit int) []*EnergyOrder {
	o.l.Lock()
	defer o.l.Unlock()
	h, ok := o.orders[pair]
	if !ok {
		return nil
	}
	items := h.Items()
	arrLen := len(items)
	if limit < arrLen {
		arrLen = limit
	}
	orders := make([]*EnergyOrder, arrLen)
	for i := 0; i < arrLen; i++ {
		orders[i] = items[i].Item
	}
	return orders
}
