package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type Server struct {
	client   *http.Client
	mux      *http.ServeMux
	port     string
	host     string
	services []service
}

func NewServer(host, port string) *Server {
	return &Server{
		host:     host,
		port:     port,
		services: make([]service, 10),
		client:   &http.Client{},
		mux:      http.NewServeMux(),
	}
}

func (s *Server) AddService(url *url.URL, name, endpoint string) error {
	se, err := newService(url, name, endpoint)
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("Adding service [%s - %s]", se.name, se.url.String()))

	s.services = append(s.services, *se)
	return nil
}

func (s *Server) Run() error {
	for _, svc := range s.services {
		s.mux.HandleFunc(svc.endpoint, svc.HandleProxy())
	}

	if err := http.ListenAndServe(s.host+":"+s.port, s.mux); err != nil {
		slog.Error("could not start the server: " + err.Error())
		return fmt.Errorf("could not start the server: %v", err)

	}
	return nil
}
