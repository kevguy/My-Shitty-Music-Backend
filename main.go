package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"Redis-Exploration/api"
	"Redis-Exploration/dao"
	"Redis-Exploration/googleauth"
	"Redis-Exploration/mywebsocket"
	"Redis-Exploration/redis"
	"Redis-Exploration/util"

	"github.com/gorilla/mux"
)

var shittyMusicDAO = dao.ShittyMusicDAO{}
var googleCredientials = googleauth.Credentials{}
var shittyMusicRedisDAO = redisclient.ShittyMusicRedisDAO{}

// var store = sessions.NewCookieStore([]byte("something-very-secret"))

// Parse the configuration file 'config.toml', and establish a connection to DB
func initEnv() {
	fmt.Println("initEnv")
	// Read .env
	// then you can use for instance, os.Getenv("S3_BUCKET_NAME") to get the value
	fEnvFile := flag.String("env-file", "", "path to environment file")
	flag.Parse()

	if *fEnvFile != "" {
		err := util.LoadEnvFile(*fEnvFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Init MongoDB's Song DAO
	shittyMusicDAO = dao.ShittyMusicDAO{
		Server:   os.Getenv("MONGOLAB_SERVER"),
		Database: os.Getenv("MONGOLAB_DATABASE"),
		Addr:     os.Getenv("MONGOLAB_ADDR"),
		Username: os.Getenv("MONGOLAB_USER"),
		Password: os.Getenv("MONGOLAB_PASSWORD"),
	}
	shittyMusicDAO.Connect()

	googleCredientials = googleauth.Credentials{
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

	r := mux.NewRouter()

	// Setup websocket
	mywebsocket.CreateWebsocket(r, &shittyMusicDAO, &shittyMusicRedisDAO)

	authentication := googleauth.InitJWTAuthentication()
	googleauth.CreateRoutes(googleCredientials, r, &shittyMusicDAO)
	googleauth.CreateAuthenticationRoutes(r, &shittyMusicDAO, &shittyMusicRedisDAO, authentication)

	// Setup API Calls
	api.HandleAPI(r, &shittyMusicDAO, &shittyMusicRedisDAO, authentication)

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
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
