package monitoring

import (
	"fmt"

	monitoringEntity "github.com/kukinsula/boxy/entity/monitoring"
)

type monitoringGatewayMock struct{}

func newMonitoringGatewayMock() *monitoringGatewayMock {
	return &monitoringGatewayMock{}
}

func (gateway *monitoringGatewayMock) Send(metrics *monitoringEntity.Metrics) error {

	fmt.Printf("%s\n%s\n%s\n",
		metrics.CPU, metrics.Memory, metrics.Network)

	return nil
}
