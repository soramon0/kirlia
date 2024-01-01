package server

import (
	"errors"
	"fmt"
	"net/http"

	termfreq "github.com/soramon0/kirlia/pkg/term_freq"
)

type api struct {
	*http.Server
	tfIndex *termfreq.TermFreqIndex
}

func NewServer(addr string, tfIndex *termfreq.TermFreqIndex) *api {
	api := &api{
		Server:  &http.Server{Addr: addr},
		tfIndex: tfIndex,
	}

	api.Server.Handler = api.newMux()

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

func (a *api) newMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", a.serveHomePage)
	mux.HandleFunc("/api/search", a.searchPage)

	return mux
}

func (a *api) serveHomePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.notFound(w)
		return
	}

	path := r.URL.Path
	if path != "/" && path != "/index.html" {
		a.notFound(w)
		return
	}

	w.Header().Add("Content-Type", "text/html")
	fmt.Fprint(w, "<h1>Hello world</h1>")
}

func (a *api) searchPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Add("Content-Type", "text/html")
		fmt.Fprint(w, "<h1>Query Index</h1>")
	}
}

func (a *api) notFound(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(404)
	fmt.Fprint(w, "<h1>Page not found</h1>")
}
