package client

import (
	"fmt"

	loginEntity "github.com/kukinsula/boxy/entity/login"
	loginUsecase "github.com/kukinsula/boxy/usecase/login"
)

type Login struct {
	*client
}

func NewLogin(
	URL string,
	requestLogger RequestLogger,
	responseLogger ResponseLogger) *Login {

	return &Login{
		client: newClient(
			URL,
			newRequester(),
			&JSONCodec{},
			requestLogger,
			responseLogger),
	}
}

func (login *Login) Signup(
	uuid string, params *loginUsecase.CreateUserParams) (*loginEntity.User, error) {

	result := &loginEntity.User{}
	resp, err := login.POST(&Request{
		UUID: uuid,
		Path: "/login/signup",
		Headers: map[string][]string{
			"Encoding-Type": []string{"application/json"},
		},
		Body: map[string]interface{}{
			"email":     params.Email,
			"password":  params.Password,
			"firstName": params.FirstName,
			"lastName":  params.LastName,
		},
	}).Decode(result)

	if err != nil {
		return nil, err
	}

	if resp.Status != 201 {
		return nil, fmt.Errorf("Signup should return Status code 201, not %d", resp.Status)
	}

	return result, nil
}

func (login *Login) CheckActivate(uuid string, params *loginUsecase.EmailAndTokenParams) error {
	resp := login.GET(&Request{
		UUID: uuid,
		Path: "/login/activate",
		Query: map[string]interface{}{
			"email": params.Email,
			"token": params.Token,
		},
	})

	if resp.Error != nil {
		return resp.Error
	}

	if resp.Status != 204 {
		return fmt.Errorf(
			"CheckActivation should return Status code 204, not %d", resp.Status)
	}

	return nil
}

func (login *Login) Activate(uuid string, params *loginUsecase.EmailAndTokenParams) error {
	resp := login.POST(&Request{
		UUID: uuid,
		Path: "/login/activate",
		Headers: map[string][]string{
			"Encoding-Type": []string{"application/json"},
		},
		Body: map[string]interface{}{
			"email": params.Email,
			"token": params.Token,
		},
	})

	if resp.Error != nil {
		return resp.Error
	}

	if resp.Status != 204 {
		return fmt.Errorf("Signup should return Status code 204, not %d", resp.Status)
	}

	return nil
}

func (login *Login) Signin(
	uuid string, params *loginUsecase.SigninParams) (*loginUsecase.SigninResult, error) {

	result := &loginUsecase.SigninResult{}
	resp, err := login.POST(&Request{
		UUID: uuid,
		Path: "/login/signin",
		Headers: map[string][]string{
			"Encoding-Type": []string{"application/json"},
		},
		Body: map[string]interface{}{
			"email":    params.Email,
			"password": params.Password,
		},
	}).Decode(result)

	if err != nil {
		return nil, err
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("Signin should return Status code 200, not %d", resp.Status)
	}

	return result, nil
}

func (login *Login) Me(uuid, token string) (*loginUsecase.SigninResult, error) {
	result := &loginUsecase.SigninResult{}
	resp, err := login.GET(&Request{
		UUID: uuid,
		Path: "/login/me",
		Headers: map[string][]string{
			"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
		},
	}).Decode(result)

	if err != nil {
		return nil, err
	}

	if resp.Status != 200 {
		return nil, fmt.Errorf("Me should return Status code 200, not %d", resp.Status)
	}

	return result, nil
}

func (login *Login) Logout(uuid, token string) error {
	resp := login.DELETE(&Request{
		UUID: uuid,
		Path: "/login/logout",
		Headers: map[string][]string{
			"Authorization": []string{fmt.Sprintf("Bearer %s", token)},
		},
	})

	if resp.Error != nil {
		return resp.Error
	}

	if resp.Status != 204 {
		return fmt.Errorf("Logout should return Status code 200, not %d", resp.Status)
	}

	return nil
}
