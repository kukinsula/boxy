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
	loginEntity "github.com/kukinsula/boxy/entity/login"
	"github.com/kukinsula/boxy/framework/api/client"
	loginUsecase "github.com/kukinsula/boxy/usecase/login"
)

func main() {
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
		workers := 10

		for index := 0; index < workers; index++ {
			user := &loginEntity.User{
				Password:  GenerateRandomString(8),
				FirstName: GenerateRandomString(4),
				LastName:  GenerateRandomString(5),
			}

			user.Email = fmt.Sprintf("%s.%s@mail.io", user.FirstName, user.LastName)

			go worker(service, logger, user, 1000, 2000)
		}
	}()

	// go Stream(service, logger)

	<-signals

	fmt.Println("Finished!")
}

func worker(
	service *client.Service,
	logger log.Logger,
	user *loginEntity.User,
	min, max int) {

	var signinResult *loginUsecase.SigninResult
	var meResult *loginUsecase.SigninResult
	var err error

	randomer := rand.New(rand.NewSource(time.Now().UnixNano()))

	time.Sleep(getRandomDuration(randomer, min, max))

	signupResult, err := service.Login.Signup(entity.NewUUID(), &loginUsecase.CreateUserParams{
		Email:     user.Email,
		Password:  user.Password,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})

	if err != nil {
		logger(entity.NewUUID(), log.ERROR, "worker Signup failed",
			map[string]interface{}{"error": err})
		return
	}

	err = service.Login.CheckActivate(entity.NewUUID(), &loginUsecase.EmailAndTokenParams{
		Email: signupResult.Email,
		Token: signupResult.ActivationToken,
	})

	if err != nil {
		logger(entity.NewUUID(), log.ERROR, "worker CheckActivate failed",
			map[string]interface{}{"error": err})
		return
	}

	err = service.Login.Activate(entity.NewUUID(), &loginUsecase.EmailAndTokenParams{
		Email: signupResult.Email,
		Token: signupResult.ActivationToken,
	})

	if err != nil {
		logger(entity.NewUUID(), log.ERROR, "worker Activate failed",
			map[string]interface{}{"error": err})
		return
	}

	for err == nil {
		time.Sleep(getRandomDuration(randomer, min, max))

		signinResult, err = service.Login.Signin(entity.NewUUID(), &loginUsecase.SigninParams{
			Email:    user.Email,
			Password: user.Password,
		})

		if err != nil {
			break
		}

		time.Sleep(getRandomDuration(randomer, min, max))

		meResult, err = service.Login.Me(entity.NewUUID(), signinResult.AccessToken)
		if err != nil {
			break
		}

		time.Sleep(getRandomDuration(randomer, min, max))

		err = service.Login.Logout(entity.NewUUID(), meResult.AccessToken)
		if err != nil {
			break
		}
	}

	logger(entity.NewUUID(), log.ERROR, "worker ended",
		map[string]interface{}{"error": err})
}

func Stream(service *client.Service, logger log.Logger) {
	signinResult, err := service.Login.Signin(entity.NewUUID(), &loginUsecase.SigninParams{
		Email:    "titi@mail.io",
		Password: "Azerty1234.",
	})

	if err != nil {
		return
	}

	logger(entity.NewUUID(), log.DEBUG, "Signin result",
		map[string]interface{}{"result": *signinResult, "error": err})

	channel, err := service.Streaming.Stream(entity.NewUUID(), signinResult.AccessToken)
	if err != nil {
		logger(entity.NewUUID(), log.ERROR, "streaming failed",
			map[string]interface{}{"error": err})
	}

	for metrics := range channel {
		fmt.Printf("METRICS %s\n", metrics)
	}
}

func getRandomDuration(randomer *rand.Rand, min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min)+min) * time.Millisecond
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateRandomString(n int) string {
	b := make([]rune, n)

	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}
