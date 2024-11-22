package server

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
)

type Server struct {
	client   *http.Client
	mux      *http.ServeMux
	port     string
	host     string
	Services []service `json:"services"`
}

func NewServer(host, port string) *Server {
	return &Server{
		host:     host,
		port:     port,
		Services: make([]service, 0),
		client:   &http.Client{},
		mux:      http.NewServeMux(),
	}
}

func (s *Server) addService(url *url.URL, name, endpoint string) error {
	se, err := newService(url, name, endpoint)
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Adding service [%s - %s]", se.Name, se.URL.String()))

	s.Services = append(s.Services, *se)

	s.mux.HandleFunc(se.Endpoint, se.handleProxy())
	return nil
}

func (s *Server) listServicesHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info(fmt.Sprintf("[SERVER] Listing services: %v", s.Services))

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(s.Services); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding services: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) addServiceHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Adding services")

		var body map[string]any

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, fmt.Sprintf("Error parsing service: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		urlStr, _ := body["url"].(string)
		name, _ := body["name"].(string)
		endpoint, _ := body["endpoint"].(string)

		parsedUrl, err := url.Parse(urlStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error parsing url: %v", err), http.StatusInternalServerError)
			return
		}

		err = s.addService(parsedUrl, name, endpoint)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error adding service: %v", err), http.StatusInternalServerError)
			return
		}

		if err = json.NewEncoder(w).Encode(body); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding service: %v", err), http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) Run() {
	s.mux.HandleFunc("GET /services", s.listServicesHandler())
	s.mux.HandleFunc("POST /services", s.addServiceHandler())

	log.Fatal(http.ListenAndServe(s.host+":"+s.port, s.mux))
}
