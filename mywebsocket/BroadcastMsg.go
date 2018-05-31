package mywebsocket

import "log"

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
