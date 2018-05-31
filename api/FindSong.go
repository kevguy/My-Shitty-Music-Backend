package api

import (
	"net/http"

	"github.com/kevguy/My-Shitty-Music-Backend/util"

	"github.com/gorilla/mux"
)

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
