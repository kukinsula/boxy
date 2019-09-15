package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kukinsula/boxy/entity/codec"
	"github.com/kukinsula/boxy/entity/log"
	"github.com/kukinsula/boxy/framework/mongo"
	redis "github.com/kukinsula/boxy/framework/redis"
	redisServer "github.com/kukinsula/boxy/framework/redis/server"
	"github.com/kukinsula/boxy/usecase"
	loginUsecase "github.com/kukinsula/boxy/usecase/login"
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

	ctx := context.Background()
	database, err := mongo.NewDatabase(mongo.NewDatabaseParams{
		Context:  ctx,
		URI:      "mongodb://localhost:27017",
		Database: "boxy",
		Logger:   logger,
	})

	if err != nil {
		fmt.Println("NewDatabase failed: %s", err)
		return
	}

	err = database.Init(ctx)

	if err != nil {
		fmt.Println("Database.Init failed: %s", err)
		return
	}

	tokener := usecase.NewTokener("TopSecret")
	passworder := usecase.NewPassworder(10)
	login := loginUsecase.NewLogin(database.User, tokener, passworder)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go redisServer.HandleSignup(client, login)
	go redisServer.HandleSignin(client, login)
	go redisServer.HandleCheckActivate(client, login)
	go redisServer.HandleActivate(client, login)
	go redisServer.HandleMe(client, login)
	go redisServer.HandleLogout(client, login)

	<-signals
	fmt.Println("Finished!")
}
