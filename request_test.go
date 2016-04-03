package tpt

import (
	"io/ioutil"
	"net/url"
	"strings"
	"testing"
)

var baseURL url.URL

func init() {
	baseURLPtr, _ := url.Parse("http://localhost:1234/base")
	baseURL = *baseURLPtr
}

func cmp(t *testing.T, name string, expect, got interface{}) {
	if expect != got {
		t.Errorf("%s should be %#v, got %#v", name, expect, got)
	}
}

func TestGetRequest(t *testing.T) {
	req := NewRequest(baseURL, "/path")
	cmp(t, "1 Request.Host", req.Request.Host, "localhost:1234")
	cmp(t, "1 Request.Method", req.Method, "GET")
	cmp(t, "1 Request.URL as a string",
		req.Request.URL.String(),
		"http://localhost:1234/base/path")
	cmp(t, "1 Number of request headers", len(req.Request.Header), int(0))
	cmp(t, "1 Request Body", req.Request.Body, nil)
}

func TestPostRequest(t *testing.T) {
	req := NewRequest(baseURL, "/path").PostJSON("Hello")
	cmp(t, "2 Request.Host", req.Request.Host, "localhost:1234")
	cmp(t, "2 Request.Method", req.Method, "POST")
	cmp(t, "2 Request.URL as a string",
		req.Request.URL.String(),
		"http://localhost:1234/base/path")
	cmp(t, "2 Number of request headers", len(req.Request.Header), int(1))
	cmp(t, "2 Content-Type header", req.Request.Header.Get("Content-Type"), "application/json")
	bbytes, err := ioutil.ReadAll(req.Request.Body)
	if err != nil {
		t.Fatalf("Could not read POST body: %s", err.Error)
	}
	cmp(t, "2 Request body", strings.TrimSpace(string(bbytes)), `"Hello"`)
}
