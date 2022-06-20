package oauth

import (
	"net/http"

	"github.com/resonatecoop/id/oauth/tokentypes"
	"github.com/resonatecoop/user-api/model"
)

func (s *Service) refreshTokenGrant(r *http.Request, client *model.Client) (*AccessTokenResponse, error) {
	// Fetch the refresh token
	theRefreshToken, err := s.GetValidRefreshToken(r.Form.Get("refresh_token"), client)
	if err != nil {
		return nil, err
	}

	// Get the scope
	scope, err := s.getRefreshTokenScope(theRefreshToken, r.Form.Get("scope"))
	if err != nil {
		return nil, err
	}

	// Log in the user
	accessToken, refreshToken, err := s.Login(
		theRefreshToken.Client,
		theRefreshToken.User,
		scope,
	)
	if err != nil {
		return nil, err
	}

	// Create response
	accessTokenResponse, err := NewAccessTokenResponse(
		accessToken,
		refreshToken,
		s.cnf.Oauth.AccessTokenLifetime,
		tokentypes.Bearer,
	)
	if err != nil {
		return nil, err
	}

	return accessTokenResponse, nil
}
