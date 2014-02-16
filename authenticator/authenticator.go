package authenticator

import (
	"crypto/md5"
	"encoding/json"
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

	tokenFile, err := GetTokenFilePath(httpClient.BaseURL)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		token, err := GetTokenByPromptingUser(&httpClient)
		if err != nil {
			return nil, err
		}
		if err := ioutil.WriteFile(tokenFile, []byte(token), 0644); err != nil {
			return nil, err
		}
		httpClient.Token = token
	} else {
		token, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return nil, err
		}

		if token[0] == "{"[0] {
			var tokenJSON map[string]interface{}
			json.Unmarshal(token, &tokenJSON)
			token = []byte(tokenJSON["token"].(string))
		}

		httpClient.Token = string(token)
	}

	return &httpClient, nil
}

func GetAPIURL() string {
	apiURL := os.Getenv("ORCHARD_API_URL")

	if apiURL == "" {
		apiURL = "https://orchardup.com/api/v2"
	}

	return apiURL
}

func GetTokenFilePath(baseURL string) (string, error) {
	tokenDir, err := GetTokenDir()
	if err != nil {
		return "", err
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

func GetTokenByPromptingUser(httpClient *api.HTTPClient) (string, error) {
	username, password := Prompt()

	token, err := httpClient.GetAuthToken(username, password)
	if err != nil {
		return "", err
	}
	httpClient.Token = token

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
