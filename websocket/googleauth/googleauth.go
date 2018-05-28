package googleauth

import (
	"Redis-Exploration/websocket/dao"
	"Redis-Exploration/websocket/util"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var cred Credentials
var conf *oauth2.Config
var state string
var store = sessions.NewCookieStore([]byte("something-very-secret"))

// Credentials which stores google ids.
type Credentials struct {
	Cid     string `json:"cid"`
	Csecret string `json:"csecret"`
}

const (
	defaultSessionID = "default"
	// The following keys are used for the default session. For example:
	googleProfileSessionKey = "google_profile"
	oauthTokenSessionKey    = "oauth_token"

	// This key is used in the OAuth flow session to store the URL to redirect the
	// user to after the OAuth flow is complete.
	oauthFlowRedirectKey = "redirect"
)

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

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func initGoogleAuth(c Credentials) {
	// Gob encoding for gorilla/sessions
	gob.Register(&oauth2.Token{})
	gob.Register(&Profile{})

	conf = &oauth2.Config{
		ClientID:     c.Cid,
		ClientSecret: c.Csecret,
		RedirectURL:  "http://localhost:3000/googleauth/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", // You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
		},
		Endpoint: google.Endpoint,
	}
}

// AllSongsEndPoint finds all songs
func getLoginURL(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.Must(uuid.NewV4()).String()
	fmt.Println("sessionID")
	fmt.Println(sessionID)

	oauthFlowSession, err := store.New(r, sessionID)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	oauthFlowSession.Options.MaxAge = 10 * 60 // 10 minutes

	// redirectURL, err := validateRedirectURL(r.FormValue("redirect"))
	// if err != nil {
	// 	util.RespondWithError(w, http.StatusInternalServerError, err.Error())
	// }
	oauthFlowSession.Values[oauthFlowRedirectKey] = "/"
	oauthFlowSession.Values["fuck"] = "fuck"

	if err := oauthFlowSession.Save(r, w); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "could not save session")
	}

	// fmt.Println("oauthFlowSession found")
	fmt.Println(oauthFlowSession.Values[oauthFlowRedirectKey])
	fmt.Println(oauthFlowSession.Values["fuck"])

	// Use the session ID for the "state" parameter.
	// This protects against CSRF (cross-site request forgery).
	// See https://godoc.org/golang.org/x/oauth2#Config.AuthCodeURL for more detail.
	state = randToken()
	loginURL := conf.AuthCodeURL(sessionID)
	// url := bookshelf.OAuthConfig.AuthCodeURL(sessionID, oauth2.ApprovalForce,
	// 	oauth2.AccessTypeOnline)
	fmt.Println("login url")
	fmt.Println(loginURL)
	util.RespondWithJSON(w, http.StatusOK, loginURL)
}

func fetchProfile(ctx context.Context, tok *oauth2.Token) (*Profile, error) {
	client := conf.Client(ctx, tok)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return nil, err
	}

	defer email.Body.Close()
	data, _ := ioutil.ReadAll(email.Body)
	log.Println("Email body: ", string(data))
	var res map[string]interface{}
	if err := json.Unmarshal(data, &res); err != nil {
		panic(err)
	}
	// Sample Email body
	// 	Email body:  {
	//  "sub": "105524006654987809707",
	//  "name": "Kev Lai",
	//  "given_name": "Kev",
	//  "family_name": "Lai",
	//  "profile": "https://plus.google.com/105524006654987809707",
	//  "picture": "https://lh3.googleusercontent.com/-XdUIqdMkCWA/AAAAAAAAAAI/AAAAAAAAAAA/4252rscbv5M/photo.jpg",
	//  "email": "kevatuk@gmail.com",
	//  "email_verified": true
	// }

	return &Profile{
		ID:          res["sub"].(string),
		DisplayName: res["given_name"].(string),
		FullName:    res["name"].(string),
		ImageURL:    res["picture"].(string),
		Email:       res["email"].(string),
	}, nil
}

// oauthCallbackHandler completes the OAuth flow, retreives the user's profile
// information and stores it in a session.
func authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("fuckin given state")
	fmt.Println(r.FormValue("state"))
	oauthFlowSession, err := store.Get(r, r.FormValue("state"))
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "invalid state parameter. try logging in again.")
		return
	}
	fmt.Println("oauthFlowSession found")
	// b, err := json.MarshalIndent(oauthFlowSession.Values, "", "  ")
	// if err != nil {
	// 	fmt.Println("error:", err)
	// }
	// fmt.Print(string(b))
	// fmt.Println(oauthFlowSession.Values[oauthFlowRedirectKey].(string))
	fmt.Println(oauthFlowSession.Values["fuck"])

	// redirectURL, ok := oauthFlowSession.Values[oauthFlowRedirectKey].(string)
	// // Validate this callback request came from the app.
	// if !ok {
	// 	util.RespondWithError(w, http.StatusInternalServerError, "invalid state parameter. try logging in again.")
	// 	return
	// }

	code := r.FormValue("code")
	if string(code) == "" {
		util.RespondWithError(w, http.StatusInternalServerError, "Code Not Found")
		return
	}
	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	session, err := store.New(r, defaultSessionID)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "could not get default session")
		return
	}

	profile, err := fetchProfile(oauth2.NoContext, tok)
	// ctx := context.Background()
	// profile, err := fetchProfile(ctx, tok)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "could not fetch profile")
		return
	}

	session.Values[oauthTokenSessionKey] = tok
	session.Values[googleProfileSessionKey] = profile
	if err := session.Save(r, w); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "could not save session")
		return
	}

	http.Redirect(w, r, "http://localhost:3000/songs", http.StatusFound)

}

func CreateRoutes(c Credentials, r *mux.Router, _dao *dao.ShittyMusicDAO) {

	// store = *cookieStore
	store.Options = &sessions.Options{
		Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 8, // 8 hours
		HttpOnly: true,
	}
	initGoogleAuth(c)

	r.HandleFunc("/googleauth/loginurl", getLoginURL).Methods("GET", "OPTIONS")
	r.HandleFunc("/googleauth/auth", authCallbackHandler).Methods("GET", "OPTIONS")
}

type Profile struct {
	ID, DisplayName, FullName, ImageURL, Email string
}
