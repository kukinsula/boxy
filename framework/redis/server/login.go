package server

import (
	"context"

	redisFramework "github.com/kukinsula/boxy/framework/redis"
	loginUsecase "github.com/kukinsula/boxy/usecase/login"
)

// Signup

type signupHandler struct {
	login  *loginUsecase.Login
	params *loginUsecase.CreateUserParams
}

func (handler *signupHandler) Params() interface{} { return handler.params }

func (handler *signupHandler) Exec(uuid string, ctx context.Context) (interface{}, error) {
	return handler.login.Signup(uuid, ctx, handler.params)
}

func HandleSignup(
	client *redisFramework.Client,
	login *loginUsecase.Login) error {

	return client.Handle(redisFramework.LOGIN_SIGNUP, func() redisFramework.Handler {
		return &signupHandler{login: login, params: &loginUsecase.CreateUserParams{}}
	})
}

// CheckActivate

type checkActivateHandler struct {
	login  *loginUsecase.Login
	params *loginUsecase.EmailAndTokenParams
}

func (handler *checkActivateHandler) Params() interface{} { return handler.params }

func (handler *checkActivateHandler) Exec(uuid string, ctx context.Context) (interface{}, error) {
	return nil, handler.login.CheckActivate(uuid, ctx, handler.params)
}

func HandleCheckActivate(
	client *redisFramework.Client,
	login *loginUsecase.Login) error {

	return client.Handle(redisFramework.LOGIN_CHECK_ACTIVATE, func() redisFramework.Handler {
		return &checkActivateHandler{login: login, params: &loginUsecase.EmailAndTokenParams{}}
	})
}

// Activate

type activateHandler struct {
	login  *loginUsecase.Login
	params *loginUsecase.EmailAndTokenParams
}

func (handler *activateHandler) Params() interface{} { return handler.params }

func (handler *activateHandler) Exec(uuid string, ctx context.Context) (interface{}, error) {
	return nil, handler.login.Activate(uuid, ctx, handler.params)
}

func HandleActivate(
	client *redisFramework.Client,
	login *loginUsecase.Login) error {

	return client.Handle(redisFramework.LOGIN_ACTIVATE, func() redisFramework.Handler {
		return &activateHandler{login: login, params: &loginUsecase.EmailAndTokenParams{}}
	})
}

// Signin

type signinHandler struct {
	login  *loginUsecase.Login
	params *loginUsecase.SigninParams
}

func (handler *signinHandler) Params() interface{} { return handler.params }

func (handler *signinHandler) Exec(uuid string, ctx context.Context) (interface{}, error) {
	return handler.login.Signin(uuid, ctx, handler.params)
}

func HandleSignin(
	client *redisFramework.Client,
	login *loginUsecase.Login) error {

	return client.Handle(redisFramework.LOGIN_SIGNIN, func() redisFramework.Handler {
		return &signinHandler{login: login, params: &loginUsecase.SigninParams{}}
	})
}

// Me

type meHandler struct {
	login  *loginUsecase.Login
	params *loginUsecase.AccessTokenParams
}

func (handler *meHandler) Params() interface{} { return handler.params }

func (handler *meHandler) Exec(uuid string, ctx context.Context) (interface{}, error) {
	return handler.login.Me(uuid, ctx, handler.params)
}

func HandleMe(
	client *redisFramework.Client,
	login *loginUsecase.Login) error {

	return client.Handle(redisFramework.LOGIN_ME, func() redisFramework.Handler {
		return &meHandler{login: login, params: &loginUsecase.AccessTokenParams{}}
	})
}

// Logout

type logoutHandler struct {
	login  *loginUsecase.Login
	params *loginUsecase.AccessTokenParams
}

func (handler *logoutHandler) Params() interface{} { return handler.params }

func (handler *logoutHandler) Exec(uuid string, ctx context.Context) (interface{}, error) {
	return nil, handler.login.Logout(uuid, ctx, handler.params)
}

func HandleLogout(
	client *redisFramework.Client,
	login *loginUsecase.Login) error {

	return client.Handle(redisFramework.LOGIN_LOGOUT, func() redisFramework.Handler {
		return &logoutHandler{login: login, params: &loginUsecase.AccessTokenParams{}}
	})
}
