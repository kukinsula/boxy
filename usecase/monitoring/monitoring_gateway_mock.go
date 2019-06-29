package monitoring

import (
	"fmt"

	monitoringEntity "github.com/kukinsula/boxy/entity/monitoring"
)

type monitoringGatewayMock struct{}

func newMonitoringGatewayMock() *monitoringGatewayMock {
	return &monitoringGatewayMock{}
}

func (gateway *monitoringGatewayMock) Send(
	cpu *monitoringEntity.CPU,
	memory *monitoringEntity.Memory,
	network *monitoringEntity.Network) error {

	fmt.Printf("%s\n%s\n%s\n", cpu, memory, network)

	return nil
}
