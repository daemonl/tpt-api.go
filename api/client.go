package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"
)

type Config struct {
	Endpoint     string
	ClientID     string
	ClientSecret string
}

type Client struct {
	sync.RWMutex
	Config
	*BearerToken
}

type BearerToken struct {
	Token  string `json:"bearer"`
	Expiry int64  `json:"expiry"`
}

func NewClient(config Config) (*Client, error) {
	return &Client{
		Config: config,
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
	return nil
}

func (c *Client) Get(path string, query url.Values) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.Endpoint+path+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	//req.Header.Add("User-Token")
	req.Header.Add("Authorization", "Bearer "+c.BearerToken.Token)
	return http.DefaultClient.Do(req)
}
