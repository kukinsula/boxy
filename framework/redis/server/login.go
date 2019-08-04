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

	return client.Handle(&redisFramework.Handler{
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

	params := &loginUsecase.AccessTokenParams{}

	return client.Handle(&redisFramework.Handler{
		Channel: redisFramework.LOGIN_ME,
		Params:  params,
		Handle: func(uuid string, ctx context.Context) (interface{}, error) {
			return login.Me(uuid, ctx, params)
		},
	})
}

func HandleLogout(
	client *redisFramework.Client,
	login *loginUsecase.Login) error {

	params := &loginUsecase.AccessTokenParams{}

	return client.Handle(&redisFramework.Handler{
		Channel: redisFramework.LOGIN_LOGOUT,
		Params:  params,
		Handle: func(uuid string, ctx context.Context) (interface{}, error) {
			return nil, login.Logout(uuid, ctx, params)
		},
	})
}
