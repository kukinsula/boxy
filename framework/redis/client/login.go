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

func (login *Login) Signup(
	uuid string,
	context context.Context,
	params *loginUsecase.CreateUserParams) (*loginEntity.User, error) {

	result := &loginEntity.User{}
	err := login.Request(&redisFramework.Request{
		UUID:    uuid,
		Context: context,
		Channel: redisFramework.LOGIN_SIGNUP,
		Params:  params,
		Ping:    time.Minute,
	}).Decode(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (login *Login) CheckActivate(
	uuid string,
	context context.Context,
	params *loginUsecase.EmailAndTokenParams) error {

	return login.Request(&redisFramework.Request{
		UUID:    uuid,
		Context: context,
		Channel: redisFramework.LOGIN_CHECK_ACTIVATE,
		Params:  params,
		Ping:    time.Minute,
	}).Error
}

func (login *Login) Activate(
	uuid string,
	context context.Context,
	params *loginUsecase.EmailAndTokenParams) error {

	return login.Request(&redisFramework.Request{
		UUID:    uuid,
		Context: context,
		Channel: redisFramework.LOGIN_ACTIVATE,
		Params:  params,
		Ping:    time.Minute,
	}).Error
}

func (login *Login) Signin(
	uuid string,
	context context.Context,
	params *loginUsecase.SigninParams) (*loginUsecase.SigninResult, error) {

	result := &loginUsecase.SigninResult{}
	err := login.Request(&redisFramework.Request{
		UUID:    uuid,
		Context: context,
		Channel: redisFramework.LOGIN_SIGNIN,
		Params:  params,
		Ping:    time.Minute,
	}).Decode(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (login *Login) Me(
	uuid string,
	context context.Context,
	token string) (*loginUsecase.SigninResult, error) {

	result := &loginUsecase.SigninResult{}
	err := login.Request(&redisFramework.Request{
		UUID:    uuid,
		Context: context,
		Channel: redisFramework.LOGIN_ME,
		Params:  &loginUsecase.AccessTokenParams{Token: token},
		Ping:    time.Minute,
	}).Decode(result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (login *Login) Logout(
	uuid string,
	context context.Context,
	token string) error {

	resp := login.Request(&redisFramework.Request{
		UUID:    uuid,
		Context: context,
		Channel: redisFramework.LOGIN_LOGOUT,
		Params:  &loginUsecase.AccessTokenParams{Token: token},
		Ping:    time.Minute,
	})

	return resp.Error
}
