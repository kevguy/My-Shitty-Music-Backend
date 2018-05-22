package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"Redis-Exploration/websocket/api"
	"Redis-Exploration/websocket/dao"
	"Redis-Exploration/websocket/util"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var shittyMusicDAO = dao.ShittyMusicDAO{}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// For different origins
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

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

	fmt.Println(os.Getenv("MONGOLAB_SERVER"))
	fmt.Println(os.Getenv("MONGOLAB_DATABASE"))
	shittyMusicDAO.Server = os.Getenv("MONGOLAB_SERVER")
	shittyMusicDAO.Database = os.Getenv("MONGOLAB_DATABASE")
	shittyMusicDAO.Addr = os.Getenv("MONGOLAB_ADDR")
	shittyMusicDAO.Username = os.Getenv("MONGOLAB_USER")
	shittyMusicDAO.Password = os.Getenv("MONGOLAB_PASSWORD")

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

	r.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		for {
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			if string(msg) == "ping" {
				fmt.Println("ping")
				time.Sleep(2 * time.Second)
				err = conn.WriteMessage(msgType, []byte("pong"))
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				// conn.Close()
				fmt.Println(string(msg))
				var dat map[string]interface{}
				if err := json.Unmarshal(msg, &dat); err != nil {
					panic(err)
				}
				fmt.Println(dat)
				err = conn.WriteMessage(msgType, []byte(msg))
				if err != nil {
					fmt.Println(err)
					return
				}
				// return
			}
		}
	})

	api.HandleApi(r, &shittyMusicDAO)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(index))
	})
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}
}
