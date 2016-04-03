package tpt

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

	"github.com/daemonl/tpt.go/tptobjects"
)

// BearerToken is given to an API client to authenticate the
// application
type BearerToken struct {
	Token  string `json:"bearer"`
	Expiry int64  `json:"expiry"`
}

// Config represents the url and authentication details for
// connecting to the TPT API
type Config struct {
	Endpoint     string `json:"endpoint"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

// Client is the root of the connection to the TPT API
type Client struct {
	BaseURL *url.URL
	sync.RWMutex
	Config
	*BearerToken
}

// NewClient builds the default client
func NewClient(config Config) (*Client, error) {
	u, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	return &Client{
		Config:  config,
		BaseURL: u,
	}, nil
}

// New starts a new builder chain
func (c *Client) NewRequest(reqPath string) *Request {
	return NewRequest(*c.BaseURL, reqPath).
		AddHeader("Authorization", "Bearer "+c.BearerToken.Token)

}

// OAuth fetches a new Bearer token from the configured credentials, and sets
// up the client's TokenRequestBuilder
// TODO: This should be automatic, and handle expiry automatically.
func (c *Client) OAuth() error {
	c.Lock()
	defer c.Unlock()

	reqBodyBuf := &bytes.Buffer{}
	if err := json.NewEncoder(reqBodyBuf).Encode(&struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
	}); err != nil {
		return err
	}

	resp, err := http.Post(
		c.Endpoint+"/v1/oauth/token",
		"application/json", reqBodyBuf)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	token := &BearerToken{}
	if err := json.NewDecoder(resp.Body).Decode(token); err != nil {
		return err
	}

	c.BearerToken = token

	return nil
}

// ExchangeUserCode returns gets a user_token in eschange for the code given at
// the end of the client oAuth2 flow. It returns a User object which can be
// used for user authenticated API calls
func (c *Client) ExchangeUserCode(code string) (*User, error) {
	respBody := &struct {
		Token string `json:"user_token"`
	}{}

	err := c.NewRequest("/v1/user/oauth/token").PostJSON(map[string]string{
		"code": code,
	}).DecodeInto(respBody)

	if err != nil {
		return nil, err
	}

	return c.User(respBody.Token), nil
}

// User returns a User object from an oAuth-ish token which can be used for
// user authenticated API calls
func (c *Client) User(token string) *User {
	return &User{
		Token:  token,
		Client: c,
	}
}

/////////////////////////
// Wrapped API Methods //
/////////////////////////

func (c *Client) GetNews(symbol string) (*tptobjects.NewsResponse, error) {
	resp := &tptobjects.NewsResponse{}
	err := c.NewRequest("/v1/news").
		AddQuery("symbol", symbol).
		DecodeInto(resp)
	return resp, err
}
