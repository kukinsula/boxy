package monitoring

import (
	"testing"
)

func TestMockMonitoring(t *testing.T) {
	gateway := newMonitoringGatewayMock()
	monitoring := NewMonitoring(gateway)

	monitoring.Start()
}
