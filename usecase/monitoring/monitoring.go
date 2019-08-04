package monitoring

import (
	"time"

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
}

func NewMonitoring(monitoringGateway MonitoringGateway) *Monitoring {
	return &Monitoring{
		Interval:          1,
		monitoringGateway: monitoringGateway,
		cpu:               monitoringEntity.NewCPU(),
		memory:            monitoringEntity.NewMemory(),
		network:           monitoringEntity.NewNetwork(),
	}
}

func (monitoring *Monitoring) Start() error {
	for {
		metrics, err := monitoring.Update()
		if err != nil {
			return err
		}

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
