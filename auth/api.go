package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/codegangsta/negroni"
	. "github.com/kevguy/My-Shitty-Music-Backend/models"
	"github.com/kevguy/My-Shitty-Music-Backend/mongodb"
	"github.com/kevguy/My-Shitty-Music-Backend/redis"

	"github.com/gorilla/mux"

	"github.com/kevguy/My-Shitty-Music-Backend/util"
)

var shittyMusicDao mongodb.ShittyMusicDAO
var shittyMusicRedisDao redisclient.ShittyMusicRedisDAO
var authentication JWTAuthentication

type GoogleRequest struct {
	Type string `bson:"type" json:"type"`
	Code string `bson:"code" json:"code"`
}

type VerifyRequest struct {
	UserID string `bson:"user_id" json:"user_id"`
	Token  string `bson:"token" json:"token"`
}

func HandlePreflight(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fmt.Println(r.Method)
	if r.Method == "OPTIONS" {
		fmt.Println("preflight")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		util.RespondWithJSON(w, http.StatusOK, "")

		return
	}
	next(w, r)
}

func AuthenticateEndPoint(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	log.Println("Authentication Request body: ", string(data))
	// var res map[string]interface{}
	var googleRequest GoogleRequest
	if err := json.Unmarshal(data, &googleRequest); err != nil {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid Authentication Request")
		return
	}
	fmt.Println(googleRequest)
	fmt.Println(googleRequest.Code)
	fmt.Println(googleRequest.Type)
	if googleRequest.Type == "" {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid Authentication Request")
		return
	}

	if googleRequest.Type == "google" {
		// authentication := InitJWTAuthentication()
		fmt.Println(authentication)

		code := googleRequest.Code
		// Retrieve Google Profile
		profile := RetrieveGoogleUserProfile(code)
		fmt.Println(profile.DisplayName)
		fmt.Println(profile.ID)
		fmt.Println(profile)

		// Find user
		var user User
		var err error
		user, err = shittyMusicDao.FindGoogleUser(profile.ID)
		if err != nil {
			if err.Error() == "not found" {
				// Create a new user
				err = shittyMusicDao.InsertGoogleUser(profile)
				if err != nil {
					fmt.Println(err.Error())
					panic(err)
				}
				user, err = shittyMusicDao.FindGoogleUser(profile.ID)
				if err != nil {
					fmt.Println(err.Error())
					panic(err)
				}
			} else {
				panic(err)
			}
		}

		// Generate token
		token, err := authentication.GenerateToken(user.ID.Hex())
		if err != nil {
			panic(err)
		}
		if err != nil {
			util.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Send token and basic user info
		response, _ := json.Marshal(map[string]string{"token": token})
		fmt.Println(string(response))
		util.RespondWithJSON(w, http.StatusOK, map[string]string{
			"token":     token,
			"user_name": user.Name,
			"user_id":   user.ID.Hex(),
		})
	} else {
		util.RespondWithError(w, http.StatusUnauthorized, "Not Supported")
	}
	return
}

func CheckLoginEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, _ := ioutil.ReadAll(r.Body)
	log.Println("Authentication Request body: ", string(data))

	var verifyRequest VerifyRequest
	if err := json.Unmarshal(data, &verifyRequest); err != nil {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid Authentication Request")
		return
	}
	fmt.Println(verifyRequest)
	fmt.Println(verifyRequest.UserID)
	fmt.Println(verifyRequest.Token)
	if verifyRequest.UserID == "" || verifyRequest.Token == "" {
		util.RespondWithError(w, http.StatusUnauthorized, "Invalid Authentication Request")
		return
	}

	// Verify Login
	// authentication := InitJWTAuthentication()
	fmt.Println(authentication)

	// Verify Token
	fmt.Println("Verifying Token")
	if result := authentication.VerifyToken(verifyRequest.UserID, verifyRequest.Token); result {
		util.RespondWithJSON(w, http.StatusOK, map[string]bool{"status": true})
	} else {
		util.RespondWithJSON(w, http.StatusOK, map[string]bool{"status": false})
	}
	return
}

func CreateAuthenticationRoutes(r *mux.Router,
	_dao *mongodb.ShittyMusicDAO,
	_redisDao *redisclient.ShittyMusicRedisDAO,
	_authentication *JWTAuthentication) {
	fmt.Println("Setting up Authentication Routes")
	shittyMusicDao = *_dao
	shittyMusicRedisDao = *_redisDao
	authentication = *_authentication

	r.Handle("/authenticate", negroni.New(
		negroni.HandlerFunc(HandlePreflight),
		negroni.HandlerFunc(AuthenticateEndPoint),
	)).Methods("POST", "OPTIONS")

	// r.HandleFunc("/authenticate", AuthenticateEndPoint).Methods("POST", "OPTIONS")
	r.HandleFunc("/check-login", CheckLoginEndPoint).Methods("POST", "OPTIONS")
	// r.HandleFunc("/refresh-token-auth").Methods("GET", "OPTIONS")
	// r.HandleFunc("/logout").Methods("GET", "OPTIONS")
}
