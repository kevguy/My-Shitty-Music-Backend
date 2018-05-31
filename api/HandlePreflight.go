package api

import (
	"fmt"
	"net/http"

	"github.com/kevguy/My-Shitty-Music-Backend/util"
)

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
