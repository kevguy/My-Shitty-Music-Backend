package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"Redis-Exploration/websocket/api"
	"Redis-Exploration/websocket/dao"
	"Redis-Exploration/websocket/mywebsocket"
	"Redis-Exploration/websocket/util"

	"github.com/gorilla/mux"
)

var shittyMusicDAO = dao.ShittyMusicDAO{}

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

	shittyMusicDAO = dao.ShittyMusicDAO{
		Server:   os.Getenv("MONGOLAB_SERVER"),
		Database: os.Getenv("MONGOLAB_DATABASE"),
		Addr:     os.Getenv("MONGOLAB_ADDR"),
		Username: os.Getenv("MONGOLAB_USER"),
		Password: os.Getenv("MONGOLAB_PASSWORD"),
	}

	shittyMusicDAO.Connect()
}

func main() {
	initEnv()

	indexFile, err := os.Open("html/index.html")
	if err != nil {
		fmt.Println(err)
	}
	index, err := ioutil.ReadAll(indexFile)
	if err != nil {
		fmt.Println(err)
	}

	r := mux.NewRouter()

	mywebsocket.CreateWebsocket(r, &shittyMusicDAO)

	api.HandleApi(r, &shittyMusicDAO)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(index))
	})
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
