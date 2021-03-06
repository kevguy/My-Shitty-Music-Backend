package fcm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/kevguy/My-Shitty-Music-Backend/auth"
	"github.com/kevguy/My-Shitty-Music-Backend/mongodb"
	"github.com/kevguy/My-Shitty-Music-Backend/redis"
	"github.com/kevguy/My-Shitty-Music-Backend/util"
)

var shittyMusicDao mongodb.ShittyMusicDAO
var shittyMusicRedisDao redisclient.ShittyMusicRedisDAO
var authentication auth.JWTAuthentication
var fcmClient FcmClient

type TokenRequest struct {
	Token string `bson:"token" json:"token"`
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

func UpdateFCMTokenEndPoint(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// retrieve access_token
	// find user
	// grad fcm token
	// update user and save back to mongodb
	fmt.Println("fuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuckfuck")

	// data, _ := ioutil.ReadAll(r.Body)
	// log.Println("Authentication Request body: ", string(data))

	accessToken := r.Header.Get("x-access-token")
	fmt.Println("accessToken")
	fmt.Println(accessToken)
	userID := authentication.GetUserID(accessToken)
	fmt.Println("userID")
	fmt.Println(userID)

	if result := authentication.VerifyToken(userID, accessToken); !result {
		fmt.Println("token failed")
		util.RespondWithError(w, http.StatusBadRequest, "Invalid token")
		return
	}
	defer r.Body.Close()
	// fuck, _ := ioutil.ReadAll(r.Body)
	// log.Println("Authentication Request body: ", string(fuck))

	var tokenRequest TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&tokenRequest); err != nil {
		fmt.Println("Invalid request payload")
		fmt.Println(err.Error())
		util.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	fcmToken := tokenRequest.Token
	fcmClient.SubscribeToBroadcastTopic(fcmToken)

	user, err := shittyMusicDao.FindUserByID(userID)
	if err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Can't find user.")
		return
	}

	user.FcmToken = fcmToken
	if err := shittyMusicDao.UpdateUser(user); err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	util.RespondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
	return
}

func SendHelloEndPoint(w http.ResponseWriter, r *http.Request) {
	fcmClient.BroadcastHello()
	util.RespondWithJSON(w, http.StatusOK, "Okay")
}

func CreateFCMRoutes(r *mux.Router,
	_dao *mongodb.ShittyMusicDAO,
	_redisDao *redisclient.ShittyMusicRedisDAO,
	_authentication *auth.JWTAuthentication,
	_fcmClient *FcmClient) {
	fmt.Println("Setting up FCM Routes")
	shittyMusicDao = *_dao
	shittyMusicRedisDao = *_redisDao
	authentication = *_authentication
	fcmClient = *_fcmClient

	r.HandleFunc("/send-fcm-hello", SendHelloEndPoint).Methods("GET")

	r.Handle("/update-fcm-token", negroni.New(
		negroni.HandlerFunc(HandlePreflight),
		negroni.HandlerFunc(UpdateFCMTokenEndPoint),
	)).Methods("POST", "OPTIONS")

	// r.HandleFunc("/authenticate", AuthenticateEndPoint).Methods("POST", "OPTIONS")
	// r.HandleFunc("/check-login", CheckLoginEndPoint).Methods("POST", "OPTIONS")
	// r.HandleFunc("/refresh-token-auth").Methods("GET", "OPTIONS")
	// r.HandleFunc("/logout").Methods("GET", "OPTIONS")
}
