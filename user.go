package tpt

import (
	"fmt"
	"net/http"
)

// User wraps an oAuth-ish token which can be used for user authenticated API
// calls. It extends RequestBuilder for custom http calls, and wraps some of
// the API calls to the TPT API
type User struct {
	*RequestBuilder
	Token string
}

// Modify implements Modifier, to add the User-Token to the request
func (u *User) Modify(req *http.Request) {
	req.Header.Add("User-Token", u.Token)
}

/////////////////////////
// Wrapped API Methods //
/////////////////////////

// RevokeToken revokes an existing access token. This token will no longer be
// able to be used for authentication.
func (u *User) RevokeToken() error {
	resp := &struct {
		Revoked bool `json:"revoked"`
	}{}
	err := u.New("/v1/user/oauth/revoke").PostJSON(map[string]string{
		"token": u.Token,
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

// GetAccountDetails returns the user’s account details.
func (u *User) GetAccountDetails() (*UserAccountDetails, error) {
	resp := &UserAccountDetails{}
	err := u.New("/v1/user/account").DecodeInto(resp)
	return resp, err
}
