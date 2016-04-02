package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/daemonl/tpt.go"
)

func main() {
	err := do()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(127)
		return
	}
}

func loadConfig(filename string) (tpt.Config, error) {
	devConfig := tpt.Config{}
	file, err := os.Open(filename)
	if err != nil {
		return devConfig, err
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&devConfig)
	return devConfig, nil
}

func do() error {

	config, err := loadConfig("config.json")
	if err != nil {
		return err
	}
	c, err := tpt.NewClient(config)
	if err != nil {
		return err
	}
	err = c.OAuth()
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", *c.BearerToken)

	r, err := c.New("/v1/news").Query(&url.Values{"symbol": {"AAPL"}}).RawResponse()
	if err != nil {
		return err
	}
	b, _ := httputil.DumpResponse(r, true)
	fmt.Println(string(b))

	http.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {

		query := &url.Values{
			"client_id":    {config.ClientID},
			"redirect_uri": {"http://localhost:8080/oauth"},
		}
		authorizeUrl := config.Endpoint + "/v1/user/oauth/authorize?" + query.Encode()

		rw.Header().Add("Content-Type", "text/html")
		fmt.Fprintf(rw, `<!DOCTYPE html>
		<html>
			<head></head>
			<body>
				<a href="%s">Link</a>
			</body>
		</html>`, authorizeUrl)
	})

	http.HandleFunc("/oauth", func(rw http.ResponseWriter, req *http.Request) {
		code := req.URL.Query().Get("code")
		if len(code) < 1 {
			http.Redirect(rw, req, "/", http.StatusTemporaryRedirect)
			return
		}
		user, err := c.ExchangeUserCode(code)
		if err != nil {
			fmt.Println(err.Error())
			http.Redirect(rw, req, "/", http.StatusTemporaryRedirect)
			return
		}
		// TODO: Store This
		fmt.Printf("User Access Token: %s\n", user.Token)

		accDetails, err := user.GetAccountDetails()
		if err != nil {
			fmt.Println(err.Error())
			http.Error(rw, err.Error(), 500)
			return
		}
		fmt.Println(accDetails)
		json.NewEncoder(rw).Encode(accDetails)

	})

	http.ListenAndServe(":8080", nil)

	return nil
}
