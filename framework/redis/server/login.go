package server

import (
	"context"

	redisFramework "github.com/kukinsula/boxy/framework/redis"
	loginUsecase "github.com/kukinsula/boxy/usecase/login"
)

func HandleSignin(
	client *redisFramework.Client,
	login *loginUsecase.Login) error {

	params := &loginUsecase.SigninParams{}

	return client.Handle(redisFramework.Handler{
		Channel: redisFramework.LOGIN_SIGNIN,
		Params:  params,
		Handle: func(uuid string, ctx context.Context) (interface{}, error) {
			return login.Signin(uuid, ctx, params)
		},
	})
}

func HandleMe(
	client *redisFramework.Client,
	login *loginUsecase.Login) error {

	var token string

	return client.Handle(redisFramework.Handler{
		Channel: redisFramework.LOGIN_ME,
		Params:  &token,
		Handle: func(uuid string, ctx context.Context) (interface{}, error) {
			return login.Me(uuid, ctx, token)
		},
	})
}

func HandleLogout(
	client *redisFramework.Client,
	login *loginUsecase.Login) error {

	var token string

	return client.Handle(redisFramework.Handler{
		Channel: redisFramework.LOGIN_LOGOUT,
		Params:  &token,
		Handle: func(uuid string, ctx context.Context) (interface{}, error) {
			return login.Logout(uuid, ctx, token)
		},
	})
}
