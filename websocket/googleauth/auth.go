package googleauth

import (
	"Redis-Exploration/websocket/dao"
	"Redis-Exploration/websocket/util"
	"context"
	"encoding/gob"
	"errors"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	uuid "github.com/satori/go.uuid"
)

var cred Credentials
var conf *oauth2.Config
var state string
var store *sessions.CookieStore

// Credentials which stores google ids.
type Credentials struct {
	Cid     string `json:"cid"`
	Csecret string `json:"csecret"`
}

const (
	defaultSessionID = "default"
	// The following keys are used for the default session. For example:
	// session, _ := store,New(rm defaultSessionID)
	// session.Values[oauthTokenSessionKey]
	googleProfileSessionKey = "google_profile"
	oauthTokenSessionKey    = "oauth_token"

	// This key is used in the OAuth flow session to store the URL to redirect the
	// user to after the OAuth flow is complete.
	oauthFlowRedirectKey = "redirect"
)

func init() {
	// Gob encoding for gorilla/sessions
	gob.Register(&oauth2.Token{})
	// gob.Register(&Profile{})
}

// validateRedirectURL checks that the URL provided is valid.
// If the URL is missing, redirect the user to the application's root.
// The URL must not be absolute (i.e., the URL must refer to a path within this
// application).
func validateRedirectURL(path string) (string, error) {
	if path == "" {
		return "/", nil
	}

	// Ensure redirect URL is valid and not pointing to a different server.
	parsedURL, err := url.Parse(path)
	if err != nil {
		return "/", err
	}
	if parsedURL.IsAbs() {
		return "/", errors.New("URL must not be absolute")
	}
	return path, nil
}

// loginHandler initiates an OAuth flow to authenticate the user.
func loginHandler(w http.ResponseWriter, r *http.Request) *appError {
	sessionID := uuid.Must(uuid.NewV4()).String()

	oauthFlowSession, err := store.New(r, sessionID)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "Could not create oauth session")
		return
	}
	oauthFlowSession.Options.MaxAge = 10 * 60 // 10 minutes

	redirectURL, err := validateRedirectURL(r.FormValue("redirect"))
	if err != nil {
		// return appErrorf(err, "invalid redirect URL: %v", err)
		util.RespondWithError(w, http.StatusInternalServerError, "Invalid redirect URL")

	}
	oauthFlowSession.Values[oauthFlowRedirectKey] = redirectURL

	if err := oauthFlowSession.Save(r, w); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "Could not save session")
		return
	}

	// Use the session ID for the "state" parameter.
	// This protects against CSRF (cross-site request forgery).
	// See https://godoc.org/golang.org/x/oauth2#Config.AuthCodeURL for more detail.
	url := conf.AuthCodeURL(sessionID, oauth2.ApprovalForce,
		oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusFound)
	return nil
}

func oauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	oauthFlowSession, err := store.Get(r, r.FormValue("state"))
	if err != nil {
		// return appErrorf(err, "invalid state parameter. try logging in again.")
		util.RespondWithError(w, http.StatusInternalServerError, "invalid state parameter. try logging in again.")
		return
	}

	redirectURL, ok := oauthFlowSession.Values[oauthFlowRedirectKey].(string)
	// Validate this callback request came from the app.
	if !ok {
		// return appErrorf(err, "invalid state parameter. try logging in again.")
		util.RespondWithError(w, http.StatusInternalServerError, "invalid state parameter. try logging in again.")
		return
	}

	code := r.FormValue("code")
	tok, err := conf.Exchange(context.Background(), code)
	if err != nil {
		// return appErrorf(err, "could not get auth token: %v", err)
		util.RespondWithError(w, http.StatusInternalServerError, "could not get auth token")
		return
	}

	session, err := store.New(r, defaultSessionID)
	if err != nil {
		// return appErrorf(err, "could not get default session: %v", err)
		util.RespondWithError(w, http.StatusInternalServerError, "could not get default session")
		return
	}

	ctx := context.Background()
	profile, err := fetchProfile(ctx, tok)
	if err != nil {
		return appErrorf(err, "could not fetch Google profile: %v", err)
	}

}

func CreateRoutes(c Credentials, r *mux.Router, cookieStore *sessions.CookieStore, _dao *dao.ShittyMusicDAO) {
	store = cookieStore
	initGoogleAuth(c)

	r.HandleFunc("/googleauth/loginurl", getLoginURL).Methods("GET")
	r.HandleFunc("/googleauth/auth", authHandler).Methods("GET")
}
