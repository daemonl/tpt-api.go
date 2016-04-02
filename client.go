package tpt

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
)

type BearerToken struct {
	Token  string `json:"bearer"`
	Expiry int64  `json:"expiry"`
}

type Config struct {
	Endpoint     string `json:"endpoint"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

type Client struct {
	sync.RWMutex
	TokenRequestBuilder *RequestBuilder
	*RequestBuilder
	Config
	*BearerToken
}

func NewClient(config Config) (*Client, error) {
	u, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	return &Client{
		Config:              config,
		TokenRequestBuilder: NewRequestBuilder(u),
	}, nil
}

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
	c.RequestBuilder = c.TokenRequestBuilder.WithModifier(
		&Headers{
			"Authorization": "Bearer " + c.BearerToken.Token,
		},
	)

	return nil
}

func (c *Client) ExchangeUserCode(code string) (*User, error) {
	respBody := &struct {
		Token string `json:"user_token"`
	}{}

	err := c.New("/v1/user/oauth/token").PostJSON(map[string]string{
		"code": code,
	}).DecodeInto(respBody)

	if err != nil {
		return nil, err
	}

	return c.User(respBody.Token), nil
}

func (c *Client) User(token string) *User {
	return &User{
		Token: token,
		RequestBuilder: c.RequestBuilder.WithModifier(
			&Headers{"User-Token": token},
		),
	}
}
