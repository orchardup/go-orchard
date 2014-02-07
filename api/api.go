package api

import "fmt"
import "errors"
import "strings"
import "io/ioutil"
import "net/url"
import "net/http"
import "encoding/json"

type Client interface {
	GetAuthToken(string, string) (string, error)
	GetHosts(string) ([]Host, error)
	CreateHost(string) (Host, error)
}

type Host struct {
	ID   string
	Name string
	URL  string
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

func (api HTTPClient) GetHosts(token string) ([]Host, error) {
	var hosts []Host

	req, err := http.NewRequest("GET", api.BaseURL+"/hosts", nil)
	if err != nil {
		return []Host{}, err
	}
	if err := DoRequest(req, token, &hosts); err != nil {
		return []Host{}, err
	}

	return hosts, nil
}

func (api HTTPClient) CreateHost(token string, name string) (Host, error) {
	var host Host

	v := url.Values{}
	v.Set("name", name)
	req, err := http.NewRequest("POST", api.BaseURL+"/hosts", strings.NewReader(v.Encode()))
	if err != nil {
		return Host{}, err
	}
	if err := DoRequest(req, token, &host); err != nil {
		return Host{}, err
	}

	return host, nil
}

func DoRequest(req *http.Request, token string, v interface{}) error {
	client := &http.Client{}
	req.Header.Set("Authorization", "Token "+token)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if err := DecodeResponse(resp, &v); err != nil {
		return err
	}
	return nil
}

func DecodeResponse(resp *http.Response, v interface{}) error {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return errors.New(fmt.Sprintf("erroneous API response: %s", body))
	}

	if err := json.Unmarshal(body, &v); err != nil {
		return err
	}

	return nil
}
