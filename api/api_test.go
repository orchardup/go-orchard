package api

import "testing"
import "fmt"
import "io/ioutil"
import "net/url"
import "net/http"
import "net/http/httptest"

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

		body, _ := ioutil.ReadAll(r.Body)
		v, _ := url.ParseQuery(string(body))

		if v.Get("name") != "newhost" {
			t.Errorf("expected 'newhost', got '%s'", v.Get("name"))
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
