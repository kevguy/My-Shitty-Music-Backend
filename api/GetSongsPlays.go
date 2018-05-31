package api

import (
	"net/http"

	"github.com/kevguy/My-Shitty-Music-Backend/util"
)

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
