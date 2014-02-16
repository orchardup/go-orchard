package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/orchardup/go-orchard/constants"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Host struct {
	ID         string
	Name       string
	URL        string
	Size       int64
	IPAddress  string `json:"ipv4_address"`
	ClientKey  string `json:"client_key"`
	ClientCert string `json:"client_cert"`
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

func (client *HTTPClient) GetHosts() ([]*Host, error) {
	req, err := http.NewRequest("GET", client.BaseURL+"/hosts", nil)
	if err != nil {
		return nil, err
	}

	var hosts []*Host
	if err := client.DoRequest(req, &hosts); err != nil {
		return nil, err
	}
	return hosts, nil
}

func (client *HTTPClient) GetHost(name string) (*Host, error) {
	req, err := http.NewRequest("GET", client.BaseURL+"/hosts/"+name, nil)
	if err != nil {
		return nil, err
	}
	var host Host
	if err := client.DoRequest(req, &host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (client *HTTPClient) CreateHost(name string, ramInMB int) (*Host, error) {
	v := make(map[string]interface{})
	v["name"] = name
	v["size"] = ramInMB
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", client.BaseURL+"/hosts", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var host Host
	if err := client.DoRequest(req, &host); err != nil {
		return nil, err
	}
	return &host, nil
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("orchard/%s", constants.Version))
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	if err := DecodeResponse(resp, v); err != nil {
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
		explanation := string(body)
		var jsonError map[string]interface{}

		if err := json.Unmarshal(body, &jsonError); err == nil {
			if jsonError["detail"] != nil {
				explanation = jsonError["detail"].(string)
			}
		}

		return fmt.Errorf("The Orchard API returned an error: %s", explanation)
	}

	if v != nil {
		if err := json.Unmarshal(body, &v); err != nil {
			return err
		}
	}

	return nil
}
