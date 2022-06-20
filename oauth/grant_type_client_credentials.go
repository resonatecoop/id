package oauth

import (
	"net/http"

	"github.com/resonatecoop/id/oauth/tokentypes"
	"github.com/resonatecoop/user-api/model"
)

func (s *Service) clientCredentialsGrant(r *http.Request, client *model.Client) (*AccessTokenResponse, error) {
	// Get the scope string
	scope, err := s.GetScope(r.Form.Get("scope"))
	if err != nil {
		return nil, err
	}

	// Create a new access token
	accessToken, err := s.GrantAccessToken(
		client,
		nil,                             // empty user
		s.cnf.Oauth.AccessTokenLifetime, // expires in
		scope,
	)
	if err != nil {
		return nil, err
	}

	// Create response
	accessTokenResponse, err := NewAccessTokenResponse(
		accessToken,
		nil, // refresh token
		s.cnf.Oauth.AccessTokenLifetime,
		tokentypes.Bearer,
	)
	if err != nil {
		return nil, err
	}

	return accessTokenResponse, nil
}
