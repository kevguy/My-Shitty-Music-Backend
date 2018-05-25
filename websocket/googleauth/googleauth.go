package googleauth

import (
	"Redis-Exploration/websocket/dao"
	"Redis-Exploration/websocket/util"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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
	googleProfileSessionKey = "google_profile"
	oauthTokenSessionKey    = "oauth_token"

	// This key is used in the OAuth flow session to store the URL to redirect the
	// user to after the OAuth flow is complete.
	oauthFlowRedirectKey = "redirect"
)

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
	fmt.Println("I got you babe")
	fmt.Println(r.FormValue("redirect"))
	state = randToken()

	// Get a session. Get() always returns a session, even if empty.
	session, err := store.Get(r, "session-name")
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set some session values.
	// session.Values["foo"] = "bar"
	// session.Values[42] = 43
	session.Values["state"] = state

	// Save it before we write to the response/return from the handler.
	session.Save(r, w)

	// State can be some kind of random generated hash string.
	// See relevant RFC: http://tools.ietf.org/html/rfc6749#section-10.12
	fmt.Println("hihi")
	loginURL := conf.AuthCodeURL(state)

	util.RespondWithJSON(w, http.StatusOK, loginURL)

	// songs, err := shittyMusicDao.FindAllSongs()
	// if err != nil {
	// 	util.RespondWithError(w, http.StatusInternalServerError, err.Error())
	// 	return
	// }
	// util.RespondWithJSON(w, http.StatusOK, songs)
	// mywebsocket.BroadcastMsg(mywebsocket.Message{
	// 	Type:    "text",
	// 	Content: "AllSongsEndPoint",
	// })
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("handling Login")

	// Handle the exchange code to initiate a transport.
	session, err := store.Get(r, "session-name")
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	retrievedState := session.Values["state"]
	state := r.URL.Query().Get("state")
	log.Println(retrievedState)
	log.Println(state)
	// if retrievedState != state {
	// 	util.RespondWithError(w, http.StatusUnauthorized, "State not match")
	// 	return
	// }

	code := r.URL.Query().Get("code")
	if string(code) == "" {
		util.RespondWithError(w, http.StatusInternalServerError, "Code Not Found")
		return
	}
	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	client := conf.Client(oauth2.NoContext, tok)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	defer email.Body.Close()
	data, _ := ioutil.ReadAll(email.Body)
	log.Println("Email body: ", string(data))

	util.RespondWithJSON(w, http.StatusOK, "Login OK")

	// // Handle the exchange code to initiate a transport.
	//   session := sessions.Default(c)
	//   retrievedState := session.Get("state")
	//   if retrievedState != c.Query("state") {
	//       c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Invalid session state: %s", retrievedState))
	//       return
	//   }
	//
	// tok, err := conf.Exchange(oauth2.NoContext, c.Query("code"))
	// if err != nil {
	// 	c.AbortWithError(http.StatusBadRequest, err)
	//       return
	// }
	//
	// client := conf.Client(oauth2.NoContext, tok)
	// email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	//   if err != nil {
	// 	c.AbortWithError(http.StatusBadRequest, err)
	//       return
	// }
	//   defer email.Body.Close()
	//   data, _ := ioutil.ReadAll(email.Body)
	//   log.Println("Email body: ", string(data))
	//   c.Status(http.StatusOK)

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

}

func CreateRoutes(c Credentials, r *mux.Router, cookieStore *sessions.CookieStore, _dao *dao.ShittyMusicDAO) {
	store = cookieStore
	initGoogleAuth(c)

	r.HandleFunc("/googleauth/loginurl", getLoginURL).Methods("GET", "OPTIONS")
	r.HandleFunc("/googleauth/auth", authHandler).Methods("GET")
}

type Profile struct {
	ID, DisplayName, FullName, ImageURL, Email string
}
