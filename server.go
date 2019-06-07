package main

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/oschwald/geoip2-golang"
	"github.com/sirupsen/logrus"
)

func newServer(log *logrus.Logger, db *geoip2.Reader, mw ...func(http.Handler) http.Handler) *chi.Mux {
	mux := chi.NewRouter()
	mux.Use(mw...)
	mux.Route("/", func(r chi.Router) {
		r.Get("/{ip}", byIP(db))
	})

	return mux
}

type value struct {
	Status  int         `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func errorR(w http.ResponseWriter, code int, message string) {
	sendResponse(w, code, value{Status: code, Message: message}) //, nil)
}

func jsonR(w http.ResponseWriter, code int, payload interface{}) {
	sendResponse(w, code, value{Status: code, Data: payload})
}

func sendResponse(w http.ResponseWriter, code int, payload interface{}) {
	resp, err := json.Marshal(payload)
	if err != nil {
		logrus.Println(err.Error())
		errorR(w, http.StatusInternalServerError, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(resp)
}
func byIP(db *geoip2.Reader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ipFromReq := chi.URLParam(r, "ip")
		if ipFromReq == "" {
			errorR(w, 404, "Invalid Request")
			return
		}

		ip := net.ParseIP(ipFromReq)
		if ip == nil {
			errorR(w, 404, "Invalid IP Address")
			return
		}

		record, err := db.City(ip)
		if err != nil {
			errorR(w, 404, "Not Found")
			return
		}

		jsonR(w, http.StatusOK, record)
	}
}
