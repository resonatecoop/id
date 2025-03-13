package web

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/resonatecoop/id/session"
)

func (s *Service) indexForm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	state := NewGuestInitialState(s.cnf)

	query := r.URL.Query()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	Index(
		s.cnf.IsDevelopment,
		r.URL.Path,
		getQueryString(query),
		string(csrf.TemplateField(r)),
		"Resonate",
		"Resonate ID Server",
		&Profile{},
		&session.Flash{},
		state.toFragment(),
		state,
		w,
	)
}
