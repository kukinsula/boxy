package client

import (
	monitoringEntity "github.com/kukinsula/boxy/entity/monitoring"
	redisFramework "github.com/kukinsula/boxy/framework/redis"
)

type Monitoring struct {
	*redisFramework.Client
}

func NewMonitoring(client *redisFramework.Client) *Monitoring {
	return &Monitoring{Client: client}
}

func (monitoring *Monitoring) Send(metrics *monitoringEntity.Metrics) error {
	return monitoring.Publish(redisFramework.STREAMING, metrics)
}
