package auth

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/kevguy/My-Shitty-Music-Backend/models"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var conf *oauth2.Config

// Credentials which stores google ids.
type Credentials struct {
	Cid     string `json:"cid"`
	Csecret string `json:"csecret"`
}

func initGoogleAuth(c Credentials) {
	// Gob encoding for gorilla/sessions
	gob.Register(&oauth2.Token{})
	gob.Register(&models.GoogleProfile{})

	conf = &oauth2.Config{
		ClientID:     c.Cid,
		ClientSecret: c.Csecret,
		// no redirect url is set, refer to https://stackoverflow.com/questions/28321570/google-oauth-2-0-error-redirect-uri-mismatch
		RedirectURL: "postmessage",
		// RedirectURL: "",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email", // You have to select your own scope from here -> https://developers.google.com/identity/protocols/googlescopes#google_sign-in
		},
		Endpoint: google.Endpoint,
	}
}

func fetchProfile(ctx context.Context, tok *oauth2.Token) (*models.GoogleProfile, error) {
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

	return &models.GoogleProfile{
		ID:          res["sub"].(string),
		DisplayName: res["given_name"].(string),
		FullName:    res["name"].(string),
		ImageURL:    res["picture"].(string),
		Email:       res["email"].(string),
	}, nil
}

func RetrieveGoogleUserProfile(code string) models.GoogleProfile {
	fmt.Println("Starting RetrieveGoogleUserProfile")
	tok, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		panic(err)
	}
	fmt.Println("after exchange")
	fmt.Println(tok)

	profile, err := fetchProfile(oauth2.NoContext, tok)
	// ctx := context.Background()
	// profile, err := fetchProfile(ctx, tok)
	if err != nil {
		panic(err)
	}

	return *profile
}

func CreateRoutes(c Credentials, r *mux.Router) {
	initGoogleAuth(c)
}
