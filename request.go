package tpt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"net/http"
	"net/url"
	"path"
)

// Request is a chainable request being built. It supports a very limited
// subset of standard the net/http request, but is sufficient for the TPT API
// Errors in the chain are stored, and only the first is returned at the end of
// the build process. This makes the chaining possible and clean, otherwise
// each step would have to be checked for its own error.  Consequently, while
// this is 'prettry cool' to work with, it's not really idiomatic go.
type Request struct {
	*http.Request
	firstError error
}

// NewRequest builds a default request from a url and path. the base url is
// passed by value to avoid any accidental modification / proove immutability
func NewRequest(base url.URL, reqPath string) *Request {
	base.Path = path.Join(base.Path, reqPath)
	req := &http.Request{
		Method:     "GET", // Default, can be changed by the chainer
		URL:        &base,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
		Host:       base.Host,
	}
	return &Request{
		Request: req,
	}
}

// err sets the internal error. Only the first error is ever returned, see
// comment for Reqeust
func (req *Request) err(err error) {
	if req.firstError == nil {
		req.firstError = err
	}
}

// AddQuery adds a new key and value to the querystring. It uses .Add(), so
// will *not* replace existing values
func (req *Request) AddQuery(key, value string) *Request {
	// This is slower than it needs to be. Parses the query, appends, then
	// re-encodes it
	q := req.Request.URL.Query()
	q.Add(key, value)
	req.Request.URL.RawQuery = q.Encode()
	return req
}

// AddHeader adds a new key and value to the request's headers. It uses .Add(),
// so will *not* replace existing headers
func (req *Request) AddHeader(key string, value string) *Request {
	req.Request.Header.Add(key, value)
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
	req.AddHeader("Content-Type", "application/json")
	req.Body = ioutil.NopCloser(bodyBytes)
	return req
}

// RawResponse performs the request, and just returns the response. It is the
// last point of the chaining, but is wrapped by other methods as well, so may
// not be called directly
func (req *Request) RawResponse() (*http.Response, error) {
	if req.firstError != nil {
		return nil, req.firstError
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
	//log.Printf("API %s %s -> %s\n", req.Method, req.URL.String(), resp.Status)
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(responseInto)
}
