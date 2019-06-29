package login

import (
	"context"
	"fmt"
	"time"

	loginEntity "github.com/kukinsula/boxy/entity/login"
	"github.com/kukinsula/boxy/usecase"
)

// TODO
//
// * Mongo:
//   * transactions

type LoginGateway interface {
	Create(
		uuid string,
		ctx context.Context,
		user *loginEntity.User) (*loginEntity.User, error)

	FindByEmailAndActivationToken(
		uuid string,
		ctx context.Context,
		email, token string,
		projection map[string]interface{}) (*loginEntity.User, error)

	FindByEmailAndInitializationToken(
		uuid string,
		ctx context.Context,
		email, token string,
		projection map[string]interface{}) (*loginEntity.User, error)

	FindByEmail(
		uuid string,
		ctx context.Context,
		email string,
		projection map[string]interface{}) (*loginEntity.User, error)

	FindByAccessToken(
		uuid string,
		ctx context.Context,
		token string,
		projection map[string]interface{}) (*loginEntity.User, error)

	Update(
		uuid string,
		ctx context.Context,
		conditions map[string]interface{},
		update map[string]interface{}) error
}

type Login struct {
	loginGateway LoginGateway
	tokener      *usecase.Tokener
	passworder   *usecase.Passworder
}

func NewLogin(
	loginGateway LoginGateway,
	tokener *usecase.Tokener,
	passworder *usecase.Passworder) *Login {

	return &Login{
		loginGateway: loginGateway,
		tokener:      tokener,
		passworder:   passworder,
	}
}

type CreateUserParams struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type InitializeParams struct {
	Email    string `json:"email"`
	Token    string `json:"token"`
	Password string `json:"password"`
}

func (login *Login) Signup(
	uuid string,
	ctx context.Context,
	params CreateUserParams) (*loginEntity.User, error) {

	token, err := login.tokener.Generate(usecase.GenerateTokenParams{
		Audience:  "Users",
		ExpiresIn: time.Hour * 24,
		Issuer:    "Login",
		Subject:   "Signup",
		Email:     params.Email,
		UUID:      uuid,
	})

	if err != nil {
		return nil, err
	}

	encrypted, err := login.passworder.Hash(params.Password)
	if err != nil {
		return nil, err
	}

	user := loginEntity.NewUserBuilder().
		UUID(uuid).
		Email(params.Email).
		FirstName(params.FirstName).
		LastName(params.LastName).
		Password(string(encrypted)).
		State(loginEntity.ACTIVATING).
		ActivationToken(token).
		Build()

	return login.loginGateway.Create(uuid, ctx, user)
}

func (login *Login) CheckActivation(
	uuid string,
	ctx context.Context,
	email, token string) error {

	_, err := login.tokener.Verify(token)
	if err != nil {
		return err
	}

	user, err := login.loginGateway.FindByEmailAndActivationToken(
		uuid, ctx, email, token, map[string]interface{}{})

	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("CheckActivation failed: cannot find User with token %s", token)
	}

	return nil
}

func (login *Login) Activate(
	uuid string,
	ctx context.Context,
	email, token string) error {

	_, err := login.tokener.Verify(token)
	if err != nil {
		return err
	}

	user, err := login.loginGateway.FindByEmailAndActivationToken(uuid, ctx,
		email, token, map[string]interface{}{"uuid": 1})

	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("Activate failed: cannot find User with email %s", email)
	}

	err = login.loginGateway.Update(uuid, ctx,
		map[string]interface{}{"uuid": user.UUID},
		map[string]interface{}{
			"$set":   map[string]interface{}{"state": loginEntity.VALID},
			"$unset": map[string]interface{}{"activationToken": 1},
		})

	if err != nil {
		return err
	}

	return nil
}

type SigninParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SigninResult struct {
	UUID        string `json:"uuid""`
	Email       string `json:"email"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	AccessToken string `json:"access-token"`
}

func (login *Login) Signin(
	uuid string,
	ctx context.Context,
	params *SigninParams) (*SigninResult, error) {

	user, err := login.loginGateway.FindByEmail(uuid, ctx, params.Email,
		map[string]interface{}{
			"uuid":      1,
			"email":     1,
			"firstName": 1,
			"lastName":  1,
			"password":  1,
		})

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("Signin failed: cannot find User with email %s", params.Email)
	}

	err = login.passworder.Compare([]byte(user.Password), []byte(params.Password))
	if err != nil {
		return nil, err
	}

	token, err := login.tokener.Generate(usecase.GenerateTokenParams{
		Audience:  "Users",
		ExpiresIn: time.Hour * 24,
		Issuer:    "Login",
		Subject:   "Signin",
		Email:     params.Email,
	})

	if err != nil {
		return nil, err
	}

	err = login.loginGateway.Update(uuid, ctx,
		map[string]interface{}{"uuid": user.UUID},
		map[string]interface{}{"$set": map[string]interface{}{"accessToken": token}})

	if err != nil {
		return nil, err
	}

	return &SigninResult{
		UUID:        user.UUID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		AccessToken: token,
	}, nil
}

func (login *Login) Me(
	uuid string,
	ctx context.Context,
	token string) (*SigninResult, error) {

	_, err := login.tokener.Verify(token)
	if err != nil {
		return nil, err
	}

	user, err := login.loginGateway.FindByAccessToken(uuid, ctx, token,
		map[string]interface{}{
			"uuid":      1,
			"email":     1,
			"firstName": 1,
			"lastName":  1,
		})

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("Me failed: cannot find user with access token %s", token)
	}

	return &SigninResult{
		UUID:        user.UUID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		AccessToken: token,
	}, nil
}

func (login *Login) Create(
	uuid string,
	ctx context.Context,
	params CreateUserParams) (*loginEntity.User, error) {

	token, err := login.tokener.Generate(usecase.GenerateTokenParams{
		Audience:  "Users",
		ExpiresIn: time.Hour * 24,
		Issuer:    "Login",
		Subject:   "Create",
		Email:     params.Email,
		UUID:      uuid,
	})

	if err != nil {
		return nil, err
	}

	encrypted, err := login.passworder.Hash(params.Password)
	if err != nil {
		return nil, err
	}

	user := loginEntity.NewUserBuilder().
		UUID(uuid).
		Email(params.Email).
		Password(string(encrypted)).
		InitializationToken(token).
		State(loginEntity.INITIALIZING).
		Build()

	return login.loginGateway.Create(uuid, ctx, user)
}

func (login *Login) CheckInitialization(
	uuid string,
	ctx context.Context,
	email, token string) error {

	_, err := login.tokener.Verify(token)
	if err != nil {
		return err
	}

	user, err := login.loginGateway.FindByEmailAndInitializationToken(
		uuid, ctx, email, token, loginEntity.UserFullProjection)

	if err != nil {
		return err
	}

	if user == nil {
		return fmt.Errorf("CheckInitialization failed: cannot find user with email %s and initialization token %s", email, token)
	}

	return nil
}

func (login *Login) Initialize(
	uuid string,
	ctx context.Context,
	params InitializeParams) (*loginEntity.User, error) {

	_, err := login.tokener.Verify(params.Token)
	if err != nil {
		return nil, err
	}

	user, err := login.loginGateway.FindByEmailAndInitializationToken(
		uuid, ctx, params.Email, params.Token, map[string]interface{}{"uuid": 1})

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("Initialize failed: cannot find user with email %s and initialization token %s",
			params.Email, params.Token)
	}

	encrypted, err := login.passworder.Hash(params.Password)
	if err != nil {
		return nil, err
	}

	err = login.loginGateway.Update(uuid, ctx,
		map[string]interface{}{"uuid": user.UUID},
		map[string]interface{}{
			"$set": map[string]interface{}{
				"state":    loginEntity.VALID,
				"password": encrypted,
			},
			"$unset": map[string]interface{}{"initializationToken": 1},
		})

	if err != nil {
		return nil, err
	}

	user.AccessToken = params.Token

	return user, nil
}

func (login *Login) Logout(
	uuid string,
	ctx context.Context,
	token string) (*loginEntity.User, error) {

	_, err := login.tokener.Verify(token)
	if err != nil {
		return nil, err
	}

	user, err := login.loginGateway.FindByAccessToken(
		uuid, ctx, token, loginEntity.UserFullProjection)

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("Logout failed: cannot find user with access token %s", token)
	}

	err = login.loginGateway.Update(uuid, ctx,
		map[string]interface{}{"uuid": user.UUID},
		map[string]interface{}{"$unset": map[string]interface{}{"accessToken": 1}})

	if err != nil {
		return nil, err
	}

	user.AccessToken = ""

	return user, nil
}
