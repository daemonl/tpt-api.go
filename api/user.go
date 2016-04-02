package api

import "net/http"

type User struct {
	Client
	Token string
}

func (u *User) Do(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Token", u.Token)
	return u.Client.Do(req)
}
