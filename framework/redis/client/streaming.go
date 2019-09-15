package client

import (
	"context"
	"time"

	redisFramework "github.com/kukinsula/boxy/framework/redis"
)

type Streaming struct {
	*redisFramework.Client
}

func NewStreaming(client *redisFramework.Client) *Streaming {
	return &Streaming{Client: client}
}

func (streaming *Streaming) Subscribe(ctx context.Context) *redisFramework.Subscription {
	return streaming.Client.Subscribe(ctx,
		redisFramework.STREAMING,
		time.Minute)
}
