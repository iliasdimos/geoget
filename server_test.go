package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oschwald/geoip2-golang"
)

type MockDatabase struct {
	store map[string]*geoip2.City
}

func (d MockDatabase) City(ip net.IP) (*geoip2.City, error) {
	c, ok := d.store[ip.String()]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func TestGetCity(t *testing.T) {
	city1 := geoip2.City{
		Country: struct {
			GeoNameID         uint              "maxminddb:\"geoname_id\""
			IsInEuropeanUnion bool              "maxminddb:\"is_in_european_union\""
			IsoCode           string            "maxminddb:\"iso_code\""
			Names             map[string]string "maxminddb:\"names\""
		}{
			IsInEuropeanUnion: false,
		},
	}

	db := MockDatabase{
		map[string]*geoip2.City{
			"8.8.8.8": &city1},
	}

	server := GeoServer{db: db}

	t.Run("Get an address that exists ", func(t *testing.T) {
		request := newGetRequest("8.8.8.8")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusOK)
		if response.Body.String() == "" {
			t.Errorf("response body is wrong, got '%s' want '%s'", response.Body.String(), "")
		}
	})
	t.Run("Get an address that does not exists ", func(t *testing.T) {
		request := newGetRequest("1.1.1.1")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusNotFound)
		if response.Body.String() != "" {
			t.Errorf("response body is wrong, got '%s' want '%s'", response.Body.String(), "")
		}
	})
	t.Run("Sent empty ip", func(t *testing.T) {
		request := newGetRequest("")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusNotFound)
	})
	t.Run("Sent malformed ip", func(t *testing.T) {
		request := newGetRequest("1.1.a.2")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusNotFound)
	})
	t.Run("Sent POST request ", func(t *testing.T) {
		request := newPostRequest("1.1.1.1")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(t, response.Code, http.StatusMethodNotAllowed)
	})
}

func newGetRequest(ip string) *http.Request {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/%s", ip), nil)
	return req
}
func newPostRequest(ip string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/%s", ip), nil)
	return req
}
func assertStatus(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}
