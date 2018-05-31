package api

import (
	"net/http"

	"github.com/kevguy/My-Shitty-Music-Backend/util"
)

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
