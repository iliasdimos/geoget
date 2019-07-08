package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/oschwald/geoip2-golang"
	"github.com/sirupsen/logrus"
)

var (
	ErrInvalidIP = errors.New("Invalid ip address")
	ErrNotFound  = errors.New("Not Found")
)

type GeoServer struct {
	db  Database
	log *logrus.Logger
}

func NewGeoServer(log *logrus.Logger, db Database) *GeoServer {
	return &GeoServer{db: db, log: log}
}

// Chain applies middlewares to a http.HandlerFunc
func Chain(f http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

func (g *GeoServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Error(w, http.StatusMethodNotAllowed, fmt.Errorf("Method %s not allowed", r.Method))
		return
	}

	in := r.URL.Path[len("/"):]
	ip, err := ipHelper(in)

	if err != nil {
		Error(w, http.StatusNotFound, err)
		return
	}
	data, err := g.getData(ip)

	if err != nil {
		Error(w, http.StatusNotFound, err)
		return
	}

	Json(w, data)
}

func ipHelper(in string) (net.IP, error) {
	if in == "" {
		return nil, ErrInvalidIP
	}

	ip := net.ParseIP(in)
	if ip == nil {
		return nil, ErrInvalidIP
	}
	return ip, nil
}
func (g *GeoServer) getData(ip net.IP) (*geoip2.City, error) {
	data, err := g.db.City(ip)
	if err != nil {
		return nil, ErrNotFound
	}
	return data, nil
}

func Json(w http.ResponseWriter, data *geoip2.City) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&data)
}

type ErrorStruct struct {
	Data string `json:"data"`
}

func Error(w http.ResponseWriter, code int, err error) {
	resp := new(ErrorStruct)
	resp.Data = err.Error()
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(&resp)
}

type Database interface {
	City(net.IP) (*geoip2.City, error)
}
