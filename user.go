package tpt

import (
	"fmt"

	"github.com/daemonl/tpt.go/tptobjects"
)

// User wraps an oAuth-ish token which can be used for user authenticated API
// calls. It extends RequestBuilder for custom http calls, and wraps some of
// the API calls to the TPT API
type User struct {
	Token  string
	Client *Client
}

// NewRequest returns a request builder which is pre-populated with
// authentication headers for the user and application
func (u *User) NewRequest(reqPath string) *Request {
	if len(u.Token) < 1 {
		return &Request{
			firstError: fmt.Errorf("User has no token"),
		}
	}
	return u.Client.NewRequest(reqPath).AddHeader("User-Token", u.Token)
}

/////////////////////////
// Wrapped API Methods //
/////////////////////////

// GetAccountDetails returns the user’s account details.
func (u *User) GetAccountDetails() (*tptobjects.UserAccountDetails, error) {
	resp := &tptobjects.UserAccountDetails{}
	err := u.NewRequest("/v1/user/account").DecodeInto(resp)
	return resp, err
}
