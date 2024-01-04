package server

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"

	termfreq "github.com/soramon0/kirlia/pkg/term_freq"
)

type api struct {
	*http.Server
	tfIndex *termfreq.TermFreqIndex
	templ   *template.Template
}

func NewServer(addr string, tfIndex *termfreq.TermFreqIndex) (*api, error) {
	if tfIndex == nil || len(*tfIndex) == 0 {
		return nil, fmt.Errorf("error: index cannot be empty")
	}

	t, err := template.ParseFiles("resources/index.html")
	if err != nil {
		return nil, err
	}

	api := &api{
		Server:  &http.Server{Addr: addr},
		tfIndex: tfIndex,
		templ:   t,
	}

	api.Server.Handler = api.newMux()

	return api, nil
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

	var buf bytes.Buffer
	if err := a.templ.Execute(&buf, nil); err != nil {
		fmt.Printf("error: %s\n", err)
		w.WriteHeader(500)
		fmt.Fprint(w, "<h1>Internal Server Error</h1")
		return
	}
	io.Copy(w, &buf)
}

func (a *api) notFound(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(404)
	fmt.Fprint(w, "<h1>Page not found</h1>")
}
