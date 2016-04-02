package main

import (
	"encoding/json"
	"fmt"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/daemonl/tpt/api"
)

func main() {
	err := do()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(127)
		return
	}
}

func loadConfig(filename string) (api.Config, error) {
	devConfig := api.Config{}
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
	c, err := api.NewClient(config)
	if err != nil {
		return err
	}
	err = c.OAuth()
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", *c.BearerToken)

	r, err := c.Get("/v1/news", url.Values{"symbol": {"AAPL"}})
	if err != nil {
		return err
	}
	b, _ := httputil.DumpResponse(r, true)
	fmt.Println(string(b))

	return nil
}
