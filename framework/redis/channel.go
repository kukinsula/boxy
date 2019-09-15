package redis

type Channel string

const (
	STREAMING = Channel("streaming")

	LOGIN_SIGNUP         = Channel("login.signup")
	LOGIN_CHECK_ACTIVATE = Channel("login.check_activate")
	LOGIN_ACTIVATE       = Channel("login.activate")
	LOGIN_SIGNIN         = Channel("login.signin")
	LOGIN_ME             = Channel("login.me")
	LOGIN_LOGOUT         = Channel("login.logout")
)
