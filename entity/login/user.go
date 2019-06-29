package login

import (
	"fmt"
	"strings"
)

type UserState int

const (
	NONE = iota - 1
	VALID
	INITIALIZING
	ACTIVATING
	ARCHIVED
)

var UserFullProjection = map[string]interface{}{
	"email":               1,
	"uuid":                1,
	"firstName":           1,
	"lastName":            1,
	"accessToken":         1,
	"initializationToken": 1,
	"activationToken":     1,
	"state":               1,
	"password":            1,
}

type User struct {
	UUID                string    `json:"uuid" bson:"uuid"`
	Email               string    `json:"email" bson:"email"`
	FirstName           string    `json:"firstName" bson:"firstName"`
	LastName            string    `json:"lastName" bson:"lastName"`
	AccessToken         string    `json:"access-token" bson:"accessToken"`
	ActivationToken     string    `json:"activation-token" bson:"activationToken"`
	InitializationToken string    `json:"initialization-token" bson:"initializationToken"`
	State               UserState `json:"state" bson:"state"`
	Password            string    `bson:"password"`
}

func NewUser(uuid,
	email,
	password,
	firstName,
	lastName,
	accessToken,
	activationToken,
	initializationToken string,
	state UserState) *User {

	return &User{
		UUID:                uuid,
		Email:               email,
		Password:            password,
		FirstName:           firstName,
		LastName:            lastName,
		AccessToken:         accessToken,
		ActivationToken:     activationToken,
		InitializationToken: initializationToken,
		State:               state,
	}
}

func (user *User) String() string {
	str := ""

	if user.UUID != "" {
		str = fmt.Sprintf("%s UUID:%s", str, user.UUID)
	}

	if user.Email != "" {
		str = fmt.Sprintf("%s Email:%s", str, user.Email)
	}

	if user.Password != "" {
		str = fmt.Sprintf("%s Password:%s", str, user.Password)
	}

	if user.FirstName != "" {
		str = fmt.Sprintf("%s FirstName%s", str, user.FirstName)
	}

	if user.LastName != "" {
		str = fmt.Sprintf("%s LastName:%s", str, user.LastName)
	}

	if user.State != -1 {
		str = fmt.Sprintf("%s State:%d", str, user.State)
	}

	if user.AccessToken != "" {
		str = fmt.Sprintf("%s AccessToken:%s", str, user.AccessToken)
	}

	if user.ActivationToken != "" {
		str = fmt.Sprintf("%s ActivationToken:%s", str, user.ActivationToken)
	}

	if user.InitializationToken != "" {
		str = fmt.Sprintf("%s InitializationToken:%s", str, user.InitializationToken)
	}

	return strings.Trim(str, " ")
}

type userBuilder struct {
	user *User
}

func NewUserBuilder() *userBuilder {
	return &userBuilder{
		user: &User{
			State: NONE,
		},
	}
}

func (builder *userBuilder) UUID(uuid string) *userBuilder {
	builder.user.UUID = uuid

	return builder
}

func (builder *userBuilder) Email(email string) *userBuilder {
	builder.user.Email = email

	return builder
}

func (builder *userBuilder) FirstName(firstName string) *userBuilder {
	builder.user.FirstName = firstName

	return builder
}

func (builder *userBuilder) LastName(lastName string) *userBuilder {
	builder.user.LastName = lastName

	return builder
}

func (builder *userBuilder) Password(password string) *userBuilder {
	builder.user.Password = password

	return builder
}

func (builder *userBuilder) AccessToken(token string) *userBuilder {
	builder.user.AccessToken = token

	return builder
}

func (builder *userBuilder) InitializationToken(token string) *userBuilder {
	builder.user.InitializationToken = token

	return builder
}

func (builder *userBuilder) ActivationToken(token string) *userBuilder {
	builder.user.ActivationToken = token

	return builder
}

func (builder *userBuilder) State(state UserState) *userBuilder {
	builder.user.State = state

	return builder
}

func (builder *userBuilder) Build() *User {
	return builder.user
}
