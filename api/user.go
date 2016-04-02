package api

import "fmt"

type User struct {
	*RequestBuilder
	token string
}

func (u *User) RevokeToken() error {
	resp := &struct {
		Revoked bool `json:"revoked"`
	}{}
	err := u.New("/v1/user/oauth/revoke").PostJSON(map[string]string{
		"token": u.token,
	}).DecodeInto(resp)
	if err != nil {
		return err
	}
	if resp.Revoked {
		u.RequestBuilder = nil
		return nil
	}
	return fmt.Errorf("Not Revoked")
}

func (u *User) GetAccountDetails() (*UserAccountDetails, error) {
	resp := &UserAccountDetails{}
	err := u.New("/v1/user/account").DecodeInto(resp)
	return resp, err
}
