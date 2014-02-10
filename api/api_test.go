package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetHosts(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hosts" {
			t.Errorf("expected HTTP request to /hosts, got %s", r.URL.Path)
		}

		fmt.Fprintln(w, `[
      {
        "id": "14dff6d8-3b9a-41be-9ffd-d0d054a17492",
        "url": "http://dummy/hosts/default_bfirsh",
        "name": "default_bfirsh"
      }
    ]`)
	}))
	defer ts.Close()

	client := HTTPClient{ts.URL, "dummy_token"}

	hosts, err := client.GetHosts()
	if err != nil {
		t.Error(err)
	}
	if len(hosts) != 1 {
		t.Errorf("expected 1 element, got %d (hosts: %v)", len(hosts), hosts)
	}
	if hosts[0].Name != "default_bfirsh" {
		t.Errorf("expected default_bfirsh, got %s (hosts: %v)", hosts[0].Name, hosts)
	}
}

func TestCreateHost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/hosts" {
			t.Errorf("expected HTTP request to /hosts, got %s", r.URL.Path)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json, got %s", r.Header.Get("Content-Type"))
		}

		body, _ := ioutil.ReadAll(r.Body)
		var data map[string]string
		json.Unmarshal(body, &data)

		if data["name"] != "newhost" {
			t.Errorf("expected 'newhost', got '%s'", data["name"])
		}

		w.WriteHeader(201)
		fmt.Fprintln(w, `{
      "id": "14dff6d8-3b9a-41be-9ffd-d0d054a17492",
      "url": "http://dummy/hosts/newhost",
      "name": "newhost"
    }`)
	}))

	client := HTTPClient{ts.URL, "dummy_token"}

	host, err := client.CreateHost("newhost")
	if err != nil {
		t.Error(err)
	}
	if host.Name != "newhost" {
		t.Errorf("expected 'newhost', got '%s' (host: %v)", host.Name, host)
	}
}

func TestDeleteHost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/hosts/myhost" {
			t.Errorf("expected DELETE request to /hosts/myhost, got %s request to %s", r.Method, r.URL.Path)
		}

		fmt.Fprintln(w, "")
	}))

	client := HTTPClient{ts.URL, "dummy_token"}

	err := client.DeleteHost("myhost")
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteHostError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		fmt.Fprintln(w, "I broke :(")
	}))

	client := HTTPClient{ts.URL, "dummy_token"}

	err := client.DeleteHost("myhost")
	if err == nil {
		t.Error("expected DeleteHost() to return an error")
	}
}
