package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kevguy/My-Shitty-Music-Backend/models"
	"github.com/kevguy/My-Shitty-Music-Backend/mywebsocket"
	"github.com/kevguy/My-Shitty-Music-Backend/util"

	"gopkg.in/mgo.v2/bson"
)

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
