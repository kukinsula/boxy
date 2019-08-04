package mongo

import (
	"context"

	"github.com/kukinsula/boxy/entity/log"
	loginEntity "github.com/kukinsula/boxy/entity/login"

	"go.mongodb.org/mongo-driver/bson"
)

// TODO
//
// * transactions

type UserModel struct {
	*model
}

func NewUserModel(
	ctx context.Context,
	database *Database,
	logger log.Logger) (*UserModel, error) {

	model, err := newModel(modelParams{
		Context:  ctx,
		Database: database,
		Name:     "users",
		Logger:   logger,

		Indexes: []indexParams{
			indexParams{
				Name:       "email",
				Value:      1,
				Unique:     true,
				Background: true,
				Sparse:     false,
			},

			indexParams{
				Name:       "uuid",
				Value:      1,
				Unique:     true,
				Background: true,
				Sparse:     false,
			},

			indexParams{
				Name:       "accessToken",
				Value:      1,
				Unique:     true,
				Background: true,
				Sparse:     true,
			},
		},
	})

	if err != nil {
		return nil, err
	}

	return &UserModel{model: model}, nil
}

func (model *UserModel) Create(
	uuid string,
	ctx context.Context,
	user *loginEntity.User) (*loginEntity.User, error) {

	err := model.InsertOne(uuid, ctx, bson.M{
		"email":           user.Email,
		"uuid":            user.UUID,
		"firstName":       user.FirstName,
		"lastName":        user.LastName,
		"password":        user.Password,
		"activationToken": user.ActivationToken,
	})

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (model *UserModel) FindByEmailAndActivationToken(
	uuid string,
	ctx context.Context,
	email, token string,
	projection map[string]interface{}) (*loginEntity.User, error) {

	user := &loginEntity.User{}

	err := model.FindOne(uuid, ctx,
		map[string]interface{}{"email": email, "activationToken": token},
		projection,
		user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (model *UserModel) FindByEmailAndInitializationToken(
	uuid string,
	ctx context.Context,
	email, token string,
	projection map[string]interface{}) (*loginEntity.User, error) {

	user := &loginEntity.User{}

	err := model.FindOne(uuid, ctx,
		map[string]interface{}{"email": email, "initializationtoken": token},
		projection,
		user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (model *UserModel) FindByEmail(
	uuid string,
	ctx context.Context,
	email string,
	projection map[string]interface{}) (*loginEntity.User, error) {

	user := &loginEntity.User{}

	err := model.FindOne(uuid, ctx,
		map[string]interface{}{"email": email},
		projection,
		user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (model *UserModel) FindByAccessToken(
	uuid string,
	ctx context.Context,
	token string,
	projection map[string]interface{}) (*loginEntity.User, error) {

	user := &loginEntity.User{}

	err := model.FindOne(uuid, ctx,
		map[string]interface{}{"accessToken": token},
		projection,
		user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (model *UserModel) Update(
	uuid string,
	ctx context.Context,
	conditions map[string]interface{},
	update map[string]interface{}) error {

	_, err := model.UpdateOne(uuid, ctx, conditions, update)
	if err != nil {
		return err
	}

	return nil
}
