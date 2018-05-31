package api

import (
	"encoding/json"
	"net/http"

	"github.com/kevguy/My-Shitty-Music-Backend/models"
	"github.com/kevguy/My-Shitty-Music-Backend/util"
)

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
