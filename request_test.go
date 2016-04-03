package tpt

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var baseURL url.URL

func init() {
	baseURLPtr, _ := url.Parse("http://localhost:1234/base")
	baseURL = *baseURLPtr
}

func cmp(t *testing.T, name string, got, expect interface{}) {
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

func TestSomeComplexity(t *testing.T) {
	req := NewRequest(baseURL, "/path/one/two")
	cmp(t, "3 Request.URL", req.Request.URL.String(), "http://localhost:1234/base/path/one/two")
	req.AddQuery("q1", "qv1")
	cmp(t, "3 Request.URL added", req.Request.URL.String(), "http://localhost:1234/base/path/one/two?q1=qv1")
	req.AddHeader("h1", "hv1")
	cmp(t, "3 Header", req.Request.Header.Get("h1"), "hv1")
}

type WontEncode struct {
}

func (wo *WontEncode) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("ERROR 12345")
}

func TestErrors(t *testing.T) {
	r := WontEncode{}
	_, err := NewRequest(baseURL, "/path").PostJSON(&r).RawResponse()
	if err == nil {
		t.Error("Should be an error encoding recursive JSON")
	}
	if !strings.Contains(err.Error(), "ERROR 12345") {
		t.Errorf("Not a JSON error: %s", err.Error())
	}
}

func TestSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{"key":"value"}`))
	}))
	defer ts.Close()
	baseURL, _ := url.Parse(ts.URL)

	val := &struct {
		Key string
	}{}
	err := NewRequest(*baseURL, "/").DecodeInto(val)
	if err != nil {
		t.Error(err.Error())
	}
	cmp(t, "TestSuccess: Response Key", val.Key, "value")

}
