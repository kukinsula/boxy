package server

import (
	"context"

	loginEntity "github.com/kukinsula/boxy/entity/login"
	redisFramework "github.com/kukinsula/boxy/framework/redis"
	loginUsecase "github.com/kukinsula/boxy/usecase/login"
)

type Backend struct {
	Login     LoginBackender
	Streaming StreamingBackender
}

func NewBackend(
	login LoginBackender,
	streaming StreamingBackender) *Backend {

	return &Backend{
		Login:     login,
		Streaming: streaming,
	}
}

// TODO: enlever ça, ne passer faire une abstraction pour le moment
type LoginBackender interface {
	Signup(uuid string,
		context context.Context,
		params *loginUsecase.CreateUserParams) (*loginEntity.User, error)

	CheckActivate(uuid string,
		context context.Context,
		params *loginUsecase.EmailAndTokenParams) error

	Activate(uuid string,
		context context.Context,
		params *loginUsecase.EmailAndTokenParams) error

	Signin(uuid string,
		context context.Context,
		params *loginUsecase.SigninParams) (*loginUsecase.SigninResult, error)

	Me(uuid string,
		context context.Context,
		token string) (*loginUsecase.SigninResult, error)

	Logout(uuid string,
		context context.Context,
		token string) error
}

type StreamingBackender interface {
	Subscribe(context context.Context) *redisFramework.Subscription
}
