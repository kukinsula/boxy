package redis

import (
	"context"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Subscription struct {
	Context    context.Context
	channel    Channel
	ping       time.Duration
	Subscribed chan struct{}
	Message    chan []byte
}

func NewSusbcription(
	ctx context.Context,
	channel Channel,
	ping time.Duration) *Subscription {

	return &Subscription{
		Context:    ctx,
		channel:    channel,
		ping:       ping,
		Subscribed: make(chan struct{}),
		Message:    make(chan []byte),
	}
}

func (subscription *Subscription) Start(conn redis.Conn) error {
	pubsub := redis.PubSubConn{Conn: conn}
	err := pubsub.Subscribe(string(subscription.channel))
	if err != nil {
		return err
	}

	failure := make(chan error)
	done := make(chan struct{})

	go func() {
		failure <- subscription.receive(pubsub, done)
		close(failure)
	}()

	ticker := time.NewTicker(subscription.ping)
	defer ticker.Stop()

	for goOn := true; goOn; goOn = goOn && err == nil {
		select {
		case err = <-failure: // receive failed or ended

		case <-ticker.C: // Connection health check
			err = pubsub.Ping("")

		case <-subscription.Context.Done():
			goOn = false
		}
	}

	pubsub.Unsubscribe(string(subscription.channel))

	<-done

	conn.Close()

	return err
}

func (subscription *Subscription) receive(
	pubsub redis.PubSubConn,
	done chan struct{}) (err error) {

	for goOn := true; goOn; goOn = goOn && err == nil {
		switch result := pubsub.Receive().(type) {
		case error:
			err = result

		case redis.Subscription:
			switch result.Count {
			case 0:
				goOn = false

			case 1:
				subscription.Subscribed <- struct{}{}
			}

		case redis.Message:
			subscription.Message <- result.Data
		}
	}

	close(subscription.Subscribed)
	close(subscription.Message)

	done <- struct{}{}
	close(done)

	return err
}
