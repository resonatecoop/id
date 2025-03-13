package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/id/util/response"
	"github.com/resonatecoop/user-api/model"
)

func (s *Service) passwordResetForm(w http.ResponseWriter, r *http.Request) {
	sessionService, err := s.passwordResetCommon(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	state := NewGuestInitialState(
		s.cnf,
	)

	query := r.URL.Query()
	token := r.Form.Get("token")

	if token != "" {
		emailToken, user, err := s.oauthService.GetValidEmailToken(
			token,
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
			query.Del("token")
			redirectWithQueryString("/web/password-reset", query, w, r)
			return
		}

		// we delete the validated email token
		err = s.oauthService.DeleteEmailToken(emailToken, true)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// we creates a new one without sending
		emailToken, err = s.oauthService.CreateEmailToken()

		email := user.Username
		claims := model.NewOauthEmailTokenClaims(email, emailToken)

		token, err := s.oauthService.CreateJwtEmailTokenClaims(claims)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("X-CSRF-Token", csrf.Token(r))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		flash, err := sessionService.GetFlashMessage()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		PasswordResetUpdatePassword(
			s.cnf.IsDevelopment,
			r.URL.Path,
			getQueryString(query),
			string(csrf.TemplateField(r)),
			"Update your password",
			"",
			&Profile{},
			flash,
			state.toFragment(),
			state,
			token,
			w,
		)
		return
	}

	flash, err := sessionService.GetFlashMessage()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	PasswordReset(
		s.cnf.IsDevelopment,
		r.URL.Path,
		getQueryString(query),
		string(csrf.TemplateField(r)),
		"Reset your password",
		"",
		&Profile{},
		flash,
		state.toFragment(),
		state,
		w,
	)
}

func (s *Service) passwordReset(w http.ResponseWriter, r *http.Request) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if strings.ToLower(r.Form.Get("_method")) == "put" || r.Method == http.MethodPut {
		err = s.passwordResetUpdatePassword(r)

		if err != nil {
			if r.Header.Get("Accept") == "application/json" {
				response.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
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

		message := "Your password was updated successfully. A confirmation email has been sent."

		if r.Header.Get("Accept") == "application/json" {
			response.WriteJSON(w, map[string]interface{}{
				"message": message,
			}, http.StatusAccepted)
			return
		}

		err = sessionService.SetFlashMessage(&session.Flash{
			Type:    "Info",
			Message: message,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		redirectWithQueryString("/web/login", r.URL.Query(), w, r)
		return
	}

	// send password reset token
	go func() {
		_, _ = s.oauthService.SendEmailToken(
			model.NewOauthEmail(
				r.Form.Get("email"),
				"Reset your password",
				"password-reset",
			),
			fmt.Sprintf(
				"https://%s/password-reset",
				s.cnf.Hostname,
			),
		)
		// maybe log something later
	}()

	message := "If you do have a Resonate account, you should receive an email shortly"

	if r.Header.Get("Accept") == "application/json" {
		response.WriteJSON(w, map[string]interface{}{
			"message": message,
			"status":  http.StatusAccepted,
		}, http.StatusAccepted)
		return
	}

	err = sessionService.SetFlashMessage(&session.Flash{
		Type:    "Info",
		Message: message,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, r.RequestURI, http.StatusFound)
	return
}

func (s *Service) passwordResetUpdatePassword(r *http.Request) error {
	emailToken, user, err := s.oauthService.GetValidEmailToken(r.Form.Get("token"))

	if err != nil {
		return err
	}

	if r.Form.Get("password_new") != r.Form.Get("password_confirm") {
		return ErrPasswordMismatch
	}

	err = s.oauthService.SetPassword(user, r.Form.Get("password_new"))

	if err != nil {
		return err
	}

	if !user.EmailConfirmed {
		err = s.oauthService.ConfirmUserEmail(user.Username)

		if err != nil {
			return err
		}
	}

	softDelete := true

	err = s.oauthService.DeleteEmailToken(emailToken, softDelete)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) passwordResetCommon(r *http.Request) (
	session.ServiceInterface,
	error,
) {
	// Get the session service from the request context
	sessionService, err := getSessionService(r)
	if err != nil {
		return nil, err
	}

	return sessionService, nil
}
