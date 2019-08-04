package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/entity/log"
	"github.com/kukinsula/boxy/framework/api/client"
)

func main() {
	nbWorkers := 1
	logger := log.CleanMetaLogger(log.StdoutLogger)
	service := client.NewService("http://127.0.0.1:9000",
		func(req *client.Request) {
			logger(req.UUID, log.DEBUG,
				fmt.Sprintf("HTTP -> %s %s", req.Method, req.URL),
				map[string]interface{}{
					"headers": req.Headers,
					"body":    req.Body,
				})
		},

		func(resp *client.Response) {
			logger(resp.Request.UUID, log.DEBUG,
				fmt.Sprintf("HTTP <- %s %s", resp.Request.Method, resp.Request.URL),
				map[string]interface{}{
					"duration": resp.Duration,
					"headers":  resp.Headers,
					"status":   resp.Status,
					"error":    resp.Error,
				})
		})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	rand.Seed(time.Now().UnixNano())

	go func() {
		for index := 0; index < nbWorkers; index++ {
			go worker(service, logger, 500, 1500)
		}
	}()

	<-signals

	fmt.Println("Finished!")
}

func worker(service *client.Service, logger log.Logger, min, max int) {
	var err error

	randomer := rand.New(rand.NewSource(time.Now().UnixNano()))

	for err == nil {
		time.Sleep(getRandomDuration(randomer, min, max))

		resultSignin, err := service.Login.Signin(entity.NewUUID(), "titi@mail.io", "Azerty1234.")
		if err != nil {
			break
		}

		logger(entity.NewUUID(), log.DEBUG, "Signin result",
			map[string]interface{}{"result": *resultSignin, "error": err})

		time.Sleep(getRandomDuration(randomer, min, max))

		resultMe, err := service.Login.Me(entity.NewUUID(), resultSignin.AccessToken)
		if err != nil {
			break
		}

		logger(entity.NewUUID(), log.DEBUG, "Me result",
			map[string]interface{}{"result": *resultMe, "error": err})

		time.Sleep(getRandomDuration(randomer, min, max))

		err = service.Login.Logout(entity.NewUUID(), resultMe.AccessToken)
		if err != nil {
			break
		}

		logger(entity.NewUUID(), log.DEBUG, "Logout result",
			map[string]interface{}{"error": err})
	}

	logger(entity.NewUUID(), log.ERROR, "worker failed",
		map[string]interface{}{"error": err})
}

func getRandomDuration(randomer *rand.Rand, min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min)+min) * time.Millisecond
}
