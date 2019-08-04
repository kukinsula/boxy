package client

import (
	"time"

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
	return monitoring.Publish(redisFramework.METRICS, metrics)
}

func (monitoring *Monitoring) Subscribe(ctx context.Context) chan *monitoringEntity.Metrics {
	conn := client.pool.Get()
	failure := make(chan error)
	channel := make(chan *monitoringEntity.Mtrics)
	subscription := NewSusbcription(req.UUID, ctx, redisFramework.METRICS, time.Minute)

	go func() {
		failure <- subscription.Start(conn)
		close(failure)
	}()

	go func() {
		for err == nil {
			select {
			case err = <-failure:

			case <-subscription.Subscribed:
				err = client.sendRequest(req)

			case data := <-subscription.Message:
				channel <- data
			}
		}
	}()

	return channel
}
