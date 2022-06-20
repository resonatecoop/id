package oauth

import (
	"errors"
	"net/http"

	"github.com/resonatecoop/id/util/response"
	"github.com/resonatecoop/user-api/model"
)

var (
	// ErrInvalidGrantType ...
	ErrInvalidGrantType = errors.New("Invalid grant type")
	// ErrInvalidClientIDOrSecret ...
	ErrInvalidClientIDOrSecret = errors.New("Invalid client ID or secret")
)

// tokensHandler handles all OAuth 2.0 grant types
// (POST /v1/oauth/tokens)
func (s *Service) tokensHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the form so r.Form becomes available
	if err := r.ParseForm(); err != nil {
		response.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Map of grant types against handler functions
	grantTypes := map[string]func(r *http.Request, client *model.Client) (*AccessTokenResponse, error){
		"authorization_code": s.authorizationCodeGrant,
		"password":           s.passwordGrant,
		"client_credentials": s.clientCredentialsGrant,
		"refresh_token":      s.refreshTokenGrant,
	}

	// Check the grant type
	grantHandler, ok := grantTypes[r.Form.Get("grant_type")]
	if !ok {
		response.Error(w, ErrInvalidGrantType.Error(), http.StatusBadRequest)
		return
	}

	// Client auth
	client, err := s.basicAuthClient(r)
	if err != nil {
		response.UnauthorizedError(w, err.Error())
		return
	}

	// Grant processing
	resp, err := grantHandler(r, client)
	if err != nil {
		response.Error(w, err.Error(), getErrStatusCode(err))
		return
	}

	// Write response to json
	response.WriteJSON(w, resp, 200)
}

// introspectHandler handles OAuth 2.0 introspect request
// (POST /v1/oauth/introspect)
func (s *Service) introspectHandler(w http.ResponseWriter, r *http.Request) {
	// Client auth
	client, err := s.basicAuthClient(r)
	if err != nil {
		response.UnauthorizedError(w, err.Error())
		return
	}

	// Introspect the token
	resp, err := s.introspectToken(r, client)
	if err != nil {
		response.Error(w, err.Error(), getErrStatusCode(err))
		return
	}

	// Write response to json
	response.WriteJSON(w, resp, 200)
}

// Get client credentials from basic auth and try to authenticate client
func (s *Service) basicAuthClient(r *http.Request) (*model.Client, error) {
	// Get client credentials from basic auth
	clientID, secret, ok := r.BasicAuth()
	if !ok {
		return nil, ErrInvalidClientIDOrSecret
	}

	// Authenticate the client
	client, err := s.AuthClient(clientID, secret)
	if err != nil {
		// For security reasons, return a general error message
		return nil, ErrInvalidClientIDOrSecret
	}

	return client, nil
}
