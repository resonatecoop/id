package web

import (
	"net/http"

	"github.com/resonatecoop/id/session"
)

func (s *Service) contactForm(w http.ResponseWriter, r *http.Request) {
	// Set the header and write the buffer to the http.ResponseWriter
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	Contact(
		s.cnf.IsDevelopment,
		s.cnf.Version,
		r.URL.Path,
		getQueryString(r.URL.Query()),
		"",
		"Contact",
		"This is a blank contact form page just for testing",
		&Profile{},
		&session.Flash{},
		"",
		&InitialState{},
		w,
	)
}
