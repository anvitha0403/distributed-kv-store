package main

import (
	"encoding/json"
	"net/http"

	"github.com/anvitha0403/golog/internal/server/kvstore"
	"github.com/gorilla/mux"
)

type keyValue struct {
	Key   string
	Value string
}
type httpServer struct {
	store kvstore.Store
}

func newHTTPServer(config *Config) (*httpServer, error) {
	return &httpServer{
		store: config.Store,
	}, nil
}

func (s *httpServer) putHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var pair keyValue
	if err := json.NewDecoder(r.Body).Decode(&pair); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	op, err := s.store.Put(pair.Key, pair.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if op == kvstore.OPERATION_PUT_UPDATE {
		w.WriteHeader(http.StatusOK)

	} else {
		w.WriteHeader(http.StatusCreated)

	}

}

func (s *httpServer) getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("key")
	value, err := s.store.Get(key)

	if err == kvstore.ErrKeyDoesntExist {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := keyValue{Key: key, Value: value}

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func (s *httpServer) deleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("key")
	err := s.store.Del(key)

	if err == kvstore.ErrKeyDoesntExist {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func NewHTTPServer(addr string, config *Config) (*http.Server, error) {
	httpsrv, err := newHTTPServer(config)
	if err != nil {
		return nil, err
	}
	r := mux.NewRouter()

	r.HandleFunc("/kv", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {

		case http.MethodPut:
			httpsrv.putHandler(w, r)
		case http.MethodGet:
			httpsrv.getHandler(w, r)
		case http.MethodDelete:
			httpsrv.deleteHandler(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}, nil
}
