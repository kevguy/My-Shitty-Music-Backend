package googleauth

import (
	"Redis-Exploration/websocket/dao"
	. "Redis-Exploration/websocket/models"
	"Redis-Exploration/websocket/redis"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"Redis-Exploration/websocket/util"
)

var shittyMusicDao dao.ShittyMusicDAO
var shittyMusicRedisDao redisclient.ShittyMusicRedisDAO

type GoogleRequest struct {
	Type string `bson:"type" json:"type"`
	Code string `bson:"code" json:"code"`
}

func AuthenticateEndPoint(w http.ResponseWriter, r *http.Request) {
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
	// util.RespondWithJSON(w, http.StatusOK, googleRequest)
	if googleRequest.Type == "google" {
		authentication := InitJWTAuthentication()
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
		token, err := authentication.GenerateToken(user.ID.String())
		if err != nil {
			panic(err)
		}
		if err != nil {
			util.RespondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		response, _ := json.Marshal(map[string]string{"token": token})
		fmt.Println(string(response))
		util.RespondWithJSON(w, http.StatusOK, map[string]string{"token": token})
		// util.RespondWithJSON(w, http.StatusOK, googleRequest)
	} else {
		util.RespondWithError(w, http.StatusUnauthorized, "Not Supported")
	}
	return
	/*
		if googleRequest.Type == "google" {
			authentication := InitJWTAuthentication()

			code := googleRequest.Code
			// Retrieve Google Profile
			profile := RetrieveGoogleUserProfile(code)

			// Find user
			fmt.Println(profile.DisplayName)
			user, err := shittyMusicDao.FindGoogleUser(profile.ID)

			// User not found ==> Create New user

			// Generate token
			token, err := authentication.GenerateToken(user.UUID)
			if err != nil {
				panic(err)
			}
			if err != nil {
				util.RespondWithError(w, http.StatusInternalServerError, err.Error())
				return
			}

			response, _ := json.Marshal(map[string]string{"token": token})
			util.RespondWithJSON(w, http.StatusOK, response)

		} else {
			util.RespondWithError(w, http.StatusUnauthorized, "Not Supported")
			return
		}
	*/
}

func CreateAuthenticationRoutes(r *mux.Router, _dao *dao.ShittyMusicDAO, _redisDao *redisclient.ShittyMusicRedisDAO) {
	fmt.Println("Setting up Authentication Routes")
	shittyMusicDao = *_dao
	shittyMusicRedisDao = *_redisDao

	r.HandleFunc("/authenticate", AuthenticateEndPoint).Methods("POST", "OPTIONS")
	// r.HandleFunc("/refresh-token-auth").Methods("GET", "OPTIONS")
	// r.HandleFunc("/logout").Methods("GET", "OPTIONS")
}
