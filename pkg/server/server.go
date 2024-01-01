package server

import (
	"errors"
	"fmt"
	"net/http"
)

type api struct {
	*http.Server
}

func NewServer(addr string) *api {
	mux := http.NewServeMux()
	api := &api{
		Server: &http.Server{
			Addr: addr,
		},
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			api.notFound(w)
			return
		}

		path := r.URL.Path
		if path != "/" && path != "/index.html" {
			api.notFound(w)
			return
		}

		w.Header().Add("Content-Type", "text/html")
		fmt.Fprint(w, "<h1>Hello world</h1>")
	})

	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Add("Content-Type", "text/html")
			fmt.Fprint(w, "<h1>Query Index</h1>")
		}
	})

	api.Server.Handler = mux
	return api
}

func (a *api) Serve() error {
	host := a.Addr
	if host[0] == ':' {
		host = "0.0.0.0" + host
	}
	fmt.Printf("- Server listening at http://%s\n", host)

	if err := a.Server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			fmt.Println("Server closed")
			return nil
		}

		return fmt.Errorf("error: failed to start server. %s", err)
	}

	return nil
}

func (a *api) notFound(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(404)
	fmt.Fprint(w, "<h1>Page not found</h1>")
}
