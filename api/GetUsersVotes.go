package api

import (
	"fmt"
	"net/http"

	"github.com/kevguy/My-Shitty-Music-Backend/util"

	"github.com/gorilla/mux"
)

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
