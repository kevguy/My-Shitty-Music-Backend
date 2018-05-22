package mywebsocket

import (
	"Redis-Exploration/websocket/dao"
	"fmt"
	"log"
	"net/http"
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

func broadcastMsg(msg Message) error {
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
		msg := <-broadcast

		if msg.Type == "text" {
			if string(msg.Content) == "ping" {
				msg.Content = "pong"
				broadcastMsg(msg)
			} else {
				broadcastMsg(msg)
			}
		} else {
			// Parse content (i.e. msg.Content)
		}
	}
}

func CreateWebsocket(r *mux.Router, _dao *dao.ShittyMusicDAO) {
	// Configure websocket route
	r.HandleFunc("/websocket", handleConnections)

	// Start listening for incoming chat messages
	go handleMessage()

	// r.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
	// 	ws, err := upgrader.Upgrade(w, r, nil)
	// 	if err != nil {
	// 		fmt.Printf("creating websocket: %s\n", err)
	// 		log.Fatalf("creating websocket: %s\n", err)
	// 		return
	// 	}
	//
	// 	for {
	// 		msgType, msg, err := ws.ReadMessage()
	// 		if err != nil {
	// 			fmt.Printf("message error: %s\n", err)
	// 			// return
	// 		}
	//
	// 		// Handle Ping Message
	// 		fmt.Println(string(msg))
	// 		fmt.Println(util.IsJSON(string(msg)))
	// 		if string(msg) == "ping" {
	// 			err = pingpong(ws, msgType)
	// 			if err != nil {
	// 				fmt.Printf("ping message error: %s\n", err)
	// 				// return
	// 			}
	// 		} else if util.IsJSON(string(msg)) {
	// 			var data map[string]interface{}
	// 			if err = json.Unmarshal(msg, &data); err != nil {
	// 				fmt.Printf("parsing JSON string: %s\n", err)
	// 				// return
	// 			}
	//
	// 			fmt.Println(data)
	// 			fmt.Println(data["a"])
	// 			if err = ws.WriteMessage(msgType, []byte(msg)); err != nil {
	// 				fmt.Printf("sending ws reply: %s\n", err)
	// 			}
	// 		}
	// 	}
	// })
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
