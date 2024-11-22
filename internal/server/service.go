package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type service struct {
	TTL      time.Time              `json:"ttl"`
	URL      *url.URL               `json:"url"`
	Proxy    *httputil.ReverseProxy `json:"-"`
	Name     string                 `json:"name"`
	Endpoint string                 `json:"endpoint"`
}

func newService(url *url.URL, name, endpoint string) (*service, error) {
	se := &service{
		URL:      url,
		Name:     name,
		Endpoint: endpoint,
		TTL:      time.Now().Add(30 * time.Millisecond),
		Proxy:    httputil.NewSingleHostReverseProxy(url),
	}
	err := se.verifyHealthy()
	if err != nil {
		return nil, err
	}

	return se, nil
}

func (se *service) verifyHealthy() error {
	c := http.Client{}

	res, err := c.Get(se.URL.String() + "/health")
	if err != nil {
		slog.Error(fmt.Sprintf("Service [%s - %s] isn't healthy: %s", se.Name, se.URL.String(), err.Error()))
		return err
	}

	if res.StatusCode != http.StatusOK {
		err = errors.New("Health request returns status code != 200")
		slog.Error(fmt.Sprintf("Service [%s - %s] isn't healthy: %s", se.Name, se.URL.String(), err.Error()))
		return err
	}

	slog.Info(fmt.Sprintf("Service [%s - %s]", se.Name, se.URL.String()))
	return nil
}

func (s *service) handleProxy() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.isValid() {
			if err := s.verifyHealthy(); err != nil {
				w.Write([]byte("Service is down"))

				w.WriteHeader(502)
				return
			}
		}

		slog.Info(fmt.Sprintf("[PROXY] Request received at %s at %s", r.URL, time.Now().UTC()))
		r.URL.Host = s.URL.Host
		r.URL.Scheme = s.URL.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = s.URL.Host

		path := r.URL.Path
		r.URL.Path = strings.TrimLeft(path, s.Endpoint)

		slog.Info(fmt.Sprintf("[PROXY] Proxying request to %s at %s", r.URL, time.Now().UTC()))
		s.Proxy.ServeHTTP(w, r)
	}
}

func (s *service) isValid() bool {
	slog.Info("Validando...")
	return time.Now().Before(s.TTL)
}
