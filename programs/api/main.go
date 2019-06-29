package main

import (
	"fmt"
	"time"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/framework/api/server"
	redis "github.com/kukinsula/boxy/framework/redis"
	redisClient "github.com/kukinsula/boxy/framework/redis/client"
)

func main() {
	client, err := redis.NewClient(redis.Config{
		Address:     "127.0.0.1:6379",
		MaxActive:   10,
		MaxIdle:     5,
		IdleTimeout: 200 * time.Second,
		Codec:       &entity.JSONCodec{},
		Logger:      entity.StdoutLogger,
	})

	if err != nil {
		fmt.Println("redis.NewClient failed: %s", err)
		return
	}

	service := redisClient.NewService(client)
	backend := server.NewBackend(service.Login)
	api := server.NewAPI(server.Config{
		Address: "127.0.0.1:8000",
		Backend: backend,
		Logger:  entity.StdoutLogger,
	})

	api.Run()
}
