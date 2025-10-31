package transport

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"time"
)

type Kitty struct {
	Name string `json:"name"`
}

func Router() http.Handler {
	r := mux.NewRouter()
	s := r.PathPrefix("/api/v1").Subrouter()

	s.HandleFunc("/hello-world", helloWorld).Methods(http.MethodGet)
	s.HandleFunc("/kitty", getKitty).Methods(http.MethodGet)
	s.HandleFunc("/order/{ID}", order).Methods(http.MethodGet)
	return logMiddleware(r)
}

func helloWorld(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{
			"method":     r.Method,
			"time":       time.Now(),
			"url":        r.URL,
			"remoteAddr": r.RemoteAddr,
			"userAgent":  r.UserAgent(),
		}).Info("got a new request")
		h.ServeHTTP(w, r)
	})
}

func getKitty(w http.ResponseWriter, _ *http.Request) {
	cat := Kitty{"Кот"}
	b, err := json.Marshal(cat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err = io.WriteString(w, string(b)); err != nil {
		log.WithField("err", err).Error("write response error")
	}
}

func order(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["ID"]

	some := r.URL.Query().Get("some")

	if _, err := io.WriteString(w, id+some); err != nil {
		log.WithField("err", err).Error("write response error")
	}
}
