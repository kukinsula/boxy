package server

import (
	"time"

	loginEntity "github.com/kukinsula/boxy/entity/login"

	"github.com/gin-gonic/gin"
)

type Context struct {
	*gin.Context

	UUID  string
	Start time.Time

	AccessToken string
	Requester   *loginEntity.User
}
