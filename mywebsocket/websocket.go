package mywebsocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kevguy/My-Shitty-Music-Backend/auth"
	"github.com/kevguy/My-Shitty-Music-Backend/mongodb"
	"github.com/kevguy/My-Shitty-Music-Backend/redis"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Define our message object
type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type UpvoteMsg struct {
	SongID string `json:"song_id"`
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

var shittyMusicDao mongodb.ShittyMusicDAO
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

func handleUpvote(input string) {
	result := strings.Split(input, ":")
	if len(result) != 3 {
		return
	}

	token := result[0]
	userID := result[1]
	songID := result[2]

	authentication := auth.InitJWTAuthentication()
	fmt.Println(authentication)

	// Verify Token
	fmt.Println("Verifying Token")
	tokenValid := authentication.VerifyToken(userID, token)
	if !tokenValid {
		return
	}

	// Find song
	song, err := shittyMusicDao.FindSongByID(songID)
	if err != nil {
		// can't find song
		return
	}

	// Update song
	song.Upvotes++
	if err = shittyMusicDao.UpdateSong(song); err != nil {
		// can't update database
		return
	}

	user, err := shittyMusicDao.FindUserByID(userID)
	if err != nil {
		// can't find song
		return
	}
	user.Hearts = append(user.Hearts, songID)

	err = shittyMusicDao.UpdateUser(user)
	if err != nil {
		// can't update user
		return
	}

	str := strconv.Itoa(song.Upvotes)
	content, _ := json.Marshal(map[string]string{
		"userid":   user.ID.Hex(),
		"username": user.Name,
		"songid":   song.ID.Hex(),
		"song":     song.Name,
		"upvotes":  str,
	})
	msg := Message{
		Type:    "upvote",
		Content: string(content),
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
func CreateWebsocket(r *mux.Router, _dao *mongodb.ShittyMusicDAO, _redisDao *redisclient.ShittyMusicRedisDAO) {
	fmt.Println("Setting up Websocket")
	shittyMusicDao = *_dao
	shittyMusicRedisDao = *_redisDao

	// Configure websocket route
	r.HandleFunc("/websocket", handleConnections)

	// Start listening for incoming chat messages
	go handleMessage()
}
