package oauth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	uuid "github.com/google/uuid"
	"github.com/resonatecoop/id/oauth"
	"github.com/resonatecoop/id/oauth/tokentypes"
	testutil "github.com/resonatecoop/id/test-util"
	"github.com/resonatecoop/user-api/model"
	"github.com/stretchr/testify/assert"
)

func (suite *OauthTestSuite) TestNewIntrospectResponseFromAccessToken() {

	accessToken := &model.AccessToken{
		Token:     "test_token_introspect_1",
		ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
		ClientID:  suite.clients[0].ID,
		UserID:    suite.users[0].ID,
		Scope:     "read_write",
	}
	expected := &oauth.IntrospectResponse{
		Active:    true,
		Scope:     accessToken.Scope,
		TokenType: tokentypes.Bearer,
		ExpiresAt: int(accessToken.ExpiresAt.Unix()),
		ClientID:  suite.clients[0].Key,
		UserID:    accessToken.UserID.String(),
		Username:  suite.users[0].Username,
	}

	actual, err := suite.service.NewIntrospectResponseFromAccessToken(accessToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, actual)

	accessToken.ClientID = uuid.Nil
	expected.ClientID = ""
	actual, err = suite.service.NewIntrospectResponseFromAccessToken(accessToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, actual)

	accessToken.UserID = uuid.Nil
	expected.Username = ""
	expected.UserID = ""
	actual, err = suite.service.NewIntrospectResponseFromAccessToken(accessToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, actual)
}

func (suite *OauthTestSuite) TestNewIntrospectResponseFromRefreshToken() {
	refreshToken := &model.RefreshToken{
		Token:     "test_token_introspect_1",
		ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
		ClientID:  suite.clients[0].ID,
		UserID:    suite.users[0].ID,
		Scope:     "read_write",
	}
	expected := &oauth.IntrospectResponse{
		Active:    true,
		Scope:     refreshToken.Scope,
		TokenType: tokentypes.Bearer,
		ExpiresAt: int(refreshToken.ExpiresAt.Unix()),
		ClientID:  suite.clients[0].Key,
		UserID:    refreshToken.UserID.String(),
		Username:  suite.users[0].Username,
	}

	actual, err := suite.service.NewIntrospectResponseFromRefreshToken(refreshToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, actual)

	refreshToken.ClientID = uuid.Nil
	expected.ClientID = ""
	actual, err = suite.service.NewIntrospectResponseFromRefreshToken(refreshToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, actual)

	refreshToken.UserID = uuid.Nil
	expected.Username = ""
	expected.UserID = ""
	actual, err = suite.service.NewIntrospectResponseFromRefreshToken(refreshToken)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, actual)
}

func (suite *OauthTestSuite) TestHandleIntrospectMissingToken() {
	// Make a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/introspect", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{}

	// And serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrTokenMissing.Error(),
		400,
	)
}

func (suite *OauthTestSuite) TestHandleIntrospectInvailidTokenHint() {
	// Make a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/introspect", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")
	r.PostForm = url.Values{"token": {"token"}, "token_type_hint": {"wrong"}}

	// And serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrTokenHintInvalid.Error(),
		400,
	)
}

func (suite *OauthTestSuite) TestHandleIntrospectAccessToken() {
	// Insert a test access token with a user
	accessToken := &model.AccessToken{
		IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
		Token:     "test_token_introspect_1",
		ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
		ClientID:  suite.clients[0].ID,
		UserID:    suite.users[0].ID,
		Scope:     "read_write",
	}

	ctx := context.Background()

	_, err := suite.db.NewInsert().
		Model(accessToken).
		Exec(ctx)

	// Insertion worked
	assert.Nil(suite.T(), err)

	// Make a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/introspect", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")

	// With correct token hint
	r.PostForm = url.Values{
		"token":           {accessToken.Token},
		"token_type_hint": {oauth.AccessTokenHint},
	}

	// And serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	expected, err := suite.service.NewIntrospectResponseFromAccessToken(accessToken)
	assert.NoError(suite.T(), err)
	testutil.TestResponseObject(suite.T(), w, expected, 200)

	// With incorrect token hint
	r.PostForm = url.Values{
		"token":           {accessToken.Token},
		"token_type_hint": {oauth.RefreshTokenHint},
	}

	// Serve the request
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrRefreshTokenNotFound.Error(),
		404,
	)

	// Without token hint
	r.PostForm = url.Values{
		"token": {accessToken.Token},
	}

	// Serve the request
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	expected, err = suite.service.NewIntrospectResponseFromAccessToken(accessToken)
	assert.NoError(suite.T(), err)
	testutil.TestResponseObject(suite.T(), w, expected, 200)
}

func (suite *OauthTestSuite) TestHandleIntrospectRefreshToken() {
	// Insert a test refresh token with a user
	refreshToken := &model.RefreshToken{
		IDRecord:  model.IDRecord{CreatedAt: time.Now().UTC()},
		Token:     "test_token_introspect_1",
		ExpiresAt: time.Now().UTC().Add(+10 * time.Second),
		ClientID:  suite.clients[0].ID,
		UserID:    suite.users[0].ID,
		Scope:     "read_write",
	}

	ctx := context.Background()

	_, err := suite.db.NewInsert().
		Model(refreshToken).
		Exec(ctx)

	// Insertion worked
	assert.Nil(suite.T(), err)

	// Make a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/introspect", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")

	// With correct token hint
	r.PostForm = url.Values{
		"token":           {refreshToken.Token},
		"token_type_hint": {oauth.RefreshTokenHint},
	}

	// And serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check the response
	expected, err := suite.service.NewIntrospectResponseFromRefreshToken(refreshToken)
	assert.NoError(suite.T(), err)
	testutil.TestResponseObject(suite.T(), w, expected, 200)

	// With incorrect token hint
	r.PostForm = url.Values{
		"token":           {refreshToken.Token},
		"token_type_hint": {oauth.AccessTokenHint},
	}

	// Serve the request
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrAccessTokenNotFound.Error(),
		404,
	)

	// Without token hint
	r.PostForm = url.Values{
		"token": {refreshToken.Token},
	}

	// Serve the request
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrAccessTokenNotFound.Error(),
		404,
	)
}

func (suite *OauthTestSuite) TestHandleIntrospectInactiveToken() {
	// Make a request
	r, err := http.NewRequest("POST", "http://1.2.3.4/v1/oauth/introspect", nil)
	assert.NoError(suite.T(), err, "Request setup should not get an error")
	r.SetBasicAuth("test_client_1", "test_secret")

	// With access token hint
	r.PostForm = url.Values{
		"token":           {"unexisting_token"},
		"token_type_hint": {oauth.AccessTokenHint},
	}

	// And serve the request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrAccessTokenNotFound.Error(),
		404,
	)

	// With refresh token hint
	r.PostForm = url.Values{
		"token":           {"unexisting_token"},
		"token_type_hint": {oauth.RefreshTokenHint},
	}

	// Serve the request
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrRefreshTokenNotFound.Error(),
		404,
	)

	// Without token hint
	r.PostForm = url.Values{
		"token": {"unexisting_token"},
	}

	// Serve the request
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, r)

	// Check response
	testutil.TestResponseForError(
		suite.T(),
		w,
		oauth.ErrAccessTokenNotFound.Error(),
		404,
	)
}
