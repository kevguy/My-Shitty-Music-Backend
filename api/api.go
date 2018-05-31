package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"Redis-Exploration/dao"
	. "Redis-Exploration/models"
	"Redis-Exploration/mywebsocket"
	"Redis-Exploration/redis"
	"Redis-Exploration/util"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

var shittyMusicDao dao.ShittyMusicDAO
var shittyMusicRedisDao redisclient.ShittyMusicRedisDAO

// AllSongsEndPoint finds all songs
func AllSongsEndPoint(w http.ResponseWriter, r *http.Request) {
	songs, err := shittyMusicDao.FindAllSongs()
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	util.RespondWithJSON(w, http.StatusOK, songs)
	mywebsocket.BroadcastMsg(mywebsocket.Message{
		Type:    "text",
		Content: "AllSongsEndPoint",
	})
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
	mywebsocket.BroadcastMsg(mywebsocket.Message{
		Type:    "text",
		Content: "FindSongEndpoint",
	})
}

// CreateSongEndPoint saves a new song
func CreateSongEndPoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println("CreateSongEndPoint")
	defer r.Body.Close()
	var song Song
	fmt.Println(song.Name)
	if err := json.NewDecoder(r.Body).Decode(&song); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
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
		Content: "CreateSongEndPoint",
	})
}

// UpdateSongEndPoint updates the information of a song
func UpdateSongEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var song Song
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
	var song Song
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

func GetUserUpvotesEndPoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
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

// HandleAPI sets up how to handle API calls
func HandleAPI(r *mux.Router, _dao *dao.ShittyMusicDAO, _redisDao *redisclient.ShittyMusicRedisDAO) {
	fmt.Println("Setting up Api Calls")
	shittyMusicDao = *_dao
	shittyMusicRedisDao = *_redisDao

	r.HandleFunc("/songs/plays", GetSongsPlaysEndPoint).Methods("GET")
	r.HandleFunc("/songs", AllSongsEndPoint).Methods("GET")
	r.HandleFunc("/songs", CreateSongEndPoint).Methods("POST", "OPTIONS")
	r.HandleFunc("/songs", UpdateSongEndPoint).Methods("PUT", "OPTIONS")
	r.HandleFunc("/songs", DeleteSongEndPoint).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/songs/{id}", FindSongEndpoint).Methods("GET")

	r.HandleFunc("/users/upvotes/{id}", GetUserUpvotesEndPoint).Methods("GET", "OPTIONS")
}
