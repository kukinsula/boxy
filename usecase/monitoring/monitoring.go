package monitoring

import (
	"time"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/entity/log"
	monitoringEntity "github.com/kukinsula/boxy/entity/monitoring"
)

type MonitoringGateway interface {
	Send(metrics *monitoringEntity.Metrics) error
}

type Monitoring struct {
	Interval          int
	monitoringGateway MonitoringGateway
	cpu               *monitoringEntity.CPU
	memory            *monitoringEntity.Memory
	network           *monitoringEntity.Network
	logger            log.Logger
}

func NewMonitoring(monitoringGateway MonitoringGateway, logger log.Logger) *Monitoring {
	return &Monitoring{
		Interval:          1,
		monitoringGateway: monitoringGateway,
		cpu:               monitoringEntity.NewCPU(),
		memory:            monitoringEntity.NewMemory(),
		network:           monitoringEntity.NewNetwork(),
		logger:            logger,
	}
}

func (monitoring *Monitoring) Start() error {
	for {
		metrics, err := monitoring.Update()
		if err != nil {
			return err
		}

		monitoring.logger(entity.NewUUID(), log.DEBUG, "New Metrics calculated",
			map[string]interface{}{
				"CPU":     metrics.CPU,
				"Memory":  metrics.Memory,
				"Network": metrics.Network,
			})

		err = monitoring.monitoringGateway.Send(metrics)
		if err != nil {
			return err
		}

		time.Sleep(time.Second)
	}
}

func (monitoring *Monitoring) Update() (*monitoringEntity.Metrics, error) {
	err := monitoring.cpu.Update()
	if err != nil {
		return nil, err
	}

	err = monitoring.memory.Update()
	if err != nil {
		return nil, err
	}

	err = monitoring.network.Update()
	if err != nil {
		return nil, err
	}

	return &monitoringEntity.Metrics{
		CPU:     monitoring.cpu,
		Memory:  monitoring.memory,
		Network: monitoring.network,
	}, nil
}
