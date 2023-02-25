package ws

import (
	"fmt"
	"net/url"
)

type RunnerArguments struct {
	Flow  string `json:"flow"`
	Data  string `json:"data"`
	Proxy string `json:"proxy"`
}

func (a *RunnerArguments) ToArray() []string {
	args := []string{a.Flow}

	if len(a.Data) > 0 {
		args = append(args, "--data", a.Data)
	}

	if len(a.Proxy) > 0 {
		u, err := url.Parse(a.Proxy)
		if err == nil {
			address := fmt.Sprintf("%s:%s", u.Hostname(), u.Port())
			if u.User != nil {
				address = fmt.Sprintf("%s@%s", u.User.String(), address)
			}

			args = append(args, "--proxy-protocol", u.Scheme)
			args = append(args, "--proxy-address", address)
		}
	}

	return args
}
