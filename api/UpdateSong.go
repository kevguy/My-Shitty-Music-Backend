package api

import (
	"encoding/json"
	"net/http"

	"github.com/kevguy/My-Shitty-Music-Backend/models"
	"github.com/kevguy/My-Shitty-Music-Backend/util"
)

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
