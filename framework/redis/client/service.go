package client

import (
	redisFramework "github.com/kukinsula/boxy/framework/redis"
)

type Service struct {
	*Login
}

func NewService(client *redisFramework.Client) *Service {
	return &Service{
		Login: NewLogin(client),
	}
}
