package main

import (
	"fmt"
	"time"

	"github.com/kukinsula/boxy/entity/codec"
	"github.com/kukinsula/boxy/entity/log"
	"github.com/kukinsula/boxy/framework/redis"
	redisClient "github.com/kukinsula/boxy/framework/redis/client"
	monitoringUsecase "github.com/kukinsula/boxy/usecase/monitoring"
)

func main() {
	logger := log.CleanMetaLogger(log.StdoutLogger)
	client, err := redis.NewClient(redis.Config{
		Address:     "127.0.0.1:6379",
		MaxActive:   10,
		MaxIdle:     5,
		IdleTimeout: 200 * time.Second,
		Codec:       &codec.JSONCodec{},
		Logger:      logger,
	})

	if err != nil {
		fmt.Println("redis.NewClient failed: %s", err)
		return
	}

	gateway := redisClient.NewMonitoring(client)
	monitoring := monitoringUsecase.NewMonitoring(gateway)

	monitoring.Start()
}
