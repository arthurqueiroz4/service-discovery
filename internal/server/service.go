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
	ttl      time.Time
	url      *url.URL
	proxy    *httputil.ReverseProxy
	name     string
	endpoint string
}

func newService(url *url.URL, name, endpoint string) (*service, error) {
	se := &service{
		url:      url,
		name:     name,
		endpoint: endpoint,
		ttl:      time.Now().Add(30 * time.Minute),
		proxy:    httputil.NewSingleHostReverseProxy(url),
	}
	err := se.verifyHealthy()
	if err != nil {
		return nil, err
	}

	return se, nil
}

func (se *service) verifyHealthy() error {
	c := http.Client{}

	res, err := c.Get(se.url.String())
	if err != nil {
		slog.Error(fmt.Sprintf("Service [%s - %s] isn't healthy: %s", se.name, se.url.String(), err.Error()))
		return err
	}

	if res.StatusCode != http.StatusOK {
		err = errors.New("Health returns status code != 200: " + se.url.String())
		slog.Error(fmt.Sprintf("Service [%s - %s] isn't healthy: %s" + err.Error()))
		return err
	}

	slog.Info(fmt.Sprintf("Service [%s - %s]", se.name, se.url.String()))
	return nil
}

func (s *service) HandleProxy() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Handle ttl
		slog.Info(fmt.Sprintf("[PROXY] Request received at %s at %s", r.URL, time.Now().UTC()))
		r.URL.Host = s.url.Host
		r.URL.Scheme = s.url.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = s.url.Host

		path := r.URL.Path
		r.URL.Path = strings.TrimLeft(path, s.endpoint)

		slog.Info(fmt.Sprintf("[PROXY] Proxying request to %s at %s", r.URL, time.Now().UTC()))
		s.proxy.ServeHTTP(w, r)
	}
}
