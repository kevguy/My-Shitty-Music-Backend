package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"My-Shitty-Music-Backend/dao"
	"My-Shitty-Music-Backend/googleauth"
	"My-Shitty-Music-Backend/models"
	"My-Shitty-Music-Backend/mywebsocket"
	"My-Shitty-Music-Backend/redis"
	"My-Shitty-Music-Backend/util"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

var shittyMusicDao dao.ShittyMusicDAO
var shittyMusicRedisDao redisclient.ShittyMusicRedisDAO
var authentication googleauth.JWTAuthentication

// AllSongsEndPoint finds all songs
func AllSongsEndPoint(w http.ResponseWriter, r *http.Request) {
	songs, err := shittyMusicDao.FindAllSongs()
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.RespondWithJSON(w, http.StatusOK, songs)
	// mywebsocket.BroadcastMsg(mywebsocket.Message{
	// 	Type:    "text",
	// 	Content: "AllSongsEndPoint",
	// })
}

// FindSongEndpoint finds a song
func FindSongEndpoint(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	song, err := shittyMusicDao.FindSongByID(params["id"])
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid Song ID")
		return
	}
	util.RespondWithJSON(w, http.StatusOK, song)
	// mywebsocket.BroadcastMsg(mywebsocket.Message{
	// 	Type:    "text",
	// 	Content: "FindSongEndpoint",
	// })
}

// UpdateSongEndPoint updates the information of a song
func UpdateSongEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var song models.Song
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := shittyMusicDao.UpdateSong(song); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.RespondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// DeleteSongEndPoint deletes a song
func DeleteSongEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var song models.Song
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if err := shittyMusicDao.DeleteSong(song); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.RespondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// GetSongsPlaysEndPoint retrieves the number of plays for every song (from Redis)
func GetSongsPlaysEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	vals, err := shittyMusicRedisDao.GetPlays()
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.RespondWithJSON(w, http.StatusOK, vals)
}

// CreateSongEndPoint saves a new song
func CreateSongEndPoint(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fmt.Println("CreateSongEndPoint")

	fmt.Println(r.Header.Get("Accept"))
	fmt.Println(r.Header.Get("Content-Type"))
	fmt.Println(r.Header.Get("x-access-token"))
	token := r.Header.Get("x-access-token")
	fmt.Println(token)
	userID := authentication.GetUserID(token)
	fmt.Println(userID)

	if result := authentication.VerifyToken(userID, token); !result {
		fmt.Println("token failed")
		util.RespondWithError(w, http.StatusBadRequest, "Invalid token")
		return
	}

	defer r.Body.Close()

	var song models.Song
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		fmt.Println(err.Error())
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := shittyMusicDao.FindUserByID(userID)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Can't find user.")
		return
	}
	fmt.Println(user)

	song.ID = bson.NewObjectId()
	song.Upvotes = 0
	song.Plays = 0
	fmt.Println("Trying to insert song")
	fmt.Println(song.Name)
	if err := shittyMusicDao.InsertSong(song); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := shittyMusicRedisDao.InitSong(song.ID.String(), 0, 0); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.RespondWithJSON(w, http.StatusCreated, song)
	mywebsocket.BroadcastMsg(mywebsocket.Message{
		Type:    "text",
		Content: user.Name + " added a new song",
	})
}

func GetUserUpvotesEndPoint(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fmt.Println("GetUserUpvotesEndPoint")
	defer r.Body.Close()

	fmt.Println(r.Header.Get("Accept"))
	fmt.Println(r.Header.Get("Content-Type"))
	fmt.Println(r.Header.Get("x-access-token"))

	token := r.Header.Get("x-access-token")

	userID := authentication.GetUserID(r.Header.Get("x-access-token"))
	fmt.Println(userID)

	if result := authentication.VerifyToken(userID, token); !result {
		fmt.Println("token failed")
		util.RespondWithError(w, http.StatusBadRequest, "Invalid token")
	}

	params := mux.Vars(r)
	user, err := shittyMusicDao.FindUserByID(params["id"])
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid User ID")
		return
	}
	util.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"upvotes": user.Hearts,
	})
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

// HandleAPI sets up how to handle API calls
func HandleAPI(r *mux.Router,
	_dao *dao.ShittyMusicDAO,
	_redisDao *redisclient.ShittyMusicRedisDAO,
	_authentication *googleauth.JWTAuthentication) {
	fmt.Println("Setting up Api Calls")
	shittyMusicDao = *_dao
	shittyMusicRedisDao = *_redisDao
	authentication = *_authentication
	fmt.Println(authentication)

	r.HandleFunc("/songs/plays", GetSongsPlaysEndPoint).Methods("GET")
	r.HandleFunc("/songs", AllSongsEndPoint).Methods("GET")
	// r.HandleFunc("/songs", CreateSongEndPoint).Methods("POST", "OPTIONS")
	r.HandleFunc("/songs", UpdateSongEndPoint).Methods("PUT", "OPTIONS")
	r.HandleFunc("/songs", DeleteSongEndPoint).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/songs/{id}", FindSongEndpoint).Methods("GET")

	// r.HandleFunc("/songs", CreateSongEndPoint).Methods("POST", "OPTIONS")
	r.Handle("/add-song", negroni.New(
		negroni.HandlerFunc(HandlePreflight),
		negroni.HandlerFunc(CreateSongEndPoint),
	)).Methods("POST", "OPTIONS")

	r.Handle("/users/upvotes/{id}", negroni.New(
		negroni.HandlerFunc(HandlePreflight),
		negroni.HandlerFunc(GetUserUpvotesEndPoint),
	)).Methods("GET", "OPTIONS")

	// r.HandleFunc("/users/upvotes/{id}", GetUserUpvotesEndPoint).Methods("GET", "OPTIONS")
}
