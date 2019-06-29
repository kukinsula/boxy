package client

import (
	"context"
	"time"

	loginEntity "github.com/kukinsula/boxy/entity/login"
	redisFramework "github.com/kukinsula/boxy/framework/redis"
	loginUsecase "github.com/kukinsula/boxy/usecase/login"
)

type Login struct {
	*redisFramework.Client
}

func NewLogin(client *redisFramework.Client) *Login {
	return &Login{Client: client}
}

func (login *Login) Signin(
	uuid string,
	context context.Context,
	params loginUsecase.SigninParams) (*loginEntity.User, error) {

	user := &loginEntity.User{}

	err := login.Request(redisFramework.Request{
		UUID:    uuid,
		Context: context,
		Channel: redisFramework.LOGIN_SIGNIN,
		Params:  params,
		Result:  user,
		Ping:    time.Minute,
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (login *Login) Me(
	uuid string,
	context context.Context,
	token string) (*loginEntity.User, error) {

	user := &loginEntity.User{}

	err := login.Request(redisFramework.Request{
		UUID:    uuid,
		Context: context,
		Channel: redisFramework.LOGIN_ME,
		Params:  token,
		Result:  user,
		Ping:    time.Minute,
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (login *Login) Logout(
	uuid string,
	context context.Context,
	token string) (*loginEntity.User, error) {

	user := &loginEntity.User{}

	err := login.Request(redisFramework.Request{
		UUID:    uuid,
		Context: context,
		Channel: redisFramework.LOGIN_LOGOUT,
		Params:  token,
		Result:  user,
		Ping:    time.Minute,
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}
