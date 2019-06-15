package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/oschwald/geoip2-golang"
	"github.com/sirupsen/logrus"
)

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type GeoServer struct {
	db  Database
	log *logrus.Logger
}

func NewGeoServer(log *logrus.Logger, db Database) *GeoServer {
	return &GeoServer{db: db}
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
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	in := r.URL.Path[len("/"):]
	ip, err := ipHelper(in)

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	g.respondIP(w, ip)
}

var ErrInvalidIP = fmt.Errorf("Invalid ip address")
var ErrNotFound = fmt.Errorf("Not Found")

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

func (g *GeoServer) respondIP(w http.ResponseWriter, ip net.IP) {
	data, err := g.db.City(ip)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(&data)
	w.WriteHeader(http.StatusOK)
}

type Database interface {
	City(net.IP) (*geoip2.City, error)
}
