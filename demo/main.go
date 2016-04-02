package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"

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

	news, err := c.GetNews("AAPL")
	if err != nil {
		return err
	}
	fmt.Println(news)

	user, err := oauthClientFlow(c)

	accDetails, err := user.GetAccountDetails()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(accDetails)

	return nil
}

func oauthClientFlow(c *tpt.Client) (*tpt.User, error) {

	userChan := make(chan *tpt.User)
	errChan := make(chan error)

	s := &http.Server{
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			code := req.URL.Query().Get("code")
			if len(code) < 1 {
				query := &url.Values{
					"client_id":    {c.Config.ClientID},
					"redirect_uri": {"http://localhost:8080/oauth"},
				}
				authorizeUrl := c.Config.Endpoint + "/v1/user/oauth/authorize?" + query.Encode()

				http.Redirect(rw, req, authorizeUrl, http.StatusTemporaryRedirect)
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
			userChan <- user
			rw.Write([]byte("OK - You can close the browser now"))
		}),
	}

	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		return nil, err
	}
	defer l.Close()

	go func() {
		err = s.Serve(l)
		errChan <- err
	}()

	fmt.Println("Visit http://localhost:8080 to authorize the application")
	fmt.Println("Attempting to do it for you...")

	// Cross Platform? These will exit one way or another
	exec.Command("start", "http://localhost:8080").Start()
	exec.Command("open", "http://localhost:8080").Start()
	exec.Command("xdg-open", "http://localhost:8080").Start()

	select {
	case err := <-errChan:
		return nil, err
	case user := <-userChan:
		return user, nil
	}
}
