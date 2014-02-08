package main

import "fmt"
import "os"
import "io"
import "io/ioutil"
import "crypto/md5"
import "github.com/orchardup/orchard/api"
import "code.google.com/p/gopass"

func main() {
	httpClient, err := Authenticate()
	if err != nil {
		fmt.Printf("Error authenticating:\n%s\n", err)
	}

	hosts, err := httpClient.GetHosts()
	if err != nil {
		fmt.Printf("Error getting hosts:\n%s\n", err)
		return
	}

	fmt.Printf("hosts: %v\n", hosts)
}

func Authenticate() (*api.HTTPClient, error) {
	httpClient := api.HTTPClient{"http://localdocker:8000/api/v1", ""}

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
		httpClient.Token = string(token)
	}

	return &httpClient, nil
}

func GetTokenFilePath(baseURL string) (string, error) {
	tokenDir, err := GetTokenDir()
	if err != nil {
		return "", err
	}

	h := md5.New()
	io.WriteString(h, baseURL)
	hash := fmt.Sprintf("%x", h.Sum(nil))

	return tokenDir + hash, nil
}

func GetTokenDir() (string, error) {
	tokenDir := os.Getenv("HOME") + "/.orchard/api_tokens/"
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
	fmt.Print("Username you signed up for Orchard with: ")
	fmt.Scanln(&username)
	password, _ = gopass.GetPass("Password: ")
	return username, password
}
