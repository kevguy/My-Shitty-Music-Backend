package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/kevguy/My-Shitty-Music-Backend/api"
	"github.com/kevguy/My-Shitty-Music-Backend/auth"
	"github.com/kevguy/My-Shitty-Music-Backend/fcm"
	"github.com/kevguy/My-Shitty-Music-Backend/mongodb"
	"github.com/kevguy/My-Shitty-Music-Backend/mywebsocket"
	"github.com/kevguy/My-Shitty-Music-Backend/redis"
	"github.com/kevguy/My-Shitty-Music-Backend/util"

	"github.com/gorilla/mux"
)

var shittyMusicDAO = mongodb.ShittyMusicDAO{}
var googleCredientials = auth.Credentials{}
var shittyMusicRedisDAO = redisclient.ShittyMusicRedisDAO{}
var fcmClient = fcm.FcmClient{}

// var store = sessions.NewCookieStore([]byte("something-very-secret"))

// Parse the configuration file 'config.toml', and establish a connection to DB
func initEnv() {
	fmt.Println("initEnv")

	// Read .env
	// then you can use for instance, os.Getenv("S3_BUCKET_NAME") to get the value
	fEnvFile := flag.String("env-file", "", "path to environment file")
	mode := flag.String("mode", "", "dev/production mode")
	flag.Parse()
	if *mode == "dev" {
		if *fEnvFile != "" {
			err := util.LoadEnvFile(*fEnvFile)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// Init MongoDB's Song DAO
	shittyMusicDAO = mongodb.ShittyMusicDAO{
		Server:   os.Getenv("MONGOLAB_SERVER"),
		Database: os.Getenv("MONGOLAB_DATABASE"),
		Addr:     os.Getenv("MONGOLAB_ADDR"),
		Username: os.Getenv("MONGOLAB_USER"),
		Password: os.Getenv("MONGOLAB_PASSWORD"),
	}
	shittyMusicDAO.Connect()

	googleCredientials = auth.Credentials{
		Cid:     os.Getenv("GOOGLE_AUTH_CLIENT_ID"),
		Csecret: os.Getenv("GOOGLE_AUTH_CLIENT_SECRET"),
	}

	// Init Redis's DAO
	shittyMusicRedisDAO = redisclient.ShittyMusicRedisDAO{
		Addr:     os.Getenv("REDIS_URI") + ":" + os.Getenv("REDIS_PORT"),
		Password: "",
		DB:       0,
	}
	shittyMusicRedisDAO.Connect()

	// store.Options = &sessions.Options{
	// 	// Domain:   "localhost",
	// 	// Path:     "/",
	// 	// MaxAge:   3600 * 8, // 8 hours
	// 	HttpOnly: true,
	// }
}

func main() {
	initEnv()

	fcmClient = *fcm.InitFcmClient()

	r := mux.NewRouter()

	// Setup websocket
	mywebsocket.CreateWebsocket(r, &shittyMusicDAO, &shittyMusicRedisDAO)

	authentication := auth.InitJWTAuthentication()
	auth.CreateRoutes(googleCredientials, r)
	auth.CreateAuthenticationRoutes(r, &shittyMusicDAO, &shittyMusicRedisDAO, authentication)

	// Set Fcm Calls
	fcm.CreateFCMRoutes(r, &shittyMusicDAO, &shittyMusicRedisDAO, authentication, &fcmClient)

	// Setup API Calls
	api.HandleAPI(r, &shittyMusicDAO, &shittyMusicRedisDAO, authentication, &fcmClient)

	// Serves index html page
	indexFile, err := os.Open("html/index.html")
	if err != nil {
		fmt.Println(err)
	}
	index, err := ioutil.ReadAll(indexFile)
	if err != nil {
		fmt.Println(err)
	}
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(index))
	})

	// corsObj := handlers.AllowedOrigins([]string{"*"})
	//
	// if err := http.ListenAndServe(":3000", handlers.CORS(corsObj)(r)); err != nil {
	// 	log.Fatal(err)
	// }

	port := os.Getenv("PORT")

	if port == "" {
		// log.Fatal("$PORT must be set")
		port = "3000"
	}

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
