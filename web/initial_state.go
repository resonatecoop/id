package web

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/resonatecoop/id/config"
	"github.com/resonatecoop/id/session"
	"github.com/resonatecoop/user-api-client/models"
	"github.com/resonatecoop/user-api/model"
)

// Usergroup public
type UserGroup struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Banner      string `json:"banner"`
}

// Profile user public profile
type Profile struct {
	ID          string `json:"id"`
	Role        string `json:"role"`
	LegacyID    int32  `json:"legacyID"`
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Credits     string `json:"credits"`
	// FullName               string                                 `json:"fullName"`
	// FirstName              string                                 `json:"firstName"`
	// LastName               string                                 `json:"lastName"`
	Country                string      `json:"country"`
	NewsletterNotification bool        `json:"newsletterNotification"`
	EmailConfirmed         bool        `json:"emailConfirmed"`
	Member                 bool        `json:"member"`
	Complete               bool        `json:"complete"`
	Usergroups             []UserGroup `json:"usergroups"`
}

// NewProfile
func NewProfile(
	user *model.User,
	usergroups []*models.UserUserGroupPrivateResponse,
	isUserAccountComplete bool,
	credits string,
	role string,
) *Profile {
	displayName := ""

	if len(usergroups) > 0 {
		displayName = usergroups[0].DisplayName
	}

	var usergroupList []UserGroup

	for i := range usergroups {
		usergroupList = append(usergroupList, UserGroup{
			ID:          usergroups[i].ID,
			DisplayName: usergroups[i].DisplayName,
			Avatar:      usergroups[i].Avatar,
			Banner:      usergroups[i].Banner,
		})
	}

	return &Profile{
		ID:                     user.ID.String(),
		Complete:               isUserAccountComplete,
		Country:                user.Country,
		Credits:                credits,
		DisplayName:            displayName,
		Role:                   role,
		Email:                  user.Username,
		EmailConfirmed:         user.EmailConfirmed,
		LegacyID:               user.LegacyID,
		Member:                 user.Member,
		NewsletterNotification: user.NewsletterNotification,
		Usergroups:             usergroupList,
	}
}

type InitialState struct {
	ApplicationName       string                `json:"applicationName"`
	StaticURL             string                `json:"staticURL"`
	AppURL                string                `json:"appURL"`
	IsUserAccountComplete bool                  `json:"isUserAccountComplete"`
	ClientID              string                `json:"clientID"`
	QueryString           string                `json:"queryString"`
	UserGroup             string                `json:"usergroup"`
	Token                 string                `json:"token"`
	Clients               []config.ClientConfig `json:"clients"`
	Profile               *Profile              `json:"profile"`
	Memberships           []Membership          `json:"memberships"`
	Shares                []Share               `json:"shares"`
	CSRFToken             string                `json:"csrfToken"`
	CountryList           []Country             `json:"countries"`
	CurrentDate           time.Time             `json:"currentDate"`
}

func (state InitialState) toFragment() string {
	initialState, err := json.Marshal(state)

	if err != nil {
		panic(err)
	}

	replacer := strings.NewReplacer(`'`, `\'`)

	escaped := replacer.Replace(string(initialState))

	// Inject initial state into frontend
	fragment := fmt.Sprintf(
		`<script>window.initialState=JSON.parse('%s')</script>`,
		escaped,
	)
	return fragment
}

func NewGuestInitialState(
	cnf *config.Config,
) *InitialState {
	return &InitialState{
		Clients:     cnf.Clients,
		StaticURL:   cnf.StaticURL,
		AppURL:      cnf.AppURL,
		CountryList: getCountryList(),
		CurrentDate: time.Now(),
	}
}

func NewInitialState(
	cnf *config.Config,
	client *model.Client,
	user *model.User,
	userSession *session.UserSession,
	isUserAccountComplete bool,
	credits string,
	usergroups []*models.UserUserGroupPrivateResponse,
	memberships []Membership,
	shares []Share,
	csrfToken string,
	countryList []Country,
) *InitialState {
	if userSession != nil {
		accessToken := userSession.AccessToken

		profile := NewProfile(
			user,
			usergroups,
			isUserAccountComplete,
			credits,
			userSession.Role,
		)

		return &InitialState{
			ApplicationName:       client.ApplicationName.String,
			ClientID:              client.Key,
			StaticURL:             cnf.StaticURL,
			AppURL:                cnf.AppURL,
			IsUserAccountComplete: isUserAccountComplete,
			Clients:               cnf.Clients,
			Profile:               profile,
			Token:                 accessToken,
			Memberships:           memberships,
			Shares:                shares,
			CSRFToken:             csrfToken,
			CountryList:           countryList,
			CurrentDate:           time.Now(),
		}
	}

	return NewGuestInitialState(
		cnf,
	)
}
