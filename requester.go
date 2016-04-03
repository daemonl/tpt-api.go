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

// Headers is a Modifier which adds to the request headers
type Headers map[string]string

// Modify implements the Modifier interface
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
func (rb *RequestBuilder) WithModifier(mod Modifier) *RequestBuilder {
	n := &RequestBuilder{
		BaseURL:   rb.BaseURL,
		Modifiers: append(rb.Modifiers, mod),
	}
	return n
}

// New starts a new builder chain
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

// Request is a chainable request being built
type Request struct {
	*http.Request
	gotError error
}

// err sets the internal error. Only the first error is ever returned, and only
// at the end of the build process. This makes the chaining possible and clean,
// otherwise each step would have to be checked for its own error.
// Consequently, while this is 'prettry cool' to work with, it's not really
// idiomatic go.
func (req *Request) err(err error) {
	if req.gotError == nil {
		req.gotError = err
	}
}

// Query merges provided querystring parameters into the request
func (req *Request) Query(query *url.Values) *Request {
	for k, vals := range *query {
		for _, v := range vals {
			req.Request.URL.Query().Add(k, v)
		}
	}
	return req
}

// PostJSON adds a request reader for the JSON encoding of the provided body,
// sets the method to POST, and adds an application/json content type
func (req *Request) PostJSON(body interface{}) *Request {
	req.Request.Method = "POST"
	bodyBytes := &bytes.Buffer{}
	err := json.NewEncoder(bodyBytes).Encode(body)
	if err != nil {
		req.err(err)
		return req
	}
	req.Header.Add("Content-Type", "application/json")
	req.Body = ioutil.NopCloser(bodyBytes)
	return req
}

// RawResponse performs the request, and just returns the response. It is the
// last point of the chaining, but is wrapped by other methods as well, so may
// not be called directly
func (req *Request) RawResponse() (*http.Response, error) {
	if req.gotError != nil {
		return nil, req.gotError
	}

	return http.DefaultClient.Do(req.Request)
}

// DecodeInto is an extension of RawResponse, which decodes a JSON body into
// the provided object. Just for fun, it also sets the Accept header to
// application/json,
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
