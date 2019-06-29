package monitoring

import (
	"time"

	monitoringEntity "github.com/kukinsula/boxy/entity/monitoring"
)

type MonitoringGateway interface {
	Send(
		cpu *monitoringEntity.CPU,
		memory *monitoringEntity.Memory,
		network *monitoringEntity.Network) error
}

type Monitoring struct {
	Interval          int
	monitoringGateway MonitoringGateway

	cpu     *monitoringEntity.CPU
	memory  *monitoringEntity.Memory
	network *monitoringEntity.Network
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
		err := monitoring.Update()
		if err != nil {
			return err
		}

		err = monitoring.monitoringGateway.Send(
			monitoring.cpu,
			monitoring.memory,
			monitoring.network)

		if err != nil {
			return err
		}

		time.Sleep(time.Second)
	}
}

func (monitoring *Monitoring) Update() error {
	err := monitoring.cpu.Update()
	if err != nil {
		return err
	}

	err = monitoring.memory.Update()
	if err != nil {
		return err
	}

	err = monitoring.network.Update()
	if err != nil {
		return err
	}

	return nil
}
