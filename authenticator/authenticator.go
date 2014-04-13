package authenticator

import (
	"crypto/md5"
	"fmt"
	"github.com/orchardup/go-orchard/api"
	"github.com/orchardup/go-orchard/vendor/code.google.com/p/gopass"
	"io"
	"io/ioutil"
	"os"
	"path"
)

func Authenticate() (*api.HTTPClient, error) {
	httpClient := api.HTTPClient{GetAPIURL(), ""}
	err := PopulateToken(&httpClient)
	if err != nil {
		return nil, err
	}
	return &httpClient, nil
}

func PopulateToken(httpClient *api.HTTPClient) error {
	envVar := os.Getenv("ORCHARD_API_TOKEN")
	if envVar != "" {
		httpClient.Token = envVar
		return nil
	}

	tokenFile, err := GetTokenFilePath(httpClient.BaseURL)
	if err != nil {
		return err
	}

	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		token, err := GetTokenByPromptingUser(*httpClient)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(tokenFile, []byte(token), 0644); err != nil {
			return err
		}
		httpClient.Token = token
	} else {
		token, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return err
		}

		httpClient.Token = string(token)
	}

	return nil
}

func GetAPIURL() string {
	apiURL := os.Getenv("ORCHARD_API_URL")

	if apiURL == "" {
		apiURL = "https://api.orchardup.com/v2"
	}

	return apiURL
}

func GetTokenFilePath(baseURL string) (string, error) {
	tokenDir, err := GetTokenDir()
	if err != nil {
		return "", err
	}

	// HACK: API URL used to be orchard.com/api, don't invalidate those
	// tokens
	if baseURL == "https://api.orchardup.com/v2" {
		baseURL = "https://orchardup.com/api/v2"
	}

	h := md5.New()
	io.WriteString(h, baseURL)
	hash := fmt.Sprintf("%x", h.Sum(nil))

	return path.Join(tokenDir, hash), nil
}

func GetTokenDir() (string, error) {
	tokenDir := path.Join(os.Getenv("HOME"), ".orchard", "api_tokens")
	err := os.MkdirAll(tokenDir, 0700)
	if err != nil {
		return "", err
	}
	return tokenDir, nil
}

func GetTokenByPromptingUser(httpClient api.HTTPClient) (string, error) {
	username, password := Prompt()

	token, err := httpClient.GetAuthToken(username, password)
	if err != nil {
		return "", err
	}

	return token, nil
}

func Prompt() (string, string) {
	var (
		username string
		password string
	)
	fmt.Print("Orchard username: ")
	fmt.Scanln(&username)
	password, _ = gopass.GetPass("Password: ")
	return username, password
}
