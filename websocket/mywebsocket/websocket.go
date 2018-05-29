package mywebsocket

import (
	"Redis-Exploration/websocket/dao"
	"Redis-Exploration/websocket/redis"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Define our message object
type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

var shittyMusicDao dao.ShittyMusicDAO
var shittyMusicRedisDao redisclient.ShittyMusicRedisDAO

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// For different origins
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func pingpong(ws *websocket.Conn, msgType int) error {
	time.Sleep(2 * time.Second)
	err := ws.WriteMessage(msgType, []byte("pong"))
	return err
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("creating websocket: %s\n", err)
		log.Fatalf("creating websocket: %s\n", err)
		return
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	for {
		var msg Message
		fmt.Printf("%+v\n", msg)
		// Read in a new message as JSON and map it to a Message object
		if err := ws.ReadJSON(&msg); err != nil {
			log.Printf("connection error: %s", err)
			delete(clients, ws)
			break
		}

		// Send the newly received message to the broadcast channel
		fmt.Printf("%+v\n", msg)
		broadcast <- msg
	}
}

func BroadcastMsg(msg Message) error {
	// Send it out to every client that is currently connected
	for client := range clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("error: %v", err)
			client.Close()
			delete(clients, client)
			return err
		}
	}
	return nil
}

func handleMessage() {
	for {
		// Grab the next message from the broadcast channel
		switch msg := <-broadcast; msg.Type {
		case "text":
			if string(msg.Content) == "ping" {
				msg.Content = "pong"
				BroadcastMsg(msg)
			} else {
				BroadcastMsg(msg)
			}
		case "play":
			fmt.Println("got play")
			handlePlay(msg.Content)
		case "upvote":
			handleUpvote(msg.Content)
		case "add_new_song":
			fmt.Println("Haven't implemented yet")
		default:
			fmt.Println("Haven't implemented yet")
		}
	}
}

func handleUpvote(songID string) {
	// Find song
	song, err := shittyMusicDao.FindSongByID(songID)
	if err != nil {
		// can't find song
		return
	}

	// Update song
	song.Upvotes++
	if err := shittyMusicDao.UpdateSong(song); err != nil {
		// can't update database
		return
	}

	msg := Message{
		Type:    "upvote",
		Content: songID + ":" + string(song.Upvotes),
	}

	BroadcastMsg(msg)
}

func handlePlay(songID string) {
	if err := shittyMusicRedisDao.PlaySong(songID); err != nil {
		if err != nil {
			// can't find song
			return
		}
	}

	// Find song
	song, err := shittyMusicDao.FindSongByID(songID)
	if err != nil {
		// can't find song
		return
	}

	// Update song
	song.Plays++
	if err := shittyMusicDao.UpdateSong(song); err != nil {
		// can't update database
		return
	}

	str := strconv.Itoa(song.Plays)
	msg := Message{
		Type:    "play",
		Content: songID + ":" + str,
	}

	BroadcastMsg(msg)
}

// CreateWebsocket sets up the websocket
func CreateWebsocket(r *mux.Router, _dao *dao.ShittyMusicDAO, _redisDao *redisclient.ShittyMusicRedisDAO) {
	fmt.Println("Setting up Websocket")
	shittyMusicDao = *_dao
	shittyMusicRedisDao = *_redisDao

	// Configure websocket route
	r.HandleFunc("/websocket", handleConnections)

	// Start listening for incoming chat messages
	go handleMessage()
}

// r.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
// 	ws, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	for {
// 		msgType, msg, err := ws.ReadMessage()
// 		if err != nil {
// 			fmt.Println(err)
// 			return
// 		}
// 		if string(msg) == "ping" {
// 			fmt.Println("ping")
// 			time.Sleep(2 * time.Second)
// 			err = ws.WriteMessage(msgType, []byte("pong"))
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 		} else {
// 			// ws.Close()
// 			fmt.Println(string(msg))
// 			var dat map[string]interface{}
// 			if err := json.Unmarshal(msg, &dat); err != nil {
// 				panic(err)
// 			}
// 			fmt.Println(dat)
// 			err = ws.WriteMessage(msgType, []byte(msg))
// 			if err != nil {
// 				fmt.Println(err)
// 				return
// 			}
// 			// return
// 		}
// 	}
// })
