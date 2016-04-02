package tpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
)

type Headers map[string]string

func (h *Headers) Modify(req *http.Request) {
	for k, v := range *h {
		req.Header.Add(k, v)
	}
}

// Modifier adds to or changes a request on the way through
type Modifier interface {
	Modify(*http.Request)
}

// RequestBuilder handles client requests after passing through the Modifier
type RequestBuilder struct {
	Modifiers []Modifier
	BaseURL   *url.URL
}

// NewRequestBuilder creates a requester
func NewRequestBuilder(url *url.URL) *RequestBuilder {
	return &RequestBuilder{
		BaseURL:   url,
		Modifiers: []Modifier{},
	}
}

// WithModifier derives a new requester with the given modifier
func (r *RequestBuilder) WithModifier(mod Modifier) *RequestBuilder {
	n := &RequestBuilder{
		BaseURL:   r.BaseURL,
		Modifiers: append(r.Modifiers, mod),
	}
	return n
}

func (rb *RequestBuilder) New(reqPath string) *Request {
	u := *rb.BaseURL
	u.Path = path.Join(u.Path, reqPath)
	req := &http.Request{
		Method:     "GET",
		URL:        &u,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		//Body:       rc,
		Host: u.Host,
	}
	for _, mod := range rb.Modifiers {
		mod.Modify(req)
	}
	return &Request{
		Request: req,
	}
}

type Request struct {
	*http.Request
	gotError error
}

func (r *Request) err(err error) {
	if r.gotError == nil {
		r.gotError = err
	}
}

func (r *Request) Query(query *url.Values) *Request {
	for k, vals := range *query {
		for _, v := range vals {
			r.Request.URL.Query().Add(k, v)
		}
	}
	return r
}

// Post performs a http POST request, writing the body as application/json
func (r *Request) PostJSON(body interface{}) *Request {
	r.Request.Method = "POST"
	bodyBytes := &bytes.Buffer{}
	err := json.NewEncoder(bodyBytes).Encode(body)
	if err != nil {
		r.err(err)
		return r
	}
	r.Header.Add("Content-Type", "application/json")
	r.Body = ioutil.NopCloser(bodyBytes)
	return r
}

func (r *Request) RawResponse() (*http.Response, error) {
	if r.gotError != nil {
		return nil, r.gotError
	}

	return http.DefaultClient.Do(r.Request)
}

func (req *Request) DecodeInto(responseInto interface{}) error {
	req.Header.Add("Accept", "application/json")

	resp, err := req.RawResponse()
	if err != nil {
		return err
	}
	log.Printf("API %s %s -> %s\n", req.Method, req.URL.String(), resp.Status)
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(responseInto)
}
