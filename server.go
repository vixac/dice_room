package main

import (
	"dice_room/model"
	"dice_room/store"
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*
//go:embed static/*
var content embed.FS

// Server holds all dependencies and serves as the receiver for HTTP handlers.
type Server struct {
	store         store.Store
	broadcaster   *Broadcaster
	templates     *template.Template
	hostPrefix    string
	secureCookies bool
}

func NewServer(store store.Store, broadcaster *Broadcaster, hostPrefix string, secureCookies bool) *Server {
	tmpl := template.Must(template.New("").Funcs(template.FuncMap{
		"static": func(path string) string {
			return hostPrefix + "/static/" + path
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"reverse": func(xs []model.LogEntry) []model.LogEntry {
			out := make([]model.LogEntry, len(xs))
			for i := range xs {
				out[i] = xs[len(xs)-1-i]
			}
			return out
		},
	}).ParseFS(content, "templates/*.html"))

	return &Server{
		store:         store,
		broadcaster:   broadcaster,
		templates:     tmpl,
		hostPrefix:    hostPrefix,
		secureCookies: secureCookies,
	}
}

// routes wires all URL patterns to their handlers and returns the mux.
func (s *Server) routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(http.FS(content)))
	mux.HandleFunc("/", s.indexHandler)
	mux.HandleFunc("/room/", s.roomHandler)
	mux.HandleFunc("/events/", s.eventsHandler)
	mux.HandleFunc("/privacy", s.privacyHandler)
	mux.HandleFunc("/terms", s.termsHandler)
	return mux
}
