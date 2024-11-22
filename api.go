package main

import "net/http"

func main() {
	sm := http.NewServeMux()

	sm.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	sm.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("pong"))
	})

	sm.HandleFunc("/pong", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ping"))
	})

	http.ListenAndServe("localhost:9000", sm)
}
