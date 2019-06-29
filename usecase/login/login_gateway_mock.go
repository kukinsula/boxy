package login

import (
	"context"
	"fmt"
	"sync"

	loginEntity "github.com/kukinsula/boxy/entity/login"
)

type loginGatewayMock struct {
	users map[string]*loginEntity.User
	mutex *sync.RWMutex
}

func newUserDatabaseMock() *loginGatewayMock {
	return &loginGatewayMock{
		users: map[string]*loginEntity.User{},
		mutex: &sync.RWMutex{},
	}
}

func (database *loginGatewayMock) Create(
	ctx context.Context,
	user *loginEntity.User) (*loginEntity.User, error) {

	database.mutex.Lock()
	defer database.mutex.Unlock()

	result, err := database.findByEmail(user.Email)
	if err != nil {
		return nil, err
	}

	if result != nil {
		return nil, fmt.Errorf("Create failed: email %s is already used", user.Email)
	}

	database.users[user.UUID] = user

	return user, nil
}

func (database *loginGatewayMock) findUserBy(
	comparator func(user *loginEntity.User) bool) (*loginEntity.User, error) {

	for _, user := range database.users {
		if comparator(user) {
			return user, nil
		}
	}

	return nil, nil
}

func (database *loginGatewayMock) findByEmailAndActivationToken(email, token string) (*loginEntity.User, error) {
	return database.findUserBy(func(user *loginEntity.User) bool {
		return user.Email == email && user.ActivationToken == token
	})
}

func (database *loginGatewayMock) findByEmailAndInitializationToken(email, token string) (*loginEntity.User, error) {
	return database.findUserBy(func(user *loginEntity.User) bool {
		return user.Email == email && user.InitializationToken == token
	})
}

func (database *loginGatewayMock) findByEmail(email string) (*loginEntity.User, error) {
	return database.findUserBy(func(user *loginEntity.User) bool {
		return user.Email == email
	})
}

func (database *loginGatewayMock) findByAccessToken(token string) (*loginEntity.User, error) {
	return database.findUserBy(func(user *loginEntity.User) bool {
		return user.AccessToken == token
	})
}

func (database *loginGatewayMock) FindByEmailAndActivationToken(
	ctx context.Context,
	email, token string,
	projection map[string]interface{}) (*loginEntity.User, error) {

	database.mutex.RLock()
	defer database.mutex.RUnlock()

	return database.findByEmailAndActivationToken(email, token)
}

func (database *loginGatewayMock) FindByEmailAndInitializationToken(
	ctx context.Context,
	email, token string,
	projection map[string]interface{}) (*loginEntity.User, error) {

	database.mutex.RLock()
	defer database.mutex.RUnlock()

	return database.findByEmailAndInitializationToken(email, token)
}

func (database *loginGatewayMock) FindByEmail(
	ctx context.Context,
	email string,
	projection map[string]interface{}) (*loginEntity.User, error) {

	database.mutex.RLock()
	defer database.mutex.RUnlock()

	return database.findByEmail(email)
}

func (database *loginGatewayMock) FindByAccessToken(
	ctx context.Context,
	token string,
	projection map[string]interface{}) (*loginEntity.User, error) {

	database.mutex.RLock()
	defer database.mutex.RUnlock()

	return database.findByAccessToken(token)
}

func (database *loginGatewayMock) update(
	uuid string,
	conditions map[string]interface{}) error {

	user, ok := database.users[uuid]
	if !ok {
		return fmt.Errorf("Upate failed: cannot find User with UUID %s", user.UUID)
	}

	database.users[user.UUID] = user

	return nil
}

func (database *loginGatewayMock) Update(
	ctx context.Context,
	conditions map[string]interface{},
	update map[string]interface{}) error {

	database.mutex.Lock()
	defer database.mutex.Unlock()

	rawUUID, ok := conditions["uuid"]
	if !ok {
		return fmt.Errorf("Update failed: cnnot find UUID in conditions %v", conditions)
	}

	uuid, ok := rawUUID.(string)
	if !ok {
		return fmt.Errorf("Update failed: %v is not a string", rawUUID)
	}

	return database.update(uuid, update)
}

func (database *loginGatewayMock) String() string {
	database.mutex.RLock()
	defer database.mutex.RUnlock()

	str := ""

	for _, user := range database.users {
		str = fmt.Sprintf("%s%#v\n", str, user)
	}

	return str
}
