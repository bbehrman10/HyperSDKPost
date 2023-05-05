package controller

import (
	ametrics "github.com/ava-labs/avalanchego/api/metrics"
	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/ava-labs/hypersdk/examples/tokenvm/consts"
	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	initializeEnergyAsset prometheus.Counter
	produceEnergy         prometheus.Counter
	consumeEnergy         prometheus.Counter
	createEnergyOrder     prometheus.Counter
	fillEnergyOrder       prometheus.Counter
	closeEnergyOrder      prometheus.Counter
}

func newMetrics(gatherer ametrics.MultiGatherer) (*metrics, error) {
	m := &metrics{
		initializeEnergyAsset: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "initialize_energy_asset",
			Help:      "number of initialize asset actions",
		}),
		produceEnergy: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "produceEnergy",
			Help:      "number of produce energy actions",
		}),
		consumeEnergy: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "consumeEnergy",
			Help:      "number of consume energy actions",
		}),
		createEnergyOrder: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "create_energy_order",
			Help:      "number of create order actions",
		}),
		fillEnergyOrder: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "fill_energy_order",
			Help:      "number of fill energy order actions",
		}),
		closeEnergyOrder: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "actions",
			Name:      "close_energy_order",
			Help:      "number of close energy order actions",
		}),
	}
	r := prometheus.NewRegistry()
	errs := wrappers.Errs{}
	errs.Add(
		r.Register(m.initializeEnergyAsset),
		r.Register(m.produceEnergy),
		r.Register(m.consumeEnergy),
		r.Register(m.createEnergyOrder),
		r.Register(m.fillEnergyOrder),
		r.Register(m.closeEnergyOrder),
		gatherer.Register(consts.Name, r),
	)
	return m, errs.Err
}
