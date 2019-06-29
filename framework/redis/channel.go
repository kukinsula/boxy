package redis

type Channel string

const (
	METRICS = Channel("metrics")

	LOGIN_SIGNIN = Channel("login.signin")
	LOGIN_ME     = Channel("login.me")
	LOGIN_LOGOUT = Channel("login.logout")
)
