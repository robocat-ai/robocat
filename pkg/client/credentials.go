package robocat

import "net/url"

type Credentials struct {
	Username string
	Password string
}

func (c *Credentials) GetUserInfo() *url.Userinfo {
	if len(c.Password) == 0 {
		return url.User(c.Username)
	}

	return url.UserPassword(c.Username, c.Password)
}
