package util

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
)

// apiClient is built on first use rather than at init time: viper is only
// populated by cobra.OnInitialize, which runs after package initialization.
// Sharing one client also keeps connections alive across the transaction loop
// in 'txlog build'.
var apiClient = sync.OnceValue(func() *resty.Client {
	return resty.New().
		SetTimeout(30 * time.Second).
		SetBaseURL(viper.GetString("server.url"))
})

// APIRequest returns a request against the txlog server, with the JSON content
// type and authentication already configured. Paths passed to it are relative
// to server.url.
func APIRequest() *resty.Request {
	request := apiClient().R().SetHeader("Content-Type", "application/json")
	SetAuthentication(request)
	return request
}

// CheckStatus turns a non-200 response into an error describing it.
func CheckStatus(response *resty.Response) error {
	if response.StatusCode() != 200 {
		return fmt.Errorf("server returned status code %d: %s", response.StatusCode(), response.String())
	}
	return nil
}
