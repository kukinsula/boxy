package redis

import (
	"context"
	"time"

	"github.com/kukinsula/boxy/entity"

	"github.com/gomodule/redigo/redis"
)

type Subscription struct {
	UUID       string
	Context    context.Context
	channel    Channel
	pubsub     redis.PubSubConn
	ping       time.Duration
	logger     entity.Logger
	Subscribed chan struct{}
	Message    chan []byte
}

func NewSusbcription(
	uuid string,
	ctx context.Context,
	channel Channel,
	pubsub redis.PubSubConn,
	ping time.Duration,
	logger entity.Logger) *Subscription {

	return &Subscription{
		UUID:       uuid,
		Context:    ctx,
		channel:    channel,
		pubsub:     pubsub,
		ping:       ping,
		logger:     logger,
		Subscribed: make(chan struct{}),
		Message:    make(chan []byte),
	}
}

func (subscription *Subscription) Start() error {
	err := subscription.pubsub.Subscribe(string(subscription.channel))
	if err != nil {
		return err
	}

	failure := make(chan error)
	done := make(chan struct{})

	go func() {
		failure <- subscription.receive(done)
		close(failure)
	}()

	ticker := time.NewTicker(subscription.ping)
	defer ticker.Stop()

	for goOn := true; goOn; goOn = goOn && err == nil {
		select {
		case err = <-failure: // receive failed or ended

		case <-ticker.C: // Connection health check
			err = subscription.pubsub.Ping("")

		case <-subscription.Context.Done():
			goOn = false
		}
	}

	subscription.pubsub.Unsubscribe(string(subscription.channel))
	subscription.pubsub.Close()

	<-done
	close(done)

	subscription.logger(entity.Log{
		UUID:    subscription.UUID,
		Level:   "debug",
		Message: "REDIS subscription ended",
		Meta:    map[string]interface{}{"channel": subscription.channel},
	})

	return err
}

func (subscription *Subscription) receive(done chan struct{}) error {
	var err error

	for goOn := true; goOn; goOn = goOn && err == nil {
		switch resp := subscription.pubsub.Receive().(type) {
		case error:
			err = resp

		case redis.Subscription:
			switch resp.Count {
			case 0:
				goOn = false

			case 1:
				subscription.Subscribed <- struct{}{}

				subscription.logger(entity.Log{
					UUID:    subscription.UUID,
					Level:   "debug",
					Message: "REDIS subscribed",
					Meta:    map[string]interface{}{"channel": subscription.channel},
				})
			}

		case redis.Message:
			subscription.Message <- resp.Data
		}
	}

	close(subscription.Subscribed)
	close(subscription.Message)

	done <- struct{}{}

	return err
}
