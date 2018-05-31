package api

import (
	"fmt"

	"github.com/kevguy/My-Shitty-Music-Backend/auth"
	"github.com/kevguy/My-Shitty-Music-Backend/mongodb"
	"github.com/kevguy/My-Shitty-Music-Backend/redis"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

var shittyMusicDao mongodb.ShittyMusicDAO
var shittyMusicRedisDao redisclient.ShittyMusicRedisDAO
var authentication auth.JWTAuthentication

// HandleAPI sets up how to handle API calls
func HandleAPI(r *mux.Router,
	_dao *mongodb.ShittyMusicDAO,
	_redisDao *redisclient.ShittyMusicRedisDAO,
	_authentication *auth.JWTAuthentication) {

	fmt.Println("Setting up Api Calls")
	shittyMusicDao = *_dao
	shittyMusicRedisDao = *_redisDao
	authentication = *_authentication
	fmt.Println(authentication)

	r.HandleFunc("/songs/plays", GetSongsPlaysEndPoint).Methods("GET")
	r.HandleFunc("/songs", AllSongsEndPoint).Methods("GET")
	r.HandleFunc("/songs", UpdateSongEndPoint).Methods("PUT", "OPTIONS")
	r.HandleFunc("/songs", DeleteSongEndPoint).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/songs/{id}", FindSongEndpoint).Methods("GET")

	r.Handle("/add-song", negroni.New(
		negroni.HandlerFunc(HandlePreflight),
		negroni.HandlerFunc(CreateSongEndPoint),
	)).Methods("POST", "OPTIONS")

	r.Handle("/users/upvotes/{id}", negroni.New(
		negroni.HandlerFunc(HandlePreflight),
		negroni.HandlerFunc(GetUserUpvotesEndPoint),
	)).Methods("GET", "OPTIONS")
}
