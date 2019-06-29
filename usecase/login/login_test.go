package login

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kukinsula/boxy/entity"
	loginEntity "github.com/kukinsula/boxy/entity/login"
	mongo "github.com/kukinsula/boxy/framework/mongo"
	"github.com/kukinsula/boxy/usecase"
)

// func TestMockLogin(t *testing.T) {
// 	database := newUserDatabaseMock()

// 	testLogin(t, database)
// }

func TestMongoLogin(t *testing.T) {
	ctx := context.Background()
	database, err := mongo.NewDatabase(mongo.NewDatabaseParams{
		Context:  ctx,
		URI:      "mongodb://localhost:27017",
		Database: "boxy_test",
		Logger: func(uuid interface{},
			collection, operation string,
			datas ...interface{}) error {

			_, err := fmt.Printf("[%s] %s.%s(%v)\n",
				[]interface{}{uuid, collection, operation, datas}...)

			return err
		},
	})

	if err != nil {
		t.Errorf("NewDatabase failed: %s", err)
		t.FailNow()
	}

	err = database.Drop(ctx)

	if err != nil {
		t.Errorf("Drop failed: %s", err)
		t.FailNow()
	}

	err = database.Init(ctx)

	if err != nil {
		t.Errorf("Drop failed: %s", err)
		t.FailNow()
	}

	testLogin(t, database.User)

	// err = database.Drop(ctx)
	// if err != nil {
	// 	t.Errorf("Drop failed: %s", err)
	// 	t.FailNow()
	// }
}

func testLogin(t *testing.T, database UserDatabase) {
	tokener := usecase.NewTokener("TopSecret")
	fakeTokener := usecase.NewTokener("WrongSecret")
	passworder := usecase.NewPassworder(10)
	login := NewLoginUseCase(database, tokener, passworder)

	uuid := entity.NewUUID()
	ctx := context.WithValue(context.Background(), "id", uuid)

	builder := loginEntity.NewUserBuilder()
	builder.Email("titi@mail.io").Password("Azerty1234.")
	builder.FirstName("Ti").LastName("Ti")
	user := builder.Build()

	badEmail := "toto@mail.io"

	absentToken, err := tokener.Generate(usecase.GenerateTokenParams{
		Audience:  "Users",
		ExpiresIn: time.Hour * 24,
		Issuer:    "Login",
		Subject:   "Signup",
		Email:     user.Email,
		UUID:      uuid,
	})

	if err != nil {
		t.Errorf("Tokener (absent token) should not fail: %s", err)
		t.FailNow()
	}

	invalidToken, err := fakeTokener.Generate(usecase.GenerateTokenParams{
		Audience:  "Users",
		ExpiresIn: time.Hour * 24,
		Issuer:    "Login",
		Subject:   "Signup",
		Email:     user.Email,
		UUID:      uuid,
	})

	if err != nil {
		t.Errorf("Tokener (invalid token) should not fail: %s", err)
		t.FailNow()
	}

	// Signup

	result, err := login.Signup(ctx, CreateUserParams{
		Email:     user.Email,
		Password:  string(user.Password),
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})

	if err != nil {
		t.Errorf("Signup should not fail: %s", err)
		t.FailNow()
	}

	if result == nil {
		t.Error("Signup should return a valid User")
		t.FailNow()
	}

	tmp, err := login.Signup(ctx, CreateUserParams{
		Email:     user.Email,
		Password:  string(user.Password),
		FirstName: user.FirstName,
		LastName:  user.LastName,
	})

	if err == nil || tmp != nil {
		t.Errorf("Signup should fail because User %s should already exist", user.Email)
		t.FailNow()
	}

	// CheckActivation

	err = login.CheckActivation(ctx, user.Email, result.ActivationToken)
	if err != nil {
		t.Errorf("CheckActivation should not fail: %s", err)
		t.FailNow()
	}

	err = login.CheckActivation(ctx, user.Email, absentToken)
	if err == nil {
		t.Errorf("CheckActivation should fail: %s token should not exist", absentToken)
		t.FailNow()
	}

	err = login.CheckActivation(ctx, user.Email, invalidToken)
	if err == nil {
		t.Errorf("CheckActivation should fail: %s token should not be valid", invalidToken)
		t.FailNow()
	}

	err = login.CheckActivation(ctx, badEmail, result.ActivationToken)
	if err == nil {
		t.Errorf("CheckActivation should fail: user with email %s should not be found", badEmail)
		t.FailNow()
	}

	// Activate

	token := result.ActivationToken

	err = login.Activate(ctx, user.Email, token)
	if err != nil {
		t.Errorf("Activate should not fail: %s", err)
		t.FailNow()
	}

	err = login.Activate(ctx, user.Email, token)
	if err == nil {
		t.Errorf("Activate should fail: user with email %s already activated", user.Email)
		t.FailNow()
	}

	err = login.Activate(ctx, user.Email, absentToken)
	if err == nil {
		t.Errorf("Activate should fail: %s should not be found", absentToken)
		t.FailNow()
	}

	err = login.Activate(ctx, user.Email, invalidToken)
	if err == nil {
		t.Errorf("Activate should fail: %s should not be valid", invalidToken)
		t.FailNow()
	}

	// Signin

	result, err = login.Signin(ctx, SigninParams{
		Email:    user.Email,
		Password: string(user.Password),
	})
	if err != nil {
		t.Errorf("Signin should not fail: %s", err)
		t.FailNow()
	}

	if result == nil {
		t.Error("Signin should return a valid User")
		t.FailNow()
	}

	badPassword := "BadPassword"
	tmp, err = login.Signin(ctx, SigninParams{
		Email:    user.Email,
		Password: badPassword,
	})
	if err == nil || result == nil {
		t.Errorf("Signin should fail: password %s is incorrect", badPassword)
		t.FailNow()
	}

	tmp, err = login.Signin(ctx, SigninParams{
		Email:    badEmail,
		Password: user.Password,
	})
	if err == nil || result == nil {
		t.Errorf("Signin should fail: user %s doesn't exist", badEmail)
		t.FailNow()
	}

	// Me

	result, err = login.Me(ctx, result.AccessToken)
	if err != nil {
		t.Errorf("Me should not fail: %s", err)
		t.FailNow()
	}

	if result == nil {
		t.Error("Me should return a valid User")
		t.FailNow()
	}

	tmp, err = login.Me(ctx, absentToken)
	if err == nil || result == nil {
		t.Errorf("Me should fail: access token %s is absent", absentToken)
		t.FailNow()
	}

	tmp, err = login.Me(ctx, invalidToken)
	if err == nil || result == nil {
		t.Errorf("Me should fail: access token %s is invalid", invalidToken)
		t.FailNow()
	}

	// Logout

	token = result.AccessToken

	result, err = login.Logout(ctx, token)
	if err != nil {
		t.Errorf("Logout should not fail: %s", err)
		t.FailNow()
	}

	if result == nil {
		t.Errorf("Logout should return a valid User")
		t.FailNow()
	}

	tmp, err = login.Me(ctx, token)
	if err == nil || tmp != nil {
		t.Errorf("Me should fail: access token %s should not be valid anymore", result.AccessToken)
		t.FailNow()
	}
}
