package api

import "fmt"
import "errors"
import "io/ioutil"
import "net/url"
import "net/http"
import "encoding/json"

type Client interface {
	GetAuthToken(string, string) (string, error)
}

type HTTPClient struct {
	BaseURL string
}

type AuthResponse struct {
	Token string
}

func (api HTTPClient) GetAuthToken(username string, password string) (string, error) {
	resp, err := http.PostForm(api.BaseURL+"/signin",
		url.Values{"username": {username}, "password": {password}})

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var authResponse AuthResponse
	if err := DecodeResponse(resp, &authResponse); err != nil {
		return "", err
	}

	return authResponse.Token, nil
}

func DecodeResponse(resp *http.Response, v interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("erroneous API response: %s", body))
	}

	if err := json.Unmarshal(body, &v); err != nil {
		return err
	}

	return nil
}
