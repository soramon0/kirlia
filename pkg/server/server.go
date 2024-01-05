package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"

	termfreq "github.com/soramon0/kirlia/pkg/term_freq"
)

type api struct {
	*http.Server
	tfIndex  *termfreq.TermFreqIndex
	pageHome *template.Template
	page404  *template.Template
	page500  *template.Template
}

func NewServer(addr string, tfIndex *termfreq.TermFreqIndex) (*api, error) {
	if tfIndex == nil || len(*tfIndex) == 0 {
		return nil, fmt.Errorf("error: index cannot be empty")
	}

	root, err := getAbsRootPath("resources")
	if err != nil {
		return nil, err
	}
	t, err := template.ParseFiles(root + "/index.html")
	if err != nil {
		return nil, err
	}
	t404, err := template.ParseFiles(root + "/404.html")
	if err != nil {
		return nil, err
	}
	t500, err := template.ParseFiles(root + "/500.html")
	if err != nil {
		return nil, err
	}

	api := &api{
		Server:   &http.Server{Addr: addr},
		tfIndex:  tfIndex,
		pageHome: t,
		page404:  t404,
		page500:  t500,
	}

	api.Server.Handler = logRequestHandler(api.newMux())

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
	mux.HandleFunc("/api/search", a.searchDocuments)

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

	a.render(w, http.StatusOK, func(buf *bytes.Buffer) error {
		return a.pageHome.Execute(buf, nil)
	})
}

func findTermFreq(term string, doc *termfreq.TermFreq) float32 {
	freq, ok := (*doc)[term]
	if !ok {
		return 0.0
	}

	var docFreqSum uint = 0
	for _, f := range *doc {
		docFreqSum += f
	}

	return float32(freq) / float32(docFreqSum)
}

func findDocsFreq(terms string, docs *termfreq.TermFreqIndex) map[string]float32 {
	data := make(map[string]float32, 0)

	for filename, doc := range *docs {
		fmt.Println("Searching", filename)
		l := termfreq.NewLexer(terms)
		term := l.NextToken()
		var total_freq float32 = 0.0
		for {
			if term == nil {
				fmt.Println()
				break
			}
			freq := findTermFreq(*term, &doc)
			total_freq += freq
			fmt.Printf("Term %s => %f\n", *term, freq)

			term = l.NextToken()
		}

		data[filename] = total_freq
	}

	return data
}

func (a *api) searchDocuments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		a.jsonResponse(w, http.StatusBadRequest, apiResponse{Msg: "invalid request method"})
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		a.jsonResponse(w, http.StatusBadRequest, apiResponse{Msg: "q query param is required"})
		return
	}

	data := findDocsFreq(query, a.tfIndex)
	res := apiResponse{Data: data, Msg: fmt.Sprintf("Searched %d files", len(*a.tfIndex))}
	a.jsonResponse(w, http.StatusOK, &res)
}

func (a *api) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Add("Content-Type", "application/json")

	result, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(apiResponse{Msg: "Failed to render json"})
		return
	}

	w.WriteHeader(status)
	w.Write(result)
}

func (a *api) notFound(w http.ResponseWriter) {
	a.render(w, http.StatusNotFound, func(buf *bytes.Buffer) error {
		return a.page404.Execute(buf, nil)
	})
}

func (a *api) internalErr(w http.ResponseWriter) {
	a.render(w, http.StatusInternalServerError, func(buf *bytes.Buffer) error {
		return a.page500.Execute(buf, nil)
	})
}

func (a *api) render(
	w http.ResponseWriter,
	status int,
	executeTempl func(buf *bytes.Buffer) error,
) {
	w.Header().Add("Content-Type", "text/html")

	var buf bytes.Buffer
	if err := executeTempl(&buf); err != nil {
		fmt.Printf("error: %s\n", err)

		w.WriteHeader(500)
		buf.Reset()
		if err := a.page500.Execute(&buf, nil); err != nil {
			fmt.Fprint(w, "<h1>Internal Server Error</h1")
		}
		return
	}

	w.WriteHeader(status)
	io.Copy(w, &buf)
}

type apiResponse struct {
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func getAbsRootPath(path string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	if wd[len(wd)-3:] == "cmd" {
		wd = wd[0 : len(wd)-4]
	}

	return fmt.Sprintf("%s/%s", wd, path), nil
}
