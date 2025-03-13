package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/csrf"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/util/response"
	"github.com/resonatecoop/user-api/model"
)

func (s *Service) loginForm(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("X-CSRF-Token", csrf.Token(r))

	state := NewGuestInitialState(s.cnf)

	flash, _ := sessionService.GetFlashMessage()

	// Get the client from the request context
	client, err := getClient(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	query := r.URL.Query()

	Login(
		s.cnf.IsDevelopment,
		r.URL.Path,
		getQueryString(query),
		string(csrf.TemplateField(r)),
		"Log in",
		"Log in to your Resonate account",
		&Profile{},
		flash,
		state.toFragment(),
		state,
		client,
		w,
	)
}

func (s *Service) login(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get the client from the request context
	client, err := getClient(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Authenticate the user
	user, err := s.oauthService.AuthUser(
		r.Form.Get("email"),    // email/username
		r.Form.Get("password"), // password
	)

	if err != nil {
		switch r.Header.Get("Accept") {
		case "application/json":
			response.Error(w, err.Error(), http.StatusBadRequest)
		default:
			err = sessionService.SetFlashMessage(&session.Flash{
				Type:    "Error",
				Message: err.Error(),
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, r.RequestURI, http.StatusFound)
		}
		return
	}

	t := time.Date(2022, 8, 17, 0, 0, 0, 0, time.UTC)

	if user.CreatedAt.Before(t) && user.LastPasswordChange.IsZero() {
		message := "Please reset your password. You should receive an e-mail shortly."
		switch r.Header.Get("Accept") {
		case "application/json":
			response.Error(w, message, http.StatusBadRequest)
		default:
			err = sessionService.SetFlashMessage(&session.Flash{
				Type:    "Info",
				Message: message,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, r.RequestURI, http.StatusFound)
		}

		go func() {
			email := model.NewOauthEmail(
				user.Username,
				"Reset your password",
				"password-reset",
			)
			_, _ = s.oauthService.SendEmailToken(
				email,
				fmt.Sprintf(
					"https://%s/password-reset",
					s.cnf.Hostname,
				),
			)
		}()

		return
	}

	// Email should be confirmed (click autologin link in email)
	if !user.EmailConfirmed {
		switch r.Header.Get("Accept") {
		case "application/json":
			response.Error(w, "Please confirm your email", http.StatusBadRequest)
		default:
			err = sessionService.SetFlashMessage(&session.Flash{
				Type:    "Error",
				Message: "Please confirm your email",
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, r.RequestURI, http.StatusFound)
		}

		go func() {
			email := model.NewOauthEmail(
				user.Username,
				"Confirm your email",
				"email-confirmation",
			)
			_, _ = s.oauthService.SendEmailToken(
				email,
				fmt.Sprintf(
					"https://%s/email-confirmation",
					s.cnf.Hostname,
				),
			)
		}()

		return
	}

	// Get the scope string
	scope, err := s.oauthService.GetScope("read_write")
	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
		return
	}

	// Log in the user
	accessToken, refreshToken, err := s.oauthService.Login(
		client,
		user,
		scope,
	)
	if err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
		return
	}

	scopes := strings.Split(accessToken.Scope, " ")

	// Log in the user and store the user session in a cookie
	userSession := &session.UserSession{
		ClientID:     client.Key,
		Username:     user.Username,
		Role:         scopes[1],
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken.Token,
	}
	if err := sessionService.SetUserSession(userSession); err != nil {
		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Error",
			Message: err.Error(),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.RequestURI, http.StatusFound)
		return
	}

	// Redirect to the authorize page by default but allow redirection to other
	// pages by specifying a path with login_redirect_uri query string param
	loginRedirectURI := r.URL.Query().Get("login_redirect_uri")
	if loginRedirectURI == "" {
		loginRedirectURI = "/web/authorize"
	}

	query := r.URL.Query()

	redirectWithQueryString(loginRedirectURI, query, w, r)
	return
}
