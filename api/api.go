package api

import "fmt"
import "errors"
import "strings"
import "io/ioutil"
import "net/url"
import "net/http"
import "encoding/json"

type Host struct {
	ID   string
	Name string
	URL  string
}

type HTTPClient struct {
	BaseURL string
	Token   string
}

type AuthResponse struct {
	Token string
}

func (client *HTTPClient) GetAuthToken(username string, password string) (string, error) {
	resp, err := http.PostForm(client.BaseURL+"/signin",
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

func (client *HTTPClient) GetHosts() ([]Host, error) {
	var hosts []Host

	req, err := http.NewRequest("GET", client.BaseURL+"/hosts", nil)
	if err != nil {
		return []Host{}, err
	}
	if err := client.DoRequest(req, &hosts); err != nil {
		return []Host{}, err
	}

	return hosts, nil
}

func (client *HTTPClient) CreateHost(name string) (Host, error) {
	var host Host

	v := url.Values{}
	v.Set("name", name)
	req, err := http.NewRequest("POST", client.BaseURL+"/hosts", strings.NewReader(v.Encode()))
	if err != nil {
		return Host{}, err
	}
	if err := client.DoRequest(req, &host); err != nil {
		return Host{}, err
	}

	return host, nil
}

func (client *HTTPClient) DeleteHost(name string) error {
	req, err := http.NewRequest("DELETE", client.BaseURL+"/hosts/"+name, nil)
	if err != nil {
		return err
	}
	if err := client.DoRequest(req, nil); err != nil {
		return err
	}

	return nil
}

func (client *HTTPClient) DoRequest(req *http.Request, v interface{}) error {
	cl := &http.Client{}
	req.Header.Set("Authorization", "Token "+client.Token)
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	if v != nil {
		if err := DecodeResponse(resp, &v); err != nil {
			return err
		}
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
